package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase19LiteralFrontier(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase19", "literal_tracer.lang"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(source), "let count = 42") {
		t.Fatal("literal_tracer.lang lost its direct numeric let")
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("literal_tracer.lang unexpectedly passed production checking")
	}
	first := checked.Diagnostics[0]
	if first.ID != "diagnostic:b8cbd5b550316bf004395769" || first.Code != "syntax.unexpected_byte" || first.Primary.Start != 97 || first.Primary.End != 98 {
		t.Fatalf("literal_tracer.lang refusal moved: got id=%q code=%q span=%+v", first.ID, first.Code, first.Primary)
	}
}
