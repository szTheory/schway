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

## Out-of-bounds control

D-20: a byte-index defect exits 65 without running release, and the operating
system reclaims the allocation. The commands that demonstrate this are added
by plan 29-10.
