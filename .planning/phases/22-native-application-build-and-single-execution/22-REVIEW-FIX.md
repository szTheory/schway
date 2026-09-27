---
phase: 22-native-application-build-and-single-execution
fixed_at: 2026-09-27T20:22:44Z
review_path: .planning/phases/22-native-application-build-and-single-execution/22-REVIEW.md
iteration: 1
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 22: Code Review Fix Report

**Fixed at:** 2026-09-27T20:22:44Z  
**Source review:** `.planning/phases/22-native-application-build-and-single-execution/22-REVIEW.md`  
**Iteration:** 1

**Verification location:** Targeted checks ran in the isolated review-fix worktree (`.claude/worktrees/rf-22-9051-1790540472`), before teardown.

**Summary:**
- Findings in scope: 1
- Fixed: 1
- Skipped: 0

## Fixed Issues

### WR-01: Replay case decoder accepts duplicate JSON keys

**Files modified:** `internal/compiler/native/native.go`, `internal/compiler/session/session_app_verify.go`, `internal/compiler/session/session_app_verify_test.go`  
**Commit:** `340fd56`  
**Applied fix:** Exported the existing recursive native JSON duplicate-key validator for internal compiler package reuse and invoke it before replay-case decoding. Added negative tests for repeated top-level `schema` and nested `expected.value` keys.

**Verification:** `gofmt`, `git diff --check`, and targeted `go test ./internal/compiler/session -run '^TestDecodeReplayCasesRejectsDuplicateJSONKeys$' -count=1` passed. Both duplicate-key inputs are rejected. The full test suite was not run.

---

_Fixed: 2026-09-27T20:22:44Z_  
_Fixer: the agent (gsd-code-fixer)_  
_Iteration: 1_
