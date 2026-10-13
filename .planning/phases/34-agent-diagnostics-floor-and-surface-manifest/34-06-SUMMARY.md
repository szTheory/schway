---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "06"
subsystem: compiler-authoring-surface
tags: [go, cli, manifest, parser-bounds, agent-surface]

requires:
  - phase: 34-05
    provides: Seven current source refusal codes and literal hints
provides:
  - Deterministic source-free schway.surface/1 JSON through two exact CLI placements
  - Typed checked-context facts, scoped limits, and independently pinned refusal hints
  - Closed route inventory for both placements without expanding module-aware authority
affects: [34-07, 34-08, authoring-clients, CLI-route-audit]

actuals:
  tokens: 89264
  tasks: 3
  commits: 4
commits: 4
plan_head_before: 2697d93c763c927d27c425437b9d99db3d8c1bc4
plan_head_after: a9897027cd9157548a4cb18bffa059c4421c7342

tech-stack:
  added: []
  patterns:
    - The typed manifest is a leaf consumer of checked facts; checker admission never imports its registry.
    - Canonical serialization sorts copied slices and emits every required array.
    - Source-free CLI routes validate the complete raw argument vector before global flag extraction.

key-files:
  created:
    - internal/compiler/surfacemanifest/manifest.go
    - internal/compiler/surfacemanifest/manifest_test.go
    - cmd/schway/main_phase34_surface_test.go
    - docs/agent-surface.md
  modified:
    - cmd/schway/main.go
    - cmd/schway/main_phase32_routes_test.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/check/scalar.go

key-decisions:
  - "Publish only the two exact surface JSON argument placements; reject other raw flags before global stripping."
  - "Describe checked contexts separately from grammar reachability, ability derivation, and ordinary versus scalar-expression calls."
  - "Keep FileBytes size configuration-scoped and identify every numeric bound by category, scope, unit, meaning, and route."
  - "Keep the module-aware route set unchanged and classify surface as no-entry-source."

requirements-completed: [DX-23]

coverage:
  - id: D1
    description: Exact source-free route placements emit a deterministic standalone current manifest and malformed raw arguments retain usage behavior.
    requirement: DX-23
    verification:
      - kind: unit
        ref: cmd/schway/main_phase34_surface_test.go#TestPhase34SurfaceRoute
        status: pass
      - kind: unit
        ref: .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-06-task1-tracer-verify.json
        status: pass
    human_judgment: false
  - id: D2
    description: The typed document pins checked contexts, call and builtin arities, and all seven current refusal literals without future-owned claims.
    requirement: DX-23
    verification:
      - kind: unit
        ref: internal/compiler/surfacemanifest/manifest_test.go#TestPhase34SurfaceDocument
        status: pass
    human_judgment: false
  - id: D3
    description: Finite compiler and tool limits expose independently pinned values, scope, units, meanings, and applicable routes.
    requirement: DX-23
    verification:
      - kind: unit
        ref: internal/compiler/surfacemanifest/manifest_test.go#TestPhase34SurfaceLimits
        status: pass
    human_judgment: false
  - id: D4
    description: Both route placements are closed-inventory no-entry-source rows; removal, module-class drift, and raw-flag bypass controls are exercised.
    requirement: DX-23
    verification:
      - kind: unit
        ref: cmd/schway/main_phase34_surface_test.go#TestPhase34SurfaceRouteInventory
        status: pass
      - kind: unit
        ref: cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit
        status: pass
    human_judgment: false

duration: 15min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 06: Source-Free Agent Surface Manifest Summary

**Schway now serves a deterministic `schway.surface/1` document with checked contexts, scoped limits, seven current refusals, and closed source-free route controls.**

## Performance

- **Duration:** 15 minutes, from the pinned execution start through summary creation.
- **Started:** 2026-10-13T23:36:09Z.
- **Completed:** 2026-10-13T23:51:10Z.
- **Tasks:** 3.
- **Files changed:** 32 plan and evidence files.
- **Actual-size method:** 357,056 characters in the measured base-to-task diff, divided by four; the stored token estimate is 89,264.

## Accomplishments

- Added the typed `surfacemanifest` package and routed only `surface --json` and `--json surface` before global flag stripping. Successful output is standalone UTF-8 JSON with a fixed field order, sorted copied arrays, all required arrays present, and one final newline. Repeated calls work from an empty directory without creating files.
- Published current checked type, operator, statement, builtin, call, resource, restriction, and refusal facts. The document distinguishes ordinary one-argument calls from absent scalar-expression helper calls, pins `byte_at` at arity two, and keeps parse-only and ability-only facts out of admission claims.
- Exposed existing parser bounds through `syntax.CurrentAuthoringBounds()` and named the existing 65,536 scalar analysis budget without changing their consumers. Limits identify category, scope, unit, meaning, and routes; the `FileBytes` maximum remains configuration-scoped with no universal language-wide cap claim.
- Added `docs/agent-surface.md` and extended the closed route census for both surface placements as `no-entry-source`. The existing module-aware route set remains unchanged.

## Task Commits

1. `567fe317` — `test(34-06): add failing source-free surface route test` (valid RED, committed before implementation).
2. `48e7777e` — `feat(34-06): publish source-free current surface manifest`.
3. `b04a030d` — `fix(34-06): scope and pin current surface limits`.
4. `a9897027` — `test(34-06): register source-free surface route inventory`.

The measured plan count is four commits from `2697d93c763c927d27c425437b9d99db3d8c1bc4` through `a9897027cd9157548a4cb18bffa059c4421c7342`. The plan’s task output and test evidence touch 32 files. No commit deleted tracked files.

## Verification Evidence

All checks used Go 1.24, the `/tmp/ai-lang-verification-gocache` cache, focused selectors, and the standard-library `/tmp/schway-go-junit.py` reporter. The JSON records preserve actual Go JSON events, exit codes, JUnit paths, reporter source hashes, freshness, and semantic assessments.

- **Task 1 RED:** `TestPhase34SurfaceRoute`, exit 1, 1.695 s. The intended assertion saw the missing command return usage exit 64. `gsd_run check tdd-red-evidence` returned `RED_EVIDENCE_OK`; this test commit preceded the implementation commit.
- **Task 1 GREEN:** `TestPhase34SurfaceRoute`, exit 0, 0.329 s. The post-commit tracer rerun also passed, exit 0, 0.588 s.
- **Task 2 GREEN:** `TestPhase34SurfaceDocument` and `TestPhase34SurfaceLimits`, exit 0, 0.223 s. The tracer already delivered the complete manifest, so this task’s independent checks were already green; no failure was manufactured.
- **Task 3 RED:** `TestPhase34SurfaceRouteInventory`, exit 1, 0.848 s because both routes lacked inventory rows. The red-evidence checker returned `RED_EVIDENCE_OK`.
- **Task 3 GREEN:** `TestPhase34SurfaceRoute`, `TestPhase34SurfaceRouteInventory`, and `TestPhase32PublicRouteAudit` plus subtests, exit 0, 4.072 s; 17 named test events passed. Inventory removal reported `missing_route`, class mutation reported `wrong_module_class`, and the simulated strip-first mutation reported `unexpected_flag_acceptance`, then the raw route was checked again.

The exact run records and JUnit reports are in this phase directory under `34-06-task*.json` and matching `.junit.xml` files. Four temporary Task 2 test failures were test-oracle wording mismatches, not implementation failures; each raw run is retained with `semanticAssessment: invalid_test_expectation`, and the corrected final selector passed.

## TDD Gate Compliance

- Task 1’s intended behavioral RED was independently accepted before the feature commit. Its production implementation and post-commit tracer both passed the named route selector.
- Task 2’s typed facts and limits had already shipped in the tracer. The independent document and limit tests passed without a manufactured RED.
- Task 3’s route inventory RED was independently accepted before the inventory change. The final focused route and module-authority selector passed.

## Deviations from Plan

**1. [Rule 3 - Blocking compile issue] Corrected manifest helper signatures before verification.**
- **Found during:** Task 1.
- **Issue:** The first compile exposed invalid variadic helper signatures and an unused checked-route slice.
- **Fix:** Made contexts explicit slice parameters and used the checked-route list for checker limits.
- **Files modified:** `internal/compiler/surfacemanifest/manifest.go`.
- **Verification:** The focused route selector passed after correction.
- **Committed in:** `48e7777e`.

**Total deviations:** 1 auto-fixed blocking issue. The plan’s scope and public contract were preserved.

## Issues Encountered

The route test initially treated the pretty-printed JSON’s internal line breaks as extra trailing newlines. The Task 2 oracle also needed wording corrections to match the already-published configuration-scoped and Bool restrictions. Those exact attempts remain in the evidence records as invalid test expectations; final task selectors passed.

## User Setup Required

None. No external service, dependency install, Clang invocation, or human verification step was required.

## Next Phase Readiness

Plan 34-07 can independently verify the published document against source witnesses and negative controls. The manifest remains a consumer-facing artifact; checker admission stays in the existing parser and checker paths.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*

## Self-Check: PASSED

All eight declared plan files and this summary exist, and all four task commits are ancestors of the current `HEAD`.
