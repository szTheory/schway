# Phase 37.1 Older Comparator Regression Repair

This is a narrow discovered-regression repair after the active Phase 37.1 plans completed. It restores the old C comparators' strict identity check for the existing `sum_to_n` exception by reusing the measured Phase371 amendments. No phase plan/summary, source fixture, frozen golden, corpus, amendment ledger, production emitter, or Phase371 adapter/projection was changed.

## Measured binding

The live source and checked core still match the Phase371 census:

| Input | SHA-256 |
| --- | --- |
| `examples/sum_to_n.schway` source | `6e3a5c992eb45a741ba17d55df3c378e681a7b151544d6625000cc23411f3aa6` |
| Checked core | `b149a1761f25983ae17d817bc034bc03da8a3f12e5a699c236deb929a0055b36` |
| Old C for all five existing amendment routes | `163eb1d42ad269797fb0ea95be9252c643fb495e1f10d4bba372979186bed02c` |
| Exact post-amendment C for all five existing amendment routes | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |

The measured route lookup matched exactly for Phase34 `control_native_c`, `emit`, and `emit_native`, plus Phase36 `emit` and `emit_native`. The five existing ledger keys remain unchanged. Each older comparator now verifies its source/core route binding and exact amended C through a test-only bridge to Phase371's existing strict comparator. Phase35 additionally requires its frozen C bytes to equal the bound old C before applying the amendment. Phase34's historical corpus did not record a route schema, so that comparison remains source/core/C-bound; Phase36 continues to check its recorded schema.

For Phase273's frozen golden, reversing the single named capacity-exit line in the measured old C yields SHA-256 `847c72c45f5338ea1ea476eaf5ccab50aef88eb169c449268d8df2e93d766e83`, the immutable golden hash. This composes the prior named-exit delta with the existing Phase371 projection without changing either receipt.

## Verification

The following focused session roots passed with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, and the listed timeout:

```text
go test ./internal/compiler/session -run '^(TestPhase273ConformanceCorpusByteIdentity|TestPhase28PrunedEmitMatchesGolden|TestPhase35PreChangeEmitterCorpus|TestPhase35PreChangeCorpusControls|TestPhase371PreEmitterRouteCensus|TestPhase371CompatibilityAmendmentControls)$' -count=1 -v -timeout=120s
```

This reached all four regression roots, the census, and the amendment-control root. The controls passed for missing, duplicate, extra, wrong-path, wrong-route, stale-source, stale-core, stale-old-C, wrong-expected-digest, wrong-anchor, wrong-offset, wrong-anchor-count, wrong-delta, unrelated post-amendment bytes, and repeated anchors. The older-comparator control also rejects an unrelated output-byte mutation, a stale source digest, and an `emit` C value presented as `emit_application`.

The inherited six-root baseline also passed, bounded at 600 seconds:

```text
go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s
```

All non-target rows and routes continue through their original exact comparators. Frozen raw goldens, the Phase34/36 corpora, and the five-row Phase371 ledger remain byte-for-byte unchanged.

## Changed bound test inputs

These post-repair file hashes are recorded for the existing source-evolution history refresh; they are not new allowance keys:

| Test input | SHA-256 |
| --- | --- |
| `internal/compiler/session/session_phase273_identity_test.go` | `25cb86e2a577b1cf9719b94717e11d74aecfa7755d552014888fdcfb6175ff87` |
| `internal/compiler/session/session_phase28_prune_test.go` | `ec3cbe3a19efa1b0e364a5f2a431d73481addc4d271934761beccdfe163e5036` |
| `internal/compiler/session/session_phase35_prechange_test.go` | `e3f258f3c4a3b4dc66825e1437ad464cc13f32bbd0a48326cb275c38e3627f64` |
| `internal/compiler/session/session_phase371_compatibility_test.go` | `1fcfb16979194559307a3e9c494052889a5066d42fdf011d154c3479f112e5e4` |
