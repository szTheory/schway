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
  tokens: 19417
  tasks: 2
  commits: 13
  plan_head_before: 35e057b62f3a60cc55cdb3fe92aec2ebbe546c0e
tech-stack:
  added: []
  patterns: [per-pair event ordinals, peer-derived scalar-cycle site admission, focused hosted TDD gates]
key-files:
  created: [.planning/phases/26-checked-scalar-sum/26-04-red-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-interp-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-cgen-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-red-peer-evidence.json, .planning/phases/26-checked-scalar-sum/26-04-integration-red-evidence.json]
  modified: [.github/workflows/ci.yml, internal/compiler/execution/execution.go, internal/compiler/executionpeer/executionpeer.go, internal/compiler/interp/interp.go, internal/compiler/cgen/cgen.go, internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_names_test.go, internal/compiler/native/native.go, internal/compiler/session/session_phase5_compare.go, internal/compiler/session/session_phase5_compare_test.go, testdata/phase16/public-emitter-consumers.json]
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

**Core-level repeated scalar-copy events carry canonical per-invocation occurrence identity from both engines, and an independent peer validates each ordinal and cycle site.**

## Performance

- **Duration:** 53 min
- **Started:** 2026-10-03T01:21:19Z
- **Completed:** 2026-10-03T02:13:54Z
- **Tasks:** 2
- **Files modified:** 24 plan-delivery files

## Accomplishments

- Added optional `occurrence` to schema-2 events, with zero omitted so existing one-shot canonical bytes remain unchanged.
- Interpreter and C17 independently assign checked ordinal sequences per `(Invocation, ID)`. The fixtures inject `OpCopy` into checked loop core, return 6 and emit 0/1/2 for input 3, while input 0 emits no copy events. This proves core/backend behavior, not a source-level `let snapshot = i` witness.
- Updated native decoding and the independent peer to accept only contiguous repeated scalar-copy occurrences within a derived cyclic CFG site; duplicate, missing, reordered, wrong-site, wrong-invocation, and non-scalar controls remain refused.
- Expanded focused CI selectors to execute all named occurrence tests and affected packages, while keeping the public emitter consumer registry aligned.

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

**3. [Rule 1 - Regression] Occurrence writer support changed unrelated generated C and broke integration contracts**
- **Found during:** Wave 3 full hosted gate 37089290645
- **Issue:** Unconditional schema-2 occurrence support changed legacy bytes and required exact-width identifiers in programs without U64 support; the new field also needed comparison routing, namespace reservations, actionable duplicate diagnostics, and refreshed consumer locations.
- **Fix:** Gate occurrence emission on an eligible cyclic scalar-copy site; preserve prior writer output otherwise, retain duplicate-pair refusal codes, route `Occurrence` through semantic comparison, reserve generated identifiers, and update every moved emitter consumer call.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `cgen_program.go`, `cgen_names_test.go`, `internal/compiler/executionpeer/executionpeer.go`, `internal/compiler/session/session_phase5_compare.go`, `session_phase5_compare_test.go`, `testdata/phase16/public-emitter-consumers.json`, `.github/workflows/ci.yml`
- **Verification:** Full-package focused run 37090502900 passed cgen and executionpeer on both hosts and passed Phase 26 occurrence, authority, and order tests. Session package completed with pre-existing corpus/frontier/lifecycle failures; Ubuntu also lacked `-lc++` for unrelated Phase 5 verification tests. Full gate 37090678253 is recorded below.
- **Committed in:** `fa85fac`, `9784f76`

**Total deviations:** 2 auto-fixed (1 Rule 2, 1 Rule 3). Both were required to execute the planned evidence gates and accept valid native traces. Plan scope did not expand.

## TDD Gate Compliance

- Task 1 RED was recorded in `.planning/phases/26-checked-scalar-sum/26-04-red-evidence.json`, `26-04-red-interp-evidence.json`, and `26-04-red-cgen-evidence.json`. `check tdd-red-evidence` returned `RED_EVIDENCE_OK`; hosted run 37088120033 failed the expected encoding, interpreter occurrence, and C17 copy-site assertions on both hosts.
- Task 2 RED was recorded in `26-04-red-peer-evidence.json`. `check tdd-red-evidence` returned `RED_EVIDENCE_OK`; run 37088704306 failed because the peer rejected the valid second occurrence as a duplicate pair on both hosts.
- Task 1 GREEN: hosted run 37088525147 passed focused encoding/interpreter/native tests on Ubuntu and macOS; these are checked-core fixtures that inject `OpCopy`, not public source witnesses.
- Task 2 GREEN: hosted run 37088956514 passed on Ubuntu and macOS, including all six peer mutation controls.
- Wave 3 RED: full hosted run 37089290645 failed on both hosts after the integration regressions listed in `.planning/phases/26-checked-scalar-sum/26-04-integration-red-evidence.json`.
- Integration correction: focused run 37090502900 on both hosts passed cgen and executionpeer full packages, Phase 26 occurrence/authority tests, and `TestPhase16PublicEmitterConsumerInventory`; its whole session package still fails on the known validation-corpus digest/frontier/reconciliation/lifecycle set, with Ubuntu additionally failing Phase 5 tests because Clang cannot find `-lc++` in that focused job.
- Final full gate 37090678253 passed vet and build on both hosts. Full tests failed on both hosts only in the active Phase 26 evidence/lifecycle set: stale corpus digest, an unreconciled R2 row, active R2b rows, draft lifecycle status, and the corresponding non-empty R1/R2/R3/R2b groundedness audit. Phase 23/24/25 aggregate checks passed; Phase 6 aggregates failed on the same active Phase 26 rows. Race checks, Phase 15 aggregate checks, and the corpus receipt job were skipped by workflow dependencies after failures.
- No local Go test suite, native program, or verification script was run.
- Compile-only `GOCACHE=/tmp/schway-phase26-build-cache go build ./...`, `gofmt`, and `git diff --check` passed locally.

## Issues Encountered

Several early RED attempts exposed incomplete fixtures or a too-narrow CI selector and were not treated as valid RED evidence. The focused GREEN run passed the occurrence tests on both hosts. Plan 04 did not produce the requested public source witness: its core tests inject `OpCopy` into checked loop core. Plan 05 owns the source-level `let snapshot = i` consumer and must not assume that source form is established by these tests.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 05 owns both the public source-copy witness (`let snapshot = i`) and application capture capacity. It can rely on stable event identity and independent peer sequence checks after implementing that source consumer. This summary does not mark Phase 26 complete; Plan 05 remains.

---
*Phase: 26-checked-scalar-sum*
*Completed: 2026-10-03*

## Self-Check: PASSED

- Summary file exists at the planned path.
- The ten original task commits and integration corrections `fa85fac`, `9784f76` are present in Git history.
- The measured commit count through the integration correction is 13 from the recorded base SHA, including the prior metadata closeout and both integration commits; this SUMMARY/state metadata update is a separate commit.
