---
phase: "18"
slug: "branch-on-a-computed-value"
status: validated
nyquist_compliant: false
wave_0_complete: false
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
| 18-01-T1 | 01 | 0 | CTL-01 | T-18-01 | Pin the existing computed-scrutinee refusal; independent validators fail closed for malformed or forged core | fixture, unit | `go test ./internal/compiler/check -run '^TestPhase18ComputedScrutineeProductionPathAccepted$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 18-01-T2 | 01 | 0 | CTL-02 | T-18-02 | A Result-returning callee's computed value reaches all five comparator axes through independent peer admission | integration, differential, smoke | `go test ./internal/compiler/session -run '^TestPhase18ResultComputedMatch$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 18-01-T3 | 01 | 0 | CTL-03 | T-18-03 | Payload place is returned; injected wrong-slot corruption diverges at `axis:terminal-outcome` | integration, mutation-kill | `go test ./internal/compiler/session -run '^TestPhase18ComputedSourceFourTierDifferential$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 18-01-T4 | 01 | 0 | CTL-01 | T-18-01 | A pre-match loan live in one arm is classified with existing endpoints and bounded fixpoint work | source integration, ownership | `go test ./internal/compiler/session -run '^TestPhase18WrongSlotMutation$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 18-08-T1 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-01 | Repeated cold/warm distributions and host/tool provenance are recorded and internally consistent; plan 08 will deliver the phase-local evidence verifier | latency evidence | `go test ./internal/compiler/session -run '^TestPhase18FiveAxis$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |
| 18-08-T2 | 08 | 6 | CTL-01, CTL-02, CTL-03 | T-18-03 | CI disposition is evidence-backed; added command runs in both matrix hosts or unchanged CI blob is proven; plan 08 will deliver the phase-local evidence verifier | CI configuration evidence | `go test ./internal/compiler/session -run '^TestPhase18ComparatorControlRejectsMissingExecutionTierAndRequiresPeer$' -count=1` | ❌ W0 | EXERCISED | — | ⬜ pending |

Threat refs:

- **T-18-01:** Untrusted input causes unbounded parser, CFG, or liveness work; preserve limits and fail closed at `4 × blocks × (loans+1)`.
- **T-18-02:** Checker and independent admission peers derive computed-place facts differently; independently validate place, type, and ownership.
- **T-18-03:** Native code writes the wrong payload slot but the comparator misses it; require mutation injection and terminal-outcome divergence.

The five-axis acceptance must use the full session/program comparator and peer-validation route. `Phase5CompareEngines` alone covers only four execution-bearing axes; diagnostic-ID refusal comparison is separate.

---

## Wave 0 Requirements

- [ ] Add and pin the refused CTL-01 computed-place `.lang` frontier fixture before changing production admission.
- [ ] Add source fixtures for a Result-returning computed match, a destructured payload-place return, and a pre-branch loan live in one arm.
- [ ] Add focused requirement, independent-peer, comparator-axis, and wrong-slot mutation-kill tests.
- [ ] Confirm the full five-axis evidence route and measure focused/full test latency.
- [ ] Add or extend recurring CI coverage only where the focused regression value justifies maintenance and runtime cost; use existing macOS/Linux CI lanes.

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

- [ ] All plan tasks have an `<automated>` verify command or explicit Wave 0 dependency.
- [ ] Sampling continuity: no 3 consecutive tasks without automated verification.
- [ ] Wave 0 covers all missing tests and fixtures above.
- [ ] No watch-mode flags.
- [ ] Feedback latency is measured and documented; recurring CI cost is justified.
- [ ] The phase-local evidence verifier passes in both measurement and final-disposition modes.
- [ ] `nyquist_compliant: true` set after Wave 0 and plan mapping are confirmed.

**Approval:** pending plan and Wave 0 validation.


---

Post-G-18-16 default-parallel and aggregate verification receipts.

```phase18-postfix
version=1
run_count=7
host=macOS-26.6.2-arm64-arm-64bit-Mach-O
go_version=go version go1.24.0 darwin/arm64
clang_version=Apple clang version 21.0.0 (clang-2100.1.1.101)
source_blob_internal/compiler/cache/cache.go=c16a3647a69fae03782719631ebd06de9011c689
source_blob_internal/compiler/cache/probe.go=0da718c2e64e3a2155838a66ed9b79204fc9ad6a
source_blob_internal/compiler/cache/probe_test.go=146e0246a42b580d509662b08b23a50074344383
source_blob_internal/compiler/native/native.go=383ee701c669ad437c749a6ed5ae64f0441a4941
source_blob_internal/compiler/native/native_test.go=f01a09aff1c2e5a9dad38bfc3b7e747d2e4a373e
run_1_lane=default_1
run_1_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_1_exit=1
run_1_duration=268.5899107080186
run_1_offset=0
run_1_length=2156
run_1_digest=c58fd1475740e1c1aaab2845d03314f92bbcfe6369eff5d37addb077060860ab
run_2_lane=default_2
run_2_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_2_exit=1
run_2_duration=207.01748933293857
run_2_offset=2156
run_2_length=2155
run_2_digest=b31d2315083c570b1af9f985bda4e30b950e6148e5752804e193cfd82c5baceb
run_3_lane=default_3
run_3_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_3_exit=1
run_3_duration=191.05879808298778
run_3_offset=4311
run_3_length=2155
run_3_digest=bfbca21b3b2976096c7454b38a3b3453eabb403508970822bd767ff5203b2fb2
run_4_lane=parallel_4
run_4_command=GOCACHE=/tmp/ai-lang-gocache go test -p=4 ./... -count=1
run_4_exit=1
run_4_duration=198.35185612505302
run_4_offset=6466
run_4_length=2158
run_4_digest=5d2dc1b741f0e295c5fc08b3dd71ac358acfe94a0707aa158d89b2bfabd27114
run_5_lane=race
run_5_command=GOCACHE=/tmp/ai-lang-gocache go test -race ./... -count=1
run_5_exit=1
run_5_duration=369.1534314589808
run_5_offset=8624
run_5_length=2382
run_5_digest=86df8a4566c430041b3ae9dfb440d3f49632d45917425bbe37731dcda3ed22b4
run_6_lane=vet
run_6_command=go vet ./...
run_6_exit=1
run_6_duration=0.3890865419525653
run_6_offset=11006
run_6_length=1967
run_6_digest=cadf69946c07548b14bd319a0a3ab30c80de95d35cb57db3162863a61f448bda
run_7_lane=build
run_7_command=go build ./...
run_7_exit=1
run_7_duration=0.5095293750055134
run_7_offset=12973
run_7_length=1971
run_7_digest=364482e1b69776f8ddff06b0ae070f92436d4d706a53a352321ada15e2cef95a
```


---

Post-G-18-16 default-parallel and aggregate verification receipts.

```phase18-postfix
version=1
attempt_id=20260925T182259Z-90626
run_count=7
host=macOS-26.6.2-arm64-arm-64bit-Mach-O
go_version=go version go1.24.0 darwin/arm64
clang_version=Apple clang version 21.0.0 (clang-2100.1.1.101)
source_blob_internal/compiler/cache/cache.go=c16a3647a69fae03782719631ebd06de9011c689
source_blob_internal/compiler/cache/probe.go=0da718c2e64e3a2155838a66ed9b79204fc9ad6a
source_blob_internal/compiler/cache/probe_test.go=146e0246a42b580d509662b08b23a50074344383
source_blob_internal/compiler/native/native.go=383ee701c669ad437c749a6ed5ae64f0441a4941
source_blob_internal/compiler/native/native_test.go=f01a09aff1c2e5a9dad38bfc3b7e747d2e4a373e
run_1_lane=default_1
run_1_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_1_exit=0
run_1_duration=282.7226976250531
run_1_offset=14944
run_1_length=1991
run_1_digest=6aa64010f424abfeaa6240e92316b8153d321d537cf9d4e3b9afbb90223c2f86
run_2_lane=default_2
run_2_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_2_exit=0
run_2_duration=232.75530312489718
run_2_offset=16935
run_2_length=1989
run_2_digest=b51924e24e50a591db1f662e87545ce651db920a0cb01938b2282109446cc095
run_3_lane=default_3
run_3_command=GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1
run_3_exit=0
run_3_duration=275.61037775001023
run_3_offset=18924
run_3_length=1989
run_3_digest=f10005d2a013ef165b0dc3d5b13eaf1633863bf42a6c163134863bb1f36cc715
run_4_lane=parallel_4
run_4_command=GOCACHE=/tmp/ai-lang-gocache go test -p=4 ./... -count=1
run_4_exit=0
run_4_duration=243.86262187501416
run_4_offset=20913
run_4_length=1994
run_4_digest=63e2796b181b31a598cf30558a3c6f55d2d2a40bb23f4f97a4426dfac721b410
run_5_lane=race
run_5_command=GOCACHE=/tmp/ai-lang-gocache go test -race ./... -count=1
run_5_exit=0
run_5_duration=198.48179300001357
run_5_offset=22907
run_5_length=1991
run_5_digest=047a7ca91a02b8a1a64e055b77d33edc67f81eccba107502fe7acf74cea792d4
run_6_lane=vet
run_6_command=go vet ./...
run_6_exit=0
run_6_duration=0.42519120895303786
run_6_offset=24898
run_6_length=92
run_6_digest=0ca914e3f228c294aa10ea4a10cf569b56359a5cb250501a268387e56d3b5ef6
run_7_lane=build
run_7_command=go build ./...
run_7_exit=0
run_7_duration=0.25219454104080796
run_7_offset=24990
run_7_length=96
run_7_digest=da99a749654165602f63f8157048c0a126e743814da9f049f8c66c5e04705987
```
