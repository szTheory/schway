---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "08"
subsystem: compiler-diagnostics
tags: [go, diagnostics, source-attribution, repair, compatibility]

requires:
  - phase: 34-07
    provides: independently checked surface producer inventory and admitted-surface witnesses
provides:
  - diagnostic presentation from the already-held checked source snapshot across affected command routes
  - repair compatibility evidence proving byte edits ignore additive locations, messages, and hints
  - exact compatibility delta ledger and current transition evidence for the next M007 capabilities
affects: [phase-34-verification, phase-35, phase-36, phase-37]

actuals:
  tokens: 71752
  tasks: 3
  commits: 8
commits: 8
plan_head_before: 8e37e8072036060b6f04945ce37d548336dcf4a8
plan_head_after: ae194a952fd3406c646cdb733b45ea10bb2617a1

tech-stack:
  added: []
  patterns:
    - enrich a command result from the same source snapshot that produced its diagnostics
    - compare repair behavior with additive presentation fields mutated independently
    - keep compatibility deltas literal and finite per fixture

key-files:
  created:
    - internal/compiler/session/session_phase34_routes_test.go
    - cmd/schway/main_phase34_routes_test.go
    - cmd/schway-repair/repair_phase34_test.go
    - testdata/phase34/compat-deltas.json
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_app_verify.go
    - internal/compiler/session/session_phase6_explain.go
    - internal/compiler/session/session_phase6_query.go
    - cmd/schway/main.go
    - internal/compiler/session/session_phase34_compat_test.go
    - testdata/phase16/public-emitter-consumers.json
    - internal/compiler/session/session_phase30_inventory_test.go
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md

key-decisions:
  - "Keep BuildApplicationFile's public signature; add the command-result path for callers that need the captured diagnostic snapshot."
  - "Attach source locations only when a diagnostic span is valid against the captured source; summary identity spans remain byte-only."
  - "Treat byte spans and replacements as the repair authority; prose, hints, and presentation locations do not change selected edits."
  - "Retain z1g, trz, and z1s as dormant until their named consumer triggers are observed."

metrics:
  duration: 32min
  completed: 2026-10-13
  started: 2026-10-14T00:37:20Z
duration: 32min
completed: 2026-10-13
status: complete
requirements-completed: [DX-20, DX-21, DX-22, DX-23]

coverage:
  - id: D1
    description: "Affected diagnostic routes render locations from their held source snapshot and preserve source-less boundaries."
    requirement: DX-20
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase34_routes_test.go#TestPhase34SourceRoutePresentation"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase34_routes_test.go#TestPhase34DiagnosticRouteInventory"
        status: pass
      - kind: integration
        ref: "cmd/schway/main_phase34_routes_test.go#TestPhase34DiagnosticRouteInventory"
        status: pass
    human_judgment: false
  - id: D2
    description: "The real repair consumer applies the same eligible byte edit when additive locations and prose change, and declines cross-unit edits."
    requirement: DX-20
    verification:
      - kind: integration
        ref: "cmd/schway-repair/repair_phase34_test.go#TestPhase34RepairJSONCompatibility"
        status: pass
      - kind: integration
        ref: "cmd/schway-repair/repair_phase30_test.go#TestPhase30RepairDriverDeclinesCrossUnit"
        status: pass
    human_judgment: false
  - id: D3
    description: "Pre-change C bytes, admitted Bool-looking identifiers, identity projections, and the exact permitted diagnostic deltas remain pinned."
    requirement: DX-21
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34CompatibilityDeltaLedger"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34BoolNameBaseline"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34IdentityBaseline"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCGoldenIsNotInert"
        status: pass
    human_judgment: false
  - id: D4
    description: "The two dated transition amendments retain all capability fields, dormant-seed triggers, and distinct evidence provenance under reached mutations."
    requirement: DX-21
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase34_compat_test.go#TestPhase34TransitionDocuments"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses"
        status: pass
    human_judgment: false

---

# Phase 34 Plan 08: Source Diagnostic Presentation and Transition Evidence Summary

**Command diagnostics now reuse the original checked-source snapshot for presentation while repair authority remains byte-based.**

## Performance

- **Duration:** 32 min
- **Started:** 2026-10-14T00:37:20Z (first actual RED run)
- **Completed:** 2026-10-14T01:09:23Z
- **Tasks:** 3
- **Files modified:** 23 across the plan commits, including focused evidence records.

## Accomplishments

- Added a command-result path that preserves the original `BuildApplicationFile` API and carries source presentation through build, run, app verification, format, evidence, interface, debug-map, explain, and query routes. Explain/query identity spans remain byte-only; unavailable and global failures remain source-less.
- Verified the existing repair driver makes the same eligible byte replacement with JSON locations, messages, and hints changed, and refuses cross-unit repairs. Emitted C, live legal Bool-looking names, and identity baselines remain pinned; `compat-deltas.json` permits only literal fixture-specific recovery/refusal changes.
- Added dated transition tables that retain Phase 35 and Phase 36/37 ownership, preserve all three dormant seed thresholds, and distinguish source inspection, newly executed focused checks, historical M006 receipts, pending Phase 38 binding, and unmeasured readability/latency.

## Task Commits

1. **Task 1: Captured source presentation** — `aeecf9e7` RED, `41002f6f` implementation, `fd12dd9c` tracer verification evidence.
2. **Task 2: Diagnostic route inventory** — `b1c592cf` RED, `7805f8db` implementation.
3. **Task 3: Repair compatibility, exact deltas, and transition evidence** — `484ee724` RED, `3bdbf2c5` implementation and documents, `ae194a95` related emitter-source inventory correction.

**Plan head before:** `8e37e8072036060b6f04945ce37d548336dcf4a8`  
**Plan head after production commits:** `ae194a952fd3406c646cdb733b45ea10bb2617a1`  
**Plan production commits measured from the ledger:** 8

## TDD Gate Compliance

- Task 1 RED was classified `RED_EVIDENCE_OK`; `TestPhase34SourceRoutePresentation` failed on the intended missing captured presentation. Its GREEN selectors passed, and the tracer feedback gate reran the route verification successfully.
- Task 2 RED was classified `RED_EVIDENCE_OK`; `TestPhase34DiagnosticRouteInventory` failed on the intended missing format-route presentation while the other selected controls passed. The focused route inventory GREEN selectors passed.
- Task 3 RED was classified `RED_EVIDENCE_OK`; the repair JSON test was already covered and passed, while the compatibility ledger failed on the intentionally absent literal delta file and the transition producer failed because the dated amendments were absent. The intended implementation assertions passed afterward.
- Each RED record contains the real Go JSON output, a fresh JUnit report produced by `/tmp/schway-go-junit.py`, reporter source hash, and an inspected `semanticAssessment`. No compile/setup failure or manufactured failing test authorized GREEN. No refactor was needed.

## Focused Verification

Passing selectors included `TestPhase34SourceRoutePresentation`, both `TestPhase34DiagnosticRouteInventory` implementations, `TestPhase32PublicRouteAudit`, `TestPhase321RunInputRequiredCLI`, `TestPhase34RepairJSONCompatibility`, `TestPhase34CompatibilityDeltaLedger`, `TestPhase34TransitionDocuments`, `TestPhase34PreChangeCProgramGolden`, `TestPhase34PreChangeCGoldenIsNotInert`, `TestPhase34BoolNameBaseline`, `TestPhase34IdentityBaseline`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase16PublicEmitterConsumerInventory`, `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28PrunedEmitMatchesGolden`, `TestEmittedCExitLiteralsMatchTable`, `TestPhase32CodeDigest`, and `TestPhase32ReceiptLivingClaims`. All runs used `GOCACHE=/tmp/ai-lang-verification-gocache`; no broad suite, native matrix, push, or hosted dispatch ran.

**Outstanding parent follow-up:** The required `TestVerificationGroundednessThreeClassesAreEmpty` selector was run and failed on scoped Phase 34 evidence metadata: one stale package path at `34-RESEARCH.md:270` and three unowned exact R2b rows at lines 265, 271, and 272. The current test paths and selectors have been confirmed; the parent executor owns reconciling the pinned research/frontier records and rerunning this gate. This report does not claim that gate passed. Phase 34 remains unverified and is not marked complete here.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Refreshed a moved public-emitter source location**
- **Found during:** final focused compatibility verification
- **Issue:** Task 1's command-result additions moved the existing `EmitNative` call from line 3104 to line 3192, so the required closed inventory selector reported one missing and one stale registry entry.
- **Fix:** Updated only the literal source location in `testdata/phase16/public-emitter-consumers.json` after explicit user approval.
- **Verification:** `TestPhase16PublicEmitterConsumerInventory` passed after the correction.
- **Committed in:** `ae194a95`.

**2. [Rule 1 - Bug] Made transition mutation controls identify blank table cells precisely**
- **Found during:** Task 3
- **Issue:** The new table parser treated any blank first cell as a Markdown separator row, and blank owner/status mutations could produce duplicate findings.
- **Fix:** Recognized a separator only when every cell contains separator characters and suppressed secondary mismatch findings for a missing cell; owner/status mutations still produce their specific findings.
- **Files modified:** `internal/compiler/session/session_phase34_compat_test.go`.
- **Verification:** `TestPhase34TransitionDocuments` passed all clean-document and reached/restored mutation controls for both documents.
- **Committed in:** `3bdbf2c5`.

**Total deviations:** 2 auto-fixed issues (Rule 1).  
**Impact on plan:** Both fixes preserve the plan's focused evidence and exact inventory/mutation contracts.

## Decisions Made

- Keep presentation additive and derive source locations from the captured bytes already used for the command result.
- Preserve byte-only explanation/query identity spans and all existing output stream, exit, and report failure behavior.
- Keep the exact transition capability order and retain z1g, trz, and z1s as dormant under their named consumer triggers.

## Issues Encountered

The final evidence-groundedness selector remains a parent-owned planning-metadata follow-up as stated above. All source-route, repair, compatibility, transition-document, and inherited C/identity selectors listed as passing completed successfully.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The dated amendments record Phase 35's literal bitwise/shift work and Phase 36/37's imported scalar-helper consumer as the next owned capabilities. Phase 38 remains responsible for hosted binding. Phase 34 still needs the parent-owned groundedness metadata follow-up and its rerun before phase verification.

## Self-Check: PASSED

All listed created files exist, all eight measured plan commits are ancestors of the current plan head, and `git diff --check` reported no whitespace errors.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*

## Parent follow-up: metadata and retained-header closure

The outstanding research/frontier failure recorded above was resolved after terminal handoff. The current research map now points the manifest witness to its actual session package and writes three equivalent selectors as separately anchored alternatives understood by the legacy branch scanner. No archived finding, classifier, frontier pin, owner map, command count, or selected test was removed. Both groundedness gates passed: R1=0, R2=0, R3=0, R2b=45 with every member owned, over 1122 enforced-tier documents and 968 commands.

The seven named research-map tests were rerun across protocol, CLI and session. That first rerun found two stale rendering expectations from before Plan 34-03 deliberately retained the legacy byte-range prefix. The parent aligned those expectations and docs with the committed format, preserving exact bounds/escape/caret assertions. The subsequent complete seven-selector run exited 0: package times 0.008s (protocol), 0.112s (CLI) and 0.752s (session), including both literal C comparator selectors and the independent manifest witness. Original execution actuals above remain historical measurements; these are additional parent follow-ups.

The final 22-task validation audit also caught a stale Phase 30 bounded-read inventory function identity after the query wrapper was moved. Commit `0606fe1d` updates that single identity to resolveQueryDiagnosticFactsPresented; the route stays single-file and the exact new/unclassified-site controls remain armed. The inventory/snapshot/route selector passed, and the final focused task-map run passed all 42 distinct named selectors in 4.629 seconds, with no missing selector. This is another parent follow-up beyond the original eight measured task commits.

## Parent follow-up: CLI audit

The configured UI review found no graphical frontend and one concrete CLI copy warning. Commit `6da4d10a` corrects generic usage to `schway`, pins the name in invalid text/JSON surface calls, and updates the inherited route-inventory prefix. All three focused CLI selectors passed (3.386s package time). The UI report records this closure separately from the original source-review observations; no human-readability or graphical pass is claimed.
