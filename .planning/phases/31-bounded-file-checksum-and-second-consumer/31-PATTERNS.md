# Phase 31: Bounded File Checksum and Second Consumer - Pattern Map

**Mapped:** 2026-10-12
**Files analyzed:** 22 new/modified (grouped)
**Analogs found:** 21 / 22 (all tracked via `git ls-files`)

Governing rule: every new shape is a PARALLEL COPY of the Phase 29 fixture shape (5 ops) in each peer; never generalize Phase 23 / Phase 29 matchers in place and never share helpers between peers. New wiring goes in new files (`session_filebytes.go`) to keep pinned `session.go` line numbers stable (`TestPhase16PublicEmitterConsumerInventory`).

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match |
|---|---|---|---|---|
| `examples/checksum/adler32.schway`, `adler32_cli.schway` (git mv) | library/entry | transform | `examples/phase30/*` (same bytes) | exact |
| `examples/checksum/checksum.schway`, `adler_low.schway` | entry | file-I/O | `examples/phase30/adler32_cli.schway` | exact (PathToken param instead of U64) |
| `examples/checksum/adapter.c` | adapter | file-I/O | `examples/phase23/adapter.c` | exact (silent, bounded 64 KiB, realloc) |
| `examples/checksum/adapter.h` | ABI header | - | `examples/phase29/adapter.h` | exact (rename prefix, `const char *path`) |
| `examples/checksum/file_bytes.bindings.json` | config | - | `examples/phase29/fixture_bytes.bindings.json` | exact |
| `check.go` (`fileBytesEntryCandidate`, `checkFileBytesEntry`, post-assembly) | validator | transform | `check.go:4162` `checkFixtureBytesEntry` | exact |
| `corevalidate.go` (`validateFileBytesForeignOperation`, lifecycle, ability) | validator | transform | `corevalidate.go:1469-1515, 3011, 4986` fixture branch | exact |
| `ability/ability.go` | policy | - | `ability.go:181-191` `case "FixtureBytes"` | exact |
| `originvalidate.go` (`fileBytesOwnerPresent/OriginProblem`) | validator | transform | `originvalidate.go:1665-1760` fixture | exact |
| `pathoracle.go` (`HasFileBytesOwnerOperations`, `ValidateFileBytesOwnerPaths`) | validator | transform | `pathoracle.go:1343-1520` fixture | exact |
| `cgen_program.go` (`checkFileBytesOwnerProgram`, validator, emitter, FileError stderr table) | codegen | request-response | `cgen_program.go:2842-2985` fixture | exact + new stderr emitter |
| `interp.go` `RunWithFileBytes` | runtime | request-response | `interp.go:277` `RunWithFixtureBytes` | exact |
| `session/session_filebytes.go` + minimal `session.go:1437` ordering | service | request-response | `session.go:1437-1445` fixtureOwnerEntry branch | role-match |
| `native` `Modules []ModuleDigest` (bindings.go, native_app.go) | model | transform | `evidence.go:68-75` `ForeignDigest` omitempty; `bindings.go:322-352` | role-match |
| `testsupport/checksum_oracle.go` (rows, host reader, case matrix) | utility | file-I/O | `testsupport/phase29_fixture_bytes.go` | role-match |
| `native/phase31_*_test.go` (matrix, mutants, controls, ASan) | test | batch | `native_phase29_test.go:36-345` | exact |
| adapter fault harness helper | test | event-driven | `native/phase23_observer_test.go:162-230` + `testdata/phase23_observer.c` | exact |
| `session_phase31_*_test.go` (single copy, dual flip) | test | batch | `session_phase30_adler_test.go:20-200` | exact |
| `cmd/schway/main_phase31_test.go` (public route) | test | request-response | `cmd/schway/main_phase30_test.go` | exact |
| D-01 re-pin sites (see list) | test config | - | existing rows | edit |
| `session_test.go:~2797` `root` token | config | - | `debtRegisterEnvironmentalSet` | edit |
| Interp stderr for typed_failure | - | - | none (hard-coded `Alternatives[0]`) | no analog |

## Pattern Assignments

### `examples/checksum/checksum.schway` and `adler_low.schway` (entry, file-I/O)

**Analog:** `examples/phase30/adler32_cli.schway` (moved verbatim to `examples/checksum/adler32_cli.schway`, header renamed to `examples.checksum.adler32_cli`).

Copy the import/export/foreign shape; change `case: U64` -> `path: PathToken`, `FixtureBytes` -> `FileBytes`, `FixtureError` -> `FileError`, symbols `schway_file_bytes_*`:

```
module examples.checksum.checksum
import adler32 {
  fn adler32
}
export {
  fn main
}
foreign C {
  fn schway_fixture_bytes_acquire(case: U64) -> FixtureBytes {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_fixture_bytes_acquire_fn
    allocator: "libc_malloc"
    release: "schway_fixture_bytes_release"
    fails: FixtureError
  }
  fn schway_fixture_bytes_view(owner: FixtureBytes) -> ByteView { ... mode: "borrow" ... }
  fn schway_fixture_bytes_release(owner: FixtureBytes) -> Unit { ... mode: "consume" ... allocator: "libc_malloc" }
}
data FixtureError =
  | FixtureMissing
fn main(case: U64) -> U64 {
  let owner = try schway_fixture_bytes_acquire(case)
  let view = schway_fixture_bytes_view(owner)
  let sum = adler32(view)
  sum
}
```
(`adler32_cli.schway` lines 1-47.) Body is `let ... then result identifier` (avoids the formatter todo). `data FileError = | FileTooLarge | FileUnavailable | AllocFailed` order is load-bearing: adapter status 1,2,3 map by position (RESEARCH). Note the analog returns `sum` with NO release in source; the closed shape's release is a compiler-inserted OpRelease (5 ops: acquire, view, call, release, return). `adler_low` adds `let low = low16(sum)` and a `fn low16(sum: U64) -> U64 { var r = sum  r = r % 65536  r }` helper defined in the entry (D-16); 6 ops. Library `adler32.schway` is not edited (sha256 `317ac139...cb4b` must stay unchanged by the move).

### `examples/checksum/adapter.c` (adapter, file-I/O)

**Analog:** `examples/phase23/adapter.c` (read-only reference). Copy-local, do not edit Phase 23.

**Prologue and hooks** (lines 1-34): `_POSIX_C_SOURCE` guard (needed for `-std=c17 -pedantic -Werror`), `#include "adapter.h"`, then `#ifndef SCHWAY_..._X / #define SCHWAY_..._X real_fn / #endif` for OPEN FSTAT MALLOC FREE READ CLOSE. Rename the macro prefix to `SCHWAY_FILE_BYTES_` and add `REALLOC` (RESEARCH A2: `malloc(65537)` -> read -> `realloc(max(1,n))`).

**Open / regular-file test** (lines 89-94), copy as is:
```c
if (path == NULL || path[0] == '\0') return acquire_failure(SCHWAY_FILE_BYTE_INVALID_PATH);
descriptor = SCHWAY_FILE_BYTE_OPEN(path, O_RDONLY | O_NONBLOCK);
if (descriptor < 0) return acquire_failure(SCHWAY_FILE_BYTE_OPEN_FAILED);
if (SCHWAY_FILE_BYTE_FSTAT(descriptor, &details) != 0 || !S_ISREG(details.st_mode)) {
  return acquire_cleanup_failure(descriptor, NULL, SCHWAY_FILE_BYTE_NOT_REGULAR);
}
```
**EINTR-retry read loop** (lines 99-119): `if (amount > 0) {...continue;} if (amount == 0) ...; if (errno == EINTR) continue; return ...READ_FAILED`. Change to a loop up to 65537 bytes; `received > 65536` -> FileTooLarge (cleanup frees the buffer exactly once).
**Cleanup-once helper** (lines 70-78): free data if non-NULL, close, close failure overrides status. **Release** (lines 145-147): `SCHWAY_FILE_BYTE_FREE(owner.data)`.

**Divergence (D-17):** drop `acquire_failure`'s `fputs(... stderr)` (lines 62-68) and `acquire_status_name`; the adapter is silent and returns only `status` 0..3 (collapse OPEN/NOT_REGULAR/READ/CLOSE failures to 2 = FileUnavailable, alloc to 3). Drop `use`; add `view` returning `{data,length}` (Phase 29 `adapter.h` view result).

### `examples/checksum/adapter.h` and `file_bytes.bindings.json`

**Analog:** `examples/phase29/adapter.h` (lines 1-50) and `examples/phase29/fixture_bytes.bindings.json`. Copy the struct trio (owner 16B, acquire_result 24B, view_result 16B), all `_Static_assert` layout pins, and the three `typedef ..._fn` types; change acquire to `(const char *path)`; rename `schway_fixture_bytes_` -> `schway_file_bytes_` (plural; avoid Phase 23 `schway_file_byte_` prefix). Manifest (all 12 lines) is `schway.local-c/1` with `sources:["adapter.c"]`, `headers`, `include_dirs:["."]`, three `symbols` with `function_type`, `runtime_dependencies:["platform-c-runtime"]`.

### Peers: check / corevalidate / originvalidate / pathoracle / ability (validator, transform)

Each: add a `FileBytes` marker set (symbol prefix `schway_file_bytes_` OR `ParameterType/ResultType == "FileBytes"`) and a 5-or-6 operation closed matcher, copying the fixture one in the SAME package with its own code. Dispatch order is critical (new branch BEFORE the Phase 23 / PathToken branch).

- **check:** `check.go:4162-4165`:
  ```go
  if fixtureBytesEntryCandidate(function) {
  	return checkFixtureBytesEntry(functionID, function, foreignSymbols, calleeContracts)
  }
  if parameterType.Constructor == "PathToken" {
  ```
  Insert `fileBytesEntryCandidate` ahead of the PathToken branch. Copy `checkFixtureOwnerPostAssembly`; do NOT add a fifth `loanLivenessFixpoint` caller (`TestSingleLoanLivenessLaw`).
- **corevalidate:** dispatch `corevalidate.go:1469-1470` (`if fixtureForeignContract(contract)`), marker `1505-1515`, lifecycle `3011-3014`, ability `4986-4992`. Add own `fileBytes*` copies placed before the fixture check. `fixtureViewCall` exemption (1555) NOT needed.
- **ability:** `ability.go:181-191` add `case "FileBytes":` identical to `"FixtureBytes"` (grants drop/share/send/escape, denies copy).
- **originvalidate:** marker `originvalidate.go:1665-1680` (uses slice-prefix compare, no `strings`) and step-machine `fixtureOwnerOriginProblem` 1694-1760:
  ```go
  fail := func(operationID string) *Problem {
  	return &Problem{Code: "core.fixture_owner_origin", Detail: fmt.Sprintf("%s: invalid fixture owner operation %q", function.ID, operationID), FunctionID: function.ID}
  }
  ```
  New code `core.file_bytes_owner_origin`; parameter type `PathToken`; accept optional trailing `U64->U64` call after the release (D-16) and refuse second trailing call / call taking the view.
- **pathoracle:** marker + `HasFixtureOwnerOperations` (`pathoracle.go:1343-1367`):
  ```go
  const prefix = "schway_fixture_bytes_"
  return strings.HasPrefix(contract.Symbol, prefix) || contract.ParameterType == "FixtureBytes" || contract.ResultType == "FixtureBytes"
  ```
  Add `HasFileBytesOwnerOperations` / `ValidateFileBytesOwnerPaths` (1376-1520 pattern: start from parameter, replay acquire -> borrow -> call -> release).

Per-peer seeded-fault tests: copy the existing fixture seeded-fault tests in each package (`grep -n Fixture *_test.go` per package) so each peer refuses a malformed trailing call independently.

### `cgen_program.go` third shape (codegen)

**Analog:** `cgen_program.go:2842-2985` (`hasFixtureOwnerFacts`, `checkFixtureOwnerProgram`, `validateFixtureOwnerFunction`, `emitProgramFixtureBytesFunction`).

**Validator core** (2886-2925): whole-struct compare of each contract:
```go
wantAcquire := core.ForeignOperationContract{
	Symbol: "schway_fixture_bytes_acquire", ABIType: "schway_fixture_bytes_acquire_fn", Mode: "acquire",
	ParameterType: "U64", ResultType: "FixtureBytes", Fails: "FixtureError",
	Allocator: "libc_malloc", Release: "schway_fixture_bytes_release", Unwind: "forbidden", NonlocalExit: "forbidden",
}
...
if acquire.Kind != core.OpForeignCall || acquire.Foreign == nil || *acquire.Foreign != wantAcquire {
```
Change `ParameterType` to `PathToken`, `ResultType` to `FileBytes`, `Fails` to `FileError`, entry `Parameter.Type == "PathToken"`, operations 5 or 6.

**Emitter** (2929-2985): reuse the whole structure (names via `newCNames`, view-struct copy, `schway_record_event` calls, paired release). One edit: line 2955
```go
fmt.Fprintf(out, "  if (%s.status != 0) exit(65);\n", acquireResultName)
```
becomes a status-to-name table print `fputs("FileError.<Alt>", stderr); fputc('\n', stderr); exit(65);` (names from `program.DataTypes` alternatives, `strconv.Quote`). Pinned bytes: `FileError.FileTooLarge\n`, `FileError.FileUnavailable\n`, `FileError.AllocFailed\n`.

**Gates to patch (RESEARCH pitfall 1):** `cgen_program.go:671-676` `hasLocalOwnerFacts(entry)` errors `unsupported local-owner C shape` for any PathToken param unless Phase 23/24 matched; exclude the new entry. `678-681` application-entry admission must accept the new shape; `~808` `#include "local/adapter.h"` gate; call-site `661`; emit dispatch `972`. PathToken argv at `1063-1065` is already correct:
```go
out.WriteString("  if (argv[1][0] == '\\0' || strlen(argv[1]) > 4096u) return 65;\n")
out.WriteString("  const char *schway_entry_input = argv[1];\n")
```

### `interp.go` `RunWithFileBytes` (runtime, request-response)

**Analog:** `RunWithFixtureBytes` (`interp.go:277-330` region): nil-route check, `corevalidate.Validate`, `findFunction`, entry-shape check, count acquires/borrows, `inputValue`, `newFlatFrame`, set `base.fixtureBytes = bytes`, `runProgramFrameStack`, live-resources check (`isByteIndexDefect` exemption).
```go
if acquires != 1 || borrows != 1 {
	return Execution{}, fmt.Errorf("fixture bytes route requires one fixture acquire and one ByteView borrow in %q", functionName)
}
```
New: parameter `PathToken`, acquire cases keyed `ParameterType == "PathToken"`, route `func(path string) ([]byte, string)` (alternative name validated against `Fails` type's `Alternatives`, unlike the hard-coded `Alternatives[0]` at ~383-395). Touch points: frame field (`~985`), dispatch `1733` (`if top.fixtureBytes != nil { fixtureForeignCall(...) }`), propagation `1877` (`result.frame.fixtureBytes = top.fixtureBytes`), `terminalLiveResources` `353`. `inputValue` returns `value{payload: input}` for non-U64 (`interp.go:903`).

### `testsupport/checksum_oracle.go` (utility, file-I/O)

**Analog:** `testsupport/phase29_fixture_bytes.go`. Copy: Go-side independent generator written separately from C (switch/`bytes.Repeat`), pinned expected table with a comment naming the external oracle (Python 3.14.4 `zlib.adler32` + Go `hash/adler32`), exported accessor. Differences (D-12): rows become structs `{Input, Want, Low, Cmd, Python}`; add the host reader (`os.OpenFile(p, os.O_RDONLY|syscall.O_NONBLOCK, 0)`, `Stat().Mode().IsRegular()`, `io.ReadAll(io.LimitReader(f, 65537))`) as non-test code here, not in session/cmd (the Phase 30 I/O inventory pins `os.Open*` there). Values: RESEARCH "Oracle rows" table.

### `native/phase31_*_test.go` (test, batch) - lanes, mutants, ASan, case matrix

**Analog:** `native_phase29_test.go`.
- Lanes (36-45): `phase29Lanes()`, `phase29BaseFlags`; `phase29BuildLane` (50-62) swaps `native.SetApplicationFlagsForTest(flags)` and calls `session.BuildApplicationFile(ctx, sourcePath, artifact, native.DefaultRunner(), manifestPath)`. `phase29Run` (64-72) uses `native.RunApplication(ctx, artifact, input, &stdout, &stderr)` - pass the PATH string. Never `t.Parallel()` with seams.
- ASan gate (`phase29RequireASan`, ~86-98): probes `clang -fsanitize=address` and `t.Skipf("env:clang address sanitizer unavailable ...")`; reuse in new deleted-bounds-check control with `-O0 -fsanitize=address -fno-sanitize-ignorelist -fno-omit-frame-pointer`, `phase29LeakOption`.
- Source mutants (`phase29ReplaceOnce` ~214, table ~232-248): exact-count-1 replace on temp copy; add rows `var a = 1 -> var a = 0`, `<` -> `<=` (asserts exit 65 + OOB stderr, `phase29AssertOOB`). Must also write the mutated tree to `t.TempDir()` with the adapter + manifest copied (see `phase30CopyDir`).
- C-mutants: `native.SetApplicationCSourceMutationForTest(phase29SignedCharMutation(t))` via `export_phase29_test.go:15-19`; the retype mutant stays the must-fail control, and a `-fsigned-char` lane (extra flag in a lane) must PASS on `0x80`/`0xFF`.
- "Killed" structure: loop vectors, compute `differs`, `t.Fatalf` if survived; always include a `control` subtest (unmutated passes). Copy this triple (mutant / equivalent-mutant-passes / control).
- Inventory meta-test: one-to-one control-name -> subtest (pattern in `session_phase30_inventory_test.go`).

### Adapter fault harness (test, event-driven)

**Analog:** `native/phase23_observer_test.go:162-230` + `native/testdata/phase23_observer.c`. Copy: copy adapter/header/manifest into `t.TempDir()`, `strings.Replace(adapter, "#include \"adapter.h\"\n", observerHooks, 1)` with `#define SCHWAY_..._MALLOC phase31_observer_malloc` etc. (fail if unchanged), append observer `.c` to `manifest.Sources`, rewrite JSON, build, run. Observer `.c` pattern (testdata/phase23_observer.c): `_POSIX_C_SOURCE`, `atexit` finalizer, receipt file from env var, counting alloc/free. Add counting FREE (release once), READ hook returning `-1/EIO`, MALLOC/REALLOC NULL, growth-after-first-read hook. New observer file goes in `native/testdata/phase31_observer.c`.

### `session_phase31_*_test.go` (single copy, dual flip)

**Analog:** `session_phase30_adler_test.go`. `phase30CopyDir` (24-40), `phase30MutateLibrary` (44-), `phase30AdlerDefinitions` (WalkDir scanning `fn adler32(`, ~139-160), `TestPhase30LibraryIsSingleCopy` (~165-): `session.CheckFile(entry)`, count functions named `adler32`, assert `function.Module == "adler32"`. Retarget `phase30AdlerDir()` to `examples/checksum` (D-01). New: per-app loop over both entries, dual-flip (mutate in TempDir, compare both pins and BuildIDs), reverse control (B-only mutation flips `checksum` not `adler_low`), repo-wide scan with sha256 allow-list: `examples/phase29/adler32_fixture.schway` = `96ae56eb...a1bf`, `testdata/phase29/adler32_library.schway` = `ec1f3a40...7a36` (full hashes in RESEARCH pitfall 7).

### Identity / receipt / evidence `Modules` (model, transform)

**Analog:** additive-omitempty precedent `evidence.go:68-75` (`ForeignDigest ... json:"foreign_digest,omitempty"`, trailing, set only when applicable) and identity struct `bindings.go:322-352`:
```go
	RuntimeDependencies []string        `json:"runtime_dependencies"`
	ExecutableDigest    string          `json:"executable_digest"`
}
func (in BuildIdentityInputs) ID() string { h := sha256.New(); encoded, _ := json.Marshal(in) ... "lang.build-identity/1" ...
```
Add trailing `Modules []ModuleDigest \`json:"modules,omitempty"\`` to `BuildIdentityInputs`, `BuildReceipt` (`native_app.go:29-47`), `ApplicationEvidenceReport` (`native_app.go:87-98`). `InputID()` zeroes only `ExecutableDigest`; unaffected. Populate in a new session sibling of `buildApplicationChecked` from `modload.Set.Sources` (sorted via `sort.Strings`; D-19 includes the entry). Add native API variant (do not change variadic `BuildApplication`). Validate `Modules` in `readApplicationReceipt` (~656-680).

### `session.go:1437` admission ordering (service)

**Analog:** `session.go:1437-1445`:
```go
localOwnerEntry := entry.Parameter.Type == "PathToken" && hasLocalForeignOperations(program)
u64Application := ...
fixtureOwnerEntry := entry.Parameter.Type == "U64" && entry.ReturnType == "U64" && pathoracle.HasFixtureOwnerOperations(program)
...
if fixtureOwnerEntry { pathoracle.ValidateFixtureOwnerPaths(program) } else if err := pathoracle.ValidateLocalOwnerPaths(program)
```
Add `fileBytesEntryOnly := pathoracle.HasFileBytesOwnerOperations(program)` and make `localOwnerEntry` false for it; branch to `ValidateFileBytesOwnerPaths`. Keep diff line-neutral or move into `session_filebytes.go`.

## Shared Patterns

### Exit literal 65
**Source:** `exitcode.AppInputRefused` (`internal/compiler/exitcode/exitcode.go:17-20`). In Go non-test code use the constant (`TestNoDuplicatedExitLiterals`); generated C emits `exit(65)` literal as existing emitters do. No new constant.

### Source-mutant exact-count replace
**Source:** `native_phase29_test.go` `phase29ReplaceOnce`. Apply to all EVD-14 source mutants.

### Skip tokens
**Source:** `session_test.go:~2797` `debtRegisterEnvironmentalSet` (`clang, otool, symlink, mkfifo, case-insensitive-fs`). Add `"root": true` in the same change as the mode-000 skip (`t.Skip("env:root ...")`); FIFO rows cite `env:mkfifo`; ASan `env:clang`.

### FIFO deadline
**Source:** `modload/resolve_unix_test.go:26-58` (goroutine + `time.After` + unblocking writer). Native mutant lane uses Runner/context timeout (default 30 s, `native.go:30`).

### D-01 re-pin site list (atomic plan 31-00)
`cmd/schway/main_phase30_test.go:17,48,97,110`; `session/session_phase30_adler_test.go:26`; `session_phase272_identity_test.go:51`; `session_phase273_identity_test.go:67`; `session_phase30_prechange_test.go:67-68` (`phase30NewCommittedPaths` must list all FOUR `examples/checksum/*.schway`); `session_phase28_manifest_test.go:290-309` plus `phase28NewlyAddedEntries` (~232) rows for new `checksum.schway` and `adler_low.schway`. Do not touch `examples/checksum.schway`, `session_phase20_test.go:14`, `testdata/phase30/prechange-corpus.json`, inline strings in `syntax_phase30_test.go:26,68,69`. `LANGUAGE-MATURITY.md:303` counts and `PRODUCT-ROADMAP.md` live rows move with the new `.schway` files.

## No Analog Found

| File / Concern | Role | Reason |
|---|---|---|
| Typed-failure stderr `FileError.<Alt>\n` in cgen | emitter | No generic emitter exists; sole occurrence is a hard-coded literal at `cgen_program.go:1209` (`phase24.error` only). Fixture shape exits 65 silently (`cgen_program.go:2955`). New code: static name table from `program.DataTypes`. |
| Optional trailing `U64->U64` call (D-16) in a closed shape | matcher | Fixture shape has exactly 5 ops; extend to 5-or-6 independently in check, corevalidate, originvalidate, pathoracle, cgen. |
| Interpreter alternative-from-route | runtime | `interp.go:~383-395` hard-codes `Alternatives[0]`; new route returns and validates the name. Compare both engines against literal strings, not a shared formatter. |

## Metadata

**Analog search scope:** `examples/phase23,29,30`, `internal/compiler/{cgen,check,corevalidate,originvalidate,pathoracle,ability,interp,native,session,testsupport,evidence}`, `cmd/schway`.
**Tracked check:** `git ls-files` confirmed for `examples/phase23|29|30/*`, `native_phase29_test.go`, `phase29_fixture_bytes.go`, `session_phase30_adler_test.go`, `main_phase30_test.go`, `phase23_observer_test.go`, `evidence.go`, `native/testdata/phase23_observer.c`. Line numbers are as observed 2026-10-12; edits shift them.
**Pattern extraction date:** 2026-10-12
