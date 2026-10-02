package session_test

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/check"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/interp"
	"github.com/szTheory/schway/internal/compiler/syntax"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase24ModelReplayIsDeterministicAndModelOnly(t *testing.T) {
	program, acquireID, useID := checkedPhase24ModelProgram(t)
	for _, test := range []struct {
		name      string
		inputByte byte
		wantKind  string
		wantValue string
		wantType  string
	}{
		{name: "0x41 returns independent answer 65", inputByte: 0x41, wantKind: "returned", wantValue: "65"},
		{name: "0x42 returns independent answer 66", inputByte: 0x42, wantKind: "returned", wantValue: "66"},
		{name: "0x43 returns typed UseError.UnsupportedByte", inputByte: 0x43, wantKind: "typed_failure", wantValue: "UnsupportedByte", wantType: "UseError"},
	} {
		t.Run(test.name, func(t *testing.T) {
			byteValue := "65"
			use := interp.ForeignOutcome{Kind: "success", Type: "U64", Value: test.wantValue}
			if test.inputByte == 0x42 {
				byteValue = "66"
			}
			if test.inputByte == 0x43 {
				byteValue = "67"
				use = interp.ForeignOutcome{Kind: "failure", Type: test.wantType, Value: test.wantValue}
			}
			result, err := interp.RunWithForeignOutcomes(program, "main", "opaque-path-token", map[string]interp.ForeignOutcome{
				acquireID: {Kind: "success", Type: "FileByteOwner", Value: byteValue},
				useID:     use,
			})
			if err != nil {
				t.Fatalf("model replay failed: %v", err)
			}
			if result.Execution.Outcome.Kind != test.wantKind || result.Execution.Outcome.Value != test.wantValue {
				t.Fatalf("model answer=%+v, want independently fixed %s:%s", result.Execution.Outcome, test.wantKind, test.wantValue)
			}
			if result.EvidenceScope != interp.EvidenceScopeModelOnly || result.ActualHostIO || result.PhysicalCleanup || len(result.Execution.LiveResources) != 0 {
				t.Fatalf("model result crossed its evidence boundary or retained obligations: %+v", result)
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatalf("encode model-only evidence: %v", err)
			}
			for _, field := range []string{`"evidence_scope":"interpreter_model_only"`, `"actual_host_io":false`, `"physical_cleanup":false`} {
				if !strings.Contains(string(encoded), field) {
					t.Errorf("serialized model evidence omits %s: %s", field, encoded)
				}
			}
			if test.wantType != "" {
				foundTypedFailure := false
				for _, event := range result.Execution.Events {
					if event.Kind == "function.failed" && event.TypeID == test.wantType {
						foundTypedFailure = true
					}
				}
				if !foundTypedFailure {
					t.Fatalf("model event stream omitted typed failure %s: %+v", test.wantType, result.Execution.Events)
				}
			}
		})
	}
}

func TestPhase24ReadmeAndVerifierScriptContract(t *testing.T) {
	readmeBytes, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "README.md"))
	if err != nil {
		t.Fatal(err)
	}
	readme := string(readmeBytes)
	for _, required := range []string{
		"examples/phase24/transfer.schway",
		"examples/phase24/error.schway",
		"examples/phase23/file_byte.bindings.json",
		"0x41",
		"65",
		"0x42",
		"66",
		"UseError.UnsupportedByte",
		"actual host IO",
		"physical cleanup",
		"historical Phase 23",
	} {
		if !strings.Contains(readme, required) {
			t.Errorf("Phase 24 public contract omits %q", required)
		}
	}

	scriptBytes, err := os.ReadFile(testsupport.ProjectPath("scripts", "verify-phase24.sh"))
	if err != nil {
		t.Fatal(err)
	}
	script := string(scriptBytes)
	for _, required := range []string{
		"for tool in go clang git uname date mktemp rm cat grep sed",
		"clang -dumpmachine",
		"host identity is incomplete",
		"status=incomplete",
		"status=pass",
		"grep -Fq -- '--- SKIP:'",
		"testing: warning: no tests to run",
		"run_phase24_tests source-refusal",
		"run_phase24_tests independent-peer",
		"run_phase24_tests model-only",
		"run_phase24_tests emitter-admission",
		"run_phase24_tests native-public-app",
		"run_phase24_tests native-observer",
		"run_phase24_tests public-contract",
	} {
		if !strings.Contains(script, required) {
			t.Errorf("Phase 24 focused verifier omits fail-closed contract %q", required)
		}
	}
	for _, duplicateLane := range []string{"go test ./...", "go test -race", "go vet ./...", "scripts/verify-phase6.sh", "scripts/verify-phase23.sh"} {
		if strings.Contains(script, duplicateLane) {
			t.Errorf("Phase 24 focused verifier duplicates a separately owned lane: %q", duplicateLane)
		}
	}

	workflowBytes, err := os.ReadFile(testsupport.ProjectPath(".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(workflowBytes)
	start := strings.Index(workflow, "  evidence-aggregate:")
	if start < 0 {
		t.Fatal("CI workflow has no bounded existing evidence-aggregate job")
	}
	end := strings.Index(workflow[start:], "\n  validation-corpus-receipt:")
	if end < 0 {
		t.Fatal("CI workflow has no following validation-corpus job boundary")
	}
	job := workflow[start : start+end]
	if count := strings.Count(job, "sh scripts/verify-phase24.sh"); count != 1 {
		t.Fatalf("existing evidence-aggregate job invokes Phase 24 focused receipt %d times, want one shared step", count)
	}
	for _, host := range []string{"ubuntu-latest", "macos-latest"} {
		if !strings.Contains(job, host) {
			t.Errorf("existing evidence-aggregate job omits required Phase 24 host %s", host)
		}
	}
}

func checkedPhase24ModelProgram(t *testing.T) (core.Program, string, string) {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("examples", "phase24", "error.schway"))
	if err != nil {
		t.Fatal(err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("Phase 24 model fixture parse diagnostics: %+v", parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 {
		t.Fatalf("Phase 24 model fixture check diagnostics: %+v", checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("Phase 24 model fixture core problems: %+v", validated.Problems)
	}
	program := validated.Program()
	var acquireID, useID string
	for _, function := range program.Functions {
		if function.Linear == nil {
			continue
		}
		for _, operation := range function.Linear.Operations {
			if operation.Foreign == nil {
				continue
			}
			if function.Name == "acquire" && operation.Foreign.Mode == "acquire" {
				acquireID = operation.ID
			}
			if function.Name == "probe" && operation.Foreign.Mode == "borrow" {
				useID = operation.ID
			}
		}
	}
	if acquireID == "" || useID == "" {
		t.Fatalf("Phase 24 model fixture omitted static acquire/use operation IDs: acquire=%q use=%q", acquireID, useID)
	}
	return program, acquireID, useID
}
