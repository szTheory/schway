package testsupport_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
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
