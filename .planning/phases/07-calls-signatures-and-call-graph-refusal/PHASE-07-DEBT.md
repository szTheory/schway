# Phase 07 — Declared Deferred Scope (D-07-27)

**Written:** 2026-09-08, at *planning* time — not at phase end.
Per Key Lesson 4: declare deferred scope in writing at the moment it is decided.

## Deferred item

**Interprocedural loan-*liveness* re-derivation in `corevalidate`.**

Phase 07 ships `corevalidate`'s independent **signature-summary** re-derivation
(D-07-20) and its independent **call-graph cycle** traversal (D-07-19). It does
**not** ship an independent re-derivation of interprocedural *loan liveness*.

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

The **signature-summary peer ships in Phase 07**, in the same plan as its
producer (`07-01-PLAN.md`, D-07-20). Deferring the liveness peer does not defer
the summary peer, and does not defer the three seeded faults (D-07-24).

## Phase 07's own scope-cut trigger (D-07-26)

If **Stage 0 + Stage 1** (`07-01-PLAN.md` + `07-02-PLAN.md` + `07-03-PLAN.md`)
exceed **~2x their initial plan estimate**, Stage 2's `callgraph` cycle refusal
(`07-04-PLAN.md`, `07-05-PLAN.md`) renegotiates out to Phase 08.

**Never cut under this trigger:**
- the Stage 0 `corevalidate` signature-summary peer (D-07-20)
- the three seeded faults (D-07-24)

## Accepted, declared limitations of Phase 07 (declared, not implied)

- **`cgen`'s `OpCall` case is structurally unreachable this phase** (A-02).
  `cgen.Emit`/`EmitNative` hard-fail on `len(Functions) != 1` (`cgen.go:22,52`)
  and both exhaustive-dispatch controls gate the `cgen` site behind
  `len(program.Functions) == 1`. Phase 07 registers the case as forward hygiene
  for Phase 11 and claims **no** `cgen` runtime coverage for `OpCall`.
- **`restrict` at arity 1 is correct but never load-bearing** (D-07-04).
  Phase 11 must scope its `restrict` evidence to single-parameter derivation and
  name the two-parameter case as an M003 obligation.
- **No maximum call-graph depth is declared this phase** (D-07-17). The fixed
  documented interpreter ceiling is SEM-08, Phase 10.

---
*Written at planning time per D-07-27.*
