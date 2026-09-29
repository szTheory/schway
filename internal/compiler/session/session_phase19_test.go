package session_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase19FourTierLiteral(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, spelling, want string }{
		{"decimal", "42", "42"},
		{"zero", "0", "0"},
		{"maximum", "18446744073709551615", "18446744073709551615"},
		{"hexadecimal", "0x2A", "42"},
		{"binary_with_separator", "0b10_1010", "42"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			variant := strings.Replace(string(source), "let count = 42", "let count = "+tc.spelling, 1)
			checked := session.Check([]byte(variant))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("literal %q refused: %+v", tc.spelling, checked.Diagnostics)
			}
			engines := phase11RunFourTiers(t, context.Background(), checked.Program, "main", "7")
			engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
			if err := phase19CheckExpectedAndCompare("phase19-literal-"+tc.name, checked.Program, engines, tc.want); err != nil {
				t.Fatalf("four-tier result check failed: %v", err)
			}
		})
	}
}

func TestPhase19WrongResultControl(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("literal fixture refused: %+v", checked.Diagnostics)
	}
	program, entryName := checked.Program, "main"
	engines := phase11RunFourTiers(t, context.Background(), program, entryName, "7")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, program, engines["interpreter"])
	if err := phase19CheckExpectedAndCompare("phase19-literal-control", program, engines, "42"); err != nil {
		t.Fatalf("clean four-tier result: %v", err)
	}
	wrong := make(map[string]execution.Execution, len(engines))
	for name, result := range engines {
		result.Outcome.Value = "43"
		wrong[name] = result
	}
	// All tiers can agree on the same wrong answer; explicit observation of
	// the source's expected value must still reject that shared mutation.
	if err := session.Phase5CompareProgramEngines("phase19-shared-wrong-result", program, wrong); err != nil {
		t.Fatalf("shared wrong-result control should show why equality alone is insufficient: %v", err)
	}
	if err := phase19CheckExpectedAndCompare("phase19-shared-wrong-result", program, wrong, "42"); err == nil {
		t.Fatal("shared wrong-result mutation passed the explicit expected-value assertion")
	}
}

func phase19CheckExpectedAndCompare(fixture string, program core.Program, engines map[string]execution.Execution, expected string) error {
	for name, result := range engines {
		if result.Outcome.Value != expected {
			return fmt.Errorf("%s returned %q, want %q", name, result.Outcome.Value, expected)
		}
	}
	return session.Phase5CompareProgramEngines(fixture, program, engines)
}

func TestPhase19LiteralRun(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway")
	interpreted, err := session.RunInterpreterCommandFile(path)
	if err != nil || interpreted.Status != "pass" || len(interpreted.Executions) != 1 {
		t.Fatalf("public interpreter run: status=%q executions=%d err=%v diagnostics=%+v", interpreted.Status, len(interpreted.Executions), err, interpreted.Diagnostics)
	}
	nativeResult, err := session.RunNativeCommandFile(context.Background(), path, native.DefaultRunner())
	if err != nil || nativeResult.Status != "pass" || len(nativeResult.Executions) != 1 {
		t.Fatalf("public native run: status=%q executions=%d err=%v diagnostics=%+v", nativeResult.Status, len(nativeResult.Executions), err, nativeResult.Diagnostics)
	}
	for name, result := range map[string]struct{ value string }{
		"interpreter": {interpreted.Executions[0].Outcome.Value},
		"native":      {nativeResult.Executions[0].Outcome.Value},
	} {
		if result.value != "42" {
			t.Fatalf("public %s run returned %q, want 42", name, result.value)
		}
	}
}

func TestPhase19Dispatch(t *testing.T) {
	result, err := session.VerifyPhase7ControlsAndWork(context.Background())
	if err != nil || result.Status != "pass" {
		t.Fatalf("CLI-observable exhaustive dispatch lane: status=%q err=%v diagnostics=%+v", result.Status, err, result.Diagnostics)
	}
	found := false
	for _, lane := range result.Lanes {
		if lane.ID == "lane:kind-exhaustive-dispatch-phase07" && lane.Status == "pass" {
			found = true
		}
	}
	if !found {
		t.Fatalf("phase07 lane did not report its OpConst-bearing dispatch pass: %+v", result.Lanes)
	}
}

func TestPhase19LiteralFrontier(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "let count = 42") {
		t.Fatal("literal_tracer.schway lost its direct numeric let")
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("literal_tracer.schway remains refused: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) != 1 || checked.Program.Functions[0].Linear == nil || checked.Program.Functions[0].Linear.Operations[0].Kind != "const" {
		t.Fatalf("literal_tracer.schway did not reach typed constant core: %+v", checked.Program.Functions)
	}
}

func TestPhase19NumericRefusalFrontiers(t *testing.T) {
	for _, tc := range []struct {
		fixture string
		literal string
		code    string
		start   int
	}{
		{fixture: "literal_overflow.schway", literal: "18446744073709551616", code: "check.literal_out_of_range", start: 99},
		{fixture: "literal_malformed.schway", literal: "0x_FF", code: "syntax.malformed_numeric_literal", start: 100},
	} {
		t.Run(tc.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", tc.fixture))
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(source), "let count = "+tc.literal) {
				t.Fatalf("%s lost literal witness %q", tc.fixture, tc.literal)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) == 0 {
				t.Fatalf("%s unexpectedly passed production checking", tc.fixture)
			}
			first := checked.Diagnostics[0]
			if first.Code != tc.code || first.Primary.Start != tc.start || first.Primary.End != tc.start+len(tc.literal) {
				t.Fatalf("%s refusal moved: got id=%q code=%q span=%+v", tc.fixture, first.ID, first.Code, first.Primary)
			}
		})
	}
}

func TestPhase19ScalarGate(t *testing.T) {
	// Reuse the established D-12-18 mutation control so this gate inherits its
	// pinned corpus comparison instead of creating a second comparison law.
	TestPayloadCorpusCharacterizationReplayMutationKilled(t)
	TestPhase19LiteralFrontier(t)
}
