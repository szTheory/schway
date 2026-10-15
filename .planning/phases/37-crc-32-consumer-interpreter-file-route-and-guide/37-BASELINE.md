# Phase 37 Pre-Change Baseline

Captured: 2026-10-15T14:01:58Z, before creating any Phase 37 Schway source or editing either historical inventory owner.

## Checkout and toolchain

- HEAD: `b1b00b84a5d46d0075424a67461df1dee9d4166f`
- Branch: `phase26-validation`
- Dirty implementation sources (`*.go`, `*.schway`, `*.c`, `*.h`): none.
- Pre-existing planning changes: `.planning/STATE.md`, `.planning/config.json`, `.planning/state.json`; untracked `.planning/milestone.lock` and `.planning/research/.cache/`. These were present before this baseline and are not source changes for this task.
- Tracked `.schway` source count before additions: 166.
- Host: Darwin 25.6.0 arm64.
- Go: `go1.24.0 darwin/arm64`.
- Clang: Apple clang 21.0.0 (`clang-2100.1.1.101`).
- Oracle environment: Python 3.14.4; zlib compile and runtime versions 1.2.12.
- Phase 34 pre-change corpus SHA-256: `903010268526ce71dab023e562714a5a8e18329a4d91adca9678e5d64d50042f`.
- Phase 36 pre-change corpus SHA-256: `b9db7faa10bae93e741f2b24e602a9eda3eb393ed38efb0df6364a9ef9873001`.
- Original Phase 36 additive ledger SHA-256: `99ee245b707175e1b27846ce825ae6371e7eb7da0e810b2d8edf486c3517eaf4`.

## Focused pre-change roots

All commands used `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, verbose output and a 600-second test timeout. Every named root reached `RUN` and `PASS`; no root or subtest was skipped.

1. Legacy public file/conformance route:

   ```sh
   GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s
   ```

   Result: 3/3 roots passed. `TestPhase23PublicFileByte` reached and passed its 13 listed cases/subtests, including the empty-token boundary. Current empty path behavior remains `(65, "", "")`. `TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus` and `TestPhase22ConformanceRunStillEmitsExecutionDocument` also passed. Package result: `ok .../cmd/schway 0.528s`.

2. Isolated Phase 36 no-call baseline:

   ```sh
   GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s
   ```

   Result: 1/1 root passed; package result `ok .../internal/compiler/session 1.235s`.

3. Historical inventory and baseline roots:

   ```sh
   GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s
   ```

   Result: 4/4 roots passed. Both Phase 34 golden roots and both Phase 36 baseline/control roots passed; all 13 named Phase 34 addition controls ran and passed. Package result: `ok .../internal/compiler/session 3.212s`.

The commands emitted only Go test runner output; test assertions capture application streams internally. The Phase 23 public application root pins its exact exit/stdout/stderr tuples. The Phase 34/36 roots validate frozen corpus and C emission equality, their current tracked-source census, and armed negative controls. This receipt records local macOS evidence only; it makes no Linux, sanitizer, maximum-CRC, or hosted-receipt claim.
