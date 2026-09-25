---
phase: "18"
slug: "branch-on-a-computed-value"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-24"
---

# Phase 18 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package (Go 1.24) |
| **Config file** | `go.mod`; `.github/workflows/ci.yml` |
| **Quick run command** | `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -count=1` |
| **Full suite command** | `go test ./...`; phase gate also runs `go vet ./...`, `go build ./...`, `go test -race ./...`, and native differential checks with installed Clang |
| **Estimated runtime** | Measure in Wave 0 and record; do not assume a latency budget without evidence |

---

## Sampling Rate

- **After every task commit:** Run the narrow owning-package test(s) for changed parser, checker, validator, interpreter, or backend behavior.
- **After every plan wave:** Run the focused Phase 18 source/differential/mutation checks and `go test ./...`.
- **Before `$gsd-verify-work`:** Run `go vet ./...`, `go build ./...`, `go test ./...`, `go test -race ./...`, and the native differential evidence with installed Clang.
- **Max feedback latency:** Measure the focused and full-suite durations in Wave 0; keep the recurring focused CI lane bounded and justify any broader recurring lane by regression value and runtime.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 18-01-T1 | 01 | 0 | CTL-01 | T-18-01 | Computed-place source reaches production checking; invalid, shadowed, and non-data places refuse | fixture, unit | `go test ./internal/compiler/check -run '^TestPhase18(ComputedScrutineeProductionPathAccepted|ComputedScrutineeRefusals)$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-01-T2 | 01 | 0 | CTL-02, CTL-03 | T-18-02 | Separate Result-call and payload-return witnesses preserve terminal computed matches | fixture, integration | `go test ./internal/compiler/session -run '^TestPhase18(ResultFixtureFrontier|PayloadFixtureFrontier)$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-01-T3 | 01 | 0 | CTL-01 | T-18-01 | Production source fixture has a pre-match borrow used in one arm | fixture, ownership | `go test ./internal/compiler/check -run '^TestPhase18LoanAcrossBranchFixture$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-02-T1 | 02 | 1 | CTL-01 | T-18-01, T-18-02 | Computed match parses, checks, passes independent peers, and agrees across execution tiers | source integration, differential | `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/session -run 'TestPhase18Computed|TestGeneratedBranchBodiesRoundTrip' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-02-T2 | 02 | 1 | CTL-01 | T-18-02 | Source spans/round trips persist and invalid computed-place controls refuse | syntax, negative controls | `go test ./internal/compiler/syntax ./internal/compiler/check -run 'Test(GeneratedBranchBodiesRoundTrip|Phase18Computed)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-03-T1 | 03 | 2 | CTL-01 | T-18-02 | Core peer admits valid places and rejects forged/malformed place facts | peer, negative controls | `go test ./internal/compiler/corevalidate -run 'Test(Phase18|CoreValidateBranch)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-03-T2 | 03 | 2 | CTL-01, CTL-03 | T-18-02 | Origin peer derives computed payload origins and rejects forged claims | peer, negative controls | `go test ./internal/compiler/originvalidate -run 'Test(Phase18|OriginValidate.*Payload)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-04-T1 | 04 | 3 | CTL-02 | T-18-02 | Result match typing keeps ordinary return mismatch and invalid-arm refusals | checker, negative controls | `go test ./internal/compiler/check -run '^TestPhase18(ResultComputedMatchChecker|ComputedScrutinee)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-04-T2 | 04 | 3 | CTL-02 | T-18-02 | Peer validates typed arm return; interpreter runs prefix before match; native emitter builds | peer, interpreter, native smoke | `go test ./internal/compiler/corevalidate ./internal/compiler/interp ./internal/compiler/cgen -run 'TestPhase18(ResultArmValuePlace|ComputedScrutinee|CallComputedMatchPrefix)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-04-T3 | 04 | 3 | CTL-02 | T-18-02 | Result caller is peer-admitted, returns expected value, and compares execution/refusal evidence | integration, differential | `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase18(ResultComputedMatchNativeReturn|ResultComputedMatchAdmission|ResultComputedMatch|FiveAxis)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-05-T1 | 05 | 4 | CTL-03 | T-18-03 | Interpreter/native terminal records include the actual returned payload | runtime integration | `go test ./internal/compiler/cgen -run '^TestProgramMatchPayloadLowering$' -count=1`; `go test ./internal/compiler/session -run '^TestPhase18PayloadPlaceReturn$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-05-T2 | 05 | 4 | CTL-03 | T-18-03 | Session preserves payload bytes and unmutated executions agree | integration, differential | `go test ./internal/compiler/session -run '^TestPhase18PayloadPlaceReturn$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-05-T3 | 05 | 4 | CTL-03 | T-18-03 | Wrong-slot mutation is observed and diverges exactly at terminal outcome | mutation-kill | `go test ./internal/compiler/session -run '^TestPhase18(WrongSlotMutation|PayloadPlaceReturn)$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-06-T1 | 06 | 5 | CTL-01 | T-18-01 | Source loan materializes point and sibling-edge endpoints; both/neither-arm controls distinguish topology | source ownership | `go test ./internal/compiler/check ./internal/compiler/session -run 'TestPhase18.*Loan|TestEdgeSpecificLiveOut' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-06-T2 | 06 | 5 | CTL-01 | T-18-01 | Liveness bound and cycle controls fail closed without partial facts | boundary, negative controls | `go test ./internal/compiler/check -run 'TestPhase18.*Loan|TestLoanLiveness(Fixpoint|Bound|Cycle)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-06-T3 | 06 | 5 | CTL-01 | T-18-01 | Production S-010 passes peers and interpreter/native execution for both alternatives | integration, differential | `go test ./internal/compiler/session -run '^TestPhase18LoanAcrossBranchProduction$' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-07-T1 | 07 | 6 | CTL-01, CTL-03 | T-18-02 | Peer-local forged place/type/origin controls reject independently of checker admission | adversarial peer tests | `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run 'TestPhase18.*(Forged|Mutation|Origin|Place)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-07-T2 | 07 | 6 | CTL-01, CTL-02, CTL-03 | T-18-02, T-18-03 | Comparator rejects missing peer/engine and seeded axis divergences; mutation seam is non-inert | adversarial integration, mutation-kill | `go test ./internal/compiler/session -run 'TestPhase18(ComparatorControl|ComparatorAxes|FiveAxis|PayloadWrongSlot|WrongSlotMutation)' -count=1` | ✅ | EXERCISED | — | ✅ green |
| 18-08-T1 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-01, T-18-02, T-18-03 | Five lanes have repeated cold/warm distributions and validated host/tool provenance | evidence verifier | `bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --measurements` | ✅ | EXERCISED | — | ✅ green |
| 18-08-T2 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-01, T-18-03 | CI disposition is justified and unchanged Ubuntu/macOS workflow matches measured evidence | CI configuration evidence | `bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --final` | ✅ | EXERCISED | — | ✅ green |

Threat refs:

- **T-18-01:** Untrusted input causes unbounded parser, CFG, or liveness work; preserve limits and fail closed at `4 × blocks × (loans+1)`.
- **T-18-02:** Checker and independent admission peers derive computed-place facts differently; independently validate place, type, and ownership.
- **T-18-03:** Native code writes the wrong payload slot but the comparator misses it; require mutation injection and terminal-outcome divergence.

The five-axis acceptance must use the full session/program comparator and peer-validation route. `Phase5CompareEngines` alone covers only four execution-bearing axes; diagnostic-ID refusal comparison is separate.

---

## Wave 0 Requirements

- [x] Add and pin the refused CTL-01 computed-place `.lang` frontier fixture before changing production admission; historical refusal evidence is recorded in `18-VERIFICATION.md`, with current accepted/refusal controls green.
- [x] Add source fixtures for a Result-returning computed match, a destructured payload-place return, and a pre-branch loan live in one arm.
- [x] Add focused requirement, independent-peer, comparator-axis, and wrong-slot mutation-kill tests.
- [x] Confirm the full five-axis evidence route and measure focused/full test latency.
- [x] Add or extend recurring CI coverage only where the focused regression value justifies maintenance and runtime cost; measured overlap supports existing macOS/Linux full and race lanes.

---

## Phase 18 Feedback and CI Evidence Record

Plan 08 fills the fenced key/value record below after measuring each named lane
with at least three cold and three warm runs. Times are seconds; include each
sample and matching count/min/median/max. The helper validates arithmetic and
provenance without imposing a time threshold. Record actual host, Go, Clang,
and pre-decision workflow blob hash; do not leave placeholders in final evidence.

Cold samples used a fresh isolated `GOCACHE` directory per run; warm samples
used the shared writable `/tmp/ai-lang-gocache`. The Go module cache and host
filesystem cache were retained. Every measured command completed successfully.

```text
phase18_evidence_version=1
host=macOS 26.6.2 (Darwin arm64)
go_version=go version go1.24.0 darwin/arm64
clang_version=Apple clang version 21.0.1 (clang-2100.1.1.101)
focused_command=go test ./internal/compiler/session -count=1
focused_cold_seconds=123.411,118.698,125.921
focused_cold_count=3
focused_cold_min_seconds=118.698
focused_cold_median_seconds=123.411
focused_cold_max_seconds=125.921
focused_warm_seconds=109.096,114.901,119.355
focused_warm_count=3
focused_warm_min_seconds=109.096
focused_warm_median_seconds=114.901
focused_warm_max_seconds=119.355
vet_command=go vet ./...
vet_cold_seconds=5.417,6.418,6.652
vet_cold_count=3
vet_cold_min_seconds=5.417
vet_cold_median_seconds=6.418
vet_cold_max_seconds=6.652
vet_warm_seconds=0.631,0.548,0.563
vet_warm_count=3
vet_warm_min_seconds=0.548
vet_warm_median_seconds=0.563
vet_warm_max_seconds=0.631
build_command=go build ./...
build_cold_seconds=3.363,4.039,3.660
build_cold_count=3
build_cold_min_seconds=3.363
build_cold_median_seconds=3.660
build_cold_max_seconds=4.039
build_warm_seconds=0.295,0.306,0.291
build_warm_count=3
build_warm_min_seconds=0.291
build_warm_median_seconds=0.295
build_warm_max_seconds=0.306
full_test_command=go test ./... -count=1
full_test_cold_seconds=147.574,146.818,142.119
full_test_cold_count=3
full_test_cold_min_seconds=142.119
full_test_cold_median_seconds=146.818
full_test_cold_max_seconds=147.574
full_test_warm_seconds=135.031,137.819,165.716
full_test_warm_count=3
full_test_warm_min_seconds=135.031
full_test_warm_median_seconds=137.819
full_test_warm_max_seconds=165.716
race_command=go test -race ./... -count=1
race_cold_seconds=283.896,314.999,307.101
race_cold_count=3
race_cold_min_seconds=283.896
race_cold_median_seconds=307.101
race_cold_max_seconds=314.999
race_warm_seconds=232.084,173.856,176.835
race_warm_count=3
race_warm_min_seconds=173.856
race_warm_median_seconds=176.835
race_warm_max_seconds=232.084
ci_before_blob=ca3c80918b5942e429d0cf8c64de86a9afe42de3
ci_disposition=not_added
ci_command=none
ci_decision_rationale=The focused native differential command has a 123.411s cold median and 114.901s warm median. Existing checks already run go test ./... and go test -race ./... on both Ubuntu and macOS, and the Phase 18 tests are included in those suites. A separate focused step would repeat the same session package coverage for about two additional minutes on each host without adding a distinct acceptance signal; retain the existing full and race lanes.
```

The verifier is phase-local at
`.planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh`.
`--measurements` validates all five lane distributions, actual host/tool
versions, and `ci_before_blob` against `git hash-object .github/workflows/ci.yml`.
`--final` additionally requires `ci_disposition=added` or `not_added`; for
`added`, the exact `ci_command` must appear in the existing `checks` job whose
matrix includes Ubuntu and macOS. For `not_added`, the workflow blob must still
match `ci_before_blob`, and the rationale must explain the measured tradeoff.

---

## Manual-Only Verifications

All objective phase behaviors have automated verification. Human review may assess code/API readability, but no acceptance criterion is delegated to manual UAT when source fixtures, peer checks, differential execution, mutation controls, and CI can establish it.

---

## Validation Sign-Off

- [x] All 20 plan tasks have an `<automated>` verify command or explicit Wave 0 dependency; each task is mapped above to a behavioral check or machine-checked evidence.
- [x] Sampling continuity: no 3 consecutive tasks without automated verification.
- [x] Wave 0 covers all planned tests and fixtures.
- [x] No watch-mode flags.
- [x] Feedback latency is measured and documented; recurring CI cost is justified.
- [x] The phase-local evidence verifier passes in both measurement and final-disposition modes.
- [x] `nyquist_compliant: true` set after Wave 0 and plan mapping are confirmed.

**Approval:** automated validation complete; all planned tasks map to green behavioral tests or machine-checked evidence. No human UAT remains for objective acceptance criteria.

## Validation Audit 2026-09-25

| Metric | Count |
|--------|-------|
| Planned tasks mapped | 20/20 |
| Automated behavioral/evidence checks green | 20/20 |
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

The previous map had six rows with stale command names and omitted planned tasks. It now maps each of the 20 tasks across plans 01–08. The consolidated adversarial acceptance command `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/session -run 'Test(GeneratedBranchBodiesRoundTrip|Phase18|LoanLiveness(Fixpoint|Bound|Cycle)|EdgeSpecificLiveOut)' -count=1` passed in all seven packages.

The parent verification run also passed `go vet ./...`, `go build ./...`, serialized `go test -p=1 ./... -count=1`, `go test -race ./... -count=1`, and the evidence verifier's `--measurements` and `--final` modes. An initial default-parallel full-suite run hit transient 5-second cache/native timeouts; the serialized full suite passed. No implementation change or additional CI step was needed.
