# Phase 19: Numeric Literals and `OpConst` - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-24
**Phase:** 19-Numeric Literals and `OpConst`
**Areas discussed:** Unsigned type and bounds, Literal spelling and formatting, How literals get typed

---

## Unsigned type and bounds

| Option | Description | Selected |
|--------|-------------|----------|
| `U64` | Architecture-independent 64-bit unsigned value; reject out-of-range literals and require exact target support. | ✓ |
| `U32` | Smaller fixed range and potentially simpler on narrower targets. | |
| `Other` | User-specified width or target rule. | |

**User's choice:** Automatically follow the recommendation: `U64`.
**Notes:** No truncation, wrap, or target-dependent source width.

---

## Literal spelling and formatting

| Option | Description | Selected |
|--------|-------------|----------|
| Recommended compact spelling set | Decimal, `0x` hexadecimal, `0b` binary; separators only between digits; no octal or suffixes; preserve lexeme and canonicalize layout. | ✓ |
| Decimal only | Smallest initial grammar, with less systems-oriented notation. | |
| Broad C/Rust-like spelling set | More bases and suffix choices, with additional ambiguity and language rules. | |

**User's choice:** Automatically follow the recommendation.
**Notes:** Keep source spelling distinct from canonical decimal runtime serialization.

---

## How literals get typed

| Option | Description | Selected |
|--------|-------------|----------|
| One implicit `U64` type | `let count = 42`; all literals have the only numeric type. | ✓ |
| Contextual inference | Infer a literal's type from surrounding expressions or signatures. | |
| Explicit suffixes | Let literal suffixes select among numeric types. | |

**User's choice:** Automatically follow the recommendation.
**Notes:** No suffix-driven, contextual, or arbitrary-precision constant system.

---

## the agent's Discretion

Parser/core representation, helper boundaries, fixture names, diagnostic text,
and exact test factoring remain with the planner under the locked decisions in
`19-CONTEXT.md`.

## Deferred Ideas

None raised beyond the roadmap's existing deferrals.
