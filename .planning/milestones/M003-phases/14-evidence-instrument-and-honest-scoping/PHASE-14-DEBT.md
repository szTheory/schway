---
phase: 14-evidence-instrument-and-honest-scoping
recorded: 2026-09-18
status: accepted
disposition: phase-in-progress
items: 100
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
| D-14-45 | PROJECT.md `## Current State` (verified already correct, commit `532ed0b`); 14-RESEARCH.md Open Question 1 | EVD-07 | warning | P21 | EXERCISED | probe:TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison | `emitProgram` emits one translation unit without `restrict` and continues to refuse foreign/by-pointer program shapes; these structural facts define scope, not optimizer behavior. The opt-in `TestPhase21EmittedMultiFunctionLTOComparison` and `21-LTO-EVIDENCE.md` record semantic equality across interpreter, `-O0`, `-O3`, and `-O3 -flto` for `testdata/phase14/multi_function_match_refusal.lang` only, using Apple Clang 21.0.0 on Darwin arm64 and the recorded fixture/generated-C digests. D-11-25's structural rationale is distinct from this measured observation. The recurring structural probe binds that tagged test to its receipt without rerunning the compiler matrix. This does not measure optimizer activity or performance, prove resource cleanup, or establish any other host/toolchain result. The existing independent comparator negative control remains `TestPhase16EmitterPortSemanticGuardIsNotInert`. |
| D-14-46 | 14-02-SUMMARY.md "Debt rows to register" item 1 | DX-08 | info | UNOWNED(conditional-surface-lands-in-language) | DEFINED | n/a | NOTHING FORCES THE DISTINCTNESS CORPUS TO GROW AS THE LANGUAGE GROWS. When the conditional form (`if`) enters the language, the spiral trio (`spiral_full.lang`, `spiral_narrow.lang`, `spiral_bare.lang`) becomes a VALID program and must be replaced in `testdata/distinctness/` by the then-current not-in-language surface, or the corpus predicate would need to be re-derived against a program that no longer refuses to parse |
| D-14-47 | 14-02-SUMMARY.md "Debt rows to register" item 2 | DX-08 | info | UNOWNED(identity-bearing-cause-span-on-parse-success-path) | DEFINED | n/a | PUTTING MORE CONTENT ON `Cause.Span` DEEPENS THE ALREADY-RECORDED, HALF-ENFORCED RULE THAT A COORDINATE SHIFT MUST NEVER MOVE A DIAGNOSTIC'S ID. This is now bounded-widened for broken programs only -- a whitespace edit inside a broken program's tail (past the point where declaration recovery starts) can now move that program's diagnostic ID, where before plan 14-02's fix the swallowed region was never in the identity basis. This is an acceptable, deliberately bounded widening of an already-recorded hole (Phase 13's coordinate-shift discipline), not a new defect, but it should be reconciled explicitly rather than rediscovered |
| D-14-48 | plan 14-09's grade derivation over 09-VALIDATION.md:85 | EVD-02 | warning | P14 | WIRED | probe:TestVerificationGroundednessFrontierIsPinned | 09-VALIDATION.MD ROW 09-01-01's SECOND EVIDENCE CELL (A `LoanChainIndex`-PATTERN GO TEST INVOCATION) CITES A PATTERN THAT MATCHES ZERO TEST NAMES -- IT WAS SHIPPED `✅ GREEN` BUT DERIVES ONLY WIRED. This is the same row plan 14-01's groundedness lint independently pinned as an R2 dead pattern in its own 125-entry frontier; the two instruments agree from two different mechanisms. Plan 14-10's Nyquist reconciliation is the scheduled closer for the groundedness frontier's R1/R2 class, which this row is a member of |
| D-14-49 | plan 14-09's grade derivation over 09-VALIDATION.md:93 | EVD-02 | warning | P14 | WIRED | probe:TestVerificationGroundednessFrontierIsPinned | 09-VALIDATION.MD ROW 09-05-02's `Mode.*Invalid`-ALTERNATION EVIDENCE CELL ALSO MATCHES ZERO TEST NAMES -- SHIPPED `✅ GREEN`, DERIVES ONLY WIRED. Same cross-instrument agreement as D-14-48: also a member of plan 14-01's pinned R2 frontier, closure scheduled at plan 14-10 |
| D-14-50 | plan 14-09's grade derivation over 12-VALIDATION.md:52 (row 12-04-01) | EVD-02 | info | CLOSED(b8cf216) | EXERCISED | probe:TestValidationGradeCapBarePackageRowHasNoNamedTest | 12-VALIDATION.MD ROW 12-04-01's EVIDENCE CELL NAMES TWO PACKAGES BUT NO `-run`/`-list`/`-fuzz`/`-bench` PATTERN, SO IT NAMES NO EXACT TEST IDENTIFIER THE CAP CAN CONFIRM EXECUTED -- repaired and directly exercised by commit b8cf216; probe:TestValidationGradeCapBarePackageRowHasNoNamedTest. |
| D-14-51 | plan 14-09's grade derivation over 04-VALIDATION.md:74 (row 04-06-03) | EVD-02 | warning | CLOSED(b8cf216) | EXERCISED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 04-VALIDATION.MD ROW 04-06-03's EVIDENCE CELL NAMES `TestLastUseDiscoveryWorkIsCounted` AND `TestLastUseDiscoveryWorkSeries`, NEITHER OF WHICH EXISTS IN THE CURRENT MODULE -- repaired and directly exercised by commit b8cf216; probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent. |
| D-14-52 | plan 14-09's grade derivation over 06-VALIDATION.md:69 (row 06-08-T3) | EVD-02 | warning | CLOSED(b8cf216) | EXERCISED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 06-VALIDATION.MD ROW 06-08-T3's EVIDENCE CELL NAMES `TestOnlyRecomputedWorkIsGateEligible`, WHICH DOES NOT EXIST -- repaired and directly exercised by commit b8cf216; probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent. |
| D-14-53 | plan 14-09's own satisfying-bar exemption for 14-VALIDATION.md | EVD-02 | info | CLOSED(172a893) | WIRED | probe:TestValidationGradeBarAppliesToPhase14 | 14-VALIDATION.MD'S OWN PER-TASK VERIFICATION MAP, ORIGINALLY EXACTLY ONE UNFILLED PLACEHOLDER ROW, IS NOW POPULATED FOR REAL: 31 rows plan 14-10 filled plus 8 this plan (14-12) added (14-11-T1..T3, 14-13-T1..T2, 14-12-T1..T3), all executed and confirmed to resolve, all judged by the now-unexempted `>=EXERCISED` satisfying bar via `TestValidationRowGradesAreEarnedOverArchivedCorpus`. The premise this row recorded (the table was never filled in beyond its plan-time placeholder) no longer holds |
| D-14-54 | plan 14-09's grade derivation over 06-VALIDATION.md:71 (row 06-09-T3) | EVD-02 | warning | CLOSED(b8cf216) | EXERCISED | probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent | 06-VALIDATION.MD ROW 06-09-T3's EVIDENCE CELL NAMES `TestRecomputedWorkIsTheOnlyHardGate`, WHICH DOES NOT EXIST -- repaired and directly exercised by commit b8cf216; probe:TestValidationGradeCapArchivedDeadCitationsRemainAbsent. |
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
| D-14-70 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-71 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-72 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-RESEARCH.md:706` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-73 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-RESEARCH.md:708` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-74 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-VALIDATION.md:87` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-75 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `09-VALIDATION.md:95` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-76 | plan 14-10's reconciliation of the pinned groundedness frontier's R3 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `09-VALIDATION.md:114` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-77 | plan 14-10's reconciliation of the pinned groundedness frontier's R3 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `09-VALIDATION.md:115` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-78 | plan 14-10's reconciliation of the pinned groundedness frontier's R2 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `10-RESEARCH.md:684` cites a dead/elided verification command, corrected outside the archive as a RENAMED verdict (see Detail section for the full command text and its replacement, which resolves). |
| D-14-79 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-RESEARCH.md:1041` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-80 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R1) — `11-VALIDATION.md:28` cites a verification command whose target has moved, corrected outside the archive as a SUPERSEDED verdict naming the superseding phase, commit and a live covering command (see Detail section). |
| D-14-81 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-82 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-83 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-84 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-85 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-86 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-87 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-88 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-89 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
| D-14-90 | plan 14-10's reconciliation of the pinned groundedness frontier's R1 class | EVD-03 | info | CLOSED(5c210a5) | DEFINED | n/a | CLOSED: Phase 02 corrected the archived command in commit 5c210a5; this finding no longer appears in the live scan. |
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
| D-14-121 | plan 14-10's attempt to remove 14-VALIDATION.md's satisfying-bar exemption | EVD-02 | warning | CLOSED(172a893) | WIRED | probe:TestValidationGradeBarAppliesToPhase14 | RESOLVED. `14-VALIDATION.MD` IS REMOVED FROM `validationGradeBarExemptions` (plan 14-12), AND `TestValidationRowGradesAreEarnedOverArchivedCorpus` PASSES FOR IT with a run record that reported complete and a measured elapsed of ~90-92s against the 480s budget (roughly 19%, comfortably inside the 75% margin) -- see Detail section for the full root-cause fix and measured numbers |
| D-14-122 | plan 14-11's fail-closed run-record margin check, which surfaced this while measuring the corpus-wide batch's real cost | EVD-01 | info | P14 | DEFINED | n/a | RETIRED — the text this row reconciled (`14-VERIFICATION.md:133`'s seeded-fault placeholder `TestThisNameDoesNotExistAnywhereZZQQ`) was deleted outright, not merely moved, when commit 5adcd32's verification-artifact rewrite replaced the whole Non-Inertness Spot-Check narrative; the reconciliation entry is removed as moot rather than repointed at a new anchor (see Detail section). |
| D-14-123 | plan 14-12's re-derivation of `14-VALIDATION.md` row `14-01-T2` under a complete run record | EVD-02 | info | P14 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `14-VALIDATION.MD` ROW `14-01-T2` DECLARES `WIRED`, BELOW THE NOW-ENFORCED `>=EXERCISED` SATISFYING BAR. Re-deriving its evidence cell against plan 14-11's completion-witnessed, margin-checked run record shows the cited test (`TestVerificationGroundednessFrontierIsPinned`) genuinely passed and is matched -- the row's TRUE ceiling is `EXERCISED`, not `WIRED`. The declared cell is kept byte-unchanged (this plan's own prohibition on rewriting a row belonging to plans 14-01..14-10 to force the bar to pass) and is instead narrowed here via `validationGradeBarRowExemptions`, mechanically checked by `TestValidationGradeBarRowExemptionsAreOwned`. |
| D-14-124 | plan 14-12's re-derivation of `14-VALIDATION.md` row `14-02-T1` under a complete run record | EVD-02 | info | P14 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `14-VALIDATION.MD` ROW `14-02-T1`'S EVIDENCE CELL IS A go-run invocation (`./cmd/lang --json check testdata/distinctness/spiral_full.lang`), not a go-test command -- IT IS STRUCTURALLY CAPPED AT `REACHABLE` BY `deriveCeiling`'s OWN LADDER AND CAN NEVER DERIVE `EXERCISED` WITHOUT REWRITING THE EVIDENCE CELL TO a go-test invocation with a -run pattern naming a specific test. The declared `REACHABLE` cell is kept byte-unchanged (row belongs to plan 14-02, covered by this plan's never-rewrite prohibition) and is narrowed here via `validationGradeBarRowExemptions`. |
| D-14-125 | plan 14-12's re-derivation of `14-VALIDATION.md` row `14-02-T2` under a complete run record | EVD-02 | info | P14 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `14-VALIDATION.MD` ROW `14-02-T2`'S EVIDENCE CELL NAMES THREE PACKAGES BUT NO `-run` PATTERN, SO `deriveCeiling` DERIVES `WIRED` STRUCTURALLY (THE SAME BARE-PACKAGE SHAPE GAP D-14-50 ALREADY RECORDED FOR `12-VALIDATION.md:52`) -- IT CAN NEVER REACH `EXERCISED` WITHOUT REWRITING THE CELL TO NAME A SPECIFIC TEST. Kept byte-unchanged (row belongs to plan 14-02) and narrowed here. |
| D-14-126 | plan 14-12's re-derivation of `14-VALIDATION.md` row `14-03-T1` under a complete run record | EVD-02 | info | P14 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `14-VALIDATION.MD` ROW `14-03-T1`'S EVIDENCE CELL, `go test ./cmd/lang-repair/... -count=1`, NAMES NO `-run` PATTERN -- SAME STRUCTURAL SHAPE GAP AS D-14-125, STRUCTURALLY CAPPED AT `WIRED`. Kept byte-unchanged (row belongs to plan 14-03) and narrowed here. |
| D-14-127 | plan 14-12's re-derivation of `14-VALIDATION.md` row `14-10-T3` under a complete run record | EVD-02 | info | P14 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `14-VALIDATION.MD` ROW `14-10-T3` DECLARES `WIRED`, BELOW THE NOW-ENFORCED BAR. Re-deriving its evidence cell (`TestVerificationGroundednessThreeClassesAreEmpty`) against the complete run record shows it genuinely passed and is matched -- TRUE ceiling `EXERCISED`, same finding shape as D-14-123. Kept byte-unchanged (row belongs to plan 14-10) and narrowed here. |
| D-14-128 | Phase 20 planned validation map row 20-08-01 exact-test ceiling | EVD-02 | info | P20 | WIRED | probe:TestValidationGradeBarRowExemptionsAreOwned | `20-VALIDATION.md` row `20-08-01` requires the unfiltered `go test ./... -count=1` preflight; the grader deliberately caps a bare package command with no `-run` at WIRED. Preserve the actual full-suite gate and narrow only this plan-time row until Plan 08 records its result. |
| D-14-129 | Phase 20 planned validation map row 20-08-02 script ceiling | EVD-02 | info | P20 | REACHABLE | probe:TestValidationGradeBarRowExemptionsAreOwned | `20-VALIDATION.md` row `20-08-02` invokes the bounded full-suite timing script, which the grade ladder classifies as REACHABLE rather than an exact named Go test. Preserve the actual measurement command and narrow only this plan-time row; Plan 08 owns its recorded timing outcome. |
| D-14-130 | plan 14-10's reconciliation of an archived Phase 11 groundedness finding | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R3) — `11-VALIDATION.md:56` cites a verification command whose roadmap target moved into the M002 archive; the historical command remains unchanged and the replacement is recorded in the reconciliation block (see Detail section). |
| D-14-131 | public identity migration: 13-RESEARCH.md:581 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-RESEARCH.md:581` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-132 | public identity migration: 13-RESEARCH.md:582 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-RESEARCH.md:582` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-133 | public identity migration: 13-VALIDATION.md:62 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:62` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-134 | public identity migration: 13-VALIDATION.md:63 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:63` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-135 | public identity migration: 13-VALIDATION.md:64 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:64` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-136 | public identity migration: 13-VALIDATION.md:66 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:66` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-137 | public identity migration: 13-VALIDATION.md:67 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:67` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-138 | public identity migration: 13-VALIDATION.md:68 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:68` preserves the recorded pre-public repair command and maps its package path to the current `schway-repair` command (see Detail section). |
| D-14-139 | public identity migration: 13-VALIDATION.md:84 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:84` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |
| D-14-140 | public identity migration: 13-VALIDATION.md:103 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `13-VALIDATION.md:103` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |
| D-14-141 | public identity migration: 14-RESEARCH.md:463 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-RESEARCH.md:463` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |
| D-14-142 | public identity migration: 14-VALIDATION.md:62 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-VALIDATION.md:62` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |
| D-14-143 | public identity migration: 14-VALIDATION.md:63 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `14-VALIDATION.md:63` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |
| D-14-144 | public identity migration: 23-VALIDATION.md:40 | EVD-03 | info | P14 | DEFINED | n/a | RECONCILIATION (R2) — `23-VALIDATION.md:40` preserves the recorded pre-public command and maps it to the current Schway package path (see Detail section). |

## Detail

### D-14-45 — measured multi-function LTO semantic comparison is scoped to one host and fixture

first-recorded: M003

**Historical Phase 14 context:** PROJECT.md's `## Current State` section had
already recorded a broad structural rationale before this row was added. That
rationale is not itself a compiler execution measurement. Phase 21 adds the
bounded executed comparison below; do not interpret it as optimizer-activity
or performance evidence. The original recorded text was:

> "**`-flto` is structurally inert on every cgen-emitted multi-function
> program** — one translation unit, no `restrict`, foreign refused. Both live
> non-inertness proofs are non-multi-function: `TestLTOTierIsNotInert` uses a
> single-function foreign fixture, and NAT-07's composition-only control is
> hand-written C explicitly not emitted by cgen (D-11-24). D-11-25 names this
> only in a test doc comment — it has **no debt row** and is **not** among
> the ten unowned items — while this document's NAT-07 bullet previously read
> as though the proof covered the interprocedural corpus. It does not."

This correction landed in commit `532ed0b` (2026-09-17, "docs: start
milestone M003"), before Phase 14 opened. EVD-07 therefore had exactly **one**
remaining task for this plan: add the missing debt row PROJECT.md's own text
already names as absent. No PROJECT.md rewrite was needed or performed.

**Debt-row absence independently confirmed** before this row was added:
`grep -rn "D-11-25" .planning/milestones/M002-phases/*/PHASE-*-DEBT.md`
returns zero matches. The only prior mentions of D-11-25 are a test doc
comment (`session_phase11_differential_test.go`) and this phase's own
planning documents (14-RESEARCH.md, 14-01-SUMMARY.md), neither of which is a
debt register.

**2026-09-26 — Phase 21 measured evidence.** The opt-in
`TestPhase21EmittedMultiFunctionLTOComparison` executed direct emitted output
for `testdata/phase14/multi_function_match_refusal.lang` on Darwin arm64 with
Apple Clang 21.0.0. The interpreter, `-O0`, `-O3`, and `-O3 -flto` documents
were semantically equal under the existing all-pairs comparator. Exact fixture
and emitted-C digests and flags are in
`.planning/milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md`.
The recurring `TestPhase21LTOEvidenceReceiptIsBoundToTaggedComparison` guard
keeps the tagged test and receipt discoverable without repeating the compiler
run. This evidence does not establish optimizer activity, performance, native
resource cleanup, or any other host/toolchain result. The existing structural
test `TestLTOInertnessOnMultiFunctionEmission` remains a distinct recurring
emission witness, not the executed LTO comparison.

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

Closed by commit `b8cf216`: the archived row now names its exact live tests, the retired citation control remains red on the old names, and the replacement command passed during Plan 20 execution. The complete 33-pair corpus run record at revision `0d58aac` corroborates this closure; its SHA-256 is `3c57e0937d0da9379578b0707c2a40428f36d336adc61549e07478c7a1ab1940` (manifest: `testdata/phase16/validation-corpus-run-record.manifest.json`).

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

Closed by commit `b8cf216`: the archived row now names its exact live tests, the retired citation control remains red on the old names, and the replacement command passed during Plan 20 execution. The complete 33-pair corpus run record at revision `0d58aac` corroborates this closure; its SHA-256 is `3c57e0937d0da9379578b0707c2a40428f36d336adc61549e07478c7a1ab1940` (manifest: `testdata/phase16/validation-corpus-run-record.manifest.json`).

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

Closed by commit `b8cf216`: the archived row now names its exact live tests, the retired citation control remains red on the old names, and the replacement command passed during Plan 20 execution. The complete 33-pair corpus run record at revision `0d58aac` corroborates this closure; its SHA-256 is `3c57e0937d0da9379578b0707c2a40428f36d336adc61549e07478c7a1ab1940` (manifest: `testdata/phase16/validation-corpus-run-record.manifest.json`).

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

**Update (plan 14-12) — RESOLVED.** Plan 14-11 fixed D-14-121's root
cause (raised `evidenceRunRecordTimeout` to `480s`, added the 0.75 margin
fraction, anchored the producer/consumer name contract), and plan 14-12
removed the `14-VALIDATION.md` exemption entry for real (commit
`172a893`), adding 8 more real rows on top of plan 14-10's 31
(`14-11-T1..T3`, `14-13-T1..T2`, `14-12-T1..T3` — 39 total,
`graded_rows: 39`). `TestValidationRowGradesAreEarnedOverArchivedCorpus`
now passes for this file with a complete run record. The premise this
row recorded — that the table would sit at a single unfilled placeholder
forever, with no plan ever closing that gap — no longer holds; the table
is filled, the exemption is gone, and a permanent guard
(`TestValidationGradeBarAppliesToPhase14`) makes both facts checkable
rather than merely asserted.

**Landing phase:** `CLOSED(172a893)` — plan 14-12's commit.

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

Closed by commit `b8cf216`: the archived row now names its exact live tests, the retired citation control remains red on the old names, and the replacement command passed during Plan 20 execution. The complete 33-pair corpus run record at revision `0d58aac` corroborates this closure; its SHA-256 is `3c57e0937d0da9379578b0707c2a40428f36d336adc61549e07478c7a1ab1940` (manifest: `testdata/phase16/validation-corpus-run-record.manifest.json`).

### D-14-55 — `03-RESEARCH.md:849` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 849
command: go test ./internal/compiler/check/... -run TestLoanConflict
classification: R2
verdict: superseded
superseding-phase: P03
superseding-commit: 64ce0a0
covering-command: go test ./internal/compiler/check/... -run TestLoanLivenessFixpoint
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `64ce0a0` (phase P03) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-56 — `03-RESEARCH.md:850` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 850
command: go test ./internal/compiler/check/... -run TestCFGLastUse
classification: R2
verdict: superseded
superseding-phase: P08
superseding-commit: 5917616
covering-command: go test ./internal/compiler/check/... -run TestCFGBackEdgeWalkIsIterative
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `5917616` (phase P08) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-57 — `03-RESEARCH.md:852` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 852
command: go test ./internal/compiler/corevalidate/... -run TestPublicOrigin
classification: R2
verdict: superseded
superseding-phase: P03
superseding-commit: 8e2133c
covering-command: go test ./internal/compiler/check/... -run TestPublicOriginFactLowered
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `8e2133c` (phase P03) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-58 — `04-RESEARCH.md:560` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 560
command: go test ./internal/compiler/session/... -run TestResourceReleaseOrder
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 17e9cb3
covering-command: go test ./internal/compiler/session/... -run TestReleaseTranspositionMutationIsMismatch
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `17e9cb3` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-59 — `04-RESEARCH.md:561` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 561
command: go test ./internal/compiler/cgen/... -run TestForeignLayoutConformance
classification: R2
verdict: superseded
superseding-phase: P05
superseding-commit: de7f810
covering-command: go test ./internal/compiler/cgen/... -run TestForeignManifestBytesUnchangedForPriorPhases
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `de7f810` (phase P05) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-60 — `04-RESEARCH.md:563` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 563
command: go test ./internal/compiler/syntax/... -run TestFallibleOperationConsumers
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 46b8832
covering-command: go test ./internal/compiler/syntax/... -run TestFallibleCallUnconsumedRejected
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `46b8832` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-61 — `04-RESEARCH.md:565` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 565
command: go test ./internal/compiler/interp/... -run TestNoCancelledOutcomeEmitted
classification: R2
verdict: superseded
superseding-phase: P10
superseding-commit: 38aa24a
covering-command: go test ./internal/compiler/interp/... -run TestRunRefusesInvalidBodyUnion
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `38aa24a` (phase P10) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-62 — `04-RESEARCH.md:566` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 566
command: go test ./internal/compiler/native/... -run TestNoUnauditedUndefinedSymbols
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 326e311
covering-command: go test ./internal/compiler/native/... -run TestUndefinedSymbolAllowlistRejectsNewSymbol
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `326e311` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-63 — `04-RESEARCH.md:567` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 567
command: go test ./internal/compiler/native/... -run TestNonlocalExitDetected
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: f1b3bad
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f1b3bad` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-64 — `04-RESEARCH.md:568` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 568
command: go test ./internal/compiler/evidence/... -run TestInterpNativeAgreementIncludingDefect
classification: R2
verdict: superseded
superseding-phase: P11
superseding-commit: fe01d93
covering-command: go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `fe01d93` (phase P11) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-65 — `04-RESEARCH.md:569` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 569
command: go test ./internal/compiler/originvalidate/... -run TestWalksAllTerminators
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 77b3ef2
covering-command: go test ./internal/compiler/originvalidate/... -run TestOriginWalksEveryTerminator
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `77b3ef2` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-66 — `04-RESEARCH.md:569` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 569
command: go test ./internal/compiler/pathoracle/... -run TestWalksAllTerminators
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: 1ecdf87
covering-command: go test ./internal/compiler/pathoracle/... -run TestPathOracleClosesOnEveryTerminator
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `1ecdf87` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-67 — `04-RESEARCH.md:570` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 570
command: go test ./internal/compiler/native/... -run TestAbortSignalHandling
classification: R2
verdict: superseded
superseding-phase: P04
superseding-commit: f1b3bad
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f1b3bad` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

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
deleting-commit: 9b54f95
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `9b54f95` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

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


Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `9b54f95` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-71 — `08-VALIDATION.md:65` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-72 — `09-RESEARCH.md:706` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-RESEARCH.md
line: 706
command: go test ./internal/compiler/corevalidate -run TestBuildLoanChainIndex
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: f093310
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f093310` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

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
superseding-commit: f093310
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f093310` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-75 — `09-VALIDATION.md:95` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 95
command: go test ./internal/compiler/corevalidate -run 'Mode.*Invalid|DecodeMode' -v
classification: R2
verdict: superseded
superseding-phase: P07
superseding-commit: 09f98cc
covering-command: go test ./internal/compiler/corevalidate -run TestDerivePeerSignatureModeMutantPairing
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `09f98cc` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-76 — `09-VALIDATION.md:114` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 114
command: grep -c "OWN-05a" .planning/REQUIREMENTS.md
classification: R3
verdict: superseded
superseding-phase: P09
superseding-commit: c738976
covering-command: grep -c "OWN-05a" .planning/MILESTONES.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `c738976` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-77 — `09-VALIDATION.md:115` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/09-peer-re-derivation-and-d-03-02-closure/09-VALIDATION.md
line: 115
command: grep -c "S-008" .planning/ROADMAP.md
classification: R3
verdict: superseded
superseding-phase: P09
superseding-commit: 92cb8a0
covering-command: grep -c "S-008" .planning/MILESTONES.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `92cb8a0` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

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
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-80 — `11-VALIDATION.md:28` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 28
command: go test ./internal/compiler/<touched-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-81 — `11-VALIDATION.md:53` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-82 — `11-VALIDATION.md:58` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-83 — `11-VALIDATION.md:60` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-84 — `11-VALIDATION.md:62` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-85 — `11-VALIDATION.md:63` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-86 — `11-VALIDATION.md:65` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-87 — `11-VALIDATION.md:66` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-88 — `11-VALIDATION.md:70` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-89 — `11-VALIDATION.md:71` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-90 — `11-VALIDATION.md:73` cites a dead pattern, corrected as a rename

first-recorded: M003


Plan 14-10's groundedness-lint reconciliation. The archived command above no longer resolves to any test the static index can find. The replacement command above does resolve (verified via `classifyCommand` over the static test index, the same primitive the lint itself uses) -- the archived row is left byte-unmodified; only this register records the correction.

Closed after the archived validation command was corrected in Plan 02, commit `5c210a5`; this identity is no longer present in the live reconciliation scan.

### D-14-91 — `12-RESEARCH.md:869` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/12-result-payloads/12-RESEARCH.md
line: 869
command: go test ./internal/compiler/<package>/... -run <TestName>
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-92 — `12-VALIDATION.md:26` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/12-result-payloads/12-VALIDATION.md
line: 26
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-93 — `07-VERIFICATION.md:87` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M002-phases/07-calls-signatures-and-call-graph-refusal/07-VERIFICATION.md
line: 87
command: grep -nE 'TBD|FIXME|XXX'
classification: R1
verdict: superseded
superseding-phase: P14
superseding-commit: 0a78834
covering-command: go test ./internal/compiler/session/... -run TestVerificationGroundedness
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `0a78834` (phase P14) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-94 — `03-RESEARCH.md:853` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/03-borrowed-views-and-cfg-lifetimes/03-RESEARCH.md
line: 853
command: go run ./cmd/lang -- <interface-export/import flow>
classification: R1
verdict: superseded
superseding-phase: P07
superseding-commit: dd051bc
covering-command: go test ./internal/compiler/corevalidate -run TestSummaryPeerCallableAgreesOnOriginOmittedClass
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `dd051bc` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-95 — `04-RESEARCH.md:554` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/04-fallible-resources-and-c-boundary/04-RESEARCH.md
line: 554
command: go test ./internal/compiler/... -run <Test...>
classification: R1
verdict: superseded
superseding-phase: P04
superseding-commit: 17e9cb3
covering-command: go test ./internal/compiler/session/... -run TestReleaseTranspositionMutationIsMismatch
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `17e9cb3` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-96 — `05-RESEARCH.md:420` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M001-phases/05-native-equivalence-and-adversarial-evidence/05-RESEARCH.md
line: 420
command: go test ./internal/compiler/... -run <TestName>
classification: R1
verdict: superseded
superseding-phase: P04
superseding-commit: f1b3bad
covering-command: go test ./internal/compiler/native/... -run TestAbortSignalAdjudicatedByWaitStatus
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f1b3bad` (phase P04) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-97 — `14-01-SUMMARY.md:160` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 160
command: grep -nE 'TBD|FIXME|XXX'
classification: R1
verdict: superseded
superseding-phase: P14
superseding-commit: 0a78834
covering-command: go test ./internal/compiler/session/... -run TestVerificationGroundedness
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `0a78834` (phase P14) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-98 — `14-01-SUMMARY.md:162` cites a symbol deleted by design

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 162
command: go test ./internal/compiler/check/... -run TestComputeLoanLastUsesAndDerivePlaceLoansAgree
classification: R2
verdict: obsolete-by-design
deleted-package: internal/compiler/check
deleted-symbol: computeLoanLastUses
deleting-phase: P09-09
deleting-commit: 9b54f95
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `9b54f95` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-99 — `14-01-SUMMARY.md:163` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 165
command: go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -v
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: f093310
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f093310` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-101 — `14-01-SUMMARY.md:166` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 166
command: go test ./internal/compiler/corevalidate -run 'Mode.*Invalid|DecodeMode' -v
classification: R2
verdict: superseded
superseding-phase: P07
superseding-commit: 09f98cc
covering-command: go test ./internal/compiler/corevalidate -run TestDerivePeerSignatureModeMutantPairing
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `09f98cc` (phase P07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-102 — `14-01-SUMMARY.md:169` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 169
command: go test ./internal/compiler/<touched-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-103 — `14-01-SUMMARY.md:170` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 181
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-114 — `14-01-SUMMARY.md:183` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-01-SUMMARY.md
line: 183
command: go test ./<changed-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-115 — `14-RESEARCH.md:448` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md
line: 448
command: go test ./internal/compiler/<package>/... -run <TestName> -count=1
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-116 — `14-RESEARCH.md:459` cites a dead pattern, corrected as a rename

first-recorded: M003

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md
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
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
line: 26
command: go test ./<changed-package>/...
classification: R1
verdict: superseded
superseding-phase: P14-09
superseding-commit: 2a48e27
covering-command: go test ./internal/compiler/session/... -run TestValidationRowGradesAreEarned
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2a48e27` (phase P14-09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

### D-14-118 — `ADVERSARIAL-SYNTHESIS.md:213` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/research/M003/ADVERSARIAL-SYNTHESIS.md
line: 213
command: grep 'D-11-25' PHASE-11-DEBT.md
classification: R3
verdict: superseded
superseding-phase: P14-07
superseding-commit: 2b942d4
covering-command: grep -c "D-11-25" .planning/UNREACHABLE-CLAIMS.md
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `2b942d4` (phase P14-07) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

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
deleting-commit: 9b54f95
```

Plan 14-10's groundedness-lint reconciliation. The archived command above exercises `computeLoanLastUses`, deleted from `internal/compiler/check` at commit `9b54f95` (phase P09-09). The lint confirms `computeLoanLastUses` is absent from `internal/compiler/check`'s production declarations via the same AST scan the callsite: witness grammar uses (D-14-23) -- a falsifiable positive claim, never a textual grep returning zero matches. The archived row is left byte-unmodified; only this register records the correction.

### D-14-120 — `ADVERSARIAL-SYNTHESIS.md:214` cites a superseded target

first-recorded: M003

```reconciliation
file: .planning/research/M003/ADVERSARIAL-SYNTHESIS.md
line: 214
command: go test ./internal/compiler/corevalidate -run 'LoanChainIndex' -count=1
classification: R2
verdict: superseded
superseding-phase: P09
superseding-commit: f093310
covering-command: go test ./internal/compiler/corevalidate -run TestPeerLoanCarryDerivesForwardFromOperations
```

Plan 14-10's groundedness-lint reconciliation. The archived command above names a test, path, or artifact that no longer serves as evidence for the claim it once made. Commit `f093310` (phase P09) is where the surviving, equivalent coverage landed; the covering command above resolves today (verified via `classifyCommand`/a real subprocess check, the same mechanism the lint's own R2/R3 classification uses). The archived row is left byte-unmodified; only this register records the correction.

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

**Update (plan 14-12) — RESOLVED.** Plan 14-11 fixed the root cause this
row identified as suspected but unconfirmed: `evidenceRunRecordTimeout`
raised from the then-300s ceiling to a measured-honest `480s`, a new
`evidenceRunRecordMarginFraction` (`0.75`) that fails closed on a
near-timeout run instead of silently degrading to a lower grade ceiling,
and `resolvedPkgPatterns` anchoring the run-record producer to execute
exactly the resolved, exact top-level test names the consumer
(`deriveCeiling`) consults — closing the substring-match risk that could
have caused the original 300.11s run's own timing pressure. Plan 14-11's
own measured corpus-wide elapsed dropped to ~90.30s (recorded in
`qlt02_budget_manifest.json`'s `evidence_run_record_wall_clock_ns`).

Plan 14-12 then removed the `14-VALIDATION.md` entry from
`validationGradeBarExemptions` for real (commit `172a893`) and re-ran
`TestValidationRowGradesAreEarnedOverArchivedCorpus`: the run record
reported complete, with a measured elapsed of `1m32.646s` (92.6s) against
the `480s` budget — 19.3% of budget, comfortably inside the 75% margin,
and nowhere near the original `300.11s`/`300s`-ceiling near-miss. Five
pre-existing rows in `14-VALIDATION.md`'s own table failed the bar for
real, un-rewritable reasons (see D-14-123 through D-14-127) and are
narrowed via `validationGradeBarRowExemptions`, not via re-adding the
file-scoped exemption; the guard that makes the exemption's return
mechanically impossible (`TestValidationGradeBarAppliesToPhase14`) is
this row's own closing witness.

None of the nine other archived files' rows regressed — the original
spurious cross-file failures were exactly the incomplete-run-record
symptom plan 14-11 fixed, not a real defect in any of those nine files.

**Landing phase:** `CLOSED(172a893)` — plan 14-12's commit removing the
`14-VALIDATION.md` exemption entry and landing the permanent guard.

### D-14-122 — `14-VERIFICATION.md:133` cited a seeded-fault placeholder name; row retired, reconciliation removed

first-recorded: M003

Plan 14-11's fail-closed margin check (Task 2) originally surfaced this
while measuring the corpus-wide run record's real cost: with D-14-121's
own risk closed (the run finished well under its margin, see plan 14-11's
own SUMMARY), `TestValidationRowGradesAreEarnedOverArchivedCorpus` derived
a genuine `WIRED` ceiling for `14-VALIDATION.md` row `14-10-T1` -- not a
false negative from an incomplete run, but a real, previously-unreported
gap in `14-VERIFICATION.md`'s own Non-Inertness Spot-Check table (row 1),
added after plan 14-10's reconciliation pass closed the R1/R2/R3 frontier
and therefore never reconciled at the time. Plan 14-11 recorded a
`reconciliation` block here classifying that cell as R1 with a `renamed`
verdict pointing at `TestVerificationGroundednessFrontierIsPinned`.

Commit `5adcd32` ("docs(phase-14): complete phase execution") then
rewrote `14-VERIFICATION.md` wholesale as part of the phase's final
verification pass. That rewrite replaced the entire Non-Inertness
Spot-Check narrative -- the seeded-fault placeholder text
`TestThisNameDoesNotExistAnywhereZZQQ` this row was reconciling no longer
appears anywhere in `14-VERIFICATION.md` (confirmed by direct grep: zero
matches). This is not a case of the cited line moving to a new location
inside the same document (which `renamed`/`superseded` verdicts exist to
cover) -- the specific text the `reconciliation` block pointed at was
deleted outright, together with the row that contained it. The
reconciliation entry has nothing left to reconcile: keeping it would mean
carrying forward a verdict about a command that no longer exists in the
tree, which is exactly the kind of stale ballast
`TestEvidenceReconciliationViewCatchesStaleEntry` (D-14-14b) exists to
catch on the *generated view* side; on the *authored register* side, the
honest move is to remove the block rather than repoint it at a new anchor
that would misrepresent what happened.

The `reconciliation` fenced block is therefore removed from this Detail
section (and `.planning/EVIDENCE-RECONCILIATION.md` regenerated to drop
the corresponding row). The `D-14-122` identifier and this Detail section
are kept -- not retired outright -- because `PHASE-14-DEBT.md:1582`
still cites it by name as the last entry in the "archived row left
untouched, only the register records the correction" precedent chain
(D-14-48 through D-14-122); retiring the ID would break that citation
without fixing anything real. No future phase needs to act on this row.

**Landing phase:** `P14` -- closed by plan 14-11 recording the original
correction; this entry (the fix you are reading) removes the now-stale
reconciliation block after commit `5adcd32` deleted its target. No
further work is scheduled.

### D-14-123 — `14-VALIDATION.md:14-01-T2` under-declares its now-measurable ceiling

first-recorded: M003

Plan 14-12's removal of `14-VALIDATION.md`'s file-scoped exemption
(`validationGradeBarExemptions`) subjects every row in the file's own
Per-Task Verification Map to the `>=EXERCISED` satisfying bar for the
first time. Row `14-01-T2` (plan 14-01, "Pin the measured violation
frontier as an exact committed literal") declares `WIRED`. Re-deriving its
evidence cell (`go test ./internal/compiler/session/... -run
'TestVerificationGroundednessFrontierIsPinned' -count=1 -v`) against plan
14-11's completion-witnessed, margin-checked run record confirms the cited
test resolves to exactly one top-level name and that name is
`passedNoSkip` in the real corpus-wide run -- the row's TRUE ceiling is
`EXERCISED`, strictly above its declared `WIRED`. This is not a cap
violation (declared never exceeds derived), only a bar violation, and only
because the bar itself did not exist -- or was file-exempted -- when this
row was authored.

This plan's own prohibition ("never rewrite an archived
`*-VALIDATION.md` row belonging to plans 14-01..14-10 to make the bar
pass") is read to cover this row: it belongs to plan 14-01, predates the
bar, and bumping its declared cell now, in the very plan that is
introducing the bar, would be indistinguishable in the historical record
from silently inflating a grade to pass a new gate. The row is kept
byte-unchanged; the correction is recorded here instead, following the
exact "archived row left untouched, only the register records the
correction" precedent D-14-48 through D-14-122 already established for
every other pre-existing finding in this phase.

**Closure mechanism:** `validationGradeBarRowExemptions["14-VALIDATION.md:14-01-T2"] = "D-14-123"`
(this row), mechanically checked by `TestValidationGradeBarRowExemptionsAreOwned`
-- the row-scoped narrowing plan 14-12's Task 2 introduces specifically for
findings, like this one, that a `*-VALIDATION.md` row cannot fix without
being rewritten.

**Landing phase:** `P14` -- closed by this same plan (14-12): the
row-scoped exemption and its ownership guard ARE the closure. No future
phase needs to act; a future plan MAY choose to retroactively correct
`14-VALIDATION.md`'s own cell if it ever touches that row for an unrelated
reason, at which point this row and its exemption entry should be retired
together.

### D-14-124 — `14-VALIDATION.md:14-02-T1` is structurally capped below the bar

first-recorded: M003

Row `14-02-T1` (plan 14-02, "Build the distinctness corpus and freeze the
pre-fix collision control") declares `REACHABLE`. Its evidence cell,
`go run ./cmd/lang --json check testdata/distinctness/spiral_full.lang`,
is not a `go test` invocation at all -- `deriveCeiling` has no path from a
`go run` command to `EXERCISED`; the ladder structurally stops this row at
`REACHABLE` regardless of any run record's completeness. This is
permanent and cannot be closed by a more complete run: only rewriting the
evidence cell to a `go test -run` invocation naming a specific test could
raise it, and that rewrite is exactly what this plan's never-rewrite
prohibition forbids for a plan-14-01..14-10 row.

**Closure mechanism:** `validationGradeBarRowExemptions["14-VALIDATION.md:14-02-T1"] = "D-14-124"`,
checked the same way as D-14-123.

**Landing phase:** `P14` -- closed by this same plan (14-12) via the
row-scoped exemption. No phase currently owns rewriting plan 14-02's
evidence cell to a `go test`-shaped invocation; if a future plan touches
that row for an unrelated reason, retiring this exemption entry alongside
is the natural follow-up, but nothing is scheduled.

### D-14-125 — `14-VALIDATION.md:14-02-T2` names no `-run` pattern, structurally capped

first-recorded: M003

Row `14-02-T2` (plan 14-02, "Attach the discarded recovery extent as a
skipped_region cause") declares `WIRED`. Its evidence cell, `go test
./internal/compiler/syntax/... ./internal/compiler/diagnostic/...
./internal/compiler/check/... -count=1`, names three packages and no
`-run`/`-list`/`-fuzz`/`-bench` pattern. `deriveCeiling` returns `WIRED`
whenever `parsed.Pattern == ""` -- the exact bare-package shape gap
D-14-50 already recorded for `12-VALIDATION.md:52`, now found a second
time inside this phase's own document. Permanent and evidence-shape-bound,
same reasoning as D-14-124.

**Closure mechanism:** `validationGradeBarRowExemptions["14-VALIDATION.md:14-02-T2"] = "D-14-125"`.

**Landing phase:** `P14` -- closed by this same plan (14-12) via the
row-scoped exemption, same disposition as D-14-124.

### D-14-126 — `14-VALIDATION.md:14-03-T1` names no `-run` pattern, structurally capped

first-recorded: M003

Row `14-03-T1` (plan 14-03, "Carry the decline reason through selectRepair
onto the Outcome envelope") declares `WIRED`. Its evidence cell, `go test
./cmd/lang-repair/... -count=1`, names one package and no `-run` pattern
-- the identical structural shape as D-14-125, in a different package.
Permanent and evidence-shape-bound.

**Closure mechanism:** `validationGradeBarRowExemptions["14-VALIDATION.md:14-03-T1"] = "D-14-126"`.

**Landing phase:** `P14` -- closed by this same plan (14-12) via the
row-scoped exemption, same disposition as D-14-124/D-14-125.

### D-14-127 — `14-VALIDATION.md:14-10-T3` under-declares its now-measurable ceiling

first-recorded: M003

Row `14-10-T3` (plan 14-10, "Empty three of four frontier classes and
close the phase's evidence") declares `WIRED`. Re-deriving its evidence
cell (`go test ./internal/compiler/session/... -run
'TestVerificationGroundednessThreeClassesAreEmpty' -count=1 -v`) against
the complete corpus-wide run record confirms the cited test resolves and
passed -- TRUE ceiling `EXERCISED`, the identical finding shape as
D-14-123, in a row belonging to plan 14-10 instead of plan 14-01.

**Closure mechanism:** `validationGradeBarRowExemptions["14-VALIDATION.md:14-10-T3"] = "D-14-127"`.

**Landing phase:** `P14` -- closed by this same plan (14-12) via the
row-scoped exemption, same disposition as D-14-123.

### D-14-128 — Phase 20's unfiltered suite gate is a bare package command

first-recorded: M003

The unfiltered `go test ./... -count=1` preflight in Phase 20 row
`20-08-01` is intentionally executed by Plan 08 as a full-suite gate. The
grade ladder does not infer an exact test identity from a package-wide
command, so WIRED is the honest ceiling until that gate is run. The row is
narrowed through `validationGradeBarRowExemptions` and is owned by Phase 20.

### D-14-129 — Phase 20's timing script is a reachable non-test command

first-recorded: M003

The bounded full-suite timing runner in row `20-08-02` is a script, not a
Go test identifier. The grade ladder therefore caps it at REACHABLE. The
row remains tied to its actual measurement command and Plan 08 owns the
result; the row-scoped exception is checked against this item.

### D-14-130 — `11-VALIDATION.md:56` still targets the pre-archive roadmap

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VALIDATION.md
line: 56
command: grep -c -E 'flto|escalation, not a pass' .planning/ROADMAP.md
classification: R3
verdict: renamed
replacement: grep -c -E 'flto|escalation, not a pass' .planning/milestones/M002-ROADMAP.md
```

The root roadmap is now the milestone index, while the cited Phase 11 evidence
remains in the M002 archive. The replacement targets that archived roadmap and
resolves against its retained Phase 11 entry; the historical validation row is
left unchanged.

---

### D-14-131 — `13-RESEARCH.md:581` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-RESEARCH.md
line: 581
command: go test ./cmd/lang-repair/... -run TestTwinPairBlame
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestTwinPairBlame
```

The Phase 13 research command remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived record.

### D-14-132 — `13-RESEARCH.md:582` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-RESEARCH.md
line: 582
command: go test ./cmd/lang-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass
```

The Phase 13 research command remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived record.

### D-14-133 — `13-VALIDATION.md:62` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 62
command: go test ./cmd/lang-repair/... -run TestTwinPairBlame -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestTwinPairBlame -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-134 — `13-VALIDATION.md:63` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 63
command: go test ./cmd/lang-repair/... -run TestTwinPairBlameGuard -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestTwinPairBlameGuard -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-135 — `13-VALIDATION.md:64` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 64
command: go test ./cmd/lang-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestRepairDriverFixesEveryDefectClassSinglePass -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-136 — `13-VALIDATION.md:66` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 66
command: go test ./cmd/lang-repair/... -run TestUnrepairableDefectFailsTheGate -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestUnrepairableDefectFailsTheGate -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-137 — `13-VALIDATION.md:67` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 67
command: go test ./cmd/lang-repair/... -run TestWrapCallInTryReverifyFailed -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestWrapCallInTryReverifyFailed -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-138 — `13-VALIDATION.md:68` records the pre-public repair command path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 68
command: go test ./cmd/lang-repair/... -run 'TestRepairDriverSourceNeverReferencesHeldoutFixtures|TestRepairDriverSourceHeldoutScanIsNotInert' -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run 'TestRepairDriverSourceNeverReferencesHeldoutFixtures|TestRepairDriverSourceHeldoutScanIsNotInert' -v -count=1
```

The Phase 13 validation row remains an accurate historical receipt. The
publication migration renamed the current command directory; this entry maps
the old path to its current equivalent without changing the archived row.

### D-14-139 — `13-VALIDATION.md:84` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 84
command: go test ./cmd/lang-repair/... -run 'TestRepairDriverImportsStayOutsideInternal|TestImportBoundaryTestIsNotInert' -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run 'TestRepairDriverImportsStayOutsideInternal|TestImportBoundaryTestIsNotInert' -v -count=1
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

### D-14-140 — `13-VALIDATION.md:103` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/13-VALIDATION.md
line: 103
command: go test ./cmd/lang-repair/... -run TestTwinPairBlameGuard -v -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestTwinPairBlameGuard -v -count=1
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

### D-14-141 — `14-RESEARCH.md:463` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-RESEARCH.md
line: 463
command: go test ./cmd/lang-repair/... -run TestUnrepairableAlwaysCarriesDiagnosis -v
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run TestUnrepairableAlwaysCarriesDiagnosis -v
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

### D-14-142 — `14-VALIDATION.md:62` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
line: 62
command: go test ./cmd/lang-repair/... -run 'TestUnrepairableAlwaysCarriesDiagnosis' -count=1 -v
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run 'TestUnrepairableAlwaysCarriesDiagnosis' -count=1 -v
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

### D-14-143 — `14-VALIDATION.md:63` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M003-phases/14-evidence-instrument-and-honest-scoping/14-VALIDATION.md
line: 63
command: go test ./cmd/lang-repair/... -run 'TestUnrepairableDiagnosisGuardIsNotInert' -count=1 -v
classification: R2
verdict: renamed
replacement: go test ./cmd/schway-repair/... -run 'TestUnrepairableDiagnosisGuardIsNotInert' -count=1 -v
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

### D-14-144 — `23-VALIDATION.md:40` records a pre-public command package path

first-recorded: M004

```reconciliation
file: .planning/milestones/M004-phases/23-live-local-allocation-and-discharge/23-VALIDATION.md
line: 40
command: go test ./cmd/lang ./internal/compiler/native -run '^TestPhase23PublicFileByte' -count=1
classification: R2
verdict: renamed
replacement: go test ./cmd/schway ./internal/compiler/native -run '^TestPhase23PublicFileByte' -count=1
```

The recorded command remains historical evidence. The public Schway migration
renamed the current command package; this entry maps the old package path to
its current equivalent without rewriting the validation or research record.

*Dated amendment: 2026-09-29, Schway public repository preparation.*
*Register: PHASE-14-DEBT.md*

*Dated amendment: 2026-10-02, M005 Phase 26 integration.* D-14-144's
document locator now follows M004 archival. Its historical command, line,
verdict, and replacement remain unchanged.
