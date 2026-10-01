---
phase: 24-ownership-transfer-through-calls-and-errors
plan: "03"
subsystem: compiler
tags: [ownership, typed-errors, physical-cleanup, native-observer, hosted-evidence]

# Dependency graph
requires:
  - phase: 24-02
    provides: Repeated activation identity, typed-error cleanup across frames, and the bounded native error application.
provides:
  - Independent physical observation of transferred owner use and normal/error cleanup.
  - Five reached native destructor controls that remain decisive when compiler events are plausible.
  - Model-only Phase 24 replay and a focused Ubuntu/macOS evidence receipt.
affects: [phase-24-verification, phase-25, ownership-evidence, native-emission]

# Actuals (#2632)
actuals:
  tokens: 9209
  tasks: 2
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Keep actual host pointers private to the observer; publish stable acquisition operation and activation identities.
    - Separate physical native lifecycle evidence from deterministic model outcomes and compiler event claims.

key-files:
  created:
    - internal/compiler/native/phase24_observer_test.go
    - internal/compiler/native/testdata/phase24_observer.c
    - examples/phase24/README.md
    - scripts/verify-phase24.sh
  modified:
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/session/session_phase24_model_test.go
    - .github/workflows/ci.yml

key-decisions:
  - "Match allocation, use, and destruction with private pointers while receipts expose only stable semantic operation and activation IDs."
  - "Keep interpreter replay explicitly model-only with actual_host_io=false and physical_cleanup=false."
  - "Instrument generated transfer-return and typed-error reporting boundaries only in the test harness; the observer is the physical cleanup oracle."

patterns-established:
  - "A physical cleanup control must reach the changed operation and fail independently of plausible compiler release events."
  - "Hosted evidence receipts bind host, target, source revision, clean tree, and focused elapsed time."

requirements-completed: [EVD-09, RES-05, RES-06, OWN-11, OWN-12]
coverage:
  - id: D1
    description: The native observer proves post-transfer use, matching physical release, C,B,A cleanup order on the typed-error path, and zero outstanding allocations.
    requirement: EVD-09
    verification:
      - kind: integration
        ref: "CI 36852275511: TestPhase24ObserverPublicLifecycleAndTypedError on Ubuntu and macOS"
        status: pass
    human_judgment: false
  - id: D2
    description: Omitted, premature, duplicate, wrong-resource, and semantic-identity-collision controls are reached and rejected by the physical observer.
    requirement: EVD-09
    verification:
      - kind: integration
        ref: "CI 36852275511: TestPhase24ObserverReachedPhysicalDestructorControls on Ubuntu and macOS"
        status: pass
    human_judgment: false
  - id: D3
    description: Deterministic replay retains model-only claims, and the focused Phase 24 aggregate passes with host-bound receipts on both required hosts.
    verification:
      - kind: integration
        ref: "CI 36852275511: scripts/verify-phase24.sh passed on Linux/x86_64 (28s) and Darwin/arm64 (20s) at c0d6a17ad51a952aaa11921dcc0f05ab8bc60bf2"
        status: pass
      - kind: integration
        ref: "CI 36852275511: go vet, go build, go test, and go test -race passed on Ubuntu and macOS"
        status: pass
    human_judgment: false

# Metrics
duration: 600min
completed: 2026-10-01
status: complete
---

# Phase 24 Plan 03: Independent Physical Transfer and Hosted Evidence Summary

**A private-pointer native observer proves transferred owner use and C,B,A physical cleanup, while model replay and dual-host receipts make no unsupported host or cleanup claims.**

## Performance

- **Duration:** 10 hours wall time, including the hosted CI wait
- **Started:** 2026-09-30T21:12:12-04:00
- **Completed:** 2026-10-01T07:12:27-04:00
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments

- Added an independent C observer for the generated Phase 24 application. It privately matches actual allocation, post-transfer use, and physical destruction, proves the `0x41`/`0x42` normal cases and the `0x43` typed-error path, observes exact C,B,A release order, and confirms zero outstanding allocations before exit.
- Added five reached physical controls: omitted, premature, duplicate, wrong-resource, and semantic-identity-collision destruction. Each fails under the physical observer even while plausible compiler events remain available to the semantic checks.
- Added deterministic replay cases with independently fixed answers and `actual_host_io=false` / `physical_cleanup=false`, public scope documentation, and a seven-group focused verifier in the existing Ubuntu/macOS evidence aggregate.

## Hosted Verification

Hosted workflow [CI run 36852275511](https://github.com/szTheory/schway/actions/runs/36852275511) completed successfully at source revision `c0d6a17ad51a952aaa11921dcc0f05ab8bc60bf2` on branch `worktree-agent-p24-01-retry`.

- Ubuntu: full `go vet ./...`, `go build ./...`, `go test ./...`, and race suite passed. `scripts/verify-phase24.sh` passed in 28 seconds on Linux/x86_64 (`linux/amd64`, target `x86_64-pc-linux-gnu`); the host receipt recorded a clean tree and the source revision.
- macOS: the same full and race suites passed. `scripts/verify-phase24.sh` passed in 20 seconds on Darwin/arm64 (`darwin/arm64`, target `arm64-apple-darwin25.6.0`); the host receipt recorded a clean tree and the source revision.
- Both current evidence aggregates also passed the existing Phase 23, Phase 6, and Phase 15 evidence steps. The optional validation-corpus receipt was skipped for this manual dispatch and is not claimed.
- No local project tests or scripts were run; this plan requires hosted-only automated validation. `gofmt` and `git diff --check` were used during implementation.

## Task Commits

1. **Observe physical transfer, cleanup and five reached mutations** — `3828e9c6` (`test(24-03): observe physical transfer cleanup`)
2. **Bind model-only replay and public contract to the existing two-host CI lane** — `c0d6a17a` (`feat(24-03): bind replay and hosted evidence contract`)

## Decisions Made

- Keep host addresses private to the observer and publish only stable operation and activation identities.
- Keep model replay and physical native observation as distinct evidence types; interpreter output never implies actual host IO or physical destruction.
- Keep the observer's transfer-return and error-report callbacks test-only, without adding a production event serializer.

## Deviations from Plan

The special Phase 24 typed-error application has no serialized compiler event stream. The test harness therefore inserts callbacks at generated helper-return and typed-error-report boundaries and uses the instrumented adapter's actual pointer observations as the physical oracle. It does not present an event stream as proof of cleanup. Semantic model events and their negative controls remain separately checked. This evidence boundary is documented in `examples/phase24/README.md`; it does not weaken the physical lifecycle claim.

**Total deviations:** 1 documented evidence-boundary adjustment. No production scope was added.

## Issues Encountered

No implementation or CI failures remained. The workflow was started with `workflow_dispatch` on the feature branch; the optional validation-corpus job was skipped and is outside this plan's acceptance criteria.

## User Setup Required

None.

## Next Phase Readiness

Plan 24-03's hosted evidence is complete. Phase-level review, regression, and goal-verification gates remain before Phase 24 is marked complete. After those gates, Phase 25 is the next capability: separate bounded shared/exclusive read-copy pointer witnesses and the integrated utility.

---
*Phase: 24-ownership-transfer-through-calls-and-errors*
*Completed: 2026-10-01*
