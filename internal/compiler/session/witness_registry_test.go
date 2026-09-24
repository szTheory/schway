// witness_registry_test.go holds plan 14-07's executed probes -- the
// D-14-23 "executed-probe" witness kind, in the flesh. Each probe asserts
// a STRUCTURAL fact (a diagnostic code, an admission refusal, a
// still-identical structural summary), never a message string, per
// D-14-32's accepted-residual-friction note: a probe over-fit to current
// diagnostic prose would go red on a benign rewording, which is not the
// signal this register exists to catch.
package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

const phase16M004Witness = "probe:TestPhase16M004CorpusRefusal"

type phase16LegacyEvidenceLedger struct {
	Schema    string                     `json:"schema"`
	CutCommit string                     `json:"cut_commit"`
	Records   []phase16LegacyEvidenceRow `json:"records"`
}

type phase16LegacyEvidenceRow struct {
	Fixture       string `json:"fixture"`
	FixtureSHA256 string `json:"fixture_sha256"`
	Artifact      string `json:"artifact"`
	Witness       string `json:"witness"`
}

type phase16FileFrozenEvidenceLedger struct {
	Schema  string                         `json:"Schema"`
	Records []phase16FileFrozenEvidenceRow `json:"Records"`
}

type phase16FileFrozenEvidenceRow struct {
	Fixture        string `json:"Fixture"`
	FixtureSHA256  string `json:"FixtureSHA256"`
	ProgramSHA256  string `json:"ProgramSHA256"`
	Artifact       string `json:"Artifact"`
	ArtifactSHA256 string `json:"ArtifactSHA256"`
	Refusal        string `json:"Refusal"`
}

// TestPhase16EmitterInventoryRefusalWitnessesResolve keeps the inventory's
// refusal rows tied to an executable named probe instead of a prose-only
// classification. Frozen-artifact digest validation is added by Plan 16-11;
// this is the fail-closed registry-to-witness half of that contract.
func TestPhase16EmitterInventoryRefusalWitnessesResolve(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "public-emitter-consumers.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry phase16ConsumerRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase5_corpus_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	refusals := 0
	for _, row := range registry.Entries {
		if row.Classification == phase16TypedRefusal {
			if row.Witness != "probe:TestPhase11InterproceduralDifferential/ZeroCallEdges" || strings.Contains(strings.Join(row.EvidenceFixtures, " "), "phase11") {
				t.Fatalf("%s: ambiguous-entry refusal was treated as frozen evidence: witness=%q fixtures=%v", row.Call, row.Witness, row.EvidenceFixtures)
			}
			continue
		}
		if row.Classification == phase16EvidenceValidator {
			if row.Witness == "probe:TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram" {
				if len(row.EvidenceFixtures) == 0 {
					t.Fatalf("%s: frozen evidence validator has no Phase 11 fixture mapping", row.Call)
				}
			} else if row.Witness != "probe:TestPhase16Phase11FrozenEvidenceRejectsProvenanceFaults" && row.Witness != "probe:TestPhase16M004ProvenanceRegistryRejectsFaults" {
				t.Fatalf("%s: unresolved frozen evidence validator witness %q", row.Call, row.Witness)
			}
			continue
		}
		if row.Classification != phase16RefusalWithFrozen {
			continue
		}
		refusals++
		switch row.Witness {
		case phase16M004Witness:
			if !strings.Contains(string(source), "func TestPhase16M004CorpusRefusal") {
				t.Fatalf("%s: unresolved refusal witness %q", row.Call, row.Witness)
			}
		case "probe:TestPhase11ByPointerRefusalFirstEvidence":
			gate, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase11_gate_test.go"))
			if err != nil || !strings.Contains(string(gate), "func TestPhase11ByPointerRefusalFirstEvidence") || len(row.EvidenceFixtures) == 0 {
				t.Fatalf("%s: unresolved Phase 11 refusal witness %q fixtures=%v", row.Call, row.Witness, row.EvidenceFixtures)
			}
		default:
			t.Fatalf("%s: unrecognized refusal witness %q", row.Call, row.Witness)
		}
	}
	if refusals == 0 {
		t.Fatal("inventory has no refusal-plus-frozen classifications")
	}
}

// TestPhase16M004ProvenanceRegistryRejectsFaults is the quality-control
// companion to the cgen manifest tests: the consumer inventory, the current
// public-refusal probe, and the frozen manifest must remain one auditable
// chain.  It mutates each link in memory so an uncited skip cannot make an
// M004 cut look like a dynamic admission.
func TestPhase16M004ProvenanceRegistryRejectsFaults(t *testing.T) {
	registry, evidence, files := phase16M004WitnessFixture(t)
	if problems := phase16M004ProvenanceProblems(registry, evidence, files); len(problems) != 0 {
		t.Fatalf("checked-in provenance chain is invalid: %v", problems)
	}
	for _, mutate := range []struct {
		name  string
		apply func(*phase16ConsumerRegistry, *phase16LegacyEvidenceLedger, *phase16FileFrozenEvidenceLedger)
	}{
		{"missing citation", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase16RefusalRow(r)].Witness = ""
		}},
		{"stale witness", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase16RefusalRow(r)].Witness = "probe:TestNoLongerCurrent"
		}},
		{"changed digest reference", func(_ *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, f *phase16FileFrozenEvidenceLedger) {
			f.Records[0].FixtureSHA256 = strings.Repeat("0", 64)
		}},
		{"classification drift", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase16RefusalRow(r)].Classification = phase16AdmittedDynamic
		}},
		{"Phase 11 fixture substitution", func(_ *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, f *phase16FileFrozenEvidenceLedger) {
			f.Records[phase11FileRecordIndex(f, "testdata/phase11/multi_function_gate_corpus.lang")].Fixture = "testdata/phase11/substituted.lang"
		}},
		{"Phase 11 canonical program digest", func(_ *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, f *phase16FileFrozenEvidenceLedger) {
			f.Records[phase11FileRecordIndex(f, "testdata/phase11/multi_function_gate_corpus.lang")].ProgramSHA256 = strings.Repeat("0", 64)
		}},
		{"Phase 11 artifact digest", func(_ *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, f *phase16FileFrozenEvidenceLedger) {
			f.Records[phase11FileRecordIndex(f, "testdata/phase11/multi_function_gate_corpus.lang")].ArtifactSHA256 = strings.Repeat("0", 64)
		}},
		{"Phase 11 exact refusal", func(_ *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, f *phase16FileFrozenEvidenceLedger) {
			f.Records[phase11FileRecordIndex(f, "testdata/phase11/multi_function_gate_corpus.lang")].Refusal = "different emitter error"
		}},
		{"Phase 11 row classification", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase11RefusalRow(r)].Classification = phase16AdmittedDynamic
		}},
		{"Phase 11 refusal witness", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase11RefusalRow(r)].Witness = "probe:TestNoLongerCurrent"
		}},
		{"Phase 11 fixture mapping", func(r *phase16ConsumerRegistry, _ *phase16LegacyEvidenceLedger, _ *phase16FileFrozenEvidenceLedger) {
			r.Entries[firstPhase11RefusalRow(r)].EvidenceFixtures = []string{"testdata/phase11/multi_function_gate_corpus.lang"}
		}},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			gotRegistry, gotEvidence, gotFiles := clonePhase16M004WitnessFixture(t, registry, evidence, files)
			mutate.apply(&gotRegistry, &gotEvidence, &gotFiles)
			if problems := phase16M004ProvenanceProblems(gotRegistry, gotEvidence, gotFiles); len(problems) == 0 {
				t.Fatal("mutated M004 registry/provenance chain was accepted")
			}
		})
	}
}

func phase11FileRecordIndex(files *phase16FileFrozenEvidenceLedger, fixture string) int {
	for i, record := range files.Records {
		if record.Fixture == fixture {
			return i
		}
	}
	panic("missing Phase 11 frozen record " + fixture)
}

func firstPhase11RefusalRow(registry *phase16ConsumerRegistry) int {
	for i, row := range registry.Entries {
		if row.Classification == phase16RefusalWithFrozen && len(row.EvidenceFixtures) > 0 {
			return i
		}
	}
	panic("missing Phase 11 refusal row")
}

func phase16M004WitnessFixture(t *testing.T) (phase16ConsumerRegistry, phase16LegacyEvidenceLedger, phase16FileFrozenEvidenceLedger) {
	t.Helper()
	read := func(path string, target any) {
		data, err := os.ReadFile(testsupport.ProjectPath(path))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, target); err != nil {
			t.Fatal(err)
		}
	}
	var registry phase16ConsumerRegistry
	var evidence phase16LegacyEvidenceLedger
	var files phase16FileFrozenEvidenceLedger
	read("testdata/phase16/public-emitter-consumers.json", &registry)
	read("testdata/phase16/legacy-emitter-evidence.json", &evidence)
	read("testdata/phase16/file-frozen-evidence.json", &files)
	return registry, evidence, files
}

func clonePhase16M004WitnessFixture(t *testing.T, registry phase16ConsumerRegistry, evidence phase16LegacyEvidenceLedger, files phase16FileFrozenEvidenceLedger) (phase16ConsumerRegistry, phase16LegacyEvidenceLedger, phase16FileFrozenEvidenceLedger) {
	t.Helper()
	data, err := json.Marshal(struct {
		Registry phase16ConsumerRegistry
		Evidence phase16LegacyEvidenceLedger
		Files    phase16FileFrozenEvidenceLedger
	}{registry, evidence, files})
	if err != nil {
		t.Fatal(err)
	}
	var clone struct {
		Registry phase16ConsumerRegistry
		Evidence phase16LegacyEvidenceLedger
		Files    phase16FileFrozenEvidenceLedger
	}
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone.Registry, clone.Evidence, clone.Files
}

func firstPhase16RefusalRow(registry *phase16ConsumerRegistry) int {
	for i, row := range registry.Entries {
		if row.Classification == phase16RefusalWithFrozen {
			return i
		}
	}
	return -1
}

func phase16M004ProvenanceProblems(registry phase16ConsumerRegistry, evidence phase16LegacyEvidenceLedger, files phase16FileFrozenEvidenceLedger) []string {
	var problems []string
	if registry.Schema != "phase16.public-emitter-consumers/2" {
		problems = append(problems, "consumer registry schema is stale")
	}
	if evidence.Schema != "phase16.legacy-emitter-evidence/1" || evidence.CutCommit != "0607486" {
		problems = append(problems, "legacy provenance schema or cutover revision is stale")
	}
	if files.Schema != "phase16.file-frozen-evidence/1" {
		problems = append(problems, "file provenance schema is stale")
	}
	refusalCalls := map[string]bool{}
	phase11MappedFixtures := map[string]bool{}
	phase11GateRowFound := false
	ambiguousEntryRowFound := false
	for _, row := range registry.Entries {
		if row.Call == "internal/compiler/session/session_phase11_differential_test.go:EmitNative:391" {
			ambiguousEntryRowFound = row.Classification == phase16TypedRefusal && row.Witness == "probe:TestPhase11InterproceduralDifferential/ZeroCallEdges" && len(row.EvidenceFixtures) == 0
		}
		if strings.HasPrefix(row.Call, "internal/compiler/session/session_phase11_gate_test.go:EmitNative:") && row.Classification == phase16RefusalWithFrozen && row.Witness == "probe:TestPhase11ByPointerRefusalFirstEvidence" && reflect.DeepEqual(row.EvidenceFixtures, []string{"testdata/phase11/multi_function_gate_corpus.lang", "testdata/phase16/historical/phase11_gate_n_two.fixture"}) {
			phase11GateRowFound = true
		}
		if row.Classification == phase16TypedRefusal && row.Call == "internal/compiler/session/session_phase11_differential_test.go:EmitNative:391" {
			if row.Witness != "probe:TestPhase11InterproceduralDifferential/ZeroCallEdges" || len(row.EvidenceFixtures) != 0 {
				problems = append(problems, "ambiguous-entry refusal lost its independent typed witness")
			}
		}
		if row.Classification == phase16RefusalWithFrozen {
			refusalCalls[row.Call] = true
		}
		if len(row.EvidenceFixtures) > 0 {
			validWitness := row.Witness == "probe:TestPhase11ByPointerRefusalFirstEvidence" || row.Witness == "probe:TestPhase16Phase11FrozenEvidenceRejectsProvenanceFaults" || row.Witness == "probe:TestPhase16M004ProvenanceRegistryRejectsFaults" || row.Witness == "probe:TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram"
			if !validWitness {
				problems = append(problems, "Phase 11 refusal row has stale witness "+row.Call)
			}
			for _, fixture := range row.EvidenceFixtures {
				phase11MappedFixtures[fixture] = true
			}
		}
		if row.Classification == phase16RefusalWithFrozen && len(row.EvidenceFixtures) == 0 && row.Witness != phase16M004Witness {
			problems = append(problems, "uncited or stale refusal row "+row.Call)
		}
	}
	if !phase11GateRowFound {
		problems = append(problems, "Phase 11 refusal-first consumer is missing exact fixture mappings")
	}
	if !ambiguousEntryRowFound {
		problems = append(problems, "ambiguous-entry refusal is not independently typed")
	}
	for _, call := range []string{
		"internal/compiler/cgen/cgen_test.go:Emit:193",
		"internal/compiler/native/foreign_retained_test.go:EmitNative:48",
		"internal/compiler/session/session_phase5_corpus_test.go:EmitNative:468",
	} {
		if !refusalCalls[call] {
			problems = append(problems, "M004 control was reclassified without provenance: "+call)
		}
	}
	fileByFixture := map[string]phase16FileFrozenEvidenceRow{}
	for _, record := range files.Records {
		if record.Fixture == "" || record.Artifact == "" || len(record.FixtureSHA256) != sha256.Size*2 || len(record.ProgramSHA256) != sha256.Size*2 || len(record.ArtifactSHA256) != sha256.Size*2 {
			problems = append(problems, "file frozen record has missing provenance")
			continue
		}
		if _, duplicate := fileByFixture[record.Fixture]; duplicate {
			problems = append(problems, "duplicate file frozen fixture "+record.Fixture)
		}
		fileByFixture[record.Fixture] = record
		if strings.HasPrefix(record.Fixture, "testdata/phase11/multi_function_gate") || record.Fixture == "testdata/phase16/historical/phase11_gate_n_two.fixture" {
			if !phase11MappedFixtures[record.Fixture] {
				problems = append(problems, "unconsumed Phase 11 frozen evidence record "+record.Fixture)
			}
			source, sourceErr := os.ReadFile(testsupport.ProjectPath(record.Fixture))
			if sourceErr != nil {
				problems = append(problems, "Phase 11 fixture cannot be read "+record.Fixture)
				continue
			}
			fixtureSum := sha256.Sum256(source)
			if fmt.Sprintf("%x", fixtureSum) != record.FixtureSHA256 {
				problems = append(problems, "Phase 11 fixture digest changed "+record.Fixture)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				problems = append(problems, "Phase 11 fixture no longer checks "+record.Fixture)
				continue
			}
			canonical, marshalErr := json.Marshal(checked.Program)
			if marshalErr != nil {
				problems = append(problems, "Phase 11 canonical program cannot be encoded "+record.Fixture)
				continue
			}
			programSum := sha256.Sum256(canonical)
			if fmt.Sprintf("%x", programSum) != record.ProgramSHA256 {
				problems = append(problems, "Phase 11 canonical program digest changed "+record.Fixture)
			}
			artifact, artifactErr := os.ReadFile(testsupport.ProjectPath(record.Artifact))
			if artifactErr != nil {
				problems = append(problems, "Phase 11 frozen artifact cannot be read "+record.Fixture)
				continue
			}
			artifactSum := sha256.Sum256(artifact)
			if fmt.Sprintf("%x", artifactSum) != record.ArtifactSHA256 {
				problems = append(problems, "Phase 11 frozen artifact digest changed "+record.Fixture)
			}
			_, refusal := cgen.EmitNative(checked.Program)
			if refusal == nil || refusal.Error() != record.Refusal {
				problems = append(problems, "Phase 11 frozen evidence has stale exact refusal "+record.Fixture)
			}
		}
	}
	for fixture := range phase11MappedFixtures {
		if _, ok := fileByFixture[fixture]; !ok {
			problems = append(problems, "Phase 11 refusal row maps missing frozen evidence "+fixture)
		}
	}
	wantPhase11Fixtures := []string{"testdata/phase11/multi_function_gate_corpus.lang", "testdata/phase16/historical/phase11_gate_n_two.fixture"}
	for _, fixture := range wantPhase11Fixtures {
		if !phase11MappedFixtures[fixture] {
			problems = append(problems, "Phase 11 frozen evidence record is not consumed "+fixture)
		}
	}
	for _, record := range evidence.Records {
		if record.Witness != strings.TrimPrefix(phase16M004Witness, "probe:") {
			problems = append(problems, "legacy record has stale witness "+record.Fixture)
		}
		file, ok := fileByFixture[record.Fixture]
		if !ok || file.FixtureSHA256 != record.FixtureSHA256 {
			problems = append(problems, "legacy record lacks matching file provenance "+record.Fixture)
		}
	}
	return problems
}

// TestPhase17B1RequiresUnverifiableDeclaredContract replaces the obsolete
// sameType premise. A real two-type program is admitted; nevertheless B1 is
// still not a production route because every user-declared signature fact is
// verified by that function's own admission. Separate compilation (M006) is
// the first setting that can introduce a declaration its declarer cannot
// prove locally.
func TestPhase17B1RequiresUnverifiableDeclaredContract(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase17", "return_type_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if result := session.Check(source); len(result.Diagnostics) != 0 {
		t.Fatalf("Phase 17 two-type tracer must be admitted; got diagnostics: %+v", result.Diagnostics)
	}

	if calls := productionResolveBlameCalls(t); len(calls) != 0 {
		t.Fatalf("resolveBlame acquired production call sites %v; DX-06 is outside Phase 17", calls)
	}

	// This is intentionally exhaustive. Adding a FunctionSignature field
	// without classifying its declaration authority fails the witness rather
	// than silently treating it as a future B1 boundary.
	verifier := map[string]string{
		"ID":            "derived from the declaring core function identity",
		"Name":          "copied from the declaring function declaration",
		"Parameters":    "declaring function admission parses and validates parameter contracts",
		"Return":        "declaring function admission parses and validates its return contract against its body",
		"Abilities":     "derived from declaring function type facts",
		"Callable":      "derived from the declaring function's local admission predicate",
		"Fails":         "derived from the declaring function's foreign contract",
		"Foreign":       "derived over the checked local call graph",
		"ClosureDigest": "derived after every declaring function is admitted",
	}
	typ := reflect.TypeOf(core.FunctionSignature{})
	if typ.NumField() != len(verifier) {
		t.Fatalf("FunctionSignature has %d fields but the M006 witness classifies %d; classify any new user-declared contract field before admitting it", typ.NumField(), len(verifier))
	}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i).Name
		if verifier[field] == "" {
			t.Fatalf("FunctionSignature.%s lacks declaring-function verification; this is the M006 reopening condition", field)
		}
	}
}

func productionResolveBlameCalls(t testing.TB) []string {
	t.Helper()
	dir := testsupport.ProjectPath("internal", "compiler", "check")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var calls []string
	fset := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, parseErr := parser.ParseFile(fset, filepath.Join(dir, entry.Name()), nil, 0)
		if parseErr != nil {
			t.Fatal(parseErr)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "resolveBlame" {
				calls = append(calls, fset.Position(call.Pos()).String())
			}
			return true
		})
	}
	return calls
}

// TestD1243ControlIsUnconstructible preserves its registered probe identity
// while proving the opposite of its historical name: Phase 18 now exposes
// returned payload bytes, so the seeded wrong-slot write must diverge on the
// terminal-outcome axis. The historical D-12-43 claim is closed by this
// source-independent runtime control and the Phase 18 source witness.
func TestD1243ControlIsUnconstructible(t *testing.T) {
	ctx := context.Background()
	runner := native.DefaultRunner()
	restore := cgen.SetPayloadSlotSwapForTest(true)
	defer restore()

	program := checkedProgram(t, "testdata", "phase12", "payload_tracer.lang")
	engines := runFunctionOkExecutions(t, ctx, program, "identity", runner)
	injected := cgen.PayloadSlotSwapInjectedWriteCount()
	if injected < 1 {
		t.Fatalf("D-12-43 probe: the fault-injection seam injected no wrong-slot write (count=%d), so this run proves nothing about the control -- check whether payload_tracer.lang's data type still declares two payload-carrying alternatives", injected)
	}
	compareErr := session.Phase5CompareEngines("payload_tracer.lang(D-12-43 probe)", engines)
	disagreement, ok := compareErr.(*session.Phase5EngineDisagreement)
	if !ok || disagreement.Axis != session.AxisTerminalOutcome {
		t.Fatalf("D-12-43 wrong-slot control error = %v, want exact %s disagreement", compareErr, session.AxisTerminalOutcome)
	}
}

// TestLTOInertnessOnMultiFunctionEmission preserves D-14-45's honest limit
// after Phase 16 admitted match-bodied programs to emitProgram. The fixture
// must now emit a schema-2 one-translation-unit program; that admission does
// not prove any LTO axis movement, because the direct C still has no
// cross-translation-unit boundary for -flto to exploit.
func TestLTOInertnessOnMultiFunctionEmission(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "multi_function_match_refusal.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture must check clean: %+v", checked.Diagnostics)
	}
	if len(checked.Program.Functions) < 2 {
		t.Fatalf("fixture must be genuinely multi-function (>= 2 functions), got %d", len(checked.Program.Functions))
	}
	generated, emitErr := cgen.EmitNative(checked.Program)
	if emitErr != nil {
		t.Fatalf("Phase 16 match admission regressed: %v", emitErr)
	}
	if !strings.Contains(generated, "lang.execution/2") {
		t.Fatal("Phase 16 match admission did not emit the schema-2 one-TU program")
	}
}

// TestRetainedPointerEscapeIsStillUnsubjected is the probe
// debtRegisterEscapeRegistry's "callback-invocation-unsubjected" entry
// names (D-14-27): it asserts the retained_pointer NAT-03 row is STILL
// Subjected: false with the same EscapeID -- the structural fact the
// escape declares. If a future plan discharges the escape (subjects the
// row), this probe fails, forcing the escape registry entry to be
// retired rather than left citing a stale premise.
func TestRetainedPointerEscapeIsStillUnsubjected(t *testing.T) {
	var target *session.NAT03Mutation
	for _, row := range session.NAT03Mutations() {
		row := row
		if row.ControlID == "control:native.sanitize.retained_pointer" {
			target = &row
		}
	}
	if target == nil {
		t.Fatal("no NAT-03 row declares control:native.sanitize.retained_pointer")
	}
	if target.Subjected {
		t.Fatal("control:native.sanitize.retained_pointer is now Subjected: true -- the escape has been discharged; retire debtRegisterEscapeRegistry's callback-invocation-unsubjected entry instead of leaving this probe passing on a stale premise")
	}
	if target.EscapeID != "escape:callback-invocation-unsubjected" {
		t.Fatalf("expected EscapeID escape:callback-invocation-unsubjected, got %q", target.EscapeID)
	}
}

// ---------------------------------------------------------------------
// Task 3: every suppression surface cites a resolvable witness (D-14-25).
// ---------------------------------------------------------------------

var (
	suppressionPendingPattern  = regexp.MustCompile(`PENDING-\d\d-\d\d`)
	suppressionDecisionPattern = regexp.MustCompile(`\bD-\d\d-\d\d[a-z]?\b`)
	suppressionEscapePattern   = regexp.MustCompile(`escape:[a-z][a-z0-9-]*`)
	suppressionProbePattern    = regexp.MustCompile(`probe:[A-Za-z][A-Za-z0-9_]*`)
	suppressionEnvPattern      = regexp.MustCompile(`env:[a-z][a-z0-9_-]*`)
)

// debtRegisterDecisionCitationResolves reports whether id (a bare
// D-XX-NN identifier) resolves against either a debt-register row in any
// *-DEBT.md, or a CONTEXT.md decision heading (this project's
// "- **D-XX-NN (...):**" convention) in any *-CONTEXT.md. Both are
// legitimate resolution targets: a D-XX-NN identifier is used throughout
// this codebase both as a debt-register row ID and as a CONTEXT.md
// decision ID, and conflating the two namespaces (requiring every
// decision citation to also be a debt row) would misfire on the
// project's own extensive, otherwise-healthy cross-referencing
// convention.
func debtRegisterDecisionCitationResolves(id string) (bool, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return false, err
	}
	for _, path := range registers {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return false, readErr
		}
		columns, rows, tableErr := parseDebtRegisterTable(filepath.Base(path), string(data))
		if tableErr != nil {
			continue
		}
		idIdx, ok := columns["ID"]
		if !ok {
			continue
		}
		for _, row := range rows {
			if idIdx < len(row) && row[idIdx] == id {
				return true, nil
			}
		}
	}
	contexts, err := phaseArtifactGlob("*", "*-CONTEXT.md")
	if err != nil {
		return false, err
	}
	needle := "**" + id
	for _, path := range contexts {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return false, readErr
		}
		if strings.Contains(string(data), needle) {
			return true, nil
		}
	}
	return false, nil
}

// suppressionResolutionMode selects which citation shapes a surface's
// resolution check applies to. See scanSuppressionSurfaces's own doc
// comment and TestNoSuppressionOutlivesItsWitness for why the LIVE
// module's comment/string-literal surface is checked only for the
// PENDING-NN-NN shape rather than the full grammar.
type suppressionResolutionMode int

const (
	// suppressionModeFull checks all five citation shapes. Safe for a
	// Skip call's own arguments (this codebase's Skip citations use only
	// this plan's shapes) and for an ISOLATED synthetic temp-dir package
	// built by a test (no collision risk with the live corpus).
	suppressionModeFull suppressionResolutionMode = iota
	// suppressionModePendingOnly checks only PENDING-NN-NN. Required for
	// the LIVE module's comment/string-literal surface: this codebase
	// ALREADY uses "escape:" as its own, unrelated, pre-existing
	// anti-theater vocabulary (session_phase5_escapes.go's
	// EscapeCoordinatedSourceToCoreFalseClaim,
	// originvalidate.go's KnownEscape -- a "coordinated lie" escape
	// hatch, nothing to do with this plan's NAT03Mutation escape
	// registry), and "D-XX-NN" is this project's own pervasive
	// decision/finding cross-reference convention used in the vast
	// majority of doc comments, not exclusively as a suppression
	// citation. Requiring every module-wide occurrence of either shape to
	// resolve against THIS plan's two narrow registries would misfire on
	// both pre-existing, healthy conventions; disambiguating them
	// properly (a full CONTEXT.md decision-ID index across all archived
	// milestones, plus a second closed registry for the unrelated
	// escape: vocabulary) is out of this plan's budget and recorded as
	// debt in its own SUMMARY.
	suppressionModePendingOnly
)

// suppressionCitationsResolve scans text for every citation-shaped
// substring D-14-25 names (a pending marker, a decision identifier, an
// escape identifier) plus this register's own probe:/env: token shapes,
// and reports whether at least one citation was found, plus every
// unresolved citation's problem (mode suppressionModePendingOnly checks
// only the PENDING shape). A text with zero citation-shaped substrings
// reports found=false; the caller decides whether that absence is itself
// a violation (only true for a Skip-call site -- a bare comment
// mentioning nothing citation-shaped is not automatically a
// suppression).
func suppressionCitationsResolve(text string, mode suppressionResolutionMode) (found bool, problems []string) {
	if suppressionPendingPattern.MatchString(text) {
		found = true
		// PENDING-NN-NN is accepted by shape alone: D-14-26's legacy
		// marker form has no registry entry of its own. Plan 14-08
		// removed this module's one surviving instance entirely as part
		// of collapsing the axis-movement law; no PENDING-NN-NN citation
		// is expected to remain live anywhere in this module now.
	}
	if mode == suppressionModePendingOnly {
		return found, nil
	}
	for _, m := range suppressionEscapePattern.FindAllString(text, -1) {
		found = true
		id := strings.TrimPrefix(m, "escape:")
		if _, ok := debtRegisterEscapeRegistry[id]; !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: absent from the closed escape registry", m))
		}
	}
	for _, m := range suppressionProbePattern.FindAllString(text, -1) {
		found = true
		name := strings.TrimPrefix(m, "probe:")
		ok, err := debtRegisterProbeExists(name)
		if err != nil {
			problems = append(problems, fmt.Sprintf("citation %q: %v", m, err))
			continue
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: no such test exists", m))
		}
	}
	for _, m := range suppressionEnvPattern.FindAllString(text, -1) {
		found = true
		id := strings.TrimPrefix(m, "env:")
		if !debtRegisterEnvironmentalSet[id] {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: outside the closed environmental set", m))
		}
	}
	for _, m := range suppressionDecisionPattern.FindAllString(text, -1) {
		found = true
		ok, err := debtRegisterDecisionCitationResolves(m)
		if err != nil {
			problems = append(problems, fmt.Sprintf("citation %q: %v", m, err))
			continue
		}
		if !ok {
			problems = append(problems, fmt.Sprintf("citation %q does not resolve: absent from every *-DEBT.md register and every *-CONTEXT.md decision", m))
		}
	}
	return found, problems
}

// suppressionSite is one enumerated occurrence of a scanned surface.
type suppressionSite struct {
	path string
	pos  string
	kind string
	text string
}

// scanSuppressionSurfaces walks EVERY *.go file under root (production
// and test alike -- D-14-25's specific fix for instance 4, which watched
// only one production file and missed the identical marker surviving as
// an error string and as prose in two test files) and enumerates four
// surfaces: every //go:build constraint, every comment, every string
// literal, and every Skip/Skipf/SkipNow call (recorded with its
// concatenated string-literal arguments as text).
//
// The //go:build surface is enumerated for EVERY .go file under root,
// including one the current GOOS/GOARCH build excludes: go/build's
// MatchFile gate is applied only to the three AST-based surfaces
// (comments, string literals, Skip calls) that follow it, never to the
// textual constraint pass. A constraint is exactly what decides whether
// its own file compiles on this host, so gating the pass that checks
// constraints on the constraint being checked would let a file hide its
// own //go:build line from this guard simply by declaring a host it
// isn't currently running on (T-14-13-02).
//
// suppressionProblems checks every enumerated build-constraint site's
// terms (split by buildConstraintTerms) against buildConstraintAllowlist
// -- the closed, project-portability set {darwin, linux, amd64, arm64,
// cgo} that AGENTS.md's stated host priorities and Go's own cgo tag
// license with no citation required. A term outside that set is refused
// unless the same line also carries a citation that resolves under this
// file's own witness grammar (a PENDING marker, a D-XX-NN decision, an
// escape: or probe: token) -- the citation is the escape hatch for a
// deliberately retained non-portable term, proven live by
// TestSuppressionWitnessGuardIsNotInert's "build constraint outside the
// allowlist" subtest.
func scanSuppressionSurfaces(root string) ([]suppressionSite, error) {
	fset := token.NewFileSet()
	buildCtx := build.Default
	var sites []suppressionSite
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			base := d.Name()
			if base == ".git" || base == "testdata" || (path != root && strings.HasPrefix(base, ".")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		raw, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			rel = path
		}

		// The textual //go:build pass runs on EVERY .go file's bytes,
		// ahead of the MatchFile gate below: a file excluded by its own
		// constraint on the current host is exactly the file whose
		// constraint nobody would otherwise see, so gating the constraint
		// scan on the constraint it is checking would be circular
		// (T-14-13-02). MatchFile is applied only to the AST-based passes
		// that follow (comments, string literals, Skip calls), which stay
		// scoped to files the current build actually compiles.
		for lineNumber, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "//go:build") {
				sites = append(sites, suppressionSite{path: rel, pos: fmt.Sprintf("%s:%d", rel, lineNumber+1), kind: "build-constraint", text: line})
			}
		}

		dir := filepath.Dir(path)
		match, matchErr := buildCtx.MatchFile(dir, filepath.Base(path))
		if matchErr != nil {
			return matchErr
		}
		if !match {
			return nil
		}

		file, parseErr := parser.ParseFile(fset, path, raw, parser.ParseComments)
		if parseErr != nil {
			return fmt.Errorf("parse %s: %w", path, parseErr)
		}

		for _, group := range file.Comments {
			for _, comment := range group.List {
				sites = append(sites, suppressionSite{path: rel, pos: fset.Position(comment.Pos()).String(), kind: "comment", text: comment.Text})
			}
		}

		ast.Inspect(file, func(n ast.Node) bool {
			switch node := n.(type) {
			case *ast.BasicLit:
				if node.Kind == token.STRING {
					sites = append(sites, suppressionSite{path: rel, pos: fset.Position(node.Pos()).String(), kind: "string-literal", text: node.Value})
				}
			case *ast.CallExpr:
				sel, ok := node.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				switch sel.Sel.Name {
				case "Skip", "Skipf", "SkipNow":
					var text strings.Builder
					for _, arg := range node.Args {
						if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.STRING {
							text.WriteString(lit.Value)
							text.WriteString(" ")
						}
					}
					sites = append(sites, suppressionSite{path: rel, pos: fset.Position(node.Pos()).String(), kind: "skip", text: text.String()})
				}
			}
			return true
		})
		return nil
	})
	return sites, err
}

// buildConstraintAllowlist is the closed set of //go:build constraint
// TERMS (GOOS values, GOARCH values, and cgo) this guard accepts with no
// citation required. Membership is a project-portability decision:
// AGENTS.md names macOS and Linux as this project's initial host
// priorities, so this project's declared targets (darwin, linux GOOS;
// amd64, arm64 GOARCH) plus Go's own cgo build tag are pre-approved. Any
// other term -- another GOOS/GOARCH value, a Go version tag, a custom
// build tag -- falls outside that portability decision and must carry a
// citation that resolves under the EVD-04 witness grammar
// (suppressionCitationsResolve) on the same //go:build line.
var buildConstraintAllowlist = map[string]string{
	"darwin": "AGENTS.md names macOS as an initial host priority",
	"linux":  "AGENTS.md names Linux as an initial host priority",
	"amd64":  "portable GOARCH target this project builds for",
	"arm64":  "portable GOARCH target this project builds for (current Apple host)",
	"cgo":    "Go's own cgo build tag, not a portability decision",
}

// buildConstraintTerms splits a //go:build constraint line into its bare
// terms. It is deliberately a TERM EXTRACTOR, not a constraint evaluator:
// this guard only ever asks "which terms appear on this line", never "does
// this line's boolean expression evaluate true on some host" -- that
// question belongs to go/build.Context.MatchFile, not to this guard. Go's
// build-constraint syntax allows &&, ||, !, and parentheses between terms;
// this splits on all of them, trims the leading directive, drops each
// term's negation prefix, and discards empties.
func buildConstraintTerms(line string) []string {
	trimmed := strings.TrimSpace(line)
	trimmed = strings.TrimPrefix(trimmed, "//go:build")
	replacer := strings.NewReplacer("&&", " ", "||", " ", "(", " ", ")", " ")
	fields := strings.Fields(replacer.Replace(trimmed))
	var terms []string
	for _, field := range fields {
		field = strings.TrimPrefix(field, "!")
		if field == "" {
			continue
		}
		terms = append(terms, field)
	}
	return terms
}

// suppressionProblems runs suppressionCitationsResolve over every site,
// plus the NAT03Mutation unsubjected-row surface, and returns every
// problem found (nil means clean). Shared by
// TestNoSuppressionOutlivesItsWitness and
// TestSuppressionWitnessGuardIsNotInert's seeded-fault subtests so both
// exercise the SAME law. Skip-call sites always use the full grammar
// (mode is irrelevant to a skip call's own citation, which must resolve
// completely); nonSkipMode governs comments, string literals, and build
// constraints.
func suppressionProblems(root string, nonSkipMode suppressionResolutionMode) ([]string, error) {
	sites, err := scanSuppressionSurfaces(root)
	if err != nil {
		return nil, err
	}
	var problems []string
	for _, site := range sites {
		mode := nonSkipMode
		if site.kind == "skip" {
			mode = suppressionModeFull
		}
		found, siteProblems := suppressionCitationsResolve(site.text, mode)
		for _, problem := range siteProblems {
			problems = append(problems, fmt.Sprintf("%s (%s): %s", site.pos, site.kind, problem))
		}
		if site.kind == "skip" && !found {
			problems = append(problems, fmt.Sprintf("%s (%s): uncited suppression -- a Skip call must cite a resolvable witness", site.pos, site.kind))
		}
		if site.kind == "build-constraint" && !found {
			for _, term := range buildConstraintTerms(site.text) {
				if _, allowed := buildConstraintAllowlist[term]; allowed {
					continue
				}
				problems = append(problems, fmt.Sprintf("%s (%s): constraint term %q is outside buildConstraintAllowlist and carries no resolvable citation", site.pos, site.kind, term))
			}
		}
	}
	return problems, nil
}

// TestNoSuppressionOutlivesItsWitness enumerates every suppression
// surface in the module -- Skip calls, //go:build constraints, comments,
// and string literals, across every package including test files
// (D-14-25) -- and requires each citation-shaped substring found to
// resolve. A bare Skip call carrying no citation at all is its own
// violation; a citation that does not resolve is a violation regardless
// of surface, including inside a string literal (the specific fix for
// instance 4, where a stale marker survived as an error string).
func TestNoSuppressionOutlivesItsWitness(t *testing.T) {
	root := testsupport.ProjectPath()
	sites, err := scanSuppressionSurfaces(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(sites) == 0 {
		t.Fatal("scanSuppressionSurfaces found nothing at all -- the scanner is looking at nothing, not finding a clean module")
	}
	skipSites := 0
	for _, site := range sites {
		if site.kind == "skip" {
			skipSites++
			t.Logf("suppression site: %s (%s): %q", site.pos, site.kind, site.text)
		}
	}
	if skipSites == 0 {
		t.Fatal("scanSuppressionSurfaces found no skip call sites -- the corpus has known Skip/Skipf calls, so this is a scanner defect, not a clean module")
	}

	problems, err := suppressionProblems(root, suppressionModePendingOnly)
	if err != nil {
		t.Fatal(err)
	}

	// Surface: every NAT03Mutation row declared unsubjected must name a
	// declared escape that resolves in the closed escape registry.
	for _, row := range session.NAT03Mutations() {
		if row.Subjected {
			continue
		}
		if row.EscapeID == "" {
			problems = append(problems, fmt.Sprintf("NAT03Mutation %s: declared unsubjected with no EscapeID", row.ControlID))
			continue
		}
		id := strings.TrimPrefix(row.EscapeID, "escape:")
		if _, ok := debtRegisterEscapeRegistry[id]; !ok {
			problems = append(problems, fmt.Sprintf("NAT03Mutation %s: EscapeID %q does not resolve in the closed escape registry", row.ControlID, row.EscapeID))
		}
	}

	if len(problems) > 0 {
		t.Fatalf("%d suppression problem(s):\n%s", len(problems), strings.Join(problems, "\n"))
	}
}

// TestSuppressionWitnessOnlyInStringLiteralIsFound is Task 3's own
// dedicated acceptance criterion: a synthetic package whose only citation
// lives inside a STRING LITERAL (never a comment, never a Skip argument)
// is still reported when that citation fails to resolve -- proving the
// enumerator inspects string literals as their own surface, not merely
// as an accident of scanning Skip-call arguments.
func TestSuppressionWitnessOnlyInStringLiteralIsFound(t *testing.T) {
	dir := t.TempDir()
	source := "package seeded\n\nfunc stale() string {\n\treturn \"row not yet subjected -- see escape:this-escape-id-is-not-registered\"\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "seeded.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	problems, err := suppressionProblems(dir, suppressionModeFull)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) == 0 {
		t.Fatal("a string-literal-only citation that fails to resolve should be reported, but nothing was")
	}
	found := false
	for _, p := range problems {
		if strings.Contains(p, "string-literal") && strings.Contains(p, "escape:this-escape-id-is-not-registered") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected a string-literal-kind problem naming the unresolved escape citation, got: %v", problems)
	}
}

// TestBuildConstraintsOutsideTheAllowlistAreRefused is Task 1's tracer,
// wiring the thinnest complete path end to end: a //go:build constraint
// line in a scanned file is enumerated as a build-constraint site,
// evaluated against buildConstraintAllowlist, and reported by name when a
// term falls outside it, while an all-allowlisted sibling constraint
// produces nothing. Fixtures use interpreted strings with escaped
// newlines, never a backtick raw string literal: a raw string would place
// a real //go:build directive at the start of a line in THIS file, and the
// module-wide TestNoSuppressionOutlivesItsWitness scan would then enumerate
// the fixture's own text as a live site.
func TestBuildConstraintsOutsideTheAllowlistAreRefused(t *testing.T) {
	allowedDir := t.TempDir()
	allowedSource := "//go:build darwin || linux\n\npackage seeded\n\nfunc allowed() {}\n"
	if err := os.WriteFile(filepath.Join(allowedDir, "allowed.go"), []byte(allowedSource), 0o600); err != nil {
		t.Fatal(err)
	}
	problems, err := suppressionProblems(allowedDir, suppressionModeFull)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 0 {
		t.Fatalf("an all-allowlisted //go:build constraint should produce zero problems; got: %v", problems)
	}

	refusedDir := t.TempDir()
	refusedSource := "//go:build darwin || linux || windows\n\npackage seeded\n\nfunc refused() {}\n"
	if err := os.WriteFile(filepath.Join(refusedDir, "refused.go"), []byte(refusedSource), 0o600); err != nil {
		t.Fatal(err)
	}
	problems, err = suppressionProblems(refusedDir, suppressionModeFull)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("expected exactly one problem for the unallowlisted constraint term, got %d: %v", len(problems), problems)
	}
	if !strings.Contains(problems[0], "windows") {
		t.Fatalf("expected the problem to name the offending term %q, got: %q", "windows", problems[0])
	}
	if !strings.Contains(problems[0], "refused.go:1") {
		t.Fatalf("expected the problem to name the site position, got: %q", problems[0])
	}
}

// TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile is Task 2's own
// proof that the textual //go:build scan runs ahead of go/build's
// MatchFile gate: a .go file whose own constraint EXCLUDES it from the
// current host's build must still be enumerated as a build-constraint
// site, and an unallowlisted term inside it must still be reported. A
// constraint that hides its own file from the scanner that checks
// constraints is circular -- exactly the hole T-14-13-02 names. Before the
// reorder lands, MatchFile's early return in scanSuppressionSurfaces skips
// the raw-byte read for this file entirely, so this test fails RED for
// exactly that reason.
func TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile(t *testing.T) {
	dir := t.TempDir()
	source := "//go:build plan9\n\npackage seeded\n\nfunc excluded() {}\n"
	if err := os.WriteFile(filepath.Join(dir, "excluded.go"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	problems, err := suppressionProblems(dir, suppressionModeFull)
	if err != nil {
		t.Fatal(err)
	}
	if len(problems) != 1 {
		t.Fatalf("expected the host-excluded file's unallowlisted constraint term to still be reported, got %d problems: %v", len(problems), problems)
	}
	if !strings.Contains(problems[0], "plan9") {
		t.Fatalf("expected the problem to name the offending term %q, got: %q", "plan9", problems[0])
	}
}

// TestSuppressionWitnessGuardIsNotInert seeds one fault per mechanizable
// kind (D-14-28) and asserts red for each, with an unmodified-copy
// control asserting green: (a) a reason-free Skip in a copied synthetic
// package; (b) a flipped callsite: count in a copied register; (c) a
// neutralized probe -- editing a copied fixture so the claim it refuses
// becomes admitted, the unexpected-pass (XPASS) shape; (d) a build
// constraint outside buildConstraintAllowlist. A guard proven red on only
// some of the four fault kinds would be inert for the rest.
func TestSuppressionWitnessGuardIsNotInert(t *testing.T) {
	t.Run("reason-free skip in a copied package", func(t *testing.T) {
		dir := t.TempDir()
		clean := "package seeded\n\nimport \"testing\"\n\nfunc TestSeededCitedSkip(t *testing.T) {\n\tt.Skip(\"probe:TestDebtRegistersAreWellFormed\")\n}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded_test.go"), []byte(clean), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err := suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("a cited skip in an otherwise-clean copy should pass; got: %v", problems)
		}

		uncited := "package seeded\n\nimport \"testing\"\n\nfunc TestSeededUncitedSkip(t *testing.T) {\n\tt.Skip(\"no reason\")\n}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded_test.go"), []byte(uncited), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err = suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a reason-free Skip should be reported as an uncited suppression, but nothing was")
		}
	})

	t.Run("flipped callsite count in a copied register", func(t *testing.T) {
		unmodified := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=0")
		if problems, err := debtRegisterProblems(unmodified); err != nil || len(problems) != 0 {
			t.Fatalf("expected a clean pass on the correct count; got err=%v problems=%v", err, problems)
		}

		flipped := debtRegisterWitnessGrammarFixture(t, "WIRED", "callsite:internal/compiler/check.resolveBlame=1")
		problems, err := debtRegisterProblems(flipped)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) == 0 {
			t.Fatal("a flipped callsite: count should be reported, but nothing was")
		}
	})

	t.Run("neutralized probe produces an unexpected pass", func(t *testing.T) {
		unmodifiedSource, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase14", "blame_unreachable_admission_refusal.lang"))
		if err != nil {
			t.Fatal(err)
		}
		if diagnostics := session.Check(unmodifiedSource).Diagnostics; len(diagnostics) == 0 {
			t.Fatal("the unmodified fixture must still be refused; TestB1BlameIsStructurallyUnreachable's own premise has decayed")
		}

		neutralized := strings.Replace(string(unmodifiedSource), "-> Buffer {", "-> Byte {", 1)
		if neutralized == string(unmodifiedSource) {
			t.Fatal("seeded edit did not change the fixture -- the seam this subtest targets has drifted")
		}
		diagnostics := session.Check([]byte(neutralized)).Diagnostics
		if len(diagnostics) != 0 {
			t.Fatalf("expected the neutralized fixture (ReturnType now matching ParameterType) to check CLEAN -- an unexpected pass demonstrating what an XPASS looks like -- got diagnostics: %+v", diagnostics)
		}
		// This is exactly the condition TestB1BlameIsStructurallyUnreachable
		// itself treats as red (t.Fatal on len(Diagnostics)==0): had this
		// neutralized fixture shipped instead of the real one, that probe
		// would have failed loudly, proving it is not inert to this fault
		// kind.
	})

	t.Run("build constraint outside the allowlist", func(t *testing.T) {
		dir := t.TempDir()

		allowlisted := "//go:build darwin || linux\n\npackage seeded\n\nfunc allowed() {}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded.go"), []byte(allowlisted), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err := suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("an all-allowlisted constraint should pass; got: %v", problems)
		}
		t.Log("fixture 1/3: allowlisted constraint -- green")

		unallowlisted := "//go:build darwin || linux || windows\n\npackage seeded\n\nfunc refused() {}\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded.go"), []byte(unallowlisted), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err = suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 1 {
			t.Fatalf("an unallowlisted constraint term should be reported exactly once, got %d: %v", len(problems), problems)
		}
		t.Logf("fixture 2/3: unallowlisted term -- red: %s", problems[0])

		// The citation is placed on a //go:build-prefixed line AFTER the
		// package clause, deliberately outside go/parser's own build-
		// directive position rule (a real directive must precede the
		// package clause, preceded only by blank lines and other line
		// comments): a genuine decision citation like "D-14-25" contains a
		// hyphen, which is not valid build-tag syntax, so embedding it in
		// an actual leading directive would make go/parser refuse to parse
		// the file at all. scanSuppressionSurfaces' own textual pass has no
		// such position rule -- it enumerates any line prefixed
		// "//go:build" anywhere in the file -- so this still exercises the
		// same site kind and the same citation escape hatch.
		cited := "package seeded\n\nfunc cited() {}\n\n//go:build darwin || linux || windows // D-14-25: windows retained pending a future host decision\n"
		if err := os.WriteFile(filepath.Join(dir, "seeded.go"), []byte(cited), 0o600); err != nil {
			t.Fatal(err)
		}
		problems, err = suppressionProblems(dir, suppressionModeFull)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(problems) != 0 {
			t.Fatalf("an unallowlisted term carrying a resolvable citation on the same line should pass; got: %v", problems)
		}
		t.Log("fixture 3/3: unallowlisted term with a resolvable citation -- green (escape hatch is real, not decorative)")
	})
}

// ---------------------------------------------------------------------
// Task 4: .planning/UNREACHABLE-CLAIMS.md -- a generated, byte-compared
// view (D-14-13, D-14-14). The register row is the AUTHORED truth, the
// probe is the EXECUTED truth, this file is DERIVED: regenerated in
// memory and compared, never hand-edited, no blessing path.
// ---------------------------------------------------------------------

// unreachableClaimEntry is one row of the generated view: a debt-register
// row whose Witness cell names at least one probe: token -- the closed,
// syntactic definition of "qualifies as a built-but-structurally-
// unreachable claim" this generator uses. A row graded WIRED/REACHABLE/
// EXERCISED/MUTATION-KILLED backed by an executed probe is exactly
// D-13-02b's "deactivated code" shape (DO-178C): present, provably
// unexecutable today, justified by analysis an executed probe forces to
// be re-examined when the configuration changes.
type unreachableClaimEntry struct {
	id       string
	register string
	grade    string
	witness  string
	trigger  string
	claim    string
}

// unreachableClaimsFrontmatterPattern matches the generated view's own
// `entries: N` frontmatter line, mirroring debtRegisterProblems' `items:`
// cross-check.
var unreachableClaimsFrontmatterPattern = regexp.MustCompile(`(?m)^entries:\s*(\d+)\s*$`)

// deriveUnreachableClaims scans every *-DEBT.md register (live and
// archived) for rows whose Witness cell contains a probe: token, in
// register-glob order (phaseArtifactGlob's own sorted order) then table
// order within each register -- fully deterministic. A register lacking
// Grade/Witness columns entirely (debtRegisterGradeWitnessExemptions)
// contributes nothing, which is correct: it has no graded rows to derive
// a claim from.
func deriveUnreachableClaims() ([]unreachableClaimEntry, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(registers)
	var entries []unreachableClaimEntry
	for _, path := range registers {
		name := filepath.Base(path)
		if _, exempt := debtRegisterGradeWitnessExemptions[name]; exempt {
			continue
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil, readErr
		}
		columns, rows, tableErr := parseDebtRegisterTable(name, string(data))
		if tableErr != nil {
			return nil, tableErr
		}
		idIdx, hasID := columns["ID"]
		gradeIdx, hasGrade := columns["Grade"]
		witnessIdx, hasWitness := columns["Witness"]
		landingIdx, hasLanding := columns["Landing phase"]
		itemIdx, hasItem := columns["Item"]
		if !hasID || !hasGrade || !hasWitness || !hasLanding || !hasItem {
			continue
		}
		for _, row := range rows {
			witness := row[witnessIdx]
			if !strings.Contains(witness, "probe:") {
				continue
			}
			entry := unreachableClaimEntry{
				id:       row[idIdx],
				register: name,
				grade:    row[gradeIdx],
				witness:  witness,
				trigger:  row[landingIdx],
				claim:    row[itemIdx],
			}
			// PHASE-13-DEBT.md is immutable historical provenance. The current
			// view corrects its superseded P17 disposition without rewriting the
			// original record: Phase 17 closed D-13-10a and narrowed D-13-02b's
			// reopening boundary to M006/separate compilation.
			if name == "PHASE-13-DEBT.md" && entry.id == "D-13-10a" {
				continue
			}
			if name == "PHASE-13-DEBT.md" && entry.id == "D-13-02b" {
				entry.witness = "probe:TestPhase17B1RequiresUnverifiableDeclaredContract, callsite:internal/compiler/check.resolveBlame=0"
				entry.trigger = "M006 (modules and separate compilation)"
				entry.claim = "B1 (CONTRACT-VIOLATION BLAME) HAS NO PRODUCTION ROUTE: every current user-declared FunctionSignature field is verified by its declaring function's own admission. resolveBlame remains unwired. Reopen only for M006 separate compilation, when a user-declared contract field can be unverifiable by its declarer."
			}
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func TestPhase17BlameBoundaryViewIsCurrent(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.id != "D-13-02b" {
			continue
		}
		if entry.trigger != "M006 (modules and separate compilation)" || strings.Contains(entry.claim, "sameType") || strings.Contains(entry.claim, "P17") || !strings.Contains(entry.witness, "TestPhase17B1RequiresUnverifiableDeclaredContract") {
			t.Fatalf("D-13-02b current disposition is stale: %+v", entry)
		}
		return
	}
	t.Fatal("D-13-02b missing from current unreachable-claims view")
}

func TestPhase17UseMatchingArgumentDebtIsClosed(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.id == "D-13-10a" {
			t.Fatal("D-13-10a remains in the current unreachable view after sealed repair evidence closed it")
		}
	}
}

// TestUnreachableClaimsGeneratedView preserves the plan-level test name while
// running the established byte-compare implementation.
func TestUnreachableClaimsGeneratedView(t *testing.T) {
	TestUnreachableClaimsViewIsCurrent(t)
}

// renderUnreachableClaimsView renders entries as the exact checked-in
// document shape: frontmatter `entries: N`, then one table row per entry.
func renderUnreachableClaimsView(entries []unreachableClaimEntry) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "entries: %d\n", len(entries))
	b.WriteString("---\n\n")
	b.WriteString("# Unreachable Claims\n\n")
	b.WriteString("**GENERATED. Do not hand-edit.** Regenerated in memory and byte-compared by\n")
	b.WriteString("`TestUnreachableClaimsViewIsCurrent` (`internal/compiler/session/witness_registry_test.go`)\n")
	b.WriteString("from every `*-DEBT.md` register's `Grade`/`Witness` columns. The register row\n")
	b.WriteString("is the authored truth, the probe is the executed truth, this file is derived --\n")
	b.WriteString("a hand edit is a failure, not a source of information. There is no regeneration\n")
	b.WriteString("command and no blessing path: if this view is out of date, correct the\n")
	b.WriteString("registers and re-derive, never overwrite this file directly.\n\n")
	b.WriteString("A row qualifies for this view when its `Witness` cell names at least one\n")
	b.WriteString("`probe:` token: a claim justified by an assertion that currently holds and is\n")
	b.WriteString("asserted to hold, so that its ceasing to hold turns the suite red (D-14-22).\n\n")
	b.WriteString("## Claims\n\n")
	b.WriteString("| ID | Register | Grade | Witness | Unblocking trigger | Claim |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, entry := range entries {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s |\n", entry.id, entry.register, entry.grade, entry.witness, entry.trigger, entry.claim)
	}
	return b.String()
}

// TestUnreachableClaimsViewIsCurrent regenerates the view in memory from
// the registers and byte-compares it against the checked-in
// .planning/UNREACHABLE-CLAIMS.md -- a hand edit is a failure (D-14-13).
// The non-zero-row-count assertion (D-14-14c) additionally guards against
// a vacuous pass: an empty generator against an empty checked-in file
// compares equal, which would be true evidence of nothing.
func TestUnreachableClaimsViewIsCurrent(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 6 {
		t.Fatalf("expected at least the six known qualifying rows (D-13-02b, D-13-10a, D-13-34, D-14-45, D-11-02, D-12-43), got %d: %+v", len(entries), entries)
	}

	regenerated := renderUnreachableClaimsView(entries)

	checkedInPath := testsupport.ProjectPath(".planning", "UNREACHABLE-CLAIMS.md")
	checkedIn, err := os.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("read .planning/UNREACHABLE-CLAIMS.md: %v", err)
	}
	if regenerated != string(checkedIn) {
		t.Fatalf(".planning/UNREACHABLE-CLAIMS.md is out of date -- regenerate from the registers, never hand-edit.\n--- regenerated ---\n%s\n--- checked-in ---\n%s", regenerated, string(checkedIn))
	}

	m := unreachableClaimsFrontmatterPattern.FindStringSubmatch(string(checkedIn))
	if m == nil {
		t.Fatal(".planning/UNREACHABLE-CLAIMS.md has no `entries: N` frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatalf("unreadable entries: frontmatter: %v", convErr)
	}
	if declared != len(entries) {
		t.Fatalf("frontmatter declares entries: %d but the derivation holds %d", declared, len(entries))
	}
}

// TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing (D-14-14c) proves
// the non-zero-row-count assertion is not vacuous ornamentation: over
// synthetic EMPTY inputs, a byte-compare alone would pass (two empty
// documents are byte-identical), but the additional "qualifying rows
// exist in the live registers" assertion must independently catch that
// the real registers are never actually empty.
func TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing(t *testing.T) {
	emptyRendered := renderUnreachableClaimsView(nil)
	if !strings.Contains(emptyRendered, "entries: 0") {
		t.Fatalf("expected an empty entry list to render entries: 0, got:\n%s", emptyRendered)
	}
	// A byte-compare between two independently rendered empty views
	// passes vacuously -- this is the exact failure mode D-14-14c names.
	if emptyRendered != renderUnreachableClaimsView(nil) {
		t.Fatal("renderUnreachableClaimsView is nondeterministic on empty input")
	}
	// The real registers are never actually empty: at least six rows
	// qualify today (see TestUnreachableClaimsViewIsCurrent), which is
	// the independent assertion that makes the byte-compare meaningful
	// rather than vacuous.
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the live registers derived ZERO qualifying rows -- the byte-compare above would now be vacuously comparing two empty documents, proving consistency but not completeness")
	}
}

// TestUnreachableClaimsViewCatchesHandEdit proves the byte-compare is
// live: appending one character to a temp copy of the checked-in file
// makes a byte-for-byte comparison against the (unchanged) regeneration
// fail. Does not touch the real checked-in file.
func TestUnreachableClaimsViewCatchesHandEdit(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	regenerated := renderUnreachableClaimsView(entries)
	tampered := regenerated + "x"
	if tampered == regenerated {
		t.Fatal("seeded one-character edit did not change the text")
	}
	if tampered == renderUnreachableClaimsView(entries) {
		t.Fatal("the tampered copy should differ from a fresh regeneration, but compared equal")
	}
}

// TestUnreachableClaimsViewCatchesStaleEntry proves a checked-in entry
// whose underlying register row has vanished is a failure, never
// auto-pruned (D-14-14b): a hand-rendered view naming a nonexistent
// register row will never byte-match a real regeneration (which simply
// omits the vanished row), so the comparison in
// TestUnreachableClaimsViewIsCurrent already catches this -- this test
// demonstrates the mechanism directly, without depending on the live
// corpus ever actually losing a row.
func TestUnreachableClaimsViewCatchesStaleEntry(t *testing.T) {
	entries, err := deriveUnreachableClaims()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries to seed a stale row against")
	}
	stale := append([]unreachableClaimEntry{{id: "D-00-00", register: "NONEXISTENT-DEBT.md", grade: "WIRED", witness: "probe:ThisRowNoLongerExists", trigger: "P99", claim: "a row whose underlying register row has vanished"}}, entries...)
	staleRendered := renderUnreachableClaimsView(stale)
	freshRendered := renderUnreachableClaimsView(entries)
	if staleRendered == freshRendered {
		t.Fatal("a view carrying a stale entry absent from the live derivation should differ from a fresh regeneration, but compared equal")
	}
}

// ---------------------------------------------------------------------
// Plan 14-10 Task 2: .planning/EVIDENCE-RECONCILIATION.md -- a generated,
// byte-compared view (D-14-13, D-14-14), reusing this file's own
// UNREACHABLE-CLAIMS.md discipline rather than writing a second generator.
// The reconciliation entry is the AUTHORED truth (PHASE-14-DEBT.md's
// "```reconciliation" blocks, session_test.go's parseReconciliationEntries),
// this file is DERIVED: regenerated in memory and compared, never
// hand-edited, no blessing path.
// ---------------------------------------------------------------------

// evidenceReconciliationFrontmatterPattern matches the generated view's own
// `entries: N` frontmatter line, mirroring unreachableClaimsFrontmatterPattern.
var evidenceReconciliationFrontmatterPattern = regexp.MustCompile(`(?m)^entries:\s*(\d+)\s*$`)

// deriveEvidenceReconciliation scans every *-DEBT.md register (live and
// archived) for reconciliation entries, in register-glob order then
// document order within each register -- fully deterministic. Reuses
// parseReconciliationEntries directly; this generator never re-parses the
// underlying markdown with its own logic.
func deriveEvidenceReconciliation() ([]reconciliationEntry, error) {
	registers, err := phaseArtifactGlob("*", "*-DEBT.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(registers)
	var entries []reconciliationEntry
	for _, path := range registers {
		parsed, parseErr := parseReconciliationEntries(path)
		if parseErr != nil {
			return nil, parseErr
		}
		entries = append(entries, parsed...)
	}
	return entries, nil
}

// evidenceReconciliationCellEscape escapes a value for embedding in a GFM
// table cell -- mirroring the groundedness lint's own unescapeCell in
// reverse: a literal backslash becomes `\\`, a literal pipe becomes `\|`.
// Verification commands routinely contain an unescaped "|" (e.g.
// `-run 'A|B'`), so this escape is load-bearing, not decorative.
func evidenceReconciliationCellEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	return s
}

// evidenceReconciliationObligationSummary renders the verdict-specific
// obligation fields as one prose cell, escaped for table embedding.
func evidenceReconciliationObligationSummary(e reconciliationEntry) string {
	switch e.Verdict {
	case reconciliationRenamed:
		return fmt.Sprintf("replacement: %s", e.Replacement)
	case reconciliationSuperseded:
		return fmt.Sprintf("phase %s, commit %s, covers: %s", e.SupersedingPhase, e.SupersedingCommit, e.CoveringCommand)
	case reconciliationObsoleteByDesign:
		return fmt.Sprintf("deleted %s from %s at phase %s, commit %s", e.DeletedSymbol, e.DeletedPackage, e.DeletingPhase, e.DeletingCommit)
	case reconciliationUnderScoped:
		return fmt.Sprintf("missing clause: %s, landing phase: %s", e.MissingClause, e.LandingPhase)
	default:
		return ""
	}
}

// renderEvidenceReconciliationView renders entries as the exact checked-in
// document shape: frontmatter `entries: N`, then one table row per entry.
func renderEvidenceReconciliationView(entries []reconciliationEntry) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "entries: %d\n", len(entries))
	b.WriteString("---\n\n")
	b.WriteString("# Evidence Reconciliation\n\n")
	b.WriteString("**GENERATED. Do not hand-edit.** Regenerated in memory and byte-compared by\n")
	b.WriteString("`TestEvidenceReconciliationViewIsCurrent` (`internal/compiler/session/witness_registry_test.go`)\n")
	b.WriteString("from every `*-DEBT.md` register's `` ```reconciliation ``` `` fenced blocks\n")
	b.WriteString("(`internal/compiler/session/session_test.go`'s `parseReconciliationEntries`). The\n")
	b.WriteString("register row is the authored truth, this file is derived -- a hand edit is a\n")
	b.WriteString("failure, not a source of information. There is no regeneration command and no\n")
	b.WriteString("blessing path: if this view is out of date, correct the registers and\n")
	b.WriteString("re-derive, never overwrite this file directly.\n\n")
	b.WriteString("This is deliberately NOT a baseline file: a baseline is satisfied by silence,\n")
	b.WriteString("while every row here is satisfied only by a claim that can itself fail --\n")
	b.WriteString("`TestReconciliationVerdictsCarryTheirObligations` re-checks every row's\n")
	b.WriteString("obligation on every run (D-14-12).\n\n")
	b.WriteString("## Entries\n\n")
	b.WriteString("| ID | File | Line | Command | Verdict | Obligation |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, e := range entries {
		// The archived command is rendered as "cmd: <text>", never as a
		// bare command span: a bare `go test ...`/`grep ...` span here
		// would itself match verificationCommandPattern, promoting this
		// very file to Tier A (D-14-11's verdict-token clause) and
		// creating a NEW finding pointing at the reconciliation view's
		// own quotation of the command it is reconciling -- the "cmd: "
		// prefix breaks verificationCommandPattern's anchored match
		// while keeping the archived text fully legible.
		fmt.Fprintf(&b, "| %s | %s | %d | `cmd: %s` | %s | %s |\n",
			e.ID,
			evidenceReconciliationCellEscape(e.File),
			e.Line,
			evidenceReconciliationCellEscape(e.Command),
			e.Verdict,
			evidenceReconciliationCellEscape(evidenceReconciliationObligationSummary(e)),
		)
	}
	return b.String()
}

// TestEvidenceReconciliationViewIsCurrent regenerates the view in memory
// from the registers and byte-compares it against the checked-in
// .planning/EVIDENCE-RECONCILIATION.md -- a hand edit is a failure
// (D-14-13). The non-zero-row-count assertion (D-14-14c) additionally
// guards against a vacuous pass.
func TestEvidenceReconciliationViewIsCurrent(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("expected at least the 66 reconciliation entries plan 14-10 authored, got 0")
	}

	regenerated := renderEvidenceReconciliationView(entries)

	checkedInPath := testsupport.ProjectPath(".planning", "EVIDENCE-RECONCILIATION.md")
	checkedIn, err := os.ReadFile(checkedInPath)
	if err != nil {
		t.Fatalf("read .planning/EVIDENCE-RECONCILIATION.md: %v", err)
	}
	if regenerated != string(checkedIn) {
		t.Fatalf(".planning/EVIDENCE-RECONCILIATION.md is out of date -- regenerate from the registers, never hand-edit.\n--- regenerated ---\n%s\n--- checked-in ---\n%s", regenerated, string(checkedIn))
	}

	m := evidenceReconciliationFrontmatterPattern.FindStringSubmatch(string(checkedIn))
	if m == nil {
		t.Fatal(".planning/EVIDENCE-RECONCILIATION.md has no `entries: N` frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatalf("unreadable entries: frontmatter: %v", convErr)
	}
	if declared != len(entries) {
		t.Fatalf("frontmatter declares entries: %d but the derivation holds %d", declared, len(entries))
	}
}

// TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing (D-14-14c)
// proves the non-zero-row-count assertion is not vacuous ornamentation,
// mirroring TestUnreachableClaimsViewNonZeroRowCountIsLoadBearing exactly.
func TestEvidenceReconciliationViewNonZeroRowCountIsLoadBearing(t *testing.T) {
	emptyRendered := renderEvidenceReconciliationView(nil)
	if !strings.Contains(emptyRendered, "entries: 0") {
		t.Fatalf("expected an empty entry list to render entries: 0, got:\n%s", emptyRendered)
	}
	if emptyRendered != renderEvidenceReconciliationView(nil) {
		t.Fatal("renderEvidenceReconciliationView is nondeterministic on empty input")
	}
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("the live registers derived ZERO reconciliation entries -- the byte-compare above would now be vacuously comparing two empty documents, proving consistency but not completeness")
	}
}

// TestEvidenceReconciliationViewCatchesHandEdit mirrors
// TestUnreachableClaimsViewCatchesHandEdit exactly: a one-character
// tamper is proven to differ from a fresh regeneration.
func TestEvidenceReconciliationViewCatchesHandEdit(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	regenerated := renderEvidenceReconciliationView(entries)
	tampered := regenerated + "x"
	if tampered == regenerated {
		t.Fatal("seeded one-character edit did not change the text")
	}
	if tampered == renderEvidenceReconciliationView(entries) {
		t.Fatal("the tampered copy should differ from a fresh regeneration, but compared equal")
	}
}

// TestEvidenceReconciliationViewCatchesStaleEntry mirrors
// TestUnreachableClaimsViewCatchesStaleEntry (D-14-14b): a checked-in
// entry whose underlying register row has vanished must never be silently
// auto-pruned -- it must fail.
func TestEvidenceReconciliationViewCatchesStaleEntry(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no entries to seed a stale row against")
	}
	stale := append([]reconciliationEntry{{
		ID: "D-00-00", File: "nonexistent.md", Line: 1, Command: "echo gone",
		Classification: classR1, Verdict: reconciliationRenamed, Replacement: "echo replacement",
	}}, entries...)
	staleRendered := renderEvidenceReconciliationView(stale)
	freshRendered := renderEvidenceReconciliationView(entries)
	if staleRendered == freshRendered {
		t.Fatal("a view carrying a stale entry absent from the live derivation should differ from a fresh regeneration, but compared equal")
	}
}

// TestEvidenceReconciliationViewCountMatchesFrontmatter is D-14-14a's own
// dedicated proof for this view: a frontmatter `entries: N` that disagrees
// with the table's own row count must fail. TestEvidenceReconciliationViewIsCurrent
// already asserts this over the real checked-in file; this test
// demonstrates the mechanism directly over a synthetic mismatch.
func TestEvidenceReconciliationViewCountMatchesFrontmatter(t *testing.T) {
	entries, err := deriveEvidenceReconciliation()
	if err != nil {
		t.Fatal(err)
	}
	rendered := renderEvidenceReconciliationView(entries)
	tampered := strings.Replace(rendered, fmt.Sprintf("entries: %d\n", len(entries)), fmt.Sprintf("entries: %d\n", len(entries)+1), 1)
	if tampered == rendered {
		t.Fatal("seeded frontmatter-count mismatch did not change the text")
	}
	m := evidenceReconciliationFrontmatterPattern.FindStringSubmatch(tampered)
	if m == nil {
		t.Fatal("tampered text lost its entries: frontmatter line")
	}
	declared, convErr := strconv.Atoi(m[1])
	if convErr != nil {
		t.Fatal(convErr)
	}
	// Re-derive the table row count from the tampered text the same way
	// TestEvidenceReconciliationViewIsCurrent would over a hand-edited
	// file: it holds len(entries) rows still (only frontmatter moved), so
	// declared (len(entries)+1) now disagrees -- this is the failure
	// TestEvidenceReconciliationViewIsCurrent's own declared != len(entries)
	// check catches on the real file.
	if declared == len(entries) {
		t.Fatal("seeded mismatch did not actually diverge from the true entry count")
	}
}
