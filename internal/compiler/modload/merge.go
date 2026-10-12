package modload

import (
	"fmt"
	"sort"

	"github.com/szTheory/schway/internal/compiler/ast"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const (
	codeProgramLimit      = "module.program_limit"
	codeImportedData      = "module.imported_data_unsupported"
	codeImportedForeign   = "module.imported_foreign_unsupported"
	codeInvalidImportName = "module.invalid_import_name"
	codeCountLimit        = "module.count_limit"
	codeDepthLimit        = "module.depth_limit"
	codeTotalBytesLimit   = "module.total_bytes_limit"
	codePathMismatch      = "module.path_mismatch"
	codeDuplicate         = "module.duplicate"
	codeImportCycle       = "module.import_cycle"
	causeKindModuleHeader = "module_header"
	causeKindImportEdge   = "import_edge"
)

// parsedModule is one parsed module file. name is the canonical dotted path
// name, "" for the entry.
type parsedModule struct {
	name    string
	program ast.Program
	tree    syntax.Tree
	source  Source
}

// mergeInImportOrderForTest appends imported modules in first-import order
// instead of canonical sorted order (D-12 control). Production leaves it false.
var mergeInImportOrderForTest bool

// merge composes one whole program (D-12, D-20, D-21). The entry's functions
// come first in source order with Module "", then every imported module sorted
// by name with its functions in source order and FuncDecl.Module set. Units[0]
// is the entry. The entry's data, foreign blocks, imports and exports stay on
// the Program; an imported module's data or foreign declarations are refused.
func merge(entry parsedModule, imported []parsedModule) (ast.Program, []diagnostic.Diagnostic) {
	ordered := append([]parsedModule(nil), imported...)
	if !mergeInImportOrderForTest {
		sort.SliceStable(ordered, func(a, b int) bool { return ordered[a].name < ordered[b].name })
	}

	merged := entry.program
	merged.Funcs = append([]ast.FuncDecl(nil), entry.program.Funcs...)
	merged.Units = []ast.Unit{{Name: "", Imports: entry.program.Imports, Exports: entry.program.Exports}}

	var diags []diagnostic.Diagnostic
	functions := len(entry.program.Funcs)
	declarations := functions + len(entry.program.Data)
	limited := false
	check := func(module parsedModule, before, beforeDecls int) {
		if limited {
			return
		}
		program := module.program
		overFunctions := before+len(program.Funcs) > syntax.MaxFunctions
		overDeclarations := beforeDecls+len(program.Funcs)+len(program.Data) > syntax.MaxDeclarations
		if !overFunctions && !overDeclarations {
			return
		}
		limited = true
		span := diagnostic.Span{}
		switch {
		case overFunctions:
			if index := syntax.MaxFunctions - before; index >= 0 && index < len(program.Funcs) {
				span = program.Funcs[index].Span
			}
		default:
			remaining := syntax.MaxDeclarations - beforeDecls
			switch {
			case remaining < 0:
			case remaining < len(program.Data):
				span = program.Data[remaining].Span
			case remaining-len(program.Data) < len(program.Funcs):
				span = program.Funcs[remaining-len(program.Data)].Span
			}
		}
		diags = append(diags, diagnostic.Error(codeProgramLimit, span,
			fmt.Sprintf("the merged program exceeds the limit of %d functions or %d declarations", syntax.MaxFunctions, syntax.MaxDeclarations),
		).InUnit(module.name))
	}
	check(entry, 0, 0)

	for _, module := range ordered {
		program := module.program
		if len(program.Data) > 0 {
			diags = append(diags, diagnostic.Error(codeImportedData, program.Data[0].Span,
				"an imported module cannot declare data: this phase admits data only in the entry module",
			).InUnit(module.name))
		}
		if len(program.Foreign) > 0 {
			diags = append(diags, diagnostic.Error(codeImportedForeign, program.Foreign[0].Span,
				"an imported module cannot declare foreign symbols: this phase admits foreign blocks only in the entry module",
			).InUnit(module.name))
		}
		check(module, functions, declarations)
		functions += len(program.Funcs)
		declarations += len(program.Funcs) + len(program.Data)
		for _, function := range program.Funcs {
			function.Module = module.name
			merged.Funcs = append(merged.Funcs, function)
		}
		merged.Units = append(merged.Units, ast.Unit{Name: module.name, Imports: program.Imports, Exports: program.Exports})
	}
	return merged, diags
}
