---
phase: 23-live-local-allocation-and-discharge
plan: 06
subsystem: native ownership
tags: [native, ownership, allocation, cleanup, evidence]

requires:
  - phase: 23-01
    provides: Checked PathToken acquire, borrow, and release operations
  - phase: 23-04
    provides: Public native application and injected acquisition cleanup controls
provides:
  - Independent pointer-identity receipts for the public success and use-error paths
  - Reached safe controls for omitted, premature, duplicate, and wrong-resource release
  - Schema-2 return evidence for the specialized local-owner native lowering
affects: [23-07, Phase 24 ownership transfer]

actuals:
  tokens: 7758
  tasks: 2
  commits: 3
  plan_head_before: 986d048853c88e536ba9441f9c4aba54ee52f5af

tech-stack:
  added: []
  patterns:
    - Test-only C hooks delegate valid allocation, use, and free operations to the real libc path
    - A private process receipt independently records pointer identity and cleanup order

key-files:
  created:
    - internal/compiler/native/testdata/phase23_observer.c
    - internal/compiler/native/phase23_observer_test.go
  modified:
    - examples/phase23/adapter.c
    - examples/phase23/adapter.h
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/native/native_app_test.go
    - cmd/lang/main_test.go
    - internal/compiler/core/core_convention_absence_test.go

key-decisions:
  - "Keep physical pointer lifetime receipts separate from compiler semantic events."
  - "Emit the checked function.returned semantic event for the specialized local-owner body so its schema-2 capture remains valid."
  - "Route only PathToken/FileByteOwner facts through the narrow local-owner emitter; unrelated foreign operations retain their existing refusals."

requirements-completed: [RES-04, RES-07, RES-09]

coverage:
  - id: D1
    description: "Public 0x41/0x42 success and 0x43 typed-use failure each prove malloc, later use of the same live pointer, release, and zero outstanding allocations before exit/reporting."
    requirement: RES-04
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/native ./cmd/lang -run '^TestPhase23(Observer|PublicFileByte|PublicUseError)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Omitted, premature, duplicate, and wrong-resource destruction mutations are reached, safely rejected, and remain distinct from plausible semantic compiler records."
    requirement: RES-09
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/native ./internal/compiler/cgen -run '^TestPhase23(ObserverMutation|PhysicalDestructorControl)' -count=1"
        status: pass
    human_judgment: false

duration: about 3h
completed: 2026-09-28
status: complete
---

# Phase 23 Plan 06: Independent native discharge evidence

**A separately compiled observer now proves that the public app uses and frees the same real pointer, including before reporting a typed use error.**

## Performance

- **Duration:** About 3 hours, including isolated-run recovery and integration checks.
- **Started:** 2026-09-28 (exact worktree start time was not recorded).
- **Completed:** 2026-09-28.
- **Tasks:** 2.
- **Files modified:** 9.

## Accomplishments

- Added a private C17 observer receipt for pointer identity and physical malloc/use/free order. It delegates valid operations to libc and rejects unsafe mutation paths before invalid libc use.
- Proved 0x41 and 0x42 return 65 and 66, and the public 0x43 path reports `UnsupportedByte` only after freeing the acquired pointer.
- Reached omitted, premature, duplicate, and wrong-resource destructor controls. The positive control passes; all mutations fail with a distinct observer receipt while schema-2 semantic evidence remains plausible.
- Restored the ordinary `function.returned` semantic event in the specialized local-owner lowering so evidence decoding accepts the returned execution record.

## Task Commits

1. **Task 1 RED:** `a58b8e4` (`test(23-06): observe physical file-byte lifecycles`).
2. **Task 1 GREEN:** `56cdcb7` (`feat(23-06): observe physical allocation and use cleanup`).
3. **Task 2 and integration fixes:** `75c35ba` (`test(23-06): verify reached native discharge controls`).

## Files Created/Modified

- `internal/compiler/native/testdata/phase23_observer.c` — Separately compiled pointer identity and lifetime observer.
- `internal/compiler/native/phase23_observer_test.go` — Public success/error receipts and four reached destructor mutations.
- `examples/phase23/adapter.c` and `adapter.h` — Real byte load hook and typed unsupported-byte status.
- `internal/compiler/cgen/cgen_program.go` — Post-release error diagnostic and schema-2 return event; local-owner shape remains narrowly recognized.
- `internal/compiler/cgen/cgen_program_test.go` — Confirms semantic return evidence remains separate from physical discharge claims.
- `internal/compiler/native/native_app_test.go` and `cmd/lang/main_test.go` — Public cleanup ordering, bounded output, and use-error checks.
- `internal/compiler/core/core_convention_absence_test.go` — Keeps the exact core field inventory current with the existing `Foreign` field.

## Test Evidence

- The two plan verification commands pass on the current macOS host.
- Full package tests pass: `go test ./internal/compiler/native ./internal/compiler/cgen ./cmd/lang`.
- A repository-wide `go test ./...` run is not yet green. The remaining failures are project-integrity updates that belong at the Phase 23 closeout: stale `.planning/LANGUAGE-MATURITY.md` corpus/guard counts, stale Phase 16 emitter-call registry locations, the Phase 23 validation commands missing from the pinned groundedness frontier, and an uncited optional-Clang skip. The first full run also identified the stale `LinearOperation` expected-field list; that was corrected here. These closeout items must be resolved before enabling the repo-wide gate in recurring CI.
- This plan has no Linux run. Cross-host evidence remains with Plan 23-07 and its CI lanes.

## Deviations from Plan

### Auto-fixed integration issues

1. **Specialized return evidence was incomplete.** The independent mutation harness exposed that the Phase 23 local-owner lowering serialized a completed capture with an empty event list. It now records the real semantic `function.returned` event; the physical observer remains the separate source of allocation evidence.
2. **Local-owner detection intercepted unrelated foreign paths.** The full compiler-package run showed historical foreign-call refusal tests receiving a Phase 23 shape error. Detection now relies on the narrow PathToken/FileByteOwner facts, preserving unrelated refusal behavior.
3. **The test subprocess helper exceeded the bounded-output policy.** Replaced three `CombinedOutput` calls in the acquisition-fault harness with separate bounded stdout/stderr writers.
4. **The core field inventory had not recorded the already-added `Foreign` field.** Updated its explicit expected set so the full suite checks the current structure.

These fixes were needed to keep Phase 23's public evidence valid and preserve existing compiler contracts. On 2026-09-28, GSD cleanup reported `scope_out_of_declared` for `internal/compiler/core/core_convention_absence_test.go`; that exact-field expectation update was needed because the repository-wide run caught the already-added `Foreign` field.

## Issues Encountered

- The isolated executor repeated a stale-context patch and stopped producing test results. Its first two task commits were preserved; execution resumed in the same worktree, and the focused and package tests passed after the missing event and detector boundary were corrected.
- The shared Git index required the normal GSD commit escalation. No working changes were lost.
- Phase 23 focused native checks are green locally; Linux and recurring CI evidence are still pending Plan 23-07.

## User Setup Required

None.

## Next Plan Readiness

Plan 23-07 can add the focused recurring macOS/Linux evidence job and update the clean-checkout instructions. Before phase completion, close the repository-integrity failures listed above, refresh the living maturity/roadmap documents against source witnesses, and rerun `go test ./...`. No conversational UAT is needed for this plan's objective checks.

---
*Phase: 23-live-local-allocation-and-discharge*
*Completed: 2026-09-28*

## Self-Check: PASSED

- All listed created and modified files exist.
- All three task commits are present in the isolated worktree.
- Both planned Phase 23 verification commands and the native/cgen/CLI package suites pass.
- The repository-wide suite findings are documented above and remain a phase closeout item; no Linux result is claimed.
