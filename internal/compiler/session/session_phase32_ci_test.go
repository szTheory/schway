package session_test

import (
	"os"
	"reflect"
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

func TestPhase32TestIndexCacheIsolation(t *testing.T) {
	first := buildTestIndex(t)
	second := buildTestIndex(t)
	if first.root != second.root || !reflect.DeepEqual(first.byImportPath, second.byImportPath) {
		t.Fatal("two buildTestIndex calls differ")
	}
	first.byImportPath["phase32/injected"] = map[string]bool{"TestInjected": true}
	for importPath := range second.byImportPath {
		first.byImportPath[importPath]["TestInjectedToo"] = true
		break
	}
	third := buildTestIndex(t)
	if _, leaked := third.byImportPath["phase32/injected"]; leaked {
		t.Fatal("outer-map mutation leaked into a later caller")
	}
	if !reflect.DeepEqual(second.byImportPath, third.byImportPath) {
		t.Fatal("inner-map mutation leaked into a later caller")
	}
	fresh, err := buildTestIndexWithTags(testsupport.ProjectPath(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if fresh.root != third.root || !reflect.DeepEqual(fresh.byImportPath, third.byImportPath) {
		t.Fatal("cached index differs from a fresh parse")
	}
}

// phase32CIHostMatrixProblems reports missing Phase 32 receipt wiring in the
// evidence-aggregate job (D-08b).
func phase32CIHostMatrixProblems(workflow string) []string {
	start := strings.Index(workflow, "\n  evidence-aggregate:\n")
	end := strings.Index(workflow, "\n  phase26-focused:\n")
	if start < 0 || end < start {
		return []string{"evidence-aggregate job boundaries not found"}
	}
	job := workflow[start:end]
	var problems []string
	step := strings.Index(job, "- name: scripts/verify-phase32.sh")
	upload := strings.Index(job, "- name: Upload Phase 32 host receipt")
	phase27 := strings.Index(job, "- name: Upload Phase 27 host receipt")
	if step < 0 || upload < 0 || phase27 < 0 {
		return append(problems, "phase32 step, phase32 upload or phase27 upload not found")
	}
	if step < phase27 || upload < step {
		problems = append(problems, "phase32 steps must follow the phase27 upload, run then upload")
	}
	runBlock := job[step:upload]
	for _, want := range []string{
		"shell: bash",
		`mkdir -p "$RUNNER_TEMP/phase32"`,
		`sh scripts/verify-phase32.sh 2>&1 | tee "$RUNNER_TEMP/phase32/receipt-${{ matrix.os }}.log"`,
	} {
		if !strings.Contains(runBlock, want) {
			problems = append(problems, "phase32 run step lacks "+want)
		}
	}
	uploadBlock := job[upload:]
	if next := strings.Index(uploadBlock[1:], "\n      - "); next >= 0 {
		uploadBlock = uploadBlock[:next+1]
	}
	for _, want := range []string{
		"if: ${{ always() }}",
		"uses: actions/upload-artifact@v4",
		"name: phase32-receipt-${{ matrix.os }}-${{ github.sha }}",
		"path: ${{ runner.temp }}/phase32/receipt-${{ matrix.os }}.log",
		"if-no-files-found: error",
	} {
		if !strings.Contains(uploadBlock, want) {
			problems = append(problems, "phase32 upload lacks "+want)
		}
	}
	if strings.Contains(job, "continue-on-error") {
		problems = append(problems, "evidence-aggregate uses continue-on-error")
	}
	return problems
}

func TestPhase32CIHostMatrix(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath(".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(data)
	if problems := phase32CIHostMatrixProblems(workflow); len(problems) != 0 {
		t.Fatalf("phase32 CI host matrix problems: %v", problems)
	}
	t.Run("armed", func(t *testing.T) {
		const upload = "name: phase32-receipt-${{ matrix.os }}-${{ github.sha }}\n          path: ${{ runner.temp }}/phase32/receipt-${{ matrix.os }}.log\n          if-no-files-found: error\n"
		if !strings.Contains(workflow, upload) {
			t.Fatal("phase32 upload block not found for the control")
		}
		mutated := strings.Replace(workflow, upload, strings.Replace(upload, "          if-no-files-found: error\n", "", 1), 1)
		problems := phase32CIHostMatrixProblems(mutated)
		if len(problems) == 0 || !strings.Contains(strings.Join(problems, ";"), "if-no-files-found") {
			t.Fatalf("armed control not detected: %v", problems)
		}
	})
}
