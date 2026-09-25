# Phase 19 Scalar and Frozen C Review

## D-12-18 and D-12-19 scalar execution baselines

**Result: PASS; no baseline digest changed.** `TestPayloadCorpusCharacterizationReplay` passed. It verified all 62 pinned interpreter corpus digests in `payloadCorpusBaseline`, using the 18 explicit Phase 15 schema-2 successor digests where the replay ledger names them, and confirmed all 35 named expected skips remained skips. The baseline is a post-interpreter-widening snapshot, as already documented by `TestPayloadCorpusCharacterizationReplay`; it is a forward regression tripwire rather than a literal pre-widening capture. No source or golden change was needed.

The Phase 19 scalar checks also passed for decimal `42`, zero, maximum U64, hexadecimal `0x2A`, and binary `0b10_1010`. Every tier returned the expected canonical decimal value, and the five-axis comparator passed all pairs among interpreter, `-O0`, `-O3`, and `-O3 -flto`. The shared-wrong-result control confirmed that comparator equality alone is insufficient: all four tiers can agree on `43`, while the explicit expected-value assertion rejects it.

## Frozen generated-C digests

`TestPreviousPhaseGoldenCUnchanged` and `TestPhase16GoldenChangeLedger` passed. The four exact frozen files and their current SHA-256 results are:

| Baseline | Expected SHA-256 | Current SHA-256 | Result |
|---|---|---|---|
| `testdata/phase1/generated.golden.c` | `1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a` | `1fd8aff8ee28de7ec39e559a7ca9ce50e480ecfffede617c36b2282c60cc122a` | unchanged |
| `testdata/phase2/owned_transfer.golden.c` | `f324f24db3ca0dfa8006b5c7fbec4263a6daf2dcbe220d49ea19167920799686` | `f324f24db3ca0dfa8006b5c7fbec4263a6daf2dcbe220d49ea19167920799686` | unchanged |
| `testdata/phase4/foreign_layout_mismatch.golden.c` | `3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031` | `3be6ebc36032ac9cc29bb916c1cdb8a8a996c0028ddf6982546f4c3dd5ffd031` | unchanged |
| `testdata/phase5/restrict_borrow.golden.c` | `05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0` | `05a16af7e57c3a1a1e2b9af1eb4bed689d89fa53ff91e51328d51dd6f64e38f0` | unchanged |

The file-frozen and generated-frozen Phase 16 artifact ledgers also passed `TestLegacyEmitterEvidence`; their artifact and provenance digests remain unchanged. No generated-C or scalar golden movement has a causal diff to review.

## Commands and outcomes

- `go test ./internal/compiler/session ./internal/compiler/core -run 'TestPhase19(FourTier|LiteralRun|WrongResult|Dispatch)|TestPayloadCorpusCharacterizationReplay|TestAllOperationKindsHandledAtEverySite' -count=1` — PASS.
- `go test ./internal/compiler/core -run 'TestPreviousPhaseGoldenCUnchanged|TestPhase16GoldenChangeLedger' -count=1` — PASS.
- `go test ./internal/compiler/cgen -run '^TestLegacyEmitterEvidence$' -count=1` — PASS.
- `go test ./internal/compiler/session -run '^TestPayloadCorpusCharacterizationReplay$' -count=1` — PASS.
