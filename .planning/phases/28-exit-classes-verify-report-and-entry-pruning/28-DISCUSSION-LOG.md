# Phase 28: Exit Classes, Verify Report, and Entry Pruning - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-10
**Phase:** 28-exit-classes-verify-report-and-entry-pruning
**Areas discussed:** Entry rule and pruning point, Exit-code table, App verify outcome digest
**Mode:** Advisor (minimal_decisive). All areas were auto-selected and the recommendations auto-followed per the user's standing instruction ("synthesize coherent one-shot recommendations and auto-follow them"). There were no interactive picks.

---

## Entry rule and pruning point

| Option | Description | Selected |
|--------|-------------|----------|
| A | `main` wins, else unique root, else refuse; check all, then prune once | ✓ (modified) |
| B | Keep current rule, add `main` only as a tiebreak, no pruning | |
| C | Pure `fn main` rule or an exported-entry marker | |
| D | Prune per command | |

**Choice:** A, but the closure-size tiebreak is kept for programs without `main` (D-03). The research recommended removing it. Code check showed it is the adopted D-11-05 contract with a dedicated test, and removing it would refuse programs that run today, which contradicts DX-18.

---

## Exit-code table

| Option | Description | Selected |
|--------|-------------|----------|
| A | New leaf package `internal/compiler/exitcode` with named domains | ✓ |
| B | Extend the table inside `protocol` | |
| C | Per-package consts plus a grep test (status quo) | |

**Choice:** A. `app run` stays in its env/timeout domain, which was verified: it takes a built artifact, not source. The W2/W3 fix is a shared `admitProgram` that runs before any tool. The static guard is an AST scan with a negative-control self-test.

---

## App verify outcome digest

| Option | Description | Selected |
|--------|-------------|----------|
| A | Digest-only for exhausted/diverged cases, schema /2, case-ID cap, static budget test | ✓ |
| B | Truncate/degrade evidence | (rejected at milestone level) |
| C | Runtime size check only | |
| D | Evidence sidecar file | |

**Choice:** A.

---

## Claude's Discretion

Helper names, file split, golden format, and whether cgen formats exit literals from constants or test-asserts them.

## Deferred Ideas

Explicit-entry language change; an evidence sidecar; loop-aware capacity (D-27.3-03); N4/R1b.
