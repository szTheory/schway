---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "07"
subsystem: compiler-surface
tags: [Go, U64, bitwise, literal-shifts, hex-literals, manifest]

requires:
  - phase: 35-05
    provides: Public bitwise example and independent literal value fixture
  - phase: 35-06
    provides: Native C lowering and unused-Byte compatibility repair
provides:
  - Public interpreter and native runs pin the unused-Byte witness at value 264 and the alternate grouping at value 8
  - U64 application seams pin 264 newline independently of public execution documents
  - All committed value vectors run through O0, O3, O3-LTO; all shift vectors also run under UBSan
  - The source-free manifest publishes Phase 35 operators and limits with additive independent source witnesses and census controls
affects: [35-08, compiler-surface, BIT-01, BIT-02, BIT-03]

actuals:
  tokens: 23124
  tasks: 3
  commits: 3
commits: 3
plan_head_before: aff43258f4633ce5471ea04e7e7b4b181134f884
plan_head_after: f36049461dc8fc688c1b6abf259fdf402ce2fd3b

tech-stack:
  added: []
  patterns:
    - Public conformance values and application stdout are asserted through separate existing routes.
    - Literal expected answers remain independent of engine agreement.
    - Phase 35 surface evidence composes additively with immutable Phase 34 witness and census fixtures.

key-files:
  created:
    - cmd/schway/main_phase35_test.go
    - internal/compiler/session/session_phase35_surface_test.go
    - testdata/phase35/surface-witnesses.json
    - testdata/phase35/surface-census.json
  modified:
    - internal/compiler/session/session_phase35_values_test.go
    - internal/compiler/session/session_phase34_surface_test.go
    - internal/compiler/session/session_phase34_limit_consumers_test.go
    - internal/compiler/surfacemanifest/manifest.go
    - internal/compiler/surfacemanifest/manifest_test.go

key-decisions:
  - "Keep the public unused-Byte conformance witness separate from the U64 application stdout seam."
  - "Use literal expected answers for numeric correctness; execution-event controls establish causal admission only."
  - "Preserve Phase 34 witness and census bytes and compose Phase 35 evidence with an explicit empty supersession list."
  - "Keep BIT-01, BIT-02, and BIT-03 pending until shared phase summaries and verification close the readiness gate."

patterns-established:
  - "Manifest claims are checked against independent complete-source witnesses and a finite named source-branch census."
  - "Native optimization and sanitizer lanes are reached explicitly and assert every authored vector."

requirements-completed: []
coverage:
  - id: D1
    description: Public interpreter and native conformance projections pin value 264, alternate grouping pins 8, and the separate U64 application routes print 264 newline.
    requirement: BIT-01
    verification:
      - kind: integration
        ref: "cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase35_values_test.go#TestPhase35FocusedValueAnswers"
        status: pass
    human_judgment: false
  - id: D2
    description: Independent value vectors pass all native optimization lanes, with wrong-answer controls per operator and UBSan coverage for every committed shift vector.
    requirement: BIT-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase35_values_test.go#TestPhase35BitwiseOptimizationMatrix"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase35_values_test.go#TestPhase35ShiftUBSan"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase35_values_test.go#TestPhase35EngineWrongAnswerControls"
        status: pass
    human_judgment: false
  - id: D3
    description: Phase 35 source-free claims, expression-depth limit routes, and source census agree with full-source witnesses and copied drift controls while Phase 34 fixtures stay byte-identical.
    requirement: BIT-03
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase35_surface_test.go#TestPhase35SurfaceManifestWitnesses"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase35_surface_test.go#TestPhase35SurfaceManifestCensus"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase35_surface_test.go#TestPhase35SurfaceManifestControls"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase35_surface_test.go#TestPhase35SurfaceLimitConsumers"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestControls"
        status: pass
    human_judgment: false

duration: 20min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 07: Public Bitwise Evidence and Surface Contract Summary

**Both public engines report the bitwise witness as 264, and the source-free manifest now publishes the verified operators, limits, and independent source evidence.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-10-14T17:26:22Z
- **Completed:** 2026-10-14T17:46:31Z
- **Tasks:** 3
- **Files modified:** 9

## Accomplishments

- The public interpreter and native CLI routes report `outcome.value=264` for the unused-Byte witness. The alternate grouping reports 8 through both engines. Existing U64 application routes return 264 and write one `264` line; the public U64 missing-input refusal and closed operation dispatch controls still pass.
- Every row present in the committed value fixture passed interpreter, Clang O0, O3, and O3-LTO with literal expected answers. The shift subset also passed the UBSan lane; no required native lane skipped. Per-operator wrong-answer controls fail with the expected lane attribution.
- The manifest now publishes five U64 operators, precedence and grouping, literal shift and hex facts, and the 64-level expression-depth limit. Additive Phase 35 witnesses and census rows preserve the Phase 34 fixture bytes and keep source discovery finite and named.

## Task Commits

1. **Task 1: Run the exact unused-Byte witness through both public conformance engines** — `e4a17c13` (`test`)
2. **Task 2: Pin focused optimization/sanitizer answers and installed source facts** — `994c47ed` (`feat`)
3. **Task 3: Keep complete independent manifest and limit census controls current** — `f3604946` (`test`)

## Files Created/Modified

- `cmd/schway/main_phase35_test.go` — actual public command projections and alternate-association control.
- `internal/compiler/session/session_phase35_values_test.go` — application seams, optimization matrix, UBSan, and independent expected-answer controls.
- `internal/compiler/surfacemanifest/manifest.go` and `manifest_test.go` — current operator, restriction, refusal, and parser-depth facts.
- `internal/compiler/session/session_phase34_surface_test.go` and `session_phase34_limit_consumers_test.go` — additive Phase 35 composition and retained Phase 34 drift controls.
- `internal/compiler/session/session_phase35_surface_test.go`, `testdata/phase35/surface-witnesses.json`, and `surface-census.json` — independent full-source evidence, branch census, and controls.

## Decisions Made

D-07 keeps the unused Byte public conformance entry distinct from the U64 application route. D-18 assigns numeric correctness to independent literal answers and uses execution-event evidence only to establish causal admission. Phase 34 evidence remains an immutable base; Phase 35 additions explicitly record no scoped supersessions.

## Deviations from Plan

The plan handoff described 34 committed value vectors, but the read-only 35-05-owned fixture contains 33. I left the fixture unchanged and ran all 33 present rows through every required optimization lane; all 16 shift rows also ran under UBSan. No required lane was skipped.

## Issues Encountered

The Phase 34 census drift-control anchor referred to the old addition precedence value. I updated the mutation anchor to the current precedence while preserving its ability to expose an unaccounted branch. The census reader also canonicalizes the implementation's descriptive `operator.Kind` selector to the immutable Phase 34 `kind` key. The inherited tests and new exact drift controls all pass, and both Phase 34 fixture digests remain unchanged.

## User Setup Required

None — no external service or manual setup was needed.

## Next Phase Readiness

Plan 35-07 is complete, with Plan 35-08 remaining. BIT-01, BIT-02, and BIT-03 remain pending until the shared phase summaries and verification complete the readiness gate.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*

## Self-Check: PASSED
