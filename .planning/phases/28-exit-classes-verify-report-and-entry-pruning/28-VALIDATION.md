---
phase: "28"
slug: "exit-classes-verify-report-and-entry-pruning"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-10"
---

# Phase 28 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Test map source: `28-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` (go1.24.0, stdlib only) |
| **Config file** | none (go.mod only) |
| **Quick run command** | `go test ./internal/compiler/exitcode ./internal/compiler/callgraph -count=1` |
| **Full suite command** | Hosted CI only (`go test ./...`). Never run it locally (AGENTS.md). |
| **Estimated runtime** | ~5–60 s per focused run (clang-backed CLI tests are slower) |

---

## Sampling Rate

- **After every task commit:** run the task's focused `-run` command for the touched package.
- **After every plan wave:** run all focused rows for that wave, plus `go test ./internal/compiler/session -run 'TestPhase16|TestPhase271|TestPhase272|TestPhase273|TestSelfDescribing|Inventory|Groundedness|Reconciliation|ValidationLifecycle' -count=1`.
- **Before `/gsd-verify-work`:** hosted full and race CI must be green on both operating systems.
- **Max feedback latency:** about 120 s per focused run.

---

## Per-Task Verification Map

Plan and task IDs are filled in once the plans exist. The requirement-level map:

| Requirement | Behavior | Test Type | Automated Command | File Exists | Status |
|-------------|----------|-----------|-------------------|-------------|--------|
| DX-17 | Table pin: 65, 69, uniqueness | unit | `go test ./internal/compiler/exitcode -run 'TestExitCodeTable' -count=1` | ❌ W0 | ⬜ pending |
| DX-17 | AST scan finds no exit literals; negative-control self-test | unit/AST | `go test ./internal/compiler/exitcode -run 'TestNoDuplicatedExitLiterals\|TestExitScanDetectorFires' -count=1` | ❌ W0 | ⬜ pending |
| DX-17 | Emitted-C exit literals stay within the table | unit | `go test ./internal/compiler/cgen -run 'TestEmittedCExitLiteralsMatchTable\|TestPhase273CapacityContractLiteralsAgree' -count=1` | partial | ⬜ pending |
| DX-17 | Same class and code from check, build, run interp and run native | CLI | `go test ./internal/compiler/native -run 'TestShippedBinaryFourSubcommandCorpusMatrix\|TestExitClassesUniformAcrossSurfaces' -count=1` | partial | ⬜ pending |
| DX-17 | Tool or IO failure gives 3, mismatch gives 4 | CLI | `go test ./internal/compiler/session -run 'TestPhase16ProductionPathsPreserveM004Refusal\|TestNativeToolFailureIsOperational' -count=1` | ✅ (edit) | ⬜ pending |
| DX-17 | app run exits 124, 125, 128+n and 65 | unit | `go test ./cmd/schway -run 'TestAppRunExitCode\|TestPhase22\|TestPhase27PublicAppCLI' -count=1` | partial | ⬜ pending |
| DX-18 | PruneToReachable: deep-equal, order preserved, unique root | unit | `go test ./internal/compiler/callgraph -run 'TestPrune\|TestEntryFunction' -count=1` | ❌ W0 | ⬜ pending |
| DX-18 | EntryFunction is called only through resolveEntry | AST | `go test ./internal/compiler/session -run 'TestEntryFunctionCalledOnlyViaResolveEntry' -count=1` | ❌ W0 | ⬜ pending |
| DX-18 | Corpus byte identity except allow-listed dead-function files | golden | `go test ./internal/compiler/session -run 'TestPhase273ConformanceCorpusByteIdentity\|TestPhase272\|TestPhase28' -count=1` | partial | ⬜ pending |
| DX-18 | An unreachable-function program runs; an ambiguous entry exits 2 | CLI | `go test ./internal/compiler/native -run 'TestUnreachableFunctionProgramRuns' -count=1` | ❌ W0 | ⬜ pending |
| EVD-13 | n=14 passes; n=15, 172, 173, 1000 are unsupported with the named reason | integration | `go test ./internal/compiler/session -run 'TestPhase273AppVerify\|TestAppVerifyExhaustedCases' -count=1` | partial | ⬜ pending |
| EVD-13 | 16 exhausted cases fit the budget and are never counted as success | integration/static | `go test ./internal/compiler/session -run 'TestAppVerifyReportBudget\|TestAppVerifySixteenExhausted' -count=1` | ❌ W0 | ⬜ pending |
| EVD-13 | Digest re-hash, named oversize error, ID bound | unit | `go test ./internal/compiler/session -run 'TestAppVerifyOutcomeDigest\|TestAppVerifyReportOversize\|TestDecodeReplayCases' -count=1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `testdata/phase28/entry-exit-class-manifest.json`, with an env-guarded generator test, committed before any behavior change
- [ ] Green baseline at HEAD for `TestPhase272*` and `TestPhase273ConformanceCorpusByteIdentity`
- [ ] Exit-code pin inventory (re-grepped at plan time) committed in the phase directory
- [ ] `internal/compiler/exitcode` package with its pin test, AST scan and self-test (behavior-neutral; a temporary allow-list shrinks as call sites move)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Hosted macOS and Linux CI green | DX-17, DX-18, EVD-13 | The full suite is not run locally | Dispatch hosted CI after the user approves a push |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
