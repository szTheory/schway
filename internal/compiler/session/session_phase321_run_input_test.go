package session_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
)

// phase321Copy copies a repository fixture into t.TempDir() (D-26: no new
// corpus file) and returns the copy's path.
func phase321Copy(t *testing.T, relative string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", relative))
	if err != nil {
		t.Fatalf("read fixture %s: %v", relative, err)
	}
	path := filepath.Join(t.TempDir(), "fixture_main.schway")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// TestPhase321RunInputRequired pins D-01..D-06: run on an entry it cannot bind
// refuses as usage (64) before any engine work, identically on both engines,
// without echoing path or parameter, and the D-05 controls stay put.
func TestPhase321RunInputRequired(t *testing.T) {
	path := phase321Copy(t, "examples/sum_to_n.schway")
	counting := &countingNativeRunner{}

	interpreted, err := session.RunInterpreterCommandFile(path)
	if err != nil {
		t.Fatal(err)
	}
	native, err := session.RunNativeCommandFile(context.Background(), path, counting)
	if err != nil {
		t.Fatal(err)
	}
	for name, result := range map[string]protocol.Result{"interpreter": interpreted, "native": native} {
		if result.Status != protocol.StatusUsage || protocol.ExitCode(result.Status) != 64 {
			t.Fatalf("%s: want usage/64, got status %q exit %d", name, result.Status, protocol.ExitCode(result.Status))
		}
		if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "tool.run_input_required" {
			t.Fatalf("%s: want one tool.run_input_required, got %+v", name, result.Diagnostics)
		}
		if result.Diagnostics[0].Primary != (diagnostic.Span{}) || result.Metrics.RecomputedWork != 1 {
			t.Fatalf("%s: span/work mismatch %+v", name, result)
		}
	}
	if counting.calls != 0 {
		t.Fatalf("native runner invoked %d times; refusal must precede engine work", counting.calls)
	}
	interpretedJSON, err := protocol.JSON(interpreted)
	if err != nil {
		t.Fatal(err)
	}
	nativeJSON, err := protocol.JSON(native)
	if err != nil {
		t.Fatal(err)
	}
	if string(interpretedJSON) != string(nativeJSON) {
		t.Fatalf("engines disagree:\n%s\n%s", interpretedJSON, nativeJSON)
	}
	rendered := string(interpretedJSON)
	if strings.Contains(rendered, filepath.Dir(path)) || strings.Contains(rendered, "n:") {
		t.Fatalf("refusal echoed path or parameter: %s", rendered)
	}

	// Armed control A: a bound main still runs, and the counter counts.
	bound := phase321Copy(t, "testdata/phase07/call_basic.schway")
	ran, err := session.RunInterpreterCommandFile(bound)
	if err != nil || ran.Status != protocol.StatusPass {
		t.Fatalf("bound main must still run on the interpreter: %v %+v", err, ran)
	}
	armed := &countingNativeRunner{}
	if _, err := session.RunNativeCommandFile(context.Background(), bound, armed); err != nil {
		t.Fatal(err)
	}
	if armed.calls < 1 {
		t.Fatalf("counter is vacuous: bound main never reached the native runner")
	}

	// Armed control B: source refusal (exit 2) wins over the run-input refusal.
	refused := phase321Copy(t, "testdata/phase3/public_view_omitted.schway")
	for name, run := range map[string]func() (protocol.Result, error){
		"interpreter": func() (protocol.Result, error) { return session.RunInterpreterCommandFile(refused) },
		"native": func() (protocol.Result, error) {
			return session.RunNativeCommandFile(context.Background(), refused, &countingNativeRunner{})
		},
	} {
		result, err := run()
		if err != nil {
			t.Fatal(err)
		}
		if result.Status != protocol.StatusInvalid || len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "core.origin_omitted" {
			t.Fatalf("%s: source refusal must win, got %+v", name, result)
		}
	}
}
