---
phase: 13-agent-loop-for-interprocedural-defects
recorded: 2026-09-13
status: accepted
disposition: phase-close
items: 3
blocking: 0
---

# Phase 13 — Declared Deferred Scope and Ratified Terminal Findings

**Written:** 2026-09-13, at phase close (plan 07), after all three findings below
were ratified by the developer. Shape follows the mechanically-checked debt-register
format `TestDebtRegistersAreWellFormed`
(`internal/compiler/session/session_test.go:2613`) enforces on every `*-DEBT.md`
register: an `items:` count matching the `## Items` table, one `### <ID>` detail
section per row, a severity from the closed vocabulary (blocker, warning, info),
and a non-empty landing phase.

This register preserves two ratified terminal findings and one later-resolved
finding. D-13-02b records DX-06's structurally unreachable B1 contract-violation
blame at this language maturity; D-13-10a records `use_matching_argument`
withdrawn as unrepairable, narrowing DX-07 to two classes; D-13-34 records the
M001 move/borrow alpha-rename weakness found by D-13-33 and its replacement
closure in M003 Phase 20.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-13-02b | 13-CONTEXT.md (D-13-02b), established by 13-04, ratified at plan 07's checkpoint on the D-12-43 precedent | DX-06 | warning | P17 | WIRED | probe:TestPhase17B1RequiresUnverifiableDeclaredContract, callsite:internal/compiler/check.resolveBlame=0 | B1 (CONTRACT-VIOLATION BLAME) IS STRUCTURALLY UNREACHABLE IN PRODUCTION AT THIS LANGUAGE MATURITY. `sameType(function.ReturnType, function.Parameter.Type)` is enforced as an admission precondition on every function (`internal/compiler/check/check.go:255`, `:3148`, `:3399`), independent of any call — a callee whose own body contradicts its own declared contract is refused before the interprocedural pass could ever reach it. This is a structural property of the checker's ordering, not a gap in the fixture corpus. The contract-boundary rule (`resolveBlame`/`resolveCycleBlame`) is implemented, exhaustively unit-tested, and correct, and remains ready for the first B1-shaped class. Criterion 3's twin pair (`D-13-28`) genuinely discriminates detection-site blame from caller blame, which is real and useful, but does not exercise the contract-boundary rule's distinctive claim; DX-06 is reported as partially met rather than absorbed as fully met. The Phase 13 probe was superseded by Phase 17's `TestPhase17B1RequiresUnverifiableDeclaredContract`, which tests the current user-declared contract boundary and preserves the original finding's successor provenance |
| D-13-10a | 13-CONTEXT.md (D-13-10a), found empirically by plan 13-05, ratified at plan 07's checkpoint | DX-07 | warning | P17 | DEFINED (withdrawn) | probe:TestPhase17B1RequiresUnverifiableDeclaredContract | `USE_MATCHING_ARGUMENT` IS WITHDRAWN AS UNREPAIRABLE AT THIS LANGUAGE MATURITY, NARROWING DX-07 TO TWO GENUINELY REPAIRABLE CLASSES. Every Lang function has exactly one parameter and one type fact, so every initialized in-scope place at a call-argument-type-mismatch site shares that one constructor — the unique match the uniqueness gate (D-13-10) could ever find is always the argument's own already-passed place, so the emitted `Replacement` reproduces the original token byte-for-byte. A no-op splice cannot make a mismatched program re-check clean; the driver reports `unrepairable`, not `repaired`. DX-07 ships with two classes instead of three: `move_after_interprocedural_loan` (backward direction only, per D-13-09b) and `wrap_call_in_try`. ROADMAP's scope-cut note already sanctions narrowing to "the two highest-value defect classes," though its stated trigger (phases 07-12 overrunning) did not occur — this is a different and better-evidenced reason. The former Phase 13 witness was retired when Phase 17 replaced that probe with `TestPhase17B1RequiresUnverifiableDeclaredContract`; the generated current view excludes this closed row |
| D-13-34 | 13-07-PLAN.md Task 1, initially adjudicated Option B; replaced by M003 Phase 20 on 2026-09-25 at the user's explicit choice | M001 held-out evidence integrity (no live M002 requirement; retro-strengthening obligation D-13-33) | warning | CLOSED(4658eb1) | EXERCISED | probe:TestPhase6HeldoutPairsAreStructurallyDistinct | M001'S ORIGINAL MOVE/BORROW HELD-OUT PAIRS WERE ALPHA-RENAME ONLY, AS D-13-33'S FIVE-FIELD STRUCTURAL PREDICATE FOUND. On 2026-09-25 the user selected replacement fixtures in M003 Phase 20. The held-out move summary is now `{bindingCount:3, matchArmCount:0, borrowCount:0, takeCount:3, maxDepth:1}` versus derivation `{bindingCount:2, matchArmCount:0, borrowCount:0, takeCount:2, maxDepth:1}`; held-out borrow is `{bindingCount:4, matchArmCount:0, borrowCount:4, takeCount:0, maxDepth:1}` versus derivation `{bindingCount:3, matchArmCount:0, borrowCount:3, takeCount:0, maxDepth:1}`. All three fixture pairs now differ structurally; existing injectors still produce the intended single ownership defect. D-06-29 structural separation inference is restored for move and borrow while its small-corpus limitation remains.

## Detail

### D-13-02b — B1 contract-violation blame is structurally unreachable — RATIFIED as terminal (plan 07)

first-recorded: M002

See 13-CONTEXT.md's own `D-13-02b` entry for the full derivation; it was
established empirically by plan 13-04 and independently verified by the
orchestrator before this phase's planning closed. This entry records only the
phase-close ratification.

**What this means for DX-06, stated plainly:**
- B1 (contract-violation blame) is implemented, exhaustively unit-tested, and
  correct — but nothing in the shipped language can trigger it, because
  `sameType(ReturnType, Parameter.Type)` refuses a self-contradicting callee at
  admission time, before any interprocedural pass runs.
- Every constructible interprocedural defect is B2 (caller misuse), and B2
  correctly selects the right fix site for those — D-13-04's empirical
  zero-span-movement result holds.
- Criterion 3's twin pair (D-13-28: shared callee `sink`, two callers `alpha`
  and `beta`, one legal and one illegal) discriminates detection-site blame
  from caller blame, which is real and useful, but does not exercise the
  contract-boundary rule's distinctive claim — the rule is ready infrastructure
  for the first B1-shaped class, not yet load-bearing in production.

**Addendum (added at phase-close verification, on the verifier's
recommendation): a third instance of this same structural cause.** D-13-28's
twin-pair design assumed the true fix would land in the shared callee `sink`.
Plan 13-06 found empirically that it does not — the shipped repair operates
entirely **inside the caller** for *both* fixture halves. That is the same root
cause as this entry (the single-type-per-function / `sameType` invariant),
surfacing for a third time after D-13-02b itself and D-13-10a.

It was resolved inside 13-06 as a Rule 1 deviation — tests were built around the
verified behavior rather than the plan's assumed behavior — and is documented in
that plan's own `## Criterion 3 verdict` section rather than routed through this
phase's `gate="blocking-human"` checkpoint, which covered only D-13-33/D-13-34.
Nothing was weakened or concealed, and DX-06's status is unchanged. This
addendum exists so a reader of the debt register **alone** gets the full picture
without having to cross-reference the SUMMARYs.

Full detail: `13-06-SUMMARY.md`, `## Criterion 3 verdict`.

**Ratified 2026-09-13 at plan 13-07's `gate="blocking-human"` checkpoint**, on
the D-12-43 precedent (Phase 12's own unconstructible-decisive-control
ratification). No enabling work is scheduled and no future phase is named as
the owner of reopening this finding. The reopening condition: the language
gains a signature whose return type may differ from its parameter type, i.e.
`sameType` stops being an admission precondition.

### D-13-10a — `use_matching_argument` withdrawn as unrepairable — RATIFIED (plan 07)

first-recorded: M002

See 13-CONTEXT.md's own `D-13-10a` entry for the full derivation, found
empirically by plan 13-05 and confirmed by reading the emission site.

**Root cause, and why it is the same structural property as D-13-02b:** the
branch's own guard establishes `typeFact.Shape.Constructor !=
contract.ParameterType` before the gate runs, so no in-scope place can ever
equal `contract.ParameterType` — 13-05 compared against the caller's own
`typeFact.ID` instead (a Rule 1 deviation). But every Lang function has
exactly one parameter and one type fact, so every initialized place in the
caller shares that one constructor, and the unique match the uniqueness gate
finds is always the argument's own already-passed place. The replacement
reproduces the original token — a no-op splice, verified through the real
driver on a held-out fixture rather than assumed.

**Consequence, ratified:** DX-07 ships with **two** genuinely repairable
classes, not three: `move_after_interprocedural_loan` (backward direction
only, per D-13-09b) and `wrap_call_in_try`. `use_matching_argument` emits no
repair at all (`unrepairable`, never a `MachineApplicable` no-op) rather than
shipping a diagnostic that claims to repair and does not.

**Ratified 2026-09-13 at plan 13-07's `gate="blocking-human"` checkpoint.**
ROADMAP's scope-cut note already sanctions narrowing to "the two
highest-value defect classes," though its stated trigger (phases 07-12
overrunning) did not occur — this is a different and better-evidenced reason,
recorded rather than silently substituted. The reopening condition: the
single-type-per-function invariant relaxes (this language currently admits
exactly one parameter and one type fact per function; a future generics or
multi-parameter surface could reintroduce a genuine uniqueness question).

### D-13-34 — `testdata/phase6` move and borrow held-out pairs — CLOSED (M003 Phase 20)

first-recorded: M002
closed: 2026-09-25

Plan 13-07 strengthened `TestPhase6DefectCorpusIsHeldOut` with the
identifier-independent five-component predicate required by D-13-33. Its
measured finding was real: the original move and borrow pairs were
structurally identical, despite byte inequality. The Phase 13 developer
checkpoint initially ratified keeping those fixtures as evidence debt
(Option B).

At the Phase 20 blocking-human checkpoint, the user selected **replace**.
The held-out move fixture now adds a third take hop and keeps the injector
marker on that hop; the held-out borrow fixture adds a reborrow after the
marked shared borrow. The derivation fixtures and predicate are unchanged.
The probes now require structural inequality for all three classes, with no
known-identical exceptions or skips. Measured summaries are recorded in
`20-D-13-34-DECISION.md` and `20-06-SUMMARY.md`; injector and debt-cap evidence
is recorded there as well.

D-06-29's structural separation inference is restored for move and borrow.
Its residual limit remains explicit: three classes and six fixtures cannot
establish generalization across the full language program space. The initial
Option B adjudication remains part of the chronology; this later replacement
supersedes it. Closure landing commit is recorded in the Phase 20 summary.
