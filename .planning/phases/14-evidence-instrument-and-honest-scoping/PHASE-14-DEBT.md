---
phase: 14-evidence-instrument-and-honest-scoping
recorded: 2026-09-18
status: accepted
disposition: phase-in-progress
items: 3
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
"Debt rows to register" section.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-14-45 | PROJECT.md `## Current State` (verified already correct, commit `d21db90`); 14-RESEARCH.md Open Question 1 | EVD-07 | warning | UNOWNED(probe:TestLTOInertnessOnMultiFunctionEmission) | WIRED | probe:TestLTOInertnessOnMultiFunctionEmission | `-flto` IS STRUCTURALLY INERT ON EVERY CGEN-EMITTED MULTI-FUNCTION PROGRAM, AND THIS CLAIM PREVIOUSLY HAD NO DEBT ROW. `cgen.emitProgram` emits one translation unit with no `restrict` attribute, and a foreign-boundary program is refused outright by multi-function emission -- so LTO's whole-program optimization has no cross-TU boundary to exploit and no alias attribute to hoist across. The two existing non-inertness proofs are both non-multi-function: `TestLTOTierIsNotInert` uses a single-function foreign fixture, and NAT-07's composition-only control is hand-written C explicitly not emitted by `cgen` (D-11-24). D-11-25 names this only in a test doc comment -- it has no debt row and was not among the ten unowned items PROJECT.md's `## Current State` enumerates. This row closes that gap |
| D-14-46 | 14-02-SUMMARY.md "Debt rows to register" item 1 | DX-08 | info | UNOWNED(conditional-surface-lands-in-language) | DEFINED | n/a | NOTHING FORCES THE DISTINCTNESS CORPUS TO GROW AS THE LANGUAGE GROWS. When the conditional form (`if`) enters the language, the spiral trio (`spiral_full.lang`, `spiral_narrow.lang`, `spiral_bare.lang`) becomes a VALID program and must be replaced in `testdata/distinctness/` by the then-current not-in-language surface, or the corpus predicate would need to be re-derived against a program that no longer refuses to parse |
| D-14-47 | 14-02-SUMMARY.md "Debt rows to register" item 2 | DX-08 | info | UNOWNED(identity-bearing-cause-span-on-parse-success-path) | DEFINED | n/a | PUTTING MORE CONTENT ON `Cause.Span` DEEPENS THE ALREADY-RECORDED, HALF-ENFORCED RULE THAT A COORDINATE SHIFT MUST NEVER MOVE A DIAGNOSTIC'S ID. This is now bounded-widened for broken programs only -- a whitespace edit inside a broken program's tail (past the point where declaration recovery starts) can now move that program's diagnostic ID, where before plan 14-02's fix the swallowed region was never in the identity basis. This is an acceptable, deliberately bounded widening of an already-recorded hole (Phase 13's coordinate-shift discipline), not a new defect, but it should be reconciled explicitly rather than rediscovered |

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

---

*Register: PHASE-14-DEBT.md*
*Opened: 2026-09-18 (plan 14-04)*
