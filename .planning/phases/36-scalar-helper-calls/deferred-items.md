# Deferred Items: Phase 36 Plan 36-05

## Frozen no-call corpus refusal shifts

**Kind:** unmet truth / out-of-scope pre-existing validator change  
**Evidence:** `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s`  
**Result:** Generated-C byte snapshots match. Three fixtures differ only in the current core-validator refusal occurring before the historical cgen refusal/application route:

- `examples/checksum/adler_low.schway` and `examples/checksum/checksum.schway` now refuse with `core.scalar_call_context_invalid` where the frozen receipt expected the cgen file-bytes owner refusal and emitted application-shell route.
- `examples/phase24/error.schway` now refuses with `core.scalar_backedge_authority` where the frozen receipt expected the Phase24 application-shell refusal and emitted application-shell route.

These changes originate in the core validator, outside this plan's owned files. Plan 36-05 preserves the independent cgen inventory and frozen C expectations and does not weaken the corpus guard or alter core/peer behavior. Reconcile the historical refusal expectations with the phase that owns those validator changes.
