# Phase 23 native file-byte application

`file_byte.lang` is a retained native application that receives one path token,
acquires a live `FileByteOwner`, borrows it to read one byte, and discharges the
owner before returning or reporting a typed use error. This example uses the
explicit local C binding manifest; it does not turn on general foreign calls
or pointer support.

## Build from a clean checkout

Run these commands from the repository root with Go 1.24 and Clang installed.
The `printf` commands create caller-created raw-byte files; they do not depend
on locale or text encoding.

```sh
go build -o ./lang ./cmd/lang
./lang build examples/phase23/file_byte.lang \
  --manifest examples/phase23/file_byte.bindings.json --output ./file-byte
printf '\101' > ./byte-41.bin
printf '\102' > ./byte-42.bin
./lang app run ./file-byte -- ./byte-41.bin
./lang app run ./file-byte -- ./byte-42.bin
```

The public app commands print `65` and `66`. The adapter accepts exactly one
raw file byte: it reads one byte and probes for EOF. An empty file fails during
acquisition, and a file with a second byte fails during acquisition before an
owner is published. `PathToken` has a 4,096-byte input bound.

```sh
: > ./empty.bin
printf '\101\102' > ./two-bytes.bin
printf '\103' > ./byte-43.bin
./lang app run ./file-byte -- ./empty.bin
./lang app run ./file-byte -- ./two-bytes.bin
./lang app run ./file-byte -- ./byte-43.bin
```

The independent expected answers are checked into
[`file_byte.expected.json`](./file_byte.expected.json). They name the public
`0x41 → 65` and `0x42 → 66` successes, the empty and two-byte acquisition
errors, and the `0x43` use error. These expected values and diagnostics are
authored constants, not values copied from compiler events. Acquisition errors
are `lang_file_byte_acquire: EmptyFile` and
`lang_file_byte_acquire: FileTooLong`. The `0x43` case reaches
`lang_file_byte_use: UnsupportedByte`, keeps its owner live through the failed
borrow, performs the matching release, and only then reports the error. CLI
diagnostics are bounded to 128 bytes.

## Evidence limits

The independent native observer records allocation, use, pointer identity, and
release for the exact binary and host exercised by that test. The public output
alone does not prove physical cleanup, and a receipt from one host does not
establish another host's ABI or runtime behavior. Interpreter receipts are
model-only: they do not perform host IO or prove physical cleanup. The
Phase 22 U64 application example remains available and still prints `7`.
