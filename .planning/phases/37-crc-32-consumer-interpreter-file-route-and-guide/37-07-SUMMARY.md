---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "07"
subsystem: testing
tags: [cli, file-run, exit-codes, surface-manifest, consumer-census]

# Dependency graph
requires:
  - phase: 37-03
    provides: "Public interpreter/native source-file run routes, input tuples, and route outcome tests."
provides:
  - "Finite usage, route, diagnostic, and owning-test facts for both source-file run forms."
  - "README exit semantics tied to existing source-file input and outcome tests, with all named exit values unchanged."
  - "Route-scoped 4096-byte PathToken and fixed interpreter FileBytes limits in the existing ten-field surface contract."
affects: [phase-37-plan-08, phase-38, surface-manifest, cli-route-audits]

# Actuals (#2632); chars/4 over the realized implementation diff, not harness tokens.
actuals:
  tokens: 13380
  tasks: 3
  commits: 6
commits: 6
plan_head_before: 5655e7f1c90b7a54ec1efbc9adf5cff36aa12e31
plan_head_after: 547e21564b5e99ab65e93cfe7ed25f73438e7be0

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Route-specific input and reader limits remain separate from general FileBytes language facts."
    - "Historical surface witnesses stay immutable; later facts compose through additive witness documents."

key-files:
  created:
    - cmd/schway/main_phase37_inventory_test.go
    - internal/compiler/session/session_phase37_surface_test.go
    - testdata/phase37/surface-witnesses.json
  modified:
    - cmd/schway/main.go
    - cmd/schway/main_phase32_routes_test.go
    - cmd/schway/main_phase34_routes_test.go
    - README.md
    - internal/compiler/session/session_phase28_exitdocs_test.go
    - internal/compiler/session/session_phase34_limit_consumers_test.go
    - internal/compiler/session/session_phase34_surface_test.go
    - internal/compiler/surfacemanifest/manifest.go
    - internal/compiler/surfacemanifest/manifest_test.go

key-decisions:
  - "Keep the 4096-byte opaque source-file PATH bound separate from app run's existing argument fact and the interpreter's fixed 65536-byte FileBytes read cap."
  - "Preserve the native binding-selected FileBytes maximum and the existing schway.surface/1 ten-field wire shape."
  - "Preserve the Phase 34 witness receipt pinned by Phase 35 and publish these facts through an additive Phase 37 witness document."
  - "Leave APP-14, APP-13, MOD-07, and DX-24 pending parent acceptance."

patterns-established:
  - "Finite route consumers name the actual interpreter/native handler roots and the source programs that consume FileBytes."
  - "New surface limits are independently witnessed and checked for wrong scope, missing consumers, and wire-field drift."

requirements-completed: []

# Coverage routes automated evidence only. Parent acceptance remains outstanding.
coverage:
  - id: D1
    description: "Both source-file run forms appear in usage and finite route, diagnostic, and owner inventories."
    verification:
      - kind: integration
        ref: "cmd/schway#TestPhase32PublicRouteAudit"
        status: pass
      - kind: integration
        ref: "cmd/schway#TestPhase34SourceRoutePresentation"
        status: pass
      - kind: unit
        ref: "cmd/schway#TestPhase34DiagnosticRouteInventory"
        status: pass
      - kind: unit
        ref: "cmd/schway#TestPhase37FileRunRouteInventoryControls"
        status: pass
    human_judgment: false
  - id: D2
    description: "The exit table retains every existing value, and README semantics distinguish file-run input outcomes from no-tail conformance and app run."
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestExitCodeDocsTable"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase37FileRunExitContract"
        status: pass
    human_judgment: false
  - id: D3
    description: "The source-free surface reports the opaque PATH bound, fixed interpreter reader cap, native binding-selected scope, and independent CRC consumers without changing its ten-field schema."
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestPhase34SurfaceLimitConsumers"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestPhase37FileRouteSurfaceFacts"
        status: pass
      - kind: unit
        ref: "internal/compiler/surfacemanifest#TestPhase34SurfaceDocument"
        status: pass
      - kind: unit
        ref: "internal/compiler/surfacemanifest#TestPhase34SurfaceLimits"
        status: pass
    human_judgment: false

duration: 14m
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 07: Public file-run facts Summary

**Finite route inventories, exit documentation, and source-free surface facts now describe both public file-run forms with their actual limits and owners.**

## Performance

- **Duration:** 14m
- **Started:** 2026-10-15T16:52:50Z
- **Completed:** 2026-10-15T17:07:13Z
- **Tasks:** 3
- **Files modified:** 12

## Accomplishments

- Added interpreter and native source-file forms to usage, module-aware route ownership, imported-source diagnostic, and source-presentation inventories. Existing route rows and parser enforcement remain present.
- Documented source refusal 2, tool/build failure 3, bad usage 64, typed file/input refusal 65, and the distinct existing app-run process classes without changing any exit constant. README anchors the silent empty-path `(65, "", "")` tuple to `TestPhase37FileRunInputBoundary`.
- Published a 4096-byte opaque PathToken limit for both file-run forms and a 65536-byte fixed interpreter FileBytes reader limit. The native binding-selected FileBytes maximum remains independent; `schway.surface/1` still has ten top-level fields.
- Extended the independent consumer census with both production handler roots and the two CRC file-consuming entries. Missing consumer, incorrect universal-cap wording, and added wire-field controls are checked.

## Task Commits

Each task was committed atomically with a test-first RED commit and a completing commit:

1. **Task 1: Register public source-file routes** — `81af589d` (RED test), `47fe0fb0` (route facts).
2. **Task 2: Extend the exit contract** — `a6382cd8` (RED test), `9d3cf8b1` (README semantics).
3. **Task 3: Publish scoped surface facts** — `1e7c791f` (RED test), `547e2156` (surface facts).

The measured task-commit count from the persisted plan ledger at summary creation is six. State and roadmap tracking is committed separately by the executor metadata commit.

## Verification

All plan selectors were run with `GOCACHE=/tmp/ai-lang-verification-gocache`; every named root reported RUN and PASS with no skips:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase32PublicRouteAudit|TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory|TestPhase37FileRunRouteInventoryControls)$' -count=1 -v -timeout=600s` — all four roots passed, including both engines' imported-source presentation subtests.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestExitCodeDocsTable|TestPhase37FileRunExitContract)$' -count=1 -v -timeout=600s` — both roots and semantic negative controls passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34SurfaceLimitConsumers|TestPhase37FileRouteSurfaceFacts)$' -count=1 -v -timeout=600s` — both roots and surface drift controls passed.
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest -run '^(TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument|TestPhase36SurfaceDocument)$' -count=1 -v -timeout=600s` — all four catalog/schema roots passed.
- Existing Phase 34–36 finite surface witness, census, control, and limit-consumer roots were also run with one exact anchored selector and the same `GOCACHE`; all thirteen passed.

The source-file runtime/outcome tuples from Plan 37-03 remain the linked executable evidence; this plan did not rerun broad package suites, native matrices, hosted receipts, or manual UAT. No human judgment is required for these documentation and deterministic-inventory deliverables. Parent acceptance still owns requirement completion.

## Files Created/Modified

- `cmd/schway/main.go` — public usage forms.
- `cmd/schway/main_phase32_routes_test.go`, `cmd/schway/main_phase34_routes_test.go`, `cmd/schway/main_phase37_inventory_test.go` — finite route ownership, module classification, diagnostic, and drift controls.
- `README.md`, `internal/compiler/session/session_phase28_exitdocs_test.go` — route semantics and exit-documentation controls.
- `internal/compiler/surfacemanifest/manifest.go`, `internal/compiler/surfacemanifest/manifest_test.go` — scoped limits and independently pinned limit catalog.
- `internal/compiler/session/session_phase34_limit_consumers_test.go`, `internal/compiler/session/session_phase34_surface_test.go`, `internal/compiler/session/session_phase37_surface_test.go` — actual consumer graph, additive witness composition, surface wire and scope controls.
- `testdata/phase37/surface-witnesses.json` — additive route and limit witnesses.

## Decisions Made

- The 4096-byte PATH argument bound is a route input fact; it does not become a FileBytes content cap.
- The 65536-byte value applies to the fixed interpreter adapter. Native FileBytes maximum selection remains binding-dependent, and no universal FileBytes cap is published.
- Phase 34 surface witness bytes remain unchanged because Phase 35 pins their digest. Phase 37 facts are additive and compose through the existing validator.
- Requirements remain pending parent acceptance; no requirement checkboxes were changed.

## Deviations from Plan

### Rule 3 — Blocking compatibility issue

**1. Preserved the immutable Phase 34 witness receipt with an additive Phase 37 witness document.**

- **Found during:** Task 3.
- **Issue:** The plan proposed editing `testdata/phase34/surface-witnesses.json`, while `phase35AssertImmutablePhase34Fixtures` requires its historical SHA-256 digest to remain unchanged.
- **Fix:** Added `testdata/phase37/surface-witnesses.json` and composed its witnesses through the existing finite surface validator. Updated the independent manifest limit catalog for the new route-scoped facts.
- **Files modified:** `testdata/phase37/surface-witnesses.json`, `internal/compiler/session/session_phase34_surface_test.go`, `internal/compiler/session/session_phase34_limit_consumers_test.go`, `internal/compiler/session/session_phase37_surface_test.go`, `internal/compiler/surfacemanifest/manifest_test.go`.
- **Verification:** Existing Phase 34–36 witness/census/control roots and the exact Plan 37-07 surface selector passed.
- **Committed in:** `547e2156`.

**Total deviations:** 1 compatibility-preserving adjustment. **Impact:** no historical witness rewrite, wire-field change, dependency, workflow, or requirement acceptance.

## Issues Encountered

The Phase 34 witness was intentionally immutable under its Phase 35 digest guard. The additive Phase 37 witness pattern resolved that constraint while keeping the old finite controls active.

## User Setup Required

None.

## Next Phase Readiness

The route facts and checks are ready for Plan 37-08’s executable guide work. The guide remains owned by Plan 37-08. Requirement acceptance and phase-level verification remain with the parent workflow.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

## Self-Check: PASSED

The summary and new artifacts exist; all six task commits are ancestors of the current plan head.
