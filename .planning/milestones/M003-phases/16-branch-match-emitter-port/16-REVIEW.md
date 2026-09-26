---
phase: 16-branch-match-emitter-port
reviewed: 2026-09-26T12:59:22Z
depth: standard
files_reviewed: 14
files_reviewed_list:
  - .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/session_phase11_gate.go
  - internal/compiler/session/session_phase11_gate_test.go
  - internal/compiler/session/session_phase16_emitter_inventory_test.go
  - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  - internal/compiler/session/session_phase6_pin_test.go
  - internal/compiler/session/witness_registry_test.go
  - testdata/phase16/file-frozen-evidence.json
  - testdata/phase16/historical/phase11_gate_n_two.fixture
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
  - internal/compiler/session/session_test.go
findings:
  critical: 0
  warning: 1
  info: 0
  total: 1
status: issues_found
---

# Phase 16: Code Review Report

**Reviewed:** 2026-09-26T12:59:22Z
**Depth:** standard
**Files Reviewed:** 14
**Status:** issues_found

## Summary

The existing review covered the Phase 16 Plans 21–24 gap-closure range (`1b5fa64^..04d6154`) across 13 files and found no issues. This update adds a standard-depth review of `internal/compiler/session/session_test.go` for Plan 25.

The previously reviewed by-pointer route proves the exact current public refusal before reading frozen C, and the frozen artifact is bound to the checked fixture, canonical program, artifact digest, and refusal. The inventory derives call sites from the Go AST and requires an exact registry bijection. The validation grader selects completed documents plus only draft/planned tables already carrying the Grade schema; its live record is bound to the current requested pairs and completion witnesses. The Plan 25 guard checks expected debt rows and several seeded regressions, but its document-wide string checks do not bind the owner and family assertions to their intended roadmap section and requirements amendment.

## Narrative Findings (AI reviewer)

### WR-01: Ownership guard does not bind text to its required sections

**File:** `internal/compiler/session/session_test.go:4631`
**Issue:** `phase16EmitterCutProblems` checks the NAT-09 header and marker in `requirements`, but later checks family names and the Phase 21 owner title with document-wide `strings.Contains` calls (lines 4634 and 4645). Likewise, it checks that M003 precedes M004 and that the Phase 21 title exists somewhere in the roadmap, without proving that the titled Phase 21 entry is inside M004 and after the M003 completion boundary. As a result, moving family names outside the amendment or moving the titled owner elsewhere while leaving the M004 heading intact can satisfy the guard even though the stated linkage is broken.
**Fix:** Extract the NAT-09 amendment section and the relevant M004 roadmap section (bounded by the next headings), then assert the family names and owner title within those sections and verify the Phase 21 heading's position relative to the M003 completion marker.

## Verification Evidence

Passed:

```text
GOCACHE=/tmp/ai-lang-review-gocache go test ./internal/compiler/session -run '^(TestPhase11ZeroAttributeGate|TestPhase11GateIsNonVacuous|TestPhase11GateFailsAtNZero|TestPhase11GateCountsAdjacentWouldCarryFunctions|TestPhase11GateMutationKill|TestPhase11SuppressionIsDiffLocal|TestPhase11ByPointerRefusalFirstEvidence|TestPhase16PublicEmitterConsumerInventory|TestPhase16EmitterInventoryMutationControls|TestPhase16EmitterInventoryRefusalWitnessesResolve|TestPhase16M004ProvenanceRegistryRejectsFaults|TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram|TestPhase16Phase11FrozenEvidenceRejectsProvenanceFaults|TestAllPrimaryValidationRowsExcludeDrafts|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestRunRecordCarriesACompletionWitness|TestRunRecordCompletenessGuardIsNotInert|TestLaneSchemaLiteralSiteCountIsPinned)$' -count=1 -v
```

Also verified JSON syntax for the three changed evidence manifests and record metadata, found no remaining references to the retired `multi_function_gate_n_two.lang` fixture, and found no whitespace errors in the reviewed commit range.

---

_Reviewed: 2026-09-26T12:59:22Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
