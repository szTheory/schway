---
quick_task: 261001-ikw
status: complete
completed: 2026-10-01
subsystem: workspace-handoff
source_baseline: e5e8e245fc4a3229a8bc97adea387f5524ef6e1b
archive_head: 21bbd599d8b02309851d66db7a5d5502e4dd37f0
next_command: "$gsd-discuss-phase 25"
---

# Canonical Schway Checkout Restored

The user requested one clear development location. The original `schway`
checkout was renamed to sibling `schway-archive`; the current `schway-public`
checkout was then renamed to `schway`. Neither repository was overwritten,
reset, merged, or replaced with a partial copy.

## Preservation checks

- Both directories retain their original device/inode identities after the
  rename. Their pre/post-move heads, branches, remotes, and working-tree states
  match exactly.
- The archive remains at `21bbd59`, including its existing modified, deleted,
  untracked, and ignored local material. It still has no remote.
- All six private audit files retain their exact SHA-256 digests. They remain
  in the archive, outside the active repository.
- The current checkout retained `e5e8e24` and branch
  `worktree-agent-p24-01-retry` before this task's additive documentation
  commit. The prior local handoff commit is preserved.
- The former `schway-public` directory is absent. The private audit helper's
  original-source and report paths now target the archive; its previous script
  version and the move verification manifest were retained privately.

## Active handoff

Updated STATE and the Phase 25 `continue.md` to explain the canonical directory
and the archive boundary. Fresh `init.progress` at the canonical root confirms
Phases 22, 23, and 24 are complete, all 13 plans have summaries, and their
verification reports pass. Phase 25 has no CONTEXT or PLAN yet.

Start the next session in the canonical `schway` project folder and run:

```text
$gsd-discuss-phase 25
```

The latest completed feature phase remains **24 — Ownership Transfer Through
Calls and Errors**. Its decisions, verification, and hosted evidence remain
unchanged. This directory cleanup does not start Phase 25 or repeat prior work.

## Verification boundary

Directory/Git-state comparisons, private audit hashes, fresh structured GSD
routing, handoff references, and whitespace checks cover this move. No source,
runtime, CI configuration, or covered verification input changed. No local
project suite, push, or branch merge was performed. The directory move itself
has no source commit; the final task commit contains only planning records.
