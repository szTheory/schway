package session_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"unicode/utf8"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

const phase371ReasonProgram = `module phase371.reason_utf8
export { fn main }
fn increment(value: U64) -> U64 {
  let next = value + 1
  next
}
fn main(input: U64) -> U64 {
  var result = increment(input)
  if result < 12 {
    defect "__PHASE371_REASON__"
  } else {
    result = result + 1
  }
  result
}
fn after(input: U64) -> U64 { input }
`

func phase371ReasonSource(reason []byte) []byte {
	marker := []byte("__PHASE371_REASON__")
	template := []byte(phase371ReasonProgram)
	index := bytes.Index(template, marker)
	if index < 0 {
		panic("Phase 37.1 reason template marker missing")
	}
	source := make([]byte, 0, len(template)-len(marker)+len(reason))
	source = append(source, template[:index]...)
	source = append(source, reason...)
	source = append(source, template[index+len(marker):]...)
	return source
}

func phase371FindDefectReason(program *core.Program) (*string, bool) {
	for functionIndex := range program.Functions {
		function := &program.Functions[functionIndex]
		if function.Linear == nil {
			continue
		}
		for operationIndex := range function.Linear.Operations {
			operation := &function.Linear.Operations[operationIndex]
			if operation.Kind == core.OpDefect {
				return &operation.Reason, true
			}
		}
	}
	return nil, false
}

func phase371HasProblemCode(problems []corevalidate.Problem, code string) bool {
	for _, problem := range problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func phase371StraightLineDefectCore(t *testing.T) core.Program {
	t.Helper()
	checked := session.Check([]byte(`module phase371.linear_core
export { fn main }
fn main(input: U64) -> U64 { input }
`))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("straight-line source was refused: %+v", checked.Diagnostics)
	}
	found := false
	for functionIndex := range checked.Program.Functions {
		function := &checked.Program.Functions[functionIndex]
		if function.Name != "main" || function.Match != nil || function.Linear == nil || len(function.Linear.Blocks) != 0 {
			continue
		}
		for operationIndex := range function.Linear.Operations {
			operation := &function.Linear.Operations[operationIndex]
			if operation.Kind == core.OpReturn {
				operation.Kind = core.OpDefect
				operation.Reason = "linear valid reason"
				found = true
			}
		}
	}
	if !found {
		t.Fatal("straight-line source did not produce an OpReturn to adapt for the raw-core path")
	}
	if validated := corevalidate.Validate(checked.Program); !validated.Valid {
		t.Fatalf("valid straight-line OpDefect raw core was refused: %+v", validated.Problems)
	}
	return checked.Program
}

func TestPhase371DefectReasonUTF8Admission(t *testing.T) {
	for _, malformed := range []struct {
		name   string
		reason []byte
	}{
		{name: "raw FF followed by ASCII", reason: []byte{0xff, 'A'}},
		{name: "invalid continuation", reason: []byte{0xe2, 'A'}},
		{name: "truncated sequence", reason: []byte{0xe2, 0x82}},
	} {
		t.Run("source/"+malformed.name, func(t *testing.T) {
			source := phase371ReasonSource(malformed.reason)
			parsed := syntax.Parse(source)
			if !bytes.Equal(parsed.Tree.Bytes(), source) {
				t.Fatal("parser did not preserve malformed source bytes losslessly")
			}
			if len(parsed.Program.Funcs) != 3 || parsed.Program.Funcs[2].Name != "after" {
				t.Fatalf("parser did not recover to the following declaration: funcs=%+v", parsed.Program.Funcs)
			}
			reasonStart := bytes.Index(source, malformed.reason)
			if reasonStart < 0 {
				t.Fatal("malformed reason bytes are absent from source")
			}
			wantInvalid := make(map[[2]int]bool)
			for offset := 0; offset < len(malformed.reason); {
				runeValue, size := utf8.DecodeRune(malformed.reason[offset:])
				if runeValue == utf8.RuneError && size == 1 {
					wantInvalid[[2]int{reasonStart + offset, reasonStart + offset + 1}] = false
				}
				offset += size
			}
			for _, problem := range parsed.Diagnostics {
				if problem.Code != "syntax.invalid_utf8" {
					continue
				}
				span := [2]int{problem.Primary.Start, problem.Primary.End}
				if _, expected := wantInvalid[span]; expected {
					wantInvalid[span] = true
				}
			}
			for span, found := range wantInvalid {
				if !found {
					t.Errorf("missing stable syntax.invalid_utf8 byte span [%d:%d]; diagnostics=%+v", span[0], span[1], parsed.Diagnostics)
				}
			}
			if len(wantInvalid) == 0 {
				t.Fatalf("test vector has no invalid UTF-8 byte: % x", malformed.reason)
			}
		})
	}

	t.Run("outside-quote baseline", func(t *testing.T) {
		source := append([]byte{0xff}, []byte(phase371ReasonProgram)...)
		parsed := syntax.Parse(source)
		if !bytes.Equal(parsed.Tree.Bytes(), source) {
			t.Fatal("outside-quote invalid UTF-8 changed lossless source bytes")
		}
		found := false
		for _, problem := range parsed.Diagnostics {
			if problem.Code == "syntax.invalid_utf8" && problem.Primary.Start == 0 && problem.Primary.End == 1 {
				found = true
			}
		}
		if !found {
			t.Fatalf("existing outside-quote syntax.invalid_utf8 [0:1] changed: %+v", parsed.Diagnostics)
		}
	})

	validReasons := []struct {
		name   string
		reason []byte
	}{
		{name: "replacement character and non-ASCII", reason: []byte("\uFFFD café")},
		{name: "control byte", reason: []byte{0x01, 'A'}},
		{name: "tab control", reason: []byte{'\t', 'A'}},
		{name: "NUL followed by A", reason: []byte{0x00, 'A'}},
		{name: "NUL followed by Z", reason: []byte{0x00, 'Z'}},
		{name: "literal backslash-u spelling", reason: []byte(`\u0041`)},
		{name: "literal backslash-x spelling", reason: []byte(`\xFF`)},
	}
	var replacementSource []byte
	var replacementReason []byte
	var replacementProgram core.Program
	for _, valid := range validReasons {
		t.Run("valid/"+valid.name, func(t *testing.T) {
			source := phase371ReasonSource(valid.reason)
			parsed := syntax.Parse(source)
			if len(parsed.Diagnostics) != 0 || !bytes.Equal(parsed.Tree.Bytes(), source) {
				t.Fatalf("valid reason parse changed source or produced diagnostics: %+v", parsed.Diagnostics)
			}
			checked := session.Check(source)
			if len(checked.Diagnostics) != 0 {
				t.Fatalf("checker refused valid reason % x: %+v", valid.reason, checked.Diagnostics)
			}
			if !utf8.Valid(valid.reason) {
				t.Fatalf("valid-case test reason is invalid UTF-8: % x", valid.reason)
			}
			actual, found := phase371FindDefectReason(&checked.Program)
			if !found || !bytes.Equal([]byte(*actual), valid.reason) {
				t.Fatalf("core reason=% x found=%v, want exact bytes % x", func() []byte {
					if actual == nil {
						return nil
					}
					return []byte(*actual)
				}(), found, valid.reason)
			}
			if validated := corevalidate.Validate(checked.Program); !validated.Valid {
				t.Fatalf("independent core validator refused valid reason % x: %+v", valid.reason, validated.Problems)
			}
			if bytes.Equal(valid.reason, []byte("\uFFFD café")) {
				replacementSource = source
				replacementReason = append([]byte(nil), valid.reason...)
				replacementProgram = checked.Program
			}
		})
	}

	t.Run("raw-core boundary", func(t *testing.T) {
		if len(replacementProgram.Functions) == 0 {
			t.Fatal("valid checked program needed for independent raw-core mutation is absent")
		}
		encoded, err := json.Marshal(replacementProgram)
		if err != nil {
			t.Fatalf("marshal independently checked core: %v", err)
		}
		var mutant core.Program
		if err := json.Unmarshal(encoded, &mutant); err != nil {
			t.Fatalf("clone independently checked core: %v", err)
		}
		reason, found := phase371FindDefectReason(&mutant)
		if !found {
			t.Fatal("cloned checked core has no OpDefect to mutate")
		}
		*reason = string([]byte{0xff, 'A'})
		validated := corevalidate.Validate(mutant)
		if validated.Valid || !phase371HasProblemCode(validated.Problems, "core.final_claim_mismatch") {
			t.Fatalf("raw-core malformed reason was not independently refused as core.final_claim_mismatch: valid=%v problems=%+v", validated.Valid, validated.Problems)
		}
		if !utf8.Valid(replacementReason) || !bytes.Contains(replacementSource, replacementReason) {
			t.Fatal("raw-core mutation was not based on a separately checked valid source reason")
		}
	})

	t.Run("raw-core boundary/straight-line", func(t *testing.T) {
		straightLine := phase371StraightLineDefectCore(t)
		encoded, err := json.Marshal(straightLine)
		if err != nil {
			t.Fatalf("marshal straight-line core: %v", err)
		}
		var mutant core.Program
		if err := json.Unmarshal(encoded, &mutant); err != nil {
			t.Fatalf("clone straight-line core: %v", err)
		}
		reason, found := phase371FindDefectReason(&mutant)
		if !found {
			t.Fatal("cloned straight-line core has no OpDefect to mutate")
		}
		*reason = string([]byte{0xff, 'A'})
		validated := corevalidate.Validate(mutant)
		if validated.Valid || !phase371HasProblemCode(validated.Problems, "core.final_claim_mismatch") {
			t.Fatalf("straight-line raw-core malformed reason was not independently refused as core.final_claim_mismatch: valid=%v problems=%+v", validated.Valid, validated.Problems)
		}
	})

	t.Run("interpreter peer and native O0", func(t *testing.T) {
		if len(replacementProgram.Functions) == 0 {
			t.Fatal("valid checked tracer program is absent")
		}
		run, err := interp.Run(replacementProgram, "main", "7")
		if err != nil {
			t.Fatalf("interpreter run: %v", err)
		}
		if run.Outcome.Kind != "defect" || run.Outcome.Value != "" || len(run.Events) == 0 || run.Events[len(run.Events)-1].Output != string(replacementReason) {
			t.Fatalf("interpreter terminal=%+v events=%+v, want exact defect reason % x", run.Outcome, run.Events, replacementReason)
		}
		if err := executionpeer.Validate(replacementProgram, run); err != nil {
			t.Fatalf("independent peer refused interpreter trace: %v", err)
		}
		generated, err := cgen.EmitNative(replacementProgram)
		if err != nil {
			t.Fatalf("native emission: %v", err)
		}
		runner := native.DefaultRunner()
		runner.Expect = native.ExpectDefect
		result, err := runner.Run(context.Background(), generated, "-O0", []string{"7"})
		if err != nil || len(result.Pairs) != 1 {
			t.Fatalf("native O0 defect trace pairs=%d err=%v", len(result.Pairs), err)
		}
		nativeRun := result.Pairs[0].Execution
		if nativeRun.Outcome.Kind != "defect" || nativeRun.Outcome.Value != "" || len(nativeRun.Events) == 0 || nativeRun.Events[len(nativeRun.Events)-1].Output != string(replacementReason) {
			t.Fatalf("native O0 terminal=%+v events=%+v, want exact reason % x", nativeRun.Outcome, nativeRun.Events, replacementReason)
		}
		if err := executionpeer.Validate(replacementProgram, nativeRun); err != nil {
			t.Fatalf("independent peer refused native O0 trace: %v", err)
		}
	})
}
