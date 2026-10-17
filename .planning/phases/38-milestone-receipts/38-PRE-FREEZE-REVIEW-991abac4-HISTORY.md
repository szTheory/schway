---
phase: 38-milestone-receipts
reviewed: 2026-10-17T00:22:56Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - .github/workflows/ci.yml
  - scripts/verify-phase32.sh
  - scripts/check-milestone-binding.sh
  - testdata/receipts/M007.manifest
  - internal/compiler/session/session_phase38_receipts_test.go
source_sha: 991abac4d1c9ac63c7624192a956b00bf3cbef3b
codedigest: c6b626c73f6fc11d912ef5e309065df64d30158013430b821e82de3ada5992d0
source_hashes:
  .github/workflows/ci.yml: a8a913b4adfa36970bfd953e44ef541460acb302a9bfb78c75ddb05dc5b12c3b
  scripts/verify-phase32.sh: bffa9aeff47c7b9ba8393d73944f4203fabd926631e089f052f9165940e1eece
  scripts/check-milestone-binding.sh: 11124f3a575bbfb0285d1b16fcebb9e74bef246ac71bce2c31d376cadc514ab0
  testdata/receipts/M007.manifest: 7fcfeeab05e89d9253006e2a333ac9b7e670fb2412d57a55a252b0041d515985
  internal/compiler/session/session_phase38_receipts_test.go: dcf810732d0f211da933ccdecda17862a510139aa3356fa2f4b508061465bdbc
findings:
  critical: 1
  warning: 0
  info: 0
  total: 1
status: issues_found
---
<!-- phase38-preparation/1 kind=review source_sha=991abac4d1c9ac63c7624192a956b00bf3cbef3b codedigest=c6b626c73f6fc11d912ef5e309065df64d30158013430b821e82de3ada5992d0 disposition=blocked -->

# Phase 38: Pre-freeze Code Review

**Reviewed:** 2026-10-17T00:22:56Z
**Depth:** standard
**Files Reviewed:** 5
**Status:** issues_found

## Summary

This incremental review covers `991abac4d1c9ac63c7624192a956b00bf3cbef3b` against `efff39f98bf9f7ac9aba6dc5f514d1b49acfc5a6`. Only the binding checker and its receipt-control tests changed; the other three scoped files match their prior hashes. The two previously reported final-mode bypasses are closed: committed config changes after the freeze now refuse, and merge-aware history checks reject merge-only source and non-planning paths. One config-preservation gap remains at the freeze commit boundary.

**Newly run:** `GOCACHE=/tmp/schway-phase38-pre-freeze-rereview-gocache go test ./internal/compiler/session -run '^(TestPhase38FreezeWorktreeControls|TestPhase38FinalCandidateControls)$' -count=1 -v -timeout=120s` passed, including the merge-only digest-path, merge-only README, and config commit/restore controls. The supplied source SHA, codedigest, and all five source hashes match the checkout. The executor's four-root post-commit union was reported separately as historical evidence; it was not rerun here.

## Narrative Findings (AI reviewer)

### Critical Issues

#### CR-01: [BLOCKER] Freeze commit can include the protected workspace config

**File:** `scripts/check-milestone-binding.sh:465-467`
**Severity:** BLOCKER
**Issue:** The config history scan starts at `freeze_sha` and therefore examines only commits after it. If `.planning/config.json` was accidentally staged into the freeze commit itself, the scan is empty at that boundary. The worktree hash can then match `workspace_config.sha256` and the index is clean, so both freeze and final modes can accept a freeze commit that persisted the local workspace config. This violates the requirement that the pre-existing config remain unchanged and unstaged.
**Fix:** Inspect the freeze commit's own changed paths against its parent and refuse if it touches `.planning/config.json`; add a control where the freeze commit records the workspace config and assert freeze/final refusal.

---

_Reviewed: 2026-10-17T00:22:56Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
