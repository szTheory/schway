---
phase: 05-native-equivalence-and-adversarial-evidence
verified: 2026-09-06T19:21:45Z
status: passed
score: 4/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
---

# Phase 5: Native Equivalence and Adversarial Evidence Verification Report

**Phase Goal:** The source-to-native subset survives optimization and deliberately seeded boundary defects with independently meaningful evidence.
**Verified:** 2026-09-06T19:21:45Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth (ROADMAP SC) | Status | Evidence |
|---|------|--------|----------|
| 1 | Interpreter, `-O0`, and `-O3` match on terminal outcome, semantic-event order, and live-resource state across the milestone corpus | ✓ VERIFIED | `TestPhase5CorpusThreeEngineAgreement` (`internal/compiler/session/session_phase5_corpus_test.go`) runs the union of every Phase 1-5 corpus fixture plus `EnumeratePhase5Closure()`'s bounded enumeration (112 cases observed, including 84+ enumerated closure cases) through interpreter/`-O0`/`-O3` — ran directly in this session, all PASS. The flagged flaky signal (one prior `-race` timeout) did not reproduce; the test passed cleanly here. `control:interpreter-o0-o3-lto` (the `-O3 -flto` tier) also fires with `recomputed_work: 4` in the live `verify-phase5.sh` run (not zero-work), and `TestLTOTierIsNotInert` independently proves `-flto` changes codegen — the LTO tier is non-inert, not vacuous. D-05-41 (LTO tier samples one of six adversarial fixtures inside the gate's own lane, not all six) is recorded debt, not a gap — the full six-fixture set is covered at `-O0`/`-O3` by the exhaustive corpus test above. |
| 2 | All seven native hostile mutations and all applicable ownership/certificate mutations are detected by the intended independent lane | ✓ VERIFIED | `NAT03Mutations()` (`internal/compiler/session/session_phase5_alias.go`) lists all 7 rows; `TestNAT03MutationTableHasSevenRows`, `TestNAT03IsSixSubjectedPlusOneEscape`, `TestNAT03MutationsCiteExistingPrograms`, and `TestEveryMutationMovesItsClaimedAxis` all PASS (run directly). Verified all 6 cited corpus programs exist on disk (`testdata/phase4/foreign_layout_mismatch.golden.c`, `testdata/phase4/acquire_three_success.lang`, `testdata/phase4/nonlocal_exit_probe.lang`, `testdata/phase5/false_restrict_hoist.lang`, `testdata/phase5/allocator_mismatch.lang`, `testdata/phase5/retained_pointer.lang`). Row 7 (stale callback retention) is honestly declared `Subjected: false` with `EscapeID: "escape:callback-invocation-unsubjected"` — never claimed as a detected control (confirmed absent from `Phase5RequiredControls()`/`AllShippedControlIDs()` via `TestBothPhase5EscapesAreVisible`/`TestCoordinatedLieEscapeIsNeverDetected`, both PASS). |
| 3 | ASan/UBSan evidence is isolated from semantic equivalence evidence and catches retained-pointer lifetime defects | ✓ VERIFIED | `TestSanitizerBuildIsSeparateArtifact` and `TestSanitizerReportIsNotNativeResult` PASS (run directly), confirming structural isolation from the equivalence comparator. `TestSanitizeLaneRetainedPointerAlwaysReports` and `TestSanitizeLaneAllocatorMismatchIsDetected` PASS, confirming real detection over `native/lang_foreign_retained.c` (heap-UAF) and `native/lang_foreign_arena.c` (allocator-identity mismatch). Live gate run shows `control:native.sanitize.retained_pointer`, `control:native.sanitize.use_after_free`, `control:native.sanitize.allocator_mismatch`, and `control:native.sanitize.ubsan_no_recover` all firing with nonzero `recomputed_work`. |
| 4 | Any injected mismatch reports a minimized source/core case and causal event trace; the coordinated source-to-core false claim remains a documented escape | ✓ VERIFIED | `MismatchDocument` (`internal/compiler/reduce/mismatch.go`, `lang.mismatch/0`) carries `ReducedCore`, `ReducedSource`, `Minimality`, and `CausalSteps`/event-window fields; `TestMismatchDocumentHasExactlyTheSpecifiedFields`, `TestMismatchEventWindowIsBounded`, `TestNoOpReducerGoesRed`, `TestPredicateTooLooseGoesRed`, and `TestReducerNonDeterminismGoesRed` all PASS (run directly) — the three reducer vacuity mutation-kills prove the reducer is not a rubber stamp. `TestCoordinatedLieArtifactsBothValidate`, `TestCoordinatedLieEscapeIsDeclared`, `TestCoordinatedLieEscapeIsNeverDetected`, and `TestCoordinatedLiePassesTheGateUnderTheNamedEscape` all PASS — the coordinated source-to-core lie (`testdata/phase5/coordinated_lie.{lang,core.json}`) is demonstrated reachable, passes the gate, and is attributed explicitly to `escape:coordinated-source-to-core-false-claim`, never claimed closed. |

**Score:** 4/4 truths verified (0 present, behavior-unverified)

### Required Artifacts (spot-checked, not exhaustive)

| Artifact | Expected | Status |
|----------|----------|--------|
| `internal/compiler/session/session_phase5.go` | `Phase5RequiredControls()` (15 identifiers), `Phase5ExpectedEscapes()` (2), gate orchestration, zero-work sweep | ✓ VERIFIED — read in full; 15 controls confirmed by literal count; `for _, lane := range result.Lanes { if lane.RecomputedWork == 0 { ... fail ... } }` present, proving no lane can pass at zero work |
| `scripts/verify-phase5.sh` | Peer gate script, does not invoke prior gate scripts, pins ASAN/UBSAN options | ✓ VERIFIED — ran live, exit 0, all lanes report nonzero `recomputed_work` |
| `internal/compiler/session/qlt01_registry.json` | Machine-readable control registry from spikes 001-005 | ✓ VERIFIED — 22 rows confirmed by direct JSON parse |
| `internal/compiler/reduce/mismatch.go` | `lang.mismatch/0` document, 13 fields | ✓ VERIFIED — `EvidenceID`, `ReducedCore`, `ReducedSource`, `Minimality` fields present; `TestNoNewSchemaVersionsIntroduced` and `TestMismatchDocumentHasExactlyTheSpecifiedFields` PASS |
| `testdata/phase5/coordinated_lie.lang` + `.core.json` | Adversarial artifact pair | ✓ VERIFIED — both exist on disk, `TestCoordinatedLieArtifactsBothValidate` PASS |

### Behavioral Spot-Checks / Test Execution

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Full Phase 5 gate | `sh scripts/verify-phase5.sh` (run live) | exit 0; `go test ./...` and `go test -race ./...` both green; 5 lanes-groups all `"status":"pass"`; every individual lane's `recomputed_work` > 0 (no vacuous pass observed) | ✓ PASS |
| No-regression: Phase 1 gate | `sh scripts/verify-phase1.sh` (run live) | exit 0, `"status":"pass"` | ✓ PASS |
| No-regression: Phase 2 gate | `sh scripts/verify-phase2.sh` (run live) | exit 0, `"status":"pass"` | ✓ PASS |
| No-regression: Phase 3 gate | `sh scripts/verify-phase3.sh` (run live) | exit 0, `"status":"pass"` | ✓ PASS |
| No-regression: Phase 4 gate | `sh scripts/verify-phase4.sh` (run live) | exit 0, `"status":"pass"` | ✓ PASS |
| Three-engine corpus agreement (SC1) | `go test ./internal/compiler/session/... -run TestPhase5CorpusThreeEngineAgreement -v` | all subtests PASS (112 cases incl. enumerated closure) | ✓ PASS |
| Prior-phase byte-identity | `go test ./internal/compiler/core/... -run Golden` | `TestPreviousPhaseGoldenCUnchanged` PASS across phase1/2/4 goldens | ✓ PASS |
| Byte-frozen prior artifacts | `git diff --stat 4ca704b^..HEAD -- internal/compiler/session/session.go testdata/phase1..4 scripts/verify-phase1..4.sh` | empty diff | ✓ PASS |
| QLT-01 registry mutation-kills | `go test ./internal/compiler/session/... -run TestQLT01 -v` | all 8 subtests PASS, including 3 "goes red" mutation-kill tests | ✓ PASS |
| Reducer mutation-kills (D-05-27) | `go test ./internal/compiler/reduce/... -run "TestNoOpReducerGoesRed|TestPredicateTooLooseGoesRed|TestReducerNonDeterminismGoesRed" -v` | all PASS | ✓ PASS |
| Coordinated-lie escape (D-05-30) | `go test ./internal/compiler/session/... -run TestCoordinatedLie -v` | all 4 subtests PASS | ✓ PASS |
| Sanitizer isolation + detection (SC3) | `go test ./internal/compiler/native/... -run TestSanitiz -v` and session sanitize tests | all PASS | ✓ PASS |
| NAT-03 citation/axis-movement proofs | `go test ./internal/compiler/session/... -run "TestNAT03|TestAssertMutationMovesAnAxis|TestEveryNAT03" -v` | all PASS | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan(s) | Status | Evidence |
|---|---|---|---|
| INT-02 | 05-10, 05-12, 05-13, 05-14 | ✓ SATISFIED | `lang.mismatch/0` document with minimized source/core + causal trace; matches REQUIREMENTS.md's checked entry |
| NAT-02 | 05-02, 05-05, 05-06, 05-07, 05-09, 05-14 | ✓ SATISFIED | Three-engine agreement over milestone corpus, live-tested |
| NAT-03 | 05-01, 05-03, 05-04, 05-07, 05-08, 05-09, 05-14 | ✓ SATISFIED | Seven-row mutation table, six subjected + one honest escape, all corpus programs verified to exist |
| QLT-01 | 05-11, 05-13, 05-14 | ✓ SATISFIED | 22-row registry, fail-closed audit with mutation-kill tests, coordinated lie made an explicit gate-visible escape |

No orphaned requirements found in REQUIREMENTS.md for Phase 5 — all four declared IDs (INT-02, NAT-02, NAT-03, QLT-01) are covered by at least one plan's `requirements` field, and all four are marked `Complete` in REQUIREMENTS.md's own tracking table.

### Anti-Patterns Found

Carried forward from `05-REVIEW.md` (status `issues_found`, 0 critical, 2 warnings, 1 info) — re-confirmed present and unresolved but non-blocking:

| File | Issue | Severity | Impact |
|------|-------|----------|--------|
| `internal/compiler/reduce/reduce.go:937-968` | `removeIndices`/`scrubBlocks` unreachable dead code (WR-01) | ⚠️ Warning | Cosmetic; does not affect any gate or control |
| `internal/compiler/session/qlt01.go:279-294` | `QLT01LaneFromRows` only populates `Controls` on success branch, causing overlapping diagnostics on failure (WR-02) | ⚠️ Warning | Gate still correctly fails; triage-clarity issue only |
| `internal/compiler/session/session_phase5_mismatch.go:135-148` | `mismatchEvidenceID` swallows `json.Marshal` errors silently (IN-01) | ℹ️ Info | Currently unreachable given `core.Program`'s exhaustive JSON tagging |

No debt-marker (`TBD`/`FIXME`/`XXX`) or unresolved `TODO`/`HACK`/`PLACEHOLDER` found in the Phase 5 file set reviewed.

### Known-and-Accepted Items (verified as recorded, not new gaps)

- **D-05-41**: `control:interpreter-o0-o3-lto`'s own gate lane samples one adversarial fixture (`inline_across_foreign.lang`) at the LTO tier, not all six — confirmed present in `05-DEBT.md` with severity `info`, landing phase `Phase 5+`. Confirmed NOT a correctness gap: the full six-fixture set is proven at `-O0`/`-O3` via `TestPhase5CorpusThreeEngineAgreement`, and LTO non-inertness is proven independently via `TestLTOTierIsNotInert`.
- **D-05-42**: `OpCall`/interprocedural equivalence deferred to M002 as its lead charter item; D-03-02 remains open past the milestone. Confirmed recorded verbatim in `05-DEBT.md` and in `ROADMAP.md`'s "M002 Charter" section, per D-05-33's required wording.
- **`lang.mismatch/0`'s thirteen fields** (D-05-26's twelve plus additive `evidence_id`) — confirmed by direct read of `mismatch.go`'s struct tags.
- **Two named escapes never counted as detected controls**: `escape:callback-invocation-unsubjected` and `escape:coordinated-source-to-core-false-claim` — both confirmed absent from `Phase5RequiredControls()`/`AllShippedControlIDs()` by passing tests, and both confirmed visible/declared.

### Deferred Items

None beyond the M001-to-M002 deferrals already recorded as accepted debt (D-05-42) — not treated as gaps per the phase's own explicit scope decision.

### Human Verification Required

None. All four ROADMAP success criteria resolve to concrete, directly-executed test/gate evidence rather than requiring visual, real-time, or subjective judgment.

### Gaps Summary

No gaps found. All 15 required Phase 5 controls fire with nonzero recomputed work in a live run of `scripts/verify-phase5.sh` (exit 0). Both declared escapes remain visible and are never claimed as detected controls. All NAT-03 corpus citations resolve to files that exist on disk, and each mutation is independently proven to move its claimed comparison axis rather than passing vacuously. The `-O3 -flto` tier is proven non-inert by a dedicated codegen-diff test, distinct from the sampling-scope debt item (D-05-41) which is honestly recorded rather than hidden. Prior-phase gates (`verify-phase1.sh` through `verify-phase4.sh`) all remain green with no regression, and `session.go` plus every frozen prior-phase testdata file and gate script are confirmed byte-identical via `git diff --stat`. The three code-review findings (WR-01, WR-02, IN-01) are real but non-blocking quality issues, already disclosed in `05-REVIEW.md` with `status: issues_found` and 0 critical findings — they do not threaten the correctness of any shipped control or the phase's evidentiary claims.

---

_Verified: 2026-09-06T19:21:45Z_
_Verifier: Claude (gsd-verifier)_
