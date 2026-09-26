# Phase 15: Event Identity (`lang.execution/2`) - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-19
**Phase:** 15-event-identity-lang-execution-2
**Areas discussed:** Invocation path grammar, Observable OpCall event contract, Independent peer validation strictness, Static invocation-path table ceiling

---

## Invocation Path Grammar

| Option | Description | Selected |
|--------|-------------|----------|
| Escaped readable ancestry path | Deterministic, peer-rederivable root/call-site chain with canonical escaping and explicit ordinals. | ✓ |
| Length-prefixed text segments | Unambiguous arbitrary bytes, but opaque and easy to mis-parse. | |
| Structured JSON path | Extensible nested metadata at the cost of schema and canonical-byte complexity. | |
| Fixed-width path hash | Compact, but opaque and needlessly reintroduces collision policy. | |

**User's choice:** Accepted the synthesized recommendation set in full.
**Notes:** Always emit `#0` now; reject non-canonical escaping; occurrence identity is not content identity.

---

## Observable `OpCall` Event Contract

| Option | Description | Selected |
|--------|-------------|----------|
| Caller-owned pre-activation event | One `function.called` edge owned by the caller and emitted immediately before child activation. | ✓ |
| Callee-owned activation event | Groups by callee lifecycle but obscures ownership of the source call operation. | |
| Paired call/completion events | Span-like lifecycle detail with doubled cardinality and pairing failure modes. | |

**User's choice:** Accepted the synthesized recommendation set in full.
**Notes:** The removal control asserts exactly one loss in the `function.called` projection; full-document callee events necessarily also disappear when a real call is removed.

---

## Independent Peer Validation Strictness

| Option | Description | Selected |
|--------|-------------|----------|
| Membership only | Verifies admissibility but accepts omission, reparenting, duplication, and impossible order. | |
| Membership, uniqueness, and binding | Catches collision and frame attribution but not malformed causal sequence. | |
| Exact observed structural validation | Validates membership, uniqueness, function binding, ownership, and preorder without requiring untaken paths. | ✓ |
| Exact static-tree multiplicity | Strong for straight-line fixtures but becomes unsound once branches exist. | |

**User's choice:** Accepted the synthesized recommendation set in full.
**Notes:** Static-tree equality remains a dedicated straight-line coverage control; producer and peer seams each receive a directional seeded fault.

---

## Static Invocation-Path Table Ceiling

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed 4096-node preflight cap | Deterministic bound with 67.1× measured headroom and a named refusal. | ✓ |
| Node cap plus byte cap | Adds a separate generated-size budget before evidence shows it is needed. | |
| Adaptive source-derived budget | Looks flexible but cannot soundly bound exponential DAG unfolding. | |
| Runtime-computed paths | Avoids static materialization but adds runtime allocation, overflow, and audit costs. | |

**User's choice:** Accepted the synthesized recommendation set in full.
**Notes:** Independent measurement confirmed 61 nodes including entry. Boundary controls pin 4096 accepted, 4097 refused, and mutation-kill preflight removal.

---

## the agent's Discretion

- Internal factoring, helper names, and exact diagnostic prose, within the locked wire and evidence contracts.

## Deferred Ideas

None.
