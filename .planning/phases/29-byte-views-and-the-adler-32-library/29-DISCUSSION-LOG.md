# Phase 29: Byte Views and the Adler-32 Library - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-11
**Phase:** 29-byte-views-and-the-adler-32-library
**Areas discussed:** Byte-read surface syntax, Fixture byte delivery, Spike sequencing and cap evidence

Mode: advisor (`minimal_decisive`), one gsd-advisor-researcher per area, recommendations
auto-followed per the user's standing fan-out-then-auto-follow preference (no per-area prompt).

---

## Byte-read surface syntax

| Option | Description | Selected |
|--------|-------------|----------|
| Builtins `byte_len(data)` / `byte_at(data, i)` | Name-recognized atoms like `write_u64`; one LL(1) branch; no lexer change; first 2-arg call form, special-cased | ✓ |
| `data[i]` + `len(data)` / `data.len` | Familiar; adds `[`/`]`, postfix parse, still needs a length spelling; pre-commits array syntax | |
| Method-ish `data.at(i)` / `data.len()` | Dropped: `scalarLetStart` rejects `ident.`; implies a method system | |

**Choice:** Builtins (auto-followed).
**Notes:** `[]` can later be sugar over the same core op.

---

## Fixture byte delivery

| Option | Description | Selected |
|--------|-------------|----------|
| Test-only foreign fixture owner | acquire(case U64) / borrow→ByteView / consume; straight-line caller; independent C and Go byte generators; runs through unmodified `schway run` | ✓ |
| Synthetic harness injection / new run flag / ByteView entry | Smallest change but bypasses `schway run`, common-mode risk, never exercises caller loan/release, pre-empts Phase 31 | |
| Byte-string literal syntax | New surface across parser/formatter/six consumers; needs a repeat form | |

**Choice:** Foreign fixture owner (auto-followed).
**Notes:** Verified `schway run --engine=X SOURCE` takes no input; case id is baked into each fixture.

---

## Spike sequencing and cap evidence

| Option | Description | Selected |
|--------|-------------|----------|
| RESEARCH.md settles (a)(b)(c)(e)(f); Wave-0 plan 29-00 gate measures (d) on the real path | No extra cycle; measures product path; gate before six-consumer spread | ✓ |
| Standalone `/gsd-spike` 008 before planning | Clean separation, but ByteView doesn't exist yet, so it measures a proxy | |

**Choice:** Wave-0 gate (auto-followed), with the D-12 measurement protocol and D-13 lowering rule.

---

## Claude's Discretion

- Expr kind names, fixture owner type name, adapter layout, case-id numbering, generated vs hand-written fixtures, repair coverage depth.

## Deferred Ideas

- `[]` sugar (arrays milestone); byte-string literals; public `schway run` input; all Phase 31 file/route work.
