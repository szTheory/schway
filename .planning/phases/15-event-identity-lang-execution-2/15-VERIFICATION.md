---
phase: 15-event-identity-lang-execution-2
verified: 2026-09-19T20:26:08Z
status: passed
score: 5/5 must-haves verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/debug/resolved/phase15-lto-diagnostic-order.md
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
  - .planning/phases/15-event-identity-lang-execution-2/15-REVIEW.md
  - .planning/phases/15-event-identity-lang-execution-2/15-VALIDATION.md
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/execution/invocation.go
  - internal/compiler/execution/invocation_test.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_test.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase15_frontier_test.go
  - internal/compiler/session/session_phase5_compare.go
  - internal/compiler/session/session_phase5_compare_test.go
  - internal/compiler/session/witness_registry_test.go
covered_digest: "v1:sha256:327117c580f56d50cf681ea0ee8a459c31e473ede2dc30b550ff62aaaa99fe5d"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 5/5
  gaps_closed:
    - "Repository-wide automated acceptance remains green after Phase 15 native-emitter changes."
  gaps_remaining: []
  regressions: []
---

# Phase 15: Event Identity (`lang.execution/2`) Verification Report

**Phase Goal:** Two activations of the same callee through a shared-leaf diamond are distinguishable, and the causal edge between caller and callee is observed rather than inferred.
**Verified:** 2026-09-19T20:26:08Z
**Status:** passed
**Re-verification:** Yes — after LTO diagnostic-order gap closure

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | The shared-leaf diamond is distinct and agrees across interpreter, O0, O3, and O3-LTO. | ✓ VERIFIED | `DiamondSharedLeaf`, `TestPhase15DiamondFrontierMoved`, and `TestPhase15CollisionGuardIsNotInert` passed. The positive gate runs the checked fixture on four real tiers, rejects restored duplicate `(invocation, id)` pairs, and peer-gates all pairwise comparisons. |
| 2 | `OpCall` emits an observed, caller-owned causal edge; removing or corrupting it is caught. | ✓ VERIFIED | `calledEvent` records the caller invocation and `CalleeFunctionID` before child execution; native emission does the same. Focused edge, projection, and mutation controls passed. |
| 3 | A non-importing peer independently re-derives occurrence membership and rejects producer and peer-boundary faults. | ✓ VERIFIED | `executionpeer.Validate` reconstructs traversal from `core` plus public execution data; producer-corruption, peer-acceptance, import-boundary, and peer-required comparison controls pass. |
| 4 | `/0` and `/1` stay frozen while `/2` carries invocation identity. | ✓ VERIFIED | `schemaForProgram` selects `/2` only for multi-function programs; grammar and legacy-byte/frozen-writer tests pass. |
| 5 | Native occurrence expansion is measured and bounded, with usable output capacity and preserved refusal precedence. | ✓ VERIFIED | The 61-node real fixture, 4096/4097 boundary, preflight mutation, exact output N-1/N, and O0/O3 execution tests pass. `TestUnsupportedProgramShapePrecedesSchema2Preflight` and the original LTO witness prove unsupported `Match` bodies retain their named refusal before size estimation. |

**Score:** 5/5 truths verified (0 present, behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `internal/compiler/execution/invocation.go` | Canonical invocation grammar | ✓ VERIFIED | Strict formatter/parser and invalid-spelling tests are substantive and used by all `/2` producers/peer. |
| `internal/compiler/interp/interp.go` | Frame-local occurrences and call edges | ✓ VERIFIED | `entryIdentity`, child configuration, and `calledEvent` wire canonical identities through interpreter event paths. |
| `internal/compiler/executionpeer/executionpeer.go` | Independent causal peer | ✓ VERIFIED | Imports public/core model rather than interpreter or cgen, and validates exact observed `/2` structure. |
| `internal/compiler/cgen/cgen_program.go` and `cgen.go` | Native occurrence tables, causal writer, and bounded preflight | ✓ VERIFIED | Supported body shapes are collected before invocation/event/output preflight; static parent-indexed table and C writer remain live after the ordering repair. |
| `internal/compiler/session/session_phase5_compare.go` | `/2` peer gate before comparison | ✓ VERIFIED | Every schema-2 document reaches `executionpeer.Validate` before pair comparison. |
| `internal/compiler/session/session_phase11_differential_test.go` | Four-tier diamond gate | ✓ VERIFIED | Requires all four documents, pair uniqueness, and peer-gated all-pairs comparison for the permanent fixture. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| Interpreter frames and `OpCall` | Invocation grammar | ✓ WIRED | Frame creation/child setup use `execution.FormatInvocation`; call edge has caller ownership. |
| Native shape validation and preflight | Native C writer | ✓ WIRED | `emitProgram` validates ordered supported functions, then preserves bounded occurrence, event-capacity, and exact-size checks before serialization. |
| Four-tier driver | Independent peer | ✓ WIRED | `Phase5CompareProgramEngines` validates each `/2` engine result before comparing pairs. |
| Diamond test | Tracked fixture | ✓ WIRED | `DiamondSharedLeaf` executes `multi_function_diamond_call.lang` on interpreter, O0, O3, and O3-LTO. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| LTO diagnostic precedence plus native capacity/output safeguards | Focused cgen regression/boundary run | All named tests passed. | ✓ PASS |
| Four-tier occurrence identity and collision guard | `go test ./internal/compiler/session -run 'TestPhase11InterproceduralDifferential/DiamondSharedLeaf|TestPhase15CollisionGuardIsNotInert|TestPhase15DiamondFrontierMoved' -count=1 -v` | All matching named tests passed. | ✓ PASS |
| Repository automated acceptance | `go build ./... && go vet ./... && go test ./... -count=1` | Completed successfully after the focused checks; no package failure observed. | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
| --- | --- | --- | --- |
| OBS-01 | 15-01, 15-02, 15-05, 15-07 | ✓ SATISFIED | Four-tier diamond acceptance and restored-collision negative prove occurrence-specific identity. |
| OBS-02 | 15-01, 15-02, 15-05, 15-06 | ✓ SATISFIED | Caller-owned `function.called` appears across engines; focused corrupt/remove controls exercise the edge. |
| OBS-03 | 15-03, 15-06 | ✓ SATISFIED | Independent peer traversal and two-direction fault controls pass. |
| OBS-04 | 15-01 through 15-06 | ✓ SATISFIED | Frozen `/0`/`/1` bytes and `/2`-only invocation grammar are covered by active tests. |
| NAT-10 | 15-05, 15-07 | ✓ SATISFIED | The re-invoking multi-function diamond agrees on interpreter, O0, O3, and O3-LTO. |

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `session_phase11_differential_test.go` | OBS-01, NAT-10 | Yes | 0 | No | Behavioral four-tier and pairwise comparisons | Strong |
| `session_phase15_frontier_test.go` | OBS-01, OBS-02, OBS-03 | Yes | 0 | No | Behavioral/mutation | Strong |
| `cgen_program_test.go` | OBS-04 | Yes | 0 | No | Boundary, refusal precedence, and native execution | Strong |

No disabled requirement-only test or circular expected-value generator was found. Differential expected values come from independently executed engines and the independently-derived peer.

### Anti-Patterns Found

No Phase 15 debt markers (`TBD`, `FIXME`, or `XXX`) were found in the changed compiler implementation. The previous deterministic failure is closed by `b55c1fc`: the emitter validates supported structural shapes before schema-2 estimation, and the direct regression plus original LTO witness pass.

### Decision Coverage

All 20 trackable Phase 15 context decisions are honored by shipped artifacts.

## Human Verification

N/A — compiler/foundation phase. Per the user's explicit automated-acceptance policy, integration, differential, mutation, boundary, and uncached repository-suite evidence are the acceptance gate; no manual UAT is required.

## Gaps Summary

None. The prior LTO diagnostic-order regression is closed without weakening capacity or output safeguards, and all five Phase 15 requirements have deterministic automated evidence.

---

_Verified: 2026-09-19T20:26:08Z_
_Verifier: the agent (gsd-verifier)_
