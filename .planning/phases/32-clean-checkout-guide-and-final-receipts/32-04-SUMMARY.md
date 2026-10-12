---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 04
subsystem: examples/checksum guide
tags: [dx-19, doc-test, clean-checkout, guide]
requires: ["32-03"]
provides:
  - examples/checksum/README.md as an executable guide (marked doc-test fences, case-keyed tables)
  - TestPhase32GuideCleanCheckout and TestPhase32GuideMatchesExpectedJSON
  - phase32ParseGuide, phase32CopyCheckout, phase32GuideOutputTargets, phase32GuideJSONProblems, phase32NativePolicy
affects: [32-07]
key-files:
  created:
    - cmd/schway/main_phase32_guide_test.go
  modified:
    - examples/checksum/README.md
key-decisions:
  - "README is the only source of commands and numbers; the test cross-checks them against checksum.expected.json, ChecksumOracleRows and Phase29FixtureExpected"
  - "Each row materialises its input in a per-app temp subdirectory because MaterializeChecksumCase reuses the case file name"
requirements-completed: []
status: complete
commits: 2
plan_head_before: 7018bbdbd45796d62f895bc19dac217474eb3412
plan_head_after: 739c4eea
actuals:
  tokens: 30000
  tasks: 2
  commits: 2
coverage:
  - id: SC1-D-01-D-02
    description: "The checksum guide runs line by line from a git ls-files clean-checkout copy (go build, three native builds, one app run per table row for checksum, adler_low and adler32_cli) and every result matches the README cell"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout"
        status: pass
    human_judgment: false
  - id: D-01-D-03
    description: "Limit, error exit, case names, stdout/stderr, adler_low = checksum mod 65536 and fixture numbers in the README agree with checksum.expected.json, the independent oracle rows and Phase29FixtureExpected; armed-json control observed to fail"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32GuideMatchesExpectedJSON"
        status: pass
    human_judgment: false
  - id: D-01a
    description: "Every guide output target and sidecar is under .schway/ and mkdir -p precedes the first build"
    requirement: "DX-19"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout"
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-10-12
---

# Phase 32 Plan 04: Executable checksum guide Summary

`examples/checksum/README.md` now carries marked `doc-test` fences that a test runs from a simulated clean checkout, with every table number cross-checked against the expected JSON and the independent oracle.

## Accomplishments

- README rewritten: setup fence (go build, `mkdir -p .schway/checksum`, three builds into `.schway/checksum/`), one run fence per app, case-keyed results table (7 cases), fixture table (3, 10), error prose line, Earlier programs and Known limits sections. The Limit and Every-error-exits sentences are kept verbatim.
- `TestPhase32GuideCleanCheckout` copies the checkout, runs the setup lines and 17 run lines (7 x 2 + 2 fixture rows), asserts exit/stdout/stderr per cell, absence of the secret path segment in error stderr, anti-vacuity counts and output-target confinement.
- `TestPhase32GuideMatchesExpectedJSON` is clang-free, with an `armed-json` subtest that mutates the limit.

## Deviations from Plan

None in behavior. Task 1 gated the anti-vacuity checks (at least 4 blocks, all three apps) behind a temporary constant because only one run fence existed at the tracer commit; Task 2 removed it.

## Known Stubs

None.

## Verification

Both Phase 32 guide tests, `go test ./cmd/schway`, the spawn lint, the session meta set and `TestValidationRowGradesAreEarnedOverArchivedCorpus` pass locally with clang. DX-19 is not marked complete (32-11 closes it).

## Self-Check: PASSED
