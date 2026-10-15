package interp

import (
	"testing"
)

func TestPhase37CaptureDisabledRetention(t *testing.T) {
	program := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  let n = byte_len(bytes)
  n
}
`)
	probe := &capturePolicyProbe{}
	result, err := runFileApplicationWithProbe(program, "main", "input", func(string) ([]byte, string) {
		return []byte("hello"), ""
	}, false, probe)
	if err != nil {
		t.Fatal(err)
	}
	if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != "5" || string(result.Output) != "5\n" {
		t.Fatalf("execution=%+v output=%q, want returned 5 with 5 newline", result.Execution.Outcome, result.Output)
	}
	if result.Execution.Events != nil || len(result.Execution.Events) != 0 {
		t.Fatalf("events len=%d cap=%d; want nil and zero", len(result.Execution.Events), cap(result.Execution.Events))
	}
	if probe.evidenceCreations != 0 || probe.eventConstructions != 0 || probe.invocationBuilds != 0 || probe.segmentAllocations != 0 || probe.ordinalMapAllocations != 0 || probe.initialEventCapacity != 0 {
		t.Fatalf("capture probe=%+v; all optional-capture retention must stay zero", probe)
	}
	if len(probe.frames) < 2 {
		t.Fatalf("observed %d frames, want root and scalar consumer", len(probe.frames))
	}
	for index, frame := range probe.frames {
		if frame.invocation != "" || frame.entryID != "" || frame.segments || frame.ordinals || frame.evidence {
			t.Errorf("frame %d retained capture identity/state: %+v", index, frame)
		}
	}
}
