---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
reviewed: 2026-10-14T22:21:53Z
depth: standard
files_reviewed: 51
files_reviewed_list:
  - cmd/schway/main_phase35_test.go
  - docs/agent-surface.md
  - examples/bitwise.schway
  - examples/checksum/adler32.schway
  - internal/compiler/cgen/cgen_phase35_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase35_baseline_test.go
  - internal/compiler/check/check_phase35_numeric_test.go
  - internal/compiler/check/check_phase35_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_phase35_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase35_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_phase35_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_phase35_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_phase35_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_phase35_test.go
  - internal/compiler/session/session_phase34_compat_test.go
  - internal/compiler/session/session_phase34_limit_consumers_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase35_adler_test.go
  - internal/compiler/session/session_phase35_closeout_test.go
  - internal/compiler/session/session_phase35_peer_mutation_test.go
  - internal/compiler/session/session_phase35_prechange_test.go
  - internal/compiler/session/session_phase35_surface_test.go
  - internal/compiler/session/session_phase35_values_test.go
  - internal/compiler/surface/refusals.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/numeric_phase35_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase34_recovery_test.go
  - internal/compiler/syntax/parser_phase35_baseline_test.go
  - internal/compiler/syntax/parser_phase35_test.go
  - testdata/phase35/numeric-delta.json
  - testdata/phase35/numeric-policy-baseline.json
  - testdata/phase35/prechange-corpus.json
  - testdata/phase35/recovery-after.json
  - testdata/phase35/shift-sites.json
  - testdata/phase35/surface-census.json
  - testdata/phase35/surface-witnesses.json
  - testdata/phase35/value-vectors.json
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 35: Code Review Report

**Reviewed:** 2026-10-14T22:21:53Z  
**Depth:** standard  
**Files Reviewed:** 51  
**Status:** issues_found

## Summary

Reviewed the Phase 35 source, tests, fixtures, and agent-facing documentation at standard depth. Two parser/formatter edge cases remain: formatting can join separated `>` tokens into a shift operator, and a parenthesized negative shift amount bypasses the dedicated refusal diagnostic. No critical security or data-loss issue was identified. No tests were run.

## Narrative Findings (AI reviewer)

### Warnings

#### WR-01: Formatter joins separated right-angle tokens into a shift operator

**Severity:** WARNING  
**File:** `internal/compiler/syntax/format.go:431-434`  
**Issue:** `Format` combines adjacent `<`/`>` tokens into an operator only when their source spans touch (`format.go:37-45`), but the `TokenRAngle` renderer always trims whitespace and writes `>` with no trailing separator. Consequently malformed or recovering scalar input such as `x > > x` is emitted with contiguous `>>` and reparses as a right-shift expression. The parser explicitly treats spaced `>` tokens as non-shift tokens (`parser_phase35_test.go:124-128`), so formatting changes the tokenization and recovery result.
**Fix:** In scalar-expression context, preserve a separator between `TokenRAngle` tokens unless their source spans were contiguous and deliberately combined. Add a format/reparse regression for `x > > x` and `x > /* separated */ > x` that verifies formatting does not create a `shift_right` node.

#### WR-02: Parenthesized negative shift amounts miss the shift refusal

**Severity:** WARNING  
**File:** `internal/compiler/syntax/parser.go:833-847` and `internal/compiler/syntax/parser.go:984-987`  
**Issue:** The dedicated `surface.unsupported_shift_amount` path recognizes a `-` only when it immediately follows the shift operator. For `x << (-1)`, parsing enters the generic parenthesized-expression path instead; `-` is reported as an expected scalar operand and recovery then reports an expected `)` at `1`. The input is still rejected, but it loses the registered shift-amount diagnostic and its promised guidance for a negative amount. The current negative-amount test covers only `x << -1` (`parser_phase35_test.go:86`).
**Fix:** Preserve shift-amount context while parsing transparent groups, or recognize a grouped unary-minus integer literal as one shift amount. Emit `surface.unsupported_shift_amount` for `x << (-1)` and nested/comment-separated variants, and add tests asserting that the dedicated refusal is present.

---

_Reviewed: 2026-10-14T22:21:53Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_
