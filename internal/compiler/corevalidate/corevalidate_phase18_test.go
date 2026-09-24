package corevalidate_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
)

func TestPhase18ResultArmValuePlace(t *testing.T) {
	path := filepath.Join("..", "..", "..", "testdata", "phase18", "result_computed_match.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check: %+v", checked.Diagnostics)
	}
	if got := corevalidate.Validate(checked.Program); !got.Valid {
		t.Fatalf("valid typed arm rejected: %+v", got.Problems)
	}
	clone := func() core.Program {
		data, err := json.Marshal(checked.Program)
		if err != nil {
			t.Fatal(err)
		}
		var program core.Program
		if err := json.Unmarshal(data, &program); err != nil {
			t.Fatal(err)
		}
		return program
	}
	callee := func(program *core.Program) *core.Function {
		for i := range program.Functions {
			if program.Functions[i].Name == "produce" {
				return &program.Functions[i]
			}
		}
		t.Fatal("produce absent")
		return nil
	}
	mutations := []struct {
		name  string
		apply func(*core.Function)
	}{
		{"source", func(fn *core.Function) {
			for i := range fn.Linear.Operations {
				if fn.Linear.Operations[i].Kind == core.OpReturn {
					fn.Linear.Operations[i].SourceID = fn.Parameter.ID
				}
			}
		}},
		{"alternative", func(fn *core.Function) { fn.Match.Arms[0].Value = "Raw" }},
		{"place", func(fn *core.Function) { fn.Match.Arms[0].ValuePlaceID = "missing" }},
		{"type", func(fn *core.Function) {
			for i := range fn.Linear.Places {
				if fn.Linear.Places[i].ID == fn.Match.Arms[0].ValuePlaceID {
					fn.Linear.Places[i].TypeID = fn.Parameter.Type
				}
			}
		}},
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			program := clone()
			mutation.apply(callee(&program))
			if got := corevalidate.Validate(program); got.Valid {
				t.Fatal("forged arm return accepted")
			}
		})
	}
}

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

func TestPhase18ComputedScrutineeRejectsWrongType(t *testing.T) {
	checked := session.Check([]byte(`module phase18.core_peer_wrong_type
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
	function := &program.Functions[0]
	var computedID string
	for _, place := range function.Linear.Places {
		if place.Name == "computed" {
			computedID = place.ID
			break
		}
	}
	for index := range function.Linear.Places {
		if function.Linear.Places[index].ID == computedID {
			function.Linear.Places[index].TypeID = "forged-non-data-type"
		}
	}
	function.Linear.Types = append(function.Linear.Types, core.TypeFact{
		ID: "forged-non-data-type", Shape: core.TypeRef{Constructor: "Byte"},
	})
	if result := corevalidate.Validate(program); result.Valid {
		t.Fatal("computed scrutinee whose place claims a non-data type was accepted")
	}
}
