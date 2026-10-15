---
phase: "34"
slug: "agent-diagnostics-floor-and-surface-manifest"
status: verified
threats_open: 0
asvs_level: 1
register_authored_at_plan_time: true
created: "2026-10-13"
---

# Phase 34 — Security

The eight plan-time threat models supply this deduplicated register. The configured audit is ASVS level 1, blocking at high severity. Existing mitigations were checked through source/test presence and the fresh 42-selector focused execution; this is a plan-register audit, not a new threat discovery pass or a hosted/native certification. All 19 authored threats have closure evidence, so the L1 short-circuit applies.

## Trust Boundaries

Captured source and loader-confirmed labels cross into diagnostic presentation; original byte spans remain edit authority. Untrusted text crosses a bounded ASCII renderer. Token recovery claims exact spans within parser budgets. Raw surface arguments are validated before flag stripping. Discovery claims cross independent checked witnesses and a bounded source census. Enriched JSON crosses the existing repair consumer. Current transition evidence stays distinct from historical receipts and future ownership.

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-34-01 | Tampering | pre-change corpus | high | mitigate | `internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden`; `internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCGoldenIsNotInert` — source inspected, final focused run passed. | closed |
| T-34-02 | Tampering | draft and name baselines | high | mitigate | `internal/compiler/syntax/parser_phase34_baseline_test.go#TestPhase34DraftBaseline`; `internal/compiler/session/session_phase34_compat_test.go#TestPhase34BoolNameBaseline` — source inspected, final focused run passed. | closed |
| T-34-03 | Tampering | diagnostic identity | high | mitigate | `internal/compiler/diagnostic/diagnostic_phase34_test.go#TestPhase34DiagnosticIdentityProjection` — source inspected, final focused run passed. | closed |
| T-34-04 | Spoofing | Human filename/message/snippet | high | mitigate | `internal/compiler/protocol/protocol_phase34_test.go#TestPhase34HumanSnippetRendering` — source inspected, final focused run passed. | closed |
| T-34-05 | Tampering | span presentation | high | mitigate | `internal/compiler/protocol/protocol_phase34_test.go#TestPhase34DiagnosticLocations` — source inspected, final focused run passed. | closed |
| T-34-06 | Information Disclosure | module labels | high | mitigate | `internal/compiler/session/session_phase34_presentation_test.go#TestPhase34ImportedDiagnosticPresentation` — source inspected, final focused run passed. | closed |
| T-34-07 | Tampering | source snapshot/provenance | high | mitigate | `internal/compiler/session/session_phase34_presentation_test.go#TestPhase34PresentationUsesCapturedSource` — source inspected, final focused run passed. | closed |
| T-34-08 | Denial of Service | delimiter recovery | high | mitigate | `internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryControls` — source inspected, final focused run passed. | closed |
| T-34-09 | Tampering | diagnostic cause completeness | high | mitigate | `internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryCorpus`; `internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryControls` — source inspected, final focused run passed. | closed |
| T-34-10 | Tampering | identifier/type admission | high | mitigate | `internal/compiler/check/check_phase34_refusal_test.go#TestPhase34BoolSpellingsRemainIdentifiers`; `internal/compiler/check/check_phase34_refusal_test.go#TestPhase34CallArityBoundaries` — source inspected, final focused run passed. | closed |
| T-34-11 | Tampering | refusal repair authority | high | mitigate | `internal/compiler/check/check_phase34_refusal_test.go#TestPhase34RefusalHintsHaveNoRepairs` — source inspected, final focused run passed. | closed |
| T-34-12 | Elevation of Privilege | surface dispatch | high | mitigate | `cmd/schway/main_phase34_surface_test.go#TestPhase34SurfaceRoute`; `cmd/schway/main_phase34_surface_test.go#TestPhase34SurfaceRouteInventory` — source inspected, final focused run passed. | closed |
| T-34-13 | Tampering | current manifest facts | high | mitigate | `internal/compiler/surfacemanifest/manifest_test.go#TestPhase34SurfaceDocument`; `internal/compiler/surfacemanifest/manifest_test.go#TestPhase34SurfaceLimits` — source inspected, final focused run passed. | closed |
| T-34-14 | Tampering | manifest admission drift | high | mitigate | `internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestWitnesses`; `internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestCensus` — source inspected, final focused run passed. | closed |
| T-34-15 | Tampering | restriction/hint drift | high | mitigate | `internal/compiler/session/session_phase34_surface_test.go#TestPhase34SurfaceManifestControls` — source inspected, final focused run passed. | closed |
| T-34-16 | Tampering | repair consumer | high | mitigate | `cmd/schway-repair/repair_phase34_test.go#TestPhase34RepairJSONCompatibility` — source inspected, final focused run passed. | closed |
| T-34-17 | Information Disclosure | public route snapshots | high | mitigate | `internal/compiler/session/session_phase34_routes_test.go#TestPhase34SourceRoutePresentation`; `cmd/schway/main_phase34_routes_test.go#TestPhase34SourceRoutePresentation`; `internal/compiler/session/session_phase34_routes_test.go#TestPhase34DiagnosticRouteInventory`; `cmd/schway/main_phase34_routes_test.go#TestPhase34DiagnosticRouteInventory` — source inspected, final focused run passed. | closed |
| T-34-18 | Tampering | compatibility and evidence claims | high | mitigate | `internal/compiler/session/session_phase34_compat_test.go#TestPhase34CompatibilityDeltaLedger`; `internal/compiler/session/session_phase34_prechange_test.go#TestPhase34PreChangeCProgramGolden`; `internal/compiler/session/session_phase34_compat_test.go#TestPhase34TransitionDocuments` — source inspected, final focused run passed. | closed |
| T-34-SC | Tampering | package installation | high | mitigate | `go.mod` has no dependencies; all new packages/reporting tooling use installed standard libraries. No installation task or package install occurred. | closed |

## Accepted Risks Log

No accepted risks. No threat was closed by assuming user acceptance or by transferring its disposition.

## Audit Scope and Limits

The final task-map run reached all 42 distinct requested names across eight packages, exit 0, wall time 4.629s. Controls covered invalid/source-less spans, hostile text, wrong-unit snapshots, recovery cause/sibling drops, manifest drift, raw flags, route classification, repair authority, exact C bytes and document overclaims, with restoration. The inherited bounded-read inventory remains closed with its negative controls after the query helper rename. These checks cover the named finite inputs and planned boundaries; arbitrary programs, human readability, whole-loop latency and final M007 hosted binding are not established. Phase 38 owns the hosted binding.

## Sign-Off

- [x] All 19 unique plan-time threats have a disposition and concrete mitigation evidence.
- [x] No accepted risk is substituted for an implementation check.
- [x] `threats_open: 0` at the configured high block threshold.
- [x] ASVS level 1 source/control depth; no L2/L3 claim.

**Audit:** Verified 2026-10-13 by the Phase 34 GSD orchestrator.

## Security Audit 2026-10-14

| Metric | Count |
|---|---|
| Threats found | 19 |
| Closed | 19 |
| Open | 0 |

## Phase 35 successor refresh — 2026-10-14

All original plans remain complete; no implementation or UAT is replayed. The
Phase 35 final inherited regression refresh passed 94 distinct names / 98
package-qualified executions across eleven packages with no missing/fail/skip,
including this phase's mapped selectors. See `../35-bitwise-operators-literal-shifts-and-hex-literals/35-REGRESSION.md`.
This refresh labels the original execution receipts historical and recomputes
the current verification fingerprint through its canonical owner.

L1 authored-register review retains every closed mitigation: bounded native capture and root/digest shell gates are byte-unchanged; source presentation and byte identity/repair authority are byte-unchanged. Shift/recovery/manifest additions have passing named independent controls and the current Phase 35 review. No new installation or accepted-risk substitution. The configured L1 zero-open short-circuit applies; this is a current source/control applicability review, not a new hosted security run.

## Phase 37 successor applicability review — 2026-10-15

The current State A review found all 30 named mitigation anchors in source,
all eight plan-time threat models and all 19 authored closed threat rows.
The fresh 22-row task-map execution passed 42 distinct names / 45 package roots,
including captured-source immutability, hostile ASCII rendering, raw surface
argument handling, recovery controls, manifest drift and byte-only repair
identity. See `34-REFRESH-RUN.json` for exact commands and source/log hashes.
The existing resolved public-name warning is preserved in UI history; the
current surface-route check pins `usage: schway` in text and JSON.

Successor inspection retains the explicit file-input route and captured-source
presentation boundary; current helper-call and bitwise admissions belong to
Phases 35/36/37, not a new Phase 34 authority grant. `go.mod` still contains
only the module identity and Go 1.24 directive. No new dependency, installation,
accepted-risk substitution or hosted security result was introduced. The L1
zero-open authored-register short-circuit applies, with no L2/L3 claim.
