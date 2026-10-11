package session_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase29OOBEvidenceSource = `module phase29.oob.evidence

export {
  fn probe
}

fn probe(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}
`

// TestPhase29OOBEvidenceProjection carries one out-of-range byte_at from the
// interpreter through the schema-2 projection to an accepted peer verdict.
func TestPhase29OOBEvidenceProjection(t *testing.T) {
	checked := session.Check([]byte(phase29OOBEvidenceSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	run, err := interp.RunByteViewForTest(checked.Program, "probe", []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	if run.Outcome.Kind != execution.OutcomeDefect || run.Outcome.Value != "" {
		t.Fatalf("outcome = %+v, want empty defect", run.Outcome)
	}
	final := run.Events[len(run.Events)-1]
	if final.Kind != "function.defected" || final.Output != "byte index out of bounds" || final.SourcePlace == "" {
		t.Fatalf("final event = %+v, want source-attributed byte index defect", final)
	}
	projected, err := session.ProjectExecutionSchema2(checked.Program, run)
	if err != nil {
		t.Fatalf("project: %v", err)
	}
	got := projected.Events[len(projected.Events)-1]
	if got.Kind != "function.defected" || got.Output != "byte index out of bounds" {
		t.Fatalf("projected final event = %+v, want reason byte index out of bounds", got)
	}
	if err := executionpeer.Validate(checked.Program, projected); err != nil {
		t.Fatalf("peer rejected interpreter OOB evidence: %v", err)
	}
}
