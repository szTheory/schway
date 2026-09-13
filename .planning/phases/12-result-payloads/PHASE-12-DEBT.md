---
phase: 12-result-payloads
recorded: 2026-09-12
status: accepted
disposition: planning-time
items: 6
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
| D-12-04c | 12-CONTEXT.md (D-12-04c) | none — stale documentation, not a live risk | info | Phase 12, plan 04 (the header correction) | `testdata/phase08/relay_depth2_accept.lang`'s HEADER IS STALE. It claims `corevalidate` refuses the fixture with `core.move_while_borrowed`, which D-09-03 closed; the pre-flight probe measured `corevalidate.Valid == true` on the unmodified fixture. Recorded per D-09-45 rather than silently fixed |
| D-12-36 | 12-CONTEXT.md (D-12-36), RATIFIED at plan 01 Task 2, `12-01-SUMMARY.md` | NAT-04..NAT-07 | warning | OPEN and UNOWNED — re-deferred; no phase currently claims porting branch bodies, foreign-call block bodies, and both by-pointer lowering variants into `emitProgram`, converging the preamble, and re-pinning the four frozen golden-C digests | THE INHERITED D-11-02 SIX-EMITTER DELETION IS RE-DEFERRED WITH A STATED REVERSAL, gated on the N=1 convergence differential per `TestN1ConvergenceDifferential` (`internal/compiler/cgen/cgen_n1_convergence_test.go`). The measured five-shape table shows `emitProgram` refuses four of five single-function shapes outright and diverges textually and structurally on the fifth (`owned_transfer.lang`); making the differential green requires ~1,500 lines of `cgen.go` emitter porting work, which is not "cheap" under any reading — D-12-36's single named trigger. The superseded PHASE-11-DEBT.md landing-phase text ("Phase 12 — a green N=1 convergence differential (Q-05), landing inside Phase 12's per-dispatch-site `Result` plans, never as a second sweep") is withdrawn; see PHASE-11-DEBT.md's `### D-11-02` section for the full stated reversal |

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

### D-12-04c — `relay_depth2_accept.lang`'s stale header

The fixture's header states that "corevalidate ... also refuses this
fixture (`core.move_while_borrowed`) even though `check`'s own
interprocedural law ... correctly admits it." Verified no longer true: the
Phase 12 pre-flight probe measured `corevalidate.Valid == true` on the
unmodified fixture. Phase 09's D-09-03 closed it deliberately — retiring
`relay_depth2_accept.lang` and `twin_a_accept.lang` from
`peerDivergenceExpected` was a required assertion of Phase 09, and the
register's retirement comment says so — but the fixture's own header was
never updated. Recorded per D-09-45 rather than silently fixed. Plan 04
corrects the header.

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
