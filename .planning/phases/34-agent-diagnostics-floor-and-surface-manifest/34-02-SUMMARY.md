---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "02"
subsystem: diagnostics
tags: [go, diagnostics, json, source-location, ascii]
requires:
  - phase: 34-01
    provides: structured diagnostic and repair identity floor
provides:
  - captured entry-source coordinates for text and JSON diagnostics
  - byte-only diagnostic identity with validated scalar line and column projection
  - bounded printable-ASCII source excerpts and documented wire coordinates
affects: [34-03, 34-04, 34-05, 34-06, 34-07, 34-08, check CLI, diagnostic clients]
actuals:
  tokens: 60788.25
  tasks: 3
  commits: 7
commits: 7
plan_head_before: dab49c596bc26155421bc827a539429e97b92525
plan_head_after: 72109da51256ce9b9ac43485d5f69608602a1336
tech-stack:
  added: []
  patterns:
    - invocation-local captured-source sidecar omitted from serialization and identity
    - source-origin provenance for real zero-width primary spans
    - explicit rendered-column boundaries for bounded escaped snippets
key-files:
  created:
    - internal/compiler/protocol/presentation.go
    - internal/compiler/diagnostic/diagnostic_phase34_test.go
    - internal/compiler/protocol/protocol_phase34_test.go
    - docs/diagnostics.md
  modified:
    - internal/compiler/diagnostic/diagnostic.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/session/session.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/lexer.go
    - cmd/schway/main_phase34_diagnostics_test.go
key-decisions:
  - "Half-open byte offsets remain authoritative for identity and edits; coordinates derive only from captured, bounds-checked source."
  - "A zero-width primary needs explicit source-origin provenance; synthetic operational zero spans stay unlocated."
  - "Text rendering uses bounded printable ASCII while JSON carries coordinates without source bodies."
patterns-established:
  - "Presentation enriches a copied diagnostic result and preserves semantic input and identity."
  - "Cropping returns explicit visible source bounds so omission markers cannot be confused with literal source text."
requirements-completed: [DX-20]
coverage:
  - id: D1
    description: Public check reports a broken entry-file diagnostic with matching file, line, and column in text and JSON.
    requirement: DX-20
    verification:
      - kind: integration
        ref: cmd/schway/main_phase34_diagnostics_test.go#TestPhase34CheckDiagnosticPresentation
        status: pass
    human_judgment: false
  - id: D2
    description: Validated scalar coordinates cover parser and lexer source origins, CRLF and EOF boundaries, nested units, and frozen diagnostic/result identities.
    requirement: DX-20
    verification:
      - kind: unit
        ref: internal/compiler/protocol/protocol_phase34_test.go#TestPhase34DiagnosticLocations
        status: pass
      - kind: unit
        ref: internal/compiler/diagnostic/diagnostic_phase34_test.go#TestPhase34DiagnosticIdentityProjection
        status: pass
    human_judgment: false
  - id: D3
    description: Human excerpts and detail fields are printable ASCII, bounded to 120 columns, escape-safe, and aligned with visible source ranges.
    requirement: DX-20
    verification:
      - kind: unit
        ref: internal/compiler/protocol/protocol_phase34_test.go#TestPhase34HumanSnippetRendering
        status: pass
    human_judgment: false
duration: 19 min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 02: Source-Located Diagnostics and Safe Presentation Summary

**Schway check now maps validated byte spans to scalar source positions and bounded escaped excerpts while preserving diagnostic and result identities.**

## Performance

- **Duration:** 19 min
- **Started:** 2026-10-13T21:46:15Z
- **Completed:** 2026-10-13T22:05:29Z
- **Tasks:** 3
- **Files modified:** 24, including focused evidence records and reports

## Accomplishments

- Attached the entry bytes already returned by checking the source tree, without another read or check; public text and JSON now share validated file, line, and column locations.
- Marked parser and lexer source origins, including empty-file EOF, while keeping synthetic zero-width diagnostics locationless and preserving legacy diagnostic and result IDs.
- Bounded human excerpts, filenames, and messages to printable ASCII; pinned escaped-column carets, crop boundaries, literal ellipses, CRLF ranges, and EOF insertions, then documented the wire contract.

## Task Commits

1. **Task 1: Entry broken-draft check from source to text and JSON** — `acb082a9` RED, `d6148eb6` GREEN.
2. **Task 2: Close location boundaries and source-presence identity** — `77ff7816` RED, `b9f14360` GREEN.
3. **Task 3: Pin bounded ASCII text and document wire coordinates** — `b47f9e7e` RED, `13607d8b` additional literal-ellipsis RED control, `72109da5` GREEN.

The GSD metadata commit records this summary with the updated state and roadmap after the self-check.

## Files Created/Modified

- `internal/compiler/protocol/presentation.go` — Captured-source projection, coordinate mapping, and bounded ASCII rendering.
- `internal/compiler/diagnostic/diagnostic.go` — Derived coordinates and ignored source-origin provenance with byte-only identity projections.
- `internal/compiler/protocol/protocol.go` — Presentation sidecar omitted from JSON and final result identity.
- `internal/compiler/session/session.go` — Reuses captured entry-source bytes for both check result paths.
- `internal/compiler/syntax/parser.go`, `internal/compiler/syntax/lexer.go` — Mark diagnostics originating at source positions.
- `cmd/schway/main_phase34_diagnostics_test.go`, `internal/compiler/diagnostic/diagnostic_phase34_test.go`, `internal/compiler/protocol/protocol_phase34_test.go` — Public route, identity, coordinate, and rendering goldens.
- `docs/diagnostics.md` — Byte authority, scalar coordinates, provenance, escaping, crop limits, and JSON omission rules.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-02-*-evidence.json` and matching JUnit reports — Actual named RED/GREEN runs and final tracer verification.

## Decisions Made

- Kept original half-open byte offsets as the only identity and edit coordinates; line and column remain derived presentation fields.
- Required explicit provenance for zero-width primary spans so source EOF and synthetic diagnostics remain distinct.
- Kept source capture invocation-local and out of diagnostic JSON, command identity, and result identity.

## TDD Gate Compliance

- **Task 1 RED:** `TestPhase34CheckDiagnosticPresentation` executed and failed at the planned assertion because JSON returned `line=0,col=0`; the evidence record classifies as `RED_EVIDENCE_OK`. **GREEN:** the final public check selector passed with exit 0 in 586 ms.
- **Task 2 RED:** the named empty-source parser subtest failed because real EOF diagnostics lacked source-origin provenance; the record classifies as `RED_EVIDENCE_OK`. **GREEN:** all eight planned selectors passed (50 named test events, exit 0, 566 ms), including literal identity bytes, nested spans, `InUnit`, and `Result.Finalize`.
- **Task 3 RED:** the named crop test showed the caret extending through the trailing omission marker; its record classifies as `RED_EVIDENCE_OK`. A second named control caught literal source `...` being mistaken for a crop marker and also classifies as `RED_EVIDENCE_OK`. **GREEN:** `TestPhase34HumanSnippetRendering` passed (7 named test events, exit 0, 565 ms), including the byte-offset mutation control.
- The installed classifier did not accept Go's native `go test -json` event stream. A copy-local Python standard-library reporter generated JUnit only from actual named Go test outcomes and retained raw events, stderr, exit codes, report freshness, and reporter source in the evidence records. No test outcome or human approval was fabricated.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Adapted actual Go test events to the classifier's supported JUnit format**
- **Found during:** Task 1 RED evidence capture.
- **Issue:** The installed TDD evidence classifier did not accept Go's native JSON event format.
- **Fix:** Used a copy-local Python standard-library adapter that preserves the real command result and raw events, rejects incomplete or unnamed outcomes, and writes a fresh JUnit report for classifier validation.
- **Files modified:** Task RED evidence records and matching JUnit reports under `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/`.
- **Verification:** All four RED records returned `RED_EVIDENCE_OK`; the final focused selectors passed.
- **Committed in:** `acb082a9`, `77ff7816`, `b47f9e7e`, and `13607d8b`.

**2. [Rule 1 - Bug] Distinguished literal ellipses from crop omission markers**
- **Found during:** Task 3.
- **Issue:** Inferring suffix omission from a snippet ending in `...` could hide the source range for literal dots.
- **Fix:** Made cropping return explicit visible source bounds, then clip carets to those bounds.
- **Files modified:** `internal/compiler/protocol/presentation.go`, `internal/compiler/protocol/protocol_phase34_test.go`.
- **Verification:** The literal-ellipsis RED control failed before the fix and passed in the final rendering selector.
- **Committed in:** `72109da5`.

**Total deviations:** 2 auto-fixed (1 Rule 3 blocking evidence adapter, 1 Rule 1 rendering bug).
**Impact on plan:** No semantic or schema expansion, production dependency, broad suite, or additional backend was introduced.

## Issues Encountered

The RED evidence path needed an adapter because the installed classifier did not support Go's event format. The adapter and actual source are preserved in the committed evidence records.

## User Setup Required

None — no external service configuration is required.

## Next Phase Readiness

Plan 34-03 can build on the captured-source sidecar, stable byte-only identities, and tested coordinate and text-presentation boundaries. Task 34-02 leaves no blocker for the next planned slice.

## Self-Check: PASSED

All listed implementation, test, documentation, and evidence files exist, and all seven task commits are ancestors of the current branch head.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*
