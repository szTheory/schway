---
phase: 06-agent-feedback-and-performance-ratification
plan: 10
subsystem: performance-ratification
tags: [go, stage-attribution, timing, honest-unavailability, go-ast]

requires:
  - phase: 06-agent-feedback-and-performance-ratification
    provides: "06-06's protocol.StageTiming type and Lane.StageBreakdown field (schema shape only, unpopulated); 06-09's session.EvaluateBudget/Observation.StageDelta citation requirement this stage data now feeds"
provides:
  - "session.StageNames/StageRecorder/StageRecordError: explicit start/stop timestamp pairs at D-06-21's five fixed pipeline boundaries, no tracing runtime, no global registry"
  - "session.TimingObservationEnabled: the sole exported reader of LANG_OBSERVE_TIMING, consumed by both completeCommand and StageRecorder.Breakdown"
  - "verifyPhase6NativeDifferentialLane instrumented end-to-end with all five stage boundaries, proven through the shipped verify path on a real corpus"
  - "protocol.PeakRSSUnavailable: the D-06-20 rationale recorded as an enforceable doc comment, replacing ~15 hardcoded string-literal producer sites"
affects: ["06-15 (phase gate: references TestNoGetrusageAnywhere/TestPeakRSSStaysUnavailable as required controls; stage_breakdown available for its own verification pass)"]

actuals:
  tokens: 15200
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Fixed-size-array stage recorder (StageRecorder) as the entire mechanism for D-06-21's runtime-posture constraint -- no map, no span tree, no global registry, no goroutine, no package-level mutable state, proven by go/ast scan"
    - "Two genuinely separate clang invocations (compile-to-object, then link-to-executable) replacing one combined compile-and-link command, so native_compile and link are real, independently timed boundaries rather than two labels on one span"
    - "Duplicate-rather-than-modify-session.go for parse/check timing: syntax.Parse and check.Program called directly in the sibling file rather than through session.Check, which bundles both into one call"
    - "Single named constant (protocol.PeakRSSUnavailable) with the full D-06-20 rationale as its doc comment, replacing repeated string literals -- mirrors protocol.Schema1/LaneSchema1's own two-constant-coexistence precedent"

key-files:
  created:
    - internal/compiler/session/session_phase6_stages.go
    - internal/compiler/session/session_phase6_stages_test.go
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase6_verify.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_mismatch.go
    - internal/compiler/session/session_phase5_sanitize.go

key-decisions:
  - "parse and check are timed as two genuinely separate steps by calling syntax.Parse and check.Program directly in session_phase6_verify.go, duplicating session.Check's own parse-then-check shape locally, rather than wrapping one combined session.Check(source) call in a single timed span -- matches this sibling file's own established 'duplicate rather than modify session.go' discipline (classifyPhase6FixtureKind, phase6CompileBinary) and gives parse/check real, independently measured boundaries instead of two identical-duration labels."
  - "native_compile and link are split into two real clang invocations (compile-to-object with -c, then link-to-executable) inside phase6CompileBinary, replacing the prior single combined compile-and-link command -- this is what makes native_compile and link genuine, separately-timed pipeline boundaries rather than two labels stamped on one measured span, and is a legitimate two-step toolchain flow, not a fabricated one."
  - "Breakdown() gates on TimingObservationEnabled() while Start/Stop always track state -- StageRecorder.Stop's error-checking behavior (unknown stage, Stop without Start) must hold regardless of the env var, so only the OUTPUT (Breakdown) is gated, not the internal bookkeeping."
  - "TestStageBreakdownAbsentWhenUnobserved interprets D-06-21's 'byte-stable across runs with the switch unset' as stability of stage_breakdown's own absence, not full-Result byte identity across two real invocations -- lane.ElapsedNS is set unconditionally in verifyPhase6NativeDifferentialLane (pre-existing behavior, unrelated to this plan) and already varies run to run, so literal full-document byte identity was never achievable independent of this plan's scope. Documented here as a load-bearing interpretive choice."
  - "protocol.PeakRSSUnavailable is a bare identifier when referenced from inside package protocol itself (protocol.New()) and a qualified protocol.PeakRSSUnavailable selector everywhere else -- TestPeakRSSStaysUnavailable's AST scan (peakRSSAssignmentValid) accepts both forms as valid, matching how Go constants are naturally referenced across package boundaries."
  - "requirements-completed left empty per the plan's own instruction: QLT-02 and FND-04 are also declared by 06-09 (already landed) and 06-15 (the phase gate, not yet run) -- the shared-ID ready-check correctly defers marking either complete until every declaring plan finishes."

requirements-completed: []

coverage:
  - id: D1
    description: "session.StageRecorder records explicit start/stop timestamp pairs at D-06-21's five fixed pipeline stages (parse, check, lower, native_compile, link) with no tracing runtime, span model, or global registry"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestStageBreakdownCoversEveryPipelineStage"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestStageRecorderRefusesUnknownStage"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestStageRecorderFileDeclaresNoMutableStateOrGoroutines"
        status: pass
    human_judgment: false
  - id: D2
    description: "Stage timing is single-gated on LANG_OBSERVE_TIMING via TimingObservationEnabled, the sole reader of that env var; a run with the switch unset emits no stage_breakdown at all, never an array of zeros"
    requirement: "FND-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestStageTimingIsGatedOnObserveTiming"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestStageBreakdownAbsentWhenUnobserved"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_stages_test.go#TestTimingEnvironmentIsReadInExactlyOnePlace"
        status: pass
    human_judgment: false
  - id: D3
    description: "stage_breakdown never enters Result.Finalize()'s identity struct, so stage attribution cannot change semantic output"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestStageBreakdownExcludedFromIdentity"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestMetricsAndLaneFieldsExcludedFromIdentity"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestIdentityFieldEnumerationIsExhaustive"
        status: pass
    human_judgment: false
  - id: D4
    description: "peak_rss_status remains unavailable for M001; the reason is recorded in a code comment at protocol.PeakRSSUnavailable and mechanically enforced -- no raw literal, no getrusage call, no PeakRSSBytes assignment anywhere in the tree"
    requirement: "QLT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestPeakRSSStaysUnavailable"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestPeakRSSUnavailabilityIsDocumented"
        status: pass
      - kind: unit
        ref: "internal/compiler/protocol/protocol_test.go#TestNoGetrusageAnywhere"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-07
status: complete
---

# Phase 6 Plan 10: Stage Attribution and the Peak-RSS Decision Summary

**Five real clang/parse/check/lower boundaries measured end-to-end on the native-differential lane via a fixed-size-array `StageRecorder` (no tracing runtime), single-gated on `LANG_OBSERVE_TIMING`, identity-excluded, plus the peak-RSS `"unavailable"` gap now recorded as an enforceable, documented decision via `protocol.PeakRSSUnavailable`.**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-09-07 (approx, per session)
- **Completed:** 2026-09-07
- **Tasks:** 3 completed, 3 commits
- **Files modified:** 9 (2 created, 7 modified)

## Accomplishments

- `session.StageNames()`/`StageRecorder`/`StageRecordError` (new `session_phase6_stages.go`): D-06-21's five fixed, sequential pipeline stages (`parse`, `check`, `lower`, `native_compile`, `link`) recorded via explicit `Start`/`Stop` timestamp pairs into a fixed-size array, indexed by stage ordinal -- no map, no span tree, no global registry, no goroutine, no package-level mutable state, proven by a `go/ast` structural test. `Stop` on an unknown stage or without a matching prior `Start` returns a typed `*StageRecordError` rather than silently recording a zero.
- `session.TimingObservationEnabled()` is now the sole exported reader of `LANG_OBSERVE_TIMING` in the entire tree (`session.go`'s `completeCommand` and `StageRecorder.Breakdown` both route through it), proven by `TestTimingEnvironmentIsReadInExactlyOnePlace`'s `go/ast` scan over every non-test file under `internal/` and `cmd/`.
- `verifyPhase6NativeDifferentialLane` (`session_phase6_verify.go`) is instrumented end-to-end with all five boundaries: `parse`/`check` are timed as two genuinely separate steps (calling `syntax.Parse`/`check.Program` directly rather than the combined `session.Check`), `lower` times `cgen.EmitNative`, and `native_compile`/`link` are two real, separately-timed clang invocations inside `phase6CompileBinary` (compile-to-object, then link-to-executable), replacing the prior single combined compile-and-link command. Proven through the shipped verify path: with `LANG_OBSERVE_TIMING=1` and a cold cache, the lane's `stage_breakdown` carries one entry per `StageNames()` name, in pipeline order.
- `protocol.PeakRSSUnavailable` (new constant in `protocol.go`) replaces every one of the ~15 hardcoded `"unavailable"` `peak_rss_status` producer-site string literals across `session.go`, `session_phase5.go`, `session_phase5_mismatch.go`, `session_phase5_sanitize.go`, and `protocol.New()`. Its doc comment carries D-06-20's full rationale (bytes-vs-kilobytes unit discrepancy, Go-runtime-RSS confound, the host-encoding trap PROJECT.md warns against) and the second-machine revisit condition. Three tests make this mechanically enforced rather than merely documented: `TestPeakRSSStaysUnavailable` (every non-test assignment uses the constant, no raw literal, no `PeakRSSBytes` assignment), `TestPeakRSSUnavailabilityIsDocumented` (the doc comment exists and names the revisit condition), and `TestNoGetrusageAnywhere` (a `go/ast` scan proving no `syscall.Getrusage`/`unix.Getrusage`/`ru_maxrss` reference exists anywhere in `internal/` or `cmd/`).

## Task Commits

Each task was committed atomically:

1. **Task 1: One lane's five stage boundaries, measured end-to-end** - `8dc2b44` (feat)
2. **Task 2: The observation gate and identity exclusion** - `e2a4359` (test)
3. **Task 3: Record the peak-RSS gap as a deliberate decision** - `a3a1666` (docs)

**Plan metadata:** commit pending (this SUMMARY + STATE/ROADMAP)

## Files Created/Modified

- `internal/compiler/session/session_phase6_stages.go` - `StageNames`, `StageRecorder`, `StageRecordError`, `Breakdown` -- the entire stage-attribution mechanism
- `internal/compiler/session/session_phase6_stages_test.go` - all nine named tests from the plan's `<artifacts_this_phase_produces>` list
- `internal/compiler/session/session.go` - `TimingObservationEnabled` (the sole `LANG_OBSERVE_TIMING` reader), `completeCommand` now calls it; 8 `PeakRSSStatus` producer sites moved to `protocol.PeakRSSUnavailable`
- `internal/compiler/session/session_phase6_verify.go` - `verifyPhase6NativeDifferentialLane` instrumented with all five stage boundaries; `phase6CompileBinary` split into two clang invocations (compile, then link) recording `native_compile`/`link`; 2 `PeakRSSStatus` sites moved to the constant
- `internal/compiler/protocol/protocol.go` - `PeakRSSUnavailable` constant with D-06-20's full rationale doc comment; `protocol.New()` moved to it
- `internal/compiler/protocol/protocol_test.go` - `TestStageBreakdownExcludedFromIdentity`, `TestPeakRSSStaysUnavailable`, `TestPeakRSSUnavailabilityIsDocumented`, `TestNoGetrusageAnywhere`
- `internal/compiler/session/session_phase5.go`, `session_phase5_mismatch.go` (2 sites), `session_phase5_sanitize.go` - remaining `PeakRSSStatus` producer sites moved to `protocol.PeakRSSUnavailable`

## Decisions Made

See `key-decisions` in frontmatter. Most consequential: splitting `phase6CompileBinary`'s single combined compile-and-link clang invocation into two genuinely separate invocations, so `native_compile` and `link` are real, independently timed pipeline boundaries this plan can measure -- rather than fabricating a boundary inside one subprocess call that has no real internal seam.

## Deviations from Plan

None - plan executed exactly as written. All acceptance criteria across all three tasks were verified passing, including the tracer's real-corpus end-to-end proof and the three go/ast structural tests (no mutable state/goroutine, exactly-one-env-reader, no getrusage).

**Total deviations:** 0.
**Impact on plan:** None.

## Known Stubs

None. `StageRecorder`, `TimingObservationEnabled`, and `protocol.PeakRSSUnavailable` are all real, tested implementations with no placeholder data paths. The peak-RSS gap itself is an intentional, documented, mechanically enforced decision (D-06-20), not an undocumented stub -- distinguishing it from a stub is the entire point of this plan's Task 3.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `stage_breakdown` is real, gated, identity-excluded, and available for 06-15's own phase-gate verification pass and for any future D-06-22 blocking-regression stage-delta citation.
- `protocol.PeakRSSUnavailable` and its three enforcing tests are ready for 06-15 to reference as required controls.
- `QLT-02` and `FND-04` remain in `requirements-completed: []` per the plan's own instruction -- 06-15 (the phase gate) also declares both and has not yet finished; the shared-ID ready-check will mark them complete once 06-15 lands.
- `go test ./...`, `go test -race ./...`, and `go vet ./...` are all clean across the full tree at HEAD after this plan's three commits.
- No blockers for 06-15.

---
*Phase: 06-agent-feedback-and-performance-ratification*
*Completed: 2026-09-07*

## Self-Check: PASSED

- FOUND: internal/compiler/session/session_phase6_stages.go
- FOUND: internal/compiler/session/session_phase6_stages_test.go
- FOUND: commit 8dc2b44 (Task 1)
- FOUND: commit e2a4359 (Task 2)
- FOUND: commit a3a1666 (Task 3)
- All plan `<acceptance_criteria>` re-verified passing (see coverage block above)
- Plan-level `<verification>` re-run: `go build ./...`, `go test ./...`, `go test -race ./...`, `go vet ./...` all pass; all three prior-phase pins green; shipped binary emits stage_breakdown end-to-end under LANG_OBSERVE_TIMING=1
