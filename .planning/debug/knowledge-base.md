---
status: complete
---

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

## phase15-lto-diagnostic-order — Schema-2 preflight masked the established unsupported-Match refusal
- **Date:** 2026-09-19
- **Error patterns:** unsupported linear C type "Switch", multi-function branch bodies, diagnostic precedence, TestLTOInertnessOnMultiFunctionEmission
- **Root cause(s):** emitProgram performed schema-2 occurrence/output preflight before supported-body validation, allowing schema2ExecutionDocumentSize's linearInput call on a Match entry to mask the established multi-function branch-body refusal
- **Fix:** Move ordered supported-shape validation ahead of the unchanged invocation-path, event-capacity, and exact output-size preflights; add a direct cgen diagnostic-precedence regression
- **Files changed:** internal/compiler/cgen/cgen_program.go, internal/compiler/cgen/cgen_program_test.go, .planning/phases/15-event-identity-lang-execution-2/15-VERIFICATION.md
- **Why not caught:** The native-capacity regression suite proved bounds for supported straight-line programs but had no unsupported-body diagnostic-precedence case
- **Recurrence guard:** Regression test internal/compiler/cgen/cgen_program_test.go:TestUnsupportedProgramShapePrecedesSchema2Preflight, plus existing 4096/4097 occurrence and N-1/N output-bound controls
---

## phase-15-schema-admission-ci — Schema 2 wire-admission evidence gap
- **Date:** 2026-09-26
- **Error patterns:** native Schema 2 admission, ToolError diagnostics, UAT coverage metadata
- **Root cause(s):** The validator's direct unit test did not cover canonical JSON decoding or ToolError wrapping, and Plan 01 omitted machine-readable coverage.
- **Fix:** Plan 15-08 added the canonical-bytes-to-decoder-to-ToolError seam, refusal-specific diagnostics, and coverage metadata; Phase 15 UAT gap G-15-1 is resolved.
- **Files changed:** internal/compiler/native/native_test.go, .planning/phases/15-event-identity-lang-execution-2/15-01-SUMMARY.md
- **Why not caught:** The direct validator test bypassed the wire boundary and summary prose was not machine-readable to verify-work.
- **Recurrence guard:** TestDecodeExecutionSchema2AdmissionSeam and the named Phase 15 native admission CI step.
---

## phase-15-peer-gate-ci — Legacy comparator and peer-gate evidence gap
- **Date:** 2026-09-26
- **Error patterns:** Schema 0/1 preservation, Schema 2 peer, UAT coverage, CI provenance
- **Root cause(s):** Plan 06 lacked machine-readable coverage and a direct legacy-wrapper regression; the named phase gate was stale.
- **Fix:** Plan 15-09 added Schema 0/1 wrapper tests, coverage metadata, and a focused cross-platform Phase 15 CI aggregate; G-15-11 is resolved.
- **Files changed:** internal/compiler/session/session_phase5_compare_test.go, internal/compiler/session/session_phase6_test.go, .github/workflows/ci.yml, Phase 15 Plan 06 artifacts
- **Why not caught:** Existing Schema 2 peer tests did not prove legacy wrapper behavior or publish coverage to the UAT classifier.
- **Recurrence guard:** TestPhase5CompareProgramEnginesPreservesLegacySchemas and TestCIWorkflowRunsCurrentAggregateGate.
---

## phase-15-diamond-gate-ci — Four-tier diamond evidence gap
- **Date:** 2026-09-26
- **Error patterns:** shared-leaf diamond, collision control, UAT coverage, CI provenance
- **Root cause(s):** Plan 07 lacked machine-readable coverage and the current CI aggregate did not identify Phase 15's diamond seams.
- **Fix:** Plan 15-09 published automated coverage and added the current cross-platform aggregate; G-15-12 is resolved and Phase 15 UAT passes 12/12.
- **Files changed:** internal/compiler/session/session_phase6_test.go, .github/workflows/ci.yml, Phase 15 Plan 07 artifacts
- **Why not caught:** The tests ran under general Go CI, but their phase-specific traceability and named provenance were absent.
- **Recurrence guard:** TestPhase11InterproceduralDifferential/DiamondSharedLeaf, TestPhase15CollisionGuardIsNotInert, and the current Phase 15 CI aggregate.
---

## phase16-debt-and-machine-probe — Stale witness and sandbox-dependent tests
- **Date:** 2026-09-26
- **Error patterns:** Phase 13 witness references, TestDebtRegistersAreWellFormed, Darwin sysctl, measure.probe_failed
- **Root cause(s):** Archived debt rows named a replaced Phase 17 witness, and budget tests depended on a host sysctl probe denied by the sandbox.
- **Fix:** Plan 16-19 synchronized the archive witness and used injected MachineFacts for deterministic budget assertions while retaining bounded probe tests; G-16-21-C is resolved.
- **Files changed:** .planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md, internal/compiler/session/session_phase6_budget_test.go
- **Why not caught:** Historical witness tokens and host-specific subprocess permissions were not represented by deterministic test fixtures.
- **Recurrence guard:** TestDebtRegistersAreWellFormed, TestUnreachableClaimsViewIsCurrent, TestBudgetLaneCarriesMachineIDAndVerdict, and bounded commandFactory probe tests.
---

## phase16-groundedness-owner — Grouped alternation caused a false frontier finding
- **Date:** 2026-09-26
- **Error patterns:** groundedness, Go -run alternation, unowned R2b
- **Root cause(s):** The scanner splits literal alternation branches and misclassified a grouped selector as an unowned finding.
- **Fix:** Plan 16-26 expressed the same test selection as independently anchored branches; G-16-21-E is resolved.
- **Files changed:** .planning/phases/16-branch-match-emitter-port/16-VERIFICATION.md
- **Why not caught:** The selector was semantically valid to Go but not independently parseable by the project's groundedness scanner.
- **Recurrence guard:** TestVerificationGroundednessFrontierIsPinned, TestVerificationGroundednessThreeClassesAreEmpty, and the recorded selector pattern in Plan 16-26.
---

## phase16-m004-phase11-gates — Phase 11 consumers used a retired live-emission path
- **Date:** 2026-09-26
- **Error patterns:** Phase 11 zero-attribute gates, by-pointer emission, cut-m004, frozen evidence
- **Root cause(s):** The emitter inventory classified by-pointer Phase 11 consumers as dynamically admitted after Phase 16 had correctly cut that path.
- **Fix:** Plan 16-21 moved the checks to test-only refusal-first use of provenance-bound frozen C and added inventory/mutation controls; G-16-21-A is resolved.
- **Files changed:** internal/compiler/session/session_phase11_gate_test.go, internal/compiler/session/session_phase16_emitter_inventory_test.go, internal/compiler/session/witness_registry_test.go, testdata/phase16/public-emitter-consumers.json
- **Why not caught:** The source-derived consumer inventory did not initially encode the M004 refusal disposition for those exact consumers.
- **Recurrence guard:** TestPhase16PublicEmitterConsumerInventory, TestPhase16M004ProvenanceRegistryRejectsFaults, and TestPhase16ProductionPathsPreserveM004Refusal.
---

## phase16-validation-frontier-drift — Corpus and groundedness snapshots lagged source evolution
- **Date:** 2026-09-26
- **Error patterns:** validation corpus digest, groundedness frontier, R2b ownership
- **Root cause(s):** Committed test-index and evidence-document changes landed after the corpus record and exact frontier were sealed.
- **Fix:** Plans 16-17/18 refreshed the derived records; Phase 20 Plans 07/10 reconciled the live frontier and final 33-pair corpus record; G-16-21-B is resolved.
- **Files changed:** testdata/phase16/validation-corpus-run-record.jsonl, testdata/phase16/validation-corpus-run-record.manifest.json, internal/compiler/session/verification_groundedness_test.go, Phase 14/16 validation artifacts
- **Why not caught:** Snapshot inputs changed after their prior pinned revision, while the exact-set and digest checks correctly failed closed.
- **Recurrence guard:** TestVerificationGroundednessFrontierIsPinned, TestValidationRowGradesAreEarnedOverArchivedCorpus, and the Phase 20 corpus manifest.
---

## phase25-pathoracle-regression — Local loan replay misclassified call and declared borrowed returns
- **Date:** 2026-10-02
- **Error patterns:** core.pathoracle_refused, pathoracle.pointer_escape, OpCall result ancestry, repair reverify_failed, Phase 25 full-suite regression
- **Root cause(s):** The straight-line local loan replay propagated source ancestry through every non-copy operation, including OpCall whose return provenance depends on callee contracts, and treated every loan-derived return as an owned-result escape even when a borrow return was declared.
- **Fix:** Break local result ancestry at OpCall while continuing to count source loans as call uses; only reject loan-derived returns when the function has no declared return origin; keep contract-aware peers authoritative for call result and borrowed-return provenance.
- **Files changed:** internal/compiler/pathoracle/pathoracle.go, internal/compiler/pathoracle/pathoracle_pointer_successor_test.go, internal/compiler/session/verification_groundedness_test.go, testdata/phase16/public-emitter-consumers.json
- **Why not caught:** The original focused oracle controls did not exercise OpCall-return provenance or declared borrowed-return exits, and the local package check did not compose origin/admission/repair consumers; the uncached repository-wide suite exposed the gap.
- **Recurrence guard:** `TestPhase25PointerPathDefersCallResultProvenance`, `TestPhase25PointerPathAllowsDeclaredBorrowReturn`, `TestPhase25PointerPathChecksOverlapAcrossCalls`, and `TestPhase25PointerPathChecksEscapeInCallFunction` in `internal/compiler/pathoracle/pathoracle_pointer_successor_test.go`, plus the uncached full Go suite.
---
