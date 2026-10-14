package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/surfacemanifest"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase35ApplicationWitness = `module phase35.application
export { fn main }
fn main(x: U64) -> U64 {
  let y = 0x21
  let witness = x ^ (y << 3) & 0xFF
  witness
}
`

func phase35GuidePath(t *testing.T) string {
	t.Helper()
	return testsupport.ProjectPath("docs", "agent-surface.md")
}

func phase35ReadGuide(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(phase35GuidePath(t))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func phase35OperatorDocumentProblems(guide string, manifest surfacemanifest.Manifest) []string {
	var problems []string
	guide = strings.Join(strings.Fields(guide), " ")
	required := []string{
		"Precedence is, from loosest to tightest, comparisons (`<`, `==`), bitwise OR (`|`), XOR (`^`), AND (`&`), shifts (`<<`, `>>`), addition (`+`), then remainder (`%`). Operators at the same tier associate left to right; `<` and `==` share one tier.",
		"The written expression `x ^ (y << 3) & 0xFF` means `x ^ ((y << 3) & 0xFF)`. With `x=0x100` and `y=0x21`, it returns decimal `264`; the alternate `(x ^ (y << 3)) & 0xFF` returns `8`.",
		"Shift amounts must be integer literals from `0` through `63`. Parentheses around a numeric literal preserve that literal; names and computed expressions remain unsupported, even when their value is known. Left shifts keep the low 64 bits; right shifts are logical.",
		"Scalar expression nesting is separately capped at 64 levels, including nested parentheses and recursively nested `byte_at` indices; this is not the type-nesting limit.",
		"Hexadecimal source tokens accept lowercase and uppercase `0x`/`0X` in scalar expressions and preserve their spelling when formatted. Numeric conversion remains route-specific: scalar `010` is `8` and `0B10` is `2`; the legacy route reads `010` as `10` and refuses `0B10`. Phase 35 adds uppercase `0X` to the legacy route without making the two routes identical.",
		"`examples/bitwise.schway` uses one entry parameter named `unused: Byte` so the existing Byte fixture route can reach the U64/Bool witness. The parameter is not read, copied, stored, or converted.",
		"Public conformance returns an execution document with `outcome.value` equal to `264` on both engines. The U64 application seam separately emits `264\\n`; this is not the public conformance projection.",
		"The existing U64 run input requirement remains in force: there is no default U64 value or new input flag. File input through the interpreter CLI belongs to Phase 37.",
		"`executionpeer` checks causal event admission and fabricated-event refusal; it does not prove numeric answers. Numeric correctness uses independently authored literal answers and reached wrong-result controls.",
		"cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness",
		"internal/compiler/session/session_phase35_values_test.go#TestPhase35FocusedValueAnswers",
		"internal/compiler/session/session_phase35_adler_test.go#TestPhase35ExistingAdlerAnswers",
	}
	for _, clause := range required {
		if !strings.Contains(guide, strings.Join(strings.Fields(clause), " ")) {
			problems = append(problems, "missing or altered guide clause: "+clause)
		}
	}
	if manifest.Schema != surfacemanifest.Schema {
		problems = append(problems, "current surface manifest schema changed")
	}
	wantOperators := map[string]string{
		"u64.bit_xor": "^", "u64.bit_and": "&", "u64.bit_or": "|",
		"u64.shift_left": "<<", "u64.shift_right": ">>",
	}
	for _, operator := range manifest.Operators {
		if want, ok := wantOperators[operator.ID]; ok {
			if operator.Syntax != want {
				problems = append(problems, "manifest operator changed: "+operator.ID)
			}
			delete(wantOperators, operator.ID)
		}
	}
	if len(wantOperators) != 0 {
		problems = append(problems, "manifest is missing documented operators")
	}
	for _, operator := range manifest.Operators {
		constraints := strings.Join(operator.Constraints, " ")
		if operator.ID == "u64.shift_left" && (!strings.Contains(constraints, "0 through 63") || !strings.Contains(constraints, "low 64 bits")) {
			problems = append(problems, "left-shift manifest contract differs from the guide")
		}
		if operator.ID == "u64.shift_right" && (!strings.Contains(constraints, "0 through 63") || !strings.Contains(constraints, "logical")) {
			problems = append(problems, "right-shift manifest contract differs from the guide")
		}
	}
	shiftRefusal := false
	depthLimit := false
	for _, refusal := range manifest.Refusals {
		if refusal.Code == "surface.unsupported_shift_amount" && refusal.Hint == "Shift amounts must be integer literals from 0 to 63; write a literal in that range." {
			shiftRefusal = true
		}
	}
	for _, limit := range manifest.Limits {
		if limit.ID == "parser.expression_depth" && limit.Value == 64 {
			depthLimit = true
		}
	}
	if !shiftRefusal {
		problems = append(problems, "manifest shift refusal or hint differs from the documented range")
	}
	if !depthLimit {
		problems = append(problems, "manifest expression-depth limit is not 64")
	}
	return problems
}

func TestPhase35OperatorDocumentation(t *testing.T) {
	guide := phase35ReadGuide(t)
	normalizedGuide := strings.Join(strings.Fields(guide), " ")
	manifest := surfacemanifest.Current()
	if problems := phase35OperatorDocumentProblems(guide, manifest); len(problems) != 0 {
		t.Fatalf("operator guide does not match current contracts: %v", problems)
	}
	controls := []struct {
		name, old, replacement string
	}{
		{"wrong-value", "it returns decimal `264`", "it returns decimal `265`"},
		{"wrong-count-rule", "from `0` through `63`", "from `0` through `64`"},
		{"causal-as-numeric", "it does not prove numeric answers", "it proves numeric answers"},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			if strings.Count(normalizedGuide, control.old) != 1 {
				t.Fatalf("control anchor %q is absent or ambiguous", control.old)
			}
			mutated := strings.Replace(normalizedGuide, control.old, control.replacement, 1)
			if problems := phase35OperatorDocumentProblems(mutated, manifest); len(problems) == 0 {
				t.Fatal("copied misleading documentation control was not detected")
			}
		})
	}
}

func TestPhase35PublicDocumentValues(t *testing.T) {
	guide := phase35ReadGuide(t)
	manifest := surfacemanifest.Current()
	if problems := phase35OperatorDocumentProblems(guide, manifest); len(problems) != 0 {
		t.Fatalf("documented values do not match current surface: %v", problems)
	}

	publicSource, err := os.ReadFile(testsupport.ProjectPath("examples", "bitwise.schway"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(publicSource), "fn main(unused: Byte) -> U64") != 1 || strings.Count(string(publicSource), "unused") != 1 || !strings.Contains(string(publicSource), "let witness = x ^ (y << 3) & 0xFF") {
		t.Fatal("public witness no longer matches the documented unused-Byte source form")
	}
	publicTest, err := os.ReadFile(testsupport.ProjectPath("cmd", "schway", "main_phase35_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	publicTestText := string(publicTest)
	for _, observation := range []string{
		"TestPhase35PublicConformanceWitness",
		`phase35RunPublicCLI(t, engine, path, "264", operationIDs)`,
		`phase35RunPublicCLI(t, engine, alternatePath, "8", alternateIDs)`,
		"got.Outcome.Value != want",
	} {
		if !strings.Contains(publicTestText, observation) {
			t.Fatalf("public conformance observation %q is not pinned by its named test", observation)
		}
	}

	interpreted, diagnostics, err := session.RunInterpreterApplication([]byte(phase35ApplicationWitness), "256")
	if err != nil || len(diagnostics) != 0 || string(interpreted.Output) != "264\n" {
		t.Fatalf("interpreter application output=%q diagnostics=%+v err=%v, want 264\\n", interpreted.Output, diagnostics, err)
	}
	artifact := filepath.Join(t.TempDir(), "phase35-document-values")
	if _, diagnostics, err := session.BuildApplication(context.Background(), []byte(phase35ApplicationWitness), artifact, native.DefaultRunner()); err != nil || len(diagnostics) != 0 {
		t.Fatalf("build application diagnostics=%+v err=%v", diagnostics, err)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "256", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != "264\n" || stderr.Len() != 0 {
		t.Fatalf("native application outcome=%+v stdout=%q stderr=%q err=%v, want 264\\n", outcome, stdout.String(), stderr.String(), err)
	}
}
