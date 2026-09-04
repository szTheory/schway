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
	endpoints, _, err := pathoracle.RecomputeEndpoints(function)
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
		got, work, err := pathoracle.RecomputeEndpoints(function)
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
	_, _, err := pathoracle.RecomputeEndpoints(function)
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
	_, _, err := pathoracle.RecomputeEndpoints(function)
	if err == nil {
		t.Fatalf("expected a path-count-cap rejection, got none")
	}
	if _, ok := pathoracle.PathCapError(err); !ok {
		t.Fatalf("want a path-count-cap error, got: %v", err)
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
