# Deferred Items: Phase 36 Plan 36-05

## Frozen no-call corpus refusal shifts

**Kind:** unmet truth / out-of-scope pre-existing validator change  
**Evidence:** `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s`  
**Result:** Generated-C byte snapshots match. Three fixtures differ only in the current core-validator refusal occurring before the historical cgen refusal/application route:

- `examples/checksum/adler_low.schway` and `examples/checksum/checksum.schway` now refuse with `core.scalar_call_context_invalid` where the frozen receipt expected the cgen file-bytes owner refusal and emitted application-shell route.
- `examples/phase24/error.schway` now refuses with `core.scalar_backedge_authority` where the frozen receipt expected the Phase24 application-shell refusal and emitted application-shell route.

These changes originate in the core validator, outside this plan's owned files. Plan 36-05 preserves the independent cgen inventory and frozen C expectations and does not weaken the corpus guard or alter core/peer behavior. Reconcile the historical refusal expectations with the phase that owns those validator changes.

### Resolution amendment — 2026-10-14

The authorized `phase36-legacy-call-routes` debug investigation reproduced all
three failures and confirmed actual application admission regressions, not
refusal-text drift. A one-file pre-feature corevalidate overlay restored the
frozen baseline and controls. The correction scopes call-only scalar analysis
to scalar callers, retains the complete existing file-owner ByteView lifecycle,
and keeps the typed fallible PathToken-to-U64 probe on each peer's independent
owner/error path. Held-out origin tests additionally required the same caller
scope for the file-owner scalar tail and restoration of owner refusal precedence.

Fresh local verification passes the unchanged
`TestPhase36NoCallGeneratedCByteBaseline` and
`TestPhase36NoCallGeneratedCByteBaselineControls`, the four new applicability and
legacy-call regression roots, Phase24 cleanup/activation controls, Phase29
ownership controls, all three peers' `TestPhase31FileBytesSeededFaults`, and the
Phase36 raw call safety/refusal-removal selectors. Seven single-site overlay
mutants were killed at semantic assertions. No frozen corpus/golden or example
source was changed. Full commands, results and prevention are in
`.planning/debug/resolved/phase36-legacy-call-routes.md`; fresh logs are
`/tmp/schway36-final-session.log` and `/tmp/schway36-final-validators.log`.

This closes the implementation residual before Plan36-06, subject to the debug
workflow's final review/archive; earlier failure receipts above remain historical
evidence. No full suite, native matrix or cross-host result is claimed.
