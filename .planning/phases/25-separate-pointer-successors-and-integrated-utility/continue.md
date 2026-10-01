# Continue — Phase 25 discussion

## Last verified state

Phase 24, **Ownership Transfer Through Calls and Errors**, is complete:
three plans summarized, five of five roadmap truths verified. Hosted
[CI 36856048690](https://github.com/szTheory/schway/actions/runs/36856048690)
passes Ubuntu/macOS checks and evidence aggregates at `ed94ef79`. Current
implementation is unchanged from that tested source. Phases 22 and 23 also
have current passing verification. Public-repository preparation is complete.

## Next action

Run **`$gsd-discuss-phase 25`** in this checkout, on
`worktree-agent-p24-01-retry`. Phase 25 has no CONTEXT or plans yet. Capture its
remaining product decisions before planning; do not skip directly to execution.

## Scope and evidence to carry forward

Phase 25 is **Separate Pointer Successors and Integrated Utility**: one bounded
shared read-copy helper, one exclusive read-copy helper, and a reproducible
file-byte utility with a pinned result. Native by-pointer bodies remain
refused. Family-specific admission, independent peer checks, positive and
conflict/escape controls, and checked ABI facts must accompany the runnable gain.

Read `.planning/ROADMAP.md` and `.planning/REQUIREMENTS.md` for NAT-11/12/13,
EVD-10, and DX-14/15; then `.planning/PRODUCT-ROADMAP.md` and
`.planning/LANGUAGE-MATURITY.md` for current priorities. Carry forward the
Phase 24 directory's `24-CONTEXT.md`, `24-03-SUMMARY.md`, and
`24-VERIFICATION.md`. Start from `examples/phase24/transfer.schway` and the
independent physical-observer evidence, not model events alone.

## Avoid repeated or expanded work

- Do not repeat publication, Phase 22–24 execution, completed UAT, or passing
  verification. Fresh GSD routing confirms those phases complete.
- Run project suites through hosted CI; no local suite is authorized by the
  standing validation contract. No new runtime receipt is claimed here.
- Keep shared/exclusive witnesses separate and the FFI-03 trust boundary
  explicit. Add no unsupported alias, alignment, or capture promises.
- Arithmetic, loops, FizzBuzz, and general JSON remain subsequent work. Phase
  25 decisions are still to be discussed; this handoff does not settle them.
- This branch contains later work than public `main`; do not reset to main.
  Handoff-only local commits are safe to retain. Publish only after the
  required privacy scan; no push is part of this handoff.
