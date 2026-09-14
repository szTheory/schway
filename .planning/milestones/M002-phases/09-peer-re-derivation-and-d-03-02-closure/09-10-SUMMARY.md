---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 10
subsystem: compiler-validation
tags: [nyquist-closure, requirements, roadmap, debt-register, ownership, documentation]

requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "All ten plans' evidence (09-01 through 09-09), and plan 09-08's mandatory mid-phase gate adjudication of every open debt item"
provides:
  - "QLT-07's loan-liveness subset closed and ratified in 09-VALIDATION.md's 'M001 Phase 3 Debt Closure (loan-liveness subset)' section, re-verified live rather than trusted from planning time"
  - "A single non-mutating pointer line appended to the archived M001 03-VALIDATION.md, its frontmatter and checkboxes byte-identical (git diff --numstat: 1 added, 0 deleted)"
  - "OWN-05 split into OWN-05a (Phase 09, Complete) and OWN-05b (Phase 10, Pending), REQUIREMENTS.md traceability arithmetic reconciled (31 total)"
  - "OWN-09's own requirement text corrected to Phase 09, closing D-08-27/D-09-14"
  - "TRU-04's own text carries the cycle-peer differential disposition (D-09-22)"
  - "ROADMAP's S-008 spike row amended to record its replacement by a bounded inventory pass, as an explicit process amendment (D-09-39); Phase 09 marked complete"
  - "PHASE-09-DEBT.md closing state: 14 of 16 items resolved/superseded-closed; D-09-51 and D-09-53 both given an explicit landing phase (Phase 10)"
affects: ["Phase 10 (candidate landing site for D-09-51 and D-09-53; inherits OWN-05b, SEM-08/09, TRU-02/03, QLT-04)"]

actuals:
  tokens: 29250
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Supplemental-filing pattern: an archived approval record is never rewritten in place, only pointed to by a single appended, non-mutating line from the superseding document"
    - "Re-verify, don't trust: a planning-time inventory result is re-run live at closure time before being ratified, rather than assumed still true after nine plans landed"

key-files:
  created: []
  modified:
    - .planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
    - .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md
    - .planning/REQUIREMENTS.md
    - .planning/ROADMAP.md
    - .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md

key-decisions:
  - "09-09-03's row (Pattern B's twin pair demonstrating a differing CLI verdict) is marked RED in 09-VALIDATION.md's Per-Task Verification Map rather than silently folded into a green closure, because the split does not materialize (D-09-53). This does not block nyquist_compliant: true, because OWN-09's core claim (exactly one loan-liveness decision point) is independently evidenced by 09-09-SUMMARY.md's D1-D3 and does not depend on Pattern B's split — the split was the phase's strongest possible corroborating observation, not its defining criterion, and its own plan (09-09 Task 4(b)) named this exact outcome as an anticipated contingency with a prescribed handling (report, don't accommodate)."
  - "D-09-53's landing phase, left undecided at plan 09-09's end, is decided here as Phase 10 (alongside D-09-51), rather than left dangling per Key Lesson 4 -- both are pre-existing, Phase-08-vintage validator defects this phase's authorized deletions unmasked one layer at a time, neither a Phase 09 regression."
  - "OWN-08 and TRU-04 flipped to Complete in this plan (previously carried Pending pending this plan's own gate, per 09-08-SUMMARY.md's explicit requirement-marking rule) -- verified against 09-03-SUMMARY.md (D-03-02 closed in both admission layers) and the combined evidence of 09-01/09-04/09-06 (zero-divergence differential, seeded-fault companion, cost bound, cycle-peer differential) respectively."
  - "The M001 03-VALIDATION.md pointer line was written without a preceding blank line, specifically because git counts a blank separator line as an additional added line under --numstat -- the plan's own verify command requires exactly one added line, so the blockquote is appended directly after the prior paragraph's closing line."

patterns-established: []

requirements-completed: [QLT-07, OWN-05a, OWN-09, TRU-04]

coverage:
  - id: D1
    description: "QLT-07's loan-liveness subset closed in 09-VALIDATION.md, naming exactly the 03-03/03-04/03-05 rows it claims and re-stating all three exclusions (OWN-04 rows, loop-carried-liveness limitation, 03-VALIDATION.md's own nyquist_compliant: false staying false); the inventory verification command re-run live and its result recorded"
    requirement: "QLT-07"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
  - id: D2
    description: "03-VALIDATION.md (archived M001 document) receives exactly one appended, non-mutating pointer line; frontmatter and checkboxes byte-identical before and after; nyquist_compliant: false unchanged"
    requirement: "QLT-07"
    verification:
      - kind: other
        ref: "git diff --numstat -- .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md (1 added, 0 deleted)"
        status: pass
      - kind: other
        ref: "grep -c 'nyquist_compliant: false' .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md = 2 (frontmatter + reaffirming pointer-line prose)"
        status: pass
    human_judgment: false
  - id: D3
    description: "OWN-05 split into OWN-05a (Phase 09, Complete) and OWN-05b (Phase 10, Pending); no unsplit OWN-05 row remains; traceability arithmetic (31 total, per-phase counts) reconciled"
    requirement: "OWN-05a"
    verification:
      - kind: other
        ref: "grep -c OWN-05a .planning/REQUIREMENTS.md = 5; grep -c OWN-05b = 4; grep -n '| OWN-05 |' .planning/REQUIREMENTS.md exits 1 with no output"
        status: pass
    human_judgment: false
  - id: D4
    description: "OWN-09's own requirement text corrected to say Phase 09 and 'decision point' rather than 'law', closing D-08-27/D-09-14; TRU-04's own text carries the cycle-peer differential disposition (D-09-22)"
    requirement: "OWN-09"
    verification:
      - kind: other
        ref: "grep -n cycle-peer .planning/REQUIREMENTS.md"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
  - id: D5
    description: "ROADMAP's S-008 row amended (never deleted) to record its replacement by a bounded inventory pre-flight pass, with an amendment note stating both reasons; Phase 09 marked complete with no TBD remaining under its heading; PHASE-09-DEBT.md carries a dated closing paragraph giving every row an end-of-phase state"
    requirement: "QLT-07"
    verification:
      - kind: other
        ref: "grep -c S-008 .planning/ROADMAP.md = 6; grep -n TBD .planning/ROADMAP.md shows none under the Phase 09 heading"
        status: pass
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: true
    rationale: "Whether the QLT-07 closure's wording claims no more than it proves, and whether OWN-05's split (rather than a partial single row) is the honest disposition, are judgments about scope wording that this plan's own Manual-Only Verifications table requires a human reviewer to confirm -- not a testable predicate."

duration: ~50min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 10: QLT-07's Scoped Closure, the OWN-05a/OWN-05b Split, and the Requirement and Roadmap Corrections Summary

**Closed QLT-07's loan-liveness Nyquist subset in Phase 09's own 09-VALIDATION.md (re-verified live, not trusted from planning time), appended exactly one non-mutating pointer line to the archived M001 03-VALIDATION.md, split OWN-05 into OWN-05a/OWN-05b, corrected OWN-09's and TRU-04's own requirement texts, amended ROADMAP's S-008 spike row as an explicit process substitution, and gave every PHASE-09-DEBT.md item an explicit end-of-phase state — marking Phase 09 complete.**

## Performance

- **Duration:** ~50 min
- **Tasks:** 3 completed
- **Files modified:** 5 (0 created)

## Accomplishments

- **Task 1 — QLT-07 closure, pointer-only M001 edit:** Ratified the pre-drafted "M001 Phase 3 Debt Closure (loan-liveness subset)" section in `09-VALIDATION.md`, re-running D-09-40a's inventory verification command live (`go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle`, all three `ok`) rather than trusting the planning-time result after nine plans had landed. Ticked every Per-Task Verification Map row, Required Negative Control, Wave 0 Requirement, Mutation-Kill Register row, and Generator Reachability Register row from a named plan SUMMARY's evidence — except 09-09-03 (Pattern B's promised twin split), which is marked **red** with an explanation naming D-09-53, honestly documenting the one place evidence did not support the original expectation. Appended exactly one non-mutating pointer line to the archived `03-VALIDATION.md`, verified via `git diff --numstat` (1 added, 0 deleted) and a grep confirming `nyquist_compliant: false` still reads false.
- **Task 2 — Requirement corrections:** Corrected OWN-09's own text to name Phase 09 as the intraprocedural decision point's retirement phase (one phase after Phase 08's interprocedural law), using "decision point" instead of "law" per D-09-08's own finding that there was never a second law. Split OWN-05 into OWN-05a (Phase 09: `check` + `corevalidate`, Complete) and OWN-05b (Phase 10: `interp`, Pending), updating the traceability table's total (31), mapped count, and per-phase counts. Wrote TRU-04's cycle-peer differential disposition directly into its own requirement text. Flipped OWN-05a, OWN-08, OWN-09, TRU-04, and QLT-07 to Complete (OWN-07 was already Complete from plan 09-08); OWN-05b stays Pending against Phase 10.
- **Task 3 — S-008 amendment, ROADMAP closure, debt-register finalization:** Amended (never deleted) S-008's spike-table row to record its replacement by a bounded inventory/estimation pre-flight pass, with a dedicated amendment note beneath the table stating both reasons (category error against the spike discipline; avoiding a second D-08-43-shaped registry-gap reproduction). Confirmed spike 006's `MANIFEST.md` entry already existed (added by plan 09-02). Marked Phase 09 complete in ROADMAP's Phases list and its own Wave 6 plan checkbox — all five Phase 09 success criteria landed across plans 09-01 through 09-09, and the Phase 09 section already listed every plan with no `TBD`. Added a dated "Phase 09 Closing State" section to `PHASE-09-DEBT.md` giving all 16 items an explicit end-of-phase state: 14 resolved/superseded-and-closed, and D-09-51/D-09-53 both explicitly landing in Phase 10 (D-09-53's landing phase was undecided at plan 09-09's end; decided here).

## Task Commits

1. **Task 1: Close QLT-07's loan-liveness subset, and touch the archived M001 document exactly once** — `7396108` (docs)
2. **Task 2: Correct the requirement documents — including the two that must stay partial** — `2e085a4` (docs)
3. **Task 3: Record the S-008 process amendment, close the spike-table gap, and finalize the debt register** — `2598ef5` (docs)

**Plan metadata:** (this commit)

## Files Created/Modified

- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md` — closure section ratified, every table's Status column ticked from named evidence (except 09-09-03, marked red), frontmatter flipped to `status: validated`, `nyquist_compliant: true`, `wave_0_complete: true`
- `.planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md` — exactly one non-mutating pointer line appended; frontmatter and checkboxes byte-identical
- `.planning/REQUIREMENTS.md` — OWN-09 text corrected; OWN-05 split into OWN-05a/OWN-05b; TRU-04 text carries cycle-peer disposition; QLT-07 text names its exact scope; traceability table and coverage arithmetic updated; six requirement checkboxes flipped
- `.planning/ROADMAP.md` — S-008 row amended and an amendment note added; Phase 09 marked complete in the Phases list and its Wave 6 plan checkbox
- `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md` — D-09-53's landing-phase cell and detail-section status updated to name Phase 10 explicitly; dated Phase 09 Closing State section added

## Decisions Made

See `key-decisions` in frontmatter for the full account: 09-09-03's honest red marking and why it does not block `nyquist_compliant: true`; D-09-53's landing-phase decision (Phase 10); OWN-08/TRU-04's Complete flip and the evidence behind it; and the pointer-line formatting choice driven by `git diff --numstat`'s blank-line counting behavior.

## Deviations from Plan

### Auto-fixed Issues

None — plan executed exactly as written. No production code was touched, matching the plan's own prohibition.

---

**Total deviations:** 0.
**Impact on plan:** None.

## Known Stubs

None introduced by this plan. Two carried debt items remain open in `PHASE-09-DEBT.md`, both explicitly landing in Phase 10 (not silently dropped, per this plan's own Task 3(d)):

- **D-09-51** — `originvalidate.walkReturnOrigin` has no `case core.OpCall`, so it treats a call boundary as transparent rather than consulting the callee's declared return contract. Pre-existing, Phase-08-vintage, unmasked by this phase's `corevalidate` fix. Landing phase: Phase 10 (ROADMAP's own Success Criterion 1 already names the fix).
- **D-09-53** — `deriveFunctionUsesParam`'s `default` branch treats a plain `OpMove`/`OpCopy` forward-to-return as a "use," contradicting its own doc comment's exemption, collapsing Pattern B's promised twin-fixture split. Pre-existing, Phase-08-vintage, unmasked by this phase's `computeLoanLastUses` deletion. Landing phase: Phase 10, decided at this plan.

## Threat Flags

None — this plan's threat register items (T-09-30, T-09-16, T-09-31, T-09-32, T-09-33) are all directly mitigated: T-09-30 by the `git diff --numstat` verification (1 added, 0 deleted) and the `nyquist_compliant: false` grep; T-09-16 by the OWN-05 split and the greps confirming no unsplit row remains; T-09-31 by the closure section naming its exact rows and re-stating all three exclusions, confirmed by this plan's own Manual-Only Verification; T-09-32 by the Scope-Cut Order confirmation (QLT-07 committed, item 2 never triggered, no cut recorded); T-09-33 by the S-008 row being amended in place rather than deleted, with both reasons stated in the amendment note.

## Issues Encountered

**`git diff --numstat`'s blank-line counting.** The first attempt at the M001 pointer-line append included a blank separator line before the blockquote (for markdown readability), which `git diff --numstat` counted as 2 added lines — failing the plan's own "exactly one added line" verification. Reverted and re-applied the blockquote directly after the prior paragraph's closing line, with no blank separator, yielding exactly 1 added / 0 deleted. Documented here since a future editor of this file should not "fix" the missing blank line without re-checking this constraint.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- Phase 09 is complete: all six requirements (OWN-05a, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07) are Complete; OWN-05b is correctly Pending against Phase 10.
- QLT-07's loan-liveness Nyquist subset is closed and ratified; M001's `03-VALIDATION.md` is referenced, never rewritten.
- Two carried debt items (D-09-51, D-09-53) both land explicitly in Phase 10 — Phase 10 planning should treat both as known agenda items alongside its own charter (SEM-08/09, OWN-05b, TRU-02/03, QLT-04).
- `go test ./... && go vet ./...` exits 0 at HEAD.
- No blockers for Phase 10.

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

- FOUND: .planning/phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
- FOUND: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-VALIDATION.md
- FOUND: .planning/REQUIREMENTS.md
- FOUND: .planning/ROADMAP.md
- FOUND: .planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md
- FOUND commit 7396108 (docs(09-10): ratify QLT-07's loan-liveness closure, pointer-only M001 edit)
- FOUND commit 2e085a4 (docs(09-10): correct OWN-09 text, split OWN-05, write TRU-04's cycle-peer disposition)
- FOUND commit 2598ef5 (docs(09-10): record S-008's replacement, close Phase 09 in ROADMAP, finalize debt register)
- `go test ./... && go vet ./...` exits 0 (verified before this SUMMARY was finalized)
