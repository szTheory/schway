---
phase: "16"
slug: "branch-match-emitter-port"
status: complete
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
