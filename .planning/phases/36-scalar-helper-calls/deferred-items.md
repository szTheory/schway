# Phase 36 Plan 07 blocker

Plan 36-07 is not complete. The required Task 2 focused verification reaches the existing complete-program consumer controls and exposes a production PathOracle refusal in the frozen Phase 24 ownership-transfer witness. Do not mark the plan complete until the compatibility defect is repaired and the exact verification command passes.

The exact failing command is:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session ./internal/compiler/surfacemanifest -run '^(TestPhase36SurfaceManifestCallCensus|TestPhase36SurfaceManifestCallControls|TestPhase36SurfaceLimitConsumers|TestPhase36SurfaceDocument|TestPhase34SurfaceManifestWitnesses|TestPhase34SurfaceManifestCensus|TestPhase34SurfaceManifestControls|TestPhase34SurfaceLimitConsumers|TestPhase35SurfaceManifestWitnesses|TestPhase35SurfaceManifestCensus|TestPhase35SurfaceManifestControls|TestPhase35SurfaceLimitConsumers|TestPhase34SurfaceDocument|TestPhase34SurfaceLimits|TestPhase35SurfaceDocument)$' -count=1 -v -timeout=600s
```

`TestPhase36SurfaceManifestCallControls` fails before any mutation control because eight retained complete-program witnesses no longer pass. Each reports `invalid/core.pathoracle_refused`, with hint `pathoracle.scalar_call_context: caller "s1:phase24.transfer:fn:main" operation "s1:phase24.transfer:fn:main:op:2" has an invalid scalar call contract or carries non-value authority`:

| Legacy witness | Facet |
| --- | --- |
| `file-byte-owner-adapter` | `type:file_byte_owner` |
| `path-token-checked-parameter` | `type:path_token` |
| `path-token-adapter-input` | `type:path_token` |
| `unit-foreign-release` | `type:unit` |
| `borrow-linear` | `statement:borrow` |
| `borrow-mut-linear` | `statement:borrow_mut` |
| `foreign-block-policy` | `statement:foreign_block` |
| `try-foreign-adapter` | `statement:try` |

Exact repro: `examples/phase24/transfer.schway` is loaded by those full-session witness IDs. `main` obtains a resource-backed value, then calls `shared_copy(value)` and `exclusive_copy(shared)`; `shared_copy` uses `borrow value`, while `exclusive_copy` uses `borrow mut value`. PathOracle refuses the scalar-call context at `main` operation 2. The witness contract expects a complete checked pass. Preserve these owner/path proofs and their expected outcomes; do not replace them with checker-only evidence or supersede them as obsolete scalar-call claims. Route the production PathOracle compatibility correction through GSD debug, then rerun the exact command above.

The focused command did pass the Phase 36 call census, Phase 36 limit consumers and document, Phase 34/35 census and limit consumers, Phase 34 document/limits, and Phase 35 document. Phase 34/35 full witness/control tests and the Phase 36 control baseline are blocked by the same eight outcomes. The earlier frozen fixture bytes remain unchanged:

| Fixture | SHA-256 |
| --- | --- |
| `testdata/phase34/surface-witnesses.json` | `823ac61fac34274ba6ee4cb8719ab0618443600fc976ca5587e1ffee76869742` |
| `testdata/phase34/surface-census.json` | `3934469136ed0f741dbbbca79428a98bc8134c7bfc0e5862ca440e85a94c256a` |
| `testdata/phase35/surface-witnesses.json` | `0fcfe36a942a5670a9819bdc789b4aebcf2182e336bbef9b9ef9d8a11fd2f64a` |
| `testdata/phase35/surface-census.json` | `9af1ecafc10e01bfcdca8882b44c85019f637c51790207f01808013e2461c98f` |
