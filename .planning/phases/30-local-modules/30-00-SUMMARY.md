---
phase: 30-local-modules
plan: 00
subsystem: testing
tags: [go, os.Root, byte-identity-golden, gate-verdict, modload, cgen-naming]

requires:
  - phase: 29-byte-views-and-the-adler-32-library
    provides: verdict shape, adler32 library fixtures
provides:
  - no-import byte-identity golden (testdata/phase30/prechange-corpus.json)
  - closed module resolver primitive (internal/compiler/modload/fs.go)
  - pure moduleCName injective C naming (unused until 30-07)
  - 30-00-GATE-VERDICT.md with verdict PROCEED
affects: [30-01, 30-02, 30-03, 30-04, 30-05, 30-06, 30-07, 30-08, 30-09]

actuals:
  tokens: 36000
  tasks: 3
  commits: 4
plan_head_before: 4ea05fd7e2f8dec830fb6392331e44c86bda0e8a
plan_head_after: e2cf70e301735d083a88145553ad6f5501fbec42
commits: 4

tech-stack:
  added: []
  patterns:
    - "os.Root with per-component Lstat, exact-case listing, regular-before-open, SameFile"
    - "armed fault-injection seams (unexported bool vars) with a control per gate flag"

key-files:
  created:
    - testdata/phase30/prechange-corpus.json
    - internal/compiler/session/session_phase30_prechange_test.go
    - internal/compiler/session/session_phase30_gate_test.go
    - internal/compiler/modload/fs.go
    - internal/compiler/modload/modload_phase30_gate_test.go
    - internal/compiler/modload/modload_phase30_gate_unix_test.go
    - internal/compiler/diagnostic/diagnostic_phase30_gate_test.go
    - internal/compiler/cgen/cgen_phase30_gate_test.go
    - internal/compiler/check/check_phase30_gate_test.go
    - .planning/phases/30-local-modules/30-00-GATE-VERDICT.md
  modified:
    - internal/compiler/cgen/cgen.go

key-decisions:
  - "Gate verdict is PROCEED: flags (a)-(f) all passed with armed controls and no D-16 split condition holds."
  - "The golden pins the recovered debug-map panic text for five sources that already panic in debugmap.Build at PLAN_BASE."
  - "Case mismatch is classified by listing the parent on Lstat ErrNotExist as well as on Lstat success, so it is the same on every OS."

patterns-established:
  - "Golden first commit: test and golden only, generated at PLAN_BASE, revision recorded"
  - "Seam-based controls: each gate flag has a variant that must be observed to fail"

requirements-completed: [MOD-03, MOD-04, MOD-05]

duration: 40min
completed: 2026-10-12
status: complete
---

# Phase 30 Plan 00: Wave-0 gate Summary

**No-import byte-identity golden committed first, then a closed os.Root module resolver primitive and six armed gate flags, ending in verdict: PROCEED.**

## Performance

- **Tasks:** 3 (golden and resolver tracer, flags c/d, flags e/f and verdict)
- **Files modified:** 11 (10 created, 1 production edit besides fs.go: cgen.go)

## Accomplishments
- `TestPhase30PreChangeGolden` pins six hashes for 159 committed `.schway` files plus one generated over-work row (160 entries) at revision 4ea05fd; `TestPhase30PreChangeGoldenIsNotInert` flips one hash and sees the mismatch.
- `modload` refuses escaping and in-root symlinks, FIFOs (non-blocking, within 2 s), and case variants by name; controls for skipped Lstat, open-before-Lstat and case-fold accept each observably fail the matching test.
- Unit omitempty keeps empty-unit diagnostic IDs identical (diagnostic:d34396fcc8ea6530a1f04dc3); consumer inventory of 19 non-test files pinned.
- `moduleCName` is injective over 4,672 pairs, free of `__`, namespace-confined and order independent; the plain-join seam collides.
- Merged-program gap confirmed: 1,025 functions pass check.Program with no limit; plan 30-04 owns `module.program_limit`.

## Task Commits

1. **Task 1 Commit A: golden** - `7ef4c45` (test)
2. **Task 1 Commit B: resolver primitive** - `219c8b4` (feat)
3. **Task 2: flags c and d** - `8c0bd34` (test)
4. **Task 3: flags e and f, verdict** - `e2cf70e` (docs)

## Decisions Made
- Verdict PROCEED. Over-work generator uses 400 functions of 10 statements (plus main); count only, no timing pinned (D-28).
- Walked files with check diagnostics record `skipped: check diagnostics` for core, EmitApplication and control C fields.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] DebugMapCommandFile panics on five committed sources**
- **Found during:** Task 1 (golden generation)
- **Issue:** debugmap.Build dereferences nil for testdata/phase1/comments.schway, testdata/phase1/toggle.schway, testdata/phase17/return_type_tracer.schway, testdata/phase6/derivation_match_defect.schway, testdata/phase6/heldout_match_defect.schway at PLAN_BASE, aborting golden generation.
- **Fix:** the test recovers and records `panic: <text>` as the debug_map field, so the existing behavior is pinned rather than hidden. No production edit.
- **Files modified:** internal/compiler/session/session_phase30_prechange_test.go
- **Committed in:** 7ef4c45

---

**Total deviations:** 1 auto-fixed (1 blocking). **Impact:** none on scope; pre-existing defect recorded in the verdict.

## Issues Encountered
The pre-existing debug-map panics above are a latent defect outside this phase; they are not fixed here.

## TDD Gate Compliance
Plan type is `execute` with a tracer, not `type: tdd`. Commit A (test + golden) precedes all production edits as required by D-15.

## Known Stubs
None. `moduleCName` is intentionally unused until plan 30-07 (documented beside the function).

## Threat Flags
None beyond the plan's threat model: `os.OpenRoot` appears only in internal/compiler/modload/fs.go and nothing wires it into a CLI route.

## User Setup Required
None.

## Next Phase Readiness
Verdict PROCEED; plans 30-01 to 30-09 may proceed with `TestPhase30PreChangeGolden` in their verify. The golden must not be regenerated.

## Self-Check: PASSED
