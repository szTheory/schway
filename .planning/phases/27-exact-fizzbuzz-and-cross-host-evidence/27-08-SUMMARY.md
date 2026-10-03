---
phase: 27-exact-fizzbuzz-and-cross-host-evidence
plan: 08
subsystem: testing
tags: [independent-oracles, occurrence-evidence, native-apps, capacity-controls]

requires:
  - phase: 27-exact-fizzbuzz-and-cross-host-evidence
    provides: FizzBuzz and sum application output with bounded stdout and separate evidence capacity
provides:
  - Independent literal-answer adjudication for interpreter and native results for both M005 apps
  - Reached FizzBuzz remainder, increment, and skipped-iteration mutation controls
  - Occurrence-aware event identity and fail-closed collision, omission, reorder, and capacity controls
  - Complete Unit application capture support with child outcome kept separate
affects: [phase-27-cross-host-receipts, phase-27-documentation, app-evidence]

actuals:
  tokens: 12015
  tasks: 2
  commits: 1

tech-stack:
  added: []
  patterns: [per-engine-literal-oracles, reached-source-mutations, occurrence-order-negative-controls]
key-files:
  created:
    - internal/compiler/session/session_phase27_evidence_test.go
    - internal/compiler/executionpeer/executionpeer_phase27_test.go
    - internal/compiler/interp/interp_phase27_evidence_test.go
    - internal/compiler/cgen/cgen_phase27_evidence_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_app.go
    - internal/compiler/native/native_app_test.go
    - internal/compiler/session/session_phase27_output_test.go
    - cmd/schway/main_test.go
key-decisions:
  - "Pin each interpreter and native answer to independently authored bytes before comparing the engines; equality alone never supplies a pass."
  - "Use FizzBuzz input 15 for the combined-remainder, altered-increment, and skipped-iteration mutants so each changed path reaches the fixed oracle."
  - "On a native stdout limit crossing, stop further app writes, record the failure once, serialize requested evidence, then return child exit 65."
  - "Validate retained app captures under a private returned-application contract that permits Unit's empty result while preserving the strict U64 conformance contract."
patterns-established:
  - "Occurrence controls run against real FizzBuzz interpreter and native traces at 0, 1, and 2 loop iterations before collision, omission, and reorder mutations."
  - "Application process outcome, app bytes, and evidence completeness remain separately asserted, including a failed child with complete evidence."
requirements-completed: [EVD-12, U64-02, APP-08, APP-09]
coverage:
  - id: D1
    description: "Each engine separately matches pinned literal answers for FizzBuzz 0, 1, 15, and 1000 and sum 0, 10, and 1000; both reject 1001 before output."
    requirement: APP-08
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase27(ExactAnswerMatrix|ReachedWrongRemainder|ReachedWrongArithmetic|SkippedIteration)$|^TestPhase26(ExactSumMatrix|ReachedWrongResult|SkippedIteration)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "Reached wrong-remainder, changed-increment, and skipped-first-iteration FizzBuzz mutations all fail the original independent 15-line oracle in both engines."
    requirement: U64-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase27_evidence_test.go:TestPhase27ReachedWrongRemainder, TestPhase27ReachedWrongArithmetic, TestPhase27SkippedIteration"
        status: pass
    human_judgment: false
  - id: D3
    description: "Overflow and dynamic-zero failures, repeated event identity, peer rejection controls, evidence capacity, and complete Unit capture are independently observable."
    requirement: EVD-12
    verification:
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./internal/compiler/executionpeer ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/native ./internal/compiler/session -run '^TestPhase27(OverflowZeroControls|RepeatedEventIdentity|PeerOccurrenceOrder|EvidenceCapacity|OutputEvidenceSeparation|LegacyDocumentCompatibility|PublicAppCLI)$|^TestPhase26(EventOccurrenceEncoding|PeerOccurrenceOrder|OverflowProcessOutcome|RepeatedCopyCapacity)$' -count=1"
        status: pass
    human_judgment: false
  - id: D4
    description: "Public CLI pins both apps' input bounds, output-limit status and stderr, and the dynamic-zero remainder failure channel."
    requirement: APP-09
    verification:
      - kind: integration
        ref: "cmd/schway/main_test.go:TestPhase27PublicAppCLI"
        status: pass
    human_judgment: false
  - id: D5
    description: "A stdout limit failure retains the completed prefix and exit 65 while a complete Unit execution capture remains available; evidence-only exhaustion does not change successful child streams."
    requirement: EVD-12
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase27_output_test.go:TestPhase27OutputEvidenceSeparation"
        status: pass
    human_judgment: false

# Metrics
duration: 34min
completed: 2026-10-03
tasks: 2
files_modified: 11
status: complete
---

# Phase 27 Plan 08: Independent Evidence Adjudication Summary

**Interpreter and native answers are now pinned independently for both apps, with reached semantic mutations and incomplete event or capture controls rejected.**

## Performance

- **Duration:** 34 min
- **Started:** 2026-10-03T17:24:10-04:00
- **Completed:** 2026-10-03T17:57:55-04:00
- **Tasks:** 2
- **Files modified:** 11

## Accomplishments

- Added an exact-answer matrix for FizzBuzz inputs 0, 1, 15, and 1000 and sum inputs 0, 10, and 1000. Interpreter output and native status/stdout/stderr are each checked against independent literals before the engine results are compared. Both applications reject 1001 without stdout.
- Added three reached FizzBuzz mutations: the multiple-of-15 remainder condition, scalar increment, and skipped first iteration. Every mutant parses, builds, executes in both engines at input 15, and differs from the original fixed answer.
- Checked exact U64 maximum success, checked-add overflow, and dynamic-zero remainder failure. Added real loop event identity controls for 0, 1, and 2 iterations plus peer refusals for collision, omitted first occurrence, and reordered events.
- Exercised lowered event-array and evidence-byte capacities while preserving successful child status and exact app output. Also verified output-limit exit 65 can coexist with complete Unit evidence, while evidence-only exhaustion does not alter the successful app process result.
- Expanded public CLI coverage for both committed apps, the dynamic-zero witness, exact accepted inputs, early 1001 refusal, and output-limit status/stderr. Confirmed the legacy U64 execution-document verifier still passes while Unit app bytes remain on the separate output channel.

## Task Commits

1. **Tasks 1–2: Add independent answer, mutation, event, and capacity evidence** - `21fb123` (feat)

**Plan metadata:** pending.

## Files Created/Modified

- `internal/compiler/session/session_phase27_evidence_test.go` - Pins both app answers, reached source mutants, U64 overflow/zero controls, repeated event identities, and legacy execution-document compatibility.
- `internal/compiler/executionpeer/executionpeer_phase27_test.go` - Validates real interpreter/native FizzBuzz traces and rejects repeated-copy collisions, omissions, and reordering.
- `internal/compiler/interp/interp_phase27_evidence_test.go` - Confirms 0/1/2 loop occurrence identity stays in execution evidence while app bytes remain separate.
- `internal/compiler/cgen/cgen_phase27_evidence_test.go` - Lowers generated event-array and evidence-byte capacities and checks incomplete evidence without changing child output.
- `internal/compiler/cgen/cgen.go` and `internal/compiler/cgen/cgen_program.go` - Delay native output-limit child exit until requested evidence has been serialized; block later app writes and retain the first exact stderr reason.
- `internal/compiler/native/native.go` and `internal/compiler/native/native_app.go` - Add a private application capture validation path for returned Unit documents without relaxing the conformance U64 outcome contract.
- `internal/compiler/native/native_app_test.go` and `internal/compiler/session/session_phase27_output_test.go` - Assert complete evidence with a failed child and keep app/evidence capacity verdicts separate.
- `cmd/schway/main_test.go` - Pins end-to-end CLI behavior for both apps, output limits, early input refusal, and dynamic zero.

## Decisions Made

- Literal answers are the independent adjudicator. Engine agreement is checked only after each engine passes its own expected bytes and status.
- The native writer blocks subsequent stdout writes after the first crossing chunk, then finishes normal function and resource cleanup before returning exit 65. This also leaves room for optional evidence serialization.
- Application capture accepts the generated Unit return shape only under an internal application-specific validation mode. Ordinary conformance execution still requires a non-empty U64 value.

## Deviations from Plan

- Changed the generated native output failure path to defer child exit until after optional evidence serialization. The plan requires a child failure even when evidence serialization succeeds; the previous immediate exit made that combination impossible.
- Added application-specific Unit capture validation and return-event field handling after the new FizzBuzz evidence case exposed that the existing runner applied its U64-only conformance contract to Unit app captures.

**Total deviations:** 2 necessary correctness updates.

## Issues Encountered

- The first peer and evidence run showed Unit app captures were reported incomplete even when generated JSON was present and valid for a Unit entry; decoding was applying the legacy non-empty U64 outcome check. The private app validation mode fixed this without changing the public evidence schema.
- The same run also showed immediate exit on the first output-limit crossing prevented requested evidence serialization. The generated C now blocks further app writes, emits the limit reason once, serializes evidence if requested, then exits 65.
- Both plan selectors passed, as did the focused output-bound, dynamic-zero, Phase 26 event/overflow, native exact-evidence-limit, generated-name, and emitter preflight controls. No full project suite was run.

## Next Plan Readiness

- Plan 27-09 can now produce host-bound receipts for both programs using a stable exact-output oracle and separately observed child/evidence outcomes.
- The remaining host receipts should bind the already-tested application sources and fixed fixtures to their actual host/compiler facts.

---
*Phase: 27-exact-fizzbuzz-and-cross-host-evidence*
*Completed: 2026-10-03*
