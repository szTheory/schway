# Feature Landscape

**Domain:** Bounded checksum file utility + minimal local module system (Schway M006)
**Researched:** 2026-10-10
**Confidence:** MEDIUM-HIGH (checksum conventions and boundary behavior HIGH; module-system comparison from training knowledge of Go/Rust/Zig/OCaml/Hare/Odin/Gleam/Roc, MEDIUM; repo grounding HIGH)

## Repo grounding (read, not assumed)

- `examples/checksum.schway` is a provisional refused fixture (`foreign C` `checksum_read`, a `loop buffer { next(buffer) }` form). Its loop syntax is NOT the admitted syntax; M005 programs use `while i < n { ... }` (`examples/fizzbuzz.schway`). Treat it as a witness to be rewritten, not a spec.
- The parser already accepts `module a.b.c` and `export { fn name }` (`internal/compiler/syntax/parser.go:158,258`; lexer keywords `module`/`export`; `format.go` has an `export` context that also handles `type`). There is NO `import` token or parse rule. Every shipped program is single-file and starts with `module x.y` + `export { fn main }`.
- File input already exists as `PathToken` (4,096-byte bound) -> audited C adapter -> owned bounded buffer with generated cleanup (`examples/phase23/README.md`, `file_byte.schway`; typed-error cleanup in `examples/phase24/error.schway`). The app route takes ONE argv token; output is auto-printed `U64` + newline or explicit `write_*`, shared 65,536-byte output ceiling; defects exit 65 (`examples/phase27/README.md`).
- Call graph is already a guaranteed DAG with three-color cycle refusal; caller admission never reads a callee body (PROJECT.md, M002 Phase 07). Emission is a single translation unit (`emitProgram`). This is exactly the substrate a whole-program local-module system needs.
- PRODUCT-ROADMAP (Mid term row, rec. 2): "Second consumer imports a library without copying it; malformed/truncated/oversized input has specified behavior"; "A local module system need not introduce separately verified binary contracts."

---

## Decision 1: Checksum CLI contract

### Landscape (conventions)

| Tool | Algorithm | Output | Input | Notes |
|------|-----------|--------|-------|-------|
| POSIX `cksum` | CRC-32 (poly 0x04C11DB7) over bytes **plus length bytes appended**, complemented | `<crc-decimal> <byte-count> [file]` | files or stdin | Decimal; needs XOR/shift/table or bitwise ops and a length-fold loop. Empty file -> `4294967295 0`. `a` -> `1220704766 1` (verified on this host). |
| GNU/BSD `sum` | 16-bit rotate-right+add (BSD) / plain byte sum mod 2^16 (SysV) | decimal + blocks | files/stdin | Legacy; BSD vs SysV disagree. Not recommended as a spec. |
| `sha256sum`/`md5sum` | cryptographic | `<hex>  <file>` (two spaces); `-c` verify mode | files/stdin | Hex + filename is the "verification list" format. Far beyond Schway's current ops (32-bit rotates, tables). |
| zlib `adler32` | Adler-32 (RFC 1950) | integer; typically shown hex | library | `a = 1 + sum(bytes) mod 65521`, `b = sum of successive a mod 65521`, `result = b*65536 + a`. Needs only add, mod, and one scale. |

### Recommendation (one per point)

**Algorithm: Adler-32 (RFC 1950).** Reasons: it is the only widely-specified checksum computable with operations Schway already has (checked `+`, `%`) plus byte indexing and one constant scale (`b * 65536`). CRC-32/`cksum` would force XOR, shifts, and either table data (aggregates, out of scope) or 8x bit loops, each a new checked-operator family with its own interpreter/native/peer evidence. That violates the milestone's "admit only the byte operations the checksum needs." Weakness: Adler-32 is weak for short inputs (documented in RFC 3309); acceptable because the goal is a specified, checkable checksum, not integrity against adversaries. Say so in the README ("not cryptographic, not CRC").

**Minimal new arithmetic:** the scale `b * 65536` needs a checked U64 multiply OR can be avoided by `b <<`-free alternatives. Recommend admitting checked U64 `*` (same family as `+`: overflow -> exit 65, same four-layer evidence). Max-input check: with a 65,536-byte cap, `b < 65521` and `a < 65521` always, so `b*65536 + a < 2^32`, no overflow in practice; the overflow path is still required evidence because the operator is general. If the roadmap wants zero new operators, the fallback is printing `a` and `b` as two decimals (non-standard; loses third-party comparability) - NOT recommended.

**Input: one path argument, via existing `PathToken`.** Matches `cksum file`, fits the one-token app ABI, avoids stdin (stdin would need a second input channel in the app route and would conflict with the "evidence capture is separate from ordinary run" rule). No multi-file, no `-` stdin, no flags.

**Output: bare decimal on stdout + `\n`, nothing else.** This is what the existing `U64`-return auto-print already does, so no new text surface is needed. No filename suffix (one file, caller already knows it; avoids filename-escaping questions that sha256sum has to solve with `\` prefixes). Decimal is also the `cksum` convention. Independent check: `python3 -c "import zlib,sys;print(zlib.adler32(open(sys.argv[1],'rb').read()))" FILE` (pinned fixture answers computed here: empty -> `1`; `a` -> `6422626`; `Wikipedia` -> `300286872`; 65,536 x 0xFF -> `2006388466`). Hex is a differentiator, not table stakes.

**Exit codes / stderr (align with the existing app contract and with W2/W3):**

| Outcome | stdout | exit | stderr (exact, fixed text) |
|---------|--------|------|-----------------------------|
| Success (incl. empty) | `<n>\n` | 0 | empty |
| Missing / permission denied / is a directory / not a regular file | empty | named typed-error class, recommend 66 (sysexits EX_NOINPUT) if the exit-class split from W2/W3 can carry it; else 65 as for other app defects | one fixed line, e.g. `schway: checksum: cannot read file` |
| Larger than limit | empty | same class as above or a distinct "too large" 65 | `schway: checksum: file exceeds 65536 bytes` |
| Path token too long / invalid | empty | existing app refusal behavior (exit 65 `schway app: ...`) | existing text |

Rules: stdout is empty on every error (never a partial/wrong number); error text is fixed and contains NO attacker-controlled path bytes (log-injection footgun; the Phase 23 adapter already bounds diagnostics to 128 bytes). Open item for the roadmap: pick ONE exit number per class and pin it; the exit-class conflation in W2/W3 is precisely this decision, so decide them together.

**Size limit: 65,536 bytes.** Reuses the shared ceiling number (one mental constant), keeps `a`/`b` far from overflow, and keeps the interpreter oracle's event budget meaningful. Note the N1/event-capacity issue: a loop that visits 65,536 bytes produces tens of thousands of events; `app verify` already reports `unsupported` past its capacity. The checksum verify route must therefore use a bounded set of replay cases (empty, 1, small, boundary-adjacent) and rely on native-vs-pinned-answer for the large cases. This is a design constraint the roadmap must schedule, not a surprise to find in execution.

### Adversarial pass on Decision 1

- "Adler-32 is a poor checksum, why not cksum?" Counter: cksum drags in bitwise/table features; the milestone value is the byte-view + module + error path, not the algorithm. Keep the algorithm swappable inside the module (a later `crc32` export is a clean second use of the same library).
- "Bare decimal loses the filename, can't feed `-c`." True; sha256sum-style verification lists are a future differentiator, not needed for one-file input.
- "Reusing 65,536 conflates input and output limits." Accepted: they are independent facts that happen to share a number; name them separately in code (this is exactly debt N3, duplicated capacity literal).

---

## Decision 2: Exact boundary behavior

| Case | Required behavior | Independent answer source |
|------|-------------------|---------------------------|
| Empty file (0 bytes) | success, prints `1`, exit 0 (loop body never runs; no OOB; allocation of zero bytes must be defined: adapter returns an empty owned buffer or a distinct empty state, and cleanup still runs exactly once) | `zlib.adler32(b'')` = 1 |
| 1 byte (`a`) | `6422626` | zlib |
| Typical (`Wikipedia`) | `300286872` | zlib / RFC 1950 example |
| Exactly 65,536 bytes of 0xFF | success `2006388466` | zlib |
| 65,537 bytes | refused, empty stdout, named "exceeds limit" outcome, buffer never partially processed, no leak | expected = refusal (zlib value `2273841137` must NOT appear: wrong-result control) |
| Missing path | named read failure, empty stdout, typed-error cleanup path exercised | shell `cat` fails |
| Unreadable (mode 000) | same class as missing (do not distinguish ENOENT/EACCES in output; distinguishing leaks filesystem layout and adds surface). Skip in tests when running as root | |
| Directory | same class; `open` succeeds on a directory on POSIX but `read` returns EISDIR, so the adapter must check/handle it explicitly | |
| FIFO / `/dev/zero` / `/dev/stdin` | refuse non-regular files (fstat after open; open with `O_NONBLOCK` to avoid blocking on a FIFO). Otherwise `/dev/zero` is a hang/OOM and a FIFO is a hang | |
| File changes during read | Defined as: the adapter reads at most max+1 bytes in a read loop to EOF into a buffer it owns, and the checksum is over exactly the bytes returned; no `stat`-size-then-read (TOCTOU). Result is a checksum of some prefix/mixture, never a memory error; document "no snapshot guarantee". Growth past max during read -> oversized outcome. Not tested with races (nondeterministic); tested by code review of the single-read-loop structure plus a truncate-before-read unit control. | |
| Bytes >= 0x80 | counted as unsigned 0-255 (a signed-char bug would change the answer) | the 0xFF fixture is the control |

Mutation controls that must exist (wrong-result class): off-by-one on the loop bound (skip last byte / read one past), `% 65521` replaced with `% 65520`, `a` initial value 0 instead of 1, swapped `a`/`b` in the result, unsigned vs signed byte. Each must move at least one pinned answer above. Out-of-bounds control: an index program with `i == len` must fail on interpreter and native with the SAME named outcome (bounds-checked index, defined exit).

### Adversarial pass
Hardest footgun is "oversize detected after allocating". Require the adapter to cap the allocation at max+1 before read, never trust `st_size`. Second is empty-file zero-size `malloc(0)` returning NULL being misreported as a read failure.

---

## Decision 3: Local modules

### Comparison (training-knowledge level; no new dependency or web source needed)

| Language | Unit | Import | Visibility | Cycles | Takeaway for Schway |
|----------|------|--------|-----------|--------|---------------------|
| Go | package = directory | `import "path"`, qualified `pkg.Name` | capitalized = exported | refused | Qualified access; compiler-enforced acyclic graph |
| Rust | `mod`/file tree + crate | `use`, `mod x;` | `pub` default-private | allowed within a crate | Too rich (re-exports, glob, nested) |
| Zig | file = struct | `@import("file.zig")` | `pub` | allowed lazily | File-as-module, explicit `pub`; imports are values |
| OCaml | file = module, `.mli` sig | qualified `M.f`, `open` | signature file | refused | Separate interface file is the cleanest separate-compilation story, but heavy |
| Hare | file/dir = module | `use mod;` qualified | `export` keyword | refused | Closest to Schway: explicit `export`, qualified use |
| Odin | package = directory | `import "dir"`, qualified | `@private` | refused | Qualified by default |
| Gleam | file = module | `import a/b` qualified `b.f` | `pub` | refused | Tiny, explicit, no cycles; best minimal model |
| Roc | module file with `exposes` header | explicit imports | exposes list | refused | Header-declared interface (like Schway `export {}`) |

Convergent minimum: file = module, explicit export list, explicit import, qualified use, acyclic graph. Nobody small and recent (Hare, Gleam, Roc, Odin) permits cycles or unqualified wildcard imports.

### Table stakes (module system)

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| One file = one module, declared `module a.b` | Already parsed; path must agree with file location (`a/b.schway` under the entry root) or be refused (declared-vs-actual mismatch is the classic confusing bug) | LOW | Existing `modulePath()` |
| `import a.b` declaration at file top, explicit, no side effects | Gleam/Hare convention | LOW | New token + parse rule + formatter canonical form + `lang format` idempotence evidence |
| Qualified access only (`b.adler32(...)`) | Removes name collision as a class; no `open`/glob/unqualified; no aliasing in v1 | LOW-MED | Collision between a local fn and imported name becomes impossible |
| Explicit exports; everything else private; use of an unexported name refused with a named, repairable diagnostic | "explicit visibility" is in the milestone goal; `export { fn ... }` already exists | MED | Existing `export` currently only marks `main`; extend semantic meaning. `type` in export lists already formats (needed so a consumer can name the error type of an exported fallible fn) |
| Import cycle refusal with a named code and the cycle path | Go/Gleam/Hare behavior; M002 callgraph already refuses call cycles | LOW-MED | Reuse three-color DFS pattern at module level |
| Missing module / duplicate module path / module-name-vs-path mismatch refused deterministically | Basic hygiene | LOW | |
| Whole-program resolution: modules merge into one core program, one translation unit | Smallest thing that works; `emitProgram` already single-TU | MED | Resolution roots only at the entry file's directory tree; no `..`, no absolute paths, no symlink escapes (bounded build authority per Constraints: Security) |
| Independent validation: peers (`corevalidate`/`originvalidate`) see merged, qualified-name functions and must independently check that a caller only uses exported callees | Repo's evidence style | MED-HIGH | This is the real cost, not the parser |

### Differentiators (defer unless cheap)

| Feature | Value | Complexity | Notes |
|---------|-------|-----------|-------|
| Module interface digest (hash of exported signatures) recorded in the receipt | Directly serves D-13-02b: separates "interface changed" from "body changed" and is the natural seam for separate compilation later | LOW-MED | Signature summaries are already digest-bound (M002 Phase 07). Cheap to expose; high future value. Recommended as the single "not foreclosing" investment, only if it costs < one plan |
| `import a.b as x` alias | Convenience | LOW | Defer: aliasing reintroduces collisions |
| Exporting `type`/`data` and abilities across modules | Needed only if the second consumer must name the library's error type | MED | Include only the minimum: exported `data` error type |

### Anti-features (do NOT build)

| Anti-Feature | Why Avoid | Instead |
|--------------|-----------|---------|
| Cyclic imports | Breaks DAG assumption of liveness/summary derivation (TRU-04, D-09-22) | Refuse with a named code |
| Unqualified/glob import, re-exports, nested submodules, conditional imports | Search-space growth; hurts AI-authored predictability | One flat qualified import |
| Package manager / registry / versions / remote imports / `Schway.toml` | Out of Scope in PROJECT.md | Local relative-to-root paths only |
| Separately compiled artifacts, binary interface files, user-declared contracts | DX-06/B1 needs a user-declared contract field; modules alone do not establish it (PRODUCT-ROADMAP capability map) | Whole-program merge now; keep exports as the signature surface |
| Generic/polymorphic exports, macros, module-level state | Not needed; generics are Out of Scope | Plain monomorphic `fn` only |
| Module-level `let`/constants as exports | Initialization-order and cycle semantics | Export only functions (and error `data`) |
| Implicit module search path / env var | Ambient authority | Explicit root = entry file's directory |

### Smallest viable module system (one recommendation)

Gleam/Hare-style: file = module (`module checksum.adler`), `export { fn adler32; type ReadError }` is the whole public surface, `import checksum.adler` then `adler.adler32(buf)`, root = entry file's directory, acyclic, qualified-only, whole-program merge into the existing single-TU pipeline. This does not foreclose separate compilation because (a) callers are already admitted from signatures only, (b) the export list IS the interface, and (c) qualified names + acyclic graph are exactly what a later per-module compilation needs. What it does NOT claim: contract blame (DX-06) stays tied to a user-declared contract.

### Adversarial pass on Decision 3

- "Whole-program merge is just copy-paste with extra steps." Counter: the evidence differs, since the consumer must be rejected when it touches a private name, cycles must be refused, and the receipt must bind the module files' content. Make "second consumer's receipt names both module files by hash" a requirement; otherwise the claim "imports without copying" is unobservable. A negative control: delete/modify the library file and the consumer's receipt/answer must change.
- "Qualified-only will annoy people." For AI-authored, human-audited code that is a feature: every call site shows its origin.
- "Should export lists remain `fn main` for the entrypoint?" Yes: `main` stays the only entry; library modules have no `main`; entry vs library roles must be a checked distinction (refuse `schway build` on a library, refuse importing an entry).
- Risk: the `module` name vs path mismatch rule and file-resolution rules are where diagnostics will multiply. Cap the diagnostic set to ~6 codes (not found, cycle, private use, duplicate, mismatch, import-of-entry).

---

## Decision 4: The second consumer

Candidates:

| Candidate | Value | Cost | Evidence value |
|-----------|-------|------|----------------|
| A. `checksum_verify`: reads a file and an expected checksum, prints `ok` / exits nonzero on mismatch (the `sha256sum -c` analog) | Highest: real tool; exercises the library from a different `main` with different control flow and a different error class (mismatch) | Needs two inputs. The app ABI takes ONE argv token today; (path, expected) requires ABI growth (arity-N) or a second token form | Strong |
| B. Two-file tool (checksum two files, print both / equal?) | Shows reuse twice | Same two-argument problem; also requires two live owned buffers in one frame (resource carry) | Medium |
| C. Path-only consumer using a DIFFERENT export of the same module (e.g. `byte_sum`/`length`-style export, or `adler_ok` comparing against a source-embedded pinned value) | Proves import, privacy, and non-copy with zero ABI change | Minimal | Medium: real reuse, weaker "tool" story |

**Recommendation: C now, with A as the stated target only if the one-token ABI is extended for another reason.** Concretely: the library module exports `adler32(path) -> U64` fallibly (owning buffer, loop, cleanup inside); consumer 1 is `checksum` (prints it); consumer 2 is a path-only program that imports the same module and returns a derived judgment from the same library result (for example `checksum_verify` with the expected value fixed in source, or a "matches known value" pinned to an embedded literal). The evidence is the same: second `main` file, no copied body, receipt binds the library hash, private-name and cycle refusals. Do not extend the app ABI to two tokens inside this milestone: it drags in arity-N and widens scope with no module-system evidence gain. If the roadmap prefers a real tool story, the one-token trick is for the second consumer to reuse the same path and an `expected` baked from a second file path - no, that is B; reject it.

Adversarial: C looks contrived. Accept: its job is to prove reuse, not product value; name it honestly in the README as a reuse witness.

---

## Decision 5: Carry-over debt classification (from `.planning/milestones/M005-MILESTONE-AUDIT.md`)

| Item | What | Class for M006 | Reason |
|------|------|----------------|--------|
| N2 | Re-acquire two-host hosted receipts at current revision | **Table stakes** (closing gate) | Every M005/M006 claim of cross-host evidence is only valid at its SHA; M006 adds new ABI/adapter code that needs fresh hosted receipts anyway. Do as the final phase, not first |
| N1 | `app verify` report exceeds 65,536 bytes at n >= 173 and loses the named `native.event_capacity_exhausted` reason | **Table stakes, and it is a hard dependency** | The checksum program iterates up to 65,536 bytes; its app-verify cases will hit the same report-size/event-capacity wall far sooner than `sum_to_n`. Without N1 the verify route cannot report the capacity reason for the new program. Fix once (shrink per-case evidence for exhausted cases / keep reason before size check) and reuse |
| W2 | `build` reports origin refusal as exit 2 but corevalidate refusal as 3 | **Table stakes** | Module refusals (private use, cycle, missing module) will flow through the same admission helper; exit classes must be right before adding more refusal kinds, and the checksum error exit classes in Decision 1 depend on the split |
| W3 | `run --engine=native` vs interpreter exit class for the same refusal | **Table stakes (same fix as W2)** | One shared exit-class mapping; the audit marks it pre-existing, but do it in the same plan to avoid a second pass over `runBuild`/native run mapping |
| N3 (duplicated capacity literal), N4 (guide n <= 14 note), R1b, D-27.1-10 dispositions | | **Not table stakes** | Keep tracked. N3 becomes cheap to fix incidentally when the 65,536 constant is named for the checksum cap; opportunistic only |

---

## Feature Dependencies

```
Checked U64 *  (only if Adler scale used) -> Adler-32 result composition
Byte view + length + bounds-checked index -> Adler loop over buffer
Existing: PathToken + owned bounded buffer + typed-error cleanup (Phase 23/24) -> checksum_read
Borrow of owned buffer across scalar while loop (loan state live over back edge: currently REFUSED) -> Adler loop   <-- biggest new checker work
Oversize/empty/special-file handling in C adapter -> exact error outcomes
W2/W3 exit-class split -> checksum error exit codes -> module refusal exit codes
Parser `import` + module resolution -> export visibility -> cycle refusal -> second consumer
N1 -> app verify for checksum (event capacity at 65,536 bytes)
All of the above -> N2 hosted two-host receipts (last)
```

The one real hazard: the Phase 26 rule "resource/loan state across loop back edges stays refused." The checksum loop borrows an owned buffer for the whole loop. Either admit a narrowly-scoped "shared borrow of an owned buffer live across a scalar loop, no mutation, no move" with its own independent fixed point and refusal controls, or restructure so the loop sees only a scalar length plus an index-read primitive taking the buffer by shared borrow each iteration (loan starts and ends within one iteration, no loan across the back edge). The second is smaller and keeps the Phase 26 refusal intact; recommend it as the first attempt, the broader admission only if it fails.

## MVP Recommendation (ordering for roadmap)

1. Byte primitives: length, bounds-checked index (per-iteration borrow), unsigned byte, plus checked `*` if needed; each with interpreter/native/peer evidence and OOB + wrong-result controls.
2. Checksum program in a single file (no modules yet) with full boundary table, adapter hardening (cap max+1, regular-file check, empty buffer), exact exit/stderr, independent zlib-derived answers. Include N1 and W2/W3 here since the program is the first consumer of both.
3. Local modules: `import`, qualified access, export visibility, cycle/missing/private refusals, move checksum into a library module, `checksum` becomes the first consumer.
4. Second consumer + content-bound receipt of library hash + mutation control ("library changed -> consumer answer changes").
5. N2 two-host receipts at the final SHA.

Defer: hex output, filename suffix, stdin, multi-file, CRC-32/`cksum` algorithm, two-token app ABI, aliases, interface digest (unless free), separate compilation, general strings.

## Sources

- Repo (HIGH): `/Users/jon/projects/schway/.planning/PROJECT.md`, `.planning/PRODUCT-ROADMAP.md`, `examples/checksum.schway`, `examples/fizzbuzz.schway`, `examples/phase23/` and `examples/phase24/` (adapter, README, PathToken bound), `examples/phase27/README.md` (exit 65, output ceiling, error text), `internal/compiler/syntax/parser.go` (module/export parsing), `.planning/milestones/M005-MILESTONE-AUDIT.md` (N1-N4, W2, W3 definitions).
- Executed locally (HIGH): `python3 zlib.adler32` for pinned fixture values; `/usr/bin/cksum` output for `a`.
- RFC 1950 (Adler-32 definition) and RFC 3309 (short-message weakness), POSIX `cksum` specification: recalled from training, not re-fetched this session (MEDIUM).
- Module-system comparisons (Go, Rust, Zig, OCaml, Hare, Odin, Gleam, Roc): training knowledge, not re-verified against current docs (MEDIUM); conclusions rest on convergent design, not on any single language's latest syntax.
