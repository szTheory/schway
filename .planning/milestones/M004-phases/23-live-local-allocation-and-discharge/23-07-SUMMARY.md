---
phase: 23-live-local-allocation-and-discharge
plan: "07"
subsystem: native-evidence
tags: [native, Go, Clang, CI, allocation, verification]

# Dependency graph
requires:
  - phase: 23-06
    provides: [independent native observer controls and public allocation lifecycle evidence]
provides:
  - Clean-checkout public file-byte example with independently authored expected outcomes.
  - Fail-closed focused Phase 23 verifier wired into the existing macOS/Linux CI matrix.
  - Measured macOS focused-gate latency with Linux explicitly pending an executed CI receipt.
affects: [Phase 23 validation, evidence-aggregate CI]

# Actuals measured from the plan ledger at summary authoring.
actuals:
  tokens: 10559
  tasks: 2
  commits: 2
  plan_head_before: ee077e7181ac3e3cfe8a3adddd6038e2a2bf168c

# Tech tracking
tech-stack:
  added: []
  patterns:
    - One focused shell aggregate under the existing host matrix, separate from full, race, vet, and sanitizer lanes.
    - Independent checked-in expected answers for public CLI outcomes.
    - Host receipts remain incomplete until all focused groups pass on that host.

key-files:
  created:
    - examples/phase23/README.md
    - examples/phase23/file_byte.expected.json
    - scripts/verify-phase23.sh
  modified:
    - .github/workflows/ci.yml
    - .planning/LANGUAGE-MATURITY.md
    - .planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md
    - internal/compiler/native/native_app_test.go
    - internal/compiler/session/session_phase23_contract_test.go
    - internal/compiler/session/session_phase6_test.go
    - internal/compiler/session/verification_groundedness_test.go
    - testdata/phase16/public-emitter-consumers.json

key-decisions:
  - Keep the Phase 23 aggregate in the existing Ubuntu/macOS evidence matrix without repeating full, race, vet, or sanitizer suites.
  - Keep Phase 23 validation in-progress and nyquist_compliant false until the Linux CI receipt actually exists.
  - Use authored fixture constants as expected CLI answers; do not derive them from compiler events.

requirements-completed: [FFI-03, RES-04, RES-07, RES-08, RES-09]

# Coverage metadata (#1602)
coverage:
  - id: D1
    description: Public clean-checkout file-byte commands and independent answers cover both successful bytes and acquisition/use error families.
    requirement: RES-04
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang ./internal/compiler/session -run '^TestPhase23(Readme|Public|Contract)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Focused Phase 23 verification fails closed and is included once in the existing macOS/Linux host matrix.
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache sh scripts/verify-phase23.sh"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase23_contract_test.go#TestPhase23VerifierScriptContract"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_test.go#TestCIWorkflowRunsCurrentAggregateGate"
        status: pass
    human_judgment: false
  - id: D3
    description: Repository integrity inventories, source-location witnesses, and focused-command ownership agree with the current tree.
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./..."
        status: pass
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestPhase20ValidationLifecycle"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/evidence_grade_test.go#TestValidationRowGradesAreEarnedOverArchivedCorpus"
        status: pass
    human_judgment: false

# Metrics
duration: 2h 1m
completed: 2026-09-28
status: complete
---

# Phase 23 Plan 07: Live Local Allocation and Discharge Summary

**Public file-byte walkthrough with independent answers and a fail-closed focused evidence gate wired to the existing macOS/Linux CI matrix**

## Performance

- **Duration:** 2h 1m
- **Started:** 2026-09-28T12:41:11Z
- **Completed:** 2026-09-28T14:42:24Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- Added clean-checkout instructions for building and running the public file-byte example, with independent expected JSON for 0x41/0x42 success, empty/two-byte acquisition failures, and the 0x43 use failure after cleanup.
- Added a focused verifier for source admission, independent peers, model-only cases, native ABI/acquisition/observer controls, and public CLI contracts; it records host/compiler/target/revision identity and elapsed time and fails on missing tools, skipped tests, or no matching tests.
- Wired that verifier once into the existing Ubuntu/macOS `evidence-aggregate` matrix, and updated validation status without claiming an unobserved Linux receipt.
- Reconciled the known repo-wide closeout drift in the maturity census, Phase 16 emitter call inventory, and groundedness frontier; the final full Go suite passes.

## Host Evidence

The focused gate passed on macOS Darwin/arm64 with Go 1.24.0 and Apple Clang 21.0.0 targeting `arm64-apple-darwin25.6.0`. Three fresh Go build-cache runs took 18, 23, and 23 seconds (median 23 seconds). Three warm-cache runs took 8, 9, and 11 seconds (median 9 seconds). The Go module cache and host toolchain were shared. The workflow source wires the same step into Ubuntu and macOS jobs, but no Linux CI job ran during this plan; Linux remains incomplete in `23-VALIDATION.md`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Document the public file-byte command and independent answers** — `13b8900` (`docs`).
2. **Task 2: Add the focused Phase 23 host evidence gate** — `d6ac1e6` (`feat`).

**Plan metadata:** committed with this summary; its commit is recorded in the execution receipt.

## Files Created/Modified

- `examples/phase23/README.md` — clean-checkout build/run instructions and evidence scope.
- `examples/phase23/file_byte.expected.json` — independent expected public outcomes.
- `scripts/verify-phase23.sh` — focused fail-closed host evidence aggregate.
- `.github/workflows/ci.yml` — invokes the focused aggregate in the existing host matrix.
- `.planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md` — measured macOS distributions; Linux receipt pending.
- `.planning/LANGUAGE-MATURITY.md`, `testdata/phase16/public-emitter-consumers.json`, and session/native tests — reconciled closeout inventories and regression checks.

## Decisions Made

- The Phase 23 script owns only the focused allocation/discharge evidence; existing CI jobs retain ownership of full, race, vet, and sanitizer suites.
- Phase validation remains `in-progress` with `nyquist_compliant: false` until an actual Linux matrix receipt is available.
- No human UAT is needed for objectively asserted command outcomes, script contracts, or workflow wiring.

## Deviations from Plan

**1. [Rule 3 - Blocking] Reconciled repository-wide integrity closeout failures**

- **Found during:** Task 2, full-suite closeout.
- **Issue:** The 23-06 closeout findings still caused repo-wide tests to fail: maturity corpus/guard counts were stale, Phase 16 emitter call locations had moved, the new Phase 23 validation commands lacked pinned groundedness ownership, and an existing Clang skip lacked a source-witness citation.
- **Fix:** Refreshed the current census and call locations, registered the seven Phase 23 validation commands with their P23 landing owner, and cited the existing public probe at the Clang skip site.
- **Files modified:** `.planning/LANGUAGE-MATURITY.md`, `testdata/phase16/public-emitter-consumers.json`, `internal/compiler/session/verification_groundedness_test.go`, `internal/compiler/native/native_app_test.go`.
- **Verification:** `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` passed after reconciliation.
- **Committed in:** `d6ac1e6`.

**Total deviations:** 1 auto-fixed (Rule 3: blocking repository integrity failures). **Impact:** Closeout failures were repaired in scope; no unrelated source behavior changed.

## Issues Encountered

- The first sandboxed script invocation could not write to Go's default build-cache directory. Re-running with `GOCACHE=/tmp/ai-lang-verification-gocache` passed; CI uses its existing writable runner cache.
- The repo-wide suite rejected a `draft` validation document. A `partial` status would enroll its ungraded table in the Phase 16 corpus digest, so the document remains `in-progress`, with `nyquist_compliant: false` and the Linux receipt explicitly pending. The lifecycle and archived-corpus checks pass in that state.
- No Linux workflow execution was available in this run. The matrix wiring is present, but no Linux evidence is claimed.

## User Setup Required

None.

## Cleanup Scope Warnings

The worktree merge reported these six files outside the plan's declared file list. Each was needed to close an observed integrity or documentation dependency:

- `.planning/LANGUAGE-MATURITY.md` — refreshed the source-derived program and guard census required by the project's living-document workflow.
- `.planning/REQUIREMENTS.md` — marked FFI-03, RES-04, RES-07, RES-08, and RES-09 complete after their verification records passed.
- `internal/compiler/native/native_app_test.go` — added the source witness to an existing Clang-dependent test skip so the repository's suppression-witness check passes.
- `internal/compiler/session/session_phase6_test.go` — extended the CI workflow contract to pin the new aggregate step.
- `internal/compiler/session/verification_groundedness_test.go` — pinned the seven Phase 23 validation commands to their groundedness owner.
- `testdata/phase16/public-emitter-consumers.json` — refreshed two emitter source locators that moved as Phase 23 tests were added.

## Next Phase Readiness

Plan 23-07 implementation and local verification are complete. Phase 23 host validation remains in progress until the existing Linux CI lane produces a passing focused receipt; do not treat the macOS receipt or workflow wiring as Linux evidence.

## Self-Check: PASSED

- Task commits `13b8900` and `d6ac1e6` exist.
- Required example, focused script, CI workflow, and validation files exist.
- `git diff --check` passed before summary creation; `STATE.md` and `ROADMAP.md` are unchanged.

---
*Phase: 23-live-local-allocation-and-discharge*
*Completed: 2026-09-28*
