---
quick_task: 261001-ikw
status: complete
files_modified:
  - .planning/STATE.md
  - .planning/phases/25-separate-pointer-successors-and-integrated-utility/continue.md
  - .planning/quick/261001-ikw-archive-the-original-checkout-and-restor/
---

# Restore the canonical Schway checkout location

## Goal

The `schway` project folder contains the current sanitized source and Phase 25
handoff. The original private checkout remains recoverable in the sibling
`schway-archive` directory.

## Tasks

1. Confirm the archive destination is absent, both checkouts are independent
   repositories, and their expected heads and working-tree states still match.
   Record a private pre-move manifest, including audit-file digests. Rename the
   original `schway` directory to `schway-archive`, then rename `schway-public`
   to `schway`. If the second rename fails, restore the first directory name.
   Do not overwrite, delete, reset, merge, or copy source between repositories.
2. Verify both directory identities, heads, branches, working-tree states, and
   archived private audit digests after relocation. Update the private audit
   helper's original-checkout and report destinations to the archive so it
   cannot place private reports into the active tree.
3. Update active STATE and the Phase 25 handoff for the canonical location,
   record this quick task, and commit only these planning documents. Verify
   the source remains unchanged and structured GSD routing still reports
   Phases 22–24 complete and Phase 25 ready for `$gsd-discuss-phase 25`.

## Constraints and verification

- Preserve original head `21bbd59`, all dirty/untracked material, and private
  audit records in the archive. Preserve current head `e5e8e24` and its branch
  `worktree-agent-p24-01-retry` before the additive handoff commit.
- Keep publication-facing documents free of personal absolute paths. Reference
  the private archive by its sibling directory name; do not copy its contents.
- Require matching pre/post-move directory identities, Git state, and private
  audit hashes; no `schway-public` directory remains afterward.
- Verify all 13 prior plans are summarized, prior verification remains passed,
  Phase 25 has no CONTEXT or PLAN, and handoff references resolve.
- Run whitespace and documentation consistency checks. No source change,
  local project test suite, push, or branch merge is part of this task.
