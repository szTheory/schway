package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 27.2 witnesses for R1: branch-free scalar bodies (no if/while) that
// the checker accepts must also pass corevalidate/originvalidate, build, and
// run identically on both engines (D-27.2-01, D-27.2-06, D-27.2-13). No public
// cgen emitter is called here (Phase 16 inventory, D-27.2-14 amendment).

type phase272Witness struct {
	name, ret, body, input, wantStdout, wantReturn string
	wantDefect                                     bool
}

func phase272Source(ret, body string) []byte {
	return []byte("module t.sl\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> " + ret + " {\n" + body + "}\n")
}

// phase272RequireWitness runs one witness across check, corevalidate, both
// engines, build and executionpeer. The corevalidate step runs before any
// engine so the red output names the first refusal code.
func phase272RequireWitness(t *testing.T, w phase272Witness) {
	t.Helper()
	phase272RequireSource(t, phase272Source(w.ret, w.body), w)
}

func phase272RequireSource(t *testing.T, source []byte, w phase272Witness) {
	t.Helper()
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("phase27.2 witness not checked: %+v", checked.Diagnostics)
	}
	if validated := corevalidate.Validate(checked.Program); !validated.Valid {
		t.Fatalf("phase27.2 witness %s refused by corevalidate: first problem %s (%d problems)", w.name, validated.Problems[0].Code, len(validated.Problems))
	}
	interpreted, diagnostics, err := session.RunInterpreterApplication(source, w.input)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("interpreter application diagnostics=%+v err=%v", diagnostics, err)
	}
	artifact := filepath.Join(t.TempDir(), "witness")
	if _, diagnostics, err := session.BuildApplication(context.Background(), source, artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build diagnostics=%+v err=%v", diagnostics, err)
	}
	if w.wantDefect {
		phase272RequireOverflowDefect(t, checked, interpreted.Execution, interpreted.Output, artifact, w.input)
		return
	}
	var stdout, stderr bytes.Buffer
	reportPath := filepath.Join(t.TempDir(), "report.json")
	outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, w.input, reportPath, native.EvidenceEvents, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stderr.Len() != 0 {
		t.Fatalf("native outcome=%+v stderr=%q err=%v, want exit 0 and empty stderr", outcome, stderr.String(), err)
	}
	if stdout.String() != w.wantStdout {
		t.Fatalf("native stdout=%q, want %q", stdout.String(), w.wantStdout)
	}
	if interpreted.Execution.Outcome.Kind != execution.OutcomeReturned || interpreted.Execution.Outcome.Value != w.wantReturn || string(interpreted.Output) != w.wantStdout {
		t.Fatalf("interpreter outcome=%+v stdout=%q, want return %q and stdout %q", interpreted.Execution.Outcome, interpreted.Output, w.wantReturn, w.wantStdout)
	}
	if !bytes.Equal(interpreted.Output, stdout.Bytes()) {
		t.Fatalf("engines disagree: interpreter=%q native=%q", interpreted.Output, stdout.Bytes())
	}
	projected, err := session.ProjectExecutionSchema2(checked.Program, interpreted.Execution)
	if err != nil {
		t.Fatalf("project interpreter execution: %v", err)
	}
	if err := executionpeer.Validate(checked.Program, projected); err != nil {
		t.Fatalf("executionpeer rejected interpreter execution: %v", err)
	}
	if report.Execution == nil {
		t.Fatalf("native report carried no execution (capture status %q)", report.CaptureStatus)
	}
	if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
		t.Fatalf("executionpeer rejected native execution: %v", err)
	}
}

func phase272RequireOverflowDefect(t *testing.T, checked session.CheckResult, interpreted execution.Execution, interpretedOutput []byte, artifact, input string) {
	t.Helper()
	if interpreted.Outcome.Kind != execution.OutcomeDefect || len(interpretedOutput) != 0 || len(interpreted.Events) == 0 {
		t.Fatalf("interpreter outcome=%+v stdout=%q, want empty defect with events", interpreted.Outcome, interpretedOutput)
	}
	event := interpreted.Events[len(interpreted.Events)-1]
	if event.Kind != "function.defected" || event.Output != "U64 addition overflow" {
		t.Fatalf("interpreter terminal event=%+v, want function.defected U64 addition overflow", event)
	}
	projected, err := session.ProjectExecutionSchema2(checked.Program, interpreted)
	if err != nil {
		t.Fatalf("project interpreter defect: %v", err)
	}
	if err := executionpeer.Validate(checked.Program, projected); err != nil {
		t.Fatalf("executionpeer rejected interpreter defect: %v", err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, input, &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway: U64 addition overflow\n" {
		t.Fatalf("native overflow outcome=%+v stdout=%q stderr=%q err=%v, want exit 65 and fixed stderr only", outcome, stdout.String(), stderr.String(), err)
	}
}

func TestPhase272HeadlineStraightLineCopy(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase272", "straight_line_copy.schway"))
	if err != nil {
		t.Fatal(err)
	}
	phase272RequireSource(t, source, phase272Witness{name: "headline", input: "41", wantStdout: "41\n", wantReturn: "41"})
}

func TestPhase272StraightLineScalarAdmission(t *testing.T) {
	witnesses := []phase272Witness{
		{name: "pure-copy", ret: "U64", body: "  let total = n\n  total\n", input: "41", wantStdout: "41\n", wantReturn: "41"},
		{name: "pure-const", ret: "U64", body: "  let k = 7\n  k\n", input: "41", wantStdout: "7\n", wantReturn: "7"},
		{name: "var-only", ret: "U64", body: "  var total = n\n  total\n", input: "41", wantStdout: "41\n", wantReturn: "41"},
		{name: "var-assign", ret: "U64", body: "  var t = n\n  t = t + 1\n  t\n", input: "41", wantStdout: "42\n", wantReturn: "42"},
		{name: "let-add", ret: "U64", body: "  let t = n + 1\n  t\n", input: "41", wantStdout: "42\n", wantReturn: "42"},
		{name: "mixed", ret: "U64", body: "  let a = n\n  var b = a + 1\n  b = b + a\n  b\n", input: "41", wantStdout: "83\n", wantReturn: "83"},
		{name: "chain", ret: "U64", body: "  let a = n\n  let b = a\n  let c = 5\n  let d = b\n  d\n", input: "41", wantStdout: "41\n", wantReturn: "41"},
		{name: "unit-write-u64", ret: "Unit", body: "  let x = n\n  write_u64(x)\n  write_newline()\n", input: "41", wantStdout: "41\n", wantReturn: ""},
		{name: "unit-write-text", ret: "Unit", body: "  write_text(\"hi\\n\")\n", input: "41", wantStdout: "hi\n", wantReturn: ""},
		{name: "unit-newline", ret: "Unit", body: "  write_newline()\n", input: "41", wantStdout: "\n", wantReturn: ""},
		{name: "unit-copy-only", ret: "Unit", body: "  let x = n\n", input: "41", wantStdout: "", wantReturn: ""},
		// Constant-only overflow reads no n, so it doubles as the unused-parameter
		// regression. It checks and builds, then defects at runtime (Pitfall 3).
		{name: "overflow-defect", ret: "U64", body: "  let m = 18446744073709551615\n  let t = m + 1\n  t\n", input: "41", wantDefect: true},
	}
	for _, w := range witnesses {
		t.Run(w.name, func(t *testing.T) { phase272RequireWitness(t, w) })
	}
}

func TestPhase272NegativeControls(t *testing.T) {
	checkerRefusals := []struct{ name, ret, body, code string }{
		{"uninitialized-read", "U64", "  let t = n\n  if n < 1 {\n    let q = 1\n  } else {\n    let r = 2\n  }\n  q\n", "name.uninitialized_place"},
		{"unknown-place", "U64", "  var t = n\n  q\n", "name.unknown_place"},
		{"type-mismatch", "U64", "  let b = n < 5\n  b\n", "type.scalar_result"},
		{"unit-returns-value", "Unit", "  let x = n\n  x\n", "type.unit_result"},
		{"constant-zero-divisor", "U64", "  let z = 0\n  let t = n % z\n  t\n", "type.scalar_remainder_zero"},
	}
	for _, c := range checkerRefusals {
		t.Run(c.name, func(t *testing.T) {
			checked := session.Check(phase272Source(c.ret, c.body))
			if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != c.code {
				t.Fatalf("diagnostics=%+v, want first code %s", checked.Diagnostics, c.code)
			}
		})
	}
	t.Run("unreachable-trailing-result", func(t *testing.T) {
		body := "  let t = n\n  if n < 1 {\n    defect \"a\"\n  } else {\n    defect \"b\"\n  }\n  t\n"
		checked := session.Check(phase272Source("U64", body))
		if len(checked.Diagnostics) != 0 {
			t.Fatalf("phase27.2 control not reached: checker diagnostics %+v", checked.Diagnostics)
		}
		validated := corevalidate.Validate(checked.Program)
		if validated.Valid || validated.Problems[0].Code != "core.terminal_block_unreachable" {
			t.Fatalf("corevalidate valid=%v problems=%+v, want first problem core.terminal_block_unreachable", validated.Valid, validated.Problems)
		}
	})
}

// TestPhase272R1bScalarFirstLetThenCallCharacterization pins KNOWN DEBT R1b
// (D-27.2-12): a671308 routes the body to the scalar parser on the first let
// only, so a scalar-first let followed by a call let is refused. This records
// current behavior, not desired behavior.
func TestPhase272R1bScalarFirstLetThenCallCharacterization(t *testing.T) {
	source := []byte("module t.r1b\n\nexport {\n  fn main\n}\n\nfn helper(x: U64) -> U64 {\n  x\n}\n\nfn main(n: U64) -> U64 {\n  let a = n\n  let b = helper(a)\n  b\n}\n")
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 || checked.Diagnostics[0].Code != "syntax.expected_linear_result" {
		t.Fatalf("diagnostics=%+v, want first code syntax.expected_linear_result (debt R1b characterization)", checked.Diagnostics)
	}
}
