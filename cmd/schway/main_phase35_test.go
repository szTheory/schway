package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase35PublicConformanceWitness(t *testing.T) {
	path := testsupport.ProjectPath("examples", "bitwise.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read tracked Phase 35 witness: %v", err)
	}
	operationIDs := phase35PublicPureOperationIDs(t, source)
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine+"/locked-264", func(t *testing.T) {
			phase35RunPublicCLI(t, engine, path, "264", operationIDs)
		})
	}

	const locked = "x ^ (y << 3) & 0xFF"
	const alternate = "(x ^ (y << 3)) & 0xFF"
	alternateSource := strings.Replace(string(source), locked, alternate, 1)
	if alternateSource == string(source) {
		t.Fatal("tracked witness did not contain the locked association expression")
	}
	alternatePath := filepath.Join(t.TempDir(), "alternate.schway")
	if err := os.WriteFile(alternatePath, []byte(alternateSource), 0o600); err != nil {
		t.Fatalf("write alternate-association source: %v", err)
	}
	alternateIDs := phase35PublicPureOperationIDs(t, []byte(alternateSource))
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine+"/alternate-8", func(t *testing.T) {
			phase35RunPublicCLI(t, engine, alternatePath, "8", alternateIDs)
		})
	}
}

type phase35PublicCommand struct {
	Schema     string                `json:"schema"`
	Command    string                `json:"command"`
	Status     string                `json:"status"`
	Executions []execution.Execution `json:"executions"`
}

func phase35RunPublicCLI(t *testing.T, engine, path, want string, pureOperationIDs []string) {
	t.Helper()
	code, stdout, stderr := captureLangRun(t, []string{"--json", "run", "--engine=" + engine, path})
	if code != exitSuccess || stderr != "" {
		t.Fatalf("public %s run code=%d stderr=%q stdout=%q", engine, code, stderr, stdout)
	}
	var command phase35PublicCommand
	if err := json.Unmarshal([]byte(stdout), &command); err != nil {
		t.Fatalf("decode public %s command result: %v; output=%q", engine, err, stdout)
	}
	if command.Command != "run" || command.Status != "pass" || len(command.Executions) != 1 {
		t.Fatalf("public %s command schema=%q command=%q status=%q executions=%d", engine, command.Schema, command.Command, command.Status, len(command.Executions))
	}
	got := command.Executions[0]
	if got.Outcome.Kind != execution.OutcomeReturned || got.Outcome.Value != want {
		t.Fatalf("public %s projection outcome=%+v, want returned value %q", engine, got.Outcome, want)
	}
	returns := 0
	for _, event := range got.Events {
		if event.Kind == "function.returned" {
			returns++
		}
		if event.Kind == "function.defected" {
			t.Fatalf("public %s pure-operation witness produced a defect event: %+v", engine, event)
		}
		for _, operationID := range pureOperationIDs {
			if strings.HasPrefix(event.ID, operationID+":event") {
				t.Fatalf("public %s pure operation %q emitted an event: %+v", engine, operationID, event)
			}
		}
	}
	if returns != 1 || len(got.Events) == 0 || got.Events[len(got.Events)-1].Kind != "function.returned" {
		t.Fatalf("public %s events=%+v, want the existing terminal return evidence", engine, got.Events)
	}
	if len(got.LiveResources) != 0 {
		t.Fatalf("public %s witness left live resources: %v", engine, got.LiveResources)
	}
}

func phase35PublicPureOperationIDs(t *testing.T, source []byte) []string {
	t.Helper()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 35 public witness failed checking: %+v", checked.Diagnostics)
	}
	seen := map[core.OperationKind]bool{}
	var ids []string
	for _, function := range checked.Program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			switch operation.Kind {
			case core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64, core.OpShiftLeftU64, core.OpShiftRightU64:
				seen[operation.Kind] = true
				ids = append(ids, operation.ID)
			}
		}
	}
	for _, kind := range []core.OperationKind{core.OpBitXorU64, core.OpBitAndU64, core.OpBitOrU64, core.OpShiftLeftU64, core.OpShiftRightU64} {
		if !seen[kind] {
			t.Fatalf("public witness did not lower the required operation kind %s", kind)
		}
	}
	return ids
}
