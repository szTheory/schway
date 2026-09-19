---
phase: 16-branch-match-emitter-port
plan: 05
subsystem: native-emission-evidence
tags: [c17, clang, restrict, sanitizer, lto, m004]
requires:
  - phase: 16-branch-match-emitter-port
    provides: D-16-06/D-16-07 restrict probe lanes and structural refusal fence
provides:
  - Human-selected, evidence-linked M004 cut for every by-pointer emitter family
  - Recorded macOS lane, Linux-unavailable, and refusal-fence verdicts
affects: [16-06, 16-07, 16-08, 16-09, M004-debt]
actuals:
  tokens: 1421
  tasks: 1
  commits: 1
commits: 1
plan_head_before: 3b2921e2bf07d8fd1877591c2927cc09dc6a39fa
tech-stack:
  added: []
  patterns: [blocking-human evidence disposition, cross-host admission threshold, explicit M004 cut]
key-files:
  created:
    - .planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md
  modified: []
key-decisions:
  - "Selected cut-m004: macOS-only evidence is insufficient, so emitLinearBorrowedByPointer and every other by-pointer family remain unavailable in M003."
  - "Retain the source/refusal fence and macOS results as evidence, but do not treat the unavailable Linux lane as a passing proxy."
requirements-completed: []
coverage:
  - id: D1
    description: "The by-pointer admission boundary records exact source, macOS lane, Linux availability, and refusal-fence evidence before selecting a disposition."
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/cgen -run 'TestRestrictReadonlyProbe(Darwin|Linux)Lanes|TestRestrictReadonlyProbeSourceShape|TestRestrictAdmissionRejectsExtensions|TestRestrictAdmissionFenceIsNotInert' -count=1 -v"
        status: pass
    human_judgment: true
    rationale: "The blocking-human checkpoint selected cut-m004 because the required Linux evidence is unavailable; automated evidence cannot substitute for that disposition."
duration: 3 min
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 05: Bounded By-Pointer Disposition Summary

**The resolved `cut-m004` checkpoint retains every by-pointer emitter family as owned M004 debt: macOS evidence and the source/refusal fence passed, but Linux is unavailable and macOS-only evidence cannot authorize admission.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-09-19T21:39:39Z
- **Completed:** 2026-09-19T21:42:39Z
- **Tasks:** 1/1
- **Files modified:** 1

## Accomplishments

- Recorded the explicit human selection `cut-m004`, as required by the blocking decision checkpoint.
- Preserved the exact probe identity: SHA-256 `1e06f052845b75a738e481b87d3076d46fbafef4ca3ac137830a36e2ec35496e`, compiled on Darwin with `/usr/bin/clang`.
- Recorded macOS `O0`, `O3`, `O3-LTO`, and `ASan-UBSan` as `PASS`, every named D-16-07 extension refusal as passing, and Linux `all` as `UNAVAILABLE` rather than a proxy pass.

## Evidence and Disposition

| Evidence | Result | Disposition relevance |
| --- | --- | --- |
| Source-shape digest | PASS (`1e06f…35496e`) | Establishes the immutable one-TU read/copy-only probe bytes. |
| macOS `/usr/bin/clang` O0 / O3 / O3-LTO / ASan+UBSan | PASS / PASS / PASS / PASS | Necessary evidence, but not sufficient by itself. |
| Linux required-host lane | UNAVAILABLE | Blocks admission; no Linux result may be inferred from Darwin. |
| D-16-07 refusal matrix | PASS | Rejects pointee mutation, second pointer, pointer escape/forwarding, callback, foreign call, volatile, atomic, and separate compilation. |
| Fence mutation control | PASS | Confirms the refusal fence is live rather than inert. |

The human checkpoint resolution is `cut-m004`. `emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`, and any other by-pointer family are not admitted in M003. The M004 debt work must provide the complete discharge-pair design and the missing cross-host evidence before revisiting a by-pointer lowering family. This does not claim that LTO is non-inert; that remains an independent evidence obligation under D-16-08.

## Task Commits

1. **Task 1: Decide the bounded by-pointer disposition** — pending metadata commit (docs)

## Files Created/Modified

- `.planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md` — evidence-compatible `cut-m004` disposition and lane/refusal digest.

## Decisions Made

- Selected `cut-m004` because a required Linux lane is unavailable. Successful macOS lanes must never authorize by-pointer admission alone.
- Kept all by-pointer families as explicit M004 debt; no production emitter or alias-policy change is authorized by this plan.

## Deviations from Plan

None - the resolved blocking-human checkpoint selected the plan's evidence-compatible recommended option.

## Issues Encountered

- The Linux lane is unavailable on this Darwin executor. This is a measured `UNAVAILABLE` result, not an unrun verification or a test failure, and it is the reason the admission threshold is not met.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Later Phase 16 plans must carry the formal M004 debt disposition forward and must not route `emitLinearBorrowedByPointer` or another by-pointer family into production. A future admission requires Linux and macOS exact-shape lanes plus the complete structural fence; LTO non-inertness remains separate evidence debt.

## Self-Check: PASSED

- Found `.planning/phases/16-branch-match-emitter-port/16-05-SUMMARY.md`.
- Focused source-shape, refusal-matrix, and mutation-control verification passed; the explicit Linux test reported the expected `UNAVAILABLE` result.
