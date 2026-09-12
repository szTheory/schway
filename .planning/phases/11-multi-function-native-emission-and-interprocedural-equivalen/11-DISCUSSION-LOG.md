# Phase 11: Multi-Function Native Emission and Interprocedural Equivalence - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-11
**Phase:** 11-multi-function-native-emission-and-interprocedural-equivalence
**Areas discussed:** cgen emission structure; NAT-05 fact-naming channel; mid-phase gate zero-attribute mechanism; HDD reducer widening (QLT-05); QLT-06 cache scope; QLT-03 shape register; NAT-07 composition-only negative control

---

## Method

This discussion did not use the default question-per-area flow. The user supplied a
standing research mandate ("for each fan out for each decision point consider
breadth+depth through all relevant stakeholder-role lenses ... adversarial pass as
well then synthesize perfect one-shot recommendations for each") and a USER-PROFILE
calibration of `opinionated` → `minimal_decisive`, which explicitly directs Claude
not to bounce low/medium-stakes choices back.

Seven decision points were therefore identified by codebase scout, then researched
in parallel by seven `gsd-advisor-researcher` subagents (4 on opus, 3 on sonnet),
each instructed to: fan out across role lenses, research external ecosystems, run an
explicit adversarial pass against its own leading recommendation, and return 1-2
options with one decisive pick. Two agents were given standing permission to return
"the requirement as written is wrong."

Each agent was pointed at specific prior art where the analogue was unusually strong:
Rust's `noalias` saga (G), Build Systems à la Carte / early cutoff (E), Csmith's
exclusion list + swarm testing (F), LLVM `opt-bisect-limit` (C), C-Reduce and
Hypothesis's shrinker on reduction slippage (D).

---

## A. cgen multi-function emission structure

Scout finding that reframed the question: the six emit paths are not function
emitters — each writes a complete translation unit including its own
`int main(int argc, char **argv)`. There is no per-function seam to loop over.

| Option | Description | Selected |
|--------|-------------|----------|
| Extract a body-emitter seam out of all six paths + one shared TU assembler | One convention immediately, smallest steady state | |
| Additive `emitProgram` path, six single-function emitters untouched | Goldens and pinned digests frozen by construction; build-then-delete, deletion gated on an N=1 convergence differential | ✓ |
| Per-function dispatch loop reusing existing selectors | — | |

**Selected:** additive `emitProgram`.
**Notes:** decisive argument was diff attribution — extraction puts
`restrict_borrow.golden.c` and the four pinned digests at `core_test.go:156-159` at
risk in the same diff that introduces calls, so a moved digest is ambiguous between
refactor and feature. Agent conceded the two-emitter-families cost honestly
(two support-block owners, split fault-injection seam, Phase 12 touches both) and
defended it as *scheduled* duplication on the D-09-10 precedent. Also produced two
findings not asked for: the entry-point resolver already has the data it needs in
`buildAdjacency`, and adding an `EntryFunctionID` field to `core.Program` would move
frozen core bytes.

---

## B. NAT-05 fact-naming channel

| Option | Description | Selected |
|--------|-------------|----------|
| Sidecar only (pure D-05-01 extension) | Smallest delta, zero drift surface; but a reviewer reading the C sees `*restrict p` with no named fact | |
| Sidecar normative + inline C comment generated from the identical struct value, plus a text↔manifest cross-scan lane | One construction, two renderers; the D-04-12 ruling already on record | ✓ |

**Selected:** sidecar normative + generated comment.
**Notes:** agent found the repo had already made this exact ruling at
`cgen.go:1980-1983` and argued against re-litigating it. Contributed the
**discharge-pair** identity (callee `justified_by` + caller-side `discharged_by`,
refused on equality not containment because a missing discharge is the forgery
shape). Schema verdict: mint `lang.attributes/0` rather than bump `lang.foreign/0`,
because that manifest is already structurally single-function. Found a live
pre-existing gap: `evidence.go:233` binds `ForeignDigest` only under
`hasForeignContract`, so a pure-Lang `restrict` claim is not digest-bound into
evidence today. Named three residual holes rather than papering them
(`escape:coordinated-source-to-core-false-claim` inherited not closed; naming proves
derivation not survival under inlining; `ClosureDigest` is staleness not
authenticity).

**Superseded in part by G:** since Phase 11 emits zero call-boundary attributes
(D-11-09), this design is recorded and not built (D-11-11/D-11-12).

---

## C. Mid-phase gate zero-attribute mechanism

| Option | Description | Selected |
|--------|-------------|----------|
| Fixture selection — pick shapes that never reach `selectsByPointerLowering` | No new code; but proves zero attributes by proving zero alias facts — tautological | |
| Attribute-suppression profile + `check`-derived N ≥ 1 second knower | The only construction that can fail non-vacuously | ✓ |

**Selected:** suppression profile, permanent, as a lane.
**Notes:** agent's central argument — fixture selection cannot distinguish "we
suppressed attributes" from "this program had nothing to say," which is the M001
three-failure shape. Insisted the manifest's empty `emitted_attributes` is *not* the
second knower (one derivation read twice). Answered the second-convention objection
with a diff-locality test rather than an assurance. Argued permanence on flag-debt
grounds (undeclared permanence is the failure mode, not permanence). Volunteered the
honest weakness: at `-O0` Clang does little with `restrict`, so the gate's claim is
modest — and declined to renegotiate the gate's optimization level over it.

**Conflict with A:** A proposed hard-wiring `ByPointer = false` with no flag.
Adjudicated in favour of C's second knower and anti-vacuity floor; suppression
mechanism folded into A's single `lowering.ByPointer` field.

**Risk flagged, carried into planning:** uncertain whether `check` can report a
would-be-attribute count without importing `cgen`. If not, the second knower
collapses and the gate's independence claim needs redesign.

---

## D. HDD reducer widening (QLT-05)

| Option | Description | Selected |
|--------|-------------|----------|
| Full multi-function HDD with callee inlining | Genuinely minimal output | |
| Two whole-program moves, no inlining, strict re-verification, named refusal register | Validity-preserving by construction; IDs never renumber, so re-verification can be strict | ✓ |

**Selected:** bounded widening, `drop-call-site` + `drop-orphan-function`.
**Notes:** the decisive reframing — **inlining destroys the defect class Phase 11 is
about**, because the call boundary *is* where the promise lives. Agent verified four
premises by compiling probe programs rather than reasoning about them (orphan
functions check clean; devirtualized calls check clean; forward references check
clean; operation IDs are function-prefixed). The ID-prefix fact is what lets
re-verification demand strict `OperationID` equality instead of inheriting the
`CausalRole` relaxation — turning HDD's slippage trap into a falsifiable assertion.
Found a second silent-slippage hole: `foreignCallSequenceFor` returns `nil` for
multi-function programs, so a dropped call into a foreign-acquiring callee would
compare `nil` to `nil` and pass re-verification while reducing to a different bug.
Own adversarial pass found the vacuity mode (zero-move reduction re-verifies
perfectly) and fixed it with one assertion.

**Scope verdict:** criterion 4 does **not** need renegotiating — roughly one plan's
work — *conditional on* Q-01, the core-level `OpCall`→`OpCopy` probe, which is
untested and may be blocked by the Phase 10 `peerDeriveOriginFacts` gap.

---

## E. QLT-06 interprocedural cache-key scope

| Option | Description | Selected |
|--------|-------------|----------|
| Mint a native-tier closure key; feed `ClosureDigest` into `cache.ArtifactSpec` | Matches ThinLTO's key shape | |
| Cache nothing new; split the requirement row | `FixtureSource` already hashes the whole program, which strictly dominates any closure key in a single-unit language | ✓ |

**Selected:** cache nothing new; split QLT-06 into a (Complete, Phase 07) and
b (Complete-by-abstention, Phase 11, structurally gated).
**Notes:** the premise in the prompt was inverted by the finding — a closure key here
is a soundness-*loosening* change, not a tightening, because it can only admit hits
the whole-program hash would reject. Agent also corrected carried evidence: S-006's
eviction figures are an upper bound on a model, not a measurement, because
`ClosureDigest` chains over signature summaries and a body-only edit doesn't move it
(early cutoff by construction). Its asymmetry argument — a miss is a slow build, a
false hit at `-flto` is a published false proof — is the cleanest statement of this
project's caching posture on record.

**Unrelated finding, higher urgency than the question asked:** `cgen`'s source is
not a declared cache input, so editing `cgen` and re-running with unchanged fixtures
can serve a binary built by the old `cgen` against the new interpreter. Not among
D-06-13's four declared escapes. Phase 11 is the phase that rewrites `cgen`.
Confirmation experiment Q-02 defined.

---

## F. QLT-03 call-graph-shape reachability register

| Option | Description | Selected |
|--------|-------------|----------|
| Pure test-emitted artifact | Nothing to rot; but cannot make a negative claim at all | |
| Mechanically-enumerated cell grid, hand-classified `unreachable` only, committed as `qlt03_shape_register.json` | Rows are an axis product, so a cell cannot be absent; unclassified defaults to `gap` and fails the audit | ✓ |

**Selected:** enumerated grid, new file (not an extension of `qlt01_registry.json`).
**Notes:** agent refused to merge into qlt01 because its row key is `spike_id` over
an unrelated universe, making "missing row" ambiguous in both audits. Contributed the
five-class **proof-mechanism table** with one admissible mechanism per class and no
free-text unreachable — and explicitly rejected two candidate mechanisms from the
prompt as theater, including "assert the shape never appears across the corpus"
(tautological for a deterministic generator; keep as drift detector only). Swarm
testing produced `reached_thin` counting as a gap. Confirmed the register is the
honest home for disclosed trust gaps *provided* each gap row carries a falsifier that
goes red when the gap closes (SPARK's justified-unproved-check pattern). Its cheapest
experiment carries a falsifiable prediction.

---

## G. NAT-07 composition-only negative control

| Option | Description | Selected |
|--------|-------------|----------|
| No-mutation "honest defect" fixture | Literally satisfies "fails red before its fix" — but proven **not constructible** in M002 | |
| Coordinated two-site mutation across two TUs | Only honest construction available; divergence empirically verified on this host | ✓ |

**Selected:** two-site mutation, three TUs, 4×3 matrix with exactly one red cell.
**Notes:** the highest-yield agent. Compiled and ran the matrix on this host
(Apple clang 21.0.0, arm64) and dumped IR. Established four things by measurement:
the roadmap's `interpreter == -O0 != -O3` signature is factually wrong for the
interprocedural case (measured: `interp == -O0 == -O3 != -O3 -flto`); `-flto` is
inert by construction if all functions land in one TU, which today's emitter does;
exploitation is non-monotonic in inlining aggressiveness (three behaviours for
morally identical programs); and Clang re-expresses parameter `restrict` as
block-scoped `!alias.scope`/`!noalias` metadata at inline time, whose dominance
information is known-weaker than the attribute it replaces.

Proved Option B impossible rather than merely hard: Lang functions take exactly one
parameter, have no globals, no callbacks, no address-escaping foreign contracts, and
`check` already refuses passing one place to two calls — so two pointers to one
object cannot exist across a Lang call boundary in M002.

**Consequent recommendation, accepted:** emit **zero call-boundary alias attributes**
in Phase 11; make the mid-phase gate's zero-attribute state terminal rather than a
waypoint.

**Exercised its permission to reject the requirement:** criterion 3 needs two
amendments (divergence signature; "fails red before its fix") plus one addition
(toolchain-pinned re-measurement, with an all-green matrix being a lane failure).

**Own verdict on its own control, recorded verbatim in CONTEXT.md:** it survives as
tier evidence plus a mutation-killed validator test, and does **not** survive as
evidence that a shipped Lang `restrict` is sound.

---

## Cross-cutting: one carry-forward item, four independent hits

`corevalidate.peerDeriveOriginFacts` has no `core.OpCall` case. It entered this
session as one of five disclosed Phase 10 carry-forwards and was independently
surfaced by C (gate second-knower risk), D (blocks the reducer's central move), F
(the register's headline negative row) and G (via the unreviewed D-09-51 flip). It
leaves this session as the phase's critical path, with a both-branches planning
instruction in CONTEXT.md.

---

## Adjudications made by the orchestrator

1. **A vs C on attribute suppression.** C's anti-vacuity second knower adopted;
   suppression mechanism folded into A's single `lowering.ByPointer` field rather
   than a parallel emitter.
2. **A vs G on translation-unit topology.** Production corpus stays single-TU (A);
   multi-TU confined to NAT-07's hand-written-C matrix test (G's own cheapest
   experiment), so `native.Runner` is not widened this phase. Consequence — the
   corpus's LTO tier is inert for Lang-to-Lang code — is declared in writing rather
   than left implied.
3. **B vs G on NAT-05.** B's discharge schema recorded as design, not built, since
   G's D-11-09 removes the promises it would describe. NAT-05 is satisfied by an
   explicit empty set, flagged as a requirement weakened by evidence.

## Claude's Discretion

All seven decision points were resolved under the user's standing mandate to
synthesize and proceed. Three are marked in CONTEXT.md for explicit human attention
at the planning checkpoint rather than silent adoption: D-11-09/D-11-10 (NAT-05
weakened), D-11-20/21/22 (three roadmap amendments), D-11-39 (QLT-06 split, one half
satisfied by abstention).

## Deferred Ideas

Recorded in full in CONTEXT.md `<deferred>`. Headline items: deleting the six
single-function emitters (Phase 12, gated on Q-05); call-boundary attributes and the
`lang.attributes/0` discharge schema (no roadmap home — needs a language construct
that can create two pointers to one object); per-function TUs / widening
`native.Runner`; a closure-keyed cache (M003, only if Lang gains compilation units);
reviewing the D-09-51 flip; the two 10-REVIEW items that remain review-enforced
rather than mechanism-enforced.
