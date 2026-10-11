# Technology Stack: M006 Bounded Byte Utility and Local Modules

**Project:** Schway
**Researched:** 2026-10-10
**Scope:** Additions/changes only. Everything in PROJECT.md "Validated" is out of scope.
**Overall confidence:** MEDIUM-HIGH. Algorithm and vectors are HIGH (RFC 1950, vectors executed with Python zlib on 2026-10-10). Compiler integration claims are MEDIUM: grounded by reading source, not by prototyping.

## Headline Decision

**No new dependencies. No new Go module, no new C library, no new backend.** The milestone adds
(1) Adler-32 as the one checksum, (2) two read-only byte builtins that widen bytes to U64,
(3) one generalized copy-local C file adapter, (4) a stdlib-only module loader that flattens
imports into the single `ast.Program` every existing peer already consumes.

Facts about the existing code that shape every decision below:

- `ast.FuncDecl` has exactly one `Parameter`; `ast.Program` has one `Module` and `Exports` of `Kind`+`Name`
  (`internal/compiler/ast/ast.go`). There is no `import`. The parser's `module` is a path only (`syntax/parser.go:158,258`).
- Scalar-loop mode is gated at parse time on `parameterType.Constructor == "U64"` (`syntax/parser.go` ~L362, `scalarLetStart`).
  Scalar admission (`check/scalar.go`) classifies `move`/`borrow`/`foreign_call`/`release` and Buffer/Owner copies as
  `owner`/`loan`/`resource` authority and refuses them across back edges. So a loop that touches a Buffer is a
  **new admission shape**, not a tweak.
- `Byte` and `Buffer` are already executable shapes (`check/check.go` `executableShape`, `payloadFieldShape`:
  Byte = 1 byte `unsigned char`, Buffer = `SCHWAY_BUFFER`). `PathToken` is already the native entry type.
- Generated C identifiers are confined by `cName` -> `^SCHWAY_[A-Z0-9_]*$` and `cLocal` -> `schway_value_*`
  (`cgen/cgen.go:1578`), enforced by `TestGeneratedIdentifierNamespacesStayConfined`. `emitProgram` is one translation unit.
- The M004 file adapter (`examples/phase23/adapter.{h,c}`) is hard-wired to 1-byte files: `EmptyFile` is an error,
  `TooLong` is anything over 1 byte, owner = `{unsigned char *data; uint64_t length}` with `_Static_assert`s on 64-bit pointers.

## Recommended Stack

### Decision 1: Checksum algorithm -> **Adler-32 (RFC 1950 sec. 8.2), printed as unsigned decimal**

| Candidate | Verdict | Reason |
|-----------|---------|--------|
| **Adler-32** | **CHOOSE** | Needs only `+` and `% 65521` per byte, both already admitted (checked U64 `+`, `%`). Real, standard, externally checkable. |
| CRC-32/IEEE, POSIX `cksum` | Reject now | Needs xor + shift (or a 256-entry table built with xor). Neither exists. Multi-operator milestone before any user value. Reopen when a bitwise-operator milestone exists. |
| FNV-1a | Reject | Needs xor and wrapping multiply. U64 `*` would be checked (overflow = exit 65), the opposite of FNV's definition. Also no standard CLI to verify. |
| Fletcher-16/32 | Reject | Arithmetic-only, but no ubiquitous reference tool/library; weaker than Adler for no gain. |
| BSD `sum` | Reject | Needs rotate-right (shift + or). |
| SysV `sum -s` | Reject | Arithmetic-only but tool behavior differs across macOS/Linux; weak and obscure. |

**Definition (RFC 1950 sec. 8.2):** `A = 1 + D1 + ... + Dn (mod 65521)`; `B = (1+D1) + (1+D1+D2) + ... (mod 65521)`;
`ADLER32 = B * 65536 + A`. 65521 is the largest prime below 65536. Initial `A=1, B=0`. Source: https://www.rfc-editor.org/rfc/rfc1950 (HIGH).

**Algorithm in admitted operators only:**

```
a = 1; b = 0; i = 0
while i < len(data):
    a = (a + data[i]) % 65521      # + and % are admitted
    b = (b + a) % 65521
    i = i + 1
result = b doubled 16 times, then + a   # B*65536 via 16 checked additions x = x + x; no multiply needed
```

Reducing mod 65521 on every byte (not batched like zlib's NMAX=5552) removes any dependence on U64 overflow reasoning.
Max `b` < 65521 so `b * 65536 + a` < 2^32; checked `+` can never trip. Final output via the existing `write_u64` + `write_newline`.
No hex formatting is added (that would be a new output primitive).

**Interop / verification (all executed 2026-10-10 with Python `zlib.adler32`, HIGH):**

| Input | Decimal | Hex |
|-------|---------|-----|
| empty (0 bytes) | 1 | 0x00000001 |
| `A` (1 byte) | 4325442 | 0x00420042 |
| `abc` | 38600999 | 0x024D0127 |
| `Wikipedia` | 300286872 | 0x11E60398 |
| `\xff` (1 byte) | 16777472 | 0x01000100 |
| 4096 x `\x00` | 268435457 | 0x10000001 |
| 5552 x `\xff` (zlib NMAX boundary) | 4052720524 | 0xF18F9B8C |


External check commands: `python3 -c "import zlib,sys;print(zlib.adler32(open(sys.argv[1],'rb').read()))" FILE`.
Test oracle inside Go tests: stdlib `hash/adler32` (tests only; the interpreter itself must execute the Schway
program so the oracle stays independent of the thing under test). No `cksum`/`crc32` match this output; document that so no one
compares them.

**Adversarial pass:** (a) Adler-32 is weak on short inputs and is not a security hash. Mitigation: name the output
`adler32` in docs/README, never "checksum" unqualified in a security context. (b) "Real-world usefulness" is lower than
CRC-32. Counter: usefulness is dominated by being verifiable against zlib on every dev machine; CRC-32 would force
xor/shift semantics, `<<` overflow rules, and 2+ more operators through all six peers. (c) Pinning B*65536 via 16 doublings looks
odd; counter: it is derived entirely from existing evidence-covered operators, and a comment in the library source states why.

### Decision 2: Byte representation -> **bytes widen to U64 at the read; no new scalar type**

| Option | Verdict |
|--------|---------|
| Bytes read as U64 (`0..255`) | **CHOOSE** |
| New `U8` scalar with arithmetic | Reject: needs U8 `+`, conversions, a second overflow story, and new scalar-loop state typing in three independent peers. |

- The existing `Byte` shape stays an opaque payload/FFI shape; M006 does not add arithmetic to it.
- C17: `(uint64_t)data[(size_t)i]` where `data` is `const unsigned char *` (same type as the Phase 23 owner; `unsigned char` is
  the aliasing-safe character type, and `uint8_t` is the same type on Clang targets). The explicit cast to `uint64_t` before any
  arithmetic avoids integer-promotion-to-`int` surprises. Length stays `uint64_t` (matches Phase 23's `_Static_assert(sizeof(void*)==8)`);
  convert to `size_t` only at the subscript.
- Interpreter: `[]byte` element -> `uint64`, same widening.

### Decision 3: Byte view / length / indexing -> **two call-form builtins, dynamic bounds check, defect exit**

Add exactly two expression builtins (parsed like the existing `write_u64(...)` call-forms; no lexer change for `[` `]`, which would be a larger syntax and formatter change):

| Builtin | Type | Semantics |
|---------|------|-----------|
| `buffer_len(b)` | `-> U64` | Length of the borrowed buffer (non-negative, `uint64_t`). |
| `buffer_byte(b, i)` | `-> U64` | Byte at index `i` (0..255). `i >= len` is a **defect**, not UB. |

Rationale for call-form over `b[i]`: builds on the already-used builtin path (`isWriteCallStart`, `scalarWriteStatement`), keeps formatter and token surface
unchanged, and PRODUCT-ROADMAP states unsupported syntax is "not settled", so the cheap reversible form is right. Index sugar can be added later without changing core ops.

- **New core ops** (name suggestions): `OpBufferLen`, `OpBufferByteAt` in `core.OperationKind` and `core.Operations()` (`core/core.go` ~L859); all six consumers (`check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, `cgen`) must handle them, each deriving independently as the activation map requires.
- **Bounds-check strategy:** always dynamic. Emitted C: `if (idx >= len) { fputs("schway: byte index out of bounds\n", stderr); exit(65); }` then the load. Same `exit 65` and message class as the existing "U64 addition overflow" defect (`examples/phase26`, hosted receipt: exit 65 with exact stderr). The interpreter emits the identical defect text. Do **not** try static range proof in M006 (`i < buffer_len(b)` loop-guard elision is an optimization; it would add an unproven analysis and a mutation target for no consumer need).
- **Borrow rule:** `buffer_len`/`buffer_byte` take a *shared borrow* of a Buffer-shaped place and never move/consume it. A loop whose body only does shared-borrow reads of an owner that is not moved, released, or reassigned in the loop is admissible; any move/release/exclusive borrow of that owner inside the loop stays refused (existing `scalarAuthorityCategory` refusals remain for those kinds).
  This is the one genuinely hard checker/validator change; it needs a "loan taken before loop, not carried as mutable state, read-only per iteration" proof derived independently in `check`, `corevalidate`, and `originvalidate`. Flag for deep phase research.
- **Library function shape:** `fn adler32(data: Buffer) -> U64` has a Buffer parameter, but scalar mode is currently gated on a `U64` parameter (`parser.go`, `scalarLetStart` condition). Parser gate and `check/scalar.go` admission must be widened to "single parameter U64 **or** borrowed Buffer", with the Buffer treated as a read-only loan, not scalar state.

### Decision 4: File input -> **new copy-local, generalized adapter; reuse the Phase 23 contract pattern, not the Phase 23 file**

Phase 23's adapter cannot serve checksum (1-byte only; empty = error). Do not edit it (its expected JSON, hosted receipts, and observer tests pin it). Copy it to a new adapter and generalize:

- Location: `native/` is for shared runtime C (`schway_foreign_*.c`); the Phase 23 adapter lives under `examples/`. Follow the Phase 23 layout: `examples/checksum/adapter.{h,c}` + `*.bindings.json` + expected JSON, so the "trusted foreign assertion / build authority" boundary stays visibly per-example (PROJECT.md Security constraint).
- Contract: `schway_file_bytes_acquire(const char *path)` -> `{int32_t status; {unsigned char *data; uint64_t length;} owner}`; `schway_file_bytes_release(owner)` -> `free(data)`. Keep `open(O_RDONLY|O_NONBLOCK)` + `fstat` + `S_ISREG` + `read` loop with `EINTR` retry + `close` error status exactly as Phase 23 does.
- **Size limit: 4096 bytes** (`SCHWAY_FILE_BYTES_MAX`), enforced by reading `limit+1` bytes (do not trust `st_size`; files can grow). Oversize -> `TooLong` status, buffer freed. Rationale for 4096 rather than 65536: loop iterations create dynamic occurrence events and the shared evidence/report ceiling is 65,536 (M005, and the N1 debt is exactly about the verify report exceeding it). Even at ~16 events/iteration, 4096 fits with headroom; raising the constant later is a one-line, one-test change. If phase research shows per-byte events are cheap, raising to 65,536 is allowed but not assumed.
- **Empty file is valid** (Adler-32 of empty = 1), so unlike Phase 23 it is not an error. `malloc(0)` is implementation-defined (may return NULL): always allocate `max(1, length)` and keep `length = 0`; a NULL return with `length==0` must not be treated as an alloc failure. Ownership/cleanup unchanged: acquire creates the obligation, release discharges, typed error path frees on every failure.
- **Typed errors** reuse Phase 24's `try`/typed-error cleanup: `data ReadError = | OpenFailed | NotRegular | ReadFailed | TooLong | AllocFailed | CloseFailed | InvalidPath`; stable stderr token per variant; error exit class follows the existing typed-error path (do not invent a new exit class; W2/W3 below are the only exit-class changes).
- **Interpreter**: implement the equivalent in Go (`os.Open` + `io.LimitReader(f, max+1)` + `Stat().Mode().IsRegular()`), written independently of the C so differential runs mean something. Go stdlib only.
- **`examples/checksum.schway`** (current file): provisional, uses invalid syntax (`loop buffer { next(buffer) ... }`, `add(...)`, two-arg `try`, `Buffer` path). It is a **refusal witness today**; replace it in place with the real program once admitted and update/retire whatever test pins it as refused. Its `checksum_print` foreign function is not needed: output uses `write_u64`/`write_newline`, so the program has **one** foreign symbol (acquire) plus release.
- Entry type: `main(path: PathToken) -> ...` as Phase 23 (not `Buffer path`). Public CLI shape for passing a path to `app run` must be confirmed in `cmd/schway/main.go`/`native_app.go` (today `app run` accepts bounded U64 input); this is a phase-research flag.

### Decision 5: Local modules -> **stdlib-only loader that flattens to one `ast.Program` before `check`**

| Concern | Recommendation |
|---------|----------------|
| Syntax | Add `import <module.path>` (one per line, after `module`, before `export`/decls). Reuse `modulePath()`. Add `TokenImport`. Only `fn` exports in M006 (no cross-module `data` types; the library needs only `Buffer`, `U64`). |
| Resolution | Module `a.b.c` -> file `<root>/a/b/c.schway`. `<root>` = entry file's directory with the entry's own module path suffix stripped (entry `examples.checksum` at `examples/checksum.schway` -> root = repo dir). Reject path segments containing `/`, `..`, or non-identifier characters (segments are already identifiers from the lexer), refuse symlinks that resolve outside root, and enforce a file-size cap and a module-count cap (suggest 16 modules, 64 KiB each) before reading. |
| Cycles | Refuse with a named diagnostic (`module.import_cycle`) listing the cycle. No transitive re-export. |
| Visibility | Only names in the target's `export { fn x }` are callable from importers; an unexported call is `module.not_exported` with the declaring module/span. Default private. |
| Merge | Loader parses each file, validates exports and visibility per module, then produces one `ast.Program` with each non-entry function renamed to an internal qualified name and calls rewritten. All six peers keep consuming a single program unchanged. This also matches `emitProgram`'s single-TU design and the roadmap's note that modules need not create separately verified binary contracts. |
| C symbols | Qualified internal name uses an injective encoding, e.g. `<len>_<seg>` per segment (`7_examples` `4_lib` `5_adler` `8_adler32`), so `a.b_c` and `a_b.c` cannot collide. `cName` then maps it into `SCHWAY_*` as today. `cName` uppercases, so add a linker-level uniqueness check on the resulting C name and refuse collisions (e.g. `Foo` vs `foo`) instead of silently merging. Do not loosen the `^SCHWAY_[A-Z0-9_]*$` invariant. |
| Diagnostics | `diagnostic.Span` carries offsets with no file identity. Add a source/file identifier (or per-module offset table) so errors in an imported file are attributed to the right file. This is a real cross-cutting cost; flag it. |
| Evidence/receipts | App build inputs and source hashes must cover **every** loaded module file, in sorted deterministic order, or `app verify` replay can pass against a stale library. |
| Dependencies | `os`, `path/filepath`, `io/fs`, `sort` only. No package manager, no lockfile, no versioning. |

**Second consumer.** Keep the library at `examples/lib/adler.schway` (`module examples.lib.adler`, `export { fn adler32 }`). Consumer 1:
`examples/checksum.schway` prints the full value. Consumer 2 (e.g. `examples/checksum_low16.schway`) imports the same module and prints
`adler32(...) % 65536`, so it is distinct behavior from the same unmodified library. A test must assert the library bytes are not duplicated (both consumers' flattened
programs reference the same source file hash, and neither file contains the loop).

**Rejected: separate compilation / per-module object files / header-based contracts.** That is D-13-02b's separate-compilation capability,
not required by the stated goal and would introduce user-declared ABI facts the producer cannot verify. Revisit only with a real B1 boundary witness.

## Core Technologies (unchanged, restated for roadmap)

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.24.0 (stdlib only) | Compiler, interpreter oracle, module loader | Existing host; `go.mod` has no deps and must stay that way. |
| Clang / C17 | installed Clang (Apple Clang 21.0.0 locally) | Native backend | Existing path; new C is hand-written adapter + emitted bounds checks only. |
| libc/POSIX | `open`, `fstat`, `read`, `close`, `malloc`, `free` | File adapter | Same calls as audited Phase 23 adapter. |

## Carry-over debt (no stack change; Go-only)

| Item | Stack implication |
|------|-------------------|
| N1 verify report >65536 loses capacity reason | Go change in app verify report assembly: truncate bounded detail, never the reason field. No new tech. |
| N2 two-host receipts at current SHA | Re-run existing CI matrix after final commit; no new tech. Module files must be in the declared input set (see Evidence/receipts). |
| W2 build origin vs corevalidate exit class | Split exit classes in `cmd/schway`; add distinct class constants and a CLI test per class. |
| W3 `run --engine=native foreign_origin_omitted` exits 3 vs interpreter 2 | Map to the interpreter's class; add a cross-engine exit-code parity test. |

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| Checksum | Adler-32 | CRC-32 / POSIX cksum | Needs xor/shift; multiple new operators across six peers. |
| Byte type | Widen to U64 | New `U8` scalar | Second arithmetic/overflow story; more peer work for no consumer. |
| Byte access syntax | `buffer_len(b)`, `buffer_byte(b, i)` | `b[i]`, `len(b)` | Lexer/formatter/parser surface larger; add sugar later over the same ops. |
| Bounds | Always dynamic check -> defect exit 65 | Static proof from loop guard | Extra unproven analysis; no consumer need. |
| File read | New copy-local generalized adapter | Edit Phase 23 adapter | Pinned by receipts/observer tests; copy-and-diverge is the stated project preference. |
| File read | C adapter | Schway-owned file handles | Roadmap: not demonstrated yet; needs close-error semantics. |
| Modules | Flatten to one `ast.Program` | Per-module IR + linker | Needs separate-compilation contracts (D-13-02b). |
| Dependencies | Go stdlib + libc | zlib / any hashing lib | Defeats the point (language must compute the checksum), adds a link dependency. |

## What NOT to Add

- No third-party Go modules; no zlib/libz link; no `.so`/`.a` packaging; no package manager/lockfile.
- No xor/shift/rotate/multiply operators, no hex output formatting, no U8 type, no `b[i]` sugar, no slices/sub-views, no mutation of buffers, no resource/loan carry across back edges beyond the single read-only shared borrow.
- No change to Phase 23/24/25/26/27 adapters, fixtures, or expected receipts.
- No cross-module `data` types or re-exports, no versioned or remote imports, no import-time code.

## Installation

```bash
# Nothing to install. Go 1.24 + installed Clang only.
go build ./... && go test ./internal/compiler/...
# External oracle (dev only):
python3 -c "import zlib,sys;print(zlib.adler32(open(sys.argv[1],'rb').read()))" FILE
```

## Phase research flags

1. **Read-only Buffer loan across a scalar back edge** (checker + two independent peers): highest risk; needs a design spike before planning.
2. **Public CLI path input for `app run`** (today bounded U64): confirm the exact surface and how the path reaches `PathToken`.
3. **Span file identity** for multi-file diagnostics and its effect on diagnostic JSON/protocol stability.
4. **Per-byte event volume vs evidence/report ceiling** to confirm 4096 is the right limit (and to fix N1 properly).

## Sources

- RFC 1950 (ZLIB Compressed Data Format) sec. 8.2 Adler-32: https://www.rfc-editor.org/rfc/rfc1950 (HIGH)
- Python `zlib.adler32` vectors, executed locally 2026-10-10 (HIGH)
- Go stdlib `hash/adler32` (test oracle only): https://pkg.go.dev/hash/adler32 (HIGH, not fetched this session; known stdlib package)
- Repo (read this session): `examples/checksum.schway`, `examples/phase23/adapter.{h,c}`, `examples/sum_to_n.schway`, `examples/fizzbuzz.schway`, `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go`, `internal/compiler/check/{check,scalar}.go`, `internal/compiler/cgen/cgen.go`, `internal/compiler/core/core.go`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, `.planning/PROJECT.md`
- Not independently researched this session: POSIX `cksum` spec details (CRC choice was rejected on operator grounds, not on spec reading); behavior of other languages' module systems (decisions rest on repo constraints, not ecosystem comparison). MEDIUM on those points.
