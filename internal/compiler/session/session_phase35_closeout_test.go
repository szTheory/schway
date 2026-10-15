package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
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

type phase35AdlerCurrentProjection struct {
	sourceBytes   int
	sourceSHA     string
	coreSchema    string
	coreSHA       string
	semanticIDs   []string
	semanticIDSHA string
	emissions     map[string]phase35Emission
}

func phase35CurrentAdlerProjection(t *testing.T) phase35AdlerCurrentProjection {
	t.Helper()
	path := testsupport.ProjectPath("examples", "checksum", "adler32.schway")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(source), "r = b << 16") != 1 || strings.Count(string(source), "r = r + r") != 0 || strings.Count(string(source), "r = r + a") != 1 {
		t.Fatal("Adler source no longer contains exactly the one-line shift and preserved final addition")
	}
	checked, err := session.CheckFile(path)
	if err != nil || len(checked.Diagnostics) != 0 {
		t.Fatalf("check current Adler module: diagnostics=%+v err=%v", checked.Diagnostics, err)
	}
	coreBytes, err := json.Marshal(checked.Program)
	if err != nil {
		t.Fatal(err)
	}
	emitters := []struct {
		name string
		call func() (string, error)
	}{
		{"emit", func() (string, error) { return cgen.Emit(checked.Program) }},
		{"emit_native", func() (string, error) { return cgen.EmitNative(checked.Program) }},
		{"emit_application", func() (string, error) { return cgen.EmitApplication(checked.Program) }},
	}
	projection := phase35AdlerCurrentProjection{
		sourceBytes: len(source),
		sourceSHA:   phase35Digest(source),
		coreSchema:  checked.Program.Schema,
		coreSHA:     phase35Digest(coreBytes),
		semanticIDs: phase35SemanticIDs(t, coreBytes),
		emissions:   make(map[string]phase35Emission, len(emitters)),
	}
	projection.semanticIDSHA = phase35Digest([]byte(strings.Join(projection.semanticIDs, "\n")))
	for _, emitter := range emitters {
		cSource, err := emitter.call()
		projection.emissions[emitter.name] = phase35EmitResult(cSource, err)
		if err == nil {
			t.Logf("current Adler C %s bytes=%d sha=%s", emitter.name, len(cSource), phase35Digest([]byte(cSource)))
		} else {
			t.Logf("current Adler %s outcome=refused code=%s error=%s", emitter.name, phase35ErrorCode(err.Error()), err.Error())
		}
	}
	t.Logf("current Adler source bytes=%d sha=%s core schema=%s core sha=%s semantic IDs=%d/%s", projection.sourceBytes, projection.sourceSHA, projection.coreSchema, projection.coreSHA, len(projection.semanticIDs), projection.semanticIDSHA)
	return projection
}

func phase35EvidenceProblems(evidence string, index *testIndex) []string {
	var problems []string
	normalizedEvidence := strings.Join(strings.Fields(evidence), " ")
	requiredRefs := []string{
		"internal/compiler/session/session_phase35_adler_test.go#TestPhase35ExistingAdlerAnswers",
		"internal/compiler/session/session_phase35_adler_test.go#TestPhase35AdlerWrongShiftControl",
		"internal/compiler/session/session_phase29_interp_fixture_test.go#TestPhase29InterpreterFixtureVectors",
		"internal/compiler/session/session_phase29_native_fixture_test.go#TestPhase29NativeFixtureVectors",
		"internal/compiler/native/native_phase31_matrix_test.go#TestPhase31CaseMatrix",
		"internal/compiler/session/session_phase35_closeout_test.go#TestPhase35OperatorDocumentation",
		"internal/compiler/session/session_phase35_closeout_test.go#TestPhase35PublicDocumentValues",
		"cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness",
		"internal/compiler/session/session_phase35_values_test.go#TestPhase35FocusedValueAnswers",
		"internal/compiler/session/session_phase35_closeout_test.go#TestPhase35EvidenceGrounding",
		"internal/compiler/session/summary_coverage_test.go#TestSummaryCoverageGroundedness",
		"internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeEmitterCorpus",
		"internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeCorpusControls",
	}
	for _, ref := range requiredRefs {
		if !strings.Contains(evidence, ref) {
			problems = append(problems, "evidence missing required test ref: "+ref)
		}
	}
	for _, required := range []string{
		"Fresh task selectors ran from checkout HEAD `26380801`",
		"O0/O3/O3-LTO and UBSan were not rerun for this plan.",
		"Phase 38 hosted binding remains pending.",
		"executionpeer supplies causal admission evidence only; it does not prove numeric answers.",
		"independently authored literal answers and reached wrong-result controls",
	} {
		if !strings.Contains(normalizedEvidence, strings.Join(strings.Fields(required), " ")) {
			problems = append(problems, "evidence missing or alters required scope: "+required)
		}
	}
	knownTests := make(map[string]bool)
	for _, names := range index.byImportPath {
		for name := range names {
			knownTests[name] = true
		}
	}
	kind := ""
	lines := strings.Split(evidence, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "Coverage kind: ") {
			kind = strings.Trim(strings.TrimPrefix(trimmed, "Coverage kind: "), "`.")
		}
	}
	refCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ref: ") {
			continue
		}
		ref := strings.Trim(strings.TrimPrefix(trimmed, "- ref: "), "`")
		refCount++
		file, name, ok := strings.Cut(ref, "#")
		if !ok {
			problems = append(problems, "test evidence reference lacks file#TestName: "+ref)
			continue
		}
		packageDir := filepath.Join(index.root, filepath.FromSlash(filepath.Dir(file)))
		importPath := index.importPathFor(packageDir)
		if !index.byImportPath[importPath][name] {
			problems = append(problems, "unknown or wrong-package test reference: "+ref)
			continue
		}
		coverage := "---\ncoverage:\n  - id: evidence\n    description: Named Phase 35 execution evidence\n    verification:\n      - kind: " + kind + "\n        ref: \"" + ref + "\"\n        status: pass\n    human_judgment: false\nstatus: complete\n---\n"
		if coverageProblems := summaryCoverageProblems(coverage, knownTests); len(coverageProblems) != 0 {
			problems = append(problems, "evidence coverage contract rejected "+ref+": "+strings.Join(coverageProblems, "; "))
		}
	}
	if kind == "" {
		problems = append(problems, "evidence does not declare a coverage kind")
	}
	if refCount == 0 {
		problems = append(problems, "evidence contains no test references")
	}
	return problems
}

func TestPhase35EvidenceGrounding(t *testing.T) {
	current := phase35CurrentAdlerProjection(t)
	index := buildTestIndex(t)
	evidencePath := testsupport.ProjectPath(".planning", "phases", "35-bitwise-operators-literal-shifts-and-hex-literals", "35-EVIDENCE.md")
	// GSD's prohibition prover supplies a known-bad or clean receipt to this
	// same checker; normal Go test runs retain the committed historical input.
	if subject := os.Getenv("GSD_PROHIB_SUBJECT"); subject != "" {
		if !filepath.IsAbs(subject) {
			subject = testsupport.ProjectPath(filepath.FromSlash(subject))
		}
		evidencePath = subject
	}
	evidenceBytes, err := os.ReadFile(evidencePath)
	if err != nil {
		t.Fatal(err)
	}
	evidence := string(evidenceBytes)
	if problems := phase35EvidenceProblems(evidence, index); len(problems) != 0 {
		t.Fatalf("Phase 35 evidence is not grounded: %v", problems)
	}
	controls := []struct {
		name, old, replacement string
	}{
		{"unknown-test", "TestPhase35ExistingAdlerAnswers", "TestPhase35GhostAnswers"},
		{"unsupported-kind", "Coverage kind: `unit`", "Coverage kind: `checksum`"},
		{"stale-revision", "checkout HEAD `26380801`", "checkout HEAD `64dfec6f`"},
		{"missing-lane", "O0/O3/O3-LTO and UBSan were not rerun for this plan.", "All optimization and sanitizer lanes passed in this plan."},
		{"causal-as-numeric", "it does not prove numeric answers", "it proves numeric answers"},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			normalizedEvidence := strings.Join(strings.Fields(evidence), " ")
			if !strings.Contains(normalizedEvidence, strings.Join(strings.Fields(control.old), " ")) {
				t.Fatalf("control anchor %q is absent", control.old)
			}
			mutated := strings.Replace(normalizedEvidence, strings.Join(strings.Fields(control.old), " "), strings.Join(strings.Fields(control.replacement), " "), 1)
			if problems := phase35EvidenceProblems(mutated, index); len(problems) == 0 {
				t.Fatal("copied misleading evidence control was not detected")
			}
		})
	}

	deltaBytes, err := os.ReadFile(testsupport.ProjectPath(".planning", "phases", "35-bitwise-operators-literal-shifts-and-hex-literals", "35-COMPATIBILITY-DELTA.md"))
	if err != nil {
		t.Fatal(err)
	}
	delta := string(deltaBytes)
	corpus := phase35ReadPreChangeCorpus(t)
	var old *phase35PreChangeEntry
	for i := range corpus.Entries {
		if corpus.Entries[i].Path == "examples/checksum/adler32.schway" {
			old = &corpus.Entries[i]
			break
		}
	}
	if old == nil {
		t.Fatal("frozen baseline is missing the declared Adler source row")
	}
	if old.CoreSchema != current.coreSchema || current.coreSchema != "lang.core/1" || old.Emit.Schema != cgen.Schema || cgen.Schema != "lang.c17/0" {
		t.Fatalf("Adler delta changed a schema: old=%s current=%s cgen=%s", old.CoreSchema, current.coreSchema, cgen.Schema)
	}
	if strings.Count(old.Source, "r = r + r") != 16 || strings.Count(old.Source, "r = r + a") != 1 {
		t.Fatal("frozen Adler source no longer records exactly sixteen old doublings and the final addition")
	}
	for _, route := range phase35RouteAccessors {
		before := route.get(old)
		after := current.emissions[route.name]
		if before.Outcome != after.Outcome {
			t.Fatalf("Adler %s outcome changed: before=%s after=%s", route.name, before.Outcome, after.Outcome)
		}
		var row string
		if before.Outcome == "emitted" {
			if before.CSource == after.CSource {
				t.Fatalf("Adler %s output did not record the source delta", route.name)
			}
			row = "| `" + route.name + "` | emitted | " + fmt.Sprint(len(before.CSource)) + " | `" + phase35Digest([]byte(before.CSource)) + "` | emitted | " + fmt.Sprint(len(after.CSource)) + " | `" + phase35Digest([]byte(after.CSource)) + "` |"
			t.Logf("Adler C %s before=%d/%s after=%d/%s", route.name, len(before.CSource), phase35Digest([]byte(before.CSource)), len(after.CSource), phase35Digest([]byte(after.CSource)))
		} else {
			if before.Code != after.Code || before.Error != after.Error {
				t.Fatalf("Adler %s refusal changed: before=%s/%s after=%s/%s", route.name, before.Code, before.Error, after.Code, after.Error)
			}
			row = "| `" + route.name + "` | " + before.Outcome + " (`" + before.Code + "`) | — | — | " + after.Outcome + " (`" + after.Code + "`) | — | — |"
			t.Logf("Adler %s refusal unchanged: %s (%s)", route.name, before.Code, before.Error)
		}
		if !strings.Contains(delta, row) {
			t.Fatalf("compatibility delta lacks exact %s route projection row: %s", route.name, row)
		}
	}
	currentSourceRow := "| source_sha256 | `" + old.SourceSHA + "` | `" + current.sourceSHA + "` |"
	currentSourceBytesRow := "| source_bytes | " + fmt.Sprint(len(old.Source)) + " | " + fmt.Sprint(current.sourceBytes) + " |"
	currentCoreRow := "| checked_core_sha256 | `" + old.CoreSHA + "` | `" + current.coreSHA + "` |"
	currentIDsRow := "| checked_core_semantic_id_count | " + fmt.Sprint(len(old.SemanticIDs)) + " | " + fmt.Sprint(len(current.semanticIDs)) + " |"
	currentIDSHArow := "| checked_core_semantic_ids_sha256 | `" + phase35Digest([]byte(strings.Join(old.SemanticIDs, "\n"))) + "` | `" + current.semanticIDSHA + "` |"
	if !strings.Contains(delta, currentSourceRow) || !strings.Contains(delta, currentSourceBytesRow) || !strings.Contains(delta, currentCoreRow) || !strings.Contains(delta, currentIDsRow) || !strings.Contains(delta, currentIDSHArow) {
		t.Fatalf("compatibility delta lacks exact Adler source/core projections: %q / %q / %q / %q / %q", currentSourceRow, currentSourceBytesRow, currentCoreRow, currentIDsRow, currentIDSHArow)
	}
	t.Logf("Adler source before=%d/%s after=%d/%s; core before=%s after=%s; semantic IDs before=%d/%s after=%d/%s", len(old.Source), old.SourceSHA, current.sourceBytes, current.sourceSHA, old.CoreSHA, current.coreSHA, len(old.SemanticIDs), phase35Digest([]byte(strings.Join(old.SemanticIDs, "\n"))), len(current.semanticIDs), current.semanticIDSHA)
}
