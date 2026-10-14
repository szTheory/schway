---
phase: 34
status: passed
selected_named_tests: 70
packages: 11
exit: 0
missing_selectors: 0
---

# Phase 34 — Focused Regression Gate

Fresh execution after the source-review correction passed all 70 selected named tests across 11 packages, including the original 42-selector task map, three new limit-consumer/guard checks, and relevant prior-phase controls. The canonical GSD normalizer and 600-second timeout wrapper ran this one-shot command; elapsed wall time was 23.882 seconds. This is a focused local regression receipt, not a full suite or a cold/warm whole-loop latency measurement.

Command:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway ./cmd/schway-repair ./internal/compiler/cgen ./internal/compiler/check ./internal/compiler/diagnostic ./internal/compiler/interp ./internal/compiler/native ./internal/compiler/protocol ./internal/compiler/session ./internal/compiler/surfacemanifest ./internal/compiler/syntax -run '^TestDiagnosticZeroBytesUnchanged$|^TestEmittedCExitLiteralsMatchTable$|^TestFreezeGate$|^TestFreezeGateDifferential$|^TestHumanJSONProjectionParity$|^TestNoCheckPhase32Caller$|^TestPhase16PublicEmitterConsumerInventory$|^TestPhase20ValidationLifecycle$|^TestPhase25LivingRoadmap$|^TestPhase26AnalysisExhaustion$|^TestPhase272AcceptedCorpusByteIdentity$|^TestPhase273ConformanceCorpusByteIdentity$|^TestPhase28PrunedEmitMatchesGolden$|^TestPhase30DebugMapNoImportUnguarded$|^TestPhase30ImportedFileDiagnosticNamesFile$|^TestPhase30ModuleWorkLimit$|^TestPhase30RepairDriverDeclinesCrossUnit$|^TestPhase30SingleIOSite$|^TestPhase30UnitIdentity$|^TestPhase30UnitProjection$|^TestPhase321RunInputRequiredCLI$|^TestPhase32PublicRouteAudit$|^TestPhase32ReceiptLivingClaims$|^TestPhase33AbortLikeNarrowed$|^TestPhase33AbortLikeWideningMutantKilled$|^TestPhase33CaptureReadBounded$|^TestPhase34BoolNameBaseline$|^TestPhase34BoolSpellingsRemainIdentifiers$|^TestPhase34CallArityBoundaries$|^TestPhase34CheckDiagnosticPresentation$|^TestPhase34CompatibilityDeltaLedger$|^TestPhase34DiagnosticIdentityProjection$|^TestPhase34DiagnosticLocations$|^TestPhase34DiagnosticRouteInventory$|^TestPhase34DraftBaseline$|^TestPhase34HumanSnippetRendering$|^TestPhase34IdentityBaseline$|^TestPhase34ImportedDiagnosticPresentation$|^TestPhase34InterfaceLimitWitnesses$|^TestPhase34LinearBindingLimitScope$|^TestPhase34PreChangeCGoldenIsNotInert$|^TestPhase34PreChangeCProgramGolden$|^TestPhase34PresentationUsesCapturedSource$|^TestPhase34RecoveryControls$|^TestPhase34RecoveryCorpus$|^TestPhase34RefusalHintsHaveNoRepairs$|^TestPhase34RepairJSONCompatibility$|^TestPhase34SourceRoutePresentation$|^TestPhase34SurfaceDocument$|^TestPhase34SurfaceLimitConsumers$|^TestPhase34SurfaceLimits$|^TestPhase34SurfaceManifestCensus$|^TestPhase34SurfaceManifestControls$|^TestPhase34SurfaceManifestWitnesses$|^TestPhase34SurfaceRoute$|^TestPhase34SurfaceRouteInventory$|^TestPhase34TransitionDocuments$|^TestPhase34UnsupportedSurfaceRefusals$|^TestReceiptGateArchivedM006$|^TestReceiptGateDifferential$|^TestReceiptGateMalformedControls$|^TestRepairDriverDecodesNoProseFields$|^TestRepairIdentityUsesKindOnly$|^TestRepairSpanShiftDoesNotChangeDiagnosticID$|^TestSkippedRegionCauseSeparatesSpiralTrio$|^TestStaticTestIndexMatchesGoTestList$|^TestSummaryCoverageGroundedness$|^TestValidProgramUnaffectedBySkippedRegionFix$|^TestVerificationGroundednessFrontierIsPinned$|^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1 -v
```

| Package | Reported time |
|---|---|
| cmd/schway | 4.594s |
| cmd/schway-repair | 0.126s |
| internal/compiler/cgen | 0.060s |
| internal/compiler/check | 0.092s |
| internal/compiler/diagnostic | 0.016s |
| internal/compiler/interp | 0.027s |
| internal/compiler/native | 0.115s |
| internal/compiler/protocol | 0.025s |
| internal/compiler/session | 22.862s |
| internal/compiler/surfacemanifest | 0.009s |
| internal/compiler/syntax | 0.037s |

The passing controls include exact/oversize capture, narrowed interpreter defects and widening mutant, generic freeze/receipt differentials and malformed controls, historical M006 receipt acceptance, reopened living markers, byte-identical C corpora, public emitter inventory, static test-index accuracy, summary grounding, and validation lifecycle. The route, recovery, location, identity, repair, manifest and transition tests retain their own negative controls. Source-reference consumer evidence is finite and branch-insensitive; interface thresholds and the aggregate AST guard are separately executed.

No whole-package/project suite, native matrix, push, hosted dispatch, or final receipt binding ran. Phase 38 retains final hosted ownership. Human readability and whole-loop latency remain unmeasured.
