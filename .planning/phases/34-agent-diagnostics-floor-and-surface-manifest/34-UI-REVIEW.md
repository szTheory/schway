# Phase 34 — UI Review

**Audited:** 2026-10-13  
**Status:** CLI-facing copy and command behavior reviewed; graphical UI audit not applicable  
**Baseline:** No UI-SPEC.md; Phase 34 plans, summaries, and context define CLI diagnostic and manifest behavior  
**Screenshots:** Not captured — `check ui.safety-gate 34 --raw` reported `frontend=false`, `hasUiFiles=false`, `hasUiSpec=false`, `block=false`  
**Interaction captures:** off (`workflow.ui_interaction_capture=false`); no graphical interaction surface  
**Execution evidence:** The safety gate was run for this audit. Phase summaries record their historical focused test receipts; tests and browser checks were not rerun.

## Applicability and Scores

There is no graphical frontend in this phase. Visual hierarchy, color, typography, and spacing therefore have no applicable UI evidence and receive no score. The two applicable observations below cover the user-facing CLI only; they are not a visual score or a combined score out of 24.

| Pillar | Result | Evidence |
|--------|--------|----------|
| 1. Copywriting | 3/4 — CLI only | Refusal hints give a direct alternative, but generic usage output names the command `lang` instead of `schway`. |
| 2. Visuals | N/A | Safety gate reports no frontend or UI files. |
| 3. Color | N/A | No graphical interface or color contract exists in this phase. |
| 4. Typography | N/A | No graphical interface or typography contract exists in this phase. |
| 5. Spacing | N/A | No graphical layout or spacing contract exists in this phase. |
| 6. Experience Design | 3/4 — CLI only | Diagnostic and manifest routes have concrete text/JSON behavior and bounded recovery; there is no graphical interaction surface to assess. |

## Priority Fixes

1. **Replace the stale `lang` name in generic usage output** — usage errors, including malformed `surface` invocations, currently identify the program as `lang`, which can confuse users following the Schway CLI — change the shared usage string to `schway` and pin the displayed name in the route test.

## Detailed Findings

### Pillar 1: Copywriting (3/4, CLI only)

- **WARNING — stale command name.** [`cmd/schway/main.go:720`](/Users/jon/projects/schway/cmd/schway/main.go:720) builds generic usage copy beginning `usage: lang`. The `surface` route delegates invalid arguments to this shared message at lines 134–143, so the mismatch also appears while users are trying the new route. The project publication name is Schway. Update the usage text to `schway` and assert that name in the usage-route test.
- The explicit refusal catalog gives short, form-specific hints, such as “Logical `&&` is unsupported; use explicit if/else control flow” and “An if statement requires an else block” ([`internal/compiler/surface/refusals.go:14`](/Users/jon/projects/schway/internal/compiler/surface/refusals.go:14)). These name a supported next step without promising an automatic repair.

### Pillar 2: Visuals (N/A)

No graphical screens or UI files are present according to the canonical Phase 34 safety gate. No screenshots were captured, and no visual hierarchy or icon treatment is claimed.

### Pillar 3: Color (N/A)

No graphical component or color usage is in scope. A 60/30/10 palette audit would have no subject.

### Pillar 4: Typography (N/A)

No graphical text styles or font scale are in scope. CLI text was considered under Copywriting only.

### Pillar 5: Spacing (N/A)

No graphical layout or spacing scale is in scope.

### Pillar 6: Experience Design (3/4, CLI only)

- [`cmd/schway/main.go:39`](/Users/jon/projects/schway/cmd/schway/main.go:39) routes `surface` before global flag extraction. [`runSurface` at line 134](/Users/jon/projects/schway/cmd/schway/main.go:134) accepts the two documented JSON argument orders and rejects extra arguments. The route test specifies stable standalone JSON, exactly one trailing newline, no filesystem output, and invalid-argument behavior ([`cmd/schway/main_phase34_surface_test.go:14`](/Users/jon/projects/schway/cmd/schway/main_phase34_surface_test.go:14)); this test was inspected, not rerun.
- The manifest exposes a fixed `schway.surface/1` schema for coordinates, checked types, operators, statements, builtins, call rules, restrictions, limits, and refusals ([`internal/compiler/surfacemanifest/manifest.go:21`](/Users/jon/projects/schway/internal/compiler/surfacemanifest/manifest.go:21)). Its documented purpose is source-free discovery; the route test checks that it runs from an empty directory without creating files.
- Human diagnostics include stable ID/code, source label, line and column, a bounded excerpt, and a caret. The paired JSON test checks matching locations and ensures source bytes are not included ([`internal/compiler/protocol/presentation.go:146`](/Users/jon/projects/schway/internal/compiler/protocol/presentation.go:146), [`cmd/schway/main_phase34_diagnostics_test.go:12`](/Users/jon/projects/schway/cmd/schway/main_phase34_diagnostics_test.go:12)).
- These are source and test-contract observations. This audit did not exercise a live CLI, browser, keyboard navigation, or GUI state.

## Files Audited

- Project context: `AGENTS.md`, `.planning/PROJECT.md`, and `34-CONTEXT.md`.
- Phase plans and execution summaries: `34-01-PLAN.md`, `34-01-SUMMARY.md`, `34-02-PLAN.md`, `34-02-SUMMARY.md`, `34-03-PLAN.md`, `34-03-SUMMARY.md`, `34-04-PLAN.md`, and `34-04-SUMMARY.md`.
- CLI and presentation: `cmd/schway/main.go`, `cmd/schway/main_phase34_surface_test.go`, `cmd/schway/main_phase34_diagnostics_test.go`, `internal/compiler/protocol/presentation.go`, `internal/compiler/surface/refusals.go`, and `internal/compiler/surfacemanifest/manifest.go`.

**Recommendation count:** 1 concrete CLI fix. Graphical recommendations: N/A.

## Parent Follow-up — 2026-10-13

The stale command-name warning is resolved. Generic usage now begins `usage: schway`; malformed JSON and text `surface` calls explicitly pin that public name. The inherited Phase 32 usage parser follows the corrected prefix and retains its route and omission controls.

Fresh execution: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase34SurfaceRoute$|^TestPhase34SurfaceRouteInventory$|^TestPhase32PublicRouteAudit$' -count=1 -v` exited 0; all three named tests passed (package time 3.386s). These checks support the CLI correction only. The original source-review scores and graphical N/A findings above remain historical observations, without a new visual or human-readability claim.
