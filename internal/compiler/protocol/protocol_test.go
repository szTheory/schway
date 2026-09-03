package protocol_test

import (
	"testing"

	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/protocol"
)

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
