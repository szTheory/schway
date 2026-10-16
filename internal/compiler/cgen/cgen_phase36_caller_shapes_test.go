package cgen_test

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase36UnusedByteStructuredHelper(t *testing.T) {
	// This is the only helper call in the whole program. An unrelated call in
	// another U64/Bool body must not be needed to select dynamic emission.
	checked := session.Check([]byte(`module phase36.unused_byte_call
export { fn main }
fn increment(value: U64) -> U64 {
  let result = value + 1
  result
}

fn main(unused: Byte) -> U64 {
  var result = increment(10)
  result
}
`))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("checker refused unused-Byte helper: %+v", checked.Diagnostics)
	}
	interpreted, err := interp.Run(checked.Program, "main", "7")
	if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != "11" {
		t.Fatalf("interpreter literal answer=%+v err=%v, want returned 11", interpreted.Outcome, err)
	}
	if err := executionpeer.Validate(checked.Program, interpreted); err != nil {
		t.Fatalf("interpreter causal attribution: %v", err)
	}
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatalf("unused-Byte helper native emission: %v", err)
	}
	for _, lane := range []struct {
		name, optimization string
		lto                bool
	}{{"O0", "-O0", false}, {"O3", "-O3", false}, {"O3-LTO", "-O3", true}} {
		t.Run(lane.name, func(t *testing.T) {
			runner := native.DefaultRunner()
			runner.LTO = lane.lto
			result, err := runner.Run(context.Background(), generated, lane.optimization, []string{"7"})
			if err != nil || len(result.Pairs) != 1 {
				t.Fatalf("native execution pairs=%d err=%v", len(result.Pairs), err)
			}
			actual := result.Pairs[0].Execution
			if actual.Outcome.Kind != "returned" || actual.Outcome.Value != "11" {
				t.Fatalf("native literal answer=%+v, want returned 11", actual.Outcome)
			}
			if err := executionpeer.Validate(checked.Program, actual); err != nil {
				t.Fatalf("native causal attribution: %v", err)
			}
		})
	}
}

func TestPhase36ByteViewUnitHelperEmission(t *testing.T) {
	// Keep the declared Unit body in the translation unit even though main
	// does not call it. This tests emission and linking, not a new ByteView
	// value-call signature or a public ByteView entry route.
	source := []byte(`module phase36.byteview_unit_call
export { fn main }
fn increment(value: U64) -> U64 {
  let result = value + 1
  result
}
fn consume(bytes: ByteView) -> Unit {
  var result = increment(10)
  write_u64(result)
}
fn main(input: U64) -> U64 {
  let result = input
  result
}
`)
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("checker refused ByteView Unit body: %+v", checked.Diagnostics)
	}
	generated, err := cgen.EmitApplication(checked.Program)
	if err != nil {
		t.Fatalf("ByteView Unit helper native emission: %v", err)
	}
	artifact := filepath.Join(t.TempDir(), "byteview-unit")
	if _, err := native.BuildApplication(context.Background(), source, generated, artifact); err != nil {
		t.Fatalf("ByteView Unit helper translation unit: %v", err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "7", &stdout, &stderr)
	if err != nil || outcome.ExitCode != 0 || stdout.String() != "7\n" || stderr.Len() != 0 {
		t.Fatalf("unrelated entry outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
	}
}
