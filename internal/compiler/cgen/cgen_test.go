package cgen_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// foreignReleaseCheckedProgram checks the three-acquisition resource
// fixture, reused by the D-04-20 streaming-emitter tests below.
func foreignReleaseCheckedProgram(t *testing.T) session.CheckResult {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "acquire_three_success.lang"))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("fixture failed to check: %+v", checked.Diagnostics)
	}
	return checked
}

// TestStreamingEmitterWritesAtPointOfOccurrence proves D-04-20's additive
// streaming emitter is selected for a foreign-acquiring function: the
// buffered lang_write_events replay is absent, and the events array opens
// BEFORE the first event is recorded, so a lang_record_event call always
// lands inside an already-open JSON array rather than one materialized only
// at the very end.
func TestStreamingEmitterWritesAtPointOfOccurrence(t *testing.T) {
	checked := foreignReleaseCheckedProgram(t)
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(generated, "lang_write_events") {
		t.Fatalf("a foreign-acquiring function must not use the buffered event replay:\n%s", generated)
	}
	if !strings.Contains(generated, `\"events\":[`) {
		t.Fatalf("expected the streamed events array literal, got:\n%s", generated)
	}
	openIndex := strings.Index(generated, `\"events\":[`)
	firstRecord := strings.Index(generated, `lang_record_event("`) // a CALL site, not the definition
	if openIndex < 0 || firstRecord < 0 || openIndex > firstRecord {
		t.Fatalf("events array must open before the first event is recorded: open=%d first=%d\n%s", openIndex, firstRecord, generated)
	}
}

// TestExistingEmittersAreByteIdentical is D-04-23's regression proof: adding
// the streaming emitter must not move a single byte of any Phase 1/2/3
// generated-C golden, since the buffered emitEventSupport those emitters
// use is completely unmodified by this plan.
func TestExistingEmittersAreByteIdentical(t *testing.T) {
	tests := []struct {
		source []string
		golden []string
		native bool // Emit (false, Phase 1's plain non-JSON mode) vs EmitNative (true)
	}{
		{[]string{"testdata", "phase1", "toggle.lang"}, []string{"testdata", "phase1", "generated.golden.c"}, false},
		{[]string{"testdata", "phase2", "owned_transfer.lang"}, []string{"testdata", "phase2", "owned_transfer.golden.c"}, true},
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
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "foreign_acquire_one.lang"))
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
// hoc namer: the emitted name lives in the closed "LANG_" namespace.
func TestGeneratedForeignHeaderNamesAreAllocated(t *testing.T) {
	checked := foreignAcquireCheckedProgram(t)
	header, err := cgen.EmitForeignHeader(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(header, "typedef struct LANG_LANG_RES_OPEN_RESULT") {
		t.Fatalf("expected an allocated LANG_-namespaced result type, got:\n%s", header)
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
	unit, err := cgen.EmitForeignConformance(checked.Program, testsupport.ProjectPath("native", "lang_foreign_resource_private.h"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{
		`_Static_assert(sizeof(lang_foreign_resource_block) == 1,`,
		`_Static_assert(_Alignof(lang_foreign_resource_block) == 1,`,
		`_Static_assert(offsetof(lang_foreign_resource_block, payload) == 0,`,
		`#include "` + testsupport.ProjectPath("native", "lang_foreign_resource_private.h") + `"`,
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
	privateHeaderPath := testsupport.ProjectPath("native", "lang_foreign_resource_private.h")
	includeLine := `#include "` + privateHeaderPath + `"`

	ordinary, err := cgen.Emit(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(ordinary, privateHeaderPath) {
		t.Fatalf("ordinary Emit output mentions the private header path:\n%s", ordinary)
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
	{"semicolon and brace closing the extern and opening a new definition", "lang_res_open;}\nint injected(void){return 0;}//"},
	{"parenthesis-bearing fragment", "lang_res_open(int x)"},
	{"embedded newline", "lang_res_open\ninjected"},
	{"leading digit", "1lang_res_open"},
	{"embedded space", "lang_res_open injected"},
	{"comment terminator escaping the header comment", "lang_res_open*/int injected(void){return 0;}/*"},
	{"non-ASCII rune", "lang_res_open\u00e9"},
}

// TestForeignSymbolInjectionNeverReachesGeneratedC proves cgen.Emit and
// cgen.EmitNative are never reached with a non-identifier Symbol: both call
// corevalidate.Validate first, so a hostile Symbol is refused with an error
// naming foreign.symbol_not_identifier and an EMPTY generated string -- the
// emptiness is load-bearing, proving no C source containing the injected
// fragment was ever produced.
func TestForeignSymbolInjectionNeverReachesGeneratedC(t *testing.T) {
	positiveControl := foreignAcquireCheckedProgram(t)
	if _, err := cgen.Emit(positiveControl.Program); err != nil {
		t.Fatalf("positive control: unmutated program failed to Emit: %v", err)
	}
	if _, err := cgen.EmitNative(positiveControl.Program); err != nil {
		t.Fatalf("positive control: unmutated program failed to EmitNative: %v", err)
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
	privateHeaderPath := testsupport.ProjectPath("native", "lang_foreign_resource_private.h")

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
	privateHeaderPath := testsupport.ProjectPath("native", "lang_foreign_resource_private.h")

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
			outcome:  "lang_write_buffer_hex(&lang_value_delivered)",
			transfer: "lang_record_event(\"value.transferred\"",
		},
		{
			name:     "Byte copy",
			source:   "module owned.copy\nexport { fn retain }\nfn retain(code: Byte) -> Byte {\n  let kept = code\n  kept\n}\n",
			outcome:  "lang_write_byte(lang_value_kept)",
			transfer: "lang_record_event(\"value.copied\"",
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
				t.Fatal(err)
			}
			for _, required := range []string{
				"#define LANG_OUTPUT_LIMIT 65536u",
				"static int lang_write_json_string(",
				"static int lang_record_event(",
				test.transfer,
				"lang_record_event(\"function.returned\"",
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
			returned := strings.Index(generated, `lang_record_event("function.returned"`)
			outcome := strings.Index(generated, test.outcome)
			if operation < 0 || returned <= operation || outcome <= returned {
				t.Fatalf("runtime event/outcome order is not operation -> return -> serialization: operation=%d return=%d outcome=%d\n%s", operation, returned, outcome, generated)
			}
		})
	}
}

// TestExactlyOneLandingPadIsInstalled proves D-04-17's cost constraint is
// falsifiable, not merely stated: regardless of how many acquisitions or
// foreign calls a function performs, emitLinearForeign installs the
// process-root setjmp landing pad exactly once.
func TestExactlyOneLandingPadIsInstalled(t *testing.T) {
	tests := []struct {
		fixture      string
		acquisitions int
	}{
		{"foreign_acquire_one.lang", 1},
		{"nonlocal_exit_probe.lang", 2},
		{"acquire_three_success.lang", 3},
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
			if err != nil {
				t.Fatal(err)
			}
			count := strings.Count(generated, "setjmp(")
			if count != 1 {
				t.Fatalf("%s (%d acquisitions): setjmp( appears %d times, want exactly 1:\n%s", test.fixture, test.acquisitions, count, generated)
			}
		})
	}
}

// TestLedgerIsStaticStorage proves D-04-17/D-04-18's ledger declaration
// carries static storage duration -- never automatic -- since C17 leaves an
// automatic object indeterminate after a longjmp crosses its setjmp call.
func TestLedgerIsStaticStorage(t *testing.T) {
	checked := nonlocalPadCheckedProgram(t, "nonlocal_exit_probe.lang")
	generated, err := cgen.EmitNative(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generated, "static int lang_resource_live[") {
		t.Fatalf("expected the resource ledger to be declared with static storage duration:\n%s", generated)
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
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase4", "nonlocal_exit_probe.lang"))
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
		{"testdata", "phase4", "nonlocal_exit_probe.lang"},
		{"internal", "compiler", "cgen", "cgen.go"},
		{"native", "lang_foreign_nonlocal.c"},
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
