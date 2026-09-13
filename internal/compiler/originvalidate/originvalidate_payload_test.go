package originvalidate_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/originvalidate"
)

// payloadDestructureThroughBorrowFunction builds, by hand -- directly at
// core-IR level, bypassing check/parser entirely, matching this package's
// own synthetic-program test style -- a function whose body borrows its
// parameter, destructures a payload out of the borrowed alias, and returns
// the destructured payload unchanged. This is Task 1's own darkest-corner
// shape: a payload whose origin traces back through an OpDestructurePayload
// to an original OpBorrowShared. function.PublicOrigin declares exactly the
// origin the body derives, so a correctly-derived walk must publish clean.
func payloadDestructureThroughBorrowFunction() core.Function {
	functionID := "s1:test:fn:payload_relay"
	typeID := functionID + ":type:0"
	parameterID := functionID + ":place:0"
	viewID := functionID + ":place:1"
	payloadID := functionID + ":place:2"
	return core.Function{
		ID: functionID, Name: "payload_relay",
		EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
		Parameter:  core.Parameter{ID: parameterID, Name: "buffer", Type: "Outcome"},
		ReturnType: "Buffer",
		PublicOrigin: &core.PublicOrigin{
			Paths: []string{"buffer"}, Access: "shared",
		},
		Linear: &core.LinearBody{
			ID: functionID + ":linear",
			Places: []core.Place{
				{ID: parameterID, Name: "buffer", TypeID: typeID},
				{ID: viewID, Name: "view", TypeID: typeID},
				{ID: payloadID, Name: "payload", TypeID: typeID},
			},
			Operations: []core.LinearOperation{
				{ID: functionID + ":op:0", PointID: functionID + ":point:linear:0", Kind: core.OpBorrowShared, SourceID: parameterID, TargetID: viewID, LoanID: functionID + ":loan:0", TypeID: typeID},
				{ID: functionID + ":op:1", PointID: functionID + ":point:linear:1", Kind: core.OpDestructurePayload, SourceID: viewID, PayloadTargetID: payloadID, TypeID: typeID},
				{ID: functionID + ":op:2", PointID: functionID + ":point:linear:2", Kind: core.OpReturn, SourceID: payloadID, TypeID: typeID},
			},
		},
	}
}

// TestOriginWalksThroughPayloadDestructureToBorrow is Task 1's own criterion:
// a payload whose origin traces back through a destructure to an original
// borrow is walked correctly by originvalidate, so a declared PublicOrigin
// covering that borrow is not reported as understated (D-12-28's guard is
// not tripped -- this walks a NEW kind's own semantics, not core.OpCall's
// gap).
func TestOriginWalksThroughPayloadDestructureToBorrow(t *testing.T) {
	function := payloadDestructureThroughBorrowFunction()
	problems := originvalidate.PublishProblemsFor(function, nil)
	if len(problems) != 0 {
		t.Fatalf("expected the payload's origin to be traced through the destructure to the borrow with no problems, got %+v", problems)
	}

	paths, access, ok := originvalidate.RecomputeOrigin(function, nil)
	if !ok {
		t.Fatalf("expected RecomputeOrigin to derive an origin through the payload destructure")
	}
	if access != "shared" {
		t.Fatalf("expected access shared (from the OpBorrowShared hop), got %q", access)
	}
	if len(paths) != 1 || paths[0] != "buffer" {
		t.Fatalf("expected origin path [buffer], got %+v", paths)
	}
}

// TestOriginUnderstatedAcrossPayloadDestructure proves the walk is not
// merely permissive: an UNDER-declared origin (empty Paths) covering a
// borrow-derived payload is still refused as understated -- the new arm
// must not make the refusal a no-op.
func TestOriginUnderstatedAcrossPayloadDestructure(t *testing.T) {
	function := payloadDestructureThroughBorrowFunction()
	function.PublicOrigin = &core.PublicOrigin{Paths: []string{}, Access: "shared"}
	problems := originvalidate.PublishProblemsFor(function, nil)
	if len(problems) != 1 || problems[0].Code != "core.origin_understated" {
		t.Fatalf("expected exactly core.origin_understated, got %+v", problems)
	}
}
