package testsupport_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestCLIOutputContract(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	fixture := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	for _, arguments := range [][]string{
		{"--json", "check", fixture},
		{"check", "--json", fixture},
		{"check", fixture, "--json"},
	} {
		result := testsupport.RunCLI(t, binary, nil, arguments...)
		if result.Exit != 0 || len(result.Stderr) != 0 {
			t.Fatalf("JSON success stream contract failed for %v: exit=%d stdout=%q stderr=%q", arguments, result.Exit, result.Stdout, result.Stderr)
		}
		var decoded protocol.Result
		if err := json.Unmarshal(result.Stdout, &decoded); err != nil {
			t.Fatalf("invalid JSON for %v: %v\n%s", arguments, err, result.Stdout)
		}
		if decoded.Metrics.OutputBytes != len(result.Stdout) || decoded.Metrics.RecomputedWork == 0 {
			t.Fatalf("missing command observations for %v: metrics=%+v actual_bytes=%d", arguments, decoded.Metrics, len(result.Stdout))
		}
		canonical, err := protocol.JSON(decoded)
		if err != nil || !bytes.Equal(canonical, result.Stdout) {
			t.Fatalf("noncanonical JSON for %v:\ngot=%q\nwant=%q err=%v", arguments, result.Stdout, canonical, err)
		}
	}
}

func TestCLIExitTaxonomy(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	valid := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	invalid := testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang")

	cases := []struct {
		name string
		env  []string
		args []string
		exit int
	}{
		{name: "success", args: []string{"check", valid}, exit: 0},
		{name: "invalid source", args: []string{"check", invalid}, exit: 2},
		{name: "operational read", args: []string{"check", filepath.Join(t.TempDir(), "absent.lang")}, exit: 3},
		{name: "operational tool", env: []string{"PATH=" + t.TempDir()}, args: []string{"run", "--engine=native", valid}, exit: 3},
		{name: "usage", args: []string{"unknown"}, exit: 64},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			result := testsupport.RunCLI(t, binary, test.env, test.args...)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%q stderr=%q", result.Exit, test.exit, result.Stdout, result.Stderr)
			}
		})
	}
	if got := protocol.ExitCode(protocol.StatusMismatch); got != 4 {
		t.Fatalf("semantic mismatch exit=%d want=4", got)
	}
}

func TestCLISourceByteLimitBoundary(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	prefix := []byte("module limits.source\nexport { fn keep }\nfn keep(code: Byte) -> Byte { code }\n//")
	for _, test := range []struct {
		name string
		size int
		exit int
		code string
	}{
		{name: "exact", size: 1 << 20, exit: 0},
		{name: "one over", size: 1<<20 + 1, exit: 2, code: "syntax.input_limit"},
	} {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.lang")
			source := append(append([]byte(nil), prefix...), bytes.Repeat([]byte{'x'}, test.size-len(prefix))...)
			if err := os.WriteFile(path, source, 0o600); err != nil {
				t.Fatal(err)
			}
			result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%s", result.Exit, test.exit, result.Stdout)
			}
			if test.code != "" && !bytes.Contains(result.Stdout, []byte(test.code)) {
				t.Fatalf("missing %s: %s", test.code, result.Stdout)
			}
		})
	}
}

func TestCLICheckWorkScalesWithSource(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	work := make([]int, 0, 3)
	for _, count := range []int{1, 10, 100} {
		var source strings.Builder
		source.WriteString("module work.scale\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n")
		for index := 0; index < count; index++ {
			fmt.Fprintf(&source, "  let value%d = code\n", index)
		}
		source.WriteString("  code\n}\n")
		path := filepath.Join(t.TempDir(), fmt.Sprintf("scale-%d.lang", count))
		if err := os.WriteFile(path, []byte(source.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
		var decoded protocol.Result
		if result.Exit != 0 || json.Unmarshal(result.Stdout, &decoded) != nil {
			t.Fatalf("count=%d result=%+v", count, result)
		}
		work = append(work, decoded.Metrics.RecomputedWork)
	}
	if !(work[0] < work[1] && work[1] < work[2]) {
		t.Fatalf("check work does not scale: %v", work)
	}
}

func TestCLICheckRejectsWorkBeyondLimit(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	var source strings.Builder
	source.WriteString("module work.limit\nexport { fn keep }\nfn keep(code: Byte) -> Byte {\n")
	for index := 0; index < 12_000; index++ {
		fmt.Fprintf(&source, "  let value%d = code\n", index)
	}
	source.WriteString("  code\n}\n")
	path := filepath.Join(t.TempDir(), "over-work.lang")
	if err := os.WriteFile(path, []byte(source.String()), 0o600); err != nil {
		t.Fatal(err)
	}
	result := testsupport.RunCLI(t, binary, nil, "--json", "check", path)
	if result.Exit != 2 || !bytes.Contains(result.Stdout, []byte("check.work_limit")) {
		t.Fatalf("work limit result=%+v", result)
	}
}

func TestHumanJSONIdentityParity(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, test := range []struct {
		name string
		args []string
	}{
		{name: "check success", args: []string{"check", testsupport.ProjectPath("testdata", "phase1", "toggle.lang")}},
		{name: "check invalid", args: []string{"check", testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang")}},
		{name: "interpreter", args: []string{"run", "--engine=interpreter", testsupport.ProjectPath("testdata", "phase1", "toggle.lang")}},
	} {
		t.Run(test.name, func(t *testing.T) {
			human := testsupport.RunCLI(t, binary, nil, test.args...)
			jsonArguments := append([]string{"--json"}, test.args...)
			machine := testsupport.RunCLI(t, binary, nil, jsonArguments...)
			if human.Exit != machine.Exit || len(machine.Stderr) != 0 {
				t.Fatalf("projection stream/exit mismatch: human=%+v machine=%+v", human, machine)
			}
			var decoded protocol.Result
			if err := json.Unmarshal(machine.Stdout, &decoded); err != nil {
				t.Fatal(err)
			}
			humanBytes := human.Stdout
			if human.Exit != 0 {
				humanBytes = human.Stderr
				if len(human.Stdout) != 0 {
					t.Fatalf("human failure wrote stdout: %q", human.Stdout)
				}
			}
			if !strings.Contains(string(humanBytes), decoded.ID) {
				t.Fatalf("human projection lacks result ID %q: %q", decoded.ID, humanBytes)
			}
			for _, problem := range decoded.Diagnostics {
				if !strings.Contains(string(humanBytes), problem.ID) {
					t.Fatalf("human projection lacks diagnostic ID %q: %q", problem.ID, humanBytes)
				}
			}
			for _, execution := range decoded.Executions {
				for _, event := range execution.Events {
					if !strings.Contains(string(humanBytes), event.ID) {
						t.Fatalf("human projection lacks event ID %q: %q", event.ID, humanBytes)
					}
				}
			}
		})
	}

	jsonFailure := testsupport.RunCLI(t, binary, os.Environ(), "--json", "check", testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"))
	if jsonFailure.Exit != 2 || len(jsonFailure.Stderr) != 0 || len(jsonFailure.Stdout) == 0 {
		t.Fatalf("JSON failure stream contract: %+v", jsonFailure)
	}
}

func TestHumanJSONMixedDiagnosticVersionParity(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, fixture := range []string{
		testsupport.ProjectPath("testdata", "phase1", "non_exhaustive.lang"),
		testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang"),
	} {
		human := testsupport.RunCLI(t, binary, nil, "check", fixture)
		machine := testsupport.RunCLI(t, binary, nil, "--json", "check", fixture)
		if human.Exit != 2 || machine.Exit != 2 || len(human.Stdout) != 0 || len(machine.Stderr) != 0 {
			t.Fatalf("projection stream/exit mismatch for %s: human=%+v machine=%+v", fixture, human, machine)
		}
		var decoded protocol.Result
		if err := json.Unmarshal(machine.Stdout, &decoded); err != nil {
			t.Fatal(err)
		}
		if decoded.Schema != "lang.command/0" || len(decoded.Diagnostics) != 1 {
			t.Fatalf("command envelope or diagnostic missing: %+v", decoded)
		}
		if !strings.Contains(string(human.Stderr), decoded.ID) || !strings.Contains(string(human.Stderr), decoded.Diagnostics[0].ID) {
			t.Fatalf("human projection lost machine identities: human=%q machine=%+v", human.Stderr, decoded)
		}
	}
}

func TestVerifyCLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	corpus := testsupport.ProjectPath("testdata", "phase1")
	machine := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
	if machine.Exit != 0 || len(machine.Stderr) != 0 {
		t.Fatalf("verify CLI failed: %+v", machine)
	}
	var result protocol.Result
	if err := json.Unmarshal(machine.Stdout, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass || len(result.Lanes) != 5 {
		t.Fatalf("verify result omitted lanes: %+v", result)
	}
	for _, required := range []string{"control:match.non_exhaustive", "control:evidence.source_mismatch", "control:interpreter-o0-o3"} {
		if !bytes.Contains(machine.Stdout, []byte(required)) {
			t.Fatalf("verify JSON omitted %s: %s", required, machine.Stdout)
		}
	}
}

func TestVerifyPhase2CLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	machine := testsupport.RunCLI(t, binary, nil, "--json", "verify", testsupport.ProjectPath("testdata", "phase2"))
	if machine.Exit != 0 || len(machine.Stderr) != 0 {
		t.Fatalf("Phase 2 verify CLI failed: %+v", machine)
	}
	var result protocol.Result
	if err := json.Unmarshal(machine.Stdout, &result); err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusPass || !reflect.DeepEqual(result.ExpectedEscapes, []string{"escape:coordinated-source-core-lie"}) {
		t.Fatalf("Phase 2 verify result omitted expected escape: %+v", result)
	}
	for _, required := range []string{"control:ownership.use_after_move", "control:ownership.move_while_borrowed", "control:ownership.transfer_requires_take", "control:ability.forged_copy", "control:core.duplicate_operation_id", "control:interpreter-o0-o3-owned", "control:evidence.core_mismatch", "control:backend.runtime_causality"} {
		if !bytes.Contains(machine.Stdout, []byte(required)) {
			t.Fatalf("verify JSON omitted %s", required)
		}
	}
}

func TestVerifyCorpusSourceByteLimitBoundary(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	for _, test := range []struct {
		name  string
		phase string
		file  string
		size  int
		exit  int
	}{
		{name: "base exact", phase: "phase1", file: "toggle.lang", size: 1 << 20, exit: 0},
		{name: "base over", phase: "phase1", file: "toggle.lang", size: 1<<20 + 1, exit: 2},
		{name: "owned exact", phase: "phase2", file: "owned_transfer.lang", size: 1 << 20, exit: 0},
		{name: "owned over", phase: "phase2", file: "owned_transfer.lang", size: 1<<20 + 1, exit: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			corpus := t.TempDir()
			sourceCorpus := testsupport.ProjectPath("testdata", test.phase)
			entries, err := os.ReadDir(sourceCorpus)
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".lang") {
					continue
				}
				data, readErr := os.ReadFile(filepath.Join(sourceCorpus, entry.Name()))
				if readErr != nil {
					t.Fatal(readErr)
				}
				if entry.Name() == test.file {
					data = append(data, '/', '/')
					data = append(data, bytes.Repeat([]byte{'x'}, test.size-len(data)-1)...)
					data = append(data, '\n')
				}
				if writeErr := os.WriteFile(filepath.Join(corpus, entry.Name()), data, 0o600); writeErr != nil {
					t.Fatal(writeErr)
				}
			}
			result := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
			if result.Exit != test.exit {
				t.Fatalf("exit=%d want=%d stdout=%s", result.Exit, test.exit, result.Stdout)
			}
			if test.exit != 0 && !bytes.Contains(result.Stdout, []byte("verify.fixture_input_limit")) {
				t.Fatalf("missing bounded fixture diagnostic: %s", result.Stdout)
			}
		})
	}
}

func TestPhase2VerifierScriptContract(t *testing.T) {
	script, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase2.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	if strings.Contains(text, "verify-phase1.sh") || strings.Count(text, "\ngo test ./...\n") != 1 || strings.Count(text, "\ngo test -race ./...\n") != 1 || strings.Count(text, "\ngo vet ./...\n") != 1 {
		t.Fatalf("Phase 2 gate duplicates or nests shared verification:\n%s", text)
	}
	for _, required := range []string{"assert-go-tests.sh --self-test", "TestOwnedBackendMutationIsMismatch", "control:backend.runtime_causality", "verify testdata/phase1", "verify testdata/phase2", "warm_samples=20", "peak_rss=unavailable", "p50_ns=", "p95_ns=", "min_ns=", "max_ns=", "output_bytes=", "work="} {
		if !strings.Contains(text, required) {
			t.Fatalf("Phase 2 gate omitted %q", required)
		}
	}
}
