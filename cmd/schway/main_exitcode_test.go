package main

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
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
