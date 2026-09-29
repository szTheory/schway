package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase23DischargeContract struct {
	Schema              string                      `json:"schema"`
	ContractOnly        bool                        `json:"contract_only"`
	ProductionAdmission bool                        `json:"production_admission"`
	Claims              phase23ContractClaims       `json:"claims"`
	Transitions         []phase23ContractTransition `json:"transitions"`
	Acquisition         phase23ContractAcquisition  `json:"acquisition"`
	ErrorPath           phase23ContractErrorPath    `json:"error_path"`
	EvidenceScopes      phase23EvidenceScopes       `json:"evidence_scopes"`
	Exits               []phase23ContractExit       `json:"exits"`
}

type phase23ContractClaims struct {
	FailedOrPartialAcquisitionMintsOwner     bool `json:"failed_or_partial_acquisition_mints_owner"`
	UseErrorCleanupPrecedesOrdinaryError     bool `json:"use_error_cleanup_precedes_ordinary_error_reporting"`
	StructuralValidationProvesRuntimeCleanup bool `json:"structural_validation_proves_runtime_cleanup"`
	InterpreterModelProvesHostIO             bool `json:"interpreter_model_proves_host_io"`
	InterpreterModelProvesPhysicalCleanup    bool `json:"interpreter_model_proves_physical_cleanup"`
	ModelAndNativeReceiptsAreDistinct        bool `json:"model_and_native_receipts_are_distinct"`
	TransferImplementedInPhase23             bool `json:"transfer_implemented_in_phase_23"`
}

type phase23ContractTransition struct {
	ID               string `json:"id"`
	Effect           string `json:"effect"`
	Infallible       bool   `json:"infallible"`
	Phase23Execution string `json:"phase_23_execution"`
	TargetPhase      int    `json:"target_phase,omitempty"`
	EvidenceScope    string `json:"evidence_scope"`
}

type phase23ContractAcquisition struct {
	SuccessfulAcquire        string `json:"successful_acquire"`
	FailedOrPartialAcquire   string `json:"failed_or_partial_acquire"`
	PartialAllocationCleanup string `json:"partial_allocation_cleanup"`
}

type phase23ContractErrorPath struct {
	UseError      string   `json:"use_error"`
	RequiredOrder []string `json:"required_order"`
}

type phase23EvidenceScopes struct {
	StructuralValidation phase23StructuralEvidence `json:"structural_validation"`
	InterpreterModel     phase23ModelEvidence      `json:"interpreter_model"`
	NativeReceipt        phase23NativeEvidence     `json:"native_receipt"`
}

type phase23StructuralEvidence struct {
	Scope                string `json:"scope"`
	ProvesRuntimeCleanup bool   `json:"proves_runtime_cleanup"`
}

type phase23ModelEvidence struct {
	Scope                 string `json:"scope"`
	ProvesHostIO          bool   `json:"proves_host_io"`
	ProvesPhysicalCleanup bool   `json:"proves_physical_cleanup"`
}

type phase23NativeEvidence struct {
	Scope                                 string `json:"scope"`
	RequiredForPhysicalIOAndCleanupClaims bool   `json:"required_for_physical_io_and_cleanup_claims"`
}

type phase23ContractExit struct {
	ID             string `json:"id"`
	Disposition    string `json:"disposition"`
	DischargeRule  string `json:"discharge_rule,omitempty"`
	FutureEvidence string `json:"future_evidence,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

var phase23RequiredTransitions = map[string]phase23ContractTransition{
	"release": {
		ID: "release", Effect: "consumes_one_local_obligation", Infallible: true,
		Phase23Execution: "runnable_local_release", EvidenceScope: "native_observer_required_for_physical_cleanup",
	},
	"borrow": {
		ID: "borrow", Effect: "preserves_same_owner_obligation", Infallible: false,
		Phase23Execution: "runnable_borrowed_use", EvidenceScope: "native_observer_required_for_physical_use",
	},
	"transfer": {
		ID: "transfer", Effect: "preserves_live_resource_under_new_owner", Infallible: false,
		Phase23Execution: "contract_only", TargetPhase: 24, EvidenceScope: "future_phase_native_observer",
	},
}

var phase23RequiredExits = map[string]string{
	"normal_return":       "supported",
	"error_return":        "supported",
	"supported_unwind":    "refused",
	"nonlocal_transfer":   "refused",
	"defect":              "outside_cleanup_guarantee",
	"cancellation":        "refused",
	"process_termination": "outside_cleanup_guarantee",
}

func TestPhase23ContractTransitionsAndEvidenceScopes(t *testing.T) {
	path := testsupport.ProjectPath(".planning", "phases", "23-live-local-allocation-and-discharge", "23-RESOURCE-DISCHARGE-CONTRACT.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	contract, err := decodePhase23Contract(data)
	if err != nil {
		t.Fatal(err)
	}
	if problems := phase23ContractProblems(contract); len(problems) != 0 {
		t.Fatalf("checked-in Phase 23 resource-discharge contract is invalid: %v", problems)
	}

	for _, mutate := range []struct {
		name  string
		apply func(*phase23DischargeContract)
	}{
		{"missing transition", func(c *phase23DischargeContract) { c.Transitions = c.Transitions[1:] }},
		{"duplicate transition", func(c *phase23DischargeContract) { c.Transitions[1].ID = c.Transitions[0].ID }},
		{"unknown transition", func(c *phase23DischargeContract) { c.Transitions[0].ID = "opaque_transition" }},
		{"release does not consume", func(c *phase23DischargeContract) { c.Transitions[0].Effect = "preserves_same_owner_obligation" }},
		{"borrow consumes", func(c *phase23DischargeContract) { c.Transitions[1].Effect = "consumes_one_local_obligation" }},
		{"transfer lacks new owner", func(c *phase23DischargeContract) { c.Transitions[2].Effect = "preserves_same_owner_obligation" }},
		{"transfer admitted in phase 23", func(c *phase23DischargeContract) { c.Transitions[2].Phase23Execution = "runnable_transfer" }},
		{"production admitted", func(c *phase23DischargeContract) { c.ProductionAdmission = true }},
		{"failed acquire creates owner", func(c *phase23DischargeContract) { c.Claims.FailedOrPartialAcquisitionMintsOwner = true }},
		{"structural checks claim cleanup", func(c *phase23DischargeContract) { c.Claims.StructuralValidationProvesRuntimeCleanup = true }},
		{"model claims physical cleanup", func(c *phase23DischargeContract) { c.EvidenceScopes.InterpreterModel.ProvesPhysicalCleanup = true }},
		{"missing exit class", func(c *phase23DischargeContract) { c.Exits = c.Exits[1:] }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			copyOfContract := clonePhase23Contract(t, contract)
			mutate.apply(&copyOfContract)
			if problems := phase23ContractProblems(copyOfContract); len(problems) == 0 {
				t.Fatal("invalid in-memory contract mutation was accepted")
			}
		})
	}

	unknownField := bytes.Replace(data, []byte(`"contract_only": true`), []byte(`"contract_only": true, "unreviewed_claim": true`), 1)
	if _, err := decodePhase23Contract(unknownField); err == nil {
		t.Fatal("strict contract decoder accepted an unknown field")
	}
	missingFalseClaim := bytes.Replace(data, []byte(`"proves_host_io": false,`), nil, 1)
	if _, err := decodePhase23Contract(missingFalseClaim); err == nil {
		t.Fatal("strict contract decoder accepted an omitted false-valued evidence claim")
	}
}

func TestPhase23ReadmeExpectedAnswersAreIndependentConstants(t *testing.T) {
	type publicCase struct {
		ID           string `json:"id"`
		FileBytesHex string `json:"file_bytes_hex"`
		ExitCode     int    `json:"exit_code"`
		Stdout       string `json:"stdout"`
		Stderr       string `json:"stderr"`
	}
	var expected struct {
		Schema string       `json:"schema"`
		Cases  []publicCase `json:"cases"`
	}
	path := testsupport.ProjectPath("examples", "phase23", "file_byte.expected.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&expected); err != nil {
		t.Fatal(err)
	}
	if expected.Schema != "schway.phase23-file-byte-expected/1" || len(expected.Cases) != 5 {
		t.Fatalf("unexpected independent Phase 23 answer fixture: schema=%q cases=%+v", expected.Schema, expected.Cases)
	}
	want := []publicCase{
		{ID: "byte-41", FileBytesHex: "41", ExitCode: 0, Stdout: "65\n", Stderr: ""},
		{ID: "byte-42", FileBytesHex: "42", ExitCode: 0, Stdout: "66\n", Stderr: ""},
		{ID: "empty-acquire-error", FileBytesHex: "", ExitCode: 65, Stdout: "", Stderr: "schway_file_byte_acquire: EmptyFile\n"},
		{ID: "two-byte-acquire-error", FileBytesHex: "4142", ExitCode: 65, Stdout: "", Stderr: "schway_file_byte_acquire: FileTooLong\n"},
		{ID: "unsupported-byte-use-error", FileBytesHex: "43", ExitCode: 65, Stdout: "", Stderr: "schway_file_byte_use: UnsupportedByte\n"},
	}
	for index, test := range expected.Cases {
		if test != want[index] {
			t.Errorf("independent answer %d is %+v, want literal %+v", index, test, want[index])
		}
	}
}

func TestPhase23VerifierScriptContract(t *testing.T) {
	path := testsupport.ProjectPath("scripts", "verify-phase23.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, required := range []string{
		"for tool in go clang git uname date mktemp rm cat grep sed",
		"clang -dumpmachine",
		"host identity is incomplete",
		"status=incomplete",
		"status=pass",
		"grep -Fq -- '--- SKIP:'",
		"testing: warning: no tests to run",
		"run_phase23_tests source-admission",
		"run_phase23_tests independent-peers",
		"run_phase23_tests model-only",
		"run_phase23_tests native-observer",
		"run_phase23_tests public-contract",
	} {
		if !strings.Contains(text, required) {
			t.Errorf("focused Phase 23 verifier is missing fail-closed contract %q", required)
		}
	}
	for _, duplicateLane := range []string{
		"go test ./...",
		"go test -race",
		"go vet ./...",
		"ASAN_OPTIONS=",
		"UBSAN_OPTIONS=",
		"scripts/verify-phase6.sh",
	} {
		if strings.Contains(text, duplicateLane) {
			t.Errorf("focused Phase 23 verifier duplicates an existing full or sanitizer lane: %q", duplicateLane)
		}
	}
}

func decodePhase23Contract(data []byte) (phase23DischargeContract, error) {
	var contract phase23DischargeContract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		return phase23DischargeContract{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return phase23DischargeContract{}, fmt.Errorf("unexpected trailing JSON value")
		}
		return phase23DischargeContract{}, err
	}
	if err := requirePhase23BooleanFields(data); err != nil {
		return phase23DischargeContract{}, err
	}
	return contract, nil
}

func requirePhase23BooleanFields(data []byte) error {
	paths := []string{
		"contract_only", "production_admission",
		"claims.failed_or_partial_acquisition_mints_owner",
		"claims.use_error_cleanup_precedes_ordinary_error_reporting",
		"claims.structural_validation_proves_runtime_cleanup",
		"claims.interpreter_model_proves_host_io",
		"claims.interpreter_model_proves_physical_cleanup",
		"claims.model_and_native_receipts_are_distinct",
		"claims.transfer_implemented_in_phase_23",
		"evidence_scopes.structural_validation.proves_runtime_cleanup",
		"evidence_scopes.interpreter_model.proves_host_io",
		"evidence_scopes.interpreter_model.proves_physical_cleanup",
		"evidence_scopes.native_receipt.required_for_physical_io_and_cleanup_claims",
	}
	for _, path := range paths {
		var value json.RawMessage = data
		for _, key := range strings.Split(path, ".") {
			var object map[string]json.RawMessage
			if err := json.Unmarshal(value, &object); err != nil {
				return fmt.Errorf("decode %q: %w", path, err)
			}
			var ok bool
			value, ok = object[key]
			if !ok {
				return fmt.Errorf("required boolean %q is missing", path)
			}
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("required boolean %q cannot be null", path)
		}
		var boolean bool
		if err := json.Unmarshal(value, &boolean); err != nil {
			return fmt.Errorf("required field %q is not a boolean: %w", path, err)
		}
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(data, &root); err != nil {
		return err
	}
	var transitions []map[string]json.RawMessage
	if err := json.Unmarshal(root["transitions"], &transitions); err != nil {
		return fmt.Errorf("decode transitions: %w", err)
	}
	for i, transition := range transitions {
		raw, ok := transition["infallible"]
		if !ok {
			return fmt.Errorf("required boolean transitions.%d.infallible is missing", i)
		}
		if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
			return fmt.Errorf("required boolean transitions.%d.infallible cannot be null", i)
		}
		var boolean bool
		if err := json.Unmarshal(raw, &boolean); err != nil {
			return fmt.Errorf("required field transitions.%d.infallible is not a boolean: %w", i, err)
		}
	}
	return nil
}

func phase23ContractProblems(contract phase23DischargeContract) []string {
	var problems []string
	if contract.Schema != "codename.schway.resource-discharge-contract.v2" {
		problems = append(problems, "unexpected schema")
	}
	if !contract.ContractOnly || contract.ProductionAdmission {
		problems = append(problems, "contract must remain design-only and non-admitting")
	}
	claims := contract.Claims
	if claims.FailedOrPartialAcquisitionMintsOwner || !claims.UseErrorCleanupPrecedesOrdinaryError || claims.StructuralValidationProvesRuntimeCleanup || claims.InterpreterModelProvesHostIO || claims.InterpreterModelProvesPhysicalCleanup || !claims.ModelAndNativeReceiptsAreDistinct || claims.TransferImplementedInPhase23 {
		problems = append(problems, "contract overstates ownership transitions or evidence")
	}

	seenTransitions := make(map[string]bool, len(contract.Transitions))
	for _, transition := range contract.Transitions {
		expected, required := phase23RequiredTransitions[transition.ID]
		if !required {
			problems = append(problems, "unknown transition: "+transition.ID)
			continue
		}
		if seenTransitions[transition.ID] {
			problems = append(problems, "duplicate transition: "+transition.ID)
		}
		seenTransitions[transition.ID] = true
		if transition != expected {
			problems = append(problems, transition.ID+" transition semantics or evidence scope drifted")
		}
	}
	for id := range phase23RequiredTransitions {
		if !seenTransitions[id] {
			problems = append(problems, "missing transition: "+id)
		}
	}
	if contract.Acquisition.SuccessfulAcquire != "creates_one_noncopyable_local_owner" || contract.Acquisition.FailedOrPartialAcquire != "creates_no_owner" || contract.Acquisition.PartialAllocationCleanup != "adapter_before_failure_return" {
		problems = append(problems, "acquisition ownership semantics drifted")
	}
	if contract.ErrorPath.UseError != "ordinary_typed_failure" || !equalPhase23Strings(contract.ErrorPath.RequiredOrder, []string{"keep_owner_live", "run_infallible_local_release", "report_ordinary_error"}) {
		problems = append(problems, "use-error cleanup order drifted")
	}
	if contract.EvidenceScopes.StructuralValidation.Scope != "checked_core_structure" || contract.EvidenceScopes.StructuralValidation.ProvesRuntimeCleanup || contract.EvidenceScopes.InterpreterModel.Scope != "deterministic_supplied_operation_outcomes" || contract.EvidenceScopes.InterpreterModel.ProvesHostIO || contract.EvidenceScopes.InterpreterModel.ProvesPhysicalCleanup || contract.EvidenceScopes.NativeReceipt.Scope != "independent_native_observer" || !contract.EvidenceScopes.NativeReceipt.RequiredForPhysicalIOAndCleanupClaims {
		problems = append(problems, "evidence scopes are missing, conflated, or overstated")
	}
	seenExits := make(map[string]bool, len(contract.Exits))
	for _, exit := range contract.Exits {
		disposition, required := phase23RequiredExits[exit.ID]
		if !required {
			problems = append(problems, "unknown exit class: "+exit.ID)
			continue
		}
		if seenExits[exit.ID] {
			problems = append(problems, "duplicate exit class: "+exit.ID)
		}
		seenExits[exit.ID] = true
		if exit.Disposition != disposition {
			problems = append(problems, fmt.Sprintf("%s disposition must be %s", exit.ID, disposition))
		}
		if disposition == "supported" {
			if exit.DischargeRule == "" || exit.FutureEvidence == "" {
				problems = append(problems, exit.ID+" requires local discharge and future evidence")
			}
		} else if exit.Reason == "" {
			problems = append(problems, exit.ID+" requires a refusal or guarantee-boundary reason")
		}
	}
	for id := range phase23RequiredExits {
		if !seenExits[id] {
			problems = append(problems, "missing exit class: "+id)
		}
	}
	return problems
}

func equalPhase23Strings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func clonePhase23Contract(t *testing.T, contract phase23DischargeContract) phase23DischargeContract {
	t.Helper()
	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	clone, err := decodePhase23Contract(data)
	if err != nil {
		t.Fatal(err)
	}
	return clone
}
