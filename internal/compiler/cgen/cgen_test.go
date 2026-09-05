package cgen_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

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
