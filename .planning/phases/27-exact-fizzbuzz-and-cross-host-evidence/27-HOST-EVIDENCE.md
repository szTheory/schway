# Phase 27 Host Evidence

Status: EVD-12 satisfied by one hosted run at one declared revision.

## Claimed hosted run

| Field | Value |
|---|---|
| Claimed source revision | `fa5a169216b60cce94929c881001b1ff02fbf09a` (branch `phase26-validation`) |
| Hosted run | 38052601527, `workflow_dispatch` of `ci.yml`, default inputs |
| URL | https://github.com/szTheory/schway/actions/runs/38052601527 |
| Run conclusion | success |
| `current evidence aggregate (macos-latest)` | success (job 114214599385) |
| `current evidence aggregate (ubuntu-latest)` | success (job 114214599442) |
| `checks (macos-latest)` (full and race lanes) | success (job 114214599270) |
| `checks (ubuntu-latest)` (full and race lanes) | success (job 114214599562) |
| Skipped by design | Phase 26 checked scalar sum, validation corpus receipt (input flags off) |

## Per-host receipts

Rows are the verbatim `phase27 ` log lines from the downloaded artifacts
`phase27-receipt-macos-latest-<sha>` and `phase27-receipt-ubuntu-latest-<sha>`:

| Host | Saved file | Target | Toolchain | Cases | Controls | Result |
|---|---|---|---|---|---|---|
| macOS | `host-macos.receipt` | arm64, `arm64-apple-darwin25.6.0` | go1.24.13 darwin/arm64, Apple clang 21.0.0 | 10/10 | 3/3 | pass, tree=clean |
| Linux | `host-linux.receipt` | x86_64, `x86_64-pc-linux-gnu` | go1.24.13 linux/amd64, Ubuntu clang 18.1.3 | 10/10 | 3/3 | pass, tree=clean |

Both rows declare these inputs and hashes (git blob hashes):

- sources: `examples/fizzbuzz.schway=a274e64850b1727a85a34ee2cf0b5545b0b052ba`,
  `examples/sum_to_n.schway=958739bc74fb3cb3da00c5e5cf285172fce82aa9`
- fixtures: `examples/phase27/fizzbuzz-15.out=0a934f213a5145cb1ee6f760fda496bcde6d7130`,
  `examples/phase27/fizzbuzz-1000.out=917dcf12665f09c11d98fe21f893522a516d799b`
- inputs: `fizzbuzz:0,1,15,1000,1001;sum_to_n:0,1,15,1000,1001`

## Gate result

```sh
PHASE27_EXPECTED_REVISION=fa5a169216b60cce94929c881001b1ff02fbf09a \
  sh scripts/check-phase27-receipts.sh \
  .planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/host-macos.receipt \
  .planning/phases/27-exact-fizzbuzz-and-cross-host-evidence/host-linux.receipt
```

Output: `phase27 receipt gate: status=pass hosts=macOS,Linux revision=fa5a169216b60cce94929c881001b1ff02fbf09a` with the source and fixture hashes above.

The gate recomputes expected hashes from its own checkout. Commits made after
the claimed revision touch only `.planning/` evidence files, so the source and
fixture hashes are unchanged; the expected revision is therefore pinned with
`PHASE27_EXPECTED_REVISION` rather than the checkout HEAD.

## Superseded run

Run 38052011600 at `c7bf17c54591241392d15a4b2f0e4a64694ac212` passed the Phase 27
driver steps on both hosts but FAILED the hosted full-suite lanes on both hosts
(regressions from earlier Phase 27 plans: frozen legacy emitter bytes, the
canonical formatter for `let x = y % N`, and several guard inventories). It is
not claimed as evidence. The fixes were committed as `fa5a169` and a fresh run
was dispatched; every receipt above is from that fresh run.

## Clean-checkout guide execution (current execution)

Executed in fresh clones of this repository (not the working checkout), at
`c7bf17c` and again at `fa5a169`, following `examples/phase27/README.md`:
`go build -o ./schway ./cmd/schway`; `./schway build` of both examples;
`./schway app run` for FizzBuzz 15 and 1000 (byte-identical via `cmp` to the
committed fixtures) and sum_to_n 10 (`55`), 1000 (`500500`); and input 1001
for both programs (empty stdout, exit 65, `schway app: input exceeds 1000`).
These local runs were on Darwin arm64 and are guide verification, not
cross-host evidence.

## Scope

Phase 25 and Phase 26 receipts remain historical, identified by their own
revisions, and are not cited here as Phase 27 execution.
