---
phase: "33"
slug: "carried-debt-bounded-capture-read-and-generic-freeze-gate"
# status lifecycle: planned (seeded by plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (set by validate-phase §6)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-13"
---

# Phase 33 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Source: `33-RESEARCH.md` § Validation Architecture. All five plans are complete. A fresh focused validation run covers six packages and 34 named test selectors; no hosted run is claimed. Every command is a focused
> `go test -run` or a static check; full suites run on hosted CI only (STATE.md).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing`, go1.24.0 |
| **Config file** | none (go.mod only) |
| **Quick run command** | the focused per-requirement commands below for each touched package |
| **Full suite command** | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest); never run locally |
| **Estimated runtime** | ~30 seconds for all focused rows |

---

## Sampling Rate

- **After every task commit:** the focused command(s) for the touched package, plus `TestPhase16PublicEmitterConsumerInventory` whenever `session.go` or an emitter consumer changed, plus the session meta-test set when a planning document with a `go test` row changed.
- **After every plan wave:** every row below for the wave's packages, plus the session meta-test set.
- **Before `/gsd-verify-work`:** every row below green as focused commands. The phase binds no hosted receipt (D-03).
- **Max feedback latency:** 60 seconds

---

## Per-Requirement Verification Map

| Requirement | Behavior | Test Type | Automated Command | File Exists | Status |
|-------------|----------|-----------|-------------------|-------------|--------|
| QLT-13 | exactly-limit completes; limit+1 is the named capacity outcome; sparse 1 GiB capture allocates < 16*limit | integration | `go test ./internal/compiler/native -run 'TestPhase33CaptureReadBounded' -count=1` | ✅ | ✅ green |
| QLT-13 | existing capture controls unchanged | integration | `go test ./internal/compiler/native -run 'TestPhase22EvidenceMissingPartialAndCapacityControls|TestPhase27OutputEvidenceExactLimit|TestPhase27EvidenceCapacity|TestPhase26FailureChannels' -count=1` | ✅ | ✅ green |
| QLT-14 | freeze-gate verdict identical from repo root and from a subdirectory (WR-01) | integration | `go test ./internal/compiler/session -run 'TestFreezeGate' -count=1` | ✅ | ✅ green |
| QLT-14 | abort-like parity rows pinned before narrowing (WR-02) | integration | `go test ./internal/compiler/native -run 'TestPhase321DefectParityWithLiveOwner|TestPhase33DefectParity' -count=1` | ✅ | ✅ green |
| QLT-14 | predicate accepts only the named reason with the live owner; widening mutant killed (WR-02) | unit | `go test ./internal/compiler/interp -run 'TestPhase321AbortLikeDefect|TestPhase33AbortLike' -count=1` | ✅ | ✅ green |
| QLT-14 | emitted C unchanged by the reason-constant refactor | integration | `go test ./internal/compiler/cgen ./internal/compiler/session -run 'TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestPhase28PrunedEmitMatchesGolden|TestEmittedCExitLiteralsMatchTable' -count=1` | ✅ | ✅ green |
| QLT-14 | IN-01 route audit; IN-02 marker fixture | unit | `go test ./cmd/schway -run 'TestPhase32PublicRouteAudit|TestPhase321RunInputRequiredCLI' -count=1` | ✅ | ✅ green |
| QLT-14 | IN-03 `.md`-only change moves neither digest nor freeze verdict | integration | `go test ./internal/compiler/session -run 'TestPhase32CodeDigest|TestFreezeGate' -count=1` | ✅ | ✅ green |
| QLT-14 | coverage `kind` outside the enum is rejected | unit | `go test ./internal/compiler/session -run 'TestSummaryCoverageGroundedness' -count=1` | ✅ | ✅ green |
| QLT-15 | Phase 29 WR-02/WR-03 witnesses green | integration | `go test ./internal/compiler/check ./internal/compiler/native -run 'TestPhase321ByteViewEntryRefused|TestPhase321DefectParityWithLiveOwner' -count=1` | ✅ | ✅ green |
| QLT-15 | Phase 28 VALIDATION/SECURITY accepted; planning-doc meta-tests pass | meta | `go test ./internal/compiler/session -run 'TestPhase20ValidationLifecycle|TestVerificationGroundedness|TestValidationRowGradesAreEarnedOverArchivedCorpus|TestStaticTestIndexMatchesGoTestList' -count=1` | ✅ | ✅ green |
| EVD-16 | generic gates preserve the catalog through literal goldens (migration step 2) | integration | `go test ./internal/compiler/session -run 'TestReceiptGateDifferential|TestFreezeGateDifferential' -count=1` | ✅ | ✅ green |
| EVD-16 | archived M006 receipts accepted; armed controls reject | integration | `go test ./internal/compiler/session -run 'TestReceiptGateArchivedM006|TestReceiptGate' -count=1` | ✅ | ✅ green |
| EVD-16 | malformed control/program lists cannot waive required names or case-program membership | integration | `go test ./internal/compiler/session -run 'TestReceiptGateMalformedControls|TestReceiptGateProgramInventory' -count=1` | ✅ | ✅ green |
| EVD-16 | manifest sha256 pinned and rows grounded in archived logs | unit | `go test ./internal/compiler/session -run 'TestPhase32ProducerCaseTable' -count=1` | ✅ | ✅ green |
| EVD-16 | no `check-phase32-*` caller; both scripts absent | static | `go test ./internal/compiler/session -run 'TestNoCheckPhase32Caller' -count=1` | ✅ | ✅ green |
| EVD-16 | living-claim marker check generalized by label | unit | `go test ./internal/compiler/session -run 'TestPhase32ReceiptLivingClaims' -count=1` | ✅ | ✅ green |
| (guard) | emitter inventory, CI wiring, living-roadmap word check | meta | `go test ./internal/compiler/session ./internal/compiler/native -run 'TestPhase16PublicEmitterConsumerInventory|TestPhase32CIHostMatrix|TestPhase32HostedCIEnv|TestPhase25LivingRoadmap' -count=1` | ✅ | ✅ green |

*Status: ✅ green · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [x] `internal/compiler/native/native_phase33_capture_test.go` — QLT-13
- [x] `internal/compiler/interp/interp_phase33_defect_test.go` — WR-02 narrowing and mutant
- [x] file-route parity rows (new `native_phase33_defect_parity_test.go` or extension) — WR-02 pin
- [x] `testdata/receipts/M006.manifest`, `scripts/check-freeze.sh`, `scripts/check-receipts.sh` — EVD-16
- [x] differential tests and `TestReceiptGateArchivedM006` — EVD-16
- [x] `TestNoCheckPhase32Caller` — EVD-16

No framework install needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| seeded unbounded-read mutant is caught | QLT-13 | D-16: one-time mutation was executed and failed; preserved in `33-01-SUMMARY.md` | completed during plan 33-01; no human verification remains |

Static checks: `git diff --stat -- go.mod go.sum` empty; `sh -n` on the gate scripts; `git grep -n 'check-phase32' -- . ':!.planning'` empty after the migration.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** verified 2026-10-13

## Per-Task Completion Map

| Plan tasks | Requirements | Evidence | Status |
|------------|--------------|----------|--------|
| 33-01 tasks 1–3 | QLT-13, QLT-14 | `33-01-SUMMARY.md` and capture/marker/coverage controls above | green |
| 33-02 tasks 1–3 | QLT-14 | `33-02-SUMMARY.md` and parity/narrowing controls above | green |
| 33-03 tasks 1–3 | EVD-16, QLT-14 | `33-03-SUMMARY.md` and generic/archived gate controls above | green |
| 33-04 tasks 1–3 | EVD-16 | `33-04-SUMMARY.md` and goldens/no-caller/producer controls above | green |
| 33-05 tasks 1–3 | EVD-16, QLT-14, QLT-15 | `33-05-SUMMARY.md` and living-claim/groundedness controls above | green |

## Validation Audit 2026-10-13

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

## Successor Refresh — 2026-10-13

The existing compliant contract was re-audited after Phase 34 changed covered shared inputs. All 18 focused rows resolved to 47 current named tests; the exact anchored run passed across six packages with no missing selector (26.325s). No new coverage gap or test generation was needed. See `33-REFRESH-AUDIT.md`; the earlier receipts above remain historical.

## Phase 35 successor refresh — 2026-10-14

All original plans remain complete; no implementation or UAT is replayed. The
Phase 35 final inherited regression refresh passed 94 distinct names / 98
package-qualified executions across eleven packages with no missing/fail/skip,
including this phase's mapped selectors. See `../35-bitwise-operators-literal-shifts-and-hex-literals/35-REGRESSION.md`.
This refresh labels the original execution receipts historical and recomputes
the current verification fingerprint through its canonical owner.

State A coverage review found no additional unmapped requirement or test gap. Successor-specific syntax and inventory adaptations are explicitly bound in the Phase 35 compatibility/mutation evidence; their frozen before-state fixtures are preserved.


## Phase 37 successor validation refresh — 2026-10-15

State A was re-audited after the Phase 37 successor edits. The first complete
18-row focused run exposed inherited inventory, finite-source allowance and
evidence-metadata gaps. Those gaps are corrected without changing production
or frozen receipts. A fresh normalized, timeout-bounded rerun passed all
47 distinct package-qualified root executions across six packages, with zero
missing, failed or skipped tests (48.447 seconds wall time). Five supplemental
ledger/registry roots also passed, including the new 21-task/C01-C28 registry
guard and its 23 reached refusal controls. All four requirement mappings remain
covered; no human-only verification item or test gap remains.

The exact commands, root sets, log hashes and repaired worktree hashes are in
[33-PHASE37-REFRESH-RUN.json](33-PHASE37-REFRESH-RUN.json). Earlier receipts
remain historical. These are parent-run focused checks on Darwin/arm64,
not a new hosted or full-suite result. The independent goal verdict is pending.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 4 |
| Resolved | 4 |
| Escalated | 0 |

## Review-control closure — 2026-10-15

Independent review found two additional fail-open cases in existing guards.
WR-01 now checks each coverage verification item rather than aggregate field
slices; all 29 controls and 60 summaries pass. WR-02 now enforces literal
case/program membership in both directions while preserving prior refusal
positions; the five-root receipt compatibility run passes. Both controls
failed for their intended reasons before the fix. The EVD-16 row now names
TestReceiptGateProgramInventory explicitly. The corpus grade law and frozen
receipt are unchanged. See 33-REVIEW-FIX.md for exact failures and logs;
the final parent rerun and independent goal verdict remain separate gates.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 6 |
| Resolved | 6 |
| Escalated | 0 |
