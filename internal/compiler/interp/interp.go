package interp

import (
	"encoding/json"
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
)

const Schema = "lang.execution/0"

type Outcome struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

type Event struct {
	Schema     string `json:"schema"`
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	FunctionID string `json:"function_id"`
	Input      string `json:"input"`
	Output     string `json:"output"`
}

type Execution struct {
	Schema        string   `json:"schema"`
	Outcome       Outcome  `json:"outcome"`
	Events        []Event  `json:"events"`
	LiveResources []string `json:"live_resources"`
}

func Run(program core.Program, functionName, input string) (Execution, error) {
	function, ok := findFunction(program, functionName)
	if !ok {
		return Execution{}, fmt.Errorf("function %q is absent from checked core", functionName)
	}
	for _, arm := range function.Match.Arms {
		if arm.Pattern == input {
			event := Event{
				Schema: Schema, ID: arm.ID + ":event:returned", Kind: "function.returned",
				FunctionID: function.ID, Input: input, Output: arm.Value,
			}
			return Execution{
				Schema: Schema, Outcome: Outcome{Kind: "returned", Value: arm.Value},
				Events: []Event{event}, LiveResources: []string{},
			}, nil
		}
	}
	return Execution{}, fmt.Errorf("checked match %q has no arm for %q", function.Match.ID, input)
}

func CanonicalBytes(execution Execution) ([]byte, error) {
	return json.Marshal(execution)
}

func findFunction(program core.Program, name string) (core.Function, bool) {
	for _, function := range program.Functions {
		if function.Name == name {
			return function, true
		}
	}
	return core.Function{}, false
}
