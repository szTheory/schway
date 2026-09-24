# Phase 18: Branch on a Computed Value - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-24
**Phase:** 18-branch-on-a-computed-value
**Areas discussed:** Computed scrutinee and source boundary; ownership, calls, and native evidence

---

## Computed scrutinee and source boundary

| Option | Description | Selected |
|--------|-------------|----------|
| Generalize terminal `match` to an in-scope `data` place | Follows the Phase 18 goal and existing construct; keeps one control-flow law | ✓ |
| Add `if`/`else` or a separate branch form | Conflicts with the roadmap's explicit out-of-scope boundary | |

**Choice:** Generalize the terminal `match` to an in-scope `data` place (roadmap-locked; recommended).
**Notes:** `--auto` selected the scoped option; no additional syntax capability was added.

---

## Ownership, calls, and native evidence

| Option | Description | Selected |
|--------|-------------|----------|
| Treat S-010 as planning permission and re-prove it on the production source path | Preserves the distinction between spike evidence and phase acceptance | ✓ |
| Treat the spike alone as production proof | Contradicts Phase 18 success criterion 4 | |

**Choice:** Re-run the borrowed-value shape through production code and the declared liveness bound (roadmap-locked; recommended).
**Notes:** `--auto` selected objective automated source, differential, mutation, smoke, and CI evidence wherever recurring value justifies its cost, in line with the recorded user preference.

---

## the agent's Discretion

- Parser/core representation, helper boundaries, fixture names, and test factoring within the locked phase boundary.

## Deferred Ideas

- `if`/`else`, `Bool`, arithmetic, and loops remain deferred by the M003 roadmap.
