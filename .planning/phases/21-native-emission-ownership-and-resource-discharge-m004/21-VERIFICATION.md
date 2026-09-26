---
phase: 21-native-emission-ownership-and-resource-discharge-m004
verified: 2026-09-26T02:55:14Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - .planning/PROJECT.md
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/UNREACHABLE-CLAIMS.md
  - .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md
  - .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
  - .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md
  - .planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md
  - .planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-01-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-01-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-02-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-02-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-03-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-03-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-04-PLAN.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-04-SUMMARY.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-CONTEXT.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESEARCH.md
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json
  - .planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-VALIDATION.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/core/core.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interptestdirect/interptestdirect.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/session/session_phase16_frozen_evidence_external_test.go
  - internal/compiler/session/session_phase21_contract_test.go
  - internal/compiler/session/session_phase21_lto_evidence_test.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/witness_registry_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
covered_digest: "v1:sha256:aa9425fc46b3fa62195fa0f2b25cba1638b273ab9980d31d10516a9d2958c311"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 21: Native Emission Ownership and Resource Discharge — Verification

**Phase Goal:** Design checked resource discharge and foreign-boundary ownership contracts required before any M003-cut emitter family can be reconsidered.

**Verified:** 2026-09-26T02:55:14Z  
**Status:** passed  
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | The contract enumerates the seven modeled foreign exits, records dispositions and required proof/reasons, and refuses incomplete or over-admitting mutations. | ✓ VERIFIED | `21-RESOURCE-DISCHARGE-CONTRACT.json` lists normal return, error return, unwind, nonlocal transfer, defect, cancellation, and process termination. `TestPhase21ResourceDischargeContract` passed with mutation subtests for missing/unknown/duplicate exit, bad disposition, missing discharge evidence, admitted family, and missing family. |
| 2 | The three M003-cut emitter families remain unadmitted; public lowering retains one production authority and unsupported foreign/by-pointer program shapes remain refused. | ✓ VERIFIED | `production_admission` is false for D-16-11..13. `TestPublicDispatchUsesOnlyEmitProgram`, `TestPhase21LegacyEmitterBodiesRetired`, and `TestForeignResourceLedgerEmitterRemainsRefused` passed. `emitProgram` retains explicit foreign and by-pointer rejection before serialization. The heuristic key-link query missed paths assembled through `testsupport.ProjectPath`; direct source inspection confirmed the test reads and validates the checked-in contract. |
| 3 | One directly emitted multi-function fixture is compared across interpreter, `-O0`, `-O3`, and `-O3 -flto`, with bounded provenance and an independent comparator control. | ✓ VERIFIED | Ran `GOCACHE=/private/tmp/ai-lang-phase21-gocache go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v`: PASS. Observed `darwin/arm64`, `/usr/bin/clang`, Apple Clang 21.0.0, fixture SHA-256 `a86d9e90…f3e556`, emitted-C SHA-256 `d096fca6…6416`, and all-pairs semantic equality. `TestPhase16EmitterPortSemanticGuardIsNotInert` also passed. The receipt correctly limits the result to this fixture, host, compiler, and lanes; it makes no optimizer-activity, performance, cleanup, or cross-host claim. |
| 4 | Historical debt rows and generated claims agree with the retired code and bounded measurement. | ✓ VERIFIED | `TestDebtRegistersAreWellFormed`, `TestUnreachableClaimsViewIsCurrent`, and `TestPhase16EmitterCutsAreAmendedAndOwned` passed. D-11-02/D-12-36 record retirement evidence; D-16-11..13 retain family-specific gates; D-14-45 points to the measured receipt without expanding its scope. |
| 5 | Cheap contract, receipt-binding, retirement, refusal, and evidence-grade checks recur in the normal Go suite; the expensive compiler comparison remains intentionally opt-in. | ✓ VERIFIED | Untagged `TestPhase21ResourceDischargeContract` and `TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison` passed. `TestValidationRowGradesAreEarnedOverArchivedCorpus` passed all 21 validation documents. Latest phase-wide `go test ./... -count=1 -p 2` completed with exit 0 (executor run). GSD phase-completeness check also passed: 4/4 summaries, no orphan plans. |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `21-RESOURCE-DISCHARGE-CONTRACT.json` | Checked exit and family contract | ✓ VERIFIED | Substantive versioned JSON; normal/error returns carry discharge plus future evidence, opaque paths are refused or outside guarantees, all cut families remain unadmitted. Loaded and mutation-checked by the recurring session test. |
| `session_phase21_contract_test.go` | Structural guard wired to contract | ✓ VERIFIED | Exists and substantively validates contract and mutation cases. Reads the exact project artifact via `testsupport.ProjectPath`; active under untagged Go tests. |
| `cgen.go` and `cgen_program_test.go` | Single dispatch and retired bodies remain absent | ✓ VERIFIED | Public API routes through `emitProgram`; AST regression guard checks dispatch and absent legacy declarations. Refusal tests pass. |
| `session_phase21_lto_evidence_test.go` and `21-LTO-EVIDENCE.md` | One-shot emitted-fixture comparison and bounded receipt | ✓ VERIFIED | Tagged integration test was run in this verification. Receipt binds compiler, host, fixture/emitted digests, flags, lanes, and evidence ceiling. Untagged binding test preserves discoverability. |
| Debt records and `.planning/UNREACHABLE-CLAIMS.md` | Source rows and derived view stay aligned | ✓ VERIFIED | Parser, generated-view, ownership, and validation-grade checks pass. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `session_phase21_contract_test.go` | `21-RESOURCE-DISCHARGE-CONTRACT.json` | `testsupport.ProjectPath` + `os.ReadFile` + validator | ✓ WIRED | Manually traced after the generic key-link heuristic did not resolve the dynamically assembled path; focused mutation test passed. |
| `cgen_program_test.go` | `cgen.go` | Go AST parse of production source | ✓ WIRED | Dispatch and absence checks passed. |
| `session_phase21_lto_evidence_test.go` | Phase 14 fixture and shared four-tier runner/comparator | direct emission + existing runner and `Phase5CompareEngines` | ✓ WIRED | Tagged run compiled/executed and compared the four lanes; recorded hashes matched receipt. |
| `PHASE-16-DEBT.md` | Contract and current refusal witness | named witness + contract reference | ✓ WIRED | `TestPhase16EmitterCutsAreAmendedAndOwned` passed; each family keeps separate reopening requirements. |

### Data-Flow Trace (Level 4)

N/A — this compiler-contract phase has no rendered or user-facing data flow. The emitted-C integration test traces fixture input through the real emitter and native execution comparator as recorded above.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Contract mutation refusal | `go test ./internal/compiler/session -run '^TestPhase21ResourceDischargeContract$' -count=1 -v` | All seven seeded mutation subtests passed. | ✓ PASS |
| Public dispatch and legacy-body retirement/refusal | `go test ./internal/compiler/cgen -run '^TestPhase21LegacyEmitterBodiesRetired$' -count=1 -v` plus focused checks of `TestForeignResourceLedgerEmitterRemainsRefused` and `TestPublicDispatchUsesOnlyEmitProgram` | All three named tests passed. | ✓ PASS |
| Emitted multi-function semantics under compiler lanes | `go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v` | All-pairs semantic equality on recorded Darwin/arm64 + Clang configuration. | ✓ PASS |
| Comparator detects seeded disagreement | `go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v` | Independent negative control passed. | ✓ PASS |

### Probe Execution

No shell probe scripts are declared by Phase 21. The phase-declared tagged compiler comparison was run directly and passed; see the behavioral spot-check and receipt.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| NAT-09 | Roadmap / Phase 16 amendment; no Phase 21 plan requirement IDs | Three M003-cut emitter families remain formally owned with family-specific reopening gates and stated one-TU/`-flto` boundary. | ✓ SATISFIED | Contract and Phase 16 debt retain all three refusals and distinct conditions; `TestPhase16EmitterCutsAreAmendedAndOwned` passed. Phase 21 plan `requirements` arrays are empty; no orphan Phase 21 requirement IDs were found. |

### Test Quality Audit

| Test File | Linked criterion | Active | Circular | Assertion strength | Verdict |
|---|---|---:|---:|---|---|
| `session_phase21_contract_test.go` | Contract completeness/refusal | Yes | No | Value and seeded-mutation rejection | PASS |
| `cgen_program_test.go` | Sole dispatch, absence, refusal | Yes | No | AST structure plus executable refusal checks | PASS |
| `session_phase21_lto_evidence_test.go` | Four-lane semantic agreement | Yes (opt-in tag) | No | Actual compiled/executed outputs compared against interpreter and each other | PASS |
| `session_phase21_contract_test.go` receipt guard | Recurring evidence linkage | Yes (untagged) | No | Exact source/receipt fields and bounded-claim assertions | PASS |

No disabled Phase 21 acceptance tests or circular expected-value capture were found. The comparator's independent seeded-disagreement test passed.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `session_phase5_alias.go` | 284 | “operational placeholder” in a comment describing a diagnostic fallback when Clang cannot be located | Info | Not an implementation stub; the surrounding comment explicitly says the value is diagnostic and non-gating. No TBD/FIXME/XXX/TODO/HACK markers were found in changed implementation files. |

### Decision Coverage

All trackable CONTEXT.md decisions are honored by shipped artifacts (3/3); no unhonored decisions.

### Human Verification Required

N/A — infrastructure/compiler-contract phase with no user-facing behavior. All phase acceptance evidence is automated. Runtime cleanup witnesses and any emitter admission are explicitly future evidence requirements, outside this phase's design-only goal; their absence is not represented as runtime proof.

### Gaps Summary

No goal-blocking gaps. All five observable truths are verified. ROADMAP §Phase 21 now records this work as complete while preserving the no-emitter-admission boundary. No human UAT is needed.

---

_Verified: 2026-09-26T02:55:14Z_  
_Verifier: gsd-verifier_
