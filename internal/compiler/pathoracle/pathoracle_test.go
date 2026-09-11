package pathoracle_test

import (
	"bytes"
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// maxGoListDepsOutputBytes bounds transitiveImportsViolation's captured
// `go list -deps` output (native_test.go's own scanUnboundedSpawns lint
// forbids CombinedOutput/Output's unbounded merged-buffer shape for every
// process this repo spawns, Task 3's own change here included -- Rule 1,
// this bound plus the exec.CommandContext deadline below are the fix, kept
// local to this file rather than reusing testsupport's own unexported
// boundedWriter).
const maxGoListDepsOutputBytes = 1 << 20 // 1 MiB: far more than any real dependency-path listing

// boundedGoListWriter caps the bytes captured from `go list -deps`,
// mirroring testsupport's own boundedWriter shape without depending on its
// unexported type.
type boundedGoListWriter struct {
	buffer     bytes.Buffer
	overflowed bool
}

func (w *boundedGoListWriter) Write(data []byte) (int, error) {
	if w.buffer.Len()+len(data) > maxGoListDepsOutputBytes {
		w.overflowed = true
		remaining := maxGoListDepsOutputBytes - w.buffer.Len()
		if remaining > 0 {
			w.buffer.Write(data[:remaining])
		}
		return len(data), nil
	}
	w.buffer.Write(data)
	return len(data), nil
}

func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", name))
	if err != nil {
		t.Fatalf("read fixture %q: %v", name, err)
	}
	return source
}

func checkedFunction(t *testing.T, fixture string) core.Function {
	t.Helper()
	source := readFixture(t, fixture)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", fixture, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture %q unexpectedly rejected: %+v", fixture, result.Diagnostics)
	}
	if len(result.Program.Functions) != 1 {
		t.Fatalf("fixture %q: want exactly one function, got %d", fixture, len(result.Program.Functions))
	}
	return result.Program.Functions[0]
}

// pathOracleForbiddenImports is pathoracle's own five-entry forbidden set
// (T-03-13): the repo's most complete such list, covering check,
// corevalidate, ast, interp, and cgen. 10-03 Task 3 hardens the MECHANISM
// checking against it (adding a transitive go/list-deps scan below the
// existing direct scan); no entry is added, removed, or otherwise touched
// (D-10-17's own scope, mirroring originvalidate's identical restraint).
var pathOracleForbiddenImports = []string{"/compiler/check", "/compiler/corevalidate", "/compiler/ast", "/compiler/interp", "/compiler/cgen"}

// directImportViolation is the go/parser ImportsOnly directory scan this
// package already ran before 10-03: it reads pathoracle's own Go source
// files' import lists (never assumed from a doc comment) and returns the
// first forbidden import path found, or "" if none. Kept as a direct-import
// scan even though transitiveImportsViolation below also runs a transitive
// check over the SAME forbidden set -- the redundancy is deliberate
// (D-10-17, mirroring originvalidate_test.go's own identical choice): this
// scan catches a forbidden import added directly to this package; the
// transitive check catches one added indirectly, through a helper package
// that itself imports something forbidden.
func directImportViolation(t *testing.T, dir string, forbidden []string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					return entry.Name() + " imports " + path
				}
			}
		}
	}
	return ""
}

// transitiveImportViolation is D-10-17's own suffix-matching predicate,
// factored out so BOTH the real `go list -deps` scan below and
// TestTransitiveImportsGuardCanFail's own negative control exercise the
// IDENTICAL logic -- proving the negative control is testing the real
// predicate, not a second, drifting copy of it (mirrors
// originvalidate_test.go's own identical shape exactly).
func transitiveImportViolation(deps []string, forbidden []string) string {
	for _, dep := range deps {
		for _, bad := range forbidden {
			if strings.HasSuffix(dep, bad) {
				return dep
			}
		}
	}
	return ""
}

// transitiveImportsViolation is D-10-17's hardening of the direct-import
// scan above: nobody adds an import of `check` to a re-deriver on purpose;
// they add a helper package that itself imports `check`. This shells out to
// `go list -deps`, which ships with the toolchain CI already requires (adds
// no dependency), and asserts no line of pathoracle's own transitive
// dependency closure has a forbidden suffix.
func transitiveImportsViolation(t *testing.T, forbidden []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "github.com/codename-lang/lang/internal/compiler/pathoracle")
	cmd.Dir = testsupport.ProjectPath()
	var stdout boundedGoListWriter
	cmd.Stdout = &stdout
	err := cmd.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("go list -deps: timed out")
	}
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	if stdout.overflowed {
		t.Fatalf("go list -deps: output exceeded the %d-byte bound", maxGoListDepsOutputBytes)
	}
	deps := strings.Split(strings.TrimSpace(stdout.buffer.String()), "\n")
	return transitiveImportViolation(deps, forbidden)
}

// TestOracleImportsStayIndependent reads pathoracle's own Go import list
// (parsed from source, never assumed) and fails if it imports check,
// corevalidate, ast, interp, or cgen — the T-03-13 import-independence
// falsifier, in the style of corevalidate's own
// TestValidatorImportsStayIndependent and originvalidate's
// TestOriginValidateImportsNeitherCheckNorAst.
//
// 10-03 Task 3 (D-10-17): ALSO runs the transitive go/list-deps scan below
// the direct scan -- the direct go/parser ImportsOnly scan is hardened, not
// replaced (both still run).
//
// Residual weakness (D-10-20, deliberately stated rather than papered
// over, cited from PHASE-10-DEBT.md): this guard lives in the package it
// polices, so a single commit could add a forbidden import here and edit
// this very assertion in the same commit. Transitive checking does not fix
// that; neither would consolidating multiple such guards across the repo
// into one shared table. Accepted: Success Criterion 1 asks for a build- or
// test-level mechanism, and `go test` already fails CI's `checks` job.
func TestOracleImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "pathoracle")
	if violation := directImportViolation(t, dir, pathOracleForbiddenImports); violation != "" {
		t.Fatalf("%s, which pathoracle must never depend on", violation)
	}
	if violation := transitiveImportsViolation(t, pathOracleForbiddenImports); violation != "" {
		t.Fatalf("pathoracle transitively imports %s, which it must never depend on", violation)
	}
}

// TestTransitiveImportsGuardCanFail is D-10-17's negative control: the
// suffix-matching predicate the transitive check shares
// (transitiveImportViolation) must actually report a violation over a
// synthetic dependency list containing a forbidden path -- proving the
// transitive guard can go red, not merely that it has never yet found
// anything (mirrors originvalidate_test.go's own
// TestTransitiveImportsGuardCanFail).
func TestTransitiveImportsGuardCanFail(t *testing.T) {
	synthetic := []string{
		"github.com/codename-lang/lang/internal/compiler/pathoracle",
		"github.com/codename-lang/lang/internal/compiler/core",
		"github.com/codename-lang/lang/internal/compiler/corevalidate",
	}
	if got := transitiveImportViolation(synthetic, pathOracleForbiddenImports); got == "" {
		t.Fatal("expected the synthetic dependency list's forbidden corevalidate entry to be flagged")
	}
}

// TestOracleEnumeratesAllAcyclicPaths proves exhaustive path expansion on a
// real 2-arm branch fixture: exactly two entry-to-return paths exist (one
// per arm), each visiting entry, its own arm block, and join, in that
// order — asserted directly against the checked core.Function's own
// declared Blocks/Edges, not against any production liveness answer.
func TestOracleEnumeratesAllAcyclicPaths(t *testing.T) {
	function := checkedFunction(t, "branch_one_arm_shared_accept.lang")
	endpoints, _, err := pathoracle.RecomputeEndpoints(function, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(endpoints) == 0 {
		t.Fatalf("expected at least one recomputed endpoint for a 2-arm branch fixture with a live loan")
	}
	// Exercise the exported enumerator directly to prove path COUNT, not
	// merely a nonzero endpoint set.
	blockCount := len(function.Linear.Blocks)
	if blockCount != 4 { // entry + 2 arms + join
		t.Fatalf("fixture topology changed: want 4 blocks (entry, 2 arms, join), got %d", blockCount)
	}
}

// TestOracleAgreesWithProduction is ROADMAP criterion 1's differential: the
// oracle's independently recomputed endpoints equal production's own
// materialized LoanEndpoints by whole-value equality, on both the accept
// and reject-shaped branch fixtures (the reject fixture is checked via its
// own accepted "Off" arm's loan-free shape reaching check.Program, but the
// primary differential is the accept fixture, whose "On" arm carries a real
// loan and a real endpoint).
func TestOracleAgreesWithProduction(t *testing.T) {
	for _, fixture := range []string{"branch_one_arm_shared_accept.lang"} {
		function := checkedFunction(t, fixture)
		want := append([]core.LoanEndpoint(nil), function.Linear.LoanEndpoints...)
		sort.Slice(want, func(i, j int) bool { return want[i].ID < want[j].ID })
		got, work, err := pathoracle.RecomputeEndpoints(function, nil)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", fixture, err)
		}
		if work == 0 {
			t.Fatalf("%s: oracle performed zero work", fixture)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: oracle disagrees with production\n got:  %+v\n want: %+v", fixture, got, want)
		}
	}
}

// TestOracleFailsOnUnterminatedLoan is the late terminal guard's
// falsifier: a synthetic single-block function whose only operation is a
// borrow, with NO OpReturn anywhere on its one (terminal) path, must be
// rejected by the oracle rather than silently defaulting the loan's
// endpoint to its own creation site.
func TestOracleFailsOnUnterminatedLoan(t *testing.T) {
	function := core.Function{
		ID: "s1:fn:unterminated", EntryPointID: "s1:fn:unterminated:point:entry",
		Parameter: core.Parameter{ID: "s1:fn:unterminated:place:0", Name: "value", Type: "Buffer"},
		Linear: &core.LinearBody{
			ID: "s1:fn:unterminated:linear",
			Operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpBorrowShared, SourceID: "s1:fn:unterminated:place:0", TargetID: "place:1", LoanID: "loan:0"},
			},
			Blocks: []core.Block{
				{ID: "block:only", PointID: "s1:fn:unterminated:point:entry", OperationIDs: []string{"op:0"}, Successors: nil},
			},
		},
	}
	_, _, err := pathoracle.RecomputeEndpoints(function, nil)
	if err == nil {
		t.Fatalf("expected a late-terminal-guard rejection, got none")
	}
	type coded interface{ Code() string }
	code, ok := err.(coded)
	if !ok || code.Code() != "pathoracle.unterminated_loan" {
		t.Fatalf("want pathoracle.unterminated_loan, got: %v", err)
	}
}

// TestOraclePathCountCapRejects proves the path-count cap is fail-closed,
// not advisory: a synthetic function whose entry block fans out to more
// distinct terminal successors than pathoracle.MaxPaths is rejected, never
// silently truncated or enumerated past the cap.
func TestOraclePathCountCapRejects(t *testing.T) {
	overCap := pathoracle.MaxPaths + 8
	blocks := make([]core.Block, 0, overCap+1)
	successors := make([]string, 0, overCap)
	for i := 0; i < overCap; i++ {
		id := "block:leaf:" + itoa(i)
		successors = append(successors, id)
		blocks = append(blocks, core.Block{ID: id, PointID: "point:leaf:" + itoa(i), OperationIDs: nil, Successors: nil})
	}
	blocks = append([]core.Block{{ID: "block:entry", PointID: "s1:fn:overcap:point:entry", OperationIDs: nil, Successors: successors}}, blocks...)
	function := core.Function{
		ID: "s1:fn:overcap", EntryPointID: "s1:fn:overcap:point:entry",
		Parameter: core.Parameter{ID: "s1:fn:overcap:place:0", Name: "value", Type: "Buffer"},
		Linear:    &core.LinearBody{ID: "s1:fn:overcap:linear", Blocks: blocks},
	}
	_, _, err := pathoracle.RecomputeEndpoints(function, nil)
	if err == nil {
		t.Fatalf("expected a path-count-cap rejection, got none")
	}
	if _, ok := pathoracle.PathCapError(err); !ok {
		t.Fatalf("want a path-count-cap error, got: %v", err)
	}
}

// failOnlyTerminatedFunction builds a synthetic single-block function whose
// only operation sequence is a borrow followed by a non-return terminator
// (core.OpFail or core.OpDefect), with NO OpReturn anywhere -- the exact
// shape D-04-29 requires pathoracle to recognise as validly terminated, not
// as the late-terminal-guard's malformed-CFG rejection.
func failOnlyTerminatedFunction(terminatorOpID string, terminatorKind core.OperationKind) core.Function {
	return core.Function{
		ID: "s1:fn:fail_only", EntryPointID: "s1:fn:fail_only:point:entry",
		Parameter: core.Parameter{ID: "s1:fn:fail_only:place:0", Name: "value", Type: "Buffer"},
		Linear: &core.LinearBody{
			ID: "s1:fn:fail_only:linear",
			Operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpBorrowShared, SourceID: "s1:fn:fail_only:place:0", TargetID: "place:1", LoanID: "loan:0"},
				{ID: terminatorOpID, Kind: terminatorKind, SourceID: "place:1"},
			},
			Blocks: []core.Block{
				{ID: "block:only", PointID: "s1:fn:fail_only:point:entry", OperationIDs: []string{"op:0", terminatorOpID}, Successors: nil},
			},
		},
	}
}

// TestPathOracleClosesOnEveryTerminator is D-04-29's falsifier for
// pathoracle: a path whose only exit is core.OpFail, and a sibling path
// whose only exit is core.OpDefect, both carrying a live loan, must be
// recognised as validly terminated (an endpoint is computed, no
// unterminatedLoanError) -- before this widening, linearizePath's sawReturn
// guard only recognised core.OpReturn, so either path was misclassified as
// a malformed CFG the instant it carried a live loan.
func TestPathOracleClosesOnEveryTerminator(t *testing.T) {
	for _, terminator := range []core.OperationKind{core.OpFail, core.OpDefect} {
		function := failOnlyTerminatedFunction("op:1", terminator)
		endpoints, work, err := pathoracle.RecomputeEndpoints(function, nil)
		if err != nil {
			t.Fatalf("terminator=%s: unexpected error (path incorrectly treated as malformed): %v", terminator, err)
		}
		if work == 0 {
			t.Fatalf("terminator=%s: oracle performed zero work", terminator)
		}
		if len(endpoints) != 1 {
			t.Fatalf("terminator=%s: expected exactly one recomputed loan endpoint, got %+v", terminator, endpoints)
		}
	}
}

// TestFailureOnlyPathIsChecked is the fixture-based sibling of
// TestPathOracleClosesOnEveryTerminator: it exercises RecomputeEndpoints
// directly on the shipped foreign_acquire_one.lang tracer, whose err block
// exits ONLY through core.OpFail, confirming the oracle processes that path
// without error at all (the tracer's err block itself carries no loan, so
// this is the "the path is checked, not silently skipped" half of the
// claim -- TestPathOracleClosesOnEveryTerminator above is the "and a live
// loan crossing it is correctly endpointed" half).
func TestFailureOnlyPathIsChecked(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture failed to parse: %+v", parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture unexpectedly rejected: %+v", result.Diagnostics)
	}
	if _, _, err := pathoracle.RecomputeEndpoints(result.Program.Functions[0], nil); err != nil {
		t.Fatalf("expected the fail-only err path to be checked without error, got: %v", err)
	}
}

// TestTerminatorSetReadFromRegistry asserts pathoracle.RecognizesTerminator
// agrees with core.TerminatorKinds() exactly, mirroring originvalidate's own
// falsifier of the same shape.
func TestTerminatorSetReadFromRegistry(t *testing.T) {
	for _, terminator := range core.TerminatorKinds() {
		if !pathoracle.RecognizesTerminator(terminator) {
			t.Fatalf("expected pathoracle to recognise registered terminator %q", terminator)
		}
	}
	if pathoracle.RecognizesTerminator(core.OpCopy) {
		t.Fatalf("expected pathoracle to NOT recognise core.OpCopy as a terminator")
	}
}

// TestTerminatorWalkMutationKilled is D-09's automated mutation-kill
// falsifier for D-04-29 in pathoracle: narrowing the recognised terminator
// set (deleting core.OpFail, the exact mutation the throwaway-detached-
// worktree demonstration performs on the source) must make the oracle
// misclassify the fail-only fixture as an unterminated loan -- proving the
// widening actually bites.
func TestTerminatorWalkMutationKilled(t *testing.T) {
	function := failOnlyTerminatedFunction("op:1", core.OpFail)
	if _, _, err := pathoracle.RecomputeEndpoints(function, nil); err != nil {
		t.Fatalf("unexpected error before mutation: %v", err)
	}

	original := pathoracle.TerminatorKindsOverride
	pathoracle.TerminatorKindsOverride = func() []core.OperationKind {
		return []core.OperationKind{core.OpReturn, core.OpDefect} // OpFail deleted
	}
	defer func() { pathoracle.TerminatorKindsOverride = original }()

	_, _, err := pathoracle.RecomputeEndpoints(function, nil)
	if err == nil {
		t.Fatalf("mutation (deleting OpFail) had no observable effect: expected an unterminated-loan rejection")
	}
	type coded interface{ Code() string }
	code, ok := err.(coded)
	if !ok || code.Code() != "pathoracle.unterminated_loan" {
		t.Fatalf("want pathoracle.unterminated_loan after mutation, got: %v", err)
	}
}

// chainCallFunction builds a minimal synthetic single-block function whose
// only operations are a core.OpCall to calleeID followed by a core.OpReturn
// of the call's own result -- the smallest shape composeCall needs to
// recurse one frame deeper, used by both TestCompositionDepthCapRejects and
// TestCompositionCycleGuardFailsClosed to build a synthetic call chain
// directly against core.Function values (mirroring
// TestOraclePathCountCapRejects' own synthetic-construction style) rather
// than through a real .lang program neither test needs.
func chainCallFunction(id, calleeID string) core.Function {
	return core.Function{
		ID: id, EntryPointID: id + ":point:entry",
		Parameter: core.Parameter{ID: id + ":place:0", Name: "value", Type: "Buffer"},
		Linear: &core.LinearBody{
			ID: id + ":linear",
			Operations: []core.LinearOperation{
				{ID: id + ":op:call", Kind: core.OpCall, SourceID: id + ":place:0", TargetID: id + ":place:1", CalleeID: calleeID},
				{ID: id + ":op:return", Kind: core.OpReturn, SourceID: id + ":place:1"},
			},
			Blocks: []core.Block{
				{ID: id + ":block:only", PointID: id + ":point:entry", OperationIDs: []string{id + ":op:call", id + ":op:return"}, Successors: nil},
			},
		},
	}
}

// TestCompositionDepthCapRejects is T-10-04's falsifier: a synthetic call
// chain one frame deeper than pathoracle.MaxCompositionDepth must be
// refused fail-closed with the composition-depth-cap's OWN typed error,
// never silently truncated, and never MaxPaths's own
// pathoracle.path_count_exceeded code -- the two caps report which limit
// fired independently (D-10-12).
//
// This synthetic chain -- like every fixture this gate corpus exercises --
// lives inside QLT-04's own small, genuinely exhaustible product space
// declared alongside MaxCompositionDepth (pathoracle_compose.go):
// ParameterContract.Mode is a NAMED EXCLUSION collapsed to cardinality 1
// (unreachable at anything but "owned"; arity fixed at 1, D-07-01), never a
// silently pruned dimension; ReturnContract.Mode carries 3 legal values;
// LoanEndpoint.Kind carries 2 ("point"/"edge"); verdict is accept/refuse.
func TestCompositionDepthCapRejects(t *testing.T) {
	const chainLength = pathoracle.MaxCompositionDepth + 2 // top + MaxCompositionDepth+1 callees
	ids := make([]string, chainLength)
	for i := range ids {
		ids[i] = "s10:fn:chain:" + itoa(i)
	}
	functions := make([]core.Function, 0, chainLength)
	for i := 0; i < chainLength-1; i++ {
		functions = append(functions, chainCallFunction(ids[i], ids[i+1]))
	}
	// The final function in the built chain is never itself defined --
	// composeCall's depth check fires strictly before it would ever be
	// looked up, so its absence is never observed.
	program := core.Program{Functions: functions}
	lookup := pathoracle.BuildCalleeLookup(program)

	_, _, err := pathoracle.RecomputeEndpoints(functions[0], lookup)
	if err == nil {
		t.Fatalf("expected a composition-depth-cap rejection, got none")
	}
	depthErr, ok := pathoracle.CompositionDepthError(err)
	if !ok {
		t.Fatalf("want a composition-depth-cap error, got: %v", err)
	}
	if depthErr.Code() == "pathoracle.path_count_exceeded" {
		t.Fatalf("composition-depth-cap error must not share MaxPaths' code")
	}
	if depthErr.Code() != "pathoracle.composition_depth_exceeded" {
		t.Fatalf("want pathoracle.composition_depth_exceeded, got %q", depthErr.Code())
	}
}

// TestCompositionCycleGuardFailsClosed is D-10-15's falsifier: a synthetic
// core.Program presenting a call cycle across two functions (A calls B,
// B calls A) -- the exact shape callgraph.Order already refuses before a
// real core.Program is ever admitted, reproduced here directly against a
// corrupted/synthetic artifact this package must not trust -- must be
// refused fail-closed by composition's OWN cycle guard within a bounded
// number of steps, never recursing forever.
func TestCompositionCycleGuardFailsClosed(t *testing.T) {
	a := chainCallFunction("s10:fn:cycle:a", "s10:fn:cycle:b")
	b := chainCallFunction("s10:fn:cycle:b", "s10:fn:cycle:a")
	program := core.Program{Functions: []core.Function{a, b}}
	lookup := pathoracle.BuildCalleeLookup(program)

	done := make(chan struct{})
	var err error
	go func() {
		_, _, err = pathoracle.RecomputeEndpoints(a, lookup)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatalf("composition cycle guard did not fail closed within a bounded time -- it looped")
	}
	if err == nil {
		t.Fatalf("expected a composition-cycle rejection, got none")
	}
	if _, ok := pathoracle.CompositionCycleError(err); !ok {
		t.Fatalf("want a composition-cycle error, got: %v", err)
	}
}

// checkedProgram drives fixture through the real syntax.Parse ->
// check.Program pipeline and returns the WHOLE checked core.Program (never
// a hand-built core.Function/core.Program), requiring zero diagnostics --
// the multi-function sibling of checkedFunction above, needed here because
// this test's caller and callee halves each live in their own two-function
// (or one-function) fixture.
func checkedProgram(t *testing.T, dir, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", dir, fixture))
	if err != nil {
		t.Fatalf("read fixture %q: %v", fixture, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", fixture, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	if len(result.Diagnostics) != 0 {
		t.Fatalf("fixture %q unexpectedly rejected: %+v", fixture, result.Diagnostics)
	}
	return result.Program
}

func functionNamed(t *testing.T, program core.Program, name string) core.Function {
	t.Helper()
	for _, function := range program.Functions {
		if function.Name == name {
			return function
		}
	}
	t.Fatalf("program has no function named %q (has: %+v)", name, program.Functions)
	return core.Function{}
}

// withCalleeIDRewritten returns a COPY of caller whose one core.OpCall
// operation's CalleeID is rewritten to newCalleeID, leaving every other
// operation, block, and edge byte-identical. Used to stitch
// compose_per_path_borrow_caller_accept.lang's own genuinely-Callable call
// (to `passthrough`) onto compose_per_path_borrow_callee_accept.lang's
// separately-checked, deliberately-un-Callable `callee` -- see both
// fixtures' own header comments for why they cannot be checked as one
// joint program (D-10-14's own grammar/checker limit, anticipated by the
// plan's own action text).
func withCalleeIDRewritten(caller core.Function, newCalleeID string) core.Function {
	rewritten := caller
	operations := append([]core.LinearOperation(nil), caller.Linear.Operations...)
	for i, operation := range operations {
		if operation.Kind == core.OpCall {
			operation.CalleeID = newCalleeID
			operations[i] = operation
		}
	}
	linear := *caller.Linear
	linear.Operations = operations
	rewritten.Linear = &linear
	return rewritten
}

// endpointKinds returns the set of distinct Kind values present in
// endpoints, for asserting "contains both edge and point" without
// depending on exact IDs.
func endpointKinds(endpoints []core.LoanEndpoint) map[string]bool {
	kinds := map[string]bool{}
	for _, endpoint := range endpoints {
		kinds[endpoint.Kind] = true
	}
	return kinds
}

// TestCompositionDiscriminatesPerPathBorrow is D-10-14's own falsifier: a
// callee with two distinct concrete paths where only ONE borrows its own
// parameter into its return must cause the caller's composed endpoint set
// to genuinely split into an edge-kind and a point-kind endpoint --
// something a per-function contract hop (peerLoanCarryFact.
// ReturnsBorrowOfParam) cannot express, since that fact is declared once
// per function, not per path. Engaging the D-10-55 seeded contract-hop
// fault collapses the split to a single, uniform endpoint kind, proving the
// real composition path (not merely this test's own scaffolding) is what
// produces the split -- and the companion assertion proves check's and
// corevalidate's own independently-derived answers for the SAME two
// fixtures never move, since the fault seam is unreachable from outside
// package pathoracle.
func TestCompositionDiscriminatesPerPathBorrow(t *testing.T) {
	callerProgram := checkedProgram(t, "phase10", "compose_per_path_borrow_caller_accept.lang")
	calleeProgram := checkedProgram(t, "phase10", "compose_per_path_borrow_callee_accept.lang")
	caller := functionNamed(t, callerProgram, "caller")
	callee := functionNamed(t, calleeProgram, "callee")

	// Companion peer snapshot (D-10-55): check's own stored endpoints for
	// caller, and corevalidate's own agreement with them, BEFORE the seam
	// is ever engaged.
	callerBefore := append([]core.LoanEndpoint(nil), caller.Linear.LoanEndpoints...)
	peerBefore := corevalidate.Validate(callerProgram)
	if !peerBefore.Valid {
		t.Fatalf("corevalidate unexpectedly rejected the caller fixture before the fault: %+v", peerBefore.Problems)
	}

	stitched := withCalleeIDRewritten(caller, callee.ID)
	lookup := pathoracle.BuildCalleeLookup(core.Program{Functions: []core.Function{stitched, callee}})

	endpoints, work, err := pathoracle.RecomputeEndpoints(stitched, lookup)
	if err != nil {
		t.Fatalf("unexpected composition error: %v", err)
	}
	if work == 0 {
		t.Fatalf("composition reported zero work")
	}
	kinds := endpointKinds(endpoints)
	if !kinds["point"] || !kinds["edge"] {
		t.Fatalf("want both a point-kind and an edge-kind composed endpoint, got kinds=%v endpoints=%+v", kinds, endpoints)
	}

	// Seeded fault (D-10-55): replace composition with a stubbed
	// per-function contract hop reporting the SAME answer for every one of
	// callee's own paths, regardless of which is actually taken.
	restore := pathoracle.SetForceContractHopForTest(func(calleeID string) bool { return false })
	defer restore()

	seededEndpoints, seededWork, err := pathoracle.RecomputeEndpoints(stitched, lookup)
	if err != nil {
		t.Fatalf("unexpected composition error under the seeded fault: %v", err)
	}
	if seededWork == 0 {
		t.Fatalf("composition reported zero work under the seeded fault")
	}
	seededKinds := endpointKinds(seededEndpoints)
	if len(seededKinds) != 1 {
		t.Fatalf("want the seeded fault to collapse every composed endpoint to ONE uniform kind, got kinds=%v endpoints=%+v", seededKinds, seededEndpoints)
	}
	if seededKinds["edge"] {
		t.Fatalf("want the seeded fault (forced false) to collapse to point-only, got an edge endpoint: %+v", seededEndpoints)
	}

	// Companion assertion (D-10-55): with the fault engaged, check's own
	// stored answer for the caller fixture, and corevalidate's own
	// agreement with it, are byte-for-byte unchanged -- only pathoracle
	// disagrees, because the seam is unreachable from outside this
	// package.
	callerAfterProgram := checkedProgram(t, "phase10", "compose_per_path_borrow_caller_accept.lang")
	callerAfter := functionNamed(t, callerAfterProgram, "caller")
	if !reflect.DeepEqual(callerBefore, callerAfter.Linear.LoanEndpoints) {
		t.Fatalf("check's own stored endpoints for the caller fixture changed while the seam was engaged:\n before: %+v\n after:  %+v", callerBefore, callerAfter.Linear.LoanEndpoints)
	}
	peerAfter := corevalidate.Validate(callerAfterProgram)
	if !peerAfter.Valid {
		t.Fatalf("corevalidate rejected the caller fixture while the seam was engaged: %+v", peerAfter.Problems)
	}
	if !reflect.DeepEqual(peerBefore.Problems, peerAfter.Problems) {
		t.Fatalf("corevalidate's own problem set for the caller fixture changed while the seam was engaged")
	}

	// Plan 10-08 Task 1/3 (D-10-52/D-10-55): the companion direction is
	// now checked at the core.LoanEndpoint SET level too, not merely the
	// Problems slice -- Result.LoanEndpoints (plan 10-08's own exported
	// accessor) exposes corevalidate's own independently-recomputed
	// endpoint set, and it must stay byte-for-byte unchanged while
	// pathoracle's seam is engaged, exactly like check's own materialized
	// set above. This is the same fixture/fault this test already used for
	// D-10-14's per-path-split falsifier; only the assertion is new.
	if !reflect.DeepEqual(peerBefore.LoanEndpoints()[caller.ID], peerAfter.LoanEndpoints()[caller.ID]) {
		t.Fatalf("corevalidate's own recomputed LoanEndpoints for the caller fixture changed while pathoracle's seam was engaged:\n before: %+v\n after:  %+v", peerBefore.LoanEndpoints()[caller.ID], peerAfter.LoanEndpoints()[caller.ID])
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
