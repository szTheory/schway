# Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md. This log keeps the alternatives that were considered.

**Date:** 2026-10-13
**Phase:** 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
**Areas discussed:** M006 receipt reopen, Generic freeze/receipt gate, isAbortLikeDefect narrowing, Small debt + bounded read + Phase 28 backfill
**Mode:** Advisor (calibration `minimal_decisive`). Four parallel gsd-advisor-researcher agents used multi-lens and adversarial framing. The user's standing instruction was to fan out, synthesize, and auto-follow, so they made no per-area picks.

---

## M006 receipt reopen

| Option | Description | Selected |
|--------|-------------|----------|
| Docs flip first, as a docs-only commit | The markers flip to "reopened" before any digest-path commit, so CI stays green | ✓ |
| Flip in the last commit | Every commit in between fails `stale_bound_claim` | |
| Hosted re-take in 33 | The evidence goes stale by Phase 34, and the roadmap says the phase binds no receipt | |

**User's choice:** Auto-followed the recommendation.
**Notes:** Archived receipt artifacts stay immutable.

## Generic freeze/receipt gate

| Option | Description | Selected |
|--------|-------------|----------|
| Two POSIX sh scripts plus a pre-freeze manifest in testdata, with freeze commit and digest as CLI args/env | Byte parity with the M006 output, no deps | ✓ |
| Go tool/helper | Typed, but adds a new binary to the digest and breaks sh byte parity | |
| M006 manifest under .planning receipts dir | Outside the digest. Rejected in favor of a uniform testdata location plus a sha256 pin | |

**User's choice:** Auto-followed the recommendation.
**Notes:** git `ls-tree` rejects `:(exclude)`, so IN-03 is solved by making `codedigest.sh` the single source of the path rules (`--paths`, `--changed`).

## isAbortLikeDefect narrowing

| Option | Description | Selected |
|--------|-------------|----------|
| Owner identity plus a closed per-kind reason, using shared constants in `execution` | Meets SC2 and removes the three-way string copy | ✓ |
| Owner check only | Does not meet "the reasons it names" | |
| Non-owner live resource at a defect: refuse | Interpreter guard against leak bugs, in a state that should be unreachable | ✓ |
| Non-owner live resource at a defect: accept | Native parity, but the guard becomes a no-op | |

**User's choice:** Auto-followed the recommendation.

## Small debt, bounded read, Phase 28 backfill

| Option | Description | Selected |
|--------|-------------|----------|
| Fix each cheap item with a named test; record covered items with their existing witness test | Honest, about 6 files | ✓ |
| Disposition-only | Launders debt | |
| Phase 28 docs hand-authored from evidence | Stays in scope, marked retroactive | ✓ |
| Run /gsd-validate-phase and /gsd-secure-phase retroactively | Writes new tests into frozen history and breaks the plan cap | |

**User's choice:** Auto-followed the recommendation.
**Notes:** The QLT-13 mutant is caught by a sparse 1 GiB file plus a TotalAlloc bound. There is no Stat-by-path pre-check (TOCTOU).

## Claude's Discretion

- Marker date and prose wording, reason-constant names, manifest line syntax, golden normalization token, and where the `.md`-only test lives.

## Deferred Ideas

- Migrating the older `check-phase27-receipts.sh` and similar gates.
- Moving test seams out of production packages (29 WR-04).
- Module-awareness for `evidence`/`interface`/`query`.
