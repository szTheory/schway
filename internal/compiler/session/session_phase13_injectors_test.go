package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase13Fixture reads a fixture from testdata/phase13 by name, mirroring
// phase6Fixture's shape exactly (session_phase6_injectors_test.go).
func phase13Fixture(t *testing.T, name string) []byte {
	t.Helper()
	path := testsupport.ProjectPath("testdata", "phase13", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %s: %v", name, err)
	}
	return source
}

// The three inline base programs below are Task 1's own self-contained
// fixtures -- Task 1 lands BEFORE Task 2 authors the real testdata/phase13
// corpus, so TestPhase13InjectorsProduceExactlyOneDefect and
// TestPhase13InjectorGuardsAreNotInert (Task 1's own required tests) must
// not depend on files Task 2 has not written yet. Each mirrors the SHAPE
// of its corresponding testdata/phase13 corpus member Task 2 later adds,
// so both find and exercise the identical defect mechanism.

var phase13InjectorLoanBase = []byte(`module phase13.injector_base_loan

export {
  fn relay
}

fn sink(buffer: Buffer) -> Buffer {
  let held = take buffer
  held
}

fn relay(buffer: Buffer) -> Buffer {
  let borrowed = borrow buffer
  let routed = sink(borrowed)
  let delivered = take buffer // lang:interprocedural-loan-target
  routed
}`)

var phase13InjectorConsumeBase = []byte(`module phase13.injector_base_consume

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request) // lang:fallible-consume-target
  handle
}`)

var phase13InjectorArgumentBase = []byte(`module phase13.injector_base_argument

export {
  fn main
}

fn identity(value: Byte) -> Byte { // lang:call-argument-target
  value
}

fn main(value: Byte) -> Byte {
  let result = identity(value)
  result
}`)

// TestPhase13InjectorsProduceExactlyOneDefect is Task 1's central
// assertion: each of the three phase-13 injectors, applied to its own
// clean base, produces a mutated source that checks with EXACTLY ONE
// diagnostic, of exactly the code that injector's defect class names.
func TestPhase13InjectorsProduceExactlyOneDefect(t *testing.T) {
	cases := []struct {
		name     string
		injector Injector
		base     []byte
		wantCode string
	}{
		{"interprocedural_loan", InterproceduralLoanInjector{}, phase13InjectorLoanBase, "check.interprocedural_loan_liveness"},
		{"fallible_consume", FallibleConsumeInjector{}, phase13InjectorConsumeBase, "syntax.fallible_call_not_consumed"},
		{"call_argument_type", CallArgumentTypeInjector{}, phase13InjectorArgumentBase, "check.call_argument_type_mismatch"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			baseResult := Check(tc.base)
			if len(baseResult.Diagnostics) != 0 {
				t.Fatalf("%s: base did not check clean: %+v", tc.name, baseResult.Diagnostics)
			}
			mutated, err := tc.injector.Inject(tc.base)
			if err != nil {
				t.Fatalf("%s: Inject failed: %v", tc.name, err)
			}
			result := Check(mutated)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("%s: got %d diagnostics, want 1: %+v", tc.name, len(result.Diagnostics), result.Diagnostics)
			}
			if result.Diagnostics[0].Code != tc.wantCode {
				t.Fatalf("%s: got code %q, want %q", tc.name, result.Diagnostics[0].Code, tc.wantCode)
			}
		})
	}
}

// TestPhase13InjectorGuardsAreNotInert is Task 1's D-13-30(a) assertion:
// each injector's guard-disabled twin, given marker-free source, returns
// it completely unchanged (which therefore still checks clean) -- the
// silent pass its real, guarded Inject refuses instead.
func TestPhase13InjectorGuardsAreNotInert(t *testing.T) {
	cases := []struct {
		name      string
		marker    string
		base      []byte
		guardSkip func([]byte) []byte
		injector  Injector
	}{
		{"interprocedural_loan", loanTargetMarker, phase13InjectorLoanBase, interproceduralLoanInjectSkippingGuard, InterproceduralLoanInjector{}},
		{"fallible_consume", fallibleConsumeTargetMarker, phase13InjectorConsumeBase, fallibleConsumeInjectSkippingGuard, FallibleConsumeInjector{}},
		{"call_argument_type", callArgumentTargetMarker, phase13InjectorArgumentBase, callArgumentTypeInjectSkippingGuard, CallArgumentTypeInjector{}},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			markerFree := bytes.ReplaceAll(tc.base, []byte(tc.marker), []byte(""))
			unguarded := tc.guardSkip(markerFree)
			if !bytes.Equal(unguarded, markerFree) {
				t.Fatalf("%s: guard-disabled twin mutated marker-free source; want unchanged", tc.name)
			}
			uncheckedResult := Check(unguarded)
			if len(uncheckedResult.Diagnostics) != 0 {
				t.Fatalf("%s: guard-disabled path's unmutated source did not check clean: %+v", tc.name, uncheckedResult.Diagnostics)
			}
			_, err := tc.injector.Inject(markerFree)
			typed := injectorError(err)
			if typed == nil || typed.Code != InjectorTargetMissingCode {
				t.Fatalf("%s: guarded Inject did not refuse on marker-free input: %v", tc.name, err)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Task 2 -- the corpus, including D-13-28's twin pair, and the
// topology-disjointness control (D-13-25, D-13-26).
// -----------------------------------------------------------------------

// corpusTopology is the (function count, call-edge count,
// detection-to-fix hop distance) triple D-13-26.1 requires TestCorpus
// TopologyDisjoint to compute independently over ast.Program -- re-derived
// here, never read from the callgraph package (the planner's own
// discretion note, 13-04-PLAN.md's objective). Hop distance is defined as
// the shortest call-graph path, in CALLER -> CALLEE edges, from the
// program's root/entry function to the function containing the fixture's
// own marked (detection) statement -- the axis this plan actually varies
// between held-out and derivation members: a shallow chain (derivation,
// hop <= 1) versus a chain routed through one or more pure relay
// functions before reaching the defect (held-out, hop >= 2). This is a
// call-graph-DEPTH reading of "detection-to-fix hop distance", chosen
// because D-13-28's twin-pair fixtures deliberately keep detection and
// the eventual repair-blamed function DIFFERENT identities depending on
// direction (D-13-02a is still open) -- the axis every fixture in this
// corpus genuinely varies, and the one a blame rule keyed on call-graph
// shape could actually overfit, is how deep the defect sits in the call
// graph, not which specific function a not-yet-written repair rule will
// eventually point at.
type corpusTopology struct {
	functionCount int
	edgeCount     int
	hopDistance   int
}

// callEdges walks every function's own linear body bindings and returns
// the caller-name -> callee-name adjacency list, counting every "call",
// "try_call", and "discard_call" RHS kind as one edge -- every Lang-to-Lang
// or Lang-to-foreign call site this language's grammar can express
// (internal/compiler/ast/ast.go's own RHS.Kind doc comment).
func callEdges(t *testing.T, source []byte) (graph map[string][]string, edgeCount int) {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("callEdges: parse failed: %+v", parsed.Diagnostics)
	}
	graph = map[string][]string{}
	for _, function := range parsed.Program.Funcs {
		if function.Body.Linear == nil {
			continue
		}
		for _, binding := range function.Body.Linear.Bindings {
			switch binding.RHS.Kind {
			case "call", "try_call", "discard_call":
				graph[function.Name] = append(graph[function.Name], binding.RHS.Callee)
				edgeCount++
			}
		}
	}
	return graph, edgeCount
}

// bfsHopDistance returns the shortest number of edges from root to target
// in graph, walking CALLER -> CALLEE edges forward.
func bfsHopDistance(graph map[string][]string, root, target string) (int, bool) {
	if root == target {
		return 0, true
	}
	visited := map[string]bool{root: true}
	queue := []string{root}
	dist := map[string]int{root: 0}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range graph[current] {
			if visited[next] {
				continue
			}
			visited[next] = true
			dist[next] = dist[current] + 1
			if next == target {
				return dist[next], true
			}
			queue = append(queue, next)
		}
	}
	return 0, false
}

// computeCorpusTopology parses source independently (never consulting the
// callgraph package) and returns its (function count, call-edge count,
// hop distance) triple, where hop distance is the shortest root ->
// detection path over the re-derived adjacency list.
func computeCorpusTopology(t *testing.T, source []byte, root, detection string) corpusTopology {
	t.Helper()
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("computeCorpusTopology: parse failed: %+v", parsed.Diagnostics)
	}
	graph, edgeCount := callEdges(t, source)
	hop, ok := bfsHopDistance(graph, root, detection)
	if !ok {
		t.Fatalf("computeCorpusTopology: detection function %q unreachable from root %q", detection, root)
	}
	return corpusTopology{functionCount: len(parsed.Program.Funcs), edgeCount: edgeCount, hopDistance: hop}
}

// TestCorpusTopologyDisjoint is D-13-26.1/D-13-26.2's central assertion:
// for each interprocedural class with both a held-out and a derivation
// member, the two members' independently re-derived topology triples are
// NOT equal, and the held-out member's own hop distance is >= 2 -- this
// replaces byte-inequality, which alpha-renaming defeats (D-13-24, the
// M001 lesson).
func TestCorpusTopologyDisjoint(t *testing.T) {
	cases := []struct {
		name             string
		heldoutFile      string
		heldoutRoot      string
		heldoutDetection string
		derivationFile   string
		derivationRoot   string
		derivationDetect string
	}{
		{
			name: "fallible_call_unconsumed",
			heldoutFile: "heldout_fallible_call_unconsumed.lang", heldoutRoot: "main", heldoutDetection: "acquire",
			derivationFile: "derivation_fallible_call_unconsumed.lang", derivationRoot: "main", derivationDetect: "main",
		},
		{
			name: "call_argument_type_mismatch",
			heldoutFile: "heldout_call_argument_mismatch.lang", heldoutRoot: "main", heldoutDetection: "dispatch",
			derivationFile: "derivation_call_argument_mismatch.lang", derivationRoot: "main", derivationDetect: "main",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			heldout := computeCorpusTopology(t, phase13Fixture(t, tc.heldoutFile), tc.heldoutRoot, tc.heldoutDetection)
			derivation := computeCorpusTopology(t, phase13Fixture(t, tc.derivationFile), tc.derivationRoot, tc.derivationDetect)
			if heldout == derivation {
				t.Fatalf("%s: held-out and derivation topology triples are EQUAL (%+v) -- the split varies nothing a blame rule could overfit", tc.name, heldout)
			}
			if heldout.hopDistance < 2 {
				t.Fatalf("%s: held-out hop distance %d, want >= 2 (D-13-26.2)", tc.name, heldout.hopDistance)
			}
			if derivation.hopDistance > 1 {
				t.Fatalf("%s: derivation hop distance %d, want <= 1", tc.name, derivation.hopDistance)
			}
		})
	}
}

// alphaRenameDerivationCallArgumentMismatch returns a byte-for-byte
// structural copy of derivation_call_argument_mismatch.lang with every
// identifier (module suffix, function names, parameter/binding names)
// renamed -- exactly M001's documented weakness shape (item -> buffer,
// moved_once -> delivered): a rename that changes no call-graph shape at
// all, so its topology triple must come out IDENTICAL to the original.
func alphaRenameDerivationCallArgumentMismatch(t *testing.T, source []byte) []byte {
	t.Helper()
	renamed := string(source)
	replacements := []struct{ from, to string }{
		{"derivation_call_argument_mismatch", "alpha_renamed_derivation_call_argument_mismatch"},
		{"identity", "principal"},
		{"result", "outcome"},
		{"value", "payload"},
		{"main", "entry"},
	}
	for _, replacement := range replacements {
		pattern := regexp.MustCompile(`\b` + regexp.QuoteMeta(replacement.from) + `\b`)
		renamed = pattern.ReplaceAllString(renamed, replacement.to)
	}
	return []byte(renamed)
}

// TestCorpusTopologyGuardIsNotInert is D-13-30(b)'s required mutation
// kill -- the topology-distinctness control is the one most in need of a
// not-inert proof, since it is what proves the split is not ceremonial.
// An alpha-renamed copy of a derivation fixture, fed through the SAME
// topology computation TestCorpusTopologyDisjoint uses, must come out
// topologically EQUAL to the original -- proving that had this renamed
// copy been submitted in place of a real held-out member, TestCorpus
// TopologyDisjoint's own `heldout == derivation` fatal would have fired.
// A control that a pure rename can slip past is exactly the M001
// weakness this phase exists to close.
func TestCorpusTopologyGuardIsNotInert(t *testing.T) {
	original := phase13Fixture(t, "derivation_call_argument_mismatch.lang")
	dir := t.TempDir()
	renamed := alphaRenameDerivationCallArgumentMismatch(t, original)
	path := filepath.Join(dir, "alpha_renamed_derivation_call_argument_mismatch.lang")
	if err := os.WriteFile(path, renamed, 0o644); err != nil {
		t.Fatal(err)
	}
	renamedSource, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	renamedResult := Check(renamedSource)
	if len(renamedResult.Diagnostics) != 0 {
		t.Fatalf("alpha-renamed copy did not check clean: %+v", renamedResult.Diagnostics)
	}
	originalTriple := computeCorpusTopology(t, original, "main", "main")
	renamedTriple := computeCorpusTopology(t, renamedSource, "entry", "entry")
	if renamedTriple != originalTriple {
		t.Fatalf("control is INERT in the wrong direction: alpha-renaming changed the topology triple (original=%+v, renamed=%+v) -- rename must preserve topology for this test to demonstrate anything", originalTriple, renamedTriple)
	}
	// This IS the violation: had the renamed copy been submitted as the
	// held-out member, TestCorpusTopologyDisjoint's `heldout == derivation`
	// fatal above would have fired on exactly this pair. Demonstrating
	// that the two triples are equal -- already asserted above via
	// renamedTriple != originalTriple -- is D-13-30(b)'s not-inert proof:
	// a pure identifier rename cannot escape the disjointness check.
}

// TestHeldoutBaselinesAreFailClosed is D-13-26.3's central assertion:
// every heldout_* fixture checks clean unmutated, and produces exactly
// one diagnostic, of the expected code, once its injector runs. This
// plan does NOT yet assert the presence of a driver-eligible repair for
// classes 2 and 3 -- those repair rules land in plan 13-05, and plan
// 13-06 adds that assertion once they exist.
func TestHeldoutBaselinesAreFailClosed(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		injector Injector
		wantCode string
	}{
		{"fallible_call_unconsumed", "heldout_fallible_call_unconsumed.lang", FallibleConsumeInjector{}, "syntax.fallible_call_not_consumed"},
		{"call_argument_mismatch", "heldout_call_argument_mismatch.lang", CallArgumentTypeInjector{}, "check.call_argument_type_mismatch"},
		{"call_argument_ambiguous", "heldout_call_argument_ambiguous.lang", CallArgumentTypeInjector{}, "check.call_argument_type_mismatch"},
		{"shared_callee_twin_alpha", "heldout_shared_callee_twin_alpha.lang", InterproceduralLoanInjector{}, "check.interprocedural_loan_liveness"},
		{"shared_callee_twin_mirror", "heldout_shared_callee_twin_mirror.lang", InterproceduralLoanInjector{}, "check.interprocedural_loan_liveness"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			base := phase13Fixture(t, tc.file)
			baseResult := Check(base)
			if len(baseResult.Diagnostics) != 0 {
				t.Fatalf("%s: base did not check clean: %+v", tc.file, baseResult.Diagnostics)
			}
			mutated, err := tc.injector.Inject(base)
			if err != nil {
				t.Fatalf("%s: Inject failed: %v", tc.file, err)
			}
			result := Check(mutated)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("%s: got %d diagnostics, want 1: %+v", tc.file, len(result.Diagnostics), result.Diagnostics)
			}
			if result.Diagnostics[0].Code != tc.wantCode {
				t.Fatalf("%s: got code %q, want %q", tc.file, result.Diagnostics[0].Code, tc.wantCode)
			}
		})
	}
}

// -----------------------------------------------------------------------
// Task 3 -- sealing the held-out corpus (D-13-27) and proving the seal
// is not inert (D-13-30(c)).
// -----------------------------------------------------------------------

// heldoutManifestPath is testdata/phase13/HELDOUT.sha256's own path.
func heldoutManifestPath() string {
	return testsupport.ProjectPath("testdata", "phase13", "HELDOUT.sha256")
}

// parseHeldoutManifest reads and parses a sha256sum-shaped manifest file
// (one "<hex digest>  <path>" line per entry) into an ordered slice.
func parseHeldoutManifest(t *testing.T, path string) []struct{ digest, relPath string } {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read manifest %s: %v", path, err)
	}
	var entries []struct{ digest, relPath string }
	for _, line := range strings.Split(strings.TrimRight(string(raw), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.SplitN(line, "  ", 2)
		if len(fields) != 2 {
			t.Fatalf("manifest line does not match '<digest>  <path>': %q", line)
		}
		entries = append(entries, struct{ digest, relPath string }{digest: fields[0], relPath: fields[1]})
	}
	return entries
}

// TestHeldoutCorpusSealed is D-13-27's central assertion: the manifest
// lists exactly the heldout_*.lang files present under testdata/phase13,
// and every listed digest matches the file's current bytes. No
// regenerate helper, bless flag, or environment-variable escape exists
// anywhere in this path -- editing a held-out fixture after this test is
// written is a loud, reviewable diff, never a silent pass.
func TestHeldoutCorpusSealed(t *testing.T) {
	repoRoot := testsupport.ProjectPath()
	manifestPath := heldoutManifestPath()
	entries := parseHeldoutManifest(t, manifestPath)

	listed := map[string]bool{}
	for _, entry := range entries {
		listed[entry.relPath] = true
		fullPath := filepath.Join(repoRoot, entry.relPath)
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("manifest lists %s but it does not exist: %v", entry.relPath, err)
		}
		sum := sha256.Sum256(data)
		got := hex.EncodeToString(sum[:])
		if got != entry.digest {
			t.Fatalf("%s: digest mismatch -- manifest says %s, current bytes hash to %s (a sealed held-out fixture was edited)", entry.relPath, entry.digest, got)
		}
	}

	corpusDir := testsupport.ProjectPath("testdata", "phase13")
	dirEntries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", corpusDir, err)
	}
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		if !strings.HasPrefix(name, "heldout_") || !strings.HasSuffix(name, ".lang") {
			continue
		}
		relPath := "testdata/phase13/" + name
		if !listed[relPath] {
			t.Fatalf("%s exists on disk but is not listed in HELDOUT.sha256 -- the manifest must list every held-out fixture", relPath)
		}
	}
}

// TestHeldoutCorpusSealGuardIsNotInert is D-13-30(c)'s required mutation
// kill: flipping exactly one byte in a TEMP COPY of a sealed fixture must
// make the same digest comparison report a mismatch. The real fixture is
// never touched.
func TestHeldoutCorpusSealGuardIsNotInert(t *testing.T) {
	entries := parseHeldoutManifest(t, heldoutManifestPath())
	if len(entries) == 0 {
		t.Fatal("manifest is empty -- nothing to prove the seal guard against")
	}
	repoRoot := testsupport.ProjectPath()
	target := entries[0]
	original, err := os.ReadFile(filepath.Join(repoRoot, target.relPath))
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	tampered := append([]byte(nil), original...)
	flipIndex := len(tampered) / 2
	tampered[flipIndex] ^= 0xFF // exactly one byte flipped
	tamperedPath := filepath.Join(dir, filepath.Base(target.relPath))
	if err := os.WriteFile(tamperedPath, tampered, 0o644); err != nil {
		t.Fatal(err)
	}

	tamperedBytes, err := os.ReadFile(tamperedPath)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(tamperedBytes)
	got := hex.EncodeToString(sum[:])
	if got == target.digest {
		t.Fatalf("one-byte flip did not change the digest -- seal guard is inert (original=%s, tampered=%s)", target.digest, got)
	}

	// The real fixture on disk must be untouched.
	untouched, err := os.ReadFile(filepath.Join(repoRoot, target.relPath))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(untouched, original) {
		t.Fatal("the real sealed fixture was mutated -- this test must only ever touch a temp copy")
	}
}

// phase13CorpusDispatchMarkers are the marker filenames cmd/lang/main.go's
// existing isPhase5Corpus/isPhase6Corpus/isPhase7Corpus dispatch functions
// key on -- a testdata/phase13 file sharing any of these names would cause
// the directory to be misrouted as an earlier phase's corpus by any tool
// that walks testdata/ looking for these markers (testdata/phase13/README).
var phase13CorpusDispatchMarkers = []string{
	"heldout_match_defect.lang", // isPhase6Corpus
	"call_basic.lang",           // isPhase7Corpus
	"restrict_borrow.lang",      // isPhase5Corpus
}

// TestPhase13CorpusIsNotMisroutedByCorpusDispatch asserts that no file
// under testdata/phase13 carries a name cmd/lang/main.go's existing
// marker-file corpus dispatch keys on -- the planner's chosen alternative
// to adding a fourth isPhaseNCorpus sibling (13-04-PLAN.md's objective):
// this corpus is consumed by Go tests and by per-file `lang --json check`,
// never by a corpus-level `lang verify`.
func TestPhase13CorpusIsNotMisroutedByCorpusDispatch(t *testing.T) {
	corpusDir := testsupport.ProjectPath("testdata", "phase13")
	dirEntries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("failed to read %s: %v", corpusDir, err)
	}
	present := map[string]bool{}
	for _, dirEntry := range dirEntries {
		present[dirEntry.Name()] = true
	}
	for _, marker := range phase13CorpusDispatchMarkers {
		if present[marker] {
			t.Fatalf("testdata/phase13 contains %q, which cmd/lang/main.go's existing corpus dispatch would use to misroute this directory", marker)
		}
	}
}
