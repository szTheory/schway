---
phase: 36
fixed_at: 2026-10-15T06:26:21Z
review_path: .planning/phases/36-scalar-helper-calls/36-REVIEW.md
iteration: 1
findings_in_scope: 1
fixed: 1
skipped: 0
status: all_fixed
---

# Phase 36: Code Review Fix Report

**Fixed at:** 2026-10-15T06:26:21Z
**Source review:** .planning/phases/36-scalar-helper-calls/36-REVIEW.md
**Iteration:** 1

**Summary:**
- Findings in scope: 1
- Fixed: 1
- Skipped: 0

## Fixed Issues

### WR-01: Truncated repeated-call traces omit the expected callee

**Files modified:** `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/executionpeer/executionpeer_phase36_test.go`
**Commit:** 61217af4
**Applied fix:** End-of-stream refusals now read expected callee identity from dynamically observed occurrences when dynamic call identity is active, while retaining the static occurrence map for legacy traces. Added repeated-trace truncation controls at both the pending-child and open-activation boundaries for ordinal 1; they assert the exact refusal code, invocation context, and checked call's callee ID.

## Verification

Verification ran in the main checkout. `gofmt -d` reported no formatting changes. The exact focused selectors below passed, including first-call ordinal-zero compatibility:

```text
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/executionpeer -run '^(TestPhase36DynamicCallOrdinalControls|TestPhase36CallCausalAttributionControls|TestExecutionPeerFailuresAreActionable|TestValidateObservedCausalStructure)$' -count=1 -v -timeout=600s
```

The test-first run failed specifically because the truncated ordinal-1 `orphan_child` refusal had an empty `ExpectedFunction`; after the lookup repair, the focused selector set passed.

---

_Fixed: 2026-10-15T06:26:21Z_
_Fixer: the agent (gsd-code-fixer)_
_Iteration: 1_
