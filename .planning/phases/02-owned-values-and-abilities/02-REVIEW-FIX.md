---
phase: 02-owned-values-and-abilities
fixed_at: 2026-09-04
review_path: .planning/phases/02-owned-values-and-abilities/02-REVIEW.md
iteration: 7
findings_in_scope: 12
fixed: 8
accepted_as_debt: 4
skipped: 0
status: closed_with_accepted_debt
code_head: 3399ddc
---

# Phase 02: Fix Report — Waves 6 and 7

**Iterations 1–5** are recorded in git history and in the prior revision of this
file (post-gate closure through `cbba405`). This revision covers the two waves run
during Phase 02 close-out, after an independent deep review overturned the previous
`clean` verdict.

## Wave 6 — the blockers the `clean` verdict missed (9 commits)

An independent deep re-review at `f1206bb` found **2 critical, 4 warning, 3 info**,
overturning the prior `status: clean`. All three headline findings were reproduced
through the shipped CLI before any edit.

| ID | Finding | Commit(s) |
|---|---|---|
| CR-02 | `lang format` emitted non-reparsable source at exit 0 for a linear body with a generic return type plus a binding — identifiers fused (`take vww`), function left unterminated | `890bcd0` |
| WR-03 | The property generator emitted only Phase 1 match programs, so the round-trip property could never reach the linear surface — this is *why* CR-02 shipped | `2d98a78` |
| WR-01 | `evidence.Build` discarded the canonical reparse diagnostics; without the guard it bound a **complete manifest to corrupted bytes** | `d9b370f` |
| WR-04 | `testsupport` spawned processes with `CombinedOutput`, unbounded buffers, no deadline | `2f2e2e2` |
| SEC-02-C | The `CombinedOutput` assertion scanned only `native.go`, so `evidence.go`'s two spawns were unguarded | `8d87a1a` |
| SEC-02-A / IN-01 | cgen reserved sets omitted emitted identifiers; the load-bearing prefix-confinement invariant was unstated and untested | `dd9c0a8` |
| CR-01 | `check.go` had no ability gate on `borrow` while `corevalidate` did, so an admitted program died as opaque `exit 3 / [0:0]` instead of a causal diagnostic | `2549311` |
| WR-02 | Loan liveness was not transitive — one reborrow hop let an illegal move through — **and the "independent" oracle encoded the same wrong law** | `3d9493a` |
| — | The ninth control was required fail-closed in `session.go` but not asserted at the CLI layer | `da75e95` |

CR-01 required a semantic decision (the plans contradicted themselves about whether
`Buffer` is shareable). The developer chose to give `Buffer` the `share` ability;
recorded as **OV-02-01** in `02-OVERRIDES.md`.

## Wave 7 — the regression wave 6 introduced (1 commit)

A second independent deep review at `da75e95` found that **`890bcd0` was a net
regression**. Reproduced first-hand before editing:

```
fn f(v: Byte) -> Byte // c
{ let w = take v
  w }
```

formatted to `take vw` — `v` and `w` fused — at **exit 0**, output failing to
reparse (`syntax.expected_linear_result`), and with two declarations the leaked
indent swallowed `fn g` into `f`'s body. Broader than the original CR-02: it hit
every parameter type, not only generics, and the pre-`890bcd0` classifier did not
have this failure.

**Root cause:** `890bcd0` classified the opening brace from a `header` field, but
`newline()` cleared it unconditionally and `TokenComment` routes through
`newline()`.

**Fixed in `3399ddc`:** preserve `header` across the comment-induced line break — a
comment is the one token that ends a line without ending the construct that opened
it — and emit the brace without a separator space when it opens its own line (this
also closes IN-04). Both generators now emit a comment trailing a declaration header
and a `match` header, closing the placement blind spot that let 1000 generated cases
and two fuzz seeds miss it.

**Falsification:** with generators updated and `format.go` alone reverted,
`TestGeneratedLinearRoundTrips` fails at `case=11239299577418040656` with
`borrow hold1hold2`. The coverage is load-bearing.

## Accepted as debt, not fixed

Four warnings from the wave-7 review plus the carried security and info items are
recorded in `02-DEBT.md` (D-02-01 … D-02-09) with reproduction, non-blocking
rationale, and a Phase 03 remedy each:

- **D-02-01** the repo-wide spawn guard is a substring scan and was defeated four ways
- **D-02-02** `evidence.canonical_unstable` never reaches a user
- **D-02-03** transitive loan sets made `check` Θ(N²)
- **D-02-04 … D-02-09** `native.timeout` falsifier, `__LANG_` C17 §7.1.3, the 8 MiB
  CLI ceiling, the literal-line causality seam, `protocol.Human` divergence, and
  `Box`/`Pair` spanless exit 3

The developer decided on 2026-09-04 to close Phase 02 recording these rather than
run an eighth wave: all are control strength, cost, or observability, and waves 5,
6 and 7 each introduced defects while fixing others.

## Verification at `3399ddc`

- `go test ./...` — exit 0
- `go test -race ./...` — exit 0
- `go vet ./...` — exit 0
- `sh scripts/verify-phase2.sh` — exit 0; Phase 1 work 18, Phase 2 work **52**,
  **nine** controls, `escape:coordinated-source-core-lie` only in `expected_escapes`
- `git diff 91a6206^ --exit-code` over `testdata/phase1/generated.golden.c`,
  `testdata/phase1/evidence.golden.json`, `testdata/phase2/owned_transfer.golden.c` —
  **empty**
- `testdata/phase2/evidence.golden.json` moved in `core_digest` and `id` only, as the
  causal consequence of OV-02-01
- `git diff --check` — clean

All three originally reported blockers re-tested through the shipped CLI: the
formatter round-trips and is a fixed point; `borrow buffer` checks and runs at exit
0; the reborrow-then-move program is rejected at exit 2 with a span, while the
direct form is still rejected (no over-blocking).

---

_Fixed: 2026-09-04_
_Iteration: 7 (close-out)_
