# Phase 31: Bounded File Checksum and Second Consumer - Context

**Gathered:** 2026-10-12
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase delivers two things:
- A `checksum` application. A caller runs `schway app run <artifact> -- PATH`
  on one regular file of at most 65,536 bytes and gets the exact Adler-32 value
  as bare decimal followed by `\n`, or a named error. On every error stdout is
  empty and stderr is fixed text that never contains the path.
- A second application that imports the same `adler32` library module and
  prints a different pinned result. The library is not copied.

Build identity and evidence hash every module file in sorted order. Every
pinned answer comes from Python `zlib.adler32`, with its generating command
recorded. Every reached EVD-14 control is observed to fail.

The following are out of scope here:
- a new CLI flag;
- a new exit-code value;
- a `read_file` builtin;
- `run --engine` input;
- the clean-checkout guide and hosted receipts, which belong to Phase 32.

</domain>

<decisions>
## Implementation Decisions

Every decision below came from advisor research, with one researcher per gray
area at the `minimal_decisive` tier. Each was **auto-followed**, per the user's
standing preference to fan out, synthesize and auto-follow. Where two
researchers disagreed, the synthesis is recorded inline.

M006 research Decisions 1–10 in `.planning/research/SUMMARY.md` stay settled
and are not repeated here. That applies in particular to:
- Decision 4 (dynamic bounds checks, OOB exit 65);
- Decision 5 (cap 65,536);
- Decision 6 (adapter shape: `O_RDONLY|O_NONBLOCK`, `fstat` on the descriptor,
  `S_ISREG`, an `EINTR`-retrying read loop with a limit+1 probe, never
  `st_size`, `max(1,n)` allocation, the empty file succeeds with 1);
- Decision 8 (second consumer, path-only).

Phase 30 D-05 (the root is the entry file's directory) is the constraint that
drives D-01.

### Library home and app layout
- **D-01:** Use `git mv` to move `examples/phase30/adler32.schway` and
  `examples/phase30/adler32_cli.schway` into `examples/checksum/`.
  - `checksum.schway`, the second consumer and the moved Phase 30 CLI then sit
    side by side, all importing one `examples/checksum/adler32.schway`.
  - Under D-05 an entry can only import files below its own directory, so every
    consumer has to sit next to the library. This is the same pattern as Go's
    `cmd/*`, a dune library with several executables, and Zig.
  - Rename the moved CLI's header to `examples.checksum.adler32_cli` in the
    same change that re-pins the goldens.
  - Re-pin the live path sites:
    - `main_phase30_test.go`
    - `session_phase28_manifest_test.go`
    - the 27.2/27.3 identity corpora
    - `session_phase30_adler_test.go`
    - the D-15 pre-change golden
    - the Phase 30 corpus registers
  - Fold `examples/phase30/README.md` into `examples/checksum/README.md`.
  - Leave dated Phase 30 receipts, SUMMARYs and VERIFICATION records
    untouched, because they are historical. Update only the live rows,
    `LANGUAGE-MATURITY.md` and `PRODUCT-ROADMAP.md`.

  — **Reversibility:** costly — this re-pins about 6 test files and 5
  registers, and the planner must sequence it as one atomic re-pin commit.

  Rejected alternative: freezing the Phase 30 copy and adding a second
  library. That leaves two live importable copies, which contradicts MOD-06
  and the Phase 30 "imported rather than copied" claim.
- **D-02:** `examples/checksum.schway` (the provisional frontier) is **not
  touched**. `TestPhase20ChecksumFrontier` and the Phase 28 exit manifest pin
  it byte for byte. Phase 32 retires it. The new app lives at
  `examples/checksum/checksum.schway`.

### Second consumer (MOD-06)
- **D-03:** The second consumer is `examples/checksum/adler_low.schway`. It
  prints `adler32(view) % 65536`, the A half.
  - It uses only the existing `%` operator.
  - Its distinguishing pin is `Wikipedia` → `920`, against `300286872` from
    `checksum`. That matches the documented A=920 and B=4582.
  - Empty and `0x00` inputs print `1` in both apps, so they are not the
    "different result" pin.
  - It uses the same path input and the same adapter as `checksum`.
- **D-04:** Single copy and the dual-flip proof:
  - **Static scan.** Scan every `*.schway` under `examples/` and `testdata/`
    for `fn adler32(` together with `% 65521`. Expect exactly
    `examples/checksum/adler32.schway`, plus a frozen allow-list of the two
    Phase 29 files (`examples/phase29/adler32_fixture.schway` and
    `testdata/phase29/adler32_library.schway`), each pinned by sha256 so the
    list cannot quietly grow.
  - **Per-app merged-program assertion.** For each app, assert exactly one
    `adler32` with `Module == "adler32"`. Keep the existing
    `copied-library-is-detected` subtest.
  - **Dual flip.** Copy `examples/checksum/` into `t.TempDir()` and change
    `var a = 1` to `var a = 2`. Then assert that both `checksum` and
    `adler_low` differ from their pins and both BuildIDs change. Assert that
    the library's `Modules` digest changes identically in both records and that
    the entry digests do not change.
  - **Reverse control.** A mutation that touches only B flips `checksum` but
    not `adler_low`. This proves both pins genuinely depend on the library.

### Path input and file adapter
- **D-05:** The entry is `fn main(path: PathToken) -> U64`. The entry's
  parameter type selects how argv is handled; this already exists in
  `cgen_program.go` around line 1061.
  - `app run ARTIFACT -- PATH` is unchanged. There is no `--path` flag.
  - The usage/help text should name PATH where the entry type is a path.
  - Existing refusals cover these cases: an empty path exits 65 in generated
    C; NUL and a path over 4096 bytes get the host refusals at
    `native_app.go:390-395`.
  - The path is passed as given: no canonicalization and no `filepath.Abs`.
  - Pin one new case, a path with a leading `-`, which is safe because it comes
    after `--` as `argv[1]`.
- **D-06:** The adapter is a copy-local triple in `examples/checksum/`,
  wired through the existing `build --manifest` schema (`schway.local-c/1`):
  - `acquire(path: PathToken) -> FileBytes fails FileError`, mode acquire;
  - `view(owner) -> ByteView`, mode borrow;
  - `release`, mode consume.

  The caller shape copies `examples/phase30/adler32_cli.schway`: `try
  acquire`, then `view`, then `adler32(view)`, then release, all in
  straight-line code, so no loan crosses a back edge.
  - The C adapter copies the Phase 23 adapter's macro-hooked syscalls
    (`OPEN`, `READ`, `MALLOC`, `FREE`, …) for fault injection.
  - cgen, pathoracle and originvalidate gain a **third parallel shape** next
    to the frozen Phase 23 `FileByteOwner` shape and the Phase 29/30
    `FixtureBytes` shape (ARCHITECTURE F5).
  - The Phase 23 matchers, adapter, receipts and observer tests stay
    untouched.

  Rejected alternative: a `read_file` builtin. It would be new language
  surface and would collapse the FFI evidence boundary.
- **D-07:** The interpreter reads the file independently through a
  host-supplied byte route in the `RunWithFixtureBytes` closure style
  (`interp.go:277`), keyed by path:
  - `os.OpenFile(p, O_RDONLY|syscall.O_NONBLOCK)`;
  - `f.Stat().Mode().IsRegular()`;
  - a read bounded to 65,537 bytes via `io.LimitReader`;
  - never `os.ReadFile`, never the stat size.

  Parity is one shared case matrix run on both engines. It compares exit,
  stdout and stderr byte for byte. The cases are:
  - empty, `0x00`, `Wikipedia`, 65,536×`0xFF`, 65,537×`0xFF`;
  - missing, directory, FIFO (with a deadline), `/dev/null`;
  - a mode-000 file, skipped with an `env:` reason when `os.Geteuid()==0`.

  The `/dev/null` row is mandatory: if `S_ISREG` were dropped it would
  checksum to `1`.

### File-error outcomes and exit codes
- **D-08:** Every file outcome exits **65** (`exitcode.AppInputRefused`).
  Outcomes are told apart by a fixed stderr line and a named alternative.
  **No new exit-code value** is added.
  - This is the Phase 23 precedent: every acquire failure exits 65.
  - Phase 28 marks 65 as "never renumber".
  - coreutils `cksum` and `sha256sum` likewise use one code for every file
    error.
  - APP-12 requires named outcomes, not distinct codes.
  - These outcomes are program semantics that are identical on both engines,
    so they are not exit 3 (tool failure).
  - Record the outcome names in the checksum expected-matrix JSON and in the
    exit-class table's row description for 65.

  Rejected alternative: a sysexits split into 66, 71 and 74. That adds rows to
  a one-way table and causes exit-pin ripple (PITFALLS 13).
- **D-09:** The outcome type is `data FileError = | FileTooLarge |
  FileUnavailable | AllocFailed`:

  | Outcome | Cause | Engines | Exit | Stdout |
  |---|---|---|---|---|
  | `FileTooLarge` | the 65,537th byte read succeeds | both | 65 | empty |
  | `FileUnavailable` | missing, `EACCES`, directory, FIFO, device, `fstat` failure, `EIO` or another read error mid-read, `close` failure | both | 65 | empty |
  | `AllocFailed` | `malloc(max(1,n))` returns NULL | native only, injected via the `MALLOC` hook | 65 | empty |

  - `EIO` is reachable only on native, through the `READ` hook, so it is
    tested there and kept out of the cross-engine comparison. The same applies
    to `AllocFailed`.
  - `EINTR` is retried and is never an outcome. A short read just continues
    the loop.

  **Stderr synthesis.** The two researchers disagreed. One proposed new
  `checksum: …` lines; the other pointed to the existing generated
  `Type.Alternative` failure line (`cgen_program.go:1209`, e.g.
  `AcquireError.AcquireFailed`).

  Resolution: use the **existing generated convention**, giving one line per
  alternative, such as `FileError.FileTooLarge` and
  `FileError.FileUnavailable`, in exactly the form cgen already prints. That
  way the interpreter and native produce byte-identical stderr from the same
  mechanism, and no new stderr emission path is added. The planner pins the
  exact bytes from the existing form.
- **D-10:** Release exactly once on every path:
  - On native, a counting `FREE` hook in the adapter.
  - On the interpreter, the event trace.
  - The controls cover the success path, `FileTooLarge` (the buffer exists
    when the probe trips), `AllocFailed`, and mid-read `EIO`. A
    double-release mutant on the oversize path runs under ASan and must be
    observed to fail.

### Build identity and evidence (MOD-06 hashing)
- **D-11:** Add `Modules []ModuleDigest{Module, Path, Digest}` as a
  **trailing** `json:"modules,omitempty"` field on `BuildIdentityInputs`,
  `BuildReceipt` and `ApplicationEvidenceReport`.
  - It is sorted by module name with `sort.Strings`.
  - It hashes each module's raw bytes from the single `modload` read (Phase 30
    D-10).
  - It is populated **only when the program has imports**, so every
    no-import ID, receipt and golden stays byte-identical. This follows the
    `ForeignDigest` omitempty precedent.
  - `SourceDigest` stays the entry-only digest.
  - `session.go:~1466` passes the `modload.Set` instead of the entry bytes.
  - Controls:
    - a comment-only library edit flips the library's `Modules` digest and the
      `BuildID`, while `EmittedCDigest` and the entry digest stay the same;
    - reversing the import order gives an identical `BuildID`;
    - existing no-import goldens stay unchanged.

  — **Reversibility:** costly — the receipt schema is additive, but once
  receipts are published the field shape is a contract.

### Oracle literals and EVD-14 controls
- **D-12:** Pinned literals become struct rows in testsupport:
  `{input, want, cmd, python}`.
  - `cmd` is the exact generating command, for example
    `python3 -c 'import zlib;print(zlib.adler32(b"\xff"*65536))'`.
  - `python` is the version, for example `Python 3.14.4`.
  - Python is not needed at test time. A `_test.go` cross-check against Go's
    `hash/adler32` acts as a second oracle.
  - A meta-test requires non-empty `cmd` and `python` on every row.
  - The rows include `adler_low`'s `% 65536` values.
- **D-13:** The control harness extends the Phase 29 one
  (`native_phase29_test.go:229-345`). Source-level mutants use `ReplaceOnce`
  with exact anchor counts on a temp copy of the app tree. C-level mutants use
  the `applicationCSourceMutationForTest` seam. New rows on top of Phase 29's:
  - **Wrong initial state:** `var a = 1` becomes `var a = 0`.
  - **Off-by-one loop bound:** `<` becomes `<=`. It is caught by the bounds
    check, so the test asserts exit 65 and the exact OOB stderr.
  - **Signed char.** A literal `-fsigned-char` flag is an *equivalent mutant*,
    because bytes are `const unsigned char *`. So:
    - the Phase 29 retype mutant stays the must-fail control;
    - a `-fsigned-char` flag lane must **pass** with the pinned value on the
      `0x80`/`0xFF` fixtures. That proves independence from the platform's
      `char` signedness.
  - **Deleted bounds check under ASan.** Built at `-O0 -fsanitize=address`
    with a **non-empty** out-of-bounds input. Stderr must contain
    `AddressSanitizer: heap-buffer-overflow` and must not contain the exit-65
    line. An empty input stays inside the `max(1,n)` allocation, so ASan would
    not fire. Reuse the `env:` skip in `phase29RequireASan`.
  - The refused loop neighbors reuse the Phase 29 fixtures, each asserting its
    exact code and exit 2.
  - The library mutation is D-04.
  - An inventory meta-test maps each EVD-14 control name one-to-one to a
    subtest, so no control is wired but never run. CI already runs
    `go test -race ./...` with no `-run` filter.
- **D-14:** The cap runs on every `go test -race` run with no `-short` gate
  and no env opt-in. The 65,536×`0xFF` file and the 65,537-byte file are
  generated in `t.TempDir()` and run on both engines through the public route
  on native. Spike 008 measured the worst hosted interpreter-race max at
  376 ms, against a 12 s budget. Wall time stays out of goldens (Phase 29
  D-12).

### Verification posture
- **D-15:** Zero human UAT (PROJECT.md § Verification Operating Preference).
  Every success criterion, outcome row and control is an automated
  `go test` lane running in hosted CI on macOS and ubuntu. The unreadable-file
  and ASan lanes skip only with cited `env:` reasons.

### Amendments after research (2026-10-12, auto-followed; resolves 31-RESEARCH open questions)
- **D-16 (amends D-03; resolves research Q1):** `adler32(view) % 65536` cannot be
  written in a scalar body today. User calls are not admitted there
  (`parser.go:715-762`), and a `let x = f(view)` body is refused with
  `check.unexecutable_shape`. Research option A is adopted. The new third
  owner shape (D-06) admits **one optional trailing `U64 -> U64` call** on the
  library result, e.g. `low16`, defined in the entry. That gives:
  - `checksum`: acquire, view, `adler32`, release, return (5 operations);
  - `adler_low`: the same plus `low16(sum)` (6 operations).

  It widens a closed shape, adds no language surface, and is decided in the
  tracer plan, so every matcher learns it at once. Each refused neighbour,
  such as a second trailing call or a call taking the view, gets a named
  refusal pin.
- **D-17 (amends D-09; resolves research finding 2):** There is no generic
  `Type.Alternative` stderr emitter. The new cgen emitter for the third shape
  prints `FileError.<Alternative>\n` itself, and the C adapter stays silent.
  The interpreter prints the same bytes from the same alternative name. The
  three stderr byte strings are pinned exactly as given in 31-RESEARCH.md.
- **D-18 (amends D-05; resolves research Q2):** The `app run` usage text is
  unchanged, because the CLI cannot know the entry type before reading the
  receipt. PATH input is documented in `examples/checksum/README.md`.
- **D-19 (amends D-11; resolves research Q3):** `Modules` lists **every**
  module, including the entry, sorted by module name. This satisfies MOD-06's
  "hash every module in sorted order" literally. `SourceDigest` is kept
  unchanged, and the redundancy is accepted. The list is populated only when
  the program has imports, so no-import bytes are unchanged.
- **D-20:** The mode-000 root skip uses a new closed-set `env:` token (for
  example `root`), added to the allow-list at `session_test.go:~2797` in the
  same change.
- **D-21 (2026-10-12, amends D-11's import-order control; plan-check
  revision 1):** D-11's control "reversing the import order gives an
  identical `BuildID`" cannot hold as written. Reversing the import
  declarations changes the entry file's bytes. Under D-11 `SourceDigest` is
  the entry digest, and under D-19 the entry row of `Modules` carries the
  same digest; both feed `BuildIdentityInputs.ID()`, so the `BuildID` must
  change. The control is replaced by these assertions (plan 31-05):
  - **Set function.** `moduleDigests` over one loaded `modload.Set`, with
    the imported sources (`Sources[1:]`) presented in reversed order and the
    entry kept first, returns a slice identical to the unreversed call.
  - **Two real builds.** Two entries that differ only in import declaration
    order are built. Their real receipts (`<artifact>.schway-build.json`)
    have identical library rows (every `modules` row except the entry row),
    identical row order, and identical `EmittedCDigest` and
    `ExecutableDigest`.
  - **Identity.** After zeroing exactly `SourceDigest` and the entry row's
    `Digest` in both receipts' `Identity`, the two `ID()` values are equal.
    No other field, `ExecutableDigest` included, is zeroed.

### Claude's Discretion
- The exact module, file and symbol names of the adapter triple and the
  owner type (`FileBytes` is suggested).
- The plan count and wave split. The D-01 re-pin should be an early, atomic
  plan.
- The exact helper layout of the shared case matrix.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Scope and requirements
- `.planning/ROADMAP.md` § Phase 31 — goal, success criteria 1–4, research flag
- `.planning/REQUIREMENTS.md` — APP-11, APP-12, MOD-06, EVD-14
- `.planning/research/SUMMARY.md` — Resolved Decisions 1–10 (settled), pinned literal table, pitfalls 1–5, 12–15
- `.planning/research/ARCHITECTURE.md` — the F5 file-bytes owner shape and the `Program.Modules[]` proposal (lines ~94, ~230)
- `.planning/research/PITFALLS.md` — #5 (file semantics), #13 (exit-pin ripple), #15 (wired-not-run)

### Prior phase decisions
- `.planning/phases/30-local-modules/30-CONTEXT.md` — D-05 (the root is the entry directory), D-10 (single read), D-11 (Unit), D-15 (no-import golden), the deferred list
- `.planning/phases/29-byte-views-and-the-adler-32-library/29-CONTEXT.md` — byte views, the mutant harness, D-12 (wall time out of goldens)
- `.planning/phases/28-exit-classes-verify-report-and-entry-pruning/28-CONTEXT.md` — the exit-class table and 65 "never renumber"
- `.planning/spikes/008-byteview-cap/README.md` — cap timings

### Code anchors
- `internal/compiler/exitcode/exitcode.go` — the exit table (`TestExitCodeTable`, `TestNoDuplicatedExitLiterals`)
- `internal/compiler/cgen/cgen_program.go` — the PathToken argv path at ~1061, the `Type.Alternative` failure line at ~1209, the fixture shape at ~2862-2975, the Phase 23 shape at ~1180/1585/1795
- `internal/compiler/native/native_app.go` — the input refusals at ~390, `BuildIdentityInputs` at ~213/252; `internal/compiler/native/bindings.go` at ~342 (the ID marshal)
- `internal/compiler/session/session_modules.go` (~35-41) and `session.go` (~1422, ~1466) — the modload set and identity inputs
- `internal/compiler/interp/interp.go` — `RunWithFixtureBytes` at ~277
- `examples/phase23/adapter.c` — the macro-hooked syscall pattern to copy (read-only reference; do not edit)
- `examples/phase30/` — the library, the CLI and the README to move (D-01)
- `internal/compiler/native/native_phase29_test.go` (~36-42 lanes, ~93 skips, ~229-345 mutants) and `internal/compiler/testsupport/phase29_fixture_bytes.go`
- `internal/compiler/session/session_phase30_adler_test.go` (~54 mutation helper, ~174 single copy)
- `internal/compiler/session/session_phase20_test.go:14` — the frontier pin that must stay untouched (D-02)

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- PathToken entry argv handling already exists in cgen and the native runner, so no CLI change is needed.
- The `schway.local-c/1` bindings manifest and the Phase 29/30 acquire/view/release caller shape.
- The Phase 23 adapter's macro hooks for syscall fault injection, as a copy-local template.
- The Phase 29 mutant harness (`ReplaceOnce`, the C-source mutation seam, the ASan lanes, `phase29RequireASan`).
- The `ForeignDigest` omitempty precedent for additive identity fields.
- The `modload.Set.Sources` list, already sorted by name with the entry first.

### Established Patterns
- Name-keyed shape matchers in cgen, pathoracle and originvalidate, kept as parallel copies (never generalized in place).
- App-level failures exit 65 with fixed stderr, and the differential harness compares exit, stdout and stderr.
- Skips must carry a cited `env:` reason (`TestNoSuppressionOutlivesItsWitness`).
- A new exit literal must be a named constant (`TestNoDuplicatedExitLiterals`); D-08 avoids adding one.

### Integration Points
- cgen, pathoracle and originvalidate: the third owner shape.
- interp: a path-keyed byte route.
- session/native: module digests feed identity, receipts and evidence.
- `examples/checksum/`: the new home for the library and all three consumers.

</code_context>

<specifics>
## Specific Ideas

- Pins:

  | Input | `checksum` | `adler_low` |
  |---|---|---|
  | empty | 1 | 1 |
  | `0x00` | 65537 | 1 |
  | `Wikipedia` | 300286872 | 920 |
  | 65,536×`0xFF` | 2006388466 | `% 65536` of 2006388466, generated by the Python command |
  | 65,537×`0xFF` | `FileTooLarge` | `FileTooLarge` |

  The value 2273841137 must never be printed.
- Adversarial rows to keep:
  - a FIFO without `O_NONBLOCK` hangs, so the FIFO row runs with a deadline;
  - a file that grows between `fstat` and `read` is caught by the read-based
    size check;
  - an `st_size`-trusting mutant;
  - `>=` vs `>` at the 65,536/65,537 boundary.

</specifics>

<deferred>
## Deferred Ideas

- `run --engine=X FILE -- INPUT`, giving the interpreter a public input route.
  Not required by APP-11 or APP-12; "every engine" is met at Go level, as in
  Phases 29 and 30.
- A sysexits-style exit split (66/71/74). Rejected for M006; reopen only with a
  dedicated exit-table change.
- Imported foreign declarations, so the adapter block isn't repeated in
  `checksum` and `adler_low`. Out of M006 scope; the block is the adapter
  contract, not the library loop.

### Reviewed Todos (not folded)
- **run --engine on a parameterised main reports an operational failure**
  (`.planning/todos/pending/2026-10-12-run-parameterised-main-reports-operational-failure.md`).
  Left pending: it is outside APP-11/12, it would re-pin the Phase 28 manifest,
  and the Phase 31 public route is `build` + `app run`.
- **Formatter merges a trailing result identifier into the preceding var
  assignment**
  (`.planning/todos/pending/2026-10-12-formatter-merges-trailing-result-into-var-assignment.md`).
  Not folded, because it is a formatter correctness fix outside the Phase 31
  requirements. D-01 moves the library verbatim (not reformatted). Best done as
  `/gsd-quick` or in Phase 32, before the guide documents `format`.

</deferred>

---

*Phase: 31-bounded-file-checksum-and-second-consumer*
*Context gathered: 2026-10-12 (advisor fan-out, auto-followed)*
