# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: M001 — Source-to-Native Semantic Spine

**Shipped:** 2026-09-07
**Phases:** 6 | **Plans:** 62 | **Tasks:** 174 | **Commits:** 335 | **Timeline:** 5 days (2026-09-03 → 2026-09-07)
**Code:** ~59,150 lines Go, ~916 lines C/H, zero external dependencies

### What Was Built

- **A real vertical compiler**, not a set of layers: lossless canonical frontend
  → typed core with per-function CFG → affine ownership and borrow checker →
  independent re-derivation at two trust crossings (`corevalidate`,
  `originvalidate`) → deterministic interpreter oracle → readable C17 through
  Clang at `-O0`/`-O3`/`-flto`.
- **Ownership and borrowing that survive native lowering**: affine transfer,
  independently derived copy/drop/share/send/escape abilities, shared/exclusive
  loans decided by a CFG liveness fixpoint, and declared public borrow origins
  verified against the core rather than trusted.
- **Fallible resources and a real C boundary**: three-stage acquisition with
  exact reverse-order partial cleanup, one authoritative `core.ForeignContract`
  generating three lockstep-derived inspectable layers, a setjmp/longjmp
  nonlocal-exit landing pad, and `nm -u` symbol gating.
- **Adversarial native evidence**: a five-axis interpreter/`-O0`/`-O3`
  comparator, `restrict` emitted only on checker-proven no-alias parameters
  (with an engineered `false_no_alias` divergence as its negative control), an
  always-on ASan+UBSan lane with real detectable subjects, and a core-level HDD
  reducer with a strict interestingness predicate.
- **An agent-facing feedback surface**: `lang explain`'s bounded cause DAG,
  rustfix-style repair applicability, a verdict-incapable artifact cache,
  changed-risk lane selection over a 38-row declared substrate, and a measured
  p50/p95/CoV budget manifest.
- **`cmd/lang-repair`**: a standalone repair driver for five defect classes that
  talks to the shipped binary through `--json check` alone, with the
  protocol-only boundary enforced by a build-failing lint.

### What Worked

- **Vertical slices caught cross-layer bugs that layered construction would have
  hidden until integration.** 03-02's own audit surfaced a fifth consumer
  (native's execution-document validator) nobody had listed. Phase 4's
  three-engine differential found two real `lang run --engine=native` bugs.
- **Independent re-derivation as an architectural rule, not a testing tactic.**
  `corevalidate` deciding loan endpoints by reachability-closure-plus-reduction
  while `check` uses an iterative worklist is the reason agreement between them
  means something. The same discipline gave `originvalidate` its authority.
- **Mutation-killing every differential.** A control that has never been made to
  fail is a claim, not evidence. Phase 5's `control:alias.false_no_alias`
  becoming a real engineered subject — an actual `interpreter == -O0 (2) != -O3
  (7)` divergence on Apple clang 21 — is the shape this should always take.
- **Shadow-run before delete.** 05-02 ran `loanLivenessFixpoint` against
  `discoverLoanLastUses` over 230,692 admission-site comparisons with zero
  divergences before deleting the old law. That is how you retire a derivation.
- **Recording debt as structured, machine-checkable registers.** `*-DEBT.md`
  with a closed severity vocabulary, an identifier, a source, and a landing
  phase — plus a test asserting the shape — kept deferrals honest across six
  phases.
- **Deferring `OpCall` rather than landing it under load.** The call to keep it
  out of Phase 5 was made explicitly, with the reasoning written down, and it
  was correct for the risk budget.

### What Was Inefficient

- **Three gate failures shared one shape**: a green test whose reachable input
  space omitted the hard case. Phase 2 cost a remediation round, Phase 3 a
  mid-phase gate plus three gap-closure plans, Phase 4 thirteen plans and five
  review rounds. The lesson was learned three times before it became a rule.
- **Two liveness derivations coexisted in the checker from 03-03 until 05-02** —
  roughly two phases of carrying a known duplicate law because the phase that
  introduced the better one scoped itself to arm bodies only.
- **A quadratic cost fix landed in the wrong half.** D-02-03's validator half
  closed in 03-04; D-03-01 recorded that the fix never reached the
  admission-deciding path. Fixing one of two peers is not fixing the item.
- **`originvalidate.RecomputeOrigin` needed three corrective plans** (03-08,
  03-09, 03-10) after its first landing: a first-seen guard, a
  publish-every-function rule, and a per-return conservative combiner. The
  original derivation looked at one return and one arm.
- **Phase counts drifted upward under review pressure** — 3, 7, 10, 13, 14, 15
  plans. Later phases absorbed gap-closure work as plan count rather than as
  scope renegotiation.
- **Nyquist validation fell behind.** Phases 1, 2, and 4 are compliant; 3, 5,
  and 6 are not-validated. Validation ran when it was convenient, not when the
  phase closed.

### Patterns Established

- **Independent re-derivation at every trust crossing.** The producer's claim is
  never the consumer's evidence. Shared inert records only.
- **Every differential gets a mutation-kill.** Declare the control, then break
  the thing it watches and prove the control goes red.
- **Interrogate what a property test actually reaches**, not what it nominally
  covers — see D-04-21's Generator and Probe Reachability Register, where each
  row must name a shape the generator does *not* reach.
- **Drive the shipped binary on hand-written programs**, not only the gate's own
  corpus. Out-of-corpus proofs caught what in-corpus tests could not.
- **Schema evolution by additive fields with frozen prior bytes.** The `/0`→`/1`
  bump landed across 12 literal sites with `/0` bytes provably unchanged, and
  identity exclusion is asserted by reflection tests that self-invalidate.
- **Escapes are declared and executable, not hypothetical.**
  `escape:coordinated-source-to-core-false-claim` ships as a runnable pair of
  artifacts that each pass their own validator while contradicting each other.
- **Fail closed on ambiguity.** The cache resolves to `not_cacheable`; the lane
  selector refuses drift; graph walks carry visited-set guards and named
  refusal codes rather than hanging.

### Key Lessons

1. **A control you have never seen fail is a claim.** Budget the mutation-kill
   into the plan that introduces the control, not into a later gate.
2. **When a phase introduces a better derivation, retiring the old one is part
   of that phase**, not a follow-up. Two coexisting laws is a defect with a
   delayed fuse.
3. **Fixing one of N independent peers is not fixing the item.** The peer
   structure that makes re-derivation valuable also means every fix has N sites.
4. **Deferring a large structural change is legitimate and should be written
   down at the moment it is decided**, with the consequence stated plainly.
   "M001 ships without Lang-to-Lang calls" cost nothing to say and would have
   cost a great deal to discover at close.
5. **Validation that is not gated slips.** Nyquist compliance tracked exactly
   the phases where it was a gate condition.
6. **Archival is a code change.** Moving `.planning/phases/` at milestone close
   broke two tests that hardcoded the path — and one of them would have failed
   *open*, silently passing on an empty glob.

### Cost Observations

- Model mix: not instrumented this milestone.
- Sessions: not instrumented; 335 commits across 5 calendar days.
- Notable: recorded per-plan durations rose sharply from Phase 2 (~10 min/plan)
  to Phase 3 (70-95 min/plan) as plans moved from single-consumer changes to
  five-consumer coordinated changes. Phase 3 is where the cost of the peer
  architecture became visible — and where it started paying.

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Commits | Phases | Plans | Key Change |
|-----------|---------|--------|-------|------------|
| M001 | 335 | 6 | 62 | Established vertical slicing, independent re-derivation at trust crossings, and mutation-killed controls as standing rules |

### Cumulative Quality

| Milestone | Requirements | Phases Verified | Nyquist | External Deps |
|-----------|--------------|-----------------|---------|---------------|
| M001 | 28/28 | 6/6 | 3/6 (partial) | 0 |

### Top Lessons (Verified Across Milestones)

1. *(Awaiting a second milestone to cross-validate M001's lessons.)*
