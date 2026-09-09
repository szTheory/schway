---
phase: 08-interprocedural-loan-liveness-in-check
recorded: 2026-09-09
status: accepted
disposition: in-progress
items: 5
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
| D-08-26 | 08-CONTEXT.md (D-08-26), 08-03-PLAN.md | OWN-06 (success criterion 4) | warning | Phase 09 — Peer Re-Derivation and D-03-02 Closure (cost question adjudicated at Phase 08's mid-phase gate) | Success criterion 4's ACCEPTED-program half ships no runtime artifact this phase. On refusal the disclosure is real (cause 3 names the one consulted callee-signature field); on acceptance no shipped vehicle can carry it — `protocol.ExplainSummary` is synthesized only from an existing diagnostic's causes, `internal/compiler/cache` is structurally incapable of holding a verdict (`TestCacheExportedSurfaceStoresNoVerdict`), and a sibling `lang.*/1` document contradicts `protocol.InterfaceSummary`'s own recorded anti-pattern. Satisfied this phase by a table-driven test asserting the consulted field set, not by a shipped artifact |
| D-08-27 | 08-CONTEXT.md (D-08-27), REQUIREMENTS.md:173 vs OWN-09's own text | OWN-09 | warning | Phase 09 — Peer Re-Derivation and D-03-02 Closure (conflict surfaced at Phase 08's mid-phase gate) | The project's own documents disagree about when the intraprocedural loan-liveness law retires: OWN-09's text says "retired in the same phase the interprocedural law lands" (Phase 08); `REQUIREMENTS.md:173` maps OWN-09 to Phase 09. Only a human decision closes this. Interim rule locked for Phase 08: `ownership.*` codes are NOT retired, and the interprocedural check runs LAST, after intraprocedural admission has already passed, so no program is judged by both laws for the same fact |
| D-08-37 | 08-CONTEXT.md (D-08-37), spike S-006 iteration 6 | QLT-06 | info | Phase 11 — Multi-Function Native Emission and Equivalence | NON-GOAL (Phase 08): a persistent CROSS-RUN summary cache is explicitly deferred. Spike S-006 measured that a single leaf edit invalidates 92% of a call-graph-closure-keyed cache worst-case, 43% mean on a realistic parser-shaped corpus, and 100% on a chain. Within-run memoization is mandatory and sufficient for EFF-02; cross-run caching is a separate, unproven claim that must not be assumed or quoted as a production win until QLT-06's callee-changes-invalidates-caller regression test gates it |
| D-08-38 | 08-CONTEXT.md (D-08-38), ROADMAP.md M002 scope-cut order | EFF-02 | info | Phase 09 — Peer Re-Derivation and D-03-02 Closure (only if the trigger fires) | Declared scope-cut trigger: if the summary-table + liveness-law work exceeds ~2x its initial plan estimate, the cost-gate INSTRUMENT work (chokepoint widening, corpus generator, risk lane, manifest row) renegotiates into Phase 09 — never the criterion-1 corpus, never the seeded mutation-kills, never the two-path D-07-49 fix. If it is cut, the liveness law may NOT be declared final until it lands |

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
witness — the unexported `loanLivenessBoundSeam` package-level var, false in
production, flipped and deferred-restored by a same-package test that asserts
the named refusal fires — so it is never a control that has never been seen to
fail.

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
case should become a runtime artifact, and what it costs, is on the mid-phase
gate's agenda (D-08-40b).

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
with coverage). This item is on the mid-phase gate's agenda (D-08-40a) and is
recorded as a conflict, not as a resolution.

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
