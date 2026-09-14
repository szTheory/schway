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

## Milestone: M002 — Interprocedural Semantic Spine

**Shipped:** 2026-09-14
**Phases:** 7 (07-13) | **Plans:** 61 | **Tasks:** 183 | **Commits:** 386 | **Timeline:** 6 days (2026-09-08 → 2026-09-14)
**Code:** ~102,280 lines Go across 25 packages, 61 new `.lang` fixtures, zero external dependencies
**Close:** `override_closeout` — 29/31 requirements satisfied, 2 ratified partial, 3 deferred items acknowledged

### What Was Built

M002 added exactly one new `core.OperationKind` — `OpCall` — and then spent
seven phases proving that every semantic guarantee M001 established
intraprocedurally still holds across a function boundary.

- **Phase 07** made `OpCall` real at all six dispatch sites with a digest-bound
  `lang.interface/1` signature summary, a `callgraph` package whose iterative
  three-color DFS refuses every cycle by name, and `Callable` defined as
  publication safety rather than export membership.
- **Phase 08** derived interprocedural loan liveness in `check` from callee
  signatures alone, under a derived, block-count-scaled, fail-closed iteration
  bound and a measured cost curve ratified into the feedback-budget manifest.
- **Phase 09** had `corevalidate` re-derive the same fact by a structurally
  opposite walk (forward set-propagation vs `check`'s backward memoized walk),
  closed **D-03-02** — the one debt item knowingly carried past M001 — and
  deleted `computeLoanLastUses` only after a second detector existed.
- **Phase 10** extended `originvalidate` and `pathoracle` across `OpCall`
  without either importing `check` or `corevalidate`, and gave the interpreter a
  bounded call stack with a documented ceiling.
- **Phase 11** emitted multi-function C17 and got interpreter, `-O0`, `-O3`, and
  `-O3 -flto` to agree on five axes — the claim M001 structurally could not make.
- **Phase 12** shipped payload-carrying `Result` alternatives, storable,
  matchable, and affine-correct in all three engines (closing M001's D-04-30).
- **Phase 13** took `lang explain`'s cause DAG across function boundaries and
  taught `lang-repair` two new interprocedural repair classes, reached through
  the JSON protocol alone on a sealed held-out split.

### What Worked

- **Cutting phases where a gate becomes meaningful, not where implementation
  could parallelize.** All seven phases verified passed, and each gate caught
  something real: Phase 09's requirement-vs-traceability defect, Phase 10's
  fixture escape hatch, Phase 12's one and only blocker.
- **Splitting OWN-05 into OWN-05a/OWN-05b (D-09-37)** rather than flipping one
  row Complete on two of three derivers. This is the direct reason the milestone
  audit found zero orphaned or overclaimed requirements across three independent
  sources.
- **Eliminating human-judgment items with tests instead of adjudicating them**
  (Phase 11). `TestSeedEntryHazardIsReal` and `session_admission_divergence_test.go`
  both fail in *both* directions — catching a silently resolved divergence as
  well as a new one. A debt note is read once; a test runs on every CI
  invocation.
- **Escape hatches that report rather than fall back.** Plan 10-07's fixture
  instruction told the executor to report an obstruction rather than silently
  substitute a weaker fixture. That is how the `peerDeriveOriginFacts` `OpCall`
  hole surfaced — with a full empirical trail, including the discovery that an
  untouched Phase 08 fixture had been refused for the same hidden reason all
  along.
- **Reporting what you find instead of quietly fixing it.** Phase 13 chose the
  stricter retro-strengthening branch and discovered that *M001's* held-out and
  derivation fixtures were alpha-renames of each other (D-13-34) — a hole in
  already-shipped evidence that no M002 requirement would have exposed.

### What Was Inefficient

- **The single-function emitter deletion was deferred twice** (D-11-02 →
  D-12-36) despite Phase 10 writing D-10-60, a no-third-deferral rule, precisely
  to catch this pattern. Each deferral was individually defensible; the pattern
  is the problem, and it now crosses a milestone boundary unowned.
- **A dependency chain shipped with no owner.** D-12-21 cannot close until
  D-11-51 does, and neither has a phase. A debt item that blocks another debt
  item deserves an owner at the moment the dependency is recorded, not at close.
- **A flagged human review never happened.** Plan 10-02's D-09-51 negative-control
  verdict flip was documented and explicitly flagged — and then sat unreviewed
  through three more phases (D-10-C04, D-11-27). Flagging is not routing.
- **Nyquist validation slipped again**, in exactly the way M001's Lesson 5
  predicted: compliant for the two phases where it was a gate condition (09, 10),
  not-validated for the five where it was not. Phase 13's substance was actually
  there — plan 13-07 re-ran every verification-map row and found two ungrounded
  `-run` patterns — but the lifecycle marker stayed `draft` because
  `validate-phase` never ran as a skill.
- **Two of three DX-07 repair classes and all of DX-06 were blocked by one
  unexamined invariant.** `sameType(ReturnType, Parameter.Type)` at function
  admission was not identified as a requirement-level constraint until Phase 13
  was already building against it. The blame resolver and its exhaustiveness
  guard are now built, tested, and unreachable.

### Patterns Established

- **Structural opposition as the independence proof.** `check` walks backward
  and memoizes; `corevalidate` propagates forward over its existing postorder.
  Agreement between two algorithms chosen to be structurally opposite is worth
  more than agreement between two implementations of the same idea — and it is
  falsifiable, which M002 proved with bidirectional seeded faults.
- **Transitive-dependency guards as mechanism, not review.** `go list -deps` on
  each peer is a real check that peers do not import each other. M002's own
  10-REVIEW.md WR-02 records where this is still hand-curated and therefore
  weaker than it looks.
- **Requirement splitting as an anti-overclaim device.** When a requirement
  names N derivers and a phase ships M < N, split the row.
- **Escalate a defect in the criterion rather than downgrade the assertion.**
  D-12-43 and D-13-10a both took a criterion that could not be met, proved
  empirically why, and recorded it as a terminal finding at a blocking
  checkpoint — instead of rewriting the criterion to match what shipped.
- **Reject a checker's conclusions where the tree disagrees.** The integration
  checker's structural findings were adopted; its requirement-satisfaction
  column was rejected against a `grep`.

### Key Lessons

1. **An integration checker that grades requirements from wiring will convert an
   honest partial into a false green.** Wiring is exactly what a structurally
   unreachable defect class still has. Grade requirements against the tree.
2. **A rule you write to prevent a pattern does not prevent it.** D-10-60 was
   written in Phase 10 specifically to stop a third deferral, and Phase 12
   deferred anyway. Rules need an owner and a trigger, not just a statement.
3. **"Flagged for human review" is not a work item.** If a flip, a deviation, or
   a residual risk needs a human, it needs a phase — otherwise it accumulates.
4. **One invariant can silently define what a whole phase can prove.** Identify
   the load-bearing type-system invariants *before* planning a phase whose
   requirements depend on contradicting them.
5. **M001 Lesson 5 reconfirmed: validation that is not gated slips.** Two
   milestones, same result, same mechanism. Make it a gate or expect it not to
   happen.
6. **An escape hatch that reports beats a fallback that substitutes.** Every
   genuinely new finding in Phase 10 came through an instruction to report
   obstruction rather than quietly weaken the fixture.

### Cost Observations

- Model mix: not instrumented this milestone.
- Sessions: not instrumented; 386 commits across 6 calendar days.
- Notable: 61 plans in 6 days versus M001's 62 in 5 — per-plan cost did not fall
  despite the milestone adding exactly one operation kind. The peer architecture
  that M001's Phase 3 made visible is now the steady-state cost: every semantic
  change lands at six dispatch sites plus two exhaustive-dispatch controls, which
  is the project's own D-04-22 standard for "real."

---

## Cross-Milestone Trends

### Process Evolution

| Milestone | Commits | Phases | Plans | Key Change |
|-----------|---------|--------|-------|------------|
| M001 | 335 | 6 | 62 | Established vertical slicing, independent re-derivation at trust crossings, and mutation-killed controls as standing rules |
| M002 | 386 | 7 | 61 | Added gate-driven phase cuts, requirement splitting against overclaim, and "escalate a defect in the criterion" over downgrading an assertion |

### Cumulative Quality

| Milestone | Requirements | Phases Verified | Nyquist | External Deps |
|-----------|--------------|-----------------|---------|---------------|
| M001 | 28/28 | 6/6 | 3/6 (partial) | 0 |
| M002 | 29/31 (2 ratified partial) | 7/7 | 2/7 (partial) | 0 |

### Top Lessons (Verified Across Milestones)

1. **Validation that is not gated slips.** M001 Lesson 5, reconfirmed exactly in
   M002: compliant for the phases where Nyquist was a gate condition, not
   validated for the rest. Two milestones, same mechanism. Either gate it or
   stop counting on it.
2. **A control you have never seen fail is a claim.** M001 Lesson 1 held through
   M002 as standing discipline (QLT-08), and the two places M002 nearly shipped
   a vacuous control (WR-02's payload-slot swap, plan 11-08's positional-match
   relaxation) were both caught by applying it.
3. **Write the consequence down at the moment you decide to defer.** M001's
   "ships without Lang-to-Lang calls" cost nothing to say and framed all of
   M002. M002's inverse case — D-11-02 deferred twice against its own D-10-60
   rule — shows that writing it down is necessary but not sufficient: a deferral
   also needs an owner.
4. **Fixing one of N independent peers is not fixing the item.** M001 Lesson 3,
   reconfirmed: `peerDeriveOriginFacts`'s missing `OpCall` case is exactly this
   shape, one phase after D-10-28 found the same shape in the same package's
   other peer.
5. **Archival is a code change.** M001 Lesson 6 — worth re-checking after every
   milestone close, since phase directories move again each time.
