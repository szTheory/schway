# Phase 27 exact FizzBuzz and sum applications

This is the one command owner for the two M005 programs:
[`examples/fizzbuzz.schway`](../fizzbuzz.schway) (returns `Unit`, writes
explicitly) and [`examples/sum_to_n.schway`](../sum_to_n.schway) (returns `U64`,
auto-printed). Fixed FizzBuzz answers live in
[`fizzbuzz-15.out`](./fizzbuzz-15.out) and
[`fizzbuzz-1000.out`](./fizzbuzz-1000.out).

## Clean-checkout run

From the repository root, with Go 1.24 and Clang installed:

```sh
go build -o ./schway ./cmd/schway
./schway build examples/fizzbuzz.schway --output ./phase27-fizzbuzz
./schway build examples/sum_to_n.schway --output ./phase27-sum
./schway app run ./phase27-fizzbuzz -- 15 > ./fizzbuzz-15.actual
cmp ./fizzbuzz-15.actual examples/phase27/fizzbuzz-15.out
./schway app run ./phase27-fizzbuzz -- 1000 > ./fizzbuzz-1000.actual
cmp ./fizzbuzz-1000.actual examples/phase27/fizzbuzz-1000.out
./schway app run ./phase27-sum -- 10
```

## Exact answers

- FizzBuzz `0` prints nothing; `1` prints `1`; `15` and `1000` print the bytes in
  `examples/phase27/fizzbuzz-15.out` and `examples/phase27/fizzbuzz-1000.out`.
- sum_to_n `0` prints `0`, `10` prints `55`, `1000` prints `500500` (and `1`
  prints `1`, `15` prints `120`).

## Bounds and failure behavior

- Inputs `0` through `1000` are accepted. Input `1001` is refused before any
  output: empty stdout, exit status 65, stderr
  `schway app: input exceeds 1000`.
- Application output is bounded incrementally by a shared 65536-byte limit. A
  write that would cross the limit is a non-success: a prefix of completed
  writes may already have been streamed to stdout, the child exits with status
  65, and stderr is `schway: application output limit exceeded`. The `app run`
  command forwards that child status.
- A `Unit` entry (FizzBuzz) uses explicit `write_text`, `write_u64` and
  `write_newline` calls with literal text only. A `U64` return (sum_to_n) is
  auto-printed with a trailing newline.
- `U64` remainder by a statically known zero divisor is refused at check time
  (`type.scalar_remainder_zero`). A dynamic zero divisor fails at run time with
  exit status 65 and stderr `schway: U64 remainder by zero`.

## Focused controls and host verification

```sh
go test ./internal/compiler/session -run '^TestPhase27' -count=1
go test ./internal/compiler/cgen -run '^TestPhase27Output' -count=1
go test ./cmd/schway -run '^TestPhase27' -count=1
sh scripts/verify-phase27.sh
```

`scripts/verify-phase27.sh` runs the cases and native controls on the available
host only and prints a receipt with host, toolchain, revision, source and
fixture hashes, and inputs. It prints `status=pass` only after every case and
control finishes. A single-host run, especially from a modified tree, is a
development receipt, not cross-host evidence.

## Two-host evidence

CI runs the driver in both the Ubuntu and macOS `evidence-aggregate` lanes and
uploads each host's log as `phase27-receipt-<os>-<sha>`. Download the two
artifacts for one run, save one result row per host, and check them:

```sh
sh scripts/check-phase27-receipts.sh host-macos.receipt host-linux.receipt
```

The gate requires a clean tree, the same revision, and the same source/fixture
hashes and inputs on both hosts. A missing, failed, stale, different-input or
unavailable host is recorded as unavailable or failing, never as a pass. The
recorded run, revisions and conclusions are in
[`27-HOST-EVIDENCE.md`](../../.planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/27-HOST-EVIDENCE.md).

Historical Phase 25 and Phase 26 receipts are identified by their own revisions;
they are not Phase 27 execution evidence.
