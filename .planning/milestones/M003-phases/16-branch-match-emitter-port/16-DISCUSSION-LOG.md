# Phase 16: Branch/Match Emitter Port - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-19
**Phase:** 16-Branch/Match Emitter Port
**Areas discussed:** emitter-port boundary, C `restrict` probe, compatibility and golden evidence, formal M004 cut

---

## Emitter Port Boundary

| Option | Description | Selected |
|---|---|---|
| Narrow capability port | Port match, branch, and ordinary linear lowering into `emitProgram`; delete three legacy emitters at cutover. | ✓ |
| Six-emitter rewrite | Rebuild foreign and both by-pointer families in the surviving backend. | |
| Retain N=1 dispatch | Add partial support while preserving the function-count fork. | |

**User's choice:** Approved the research-backed narrow capability port.
**Notes:** The port must preserve checker-derived payload layout, Phase 15 `/2` event behavior, validation precedence, and same-commit deletion.

---

## C `restrict` Probe

| Option | Description | Selected |
|---|---|---|
| Run narrow probe now | Test the exact read-only, one-pointer, one-TU static-callee shape before planning. | ✓ |
| Defer by-pointer lowering | Retain the M004 cut pending a full caller-side discharge design. | |

**User's choice:** Approved the narrow probe with structural acceptance guards.
**Notes:** A positive result cannot prove LTO non-inert or generalize to mutation, escape, foreign calls, callbacks, arity-N, or separate compilation.

---

## Compatibility and Golden Evidence

| Option | Description | Selected |
|---|---|---|
| Ledger plus identity gate | Require byte identity for three landed shapes and a machine-linked four-entry golden-change ledger. | ✓ |
| Digest updates only | Treat revised hashes as sufficient explanation. | |
| Semantic comparison only | Skip exact-C characterization after the port. | |

**User's choice:** Approved ledger plus byte identity and four-tier semantic evidence.
**Notes:** The ledger records path, old/new hash, structural cause, responsibility moved, semantic witness, fixture, and disposition.

---

## Formal M004 Cut

| Option | Description | Selected |
|---|---|---|
| Formal amendment and owned debt | Use D-10-60's amendment clause and name real prerequisites and LTO consequence. | ✓ |
| Silent deferral | Carry unsupported families forward without a formal disposition. | |
| Expand Phase 16 | Build foreign-resource and alias-discharge subsystems now. | |

**User's choice:** Approved the formal cut, conditional only on the bounded by-pointer probe.
**Notes:** Unsupported foreign and by-pointer families must not be forced into Phase 16 merely to satisfy an old deletion target.

---

## the agent's Discretion

Helper names, test factoring, ledger serialization, and diagnostic prose remain flexible inside the locked evidence and scope boundaries.

## Deferred Ideas

None — all related follow-on work is explicitly owned by M004.
