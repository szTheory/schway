---
phase: "34"
slug: "agent-diagnostics-floor-and-surface-manifest"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-13"
---

# Phase 34 — Validation Strategy

This is a planned validation contract, not an execution receipt. New test names below are prospective and must be created before their selector is treated as evidence. The planner will align task/plan/wave IDs with its actual decomposition. Research executed only two focused Bool-name/tail CLI probes; those are recorded in `34-RESEARCH.md`.

## Test Infrastructure

| Property | Value |
|----------|-------|
| Framework | Go 1.24 standard-library `testing` |
| Config file | `go.mod` |
| Quick command after its test exists | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/diagnostic -run '^TestPhase34DiagnosticIdentityProjection$' -count=1` |
| Broader command | Existing hosted CI only; no local project or whole-package suites |
| Feedback target | Focused selectors; aim below 60 seconds per cheap check, record actual durations during execution |
| Hosted receipt owner | Phase 38; no new binding in Phase 34 |

## Sampling Rate

- After each task, run its exact focused selector and inspect exit/output; a selector matching zero tests is not a pass.
- After each wave, run the combined relevant focused selectors once, including applicable legacy controls.
- Before phase verification, run the final claim selectors and planning-integrity checks on the final source.
- Preserve the full-suite prohibition in STATE. CI receipts are historical until Phase 38 binds final M007 source.

## Planned Verification Map

| Claim | Requirement | Secure behavior / refusal | Type | Prospective selector | File exists | Status |
|-------|-------------|---------------------------|------|----------------------|-------------|--------|
| Pre-change corpus, names and emission baselines | DX-20..23 | No admitted source/C/identifier meaning changes silently | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1` | W0 new | pending |
| Byte-only identity under enrichment/unit/finalization | DX-20 | Coordinates do not change diagnostic IDs or patch authority | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/diagnostic -run '^TestPhase34DiagnosticIdentityProjection$' -count=1` | new | pending |
| UTF-8/CRLF/EOF and every present diagnostic span | DX-20 | Validate source span; synthetic locations remain unavailable | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/protocol -run '^TestPhase34DiagnosticLocations$' -count=1` | new | pending |
| Bounded escaped snippet/caret | DX-20 | No terminal controls, hidden crop, guessed cell width or unbounded output | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/protocol -run '^TestPhase34HumanSnippetRendering$' -count=1` | new | pending |
| Entry/imported source snapshot and public projection | DX-20 | No reread/path inference; correct effective unit | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase34ImportedDiagnosticPresentation$' -count=1` | new | pending |
| Public broken-draft text/JSON | DX-20 | Existing stream/exit/repair compatibility retained | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase34CheckDiagnosticPresentation$' -count=1` | new | pending |
| Exact reconstructed before/after draft counts | DX-21 | One cause once; separate causes and siblings survive | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34RecoveryCorpus$' -count=1` | W0 new corpus | pending |
| Bounded recovery and reached controls | DX-21 | Delimiter ownership, lossless source, progress and named cap retained | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34RecoveryControls$' -count=1` | new | pending |
| Seven dedicated codes and one-line hints | DX-22 | Precise context, no misleading cascades or automatic edits | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax ./internal/compiler/check -run '^TestPhase34UnsupportedSurfaceRefusals$' -count=1` | new | pending |
| Bool-like names remain identifiers | DX-22 | No accidental reserved word or widened literal/type semantics | unit | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase34BoolSpellingsRemainIdentifiers$' -count=1` | W0 name witnesses | pending |
| Current surface and finite bidirectional proof | DX-23 | Independent accepted/refused contexts; omission/false-claim/hint/constraint controls | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase34SurfaceManifestWitnesses$' -count=1` | new | pending |
| Deterministic source-free CLI and route audit | DX-23 | No ambient source/host probes; raw extra flags refused | integration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase34SurfaceRoute|TestPhase34SurfaceRouteInventory)$' -count=1` | new | pending |

The exact test ownership/package may be refined by the planner without weakening the claim or control. Add phase-unique threat references after `<threat_model>` registers exist; do not invent threat IDs or mark tests green before execution.

## Existing Compatibility Anchors

Retain focused checks for `TestDiagnosticZeroBytesUnchanged`, `TestRepairIdentityUsesKindOnly`, `TestRepairSpanShiftDoesNotChangeDiagnosticID`, `TestPhase30UnitIdentity`, `TestHumanJSONProjectionParity`, `TestPhase30UnitProjection`, `TestSkippedRegionCauseSeparatesSpiralTrio`, `TestValidProgramUnaffectedBySkippedRegionFix`, `TestPhase30SingleIOSite`, `TestPhase30ImportedFileDiagnosticNamesFile`, `TestRepairDriverDecodesNoProseFields`, `TestPhase30RepairDriverDeclinesCrossUnit`, and `TestPhase32PublicRouteAudit`. These names were found in source; no new pass is claimed here.

## Wave 0 Requirements

- [ ] Reconstructed CRC/word-counter and isolated/near-miss fixtures with explicit provenance and literal pre-change code/span/count baseline before parser edits.
- [ ] Complete emitting-corpus enumeration, explicit non-emitting cases, literal C goldens and a reached comparator mutation before compiler edits.
- [ ] Accepted `true`/`false` parameter/binding-name witnesses and exact unbound-tail baseline before refusal classification.
- [ ] Create the new focused tests before executing their selectors; no new test framework or dependency.

## Manual-Only Verifications

None expected. All required contracts have deterministic proxies. Human readability preference and whole-loop latency are unmeasured; they are not acceptance claims.

## Validation Sign-Off

- [ ] Every task has an executable focused command and observable failure signal, or a real Wave 0 dependency.
- [ ] No three consecutive tasks without automated evidence.
- [ ] All source/emission baselines precede the changes they constrain.
- [ ] Bidirectional manifest and recovery negative controls are reached and restored.
- [ ] Final selectors match real named tests; coverage references and kinds are grounded.
- [ ] Actual feedback durations recorded without inventing distributions.
- [ ] `nyquist_compliant: true` and `wave_0_complete: true` set only after execution audit.

**Approval:** Planned under the user's delegated decisions; execution evidence pending.
