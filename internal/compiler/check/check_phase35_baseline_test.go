package check_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase35NumericBaselineSchema = "phase35.numeric-policy-baseline/1"

type phase35NumericBaseline struct {
	Schema   string               `json:"schema"`
	Revision string               `json:"revision"`
	Numeric  []phase35NumericCase `json:"numeric_cases"`
	Bool     []phase35BoolCase    `json:"bool_identity_cases"`
}

type phase35Observation struct {
	Outcome     string                 `json:"outcome"`
	Diagnostics []phase35DiagnosticRow `json:"diagnostics,omitempty"`
}

type phase35DiagnosticRow struct {
	ID      string          `json:"id"`
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Primary diagnostic.Span `json:"primary_span"`
}

type phase35NumericCase struct {
	Name        string             `json:"name"`
	Route       string             `json:"route"`
	Literal     string             `json:"literal"`
	Source      string             `json:"source"`
	SourceSHA   string             `json:"source_sha256"`
	Parse       phase35Observation `json:"parse"`
	Check       phase35Observation `json:"check"`
	CoreSchema  string             `json:"core_schema,omitempty"`
	CoreSHA     string             `json:"core_sha256,omitempty"`
	SemanticIDs []string           `json:"semantic_ids,omitempty"`
	ConstU64    []string           `json:"const_u64,omitempty"`
}

type phase35BoolCase struct {
	Name        string             `json:"name"`
	Source      string             `json:"source"`
	SourceSHA   string             `json:"source_sha256"`
	Check       phase35Observation `json:"check"`
	CoreSchema  string             `json:"core_schema,omitempty"`
	CoreSHA     string             `json:"core_sha256,omitempty"`
	SemanticIDs []string           `json:"semantic_ids,omitempty"`
}

func TestPhase35NumericPolicyBaseline(t *testing.T) {
	fixture := phase35LoadNumericBaseline(t)
	if err := exec.Command("git", "cat-file", "-e", fixture.Revision+"^{commit}").Run(); err != nil {
		t.Fatalf("frozen numeric baseline revision %s is unavailable: %v", fixture.Revision, err)
	}
	phase35ValidateNumericHistory(t, fixture.Numeric)
	observed := phase35ObserveNumericAndBool(t)
	if !phase35EqualJSON(t, phase35StableNumeric(fixture.Numeric), phase35StableNumeric(observed.Numeric)) {
		t.Fatal("stable route-local 010/0B10 observations differ from the live checker")
	}
	if !phase35EqualJSON(t, fixture.Bool, observed.Bool) {
		t.Fatal("Bool identity observations differ from the live checker")
	}
	for _, route := range []string{"legacy", "scalar"} {
		for _, literal := range []string{"010", "0B10"} {
			row := phase35FindNumeric(fixture.Numeric, route, literal)
			if row == nil {
				t.Fatalf("frozen fixture lacks %s %s witness", route, literal)
			}
		}
	}
}

func phase35ValidateNumericHistory(t *testing.T, rows []phase35NumericCase) {
	t.Helper()
	if len(rows) != 12 {
		t.Fatalf("frozen numeric case count=%d, want 12 route/literal pairs", len(rows))
	}
	for _, route := range []string{"legacy", "scalar"} {
		for _, literal := range []string{"010", "0B10", "0b1_0", "0xA_B", "0X1A", "1_000"} {
			row := phase35FindNumeric(rows, route, literal)
			if row == nil {
				t.Fatalf("frozen fixture lacks %s %s witness", route, literal)
			}
			if phase35SHA([]byte(row.Source)) != row.SourceSHA {
				t.Fatalf("frozen %s %s source digest does not match its bytes", route, literal)
			}
			if !strings.Contains(row.Source, "module phase35.numeric."+route+"\n") || !strings.Contains(row.Source, "let value = "+literal+"\n") {
				t.Fatalf("frozen %s %s source does not identify its route and literal", route, literal)
			}
			if row.Parse.Outcome == "" || row.Check.Outcome == "" {
				t.Fatalf("frozen %s %s lacks explicit parse/check outcomes", route, literal)
			}
			for _, problem := range row.Parse.Diagnostics {
				if problem.Primary.Start < 0 || problem.Primary.End < problem.Primary.Start || problem.Primary.End > len(row.Source) {
					t.Fatalf("frozen %s %s parse diagnostic %s span is outside source: %+v", route, literal, problem.Code, problem.Primary)
				}
			}
			for _, problem := range row.Check.Diagnostics {
				if problem.Primary.Start < 0 || problem.Primary.End < problem.Primary.Start || problem.Primary.End > len(row.Source) {
					t.Fatalf("frozen %s %s check diagnostic %s span is outside source: %+v", route, literal, problem.Code, problem.Primary)
				}
			}
			if row.Check.Outcome == "admitted" && (row.CoreSchema == "" || row.CoreSHA == "" || len(row.SemanticIDs) == 0) {
				t.Fatalf("frozen %s %s admitted row lacks core/schema/identity evidence", route, literal)
			}
			if row.Check.Outcome == "refused" {
				if len(row.Check.Diagnostics) != 1 || row.Check.Diagnostics[0].ID == "" || row.Check.Diagnostics[0].Code != "check.literal_out_of_range" {
					t.Fatalf("frozen %s %s refusal lost its exact range diagnostic: %+v", route, literal, row.Check.Diagnostics)
				}
				span := row.Check.Diagnostics[0].Primary
				if row.Source[span.Start:span.End] != literal {
					t.Fatalf("frozen %s %s diagnostic span covers %q", route, literal, row.Source[span.Start:span.End])
				}
			}
		}
	}
	for _, expected := range []struct {
		route, literal, outcome, value string
	}{
		{"legacy", "010", "admitted", "10"},
		{"legacy", "0B10", "refused", ""},
		{"legacy", "0b1_0", "admitted", "2"},
		{"legacy", "0xA_B", "admitted", "171"},
		{"legacy", "0X1A", "refused", ""},
		{"legacy", "1_000", "admitted", "1000"},
		{"scalar", "010", "admitted", "8"},
		{"scalar", "0B10", "admitted", "2"},
		{"scalar", "0b1_0", "admitted", "2"},
		{"scalar", "0xA_B", "admitted", "171"},
		{"scalar", "0X1A", "admitted", "26"},
		{"scalar", "1_000", "admitted", "1000"},
	} {
		row := phase35FindNumeric(rows, expected.route, expected.literal)
		if row.Check.Outcome != expected.outcome {
			t.Fatalf("frozen %s %s outcome=%s, want %s", expected.route, expected.literal, row.Check.Outcome, expected.outcome)
		}
		if expected.value != "" && (len(row.ConstU64) != 1 || row.ConstU64[0] != expected.value) {
			t.Fatalf("frozen %s %s constants=%v, want %s", expected.route, expected.literal, row.ConstU64, expected.value)
		}
	}
}

func phase35StableNumeric(rows []phase35NumericCase) []phase35NumericCase {
	var stable []phase35NumericCase
	for _, row := range rows {
		if row.Literal == "010" || row.Literal == "0B10" {
			stable = append(stable, row)
		}
	}
	return stable
}

func phase35LoadNumericBaseline(t *testing.T) phase35NumericBaseline {
	t.Helper()
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-policy-baseline.json"))
	if err != nil {
		t.Fatalf("read numeric baseline: %v", err)
	}
	var fixture phase35NumericBaseline
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode numeric baseline: %v", err)
	}
	if fixture.Schema != phase35NumericBaselineSchema || fixture.Revision == "" {
		t.Fatalf("invalid numeric baseline header: schema=%q revision=%q", fixture.Schema, fixture.Revision)
	}
	return fixture
}

func phase35ObserveNumericAndBool(t *testing.T) phase35NumericBaseline {
	t.Helper()
	result := phase35NumericBaseline{Schema: phase35NumericBaselineSchema}
	for _, route := range []string{"legacy", "scalar"} {
		parameterType := "Byte"
		if route == "scalar" {
			parameterType = "U64"
		}
		for _, literal := range []string{"010", "0B10", "0b1_0", "0xA_B", "0X1A", "1_000"} {
			name := route + "_" + strings.NewReplacer("_", "separator", "0", "zero", "x", "x", "B", "B").Replace(literal)
			source := []byte(fmt.Sprintf("module phase35.numeric.%s\nexport {\n  fn main\n}\nfn main(input: %s) -> U64 {\n  let value = %s\n  value\n}\n", route, parameterType, literal))
			row := phase35NumericCase{Name: name, Route: route, Literal: literal, Source: string(source), SourceSHA: phase35SHA(source)}
			parsed := syntax.Parse(source)
			row.Parse = phase35Observation{Outcome: phase35Outcome(len(parsed.Diagnostics) == 0), Diagnostics: phase35DiagnosticRows(parsed.Diagnostics)}
			checked := session.Check(source)
			row.Check = phase35Observation{Outcome: phase35Outcome(len(checked.Diagnostics) == 0), Diagnostics: phase35DiagnosticRows(checked.Diagnostics)}
			if len(checked.Diagnostics) == 0 {
				row.CoreSchema, row.CoreSHA, row.SemanticIDs = phase35CoreWitness(t, checked.Program)
				row.ConstU64 = phase35Constants(checked.Program)
			}
			result.Numeric = append(result.Numeric, row)
		}
	}
	for _, item := range []struct {
		name   string
		source string
	}{
		{"parameter_true", "module phase35.bool.parameter_true\nexport { fn main }\nfn main(true: U64) -> U64 { true }\n"},
		{"parameter_false", "module phase35.bool.parameter_false\nexport { fn main }\nfn main(false: U64) -> U64 { false }\n"},
		{"local_true", "module phase35.bool.local_true\nexport { fn main }\nfn main(n: U64) -> U64 { let true = n true }\n"},
		{"local_false", "module phase35.bool.local_false\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n false }\n"},
	} {
		source := []byte(item.source)
		checked := session.Check(source)
		row := phase35BoolCase{Name: item.name, Source: item.source, SourceSHA: phase35SHA(source), Check: phase35Observation{Outcome: phase35Outcome(len(checked.Diagnostics) == 0), Diagnostics: phase35DiagnosticRows(checked.Diagnostics)}}
		if len(checked.Diagnostics) == 0 {
			row.CoreSchema, row.CoreSHA, row.SemanticIDs = phase35CoreWitness(t, checked.Program)
		}
		result.Bool = append(result.Bool, row)
	}
	return result
}

func phase35FindNumeric(rows []phase35NumericCase, route, literal string) *phase35NumericCase {
	for index := range rows {
		if rows[index].Route == route && rows[index].Literal == literal {
			return &rows[index]
		}
	}
	return nil
}

func phase35CoreWitness(t *testing.T, program core.Program) (string, string, []string) {
	t.Helper()
	data, err := json.Marshal(program)
	if err != nil {
		t.Fatalf("serialize checked core: %v", err)
	}
	ids := []string{program.ModuleID}
	for _, function := range program.Functions {
		ids = append(ids, function.ID, function.EntryPointID, function.ReturnPointID)
		if function.Linear != nil {
			for _, operation := range function.Linear.Operations {
				ids = append(ids, operation.ID)
			}
		}
	}
	sort.Strings(ids)
	return program.Schema, phase35SHA(data), ids
}

func phase35Constants(program core.Program) []string {
	var values []string
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpConst {
				values = append(values, operation.ConstU64)
			}
		}
	}
	return values
}

func phase35DiagnosticRows(items []diagnostic.Diagnostic) []phase35DiagnosticRow {
	rows := make([]phase35DiagnosticRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, phase35DiagnosticRow{ID: item.ID, Code: item.Code, Message: item.Message, Primary: item.Primary})
	}
	return rows
}

func phase35Outcome(admitted bool) string {
	if admitted {
		return "admitted"
	}
	return "refused"
}

func phase35SHA(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func phase35EqualJSON(t *testing.T, want, got any) bool {
	t.Helper()
	wantJSON, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("marshal expected observations: %v", err)
	}
	gotJSON, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal current observations: %v", err)
	}
	return string(wantJSON) == string(gotJSON)
}
