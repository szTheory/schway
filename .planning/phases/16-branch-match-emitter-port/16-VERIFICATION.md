---
phase: 16-branch-match-emitter-port
verified: 2026-09-24T12:48:31Z
status: passed
score: 4/4 must-haves verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/STATE.md
  - .planning/phases/16-branch-match-emitter-port/16-01-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-01-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-02-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-02-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-03-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-03-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-04-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-04-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-05-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-06-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-06-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-07-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-07-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-08-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-08-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-09-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-09-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-10-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-10-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-11-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-11-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-12-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-12-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-13-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-13-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-14-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-14-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-15-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-15-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-16-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-16-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-17-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-17-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-18-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-18-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-19-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-19-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-20-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-20-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-21-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-21-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-22-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-22-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-23-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-23-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-24-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-24-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-25-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-25-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-26-PLAN.md
  - .planning/phases/16-branch-match-emitter-port/16-26-SUMMARY.md
  - .planning/phases/16-branch-match-emitter-port/16-CONTEXT.md
  - .planning/phases/16-branch-match-emitter-port/16-RESEARCH.md
  - .planning/phases/16-branch-match-emitter-port/16-UAT.md
  - .planning/phases/16-branch-match-emitter-port/16-SECURITY.md
  - .planning/phases/16-branch-match-emitter-port/16-VALIDATION.md
  - .planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_emitter_retirement_test.go
  - internal/compiler/cgen/cgen_n1_convergence_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase16_control_test.go
  - internal/compiler/session/session_phase16_production_paths_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/session/verification_groundedness_test.go
covered_digest: "v1:sha256:776f4a76bba8c14ac5d0d5b0ff51f72d7f5e701bed4e9f8b057c1ef366c42e76"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 4/4
  gaps_closed:
    - "G-16-21-E: scanner rejected the grouped NAT-09 selector despite its tests running; report branches now anchor independently."
  gaps_remaining: []
  regressions: []
decision_coverage:
  honored: 11
  total: 11
  not_honored: []
---

# Phase 16: Branch/Match Emitter Port Verification Report

**Phase Goal:** One emission law lowers every admissible program, instead of two laws split by a function-count guard.

**Verified:** 2026-09-24T12:48:31Z

**Status:** passed

**Re-verification:** Yes — refreshed after Plan 16-26 groundedness gap closure and the legitimate Phase 17 → 18 tracking transition. All roadmap truths and NAT-08/NAT-09 were rechecked against source, tests, UAT, and the settled roadmap/state artifacts. The fingerprint includes the current `.planning/ROADMAP.md` and `.planning/STATE.md`.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | The superseded `emitMatch`, `emitBranch`, and `emitLinear` definitions are absent, and deletion shares the dispatch-cutover commit. | ✓ VERIFIED | `go test ./internal/compiler/cgen -run '^TestSupersededEmitterDefinitionsRemoved$' -count=1` parses all production cgen Go files and asserts no superseded emitter definition remains. `TestPublicDispatchUsesOnlyEmitProgram` checks both public dispatchers call only `emitProgram`. Commit `0607486` changes the public emitter and deletes the legacy bodies in the same changeset. |
| 2 | The convergence table proves byte identity for admitted shapes and every moved golden digest has an auditable justification. | ✓ VERIFIED | `TestN1ConvergenceDifferential` compares public and direct output for five fixtures in both modes; four admitted rows require byte identity and the foreign row must refuse. `TestPhase16GoldenChangeLedger` and `TestPreviousPhaseGoldenCUnchanged` pass; `internal/compiler/core/core_test.go` pins four old/new digests, moved responsibility, structural reason, semantic witness, fixture, and disposition per row. The independent interpreter/O0/O3/O3-LTO differential test also passes. |
| 3 | NAT-09 records the three cut families, M004 landing, and the `-flto` limitation with an owning phase for each. | ✓ VERIFIED | REQUIREMENTS amendment, ROADMAP Phase 21, and `PHASE-16-DEBT.md` agree on `emitLinearForeign`, `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`, and owner P21. Each debt row states pure-Lang one-TU `-flto` inertness. `TestPhase16EmitterCutsAreAmendedAndOwned` passes, including seeded missing-family, unowned-row, mistitled-owner, and missing-prerequisite mutations. |
| 4 | Admissible source has no function-count dispatch fork, and `emitProgram` serializes a derived live-resource value. | ✓ VERIFIED | `Emit` and `EmitNative` have no `len(program.Functions) != 1` dispatch. `emitProgram` calls `deriveProgramLiveResources(program)` and passes that value to size calculation and the generated `lang_write_live_resources` writer; the string literal is emitted from the encoded derived slice. `TestProgramLiveResourcesAreDerived` and its mutation control pass. The full test suite also passes. |

The additional production-path smoke controls are also verified: `TestPhase16ProductionPathsPreserveM004Refusal` runs both an admitted-path caller and a cut-family fixture through `RunNativeCommandFile` and requires terminal refusal with no execution document; `TestPhase16ProductionBypassMutationIsKilled` proves a seeded frozen-C route is detected by the source guard. These extend the machine-checked validation coverage without adding a user-facing or subjective acceptance criterion.

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/compiler/cgen/cgen.go` | Single validated public dispatch | ✓ VERIFIED | Both exported emitters validate, then delegate directly to `emitProgram`; the same-commit cutover is visible in `0607486`. |
| `internal/compiler/cgen/cgen_program.go` | Whole-program implementation and derived resource output | ✓ VERIFIED | Substantive branch/match/program lowering, explicit refusals for cut families, and derived-resource serialization. |
| `internal/compiler/cgen/cgen_n1_convergence_test.go` | Convergence proof | ✓ VERIFIED | Five checked fixtures and both APIs; byte equality is asserted on every admitted row, with cut refusal asserted on the excluded row. |
| `internal/compiler/cgen/cgen_emitter_retirement_test.go` | Legacy-definition removal proof | ✓ VERIFIED | Parses production cgen Go sources and fails if any superseded `emitMatch`, `emitBranch`, or `emitLinear` definition remains. |
| `internal/compiler/core/core_test.go` | Golden provenance ledger | ✓ VERIFIED | Four digest entries and anti-decay controls require coherent state, changed digests, reasons, and executable semantic witnesses. |
| `internal/compiler/session/session_phase11_differential_test.go` | Independent semantic differential | ✓ VERIFIED | Calls the direct program emitter and compares interpreter/O0/O3/O3-LTO engine documents; a seeded semantic mutation is rejected. |
| `internal/compiler/session/session_phase16_production_paths_test.go` | Production refusal and bypass proof | ✓ VERIFIED | Exercises two fixtures through the production native command path and kills a seeded frozen-C bypass selector. |
| `.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`, `PHASE-16-DEBT.md` | NAT-09 formal cut and ownership | ✓ VERIFIED | Amendment, roadmap owner, debt IDs, P21 ownership, reopening conditions, and LTO limitation are cross-checked by a mutation-backed test. |
| `.planning/phases/16-branch-match-emitter-port/16-SECURITY.md` | Phase threat register and audit | ✓ VERIFIED | 49 threats have dispositions; the auditor records 49 closed, 0 open, and verified sign-off. Accepted event-attribution risk is documented. |
| `.planning/phases/16-branch-match-emitter-port/16-VALIDATION.md` | Nyquist task map and host-specific evidence policy | ✓ VERIFIED | Frontmatter is `complete` and `nyquist_compliant: true`; plans 07–26 are mapped. Linux restrict evidence remains a prerequisite only for any future by-pointer admission, which is currently cut/refusal-only. |
| `.github/workflows/ci.yml` | Recurring integration/regression gate | ✓ VERIFIED | The macOS/Linux matrix runs `go build ./...` and `go test ./...`. Current final fresh runs of both commands passed. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `cgen.go:Emit` | `cgen_program.go:emitProgram` | validated `core.Program` | ✓ WIRED | Direct return of `emitProgram(validated.Program(), false)`. |
| `cgen.go:EmitNative` | `cgen_program.go:emitProgram` | validated `core.Program` | ✓ WIRED | Direct return of `emitProgram(validated.Program(), true)`. |
| `emitProgram` | live resource JSON | derive → size → generated writer | ✓ WIRED | One `liveResources` value feeds size accounting and the generated writer; mutation test proves serialization changes when the derivation seam changes. |
| NAT-09 amendment | M004 debt rows | requirement / roadmap / debt semantic guard | ✓ WIRED | The session guard reads and cross-validates the three authorities and rejects seeded drift. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `cgen_program.go` | `liveResources` | `deriveProgramLiveResources(program)` | Yes; its value feeds generated schema-2 JSON and output-size accounting. The currently admitted resource shapes derive an empty list; a mutation control verifies the writer consumes the value. | ✓ FLOWING |
| `cgen_program.go` | generated C | checked `core.Program` passed through `Emit` / `EmitNative` | Yes; lowering reads call graph, functions, operations, layouts, and admission facts before emitting C. | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Public/direct convergence and derived-resource serialization | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run '^(TestN1ConvergenceDifferential|TestProgramLiveResourcesAreDerived|TestProgramLiveResourceDerivationIsNotInert)$' -count=1` | exit 0 (0.341s) | ✓ PASS |
| Golden-C immutability and change ledger | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/core -run '^(TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger|TestPhase16GoldenChangeLedgerRejectsFaults)$' -count=1` | exit 0 (0.215s) | ✓ PASS |
| NAT-09 ownership and groundedness frontier | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^(TestPhase16EmitterCutsAreAmendedAndOwned|TestDebtRegistersAreWellFormed|TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty)$' -count=1` | exit 0 (1.580s) | ✓ PASS |
| Independent native semantic differential | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestPhase16DirectProgramFourTierDifferential$' -count=1` | exit 0 (2.134s) | ✓ PASS |
| Superseded-emitter deletion and production-path safety | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run '^TestSupersededEmitterDefinitionsRemoved$' -count=1`; `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^(TestPhase16ProductionBypassMutationIsKilled|TestPhase16ProductionPathsPreserveM004Refusal|TestPhase16EmitterCutsAreAmendedAndOwned|TestDebtRegistersAreWellFormed|TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty)$' -count=1` | both exit 0 (0.260s and 1.344s) | ✓ PASS |
| Re-verification regression selection | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/session -run '^(TestSupersededEmitterDefinitionsRemoved|TestN1ConvergenceDifferential|TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger|TestPhase16EmitterCutsAreAmendedAndOwned|TestDebtRegistersAreWellFormed|TestVerificationGroundednessFrontierIsPinned|TestVerificationGroundednessThreeClassesAreEmpty)$' -count=1` | exit 0 across all three packages (1.773s) | ✓ PASS |
| Full workspace regression and build | `GOCACHE=/tmp/ai-lang-gocache go build ./... && GOCACHE=/tmp/ai-lang-gocache go test ./...` | exit 0; session package completed in 86.398s; all packages passed | ✓ PASS |

### Validation Coverage Outcome

The supplemental task map in `16-VALIDATION.md` covers Plans 07–26, including all planned gap-closure tasks, and the new production-path smoke tests are compiled and run by the same Go package suite. The repository CI workflow invokes `go build ./...` and `go test ./...` on macOS and Linux; current final fresh runs of both commands passed. `16-SECURITY.md` records 49/49 threats closed or accepted with `threats_open: 0`. Focused deletion, production-refusal, bypass-mutation, NAT-09 ownership, and groundedness controls also pass. Linux restrict evidence remains a prerequisite only before any future by-pointer admission; the accepted `cut-m004` disposition keeps that path refused, so it creates no Phase 16 human UAT item.

### Probe Execution

Step 7c: SKIPPED — no Phase 16 probe is declared in its plans or summaries, and no `scripts/**/tests/probe-*.sh` file exists. Native compiler behavior is exercised by the named differential and Go tests.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| NAT-08 | Phase 16 plans claiming NAT-08, including 16-01 through 16-26 | One emission law lowers every admissible program; superseded emitters are deleted in the atomic public cutover. | ✓ SATISFIED | `TestSupersededEmitterDefinitionsRemoved` asserts zero legacy definitions; commit `0607486` contains dispatch cutover and deletions; convergence, golden ledger, and independent four-tier tests pass. |
| NAT-09 | Phase 16 plans claiming NAT-09, including 16-05, 16-06, 16-11 through 16-26 | The three no-consumer families are formally cut, assigned to M004, and disclose the `-flto` consequence. | ✓ SATISFIED | Requirement/roadmap/debt agreement plus mutation-backed ownership control and current groundedness controls pass. |

No additional Phase 16 requirement mapping is orphaned in REQUIREMENTS.md.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |

No blocking debt marker or stub was found in the inspected implementation and requirement-linked controls. Test-quality review found active convergence, digest-ledger, mutation, and independent interpreter/native comparison controls; none of the named requirement proofs is skipped. Expected values for the four-tier comparator come from the interpreter and separately compiled native lanes, rather than from the emitter's own golden hashes.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `cgen_n1_convergence_test.go` | NAT-08 | Yes | No | No | Byte value equality / explicit refusal | PASS |
| `core_test.go` | NAT-08 | Yes | No | No | Digest value and seeded ledger mutations | PASS |
| `session_phase11_differential_test.go` | NAT-08 | Yes | No | No | Four-tier behavioral comparison and seeded semantic mutation | PASS |
| `cgen_emitter_retirement_test.go` | NAT-08 | Yes | No | No | AST-level absence assertion over production emitter definitions | PASS |
| `session_phase16_production_paths_test.go` | NAT-08, NAT-09 | Yes | No | No | Production native refusal behavior plus seeded bypass mutation | PASS |
| `session_test.go` | NAT-09 | Yes | No | No | Cross-document semantic contract plus seeded mutations | PASS |
| `verification_groundedness_test.go` | NAT-09 | Yes | No | No | Exact frontier set and empty-class assertions | PASS |

**Disabled tests on requirement-linked proofs:** 0. **Circular patterns detected:** 0. **Insufficient assertions:** 0.

### Decision Coverage

The decision-coverage gate found **11/11** trackable CONTEXT.md decisions honored by shipped artifacts.

### Human Verification

None. This compiler/infrastructure phase has no visual or user-flow criteria. The only decision checkpoint, the `cut-m004` disposition, is already resolved in UAT item 5. Current code behavior, source deletion, ownership, production-path refusal, security dispositions, and recurring integration checks are covered by deterministic source checks, tests, and CI. Linux restrict evidence remains necessary only if the explicitly cut by-pointer path is reconsidered for admission.

## Gaps Summary

No must-have gaps remain. Plan 16-26 corrected the scanner-visible owner-law selector into independently anchored branches. Fresh focused emitter-retirement, production-path refusal, bypass-mutation, owner-law, and groundedness tests pass; the complete Go test suite and build also pass. `16-VALIDATION.md` is Nyquist-compliant through Plan 26; `16-SECURITY.md` records 49/49 threats closed or accepted and zero open. The macOS/Linux CI matrix runs the full build and test commands. UAT records 21/21 checks passed, including the previously recorded `cut-m004` decision and fresh full-suite check.

_Verified: 2026-09-24T12:48:31Z_

_Verifier: the agent (gsd-verifier)_
