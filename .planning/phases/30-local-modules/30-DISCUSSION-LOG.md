# Phase 30: Local Modules - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md. This log preserves the alternatives considered.

**Date:** 2026-10-12
**Phase:** 30-local-modules
**Areas discussed:** Import surface, Name/path resolution, Unit reach and C names, Caps and gate shape
**Mode:** Advisor (`minimal_decisive`). Four parallel researchers, each recommendation **auto-followed** per the user's standing "fan out, then auto-follow" preference. No interactive questions were asked.

---

## Import surface

| Option | Description | Selected |
|--------|-------------|----------|
| Named import block, bare call | `import m { fn f }`, then `f(x)`. The callee stays a bare string. | ✓ |
| Plain import, qualified call | `import m`, then `m.f(x)`. Needs dotted-callee changes in three parsers and six consumers. | |
| Hybrid | Both forms. Doubles the surface. | |

**Choice:** named import block. Imports are unsorted (source order), with no commas, and come before `export`.

## Name/path resolution

| Option | Description | Selected |
|--------|-------------|----------|
| Root = entry dir, entry header unchecked, imported header = path | No new check on programs without imports | ✓ |
| Entry header must end with its stem | Would add refusals to existing programs | |
| Project-root marker | An upward search contradicts MOD-03 | |
| Allow in-root symlinks (plain `os.Root`) | Path aliasing breaks duplicate detection and identity | |
| Refuse all symlinks + ReadDir exact case | Deterministic on macOS and Linux | ✓ |

**Choice:**
- Root is the entry directory. The entry header is unchecked.
- Segments are `[a-z][a-z0-9_]*`.
- Symlinks are refused, and case must match exactly through ReadDir.
- Each file is Lstat'd and confirmed regular before it is opened.
- The empirical `os.Root` spike ran on Go 1.24.0 darwin/arm64.

## Unit reach and C names

| Option | Description | Selected |
|--------|-------------|----------|
| Unit = module name, omitempty, on Diagnostic/Cause/Repair/debugmap | IDs and goldens unchanged | ✓ |
| Unit = relative path | Leaks layout into the protocol | |
| Both | Two sources of truth | |
| Span.File | Changes every existing JSON | |
| Length-prefixed `SCHWAY_M_<len>_<SEG>…` | Injective, keeps the namespace invariant | ✓ |
| Escape-based | Risks `__` and bijection bugs | |
| Numeric module index | Unstable | |

## Caps and gate shape

| Option | Description | Selected |
|--------|-------------|----------|
| ARCHITECTURE §5.1 caps: per-file 1 MiB (existing), 16 modules, 4 MiB total, depth 8, plus 16 imports per module | Aligned with the existing syntax limit | ✓ |
| Researcher A1: 32 modules | Looser than the settled architecture | |
| STACK 64 KiB per file | A second size number that could drift | |
| Gate-plan 30-00 + GATE-VERDICT, ceiling 10, split trigger | Phase 29 precedent | ✓ |
| Standalone /gsd-spike | Contradicts the Phase 29 D-10 precedent | |

## Claude's Discretion
- Exact `module.*` code names, loader package layout, and how the library file is placed under `examples/phase30/`.

## Deferred Ideas
- Qualified calls and aliases.
- Cross-module types.
- An entry-header check.
- The `SCHWAY__X` leading-underscore fix.
- MOD-06 identity hashing, deferred to Phase 31.
