---
status: resolved
trigger: "G-16-21-C: TestDebtRegistersAreWellFormed reports missing Phase 13 witness test names; TestBudgetLaneCarriesMachineIDAndVerdict, TestBudgetAuditRefusesUndeclaredMachine, and TestQLT02InterproceduralGrowthExponent fail at ProbeMachine."
created: 2026-09-23T15:00:00Z
updated: 2026-09-26T19:47:39Z
---

## Current Focus

hypothesis: Two independent deterministic causes explain this grouped UAT gap: the Phase 13 archived debt register retains probe names removed/replaced by Phase 17 while the global debt-register validator still requires every probe name to resolve to a current Go test; and the three Phase 6 budget tests call the real Darwin machine probe, whose `sysctl` subprocess is denied by this sandbox.
test: Re-run the four reported tests in isolation, inspect the Phase 13 debt row and Phase 17 witness migration, and invoke the exact Darwin `sysctl` probe directly.
expecting: The debt-register test names only the stale Phase 13 witness tokens; all three budget tests fail with the same typed probe error; direct `sysctl` shows the host permission denial.
next_action: Return the two confirmed root causes and supporting file/command evidence to the UAT orchestrator for gap-closure planning.
bug_class: Bohrbug (stale evidence names) plus environment/config restriction (host probe).
reasoning_checkpoint:
  hypothesis: "The D-13-02b and D-13-10a register witness tokens name a test removed/replaced in Phase 17, and the Phase 6 budget tests fail because sandbox policy denies Darwin's `sysctl -n machdep.cpu.brand_string` probe."
  confirming_evidence:
    - "Focused `go test` reports exactly two stale witness tokens in PHASE-13-DEBT.md, both `probe:TestB1BlameIsStructurallyUnreachable`."
    - "Phase 17 defines `TestPhase17B1RequiresUnverifiableDeclaredContract` and its commit updates `.planning/UNREACHABLE-CLAIMS.md` while leaving the archival Phase 13 debt register untouched."
    - "All three budget tests fail at `liveMachineIDForTest` with `measure.probe_failed`; direct sysctl exits 1 with `Operation not permitted` while clang succeeds."
  falsification_test: "The explanation would be false if `TestB1BlameIsStructurallyUnreachable` existed in the current Go test list, if the Phase 13 witness tokens resolved to another current test, or if direct sysctl succeeded in this environment."
  fix_rationale: "The obsolete witness facts must be reconciled with the extant Phase 17 evidence in the archived register/validator contract; the live machine probe failures require a test/host-probe boundary that remains verifiable when Darwin sysctl access is restricted, or execution in a host context where that probe is permitted."
  blind_spots: "Diagnosis did not change code or test alternate execution environments. It does not determine the precise preferred policy for historical witness tokens or whether the CI runner permits the same sysctl query."
  candidate_causes:
    - "data: Phase 13 debt register witness rows were not synchronized when Phase 17 superseded their test evidence."
    - "environment: sandbox host policy denies the sysctl CPU-model probe required by ProbeMachine."
  and_gate: "No. The stale witness failure and host-probe failures are independent sub-failures grouped under one UAT gap; either can occur without the other."

## Symptoms

expected: "The four focused tests pass: every debt witness token resolves to a current test and host-dependent budget checks can obtain a declared machine identity."
actual: "The debt-register validator reports two nonexistent Phase 13 test names. Three budget tests terminate at ProbeMachine with measure.probe_failed."
errors: "PHASE-13-DEBT.md D-13-02b and D-13-10a: witness token `probe:TestB1BlameIsStructurallyUnreachable` names a test that does not exist. Budget tests: `ProbeMachine: measure.probe_failed`. Direct sysctl: `Operation not permitted`."
reproduction: "Run `GOCACHE=/tmp/ai-lang-go-cache go test ./internal/compiler/session -run '^(TestDebtRegistersAreWellFormed|TestBudgetLaneCarriesMachineIDAndVerdict|TestBudgetAuditRefusesUndeclaredMachine|TestQLT02InterproceduralGrowthExponent)$' -count=1 -v` in this workspace. Run `sysctl -n machdep.cpu.brand_string` to observe the host restriction."
started: "Observed during Phase 16 UAT on 2026-09-23."

## Eliminated

- hypothesis: "The Phase 13 debt register has an invalid shape or another register causes TestDebtRegistersAreWellFormed to fail."
  evidence: "Verbose isolated run passes every listed debt-register subtest except PHASE-13-DEBT.md; that one emits only the two stale witness-token diagnostics."
  timestamp: 2026-09-23
- hypothesis: "Clang is missing or the budget manifest lacks an entry for this host."
  evidence: "`/usr/bin/clang --version` succeeds. Each affected test fails in `liveMachineIDForTest` before comparing the manifest, and direct `sysctl -n machdep.cpu.brand_string` returns `Operation not permitted`."
  timestamp: 2026-09-23

## Evidence

- timestamp: 2026-09-23
  checked: "Focused isolated Go command for the four reported tests."
  found: "`TestDebtRegistersAreWellFormed` fails only under PHASE-13-DEBT.md rows D-13-02b and D-13-10a, with exact stale token `probe:TestB1BlameIsStructurallyUnreachable`. The three budget tests all fail at their call to liveMachineIDForTest with `measure.probe_failed`."
  implication: "The debt failure is data/test-registry drift. The shared probe failure occurs before each test's distinct assertion and has one common lower-level cause."
- timestamp: 2026-09-23
  checked: "Current Go test definitions and commit 163fc8b (`test(17-09): ratify M006 blame boundary`)."
  found: "No `TestB1BlameIsStructurallyUnreachable` function exists. `internal/compiler/session/witness_registry_test.go` defines `TestPhase17B1RequiresUnverifiableDeclaredContract`; commit 163fc8b migrates the derived `.planning/UNREACHABLE-CLAIMS.md` witness to that test but does not modify the Phase 13 archived source register."
  implication: "The Phase 13 source register and generated view have diverged: the generated claim now tracks Phase 17 evidence while historical source rows still cite the superseded Phase 13-era probe. The all-register test catches this source-level rot."
- timestamp: 2026-09-23
  checked: "`clang --version` and direct `sysctl -n machdep.cpu.brand_string`."
  found: "Clang succeeds (`Apple clang version 21.0.0`); sysctl exits 1 with `sysctl fmt -1 1024 1: Operation not permitted`."
  implication: "`ProbeMachine` first completes the clang probe, then the Darwin CPU-model probe fails. `runBoundedProbe` maps the failed sysctl process to `measure.probe_failed`; this is specific to host sandbox policy, not an absent compiler."
- timestamp: 2026-09-23
  checked: "Probe implementation and test helper (`internal/compiler/measure/machine.go`, `internal/compiler/session/session_phase6_budget_test.go`)."
  found: "`ProbeMachine` calls clang then Darwin `sysctl`; the three reported session tests use `liveMachineIDForTest`, which invokes the real host probe rather than an injected fixture."
  implication: "The three test names are one environmental failure family."

## Resolution

root_cause: "Independent causes: (1) stale witness-data synchronization — Phase 17 replaced the Phase 13-era B1 probe with TestPhase17B1RequiresUnverifiableDeclaredContract and updated the generated claims view, but PHASE-13-DEBT.md D-13-02b and D-13-10a still reference nonexistent TestB1BlameIsStructurallyUnreachable, so the global TestDebtRegistersAreWellFormed check fails; (2) environment restriction — the three budget tests call the real Darwin ProbeMachine, and its sysctl CPU-model subprocess is denied by the sandbox with Operation not permitted, which ProbeMachine surfaces as measure.probe_failed."
fix: "Not applied (diagnosis-only task)."
verification: "Reproduced all four failures with an isolated uncached test command and directly confirmed the denied sysctl operation."
files_changed: [".planning/debug/phase16-debt-and-machine-probe.md"]

## Post-diagnosis closure (2026-09-26)

Phase 16 Plan 19 replaced the stale Phase 13 witness names with the current Phase 17 witness and changed the budget tests to use deterministic injected machine facts while keeping bounded real-probe behavior separately tested. Phase 16 UAT gap G-16-21-C is resolved by Plan 19; its verification and the current M003 audit pass. The original diagnosis-only scope is preserved above; this section records the later closure evidence.
