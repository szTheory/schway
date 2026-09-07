---
phase: 06-agent-feedback-and-performance-ratification
verified: 2026-09-07T05:37:39Z
status: passed
score: 4/4 roadmap success criteria met (1 with a documented, pre-accepted partial gap — see note on SC4)
behavior_unverified: 0
overrides_applied: 0
accepted_known_debt:
  - id: WR-01
    source: 06-REVIEW.md
    finding: "D-06-19's cold-sample distribution (measure.ColdSampleCount) is declared but no production path collects it, and Samples.Summary() structurally refuses any sample set shorter than WarmSampleCount (20). scripts/verify-phase6.sh's observe() only collects 20 warm samples."
    disposition: "Accepted, pre-existing, documented in 06-REVIEW.md as a Warning (not Critical). Not re-flagged as a new blocking gap per verification task instructions. Affects SC4's literal wording ('cold/warm feedback distributions') — the warm half is real and gated correctly; the cold half is absent."
  - id: WR-02
    source: 06-REVIEW.md
    finding: "query's cursor pagination uses strict-inequality tie-breaking that would silently drop ties if any resolver ever produced them. Not reachable today — all five vocabulary resolvers return unique-keyed facts."
    disposition: "Accepted latent fragility, not a live bug. Recorded, not re-flagged."
  - id: D-06-13
    disposition: "Four named, permanent artifact-cache soundness holes — declared and asserted never-detected by session.Phase6ExpectedEscapes(). Accepted residual per 06-DEBT.md."
  - id: D-06-20
    disposition: "Peak RSS permanently unavailable for M001 (no getrusage anywhere, TestNoGetrusageAnywhere enforces this). Accepted, deliberate gap per 06-DEBT.md."
  - id: D-06-29
    disposition: "Held-out corpus discipline blunts, does not eliminate, repair-driver overfitting risk. Accepted residual per 06-DEBT.md."
  - id: D-06-30
    disposition: "The non-gating agent-legibility exercise (repair_rounds vs p95:2) has not been run as part of phase close. Explicitly non-gating, per-milestone/on-demand cadence. Accepted per 06-DEBT.md."
  - id: D-06-33
    disposition: "Four unclassified-category edge probes (one per requirement area) never resolved across the phase. Carried to a future verification pass per 06-DEBT.md."
---

# Phase 6: Agent Feedback and Performance Ratification Verification Report

**Phase Goal:** AI agents and human reviewers can inspect, repair, and verify the
compiler corpus through bounded stable protocols whose cost is measured and enforced.

**Verified:** 2026-09-07T05:37:39Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths (ROADMAP Success Criteria)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `explain`/`query` return bounded versioned facts and cause graphs by stable identity without prose scraping or whole-program dumps | ✓ VERIFIED | `lang.explain/0`, `lang.query/0` minted (`protocol.go`); `TestQueryResolvesEveryStableIDVocabulary`, `TestQueryMintsNoSixthVocabulary` confirm the joined addressing surface (D-06-01); `TestExplainCauseGraphIsDeterministic` proves the per-cold-invocation synthesis obligation (D-06-02) is actually tested, not assumed; `TestQueryCursorPaginationIsBounded`, `TestQueryPagesConcatenateExactlyOnce`, `TestQueryDepthDoesNotPaginate` cover D-06-03's bounding discipline; `cmd/lang/main.go:47-50` wires `explain SRC ID` / `query SRC ID` positional dispatch per D-06-05's amended operand form; no persisted graph store exists anywhere in `internal/compiler/session/session_phase6_explain.go` / `_query.go`. |
| 2 | `verify` selects evidence lanes by changed risk and reports reused cache inputs; `evidence` validates compact manifests and expands traces on failure/request | ✓ VERIFIED | `internal/compiler/cache` package exists with `TestCacheExportedSurfaceStoresNoVerdict`, `TestOutcomeFieldSetIsPinned` (pins `Outcome{Artifact,Key,Status}` — no verdict field can be silently added), `TestCacheImportsStayIndependent` (cache never imports protocol/diagnostic/session) — confirms D-06-06's one-way "artifacts never verdicts" property is real, not aspirational. `session_phase6_risklanes.go` + `risk_lanes.json` + `TestRiskLaneUnclassifiedFixtureWidensToAllLanes`/`TestRiskLaneUndeclaredInputIsNotCacheable` confirm D-06-11's fail-closed widening. **CR-01 (critical finding, 06-REVIEW.md) was fixed in commit `6f536ea`**: `VerifyPhase6ControlsAndWork`'s `CacheInputsReusedCount` now increments on `cache.StatusArtifactReused`, matching the sibling `VerifyPhase6ChangedRisk`. Verified live, not just by reading the diff: a fresh `sh scripts/verify-phase6.sh` run (this session) shows `"cache_status":"artifact_reused"` on `lane:native-differential` and `"cache_inputs_reused_count":1` in the phase-6 gate's own JSON output — the exact code path `scripts/verify-phase6.sh` drives end-to-end. `TestEvidenceExpandsTraceOnFailure`, `TestEvidenceExpandsTraceOnRequest`, `TestEvidenceTraceRespectsOutputBound` cover on-demand trace expansion. |
| 3 | An automated repair exercise fixes representative match, move, borrow, cleanup, and stale-evidence defects using only the supported command protocol | ✓ VERIFIED | `cmd/lang-repair` is a separate binary; `import_boundary_test.go` enforces D-06-28's structural (not conventional) protocol-only boundary. All three D-06-27 anti-theater guards are real and independently self-verifying: **Guard 1** `TestProseScrambleLeavesRepairBehaviourIdentical` covers all five classes and is backed by `TestProseScrambleFixtureKeepsStructuredFieldsIntact` (proves the scrambler changed prose and nothing else — not vacuous). **Guard 2** `TestVocabularyRemovalDrivesTheDriverRed` (four strip modes) is paired with `TestVocabularyRemovalGuardIsNotInert`, a load-bearing control on the identical harness proving the guard can also go green — a guard that can only ever assert RED would not detect its own inertness. **Guard 3** `TestEveryInjectorRefusesWhenMarkerDisappears` is driven from `AllInjectors()` (fails loudly if a future 6th injector isn't covered) and paired with `TestInjectorMarkerCountGuardIsNotInert`, which demonstrates concretely what the unguarded path would silently do. `TestRepairOracleRejectsDeleteTheCode` confirms D-06-26's anti-degenerate-repair oracle. |
| 4 | Declared-machine cold/warm feedback distributions, peak memory, output bytes, and affected work meet ratified budgets or identify a specific blocking regression | ⚠️ PARTIALLY MET (accepted, documented) | `session/qlt02_budget_manifest.json` + `AuditQLT02BudgetManifest` + `machine_id` probe (D-06-17) + observation-only mode for undeclared machines (D-06-18) are all real and gated (`lane:qlt02-budget-audit` fires green in the live run). `recomputed_work`-only blocking rule (D-06-14/D-06-22) is exact and deterministic. `peak_rss_status` is honestly `"unavailable"` everywhere, mechanically enforced by `TestNoGetrusageAnywhere` (no `getrusage` reference anywhere in the tree) — this is a **deliberate, accepted D-06-20 design choice**, not a gap. `output_bytes` and `recomputed_work` are live and correctly reported in every lane of the fresh run. **The warm half of D-06-19 is real** (`sh scripts/verify-phase6.sh`'s `observe()` collects 20 warm samples through `lang stats`, producing `{"p50":...,"p95":...,"cov":...,"count":20}` in this session's live run). **The cold half is not implemented**: `measure.ColdSampleCount` is declared but no production code path anywhere collects it, and `Samples.Summary()` structurally refuses any sample set shorter than 20 — this is 06-REVIEW.md's WR-01, a pre-existing, already-documented finding (not a new discovery), recorded here as an accepted partial gap in the literal "cold/warm" wording of this success criterion rather than a full failure, per the phase's own D-06-15 stance that wall-clock distributions are non-gating observations. |

**Score:** 3/4 success criteria fully verified; 1/4 (SC4) partially verified — the deterministic hard gate (`recomputed_work`), machine ratification, output-bytes, and honest peak-RSS unavailability are all real and correctly wired; the warm-distribution reporting is real and live; the cold-distribution reporting named in D-06-19 is declared but unwired (WR-01, pre-existing, documented, non-blocking per this project's own non-gating-observation design for wall-clock metrics).

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/cache/` | New cache package, artifacts-never-verdicts | ✓ VERIFIED | `cache.go`, `probe.go` exist; `TestCacheExportedSurfaceStoresNoVerdict`/`TestOutcomeFieldSetIsPinned`/`TestCacheImportsStayIndependent` all pass |
| `internal/compiler/session/session_phase6_explain.go` | `explain` synthesis | ✓ VERIFIED | Exists, wired via `cmd/lang/main.go:47-49`, `lang.explain/0` minted |
| `internal/compiler/session/session_phase6_query.go` | `query` addressing/pagination | ✓ VERIFIED | Exists, wired via `cmd/lang/main.go:50-52`, `lang.query/0` minted |
| `internal/compiler/session/session_phase6_risklanes.go` + `session/risk_lanes.json` | Declared risk→lane table + audit | ✓ VERIFIED | Registry loads, audit refuses stale/undeclared lane IDs (`TestRiskLaneAuditRefusesStaleLaneID` etc.) |
| `session/qlt02_budget_manifest.json` + budget code | Ratified budgets, declared machines | ✓ VERIFIED | `AuditQLT02BudgetManifest`, live `lane:qlt02-budget-audit` passes in fresh run |
| `cmd/lang-repair/` | Protocol-only repair driver | ✓ VERIFIED | Separate binary, subprocess-only, import-boundary linted |
| `internal/compiler/session/session_phase6_injectors.go` | Five defect injectors | ✓ VERIFIED | `AllInjectors()` returns 5, each with marker-mutation-kill coverage |
| `internal/compiler/protocol/protocol.go` (`/1` bump) | Coordinated schema bump | ✓ VERIFIED | `lang.command/1`, `lang.verify-lane/1` both minted; `/0` constants still declared; no non-test production site left on `/0` (`grep` returned zero hits) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `cmd/lang/main.go` | `session.VerifyPhase6ControlsAndWork` / `session_phase6_verify.go` | `verify testdata/phase6` dispatch | ✓ WIRED | Confirmed by live run producing real lane JSON with `cache_status`/`cache_inputs_reused_count` |
| `session_phase6.go` (CI-gated path) | `cache.StatusArtifactReused` | `CacheInputsReusedCount` increment | ✓ WIRED (post-fix) | CR-01 fixed in `6f536ea`; live run shows `cache_inputs_reused_count:1` on a real cache-reuse lane |
| `cmd/lang-repair` | shipped `lang` binary | subprocess + `--json` only | ✓ WIRED | `import_boundary_test.go` structurally enforces no `internal/` import |
| `Result.Finalize()` | identity struct | D-06-32 constraint | ✓ VERIFIED | Read `protocol.go:455-519` directly — the anonymous `identity` struct contains `Schema/Command/Status/ModuleID/FormattedDigest/DiagnosticIDs/ExecutionDigests/EvidenceID/InterfaceID/DebugMapID/ExplainID/QueryID/LaneIDs/ExpectedEscapes`. No `Metrics` or per-lane `CacheStatus`/`GateVerdict`/timing field reaches it — only `lane.ID+":"+lane.Status` per lane. |
| `scripts/verify-phase6.sh` `observe()` | `measure.ColdSampleCount` | cold-sample collection | ✗ NOT WIRED (accepted, WR-01) | `observe()` only loops `phase6_warm_sample_count=20`; no cold-sample loop exists anywhere |

### Behavioral Spot-Checks / Live Run

| Check | Command | Result | Status |
|-------|---------|--------|--------|
| Full phase-6 gate | `sh scripts/verify-phase6.sh` | Exit code 0; all lanes `"status":"pass"`; `cache_inputs_reused_count:1` observed on a live cache-reuse lane; warm distribution `{"p50":153311125,"p95":162049917,"cov":0.0408,"count":20}` | ✓ PASS |
| Full test suite | `go test ./...` | All packages `ok` | ✓ PASS |
| Anti-theater guard 1 | `cmd/lang-repair/antitheater_test.go::TestProseScrambleLeavesRepairBehaviourIdentical` | Present, covers 5 classes, backed by non-vacuous control | ✓ PASS (enumerated) |
| Anti-theater guard 2 | `TestVocabularyRemovalDrivesTheDriverRed` + `TestVocabularyRemovalGuardIsNotInert` | Both present; RED assertion paired with GREEN control | ✓ PASS (enumerated) |
| Anti-theater guard 3 | `TestEveryInjectorRefusesWhenMarkerDisappears` + `TestInjectorMarkerCountGuardIsNotInert` | Driven from `AllInjectors()`, all 5 classes covered, paired with non-inert control | ✓ PASS (enumerated) |
| Cache one-way property | `TestCacheExportedSurfaceStoresNoVerdict`, `TestOutcomeFieldSetIsPinned`, `TestCacheImportsStayIndependent` | All present and substantive (AST-based identifier/import scans, not string matches) | ✓ PASS (enumerated) |

### Requirements Coverage

| Requirement | Source Plans | Status | Evidence |
|-------------|--------------|--------|----------|
| FND-04 | 06-01, 06-06, 06-07, 06-08, 06-09, 06-10, 06-15 | ✓ SATISFIED | `cache_status`, `recomputed_work`, `output_bytes` reported without changing `Result.Finalize()` identity (D-06-32 confirmed directly in code) |
| DX-02 | 06-02, 06-03, 06-15 | ✓ SATISFIED | `explain`/`query` live, tested, no 6th ID vocabulary minted |
| DX-03 | 06-01, 06-04, 06-05, 06-06, 06-07, 06-15 | ✓ SATISFIED | Cache substrate + risk-lane selector live and gated; CR-01 fix confirmed live |
| DX-04 | 06-11, 06-12, 06-13, 06-14, 06-15 | ✓ SATISFIED | Repair driver, 5 injectors, 3 anti-theater guards, anti-degenerate oracle all real |
| QLT-02 | 06-01, 06-06, 06-08, 06-09, 06-10, 06-15 | ⚠️ PARTIALLY SATISFIED | Budget ratification, machine declaration, deterministic gate, honest peak-RSS unavailability all real; cold-distribution half of D-06-19 unwired (WR-01, accepted pre-existing finding) |

No orphaned requirements — REQUIREMENTS.md maps exactly FND-04, DX-02, DX-03, DX-04, QLT-02 to Phase 6, and all five appear across the fifteen plans' `requirements:` frontmatter.

### Anti-Patterns Found

None. Scanned all Phase 6 production files (`session_phase6*.go`, `cache/*.go`, `protocol.go`, `cmd/lang-repair/*.go`, `scripts/verify-phase6.sh`) for `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`placeholder` — zero unreferenced debt markers found (the sole `XXXXXX` hit is a `mktemp` template pattern, not a debt marker).

### Human Verification Required

None. All four success criteria are either fully verified against the live codebase and a fresh end-to-end run, or (SC4) precisely and honestly characterized as partially met with a specific, already-documented, non-blocking gap (WR-01) rather than an open question requiring human judgment.

### Gaps Summary

No new gaps found. This verification pass confirms:

1. The one Critical finding from 06-REVIEW.md (CR-01, `cache_inputs_reused_count` dead code) was genuinely fixed in commit `6f536ea` — confirmed both by reading the diff and by observing `cache_inputs_reused_count:1` in a live, fresh `sh scripts/verify-phase6.sh` run against the exact CI-gated code path the bug lived in.
2. The two Warnings from 06-REVIEW.md (WR-01 cold-sample distribution unwired, WR-02 latent cursor tie-break fragility) remain open exactly as recorded — neither is a new discovery, and per this verification's brief they are accepted, documented scope rather than blocking gaps. WR-02 is confirmed unreachable given the current five resolvers' unique-ID guarantee.
3. The falsifiability core of DX-04 (the three D-06-27 anti-theater guards) is real: each has a paired, load-bearing control proving it is not itself inert or vacuous, which is the exact failure shape this project's process rules exist to prevent.
4. D-06-06's cache one-way property (artifacts, never verdicts) is enforced by AST-based structural tests, not string conventions, and the identifier-denylist hole (the ambiguous "status"/"outcome" words) is closed by a separate field-set pin test.
5. D-06-32 (no metrics field reaches `Result.Finalize()`'s identity struct) is confirmed by direct code reading, not just test presence.
6. All items in 06-DEBT.md (D-06-13, D-06-20, D-06-29, D-06-30, D-06-33) remain accurately characterized as accepted, non-blocking residual debt — none silently escalated to a blocker, none silently dropped.

The only place a genuine limitation touches a literal success-criterion wording is SC4's "cold/warm feedback distributions" — the warm half is real and live; the cold half (D-06-19) is declared but not produced by any code path. This is recorded as an accepted, pre-existing, non-blocking partial gap rather than a phase-blocking failure, consistent with the project's own D-06-15 design stance that wall-clock distribution metrics are observations, never hard gates.

---

## Re-Verification: 2026-09-07 (UAT automation pass)

**Trigger:** the fifteen Phase 6 SUMMARYs were edited after the initial pass, to
replace hand-driven UAT evidence with automated evidence. That makes the report
stale by mtime, so it is re-verified here rather than left to claim a result it
no longer covers.

**What changed since the initial verification.** Additive test coverage and CI
wiring only — no production behavior was modified:

1. `internal/compiler/testsupport/cli_phase6_test.go` (new) — six subprocess
   e2e tests driving the shipped binary for `explain`, `query` (resolution,
   closed-vocabulary, and honest-absence), the coordinated `/1` schema bump,
   and `lang stats`. These replace the five deliverables previously confirmed
   by a human at a terminal. `lang stats` had no test coverage of any kind
   before this pass, despite `scripts/verify-phase6.sh:151` depending on it.
2. `internal/compiler/testsupport/testsupport.go` — added `RunCLIStdin` /
   `RunCLIStdinErr`, sharing one `runCLI` spawn path with the existing
   `RunCLIErr`. Still `exec.CommandContext` + `WithTimeout` + `boundedWriter`
   on both output streams; `native_test.go`'s `TestSourceNeverSpawnsUnbounded
   Processes` and `TestSpawnGuardCatchesKnownEvasions` both re-run clean.
3. `internal/compiler/session/session_phase6_test.go` — added
   `TestCIWorkflowRunsPhase6Gate`, pinning `.github/workflows/ci.yml` to the
   current phase's gate. It reads the workflow as text and never executes it
   (the gate runs `go test ./...`, so executing it from a test would recurse).
4. `.github/workflows/ci.yml` — the `phase-gate` job ran `verify-phase4.sh`,
   two phases stale, against its own comment promising it would move forward.
   Retargeted to `scripts/verify-phase6.sh`. The job key is unchanged, so
   branch-protection rules keyed on it still match. **This was a real, live
   gap: Phase 5's and Phase 6's gates existed but nothing in CI ran them.**

`scripts/verify-phase6.sh` and `scripts/verify-phase5.sh` are byte-unchanged;
all five Phase 6 script-text contract tests still pass.

**Evidence.** Re-run end to end on the live tree for this pass: `go vet ./...`
clean; `go test ./...` clean; `go test -race ./...` clean (18 packages);
`sh scripts/verify-phase6.sh` green, exit 0. The new assertions were
mutation-checked rather than merely observed passing — the explain schema
literal, the explain node bound, the `not_captured` absence contract, and the
`/1` schema assertion were each temporarily inverted and each produced a
failure, then restored.

**Effect on this report's findings.** None are changed. Every accepted item in
the frontmatter (WR-01, WR-02, D-06-13, D-06-20, D-06-29, D-06-30, D-06-33)
remains open and accurately characterized; nothing was silently escalated or
dropped. SC4's cold-half gap is untouched. The `## Human Verification Required`
section above said "None" for the success criteria, and that is now also true
of the deliverable-level UAT: `uat.classify-coverage` reports
`all_auto_covered: true, present: 0, errors: 0` for all fifteen SUMMARYs, 68
deliverables, none requiring human judgment.

_Verified: 2026-09-07T05:37:39Z_
_Re-verified: 2026-09-07 (UAT automation pass)_
_Verifier: Claude (gsd-verifier)_
