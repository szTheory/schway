package testsupport

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// MaxCLIStreamBytes bounds each captured child stream independently, mirroring
// native.MaxStreamBytes and evidence.MaxToolProbeBytes. It is larger than
// those production bounds because CLI tests legitimately capture whole `--json
// verify` corpora, but it is still a hard ceiling: a runaway child cannot
// exhaust the test process.
const MaxCLIStreamBytes = 8 << 20

// BuildCLITimeout and RunCLITimeout give every spawn a deadline so a hung
// child fails with a typed timeout instead of hanging the whole test binary.
const (
	BuildCLITimeout = 5 * time.Minute
	RunCLITimeout   = 2 * time.Minute
)

// CLIError is the stable typed failure of a test-support spawn. Codes are
// stable strings so callers can assert on them the way native.ToolError and
// evidence.ToolProbeError codes are asserted.
type CLIError struct {
	Code string
	Err  error
}

func (e *CLIError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *CLIError) Unwrap() error { return e.Err }

// boundedWriter captures at most MaxCLIStreamBytes+1 bytes. The extra
// detection byte makes overflow observable without silently truncating into a
// plausible-looking result: total is the true byte count, and any caller that
// overflows fails rather than reading a truncated buffer.
type boundedWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *boundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := MaxCLIStreamBytes + 1 - w.buffer.Len()
	if remaining > len(data) {
		remaining = len(data)
	}
	if remaining > 0 {
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *boundedWriter) bytes() []byte    { return w.buffer.Bytes() }
func (w *boundedWriter) overflowed() bool { return w.total > MaxCLIStreamBytes }

func streamError(code string) error {
	return &CLIError{Code: code, Err: fmt.Errorf("process stream exceeded %d bytes", MaxCLIStreamBytes)}
}

func ProjectPath(parts ...string) string {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	return filepath.Join(append([]string{root}, parts...)...)
}

// BuildCLI builds ./cmd/lang into a temporary directory, failing the test on
// any build, timeout, or stream-overflow error.
func BuildCLI(t testing.TB) string {
	t.Helper()
	binary, err := BuildCLIErr(context.Background(), filepath.Join(t.TempDir(), "lang"))
	if err != nil {
		t.Fatalf("build CLI: %v", err)
	}
	return binary
}

// BuildCLIErr is the goroutine-safe form of BuildCLI: it reports failures as
// typed errors instead of calling t.Fatalf, which may only be called from the
// goroutine running the test.
func BuildCLIErr(parent context.Context, binary string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, BuildCLITimeout)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "build", "-o", binary, "./cmd/lang")
	command.Dir = ProjectPath()
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", &CLIError{Code: "testsupport.build_timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() {
		return "", streamError("testsupport.build_stdout_truncated")
	}
	if stderr.overflowed() {
		return "", streamError("testsupport.build_stderr_truncated")
	}
	if err != nil {
		return "", &CLIError{Code: "testsupport.build_failed", Err: fmt.Errorf("%w\n%s%s", err, stdout.bytes(), stderr.bytes())}
	}
	return binary, nil
}

type CLIResult struct {
	Stdout []byte
	Stderr []byte
	Exit   int
}

// RunCLI runs the built binary, failing the test on any spawn, timeout, or
// stream-overflow error. A nonzero child exit is returned in the result rather
// than failing, because CLI tests assert on exit codes.
func RunCLI(t testing.TB, binary string, environment []string, arguments ...string) CLIResult {
	t.Helper()
	result, err := RunCLIErr(context.Background(), binary, environment, arguments...)
	if err != nil {
		t.Fatalf("run CLI: %v", err)
	}
	return result
}

// RunCLIErr is the goroutine-safe form of RunCLI.
func RunCLIErr(parent context.Context, binary string, environment []string, arguments ...string) (CLIResult, error) {
	ctx, cancel := context.WithTimeout(parent, RunCLITimeout)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	if environment == nil {
		command.Env = os.Environ()
	} else {
		command.Env = environment
	}
	var stdout, stderr boundedWriter
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return CLIResult{}, &CLIError{Code: "testsupport.run_timeout", Err: ctx.Err()}
	}
	if stdout.overflowed() {
		return CLIResult{}, streamError("testsupport.run_stdout_truncated")
	}
	if stderr.overflowed() {
		return CLIResult{}, streamError("testsupport.run_stderr_truncated")
	}
	exit := 0
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return CLIResult{}, &CLIError{Code: "testsupport.run_failed", Err: err}
		}
		exit = exitError.ExitCode()
	}
	return CLIResult{Stdout: stdout.bytes(), Stderr: stderr.bytes(), Exit: exit}, nil
}
