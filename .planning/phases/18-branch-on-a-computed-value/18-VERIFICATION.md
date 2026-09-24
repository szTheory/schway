---
phase: 18-branch-on-a-computed-value
verified: 2026-09-24T20:33:39Z
status: passed
score: 21/21 must-haves verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/LANGUAGE-MATURITY.md
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/STATE.md
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
  - .planning/phases/18-branch-on-a-computed-value/18-CONTEXT.md
  - .planning/phases/18-branch-on-a-computed-value/18-DISCUSSION-LOG.md
  - .planning/phases/18-branch-on-a-computed-value/18-PATTERNS.md
  - .planning/phases/18-branch-on-a-computed-value/18-RESEARCH.md
  - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
  - .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh
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
  - .planning/spikes/007-loan-across-branch/README.md
  - .planning/state.json
  - internal/compiler/ast/ast.go
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
  - testdata/phase18/computed_match.lang
  - testdata/phase18/loan_across_branch.lang
  - testdata/phase18/payload_return.lang
  - testdata/phase18/result_computed_match.lang
covered_digest: "v1:sha256:b5af9d7884e86be9b63fb4041c8ef18e0966a9d20550540fd898806221c89d01"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 5
  total: 5
  not_honored: []
human_verification: []
---

# Phase 18: Branch on a Computed Value — Verification

**Phase Goal:** A branch can discriminate a value the function computed, not only its own parameter, by generalizing `match` to the terminal form of a linear body over an in-scope `data` place.
**Verified:** 2026-09-24T20:33:39Z
**Status:** passed
**Re-verification:** No — initial phase verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A linear prefix binds an in-scope `data` place and its terminal `match` discriminates that place through the source-to-core path. | ✓ VERIFIED | `computed_match.lang`; `TestPhase18ComputedScrutineeProductionPathAccepted` asserts the scrutinee ID names the `computed` place; parser/check/session tests pass. The earlier refusal and exact `syntax.expected_linear_result` span are pinned in commit `4f40dce`. |
| 2 | A `Result` returned by a callee can be matched by its caller, with independent admission and cross-engine agreement. | ✓ VERIFIED | `result_computed_match.lang`; `TestPhase18ResultComputedMatch` proves the match scrutinee is the `OpCall` target typed `Result`, validates core and origin peers, and checks `Accepted` on interpreter, `-O0`, `-O3`, and `-O3 -flto`. The established fifth axis is separately exercised by stable diagnostic-ID comparison on a refused program in `TestPhase18FiveAxis`; accepted executions have four execution-bearing axes because refusals do not execute. |
| 3 | A computed match arm returns its destructured payload place, and wrong-slot corruption is observable at terminal outcome. | ✓ VERIFIED | `TestPhase18PayloadPlaceReturn` expects `Ok:01020304` from all four execution tiers and compares the program evidence. `TestPhase18WrongSlotMutation` asserts a positive injected-write count and exact `axis:terminal-outcome` disagreement. |
| 4 | The production one-arm-live loan crosses computed-match fan-out using existing endpoints and bounded, fail-closed liveness. | ✓ VERIFIED | S-010 production fixture `loan_across_branch.lang`; `TestPhase18LoanAcrossBranchProduction` checks peer admission, a `point` and `edge` endpoint, and both branch outcomes on interpreter/native tiers. Checker test asserts the unused `Off` sibling receives the edge endpoint and the one-arm use receives the point endpoint. `loanLivenessFixpoint` enforces `loanLivenessBound = 4 × blocks × (distinct loans + 1)` and returns no partial facts on bound/cycle refusal; Phase 18 production checking succeeds. Source check reports `recomputed_work=173`. |
| 5 | The original terminal-match parser frontier was pinned before admission changed. | ✓ VERIFIED | `git show 4f40dce:internal/compiler/check/check_phase18_test.go` shows an assertion for `syntax.expected_linear_result` and a nonempty exact source span. |
| 6 | Separate source witnesses distinguish computed match, Result-returning call, payload return, and pre-branch loan. | ✓ VERIFIED | Four independent fixtures exist under `testdata/phase18/`; `TestPhase18ResultFixtureFrontier`, `TestPhase18PayloadFixtureFrontier`, checker tests, and loan source tests load and inspect the intended witnesses. |
| 7 | Fixture controls are non-vacuous and distinguish their CTL claims. | ✓ VERIFIED | Tests assert fixture needles, AST terminal-match shape, exact place IDs/types, positive mutation injection, exact divergence axis, and peer invocation; adversarial tests reject omitted execution tiers and seeded comparator divergences. |
| 8 | Computed terminal match extends the existing branch law rather than introducing a second branch construct. | ✓ VERIFIED | `ast.LinearBody.TerminalMatch`, parser `linearBody`, and checker lowering route the terminal match into the existing branch CFG; syntax/checker/source differential tests pass. |
| 9 | Independent peers derive computed scrutinee place, type, ownership, and payload origin, and refuse malformed or forged facts. | ✓ VERIFIED | `corevalidate_phase18_test.go` and `originvalidate_phase18_test.go` contain independent derivation and forged/wrong-type/sibling-arm negative controls; phase tests pass. |
| 10 | Result arm return typing remains distinct from the callee parameter and invalid or forged returns are refused independently. | ✓ VERIFIED | Checker asserts `Resource -> Result`, a Result-typed `ValuePlaceID`, and matching `OpReturn`; core peer tests reject Resource-sourced and forged values. `TestPhase18ResultArmValuePlace` passes. |
| 11 | Called callee prefix operations run before dispatch on the computed place. | ✓ VERIFIED | `interp_test.go:TestPhase18CallComputedMatchPrefix` and session Result call differential test exercise production call execution; interpreter and C emitter use the checked scrutinee place for dispatch. |
| 12 | Accepted computed-match source executes through interpreter and all three native optimization tiers. | ✓ VERIFIED | Named session checks executed in this verification pass; all four tiers assert returned values, and full workspace suite is reported green by the orchestrator. |
| 13 | Unmutated payload execution agrees before corruption is injected. | ✓ VERIFIED | `TestPhase18PayloadPlaceReturn` runs the unmutated companion across four tiers before the positive mutation test. |
| 14 | Shared-prefix places and arm-specific last-use facts remain isolated and sound. | ✓ VERIFIED | Checker tests exercise the one-arm source, a neither-arm source, and a valid synthetic both-arm CFG; they assert endpoint shape on exact sibling edges. |
| 15 | Invalid-place, wrong-type, forged-peer, missing-peer, missing-engine, and comparator-axis controls fail closed. | ✓ VERIFIED | `TestPhase18ComputedScrutineeRefusals`, core/origin forged-input tests, and `session_phase18_adversarial_test.go` provide separate negative controls; all named tests pass. |
| 16 | Acceptance cannot pass when peer validation, required execution engines, or mutation injection is omitted. | ✓ VERIFIED | `TestPhase18ComparatorControlRejectsMissingExecutionTierAndRequiresPeer` counts peer calls and refuses a missing `O3-LTO` document; wrong-slot test requires positive injection. |
| 17 | The exact claimed comparator axes and loan endpoint semantics are exercised. | ✓ VERIFIED | Comparator anti-controls seed terminal outcome, event order, resource ledger, and exit-signal divergences; diagnostic-ID refusal has its own exact axis test. Loan tests assert point/edge kinds and the `Off` edge identity. |
| 18 | Focused and full verification feedback costs have repeatable cold/warm distributions with provenance. | ✓ VERIFIED | The phase-local `--measurements` verifier passes against all five cold/warm lane distributions, host/tool versions, and the current workflow blob. |
| 19 | Recurring focused CI is used only when measured regression value justifies the duplication. | ✓ VERIFIED | `--final` passes; it proves the unchanged workflow hash and records why the focused native session command (about two minutes per host) duplicates the existing full/race coverage without a distinct signal. |
| 20 | Existing macOS/Linux CI checks retain build, vet, full tests, race tests, and installed Clang coverage. | ✓ VERIFIED | `.github/workflows/ci.yml` matrix is Ubuntu/macOS, requires Clang, and runs `go vet ./...`, `go build ./...`, `go test ./...`, and `go test -race ./...`; the Phase 18 tests are in those packages. |
| 21 | S-010 was clean for its pre-planning hard gate and was rerun through production source code. | ✓ VERIFIED | Spike 007 records the explicit CFG test and killed third-kind mutation. Phase 18 adds the production source fixture, production checker endpoint assertions, peer checks, and differential execution; the source validation passes. |

**Score:** 21/21 verified; behavior-unverified: 0.

### Required Artifacts

All plan-declared artifacts passed `gsd-tools query verify.artifacts`: Plan 01 2/2, 02 2/2, 03 2/2, 04 10/10, 05 5/5, 06 2/2, 07 1/1, and 08 3/3. The checked artifacts are substantive; source fixtures are consumed by the named tests, checker/peer/emitter artifacts are invoked by the integration paths, and the evidence verifier is runnable in both modes.

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Linear source prefix | `Match.ScrutineeID` place | Parser → checker → core | ✓ WIRED | Test asserts the ID points to the source-named place produced by the prefix. |
| Result-returning call target | Caller terminal match | Checker place/type facts → interpreter and C emitter | ✓ WIRED | Session test checks the exact OpCall target identity, `Result` type, and expected outcomes in all four execution tiers. |
| Destructured payload | Terminal outcome value | Arm value place → OpReturn → execution serialization → comparator | ✓ WIRED | Expected payload bytes survive native/interpreter execution; injected wrong-slot mutation diverges on the terminal axis. |
| Pre-match borrow | Liveness endpoints on match edges | Production CFG → `loanLivenessFixpoint` → `materializeLoanEndpoints` | ✓ WIRED | Source test observes the point use and unused sibling edge endpoint, and production checker accepts both alternatives. |
| Measured evidence | CI disposition | Evidence verifier → workflow blob and matrix checks | ✓ WIRED | `--measurements` and `--final` both pass; workflow retains Ubuntu/macOS full/race jobs. |

The generic key-link query cannot interpret the plans' descriptive component names as filesystem paths; these links were traced directly through implementation and the named behavioral tests above.

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| Computed match | `Match.ScrutineeID` | Checked prefix operation's target place | Yes; branch reads that computed place | ✓ FLOWING |
| Result caller | caller's `result` place | `OpCall` target returned by `produce` | Yes; `Result` value selects `Accepted` | ✓ FLOWING |
| Payload return | arm's typed payload value place | Destructured runtime payload | Yes; `Ok:01020304` appears in execution outcome | ✓ FLOWING |
| Loan witness | loan endpoint records | CFG branch liveness | Yes; endpoint records refer to real match edge and point use | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Computed source, Result call, payload mutation, loan, axis controls | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^(TestPhase18ComputedSourceFourTierDifferential|TestPhase18ResultComputedMatch|TestPhase18FiveAxis|TestPhase18PayloadPlaceReturn|TestPhase18WrongSlotMutation|TestPhase18LoanAcrossBranchProduction|TestPhase18ComparatorControlRejectsMissingExecutionTierAndRequiresPeer|TestPhase18ComparatorAxesRejectSeededDivergence)$' -count=1` | Passed in 8.2s | ✓ PASS |
| Parser frontier and computed terminal form | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/syntax -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| Production checker/refusal/loan controls | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/check -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| Independent core peer | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/corevalidate -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| Independent origin peer | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/originvalidate -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| Interpreter call-prefix regression | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/interp -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| C emitter regression | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run '^TestPhase18' -count=1` | Passed | ✓ PASS |
| Entire Go suite and race suite | `go test ./... -count=1`; `go test -race ./... -count=1` | Full suite and 30 cold/warm lane samples reported passing by the orchestrator; the phase-local verifier confirms the recorded runs and distribution metadata. | ✓ PASS |
| Production S-010 source check | `GOCACHE=/tmp/ai-lang-gocache go run ./cmd/lang check testdata/phase18/loan_across_branch.lang` | Exit 0; `recomputed_work=173` | ✓ PASS |

### Probe Execution

No `probe-*.sh` files or phase-declared probe paths apply. The phase-local validation verifier was executed directly in both prescribed modes; both exited 0.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| CTL-01 | 01, 02, 03, 06, 07, 08 | Match a computed in-scope value and preserve production loan behavior | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; computed-source differential, independent peers, and production S-010 fixture with existing point/edge endpoints and bounded fail-closed liveness. |
| CTL-02 | 01, 04, 07, 08 | Match a Result-returning callee's returned value with cross-engine agreement | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; `TestPhase18ResultComputedMatch`; four execution tiers plus separate diagnostic-ID refusal control for the fifth comparator axis. |
| CTL-03 | 01, 03, 05, 07, 08 | Return a destructured payload and detect seeded wrong-slot corruption | ✓ COMPLETE | `REQUIREMENTS.md` marks complete; payload value assertion and positive injected mutation with exact terminal-outcome divergence. |

All three Phase 18 requirements mapped in `REQUIREMENTS.md` are complete and have direct automated evidence; none are orphaned from the plans.

### Decision Coverage

The GSD decision-coverage query reports all 5 trackable context decisions honored and no missing decisions.

### Test Quality Audit

| Test File Group | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|---|---|---:|---:|---:|---|---|
| Phase 18 syntax/check/corevalidate/originvalidate tests | CTL-01, CTL-02, CTL-03 | Yes | 0 | 0 | Behavioral/value and negative-control assertions | PASS |
| Phase 18 session, payload, loan, and adversarial tests | CTL-01, CTL-02, CTL-03 | Yes | 0 | 0 | Cross-engine values, exact axes, mutation injection, peer execution | PASS |

No disabled requirement tests or expected-value writers were found in the linked test files. The fixtures and payload expectations are authored inputs/contracts; they are not captured from the system under test.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| None | — | No unreferenced `TBD`, `FIXME`, or `XXX` markers; no empty phase implementation or disabled linked tests | — | No blockers found. |

### Human Verification Required

N/A — compiler/toolchain foundation phase with no user-facing UX. Every acceptance criterion is established by deterministic source fixtures, peer validation, native/interpreter differential tests, mutation controls, production liveness checks, and measured build/test evidence. `human_verification` is empty.

### Gaps Summary

No gaps remain against the Phase 18 roadmap success criteria or the merged plan must-haves. CTL-01 through CTL-03 are now marked complete in `REQUIREMENTS.md`, and this refreshed report fingerprints that updated state. The roadmap updater still reports Phase 18 `roadmap_complete: false` because the phase entry is maintained differently from writable phase details; no roadmap text was modified to mask that condition.

---

_Verified: 2026-09-24T20:33:39Z_  
_Verifier: the agent (gsd-verifier)_
