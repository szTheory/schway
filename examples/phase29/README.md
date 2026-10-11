# Phase 29 fixture bytes owner

A test-only foreign owner that hands out fixed byte buffers keyed by a
`uint64_t` case id (D-05). It exists so the Phase 29 byte-view and Adler-32
vectors read real foreign-owned bytes. It is deliberately not the Phase 23
file-byte owner and not the future Phase 31 file-bytes owner (D-09); no symbol
shares a prefix with either.

Symbols (manifest `fixture_bytes.bindings.json`, schema `schway.local-c/1`):

- `schway_fixture_bytes_acquire(case_id)` returns a status and an owner. Any
  case id outside 0..9 returns status 1 with no owner published.
- `schway_fixture_bytes_view(owner)` returns a read-only data pointer and
  length; it is infallible and does not copy.
- `schway_fixture_bytes_release(owner)` frees the owner once.

An empty case allocates one byte, publishes length 0, and releases normally.

## Case table

Bytes are generated independently in C (`adapter.c`) and in Go
(`internal/compiler/testsupport/phase29_fixture_bytes.go`, D-07), and the two
are compared byte-for-byte by a test. Neither generator computes a checksum.
Expected values come from Python 3.14.4 `zlib.adler32` and Go 1.24.0
`hash/adler32`, never from a Schway implementation.

| Case | Bytes | Expected checksum |
|------|-------|-------------------|
| 0 | empty | 1 |
| 1 | one 0x00 | 65537 |
| 2 | "a" | 6422626 |
| 3 | "Wikipedia" | 300286872 |
| 4 | 4096 x 0x00 | 268435457 |
| 5 | 5552 x 0xFF | 4052720524 |
| 6 | 0x80 0xFF | 33620352 |
| 7 | 4096 x 0xFF | 2170679522 |
| 8 | 16384 x 0xFF | 2967061426 |
| 9 | 65536 x 0xFF | 2006388466 |

## Build and run the Adler-32 fixture

Run these commands from the repository root with Go 1.24 and Clang installed.
One artifact serves every case id; the case id is the single app argument.

```sh
go build -o ./schway ./cmd/schway
./schway build examples/phase29/adler32_fixture.schway \
  --manifest examples/phase29/fixture_bytes.bindings.json --output ./adler32-fixture
./schway app run ./adler32-fixture -- 3
./schway app run ./adler32-fixture -- 6
./schway app run ./adler32-fixture -- 10
```

Case 3 prints `300286872` and case 6 (`0x80 0xFF`, read as 128 and 255) prints
`33620352`, each with exit 0 and empty stderr. Case 10 is outside the table:
the acquire reports `FixtureMissing`, nothing is printed, and the exit code is
65. Every case 0..9 prints the checksum in the table above.

## Out-of-bounds control

`byte_index_oob.schway` is the fixture program with the callee replaced by
`read_past_end`, which reads `byte_at(bytes, n)` after `var n = byte_len(bytes)`.

```sh
./schway build examples/phase29/byte_index_oob.schway \
  --manifest examples/phase29/fixture_bytes.bindings.json --output ./byte-index-oob
./schway app run ./byte-index-oob -- 0
./schway app run ./byte-index-oob -- 2
```

Both runs print nothing on stdout, print exactly
`schway: byte index out of bounds` plus a newline on stderr, and exit 65. The
result is the same when the generated C is compiled at `-O3` and at
`-O3 -flto` by the test-only flag seam (D-21); the production build stays `-O0`.
The interpreter reports the same failure as a defect with empty output (D-19).

D-20: the byte-index defect is abort-like. It exits 65 without running the
release, so the fixture allocation is not freed by the program; the operating
system reclaims it when the process ends. For that reason the sanitizer lane for
this control runs with `ASAN_OPTIONS=detect_leaks=0`, while the ten checksum
vectors run with leak detection on wherever the host supports it.

## Sanitizer evidence

Each lane is a separate subtest with its own marker. Host for the rows below:
darwin/arm64, Apple clang version 21.0.0 (clang-2100.1.1.101). AddressSanitizer
is available here. LeakSanitizer is not supported on darwin, so the vector lane
here checks for no sanitizer report and no use-after-free or overflow, and
`detect_leaks=1` is exercised only on hosts that support it (the lane requests
`detect_leaks=0` on darwin).

| Lane | Subtest | Result |
|------|---------|--------|
| Adapter acquire/release of every case | `TestPhase29FixtureBytesUnknownCaseRefused/asan` | ran: no report (darwin/arm64, Apple clang 21.0.0) |
| Ten Adler-32 app vectors | `TestPhase29OOBApplicationLanes/asan/vectors` | ran: no report (darwin/arm64, Apple clang 21.0.0; no leak detection on darwin) |
| Out-of-bounds control | `TestPhase29OOBApplicationLanes/asan/oob-control` | ran: no report (darwin/arm64, Apple clang 21.0.0; `detect_leaks=0` per D-20) |

A host without AddressSanitizer records each lane as
`PARTIAL (address sanitizer unavailable on <GOOS>/<GOARCH>, <clang version>)`.
