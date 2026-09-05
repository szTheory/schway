package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/codename-lang/lang/internal/compiler/execution"
)

const MaxStreamBytes = 64 * 1024

type ToolError struct {
	Code string
	Err  error
}

func (e *ToolError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *ToolError) Unwrap() error { return e.Err }

type Pair struct {
	Input     string
	Execution execution.Execution
}

type Result struct {
	Optimization string
	Pairs        []Pair
	CompileTime  time.Duration
	RunTime      time.Duration
	OutputBytes  int
}

type Runner struct {
	ClangPath string
	Timeout   time.Duration
	// Expect names the terminal outcome this runner's caller expects every
	// execution document to carry (D-04-08). The zero value behaves exactly
	// as ExpectValue, so every pre-Phase-4 caller that never sets this field
	// keeps the original "returned"-only contract unchanged.
	Expect TerminalOutcome
	// ForeignSources names additional C files (the frozen foreign
	// translation unit, D-04-10) to compile as their own separate, bounded,
	// timed invocations and link into the final binary. Empty for every
	// pre-Phase-4 caller, which keeps the original single-file compile-and-
	// link invocation byte-for-byte unchanged.
	ForeignSources []string
	command        func(context.Context, string, ...string) *exec.Cmd
}

func DefaultRunner() Runner { return Runner{ClangPath: "clang", Timeout: 5 * time.Second} }

func (r Runner) Run(parent context.Context, cSource, optimization string, inputs []string) (Result, error) {
	if r.ClangPath == "" {
		r.ClangPath = "clang"
	}
	if r.Timeout <= 0 {
		r.Timeout = 5 * time.Second
	}
	directory, err := os.MkdirTemp("", "lang-native-")
	if err != nil {
		return Result{}, &ToolError{Code: "native.temp_failed", Err: err}
	}
	defer os.RemoveAll(directory)

	sourcePath := filepath.Join(directory, "program.c")
	binaryPath := filepath.Join(directory, "program")
	if err := os.WriteFile(sourcePath, []byte(cSource), 0o600); err != nil {
		return Result{}, &ToolError{Code: "native.temp_failed", Err: err}
	}

	// Per D-04-10, the frozen foreign translation unit is compiled as its
	// own separate, bounded, timed invocation -- never merged into one
	// clang invocation with program.c's own compile -- and only the
	// resulting object is linked into the final binary below.
	objectPaths := make([]string, 0, len(r.ForeignSources))
	for index, foreignSource := range r.ForeignSources {
		objectPath := filepath.Join(directory, fmt.Sprintf("foreign_%d.o", index))
		foreignCtx, foreignCancel := context.WithTimeout(parent, r.Timeout)
		foreignCommand := r.commandContext(foreignCtx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, "-c", foreignSource, "-o", objectPath)
		var foreignStdout, foreignStderr boundedWriter
		foreignCommand.Stdout = &foreignStdout
		foreignCommand.Stderr = &foreignStderr
		foreignErr := foreignCommand.Run()
		deadlineExceeded := errors.Is(foreignCtx.Err(), context.DeadlineExceeded)
		foreignCancel()
		if deadlineExceeded {
			return Result{}, &ToolError{Code: "native.timeout", Err: foreignCtx.Err()}
		}
		if foreignStdout.overflowed() {
			return Result{}, streamError("native.compile_stdout_truncated")
		}
		if foreignStderr.overflowed() {
			return Result{}, streamError("native.compile_stderr_truncated")
		}
		if foreignErr != nil {
			code := "native.compile_failed"
			if errors.Is(foreignErr, exec.ErrNotFound) || errors.Is(foreignErr, os.ErrNotExist) {
				code = "native.tool_missing"
			}
			return Result{}, &ToolError{Code: code, Err: withStderr(foreignErr, foreignStderr.bytes())}
		}
		objectPaths = append(objectPaths, objectPath)
	}

	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	started := time.Now()
	arguments := append([]string{"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, sourcePath}, objectPaths...)
	arguments = append(arguments, "-o", binaryPath)
	command := r.commandContext(ctx, r.ClangPath, arguments...)
	var compileStdout, compileStderr boundedWriter
	command.Stdout = &compileStdout
	command.Stderr = &compileStderr
	err = command.Run()
	compileTime := time.Since(started)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Result{}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if compileStdout.overflowed() {
		return Result{}, streamError("native.compile_stdout_truncated")
	}
	if compileStderr.overflowed() {
		return Result{}, streamError("native.compile_stderr_truncated")
	}
	if err != nil {
		code := "native.compile_failed"
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return Result{}, &ToolError{Code: code, Err: withStderr(err, compileStderr.bytes())}
	}

	result := Result{Optimization: optimization, CompileTime: compileTime, Pairs: make([]Pair, 0, len(inputs))}
	for _, input := range inputs {
		runStarted := time.Now()
		runCtx, runCancel := context.WithTimeout(parent, r.Timeout)
		runCommand := r.commandContext(runCtx, binaryPath, input)
		var runStdout, runStderr boundedWriter
		runCommand.Stdout = &runStdout
		runCommand.Stderr = &runStderr
		runErr := runCommand.Run()
		runCancel()
		result.RunTime += time.Since(runStarted)
		result.OutputBytes += runStdout.total + runStderr.total
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return Result{}, &ToolError{Code: "native.timeout", Err: runCtx.Err()}
		}
		if runStdout.overflowed() {
			return Result{}, streamError("native.run_stdout_truncated")
		}
		if runStderr.overflowed() {
			return Result{}, streamError("native.run_stderr_truncated")
		}
		if runErr != nil {
			return Result{}, &ToolError{Code: "native.run_failed", Err: withStderr(runErr, runStderr.bytes())}
		}
		if len(runStderr.bytes()) != 0 {
			return Result{}, &ToolError{Code: "native.run_stderr", Err: fmt.Errorf("native process wrote stderr: %s", bounded(string(runStderr.bytes()), 2048))}
		}
		expect := r.Expect
		if expect == "" {
			expect = ExpectValue
		}
		decoded, decodeErr := decodeExecution(runStdout.bytes(), expect)
		if decodeErr != nil {
			return Result{}, decodeErr
		}
		result.Pairs = append(result.Pairs, Pair{Input: input, Execution: decoded})
	}
	return result, nil
}

func (r Runner) commandContext(ctx context.Context, name string, arguments ...string) *exec.Cmd {
	if r.command != nil {
		return r.command(ctx, name, arguments...)
	}
	return exec.CommandContext(ctx, name, arguments...)
}

type boundedWriter struct {
	buffer bytes.Buffer
	total  int
}

func (w *boundedWriter) Write(data []byte) (int, error) {
	w.total += len(data)
	remaining := MaxStreamBytes + 1 - w.buffer.Len()
	if remaining > 0 {
		if remaining > len(data) {
			remaining = len(data)
		}
		_, _ = w.buffer.Write(data[:remaining])
	}
	return len(data), nil
}

func (w *boundedWriter) bytes() []byte    { return w.buffer.Bytes() }
func (w *boundedWriter) overflowed() bool { return w.total > MaxStreamBytes }

func streamError(code string) error {
	return &ToolError{Code: code, Err: fmt.Errorf("process stream exceeded %d bytes", MaxStreamBytes)}
}

func withStderr(err error, stderr []byte) error {
	if len(stderr) == 0 {
		return err
	}
	return fmt.Errorf("%w: %s", err, bounded(string(stderr), 2048))
}

func decodeExecution(stdout []byte, expect TerminalOutcome) (execution.Execution, error) {
	if len(stdout) > MaxStreamBytes {
		return execution.Execution{}, streamError("native.run_stdout_truncated")
	}
	if err := rejectDuplicateJSONKeys(stdout); err != nil {
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	var value execution.Execution
	decoder := json.NewDecoder(bytes.NewReader(stdout))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			err = errors.New("trailing execution document")
		}
		return execution.Execution{}, &ToolError{Code: "native.trailing_execution", Err: err}
	}
	if err := validateExecution(value, ExpectValue); err != nil {
		return execution.Execution{}, &ToolError{Code: "native.invalid_execution", Err: err}
	}
	return value, nil
}

// TerminalOutcome names the closed axis validateExecution accepts (D-04-08:
// value | typed_failure | defect, with "defect" reserved for a later plan --
// see Pitfall 3, 04-RESEARCH.md). It exists so a caller states in advance
// which terminal shape it expects, rather than validateExecution silently
// discovering the shape from the document -- an execution document is
// validated AGAINST an expectation, never used to infer one, matching the
// project's general "never trust the producer" posture.
type TerminalOutcome string

const (
	ExpectValue        TerminalOutcome = "value"
	ExpectTypedFailure TerminalOutcome = "typed_failure"
)

func rejectDuplicateJSONKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var scan func() error
	scan = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]struct{})
			for decoder.More() {
				keyToken, keyErr := decoder.Token()
				if keyErr != nil {
					return keyErr
				}
				key, ok := keyToken.(string)
				if !ok {
					return errors.New("object key is not a string")
				}
				if _, duplicate := seen[key]; duplicate {
					return fmt.Errorf("duplicate JSON key %q", key)
				}
				seen[key] = struct{}{}
				if err := scan(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		case '[':
			for decoder.More() {
				if err := scan(); err != nil {
					return err
				}
			}
			_, err = decoder.Token()
			return err
		default:
			return errors.New("unexpected JSON delimiter")
		}
	}
	return scan()
}

// validateExecution asserts the execution document against the caller's
// declared expectation (D-04-08's closed value|typed_failure|defect axis;
// "defect" is not yet constructible and is intentionally absent from this
// switch -- a document claiming it is rejected by the default case below).
// Per Pitfall 3 (04-RESEARCH.md), the pre-Phase-4 "value" contract is kept
// byte-for-byte: every one of its five original rejection grounds still
// rejects (TestValidateExecutionStillRejectsOldGrounds), this function
// simply no longer treats every OTHER expectation as automatically invalid.
func validateExecution(value execution.Execution, expect TerminalOutcome) error {
	if value.Schema != execution.Schema0 && value.Schema != execution.Schema1 {
		return errors.New("unsupported execution schema")
	}
	switch expect {
	case ExpectValue, "":
		if value.Outcome.Kind != "returned" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 || len(value.LiveResources) != 0 {
			return errors.New("execution document violates required contract")
		}
	case ExpectTypedFailure:
		if value.Outcome.Kind != "typed_failure" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 {
			return errors.New("execution document violates required contract")
		}
	default:
		return fmt.Errorf("unsupported expected terminal outcome %q", expect)
	}
	seenIDs := make(map[string]struct{}, len(value.Events))
	functionID := value.Events[0].FunctionID
	for index, event := range value.Events {
		if event.Schema != value.Schema || event.ID == "" || event.FunctionID == "" || event.FunctionID != functionID {
			return errors.New("execution event identity or schema mismatch")
		}
		if _, duplicate := seenIDs[event.ID]; duplicate {
			return errors.New("duplicate execution event id")
		}
		seenIDs[event.ID] = struct{}{}
		isLast := index == len(value.Events)-1
		switch event.Kind {
		case "function.returned":
			if !isLast || event.TargetPlace != "" {
				return errors.New("return event must be last and have no target")
			}
			if value.Schema == execution.Schema0 {
				if event.Input == "" || event.Output == "" || event.SourcePlace != "" || event.TypeID != "" {
					return errors.New("match return event fields are invalid")
				}
			} else if event.SourcePlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("linear return event fields are invalid")
			}
		case "value.copied", "value.transferred", "value.borrowed", "value.borrowed_exclusive", "foreign.called":
			if value.Schema != execution.Schema1 || isLast || event.SourcePlace == "" || event.TargetPlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("linear transition event fields are invalid")
			}
		case "function.failed":
			if !isLast || value.Schema != execution.Schema1 || event.SourcePlace == "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("typed-failure event fields are invalid")
			}
		case "resource.released":
			// D-04-07's non-terminal release transition: like the other
			// linear transitions it is never last (a terminator always
			// follows), but unlike them it produces no new place, so
			// TargetPlace must stay empty rather than required.
			if value.Schema != execution.Schema1 || isLast || event.SourcePlace == "" || event.TargetPlace != "" || event.TypeID == "" || event.Input != "" || event.Output != "" {
				return errors.New("resource release event fields are invalid")
			}
		default:
			return errors.New("unknown execution event kind")
		}
	}
	return nil
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
