---
phase: 36-scalar-helper-calls
plan: "01"
subsystem: testing
tags: [go, c17, compatibility, generated-c, phase36]
requires:
  - phase: 35-bitwise-operators-literal-shifts-and-hex-literals
    provides: Existing scalar emitters, generated-C routes, and frozen-corpus patterns
provides:
  - Literal no-call generated-C outcomes for all 163 tracked Schway sources at the pre-feature revision
  - Exact imported-module source paths and digests for every captured module composition
  - Live drift, omission, false-exemption, and ledger path-containment controls
affects: [Phase 36 plans 02-08, CALL-01, CALL-02]
actuals:
  tokens: 1078872
  tasks: 2
  commits: 3
plan_head_before: bf986b6515417acefa662e7c3541b8ef85bad523
plan_head_after: b7574b195299c0cd111f95ba8a51dba1008bd415
commits: 3
tech-stack:
  added: []
  patterns:
    - Enumerate tracked source inventory and freeze exact outputs per emitter route
    - Bind imported module identities and digests to exact source rows
key-files:
  created:
    - internal/compiler/session/session_phase36_prechange_test.go
    - testdata/phase36/prechange-corpus.json
    - testdata/phase36/new-fixture-ledger.json
    - .planning/phases/36-scalar-helper-calls/36-01-SUMMARY.md
  modified:
    - .planning/STATE.md
    - .planning/ROADMAP.md
    - .planning/WINDOWS.md
key-decisions: []
requirements-addressed: [CALL-02]
requirements-completed: []
coverage:
  - id: D1
    description: Frozen source bytes, checked identities, imported-module digests, and explicit outcomes for each of 163 tracked pre-change sources across three emitter routes.
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline
        status: pass
    human_judgment: false
  - id: D2
    description: The comparator reaches C-byte drift, omitted-row, false route-exemption, old-source ledger, and unsafe ledger-path controls.
    verification:
      - kind: integration
        ref: internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaselineControls
        status: pass
    human_judgment: false
patterns-established:
  - Every tracked pre-change source has a row for emit, emit_native, and emit_application, including refusal and non-emission outcomes.
  - New fixture exemptions require a sorted, source-hash-bound ledger row and cannot hide old corpus rows.
duration: 9min
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 01: Freeze No-Call Generated C Baseline Summary

**Literal C snapshots now cover all 163 tracked Schway sources across three emitter routes before scalar-call compiler changes.**

## Performance

- **Duration:** 9 min
- **Started:** 2026-10-15T01:22:10Z
- **Completed:** 2026-10-15T01:31:07Z
- **Tasks:** 2/2
- **Files changed by the tasks:** 3

## Accomplishments

- Captured every `.schway` source from revision `bf986b6515417acefa662e7c3541b8ef85bad523`, with exact source bytes and hashes, checked-core identities, and explicit emitted, refused, or non-emitting outcomes for `Emit`, `EmitNative`, and `EmitApplication`.
- Bound all six module rows from the three import-bearing roots to their exact tracked source paths and SHA-256 digests. The intentional new-fixture ledger starts empty.
- Added controls that reach byte drift, an omitted legacy row, a falsely exempted emitting route, an old source placed in the new-fixture ledger, and a traversal path rejected before file access.

## Verification

Fresh focused commands passed with nonzero named roots:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s` — `TestPhase36NoCallGeneratedCByteBaseline` RUN/PASS.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s` — both named roots RUN/PASS after the final test changes.

The baseline was recorded once from the pinned pre-feature revision. The capture-only helper was removed before commit; normal tests read the frozen fixture and do not regenerate it. No full suite, whole-package run, or broad native matrix was run.

## Task Commits

1. **Task 1: Commit the literal source-to-C corpus and current inventory** — `9b8285de` (`test(36-01): freeze no-call generated C baseline`)
2. **Task 2: Arm drift, omission, and false-exemption controls** — `7588edf6` (`test(36-01): arm generated C baseline controls`)
3. **Rule 2 security fix:** Contain fixture-ledger paths before reading source files — `b7574b19` (`fix(36-01): contain fixture ledger paths`)

**Measured task commits:** 3, from `bf986b6515417acefa662e7c3541b8ef85bad523` through `b7574b195299c0cd111f95ba8a51dba1008bd415`.

## Files Created and Updated

- `internal/compiler/session/session_phase36_prechange_test.go` — frozen comparator, tracked-inventory and module-digest checks, and reached negative controls.
- `testdata/phase36/prechange-corpus.json` — literal C and explicit route outcomes for 163 sources at the captured revision.
- `testdata/phase36/new-fixture-ledger.json` — empty, versioned ledger for intentional later fixture additions.
- `.planning/STATE.md` and `.planning/ROADMAP.md` — advance the plan pointer and record the completed plan count.
- `.planning/WINDOWS.md` — records the path-containment deviation as fixed entry 9.

## Decisions Made

Followed D-19 and D-25 as written. CALL-02 is addressed by this prerequisite but remains incomplete; later Phase 36 plans own the full requirement.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Security] Reject unsafe fixture-ledger paths before file access**
- **Found during:** Task 2 review
- **Issue:** A malformed new-fixture ledger path could cause the test to read outside the repository.
- **Fix:** Require normalized repository-relative paths and add a reached traversal rejection control.
- **Files modified:** `internal/compiler/session/session_phase36_prechange_test.go`
- **Verification:** Both required focused roots pass; the controls root asserts traversal rejection before file access.
- **Committed in:** `b7574b19`

**Total deviations:** 1 auto-fixed (Rule 2 - Security). The change closes a direct file-read boundary in the planned ledger validation.

## Issues Encountered

The temporary one-time corpus capture harness needed two local helper-name corrections before it ran. The capture then completed successfully, and the helper file was removed before the baseline commit.

## User Setup Required

None.

## Next Phase Readiness

Plan 36-02 can proceed with compiler changes against the committed no-call baseline. This plan proves the compatibility prerequisite only; CALL-01/CALL-02 and Phase 36 acceptance remain open.

## Self-Check: PASSED

- All three task artifacts and this summary exist.
- Task commits `9b8285de`, `7588edf6`, and `b7574b19` are ancestors of HEAD.
- Both named baseline roots pass with the required RUN/PASS output.
- The coverage classifier recognizes both deliverables as automated and error-free.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*
