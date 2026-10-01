# Phase 24 owner transfer and typed-error applications

This example contains two ordinary native application entries:

- [`transfer.schway`](./transfer.schway) acquires a `FileByteOwner` in a helper,
  returns that live owner to its caller, borrows it to read one byte, and then
  releases it.
- [`error.schway`](./error.schway) performs two caller acquisitions and a third
  acquisition through `probe`. Its `0x43` input reaches a typed
  `UseError.UnsupportedByte` after all three owners exist.

Both entries receive the same one-path input and use the explicit Phase 23
local-C binding contract. Build and run them from the repository root with Go
1.24 and Clang installed:

```sh
go build -o ./schway ./cmd/schway
./schway build examples/phase24/transfer.schway \
  --manifest examples/phase23/file_byte.bindings.json --output ./phase24-transfer
./schway build examples/phase24/error.schway \
  --manifest examples/phase23/file_byte.bindings.json --output ./phase24-error
printf '\101' > ./byte-41.bin
printf '\102' > ./byte-42.bin
printf '\103' > ./byte-43.bin
./schway app run ./phase24-transfer -- ./byte-41.bin
./schway app run ./phase24-transfer -- ./byte-42.bin
./schway app run ./phase24-error -- ./byte-41.bin
./schway app run ./phase24-error -- ./byte-42.bin
./schway app run ./phase24-error -- ./byte-43.bin
```

The independently fixed public answers are `0x41 → 65` and `0x42 → 66` for
both entries. The error entry reports `UseError.UnsupportedByte` on `0x43` and
exits with status 65. The adapter accepts one raw byte and probes for EOF;
`PathToken` remains bounded to 4,096 bytes.

## Evidence scope

The Phase 24 native observer test links the generated error application to an
instrumented copy of the declared adapter. It privately matches actual C
pointers across allocation, the successful return from the owner-producing
helper called by `probe`, use, and release, then writes only the static
acquisition operation ID and its dynamic activation number. The passing `0x43`
receipt records allocations A, B, C; helper return and use of transferred owner
C; frees C, B, A; and zero outstanding allocations before the typed error is
reported. The observer never prints pointer addresses. Five reached controls
reject an omitted, premature, duplicate, or wrong-resource release and a semantic
identity collision. The generated Phase 24 error application has no serialized
compiler event stream; model events remain a separate semantic view, and no
compiler event is treated as physical-cleanup evidence.

The deterministic interpreter replay supplies acquisition and use outcomes.
It fixes `65`, `66`, and `UseError.UnsupportedByte` independently. It does not
claim actual host IO or physical cleanup, reports `actual_host_io=false` and
`physical_cleanup=false`, and does not open the path.
The focused hosted receipt from `scripts/verify-phase24.sh` records its host,
target, source revision, tree state, and elapsed time. New Ubuntu and macOS
Phase 24 receipts are kept distinct from the historical Phase 23 receipts.

These guarantees cover only the admitted `PathToken`/`FileByteOwner` programs,
the declared C ABI, infallible local release, and the supported success and
typed-error exits. They do not establish general file IO, arbitrary foreign
calls, pointer safety, unwinding, or physical cleanup from model output alone.
