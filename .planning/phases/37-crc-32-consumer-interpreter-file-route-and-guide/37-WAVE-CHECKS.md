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
