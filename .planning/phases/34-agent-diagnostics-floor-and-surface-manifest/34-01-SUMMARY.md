---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "01"
subsystem: testing
tags: [pre-change-baseline, cgen, parser, diagnostics, protocol]
requires: []
provides:
  - "A revision-pinned literal C and admission corpus for all 162 tracked Schway inputs and both emitters."
  - "Twenty provenance-labeled parser drafts with literal pre-change diagnostics, causes, spans, IDs, counts, and recovered siblings."
  - "Bool-name admission witnesses and source-less diagnostic, repair, unit, and result identity snapshots."
affects: [34-02, 34-03, 34-04, 34-05, 34-06, 34-07, 34-08]
actuals:
  tokens: 321787
  tasks: 3
  commits: 4
plan_head_before: 0a3cdc1f8903d6cff82c97c2f986cbdd11fa0168
plan_head_after: 74fcfdd91449ea353f3857b124e1951ed10316cd
tech-stack:
  added: []
  patterns: [literal-output-golden, reconstructed-fixture-provenance, source-less-identity-snapshot]
key-files:
  created:
    - internal/compiler/session/session_phase34_prechange_test.go
    - testdata/phase34/prechange-corpus.json
    - internal/compiler/syntax/parser_phase34_baseline_test.go
    - testdata/phase34/drafts.json
    - testdata/phase34/draft-baseline.json
    - internal/compiler/session/session_phase34_compat_test.go
    - testdata/phase34/identity-baseline.json
  modified: []
key-decisions:
  - "Freeze successful emitter output as complete literal C; compare explicit emission outcome classes and prior successful admissions."
  - "Label all twenty draft sources reconstructed-2026-10-13; CRC32 and word-counter examples are representative reconstructions, not recovered scratch files."
  - "Keep the historical unbound true tail as name.unknown evidence while testing legal true/false names live through the existing routes."
patterns-established:
  - "Baseline writers run only for initial capture and are removed before commit; ordinary tests have no regeneration path."
  - "Failed-diagnostic text remains historical evidence while emitted C and explicit emitter outcomes stay guarded."
requirements-completed: [DX-20, DX-21, DX-22, DX-23]
coverage:
  - id: D1
    description: "The 162-row source-to-C baseline records both emitter outcomes and literal successful C, with reached mutation controls."
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCGoldenIsNotInert"
        status: pass
    human_judgment: false
  - id: D2
    description: "Twenty reconstructed parser drafts retain their literal pre-change diagnostics and recovery evidence."
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/parser_phase34_baseline_test.go#TestPhase34DraftBaseline"
        status: pass
    human_judgment: false
  - id: D3
    description: "Accepted true/false names and imported true are witnessed; the unbound true before-state is frozen at both routes."
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34BoolNameBaseline"
        status: pass
    human_judgment: false
  - id: D4
    description: "Legacy diagnostic bytes, repair identity, unit attribution, and Result.Finalize IDs are pinned."
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34IdentityBaseline"
        status: pass
    human_judgment: false
duration: 13min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 01: Pre-change Diagnostics and Emission Baselines Summary

**Literal baselines now preserve 162 source-to-C outcomes, twenty reconstructed parser drafts, Bool-name admissions, and legacy identity bytes.**

## Performance

- **Duration:** 12 min 54 sec
- **Started:** 2026-10-13T21:20:53Z
- **Completed:** 2026-10-13T21:33:47Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Captured every one of the 162 tracked Schway source paths at revision `0a3cdc1f8903d6cff82c97c2f986cbdd11fa0168`. `EmitApplication` has 17 emitted and 145 non-emitted outcomes; `Phase16ControlNativeC` has 70 emitted and 92 non-emitted outcomes. Successful outputs are stored as literal C strings, and refusal/non-emission details remain explicit per emitter.
- Added twenty `reconstructed-2026-10-13` drafts with 77 literal pre-change parser diagnostics. The baseline preserves code, ID, message, byte spans, cause spans, exact counts, source bytes, and recovered sibling names. The two independently authored top-level errors retain two disjoint skipped-region spans; the sibling control retains its real recovery span. CRC32 and word-counter sources are representative reconstructions, not recovered originals.
- Recorded nine true/false name cases and eight source-less identity cases, with six frozen `Result.Finalize` IDs. Seven single-file legal-name cases pass both existing routes; the imported `true` function passes module-aware `CheckCommandFile`, while source-only `Check` retains `module.imports_not_loaded`. The unbound linear true tail remains frozen as `name.unknown` at both routes.
- The comparator now checks non-emitting outcome classes while ignoring failed-diagnostic wording as C identity. Its C-byte mutation and non-emitting-to-emitted status mutation both fail with the expected row/emitter attribution and pass again after restoration.

## Task Commits

1. **Task 1: Commit exhaustive literal pre-change C corpus** — `a788aa83` (`test(34-01): freeze literal pre-change C corpus`)
2. **Task 2: Measure reconstructed drafts and isolated causes** — `14b44fa3` (`test(34-01): freeze reconstructed draft diagnostics`)
3. **Task 3: Freeze legal Bool names and identity/repair compatibility** — `58b69c1b` (`test(34-01): freeze Bool names and identity bytes`)
4. **Rule 1 comparator correction** — `74fcfdd9` (`fix(34-01): compare explicit non-emission outcomes`)

## Files Created/Modified

- `internal/compiler/session/session_phase34_prechange_test.go` — exhaustive source inventory, full-C comparator, admission/emitter outcomes, and reached/restored mutation controls.
- `testdata/phase34/prechange-corpus.json` — literal source-to-C baseline with revision `0a3cdc1f8903d6cff82c97c2f986cbdd11fa0168`.
- `internal/compiler/syntax/parser_phase34_baseline_test.go` — immutable draft-baseline assertions without comparing future parser output to superseded before-state diagnostics.
- `testdata/phase34/drafts.json` — twenty authored draft sources, explicit provenance, cause relationships, and sibling expectations.
- `testdata/phase34/draft-baseline.json` — exact parser diagnostic before-state at revision `a788aa8371716b35aa6cb25d609e8b6476c38f22`.
- `internal/compiler/session/session_phase34_compat_test.go` — live Bool-name route checks and literal legacy identity/result snapshots.
- `testdata/phase34/identity-baseline.json` — route observations and source-less identity evidence at revision `14b44fa3102da99a916f12551a99363ecbb44b75`.

## Decisions Made

- Kept complete emitted C in the golden rather than digest-only records. The permanent comparator enforces emitted/refused/not-emitted outcome classes and literal C, and requires previously successful admissions to stay successful.
- Kept failure diagnostics as historical facts in the corpus but excluded their message/code presentation from C comparisons, so parser recovery work can change diagnostics without masquerading as a C byte regression.
- Kept reconstructed provenance explicit for every draft. No original CRC32 or word-counter scratch source is claimed as recovered.
- Recorded both source-only and module-aware behavior for the imported-name case. `Check` correctly reports that imports are not loaded by its single-file route; `CheckCommandFile` resolves the temp-directory module and accepts the imported `true` function.

## Verification Evidence

Final focused checks used only `GOCACHE=/tmp/ai-lang-verification-gocache` and all five named selectors passed:

- `go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1 -v` — both passed; package time 0.689 sec. The comparator mutation controls were reached and restored.
- `go test ./internal/compiler/syntax -run '^TestPhase34DraftBaseline$' -count=1 -v` — passed; package time 0.008 sec.
- `go test ./internal/compiler/session -run '^(TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline)$' -count=1 -v` — both passed; package time 0.011 sec.

Source inspection established the existing `CheckFile` module-loading path, both public C emitter boundaries, parser recovery behavior, legacy diagnostic schemas, and `Result.Finalize` identity inputs. The baseline captures were newly executed at their recorded revisions. No project-wide test suite, native matrix, hosted dispatch, or receipt binding was run.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Compare non-emitting emitter outcomes**

- **Found during:** Task 3 cross-task review
- **Issue:** The first Task 1 comparator stored `not_emitted` rows but skipped comparing their outcome class, so a transition from non-emitting to emitting could pass unnoticed.
- **Fix:** Compare explicit outcome classes for every emitter while leaving diagnostic reason text outside C identity; add a reached status mutation control and verify restoration.
- **Files modified:** `internal/compiler/session/session_phase34_prechange_test.go`
- **Verification:** The focused Task 1 selectors passed after the fix; both C-byte and non-emission status mutations were caught and restored.
- **Committed in:** `74fcfdd9`

---

**Total deviations:** 1 auto-fixed (Rule 1)
**Impact on plan:** The fix strengthens the planned comparator without production compiler changes or scope expansion.

## Issues Encountered

- The imported-name fixture intentionally produces `module.imports_not_loaded` from source-only `Check`; the module-aware `CheckCommandFile` route passes. Both exact route observations are recorded rather than conflated.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 34-01's three baseline tasks and the comparator correction are committed before any production compiler edits. Plan 34-02 can proceed with the draft, admission, and identity before-states frozen. The concurrent `.planning/config.json` edit remains unstaged and uncommitted.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*

## Self-Check: PASSED

All seven created artifacts and this summary exist. Task commits `a788aa8371716b35aa6cb25d609e8b6476c38f22`, `14b44fa3102da99a916f12551a99363ecbb44b75`, `58b69c1b4104b008bdaaaac49470e6d29046a602`, and `74fcfdd91449ea353f3857b124e1951ed10316cd` are ancestors of HEAD.
