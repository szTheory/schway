package protocol_test

import (
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

func TestOwnershipProjectionIdentityParity(t *testing.T) {
	result := protocol.New("run", protocol.StatusInvalid)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs("ownership.use_after_move", diagnostic.Span{Start: 1, End: 2}, "moved value", nil, diagnostic.Repair{Kind: "use_transfer_target"})}
	result.Executions = []execution.Execution{{Schema: execution.Schema1, Outcome: execution.Outcome{Kind: "returned", Value: "01020304"}, Events: []execution.Event{{Schema: execution.Schema1, ID: "owned:event:0", Kind: "value.transferred", FunctionID: "owned:fn", SourcePlace: "owned:p0", TargetPlace: "owned:p1", TypeID: "owned:t0"}}, LiveResources: []string{}}}
	machine := result.Finalize()
	human := protocol.Human(result)
	if machine.Schema != "lang.command/0" || !strings.Contains(human, machine.ID) || !strings.Contains(human, machine.Diagnostics[0].ID) || !strings.Contains(human, machine.Executions[0].Events[0].ID) {
		t.Fatalf("human/JSON ownership identities diverged: result=%+v human=%q", machine, human)
	}
	for _, want := range []string{"source_place=owned:p0", "target_place=owned:p1", "type_id=owned:t0"} {
		if !strings.Contains(human, want) { t.Fatalf("human ownership event omitted %q: %q", want, human) }
	}
}

func TestResultIdentitySemanticSensitivity(t *testing.T) {
	base := protocol.New("run", protocol.StatusPass)
	base.Executions = []interp.Execution{{
		Schema:        interp.Schema,
		Outcome:       interp.Outcome{Kind: "returned", Value: "On"},
		Events:        []interp.Event{{Schema: interp.Schema, ID: "event:one", Kind: "function.returned", Input: "Off", Output: "On"}},
		LiveResources: []string{},
	}}
	first := base.Finalize()

	metricsChanged := base
	metricsChanged.Metrics.ElapsedNS = 999
	if got := metricsChanged.Finalize().ID; got != first.ID {
		t.Fatalf("operational metrics changed semantic result ID: first=%s got=%s", first.ID, got)
	}

	outcomeChanged := base
	outcomeChanged.Executions = append([]interp.Execution(nil), base.Executions...)
	outcomeChanged.Executions[0].Outcome.Value = "Off"
	if got := outcomeChanged.Finalize().ID; got == first.ID {
		t.Fatalf("semantic outcome change retained result ID %s", got)
	}
}
