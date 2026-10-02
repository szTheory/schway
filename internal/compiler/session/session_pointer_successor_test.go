package session_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase25FamilyConflict(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"shared", "module phase25.shared_conflict\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n let first = borrow value\n let second = borrow mut value\n let observed = take first\n second\n}\n"},
		{"exclusive", string(mustPhase25Fixture(t, "exclusive_conflict_reject.schway"))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checked := session.Check([]byte(test.source))
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != "ownership.borrow_conflict" {
				t.Fatalf("expected a semantic borrow conflict, got %+v", checked.Diagnostics)
			}
			problem := checked.Diagnostics[0]
			if problem.Primary.Start == problem.Primary.End || len(problem.Causes) < 3 {
				t.Fatalf("conflict lost primary source attribution or causes: %+v", problem)
			}
		})
	}
}

func TestPhase25FamilyEscape(t *testing.T) {
	result, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", "exclusive_escape_reject.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "invalid" || len(result.Diagnostics) != 1 {
		t.Fatalf("expected exclusive escape refusal, got %+v", result)
	}
	problem := result.Diagnostics[0]
	if problem.Code != "core.origin_omitted" {
		t.Fatalf("escape must identify its source and semantic boundary: %+v", problem)
	}
}

func TestPhase25EscapeDiagnosticSourceAttribution(t *testing.T) {
	tests := []struct {
		name, source, primary, cause string
	}{
		{"exclusive", string(mustPhase25Fixture(t, "exclusive_escape_reject.schway")), "exclusive", "borrow mut value"},
		{"shared", "module phase25.shared_escape\nexport { fn relay }\nfn relay(value: Buffer) -> Buffer {\n  let shared = borrow value\n  shared\n}\n", "shared", "borrow value"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "escape.schway")
			if err := os.WriteFile(path, []byte(test.source), 0600); err != nil {
				t.Fatal(err)
			}
			result, err := session.CheckCommandFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != "invalid" || len(result.Diagnostics) != 1 {
				t.Fatalf("expected one command refusal, got %+v", result)
			}
			problem := result.Diagnostics[0]
			if problem.Code != "core.origin_omitted" || problem.Schema != diagnostic.Schema || problem.Primary.Start >= problem.Primary.End || len(problem.Repairs) != 0 {
				t.Fatalf("escape diagnostic lost its code/schema/span or suggested a repair: %+v", problem)
			}
			if got := test.source[problem.Primary.Start:problem.Primary.End]; got != test.primary {
				t.Fatalf("primary span selects %q, want escaping result %q", got, test.primary)
			}
			if len(problem.Causes) == 0 || problem.Causes[0].Kind != "borrow_created_here" || problem.Causes[0].Span == nil {
				t.Fatalf("escape diagnostic lost its source-located borrow cause: %+v", problem.Causes)
			}
			span := *problem.Causes[0].Span
			if got := test.source[span.Start:span.End]; got != test.cause {
				t.Fatalf("borrow cause span selects %q, want %q", got, test.cause)
			}
		})
	}
}

func TestPhase25EscapeDiagnosticFunctionIdentity(t *testing.T) {
	source := "module phase25.escape_identity\nexport { fn relay }\nfn owned(value: Buffer) -> Buffer {\n  let result = take value\n  result\n}\nfn relay(value: Buffer) -> Buffer {\n  let result = borrow value\n  result\n}\n"
	path := filepath.Join(t.TempDir(), "identity.schway")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	result, err := session.CheckCommandFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "invalid" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "core.origin_omitted" {
		t.Fatalf("expected an escape refusal attributed to relay, got %+v", result)
	}
	primary := result.Diagnostics[0].Primary
	if got := source[primary.Start:primary.End]; got != "result" || primary.Start != strings.LastIndex(source, "result") {
		t.Fatalf("similar earlier return redirected the escape source selection to %q at %d", got, primary.Start)
	}
}

func TestPhase25IndependentPeer(t *testing.T) {
	checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase3", "sequential_shared_then_exclusive_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("ended shared loan followed by exclusive access was rejected: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("corevalidate rejected sequential families: %+v", result.Problems)
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("originvalidate rejected sequential families: %+v", problems)
	}
	if err := pathoracle.ValidateLocalOwnerPaths(checked.Program); err != nil {
		t.Fatalf("pathoracle rejected sequential families: %v", err)
	}
}

func TestPhase25IndependentPeerMutations(t *testing.T) {
	for _, positive := range []struct{ phase, fixture string }{
		{"phase3", "sequential_shared_then_exclusive_accept.schway"},
		{"phase3", "shared_shared_accept.schway"},
		{"phase25", "exclusive_copy_accept.schway"},
	} {
		t.Run("positive/"+positive.fixture, func(t *testing.T) {
			program := checkedPhase25Program(t, positive.phase, positive.fixture)
			assertPhase25Peers(t, program, "", true)
		})
	}
	for _, mutation := range []struct {
		name    string
		phase   string
		fixture string
		mutate  func(*core.Program)
	}{
		{"overlap", "phase3", "sequential_shared_then_exclusive_accept.schway", mutatePhase25Overlap},
		{"escape", "phase25", "exclusive_copy_accept.schway", mutatePhase25Escape},
	} {
		t.Run("mutation/"+mutation.name, func(t *testing.T) {
			program := checkedPhase25Program(t, mutation.phase, mutation.fixture)
			mutation.mutate(&program)
			assertPhase25Peers(t, program, mutation.name, false)
		})
	}
}

func TestPhase25CheckCommandPeerGate(t *testing.T) {
	for _, fixture := range []struct {
		name, phase, file string
		status            string
	}{
		{"ended shared then exclusive", "phase3", "sequential_shared_then_exclusive_accept.schway", "pass"},
		{"compatible readers", "phase3", "shared_shared_accept.schway", "pass"},
		{"overlap", "phase25", "exclusive_conflict_reject.schway", "invalid"},
		{"borrowed result escape", "phase25", "exclusive_escape_reject.schway", "invalid"},
	} {
		t.Run(fixture.name, func(t *testing.T) {
			result, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", fixture.phase, fixture.file))
			if err != nil {
				t.Fatal(err)
			}
			if result.Status != fixture.status {
				t.Fatalf("command gate returned %q with diagnostics %+v, want %q", result.Status, result.Diagnostics, fixture.status)
			}
			if fixture.status == "invalid" && len(result.Diagnostics) == 0 {
				t.Fatal("command gate refused without a diagnostic")
			}
		})
	}
}

func assertPhase25Peers(t *testing.T, program core.Program, mutation string, wantValid bool) {
	t.Helper()
	coreResult := corevalidate.Validate(program)
	originProblems := originvalidate.ValidatePublished(program)
	pathErr := pathoracle.ValidateLocalOwnerPaths(program)
	coreValid := coreResult.Valid
	var coreCode string
	if len(coreResult.Problems) > 0 {
		coreCode = coreResult.Problems[0].Code
	}
	if mutation == "escape" {
		function := phase25SessionFunction(&program, "exclusive_copy")
		peer, exists := coreResult.PeerSignatures()[function.ID]
		coreValid = exists && peer.Callable
		if exists && !peer.Callable {
			coreCode = "core.peer_callable_false"
		}
	}
	originValid := len(originProblems) == 0
	pathValid := pathErr == nil
	if coreValid != wantValid || originValid != wantValid || pathValid != wantValid {
		t.Fatalf("independent peers disagree with expected %v for %q: core valid=%v problems=%+v, origin valid=%v problems=%+v, path err=%v", wantValid, mutation, coreValid, coreResult.Problems, originValid, originProblems, pathErr)
	}
	if !wantValid {
		if mutation == "overlap" && len(coreResult.Problems) == 0 || mutation == "escape" && coreValid || len(originProblems) == 0 || pathErr == nil {
			t.Fatalf("each peer must report its own refusal for %q", mutation)
		}
		if mutation == "overlap" && (coreCode != "core.operation_order" && coreCode != "core.borrow_conflict" || originProblems[0].Code != "core.borrow_conflict" || !strings.HasPrefix(pathErr.Error(), "pathoracle.pointer_borrow_conflict:")) {
			t.Fatalf("overlap mutation did not produce distinct peer refusals: core=%+v origin=%+v path=%v", coreResult.Problems, originProblems, pathErr)
		}
		if mutation == "escape" && (coreCode != "core.peer_callable_false" || originProblems[0].Code != "core.origin_omitted" || !strings.HasPrefix(pathErr.Error(), "pathoracle.pointer_escape:")) {
			t.Fatalf("escape mutation did not produce its expected peer refusals: core=%+v origin=%+v path=%v", coreResult.Problems, originProblems, pathErr)
		}
	}
}

func checkedPhase25Program(t *testing.T, phase, fixture string) core.Program {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", phase, fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(data)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("accepted fixture %s had diagnostics: %+v", fixture, checked.Diagnostics)
	}
	return clonePhase25Program(checked.Program)
}

func clonePhase25Program(program core.Program) core.Program {
	program.Functions = append([]core.Function(nil), program.Functions...)
	for i := range program.Functions {
		if program.Functions[i].Linear == nil {
			continue
		}
		linear := *program.Functions[i].Linear
		linear.Types = append([]core.TypeFact(nil), linear.Types...)
		linear.Places = append([]core.Place(nil), linear.Places...)
		linear.Operations = append([]core.LinearOperation(nil), linear.Operations...)
		linear.LoanEndpoints = append([]core.LoanEndpoint(nil), linear.LoanEndpoints...)
		program.Functions[i].Linear = &linear
	}
	return program
}

func mutatePhase25Overlap(program *core.Program) {
	function := phase25SessionFunction(program, "relay")
	ops := function.Linear.Operations
	use, borrow := -1, -1
	for i, operation := range ops {
		if operation.Kind == core.OpBorrowShared && use < 0 {
			for j := i + 1; j < len(ops); j++ {
				if ops[j].SourceID == operation.TargetID {
					use = j
					break
				}
			}
		}
		if operation.Kind == core.OpBorrowExclusive {
			borrow = i
		}
	}
	if use < 0 || borrow < 0 || use >= borrow {
		panic("sequential fixture did not contain an ended shared loan before an exclusive borrow")
	}
	firstUse := ops[use]
	moved := append([]core.LinearOperation(nil), ops[:use]...)
	moved = append(moved, ops[use+1:]...)
	moved = append(moved[:borrow], append([]core.LinearOperation{firstUse}, moved[borrow:]...)...)
	for i := range moved {
		moved[i].ID = fmt.Sprintf("%s:op:%d", function.ID, i)
		moved[i].PointID = fmt.Sprintf("%s:point:linear:%d", function.ID, i)
	}
	function.Linear.Operations = moved
}

func mutatePhase25Escape(program *core.Program) {
	function := phase25SessionFunction(program, "exclusive_copy")
	ops := function.Linear.Operations
	if len(ops) != 3 || ops[0].Kind != core.OpBorrowExclusive || ops[2].Kind != core.OpReturn {
		panic("exclusive copy fixture did not contain a borrow and terminal return")
	}
	ops[2].SourceID = ops[0].TargetID
}

func phase25SessionFunction(program *core.Program, name string) *core.Function {
	for i := range program.Functions {
		if program.Functions[i].Name == name {
			return &program.Functions[i]
		}
	}
	panic("Phase 25 mutation function not found: " + name)
}

func TestPhase25UnsupportedPointerShape(t *testing.T) {
	for _, test := range []struct{ name, fixture, code string }{
		{"exclusive mutation conflicts with active loan", "exclusive_conflict_reject.schway", "ownership.borrow_conflict"},
		{"exclusive result escape stays unsupported", "exclusive_escape_reject.schway", "core.origin_omitted"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if test.fixture == "exclusive_escape_reject.schway" {
				result, err := session.CheckCommandFile(testsupport.ProjectPath("testdata", "phase25", test.fixture))
				if err != nil {
					t.Fatal(err)
				}
				if result.Status != "invalid" || len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != test.code {
					t.Fatalf("source control %q produced %+v, want %s", test.fixture, result.Diagnostics, test.code)
				}
				return
			}
			checked, err := session.CheckFile(testsupport.ProjectPath("testdata", "phase25", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != test.code {
				t.Fatalf("source control %q produced %+v, want %s", test.fixture, checked.Diagnostics, test.code)
			}
			if checked.Diagnostics[0].Primary.Start == checked.Diagnostics[0].Primary.End {
				t.Fatalf("source control %q lost its offending source span: %+v", test.fixture, checked.Diagnostics[0])
			}
		})
	}
}

func TestPhase25OwnershipDiagnosticBoundary(t *testing.T) {
	for _, test := range []struct{ name, phase, fixture, code string }{
		{"moved-from use", "phase2", "use_after_move.schway", "ownership.use_after_move"},
		{"discarded acquired owner", "phase23", "discard_owner.schway", "check.local_owner_discarded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			checked, err := session.CheckFile(testsupport.ProjectPath("testdata", test.phase, test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != test.code {
				t.Fatalf("source control %q produced %+v, want %s", test.fixture, checked.Diagnostics, test.code)
			}
			if checked.Diagnostics[0].Primary.Start == checked.Diagnostics[0].Primary.End {
				t.Fatalf("source control %q lost source attribution: %+v", test.fixture, checked.Diagnostics[0])
			}
		})
	}
}

func mustPhase25Fixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}
