---
phase: quick
plan: 260924-k1v
verified: 2026-09-24
status: failed
score: 3/4 must-haves verified
covered_files:
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_peer_gate_test.go
  - .planning/LANGUAGE-MATURITY.md
---

# Quick Task 260924-k1v Verification

**Goal:** Reconcile pathoracle's concrete-path loan endpoints with the accepted Phase 18 computed-match fixture and refresh the corpus count.

**Status:** The fixture, retained refusal guards, corpus gate, and maturity count pass. The full Go suite exits nonzero on additional check/core failures and one cgen timeout; parent triage is required before treating the whole-plan success criterion as green.

## Must-Haves

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | The `loan_across_branch` fixture agrees exactly with checker and corevalidate, with an On-arm point after take and an Off successor edge. | VERIFIED | `TestOracleAgreesWithComputedMatchLoan` passed and asserts whole-value equality plus endpoint shape and IDs. |
| 2 | Path enumeration remains independent and bounded, retaining malformed-path, cycle, path-cap, and composition guards. | VERIFIED | Focused refusal/composition tests passed; full pathoracle and session packages passed. |
| 3 | The corpus-wide gate includes the new fixture and reports no endpoint divergence; maturity states 137 programs and 4,572 lines. | VERIFIED | Focused corpus gate, maturity count, and self-describing-doc checks passed. |
| 4 | The full Go suite passes. | NOT VERIFIED | `GOCACHE=/tmp/ai-lang-gocache go test ./...` exited 1; failures listed below. |

## Verification Commands

- Focused pathoracle regression and retained controls — PASS.
- Focused session corpus and maturity checks — PASS.
- `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/pathoracle ./internal/compiler/session -count=1` — PASS.
- `git diff --check -- internal/compiler/pathoracle/pathoracle.go internal/compiler/pathoracle/pathoracle_test.go internal/compiler/session/session_peer_gate_test.go .planning/LANGUAGE-MATURITY.md` — PASS.
- `GOCACHE=/tmp/ai-lang-gocache go test ./...` — FAIL, exit 1.

## Full-Suite Failures

- `internal/compiler/cgen`: `TestPayloadTracerThreeEngineAgreement` timed out in its native run: `native.timeout: context deadline exceeded`.
- `internal/compiler/check`: `TestPhase18LoanAcrossBranchFixture/both_arms` rejected the control source with `ownership.use_after_move [435:439]: value was used after ownership transferred`.
- `internal/compiler/core`: `TestPreviousPhaseCoreBytesUnchanged` and `TestPreviousPhaseManifestIDsUnchanged` report changed hashes/IDs for seven historical Phase 3–5 fixtures: `borrowed_view`, `branch_one_arm_shared_accept`, `branch_view`, `public_view_multi_arm_access_conflict`, `public_view_multi_arm_omitted`, `defect_terminal`, and `defect_dies_by_signal`. The exact observed pairs are available in the full-suite command output from this run; these are recorded for parent triage because they may be downstream effects of recent Phase 18 representation changes.

Other reported package results included passing `corevalidate`, `native`, `pathoracle`, `protocol`, `reduce`, `session`, `syntax`, and `testsupport` packages.

## Scope

Changed production/test files are limited to pathoracle; the existing corpus gate needed no edit. `.planning/LANGUAGE-MATURITY.md` and the planning execution records were updated. Pre-existing `.planning/config.json` changes and untracked artifacts were preserved.
