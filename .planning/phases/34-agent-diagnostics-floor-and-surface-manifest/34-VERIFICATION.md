---
phase: 34-agent-diagnostics-floor-and-surface-manifest
verified: 2026-10-16T21:37:48Z
status: passed
score: 38/38 must-haves verified
covered_files:
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-01-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-01-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-02-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-02-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-03-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-05-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-05-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-06-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-06-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-07-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-07-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-08-PLAN.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-08-SUMMARY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-371-REFRESH-GATES.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REFRESH-REGRESSION.json"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REFRESH-RUN.json"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW-DISPOSITION.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW-FIX.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-REVIEW.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-SECURITY.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-UAT.md"
  - ".planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-VALIDATION.md"
  - ".planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-REGRESSION.md"
  - ".planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/36-scalar-helper-calls/36-REGRESSION.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-ACCEPTANCE-EVIDENCE.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-01.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-02.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-05.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-08.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-14.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-15.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-COMPLETION-2026-10-16.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DISPOSITION-AMENDMENT-2026-10-16.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-INVENTORY-20261016.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-SURFACEMANIFEST-20261016.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-SYNTAX-20261016.jsonl"
  - "cmd/schway-repair/repair_phase34_test.go"
  - "cmd/schway/main.go"
  - "cmd/schway/main_phase34_diagnostics_test.go"
  - "cmd/schway/main_phase34_routes_test.go"
  - "cmd/schway/main_phase34_surface_test.go"
  - "cmd/schway/main_phase35_test.go"
  - "cmd/schway/main_phase36_test.go"
  - "cmd/schway/main_phase37_file_run_test.go"
  - "docs/agent-surface.md"
  - "docs/diagnostics.md"
  - "internal/compiler/ast/ast.go"
  - "internal/compiler/check/check.go"
  - "internal/compiler/check/check_phase34_refusal_test.go"
  - "internal/compiler/check/check_phase36_scalar_calls_test.go"
  - "internal/compiler/check/scalar.go"
  - "internal/compiler/diagnostic/diagnostic.go"
  - "internal/compiler/diagnostic/diagnostic_phase34_test.go"
  - "internal/compiler/protocol/presentation.go"
  - "internal/compiler/protocol/protocol.go"
  - "internal/compiler/protocol/protocol_phase34_test.go"
  - "internal/compiler/session/session.go"
  - "internal/compiler/session/session_app_verify.go"
  - "internal/compiler/session/session_modules.go"
  - "internal/compiler/session/session_phase34_compat_test.go"
  - "internal/compiler/session/session_phase34_limit_consumers_test.go"
  - "internal/compiler/session/session_phase34_prechange_test.go"
  - "internal/compiler/session/session_phase34_presentation_test.go"
  - "internal/compiler/session/session_phase34_routes_test.go"
  - "internal/compiler/session/session_phase34_surface_test.go"
  - "internal/compiler/session/session_phase35_surface_test.go"
  - "internal/compiler/session/session_phase36_surface_test.go"
  - "internal/compiler/session/session_phase37_surface_test.go"
  - "internal/compiler/session/session_presentation.go"
  - "internal/compiler/surface/refusals.go"
  - "internal/compiler/surfacemanifest/manifest.go"
  - "internal/compiler/surfacemanifest/manifest_test.go"
  - "internal/compiler/syntax/lexer.go"
  - "internal/compiler/syntax/parser.go"
  - "internal/compiler/syntax/parser_phase34_recovery_test.go"
  - "internal/compiler/syntax/parser_phase34_refusal_test.go"
  - "internal/compiler/syntax/parser_phase35_test.go"
  - "internal/compiler/syntax/parser_phase36_test.go"
  - "testdata/phase34/compat-deltas.json"
  - "testdata/phase34/draft-after.json"
  - "testdata/phase34/draft-baseline.json"
  - "testdata/phase34/drafts.json"
  - "testdata/phase34/identity-baseline.json"
  - "testdata/phase34/prechange-corpus.json"
  - "testdata/phase34/surface-census.json"
  - "testdata/phase34/surface-witnesses.json"
  - "testdata/phase35/shift-sites.json"
  - "testdata/phase35/surface-census.json"
  - "testdata/phase35/surface-witnesses.json"
  - "testdata/phase36/surface-census.json"
  - "testdata/phase36/surface-witnesses.json"
  - "testdata/phase37/new-fixture-ledger.json"
  - "testdata/phase37/surface-witnesses.json"
  - "testdata/phase371/compatibility-amendments.json"
  - "testdata/phase371/pre-emitter-census.json"
covered_digest: "v3:sha256:d1e0db6f377c637ab4ca269f1cfe661b01d1d7a18f4815fddbfaaa576e6bf39b"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 26
  total: 26
  not_honored: []
re_verification:
  previous_status: passed
  previous_score: 38/38
  gaps_closed: []
  gaps_remaining: []
  regressions: []
  scope: "Stale-report refresh after the Phase 37.1 source delta, followed by the scoped user-delegated disposition of the one CLI-copy clarity item; all Phase 34 truths remain verified against current source and receipts."
---

# Phase 34: Agent Diagnostics Floor and Surface Manifest Verification Report

**Phase Goal:** An agent drafting Schway gets one precise, locatable diagnostic per real cause and can ask the toolchain what surface is admitted, instead of guessing from cascading, offset-only or misleading errors.
**Verified:** 2026-10-16T21:37:48Z
**Status:** passed
**Re-verification:** Yes — stale-report refresh after the Phase 37.1 source delta, with a later scoped disposition of the pending copy item; the previous report passed without gaps.

The dated initial verification and successor amendments below are preserved as history. The latest 2026-10-16 refresh and disposition sections are authoritative for the current source check and terminal status.

## Goal Achievement

### Observable Truths

Statuses below cover the four roadmap success criteria plus all 34 PLAN `must_haves.truths`. Tests cited here were freshly executed as 70 distinct exact named selectors across 11 packages. The retained structured verbose-run receipt records 73 package-qualified root `run` events and 73 root `pass` events: three names were selected in two packages each, so event count exceeds distinct-name count. It records 70 expected and 70 passed names, no missing selectors, and exit 0. The full command and package results are recorded in `34-REGRESSION.md`. No prior verification existed.

| # | Truth | Status | Evidence |
|---|---|---|---|
| R1 | Text diagnostics provide `file:line:col`, one escaped source line and a caret; JSON adds 1-based coordinates beside unchanged byte offsets, including UTF-8, CRLF and final-line EOF. | ✓ VERIFIED | `TestPhase34DiagnosticLocations`, `TestPhase34HumanSnippetRendering`, `TestPhase34CheckDiagnosticPresentation`; `protocol/presentation.go:112` maps source offsets. |
| R2 | The pinned agent-draft corpus reduces one-cause cascades while retaining independent causes and siblings. | ✓ VERIFIED | `TestPhase34DraftBaseline`, `TestPhase34RecoveryCorpus`, `TestPhase34RecoveryControls`; the 20 fixtures are explicitly reconstructed and do not claim recovery of the unavailable CRC/word-counter originals. |
| R3 | The seven current absent forms receive dedicated `surface.unsupported_*` codes, exact spans and hints instead of misleading parser errors. | ✓ VERIFIED | `TestPhase34UnsupportedSurfaceRefusals`, `TestPhase34RefusalHintsHaveNoRepairs`, `TestPhase34SurfaceManifestControls`; the separately scoped dynamic-shift item is recorded under Deferred Items. |
| R4 | `schway surface --json` emits a versioned deterministic manifest and bidirectional controls detect missing or invented admission/refusal claims. | ✓ VERIFIED | `TestPhase34SurfaceRoute`, `TestPhase34SurfaceDocument`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestCensus`, `TestPhase34SurfaceManifestControls`. |
| P1.1 | Every enumerated tracked Schway input has explicit pre-change emission outcomes and literal successful C; a reached byte mutation fails comparison. | ✓ VERIFIED | `TestPhase34PreChangeCProgramGolden` and `TestPhase34PreChangeCGoldenIsNotInert`; fixture inventory and byte-mutant are compared in `session_phase34_prechange_test.go`. |
| P1.2 | Every reconstructed draft has labeled provenance and literal pre-change diagnostics, IDs, spans, causes and counts; originals are not claimed recovered. | ✓ VERIFIED | `TestPhase34DraftBaseline`; fixture provenance and immutable before rows are in `testdata/phase34/drafts.json` and `draft-baseline.json`. |
| P1.3 | Legal `true`/`false` parameter and binding names remain legal, while the exact unbound true-tail baseline is preserved. | ✓ VERIFIED | `TestPhase34BoolNameBaseline`, `TestPhase34BoolSpellingsRemainIdentifiers`, `TestPhase34UnsupportedSurfaceRefusals`; the unbound tail is now refused contextually. |
| P1.4 | Source-less legacy schemas, IDs, repairs, units and result identity are frozen before additive presentation. | ✓ VERIFIED | `TestPhase34IdentityBaseline`, `TestPhase34DiagnosticIdentityProjection`, `TestPhase34CompatibilityDeltaLedger`; literal baseline is `testdata/phase34/identity-baseline.json`. |
| P2.1 | Adjacent and zero-width spans keep their half-open byte boundaries; projection neither merges spans nor misplaces EOF/CRLF. | ✓ VERIFIED | `TestPhase34DiagnosticLocations`; primary and nested projections retain byte offsets in `protocol/presentation.go`. |
| P2.2 | Known empty-file EOF gets a 1:1 caret; unverified synthetic spans and nil cause/repair spans omit coordinates. | ✓ VERIFIED | `TestPhase34DiagnosticLocations`, `TestPhase34PresentationUsesCapturedSource`; position projection rejects unknown origins and invalid ranges. |
| P2.3 | Columns count Unicode scalar values, invalid bytes count individually, escaped snippet carets align to printable text, and byte offsets do not change. | ✓ VERIFIED | `TestPhase34DiagnosticLocations`, `TestPhase34HumanSnippetRendering`, `TestPhase34DiagnosticIdentityProjection`. |
| P2.4 | Repeated enrichment preserves diagnostic/cause/repair ordering and IDs; `InUnit` and location-only changes do not alter identity. | ✓ VERIFIED | `TestPhase34DiagnosticIdentityProjection`, `TestPhase34IdentityBaseline`; `SourcePosition` is presentation-only and excluded from identity projection. |
| P2.5 | Human source errors retain code/ID and bounded location/snippet/caret, while JSON never contains source bodies. | ✓ VERIFIED | `TestPhase34CheckDiagnosticPresentation`, `TestPhase34HumanSnippetRendering`; public CLI test checks both streams and source-body absence in JSON. |
| P3.1 | Presentation uses only the invocation's captured bytes and path labels; later file mutation cannot change output or trigger another read. | ✓ VERIFIED | `TestPhase34PresentationUsesCapturedSource`, `TestPhase34SourceRoutePresentation`; a read observer and post-capture mutation exercise this invariant. |
| P3.2 | Imported primary/cause/repair spans use `EffectiveUnit` and loader-confirmed relative labels; JSON retains dotted units and nil spans. | ✓ VERIFIED | `TestPhase34ImportedDiagnosticPresentation`, `TestPhase34DiagnosticRouteInventory`; module source maps are retained by session presentation. |
| P3.3 | Missing/global positions remain unlocated; entry labels follow CLI authority and imported labels do not leak canonical absolute paths. | ✓ VERIFIED | `TestPhase34DiagnosticRouteInventory`, `TestPhase34SourceRoutePresentation`; operational errors remain unlocated and source-bearing routes use captured labels. |
| P4.1 | One recovery cause yields one primary diagnostic while the next sibling and independent missing closer/mistake remain visible. | ✓ VERIFIED | `TestPhase34RecoveryCorpus`, `TestPhase34RecoveryControls`; controls specifically restore cascade, sibling swallowing and lost-cause defects. |
| P4.2 | Recovery terminates within existing source/token/diagnostic/type bounds and retains the overflow diagnostic when capped. | ✓ VERIFIED | `TestPhase34RecoveryCorpus`, `TestPhase34SurfaceLimitConsumers`, `TestPhase34LinearBindingLimitScope`; the finite census and parser bounds are exercised separately. |
| P4.3 | Only exact parser-claimed token spans suppress lexer errors; rejected source stays lossless and parse failures cannot reach checking/native emission. | ✓ VERIFIED | `TestPhase34UnsupportedSurfaceRefusals`, `TestPhase34RecoveryCorpus`, `TestPhase34SourceRoutePresentation`; exact-span and near-miss controls pass. |
| P4.4 | The spiral trio retains distinct discarded spans and diagnostic/result identities. | ✓ VERIFIED | `TestSkippedRegionCauseSeparatesSpiralTrio`; phase-34 recovery checks preserve independent causes. |
| P5.1 | Each of the seven current forms has its stable code, exact span and independently pinned one-line hint without misleading/cascading errors. | ✓ VERIFIED | `TestPhase34UnsupportedSurfaceRefusals` in syntax and checker packages; the seven literals are independently checked by `TestPhase34SurfaceManifestControls`. |
| P5.2 | Bound `true`/`false` names retain checked meanings; only the contextually unbound result-tail spelling is refused, with no literal/result widening. | ✓ VERIFIED | `TestPhase34BoolSpellingsRemainIdentifiers`, `TestPhase34BoolNameBaseline`, `TestPhase34CallArityBoundaries`. |
| P5.3 | Schway/foreign declaration arity, ordinary call arity and two-operand `byte_at` remain distinct; scalar helper calls stay absent. | ✓ VERIFIED | `TestPhase34CallArityBoundaries`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestControls`. |
| P5.4 | Refusal hints are prose only and create no structured repair or automatic replacement authority. | ✓ VERIFIED | `TestPhase34RefusalHintsHaveNoRepairs`, `TestPhase34RepairJSONCompatibility`; repair authority remains exact byte spans and replacements. |
| P6.1 | Source-free invocation emits a complete nonempty typed current manifest; no source is required and extra/empty arguments fail as usage. | ✓ VERIFIED | `TestPhase34SurfaceRoute` runs from an empty temporary directory, checks exact argument placements and rejects malformed raw args. |
| P6.2 | Standalone manifest JSON is deterministic UTF-8 with fixed fields/canonical arrays and one newline, and excludes user/host/timing data. | ✓ VERIFIED | `TestPhase34SurfaceRoute`, `TestPhase34SurfaceDocument`; repeated bytes are equal and output is checked for its one-newline framing. |
| P6.3 | `byte_at` arity two, ordinary arity restrictions and absent scalar helpers are separately represented; no unsupported helper claim is invented. | ✓ VERIFIED | `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestControls`; actual complete-source outcomes are compared with the typed facts. |
| P6.4 | Each published limit identifies category, scope, unit, meaning and applicable route; compiler bounds stay distinct from tool budgets. | ✓ VERIFIED | `TestPhase34SurfaceLimitConsumers`, `TestPhase34InterfaceLimitWitnesses`, `TestPhase34LinearBindingLimitScope`; census is finite and branch-insensitive by design. |
| P7.1 | Each admitted facet/context has an independent complete-source witness; refusal witnesses pin rejected status, code and hint; both comparison directions execute. | ✓ VERIFIED | `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestCensus`; source witnesses are independent of the typed manifest registry. |
| P7.2 | Removing/inventing admissions, refusal codes/hints or weakening contexts/arity/limits is caught for its named reason, with controls restored. | ✓ VERIFIED | `TestPhase34SurfaceManifestControls` reaches and restores each declared mutation. |
| P7.3 | The matrix independently pins `byte_at`/ordinary arities, absent scalar calls, seven refusals, limit values/scopes/units and context restrictions. | ✓ VERIFIED | `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestControls`, `TestPhase34SurfaceLimitConsumers`. |
| P7.4 | Coverage is expressly finite; neither source census alone nor production registry output is treated as a complete admission oracle. | ✓ VERIFIED | `TestPhase34SurfaceManifestCensus` checks its finite owned-file witness set; code/report make no universal grammar-reachability claim. |
| P8.1 | Affected diagnostic-bearing routes present captured source locations when proven and honestly omit unavailable/global locations while preserving streams/exits and route authority. | ✓ VERIFIED | `TestPhase34SourceRoutePresentation`, `TestPhase34DiagnosticRouteInventory`; public check/build/run/app-verify and other constructor families are inventoried. |
| P8.2 | The existing repair decoder ignores additive positions and prose, preserves eligible byte edits, and declines cross-unit changes. | ✓ VERIFIED | `TestPhase34RepairJSONCompatibility`, `TestRepairDriverDecodesNoProseFields`, `TestPhase30RepairDriverDeclinesCrossUnit`. |
| P8.3 | Admitted source/name witnesses and successful C remain unchanged; authorized diagnostic deltas are exact and unrelated diagnostics stay fixed. | ✓ VERIFIED | `TestPhase34PreChangeCProgramGolden`, `TestPhase34BoolSpellingsRemainIdentifiers`, `TestPhase34CompatibilityDeltaLedger`. |
| P8.4 | Final evidence names real tests/selectors; hosted receipt binding remains Phase 38 and readability/whole-loop latency are unmeasured. | ✓ VERIFIED | Fresh run of 70 distinct named selectors passed; no whole suite, native matrix, hosted dispatch or receipt binding ran. `34-REGRESSION.md` carries the exact command. |
| P8.5 | All three dormant seeds retain scoped dispositions and the eight-slice plan includes both runnable check and source-free discovery. | ✓ VERIFIED | `TestPhase34TransitionDocuments` checks the dated amendments; runnable gains are covered by `TestPhase34SourceRoutePresentation` and `TestPhase34SurfaceRoute`. |
| P8.6 | Both living-document amendments retain the three capability fields, seed dispositions and honest evidence statuses; targeted mutants fail for their defect. | ✓ VERIFIED | `TestPhase34TransitionDocuments` exercises field removal, changed status/disposition and evidence-overclaim controls. |

**Score:** 38/38 truths verified (0 present, behavior-unverified)

### Deferred Items (historical at the prior verification; resolved by Phase 35)

| # | Item | Addressed In | Evidence |
|---|---|---|---|
| 1 | The dynamic-shift-amount refusal code/hint is not in the current seven-code catalog. | Phase 35 | `ROADMAP.md` Phase 35 success criteria 2 and 5 require a named `surface.unsupported_*` code/hint registered in the manifest; `REQUIREMENTS.md` explicitly assigns this DX-22 sub-item to BIT-02/Phase 35. The current Phase 34 requirement tests and plans intentionally cover seven implemented forms. |

## Required Artifacts

The eight `verify.artifacts` queries passed every declared artifact (57/57 declarations; duplicates across plans were checked in each plan). The path groups below match each plan's `files_modified` inventory; all were present and substantive. The 70 distinct named selectors below exercise their behavior, rather than treating file existence as proof.

| Artifact group | Expected | Status | Details |
|---|---|---|---|
| 34-01 baselines | Pre-change source/emission, draft and identity snapshots | ✓ VERIFIED | `session_phase34_prechange_test.go`, `parser_phase34_baseline_test.go`, `session_phase34_compat_test.go`; `prechange-corpus.json`, `drafts.json`, `draft-baseline.json`, `identity-baseline.json`. |
| 34-02 presentation | Byte identity, positions, rendering and CLI behavior | ✓ VERIFIED | `diagnostic.go`, `diagnostic_phase34_test.go`, `protocol.go`, `presentation.go`, `protocol_phase34_test.go`, `session.go`, `parser.go`, `lexer.go`, `main_phase34_diagnostics_test.go`, `docs/diagnostics.md`. |
| 34-03 captured sources | Entry/import source snapshot and per-unit location projection | ✓ VERIFIED | `session.go`, `session_modules.go`, `session_presentation.go`, `session_phase34_presentation_test.go`. |
| 34-04 recovery | Bounded local recovery and literal after-state corpus | ✓ VERIFIED | `parser.go`, `lexer.go`, `parser_phase34_recovery_test.go`, `draft-after.json`. |
| 34-05 refusals | Contextual recognizers, AST/checker guards and refusal tests | ✓ VERIFIED | `parser.go`, `lexer.go`, `parser_phase34_refusal_test.go`, `surface/refusals.go`, `ast.go`, `check.go`, `scalar.go`, `check_phase34_refusal_test.go`, `draft-after.json`. |
| 34-06 manifest | Typed facts, CLI route, scope tests and guide | ✓ VERIFIED | `surfacemanifest/manifest.go`, `manifest_test.go`, `main.go`, `main_phase34_surface_test.go`, `main_phase32_routes_test.go`, `parser.go`, `scalar.go`, `docs/agent-surface.md`. |
| 34-07 witnesses | Independent complete-source witness and finite census controls | ✓ VERIFIED | `session_phase34_surface_test.go`, `surface-witnesses.json`, `surface-census.json`. |
| 34-08 integration | Public source routes, repair compatibility, delta ledger and living amendments | ✓ VERIFIED | `session.go`, `session_app_verify.go`, `main.go`, `session_phase34_routes_test.go`, `main_phase34_routes_test.go`, `session_phase6_explain.go`, `session_phase6_query.go`, `repair_phase34_test.go`, `session_phase34_compat_test.go`, `compat-deltas.json`, `PRODUCT-ROADMAP.md`, `LANGUAGE-MATURITY.md`. |

`verify.key-links` could not interpret plan `from` values expressed as prose descriptions as paths, so those machine results were not accepted as wiring evidence. I traced the following links manually through implementation and their focused behavioral tests:

## Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `CheckCommandFile` and command constructors | `protocol.Result.Presentation` | One checked invocation retains entry bytes and route label | ✓ WIRED | `session.go:1108`, `session.go:1542`, `session_presentation.go`; route inventory and source-route tests pass. |
| module loader sources / unit trees | presentation source map | `EffectiveUnit`, captured bytes and loader relative path | ✓ WIRED | `TestPhase34ImportedDiagnosticPresentation` and `TestPhase34PresentationUsesCapturedSource` verify unit mapping and no reread. |
| `protocol.WithPresentation` | primary/cause/repair spans | Copy-and-enrich from verified source origin | ✓ WIRED | `protocol/presentation.go:32-112`; nested coordinate and absent-origin tests pass. |
| `protocol.JSON` / `Human` | diagnostic locations and source snippet | JSON adds coordinates; text renderer adds bounded escaped excerpt | ✓ WIRED | `TestPhase34DiagnosticLocations`, `TestPhase34HumanSnippetRendering`, `TestPhase34CheckDiagnosticPresentation`. |
| `diagnostic.InUnit` | byte-only identity projection | Coordinates remain outside the legacy hash shape | ✓ WIRED | `diagnostic.go`; `TestPhase34DiagnosticIdentityProjection` and literal identity baselines pass. |
| parser refusal recognizers | `surface.LookupRefusal` and exact token suppression | Contextual form, exact span, stable code/hint | ✓ WIRED | `TestPhase34UnsupportedSurfaceRefusals` exercises exact matches and near misses. |
| contextual Bool-tail classification | checker/parser result-tail refusal | Bound names remain identifiers; only unbound tail is refused | ✓ WIRED | `TestPhase34BoolSpellingsRemainIdentifiers`, `TestPhase34BoolNameBaseline`. |
| `surface.Refusals()` | `surfacemanifest.Current()` | Typed manifest copies current codes/forms/hints and canonicalizes arrays | ✓ WIRED | `surfacemanifest/manifest.go:108`; manifest witness and mutation controls pass. |
| parser/check limits | manifest limits | Declared constants, bounded route/context lists and source-consumer census | ✓ WIRED | `TestPhase34SurfaceLimitConsumers`, `TestPhase34InterfaceLimitWitnesses`, `TestPhase34LinearBindingLimitScope`; scope is expressly finite. |
| raw CLI arguments | `runSurface` | Exact supported placements dispatched before global argument stripping | ✓ WIRED | `cmd/schway/main.go:134`; `TestPhase34SurfaceRoute` checks both placements, bad args and source-free execution. |
| manifest facts | independent source witnesses and census | Public check outcomes compare both directions with owned contexts | ✓ WIRED | `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestCensus`, `TestPhase34SurfaceManifestControls`. |
| enriched protocol JSON | `schway-repair` decode/apply | Existing byte span/replacement authority; hints and coordinates ignored | ✓ WIRED | `TestPhase34RepairJSONCompatibility`, `TestRepairDriverDecodesNoProseFields`, cross-unit decline test. |
| pre-change snapshots | after-state compatibility ledger | Literal C/admission baselines plus exact authorized diagnostic deltas | ✓ WIRED | Golden, baseline and `TestPhase34CompatibilityDeltaLedger` pass. |
| transition amendments | current source witnesses and dormant seed dispositions | Independently authored expected fields and mutation controls | ✓ WIRED | `TestPhase34TransitionDocuments` passes. |

## Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| diagnostic presentation | per-unit source bytes and module labels | Captured parser/module-loader input from the same invocation | Yes; coordinates/snippets are derived from original bytes | ✓ FLOWING |
| parser recovery | lossless source tree and diagnostics | Actual draft bytes parsed by lexer/parser; `Tree.Bytes()` compared to input | Yes; rejected AST is stopped before checker/emitter | ✓ FLOWING |
| surface manifest | typed current restrictions, bounds, routes and refusal catalog | Checked constants and `surface.Refusals()` mapped by `surfacemanifest.Current()` | Yes; generated from current implementation facts, not a placeholder/mock | ✓ FLOWING |
| independent manifest check | witness outcomes | Complete Schway source passed through public check; separate finite census | Yes; tests compare observed behavior to manifest both ways | ✓ FLOWING |

The manifest intentionally describes checked contexts, not reachability of every form through the full grammar. The seven-file reference census is finite and branch-insensitive; executable interface threshold controls and the aggregate AST binding-body guard supply distinct evidence. Neither is reported as a proof of universal source reachability.

## Behavioral Spot-Checks

| Behavior | Command / named test | Result | Status |
|---|---|---|---|
| 70 distinct focused selectors across 11 packages, including all phase-34 named checks and prior-phase compatibility controls | Exact anchored-selector command in `34-REGRESSION.md`, run with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1` | Structured receipt shows exit 0, 70 expected/passed distinct names and no missing selectors. Its verbose output contains 73 package-qualified root `run` and 73 root `pass` records; `TestPhase34SourceRoutePresentation` and `TestPhase34DiagnosticRouteInventory` each run in `cmd/schway` and `internal/compiler/session`, while `TestPhase34UnsupportedSurfaceRefusals` runs in `internal/compiler/check` and `internal/compiler/syntax`. | ✓ PASS |
| UTF-8/CRLF/EOF location mapping, snippets and JSON source exclusion | `TestPhase34DiagnosticLocations`, `TestPhase34HumanSnippetRendering`, `TestPhase34CheckDiagnosticPresentation` | All passed in the focused run | ✓ PASS |
| Snapshot immutability, imported source attribution and route presentation | `TestPhase34PresentationUsesCapturedSource`, `TestPhase34ImportedDiagnosticPresentation`, `TestPhase34SourceRoutePresentation`, `TestPhase34DiagnosticRouteInventory` | All passed in the focused run | ✓ PASS |
| One-cause recovery, independent causes, lossless source and mutation controls | `TestPhase34RecoveryCorpus`, `TestPhase34RecoveryControls`, `TestSkippedRegionCauseSeparatesSpiralTrio` | All passed in the focused run | ✓ PASS |
| Manifest command, independent witness/census, bidirectional drift controls and limits | `TestPhase34SurfaceRoute`, `TestPhase34SurfaceManifestWitnesses`, `TestPhase34SurfaceManifestCensus`, `TestPhase34SurfaceManifestControls`, `TestPhase34SurfaceLimitConsumers` | All passed in the focused run | ✓ PASS |

## Probe Execution

No probe script was declared by the eight plans/summaries, and no conventional `scripts/**/tests/probe-*.sh` file was present. Probe execution is not applicable.

## Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| DX-20 | 34-01, 34-02, 34-03, 34-08 | Locatable text and additive JSON coordinates with unchanged byte authority | ✓ SATISFIED | Location, rendering, imported-source, identity and public-route selectors pass. |
| DX-21 | 34-01, 34-04, 34-08 | Pinned draft corpus reduces one-cause cascades and retains independent causes | ✓ SATISFIED | Twenty explicitly reconstructed fixtures with immutable before rows; recovery corpus/controls pass. |
| DX-22 | 34-01, 34-05, 34-06, 34-07, 34-08 | Dedicated absent-form refusals with one-line hints | ✓ SATISFIED FOR PHASE 34 SCOPE | Seven current forms and their manifest controls pass. Dynamic-shift amount is tracked as the explicit Phase 35/BIT-02 deferred sub-item above, not claimed as already implemented. |
| DX-23 | 34-01, 34-06, 34-07, 34-08 | Deterministic source-free surface manifest and two-way drift controls | ✓ SATISFIED | Source-free CLI, canonical JSON and independent admission/refusal comparisons pass. |

All four IDs in plan `requirements` fields are defined in `REQUIREMENTS.md` and mapped to Phase 34. No additional requirement is mapped to Phase 34 without appearing in a plan. DX-22's dynamic-shift sub-item is explicitly owned by BIT-02/Phase 35 in the roadmap note.

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unreferenced `TBD`/`FIXME`/`XXX` markers, executable placeholders, empty user-visible implementations, or console-only implementations found in Phase 34 files. | — | No completion blocker. The word “placeholder” occurs only in explanatory comments about an unused IR place and a one-byte layout field; neither is user-visible stub behavior. |

The Plan 34-05 task-3 test and implementation were committed together, as documented. This is a process-sequencing deviation, not evidence that any verified phase behavior is absent; no red-first sequence is claimed. No new dependency, backend, general framework or native semantic rule was found in the phase's declared source scope.

## Human Verification Required

None. The phase's deterministic diagnostic, recovery, route, repair and manifest contracts are exercised by focused tests. Per D-24, human readability and cold/warm whole-loop latency remain explicitly unmeasured and are not claimed as completed truths; there is no graphical UI or external-service integration to inspect.

## Gaps Summary

The phase goal and its four roadmap success criteria were achieved in the initial verification. The 34 plan truths and four roadmap truths were behaviorally verified by named tests, the 57 declared artifact occurrences were present and substantive, and manual traces closed the key links that the path-based query could not parse. At that time the dynamic-shift refusal was deferred to Phase 35/BIT-02; the current refresh below records its delivery. The pre-existing CRC-32 and word-counter research scratch sources were unavailable; their representative reconstructed fixtures are explicitly labeled and never presented as recovered originals.

## Current successor refresh — 2026-10-14T23:46:50Z

This refresh rechecked the Phase 34 goal against current source and the latest scoped regression receipt after Phase 35. It did not replay plans, UAT, or the prior Phase 34 tests. The resulting score remains **38/38 truths verified**, with no present-behavior-unverified truth, no open gap, and no current human-verification item. The plans declare no prohibitions.

The current local receipt in `35-REGRESSION.md` records the final after-repair run: 94 distinct selectors, 98 package-qualified root RUN/PASS executions across 11 packages, exit 0 in 25.854 seconds, and zero missing, failed, or skipped selectors. The receipt says the final run followed the source repairs; the workspace has no source changes since that run. This is the fresh shared successor evidence. The earlier 70-selector Phase 34 receipt above remains historical and is not presented as rerun in this refresh.

Current source inspection confirms diagnostics retain byte offsets while adding 1-based source coordinates from the captured source; human presentation uses the captured line and caret, and JSON presentation preserves the source-less identity projection. Parser recovery/refusal paths remain wired to dedicated diagnostics. The current surface manifest is derived from the parser/checker refusal catalog and the `schway surface` route. Its nine entries preserve the original seven and include the authorized Phase 35 `surface.unsupported_shift_amount` and `surface.unsupported_expression_depth` additions. The prior DX-22 dynamic-shift deferral is therefore resolved; these Phase 35 catalog additions are authorized compatibility deltas, not Phase 34 regressions.

The current contract inventory remains **57/57 declared artifacts** and **22/22 key links**. The path-based key-link query could not parse prose-origin links and some JSON fixture targets, so those links were manually traced through source and current named controls; the prior report's “24 links” count was incorrect, and the plan declarations total 22. Requirements DX-20, DX-21, DX-22, and DX-23 are satisfied. The authored L1 security register has 19/19 items closed and zero open; its applicability check is current, not a new full security audit. State A coverage was reevaluated with no new gap. The review verdict is clean for 60 reviewed files, with the required degraded `no-task-commits-since` fallback disclosed; its disposition records zero open findings. The UI review was reused in automatic mode because this phase has no frontend. The completed UAT remains preserved and was not replayed.

The receipt is a scoped local run, not a full workspace suite, native matrix, hosted CI run, performance claim, or proof of hosted behavior. Existing source-less identity goldens remain unchanged. Archived M006 receipts remain historical; Phase 38 owns the final hosted binding. No new human-verification item is outstanding.

---

_Verified: 2026-10-14T23:46:50Z_
_Verifier: the agent (gsd-verifier)_

## Current successor refresh — 2026-10-15T07:22:07Z

This bounded amendment revalidated the Phase 34 goal after the Phase 36 successor made the former fingerprint stale. It preserves the dated Phase 34 verification and review/UAT dispositions as history; it did not replay old plans, prior Phase 34 tests, UAT, or phase.complete.

The Phase 36 prior-phase receipt in `.planning/phases/36-scalar-helper-calls/36-REGRESSION.md` binds source/test state to `b51043c6566f793893c8c4e143493e7ec3af5bcd`, after the last compiler/test edit. I checked the owner-results manifest and all 13 raw log hashes; every expected root was anchored as RUN/PASS, all 13 commands exited 0, with 80 occurrences across 69 selected names and 13 packages, and no FAIL/SKIP. Current Phase 34 roots include diagnostic route/source presentation, call-arity and refusal boundaries, recovery corpus/controls, and surface-manifest witnesses/census/controls. The final acceptance receipt in `36-ACCEPTANCE-EVIDENCE.md` binds a clean source tree at `979357388171339005fd6dbd65aa760810ae9384`; 45 actual roots across 12 packages and matching log hashes include Phase 36 call grammar, checker refusal/source-intent, manifest, and public helper route tests. A tracked source/test comparison through current `HEAD 66cebc9f` found no changes after the regression revision; later commits are docs-only. This is bounded exact-selector evidence, not a full suite or hosted result.

Phase 36 adds scalar-call grammar and refusal contracts to the shared compiler surface. `parser_phase36_test.go` exercises direct/nested calls, arity refusal, lossless recovery, and bounded depth; `check_phase36_scalar_calls_test.go` covers signature/type, import, refusal, and source-intent controls; the Phase 36 surface witness/census/control roots passed. The new local/imported U64 examples are covered by `TestPhase36PublicHelperConformance`; the regression receipt also ran `TestPhase321RunInputRequiredCLI`, preserving required-input behavior without adding a default.

The exact changed signature copy at `internal/compiler/check/scalar.go:52` names a U64/Bool pair, an existing U64/ByteView Unit body, or an unused Byte-to-U64 bridge. Checker predicates at lines 46–50 and signature/refusal controls show those are the admitted shapes. Phase 36 grammar tests assert refusal codes and recovery behavior, not exact message bytes; no plan requirement asks for byte-exact explanatory text. This dated source-inspection/copy disposition is not a new human copy review or human-verification gate.

The prior Phase 34 human/UAT disposition remains historical: the verification dated 2026-10-14 records completed UAT preserved and no open human item; this refresh did not replay it. The CLI-only copy review in `34-UI-REVIEW.md` is dated 2026-10-13; its follow-up fixed generic `usage: lang` to `usage: schway` and records three focused route tests passing. That review applies to its tested wording and is not relabeled as reviewing new Phase 36 text. Existing location, presentation, hint, and cascade controls remain intact; source inspection found no inaccurate admission shape.

All eight Phase 34 plan artifact queries passed (57 declared artifact occurrences); manual source traces for prose-origin and JSON-target links remain in force. DX-20 through DX-23 remain satisfied, including Phase 35's closure of the earlier dynamic-shift deferral. Nyquist remains validated; the authored L1 register retains 19/19 closed items and zero open, reused for applicability rather than a new full security audit. The UI safety gate remains inapplicable because there is no frontend. No new goal gap or human-verification item is outstanding; the score remains **38/38**.

---

_Verified: 2026-10-15T07:22:07Z_
_Verifier: the agent (gsd-verifier)_


## Current successor refresh — 2026-10-15T20:38:59Z

This refresh independently checked the Phase 34 goal against current compiler and CLI code after Phases 35–37. It does not replay the eight plans or completed UAT. The verdict remains **passed, 38/38 truths verified**: no missing artifact, unwired link, failing truth, blocker, or new human-verification item was found.

### Current truth check

| Truth | Current result | Evidence |
|---|---|---|
| Precise, locatable diagnostics | ✓ VERIFIED | Current protocol presentation still derives 1-based scalar columns and bounded snippets from captured source bytes while preserving byte offsets and source-less diagnostic identity. The focused location, snippet, route, and repair-compatibility checks passed. |
| One diagnostic per real cause | ✓ VERIFIED | The pinned draft/recovery corpus and negative controls remain in the current source-hash-bound regression receipt; the fresh recovery corpus and controls passed, including cascade, sibling, and independent-cause controls. |
| Honest admission/refusal surface | ✓ VERIFIED | Phase 34’s seven absent forms retain their dedicated refusals. The current catalog also contains Phase 35’s shift-amount and expression-depth refusals. Phase 35 bitwise operators, Phase 36 scalar helper calls, and the Phase 37 interpreter file route are now admitted and represented by current witnesses. |
| Source-free current manifest | ✓ VERIFIED | The schway surface --json path still emits the typed, deterministic manifest derived from current compiler facts. Current Phase 34 witness/census/mutation controls and Phase 35/36/37 successor surface tests pass. |

The historical truth rows above remain accurate for Phase 34’s original contract. In particular, the earlier “scalar calls stay absent” rows describe the Phase 34 boundary at that time; Phase 36 deliberately amended that boundary. The current manifest and successor witnesses now describe the admitted scalar helper shapes and retain their arity and signature restrictions. Likewise, Phase 35 closed the historical dynamic-shift deferral, added the shift-amount and expression-depth refusal codes, and admitted bitwise/hex forms. Phase 37 adds the interpreter file-input route and its published bounds/restrictions. These are authorized successor amendments, not regressions or claims that Phase 34 originally shipped those features.

### Current receipts and independent checks

I independently checked both current Phase 34 refresh receipts against their input and raw-log hashes. All 48 source/fixture hashes and 8 raw logs in 34-REFRESH-RUN.json matched; its 22 rows cover 42 distinct selectors and 45 package-qualified root executions across 8 packages (8.343 seconds). All 105 input hashes and 11 raw logs in 34-REFRESH-REGRESSION.json matched; its 43 rows cover 107 distinct selectors and 110 package-qualified root executions across 11 packages (46.825 seconds). The selected root events matched the expected RUN/PASS sets, with zero missing, failed, or skipped roots. The regression set includes all 70 original Phase 34 names, all Phase 33 validation-map matches, and the 11 successor roots. The current checkout adds planning changes after the receipt source revisions; no compiler or test source changes occurred, and the receipt-bound input hashes still match.

I also ran this bounded independent set on the current checkout with GOCACHE=/tmp/ai-lang-verification-gocache, exact -run selectors, -count=1, and a 10-second timeout. Every named test passed; no selected test was skipped:

| Focus | Named selectors | Result |
|---|---|---|
| Diagnostic location and text rendering (./internal/compiler/protocol) | TestPhase34DiagnosticLocations, TestPhase34HumanSnippetRendering | PASS (0.56 s) |
| Recovery corpus and negative controls (./internal/compiler/syntax) | TestPhase34RecoveryCorpus, TestPhase34RecoveryControls | PASS (0.58 s) |
| Phase 34 manifest witnesses, finite census, and drift controls (./internal/compiler/session) | TestPhase34SurfaceManifestWitnesses, TestPhase34SurfaceManifestCensus, TestPhase34SurfaceManifestControls | PASS (0.88 s) |
| Public diagnostic and surface routes (./cmd/schway) | TestPhase34CheckDiagnosticPresentation, TestPhase34SurfaceRoute | PASS (0.58 s) |
| Additive JSON repair compatibility (./cmd/schway-repair) | TestPhase34RepairJSONCompatibility | PASS (0.61 s) |
| Successor manifest witnesses (./internal/compiler/session) | TestPhase35SurfaceManifestWitnesses, TestPhase36SurfaceManifestCallWitnesses, TestPhase37FileRouteSurfaceFacts | PASS (0.63 s) |
| Phase 37 file-route grammar (./cmd/schway) | TestPhase37FileRunGrammar | PASS (0.46 s) |

The eight plan-level artifact queries report all 57/57 declared occurrences present and substantive. The path-based link query cannot parse some prose-origin and fixture-target declarations; I manually traced all 22/22 declared links through the implementation and named behavioral controls. The current traces cover captured-source handoff, imported-unit mapping, diagnostic JSON/text projection and identity, parser refusal/recovery wiring, current manifest derivation, CLI argument routing, independent witness comparison, and repair decoding. Level-4 data-flow remains real: presentation uses captured compiler input, recovery consumes the actual draft bytes, and manifest facts come from current typed compiler state rather than static fallback data.

DX-20, DX-21, DX-22, and DX-23 remain satisfied, with no additional Phase 34 requirement orphaned from the plans. The decision-coverage query for the Phase 34 context records 26/26 honored. The existing security register remains 19/19 authored mitigations closed, zero open, ASVS L1 only. The current explicit review scope remains 48 files with zero current critical, warning, or info findings; the three historical warnings are recorded as fixed and none remain open. The wider degraded resolver result is not counted as reviewed.

### Anti-patterns, probes, and limits

The current review-listed implementation and evidence files contain no unresolved TBD, FIXME, or XXX markers. Two uses of “placeholder” are explanatory comments about an unused internal IR place and a layout field; neither is an empty implementation or user-visible behavior. The bounded selector runs above skipped nothing. A separate optional Phase 37 native sanitizer test contains a host-capability skip, but that native test was not selected or represented as executed here. No Phase 34 probe was declared, and no conventional project probe script applies.

No local whole-package/project suite, native matrix, hosted dispatch, new UAT session, or hosted receipt binding was run. Human readability and cold/warm whole-loop latency remain unmeasured as required by the existing scope; this refresh does not claim those qualities. Phase 38 remains the owner of hosted receipt binding. At that date, no goal gap or human-verification item was outstanding; the current refresh below applies the present verifier-phase gate to the visible CLI diagnostic.

## Current execute-phase refresh — 2026-10-16T21:11:45Z

This is an initial-mode goal-backward refresh because the prior report had no `gaps:` section. The prior `passed` report became stale when its covered source inputs changed. The current code state is commit `20a5194`; `6572cd0` adds documentation only, and the inspected Phase 34 source inputs have no later code changes. The relevant source delta is a parser check for invalid UTF-8 within quoted defect reasons; the adjacent Phase 34 compatibility test now also checks the successor source inventory/ledger. The new parser behavior has its own Phase 37.1 tests, while the existing Phase 34 recovery and refusal cases were rerun in current receipts.

### Current observable truths

| # | Roadmap truth | Status | Current evidence |
|---|---|---|---|
| R1 | Text diagnostics show `file:line:col`, one source line and caret; JSON adds 1-based coordinates beside unchanged byte offsets, including UTF-8, CRLF and EOF cases. | ✓ VERIFIED | `protocol/presentation.go` derives positions from captured bytes; current receipts include `TestPhase34DiagnosticLocations`, `TestPhase34CheckDiagnosticPresentation`, and `TestPhase34SourceRoutePresentation`. The prior exact rendering goldens remain covered and presentation code is unchanged. |
| R2 | The pinned agent-draft corpus emits one diagnostic per real syntax cause, without follow-on expected-brace/declaration cascades. | ✓ VERIFIED | `testdata/phase34/drafts.json` contains 20 explicitly reconstructed fixtures, including CRC-32 and word-counter examples; current syntax receipts show `TestPhase34DraftBaseline`, `TestPhase34RecoveryCorpus`, and `TestPhase34RecoveryControls` passing. Original research scratch sources remain unavailable and are not claimed recovered. |
| R3 | The seven Phase 34 absent forms each receive a dedicated `surface.unsupported_*` code and one-line hint; prior misleading errors disappear. | ✓ VERIFIED | Current syntax/checker receipts show refusal, Bool-name, call-arity, and hint controls passing. `surface.Refusals()` carries Phase 34 refusals plus successor-owned shift/depth entries; dynamic-shift support is now covered by Phase 35. |
| R4 | `schway surface --json` emits a versioned deterministic manifest and bidirectional controls reject missing or invented claims. | ✓ VERIFIED | `cmd/schway/main.go` dispatches both documented argument orders to `runSurface`; `surfacemanifest.JSON` canonicalizes arrays before encoding. Current receipts include `TestPhase34SurfaceDocument`, `TestPhase34SurfaceLimits`, `TestPhase34SurfaceRoute`, and the independent witness/census/control tests. |

**Current score:** 38/38 truths verified (0 present, behavior-unverified). The 34 plan-specific truths remain itemized above in the preserved truth table; the source changes do not alter their frozen baselines or wiring contracts. Phase 35 closed the prior dynamic-shift deferral. No requirement is orphaned.

### Current artifact, link, and data-flow checks

The eight current `verify.artifacts` queries passed all 57 declared artifact occurrences. Each has a source, test, fixture, or public route role; none is missing or a stub. The path-based key-link query cannot parse prose-origin declarations, so I manually traced all 22 declared links through current code and behavioral controls. The existing Key Link Verification table records the traced connections: source snapshots reach diagnostic presentation, parser refusal recognizers reach the stable catalog, the catalog feeds the typed manifest, and the CLI route emits its canonical JSON.

Level-4 data flow remains live: source locations and snippets derive from bytes captured for the current invocation; recovery parses each actual reconstructed draft; and the manifest is built from current compiler bounds, route declarations, and refusal entries. The finite seven-file AST census is expressly not treated as a universal admission proof.

### Current behavioral receipts

The current bounded 37.1 completion receipt records 250 selectors, 251 root execution events, 240 distinct names, 16 groups, and zero missing, failed, or skipped selectors. It reuses 220 unchanged passing roots, reruns the five previously failing roots successfully, and passes all 26 previously unrun roots. The explicitly pressure-sensitive Phase 33 capture-read test is excluded. Across current raw receipts, Phase 34 test names account for 122 run events; none has a fail or skip event. Direct current roots include the Phase 34 diagnostic routes, coordinate mapping, repair compatibility, unsupported-form refusals, baseline/recovery controls, surface route, manifest witnesses, finite census, mutation controls, and limit consumers. Relevant logs are fingerprinted above. This is bounded selector evidence, not a full suite or a claim about native pressure, hosted behavior, or whole-loop latency.

| Behavior | Current evidence | Result |
|---|---|---|
| Captured-source location, JSON/text presentation, and public diagnostic routes | `37.1-371-REFRESH-REGRESSION-01.jsonl`, `37.1-371-REFRESH-REGRESSION-14.jsonl` | PASS |
| Exact reconstructed-draft baseline, one-cause recovery and independent-cause controls | `371-REFRESH-SYNTAX-20261016.jsonl`, `37.1-371-REFRESH-REGRESSION-15.jsonl` | PASS |
| Unsupported forms, preserved Bool identifiers, refusal hints and prior-code controls | `371-REFRESH-SYNTAX-20261016.jsonl`, `37.1-371-REFRESH-REGRESSION-05.jsonl` | PASS |
| Stable source-free manifest, limits, route, independent witnesses/census and bidirectional controls | `371-REFRESH-SURFACEMANIFEST-20261016.jsonl`, `37.1-371-REFRESH-REGRESSION-01.jsonl`, `37.1-371-REFRESH-REGRESSION-15.jsonl` | PASS |
| Repair decoder ignores additive presentation while preserving byte-edit authority | `37.1-371-REFRESH-REGRESSION-02.jsonl` | PASS |

### Requirement coverage

| Requirement | Status | Evidence |
|---|---|---|
| DX-20 | ✓ SATISFIED | Coordinate, rendering, identity, imported-source, route, and repair-compatibility tests; see current rows and prior exact goldens. |
| DX-21 | ✓ SATISFIED | Reconstructed 20-draft before/after baselines and current recovery corpus/negative controls. |
| DX-22 | ✓ SATISFIED | Seven Phase 34 unsupported forms are tested; Phase 35 owns the subsequently added shift-amount refusal, now listed by the current manifest. |
| DX-23 | ✓ SATISFIED | Versioned canonical manifest, source-free command route, current admission/refusal witnesses and finite bidirectional controls. |

No additional requirement mapped to Phase 34 in `REQUIREMENTS.md` is missing from the plans.

### Decision coverage

All trackable CONTEXT decisions are honored by shipped artifacts: **26/26**, with none unhonored. The gate is non-blocking.

### Test quality audit

| Requirement | Linked tests | Active | Skipped | Circular expected-value generation | Strongest assertion | Verdict |
|---|---|---:|---:|---:|---|---|
| DX-20 | `diagnostic_phase34_test.go`, `protocol_phase34_test.go`, `session_phase34_presentation_test.go`, `main_phase34_diagnostics_test.go`, route/repair tests | Yes | 0 | 0 | Exact coordinate/value and rendered-output assertions | PASS |
| DX-21 | `parser_phase34_baseline_test.go`, `parser_phase34_recovery_test.go`, `session_phase34_compat_test.go` | Yes | 0 | 0 | Literal before/after codes, counts, spans, causes and sibling behavior | PASS |
| DX-22 | `parser_phase34_refusal_test.go`, `check_phase34_refusal_test.go`, `manifest_test.go`, surface witness controls | Yes | 0 | 0 | Exact refusal code, hint, span, near-miss and admission assertions | PASS |
| DX-23 | `manifest_test.go`, `main_phase34_surface_test.go`, `session_phase34_surface_test.go` | Yes | 0 | 0 | Canonical bytes plus independent witnesses and armed drift controls | PASS |

**Disabled requirement tests:** 0. **Circular test patterns:** 0. **Insufficient assertions:** 0. The inspected `os.WriteFile` calls create temporary Schway inputs, malformed fixture inputs, or a fake tool executable; tests do not generate their expected values from the system under test. The before/after baselines are committed literal fixtures, and manifest witnesses are separately authored source programs.

### Anti-patterns and probes

No unreferenced `TBD`, `FIXME`, or `XXX` debt markers, empty implementations, or user-visible stubs were found in the Phase 34 implementation scope. The two existing “placeholder” mentions are explanatory comments about an unused internal IR place and a single-byte layout field. No phase-declared or conventional `scripts/**/tests/probe-*.sh` file applies.

### Human verification disposition (as of 2026-10-16T21:11:45Z)

At this refresh checkpoint the item below was pending, so the interim verifier decision was `human_needed`. The report preserves this as-of assessment; it is superseded by the scoped disposition recorded below, not by a claim that readability was independently measured.

#### 1. Inspect CLI diagnostic copy

**Test:** Run `schway check` on one unsupported-form draft, such as the reconstructed logical-`&&` case, and inspect the plain-text output.
**Expected:** The location, one-line source excerpt, caret, and short hint let a developer identify the cause and next step without a misleading cascade.
**Why human:** Current tests pin exact bytes and diagnostic counts, but cannot establish whether the copy is understandable in context.

### Gaps summary

No blocking code or artifact gaps were found. All 38 technical truths are verified, all 57 artifact declarations pass, all 22 links are manually traced, and DX-20 through DX-23 are satisfied. At this checkpoint the CLI copy-clarity item remained pending, so the interim status was `human_needed`.

---

_Verified: 2026-10-16T21:11:45Z_
_Verifier: the agent (gsd-verifier)_

## Current disposition — 2026-10-16T21:37:48Z

The user delegated the open wording choice with “what do u rec? do that”. The scoped UAT at `34-UAT.md` records the recommendation to keep the existing diagnostic wording, the user's delegated acceptance, and a complete 1/1 result with zero issues or pending items. It covers only this outstanding copy checkpoint; it is not a new full Phase 34 UAT campaign or an independent human readability measurement.

The existing CLI diagnostic identifies the exact `&&` operator and says `Logical && is unsupported; use explicit if/else control flow.` The verified source-backed output remains pinned by the technical checks. The delegated editorial choice resolves the pending checkpoint without changing source or wording. Human readability remains unmeasured; no independent readability result is claimed. The current human-verification list is empty, and the terminal status is `passed` at 38/38.

`34-UAT.md` is included in the current fingerprint. The canonical fingerprint command returned `covered_files` and digest `v3:sha256:d1e0db6f377c637ab4ca269f1cfe661b01d1d7a18f4815fddbfaaa576e6bf39b`.

---

_Verified: 2026-10-16T21:37:48Z_
_Verifier: the agent (gsd-verifier)_
