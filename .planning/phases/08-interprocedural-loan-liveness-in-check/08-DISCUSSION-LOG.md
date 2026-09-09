# Phase 08: Interprocedural Loan Liveness in `check` - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-09
**Phase:** 08-interprocedural-loan-liveness-in-check
**Areas discussed:** Summary source & memo table, Fixpoint bound & its refusal, Refusal blame & disclosure, Cost gate shape (EFF-02)
**Mode:** advisor (4 parallel `gsd-advisor-researcher` agents, `minimal_decisive` calibration, `NON_TECHNICAL_OWNER = false` — `technical_background: true` overrides)
**Entry gate:** spike S-006 answered VALIDATED (2026-09-09) — hard entry gate on this phase's planning, released

---

## Gray area selection

The developer selected **all four** areas and attached their standing boilerplate
mandate: fan out across breadth and depth through every relevant
stakeholder-role lens, weigh pros/cons/tradeoffs/anti-patterns/best-practices/
footguns/lessons-learned, research online and draw on other products and
ecosystems, run an adversarial pass, then synthesize one perfect one-shot
recommendation per decision point. They characterized it as "my boilerplate
prompt saying research and synthesize best recommendation for everything."

Four `gsd-advisor-researcher` agents were spawned in parallel, one per area.
**Every load-bearing claim in the four returns was independently re-verified
against the shipped tree by the orchestrator before being locked.** Three claims
required correction; the corrections are recorded below and at the affected
decisions in CONTEXT.md.

---

## Summary source & memo table

| Option | Description | Selected |
|--------|-------------|----------|
| Design 1 — DECLARED | The summary table IS the already-shipped `callSignatureTable`; both liveness bits read from `/1`'s existing `ParameterContract.Mode`/`Drops` and `ReturnContract.Mode`/`Paths`/`Fresh`. No new derivation. Reads OWN-06 most literally. | |
| Design 2 — DERIVED | A new derivation pass over `callgraph.Order`'s RPO computing both bits from checked `core.Function` operations, matching spike S-006 literally. Textually re-walks callee bodies. | |
| Design 3 — HYBRID | Derive once per function in RPO, memoized within-run, carried on the existing in-`check` signature table alongside `Callable`; per-call admission consumes signature fields only. | ✓ |

**User's choice:** synthesized recommendation adopted (Design 3, hybrid).

**Notes:** Design 1 was **falsified, not merely disfavoured**, and that
falsification is what decides the area. `ReturnContract.Mode`/`Paths` does state
`ReturnsBorrowOfParam` declaratively, but **no `/1` field anywhere states
`UsesParam`** — `ParameterContract.Mode` is a declared *convention* (D-07-02),
not an observed body fact, and `core.FunctionSignature`'s own doc comment records
it "deliberately has no Linear/Match field at all." The only sound declared-only
fallback (assume `UsesParam = true` for every non-owned parameter) refuses the
`borrow; move; call` safe program — which is exactly the safe twin criterion 1
requires `check` to accept. A declared-only design fails the phase's own gate by
construction.

Design 3 avoids a `/1`→`/2` schema bump entirely, because `Callable` is *already*
a body-derived bit on `callSignatureTable` (its own doc comment records the table
is built after every function's body is checked). The two liveness bits are the
same shape of fact, so this extends a shipped pattern rather than opening a
one-way published-schema commitment.

The researcher's initial framing that this "textually brushes against OWN-06"
was tightened: D-07-34 **already recorded** the consumption-vs-production reading
for `Callable`, so D-08-05 restates an existing interpretation rather than taking
a new liberty — but it must be written down explicitly, since an undeclared
version of that reading is precisely what the debt register exists to prevent.

Also locked from the return: the mechanism must be a **forward canonicalization
pre-pass in `derivePlaceLoans`**, never a backward transfer clause (S-006
iteration 2 — the backward walk reaches the `move` before the call that
establishes the alias, so the refusal never fires and every accept-case test
still passes); and D-07-49's fix lands in **both** admission paths per D-07-21.

---

## Fixpoint bound & its refusal

| Option | Description | Selected |
|--------|-------------|----------|
| Derived internal-consistency assertion | Bound the intraprocedural worklist's existing `work` counter at a lattice-height-derived ceiling; provoke via an unexported seam; declare the source-unreachability as debt in D-07-47's shape. | ✓ |
| Fixed constant as a DoS ceiling | A flat iteration cap (e.g. 10,000) justified as future-proofing against a language version with loops. | |

**User's choice:** synthesized recommendation adopted (derived assertion + seam +
declared debt).

**Notes:** The decisive reframing came from the research rather than from the
question as posed: there is **no call-graph fixpoint to bound at all** (S-006
finding #2 — RPO is one pass), and summary derivation is two passes in program
order. The only genuine iterative fixpoint is the intraprocedural worklist, which
converges within lattice height by construction. So the bound is an **internal
consistency assertion against future refactors, not a DoS defense against hostile
input** — and criterion 2's "a program engineered to exceed that bound" is
**unsatisfiable from source in this language**. That is declared at planning time
per Key Lesson 4, in D-07-47's exact disposition, rather than discovered
mid-phase.

**Correction applied (C-1):** the researcher initially reported
`callReturnTypeDerivationSeam` as existing but was not able to cite it; a first
orchestrator grep with too narrow a pattern appeared to contradict them. Re-checked
and **confirmed present** at `check.go:493` (`var callReturnTypeDerivationSeam =
false`, read at `:1928`, flipped and restored at `check_test.go:3412`). The
researcher was right. It is unexported, which is what D-07-42 requires —
`pathoracle.TerminatorKindsOverride` (`pathoracle.go:51,59`) is **exported** and
is the shape D-07-42 cites as precedent while forbidding its multiplication.

Two further items were pulled into scope from the return rather than deferred:
converting `loanLivenessFixpoint`'s native-recursion pre-walk (`check.go:1157-1180`)
to an explicit stack — since `callgraph.Order` was deliberately made iterative for
the identical Pitfall-4 reason and shipping without it while citing D-07-18 would
be an inconsistency an auditor flags — and regularizing the bare
`check.cfg_back_edge` string into a coded diagnostic while in the same function.

Namespace resolved as `check.*`, not `core.*`, by inverting D-07-15's own
reasoning: Phase 09's peer uses a reachability closure, so the two mechanisms
**cannot fail the same way**, and a shared `core.*` code would assert an agreement
they are structurally incapable of having.

---

## Refusal blame & disclosure

| Option | Description | Selected |
|--------|-------------|----------|
| A — Mint a new `check.*` code | New `check.interprocedural_loan_liveness`, caller-blamed, fixed-role cross-function causes, non-repairable; `ownership.*` codes untouched this phase; promoted to `core.*` in Phase 09 with the peer. | ✓ |
| B — Reuse `ownership.*` verbatim | Distinguish the interprocedural case only via a new `Cause.Kind` inside the existing code's cause list. Satisfies "one law, not two" most literally. | |

**User's choice:** synthesized recommendation adopted (Option A).

**Notes:** Option B was rejected on the AI-agent-consumer lens, which this project
weights heaviest ("AI effectiveness depends more on precise verifier feedback than
exotic syntax"). `ownership.borrow_conflict` carries **local** repairs
(`narrow_to_shared_borrow`, `create_loan_after_conflicting_loan_ends`,
`check.go:2809-2825`); attaching a cross-function conflict to it would force
`lang-repair` to inspect `Causes[].Kind` before deciding whether to propose an
edit, degrading `Code` as a dispatch key and risking a repair to the wrong
function.

The cause chain was resolved as a **fixed three-role template**, which sidesteps
the D-07-16/D-07-43 determinism problem rather than re-solving it — role-keyed
positions are never selected from a candidate set, so there is nothing to rotate,
and Pitfall 7's unbounded-cause-DAG risk is answered structurally because the
chain never recurses into a middle relay's own causes.

**Correction applied (C-2):** the researcher's proposal to carry a callee
declaration span was checked and **cannot be built** — `diagnostic.Span` is
`{Start, End int}` (`diagnostic.go:17-20`) with no file or module field, and
`Cause.Span *Span` is likewise fileless. Widening it would be a one-way
published-schema change on research-only justification, against D-07-08. The
callee cause is therefore ID-only and spanless, which is already an established
shape in the same file (`{Kind: "loan", Detail: blocking.id}`, `check.go:2809`).

**Two gaps were surfaced and are declared rather than papered over:**

1. **Criterion 4's accepted-program half cannot be met by any shipped runtime
   artifact in Phase 08.** All three candidate vehicles were checked and fail:
   `protocol.ExplainSummary` is synthesized only from an *existing diagnostic's*
   causes, and there is no diagnostic on acceptance; `internal/compiler/cache` is
   structurally incapable of holding a verdict (the guard is literally named
   `TestCacheExportedSurfaceStoresNoVerdict`, `cache_test.go:120`); a sibling
   `lang.*/1` document contradicts `protocol.InterfaceSummary`'s own recorded
   anti-pattern against a second copy of the `/1` artifact. Disposition: a
   table-driven test as documentation-of-record, with the runtime-artifact cost
   question taken to the mid-phase gate.

2. **OWN-09's phase mapping conflicts with its own text** — the requirement says
   the intraprocedural law retires "in the same phase the interprocedural law
   lands" (Phase 08), while `REQUIREMENTS.md:173` maps it to Phase 09.
   **Flagged for the developer, deliberately not resolved by Claude.** Interim
   rule locked so planning is unblocked: `ownership.*` is not retired this phase,
   and the interprocedural check runs last so the two laws never overlap in scope.

The corpus section was kept uncompressed in CONTEXT.md because its per-shape
"or the gate is decorative" arguments are the direct analogue of Phase 07's
diamond insight — in particular, a twin pair that varies the *caller's* shape
rather than the *callee's declared contract field* would pass with the
interprocedural law entirely deleted.

---

## Cost gate shape (EFF-02)

| Option | Description | Selected |
|--------|-------------|----------|
| Deterministic work counter + fitted growth exponent, `elapsed_ns` observed-only | Hard gate on a machine-independent counter fitted against operation count across a size sweep; extends the shipped `recomputed_work`-hard / `elapsed_ns`-observed manifest split. | ✓ |
| Timing-only p50/p95/CoV as the hard gate | A literal reading of EFF-02's wording; zero new code, reuses `measure/statistics.go` exactly as built. | |

**User's choice:** synthesized recommendation adopted (deterministic counters +
fitted exponent).

**Notes:** The timing-only option was rejected as **structurally decorative**. The
roadmap states the failure mode "does not fail a correctness test; it fails a cost
curve," and S-006 states its instrument rule verbatim: wall-clock "is never used
to classify a mechanism." `measure/machine.go`'s own `CoVDemotionThreshold`
comment names the confounds (battery state, P/E core scheduling, thermal
throttling, single-host, no CI fleet) that would mask a super-linear curve at any
affordable corpus size. Same reasoning rustc used to gate perf regressions on
instruction counts rather than wall time.

Verified that `measure.Samples` is a bare `[]int64` and `Summary()` has no
nanosecond-specific logic, so applying the existing p50/p95/CoV protocol to work
counts satisfies EFF-02's wording **literally** and invents no parallel
instrument. That interpretation is recorded, because a timing-only reading of
EFF-02 is a live misreading.

The bound is fitted against **operation count, not function count** — S-006
records under Surprises that the `dense` shape fits 1.50 against function count
and 1.11 against operation count, and that "fitting only against function count
would have produced a wrong finding."

**Correction applied (C-3) — the highest-value finding in this area, and it
changes the plan.** The researcher proposed manifest rows named
`recomputed_work_per_op_exponent_*`. Verified that **two independent chokepoints
hardcode the literal string `"recomputed_work"`**:
`measure.Demote` (`statistics.go:135`, `if metric != "recomputed_work" { return
VerdictObserved }`, structurally guarded by
`TestDemoteHasExactlyOnePromotionPassthrough`) and
`session.QLT02GateEligibleMetrics()` (`session_phase6_budget.go:55-57`, returning
exactly `[]string{"recomputed_work"}`, the sole production argument at `:256`).
Any newly-named metric is therefore **silently demoted to `observed`** — a reviewer
reading only the manifest JSON would see `"gate_type": "hard"` and wrongly
conclude the gate is live. The researcher's related claim that this rule is
structural was also imprecise: `gateEligibleMetrics` is a caller-supplied
parameter to `AuditQLT02BudgetManifest`, and it is the *production caller* that
hardcodes the single name. Widening both chokepoints and re-deriving the
promotion-passthrough guard is now a named task with its own test obligation.

**Verified all-clear worth recording:** `ClosureDigest` is **not referenced
anywhere in `internal/compiler/cache`**, and `cache.Input` is constructed only in
tests. There is **no QLT-06 violation in flight** — `07-CONTEXT.md`'s line
describing `cache.Input` as taking `ClosureDigest` is a statement of planning
intent, not shipped wiring. The cross-run-cache non-goal pre-empts a future
wiring rather than describing a current defect.

---

## Claude's Discretion

The developer directed that the synthesized recommendation be adopted for all four
gray areas without per-area selection. Every decision in CONTEXT.md is therefore
Claude's synthesis under that standing instruction, grounded in the four advisor
returns and verified against the shipped tree wherever a claim was checkable.

Planner discretion explicitly retained over: plan decomposition and wave ordering
(subject to two hard ordering constraints — summary derivation after
`callgraph.Order` proves acyclicity, and the D-08-32 chokepoint widening before
any manifest row claims to be a hard gate); Go identifier and file names for the
summary bits, memo field, seam, and generator; whether the derivation is a method
on the existing table builder or a sibling function; corpus file names, module
paths, and the generated-shape size ladder; and whether the ratio-stability
tripwire ships as a manifest row or an in-test assertion.

## Deferred Ideas

- `corevalidate`'s independent liveness peer and D-03-02 closure — Phase 09.
- Promotion of `check.interprocedural_loan_liveness` to `core.*` — Phase 09.
- OWN-09's retirement of the intraprocedural law — Phase 09 per the requirements
  table, though OWN-09's own text says Phase 08. **Conflict flagged, on the
  mid-phase gate agenda.**
- Criterion 4's accepted-program disclosure as a runtime artifact — declared gap
  with a stated reason; cost question at the mid-phase gate.
- A persistent cross-run summary cache — Phase 11 / QLT-06, non-goal recorded now.
- Widening `diagnostic.Span` with a file/module field — not this phase.
- A third summary bit or per-arm bits — re-open at arity > 1 or `Result` payloads.
- Publishing the two liveness bits into `/1` or a `/2` — deliberately not done;
  the in-memory table is the cheap direction.
- The interpreter's fixed call-stack ceiling — SEM-08, Phase 10.
- Re-measuring cache invalidation fallout on production shapes — not worth the
  sweep cost; the spike's numbers stand.
- A `tree`-shaped cost corpus — **dropped as dominated**, not deferred.
- A match-arm call fixture for liveness — recommended regression coverage, not a
  gate requirement.
