package native

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type ToolError struct {
	Code string
	Err  error
}

func (e *ToolError) Error() string { return e.Code + ": " + e.Err.Error() }
func (e *ToolError) Unwrap() error { return e.Err }

type Pair struct {
	Input  string
	Output string
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

	ctx, cancel := context.WithTimeout(parent, r.Timeout)
	defer cancel()
	started := time.Now()
	command := exec.CommandContext(ctx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, sourcePath, "-o", binaryPath)
	compileOutput, err := command.CombinedOutput()
	compileTime := time.Since(started)
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return Result{}, &ToolError{Code: "native.timeout", Err: ctx.Err()}
	}
	if err != nil {
		code := "native.compile_failed"
		if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
			code = "native.tool_missing"
		}
		return Result{}, &ToolError{Code: code, Err: fmt.Errorf("%w: %s", err, bounded(string(compileOutput), 2048))}
	}

	result := Result{Optimization: optimization, CompileTime: compileTime, Pairs: make([]Pair, 0, len(inputs))}
	for _, input := range inputs {
		runStarted := time.Now()
		runCtx, runCancel := context.WithTimeout(parent, r.Timeout)
		runCommand := exec.CommandContext(runCtx, binaryPath, input)
		output, runErr := runCommand.CombinedOutput()
		runCancel()
		result.RunTime += time.Since(runStarted)
		result.OutputBytes += len(output)
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
			return Result{}, &ToolError{Code: "native.timeout", Err: runCtx.Err()}
		}
		if runErr != nil {
			return Result{}, &ToolError{Code: "native.run_failed", Err: fmt.Errorf("%w: %s", runErr, bounded(string(output), 2048))}
		}
		result.Pairs = append(result.Pairs, Pair{Input: input, Output: strings.TrimSpace(string(output))})
	}
	return result, nil
}

func bounded(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit] + "…"
}
