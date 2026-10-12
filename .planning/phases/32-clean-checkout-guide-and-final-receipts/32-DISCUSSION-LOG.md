# Phase 32: Clean-Checkout Guide and Final Receipts - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-12
**Phase:** 32-clean-checkout-guide-and-final-receipts

**Areas discussed:**
- guide home and doc-test;
- frontier retirement;
- receipt revision binding;
- the pending todos.

**Mode:** advisor fan-out, with one `gsd-advisor-researcher` per area at the `minimal_decisive` tier. Every recommendation was auto-followed under the user's standing fan-out boilerplate, which was re-pasted this session. No AskUserQuestion was issued. A cross-area coherence pass made two adjustments, recorded as [synthesis] in CONTEXT.md:
- the code digest excludes `*.md`;
- all code changes, including the formatter fix, land before the receipt freeze.

---

## Guide home and doc-test

| Option | Description | Selected |
|--------|-------------|----------|
| README as the executable spec | Parse the marked fences and results table in `examples/checksum/README.md` and run them in a tracked-file temp copy | ✓ |
| New top-level guide | Same parser, new doc, so claims are duplicated | |
| Test-side mirror table plus drift check | Can pass while the guide is wrong | |
| `testscript` / txtar | A dependency, and it inverts the doc-test | |

**Choice:** The README is the spec. The clean checkout is a `git ls-files -co --exclude-standard` copy. Only deterministic error rows go in the guide, with anti-vacuity checks and mutated-guide controls.

## Frontier retirement

| Option | Description | Selected |
|--------|-------------|----------|
| Delete, with a tombstone digest, dated allow-lists and an amendment | Frozen goldens stay byte-identical; a resurrected file fails | ✓ |
| Move to `testdata/` or a history directory | Stays a live refusing corpus member | |
| Keep in place with a "retired" header | Re-pins anyway and leaves a duplicate | |

**Choice:** Delete, with `TestPhase32ChecksumFrontierRetired`, the `phase28RetiredEntries` and `phase30RetiredPaths` lists, and the amendment in PROJECT.md.

## Receipt revision binding

| Option | Description | Selected |
|--------|-------------|----------|
| Code-tree digest in the hosted log, plus a committed receipt | Content-addressed; not affected by its own commit; survives shallow clones and squash merges | ✓ |
| `headSha` plus `git diff` | Breaks when history is missing | |
| `rev-parse HEAD:<subtree>` | Fiddly to combine | |
| BuildID / artifact digests only | Differ per host and toolchain | |

**Choice:** Extend `verify-phase27.sh` and `check-phase27-receipts.sh` with a `codedigest`. The receipt is `32-RECEIPT.json`. The public-route audit runs first. Reopening is a documentation-truth gate that never fails on drift alone. The plans split around the single push and dispatch hand-off.

## Pending todos

| Option | Description | Selected |
|--------|-------------|----------|
| Formatter: fix only, first plan of the phase | Lands before the receipt freeze; committed sources are not reformatted | ✓ ([synthesis]: folded rather than `/gsd-quick`) |
| Formatter: fix and reformat the adler32 sources | Re-pins about 48 sites | |
| `run` on a parameterised main: keep deferred, guide avoids it | Zero code risk; the guide names it as unsupported | ✓ |
| `run`: fold or quick-fix | Re-pins the DX-17 manifest; it is a design decision | |

## Claude's Discretion

- The fence-marker syntax and table layout.
- How a missing clang is made a failure on CI.
- Test file placement.
- Whether to generalize the phase27 scripts in place or add phase32 scripts.

## Deferred Ideas

- `run` with an INPUT argument.
- Reformatting the adler32 sources.
- Authenticated (signed) receipts.
