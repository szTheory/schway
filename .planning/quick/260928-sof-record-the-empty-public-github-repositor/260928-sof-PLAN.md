---
quick_task: 260928-sof
status: planned
depends_on: []
files_modified:
  - .planning/PROJECT.md
  - .planning/STATE.md
  - .planning/phases/23-live-local-allocation-and-discharge/.continue-here.md
---

# Record the empty public GitHub repository and no-PII boundary

## Goal

Make `github.com/szTheory/schway` the recorded canonical public repository path
without implying that the local checkout is connected or that source has been
published. Preserve the required privacy review as a gate before any future
public source push.

## Tasks

1. In `.planning/PROJECT.md`, record the 2026-09-28 repository decision:
   `github.com/szTheory/schway` is public and empty, with no default branch or
   uploaded README/source/files; its description is “Schway is a general-purpose
   programming language and toolchain for AI-authored, human-audited production
   software.” Update the coordinated technical rename boundary to use this
   chosen owner/path. State that this checkout still has no Git remote. Before
   any future public source push, require an audit of reachable Git history and
   tracked and untracked worktree files for personal information, secrets,
   credentials, and identifying local paths. Mark that audit as pending, with
   no conclusion about its outcome.
2. Update the current position and next-action wording in `.planning/STATE.md`
   and the Phase 23 `.continue-here.md` to reflect the chosen repository path
   and the same pending privacy gate. Keep the hosted Ubuntu
   `evidence-aggregate` receipt as Phase 23's outstanding verification blocker;
   preserve completed plans and UAT. Remove stale requests for the user to choose
   the owner/path, while keeping the local remote explicitly unset and the first
   source push outstanding.

## Scope boundary

This is documentation only. Do not add a remote, change GitHub settings, upload
files or source, push commits, run tests, or claim a completed PII audit. The
active product API coverage, assumption-delta, and schema-push pre-plan hooks do
not apply to repository administration and planning-document updates. The
active security contribution is the explicit no-PII boundary and pending audit.

## Verification

Use read-only `gh repo view szTheory/schway` and `git remote -v` to confirm the
repository and local-remote facts. Inspect the three edited documents and their
diff for consistent wording, the pending audit, and the preserved Phase 23 CI
blocker. No tests are needed for this documentation task.
