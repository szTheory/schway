# Phase 19 Scalar Projection Gate

## Verification

Focused production gate:

```sh
go test ./internal/compiler/interp ./internal/compiler/session -run 'TestPhase19(ScalarProjection|LiteralFrontier)|TestPayloadCorpusCharacterizationReplay' -count=1
```

Result: PASS. The new U64 projection cases passed for `0`, `42`, and `18446744073709551615`; the existing scalar and tagged projections passed unchanged. `TestPayloadCorpusCharacterizationReplay` passed against the pinned corpus.

Mutation and frontier gate:

```sh
go test ./internal/compiler/session -run 'TestPhase19ScalarGate|TestPayloadCorpusCharacterizationReplayMutationKilled|TestPayloadCorpusCharacterizationReplay' -count=1
```

Result: PASS. The established empty-tag serialization mutation moved the corpus digest, and the Phase 19 tracer fixture retained its Wave 1 refusal ID, code, and span.

## Golden Disposition

No prior D-12-18 scalar golden changed. The exact old/new byte comparison is therefore unchanged; the pinned digest replay passed without updating any golden. No cause for golden movement applies.

## Dispatch Boundary

No `OpConst` operation definition, dispatch case, or consumer was changed in this gate.
