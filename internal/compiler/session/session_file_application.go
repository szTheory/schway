package session

import (
	"fmt"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/interp"
)

// RunInterpreterFileApplication checks a source file and its imported module
// closure, independently admits the complete program, and then runs its
// checked FileBytes application through the interpreter's fixed host reader.
func RunInterpreterFileApplication(sourcePath, input string) (interp.ApplicationResult, []diagnostic.Diagnostic, error) {
	checked, _, err := checkSourcePath(sourcePath)
	if err != nil {
		return interp.ApplicationResult{}, nil, err
	}
	if len(checked.Diagnostics) != 0 {
		return interp.ApplicationResult{}, checked.Diagnostics, nil
	}
	admitted, diagnostics := admitProgram(checked)
	if len(diagnostics) != 0 {
		return interp.ApplicationResult{}, diagnostics, nil
	}
	if err := validateLocalOrFileBytesOwnerPaths(admitted); err != nil {
		return interp.ApplicationResult{}, []diagnostic.Diagnostic{diagnostic.Error("core.pathoracle_refused", diagnostic.Span{}, err.Error())}, nil
	}
	entry, program, entryDiagnostics, err := resolveEntry(admitted)
	if err != nil {
		return interp.ApplicationResult{}, nil, err
	}
	if len(entryDiagnostics) != 0 {
		return interp.ApplicationResult{}, entryDiagnostics, nil
	}
	if entry.Parameter.Type != "PathToken" {
		return interp.ApplicationResult{}, nil, fmt.Errorf("file application entry %q must accept a PathToken", entry.ID)
	}
	result, err := interp.RunFileApplication(program, entry.Name, input)
	if err != nil {
		return interp.ApplicationResult{}, nil, err
	}
	return result, nil, nil
}
