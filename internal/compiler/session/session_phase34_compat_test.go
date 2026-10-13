package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase34IdentitySchema       = "phase34.identity-baseline/1"
	phase34IdentityBaselineHash = "8cdcc4243cecf24b3721a652b6835aae077e62f797826e484fc7a0618273174c"
)

type phase34CompatibilityBaseline struct {
	Schema        string                       `json:"schema"`
	Revision      string                       `json:"revision"`
	BoolCases     []phase34BoolObservation     `json:"bool_cases"`
	IdentityCases []phase34IdentityObservation `json:"identity_cases"`
	ResultIDs     map[string]string            `json:"result_ids"`
}

type phase34BoolObservation struct {
	ID           string             `json:"id"`
	Source       string             `json:"source"`
	ModuleSource string             `json:"module_source,omitempty"`
	Check        phase34RouteRecord `json:"check"`
	Command      phase34RouteRecord `json:"check_command_file"`
}

type phase34RouteRecord struct {
	Outcome     string                `json:"outcome"`
	Status      string                `json:"status"`
	Error       string                `json:"error,omitempty"`
	Diagnostics []phase34RouteProblem `json:"diagnostics"`
}

type phase34RouteProblem struct {
	ID      string          `json:"id"`
	Code    string          `json:"code"`
	Primary diagnostic.Span `json:"primary_span"`
	JSON    string          `json:"json"`
}

type phase34IdentityObservation struct {
	Name          string                  `json:"name"`
	Schema        string                  `json:"schema"`
	DiagnosticID  string                  `json:"diagnostic_id"`
	Unit          string                  `json:"unit,omitempty"`
	CausesNil     bool                    `json:"causes_nil"`
	MarshaledJSON string                  `json:"marshaled_json"`
	Repairs       []phase34RepairSnapshot `json:"repairs"`
}

type phase34RepairSnapshot struct {
	Kind           string           `json:"kind"`
	Span           *diagnostic.Span `json:"span,omitempty"`
	Replacement    string           `json:"replacement"`
	Applicability  string           `json:"applicability"`
	DriverEligible bool             `json:"driver_eligible"`
}

type phase34BoolCase struct {
	ID           string
	Source       string
	ModuleSource string
}

func phase34BoolCases() []phase34BoolCase {
	return []phase34BoolCase{
		{ID: "linear-true-parameter", Source: "module draft.bool_linear_true\nexport { fn main }\nfn main(true: U64) -> U64 { true }\n"},
		{ID: "scalar-true-parameter", Source: "module draft.bool_scalar_true\nexport { fn main }\nfn main(true: U64) -> U64 { let result = true + 1 result }\n"},
		{ID: "linear-false-binding", Source: "module draft.bool_false_binding\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n false }\n"},
		{ID: "false-parameter-true-binding", Source: "module draft.bool_mixed_names\nexport { fn main }\nfn main(false: U64) -> U64 { let true = false true }\n"},
		{ID: "bound-true-tail", Source: "module draft.bool_bound_true\nexport { fn main }\nfn main(n: U64) -> U64 { let true = n true }\n"},
		{ID: "bound-false-tail", Source: "module draft.bool_bound_false\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n false }\n"},
		{ID: "scalar-false-binding", Source: "module draft.bool_scalar_false\nexport { fn main }\nfn main(n: U64) -> U64 { let false = n + 1 false }\n"},
		{ID: "imported-true-function", Source: "module draft.bool_importer\nimport draft.bool_names { fn true }\nexport { fn main }\nfn main(n: U64) -> U64 { let result = true(n) result }\n", ModuleSource: "module draft.bool_names\nexport { fn true }\nfn true(value: U64) -> U64 { value }\n"},
		{ID: "unbound-true-tail", Source: "module draft.bool_unbound\nexport { fn main }\nfn main(n: U64) -> U64 { true }\n"},
	}
}

func phase34IdentityBaselinePath() string {
	return testsupport.ProjectPath("testdata", "phase34", "identity-baseline.json")
}

func phase34ObserveBoolCases(t *testing.T, includeUnbound bool) []phase34BoolObservation {
	t.Helper()
	var observations []phase34BoolObservation
	for _, testCase := range phase34BoolCases() {
		if !includeUnbound && testCase.ID == "unbound-true-tail" {
			continue
		}
		directory := t.TempDir()
		entryPath := filepath.Join(directory, "main.schway")
		if err := os.WriteFile(entryPath, []byte(testCase.Source), 0o644); err != nil {
			t.Fatal(err)
		}
		if testCase.ModuleSource != "" {
			modulePath := filepath.Join(directory, "draft", "bool_names.schway")
			if err := os.MkdirAll(filepath.Dir(modulePath), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(modulePath, []byte(testCase.ModuleSource), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		check := session.Check([]byte(testCase.Source))
		command, commandErr := session.CheckCommandFile(entryPath)
		observations = append(observations, phase34BoolObservation{
			ID: testCase.ID, Source: testCase.Source, ModuleSource: testCase.ModuleSource,
			Check: phase34CheckRoute(check.Diagnostics), Command: phase34CommandRoute(command, commandErr),
		})
	}
	return observations
}

func phase34CheckRoute(diagnostics []diagnostic.Diagnostic) phase34RouteRecord {
	status := protocol.StatusPass
	if len(diagnostics) != 0 {
		status = protocol.StatusInvalid
	}
	return phase34RouteRecord{Outcome: status, Status: status, Diagnostics: phase34RouteProblems(diagnostics)}
}

func phase34CommandRoute(result protocol.Result, err error) phase34RouteRecord {
	if err != nil {
		return phase34RouteRecord{Outcome: "error", Error: err.Error(), Diagnostics: []phase34RouteProblem{}}
	}
	return phase34RouteRecord{Outcome: result.Status, Status: result.Status, Diagnostics: phase34RouteProblems(result.Diagnostics)}
}

func phase34RouteProblems(diagnostics []diagnostic.Diagnostic) []phase34RouteProblem {
	problems := make([]phase34RouteProblem, 0, len(diagnostics))
	for _, item := range diagnostics {
		encoded, err := json.Marshal(item)
		if err != nil {
			panic(err)
		}
		problems = append(problems, phase34RouteProblem{ID: item.ID, Code: item.Code, Primary: item.Primary, JSON: string(encoded)})
	}
	return problems
}

func phase34IdentityObservations(t *testing.T) ([]phase34IdentityObservation, map[string]string) {
	t.Helper()
	causeSpan := diagnostic.Span{Start: 3, End: 8}
	repairSpanA := diagnostic.Span{Start: 12, End: 18}
	repairSpanB := diagnostic.Span{Start: 28, End: 34}
	shiftedRepairSpanA := diagnostic.Span{Start: 112, End: 118}
	shiftedRepairSpanB := diagnostic.Span{Start: 128, End: 134}
	baseRepairs := []diagnostic.Repair{
		{Kind: "insert_take", Detail: "ownership", Span: &repairSpanA, Replacement: "take false", Applicability: diagnostic.ApplicabilityMachineApplicable},
		{Kind: "replace_borrow", Detail: "borrow", Span: &repairSpanB, Replacement: "borrow false", Applicability: diagnostic.ApplicabilityRequiresConfirmation},
	}
	shiftedRepairs := []diagnostic.Repair{
		{Kind: "insert_take", Detail: "ownership", Span: &shiftedRepairSpanA, Replacement: "take false", Applicability: diagnostic.ApplicabilityMachineApplicable},
		{Kind: "replace_borrow", Detail: "borrow", Span: &shiftedRepairSpanB, Replacement: "borrow false", Applicability: diagnostic.ApplicabilityRequiresConfirmation},
	}
	differentKindRepairs := append([]diagnostic.Repair(nil), baseRepairs...)
	differentKindRepairs[1].Kind = "different_repair_kind"

	base := diagnostic.ErrorWithRepairs("compat.identity", diagnostic.Span{Start: 40, End: 45}, "legacy repair diagnostic", []diagnostic.Cause{{Kind: "declared_here", Detail: "binding origin", Span: &causeSpan}}, baseRepairs...)
	shifted := diagnostic.ErrorWithRepairs("compat.identity", diagnostic.Span{Start: 40, End: 45}, "legacy repair diagnostic", []diagnostic.Cause{{Kind: "declared_here", Detail: "binding origin", Span: &causeSpan}}, shiftedRepairs...)
	differentKind := diagnostic.ErrorWithRepairs("compat.identity", diagnostic.Span{Start: 40, End: 45}, "legacy repair diagnostic", []diagnostic.Cause{{Kind: "declared_here", Detail: "binding origin", Span: &causeSpan}}, differentKindRepairs...)

	diagnostics := []struct {
		name       string
		diagnostic diagnostic.Diagnostic
	}{
		{"error-nil-causes", diagnostic.Error("compat.legacy", diagnostic.Span{Start: 1, End: 2}, "legacy error without causes")},
		{"error-nested-cause-span", diagnostic.Error("compat.legacy", diagnostic.Span{Start: 9, End: 10}, "legacy error with cause", diagnostic.Cause{Kind: "nested_origin", Detail: "declared", Span: &causeSpan})},
		{"error-with-repairs-nil-causes", diagnostic.ErrorWithRepairs("compat.identity", diagnostic.Span{Start: 40, End: 45}, "legacy repair diagnostic", nil, baseRepairs...)},
		{"error-with-repairs-nested-causes", base},
		{"error-with-repairs-shifted-coordinates", shifted},
		{"error-with-repairs-different-kind", differentKind},
		{"in-unit-source", base.InUnit("draft.identity")},
		{"in-unit-other", base.InUnit("other.identity")},
	}
	rows := make([]phase34IdentityObservation, 0, len(diagnostics))
	for _, item := range diagnostics {
		encoded, err := json.Marshal(item.diagnostic)
		if err != nil {
			t.Fatal(err)
		}
		row := phase34IdentityObservation{
			Name: item.name, Schema: item.diagnostic.Schema, DiagnosticID: item.diagnostic.ID,
			Unit: item.diagnostic.Unit, CausesNil: item.diagnostic.Causes == nil,
			MarshaledJSON: string(encoded), Repairs: make([]phase34RepairSnapshot, 0, len(item.diagnostic.Repairs)),
		}
		for _, repair := range item.diagnostic.Repairs {
			row.Repairs = append(row.Repairs, phase34RepairSnapshot{
				Kind: repair.Kind, Span: repair.Span, Replacement: repair.Replacement,
				Applicability: repair.Applicability, DriverEligible: diagnostic.DriverEligible(repair),
			})
		}
		rows = append(rows, row)
	}
	phase34AssertIdentityEvidence(t, rows)

	if base.ID != shifted.ID {
		t.Fatal("repair coordinate shifts unexpectedly changed the identity ID")
	}
	if base.ID == differentKind.ID {
		t.Fatal("changing the repair kind did not change the identity ID")
	}
	if base.InUnit("draft.identity").ID == base.InUnit("other.identity").ID {
		t.Fatal("different units unexpectedly share a diagnostic ID")
	}

	empty := protocol.New("check", protocol.StatusPass)
	withDiag := protocol.New("check", protocol.StatusInvalid)
	withDiag.Diagnostics = []diagnostic.Diagnostic{base}
	withUnitDiag := protocol.New("check", protocol.StatusInvalid)
	withUnitDiag.Diagnostics = []diagnostic.Diagnostic{base.InUnit("draft.identity")}
	ordered := protocol.New("check", protocol.StatusInvalid)
	ordered.Diagnostics = []diagnostic.Diagnostic{base, base.InUnit("draft.identity")}
	reversed := protocol.New("check", protocol.StatusInvalid)
	reversed.Diagnostics = []diagnostic.Diagnostic{base.InUnit("draft.identity"), base}
	metricsOnly := empty
	metricsOnly.Metrics.ElapsedNS = 17
	metricsOnly.Metrics.OutputBytes = 123
	resultIDs := map[string]string{
		"empty-pass":                empty.Finalize().ID,
		"empty-pass-metrics-varied": metricsOnly.Finalize().ID,
		"with-legacy-diagnostic":    withDiag.Finalize().ID,
		"with-unit-diagnostic":      withUnitDiag.Finalize().ID,
		"diagnostic-order-forward":  ordered.Finalize().ID,
		"diagnostic-order-reversed": reversed.Finalize().ID,
	}
	if resultIDs["empty-pass"] != resultIDs["empty-pass-metrics-varied"] {
		t.Fatal("measurement-only fields changed Result.Finalize ID")
	}
	if resultIDs["diagnostic-order-forward"] == resultIDs["diagnostic-order-reversed"] {
		t.Fatal("Result.Finalize ID did not retain diagnostic order")
	}
	return rows, resultIDs
}

func phase34AssertIdentityEvidence(t *testing.T, rows []phase34IdentityObservation) {
	t.Helper()
	byName := make(map[string]phase34IdentityObservation, len(rows))
	for _, row := range rows {
		byName[row.Name] = row
	}
	if !byName["error-nil-causes"].CausesNil || byName["error-nested-cause-span"].CausesNil {
		t.Fatal("legacy Error evidence must retain nil and non-nil cause cases")
	}
	base, shifted := byName["error-with-repairs-nested-causes"], byName["error-with-repairs-shifted-coordinates"]
	if base.DiagnosticID != shifted.DiagnosticID || base.MarshaledJSON == shifted.MarshaledJSON {
		t.Fatal("repair coordinate shift must change serialized repair bytes without changing diagnostic identity")
	}
	if len(base.Repairs) != 2 || !base.Repairs[0].DriverEligible || base.Repairs[1].DriverEligible {
		t.Fatalf("repair snapshot must contain both eligible and non-eligible byte replacements: %+v", base.Repairs)
	}
	if byName["in-unit-source"].Unit == byName["in-unit-other"].Unit || byName["in-unit-source"].DiagnosticID == byName["in-unit-other"].DiagnosticID {
		t.Fatal("InUnit identity snapshots must distinguish unit attribution")
	}
}

func phase34LoadCompatibilityBaseline(t *testing.T) phase34CompatibilityBaseline {
	t.Helper()
	data, err := os.ReadFile(phase34IdentityBaselinePath())
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if got := hex.EncodeToString(digest[:]); got != phase34IdentityBaselineHash {
		t.Fatalf("immutable identity baseline changed: sha256 %s, want %s", got, phase34IdentityBaselineHash)
	}
	var baseline phase34CompatibilityBaseline
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatalf("decode identity baseline: %v", err)
	}
	return baseline
}

func TestPhase34BoolNameBaseline(t *testing.T) {
	baseline := phase34LoadCompatibilityBaseline(t)
	if baseline.Schema != phase34IdentitySchema || baseline.Revision == "" {
		t.Fatalf("invalid compatibility baseline header: schema=%q revision=%q", baseline.Schema, baseline.Revision)
	}
	byID := make(map[string]phase34BoolObservation, len(baseline.BoolCases))
	expectedCases := phase34BoolCases()
	for index, item := range baseline.BoolCases {
		if index >= len(expectedCases) || item.ID != expectedCases[index].ID || item.Source != expectedCases[index].Source || item.ModuleSource != expectedCases[index].ModuleSource {
			t.Fatalf("historical bool/name source row %d does not match the immutable case definition: %+v", index, item)
		}
		byID[item.ID] = item
	}
	if len(byID) != len(phase34BoolCases()) {
		t.Fatalf("baseline has %d bool/name cases, want %d", len(byID), len(phase34BoolCases()))
	}
	unbound, ok := byID["unbound-true-tail"]
	if !ok || !phase34RouteHasCode(unbound.Check, "name.unknown") || !phase34RouteHasCode(unbound.Command, "name.unknown") {
		t.Fatalf("historical unbound true tail must record name.unknown at both prior routes: %+v", unbound)
	}

	observed := phase34ObserveBoolCases(t, false)
	for _, item := range observed {
		if item.ID == "imported-true-function" {
			if item.Check.Outcome != protocol.StatusInvalid || !phase34RouteHasCode(item.Check, "module.imports_not_loaded") {
				t.Fatalf("source-only Check must retain its import-resolution refusal for the imported fixture: %+v", item.Check)
			}
			if item.Command.Outcome != protocol.StatusPass {
				t.Fatalf("module-aware CheckCommandFile did not admit the imported legal name: %+v", item.Command)
			}
			continue
		}
		if item.Check.Outcome != protocol.StatusPass {
			t.Fatalf("legal Bool/name case %q was not admitted by Check: %+v", item.ID, item.Check)
		}
		if item.Command.Outcome != protocol.StatusPass {
			t.Fatalf("legal Bool/name case %q was not admitted by CheckCommandFile: %+v", item.ID, item.Command)
		}
	}
}

func phase34RouteHasCode(route phase34RouteRecord, code string) bool {
	for _, problem := range route.Diagnostics {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func TestPhase34IdentityBaseline(t *testing.T) {
	baseline := phase34LoadCompatibilityBaseline(t)
	if baseline.Schema != phase34IdentitySchema {
		t.Fatalf("identity baseline schema %q, want %q", baseline.Schema, phase34IdentitySchema)
	}
	identity, resultIDs := phase34IdentityObservations(t)
	if !reflect.DeepEqual(identity, baseline.IdentityCases) {
		want, _ := json.Marshal(baseline.IdentityCases)
		got, _ := json.Marshal(identity)
		t.Fatalf("legacy diagnostic bytes or repair facts changed:\n got %s\nwant %s", got, want)
	}
	if !reflect.DeepEqual(resultIDs, baseline.ResultIDs) {
		t.Fatalf("Result.Finalize IDs changed: got %v want %v", resultIDs, baseline.ResultIDs)
	}
}
