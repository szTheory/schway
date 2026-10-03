---
phase: 26-checked-scalar-sum
verified: 2026-10-03T11:23:18Z
status: passed
score: 5/5 roadmap truths verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/phases/26-checked-scalar-sum/26-01-PLAN.md
  - .planning/phases/26-checked-scalar-sum/26-01-SUMMARY.md
  - .planning/phases/26-checked-scalar-sum/26-02-PLAN.md
  - .planning/phases/26-checked-scalar-sum/26-02-SUMMARY.md
  - .planning/phases/26-checked-scalar-sum/26-03-PLAN.md
  - .planning/phases/26-checked-scalar-sum/26-03-SUMMARY.md
  - .planning/phases/26-checked-scalar-sum/26-04-PLAN.md
  - .planning/phases/26-checked-scalar-sum/26-04-SUMMARY.md
  - .planning/phases/26-checked-scalar-sum/26-05-PLAN.md
  - .planning/phases/26-checked-scalar-sum/26-05-SUMMARY.md
  - cmd/schway/main_test.go
  - examples/phase26/checked_add_overflow.schway
  - examples/sum_to_n.schway
  - internal/compiler/ability/ability.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/executionpeer/executionpeer.go
  - internal/compiler/executionpeer/executionpeer_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/session/session_phase26_event_test.go
  - internal/compiler/session/session_phase26_overflow_test.go
  - internal/compiler/session/session_phase26_test.go
  - internal/compiler/session/session_phase5_compare.go
  - internal/compiler/session/session_phase5_compare_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - testdata/phase16/public-emitter-consumers.json
covered_digest: "v2:sha256:1828516bd9babb893eea41ce430d065d04efaf419b3e95b7d369219a5dfd5407"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 10
  total: 10
  not_honored: []
---

# Phase 26: Checked Scalar Sum Verification Report

**Phase Goal:** A developer can run a source-authored scalar loop that computes and prints the exact bounded `sum_to_n` result, with defined overflow and explicit refusal of unsafe loop-carried state.
**Verified:** 2026-10-03T11:23:18Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | Public application route builds and runs ordinary source with inputs 0, 10, and 1,000 producing exactly `0\n`, `55\n`, and `500500\n`. | ✓ VERIFIED | `TestPhase26PublicAppCLI` asserts each exact stdout/stderr/exit tuple; `TestPhase26ExactSumMatrix` separately runs the retained app and interpreter against literal values. Focused hosted run [37118144404](https://github.com/szTheory/schway/actions/runs/37118144404) passed on Ubuntu and macOS. |
| 2 | U64/Bool comparison, `if/else`, and predicate-controlled scalar loops are admitted consistently by checker, independent validators, interpreter, and native execution. | ✓ VERIFIED | `examples/sum_to_n.schway` is parsed and lowered to explicit typed CFG operations. `TestPhase26FixedPoint`, `TestPhase26PeerIndependence`, scalar-copy cycle controls, `TestPhase26ExactSumMatrix`, and `TestPhase26SourceRepeatedCopyEvidence` exercise admission and both engines. The focused hosted run passed both hosts. |
| 3 | In-range U64 addition is exact; interpreter and C17 fail on reached overflow before a wrapped value succeeds. | ✓ VERIFIED | `TestPhase26CheckedAddInterpreter`, `TestPhase26CheckedAddC17`, and `TestPhase26OverflowProcessOutcome` pin MAX+0 and MAX+1. The public CLI test pins empty stdout, exit 65, and exact `schway: U64 addition overflow\n` stderr. Focused hosted run passed both hosts. |
| 4 | Ownership, resource, loan, and loan-derived provenance across a loop back edge are refused with stable source attribution; bounded analysis fails closed. | ✓ VERIFIED | Checker tests cover refusal categories and forced budget exhaustion; core and origin validators each have independent `TestPhase26AnalysisExhaustion` controls and derive their own scalar facts. `TestPhase26BackEdgeAuthority` and cycle-boundary mutations check peer refusal. Focused hosted run passed both hosts. |
| 5 | Each engine is checked against pinned answers and reached wrong-result/skipped-iteration controls; input 1,001 fails before app output. | ✓ VERIFIED | `TestPhase26ExactSumMatrix` embeds expected `0`, `55`, and `500500` literals for interpreter and native, runs two source mutations, and pins 1,001 to exit 65 with empty stdout. `TestPhase26PublicAppCLI` checks the caller-visible tuples. Focused hosted run passed both hosts. |

**Score:** 5/5 roadmap truths verified (0 present, behavior-unverified).

The plan's phrase “complete event evidence, `Verified=true`” in 26-05-PLAN does not match the existing app-evidence API contract. `ApplicationEvidenceReport` documents that capture records structure and identity and is never itself a differential semantic verdict; `RunApplicationWithEvidence` leaves its `Verified` field false. This is intentional and consistent with the locked context decisions to keep app outcomes separate from evidence capture and to compare each engine using independent validation. `TestPhase26SourceRepeatedCopyEvidence` independently calls `executionpeer.Validate` on both interpreter-projected and native-captured evidence and confirms the 0/1/2 occurrence sequence. Thus the locked requirement and phase goal are met by the separate peer verdict; the report boolean must remain false. Recommend correcting the plan wording in a dated planning amendment so it says the complete capture is independently peer-validated, while capacity exhaustion stays unverified and preserves the child result.

### Decision Coverage

The decision-coverage query reports all 10 trackable CONTEXT decisions honored, with no unhonored items. The decisions are reflected in the source form, scalar-only CFG admission, separate overflow/app failure channels, independent analysis and peer checks, repeated event identity, and retained toolchain.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/sum_to_n.schway` | Ordinary bounded scalar-loop witness | ✓ VERIFIED | Substantive source declares mutable scalar state, input guard, pre-tested loop, repeated scalar copy, and return value. It is built through the public application route in hosted acceptance tests. |
| `examples/phase26/checked_add_overflow.schway` | Reached direct overflow witness | ✓ VERIFIED | Exercises MAX+0/MAX+1 branches through interpreter and native tests. |
| Parser/AST/core/checker and independent validators | Typed scalar syntax and bounded CFG admission/refusal | ✓ VERIFIED | Source constructs lower through typed operations and CFG edges; checker, core validator, and origin validator each contain independent scalar transfer analyses and budget refusals. |
| Interpreter, C17 emitter, app runner, execution peer | Exact execution, output, checked overflow, dynamic occurrence evidence | ✓ VERIFIED | Producers are reached from checked source; app output tests invoke retained artifacts; peer derives event sequence/membership and tests forged controls. |
| Phase 26 test homes | Requirements-linked acceptance and negative controls | ✓ VERIFIED | Named tests assert literal values, streams/status, fixed-point/refusal behavior, overflow, occurrence order, and capacity outcome. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Source token/CST span | Typed scalar CFG | Parser, AST lowering, checker | ✓ WIRED | `sum_to_n.schway` uses the new `var`, assignment, branch, loop, `<`, and `+` forms; syntax tests check formatting, malformed recovery, and spans. |
| Typed core CFG | Independent validation | Core and origin peer derivation | ✓ WIRED | `session.Check` outputs are revalidated by both peers; peer-independence and mutation tests cover altered edges/facts. |
| Checked program | Interpreter | `interp.Run` | ✓ WIRED | Exact matrix and overflow tests invoke the interpreter directly on checked source. |
| Checked program | Native app | C17 emission, `BuildApplicationFile`, `RunApplication` | ✓ WIRED | Public CLI and session tests build the checked-in source to retained artifacts and assert output/status. |
| Captured events | Independent event verdict | `executionpeer.Validate` | ✓ WIRED | Source event test validates both projected interpreter events and native captured events, including repeated copy ordinals. Capture does not self-assert semantic verification. |
| App defect | Caller-visible failure | C app shell and `RunOutcome` | ✓ WIRED | Overflow and rejected-input tests verify nonzero exit and no stdout; tool and evidence-capacity errors remain separate. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| `sum_to_n` application | U64 return and stdout | Parsed source input → checked CFG → interpreter/C17 execution → retained app writer | Yes; input matrix and exact caller-visible output are asserted | ✓ FLOWING |
| Native event capture | Execution document | Runtime event recorder → bounded capture file → strict decoder → independent peer | Yes; source-generated events and 0/1/2 ordinals are checked | ✓ FLOWING |
| Capacity outcome | Process result and capture status | Child process outcome plus bounded capture result | Yes; forced capacity test retains `6\n` and exit 0 while marking capture exhausted and `Verified=false` | ✓ FLOWING |

### Behavioral Spot-Checks

Local Go tests and native program runs were not executed, per the project’s hosted-only execution policy. Hosted checks are the behavioral evidence:

| Behavior | Hosted command/evidence | Result | Status |
|---|---|---|---|
| Phase 26 checker, peer, path, and native regressions | [Run 37118144404](https://github.com/szTheory/schway/actions/runs/37118144404), both host jobs | Passed | ✓ PASS |
| Full Go suite, race suite, build, vet, and current evidence aggregates | [Run 37118315516](https://github.com/szTheory/schway/actions/runs/37118315516), Ubuntu and macOS checks and aggregates | Passed | ✓ PASS |
| 35-pair validation corpus | Historical receipt [Run 37098702559](https://github.com/szTheory/schway/actions/runs/37098702559) | Previously passed; not used as current semantic proof | ℹ️ HISTORICAL |

Both decisive current receipts target source revision `46bee44ad87891d8a8547b2489f029d0fe237899`. Read-only git comparison confirms no implementation, test, example, or CI files changed between that revision and current HEAD `34b76c199396e434135fb474107fb71021713ff2`; current HEAD adds planning records only. The latest full run’s Phase 26-specific matrix job was skipped, but its full Go and race suites passed; the preceding focused Phase 26 run passed the targeted checks on both hosts.

### Probe Execution

Not applicable. Phase 26 is not a migration/tooling phase, its roadmap criteria do not declare probes, and no probe path is declared by its plans.

### Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| U64-01 | 26-01, 26-02, 26-05 | Exact U64 addition and defined checked overflow | ✓ SATISFIED | Interpreter/C17 near-MAX tests and exact app failure tuple. |
| FLOW-01 | 26-01, 26-03, 26-04, 26-05 | Bool branches and scalar loops with bounded independent fixed-point derivation | ✓ SATISFIED | Checker and both peers’ fixed-point tests; exact source loop execution in both engines. |
| FLOW-02 | 26-01, 26-03, 26-04, 26-05 | Unsafe authority carries refused and analysis bounds fail closed | ✓ SATISFIED | Category refusal/mutation controls and independent forced-exhaustion tests. |
| APP-07 | 26-01, 26-02, 26-05 | Public ordinary-source sum values | ✓ SATISFIED | Exact public CLI and retained-artifact matrix tests. |

No additional Phase 26 requirements are mapped to this phase without a plan.

### Test Quality Audit

| Test files | Linked requirements | Active | Disabled | Circular expected values | Assertion strength | Verdict |
|---|---|---:|---:|---:|---|---|
| `session_phase26_test.go`, `main_test.go`, `session_phase26_overflow_test.go`, `interp_test.go`, `cgen_program_test.go` | U64-01, APP-07 | Yes | 0 phase-linked disabled tests found | No; expected sum/overflow values are literal pins | Exact value, streams, status, reached mutations | PASS |
| `check_test.go`, `corevalidate_cycle_peer_test.go`, `originvalidate_test.go`, `executionpeer_test.go` | FLOW-01, FLOW-02 | Yes | 0 phase-linked disabled tests found | No | Admission, independent re-derivation, forged/refusal controls | PASS |
| `session_phase26_event_test.go`, `native_app_test.go`, `syntax_test.go` | FLOW-01, FLOW-02, APP-07 | Yes | 0 phase-linked disabled tests found | No | Event identity, capacity separation, spans/recovery | PASS |

Disabled-test scan found one existing unrelated Clang-dependent skip in `native_app_test.go` at line 89; it is outside the Phase 26 requirement-linked tests. Source-mutant files written by tests are inputs to reached mutation checks, not generated expected-answer fixtures.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unresolved TBD/FIXME/XXX markers or implementation stubs found in phase implementation files. Grep matches for empty slices/nulls are test helpers, inventory returns, or existing non-phase logic. | — | No blocker. |

### Human Verification Required

None. This phase’s acceptance is observable through exact source execution and hosted behavioral tests; it has no visual UI, real-time interaction, or external service behavior requiring manual judgment.

### Gaps Summary

No roadmap truth or requirement is missing. One 26-05 plan sentence overstates the app capture report’s `Verified` field: the report intentionally remains false, while the captured evidence receives an independent `executionpeer.Validate` verdict in the source integration test. This preserves the locked evidence boundary and fulfills the phase goal. Correct that sentence in a dated planning amendment; do not change the app capture API to self-claim semantic verification.

---

_Verified: 2026-10-03T11:23:18Z_  
_Verifier: the agent (gsd-verifier)_
