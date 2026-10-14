package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/session"
)

func TestPhase34SourceRoutePresentation(t *testing.T) {
	source := "module draft.route\nexport { fn main }\nfn main(n: U64) -> U64 { n + }\n"
	directory := t.TempDir()
	sourcePath := filepath.Join(directory, "route.schway")
	if err := os.WriteFile(sourcePath, []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}

	code, jsonOutput, stderr := captureLangRun(t, []string{"--json", "check", sourcePath})
	if code != exitcode.Invalid || stderr != "" {
		t.Fatalf("JSON check code=%d stderr=%q stdout=%q", code, stderr, jsonOutput)
	}
	var checkResult struct {
		Diagnostics []struct {
			ID      string `json:"id"`
			Code    string `json:"code"`
			Primary struct {
				Line int `json:"line"`
				Col  int `json:"col"`
			} `json:"primary_span"`
		} `json:"diagnostics"`
	}
	if err := json.Unmarshal([]byte(jsonOutput), &checkResult); err != nil || len(checkResult.Diagnostics) == 0 {
		t.Fatalf("decode check diagnostic: err=%v output=%q", err, jsonOutput)
	}
	want := checkResult.Diagnostics[0]
	if want.ID == "" || want.Code == "" || want.Primary.Line != 3 || want.Primary.Col < 1 {
		t.Fatalf("check diagnostic=%+v, want a stable located diagnostic", want)
	}
	var stdout string

	routes := []struct {
		name string
		args []string
	}{
		{name: "interpreter", args: []string{"run", "--engine=interpreter", sourcePath}},
		{name: "native", args: []string{"run", "--engine=native", sourcePath}},
		{name: "build", args: []string{"build", sourcePath, "--output", filepath.Join(directory, "app")}},
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			code, stdout, stderr := captureLangRun(t, route.args)
			assertPhase34RouteText(t, route.name, code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)

			jsonArgs := append([]string{"--json"}, route.args...)
			jsonCode, jsonStdout, jsonStderr := captureLangRun(t, jsonArgs)
			if jsonCode != exitcode.Invalid || jsonStderr != "" {
				t.Fatalf("JSON %s code=%d stderr=%q stdout=%q", route.name, jsonCode, jsonStderr, jsonStdout)
			}
			var result struct {
				Diagnostics []struct {
					ID      string `json:"id"`
					Code    string `json:"code"`
					Primary struct {
						Line int `json:"line"`
						Col  int `json:"col"`
					} `json:"primary_span"`
				} `json:"diagnostics"`
			}
			if err := json.Unmarshal([]byte(jsonStdout), &result); err != nil || len(result.Diagnostics) == 0 {
				t.Fatalf("decode JSON %s diagnostic: err=%v output=%q", route.name, err, jsonStdout)
			}
			got := result.Diagnostics[0]
			if got.ID != want.ID || got.Code != want.Code || got.Primary.Line != 3 || got.Primary.Col != want.Primary.Col {
				t.Fatalf("JSON %s diagnostic=%+v, want check identity/location %+v", route.name, got, want)
			}
		})
	}

	casesPath := filepath.Join(directory, "cases.json")
	if err := os.WriteFile(casesPath, []byte(`{"schema":"lang.replay-cases/1","cases":[{"id":"bad","kind":"source","input":"7","expected":{"kind":"returned","value":"7"},"foreign_outcomes":[]}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(directory, "report.json")
	if err := os.WriteFile(reportPath, []byte("stale report"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, stdout, stderr = captureLangRun(t, []string{"app", "verify", sourcePath, "--cases", casesPath, "--report", reportPath})
	assertPhase34RouteText(t, "app verify", code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)
	reportBytes, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("source refusal did not replace the stale report: %v", err)
	}
	var report session.ReplayReport
	if err := json.Unmarshal(reportBytes, &report); err != nil || report.Verified || report.Status != session.ReplayStatusFail {
		t.Fatalf("app verify refusal report=%+v err=%v", report, err)
	}

	// A failed report publish must keep the source refusal visible, retain the
	// established invalid-source exit, and include the report failure.
	missingParentReport := filepath.Join(directory, "missing", "report.json")
	code, stdout, stderr = captureLangRun(t, []string{"app", "verify", sourcePath, "--cases", casesPath, "--report", missingParentReport})
	assertPhase34RouteText(t, "app verify report-write failure", code, stdout, stderr, sourcePath, want.ID, want.Code, want.Primary.Col)
	if !strings.Contains(stderr, "verification report") {
		t.Fatalf("report-write failure was lost: stderr=%q", stderr)
	}
}

func assertPhase34RouteText(t *testing.T, route string, code int, stdout, stderr, path, id, diagnosticCode string, column int) {
	t.Helper()
	if code != exitcode.Invalid || stdout != "" {
		t.Fatalf("%s code=%d stdout=%q stderr=%q, want invalid source on stderr", route, code, stdout, stderr)
	}
	for _, part := range []string{fmt.Sprintf("%s:3:%d", path, column), id, diagnosticCode, "n + }", "^"} {
		if !strings.Contains(stderr, part) {
			t.Errorf("%s stderr lacks %q: %q", route, part, stderr)
		}
	}
}
