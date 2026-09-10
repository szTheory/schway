---
phase: 08-interprocedural-loan-liveness-in-check
recorded: 2026-09-09
status: accepted
disposition: gate-adjudicated
items: 9
blocking: 0
---

# Phase 08 — Declared Deferred Scope (D-08-39)

**Written:** 2026-09-09, at *planning* time — not at phase end.

Per RETROSPECTIVE Key Lesson 4: declare deferred scope in writing at the moment
it is decided. Shape follows the mechanically-checked debt-register format
`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`)
enforces on every `*-DEBT.md` register: an `items:` count matching the `## Items`
table, one `### <ID>` detail section per row, and a severity from the closed
vocabulary (blocker, warning, info).

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-08-15 | 08-CONTEXT.md (D-08-15), 08-03-PLAN.md | OWN-06, QLT-08 | info | Not scheduled — reopen when the language gains iteration, collections, or arity > 1 | The fail-closed loan-liveness iteration bound and its named refusal are UNREACHABLE from any legal `.lang` source program: the intraprocedural worklist is a monotone fixpoint over a finite lattice, and the language has no loops, no iteration, no collections, and arity 1, so no program can force divergence. The control's only living witness is the unexported `loanLivenessBoundSeam`, which mutation-kills it |
| D-08-26 | 08-CONTEXT.md (D-08-26), 08-03-PLAN.md | OWN-06 (success criterion 4) | warning | Phase 09 — Peer Re-Derivation and D-03-02 Closure (reviewed and carried at Phase 08's mid-phase gate, 08-06, 2026-09-10) | Success criterion 4's ACCEPTED-program half ships no runtime artifact this phase. On refusal the disclosure is real (cause 3 names the one consulted callee-signature field); on acceptance no shipped vehicle can carry it — `protocol.ExplainSummary` is synthesized only from an existing diagnostic's causes, `internal/compiler/cache` is structurally incapable of holding a verdict (`TestCacheExportedSurfaceStoresNoVerdict`), and a sibling `lang.*/1` document contradicts `protocol.InterfaceSummary`'s own recorded anti-pattern. Satisfied this phase by a table-driven test asserting the consulted field set, not by a shipped artifact |
| D-08-27 | 08-CONTEXT.md (D-08-27), REQUIREMENTS.md:173 vs OWN-09's own text | OWN-09 | warning | Phase 09 — Peer Re-Derivation and D-03-02 Closure (conflict reviewed and carried, unresolved, at Phase 08's mid-phase gate, 08-06, 2026-09-10) | The project's own documents disagree about when the intraprocedural loan-liveness law retires: OWN-09's text says "retired in the same phase the interprocedural law lands" (Phase 08); `REQUIREMENTS.md:173` maps OWN-09 to Phase 09. Only a human decision closes this. Interim rule locked for Phase 08: `ownership.*` codes are NOT retired, and the interprocedural check runs LAST, after intraprocedural admission has already passed, so no program is judged by both laws for the same fact |
| D-08-37 | 08-CONTEXT.md (D-08-37), spike S-006 iteration 6 | QLT-06 | info | Phase 11 — Multi-Function Native Emission and Equivalence | NON-GOAL (Phase 08): a persistent CROSS-RUN summary cache is explicitly deferred. Spike S-006 measured that a single leaf edit invalidates 92% of a call-graph-closure-keyed cache worst-case, 43% mean on a realistic parser-shaped corpus, and 100% on a chain. Within-run memoization is mandatory and sufficient for EFF-02; cross-run caching is a separate, unproven claim that must not be assumed or quoted as a production win until QLT-06's callee-changes-invalidates-caller regression test gates it |
| D-08-38 | 08-CONTEXT.md (D-08-38), ROADMAP.md M002 scope-cut order | EFF-02 | info | Resolved at Phase 08's mid-phase gate (08-06, 2026-09-10) — not triggered | Declared scope-cut trigger: if the summary-table + liveness-law work exceeds ~2x its initial plan estimate, the cost-gate INSTRUMENT work (chokepoint widening, corpus generator, risk lane, manifest row) renegotiates into Phase 09 — never the criterion-1 corpus, never the seeded mutation-kills, never the two-path D-07-49 fix. If it is cut, the liveness law may NOT be declared final until it lands |
| D-08-40 | 08-06-PLAN.md (mid-phase gate agenda a-adjacent finding), session_peer_gate_test.go `peerDivergenceExpected` | OWN-06, OWN-07 | warning | Phase 09 — Peer Re-Derivation and D-03-02 Closure | The mid-phase gate formally registers, as tracked debt rather than a SUMMARY-only note, the two accepted-program fixtures (`twin_a_accept.lang`, `relay_depth2_accept.lang`) whose check-admits/corevalidate-refuses divergence is already live in session's own `peerDivergenceExpected` map (07-10's mechanism): `check` correctly admits per the new interprocedural liveness law; `corevalidate`'s still-intraprocedural `loanChainIndex` refuses both via `core.move_while_borrowed` (unconditional `parent[TargetID] = SourceID` propagation through every `OpCall`). Concrete, fixture-backed input for whichever Phase 09 plan extends `corevalidate`'s own loan-liveness re-derivation to consult callee signatures the same way `check` now does |
| D-08-41 | 08-03-SUMMARY.md ("Plan-text vs. verified-reality notes", item B) | OWN-06 (success criterion 1) | info | Not scheduled — documented scope limit, already satisfied at the checked-core level | Pattern B's real `.lang` twin pair (`twin_b_refuse.lang`/`twin_b_accept.lang`) cannot demonstrate a differing END-TO-END CLI verdict: `computeLoanLastUses`' summary-blind AST-shadow admission path refuses BOTH members identically (`ownership.move_while_borrowed`) at the INTRAPROCEDURAL layer, before `check`'s interprocedural pass ever runs. The contract-driven backward gate itself is proven only at the checked-core level (08-02's own `TestInterproceduralLivenessTwinPatternB`, a synthetic `core.Program` construction). The mid-phase gate adjudicates this as a genuine, disclosed, permanent scope limitation of this phase's real-fixture corpus for Pattern B — not a defect, and not a to-do |
| D-08-42 | 08-CONTEXT.md (D-08-03), 08-01-PLAN.md's original doc comment | OWN-06 | info | Not scheduled — planning-document correction only, no production-code defect | 08-CONTEXT.md's D-08-03 and 08-01's own original doc comment both assert `callgraph.Order`'s reverse postorder is CALLEE-before-CALLER. 08-02 established empirically (`TestOrderSortsAdjacencyByCalleeID` plus this package's own `TestSummaryDerivationIsOnePassPerFunction`) that it is actually CALLER-before-CALLEE, and `buildInterproceduralSummaries`' own doc comment (`check.go`) already documents the correction and walks the order backward to get a true callee-before-caller derivation. The production code is correct; the two planning documents are stale and are recorded here rather than silently left to mislead a future reader |
| D-08-43 | `internal/compiler/session/session_test.go` (`TestQLT01RegistryCoversAllFiveSpikes`) | QLT-01 | info | Not scheduled — pre-existing `.planning/spikes` registry data gap, unrelated to Phase 08, verified present before this phase began | `TestQLT01RegistryCoversAllFiveSpikes` fails with "spike 006 has a directory under .planning/spikes but no registry row cites it." A pre-existing registry-maintenance gap in `.planning/spikes`, not caused by or in scope for Phase 08; recorded here so the one non-Phase-08 test failure surfaced by `go test ./...` during this phase is traceable rather than silently tolerated |

## Detail

### D-08-15 — the iteration bound is unreachable from source

The bound added to `loanLivenessFixpoint` is an **internal-consistency
assertion, not a DoS defense**. A monotone transfer function over a finite
lattice of live-loan sets provably converges within lattice-height iterations.
No `.lang` program can force divergence; only an implementation bug can — a
non-monotone transfer function, a mutated lattice, or a lost `queued` flag
reintroducing infinite reinsertion.

`.planning/LANGUAGE-MATURITY.md` records what the language can express today:
no arithmetic, no iteration, no collections, arity 1, A-normal form, 58 programs
averaging ~28 lines. Success criterion 2's phrase *"a program engineered to
exceed that bound"* is therefore **unsatisfiable from source in this language**.

This is the exact disposition D-07-47 already established for
`check.call_return_type_unrepresentable`. The control ships with a living
witness — the unexported `loanLivenessBoundSeam` package-level var
(`internal/compiler/check/check.go`), false in production, flipped and
deferred-restored by `TestLoanLivenessBoundMutationKilled`
(`internal/compiler/check/check_test.go`), which asserts BOTH directions: the
named `check.loan_liveness_bound_exceeded` refusal fires with the seam up, and
the same otherwise-clean input admits cleanly with the seam down — so it is
never a control that has never been seen to fail.

**Reopen when:** the language gains iteration, loops, collections, or arity
past 1 — any of which makes the lattice height program-controlled and the
refusal potentially source-reachable.

### D-08-26 — success criterion 4's accepted-program half

**On refusal, criterion 4 is met by a shipped artifact.** The
`check.interprocedural_loan_liveness` diagnostic's third cause names the one
specific callee-signature field the answer depended on —
`<calleeID>:return.mode=<Mode>` or `<calleeID>:parameters[0].mode=<Mode>` —
never a bitmask, never "all of them". The liveness law consults exactly one bit
per direction, so naming one field is both honest and sufficient. Body-blindness
holds: `Mode` is copied verbatim from the declared AST type, so disclosing it
reveals nothing about the callee's control flow.

**On acceptance, no shipped runtime artifact carries the disclosure this
phase.** All three candidate vehicles were checked against the tree and each
fails for a structural reason:

- `protocol.ExplainSummary` (`internal/compiler/protocol/protocol.go:223-249`)
  is synthesized from an **existing diagnostic's** flat `Causes` list. An
  accepted program has no diagnostic, so `lang explain` has nothing to expand.
- `internal/compiler/cache` is structurally incapable of holding a verdict —
  the guard is named literally, `TestCacheExportedSurfaceStoresNoVerdict`
  (`internal/compiler/cache/cache_test.go:120`).
- A new sibling `lang.*/1` document contradicts `protocol.InterfaceSummary`'s
  own recorded anti-pattern (`protocol.go:149-152`): mirroring `/1` fields
  "would create a second schema that can drift from the first with no producer
  forcing them together".

**Disposition this phase:** satisfied by a table-driven test that asserts, for
each corpus member, the exact set of callee-signature fields the derivation
consulted — machine-checked, not human-readable output. Whether the accepted
case should become a runtime artifact, and what it costs, was on the mid-phase
gate's agenda.

**2026-09-10 — reviewed and carried at Phase 08's mid-phase gate (08-06).**
Commissioning a runtime vehicle now would be a Rule 4 architectural change (a
new schema, a new protocol surface, or a new cache field) requiring its own
dedicated planning, not something an adjudication gate decides by itself. The
gate elects to leave this as declared debt landing in Phase 09, alongside
`corevalidate`'s own peer re-derivation work — the same phase that will need
to reason about what the peer discloses on acceptance too, so the two
disclosure questions are naturally answered together rather than the accepted
half being solved in isolation now.

### D-08-27 — OWN-09's Phase-08-vs-09 document conflict

**OWN-09's own text** says the intraprocedural law is *"retired in the same
phase the interprocedural law lands. One law, not two"* — which is Phase 08.
**`REQUIREMENTS.md:173`** maps OWN-09 to **Phase 09**. The two cannot both be
right and only a human decision closes it.

**Interim rule locked for Phase 08 planning and execution:**

1. `ownership.*` loan codes are **not retired** in this phase.
2. The interprocedural check runs **last** — a separate pass over the completed
   `core.Program`, after `callgraph.Order` has proven acyclicity and after
   every function's intraprocedural admission has already passed. The two laws
   therefore never overlap in scope and no program is judged by both for the
   same fact.
3. Retirement lands in Phase 09 alongside `corevalidate`'s independent peer.

Retiring the sole law **before** a second independent detector exists to
validate the interprocedural law's completeness would violate RETROSPECTIVE Key
Lesson 2 directly (two coexisting laws, and the one that gets deleted is the one
with coverage). This item was on the mid-phase gate's agenda and is recorded as
a conflict, not as a resolution.

**2026-09-10 — reviewed and carried, unresolved, at Phase 08's mid-phase gate
(08-06).** Auto-mode adjudication does not resolve a genuine document conflict
by assumption — the plan's own agenda item names this "only a human decision
closes this," and that has not changed. The gate reconfirms the interim rule
above stays locked through Phase 08's close-out (i.e. through this plan's own
Task 3, which does not touch `ownership.*` retirement), and treats
`REQUIREMENTS.md:173`'s Phase 09 mapping as the operative one for planning
purposes — OWN-09 remains mapped to Phase 09 in `REQUIREMENTS.md`, unmodified
by this plan, and Phase 09's own planning is where OWN-09's text should be
corrected to match, or the mapping itself is what should move. This gate does
not pick which document is "right"; it declines to let the conflict block
declaring the liveness law final, since the law's own scope (interprocedural,
run last, non-retiring) does not depend on the answer.

### D-08-37 — cross-run summary cache is a declared non-goal

> **NON-GOAL (Phase 08):** a persistent cross-run summary cache is explicitly
> deferred to Phase 11 / QLT-06. Spike S-006 measured that a single leaf edit
> invalidates **92%** of a call-graph-closure-keyed cache worst-case and **43%**
> mean on a realistic parser-shaped corpus, and **100%** on a chain. Within-run
> memoization is mandatory and sufficient for EFF-02; cross-run caching is a
> separate, unproven claim that must not be assumed or quoted as a production
> win until QLT-06's callee-changes-invalidates-caller regression test gates it.

Do **not** re-measure on production shapes — the spike's numbers are sufficient
for planning, and the milestone's declared 2x scope-cut trigger for Phases 08
and 09 makes the marginal evidentiary value a poor trade.

**Verified all-clear, recorded as a positive finding, not a defect:**
`ClosureDigest` is not referenced anywhere in `internal/compiler/cache`, and
`cache.Input` is constructed only in tests. There is **no QLT-06 violation in
flight**. `07-CONTEXT.md`'s line describing `cache.Input` as taking
`ClosureDigest` as "the named, digested input" is a statement of planning intent,
not shipped wiring. This non-goal pre-empts a future wiring; it does not describe
a current defect.

The Phase 08 memo is within-run only, keyed by function ID, rebuilt on every
`check.Program` call, never persisted — and that constraint is itself under test
(`TestSummaryMemoNeverPersisted`), because honouring it is what keeps this phase
clear of QLT-06.

### D-08-38 — the declared scope-cut trigger and its named landing phase

The milestone declares a **2x trigger** for Phases 08 and 09
(`.planning/REQUIREMENTS.md`, `.planning/ROADMAP.md`). Adopted for Phase 08 in
this exact shape:

**If the summary-table + liveness-law work (plans 08-01, 08-02, 08-03) exceeds
~2x its initial plan estimate, the cost-gate instrument work (plan 08-04:
D-08-32's two-chokepoint widening, D-08-34's corpus generator, D-08-36's risk
lane, D-08-33's manifest row) renegotiates into Phase 09.**

Never cut, under any circumstance:

- the criterion-1 adversarial corpus (D-08-28),
- the seeded mutation-kills (D-08-16),
- the two-path D-07-49 fix (D-08-09).

Criterion 3's measurement is the gate for this phase's **riskiest assumption**
("that a signature-mediated fixpoint stays a bounded, summary-based analysis
rather than becoming whole-program borrow inference in disguise"), so cutting it
is a **deferral with a named landing phase, never a silent drop**. If it is cut,
the liveness law **may not be declared final** until it lands, and Phase 09's
peer may not be declared complete on top of an unmeasured producer.

**2026-09-10 — resolved at Phase 08's mid-phase gate (08-06). The trigger did
not fire.** 08-04 (plan 4 of the summary-table + liveness-law work) landed at
its own estimated scope, and criterion 3's cost-gate instrument work (08-05)
shipped in full: `generateCallGraphCorpus`'s five-shape synthetic corpus, the
least-squares growth-exponent fit, both chokepoints (`measure.Demote`,
`session.QLT02GateEligibleMetrics()`) widened and proven to agree, and the
ratified `hard` manifest row. The gate re-measured the fit directly against
`buildInterproceduralSummaries`' own work counter at this gate's own commit:
every shape's fitted milli-exponent is 998-1003 (chain 1003, diamond 1002,
dense 1003, parser-shaped 1000, forward 998), all comfortably under the 1200
bound, with ratio-stability deltas between S=128 and 4S=512 all under 0.2%
(well inside the 15% tripwire). Because the instrument work was not cut and
the measurement genuinely shipped, the liveness law **is** declared final by
this gate — see Task 3 of 08-06-PLAN.md and its commit.

### D-08-40 — the two accepted-program peer-divergence fixtures, formalized as debt

`twin_a_accept.lang` and `relay_depth2_accept.lang` (landed 08-03) are each
registered in `internal/compiler/session/session_peer_gate_test.go`'s
`peerDivergenceExpected` map with `core.move_while_borrowed`. `check` admits
both correctly: the new interprocedural liveness law consults the callee's
declared/derived contract and finds no live-loan conflict. `corevalidate`'s
own loan-liveness peer (`loanChainIndex`) is still intraprocedural this phase
— unconditional `parent[TargetID] = SourceID` propagation through every
`OpCall`, with no callee-signature consultation at all — so it refuses both
fixtures via `core.move_while_borrowed` regardless of the callee's declared
contract.

This was previously visible only in 08-03-SUMMARY.md's key-decisions and the
`peerDivergenceExpected` map's own comments. The mid-phase gate promotes it
into the tracked debt register: it is exactly the input Phase 09's peer
re-derivation plan needs (the fixture headers already name the precise
mechanism `loanChainIndex` would need to change), and a register entry
survives past the point any one plan's SUMMARY scrolls out of context.

**Landing phase:** Phase 09 — `corevalidate`'s independent loan-liveness peer
must consult callee summaries the same way `check` now does, closing OWN-07
and this divergence together.

### D-08-41 — Pattern B's real-fixture scope limit on success criterion 1

`twin_b_refuse.lang`/`twin_b_accept.lang` (Pattern B, `UsesParam`) are a
genuine twin pair — mechanically verified caller-identical by
`TestInterproceduralTwinPairsDifferOnlyInTheCallee` — but they cannot
demonstrate a DIFFERING end-to-end CLI verdict. The caller's required
`borrow; move; call` shape is refused INTRAPROCEDURALLY
(`ownership.move_while_borrowed`) for BOTH members alike by
`computeLoanLastUses`' summary-blind AST-shadow admission path, before
`check`'s own interprocedural pass ever runs. This is `computeLoanLastUses`
working exactly as designed (summary-blind by construction, D-08-09's second
admission path) — not a bug, and not something this phase's `files_modified`
(fixtures and tests only, no changes to `computeLoanLastUses`) was scoped to
change.

The contract-driven differentiation the fixture pair exists to demonstrate is
proven at the checked-core level instead: 08-02's own
`TestInterproceduralLivenessTwinPatternB` constructs the two `core.Program`
values directly and shows the backward gate genuinely reads `UsesParam` and
produces the correct split verdict there. 08-03-SUMMARY.md's "Plan-text vs.
verified-reality notes" documented this as a note at the time; the mid-phase
gate adjudicates it explicitly now: this is an accepted, permanent, disclosed
scope limitation of the real-`.lang`-fixture corpus for Pattern B, not
deferred work and not a defect. Success criterion 1 is met for Pattern B at
the checked-core level; it is not met, and cannot be met without changing
`computeLoanLastUses` (out of this phase's charter), at the real-fixture CLI
level.

**Landing phase:** Not scheduled. Reopen only if a future phase makes
`computeLoanLastUses` summary-aware, which is not currently planned for any
numbered phase.

### D-08-42 — stale call-graph-order direction in two planning documents

`08-CONTEXT.md`'s D-08-03 and 08-01-PLAN.md's original doc comment both assert
`callgraph.Order`'s reverse postorder lists a function CALLEE-before-CALLER.
08-02's own investigation (`TestOrderSortsAdjacencyByCalleeID` plus this
package's `TestSummaryDerivationIsOnePassPerFunction`) found the opposite
empirically true: `Order` lists a CALLER before every function it calls.
`buildInterproceduralSummaries` (`internal/compiler/check/check.go`) already
carries the corrected doc comment and walks `Order`'s result BACKWARD to get
genuine callee-before-caller derivation — the production code has been
correct since 08-02 landed.

Recorded here so the two now-stale planning documents do not mislead a future
reader who trusts CONTEXT.md over the code itself. No production-code change
is implied or needed.

**Landing phase:** Not scheduled — this is a documentation-accuracy note, not
a deferred implementation item. `08-CONTEXT.md` is a planning artifact and is
not corrected retroactively by this plan (out of scope; would rewrite
phase-history at the wrong layer). A reader relying on `08-CONTEXT.md`'s
D-08-03 should instead trust `check.go`'s own doc comment and this entry.

### D-08-43 — pre-existing spike-registry gap surfaces as the one non-Phase-08 test failure

`TestQLT01RegistryCoversAllFiveSpikes` (`internal/compiler/session`) fails
with "spike 006 has a directory under .planning/spikes but no registry row
cites it." This was verified present in the working tree before Phase 08's
first plan executed, is unrelated to any Phase 08 file, and is not caused by
or fixed in this phase — it is a `.planning/spikes` registry-maintenance gap
(a row was never added for spike 006's directory).

Recorded here, rather than left as an unexplained `go test ./...` failure a
future reader might mistake for Phase 08 regression, so the one known,
tracked, pre-existing failure is traceable to this entry.

**Landing phase:** Not scheduled — no phase currently owns `.planning/spikes`
registry maintenance as a charter item. Whoever next touches the spike
registry (or plans a phase that does) should add spike 006's row and close
this entry.
