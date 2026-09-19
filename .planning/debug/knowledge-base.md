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
