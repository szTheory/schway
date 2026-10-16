---
phase: "38"
slug: "milestone-receipts"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-16"
---

# Phase 38 — Validation Strategy

This is a proposed contract, not a passing receipt. The planner must replace Wave 0 placeholders with exact new test owners and selectors before execution.

## Test Infrastructure

| Property | Value |
| --- | --- |
| Framework | Go 1.24 test, existing shell gates and hosted GitHub Actions |
| Config | go.mod; .github/workflows/ci.yml |
| Existing focused baseline | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestFreezeGateDifferential|TestReceiptGateDifferential|TestPhase32CodeDigest)$' -count=1 -v -timeout=180s` |
| Local policy | Exact bounded selectors; no full local suite, whole-package sweep, broad native matrix or pressure capture |
| Hosted policy | Required checks and evidence aggregate on Ubuntu/macOS; existing full and race lanes plus complete M007 producer inventory |
| Runtime | Measure actual focused runs; do not invent a latency estimate |

## Sampling Rate

After each implementation task, run its focused behavioral controls and verify expected non-skipped roots/subtests were reached. After a preparation wave, run the union of changed acceptance seams once. Before freeze, complete relevant review/security/freshness and the new audit. After ingest, re-gate copied logs/provenance and the actual planning-only descendant. Full/race suites run hosted; a local observation cannot replace either host.

## Per-Task Verification Map

| Task group | Plan | Wave | Requirement | Security behavior | Test type | Command/evidence | Exists | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Producer/manifest/CI | 01 | 1 | EVD-17 prerequisite | Missing, duplicate, skipped or partial required rows fail; failed producer cannot be masked by tee | focused unit/seam | MISSING — Wave 0 producer selection/inventory controls; existing differential baseline above | new controls needed | pending |
| Run/attempt/source binding | 01 | 1 | EVD-17 prerequisite | Wrong revision/attempt/job/artifact/log digest refuses; planning-only descendant allowed | focused seam | MISSING — Wave 0 binding positive/armed negative controls | new controls needed | pending |
| Audit/freeze | 01 | 1 | EVD-17 prerequisite | Clean digest worktree; completed preparation audited; actual SHA/digest/manifest recorded | gate | Existing freeze gate plus scoped current-source checks; exact invocation comes from actual freeze | existing/new data | pending |
| Remote action/hosted run | 02 | 2 | EVD-17 prerequisite | Commit-specific authorization, exact selected ref, both hosts and required jobs complete | hosted integration | CI run metadata, M007 artifacts, required executed roots, full/race conclusions | remote checkpoint pending | pending |
| Durable copied-byte ingest | 03 | 3 | EVD-17 final acceptance | Exact run/attempt/freeze/log binding; final code drift rejected; only .planning modified | gate | Generic receipt/freeze gates and new narrow binding seam against copied bytes | Wave 0 seam needed | pending |

## Wave 0 Requirements

- [ ] Exact M007 program/case/control inventory and a bounded test selection that exercises the producer's real dispatch.
- [ ] Reached omission/duplicate/wrong-result/empty-selector/skip/partial-inventory controls.
- [ ] Exact run/attempt/freeze/log binding and wrong-revision controls; unchanged generic gate refusals.
- [ ] CI artifact membership, upload-on-failure and producer pipeline failure controls.
- [ ] No new framework/dependency installation.

The planner supplies task IDs, unique threat IDs, exact files/commands and observable failing directions. Source-inspected existing tests are not newly passing evidence.

## Manual-Only Verification

Only authorization to push the named freeze and dispatch the named workflow requires the user. The checkpoint must name the full actual commit/digest, repository/ref, workflow/inputs and push scope. All objectively checkable behavior, raw bytes and hosted provenance are agent-verified. Missing external access/results leave EVD-17 pending; they do not create subjective UAT.

## Validation Sign-Off

- [ ] All tasks have automated checks or an explicit authority checkpoint.
- [ ] No three consecutive tasks lack objective automated evidence.
- [ ] Wave 0 covers missing controls without mirrored implementation tests.
- [ ] No watch mode, hidden broad local suite or skipped required case accepted.
- [ ] Both hosts and required full/race/evidence lanes pass at the selected freeze.
- [ ] Copied logs and metadata pass binding and final drift checks.
- [ ] `nyquist_compliant: true` only after execution evidence establishes it.

**Approval:** pending execution evidence.
