# Phase 11 — Collected Evidence-Weakened Claims

**Read this before writing 11-VERIFICATION.md.** The claims below are
**honest-scope reductions, deliberately made and evidenced**, not
omissions. Each one reads, in the roadmap's own prose, as a plain pass of
the requirement's literal text. Each one is NOT that: it is a requirement
satisfied by a narrower, evidence-backed fact than the prose implies. The
phase record must report every one of these as **weakened by evidence**,
never as a clean, unqualified pass.

---

## 1. NAT-05 — satisfied by an explicit, visible EMPTY SET (D-11-10)

NAT-05 asks for call-boundary alias attributes. Phase 11 emits **zero**
of them, by construction: two pointers to one object cannot exist across
a Lang call boundary in this milestone's language (one parameter per
function, no globals, no callbacks, no address-escaping foreign
contracts, and `check` already refuses passing one place to two calls).
A call-boundary `restrict` would promise about a hazard that cannot
exist, so the mid-phase gate's zero-attribute state is TERMINAL for this
phase, not a waypoint still being built toward.

The emitted artifact carries a generated comment recording that the
call-boundary attribute set is empty and why, citing D-11-09. This is a
requirement **weakened by evidence**: it reads as done and is not the
same claim the roadmap's prose implies (an attribute set with real,
exercised members).

**Evidence:** `11-MIDPHASE-GATE.md` (the N=1 would-carry count, the
zero-attribute scan result, and the ratification that this state is
terminal for Phase 11).

---

## 2. NAT-06 / criterion 2 — `-flto` is INERT BY CONSTRUCTION over the production corpus (D-11-25)

All Lang functions land in one translation unit (`emitProgram`'s own
whole-program C17 assembler), so `-flto`'s cross-TU optimization tier is
a green no-op for Lang-to-Lang code: there is no second Lang-side
translation unit for it to reach across. The existing lane's own
non-inertness is borrowed entirely from the FOREIGN translation-unit
boundary (the frozen foreign TU plan 4 already links against), not from
anything Phase 11 built.

NAT-07's hand-written negative control (a 4-way TU split: callee,
writer, a separate coordination wrapper calling the callee twice, and
caller) proves the LTO TIER ITSELF can be exploited on this toolchain.
It does NOT make the corpus's own LTO tier meaningful for Lang-to-Lang
code — the corpus never has more than one Lang TU to begin with. This is
a requirement **inert by construction** over the corpus this phase
actually ships, not a claim that LTO was meaningfully exercised end to
end by the shipped pipeline.

**Evidence:** `11-GUARD-LEDGER.md` (single-TU emission fact),
`11-NAT07-EVIDENCE.md` (the toolchain-pinned 4-way-split negative control
and its own recorded scope).

---

## 3. QLT-06 — split; 06b is COMPLETE-BY-ABSTENTION (D-11-39)

QLT-06 splits into two halves on the OWN-05a/05b precedent:

- **QLT-06a — Complete (Phase 07).** `ClosureDigest` chains over callee
  signature summaries, mutation-killed by two independent knowers
  (`check`'s own chain and corevalidate's independent re-derivation),
  including `TestCalleeChangeInvalidatesCallerClosureDigest`.
- **QLT-06b — Complete-by-abstention (Phase 11).** Nothing
  interprocedural is cached, deliberately. `ArtifactSpec.FixtureSource`
  already hashes the WHOLE program, which strictly dominates any
  call-graph-closure key in a single-unit language: a closure key here
  would be a soundness-LOOSENING change, not a soundness improvement, so
  none was built. The S-006 eviction figures (100%/92%/43%) cited earlier
  as motivation are an upper bound on a MODEL of digest-chaining
  behavior, not a measurement of this codebase's actual cache — do not
  re-quote them as measured (D-11-40).

This requirement is **weakened by evidence**: QLT-06b reads as "cache
soundness for interprocedural facts, discharged" and is discharged
structurally (by there being nothing separate to cache), not by building
and testing a cache.

**Evidence:** `11-QLT06-ABSTENTION.md` (the full split record and the
whole-program-hash-dominance argument).

---

## 4. QLT-05 — ships FULL, not narrowed (Q-01 BRANCH A)

Unlike the three items above, this one is recorded here to be explicit
about which branch shipped: plan 11-01's Q-01 spike settled BRANCH A
(accepted) — `corevalidate` accepts the core-level `OpCall`-to-`OpCopy`
rewrite `drop-call-site` performs, so `drop-call-site` ships as a full,
LIVE whole-program reduce move (plan 11-08), never narrowed behind
`reduce.RefusedShapes()`. QLT-05 is therefore **not** an evidence-weakened
claim on this axis: the reducer's own whole-program moves are unnarrowed,
and this plan (11-09) closes the remaining two gaps BRANCH A left open —
the `foreignCallSequenceFor` silent-slippage hole (D-11-33) and
re-verification inheriting the search's own relaxation (D-11-32) — both
under strict, cold-start, anti-vacuity-gated re-verification.

The one genuinely narrowed fact living under the QLT-05 umbrella is
`corevalidate.peerDeriveOriginFacts`'s own pre-existing `OpCall` gap
(a Phase 10 carry-forward, not something this phase's reducer work
narrows further): it is named in `RefusedShapes()`'s own disclosed
register and tracked as debt, not silently absorbed into a "QLT-05
complete" claim.

**Evidence:** `PHASE-11-DEBT.md` (the Phase 10 carry-forward register),
`11-08-SUMMARY.md` (Q-01 BRANCH A's own settlement and the live-move
decision), this plan's own `11-09-SUMMARY.md` (the closed
slippage/re-verification gaps).

---

## Summary table

| Claim | Roadmap prose implies | Actually satisfied by | Evidence |
|---|---|---|---|
| NAT-05 | A real, exercised call-boundary attribute set | An explicit, visible EMPTY set (D-11-10) | `11-MIDPHASE-GATE.md` |
| NAT-06 / criterion 2 | `-flto` meaningfully exercised on the shipped corpus | `-flto` inert by construction over a single-TU corpus (D-11-25) | `11-GUARD-LEDGER.md`, `11-NAT07-EVIDENCE.md` |
| QLT-06 | A closure-keyed interprocedural cache, built and tested | QLT-06a built (Phase 07); QLT-06b complete-by-abstention (D-11-39) — nothing built, by design | `11-QLT06-ABSTENTION.md` |
| QLT-05 | Full reducer re-verification, no caveats | Ships FULL (Q-01 BRANCH A); one pre-existing, disclosed, tracked debt item (`peerDeriveOriginFacts`'s `OpCall` gap) sits alongside it, not inside it | `PHASE-11-DEBT.md`, `11-08-SUMMARY.md` |

None of the four rows above may be reported in `11-VERIFICATION.md` as a
clean, unqualified pass of the requirement's literal prose. Each is a
real pass of a narrower, evidenced, and explicitly stated claim.
