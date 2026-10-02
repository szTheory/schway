package originvalidate_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase25OriginProgram(t *testing.T, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	return checked.Program
}

func TestPhase25PointerOrigin(t *testing.T) {
	for _, fixture := range []string{"shared_shared_accept.schway", "sequential_shared_then_exclusive_accept.schway"} {
		source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", fixture))
		if err != nil {
			t.Fatal(err)
		}
		checked := session.Check(source)
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
		}
		if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
			t.Errorf("independent origin peer rejected %s: %+v", fixture, problems)
		}
	}

	program := phase25OriginProgram(t, "shared_copy_accept.schway")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"value"}, Access: "shared"}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("independent origin peer rejected shared read/copy: %+v", problems)
	}
	omitted := phase25OriginProgram(t, "shared_copy_accept.schway")
	problems := originvalidate.ValidatePublished(omitted)
	if len(problems) == 0 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("origin peer accepted an undeclared returned loan: %+v", problems)
	}

	function := program.Functions[0]
	origins := originvalidate.RecomputeOriginPerReturn(function, originvalidate.BuildCalleeOriginFacts(program))
	if len(origins) != 1 || !origins[0].Derived || origins[0].Access != "shared" || len(origins[0].Paths) != 1 || origins[0].Paths[0] != "value" {
		t.Fatalf("origin peer did not independently derive the returned shared family: %+v", origins)
	}

	mutated := phase25OriginProgram(t, "shared_copy_accept.schway")
	mutated.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"value"}, Access: "shared"}
	mutated.Functions[0].Linear.Operations[0].Kind = core.OpBorrowExclusive
	problems = originvalidate.ValidatePublished(mutated)
	if len(problems) == 0 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("origin peer did not reject family/access mutation: %+v", problems)
	}

	conflictingSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "shared_shared_accept.schway"))
	if err != nil {
		t.Fatal(err)
	}
	conflicting := session.Check(conflictingSource)
	if len(conflicting.Diagnostics) != 0 {
		t.Fatalf("conflict fixture: unexpected diagnostics: %+v", conflicting.Diagnostics)
	}
	operations := conflicting.Program.Functions[0].Linear.Operations
	sharedIndex := -1
	for i := range operations {
		if operations[i].Kind == core.OpBorrowShared {
			if sharedIndex >= 0 {
				operations[i].Kind = core.OpBorrowExclusive
				break
			}
			sharedIndex = i
		}
	}
	problems = originvalidate.ValidatePublished(conflicting.Program)
	if len(problems) == 0 || problems[0].Code != "core.borrow_conflict" {
		t.Fatalf("origin peer accepted a live incompatible access: %+v", problems)
	}
}

func TestPhase25U64CopyOrigin(t *testing.T) {
	program := phase25OriginProgram(t, "exclusive_copy_accept.schway")
	function := program.Functions[0]
	origins := originvalidate.RecomputeOriginPerReturn(function, originvalidate.BuildCalleeOriginFacts(program))
	if len(origins) != 1 || origins[0].Derived {
		t.Fatalf("origin peer retained borrow provenance through a U64 copy: %+v", origins)
	}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("origin peer refused an ordinary copied U64 result: %+v", problems)
	}

	conflicting := phase25OriginProgram(t, "exclusive_copy_accept.schway")
	function = conflicting.Functions[0]
	linear := function.Linear
	if len(linear.Operations) != 3 || linear.Operations[0].Kind != core.OpBorrowExclusive || linear.Operations[1].Kind != core.OpCopy || linear.Operations[2].Kind != core.OpReturn {
		t.Fatalf("exclusive-copy fixture has unexpected operations: %+v", linear.Operations)
	}
	borrowedID := linear.Operations[0].TargetID
	borrowType := linear.Operations[0].TypeID
	linear.Places[2].Name = "overlap"
	linear.Places = append(linear.Places,
		core.Place{ID: function.ID + ":place:3", Name: "copied", TypeID: borrowType},
		core.Place{ID: function.ID + ":place:4", Name: "overlap_copy", TypeID: borrowType},
	)
	linear.Operations = []core.LinearOperation{
		linear.Operations[0],
		{ID: function.ID + ":op:1", PointID: function.ID + ":point:linear:1", Kind: core.OpBorrowShared, SourceID: function.Parameter.ID, TargetID: function.ID + ":place:2", TypeID: borrowType, LoanID: "phase25:test:overlap"},
		{ID: function.ID + ":op:2", PointID: function.ID + ":point:linear:2", Kind: core.OpCopy, SourceID: borrowedID, TargetID: function.ID + ":place:3", TypeID: borrowType},
		{ID: function.ID + ":op:3", PointID: function.ID + ":point:linear:3", Kind: core.OpCopy, SourceID: function.ID + ":place:2", TargetID: function.ID + ":place:4", TypeID: borrowType},
		{ID: function.ID + ":op:4", PointID: function.ID + ":point:linear:4", Kind: core.OpReturn, SourceID: function.ID + ":place:3", TypeID: borrowType},
	}
	problems := originvalidate.ValidatePublished(conflicting)
	if len(problems) == 0 || problems[0].Code != "core.borrow_conflict" {
		t.Fatalf("origin peer did not keep the exclusive loan live through its U64 copy: %+v", problems)
	}
}

func TestPhase25UtilityOwnerTransfer(t *testing.T) {
	program := phase25UtilityOwnerProgram(t)
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("independent origin peer rejected exact Phase 25 owner/result flow: %+v", problems)
	}
	for _, test := range []struct {
		name   string
		mutate func(*core.Program)
	}{
		{"swapped helpers", func(program *core.Program) {
			main := phase25UtilityOwnerFunction(program, "main")
			sharedID, exclusiveID := main.Linear.Operations[2].CalleeID, main.Linear.Operations[3].CalleeID
			main.Linear.Operations[2].CalleeID, main.Linear.Operations[3].CalleeID = exclusiveID, sharedID
		}},
		{"arbitrary helper", func(program *core.Program) {
			phase25UtilityOwnerFunction(program, "main").Linear.Operations[2].CalleeID = "untrusted:helper"
		}},
		{"missing exclusive call", func(program *core.Program) {
			main := phase25UtilityOwnerFunction(program, "main")
			main.Linear.Operations = append(main.Linear.Operations[:3], main.Linear.Operations[4:]...)
		}},
		{"tampered return", func(program *core.Program) {
			main := phase25UtilityOwnerFunction(program, "main")
			main.Linear.Operations[5].SourceID = main.Linear.Operations[2].TargetID
		}},
		{"exclusive argument bypasses shared", func(program *core.Program) {
			main := phase25UtilityOwnerFunction(program, "main")
			main.Linear.Operations[3].SourceID = main.Linear.Operations[1].TargetID
		}},
		{"reordered helpers", func(program *core.Program) {
			main := phase25UtilityOwnerFunction(program, "main")
			main.Linear.Operations[2], main.Linear.Operations[3] = main.Linear.Operations[3], main.Linear.Operations[2]
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			mutated := phase25CloneUtilityOwnerProgram(phase25UtilityOwnerProgram(t))
			test.mutate(&mutated)
			if problems := originvalidate.ValidatePublished(mutated); len(problems) == 0 {
				t.Fatal("origin peer accepted a non-exact Phase 25 owner/result chain")
			}
		})
	}
}

func TestPhase25PreservesExactPhase24DirectOwnerRoute(t *testing.T) {
	direct := phase25DirectUtilityOwnerProgram(t)
	main := phase25UtilityOwnerFunction(&direct, "main")
	if len(direct.Functions) != 2 || main.Linear == nil || len(main.Linear.Operations) != 4 {
		t.Fatalf("direct-route fixture is not the exact Phase 24 shape: functions=%d operations=%d", len(direct.Functions), len(main.Linear.Operations))
	}
	if problems := originvalidate.ValidatePublished(direct); len(problems) != 0 {
		t.Fatalf("independent origin peer rejected the exact direct Phase 24 route: %+v", problems)
	}

	mutated := phase25CloneUtilityOwnerProgram(phase25UtilityOwnerProgram(t))
	main = phase25UtilityOwnerFunction(&mutated, "main")
	main.Linear.Operations[5].SourceID = main.Linear.Operations[1].TargetID
	if problems := originvalidate.ValidatePublished(mutated); len(problems) == 0 {
		t.Fatal("origin peer accepted a Phase 25 chain whose return was tampered back to the borrowed-use result")
	}
}

func phase25UtilityOwnerProgram(t *testing.T) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "transfer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("integrated owner-transfer source diagnostics: %+v", checked.Diagnostics)
	}
	return phase25CloneUtilityOwnerProgram(checked.Program)
}

func phase25CloneUtilityOwnerProgram(program core.Program) core.Program {
	program.Functions = append([]core.Function(nil), program.Functions...)
	for index := range program.Functions {
		if program.Functions[index].Linear == nil {
			continue
		}
		linear := *program.Functions[index].Linear
		linear.Types = append([]core.TypeFact(nil), linear.Types...)
		linear.Places = append([]core.Place(nil), linear.Places...)
		linear.Operations = append([]core.LinearOperation(nil), linear.Operations...)
		linear.Blocks = append([]core.Block(nil), linear.Blocks...)
		linear.Edges = append([]core.Edge(nil), linear.Edges...)
		program.Functions[index].Linear = &linear
	}
	return program
}

func phase25UtilityOwnerFunction(program *core.Program, name string) *core.Function {
	for index := range program.Functions {
		if program.Functions[index].Name == name {
			return &program.Functions[index]
		}
	}
	panic("Phase 25 owner-transfer fixture function not found: " + name)
}

func phase25DirectUtilityOwnerProgram(t *testing.T) core.Program {
	t.Helper()
	program := phase25UtilityOwnerProgram(t)
	main := phase25UtilityOwnerFunction(&program, "main")
	ops := main.Linear.Operations
	main.Linear.Operations = []core.LinearOperation{ops[0], ops[1], ops[4], ops[5]}
	main.Linear.Operations[3].SourceID = ops[1].TargetID
	for index := range main.Linear.Operations {
		main.Linear.Operations[index].ID = main.ID + ":op:" + strconv.Itoa(index)
		main.Linear.Operations[index].PointID = main.ID + ":point:linear:" + strconv.Itoa(index)
	}
	main.Linear.Places = append([]core.Place(nil), main.Linear.Places[:3]...)
	filtered := make([]core.Function, 0, 2)
	for _, function := range program.Functions {
		if function.Name == "main" || function.Name == "acquire" {
			filtered = append(filtered, function)
		}
	}
	program.Functions = filtered
	return program
}
