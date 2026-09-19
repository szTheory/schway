# GSD Debug Knowledge Base

Resolved debug sessions. Used by `gsd-debugger` to surface known-pattern hypotheses at the start of new investigations.

---

## phase15-full-suite-regressions — Phase 15 aggregate gate failed across schema, process-policy, payload, and evidence contracts
- **Date:** 2026-09-19
- **Error patterns:** lang.execution/1 versus lang.execution/0, unbounded subprocess, canonical payload hash mismatch, validation grade, groundedness, full-suite regression
- **Root cause(s):** Phase 15 interpreter schema selection erased the legacy bare-match /0 versus frame-based /1 distinction; the new executionpeer import-boundary test used an unbounded subprocess constructor and output helper; the intentional multi-function /1-to-/2 migration lacked a versioned successor characterization ledger; Phase 15 evidence used ungrounded command shapes, a stale test name, and omitted the mandatory Grade column
- **Fix:** Preserve caller-selected legacy schemas and override only multi-function programs to /2; bound executionpeer's subprocess and both output streams; add an explicit 18-entry Phase 15 /2 successor digest ledger while retaining the Phase 12 baseline; ground Phase 15 evidence, derive grades, close the two event-identity debts, and synchronize the generated claims view
- **Files changed:** internal/compiler/interp/interp.go, internal/compiler/interp/interp_test.go, internal/compiler/executionpeer/executionpeer_test.go, internal/compiler/session/session_payload_replay_test.go, .planning/phases/15-event-identity-lang-execution-2/15-RESEARCH.md, .planning/phases/15-event-identity-lang-execution-2/15-VALIDATION.md, .planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md, .planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md, .planning/UNREACHABLE-CLAIMS.md
- **Why not caught:** Focused Phase 15 gates did not compose legacy schema boundaries, repository-wide subprocess policy, versioned payload migration, and planning-evidence laws before the exact build + vet + uncached full-suite closure gate.
- **Recurrence guard:** Schema boundary coverage in internal/compiler/interp/interp_test.go; repository-wide TestSourceNeverSpawnsUnboundedProcesses; TestPayloadCorpusCharacterizationReplay plus its mutation control and 18-entry /2 successor ledger; groundedness, reconciliation, validation-grade, debt-register, and TestUnreachableClaimsViewIsCurrent laws.
---

## phase15-native-capacity — Native schema-2 occurrence evidence exhausted event and output bounds
- **Date:** 2026-09-19
- **Error patterns:** native.run_signaled, exit 74, shared-leaf diamond, LANG_EVENT_CAPACITY, 65536-byte output
- **Root cause(s):** Schema-2 event capacity counted unique declared bodies instead of occurrence-weighted executed operations; schema-2 output inherited a 65,536-byte legacy writer/runner ceiling without a schema-specific preflight size contract
- **Fix:** Derive event capacity from every preflight occurrence; share a 16 MiB execution-document contract between cgen and native capture; calculate exact schema-2 bytes and refuse oversized output before C serialization with cgen.execution_output_exceeded
- **Files changed:** internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen.go, internal/compiler/cgen/cgen_program_test.go, internal/compiler/cgen/export_test.go, internal/compiler/execution/execution.go, internal/compiler/native/native.go, internal/compiler/native/native_test.go
- **Why not caught:** The 61-node test stopped at EmitNative and the 4,096-node test asserted admission only; neither executed native code or checked document bytes
- **Recurrence guard:** Regression tests internal/compiler/cgen/cgen_program_test.go:TestDeepDiamondExecutesAcrossNativeOptimizationTiers and TestSchema2ExecutionOutputBoundIsPreflighted
---
