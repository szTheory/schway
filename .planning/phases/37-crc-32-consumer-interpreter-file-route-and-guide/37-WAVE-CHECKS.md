# Phase 37 parent wave checks

## Wave 1 — 2026-10-15

Canonical worktree base-check selected sequential execution because the branch HEAD differs from origin/HEAD. User worktree preferences remain unchanged; each plan gets a fresh plan-scoped `none` dispatch sentinel and supplied root pin.

37-01 terminal handoff and committed summary were reconciled against six reachable plan-subject commits. First two created source artifacts exist; summary contains `Self-Check: PASSED` and no failed marker.

Fresh parent checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go build -o /tmp/schway37-wave1 ./cmd/schway`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s` — all four roots reached and passed, including both native outputs and ledger negative controls.
- Canonical wave-post schema-drift and UI gates returned block=false; codebase-drift skipped for no structure map. UI scope was degraded/no-task-commit-rows, so this is no UI quality claim.

The project instruction forbids local project/whole-package suites. The post-wave build and test gate therefore used the changed CLI and named focused integration selectors rather than the generic whole-repository fallback.

The executor's generic requirement update prematurely marked APP-13, MOD-07, and APP-14 complete. Parent used canonical `requirements.revert-phase APP-13 MOD-07 APP-14 DX-24`; all four remain pending phase-level acceptance. The summary's native receipts stand, but its requirement list does not establish the later interpreter route or matrix.

## Wave 2 — 2026-10-15

37-02 terminal handoff and committed summary were reconciled against seven reachable plan-subject commits. First two created artifacts exist; no failed self-check marker.

Fresh parent checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go build -o /tmp/schway37-wave2 ./cmd/schway`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/session -run '^(TestPhase37InterpreterFileApplication|TestPhase37CaptureDisabledRetention|TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderBasic)$' -count=1 -v -timeout=600s` — all four roots reached and passed. The maximum input observation was 2.536121167 seconds on darwin/arm64, Go 1.24.0; this is one host observation, not a latency distribution or speed guarantee.
- Canonical wave-post schema and UI gates block=false; codebase-drift skipped for absent structure map. The UI scope remains degraded/no-task-commit-rows, not quality evidence.

Requirements remain pending phase acceptance. The new outcome controls found and fixed exhausted capture obscuring a later checked defect. Executor canonical proof receipts are source-bound to its implementation head; final phase proof refresh remains required after later changes.

Wave 3 readiness: 37-03 and 37-04 are both ready and have disjoint declared file sets. They still serialize because this run's negotiated isolation is none.

## Wave 3 — 2026-10-15

37-03 and 37-04 terminal handoffs and committed summaries were reconciled against four reachable plan commits each; first created artifacts exist, no failed self-check markers. Their objective automated coverage is not human UAT: parent corrected 37-03's two `human_judgment` flags, which had described parent workflow ownership rather than human-dependent product qualities. No verification result was changed.

Fresh parent checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go build -o /tmp/schway37-wave3 ./cmd/schway`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunGrammar|TestPhase37OpaqueFilePaths|TestPhase37FileRunInputBoundary|TestPhase37FileRunOutcomes|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37FileApplicationAuthority$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37FileReaderFaults|TestPhase37FileReaderBoundControls|TestPhase37CaptureRetentionControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^(TestPhase37CRC32PhysicalCleanup|TestPhase37CRC32BoundControls|TestPhase37CRC32Sanitizer)$' -count=1 -v -timeout=600s`

All 13 exact roots reached and passed, with no skipped subtests in these runs. Native missing-free and one-byte-short-read mutants built and reached their physical assertions. Combined ASan/UBSan was available and passed generated CRC small/cap success plus an actual-adapter cap-plus-one returning C caller; no leak detection or Linux-host execution claim. The generated application's typed-error exit remains distinct from the returning C caller.

Wave-post schema/UI block=false; codebase-drift skipped for no structure map. Wave 4's three plans are ready and declared file sets disjoint; serialization remains in force for this run. All four phase requirements remain pending final acceptance.

## Wave 4 — 2026-10-15

37-05/06/07 terminal handoffs, summaries, first created files, and reachable plan-subject commits were reconciled (4/3/7 commits respectively, including metadata). No failed self-check markers. Actual scope deviations are retained in the summaries: 37-06's existing InputID semantics (see 37-BUILD-IDENTITY-DECISION.md) and 37-07's additive Phase37 surface witness preserving the Phase35-pinned Phase34 receipt.

Fresh parent checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go build -o /tmp/schway37-wave4 ./cmd/schway`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileMatrix|TestPhase37PublicMatrixDriftControl|TestPhase37CRC32AlgorithmMutants|TestPhase37CRC32LibraryDualFlip|TestPhase32PublicRouteAudit|TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory|TestPhase37FileRunRouteInventoryControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32SingleCopyAndIdentity|TestExitCodeDocsTable|TestPhase37FileRunExitContract|TestPhase34SurfaceLimitConsumers|TestPhase37FileRouteSurfaceFacts)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/testsupport -run '^(TestPhase37CRC32OracleRows|TestPhase37CRC32OracleDriftControls)$' -count=1 -v -timeout=600s`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest -run '^(TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument|TestPhase36SurfaceDocument)$' -count=1 -v -timeout=600s`

All 19 exact roots reached and passed with no root or subtest skips. Every representable matrix case ran for both engines/four consumers, followed by separate direct NUL, manifest and isolated process controls. All five algorithm controls reached independent numeric disagreement; no-edit InputID/module rows stayed stable and actual artifact-bound identity observations remained explicit. Raw command/session output was inspected in /tmp/schway37-wave4-{cmd,session}.log (ephemeral local files, not durable receipt artifacts).

Fresh wave-post hooks: schema and UI block=false, codebase-drift skipped for no structure map. Only37-08 is now ready. Requirements remain pending final phase acceptance.
