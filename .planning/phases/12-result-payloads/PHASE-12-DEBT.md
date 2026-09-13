---
phase: 12-result-payloads
recorded: 2026-09-12
status: accepted
disposition: planning-time
items: 8
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

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-12-24 | 12-CONTEXT.md (D-12-24), extends D-12-23 | RES-03 | info | OPEN — reopens only if a payload type is added whose declared representation has a provably invalid bit pattern (a non-null-guaranteed pointer, or a range-restricted integer); no phase currently owns this work | NICHE OPTIMIZATION IS DESIGNED AND RECORDED, NOT BUILT, because its precondition is presently UNINSTANTIABLE, not merely undischarged. The only payload types at this maturity are `Byte` and `Buffer`; `Byte` is plausibly fully inhabited (every bit pattern of an `unsigned char` is a valid `Byte`), so there is no niche bit pattern to exploit. The candidate mechanism is an extension of the checker-derived `RecordLayout`/`LayoutField` path (the pattern `standardForeignLayout` already establishes) with a new inhabitance/niche fact derived by the checker from a type's enumerated bit-pattern space — explicitly NOT a hand-declared ability |
| D-12-30 | 12-CONTEXT.md (D-12-30), extends D-12-27/D-12-28/D-12-29 | RES-02 | warning | OPEN — lands only when all three of D-10-C01, D-10-C02, and D-10-C04 are resolved; no phase currently owns closing all three | THE FULL RESOURCE-IN-PAYLOAD RULE IS DEFERRED, refused this phase by a named fail-closed diagnostic. Criterion 1's *resource* half is deferred, not met; criterion 1 is satisfied with `Byte`, `Buffer`, and nullary-ADT payloads only. The landing condition is three-part and ALL THREE are jointly required: D-10-C01 closed (the missing `OpCall` arm added to `peerDeriveOriginFacts`) AND D-10-C02 proven order-independent or fixed AND D-10-C04 reviewed |
| D-12-21 | 12-CONTEXT.md (D-12-21), flags interaction with D-11-51 | NAT-06 | warning | OPEN and UNOWNED — D-11-51 remains open and unowned; its fix needs both `interp.go` and `cgen_program.go` and therefore its own reviewed plan, which Phase 12 does not claim | D-11-51's PER-INVOCATION EVENT IDENTITY GAP IS FLAGGED AS LIKELY TO BE TRIPPED FIRST BY THIS PHASE'S FIXTURES. A multi-call-site fixture exercising a payload-carrying return may surface `validateExecution`'s duplicate-execution-event-id refusal for the first time. This is an anticipated constraint, not a mystery — Phase 12 does not attempt to fix D-11-51 |
| D-12-42 | 12-CONTEXT.md (D-12-42) | QLT-09 | info | Phase 13 or end-of-Phase-12 (decided against a pre-registered threshold in the QLT-07/D-09-40 style once Phase 12's layout work is frozen) | QLT-09's PHASE-5 NYQUIST PORTION IS NOT CLOSED MID-PHASE. Phase 12 is the native-tier change QLT-09 was waiting on, but the change is still in flight within Phase 12; closing it mid-phase would conflate a fresh semantic-verification commit with a Nyquist-debt closure (D-11-01's digest-conflation concern). `12-VALIDATION.md` scopes itself to criterion 2's new control and explicitly declines to close QLT-09's Phase-5 portion |
| D-12-04c | 12-CONTEXT.md (D-12-04c) | none — stale documentation, not a live risk | info | CLOSED — Phase 12, plan 04 (header corrected; see this file's `### D-12-04c` section) | `testdata/phase08/relay_depth2_accept.lang`'s HEADER WAS STALE. It claimed `corevalidate` refuses the fixture with `core.move_while_borrowed`, which D-09-03 closed; the pre-flight probe measured `corevalidate.Valid == true` on the unmodified fixture. Plan 04 replaced the stale claim with the correction, naming D-09-03 as the closing decision and D-12-04c as the correction, per D-09-45's record-corrections-never-silently-fix discipline |
| D-12-36 | 12-CONTEXT.md (D-12-36), RATIFIED at plan 01 Task 2, `12-01-SUMMARY.md` | NAT-04..NAT-07 | warning | OPEN and UNOWNED — re-deferred; no phase currently claims porting branch bodies, foreign-call block bodies, and both by-pointer lowering variants into `emitProgram`, converging the preamble, and re-pinning the four frozen golden-C digests | THE INHERITED D-11-02 SIX-EMITTER DELETION IS RE-DEFERRED WITH A STATED REVERSAL, gated on the N=1 convergence differential per `TestN1ConvergenceDifferential` (`internal/compiler/cgen/cgen_n1_convergence_test.go`). The measured five-shape table shows `emitProgram` refuses four of five single-function shapes outright and diverges textually and structurally on the fifth (`owned_transfer.lang`); making the differential green requires ~1,500 lines of `cgen.go` emitter porting work, which is not "cheap" under any reading — D-12-36's single named trigger. The superseded PHASE-11-DEBT.md landing-phase text ("Phase 12 — a green N=1 convergence differential (Q-05), landing inside Phase 12's per-dispatch-site `Result` plans, never as a second sweep") is withdrawn; see PHASE-11-DEBT.md's `### D-11-02` section for the full stated reversal |
| D-12-26 | 12-CONTEXT.md (D-12-26), plan 05 Task 3 | RES-03 | info | CLOSED — Phase 12, plan 05 (the claim is recorded at the strength it can actually be held, permanently; there is no future work to land) | CRITERION 2'S "ONE MEANING" CLAIM IS OBSERVABLE-BEHAVIOR AGREEMENT, NOT BYTE-IDENTICAL LAYOUT. `interp` has no byte layout at all — its value model carries no size, alignment, or offset anywhere — so "one meaning in the core IR, the interpreter, and emitted C17" cannot mean byte-identical layout across all three engines, because that claim cannot be true. It means: the same alternative is live, the same payload value is extracted, and the same events are emitted in the same order, with layout-as-bytes a C-only obligation policed by `_Static_assert`/`offsetof` pairs (plan 05 Task 1's `PayloadLayoutMutationRunner`) for internal self-consistency. Byte-identical layout across `interp`, the core IR, and emitted C17 is explicitly NOT claimed |
| D-12-43 | plan 05 Task 2's empirical measurement, extends D-12-38/D-12-39/D-12-41 | RES-03 | warning | OPEN and UNOWNED — reopens if a future plan changes how a payload-carrying return's terminal value is derived (e.g. exposing payload bytes on a match arm's return path); no phase currently owns this | D-12-38's DECISIVE WRONG-SLOT VALUE-DIVERGENCE CONTROL IS UNCONSTRUCTIBLE AGAINST THE CURRENT REPRESENTATION — an absence-of-applicable-channel finding, not a failed engineering attempt. `TestPayloadSlotSwapMutationKilled` seeds a real, type-safe bug in `cgen`'s `OpConstructPayload` codegen (a correct tag, but the payload written into a DIFFERENT alternative's struct field, via the new `cgen.SetPayloadSlotSwapForTest` seam) and drives it through interpreter/-O0/-O3 comparison. Measured result: NO disagreement on any of the five axes. A payload-carrying return's `Outcome.Value` is, on BOTH engines, always the alternative's own compile-time-known TAG NAME — `cgen`'s `returnLiteral` is a literal string baked into the generated C at emission time (never read back from the runtime struct), and `interp`'s `value.String()` resolves through the tag field, never the payload bytes (D-12-26: `interp` has no byte layout to diverge in). So a wrong-SLOT payload write is structurally invisible to every axis `session_phase5_compare.go` compares today. Per D-12-41/D-11-36 this is escalated as a defect in the criterion, not a quiet downgrade to a weaker (e.g. compile-failure-only) assertion: `TestPayloadSlotSwapMutationKilled` pins the absence as a genuine PASSING regression test, mirroring `TestC03PeerDeriveOriginFactsOpCallGapStillOpen`'s own gap-pinning precedent, rather than falsely claiming the mutation was caught |

## Detail

### D-12-24 — niche optimization is designed and recorded, not built

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

### D-12-42 — QLT-09's Phase-5 Nyquist portion is not closed this phase

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

### D-12-36 — the inherited D-11-02 deletion is re-deferred with a stated reversal

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

### D-12-26 — criterion 2's "one meaning" claim is observable-behavior agreement — CLOSED (plan 05)

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
