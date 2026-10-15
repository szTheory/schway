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
