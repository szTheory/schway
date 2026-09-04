# Phase 3: Borrowed Views and CFG Lifetimes - Context

**Gathered:** 2026-09-04
**Status:** Ready for planning
**Source:** developer direction at Phase 02 close-out, plus Phase 02 carry-forward

<domain>
## Phase Boundary

Local borrows remain ergonomic through precise last-use inference while public
borrowed results remain explicit and separately checkable. Requirements OWN-03 and
OWN-04.

This phase extends the same vertical source-to-native spine Phases 1 and 2 built. It
does **not** open new subsystems: no crash platform, no proof assistant, no symbol
server, no DWARF pipeline.

</domain>

<decisions>
## Implementation Decisions

### Debug-lineage seam (developer direction, explicit)

- **D-01:** Preserve the debug-lineage seam described in
  `wiki/debug-evidence-symbolication-and-proof.md`. Source → typed-core →
  lowering → native identity must remain joinable; do not foreclose it.
- **D-02:** Treat that note's "Next bounded experiment" as a **bounded experiment**,
  not a build-out. It is an admission test for whether the seam earns its cost —
  not a commitment to production crash capture.
- **D-03:** Explicitly out of scope this phase: DWARF/CodeView emission, crash
  storage/upload/retention, symbol servers, minidump capture, proof/SMT tiers.
  The wiki note's own guidance is to defer these until concrete workloads justify
  them; honour that.
- **D-04:** The experiment must be bounded the way every other Phase 2 lane is —
  declared caps, counted work, stable schema, timeout, output ceiling — and must
  report `available` / `optimized_out` / `not_captured` honestly rather than
  inventing a value. If it cannot be bounded, it does not ship this phase.

### Carried Phase 02 debt with a Phase 3 home

- **D-05:** D-02-03 (checker is Θ(N²) in body size after transitive loan sets) is
  *this phase's* problem, not an afterthought: OWN-03's CFG/edge-specific liveness
  should remove the quadratic factor by construction. Make `recomputed_work` count
  the propagation so the metric stops understating real cost.
- **D-06:** D-02-05 (`__LANG_` violates C17 §7.1.3 reserved identifiers) has a
  deadline, not a priority: rename to `_LANG_` in a standalone commit **before any
  further C artifact is frozen**. Every new golden widens the freeze.
- **D-07:** D-02-09 (`Box`/`Pair` type-check but die spanless exit 3 on every
  engine) is the same taxonomy smell that made CR-01 a Phase 2 blocker. Prefer a
  causal compile-time diagnostic over widening the C backend.
- **D-08:** The remaining debt (D-02-01 spawn-guard strength, D-02-02
  `evidence.canonical_unstable` unreachable at the CLI, D-02-04 missing
  `native.timeout` falsifier, D-02-06 8 MiB CLI ceiling, D-02-07 literal-line
  causality seam, D-02-08 `protocol.Human` divergence) is real but not
  phase-shaped. Fold each into whatever plan already touches its file; do not
  create a debt-cleanup plan.

### Method — non-negotiable, learned from three Phase 02 gate failures

- **D-09:** A differential test is not evidence until reverting the production hunk
  makes it fail. **Mutation-kill every oracle.** Phase 02's ownership oracle
  encoded the same wrong law as production and the differential agreed on the wrong
  answer.
- **D-10:** Interrogate what inputs a green property test actually *reaches* before
  trusting it. Two Phase 02 blockers hid behind generators whose reachable input
  space omitted the hard case — coverage was never the problem; the production code
  was fully executed both times.
- **D-11:** Drive behaviour through the shipped binary on hand-written programs,
  not only the gate's own corpus. The gate only ever sees what ships with it.
- **D-12:** Preserve the independence of the two admission layers (checker and
  `corevalidate`). Where a change must touch both, say so explicitly and record
  that the differential cannot cross-check that specific row (as OV-02-01 does).
- **D-12a (correction, from 03-PATTERNS):** "two admission layers" understates the
  blast radius of a new `OperationKind`. `interp.go:54-78` dispatches on all four
  kinds without authorizing any, and `cgen` lowers them — so adding e.g.
  `OpBorrowExclusive` requires **four** sites: `check`, `corevalidate`, `interp`,
  `cgen`. Miss `interp` or `cgen` and the operation silently vanishes from the
  execution trace that the O0/O3 differential compares, which is exactly the
  self-confirming shape Phase 2 spent a wave eliminating.

### Compatibility (inherited, still binding)

- **D-13:** Phase 1 goldens and schemas stay byte-identical. A Phase 2/3 golden may
  move only as a *causal* consequence of a deliberate semantic change, and the diff
  must be explained field-by-field.
- **D-14:** One global generated-C ordinary-identifier namespace; preferred
  source-derived names when unique, deterministic suffix only on real collision.
- **D-15:** Every input read and every spawned process stays bounded: deadline,
  independent stdout/stderr caps at max-plus-one, no `CombinedOutput`.

### Claude's Discretion

Plan decomposition, wave structure, how many plans, which fixture shapes exercise
CFG edges, and the concrete form of the debug-lineage experiment — provided it stays
bounded per D-04 and does not cross the D-03 scope fence.

</decisions>

<specifics>
## Specific Ideas

The Phase 2 close-out demonstrated that layered gates pay for themselves: each layer
found defects the previous layer missed, including one defect *introduced* by the
previous layer's fix. Keep the layering. Do not collapse review, verification,
validation and security into one pass to save time.

`.planning/spikes/002-cfg-edge-last-use/` is the direct prior art for OWN-03 and
`003-public-origins-generic-abilities/` for OWN-04. `004-independent-certificate-checker/`
is relevant to keeping the two admission layers independent.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase scope
- `.planning/ROADMAP.md` §"Phase 3" — goal, four success criteria, canonical refs
- `.planning/REQUIREMENTS.md` — OWN-03, OWN-04

### Prior art
- `.planning/spikes/002-cfg-edge-last-use/README.md` — edge-specific last use
- `.planning/spikes/003-public-origins-generic-abilities/README.md` — public origins
- `.planning/spikes/004-independent-certificate-checker/README.md` — independent checking
- `wiki/ownership-evidence-roadmap.md` — evidence trajectory

### Phase 2 carry-forward (read before planning)
- `.planning/phases/02-owned-values-and-abilities/02-DEBT.md` — nine accepted items,
  each with a Phase 3 remedy, plus the process debt in "Process debt"
- `.planning/phases/02-owned-values-and-abilities/02-OVERRIDES.md` — OV-02-01,
  `Buffer` grants `share`, and its three accepted consequences
- `.planning/phases/02-owned-values-and-abilities/02-VERIFICATION.md` — the 14 truths
  this phase must not regress
- `.planning/phases/02-owned-values-and-abilities/02-VALIDATION.md` — the sampling
  contract and its two recorded residual weaknesses

### Debug-lineage seam
- `wiki/debug-evidence-symbolication-and-proof.md` — the seam to preserve and the
  bounded experiment to scope. Note its own "Proof posture" section: do **not** turn
  the language into a proof assistant in v1.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `internal/compiler/check/check.go` — straight-line loan liveness with transitive
  loan sets; the base CFG liveness replaces. Already has counted work.
- `internal/compiler/corevalidate/corevalidate.go` — the independent admission layer;
  has its own transitive `loansForPlace`. Keep independent.
- `internal/compiler/ability/ability.go` — sealed five-ability derivation; OWN-04's
  public origins will need access modes alongside abilities.
- `scripts/verify-phase2.sh` + `scripts/assert-go-tests.sh` — the bounded gate and
  fail-closed exact-target selector. Phase 3 should extend, not fork, this pattern.
- `internal/compiler/session/session.go` — control table, lane work, and the
  fail-closed required-control set (now nine).

### Established Patterns
- Negative controls are exact IDs required fail-closed by the gate, with nonzero work.
- Semantic identity uses function-local ordinals, never source offsets — the
  debug-lineage seam must follow this, not reintroduce byte offsets as identity.
- Evidence claims content identity only; never overclaim translation proof.

### Integration Points
- Any new lane joins `verify-phase2.sh`'s control set and must carry nonzero work.
- Any new diagnostic joins the existing repair-bearing `lang.diagnostic/1` taxonomy.

</code_context>

<deferred>
## Deferred Ideas

- Full DWARF/CodeView emission, symbol bundles, crash capture/upload/retention,
  minidumps (D-03) — deferred until a concrete workload justifies the cost.
- Proof/SMT/refinement tiers (D-03) — deferred per the wiki note's own guidance.
- D2 dogfooding (a Lang-written corpus/evidence runner) — the compiler and harness
  remain Go; noted in Phase 02's decisions, not this phase's scope.

</deferred>
