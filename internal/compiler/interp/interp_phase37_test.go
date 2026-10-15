package interp

import (
	"errors"
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

func TestPhase37CaptureDisabledOutcomes(t *testing.T) {
	plainProgram := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  let n = byte_len(bytes)
  n
}
`)
	for _, row := range []struct {
		name string
		data []byte
		want string
	}{
		{name: "zero", data: nil, want: "0"},
		{name: "nonzero", data: []byte("hello"), want: "5"},
	} {
		t.Run("successful-"+row.name, func(t *testing.T) {
			probe := &capturePolicyProbe{}
			result, err := runFileApplicationWithProbe(plainProgram, "main", "input", func(string) ([]byte, string) {
				return row.data, ""
			}, false, probe)
			if err != nil {
				t.Fatal(err)
			}
			if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != row.want || string(result.Output) != row.want+"\n" {
				t.Fatalf("outcome=%+v output=%q, want returned %s", result.Execution.Outcome, result.Output, row.want)
			}
			if result.Execution.Events != nil || len(result.Execution.LiveResources) != 0 || probe.resourceReleases != 1 {
				t.Fatalf("events=%v live=%v semantic releases=%d; want no events, no live owner, one release", result.Execution.Events, result.Execution.LiveResources, probe.resourceReleases)
			}
		})
	}
	for _, alternative := range []string{"FileTooLarge", "FileUnavailable", "AllocFailed"} {
		t.Run("typed-"+alternative, func(t *testing.T) {
			probe := &capturePolicyProbe{}
			result, err := runFileApplicationWithProbe(plainProgram, "main", "input", func(string) ([]byte, string) {
				return nil, alternative
			}, false, probe)
			if err != nil {
				t.Fatal(err)
			}
			if result.Execution.Outcome.Kind != "typed_failure" || result.Execution.Outcome.Value != alternative || len(result.Output) != 0 {
				t.Fatalf("outcome=%+v output=%q, want typed failure %s with empty output", result.Execution.Outcome, result.Output, alternative)
			}
			if result.Execution.Events != nil || len(result.Execution.LiveResources) != 0 || probe.resourceReleases != 0 {
				t.Fatalf("events=%v live=%v semantic releases=%d; typed acquire failure must own/release nothing", result.Execution.Events, result.Execution.LiveResources, probe.resourceReleases)
			}
		})
	}
	defectProgram := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  var n = byte_len(bytes)
  if n < 10 {
    write_u64(n)
    write_newline()
    defect "boom"
  } else {
    n = n + 0
  }
  let result = identity(n)
  result
}
fn identity(value: U64) -> U64 {
  value
}
`)
	t.Run("prior-output-then-checked-defect", func(t *testing.T) {
		probe := &capturePolicyProbe{}
		result, err := runFileApplicationWithProbe(defectProgram, "main", "input", func(string) ([]byte, string) {
			return []byte("hello"), ""
		}, false, probe)
		if err != nil {
			t.Fatal(err)
		}
		if result.Execution.Outcome.Kind != "defect" || string(result.Output) != "5\n" || len(result.Execution.LiveResources) != 1 {
			t.Fatalf("outcome=%+v output=%q live=%v, want late defect with prior output and existing live-owner policy", result.Execution.Outcome, result.Output, result.Execution.LiveResources)
		}
		if result.Execution.Events != nil || probe.resourceReleases != 0 {
			t.Fatalf("events=%v semantic releases=%d; abort-like defect must retain no trace and leave cleanup policy unchanged", result.Execution.Events, probe.resourceReleases)
		}
	})
	t.Run("capture-exhaustion-keeps-later-terminal-outcome", func(t *testing.T) {
		restore := SetEvidenceLimitsForTest(1, 1, 1, 0)
		defer restore()
		result, err := runFileApplicationWithReader(defectProgram, "main", "input", func(string) ([]byte, string) {
			return []byte("hello"), ""
		}, true)
		if !errors.Is(err, ErrEvidenceCapacityExhausted) {
			t.Fatalf("err=%v, want capture-capacity status alongside the real terminal execution", err)
		}
		if result.Execution.Outcome.Kind != "defect" || string(result.Output) != "5\n" || len(result.Execution.LiveResources) != 1 {
			t.Fatalf("outcome=%+v output=%q live=%v, want late defect preserved after evidence exhaustion", result.Execution.Outcome, result.Output, result.Execution.LiveResources)
		}
	})
}
