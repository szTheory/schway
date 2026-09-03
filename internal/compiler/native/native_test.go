package native

import (
	"context"
	"errors"
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

func TestNativeNeverUsesCombinedOutput(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller path unavailable")
	}
	source, err := os.ReadFile(filepath.Join(filepath.Dir(file), "native.go"))
	if err != nil {
		t.Fatal(err)
	}
	forbidden := "Combined" + "Output"
	if strings.Contains(string(source), forbidden) {
		t.Fatalf("native runner merges stdout and stderr through %s", forbidden)
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
