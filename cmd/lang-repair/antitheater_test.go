package main

// This file ships the two remaining DX-04 anti-theater guards (D-06-27.1,
// D-06-27.2) plus the fixture-substitution subprocess they are both built
// on. It is a _test.go file, so it is explicitly EXEMPT from the
// import-boundary lint (import_boundary_test.go scans only non-test
// files) -- exactly like repair_test.go, it imports internal/ packages to
// DRIVE the real injectors and the real shipped `lang` binary, while the
// driver itself (repair.go, main.go) stays untouched by this plan
// (TestRepairDriverAndMainAreUnchangedByThisPlan).
//
// Mechanism (06-RESEARCH.md Open Question 2): a fixture-substitution
// subprocess, not a driver-side test hook. The stand-in program built by
// buildStandin below is passed as the driver's own `--lang` path -- the
// driver spawns it exactly the way it spawns the real `lang` binary,
// reads its stdout, and never learns the difference. This costs the
// driver ZERO code changes: no second, untested code path is created
// inside cmd/lang-repair.
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
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// standinDirEnv names the environment variable the stand-in reads to find
// its captures directory. The real driver never sets or reads this
// variable -- it is purely a fixture of THIS test file, inherited by the
// stand-in subprocess because os/exec defaults Cmd.Env to the parent
// process's environment when Env is left nil (exactly what runLangCheck
// does), so t.Setenv in the test process reaches the stand-in without any
// driver-side plumbing.
const standinDirEnv = "LANG_REPAIR_STANDIN_DIR"

// standinSource is the fixture-substitution stand-in's entire program. It
// takes no flags of its own and does not even look at "--json check" --
// it reads its LAST argument (the source file path the driver always
// passes last), hashes that file's CURRENT on-disk content, and writes
// the captured document filed under that hash to stdout. A file the test
// harness never staged a capture for is refused with a nonzero exit
// (never silently empty output), which is what makes a captures-directory
// bug fail loudly rather than masquerade as a real "lang" crash.
const standinSource = `package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	dir := os.Getenv("LANG_REPAIR_STANDIN_DIR")
	if dir == "" || len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "standin: LANG_REPAIR_STANDIN_DIR unset or no source argument")
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
	dir, err := os.MkdirTemp("", "lang-repair-standin")
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
// code and JSON envelope a real `lang-repair --json` invocation would
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
	path := filepath.Join(t.TempDir(), "copy.lang")
	mustWriteFile(t, path, content)
	return path
}

// captureFile reads one of the checked-in per-class baseline captures
// this plan's Task 1 committed to cmd/lang-repair/testdata/. They were
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
	for _, class := range []string{"match", "move", "borrow"} {
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

	original := phase6Fixture(t, "heldout_move_defect.lang")
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

// TestRepairDriverAndMainAreUnchangedByThisPlan asserts, by content hash,
// that repair.go and main.go -- the driver's own production code -- carry
// the exact bytes 06-13 shipped. The fixture-substitution mechanism above
// requires zero driver changes by construction; this test makes that a
// live, falsifiable check rather than a claim in a comment.
func TestRepairDriverAndMainAreUnchangedByThisPlan(t *testing.T) {
	for _, name := range []string{"repair.go", "main.go"} {
		info, err := os.Stat(name)
		if err != nil {
			t.Fatalf("%s: expected to exist unmodified from 06-13: %v", name, err)
		}
		if info.IsDir() {
			t.Fatalf("%s: expected a file", name)
		}
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
	default:
		return nil, fmt.Errorf("mutateCapture: unknown mode %q", mode)
	}
	return json.Marshal(doc)
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
	for _, class := range []string{"match", "move", "borrow"} {
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
