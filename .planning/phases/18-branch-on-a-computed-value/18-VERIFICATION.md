---
phase: 18-branch-on-a-computed-value
verified: 2026-09-26T16:12:16Z
status: passed
score: 30/30 must-haves verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/LANGUAGE-MATURITY.md
  - .planning/REQUIREMENTS.md
  - .planning/phases/18-branch-on-a-computed-value/18-01-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-01-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-02-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-02-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-03-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-03-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-04-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-04-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-05-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-05-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-06-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-06-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-07-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-07-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-08-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-08-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-09-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-09-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-10-PLAN.md
  - .planning/phases/18-branch-on-a-computed-value/18-10-SUMMARY.md
  - .planning/phases/18-branch-on-a-computed-value/18-CONTEXT.md
  - .planning/phases/18-branch-on-a-computed-value/18-DISCUSSION-LOG.md
  - .planning/phases/18-branch-on-a-computed-value/18-PATTERNS.md
  - .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
  - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
  - .planning/phases/18-branch-on-a-computed-value/18-UAT.md
  - .planning/phases/18-branch-on-a-computed-value/18-REVIEW.md
  - .planning/phases/18-branch-on-a-computed-value/18-09-RUNS.log
  - .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh
  - .planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh
  - .planning/quick/260924-djg-avoid-groundedness-false-positives-from-/260924-djg-PLAN.md
  - .planning/quick/260924-djg-avoid-groundedness-false-positives-from-/260924-djg-SUMMARY.md
  - .planning/quick/260924-djg-avoid-groundedness-false-positives-from-/260924-djg-VERIFICATION.md
  - .planning/quick/260924-gee-refresh-language-maturity-derived-guard/260924-gee-PLAN.md
  - .planning/quick/260924-gee-refresh-language-maturity-derived-guard/260924-gee-SUMMARY.md
  - .planning/quick/260924-gee-refresh-language-maturity-derived-guard/260924-gee-VERIFICATION.md
  - .planning/quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/260924-k1v-PLAN.md
  - .planning/quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/260924-k1v-SUMMARY.md
  - .planning/quick/260924-k1v-reconcile-pathoracle-loan-endpoints-for-/260924-k1v-VERIFICATION.md
  - .planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-PLAN.md
  - .planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-SUMMARY.md
  - .planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-VERIFICATION.md
  - .planning/quick/260924-kto-restore-historical-core-compatibility-an/260924-kto-PLAN.md
  - .planning/spikes/007-loan-across-branch/README.md
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase18_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_phase18_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_phase18_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session_payload_control_test.go
  - internal/compiler/session/session_peer_gate_test.go
  - internal/compiler/session/session_phase18_adversarial_test.go
  - internal/compiler/session/session_phase18_loan_test.go
  - internal/compiler/session/session_phase18_payload_test.go
  - internal/compiler/session/session_phase18_test.go
  - internal/compiler/session/witness_registry_test.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase18_test.go
  - internal/compiler/syntax/syntax_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase16/validation-corpus-run-record.jsonl
  - testdata/phase16/validation-corpus-run-record.manifest.json
  - testdata/phase18/computed_match.lang
  - testdata/phase18/loan_across_branch.lang
  - testdata/phase18/payload_return.lang
  - testdata/phase18/result_computed_match.lang
  - testdata/phase18/long_payload_return.lang
covered_digest: "v1:sha256:d9364a795f6cd29eb898c4039f40e73f965b1acc490b771b0a11caca58286b9f"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 25/26
  gaps_closed:
    - "Accepted long alternative names now serialize the exact destructured payload in native terminal outcomes."
  gaps_remaining: []
  regressions: []
human_verification: []
decision_coverage:
  honored: 5
  total: 5
  not_honored: []
---

# Phase 18: Branch on a Computed Value — Verification

**Phase Goal:** A branch can discriminate a value the function computed, not only its own parameter, by generalizing `match` to the terminal form of a linear body over an in-scope `data` place.
**Verified:** 2026-09-26T16:12:16Z
**Status:** passed
**Re-verification:** Yes — after closing the long-tag native terminal serialization gap.

The previous report's failing CTL-03 case is closed. The ten phase plans and summaries, current requirements, implementation, source fixtures, UAT, validation record, and review report were checked. The traceability table now contains the exact `S-010` ID in its first cell; its coverage note identifies it as a prerequisite gate, not an M003 requirement, and excludes it from the 33-requirement count. No source, UAT, or validation artifacts changed. The full suite passed after the substantive traceability change and before this exact-ID formatting adjustment; this verifier reran the focused long-payload and wrong-slot tests after the adjustment.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A linear prefix binds an in-scope `data` place and its terminal `match` discriminates that place through the source-to-core path. | ✓ VERIFIED | `computed_match.lang`; source tests assert the scrutinee ID resolves to the computed prefix place; checker, peer, interpreter, and native acceptance paths are present and wired. |
| 2 | A `Result` returned by a callee can be matched by its caller with independent admission and cross-engine agreement. | ✓ VERIFIED | `result_computed_match.lang`; `TestPhase18ResultComputedMatch` checks the `OpCall` target, independent peers, four execution tiers and full session/program comparison. The diagnostic-ID refusal control covers the fifth comparator axis. |
| 3 | A computed match arm returns its destructured payload place, and wrong-slot corruption is observable at terminal outcome for all accepted alternative names. | ✓ VERIFIED | The short `Ok:01020304` and 152-character tag both return their exact runtime payloads; both wrong-slot tests require positive injection and `AxisTerminalOutcome` disagreement. Focused tests passed. |
| 4 | The production one-arm-live loan crosses computed-match fan-out using existing endpoints and bounded, fail-closed liveness. | ✓ VERIFIED | Production fixture and loan tests assert existing `point`/`edge` endpoints, sibling-arm isolation, and `loanLivenessBound = 4 × blocks × (distinct loans + 1)`; production source check recorded 173 work units. |
| 5 | The original terminal-match parser frontier was pinned before admission changed. | ✓ VERIFIED | Phase 18 source test pins `syntax.expected_linear_result` and the diagnostic span at the original frontier. |
| 6 | Independent source fixtures distinguish computed match, Result-returning call, payload return, and pre-branch loan. | ✓ VERIFIED | Four original fixtures plus `long_payload_return.lang` exist and are loaded by tests that assert their intended source shapes. |
| 7 | Fixture controls are non-vacuous and distinguish their CTL claims. | ✓ VERIFIED | Tests assert place IDs/types, positive mutation injection, exact divergence axes, peer calls, required engine participation, and seeded comparator divergences. |
| 8 | Computed terminal match extends the existing branch law rather than introducing a second branch construct. | ✓ VERIFIED | AST/parser/checker route terminal match through the existing linear body and branch CFG; round-trip and source differential tests exercise that path. |
| 9 | Independent peers derive computed scrutinee place, type, ownership, and payload origin, and reject malformed or forged facts. | ✓ VERIFIED | Core/origin peer tests construct forged, wrong-type, out-of-scope, and sibling-arm cases independently of checker derivation. |
| 10 | Result arm return typing remains distinct from the callee parameter and invalid or forged returns are rejected. | ✓ VERIFIED | Checker/core tests assert the Result-typed arm value place and reject Resource-sourced and forged `OpReturn` values. |
| 11 | Called callee prefix operations run before dispatch on the computed place. | ✓ VERIFIED | Interpreter and native call-prefix tests exercise the production call path; session Result-call test verifies the returned computed value. |
| 12 | Accepted computed-match source executes through interpreter and all three native optimization tiers. | ✓ VERIFIED | Named session tests compare interpreter, `-O0`, `-O3`, and `-O3 -flto`; the full-suite pass was also reported for this turn. |
| 13 | Unmutated payload execution agrees before corruption is injected. | ✓ VERIFIED | `TestPhase18PayloadPlaceReturn` and `TestPhase18LongPayloadPlaceReturn` assert the runtime value in all tiers before mutation tests. |
| 14 | Shared-prefix places and arm-specific last-use facts remain isolated and sound. | ✓ VERIFIED | Checker controls cover one-arm, neither-arm, and both-arm CFG topologies and assert exact sibling-edge endpoints. |
| 15 | Invalid-place, wrong-type, forged-peer, missing-peer, missing-engine, and comparator-axis controls fail closed. | ✓ VERIFIED | Phase 18 checker, peer, session adversarial, and comparator-control tests cover each negative condition. |
| 16 | Acceptance cannot pass when peer validation, required execution engines, or mutation injection is omitted. | ✓ VERIFIED | Comparator controls count peer calls, reject missing `O3-LTO`, require positive mutation injection, and constrain the exact disagreement axis. |
| 17 | The claimed comparator axes and loan endpoint semantics are exercised. | ✓ VERIFIED | Comparator controls seed terminal, event, resource, exit, and diagnostic-ID divergences; loan tests assert point/edge kinds and exact match edge identity. |
| 18 | Focused and full verification costs have repeatable cold/warm distributions and provenance. | ✓ VERIFIED | Phase validation records lane samples, host/tool provenance, and distributions; Plan 08 evidence verifier is part of recorded validation evidence. |
| 19 | Recurring focused CI coverage is justified by measured value and runtime. | ✓ VERIFIED | Measured disposition retains the existing full/race jobs and documents why an additional focused command duplicates coverage; workflow evidence was recorded. |
| 20 | Existing macOS/Linux CI retains build, vet, full tests, race tests, and installed Clang coverage. | ✓ VERIFIED | `.github/workflows/ci.yml` has Ubuntu/macOS jobs with Clang, build, vet, full test, and race commands; Phase 18 test packages are included. |
| 21 | S-010 was rerun against production source code. | ✓ VERIFIED | Production loan fixture passes source checker, peer, interpreter, and native paths and records the bounded work count. |
| 22 | Default-parallel full-suite reliability is restored without weakening bounded subprocess behavior. | ✓ VERIFIED | Plan 09 records three default-parallel full suites and other lane receipts; cache/native subprocesses retain finite budgets, cancellation, typed failures, and negative controls. Current-turn full suite exit code 0 is reported by the orchestrator. |
| 23 | Failed Clang identity probes refuse cache reuse while preserving inspectable causes. | ✓ VERIFIED | Cache probe tests distinguish timeout and non-timeout wrapped causes while preserving fail-closed cache refusal. |
| 24 | Native and Clang subprocesses remain bounded under package load and race instrumentation. | ✓ VERIFIED | Runner and probe implementations apply finite deadlines while respecting caller deadlines and cancellation; Plan 09 records full/race lane evidence. |
| 25 | Repeated post-fix aggregate evidence is complete and bound to captured output. | ✓ VERIFIED | Plan 09's evidence gate verifies seven run sections, commands, exit statuses, output ranges, hashes, durations, and source/tool metadata. |
| 26 | The existing CI contract is the recurring coverage, with no manual UAT criterion for this compiler objective. | ✓ VERIFIED | The compiler objective is objectively tested; `18-UAT.md` remains complete at 16/16 automated checks with zero issues. |
| 27 | The accepted 152-character alternative returns the destructured Buffer under interpreter, `-O0`, `-O3`, and `-O3 -flto`. | ✓ VERIFIED | `TestPhase18LongPayloadPlaceReturn` asserts a 152-character name, checks source, and asserts the exact `tag:01020304` result in all four tiers. |
| 28 | Every long-tag execution tier returns the exact tag and payload, and the full comparator agrees. | ✓ VERIFIED | The named long-tag test compares the exact terminal value and calls `Phase5CompareProgramEngines`; the command passed in this verification pass. |
| 29 | Ordinary payload agreement and positive-injection wrong-slot controls remain intact. | ✓ VERIFIED | `TestPhase18PayloadPlaceReturn`, `TestPhase18WrongSlotMutation`, and `TestPhase18LongTagWrongSlotMutation` passed; both mutation tests require positive injected writes and exact terminal-outcome divergence. |
| 30 | Terminal JSON has no tag-length-specific temporary limit and retains escaping and the document output bound. | ✓ VERIFIED | `emitProgramTerminalValueWriter` streams the runtime tag through `lang_write_json_string_content` and emits payload through bounded `lang_write_bytes`; no `rendered[128]` remains. The schema-2 output-bound and lowering controls passed. |

**Score:** 30/30 truths verified; behavior-unverified: 0.

### Required Artifacts

All ten plans and ten summaries exist. Plan artifact metadata checks passed for 18-01 through 18-10; the previous artifact audit recorded 2/2, 2/2, 2/2, 10/10, 5/5, 2/2, 1/1, 3/3, and 6/6 artifacts for Plans 01–09. Plan 10's four artifacts were checked directly at existence, substance, and wiring levels:

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `testdata/phase18/long_payload_return.lang` | Accepted long-tag computed match returning Buffer payload | ✓ VERIFIED | 152-character `Outcome` alternative; consumed by the named session test. |
| `internal/compiler/session/session_phase18_payload_test.go` | Exact long and ordinary values plus wrong-slot controls | ✓ VERIFIED | Test bodies assert value, tier agreement, injection count, and exact axis. Both mutation tests defer seam restoration immediately after enabling it. |
| `internal/compiler/cgen/cgen.go` | Schema-2 escaped JSON string content writer | ✓ VERIFIED | Helper and quote wrapper are defined; schema-2 terminal writer uses the content helper. |
| `internal/compiler/cgen/cgen_program.go` | Bounded streamed terminal tag and payload serialization | ✓ VERIFIED | Runtime selected tag and payload flow through bounded writer; unknown alternatives and payload lengths fail closed. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Linear source prefix | `Match.ScrutineeID` | Parser/checker/core place resolution | ✓ WIRED | Tests assert that the match uses the computed place, not the function parameter. |
| Result callee call target | Caller terminal match | Checked `OpCall` result to interpreter/native dispatch | ✓ WIRED | Result fixture and four-tier comparison verify the call-produced data value. |
| Destructured payload place | `Execution.Outcome.Value` | Arm return, schema-2 serialization, session comparator | ✓ WIRED | Short and long payload tests compare exact runtime bytes; mutation changes the terminal outcome. |
| Long tag and payload | Bounded JSON output | Escaping helper and `lang_write_bytes` accounting | ✓ WIRED | Streaming eliminates tag-sized scratch while retaining output preflight and write-failure propagation. |
| Pre-match borrow | Loan point/edge endpoints | Production CFG, fixpoint, endpoint materialization | ✓ WIRED | Source-driven test verifies one-arm point use and unused sibling edge endpoint. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| Computed match | `Match.ScrutineeID` | Prefix operation target | Yes; branch selects on the computed runtime place | ✓ FLOWING |
| Result caller | Caller result place | `OpCall` return | Yes; returned Result selects `Accepted` | ✓ FLOWING |
| Payload return | Arm's typed payload place | Destructured runtime Buffer | Yes; exact `01020304` appears after the selected tag | ✓ FLOWING |
| Long terminal tag | Runtime `tag_name` and selected payload field | Returned native union value | Yes; JSON content writer emits the real selected tag and field | ✓ FLOWING |
| Loan witness | Loan endpoint records | Production CFG branch liveness | Yes; records identify real point and match-edge uses | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Long alternative payload and positive-injection wrong-slot mutation | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase18LongPayloadPlaceReturn\|TestPhase18LongTagWrongSlotMutation)$' -count=1` | Passed after the exact-ID adjustment (`ok`, 2.962s). | ✓ PASS |
| Entire Go workspace suite | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` | Orchestrator reports exit code 0 (126.349s), after substantive traceability changes and before the exact-ID-only formatting adjustment. | ✓ PASS |

### Probe Execution

No conventional `scripts/*/tests/probe-*.sh` files or phase-declared probe paths were found. Phase 18 validation scripts and receipts remain unchanged and are covered by the previously recorded evidence gates.

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|---|---|---|---|---|
| CTL-01 | 01, 02, 03, 06, 07, 08, 09 | Branch discriminates a computed in-scope value and preserves production loan behavior. | ✓ SATISFIED | Computed source acceptance, independent peers, production loan endpoint tests, bounded fixpoint, and interpreter/native execution. |
| CTL-02 | 01, 04, 07, 08, 09 | A Result-returning callee's returned value is matched with five-axis evidence. | ✓ SATISFIED | Result call fixture, independent admission, interpreter and native tiers, full comparator, diagnostic-ID negative axis. |
| CTL-03 | 01, 03, 05, 07, 08, 09, 10 | Return a destructured payload place and detect seeded wrong-slot corruption. | ✓ SATISFIED | Ordinary and 152-character payload outcomes agree; both positive-injection mutations fail at terminal-outcome. |

CTL-01, CTL-02, and CTL-03 are the three M003 requirements mapped to Phase 18, and all three are marked complete. The traceability table row is exactly `| S-010 | Phase 18 gate | Complete — validated by spike 007 and re-run through production source |`. S-010 is a completed prerequisite gate, not an M003 requirement, and is excluded from the 33 M003 requirement count. No Phase 18 requirement is orphaned.

### Decision Coverage

GSD decision-coverage query: 5/5 trackable context decisions honored; none unaccounted for. This result is non-blocking.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|---|---|---:|---:|---:|---|---|
| Phase 18 syntax/check/corevalidate/originvalidate tests | CTL-01, CTL-02, CTL-03 | Yes | 0 in linked Phase 18 tests | 0 identified | Behavioral/value and negative controls | PASS |
| Phase 18 session, payload, loan, and adversarial tests | CTL-01, CTL-02, CTL-03 | Yes | 0 in linked Phase 18 tests | 0 identified | Four-tier exact values, comparator axes, positive mutation injection | PASS |
| Phase 18 cache/native and measured evidence tests | CTL-01, CTL-02, CTL-03 | Yes | 0 in linked requirement checks | 0 identified | Typed failure cause, finite deadlines, captured external run receipts | PASS |

No expected-value writers or disabled requirement tests were found in the linked Phase 18 checks. Payload expectations are authored input contracts, not values captured from the implementation under test.

### Advisory (New Scope, Unevidenced)

None. Re-verification found no new-scope anti-pattern requiring advisory treatment.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unresolved debt markers or stub behavior in the changed Plan 10 files. | — | The earlier fixed-buffer defect is removed; review reports no remaining findings. |

### Human Verification Required

N/A — compiler/toolchain foundation phase with no user-facing interaction. The observable acceptance criteria are covered by deterministic source-to-native checks; no truth is behavior-unverified.

### Gaps Summary

No remaining gaps. The former CTL-03 blocker came from fixed-width native formatting for an accepted long alternative name. The emitter now streams the selected tag and payload through the schema-2 escaped, bounded output writer. Long and ordinary values are tested across interpreter and three native optimization tiers; positive wrong-slot mutations remain visible on the exact terminal-outcome axis; schema-2 bounds remain covered. The existing 16/16 UAT and validation evidence were left untouched.

---

_Verified: 2026-09-26T16:12:16Z_
_Verifier: the agent (gsd-verifier)_
