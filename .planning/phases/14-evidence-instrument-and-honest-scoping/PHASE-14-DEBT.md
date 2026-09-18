---
phase: 14-evidence-instrument-and-honest-scoping
recorded: 2026-09-18
status: accepted
disposition: phase-in-progress
items: 10
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
Verification Map was never filled in beyond its plan-time template.

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

---

*Register: PHASE-14-DEBT.md*
*Opened: 2026-09-18 (plan 14-04)*
