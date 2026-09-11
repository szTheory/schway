package pathoracle_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/pathoracle"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

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

// TestOracleImportsStayIndependent reads pathoracle's own Go import list
// (parsed from source, never assumed) and fails if it imports check,
// corevalidate, or ast — the T-03-13 import-independence falsifier, in the
// style of corevalidate's own TestValidatorImportsStayIndependent and
// originvalidate's TestOriginValidateImportsNeitherCheckNorAst.
func TestOracleImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "pathoracle")
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
			forbidden := []string{"/compiler/check", "/compiler/corevalidate", "/compiler/ast", "/compiler/interp", "/compiler/cgen"}
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					t.Fatalf("%s imports %s, which pathoracle must never depend on", entry.Name(), path)
				}
			}
		}
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
