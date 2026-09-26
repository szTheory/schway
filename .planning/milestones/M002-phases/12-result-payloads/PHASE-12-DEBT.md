---
phase: 12-result-payloads
recorded: 2026-09-12
status: accepted
disposition: planning-time
items: 10
blocking: 0
---

# Phase 12 — Declared Deferred Scope and Ratified Branch Decisions

**Written:** 2026-09-12, at phase **open** (plan 01), following PHASE-11-DEBT.md's
own stated discipline of recording debt at phase open rather than phase close.
Shape follows the mechanically-checked debt-register format
`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go:2613`)
enforces on every `*-DEBT.md` register: an `items:` count matching the
`## Items` table, one `### <ID>` detail section per row, a severity from the
closed vocabulary (blocker, warning, info), and a non-empty landing phase.

This register carries Phase 12's own recorded-not-built decisions
(D-12-24, D-12-30, D-12-21, D-12-42, D-12-04c) plus one re-filed item created
by plan 01's Task 2 checkpoint outcome (D-12-36).

**Amended at phase close (plan 08).** D-12-44 (CR-01's FIX disposition,
closed by plans 06-07) and D-12-45 (the WR-01/WR-02/IN-01 secondary-finding
dispositions, all closed) were added as new rows. D-12-43 was amended with
a dated paragraph recording plan 08 Task 1's blocking-human checkpoint
outcome; its original measurement narrative is unchanged.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-12-24 | 12-CONTEXT.md (D-12-24), extends D-12-23 | RES-03 | info | UNOWNED(niche-optimization-uninstantiable) | DEFINED | n/a | NICHE OPTIMIZATION IS DESIGNED AND RECORDED, NOT BUILT, because its precondition is presently UNINSTANTIABLE, not merely undischarged. The only payload types at this maturity are `Byte` and `Buffer`; `Byte` is plausibly fully inhabited (every bit pattern of an `unsigned char` is a valid `Byte`), so there is no niche bit pattern to exploit. The candidate mechanism is an extension of the checker-derived `RecordLayout`/`LayoutField` path (the pattern `standardForeignLayout` already establishes) with a new inhabitance/niche fact derived by the checker from a type's enumerated bit-pattern space — explicitly NOT a hand-declared ability |
| D-12-30 | 12-CONTEXT.md (D-12-30), extends D-12-27/D-12-28/D-12-29 | RES-02 | warning | UNOWNED(resource-in-payload-full-rule) | DEFINED | n/a | THE FULL RESOURCE-IN-PAYLOAD RULE IS DEFERRED, refused this phase by a named fail-closed diagnostic. Criterion 1's *resource* half is deferred, not met; criterion 1 is satisfied with `Byte`, `Buffer`, and nullary-ADT payloads only. The landing condition is three-part and ALL THREE are jointly required: D-10-C01 closed (the missing `OpCall` arm added to `peerDeriveOriginFacts`) AND D-10-C02 proven order-independent or fixed AND D-10-C04 reviewed |
| D-12-21 | 12-CONTEXT.md (D-12-21), flags interaction with D-11-51 | NAT-06 | warning | CLOSED(15ee061) | MUTATION-KILLED | probe:TestPhase15CollisionGuardIsNotInert | D-11-51's PER-INVOCATION EVENT IDENTITY GAP was an anticipated Phase 12 constraint, not silently absorbed there. Phase 15 closes the shared mechanism with canonical invocation paths, four-tier shared-leaf agreement, a moved-frontier assertion, and a live collision mutation; the Phase 12 digest ledger remains preserved alongside an explicit `/2` successor ledger. See the dated closure paragraph below. |
| D-12-42 | 12-CONTEXT.md (D-12-42) | QLT-09 | info | UNOWNED(qlt09-phase5-nyquist-portion) | DEFINED | n/a | QLT-09's PHASE-5 NYQUIST PORTION IS NOT CLOSED MID-PHASE. Phase 12 is the native-tier change QLT-09 was waiting on, but the change is still in flight within Phase 12; closing it mid-phase would conflate a fresh semantic-verification commit with a Nyquist-debt closure (D-11-01's digest-conflation concern). `12-VALIDATION.md` scopes itself to criterion 2's new control and explicitly declines to close QLT-09's Phase-5 portion |
| D-12-04c | 12-CONTEXT.md (D-12-04c) | none — stale documentation, not a live risk | info | CLOSED(82da6929) | DEFINED | n/a | `testdata/phase08/relay_depth2_accept.lang`'s HEADER WAS STALE. It claimed `corevalidate` refuses the fixture with `core.move_while_borrowed`, which D-09-03 closed; the pre-flight probe measured `corevalidate.Valid == true` on the unmodified fixture. Plan 04 replaced the stale claim with the correction, naming D-09-03 as the closing decision and D-12-04c as the correction, per D-09-45's record-corrections-never-silently-fix discipline |
| D-12-36 | 12-CONTEXT.md (D-12-36), RATIFIED at plan 01 Task 2, `12-01-SUMMARY.md`; residual ownership reconciled in Phase 20 | NAT-04..NAT-07 | warning | P21 | EXERCISED | probe:TestPhase21LegacyEmitterBodiesRetired | Phase 16 removed the old linear/branch/match public fallback; Phase 21 commit `3204db7` retired the remaining private foreign and by-pointer emitter bodies, with their absence and the sole `emitProgram` public dispatch pinned by the executed AST probe. All affected program shapes remain refused under D-16-11..13. The earlier N=1 convergence reversal remains historically accurate for Phase 12; this closes only the residual private-code liability. See PHASE-11-DEBT.md D-11-02 for the implementation evidence. |
| D-12-26 | 12-CONTEXT.md (D-12-26), plan 05 Task 3 | RES-03 | info | CLOSED(5559929a) | DEFINED | n/a | CRITERION 2'S "ONE MEANING" CLAIM IS OBSERVABLE-BEHAVIOR AGREEMENT, NOT BYTE-IDENTICAL LAYOUT. `interp` has no byte layout at all — its value model carries no size, alignment, or offset anywhere — so "one meaning in the core IR, the interpreter, and emitted C17" cannot mean byte-identical layout across all three engines, because that claim cannot be true. It means: the same alternative is live, the same payload value is extracted, and the same events are emitted in the same order, with layout-as-bytes a C-only obligation policed by `_Static_assert`/`offsetof` pairs (plan 05 Task 1's `PayloadLayoutMutationRunner`) for internal self-consistency. Byte-identical layout across `interp`, the core IR, and emitted C17 is explicitly NOT claimed |
| D-12-43 | plan 05 Task 2's empirical measurement, extends D-12-38/D-12-39/D-12-41 | RES-03 | warning | UNOWNED(wrong-slot-value-divergence-unconstructible) | WIRED | probe:TestD1243ControlIsUnconstructible | D-12-38's DECISIVE WRONG-SLOT VALUE-DIVERGENCE CONTROL IS UNCONSTRUCTIBLE AGAINST THE CURRENT REPRESENTATION — an absence-of-applicable-channel finding, not a failed engineering attempt. `TestPayloadSlotSwapMutationKilled` seeds a real, type-safe bug in `cgen`'s `OpConstructPayload` codegen (a correct tag, but the payload written into a DIFFERENT alternative's struct field, via the new `cgen.SetPayloadSlotSwapForTest` seam) and drives it through interpreter/-O0/-O3 comparison. Measured result: NO disagreement on any of the five axes. A payload-carrying return's `Outcome.Value` is, on BOTH engines, always the alternative's own compile-time-known TAG NAME — `cgen`'s `returnLiteral` is a literal string baked into the generated C at emission time (never read back from the runtime struct), and `interp`'s `value.String()` resolves through the tag field, never the payload bytes (D-12-26: `interp` has no byte layout to diverge in). So a wrong-SLOT payload write is structurally invisible to every axis `session_phase5_compare.go` compares today. Per D-12-41/D-11-36 this is escalated as a defect in the criterion, not a quiet downgrade to a weaker (e.g. compile-failure-only) assertion: `TestPayloadSlotSwapMutationKilled` pins the absence as a genuine PASSING regression test, mirroring `TestC03PeerDeriveOriginFactsOpCallGapStillOpen`'s own gap-pinning precedent, rather than falsely claiming the mutation was caught |
| D-12-44 | `12-REVIEW.md` CR-01, `12-VERIFICATION.md` gap 1, disposition recorded in `12-06-PLAN.md`'s `<cr01_disposition>` | RES-02, RES-03 | blocker | CLOSED(1292c2d2) | DEFINED | n/a | CR-01'S DISPOSITION IS FIX, RECORDED AS A DECISION, NOT AN INFERRED OUTCOME. `cgen` and `interp` each resolved a payload operation's declaring alternative via an ambiguous first-match linear scan, with no refusal anywhere in the pipeline for the colliding-`PayloadType` case that triggers it — the review's sole CRITICAL finding and the verifier's only hard gap. Closed by a fail-closed `check.duplicate_payload_type` declaration-time refusal (plan 06) plus a shared, ambiguity-detecting `core.AlternativeNameForPayloadType`/`core.LookupAlternativeDetail` resolver replacing both engines' independent copies (plan 07). Imposes a RESTRICTION on the source language: two alternatives of one data type may not declare the same payload type. Lifting condition: GEN-01's real generic `Result<T, E>` (M003) collides by construction at `T == E`; lifting the refusal requires promoting the alternative name to a first-class fact on the operation |
| D-12-45 | `12-REVIEW.md` WR-01, WR-02, IN-01 | RES-02 | info | CLOSED(1292c2d2) | DEFINED | n/a | THE THREE SECONDARY REVIEW FINDINGS ARE EACH CLOSED; NONE WAS DEFERRED. WR-01 (misleading binder-source message) fixed plan 06 Task 3. WR-02 (fault-injection seam could no-op vacuously) fixed plan 07 Task 3, now asserting an injected wrong-slot write count of at least 1 (observed: 2). IN-01 (`emitBranch`'s local reimplementation of alternative-detail lookup) fixed plan 07 Task 1 via `core.LookupAlternativeDetail` |

## Detail

### D-12-24 — niche optimization is designed and recorded, not built

first-recorded: M002

D-12-23 establishes that niche optimization requires a provably uninhabited
bit pattern in the payload type to hide the tag in. The only payload types at
this maturity are `Byte` and `Buffer`, and `Byte` is plausibly fully
inhabited — every bit pattern of an `unsigned char` is a valid `Byte`. There
is therefore no niche to exploit, and RES-03's niche clause is not merely
undischarged, it is **uninstantiable**. This is the same shape of finding as
D-11-09, where Phase 11 emitted zero call-boundary alias attributes because
"two pointers to one object cannot exist across a Lang call boundary in
M002" — a mechanism whose precondition cannot presently exist, so the zero
state is the phase's terminal state, not a waypoint. This is stated as an
**absence-of-applicable-input finding**, never as a silent skip.

The candidate mechanism, recorded for when the reopening condition fires: an
extension of the existing checker-derived layout path
(`RecordLayout`/`LayoutField`, following `standardForeignLayout`'s pattern)
with a new inhabitance/niche fact derived by the checker from a type's
enumerated bit-pattern space — explicitly NOT a hand-declared ability.

**Reopening condition:** a payload type is added whose declared
representation has a provably invalid bit pattern (a non-null-guaranteed
pointer, or a range-restricted integer).

**Confirmed and sharpened by plan 05 (Tasks 1-2).** Building D-12-37's
frozen-fixture layout control and D-12-38's decisive slot-swap mutation
control surfaced nothing that changes this finding: neither control needed,
nor could construct, a niche bit pattern for `Byte` or `Buffer`. The finding
stands unchanged — niche optimization remains **uninstantiable**, not merely
undischarged, at this maturity.

### D-12-30 — the resource-payload rule's three-part landing condition

first-recorded: M002

A named refusal rejects any alternative whose payload type structurally
contains a Phase-4 tracked-resource-derived value. RES-02 and criterion 1 are
satisfied this phase with `Byte`, `Buffer`, and nullary-ADT payloads —
ordinary move-by-function-end, no resource involved. The honest cost: RES-01's
resource half is deferred, not met.

The full rule lands only when **all three** of the following are true:
- **D-10-C01 is closed** — the `OpCall` arm added to
  `corevalidate.peerDeriveOriginFacts`.
- **D-10-C02 is proven order-independent or fixed** —
  `peerCalleeFrameDrained`'s unstated forward-pass ordering assumption.
- **D-10-C04 is reviewed** — the unreviewed D-09-51 negative-control verdict
  flip.

Both sub-hazards land precisely on `corevalidate`'s two open, unreviewed
gaps: the construction side (a resource moved into a payload, then returned)
risks D-10-C02's unproven ordering assumption; the match/forwarding side (a
payload constructed in a callee and forwarded out) hits D-10-C01's missing
`OpCall` case. Admitting the rule now risks discovering a genuinely new
interprocedural fact after fixtures are written — exactly the "hidden need
for a new interprocedural rule" criterion 3 says must trigger a slip, not a
silent absorption.

### D-12-21 — the D-11-51 interaction is anticipated, not absorbed

first-recorded: M002

Widening the interpreter value type (D-12-17) is orthogonal to D-11-51 —
event-ID identity derives from the callee's static `OpReturn`, untouched by
the value shape. But a multi-call-site fixture exercising a
payload-carrying return is exactly the kind of new fixture likely to trip
that open defect for the first time. This row exists so the interaction
surfaces as an anticipated constraint rather than a mystery
`validateExecution` refusal ("duplicate execution event id").

D-11-51 remains **OPEN and UNOWNED**: its fix needs both `interp.go` and
`cgen_program.go`, and therefore its own reviewed plan. Phase 12 does not
claim it.

Closed 2026-09-19 by Phase 15 (`15ee061`), without rewriting the Phase 12
history above. Canonical invocation paths now distinguish repeated callee
occurrences in both engines. The unchanged shared-leaf fixture passes
interpreter, `-O0`, `-O3`, and `-O3 -flto` in
`TestPhase11InterproceduralDifferential/DiamondSharedLeaf`;
`TestPhase15DiamondFrontierMoved` records the old refusal's movement; and
`TestPhase15CollisionGuardIsNotInert` recreates the duplicate pair and proves
the positive gate is not vacuous. `TestPayloadCorpusCharacterizationReplay` retains
the original Phase 12 digest map and separately pins every authorized
multi-function `lang.execution/2` successor, so closure does not relabel the
historical payload evidence.

### D-12-42 — QLT-09's Phase-5 Nyquist portion is not closed this phase

first-recorded: M002

Phase 12 is the native-tier change QLT-09 was waiting on ("deliberately
deferred because both surfaces are touched again by M002 (native tier, agent
surface); validating now risks validating a shape that is about to change"),
but the change is still in flight within Phase 12. Closing it mid-phase
would conflate a fresh semantic-verification commit with a Nyquist-debt
closure — D-11-01's digest-conflation concern.

`12-VALIDATION.md` scopes itself to criterion 2's new control and explicitly
declines to close QLT-09's Phase-5 portion. It becomes validatable once
Phase 12's layout work is frozen (after the D-12-31/D-12-36 branch resolves),
making end-of-Phase-12 or Phase 13 the defensible home, decided against a
pre-registered threshold in the QLT-07/D-09-40 style rather than folded
silently into this criterion.

### D-12-04c — `relay_depth2_accept.lang`'s stale header — CLOSED (plan 04)

first-recorded: M002

The fixture's header stated that "corevalidate ... also refuses this
fixture (`core.move_while_borrowed`) even though `check`'s own
interprocedural law ... correctly admits it." Verified no longer true: the
Phase 12 pre-flight probe measured `corevalidate.Valid == true` on the
unmodified fixture. Phase 09's D-09-03 closed it deliberately — retiring
`relay_depth2_accept.lang` and `twin_a_accept.lang` from
`peerDivergenceExpected` was a required assertion of Phase 09, and the
register's retirement comment says so — but the fixture's own header was
never updated. Recorded per D-09-45 rather than silently fixed.

**Closed by plan 04, Task 3.** `testdata/phase08/relay_depth2_accept.lang`'s
header comment block was replaced with a dated D-12-04c correction naming
D-09-03 as the closing decision, quoting the superseded text verbatim (never
silently erased, per D-09-45), and recording that the pre-flight probe's
first draft went red for exactly this reason — the fixture was not wrong,
the header was. Not one line of Lang source below the header changed
(verified by `git diff --stat`).

### D-12-36 — the inherited D-11-02 deletion was re-deferred, with residual scope now owned by Phase 21

first-recorded: M002

Plan 01's Task 1 committed `TestN1ConvergenceDifferential`
(`internal/compiler/cgen/cgen_n1_convergence_test.go`, commit `1bac938`),
driving both the legacy single-function `Emit` path and `emitProgram` over
five representative single-function shapes. Measured result: `emitProgram`
refuses four of the five shapes outright —
`testdata/phase1/toggle.lang` (match-only), `testdata/phase3/borrowed_view.lang`
(branch/match+linear), `testdata/phase4/foreign_acquire_one.lang`
(foreign-call blocks), and `testdata/phase4/defect_terminal.lang` (branch
with a defect terminator) — and on the fifth, `testdata/phase2/owned_transfer.lang`,
both paths succeed but their outputs differ, both textually (a call-boundary
attribute comment and an extra `#include` shift the preamble) and
structurally (`emitProgram` writes a `static` function plus a separate
`main`; `emitLinear` writes one inline `main`).

Making the differential green requires porting branch bodies, foreign-call
block bodies, and both by-pointer lowering variants into `emitProgram`,
converging the preamble, and re-pinning the four frozen golden-C digests —
roughly 1,500 lines of `cgen.go` emitter logic. That is not "cheap" under any
reading, which is D-12-36's single named trigger (12-CONTEXT.md).

At plan 01's Task 2 checkpoint, the developer ratified **RE-DEFER** over
DELETE. Reasoning, recorded verbatim in `12-01-SUMMARY.md`: the "two
coexisting laws" tax D-12-31 feared does not materialize for payload `case`
arms specifically, because `emitProgram` cannot express a match body at all
— the payload match surface stays exclusively legacy-family territory this
phase, so the payload arms are still written exactly once. Plans 02-05 as
written are scoped to this branch.

This re-files PHASE-11-DEBT.md's `D-11-02` row with a stated reversal
superseding its landing-phase text — see that file's `### D-11-02` section
for the full quoted supersession. The six emitters (`emitLinear`,
`emitLinearBorrowedByPointer`, `emitLinearBorrowedByPointerPlain`,
`emitLinearForeign`, `emitBranch`, `emitMatch`) remain in `cgen`, byte-untouched
by this decision.

**2026-09-25 — Phase 20 residual ownership reconciliation.** The statement
above records Phase 12's historical state. Phase 16 subsequently removed
`emitLinear`, `emitBranch`, and `emitMatch` from the public production path;
the foreign and two by-pointer implementations remain in `cgen.go` behind the
refusal-only schema-2 boundary. Their named Phase 21 owners are D-16-11..13.
Phase 21 now owns the residual retirement/convergence work represented by this
row. See PHASE-11-DEBT.md D-11-02 for the exact evidence and the preserved
D-10-C04 verdict-review boundary.

**2026-09-26 — Phase 21 residual-body retirement.** Commit `3204db7` removes
the remaining private foreign and by-pointer program-lowering bodies; the
executed `TestPhase21LegacyEmitterBodiesRetired` guard pins their absence and
the sole public `emitProgram` path. The family refusal boundary remains intact
under D-16-11..13. This closes the residual private-code liability without
reversing Phase 12's historical convergence measurement or admitting these
families.

### D-12-26 — criterion 2's "one meaning" claim is observable-behavior agreement — CLOSED (plan 05)

first-recorded: M002

`interp` has no byte layout at all — its value model (`interp.value{tag,
payload string}`, D-12-17) carries no size, alignment, or offset anywhere.
So criterion 2's "one meaning in the core IR, the interpreter, and emitted
C17" **cannot** mean byte-identical layout across all three engines, because
that stronger claim cannot be true — there is no `interp` layout for a C
layout to be identical to.

**What is actually claimed and proven:** the same alternative is live, the
same payload value is extracted, and the same events are emitted in the
same order, across the core IR, `interp`, and emitted C17. Layout-as-bytes
is a **C-only** obligation, policed for internal self-consistency by
`_Static_assert`/`offsetof` pairs — plan 05 Task 1's
`cgen.EmitPayloadConformance`/`session.PayloadLayoutMutationRunner`, proving
a committed, deliberately-transposed fixture is refused at compile time
under `-Werror`.

**What is explicitly NOT claimed:** byte-identical layout across `interp`,
the core IR, and emitted C17. `interp` has nothing to be byte-identical
*to*. A weaker claim stated precisely is worth more than a stronger claim
that cannot be true (12-CONTEXT.md, D-12-26).

Task 1's control (`PayloadLayoutMutationRunner`) is the C-side policing
named above. Task 2's control (`TestPayloadSlotSwapMutationKilled`) is the
observable-behavior policing D-12-26 also requires — see D-12-43 below for
what it actually measured and why the decisive form of that policing is
presently unconstructible.

### D-12-43 — D-12-38's decisive value-divergence control is unconstructible against the current representation

first-recorded: M002

D-12-38 requires a genuine interpreter-vs-native **value** divergence from a
seeded wrong-alternative-payload-slot bug, checked via
`session_phase5_compare.go`'s existing `axis:terminal-outcome`
(D-12-39 measured that this channel already exists and needs no harness
extension). Plan 05 Task 2 built this control for real:
`cgen.SetPayloadSlotSwapForTest` seeds a type-safe wrong-slot write in
`OpConstructPayload`'s codegen (a correct tag, but the payload written into
a **different** alternative's own struct field), and
`TestPayloadSlotSwapMutationKilled` drives plan 02's tracer fixture
(mutated) and plan 03's `payload_borrow_interaction.lang` (unmutated
companion) through interpreter/-O0/-O3 comparison.

**Measured result: no disagreement, on any of the five axes.** A
payload-carrying return's `Outcome.Value` is, on BOTH engines, always the
alternative's own compile-time-known TAG NAME:

- `cgen`'s `returnLiteral` (the `main`'s emitted JSON `"value"` field) is a
  literal string baked into the generated C **at emission time**, from
  `arm.Pattern` — it is never read back from the runtime struct the mutated
  code writes into. No amount of struct-memory corruption can move this
  string.
- `interp`'s `value.String()` (D-12-18) resolves through the value's own
  `tag` field when non-empty, never through `payload` — and
  `OpConstructPayload`'s interp case sets `tag` from
  `alternativeNameForPayloadType`, a structural fact independent of the
  seeded bug entirely (D-12-26: `interp` has no byte layout to corrupt in
  the first place).

So a wrong-SLOT payload write is **structurally invisible** to every axis
`session_phase5_compare.go` compares today — not because the mutation was
built wrong, but because the current grammar (`analyzePayloadArm`, D-12-05)
only ever permits a bare-value match arm's result to be an alternative
*name*, never the raw payload content, so no execution document this phase
can produce ever carries payload bytes in its terminal outcome.

**Per D-12-41/D-11-36, this is escalated as a defect in the criterion**,
never a quiet downgrade to a weaker assertion (e.g. a compile-failure-only
check, which would conflate this control with D-12-37's already-necessary
but already-insufficient one).
`TestPayloadSlotSwapMutationKilled` pins the absence as a genuine PASSING
regression test — mirroring `TestC03PeerDeriveOriginFactsOpCallGapStillOpen`'s
own gap-pinning precedent (D-12-04a) — rather than falsely reporting the
mutation as caught. If this control ever needs to become load-bearing for
real, it requires either widening the grammar to let a match arm's result
expose payload content directly, or adding a payload-value-aware
observation channel to the differential harness — both are new, reviewable
work, not a fix to this plan's own scope.

**Reopening condition:** a future plan changes how a payload-carrying
return's terminal value is derived (for example, permitting a match arm to
return the bound payload directly rather than only reconstructing an
alternative), making payload bytes reachable from `Outcome.Value` for the
first time.

**Ratified 2026-09-13 at plan 12-08 Task 1's `gate="blocking-human"`
checkpoint.** The developer was presented with three options — `ratify`,
`slip`, and `ratify-with-owner` — and replied with the option id `ratify`
verbatim: **"ratify"**, selecting "RATIFY — D-12-43 is criterion 2's
terminal state as worded." No enabling work is scheduled and no future
phase is named as the owner of reopening this finding; `.planning/ROADMAP.md`
gains no new phase and no new scope from this decision. This checkpoint
exists precisely because `12-VERIFICATION.md`'s human-verification item 2
objected that this finding was reached by measurement inside plan 05
without a blocking human checkpoint — unlike D-12-31's comparable branch
decision, which did get one. That process gap is now closed: the finding's
merit is unchanged, but its acceptance is now an explicit, dated, verbatim
human decision rather than an inference from a passing test suite.

The landing phase above remains **OPEN and UNOWNED**, governed by the same
reopening condition stated before this ratification — ratifying the finding
as terminal does not itself constitute new work, so no phase is committed.
WR-02's anti-vacuity assertion (plan 07 Task 3, observed injected wrong-slot
write count: 2 on a real run) now guards `TestPayloadSlotSwapMutationKilled`
against ever passing vacuously — the absence this control measures is a
genuine, exercised absence, not an unexercised no-op.

**2026-09-18 — Grade/Witness added (plan 14-07).** This register now
carries the Grade/Witness columns (D-14-21/D-14-24) across all ten rows.
This row is graded `WIRED` with Witness `probe:TestD1243ControlIsUnconstructible`
(`internal/compiler/session/witness_registry_test.go`) — an executed probe
that re-seeds the exact wrong-slot-payload-write bug via
`cgen.SetPayloadSlotSwapForTest` and FAILS if `session.Phase5CompareEngines`
ever reports a disagreement, unlike `TestPayloadSlotSwapMutationKilled`
above (which logs either outcome and always passes). If this control ever
becomes constructible, `TestD1243ControlIsUnconstructible` goes red before
this row's own "OPEN and UNOWNED" prose would ever be re-read by a human.
The other nine rows are `DEFINED`/`n/a` (closed decisions or recorded-not-built
designs — no executed claim beyond their own existence).

### D-12-44 — CR-01's disposition is FIX, recorded as a decision — CLOSED (plans 06, 07)

first-recorded: M002

**The disposition chosen was FIX, not ratified debt, and why.** `12-06-PLAN.md`'s
`<cr01_disposition>` records this explicitly rather than leaving it implicit
in the existence of the plan, because `12-VERIFICATION.md`'s
human-verification item 1 routed the disposition itself to a human, with
two named branches: "committed fix + re-verification" versus "explicit debt
entry with a named landing condition." Invoking this gap-closure run already
chose FIX; this row is the record, not the choice.

The verifier named two candidate closures: (a) a fail-closed `check.*`
diagnostic, and (b) carrying an `AlternativeName` fact directly on the
operation so resolution never round-trips through `PayloadType`. The chosen
disposition is **(a), plus a narrowed form of (b)'s intent** — not (b) as
literally described. (b) as literally described is an IR-schema change:
adding a field to `core.LinearOperation` touches the JSON wire shape,
`corevalidate`'s replay laws, all six dispatch sites, and the explicitly
enumerated field set `core_convention_absence_test.go`'s
`linearOperationExpectedFields` pins as a tripwire — and the phase's
must-not-regress list requires `TestPayloadCorpusCharacterizationReplay` to
stay byte-identical across the whole pre-Phase-12 corpus, a bar an IR field
addition is the single most likely change in this space to fail. A
check-time refusal cannot move those bytes, and plan 06 Task 2 measured
that directly rather than assuming it. (a) alone is not sufficient on its
own — two independent first-match derivations left in `cgen` and `interp`
is exactly the defect class D-12-25 exists to prevent, and `check`'s
refusal only governs programs that cross the parser (`session.PayloadProbeDataType`
proves a `core.DataType` can be hand-built and never see `check` at all).
So plan 07 additionally collapsed both engine copies into one shared,
ambiguity-reporting resolver in `core` — (b)'s intent, without its
IR-schema blast radius.

**What closed it.**
- Plan 06: `check.duplicate_payload_type`, a declaration-time refusal added
  to `check.Program`'s `program.Data` walk (`internal/compiler/check/check.go`),
  positioned in the same region as `check.resource_payload_refused` and
  before its early return. Shipped message, verified on
  `testdata/phase12/payload_duplicate_payload_type.lang`:
  > `alternative "Failed" shares payload type "Buffer" with alternative "Ok": a construct/destructure operation cannot be resolved to a declaring alternative unambiguously while two alternatives of one data type collide on payload type (D-12-44, see PHASE-12-DEBT.md)`
- Plan 07: `core.AlternativeNameForPayloadType` and `core.LookupAlternativeDetail`
  (`internal/compiler/core/core.go`) — the single shared, ambiguity-reporting
  derivation, replacing `cgen.alternativeNameForPayloadType` and
  `interp.alternativeNameForPayloadType` (both deleted). Ambiguity error
  text observed on the colliding fixture:
  > `data type "Colliding": payload type "Byte" is ambiguous between alternatives [First Second]`

**What the evidence is — a convergence test is NOT the evidence.** Two
controls go red if the fix is reverted:
- `TestDuplicatePayloadTypeRefused` (plan 06 Task 2). Observed RED (walk
  disabled): `check_payload_test.go:228: expected at least one diagnostic,
  got none`. Observed GREEN (walk restored): all four subtests PASS.
- `TestPayloadAlternativeResolutionHasExactlyOneDerivation` (plan 07 Task 2).
  Observed RED (a stub `alternativeNameForPayloadType` reintroduced into
  `cgen.go`): `core_single_derivation_test.go:194: internal/compiler/cgen/cgen.go:
  found a local declaration of "alternativeNameForPayloadType" -- the shared
  core.AlternativeNameForPayloadType derivation must not be reimplemented
  locally (CR-01/D-12-25)`. Observed GREEN (stub removed): all four
  subtests PASS. Neither demonstration was committed; both `git diff --stat`
  showed zero change to the affected file after restoration.

**The restriction this imposes on the source language, stated plainly:**
two alternatives of one data type may not declare the same payload type.

**The named lifting condition:** GEN-01's real generic `Result<T, E>`
(M003) collides by construction at `T == E` — `Ok(T)`/`Err(T)` share a
payload type whenever the two type parameters are instantiated identically.
Lifting the refusal to admit that case requires promoting the alternative
name to a first-class fact on the operation (the IR-schema change (b)
described, deferred above on corpus-byte-stability grounds). Until GEN-01
lands and that promotion is done, the refusal stands.

**Measured corpus-neutrality result:** `TestPayloadCorpusCharacterizationReplay`
green with the refusal in place — 97 subtests total (62 PASS, 35 SKIP)
across all 10 pre-Phase-12 corpus directories, byte-identical, no golden
re-pinned (`12-06-SUMMARY.md`).

### D-12-45 — the three secondary review findings are each closed — CLOSED (plans 06, 07)

first-recorded: M002

None of `12-REVIEW.md`'s three secondary findings was silently dropped.

**WR-01** (`check.go`, `analyzePayloadArm`'s "binder on nullary alternative"
diagnostic conflated pattern-side and construction-side binders) — fixed
plan 06 Task 3. The fix took the review's second suggested shape: a
conditional message branching on whether `arm.Binder` (pattern side) or
`arm.ConstructBinder` (construction side) supplied the binder, naming the
offending binder identifier and both alternatives on the construction-side
branch. The diagnostic code `check.binder_on_nullary_alternative` was
deliberately left unchanged — no new code was introduced, so no existing
fixture's expected code moves. Pattern-side wording is untouched; the new
construction-side wording, verified on the `Empty => Ok(w)` shape:
> `construction binder "w" names a payload source for alternative "Ok", but this arm's pattern alternative "Empty" declares no payload to bind it from`

**WR-02** (`cgen.go`, `wrongPayloadSlot`'s fault-injection seam could
silently fall back to the correct write with no signal that it had no-oped)
— fixed plan 07 Task 3. `cgen.PayloadSlotSwapInjectedWriteCount()`, a
production-visible counter reader following `SetPayloadSlotSwapForTest`'s
own cross-package precedent, is asserted `>= 1` inside
`TestPayloadSlotSwapMutationKilled`'s mutated subtest before the engines are
compared. Observed on a real run: **2** wrong-slot writes injected
(`12-07-SUMMARY.md`). D-12-43's recorded finding and measurement narrative
are unchanged by this fix — WR-02 guards the control against vacuous
passing; it does not alter what the control measures.

**IN-01** (`cgen.go`, `emitBranch`'s per-alternative `detail` lookup was a
third independent reimplementation of alternative-detail resolution) —
fixed plan 07 Task 1. `emitBranch`'s loop now calls
`core.LookupAlternativeDetail` directly, the same helper `check`'s own
`lookupAlternativeDetail` delegates to. All three consumers (`check`,
`cgen`, `interp`) now read one derivation.

No review finding — WR-01, WR-02, or IN-01 — was deferred to this or any
future phase; all three closed within plans 06-07.
