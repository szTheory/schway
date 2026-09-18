---
phase: 14-evidence-instrument-and-honest-scoping
recorded: 2026-09-18
status: accepted
disposition: phase-in-progress
items: 78
blocking: 0
---

# Phase 14 — Evidence Instrument and Honest Scoping Debt Register

**Written:** 2026-09-18, at plan 14-04, the first plan in this phase to open a
`*-DEBT.md` register. Named `PHASE-14-DEBT.md` per the M003 naming convention
(D-14-24): any register created from M003 on is named `PHASE-<NN>-DEBT.md`, no
exemption. Shape follows the mechanically-checked debt-register format
`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`)
enforces on every `*-DEBT.md` register: an `items:` count matching the
`## Items` table, one `### <ID>` detail section per row (each carrying a
`first-recorded:` milestone line, D-14-31), a severity from the closed
vocabulary (blocker, warning, info), and a landing phase drawn from the
closed three-form owning-phase vocabulary this same plan mechanized
(`P<NN>` | `CLOSED(<commit-sha>)` | `UNOWNED(<witness-id>)`, PRC-01).

This register opens with three items: EVD-07's outstanding half (the
`-flto` multi-function inertness claim, which had no debt row before this
plan), and two debt rows deferred here by plan 14-02's SUMMARY under its own
"Debt rows to register" section. Plan 14-09 (EVD-02's grade cap) adds seven
more: six rows whose declared grade derives below its shipped `✅ green`
verdict once the mechanized cap is applied -- the instrument's first real
findings, D-14-07's predicted outcome, reported honestly rather than
suppressed -- and one row recording that 14-VALIDATION.md's own Per-Task
Verification Map was never filled in beyond its plan-time template. Plan
14-10 (the Nyquist reconciliation) adds 66 more (D-14-55..D-14-120): one
row per runnability, groundedness, or grep finding in the pinned
groundedness frontier, each a "debt row with a witness" under the closed
reconciliation verdict vocabulary (`renamed` | `superseded` |
`obsolete-by-design` | `under-scoped`, D-14-12) rather than a rewrite of
the archived document the finding was found in. Every such row carries a
`` ```reconciliation ``` `` fenced block in its own Detail section, parsed
and mechanically checked by `TestReconciliationVerdictsCarryTheirObligations`
(`internal/compiler/session/session_test.go`); the Items table's own Grade
and Witness cells stay `DEFINED`/`n/a` for these rows deliberately -- the
obligation check is a separate, dedicated law, not routed through the
Grade/Witness `probe:` grammar, so these 66 rows do not also inflate
`.planning/UNREACHABLE-CLAIMS.md` (that view qualifies rows by a `probe:`
witness token, and a reconciliation finding is not an unreachable-claim
finding).

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-14-45 | PROJECT.md `## Current State` (verified already correct, commit `d21db90`); 14-RESEARCH.md Open Question 1 | EVD-07 | warning | UNOWNED(probe:TestLTOInertnessOnMultiFunctionEmission) | WIRED | probe:TestLTOInertnessOnMultiFunctionEmission | `-flto` IS STRUCTURALLY INERT ON EVERY CGEN-EMITTED MULTI-FUNCTION PROGRAM, AND THIS CLAIM PREVIOUSLY HAD NO DEBT ROW. `cgen.emitProgram` emits one translation unit with no `restrict` attribute, and a foreign-boundary program is refused outright by multi-function emission -- so LTO's whole-program optimization has no cross-TU boundary to exploit and no alias attribute to hoist across. The two existing non-inertness proofs are both non-multi-function: `TestLTOTierIsNotInert` uses a single-function foreign fixture, and NAT-07's composition-only control is hand-written C explicitly not emitted by `cgen` (D-11-24). D-11-25 names this only in a test doc comment -- it has no debt row and was not among the ten unowned items PROJECT.md's `## Current State` enumerates. This row closes that gap |
| D-14-46 | 14-02-SUMMARY.md "Debt rows to register" item 1 | DX-08 | info | UNOWNED(conditional-surface-lands-in-language) | DEFINED | n/a | NOTHING FORCES THE DISTINCTNESS CORPUS TO GROW AS THE LANGUAGE GROWS. When the conditional form (`if`) enters the language, the spiral trio (`spiral_full.lang`, `spiral_narrow.lang`, `spiral_bare.lang`) becomes a VALID program and must be replaced in `testdata/distinctness/` by the then-current not-in-language surface, or the corpus predicate would need to be re-derived against a program that no longer refuses to parse |
| D-14-47 | 14-02-SUMMARY.md "Debt rows to register" item 2 | DX-08 | info | UNOWNED(identity-bearing-cause-span-on-parse-success-path) | DEFINED | n/a | PUTTING MORE CONTENT ON `Cause.Span` DEEPENS THE ALREADY-RECORDED, HALF-ENFORCED RULE THAT A COORDINATE SHIFT MUST NEVER MOVE A DIAGNOSTIC'S ID. This is now bounded-widened for broken programs only -- a whitespace edit inside a broken program's tail (past the point where declaration recovery starts) can now move that program's diagnostic ID, where before plan 14-02's fix the swallowed region was never in the identity basis. This is an acceptable, deliberately bounded widening of an already-recorded hole (Phase 13's coordinate-shift discipline), not a new defect, but it should be reconciled explicitly rather than rediscovered |
| D-14-48 | plan 14-09's grade derivation over 09-VALIDATION.md:85 | EVD-02 | warning | P14 | WIRED | probe:TestVerificationGroundednessFrontierIsPinned | 09-VALIDATION.MD ROW 09-01-01's SECOND EVIDENCE CELL (A `LoanChainIndex`-PATTERN GO TEST INVOCATION) CITES A PATTERN THAT MATCHES ZERO TEST NAMES -- IT WAS SHIPPED `✅ GREEN` BUT DERIVES ONLY WIRED. This is the same row plan 14-01's groundedness lint independently pinned as an R2 dead pattern in its own 125-entry frontier; the two instruments agree from two different mechanisms. Plan 14-10's Nyquist reconciliation is the scheduled closer for the groundedness frontier's R1/R2 class, which this row is a member of |
| D-14-49 | plan 14-09's grade derivation over 09-VALIDATION.md:93 | EVD-02 | warning | P14 | WIRED | probe:TestVerificationGroundednessFrontierIsPinned | 09-VALIDATION.MD ROW 09-05-02's `Mode.*Invalid`-ALTERNATION EVIDENCE CELL ALSO MATCHES ZERO TEST NAMES -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED. Same cross-instrument agreement as D-14-48: also a member of plan 14-01's pinned R2 frontier, closure scheduled at plan 14-10 |
| D-14-50 | plan 14-09's grade derivation over 12-VALIDATION.md:52 (row 12-04-01) | EVD-02 | info | UNOWNED(probe:TestValidationGradeCapBarePackageRowHasNoNamedTest) | WIRED | probe:TestValidationGradeCapBarePackageRowHasNoNamedTest | 12-VALIDATION.MD ROW 12-04-01's EVIDENCE CELL NAMES TWO PACKAGES BUT NO `-run`/`-list`/`-fuzz`/`-bench` PATTERN, SO IT NAMES NO EXACT TEST IDENTIFIER THE CAP CAN CONFIRM EXECUTED -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED UNDER THE EXACT-IDENTIFIER DISCIPLINE. The package-wide invocation genuinely compiles and the code path is real; the row is simply not phrased as a resolvable claim. No phase currently owns rephrasing it |
| D-14-51 | plan 14-09's grade derivation over 04-VALIDATION.md:74 (row 04-06-03) | EVD-02 | warning | UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent) | WIRED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 04-VALIDATION.MD ROW 04-06-03's EVIDENCE CELL NAMES `TestLastUseDiscoveryWorkIsCounted` AND `TestLastUseDiscoveryWorkSeries`, NEITHER OF WHICH EXISTS IN THE CURRENT MODULE -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED. Both were real tests at the time 04-DEBT.md's "deferred retirement" item was written (04-06-SUMMARY.md cites them as coverage) and were retired alongside the `computeLoanLastUses` deletion (D-09-09) without this row being updated. This is a genuine hole in shipped M001 evidence being surfaced, not a regression to suppress -- the precedent D-13-33 and D-14-07 both name explicitly |
| D-14-52 | plan 14-09's grade derivation over 06-VALIDATION.md:69 (row 06-08-T3) | EVD-02 | warning | UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent) | WIRED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 06-VALIDATION.MD ROW 06-08-T3's EVIDENCE CELL NAMES `TestOnlyRecomputedWorkIsGateEligible`, WHICH DOES NOT EXIST -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED. Plan 08-05 renamed it to `TestOnlyGateEligibleMetricsPassThrough` when `GateEligibleMetrics()` widened from one metric to a set (08-05-SUMMARY.md's own key-decisions record the rename and why), and this M001 row was never repointed. Same class of finding as D-14-51: a real rename left a stale citation in a document already marked satisfied |
| D-14-53 | plan 14-09's own satisfying-bar exemption for 14-VALIDATION.md | EVD-02 | info | UNOWNED(none-yet-scheduled) | DEFINED | n/a | 14-VALIDATION.MD'S OWN PER-TASK VERIFICATION MAP HAS EXACTLY ONE ROW, AND ITS EVIDENCE CELL IS STILL THE LITERAL PLAN-TIME PLACEHOLDER `` `{command}` `` -- NO PLAN IN THIS PHASE EVER FILLED IT IN, BECAUSE EACH PLAN TRACKS ITS OWN VERIFICATION THROUGH ITS OWN `*-SUMMARY.md` INSTEAD. Recording this as a silent permanent bar exemption (`validationGradeBarExemptions`, `internal/compiler/session/evidence_grade_test.go`) would be exactly the suppression EVD-02 exists to retire, so it is named here instead: either a future plan populates this table for real from the phase's nine plans' own evidence, or the document is retired in favor of the per-plan SUMMARY.md convention it has in practice already lost to |
| D-14-54 | plan 14-09's grade derivation over 06-VALIDATION.md:71 (row 06-09-T3) | EVD-02 | warning | UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent) | WIRED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 06-VALIDATION.MD ROW 06-09-T3's EVIDENCE CELL NAMES `TestRecomputedWorkIsTheOnlyHardGate`, WHICH DOES NOT EXIST -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED. `08-REVIEW.md`'s own WR-01 warning named this exact staleness (the test's name asserted "only hard gate" after plan 08-05 widened the gate-eligible set to two metrics) and recommended a rename; the rename that landed is `TestRecomputedWorkHardGateBoundComparison`. A third instance of the same class as D-14-51/D-14-52 -- a real, even self-documented rename left a stale citation behind in an M001 row already marked satisfied |
| D-14-55 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `03-RESEARCH.md:849` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-56 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `03-RESEARCH.md:850` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-57 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `03-RESEARCH.md:852` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-58 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:560` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-59 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:561` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-60 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:563` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-61 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:565` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-62 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:566` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-63 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:567` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-64 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:568` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-65 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:569` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-66 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:569` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-67 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `04-RESEARCH.md:570` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-68 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `08-RESEARCH.md:507` cites a verification command over a symbol deliberately deleted from the tree, corrected outside the archive as an OBSOLETE-BY-DESIGN verdict naming the deleting phase, commit and the deleted symbol, confirmed absent (see Detail section). |
| D-14-69 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `08-RESEARCH.md:514` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-70 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `08-VALIDATION.md:53` cites a verification command over a symbol deliberately deleted from the tree, corrected outside the archive as an OBSOLETE-BY-DESIGN verdict naming the deleting phase, commit and the deleted symbol, confirmed absent (see Detail section). |
| D-14-71 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `08-VALIDATION.md:65` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-72 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-RESEARCH.md:706` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-73 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-RESEARCH.md:708` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-74 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-VALIDATION.md:87` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-75 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-VALIDATION.md:95` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-76 | plan 14-10's reconciliation of the pinned groundedness frontier's R3 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `09-VALIDATION.md:114` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-77 | plan 14-10's reconciliation of the pinned groundedness frontier's R3 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `09-VALIDATION.md:115` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-78 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `10-RESEARCH.md:684` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-79 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-RESEARCH.md:1041` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-80 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:28` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-81 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:53` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-82 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:58` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-83 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:60` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-84 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:62` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-85 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:63` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-86 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:65` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-87 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:66` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-88 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:70` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-89 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:71` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-90 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:73` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-91 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `12-RESEARCH.md:869` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-92 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `12-VALIDATION.md:26` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-93 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `07-VERIFICATION.md:87` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-94 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `03-RESEARCH.md:853` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-95 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `04-RESEARCH.md:554` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-96 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `05-RESEARCH.md:420` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-97 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:160` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-98 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-01-SUMMARY.md:162` cites a verification command over a symbol deliberately deleted from the tree, corrected outside the archive as an OBSOLETE-BY-DESIGN verdict naming the deleting phase, commit and the deleted symbol, confirmed absent (see Detail section). |
| D-14-99 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-01-SUMMARY.md:163` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-100 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-01-SUMMARY.md:165` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-101 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-01-SUMMARY.md:166` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-102 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:169` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-103 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:170` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-104 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:171` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-105 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:172` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-106 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:173` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-107 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:174` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-108 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:175` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-109 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:176` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-110 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:177` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-111 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:178` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-112 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:179` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-113 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:181` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-114 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-01-SUMMARY.md:183` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-115 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-RESEARCH.md:448` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-116 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-RESEARCH.md:459` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-117 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-VALIDATION.md:26` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-118 | plan 14-10's reconciliation of the pinned groundedness frontier's R3 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `ADVERSARIAL-SYNTHESIS.md:213` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-119 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `ADVERSARIAL-SYNTHESIS.md:214` cites a verification command over a symbol deliberately deleted from the tree, corrected outside the archive as an OBSOLETE-BY-DESIGN verdict naming the deleting phase, commit and the deleted symbol, confirmed absent (see Detail section). |
| D-14-120 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `ADVERSARIAL-SYNTHESIS.md:214` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-121 | plan 14-10's attempt to remove 14-VALIDATION.md's satisfying-bar exemption | EVD-02 | warning | UNOWNED(none-yet-scheduled) | DEFINED | n/a | REMOVING `14-VALIDATION.MD` FROM `validationGradeBarExemptions` AND RE-RUNNING `TestValidationRowGradesAreEarnedOverArchivedCorpus` PRODUCED SPURIOUS "DECLARED EXERCISED EXCEEDS THE CEILING WIRED" FAILURES ACROSS NINE OTHER, UNRELATED, ALREADY-FROZEN ARCHIVED FILES (01, 02, 06, 07, 08, 10, 11, 12, 13-VALIDATION.MD) IN THE SAME RUN. The whole corpus-wide test completed at 300.11s, suspiciously close to `evidenceRunRecordTimeout`'s 300s ceiling (lowered from 900s by plan 14-09's own deviation fix) -- consistent with the run-record generation not completing within its budget and several packages' citations falling back to an unresolved WIRED ceiling rather than their true EXERCISED grade. The exemption-removal change was reverted rather than landed under that risk; the corpus-wide grade cap over the real archived data was never actually re-verified with a complete run record in this plan |
| D-14-122 | plan 14-11's fail-closed run-record margin check, which surfaced this while measuring the corpus-wide batch's real cost | EVD-01 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `14-VERIFICATION.md:133` cites a seeded-fault placeholder name (`TestThisNameDoesNotExistAnywhereZZQQ`) inside a Non-Inertness Spot-Check's narrative description of an already-reverted perturbation, corrected outside the archive as a RENAMED verdict naming the real, resolving test the row's claim actually rests on (see Detail section). |

## Detail

### D-14-45 — the `-flto` multi-function inertness claim finally has a debt row

first-recorded: M003

**PROJECT.md's text was verified already correct before this row was added,
not re-edited.** The `## Current State` section, read in full at this plan's
Task 2, already states:

> "**`-flto` is structurally inert on every cgen-emitted multi-function
> program** — one translation unit, no `restrict`, foreign refused. Both live
> non-inertness proofs are non-multi-function: `TestLTOTierIsNotInert` uses a
> single-function foreign fixture, and NAT-07's composition-only control is
> hand-written C explicitly not emitted by cgen (D-11-24). D-11-25 names this
> only in a test doc comment — it has **no debt row** and is **not** among
> the ten unowned items — while this document's NAT-07 bullet previously read
> as though the proof covered the interprocedural corpus. It does not."

This correction landed in commit `d21db90` (2026-09-17, "docs: start
milestone M003"), before Phase 14 opened. EVD-07 therefore had exactly **one**
remaining task for this plan: add the missing debt row PROJECT.md's own text
already names as absent. No PROJECT.md rewrite was needed or performed.

**Debt-row absence independently confirmed** before this row was added:
`grep -rn "D-11-25" .planning/milestones/M002-phases/*/PHASE-*-DEBT.md`
returns zero matches. The only prior mentions of D-11-25 are a test doc
comment (`session_phase11_differential_test.go`) and this phase's own
planning documents (14-RESEARCH.md, 14-01-SUMMARY.md), neither of which is a
debt register.

**Landing phase:** `UNOWNED(probe:TestLTOInertnessOnMultiFunctionEmission)`.
No phase currently claims closing this — closing it would mean either
widening `emitProgram` to a genuine multi-translation-unit emission strategy
(a significant architectural change with no current owner) or accepting the
inertness as permanent and re-scoping NAT-07's own claim. The named witness
probe does not exist yet; per D-14-24's stated hand-off, resolving a
`UNOWNED(...)` witness identifier to an executed probe is plan 14-07's job
(the plan that adds the register's `Witness` column). This row's own witness
token is chosen to name the eventual probe's intended subject so 14-07 can
wire it without renaming.

### D-14-46 — the distinctness corpus does not grow itself

first-recorded: M003

See 14-02-SUMMARY.md's own "Debt rows to register" item 1 for the full
finding. Recorded verbatim here as the Item text requires; the trigger is
external to this phase and to M003's own roadmap (no milestone currently
plans `if`/conditional surface syntax — PROJECT.md's M003 scope explicitly
excludes it). `UNOWNED(...)` is therefore the honest landing phase: the
witness token names the trigger condition itself rather than a probe, since
there is nothing to execute today that could go red — the corpus is
correct until the trigger fires, at which point a human plan-time decision
(not a test) determines the replacement fixtures.

### D-14-47 — `Cause.Span` content on parse-failure diagnostics widens the coordinate-shift hole

first-recorded: M003

See 14-02-SUMMARY.md's own "Debt rows to register" item 2 for the full
finding. Recorded verbatim here. The widening is bounded to broken programs
only (parse-successful diagnostics are untouched), and reconciling it against
Phase 13's own half-enforced coordinate-shift rule is left to whichever
future plan next widens `Cause` content on a parse-successful diagnostic
path — no such plan exists yet, so `UNOWNED(...)` is the honest landing
phase.

### D-14-48 — `09-VALIDATION.md:85` cites a dead pattern, shipped green

first-recorded: M003

Plan 14-09's grade derivation ladder (EVD-02) parses this row's evidence
cell, `go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v`,
resolves the package's real test names via the same static index plan
14-01 built, and finds none of them matches the pattern `LoanChainIndex`.
The row was shipped `✅ green`; the cap derives only `WIRED`. This is
**independently corroborated** by a second, structurally different
instrument: plan 14-01's groundedness lint (`verification_groundedness_test.go`)
already pins this exact `(file, line)` in its own 125-entry frontier as an
R2 dead pattern. Two instruments, built from different derivations,
agreeing on the same dead reference is exactly the kind of cross-checked
evidence D-14-07 anticipated finding at least once. Not rewritten in
place — the archived row's evidence and command text are untouched; only
this register records the finding.

**Landing phase:** `P14` — plan 14-10's Nyquist reconciliation already owns
driving the groundedness lint's R1/R2 frontier toward empty, and this row
is a member of that same frontier.

### D-14-49 — `09-VALIDATION.md:93` cites a second dead pattern, shipped green

first-recorded: M003

Same finding shape as D-14-48, a different row: `go test ./internal/compiler/corevalidate -run 'Mode.*Invalid\|DecodeMode' -v`
matches zero real test names under the unescaped-pipe reading D-14-17
ratified (the reading this whole corpus uses). Shipped `✅ green`; derives
`WIRED`. Also already a member of plan 14-01's pinned R2 frontier —
independently corroborated the same way as D-14-48.

**Landing phase:** `P14`, same reconciliation plan as D-14-48.

### D-14-50 — `12-VALIDATION.md:52` names no exact test, shipped green

first-recorded: M003

Row 12-04-01's evidence cell, `go test ./internal/compiler/originvalidate/... ./internal/compiler/corevalidate/... -count=1`,
names two packages and no `-run`/`-list`/`-fuzz`/`-bench` pattern. D-14-03's
ladder requires an exact resolving test identifier to reach `EXERCISED`; a
bare package invocation names none, however genuinely it compiles and
runs. Shipped `✅ green`; derives `WIRED`. This is a shape gap in the row's
own phrasing, not evidence that the underlying code is broken — no phase
currently owns rephrasing the row with a specific `-run` pattern.

**Landing phase:** `UNOWNED(probe:TestValidationGradeCapBarePackageRowHasNoNamedTest)`.

### D-14-51 — `04-VALIDATION.md:74` cites two tests retired in a later phase

first-recorded: M003

Row 04-06-03's evidence cell names `TestLastUseDiscoveryWorkIsCounted` and
`TestLastUseDiscoveryWorkSeries`. `grep -rn "func TestLastUseDiscoveryWork"
internal/compiler/check/*.go` returns zero matches — neither exists in the
current module. Both were real when `04-06-SUMMARY.md` shipped this row
green (its own coverage section cites both by file and line), and
`04-DEBT.md` records "the deferred retirement" of the transitive scan they
covered. Grepping the M002 phase history: the retirement this deferred item
named landed alongside `computeLoanLastUses`'s deletion (D-09-09,
`09-VALIDATION.md` row `09-09-01`), which is where these two tests'
coverage was folded into the surviving single loan-liveness law without
this M001 row being repointed. Shipped `✅ green`; derives `WIRED`.
**Not rewritten in place** — this is the real hole in shipped M001
evidence D-13-33's precedent and D-14-07 both anticipated surfacing, not a
regression to suppress.

**Landing phase:** `UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent)`
— no phase currently owns repointing this M001 row at its evidence's
current equivalent (if one still exists in the surviving law's own test
suite) or retracting the claim.

### D-14-52 — `06-VALIDATION.md:69` cites a test renamed in plan 08-05

first-recorded: M003

Row 06-08-T3's evidence cell names `TestOnlyRecomputedWorkIsGateEligible`.
It does not exist in `internal/compiler/measure/statistics_test.go` today;
`08-05-SUMMARY.md`'s own key-decisions record that plan 08-05 renamed it to
`TestOnlyGateEligibleMetricsPassThrough` when `GateEligibleMetrics()`
widened from a single metric to a two-element set, because "only
recomputed_work" stopped being true. This M001 row, shipped `✅ green`
before that rename, was never repointed at the new name. Derives `WIRED`.
Same finding class as D-14-51 — a real, documented rename left a stale
citation behind in a document already marked satisfied.

**Landing phase:** `UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent)`
— the same shared probe witnesses both D-14-51 and D-14-52, since both
assert the same property (a named set of historically-real identifiers
remains absent today).

### D-14-53 — `14-VALIDATION.md`'s own verification map was never filled in

first-recorded: M003

`14-VALIDATION.md`'s Per-Task Verification Map has exactly one row
(`14-01-01`), and its Automated Command cell is still the literal
plan-time placeholder `` `{command}` `` — no plan across this phase's nine
plans ever replaced it with real evidence, because each plan tracks its
own verification through its own `*-SUMMARY.md` (this plan's included)
rather than through this shared table. Grading it honestly derives
`DEFINED` (there is no evidence cell to derive a ceiling from beyond the
floor), which would fail EVD-02's satisfying bar if the bar applied here.
Rather than let the bar's own file-scoped exemption map
(`validationGradeBarExemptions`) quietly cover this forever — exactly the
suppression EVD-02 exists to retire — this row names the gap explicitly.

**Update (plan 14-10):** the table is now populated for real — 31 rows,
one per task across all ten plans in this phase, every `Automated Command`
executed during plan 14-10's Task 3 and confirmed to resolve (see
14-10-SUMMARY.md). The `validationGradeBarExemptions` entry for
`14-VALIDATION.md` was deliberately LEFT IN PLACE rather than removed:
removing it and re-running `TestValidationRowGradesAreEarnedOverArchivedCorpus`
surfaced spurious "declared EXERCISED exceeds the ceiling WIRED" failures
across MULTIPLE unrelated already-frozen archived files (01, 02, 06, 07,
08, 10, 11, 12, 13-VALIDATION.md), not just this document — consistent
with a run-record generation that did not complete within its budget on
that attempt (the whole run finished at 300.11s, suspiciously close to
`evidenceRunRecordTimeout`'s 300s ceiling per plan 14-09's own SUMMARY).
Rather than land a corpus-wide grade change under that risk, the change
was reverted; see D-14-121 below.

**Landing phase:** `UNOWNED(none-yet-scheduled)` — no phase yet owns either
retroactively populating this table from the phase's own plan SUMMARYs or
formally retiring the document in favor of the convention the phase has,
in practice, already moved to.

### D-14-54 — `06-VALIDATION.md:71` cites a test renamed per its own reviewer's warning

first-recorded: M003

Row 06-09-T3's evidence cell names `TestRecomputedWorkIsTheOnlyHardGate`.
`grep -rn "func TestRecomputedWork" internal/compiler/session/*.go` shows
only `TestRecomputedWorkHardGateBoundComparison` and
`TestRecomputedWorkBlockingIgnoresHighCoV` today — the first is the
rename. `08-REVIEW.md`'s WR-01 warning is the documented cause: it named
this exact test as stale the moment plan 08-05 widened
`QLT02GateEligibleMetrics()`/`measure.GateEligibleMetrics()` from one
metric to a two-element set (the name's own claim, "only hard gate",
stopped being literally true), and recommended a rename along
`TestRecomputedWorkCeilingIsStrictlyEnforced` lines; the rename that
actually landed carries a different but equivalent name. This M001 row,
shipped `✅ green` before the review and the rename, was never repointed.
Derives `WIRED`. Third instance of the same finding class as D-14-51 and
D-14-52 in this same register — not a coincidence so much as a recurring
failure mode this whole plan exists to make visible: a test rename is a
routine, reviewed, well-documented refactor, and the citing archived row
is exactly what nothing currently re-checks.

**Landing phase:** `UNOWNED(probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent)`
— the same shared probe now witnesses all three (D-14-51, D-14-52, D-14-54).

### D-14-55 — `03-RESEARCH.md:849` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 849
command: go test ./internal/compiler/check/... -run TestLoanConflict
classification: R2
verdict: superseded
superseding-phase: P03
superseding-commit: a428280
covering-command: go test ./internal/compiler/check/... -run TestLoanLivenessFixpoint
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `a428280` (phase P03) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-56 — `03-RESEARCH.md:850` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 850
command: go test ./internal/compiler/check/... -run TestCFGLastUse
classification: R2
verdict: superseded
superseding-phase: P08
superseding-commit: 63246d9
covering-command: go test ./internal/compiler/check/... -run TestCFGBackEdgeWalkIsIterative
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `63246d9` (phase P08) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-57 — `03-RESEARCH.md:852` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 852
command: go test ./internal/compiler/corevalidate/... -run TestPublicOrigin
classification: R2
verdict: superseded
superseding-phase: P03
superseding-commit: 2b2649e
covering-command: go test ./internal/compiler/check/... -run TestPublicOriginFactLowered
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2b2649e` (phase P03) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-58 — `04-RESEARCH.md:560` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 560
command: go test ./internal/compiler/session/... -run TestResourceReleaseOrder
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 300ab6b
covering-command: go test ./internal/compiler/session/... -run TestReleaseTranspositionMutationIsMismatch
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `300ab6b` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-59 — `04-RESEARCH.md:561` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 561
command: go test ./internal/compiler/cgen/... -run TestForeignLayoutConformance
classification: R2
verdict: superseded
superseding-phase: P05
superseding-commit: 8a3d34b
covering-command: go test ./internal/compiler/cgen/... -run TestForeignManifestBytesUnchangedForPriorPhases
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `8a3d34b` (phase P05) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-60 — `04-RESEARCH.md:563` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 563
command: go test ./internal/compiler/syntax/... -run TestFallibleOperationConsumers
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 62a443d
covering-command: go test ./internal/compiler/syntax/... -run TestFallibleCallUnconsumedRejected
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `62a443d` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-61 — `04-RESEARCH.md:565` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 565
command: go test ./internal/compiler/interp/... -run TestNoCancelledOutcomeEmitted
classification: R2
verdict: superseded
superseding-phase: P10
superseding-commit: 2f236ef
covering-command: go test ./internal/compiler/interp/... -run TestRunRefusesInvalidBodyUnion
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2f236ef` (phase P10) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-62 — `04-RESEARCH.md:566` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 566
command: go test ./internal/compiler/native/... -run TestNoUnauditedUndefinedSymbols
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: ae83e3b
covering-command: go test ./internal/compiler/native/... -run TestUndefinedSymbolAllowlistRejectsNewSymbol
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `ae83e3b` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-63 — `04-RESEARCH.md:567` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 567
command: go test ./internal/compiler/native/... -run TestNonlocalExitDetected
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 8743bfe
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `8743bfe` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-64 — `04-RESEARCH.md:568` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 568
command: go test ./internal/compiler/evidence/... -run TestInterpNativeAgreementIncludingDefect
classification: R2
verdict: superseded
superseding-phase: P11
superseding-commit: 9448de6
covering-command: go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `9448de6` (phase P11) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-65 — `04-RESEARCH.md:569` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 569
command: go test ./internal/compiler/originvalidate/... -run TestWalksAllTerminators
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 1103fc1
covering-command: go test ./internal/compiler/originvalidate/... -run TestOriginWalksEveryTerminator
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `1103fc1` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-66 — `04-RESEARCH.md:569` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 569
command: go test ./internal/compiler/pathoracle/... -run TestWalksAllTerminators
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: c9677e2
covering-command: go test ./internal/compiler/pathoracle/... -run TestPathOracleClosesOnEveryTerminator
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `c9677e2` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-67 — `04-RESEARCH.md:570` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 570
command: go test ./internal/compiler/native/... -run TestAbortSignalHandling
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 8743bfe
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `8743bfe` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-68 — `08-RESEARCH.md:507` cites a symbol deleted by design

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-RESEARCH.md
line: 507
command: go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree
classification: R2
verdict: obsolete-by-design
deleted-package: internal/compiler/check
deleted-symbol: computeLoanLastUses
deleting-phase: P09-09
deleting-commit: b8fe3df
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `b8fe3df` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-69 — `08-RESEARCH.md:514` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-RESEARCH.md
line: 514
command: go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestBudgetLaneCarriesMachineIDAndVerdict
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-70 — `08-VALIDATION.md:53` cites a symbol deleted by design

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md
line: 53
command: go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree
classification: R2
verdict: obsolete-by-design
deleted-package: internal/compiler/check
deleted-symbol: computeLoanLastUses
deleting-phase: P09-09
deleting-commit: b8fe3df
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `b8fe3df` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-71 — `08-VALIDATION.md:65` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/08-interprocedural-loan-liveness-in-check/08-VALIDATION.md
line: 65
command: go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestBudgetLaneCarriesMachineIDAndVerdict
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-72 — `09-RESEARCH.md:706` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-RESEARCH.md
line: 706
command: go test ./internal/compiler/corevalidate -run TestBuildLoanChainIndex
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: 703fedc
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `703fedc` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-73 — `09-RESEARCH.md:708` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-RESEARCH.md
line: 708
command: go test ./internal/compiler/corevalidate -run TestPeerDoesNotRederiveNarrowedClasses
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/corevalidate -run TestPeerRederivesFormerlyNarrowedClasses
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-74 — `09-VALIDATION.md:87` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 87
command: go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: 703fedc
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `703fedc` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-75 — `09-VALIDATION.md:95` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 95
command: go test ./internal/compiler/corevalidate -run 'Mode.*Invalid|DecodeMode' -v
classification: R2
verdict: superseded
superseding-phase: P07
superseding-commit: b85d518
covering-command: go test ./internal/compiler/corevalidate -run TestDerivePeerSignatureModeMutantPairing
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `b85d518` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-76 — `09-VALIDATION.md:114` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 114
command: grep -c "OWN-05a" .planning/REQUIREMENTS.md
classification: R3
verdict: superseded
superseding-phase: P09
superseding-commit: 2e085a4
covering-command: grep -c "OWN-05a" .planning/MILESTONES.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2e085a4` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-77 — `09-VALIDATION.md:115` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 115
command: grep -c "S-008" .planning/ROADMAP.md
classification: R3
verdict: superseded
superseding-phase: P09
superseding-commit: 2598ef5
covering-command: grep -c "S-008" .planning/MILESTONES.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2598ef5` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-78 — `10-RESEARCH.md:684` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/10-trusted-interprocedural-oracle/10-RESEARCH.md
line: 684
command: go test ./internal/compiler/session/... -run TestCheckCommandFile
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestCheckCommandSurfacesPeerRefusal
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-79 — `11-RESEARCH.md:1041` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-RESEARCH.md
line: 1041
command: go test ./internal/compiler/<package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-80 — `11-VALIDATION.md:28` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 28
command: go test ./internal/compiler/<touched-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-81 — `11-VALIDATION.md:53` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 53
command: grep -c -E 'D-11-(02|07|11|12|13|27|36|40|42)' …/PHASE-11-DEBT.md
classification: R1
verdict: renamed
replacement: grep -c -E 'D-11-(02|07|11|12|13|27|36|40|42)' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-82 — `11-VALIDATION.md:58` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 58
command: go test ./internal/compiler/callgraph/... -run 'TestEntryFunction…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/callgraph/... -run 'TestEntryFunctionRefusesManyRoots|TestEntryFunctionRefusesZeroRoots' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-83 — `11-VALIDATION.md:60` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 60
command: go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSet…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSetIsExplicitlyEmpty' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-84 — `11-VALIDATION.md:62` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 62
command: grep -c -E 'function count|call-edge count|N =…' …/11-MIDPHASE-GATE.md
classification: R1
verdict: renamed
replacement: grep -c -E 'function count|call-edge count|N =' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-MIDPHASE-GATE.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-85 — `11-VALIDATION.md:63` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 63
command: awk … | wc -l
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/session/... -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-86 — `11-VALIDATION.md:65` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 65
command: grep -c 'D-11-25' …/session_phase11_differential_test.go
classification: R1
verdict: renamed
replacement: grep -c 'D-11-25' internal/compiler/session/session_phase11_differential_test.go
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-87 — `11-VALIDATION.md:66` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 66
command: go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-88 — `11-VALIDATION.md:70` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 70
command: go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames|TestCache…|TestNoClosureDigestInCache' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames|TestCacheDirectImportGuardCanFail|TestCacheTransitiveImportGuardCanFail|TestNoClosureDigestInCache' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-89 — `11-VALIDATION.md:71` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 71
command: grep -c -E 'QLT-06a|QLT-06b|strictly dominates…' …/11-QLT06-ABSTENTION.md
classification: R1
verdict: renamed
replacement: grep -c -E 'QLT-06a|QLT-06b|strictly dominates' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-QLT06-ABSTENTION.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-90 — `11-VALIDATION.md:73` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 73
command: go test ./internal/compiler/reduce/... -run 'TestDropCallSite|TestDropOrphanFunction|…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/reduce/... -run 'TestDropCallSiteRewritesOneEdge|TestDropOrphanFunctionRemovesUncalledNonEntry|TestDropUnusedBindingRemovesOnlyUnreadOperation' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-91 — `12-RESEARCH.md:869` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/12-result-payloads/12-RESEARCH.md
line: 869
command: go test ./internal/compiler/<package>/... -run <TestName>
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-92 — `12-VALIDATION.md:26` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md
line: 26
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-93 — `07-VERIFICATION.md:87` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VERIFICATION.md
line: 87
command: grep -nE 'TBD|FIXME|XXX'
classification: R1
verdict: superseded
superseding-phase: P14
superseding-commit: 93613cc
covering-command: go test ./internal/compiler/session/... -run TestVerificationGroundedness
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `93613cc` (phase P14) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-94 — `03-RESEARCH.md:853` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 853
command: go run ./cmd/lang -- <interface-export/import flow>
classification: R1
verdict: superseded
superseding-phase: P07
superseding-commit: 9b62c1a
covering-command: go test ./internal/compiler/corevalidate -run TestSummaryPeerCallableAgreesOnOriginOmittedClass
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `9b62c1a` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-95 — `04-RESEARCH.md:554` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 554
command: go test ./internal/compiler/... -run <Test...>
classification: R1
verdict: superseded
superseding-phase: P04
superseding-commit: 300ab6b
covering-command: go test ./internal/compiler/session/... -run TestReleaseTranspositionMutationIsMismatch
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `300ab6b` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-96 — `05-RESEARCH.md:420` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-RESEARCH.md
line: 420
command: go test ./internal/compiler/... -run <TestName>
classification: R1
verdict: superseded
superseding-phase: P04
superseding-commit: 8743bfe
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `8743bfe` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-97 — `14-01-SUMMARY.md:160` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 160
command: grep -nE 'TBD|FIXME|XXX'
classification: R1
verdict: superseded
superseding-phase: P14
superseding-commit: 93613cc
covering-command: go test ./internal/compiler/session/... -run TestVerificationGroundedness
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `93613cc` (phase P14) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-98 — `14-01-SUMMARY.md:162` cites a symbol deleted by design

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 162
command: go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree
classification: R2
verdict: obsolete-by-design
deleted-package: internal/compiler/check
deleted-symbol: computeLoanLastUses
deleting-phase: P09-09
deleting-commit: b8fe3df
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `b8fe3df` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-99 — `14-01-SUMMARY.md:163` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 163
command: go test ./internal/compiler/session/... -run TestAuditQLT02BudgetManifest
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestBudgetLaneCarriesMachineIDAndVerdict
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-100 — `14-01-SUMMARY.md:165` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 165
command: go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: 703fedc
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `703fedc` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-101 — `14-01-SUMMARY.md:166` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 166
command: go test ./internal/compiler/corevalidate -run 'Mode.*Invalid|DecodeMode' -v
classification: R2
verdict: superseded
superseding-phase: P07
superseding-commit: b85d518
covering-command: go test ./internal/compiler/corevalidate -run TestDerivePeerSignatureModeMutantPairing
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `b85d518` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-102 — `14-01-SUMMARY.md:169` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 169
command: go test ./internal/compiler/<touched-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-103 — `14-01-SUMMARY.md:170` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 170
command: grep -c -E 'D-11-(02|07|11|12|13|27|36|40|42)' …/PHASE-11-DEBT.md
classification: R1
verdict: renamed
replacement: grep -c -E 'D-11-(02|07|11|12|13|27|36|40|42)' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-104 — `14-01-SUMMARY.md:171` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 171
command: go test ./internal/compiler/callgraph/... -run 'TestEntryFunction…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/callgraph/... -run 'TestEntryFunctionRefusesManyRoots|TestEntryFunctionRefusesZeroRoots' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-105 — `14-01-SUMMARY.md:172` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 172
command: go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSet…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/cgen/... -run 'TestEmittedAttributeSetIsExplicitlyEmpty' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-106 — `14-01-SUMMARY.md:173` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 173
command: grep -c -E 'function count|call-edge count|N =…' …/11-MIDPHASE-GATE.md
classification: R1
verdict: renamed
replacement: grep -c -E 'function count|call-edge count|N =' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-MIDPHASE-GATE.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-107 — `14-01-SUMMARY.md:174` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 174
command: awk … | wc -l
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/session/... -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-108 — `14-01-SUMMARY.md:175` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 175
command: grep -c 'D-11-25' …/session_phase11_differential_test.go
classification: R1
verdict: renamed
replacement: grep -c 'D-11-25' internal/compiler/session/session_phase11_differential_test.go
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-109 — `14-01-SUMMARY.md:176` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 176
command: go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/session/... -run 'TestQLT03GeneratorOpKindClosure' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-110 — `14-01-SUMMARY.md:177` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 177
command: go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames|TestCache…|TestNoClosureDigestInCache' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/cache/... -run 'TestDeclaredInputNames|TestCacheDirectImportGuardCanFail|TestCacheTransitiveImportGuardCanFail|TestNoClosureDigestInCache' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-111 — `14-01-SUMMARY.md:178` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 178
command: grep -c -E 'QLT-06a|QLT-06b|strictly dominates…' …/11-QLT06-ABSTENTION.md
classification: R1
verdict: renamed
replacement: grep -c -E 'QLT-06a|QLT-06b|strictly dominates' .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-QLT06-ABSTENTION.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-112 — `14-01-SUMMARY.md:179` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 179
command: go test ./internal/compiler/reduce/... -run 'TestDropCallSite|TestDropOrphanFunction|…' -v -count=1
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/reduce/... -run 'TestDropCallSiteRewritesOneEdge|TestDropOrphanFunctionRemovesUncalledNonEntry|TestDropUnusedBindingRemovesOnlyUnreadOperation' -v -count=1
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-113 — `14-01-SUMMARY.md:181` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 181
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-114 — `14-01-SUMMARY.md:183` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 183
command: go test ./<changed-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-115 — `14-RESEARCH.md:448` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md
line: 448
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-116 — `14-RESEARCH.md:459` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md
line: 459
command: go test ./internal/compiler/session/... -run TestLanguageMaturityGuardCountIsCurrent -v
classification: R2
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestLanguageMaturityCountsAreCurrent -v
```

Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

### D-14-117 — `14-VALIDATION.md:26` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
line: 26
command: go test ./<changed-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 5050f57
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5050f57` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-118 — `ADVERSARIAL-SYNTHESIS.md:213` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/research/M003/ADVERSARIAL-SYNTHESIS.md
line: 213
command: grep 'D-11-25' PHASE-11-DEBT.md
classification: R3
verdict: superseded
superseding-phase: P14-07
superseding-commit: b8242d1
covering-command: grep -c "D-11-25" .planning/UNREACHABLE-CLAIMS.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `b8242d1` (phase P14-07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-119 — `ADVERSARIAL-SYNTHESIS.md:214` cites a symbol deleted by design

first-recorded: M003

```reconciliation
file: .planning/research/M003/ADVERSARIAL-SYNTHESIS.md
line: 214
command: go test ./internal/compiler/check/... -list 'TestComputeLoanLastUsesAndDerivePlaceLoansAgree'
classification: R2
verdict: obsolete-by-design
deleted-package: internal/compiler/check
deleted-symbol: computeLoanLastUses
deleting-phase: P09-09
deleting-commit: b8fe3df
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `b8fe3df` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-120 — `ADVERSARIAL-SYNTHESIS.md:214` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/research/M003/ADVERSARIAL-SYNTHESIS.md
line: 214
command: go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -count=1
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: 703fedc
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `703fedc` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-121 — the corpus-wide grade-cap satisfying bar was not re-verified for `14-VALIDATION.md`

first-recorded: M003

Plan 14-10's Task 3 filled `14-VALIDATION.md`'s Per-Task Verification Map
for real (31 rows, one per task across all ten plans, every command
executed and confirmed to resolve). This made
`validationGradeBarExemptions`'s `14-VALIDATION.md` entry — whose stated
reason was "template row never filled beyond plan-time placeholder" — read
as stale, so removing it and re-running
`TestValidationRowGradesAreEarnedOverArchivedCorpus` (the corpus-wide
EVD-02 satisfying-bar test) was attempted.

That attempt surfaced a problem orthogonal to `14-VALIDATION.md` itself:
the SAME run reported "declared grade EXERCISED exceeds the ceiling WIRED"
for rows in NINE other, unrelated, already-frozen archived files
(`01-VALIDATION.md`, `02-VALIDATION.md`, `06-VALIDATION.md`,
`07-VALIDATION.md`, `08-VALIDATION.md`, `10-VALIDATION.md`,
`11-VALIDATION.md`, `12-VALIDATION.md`, `13-VALIDATION.md`) whose own
exemptions were never touched. The whole corpus-wide test completed at
`300.11s` — suspiciously close to `evidenceRunRecordTimeout`'s `300s`
ceiling (`internal/compiler/session/evidence_grade_test.go`, lowered from
`900s` by plan 14-09's own deviation fix per its SUMMARY) — consistent
with the run-record generation not completing within its budget on that
attempt and several packages' citations falling back to an unresolved
`WIRED` ceiling instead of their true `EXERCISED` grade, rather than any
of those nine files having genuinely regressed.

The exemption-removal change was reverted (`git checkout --`) rather than
landed under that risk: this plan never confirmed, one way or the other,
whether `14-VALIDATION.md` would actually clear the satisfying bar under a
complete run record. `14-VALIDATION.md`'s exemption therefore stays in
`validationGradeBarExemptions`, now for a genuinely different and equally
honest reason than the one originally written there (the table is filled,
but the bar was never safely re-verified) — recorded here rather than
silently left to read as though the original "never filled in" reason
still applied.

**Landing phase:** `UNOWNED(none-yet-scheduled)` — no phase yet owns either
re-running the corpus-wide grade cap with a verified-complete run record
(possibly requiring `evidence-run-record.sh`'s own timeout/consolidation
budget to be revisited a second time, mirroring plan 14-09's own
`consolidatePkgPatterns` fix) or accepting `14-VALIDATION.md`'s exemption
as permanent with an updated, non-stale rationale.

### D-14-122 — `14-VERIFICATION.md:133` cites a seeded-fault placeholder name, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/phases/14-evidence-instrument-and-honest-scoping/14-VERIFICATION.md
line: 133
command: go test ... -run 'TestThisNameDoesNotExistAnywhereZZQQ'
classification: R1
verdict: renamed
replacement: go test ./internal/compiler/session/... -run TestVerificationGroundednessFrontierIsPinned
```

Plan 14-11's fail-closed margin check (Task 2) surfaced this while measuring
the corpus-wide run record's real cost: with D-14-121's own risk closed
(the run now finishes in well under its margin, see this plan's own
SUMMARY), `TestValidationRowGradesAreEarnedOverArchivedCorpus` derived a
genuine `WIRED` ceiling for `14-VALIDATION.md` row `14-10-T1` -- not a
false negative from an incomplete run, but a real, previously-unreported
gap in `14-VERIFICATION.md`'s own Non-Inertness Spot-Check table (row 1),
added after plan 14-10's reconciliation pass closed the R1/R2/R3 frontier
and therefore never reconciled.

The archived cell's literal text, `go test ... -run
'TestThisNameDoesNotExistAnywhereZZQQ'`, narrates a fault-injection
perturbation that was already executed and reverted: the same table row's
own "Reverted clean?" column confirms `git checkout --` restored
`14-01-SUMMARY.md` to clean, and `TestThisNameDoesNotExistAnywhereZZQQ` was
never a real test -- it was deliberately fabricated so the row could show
`TestVerificationGroundednessFrontierIsPinned` and
`TestVerificationGroundednessThreeClassesAreEmpty` catching it. The cell
therefore never named live evidence and was never meant to resolve; the
row's actual claim rests on the replacement command above, which is a
real, currently-resolving test (verified via `classifyCommand` over the
static test index, the same primitive the lint's own classification uses).
The archived row is left byte-unmodified; only this register records the
correction.

**Landing phase:** `P14` -- closed by this same plan (14-11) recording the
correction; no further work is scheduled.

---

*Register: PHASE-14-DEBT.md*
*Opened: 2026-09-18 (plan 14-04)*
