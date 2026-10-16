package executionpeer_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
)

// Terminal attribution is structural evidence, independent of numeric replay.
func TestPhase33TerminalAttribution(t *testing.T) {
	for _, terminal := range []struct {
		kind                   core.OperationKind
		event, suffix, outcome string
	}{
		{core.OpReturn, "function.returned", "returned", execution.OutcomeReturned},
		{core.OpFail, "function.failed", "failed", execution.OutcomeTypedFailure},
		{core.OpDefect, "function.defected", "defected", execution.OutcomeDefect},
	} {
		t.Run(string(terminal.kind), func(t *testing.T) {
			op := core.LinearOperation{ID: "entry:terminal", Kind: terminal.kind, SourceID: "entry:value", TypeID: "value", Reason: "expected reason"}
			p := core.Program{Functions: []core.Function{function("entry", op)}}
			e := event(inv(t, "entry"), op.ID+":event:"+terminal.suffix, terminal.event, "entry")
			e.SourcePlace, e.TypeID = op.SourceID, op.TypeID
			if terminal.kind == core.OpFail {
				// This synthetic terminal uses a constructor named like its local
				// type ID; the checked Phase24 tracer separately proves non-identity
				// TypeFact-ID-to-constructor mapping.
				p.Functions[0].Linear.Types = []core.TypeFact{{ID: op.TypeID, Shape: core.TypeRef{Constructor: op.TypeID}}}
				p.Functions[0].Linear.Places = []core.Place{{ID: op.SourceID, TypeID: op.TypeID}}
			}
			if terminal.kind == core.OpDefect {
				e.Output = op.Reason
			}
			clean := execution.Execution{Schema: execution.Schema2, Outcome: execution.Outcome{Kind: terminal.outcome}, Events: []execution.Event{e}}
			if err := executionpeer.Validate(p, clean); err != nil {
				t.Fatalf("clean terminal: %v", err)
			}
			controls := []struct {
				name   string
				mutate func(*execution.Execution)
			}{
				{"wrong source", func(d *execution.Execution) { d.Events[0].SourcePlace = "forged" }},
				{"wrong type", func(d *execution.Execution) { d.Events[0].TypeID = "forged" }},
				{"target on terminal", func(d *execution.Execution) { d.Events[0].TargetPlace = "forged" }},
				{"missing outcome", func(d *execution.Execution) { d.Outcome.Kind = "" }},
				{"unknown outcome", func(d *execution.Execution) { d.Outcome.Kind = "forged" }},
				{"event after root terminal", func(d *execution.Execution) {
					repeated := e
					repeated.Occurrence++
					d.Events = append(d.Events, repeated)
				}},
			}
			for _, outcome := range execution.TerminalOutcomeKinds() {
				if outcome != terminal.outcome {
					controls = append(controls, struct {
						name   string
						mutate func(*execution.Execution)
					}{"outcome " + outcome, func(d *execution.Execution) { d.Outcome.Kind = outcome }})
				}
			}
			if terminal.kind == core.OpDefect {
				controls = append(controls, struct {
					name   string
					mutate func(*execution.Execution)
				}{"wrong reason", func(d *execution.Execution) { d.Events[0].Output = "forged" }})
				controls = append(controls, struct {
					name   string
					mutate func(*execution.Execution)
				}{"defect value", func(d *execution.Execution) { d.Outcome.Value = "forged" }})
			} else {
				// The peer must continue to accept a structurally plausible wrong value.
				d := clean
				d.Outcome.Value = "999"
				if err := executionpeer.Validate(p, d); err != nil {
					t.Fatalf("numeric replay crossed peer boundary: %v", err)
				}
			}
			for _, control := range controls {
				t.Run(control.name, func(t *testing.T) {
					d := clean
					d.Events = append([]execution.Event(nil), clean.Events...)
					control.mutate(&d)
					if err := executionpeer.Validate(p, d); err == nil {
						t.Fatal("forged terminal accepted")
					}
				})
			}
		})
	}
}
