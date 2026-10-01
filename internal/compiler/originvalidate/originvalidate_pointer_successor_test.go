package originvalidate_test

import (
	"os"
	"testing"

	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/originvalidate"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase25OriginProgram(t *testing.T, fixture string) core.Program {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase25", fixture))
	if err != nil {
		t.Fatal(err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("%s: unexpected diagnostics: %+v", fixture, checked.Diagnostics)
	}
	return checked.Program
}

func TestPhase25PointerOrigin(t *testing.T) {
	program := phase25OriginProgram(t, "shared_copy_accept.schway")
	program.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"value"}, Access: "shared"}
	if problems := originvalidate.ValidatePublished(program); len(problems) != 0 {
		t.Fatalf("independent origin peer rejected shared read/copy: %+v", problems)
	}
	omitted := phase25OriginProgram(t, "shared_copy_accept.schway")
	problems := originvalidate.ValidatePublished(omitted)
	if len(problems) == 0 || problems[0].Code != "core.origin_omitted" {
		t.Fatalf("origin peer accepted an undeclared returned loan: %+v", problems)
	}

	function := program.Functions[0]
	origins := originvalidate.RecomputeOriginPerReturn(function, originvalidate.BuildCalleeOriginFacts(program))
	if len(origins) != 1 || !origins[0].Derived || origins[0].Access != "shared" || len(origins[0].Paths) != 1 || origins[0].Paths[0] != "value" {
		t.Fatalf("origin peer did not independently derive the returned shared family: %+v", origins)
	}

	mutated := phase25OriginProgram(t, "shared_copy_accept.schway")
	mutated.Functions[0].PublicOrigin = &core.PublicOrigin{Paths: []string{"value"}, Access: "shared"}
	mutated.Functions[0].Linear.Operations[0].Kind = core.OpBorrowExclusive
	problems = originvalidate.ValidatePublished(mutated)
	if len(problems) == 0 || problems[0].Code != "core.origin_access_mismatch" {
		t.Fatalf("origin peer did not reject family/access mutation: %+v", problems)
	}
}
