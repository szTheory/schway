package interp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPhase37CaptureRetentionControls(t *testing.T) {
	program := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  var index = 0
  var value = byte_len(bytes)
  while index < 3 {
    value = identity(value + index)
    index = index + 1
  }
  value
}
fn identity(value: U64) -> U64 {
  value
}
`)
	probe := &capturePolicyProbe{}
	result, err := runFileApplicationWithProbe(program, "main", "input", func(string) ([]byte, string) {
		return []byte("hello"), ""
	}, false, probe)
	if err != nil {
		t.Fatal(err)
	}
	if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != "8" || string(result.Output) != "8\n" {
		t.Fatalf("execution=%+v output=%q, want returned 8 after three scalar calls", result.Execution.Outcome, result.Output)
	}
	if result.Execution.Events != nil || len(result.Execution.Events) != 0 {
		t.Fatalf("events len=%d cap=%d; want nil and zero", len(result.Execution.Events), cap(result.Execution.Events))
	}
	if probe.evidenceCreations != 0 || probe.eventConstructions != 0 || probe.invocationBuilds != 0 || probe.segmentAllocations != 0 || probe.ordinalMapAllocations != 0 || probe.initialEventCapacity != 0 {
		t.Fatalf("capture probe=%+v; all optional-capture retention must stay zero", probe)
	}
	if len(probe.frames) < 5 {
		t.Fatalf("observed %d frames, want application root and three repeated scalar calls", len(probe.frames))
	}
	for index, frame := range probe.frames {
		if frame.invocation != "" || frame.entryID != "" || frame.segments || frame.ordinals || frame.evidence {
			t.Errorf("frame %d retained capture identity/state: %+v", index, frame)
		}
	}

	mutations := []struct {
		name        string
		anchor      string
		replacement string
	}{
		{
			name: "recorder",
			anchor: `if !base.captureEnabled {
		schema = schemaForProgram(program, execution.Schema1)
	} else if hasScalarValueCalls(program) {`,
			replacement: `if !base.captureEnabled && base.captureProbe == nil {
		schema = schemaForProgram(program, execution.Schema1)
	} else if hasScalarValueCalls(program) {`,
		},
		{
			name: "event",
			anchor: `func appendFrameEvent(f *frame, events []Event, build func() Event) []Event {
	if f == nil || !f.captureEnabled {`,
			replacement: `func appendFrameEvent(f *frame, events []Event, build func() Event) []Event {
	if f == nil || (!f.captureEnabled && f.captureProbe == nil) {`,
		},
		{
			name: "invocation",
			anchor: `if !base.captureEnabled {
		schema = schemaForProgram(program, execution.Schema1)
	} else if hasScalarValueCalls(program) {`,
			replacement: `if !base.captureEnabled {
		schema, invocation, err = entryIdentity(program, base.function.ID, execution.Schema1)
	} else if hasScalarValueCalls(program) {`,
		},
		{
			name: "segment",
			anchor: `if !caller.captureEnabled {
		// The file-application policy carries no invocation path or call ordinal.
	} else if caller.evidence == nil {`,
			replacement: `if !caller.captureEnabled {
		if caller.captureProbe != nil {
			caller.captureProbe.segmentAllocations++
		}
		segments = []execution.InvocationSegment{}
		segments = append(append([]execution.InvocationSegment{}, caller.segments...), execution.InvocationSegment{OpCallID: operation.ID, Ordinal: 0})
	} else if caller.evidence == nil {`,
		},
		{
			name: "ordinal",
			anchor: `if !caller.captureEnabled {
		// The file-application policy carries no invocation path or call ordinal.
	} else if caller.evidence == nil {`,
			replacement: `if !caller.captureEnabled {
		if caller.nextCallOrdinals == nil {
			caller.nextCallOrdinals = make(map[string]uint64)
			if caller.captureProbe != nil {
				caller.captureProbe.ordinalMapAllocations++
			}
		}
		caller.nextCallOrdinals[operation.ID]++
		if caller.captureProbe != nil {
			caller.captureProbe.observeFrame(caller)
		}
	} else if caller.evidence == nil {`,
		},
	}
	original, err := os.ReadFile("interp.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range mutations {
		t.Run(mutation.name, func(t *testing.T) {
			if strings.Count(string(original), mutation.anchor) != 1 {
				t.Fatalf("source mutation anchor occurs %d times; want exactly one", strings.Count(string(original), mutation.anchor))
			}
			mutant := strings.Replace(string(original), mutation.anchor, mutation.replacement, 1)
			phase37RunMutationOverlay(t, "interp.go", mutant, "TestPhase37CaptureRetentionMutationTarget", mutation.name, "RETENTION_ASSERTION["+mutation.name+"]")
		})
	}
}

// TestPhase37CaptureRetentionMutationTarget is exercised directly as a clean
// baseline and by the bounded overlay children above. Armed children fail
// after the real zero-retention discriminator observes named production-path
// growth.
func TestPhase37CaptureRetentionMutationTarget(t *testing.T) {
	mode := os.Getenv("SCHWAY_PHASE37_MUTANT")
	program := phase321CheckFile(t, `fn adler32(bytes: ByteView) -> U64 {
  var index = 0
  var value = byte_len(bytes)
  while index < 3 {
    value = identity(value + index)
    index = index + 1
  }
  value
}
fn identity(value: U64) -> U64 {
  value
}
`)
	probe := &capturePolicyProbe{}
	result, err := runFileApplicationWithProbe(program, "main", "input", func(string) ([]byte, string) {
		return []byte("hello"), ""
	}, false, probe)
	if err != nil {
		t.Fatal(err)
	}
	if result.Execution.Outcome.Kind != "returned" || result.Execution.Outcome.Value != "8" || string(result.Output) != "8\n" {
		t.Fatalf("mutant changed repeated-call result: outcome=%+v output=%q", result.Execution.Outcome, result.Output)
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
	defectProbe := &capturePolicyProbe{}
	defectResult, err := runFileApplicationWithProbe(defectProgram, "main", "input", func(string) ([]byte, string) {
		return []byte("hello"), ""
	}, false, defectProbe)
	if err != nil {
		t.Fatal(err)
	}
	if defectResult.Execution.Outcome.Kind != "defect" || string(defectResult.Output) != "5\n" || len(defectResult.Execution.LiveResources) != 1 || defectProbe.resourceReleases != 0 {
		t.Fatalf("mutant changed late-defect contract: outcome=%+v output=%q live=%v semantic releases=%d", defectResult.Execution.Outcome, defectResult.Output, defectResult.Execution.LiveResources, defectProbe.resourceReleases)
	}
	violations := phase37CaptureRetentionViolations(result, probe)
	if mode == "" {
		if len(violations) != 0 {
			t.Fatalf("clean retention target observed growth: %v", violations)
		}
		if defectViolations := phase37CaptureRetentionViolations(defectResult, defectProbe); len(defectViolations) != 0 {
			t.Fatalf("clean late-defect retention target observed growth: %v", defectViolations)
		}
		return
	}
	phase37AssertMutationChildRoot(t)
	if detail, reached := violations[mode]; reached {
		t.Fatalf("RETENTION_ASSERTION[%s]: %s", mode, detail)
	}
	t.Fatalf("RETENTION_MUTANT_SURVIVED[%s]: zero-retention assertions did not observe growth; probe=%+v frames=%+v", mode, probe, probe.frames)
}

func phase37CaptureRetentionViolations(result ApplicationResult, probe *capturePolicyProbe) map[string]string {
	violations := make(map[string]string)
	if probe.evidenceCreations != 0 {
		violations["recorder"] = fmt.Sprintf("evidence creations=%d", probe.evidenceCreations)
	}
	if probe.eventConstructions != 0 || len(result.Execution.Events) != 0 {
		violations["event"] = fmt.Sprintf("event constructions=%d retained events=%d", probe.eventConstructions, len(result.Execution.Events))
	}
	for _, frame := range probe.frames {
		if frame.invocation != "" || frame.entryID != "" {
			violations["invocation"] = fmt.Sprintf("retained invocation=%q entryID=%q", frame.invocation, frame.entryID)
			break
		}
	}
	if probe.segmentAllocations != 0 {
		violations["segment"] = fmt.Sprintf("segment allocations=%d", probe.segmentAllocations)
	} else {
		for _, frame := range probe.frames {
			if frame.segments {
				violations["segment"] = "retained invocation segments"
				break
			}
		}
	}
	if probe.ordinalMapAllocations != 0 {
		violations["ordinal"] = fmt.Sprintf("ordinal map allocations=%d", probe.ordinalMapAllocations)
	} else {
		for _, frame := range probe.frames {
			if frame.ordinals {
				violations["ordinal"] = "retained call ordinal map"
				break
			}
		}
	}
	return violations
}

func phase37RunMutationOverlay(t *testing.T, sourceName, mutantSource, childTest, mode, expectedAssertion string) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "..", "..", ".."))
	sourcePath, err := filepath.Abs(sourceName)
	if err != nil {
		t.Fatal(err)
	}
	mutantPath := filepath.Join(t.TempDir(), filepath.Base(sourceName)+".mutant.go")
	if err := os.WriteFile(mutantPath, []byte(mutantSource), 0o600); err != nil {
		t.Fatal(err)
	}
	overlayBytes, err := json.Marshal(struct {
		Replace map[string]string `json:"Replace"`
	}{Replace: map[string]string{sourcePath: mutantPath}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, overlayBytes, 0o600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "test", "./internal/compiler/interp", "-run", "^"+childTest+"$", "-count=1", "-v", "-timeout=75s", "-overlay="+overlayPath)
	command.Dir = root
	command.Env = phase37MutationEnvironment(root, mode)
	output, runErr := command.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatalf("overlay child timed out for %s: %v\n%s", mode, ctx.Err(), output)
	}
	text := string(output)
	if runErr == nil {
		t.Fatalf("overlay mutant %s survived; child passed\n%s", mode, text)
	}
	if !strings.Contains(text, "=== RUN   "+childTest) || !strings.Contains(text, "PHASE37_CHILD_ROOT="+root) || !strings.Contains(text, expectedAssertion) {
		t.Fatalf("overlay mutant %s did not reach its semantic assertion (err=%v):\n%s", mode, runErr, text)
	}
	if strings.Contains(text, "[build failed]") || strings.Contains(text, "setup failed") || strings.Contains(text, "cannot find package") {
		t.Fatalf("overlay mutant %s failed before its semantic assertion:\n%s", mode, text)
	}
	t.Logf("reached %s mutant and named semantic assertion in child root %s: %s", mode, root, phase37MutationAssertionLine(text, expectedAssertion))
}

func phase37MutationAssertionLine(output, marker string) string {
	for _, line := range strings.Split(output, "\n") {
		if strings.Contains(line, marker) {
			return strings.TrimSpace(line)
		}
	}
	return "semantic assertion marker missing"
}

func phase37MutationEnvironment(root, mode string) []string {
	keys := map[string]string{
		"SCHWAY_PHASE37_MUTANT":        mode,
		"SCHWAY_PHASE37_EXPECTED_ROOT": root,
		"GOCACHE":                      "/tmp/ai-lang-verification-gocache",
	}
	environment := make([]string, 0, len(os.Environ())+len(keys))
	for _, entry := range os.Environ() {
		name, _, _ := strings.Cut(entry, "=")
		if _, replaced := keys[name]; !replaced {
			environment = append(environment, entry)
		}
	}
	for key, value := range keys {
		environment = append(environment, key+"="+value)
	}
	return environment
}

func phase37AssertMutationChildRoot(t *testing.T) {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Clean(filepath.Join(cwd, "..", "..", ".."))
	if expected := os.Getenv("SCHWAY_PHASE37_EXPECTED_ROOT"); expected == "" || expected != root {
		t.Fatalf("PHASE37_CHILD_ROOT_MISMATCH: got %s want %s", root, expected)
	}
	t.Logf("PHASE37_CHILD_ROOT=%s", root)
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
