package check

import (
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/ast"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
)

type Result struct {
	Program     core.Program
	Diagnostics []diagnostic.Diagnostic
}

func Program(program ast.Program) Result {
	result := Result{Program: core.Program{Schema: core.Schema, Module: program.Module, ModuleID: semanticID(program.Module, "module", program.Module)}}
	types := make(map[string]core.DataType)
	for _, declaration := range program.Data {
		alternatives := make([]string, 0, len(declaration.Alternatives))
		for _, alternative := range declaration.Alternatives {
			alternatives = append(alternatives, alternative.Name)
		}
		dataType := core.DataType{ID: semanticID(program.Module, "type", declaration.Name), Name: declaration.Name, Alternatives: alternatives, Span: declaration.Span}
		types[declaration.Name] = dataType
		result.Program.DataTypes = append(result.Program.DataTypes, dataType)
	}

	for _, function := range program.Funcs {
		functionID := semanticID(program.Module, "fn", function.Name)
		matchID := semanticID(program.Module, "match", function.Name)
		dataType, ok := types[function.Parameter.Type]
		if !ok {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.unknown", function.Parameter.Span, "unknown parameter type"))
			continue
		}
		if function.ReturnType != function.Parameter.Type {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.return_mismatch", function.Span, "S1 match result must have the parameter type"))
			continue
		}
		if function.Body.Scrutinee != function.Parameter.Name {
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("name.unknown_scrutinee", function.Body.Span, "match scrutinee is not the function parameter"))
			continue
		}
		seen := make(map[string]bool)
		arms := make([]core.MatchArm, 0, len(function.Body.Arms))
		for index, arm := range function.Body.Arms {
			if seen[arm.Pattern] {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.subsumed", arm.Span, "alternative is already matched"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Pattern) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
				continue
			}
			if !contains(dataType.Alternatives, arm.Value) {
				result.Diagnostics = append(result.Diagnostics, diagnostic.Error("type.invalid_variant", arm.Span, "match result is not an alternative of the return type"))
				continue
			}
			seen[arm.Pattern] = true
			arms = append(arms, core.MatchArm{
				ID: fmt.Sprintf("%s:arm:%d", functionID, index), EdgeID: fmt.Sprintf("%s:edge:%s", matchID, arm.Pattern),
				Pattern: arm.Pattern, Value: arm.Value,
			})
		}
		missing := make([]string, 0)
		for _, alternative := range dataType.Alternatives {
			if !seen[alternative] {
				missing = append(missing, alternative)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			causes := make([]diagnostic.Cause, 0, len(missing))
			for _, name := range missing {
				causes = append(causes, diagnostic.Cause{Kind: "missing_alternative", Detail: name})
			}
			result.Diagnostics = append(result.Diagnostics, diagnostic.Error("match.non_exhaustive", function.Body.Span, "match does not cover every alternative", causes...))
		}
		result.Program.Functions = append(result.Program.Functions, core.Function{
			ID: functionID, Name: function.Name,
			EntryPointID: functionID + ":point:entry", ReturnPointID: functionID + ":point:return",
			Parameter:  core.Parameter{ID: semanticID(program.Module, "parameter", function.Name+"."+function.Parameter.Name), Name: function.Parameter.Name, Type: function.Parameter.Type},
			ReturnType: function.ReturnType,
			Match:      core.Match{ID: matchID, PointID: functionID + ":point:match", Scrutinee: function.Body.Scrutinee, Arms: arms},
			Span:       function.Span,
		})
	}
	return result
}

func semanticID(module, kind, name string) string { return "s1:" + module + ":" + kind + ":" + name }

func contains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}
