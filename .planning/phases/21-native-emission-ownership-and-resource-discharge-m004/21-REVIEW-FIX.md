---
phase: 21-native-emission-ownership-and-resource-discharge-m004
fixed_at: 2026-09-27T00:59:41Z
review_path: ~/projects/ai-lang/.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-REVIEW.md
iteration: 1
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 21: Code Review Fix Report

**Fixed at:** 2026-09-27T00:59:41Z  
**Source review:** `~/projects/ai-lang/.planning/phases/21-native-emission-ownership-and-resource-discharge-m004/21-REVIEW.md`  
**Iteration:** 1

**Summary:**
- Findings in scope: 1
- Fixed: 1
- Skipped: 0

## Fixed Issues

### CR-01: Payload return bytes are omitted from output-size preflight

**Files modified:** `internal/compiler/cgen/cgen_program.go`, `internal/compiler/cgen/cgen_program_test.go`, `internal/compiler/cgen/export_test.go`  
**Commit:** `d63e2c2`  
**Applied fix:** The preflight now bounds returned ADTs across all alternatives, including JSON-escaped tags, the maximum eight hex characters for a four-byte Buffer, the three-digit Byte maximum, and nested tag payloads. The terminal writer and preflight share payload-kind selection and width constants. Added a focused payload-return refusal control.

**Verification:** `gofmt` and `git diff --check` passed. Focused Go tests passed: `TestSchema2PayloadOutcomeBoundIsPreflighted`, `TestProgramMatchPayloadLowering`, and `TestSchema2ExecutionOutputBoundIsPreflighted`. They ran in the isolated review-fix worktree at `~/projects/ai-lang/.claude/worktrees/rf-21-26593-1790470445`, using `GOCACHE=/private/tmp/codex-ai-lang-go-build`. The logic change requires human verification.

---

_Fixed: 2026-09-27T00:59:41Z_  
_Fixer: the agent (gsd-code-fixer)_  
_Iteration: 1_
