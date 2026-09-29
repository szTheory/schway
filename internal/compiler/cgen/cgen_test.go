package cgen_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// foreignReleaseCheckedProgram checks the three-acquisition historical
// resource fixture, reused by its archival-emitter evidence checks below.
func foreignReleaseCheckedProgram(t *testing.T) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

// TestStreamingEmitterWritesAtPointOfOccurrence checks the archived D-04-20
// foreign-emitter artifact and separately pins today's public refusal. It
// does not claim that the retired streaming body remains executable.
func TestStreamingEmitterWritesAtPointOfOccurrence(t *testing.T) {
	checked := foreignReleaseCheckedProgram(t)
	if _, err := cgen.EmitNative(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("foreign cut must be a named public refusal, got %v", err)
	}
	generatedBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "historical", "acquire_three_success.c"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatedBytes)
	if strings.Contains(generated, "schway_write_events") {
		t.Fatalf("a foreign-acquiring function must not use the buffered event replay:\n%s", generated)
	}
	if !strings.Contains(generated, `\"events\":[`) {
		t.Fatalf("expected the streamed events array literal, got:\n%s", generated)
	}
	openIndex := strings.Index(generated, `\"events\":[`)
	firstRecord := strings.Index(generated, `schway_record_event("`) // a CALL site, not the definition
	if openIndex < 0 || firstRecord < 0 || openIndex > firstRecord {
		t.Fatalf("events array must open before the first event is recorded: open=%d first=%d\n%s", openIndex, firstRecord, generated)
	}
}

// TestExistingEmittersAreByteIdentical keeps the admitted Phase 1/2 output
// goldens stable while the archived foreign-emitter implementation is
// retired.
func TestExistingEmittersAreByteIdentical(t *testing.T) {
	tests := []struct {
		source []string
		golden []string
		native bool // Emit (false, Phase 1's plain non-JSON mode) vs EmitNative (true)
	}{
		{[]string{"testdata", "phase1", "toggle.schway"}, []string{"testdata", "phase1", "generated.golden.c"}, false},
		{[]string{"testdata", "phase2", "owned_transfer.schway"}, []string{"testdata", "phase2", "owned_transfer.golden.c"}, true},
	}
	for _, test := range tests {
		t.Run(strings.Join(test.golden, "/"), func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(test.source...))
			if err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
			}
			var generated string
			if test.native {
				generated, err = cgen.EmitNative(checked.Program)
			} else {
				generated, err = cgen.Emit(checked.Program)
			}
			if err != nil {
				t.Fatal(err)
			}
			golden, err := os.ReadFile(testsupport.ProjectPath(test.golden...))
			if err != nil {
				t.Fatal(err)
			}
			if generated != string(golden) {
				t.Fatalf("existing emitter output moved after adding the streaming emitter:\n--- got ---\n%s\n--- want ---\n%s", generated, golden)
			}
		})
	}
}

func foreignAcquireCheckedProgram(t *testing.T) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

// TestGeneratedForeignHeaderNamesAreAllocated proves EmitForeignHeader
// allocates its own result-type identifier through the existing cNames
// deterministic-suffix-on-collision machinery (D-04-12a), not a second ad
// hoc namer: the emitted name lives in the closed "SCHWAY_" namespace.
func TestGeneratedForeignHeaderNamesAreAllocated(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(header, "typedef struct SCHWAY_SCHWAY_RES_OPEN_RESULT") {
		t.Fatalf("expected an allocated SCHWAY_-namespaced result type, got:\n%s", header)
	}
}

// TestObligationCommentsAreGeneratedFromJSON proves the header's obligation
// comment block is generated FROM the same core.ForeignContract value the
// sidecar manifest serializes (D-04-12): mutating one contract field must
// move its corresponding comment line.
func TestObligationCommentsAreGeneratedFromJSON(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	before, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(before, "/* allocator: libc_malloc */") {
		t.Fatalf("expected the original allocator comment line, got:\n%s", before)
	}
	mutated := checked.Program
	mutatedContract := *mutated.Functions[0].ForeignContract
	mutatedContract.Allocator = "a_different_allocator"
	mutated.Functions[0].ForeignContract = &mutatedContract
	after, err := cgen.EmitForeignHeader(mutated)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(after, "/* allocator: libc_malloc */") {
		t.Fatalf("allocator comment line did not move after mutating the contract:\n%s", after)
	}
	if !strings.Contains(after, "/* allocator: a_different_allocator */") {
		t.Fatalf("expected the mutated allocator comment line, got:\n%s", after)
	}
}

// TestConformanceUnitAssertsEveryField proves EmitForeignConformance carries
// one sizeof/_Alignof/offsetof _Static_assert triple per declared record
// field, plus record-level size/alignment assertions, over the REAL private
// header's declared record.
func TestConformanceUnitAssertsEveryField(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	unit, err := cgen.EmitForeignConformance(checked.Program, testsupport.ProjectPath("native", "schway_foreign_resource_private.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		`_Static_assert(sizeof(schway_foreign_resource_block) == 1,`,
		`_Static_assert(_Alignof(schway_foreign_resource_block) == 1,`,
		`_Static_assert(offsetof(schway_foreign_resource_block, payload) == 0,`,
		`#include "` + testsupport.ProjectPath("native", "schway_foreign_resource_private.h") + `"`,
	} {
		if !strings.Contains(unit, required) {
			t.Fatalf("conformance unit missing %q:\n%s", required, unit)
		}
	}
}

// TestPrivateHeaderIsIncludedOnlyByConformanceUnit proves the private
// header's #include line appears only in EmitForeignConformance's own
// output -- never in the ordinary Emit/EmitNative artifacts or in
// EmitForeignHeader's own output, both of which the compiler forms without
// ever parsing the foreign side (D-04-10).
func TestPrivateHeaderIsIncludedOnlyByConformanceUnit(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	privateHeaderPath := testsupport.ProjectPath("native", "schway_foreign_resource_private.h")
	includeLine := `#include "` + privateHeaderPath + `"`

	if _, err := cgen.Emit(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("ordinary Emit must retain the named M004 refusal, got %v", err)
	}
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(header, privateHeaderPath) {
		t.Fatalf("EmitForeignHeader output mentions the private header path:\n%s", header)
	}
	unit, err := cgen.EmitForeignConformance(checked.Program, privateHeaderPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(unit, includeLine) {
		t.Fatalf("conformance unit does not include the private header:\n%s", unit)
	}
}

// hostileForeignSymbols is the shared table of hostile Symbol values used to
// falsify both corevalidate's foreign.symbol_not_identifier audit and cgen's
// own independent validForeignSymbol guard against the SAME inputs
// (04-VERIFICATION.md gap 2, FFI-01/D-04-12).
var hostileForeignSymbols = []struct {
	name   string
	symbol string
}{
	{"semicolon and brace closing the extern and opening a new definition", "schway_res_open;}\nint injected(void){return 0;}//"},
	{"parenthesis-bearing fragment", "schway_res_open(int x)"},
	{"embedded newline", "schway_res_open\ninjected"},
	{"leading digit", "1schway_res_open"},
	{"embedded space", "schway_res_open injected"},
	{"comment terminator escaping the header comment", "schway_res_open*/int injected(void){return 0;}/*"},
	{"non-ASCII rune", "schway_res_open\u00e9"},
}

// TestForeignSymbolInjectionNeverReachesGeneratedC proves cgen.Emit and
// cgen.EmitNative are never reached with a non-identifier Symbol: both call
// corevalidate.Validate first, so a hostile Symbol is refused with an error
// naming foreign.symbol_not_identifier and an EMPTY generated string -- the
// emptiness is load-bearing, proving no C source containing the injected
// fragment was ever produced.
func TestForeignSymbolInjectionNeverReachesGeneratedC(t *testing.T) {
	positiveControl := foreignAcquireCheckedProgram(t)
	for _, emit := range []func(core.Program) (string, error){cgen.Emit, cgen.EmitNative} {
		if _, err := emit(positiveControl.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
			t.Fatalf("positive control must retain named M004 refusal, got %v", err)
		}
	}

	for _, hostileCase := range hostileForeignSymbols {
		t.Run(hostileCase.name, func(t *testing.T) {
			// Fresh session.CheckResult per subtest (each call to
			// foreignAcquireCheckedProgram re-checks the fixture from
			// source), so mutating this subtest's contract cannot leak
			// into the positive control or any sibling subtest via a
			// shared Functions slice backing array.
			checked := foreignAcquireCheckedProgram(t)
			mutated := checked.Program
			contract := *mutated.Functions[0].ForeignContract
			contract.Symbol = hostileCase.symbol
			mutated.Functions[0].ForeignContract = &contract

			generated, err := cgen.Emit(mutated)
			if err == nil {
				t.Fatalf("Emit: expected an error, got generated C:\n%s", generated)
			}
			if !strings.Contains(err.Error(), "foreign.symbol_not_identifier") {
				t.Fatalf("Emit: expected error to name foreign.symbol_not_identifier, got %v", err)
			}
			if generated != "" {
				t.Fatalf("Emit: expected empty generated string, got:\n%s", generated)
			}

			generatedNative, err := cgen.EmitNative(mutated)
			if err == nil {
				t.Fatalf("EmitNative: expected an error, got generated C:\n%s", generatedNative)
			}
			if !strings.Contains(err.Error(), "foreign.symbol_not_identifier") {
				t.Fatalf("EmitNative: expected error to name foreign.symbol_not_identifier, got %v", err)
			}
			if generatedNative != "" {
				t.Fatalf("EmitNative: expected empty generated string, got:\n%s", generatedNative)
			}
		})
	}
}

// TestForeignEmittersRefuseNonIdentifierSymbolIndependently proves the three
// exported entry points that never call corevalidate.Validate --
// EmitForeignManifest, EmitForeignHeader, EmitForeignConformance -- refuse a
// hostile Symbol on their own terms via singleForeignFunction's independent
// validForeignSymbol guard.
func TestForeignEmittersRefuseNonIdentifierSymbolIndependently(t *testing.T) {
	privateHeaderPath := testsupport.ProjectPath("native", "schway_foreign_resource_private.h")

	positiveControl := foreignAcquireCheckedProgram(t)
	if _, err := cgen.EmitForeignManifest(positiveControl.Program); err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignManifest: %v", err)
	}
	if _, err := cgen.EmitForeignHeader(positiveControl.Program); err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignHeader: %v", err)
	}
	if _, err := cgen.EmitForeignConformance(positiveControl.Program, privateHeaderPath); err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignConformance: %v", err)
	}

	for _, hostileCase := range hostileForeignSymbols {
		t.Run(hostileCase.name, func(t *testing.T) {
			// Fresh session.CheckResult per subtest -- see the identical
			// note in TestForeignSymbolInjectionNeverReachesGeneratedC.
			checked := foreignAcquireCheckedProgram(t)
			mutated := checked.Program
			contract := *mutated.Functions[0].ForeignContract
			contract.Symbol = hostileCase.symbol
			mutated.Functions[0].ForeignContract = &contract

			manifest, err := cgen.EmitForeignManifest(mutated)
			if err == nil {
				t.Fatalf("EmitForeignManifest: expected an error, got:\n%s", manifest)
			}
			if manifest != "" {
				t.Fatalf("EmitForeignManifest: expected empty string, got:\n%s", manifest)
			}
			if strings.Contains(manifest, "injected") {
				t.Fatalf("EmitForeignManifest: injected fragment leaked into output:\n%s", manifest)
			}

			header, err := cgen.EmitForeignHeader(mutated)
			if err == nil {
				t.Fatalf("EmitForeignHeader: expected an error, got:\n%s", header)
			}
			if header != "" {
				t.Fatalf("EmitForeignHeader: expected empty string, got:\n%s", header)
			}
			if strings.Contains(header, "injected") {
				t.Fatalf("EmitForeignHeader: injected fragment leaked into output:\n%s", header)
			}

			conformance, err := cgen.EmitForeignConformance(mutated, privateHeaderPath)
			if err == nil {
				t.Fatalf("EmitForeignConformance: expected an error, got:\n%s", conformance)
			}
			if conformance != "" {
				t.Fatalf("EmitForeignConformance: expected empty string, got:\n%s", conformance)
			}
			if strings.Contains(conformance, "injected") {
				t.Fatalf("EmitForeignConformance: injected fragment leaked into output:\n%s", conformance)
			}
		})
	}
}

// hostileForeignContractFields is 04-13 Task 3's shared table of hostile
// values used to falsify corevalidate's foreign.contract_field_not_c_safe
// audit and cgen's own independent unsafeForeignContractField guard against
// the SAME inputs (04-VERIFICATION.md gap 2b, FFI-01).
var hostileForeignContractFields = []struct {
	name  string
	field string
	value string
}{
	{"Allocator: comment-terminator payload", "Allocator", "*/ int injected(void){return 1;} /*"},
	{"Unwind: comment-terminator payload", "Unwind", "*/ int injected(void){return 1;} /*"},
	{"NonlocalExit: comment-terminator payload", "NonlocalExit", "*/ int injected(void){return 1;} /*"},
	{"Fails: comment-terminator payload", "Fails", "*/ int injected(void){return 1;} /*"},
	{"InitializedState: comment-terminator payload", "InitializedState", "*/ int injected(void){return 1;} /*"},
	{"Capture: comment-terminator payload", "Capture", "*/ int injected(void){return 1;} /*"},
	{"Retention: comment-terminator payload", "Retention", "*/ int injected(void){return 1;} /*"},
	{"Aliasing: comment-terminator payload", "Aliasing", "*/ int injected(void){return 1;} /*"},
	{"ForeignTypeName: comment-terminator payload", "ForeignTypeName", "*/ int injected(void){return 1;} /*"},
	{"FieldName: comment-terminator payload", "FieldName", "*/ int injected(void){return 1;} /*"},
	{"CType: comment-terminator payload", "CType", "*/ int injected(void){return 1;} /*"},
}

func mutateForeignContractField(contract *core.ForeignContract, field, value string) {
	switch field {
	case "Allocator":
		contract.Allocator = value
	case "Unwind":
		contract.Unwind = value
	case "NonlocalExit":
		contract.NonlocalExit = value
	case "Fails":
		contract.Fails = value
	case "InitializedState":
		contract.InitializedState = value
	case "Capture":
		contract.Capture = value
	case "Retention":
		contract.Retention = value
	case "Aliasing":
		contract.Aliasing = value
	case "ForeignTypeName":
		contract.Layout.ForeignTypeName = value
	case "FieldName":
		contract.Layout.Fields[0].Name = value
	case "CType":
		contract.Layout.Fields[0].CType = value
	}
}

// TestForeignPolicyValueInjectionNeverReachesGeneratedC is 04-13 Task 3's
// cgen falsifier (04-VERIFICATION.md gap 2b, FFI-01): EmitForeignHeader,
// EmitForeignConformance and EmitForeignManifest each refuse a hostile
// core.ForeignContract field with a non-nil error and an EMPTY returned
// string, WITHOUT any help from corevalidate.Validate (none of the three
// calls it) -- cgen's own independent unsafeForeignContractField guard in
// singleForeignFunction is what refuses.
func TestForeignPolicyValueInjectionNeverReachesGeneratedC(t *testing.T) {
	privateHeaderPath := testsupport.ProjectPath("native", "schway_foreign_resource_private.h")

	positiveControl := foreignAcquireCheckedProgram(t)
	if _, err := cgen.EmitForeignManifest(positiveControl.Program); err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignManifest: %v", err)
	}
	header, err := cgen.EmitForeignHeader(positiveControl.Program)
	if err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignHeader: %v", err)
	}
	if !strings.Contains(header, "/* allocator: libc_malloc */") {
		t.Fatalf("positive control: expected the original allocator comment line, got:\n%s", header)
	}
	if _, err := cgen.EmitForeignConformance(positiveControl.Program, privateHeaderPath); err != nil {
		t.Fatalf("positive control: unmutated program failed EmitForeignConformance: %v", err)
	}

	for _, hostileCase := range hostileForeignContractFields {
		t.Run(hostileCase.name, func(t *testing.T) {
			// Fresh session.CheckResult per subtest -- see the identical note
			// in TestForeignSymbolInjectionNeverReachesGeneratedC.
			checked := foreignAcquireCheckedProgram(t)
			mutated := checked.Program
			contract := *mutated.Functions[0].ForeignContract
			layout := *contract.Layout
			layout.Fields = append([]core.LayoutField{}, contract.Layout.Fields...)
			contract.Layout = &layout
			mutateForeignContractField(&contract, hostileCase.field, hostileCase.value)
			mutated.Functions[0].ForeignContract = &contract

			manifest, err := cgen.EmitForeignManifest(mutated)
			if err == nil {
				t.Fatalf("EmitForeignManifest: expected an error, got:\n%s", manifest)
			}
			if manifest != "" {
				t.Fatalf("EmitForeignManifest: expected empty string, got:\n%s", manifest)
			}
			if strings.Contains(err.Error(), "int injected(") {
				t.Fatalf("EmitForeignManifest: error echoed the injected payload: %v", err)
			}

			generatedHeader, err := cgen.EmitForeignHeader(mutated)
			if err == nil {
				t.Fatalf("EmitForeignHeader: expected an error, got:\n%s", generatedHeader)
			}
			if generatedHeader != "" {
				t.Fatalf("EmitForeignHeader: expected empty string, got:\n%s", generatedHeader)
			}
			if strings.Contains(generatedHeader, "int injected(") {
				t.Fatalf("EmitForeignHeader: injected fragment leaked into output:\n%s", generatedHeader)
			}
			if strings.Contains(err.Error(), "int injected(") {
				t.Fatalf("EmitForeignHeader: error echoed the injected payload: %v", err)
			}

			conformance, err := cgen.EmitForeignConformance(mutated, privateHeaderPath)
			if err == nil {
				t.Fatalf("EmitForeignConformance: expected an error, got:\n%s", conformance)
			}
			if conformance != "" {
				t.Fatalf("EmitForeignConformance: expected empty string, got:\n%s", conformance)
			}
			if strings.Contains(conformance, "int injected(") {
				t.Fatalf("EmitForeignConformance: injected fragment leaked into output:\n%s", conformance)
			}
			if strings.Contains(err.Error(), "int injected(") {
				t.Fatalf("EmitForeignConformance: error echoed the injected payload: %v", err)
			}
		})
	}
}

func TestLinearCSerializesRuntimeState(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		outcome  string
		transfer string
	}{
		{
			name:     "Buffer move",
			source:   "module owned.transfer\nexport { fn relay }\nfn relay(buffer: Buffer) -> Buffer {\n  let delivered = take buffer\n  delivered\n}\n",
			outcome:  "schway_write_buffer_hex(&schway_entry_output)",
			transfer: "schway_record_event(\"value.transferred\"",
		},
		{
			name:     "Byte copy",
			source:   "module owned.copy\nexport { fn retain }\nfn retain(code: Byte) -> Byte {\n  let kept = code\n  kept\n}\n",
			outcome:  "schway_write_byte(schway_entry_output)",
			transfer: "schway_record_event(\"value.copied\"",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			checked := session.Check([]byte(test.source))
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture did not check: %+v", checked.Diagnostics)
			}
			generated, err := cgen.EmitNative(checked.Program)
			if err != nil {
				if strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
					t.Skipf("historical landing-pad lowering is M004-frozen; current public refusal: %v; see probe:TestPhase16M004CorpusRefusal", err)
				}
				t.Fatal(err)
			}
			for _, required := range []string{
				"#define SCHWAY_OUTPUT_LIMIT 16777216u",
				"static int schway_write_json_string(",
				"static int schway_record_event(",
				test.transfer,
				"schway_record_event(\"function.returned\"",
				test.outcome,
			} {
				if !strings.Contains(generated, required) {
					t.Fatalf("generated C omits runtime serialization %q:\n%s", required, generated)
				}
			}
			if strings.Contains(generated, `puts("{\"schema\":\"lang.execution/1\"`) {
				t.Fatalf("generated C embeds a precomputed execution document:\n%s", generated)
			}
			operation := strings.Index(generated, test.transfer)
			returned := strings.Index(generated, `schway_record_event("function.returned"`)
			outcome := strings.Index(generated, test.outcome)
			if operation < 0 || returned <= operation || outcome <= returned {
				t.Fatalf("runtime event/outcome order is not operation -> return -> serialization: operation=%d return=%d outcome=%d\n%s", operation, returned, outcome, generated)
			}
		})
	}
}

// TestForeignLandingPadEmitterRemainsRefused pins the current public
// whole-program refusal for fixtures whose historical C used a nonlocal
// landing pad. The archived landing-pad implementation is not a live emitter.
func TestForeignLandingPadEmitterRemainsRefused(t *testing.T) {
	tests := []struct {
		fixture      string
		acquisitions int
	}{
		{"foreign_acquire_one.schway", 1},
		{"nonlocal_exit_probe.schway", 2},
		{"acquire_three_success.schway", 3},
	}
	for _, test := range tests {
		t.Run(test.fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", test.fixture))
			if err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
			}
			generated, err := cgen.EmitNative(checked.Program)
			if err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
				t.Fatalf("%s (%d acquisitions): current public emitter must refuse the foreign body, got source=%q err=%v", test.fixture, test.acquisitions, generated, err)
			}
		})
	}
}

// TestForeignResourceLedgerEmitterRemainsRefused prevents the old resource
// ledger artifact from being mistaken for current native cleanup support.
func TestForeignResourceLedgerEmitterRemainsRefused(t *testing.T) {
	checked := nonlocalPadCheckedProgram(t, "nonlocal_exit_probe.schway")
	generated, err := cgen.EmitNative(checked.Program)
	if err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("current public emitter must refuse the foreign body, got source=%q err=%v", generated, err)
	}
}

// nonlocalPadCheckedProgram checks a Task 04-05-01 fixture by name, reused
// across this file's landing-pad tests.
func nonlocalPadCheckedProgram(t *testing.T, fixture string) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

// TestNonlocalExitReachabilityIsRecorded proves D-04-21's reachability
// register requirement: the nonlocal-exit probe fixture itself carries an
// in-code comment naming both what it reaches and, per D-10, at least one
// shape it does NOT reach -- following the Phase 3 generator-reachability
// convention (03-VALIDATION.md's Generator Reachability Register: "each
// generator must carry a comment stating what it reaches and what it does
// not").
func TestNonlocalExitReachabilityIsRecorded(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "nonlocal_exit_probe.schway"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, required := range []string{
		"REACHABILITY",
		"REACHES:",
		"DOES NOT REACH",
		"foreign-invoked callback",
		"terminates the process directly",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("expected the fixture's reachability comment to mention %q, got:\n%s", required, text)
		}
	}
}

// TestBlindSpotsAreNamedNotClaimed scans the fixture, this package's own
// source, and the frozen foreign TU for the two accepted residual
// blind-spot phrases (D-04-21/T-04-33) and asserts none of them is ever
// claimed as COVERED, DETECTED, or PROVEN anywhere near a blind-spot
// mention -- named, never quietly implied handled.
func TestBlindSpotsAreNamedNotClaimed(t *testing.T) {
	paths := [][]string{
		{"testdata", "phase4", "nonlocal_exit_probe.schway"},
		{"internal", "compiler", "cgen", "cgen.go"},
		{"native", "schway_foreign_nonlocal.c"},
	}
	blindSpots := []string{"foreign-invoked callback", "terminates the process directly"}
	forbidden := []string{"covers all nonlocal exits", "detects every nonlocal exit", "proven to catch every"}
	found := make(map[string]bool, len(blindSpots))
	for _, path := range paths {
		source, err := os.ReadFile(testsupport.ProjectPath(path...))
		if err != nil {
			t.Fatal(err)
		}
		text := string(source)
		for _, blindSpot := range blindSpots {
			if strings.Contains(text, blindSpot) {
				found[blindSpot] = true
			}
		}
		lower := strings.ToLower(text)
		for _, claim := range forbidden {
			if strings.Contains(lower, claim) {
				t.Fatalf("%s claims coverage (%q) of a named blind spot -- this must never be claimed", strings.Join(path, "/"), claim)
			}
		}
	}
	for _, blindSpot := range blindSpots {
		if !found[blindSpot] {
			t.Fatalf("blind spot %q is never named in any scanned source", blindSpot)
		}
	}
}

// TestPhase5ByPointerLoweringGolden pins the historical D-05-02 C artifact
// beside the current named public refusal. The archived marker remains
// useful to historical fixture controls; it is not emitted by production
// Emit/EmitNative.
func TestPhase5ByPointerLoweringGolden(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	if _, err := cgen.Emit(checked.Program); err == nil || !strings.Contains(err.Error(), "by-pointer bodies are not supported") {
		t.Fatalf("by-pointer cut must be a named public refusal, got %v", err)
	}
	generatedBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "historical", "restrict_borrow.c"))
	if err != nil {
		t.Fatal(err)
	}
	generated := string(generatedBytes)
	golden, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.golden.c"))
	if err != nil {
		t.Fatal(err)
	}
	if generated != string(golden) {
		t.Fatalf("archived by-pointer output moved from its committed golden:\n--- got ---\n%s\n--- want ---\n%s", generated, string(golden))
	}
	if count := strings.Count(generated, "/* schway:by-pointer-param */"); count != 1 {
		t.Fatalf("expected exactly one by-pointer-param marker, got %d in:\n%s", count, generated)
	}
}

// phase1Through4AcceptingFixtures is the same accepting-fixture domain
// internal/compiler/core/core_test.go's pinnedFixtures pins (duplicated here
// rather than imported, since core_test.go lives in an external _test
// package with no exported symbol for it) plus the Phase 4 accepting
// fixtures core_test.go's D-05-39 widening added. TestPhase5ByPointerLoweringIsAdditive
// enumerates every one of them.
var phase1Through4AcceptingFixtures = []string{
	"testdata/phase1/comments.schway",
	"testdata/phase1/toggle.schway",
	"testdata/phase2/implicit_copy.schway",
	"testdata/phase2/owned_transfer.schway",
	"testdata/phase3/borrowed_view.schway",
	"testdata/phase3/branch_one_arm_shared_accept.schway",
	"testdata/phase3/branch_view.schway",
	"testdata/phase3/public_view.schway",
	"testdata/phase3/public_view_impossible.schway",
	"testdata/phase3/public_view_mixed_access.schway",
	"testdata/phase3/public_view_multi_arm_access_conflict.schway",
	"testdata/phase3/public_view_multi_arm_omitted.schway",
	"testdata/phase3/public_view_omitted.schway",
	"testdata/phase3/public_view_understated.schway",
	"testdata/phase3/sequential_shared_then_exclusive_accept.schway",
	"testdata/phase3/shared_shared_accept.schway",
	"testdata/phase4/acquire_three_fail_second.schway",
	"testdata/phase4/acquire_three_fail_third.schway",
	"testdata/phase4/acquire_three_success.schway",
	"testdata/phase4/defect_terminal.schway",
	"testdata/phase4/discard_because.schway",
	"testdata/phase4/foreign_acquire_one.schway",
	"testdata/phase4/nonlocal_exit_probe.schway",
}

// TestPhase5ByPointerLoweringIsAdditive proves the retained structural
// classifier does not mistake earlier accepted fixtures for the cut
// by-pointer family. The classifier supports manifest description and
// refusal checks; it does not dispatch production lowering.
func TestPhase5ByPointerLoweringIsAdditive(t *testing.T) {
	for _, path := range phase1Through4AcceptingFixtures {
		t.Run(path, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(strings.Split(path, "/")...))
			if err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
			}
			for _, function := range checked.Program.Functions {
				if cgen.SelectsByPointerLowering(function, function.Linear) {
					t.Fatalf("function %q wrongly matches the cut by-pointer shape classifier", function.Name)
				}
			}
		})
	}
}

// TestPhase5ByPointerLoweringThreeEngineAgreementRemainsDeferred confirms
// the historical three-engine experiment stays behind the whole-program
// refusal until family-specific discharge evidence is admitted.
func TestPhase5ByPointerLoweringThreeEngineAgreementRemainsDeferred(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway")
	result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
	if len(diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", diagnostics)
	}
	if err == nil || !strings.Contains(err.Error(), "by-pointer bodies are not supported") {
		t.Fatalf("current public emitter must refuse the by-pointer body, got result=%+v err=%v", result, err)
	}
}

// aliasProbeSentinel is D-05-36's distinctive core.ForeignContract.Alias
// value: driven through every EmitForeign* entry point, it must appear in
// NONE of their outputs (D-04-33 closed). It is a plain identifier-shaped
// string (no comment terminator), so it stays comment-safe and is never
// itself refused by unsafeForeignContractField's new Alias check -- the
// point of this test is that the field is never spliced anywhere at all,
// not merely that a hostile value would be caught.
const aliasProbeSentinel = "schway_alias_probe_sentinel"

// TestEmitForeignNeverContainsAliasValue is D-05-36's regression proof: every
// exported EmitForeign* entry point is driven with a core.ForeignContract
// carrying aliasProbeSentinel as its Alias value, and none of their outputs
// may ever contain it. Closing the carried D-04-33 gap means a future splice
// site over Alias fails this test loudly rather than silently reopening the
// closed C-injection class.
func TestEmitForeignNeverContainsAliasValue(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	mutatedContract := *checked.Program.Functions[0].ForeignContract
	mutatedContract.Alias = aliasProbeSentinel
	mutated := checked.Program
	mutated.Functions = append([]core.Function(nil), mutated.Functions...)
	mutated.Functions[0].ForeignContract = &mutatedContract

	manifest, err := cgen.EmitForeignManifest(mutated)
	if err != nil {
		t.Fatalf("EmitForeignManifest: %v", err)
	}
	if strings.Contains(manifest, aliasProbeSentinel) {
		t.Fatalf("EmitForeignManifest output contains the Alias sentinel:\n%s", manifest)
	}

	header, err := cgen.EmitForeignHeader(mutated)
	if err != nil {
		t.Fatalf("EmitForeignHeader: %v", err)
	}
	if strings.Contains(header, aliasProbeSentinel) {
		t.Fatalf("EmitForeignHeader output contains the Alias sentinel:\n%s", header)
	}

	conformance, err := cgen.EmitForeignConformance(mutated, testsupport.ProjectPath("native", "schway_foreign_resource_private.h"))
	if err != nil {
		t.Fatalf("EmitForeignConformance: %v", err)
	}
	if strings.Contains(conformance, aliasProbeSentinel) {
		t.Fatalf("EmitForeignConformance output contains the Alias sentinel:\n%s", conformance)
	}
}

// restrictAbsentWithoutAliasFactSource is the negative half of
// TestByPointerAttributeMetadataRemainsSeparateFromBodyAdmission (mirroring check_exclusive_test.go's
// identical partialCoverageExclusiveSource fixture, duplicated here rather
// than imported since it lives in an external _test package with no
// exported symbol for it): the exclusive loan `first` never covers the
// call's own terminator -- `relay` returns `buffer` directly, bypassing the
// loan chain entirely (a borrow never moves ownership) -- so neither
// check's AliasFact nor cgen's selectsByPointerLowering ever fire for this
// function, and the ordinary whole-program emitter can serialize it without
// any by-pointer attribute claim.
const restrictAbsentWithoutAliasFactSource = `module owned.alias_fact_partial_coverage

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let first = borrow mut buffer
  let second = borrow first
  buffer
}
`

// TestByPointerAttributeMetadataRemainsSeparateFromBodyAdmission keeps the
// historical D-05-01 attribute metadata distinct from executable body
// admission. A selected by-pointer shape is refused by Emit; its separate
// manifest may still describe the historical restrict fact. Ordinary
// admitted output does not acquire that token.
func TestByPointerAttributeMetadataRemainsSeparateFromBodyAdmission(t *testing.T) {
	positive, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	positiveChecked := session.Check(positive)
	if len(positiveChecked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", positiveChecked.Diagnostics)
	}
	if _, err := cgen.Emit(positiveChecked.Program); err == nil || !strings.Contains(err.Error(), "by-pointer bodies are not supported") {
		t.Fatalf("positive by-pointer fixture must retain named M004 refusal, got %v", err)
	}
	manifest, err := cgen.EmitForeignManifest(positiveChecked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(manifest, `"attr":"restrict"`) {
		t.Fatalf("expected manifest metadata to retain the historical restrict fact, got:\n%s", manifest)
	}

	negativeChecked := session.Check([]byte(restrictAbsentWithoutAliasFactSource))
	if len(negativeChecked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", negativeChecked.Diagnostics)
	}
	negativeGenerated, err := cgen.Emit(negativeChecked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(negativeGenerated, "restrict") {
		t.Fatalf("expected no restrict token without an alias fact, got:\n%s", negativeGenerated)
	}
	if strings.Contains(negativeGenerated, "restrict") {
		t.Fatalf("ordinary emitted C must not carry the historical by-pointer attribute:\n%s", negativeGenerated)
	}
}

// TestRestrictNeverOnForeignExtern is D-05-03's declaration-site regression
// proof: a foreign-shaped fixture's emitted C, header, and conformance unit
// (the three inspectable layers D-04-12 names) carry zero occurrences of any
// BannedOptimizerAttributes token -- including "restrict" -- in any extern
// declaration. BannedOptimizerAttributes/ScanForBannedAttributes are
// completely unchanged by this plan, so this is a pure non-regression
// proof: the declaration-site ban stays exactly as strict as Phase 4.
func TestRestrictNeverOnForeignExtern(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	if _, err := cgen.Emit(checked.Program); err == nil || !strings.Contains(err.Error(), "multi-function foreign-call bodies are not supported") {
		t.Fatalf("foreign fixture must retain named M004 refusal, got %v", err)
	}
	cSourceBytes, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase16", "historical", "foreign_acquire_one.c"))
	if err != nil {
		t.Fatal(err)
	}
	cSource := string(cSourceBytes)
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	conformance, err := cgen.EmitForeignConformance(checked.Program, testsupport.ProjectPath("native", "schway_foreign_resource_private.h"))
	if err != nil {
		t.Fatal(err)
	}
	if found := cgen.ScanForBannedAttributes(cSource, header, conformance); len(found) != 0 {
		t.Fatalf("expected zero banned attributes in a foreign-shaped fixture's emitted C, got %v", found)
	}
	for name, source := range map[string]string{"c source": cSource, "header": header, "conformance": conformance} {
		if strings.Contains(source, "restrict") {
			t.Fatalf("expected zero occurrences of restrict in the %s, got:\n%s", name, source)
		}
	}
}

// TestEmittedAttributesCarryJustification is D-05-04's own falsifier: every
// EmittedAttribute the Phase 5 fixture's manifest carries has a non-empty
// JustifiedBy binding, naming the exact core node, parameter, and loan
// identity the attribute is justified by.
func TestEmittedAttributesCarryJustification(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase5", "restrict_borrow.schway"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	manifest, err := cgen.EmitForeignManifest(checked.Program)
	if err != nil {
		t.Fatalf("EmitForeignManifest: %v", err)
	}
	var document struct {
		EmittedAttributes []cgen.EmittedAttribute `json:"emitted_attributes"`
	}
	if err := json.Unmarshal([]byte(manifest), &document); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if len(document.EmittedAttributes) != 1 {
		t.Fatalf("want exactly one emitted attribute, got %d: %+v", len(document.EmittedAttributes), document.EmittedAttributes)
	}
	attribute := document.EmittedAttributes[0]
	if attribute.Attr != "restrict" {
		t.Fatalf("want Attr %q, got %q", "restrict", attribute.Attr)
	}
	if attribute.JustifiedBy == "" {
		t.Fatal("attribute carries no justification")
	}
	if len(checked.Program.Functions) != 1 {
		t.Fatalf("want one function, got %d", len(checked.Program.Functions))
	}
	function := checked.Program.Functions[0]
	if attribute.CoreNode != function.ID {
		t.Fatalf("want CoreNode %q, got %q", function.ID, attribute.CoreNode)
	}
	if attribute.Parameter != function.Parameter.ID {
		t.Fatalf("want Parameter %q, got %q", function.Parameter.ID, attribute.Parameter)
	}
	if len(function.Linear.Operations) == 0 || attribute.JustifiedBy != function.Linear.Operations[0].LoanID {
		t.Fatalf("want JustifiedBy to equal the first operation's loan id %q, got %q", function.Linear.Operations[0].LoanID, attribute.JustifiedBy)
	}
}

// TestForeignManifestBytesUnchangedForPriorPhases is D-05-39's own
// falsifier for the foreignManifestDocument.EmittedAttributes element-type
// change ([]string -> []EmittedAttribute): asserts, rather than assumes,
// that an empty slice of either element type serializes to the identical
// JSON array literal `[]`, then confirms every Phase 4 foreign-contract
// fixture's manifest still carries that exact empty array (by-pointer
// lowering and a declared foreign contract are mutually exclusive shapes,
// so these fixtures' EmittedAttributes stays empty, unaffected by D-05-04).
func TestForeignManifestBytesUnchangedForPriorPhases(t *testing.T) {
	oldShape, err := json.Marshal([]string{})
	if err != nil {
		t.Fatal(err)
	}
	newShape, err := json.Marshal([]cgen.EmittedAttribute{})
	if err != nil {
		t.Fatal(err)
	}
	if string(oldShape) != string(newShape) {
		t.Fatalf("empty-slice encodings diverge: old=%s new=%s", oldShape, newShape)
	}

	for _, fixture := range []string{"testdata/phase4/foreign_acquire_one.schway", "testdata/phase4/acquire_three_success.schway"} {
		t.Run(fixture, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(strings.Split(fixture, "/")...))
			if err != nil {
				t.Fatal(err)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
			}
			manifest, err := cgen.EmitForeignManifest(checked.Program)
			if err != nil {
				t.Fatalf("EmitForeignManifest: %v", err)
			}
			if !strings.Contains(manifest, `"emitted_attributes":[]`) {
				t.Fatalf("expected the Phase 4 foreign-contract manifest to keep an empty emitted_attributes array, got:\n%s", manifest)
			}
		})
	}
}
