package session_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34SourceRoutePresentation(t *testing.T) {
	source := []byte("module draft.route\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n")
	path := filepath.Join(t.TempDir(), "route.schway")
	if err := os.WriteFile(path, source, 0o600); err != nil {
		t.Fatal(err)
	}

	reads := map[string]int{}
	restore := modload.SetReadObserverForTest(func(module string) { reads[module]++ })
	defer restore()

	checkOnce := func(route string) {
		t.Helper()
		if reads[""] != 1 || len(reads) != 1 {
			t.Fatalf("%s source reads = %v, want one entry snapshot", route, reads)
		}
		clear(reads)
	}
	assertCaptured := func(route string, result protocol.Result) {
		t.Helper()
		if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 || result.Presentation == nil {
			t.Fatalf("%s result = %+v, want invalid diagnostic with captured source", route, result)
		}
		captured, ok := result.Presentation.Sources[""]
		if !ok || captured.Label != path || !bytes.Equal(captured.Bytes, source) {
			t.Fatalf("%s presentation = %+v, want the one loaded source snapshot", route, result.Presentation)
		}
		human, err := protocol.Human(result)
		if err != nil {
			t.Fatal(err)
		}
		problem := result.Diagnostics[0]
		if problem.Primary.Line != 0 || problem.Primary.Col != 0 || !strings.Contains(human, path+":3:") || !strings.Contains(human, problem.ID) || !strings.Contains(human, problem.Code) || !strings.Contains(human, "n + }") || !strings.Contains(human, "^") {
			t.Fatalf("%s diagnostic lost its captured source attribution: %+v\n%s", route, problem, human)
		}
	}

	interpreter, err := session.RunInterpreterCommandFile(path)
	if err != nil {
		t.Fatal(err)
	}
	assertCaptured("interpreter", interpreter)
	checkOnce("interpreter")

	runner := &phase34RouteCountingRunner{}
	nativeResult, err := session.RunNativeCommandFile(context.Background(), path, runner)
	if err != nil {
		t.Fatal(err)
	}
	assertCaptured("native", nativeResult)
	if runner.calls != 0 {
		t.Fatalf("native runner received %d calls for rejected syntax, want zero", runner.calls)
	}
	checkOnce("native")

	// BuildApplicationFile is the retained public API. Its source-refusal
	// result remains unchanged while the CLI command gains presentation.
	_, diagnostics, err := session.BuildApplicationFile(context.Background(), path, filepath.Join(t.TempDir(), "app"), native.Runner{})
	if err != nil || len(diagnostics) == 0 || diagnostics[0].Code != interpreter.Diagnostics[0].Code {
		t.Fatalf("BuildApplicationFile refusal = diagnostics %v, err %v; want the same source refusal", diagnostics, err)
	}
	checkOnce("build")
}

type phase34RouteCountingRunner struct{ calls int }

func (runner *phase34RouteCountingRunner) Run(context.Context, string, string, []string) (native.Result, error) {
	runner.calls++
	return native.Result{}, nil
}
