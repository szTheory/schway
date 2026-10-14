package surfacemanifest

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestPhase34SurfaceDocument(t *testing.T) {
	manifest := Current()
	if manifest.Schema != "schway.surface/1" {
		t.Fatalf("schema = %q", manifest.Schema)
	}
	if manifest.Coordinates != (Coordinates{
		Offsets: "zero-based byte offsets", End: "exclusive", LineBase: 1, ColumnBase: 1, ColumnUnit: "Unicode scalar values",
	}) {
		t.Fatalf("coordinates = %+v", manifest.Coordinates)
	}
	if len(manifest.Types) == 0 || len(manifest.Operators) == 0 || len(manifest.Statements) == 0 || len(manifest.Builtins) == 0 || len(manifest.Restrictions) == 0 || len(manifest.Limits) == 0 || len(manifest.Refusals) != 7 {
		t.Fatalf("manifest has empty or incomplete required arrays: types=%d operators=%d statements=%d builtins=%d restrictions=%d limits=%d refusals=%d", len(manifest.Types), len(manifest.Operators), len(manifest.Statements), len(manifest.Builtins), len(manifest.Restrictions), len(manifest.Limits), len(manifest.Refusals))
	}

	wantTypeIDs := []string{"bool", "buffer", "byte", "byte_view", "file_byte_owner", "file_bytes", "fixture_bytes", "path_token", "sealed_nominal", "u64", "unit"}
	gotTypeIDs := make([]string, 0, len(manifest.Types))
	for _, item := range manifest.Types {
		gotTypeIDs = append(gotTypeIDs, item.ID)
		if item.Contexts == nil || item.Constraints == nil {
			t.Errorf("type %s has omitted arrays", item.ID)
		}
		assertSorted(t, "type "+item.ID+" contexts", item.Contexts)
		assertSorted(t, "type "+item.ID+" constraints", item.Constraints)
	}
	sort.Strings(wantTypeIDs)
	if !reflect.DeepEqual(gotTypeIDs, wantTypeIDs) {
		t.Fatalf("type IDs = %v, want %v", gotTypeIDs, wantTypeIDs)
	}
	for id, arity := range map[string]int{"bool": 0, "buffer": 0, "byte": 0, "byte_view": 0, "file_byte_owner": 0, "file_bytes": 0, "fixture_bytes": 0, "path_token": 0, "sealed_nominal": 0, "u64": 0, "unit": 0} {
		if got := typeByID(t, manifest.Types, id).Arity; got != arity {
			t.Errorf("type %s arity = %d, want %d", id, got, arity)
		}
	}
	// Complete-source witnesses in Phase 34-07 refuse these ability shapes;
	// they belong to the parse/ability restriction, not the admission list.
	abilityOnly := entryByID(t, manifest.Restrictions, "parse_only_and_ability_only")
	if !reflect.DeepEqual(abilityOnly.Contexts, []string{"sealed ability derivation", "source parser"}) || !reflect.DeepEqual(abilityOnly.Constraints, []string{"Parsing or deriving abilities alone does not establish complete-program checker or engine admission."}) {
		t.Fatalf("parse/ability restriction = %+v", abilityOnly)
	}
	if got := entryByID(t, manifest.Statements, "if_else").Contexts; !reflect.DeepEqual(got, []string{"scalar CFG control flow"}) {
		t.Fatalf("if/else contexts = %v, want only the witnessed scalar CFG context", got)
	}
	if !strings.Contains(strings.Join(entryByID(t, manifest.Restrictions, "bool_context").Constraints, " "), "not a general") {
		t.Fatal("Bool contexts overstate general literal/result admission")
	}
	for _, entries := range [][]Entry{manifest.Operators, manifest.Statements, manifest.Restrictions} {
		assertEntryOrder(t, entries)
		for _, item := range entries {
			if item.Contexts == nil || item.Constraints == nil {
				t.Errorf("entry %s has omitted arrays", item.ID)
			}
			assertSorted(t, item.ID+" contexts", item.Contexts)
			assertSorted(t, item.ID+" constraints", item.Constraints)
		}
	}
	wantStatementIDs := []string{"assign", "borrow", "borrow_mut", "defect", "discard_because", "foreign_block", "if_else", "let", "match", "take", "try", "var", "while", "write"}
	gotStatementIDs := make([]string, 0, len(manifest.Statements))
	for _, item := range manifest.Statements {
		gotStatementIDs = append(gotStatementIDs, item.ID)
	}
	if !reflect.DeepEqual(gotStatementIDs, wantStatementIDs) {
		t.Fatalf("statement IDs = %v, want %v", gotStatementIDs, wantStatementIDs)
	}
	for _, item := range manifest.Builtins {
		if item.Inputs == nil || item.Contexts == nil || item.Constraints == nil {
			t.Errorf("builtin %s has omitted arrays", item.ID)
		}
		assertSorted(t, item.ID+" inputs", item.Inputs)
		assertSorted(t, item.ID+" contexts", item.Contexts)
		assertSorted(t, item.ID+" constraints", item.Constraints)
	}
	if !strings.Contains(strings.Join(builtinByID(t, manifest.Builtins, "byte_at").Constraints, " "), "outside") || !strings.Contains(strings.Join(builtinByID(t, manifest.Builtins, "write_text").Constraints, " "), "quoted") {
		t.Fatal("byte_at bounds or write_text literal context is not described")
	}

	wantOperators := map[string]string{"u64.add": "+", "u64.equal": "==", "u64.less_than": "<", "u64.remainder": "%"}
	for _, item := range manifest.Operators {
		if want, ok := wantOperators[item.ID]; !ok || item.Syntax != want {
			t.Errorf("unexpected operator %s = %q", item.ID, item.Syntax)
		}
		delete(wantOperators, item.ID)
	}
	if len(wantOperators) != 0 {
		t.Errorf("missing operators: %v", wantOperators)
	}
	if !strings.Contains(strings.Join(entryByID(t, manifest.Operators, "u64.add").Constraints, " "), "overflow") || !strings.Contains(strings.Join(entryByID(t, manifest.Operators, "u64.remainder").Constraints, " "), "zero") {
		t.Fatal("addition overflow or remainder-zero boundary is not stated")
	}

	wantBuiltins := map[string]struct {
		arity  int
		inputs []string
		result string
	}{
		"byte_at":       {2, []string{"ByteView", "U64"}, "U64"},
		"byte_len":      {1, []string{"ByteView"}, "U64"},
		"write_newline": {0, []string{}, "Unit"},
		"write_text":    {1, []string{"quoted text literal"}, "Unit"},
		"write_u64":     {1, []string{"U64"}, "Unit"},
	}
	for _, item := range manifest.Builtins {
		want, ok := wantBuiltins[item.ID]
		if !ok || item.Arity != want.arity || item.Result != want.result || !reflect.DeepEqual(item.Inputs, want.inputs) {
			t.Errorf("builtin %s = %+v, want %+v", item.ID, item, want)
		}
		delete(wantBuiltins, item.ID)
	}
	if len(wantBuiltins) != 0 {
		t.Errorf("missing builtins: %v", wantBuiltins)
	}
	if manifest.Calls.DeclarationArity != 1 || !reflect.DeepEqual(manifest.Calls.OrdinaryArity, []int{1}) || manifest.Calls.ScalarExpressionCalls {
		t.Fatalf("call contract = %+v", manifest.Calls)
	}
	if manifest.Calls.OrdinaryContexts == nil || manifest.Calls.Constraints == nil {
		t.Fatal("call contract omitted required arrays")
	}
	if !reflect.DeepEqual(manifest.Calls.OrdinaryContexts, []string{"admitted linear binding and match-arm call paths"}) {
		t.Fatalf("ordinary call contexts = %v", manifest.Calls.OrdinaryContexts)
	}
	assertSorted(t, "ordinary call contexts", manifest.Calls.OrdinaryContexts)
	assertSorted(t, "call constraints", manifest.Calls.Constraints)
	if !strings.Contains(strings.Join(manifest.Calls.Constraints, " "), "acyclic") || !strings.Contains(strings.Join(manifest.Calls.Constraints, " "), "Foreign") {
		t.Fatalf("ordinary/foreign call restrictions are incomplete: %v", manifest.Calls.Constraints)
	}

	wantRefusals := []Refusal{
		{"surface.unsupported_bool_literal_tail", "unbound true or false in a result tail", "Bool literals are unsupported in result tails; return an initialized value allowed by the declared result type."},
		{"surface.unsupported_if_without_else", "if statement without else block", "An if statement requires an else block."},
		{"surface.unsupported_logical_and", "&&", "Logical && is unsupported; use explicit if/else control flow."},
		{"surface.unsupported_logical_not", "!", "Logical ! is unsupported; use explicit if/else control flow."},
		{"surface.unsupported_logical_or", "||", "Logical || is unsupported; use explicit if/else control flow."},
		{"surface.unsupported_parameter_arity", "more than one declaration parameter", "Declarations require exactly one parameter; pass one admitted value."},
		{"surface.unsupported_subtraction", "-", "Subtraction is unsupported; no unchecked emulation is recommended."},
	}
	if !reflect.DeepEqual(manifest.Refusals, wantRefusals) {
		t.Fatalf("refusal catalog = %+v, want independently pinned literals %+v", manifest.Refusals, wantRefusals)
	}
	fileBytes := entryByID(t, manifest.Restrictions, "filebytes_limit")
	fileBytesText := strings.Join(fileBytes.Constraints, " ")
	if !strings.Contains(fileBytesText, "validated binding selects the maximum") || !strings.Contains(fileBytesText, "does not publish a universal 65536-byte") {
		t.Fatalf("FileBytes limit claim is not configuration-scoped: %q", fileBytesText)
	}
	for _, entry := range append(append([]Entry{}, manifest.Operators...), manifest.Statements...) {
		lower := strings.ToLower(entry.ID + " " + entry.Syntax)
		if strings.Contains(lower, "bitwise") || strings.Contains(lower, "hex") || strings.Contains(lower, "shift") {
			t.Errorf("future-owned operator/statement leaked into current surface: %+v", entry)
		}
	}

	data, err := JSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(string(data), "\n") || strings.HasSuffix(string(data), "\n\n") {
		t.Fatalf("JSON must have exactly one trailing newline")
	}
	keys := []string{`"schema"`, `"coordinates"`, `"types"`, `"operators"`, `"statements"`, `"builtins"`, `"calls"`, `"restrictions"`, `"limits"`, `"refusals"`}
	last := -1
	for _, key := range keys {
		at := strings.Index(string(data), key)
		if at <= last {
			t.Fatalf("field %s is missing or out of fixed field order", key)
		}
		last = at
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal(data, &wire); err != nil || len(wire) != len(keys) {
		t.Fatalf("wire document has an envelope or invalid JSON: fields=%d error=%v", len(wire), err)
	}
	again, err := JSON(Current())
	if err != nil || string(data) != string(again) {
		t.Fatalf("Current JSON is not deterministic: error=%v", err)
	}
	assertCanonicalCopies(t)
}

func TestPhase34SurfaceLimits(t *testing.T) {
	manifest := Current()
	want := map[string]struct {
		category string
		scope    string
		unit     string
		value    int64
		routes   []string
	}{
		"application.argument_bytes": {"application input", "one application input argument", "bytes", 4096, []string{"app run"}},
		"application.evidence_bytes": {"application capture", "one same-run application evidence capture/report", "bytes", 65536, []string{"app run"}},
		"application.output_bytes":   {"application output", "one retained application stdout stream", "bytes", 65536, []string{"app run"}},
		"check.work":                 {"deterministic checker work", "one checked source program including its loaded modules", "work units", 100000, []string{"app verify", "build", "check", "run"}},
		"evidence.manifest_bytes":    {"tool input", "one evidence manifest file", "bytes", 1048576, []string{"evidence --validate"}},
		"evidence.tool_stream_bytes": {"external tool capture", "each captured external-tool stdout or stderr stream", "bytes per stream", 65536, []string{"evidence", "evidence --validate"}},
		"execution.document_bytes":   {"execution capture", "one native execution document", "bytes", 16777216, []string{"app verify"}},
		"interface.artifact_bytes":   {"interface input", "each interface summary/core JSON input", "bytes per file", 2097152, []string{"interface check"}},
		"module.depth":               {"module loading", "import chain below the entry module", "modules", 8, []string{"app verify", "build", "check", "debug-map", "run"}},
		"module.count":               {"module loading", "all distinct modules including the entry", "modules", 16, []string{"app verify", "build", "check", "debug-map", "run"}},
		"module.imports_per_file":    {"source parser", "one module source file", "imports", 16, []string{"app verify", "build", "check", "debug-map", "run"}},
		"module.total_source_bytes":  {"module loading", "all source files in one loaded module program", "bytes", 4194304, []string{"app verify", "build", "check", "debug-map", "run"}},
		"parser.alternatives":        {"authoring parse bound", "one parsed source program", "alternatives", 4096, nil},
		"parser.declarations":        {"authoring parse bound", "one source file", "declarations", 4096, nil},
		"parser.functions":           {"authoring parse bound", "one source file", "functions", 1024, nil},
		"parser.linear_bindings":     {"authoring parse bound", "one parsed linear body", "bindings", 65536, nil},
		"parser.primary_diagnostics": {"authoring diagnostic bound", "one parsed source", "diagnostics", 20, nil},
		"parser.source_bytes":        {"authoring parse bound", "one source file", "bytes", 1048576, nil},
		"parser.tokens":              {"authoring parse bound", "one source file", "tokens", 131072, nil},
		"parser.type_depth":          {"authoring type bound", "one nested type expression", "levels", 64, nil},
		"parser.type_nodes":          {"authoring type bound", "one source file", "nodes", 4096, nil},
		"scalar.analysis_work":       {"deterministic scalar analysis", "one scalar CFG function", "work units", 65536, []string{"app verify", "build", "check", "run"}},
	}
	if len(manifest.Limits) != len(want) {
		t.Fatalf("limit count = %d, want %d", len(manifest.Limits), len(want))
	}
	for _, item := range manifest.Limits {
		expected, ok := want[item.ID]
		if !ok {
			t.Errorf("unexpected limit %q", item.ID)
			continue
		}
		if item.Category != expected.category || item.Scope != expected.scope || item.Unit != expected.unit || item.Value != expected.value {
			t.Errorf("limit %s = category %q scope %q unit %q value %d; want %q %q %q %d", item.ID, item.Category, item.Scope, item.Unit, item.Value, expected.category, expected.scope, expected.unit, expected.value)
		}
		if item.Meaning == "" || item.Routes == nil || len(item.Routes) == 0 {
			t.Errorf("limit %s omits meaning or route scope: %+v", item.ID, item)
		}
		if expected.routes != nil && !reflect.DeepEqual(item.Routes, expected.routes) {
			t.Errorf("limit %s routes = %v, want %v", item.ID, item.Routes, expected.routes)
		}
		if expected.routes == nil && !reflect.DeepEqual(item.Routes, []string{"app verify", "build", "check", "debug-map", "evidence", "evidence --validate", "explain", "format", "interface export", "query", "run", "verify"}) {
			t.Errorf("limit %s parser routes = %v", item.ID, item.Routes)
		}
		delete(want, item.ID)
	}
	if len(want) != 0 {
		t.Errorf("missing limits: %v", want)
	}
	for _, item := range manifest.Limits {
		assertSorted(t, "limit "+item.ID+" routes", item.Routes)
	}
}

func assertCanonicalCopies(t *testing.T) {
	t.Helper()
	manifest := Manifest{
		Schema: "test",
		Types: []TypeEntry{
			{Entry: Entry{ID: "z", Contexts: []string{"z", "a"}, Constraints: []string{"z", "a"}}},
			{Entry: Entry{ID: "a", Contexts: []string{"y", "b"}, Constraints: []string{"y", "b"}}},
		},
		Operators:    []Entry{{ID: "z", Contexts: []string{"z", "a"}}, {ID: "a"}},
		Statements:   []Entry{{ID: "z"}},
		Builtins:     []BuiltinEntry{{ID: "z", Inputs: []string{"z", "a"}, Contexts: []string{"z", "a"}}, {ID: "a"}},
		Restrictions: []Entry{{ID: "z"}},
		Limits:       []Limit{{ID: "z", Routes: []string{"z", "a"}}, {ID: "a"}},
		Refusals:     []Refusal{{Code: "z"}, {Code: "a"}},
		Calls:        CallRules{OrdinaryArity: []int{2, 1}, OrdinaryContexts: []string{"z", "a"}},
	}
	data, err := JSON(manifest)
	if err != nil {
		t.Fatal(err)
	}
	var got Manifest
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if got.Types[0].ID != "a" || !reflect.DeepEqual(got.Types[0].Contexts, []string{"b", "y"}) || !reflect.DeepEqual(got.Calls.OrdinaryArity, []int{1, 2}) || got.Limits[0].ID != "a" || got.Refusals[0].Code != "a" {
		t.Fatalf("JSON did not sort copied arrays: %+v", got)
	}
	if manifest.Types[0].ID != "z" || manifest.Types[0].Contexts[0] != "z" || manifest.Calls.OrdinaryArity[0] != 2 {
		t.Fatalf("JSON mutated caller-owned slices: %+v", manifest)
	}
	empty, err := JSON(Manifest{})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{`"types": []`, `"operators": []`, `"statements": []`, `"builtins": []`, `"ordinary_arity": []`, `"ordinary_contexts": []`, `"constraints": []`, `"restrictions": []`, `"limits": []`, `"refusals": []`} {
		if !strings.Contains(string(empty), expected) {
			t.Errorf("empty manifest omitted required array %s", expected)
		}
	}
}

func typeByID(t *testing.T, entries []TypeEntry, id string) TypeEntry {
	t.Helper()
	for _, item := range entries {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("missing type %q", id)
	return TypeEntry{}
}

func entryByID(t *testing.T, entries []Entry, id string) Entry {
	t.Helper()
	for _, item := range entries {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("missing entry %q", id)
	return Entry{}
}

func builtinByID(t *testing.T, entries []BuiltinEntry, id string) BuiltinEntry {
	t.Helper()
	for _, item := range entries {
		if item.ID == id {
			return item
		}
	}
	t.Fatalf("missing builtin %q", id)
	return BuiltinEntry{}
}

func assertEntryOrder(t *testing.T, entries []Entry) {
	t.Helper()
	for i := 1; i < len(entries); i++ {
		if entries[i-1].ID > entries[i].ID {
			t.Errorf("entries are not ID-sorted: %s before %s", entries[i-1].ID, entries[i].ID)
		}
	}
}

func assertSorted(t *testing.T, label string, values []string) {
	t.Helper()
	if !sort.StringsAreSorted(values) {
		t.Errorf("%s is not lexically sorted: %v", label, values)
	}
}
