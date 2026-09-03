# Phase 1: Canonical Pure Spine - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md; this log preserves alternatives considered.

**Date:** 2026-09-03
**Phase:** 01-canonical-pure-spine
**Mode:** Auto-selected from prior author intent and the approved execution plan
**Areas discussed:** Provisional source surface, repository/compiler boundaries, command/diagnostic contract, evidence/cost discipline

---

## Provisional Source Surface

| Option | Description | Selected |
|--------|-------------|----------|
| Conventional braced semantic prose | Familiar recovery landmarks, canonical formatting, named roles, replaceable projection | ✓ |
| Indentation and keyword flow | Airier source but more recovery/whitespace ambiguity | |
| Uniform tree syntax | Minimal parser ambiguity but higher human translation cost | |

**Auto-selected choice:** Conventional braced semantic prose.
**Notes:** This matches the leading syntax study and keeps S1 provisional. Exact-name argument punning is canonical; positional calls and user-defined syntax are excluded.

---

## Repository and Compiler Boundaries

| Option | Description | Selected |
|--------|-------------|----------|
| Root Go module with `cmd/lang` and `internal/compiler` | Simplest clean-checkout command and future room for runtime/tool packages | ✓ |
| Nested `compiler/` Go module | Strong physical boundary but adds root invocation/workspace ceremony now | |
| Extend a spike module | Fastest code reuse but turns disposable evidence into a shadow architecture | |

**Auto-selected choice:** Root Go module with production-only packages.
**Notes:** Spike claims and fixtures migrate conceptually; spike Go APIs are not dependencies.

---

## Command and Diagnostic Contract

| Option | Description | Selected |
|--------|-------------|----------|
| One `lang` CLI with human default and `--json` projection | Same semantics for human/agent, stable scripts, minimal process surface | ✓ |
| Separate human and agent tools | Independent evolution but duplicate behavior and drift risk | |
| JSON-only tool | Small protocol but worse direct developer experience | |

**Auto-selected choice:** One CLI, two projections.
**Notes:** Phase 1 owns format/check/run/evidence; query/explain/verify arrive only after the semantic corpus exists.

---

## Evidence and Cost Discipline

| Option | Description | Selected |
|--------|-------------|----------|
| Layered evidence with provisional guardrails | Every lane catches a named defect; cheap loop stays fast; costs recorded | ✓ |
| Full verification on every edit | Maximum immediate evidence but repeats expensive native/sanitizer work | |
| Fixtures only until later | Fastest initial code but leaves formatter/backend invariants untested | |

**Auto-selected choice:** Layered evidence with properties, mutations, and differential runs from S1.
**Notes:** Phase 1 records provisional Mac thresholds; Phase 6 ratifies portable budgets from distributions.

## the agent's Discretion

- Fine punctuation within the chosen syntax family.
- Go package granularity and internal data structures.
- Benchmark implementation and diagnostic wording.
- Clean-checkout command wrapper mechanics.

## Deferred Ideas

Ownership, borrows, resources/FFI depth, full adversarial native validation,
agent query/repair ergonomics, effects, concurrency/runtime, and ecosystem kits
remain in their mapped phases.
