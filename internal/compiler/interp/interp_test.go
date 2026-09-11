package interp

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// interpProjectRoot mirrors nat03ProjectRoot's own technique
// (internal/compiler/session/session_phase5_alias.go): a
// runtime.Caller(0)-anchored resolution, avoiding a testsupport import
// (testsupport pulls in session, and session imports interp -- an interp
// test importing testsupport would be a cycle).
func interpProjectRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

// checkedCallBasicProgram checks and corevalidates testdata/phase07/call_basic.lang
// directly (syntax.Parse + check.Program + corevalidate.Validate), never
// through package session -- session imports interp, so an interp test
// importing session would be a cycle.
func checkedCallBasicProgram(t *testing.T) core.Program {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", "call_basic.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse: unexpected diagnostics: %v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check: unexpected diagnostics: %v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate rejected: %v", validated.Problems)
	}
	return validated.Program()
}

// moveAsCopyProbeProgram hand-builds a minimal two-function core.Program
// (never fed through the parser) whose caller re-reads its OWN call
// argument place immediately after the call. No LEGAL Lang source can
// express this: a non-copyable argument used a second time after a call is
// exactly what check's and corevalidate's OWN independent move-tracking
// both refuse (D-07-11's consume rule, re-derived independently on each
// side) -- so this probe is deliberately synthetic, mirroring the
// established in-repo precedent for exercising an engine directly against
// a forged core.Program (see callgraph's own T-07-35 doc comment). It is
// run via runFrameStack directly, bypassing Run's own corevalidate.Validate
// gate, which would otherwise refuse this shape identically regardless of
// moveAsCopyForTest -- exactly the point: the probe isolates interp's OWN
// mechanism from the two peers this test proves are blind to it.
func moveAsCopyProbeProgram() (program core.Program, caller core.Function) {
	nonCopyType := core.TypeFact{ID: "probe:type:buffer"} // no AbilityCopy: non-copyable (D-07-11)

	calleeParam := core.Parameter{ID: "probe:callee:place:param", Name: "value", Type: "Buffer"}
	callee := core.Function{
		ID: "probe:callee:fn", Name: "identity", Parameter: calleeParam, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:     "probe:callee:linear",
			Types:  []core.TypeFact{nonCopyType},
			Places: []core.Place{{ID: calleeParam.ID, Name: "value", TypeID: nonCopyType.ID}},
			Operations: []core.LinearOperation{
				{ID: "probe:callee:op:0", Kind: core.OpReturn, SourceID: calleeParam.ID, TypeID: nonCopyType.ID},
			},
		},
	}

	callerParam := core.Parameter{ID: "probe:caller:place:param", Name: "buffer", Type: "Buffer"}
	callTarget := "probe:caller:place:1"
	staleReadTarget := "probe:caller:place:2"
	caller = core.Function{
		ID: "probe:caller:fn", Name: "main", Parameter: callerParam, ReturnType: "Buffer",
		Linear: &core.LinearBody{
			ID:    "probe:caller:linear",
			Types: []core.TypeFact{nonCopyType},
			Places: []core.Place{
				{ID: callerParam.ID, Name: "buffer", TypeID: nonCopyType.ID},
				{ID: callTarget, Name: "result", TypeID: nonCopyType.ID},
				{ID: staleReadTarget, Name: "stale", TypeID: nonCopyType.ID},
			},
			Operations: []core.LinearOperation{
				// The call MOVES buffer (non-copyable): its own SourceID is
				// the caller's argument place.
				{ID: "probe:caller:op:0", Kind: core.OpCall, SourceID: callerParam.ID, TargetID: callTarget, TypeID: nonCopyType.ID, CalleeID: callee.ID},
				// The stale re-read: legal ONLY if the caller-side delete
				// was skipped (moveAsCopyForTest).
				{ID: "probe:caller:op:1", Kind: core.OpCopy, SourceID: callerParam.ID, TargetID: staleReadTarget, TypeID: nonCopyType.ID},
				{ID: "probe:caller:op:2", Kind: core.OpReturn, SourceID: staleReadTarget, TypeID: nonCopyType.ID},
			},
		},
	}

	return core.Program{Schema: core.Schema1, Module: "probe.move_as_copy", Functions: []core.Function{callee, caller}}, caller
}

// TestMoveAsCopyMutationKilled is Task 2's D-10-41 mutation test (QLT-08):
// partitionFrameForCall's caller-side delete of a non-copyable argument
// place is the entire mechanism realizing OWN-05b's ownership-transfer fact
// at a call boundary. With that delete suppressed via moveAsCopyForTest
// (move silently becomes copy), the caller's moved-from place stays
// readable after the call -- and ONLY interp's own observable execution
// behavior catches this: check and corevalidate execute nothing, so they
// are structurally blind to it (D-10-36). That pairing -- one peer catches
// what the other two cannot even in principle -- is what makes "interp is
// independent" falsifiable rather than rhetorical, and it is this phase's
// mutation obligation for the ownership fact (D-10-41).
//
// Four-beat body (pathoracle_test.go's precedent): assert clean,
// save/override/defer-restore, assert an observable effect, assert the
// specific divergence.
func TestMoveAsCopyMutationKilled(t *testing.T) {
	program, caller := moveAsCopyProbeProgram()

	// Beat 1: assert clean. The unmutated interp deletes the moved-from
	// argument place, so the stale re-read finds an uninitialized place
	// and the run correctly refuses.
	base := newFlatFrame(caller, map[string]string{caller.Parameter.ID: "AB"})
	cleanExecution, cleanErr := runFrameStack(program, base)
	if cleanErr == nil {
		t.Fatalf("expected the clean run to refuse the stale re-read of a moved-from place, got success: %+v", cleanExecution)
	}
	cleanBytes, err := CanonicalBytes(cleanExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (clean): %v", err)
	}

	// Beat 2: override, restored via defer.
	previous := moveAsCopyForTest
	moveAsCopyForTest = true
	defer func() { moveAsCopyForTest = previous }()

	// Beat 3: assert an observable effect. The mutated run no longer
	// refuses -- it wrongly succeeds, reading the stale value back out,
	// exactly the OWN-05b violation this mutant proves only interp itself
	// can catch.
	mutatedBase := newFlatFrame(caller, map[string]string{caller.Parameter.ID: "AB"})
	mutatedExecution, mutatedErr := runFrameStack(program, mutatedBase)
	if mutatedErr != nil {
		t.Fatalf("expected the mutated run to succeed (skipping the caller-side delete), got error: %v", mutatedErr)
	}
	if mutatedExecution.Outcome.Value != "AB" {
		t.Fatalf("expected the mutated run's stale re-read to return the original argument value %q, got %q", "AB", mutatedExecution.Outcome.Value)
	}
	mutatedBytes, err := CanonicalBytes(mutatedExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (mutated): %v", err)
	}

	// Beat 4: assert the specific divergence -- the clean run's refusal
	// and the mutated run's wrongful success produce different canonical
	// bytes, never merely "some error occurred".
	if bytes.Equal(cleanBytes, mutatedBytes) {
		t.Fatalf("expected the move-as-copy mutation to change interp's canonical bytes, but clean and mutated runs matched: %s", cleanBytes)
	}

	// check and corevalidate execute nothing (D-10-36): re-running BOTH on
	// a real, independently checked program (call_basic.lang) while
	// moveAsCopyForTest is still engaged from Beat 2 produces the
	// IDENTICAL verdict either way, because neither ever consults an
	// interp-internal, unexported runtime seam. This is the pairing the
	// mutant's whole point rests on -- check and corevalidate are
	// structurally blind to interp's own runtime behavior, never merely
	// coincidentally in agreement with it.
	realParsed := syntax.Parse(mustReadFixture(t, "call_basic.lang"))
	realChecked := check.Program(realParsed.Program)
	if len(realChecked.Diagnostics) != 0 {
		t.Fatalf("expected check.Program's diagnostics to be unaffected by moveAsCopyForTest, got: %v", realChecked.Diagnostics)
	}
	realValidated := corevalidate.Validate(realChecked.Program)
	if !realValidated.Valid {
		t.Fatalf("expected corevalidate.Validate's verdict to be unaffected by moveAsCopyForTest, got: %v", realValidated.Problems)
	}
}

// mustReadFixture reads a named file under testdata/phase07.
func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join(interpProjectRoot(), "testdata", "phase07", name)
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return source
}

// TestCallExecutesAcrossOneFrame is Task 1's tracer test (SEM-08, OWN-05b,
// D-10-21/D-10-26): call_basic.lang's main calls identity across a real
// heap frame boundary and gets back identity's own returned value, with at
// least one emitted event attributed to the callee's own function ID
// (D-10-32) -- proof the callee frame genuinely ran rather than being
// faked by a pass-through. The same program run twice must produce
// byte-identical CanonicalBytes (D-10-26): every ordered output the frame
// stack produces comes from an ordered slice, never a Go map range.
func TestCallExecutesAcrossOneFrame(t *testing.T) {
	program := checkedCallBasicProgram(t)

	var calleeID string
	for _, function := range program.Functions {
		if function.Name == "identity" {
			calleeID = function.ID
		}
	}
	if calleeID == "" {
		t.Fatalf("call_basic.lang's checked program has no function named %q", "identity")
	}

	execution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	if execution.Outcome.Kind != "returned" {
		t.Fatalf("expected outcome kind %q, got %q", "returned", execution.Outcome.Kind)
	}
	if execution.Outcome.Value != "7" {
		t.Fatalf("expected the callee's own returned value %q, got %q", "7", execution.Outcome.Value)
	}

	foundCalleeEvent := false
	for _, event := range execution.Events {
		if event.FunctionID == calleeID {
			foundCalleeEvent = true
			break
		}
	}
	if !foundCalleeEvent {
		t.Fatalf("expected at least one event attributed to the callee's own function ID %q; events: %+v", calleeID, execution.Events)
	}

	firstBytes, err := CanonicalBytes(execution)
	if err != nil {
		t.Fatalf("CanonicalBytes: %v", err)
	}
	secondExecution, err := Run(program, "main", "7")
	if err != nil {
		t.Fatalf("second Run(main, %q) returned an unexpected error: %v", "7", err)
	}
	secondBytes, err := CanonicalBytes(secondExecution)
	if err != nil {
		t.Fatalf("CanonicalBytes (second run): %v", err)
	}
	if !bytes.Equal(firstBytes, secondBytes) {
		t.Fatalf("expected two runs of the same program to produce byte-identical CanonicalBytes, got:\n%s\nvs\n%s", firstBytes, secondBytes)
	}
}

// forbiddenOwnershipAccessors is corevalidate.Result's own ownership-bearing
// accessor set (D-10-36/D-10-37): PeerSignatures and PeerSiteCoverage both
// expose the summary peer's own call-site derivation contract. Validate,
// Valid, Problems, and Program stay permitted -- they are the fail-closed
// precondition Run's own corevalidate.Validate(program) call already relies
// on, never the ownership fact itself.
var forbiddenOwnershipAccessors = []string{"PeerSignatures", "PeerSiteCoverage"}

// scanForForbiddenOwnershipAccessors parses source with go/parser and
// reports every forbidden accessor name (forbiddenOwnershipAccessors)
// referenced anywhere as a selector expression (`x.Name`), regardless of
// x's own static type -- a purely syntactic scan, matching this repo's
// established import/reference-scanning guard-test technique (see
// originvalidate_test.go's TestOriginValidatorImportsStayIndependent).
func scanForForbiddenOwnershipAccessors(filename string, source []byte) ([]string, error) {
	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, source, 0)
	if err != nil {
		return nil, err
	}
	var found []string
	ast.Inspect(file, func(n ast.Node) bool {
		selector, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		for _, forbidden := range forbiddenOwnershipAccessors {
			if selector.Sel.Name == forbidden {
				found = append(found, forbidden)
			}
		}
		return true
	})
	return found, nil
}

// TestInterpDoesNotReadCorevalidateOwnershipFields is Task 3's OWN-05b
// guard (D-10-36/D-10-37, QLT-08): today interp references neither
// corevalidate.Result.PeerSignatures nor .PeerSiteCoverage, but that is
// true by OMISSION, not by enforcement (D-10-36) -- nothing today converts
// a future reference into a build failure or a red test. This test
// converts it into "cannot without a red test": it scans every non-test
// .go file in this package via go/parser, resolved through the existing
// interpProjectRoot() helper (never testsupport, which pulls in session,
// which imports interp -- a cycle), and fails if any references either
// forbidden accessor.
//
// NARROWED CLAIM (D-10-37, PHASE-10-DEBT.md): this proves independence of
// DERIVATION MECHANISM for the ownership-transfer fact specifically,
// nested inside a shared, unrelated validation dependency -- NOT the
// mutual non-import independence check and corevalidate have from each
// other. interp's Run still calls corevalidate.Validate(program) as its
// own fail-closed precondition (import kept deliberately: hoisting
// Validate to Run's own callers would weaken interp's fail-closed posture
// and touch every call site, for an import-graph purity the current
// one-value domain does not need), so a bug in Validate's own derivation
// would feed interp bad input too (Knight and Leveson's correlated-fault
// result). The claim this test proves is narrower: interp never
// additionally CONSULTS corevalidate's own ownership-bearing verdict to
// decide its OWN runtime behavior -- OWN-05b's fact falls out of interp's
// own execution alone.
//
// The negative control proves the guard is load-bearing rather than
// vacuously green: it runs the identical scan over a small synthetic
// source string containing a forbidden accessor reference and asserts the
// scan reports it -- a guard that has never been seen to fire is not
// evidence.
func TestInterpDoesNotReadCorevalidateOwnershipFields(t *testing.T) {
	t.Run("real_scan", func(t *testing.T) {
		dir := filepath.Join(interpProjectRoot(), "internal", "compiler", "interp")
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, entry.Name())
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			found, err := scanForForbiddenOwnershipAccessors(path, source)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}
			if len(found) > 0 {
				t.Fatalf("%s references forbidden ownership-bearing accessor(s) %v -- interp must derive OWN-05b from its own observable execution, never by reading corevalidate's verdict", entry.Name(), found)
			}
		}
	})

	t.Run("negative_control", func(t *testing.T) {
		const synthetic = `package fake

type result struct{}

func (result) PeerSignatures() int { return 0 }

func probe(r result) int {
	return r.PeerSignatures()
}
`
		found, err := scanForForbiddenOwnershipAccessors("synthetic.go", []byte(synthetic))
		if err != nil {
			t.Fatalf("parse synthetic source: %v", err)
		}
		if len(found) == 0 {
			t.Fatal("expected the scan to report a violation on synthetic source containing a forbidden accessor reference -- the guard has never been observed to fire")
		}
	})
}

// generateCallDepthChainSource emits a genuine `.lang` module of n chained
// single-call functions, generalizing testdata/phase07/call_basic.lang's
// two-function template to n links (Task 2): link0 calls link1 calls
// link2 ... calls link(n-1), which is the base case with no call and simply
// returns its own parameter. link0 is the sole exported entry point. The
// full contract (module shape, per-link template, entry point name, and
// why the chain is generated rather than committed as a 129-function
// `.lang` file) is recorded in
// testdata/phase10/call_depth_chain_generator.md so a future reader can
// reconstruct the fixture without reading this function.
func generateCallDepthChainSource(n int) []byte {
	var b strings.Builder
	b.WriteString("module phase10.call_depth_chain\n\n")
	b.WriteString("export {\n  fn link0\n}\n\n")
	for i := n - 1; i >= 0; i-- {
		if i == n-1 {
			fmt.Fprintf(&b, "fn link%d(value: Byte) -> Byte {\n  value\n}\n\n", i)
			continue
		}
		fmt.Fprintf(&b, "fn link%d(value: Byte) -> Byte {\n  let result = link%d(value)\n  result\n}\n\n", i, i+1)
	}
	return []byte(b.String())
}

// generateAndCheckCallDepthChain generates an n-link call-depth chain
// (generateCallDepthChainSource) and drives it through the REAL pipeline --
// syntax.Parse -> check.Program -> corevalidate.Validate -- asserting a
// clean result at EACH stage before returning the checked core.Program, so
// a failure at an earlier stage is reported as itself rather than being
// mistaken for a depth refusal (Task 2, D-10-24: never a hand-built
// core.Program). Parameterized by n so this ONE helper drives Task 1's
// both-directions boundary test (n = MaxCallDepth, n = MaxCallDepth+1),
// Task 2's pipeline-conformance test, and Task 3's subprocess probe.
func generateAndCheckCallDepthChain(t *testing.T, n int) core.Program {
	t.Helper()
	source := generateCallDepthChainSource(n)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("parse (n=%d): unexpected diagnostics: %v", n, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("check (n=%d): unexpected diagnostics: %v", n, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("corevalidate (n=%d) rejected: %v", n, validated.Problems)
	}
	return validated.Program()
}

// TestCallDepthAtAndOverTheCap is Task 1's boundary test (SEM-08,
// D-10-23): a chain of exactly MaxCallDepth functions must execute to
// completion and return normally, while a chain one function deeper must
// terminate in the named depth-exceeded refusal -- both driven through the
// REAL pipeline via generateAndCheckCallDepthChain, never a hand-built
// core.Program.
func TestCallDepthAtAndOverTheCap(t *testing.T) {
	t.Run("at_the_cap", func(t *testing.T) {
		program := generateAndCheckCallDepthChain(t, MaxCallDepth)
		result, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("Run at MaxCallDepth (%d): unexpected error: %v", MaxCallDepth, err)
		}
		if result.Outcome.Kind != "returned" {
			t.Fatalf("expected outcome kind %q at the cap, got %q", "returned", result.Outcome.Kind)
		}
		if result.Outcome.Value != "7" {
			t.Fatalf("expected the chain's own base-case value %q, got %q", "7", result.Outcome.Value)
		}
	})

	t.Run("over_the_cap", func(t *testing.T) {
		program := generateAndCheckCallDepthChain(t, MaxCallDepth+1)
		result, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("Run over MaxCallDepth: unexpected error: %v", err)
		}

		validKind := false
		for _, kind := range execution.TerminalOutcomeKinds() {
			if result.Outcome.Kind == kind {
				validKind = true
				break
			}
		}
		if !validKind {
			t.Fatalf("expected the refusal's outcome kind %q to be a member of execution.TerminalOutcomeKinds()", result.Outcome.Kind)
		}

		foundReason := false
		for _, event := range result.Events {
			if event.Output == callDepthExceededDefectReason {
				foundReason = true
			}
		}
		if !foundReason {
			t.Fatalf("expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
		}

		firstBytes, err := CanonicalBytes(result)
		if err != nil {
			t.Fatalf("CanonicalBytes (first): %v", err)
		}
		if !bytes.Contains(firstBytes, []byte(callDepthExceededDefectReason)) {
			t.Fatalf("expected CanonicalBytes to contain the named depth-limit reason %q, got: %s", callDepthExceededDefectReason, firstBytes)
		}

		second, err := Run(program, "link0", "7")
		if err != nil {
			t.Fatalf("second Run over MaxCallDepth: unexpected error: %v", err)
		}
		secondBytes, err := CanonicalBytes(second)
		if err != nil {
			t.Fatalf("CanonicalBytes (second): %v", err)
		}
		if !bytes.Equal(firstBytes, secondBytes) {
			t.Fatalf("expected two runs of the same over-depth program to produce byte-identical CanonicalBytes, got:\n%s\nvs\n%s", firstBytes, secondBytes)
		}
	})
}

// TestCallDepthExceeded is Task 2's pipeline-conformance test (SEM-08): a
// genuine 129-function chain -- one function past MaxCallDepth -- is
// generated, parsed, checked, and corevalidated
// (generateAndCheckCallDepthChain, which itself asserts a clean result at
// each of those three stages), and interp.Run on the checked program must
// reach the named depth-exceeded refusal, never a parse/check/corevalidate
// diagnostic mistaken for it. See
// testdata/phase10/call_depth_chain_generator.md for the generator's full
// contract. D-10-24: this NEVER hand-builds a core.Program -- the whole
// point is that the compiler's real pipeline genuinely admits this program.
func TestCallDepthExceeded(t *testing.T) {
	const n = MaxCallDepth + 1
	program := generateAndCheckCallDepthChain(t, n)

	result, err := Run(program, "link0", "9")
	if err != nil {
		t.Fatalf("Run(%d-function chain): unexpected error: %v", n, err)
	}
	if result.Outcome.Kind != execution.OutcomeDefect {
		t.Fatalf("expected the depth-exceeded refusal's outcome kind %q, got %q", execution.OutcomeDefect, result.Outcome.Kind)
	}
	found := false
	for _, event := range result.Events {
		if event.Output == callDepthExceededDefectReason {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
	}
}

// probeReducedCallDepthCap and probeCallDepthMultiplier are Task 3's
// Pitfall-4 probe constants (D-10-44): the production MaxCallDepth (128)
// times a multiplier large enough to meaningfully exercise a pinned small
// host stack would exceed the parser's own maxFunctions ceiling (1024,
// syntax/parser.go:16), so the probe expresses its multiplier against this
// smaller, in-test-only cap instead. Both constants -- and the derived
// probeChainDepth -- are fixed in source BEFORE this probe's first passing
// run, so the "threshold chosen after seeing the result" attack is
// foreclosed; the commit that first makes TestNativeStackHeadroomIndependentOfCallDepth
// pass states in its own message that these constants predate that first
// green run, so review can confirm the ordering from git history alone.
const (
	probeReducedCallDepthCap  = 8
	probeCallDepthMultiplier  = 100
	probeChainDepth           = probeReducedCallDepthCap * probeCallDepthMultiplier // 800, held well under maxFunctions=1024
	probeMaxStackBytes        = 1 << 20                                            // 1 MiB: a deliberately small, pinned host ceiling
	probeSubprocessTimeout    = 30 * time.Second
	probeChildEnv             = "LANG_INTERP_CALL_DEPTH_PROBE_CHILD"
	probeArmEnv               = "LANG_INTERP_CALL_DEPTH_PROBE_ARM"
	probeArmCapDisabled       = "cap_disabled"
	probeArmCapEnabled        = "cap_enabled"
	probeMaxCapturedOutputLen = 1 << 16
)

// probeBoundedWriter caps captured child-process output at
// probeMaxCapturedOutputLen, mirroring testsupport.boundedWriter's shape
// (internal/compiler/testsupport/testsupport.go) without depending on its
// unexported type -- an interp test cannot import testsupport (it pulls in
// session, which imports interp: a cycle). Excess bytes are discarded, never
// buffered, so a runaway child cannot exhaust test memory.
type probeBoundedWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *probeBoundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := probeMaxCapturedOutputLen - w.buffer.Len()
	if remaining > 0 {
		if len(data) > remaining {
			w.buffer.Write(data[:remaining])
		} else {
			w.buffer.Write(data)
		}
	}
	return len(data), nil
}

func (w *probeBoundedWriter) bytes() []byte { return w.buffer.Bytes() }

// runCallDepthProbeSubprocess re-execs the current test binary
// (os.Args[0]) restricted to this one test via -test.run, guarded by
// probeChildEnv so the child branches into runCallDepthProbeChild instead
// of recursing, with probeArmEnv selecting which arm the child runs.
// exec.CommandContext with a real deadline (never context.Background()) and
// two independently bounded probeBoundedWriter streams (never .Output() or
// .CombinedOutput()'s unbounded merge, and never .StdoutPipe()'s unbounded
// pipe read) -- this repository's own TestSourceNeverSpawnsUnboundedProcesses
// guard (internal/compiler/native/native_test.go, D-02-01) forbids both.
func runCallDepthProbeSubprocess(t *testing.T, arm string) (stdout, stderr []byte, err error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), probeSubprocessTimeout)
	defer cancel()

	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeStackHeadroomIndependentOfCallDepth$", "-test.v")
	command.Env = append(os.Environ(), probeChildEnv+"=1", probeArmEnv+"="+arm)

	var stdoutWriter, stderrWriter probeBoundedWriter
	command.Stdout = &stdoutWriter
	command.Stderr = &stderrWriter

	runErr := command.Run()
	return stdoutWriter.bytes(), stderrWriter.bytes(), runErr
}

// runCallDepthProbeChild is the re-exec'd child's own logic (Task 3,
// D-10-42/D-10-43): it pins runtime/debug.SetMaxStack to a small, fixed
// ceiling (probeMaxStackBytes) for determinism across platforms and across
// the Linux and macOS CI runners, then drives a genuine probeChainDepth-
// function chain -- generated and checked through the real pipeline exactly
// like every other test in this file -- under one of two arms selected by
// probeArmEnv.
//
// This is genuinely NEW machinery for this repository: no subprocess,
// SetMaxStack, or TestMain re-exec pattern existed anywhere in this
// codebase before this test (10-RESEARCH.md's "No Analog Found" note). It
// converts Assumption A1 ([ASSUMED], 10-RESEARCH.md Assumptions Log) --
// that Go's stack-exhaustion runtime.throw is fatal and unrecoverable via
// recover() -- from an assumption into a directly observed result for this
// repository's toolchain: this test never exercises that path at all,
// because it exists precisely to demonstrate the STRONGER, honest finding
// (D-10-43) that language call depth never touches host stack in the first
// place. Because interp's call stack is an explicit heap []frame slice
// (D-10-21), Lang-level call depth costs O(1) host stack; the result below
// is therefore not a safety-margin ratio or a near-miss stress test -- it
// is a direct measurement that the host-stack limit and the language-level
// call-depth bound are two STRUCTURALLY UNRELATED limits, never one limit
// wearing two names (Roadmap criterion 2's Pitfall-4 Gate).
func runCallDepthProbeChild(t *testing.T) {
	t.Helper()
	debug.SetMaxStack(probeMaxStackBytes)

	switch arm := os.Getenv(probeArmEnv); arm {
	case probeArmCapDisabled:
		// Cap lifted through the nil-default maxCallDepthOverride seam
		// (never a production-mutable exported global): the chain runs
		// probeChainDepth deep -- many multiples past MaxCallDepth -- and
		// must exit cleanly despite the pinned small host ceiling, because
		// language call depth was never consuming host stack to begin
		// with.
		previous := maxCallDepthOverride
		maxCallDepthOverride = func() int { return probeChainDepth + 1 }
		defer func() { maxCallDepthOverride = previous }()

		program := generateAndCheckCallDepthChain(t, probeChainDepth)
		result, err := Run(program, "link0", "1")
		if err != nil {
			t.Fatalf("cap-disabled arm: unexpected error at chain depth %d under a %d-byte host stack ceiling: %v", probeChainDepth, probeMaxStackBytes, err)
		}
		if result.Outcome.Kind != "returned" {
			t.Fatalf("cap-disabled arm: expected outcome kind %q, got %q", "returned", result.Outcome.Kind)
		}
		fmt.Println("cap_disabled: ok")

	case probeArmCapEnabled:
		// The cap stays at its declared production value (MaxCallDepth,
		// maxCallDepthOverride left nil): the SAME probeChainDepth chain is
		// driven under the identical pinned small host stack ceiling, and
		// the named depth-exceeded refusal must fire well before the
		// chain's own end -- the language-level bound is what stops
		// execution, never a host stack limit that was never actually
		// threatened.
		program := generateAndCheckCallDepthChain(t, probeChainDepth)
		result, err := Run(program, "link0", "1")
		if err != nil {
			t.Fatalf("cap-enabled arm: unexpected error: %v", err)
		}
		if result.Outcome.Kind != execution.OutcomeDefect {
			t.Fatalf("cap-enabled arm: expected the depth-exceeded refusal's outcome kind %q, got %q", execution.OutcomeDefect, result.Outcome.Kind)
		}
		found := false
		for _, event := range result.Events {
			if event.Output == callDepthExceededDefectReason {
				found = true
			}
		}
		if !found {
			t.Fatalf("cap-enabled arm: expected an event carrying the named depth-limit reason %q, got events: %+v", callDepthExceededDefectReason, result.Events)
		}
		fmt.Println("cap_enabled: ok")

	default:
		t.Fatalf("unknown or missing probe arm %q (%s)", arm, probeArmEnv)
	}
}

// TestNativeStackHeadroomIndependentOfCallDepth is Task 3's Pitfall-4 gate
// (Roadmap criterion 2, D-10-42/D-10-43/D-10-44): a SUBPROCESS probe, never
// an in-process headroom-ratio measurement -- an in-process measurement is
// an assertion wearing observation's clothes and is REJECTED as this gate's
// evidence, usable only as a supplementary fast sanity check. The parent
// asserts on the child's own EXIT STATUS and STDOUT/STDERR, never on an
// in-process value.
//
// Two arms, both required: cap_disabled proves the language-level call
// stack survives probeChainDepth (many multiples past MaxCallDepth) under a
// pinned small host stack ceiling with no host collapse; cap_enabled proves
// the named depth-exceeded refusal fires FIRST, before that same host
// ceiling is ever approached, at the SAME chain depth. See
// runCallDepthProbeChild's own doc comment for the full framing and its
// recorded observation for Assumption A1.
func TestNativeStackHeadroomIndependentOfCallDepth(t *testing.T) {
	if os.Getenv(probeChildEnv) != "" {
		runCallDepthProbeChild(t)
		return
	}

	t.Run(probeArmCapDisabled, func(t *testing.T) {
		stdout, stderr, err := runCallDepthProbeSubprocess(t, probeArmCapDisabled)
		if err != nil {
			t.Fatalf("cap-disabled child exited with error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if !bytes.Contains(stdout, []byte("cap_disabled: ok")) {
			t.Fatalf("expected the cap-disabled child to report success on its own stdout; stdout: %s\nstderr: %s", stdout, stderr)
		}
	})

	t.Run(probeArmCapEnabled, func(t *testing.T) {
		stdout, stderr, err := runCallDepthProbeSubprocess(t, probeArmCapEnabled)
		if err != nil {
			t.Fatalf("cap-enabled child exited with error: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
		}
		if !bytes.Contains(stdout, []byte("cap_enabled: ok")) {
			t.Fatalf("expected the cap-enabled child to report the depth refusal fired first on its own stdout; stdout: %s\nstderr: %s", stdout, stderr)
		}
	})
}
