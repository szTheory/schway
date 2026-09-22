---
phase: 15-event-identity-lang-execution-2
verified: 2026-09-22T17:10:15Z
status: passed
score: 5/5 must-haves verified
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
covered_digest: "v1:sha256:83b4ad62a4679530133f076088babf5c093de5f263c7490af10ae7fa331ba539"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 5/5
  gaps_closed: []
  gaps_remaining: []
  regressions: []
human_verification: []
---

# Phase 15: Event Identity (`lang.execution/2`) Verification Report

**Phase Goal:** Two activations of the same callee through a shared-leaf diamond are distinguishable, and the causal edge between caller and callee is observed rather than inferred.
**Verified:** 2026-09-22T17:10:15Z
**Status:** passed
**Re-verification:** Yes — refreshed stale covered-input fingerprint with independent focused evidence.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | `multi_function_diamond_call.lang` runs on interpreter, O0, O3, and O3-LTO with distinct activation identities, and the restored collision fails. | ✓ VERIFIED | `TestPhase11InterproceduralDifferential/DiamondSharedLeaf`, `TestPhase15DiamondFrontierMoved`, and `TestPhase15CollisionGuardIsNotInert` passed. The latter appends a real duplicate `(Invocation, ID)` pair and requires the independent peer's `executionpeer.duplicate_pair` refusal. |
| 2 | An admitted `OpCall` produces one caller-owned, preorder `function.called` event; an unresolved or depth-refused call produces none. | ✓ VERIFIED | `interp.go` appends `calledEvent(top, operation, result.calleeID)` only after successful call partitioning and before child-frame push. Interpreter and native preorder/removal/refusal controls passed. |
| 3 | A non-importing peer independently derives observed invocation/causal structure and rejects producer and peer-boundary corruption. | ✓ VERIFIED | `executionpeer` imports only `core` and public `execution`, builds its own function/operation index, parses public invocation grammar, and validates its own stack/preorder state. Import-boundary, actionable-refusal, producer-fault, peer-acceptance-fault, and peer-required controls passed. |
| 4 | `/0` and `/1` bytes stay frozen; only `/2` carries invocation identity and uses pair uniqueness. | ✓ VERIFIED | `schemaForProgram` selects `/2` only for multi-function programs. Legacy-byte, schema-selection, wrapper-preservation, strict validation, and serialized-admission controls passed. |
| 5 | Native occurrence expansion is measured and bounded: the real deep diamond is 61 nodes, 4096 is accepted, and the attempted 4097th fails closed before emission. | ✓ VERIFIED | `preflightInvocationPathTable` precedes path-table construction and C serialization; it returns `cgen.invocation_path_table_exceeded`. Measurement, ordering, boundary, and non-inertness controls passed. |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/compiler/execution/invocation.go` | Canonical `/2` invocation grammar | ✓ VERIFIED | Non-stub formatter/parser rejects noncanonical escaping and ordinal forms; grammar and native admission tests exercise it. |
| `internal/compiler/interp/interp.go` | Frame-local occurrences and observed caller-owned edges | ✓ VERIFIED | Schema selection, invocation threading, and `calledEvent` are on real interpreter execution paths exercised by success and refusal tests. |
| `internal/compiler/executionpeer/executionpeer.go` | Independent causal peer | ✓ VERIFIED | Substantive traversal/index/stack validation uses public model plus `core`; no interpreter or cgen import is present. |
| `internal/compiler/cgen/cgen_program.go`, `cgen.go` | Native occurrence table, call edge writer, and bounded preflight | ✓ VERIFIED | Preflight, parent-indexed lookup, threaded indices, literal invocation table, and C event writer are exercised by cgen and four-tier tests. |
| `internal/compiler/session/session_phase5_compare.go` | `/2` peer gate before comparison | ✓ VERIFIED | Focused comparator tests prove exhaustive field routing, peer-required rejection, and preserved legacy bypass. |
| Differential/frontier tests | Four-tier positive diamond and live negative controls | ✓ VERIFIED | The tracked fixture is run through all four tiers and actual documents/refusals are inspected. |
| `internal/compiler/native/native.go`, `native_test.go` | Strict serialized Schema 2 admission | ✓ VERIFIED | Canonical bytes reach `decodeExecution`; malformed invocation, duplicate pair, unknown kind, and caller-ownership faults are asserted as `native.invalid_execution`. |
| `.github/workflows/ci.yml`, `session_phase6_test.go` | Durable cross-platform aggregate | ✓ VERIFIED | Source-pinning test passed and the workflow names all Phase 15 seams on Ubuntu and macOS. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| Interpreter `OpCall` | Public invocation/call-edge model | `execution.FormatInvocation` during frame construction; `calledEvent` before child push | ✓ WIRED | Direct source trace and interpreter behavioral tests confirm caller ownership and order. |
| Native `emitProgram` | Preflight and C writer | validated entry → preflight → path table → literal C emission | ✓ WIRED | Source order is explicit; ordering and boundary tests pass. |
| `/2` comparator | `executionpeer.Validate` | each `/2` document is peer-validated before pair agreement | ✓ WIRED | `TestSchema2ComparisonRequiresPeerVerdict` passes with a real corrupted document. |
| Four-tier driver | tracked shared-leaf fixture | `DiamondSharedLeaf` loads `multi_function_diamond_call.lang` and runs all four tiers | ✓ WIRED | Differential, moved-frontier, and collision controls passed. |
| CI aggregate | focused Phase 15 evidence | named Go-test regexp in both-host workflow job | ✓ WIRED | `TestCIWorkflowRunsCurrentAggregateGate` passed. |

### Data-Flow Trace (Level 4)

No rendered UI/data artifact exists in this compiler phase. The relevant runtime path is nevertheless exercised end to end: checked fixture → interpreter/native execution document → independent peer/comparator → assertion. No static or hollow rendering path applies.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Grammar, admission, interpreter, peer, native preflight/writer, comparator, diamond, collision, legacy, and CI seams | `go test ./internal/compiler/execution ./internal/compiler/native ./internal/compiler/interp ./internal/compiler/executionpeer ./internal/compiler/cgen ./internal/compiler/session -run 'TestInvocationGrammar|TestExecutionLegacyBytesFrozen|TestValidateExecutionSchema2|TestDecodeExecutionSchema2AdmissionSeam|TestInvocationThreadsThroughAllEventPaths|TestExecutionSchemaSelectionPreservesLegacy|TestFunctionCalledPreorderAndOwnership|TestFunctionCalledProjectionRemoval|TestRejectedCallEmitsNoCalledEdge|TestIndependentEntryResolution|TestInvocationMembershipTraversal|TestExecutionPeerImportBoundary|TestFullCoverageControlIsSeparate|TestValidateObservedCausalStructure|TestExecutionPeerFailuresAreActionable|TestInvocationPathTableDeepDiamondMeasures61|TestInvocationPreflightOrdering|TestInvocationPathTableBoundary|TestInvocationPreflightGuardIsNotInert|TestProgramInvocationIndexThreading|TestParentIndexedChildLookup|TestInvocationTableEmissionIsDeterministic|TestProgramWritesExecutionSchema2|TestNativeFunctionCalledPreorder|TestNativeFunctionCalledProjectionRemoval|TestLegacyEventWritersFrozen|TestComparisonFieldRoutingIsExhaustive|TestInvocationFieldsAreCompared|TestSchema2ComparisonRequiresPeerVerdict|TestExecutionProducerFaultIsCaughtByPeer|TestExecutionPeerAcceptanceFaultIsCaughtByControl|TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert|TestPhase15DiamondFrontierMoved|TestPhase5CompareProgramEnginesPreservesLegacySchemas|TestCIWorkflowRunsCurrentAggregateGate' -count=1 -v` | All selected top-level tests and subtests passed; the slowest selected package was 5.845s. | ✓ PASS |

### Probe Execution

No Phase 15 plan or summary declares a probe, and no `scripts/*/tests/probe-*.sh` file was discovered. **SKIPPED (no declared probe).**

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| OBS-01 | 15-01, 02, 05, 07, 09 | Shared-leaf activations have distinct identities. | ✓ SATISFIED | Four-tier diamond, moved-frontier, and restored collision-negative tests passed. |
| OBS-02 | 15-01, 02, 05, 06, 07, 09 | Caller-to-callee edge is observed. | ✓ SATISFIED | Interpreter/native preorder/ownership/removal tests and peer validation passed. |
| OBS-03 | 15-03, 06, 07, 09 | Non-importing peer re-derives structure. | ✓ SATISFIED | Import boundary, independent traversal, actionable failures, and both fault-direction controls passed. |
| OBS-04 | 15-01 through 15-09 | Legacy wire bytes frozen; `/2` introduces identity. | ✓ SATISFIED | Legacy-byte, schema-selection, wrapper-preservation, and serialized admission tests passed. |
| NAT-10 | 15-05, 07, 09 | Re-invoking fixture agrees across four tiers. | ✓ SATISFIED | `DiamondSharedLeaf` executed interpreter/O0/O3/O3-LTO successfully. |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `session_phase11_differential_test.go` | OBS-01, NAT-10 | Yes | 0 | No | Behavioral four-tier comparison | Strong |
| `session_phase15_frontier_test.go` | OBS-01, OBS-02, OBS-03 | Yes | 0 | No | Mutation and refusal assertions | Strong |
| Interpreter/peer/cgen tests | OBS-02, OBS-03, OBS-04 | Yes | 0 | No | Behavioral, structural, boundary | Strong |
| Native/comparator/CI tests | OBS-04, NAT-10 | Yes | 0 | No | Serialized-admission and provenance | Strong |

No disabled requirement-only test or circular expected-value generator was found. The differential result is independently produced by the interpreter and three compiled native tiers, then checked by a separately derived peer.

### Anti-Patterns Found

No unreferenced `TBD`, `FIXME`, or `XXX` marker was found in the Phase 15 implementation or requirement-linked tests. The `return []string{}` matches in cgen are non-rendering helper branches with populated paths exercised by boundary and native-output tests. No blocker or warning resulted.

### Decision Coverage

`check.decision-coverage-verify` reported **20/20** trackable Phase 15 context decisions honored; no decision was unhonored. This is advisory and did not determine the status.

### Advisory (New Scope, Unevidenced)

None.

## Human Verification

N/A — this is a compiler/foundation phase with no user-facing element. All acceptance criteria, including state/order invariants, have focused automated behavioral evidence. The current UAT is independently consistent: zero manual UAT items are required.

## Gaps Summary

None. The prior report's input fingerprint was stale; the refreshed digest covers all Phase 15 plans and summaries, the mapped requirements, and the implementation/test/CI files changed by the phase.

---

_Verified: 2026-09-22T17:10:15Z_
_Verifier: the agent (gsd-verifier)_
