---
phase: 22-native-application-build-and-single-execution
reviewed: 2026-09-30T14:58:32Z
depth: standard
files_reviewed: 5
files_reviewed_list:
  - cmd/schway/main.go
  - internal/compiler/native/native_app.go
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - scripts/verify-phase23.sh
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 22: Code Review Report

**Reviewed:** 2026-09-30T14:58:32Z  
**Depth:** standard  
**Files Reviewed:** 5 of 443 manifest paths  
**Status:** issues_found

## Summary

The manifest contains 443 changed paths. I inspected the five files listed in the frontmatter, focusing on the renamed CLI's public usage message, retained application execution, Phase 23's C adapter, and the Phase 23 verification receipt. This is a partial review of the manifest scope; the other changed paths were not reviewed, so no clean-scope conclusion is made. No tests were run.

## Warnings

### WR-01: Usage errors advertise the obsolete executable name

**File:** `cmd/schway/main.go:664`  
**Issue:** The public executable is now `schway`, but the usage diagnostic begins `usage: lang`. Users who invoke `schway` incorrectly receive instructions for an obsolete binary name, making the public command contract misleading.
**Fix:** Change the usage string prefix to `usage: schway` and update any contract assertion that pins the old public executable name.

### WR-02: Phase 23 receipt can report a clean tree with untracked source

**File:** `scripts/verify-phase23.sh:34-38`  
**Issue:** `git diff --quiet` and `git diff --cached --quiet` do not detect untracked files. If an untracked Go source or test file affects the verification run, the script records `tree=clean` even though the checked source tree differs from the committed revision. That makes the host receipt's cleanliness claim inaccurate.
**Fix:** Include untracked files in the check, for example by testing `git ls-files --others --exclude-standard` is empty along with both diff checks before assigning `source_tree=clean`.

---

_Reviewed: 2026-09-30T14:58:32Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_
