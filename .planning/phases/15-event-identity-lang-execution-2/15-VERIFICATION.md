---
phase: 15-event-identity-lang-execution-2
verified: 2026-09-25T21:02:07Z
status: passed
score: 7/7 must-haves verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/phases/15-event-identity-lang-execution-2/15-01-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-01-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-02-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-02-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-03-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-03-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-04-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-04-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-05-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-05-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-06-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-06-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-07-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-07-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-08-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-08-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-09-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-09-SUMMARY.md
  - .planning/phases/15-event-identity-lang-execution-2/15-10-PLAN.md
  - .planning/phases/15-event-identity-lang-execution-2/15-10-SUMMARY.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/execution/invocation.go
  - internal/compiler/execution/invocation_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_test.go
  - internal/compiler/session/session_payload_replay_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase15_frontier_test.go
  - internal/compiler/session/session_phase5_compare.go
  - internal/compiler/session/session_phase5_compare_test.go
  - internal/compiler/session/session_phase6_test.go
covered_digest: "v1:sha256:59e52cf66c7b190537c0a18f09103ef5d14ce2b57f849f5e960bff932d058c71"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 7/7
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  refresh_reason: "The prior covered-file fingerprint was stale. Re-ran current Phase 15 owning-package admission, interpreter, peer, native-emission, CI source-pin, and four-tier diamond/collision evidence. All targeted commands passed; the covered-file fingerprint now matches the current tree."
advisory: []
---

# Phase 15: Event Identity (`lang.execution/2`) Verification Report

**Phase Goal:** Two activations of the same callee through a shared-leaf diamond are distinguishable, and the causal edge between caller and callee is observed rather than inferred.
**Verified:** 2026-09-25T21:02:07Z
**Status:** passed
**Re-verification:** Yes — refreshed the stale fingerprint and reran current Phase 15 automated evidence.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | The tracked shared-leaf diamond succeeds across interpreter, O0, O3, and O3-LTO with distinct activation identities, and the restored collision fails. | ✓ VERIFIED | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v` passed; the diamond subtest and live negative control both executed. |
| 2 | An admitted `OpCall` produces one caller-owned preorder `function.called` event; unresolved or depth-refused calls produce none. | ✓ VERIFIED | The interpreter event-order, projection-removal, and unresolved/depth refusal tests passed in the focused cross-package run recorded below. The four-tier diamond also passed. |
| 3 | An independent non-importing peer re-derives invocation/call structure and catches producer and peer-boundary corruption. | ✓ VERIFIED | `TestExecutionPeerImportBoundary`, `TestValidateObservedCausalStructure`, `TestExecutionPeerFailuresAreActionable`, `TestExecutionProducerFaultIsCaughtByPeer`, and `TestExecutionPeerAcceptanceFaultIsCaughtByControl` are present in their requirement-linked packages; peer and comparator tests passed in the targeted runs. |
| 4 | `/0` and `/1` wire bytes remain frozen; `/2` carries canonical invocation identity and pair uniqueness, and the program-aware wrapper preserves legacy comparison while rejecting invalid `/2`. | ✓ VERIFIED | `TestExecutionLegacyBytesFrozen`, `TestValidateExecutionSchema2`, `TestDecodeExecutionSchema2AdmissionSeam`, `TestPhase5CompareProgramEnginesPreservesLegacySchemas`, and `TestSchema2ComparisonRequiresPeerVerdict` passed in the owning packages. Coverage metadata classifies the relevant plans as fully automated. |
| 5 | Native occurrence expansion is measured at 61 nodes for the real deep diamond, admits 4096, and refuses the 4097th before emission. | ✓ VERIFIED | `TestInvocationPathTableDeepDiamondMeasures61`, `TestInvocationPreflightOrdering`, `TestInvocationPathTableBoundary`, and `TestInvocationPreflightGuardIsNotInert` passed in the focused cgen run. |
| 6 | Cross-platform CI runs all Phase 15 aggregate evidence, including the Schema 2 decoder admission seam. | ✓ VERIFIED | `.github/workflows/ci.yml` runs the native decoder selector under `./internal/compiler/native` and the peer, legacy-wrapper, four-tier diamond, and collision selectors under `./internal/compiler/session`; both remain in the Ubuntu/macOS `evidence-aggregate` matrix. The package-aware source pin rejects the synthetic wrong-package selector. The owning-package native test, the source-pin tests, and all four session evidence seams passed in this verification. |
| 7 | Plan coverage metadata classifies the deterministic admission, peer, legacy-wrapper, diamond, and collision evidence as automated, removing the prior human-UAT fallback. | ✓ VERIFIED | Plan 10 coverage metadata declares both package-correct CI selectors and the package-mismatch negative control automated with `human_judgment: false`. [15-UAT.md](15-UAT.md) remains complete with 12/12 automated checks, 0 issues, 0 pending, 0 skipped; historical gaps G-15-1, G-15-11, and G-15-12 are resolved by Plans 15-08/15-09. No human UAT is required. |

**Score:** 7/7 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/compiler/execution/invocation.go` | Canonical `/2` invocation formatter/parser | ✓ VERIFIED | Substantive grammar implementation; `TestInvocationGrammar` passed. |
| `internal/compiler/interp/interp.go` | Frame-local occurrences and caller-owned call edges | ✓ VERIFIED | Wired into frame construction and `OpCall`; event path, preorder, removal, and refusal tests passed. |
| `internal/compiler/executionpeer/executionpeer.go` | Independent causal peer | ✓ VERIFIED | Substantive traversal and observed-structure validation; import-boundary and refusal tests passed. |
| `internal/compiler/cgen/cgen_program.go`, `cgen.go` | Native invocation table, event writer, and bounded preflight | ✓ VERIFIED | Preflight and C serialization are wired; measured/boundary and native event tests passed. |
| `internal/compiler/native/native.go`, `native_test.go` | Strict Schema 2 serialized admission | ✓ VERIFIED | JSON-to-ToolError seam passes when run in the owning package; CI aggregate package selection is separately tracked as a gap. |
| `internal/compiler/session/session_phase5_compare.go` | `/2` peer validation before comparison; legacy behavior preserved | ✓ VERIFIED | Peer-gate and `/0`/`/1` wrapper regressions passed. |
| `.github/workflows/ci.yml` | Both-host durable aggregate for the phase seams | ✓ VERIFIED | Separate focused commands run the native decoder admission test and the four session seams from their owning packages. The Ubuntu/macOS matrix, historical Phase 6 baseline, and existing `checks` job remain. |
| `internal/compiler/session/session_phase6_test.go` | Package-aware source pin for CI aggregate contract | ✓ VERIFIED | `TestCIWorkflowRunsCurrentAggregateGate` and `TestCIWorkflowSelectionPinsPackageOwnership` passed, including the synthetic wrong-package rejection case. |
| `15-10-PLAN.md`, `15-10-SUMMARY.md` | Closure evidence for G-15-13 | ✓ VERIFIED | Plan declares the package ownership contract and automated checks; implementation and named tests were inspected and run directly. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| Interpreter `OpCall` | Invocation/call-edge model | Canonical formatter and caller-owned event before child push | ✓ WIRED | Source paths and behavior tests agree. |
| Native emitter | Preflight and C writer | Graph/entry validation → preflight → path table → serialization | ✓ WIRED | Ordering, measured boundary, and emission tests passed. |
| `/2` comparator | Independent execution peer | Validate each `/2` document before comparison | ✓ WIRED | `TestSchema2ComparisonRequiresPeerVerdict` and peer-fault controls passed. |
| Four-tier differential | Tracked shared-leaf fixture | `DiamondSharedLeaf` runs all four tiers and compares evidence | ✓ WIRED | Focused subtest passed. |
| CI aggregate | Legacy wrapper and diamond tests | Focused session-package command | ✓ WIRED | Wrapper, diamond, and collision tests executed successfully. |
| CI aggregate | Native Schema 2 admission seam | Focused package test invocation | ✓ WIRED | Workflow runs `go test ./internal/compiler/native -run 'TestDecodeExecutionSchema2AdmissionSeam' -count=1 -v`; confirmed by the current source and passing owning-package test. |
| CI source pin | Expected package ownership | Table-driven positive/negative source fixtures | ✓ WIRED | The source pin accepts the native package selection and rejects a session-package command for the native test; both named tests passed. |

### Data-Flow Trace (Level 4)

No rendered UI/data artifact exists in this compiler phase. The runtime path is exercised as fixture/source → interpreter/native execution document → independent peer/comparator → assertions. The Schema 2 decoder path reaches real JSON decode and validation when invoked directly, and the workflow selects it from its owning `internal/compiler/native` package.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Schema 2 JSON-to-ToolError seam in its CI owning package | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/native -run 'TestDecodeExecutionSchema2AdmissionSeam' -count=1 -v` | Rerun exit 0; canonical admission and all five refusal cases passed. | ✓ PASS |
| CI source-pin positive/negative package ownership | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestCIWorkflowSelectionPinsPackageOwnership|TestCIWorkflowRunsCurrentAggregateGate' -count=1 -v` | Rerun exit 0; correct package accepted, wrong-package selector rejected, aggregate source contract passed. | ✓ PASS |
| Cross-platform session evidence selectors | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestSchema2ComparisonRequiresPeerVerdict|TestPhase5CompareProgramEnginesPreservesLegacySchemas|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v` | Rerun exit 0; peer, `/0`/`/1` wrapper, four-tier diamond, and collision tests passed. | ✓ PASS |
| Cross-package grammar, admission, interpreter, peer, native-boundary, and cgen controls | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/execution ./internal/compiler/native ./internal/compiler/interp ./internal/compiler/executionpeer ./internal/compiler/cgen -run 'TestInvocationGrammar|TestExecutionLegacyBytesFrozen|TestValidateExecutionSchema2|TestInvocationThreadsThroughAllEventPaths|TestExecutionSchemaSelectionPreservesLegacy|TestFunctionCalledPreorderAndOwnership|TestFunctionCalledProjectionRemoval|TestRejectedCallEmitsNoCalledEdge|TestIndependentEntryResolution|TestInvocationMembershipTraversal|TestExecutionPeerImportBoundary|TestFullCoverageControlIsSeparate|TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable|TestInvocationPathTableDeepDiamondMeasures61|TestInvocationPreflightOrdering|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestParentIndexedChildLookup|TestInvocationTableEmissionIsDeterministic|TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder|TestNativeFunctionCalledProjectionRemoval|TestLegacyEventWritersFrozen' -count=1 -v` | Rerun exit 0; selected tests passed in all five packages. | ✓ PASS |
| Workflow source pin and diamond/collision controls | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestCIWorkflowRunsCurrentAggregateGate|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert' -count=1 -v` | Rerun exit 0; all named tests and diamond subtest passed. | ✓ PASS |

### Probe Execution

No Phase 15 plan or summary declares a probe, and no phase-relevant `scripts/*/tests/probe-*.sh` was identified. **SKIPPED (no declared probe).**

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| OBS-01 | 15-01, 15-02, 15-05, 15-07, 15-09 | Shared-leaf activations have distinct identities. | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; current four-tier diamond and collision-negative tests passed. |
| OBS-02 | 15-01, 15-02, 15-05, 15-06, 15-07, 15-09 | Caller-to-callee edge is observed. | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; interpreter/native preorder and removal/refusal tests passed; peer validation passed. |
| OBS-03 | 15-03, 15-06, 15-07, 15-09 | Non-importing peer re-derives structure. | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; import-boundary, independent traversal, actionable errors, and bidirectional fault tests passed. |
| OBS-04 | 15-01 through 15-10 | Legacy bytes frozen; `/2` introduces identity. | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; frozen bytes, schema admission, wrapper preservation, and package-correct recurring admission selection passed. |
| NAT-10 | 15-05, 15-07, 15-09, 15-10 | Re-invoking fixture agrees across four tiers. | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; `DiamondSharedLeaf` and collision control passed; CI keeps both selected in the session package and package ownership is pinned. |

No additional requirement IDs mapped to Phase 15 were orphaned from the plans.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `session_phase11_differential_test.go` | OBS-01, NAT-10 | Yes | 0 | No | Four-tier behavioral comparison | Strong |
| Interpreter/peer/cgen tests | OBS-02, OBS-03, OBS-04 | Yes | 0 | No | Behavioral, structural, and boundary assertions | Strong |
| `native_test.go` | OBS-04 | Yes | 0 | No | JSON decode to actionable ToolError | Strong; run in native package by the recurring aggregate |
| Comparator and CI source-pin tests | OBS-03, OBS-04, NAT-10 | Yes | 0 | No | Peer refusal, comparator behavior, workflow-source contract | Strong; source pin rejects package mismatch |
| `session_phase6_test.go` | OBS-04, NAT-10 | Yes | 0 | No | Package ownership positive/negative controls and workflow contract | Strong; both named tests passed |

No disabled requirement-only test or circular expected-value generator was found in the targeted evidence. The decoder test has independent refusal cases and is not circular; the failure is in recurring CI invocation, not test quality.

### Anti-Patterns Found

No unresolved `TBD`, `FIXME`, or `XXX` debt marker was found in the Phase 15 implementation files or requirement-linked tests. Empty-slice return matches in cgen are bounded/error helper branches; exercised success and overflow tests confirm they are not user-visible stubs. No separate code stub blocker was found.

### Decision Coverage

`check.decision-coverage-verify` reported **20/20** trackable Phase 15 context decisions honored; no decision was unhonored. This gate is advisory and does not affect the status.

### Fingerprint Refresh

The previous digest (`v1:sha256:043f8000e9d67cebe523cf86672259030acd81bf399ff5de6e729f876542158f`) recomputed to `v1:sha256:59e52cf66c7b190537c0a18f09103ef5d14ce2b57f849f5e960bff932d058c71` for the complete covered-file set. The owning-package Schema 2 decoder seam, CI ownership/source pins, four-tier diamond and collision controls, frozen `/0` and `/1` behavior, interpreter call-edge/refusal tests, independent peer checks, and native invocation-table boundary/emission checks all passed on the current tree. Requirements remain complete; no requirement or STATE edits were needed.

### Advisory (New Scope, Unevidenced)

None. No new-scope anti-pattern finding required advisory treatment in this re-verification.

### Human Verification Required

None. All phase truths, including CI package ownership, have deterministic automated evidence; this compiler phase has no irreducible visual, interactive, or external-service criterion.

### Gaps Summary

The implementation truths, including the four-tier diamond, caller-owned causal events, independent peer, frozen legacy bytes, and measured native invocation bound, are verified. Plan 15-10 closes G-15-13: CI now selects the Schema 2 decoder seam from `./internal/compiler/native`, selects the peer, legacy-wrapper, diamond, and collision seams from `./internal/compiler/session`, and retains both host priorities and the Phase 6 baseline. The CI source pin has a passing positive case and a negative synthetic wrong-package control; both were run and passed. The UAT remains complete at 12/12 automated checks, with no pending or human verification items.

---

_Verified: 2026-09-25T21:02:07Z_
_Verifier: the agent (gsd-verifier)_
