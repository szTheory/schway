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
mkdir -p .schway/phase27
./schway build examples/fizzbuzz.schway --output ./.schway/phase27/phase27-fizzbuzz
./schway build examples/sum_to_n.schway --output ./.schway/phase27/phase27-sum
./schway app run ./.schway/phase27/phase27-fizzbuzz -- 15 > ./.schway/phase27/fizzbuzz-15.actual
cmp ./.schway/phase27/fizzbuzz-15.actual examples/phase27/fizzbuzz-15.out
./schway app run ./.schway/phase27/phase27-fizzbuzz -- 1000 > ./.schway/phase27/fizzbuzz-1000.actual
cmp ./.schway/phase27/fizzbuzz-1000.actual examples/phase27/fizzbuzz-1000.out
./schway app run ./.schway/phase27/phase27-sum -- 10
```

Every build artifact, its `.schway-build.json` sidecar and every `.actual` file
land under the ignored `.schway/phase27/`, so a later
`scripts/verify-phase27.sh` receipt run sees a clean tree; `mkdir -p` ensures
the redirect target directory exists.

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
uploads each host's log as `phase27-receipt-<os>-<sha>`. The gate binds receipts
to one exact revision, so there are two labeled uses.

### Fresh CI artifacts at their commit

CI uploads `phase27-receipt-macos-latest-<sha>` and
`phase27-receipt-ubuntu-latest-<sha>`; artifacts expire after 14 days. Check
out that `<sha>`, download both with
`gh run download <run-id> -n phase27-receipt-macos-latest-<sha>` (and the ubuntu
one), save each host's whole downloaded log as `host-macos.receipt` and
`host-linux.receipt`, and run the bare command:

```sh
sh scripts/check-phase27-receipts.sh host-macos.receipt host-linux.receipt
```

At a checkout of a different commit the bare command fails with
`wrong_revision` and prints a hint naming both revisions.

### Re-checking the committed Phase 27 receipts from a later checkout

```sh
PHASE27_EXPECTED_REVISION=6224f8b2c9c15f54270a232fab2bce12628c7f19 sh scripts/check-phase27-receipts.sh .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/host-macos.receipt .planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/host-linux.receipt
```

This attests the historical revision 6224f8b, not the current HEAD. It passes
only while the four declared inputs (two sources, two fixtures) are unchanged;
otherwise it fails with `source_hash_mismatch` or `fixture_hash_mismatch`,
correctly. The gate checks consistency, not authenticity; the authenticity
anchor is the hosted run record in 27-HOST-EVIDENCE.md. The script changes to
the repository root, so relative log paths resolve from the root (run from the
root or pass absolute paths).

The gate requires a clean tree, the same revision, and the same source/fixture
hashes and inputs on both hosts. A missing, failed, stale, different-input or
unavailable host is recorded as unavailable or failing, never as a pass. The
recorded run, revisions and conclusions are in
[`27-HOST-EVIDENCE.md`](../../.planning/milestones/M005-phases/27-exact-fizzbuzz-and-cross-host-evidence/27-HOST-EVIDENCE.md).

Historical Phase 25 and Phase 26 receipts are identified by their own revisions;
they are not Phase 27 execution evidence.
