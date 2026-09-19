---
phase: "15"
slug: "event-identity-lang-execution-2"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-19"
---

# Phase 15 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 standard `testing` package |
| **Config file** | none |
| **Quick run command** | `go test ./internal/compiler/{execution,executionpeer,interp,cgen,native,session} -run 'Test(.*Invocation.*|.*FunctionCalled.*|.*DiamondSharedLeaf.*|.*PathTable.*|.*ComparisonFieldRouting.*)' -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | Measure during execution; Phase 14 observed the full suite near 193 seconds on its host |

---

## Sampling Rate

- **After every task commit:** Run the narrow package/test command named by that task.
- **After every plan wave:** Run `go test ./...`.
- **Before `$gsd-verify-work`:** Full suite must be green.
- **Max feedback latency:** Record cold and warm distributions; do not substitute a guessed single number.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 15-01-02 | 01 | 1 | OBS-01, OBS-02, OBS-04 | T-15-02, T-15-04 | Canonical `/2` tracer rejects ambiguity, freezes legacy bytes, and pins the old diamond frontier | unit/golden/frontier | `go test ./internal/compiler/execution ./internal/compiler/native ./internal/compiler/session -run 'TestInvocationGrammar|TestExecutionLegacyBytesFrozen|TestValidateExecutionSchema2|TestPhase15DiamondFrontierIsPinned' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-02-01 | 02 | 2 | OBS-01, OBS-04 | T-15-04 | Interpreter threads identity through every event path with explicit legacy selection | unit | `go test ./internal/compiler/interp -run 'TestInvocationThreadsThroughAllEventPaths|TestExecutionSchemaSelectionPreservesLegacy' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-02-02 | 02 | 2 | OBS-02 | T-15-03 | Admitted calls emit caller-owned preorder edges and projection-only removal is exact | unit/negative | `go test ./internal/compiler/interp -run 'TestFunctionCalledPreorderAndOwnership|TestFunctionCalledProjectionRemoval|TestRejectedCallEmitsNoCalledEdge' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-03-01 | 03 | 2 | OBS-03 | T-15-01, T-15-03 | Peer independently resolves and unfolds without engine imports | peer/structural | `go test ./internal/compiler/executionpeer -run 'TestIndependentEntryResolution|TestInvocationMembershipTraversal|TestExecutionPeerImportBoundary|TestFullCoverageControlIsSeparate' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-03-02 | 03 | 2 | OBS-03, OBS-04 | T-15-03 | Peer rejects forged ownership, pair, grammar, kind, and preorder facts actionably | peer/mutation | `go test ./internal/compiler/executionpeer -run 'TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-04-01 | 04 | 2 | OBS-04 | T-15-01 | Native preflight derives the real 61-node measurement before allocation | boundary | `go test ./internal/compiler/cgen -run 'TestInvocationPathTableDeepDiamondMeasures61|TestInvocationPreflightOrdering' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-04-02 | 04 | 2 | OBS-04 | T-15-01 | 4096 passes, 4097 refuses by name, and bypass is mutation-killed | boundary/mutation | `go test ./internal/compiler/cgen -run 'TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-05-01 | 05 | 3 | OBS-01, NAT-10 | T-15-03 | Native parent-indexed lookup distinguishes children for repeated parent occurrences | unit/structural | `go test ./internal/compiler/cgen -run 'TestProgramInvocationIndexThreading|TestParentIndexedChildLookup|TestInvocationTableEmissionIsDeterministic' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-05-02 | 05 | 3 | OBS-02, OBS-04 | T-15-03, T-15-04 | Native `/2` writer emits preorder edges and freezes legacy writers | unit/golden | `go test ./internal/compiler/cgen -run 'TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder|TestNativeFunctionCalledProjectionRemoval|TestLegacyEventWritersFrozen' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-06-01 | 06 | 3 | OBS-03, OBS-04 | T-15-03, T-15-04 | Comparator routes both fields and requires a peer verdict | differential | `go test ./internal/compiler/session -run 'TestComparisonFieldRoutingIsExhaustive|TestInvocationFieldsAreCompared|TestSchema2ComparisonRequiresPeerVerdict' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-06-02 | 06 | 3 | OBS-03 | T-15-03, T-15-04 | Producer and peer fault directions each demonstrate actual bad behavior | mutation | `go test ./internal/compiler/session -run 'TestExecutionProducerFaultIsCaughtByPeer|TestExecutionPeerAcceptanceFaultIsCaughtByControl' -count=1 -v` | ❌ W0 | ⬜ pending |
| 15-07-01 | 07 | 4 | OBS-01, OBS-02, OBS-03, NAT-10 | T-15-03, T-15-04 | Existing DiamondSharedLeaf passes all four tiers and restored collision fails | integration/mutation | `go test ./internal/compiler/session -run 'TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert|TestPhase15DiamondFrontierMoved' -count=1 -v` | ✅ expectation flips | ⬜ pending |
| 15-07-02 | 07 | 4 | OBS-01, OBS-02, OBS-03, OBS-04, NAT-10 | T-15-04 | Full build/vet/suite backs debt and validation closure | integration | `go build ./... && go vet ./internal/compiler/session/... && go test ./... -count=1` | ✅ infrastructure | ⬜ pending |

*Task/plan allocation matches the seven finalized PLAN.md files. Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky.*

---

## Wave 0 Requirements

- [ ] Invocation grammar canonical/non-canonical byte cases: UTF-8, `%`, `/`, `#`, empty fields, and malformed ordinals.
- [ ] Frozen `/0` and `/1` canonical-byte corpus pin before producer changes.
- [ ] Peer structural import guard and named actionable error assertions.
- [ ] Static-table 61/4096/4097 boundary plus preflight-bypass mutation proof.
- [ ] `function.called` projection removal control and strict preorder assertion.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Threat Coverage

| Ref | Threat | Mitigation / evidence |
|-----|--------|-----------------------|
| T-15-01 | Exponential path-table expansion DoS | Checked/saturating preflight; 4096 accepted, 4097 named refusal, no override; mutation-kill bypass. |
| T-15-02 | Ambiguous delimiter or non-canonical path | One strict byte escape and parser; reject alternate spellings. |
| T-15-03 | Forged/reparented causal edge | Independent peer checks membership, binding, ownership, and preorder. |
| T-15-04 | Silent evidence weakening | Frozen legacy bytes, fail-closed comparator routing, flipped collision control, bidirectional seeded faults. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency distributions recorded honestly
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
