package session_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/cgen"
	"github.com/codename-lang/lang/internal/compiler/core"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func assertPhase18ComputedFrontier(t *testing.T, fixture string, sourceNeedles ...string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", fixture))
	if err != nil {
		t.Fatalf("read %s: %v", fixture, err)
	}
	for _, needle := range sourceNeedles {
		if !strings.Contains(string(source), needle) {
			t.Fatalf("fixture %s is missing source witness %q", fixture, needle)
		}
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %s terminal match did not parse: %+v", fixture, parsed.Diagnostics)
	}
	if len(parsed.Program.Funcs) == 0 || parsed.Program.Funcs[len(parsed.Program.Funcs)-1].Body.Linear == nil || parsed.Program.Funcs[len(parsed.Program.Funcs)-1].Body.Linear.TerminalMatch == nil {
		t.Fatalf("fixture %s does not preserve its terminal computed match", fixture)
	}
}

func TestPhase18ResultFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "result_computed_match.lang", "let result = identity(input)", "match result")
}

func TestPhase18PayloadFixtureFrontier(t *testing.T) {
	assertPhase18ComputedFrontier(t, "payload_return.lang", "let result = identity(value)", "Ok(payload) => payload", "match result")
}

func TestPhase18ComputedSourceFourTierDifferential(t *testing.T) {
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase18", "computed_match.lang"))
	if err != nil {
		t.Fatalf("read computed-match fixture: %v", err)
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("check computed-match fixture: %+v", checked.Diagnostics)
	}
	var function *core.Function
	for index := range checked.Program.Functions {
		if checked.Program.Functions[index].Name == "select" {
			function = &checked.Program.Functions[index]
			break
		}
	}
	if function == nil || function.Match == nil || function.Linear == nil || function.Match.Scrutinee != "computed" {
		t.Fatalf("source did not produce a checked computed branch: %+v", checked.Program.Functions)
	}
	if peer := corevalidate.Validate(checked.Program); !peer.Valid {
		t.Fatalf("independent core admission: %+v", peer.Problems)
	}
	for _, input := range []string{"Left", "Right"} {
		engines := phase11RunFourTiersWithSupplier(t, context.Background(), checked.Program, "select", input, cgen.EmitProgramNativeForTest, "emitProgram")
		engines["interpreter"] = phase16ProjectInterpreterSchema2(t, checked.Program, engines["interpreter"])
		if len(engines) != 4 {
			t.Fatalf("input %q produced %d execution engines, want interpreter, O0, O3, and O3-LTO", input, len(engines))
		}
		if err := session.Phase5CompareProgramEngines("testdata/phase18/computed_match.lang:"+input, checked.Program, engines); err != nil {
			t.Fatalf("input %q five-axis comparison: %v", input, err)
		}
	}
}
