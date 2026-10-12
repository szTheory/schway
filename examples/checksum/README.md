# Checksum examples: a bounded file checksum, a second consumer and the Phase 30 fixture CLI

Two ordinary applications read one file and share one Adler-32 library:

- `checksum` prints the full Adler-32 of a file of at most 65,536 bytes.
- `adler_low` prints only the low 16 bits (the A half) of the same checksum.

Both import `adler32.schway`; the loop exists in exactly one place.

## Layout

- `adler32.schway` is the library: `module adler32`, exporting `fn adler32`. It is
  pure scalar: no data type, no foreign block, no `main`.
- `checksum.schway` and `adler_low.schway` are the two entries. Each imports
  `fn adler32`, declares the `FileBytes` foreign owner block, the `FileError` data
  type and `fn main(path: PathToken)`.
- `adapter.c`, `adapter.h` and `file_bytes.bindings.json` are the copy-local
  adapter triple and its `schway.local-c/1` manifest. The adapter acquires the
  bytes, lends a view and releases them; it writes nothing to stdout or stderr.
- `checksum.expected.json` is the machine-readable case table.
- `adler32_cli.schway` is the Phase 30 fixture CLI (see below).

Every consumer sits beside the one library because an import resolves only to a
file under the entry's own directory. The foreign block is repeated in each entry
because imported foreign declarations are not supported.

## checksum

Run these commands from the repository root with Go 1.24 and Clang installed.

```sh doc-test=clean-checkout
go build -o ./schway ./cmd/schway
mkdir -p .schway/checksum
./schway build examples/checksum/checksum.schway \
  --manifest examples/checksum/file_bytes.bindings.json --output .schway/checksum/checksum
./schway build examples/checksum/adler_low.schway \
  --manifest examples/checksum/file_bytes.bindings.json --output .schway/checksum/adler_low
./schway build examples/checksum/adler32_cli.schway \
  --manifest examples/phase29/fixture_bytes.bindings.json --output .schway/checksum/adler32-module
```

Outputs live under the git-ignored `.schway/` directory, so the tree stays clean.
Then run the checksum program on a file:

```sh doc-test=run app=checksum
./schway app run .schway/checksum/checksum -- PATH
```

PATH is passed as given after `--`; a leading `-` in it is a path, not a flag.
The `app run` usage line still says INPUT (D-18): the argument is a file path for
this program, and the usage text is unchanged. A path of 4,097 bytes or more, a
NUL byte or an empty argument is refused before launch with exit 65.

Limit: files up to 65,536 bytes are read; a file of 65,537 bytes is refused.
The read asks for one byte past the limit and never trusts the size the file
reports.

## adler_low

```sh doc-test=run app=adler_low
./schway app run .schway/checksum/adler_low -- PATH
```

`adler_low` calls the same `adler32` and then a trailing `low16` call that keeps
the low 16 bits (the A half of the checksum), so the result is the checksum
modulo 65,536 (D-16).

## Results

Answers come from Python `zlib.adler32` (Python 3.14.4) and are cross-checked by
Go `hash/adler32`; they are never derived by running Schway. The 65,537-byte case
is refused, so no checksum is printed for it. Each row is keyed by its case name in
`checksum.expected.json`.

| Case | checksum | adler_low | Python command (checksum) |
| --- | --- | --- | --- |
| `empty` | 1 | 1 | `python3 -c 'import zlib;print(zlib.adler32(b""))'` |
| `zero` | 65537 | 1 | `python3 -c 'import zlib;print(zlib.adler32(b"\x00"))'` |
| `wikipedia` | 300286872 | 920 | `python3 -c 'import zlib;print(zlib.adler32(b"Wikipedia"))'` |
| `ff65536` | 2006388466 | 3826 | `python3 -c 'import zlib;print(zlib.adler32(b"\xff"*65536))'` |
| `ff65537` | `FileError.FileTooLarge` | `FileError.FileTooLarge` | n/a (refused before any checksum) |
| `missing` | `FileError.FileUnavailable` | `FileError.FileUnavailable` | n/a (no such file) |
| `directory` | `FileError.FileUnavailable` | `FileError.FileUnavailable` | n/a (not a file) |

For adler_low the same command ends with ` % 65536` inside the `print`. The inputs
0x80 and 0x80 0xFF give checksums 8454273 and 33620352 (adler_low 129 and 384),
which pins unsigned byte handling.

Tests: `TestPhase31ChecksumPublicRoute`, `TestPhase31ChecksumInterpreter`,
`TestPhase31AdlerLowPublicRoute`, `TestPhase31OracleRows` and
`TestPhase31ExpectedMatrixJSON`.

## Errors

Every error exits 65 (`AppInputRefused`) with empty stdout and a single fixed
stderr line. The path is never echoed (`TestPhase31CaseMatrix`).

| Outcome | stderr | Causes |
| --- | --- | --- |
| `FileTooLarge` | `FileError.FileTooLarge` | more than 65,536 bytes |
| `FileUnavailable` | `FileError.FileUnavailable` | missing file, directory, FIFO, `/dev/null`, mode 000, or a read, stat or close failure |
| `AllocFailed` | `FileError.AllocFailed` | the buffer could not be allocated |

Other unreadable inputs give `FileError.FileUnavailable`, exit 65; FIFO, mode-000, `/dev/null`, read-error and allocation-failure cases are pinned by the Phase 31 matrix tests, not executed by this guide.

The adapter releases its buffer exactly once on success, on oversize, on
`AllocFailed` and on a read error (`TestPhase31ReleaseExactlyOnce`). The bounded
limit-plus-one read is pinned by `TestPhase31BoundedReadMutants`. The shape
refusals for neighbouring loops are pinned by `TestPhase31LoopNeighborRefusals`,
and each EVD-14 control by `TestPhase31Controls` with `TestPhase31ControlInventory`.

## Single copy and identity

`TestPhase31SingleCopyAndDualFlip` shows the loop exists once, that editing the
library changes both applications, and that a library edit which only the
`checksum` result depends on leaves `adler_low` at 920. Build identity, the receipt and the evidence report carry a
sorted `modules` list whenever the program imports
(`TestPhase31ModuleDigests`); a program with no import is byte-identical to
before. The moved library bytes are pinned by `TestPhase31MovedLibraryIdentity`.

## Phase 30 fixture CLI

`adler32_cli.schway` is the Phase 29 Adler-32 program with the library imported.
It keeps the Phase 29 `foreign C` owner and `data FixtureError`, and builds with
the Phase 29 manifest.

```sh doc-test=run app=adler32_cli
./schway app run .schway/checksum/adler32-module -- CASE
```

| Case id | adler32_cli |
| --- | --- |
| `3` | 300286872 |
| `10` | refused, no output |

Case 3 prints `300286872`; case 10 is outside the table and exits 65. The tests
are `TestPhase30AdlerModuleAppCLI`, `TestPhase30AdlerModuleInterpreter` and
`TestPhase30LibraryIsSingleCopy`.

## Earlier programs

`sum_to_n` and FizzBuzz keep their own doc pins in [`../phase27/README.md`](../phase27/README.md). The Adler-32 fixture (`adler32_fixture`) and the out-of-bounds control (`byte_index_oob`) keep theirs in [`../phase29/README.md`](../phase29/README.md) (`TestPhase29ExampleReadmeCaseTable`) and are receipt programs in a later Phase 32 plan. This guide does not run them.

## Known limits

- `schway run` on a parameterised `main` is unsupported; use `build` plus `app run`
  ([todo](../../.planning/todos/pending/2026-10-12-run-parameterised-main-reports-operational-failure.md)).
- The example sources are not claimed to be format-canonical.

## Resolution rules

Resolution is closed. There is no search path and no environment lookup: an
import resolves only to a file under the entry's directory. A path segment that
is `..`, an absolute path and a symlink are all refused.

## Non-claims

- This is whole-program composition: the entry and its imports are merged into
  one program and checked and emitted together. It is not separate compilation,
  and there are no per-module interface artifacts.
- Build identity, the receipt and the evidence report hash every loaded module
  file (sorted `modules` rows), not only the entry file.
- Imported foreign declarations are not supported, so the adapter block is
  repeated in each entry.
- There is no `schway run --engine` path input; files are read through
  `build --manifest` plus `app run`.
- The provisional `examples/checksum.schway` frontier was retired on 2026-10-12;
  see the PROJECT.md Phase 32 amendment. The guide below is executed by a test,
  and hosted receipts are recorded by the Phase 32 receipt plan.
