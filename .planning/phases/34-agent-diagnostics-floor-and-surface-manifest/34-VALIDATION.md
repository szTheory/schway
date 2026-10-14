---
phase: "34"
slug: "agent-diagnostics-floor-and-surface-manifest"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-13"
---

# Phase 34 — Validation Strategy

This contract was audited against the eight committed plans, eight execution summaries and landed tests. The final focused run covered all 22 task rows and 42 distinct named selectors with exit 0, no missing selector, and 4.629 seconds total wall time. Per-package times are recorded below. Research-time Bool-name probes remain historical; this audit is new focused execution evidence.

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

## Per-Task Verification Map

Wave 1 is the mandatory pre-change evidence lane (logical Wave 0 safety setup). Compiler production edits begin only after 34-01's three baseline tasks and committed literal snapshots. All three baseline tasks were committed before the first production change; subsequent summaries preserve actual execution and intermediate failures.

| Task | Wave | Requirements | Claim / slice | Exact focused command | Threat references | Status |
|------|------|--------------|---------------|------------------------------------|-------------------|--------|
| 34-01-T1 | 1 | DX-20, DX-21, DX-22, DX-23 | Commit exhaustive literal pre-change C corpus | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert)$' -count=1 -v` | T-34-01, T-34-02 | passed — final focused run |
| 34-01-T2 | 1 | DX-20, DX-21, DX-22, DX-23 | Measure reconstructed drafts and isolated causes | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34DraftBaseline$' -count=1 -v` | T-34-01, T-34-02 | passed — final focused run |
| 34-01-T3 | 1 | DX-20, DX-21, DX-22, DX-23 | Freeze legal Bool names and identity/repair compatibility | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline)$' -count=1 -v` | T-34-01, T-34-02 | passed — final focused run |
| 34-02-T1 | 2 | DX-20 | Entry broken-draft check from source to text and JSON | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase34CheckDiagnosticPresentation$' -count=1 -v` | T-34-03, T-34-04, T-34-05 | passed — final focused run |
| 34-02-T2 | 2 | DX-20 | Close location boundaries and source-presence identity | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/diagnostic ./internal/compiler/protocol -run '^(TestPhase34DiagnosticIdentityProjection|TestPhase34DiagnosticLocations|TestDiagnosticZeroBytesUnchanged|TestRepairIdentityUsesKindOnly|TestRepairSpanShiftDoesNotChangeDiagnosticID|TestPhase30UnitIdentity|TestPhase30UnitProjection|TestHumanJSONProjectionParity)$' -count=1 -v` | T-34-03, T-34-04, T-34-05 | passed — final focused run |
| 34-02-T3 | 2 | DX-20 | Pin bounded ASCII text and document wire coordinates | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/protocol -run '^TestPhase34HumanSnippetRendering$' -count=1 -v` | T-34-03, T-34-04, T-34-05 | passed — final focused run |
| 34-03-T1 | 3 | DX-20 | Imported syntax error through public checked result | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34ImportedDiagnosticPresentation|TestPhase30SingleIOSite|TestPhase30ImportedFileDiagnosticNamesFile)$' -count=1 -v` | T-34-06, T-34-07 | passed — final focused run |
| 34-03-T2 | 3 | DX-20 | Multi-unit nested spans, peer provenance and immutable snapshots | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34ImportedDiagnosticPresentation|TestPhase34PresentationUsesCapturedSource)$' -count=1 -v` | T-34-06, T-34-07 | passed — final focused run |
| 34-04-T1 | 3 | DX-21 | One malformed expression reaches the next valid declaration | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34RecoveryCorpus$' -count=1 -v` | T-34-08, T-34-09 | passed — final focused run |
| 34-04-T2 | 3 | DX-21 | Independent causes, delimiter nesting and reached recovery controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase34RecoveryCorpus|TestPhase34RecoveryControls|TestSkippedRegionCauseSeparatesSpiralTrio|TestValidProgramUnaffectedBySkippedRegionFix)$' -count=1 -v` | T-34-08, T-34-09 | passed — final focused run |
| 34-05-T1 | 4 | DX-22 | One logical-and draft gives a precise refusal through checking | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34UnsupportedSurfaceRefusals$' -count=1 -v` | T-34-10, T-34-11 | passed — final focused run |
| 34-05-T2 | 4 | DX-22 | Remaining syntax refusals and contextual near misses | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^(TestPhase34UnsupportedSurfaceRefusals|TestPhase34RecoveryCorpus|TestPhase34RecoveryControls)$' -count=1 -v` | T-34-10, T-34-11 | passed — final focused run |
| 34-05-T3 | 4 | DX-22 | Unbound Bool-looking tails preserve ordinary identifiers and calls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^(TestPhase34UnsupportedSurfaceRefusals|TestPhase34BoolSpellingsRemainIdentifiers|TestPhase34CallArityBoundaries|TestPhase34RefusalHintsHaveNoRepairs)$' -count=1 -v` | T-34-10, T-34-11 | passed — final focused run |
| 34-06-T1 | 5 | DX-23 | Complete current manifest through strict source-free CLI | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase34SurfaceRoute$' -count=1 -v` | T-34-12, T-34-13 | passed — final focused run |
| 34-06-T2 | 5 | DX-23 | Checked contexts, independent literals and scoped resource/tool limits | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest -run '^(TestPhase34SurfaceDocument|TestPhase34SurfaceLimits)$' -count=1 -v` | T-34-12, T-34-13 | passed — final focused run |
| 34-06-T3 | 5 | DX-23 | Register source-free route without expanding module authority | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase34SurfaceRoute|TestPhase34SurfaceRouteInventory|TestPhase32PublicRouteAudit)$' -count=1 -v` | T-34-12, T-34-13 | passed — final focused run |
| 34-07-T1 | 6 | DX-23 | Independent source witness validates one manifest admission end to end | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase34SurfaceManifestWitnesses$' -count=1 -v` | T-34-14, T-34-15 | passed — final focused run |
| 34-07-T2 | 6 | DX-23 | Complete current branch census and accepted/refused matrix | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus)$' -count=1 -v` | T-34-14, T-34-15 | passed — final focused run |
| 34-07-T3 | 6 | DX-23 | Reach all manifest drift directions and restriction controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34SurfaceManifestControls|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus)$' -count=1 -v` | T-34-14, T-34-15 | passed — final focused run |
| 34-08-T1 | 7 | DX-20, DX-21, DX-22, DX-23 | Build/run/app-verify source refusals retain their invocation | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./cmd/schway -run '^TestPhase34SourceRoutePresentation$' -count=1 -v` | T-34-16, T-34-17, T-34-18 | passed — final focused run |
| 34-08-T2 | 7 | DX-20, DX-21, DX-22, DX-23 | All remaining diagnostic routes and route inventory | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./cmd/schway -run '^(TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory|TestPhase32PublicRouteAudit)$' -count=1 -v` | T-34-16, T-34-17, T-34-18 | passed — final focused run |
| 34-08-T3 | 7 | DX-20, DX-21, DX-22, DX-23 | Repair consumer, exact compatibility deltas and independently checked transition fields/dispositions/statuses with reached controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway-repair ./internal/compiler/session -run '^(TestPhase34RepairJSONCompatibility|TestPhase34CompatibilityDeltaLedger|TestPhase34TransitionDocuments|TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline|TestRepairDriverDecodesNoProseFields|TestPhase30RepairDriverDeclinesCrossUnit)$' -count=1 -v` | T-34-16, T-34-17, T-34-18 | passed — final focused run |

New test ownership matches the exact files declared in the eight PLAN files. Existing selectors mixed into these commands are compatibility anchors and were rerun in the final focused pass. Each task command has an adjacent fails_when condition in its plan: nonzero/compilation/missing-selector output is failure, and -v must show every requested named test. 34-07 controls run the clean observed source/witness matrix and mutate copies; reachability and target-specific failure are required for each control.

## Evidence identities and deliberate deltas

- testdata/phase34/prechange-corpus.json records every frozen tracked Schway input and both existing emitter outcomes, with literal C strings for success and explicit refusal/error cases.
- testdata/phase34/drafts.json identifies twenty reconstructed sources and their independent causes. draft-baseline.json stays immutable; draft-after.json records exact current recovery/refusal rows.
- identity-baseline.json freezes legacy schema/ID/source-less forms and accepted true/false name behavior; the unbound-tail before-state remains historical after its exact authorized refusal lands.
- compat-deltas.json records only per-fixture, literal before/after differences for specific D-09 recovery causes or D-13 refusals. Presentation-only IDs/codes/byte coordinates never enter that allowance.
- surface-witnesses.json and surface-census.json are independently authored finite claim/context/branch expectations, with positive complete-source checks and target-specific armed controls.
- TestPhase34TransitionDocuments in session_phase34_compat_test.go independently checks both dated Phase 34 living amendments: all fields for the three capability programs, literal owner/status expectations, all three dormant seed dispositions with source witness/consumer/trigger/action, and separate inspected/passed/historical/pending/unmeasured provenance. It fails on a missing row/field, wrong owner/status/disposition, nonexistent evidence test, future implementation claim, current hosted claim or measured readability/latency claim. Reached copies remove each critical field and alter each status/disposition; the same helper must report the exact document/row/field and named failure reason, then clean restoration passes. This selector is implemented and passed in the final focused run; it adds no new file, framework or local suite.
- Seven resolved spec-less edge criteria appear as truths in 34-02/03/06; two unclassified DX-21/DX-22 items remain flagged unresolved assumptions in 34-04/05.
- D-02 introduces no installed packages, external integration, AI framework, database migration or UI surface; no new dependency or service setup lane exists.

## Existing Compatibility Anchors

Retain focused checks for `TestDiagnosticZeroBytesUnchanged`, `TestRepairIdentityUsesKindOnly`, `TestRepairSpanShiftDoesNotChangeDiagnosticID`, `TestPhase30UnitIdentity`, `TestHumanJSONProjectionParity`, `TestPhase30UnitProjection`, `TestSkippedRegionCauseSeparatesSpiralTrio`, `TestValidProgramUnaffectedBySkippedRegionFix`, `TestPhase30SingleIOSite`, `TestPhase30ImportedFileDiagnosticNamesFile`, `TestRepairDriverDecodesNoProseFields`, `TestPhase30RepairDriverDeclinesCrossUnit`, and `TestPhase32PublicRouteAudit`. These names were found in source and all passed in the final focused run.

## Baseline Lane Requirements (34-01, Wave 1)

- [x] Reconstructed CRC/word-counter and isolated/near-miss fixtures with explicit provenance and literal pre-change code/span/count baseline before parser edits.
- [x] Complete emitting-corpus enumeration, explicit non-emitting cases, literal C goldens and a reached comparator mutation before compiler edits.
- [x] Accepted `true`/`false` parameter/binding-name witnesses and exact unbound-tail baseline before refusal classification.
- [x] Create the new focused tests before executing their selectors; no new test framework or dependency.

## Manual-Only Verifications

None expected. All required contracts have deterministic proxies. Human readability preference and whole-loop latency are unmeasured; they are not acceptance claims.

## Validation Sign-Off

- [x] Every task has an executable focused command and observable failure signal, or a real Wave 0 dependency.
- [x] No three consecutive tasks without automated evidence.
- [x] All source/emission baselines precede the changes they constrain.
- [x] Bidirectional manifest and recovery negative controls are reached and restored.
- [x] Final selectors match real named tests; coverage references and kinds are grounded.
- [x] Actual feedback durations recorded without inventing distributions.
- [x] `nyquist_compliant: true` and `wave_0_complete: true` set only after execution audit.

**Audit:** Validated 2026-10-13 under the delegated Phase 34 workflow; all four requirements have automated evidence. No manual-only gap was found.

## Final Focused Evidence

The parent combined the exact names from the 22 task commands across their eight named packages; this was a focused selection, not a whole-package or project suite. All 42 distinct selectors were observed and passed. The final run took 4.629 seconds wall time; package times were CLI 4.161s, repair 0.142s, checker 0.056s, diagnostics 0.011s, protocol 0.018s, session 1.406s, manifest 0.007s, syntax 0.022s. These are test-command measurements, not cold/warm feedback distributions or whole-loop latency.

The first closing pass found the renamed query bounded-read site still pinned to the old wrapper. Parent commit `0606fe1d` changes that one inventory identity; the purpose, closed inventory, single-file boundary and all three negative controls remain intact. The exact inventory/snapshot/route tests passed, then the full 42-selector focused run passed. Earlier rendering and manifest stale expectations were likewise corrected with their independent assertions preserved; execution SUMMARYs record the failures and follow-ups. No new dependency or test framework was installed.

Wave 1 captured 162 tracked inputs, both emitter outcomes, twenty explicitly reconstructed drafts and byte-only identities before production edits. All current recovery, refusal, manifest, repair, C-byte and transition controls ran and restored. Coverage remains finite to the named fixtures, route tables and seven-file source census. The two unclassified research edge categories remain scoped assumptions, not an automated proof of arbitrary inputs.

## Validation Audit 2026-10-14

| Metric | Count |
|---|---|
| Gaps found | 1 |
| Resolved | 1 |
| Escalated | 0 |

## Source-review Follow-up

The initial 42-selector task-map receipt above remains historical. Source review added three focused controls in `session_phase34_limit_consumers_test.go` and `parser_phase34_limit_scope_test.go`: TestPhase34SurfaceLimitConsumers, TestPhase34InterfaceLimitWitnesses, and TestPhase34LinearBindingLimitScope. All passed after the manifest-only correction in `18844d62`; reached omissions before correction are preserved in `34-REVIEW-FIX.md`. The new source-reference census is finite and branch-insensitive, paired with executable interface thresholds and inherited conditional debug-map/scalar controls. No new test runner or production seam was added.
