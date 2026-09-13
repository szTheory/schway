# Roadmap: Codename Lang

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1-6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- 🚧 **M002 — Interprocedural Semantic Spine** — Phases 07-13 (active)

## Overview

M002 adds exactly one new `core.OperationKind` — `OpCall` — to a compiler that
already has a working intraprocedural semantic spine, and then proves that every
semantic guarantee M001 established intraprocedurally still holds across a
function boundary. The single addition is deceptively large: it lands at all six
existing dispatch sites (`check`, `corevalidate`, `interp`, `cgen`, `pathoracle`,
`originvalidate`), each paying its own cost, plus two exhaustive-dispatch
controls — a minimum of 8-10 independent edits before `OpCall` is "real" by the
project's own D-04-22 standard.

The organising principle is unchanged from M001: **no fact crosses a trust
boundary without being independently re-derived**, and **no callee body is ever
read to admit a caller** — only a digest-bound signature summary. Phases are cut
where a *gate* becomes meaningful, not where implementation could parallelize.

## Phases

- [x] **Phase 07: Calls, Signatures, and Call-Graph Refusal** - `OpCall` becomes real at all six dispatch sites; cycles are refused, never hung. (completed 2026-09-09)
- [x] **Phase 08: Interprocedural Loan Liveness in `check`** - The checker derives cross-function loan liveness from signatures alone, under a measured, bounded cost. (completed 2026-09-10)
- [x] **Phase 09: Peer Re-Derivation and D-03-02 Closure** - `corevalidate` independently reaches the same interprocedural answer; the milestone's carried debt item closes. (completed 2026-09-10)
- [x] **Phase 10: Trusted Interprocedural Oracle** - `originvalidate`, `pathoracle`, and a bounded interpreter call stack make cross-function execution trustworthy before anything is lowered. (completed 2026-09-11)
- [x] **Phase 11: Multi-Function Native Emission and Interprocedural Equivalence** - `cgen` emits multi-function C17 and the five-axis comparator agrees across `-O0`/`-O3`/`-flto`. (completed 2026-09-12)
- [ ] **Phase 12: `Result` Payloads** - Payload-carrying alternatives are storable, matchable, and affine-correct in all three engines.
- [ ] **Phase 13: Agent Loop for Interprocedural Defects** - `lang explain` and `lang-repair` reach and fix cross-function defect classes through the JSON protocol alone.

## Pre-Phase Spikes

Three bounded spikes follow the project's existing `.planning/spikes/` discipline
(falsifiable question, named competing mechanisms, an oracle that does not reuse
the mechanism under test, at least one injected defect, **stop when the gate is
answered**). None of them is production code and none carries a requirement.

| Spike | Question it must answer | Blocks | Runs |
|-------|-------------------------|--------|------|
| **S-006 Interprocedural liveness cost-scaling probe** — **ANSWERED (VALIDATED, 2026-09-09)**: memoized is linear in program size on every shape; unmemoized is quadratic-to-exponential. The cache is the mechanism, not an optimization. Phase 08 planning unblocked. | Does summary-based liveness stay linear/sub-quadratic in call-graph size, or must a memoized summary cache be designed in from the start? | **Planning of Phase 08** (hard entry gate) — released | Alongside Phase 07 |
| **S-007 Recursive / mutually-recursive stress corpus** | Do a bounded call stack and cycle refusal distinguish legal deep recursion from illegal cycles without false positives on parser-shaped programs? Is the stack ceiling a fixed constant or a declared budget? | Nothing (informs Phase 07's design) | Alongside Phase 07, non-blocking |
| **S-008 Nyquist fold-in cost measurement** — **REPLACED (2026-09-10, before Phase 09 planning)**: a bounded inventory/estimation pre-flight pass, not a formal spike, ran instead and returned QLT-07 **COMMITTED** against a threshold pre-registered before the count was taken (D-09-40/D-09-40a). See the amendment note below the table. | Is closing M001 Phase 3's Nyquist debt genuinely cheap when folded into code Phase 09 already has open? | **QLT-07's commitment status** in Phase 09 — answered by the replacement pass | Before Phase 09 planning; hours, not days |

### Decision: the cost-scaling probe is a pre-phase spike, not a phase and not a Phase 07 gate

Deliberate choice, stated per instruction:

- **Not its own phase.** It carries no M002 requirement, produces throwaway Go
  workbench code, and a requirement-less phase would violate this roadmap's own
  100%-coverage/one-phase-per-requirement discipline while adding a phase whose
  success criteria could only read as task completions.

- **Not a gate inside Phase 07.** Its finding is *Phase 08*-blocking, not
  Phase 07-blocking — Phase 07's gate must be about call-graph refusal, and
  attaching a liveness-cost question to it would let a Phase 07 pass imply a
  liveness answer it never tested.

- **Therefore: a pre-phase spike, recorded here, executed alongside Phase 07,
  and a hard entry gate on Phase 08's *planning***. Phase 08 may not be planned
  until S-006 has answered whether `loanLivenessFixpoint` extends directly or
  needs a memoized summary-caching layer designed in from the start. **Answered:
  both — the backward worklist extends directly (a program-order canonicalization
  pre-pass plus one transfer clause), and the summary table must be memoized
  within a run from the start.** The
  precedent this guards against is real, not hypothetical: Go 1.18's 15-18%
  front-end-specific generics regression and Rust's Polonius performance wall.

A spike is finished when its gate is answered. Expanding one into production
code is an explicit anti-pattern here.

### Amendment: S-008 replaced by a bounded inventory pre-flight pass, not run as a formal spike (D-09-39)

Recorded as an explicit process amendment, not a silent substitution — a
future reader must not find this spike table quietly disagreeing with what
actually happened.

S-008 asked "is closing M001 Phase 3's Nyquist debt genuinely cheap when
folded into code Phase 09 already has open?" That question fails this
project's own spike discipline (`.planning/spikes/CONVENTIONS.md`) on two
independent grounds:

- **Category error.** It has no named competing mechanisms and no oracle
  independent of the mechanism under test — it is a scoping/estimation
  question, identical in kind to the sizing `plan-phase` already does at
  every phase boundary, not a validity question like S-006's
  memoized-versus-unmemoized cost-scaling comparison. Forcing spike shape
  onto it would mean inventing artificial arms, which `CONVENTIONS.md` itself
  argues against.

- **Would reproduce an already-open registry gap a second time.** Standing up
  a `.planning/spikes/008-*` directory would have reproduced D-08-43's
  already-open, un-owned `.planning/spikes` registry-maintenance gap a
  second time (`TestQLT01RegistryCoversAllFiveSpikes` already failed before
  Phase 09 touched anything, because spike 006 had a directory and no
  registry row).

**Replacement:** a bounded inventory/estimation pre-flight pass, run
2026-09-10 during Phase 09 planning, before any `09-PLAN.md` was drafted.
The pass enumerated only the loan-liveness-scoped rows of
`03-VALIDATION.md` (03-03/03-04/03-05, excluding OWN-04's 03-06/03-07 rows)
against a threshold **pre-registered before the count was taken** (D-09-40):
QLT-07 is committed iff the inventory requires ≤ 1 additional plan and opens
zero packages/files that OWN-07/OWN-08's own work does not already touch.
Finding: all 33 named tests already existed and passed
(`go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle`,
all three `ok`). **Verdict: QLT-07 COMMITTED** (D-09-40a). The closure itself
landed in Phase 09's own `09-VALIDATION.md`, "M001 Phase 3 Debt Closure
(loan-liveness subset)" section, ratified at plan 09-10.

## Departure from the Research-Reconciled Build Order

`research/SUMMARY.md` reconciled ARCHITECTURE.md and OPTIONS.md into **6 phases
plus one pre-phase spike**. This roadmap adopts that sequence with **one
departure: SUMMARY's Phase 4 is split into two phases (10 and 11)** at the
ARCHITECTURE.md Stage 6 / Stage 7 boundary.

**Why.** SUMMARY collapsed ARCHITECTURE Stages 5-8 into one phase because they
share a single headline claim (interprocedural `-O3` equivalence) and one
dependency (a trusted interpreter oracle first). Both observations are correct,
and neither survives the cost model:

1. **Size.** That phase would carry 12 of the milestone's 30 requirements and
   two brand-new subsystems that have never existed in this codebase (an
   interpreter `Frame`/call stack, and multi-function C emission). M001's
   recorded per-plan cost for coordinated multi-consumer work is 70-95 min/plan,
   not the ~10 min/plan of single-consumer work. A 12-requirement, 4-stage phase
   at that base rate is a 20-plan phase — larger than any M001 phase, and it
   would hold the milestone's central risk unadjudicated for its entire length.

2. **The dependency SUMMARY names is itself a gate.** "cgen must differential-
   test against a working, trusted interpreter" is precisely a phase boundary:
   the oracle's trustworthiness is a claim that can be gated *before* anything
   depends on it. Phase 10's headline claim is therefore distinct and
   verifiable — *the interpreter is a trustworthy interprocedural oracle,
   independently corroborated by `originvalidate` and `pathoracle`* — and
   Phase 11 keeps SUMMARY's headline claim intact.

3. **SUMMARY's own mid-phase-gate requirement is better served.** It names Stage
   4 and Stage 7 as the two highest-risk integration points, each deserving its
   own mid-phase gate. Under the split, Stage 7 gets a mid-phase gate inside a
   phase that is actually about it, rather than a mid-phase gate buried in a
   four-stage phase whose first half is already done by then.

Everything else in SUMMARY's reconciliation is adopted as written: Stage 0
(signature/summary artifact) lands first as its own plan and gate inside
Phase 07 (ARCHITECTURE.md's finer cut, taken over OPTIONS.md's wordless fold);
Nyquist-for-Phase-3 folds into the loan-liveness peer phase (Phase 09) rather
than becoming a separate pass; `Result` payloads sequence late as the negotiable
item; the agent loop is last because the defect taxonomy is not knowable before
it.

## Parallelism: implementation-level vs. gate-level

SUMMARY flags a real granularity difference between its two sources, and it is
preserved here rather than flattened. ARCHITECTURE.md describes what *can
compile in parallel*; OPTIONS.md describes what a *phase gate* must require in
sequence. Both are true at once.

**Gate-level sequencing (binding — this is the phase order):**

```
07 → 08 → 09 → 10 → 11 → 12 → 13
```

Phase 09's two-peer-agreement question is only meaningful once `check`'s
liveness law exists as a stable target (Phase 08), so `corevalidate`'s *liveness*
peer cannot be gated earlier. Phase 11 cannot be gated before Phase 10 because
the five-axis comparator's authority rests entirely on the interpreter oracle.

**Implementation-level parallelism (non-binding — available inside/across phases
if it helps, and it does not move a gate):**

| Work that can proceed in parallel | Precondition | Caveat |
|---|---|---|
| `corevalidate`'s **non-liveness** engineering (structural re-derivation of `OpCall`, independent cycle re-check) — ARCHITECTURE Stage 3 ∥ Stage 2 | Stage 1 (`OpCall` registered) landed in Phase 07 | Only the non-liveness half. The liveness peer stays gated in Phase 09. |
| `originvalidate` (Site 3) ∥ `pathoracle` (Site 6) — ARCHITECTURE Stage 5 | Phase 09's liveness facts stable | They must not import each other or `check`/`corevalidate`; that independence is the point. |
| `interp` call stack (Stage 6) ∥ Stage 5 | Phase 07's `OpCall` operations exist | Must **stabilize** before Phase 11 begins; it is the oracle. |
| `Result` payload cases (Stage 9 / Phase 12) against Phases 07-11 | None structurally | **Scheduling hazard:** land `Result` payload `case` additions in the *same plans* that already touch each dispatch site for `OpCall`, never as a second sweep over the same six files. Two passes over one file is the repeated-touch inefficiency M001's retrospective flags. |
| S-007 stress corpus ∥ Phase 07; dev-backend timing probe ∥ Phase 11 | None | Both non-blocking; the timing probe informs M003 scope, not M002. |

## Scope-Cut Order

Carried verbatim in substance from `REQUIREMENTS.md` and `research/SUMMARY.md`,
recorded **up front** so the negotiable items are identified before pressure
arrives rather than under it (Key Lesson 4: declare deferred scope in writing at
the moment it is decided).

1. **First cut — Phase 12 (`Result` payloads, RES-02/RES-03).** Structurally
   independent of the call machinery. Slip to M003 *explicitly*; do not silently
   absorb as extra plans on committed scope.

2. **Second cut — QLT-07 (Nyquist fold-in, Phase 09).** If spike S-008 shows the
   fold-in is not actually cheap, re-scope it to a declared stretch item and
   carry it as disclosed debt alongside QLT-09, rather than discovering the cost
   mid-phase.

3. **Third cut — Phase 13 scope narrowing (DX-05/06/07).** Narrow to the two
   highest-value defect classes (cross-function borrow escape, cycle refusal)
   rather than cutting the mutation-kill / repair-then-re-check discipline for
   whatever does ship.

**Explicit trigger:** if **Phase 08 or Phase 09** — the two highest-risk phases —
exceeds **~2x its initial plan estimate**, that is the trigger to renegotiate
Phase 12 out to M003 *immediately*, not to keep silently adding plans.

**Never cut:** SEM-04, SEM-07, OWN-06, OWN-07, OWN-08, NAT-06, NAT-07. D-03-02
is the one deliberately-carried debt item this milestone exists to close;
deferring it again would repeat M001's defer-under-load pattern on the same item
for a second consecutive milestone.

## Phase Details

**Standing refs — apply to every phase 07-13, read before planning any of them:**

- `.planning/STANDING-VERDICTS.md` — dependency verdicts (the zero-external-dep
  record is a hard constraint), design anti-features, process anti-patterns, and
  load-bearing facts about the six dispatch sites, `cgen`'s single-function hard
  fail, and why `-flto` is load-bearing. Reopen an entry only on new evidence.

- `.planning/LANGUAGE-MATURITY.md` — what the language can actually express
  today. Guards against planning as if arithmetic, iteration, or collections
  exist. They do not.

Note: per-phase `Canonical refs` below point into `.planning/research/`, which is
**milestone-scoped and regenerated by the next `/gsd-new-milestone`**. The two
standing refs above are the durable ones.

### Phase 07: Calls, Signatures, and Call-Graph Refusal

**Goal**: A Lang function can call another Lang function, admitted from the
callee's signature alone, and a program whose calls form a cycle is refused by
name instead of hanging.
**Depends on**: M001 Phase 6 (complete)
**Requirements**: SEM-04, SEM-05, SEM-06, SEM-07, QLT-08
**Maps to**: ARCHITECTURE Stage 0 (signature/summary artifact) + Stage 1
(`core.OpCall` registration + `check` producer path) + Stage 2 (`callgraph`
package + cycle refusal)
**Canonical refs**: `research/ARCHITECTURE.md` §1, §2 (Design C), §3;
`research/SUMMARY.md` "The Six Dispatch Sites"; `wiki/semantic-kernel-contract.md`
**Success Criteria** (what must be TRUE):

  1. `lang check` admits a two-function Lang program in which one function calls
     another, with `core.OpCall` handled at all six dispatch sites and both
     exhaustive-dispatch controls (`core_test.go`'s in-process
     `control:kind.exhaustive_dispatch` and `session.go`'s CLI-observable lane)
     green.

  2. No caller admission path reads a callee body — the callee's digest-bound
     signature summary is the only input — and a call to a non-publishable
     callee (violating callable ⊆ publishable, D-04-03) is refused with a stable
     diagnostic code.

  3. Direct, mutual, and indirect call cycles are each refused with a named
     refusal code and never hang. **Gate (Pitfall 4):** an indirect-cycle corpus
     and a pathological-depth-but-acyclic corpus both return bounded verdicts,
     and the traversal is explicit-worklist / visited-set-guarded from day one
     rather than native Go recursion.

  4. Every new interprocedural control introduced in this phase has been
     observed to fail against a seeded mutation in the plan that introduced it.

**Riskiest assumption**: that the Stage 0 signature summary carries *everything*
a caller needs for admission — ownership requirement per parameter and the full
return-origin contract — so that "never open the callee body" survives contact
with the harder phases downstream.
**Gate that catches it being wrong**: Stage 0 lands **first, as its own plan and
its own gate**, validated against the entire existing M001 Phase 4 corpus before
`OpCall` is registered. Calls do not need to exist to prove the summary shape is
correct for every function that already exists. If the summary is
under-specified, that gate fails while the change is still a pure data-shape
change with one producer — not after five consumers depend on it.
**Sizing note**: registering `OpCall` in `AllOperationKinds()` deliberately
breaks all six exhaustive switches at once — a forcing function, converting "six
sites, easy to forget one" into "six failures, impossible to forget one". Expect
Stage 0 as one plan, then coordinated-change plans at the M001 5-consumer base
rate (70-95 min/plan), not the single-consumer rate. **Revised after cross-AI
review:** Stage 0 is **two** plans (`07-01` schema/decoder/digest, `07-02`
predicate/peer/faults) plus a third deferred to the end by D-07-38 (`07-08`
digest chaining, which terminates only on a proven DAG). Stage 1 is three plans
and Stage 2 is two, for eight sequential waves in total.
**Parallel**: S-006 and S-007 spikes run alongside. `corevalidate`'s non-liveness
`OpCall` engineering may start once Stage 1 lands.
**Plans**: 12 plans (8 planned 2026-09-08 after cross-AI review returned Risk:
HIGH on the original 5, resolutions locked as D-07-29..D-07-45; `07-09` added as
gap closure for 07-VERIFICATION.md's single failed truth; `07-10`..`07-12` added
2026-09-09 as gap closure for 07-REVIEW.md's post-verification blockers CR-04,
CR-01, and CR-03. Waves are strictly sequential — every plan overlaps `check.go`
/ `corevalidate.go` / `core.go` / `session_phase7.go` / `scripts/verify-phase7.sh`
with its neighbours, so there is no safe parallelism.)

Plans:
**Wave 1**

- [x] 07-01-PLAN.md — Stage 0a: `lang.interface/1`, `DecodeInterface` strict validation, pinned frozen `/0`, canonical non-self-referential `ClosureDigest` preimage (base case only)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 07-02-PLAN.md — Stage 0b: `Callable` as **publication safety** via `PublishProblemsFor`, the `corevalidate` summary peer at BOTH replay sites, the extracted negative control, and the three seeded faults incl. the bilateral one that must FAIL

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 07-03-PLAN.md — Stage 1a: `core.LinearOperation.CalleeID`, parser `"call"` RHS kind, `core.OpCall` registered at all six dispatch sites with explicit "recognized, unsupported" arms in `interp`/`cgen`

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 07-04-PLAN.md — Stage 1b: both exhaustive-dispatch controls with their own phase-07 lists, asserting recognition not execution, each mutation-killed in this plan

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 07-05-PLAN.md — Stage 1c: SEM-06 — the pre-body signature table, the callable-⊆-publishable refusal, the body-blindness control, and the A-normal-form relay/escort witness

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 07-06-PLAN.md — Stage 2a: the `internal/compiler/callgraph` package, `core.call_graph_cycle` with deterministic witness selection, and the diamond/shared-leaf corpus that kills the gray-vs-visited mutation

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 07-07-PLAN.md — Stage 2b: `corevalidate`'s independent synthetic-artifact traversal peer, the remaining corpora, and the phase-wide QLT-08 completeness matrix by exact set equality

**Wave 8** *(blocked on Wave 7 completion)*

- [x] 07-08-PLAN.md — Stage 0c (deferred by D-07-38): `ClosureDigest` chained over callee summary digests in reverse postorder, **after** acyclicity is proven

**Wave 1 (gap closure — from 07-VERIFICATION.md `gaps_found`)**

- [x] 07-09-PLAN.md — Gap closure: argument-type-vs-declared-parameter-type refusal in `resolveCallBinding`, an independently-derived `corevalidate` peer, `OpCall`'s `TargetID.TypeID` derived from the callee's declared return type, `call_type_mismatch.lang`, and two new mutation-killed controls

**Wave 1 (gap closure — post-verification, from 07-REVIEW.md CR-01/CR-03/CR-04)**

- [x] 07-10-PLAN.md — PVG-03 (CR-04, the amplifier): `lang check` consults `corevalidate` and reports the peer's refusal as an invalid source; `interface export`/`interface core` report a peer refusal as `StatusInvalid` instead of `tool.operation_failed`; an asserted divergence register; WR-01's user-visible half closed incidentally; PVG-04/WR-01/WR-02 dispositions recorded as D-07-49/50/51

**Wave 2 (blocked on 07-10)**

- [x] 07-11-PLAN.md — PVG-01 (CR-01): the ownership half of the call contract — a call consumes its non-copyable argument in `check.resolveCallBinding`, an independently-derived `corevalidate` consume peer, `call_argument_used_twice.lang` / `call_argument_used_once.lang`, two mutation-killed controls, and D-07-07's misleading wording CORRECTED

**Wave 3 (blocked on 07-11)**

- [x] 07-12-PLAN.md — PVG-02 (CR-03): `FunctionSignature.Foreign`/`.Fails` closure-derived by an explicit worst-case join in the acyclic second pass, an independently-implemented peer join over `corevalidate`'s own postorder, `call_fallible_foreign_reach.lang`, the IN-01 index-coupling fix, and two mutation-killed controls

**Declared scope-cut trigger (D-07-26):** if `07-01`..`07-05` exceed ~2x their
initial plan estimate, `07-06`/`07-07` renegotiate out to Phase 08 — **and
`07-08` goes with them**, because D-07-38 orders digest chaining strictly behind
cycle refusal. Never cut: the Stage 0 peer, never the seeded faults. Recorded in
`phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md`.

### Phase 08: Interprocedural Loan Liveness in `check`

**Goal**: The checker decides, from callee signatures alone, whether a loan is
still live across a call boundary — and the decision is bounded, terminating,
and measurably cheap.
**Depends on**: Phase 07; **spike S-006 answered** (hard entry gate on planning)
**Requirements**: OWN-06, EFF-02
**Maps to**: ARCHITECTURE Stage 4, `check`-side half — "the single highest-risk
integration point in the whole milestone"
**Canonical refs**: `research/ARCHITECTURE.md` §4; `research/PITFALLS.md`
Pitfalls 1 and 4; `wiki/compute-efficiency-constitution.md`;
`wiki/ownership-evidence-roadmap.md`
**Success Criteria** (what must be TRUE):

  1. **Gate (Pitfall 1):** `check` refuses a composition-depth-≥2 program whose
     loan state diverges across a call boundary, and accepts its safe twin —
     with the derivation reading callee signatures only, never re-walking a
     callee body.

  2. **Gate (Pitfall 4):** the liveness fixpoint terminates under a fail-closed
     iteration bound; a program engineered to exceed that bound produces a named
     refusal, not a hang and not a silent under-approximation.

  3. Interprocedural admission cost on realistic call-graph fan-out is measured
     under the existing p50/p95/CoV protocol, stays inside a declared bound, and
     is recorded in the feedback-budget manifest.

  4. A human can read, from the shipped artifact, which callee-signature fields
     the liveness answer depended on for a given call site.

**Riskiest assumption**: that a signature-mediated fixpoint stays a bounded,
summary-based analysis rather than becoming whole-program borrow inference in
disguise — precisely the failure mode `compute-efficiency-constitution.md` names
and the design Rust abandoned pre-1.0.
**Gate that catches it being wrong**: criterion 3's cost-scaling measurement on
realistic fan-out, pre-informed by spike S-006. Whole-program inference does not
fail a correctness test; it fails a cost curve. If admission cost grows
super-linearly in call-graph size, the analysis is reading more than the
summary, whatever the code claims.
**Mid-phase gate (mandatory, mirroring M001 Phase 3's precedent)**: the
adversarial composition-depth corpus (criterion 1) and the cost measurement
(criterion 3) are both adjudicated from code-level evidence **before** the
liveness law is declared final and before `corevalidate`'s peer is planned. Open
items are adjudicated or recorded as debt, never assumed safe.
**Parallel**: none at gate level. This phase must produce a stable target before
Phase 09 is meaningful.
**Plans**: 6 plans

Plans:
**Wave 1**

- [x] 08-01-PLAN.md — Tracer: one interprocedural refusal end-to-end (summary bit → forward canonicalization → diagnostic → CLI) plus D-07-49's entry defect in both admission paths

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 08-02-PLAN.md — The summary mechanism: `UsesParam` derived once per function in reverse postorder, memoized and transitive; the backward `OpCall` gate; program-order and never-persisted invariants

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 08-03-PLAN.md — Criterion 1's adversarial corpus: both twin pairs, depth-≥2 relay chain, negative control, match-arm regression, and criterion 4's consulted-field-set assertion

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 08-04-PLAN.md — Criterion 2: the derived fail-closed iteration bound, its named refusal, the seeded mutation-kill, and the explicit-stack pre-walk hardening

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 08-05-PLAN.md — Criterion 3 (EFF-02): the synthetic four-shape corpus, the growth-exponent fit, both gate chokepoints widened together, the manifest row and the risk lane

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 08-06-PLAN.md — The mandatory mid-phase gate, its adjudications, the cost bound ratified at the gate, and the liveness law declared final

### Phase 09: Peer Re-Derivation and D-03-02 Closure

**Goal**: A second, independently-implemented admission layer reaches the same
interprocedural answer as the first — and the one debt item M001 knowingly
carried forward is closed in both layers.
**Depends on**: Phase 08; **spike S-008 answered** (determines QLT-07's
committed-vs-stretch status)
**Requirements**: OWN-05a, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07
**Maps to**: ARCHITECTURE Stage 3 (corevalidate independent re-derivation) +
the `corevalidate` half of Stage 4, with M001 Phase 3's Nyquist closure folded in
**Canonical refs**: `research/PITFALLS.md` Pitfall 2 ("land both peers in the
same plan"); D-02-03 / D-03-01 / D-03-02; `.planning/RETROSPECTIVE.md` Key
Lesson 2
**Success Criteria** (what must be TRUE):

  1. **Gate (Pitfall 2):** a shadow run over recursion, diamond, and deep-chain
     call-graph shapes shows **zero divergence** between every peer that derives
     a given interprocedural fact, **before either peer ships**, landed in the
     same plan.

  2. A deliberately seeded endpoint-level fault in one peer makes the two
     diverge — proving criterion 1's agreement is load-bearing, not vacuous —
     and a counted-work lane shows linear-or-declared-bounded cost for the
     re-derivation.

  3. An exported borrow-derived return with no declared origin is refused in the
     **interprocedural** case by **both** admission layers. D-03-02 is closed.

  4. Ownership transfer at a call site (move vs. borrow, per the callee's
     declared convention) has exactly one meaning, derived independently by
     `check` and `corevalidate`; call-site override of a callee's declared
     convention is not expressible.

  5. Exactly one loan-liveness law exists in the checker at end of phase — the
     intraprocedural law is **deleted**, not left coexisting — and Nyquist
     validation reports compliant for the loan-liveness surface.

**Riskiest assumption**: that `corevalidate` can re-derive interprocedural
liveness *without* sharing an implementation with `check` — that the two peers
will not quietly drift into one derivation with two call sites. This is the
literal site of the D-02-03/D-03-01 "fix one of two peers" hazard the project
already paid for once.
**Gate that catches it being wrong**: criterion 2's seeded endpoint-level fault.
A shared implementation agrees perfectly *and* fails to diverge under a seeded
fault; only the seeded fault distinguishes genuine independence from shared
machinery.
**Scope note**: `interp`'s third independent derivation of ownership transfer
(OWN-05's remaining peer) lands in Phase 10 and is verified there by TRU-03 and
in Phase 11 by NAT-06. QLT-07 is the milestone's second scope-cut candidate; if
S-008 shows the fold-in is not cheap, it is re-declared a stretch item **in
writing at that moment** and carried as disclosed debt alongside QLT-09.
**Parallel**: none at gate level.
**Plans**: 10 plans, in six waves. The ordering is a safety constraint, not a
convenience: the peer lands fully first, the zero-divergence differential goes
green, and only then does the deletion land (D-09-10, build-then-delete).

Plans:
**Wave 1**

- [x] 09-01-PLAN.md — TRACER: `corevalidate`'s own interprocedural loan-liveness derivation, its `OpCall` consult, the differential authored red-first, and seeded faults in both directions
- [x] 09-02-PLAN.md — Enablement: relocate the call-graph corpus generator, widen the third gate-eligible metric chokepoint, and close the one pre-existing test failure

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 09-03-PLAN.md — OWN-08: the peer re-derives all four `PublishProblemsFor` classes; D-03-02 closed in both admission layers
- [x] 09-04-PLAN.md — TRU-04 shapes: the synthetic-shape zero-divergence differential and the cycle-peer differential with witness agreement
- [x] 09-05-PLAN.md — OWN-05: call-site transfer has one meaning; override non-expressible in source and fail-closed in core

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 09-06-PLAN.md — The peer's own closure cost curve, its bound and mutation-kill, and the two-way disclosed-field-set identity
- [x] 09-07-PLAN.md — The three pre-deletion gates: diagnostic ordering stability, D-09-49 Q1's enumeration, and the AST-shadow-path subsumption corpus

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 09-08-PLAN.md — MANDATORY MID-PHASE GATE at the build-then-delete boundary; the peer's bound ratified at the gate's own commit

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 09-09-PLAN.md — OWN-09: extend the post-assembly pass, remove the lowering-time emission, delete the shadow scaffolding — one commit

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 09-10-PLAN.md — QLT-07's scoped closure, the OWN-05a/OWN-05b split, and the requirement and roadmap corrections

### Phase 10: Trusted Interprocedural Oracle

**Goal**: Cross-function execution and origin/path re-derivation are trustworthy
enough to be the authority everything native is differential-tested against.
**Depends on**: Phase 09
**Requirements**: SEM-08, SEM-09, TRU-02, TRU-03, QLT-04, OWN-05b
**Maps to**: ARCHITECTURE Stage 5 (`originvalidate` Site 3 ∥ `pathoracle`
Site 6) + Stage 6 (interpreter call stack, Site 4) + the loan-endpoint half of
Stage 8
**Canonical refs**: `research/ARCHITECTURE.md` §5; the already-proven
`OpForeignCall` hop precedent in `originvalidate`/`pathoracle`
**Success Criteria** (what must be TRUE):

  1. `originvalidate` walks published origins across `OpCall` (mirroring the
     proven `OpForeignCall` hop) and `pathoracle` independently re-derives the
     cross-function loan-chain rule — with neither importing `check` or
     `corevalidate`, enforced by a build- or test-level import control rather
     than convention.

  2. The interpreter executes a multi-function program on a bounded call stack
     with a documented fixed ceiling; exceeding it is a **named refusal**, not a
     host stack overflow. **Gate (Pitfall 4):** a native-stack-overflow probe,
     distinct from the language-level bound, confirms the two limits are not the
     same limit.

  3. Drop and cleanup obligations run in the defined order on normal return
     **and on every nonlocal exit** across a call boundary, observed as ordered
     events, not asserted.

  4. **Gate:** cross-function loan-endpoint differentials rebuild M001 Phase 3's
     exhaustive endpoint enumeration at a **declared, bounded composition
     depth**, with that depth stated in the shipped artifact rather than implied
     by the corpus.

**Riskiest assumption**: that the interpreter's frame model agrees with the two
admission layers about what is legal across a call boundary — the assumption on
which the whole five-axis comparator's authority rests in Phase 11.
**Gate that catches it being wrong**: criterion 4's exhaustive cross-function
endpoint differential, which puts `interp`, `pathoracle`, `check`, and
`corevalidate` on the same enumerated inputs. If the oracle disagrees, it is
caught here — before `cgen` is built on top of it and before any native
equivalence claim is made.
**Parallel**: `originvalidate` ∥ `pathoracle` ∥ `interp` call stack at the
implementation level; a single gate at the end. `interp` must be **stable**, not
merely working, before Phase 11 starts.
**Plans**: 9 plans

Plans:
**Wave 1**

- [x] 10-01-PLAN.md — TRACER: one call boundary executed end to end on an explicit heap frame stack, plus the OWN-05b guard (wave 1)

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 10-02-PLAN.md — `originvalidate` walks published origins across `OpCall`, closing D-09-51, with transitive import-guard hardening (wave 2)
- [x] 10-03-PLAN.md — `pathoracle` composes callee path enumerations across `OpCall`, with its own declared depth cap and the discriminating per-path-borrow fixture (wave 2)
- [x] 10-04-PLAN.md — `MaxCallDepth` with its inverted rationale, the depth-exceeded refusal as comparable data, and the Pitfall-4 subprocess stack probe (wave 2)

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 10-05-PLAN.md — SEM-09 drop/cleanup ordering across frames on normal return and every nonlocal exit, plus the `corevalidate` drain-before-pop signature invariant (wave 3)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 10-06-PLAN.md — MID-PHASE GATE: measure the D-10-59 scope-cut trigger, confirm the four hard ordering constraints, land the D-09-53 reversal's artifact correction (wave 4)

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 10-07-PLAN.md — QLT-04's declared composition depth of 3, the first depth-3 corpus in the tree, and the bidirectional depth gate (wave 5)

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 10-08-PLAN.md — Criterion 4's cross-function loan-endpoint differential: three-way on refuse, four-way on accept, with an accountable divergence register and seeded faults (wave 6)

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 10-09-PLAN.md — The `interp` stability freeze: golden corpus, determinism at `-count=10`, named escalation path, and the structural coverage floor (wave 7)

### Phase 11: Multi-Function Native Emission and Interprocedural Equivalence

**Goal**: A multi-function Lang program lowers to readable C17 and executes
identically under the interpreter, `-O0`, `-O3`, and `-O3 -flto` — with the
optimizer proven to have actually been given something to optimize.
**Depends on**: Phase 10 (a stabilized interpreter oracle is a hard precondition)
**Carried from Phase 10**: the oracle ships with five disclosed residual trust
gaps that are Phase 11 INPUTS, not closed history — most notably
`corevalidate.peerDeriveOriginFacts` having no `core.OpCall` case (constrains
which fixtures are expressible), the `callgraph` import asymmetry that makes
peer independence review-enforced rather than mechanism-enforced (NAT-06 leans
on that independence), and an unreviewed negative-control verdict flip from
D-09-51. Full list with anchors: STATE.md § Blockers/Concerns "Phase 10
carry-forward"; detail in `.planning/phases/10-trusted-interprocedural-oracle/`
(`PHASE-10-DEBT.md`, `deferred-items.md`, `10-REVIEW.md`, `10-VERIFICATION.md`).
**Scope input verified at the planning gate (2026-09-11)**: criterion 1 names
`cgen`'s `Emit`/`EmitNative`, but the single-function assumption is **not two
sites — it is 32 non-test `len(Functions) != 1` guards across 6 files in 3
packages** (55 including tests): `session` 26 (incl. `RunInterpreter`,
`RunNative`, `interpreterInputs` — so a multi-function program checks clean but
is unrunnable from the CLI on *either* engine today — plus `verifyBorrowedCorpus`
×9 and the Phase 5/6/7 verification lanes), `cgen` 4 (`Emit`, `EmitNative`,
`emitLinear`, `emitBranchOperations`), `reduce` 2. **`reduce` is the one that
changes planning**: `reduce.Reduce` hard-errors on a multi-function seed and
`ProjectSource` returns an unsupported-projection string, so **criterion 4's
"HDD reducer output on multi-function programs is re-verified" cannot be met by
widening `cgen` alone** — the reducer is a first-class subject of this phase, not
a downstream consumer of it. Likewise the 26 `session` guards include the lanes
that *are* the five-axis equivalence proof: "runnable" and "provable" are
separate costs here. Re-verify with
`awk '/^func /{f=$0;l=NR} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')`.
Inventory detail: `.planning/LANGUAGE-MATURITY.md` § "The single-function guard
inventory".
**Requirements**: NAT-04, NAT-05, NAT-06, NAT-07, QLT-03, QLT-05, QLT-06
**Maps to**: ARCHITECTURE Stage 7 (multi-function C emission + call lowering —
"the second-highest-risk integration point") + Stage 8's closing verification
**Canonical refs**: `research/ARCHITECTURE.md` §6; `research/PITFALLS.md`
Pitfalls 3, 5, 6; M001 Phase 5's `control:alias.false_no_alias` precedent
**Success Criteria** (what must be TRUE):

  1. `cgen` emits multi-function C17 — `Emit`/`EmitNative` no longer refuse
     `len(Functions) != 1` — and every aliasing or capture promise emitted at a
     call boundary (`restrict`, noalias-shaped attributes) **names the checked
     fact it derives from** in the emitted artifact.

  2. Interpreter, native `-O0`, native `-O3`, and `-O3 -flto` produce equivalent
     semantic outcomes and events for the interprocedural corpus on the existing
     five-axis comparator.

  3. **Gate (Pitfall 3):** at least one **engineered composition-only** negative
     control — the harder `-flto` sequel to `false_no_alias` — reproduces a
     divergence in which the interpreter and `-O0` agree and at least one
     optimized tier disagrees, with the `-flto` tier required to be the
     disagreeing one (D-11-20, ratified 2026-09-12 against
     `11-NAT07-EVIDENCE.md`'s measured matrix). The control is demonstrated red
     under its injected mutation and green unmutated, with a full mutation-kill
     matrix over each injection site independently; an all-green matrix is a
     lane failure, not a pass (D-11-21). The control's red cell is re-measured
     against a recorded `clang --version`; a toolchain change that extinguishes
     the divergence is an escalation, not a pass (D-11-22). The interprocedural
     `-O3`/LTO tier is proven non-inert, not assumed.

  4. **Gate (Pitfall 5):** a call-graph-shape reachability register records which
     cross-function shapes the generator actually reaches and **names shapes it
     provably does not**; HDD reducer output on multi-function programs is
     re-verified to reproduce the same property as its input.

  5. **Gate (Pitfall 6):** no interprocedural fact is marked cacheable until a
     callee-changes-invalidates-caller regression test gates it, and
     interprocedural cache keys derive from the call-graph closure, not per-unit
     hashes.

**Riskiest assumption**: that a `restrict`/no-alias promise emitted at a call
boundary survives Clang's inliner — that the fact `check` proved about a
parameter still describes the code after the call is inlined away.
**Gate that catches it being wrong**: criterion 3's engineered composition-only
negative control. A single clean `-O3` run with no negative control is explicitly
insufficient evidence here; the control must be seen to fail red first.
**Mid-phase gate (mandatory, second of the two highest-risk points)**:
multi-function C emission must be proven independently — a multi-function program
compiles, links, runs, and agrees with the interpreter at `-O0`, with zero alias
or capture attributes emitted — **before** call lowering with alias attributes is
admitted. This separates "new infrastructure works" from "optimizer promises are
sound", so a failure in the second is not masked by a defect in the first. Land
one shared `emitCall` helper, not five independent copies across `cgen`'s
switch sites.
**Parallel**: dev-backend timing probe runs alongside, non-blocking; it informs
M003 scope, not M002.
**Plans**: 9 plans (6 waves; tracer-first, with the mandatory mid-phase gate as a
ratification checkpoint between waves 3 and 4)

Plans:
**Wave 1**

- [x] 11-01-PLAN.md — Wave 0 spikes Q-01 (core-level `OpCall`→`OpCopy` rewrite) and Q-02 (live `cgen` cache hole); open `PHASE-11-DEBT.md`
- [x] 11-02-PLAN.md — NAT-07's hand-written-C 4×3 composition-only LTO control (Q-04), toolchain pin, and ratification of the criterion-3 amendments

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 11-03-PLAN.md — TRACER: `callgraph.EntryFunction` + `cgen.emitProgram`/`emitCall` + `session`'s three run sites; a two-function program runs on both engines at `-O0`

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 11-04-PLAN.md — NAT-05's explicit empty attribute set and the mid-phase gate (four-part conjunction, mutation-killed, adjudicated)

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 11-05-PLAN.md — NAT-06's four-tier interprocedural differential and the 32-guard disposition ledger
- [x] 11-06-PLAN.md — QLT-03's call-graph-shape reachability register and its drift protections

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 11-07-PLAN.md — QLT-06: `cgen` as the eighth declared cache input, plus the structural complete-by-abstention discharge
- [x] 11-08-PLAN.md — QLT-05 part 1: `reduce`'s multi-function `Seed`, two whole-program moves, derived bound, `RefusedShapes()`

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 11-09-PLAN.md — QLT-05 part 2: strict-field-equality re-verification, the foreign-call-sequence second knower, and the anti-vacuity gate

### Phase 12: `Result` Payloads

**Goal**: A `Result` can carry a payload that is stored, matched, moved out of,
and laid out identically in all three engines.
**Depends on**: Phase 11 at gate level (structurally independent at
implementation level); **`Result` payload / interprocedural-origin probe
answered** before planning
**Requirements**: RES-02, RES-03
**Maps to**: ARCHITECTURE Stage 9 / §7; closes D-04-30, deferred from M001
Phase 4
**Canonical refs**: `research/ARCHITECTURE.md` §7; D-04-30's own additive design
plan; Rust's tagged-union/niche RFC as precedent
**Success Criteria** (what must be TRUE):

  1. A `Result` value with a payload-carrying alternative is stored and matched,
     and moving out of a matched payload obeys the affine drop obligation — the
     unmoved alternative's payload is still dropped exactly once.

  2. `Result` layout — tagged union, with niche optimization applied **only**
     where a checked ability fact permits it — has one meaning in the core IR,
     the interpreter, and emitted C17, verified on the five-axis comparator
     rather than asserted per engine.

  3. **Gate (pre-flight probe):** payload origin and ownership reduce to
     already-proven Phase 08-11 machinery. If the probe surfaces a hidden need
     for a *new* interprocedural rule, this phase is replanned or moved to M003
     — it does not silently absorb new kernel rules.

**Riskiest assumption**: that `Result` payloads are structurally independent of
the call machinery. D-04-30's design plan predates M002's interprocedural work
and its independence is assumed, not confirmed.
**Gate that catches it being wrong**: criterion 3's pre-flight probe, run
**before** this phase is planned. If independence fails, the scope-cut order
already names this phase as the first cut — the response is an explicit slip to
M003, not extra plans.
**Scheduling note**: land `Result` payload `case` additions in the *same plans*
that already touch each dispatch site for `OpCall`, not as a second sweep over
the same six files. **Superseded by D-12-35** — `OpCall` landed across Phases
07-11, so there are no remaining `OpCall` plans to fold into. The note's
*purpose* (no second sweep, no two coexisting laws) still binds and is preserved
by measurement: planning measured that `emitProgram` cannot express a match body
at all, so the payload match surface is exclusively the legacy emitter family
this phase and the payload `case` arms are written exactly once.
**Plans**: 5 plans, in four waves. Plan 01 is explicitly-labeled inherited
D-11-02 debt-closure work serving NAT-04..NAT-07, not RES-* (D-12-34); plans
02-05 carry RES-02 and RES-03.

Plans:
**Wave 1**

- [ ] 12-01-PLAN.md — Inherited D-11-02 gate: the N=1 convergence differential, the D-12-36 branch decision, and PHASE-12-DEBT.md

**Wave 2** *(blocked on Wave 1 completion)*

- [ ] 12-02-PLAN.md — Tracer: a payload constructed, bound, and re-constructed end to end through parser, core IR, check, corevalidate, interp, and cgen, proven on three engines

**Wave 3** *(blocked on Wave 2 completion)*

- [ ] 12-03-PLAN.md — RES-02 expansion: the three named pattern refusals, D-12-27's resource-payload refusal, the affine drop obligation, and real exhaustive-dispatch coverage

**Wave 4** *(blocked on Wave 3 completion)*

- [ ] 12-04-PLAN.md — Independent peers, the corpus characterization replay proving the interp widening moved no bytes, and the D-12-04c header correction

**Wave 5** *(blocked on Wave 4 completion)*

- [ ] 12-05-PLAN.md — Criterion 2's two anti-vacuity controls, and the niche clause recorded as uninstantiable

### Phase 13: Agent Loop for Interprocedural Defects

**Goal**: An AI agent hitting a cross-function defect gets a bounded, correctly
attributed cause chain and can repair it through the JSON protocol alone.
**Depends on**: Phases 07-12 (the interprocedural defect taxonomy must exist
before it can be served)
**Requirements**: DX-05, DX-06, DX-07
**Maps to**: OPTIONS.md's final phase; PITFALLS.md Pitfall 7
**Canonical refs**: M001 Phase 6's `lang.explain/0` cause DAG, `diagnostic.Repair`
applicability, and DX-04's held-out/derivation-split fixture methodology
**Success Criteria** (what must be TRUE):

  1. `lang explain`'s cause DAG stays bounded when causes span functions and
     **names the function each cause step belongs to**, with the existing
     `truncated:explain.*` codes still stable.

  2. `lang-repair` reaches and fixes at least **three** new interprocedural
     defect classes through the JSON protocol alone, proven on a **held-out**
     fixture split — not on the fixtures the repairs were derived from.

  3. **Gate (Pitfall 7):** a repair-then-re-check regression covers cases where
     the **non-obvious** function is the correct fix location, and the repair
     applied there makes the program check clean.

**Riskiest assumption**: that blame lands on the function that violated its own
declared contract, rather than on the function where the error was detected —
a genuinely open design question with no concrete algorithm yet, only a
principle. rustc's own multi-year investment here is the cited precedent.
**Gate that catches it being wrong**: criterion 3. A diagnostic that fires at the
detection site looks correct in isolation and fails a repair-then-re-check test
immediately, because repairing the wrong function does not make the program
check clean.
**Scope note**: third scope-cut candidate. If Phases 07-12 overrun, narrow to the
two highest-value defect classes (cross-function borrow escape, cycle refusal)
rather than relaxing the mutation-kill or repair-then-re-check discipline for
whatever does ship.
**Plans**: TBD

## Progress

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 07. Calls, Signatures, and Call-Graph Refusal | 12/12 | Complete    | 2026-09-09 |
| 08. Interprocedural Loan Liveness in `check` | 6/6 | Complete    | 2026-09-10 |
| 09. Peer Re-Derivation and D-03-02 Closure | 10/10 | Complete    | 2026-09-10 |
| 10. Trusted Interprocedural Oracle | 9/9 | Complete    | 2026-09-11 |
| 11. Multi-Function Native Emission and Equivalence | 9/9 | Complete    | 2026-09-12 |
| 12. `Result` Payloads | 0/? | Not started | - |
| 13. Agent Loop for Interprocedural Defects | 0/? | Not started | - |

## Requirement Coverage

30 of 30 M002 requirements mapped to exactly one phase. No orphans, no
duplicates. Full traceability table lives in `REQUIREMENTS.md`.

| Phase | Requirements | Count |
|-------|--------------|-------|
| 07 | SEM-04, SEM-05, SEM-06, SEM-07, QLT-08 | 5 |
| 08 | OWN-06, EFF-02 | 2 |
| 09 | OWN-05a, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07 | 6 |
| 10 | SEM-08, SEM-09, TRU-02, TRU-03, QLT-04, OWN-05b | 6 |
| 11 | NAT-04, NAT-05, NAT-06, NAT-07, QLT-03, QLT-05, QLT-06 | 7 |
| 12 | RES-02, RES-03 | 2 |
| 13 | DX-05, DX-06, DX-07 | 3 |

QLT-08 (mutation-kill every new interprocedural control in the plan that
introduces it) is mapped to Phase 07 as the phase that establishes it, but is a
**standing discipline enforced in every phase 07-13**. It is listed once to keep
coverage unambiguous, not because it stops applying after Phase 07.

## Archived Milestones

<details>
<summary>✅ M001 Source-to-Native Semantic Spine (Phases 1-6) — SHIPPED 2026-09-07</summary>

- [x] Phase 1: Canonical Pure Spine (3/3 plans) — completed 2026-09-03
- [x] Phase 2: Owned Values and Abilities (7/7 plans) — completed 2026-09-03
- [x] Phase 3: Borrowed Views and CFG Lifetimes (10/10 plans) — completed 2026-09-04
- [x] Phase 4: Fallible Resources and C Boundary (13/13 plans) — completed 2026-09-05
- [x] Phase 5: Native Equivalence and Adversarial Evidence (14/14 plans) — completed 2026-09-06
- [x] Phase 6: Agent Feedback and Performance Ratification (15/15 plans) — completed 2026-09-07

Full phase details: [`milestones/M001-ROADMAP.md`](milestones/M001-ROADMAP.md)
Requirements (28/28 complete): [`milestones/M001-REQUIREMENTS.md`](milestones/M001-REQUIREMENTS.md)
Audit: [`milestones/M001-MILESTONE-AUDIT.md`](milestones/M001-MILESTONE-AUDIT.md)
Execution artifacts: `milestones/M001-phases/`

</details>

---
*Roadmap created: 2026-09-08 for milestone M002*
