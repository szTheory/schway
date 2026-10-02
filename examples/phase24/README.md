# Phase 24–25 file-byte utility

[`transfer.schway`](./transfer.schway) acquires and transfers one
`FileByteOwner`, reads a byte as `U64`, then passes that scalar through separate
shared and exclusive read/copy helpers before returning it. The helpers borrow
the copied scalar value; neither points into the owner's allocation. The C
adapter continues to own its file descriptor, and the declared local release
discharges the buffer allocation.

## Clean-checkout run

From the repository root, with Go 1.24 and Clang installed, build the CLI and
the retained application using the explicit Phase 23 C-binding manifest:

```sh
go build -o ./schway ./cmd/schway
./schway build examples/phase24/transfer.schway \
  --manifest examples/phase23/file_byte.bindings.json \
  --output ./phase25-transfer
printf '\101' > ./byte-41.bin
printf '\102' > ./byte-42.bin
printf '\103' > ./byte-43.bin
./schway app run ./phase25-transfer -- ./byte-41.bin
./schway app run ./phase25-transfer -- ./byte-42.bin
```

Inputs `0x41` and `0x42` print `65` and `66`, respectively. The `0x43` input
follows the existing typed use failure before either infallible copy helper
and exits with status 65:

```sh
set +e
./schway app run ./phase25-transfer -- ./byte-43.bin
result=$?
set -e
test "$result" -eq 65
```

Its stderr is `UseError.UnsupportedByte`. The native test instruments each
helper entry and confirms that the success cases call both helpers in order
while the typed-error case calls neither.

## Evidence index

- The trusted foreign operations and local build authority are declared in
  [`file_byte.bindings.json`](../phase23/file_byte.bindings.json), with their
  adapter in [`adapter.c`](../phase23/adapter.c) and
  [`adapter.h`](../phase23/adapter.h).
- [`phase25_utility_test.go`](../../internal/compiler/native/phase25_utility_test.go)
  builds and runs the integrated native utility, reached wrong-result controls
  for both families, and separate family conflict/escape controls.
- [`cgen_pointer_successor_test.go`](../../internal/compiler/cgen/cgen_pointer_successor_test.go)
  checks shared `const uint64_t *` and exclusive `uint64_t *` emission and each
  family's pointer manifest. Those shapes carry no `restrict`, `noalias`,
  capture, alignment, or ownership promise.
- [`session_pointer_successor_test.go`](../../internal/compiler/session/session_pointer_successor_test.go)
  and the Phase 25 checker/core/origin/path tests cover independent conflict,
  escape, and fail-closed boundaries.
- [`scripts/verify-phase25.sh`](../../scripts/verify-phase25.sh) prints
  separately indexed foreign/shared/exclusive receipts for macOS and Linux at
  baseline `-O0`, optimized `-O2`, and ASan+UBSan. A missing host or lane is
  incomplete; CI runs this script once in its existing dual-host evidence
  aggregate. The [CI workflow](../../.github/workflows/ci.yml) links the job
  that owns those host receipts; the focused records are its run logs, not the
  historical Phase 23/24 receipts.

## Phase 24 transfer and cleanup evidence

[`examples/phase24/transfer.schway`](./transfer.schway) and
[`examples/phase24/error.schway`](./error.schway) retain the bounded Phase 24
owner-transfer examples. The [native observer](../../internal/compiler/native/phase24_observer_test.go)
checks actual host IO and physical cleanup from generated C, including matching
allocation/use/release on the typed-error path. The [hosted receipt for run 36856048690](../../.planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md)
records the two-host validation of that observer. Deterministic interpreter replay remains model-only:
it does not claim actual host IO or physical cleanup. Historical Phase 23
receipts remain separately identified and do not substitute for Phase 24 or
Phase 25 evidence.

The model-only interpreter replay remains separate: it fixes semantic results
and typed-error ordering but does not claim host IO, native pointer behavior,
or physical cleanup. The Phase 24 owner observer proves the allocation's
bounded lifecycle; it is not evidence for either scalar pointer family.
