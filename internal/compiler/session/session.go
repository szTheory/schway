package session

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

type CheckResult struct {
	Tree        syntax.Tree
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
}

type FormatResult struct {
	Source      []byte
	Canonical   []byte
	Diagnostics []diagnostic.Diagnostic
}

type NativeResult struct {
	CSource     string
	Interpreter []interp.Execution
	O0          native.Result
	O3          native.Result
}

type EngineMismatch struct {
	Optimization string
	Input        string
	Expected     string
	Actual       string
}

func (e *EngineMismatch) Error() string {
	return fmt.Sprintf("native %s mismatch for %s: expected %s, got %s", e.Optimization, e.Input, e.Expected, e.Actual)
}

func Check(source []byte) CheckResult {
	parsed := syntax.Parse(source)
	result := CheckResult{Tree: parsed.Tree, Diagnostics: append([]diagnostic.Diagnostic(nil), parsed.Diagnostics...)}
	if len(result.Diagnostics) > 0 {
		return result
	}
	checked := check.Program(parsed.Program)
	result.Program = checked.Program
	result.Diagnostics = append(result.Diagnostics, checked.Diagnostics...)
	return result
}

func Format(source []byte) FormatResult {
	parsed := syntax.Parse(source)
	return FormatResult{Source: append([]byte(nil), source...), Canonical: syntax.Format(parsed.Tree), Diagnostics: parsed.Diagnostics}
}

func FormatFile(path string) (FormatResult, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return FormatResult{}, err
	}
	return Format(source), nil
}

func FormatCommandFile(path string, checkOnly bool) (protocol.Result, error) {
	formatted, err := FormatFile(path)
	if err != nil {
		return protocol.Result{}, err
	}
	result := protocol.New("format", protocol.StatusPass)
	if len(formatted.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = formatted.Diagnostics
		return result.Finalize(), nil
	}
	if checkOnly && !bytes.Equal(formatted.Source, formatted.Canonical) {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("format.non_canonical", diagnostic.Span{}, "source differs from canonical projection")}
		return result.Finalize(), nil
	}
	if !checkOnly {
		result.Formatted = string(formatted.Canonical)
	}
	return result.Finalize(), nil
}

func CheckFile(path string) (CheckResult, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return CheckResult{}, err
	}
	return Check(source), nil
}

func CheckCommandFile(path string) (protocol.Result, error) {
	checked, err := CheckFile(path)
	if err != nil {
		return protocol.Result{}, err
	}
	result := protocol.New("check", protocol.StatusPass)
	if len(checked.Diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = checked.Diagnostics
	} else {
		result.ModuleID = checked.Program.ModuleID
	}
	return result.Finalize(), nil
}

func RunInterpreter(source []byte) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return nil, checked.Diagnostics, nil
	}
	if len(checked.Program.DataTypes) != 1 || len(checked.Program.Functions) != 1 {
		return nil, nil, os.ErrInvalid
	}
	executions := make([]interp.Execution, 0, len(checked.Program.DataTypes[0].Alternatives))
	for _, input := range checked.Program.DataTypes[0].Alternatives {
		execution, err := interp.Run(checked.Program, checked.Program.Functions[0].Name, input)
		if err != nil {
			return nil, nil, err
		}
		executions = append(executions, execution)
	}
	return executions, nil, nil
}

func RunInterpreterFile(path string) ([]interp.Execution, []diagnostic.Diagnostic, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}
	return RunInterpreter(source)
}

func RunInterpreterCommandFile(path string) (protocol.Result, error) {
	executions, diagnostics, err := RunInterpreterFile(path)
	if err != nil {
		result := protocol.New("run", protocol.StatusOperational)
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("tool.run_failed", diagnostic.Span{}, "interpreter operation failed")}
		return result.Finalize(), nil
	}
	result := protocol.New("run", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
	} else {
		result.Executions = executions
	}
	return result.Finalize(), nil
}

func RunNative(ctx context.Context, source []byte, runner native.Runner) (NativeResult, []diagnostic.Diagnostic, error) {
	checked := Check(source)
	if len(checked.Diagnostics) > 0 {
		return NativeResult{}, checked.Diagnostics, nil
	}
	if len(checked.Program.DataTypes) != 1 || len(checked.Program.Functions) != 1 {
		return NativeResult{}, nil, os.ErrInvalid
	}
	inputs := checked.Program.DataTypes[0].Alternatives
	interpreted := make([]interp.Execution, 0, len(inputs))
	for _, input := range inputs {
		execution, err := interp.Run(checked.Program, checked.Program.Functions[0].Name, input)
		if err != nil {
			return NativeResult{}, nil, err
		}
		interpreted = append(interpreted, execution)
	}
	cSource, err := cgen.Emit(checked.Program)
	if err != nil {
		return NativeResult{}, nil, err
	}
	o0, err := runner.Run(ctx, cSource, "-O0", inputs)
	if err != nil {
		return NativeResult{}, nil, err
	}
	o3, err := runner.Run(ctx, cSource, "-O3", inputs)
	if err != nil {
		return NativeResult{}, nil, err
	}
	for index, expected := range interpreted {
		for _, actual := range []native.Result{o0, o3} {
			if actual.Pairs[index].Output != expected.Outcome.Value {
				return NativeResult{}, nil, &EngineMismatch{Optimization: actual.Optimization, Input: inputs[index], Expected: expected.Outcome.Value, Actual: actual.Pairs[index].Output}
			}
		}
	}
	return NativeResult{CSource: cSource, Interpreter: interpreted, O0: o0, O3: o3}, nil, nil
}

func RunNativeFile(ctx context.Context, path string, runner native.Runner) (NativeResult, []diagnostic.Diagnostic, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return NativeResult{}, nil, err
	}
	return RunNative(ctx, source, runner)
}

func RunNativeCommandFile(ctx context.Context, path string, runner native.Runner) (protocol.Result, error) {
	nativeResult, diagnostics, err := RunNativeFile(ctx, path, runner)
	result := protocol.New("run", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
		return result.Finalize(), nil
	}
	if err != nil {
		result.Status = protocol.StatusOperational
		code := "native.tool_failure"
		var mismatch *EngineMismatch
		if errors.As(err, &mismatch) {
			result.Status = protocol.StatusMismatch
			code = "native.engine_mismatch"
		}
		var toolError *native.ToolError
		if errors.As(err, &toolError) {
			code = toolError.Code
		}
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, "native execution did not complete successfully")}
		return result.Finalize(), nil
	}
	result.Executions = nativeResult.Interpreter
	return result.Finalize(), nil
}

func EvidenceCommandFile(ctx context.Context, path string) (evidence.Product, protocol.Result, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return evidence.Product{}, protocol.Result{}, err
	}
	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		result := protocol.New("evidence", protocol.StatusOperational)
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("evidence.tool_failure", diagnostic.Span{}, "unable to inspect native toolchain")}
		return evidence.Product{}, result.Finalize(), nil
	}
	product, diagnostics, err := evidence.Build(source, facts)
	if err != nil {
		return evidence.Product{}, protocol.Result{}, err
	}
	result := protocol.New("evidence", protocol.StatusPass)
	if len(diagnostics) > 0 {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = diagnostics
		return evidence.Product{}, result.Finalize(), nil
	}
	result.Evidence = &protocol.EvidenceSummary{Schema: evidence.Schema, ID: product.Manifest.ID, Digest: evidence.ContentDigest(product.ManifestBytes)}
	return product, result.Finalize(), nil
}

func ValidateEvidenceCommandFile(ctx context.Context, manifestPath, sourcePath string) protocol.Result {
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read evidence manifest")
	}
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read source")
	}
	manifest, err := evidence.DecodeStrict(manifestBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusInvalid, evidence.ErrorCode(err), "evidence manifest is not valid")
	}
	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "evidence.tool_failure", "unable to inspect native toolchain")
	}
	if err := evidence.Validate(manifest, source, facts); err != nil {
		return commandProblem("evidence", protocol.StatusInvalid, evidence.ErrorCode(err), "evidence manifest does not match recomputed facts")
	}
	result := protocol.New("evidence", protocol.StatusPass)
	result.Evidence = &protocol.EvidenceSummary{Schema: evidence.Schema, ID: manifest.ID, Digest: evidence.ContentDigest(manifestBytes)}
	return result.Finalize()
}

func commandProblem(command, status, code, message string) protocol.Result {
	result := protocol.New(command, status)
	result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(code, diagnostic.Span{}, message)}
	return result.Finalize()
}
