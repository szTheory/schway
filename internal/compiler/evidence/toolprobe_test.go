package evidence

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestDefaultFactsBoundsBothToolProbes(t *testing.T) {
	for _, test := range []struct {
		name     string
		badCall  int
		mode     string
		wantCode string
		timeout  time.Duration
	}{
		{name: "version stdout", badCall: 1, mode: "stdout", wantCode: "evidence.tool_stdout_truncated"},
		{name: "version stderr", badCall: 1, mode: "stderr", wantCode: "evidence.tool_stderr_truncated"},
		{name: "target stdout", badCall: 2, mode: "stdout", wantCode: "evidence.tool_stdout_truncated"},
		{name: "target stderr", badCall: 2, mode: "stderr", wantCode: "evidence.tool_stderr_truncated"},
		{name: "version timeout", badCall: 1, mode: "block", wantCode: "evidence.tool_timeout", timeout: 50 * time.Millisecond},
		{name: "target timeout", badCall: 2, mode: "block", wantCode: "evidence.tool_timeout", timeout: 50 * time.Millisecond},
	} {
		t.Run(test.name, func(t *testing.T) {
			calls := 0
			factory := func(ctx context.Context, _ string, _ ...string) *exec.Cmd {
				calls++
				mode := "normal"
				if calls == test.badCall {
					mode = test.mode
				}
				cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=TestEvidenceToolProbeHelper", "--", mode, string(rune('0'+calls)))
				cmd.Env = append(os.Environ(), "GO_WANT_EVIDENCE_TOOL_HELPER=1")
				return cmd
			}
			ctx := context.Background()
			if test.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, test.timeout)
				defer cancel()
			}
			_, err := defaultFacts(ctx, "fake-clang", factory)
			var probe *ToolProbeError
			if !errors.As(err, &probe) || probe.Code != test.wantCode {
				t.Fatalf("error=%v code=%v want=%s", err, probe, test.wantCode)
			}
		})
	}
}

func TestEvidenceToolProbeHelper(t *testing.T) {
	if os.Getenv("GO_WANT_EVIDENCE_TOOL_HELPER") != "1" {
		return
	}
	mode := os.Args[len(os.Args)-2]
	call := os.Args[len(os.Args)-1]
	switch mode {
	case "stdout":
		_, _ = os.Stdout.WriteString(strings.Repeat("x", MaxToolProbeBytes+1))
	case "stderr":
		_, _ = os.Stderr.WriteString(strings.Repeat("x", MaxToolProbeBytes+1))
	case "block":
		time.Sleep(10 * time.Second)
	default:
		if call == "1" {
			_, _ = os.Stdout.WriteString("clang fixture 1.0\n")
		} else {
			_, _ = os.Stdout.WriteString("fixture-target\n")
		}
	}
	os.Exit(0)
}
