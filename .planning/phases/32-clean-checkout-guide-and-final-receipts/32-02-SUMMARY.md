---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 02
subsystem: ci/session-meta-tests
tags: [evd-15, ci, goflags, test-index-cache]
requires: []
provides:
  - GOFLAGS -timeout=20m on the hosted verify-phase6 step
  - SCHWAY_REQUIRE_NATIVE=1 job env on checks and evidence-aggregate
  - once-per-process cached static test index with deep-copy hand-out
affects: [32-04, 32-07]
key-files:
  created:
    - internal/compiler/session/session_phase32_ci_test.go
  modified:
    - .github/workflows/ci.yml
    - internal/compiler/session/verification_groundedness_test.go
key-decisions:
  - "Timeout budget goes in the CI step env, so scripts/verify-phase6.sh stays byte-unchanged"
  - "buildTestIndex caches only the untagged parse; tagged variants stay uncached"
status: complete
commits: 2
plan_head_before: 47b9c702a42d10115b6635f478a8233271bb1623
plan_head_after: 13cdbd0ae20d0f3afea23090d10bab766c9a52ed
actuals:
  tokens: 6000
  tasks: 2
  commits: 2
coverage:
  - id: ci-flake-timeout
    description: "The hosted verify-phase6 step carries GOFLAGS -timeout=20m, with an armed control that fails when the line is removed"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_ci_test.go#TestPhase32HostedCIEnv"
        status: pass
    human_judgment: false
  - id: D-04a
    description: "checks and evidence-aggregate both set SCHWAY_REQUIRE_NATIVE 1, with an armed control"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_ci_test.go#TestPhase32HostedCIEnv"
        status: pass
    human_judgment: false
  - id: test-index-cache
    description: "The static test index is parsed once per process, each caller gets an isolated deep copy equal to a fresh parse and to go test -list"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase32_ci_test.go#TestPhase32TestIndexCacheIsolation"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/verification_groundedness_test.go#TestStaticTestIndexMatchesGoTestList"
        status: pass
    human_judgment: false
duration: 10min
completed: 2026-10-12
---

# Phase 32 Plan 02: Hosted CI timeout budget Summary

The hosted `current evidence aggregate` job now runs `scripts/verify-phase6.sh` with `GOFLAGS=-timeout=20m` (fixing the default 10m package timeout from run 38302325295), both Go jobs require native, and the session meta-tests share one cached test index.

## Accomplishments

- `ci.yml`: step env `GOFLAGS: -timeout=20m` on the verify-phase6 step; job env `SCHWAY_REQUIRE_NATIVE: "1"` on `checks` and `evidence-aggregate`. `scripts/verify-phase6.sh` is untouched.
- `TestPhase32HostedCIEnv` with `armed` and `armed-native` controls (both observed passing, meaning the mutations are detected).
- `buildTestIndex` uses `sync.Once` and returns a deep copy; `TestPhase32TestIndexCacheIsolation` proves isolation and equality with a fresh parse.
- Observation: the meta lane (`Groundedness|TestSummaryCoverage|TestStaticTestIndex`) took 2.7s before and 2.3s after on this machine (not race), so the gain is small locally; the 20m budget is the real fix.
- GOFLAGS semantics: `go vet` and `go build` exit 0 with `GOFLAGS=-timeout=20m`. `go test` honors a GOFLAGS timeout (whole-package `-timeout=1ms` run panics with "test timed out").

## Deviations from Plan

The plan's third verify command (`-run '^TestFormatIdempotent$'` with 1ms) prints 0 because that test finishes in under 1ms. Rerunning against the whole syntax package at 1ms proved the timeout is honored. No code change.

## Verification

- The new, Phase 27, Phase 6 and CI-gate pins pass. The session meta set passes.

## Known Stubs

None.

## Self-Check: PASSED

Created test file exists; commits 06f96bbf and 13cdbd0a are on the branch.
