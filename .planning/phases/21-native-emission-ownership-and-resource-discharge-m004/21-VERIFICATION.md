---
phase: 21-native-emission-ownership-and-resource-discharge-m004
verified: 2026-09-26T21:40:34Z
status: gaps_found
score: 2/5 must-haves verified
covered_files:
  - .planning/PROJECT.md
  - .planning/ROADMAP.md
  - .planning/STATE.md
  - .planning/UNREACHABLE-CLAIMS.md
  - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
  - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
  - .planning/milestones/M003-REQUIREMENTS.md
  - .planning/milestones/M003-ROADMAP.md
  - .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
  - .planning/milestones/M003-phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-CONTEXT.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-01-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-01-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-02-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-02-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-03-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-03-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-04-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-04-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VALIDATION.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-UAT.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/core/core.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interptestdirect/interptestdirect.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  - internal/compiler/session/session_phase18_payload_test.go
  - internal/compiler/session/session_phase21_contract_test.go
  - internal/compiler/session/session_phase21_lto_evidence_test.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/witness_registry_test.go
  - internal/compiler/session/evidence_grade_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
covered_digest: "v1:sha256:fda1730b65f6650f960405378033e9b62f71e6ee542fbe56e447479a8b4a961e"
behavior_unverified: 1
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 5/5
  gaps_closed: []
  gaps_remaining:
    - "Archive-dependent validation and refusal guards no longer resolve their inputs."
  regressions:
    - "The archived M003 tree exposes pre-archive path and corpus assumptions in recurring guards."
behavior_unverified_items:
  - "The by-pointer disposition test cannot reach its behavioral assertions until it resolves the archived Phase 16 summary path."
gaps:
  - truth: "Archive-dependent Phase 16 and corpus guards pass against the current M003 tree."
    status: failed
    reason: "Current checks depend on paths and corpus fingerprints that changed when M003 artifacts were archived."
    artifacts:
      - path: "internal/compiler/cgen/cgen_program_test.go"
        issue: "TestProgramBorrowedByPointerDisposition still reads the old .planning/phases/ path."
      - path: "internal/compiler/session/session_test.go"
        issue: "TestPhase16EmitterCutsAreAmendedAndOwned still reads the removed root .planning/REQUIREMENTS.md."
      - path: "internal/compiler/session/evidence_grade_test.go"
        issue: "The checked-in validation corpus pair digest differs from the current requested corpus."
      - path: "internal/compiler/session/verification_groundedness_test.go"
        issue: "The reconciliation and landing-phase inventory no longer matches findings in the archived tree."
    missing:
      - "Update the affected path resolvers and reconciliation data for M003 archive locations."
      - "Refresh the validation corpus pair record and rerun its guard."
---

# Phase 21: Native Emission Ownership and Resource Discharge — Verification

**Phase Goal:** Design checked resource-discharge and foreign-boundary ownership contracts required before any M003-cut emitter family can be reconsidered.

**Verified:** 2026-09-26 21:40:34 UTC
**Status:** gaps_found
**Re-verification:** Yes — refreshed against the post-M003-archive tree. The completed UAT was preserved.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | The checked contract classifies all seven modeled foreign exits and rejects incomplete or over-admitting mutations. | ✓ VERIFIED | `TestPhase21ResourceDischargeContract` passed with mutation controls for missing, unknown, duplicate, and misclassified exits, missing discharge evidence, and attempted family admission. The JSON keeps all D-16-11..13 families unadmitted. |
| 2 | The three cut emitter families remain refused, and public lowering has one production authority. | ◷ PARTIAL | `TestPublicDispatchUsesOnlyEmitProgram`, `TestPhase21LegacyEmitterBodiesRetired`, and `TestForeignResourceLedgerEmitterRemainsRefused` passed. The whole cgen package currently fails in `TestProgramBorrowedByPointerDisposition` because it tries to read the archived path `.planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md`; the live summary is now under `.planning/milestones/M003-phases/`. |
| 3 | The emitted multi-function fixture agrees across interpreter, `-O0`, `-O3`, and `-O3 -flto`, with bounded provenance. | ✓ VERIFIED | Re-ran the tagged comparison successfully. It produced the same emitted-C digest `d096fca69195eb63b09566387690a7b40fdc9c364f55b0b4a24209a895ff6416`, fixture digest, Darwin/arm64 host, and Apple Clang 21.0.0 recorded in the receipt. The independent comparator negative control passed. |
| 4 | Historical emitter debt rows and the derived claims view preserve family-specific gates and current ownership. | ◷ PARTIAL | `TestDebtRegistersAreWellFormed` and `TestUnreachableClaimsViewIsCurrent` passed. `TestPhase16EmitterCutsAreAmendedAndOwned` cannot reach its assertions because it reads removed `.planning/REQUIREMENTS.md`; M003 requirements are archived at `.planning/milestones/M003-REQUIREMENTS.md`, and Phase 21 has no ratified M004 requirement IDs. |
| 5 | The recurring evidence-grade and groundedness controls match the archived project corpus. | ✗ NOT VERIFIED | `TestValidationRowGradesAreEarnedOverArchivedCorpus` fails because the checked-in corpus pair digest differs from the current requested corpus. `TestVerificationGroundednessThreeClassesAreEmpty` reports unresolved findings and landing-phase entries after M003 artifacts moved under `.planning/milestones/M003-phases/`. These checks passed in the original plan run but are stale against the archived tree. |

**Score:** 2/5 truths verified; 2 partial; 1 not verified.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `21-RESOURCE-DISCHARGE-CONTRACT.json` | Checked exit and family contract | ✓ VERIFIED | Versioned, design-only contract. Supported paths carry discharge and future evidence rules; opaque paths are refused or outside cleanup guarantees. |
| `session_phase21_contract_test.go` | Structural guard wired to contract | ✓ VERIFIED | The untagged test reads the checked-in JSON and rejects seeded mutations. |
| `cgen.go` and `cgen_program_test.go` | Single dispatch and retired bodies remain absent | ◷ PARTIAL | The dedicated dispatch, retirement, and foreign-ledger refusal probes pass. The full cgen suite is blocked by an archive-relative summary path in a separate by-pointer disposition test. |
| `session_phase21_lto_evidence_test.go` and `21-LTO-EVIDENCE.md` | One-shot emitted-fixture comparison and bounded receipt | ✓ VERIFIED | Re-run passed, and current emitted-C digest matches the committed receipt. |
| Debt records and `.planning/UNREACHABLE-CLAIMS.md` | Source rows and derived view stay aligned | ◷ PARTIAL | Register parsing and derived-view equality pass; the ownership regression test still reads the deleted root requirements file. |
| Validation corpus record and groundedness register | Current archived corpus snapshot | ✗ NOT VERIFIED | The checked-in pair digest and reconciliation inventory need refresh for the M003 archive paths. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `session_phase21_contract_test.go` | `21-RESOURCE-DISCHARGE-CONTRACT.json` | `testsupport.ProjectPath` + JSON decoder + mutation validator | ✓ WIRED | Focused test passed. |
| `cgen_program_test.go` | `cgen.go` | Go AST checks | ✓ WIRED | Sole dispatch and retired-body assertions passed; by-pointer disposition integration test is blocked by its stale fixture path. |
| `session_phase21_lto_evidence_test.go` | Phase 14 fixture and comparator | Direct emission + shared four-tier runner + all-pairs comparison | ✓ WIRED | Tagged comparison passed and receipt digest remained unchanged. |
| `PHASE-16-DEBT.md` | Contract and refusal witness | Named witness + contract reference | ◷ PARTIAL | Debt register parses, but its anti-decay ownership test cannot read the removed root requirements file. |

### Data-Flow Trace

Not applicable — no user-facing data flow. The emitted-C test traces its checked fixture through direct emission, native execution, and the existing semantic comparator.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Contract completeness and mutation refusal | `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract$' -count=1 -v` | All seeded contract mutations were rejected. | ✓ PASS |
| Current emitted multi-function comparison | `go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v` | Four-lane semantic equality; current emitted-C digest matches the receipt. | ✓ PASS |
| Sole dispatcher, retired bodies, foreign-ledger refusal | `go test ./internal/compiler/cgen -run '^(TestPhase21LegacyEmitterBodiesRetired|TestPublicDispatchUsesOnlyEmitProgram|TestForeignResourceLedgerEmitterRemainsRefused)$' -count=1 -v` | All three focused probes passed. | ✓ PASS |
| Archived validation corpus snapshot | `go test ./internal/compiler/session -run '^TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1` | Checked-in corpus pair digest does not match the current requested corpus. | ✗ FAIL |
| Archived debt ownership guard | `go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1` | Cannot open removed `.planning/REQUIREMENTS.md`. | ✗ FAIL |
| Groundedness closure | `go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1` | Unresolved R1/R2/R3 findings and R2b landing entries remain after archiving. | ✗ FAIL |
| Full cgen package suite | `go test ./internal/compiler/cgen -count=1` | Fails in `TestProgramBorrowedByPointerDisposition` while opening the pre-archive Phase 16 summary path. | ✗ FAIL |

### Requirements Coverage

Phase 21 plans declare no requirement IDs, and `phase_req_ids` is null. M004 requirements have not been ratified. The old report's NAT-09 linkage to the root M003 requirements file was removed from this refresh; no unratified M004 requirement is claimed as satisfied.

### Decision Coverage

The three decisions in `21-CONTEXT.md` remain honored: contract checks stay structural, unsupported exits remain refused or outside cleanup guarantees, and LTO evidence is bounded to the named fixture/compiler/host.

### Human Verification Required

None. `21-UAT.md` remains `status: complete` with all seven automated checks passed. It was not rerun or modified. Runtime cleanup witnesses and emitter admission remain explicit future requirements.

### Gaps Summary

Phase 21's contract and scoped LTO evidence remain valid, but archived-artifact consumers no longer pass. Refresh the validation-corpus record, update path resolution and groundedness reconciliation for M003's archived files, then rerun the affected guards before refreshing Phase 21 verification again. Do not rerun UAT.

---

_Verified: 2026-09-26T21:40:34Z_
_Verifier: Codex (inline execution fallback)_
