package session_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase17TwoTypeGeneratedC(t *testing.T) {
	_, _, generated := phase17TwoTypeFourTiers(t)
	for _, want := range []string{
		"static LANG_RESULT LANG_CLASSIFY(LANG_RESOURCE, unsigned int);",
		"static LANG_RESULT LANG_MAIN(LANG_RESOURCE, unsigned int);",
		"LANG_RESULT lang_entry_output = LANG_MAIN(lang_entry_input, 0u);",
		"LANG_RESULT_name(lang_entry_output)",
	} {
		if !strings.Contains(generated, want) {
			t.Fatalf("public EmitNative C misses %q:\n%s", want, generated)
		}
	}
}

func TestPhase17TwoTypeFourTierDifferential(t *testing.T) {
	program, engines, _ := phase17TwoTypeFourTiers(t)
	if len(engines) != 4 {
		t.Fatalf("four-tier run returned %d engine documents", len(engines))
	}
	if err := session.Phase5CompareProgramEngines("phase17-return-type-tracer", program, engines); err != nil {
		t.Fatalf("Resource -> Result tiers diverged: %v", err)
	}
}

func TestPhase17TwoTypeEmitterGuardIsNotInert(t *testing.T) {
	program, engines, generated := phase17TwoTypeFourTiers(t)
	mutatedC := strings.Replace(generated, "case LANG_RESULT_LANG_CLASSIFIED: return \"Classified\";", "case LANG_RESULT_LANG_CLASSIFIED: return \"phase17-seeded-return-mutation\";", 1)
	if mutatedC == generated {
		t.Fatal("return-side C mutation did not reach the Result renderer")
	}
	mutated, err := native.DefaultRunner().Run(context.Background(), mutatedC, "-O3", []string{"Raw"})
	if err != nil || len(mutated.Pairs) != 1 {
		t.Fatalf("mutated return-side C did not execute at -O3: pairs=%d err=%v", len(mutated.Pairs), err)
	}
	engines["O3"] = mutated.Pairs[0].Execution
	err = session.Phase5CompareProgramEngines("phase17-return-mutation", program, engines)
	if err == nil {
		t.Fatal("return-side emitter mutation stayed green")
	}
	var disagreement *session.Phase5EngineDisagreement
	if !errors.As(err, &disagreement) || disagreement.Axis != session.AxisTerminalOutcome {
		t.Fatalf("return-side mutation must move terminal-outcome axis, got %v", err)
	}
}

func phase17TwoTypeFourTiers(t *testing.T) (core.Program, map[string]execution.Execution, string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check tracer: %+v", checked.Diagnostics)
	}
	entryName := "main"
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("public EmitNative: %v", err)
	}
	engines := phase11RunFourTiersWithSupplier(t, context.Background(), checked.Program, entryName, "Raw", cgen.EmitNative, "public EmitNative")
	engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
	return checked.Program, engines, generated
}
