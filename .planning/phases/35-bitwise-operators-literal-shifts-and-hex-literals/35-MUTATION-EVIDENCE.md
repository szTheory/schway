# Phase 35 live transfer-mutation audit — 2026-10-14

This supplemental receipt closes the D-18 wrong-result implementation-mutant
gap. It is separate from the Phase 34 compatibility adaptation below and does
not change either requirement or frozen fixture. The audit ran on macOS arm64,
Go 1.24.0, from working tree based on `dac30e23de4596ab665771f0ef8db3801896c227`.

## Command and outcome

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35PeerTransferMutationControls$' -count=1 -v -timeout=600s
```

The named outer test passed. It ran three no-op source overlays (one per value
peer), each of which ran and passed every subtest in the peer's
`TestPhase35ScalarTransferMutants`. It then ran fifteen non-no-op source
overlays. Every child returned failure from its intended named behavioral
assertion. The harness also requires that the child output contains both the
exact subtest `RUN` and `FAIL` lines plus its expected sink assertion text;
compile errors, missing selectors, and surviving mutants fail the outer test.

Each mutant changes only temporary implementation-source copies through Go's
standard `-overlay`. The real transfer expression in the peer is replaced as
follows: XOR `^` → `&`, AND `&` → `|`, OR `|` → `^`, left shift `<<` → `>>`,
and right shift `>>` → `<<`. Exact source occurrence counts are asserted
before each child starts. No implementation file is written or changed by the
test.

| Value peer | Operator | Mutated transfer sites | Reached named behavioral assertion |
|---|---|---:|---|
| check | XOR | 1 | `TestPhase35ScalarTransferMutants/xor_high-bit_pattern` |
| check | AND | 1 | `TestPhase35ScalarTransferMutants/and_high-bits_to_all-zero` |
| check | OR | 1 | `TestPhase35ScalarTransferMutants/or_overlapping_high-bit` |
| check | left shift | 1 | `TestPhase35ScalarTransferMutants/left_shift_all-zero` |
| check | right shift | 1 | `TestPhase35ScalarTransferMutants/right_shift_high-bit` |
| corevalidate | XOR | 2 | `TestPhase35ScalarTransferMutants/xor_zero` |
| corevalidate | AND | 2 | `TestPhase35ScalarTransferMutants/and_zero` |
| corevalidate | OR | 2 | `TestPhase35ScalarTransferMutants/or_overlapping_bit` |
| corevalidate | left shift | 2 | `TestPhase35ScalarTransferMutants/shift_left_zero` |
| corevalidate | right shift | 2 | `TestPhase35ScalarTransferMutants/shift_right_one` |
| originvalidate | XOR | 1 | `TestPhase35ScalarTransferMutants/xor_zero` |
| originvalidate | AND | 1 | `TestPhase35ScalarTransferMutants/and_zero` |
| originvalidate | OR | 1 | `TestPhase35ScalarTransferMutants/or_overlapping_bit` |
| originvalidate | left shift | 1 | `TestPhase35ScalarTransferMutants/shift_left_zero` |
| originvalidate | right shift | 1 | `TestPhase35ScalarTransferMutants/shift_right_one` |

The independent sink inputs distinguish the real result from each changed
result. In particular, OR uses overlapping `1 | 1 == 1`, so OR-to-XOR changes
the expected result; right shift uses the high bit shifted by 63, and left
shift shifts that high bit out of the low 64 bits. Corevalidation mutates both
copies of each transfer law used by its distinct analysis routes. The earlier
operation/input substitution controls remain useful differential controls,
but are not counted as source-implementation mutants here.

The child log recomputes and records the SHA-256 digest of each live
implementation source on every invocation:

| Peer source | SHA-256 bound by this run |
|---|---|
| `internal/compiler/check/scalar.go` | `a2cacabc99cd3908cacdd11f50a3c29487576a8d3489a77d86f322c8cd6c0167` |
| `internal/compiler/corevalidate/corevalidate.go` | `9d6b957d5a4df55584e21a229ce0d51a23f5d1852fd3923b84540d4eb54be14d` |
| `internal/compiler/originvalidate/originvalidate.go` | `c1a42ed4c6af8d3431ddcbd0ef41677af7d44375597feb89a5dc2dac8171ed5b` |

These recorded digests bind this run's implementation snapshots. The test
recomputes the digests and exact transfer-site counts on future runs; this
receipt does not claim that these fixed hash strings are a recurring CI gate.

## Separate Phase 34 compatibility adaptation

The immutable `testdata/phase34/prechange-corpus.json` remains unchanged. Its
162 historical rows retain their admission and emission-outcome checks. The
current tracked-source inventory is 163 because Phase 35 adds the exact
`examples/bitwise.schway` witness named by the Phase 35 compatibility delta;
the Phase 34 test allows only that declared additive path.

The four historical C projections affected by D-20 now use an exact-row
adapter in `session_phase34_prechange_test.go`. It checks both the frozen
Phase 34 C length/SHA-256 and the current input-source and emitted-C
length/SHA-256 for every row. The direct Adler projection is bound to the
compatibility delta's current Adler source hash and current emitted-C hash.
The three imported application projections have their own exact source/output
bindings in the adapter; they are not accepted by path name alone.

| Phase 34 path | Emitter | Frozen C bytes/SHA-256 | Current source SHA-256 | Current C bytes/SHA-256 |
|---|---|---|---|---|
| `examples/checksum/adler32.schway` | `control_native_c` | 44595 / `d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b` | `16490db4e53cb9db287727a8e5918745aa412659e799dd2fd53b38805bb3bef1` | 19163 / `8653dbec998b2d39975619912b400e465b299d9d60ef59bdb1397ed9991d8e6e` |
| `examples/checksum/adler32_cli.schway` | `emit_application` | 19351 / `0bf9cda02e6806c1aeadd4912fabdb46a1856e70ae2a69ba0a9071db86ebc350` | `62478d7b2ab4e777f6adfeab1c9a4e79c230178d4e1c9ca9a41f8c45fad6aa7f` | 14911 / `80df93af9f946d5db3fc520c412f7de31cec28efa7c658ce9a54dd6080950719` |
| `examples/checksum/adler_low.schway` | `emit_application` | 21511 / `34dcb4f59f20e2234ddda81654da483c1d9f9e89a18b2567d43b5fec5b6af234` | `42a38540c3c282ab7089ee8f2ccc63c23d039879e19f9ddf5c6ffff9bcfe6987` | 17071 / `31b6410aa3ca9f43bda6db85b6aebd004d2abb30cea03df149be314753698495` |
| `examples/checksum/checksum.schway` | `emit_application` | 19598 / `9bb3babccd3d9a88b36bdedd148d0b548fe142128f2367e1a33f806d2f298751` | `da61026ce0f4a6296603e359244d526bef31db230d7339fe04608ae6e5216518` | 15158 / `baef335d68fca5df6ed61c63fb6b6248ebbae04a38cc83937cdc19c4a0cdac9e` |

Freshly rerun historical controls:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1 -v -timeout=600s
```

Both tests passed. The first checks the frozen inventory, admissions,
outcomes, and exact C comparisons/adapter. The second reaches a copied changed
C projection, requires attribution to the exact path/emitter C-byte mismatch,
and verifies clean restoration.

## Scope

The owner-specific checker, corevalidation, and originvalidation selectors
passed in this audit with their independent expected-result sinks and distinct
input-differential controls. The historical Phase 35-07 receipt remains the
source for the 33-vector O0/O3/O3-LTO matrix and all 16 UBSan shift vectors;
those lanes were not rerun during this audit. No hosted Phase 38 claim is made.
