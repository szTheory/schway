package session_test

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32HostedCIEnvProblems reports missing hosted-CI env pins (D-00, D-04a).
func phase32HostedCIEnvProblems(workflow string) []string {
	var problems []string
	cs := strings.Index(workflow, "\n  checks:\n")
	es := strings.Index(workflow, "\n  evidence-aggregate:\n")
	pe := strings.Index(workflow, "\n  phase26-focused:\n")
	if cs < 0 || es < cs || pe < es {
		return []string{"job boundaries not found"}
	}
	jobs := map[string]string{
		"checks":             workflow[cs:es],
		"evidence-aggregate": workflow[es:pe],
	}
	for name, job := range jobs {
		if !strings.Contains(job, "\n      SCHWAY_REQUIRE_NATIVE: \"1\"\n") {
			problems = append(problems, name+" lacks SCHWAY_REQUIRE_NATIVE: \"1\"")
		}
		if strings.Contains(job, "continue-on-error") {
			problems = append(problems, name+" uses continue-on-error")
		}
	}
	agg := jobs["evidence-aggregate"]
	step := strings.Index(agg, "- name: scripts/verify-phase6.sh")
	if step < 0 {
		return append(problems, "verify-phase6 step not found")
	}
	block := agg[step:]
	if next := strings.Index(block[1:], "\n      - "); next >= 0 {
		block = block[:next+1]
	}
	if !strings.Contains(block, "GOFLAGS: -timeout=20m") {
		problems = append(problems, "verify-phase6 step lacks GOFLAGS: -timeout=20m")
	}
	return problems
}

func TestPhase32HostedCIEnv(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath(".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	if problems := phase32HostedCIEnvProblems(workflow); len(problems) != 0 {
		t.Fatalf("hosted CI env problems: %v", problems)
	}
	t.Run("armed", func(t *testing.T) {
		mutated := strings.Replace(workflow, "          GOFLAGS: -timeout=20m\n", "", 1)
		if mutated == workflow {
			t.Fatal("control did not mutate the workflow")
		}
		problems := phase32HostedCIEnvProblems(mutated)
		if len(problems) == 0 || !strings.Contains(strings.Join(problems, ";"), "GOFLAGS") {
			t.Fatalf("armed control not detected: %v", problems)
		}
	})
	t.Run("armed-native", func(t *testing.T) {
		mutated := strings.Replace(workflow, "      SCHWAY_REQUIRE_NATIVE: \"1\"\n", "", 1)
		if mutated == workflow {
			t.Fatal("control did not mutate the workflow")
		}
		if len(phase32HostedCIEnvProblems(mutated)) == 0 {
			t.Fatal("armed-native control not detected")
		}
	})
}
