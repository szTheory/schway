---
verification_role: proposal
---

# Phase 31: Bounded File Checksum and Second Consumer - Research

**Researched:** 2026-10-12
**Domain:** Go 1.24 stdlib-only compiler (check / corevalidate / pathoracle / originvalidate / cgen / interp / native / session / cmd) plus a copy-local C17 adapter built by Clang. No new dependencies.
**Confidence:** HIGH on every repo fact (opened or probed this session; probe files lived in the scratchpad and were not committed). MEDIUM on design recommendations tagged `[ASSUMED]` in the Assumptions Log.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
D-01..D-15 in `31-CONTEXT.md` are locked (auto-followed from advisor fan-out); M006 research Decisions 1-10 and Phase 30 D-05 stay settled. Summary of what the planner must honor verbatim:
- D-01: `git mv` `examples/phase30/adler32.schway` and `adler32_cli.schway` into `examples/checksum/`; rename the CLI header to `examples.checksum.adler32_cli`; one atomic re-pin; fold README; leave dated receipts untouched.
- D-02: `examples/checksum.schway` (provisional frontier) is NOT touched.
- D-03: second consumer `examples/checksum/adler_low.schway` prints `adler32(view) % 65536`; `Wikipedia` -> `920` vs `300286872`.
- D-04: static single-copy scan with frozen sha256 allow-list, per-app merged-program assertion, dual flip, reverse control.
- D-05: `fn main(path: PathToken) -> U64`; no new flag; path passed as given; pin a leading-`-` path.
- D-06: copy-local adapter triple (acquire/view/release) via `schway.local-c/1`; a THIRD parallel shape in cgen, pathoracle, originvalidate (and check, corevalidate); Phase 23 shape untouched.
- D-07: interpreter reads independently through a host byte route keyed by path (open `O_RDONLY|O_NONBLOCK`, `Stat().Mode().IsRegular()`, `io.LimitReader` 65,537, never `os.ReadFile`/stat size); shared case matrix on both engines.
- D-08: every file outcome exits 65; no new exit value. D-09: `data FileError = | FileTooLarge | FileUnavailable | AllocFailed`, stderr by the existing generated `Type.Alternative` convention. D-10: release exactly once on every path.
- D-11: trailing `Modules []ModuleDigest{Module, Path, Digest}` `json:"modules,omitempty"` on `BuildIdentityInputs`, `BuildReceipt`, `ApplicationEvidenceReport`; only when imports exist.
- D-12: oracle rows `{input, want, cmd, python}`; D-13: Phase 29 harness extended with the listed controls; D-14: cap cases run on every `go test -race`; D-15: zero human UAT.

### Claude's Discretion
Adapter module/file/symbol names and owner type (`FileBytes` suggested); plan count and wave split (D-01 re-pin early and atomic); helper layout of the shared case matrix.

### Deferred Ideas (OUT OF SCOPE)
`run --engine=X FILE -- INPUT`; sysexits split 66/71/74; imported foreign declarations; the two reviewed todos (parameterised-main `run --engine`, formatter trailing-result merge). Phase 32 owns the guide and hosted receipts.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| APP-11 | `checksum FILE` exact result via public route | Third owner shape (Section "Third owner shape"), adapter design, bare-decimal writer already in application shell |
| APP-12 | Named errors, empty stdout, fixed path-free stderr, release exactly once | Stderr emission analysis, FileError mapping, hook-based release counting, ASan lane |
| MOD-06 | Second app imports one copy; build identity and evidence hash every module file | D-03 expressibility finding (MUST READ), D-11 identity mechanics, D-04 scans |
| EVD-14 | Independent oracle + every reached control observed to fail | testsupport row design, Phase 29 harness reuse, inventory meta-test |
</phase_requirements>

## Summary

The phase is mostly copy-and-specialise of Phase 29/30 machinery, but research found one blocking gap and several corrections to CONTEXT's premises. Gap: **D-03 as written cannot be expressed today.** The only admitted way to pass a `ByteView` to a callee is the closed fixture-style caller shape (acquire, view, ONE call, release, return). A `ByteView`-parameter function must be a scalar body (the parser's `scalarExprAtom` admits no user calls), and a generic `let x = f(view)` body with a ByteView parameter is refused. Probe result: `check.unexecutable_shape [...]: parameter shape has no native execution lowering this phase`. The `% 65536` therefore has to be applied in `main` as a second call (`U64 -> U64`) after the view call, which requires the new third shape to admit one optional trailing scalar call. This should be decided in the tracer plan, not discovered late (it changes the operation count in all five matchers).

Corrections: (1) there is no generic `Type.Alternative` stderr mechanism. The only occurrences are hard-coded strings in the Phase 24 emitter; the fixture shape exits 65 with NO stderr; the Phase 23 adapter itself prints `schway_file_byte_acquire: <Name>` from C. The new shape must print `FileError.<Alternative>\n` itself and the new adapter must be silent. (2) The "5 registers re-pinned" for D-01 are allow-list map keys, not golden JSON bytes: no golden contains a phase30 path. (3) The D-11 receipt and `session.go` edits shift pinned call-site line numbers; put new wiring in new files.

**Primary recommendation:** Run a 9-plan, 5-wave sequence: atomic D-01 move first, then a tracer (Wikipedia through `build --manifest` + `app run` on native and via the interpreter route) that also fixes the optional trailing-call shape, then peers, error matrix, identity, second app, controls, closure.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Path argv intake and refusals | Generated C `main` + native runner | cmd/schway | `cgen_program.go:1063-1065` empty/over-4096 -> 65; `native_app.go:390-395` NUL/over-4096 tool errors |
| Bounded file read, regular-file test, owner alloc/free | Copy-local C adapter | none | FFI boundary; syscall hooks for fault injection |
| FileError stderr + exit 65 | Generated C (cgen third shape) | none | Adapter stays silent so stderr bytes come from one emitter |
| Interpreter file read | Host closure (testsupport) | interp (route plumbing) | Independence from the C adapter (D-07); interp never opens files itself |
| Shape admission | check, corevalidate, originvalidate, pathoracle (independent peers) | session admission | Peer independence is a repo law |
| Module digests | session (set -> inputs) | native (identity, receipt, evidence) | Single modload read feeds hashing |
| Oracle literals | testsupport rows | Go `hash/adler32` cross-check | Python not needed at test time |

## Standard Stack

No new libraries. Everything is Go 1.24.0 stdlib and Apple clang 21 / hosted clang (`go version` go1.24.0 darwin/arm64, clang 21.0.0 observed) `[VERIFIED: local probe]`. Python 3.14.4 is present locally and is the generating oracle `[VERIFIED: python3 --version]`. Go `hash/adler32` is the second in-process oracle (D-12) `[CITED: pkg.go.dev/hash/adler32]`.

### Package Legitimacy Audit
No external packages are installed. **Packages removed due to SLOP:** none. **Packages flagged SUS:** none.

### Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Bounded read in Go | `os.ReadFile` / stat-size | `os.OpenFile(O_RDONLY\|syscall.O_NONBLOCK)` + `f.Stat().Mode().IsRegular()` + `io.LimitReader(f, 65537)` | st_size lies for growing files, FIFOs, devices |
| Exit literal | new numeric 65 in Go | `exitcode.AppInputRefused` | `TestNoDuplicatedExitLiterals` scans session, native, protocol, cgen, cmd non-test files `[VERIFIED: exitcode/scan_test.go:156-170]` |
| Native fault injection | new LD_PRELOAD scheme | adapter macro hooks (`OPEN`, `FSTAT`, `READ`, `MALLOC`, `FREE`, `CLOSE`) replaced at the include boundary in a temp copy | Proven in `native/phase23_observer_test.go:171-190` |
| Mutating generated C | custom compiler flags | `native.SetApplicationCSourceMutationForTest` | `export_phase29_test.go:15-19` |
| Mutating the adapter | seam on program.c | temp copy of `examples/checksum/` with adapter.c edited via exact-count replace, built with its own manifest | program.c seam does not touch the adapter TU |

## Third owner shape: exact matcher inventory

Facts read this session. The new symbol prefix must not collide: use `schway_file_bytes_` (plural). Phase 23 uses `schway_file_byte_` and type `FileByteOwner`; fixture uses `schway_fixture_bytes_` / `FixtureBytes`. Verify no `HasPrefix(..., "schway_file_byte")` without the trailing underscore exists before choosing the name `[ASSUMED: grep not exhaustive]`.

| Package | Site (file:line) | What it keys on today | Needed for the new shape |
|---|---|---|---|
| ability | `ability/ability.go:181-191` (`case "FixtureBytes"`) | type name; grants drop/share/send/escape, denies copy | add `case "FileBytes"` the same way |
| corevalidate | `corevalidate.go:4986-4992` (`case "FixtureBytes"` in `deriveAbility`) | independent copy | add `FileBytes` case (independent, not shared) |
| check | `check.go:4162-4164` (`if fixtureBytesEntryCandidate(function) { return checkFixtureBytesEntry(...) }`) precedes `if parameterType.Constructor == "PathToken"` at 4165 | first binding is `try_call` of `schway_fixture_bytes_acquire` (`check.go:4596-4606`) | new `fileBytesEntryCandidate` + `checkFileBytesEntry` + post-assembly law, inserted BEFORE the PathToken branch or the Phase 23 matcher will refuse it. Copy `checkFixtureOwnerPostAssembly` (do not add a fifth `loanLivenessFixpoint` caller: `TestSingleLoanLivenessLaw` allows exactly four) |
| corevalidate | `corevalidate.go:1469-1470` dispatch `if fixtureForeignContract(contract)` inside `validateLocalForeignOperation`; `1505-1515` marker = symbol prefix or ParameterType/ResultType name; `3011-3014` lifecycle dispatch | fixture markers | add a file-bytes marker check BEFORE the fixture check, own `validateFileBytesForeignOperation` and `fileBytesOwnerLifecycleValid`. `fixtureViewCall` exemption (1555) is NOT needed: it only matters when `function.Parameter.Type == function.ReturnType` (`corevalidate.go:2768`), and PathToken != U64 |
| originvalidate | `originvalidate.go:1349-1350` dispatch, `1665-1680` marker, `1694-1760` problem fn | fixture markers | own copy `fileBytesOwnerPresent` / `fileBytesOwnerOriginProblem` |
| pathoracle | `pathoracle.go:1343-1367` marker + `HasFixtureOwnerOperations`, `1376` `ValidateFixtureOwnerPaths`, `1453-1520` per-op pins | fixture markers | own `HasFileBytesOwnerOperations` / `ValidateFileBytesOwnerPaths` |
| cgen | `cgen_program.go:2842-2880` `hasFixtureOwnerFacts` / `checkFixtureOwnerProgram`; `2886-2925` validator; `2929-2985` emitter; call at `661`; emit dispatch at `972`; `#include "local/adapter.h"` gate at `~808` (`if localFileByteProgram \|\| fixtureOwnerProgram`) | fixture markers | new `checkFileBytesOwnerProgram`, `validate...`, `emitProgramFileBytesFunction`. ALSO: `hasLocalOwnerFacts` (`cgen_program.go:1566-1578`) returns true for any PathToken parameter, and `cgen_program.go:671-676` then errors `unsupported local-owner C shape` unless the Phase 23/24 shape matched; the new entry must be excluded from that gate. The application-entry admission at `678-681` (`u64Application \|\| localFileByteEntry`) must also accept the new shape |
| session | `session.go:1437-1445` `localOwnerEntry := entry.Parameter.Type == "PathToken" && hasLocalForeignOperations(program)` then `pathoracle.ValidateLocalOwnerPaths`; `fixtureOwnerEntry` branch; `validateLocalOperationBindings` follows | PathToken+foreign ops is assumed Phase 23 | add `fileBytesOwnerEntry` ahead of `localOwnerEntry`, else `ValidateLocalOwnerPaths` refuses the new program |
| interp | `interp.go:277` `RunWithFixtureBytes`; `frame.fixtureBytes` field at `985`; propagation at `1877`; `terminalLiveResources` `353`; dispatch `1733` `if top.fixtureBytes != nil`; `fixtureForeignCall` `~373` | U64 parameter, `source.u64` | copy to `RunWithFileBytes(program, name, path, read)` with a `fileBytes` frame field; parameter type `PathToken`; `inputValue` already returns `value{payload: input}` for a non-U64 non-data parameter (`interp.go:903`), so the path string is in `source.payload` |

**Where a PathToken entry also must pass:** `cgen.go:1525` `linearInput` already handles `PathToken`; the application main at `cgen_program.go:1063-1065` emits, verbatim, `if (argv[1][0] == '\\0' || strlen(argv[1]) > 4096u) return 65;` then `const char *schway_entry_input = argv[1];` `[VERIFIED: cgen_program.go:1063-1065]`.

### How failure stderr is emitted today (answer to research target 1)

- Phase 24 typed-error application only: `out.WriteString("  if (result.status != 0) { *error = \"AcquireError.AcquireFailed\"; return 0; }\n")` at `cgen_program.go:1209`, printed by `fprintf(stderr, "%s\n", error)` then `return 65`. This is the sole "Type.Alternative" emission and it is a hand-written literal keyed on `program.Module == "phase24.error"` (`cgen_program.go:650`).
- Phase 23 local-owner shape: `if (%s.status != 0) exit(65);` silent (`cgen_program.go:1807` area), and use failure prints `schway_file_byte_use: UnsupportedByte\n` via `fputs`. The Phase 23 ADAPTER also prints on acquire failure (`examples/phase23/adapter.c`: `fputs("schway_file_byte_acquire: ", stderr)` ... name ... `'\n'`).
- Fixture shape: `fmt.Fprintf(out, "  if (%s.status != 0) exit(65);\n", acquireResultName)` at `cgen_program.go:2955`: bare `exit(65)`, NO stderr. So the answer to "does the fixture shape differ": yes, it is silent.

**Consequence for D-09:** the "existing convention" is only the byte form `<Type>.<Alternative>\n`. The new emitter must generate it. Recommendation: the adapter returns `status` (0 ok; 1,2,3 mapped by position to `FileError` alternatives) and is silent; cgen emits a static name table built from `program.DataTypes` (alternatives are identifiers, quote with `strconv.Quote`) and `fputs(name); fputc('\n'); exit(65);`. Pinned bytes: `FileError.FileTooLarge\n`, `FileError.FileUnavailable\n`, `FileError.AllocFailed\n`. Empty-path and NUL/over-length refusals keep their existing output (empty path: silent 65 from C; over-4096/NUL: `schway app run: argument is N bytes; limit is 4096` / `argument contains an embedded NUL byte` on stderr, exit 65 via `appRunExitCode`, `cmd/schway/main.go:253-258`), neither contains the path `[VERIFIED: native_app.go:391,394]`.

Interpreter side: acquire failure returns `Outcome{Kind: "typed_failure", Value: <first alternative>}` hard-coded to `Alternatives[0]` (`interp.go:~383-395`). The file route must take the alternative from the host closure and validate it is in the `Fails` type's `Alternatives`. There is no interpreter stderr/exit; the parity harness maps typed_failure to `exit 65, stdout "", stderr "FileError."+Value+"\n"`. To avoid a tautology, compare BOTH engines to literal expected strings from the matrix, not to a shared formatter.

## D-03 expressibility finding (BLOCKING design input)

Probed with a built `schway` in the scratchpad (files deleted from repo scope):
1. `let r = sum % 65536` in a linear body: `syntax.expected_linear_result` (generic `let` takes call/ident only).
2. `var s = adler32(view)` in a scalar body: `syntax.expected_linear_result [..]: expected a scalar result` (`scalarExprAtom`, `parser.go:715-762`, admits only numbers, strings, identifiers, `byte_len`, `byte_at`).
3. `fn low(view: ByteView) -> U64 { let sum = adler32(view)  let r = low16(sum)  r }` parses but check refuses: `check.unexecutable_shape ... parameter shape has no native execution lowering this phase`. `scalarBodyStart` (`parser.go:~535-544`) only treats `let` as scalar for U64/ByteView parameters when the RHS is a literal/place/scalar op, and `scalarLetStart` rejects a call RHS.
4. `fn low16(sum: U64) -> U64 { var r = sum  r = r % 65536  r }` is fine.

So `adler32(view) % 65536` is only expressible as `main`: acquire, view, `adler32(view)`, release, then `low16(sum)`, return. Options: (A, recommended) the third shape admits an optional ONE trailing `U64 -> U64` call after the release (operation count 5 or 6; `checksum` uses 5, `adler_low` uses 6). (B) Amend D-03 to a pin that needs no post-operation: none found that satisfies "same file, different result, one library loop". The dual-flip proof needs a single loop, so duplicating the loop in the library is out. Treat (A) as a decision to surface to the user at plan time; it does not add language surface (the `%` operator and plain calls already exist), only widens a closed shape. Pin the shape in all five matchers with the 5-vs-6 operation distinction, and add a seeded fault per peer for a malformed trailing call.

## Architecture Patterns

### System Architecture Diagram
```
schway build examples/checksum/checksum.schway --manifest file_bytes.bindings.json --output A
  modload.Load(entry) -> Set{Program, Sources[entry, adler32...]}      (one read, sorted)
   -> check (3rd shape) -> corevalidate -> originvalidate -> pathoracle -> cgen.EmitApplication
   -> native.BuildApplication(modules digests) -> clang: program.c + adapter.c -> A + A.schway-build.json

schway app run A -- PATH            (argv[1] = PATH as given)
  host: len>4096 or NUL -> exit 65 (stderr, no path)
  C main: empty -> exit 65 | entry(path)
     acquire(path) -> adapter: open(O_RDONLY|O_NONBLOCK) -> fstat -> S_ISREG? -> malloc(65537)
                       -> read loop (EINTR retry, probe past 65536) -> status
        status!=0 -> stderr "FileError.<Alt>\n", exit 65, stdout empty
     view(owner) -> adler32(view) [-> low16] -> release(owner) -> stdout "<decimal>\n", exit 0

interpreter route (Go tests): CheckFile(entry) -> interp.RunWithFileBytes(program,"main",path,hostRead)
   hostRead(path) = OpenFile NONBLOCK -> Stat.IsRegular -> LimitReader(65537) -> (bytes | alternative name)
```

### Recommended structure
```
examples/checksum/
  adler32.schway  adler32_cli.schway   (git mv from phase30, header renamed)
  checksum.schway adler_low.schway
  adapter.c adapter.h file_bytes.bindings.json
  README.md  (folded Phase 30 README + new)
internal/compiler/testsupport/checksum_oracle.go   (rows {input,want,low,cmd,python}; file readers; case matrix)
internal/compiler/session/session_filebytes.go    (new wiring file; keep session.go edits minimal)
```

### Adapter design requirements (for D-06/D-10/D-13 consistency)
- Header guard `_POSIX_C_SOURCE 200809L` before includes (as `examples/phase23/adapter.c` does) because `applicationFlags` is `-std=c17 -Wall -Wextra -Werror -pedantic -O0` `[VERIFIED: native_app.go:112]` and the adapter is compiled with the same flags (`bindings.go:~537`).
- ABI header: owner `{unsigned char *data; uint64_t length;}`, acquire result `{int32_t status; owner}`, view result `{const unsigned char *data; uint64_t length;}` with the same `_Static_assert` layout pins as `examples/phase29/adapter.h`; acquire function type takes `const char *path`.
- Silent on failure (no `fputs`), unlike the Phase 23 copy. `path==NULL||path[0]==0` -> FileUnavailable (defence in depth).
- Allocation: to satisfy BOTH D-10 (a heap buffer exists when the oversize probe trips, so a double-release mutant is observable) and D-13 (an off-by-one view read must be flagged by ASan, which needs an EXACT-size heap block), use `malloc(65537)`, read, then `realloc(buf, max(1,n))` on success (ASan realloc-shrink moves and re-redzones). Oversize path frees once. Hook both `MALLOC` and `REALLOC`; a failed shrink should free and report AllocFailed. `[ASSUMED]` pending planner confirmation.
- `FIFO` and `/dev/null`: `O_NONBLOCK` makes FIFO open return immediately; `fstat` then `!S_ISREG` -> FileUnavailable. Deleting `S_ISREG` makes both rows checksum to `1` (EOF), so both rows catch that mutant.
- Empty file: success, `length 0`, non-NULL `data`, prints `1`.

## Common Pitfalls

1. **Matcher fall-through.** Without the `hasLocalOwnerFacts` exclusion (`cgen_program.go:671-676`) and the session `localOwnerEntry` ordering (`session.go:1437`), the new program is refused by Phase 23 laws with an unrelated message. Add the tracer first to find these.
2. **`env:` skip tokens are a closed set.** `debtRegisterEnvironmentalSet` at `session_test.go:2797-2806` currently holds `clang, otool, symlink, mkfifo, case-insensitive-fs`; the regex is `env:[a-z][a-z0-9_-]*` (`witness_registry_test.go:555`). The mode-000/root skip needs a new token (e.g. `root`) added to that set, FIFO rows cite `env:mkfifo`, ASan keeps `env:clang`.
3. **Registers keyed by line numbers.** `TestPhase16PublicEmitterConsumerInventory` pins `session.go:EmitNative:3098`; edits above it in `session.go` (the `session.go:1415-1520` build path) shift it. Also `cgen_program_test.go:EmitNative:N` rows. Keep session.go diffs line-neutral or re-pin; put new code in `session_filebytes.go`.
4. **`len(Functions) != 1` census.** `LANGUAGE-MATURITY.md:303` states `161 .schway programs, 5,282 lines total` and a guard census (19 non-test guards). Two new `.schway` files change both counts; write new Go matchers without that guard pattern.
5. **Phase 28 manifest.** `phase28NewlyAddedEntries` (`session_phase28_manifest_test.go:232`) needs rows for every new committed `.schway` (checksum.schway, adler_low.schway) and the two moved paths (keys AND `Path` fields change); an unregistered file fails with `absent from the golden and not in phase28NewlyAddedEntries`. Observe values by running the test; the walker uses the FILE routes (`CheckCommandFile`, `RunInterpreterCommandFile`, `RunNativeCommandFile`, `BuildApplicationFile` with a missing clang). `phase28MinCorpus = 155` is a floor only.
6. **Allow-lists to re-key (D-01), verified lines:** `session_phase272_identity_test.go:51`, `session_phase273_identity_test.go:67` (library only; entries with imports are refused by byte-taking `Check` so never join those walks), `session_phase30_prechange_test.go:67-68` (`phase30NewCommittedPaths`: must now list all FOUR examples/checksum `.schway` files, because the walker observes every `.schway` and the golden contains none of them), `session_phase28_manifest_test.go:290-309`, `cmd/schway/main_phase30_test.go:17,48,97,110`, `session_phase30_adler_test.go:26` (`phase30AdlerDir`). `testdata/phase30/prechange-corpus.json` has only the generated row `generated:phase30/no-import-over-work`; its bytes do NOT change (it is never regenerated). Header-only strings in `syntax_phase30_test.go:26,68,69` are inline constants, not file reads; leave them.
7. **Single-copy scan interplay.** `phase30AdlerDefinitions` scans a directory for the substring `fn adler32(` (`session_phase30_adler_test.go:~145`); after the move the new apps must not contain it (they call `adler32(view)`, which does not match `fn adler32(`). Frozen sha256 of the Phase 29 allow-list files, read this session: `examples/phase29/adler32_fixture.schway` = `96ae56eb593001ab77b3525396cb03dd013b35cc948516851d63758b3ff1a1bf`, `testdata/phase29/adler32_library.schway` = `ec1f3a40c576cc19d47563b5e5e76ca2213de1ad35a578845da9169c94237a36`; library file `317ac139...a36`-class hash is of the pre-move bytes and must be unchanged by a pure `git mv` (`317ac13972a53eccd04a7ab84cc914aff755e77b881db6fcaf77fa5240ebcb4b`).
8. **Native timeout is 30 s** (`native.go:30`); a non-`O_NONBLOCK` mutant on a FIFO hangs until then. Use `Runner{Timeout: ...}` or a context deadline for the mutant lane, and the goroutine plus `time.After` pattern with an unblocking writer used by `modload/resolve_unix_test.go:26-58` for Go-side FIFO rows.
9. **`--evidence=events` on an error exit.** `exit(65)` before the capture write leaves the evidence incomplete (`native.evidence_incomplete`); same as the fixture shape. Do not claim evidence records for failing runs.
10. **Usage text.** `app run ARTIFACT ... -- INPUT` is printed from `cmd/schway/main.go:186,197` and in the global usage `main.go:679`; the CLI cannot know the entry type before reading the receipt. Safest compliance with D-05: keep strings unchanged and document PATH in the README and `examples/checksum/README.md`; if the text is changed, grep tests that pin it first.
11. **Receipt size cap.** `maxApplicationReceiptBytes = 64 * 1024` (`native_app.go`); `Modules` rows are small, but assert the cap still holds for a receipt with modules.
12. **Formatter todo.** Keep the moved library byte-for-byte; write new sources with a plain `let ...` then result-identifier tail (the shape `adler32_cli.schway` already uses), avoiding the `var`-then-result pattern the pending todo says is merged.

## Runtime State Inventory (D-01 move)

| Category | Items Found | Action Required |
|----------|-------------|------------------|
| Stored data | None - verified by grep: no golden JSON contains a phase30 path | none |
| Live service config | None | none |
| OS-registered state | None | none |
| Secrets/env vars | None | none |
| Build artifacts | `.schway-build.json` receipts under TempDirs only; historical receipts in `.planning/` are intentionally untouched | none |

Live path sites (code, all verified above): cmd 4 sites; session 5 sites; docs: `.planning/LANGUAGE-MATURITY.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/STATE.md` mention phase30 (update only live rows; check each); `examples/phase30/README.md` folds into `examples/checksum/README.md` and its stale lines ("Module hashing ... is Phase 31 (MOD-06); the build receipt digests the entry file only") must be rewritten.

## D-07 interpreter route

- Copy `RunWithFixtureBytes` (do not generalize): new frame field `fileBytes func(path string) (data []byte, alternative string)`, set at `RunWithFileBytes`, propagated where `interp.go:1877` copies `fixtureBytes`, honored in `terminalLiveResources` (`353`) and the `OpForeignCall` dispatch (`1733`). Validate one acquire (`ParameterType == "PathToken"`) and one `ByteView` borrow exactly as `interp.go:~306-320` does.
- The host reader lives in `testsupport` (a non-test file, like `phase29_fixture_bytes.go`), not in session/cmd: `TestPhase30...` I/O inventory pins every `os.Open/OpenFile/ReadFile/Stat` in session and cmd production code (`session_phase30_inventory_test.go`, `phase30OSWant`), so a production reader there would need a classified row.
- Reader algorithm: `os.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)` (error -> `FileUnavailable`); `Stat()` and `Mode().IsRegular()` (else `FileUnavailable`); `io.ReadAll(io.LimitReader(f, 65537))`; read error -> `FileUnavailable`; `len > 65536` -> `FileTooLarge`. Directory open succeeds on both OSes and is rejected by `IsRegular`.
- Parity matrix cases: empty; `0x00`; `Wikipedia`; 65,536x`0xFF`; 65,537x`0xFF`; missing; directory; FIFO (deadline); `/dev/null`; mode-000 (skip when `os.Geteuid()==0`, token `env:root` once registered); path with a leading `-`; optional symlink-to-regular (follows) and growth-after-fstat (native only via READ hook).

## D-11 identity mechanics

`BuildIdentityInputs` fields end with `ExecutableDigest string json:"executable_digest"` (`bindings.go:322-338`); `ID()` hashes `json.Marshal(in)` with a domain tag `lang.build-identity/1` and length framing (`bindings.go:342-352`). A trailing `Modules ... json:"modules,omitempty"` leaves every no-import `BuildID` byte-identical. `BuildReceipt` (`native_app.go:29-47`) and `ApplicationEvidenceReport` (`native_app.go:87-98`) get the same trailing field; the report is built at `native_app.go:~345-349` from the receipt. `readApplicationReceipt` revalidation (`native_app.go:~656-680`) must also validate `Modules`: sorted unique modules, 64-hex digests, relative slash paths, equality of `receipt.Modules` and `Identity.Modules`; absent when no imports.
Session: `BuildApplicationFile` (`session.go:1515-1521`) gets entry bytes from `checkSourcePath` (`session_modules.go:36-46`) and passes them to `buildApplicationChecked(ctx, checked, source, ...)`. Add a sibling that also receives `modload.Set.Sources` (entry first, then sorted; entry `Module==""`, `RelPath` = base name; imported `Module`, `RelPath` e.g. `adler32.schway`). Recommend `Modules` list IMPORTED modules only (entry is covered by `SourceDigest`), RelPath only (never absolute). Add a native API variant rather than changing the variadic `BuildApplication(ctx, source, cSource, outputPath, manifestPath...)` signature. App verify (`session_app_verify.go`) is U64-only and unaffected.
Goldens affected: none, provided `omitempty` holds; a control test asserts a no-import build's receipt JSON has no `modules` key. Controls from D-11: comment-only library edit flips library digest and `BuildID`, not `EmittedCDigest` or entry digest; reversed import order gives identical `BuildID`.

## D-12/D-13 harness reuse (verified)

- Lanes and base flags: `phase29Lanes()`, `phase29BaseFlags = {"-std=c17","-Wall","-Wextra","-Werror","-pedantic"}`, `phase29BuildLane`, `phase29Run` (`native_phase29_test.go:36-79`), `phase29RequireASan` (probe compile, skip text begins `env:clang address sanitizer unavailable`), ASan flags `-O0 -fsanitize=address -fno-sanitize-ignorelist -fno-omit-frame-pointer`, `phase29LeakOption` (darwin forces `detect_leaks=0`).
- Source mutants: `phase29ReplaceOnce` (exact-count 1), mutants wrong-modulus, off-by-one-start, swapped-halves (`native_phase29_test.go:~214-260`). C-level: `SetApplicationCSourceMutationForTest`; the signed-char retype `phase29SignedCharMutation`.
- Generated bounds check form to delete under ASan: `if (<idx> >= <view>.length) {` then `fputs("schway: byte index out of bounds\n", stderr); exit(65);` `[VERIFIED: cgen_program.go:2742-2747]`. Kill-by-ASan recipe: apply the off-by-one `<` -> `<=` source mutant AND delete the bounds-check body at C level on a NON-empty file, assert stderr contains `AddressSanitizer: heap-buffer-overflow` and not the exit-65 line. Requires the exact-size owner allocation above.
- `-fsigned-char` equivalent-mutant lane: add the literal flag to `phase29BaseFlags`-style lane flags, rebuild checksum, assert PASS with pinned values on `0x80`/`0xFF` inputs (`zlib.adler32(b"\x80")` = 8454273 and `b"\x80\xff"` = 33620352 observed this session).
- Adapter-fault harness: copy `examples/checksum/` to a TempDir, insert `#define SCHWAY_FILE_BYTES_*` hook macros at the `#include "adapter.h"` boundary (`phase23_observer_test.go:171-190` technique), add an observer `.c` to the manifest `sources`, build with `native.BuildApplication` and the fault build helper. Counting FREE hook for release-exactly-once; READ hook returns `-1/EIO`; MALLOC/REALLOC returns NULL; a growth hook appends to the file after the first read.
- Meta-test pattern to copy for the control inventory: one-to-one subtest naming enforced by a table scan (see session meta-tests: `session_phase30_inventory_test.go`). CI already runs `go test -race -timeout=20m ./...` with no `-run` filter on ubuntu-latest and macos-latest (`.github/workflows/ci.yml`), so tests run hosted without extra wiring; no CI edit is needed.

### Oracle rows (generated this session with Python 3.14.4)
Command form: `python3 -c 'import zlib;print(zlib.adler32(b"..."))'`.

| Input | adler32 | `% 65536` (adler_low) |
|---|---|---|
| empty | 1 | 1 |
| `\x00` | 65537 | 1 |
| `Wikipedia` | 300286872 | 920 |
| `\xff`x65536 | 2006388466 | 3826 |
| `\xff`x65537 (never printed; would be 2273841137) | FileTooLarge | FileTooLarge |
| `\x80` | 8454273 | 129 |
| `\x80\xff` | 33620352 | 384 |

(`Wikipedia`: A=920, B=4582 matches D-03.) Rows live in `testsupport` as `{Input, Want, Low, Cmd, Python}` with a meta-test for non-empty `Cmd` and `Python` and an `hash/adler32` cross-check test.

## Exit table (target 6)

`exitcode.go:17-20` documents `AppInputRefused = 65` ("app run refused its input argument ... never renumber") and `README.md:93` has the row `| \`AppInputRefused\` | 65 | Input refused before launch, or a generated program's checked failure; never renumbered. |`. `TestExitCodeDocsTable` (`session_phase28_exitdocs_test.go:71-93`) checks names and values, so extending the DESCRIPTION text with the three outcome names is safe; no new constant, `TestExitCodeTable` and `TestNoDuplicatedExitLiterals` stay unchanged. The Phase 28 entry/exit-class manifest is per-.schway-file, not per-code: new `examples/checksum/*.schway` DO need `phase28NewlyAddedEntries` rows (pitfall 5). Outcome names also go in the checksum expected-matrix data in testsupport.

## Plan and wave proposal (target 7)

Phase 30 used 10 plans. AGENTS.md says growing plan counts need a named witness; this proposal is 9 (0-8), each tied to a requirement or safety obligation.

| Wave | Plan | Content | Reqs |
|---|---|---|---|
| 1 | 31-00 | Atomic D-01 move + all re-pins (6 test files, 5 registers, docs rows), no behavior change; prechange golden byte-identical | MOD-06 |
| 2 | 31-01 | **Tracer**: adapter (acquire/view/release, silent), `FileBytes` ability, check + corevalidate + cgen shape incl. optional trailing call, session ordering, `checksum.schway`, `Wikipedia` -> `300286872` through `build --manifest` + `app run` on native, and `RunWithFileBytes` on the interpreter. Gate verdict on D-03 shape option A. | APP-11 |
| 3 | 31-02 | Independent peers: originvalidate + pathoracle shapes, per-peer seeded faults, session admission | APP-11 |
| 3 | 31-03 | Error matrix native: FileError stderr bytes, hook adapter harness, release-once, EIO/alloc/growth rows | APP-12 |
| 3 | 31-04 | Host reader + shared case matrix, both engines, FIFO deadline, mode-000, `/dev/null`, leading `-`, cap rows (D-14) | APP-11, APP-12 |
| 4 | 31-05 | D-11 identity/receipt/evidence `Modules` + controls + receipt validation | MOD-06 |
| 4 | 31-06 | `adler_low`, single-copy scan, per-app assertion, dual flip, reverse control | MOD-06 |
| 5 | 31-07 | Oracle rows + meta-test, EVD-14 controls (modulus, init, swapped halves, `<=`, signed-char lane, ASan deleted check, neighbors, library mutation) + inventory meta-test | EVD-14 |
| 6 | 31-08 | Closure: exit description, README/LANGUAGE-MATURITY counts, PRODUCT-ROADMAP rows, VALIDATION finalization, session meta-test set | all |

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing` (go 1.24.0); clang required, ASan and mkfifo gated by `env:` skips |
| Config file | none |
| Quick run command | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/cgen ./internal/compiler/interp ./internal/compiler/native ./internal/compiler/session ./cmd/schway -run 'Phase31' -count=1` |
| Session meta set | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|Reconciliation\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestPhase30PreChangeGolden\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable\|TestNoSuppressionOutlivesItsWitness\|TestPhase16PublicEmitterConsumerInventory' -count=1` |
| Convention tests to run before hosted dispatch | `go test ./internal/compiler/native -run 'TestSourceNeverSpawnsUnboundedProcesses' -count=1` and `go test ./internal/compiler/check -run 'TestSingleLoanLivenessLaw' -count=1` and `go test ./internal/compiler/exitcode -count=1` |
| Full suite command | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...`); do not run the full suite locally (project policy) |

### Phase Requirements -> Test Map (proposed names)
| Req / SC | Behavior | Test Type | Automated Command | Exists |
|---|---|---|---|---|
| APP-11 / SC1 | empty->1, 0x00->65537, Wikipedia->300286872, 65,536xFF->2006388466 via build+app run | public CLI | `go test ./cmd/schway -run 'TestPhase31ChecksumPublicRoute' -count=1` | Wave 0 |
| APP-11 / SC1 | same answers on interpreter route | session/native | `go test ./internal/compiler/session -run 'TestPhase31ChecksumInterpreter' -count=1` | Wave 0 |
| APP-12 / SC2 | 65,537 -> FileTooLarge; missing/dir/FIFO/device/mode-000 -> FileUnavailable; stdout empty; stderr fixed, path-free | cross-engine matrix | `go test ./internal/compiler/native -run 'TestPhase31CaseMatrix' -count=1` | Wave 0 |
| APP-12 / SC2 | release once on success, oversize, alloc fail, EIO; double-release mutant fails under ASan | native hooks + ASan | `go test ./internal/compiler/native -run 'TestPhase31ReleaseExactlyOnce' -count=1` | Wave 0 |
| APP-12 | size check uses limit+1 read: st_size-trusting mutant and growth-after-fstat killed | native mutant | `go test ./internal/compiler/native -run 'TestPhase31BoundedReadMutants' -count=1` | Wave 0 |
| MOD-06 / SC3 | single copy; per-app one adler32; dual flip; reverse control | session | `go test ./internal/compiler/session -run 'TestPhase31SingleCopyAndDualFlip' -count=1` | Wave 0 |
| MOD-06 / SC3 | Modules digests in identity, receipt, evidence; order independence; no-import byte identity | native/session | `go test ./internal/compiler/native ./internal/compiler/session -run 'TestPhase31ModuleDigests' -count=1` | Wave 0 |
| EVD-14 / SC4 | rows carry cmd+python; Go hash/adler32 agrees | testsupport | `go test ./internal/compiler/testsupport -run 'TestPhase31OracleRows' -count=1` | Wave 0 |
| EVD-14 / SC4 | each control observed to fail; inventory maps control to subtest | native + meta | `go test ./internal/compiler/native -run 'TestPhase31Controls' -count=1` | Wave 0 |
| peers | per-peer seeded faults for the third shape | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'TestPhase31' -count=1` | Wave 0 |
| D-01 | moved files hash-identical; prechange golden bytes unchanged | session | `go test ./internal/compiler/session -run 'TestPhase30PreChangeGolden\|TestPhase30LibraryIsSingleCopy' -count=1` | exists (re-pinned) |

Sampling: per task commit run the quick command for the touched package; per wave run the quick and session meta sets; phase gate is a green hosted `go test -race ./...` on macOS and ubuntu. Finalize `31-VALIDATION.md` (`TestPhase20ValidationLifecycle` fails on `status: draft`) before dispatch.

### Wave 0 Gaps
- [ ] all `TestPhase31*` files above; `testsupport/checksum_oracle.go`; adapter fault harness helper; `root` token added to `debtRegisterEnvironmentalSet`.

## Security Domain

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | argv bounds (4096, NUL, empty), regular-file check on the descriptor, limit+1 read |
| V12 Files and Resources | yes | `O_NONBLOCK` open, fstat on the fd (no path re-resolution), no st_size trust |
| V7 Error Handling and Logging | yes | fixed stderr text, never the path |
| V2/V3/V4/V6 | no | no auth, session, access control or crypto |

### Threat model seed
| Pattern | STRIDE | Mitigation / test |
|---|---|---|
| Path disclosure in errors or ASan/tool output | Information disclosure | assert distinctive path segment absent from stderr on every error row (`app run` host refusals included) |
| TOCTOU (file grows or is swapped after stat) | Tampering | fstat on the open fd; read-based probe past 65,536; growth hook test |
| FIFO/device/directory blocks or yields bogus data | DoS / Tampering | `O_NONBLOCK`, `S_ISREG`, deadline-bound tests; `/dev/null` row |
| st_size lie (sparse, proc files) | Tampering | never size-based; mutant kills it |
| Double free / leak on error path | Tampering/DoS | counting FREE hook, ASan double-release mutant |
| Symlink following | Tampering | pin: `open` follows symlinks (both engines); not a security boundary for a user-supplied path |
| Option-looking path | Tampering | after `--` as `argv[1]`; pin leading `-` row |
| Allocation failure | DoS | AllocFailed -> 65, no partial output |

## State of the Art
No external moving parts. `zlib.adler32` semantics (RFC 1950) are stable `[CITED: RFC 1950 section 8.2]`.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Option A (optional trailing `U64->U64` call in the third shape) is acceptable to the user as the way to express `% 65536` | D-03 finding | If rejected, D-03/MOD-06 need an amended second-app pin or a language step |
| A2 | Adapter uses `malloc(65537)` then `realloc(max(1,n))` to satisfy D-10 and ASan exact-size | Adapter design | ASan control could be vacuous, or FileTooLarge double-release mutant not constructible |
| A3 | Symbol prefix `schway_file_bytes_` collides with no existing prefix check | Matcher inventory | Phase 23 matchers could capture the new shape |
| A4 | `Modules` lists imported modules only (entry covered by `SourceDigest`) | D-11 | Dual-flip assertions differ if the entry is included |
| A5 | New `phase28NewlyAddedEntries` values for checksum.schway/adler_low.schway follow the adler32_cli pattern (check exit 0, build 3, run 3) | Pitfall 5 | Observed values may differ; record what the test observes |
| A6 | Hosted ubuntu has working ASan; local macOS ASan probe compiled and ran this session | Harness | Lane skips with `env:clang`; EVD-14 deleted-bounds-check evidence then only runs locally |

## Open Questions (RESOLVED)

1. **Option A vs amending D-03** (see A1) - recommend surfacing at plan start; tracer plan carries the gate. RESOLVED: option A adopted, 31-CONTEXT D-16.
2. **Usage text for PATH** - leave strings unchanged (pitfall 10) unless the user wants the text changed. RESOLVED: usage text unchanged, PATH documented in the README, 31-CONTEXT D-18.
3. **Include the entry in `Modules`?** Recommend no (A4). RESOLVED: the orchestrator chose yes (every module, entry included), 31-CONTEXT D-19.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | all | yes | 1.24.0 | none |
| clang | native lanes | yes | Apple clang 21.0.0 | tests skip `env:clang` |
| ASan runtime | deleted-bounds / double-release | yes (local probe ran) | - | skip `env:clang address sanitizer...` |
| Python 3 | regenerating oracle rows only | yes | 3.14.4 | not needed at test time |
| mkfifo | FIFO rows | yes (darwin/linux) | - | skip `env:mkfifo` |
| non-root user | mode-000 row | yes locally (uid 501) | - | skip with `env:root` |

## Sources

### Primary (HIGH)
- Repository files read this session: `31-CONTEXT.md`, `cgen_program.go`, `interp.go`, `check.go`, `corevalidate.go`, `native_app.go`, `bindings.go`, `native_phase29_test.go`, `export_phase29_test.go`, `phase23_observer_test.go`, `examples/phase23/adapter.{c,h}`, `examples/phase29/*`, `exitcode/*`, session register tests, `ci.yml`, `parser.go`.
- Local probes: Python zlib values, built `schway` check/build probes of the D-03 forms.
### Secondary / Tertiary
- None; `[CITED]` references above are standard library and RFC facts.

## Project Constraints (from CLAUDE.md / AGENTS.md)
Follow AGENTS.md; stdlib-only, prefer copy-local over dependency; macOS and Linux; deterministic evidence with negative controls and distinct native lanes (unoptimized, optimized, oracle, sanitizer); no unaudited unwind across the C ABI; work through GSD commands; read `.planning/PRODUCT-ROADMAP.md` and `LANGUAGE-MATURITY.md` at phase transitions; independent peers must not share helpers (spike findings: independent implementations plus armed fault injection).

## Metadata
**Confidence breakdown:** stack HIGH (none new); architecture HIGH (read and probed); pitfalls HIGH; plan split MEDIUM.
**Research date:** 2026-10-12. **Valid until:** 2026-11-11 (repo-local facts; invalidated by edits to the cited files).
