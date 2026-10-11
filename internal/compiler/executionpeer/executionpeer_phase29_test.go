package executionpeer_test

import (
	"testing"

	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/session"
)

const phase29PeerOOBSource = `module phase29.peer.oob

export {
  fn probe
}

fn probe(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  var x = byte_at(bytes, n)
  x
}
`

func TestPhase29ExecutionPeerByteIndexDefect(t *testing.T) {
	checked := session.Check([]byte(phase29PeerOOBSource))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check diagnostics: %+v", checked.Diagnostics)
	}
	run, err := interp.RunByteViewForTest(checked.Program, "probe", []byte("a"))
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := session.ProjectExecutionSchema2(checked.Program, run)
	if err != nil {
		t.Fatal(err)
	}
	if err := executionpeer.Validate(checked.Program, accepted); err != nil {
		t.Fatalf("peer rejected genuine byte index defect: %v", err)
	}
	last := len(accepted.Events) - 1
	clone := func() execution.Execution {
		document := accepted
		document.Events = append([]execution.Event{}, accepted.Events...)
		return document
	}
	faults := []struct {
		name   string
		mutate func(*execution.Execution)
	}{
		{"wrong reason", func(d *execution.Execution) { d.Events[last].Output = "index out of range" }},
		{"non-final defect", func(d *execution.Execution) {
			extra := d.Events[last]
			extra.Occurrence++
			d.Events = append(d.Events, extra)
		}},
		{"other place", func(d *execution.Execution) { d.Events[last].SourcePlace = "some:other:place" }},
		{"non-empty value", func(d *execution.Execution) { d.Outcome.Value = "0" }},
	}
	for _, fault := range faults {
		t.Run(fault.name, func(t *testing.T) {
			document := clone()
			fault.mutate(&document)
			if err := executionpeer.Validate(checked.Program, document); err == nil {
				t.Fatalf("peer accepted seeded fault %q", fault.name)
			}
		})
	}
}
