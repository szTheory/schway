package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
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

type phase34CompatSpan struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

type phase34CompatCause struct {
	Kind   string             `json:"kind"`
	Span   *phase34CompatSpan `json:"span,omitempty"`
	Unit   string             `json:"unit,omitempty"`
	Detail string             `json:"detail,omitempty"`
}

type phase34CompatDiagnostic struct {
	ID          string               `json:"id"`
	Code        string               `json:"code"`
	PrimarySpan phase34CompatSpan    `json:"primary_span"`
	Causes      []phase34CompatCause `json:"causes"`
}

type phase34CompatDraft struct {
	ID          string                    `json:"id"`
	Source      string                    `json:"source"`
	Diagnostics []phase34CompatDiagnostic `json:"diagnostics"`
}

type phase34CompatDelta struct {
	Fixture     string   `json:"fixture"`
	Authority   []string `json:"authority"`
	BeforeCount int      `json:"before_count"`
	AfterCount  int      `json:"after_count"`
	Removed     []string `json:"removed"`
	Added       []string `json:"added"`
}

type phase34CompatDeltaLedger struct {
	Schema string               `json:"schema"`
	Drafts []phase34CompatDelta `json:"drafts"`
	Bool   phase34BoolNameDelta `json:"bool_name"`
}

type phase34BoolRouteState struct {
	Outcome     string                    `json:"outcome"`
	Count       int                       `json:"count"`
	Diagnostics []phase34CompatDiagnostic `json:"diagnostics"`
}

type phase34BoolNameDelta struct {
	Fixture       string                `json:"fixture"`
	Authority     string                `json:"authority"`
	BeforeCheck   phase34BoolRouteState `json:"before_check"`
	BeforeCommand phase34BoolRouteState `json:"before_command"`
	AfterCheck    phase34BoolRouteState `json:"after_check"`
	AfterCommand  phase34BoolRouteState `json:"after_command"`
}

type phase34CompatDraftBaseline struct {
	Schema   string                     `json:"schema"`
	Revision string                     `json:"revision"`
	Drafts   []phase34CompatDraftBefore `json:"drafts"`
}

type phase34CompatDraftBefore struct {
	ID          string                    `json:"id"`
	Source      string                    `json:"source"`
	Diagnostics []phase34CompatDiagnostic `json:"diagnostics"`
}

type phase34CompatDraftAfter struct {
	ID              string                    `json:"id"`
	DiagnosticCount int                       `json:"diagnostic_count"`
	Diagnostics     []phase34CompatDiagnostic `json:"diagnostics"`
}

type phase35RecoveryCompatAfter struct {
	Schema     string                           `json:"schema"`
	Date       string                           `json:"date"`
	Amendments []phase35RecoveryCompatAmendment `json:"amendments"`
}

type phase35RecoveryCompatAmendment struct {
	Fixture           string                          `json:"fixture"`
	Reason            string                          `json:"reason"`
	Phase34Projection phase35RecoveryCompatProjection `json:"phase34_projection"`
	Phase35Projection phase35RecoveryCompatProjection `json:"phase35_projection"`
}

type phase35RecoveryCompatProjection struct {
	DiagnosticCount       int                       `json:"diagnostic_count"`
	RecoveredSiblingNames []string                  `json:"recovered_sibling_names"`
	Diagnostics           []phase34CompatDiagnostic `json:"diagnostics"`
}

const phase34DraftBaselineSHA256ForCompatibility = "6607c9b5e3657c0e65ba6e46d16a1dc281153994f27ca1ab8c645f9e935b8fae"

func TestPhase34CompatibilityDeltaLedger(t *testing.T) {
	baselineBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase34", "draft-baseline.json"))
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(baselineBytes)
	if got := hex.EncodeToString(digest[:]); got != phase34DraftBaselineSHA256ForCompatibility {
		t.Fatalf("immutable draft baseline changed: sha256 %s, want %s", got, phase34DraftBaselineSHA256ForCompatibility)
	}
	var baseline phase34CompatDraftBaseline
	if err := json.Unmarshal(baselineBytes, &baseline); err != nil {
		t.Fatal(err)
	}
	var after struct {
		Schema string                    `json:"schema"`
		Drafts []phase34CompatDraftAfter `json:"drafts"`
	}
	afterBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase34", "draft-after.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(afterBytes, &after); err != nil {
		t.Fatal(err)
	}
	phase35AfterBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "recovery-after.json"))
	if err != nil {
		t.Fatalf("read Phase 35 recovery amendment: %v", err)
	}
	var phase35After phase35RecoveryCompatAfter
	if err := json.Unmarshal(phase35AfterBytes, &phase35After); err != nil {
		t.Fatalf("decode Phase 35 recovery amendment: %v", err)
	}
	phase34AssertCRC32HistoricalDeltaFiles(t)
	if baseline.Schema != "phase34.draft-baseline/1" || after.Schema != "phase34.draft-after/1" || len(baseline.Drafts) != 20 || len(after.Drafts) != 20 {
		t.Fatalf("draft corpus headers/counts changed: before=%q/%d after=%q/%d", baseline.Schema, len(baseline.Drafts), after.Schema, len(after.Drafts))
	}
	if phase35After.Schema != "phase35.recovery-after/1" || phase35After.Date != "2026-10-14" || len(phase35After.Amendments) != 1 || phase35After.Amendments[0].Fixture != "crc32-reconstructed" || strings.TrimSpace(phase35After.Amendments[0].Reason) == "" {
		t.Fatalf("Phase 35 recovery amendment header/rows changed: %+v", phase35After)
	}

	current := make(map[string][]phase34CompatDiagnostic, len(baseline.Drafts))
	currentSiblings := make(map[string][]string, len(baseline.Drafts))
	for _, draft := range baseline.Drafts {
		parsed := syntax.Parse([]byte(draft.Source))
		problems := phase34CompatProblems(parsed.Diagnostics)
		current[draft.ID] = problems
		siblings := make([]string, 0, len(parsed.Program.Funcs))
		for _, function := range parsed.Program.Funcs {
			siblings = append(siblings, function.Name)
		}
		currentSiblings[draft.ID] = siblings
	}
	unboundBefore := phase34FindBoolObservation(t, phase34LoadCompatibilityBaseline(t).BoolCases, "unbound-true-tail")
	unboundAfter := phase34FindBoolObservation(t, phase34ObserveBoolCases(t, true), "unbound-true-tail")
	if unboundBefore == nil || unboundAfter == nil {
		t.Fatal("unbound Bool-tail before/after witness is missing")
	}
	ledgerBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase34", "compat-deltas.json"))
	if err != nil {
		t.Fatalf("read literal compatibility delta ledger: %v", err)
	}
	var ledger phase34CompatDeltaLedger
	if err := json.Unmarshal(ledgerBytes, &ledger); err != nil {
		t.Fatalf("decode literal compatibility delta ledger: %v", err)
	}
	if ledger.Schema != "phase34.compat-deltas/1" {
		t.Fatalf("compat delta schema %q, want phase34.compat-deltas/1", ledger.Schema)
	}
	baseByID := make(map[string]phase34CompatDraftBefore, len(baseline.Drafts))
	afterByID := make(map[string]phase34CompatDraftAfter, len(after.Drafts))
	for _, item := range baseline.Drafts {
		baseByID[item.ID] = item
	}
	for _, item := range after.Drafts {
		if _, duplicate := afterByID[item.ID]; duplicate {
			t.Errorf("draft-after repeats %q", item.ID)
		}
		afterByID[item.ID] = item
		if item.ID == phase35After.Amendments[0].Fixture {
			amendment := phase35After.Amendments[0]
			if amendment.Phase34Projection.DiagnosticCount != item.DiagnosticCount || len(item.Diagnostics) != item.DiagnosticCount || !phase34EqualCompatProjection(amendment.Phase34Projection.Diagnostics, item.Diagnostics) {
				t.Errorf("%s Phase 35 amendment does not retain the exact historical Phase 34 code/span/cause projection", item.ID)
			}
			if amendment.Phase35Projection.DiagnosticCount != len(current[item.ID]) || amendment.Phase35Projection.DiagnosticCount != len(amendment.Phase35Projection.Diagnostics) || !phase34EqualCompatProjection(current[item.ID], amendment.Phase35Projection.Diagnostics) || !reflect.DeepEqual(currentSiblings[item.ID], amendment.Phase35Projection.RecoveredSiblingNames) {
				t.Errorf("%s current parser projection differs from the literal Phase 35 diagnostic/sibling row: diagnostics=%+v siblings=%v", item.ID, current[item.ID], currentSiblings[item.ID])
			}
		} else if len(current[item.ID]) != item.DiagnosticCount || !phase34EqualCompatProjection(current[item.ID], item.Diagnostics) {
			t.Errorf("%s current diagnostics differ from frozen draft-after code/span/cause values", item.ID)
		}
	}
	allowedAuthority := map[string][]string{
		"logical-and":                {"surface.unsupported_logical_and"},
		"logical-or":                 {"surface.unsupported_logical_or"},
		"logical-not":                {"surface.unsupported_logical_not"},
		"schway-arity-two":           {"surface.unsupported_parameter_arity"},
		"if-no-else":                 {"surface.unsupported_if_without_else"},
		"subtraction":                {"surface.unsupported_subtraction"},
		"malformed-nested-paren":     {"recovery.expected_call_argument"},
		"malformed-nested-block":     {"surface.unsupported_if_without_else", "recovery.local_assignment"},
		"sibling-after-error":        {"recovery.exact_parser_owned_token"},
		"two-independent-causes":     {"recovery.exact_parser_owned_token"},
		"independent-missing-closer": {"recovery.delimiter_local"},
		"eof-empty":                  {"recovery.module_prefix"},
		"malformed-near-miss":        {"recovery.local_boundary"},
		"crc32-reconstructed":        {"recovery.delimiter_local"},
		"word-counter-reconstructed": {"surface.unsupported_logical_or", "surface.unsupported_logical_not", "surface.unsupported_if_without_else", "recovery.exact_parser_owned_token"},
	}
	deltaByID := make(map[string]phase34CompatDelta, len(ledger.Drafts))
	for _, delta := range ledger.Drafts {
		if _, duplicate := deltaByID[delta.Fixture]; duplicate {
			t.Errorf("compat delta ledger repeats fixture %q", delta.Fixture)
		}
		deltaByID[delta.Fixture] = delta
		before, exists := baseByID[delta.Fixture]
		if !exists {
			t.Errorf("%s is absent from immutable draft baseline", delta.Fixture)
			continue
		}
		if delta.Fixture == "crc32-reconstructed" {
			wantHistorical := phase34ExpectedCRC32HistoricalDelta()
			if !reflect.DeepEqual(delta, wantHistorical) {
				t.Errorf("crc32-reconstructed historical Phase 34 delta changed: got=%+v want=%+v", delta, wantHistorical)
			}
			if got := allowedAuthority[delta.Fixture]; !reflect.DeepEqual(got, delta.Authority) {
				t.Errorf("%s authority %v is not its exact authorized cause %v", delta.Fixture, delta.Authority, got)
			}
			if historical, ok := afterByID[delta.Fixture]; !ok || historical.DiagnosticCount != 2 || len(historical.Diagnostics) != 2 || !phase34EqualCompatProjection(phase35After.Amendments[0].Phase34Projection.Diagnostics, historical.Diagnostics) {
				t.Errorf("crc32-reconstructed historical Phase 34 after projection changed: %+v", historical)
			}
			if delta.BeforeCount != len(before.Diagnostics) || delta.AfterCount != afterByID[delta.Fixture].DiagnosticCount {
				t.Errorf("crc32-reconstructed historical counts changed: baseline=%d ledger=%d/%d after=%d", len(before.Diagnostics), delta.BeforeCount, delta.AfterCount, afterByID[delta.Fixture].DiagnosticCount)
			}
			continue
		}
		removed, added := phase34CompatDifference(before.Diagnostics, current[delta.Fixture])
		expectedRemoved := append([]string{}, delta.Removed...)
		expectedAdded := append([]string{}, delta.Added...)
		sort.Strings(expectedRemoved)
		sort.Strings(expectedAdded)
		if delta.BeforeCount != len(before.Diagnostics) || delta.AfterCount != len(current[delta.Fixture]) ||
			!reflect.DeepEqual(removed, expectedRemoved) || !reflect.DeepEqual(added, expectedAdded) {
			t.Errorf("%s exact before/after delta differs: before=%d/%d removed=%v/%v after=%d/%d added=%v/%v", delta.Fixture, delta.BeforeCount, len(before.Diagnostics), delta.Removed, removed, delta.AfterCount, len(current[delta.Fixture]), delta.Added, added)
		}
		if got := allowedAuthority[delta.Fixture]; !reflect.DeepEqual(got, delta.Authority) {
			t.Errorf("%s authority %v is not its exact authorized cause %v", delta.Fixture, delta.Authority, got)
		}
		if _, exists := afterByID[delta.Fixture]; !exists {
			t.Errorf("%s is absent from draft-after.json", delta.Fixture)
		}
	}
	for _, before := range baseline.Drafts {
		if before.ID == phase35After.Amendments[0].Fixture {
			if _, listed := deltaByID[before.ID]; !listed {
				t.Errorf("%s has no frozen historical Phase 34 delta row", before.ID)
			}
			continue
		}
		if reflect.DeepEqual(before.Diagnostics, current[before.ID]) {
			if _, listed := deltaByID[before.ID]; listed {
				t.Errorf("%s has no observed change but is listed as an authorized delta", before.ID)
			}
			continue
		}
		if _, listed := deltaByID[before.ID]; !listed {
			t.Errorf("%s changed without an exact fixture delta", before.ID)
		}
	}
	for fixture := range allowedAuthority {
		if _, listed := deltaByID[fixture]; !listed {
			t.Errorf("authorized fixture %s has no literal delta row", fixture)
		}
	}

	if ledger.Bool.Fixture != "unbound-true-tail" || ledger.Bool.Authority != "surface.unsupported_bool_literal_tail" ||
		!phase34BoolRouteEqual(unboundBefore.Check, ledger.Bool.BeforeCheck) || !phase34BoolRouteEqual(unboundBefore.Command, ledger.Bool.BeforeCommand) ||
		!phase34BoolRouteEqual(unboundAfter.Check, ledger.Bool.AfterCheck) || !phase34BoolRouteEqual(unboundAfter.Command, ledger.Bool.AfterCommand) {
		t.Fatal("unbound Bool-tail literal before/after delta differs from the frozen name baseline or live check")
	}
}

func phase34ExpectedCRC32HistoricalDelta() phase34CompatDelta {
	return phase34CompatDelta{
		Fixture:     "crc32-reconstructed",
		Authority:   []string{"recovery.delimiter_local"},
		BeforeCount: 5,
		AfterCount:  2,
		Removed: []string{
			"diagnostic:e3ac47f79962d16770d1e475|syntax.unexpected_byte|184:185",
			"diagnostic:c9914990956616f1356ba6e9|syntax.expected_rbrace|194:197",
			"diagnostic:72a06b90ea232dc76b9b7aca|syntax.expected_declaration|198:199",
		},
		Added: []string{},
	}
}

func phase34AssertCRC32HistoricalDeltaFiles(t *testing.T) {
	t.Helper()
	want := map[string]string{
		"testdata/phase34/drafts.json":         "8774033822c9f47db8236d00a347e7e176a9524437314c182be8555f16a1b9df",
		"testdata/phase34/draft-baseline.json": "6607c9b5e3657c0e65ba6e46d16a1dc281153994f27ca1ab8c645f9e935b8fae",
		"testdata/phase34/draft-after.json":    "1e11ff4a8af6cc03c5d6e3200d0cdf9ef7dad41ee28e6a6fcd3aa2362b272bf3",
		"testdata/phase34/compat-deltas.json":  "356d6f087023b46542636bfa244266afe1fbcd73cdb89120cc3d33802f558011",
	}
	for path, digestWant := range want {
		contents, err := os.ReadFile(testsupport.ProjectPath(path))
		if err != nil {
			t.Fatalf("read frozen Phase 34 compatibility input %s: %v", path, err)
		}
		digest := sha256.Sum256(contents)
		if got := hex.EncodeToString(digest[:]); got != digestWant {
			t.Errorf("frozen Phase 34 compatibility input %s changed: sha256 %s, want %s", path, got, digestWant)
		}
	}
}

func phase34CompatDifference(before, after []phase34CompatDiagnostic) ([]string, []string) {
	matched := make([]bool, len(after))
	removed := make([]string, 0)
	for _, old := range before {
		found := -1
		for index, current := range after {
			if !matched[index] && reflect.DeepEqual(old, current) {
				found = index
				break
			}
		}
		if found >= 0 {
			matched[found] = true
		} else {
			removed = append(removed, phase34CompatSignature(old))
		}
	}
	added := make([]string, 0)
	for index, current := range after {
		if !matched[index] {
			added = append(added, phase34CompatSignature(current))
		}
	}
	sort.Strings(removed)
	sort.Strings(added)
	return removed, added
}

func phase34CompatSignature(problem phase34CompatDiagnostic) string {
	return fmt.Sprintf("%s|%s|%d:%d", problem.ID, problem.Code, problem.PrimarySpan.Start, problem.PrimarySpan.End)
}

func phase34CompatProblems(diagnostics []diagnostic.Diagnostic) []phase34CompatDiagnostic {
	problems := make([]phase34CompatDiagnostic, 0, len(diagnostics))
	for _, item := range diagnostics {
		problem := phase34CompatDiagnostic{ID: item.ID, Code: item.Code, PrimarySpan: phase34CompatSpan{Start: item.Primary.Start, End: item.Primary.End}, Causes: make([]phase34CompatCause, 0, len(item.Causes))}
		for _, cause := range item.Causes {
			var span *phase34CompatSpan
			if cause.Span != nil {
				span = &phase34CompatSpan{Start: cause.Span.Start, End: cause.Span.End}
			}
			problem.Causes = append(problem.Causes, phase34CompatCause{Kind: cause.Kind, Span: span, Unit: cause.Unit, Detail: cause.Detail})
		}
		problems = append(problems, problem)
	}
	return problems
}

func phase34EqualCompatProjection(actual, expected []phase34CompatDiagnostic) bool {
	if len(actual) != len(expected) {
		return false
	}
	for index := range actual {
		if actual[index].Code != expected[index].Code || actual[index].PrimarySpan != expected[index].PrimarySpan || !reflect.DeepEqual(actual[index].Causes, expected[index].Causes) {
			return false
		}
	}
	return true
}

func phase34FindBoolObservation(t *testing.T, observations []phase34BoolObservation, id string) *phase34BoolObservation {
	t.Helper()
	for index := range observations {
		if observations[index].ID == id {
			return &observations[index]
		}
	}
	return nil
}

func phase34BoolState(route phase34RouteRecord) phase34BoolRouteState {
	return phase34BoolRouteState{Outcome: route.Outcome, Count: len(route.Diagnostics), Diagnostics: phase34CompatRouteDiagnostics(route.Diagnostics)}
}

func phase34CompatRouteDiagnostics(problems []phase34RouteProblem) []phase34CompatDiagnostic {
	result := make([]phase34CompatDiagnostic, 0, len(problems))
	for _, item := range problems {
		result = append(result, phase34CompatDiagnostic{ID: item.ID, Code: item.Code, PrimarySpan: phase34CompatSpan{Start: item.Primary.Start, End: item.Primary.End}})
	}
	return result
}

func phase34BoolRouteEqual(route phase34RouteRecord, state phase34BoolRouteState) bool {
	return route.Outcome == state.Outcome && len(route.Diagnostics) == state.Count && reflect.DeepEqual(phase34CompatRouteDiagnostics(route.Diagnostics), state.Diagnostics)
}

func phase34CompactRoute(t *testing.T, route phase34RouteRecord) string {
	t.Helper()
	encoded, err := json.Marshal(phase34BoolState(route))
	if err != nil {
		t.Fatal(err)
	}
	return string(encoded)
}

type phase34TransitionFinding struct {
	Document string
	Code     string
	Row      string
	Field    string
}

type phase34TransitionRow struct {
	Key           string
	Status        string
	Owner         string
	RequiredFacts []struct {
		Field string
		Text  []string
	}
}

var phase34CapabilityHeader = []string{
	"Program", "Status", "Blocker", "Smallest complete slice", "Checker/guarantee", "Evidence/debt", "Owner/next action", "Reprioritization observation",
}

var phase34SeedHeader = []string{"Seed", "Disposition", "Current source witness", "Named consumer", "Trigger/reason", "Next action"}
var phase34ProvenanceHeader = []string{"Kind", "Status", "Evidence/claim"}

var phase34TransitionCapabilities = []phase34TransitionRow{
	{
		Key: "Broken CRC/word-counter check and schway surface --json", Status: "implemented", Owner: "Phase 34",
		RequiredFacts: []struct {
			Field string
			Text  []string
		}{
			{"Blocker", []string{"None remaining"}},
			{"Smallest complete slice", []string{"runnable", "source-free"}},
			{"Checker/guarantee", []string{"byte identity", "repair authority"}},
			{"Evidence/debt", []string{"DX-20", "TestPhase34SourceRoutePresentation", "TestPhase34SurfaceManifestWitnesses", "TestPhase34PreChangeCProgramGolden"}},
			{"Reprioritization observation", []string{"admitted", "independent cause"}},
		},
	},
	{
		Key: "CRC-32 operator expression x ^ (y << 3) & 0xFF and Adler b << 16", Status: "planned", Owner: "Phase 35",
		RequiredFacts: []struct {
			Field string
			Text  []string
		}{
			{"Blocker", []string{"bitwise", "shift"}},
			{"Smallest complete slice", []string{"literal shifts", "independent vectors"}},
			{"Checker/guarantee", []string{"literal", "range", "shift"}},
			{"Evidence/debt", []string{"BIT-01", "EVD-17"}},
			{"Reprioritization observation", []string{"dynamic shift", "different bounds"}},
		},
	},
	{
		Key: "Imported crc_step scalar helper inside a loop", Status: "planned", Owner: "Phase 36; file integration in Phase 37",
		RequiredFacts: []struct {
			Field string
			Text  []string
		}{
			{"Blocker", []string{"scalar helper calls"}},
			{"Smallest complete slice", []string{"arity-one", "file consumer"}},
			{"Checker/guarantee", []string{"arity-1", "acyclic", "result-defect"}},
			{"Evidence/debt", []string{"CALL-01", "APP-13", "MOD-07"}},
			{"Reprioritization observation", []string{"second consumer", "arity-N"}},
		},
	},
}

var phase34SeedExpectations = []struct {
	Key, Witness, Consumer, Trigger string
	ExtraWitnesses                  []string
	NextFacts                       []string
}{
	{"z1g", "internal/compiler/session/session_phase6_query.go#QueryCommandFile", "crc_step declaration and all direct call occurrences", "measured named query gap", []string{"internal/compiler/session/session_phase6_explain.go#ExplainCommandFile"}, []string{"remain dormant"}},
	{"trz", "internal/compiler/native/native_app.go#readApplicationCapture", "one named checksum/file utility failure", "failure unexplained by current capture, observers, and replay", []string{"internal/compiler/native/native_phase33_capture_test.go#TestPhase33CaptureReadBounded", "internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile"}, []string{"Phase 33", "remain dormant"}},
	{"z1s", "internal/compiler/syntax/format.go#Format", "CRC/word-counter canonical source review", "named canonical-source or contract dispute", nil, []string{"remain dormant"}},
}

var phase34ProvenanceExpectations = []struct {
	Kind, Status string
	Evidence     []string
}{
	{"Source inspection", "inspected", []string{"session.go", "repair.go", "PRODUCT-ROADMAP.md"}},
	{"Focused execution", "passed", []string{"TestPhase34SourceRoutePresentation", "TestPhase34DiagnosticRouteInventory", "TestPhase34RepairJSONCompatibility", "TestPhase34CompatibilityDeltaLedger", "TestPhase34PreChangeCProgramGolden"}},
	{"Historical M006 receipts", "historical", []string{"M006", "historical"}},
	{"Phase 38 hosted binding", "pending", []string{"Phase 38", "not yet bound"}},
	{"Human readability", "unmeasured", []string{"no human review", "unmeasured"}},
	{"Whole-loop latency", "unmeasured", []string{"no cold/warm", "unmeasured"}},
}

const phase34TransitionHeading = "## Phase 34 transition amendment — 2026-10-13"

func TestPhase34TransitionDocuments(t *testing.T) {
	documents := map[string]string{}
	for _, name := range []string{"PRODUCT-ROADMAP.md", "LANGUAGE-MATURITY.md"} {
		data, err := os.ReadFile(testsupport.ProjectPath(".planning", name))
		if err != nil {
			t.Fatal(err)
		}
		documents[name] = string(data)
		if problems := phase34TransitionProblems(name, documents[name]); len(problems) != 0 {
			t.Fatalf("%s Phase 34 amendment findings: %+v", name, problems)
		}
	}

	for document, original := range documents {
		capabilityRows := phase34ParseTable(phase34ExtractTransitionBlock(t, original), phase34CapabilityHeader)
		for _, row := range phase34TransitionCapabilities {
			for _, field := range phase34CapabilityHeader {
				mutated := phase34MutateTransitionCell(t, original, phase34CapabilityHeader, "Program", row.Key, field, "")
				phase34ExpectSingleFinding(t, document, original, mutated, "missing_capability_field", row.Key, field)
			}
			if _, exists := capabilityRows[row.Key]; !exists {
				t.Fatalf("clean %s amendment lacks capability row %q before mutation controls", document, row.Key)
			}
			mutated := phase34DeleteTransitionRow(t, original, phase34CapabilityHeader, "Program", row.Key)
			phase34ExpectSingleFinding(t, document, original, mutated, "missing_capability_field", row.Key, "Program row")
			wrongStatus := "planned"
			if row.Status == "planned" {
				wrongStatus = "implemented"
			}
			phase34ExpectCellFinding(t, document, original, phase34CapabilityHeader, "Program", row.Key, "Status", wrongStatus, "wrong_capability_status", row.Key, "Status")
			phase34ExpectCellFinding(t, document, original, phase34CapabilityHeader, "Program", row.Key, "Owner/next action", "Phase 99", "wrong_capability_owner", row.Key, "Owner/next action")
		}
		for _, seed := range phase34SeedExpectations {
			for _, field := range phase34SeedHeader {
				mutated := phase34MutateTransitionCell(t, original, phase34SeedHeader, "Seed", seed.Key, field, "")
				phase34ExpectSingleFinding(t, document, original, mutated, "missing_seed_field", seed.Key, field)
			}
			mutated := phase34DeleteTransitionRow(t, original, phase34SeedHeader, "Seed", seed.Key)
			phase34ExpectSingleFinding(t, document, original, mutated, "missing_seed_field", seed.Key, "Seed row")
			phase34ExpectCellFinding(t, document, original, phase34SeedHeader, "Seed", seed.Key, "Disposition", "run experiment", "wrong_seed_disposition", seed.Key, "Disposition")
		}
		for _, provenance := range phase34ProvenanceExpectations {
			mutated := phase34DeleteTransitionRow(t, original, phase34ProvenanceHeader, "Kind", provenance.Kind)
			phase34ExpectSingleFinding(t, document, original, mutated, "missing_provenance", provenance.Kind, "row")
			mutated = phase34MutateTransitionCell(t, original, phase34ProvenanceHeader, "Kind", provenance.Kind, "Status", "")
			phase34ExpectSingleFinding(t, document, original, mutated, "missing_provenance", provenance.Kind, "Status")
		}
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Phase 38 hosted binding", "Status", "passed", "evidence_overclaim", "Phase 38 hosted binding", "Status")
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Human readability", "Status", "measured", "evidence_overclaim", "Human readability", "Status")
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Whole-loop latency", "Status", "measured", "evidence_overclaim", "Whole-loop latency", "Status")
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Phase 38 hosted binding", "Evidence/claim", "current hosted receipt passed", "evidence_overclaim", "Phase 38 hosted binding", "Evidence/claim")
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Human readability", "Evidence/claim", "human readability measured", "evidence_overclaim", "Human readability", "Evidence/claim")
		phase34ExpectCellFinding(t, document, original, phase34ProvenanceHeader, "Kind", "Whole-loop latency", "Evidence/claim", "whole-loop latency measured with p50/p95 measured", "evidence_overclaim", "Whole-loop latency", "Evidence/claim")
		if documents[document] != original {
			t.Fatalf("%s in-memory mutation controls did not restore their source document", document)
		}
	}
}

func phase34TransitionProblems(document, text string) []phase34TransitionFinding {
	block, found, duplicate := phase34TransitionBlock(text)
	if !found || duplicate {
		return []phase34TransitionFinding{{Document: document, Code: "missing_section", Row: phase34TransitionHeading, Field: "unique section"}}
	}
	var problems []phase34TransitionFinding
	capabilityRows := phase34ParseTable(block, phase34CapabilityHeader)
	if len(capabilityRows) > len(phase34TransitionCapabilities) {
		problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: "unexpected capability", Field: "Program"})
	}
	for _, expected := range phase34TransitionCapabilities {
		cells, ok := capabilityRows[expected.Key]
		if !ok {
			cells, ok = phase34InferBlankKeyRow(capabilityRows, "Program", "Owner/next action", expected.Owner)
		}
		if !ok {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: expected.Key, Field: "Program row"})
			continue
		}
		for _, field := range phase34CapabilityHeader {
			if strings.TrimSpace(cells[field]) == "" {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: expected.Key, Field: field})
			}
		}
		if got := cells["Status"]; got != "" && got != expected.Status {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "wrong_capability_status", Row: expected.Key, Field: "Status"})
		}
		if got := cells["Owner/next action"]; got != "" && got != expected.Owner {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "wrong_capability_owner", Row: expected.Key, Field: "Owner/next action"})
		}
		for _, facts := range expected.RequiredFacts {
			if strings.TrimSpace(cells[facts.Field]) == "" {
				continue
			}
			for _, fact := range facts.Text {
				if !strings.Contains(cells[facts.Field], fact) {
					problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: expected.Key, Field: facts.Field})
					break
				}
			}
		}
	}

	seedRows := phase34ParseTable(block, phase34SeedHeader)
	if len(seedRows) > len(phase34SeedExpectations) {
		problems = append(problems, phase34TransitionFinding{Document: document, Code: "wrong_seed_disposition", Row: "unexpected seed", Field: "Disposition"})
	}
	for _, expected := range phase34SeedExpectations {
		cells, ok := seedRows[expected.Key]
		if !ok {
			cells, ok = phase34InferBlankKeyRow(seedRows, "Seed", "Named consumer", expected.Consumer)
		}
		if !ok {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: "Seed row"})
			continue
		}
		for _, field := range phase34SeedHeader {
			if strings.TrimSpace(cells[field]) == "" {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: field})
			}
		}
		if cells["Disposition"] != "" && cells["Disposition"] != "retain dormant" {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "wrong_seed_disposition", Row: expected.Key, Field: "Disposition"})
		}
		for field, required := range map[string]string{"Current source witness": expected.Witness, "Named consumer": expected.Consumer, "Trigger/reason": expected.Trigger} {
			if strings.TrimSpace(cells[field]) != "" && !strings.Contains(cells[field], required) {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: field})
			}
		}
		for _, required := range expected.ExtraWitnesses {
			if strings.TrimSpace(cells["Current source witness"]) != "" && !strings.Contains(cells["Current source witness"], required) {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: "Current source witness"})
			}
		}
		for _, reference := range append([]string{expected.Witness}, expected.ExtraWitnesses...) {
			if !phase34ReferenceExists(reference) {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: "Current source witness"})
			}
		}
		for _, required := range expected.NextFacts {
			if strings.TrimSpace(cells["Next action"]) != "" && !strings.Contains(cells["Next action"], required) {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_seed_field", Row: expected.Key, Field: "Next action"})
			}
		}
	}

	provenanceRows := phase34ParseTable(block, phase34ProvenanceHeader)
	if len(provenanceRows) > len(phase34ProvenanceExpectations) {
		problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_provenance", Row: "unexpected evidence", Field: "Kind"})
	}
	for _, expected := range phase34ProvenanceExpectations {
		cells, ok := provenanceRows[expected.Kind]
		if !ok {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_provenance", Row: expected.Kind, Field: "row"})
			continue
		}
		if strings.TrimSpace(cells["Status"]) == "" || strings.TrimSpace(cells["Evidence/claim"]) == "" {
			field := "Status"
			if strings.TrimSpace(cells["Evidence/claim"]) == "" {
				field = "Evidence/claim"
			}
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_provenance", Row: expected.Kind, Field: field})
			continue
		}
		if cells["Status"] != expected.Status {
			if expected.Kind == "Phase 38 hosted binding" || expected.Kind == "Human readability" || expected.Kind == "Whole-loop latency" || expected.Kind == "Historical M006 receipts" {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "evidence_overclaim", Row: expected.Kind, Field: "Status"})
			} else {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_provenance", Row: expected.Kind, Field: "Status"})
			}
		}
		claim := strings.ToLower(cells["Evidence/claim"])
		if overclaim := phase34OverclaimText(claim); overclaim != "" {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "evidence_overclaim", Row: expected.Kind, Field: "Evidence/claim"})
			continue
		}
		for _, fact := range expected.Evidence {
			if !strings.Contains(claim, strings.ToLower(fact)) {
				problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_provenance", Row: expected.Kind, Field: "Evidence/claim"})
				break
			}
		}
	}
	if overclaim := phase34OverclaimText(strings.ToLower(block)); overclaim != "" {
		alreadyReported := false
		for _, problem := range problems {
			if problem.Code == "evidence_overclaim" {
				alreadyReported = true
				break
			}
		}
		if !alreadyReported {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "evidence_overclaim", Row: "evidence claim", Field: overclaim})
		}
	}
	if !phase34ReferenceExists("internal/compiler/session/session.go#BuildApplicationCommandFile") ||
		!phase34ReferenceExists("cmd/schway/main.go#runSurface") {
		problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: phase34TransitionCapabilities[0].Key, Field: "Evidence/debt source reference"})
	}
	for _, reference := range []string{
		"internal/compiler/session/session_phase34_routes_test.go#TestPhase34SourceRoutePresentation",
		"internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses",
		"internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden",
	} {
		evidence := capabilityRows[phase34TransitionCapabilities[0].Key]["Evidence/debt"]
		if strings.TrimSpace(evidence) != "" && (!phase34ReferenceExists(reference) || !strings.Contains(evidence, reference)) {
			problems = append(problems, phase34TransitionFinding{Document: document, Code: "missing_capability_field", Row: phase34TransitionCapabilities[0].Key, Field: "Evidence/debt " + reference})
		}
	}
	return problems
}

func phase34OverclaimText(text string) string {
	for _, phrase := range []string{"current hosted receipt", "hosted binding complete", "human readability measured", "whole-loop latency measured", "p50/p95 measured"} {
		if strings.Contains(text, phrase) {
			return phrase
		}
	}
	return ""
}

func phase34ReferenceExists(reference string) bool {
	path, symbol, ok := strings.Cut(reference, "#")
	if !ok {
		return false
	}
	data, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "func "+symbol+"(")
}

func phase34TransitionBlock(text string) (string, bool, bool) {
	if strings.Count(text, phase34TransitionHeading) != 1 {
		return "", false, true
	}
	start := strings.Index(text, phase34TransitionHeading)
	end := len(text)
	rest := text[start+len(phase34TransitionHeading):]
	for _, marker := range []string{"\n## ", "\n# "} {
		if index := strings.Index(rest, marker); index >= 0 && start+len(phase34TransitionHeading)+index < end {
			end = start + len(phase34TransitionHeading) + index + 1
		}
	}
	return text[start:end], true, false
}

func phase34ExtractTransitionBlock(t *testing.T, text string) string {
	t.Helper()
	block, found, duplicate := phase34TransitionBlock(text)
	if !found || duplicate {
		t.Fatalf("Phase 34 transition section missing or duplicated")
	}
	return block
}

func phase34ParseTable(block string, header []string) map[string]map[string]string {
	lines := strings.Split(block, "\n")
	for index, line := range lines {
		cells := phase34SplitTableRow(line)
		if !phase34EqualStrings(cells, header) {
			continue
		}
		rows := make(map[string]map[string]string)
		for rowIndex := index + 1; rowIndex < len(lines); rowIndex++ {
			values := phase34SplitTableRow(lines[rowIndex])
			if len(values) == 0 {
				break
			}
			if len(values) != len(header) {
				continue
			}
			separator := true
			for _, value := range values {
				if strings.Trim(value, "-: ") != "" {
					separator = false
					break
				}
			}
			if separator {
				continue
			}
			row := make(map[string]string, len(header))
			for cellIndex, key := range header {
				row[key] = values[cellIndex]
			}
			rows[row[header[0]]] = row
		}
		return rows
	}
	return map[string]map[string]string{}
}

func phase34SplitTableRow(line string) []string {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") || !strings.HasSuffix(trimmed, "|") {
		return nil
	}
	parts := strings.Split(strings.Trim(trimmed, "|"), "|")
	for index := range parts {
		parts[index] = strings.TrimSpace(parts[index])
	}
	return parts
}

func phase34EqualStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func phase34InferBlankKeyRow(rows map[string]map[string]string, keyField, discriminatorField, expectedDiscriminator string) (map[string]string, bool) {
	for _, row := range rows {
		if row[keyField] == "" && row[discriminatorField] == expectedDiscriminator {
			return row, true
		}
	}
	return nil, false
}

func phase34MutateTransitionCell(t *testing.T, document string, header []string, keyField, key, field, value string) string {
	t.Helper()
	return phase34ChangeTransitionTable(t, document, header, keyField, key, field, value, false)
}

func phase34DeleteTransitionRow(t *testing.T, document string, header []string, keyField, key string) string {
	t.Helper()
	return phase34ChangeTransitionTable(t, document, header, keyField, key, "", "", true)
}

func phase34ChangeTransitionTable(t *testing.T, document string, header []string, keyField, key, field, value string, deleteRow bool) string {
	t.Helper()
	block, start, end := phase34TransitionBlockRange(t, document)
	lines := strings.Split(block, "\n")
	foundHeader := -1
	for index, line := range lines {
		if phase34EqualStrings(phase34SplitTableRow(line), header) {
			foundHeader = index
			break
		}
	}
	if foundHeader < 0 {
		t.Fatalf("transition table header %v missing", header)
	}
	keyIndex, fieldIndex := -1, -1
	for index, name := range header {
		if name == keyField {
			keyIndex = index
		}
		if name == field {
			fieldIndex = index
		}
	}
	if keyIndex < 0 || (!deleteRow && fieldIndex < 0) {
		t.Fatalf("transition mutation names unknown fields key=%q field=%q", keyField, field)
	}
	for index := foundHeader + 1; index < len(lines); index++ {
		cells := phase34SplitTableRow(lines[index])
		if len(cells) == 0 {
			break
		}
		if len(cells) != len(header) || cells[keyIndex] != key {
			continue
		}
		if deleteRow {
			lines = append(lines[:index], lines[index+1:]...)
		} else {
			cells[fieldIndex] = value
			lines[index] = "| " + strings.Join(cells, " | ") + " |"
		}
		mutatedBlock := strings.Join(lines, "\n")
		return document[:start] + mutatedBlock + document[end:]
	}
	t.Fatalf("transition mutation target row %q missing", key)
	return document
}

func phase34TransitionBlockRange(t *testing.T, document string) (string, int, int) {
	t.Helper()
	block, found, duplicate := phase34TransitionBlock(document)
	if !found || duplicate {
		t.Fatal("clean transition block missing or duplicated")
	}
	start := strings.Index(document, phase34TransitionHeading)
	end := start + len(block)
	return block, start, end
}

func phase34ExpectCellFinding(t *testing.T, document, original string, header []string, keyField, key, field, value, code, row, expectedField string) {
	t.Helper()
	mutated := phase34MutateTransitionCell(t, original, header, keyField, key, field, value)
	phase34ExpectSingleFinding(t, document, original, mutated, code, row, expectedField)
}

func phase34ExpectSingleFinding(t *testing.T, document, original, mutated, code, row, field string) {
	t.Helper()
	_, start, end := phase34TransitionBlockRange(t, original)
	mutatedBlock, _, _ := phase34TransitionBlockRange(t, mutated)
	if original[:start] != mutated[:start] || original[end:] != mutated[len(mutated)-len(original[end:]):] {
		t.Fatalf("%s mutation altered historical content outside dated amendment", document)
	}
	problems := phase34TransitionProblems(document, mutated)
	if len(problems) != 1 || problems[0].Code != code || problems[0].Row != row || problems[0].Field != field || problems[0].Document != document {
		t.Fatalf("%s target mutation for %s.%s produced %+v, want only %s", document, row, field, problems, code)
	}
	if mutatedBlock == "" {
		t.Fatal("mutated block unexpectedly empty")
	}
}
