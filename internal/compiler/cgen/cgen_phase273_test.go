package cgen_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase273PriorGuard = "  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) return 0;\n"
	phase273NamedGuard = "  if (schway_event_count >= SCHWAY_EVENT_CAPACITY) { fputs(\"schway: event capacity exhausted\\n\", stderr); exit(69); }\n"
)

func phase273Checked(t *testing.T, elements ...string) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath(elements...))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s check diagnostics: %+v", strings.Join(elements, "/"), checked.Diagnostics)
	}
	return checked
}

// TestPhase273NamedCapacityExitGate pins D-27.3-01 and D-27.3-12: the
// conformance recorder of a loop program exits 69 with a named stderr line at
// event capacity, loop-free conformance C keeps its silent guard, and the
// application shell keeps its overflow flag (D-27.3-11).
func TestPhase273NamedCapacityExitGate(t *testing.T) {
	t.Run("loop-conformance", func(t *testing.T) {
		checked := phase273Checked(t, "examples", "sum_to_n.schway")
		c, err := session.Phase16ControlNativeC(checked.Program, "phase273-gate")
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(c, phase273NamedGuard); n != 1 {
			t.Errorf("sum_to_n conformance C has the named-exit recorder line %d times, want 1", n)
		}
		if n := strings.Count(c, "exit(69)"); n != 1 {
			t.Errorf("sum_to_n conformance C has exit(69) %d times, want 1", n)
		}
	})
	t.Run("loop-free-conformance", func(t *testing.T) {
		for _, path := range [][]string{
			{"examples", "phase22", "identity.schway"},
			{"testdata", "phase19", "literal_tracer.schway"},
		} {
			checked := phase273Checked(t, path...)
			c, err := session.Phase16ControlNativeC(checked.Program, "phase273-gate")
			if err != nil {
				t.Fatalf("%s: %v", strings.Join(path, "/"), err)
			}
			if !strings.Contains(c, phase273PriorGuard) {
				t.Errorf("%s: loop-free conformance C lost its prior silent guard", strings.Join(path, "/"))
			}
			if strings.Contains(c, "exit(69)") {
				t.Errorf("%s: loop-free conformance C must not contain exit(69)", strings.Join(path, "/"))
			}
		}
	})
	t.Run("application-shell", func(t *testing.T) {
		checked := phase273Checked(t, "examples", "sum_to_n.schway")
		c, err := cgen.EmitApplication(checked.Program)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(c, "{ schway_event_overflow = 1; return 1; }") {
			t.Error("application shell lost its event overflow flag")
		}
		if strings.Contains(c, "exit(69)") {
			t.Error("application shell must not contain exit(69)")
		}
	})
}
