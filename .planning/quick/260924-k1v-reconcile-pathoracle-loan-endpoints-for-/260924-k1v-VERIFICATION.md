---
phase: quick
plan: 260924-k1v
verified: 2026-09-24
status: passed
score: 4/4 must-haves verified
covered_files:
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_peer_gate_test.go
  - .planning/LANGUAGE-MATURITY.md
---

# Quick Task 260924-k1v Verification

**Goal:** Reconcile pathoracle's concrete-path loan endpoints with the accepted Phase 18 computed-match fixture and refresh the corpus count.

**Status:** All four must-haves pass. The follow-up regression closure restored historical core/manifest compatibility, replaced the invalid both-arm source control with a valid synthetic CFG control, and the full Go suite now passes.

## Must-Haves

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | The `loan_across_branch` fixture agrees exactly with checker and corevalidate, with an On-arm point after take and an Off successor edge. | VERIFIED | `TestOracleAgreesWithComputedMatchLoan` passed and asserts whole-value equality plus endpoint shape and IDs. |
| 2 | Path enumeration remains independent and bounded, retaining malformed-path, cycle, path-cap, and composition guards. | VERIFIED | Focused refusal/composition tests passed; full pathoracle and session packages passed. |
| 3 | The corpus-wide gate includes the new fixture and reports no endpoint divergence; maturity states 137 programs and 4,572 lines. | VERIFIED | Focused corpus gate, maturity count, and self-describing-doc checks passed. |
| 4 | The full Go suite passes. | VERIFIED | `GOCACHE=/tmp/ai-lang-gocache go test ./...` passed after regression closure commit `e91628c`. |

## Verification Commands

- Focused pathoracle regression and retained controls — PASS.
- Focused session corpus and maturity checks — PASS.
- `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/pathoracle ./internal/compiler/session -count=1` — PASS.
- `git diff --check -- internal/compiler/pathoracle/pathoracle.go internal/compiler/pathoracle/pathoracle_test.go internal/compiler/session/session_peer_gate_test.go .planning/LANGUAGE-MATURITY.md` — PASS.
- `GOCACHE=/tmp/ai-lang-gocache go test ./...` — PASS after `e91628c`; this also reconfirms the pathoracle change against the full repository suite.
- Compatibility pins pass with `go test ./internal/compiler/core -run '^(TestPreviousPhaseCoreBytesUnchanged|TestPreviousPhaseManifestIDsUnchanged)$'`.
- The native payload tracer passed five isolated repetitions; no timeout adjustment was needed.

## Scope

Changed production/test files are limited to pathoracle; the existing corpus gate needed no edit. `.planning/LANGUAGE-MATURITY.md` and the planning execution records were updated. Pre-existing `.planning/config.json` changes and untracked artifacts were preserved.
