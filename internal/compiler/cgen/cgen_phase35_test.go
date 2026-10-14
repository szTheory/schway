package cgen_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase35TracerSource = `module phase35.native_shift

export {
  fn main
}

fn main(input: U64) -> U64 {
  let shifted = 1 << 63
  shifted
}
`

func phase35Check(t *testing.T, source string) core.Program {
	t.Helper()
	checked := session.Check([]byte(source))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", checked.Diagnostics)
	}
	return checked.Program
}

func phase35CloneProgram(t *testing.T, program core.Program) core.Program {
	t.Helper()
	encoded, err := json.Marshal(program)
	if err != nil {
		t.Fatal(err)
	}
	var clone core.Program
	if err := json.Unmarshal(encoded, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}

func TestPhase35NativeShiftTracer(t *testing.T) {
	program := phase35Check(t, phase35TracerSource)
	generated, err := cgen.EmitNative(program)
	if err != nil {
		t.Fatalf("EmitNative: %v", err)
	}
	function := program.Functions[0]
	var shift core.LinearOperation
	for _, operation := range function.Linear.Operations {
		if operation.Kind == core.OpShiftLeftU64 {
			shift = operation
			break
		}
	}
	if shift.ID == "" || shift.ConstU64 != "63" || shift.RightID != "" {
		t.Fatalf("checked shift = %+v, want left shift by canonical literal 63", shift)
	}
	placeIndex := func(id string) int {
		for index, place := range function.Linear.Places {
			if place.ID == id {
				return index
			}
		}
		return -1
	}
	sourceLocal := fmt.Sprintf("schway_scalar_%d", placeIndex(shift.SourceID))
	targetLocal := fmt.Sprintf("schway_scalar_%d", placeIndex(shift.TargetID))
	wantSite := fmt.Sprintf("%s = (uint64_t)((unsigned long long)%s << 63);", targetLocal, sourceLocal)
	if !strings.Contains(generated, wantSite) {
		t.Fatalf("generated C lacks unsigned exact-width shift site %q", wantSite)
	}
	for _, unsafe := range []string{"(long long)" + sourceLocal + " << 63", sourceLocal + " << UINT64_C(63)"} {
		if strings.Contains(generated, unsafe) {
			t.Fatalf("generated C contains unsafe shift fragment %q", unsafe)
		}
	}

	got, err := native.DefaultRunner().Run(context.Background(), generated, "-O0", []string{"0"})
	if err != nil || len(got.Pairs) != 1 {
		t.Fatalf("native run: pairs=%d err=%v", len(got.Pairs), err)
	}
	if value := got.Pairs[0].Execution.Outcome.Value; value != "9223372036854775808" {
		t.Fatalf("native result = %q, want 9223372036854775808", value)
	}
	events := got.Pairs[0].Execution.Events
	if len(events) != 2 || events[0].Kind != "value.copied" || events[1].Kind != "function.returned" {
		t.Fatalf("shift emitted unexpected events: %+v", events)
	}
}

func TestPhase35EmitterAdmissionGates(t *testing.T) {
	program := phase35Check(t, phase35TracerSource)
	for index := range program.Functions {
		if program.Functions[index].Linear == nil {
			continue
		}
		for operationIndex := range program.Functions[index].Linear.Operations {
			operation := &program.Functions[index].Linear.Operations[operationIndex]
			if operation.Kind == core.OpShiftLeftU64 {
				operation.ConstU64 = "64"
				break
			}
		}
	}
	for _, route := range []struct {
		name string
		emit func(core.Program) (string, error)
	}{
		{name: "Emit", emit: cgen.Emit},
		{name: "EmitNative", emit: cgen.EmitNative},
		{name: "EmitApplication", emit: cgen.EmitApplication},
	} {
		t.Run(route.name, func(t *testing.T) {
			generated, err := route.emit(phase35CloneProgram(t, program))
			if err == nil {
				t.Fatal("accepted forged shift count 64")
			}
			if generated != "" {
				t.Fatalf("malformed program emitted C (%d bytes)", len(generated))
			}
			if !strings.Contains(err.Error(), "core validation failed:") {
				t.Fatalf("refusal did not come from the independent raw-core gate: %v", err)
			}
		})
	}
}
