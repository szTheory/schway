package session

import (
	"os"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

type CheckResult struct {
	Tree        syntax.Tree
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
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
