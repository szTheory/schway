// Package surfacemanifest publishes the small, source-free description of
// checked language facts that an authoring client may use for early feedback.
// It is a leaf consumer: compiler admission never imports this package.
package surfacemanifest

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/evidence"
	"github.com/szTheory/schway/internal/compiler/execution"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surface"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const Schema = "schway.surface/1"

type Manifest struct {
	Schema       string         `json:"schema"`
	Coordinates  Coordinates    `json:"coordinates"`
	Types        []TypeEntry    `json:"types"`
	Operators    []Entry        `json:"operators"`
	Statements   []Entry        `json:"statements"`
	Builtins     []BuiltinEntry `json:"builtins"`
	Calls        CallRules      `json:"calls"`
	Restrictions []Entry        `json:"restrictions"`
	Limits       []Limit        `json:"limits"`
	Refusals     []Refusal      `json:"refusals"`
}

type Coordinates struct {
	Offsets    string `json:"offsets"`
	End        string `json:"end"`
	LineBase   int    `json:"line_base"`
	ColumnBase int    `json:"column_base"`
	ColumnUnit string `json:"column_unit"`
}

type Entry struct {
	ID          string   `json:"id"`
	Syntax      string   `json:"syntax"`
	Contexts    []string `json:"contexts"`
	Constraints []string `json:"constraints"`
}

type TypeEntry struct {
	Entry
	Arity int `json:"arity"`
}

type BuiltinEntry struct {
	ID          string   `json:"id"`
	Arity       int      `json:"arity"`
	Inputs      []string `json:"inputs"`
	Result      string   `json:"result"`
	Contexts    []string `json:"contexts"`
	Constraints []string `json:"constraints"`
}

type CallRules struct {
	DeclarationArity      int      `json:"declaration_arity"`
	OrdinaryArity         []int    `json:"ordinary_arity"`
	OrdinaryContexts      []string `json:"ordinary_contexts"`
	ScalarExpressionCalls bool     `json:"scalar_expression_calls"`
	Constraints           []string `json:"constraints"`
}

type Limit struct {
	ID       string   `json:"id"`
	Category string   `json:"category"`
	Scope    string   `json:"scope"`
	Unit     string   `json:"unit"`
	Value    int64    `json:"value"`
	Meaning  string   `json:"meaning"`
	Routes   []string `json:"routes"`
}

type Refusal struct {
	Code string `json:"code"`
	Form string `json:"form"`
	Hint string `json:"hint"`
}

func entry(id, form string, contexts []string, constraints ...string) Entry {
	return Entry{ID: id, Syntax: form, Contexts: contexts, Constraints: constraints}
}

func typed(id, form string, arity int, contexts []string, constraints ...string) TypeEntry {
	return TypeEntry{Entry: entry(id, form, contexts, constraints...), Arity: arity}
}

func builtin(id string, arity int, inputs []string, result string, contexts []string, constraints ...string) BuiltinEntry {
	return BuiltinEntry{ID: id, Arity: arity, Inputs: inputs, Result: result, Contexts: contexts, Constraints: constraints}
}

func limit(id, category, scope, unit string, value int, meaning string, routes ...string) Limit {
	return Limit{ID: id, Category: category, Scope: scope, Unit: unit, Value: int64(value), Meaning: meaning, Routes: routes}
}

// Current returns facts that are backed by the present parser, checker and
// command contracts. It intentionally does not enumerate grammar reachability
// or grant support to a syntax form that only has ability derivation.
func Current() Manifest {
	bounds := syntax.CurrentAuthoringBounds()
	moduleRoutes := check.ModuleAwareRoutes()
	interpreterFileRun := "run --engine=interpreter SOURCE -- PATH"
	nativeFileRun := "run --engine=native SOURCE --manifest MANIFEST -- PATH"
	parserRoutes := []string{"app verify", "build", "check", "debug-map", "evidence", "evidence --validate", "explain", "format", "interface core", "interface export", "query", "run", "verify"}
	checkedRoutes := []string{"app verify", "build", "check", "debug-map", "interface core", "interface export", "query", "run", "verify"}
	scalarRoutes := []string{"app verify", "build", "check", "debug-map", "evidence", "evidence --validate", "explain", "interface core", "interface export", "query", "run", "verify"}
	refusals := surface.Refusals()
	refusalEntries := make([]Refusal, 0, len(refusals))
	for _, refusal := range refusals {
		refusalEntries = append(refusalEntries, Refusal{Code: refusal.Code, Form: refusal.Form, Hint: refusal.Hint})
	}

	manifest := Manifest{
		Schema: Schema,
		Coordinates: Coordinates{
			Offsets:    "zero-based byte offsets",
			End:        "exclusive",
			LineBase:   1,
			ColumnBase: 1,
			ColumnUnit: "Unicode scalar values",
		},
		Types: []TypeEntry{
			typed("bool", "Bool", 0, []string{"scalar function input and result", "scalar local state", "scalar comparison results"}, "Bool result tails return an initialized Bool local or formal by identifier; Bool literals are not generally admitted as result tails."),
			typed("buffer", "Buffer", 0, []string{"linear executable input", "resource ownership and borrowing"}, "Noncopyable resource; ownership and loan discharge rules apply."),
			typed("byte", "Byte", 0, []string{"linear executable input", "byte-oriented checked values"}, "Only operations admitted for the current Byte contexts are described here."),
			typed("byte_view", "ByteView", 0, []string{"borrowed file-byte view", "scalar byte-reader parameter"}, "Read-only borrowed view; cannot be copied, shared, sent or returned."),
			typed("file_byte_owner", "FileByteOwner", 0, []string{"foreign adapter acquisition and release contract"}, "Resource use is restricted to the checked adapter protocol."),
			typed("file_bytes", "FileBytes", 0, []string{"FileBytes binding acquisition, borrowed view and release"}, "The maximum input size is selected by the checked binding configuration; no universal language-wide byte cap is asserted."),
			typed("fixture_bytes", "FixtureBytes", 0, []string{"fixture-only resource adapter"}, "This test fixture owner is not a general source-level resource constructor."),
			typed("path_token", "PathToken", 0, []string{"checked path-token parameter", "foreign resource adapter input"}, "Path use is restricted to the checked path and adapter contracts."),
			typed("sealed_nominal", "declared fieldless data type", 0, []string{"declared alternatives in exhaustive match contexts", "sealed nominal ability derivation"}, "Only checker-declared alternatives and their admitted branch contexts are included."),
			typed("unit", "Unit", 0, []string{"scalar no-result function return", "foreign release result"}, "Unit functions have no value result tail."),
			typed("u64", "U64", 0, []string{"linear executable input", "scalar function input and result", "scalar locals and operands"}, "Scalar operators and declarations are separately constrained below."),
		},
		Operators: []Entry{
			entry("u64.add", "+", []string{"scalar U64 operands"}, "Checked arithmetic reports overflow as a defect."),
			entry("u64.bit_and", "&", []string{"scalar U64 operands"}, "Both operands and the result are U64; Bool operands are refused."),
			entry("u64.bit_or", "|", []string{"scalar U64 operands"}, "Both operands and the result are U64; logical || remains refused."),
			entry("u64.bit_xor", "^", []string{"scalar U64 operands"}, "Both operands and the result are U64."),
			entry("u64.equal", "==", []string{"scalar U64 operands"}, "Only the checked scalar operand contexts are claimed."),
			entry("u64.less_than", "<", []string{"scalar U64 operands"}, "Only the checked scalar operand contexts are claimed."),
			entry("u64.remainder", "%", []string{"scalar U64 operands"}, "A statically zero divisor is refused; a runtime zero divisor is a defect."),
			entry("u64.shift_left", "<<", []string{"scalar U64 operands"}, "The amount must be an integer literal from 0 through 63; the result keeps the low 64 bits."),
			entry("u64.shift_right", ">>", []string{"scalar U64 operands"}, "The amount must be an integer literal from 0 through 63; right shift is logical."),
		},
		Statements: []Entry{
			entry("assign", "name = expression", []string{"initialized mutable scalar local"}, "Assignment does not grant mutation to immutable or resource bindings."),
			entry("borrow", "borrow place", []string{"linear ownership contexts"}, "Loan liveness and owner discharge remain checked."),
			entry("borrow_mut", "borrow mut place", []string{"linear mutable-loan contexts"}, "Exclusive loan and owner discharge remain checked."),
			entry("defect", "defect expression", []string{"checked defect and failure paths"}, "Only admitted typed defect forms are claimed."),
			entry("discard_because", "discard value because reason", []string{"checked fallible foreign/resource paths"}, "The value and reason must satisfy the checker’s declared failure contract."),
			entry("foreign_block", "foreign C { ... }", []string{"checked foreign policy declarations"}, "Foreign names, signatures, policies, adapters and resource discharge remain narrowly validated."),
			entry("if_else", "if condition { ... } else { ... }", []string{"scalar CFG control flow"}, "An if statement without an else block is refused."),
			entry("let", "let name = expression", []string{"scalar locals", "linear bindings and admitted calls"}, "Initializer, type and ownership rules are context-specific."),
			entry("match", "match value { ... }", []string{"declared exhaustive alternatives", "linear ownership branch contexts"}, "Arms and payloads must satisfy checked exhaustiveness and ownership rules."),
			entry("take", "take place", []string{"linear ownership transfer contexts"}, "The owner must be live and discharged according to the checked path."),
			entry("try", "try foreign_call(value)", []string{"checked fallible foreign adapter call"}, "Only the admitted fallible adapter shape and error alternative are supported."),
			entry("var", "var name = expression", []string{"initialized mutable scalar local"}, "Mutable scalar state remains subject to definite initialization."),
			entry("while", "while condition { ... }", []string{"scalar CFG loops"}, "Back edges must satisfy the scalar fixed-point and authority-free state rules."),
			entry("write", "write_text | write_u64 | write_newline", []string{"scalar function output statements"}, "Operands follow the individual builtin rules."),
		},
		Builtins: []BuiltinEntry{
			builtin("byte_at", 2, []string{"ByteView", "U64"}, "U64", []string{"scalar ByteView-parameter function"}, "Index outside the borrowed view is a defect."),
			builtin("byte_len", 1, []string{"ByteView"}, "U64", []string{"scalar ByteView-parameter function"}, "Reads only the bound ByteView parameter."),
			builtin("write_newline", 0, []string{}, "Unit", []string{"scalar output statement"}, "Takes no operands."),
			builtin("write_text", 1, []string{"quoted text literal"}, "Unit", []string{"scalar output statement"}, "Accepts one valid quoted fixed text literal only."),
			builtin("write_u64", 1, []string{"U64"}, "Unit", []string{"scalar output statement"}, "Accepts one U64 operand."),
		},
		Calls: CallRules{
			DeclarationArity:      1,
			OrdinaryArity:         []int{1},
			OrdinaryContexts:      []string{"admitted linear binding and match-arm call paths"},
			ScalarExpressionCalls: true,
			Constraints:           []string{"Ordinary calls require a checked signature and acyclic call graph.", "Call arguments and results must satisfy current type, ownership and authority checks.", "Scalar CFG expressions admit direct arity-one U64/Bool value calls with one argument; imported targets must be explicitly exported and whole-program call graphs remain acyclic.", "Foreign calls require their separate declared policy/adapter contract."},
		},
		Restrictions: []Entry{
			entry("bool_context", "true | false", []string{"retained Bool identifier and type paths"}, "Unbound Bool literals in result tails are refused; Bool is not a general literal/result feature."),
			entry("filebytes_limit", "FileBytes binding", []string{"checked build binding configuration"}, "The validated binding selects the maximum; this manifest does not publish a universal 65536-byte FileBytes language limit."),
			entry("source_file_run_input", "run --engine={interpreter|native} SOURCE -- PATH", []string{interpreterFileRun, nativeFileRun}, "After --, PATH is one opaque argument. The CLI refuses empty PATH, NUL, and values over 4096 bytes with exit 65; a 4096-byte value reaches the file adapter or operating system."),
			entry("module_loading", "import declarations", moduleRoutes, "Only the listed module-aware routes load imported source; other source routes retain their existing single-file or no-entry behavior."),
			entry("parse_only_and_ability_only", "grammar and ability facts", []string{"source parser", "sealed ability derivation"}, "Parsing or deriving abilities alone does not establish complete-program checker or engine admission."),
			entry("scalar_expression_grouping", "( expression )", []string{"scalar expressions"}, "Grouping is transparent, preserves the inner expression kind and literal provenance, admits the scalar call atom, and does not admit tuples or general tails."),
			entry("scalar_operator_precedence", "% > + > << >> > & > ^ > | > (<, ==)", []string{"scalar expressions"}, "Operators associate left within each tier; < and == retain their shared comparison tier."),
			entry("scalar_expression_calls", "function_call(expression)", []string{"nested scalar helper arguments", "scalar arithmetic and comparison operands", "scalar binding and assignment RHS", "scalar branch and loop conditions", "scalar byte_at index expressions"}, "A direct local function call has exactly one scalar expression argument and returns the declared U64 or Bool result.", "Only U64 -> U64, U64 -> Bool, Bool -> U64, and Bool -> Bool signatures are admitted; Unit and ByteView calls are absent.", "Imported targets must be explicitly exported local module functions; whole-program call graphs remain acyclic.", "Bool result tails return an initialized Bool local or formal by identifier; bind a call result before returning it.", "Argument recursion shares the 64-level expression-depth bound and lowers in source order."),
			entry("u64_hex_spelling", "0x... | 0X...", []string{"scalar U64 integer literals"}, "Hex digits are case-insensitive; the formatter preserves source case and between-digit underscores, and range is checked by value."),
			entry("unused_byte_scalar_entry", "fn main(unused: Byte) -> U64", []string{"public conformance run"}, "The Byte fixture parameter must remain unused and is never read, copied, or converted; scalar work uses U64 and Bool values."),
		},
		Limits: []Limit{
			limit("application.argument_bytes", "application input", "one application input argument", "bytes", native.MaxApplicationArgumentBytes, "Maximum byte length for the retained application input argument.", "app run"),
			limit("file_run.path_bytes", "source-file run input", "one opaque PathToken supplied after the file-run separator", "bytes", native.MaxApplicationArgumentBytes, "Maximum byte length for PATH; empty, NUL-containing or over-limit values are refused before file access.", interpreterFileRun, nativeFileRun),
			limit("interpreter.filebytes_bytes", "interpreter file input", "one regular file read by the fixed interpreter FileBytes adapter", "bytes", interp.FileBytesLimit, "The fixed interpreter adapter reads at most 65536 actual bytes and probes at most one extra byte; native binding-selected caps remain independent.", interpreterFileRun),
			limit("application.evidence_bytes", "application capture", "one same-run application evidence capture/report", "bytes", execution.MaxApplicationEvidenceBytes, "Maximum encoded application evidence bytes; separate from execution documents.", "app run"),
			limit("application.output_bytes", "application output", "one retained application stdout stream", "bytes", execution.MaxApplicationOutputBytes, "Maximum bytes streamed by the retained application.", "app run"),
			limit("check.work", "deterministic checker work", "one checked source program including its loaded modules", "work units", session.MaxCheckWork, "Token count plus explicit checker traversal work; exceeding the bound is refused. debug-map applies this guard only to import-bearing programs; query applies it only through lane/corpus queries.", checkedRoutes...),
			limit("evidence.manifest_bytes", "tool input", "one evidence manifest file", "bytes", evidence.MaxManifestBytes, "Maximum manifest bytes read for evidence validation.", "evidence --validate"),
			limit("evidence.tool_stream_bytes", "external tool capture", "each captured external-tool stdout or stderr stream", "bytes per stream", evidence.MaxToolProbeBytes, "Bounds an individual tool probe stream; overflow is reported rather than treated as complete evidence.", "evidence", "evidence --validate"),
			limit("execution.document_bytes", "execution capture", "one native execution document", "bytes", execution.MaxDocumentBytes, "Maximum native execution-document capture; distinct from application evidence.", "app verify"),
			limit("interface.artifact_bytes", "interface input", "each interface summary/core JSON input", "bytes per file", session.MaxInterfaceBytes, "Maximum bytes read from either checked interface artifact.", "interface check"),
			limit("module.depth", "module loading", "import chain below the entry module", "modules", modload.MaxImportDepth, "Entry depth is zero; this is the maximum import depth beneath it.", moduleRoutes...),
			limit("module.count", "module loading", "all distinct modules including the entry", "modules", modload.MaxModules, "Maximum number of loaded modules in a program.", moduleRoutes...),
			limit("module.imports_per_file", "source parser", "one module source file", "imports", syntax.MaxImportsPerModule, "Maximum import declarations in one module.", moduleRoutes...),
			limit("module.total_source_bytes", "module loading", "all source files in one loaded module program", "bytes", modload.MaxTotalSourceBytes, "Aggregate bytes read across the entry and imported modules.", moduleRoutes...),
			limit("parser.alternatives", "authoring parse bound", "one parsed source program", "alternatives", bounds.Alternatives, "Maximum parsed alternatives before the parser refuses the program.", parserRoutes...),
			limit("parser.declarations", "authoring parse bound", "one source file", "declarations", syntax.MaxDeclarations, "Maximum data and function declarations in one file.", parserRoutes...),
			limit("parser.expression_depth", "authoring parse bound", "one nested scalar expression including groups, byte_at indices and helper-call arguments", "levels", bounds.ExpressionDepth, "Maximum scalar expression nesting across groups, byte_at indices and helper-call arguments; deeper input receives surface.unsupported_expression_depth.", parserRoutes...),
			limit("parser.functions", "authoring parse bound", "one source file", "functions", syntax.MaxFunctions, "Maximum function declarations in one file.", parserRoutes...),
			limit("parser.linear_bindings", "authoring parse bound", "outer linear bodies across one parsed source program", "bindings", bounds.LinearBindings, "Maximum aggregate bindings in outer function linear-body lists; nested statement bindings are not counted by this guard.", parserRoutes...),
			limit("parser.primary_diagnostics", "authoring diagnostic bound", "one parsed source", "diagnostics", bounds.PrimaryDiagnostics, "Maximum primary parser diagnostics; further diagnostics are suppressed and reported as syntax.too_many_errors.", parserRoutes...),
			limit("parser.source_bytes", "authoring parse bound", "one source file", "bytes", syntax.MaxSourceBytes, "Maximum bytes accepted by a source parser entry point.", parserRoutes...),
			limit("parser.tokens", "authoring parse bound", "one source file", "tokens", syntax.MaxTokens, "Maximum tokens accepted from one parsed source file.", parserRoutes...),
			limit("parser.type_depth", "authoring type bound", "one nested type expression", "levels", bounds.TypeDepth, "Maximum parser type-expression nesting depth.", parserRoutes...),
			limit("parser.type_nodes", "authoring type bound", "one source file", "nodes", bounds.TypeNodes, "Maximum visited type nodes in parser bounds.", parserRoutes...),
			limit("scalar.analysis_work", "deterministic scalar analysis", "one scalar CFG function", "work units", check.ScalarAnalysisWorkLimit, "Maximum fixed-point work before scalar analysis refuses to converge.", scalarRoutes...),
		},
		Refusals: refusalEntries,
	}
	return canonical(manifest)
}

// JSON emits the typed document in its canonical field and array order. It
// sorts copies so callers retain ownership of their slices.
func JSON(manifest Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(canonical(manifest), "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode surface manifest: %w", err)
	}
	return append(data, '\n'), nil
}

func canonical(manifest Manifest) Manifest {
	manifest.Types = append([]TypeEntry{}, manifest.Types...)
	for i := range manifest.Types {
		manifest.Types[i].Contexts = sorted(manifest.Types[i].Contexts)
		manifest.Types[i].Constraints = sorted(manifest.Types[i].Constraints)
	}
	sort.Slice(manifest.Types, func(i, j int) bool { return manifest.Types[i].ID < manifest.Types[j].ID })
	manifest.Operators = canonicalEntries(manifest.Operators)
	manifest.Statements = canonicalEntries(manifest.Statements)
	manifest.Restrictions = canonicalEntries(manifest.Restrictions)
	manifest.Builtins = append([]BuiltinEntry{}, manifest.Builtins...)
	for i := range manifest.Builtins {
		manifest.Builtins[i].Inputs = sorted(manifest.Builtins[i].Inputs)
		manifest.Builtins[i].Contexts = sorted(manifest.Builtins[i].Contexts)
		manifest.Builtins[i].Constraints = sorted(manifest.Builtins[i].Constraints)
	}
	sort.Slice(manifest.Builtins, func(i, j int) bool { return manifest.Builtins[i].ID < manifest.Builtins[j].ID })
	manifest.Calls.OrdinaryArity = append([]int{}, manifest.Calls.OrdinaryArity...)
	sort.Ints(manifest.Calls.OrdinaryArity)
	manifest.Calls.OrdinaryContexts = sorted(manifest.Calls.OrdinaryContexts)
	manifest.Calls.Constraints = sorted(manifest.Calls.Constraints)
	manifest.Limits = append([]Limit{}, manifest.Limits...)
	for i := range manifest.Limits {
		manifest.Limits[i].Routes = sorted(manifest.Limits[i].Routes)
	}
	sort.Slice(manifest.Limits, func(i, j int) bool { return manifest.Limits[i].ID < manifest.Limits[j].ID })
	manifest.Refusals = append([]Refusal{}, manifest.Refusals...)
	sort.Slice(manifest.Refusals, func(i, j int) bool { return manifest.Refusals[i].Code < manifest.Refusals[j].Code })
	return manifest
}

func canonicalEntries(entries []Entry) []Entry {
	entries = append([]Entry{}, entries...)
	for i := range entries {
		entries[i].Contexts = sorted(entries[i].Contexts)
		entries[i].Constraints = sorted(entries[i].Constraints)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].ID < entries[j].ID })
	return entries
}

func sorted(values []string) []string {
	values = append([]string{}, values...)
	sort.Strings(values)
	return values
}
