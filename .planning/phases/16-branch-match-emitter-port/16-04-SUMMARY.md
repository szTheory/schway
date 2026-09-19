---
phase: 16-branch-match-emitter-port
plan: 04
subsystem: native-emission-evidence
tags: [go, c17, clang, restrict, sanitizer, lto]
requires:
  - phase: 16-branch-match-emitter-port
    provides: D-16-06/D-16-07 restrict probe decision boundary
provides:
  - Immutable one-TU read/copy-only C17 restrict probe with source digest evidence
  - Named host lanes that distinguish PASS from UNAVAILABLE
  - Test-local D-16-07 admission fence and mutation control
affects: [16-05, cgen, native-execution, M004-cut]
actuals:
  tokens: 2592
  tasks: 2
  commits: 4
commits: 4
plan_head_before: bd9264c0c234d1290af9c2d6b6231a414cfe1605
tech-stack:
  added: []
  patterns: [immutable C probe bytes, host-lane evidence records, test-local structural admission fence]
key-files:
  created:
    - testdata/phase16/restrict_readonly_probe.c
    - internal/compiler/cgen/cgen_restrict_probe_test.go
  modified: []
key-decisions:
  - "Keep restrict admission evidence test-local and make no production emitter routing change before Plan 05's human disposition."
  - "Treat Linux as UNAVAILABLE on a non-Linux host, never as a passing proxy for macOS evidence."
requirements-completed: []
coverage:
  - id: D1
    description: Exact one-pointer C17 probe records separate source-shape, O0, O3, O3-LTO, sanitizer, and host availability results.
    verification:
      - kind: integration
        ref: internal/compiler/cgen/cgen_restrict_probe_test.go#TestRestrictReadonlyProbeDarwinLanes
        status: pass
      - kind: integration
        ref: internal/compiler/cgen/cgen_restrict_probe_test.go#TestRestrictReadonlyProbeLinuxLanes
        status: unknown
    human_judgment: true
    rationale: Linux evidence is explicitly unavailable on the macOS executor and must be considered by the later disposition.
  - id: D2
    description: Candidate structural fence accepts only the exact shape and rejects every D-16-07 extension with a live bypass control.
    verification:
      - kind: unit
        ref: internal/compiler/cgen/cgen_restrict_probe_test.go#TestRestrictAdmissionRejectsExtensions
        status: pass
      - kind: unit
        ref: internal/compiler/cgen/cgen_restrict_probe_test.go#TestRestrictAdmissionFenceIsNotInert
        status: pass
    human_judgment: false
duration: 2 min
completed: 2026-09-19
status: complete
---

# Phase 16 Plan 04: Restrict Readonly Probe Summary

**A checked-in one-TU C17 `restrict` microprogram now records macOS lane evidence and an explicit Linux-unavailable result, alongside a non-inert test-local refusal fence.**

## Performance

- **Duration:** 2 min
- **Started:** 2026-09-19T21:24:44Z
- **Completed:** 2026-09-19T21:26:31Z
- **Tasks:** 2/2
- **Files modified:** 2

## Accomplishments

- Added the immutable D-16-06 source shape with exactly one `T *restrict p`, one TU, read/copy-only callee behavior, and caller-local access around the call.
- Added no-shell Clang host lanes that compile the same SHA-256-bound bytes for O0, O3, O3-LTO, and ASan/UBSan; macOS lanes passed and Linux emitted `UNAVAILABLE` rather than a proxy success.
- Added a smallest-possible test-local candidate fence that names and rejects mutation, second pointer, escape, forwarding, callback, foreign call, volatile, atomic, and separate-compilation extensions; its bypass control proves the checks are live.

## Task Commits

1. **Task 1: Check in and execute the exact one-pointer C17 probe** — `03e5afe` (test), `70daf29` (feat)
2. **Task 2: Prove the source-shape fence refuses every prohibited extension** — `e4ab2e5` (test), `3562bca` (feat)

## Files Created/Modified

- `testdata/phase16/restrict_readonly_probe.c` — normative, hand-written D-16-06 C17 source.
- `internal/compiler/cgen/cgen_restrict_probe_test.go` — source digest validation, host lane runner, refusal matrix, and mutation control.

## Decisions Made

- Made no `emitLinearBorrowedByPointer` routing or alias-policy change: this plan supplies evidence only and preserves Plan 05's human disposition.
- Retained Linux absence as an unresolved result; successful macOS lanes do not satisfy the dual-host prerequisite.

## TDD Gate Compliance

- Task 1 RED: the named source-shape and host-lane tests failed because the tracked C fixture did not yet exist; the fixture made them green.
- Task 2 RED: the named refusal and mutation-control tests failed against the deliberate one-pointer-only candidate; the full named fence made them green.
- Focused verification and `go test ./internal/compiler/cgen -count=1` passed after the green commits.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- Linux evidence cannot be collected from the macOS executor. The named Linux test reports `UNAVAILABLE` with the same source digest and remains a blocking input to the next plan's disposition.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 05 can inspect machine-readable lane output and the refusal matrix, but must not admit the family while Linux remains unavailable. NAT-09 is intentionally not marked complete here because its formal M004/admission disposition belongs to a later plan.

## Self-Check: PASSED

- Found `testdata/phase16/restrict_readonly_probe.c` and `internal/compiler/cgen/cgen_restrict_probe_test.go`.
- Found task commits `03e5afe`, `70daf29`, `e4ab2e5`, and `3562bca` in git history.
