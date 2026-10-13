---
phase: "28"
slug: "exit-classes-verify-report-and-entry-pruning"
# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (set by validate-phase §6)
# Reconstructed after execution from the 28-01..28-08 PLAN/SUMMARY set (State B); retroactive, written in Phase 33.
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-13"
---

# Phase 28 — Validation Strategy

> Reconstructed after execution; retroactive. Hosted CI: not run for this revision. The ambiguous-entry shipped-binary observation's numeric exit code was hidden by a pipe and is pinned by tests instead.

## Test Infrastructure

| Property | Value |
|----------|-------|
| Framework | Go `testing`, go1.24 |
| Config file | none (go.mod only) |
| Quick run command | Focused commands in the per-requirement map |
| Full suite command | hosted CI only; not run for this revision |

## Sampling Rate

- Focused package tests were run for each row below during this retroactive reconstruction.
- No hosted CI or full local suite is claimed.

## Per-Requirement Verification Map

| Requirement | Plans | Behavior | Test Type | Automated Command | File Exists | Status |
|-------------|-------|----------|-----------|-------------------|-------------|--------|
| DX-17 | 01, 06, 07, 08 | Exit classes agree across surfaces and engines | integration | `go test ./internal/compiler/native -run 'TestExitClassesUniformAcrossSurfaces' -count=1` | ✅ | ✅ green |
| DX-17 | 06, 07 | Native tool failure remains operational | integration | `go test ./internal/compiler/native ./internal/compiler/session -run 'TestNativeToolFailureIsOperational' -count=1` | ✅ | ✅ green |
| DX-17 | 02, 07 | Exit codes share one table across packages | unit | `go test ./internal/compiler/exitcode -run 'TestNoDuplicatedExitLiterals' -count=1`; `go test ./internal/compiler/session -run 'TestExitCodeDocsTable' -count=1`; `go test ./internal/compiler/cgen -run 'TestEmittedCExitLiteralsMatchTable' -count=1` | ✅ | ✅ green |
| DX-17 | 08 | Relative native paths retain the right exit class | integration | `go test ./internal/compiler/session -run 'TestPhase28NativeCommandRelativePath' -count=1` | ✅ | ✅ green |
| DX-18 | 04, 06 | Unreachable functions remain runnable | integration | `go test ./internal/compiler/native -run 'TestUnreachableFunctionProgramRuns' -count=1` | ✅ | ✅ green |
| DX-18 | 01, 04 | Entry behavior is recorded by the manifest | integration | `go test ./internal/compiler/session -run 'TestPhase28EntryExitClassManifest' -count=1` | ✅ | ✅ green |
| DX-18 | 05, 08 | Pruning properties and reachable C identity remain pinned | unit | `go test ./internal/compiler/session -run 'TestPhase28PruneCorpusProperties|TestPhase28PrunedEmitMatchesGolden|TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity' -count=1` | ✅ | ✅ green |
| EVD-13 | 03, 07 | Exhausted verification cases retain named outcomes | integration | `go test ./internal/compiler/session -run 'TestAppVerifyExhaustedCases|TestAppVerifySixteenExhausted' -count=1` | ✅ | ✅ green |
| EVD-13 | 03, 07 | Reports remain within budget and include the outcome digest | unit | `go test ./internal/compiler/session -run 'TestAppVerifyReportBudget|TestAppVerifyReportOversize|TestAppVerifyOutcomeDigest' -count=1` | ✅ | ✅ green |
| EVD-13 | 07 | Refusal survives report failure and stale reports are removed | integration | `go test ./internal/compiler/session -run 'TestAppVerifyRefusalSurvivesReportWriteFailure|TestAppVerifyRemovesStaleReportOnRefusal' -count=1` | ✅ | ✅ green |
| CR-01 | 07 | Public emitter inventory matches source | unit | `go test ./internal/compiler/session -run 'TestPhase16PublicEmitterConsumerInventory' -count=1` | ✅ | ✅ green |

## Wave 0 Requirements

- [x] Exit-class and entry-manifest characterization pins exist.
- [x] App-verification report budget and outcome tests exist.

## Manual-Only Verifications

None claimed as newly performed. The archived verification report records the shipped-binary observations and their limits.

## Validation Sign-Off

- [x] Every requirement has at least one automated row.
- [x] Every test command names an existing test in its package.
- [x] This record is retroactive and does not claim hosted CI for this revision.

**Approval:** reconstructed 2026-10-13

## Validation Audit

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
