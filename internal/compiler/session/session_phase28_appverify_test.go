package session_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
)

// Phase 28 plan 03 (EVD-13): app verify keeps the named capacity reason at any
// n by reporting an outcome digest instead of embedded interpreter evidence.

func phase28RequireClang(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skip("env:clang toolchain unavailable")
	}
}

func phase28SumToN(n int) string { return strconv.Itoa(n * (n + 1) / 2) }

// phase28VerifyFile drives the file entry point so the written report is part
// of the evidence.
func phase28VerifyFile(t *testing.T, cases session.ReplayCases) (session.ReplayReport, string) {
	t.Helper()
	dir := t.TempDir()
	sourcePath := filepath.Join(dir, "sum_to_n.schway")
	if err := os.WriteFile(sourcePath, phase273SumToNSource(t), 0o600); err != nil {
		t.Fatal(err)
	}
	encodedCases, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	casesPath := filepath.Join(dir, "cases.json")
	if err := os.WriteFile(casesPath, encodedCases, 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(dir, "report.json")
	report, err := session.VerifyApplicationCasesFile(context.Background(), sourcePath, casesPath, reportPath, native.DefaultRunner())
	if err != nil {
		t.Fatalf("exhausted cases must be per-case results, got error: %v", err)
	}
	return report, reportPath
}

func phase28RequireWrittenReport(t *testing.T, reportPath string, wantCases int) {
	t.Helper()
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("report file not written: %v", err)
	}
	if len(data) > session.MaxReplayCasesBytes {
		t.Fatalf("report is %d bytes, over %d", len(data), session.MaxReplayCasesBytes)
	}
	var decoded session.ReplayReport
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Schema != "lang.app-verification/2" || len(decoded.Cases) != wantCases || decoded.Verified {
		t.Fatalf("written report schema=%q cases=%d verified=%v", decoded.Schema, len(decoded.Cases), decoded.Verified)
	}
}

func TestAppVerifyExhaustedCases(t *testing.T) {
	phase28RequireClang(t)
	for _, n := range []int{173} {
		t.Run("n="+strconv.Itoa(n), func(t *testing.T) {
			value := phase28SumToN(n)
			report, reportPath := phase28VerifyFile(t, phase273Cases(phase273Case{"exhausted", strconv.Itoa(n), value}))
			if report.Verified || len(report.Cases) != 1 {
				t.Fatalf("verified=%v cases=%d", report.Verified, len(report.Cases))
			}
			got := report.Cases[0]
			if got.Status != session.ReplayStatusUnsupported || !strings.HasPrefix(got.Diagnostic, "native.event_capacity_exhausted") {
				t.Fatalf("status=%q diagnostic=%q", got.Status, got.Diagnostic)
			}
			phase273RequireDigest(t, got, value, "-O0")
			phase28RequireWrittenReport(t, reportPath, 1)
		})
	}
}
