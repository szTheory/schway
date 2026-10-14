# Phase 33 — UI Review

**Audited:** 2026-10-13 (successor-stale verification refresh)  
**Status:** Graphical UI review not applicable; CLI and command-output behavior reviewed by source inspection  
**Baseline:** No UI-SPEC.md; Phase 33 is CLI/toolchain hardening  
**Canonical UI gate:** `check ui.safety-gate 33 --raw` returned `frontend=false`, `hasUiFiles=false`, `hasUiSpec=false`, `block=false`  
**Screenshots:** Not captured; there is no graphical frontend to capture  
**Interaction captures:** Off (`workflow.ui_interaction_capture=false`)  
**Visual score:** N/A; no 24-point aggregate or human-readability claim is made.

## Applicability

Visuals, color, typography, and spacing are not applicable to this phase. The user-facing surfaces in scope are Schway CLI diagnostics, replay/evidence outcomes, and developer-facing freeze/receipt gate messages. This review records their observable contract from current source and tests; it does not claim a subjective usability assessment.

## Findings

No current BLOCKER or WARNING was found in the applicable Phase 33 CLI surface during this source review.

- `internal/compiler/native/native_app.go:694-700` reads evidence capture through `io.LimitReader` with a one-byte sentinel beyond the configured limit, bounding the allocation while allowing oversize detection.
- `internal/compiler/native/native_app.go:391-395` distinguishes missing/invalid requested evidence from capture-cap exhaustion with separate tool codes and direct messages. The `app run` route prefixes tool errors with `schway app run` (`cmd/schway/main.go:261-268`).
- The Phase 33 `app verify` path uses a named refusal code and renders source diagnostics when available (`cmd/schway/main.go:341-363`). Its success and failure outputs identify whether replay verification passed and where a report can be inspected (`cmd/schway/main.go:365-370`).
- Freeze and receipt gates expose stable status and reason values in their output. The focused gate tests pin representative messages, including malformed/missing receipt controls (`internal/compiler/session/session_freeze_gate_test.go:75-94`, `internal/compiler/session/session_receipt_gate_test.go:26-58`). These are developer-facing text interfaces, not graphical screens.
- The stale generic command name noted in the Phase 34 review is resolved in current source: `usageResult` now starts with `usage: schway` (`cmd/schway/main.go:720`), and the surface-route test asserts the Schway name in both text and JSON usage output (`cmd/schway/main_phase34_surface_test.go:87-101`). The parent reports those assertions pass; this auditor inspected their source but did not rerun them.

## Evidence and Scope

- The canonical UI safety gate was executed by this auditor and reported no frontend, UI files, or UI spec.
- Parent-supplied refresh evidence reports 47 exact named Go selectors across all 18 Phase 33 validation rows passed, with no missing selectors, in 26.325 seconds. This is parent-run evidence, not a test run by this auditor; the evidence file is `/tmp/schway33-refresh-run.json`.
- Current source inspection covered bounded capture ingestion, evidence outcome messages, the replay verification CLI path, current usage text and its text/JSON assertions, and the freeze/receipt gate message controls. The parent reports a clean source review of 48 files with exact degraded phase-range provenance; that broader review is not represented as this auditor's inspection.
- M006 hosted receipts remain historical/reopened. Phase 33 binds no hosted receipt; Phase 38 owns final hosted binding.
- No browser, screenshot, GUI, broad suite, or native matrix was run for this review.

## Files Audited

- Project constraints: `AGENTS.md` and `.planning/PROJECT.md`.
- Phase inputs: `33-CONTEXT.md`, `33-01-PLAN.md` through `33-05-PLAN.md`, and `33-01-SUMMARY.md` through `33-05-SUMMARY.md`.
- Current relevant source: `internal/compiler/native/native_app.go`, `internal/compiler/execution/defect_reasons.go`, `cmd/schway/main.go`, `cmd/schway/main_phase34_surface_test.go`, `internal/compiler/session/session_freeze_gate_test.go`, and `internal/compiler/session/session_receipt_gate_test.go`.

**Priority UI fixes:** N/A for graphical UI. No current CLI copy or command-flow fix was identified in the scoped source review.
