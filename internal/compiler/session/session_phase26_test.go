package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

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
	for _, malformed := range []string{
		"var = 1",
		"i = + 1",
		"if 1 { var x = 0 } else { var y = 1 }",
		"while { var x = 0 }",
	} {
		result := syntax.Parse([]byte("module malformed\nfn main(n: U64) -> U64 {\n  " + malformed + "\n  n\n}\n"))
		if len(result.Diagnostics) == 0 || result.Diagnostics[0].Primary.End <= result.Diagnostics[0].Primary.Start {
			t.Errorf("malformed statement %q did not recover with a stable source span: %+v", malformed, result.Diagnostics)
		}
	}
	wrongCondition := []byte("module phase26\nfn main(n: U64) -> U64 {\n  if n { } else { }\n  n\n}\n")
	if checked := session.Check(wrongCondition); len(checked.Diagnostics) == 0 {
		t.Fatal("U64 condition was accepted as truthy")
	}
}

func TestPhase26AppBoundary(t *testing.T) {
	sourcePath := testsupport.ProjectPath("examples", "sum_to_n.schway")
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
