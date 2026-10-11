# Pitfalls Research

**Domain:** Bounded byte utility (byte views, bounds-checked indexing, loan across a scalar loop back edge, file-reading `checksum`) plus local modules with a second importing consumer, added to an existing Go-hosted compiler with C17/Clang native path, interpreter oracle, and independent validator peers.
**Milestone:** M006 Bounded Byte Utility and Local Modules
**Researched:** 2026-10-10
**Confidence:** MEDIUM-HIGH. Repo grounding (file paths, line numbers, prior debt) is HIGH: read directly from the tree. Checksum constants were recomputed locally with Python `zlib` (HIGH). C/UB, Rust borrowck, Go `os.Root`, and zlib CVE references come from training knowledge, not fetched in this session (MEDIUM); the planner should spot-check them against primary sources before citing them in requirements.

Phase names below are placeholders for the roadmapper. Suggested shape, in dependency order:
- **P-A Carry-over debt first:** N1 (report size), W2/W3 (exit classes).
- **P-B Byte view and indexing:** single-file, no loop carry.
- **P-C Loan across back edge:** relaxes `check.cfg_back_edge` for one narrow shape.
- **P-D `checksum` program and file-read spec.**
- **P-E Local modules and visibility.**
- **P-F Second consumer and closing evidence:** N2, hosted receipts last.

---

## Critical Pitfalls

### Pitfall 1: The "maximum input" requirement collides with the event-capacity and report-size ceilings (N1 reborn)

**What goes wrong:** The checksum loop runs once per byte. Today `app verify` on `sum_to_n` only verifies n <= 14 (event capacity, exit 69 `native.event_capacity_exhausted`) and loses even the named reason above n ~ 172 because exhausted cases embed ~371 bytes of interpreter evidence per unit of n and blow `MaxReplayCasesBytes` (65,536) (M005 audit N1, `session_app_verify.go`). A "maximum-size" checksum file (spec says empty, one-byte, maximum, oversized) will be 100s to 1000s of iterations. The headline case silently becomes `unsupported`, or exits 3 with "report exceeds 65536 bytes" and no report, while every other control is green.

**Why it happens:** M005 fixed N1's neighbor (the 14/15 boundary) and tested only there; RETROSPECTIVE lesson 1 ("run the public route on the headline input") was written for exactly this. The byte-count cap for the checksum program is a free parameter that nobody ties to the evidence ceilings.

**How to avoid:**
- Fix N1 in P-A before any byte work, with a test at n = 173, 1000, and 65536-equivalent: the report must either carry the named `native.event_capacity_exhausted` reason in compact form (no full interpreter evidence on exhausted cases) or exit with a named reason in the diagnostic.
- Choose the checksum size cap as a derived number: `cap <= largest input app verify can fully verify on all engines`, or explicitly declare the max-input case verified by a different, documented route (interpreter-vs-native `run` differential, not `app verify`) and say so in the requirement. Do not leave the maximum case as `unsupported` and call it covered.
- Record the cap as a single constant referenced by spec, cgen, interpreter, tests (N3 shows literals duplicated across layers drift).

**Warning signs:** Any acceptance row that says "maximum input" but whose evidence is `unsupported` or a seam test; `app verify` exit 3 with a size error; a cap chosen as a round number (65,536) rather than derived; test fixtures that stop at ~14 bytes.

**Phase to address:** P-A (report format), then P-D (cap derivation). Gate P-D acceptance on the max-input case reaching a pass verdict on the public route.

---

### Pitfall 2: Checksum definition is under-specified, and expected answers are derived from the implementation

**What goes wrong:** Classic Adler-32 mistakes: initial `a = 1, b = 0` (zlib `adler32(0, NULL, 0) == 1`; starting at 0 gives 0 for the empty file); `b += a` must use the already-updated `a`; modulus is 65521 (largest prime below 2^16), not 65536/65535 and not a transposed literal such as 65251; final value is `(b << 16) | a`, and swapping halves passes any test on inputs where `a == b`. Deferred mod (NMAX = 5552 trick) is only needed for 32-bit accumulators; with checked U64 and a small cap, per-byte `% 65521` is simpler and obviously equals the reference. Deferral bugs appear exactly at the NMAX boundary. Separately, if expected values come from running the program under test, a wrong algorithm is "confirmed" by itself.

**Why it happens:** The existing `examples/checksum.schway` is a plain additive sum (`total = add(total, byte)`), and its syntax is declared provisional. It is easy to ship "a checksum" that is a byte sum and not any named, interoperable one. The compiler's own differential (interp vs `-O0/-O3/-flto`) proves engines agree with each other, not that the algorithm is Adler-32.

**How to avoid:**
- Name the checksum in the requirement (recommend Adler-32, RFC 1950 section 8.2: no tables, no bitwise ops, only add and remainder, which M005 already admits) and state the output format byte-exactly: decimal vs hex, if hex then fixed 8 digits lowercase with leading zeros and trailing newline, and that it prints the arithmetic value (no endianness reinterpretation, no `memcpy` of the integer).
- Pin expected answers as literals generated once, outside Schway, and commit them with the generating command and tool version in a comment: Python `zlib.adler32` (and optionally Go `hash/adler32` from the standard library as a second peer, no new dependency). Values verified locally on 2026-10-10:

  | Input | Adler-32 (hex) | Why it is in the set |
  |---|---|---|
  | empty (0 bytes) | `0x00000001` | init value; leading zeros |
  | `a` | `0x00620062` | one byte |
  | `abc` | `0x024d0127` | RFC/Wikipedia-style |
  | `Wikipedia` | `0x11e60398` | widely published value, independent cross-check |
  | `\xff` x1 | `0x01000100` | high byte, sign extension |
  | `\x80` x10 | `0x1b8a0501` | high bytes accumulate (signed-char bug flips this) |
  | `\xff` x5552 | `0xf18f9b8c` | NMAX boundary |
  | `\xff` x5553 | `0x8e299c8b` | NMAX + 1 |

  Add the cap-sized and cap+1 files (oversized must produce the specified error, not a checksum).
- If the "no checked-overflow can occur" consequence of per-byte mod is true, say so: the overflow control for this program cannot be reached by accepted inputs (the F5 situation in M005 audit). Do not invent a synthetic overflow claim; keep the overflow witness on `sum_to_n`.
- If the output needs `(b << 16) | a`, note that multiplication or shifts are not admitted yet (only checked add, `<`, `==`, `%`). Either print `a` and `b` as the specified pair, compute `b * 65536` by an admitted construct, or admit a checked multiply as its own gated change. A new `OperationKind` lands at six dispatch sites (M002 lesson); do not sneak it into the checksum phase.
- If CRC-32 is chosen instead: reflected polynomial `0xEDB88320`, init `0xFFFFFFFF`, final xor `0xFFFFFFFF`, check value for `123456789` is `0xCBF43926`, and it needs a 256-entry table or bitwise ops that Schway does not have. Recommend against for M006.

**Warning signs:** Expected files produced by `--update`/golden-regenerate flags; a test that computes the expected value by calling the same Schway function; fixtures lacking bytes >= 0x80, lacking NUL bytes, or lacking any input near 5552; empty-file expectation of 0.

**Phase to address:** P-D (spec and fixtures first, as refused-frontier fixtures with pinned literals per the constructibility gate).

---

### Pitfall 3: `char` signedness and byte handling in emitted C hide behind two signed-char hosts

**What goes wrong:** If the adapter or emitted C reads bytes through `char *`/`const char *`, bytes >= 0x80 sign-extend on x86-64 and on Apple arm64 (signed char), but not on Linux aarch64 (unsigned char). The project's dual-host receipts are macOS arm64 and Ubuntu (almost certainly x86-64), both signed, so a sign-extension bug can pass both hosts' CI. Related: `printf("%x", byte)` with negative value, `strlen`/`fopen` text mode treating NUL as terminator, reading file content through a `FILE*` text-mode abstraction.

**Why it happens:** Existing adapter (`examples/phase23/adapter.c`) already uses `unsigned char *data`, but new byte-view code, glue emitted in `internal/compiler/cgen/cgen_program.go` (the phase24 glue uses fixed-shape strings), and the checksum print helper are new code written under time pressure.

**How to avoid:**
- All byte storage and loads use `uint8_t`/`unsigned char`; promote to `uint64_t` explicitly at the load site (`(uint64_t)data[i]`).
- Make fixtures kill the bug on any host: include 0x80, 0xFF, and 0x00 bytes (table above), and add a mutation control that recompiles the generated C with the byte type replaced by `char` under `-fsigned-char` and asserts the checksum fixture goes red. That proves the fixture detects the failure class independent of host default signedness.
- Consider adding `-fsanitize=implicit-integer-sign-change` (Clang) to the existing ASan/UBSan lane for this subject; treat as optional.

**Warning signs:** Any `char` in new emitted C for byte data; only ASCII test files; mutation list lacking a sign-extension entry.

**Phase to address:** P-B (byte view representation) and P-D (fixtures and mutation).

---

### Pitfall 4: Bounds check is correct in the interpreter but unsound or dead in native C

**What goes wrong:**
- Check performed after the load or after forming a pointer: `v = data[i]; if (i >= len)` is UB before the check; the optimizer may delete the check at `-O3`/`-flto`, producing interpreter != native only in optimized tiers.
- Pointer-form checks (`data + i < end`) are UB when `data + i` goes beyond one-past-the-end (C17 6.5.6p8), and wrap with large `i`.
- `NULL + 0` is UB in C17 (UBSan pointer-overflow flags it): an empty file whose buffer is `malloc(0)` may give `NULL`, and the existing generated glue treats `owner.data == NULL` as acquire failure (`cgen_program.go` ~line 1188; adapter returns `AcquireFailed`/`AllocFailed`). An empty file would be misreported as an allocation error, or a NULL-with-length-0 view would be offset.
- Width: interpreter `U64` vs Go `int` (slice index) vs C `size_t`. Comparing `i < len` after casting `i` to `int` makes huge indexes negative or wrapped. On 32-bit `size_t`, `(size_t)u64_len` truncates.
- `assert`-based checks vanish under `NDEBUG`.
- Go interpreter relies on slice panic for OOB (exit 2, Go runtime text), so OOB "agrees" by accident of a crash, not by a named outcome.

**How to avoid:**
- Emit exactly: `if (index >= view.length) { named_outcome }` on `uint64_t` operands, then `view.data[index]`. Never form an out-of-range pointer. No `assert`.
- Allocate `max(1, n)` bytes (or a static one-byte sentinel) so `data` is never NULL for a valid empty view; the length, not the pointer, carries emptiness. Remove the "NULL means failure" test for the non-failure empty case, replacing with an explicit status.
- Add `_Static_assert(sizeof(size_t) >= sizeof(uint64_t) || cap <= SIZE_MAX, ...)` or compare against `SIZE_MAX` before any cast.
- Interpreter: compare as `uint64` before converting to `int`; implement OOB as a named outcome with a dedicated exit code and exact stderr (the exit-69 `native.event_capacity_exhausted` pattern), identical across interpreter and native. Do not let a Go panic or a signal be the OOB result.
- Do NOT add bound-check elision based on the loop guard. Clang already does it soundly; a hand-written range analysis is an unproven peer-visible claim.
- Mutation controls (each must turn a specific control red, and be run with the ASan lane): (a) delete the check; heap buffer of exact size makes ASan report overflow; (b) change `>=` to `>` (index == len off-by-one); (c) truncate index to 32 bits; (d) load-before-check ordering; (e) NULL data on empty view.
- Boundary fixtures: empty file index 0, one-byte index 0 and 1, index = len, index = len - 1, index = U64 max, index = len + 1.

**Warning signs:** OOB test passes only at `-O0`; ASan lane shows no byte-view subject; empty-file case exits as `AllocFailed`; interpreter OOB message contains Go runtime text or exit 2 equals "invalid program".

**Phase to address:** P-B.

---

### Pitfall 5: File-read semantics are implicit; the race, special-file, and zero-length cases decide the "exact error results"

**What goes wrong:** Reading by `stat` size then `read` (TOCTOU: file grows or shrinks between), trusting `st_size` for `/proc`-style files (size 0 but content), reading `/dev/zero` (infinite), opening a FIFO blocking, treating a directory (open succeeds, read fails EISDIR) as empty, ignoring short reads/`EINTR`, ignoring `close` failure, and mapping errors by OS text. Existing adapter (`examples/phase23/adapter.c`) is a sound base: `O_RDONLY|O_NONBLOCK`, `fstat` on the descriptor then `S_ISREG`, read loop with `EINTR`, one extra-byte probe for `TooLong`. But it is hard-wired to exactly one byte, and returns `EmptyFile` for zero length, which the checksum spec must treat as valid (checksum 1).

**How to avoid:**
- Generalize the adapter to `cap`: read in a loop up to `cap + 1` bytes; if more than `cap` arrive, `FileTooLong` (this handles growth, `/dev/zero` behind a regular check, and wrong `st_size`). Never size the buffer from `st_size` alone. Document the snapshot semantics: bytes up to first EOF or cap+1, file may change underneath.
- Keep `fstat` on the opened descriptor (not path `stat`) and `O_NONBLOCK` open so FIFOs do not hang; keep `S_ISREG` refusal (directory, FIFO, device, socket).
- Symlinks: `open` follows them. State that explicitly (this is runtime user authority, distinct from import-time build authority in Pitfall 11).
- Specify one outcome class per case and test by class, not OS message: success, `EmptyFile` is success with checksum 1, `FileTooLong`, `OpenFailed` (missing, permission), `NotRegular` (directory/FIFO/device), `ReadFailed`, `CloseFailed`, `InvalidPath`. Each also frees the buffer exactly once (error-path cleanup is a stated requirement; reuse the owner discharge evidence from M004 Phase 23/24).
- Decide the interpreter side now: the Go interpreter must read with the same algorithm (limit+1, regular check via `Stat` on opened file, class mapping), not `os.ReadFile`. Otherwise the differential compares two different file semantics.
- CI hazards: an unreadable-file control (`chmod 000`) is vacuous when tests run as root (common in containers; root reads it). Detect `euid == 0` and use another unreadable construction or report a named skip, never a green pass. FIFO creation (`mkfifo`) and `/dev/zero` exist on both hosts but the test must not hang; wrap with deadlines (project already uses finite subprocess deadlines).
- Fixture creation must happen in the test (temp dir), not as committed special files.

**Warning signs:** Reported "TooLong" only for files exactly cap+1 created statically; no growing-file or `/dev/zero` control; a skip counted as pass; interpreter uses `os.ReadFile`; Go test expects specific `strerror` text.

**Phase to address:** P-D.

---

### Pitfall 6: Borrow across a loop back edge: a spec misreading shared by every peer (common-mode failure)

**What goes wrong:** The current refusal lives at `check.cfg_back_edge` (`internal/compiler/check/check.go` ~2797-2830; `internal/compiler/check/scalar.go` ~96-106, 237) and is mirrored by the independent peers (`corevalidate`, `originvalidate`, `pathoracle`, `executionpeer`). Relaxing it for "a shared loan of an owned buffer live across the back edge" creates a new admission rule that is easy to get wrong in the same direction in all peers, because they are all written from one prose spec. Rust shows how: NLL "problem case #4"-style and long-standing soundness holes (for example implied-bounds issue rust-lang/rust#25860) survived because the model, not the implementation, was wrong. Unsound cases to cover:
- buffer **released or moved** inside the body (`release(buffer)` or transfer to a callee) while a view from the previous iteration is still used next iteration;
- **exclusive** loan or mutation of the same buffer inside the loop while a shared loan is carried (including through an exclusive pointer alias);
- **rebinding/shadowing** of the borrowed name in the body (the provisional `checksum.schway` already shadows `let total`) so the loan refers to a stale binding;
- a loan created in the body and still live at the loop head (accumulating loans), or escaping through the loop result or through `try` early exit;
- an early exit (`try` failure) inside the loop that must release the owner exactly once, vs. loop-exit path that releases after the loop.

**Why it happens:** The same author writes spec and all peers in a short window; "structural opposition" (backward memoized walk vs forward propagation) protects against implementation mistakes, not against a shared wrong rule. The provisional syntax (`loop buffer { next(buffer) }`) also treats the buffer as a consumed iterator (exclusive), which is not the shared-borrow-plus-index model the milestone describes.

**How to avoid:**
- Write the admission rule as a small closed predicate before coding ("loop-invariant shared loan of an owner that is never moved, released, rebound, or exclusively borrowed in any block reachable from the loop head; no loan created in the body outlives an iteration"), and have each peer derive it with its existing opposite algorithm.
- Fixture-first: check in each unsound shape above as a refused fixture with its pinned diagnostic **before** relaxing anything (constructibility gate), then show only the sound shape moves from refused to accepted. The "moved frontier" test should also assert every unsound neighbor is still refused with the same code.
- Mutation-kill: seed a fault in each peer separately (each should be detected by at least one other peer or by the oracle), and a coordinated fault in all (the declared `escape:coordinated-source-to-core-false-claim` style) must be shown to be caught by something that does not share the rule: the interpreter oracle plus the ASan lane running an engineered unsound C emission (release inside loop, then use) proving use-after-free is observable. If the engineered unsound program cannot be emitted because the checker refuses it, hand-write the C (as NAT-07 did) and say so.
- Keep interpreter loan-awareness dynamic: the interpreter should treat use-after-release as a named failure, giving the oracle an independent say.
- Keep the resource carry refusal for every other shape (owner transfer across back edge, owner acquired in body, loans of values other than the single buffer). Do not widen "scalar loops" to "loops with resources".

**Warning signs:** All peers changed in one commit with identical test names; refused-neighbor fixtures missing; the only new negative control is "OOB index"; the relaxation is expressed as removing a refusal branch rather than adding a narrow positive predicate.

**Phase to address:** P-C (after P-B proves indexing alone, before P-D). Needs deeper phase research (design of the predicate and peer derivations).

---

### Pitfall 7: Local modules bundled with new byte machinery in one milestone

**What goes wrong:** `STANDING-VERDICTS.md` lists as a process anti-pattern: "Don't bundle modules/separate compilation into a milestone already extending five-plus-consumer machinery" (M001 Phase 3 cost 70-95 min/plan versus ~10 for single-consumer work). M006 does both. Every change to name resolution lands at all six dispatch sites (`check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, `originvalidate`), plus cache keys, evidence manifests, diagnostics, formatter, and repair.

**How to avoid:**
- Sequence so a runnable gain lands before modules: single-file `checksum` working end to end (P-D) before P-E starts. A second consecutive enabling phase without a runnable gain triggers scope review (AGENTS.md rule).
- Define modules narrowly as **whole-program textual composition resolved before checking** (a deterministic, sorted, closed set of source files is resolved into one program with module-qualified names), not separate compilation. State that DX-06/B1 blame and separate-compilation contracts are not claimed.
- Add an explicit scope-review checkpoint if P-E exceeds a plan-count bound agreed in the roadmap.

**Warning signs:** Module work starting before checksum has dual-host pins; plan counts growing per phase (RETROSPECTIVE: 3, 7, 10, 13 drift); "separate compilation" language creeping into requirements.

**Phase to address:** Roadmap structure (this document's phase order).

---

### Pitfall 8: Import resolution escapes the project root (build authority)

**What goes wrong:** Import `a.b` or `"../x"` mapped to a path with `filepath.Join(root, ...)`: `..` segments, absolute paths, symlinks inside the project pointing outside (`lib -> /etc`), case-insensitive collisions on macOS vs Linux (`Checksum.schway` vs `checksum.schway`), Windows reserved names, and a module file that is a FIFO or `/dev/zero` (hang/unbounded read at build time). The project constraint says build authority stays explicit; Phase 22 introduced a closed local-input manifest (`internal/compiler/native/bindings.go`: size-limited `readBindingFile`, declared inputs, content-bound identity) which module reads should reuse.

**How to avoid:**
- Module names are validated identifiers only (segments of `[a-z][a-z0-9_]*`, no empty segments, no `.`/`..`, no separators) and mapped to a fixed layout by construction; reject before touching the filesystem.
- Resolve and read through Go 1.24's standard-library `os.Root` (`os.OpenRoot(projectRoot)`; `root.Open`, `root.Stat`), which rejects traversal and symlinks that escape the root. No new dependency. If a hand-rolled check is preferred, it must `EvalSymlinks` then verify containment, and still use `Lstat`/regular-file check on the opened file, to avoid check-then-use races.
- Regular-file only, size-limited (reuse the `limit+1` read pattern), file count capped, depth capped.
- Fixtures: `../escape`, absolute path, symlink to outside directory, symlink to outside file, FIFO as module file, oversized module, case-variant duplicate, module name containing NUL. Each refused with a named code and the offending path in the diagnostic (relative, not absolute: avoid leaking host paths into evidence).

**Warning signs:** `filepath.Join` on user-controlled strings with no containment check; tests that only test `..`; absolute host paths in diagnostics or receipts (the project already ran a privacy audit for personal paths).

**Phase to address:** P-E (before any resolution code merges).

---

### Pitfall 9: Module graph semantics: cycles, diamonds, shadowing, visibility leaks, ambiguity

**What goes wrong:**
- **Cycles:** the existing `callgraph` three-color DFS refuses call cycles, but module import cycles are a different graph. A cyclic import with an acyclic call graph slips through, or the new import DFS conflates "visited" with "on stack" and refuses diamonds as cycles.
- **Diamond:** A imports B and C, both import D: D must be loaded once (canonical identity = normalized module name + content digest), emitted once, and its functions not duplicated or double-registered.
- **Shadowing/ambiguity:** two imported modules export the same name; local `fn` or `let` vs imported name; first-wins resolution is silent. Refuse ambiguity with both origins named.
- **Visibility leak:** `export { fn main }` exists in the syntax today (`examples/checksum.schway`). Enforce it in every layer, not just `check`. Leaks to test: consumer calls a non-exported function by qualified name; exported function's signature mentions a non-exported `data` type; re-exports; an exported function whose body calls a private one (fine) but whose published interface (`lang.interface/1`) includes private details; peer validators accepting a call to a private name because they only check arity/type (fixing one of N peers is not fixing the item: M001 lesson 3).
- **Callable subset publishable:** the existing rule is that a caller admission reads only callee signatures, never bodies (D-04-03). Whole-program composition must not let `check` read an imported body to make a consumer decision.

**How to avoid:**
- Specify: acyclic import graph (named refusal `module.import_cycle` listing the path), diamond allowed, no re-export in M006, no wildcard imports, qualified references only, ambiguity refused, entry module is the only one whose `main` is the program entry.
- Controls: cycle of length 2 and 3, self-import, diamond (accepted, D emitted once), duplicate export names, private call, private type in public signature, consumer-body-independence test (change imported body without changing signature: consumer's check/admission output byte-identical; change signature: consumer result changes).
- Add each rule to every peer that sees names, with a mutation per peer.

**Warning signs:** Visibility only tested via `check` diagnostics; diamond test missing; cycle test uses direct two-module case only.

**Phase to address:** P-E.

---

### Pitfall 10: Symbol collisions and non-determinism in emitted C

**What goes wrong:** Mangling `module.fn` to `module_fn` collides: module `a` fn `b_c` vs module `a.b` fn `c`. Imported `main` collides with entry `main`. Names collide with runtime prefixes (`schway_`, `SCHWAY_`) or libc. Foreign C symbols are global (`foreign C { fn checksum_read ... }` in the provisional example): two modules declaring the same symbol with different contracts, or one library declaring the foreign contract the consumer also declares, produce ODR-style mismatches. Emission order driven by Go map iteration changes bytes between runs, breaking byte-identity goldens and cache keys.

**How to avoid:**
- Injective, collision-proof mangling (length-prefixed segments, for example `schway_m<len><seg>_f<len><name>`), with a test that enumerates adversarial names; keep the entry `main` mapping unchanged so existing programs keep their bytes (use the pre-fix-revision byte-identity golden technique from M005 27.2/27.3 for every non-module program).
- Foreign symbols: dedupe by symbol only when the full contract is identical; otherwise refuse with both locations named.
- Sort everything: module load order (lexicographic by canonical name), exports, emitted functions. Test by running the build twice and under shuffled directory-read order (inject an order-reversing fs layer in a test).
- Fixture: library function and consumer function sharing local name, nested module name overlap, imported `main`.

**Warning signs:** `for k, v := range` over maps in resolution or emission; byte-identity goldens regenerated rather than compared; unchanged programs' C diff.

**Phase to address:** P-E.

---

### Pitfall 11: Stale cache/evidence when an imported module changes; diagnostics blame the wrong file

**What goes wrong:**
- Cache keys, evidence manifests, and app receipts keyed on the entry file alone serve stale verdicts when only the library changes (the precedent already noted in `STANDING-VERDICTS.md`: "per-unit hashes will silently serve stale verdicts when an unchanged caller's callee changes"). The existing cache is keyed on the call-graph closure; module closure must be added. Phase 22 receipts already say `dependency_closure: incomplete`, `cacheable: false`; resist flipping that to complete without evidence.
- Spans carry byte offsets assumed to belong to one file: an error in the library is reported against the consumer's file at the same offset; `schway format` on a module; `schway-repair` applies a patch to the wrong file; `explain` cause DAG names a function but not its module.

**How to avoid:**
- Key = digest over the sorted list of (canonical module name, content digest) for the whole import closure plus compiler identity. Controls: edit library only (cache miss, consumer verdict re-derived), edit an unrelated sibling file (hit), swap two modules' contents (miss).
- Diagnostics: add file/module identity to span as an additive schema field with frozen prior bytes (the `/0 -> /1` precedent); existing single-file diagnostics stay byte-identical. Control: a type error in the library reports the library's path and the right line; a repair JSON patch carries the file and is rejected if the file digest differs.
- Receipts include the module closure in declared inputs.

**Warning signs:** Cache test with a single source; span struct without a file field; golden diagnostics unchanged after adding modules (suspicious: nothing was actually attributed).

**Phase to address:** P-E (design) with P-F verification.

---

### Pitfall 12: "Imports without copying" is claimed but not proven; second consumer is the same program

**What goes wrong:** The second consumer is a thin copy of `checksum`, or the library is copy-pasted and both compile; passing tests do not show sharing. Or the second consumer exercises only the trivial path.

**How to avoid:**
- Second consumer must be a different program (different `main`, different output contract, for example verifying a stored expected Adler-32 against a file) that calls the same exported function.
- Sharing proof as an executable control: grep-style static test that the library's loop body text appears in exactly one source file; mutation of the library (change modulus literal 65521 to 65251) flips both consumers' pinned outputs; both consumers' evidence records name the same library digest.
- Consumers' pinned answers are independent literals (Pitfall 2), not each other's output.

**Warning signs:** Two `.schway` files containing the same loop; a mutation that only affects one consumer.

**Phase to address:** P-F.

---

### Pitfall 13: Exit-class carry-over (W2/W3) changes ripple into pinned goldens, guides, and new byte outcomes

**What goes wrong:** Splitting `schway build` (origin refusal exit 2 vs corevalidate refusal exit 3 `native.build_failed`) and `run --engine=native` (exit 3 `native.tool_failure` vs interpreter exit 2 `core.foreign_origin_omitted`) changes exit codes observed by pinned tests, guide scripts (`examples/phase27`), and hosted gates. New checksum outcomes (OOB, file classes) also need exit codes; scattering literals across `cgen.go`, native, session repeats N3.

**How to avoid:**
- Do W2/W3 in P-A, before byte outcomes, and establish one exit-class table (single source consumed by cgen, native runner, session, CLI, tests) so P-B/P-D add rows rather than literals.
- Pin before/after: list every existing test/script that asserts the old class; change them in the same plan; add a control that a corevalidate refusal is `invalid` (exit 2) with a diagnostic, and that real toolchain failures stay exit 3.
- Do not renumber existing codes (65 overflow, 69 event capacity); allocate new codes from a documented range and assert uniqueness in a test.

**Warning signs:** Exit-code integers appearing in new Go or C literals; "operational_failure" used for an input-caused refusal.

**Phase to address:** P-A.

---

### Pitfall 14: Hosted receipts at a stale revision (N2) and late review fixes

**What goes wrong:** Hosted dual-host run 38091738463 is at d7ffbcd; later review fixes have local-only evidence. M006 can repeat it: receipts collected mid-milestone, then review/validation fixes land afterwards. Receipts are also "distinct from local or container runs" per PROJECT.md.

**How to avoid:**
- Make receipt acquisition the **last** plan of the final phase, after code review and Nyquist validation fixes, with a gate that compares the receipt revision to the tree being claimed (restrict to code-relevant paths so docs-only commits do not force a rerun; D-27.1-06 rejected ancestor acceptance for now, so use exact-revision or an explicit path-scoped rule written down).
- Any post-receipt code change either re-dispatches or is recorded as local-only in the phase artifact and the milestone audit, never silently.
- Run session meta-tests (inventory, groundedness, validation lifecycle) locally before dispatch (memory: they catch planning-doc and call-site drift).
- If no remote/runner access exists, treat as external blocker per PROJECT.md (no gap plan, no UAT).

**Warning signs:** Receipt SHA differs from HEAD for non-doc files; verification says "passed" with "hosted evidence pending".

**Phase to address:** P-F.

---

### Pitfall 15: The instrument reports green because it is wired, not run (new surfaces)

**What goes wrong:** The systemic M003 finding applies to every new subject: new differential rows with zero cases, `go test -run` patterns that match zero tests and exit 0, mutation rows skipped by a build tag, fixtures never reached by the generator, validators that do not implement a new operation kind but fail open. New concrete risks: a new byte-op `OperationKind` (if any) missing from one of the six dispatch sites or two exhaustive-dispatch controls; module fixtures that pass because loading silently fell back to the single-file path; the unreadable-file control passing under root.

**How to avoid:**
- Per-requirement: a negative control that has been seen to fail (mutation-kill executed in the plan that introduces it), recorded with the mutated diff and failing test name.
- Verification commands in plans must be runnable verbatim (no elided `...`), and each `-run` pattern must be shown to match at least one test (use `-v` count assertion or the existing groundedness lint).
- Grade evidence per the existing vocabulary (`DEFINED | WIRED | REACHABLE | EXERCISED | MUTATION-KILLED`) and require EXERCISED+ for each M006 requirement; unreachable claims go to `.planning/UNREACHABLE-CLAIMS.md` with a trigger (for example per-byte-mod overflow).
- Verify on the public route with the headline input: `schway app verify` on checksum with the cap-sized file; guide commands run in a clean clone (W4 style).

**Warning signs:** "wired" in a verification row; a table cell with `-run` that was never executed; mutation count equals zero for new rows.

**Phase to address:** Every phase; enforced in plan verification sections.

---

### Pitfall 16: Debt phases grow and audits reopen the milestone

**What goes wrong:** M005 closed three times; audit-driven decimal phases (27.1-27.3) doubled planning churn. M006 carries N1-N4, W2/W3, R1b, and nine D-27.1-10 dispositions; adding checksum plus modules plus loans invites the same.

**How to avoid:**
- Take only N1, N2, W2/W3 (as scoped) and refuse N3/N4/R1b/D-27.1-10 unless a new M006 phase touches them (N3 duplicated literals will be touched by new exit codes: fold it into P-A's exit-class table deliberately, as a scoped decision; N4 guide note is trivially touched when N1 changes the guide).
- Pre-close audit run **before** the last phase: public-route checks (app verify on headline input; guide in clean clone; hosted receipt SHA) so an audit cannot find a user-facing defect after closure.
- Plan-count watch: record per-phase plan counts; growth without a named user witness or safety obligation triggers scope review (AGENTS.md).

**Warning signs:** Third decimal phase; verification of "public routes" first appears in the audit rather than in the phase.

**Phase to address:** Roadmap/closure (P-F), with pre-audit at the start of P-F.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|----------------|-----------------|
| Regenerate golden/pinned values from the new output | Tests go green fast | Algorithm errors become "expected"; kills independence | Never for checksum values; only for byte-identity goldens regenerated from the pre-change revision |
| Reuse `os.ReadFile` in the interpreter for file bytes | Less code | Interpreter and native read different semantics (symlinks, special files, growth) | Never; implement the same limit+1 algorithm |
| Deferred-mod (NMAX) optimization | Looks like zlib | Adds boundary bugs; no benefit at the cap | Never in M006 (per-byte `%`) |
| Hand-roll path containment | No `os.Root` learning | Check-then-use symlink races | Only if `os.Root` is unusable and EvalSymlinks+Lstat is covered by escape fixtures |
| Literal exit codes in new C/Go | Quick | N3 repeat; drift | Never; extend the single exit-class table |
| Skipping a control on CI (root, no mkfifo) | Green build | Vacuous coverage | Only as a named `skip` status counted separately from pass |
| Mangling with `_` join | Readable C | Collisions | Never; use length-prefixed segments |
| Module cache keyed on entry file only | Simple | Stale verdicts | Never |

## Integration Gotchas

| Integration | Common Mistake | Correct Approach |
|-------------|----------------|------------------|
| Phase 23 file-byte adapter | Reuse as is: 1-byte exact, `EmptyFile` error, NULL-data means failure | Generalize to cap; empty is success; non-NULL data for empty view |
| Phase 24 glue in `cgen_program.go` (fixed-shape string emission, hard-coded `schway_file_byte_*` symbol matchers ~lines 1164-1195) | Add a second special-case shape beside it | Route through the general `emitProgram` law; keep byte-identity for existing programs |
| `app verify` / replay (`session_app_verify.go`) | Assume per-case evidence size is bounded | Fix N1 first; test at n=173+, cap-sized |
| ASan/UBSan lane | Assume it covers new subjects automatically | Add explicit byte-view and unsound-loop subjects; confirm ASan actually reports on the mutant |
| Event identity per loop iteration (`lang.execution/2`) | Per-byte events blow capacity | Choose cap vs. capacity (Pitfall 1) or aggregate evidence |
| Evidence/receipt closure | Add modules but keep `dependency_closure: incomplete` ambiguous | Include module digests in declared inputs; state closure status explicitly |
| `schway-repair` JSON protocol | Patches without file identity | Carry file/module and digest |
| Formatter | Canonicalize only entry file | Format every module; sort import/export lists |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|----------------|
| Per-byte event records in generated C | Event capacity exit 69; huge evidence | Cap derivation, compact exhausted-case evidence | ~15 iterations for `app verify` today; ~172 for report size |
| Fixed-point budget (65,536) in checker/validators | Loop admission refuses or times out on bigger bodies | Keep loop bodies tiny; count budget differences are F4 debt (all fail-closed) | Large loop bodies, not input size |
| Module resolution reads many files | Slow cold checks | File-count cap, content-hash memo; measure with existing `measure` p50/p95 | Dozens of modules (out of scope) |
| Interpreter step limit on cap-sized loop | Interpreter exhausts a step/stack bound before native finishes | Check interpreter loop budget against cap in P-D | cap x per-iteration steps |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Import path escapes root (`..`, absolute, symlink) | Build reads arbitrary host files into evidence/artifacts (HIGH) | Name validation plus `os.Root`; escape fixtures |
| Reading special files at build or run time (FIFO, `/dev/zero`) | Hang or unbounded memory (HIGH for build, MEDIUM for run) | `O_NONBLOCK`, `S_ISREG`, limit+1 reads, deadlines |
| Allocation size from untrusted length without cap check | Overflow/truncation (cf. zlib CVE-2022-37434 class: length-driven copy) (MEDIUM, from memory) | Compare against cap before any cast or malloc |
| Host absolute paths in diagnostics/receipts | Privacy leak (project already audited) | Project-relative paths only |
| Dual-module foreign symbol with mismatched contract | Memory-unsafe call through wrong ABI | Dedupe on full contract or refuse |
| Reporting an unreadable-file test green when running as root | False assurance | Detect euid, named skip |

## UX Pitfalls

| Pitfall | User Impact | Better Approach |
|---------|-------------|-----------------|
| Error outcomes distinguished only by exit code | Cannot tell empty vs unreadable vs oversized | Exact stderr line per class plus distinct exit class, documented in the guide |
| Exhausted `app verify` shows no reason | Developer thinks tool is broken (N1) | Always carry the named reason |
| Diagnostics in imported module cite consumer file | Misleading fix location | File identity in spans |
| Checksum prints decimal while tools print hex | Cross-checking against `python -c zlib` confusing | Specify and document format; show the matching one-liner in the guide |
| Guide mentions only n <= 14 (N4) | Stale instructions | Update guide when N1 changes |

## "Looks Done But Isn't" Checklist

- [ ] **Checksum:** Often missing independent pins and boundaries; verify literals came from `zlib` (command recorded), include empty, 0x80/0xFF, 5552/5553, cap, cap+1.
- [ ] **Bounds check:** Often missing the ASan-observed mutant; verify deleting the check turns a control red under ASan and `-O3`.
- [ ] **Empty file:** Often missing non-NULL data and success path; verify checksum 1 and no `AllocFailed`.
- [ ] **File errors:** Often missing per-class, per-cleanup evidence; verify each class frees the owner once and exits with its class on both engines.
- [ ] **Loan across back edge:** Often missing refused neighbors; verify release/move/exclusive/rebind/escape fixtures still refuse with pinned codes.
- [ ] **Peers:** Often missing the second/third peer; verify each of `check`, `corevalidate`, `originvalidate`, `pathoracle` has its own seeded fault for the new rule.
- [ ] **Modules:** Often missing visibility in non-check layers; verify private-call fixture refused by every layer that resolves names.
- [ ] **Import safety:** Often missing symlink escape; verify symlink-to-outside refused.
- [ ] **Determinism:** Often missing shuffled-order test; verify identical C bytes under reversed directory order.
- [ ] **Cache:** Often missing library-only edit; verify miss on library change, hit on unrelated file.
- [ ] **Second consumer:** Often missing sharing proof; verify single copy of the loop plus mutation flips both consumers.
- [ ] **Public route:** Often missing headline input; verify `app verify` on cap-sized file and clean-clone guide.
- [ ] **Hosted receipts:** Often at an older SHA; verify receipt revision equals the claimed code revision.
- [ ] **No `-run` zero-match:** verify every verification command matches at least one test.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|---------------|----------------|
| Wrong checksum algorithm shipped | MEDIUM | Re-pin literals from zlib, fix program, regenerate native/interp goldens from pre-fix revision byte-identity for other programs |
| Unsound back-edge admission discovered | HIGH | Revert to refusal at `check.cfg_back_edge`, keep new refused fixtures, re-derive predicate with a different peer author or the oracle; mark requirement partial rather than weaken |
| Import escape found | HIGH | Disable import resolution behind a named refusal, add escape fixtures, switch to `os.Root`, audit receipts for leaked paths |
| Stale cache served | MEDIUM | Widen key to module closure, add library-only edit control, invalidate stored entries by schema bump |
| Mangling collision | MEDIUM | Switch to injective mangling; goldens for non-module programs must remain byte-identical |
| Receipts at stale SHA | LOW | Re-dispatch hosted jobs at final tree; or record local-only explicitly |
| Debt phases ballooning | MEDIUM | Scope review; defer N3/N4/R1b with explicit owners |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|------------------|--------------|
| 1 Max-input vs capacity/report size | P-A, P-D | `app verify` on cap-sized file returns pass, or compact named reason at n=173/1000 |
| 2 Checksum spec and independent pins | P-D (fixtures first) | Literals from zlib with command; boundary table all present; wrong-modulus mutant red |
| 3 Signedness | P-B, P-D | `-fsigned-char` mutant of generated C fails the 0x80/0xFF fixtures |
| 4 Bounds/UB in C | P-B | Five mutants red under `-O0/-O3/-flto` and ASan; empty-buffer non-NULL |
| 5 File-read semantics | P-D | Per-class outcome and cleanup tests; growth, `/dev/zero`, FIFO, dir, symlink, root-skip named |
| 6 Back-edge common-mode | P-C | Refused-neighbor fixtures, per-peer seeded faults, ASan-observed unsound mutant |
| 7 Bundling modules with bytes | Roadmap order | Checksum has dual-host pins before P-E opens |
| 8 Import escape | P-E | Escape fixture set refused with named codes |
| 9 Module graph semantics | P-E | Cycle/diamond/ambiguity/private-call fixtures across all peers |
| 10 C symbol collisions and ordering | P-E | Injective mangling test; reversed-order build byte-identical; non-module goldens unchanged |
| 11 Cache and diagnostic attribution | P-E, P-F | Library-only edit invalidates; library error cites library file |
| 12 Imports without copying | P-F | Single-copy static test and library mutation flips both consumers |
| 13 Exit classes | P-A | Single exit-class table; before/after pins; uniqueness test |
| 14 Stale hosted receipts | P-F (last plan) | Receipt revision equals code revision, path-scoped rule written |
| 15 Wired-not-run instruments | All | Seen-fail mutation per control; no zero-match `-run` |
| 16 Debt phase growth | P-F pre-audit | Public-route audit before final phase closes; plan-count review |

## Sources

- Repo (HIGH): `/Users/jon/projects/schway/.planning/PROJECT.md`, `AGENTS.md`, `.planning/RETROSPECTIVE.md`, `.planning/milestones/M005-MILESTONE-AUDIT.md` (N1-N4, W2/W3), `.planning/STANDING-VERDICTS.md` (process anti-pattern on bundling modules; cache key staleness precedent), `examples/checksum.schway` (provisional, additive sum, shadowing), `examples/phase23/adapter.c` (O_NONBLOCK, fstat/S_ISREG, limit probe, 1-byte hard-wiring, EmptyFile), `internal/compiler/cgen/cgen_program.go` (~1164-1195 fixed-shape file-byte glue, NULL-data check), `internal/compiler/check/check.go` ~2797-2830 and `check/scalar.go` ~96-106, 237 (`check.cfg_back_edge`), `internal/compiler/native/bindings.go` (`readBindingFile`, closed input manifest).
- Local computation (HIGH): Python `zlib.adler32`/`crc32` values in the table above, run 2026-10-10.
- RFC 1950 section 8.2 (Adler-32 definition: a init 1, b init 0, modulus 65521, result `b*65536 + a`) and zlib `adler32.c` (NMAX 5552 deferral) (MEDIUM, from memory, verify).
- ISO C17 6.5.6 (pointer arithmetic beyond one-past-end undefined), 6.2.5 (char signedness implementation-defined); Clang UBSan docs (`pointer-overflow`, `implicit-integer-sign-change`), `-fsigned-char`/`-funsigned-char` (MEDIUM, verify).
- Go 1.24 release notes: `os.Root`/`os.OpenRoot` traversal-resistant file access (MEDIUM, verify).
- Rust borrowck soundness history, for example rust-lang/rust#25860, and NLL problem cases (MEDIUM, as architectural precedent only).
- zlib CVE-2022-37434 (untrusted length driving a copy) as a length-handling precedent (MEDIUM/LOW, from memory).

---
*Pitfalls research for: M006 bounded byte utility and local modules in Schway*
*Researched: 2026-10-10*
