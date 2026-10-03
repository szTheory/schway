package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/executionpeer"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase26TracerAppSum10(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n check diagnostics: %+v", checked.Diagnostics)
	}
	interpreted, err := interp.Run(checked.Program, "main", "10")
	if err != nil {
		t.Fatalf("interp.Run(main, 10): %v", err)
	}
	if interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != "55" {
		t.Fatalf("interpreter outcome = %+v, want returned 55", interpreted.Outcome)
	}

	artifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil {
		t.Fatalf("BuildApplicationFile: %v", err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("sum_to_n build diagnostics: %+v", diagnostics)
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "10", &stdout, &stderr)
	if err != nil {
		t.Fatalf("RunApplication: %v", err)
	}
	if outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != "55\n" || stderr.Len() != 0 {
		t.Fatalf("app outcome=%+v stdout=%q stderr=%q, want exit 0, 55 newline, empty stderr", outcome, stdout.String(), stderr.String())
	}
}

func TestPhase26Frontend(t *testing.T) {
	source := []byte("module phase26\n\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  // checked scalar path\n  var i = 0\n  var total = 0\n  if n < 1001 {\n    while i < n {\n      i = i + 1\n      total = total + i\n    }\n  } else {\n    defect \"input exceeds 1000\"\n  }\n  total\n}\n")
	first := session.Format(source)
	second := session.Format(first.Canonical)
	if len(first.Diagnostics) != 0 || len(second.Diagnostics) != 0 || !bytes.Equal(first.Canonical, second.Canonical) || !bytes.Contains(first.Canonical, []byte("// checked scalar path")) {
		t.Fatalf("scalar formatter did not preserve comment and stabilize: first=%q second=%q diagnostics=%+v/%+v", first.Canonical, second.Canonical, first.Diagnostics, second.Diagnostics)
	}
	if parsed := syntax.Parse(first.Canonical); len(parsed.Diagnostics) != 0 || len(parsed.Program.Funcs) != 1 || len(parsed.Program.Funcs[0].Body.Linear.Statements) == 0 {
		t.Fatalf("scalar syntax did not parse: %+v", parsed.Diagnostics)
	}
	operatorSites := 0
	for _, token := range syntax.Parse(first.Canonical).Tree.Tokens {
		if token.Kind != syntax.TokenPlus && token.Kind != syntax.TokenLAngle {
			continue
		}
		if token.Span.Start < 0 || token.Span.End > len(first.Canonical) || string(first.Canonical[token.Span.Start:token.Span.End]) != token.Text {
			t.Errorf("operator %q lost its exact source span: %+v", token.Text, token.Span)
		}
		operatorSites++
	}
	if operatorSites != 4 {
		t.Fatalf("scalar source retained %d operator spans, want two less-than and two addition sites", operatorSites)
	}
	for _, malformed := range []string{
		"var = 1",
		"i = + 1",
		"if n { var x = 0 }",
		"while { var x = 0 }",
	} {
		result := syntax.Parse([]byte("module malformed\nexport { fn main }\nfn main(n: U64) -> U64 {\n  " + malformed + "\n  n\n}\n"))
		if len(result.Diagnostics) == 0 || result.Diagnostics[0].Primary.End <= result.Diagnostics[0].Primary.Start {
			t.Errorf("malformed statement %q did not recover with a stable source span: %+v", malformed, result.Diagnostics)
		}
	}
	wrongCondition := []byte("module phase26\nexport { fn main }\nfn main(n: U64) -> U64 {\n  if n { } else { }\n  n\n}\n")
	if checked := session.Check(wrongCondition); len(checked.Diagnostics) == 0 {
		t.Fatal("U64 condition was accepted as truthy")
	}
	boolMutation := []byte("module phase26\nexport { fn main }\nfn main(n: U64) -> U64 {\n  var flag = n < 1\n  flag = n < 2\n  if flag { } else { }\n  n\n}\n")
	if checked := session.Check(boolMutation); len(checked.Diagnostics) != 0 {
		t.Fatalf("Bool scalar declaration, reassignment, and branch were rejected: %+v", checked.Diagnostics)
	}
	branchOnly := []byte("module phase26\nexport { fn main }\nfn main(n: U64) -> U64 {\n  if n < 10 { var x = 1 } else { var y = 2 }\n  x\n}\n")
	if checked := session.Check(branchOnly); !hasDiagnostic(checked.Diagnostics, "name.uninitialized_place") {
		t.Fatalf("branch-only initialized local was not rejected at the join: %+v", checked.Diagnostics)
	}
}

func hasDiagnostic(diagnostics []diagnostic.Diagnostic, code string) bool {
	for _, item := range diagnostics {
		if item.Code == code {
			return true
		}
	}
	return false
}

func TestPhase26AppBoundary(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n check diagnostics: %+v", checked.Diagnostics)
	}
	zero, err := interp.Run(checked.Program, "main", "0")
	if err != nil || zero.Outcome.Kind != "returned" || zero.Outcome.Value != "0" {
		t.Fatalf("zero-iteration interpreter outcome=%+v err=%v, want returned 0", zero.Outcome, err)
	}
	artifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for _, test := range []struct {
		input string
		want  string
	}{{"0", "0\n"}, {"10", "55\n"}, {"1000", "500500\n"}} {
		var stdout, stderr bytes.Buffer
		outcome, err := native.RunApplication(context.Background(), artifact, test.input, &stdout, &stderr)
		if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want || stderr.Len() != 0 {
			t.Errorf("input %s outcome=%+v stdout=%q stderr=%q err=%v, want %q", test.input, outcome, stdout.String(), stderr.String(), err, test.want)
		}
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "1001", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode == 0 || stdout.Len() != 0 || stderr.String() != "schway app: input exceeds 1000\n" {
		t.Fatalf("input 1001 outcome=%+v stdout=%q stderr=%q err=%v, want nonzero with no stdout and bounded defect", outcome, stdout.String(), stderr.String(), err)
	}
}

func TestPhase26ExactSumMatrix(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n check diagnostics: %+v", checked.Diagnostics)
	}
	artifact := filepath.Join(t.TempDir(), "sum_to_n")
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
	}
	for _, test := range []struct {
		input, want string
	}{{"0", "0"}, {"10", "55"}, {"1000", "500500"}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, err := interp.Run(checked.Program, "main", test.input)
			if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != test.want {
				t.Fatalf("interpreter outcome=%+v err=%v, want literal U64 %s", interpreted.Outcome, err, test.want)
			}
			var stdout, stderr bytes.Buffer
			outcome, err := native.RunApplication(context.Background(), artifact, test.input, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want+"\n" || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v, want %q", outcome, stdout.String(), stderr.String(), err, test.want+"\n")
			}
		})
	}
	var stdout, stderr bytes.Buffer
	outcome, err := native.RunApplication(context.Background(), artifact, "1001", &stdout, &stderr)
	if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 65 || stdout.Len() != 0 || stderr.String() != "schway app: input exceeds 1000\n" {
		t.Fatalf("input 1001 outcome=%+v stdout=%q stderr=%q err=%v, want exact pre-output defect", outcome, stdout.String(), stderr.String(), err)
	}

	for _, mutation := range []struct {
		name, from, to, want string
	}{{"wrong-accumulation", "total = total + snapshot", "total = total + 1", "10"}, {"skipped-first-iteration", "var i = 0", "var i = 1", "65"}} {
		t.Run(mutation.name, func(t *testing.T) {
			mutated := strings.Replace(string(source), mutation.from, mutation.to, 1)
			if mutated == string(source) {
				t.Fatalf("source mutation %q did not reach its target", mutation.name)
			}
			mutatedPath := filepath.Join(t.TempDir(), "mutant.schway")
			if err := os.WriteFile(mutatedPath, []byte(mutated), 0o600); err != nil {
				t.Fatal(err)
			}
			mutantArtifact := filepath.Join(t.TempDir(), "mutant")
			_, diagnostics, err := session.BuildApplicationFile(context.Background(), mutatedPath, mutantArtifact, native.DefaultRunner())
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("reached source mutant did not build: diagnostics=%+v err=%v", diagnostics, err)
			}
			mutantCheck := session.Check([]byte(mutated))
			if len(mutantCheck.Diagnostics) != 0 {
				t.Fatalf("mutated source check diagnostics: %+v", mutantCheck.Diagnostics)
			}
			interpreted, err := interp.Run(mutantCheck.Program, "main", "10")
			if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value == "55" || interpreted.Outcome.Value != mutation.want {
				t.Fatalf("interpreter mutant outcome=%+v err=%v; original literal oracle is 55", interpreted.Outcome, err)
			}
			var mutantStdout, mutantStderr bytes.Buffer
			mutantOutcome, err := native.RunApplication(context.Background(), mutantArtifact, "10", &mutantStdout, &mutantStderr)
			if err != nil || mutantOutcome.Kind != native.RunExited || mutantOutcome.ExitCode != 0 || mutantStdout.String() == "55\n" || mutantStderr.Len() != 0 {
				t.Fatalf("native mutant outcome=%+v stdout=%q stderr=%q err=%v; original literal oracle is 55", mutantOutcome, mutantStdout.String(), mutantStderr.String(), err)
			}
		})
	}
}

func TestPhase26SourceRepeatedCopyEvidence(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
	source, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("source-authored scalar-copy check diagnostics: %+v", checked.Diagnostics)
	}
	var copyID, counterID string
	for _, function := range checked.Program.Functions {
		if function.Name != "main" || function.Linear == nil {
			continue
		}
		for _, place := range function.Linear.Places {
			if place.Name == "i" {
				counterID = place.ID
			}
		}
		for _, operation := range function.Linear.Operations {
			if operation.Kind == core.OpCopy {
				copyID = operation.ID
				if operation.SourceID != counterID {
					t.Fatalf("source-authored snapshot copies %q, want loop counter %q", operation.SourceID, counterID)
				}
			}
		}
	}
	if copyID == "" || counterID == "" {
		t.Fatal("ordinary sum_to_n source did not lower let snapshot = i to a scalar OpCopy")
	}

	for _, test := range []struct {
		input string
		count int
		want  string
	}{{"3", 3, "6"}, {"0", 0, "0"}} {
		t.Run("input-"+test.input, func(t *testing.T) {
			interpreted, err := interp.Run(checked.Program, "main", test.input)
			if err != nil || interpreted.Outcome.Kind != "returned" || interpreted.Outcome.Value != test.want {
				t.Fatalf("interpreter outcome=%+v err=%v", interpreted.Outcome, err)
			}
			if err := executionpeer.Validate(checked.Program, interpreted); err != nil {
				t.Fatalf("interpreter peer verification failed: %v", err)
			}
			var interpreterCopies []uint64
			for _, event := range interpreted.Events {
				if event.ID == copyID+":event" {
					interpreterCopies = append(interpreterCopies, event.Occurrence)
				}
			}
			if len(interpreterCopies) != test.count {
				t.Fatalf("interpreter source copy events=%v, want %d", interpreterCopies, test.count)
			}
			for index, occurrence := range interpreterCopies {
				if occurrence != uint64(index) {
					t.Fatalf("interpreter copy event %d occurrence=%d, want %d", index, occurrence, index)
				}
			}

			artifact := filepath.Join(t.TempDir(), "sum_to_n")
			_, diagnostics, err := session.BuildApplicationFile(context.Background(), sourcePath, artifact, native.DefaultRunner())
			if err != nil || len(diagnostics) != 0 {
				t.Fatalf("BuildApplicationFile diagnostics=%+v err=%v", diagnostics, err)
			}
			reportPath := filepath.Join(t.TempDir(), "capture.json")
			var stdout, stderr bytes.Buffer
			outcome, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, test.input, reportPath, native.EvidenceEvents, &stdout, &stderr)
			if err != nil || outcome.Kind != native.RunExited || outcome.ExitCode != 0 || stdout.String() != test.want+"\n" || stderr.Len() != 0 {
				t.Fatalf("native outcome=%+v stdout=%q stderr=%q err=%v", outcome, stdout.String(), stderr.String(), err)
			}
			if report.CaptureStatus != native.EvidenceStatusComplete || report.Execution == nil {
				t.Fatalf("native event capture status=%q execution=%v", report.CaptureStatus, report.Execution)
			}
			if err := executionpeer.Validate(checked.Program, *report.Execution); err != nil {
				t.Fatalf("native peer verification failed: %v", err)
			}
			var nativeCopies []uint64
			for _, event := range report.Execution.Events {
				if event.ID == copyID+":event" {
					nativeCopies = append(nativeCopies, event.Occurrence)
				}
			}
			if len(nativeCopies) != test.count {
				t.Fatalf("native source copy events=%v, want %d", nativeCopies, test.count)
			}
			for index, occurrence := range nativeCopies {
				if occurrence != uint64(index) {
					t.Fatalf("native copy event %d occurrence=%d, want %d", index, occurrence, index)
				}
			}
		})
	}
}
