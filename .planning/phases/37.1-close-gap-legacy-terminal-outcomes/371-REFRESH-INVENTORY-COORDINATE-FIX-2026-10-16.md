# Phase 37.1 inventory coordinate follow-up — 2026-10-16

The resumed prior-phase inventory check first returned RED on two exact Phase 37.1 compatibility calls. Its AST scanner reported session_phase371_compatibility_test.go:Emit:782 and EmitNative:784; the frozen registry still recorded those same calls at lines 775 and 777. The source and classifications were correct, and the inventory remained strict.

The authorized data-only repair changed only those two line coordinates in testdata/phase16/public-emitter-consumers.json. Both rows retain admitted-dynamic-schema2 and probe:TestPhase371PreEmitterRouteCensus. No compiler or test source changed. The post-fix inventory and refusal-witness roots passed 2/2 with no skips.

- RED receipt: 371-REFRESH-COMPAT-20261016.jsonl, SHA-256 d0d910817d8faad635898330c55b1e87feeb0be8a9e2b4d400a526f618f247f7.
- GREEN receipt: 371-REFRESH-INVENTORY-20261016.jsonl, SHA-256 7d9ef16f5c9b2e8f11929498edf2b869078d3421d6c64637ebe12dd73b38c784.
- Current registry SHA-256: cc7f0977f4ef11d10d9113af97688fffa435bc1d2e60360d911e995e10332209.
