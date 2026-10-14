package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// TestPhase35AdmissionRefusalMutationControls removes each independent
// shift-admission refusal from temporary owner-source copies and requires the
// exact raw-IR/source assertion to report acceptance or its named diagnostic
// contract changing. The no-op and restored overlays run before a mutation can
// be credited; Go's source overlay keeps implementation files untouched.
func TestPhase35AdmissionRefusalMutationControls(t *testing.T) {
	coreShapeStart := "case core.OpShiftLeftU64, core.OpShiftRightU64:\n\t\t\ttarget, targetKnown := places[operation.TargetID]"
	coreShapeEnd := "\n\t\tcase core.OpBytesLen, core.OpBytesAt:"
	coreFixedShiftStart := "case core.OpShiftLeftU64, core.OpShiftRightU64:\n\t\t\t\t\tamount, err := strconv.ParseUint(operation.ConstU64, 10, 64)"
	coreFixedShiftEnd := "\n\t\t\t\tcase core.OpBytesLen, core.OpBytesAt:"
	coreStraightFunction := "func (v *validator) validateStraightLineScalarFacts"
	coreReplayStraight := "func (v *validator) replayStraightLine"
	coreReplayBlocks := "func (v *validator) replayBlocks"
	coreTargetHelperStart := "func scalarTargetMatches"
	coreTargetHelperEnd := "\nfunc hasScalarCFG"

	originShapeShiftStart := "case core.OpShiftLeftU64, core.OpShiftRightU64:\n\t\t\tamount, err := strconv.ParseUint(operation.ConstU64, 10, 64)"
	originShapeEnd := "\n\treturn nil"
	originTransferShiftStart := "case core.OpShiftLeftU64, core.OpShiftRightU64:\n\t\tamount, err := strconv.ParseUint(operation.ConstU64, 10, 64)"
	originTransferShiftEnd := "\n\tcase core.OpRemainderU64:"

	corePeer := phase35AdmissionPeer{
		name: "corevalidate", packagePath: "./internal/compiler/corevalidate",
		sourcePath:   "internal/compiler/corevalidate/corevalidate.go",
		selector:     "^(TestPhase35ShiftImmediateAdmission|TestPhase35ShiftTargetFreshness)$",
		noOpTests:    []string{"TestPhase35ShiftImmediateAdmission", "TestPhase35ShiftTargetFreshness"},
		noOpSubtests: append(phase35AdmissionRows("TestPhase35ShiftImmediateAdmission"), phase35FreshnessRows("TestPhase35ShiftTargetFreshness")...),
		mutations: []phase35AdmissionMutation{
			{
				name: "canonical-and-present-immediate", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "missing_immediate", "leading_zero"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "missing_immediate", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "leading_zero", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{
					{start: "func scalarShiftImmediateValid", end: "\nfunc ", before: "amount, err := strconv.ParseUint(raw, 10, 64)\n\treturn err == nil && strconv.FormatUint(amount, 10) == raw && amount <= 63", after: "amount, _ := strconv.ParseUint(raw, 10, 64)\n\treturn amount <= 63", sites: 1},
					{start: coreFixedShiftStart, end: coreFixedShiftEnd, before: "validAmount := err == nil", after: "validAmount := err == nil || operation.ConstU64 == \"\"", sites: 1},
					{start: coreFixedShiftStart, end: coreFixedShiftEnd, before: "strconv.FormatUint(amount, 10) == operation.ConstU64 &&", after: "true &&", sites: 1},
					{start: coreStraightFunction, end: "\nfunc ", before: "validAmount := err == nil", after: "validAmount := err == nil || operation.ConstU64 == \"\"", sites: 1},
					{start: coreStraightFunction, end: "\nfunc ", before: "strconv.FormatUint(amount, 10) == operation.ConstU64 &&", after: "true &&", sites: 1},
				},
			},
			{
				name: "immediate-range-0-through-63", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "count_64"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "count_64", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: "func scalarShiftImmediateValid", end: "\nfunc ", before: " && amount <= 63", after: "", sites: 1},
					{start: coreFixedShiftStart, end: coreFixedShiftEnd, before: "amount <= 63", after: "true", sites: 1},
					{start: coreStraightFunction, end: "\nfunc ", before: "amount <= 63", after: "true", sites: 1},
				},
			},
			{
				name: "shift-right-place-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "extra_right_place"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "extra_right_place", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "operation.RightID == \"\" &&", after: "true &&", sites: 1},
					{before: "operation.Kind != core.OpAddChecked && operation.Kind != core.OpLessU64 && operation.Kind != core.OpEqualU64 && operation.Kind != core.OpRemainderU64 && operation.Kind != core.OpBitXorU64 && operation.Kind != core.OpBitAndU64 && operation.Kind != core.OpBitOrU64 && operation.Kind != core.OpBytesAt && operation.RightID != \"\" || ", after: "", sites: 1},
				},
			},
			{
				name: "source-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_source_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_source_type", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "types[places[operation.SourceID].TypeID].Shape.Constructor == \"U64\" &&", after: "true &&", sites: 1},
					{start: coreReplayStraight, end: "\nfunc (v *validator) replayBlocks", before: "types[source.TypeID].Shape.Constructor == \"U64\" && types[operation.TypeID].Shape.Constructor == \"U64\"", after: "types[operation.TypeID].Shape.Constructor == \"U64\"", sites: 1},
					{start: coreReplayBlocks, end: "\nfunc ", before: "types[source.TypeID].Shape.Constructor == \"U64\" && types[operation.TypeID].Shape.Constructor == \"U64\"", after: "types[operation.TypeID].Shape.Constructor == \"U64\"", sites: 1},
				},
			},
			{
				name: "operation-result-type-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_result_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_result_type", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "operation.TypeID == target.TypeID &&", after: "true &&", sites: 1},
					{start: coreReplayStraight, end: "\nfunc (v *validator) replayBlocks", before: "types[source.TypeID].Shape.Constructor == \"U64\" && types[operation.TypeID].Shape.Constructor == \"U64\"", after: "types[source.TypeID].Shape.Constructor == \"U64\"", sites: 1},
					{start: coreReplayBlocks, end: "\nfunc ", before: "types[source.TypeID].Shape.Constructor == \"U64\" && types[operation.TypeID].Shape.Constructor == \"U64\"", after: "types[source.TypeID].Shape.Constructor == \"U64\"", sites: 1},
					{start: coreTargetHelperStart, end: coreTargetHelperEnd, before: "target.TypeID == operation.TypeID", after: "true", sites: 1},
				},
			},
			{
				name: "result-place-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_result_place_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_result_place_type", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "types[target.TypeID].Shape.Constructor == \"U64\" && operation.TypeID == target.TypeID &&", after: "", sites: 1},
					{start: coreShapeStart, end: coreShapeEnd, before: "target, targetKnown :=", after: "_, targetKnown :=", sites: 1},
					{start: coreTargetHelperStart, end: coreTargetHelperEnd, before: "target.TypeID == operation.TypeID", after: "true", sites: 1},
				},
			},
			{
				name: "source-must-be-initialized", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "uninitialized_U64_source"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "uninitialized_U64_source", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: coreFixedShiftStart, end: coreFixedShiftEnd, before: "source.initialized && validAmount &&", after: "true && validAmount &&", sites: 1},
					{start: coreReplayStraight, end: "\nfunc (v *validator) replayBlocks", before: "optionalOperationSource(function, operation) || initialized[operation.SourceID]", after: "optionalOperationSource(function, operation) || operation.Kind == core.OpShiftLeftU64 || operation.Kind == core.OpShiftRightU64 || initialized[operation.SourceID]", sites: 1},
					{start: coreReplayBlocks, end: "\nfunc ", before: "optionalOperationSource(function, operation) || initialized[operation.SourceID]", after: "optionalOperationSource(function, operation) || operation.Kind == core.OpShiftLeftU64 || operation.Kind == core.OpShiftRightU64 || initialized[operation.SourceID]", sites: 1},
				},
			},
			{
				name: "target-identity-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "parameter_target", "source_aliases_target"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "parameter_target", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "source_aliases_target", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "operation.TargetID != function.Parameter.ID && operation.TargetID != operation.SourceID &&", after: "true &&", sites: 1},
					{start: coreShapeStart, end: coreShapeEnd, before: "if !v.check(freshTarget, \"core.invalid_target\", operation.TargetID) {", after: "if false && !v.check(freshTarget, \"core.invalid_target\", operation.TargetID) {", sites: 1},
					{start: coreTargetHelperStart, end: coreTargetHelperEnd, before: "operation.TargetID != operation.SourceID && !produced[operation.TargetID]", after: "true", sites: 1},
				},
			},
			{
				name: "target-freshness", selector: "^TestPhase35ShiftTargetFreshness$",
				expected: phase35FreshnessFailures("TestPhase35ShiftTargetFreshness"),
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "if !v.check(freshTarget, \"core.invalid_target\", operation.TargetID) {", after: "if false && !v.check(freshTarget, \"core.invalid_target\", operation.TargetID) {", sites: 1},
					{start: coreTargetHelperStart, end: coreTargetHelperEnd, before: "!produced[operation.TargetID]", after: "true", sites: 1},
				},
			},
			{
				name: "unrelated-fields", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "foreign_field"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "foreign_field", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{start: coreShapeStart, end: coreShapeEnd, before: "scalarOperationHasNoUnrelatedFields(operation)", after: "true", sites: 1}},
			},
			{
				name: "store-and-edge-field-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "store_field", "edge_field"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "store_field", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "edge_field", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{
					{start: coreShapeStart, end: coreShapeEnd, before: "operation.StoreTargetID == \"\" && operation.TrueEdgeID == \"\" && operation.FalseEdgeID == \"\" &&", after: "true &&", sites: 1},
					{before: " || operation.Kind != core.OpScalarStore && operation.StoreTargetID != \"\" || operation.Kind != core.OpBranch && (operation.TrueEdgeID != \"\" || operation.FalseEdgeID != \"\")", after: "", sites: 1},
				},
			},
			{
				name: "ConstU64-kind-exclusivity", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "unrelated_const_field"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "unrelated_const_field", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{before: "if operation.Kind != core.OpConst && operation.Kind != core.OpShiftLeftU64 && operation.Kind != core.OpShiftRightU64 && operation.ConstU64 != \"\" {", after: "if false {", sites: 1},
					{before: "operation.ConstU64 == \"\")", after: "true)", sites: 1},
				},
			},
		},
	}

	originPeer := phase35AdmissionPeer{
		name: "originvalidate", packagePath: "./internal/compiler/originvalidate",
		sourcePath:   "internal/compiler/originvalidate/originvalidate.go",
		selector:     "^(TestPhase35ShiftImmediateAdmission|TestPhase35ShiftTargetFreshness)$",
		noOpTests:    []string{"TestPhase35ShiftImmediateAdmission", "TestPhase35ShiftTargetFreshness"},
		noOpSubtests: append(phase35AdmissionRows("TestPhase35ShiftImmediateAdmission"), phase35FreshnessRows("TestPhase35ShiftTargetFreshness")...),
		mutations: []phase35AdmissionMutation{
			{
				name: "canonical-and-present-immediate", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "missing_immediate", "leading_zero"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "missing_immediate", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "leading_zero", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{
					{start: originShapeShiftStart, end: originShapeEnd, before: "err != nil ||", after: "(err != nil && operation.ConstU64 != \"\") ||", sites: 1},
					{start: originShapeShiftStart, end: originShapeEnd, before: "strconv.FormatUint(amount, 10) != operation.ConstU64 ||", after: "false ||", sites: 1},
					{start: originTransferShiftStart, end: originTransferShiftEnd, before: "err != nil ||", after: "(err != nil && operation.ConstU64 != \"\") ||", sites: 1},
					{start: originTransferShiftStart, end: originTransferShiftEnd, before: "strconv.FormatUint(amount, 10) != operation.ConstU64 ||", after: "false ||", sites: 1},
				},
			},
			{
				name: "immediate-range-0-through-63", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "count_64"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "count_64", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: originShapeShiftStart, end: originShapeEnd, before: "amount > 63 ||", after: "false ||", sites: 1},
					{start: originTransferShiftStart, end: originTransferShiftEnd, before: "amount > 63 ||", after: "false ||", sites: 1},
				},
			},
			{
				name: "shift-right-place-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "extra_right_place"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "extra_right_place", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{start: originShapeShiftStart, end: originShapeEnd, before: "operation.RightID != \"\" ||", after: "false ||", sites: 1}},
			},
			{
				name: "source-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_source_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_source_type", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: originShapeShiftStart, end: originShapeEnd, before: "places[operation.SourceID] != \"U64\" ||", after: "false ||", sites: 1},
					{start: originTransferShiftStart, end: originTransferShiftEnd, before: "placeTypes[operation.SourceID] != \"U64\" ||", after: "false ||", sites: 1},
				},
			},
			{
				name: "operation-result-type-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_result_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_result_type", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{start: originShapeShiftStart, end: originShapeEnd, before: "types[operation.TypeID] != \"U64\" ||", after: "false ||", sites: 1}},
			},
			{
				name: "result-place-is-u64", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "wrong_result_place_type"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "wrong_result_place_type", marker: "ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{start: originShapeShiftStart, end: originShapeEnd, before: "places[operation.TargetID] != \"U64\" ||", after: "false ||", sites: 1},
					{start: originTransferShiftStart, end: originTransferShiftEnd, before: " || placeTypes[operation.TargetID] != \"U64\"", after: "", sites: 1},
				},
			},
			{
				name: "source-must-be-initialized", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "uninitialized_U64_source"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "uninitialized_U64_source", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{start: originTransferShiftStart, end: originTransferShiftEnd, before: "!source.initialized ||", after: "false ||", sites: 1}},
			},
			{
				name: "target-identity-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "parameter_target", "source_aliases_target"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "parameter_target", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "source_aliases_target", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{
					{start: originShapeShiftStart, end: originShapeEnd, before: "operation.TargetID == function.Parameter.ID || operation.TargetID == operation.SourceID ||", after: "false ||", sites: 1},
					{start: originShapeShiftStart, end: originShapeEnd, before: "if !freshTarget {", after: "if !freshTarget && false {", sites: 1},
				},
			},
			{
				name: "target-freshness", selector: "^TestPhase35ShiftTargetFreshness$",
				expected: phase35FreshnessFailures("TestPhase35ShiftTargetFreshness"),
				edits:    []phase35AdmissionEdit{{start: originShapeShiftStart, end: originShapeEnd, before: "if !freshTarget {", after: "if !freshTarget && false {", sites: 1}},
			},
			{
				name: "unrelated-fields", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "payload_field"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "payload_field", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{start: originShapeShiftStart, end: originShapeEnd, before: "!originScalarOperationHasNoUnrelatedFields(operation)", after: "false", sites: 1}},
			},
			{
				name: "store-and-edge-field-exclusion", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "store_field", "edge_field"),
				expected: []phase35AdmissionFailure{
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "store_field", marker: "ADMISSION_ACCEPTED:"},
					{test: "TestPhase35ShiftImmediateAdmission", subtest: "edge_field", marker: "ADMISSION_ACCEPTED:"},
				},
				edits: []phase35AdmissionEdit{{start: originShapeShiftStart, end: originShapeEnd, before: "operation.StoreTargetID != \"\" || operation.TrueEdgeID != \"\" || operation.FalseEdgeID != \"\" ||", after: "false ||", sites: 1}},
			},
			{
				name: "ConstU64-kind-exclusivity", selector: phase35AdmissionCaseSelector("TestPhase35ShiftImmediateAdmission", "unrelated_const_field"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftImmediateAdmission", subtest: "unrelated_const_field", marker: "ADMISSION_ACCEPTED:"}},
				edits:    []phase35AdmissionEdit{{before: "if operation.ConstU64 != \"\" && operation.Kind != core.OpConst && operation.Kind != core.OpShiftLeftU64 && operation.Kind != core.OpShiftRightU64 {", after: "if false {", sites: 1}},
			},
		},
	}

	checkerPeer := phase35AdmissionPeer{
		name: "check", packagePath: "./internal/compiler/check",
		sourcePath: "internal/compiler/check/scalar.go",
		selector:   "^TestPhase35ShiftAdmission$",
		noOpTests:  []string{"TestPhase35ShiftAdmission"},
		noOpSubtests: []string{
			"TestPhase35ShiftAdmission/name", "TestPhase35ShiftAdmission/computed_group",
			"TestPhase35ShiftAdmission/decimal_64", "TestPhase35ShiftAdmission/hex_64",
			"TestPhase35ShiftAdmission/maximum_U64", "TestPhase35ShiftAdmission/overflow",
			"TestPhase35ShiftAdmission/accepted/decimal_zero", "TestPhase35ShiftAdmission/accepted/decimal_63",
			"TestPhase35ShiftAdmission/accepted/hex_63",
		},
		mutations: []phase35AdmissionMutation{
			{
				name: "shift-amount-must-be-number-leaf", selector: phase35AdmissionCaseSelector("TestPhase35ShiftAdmission", "name"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftAdmission", subtest: "name", marker: "SOURCE_ADMISSION_DIAGNOSTIC_MISMATCH:"}},
				edits:    []phase35AdmissionEdit{{before: "if expr.Right == nil || expr.Right.Kind != \"number\" {", after: "if expr.Right == nil {", sites: 1}},
			},
			{
				name: "source-range-0-through-63", selector: phase35AdmissionCaseSelector("TestPhase35ShiftAdmission", "decimal_64"),
				expected: []phase35AdmissionFailure{{test: "TestPhase35ShiftAdmission", subtest: "decimal_64", marker: "SOURCE_ADMISSION_ACCEPTED:"}},
				edits: []phase35AdmissionEdit{
					{before: "if amount > 63 {", after: "if false {", sites: 1},
					{before: "amount > 63 ||", after: "false ||", sites: 1},
				},
			},
		},
	}

	peers := []phase35AdmissionPeer{corePeer, originPeer, checkerPeer}
	for index, want := range []int{12, 12, 2} {
		if got := len(peers[index].mutations); got != want {
			t.Fatalf("finite admission control count peer=%s got=%d want=%d", peers[index].name, got, want)
		}
	}
	for _, peer := range peers {
		peer := peer
		t.Run(peer.name, func(t *testing.T) {
			root := testsupport.ProjectPath()
			implementationPath := testsupport.ProjectPath(peer.sourcePath)
			original, err := os.ReadFile(implementationPath)
			if err != nil {
				t.Fatalf("read %s source: %v", peer.name, err)
			}
			sourceHash := phase35AdmissionHash(original)

			t.Run("no-op-overlay-first", func(t *testing.T) {
				output, runErr := phase35RunTransferChild(t, root, phase35TransferPeer{name: peer.name, packagePath: peer.packagePath}, implementationPath, original, peer.selector)
				if runErr != nil {
					t.Fatalf("no-op overlay failed before mutation peer=%s selector=%s: %v\n%s", peer.name, peer.selector, runErr, output)
				}
				phase35RequireAdmissionPass(t, output, peer.noOpTests, peer.noOpSubtests)
				t.Logf("NO_OP_PASS peer=%s source_sha256=%s selector=%s subtests=%d", peer.name, sourceHash, peer.selector, len(peer.noOpSubtests))
			})

			for _, mutation := range peer.mutations {
				mutation := mutation
				t.Run(mutation.name, func(t *testing.T) {
					mutated, sites, editErr := phase35ApplyAdmissionEdits(string(original), mutation.edits)
					if editErr != nil {
						t.Fatalf("build %s mutant for %s: %v", peer.name, mutation.name, editErr)
					}
					if sites == 0 || mutated == string(original) {
						t.Fatalf("%s mutant did not change any guarded source", mutation.name)
					}
					output, runErr := phase35RunTransferChild(t, root, phase35TransferPeer{name: peer.name, packagePath: peer.packagePath}, implementationPath, []byte(mutated), mutation.selector)
					if runErr == nil {
						t.Fatalf("%s mutant survived selector %s\n%s", mutation.name, mutation.selector, output)
					}
					phase35RequireAdmissionFailure(t, output, mutation.expected)
					live, readErr := os.ReadFile(implementationPath)
					if readErr != nil || phase35AdmissionHash(live) != sourceHash {
						t.Fatalf("live source changed during overlay peer=%s mutation=%s before=%s after=%s err=%v", peer.name, mutation.name, sourceHash, phase35AdmissionHash(live), readErr)
					}
					t.Logf("MUTANT_KILLED peer=%s law=%s sites=%d source_sha256=%s expected_assertions=%s", peer.name, mutation.name, sites, sourceHash, phase35AdmissionFailureNames(mutation.expected))
				})
			}

			t.Run("restored-source-hash-and-pass", func(t *testing.T) {
				live, readErr := os.ReadFile(implementationPath)
				if readErr != nil {
					t.Fatalf("read restored %s source: %v", peer.name, readErr)
				}
				if got := phase35AdmissionHash(live); got != sourceHash {
					t.Fatalf("restored source hash changed peer=%s before=%s after=%s", peer.name, sourceHash, got)
				}
				output, runErr := phase35RunTransferChild(t, root, phase35TransferPeer{name: peer.name, packagePath: peer.packagePath}, implementationPath, original, peer.selector)
				if runErr != nil {
					t.Fatalf("restored source overlay failed peer=%s: %v\n%s", peer.name, runErr, output)
				}
				phase35RequireAdmissionPass(t, output, peer.noOpTests, peer.noOpSubtests)
				t.Logf("RESTORED_PASS peer=%s source_sha256=%s selector=%s", peer.name, sourceHash, peer.selector)
			})
		})
	}
}

type phase35AdmissionPeer struct {
	name         string
	packagePath  string
	sourcePath   string
	selector     string
	noOpTests    []string
	noOpSubtests []string
	mutations    []phase35AdmissionMutation
}

type phase35AdmissionMutation struct {
	name     string
	selector string
	edits    []phase35AdmissionEdit
	expected []phase35AdmissionFailure
}

type phase35AdmissionEdit struct {
	start  string
	end    string
	before string
	after  string
	sites  int
}

type phase35AdmissionFailure struct {
	test    string
	subtest string
	marker  string
}

func phase35AdmissionCaseSelector(test string, subtests ...string) string {
	if len(subtests) == 0 {
		return "^" + test + "$"
	}
	return "^" + test + "$/^(" + strings.Join(subtests, "|") + ")$"
}

func phase35AdmissionRows(test string) []string {
	rows := []string{
		"shift_left_u64/count-0", "shift_left_u64/count-63", "shift_right_u64/count-0", "shift_right_u64/count-63",
		"missing_immediate", "leading_zero", "signed_immediate", "hex_immediate", "leading_whitespace", "trailing_whitespace",
		"count_64", "maximum_u64", "overflow", "extra_right_place", "wrong_source_type", "wrong_result_type", "wrong_result_place_type",
		"parameter_target", "source_aliases_target", "uninitialized_U64_source", "store_field", "edge_field", "callee_field", "payload_field", "foreign_field", "unrelated_const_field",
	}
	qualified := make([]string, len(rows))
	for index, row := range rows {
		qualified[index] = test + "/" + row
	}
	return qualified
}

func phase35FreshnessRows(test string) []string {
	rows := []string{
		"shift_left_u64/prior_constant", "shift_left_u64/prior_copy", "shift_left_u64/prior_mutable_store", "shift_left_u64/prior_shift",
		"shift_right_u64/prior_constant", "shift_right_u64/prior_copy", "shift_right_u64/prior_mutable_store", "shift_right_u64/prior_shift",
	}
	qualified := make([]string, len(rows))
	for index, row := range rows {
		qualified[index] = test + "/" + row
	}
	return qualified
}

func phase35FreshnessFailures(test string) []phase35AdmissionFailure {
	rows := phase35FreshnessRows(test)
	failures := make([]phase35AdmissionFailure, len(rows))
	for index, row := range rows {
		parts := strings.SplitN(row, "/", 2)
		failures[index] = phase35AdmissionFailure{test: parts[0], subtest: parts[1], marker: "ADMISSION_ACCEPTED:"}
	}
	return failures
}

func phase35ApplyAdmissionEdits(source string, edits []phase35AdmissionEdit) (string, int, error) {
	totalSites := 0
	for _, edit := range edits {
		start := 0
		if edit.start != "" {
			index := strings.Index(source, edit.start)
			if index < 0 {
				return source, totalSites, fmt.Errorf("start marker not found: %q", edit.start)
			}
			start = index
		}
		end := len(source)
		if edit.end != "" {
			searchFrom := start
			if edit.start != "" {
				searchFrom += len(edit.start)
			}
			index := strings.Index(source[searchFrom:], edit.end)
			if index < 0 {
				return source, totalSites, fmt.Errorf("end marker %q not found after %q", edit.end, edit.start)
			}
			end = searchFrom + index
		}
		if start > end {
			return source, totalSites, fmt.Errorf("invalid edit region start=%d end=%d", start, end)
		}
		region := source[start:end]
		count := strings.Count(region, edit.before)
		if count != edit.sites {
			return source, totalSites, fmt.Errorf("guard site count for %q is %d, want %d", edit.before, count, edit.sites)
		}
		region = strings.ReplaceAll(region, edit.before, edit.after)
		source = source[:start] + region + source[end:]
		totalSites += count
	}
	return source, totalSites, nil
}

func phase35RequireAdmissionPass(t *testing.T, output string, tests, subtests []string) {
	t.Helper()
	for _, test := range tests {
		if !strings.Contains(output, "=== RUN   "+test) || !strings.Contains(output, "--- PASS: "+test) {
			t.Fatalf("no-op/restored child did not run and pass %q:\n%s", test, output)
		}
	}
	for _, test := range subtests {
		if !strings.Contains(output, "=== RUN   "+test) || !strings.Contains(output, "--- PASS: "+test) {
			t.Fatalf("no-op/restored child did not run and pass subtest %q:\n%s", test, output)
		}
	}
}

func phase35RequireAdmissionFailure(t *testing.T, output string, expected []phase35AdmissionFailure) {
	t.Helper()
	// Children run serially. Attribute an exact marker token to the last RUN
	// row, so SOURCE_ADMISSION_ACCEPTED cannot also count as ADMISSION_ACCEPTED
	// and a marker from another child cannot satisfy the intended assertion.
	markers := make(map[string][]string)
	failed := make(map[string]bool)
	current := ""
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "=== RUN   ") {
			current = strings.TrimPrefix(line, "=== RUN   ")
		}
		if strings.HasPrefix(line, "--- FAIL: ") {
			fields := strings.Fields(strings.TrimPrefix(line, "--- FAIL: "))
			if len(fields) > 0 {
				failed[fields[0]] = true
			}
		}
		for _, token := range strings.Fields(line) {
			switch token {
			case "ADMISSION_ACCEPTED:", "ADMISSION_DIAGNOSTIC_MISMATCH:", "SOURCE_ADMISSION_ACCEPTED:", "SOURCE_ADMISSION_DIAGNOSTIC_MISMATCH:":
				markers[current] = append(markers[current], token)
			}
		}
	}
	wantFailures := make(map[string]bool)
	for _, failure := range expected {
		name := failure.test + "/" + failure.subtest
		wantFailures[name] = true
		if !failed[name] || len(markers[name]) != 1 || markers[name][0] != failure.marker {
			t.Fatalf("mutant failed outside intended named assertion %q marker=%q:\n%s", name, failure.marker, output)
		}
	}
	for name := range markers {
		if !wantFailures[name] {
			t.Fatalf("mutant produced an unrelated assertion in %q:\n%s", name, output)
		}
	}
	for name := range failed {
		allowed := wantFailures[name]
		for expectedName := range wantFailures {
			allowed = allowed || strings.HasPrefix(expectedName, name+"/")
		}
		if !allowed {
			t.Fatalf("child had unrelated failing test %q:\n%s", name, output)
		}
	}
}

func phase35AdmissionFailureNames(failures []phase35AdmissionFailure) string {
	names := make([]string, len(failures))
	for index, failure := range failures {
		names[index] = failure.test + "/" + failure.subtest
	}
	return strings.Join(names, ",")
}

func phase35AdmissionHash(source []byte) string {
	hash := sha256.Sum256(source)
	return hex.EncodeToString(hash[:])
}
