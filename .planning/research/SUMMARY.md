# Project Research Summary

**Project:** Schway
**Milestone:** M006 Bounded Byte Utility and Local Modules
**Domain:** Ownership-checked native language toolchain (Go 1.24 stdlib host, C17 via Clang, interpreter oracle, independent validator peers)
**Researched:** 2026-10-10
**Confidence:** MEDIUM-HIGH

<!-- Synthesizer returned this document inline (subagent write refused); orchestrator persisted it (#222 self-heal). Pinned vectors re-run with Python zlib.adler32 on 2026-10-10 by the orchestrator. -->

## Executive Summary

M006 adds three things to Schway: a bounded byte view, a checksum utility that reads one bounded file, and a minimal local module system proven by a second consumer. The checksum is Adler-32 (RFC 1950), computed with checked U64 addition and `% 65521` only. Bytes widen to U64 at the read, so no new scalar type is introduced. Modules are file = module, with explicit imports, an explicit export list, qualified access, an acyclic graph, and whole-program resolution into the single `ast.Program` that every peer already consumes.

The recommended approach keeps every new analysis inside existing trust boundaries. No loan crosses a loop back edge: the buffer is a read-only `ByteView` parameter of a scalar-CFG library function, and the caller holds the loan in straight-line code (acquire, borrow shared, call, release). The FLOW-02 refusal boundary stays unchanged. Bounds checks are dynamic and never elided; an out-of-bounds read is a defect that exits 65 with a fixed stderr line. Byte operations record no per-byte events, and the loop uses `var` plus assignment, so the event capacity does not bound input size. Cross-module call admission reads only the body-blind signature table, which keeps separate compilation (D-13-02b) open without claiming it now.

The main risks are: (1) drift across the six consumers and four independent scalar classifiers when two new operations and a new shape are added; (2) common-mode error in expected answers, mitigated by literal pins from Python `zlib.adler32` and Go `hash/adler32` (both test-only) and by wrong-modulus and off-by-one mutants; (3) import and file-read escape or TOCTOU, mitigated by `os.Root`, a single bounded read with a limit+1 probe, and regular-file checks; (4) exit-class and carry-over debt drift, which is why exit classes and the verify report are fixed first in Phase 28. The maximum-size case must pass on the public route on every engine, and that route is the authority for the cap, not `app verify`.

## Resolved Decisions (orchestrator, auto-followed)

The researchers disagreed on several points. The orchestrator resolved them; these are the synthesized recommendations. Treat them as settled unless a phase discussion records evidence to reopen one. Rejected alternatives are listed so they are not re-proposed silently.

| # | Decision | Resolved recommendation | Rejected alternatives |
|---|----------|-------------------------|-----------------------|
| 1 | Checksum algorithm | Adler-32 (RFC 1950). Per-byte `% 65521`. Compute `B*65536` from 16 checked self-additions (`x = x + x`), so no new multiply or shift operator is needed. Result `B*65536 + A`. Output is bare unsigned decimal followed by `\n`. Oracles: Python `zlib.adler32` and Go `hash/adler32`, both test-only. Answers are pinned as literals, never derived from the implementation. | CRC-32 / POSIX `cksum`: needs xor, shift and a table; reopen only with a bitwise-operator milestone. FNV-1a (xor and wrapping multiply). Fletcher-16/32, BSD `sum`, SysV `sum`. Checked U64 `*` for the scale: replaced by 16 self-additions. Hex output (a new output primitive). Deferred modulus (NMAX batching). |
| 2 | Byte representation and access | No `U8` type. A byte read is widened to U64 (0..255) and read as unsigned. Two new core operations, byte length and byte-at (`OpBytesLen` / `OpBytesAt`), go through all six consumers (check, corevalidate, originvalidate, pathoracle, interp, cgen) with independent re-derivation, and all four scalar classifiers are updated. Surface syntax is decided at phase discussion. | New `U8` scalar with arithmetic. Provisional `loop buffer { next(buffer) }` iterator form (needs a mutable cursor carried across the back edge). |
| 3 | Loops and loans | No loan crosses a back edge. The buffer is a read-only `ByteView` parameter of a scalar-CFG library function. The caller holds the loan in straight-line code. FLOW-02 refusal boundary unchanged, guarded by refused-neighbor fixtures and per-peer seeded faults. | Generalizing loan liveness over back edges (spike 002 shows the producer can, but no independent cyclic peer exists). Per-iteration shared borrow of an owner inside the loop body. Loop-invariant owner carried across the back edge. Iterator protocol. |
| 4 | Bounds checks and events | Dynamic bounds checks only, never elided. An out-of-bounds read is a defect: exit 65, stderr exactly `schway: byte index out of bounds`, stdout empty. Byte ops record no per-byte events. | Static range proof or loop-guard elision. Per-byte events. `assert`-based checks (vanish under `NDEBUG`). Go runtime panic as the out-of-bounds result. |
| 5 | Size cap | Target cap 65,536 bytes. The maximum-size case must pass on every engine through the public route. Phase discussion may lower the cap only on recorded evidence. | 4,096 bytes (rationale was per-byte event volume, removed by decision 4). Cap derived from `app verify` capacity: `app verify` stays U64-only. |
| 6 | File adapter | New copy-local generalization of Phase 23's adapter under `examples/checksum/`: `open(O_RDONLY|O_NONBLOCK)`, `fstat` on the descriptor, `S_ISREG`, read loop with `EINTR` retry and a limit+1 probe, never trusting `st_size`, allocation of `max(1,n)`. Empty file succeeds with checksum 1. Missing, unreadable and non-regular files share one named error class; oversized has its own. Stdout empty on every error; stderr never echoes the path. | Editing the Phase 23 adapter or its receipts. Sizing from `st_size`. Interpreter using `os.ReadFile`. Separate missing/unreadable classes. Schway-owned file handles. Empty as error. |
| 7 | Modules | One file is one module; the `module` header must match its path under the entry directory. No search path, no `..`, no absolute paths, no symlink escape; resolution through Go 1.24 `os.Root`. Explicit imports, no glob or unqualified wildcards; collisions refused. The `export` list is the public surface; visibility re-derived independently by corevalidate. Cycles get a named refusal; diamonds allowed and load once. Whole-program composition after resolution; cross-module admission body-blind via the signature table. Injective C mangling (length-prefixed); single-module C output byte-identical. Diagnostics gain an omitempty `Unit` field. Build identity and receipts hash every module in sorted order. Modules do not establish B1 or DX-06. | Search path / env roots. Relative or `../` imports. Glob, alias, re-export. Separate compilation artifacts. Non-injective `module__name` C naming. Span-level file identity. |
| 8 | Second consumer | A path-only program imports the same library and derives a different pinned output (e.g. `adler32(...) % 65536`). Proof of import rather than copy: single library copy, and a library mutation flips both consumers' pinned outputs. | Two-argument `checksum_verify` (app ABI growth). Two-file tool (two live owners). Thin copy of `checksum`. |
| 9 | Carry-over debt | W2/W3: one exit-class table — checker, corevalidate, originvalidate and entry-resolution refusals exit 2; tool, IO and timeout failures exit 3; engine disagreement exits 4. Reclassify `callgraph.entry_ambiguous` as a source refusal and prune to the closure reachable from `main`. N1: exhausted verify cases carry an outcome digest, not full interpreter evidence. `app verify` stays U64-only. N2 (hosted dual-host receipts) is the last plan of the last phase. N3 folded in when exit codes are named. N4 and R1b not scheduled. | `entry_ambiguous` as operational failure. Degrade-ladder N1 fix. Mid-milestone hosted receipts. Scheduling N4/R1b. |
| 10 | Phase order (from 28) | 28 exit classes, verify report, entry pruning (W2, W3, N1). 29 ByteView and bounds-checked Adler-32 scalar library with a runnable gain. 30 local modules and second consumer. 31 bounded file checksum app and adapter with the exact result matrix. 32 closure: guide, retire frozen `examples/checksum.schway` frontier, N2 hosted receipts. | Pruning in the modules phase. Separate loan-only phase. Checksum app before modules. |

## Key Findings

### Recommended Stack (STACK.md)

No new dependencies. Go 1.24 stdlib (`os`, `os.Root`, `io`, `io/fs`, `path/filepath`, `sort`) and libc/POSIX (`open`, `fstat`, `read`, `close`, `malloc`, `free`) are the whole stack. Clang (C17) remains the native path.

Facts that shape decisions:
- `ast.FuncDecl` has one `Parameter`; `ast.Program` has one `Module` and `Exports`. No `import` token or rule today.
- Scalar-loop mode is gated on a single `U64` parameter at parse time (`scalarLetStart`). Scalar admission refuses move, borrow, foreign call, release and Buffer/Owner copies across back edges. A loop that touches bytes is a new admission shape.
- `Byte` and `Buffer` are executable shapes; `PathToken` is the native entry type.
- Generated C identifiers are confined by `^SCHWAY_[A-Z0-9_]*$` and `schway_value_*` (`TestGeneratedIdentifierNamespacesStayConfined`). Emission is one translation unit.
- The Phase 23 adapter is hard-wired to 1-byte files.

Algorithm (RFC 1950 §8.2): `A = 1 + sum(D_i) mod 65521`; `B = sum of successive A mod 65521`; result `B*65536 + A` (< 2^32, so checked `+` cannot trip). Each byte read as `(uint64_t)data[i]` from `const unsigned char *`.

Pinned literals (Python `zlib.adler32`, re-run 2026-10-10; record command and version beside each literal when first used):

| Input | Decimal |
|-------|---------|
| empty (0 bytes) | 1 |
| one byte `0x00` | 65537 |
| `a` | 6422626 |
| `Wikipedia` | 300286872 |
| 4096 × `0x00` | 268435457 |
| 5552 × `0xFF` (zlib NMAX boundary) | 4052720524 |
| 65,536 × `0xFF` (cap) | 2006388466 |
| 65,537 × `0xFF` (cap+1; must be refused, never printed) | 2273841137 |

Independent check: `python3 -c "import zlib,sys;print(zlib.adler32(open(sys.argv[1],'rb').read()))" FILE`.

### Expected Features (FEATURES.md)

Table stakes: Adler-32 with the boundary table; one path argument via the single-token app route; bare decimal plus newline; empty stdout on every error; fixed path-free stderr for missing/unreadable/directory/non-regular (unreadable control skipped under root); modules with `import`, explicit exports, private-use refused, named cycle/missing/duplicate refusals, header/path agreement; second consumer plus library-mutation control; W2/W3 and N1 fixed before new refusal kinds.

Differentiators (only if cheap): module interface digest in receipts; exported error `data` types across modules.

Anti-features: hex output, filename suffix, stdin, multi-file input, CRC-32/`cksum`, cyclic or glob imports, re-exports, package manager/registry/remote imports, module-level state, env search paths, generic exports, separate compilation artifacts, two-token app ABI.

### Architecture Approach (ARCHITECTURE.md)

Single-file pipeline today: `syntax.Parse` → `ast.Program` → `check.Program` → `core.Program` → peers (corevalidate, originvalidate, pathoracle) → interp and cgen. Four deliberately-copied scalar classifiers (`check`, `corevalidate.hasScalarCFG`, `originvalidate.originHasScalarCFG`, `cgen.hasScalarOperations`) must all learn every new scalar kind, plus the exhaustive-dispatch fixtures.

Components: byte ops and `ByteView` shape; scalar library `adler32(data: ByteView) -> U64`; session-side module loader (`SourceSet`, only IO site); link step (`check.Link` + `callgraph.Prune`); new parallel file-bytes owner shape (Phase 31, Phase 23 matchers frozen); additive omitempty evidence fields (`Function.Module`, `Function.Exported`, `Program.Modules[]`, `Diagnostic.Unit`).

Anti-patterns: iterator syntax; `let` copies in the byte loop; per-byte events; eliding bounds checks; virtual-concatenated source offsets; index-ordinal C names; editing Phase 23 matchers; `entry_ambiguous` as operational; re-reading source after check.

### Critical Pitfalls (PITFALLS.md, resolutions applied)

1. Maximum-size case silently unsupported — verify the cap on the public route on every engine; fix N1 in 28 and test n = 15, 172, 173, 1000.
2. Expected answers derived from the implementation — pinned zlib literals; mutants `% 65520`, `a` init 0, swapped halves, `i <= n`, signed char.
3. Signedness — `unsigned char` storage, explicit `(uint64_t)` widen, `-fsigned-char` mutant must turn 0x80/0xFF fixtures red.
4. Bounds in C — `if (index >= length)` on `uint64_t` before load; `max(1,n)` allocation; mutants under ASan at `-O0/-O3/-flto`.
5. File semantics — limit+1 probe, `O_NONBLOCK`, regular-file only, one outcome class per case, named skip under root, temp-dir FIFOs with deadlines.
6. Common-mode back-edge error — closed admission predicate written first; refused-neighbor fixtures; one seeded fault per peer.
7. Bundling modules with byte machinery — separate phases, plan-count checkpoint on the module phase.
8. Import escape — identifier segments, `os.Root`, escape fixtures (`../`, absolute, symlinks, FIFO, oversized, case-variant).
9. Module graph semantics — cycles of length 2/3, self-import, diamond, private type in public signature, body-independence control.
10. C collisions and ordering — injective mangling, uppercase-collision check, reversed-order build byte-identical.
11. Stale caches/misattribution — closure-digest keys; library errors cite library via `Unit`.
12. Sharing unproven — single-copy static test plus dual-flip mutation.
13. Exit-class ripple — enumerate and pin every asserting test first; never renumber 65/69.
14. Stale hosted receipts — last plan of Phase 32 at exact final revision.
15. Wired-not-run — every control seen to fail; no zero-match `-run` patterns.
16. Debt growth — only N1, N2, W2, W3 (N3 folded); public-route audit before closure.

## Implications for Roadmap

Numbering continues at 28. No two enabling phases run consecutively without a runnable gain (AGENTS.md).

### Phase 28: Exit classes, verify report, entry pruning (W2, W3, N1)
One exit-class table (2 source refusal / 3 tool-IO-timeout / 4 disagreement) used by session, cmd and native; `build`/`check`/`run` agree on the corevalidate-refusal fixture; native and interpreter engines agree on `foreign_origin_omitted`; `entry_ambiguous` → exit 2 with pruning to `main`'s reachable closure (byte-identity preserved for fully-reachable programs); N1 outcome digest so `native.event_capacity_exhausted` survives at n = 15, 172, 173, 1000 and 16 simultaneous exhausted cases; N3 constants. Runnable gain: libraries with unused functions now run; `app verify sum_to_n` names the capacity reason at n = 1000. Research: standard.

### Phase 29: ByteView and bounds-checked Adler-32 scalar library
`ByteView` shape; parser gate widened; `OpBytesLen`/`OpBytesAt` in all six consumers and four classifiers; `adler32(data: ByteView) -> U64` with no per-byte events; runnable gain on both engines against pinned literals; OOB exits 65; refused-neighbor fixtures pinned. **Spike required:** ByteView through `OpCall`/originvalidate, parser lookahead, release-while-loan-live in three layers, interpreter cost at 65,536 iterations, no evented stores in the loop.

### Phase 30: Local modules and second consumer
`import` token/rule/formatter; resolution via `os.Root` with named refusals (path mismatch, cycle, not exported, collision, missing, duplicate); whole-program composition with additive fields; body-blind cross-module admission with guard test; injective mangling and single-module byte identity; `Diagnostic.Unit`; sorted module hashing in identity/cache/receipts; second consumer with single-copy test and dual-flip mutation. **Deeper research + spike required.**

### Phase 31: Bounded file checksum application and adapter
`examples/checksum/` adapter (copy-local generalization); exact result matrix (empty → 1, `0x00` → 65537, `Wikipedia` → 300286872, 65,536 × 0xFF → 2006388466, 65,537 → oversize class, missing/unreadable/non-regular → shared class, alloc fault → own class with cleanup); independent interpreter file read; public CLI path input; cap case on every engine via public route; Phase 23 untouched. Research: moderate.

### Phase 32: Closure
Clean-checkout guide; retire `examples/checksum.schway` frontier with dated amendment; pre-closure public-route audit; N2 hosted macOS + Ubuntu receipts at the exact final revision as the last plan. Research: standard.

### Research Flags
- Needs research: 29 (spike), 30 (spike + deeper), 31 (moderate).
- Standard: 28, 32.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM-HIGH | Algorithm and vectors HIGH (RFC 1950; zlib run locally twice). Integration claims from reading source, not prototyping. |
| Features | MEDIUM-HIGH | Checksum conventions HIGH; module-system comparisons from training knowledge (MEDIUM). |
| Architecture | HIGH repo facts / MEDIUM analogies | Several behaviors re-run on a scratch build. |
| Pitfalls | MEDIUM-HIGH | Repo grounding HIGH; C UB, `os.Root`, CVE references from memory. |

### Gaps to Address
- Surface syntax for byte reads and imports (Phase 29/30 discussion) — operation names fixed, surface not.
- Exit codes for file-error classes (Phase 31) — choose 2/3/65 per outcome and record in the single exit-class table.
- Cap evidence (Phase 29 spike) — measure interpreter/native cost at 65,536; lower only on recorded evidence.
- Second consumer input route before Phase 31 (Phase 30 discussion).
- Public CLI path input through `app run` (Phase 31) — today it takes bounded U64.
- Module caps (Phase 30) — count/size/depth not settled; each needs exact-N / N+1 controls.
- `Diagnostic.Unit` reach (protocol, debugmap, human output, repair patches).
- `os.Root` symlink and macOS case-insensitivity behavior — confirm by spike.
- Modules must not be described as B1/DX-06 or separate compilation.

## Sources

Primary: repo source and planning docs read 2026-10-10 (see STACK/FEATURES/ARCHITECTURE/PITFALLS for paths); RFC 1950 §8.2 (https://www.rfc-editor.org/rfc/rfc1950); Python `zlib.adler32` executed locally.
Secondary: Go `hash/adler32`; Go 1.24 `os.Root`; module-system comparisons; Rust NLL/#25860; ISO C17 6.5.6/6.2.5; Clang UBSan.
Tertiary: zlib CVE-2022-37434; POSIX `cksum` (not load-bearing).

---
*Research completed: 2026-10-10*
*Ready for roadmap: yes*
