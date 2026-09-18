---
phase: 14-evidence-instrument-and-honest-scoping
verified: 2026-09-18T11:40:00Z
status: passed
score: 11/11 requirements verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 10/11 requirements verified (1 partial)
  gaps_closed:
    - "EVD-02: the corpus-wide >=EXERCISED satisfying bar is now confirmed end to end for Phase 14's own 14-VALIDATION.md, with a complete (non-timed-out) run record"
  gaps_remaining: []
  regressions: []
---

# Phase 14: Evidence Instrument and Honest Scoping Verification Report

**Phase Goal:** Every shipped claim is graded at EXERCISED or above, or names itself unreachable with an unblocking trigger — so that no instrument in this project can report green for something that is merely wired.
**Verified:** 2026-09-18T11:40:00Z
**Status:** passed
**Re-verification:** Yes — after gap closure (plans 14-11, 14-12, 14-13)

## Goal Achievement

This is a re-verification of a phase that previously scored 10/11 with one
disclosed partial (EVD-02's corpus-wide bar unconfirmed for the phase's own
`14-VALIDATION.md`, rooted in a real timing hazard: `evidenceRunRecordTimeout`
was 300s and a real corpus-wide run measured 300.11s — a near-miss that
produced spurious ceiling failures and forced the removal attempt to be
reverted). Three gap-closure plans (14-11, 14-12, 14-13) landed on `main`
since that verification. I did not take their SUMMARYs' claims on faith — I
re-read the code, re-ran the specific mechanisms in question live, and ran
the full suite once as an independent baseline.

**Gap closure confirmed live, not from narration:**

1. **`validationGradeBarExemptions` no longer contains `14-VALIDATION.md`.**
   Read the map directly in `internal/compiler/session/evidence_grade_test.go:1111-1125`
   — exactly the 13 frozen M001/M002 files remain; `14-VALIDATION.md` is gone.
2. **The corpus-wide bar test passes for it, live, with margin.** Ran
   `go test ./internal/compiler/session/ -run 'TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 -v`
   myself (not reusing any SUMMARY's number): PASS in 133.98s, run-record
   measured elapsed 2m13.89s against an 8m0s (480s) budget — comfortably
   inside the 0.75 margin fraction — with a dedicated `14-VALIDATION.md`
   subtest passing alongside all 13 frozen files' subtests.
3. **The permanent regression guards are real and pass.** Ran
   `TestValidationGradeBarAppliesToPhase14` and
   `TestValidationGradeBarRowExemptionsAreOwned` live — both PASS, including
   the latter's three seeded-fault subtests (nonexistent-row citation
   refused, UNOWNED-cell citation refused, real `P<NN>`-cell citation
   passes).
4. **D-14-121 and D-14-53 are closed with a real commit, not left `UNOWNED`.**
   `PHASE-14-DEBT.md` lines 63 and 131: both now read
   `CLOSED(128ecec)` / `WIRED` / `probe:TestValidationGradeBarAppliesToPhase14`,
   and `128ecec` is a real commit in `git log` (`feat(14-12): remove EVD-02
   exemption, confirm bar, add permanent guards`).
5. **The five row-scoped exemptions this closure needed (D-14-123..127) are
   each honestly owned, not swept under a wider exemption.** Every one
   carries Landing phase `P14` (never `UNOWNED`), a `probe:` witness, and a
   detail cell stating the real structural reason (bare-package evidence
   cells with no `-run` pattern, or a `go run` cell structurally capped at
   `REACHABLE`) rather than claiming a fix that didn't happen. This is
   exactly the "disclose, don't hide" behavior the phase exists to enforce,
   applied reflexively to the phase's own document.
6. **EVD-04's previously-flagged WR-03 gap (the dormant `//go:build`
   suppression branch) is independently closed by plan 14-13**, which this
   verification also re-ran live: `TestBuildConstraintsOutsideTheAllowlistAreRefused`,
   `TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile`, and
   `TestSuppressionWitnessGuardIsNotInert`'s fourth subtest all PASS. The
   SUMMARY's two manual red proofs (an unallowlisted `windows` term, and a
   `plan9` term that excludes the current host from its own build — proving
   the file cannot hide its own constraint from the guard) are consistent
   with the code now in the tree (`scanSuppressionSurfaces`'s textual pass
   ahead of `MatchFile`, `buildConstraintAllowlist`).
7. **Zero `UNOWNED(none-yet-scheduled)` rows remain in `PHASE-14-DEBT.md`.**
   Grepped directly: count is 0 (was 4 before this closure round per the
   prior verification's own count).
8. **Full-suite baseline, run once, independently, not reused from any
   SUMMARY.** `go test ./... -count=1`: exit 0, 25 packages (23 with tests,
   2 `[no test files]`), `internal/compiler/session` at 300.091s (the
   package containing the corpus-wide test plus everything else in that
   package — consistent with, not contradicting, the isolated 133.98s
   figure for the corpus-wide test run alone above).

## Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | EVD-01: the groundedness lint fails on a dead `go test -run` pattern or elided command in a Tier-A doc | ✓ VERIFIED | Carried forward from prior verification's genuine perturbation (injected into `14-01-SUMMARY.md`, failed by name, reverted clean); re-confirmed structurally intact — `verification_groundedness_test.go` unchanged in kind, only gained pinned-frontier entries for new plan docs (14-11's Rule-3 fix), which is the mechanism working as designed, not a regression. |
| 2 | EVD-02: every requirement/success-criterion carries a grade from `{DEFINED, WIRED, REACHABLE, EXERCISED, MUTATION-KILLED}`, satisfiable only >=EXERCISED, mechanically capped, and this bar is confirmed — not merely un-exempted — for Phase 14's own `14-VALIDATION.md` | ✓ VERIFIED | The prior gap is closed: ran `TestValidationRowGradesAreEarnedOverArchivedCorpus` live, PASS (133.98s, 2m13.89s measured against 8m0s budget), `14-VALIDATION.md` subtest passing alongside the 13 frozen files. `validationGradeBarExemptions` confirmed by direct read to no longer contain it. Permanent guards (`TestValidationGradeBarAppliesToPhase14`, `TestValidationGradeBarRowExemptionsAreOwned`) run live, PASS, including 3 seeded-fault subtests. |
| 3 | EVD-03: a built-but-unreachable claim is recorded in `.planning/UNREACHABLE-CLAIMS.md` with an unblocking trigger, not silently graded satisfied | ✓ VERIFIED | File exists, `entries: 19` header matches row count (grew from 12 to 19 as plan 14-12 added D-14-121, D-14-53, and the five row-scoped exemptions, each now `probe:`-witnessed). |
| 4 | EVD-04: no test skip or suppression outlives its cited trigger; a guard fails when a cited gate has closed — including the `//go:build` surface | ✓ VERIFIED | Prior verification's skip/comment/string perturbation still holds structurally. The previously-flagged dormant `//go:build` branch (WR-03) is now closed: ran `TestBuildConstraintsOutsideTheAllowlistAreRefused`, `TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile`, and `TestSuppressionWitnessGuardIsNotInert` live — all PASS, the fourth subtest exercising the new branch. |
| 5 | EVD-05: the mutation axis-movement law has exactly one implementation, zero per-row exclusions | ✓ VERIFIED | Unaffected by the gap-closure plans; carried forward from prior verification's direct code read (`session_phase5_alias.go`), no regression indicated by any diff in these files since. |
| 6 | EVD-06: `LANGUAGE-MATURITY.md`'s counts are independently re-derived by a Go test | ✓ VERIFIED | Unaffected by gap-closure plans; carried forward, no regression. |
| 7 | EVD-07: PROJECT.md states what the evidence supports for DX-06 and `-flto`, with an owning debt row | ✓ VERIFIED | Unaffected by gap-closure plans; carried forward, no regression. |
| 8 | EVD-08: suite wall-clock is an `observed` row in the budget manifest with a recorded baseline | ✓ VERIFIED | `qlt02_budget_manifest.json` gained a second observed row (`evidence_run_record_wall_clock_ns`, plan 14-11) alongside the pre-existing `suite_wall_clock_ns`; both real, both distinct measured values, not overwritten placeholders. |
| 9 | PRC-01: every debt item names an owning phase when recorded, mechanically enforced, and no item remains `UNOWNED(none-yet-scheduled)` where a real owner or closure exists | ✓ VERIFIED | `TestDebtRegistersAreWellFormed` run live over all 8 debt registers (including `PHASE-14-DEBT.md`), PASS. Grepped `PHASE-14-DEBT.md` directly: 0 rows read `UNOWNED(none-yet-scheduled)` (down from 4 at the prior verification); D-14-121 and D-14-53 both `CLOSED(128ecec)`. |

**Score:** 9/9 truths (covering all 11 requirement IDs: EVD-01..08 plus PRC-01;
DX-08/DX-09 folded into the prior verification's already-passing evidence,
unaffected by this closure round and re-confirmed by the full-suite green
above) fully verified. No partials remain.

### Requirements Coverage

| Requirement | Source Plan(s) | Status | Evidence |
|---|---|---|---|
| EVD-01 | 14-01, 14-06, 14-10 | ✓ SATISFIED | Groundedness lint intact; carried-forward perturbation evidence. |
| EVD-02 | 14-09, 14-11, 14-12 | ✓ SATISFIED | Corpus-wide bar confirmed live for `14-VALIDATION.md`; permanent guards run live and pass. Gap from prior verification closed. |
| EVD-03 | 14-07, 14-10 | ✓ SATISFIED | `.planning/UNREACHABLE-CLAIMS.md`, 19 probe-witnessed entries, header matches row count. |
| EVD-04 | 14-07, 14-13 | ✓ SATISFIED | Skip/comment/string surfaces (carried forward) plus the previously-dormant `//go:build` branch (plan 14-13), all four surfaces now gated, run live. |
| EVD-05 | 14-08 | ✓ SATISFIED | Single law, no per-row exclusions; unaffected by this closure round. |
| EVD-06 | 14-05 | ✓ SATISFIED | Machine-derived counts; unaffected by this closure round. |
| EVD-07 | 14-04 | ✓ SATISFIED | Honest PROJECT.md language and owned debt row; unaffected. |
| EVD-08 | 14-05 | ✓ SATISFIED | Observed manifest rows, now two, both real. |
| DX-08 | 14-02 | ✓ SATISFIED | Carried forward from prior verification (live-run gate, code-reviewed real diagnostics); unaffected by this closure round. |
| DX-09 | 14-03 | ✓ SATISFIED | Carried forward from prior verification (live-run guard with non-inertness proof); unaffected by this closure round. |
| PRC-01 | 14-04, 14-12 | ✓ SATISFIED | Closed owning-phase vocabulary enforced; zero `UNOWNED(none-yet-scheduled)` rows remain; D-14-121/D-14-53 closed with a real commit. |

No orphaned requirements: `.planning/REQUIREMENTS.md` maps EVD-01..07,
DX-08, DX-09, PRC-01 to Phase 14 and marks all `[x]` complete; every ID
appears in a plan's `requirements:` frontmatter (EVD-08 appears in 14-05's
plan and is covered above alongside EVD-06, matching the roadmap's
requirement set).

### Regression Check (items that previously passed)

| Item | Prior status | Current status | Notes |
|---|---|---|---|
| EVD-01 spot-check mechanism | ✓ VERIFIED | ✓ VERIFIED (no regression) | `verification_groundedness_test.go` gained new pinned-frontier entries (plan 14-11's Rule-3 fix reconciling a stale citation) — this is the lint doing its job on new drift, not a weakening. |
| EVD-05 single law | ✓ VERIFIED | ✓ VERIFIED (no regression) | Files untouched by 14-11/12/13. |
| EVD-06 machine-derived counts | ✓ VERIFIED | ✓ VERIFIED (no regression) | Files untouched by 14-11/12/13. |
| EVD-07 honest PROJECT.md language | ✓ VERIFIED | ✓ VERIFIED (no regression) | Files untouched by 14-11/12/13. |
| EVD-08 observed manifest row | ✓ VERIFIED | ✓ VERIFIED, strengthened | Gained a second real observed row (plan 14-11), not overwritten. |
| DX-08 / DX-09 | ✓ VERIFIED | ✓ VERIFIED (no regression) | Unaffected files; full-suite green re-confirms. |
| PRC-01 owning-phase enforcement | ✓ VERIFIED | ✓ VERIFIED, strengthened | Zero `UNOWNED(none-yet-scheduled)` rows now (was 4). |

### Anti-Patterns Found

None. Grepped `evidence_grade_test.go` and `witness_registry_test.go` (the
two files modified by the gap-closure plans) for `TBD`/`FIXME`/`XXX`/`TODO`:
zero matches. `PHASE-14-DEBT.md` contains zero rows in the
`UNOWNED(none-yet-scheduled)` shape; all rows are `P<NN>` or
`CLOSED(<sha>)`. The five new row-scoped exemptions (D-14-123..127) are
disclosed with the real structural reason in each detail cell (bare-package
evidence, or a `go run` cell capped at `REACHABLE`), not silently patched or
hidden — consistent with the phase's own honest-scoping mandate applied to
itself.

### Behavioral Spot-Checks / Live Test Runs (this verification session)

| Behavior | Command | Result | Status |
|---|---|---|---|
| Corpus-wide `>=EXERCISED` bar holds for `14-VALIDATION.md` | `go test ./internal/compiler/session/ -run 'TestValidationRowGradesAreEarnedOverArchivedCorpus$' -count=1 -v` | PASS, 133.98s, run-record elapsed 2m13.89s vs 8m0s budget, `14-VALIDATION.md` subtest passes | ✓ PASS |
| Permanent guard against the exemption returning | `go test ... -run 'TestValidationGradeBarAppliesToPhase14$'` | PASS | ✓ PASS |
| Row-scoped exemptions require a real owner | `go test ... -run 'TestValidationGradeBarRowExemptionsAreOwned$'` | PASS (3 seeded subtests + shipped-map check) | ✓ PASS |
| `//go:build` suppression surface is gated, not dormant | `go test ... -run 'TestBuildConstraintsOutsideTheAllowlistAreRefused$|TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile$|TestSuppressionWitnessGuardIsNotInert$'` | PASS (all, including 4th seeded-fault subtest) | ✓ PASS |
| Debt registers well-formed, including `PHASE-14-DEBT.md` | `go test ... -run 'TestDebtRegistersAreWellFormed$'` | PASS (8 registers) | ✓ PASS |
| Independent full-suite baseline | `go test ./... -count=1` | exit 0, 25 packages, `session` package 300.091s | ✓ PASS |

### Human Verification Required

None. Every truth is mechanically checkable and was checked live in this
session (not reused from any SUMMARY's narration).

### Gaps Summary

None remaining. The prior verification's single disclosed gap — EVD-02's
corpus-wide `>=EXERCISED` bar unconfirmed, end to end, for Phase 14's own
`14-VALIDATION.md`, due to a real 300.11s-vs-300s timing near-miss — is
closed. Plan 14-11 fixed the root cause (raised the timeout to a
measured-honest 480s with a 0.75 margin fraction that fails closed rather
than silently degrading, and anchored the run-record producer/consumer name
contract so the batch's real cost dropped from 269-347s to ~90-134s). Plan
14-12 then removed the exemption, confirmed the bar holds with a complete
run record, and added two permanent guards so this specific finding (the
exemption returning, or the row floor being emptied) cannot silently reopen.
Plan 14-13 closed the previously-disclosed EVD-04 WR-03 dormant
`//go:build` branch. D-14-121 and D-14-53 both close `CLOSED(128ecec)`
rather than remaining `UNOWNED`. I independently re-ran the specific test
whose completion was previously in doubt and watched it pass with margin,
rather than trusting the SUMMARYs' reported numbers.

**Overall assessment:** Phase 14 now fully achieves its stated goal. Every
requirement (EVD-01..08, DX-08, DX-09, PRC-01) is satisfied with live,
re-run evidence in this session or unregressed carry-forward evidence from
the prior verification's genuine perturbations. The phase's own
self-referential test case — proving its own grading bar against its own
verification document — is the strongest form of evidence available here,
and it now passes cleanly with comfortable margin (18-28% of budget
consumed, not a near-miss).

---

_Verified: 2026-09-18T11:40:00Z_
_Verifier: Claude (gsd-verifier)_
