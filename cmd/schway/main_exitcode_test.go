package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestAppRunExitCode(t *testing.T) {
	cases := []struct {
		name    string
		outcome native.RunOutcome
		err     error
		want    int
	}{
		{"child-zero", native.RunOutcome{Kind: native.RunExited, ExitCode: 0}, nil, 0},
		{"child-seven", native.RunOutcome{Kind: native.RunExited, ExitCode: 7}, nil, 7},
		{"child-65-passthrough", native.RunOutcome{Kind: native.RunExited, ExitCode: 65}, nil, 65},
		{"child-69-passthrough", native.RunOutcome{Kind: native.RunExited, ExitCode: 69}, nil, 69},
		{"signal-6", native.RunOutcome{Kind: native.RunSignaled, SignalNumber: 6}, nil, 134},
		{"signal-9", native.RunOutcome{Kind: native.RunSignaled, SignalNumber: 9}, nil, 137},
		{"timeout", native.RunOutcome{Kind: native.RunTimedOut}, nil, 124},
		{"launch-error", native.RunOutcome{Kind: native.RunLaunchError}, nil, 125},
		{"unknown-kind", native.RunOutcome{Kind: native.RunOutcomeKind("bogus")}, nil, 125},
		{"input-too-long", native.RunOutcome{}, &native.ToolError{Code: "native.input_too_long", Err: errors.New("x")}, 65},
		{"input-contains-nul", native.RunOutcome{}, &native.ToolError{Code: "native.input_contains_nul", Err: errors.New("x")}, 65},
		{"other-tool-error", native.RunOutcome{}, &native.ToolError{Code: "native.other", Err: errors.New("x")}, 125},
		{"plain-error", native.RunOutcome{}, errors.New("boom"), 125},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := appRunExitCode(tc.outcome, tc.err); got != tc.want {
				t.Fatalf("appRunExitCode = %d, want %d", got, tc.want)
			}
		})
	}
	t.Run("nonexistent-artifact", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")
		code, _, stderr := captureLangRun(t, []string{"app", "run", missing, "--", "1"})
		if code != 125 {
			t.Fatalf("exit = %d, want 125 (stderr %q)", code, stderr)
		}
		if !strings.HasPrefix(stderr, "schway app run:") {
			t.Fatalf("stderr = %q, want prefix %q", stderr, "schway app run:")
		}
	})
}

// TestAppVerifySourceRefusalExitsInvalid pins Phase 28 D-21: a source refusal
// in app verify exits 2 with the refusing code and still writes the fail
// report, while a capacity-exhausted case stays exit 3.
func TestAppVerifySourceRefusalExitsInvalid(t *testing.T) {
	oneCase := func(input, value string) session.ReplayCases {
		return session.ReplayCases{Schema: session.ReplayCasesSchema, Cases: []session.ReplayCase{
			{ID: "c1", Kind: session.ReplayCaseSource, Input: input, Expected: session.ReplayExpected{Kind: "returned", Value: value}, ForeignOutcomes: []session.ReplayForeignOutcome{}},
		}}
	}
	refusals := []struct {
		name, path, code string
	}{
		{"corevalidate", "testdata/phase07/duplicate_function_name.schway", "core.duplicate_function_id"},
		{"entry_ambiguous", "testdata/phase12/payload_borrow_interaction.schway", "core.entry_ambiguous"},
	}
	for _, tc := range refusals {
		t.Run(tc.name, func(t *testing.T) {
			source, err := os.ReadFile(testsupport.ProjectPath(strings.Split(tc.path, "/")...))
			if err != nil {
				t.Fatal(err)
			}
			code, stderr, reportPath, report := runCLIReplayCasesSource(t, string(source), oneCase("7", "7"))
			if code != exitInvalidSource {
				t.Fatalf("exit = %d, want %d; stderr=%q", code, exitInvalidSource, stderr)
			}
			if want := "schway app verify: source refused: " + tc.code + "\n"; stderr != want {
				t.Fatalf("stderr = %q, want %q", stderr, want)
			}
			if _, err := os.Stat(reportPath); err != nil {
				t.Fatalf("fail report not written: %v", err)
			}
			if report.Verified || report.Status != session.ReplayStatusFail {
				t.Fatalf("report = %+v, want unverified fail", report)
			}
		})
	}
	t.Run("capacity_stays_3", func(t *testing.T) {
		source, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
		if err != nil {
			t.Fatal(err)
		}
		code, stderr, reportPath, _ := runCLIReplayCasesSource(t, string(source), oneCase("15", "120"))
		if code != exitOperational {
			t.Fatalf("exit = %d, want %d; stderr=%q", code, exitOperational, stderr)
		}
		if _, err := os.Stat(reportPath); err != nil {
			t.Fatalf("report not written: %v", err)
		}
	})
}
