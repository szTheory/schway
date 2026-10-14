---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
fixed_at: 2026-10-14T22:28:11Z
review_path: .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW.md
iteration: 1
findings_in_scope: 2
fixed: 2
skipped: 0
status: all_fixed
---

# Phase 35: Code Review Fix Report

**Fixed at:** 2026-10-14T22:28:11Z
**Source review:** `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REVIEW.md`
**Iteration:** 1

**Summary:**
- Findings in scope: 2
- Fixed: 2
- Skipped: 0

## Fixed Issues

### WR-01: Formatter joins separated right-angle tokens into a shift operator

**Files modified:** `internal/compiler/syntax/format.go`, `internal/compiler/syntax/parser_phase35_test.go`
**Commit:** `e392c673`
**Fix status:** fixed; confirmed by independent iteration-2 re-review. No separate human-only evaluation is needed.
**Applied fix:** Keep unmatched right-angle tokens separated in scalar-expression formatting. Added whitespace and line-comment recovery cases that verify reparsing has no `shift_right` node and formatting is idempotent.

### WR-02: Parenthesized negative shift amounts miss the shift refusal

**Files modified:** `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/parser_phase35_test.go`
**Commit:** `863d82ea`
**Fix status:** fixed; confirmed by independent iteration-2 re-review. No separate human-only evaluation is needed.
**Applied fix:** Recognize a negative integer wrapped in transparent groups only at shift-amount sites and emit `surface.unsupported_shift_amount`. The iterative recognition observes the existing expression-depth limit, and the recovery path retains the shift refusal. Tests cover grouped, nested, comment-separated, over-depth, and recovery cases.

## Verification

Verification ran in the pinned main checkout selected by the orchestrator. The exact focused command passed:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase35ScalarGrammar|TestPhase35ExpressionDepth|TestPhase35GroupedNegativeShiftAmountRefusal|TestPhase35FormatterPreservesSeparatedShiftTokens|TestPhase35FormatterRoundTrip|TestPhase35RecoveryDeltaControls|TestPhase34RecoveryCorpus|TestPhase34RecoveryControls)$' -count=1 -v -timeout=600s
```

---

_Fixed: 2026-10-14T22:28:11Z_
_Fixer: the agent (gsd-code-fixer)_
_Iteration: 1_
