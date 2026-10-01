package originvalidate_test

import (
	"os"
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
