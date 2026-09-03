package session

import (
	"context"
	"fmt"
	"os"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

type CheckResult struct {
	Tree        syntax.Tree
	Program     core.Program
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

func CheckFile(path string) (CheckResult, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return CheckResult{}, err
	}
	return Check(source), nil
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
