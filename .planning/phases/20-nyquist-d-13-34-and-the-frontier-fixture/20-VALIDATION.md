---
phase: "20"
slug: "nyquist-d-13-34-and-the-frontier-fixture"
status: planned
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-25"
evidence_vocabulary: v1
graded_rows: 18
---

# Phase 20 — Validation Strategy

This is the pre-execution validation contract. `planned` means no Phase 20 result is claimed yet. The Phase 14 groundedness lint must measure the starting frontier before archived records are reconciled. The 20-06 decision checkpoint is a blocking human choice and has no automated result row; its preparation and selected outcome do.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 standard `testing` and the existing project CLI |
| Config | `go.mod`; no new external dependency |
| Focused lane | `go test ./internal/compiler/session -run 'TestVerificationGroundedness' -count=1` |
| Full suite timing | `go test ./... -count=1`; report both the roadmap's 192.7 s reference and Phase 14's measured 191.89 s cold manifest value. |
| Host cache | Use `GOCACHE=/private/tmp/phase20-gocache` in this Codex sandbox; this is environment handling, not a repository setting. |

## Sampling Rate

- Before any VALIDATION edit, record the Phase 14 lint's raw and reconciled counts separately. Baseline: R1/R2/R3 = 0, reconciled R2b = 23, raw R2b = 24, 639 documents, 708 commands.
- After each documentation task, run the named current test/CLI commands and the focused scanner. Where an exact frontier pin moves, update only measured records and ownership.
- After the closure-cache task, run native/cache controls and the enumerated closure cold and warm. The phase gate runs the **full repository suite**, with cold and warm wall-clock evidence against Phase 14's 192.7 s baseline.
- At phase close, run the checksum, validation lifecycle, closure cache, debt cap, and held-out corpus controls plus `go test ./... -count=1`. A status transition to `validated` or `complete` requires actual run evidence, not document edits alone.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Behavior | Automated Command | File Exists | Grade | Status |
|---|---|---|---|---|---|---|---|---|
| 20-01-01 | 01 | 1 | QLT-11 | Check in refused checksum-intent fixture and pin current code/span. | `go test ./internal/compiler/session -run '^TestPhase20ChecksumFrontier$' -count=1` | ❌ W1 | DEFINED | ⬜ pending |
| 20-01-02 | 01 | 1 | QLT-11 | Compare identical source bytes against the reproduced M003-open diagnostic. | `go test ./internal/compiler/session -run '^TestPhase20ChecksumFrontier$' -count=1` | ❌ W1 | DEFINED | ⬜ pending |
| 20-02-01 | 02 | 1 | QLT-10 | Reconcile Phase 07's commands from live lint and test names. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessFrontierIsPinned$' -count=1 && go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-02-02 | 02 | 1 | QLT-10 | Retire Phase 08's zero-test and deleted-mechanism cells. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessFrontierIsPinned$' -count=1 && go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-02-03 | 02 | 1 | QLT-10 | Replace Phase 11's elided command cells and re-pin exact findings. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessFrontierIsPinned$' -count=1 && go test ./internal/compiler/session -run '^TestVerificationGroundednessThreeClassesAreEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-03-01 | 03 | 2 | QLT-10 | Adjudicate active-milestone validation statuses. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-03-02 | 03 | 2 | QLT-10 | Adjudicate remaining archived statuses and stale row evidence. | `go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-03-03 | 03 | 2 | QLT-10 | Enforce zero draft, close proven stale-row debt, and exact frontier ownership. | `go test ./internal/compiler/session -run 'TestVerificationGroundedness' -count=1 && go test ./internal/compiler/session -run '^TestPhase20ValidationLifecycle$' -count=1` | ❌ lifecycle W2 | DEFINED | ⬜ pending |
| 20-04-01 | 04 | 1 | QLT-12 | Reuse only content-bound native build artifacts and execute outputs fresh. | `go test ./internal/compiler/native ./internal/compiler/cache -count=1` | ❌ W1 | DEFINED | ⬜ pending |
| 20-04-02 | 04 | 1 | QLT-12 | Reuse unchanged closure evidence and invalidate seeded changed inputs. | `go test ./internal/compiler/session -run '^TestPhase20EnumeratedClosure' -count=1` | ❌ W1 | DEFINED | ⬜ pending |
| 20-04-03 | 04 | 1 | QLT-12 | Measure full-suite cold/warm cost against the Phase 14 baseline. | `go run scripts/phase20-closure-timing.go` | ❌ W1 | DEFINED | ⬜ pending |
| 20-05-01 | 05 | 1 | PRC-02 | Derive current debt population from the M002 ten-ID cohort plus live M003 rows. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1` | ❌ population W1 | DEFINED | ⬜ pending |
| 20-05-02 | 05 | 1 | PRC-02 | Reconcile the three Phase 16 emitter-debt closures. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1` | ❌ population W1 | DEFINED | ⬜ pending |
| 20-05-03 | 05 | 1 | PRC-02 | Assign the verified Phase 21 native/LTO owner and prove seeded cap rule. | `go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtCapRule$' -count=1` | ❌ W1 | DEFINED | ⬜ pending |
| 20-07-01 | 07 | 3 | QLT-10, PRC-02 | Repair four stale or underspecified validation rows with current test evidence. | `go test ./internal/compiler/session -run '^TestValidationGradeCapArchivedDeadCitationsRemainAbsent$' -count=1 && go test ./internal/compiler/session -run '^TestValidationGradeCapBarePackageRowHasNoNamedTest$' -count=1 && go test ./internal/compiler/session -run '^TestVerificationGroundednessCorpusIsNotEmpty$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-07-02 | 07 | 3 | QLT-10, PRC-02 | Close proved row debt and enforce a live five-item cap, seeded overflow, and current views. | `go test ./internal/compiler/session -run '^TestPhase20UnownedDebtPopulation$' -count=1 && go test ./internal/compiler/session -run '^TestPhase20UnownedDebtCap$' -count=1 && go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run 'TestUnreachableClaimsView' -count=1 && go test ./internal/compiler/session -run 'TestVerificationGroundedness' -count=1` | ❌ cap W3 | DEFINED | ⬜ pending |
| 20-06-01 | 06 | 4 | PRC-02 | Prepare measured D-13-34 alternatives for human choice. | `go test ./internal/compiler/session -run '^TestPhase6HeldoutPairsAreAlphaRenamesOnly$' -count=1 && go test ./internal/compiler/session -run '^TestPhase6DefectCorpusIsHeldOut$' -count=1` | ✅ existing | DEFINED | ⬜ pending |
| 20-06-03 | 06 | 4 | PRC-02 | Implement selected outcome and assert the live ≤5 debt cap. | `go test ./internal/compiler/session -run 'TestPhase6HeldoutPairs' -count=1 && go test ./internal/compiler/session -run '^TestPhase6DefectCorpusIsHeldOut$' -count=1 && go test ./internal/compiler/session -run '^TestDebtRegistersAreWellFormed$' -count=1 && go test ./internal/compiler/session -run 'TestPhase20UnownedDebt' -count=1` | ❌ cap W4 | DEFINED | ⬜ pending |

## Wave 0 Requirements

- `20-CHECKSUM-BASELINE.md` records the exact proposed bytes and independently reproduced old/current diagnostics; there was no original M003-open fixture pin.
- Add the Phase 20 checksum, validation lifecycle, closure cache, and debt population/cap tests in the owning tasks before citing them as run evidence.
- The D-13-34 decision record must show both concrete outcomes before the blocking human checkpoint. Neither option is selected by this document.
- Content-addressed cache hits require a complete declared native build input manifest; fail closed or bypass cache when compiler, SDK/sysroot, foreign-source/header, or linker inputs cannot be accounted for.

## Manual Decision

20-06 Task 2 is `gate="blocking-human"`: choose replacement held-out move/borrow programs or a reasoned re-ratification with an owning milestone. Phase verification records the selected outcome and D-06-29's restored or written-off inference.

## Validation Sign-Off

- [ ] All 18 automated task rows resolve to real commands and have recorded results.
- [ ] Zero draft VALIDATION files remain; current Phase 20 status is updated from actual evidence.
- [ ] The full suite passes; cold/warm suite and closure timings are recorded against the 192.7 s Phase 14 baseline.
- [ ] The live, source-derived open-unowned debt set contains no more than five IDs.
- [ ] `nyquist_compliant: true` is set only after all applicable gates pass.
