package session_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

type phase21DischargeContract struct {
	Schema              string                  `json:"schema"`
	ContractOnly        bool                    `json:"contract_only"`
	ProductionAdmission bool                    `json:"production_admission"`
	Claims              phase21ContractClaims   `json:"claims"`
	Exits               []phase21ContractExit   `json:"exits"`
	Families            []phase21ContractFamily `json:"families"`
}

type phase21ContractClaims struct {
	StructuralValidationProvesRuntimeCleanup    bool `json:"structural_validation_proves_runtime_cleanup"`
	IntroducesLangUnwindOrCancellationSemantics bool `json:"introduces_schway_unwind_or_cancellation_semantics"`
	OneTULTOComparisonProvesOptimizerInactivity bool `json:"one_tu_lto_comparison_proves_optimizer_inactivity"`
}

type phase21ContractExit struct {
	ID             string `json:"id"`
	Disposition    string `json:"disposition"`
	DischargeRule  string `json:"discharge_rule,omitempty"`
	FutureEvidence string `json:"future_evidence,omitempty"`
	Reason         string `json:"reason,omitempty"`
}

type phase21ContractFamily struct {
	Decision            string `json:"decision"`
	Emitter             string `json:"emitter"`
	ProductionAdmission bool   `json:"production_admission"`
	Prerequisite        string `json:"prerequisite"`
	ReopeningEvidence   string `json:"reopening_evidence"`
	RefusalFence        string `json:"refusal_fence"`
	LTOConsequence      string `json:"lto_consequence"`
}

var phase21RequiredExits = map[string]string{
	"normal_return":       "supported",
	"error_return":        "supported",
	"supported_unwind":    "refused",
	"nonlocal_transfer":   "refused",
	"defect":              "outside_cleanup_guarantee",
	"cancellation":        "refused",
	"process_termination": "outside_cleanup_guarantee",
}

var phase21RequiredFamilies = map[string]string{
	"D-16-11": "emitLinearForeign",
	"D-16-12": "emitLinearBorrowedByPointer",
	"D-16-13": "emitLinearBorrowedByPointerPlain",
}

func TestPhase21ResourceDischargeContract(t *testing.T) {
	data, err := os.ReadFile(testsupport.ProjectPath(".planning", "milestones", "M004-phases", "21-native-emission-ownership-and-resource-discharge-m004", "21-RESOURCE-DISCHARGE-CONTRACT.json"))
	if err != nil {
		t.Fatal(err)
	}
	var contract phase21DischargeContract
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatal(err)
	}
	if problems := phase21ContractProblems(contract); len(problems) != 0 {
		t.Fatalf("checked-in resource-discharge contract is invalid: %v", problems)
	}

	for _, mutate := range []struct {
		name  string
		apply func(*phase21DischargeContract)
	}{
		{"missing exit class", func(c *phase21DischargeContract) { c.Exits = c.Exits[1:] }},
		{"unknown exit class", func(c *phase21DischargeContract) {
			c.Exits[0].ID = "opaque_exit"
		}},
		{"duplicate exit class", func(c *phase21DischargeContract) {
			c.Exits[1].ID = c.Exits[0].ID
		}},
		{"wrong exit disposition", func(c *phase21DischargeContract) {
			c.Exits[0].Disposition = "admitted_without_proof"
		}},
		{"missing discharge evidence", func(c *phase21DischargeContract) { c.Exits[0].DischargeRule = "" }},
		{"family admitted", func(c *phase21DischargeContract) { c.Families[0].ProductionAdmission = true }},
		{"missing family", func(c *phase21DischargeContract) { c.Families = c.Families[1:] }},
	} {
		t.Run(mutate.name, func(t *testing.T) {
			copyOfContract := clonePhase21Contract(t, contract)
			mutate.apply(&copyOfContract)
			if problems := phase21ContractProblems(copyOfContract); len(problems) == 0 {
				t.Fatal("invalid in-memory contract mutation was accepted")
			}
		})
	}
}

// TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison keeps the opt-in
// measurement and its committed receipt discoverable from the recurring
// suite without running Clang's multi-lane integration test on every CI job.
func TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison(t *testing.T) {
	testSource, err := os.ReadFile(testsupport.ProjectPath("internal", "compiler", "session", "session_phase21_lto_evidence_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := os.ReadFile(testsupport.ProjectPath(".planning", "milestones", "M004-phases", "21-native-emission-ownership-and-resource-discharge-m004", "21-LTO-EVIDENCE.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, required := range []string{"TestPhase21EmittedMultiFunctionLTOComparison", "multi_function_match_refusal.schway", "emitted_c_sha256", "runtime.GOOS", "cschway_version"} {
		if !strings.Contains(string(testSource), required) {
			t.Errorf("tagged test does not expose required Phase 21 LTO evidence field %q", required)
		}
	}
	for _, required := range []string{"TestPhase21EmittedMultiFunctionLTOComparison", "multi_function_match_refusal.schway", "d096fca69195eb63b09566387690a7b40fdc9c364f55b0b4a24209a895ff6416", "darwin/arm64", "Apple clang version 21.0.0", "does not measure optimizer activity or performance", "no evidence about resource discharge", "other hosts, or other toolchains"} {
		if !strings.Contains(string(receipt), required) {
			t.Errorf("evidence receipt does not contain recorded Phase 21 LTO field %q", required)
		}
	}
	if !strings.Contains(string(testSource), "//go:build phase21_lto_evidence") {
		t.Fatal("LTO comparison must remain opt-in behind phase21_lto_evidence")
	}
	if strings.Contains(string(receipt), "proves optimizer inactivity") || strings.Contains(string(receipt), "measured optimizer activity") {
		t.Fatal("LTO evidence receipt overstates semantic equality as optimizer evidence")
	}
}

func phase21ContractProblems(contract phase21DischargeContract) []string {
	var problems []string
	if contract.Schema != "codename.schway.resource-discharge-contract.v1" {
		problems = append(problems, "unexpected schema")
	}
	if !contract.ContractOnly || contract.ProductionAdmission {
		problems = append(problems, "contract must remain design-only and non-admitting")
	}
	if contract.Claims.StructuralValidationProvesRuntimeCleanup || contract.Claims.IntroducesLangUnwindOrCancellationSemantics || contract.Claims.OneTULTOComparisonProvesOptimizerInactivity {
		problems = append(problems, "contract overstates structural, semantic, or LTO evidence")
	}
	seenExits := make(map[string]bool, len(contract.Exits))
	for _, exit := range contract.Exits {
		expectedDisposition, required := phase21RequiredExits[exit.ID]
		if !required {
			problems = append(problems, "unknown exit class: "+exit.ID)
			continue
		}
		if seenExits[exit.ID] {
			problems = append(problems, "duplicate exit class: "+exit.ID)
		}
		seenExits[exit.ID] = true
		if exit.Disposition != expectedDisposition {
			problems = append(problems, fmt.Sprintf("%s disposition must be %s", exit.ID, expectedDisposition))
		}
		if expectedDisposition == "supported" {
			if exit.DischargeRule == "" || exit.FutureEvidence == "" {
				problems = append(problems, exit.ID+" requires discharge and future evidence")
			}
		} else if exit.Reason == "" {
			problems = append(problems, exit.ID+" requires a refusal or guarantee-boundary reason")
		}
	}
	for id := range phase21RequiredExits {
		if !seenExits[id] {
			problems = append(problems, "missing exit class: "+id)
		}
	}
	seenFamilies := make(map[string]bool, len(contract.Families))
	for _, family := range contract.Families {
		emitter, required := phase21RequiredFamilies[family.Decision]
		if !required {
			problems = append(problems, "unknown family decision: "+family.Decision)
			continue
		}
		if seenFamilies[family.Decision] {
			problems = append(problems, "duplicate family decision: "+family.Decision)
		}
		seenFamilies[family.Decision] = true
		if family.Emitter != emitter || family.ProductionAdmission {
			problems = append(problems, family.Decision+" has drifted emitter identity or production admission")
		}
		if family.Prerequisite == "" || family.ReopeningEvidence == "" || family.RefusalFence == "" || family.LTOConsequence == "" {
			problems = append(problems, family.Decision+" lacks family-specific reopening or refusal conditions")
		}
	}
	for id := range phase21RequiredFamilies {
		if !seenFamilies[id] {
			problems = append(problems, "missing family decision: "+id)
		}
	}
	return problems
}

func clonePhase21Contract(t *testing.T, contract phase21DischargeContract) phase21DischargeContract {
	t.Helper()
	data, err := json.Marshal(contract)
	if err != nil {
		t.Fatal(err)
	}
	var clone phase21DischargeContract
	if err := json.Unmarshal(data, &clone); err != nil {
		t.Fatal(err)
	}
	return clone
}
