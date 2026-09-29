package main

// This file ships the two remaining DX-04 anti-theater guards (D-06-27.1,
// D-06-27.2) plus the fixture-substitution subprocess they are both built
// on. It is a _test.go file, so it is explicitly EXEMPT from the
// import-boundary lint (import_boundary_test.go scans only non-test
// files) -- exactly like repair_test.go, it imports internal/ packages to
// DRIVE the real injectors and the real shipped `lang` binary. Task 3
// (D-06-27.2) required one small, deliberate fix to repair.go itself:
// driverEligible now also requires a non-empty Kind, mirroring the same
// fix in diagnostic.DriverEligible -- see the "strip_kind" deviation note
// in 06-14-SUMMARY.md for why.
//
// Mechanism (06-RESEARCH.md Open Question 2): a fixture-substitution
// subprocess, not a driver-side test hook. The stand-in program built by
// buildStandin below is passed as the driver's own `--lang` path -- the
// driver spawns it exactly the way it spawns the real `lang` binary,
// reads its stdout, and never learns the difference. This costs the
// driver ZERO code changes: no second, untested code path is created
// inside cmd/schway-repair.
//
// The stand-in resolves WHICH captured document to serve by hashing the
// content of the source file it was told to check (its last argv, the
// same sourcePath the driver passes on every invocation) -- never by
// counting invocations or inspecting argv shape. This is what lets one
// stand-in binary, built once, correctly answer BOTH of the driver's two
// calls in a single-pass repair cycle (diagnose on the pre-repair bytes,
// reverify on the post-repair bytes) without any shared mutable state
// beyond the read-only captures directory the test populates in advance.
import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// standinDirEnv names the environment variable the stand-in reads to find
// its captures directory. The real driver never sets or reads this
// variable -- it is purely a fixture of THIS test file, inherited by the
// stand-in subprocess because os/exec defaults Cmd.Env to the parent
// process's environment when Env is left nil (exactly what runLangCheck
// does), so t.Setenv in the test process reaches the stand-in without any
// driver-side plumbing.
const standinDirEnv = "SCHWAY_REPAIR_STANDIN_DIR"

// standinSource is the fixture-substitution stand-in's entire program. It
// takes no flags of its own and does not even look at "--json check" --
// it reads its LAST argument (the source file path the driver always
// passes last), hashes that file's CURRENT on-disk content, and writes
// the captured document filed under that hash to stdout. A file the test
// harness never staged a capture for is refused with a nonzero exit
// (never silently empty output), which is what makes a captures-directory
// bug fail loudly rather than masquerade as a real "schway" crash.
const standinSource = `package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	dir := os.Getenv("SCHWAY_REPAIR_STANDIN_DIR")
	if dir == "" || len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "standin: SCHWAY_REPAIR_STANDIN_DIR unset or no source argument")
		os.Exit(2)
	}
	sourcePath := os.Args[len(os.Args)-1]
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "standin: reading source: "+err.Error())
		os.Exit(2)
	}
	sum := sha256.Sum256(data)
	key := hex.EncodeToString(sum[:])
	capture, err := os.ReadFile(dir + "/" + key + ".json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "standin: no staged capture for content hash "+key)
		os.Exit(2)
	}
	os.Stdout.Write(capture)
}
`

// standinBinaryOnce builds the stand-in exactly once per test process --
// it is stateless and content-hash-keyed, so every test in this file
// (and every subtest) safely shares the one compiled binary.
var standinBinaryOnce = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "schway-repair-standin")
	if err != nil {
		return "", err
	}
	srcPath := filepath.Join(dir, "standin.go")
	if err := os.WriteFile(srcPath, []byte(standinSource), 0o644); err != nil {
		return "", err
	}
	binPath := filepath.Join(dir, "standin")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", binPath, srcPath)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("building fixture-substitution stand-in: %w (%s)", err, stderr.String())
	}
	return binPath, nil
})

func buildStandin(t testing.TB) string {
	t.Helper()
	binary, err := standinBinaryOnce()
	if err != nil {
		t.Fatal(err)
	}
	return binary
}

// writeCaptureForContent stages capture as the document the stand-in must
// serve when it is invoked against a source file whose CURRENT bytes are
// exactly content -- keyed by content hash, never by call order.
func writeCaptureForContent(t testing.TB, dir string, content, capture []byte) {
	t.Helper()
	sum := sha256.Sum256(content)
	key := hex.EncodeToString(sum[:])
	if err := os.WriteFile(filepath.Join(dir, key+".json"), capture, 0o644); err != nil {
		t.Fatal(err)
	}
}

// runCapturingStdout invokes run (main.go's own entry point, callable
// directly since this file lives in package main) with os.Stdout
// temporarily redirected, so a test can assert on the exact same exit
// code and JSON envelope a real `schway-repair --json` invocation would
// produce, without spawning a THIRD-level subprocess to do it.
func runCapturingStdout(t testing.TB, args []string) (exitCode int, stdout []byte) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	exitCode = run(args)
	if cerr := w.Close(); cerr != nil {
		t.Fatal(cerr)
	}
	os.Stdout = old
	stdout, err = io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return exitCode, stdout
}

// writeTempCopy writes content to a fresh file inside t.TempDir() and
// returns its path -- used whenever a test needs an independent on-disk
// copy of already-captured bytes (e.g. to re-run `lang --json check`
// against a specific historical state without disturbing the file the
// driver itself is mutating).
func writeTempCopy(t testing.TB, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "copy.schway")
	mustWriteFile(t, path, content)
	return path
}

// captureFile reads one of the checked-in per-class baseline captures
// this plan's Task 1 committed to cmd/schway-repair/testdata/. They were
// produced by running the real shipped `lang` binary against each
// injector's real mutated held-out fixture -- never fabricated by hand
// (see the plan's Task 1 for the exact generation procedure).
func captureFile(t testing.TB, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// TestBaselineCaptureContainsEligibleRepair is the corpus-quality guard
// D-06-27's own falsifiability rests on: a baseline capture with zero
// eligible repairs would make every downstream guard vacuous, since there
// would be nothing for the driver to apply either scrambled or intact.
func TestBaselineCaptureContainsEligibleRepair(t *testing.T) {
	for _, class := range []string{"match", "move", "borrow", "interprocedural_loan"} {
		t.Run(class, func(t *testing.T) {
			raw := captureFile(t, class+"_diagnose_capture.json")
			decoded := decodeCheckJSON(t, raw)
			_, code, ok := selectRepair(decoded)
			if !ok {
				t.Fatalf("%s_diagnose_capture.json carries zero driver-eligible repairs (MachineApplicable with both span and replacement) -- this baseline cannot exercise any guard", class)
			}
			if code == "" {
				t.Fatalf("%s_diagnose_capture.json: selected repair carries no diagnosis code", class)
			}
		})
	}
}

// TestFixtureSubstitutionIsFaithful is the tracer (Task 1): drives the
// SAME injected move defect through the REAL built `lang` binary and,
// independently, through the fixture-substitution stand-in serving real
// captures of that exact real binary's own output -- end to end, through
// the real driver's Repair cycle both times -- and asserts byte-identical
// repaired source and identical CLI exit codes. This is what proves the
// substitution is faithful before either anti-theater guard is allowed to
// depend on it.
func TestFixtureSubstitutionIsFaithful(t *testing.T) {
	realBinary := testsupport.BuildCLI(t)
	standinBinary := buildStandin(t)

	original := phase6Fixture(t, "heldout_move_defect.schway")
	mutated, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatalf("injecting move defect: %v", err)
	}

	// Real path: the driver spawns the real binary directly.
	realSourcePath := writeTempCopy(t, mutated)
	realExit, _ := runCapturingStdout(t, []string{"--lang=" + realBinary, "--source=" + realSourcePath})
	realRepaired, err := os.ReadFile(realSourcePath)
	if err != nil {
		t.Fatal(err)
	}

	// Capture the real binary's own two documents for exactly this
	// mutated/repaired byte pair, independently of the driver run above,
	// then stage them for the stand-in keyed by content hash.
	standinDir := t.TempDir()
	preCapture := testsupport.RunCLI(t, realBinary, nil, "--json", "check", writeTempCopy(t, mutated)).Stdout
	writeCaptureForContent(t, standinDir, mutated, mutateCaptureOrFatal(t, preCapture, "identity"))
	postCapture := testsupport.RunCLI(t, realBinary, nil, "--json", "check", writeTempCopy(t, realRepaired)).Stdout
	writeCaptureForContent(t, standinDir, realRepaired, mutateCaptureOrFatal(t, postCapture, "identity"))

	standinSourcePath := writeTempCopy(t, mutated)
	t.Setenv(standinDirEnv, standinDir)
	standinExit, _ := runCapturingStdout(t, []string{"--lang=" + standinBinary, "--source=" + standinSourcePath})
	standinRepaired, err := os.ReadFile(standinSourcePath)
	if err != nil {
		t.Fatal(err)
	}

	if standinExit != realExit {
		t.Fatalf("exit codes differ: real=%d standin=%d", realExit, standinExit)
	}
	if !bytes.Equal(standinRepaired, realRepaired) {
		t.Fatalf("stand-in-driven repair bytes differ from the real-binary-driven repair bytes:\nreal:    %q\nstandin: %q", realRepaired, standinRepaired)
	}
}

func mutateCaptureOrFatal(t testing.TB, raw []byte, mode string) []byte {
	t.Helper()
	mutated, err := mutateCapture(raw, mode)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}

// ---------------------------------------------------------------------
// mutateCapture -- the shared JSON-mutation primitive every guard in this
// file drives the stand-in with. It decodes the captured document
// generically (never through the driver's own narrow structs, since a
// guard specifically needs to inspect/mutate fields -- like message and
// detail -- the driver itself never decodes at all) and re-marshals after
// applying exactly one named mutation.
// ---------------------------------------------------------------------

const loremIpsumProse = "lorem ipsum dolor sit amet consectetur"

func mutateCapture(raw []byte, mode string) ([]byte, error) {
	var doc map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("mutateCapture: decoding: %w", err)
	}
	switch mode {
	case "identity":
		// No mutation -- exists so every guard's "compare against the
		// unscrambled/unstripped baseline" step runs the SAME decode ->
		// marshal pipeline as the mutated side, isolating any observed
		// difference to the mutation itself rather than to marshal
		// nondeterminism.
	case "scramble_prose":
		scrambleProse(doc)
	case "strip_repairs":
		forEachDiagnostic(doc, func(diag map[string]interface{}) {
			delete(diag, "repairs")
		})
	case "strip_kind", "strip_span", "strip_replacement":
		field := strings.TrimPrefix(mode, "strip_")
		forEachRepair(doc, func(repair map[string]interface{}) {
			delete(repair, field)
		})
	case "corrupt_kind":
		forEachRepair(doc, func(repair map[string]interface{}) { repair["kind"] = "" })
	case "corrupt_span":
		forEachRepair(doc, func(repair map[string]interface{}) {
			if span, ok := repair["span"].(map[string]interface{}); ok {
				span["start"] = json.Number("999999")
			}
		})
	case "corrupt_replacement":
		forEachRepair(doc, func(repair map[string]interface{}) { repair["replacement"] = "value" })
	default:
		return nil, fmt.Errorf("mutateCapture: unknown mode %q", mode)
	}
	return json.Marshal(doc)
}

// TestPhase17RepairProtocolFields proves the newly reachable repair remains
// wholly protocol-driven.  The stand-in serves captures keyed by source
// bytes, so every run follows Repair's real diagnose/apply/reverify path;
// only the selected JSON field is changed.
func TestPhase17RepairProtocolFields(t *testing.T) {
	langBinary := testsupport.BuildCLI(t)
	standinBinary := buildStandin(t)
	original, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "heldout_call_argument_mismatch.schway"))
	if err != nil {
		t.Fatal(err)
	}
	wantRepaired := bytes.Replace(original, []byte("classify(value)"), []byte("classify(resource)"), 1)
	diagnose := testsupport.RunCLI(t, langBinary, nil, "--json", "check", writeTempCopy(t, original)).Stdout
	reverify := testsupport.RunCLI(t, langBinary, nil, "--json", "check", writeTempCopy(t, wantRepaired)).Stdout

	run := func(mode string) (int, Outcome, []byte) {
		t.Helper()
		return runViaStandin(t, standinBinary, original, diagnose, reverify, mode)
	}

	identityExit, identity, identityBytes := run("identity")
	if identityExit != 0 || identity.Status != OutcomeRepaired || !bytes.Equal(identityBytes, wantRepaired) {
		t.Fatalf("identity control did not repair the held-out source: exit=%d outcome=%+v", identityExit, identity)
	}
	scrambledExit, scrambled, scrambledBytes := run("scramble_prose")
	if scrambledExit != identityExit || scrambled != identity || !bytes.Equal(scrambledBytes, identityBytes) {
		t.Fatalf("prose scrambling changed repair outcome: identity=%+v scrambled=%+v", identity, scrambled)
	}
	for _, mode := range []string{"strip_kind", "strip_span", "strip_replacement", "corrupt_kind", "corrupt_replacement"} {
		t.Run(mode, func(t *testing.T) {
			exit, outcome, repaired := run(mode)
			if exit == 0 || outcome.Status == OutcomeRepaired {
				t.Fatalf("%s accepted a corrupted required protocol field: exit=%d outcome=%+v", mode, exit, outcome)
			}
			if outcome.Status == OutcomeUnrepairable && !bytes.Equal(repaired, original) {
				t.Fatalf("%s modified source despite refusing repair", mode)
			}
		})
	}
	// An out-of-range changed span cannot reach a reverify at all: applyRepair
	// refuses it before writing. Exercise that failure directly so the test can
	// distinguish a bounded-edit rejection from an unrelated stand-in response.
	corruptSpan := mutateCaptureOrFatal(t, diagnose, "corrupt_span")
	standinDir := t.TempDir()
	writeCaptureForContent(t, standinDir, original, corruptSpan)
	t.Setenv(standinDirEnv, standinDir)
	path := writeTempCopy(t, original)
	if _, err := Repair(context.Background(), standinBinary, path); err == nil {
		t.Fatal("corrupt_span was accepted; an out-of-range structured span must refuse repair")
	}
	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, original) {
		t.Fatal("corrupt_span modified source before refusing repair")
	}
}

// forEachDiagnostic walks doc["diagnostics"] (a []interface{} of
// map[string]interface{}), calling visit on each one in place.
func forEachDiagnostic(doc map[string]interface{}, visit func(map[string]interface{})) {
	diagnostics, ok := doc["diagnostics"].([]interface{})
	if !ok {
		return
	}
	for _, d := range diagnostics {
		if diag, ok := d.(map[string]interface{}); ok {
			visit(diag)
		}
	}
}

// forEachRepair walks every diagnostic's own repairs[] array, calling
// visit on each repair object in place. Scoped specifically to repairs[]
// (never causes[], which also carries its own "kind") so a "strip_kind"
// mode strips ONLY repair kinds, never cause kinds.
func forEachRepair(doc map[string]interface{}, visit func(map[string]interface{})) {
	forEachDiagnostic(doc, func(diag map[string]interface{}) {
		repairs, ok := diag["repairs"].([]interface{})
		if !ok {
			return
		}
		for _, r := range repairs {
			if repair, ok := r.(map[string]interface{}); ok {
				visit(repair)
			}
		}
	})
}

// scrambleProse walks v recursively -- maps and slices, at every nesting
// depth -- replacing every STRING value found under a "message" or
// "detail" key with a fixed lorem-ipsum string. A fixed replacement
// (never randomized) is required so a failure is reproducible (per the
// plan's own instruction).
func scrambleProse(v interface{}) {
	switch val := v.(type) {
	case map[string]interface{}:
		for k, child := range val {
			if k == "message" || k == "detail" {
				if _, isString := child.(string); isString {
					val[k] = loremIpsumProse
					continue
				}
			}
			scrambleProse(child)
		}
	case []interface{}:
		for _, item := range val {
			scrambleProse(item)
		}
	}
}

// assertOnlyProseDiffers walks before and after IN LOCKSTEP (same JSON
// document, before and after a scramble_prose mutation) and fails the
// test the moment any non-message/non-detail value differs. It reports,
// via changed, whether at least one prose value actually changed --
// callers use this to refuse a scrambler that silently matched nothing
// (T-06-THEATER-02/T-06-THEATER-05).
func assertOnlyProseDiffers(t testing.TB, before, after interface{}, path string, changed *bool) {
	t.Helper()
	switch b := before.(type) {
	case map[string]interface{}:
		a, ok := after.(map[string]interface{})
		if !ok {
			t.Fatalf("%s: type changed across the scramble (map -> %T)", path, after)
		}
		if len(a) != len(b) {
			t.Fatalf("%s: key set changed across the scramble (%d keys -> %d keys)", path, len(b), len(a))
		}
		for k, bv := range b {
			av, present := a[k]
			if !present {
				t.Fatalf("%s.%s: key disappeared across the scramble", path, k)
			}
			if k == "message" || k == "detail" {
				bs, bIsStr := bv.(string)
				as, aIsStr := av.(string)
				if bIsStr && aIsStr {
					if bs != as {
						*changed = true
					}
					continue
				}
			}
			assertOnlyProseDiffers(t, bv, av, path+"."+k, changed)
		}
	case []interface{}:
		a, ok := after.([]interface{})
		if !ok || len(a) != len(b) {
			t.Fatalf("%s: array shape changed across the scramble", path)
		}
		for i := range b {
			assertOnlyProseDiffers(t, b[i], a[i], fmt.Sprintf("%s[%d]", path, i), changed)
		}
	default:
		if !reflect.DeepEqual(before, after) {
			t.Fatalf("%s: structured (non-prose) value changed across the scramble: %v -> %v", path, before, after)
		}
	}
}

// TestMutateCaptureIdentityRoundTripIsByteStable proves the decode ->
// marshal pipeline itself is deterministic and idempotent BEFORE any
// scramble/strip mode is layered on top -- so a later guard's observed
// difference can be attributed to the mutation, never to the re-marshal
// step (the plan's own instruction for the "identity" mode).
func TestMutateCaptureIdentityRoundTripIsByteStable(t *testing.T) {
	for _, class := range []string{"match", "move", "borrow", "interprocedural_loan"} {
		t.Run(class, func(t *testing.T) {
			raw := captureFile(t, class+"_diagnose_capture.json")
			first := mutateCaptureOrFatal(t, raw, "identity")
			second := mutateCaptureOrFatal(t, raw, "identity")
			if !bytes.Equal(first, second) {
				t.Fatalf("mutateCapture(raw, \"identity\") is not deterministic across repeated calls")
			}
			twice := mutateCaptureOrFatal(t, first, "identity")
			if !bytes.Equal(first, twice) {
				t.Fatalf("mutateCapture(raw, \"identity\") is not idempotent: re-applying identity to its own output changed it")
			}
		})
	}
}

func decodeGenericOrFatal(t testing.TB, raw []byte) map[string]interface{} {
	t.Helper()
	var doc map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

// ---------------------------------------------------------------------
// Anti-theater guard 1 (D-06-27.1): the prose-scramble test.
//
// runViaStandin drives the REAL driver's Repair cycle end to end, but
// with EVERY subprocess call answered by the stand-in serving mode-
// mutated captures instead of the real binary. It is the shared harness
// both this guard and guard 2 (vocabulary removal) are built on:
//   - diagnoseCapture/reverifyCapture are the checked-in real captures
//     (cmd/schway-repair/testdata/<class>_{diagnose,reverify}_capture.json).
//   - The repair actually applied is always derived from the UNMODIFIED
//     diagnoseCapture's own repairs[] (span/replacement/applicability are
//     never touched by scramble_prose, and this lets the harness compute,
//     independent of any guard, exactly what content the driver's SECOND
//     invocation will be asked to check -- so the reverify capture can be
//     staged at the right content-hash key before the driver ever runs).
// ---------------------------------------------------------------------

func runViaStandin(t testing.TB, standinBinary string, mutatedSource, diagnoseCapture, reverifyCapture []byte, mode string) (exitCode int, outcome Outcome, repaired []byte) {
	t.Helper()

	// Determine, from the REAL (unmutated) diagnose capture, exactly what
	// repair the driver will select and apply -- this is mode-independent
	// for scramble_prose (kind/span/replacement/applicability untouched)
	// and is what lets this harness stage the reverify capture at the
	// correct content-hash key up front.
	decoded := decodeCheckJSON(t, diagnoseCapture)
	repair, _, ok := selectRepair(decoded)
	if !ok {
		t.Fatalf("runViaStandin: baseline diagnose capture carries no driver-eligible repair")
	}
	repairComputationPath := writeTempCopy(t, mutatedSource)
	if err := applyRepair(repairComputationPath, repair); err != nil {
		t.Fatalf("runViaStandin: computing expected repaired bytes: %v", err)
	}
	expectedRepaired, err := os.ReadFile(repairComputationPath)
	if err != nil {
		t.Fatal(err)
	}

	standinDir := t.TempDir()
	writeCaptureForContent(t, standinDir, mutatedSource, mutateCaptureOrFatal(t, diagnoseCapture, mode))
	writeCaptureForContent(t, standinDir, expectedRepaired, mutateCaptureOrFatal(t, reverifyCapture, mode))

	sourcePath := writeTempCopy(t, mutatedSource)
	t.Setenv(standinDirEnv, standinDir)
	exitCode, stdout := runCapturingStdout(t, []string{"--lang=" + standinBinary, "--source=" + sourcePath, "--json"})
	if err := json.Unmarshal(stdout, &outcome); err != nil {
		t.Fatalf("runViaStandin: decoding driver's own --json outcome: %v (raw: %s)", err, stdout)
	}
	repaired, err = os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	return exitCode, outcome, repaired
}

// testSourceClassProseScramble drives one of the three classes that go
// through the driver's own check/repair JSON protocol (match, move,
// borrow) twice through the stand-in -- once serving the real captures
// unmodified (mode "identity"), once serving the SAME captures with every
// message/detail string replaced by lorem ipsum -- and asserts all three
// observables (repaired bytes, exit code, reported outcome) are
// identical. This is "without scraping prose," operationalized.
func testSourceClassProseScramble(t *testing.T, standinBinary string, injector session.Injector, fixture, class string) {
	t.Helper()
	original := phase6Fixture(t, fixture)
	mutated, err := injector.Inject(original)
	if err != nil {
		t.Fatalf("injecting %s defect: %v", class, err)
	}
	diagnoseCapture := captureFile(t, class+"_diagnose_capture.json")
	reverifyCapture := captureFile(t, class+"_reverify_capture.json")

	identityExit, identityOutcome, identityRepaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, "identity")
	scrambledExit, scrambledOutcome, scrambledRepaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, "scramble_prose")

	if identityOutcome.Status != OutcomeRepaired {
		t.Fatalf("%s: identity-mode control run did not even report %q (got %q) -- test harness is broken, not exercising the guard", class, OutcomeRepaired, identityOutcome.Status)
	}
	if scrambledExit != identityExit {
		t.Fatalf("%s: exit codes differ under prose scramble: identity=%d scrambled=%d", class, identityExit, scrambledExit)
	}
	if scrambledOutcome != identityOutcome {
		t.Fatalf("%s: reported outcome differs under prose scramble: identity=%+v scrambled=%+v", class, identityOutcome, scrambledOutcome)
	}
	if !bytes.Equal(scrambledRepaired, identityRepaired) {
		t.Fatalf("%s: repaired bytes differ under prose scramble", class)
	}
}

// testCleanupClassIsProseIndependent covers the cleanup class. Its repair
// mechanism (re-deriving generated C from the untouched .schway source via
// cgen.EmitNative, then re-validating via native execution) NEVER decodes
// any JSON document at all -- there is no message/detail channel for a
// prose scramble to touch, so "prose cannot change this class's repair
// behaviour" is trivially and structurally true. Demonstrated concretely
// by running the mechanism twice and asserting byte-identical output,
// since a mechanism claimed to be prose-independent had better also be
// plain deterministic.
func testCleanupClassIsProseIndependent(t *testing.T) {
	t.Helper()
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.schway"))
	if err != nil {
		t.Fatal(err)
	}
	_, first := cgen.EmitNative(checked.Program)
	_, second := cgen.EmitNative(checked.Program)
	if first == nil || second == nil {
		t.Fatal("cleanup fixture unexpectedly regained native admission after the Phase 16 M004 cut")
	}
	if first.Error() != second.Error() {
		t.Fatalf("cleanup's M004 refusal is not deterministic: first=%q second=%q", first, second)
	}
}

// testStaleEvidenceClassIsProseIndependent covers the stale-evidence
// class. Its repair mechanism (repairStaleEvidenceClass, repair_test.go)
// DOES consume a JSON document (`lang --json evidence --validate`'s own
// response), which legitimately carries a "message" field on failure --
// but the mechanism's own validateStatus closure decodes ONLY {"status"},
// exactly like the real driver's own TestRepairDriverDecodesNoProseFields
// proves for cmd/schway-repair's structs. Asserted here by go/ast over
// repair_test.go itself, so a future edit that starts reading message
// cannot silently reintroduce a prose dependency for this class.
func testStaleEvidenceClassIsProseIndependent(t *testing.T) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "repair_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	ast.Inspect(file, func(n ast.Node) bool {
		fn, isFunc := n.(*ast.FuncDecl)
		if !isFunc || fn.Name.Name != "repairStaleEvidenceClass" {
			return true
		}
		found = true
		ast.Inspect(fn, func(inner ast.Node) bool {
			st, isStruct := inner.(*ast.StructType)
			if !isStruct {
				return true
			}
			for _, field := range st.Fields.List {
				if field.Tag == nil {
					continue
				}
				tag := strings.Trim(field.Tag.Value, "`")
				if strings.Contains(tag, `json:"message`) || strings.Contains(tag, `json:"detail`) {
					t.Fatalf("repairStaleEvidenceClass now decodes a prose field (%s) -- the stale-evidence class must remain structure-only", tag)
				}
			}
			return true
		})
		return true
	})
	if !found {
		t.Fatal("repairStaleEvidenceClass not found in repair_test.go -- this test's target moved or was renamed")
	}
}

// TestProseScrambleLeavesRepairBehaviourIdentical is the first anti-
// theater guard (D-06-27.1): the driver is run against real diagnostics
// and, separately, against the same diagnostics with every message/detail
// string replaced by lorem ipsum -- kind/span/replacement/applicability
// untouched -- and both runs must produce IDENTICAL repair behaviour and
// outcome. Covers all five defect classes (T-06-THEATER-01).
func TestProseScrambleLeavesRepairBehaviourIdentical(t *testing.T) {
	standinBinary := buildStandin(t)

	t.Run("match", func(t *testing.T) {
		testSourceClassProseScramble(t, standinBinary, session.MatchInjector{}, "heldout_match_defect.schway", "match")
	})
	t.Run("move", func(t *testing.T) {
		testSourceClassProseScramble(t, standinBinary, session.MoveInjector{}, "heldout_move_defect.schway", "move")
	})
	t.Run("borrow", func(t *testing.T) {
		testSourceClassProseScramble(t, standinBinary, session.BorrowInjector{}, "heldout_borrow_defect.schway", "borrow")
	})
	t.Run("cleanup", func(t *testing.T) {
		testCleanupClassIsProseIndependent(t)
	})
	t.Run("stale_evidence", func(t *testing.T) {
		testStaleEvidenceClassIsProseIndependent(t)
	})
	// 13-06 Task 3 (13-ANTITHEATER-CONTRACT.md obligation 6, the one
	// MANDATORY registration step): move_after_interprocedural_loan is the
	// first interprocedural class to receive real captures and be
	// registered across all four contract-named lists.
	// wrap_call_in_try/use_matching_argument are DELIBERATELY not
	// registered here -- see 13-06-SUMMARY.md's "Antitheater obligations
	// discharged" section for the recorded decision (contract obligation
	// 2 requires naming a skip explicitly, not silently leaving it
	// uncovered).
	t.Run("interprocedural_loan", func(t *testing.T) {
		testInterproceduralClassProseScramble(t, standinBinary, session.InterproceduralLoanInjector{}, "heldout_shared_callee_twin_alpha.schway", "interprocedural_loan")
	})
}

// testInterproceduralClassProseScramble is testSourceClassProseScramble's
// phase13 sibling: identical mechanism, but reads its held-out fixture
// from testdata/phase13 (via testsupport.ProjectPath) rather than
// testdata/phase6 (phase6Fixture) -- the two corpora are deliberately
// disjoint by directory (D-13-24), so this cannot simply reuse the
// existing helper.
func testInterproceduralClassProseScramble(t *testing.T, standinBinary string, injector session.Injector, fixture, class string) {
	t.Helper()
	original, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase13", fixture))
	if err != nil {
		t.Fatal(err)
	}
	mutated, err := injector.Inject(original)
	if err != nil {
		t.Fatalf("injecting %s defect: %v", class, err)
	}
	diagnoseCapture := captureFile(t, class+"_diagnose_capture.json")
	reverifyCapture := captureFile(t, class+"_reverify_capture.json")

	identityExit, identityOutcome, identityRepaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, "identity")
	scrambledExit, scrambledOutcome, scrambledRepaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, "scramble_prose")

	if identityOutcome.Status != OutcomeRepaired {
		t.Fatalf("%s: identity-mode control run did not even report %q (got %q) -- test harness is broken, not exercising the guard", class, OutcomeRepaired, identityOutcome.Status)
	}
	if scrambledExit != identityExit {
		t.Fatalf("%s: exit codes differ under prose scramble: identity=%d scrambled=%d", class, identityExit, scrambledExit)
	}
	if scrambledOutcome != identityOutcome {
		t.Fatalf("%s: reported outcome differs under prose scramble: identity=%+v scrambled=%+v", class, identityOutcome, scrambledOutcome)
	}
	if !bytes.Equal(scrambledRepaired, identityRepaired) {
		t.Fatalf("%s: repaired bytes differ under prose scramble", class)
	}
}

// ---------------------------------------------------------------------
// Anti-theater guard 2 (D-06-27.2): the structured-vocabulary-removal
// test. Four stripping modes remove exactly one structured channel from
// the diagnose capture -- repairs[] wholesale, or a single field within
// each repair -- while leaving every message/detail value byte-identical
// to the baseline. If the driver genuinely depends only on the structured
// channel, every one of these must drive it RED (a non-zero exit and a
// reported-unrepaired outcome), even though a human reading the prose
// could plainly see what needs fixing.
// ---------------------------------------------------------------------

// assertMessageAndDetailUnaffectedByStrip proves the precondition every
// strip_* subtest below depends on: the stripping mode removed only
// structured vocabulary, leaving the diagnostic's own message and each
// cause's detail exactly as the real binary produced them. Without this,
// a RED result could be coincidental (the strip could have also silently
// mangled the prose) rather than proof that the structured channel alone
// mattered.
func assertMessageAndDetailUnaffectedByStrip(t testing.TB, before, after []byte) {
	t.Helper()
	beforeDoc := decodeGenericOrFatal(t, before)
	afterDoc := decodeGenericOrFatal(t, after)
	beforeDiagnostics, _ := beforeDoc["diagnostics"].([]interface{})
	afterDiagnostics, _ := afterDoc["diagnostics"].([]interface{})
	if len(beforeDiagnostics) != len(afterDiagnostics) {
		t.Fatalf("diagnostics count changed by stripping structured vocabulary: %d -> %d", len(beforeDiagnostics), len(afterDiagnostics))
	}
	for i := range beforeDiagnostics {
		b := beforeDiagnostics[i].(map[string]interface{})
		a := afterDiagnostics[i].(map[string]interface{})
		if b["message"] != a["message"] {
			t.Fatalf("diagnostics[%d].message changed by stripping structured vocabulary: %v -> %v", i, b["message"], a["message"])
		}
		bCauses, _ := b["causes"].([]interface{})
		aCauses, _ := a["causes"].([]interface{})
		if len(bCauses) != len(aCauses) {
			t.Fatalf("diagnostics[%d].causes count changed by stripping structured vocabulary", i)
		}
		for j := range bCauses {
			bc, _ := bCauses[j].(map[string]interface{})
			ac, _ := aCauses[j].(map[string]interface{})
			if bc["detail"] != ac["detail"] {
				t.Fatalf("diagnostics[%d].causes[%d].detail changed by stripping structured vocabulary: %v -> %v", i, j, bc["detail"], ac["detail"])
			}
		}
	}
}

// testVocabularyRemovalMode drives the real driver, via the stand-in,
// against the move class's real captures with mode's structured
// vocabulary stripped from the diagnose capture, and asserts the driver
// goes RED: a non-zero exit, a reported-unrepaired outcome, and the
// source file left completely untouched (the driver never even attempts
// a splice when it selects no repair).
func testVocabularyRemovalMode(t *testing.T, standinBinary string, mode string) {
	t.Helper()
	original := phase6Fixture(t, "heldout_move_defect.schway")
	mutated, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatal(err)
	}
	diagnoseCapture := captureFile(t, "move_diagnose_capture.json")
	reverifyCapture := captureFile(t, "move_reverify_capture.json")

	assertMessageAndDetailUnaffectedByStrip(t, diagnoseCapture, mutateCaptureOrFatal(t, diagnoseCapture, mode))

	exitCode, outcome, repaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, mode)
	if exitCode == 0 {
		t.Fatalf("mode %q: driver exited 0 (success) despite the structured vocabulary being stripped -- prose alone must never be sufficient", mode)
	}
	if outcome.Status != OutcomeUnrepairable {
		t.Fatalf("mode %q: got outcome status %q, want %q", mode, outcome.Status, OutcomeUnrepairable)
	}
	if !bytes.Equal(repaired, mutated) {
		t.Fatalf("mode %q: source file was modified despite the driver reporting %q", mode, OutcomeUnrepairable)
	}
}

// TestVocabularyRemovalDrivesTheDriverRed is the second anti-theater
// guard (D-06-27.2): stripping repairs[]/kind/span/replacement from the
// JSON with prose left fully intact makes the driver go RED. This proves
// the mechanism genuinely depends on the structured channel rather than
// silently falling back to prose it could scrape.
func TestVocabularyRemovalDrivesTheDriverRed(t *testing.T) {
	standinBinary := buildStandin(t)

	t.Run("strip_repairs", func(t *testing.T) { testVocabularyRemovalMode(t, standinBinary, "strip_repairs") })
	t.Run("strip_kind", func(t *testing.T) { testVocabularyRemovalMode(t, standinBinary, "strip_kind") })
	t.Run("strip_span", func(t *testing.T) { testVocabularyRemovalMode(t, standinBinary, "strip_span") })
	t.Run("strip_replacement", func(t *testing.T) { testVocabularyRemovalMode(t, standinBinary, "strip_replacement") })
}

// TestVocabularyRemovalGuardIsNotInert is the load-bearing control: the
// SAME harness (runViaStandin), against the SAME move-class captures,
// with NOTHING stripped ("identity" mode) must go GREEN. Without this
// control, a broken harness that made every run fail for an unrelated
// reason would make all four RED assertions above pass for the wrong
// reason -- a guard that can only ever assert RED cannot detect its own
// inertness any more than one that can only ever assert green.
func TestVocabularyRemovalGuardIsNotInert(t *testing.T) {
	standinBinary := buildStandin(t)
	original := phase6Fixture(t, "heldout_move_defect.schway")
	mutated, err := session.MoveInjector{}.Inject(original)
	if err != nil {
		t.Fatal(err)
	}
	diagnoseCapture := captureFile(t, "move_diagnose_capture.json")
	reverifyCapture := captureFile(t, "move_reverify_capture.json")

	exitCode, outcome, repaired := runViaStandin(t, standinBinary, mutated, diagnoseCapture, reverifyCapture, "identity")
	if exitCode != 0 {
		t.Fatalf("control run (nothing stripped) exited nonzero (%d) -- the harness itself is broken, independent of any stripping mode", exitCode)
	}
	if outcome.Status != OutcomeRepaired {
		t.Fatalf("control run (nothing stripped) got outcome status %q, want %q -- the harness itself is broken", outcome.Status, OutcomeRepaired)
	}
	if bytes.Equal(repaired, mutated) {
		t.Fatal("control run (nothing stripped) left the source unmodified -- the harness itself is broken")
	}
}

// TestProseScrambleFixtureKeepsStructuredFieldsIntact is the scrambler's
// own verification, asserting BOTH directions: at least one prose value
// changed (a scrambler that matched nothing would make the whole guard
// pass vacuously), and every structured value is byte-identical (a
// scrambler that accidentally touched a span would make the guard fail
// for the wrong reason). Both failure modes are refused explicitly.
func TestProseScrambleFixtureKeepsStructuredFieldsIntact(t *testing.T) {
	for _, class := range []string{"match", "move", "borrow", "interprocedural_loan"} {
		for _, kind := range []string{"diagnose", "reverify"} {
			name := class + "_" + kind
			t.Run(name, func(t *testing.T) {
				raw := captureFile(t, class+"_"+kind+"_capture.json")
				before := decodeGenericOrFatal(t, mutateCaptureOrFatal(t, raw, "identity"))
				after := decodeGenericOrFatal(t, mutateCaptureOrFatal(t, raw, "scramble_prose"))
				changed := false
				assertOnlyProseDiffers(t, before, after, name, &changed)
				if kind == "diagnose" && !changed {
					t.Fatalf("%s: scramble_prose changed zero prose values -- the scrambler matched nothing, which would make the guard pass vacuously", name)
				}
			})
		}
	}
}
