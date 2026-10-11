package native

import (
	"encoding/json"
	"os"
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

// TestExitClassesUniformAcrossSurfaces pins ROADMAP SC1 (DX-17): one source
// gets the same exit class and refusing code from check, build, run
// --engine=interpreter and run --engine=native on the shipped binary. The
// corevalidate fixture is the existing checker-clean duplicate_function_name
// (D-24); no new .schway file exists for it. The relative rows pin G-28-3
// (relative native run must reach Check and admitProgram; the CLI child
// inherits the test cwd internal/compiler/native, hence the ../../.. prefix).
func TestExitClassesUniformAcrossSurfaces(t *testing.T) {
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}
	binary := testsupport.BuildCLI(t)
	assertUniformRefusal := func(t *testing.T, source, code string) {
		t.Helper()
		artifact := filepath.Join(t.TempDir(), "app")
		surfaces := [][]string{
			{"--json", "check", source},
			{"--json", "build", source, "--output", artifact},
			{"--json", "run", "--engine=interpreter", source},
			{"--json", "run", "--engine=native", source},
		}
		for _, arguments := range surfaces {
			result := testsupport.RunCLI(t, binary, nil, arguments...)
			decoded := decodePhase28Command(t, result.Stdout)
			if result.Exit != 2 || len(decoded.Diagnostics) == 0 || decoded.Diagnostics[0].Code != code {
				t.Fatalf("%v: exit=%d diagnostics=%+v, want exit 2 %s (stdout=%s)", arguments, result.Exit, decoded.Diagnostics, code, result.Stdout)
			}
		}
		if _, err := os.Stat(artifact); !os.IsNotExist(err) {
			t.Fatalf("build left an artifact after a refusal: stat err=%v", err)
		}
	}
	fixtures := []struct {
		name string
		path []string
		code string
	}{
		{"duplicate_function_name", []string{"testdata", "phase07", "duplicate_function_name.schway"}, "core.duplicate_function_id"},
		{"foreign_origin_omitted", []string{"testdata", "phase4", "foreign_origin_omitted.schway"}, "core.foreign_origin_omitted"},
		{"payload_borrow_interaction", []string{"testdata", "phase12", "payload_borrow_interaction.schway"}, "core.entry_ambiguous"},
	}
	for _, fixture := range fixtures {
		t.Run(fixture.name, func(t *testing.T) {
			assertUniformRefusal(t, testsupport.ProjectPath(fixture.path...), fixture.code)
		})
	}
	relativeSource := func(t *testing.T, name string) string {
		t.Helper()
		source := filepath.Join("..", "..", "..", "testdata", "phase07", name)
		if filepath.IsAbs(source) {
			t.Fatalf("source %q must be relative", source)
		}
		if _, err := os.Stat(source); err != nil {
			t.Fatalf("relative fixture unreachable from test cwd: %v", err)
		}
		return source
	}
	t.Run("relative_path/duplicate_function_name", func(t *testing.T) {
		assertUniformRefusal(t, relativeSource(t, "duplicate_function_name.schway"), "core.duplicate_function_id")
	})
	t.Run("relative_path/native_positive_control", func(t *testing.T) {
		source := relativeSource(t, "call_basic.schway")
		result := testsupport.RunCLI(t, binary, nil, "--json", "run", "--engine=native", source)
		decoded := decodePhase28Command(t, result.Stdout)
		if result.Exit != 0 || len(decoded.Diagnostics) != 0 {
			t.Fatalf("exit=%d diagnostics=%+v stdout=%s stderr=%s, want exit 0", result.Exit, decoded.Diagnostics, result.Stdout, result.Stderr)
		}
	})
}
