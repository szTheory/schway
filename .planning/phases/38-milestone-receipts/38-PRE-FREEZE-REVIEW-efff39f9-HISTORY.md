---
phase: 38-milestone-receipts
reviewed: 2026-10-17T00:16:31Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - .github/workflows/ci.yml
  - scripts/verify-phase32.sh
  - scripts/check-milestone-binding.sh
  - testdata/receipts/M007.manifest
  - internal/compiler/session/session_phase38_receipts_test.go
source_sha: efff39f98bf9f7ac9aba6dc5f514d1b49acfc5a6
codedigest: 007ea4a600ab060ab2a04417a6ad3650032f5e92cf4b6ff4befd991fd2cb53a9
source_hashes:
  .github/workflows/ci.yml: a8a913b4adfa36970bfd953e44ef541460acb302a9bfb78c75ddb05dc5b12c3b
  scripts/verify-phase32.sh: bffa9aeff47c7b9ba8393d73944f4203fabd926631e089f052f9165940e1eece
  scripts/check-milestone-binding.sh: 2373789f6d2be373f50b42ae8b4cca8fb3250f1db67854cc29f86eb1f64a2959
  testdata/receipts/M007.manifest: 7fcfeeab05e89d9253006e2a333ac9b7e670fb2412d57a55a252b0041d515985
  internal/compiler/session/session_phase38_receipts_test.go: 46e59c60ca7d33867569826ef92dc7ea2797b160965f20084a150e0b741fc17c
findings:
  critical: 2
  warning: 0
  info: 0
  total: 2
status: issues_found
---
<!-- phase38-preparation/1 kind=review source_sha=efff39f98bf9f7ac9aba6dc5f514d1b49acfc5a6 codedigest=007ea4a600ab060ab2a04417a6ad3650032f5e92cf4b6ff4befd991fd2cb53a9 disposition=blocked -->

# Phase 38: Pre-freeze Code Review

**Reviewed:** 2026-10-17T00:16:31Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

This incremental review covers `efff39f98bf9f7ac9aba6dc5f514d1b49acfc5a6` against the prior reviewed source `0d5d51440e66e47c53f1703bef8b521eb78e389a`. Only `scripts/check-milestone-binding.sh` and `internal/compiler/session/session_phase38_receipts_test.go` changed; the other three scoped files match the prior review's hashes and its historical validation remains prior evidence. The focused final, freeze, and binding controls pass, but final-mode history checks still have two bypasses that can admit an unsafe candidate.

**Newly run:** `GOCACHE=/tmp/schway-phase38-pre-freeze-gocache go test ./internal/compiler/session -run '^(TestPhase38BindingControls|TestPhase38FreezeWorktreeControls|TestPhase38FinalCandidateControls)$' -count=1 -v -timeout=120s` passed. The supplied source SHA, codedigest, and all five source hashes match the checkout. No package-wide, full/race, producer, native, capture-pressure, or hosted checks were run.

## Narrative Findings (AI reviewer)

### Critical Issues

#### CR-01: [BLOCKER] Final mode can commit the protected workspace config

**File:** `scripts/check-milestone-binding.sh:473-475`
**Severity:** BLOCKER
**Issue:** The config guard checks the current worktree hash and whether the current index stages `.planning/config.json`. It does not inspect committed history for that path. Final mode accepts every `.planning/` path at lines 494-496, so a planning commit can record the workspace's dirty config bytes and a later commit can restore the worktree baseline; the current index is then clean and the expected hash matches, allowing final mode to pass. This can persist local workspace configuration in repository history despite the preservation rule.
**Fix:** In final mode, reject any commit after the freeze that touches `.planning/config.json` and keep the current staged-path check. Add a final-mode control that commits the workspace config and restores it before checking; it must refuse.

#### CR-02: [BLOCKER] Merge-only path changes evade the history scope check

**File:** `scripts/check-milestone-binding.sh:506-508`
**Severity:** BLOCKER
**Issue:** `git log --name-only` does not show merge-commit diffs by default. A merge resolution can introduce a path outside `.planning/` without a non-merge commit touching that path; if a later merge resolution removes it, the endpoint `git diff` is clean and both merge commits remain absent from this history list. Final mode can therefore pass despite non-planning changes in the ingest history, including a file that was committed and later removed.
**Fix:** Enumerate path changes against every commit parent (for example, use merge-aware `git log -m --name-only` or `git diff-tree -m` over the commit range) and add a merge-only non-planning create/remove control that must be rejected.

---

_Reviewed: 2026-10-17T00:16:31Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
