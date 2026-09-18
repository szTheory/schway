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

This register carries three ratified terminal findings, all decided by the
developer at plan 13-07's blocking-human checkpoints rather than absorbed
silently: D-13-02b (DX-06's B1 contract-violation blame is structurally
unreachable at this language maturity), D-13-10a (`use_matching_argument`
withdrawn as unrepairable, narrowing DX-07 to two classes), and D-13-34
(M001's own `testdata/phase6` move/borrow class pairs fail the retro-strengthened
structural distinctness predicate D-13-33 required).

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-13-02b | 13-CONTEXT.md (D-13-02b), established by 13-04, ratified at plan 07's checkpoint on the D-12-43 precedent | DX-06 | warning | P17 | WIRED | probe:TestB1BlameIsStructurallyUnreachable, callsite:internal/compiler/check.resolveBlame=0 | B1 (CONTRACT-VIOLATION BLAME) IS STRUCTURALLY UNREACHABLE IN PRODUCTION AT THIS LANGUAGE MATURITY. `sameType(function.ReturnType, function.Parameter.Type)` is enforced as an admission precondition on every function (`internal/compiler/check/check.go:255`, `:3148`, `:3399`), independent of any call — a callee whose own body contradicts its own declared contract is refused before the interprocedural pass could ever reach it. This is a structural property of the checker's ordering, not a gap in the fixture corpus. The contract-boundary rule (`resolveBlame`/`resolveCycleBlame`) is implemented, exhaustively unit-tested, and correct, and remains ready for the first B1-shaped class. Criterion 3's twin pair (`D-13-28`) genuinely discriminates detection-site blame from caller blame, which is real and useful, but does not exercise the contract-boundary rule's distinctive claim; DX-06 is reported as partially met rather than absorbed as fully met |
| D-13-10a | 13-CONTEXT.md (D-13-10a), found empirically by plan 13-05, ratified at plan 07's checkpoint | DX-07 | warning | P17 | DEFINED (withdrawn) | probe:TestB1BlameIsStructurallyUnreachable | `USE_MATCHING_ARGUMENT` IS WITHDRAWN AS UNREPAIRABLE AT THIS LANGUAGE MATURITY, NARROWING DX-07 TO TWO GENUINELY REPAIRABLE CLASSES. Every Lang function has exactly one parameter and one type fact, so every initialized in-scope place at a call-argument-type-mismatch site shares that one constructor — the unique match the uniqueness gate (D-13-10) could ever find is always the argument's own already-passed place, so the emitted `Replacement` reproduces the original token byte-for-byte. A no-op splice cannot make a mismatched program re-check clean; the driver reports `unrepairable`, not `repaired`. DX-07 ships with two classes instead of three: `move_after_interprocedural_loan` (backward direction only, per D-13-09b) and `wrap_call_in_try`. ROADMAP's scope-cut note already sanctions narrowing to "the two highest-value defect classes," though its stated trigger (phases 07-12 overrunning) did not occur — this is a different and better-evidenced reason |
| D-13-34 | 13-07-PLAN.md Task 1, adjudicated Option B at plan 07's `gate="blocking-human"` checkpoint | M001 held-out evidence integrity (no live M002 requirement; the retro-strengthening obligation is D-13-33) | warning | UNOWNED(probe:TestPhase6HeldoutPairsAreAlphaRenamesOnly) | WIRED | probe:TestPhase6HeldoutPairsAreAlphaRenamesOnly | M001'S `testdata/phase6` MOVE AND BORROW CLASS PAIRS ARE STRUCTURALLY IDENTICAL UNDER THE RETRO-STRENGTHENED PREDICATE D-13-33 REQUIRED — a real hole in shipped M001 evidence, surfaced rather than weakened away. `TestPhase6DefectCorpusIsHeldOut`'s new structural predicate `(bindingCount, matchArmCount, borrowCount, takeCount, maxDepth)` found the `move` pair (`heldout_move_defect.lang` / `derivation_move_defect.lang`) and the `borrow` pair (`heldout_borrow_defect.lang` / `derivation_borrow_defect.lang`) identical on all five components — the pairs differ only in identifier spelling (`item`→`buffer`, `moved_once`→`delivered` for move; `item`→`buffer`, `alpha`/`beta`→`first`/`second` for borrow), exactly the alpha-rename weakness D-13-33 named. Byte-inequality passed on both pairs and always would have. Only the `match` class pair is structurally distinct (3 arms vs 2). Developer ratified **Option B**: leave the strengthened predicate in place, do not edit the shipped `testdata/phase6/` fixtures, and record this as permanent M001 evidence debt rather than closing it inside a Phase 13 budget (option C, replacement fixtures) or reversing the deliberately-chosen stricter scope (option D, scoping the predicate down to `testdata/phase13` only) |

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

### D-13-34 — `testdata/phase6` move and borrow class pairs fail the retro-strengthened structural predicate — RATIFIED (plan 07)

first-recorded: M002

Plan 13-07 Task 1 replaced `TestPhase6DefectCorpusIsHeldOut`'s
byte-inequality-only assertion with an identifier-independent structural
predicate — `(bindingCount, matchArmCount, borrowCount, takeCount, maxDepth)`,
computed from the parsed program — per D-13-33's instruction to
retro-strengthen M001's own held-out distinctness control, not only Phase
13's. `testdata/phase6` has zero interprocedural fixtures (verified by
directory listing, 13-RESEARCH.md Sec.7), so D-13-26's topology triple
degenerates completely there; this intraprocedural predicate is the mandatory
weaker substitute.

**Measured result, run and recorded rather than assumed:**

| Class | heldout summary | derivation summary | Structurally distinct? |
|-------|------------------|---------------------|--------|
| match | `{bindingCount:0 matchArmCount:3 borrowCount:0 takeCount:0 maxDepth:1}` | `{bindingCount:0 matchArmCount:2 borrowCount:0 takeCount:0 maxDepth:1}` | Yes — differs on `matchArmCount` |
| move | `{bindingCount:2 matchArmCount:0 borrowCount:0 takeCount:2 maxDepth:1}` | `{bindingCount:2 matchArmCount:0 borrowCount:0 takeCount:2 maxDepth:1}` | **No — identical on every component** |
| borrow | `{bindingCount:3 matchArmCount:0 borrowCount:3 takeCount:0 maxDepth:1}` | `{bindingCount:3 matchArmCount:0 borrowCount:3 takeCount:0 maxDepth:1}` | **No — identical on every component** |

Only the `match` pair is structurally distinct. The `move` pair
(`heldout_move_defect.lang` / `derivation_move_defect.lang`) and the `borrow`
pair (`heldout_borrow_defect.lang` / `derivation_borrow_defect.lang`) are
identical on all five components — they differ only in identifier spelling
(`item`→`buffer`, `moved_once`→`delivered` for move; `item`→`buffer`,
`alpha`/`beta`→`first`/`second` for borrow), exactly the alpha-rename
weakness D-13-33 named and the reason byte-inequality alone (which still
passes on both pairs) was never a real distinctness control.

**`TestPhase6DefectCorpusDistinctnessGuardIsNotInert`** (D-13-30(b)'s required
mutation kill at intraprocedural scale) confirms the predicate is genuinely
discriminating, not merely lenient: an alpha-renamed copy of
`derivation_match_defect.lang` produces a structural summary identical to the
original, and the real `heldout_match_defect.lang` is structurally distinct
from `derivation_match_defect.lang` — proving that had the renamed copy been
submitted as the held-out member, the equality check would have caught it.

**Adjudicated 2026-09-13 at plan 13-07's `gate="blocking-human"` checkpoint —
Option B, "accept the finding as M001 evidence debt."** Rationale recorded
verbatim from the developer: authoring replacement M001 fixtures (option C)
is new M001 work re-opening a shipped milestone's corpus inside a Phase 13
budget; scoping the predicate down to `testdata/phase13` only (option D)
would reverse the stricter branch the developer deliberately chose during
this phase's discussion (13-CONTEXT.md "Specifics"). The strengthened
predicate stays in place, unweakened; the two failing class-pair assertions
are explicitly `t.Skip`'d with a reason naming this finding and pointing at
this debt item; the `match` class assertion and
`TestPhase6DefectCorpusDistinctnessGuardIsNotInert` continue to run and pass.
No `testdata/phase6/` fixture was edited (`git diff --quiet -- testdata/phase6/`
exits 0).

**Reopening condition:** a future plan authors replacement `testdata/phase6`
move/borrow fixtures with genuinely distinct structure (option C), or needs to
reason about M001's held-out evidence for those two classes for some other
reason. No phase currently owns this.
