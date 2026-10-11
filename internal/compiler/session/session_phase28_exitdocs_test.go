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
