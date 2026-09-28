---
phase: 23-live-local-allocation-and-discharge
plan: 05
subsystem: compiler/interpreter
tags: [interpreter, ownership, foreign-outcomes, discharge, evidence]

requires:
  - phase: 23-01
    provides: Checked PathToken acquire, borrow, and release operations with per-operation foreign contracts
provides:
  - Deterministic model-only foreign outcomes with local owner lifecycle checks
  - Resource discharge contract v2 distinguishing release, borrow, and future transfer
affects: [23-02, 23-03, 23-04, 23-06, Phase 24 ownership transfer]

actuals:
  tokens: 10062
  tasks: 2
  commits: 4
  plan_head_before: 581754eea346360ace666787398a95d1a5bafd7d

tech-stack:
  added: []
  patterns:
    - Foreign model outcomes are keyed by checked operation ID and never access host files
    - Model evidence scope is separate from native physical IO and cleanup receipts

key-files:
  created:
    - .planning/phases/23-live-local-allocation-and-discharge/23-RESOURCE-DISCHARGE-CONTRACT.json
    - internal/compiler/session/session_phase23_contract_test.go
    - internal/compiler/session/session_phase23_model_test.go
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go
    - internal/compiler/session/session_app_verify.go

key-decisions:
  - "Supply deterministic outcomes independently for each checked foreign operation; never treat a model result as host IO."
  - "Keep release local and consuming, borrow obligation-preserving, and transfer contract-only until Phase 24."
  - "Preserve the archived Phase 21 contract and leave shared requirements pending while sibling plans remain incomplete."

requirements-completed: [FFI-03, RES-04, RES-07, RES-09]

coverage:
  - id: D1
    description: "Per-operation modeled byte outcomes and acquire/use failures follow local owner semantics without host IO or physical cleanup claims."
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/session -run '^TestPhase23Model' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The successor contract pins local release, preserving borrow, Phase 24 transfer, error cleanup order, and separate structural, model, and native evidence scopes."
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase23Contract' -count=1"
        status: pass
    human_judgment: false

duration: 3h 29m
completed: 2026-09-28
status: complete
---

# Phase 23 Plan 05: Modeled outcomes and discharge contract Summary

**Per-operation foreign outcomes now produce deterministic byte and cleanup models, with a v2 contract separating model receipts from native evidence.**

## Performance

- **Duration:** 3h 29m
- **Started:** 2026-09-28T03:01:13Z
- **Completed:** 2026-09-28T06:30:24Z
- **Tasks:** 2
- **Files modified:** 6

## Accomplishments

- Added deterministic supplied outcomes keyed by each checked foreign operation. The model produces 65 and 66 for the expected byte fixtures, creates no owner on failed acquisition, and retains an acquired owner through modeled use failure until local release.
- Scoped interpreter model and replay reports explicitly as model-only; they claim neither host file IO nor physical allocation cleanup. Path-based source replay remains refused where independent host evidence is unavailable.
- Published the v2 discharge contract: release consumes one local obligation, borrow preserves the same owner’s obligation, and transfer preserves a live resource under a new owner but remains contract-only for Phase 24.
- Added strict typed contract decoding and mutation checks for missing, duplicate, unknown, and overclaimed transition/evidence statements. The archived Phase 21 artifact is unchanged.

## Task Commits

1. **Task 1 RED: Model per-operation outcomes** — `09774bc` (`test(23-05): cover modeled file-byte outcomes`)
2. **Task 1 GREEN: Implement model outcomes and evidence scope** — `49ea872` (`feat(23-05): model per-operation foreign outcomes`)
3. **Task 2: Publish the successor discharge contract** — `6987161` (`feat(23-05): publish resource discharge successor contract`)
4. **Task 2 validation hardening: Reject incomplete evidence claims** — `1b3abf4` (`fix(23-05): reject incomplete evidence claims`)

## Files Created/Modified

- `.planning/phases/23-live-local-allocation-and-discharge/23-RESOURCE-DISCHARGE-CONTRACT.json` — Contract v2 ownership transitions, acquisition/error semantics, and evidence scopes.
- `internal/compiler/interp/interp.go` — Operation-keyed model runner and local owner accounting.
- `internal/compiler/interp/interp_test.go` — Independent byte and failure lifecycle cases.
- `internal/compiler/session/session_app_verify.go` — Explicit replay evidence-scope reporting.
- `internal/compiler/session/session_phase23_model_test.go` — Model/report boundary assertions.
- `internal/compiler/session/session_phase23_contract_test.go` — Strict contract decoder and mutation controls.

## Decisions Made

- Use caller-supplied semantic outcomes per checked operation; the model never opens the path token or performs native allocation/free.
- Keep Phase 23’s modeled local release and borrow semantics distinct from future ownership transfer.
- Keep `REQUIREMENTS.md` unchanged because `requirements.ready-ids` reported 0/4 ready; sibling Phase 23 plans still share these IDs.

## Deviations from Plan

None - plan executed as specified. Contract decoder hardening ensures an omitted false-valued evidence claim cannot silently pass through Go zero values.

## Issues Encountered

- The RED evidence checker expects TAP, while Go’s test runner emits Go test output. The targeted failing assertion was preserved and adapted for the checker; it returned `RED_EVIDENCE_OK` before the implementation commit.
- The sandbox initially denied writing the shared Git index for Task 2. The already-authorized commit succeeded after filesystem escalation; no working files were lost.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The model and successor contract are available to the remaining Phase 23 plans. This plan establishes semantic model behavior only and does not claim physical host IO or cleanup evidence. The four shared requirements remain pending until their sibling plans finish.

## Self-Check: PASSED

- All six created or modified files exist.
- Task commits `09774bc`, `49ea872`, `6987161`, and `1b3abf4` are present in the worktree history.
- Both focused plan verification commands pass.
- The archived Phase 21 contract has no diff from the plan’s base revision.
- No TODO, FIXME, placeholder, “coming soon”, or “not available” stubs were found in the plan’s changed files.

---
*Phase: 23-live-local-allocation-and-discharge*
*Completed: 2026-09-28*
