package reduce_test

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/reduce"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// goldenMismatchProgram is a small, hand-built core.Program -- the same
// discipline reduce_test.go's own fixture builders use -- standing in for
// a reduced core artifact.
func goldenMismatchProgram() core.Program {
	return core.Program{
		Schema:   "lang.core/1",
		Module:   "phase5.mismatch_golden",
		ModuleID: "phase5.mismatch_golden",
		Functions: []core.Function{
			{
				ID:         "fn:touch",
				Name:       "touch",
				Parameter:  core.Parameter{ID: "place:0", Name: "value", Type: "Byte"},
				ReturnType: "Byte",
				Linear: &core.LinearBody{
					ID:     "fn:touch:linear",
					Types:  []core.TypeFact{{ID: "type:0"}},
					Places: []core.Place{{ID: "place:0", Name: "value", TypeID: "type:0"}},
					Operations: []core.LinearOperation{
						{ID: "fn:touch:op:0", PointID: "fn:touch:point:linear:0", Kind: core.OpReturn, SourceID: "place:0", TypeID: "type:0"},
					},
				},
			},
		},
	}
}

func goldenMismatchDocument(t testing.TB) reduce.MismatchDocument {
	t.Helper()
	document, err := reduce.NewMismatchDocument(
		"axis:terminal-outcome",
		"O0-vs-O3",
		"fn:touch:op:0",
		goldenMismatchProgram(),
		"module phase5.mismatch_golden\n\nexport {\n  fn touch\n}\n\nfn touch(value: Byte) -> Byte {\n  value\n}\n",
		reduce.MinimalityFixpoint,
		1, 1,
		map[string][]reduce.EventRecord{
			"O0": {{Kind: "value.returned", OperationID: "fn:touch:op:0", FunctionID: "fn:touch", Detail: "2"}},
			"O3": {{Kind: "value.returned", OperationID: "fn:touch:op:0", FunctionID: "fn:touch", Detail: "7"}},
		},
		[]reduce.CausalStep{{Kind: "borrow_shared", OperationID: "fn:touch:op:0", Place: "value"}},
		[]diagnostic.Cause{{Kind: "reduce.diverging_alias_claim", Detail: "false restrict claim hoisted a stale read at -O3"}},
		"evidence:golden-mismatch-fixture",
	)
	if err != nil {
		t.Fatalf("goldenMismatchDocument: %v", err)
	}
	return document
}

func TestMismatchDocumentRoundTrips(t *testing.T) {
	document := goldenMismatchDocument(t)
	encoded, err := json.Marshal(document)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := readMismatchGolden(t, "mismatch.golden.json")
	if !bytes.Equal(bytes.TrimSpace(encoded), bytes.TrimSpace(want)) {
		t.Fatalf("mismatch document golden changed:\ngot  %s\nwant %s", encoded, want)
	}
	var decoded reduce.MismatchDocument
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !reflect.DeepEqual(decoded, document) {
		t.Fatalf("round trip lost data:\ngot  %+v\nwant %+v", decoded, document)
	}

	// Stability: marshaling the same document twice must be byte-identical.
	second, err := json.Marshal(goldenMismatchDocument(t))
	if err != nil {
		t.Fatalf("marshal (second): %v", err)
	}
	if !bytes.Equal(encoded, second) {
		t.Fatalf("mismatch document is not deterministic across identical builds")
	}
}

// TestMismatchDocumentHasExactlyTheSpecifiedFields is D-05-26's own
// reflection-based falsifier: the struct's JSON tag set must equal exactly
// the specified field names, so a fourteenth field cannot be added
// silently. The set here is D-05-26's literal twelve PLUS the additive
// `evidence_id` field -- a recorded discretionary deviation from D-05-26's
// literal enumeration, per 05-12-SUMMARY.md's "Deviations from Plan"
// section (licensed by 05-CONTEXT.md's "Claude's Discretion" block).
func TestMismatchDocumentHasExactlyTheSpecifiedFields(t *testing.T) {
	want := map[string]bool{
		"schema": true, "diverging_axis": true, "engine_pair": true,
		"diverging_operation_id": true, "reduced_core": true, "reduced_source": true,
		"minimality": true, "total_recomputed_work": true, "reduction_attempts": true,
		"event_window": true, "causal_chain": true, "causes": true,
		"evidence_id": true,
	}
	typ := reflect.TypeOf(reduce.MismatchDocument{})
	got := map[string]bool{}
	for i := 0; i < typ.NumField(); i++ {
		tag := typ.Field(i).Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == "" || name == "-" {
			t.Fatalf("field %s has no usable json tag: %q", typ.Field(i).Name, tag)
		}
		got[name] = true
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MismatchDocument field set changed:\ngot  %+v\nwant %+v", got, want)
	}
}

func TestMismatchEventWindowIsBounded(t *testing.T) {
	events := make([]reduce.EventRecord, 50)
	for i := range events {
		events[i] = reduce.EventRecord{Kind: "value.returned", OperationID: strconv.Itoa(i)}
	}
	document, err := reduce.NewMismatchDocument(
		"axis:event-order", "interpreter-vs-O0", "op:49",
		goldenMismatchProgram(), "n/a", reduce.MinimalityFixpoint, 1, 1,
		map[string][]reduce.EventRecord{"interpreter": events},
		[]reduce.CausalStep{{Kind: "borrow_shared", OperationID: "op:0"}},
		nil, "",
	)
	if err != nil {
		t.Fatalf("NewMismatchDocument: %v", err)
	}
	got := document.EventWindow["interpreter"]
	if len(got) != reduce.EventWindowSize {
		t.Fatalf("expected exactly %d events, got %d", reduce.EventWindowSize, len(got))
	}
	wantLast := events[len(events)-reduce.EventWindowSize:]
	for i := range got {
		if got[i] != wantLast[i] {
			t.Fatalf("event window is not the LAST %d events: index %d got %+v want %+v", reduce.EventWindowSize, i, got[i], wantLast[i])
		}
	}
}

func TestMismatchCausesReusesDiagnosticShape(t *testing.T) {
	document := goldenMismatchDocument(t)
	var _ []diagnostic.Cause = document.Causes
	fieldType := reflect.TypeOf(reduce.MismatchDocument{}).Field(fieldIndex(t, "Causes")).Type
	wantType := reflect.TypeOf([]diagnostic.Cause(nil))
	if fieldType != wantType {
		t.Fatalf("Causes field type = %v, want %v (the existing lang.diagnostic/1 cause-graph shape, verbatim)", fieldType, wantType)
	}
}

func fieldIndex(t testing.TB, name string) int {
	t.Helper()
	typ := reflect.TypeOf(reduce.MismatchDocument{})
	for i := 0; i < typ.NumField(); i++ {
		if typ.Field(i).Name == name {
			return i
		}
	}
	t.Fatalf("no field named %s", name)
	return -1
}

// TestNoNewSchemaVersionsIntroduced is D-05-39's own source-scan falsifier:
// the only new schema this phase introduces is lang.mismatch/0. Any
// occurrence of the QUOTED Go string literal "lang.core/2", "lang.evidence/2",
// or "lang.diagnostic/2" anywhere under internal/ is a violation -- matching
// only the quoted literal (not bare prose) is deliberate: check.go and this
// very plan's own doc comments discuss these identifiers BY NAME, in prose,
// to explain that no such schema bump occurs (D-05-39), and a bare-substring
// scan would false-positive on that explanatory prose forever. A real
// schema-version introduction is always a quoted Go string constant, never
// unquoted prose.
func TestNoNewSchemaVersionsIntroduced(t *testing.T) {
	forbidden := []string{`"lang.core/2"`, `"lang.evidence/2"`, `"lang.diagnostic/2"`}
	root := testsupport.ProjectPath("internal")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		if strings.HasSuffix(path, "mismatch_test.go") {
			// This file's own forbidden-literal list would otherwise match
			// itself -- it defines the literals it scans for, it does not
			// introduce them.
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		text := string(content)
		for _, forbiddenSchema := range forbidden {
			if strings.Contains(text, forbiddenSchema) {
				t.Errorf("forbidden schema identifier %q found in %s", forbiddenSchema, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
}

// TestMismatchDocumentRejectsInvalidMinimality is the constructor-level
// falsifier the plan's own acceptance criteria name: a Minimality value
// outside the two closed values Reduce ever produces (D-05-25) must be
// refused, never silently accepted.
func TestMismatchDocumentRejectsInvalidMinimality(t *testing.T) {
	_, err := reduce.NewMismatchDocument(
		"axis:terminal-outcome", "O0-vs-O3", "fn:touch:op:0",
		goldenMismatchProgram(), "n/a", "not-a-real-minimality-value", 1, 1, nil, nil, nil, "",
	)
	if err == nil {
		t.Fatal("expected an error for an invalid Minimality value, got nil")
	}
}

func readMismatchGolden(t testing.TB, name string) []byte {
	t.Helper()
	value, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "reduce", "testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return value
}
