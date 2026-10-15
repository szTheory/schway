---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
reviewed: 2026-10-15T21:24:53Z
depth: standard
files_reviewed: 66
files_reviewed_list:
  - cmd/schway/main_phase32_routes_test.go
  - cmd/schway/main_phase35_test.go
  - docs/agent-surface.md
  - examples/bitwise.schway
  - examples/checksum/adler32.schway
  - internal/compiler/cgen/cgen.go
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
  - internal/compiler/session/session_phase272_identity_test.go
  - internal/compiler/session/session_phase273_identity_test.go
  - internal/compiler/session/session_phase34_compat_test.go
  - internal/compiler/session/session_phase34_limit_consumers_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_presentation_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase35_adler_test.go
  - internal/compiler/session/session_phase35_admission_mutation_test.go
  - internal/compiler/session/session_phase35_closeout_test.go
  - internal/compiler/session/session_phase35_peer_mutation_test.go
  - internal/compiler/session/session_phase35_prechange_test.go
  - internal/compiler/session/session_phase35_surface_test.go
  - internal/compiler/session/session_phase35_values_test.go
  - internal/compiler/surface/refusals.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/numeric_phase35_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase34_recovery_test.go
  - internal/compiler/syntax/parser_phase34_refusal_test.go
  - internal/compiler/syntax/parser_phase35_baseline_test.go
  - internal/compiler/syntax/parser_phase35_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase35/numeric-delta.json
  - testdata/phase35/numeric-policy-baseline.json
  - testdata/phase35/prechange-corpus.json
  - testdata/phase35/recovery-after.json
  - testdata/phase35/shift-sites.json
  - testdata/phase35/surface-census.json
  - testdata/phase35/surface-witnesses.json
  - testdata/phase35/value-vectors.json
  - testdata/phase36/new-fixture-ledger.json
  - testdata/phase37/new-fixture-ledger.json
  - testdata/phase35/prohibition-evidence.test.cjs
  - testdata/phase35/prohibition-causal-bad.txt
  - testdata/phase35/prohibition-lanes-bad.txt
  - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-PLAN.md
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
status: clean
---

# Phase 35: Code Review Report

**Reviewed:** 2026-10-15T21:24:53Z
**Depth:** standard
**Files Reviewed:** 66 unique paths across two review passes
**Status:** clean

## Summary

The original explicit 62-path source scope was reviewed at standard depth at 2026-10-15T21:02:28Z. No correctness, security, or maintainability defect was confirmed. The review accounts for the current Phase 36 scalar-helper and Phase 37 file-run additions as successor changes, including their fixture ledgers and route/witness composition; it does not treat earlier absent-feature descriptions as current violations.

The current evidence adapter supplies the shared validator’s required coverage description. The refresh receipt records 72 passing package-qualified roots with zero missing, failed, or skipped roots at its recorded source head `d48227882579947201f607814585c42e6ad29`; that receipt is historical evidence, not a test run against the current HEAD. No tests were rerun during the original 62-path pass.

Earlier WR-01 and WR-02 remain preserved as fixed in [35-REVIEW-DISPOSITION.md](35-REVIEW-DISPOSITION.md).

## Follow-up Review (2026-10-15T21:24:53Z)

At HEAD `04d1407a`, the follow-up reviewed the exact five-file wiring scope: `internal/compiler/session/session_phase35_closeout_test.go`, `testdata/phase35/prohibition-evidence.test.cjs`, `testdata/phase35/prohibition-causal-bad.txt`, `testdata/phase35/prohibition-lanes-bad.txt`, and `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-08-PLAN.md`. The Go test is the one overlap with the earlier 62 paths; the report frontmatter lists their 66-path union and does not imply that the earlier 62 were reread.

The test-only subject override resolves relative inputs from the project root even when Go runs from the package directory, while ordinary runs retain the committed evidence file. The Node adapter forwards that subject environment, executes only `TestPhase35EvidenceGrounding`, has a child timeout longer than Go’s test timeout, and requires its exact root `run` and `pass` events with no failure or skip events. The two fixtures each alter only the intended false claim; the Go validator rejected the causal fixture at the causal/numeric clause and the lane fixture at the missing-lane clause. The clean fixture and the normal default subject passed.

I ran the targeted adapter four times at this HEAD: each known-bad fixture produced the expected failing Node test, and both the clean fixture and normal run passed. This confirms the current-path behavior; the historical enforcement receipt at `c5a22142` remains a separate earlier record. No evidence rule was copied into the adapter or weakened by the plan amendment. No new findings were confirmed.

## Narrative Findings (AI reviewer)

No new findings from either pass. The prior WR-01 and WR-02 findings are recorded as fixed and were not reopened.

---

_Reviewed: 2026-10-15T21:24:53Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
