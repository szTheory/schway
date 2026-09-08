# Phase 07 — Declared Deferred Scope (D-07-27)

**Written:** 2026-09-08, at *planning* time — not at phase end.
**Amended:** 2026-09-08, after cross-AI review, when the plan set was rewritten
from 5 plans to 8 and D-07-29..D-07-45 were locked.
Per Key Lesson 4: declare deferred scope in writing at the moment it is decided.

---

## Deferred item 1 — interprocedural loan-*liveness* re-derivation

**Interprocedural loan-liveness re-derivation in `corevalidate`.**

Phase 07 ships `corevalidate`'s independent **signature-summary** re-derivation
(D-07-20, `07-02-PLAN.md`) and its independent **call-graph cycle** traversal
(D-07-19, `07-07-PLAN.md`). It does **not** ship an independent re-derivation of
interprocedural *loan liveness*.

### Lineage

D-03-02, open past M001: an exported borrow-derived return with no declared
origin exports indistinguishable from a fully-owned return, in the
**interprocedural** half of the hazard. The single-function half closed in M001
Phase 3. The interprocedural half is owned by M002's `OpCall` charter
(D-05-32 / D-05-33).

### Closure phase

**Phase 09 — Peer Re-Derivation and D-03-02 Closure** (OWN-05, OWN-07, OWN-08,
OWN-09, TRU-04, QLT-07).

### Closure gate, verbatim from ROADMAP.md Phase 09

> `corevalidate` independently re-derives the same interprocedural loan-liveness
> facts without sharing an implementation with `check`; a seeded endpoint-level
> fault makes the two peers diverge (OWN-07), and D-03-02 is closed — an
> exported borrow-derived return with no declared origin is refused in the
> interprocedural case, in **both** admission layers (OWN-08).

### Explicitly NOT part of this deferral

The **signature-summary peer ships in Phase 07** (`07-02-PLAN.md`, D-07-20).
Deferring the liveness peer does not defer the summary peer, and does not defer
the seeded faults (D-07-24).

---

## Deferred item 2 — the peer's `Callable` re-derivation is NARROWED (D-07-33)

**This is new in the post-review amendment and it is stated plainly because an
undeclared version of it is exactly the failure mode this file exists to prevent.**

`Callable` **as published** is the full D-04-03 predicate: publication safety,
i.e. `originvalidate.PublishProblemsFor` returns no problems (D-07-31, D-07-32).
`ValidatePublished` can refuse with any of four classes:

  - `core.origin_omitted`
  - `core.origin_understated`
  - `core.origin_access_mismatch`
  - foreign-origin-omitted (`checkForeignOriginOmitted`)

**In Phase 07, `corevalidate`'s independent peer re-derives only the first of
those four.** For `core.origin_understated`, `core.origin_access_mismatch`, and
foreign-origin-omitted, **the Phase 07 peer can only ever falsely agree with the
producer.** It does not compute those classes at all, so its agreement on them is
not evidence of anything. This is precisely the single-producer leak D-07-23
names — `escape:coordinated-source-to-core-false-claim` promoted from origins to
the call contract — scoped and time-boxed rather than denied.

**Why narrowed:** a full peer-side origin recomputation is a second
implementation of `RecomputeOrigin` / `RecomputeOriginPerReturn`
(`originvalidate.go:234`, `:133`). That is Phase 09's size, not Phase 07's. This
is D-07-26's scope-cut trigger being pulled **deliberately, at planning time**,
rather than discovered mid-execution.

**What the phase does still guarantee for the narrowed classes:** the producer
computes them, they refuse publication, and `Callable` is false as published. What
is missing is only the *second, independent* derivation.

**Closure phase:** **Phase 09**, alongside the liveness peer, in the same plan
that builds the peer's own origin recomputation.

---

## Phase 07's own scope-cut trigger (D-07-26) — updated for the 8-plan set

If **Stage 0 + Stage 1** (`07-01` + `07-02` + `07-03` + `07-04` + `07-05`) exceed
**~2x their initial plan estimate**, Stage 2's cycle refusal (`07-06`, `07-07`)
renegotiates out to Phase 08.

**New consequence introduced by the post-review re-plan, recorded so it is not
discovered under pressure:** `07-08` (closure-digest chaining) **goes with them**.
D-07-38 orders digest chaining strictly behind cycle refusal — the chain
terminates only on a DAG — so cutting `07-06`/`07-07` necessarily cuts `07-08`.
If that cut is taken, `ClosureDigest` ships with its zero-callee base case only
(`07-01`), which is correct and complete for a corpus with no calls, and the
chaining arm plus its callee-changes-invalidates-caller regression move to
Phase 08 with the cycle refusal.

**Never cut under this trigger:**
- the Stage 0 `corevalidate` signature-summary peer (D-07-20)
- the seeded faults (D-07-24), including the bilateral one that must FAIL the gate

---

## Accepted, declared limitations of Phase 07 (declared, not implied)

- **`cgen`'s `OpCall` case is structurally unreachable this phase** (A-02).
  `cgen.Emit`/`EmitNative` hard-fail on `len(Functions) != 1` (`cgen.go:22,52`)
  and both exhaustive-dispatch controls gate the `cgen` site behind
  `len(program.Functions) == 1`. Phase 07 registers the case as forward hygiene
  for Phase 11 and claims **no** `cgen` runtime coverage for `OpCall`. Per
  **D-07-39** the arm is an explicit, dedicated "recognized, unsupported in
  Phase 07" error and is never folded into the grouped copy/move/borrow cases —
  folding it there would emit copy-like C and let a green control certify a stub.
- **`interp` recognizes but does not execute `OpCall`** (D-07-39). No call-stack
  semantics ship this phase; SEM-08's bounded call stack with its fixed
  documented ceiling is Phase 10. The exhaustive-dispatch controls assert
  **recognition**, never execution, and say so in their doc comments.
- **`restrict` at arity 1 is correct but never load-bearing** (D-07-04).
  Phase 11 must scope its `restrict` evidence to single-parameter derivation and
  name the two-parameter case as an M003 obligation, rather than implying
  two-pointer coverage it does not have.
- **No maximum call-graph depth is declared this phase** (D-07-17). Depth is
  unbounded provided the graph is acyclic; the distinguisher is on-stack (gray)
  re-entry, never visit count or depth. The fixed documented interpreter ceiling
  is SEM-08, Phase 10, informed by spike S-007.
- **No `Span` on `core.LinearOperation`** (D-07-35). Cycle-diagnostic spans are
  projected on the `check` side from operation IDs. The `corevalidate` peer emits
  the shared code string with **no spans** — its job is refusal, not diagnostics.
  Any future consumer wanting spans from the peer must add them deliberately.
- **No cross-module summary channel** (D-07-34). `check.Program(ast.Program)`
  keeps its signature; admission consults an in-process pre-body signature table.
  Cross-module summary consumption is Phase 09+.
- **No interprocedural fact is marked cacheable** (`07-08`). QLT-06 is a later
  requirement; Phase 07 builds the callee-changes-invalidates-caller regression
  and the closure-derived key it will need, and modifies `cache.Input` not at all.
- **Arity-N calls and multi-argument loan interaction** (D-07-07) are deferred.
  The `/1` schema is arity-ready (D-07-10), so the future change is a checker
  predicate plus an aliasing rule, not a `/2`.
- **Labeled call-site arguments** (M001 D-02) are deferred, not revoked (D-07-06),
  and reinstated when arity widens past 1.
- **Keyed/signed summary digests** are out of scope. D-07-13: content digests
  detect staleness, never forgery — chaining them (`07-08`) widens what staleness
  they detect and adds no authenticity. Independent re-derivation is the forgery
  answer.

---

## Explicitly rejected, not deferred

Recorded here so a later phase does not re-open them believing they were merely
postponed.

- **Call-graph edges in the signature summary** — permanently rejected (D-07-11).
  It would make every consumer whole-program-aware and re-import the
  already-rejected Design A.
- **An export set in `core.Program` / `core.Function`** — rejected (D-07-31).
  `Callable` is publication safety, not export membership;
  `originvalidate.ValidatePublished` never reads an export list, so the schema
  never wanted the field. Suggested by a reviewer; declined on the evidence.
- **The `export_callee` repair for the SEM-06 refusal** — rejected (D-07-31c).
  Exporting a function cannot fix an unsafe borrow-derived return; an agent
  applying the repair would re-submit an identically-refused program. Suggested
  by a reviewer; declined on the evidence. If a repair is ever offered for that
  code it must target the return-origin contract.
- **A depth or visit ceiling on the call-graph traversal** — rejected (D-07-17).
  It would convert a correctness property into a resource limit and silently
  refuse legal deep-but-acyclic programs. Only the emitted diagnostic is bounded.

---
*Written at planning time per D-07-27; amended after cross-AI review.*
