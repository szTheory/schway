---
phase: 26-checked-scalar-sum
plan: 04
subsystem: compiler-execution
tags: [execution, occurrence, interpreter, c17, independent-validation, ci]
requires:
  - phase: 26-02
    provides: independently checked scalar loop behavior and cross-engine arithmetic
  - phase: 26-03
    provides: scalar-cycle admission and authority refusals
provides:
  - Stable per-invocation static-site occurrence ordinals in interpreter and C17 event documents
  - Independent peer validation for contiguous scalar-copy cycle occurrences
  - Focused dual-host RED/GREEN evidence for occurrence behavior
affects: [26-05, execution-schema, interpreter, native-codegen, executionpeer]
actuals:
  tokens: 6728
  tasks: 2
  commits: 10
  plan_head_before: 35e057b62f3a60cc55cdb3fe92aec2ebbe546c0e
tech-stack:
  added: []
  patterns: [per-pair event ordinals, peer-derived scalar-cycle site admission, focused hosted TDD gates]
key-files:
  created: [.planning/phases/26-checked-scalar-sum/26-04-red-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-interp-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-cgen-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-peer-evidence.json]
  modified: [.github/workflows/ci.yml, internal/compiler/execution/execution.go, internal/compiler/executionpeer/executionpeer.go, internal/compiler/interp/interp.go, internal/compiler/cgen/cgen.go, internal/compiler/cgen/cgen_program.go, internal/compiler/native/native.go, testdata/phase16/public-emitter-consumers.json]
key-decisions:
  - "Occurrence zero remains implicit in JSON, preserving existing one-shot schema-2 bytes."
  - "The peer independently admits repeated events only for scalar OpCopy sites inside cyclic CFG blocks; it rejects non-contiguous ordinals and authority-bearing cycle events."
  - "Occurrence identity remains separate from bounded public application capture, which Plan 05 owns."
patterns-established:
  - "Interpreter and C17 producers count each (Invocation, ID) independently at event creation/recording boundaries."
  - "Peer cycle membership is derived from core CFG edges and typed places rather than producer occurrence claims."
requirements-completed: [FLOW-01, FLOW-02, APP-07]
coverage:
  - id: D1
    description: "Interpreter and native C17 preserve return value 6 and emit occurrences 0/1/2 for three scalar loop copies, while input zero emits no copy event and ordinal zero stays omitted from canonical JSON."
    requirement: FLOW-01
    verification:
      - kind: unit
        ref: "GitHub Actions run 37088525147: TestPhase26EventOccurrenceEncoding, TestPhase26InterpreterRepeatedCopy, TestPhase26NativeRepeatedCopy (Ubuntu and macOS)"
        status: pass
    human_judgment: false
  - id: D2
    description: "The independent execution peer accepts a complete 0/1/2 scalar-copy cycle and rejects duplicate, gap, swapped, wrong-invocation, wrong-site, and non-scalar-cycle traces."
    requirement: FLOW-02
    verification:
      - kind: unit
        ref: "GitHub Actions run 37088956514: TestPhase26PeerOccurrenceOrder and all six mutation controls (Ubuntu and macOS)"
        status: pass
    human_judgment: false
duration: 53min
completed: 2026-10-03
status: complete
---

# Phase 26 Plan 04: Checked Scalar Sum Summary

**Repeated scalar-copy events now carry canonical per-invocation occurrence identity from both engines, and an independent peer validates each ordinal and cycle site.**

## Performance

- **Duration:** 53 min
- **Started:** 2026-10-03T01:21:19Z
- **Completed:** 2026-10-03T02:13:54Z
- **Tasks:** 2
- **Files modified:** 16 plan-delivery files

## Accomplishments

- Added optional `occurrence` to schema-2 events, with zero omitted so existing one-shot canonical bytes remain unchanged.
- Interpreter and C17 independently assign checked ordinal sequences per `(Invocation, ID)`. The observable `sum_to_n` copy site returns 6 and emits 0/1/2 for input 3, while input 0 emits no copy events.
- Updated native decoding and the independent peer to accept only contiguous repeated scalar-copy occurrences within a derived cyclic CFG site; duplicate, missing, reordered, wrong-site, wrong-invocation, and non-scalar controls remain refused.
- Expanded focused CI selectors to execute all named occurrence tests before broader regression jobs, while keeping the public emitter consumer registry aligned.

## Task Commits

1. **Task 1: Produce canonical dynamic event occurrences in both engines** - `9bf337b`, `788bb1a`, `b74398a`, `cc49e11`, `31a6e47`, `461bf07` (RED tests/selector and fixture refinements), `0fb36dc` (implementation), `ba09f05` (native decoder fix).
2. **Task 2: Independently validate occurrence order and membership** - `0f08696` (RED test), `e88ea82` (peer implementation).

**Plan metadata:** captured in the GSD closeout commit after state and roadmap updates.

## Files Created/Modified

- `internal/compiler/execution/execution.go` and `execution_test.go` - optional occurrence coordinate and canonical zero omission coverage.
- `internal/compiler/interp/interp.go` and `interp_test.go` - interpreter ordinals and repeated-copy behavior.
- `internal/compiler/cgen/cgen.go`, `cgen_program.go`, and `cgen_program_test.go` - C17 event recording, schema writer, and native repeated-copy coverage.
- `internal/compiler/native/native.go` - strict contiguous ordinal validation while decoding native event documents.
- `internal/compiler/executionpeer/executionpeer.go` and `executionpeer_test.go` - independent loop-site derivation, occurrence sequence validation, and mutation controls.
- `.github/workflows/ci.yml` - focused execution, interpreter, C generator, peer, and existing regression selectors.
- `testdata/phase16/public-emitter-consumers.json` - updated the moved callsite and classified the new schema-2 consumer.
- `.planning/phases/26-checked-scalar-sum/26-04-red-*.json` - four hosted RED transcripts, including peer occurrence-order failure.

## Decisions Made

- The first event at a static ID and invocation uses ordinal zero, which is omitted from serialized JSON. Each producer counts independently and fails closed before ordinal overflow.
- The peer derives eligible cyclic operations from CFG edges, typed source/target places, and loan facts. Only U64/Bool scalar `OpCopy` events may repeat; Plan 03 authority refusals remain in force.
- Occurrence identity is independent of capture capacity. Plan 05 retains ownership of the public application capacity boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Native decoder needed to accept contiguous repeated event pairs**
- **Found during:** Task 1 hosted GREEN
- **Issue:** The producer emitted valid repeated `(invocation, ID)` pairs, but the native event decoder still rejected all duplicate pairs.
- **Fix:** Replaced the blanket duplicate-pair refusal with contiguous occurrence validation and overflow protection.
- **Files modified:** `internal/compiler/native/native.go`
- **Verification:** Focused hosted run 37088525147 passed on Ubuntu and macOS.
- **Committed in:** `ba09f05`

**2. [Rule 2 - Critical functionality] Focused CI and public emitter inventory needed occurrence consumer coverage**
- **Found during:** Task 1 RED selector setup and emitter inventory validation
- **Issue:** Existing focused CI selectors did not run all named tests, and the explicit consumer inventory had to track the new native event writer callsite.
- **Fix:** Expanded the focused selectors and updated the registry classification for the consumer call.
- **Files modified:** `.github/workflows/ci.yml`, `testdata/phase16/public-emitter-consumers.json`
- **Verification:** Hosted RED and GREEN focused workflows ran the named tests on both hosts; the emitter inventory test passed in the focused GREEN run.
- **Committed in:** `9bf337b`, `b74398a`, `0fb36dc`

**Total deviations:** 2 auto-fixed (1 Rule 2, 1 Rule 3). Both were required to execute the planned evidence gates and accept valid native traces. Plan scope did not expand.

## TDD Gate Compliance

- Task 1 RED was recorded in `.planning/phases/26-checked-scalar-sum/26-04-red-evidence.json`, `26-04-red-interp-evidence.json`, and `26-04-red-cgen-evidence.json`. `check tdd-red-evidence` returned `RED_EVIDENCE_OK`; hosted run 37088120033 failed the expected encoding, interpreter occurrence, and C17 copy-site assertions on both hosts.
- Task 2 RED was recorded in `26-04-red-peer-evidence.json`. `check tdd-red-evidence` returned `RED_EVIDENCE_OK`; run 37088704306 failed because the peer rejected the valid second occurrence as a duplicate pair on both hosts.
- Task 1 GREEN: hosted run 37088525147 passed the focused encoding/interpreter/native tests on Ubuntu and macOS.
- Task 2 and full focused GREEN: hosted run 37088956514 passed on Ubuntu and macOS, including all six peer mutation controls. No local Go test suite, native program, or verification script was run.
- Compile-only `GOCACHE=/tmp/schway-phase26-build-cache go build ./...`, `gofmt`, and `git diff --check` passed locally.

## Issues Encountered

Several early RED attempts exposed incomplete fixtures or a too-narrow CI selector and were not treated as valid RED evidence. The final RED transcripts preserve the expected assertion failures; the final focused GREEN run passed the intended tests on both hosts.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 05 can now verify public application capture capacity while relying on stable event identity and independent peer sequence checks. This summary does not mark Phase 26 complete; Plan 05 remains.

---
*Phase: 26-checked-scalar-sum*
*Completed: 2026-10-03*

## Self-Check: PASSED

- Summary file exists at the planned path.
- All ten task commits listed above are present in Git history.
- The measured plan commit count is 10 from the recorded base SHA.
