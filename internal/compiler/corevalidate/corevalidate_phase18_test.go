package corevalidate_test

import (
	"encoding/json"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
)

// These are source-to-core controls. They deliberately use only the public
// session entry point and corevalidate's admission API; no checker derivation
// helper supplies the peer's answer.
func TestPhase18ComputedScrutineeMustComeFromEntryPrefix(t *testing.T) {
	checked := session.Check([]byte(`module phase18.core_peer
export { type Choice fn select }
data Choice = | Left | Right
fn select(input: Choice) -> Choice {
  let computed = input
  match computed {
    Left => {
      let held = take computed
      held
    }
    Right => {
      let held = take computed
      held
    }
  }
}`))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source diagnostics: %+v", checked.Diagnostics)
	}
	if result := corevalidate.Validate(checked.Program); !result.Valid {
		t.Fatalf("valid computed-place core rejected: %+v", result.Problems)
	}

	encoded, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	var forged = checked.Program
	if err := json.Unmarshal(encoded, &forged); err != nil {
		t.Fatal(err)
	}
	function := &forged.Functions[0]
	var armPlace string
	var armName string
	for _, place := range function.Linear.Places {
		if place.Name == "held" {
			armPlace, armName = place.ID, place.Name
			break
		}
	}
	if armPlace == "" {
		t.Fatal("fixture did not produce an arm-local place")
	}
	function.Match.ScrutineeID = armPlace
	function.Match.Scrutinee = armName
	if result := corevalidate.Validate(forged); result.Valid {
		t.Fatal("arm-local place unavailable before dispatch was accepted as the computed scrutinee")
	}
}

func TestPhase18ComputedScrutineeRejectsForgedPlaceID(t *testing.T) {
	checked := session.Check([]byte(`module phase18.core_peer_forged
export { type Choice fn select }
data Choice = | Left | Right
fn select(input: Choice) -> Choice {
  let computed = input
  match computed {
    Left => Left
    Right => Right
  }
}`))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source diagnostics: %+v", checked.Diagnostics)
	}
	program := checked.Program
	program.Functions[0].Match.ScrutineeID = "forged-place"
	if result := corevalidate.Validate(program); result.Valid {
		t.Fatal("forged scrutinee place ID was accepted")
	}
}
