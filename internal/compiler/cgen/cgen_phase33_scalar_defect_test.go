package cgen_test

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase33ScalarExplicitDefectShells(t *testing.T) {
	for _, location := range []string{"entry", "helper"} {
		t.Run(location, func(t *testing.T) {
			bodyName := "main"
			entry := ""
			if location == "helper" {
				bodyName = "stop"
				entry = "fn main(input: U64) -> U64 { let result = stop(input)\n result\n}\n"
			}
			source := []byte(fmt.Sprintf(`module phase33.scalar_defect
export { fn main }
fn increment(value: U64) -> U64 {
  let result = value + 1
  result
}
fn %s(input: U64) -> U64 {
  var result = increment(input)
  if result < 12 {
    defect "scalar caller stopped"
  } else {
    result = result + 1
  }
  result
}
%s
`, bodyName, entry))
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("checker refused explicit-defect caller: %+v", checked.Diagnostics)
			}
			assertDefect := func(t *testing.T, run interp.Execution) {
				t.Helper()
				if run.Outcome.Kind != "defect" || run.Outcome.Value != "" || len(run.Events) == 0 {
					t.Fatalf("execution=%+v, want terminal defect without a value", run)
				}
				last := run.Events[len(run.Events)-1]
				if last.Kind != "function.defected" || last.Output != "scalar caller stopped" {
					t.Fatalf("terminal event=%+v, want literal defect reason", last)
				}
				if err := executionpeer.Validate(checked.Program, run); err != nil {
					t.Fatalf("defect causal attribution: %v", err)
				}
			}
			interpreted, err := interp.Run(checked.Program, "main", "7")
			if err != nil {
				t.Fatalf("interpreter explicit defect: %v", err)
			}
			assertDefect(t, interpreted)
			t.Run("native", func(t *testing.T) {
				generated, err := cgen.EmitNative(checked.Program)
				if err != nil {
					t.Fatalf("native emission: %v", err)
				}
				for _, lane := range []struct {
					name, optimization string
					lto                bool
				}{{"O0", "-O0", false}, {"O3", "-O3", false}, {"O3-LTO", "-O3", true}} {
					t.Run(lane.name, func(t *testing.T) {
						runner := native.DefaultRunner()
						runner.LTO = lane.lto
						runner.Expect = native.ExpectDefect
						result, err := runner.Run(context.Background(), generated, lane.optimization, []string{"7"})
						if err != nil || len(result.Pairs) != 1 {
							t.Fatalf("native defect document: pairs=%d err=%v", len(result.Pairs), err)
						}
						assertDefect(t, result.Pairs[0].Execution)
					})
				}
			})
			t.Run("application", func(t *testing.T) {
				generated, err := cgen.EmitApplication(checked.Program)
				if err != nil {
					t.Fatalf("application emission: %v", err)
				}
				artifact := filepath.Join(t.TempDir(), "explicit-defect")
				if _, err := native.BuildApplication(context.Background(), source, generated, artifact); err != nil {
					t.Fatalf("application build: %v", err)
				}
				var stdout, stderr bytes.Buffer
				outcome, err := native.RunApplication(context.Background(), artifact, "7", &stdout, &stderr)
				if err != nil || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway app: scalar caller stopped\n" {
					t.Fatalf("application outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
				}
			})
		})
	}
}
