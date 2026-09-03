package cgen_test

import (
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/session"
)

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
