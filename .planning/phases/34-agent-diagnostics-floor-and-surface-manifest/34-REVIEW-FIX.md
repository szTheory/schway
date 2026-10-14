---
phase: 34
review: 34-REVIEW.md
status: fixed
fixed: 3
skipped: 0
---

# Phase 34 — Parent Review Fixes

The initial review remains unchanged as a historical source-review report. Commit `18844d62` resolves its three warnings. No compiler admission, parser/checker enforcement, dependency, or hosted receipt changed.

## Fixed Issues

### WR-01: Source-checking routes are missing from limit applicability

Fixed the parser and scalar consumer lists and the checked-work list. The finite seven-file source-reference census independently follows the command entry points, including query's closed dispatch table. It found additional evidence/explain/query/verify applicability beyond the initial review. Conditional debug-map/import and query/lane-corpus guards are explicit in the published meaning.

### WR-02: Linear-binding limit reports per-body scope while enforcing an aggregate

The scope and meaning now describe the aggregate outer function linear-body binding lists. `TestPhase34LinearBindingLimitScope` directly reaches 65536 and 65537 bindings split across two individually smaller bodies, and restores the exact cap. Earlier source/token limits can prevent text from reaching this AST guard; no source-level reach is claimed.

### WR-03: Surface limit witnesses do not independently check route behavior

`TestPhase34SurfaceLimitConsumers` follows consumers independently of the registry/witness literals, covers all nine parser-limit route lists, and detects omitted/false route mutations. `TestPhase34InterfaceLimitWitnesses` runs both real interface producers at the exact function cap, above that cap, and above the global work cap; refused sources publish no artifact. Inherited import/no-import debug-map controls and scalar-analysis exhaustion are retained. This remains finite consumer evidence, not branch-sensitive whole-program analysis or proof of every budget's public threshold.

## Fresh Verification

- Session: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase34SurfaceLimitConsumers$|^TestPhase34InterfaceLimitWitnesses$|^TestPhase34SurfaceManifestWitnesses$|^TestPhase34SurfaceManifestCensus$|^TestPhase34SurfaceManifestControls$|^TestPhase30ModuleWorkLimit$|^TestPhase30DebugMapNoImportUnguarded$' -count=1 -v` — all seven named tests passed, 1.010s package time.
- Producer: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/surfacemanifest -run '^TestPhase34SurfaceDocument$|^TestPhase34SurfaceLimits$' -count=1 -v` — both tests passed, 0.006s.
- AST guard: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/syntax -run '^TestPhase34LinearBindingLimitScope$' -count=1 -v` — passed, 0.016s.
- Scalar exhaustion: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/check -run '^TestPhase26AnalysisExhaustion$' -count=1 -v` — passed, 0.008s.

These are parent-run focused receipts. No full local suite, native matrix, network operation, or hosted binding ran.
