---
phase: 23-live-local-allocation-and-discharge
plan: 03
subsystem: compiler
tags: [ownership, foreign-operations, c17-emission, refusal]
requires:
  - phase: 23-01
    provides: checked local FileByteOwner operation facts and the sole C-emitter baseline
provides:
  - Source checking refuses discarded successful FileByteOwner acquisition and unsupported owner use.
  - The sole C emitter preflights each selected local-owner operation contract before serialization.
affects: [23-02, 23-04, 23-05, 23-06, 23-07, 24]
actuals:
  tokens: 5004
  tasks: 2
  commits: 4
  plan_head_before: 581754eea346360ace666787398a95d1a5bafd7d
tech-stack:
  added: []
  patterns:
    - Successful owning acquisition must remain bound so its cleanup obligation is represented.
    - The sole emitter independently checks local-owner operation contracts before C serialization.
key-files:
  created: [testdata/phase23/discard_owner.lang]
  modified: [internal/compiler/check/check.go, internal/compiler/check/check_test.go, internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_program_test.go]
key-decisions:
  - Reject discarded FileByteOwner acquisition at source admission with an operation and source-place diagnostic.
  - Gate malformed local-owner candidates before C serialization, including exact per-operation contracts and owner facts.
requirements-completed: [FFI-03, RES-08, RES-09]
coverage:
  - id: D1
    description: Discarded acquisition and unsupported local-owner source forms are refused during checking.
    requirement: RES-08
    verification:
      - kind: unit
        ref: internal/compiler/check/check_test.go#TestPhase23DiscardRequiresOwnerBinding
        status: pass
      - kind: unit
        ref: internal/compiler/check/check_test.go#TestPhase23SourceRefusal
        status: pass
    human_judgment: false
  - id: D2
    description: Invalid owner candidates and operation contracts are rejected before C serialization while the valid local-owner path remains emit-able.
    requirement: FFI-03
    verification:
      - kind: unit
        ref: internal/compiler/cgen/cgen_program_test.go#TestPhase23DiscardRefusedBeforeCSerialization
        status: pass
      - kind: unit
        ref: internal/compiler/cgen/cgen_program_test.go#TestPhase23SourceRefusalBeforeCSerialization
        status: pass
      - kind: unit
        ref: internal/compiler/cgen/cgen_program_test.go#TestPhase23OperationContractRefusalBeforeCSerialization
        status: pass
    human_judgment: false
  - id: D3
    description: Existing Phase 22 U64 emission and the public Phase 23 file-byte program remain accepted.
    verification:
      - kind: integration
        ref: go test ./internal/compiler/cgen -run '^TestPhase22ApplicationEmitter' -count=1
        status: pass
      - kind: integration
        ref: go test ./cmd/lang -run '^TestPhase23PublicFileByte' -count=1
        status: pass
    human_judgment: false
duration: 13min
completed: 2026-09-27
status: complete
---

# Phase 23 Plan 03: Live Owner Refusal Boundaries Summary

**The checker preserves successful local allocation obligations at source admission, and the C emitter rejects altered owner contracts before serialization.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-09-27T23:09:01-04:00 (first retained task commit)
- **Completed:** 2026-09-27T23:21:57-04:00
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added source diagnostics and fixtures for discarded successful owner acquisition, copying, moved-from use, escape, unsupported call transfer, and unsupported exits.
- Added a pre-serialization emitter gate for malformed owner candidates and incorrect acquire, use, and release contracts; valid local ownership still emits.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase23(SourceRefusal|Discard)' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase23(SourceRefusal|Discard|OperationContract)' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase22ApplicationEmitter' -count=1` — passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase23PublicFileByte' -count=1` — passed.
- `git diff --check` — passed.

## Task Commits

1. **Task 1: Refuse discarded and unsupported local owner use**
   - `a7302fa` — `test(23-03): cover discarded and unsupported local owner use`
   - `881b1d8` — `feat(23-03): reject discarded local owner acquisition`
2. **Task 2: Preflight local owner operation contracts**
   - `8efedd2` — `test(23-03): pin local owner emitter refusals`
   - `9adae9d` — `feat(23-03): preflight local owner contracts before C emission`

## Files Created/Modified

- `testdata/phase23/discard_owner.lang` — invalid discarded-acquisition source fixture.
- `internal/compiler/check/check.go` — early source refusal for discarded owner acquisition.
- `internal/compiler/check/check_test.go` — source-level refusal tests.
- `internal/compiler/cgen/cgen_program.go` — pre-serialization local-owner contract gate.
- `internal/compiler/cgen/cgen_program_test.go` — emitter refusal and valid-path tests.

## Decisions Made

- Discarded successful acquisition is refused before lowering so its cleanup obligation cannot disappear.
- The sole C emitter independently validates operation and owner facts instead of trusting candidate shape alone.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

The GSD RED-evidence checker accepts TAP/Node-style failure records, while the Go tests emit package-format output. For each RED gate, the recorded output retained the observed Go assertion failure and included a labeled normalized TAP projection of that same target-test failure; the checker returned `RED_EVIDENCE_OK` for both tasks.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The source checker and sole-emitter boundaries for local allocation and discharge are in place. Other Phase 23 Wave 2 plans retain their assigned ownership; this plan introduces no blocker for them.

## Self-Check: PASSED

- All five planned files are present.
- All four task commits are present in Git history.
- The focused checker, emitter, U64 regression, and public Phase 23 verification commands passed.
