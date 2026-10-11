package session_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 27.3 app-verify capacity tests (D-27.3-06 .. D-27.3-10). examples/
// sum_to_n.schway records 15 events for n = 14 and 16 for n = 15 against the
// conformance recorder capacity of 15, so n = 15 exhausts the recorder.

type phase273Case struct{ id, input, value string }

func phase273Cases(entries ...phase273Case) session.ReplayCases {
	cases := session.ReplayCases{Schema: session.ReplayCasesSchema}
	for _, entry := range entries {
		cases.Cases = append(cases.Cases, session.ReplayCase{
			ID: entry.id, Kind: session.ReplayCaseSource, Input: entry.input,
			Expected:        session.ReplayExpected{Kind: "returned", Value: entry.value},
			ForeignOutcomes: []session.ReplayForeignOutcome{},
		})
	}
	return cases
}

func phase273SumToNSource(t *testing.T) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "sum_to_n.schway"))
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func phase273SumToNConformanceC(t *testing.T) string {
	t.Helper()
	checked := session.Check(phase273SumToNSource(t))
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("sum_to_n check diagnostics: %+v", checked.Diagnostics)
	}
	c, err := session.Phase16ControlNativeC(checked.Program, "phase273")
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func phase273RequireUnsupportedCapacity(t *testing.T, caseReport session.ReplayCaseReport) {
	t.Helper()
	if caseReport.Status != session.ReplayStatusUnsupported {
		t.Fatalf("case %q status=%q diagnostic=%q, want %q", caseReport.ID, caseReport.Status, caseReport.Diagnostic, session.ReplayStatusUnsupported)
	}
	if !strings.HasPrefix(caseReport.Diagnostic, "native.event_capacity_exhausted: ") {
		t.Fatalf("case %q diagnostic %q lacks the native.event_capacity_exhausted prefix", caseReport.ID, caseReport.Diagnostic)
	}
	for _, want := range []string{"at -O0", "not a mismatch"} {
		if !strings.Contains(caseReport.Diagnostic, want) {
			t.Fatalf("case %q diagnostic %q lacks %q", caseReport.ID, caseReport.Diagnostic, want)
		}
	}
	phase273RequireDigest(t, caseReport, "120", "-O0")
}

// phase273RequireDigest asserts the D-13 outcome digest shape: no embedded
// engines map, an interpreter digest carrying the expected outcome.
func phase273RequireDigest(t *testing.T, caseReport session.ReplayCaseReport, value, level string) {
	t.Helper()
	if len(caseReport.Engines) != 0 {
		t.Fatalf("case %q engines=%v, want none (digest replaces evidence)", caseReport.ID, caseReport.Engines)
	}
	digest := caseReport.OutcomeDigest
	if digest == nil || digest.Engine != "interpreter" || digest.OutcomeKind != "returned" || digest.OutcomeValue != value {
		t.Fatalf("case %q digest=%+v, want interpreter returned %s", caseReport.ID, digest, value)
	}
	if digest.EvidenceBytes <= 0 || len(digest.EvidenceSHA256) != 64 || digest.ExhaustedLevel != level {
		t.Fatalf("case %q digest=%+v, want evidence bytes, 64-hex digest, level %s", caseReport.ID, digest, level)
	}
}

func phase273RequireNoSignalText(t *testing.T, report session.ReplayReport) {
	t.Helper()
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"signal", "abort", "run_signaled"} {
		if strings.Contains(string(encoded), forbidden) {
			t.Fatalf("marshalled report contains %q: %s", forbidden, encoded)
		}
	}
}

func TestPhase273RealCapacityExit(t *testing.T) {
	cSource := phase273SumToNConformanceC(t)
	for _, level := range []string{"-O0", "-O3"} {
		t.Run(level, func(t *testing.T) {
			runner := native.DefaultRunner()
			runner.Expect = native.ExpectValue
			result, err := runner.Run(context.Background(), cSource, level, []string{"14"})
			if err != nil {
				t.Fatalf("n = 14 at %s must stay within capacity, got %v", level, err)
			}
			if len(result.Pairs) != 1 || result.Pairs[0].Execution.Outcome.Kind != "returned" || result.Pairs[0].Execution.Outcome.Value != "105" {
				t.Fatalf("n = 14 at %s result=%+v, want returned 105", level, result.Pairs)
			}
			_, err = runner.Run(context.Background(), cSource, level, []string{"15"})
			var toolError *native.ToolError
			if !errors.As(err, &toolError) || toolError.Code != "native.event_capacity_exhausted" {
				t.Fatalf("n = 15 at %s error=%v, want ToolError native.event_capacity_exhausted", level, err)
			}
		})
	}
}

func TestPhase273AppVerifySumToNWithinCapacity(t *testing.T) {
	report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
		phase273Cases(phase273Case{"n14", "14", "105"}), native.DefaultRunner())
	if err != nil {
		t.Fatalf("within-capacity batch returned error: %v", err)
	}
	if !report.Verified || len(report.Cases) != 1 || report.Cases[0].Status != session.ReplayStatusPass || len(report.Cases[0].Engines) != 3 {
		t.Fatalf("within-capacity report=%+v", report)
	}
}

func TestPhase273AppVerifySumToNExhausted(t *testing.T) {
	report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
		phase273Cases(phase273Case{"n15", "15", "120"}), native.DefaultRunner())
	if err != nil {
		t.Fatalf("capacity exhaustion must be a per-case result, got error: %v", err)
	}
	if report.Verified || len(report.Cases) != 1 {
		t.Fatalf("exhausted report verified=%v cases=%d", report.Verified, len(report.Cases))
	}
	phase273RequireUnsupportedCapacity(t, report.Cases[0])
	phase273RequireNoSignalText(t, report)
}

func TestPhase273AppVerifyMixedCases(t *testing.T) {
	report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
		phase273Cases(phase273Case{"n14", "14", "105"}, phase273Case{"n15", "15", "120"}), native.DefaultRunner())
	if err != nil {
		t.Fatalf("mixed batch must report per case, got error: %v", err)
	}
	if report.Verified || len(report.Cases) != 2 {
		t.Fatalf("mixed report verified=%v cases=%d", report.Verified, len(report.Cases))
	}
	if report.Cases[0].Status != session.ReplayStatusPass || len(report.Cases[0].Engines) != 3 {
		t.Fatalf("n = 14 case=%+v, want pass with three engines", report.Cases[0])
	}
	phase273RequireUnsupportedCapacity(t, report.Cases[1])
	phase273RequireNoSignalText(t, report)
}

func TestPhase273AppVerifyExhaustedWrongExpected(t *testing.T) {
	report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
		phase273Cases(phase273Case{"n15-wrong", "15", "121"}), native.DefaultRunner())
	if err != nil {
		t.Fatalf("wrong expected at capacity must be a per-case result, got error: %v", err)
	}
	if report.Verified || len(report.Cases) != 1 {
		t.Fatalf("wrong-expected report verified=%v cases=%d", report.Verified, len(report.Cases))
	}
	caseReport := report.Cases[0]
	if caseReport.Status != session.ReplayStatusFail {
		t.Fatalf("status=%q diagnostic=%q, want fail", caseReport.Status, caseReport.Diagnostic)
	}
	if !strings.Contains(caseReport.Diagnostic, "interpreter produced returned:120, expected returned:121") {
		t.Fatalf("diagnostic %q lacks the comparison text", caseReport.Diagnostic)
	}
	if strings.HasPrefix(caseReport.Diagnostic, "native.event_capacity_exhausted") {
		t.Fatalf("a mismatch must not be masked as capacity: %q", caseReport.Diagnostic)
	}
}

func TestPhase273AppVerifyNonCapacityErrorStaysOperational(t *testing.T) {
	requireOperational := func(t *testing.T, report session.ReplayReport, err error, wantErr string) {
		t.Helper()
		if err == nil || !strings.Contains(err.Error(), wantErr) {
			t.Fatalf("error=%v, want text containing %q", err, wantErr)
		}
		for _, caseReport := range report.Cases {
			if caseReport.Status != session.ReplayStatusFail {
				t.Fatalf("case %q status=%q, want fail", caseReport.ID, caseReport.Status)
			}
			if strings.HasPrefix(caseReport.Diagnostic, "native.event_capacity_exhausted") {
				t.Fatalf("non-capacity error reported as capacity: %q", caseReport.Diagnostic)
			}
		}
		if len(report.Cases) == 0 {
			t.Fatal("report has no cases")
		}
	}
	t.Run("missing-clang", func(t *testing.T) {
		report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
			phase273Cases(phase273Case{"n14", "14", "105"}), phase271MissingClangRunner(t))
		requireOperational(t, report, err, "run generated C at -O0")
		if report.Cases[0].Diagnostic != "-O0 generated-C execution failed" {
			t.Fatalf("diagnostic=%q, want -O0 generated-C execution failed", report.Cases[0].Diagnostic)
		}
	})
	t.Run("native-timeout", func(t *testing.T) {
		runner := native.Runner{ClangPath: "clang", Timeout: time.Nanosecond}
		report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
			phase273Cases(phase273Case{"n15", "15", "120"}), runner)
		requireOperational(t, report, err, "native.timeout")
	})
}

// D-27.3-10 (amended): `schway run --engine=native` cannot reach a U64 loop
// program today (interpreterInputs returns nil for U64 entries), so "run-native
// on a loop program past capacity reports the named code" is proven in two
// pieces: TestPhase273RealCapacityExit (real sum_to_n conformance C through the
// runner yields the named code) and this test (a runner returning that code
// surfaces it through RunNativeCommandFile). No CLI-level loop repro is claimed.
func TestPhase273RunNativeSurfacesNamedCode(t *testing.T) {
	runner := &fakeNativeRunner{err: &native.ToolError{
		Code: "native.event_capacity_exhausted",
		Err:  errors.New("generated C exceeded its static event capacity at -O0"),
	}}
	result, err := session.RunNativeCommandFile(context.Background(), testsupport.ProjectPath("testdata", "phase19", "literal_tracer.schway"), runner)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != protocol.StatusOperational || len(result.Diagnostics) == 0 || result.Diagnostics[0].Code != "native.event_capacity_exhausted" {
		t.Fatalf("run-native result status=%q diagnostics=%+v, want operational native.event_capacity_exhausted", result.Status, result.Diagnostics)
	}
}

// TestPhase273AppVerifyOneLevelExhaustionFails pins the full WR-01 contract
// in both directions: when exactly one optimization level exhausts the event
// recorder, the case is fail (never unsupported), carries only interpreter
// evidence, names the exhausting level exactly, and the report is unverified.
// A clang wrapper shrinks SCHWAY_EVENT_CAPACITY to 1 at one level only.
func TestPhase273AppVerifyOneLevelExhaustionFails(t *testing.T) {
	realClang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("env:clang toolchain unavailable, skipping asymmetric exhaustion exercise")
	}
	for _, level := range []string{"-O0", "-O3"} {
		t.Run(level+"-only", func(t *testing.T) {
			wrapper := filepath.Join(t.TempDir(), "clang-shrink")
			script := "#!/bin/sh\n" +
				"hit=0; src=\n" +
				"for a in \"$@\"; do case \"$a\" in " + level + ") hit=1;; *.c) src=\"$a\";; esac; done\n" +
				"if [ \"$hit\" = 1 ] && [ -n \"$src\" ]; then sed -i.bak 's/#define SCHWAY_EVENT_CAPACITY [0-9]*u/#define SCHWAY_EVENT_CAPACITY 1u/' \"$src\"; fi\n" +
				"exec " + realClang + " \"$@\"\n"
			if err := os.WriteFile(wrapper, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}
			runner := native.DefaultRunner()
			runner.ClangPath = wrapper
			report, err := session.VerifyApplicationCases(context.Background(), phase273SumToNSource(t),
				phase273Cases(phase273Case{"n14", "14", "105"}), runner)
			if err != nil {
				t.Fatalf("asymmetric exhaustion at %s must be a per-case result, got error: %v", level, err)
			}
			if report.Verified {
				t.Fatalf("report verified with asymmetric exhaustion at %s: %+v", level, report)
			}
			if len(report.Cases) != 1 {
				t.Fatalf("report cases=%d, want 1", len(report.Cases))
			}
			got := report.Cases[0]
			if got.Status != session.ReplayStatusFail {
				t.Fatalf("status=%q diagnostic=%q, want fail (never unsupported)", got.Status, got.Diagnostic)
			}
			want := "event capacity exhausted at " + level + " only; the other optimization level completed, so the levels diverge"
			if got.Diagnostic != want {
				t.Fatalf("diagnostic=%q, want %q", got.Diagnostic, want)
			}
			phase273RequireDigest(t, got, "105", level)
			phase273RequireNoSignalText(t, report)
		})
	}
}
