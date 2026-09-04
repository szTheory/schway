package native

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/execution"
)

func TestNativeStreamsIndependentlyBounded(t *testing.T) {
	tests := []struct {
		name string
		mode string
		code string
	}{
		{name: "compile stdout", mode: "compile-stdout-flood", code: "native.compile_stdout_truncated"},
		{name: "compile stderr", mode: "compile-stderr-flood", code: "native.compile_stderr_truncated"},
		{name: "run stdout", mode: "run-stdout-flood", code: "native.run_stdout_truncated"},
		{name: "run stderr", mode: "run-stderr-flood", code: "native.run_stderr_truncated"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			runner := Runner{command: func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				calls++
				stage := "compile"
				if calls > 1 {
					stage = "run"
				}
				command := exec.CommandContext(ctx, os.Args[0], "-test.run=TestNativeHelperProcess", "--", test.mode, stage)
				command.Env = append(os.Environ(), "GO_WANT_NATIVE_HELPER=1")
				return command
			}}
			_, err := runner.Run(context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"})
			var toolError *ToolError
			if !errors.As(err, &toolError) || toolError.Code != test.code {
				t.Fatalf("code=%v want=%s err=%v", toolError, test.code, err)
			}
		})
	}
}

// unboundedSpawnAllowlist is intentionally empty. Every Go file in this module
// outside .planning/spikes/ must spawn processes through exec.CommandContext
// with independently bounded stdout and stderr streams. Any entry added here
// must carry an explicit written justification; an unjustified entry is a
// review failure, not a waiver.
var unboundedSpawnAllowlist = map[string]string{}

// TestSourceNeverSpawnsUnboundedProcesses is the repo-wide successor to the
// former native.go-only single-file grep. It scans every *.go file in the
// module for the two patterns that defeat the Phase 02 bounded-stream +
// deadline discipline: the merged-output helper (which reads both streams
// into one unbounded buffer) and the context-free spawn constructor (which
// starts a child with no deadline). Both needles are assembled at runtime so
// this scanner never matches its own source.
// Only .planning/spikes/ is excluded, because throwaway spike labs are not
// part of the compiler or its test support.
func TestSourceNeverSpawnsUnboundedProcesses(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller path unavailable")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	// Assembled at runtime so this scanner does not match its own source.
	forbidden := []struct{ needle, reason string }{
		{"Combined" + "Output", "merges stdout and stderr into one unbounded buffer"},
		{"exec." + "Command(", "spawns a process without a context deadline (use exec.CommandContext)"},
	}
	spikes := filepath.Join(root, ".planning", "spikes") + string(filepath.Separator)
	scanned := 0
	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".go" || strings.HasPrefix(path, spikes) {
			return nil
		}
		relative, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		relative = filepath.ToSlash(relative)
		source, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		scanned++
		for _, pattern := range forbidden {
			if !strings.Contains(string(source), pattern.needle) {
				continue
			}
			if reason, allowed := unboundedSpawnAllowlist[relative]; allowed {
				t.Logf("allowlisted %s (%s): %s", relative, pattern.needle, reason)
				continue
			}
			t.Errorf("%s uses %s, which %s", relative, pattern.needle, pattern.reason)
		}
		return nil
	})
	if walkErr != nil {
		t.Fatal(walkErr)
	}
	if scanned == 0 {
		t.Fatal("scanner found no Go sources; the guard would pass vacuously")
	}
}

func TestExecutionDecoderRejectsMalformedOutput(t *testing.T) {
	valid, err := execution.CanonicalBytes(execution.Execution{
		Schema:  execution.Schema1,
		Outcome: execution.Outcome{Kind: "returned", Value: "01020304"},
		Events:  []execution.Event{}, LiveResources: []string{},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		data []byte
		code string
	}{
		{name: "missing document fields", data: []byte(`{}`), code: "native.invalid_execution"},
		{name: "unknown field", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[],"unknown":true}`), code: "native.invalid_execution"},
		{name: "duplicate document", data: append(append([]byte{}, valid...), valid...), code: "native.trailing_execution"},
		{name: "trailing value", data: append(append([]byte{}, valid...), []byte(` true`)...), code: "native.trailing_execution"},
		{name: "malformed", data: []byte(`{"schema":`), code: "native.invalid_execution"},
		{name: "truncated", data: append([]byte{}, valid[:len(valid)-1]...), code: "native.invalid_execution"},
		{name: "oversized", data: append(append([]byte{}, valid...), []byte(strings.Repeat(" ", MaxStreamBytes))...), code: "native.run_stdout_truncated"},
		{name: "duplicate key", data: []byte(`{"schema":"lang.execution/1","schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown schema", data: []byte(`{"schema":"lang.execution/9","outcome":{"kind":"returned","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown outcome", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"mystery","value":"x"},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "empty outcome", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":""},"events":[],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "unknown event", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"e","kind":"mystery","function_id":"f"}],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "missing transition fields", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"e","kind":"value.copied","function_id":"f"}],"live_resources":[]}`), code: "native.invalid_execution"},
		{name: "return before transition", data: []byte(`{"schema":"lang.execution/1","outcome":{"kind":"returned","value":"x"},"events":[{"schema":"lang.execution/1","id":"r","kind":"function.returned","function_id":"f","source_place":"p","type_id":"t"},{"schema":"lang.execution/1","id":"e","kind":"value.copied","function_id":"f","source_place":"p","target_place":"q","type_id":"t"}],"live_resources":[]}`), code: "native.invalid_execution"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, decodeErr := decodeExecution(test.data)
			var toolError *ToolError
			if !errors.As(decodeErr, &toolError) || toolError.Code != test.code {
				t.Fatalf("code=%v want=%s err=%v", toolError, test.code, decodeErr)
			}
		})
	}
}

func TestNativeToolFailureIsOperational(t *testing.T) {
	_, err := Runner{ClangPath: filepath.Join(t.TempDir(), "missing-clang")}.Run(
		context.Background(), "int main(void) { return 0; }", "-O0", []string{"input"},
	)
	var toolError *ToolError
	if !errors.As(err, &toolError) || toolError.Code != "native.tool_missing" {
		t.Fatalf("expected native.tool_missing, got %T %v", err, err)
	}
}

func TestNativeHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_NATIVE_HELPER") != "1" {
		return
	}
	mode, stage := os.Args[len(os.Args)-2], os.Args[len(os.Args)-1]
	flood := strings.Repeat("x", MaxStreamBytes+2)
	if stage == "compile" {
		switch mode {
		case "compile-stdout-flood":
			_, _ = os.Stdout.WriteString(flood)
		case "compile-stderr-flood":
			_, _ = os.Stderr.WriteString(flood)
		}
		return
	}
	valid, _ := execution.CanonicalBytes(execution.Execution{
		Schema:  execution.Schema0,
		Outcome: execution.Outcome{Kind: "returned", Value: "input"},
		Events:  []execution.Event{}, LiveResources: []string{},
	})
	switch mode {
	case "run-stdout-flood":
		_, _ = os.Stdout.WriteString(flood)
	case "run-stderr-flood":
		_, _ = os.Stdout.Write(valid)
		_, _ = os.Stderr.WriteString(flood)
	default:
		_, _ = os.Stdout.Write(valid)
	}
}
