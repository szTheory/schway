---
phase: "16"
slug: "branch-match-emitter-port"
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-19"
---

# Phase 16 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go standard `testing` |
| **Config file** | `go.mod` |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase16-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen/... 'TestN1ConvergenceDifferential|Test.*Program'` |
| **Full suite command** | `env GOCACHE=/tmp/ai-lang-phase16-cache go test ./...` |
| **Estimated runtime** | ~192 seconds |

## Sampling Rate

- **After every task commit:** Run the focused `scripts/assert-go-tests.sh` cgen command for the affected behavior.
- **After every plan wave:** Run `env GOCACHE=/tmp/ai-lang-phase16-cache go test ./...`.
- **Before `$gsd-verify-work`:** The full suite must be green; run the exact-shape `restrict` lane on macOS and Linux before any by-pointer admission decision.
- **Max feedback latency:** 30 seconds for focused cgen checks.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|-------|----------------|--------|
| 16-01-01 | 01 | 1 | NAT-08 | T-16-01 | N=1 legacy and whole-program paths are byte-identical for scoped fixtures in both public modes. | unit + golden | `go test ./internal/compiler/cgen -run 'TestN1ConvergenceDifferential|Test.*Program' -count=1` | ✅ | EXERCISED | TestN1ConvergenceDifferential | ✅ green |
| 16-01-02 | 01 | 1 | NAT-08 | T-16-02 | Admission preserves graph/entry → shape → preflight → serialization ordering. | unit + mutation | `go test ./internal/compiler/cgen -run 'Test.*Program|TestN1ConvergenceDifferential' -count=1` | ✅ | EXERCISED | TestN1ConvergenceDifferential | ✅ green |
| 16-02-01 | 02 | 2 | NAT-08 | T-16-03 | Golden-change ledger bijects with the digest map and rejects stale or duplicate entries. | unit | `go test ./internal/compiler/core -run 'TestPreviousPhaseGoldenCUnchanged|Test.*Golden.*Ledger' -count=1` | ✅ | EXERCISED | TestPreviousPhaseGoldenCUnchanged | ✅ green |
| 16-03-01 | 03 | 3 | NAT-08 | T-16-04 | Public native dispatch has no function-count route and production lowering uses `emitProgram`. | structural + unit | `go test ./internal/compiler/cgen -run 'Test.*Dispatch|TestN1ConvergenceDifferential' -count=1` | ✅ | EXERCISED | TestN1ConvergenceDifferential | ✅ green |
| 16-04-01 | 04 | 4 | NAT-08 | T-16-05 | Admitted fixtures align dynamically; M004 is current refusal plus frozen provenance. | integration | `go test ./internal/compiler/session -run 'TestPhase11InterproceduralDifferential|TestPhase16M004CorpusRefusal' -count=1` | ✅ | EXERCISED | TestPhase16M004CorpusRefusal | ✅ green |
| 16-05-01 | 05 | 4 | NAT-09 | T-16-06 | Amendment and M004 debt records preserve owner, prerequisite, reopening, and LTO consequence. | document + unit | `go test ./internal/compiler/session -run TestDebtRegistersAreWellFormed -count=1` | ✅ | EXERCISED | TestDebtRegistersAreWellFormed | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

## Executed Gate Records

| Date (UTC) | Revision | Command | Result | Evidence |
|---|---|---|---|---|
| 2026-09-21 | `0607486` cutover, `6827f8f` witness-provenance control | `go test ./internal/compiler/cgen -run 'Test(LegacyEmitterEvidence|FileFrozenEvidenceRejectsFaults|GeneratedFrozenEvidenceRejectsFaults)' -count=1` | PASS | Every historical artifact has a SHA-256-bound fixture, artifact, family refusal, and `probe:TestPhase16M004CorpusRefusal`. Owner: P20 (QLT-10); landing phase: P20. |
| 2026-09-21 | `6827f8f` | `go test ./internal/compiler/core ./internal/compiler/session -run 'Test.*(QLT|Admission|Payload|Witness|EmitterInventory|PreviousPhaseCore)' -count=1` | PASS | Admitted controls compare dynamic schema-2 evidence; the registry/probe/manifest chain rejects missing or stale citation, altered digest provenance, and an M004 reclassification as dynamic admission. Owner: P20 (QLT-10); landing phase: P20. |
| 2026-09-21 | `6827f8f` | `go test ./internal/compiler/cgen ./internal/compiler/core ./internal/compiler/native ./internal/compiler/session -count=1` | PASS | Final package gate. Admitted controls keep dynamic schema-2 byte/convergence evidence; M004 controls are public refusal plus digest-bound frozen provenance, never live emitter admission. |

## M004 Refusal-First Dispositions

| Family | Fixture disposition | Frozen provenance | Current refusal witness |
|---|---|---|---|
| `foreign-m004` | Phase 4 acquisition/nonlocal fixtures and Phase 5 foreign controls remain cut from public `cgen.EmitNative`. | `testdata/phase16/legacy-emitter-evidence.json`, immutable `testdata/phase16/historical/*.c`, per-record fixture/artifact SHA-256. | `probe:TestPhase16M004CorpusRefusal` |
| `by-pointer-m004` | `testdata/phase5/restrict_borrow.lang` remains cut from whole-program public native emission. | `testdata/phase16/legacy-emitter-evidence.json`, `historical/restrict_borrow.c`, fixture/artifact SHA-256. | `probe:TestPhase16M004CorpusRefusal` |
| generated/file-backed controls | Phase 5 generated closure and file controls use only the authoritative manifest matching their fixture/program identity. | `generated-frozen-evidence.json` and `file-frozen-evidence.json`, with generator/program or source/artifact digests. | `probe:TestPhase16M004CorpusRefusal` where cut; admitted rows remain dynamic. |

The public consumer registry schema is `phase16.public-emitter-consumers/2`.
Its inventory test scans every public `Emit`/`EmitNative` call, rejects stale,
duplicate, local-shadow, alias, or dot-import classification drift, and requires
every refusal row to cite a current `probe:` witness. No row claims M004 dynamic
admission.

## Wave 0 Requirements

- [x] Both-mode N=1 byte-identity rows and retained scoped-refusal rows are covered by the consumer inventory and convergence controls.
- [x] Golden/frozen ledger bijection, current-digest, duplicate, stale, and altered-artifact negative controls pass.
- [x] Public dispatch has no function-count route; production lowering uses `emitProgram` for admitted whole programs.
- [x] M004 owner, prerequisite, reopening condition, and LTO consequence remain documented by debt/amendment controls.
- [x] Exact-shape `restrict` probe remains refusal-first M004 evidence; it is not an admission claim.

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Linux exact-shape `restrict` lane | NAT-08 | The current research environment lacks Linux host evidence. | Run the unchanged one-TU microprogram at `-O0`, `-O3`, and `-O3 -flto` with its sanitizer lane on Linux; record the result before admitting by-pointer lowering. |

## Validation Sign-Off

- [x] All tasks have automated verification.
- [x] Sampling continuity has no three consecutive tasks without automated verification.
- [x] Wave 0 covers all missing verification references.
- [ ] No watch-mode flags.
- [ ] Focused feedback latency is under 30 seconds.
- [x] `nyquist_compliant: true` set in frontmatter.

**Approval:** automated evidence complete; M004 remains explicit refusal-only debt.


### Supplemental Task Map (Plans 07–26)

This index extends the original threat-focused map to every task in plans 07–26, including gap-closure plans 16–16 through 16–26. Each task retains its focused test command in its PLAN and observed evidence in its SUMMARY. The repository suite is the recurring CI gate and executes these Go tests; the focused commands remain available from the plan artifacts.

| Task ID | Requirement | Behavioral evidence | Named tests in plan | CI command | Evidence |
|---|---|---|---|---|---|
| 16-07-01 | NAT-08 | Retain, never delete, TestN1ConvergenceDifferential. Expand each admitted row to applicable source/native modes and c... | TestN1ConvergenceDifferential, TestProgramBranchValidationOrder, TestExistingEmittersAreByteIdentical | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-07-02 | NAT-08 | Add a typed fixed ledger adjacent to the existing map and a pure problem-returning validator used by the production t... | TestPreviousPhaseGoldenCUnchanged, TestPhase16GoldenChangeLedger, TestPhase16GoldenChangeLedgerRejectsFaults | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-07-03 | NAT-08 | Add a narrow internal test-facing direct-program C seam that invokes exactly emitProgram(program, true); it must not ... | TestPhase16DirectProgramFourTierDifferential, TestPhase16EmitterPortSemanticGuardIsNotInert, TestLTOInertnessOnMultiFunctionEmission | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-08-01 | NAT-08, NAT-09 | See task purpose and assertions in plan. | TestN1ConvergenceDifferential, TestPhase16GoldenChangeLedger, TestPhase16EmitterPortSemanticGuardIsNotInert, TestPhase16EmitterCutsAreAmendedAndOwned | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-09-01 | NAT-08, NAT-09 | Recognize commit 0607486 as the completed, indivisible D-01 cutover: public Emit and EmitNative dispatch only to emit... | TestN1, TestProgram, TestPublic, TestPreviousPhaseCoreBytesUnchanged | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-10-01 | NAT-08, NAT-09 | Audit and repair the schema-2 projection path introduced by 22546ae, 5fec955, 43394db, b485295, and b512519. Derive i... | TestPhase5Corpus, TestPhase11Differential, TestPhase16Schema2Projection | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-10-02 | NAT-08, NAT-09 | Make the registry bijective with every direct Emit or EmitNative call under internal/compiler. Resolve import paths a... | TestPhase16, TestWitnessRegistry | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-11-01 | NAT-08, NAT-09 | For foreign_retained, native_lto, and symbols controls that call EmitNative on cut fixtures, replace live legacy-emit... | See PLAN task automated check | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-11-02 | NAT-08, NAT-09 | Classify residual cgen/native public-emitter consumers using the Plan 16-10 registry. For every cut-M004 foreign or b... | See PLAN task automated check | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-12-01 | NAT-08, NAT-09 | Inventory every Phase 5 public emitter route, including inline-LTO, alias, mismatch, sanitizer, and corpus controls. ... | TestPhase5 | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-12-02 | NAT-08, NAT-09 | Audit Phase 4 compatibility and Phase 6 injector/session controls that directly or indirectly invoke public Emit or E... | See PLAN task automated check | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-13-01 | NAT-08, NAT-09 | Audit residual core, Phase 7, quality, admission-divergence, and payload-replay consumers against the Plan 16-10 regi... | See PLAN task automated check | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-13-02 | NAT-08, NAT-09 | Replace the draft validation template with dated command/result records for the focused cgen/core, native, session, a... | See PLAN task automated check | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-14-01 | NAT-08, NAT-09 | Start with behavioral tests for one admitted program and one known cut program, then reduce Phase16ControlNativeC to ... | TestPhase16ControlNativeCAdmittedUsesPublicEmitter, TestPhase16ControlNativeCPreservesM004Refusal | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-14-02 | NAT-08, NAT-09 | Create a _test.go helper in the external session_test package for the historical controls that genuinely need executa... | TestPhase16FileFrozenEvidenceBindsCanonicalProgram, TestPhase16FileFrozenEvidenceRejectsProgramSubstitution, TestPhase16M004CorpusRefusal, TestNonlocalExitProbeInterpreterNative | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-14-03 | NAT-08, NAT-09 | Add one source-derived non-bypass gate over non-_test.go files in internal/compiler/session plus behavioral coverage ... | TestPhase16ProductionPathsPreserveM004Refusal, TestPhase16ProductionSourcesCannotLoadFrozenC, TestPhase16ProductionBypassMutationIsKilled, TestPhase16M004ProvenanceRegistryRejectsFaults | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-15-01 | NAT-08, NAT-09 | Extract the current sorted AST call set after Plan 16-14 and reconcile public-emitter-consumers.json to it. The verif... | TestPhase16PublicEmitterConsumerInventory | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-15-02 | NAT-08, NAT-09 | Replace the current source-literal-only mutation check with table-driven in-memory registry mutations against the sam... | TestPhase16PublicEmitterConsumerInventory, TestPhase16EmitterInventoryMutationControls | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-16-01 | NAT-08, NAT-09 | Add the N=2 checked source fixture matching the existing historical artifact. Verify both existing N=1/N=2 C files ag... | TestPhase16Phase11FrozenEvidenceBindsCanonicalProgram, TestPhase16Phase11FrozenEvidenceRejectsProvenanceFaults | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-17-01 | NAT-08, NAT-09 | Use the existing scripts/evidence-run-record.sh producer with the package-pattern pairs derived by evidence_grade_tes... | TestValidationRowGradesAreEarnedOverArchivedCorpus, TestRunRecordCarriesACompletionWitness, TestRunRecordCompletenessGuardIsNotInert | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-17-02 | NAT-08, NAT-09 | Keep the consumer's exact digest and completion-witness requirements intact and ensure mutation controls reject a cha... | TestValidationRowGradesAreEarnedOverArchivedCorpus, TestRunRecordCompletenessGuardIsNotInert | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-18-01 | NAT-08, NAT-09 | Resolve the unparseable Phase 16 research command into a scanner-supported explicit executable command, and annotate ... | TestVerificationGroundedness, TestVerificationGroundednessFrontierIsPinned, TestVerificationGroundednessThreeClassesAreEmpty, TestVerificationGroundednessGrepExecution | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-18-02 | NAT-08, NAT-09 | Recompute pinnedFrontier and r2bLandingPhases from the current scanner against the full enforced-document set after T... | TestVerificationGroundednessFrontierIsPinned, TestVerificationGroundednessThreeClassesAreEmpty, TestVerificationGroundednessIsNotInert, TestDebtRegisterOwnershipGuardIsNotInert | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-19-01 | NAT-08, NAT-09 | Update only the obsolete witness references in D-13-02b and D-13-10a to the current Phase 17 B1 witness TestPhase17B1... | TestDebtRegistersAreWellFormed, TestUnreachableClaimsViewIsCurrent, TestNoSuppressionOutlivesItsWitness | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-19-02 | NAT-08, NAT-09 | Replace liveMachineIDForTest coupling in budget logic tests with deterministic MachineFacts supplied by a narrowly sc... | TestMachineProbeIsBounded, TestBudgetLaneCarriesMachineIDAndVerdict, TestBudgetAuditRefusesUndeclaredMachine, TestQLT02InterproceduralGrowthExponent | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-20-01 | NAT-08, NAT-09 | Update the current corpus snapshot date and its stated .lang program and total line counts to the independently deriv... | TestLanguageMaturityCountsAreCurrent | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-21-01 | NAT-08, NAT-09 | Move the mid-phase gate helper and its independent structural derivation out of production code and into the session ... | TestPhase11ZeroAttributeGate, TestPhase11GateIsNonVacuous, TestPhase11GateFailsAtNZero, TestPhase11GateCountsAdjacentWouldCarryFunctions | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-21-02 | NAT-08, NAT-09 | Refresh the source-derived inventory after Task 1 removes the production EmitNative calls. Add an evidence-fixture li... | TestPhase16PublicEmitterConsumerInventory, TestPhase16EmitterInventoryMutationControls, TestPhase16EmitterInventoryRefusalWitnessesResolve, TestPhase16M004ProvenanceRegistryRejectsFaults | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-22-01 | NAT-08, NAT-09 | Move the evidence-only N=2 source bytes to testdata/phase16/historical/phase11_gate_n_two.fixture without changing th... | TestPhase11ZeroAttributeGate, TestPhase11GateIsNonVacuous, TestPhase11GateFailsAtNZero, TestPhase11GateCountsAdjacentWouldCarryFunctions | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-23-01 | NAT-08, NAT-09 | Re-run the existing scripts/evidence-run-record.sh producer against package-pattern pairs freshly derived by evidence... | TestValidationRowGradesAreEarnedOverArchivedCorpus, TestRunRecordCarriesACompletionWitness, TestRunRecordCompletenessGuardIsNotInert | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-23-02 | NAT-08, NAT-09 | Update the exact per-file map and total in TestLaneSchemaLiteralSiteCountIsPinned to reflect that Plan 16-21 removed ... | TestLaneSchemaLiteralSiteCountIsPinned | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-24-01 | NAT-08, NAT-09 | Update allPrimaryValidationRows to read validation document status and primary table headers. Always include validate... | TestAllPrimaryValidationRowsExcludeDrafts, TestValidationRowGradesAreEarnedOverArchivedCorpus | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-24-02 | NAT-08, NAT-09 | Update Phase 4 rows 04-03-01 and 04-07-02 to describe the current post-M004 refusal/unreachability boundaries and rep... | TestForeignEvidenceIsRefusedAfterM004Cut | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-24-03 | NAT-08, NAT-09 | Export requested package-pattern pairs with AI_LANG_EVIDENCE_PAIR_OUTPUT=/tmp/ai-lang-validation-pairs.json GOCACHE=/... | TestValidationRowGradesAreEarnedOverArchivedCorpus, TestRunRecordCarriesACompletionWitness, TestRunRecordCompletenessGuardIsNotInert, TestLaneSchemaLiteralSiteCountIsPinned | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-25-01 | NAT-09 | Add a planned M004 milestone/Phase 21 entry titled Native Emission Ownership and Resource Discharge, explicitly after... | TestDebtRegistersAreWellFormed | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-25-02 | NAT-09 | Extend phase16EmitterCutProblems to accept the roadmap as an input, require its M004 Phase 21 title and post-M003 pla... | TestPhase16EmitterCutsAreAmendedAndOwned, TestDebtRegistersAreWellFormed | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-26-01 | NAT-08, NAT-09 | In the Behavioral Spot-Checks table, replace only the NAT-09 owner-law command's grouped -run pattern with '^TestPhas... | TestPhase16EmitterCutsAreAmendedAndOwned, TestDebtRegistersAreWellFormed | `go test ./...` | In recurring CI suite; plan focused check recorded in PLAN/SUMMARY |
| 16-26-02 | NAT-08, NAT-09 | Run the groundedness controls against the corrected report and rerun the complete Go suite. | TestVerificationGroundednessFrontierIsPinned, TestVerificationGroundednessThreeClassesAreEmpty | `go test ./...` | ✅ green; see Validation Audit 2026-09-24 |

Plans 17–21 and 23–26 explicitly declare `gap_ids`; their tests remain in the normal Go suite. Plan 26 closes G-16-21-E. Plans 16-16 and 16-22 also have automated task commands and are covered by the recurring Go suite. The Linux exact-shape `restrict` probe remains host-dependent and is documented above.

## Validation Audit — 2026-09-24

**Outcome: COMPLIANT.** Requirement-to-task coverage is indexed for all plans 16-01 through 16-26. The later gap-closure plans' automated checks are in the recurring Go test suite; plan 16-14-03 has production-path refusal and seeded bypass tests.

| Evidence | Command | Observed result |
|---|---|---|
| Superseded emitter definitions are absent | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run '^TestSupersededEmitterDefinitionsRemoved$' -count=1` | PASS (0.221s) |
| Production bypass mutation is detected | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestPhase16ProductionBypassMutationIsKilled$' -count=1` | PASS (0.364s) |
| Foreign and by-pointer production paths preserve refusal | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestPhase16ProductionPathsPreserveM004Refusal$' -count=1` | PASS (0.895s) |
| M004 ownership and debt shape are valid | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestPhase16EmitterCutsAreAmendedAndOwned$' -count=1`; `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1` | Both PASS (0.754s and 0.849s) |
| Groundedness frontier is pinned | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestVerificationGroundednessFrontierIsPinned$' -count=1` | PASS (0.482s) |
| Groundedness classes are reconciled | `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1` | PASS (0.664s) |
| Full build and test suite after report correction | `GOCACHE=/tmp/ai-lang-gocache go build ./...`; `GOCACHE=/tmp/ai-lang-gocache go test ./...` | Both PASS; fresh merged-branch results supplied by orchestrator and recorded in `16-VERIFICATION.md` |

The scanner R3 finding recorded in the prior audit is resolved: the report now uses `TestSupersededEmitterDefinitionsRemoved`, and both groundedness controls pass. The Linux exact-shape `restrict` lane remains the documented manual-only admission condition because it requires Linux host evidence; the M004 cut disposition remains refusal-first. It does not require conversational UAT.

## Validation Audit — 2026-09-26

**Outcome: COMPLIANT.** The existing map covers all 26 plans and the documented Linux-only probe remains an intentional manual-only admission condition. The execute-phase regression gate passed with `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...`.

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
