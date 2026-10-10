---
phase: 26-checked-scalar-sum
verified: 2026-10-10T14:21:07Z
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
covered_digest: "v3:sha256:0148c9dc2e4aa542644ab26e365bf37e8cc96323891c66ec7c2decedfa64ba84"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 10
  total: 10
  not_honored: []
---

# Phase 26: Checked Scalar Sum Verification Report

**Phase Goal:** A developer can run a source-authored scalar loop that computes and prints the exact bounded `sum_to_n` result, with defined overflow and explicit refusal of unsafe loop-carried state.
**Verified:** 2026-10-10T00:00:00Z
**Status:** passed
**Re-verification:** Yes — after Phase 27 modified covered source (regression check against HEAD 127ee3f)

## Re-verification Summary (Phase 27 regression check)

Prior report (2026-10-03, passed, 5/5) went stale because Phase 27 touched covered files. Re-checked at HEAD 127ee3f:

- `git diff --stat 6224f8b HEAD -- ':!.planning'` is empty: non-planning source at HEAD equals the hosted-evidence revision 6224f8b.
- Hosted run [38054848320](https://github.com/szTheory/schway/actions/runs/38054848320) at 6224f8b: success; checks (ubuntu, macos) and current-evidence aggregates passed (go vet, go test ./..., go test -race ./...). Phase 26 focused job and corpus receipt job skipped (dispatch-gated); Phase 26 tests are part of `go test ./...`.
- Source inspection at HEAD: interpreter checked add (`interp.go` OpAddChecked, `^uint64(0)-a < b` defect before any wrapped value) intact; C emitter emits `schway: U64 addition overflow` + exit 65; fixed-point budgets of 65536 remain in `check/scalar.go`, `corevalidate.go` (transferLimit) and `originvalidate.go`; `executionpeer` still has occurrence tracking; `examples/sum_to_n.schway` unchanged in shape (guard `n < 1001`, loop, repeated scalar copy, defect "input exceeds 1000"); `TestPhase26PublicAppCLI` still pins 0/55/500500 and 1001 exit 65, plus overflow tuple; 27 `TestPhase26*` functions present (ExactSumMatrix, OverflowProcessOutcome, PeerOccurrenceOrder, three AnalysisExhaustion homes, etc.); `ci.yml` still carries the dispatch-gated phase26-focused job.
- 26-VALIDATION.md and 26-SECURITY.md re-audited this session (21/21 tests exist, 12/12 threats closed). No Phase 26 task commits since the clean 26-REVIEW.md; Phase 27 diff reviewed in 27-REVIEW.md.
- No Phase 26 guarantee weakened. Phase 27 additions (fizzbuzz, `TestPhase27AppBoundary`) sit alongside, not in place of, Phase 26 assertions. Requirement IDs U64-01, FLOW-01, FLOW-02, APP-07 from PLAN frontmatter all appear in REQUIREMENTS.md (marked Complete, mapped to Phase 26); no orphans.

Limitation: no local test run (hosted-only policy); the focused Phase 26 job was not re-dispatched at 6224f8b, so behavioral evidence for Phase 26 specifically comes from the full-suite run plus the earlier focused run 37118144404.

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

_Verified: 2026-10-10T00:00:00Z_  
_Verifier: the agent (gsd-verifier)_

## Addendum — 2026-10-10 (Phase 27.1)

Phase 27.1 changed `internal/compiler/session/session.go` and `internal/compiler/session/session_app_verify.go` after 6224f8b: application admission now applies the origin peer after corevalidate on `schway build` and `schway app verify`, through the shared helper also used by the interpreter application route. This supersedes the earlier note that commits after 6224f8b are planning-only. Current-source hosted evidence is run 38068888687 at 7936dac39a6215091d4dbd02a6acec20f2601386 (both hosts green, fresh receipts pass the pinned gate); see `.planning/phases/27.1-address-tech-debt-build-route-origin-validation-and-receipt/27.1-HOST-EVIDENCE.md`. The Phase 26 and Phase 27 verifier agents were not re-run (proportionate validation, D-27.1-15); the findings above stand as written.

## Addendum — 2026-10-10 (Phase 27.2)

Phase 27.2 restores straight-line scalar bodies for FLOW-01 and U64-01. Milestone audit finding R1 (branch-free scalar bodies refused with core.release_order_indeterminate) is fixed producer-side. The root cause is the entry-block OpReturn that the corevalidate release-order rule refuses; it predates a671308 for `var` bodies, because a671308 only routed `let`-led bodies into the scalar builder and thereby exposed the same shape to `let`.

Three producer defects were fixed: the checker scalar builder now splits the return into its own block when no if or while ran (b2e43aa), the cgen scalar classifier gained a structural clause for branch-free copy, constant and Unit write-only bodies (b2e43aa), and the unused entry parameter is marked void only when no operation reads it (28d7e8f). Validators (corevalidate, originvalidate, executionpeer, interp) are unchanged. Accepted-corpus core and C bytes are pinned by TestPhase272AcceptedCorpusByteIdentity against a golden generated at 9459f37 before the fix. Example sources and fixtures remain pinned to the Phase 27 and 27.1 receipt hashes, re-verified by the receipt-pin check. Red and green stages are recorded in 27.2-01-red-evidence.json.

Hosted evidence: pending, see `.planning/phases/27.2-address-tech-debt-straight-line-scalar-let-admission-r1/27.2-HOST-EVIDENCE.md`.

R1b (deferred): the program `let a = n`, `let b = helper(a)`, `b` is refused with syntax.expected_linear_result because a671308 routes on the first `let` only; it passed at a671308~1. No repository program exercises it. It is pinned by TestPhase272R1bScalarFirstLetThenCallCharacterization and is out of 27.2 scope (parser routing). It is recorded here as prose only; no DEBT register is edited.

The milestone audit R1 row is updated by the next /gsd-audit-milestone, not in this phase. Earlier sections above are unchanged.
