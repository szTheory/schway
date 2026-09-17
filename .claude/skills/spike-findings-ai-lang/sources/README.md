# Spike sources

Point-in-time snapshots of the spike workbenches, copied verbatim so this skill
stays self-contained if the spikes are ever pruned.

**Canonical originals:** `.planning/spikes/NNN-*/` in this same repository.
**Snapshot taken at:** commit `d919e2f`, 2026-09-17.

These copies are frozen evidence, not a second implementation to maintain. If a
spike is ever re-run and its README changes, this snapshot goes stale silently —
diff against the canonical path before quoting a number from here:

```sh
diff -ru .planning/spikes/006-interprocedural-liveness-cost-scaling \
         .claude/skills/spike-findings-ai-lang/sources/006-interprocedural-liveness-cost-scaling
```

Each snapshot is its own Go module (`go.mod` at the spike root), so nothing here
is part of `github.com/codename-lang/lang` and `go test ./...` at the repo root
does not reach it.
