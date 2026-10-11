package native

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestUnreachableFunctionProgramRuns is the Phase 28 runnable gain (ROADMAP
// SC3, DX-18): programs with functions `main` never calls check, build and
// run instead of failing, and an ambiguous entry is a source refusal (exit 2).
func TestUnreachableFunctionProgramRuns(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}
	binary := testsupport.BuildCLI(t)

	t.Run("unused_function", func(t *testing.T) {
		source := testsupport.ProjectPath("examples", "phase28", "unused_function.schway")
		if result := testsupport.RunCLI(t, binary, nil, "check", source); result.Exit != 0 {
			t.Fatalf("check exit=%d stdout=%s", result.Exit, result.Stdout)
		}
		artifact := filepath.Join(t.TempDir(), "app")
		if result := testsupport.RunCLI(t, binary, nil, "build", source, "--output", artifact); result.Exit != 0 {
			t.Fatalf("build exit=%d stdout=%s stderr=%s", result.Exit, result.Stdout, result.Stderr)
		}
		for _, input := range []string{"7", "0"} {
			result := testsupport.RunCLI(t, binary, nil, "app", "run", artifact, "--", input)
			if result.Exit != 0 || string(result.Stdout) != input+"\n" {
				t.Fatalf("app run %s: exit=%d stdout=%q stderr=%q", input, result.Exit, result.Stdout, result.Stderr)
			}
		}
	})

	// Byte-entry fixtures are not build-supported (build accepts only U64
	// application entries), so only run is asserted for them.
	for _, fixture := range []string{"multi_function_unreachable.schway", "multi_function_zero_call.schway"} {
		for _, engine := range []string{"interpreter", "native"} {
			t.Run(strings.TrimSuffix(fixture, ".schway")+"/"+engine, func(t *testing.T) {
				source := testsupport.ProjectPath("testdata", "phase11", fixture)
				result := testsupport.RunCLI(t, binary, nil, "--json", "run", "--engine="+engine, source)
				decoded := decodePhase28Command(t, result.Stdout)
				if result.Exit != 0 || len(decoded.Executions) != 1 {
					t.Fatalf("exit=%d executions=%d stdout=%s", result.Exit, len(decoded.Executions), result.Stdout)
				}
				if outcome := decoded.Executions[0].Outcome; outcome.Kind != "returned" || outcome.Value != "7" {
					t.Fatalf("outcome = %+v, want returned:7", outcome)
				}
			})
		}
	}

	t.Run("payload_borrow_interaction", func(t *testing.T) {
		source := testsupport.ProjectPath("testdata", "phase12", "payload_borrow_interaction.schway")
		artifact := filepath.Join(t.TempDir(), "app")
		for _, arguments := range [][]string{
			{"--json", "check", source},
			{"--json", "build", source, "--output", artifact},
			{"--json", "run", "--engine=interpreter", source},
			{"--json", "run", "--engine=native", source},
		} {
			result := testsupport.RunCLI(t, binary, nil, arguments...)
			decoded := decodePhase28Command(t, result.Stdout)
			if result.Exit != 2 || len(decoded.Diagnostics) == 0 || decoded.Diagnostics[0].Code != "core.entry_ambiguous" {
				t.Fatalf("%v: exit=%d diagnostics=%+v stdout=%s", arguments, result.Exit, decoded.Diagnostics, result.Stdout)
			}
			if message := decoded.Diagnostics[0].Message; !strings.Contains(message, "choose") || !strings.Contains(message, "identity") || !strings.Contains(message, "add `fn main`") {
				t.Fatalf("%v: ambiguity message %q must list the roots and say to add fn main", arguments, message)
			}
		}
	})
}

type phase28Command struct {
	Diagnostics []struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"diagnostics"`
	Executions []struct {
		Outcome struct {
			Kind  string `json:"kind"`
			Value string `json:"value"`
		} `json:"outcome"`
	} `json:"executions"`
}

func decodePhase28Command(t *testing.T, stdout []byte) phase28Command {
	t.Helper()
	var decoded phase28Command
	if err := json.Unmarshal(stdout, &decoded); err != nil {
		t.Fatalf("decode command result: %v\n%s", err, stdout)
	}
	return decoded
}
