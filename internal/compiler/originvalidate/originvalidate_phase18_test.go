package originvalidate_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase18ComputedPayloadFunction(returnInSiblingArm bool) core.Function {
	id := "s1:test:fn:computed_payload"
	parameterID := id + ":place:0"
	borrowedID := id + ":place:1"
	computedID := id + ":place:2"
	payloadID := id + ":place:3"
	armB := id + ":block:arm:1"
	returnBlock := id + ":block:arm:0"
	armAOperations := []string{id + ":op:2"}
	armBOperations := []string{}
	if returnInSiblingArm {
		returnBlock = armB
		armBOperations = append(armBOperations, id+":op:3")
	} else {
		armAOperations = append(armAOperations, id+":op:3")
	}
	return core.Function{
		ID: id, Name: "computed_payload",
		EntryPointID: id + ":point:entry", ReturnPointID: id + ":point:return",
		Parameter:    core.Parameter{ID: parameterID, Name: "input", Type: "Outcome"},
		ReturnType:   "Buffer",
		PublicOrigin: &core.PublicOrigin{Paths: []string{"input"}, Access: "shared"},
		Match:        &core.Match{},
		Linear: &core.LinearBody{
			ID: id + ":linear",
			Places: []core.Place{
				{ID: parameterID, Name: "input", TypeID: id + ":type:0"},
				{ID: borrowedID, Name: "borrowed", TypeID: id + ":type:0"},
				{ID: computedID, Name: "computed", TypeID: id + ":type:0"},
				{ID: payloadID, Name: "payload", TypeID: id + ":type:1"},
			},
			Operations: []core.LinearOperation{
				{ID: id + ":op:0", Kind: core.OpBorrowShared, SourceID: parameterID, TargetID: borrowedID, LoanID: id + ":loan:0"},
				{ID: id + ":op:1", Kind: core.OpMove, SourceID: borrowedID, TargetID: computedID},
				{ID: id + ":op:2", Kind: core.OpDestructurePayload, SourceID: computedID, PayloadTargetID: payloadID},
				{ID: id + ":op:3", Kind: core.OpReturn, SourceID: payloadID},
			},
			Blocks: []core.Block{
				{ID: id + ":block:entry", OperationIDs: []string{id + ":op:0", id + ":op:1"}, Successors: []string{returnBlock, armB}},
				{ID: id + ":block:arm:0", OperationIDs: armAOperations},
				{ID: armB, OperationIDs: armBOperations},
			},
		},
	}
}

func TestPhase18ComputedPlacePayloadOriginIsDerivedLocally(t *testing.T) {
	function := phase18ComputedPayloadFunction(false)
	paths, access, ok := originvalidate.RecomputeOrigin(function, nil)
	if !ok || access != "shared" || len(paths) != 1 || paths[0] != "input" {
		t.Fatalf("computed payload origin = (%v, %q, %v), want ([input], shared, true)", paths, access, ok)
	}
	if problems := originvalidate.PublishProblemsFor(function, nil); len(problems) != 0 {
		t.Fatalf("valid computed payload origin rejected: %+v", problems)
	}

	function.PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.PublishProblemsFor(function, nil)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("understated computed payload origin problems = %+v", problems)
	}
}

func TestPhase18ComputedPayloadCannotBorrowSiblingArmOrigin(t *testing.T) {
	function := phase18ComputedPayloadFunction(true)
	function.Match = nil // Block ownership, not a producer's Match summary, must bound this peer's walk.
	if _, _, ok := originvalidate.RecomputeOrigin(function, nil); ok {
		t.Fatal("payload produced in another arm was accepted as this return's origin")
	}
	problems := originvalidate.PublishProblemsFor(function, nil)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("mismatched arm payload origin problems = %+v", problems)
	}
}

func TestPhase18OriginPeerAcceptsComputedSource(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "computed_match.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("computed payload source was refused: %+v", checked.Diagnostics)
	}
	var found bool
	for _, function := range checked.Program.Functions {
		if function.Name == "select" {
			found = true
			if function.Match == nil || function.Linear == nil || function.Match.ScrutineeID == "" {
				t.Fatalf("source did not produce a computed-place match: %+v", function)
			}
		}
	}
	if !found {
		t.Fatal("source did not contain select")
	}
	if problems := originvalidate.ValidatePublished(checked.Program); len(problems) != 0 {
		t.Fatalf("origin peer rejected valid computed match source: %+v", problems)
	}
}

// Checker admission does not bless a later forged origin summary: this test
// first admits the source, then changes only the published contract and asks
// the origin peer to recompute it from the core facts.
func TestPhase18OriginPeerRejectsForgedComputedOriginAfterCheckerAdmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", "public_view.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("checker refused control source: %+v", checked.Diagnostics)
	}
	var target *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].PublicOrigin != nil {
			target = &checked.Program.Functions[index]
			break
		}
	}
	if target == nil {
		t.Fatal("checker-admitted control source did not publish an origin")
	}
	target.PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.ValidatePublished(checked.Program)
	if len(problems) == 0 {
		t.Fatal("origin peer accepted forged empty origin after checker admission")
	}
}
