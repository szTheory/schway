package session_test

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/exitcode"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

var exitDocsRow = regexp.MustCompile("(?m)^\\| `([A-Za-z]+)` \\| ([0-9]+) \\|")

func exitCodeConstants() map[string]int {
	return map[string]int{
		"Success":         exitcode.Success,
		"Unrepaired":      exitcode.Unrepaired,
		"Invalid":         exitcode.Invalid,
		"Operational":     exitcode.Operational,
		"Mismatch":        exitcode.Mismatch,
		"Usage":           exitcode.Usage,
		"AppInputRefused": exitcode.AppInputRefused,
		"EventCapacity":   exitcode.EventCapacity,
		"AppTimeout":      exitcode.AppTimeout,
		"AppLaunchError":  exitcode.AppLaunchError,
		"SignalBase":      exitcode.SignalBase,
	}
}

// exitCodeDocsFindings compares the README "Exit codes" section to the
// constants and returns every disagreement.
func exitCodeDocsFindings(readme string) []string {
	const heading = "## Exit codes"
	start := strings.Index(readme, heading)
	if start < 0 {
		return []string{"README.md has no " + heading + " section"}
	}
	section := readme[start+len(heading):]
	if next := strings.Index(section, "\n## "); next >= 0 {
		section = section[:next]
	}
	want := exitCodeConstants()
	seen := map[string]int{}
	var findings []string
	for _, m := range exitDocsRow.FindAllStringSubmatch(section, -1) {
		name := m[1]
		value, _ := strconv.Atoi(m[2])
		seen[name]++
		constant, known := want[name]
		switch {
		case !known:
			findings = append(findings, fmt.Sprintf("unknown exit-code row %q", name))
		case constant != value:
			findings = append(findings, fmt.Sprintf("row %s = %d, constant = %d", name, value, constant))
		}
	}
	for name := range want {
		if seen[name] != 1 {
			findings = append(findings, fmt.Sprintf("constant %s appears %d times, want exactly 1", name, seen[name]))
		}
	}
	if total := len(exitDocsRow.FindAllString(section, -1)); total < len(want) {
		findings = append(findings, fmt.Sprintf("only %d rows parsed", total))
	}
	return findings
}

func TestExitCodeDocsTable(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath("README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(data)
	for _, finding := range exitCodeDocsFindings(readme) {
		t.Error(finding)
	}
	t.Run("negative control", func(t *testing.T) {
		wrong := strings.Replace(readme, "| `Invalid` | 2 |", "| `Invalid` | 9 |", 1)
		if wrong == readme {
			t.Fatal("control did not alter the section")
		}
		if findings := exitCodeDocsFindings(wrong); len(findings) == 0 {
			t.Fatal("a wrong row value was not reported")
		}
		missing := strings.Replace(readme, "| `Mismatch` | 4 |", "| `Mismatched` | 4 |", 1)
		if findings := exitCodeDocsFindings(missing); len(findings) < 2 {
			t.Fatalf("renamed row findings = %v, want unknown row and missing constant", findings)
		}
	})
}

func phase37FileRunExitContractProblems(readme, routeTests string) []string {
	const statementsHeading = "### Source-file `run` routes"
	start := strings.Index(readme, statementsHeading)
	if start < 0 {
		return []string{"README.md has no " + statementsHeading + " section"}
	}
	section := readme[start+len(statementsHeading):]
	if next := strings.Index(section, "\n### "); next >= 0 {
		section = section[:next]
	}
	want := []string{
		"`schway run --engine=interpreter SOURCE -- PATH` supplies PATH to the interpreter's FileBytes reads.",
		"`schway run --engine=native SOURCE --manifest MANIFEST -- PATH` supplies PATH to the native application's FileBytes reads.",
		"The no-tail `run --engine=interpreter|native FILE` form remains synthetic conformance and supplies no application input; use `build` plus `app run` for an artifact workflow.",
		"An empty `PATH` is refused before file access with `(65, \"\", \"\")`; `TestPhase37FileRunInputBoundary` pins this tuple.",
		"The file route keeps source refusal at 2, tool/build failures at 3, bad usage at 64, and typed file/input refusal at 65.",
	}
	var findings []string
	for _, statement := range want {
		if count := strings.Count(section, statement); count != 1 {
			findings = append(findings, fmt.Sprintf("statement %q appears %d times, want exactly 1", statement, count))
		}
	}
	if !strings.Contains(routeTests, "func TestPhase37FileRunInputBoundary(") || !strings.Contains(routeTests, `name: "empty", path: "", wantCode: 65, wantStderr: ""`) {
		findings = append(findings, "README empty-path tuple is not backed by the public input-boundary test")
	}
	return findings
}

func TestPhase37FileRunExitContract(t *testing.T) {
	readmeBytes, err := os.ReadFile(testsupport.ProjectPath("README.md"))
	if err != nil {
		t.Fatal(err)
	}
	routeTestsBytes, err := os.ReadFile(testsupport.ProjectPath("cmd", "schway", "main_phase37_file_run_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	readme, routeTests := string(readmeBytes), string(routeTestsBytes)
	for _, finding := range phase37FileRunExitContractProblems(readme, routeTests) {
		t.Error(finding)
	}
	t.Run("negative controls", func(t *testing.T) {
		const emptyTuple = "An empty `PATH` is refused before file access with `(65, \"\", \"\")`; `TestPhase37FileRunInputBoundary` pins this tuple."
		fixture := "### Source-file `run` routes\n\n" + strings.Join([]string{
			"`schway run --engine=interpreter SOURCE -- PATH` supplies PATH to the interpreter's FileBytes reads.",
			"`schway run --engine=native SOURCE --manifest MANIFEST -- PATH` supplies PATH to the native application's FileBytes reads.",
			"The no-tail `run --engine=interpreter|native FILE` form remains synthetic conformance and supplies no application input; use `build` plus `app run` for an artifact workflow.",
			emptyTuple,
			"The file route keeps source refusal at 2, tool/build failures at 3, bad usage at 64, and typed file/input refusal at 65.",
		}, "\n\n")
		if findings := phase37FileRunExitContractProblems(fixture, routeTests); len(findings) != 0 {
			t.Fatalf("valid control fixture findings = %v", findings)
		}
		missing := strings.Replace(fixture, emptyTuple, "", 1)
		if missing == fixture || len(phase37FileRunExitContractProblems(missing, routeTests)) == 0 {
			t.Fatal("missing empty-path statement was not reported")
		}
		duplicate := strings.Replace(fixture, emptyTuple, emptyTuple+"\n\n"+emptyTuple, 1)
		if duplicate == fixture || len(phase37FileRunExitContractProblems(duplicate, routeTests)) == 0 {
			t.Fatal("duplicate empty-path statement was not reported")
		}
		wrong := strings.Replace(fixture, "typed file/input refusal at 65.", "typed file/input refusal at 64.", 1)
		if wrong == fixture || len(phase37FileRunExitContractProblems(wrong, routeTests)) == 0 {
			t.Fatal("wrong exit meaning was not reported")
		}
	})
}
