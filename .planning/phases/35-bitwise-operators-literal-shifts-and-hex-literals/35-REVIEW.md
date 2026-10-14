---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
reviewed: 2026-10-14T22:29:58Z
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
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 35: Code Review Report

**Reviewed:** 2026-10-14T22:29:58Z
**Depth:** standard
**Files Reviewed:** 51
**Status:** clean

## Summary

Re-reviewed the original 51-file Phase 35 scope at standard depth, focusing source inspection on the two repaired parser/formatter routes and their regression coverage. Both prior warnings are resolved: failed-source formatting preserves separated `>` tokens through reparsing, and grouped negative shift amounts receive the dedicated refusal while respecting the expression-depth cap. Grouped positive and computed shift amounts retain their existing parse paths, and nested generic formatting remains stable. No new material issue was found. No tests were run for this re-review; the fix report records the focused eight-selector syntax test command as historical evidence.

## Narrative Findings (AI reviewer)

No material findings.

---

_Reviewed: 2026-10-14T22:29:58Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
