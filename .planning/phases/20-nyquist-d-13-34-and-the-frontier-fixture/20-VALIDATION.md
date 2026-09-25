---
phase: "20"
slug: "nyquist-d-13-34-and-the-frontier-fixture"
status: planned
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-25"
evidence_vocabulary: v1
graded_rows: 25
---

# Phase 20 — Validation Strategy

This is the pre-execution validation contract. `planned` means no Phase 20 result is claimed yet. The Phase 14 groundedness lint must measure the starting frontier before archived records are reconciled. The 20-06 decision checkpoint is a blocking human choice and has no automated result row; its preparation and selected outcome do.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 standard `testing` and the existing project CLI |
| Config | `go.mod`; no new external dependency |
| Interim corpus lane | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1`; the exact frontier pin is required only after Plan 07 reconciles all Phase 20 findings. |
| Full suite timing | Three paired cold/warm `go test ./... -count=1` runs under one 30-minute helper cap; report raw samples, min/median/max, paired deltas, the roadmap's 192.7 s reference, and Phase 14's measured 191.89 s cold manifest value. |
| Host cache | Use `GOCACHE=/private/tmp/phase20-gocache` in this Codex sandbox; this is environment handling, not a repository setting. |

## Sampling Rate

- Before any VALIDATION edit, preserve two distinct measurements: the historical research snapshot (R1/R2/R3 = 0, reconciled R2b = 23, raw R2b = 24, 639 documents, 708 commands) and a new scanner run immediately before Phase 20 execution. The 2026-09-25 planning scan was R2=10, R2b=25, 651 documents, 737 commands; remeasure at execution start rather than treating either snapshot as a required live count.
- After each documentation task, run the named current test/CLI commands and the non-vacuous corpus check. Capture the full scanner's exact findings even while prospective Phase 20 tests keep its global pin red. Plan 07 reconciles identities and R2b owners and then requires the global frontier test green.
- In Wave 1, run native/cache and enumerated-closure controls plus the timing harness's deterministic smoke tests. Reconcile maturity counts, public-emitter consumers, skip witnesses, debt, and the exact frontier in Plan 07. Plan 08 requires a green unfiltered suite preflight before gathering three controlled cold/warm **full repository suite** pairs and comparing their distributions with the Phase 14 references.
- At phase close, run the checksum, validation lifecycle, closure cache, debt cap, and held-out corpus controls plus `go test ./... -count=1`. A status transition to `validated` or `complete` requires actual run evidence, not document edits alone.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Behavior | Automated Command | File Exists | Grade | Non-inertness | Status |
|---|---|---|---|---|---|---|---|---|
| 20-01-01 | 01 | 1 | QLT-11 | Check in refused checksum-intent fixture and pin current code/span. | `go test ./internal/compiler/session -run '^TestPhase20ChecksumFrontier$' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-01-02 | 01 | 1 | QLT-11 | Compare identical source bytes against the reproduced M003-open diagnostic. | `go test ./internal/compiler/session -run '^TestPhase20ChecksumFrontier$' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-02-01 | 02 | 1 | QLT-10 | Measure the fresh execution-start frontier, then reconcile Phase 07's commands. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-02-02 | 02 | 1 | QLT-10 | Retire Phase 08's zero-test and deleted-mechanism cells. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-02-03 | 02 | 1 | QLT-10 | Replace Phase 11's elided command cells, recording exact findings for final pin. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-03-01 | 03 | 2 | QLT-10 | Adjudicate active-milestone validation statuses. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-03-02 | 03 | 2 | QLT-10 | Adjudicate remaining archived statuses and stale row evidence. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-03-03 | 03 | 2 | QLT-10 | Enforce zero draft and capture interim exact frontier findings and owners. | `go test ./internal/compiler/session -run '^TestPhase20ValidationLifecycle$' -count=1` | ❌ lifecycle W2 | EXERCISED | — | ⬜ pending |
| 20-04-01 | 04 | 1 | QLT-12 | Reuse only content-bound native build artifacts and execute outputs fresh. | `go test ./internal/compiler/native ./internal/compiler/cache -run 'TestPhase20|TestCache|TestNative' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-04-02 | 04 | 1 | QLT-12 | Reuse unchanged closure evidence and invalidate seeded changed inputs. | `go test ./internal/compiler/session -run '^TestPhase20EnumeratedClosure' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-04-03 | 04 | 1 | QLT-12 | Implement and smoke-check the bounded timing harness without launching the full suite. | `go test ./scripts -run '^TestPhase20ClosureTimingHarness$' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-05-01 | 05 | 1 | PRC-02 | Derive current debt population from the M002 ten-ID cohort plus live M003 rows. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1` | ❌ population W1 | EXERCISED | — | ⬜ pending |
| 20-05-02 | 05 | 1 | PRC-02 | Reconcile the three Phase 16 emitter-debt closures. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1` | ❌ population W1 | EXERCISED | — | ⬜ pending |
| 20-05-03 | 05 | 1 | PRC-02 | Assign the verified Phase 21 native/LTO owner and prove seeded cap rule. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtCapRule$' -count=1` | ❌ W1 | EXERCISED | — | ⬜ pending |
| 20-09-01 | 09 | 3 | QLT-10, PRC-02 | Repair four archived validation rows and their live grade probes. | `go test ./internal/compiler/session -run 'TestValidationGradeCapBarePackageRowHasNoNamedTest|TestValidationGradeCapArchivedDeadCitationsRemainAbsent' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-09-02 | 09 | 3 | QLT-10 | Regenerate exact corpus record from exported package-pattern pairs. | `go test ./internal/compiler/session -run '^TestValidationGradeCapBarePackageRowHasNoNamedTest$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-09-03 | 09 | 3 | PRC-02 | Close only evidenced archive debt and refresh generated views. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-07-02 | 07 | 4 | QLT-10, PRC-02 | Refresh maturity counts, classify final emitter calls, and cite all six new skip witnesses. | `go test ./internal/compiler/session -run '^TestLanguageMaturityCountsAreCurrent$' -count=1 && go test ./internal/compiler/session -run '^TestSelfDescribingDocsGuardIsNotInert$' -count=1 && go test ./internal/compiler/session -run '^TestPhase16PublicEmitterConsumerInventory$' -count=1 && go test ./internal/compiler/session -run '^TestNoSuppressionOutlivesItsWitness$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-07-03 | 07 | 4 | QLT-10, PRC-02 | Close proved row debt and enforce a live five-item cap, seeded overflow, and current views. | `go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtCapRule$' -count=1 && go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run 'TestUnreachableClaimsView' -count=1 && go test ./internal/compiler/session -run '^TestReconciliationVerdictsCarryTheirObligations$' -count=1 && go test ./internal/compiler/session -run 'TestVerificationGroundedness' -count=1` | ❌ cap W3 | EXERCISED | — | ⬜ pending |
| 20-10-01 | 10 | 5 | QLT-10 | Record the exact corpus after phase-wide row status and evidence repair. | `go test ./internal/compiler/session -run '^TestValidationGradeCapBarePackageRowHasNoNamedTest$' -count=1` | ✅ Plan 20 | EXERCISED | — | ✅ passed |
| 20-10-02 | 10 | 5 | QLT-10, PRC-02 | Close only evidenced row debt and regenerate the derived view. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1` | ✅ existing | EXERCISED | — | ✅ passed |
| 20-08-01 | 08 | 6 | QLT-12 | Require a green unfiltered full suite before timing starts. | `go test ./... -count=1` | ✅ Plan 20 | WIRED | — | ✅ passed |
| 20-08-02 | 08 | 6 | QLT-12 | Measure three paired cold/warm full-suite distributions on the green reconciled tree. | `go run scripts/phase20-closure-timing.go` | ✅ Plan 20 | REACHABLE | — | ✅ passed |
| 20-06-01 | 06 | 7 | PRC-02 | Prepare measured D-13-34 alternatives for human choice. | `go test ./internal/compiler/session -run '^TestPhase6HeldoutPairsAreAlphaRenamesOnly$' -count=1 && go test ./internal/compiler/session -run '^TestPhase6DefectCorpusIsHeldOut$' -count=1` | ✅ existing | EXERCISED | — | ⬜ pending |
| 20-06-03 | 06 | 7 | PRC-02 | Implement selected outcome and assert the live ≤5 debt cap. | `go test ./internal/compiler/session -run 'TestPhase6HeldoutPairs' -count=1 && go test ./internal/compiler/session -run '^TestPhase6DefectCorpusIsHeldOut$' -count=1 && go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run 'TestPhase20UnownedDebt' -count=1` | ❌ cap W5 | EXERCISED | — | ⬜ pending |

## Wave 0 Requirements

- `20-CHECKSUM-BASELINE.md` records the exact proposed bytes and independently reproduced old/current diagnostics; there was no original M003-open fixture pin.
- Add the Phase 20 checksum, validation lifecycle, closure cache, and debt population/cap tests in the owning tasks before citing them as run evidence.
- The D-13-34 decision record must show both concrete outcomes before the blocking human checkpoint. Neither option is selected by this document.
- Content-addressed cache hits require a complete declared native build input manifest; fail closed or bypass cache when compiler, SDK/sysroot, foreign-source/header, or linker inputs cannot be accounted for.

## Manual Decision

20-06 Task 2 is `gate="blocking-human"`: choose replacement held-out move/borrow programs or a reasoned re-ratification with an owning milestone. Phase verification records the selected outcome and D-06-29's restored or written-off inference.

## Validation Sign-Off

- [ ] All 21 automated task rows resolve to real commands and have recorded results.
- [ ] Zero draft VALIDATION files remain; current Phase 20 status is updated from actual evidence.
- [ ] Maturity counts, public-emitter registry, skip witnesses, reconciliation view, and exact frontier gates are green before the timing preflight.
- [ ] Plan 08's unfiltered full-suite preflight passes before any timed pair begins.
- [ ] Three cold and three warm full-suite runs pass within the helper cap; raw samples, min/median/max, and paired deltas are recorded against both Phase 14 references.
- [ ] The live, source-derived open-unowned debt set contains no more than five IDs.
- [ ] `nyquist_compliant: true` is set only after all applicable gates pass.
