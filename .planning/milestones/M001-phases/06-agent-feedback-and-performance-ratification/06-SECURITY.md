---
phase: "06"
slug: "agent-feedback-and-performance-ratification"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-09-07"
---

# Phase 06 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

Register origin: **authored at plan time**. All 15 plans (06-01 … 06-15) carried a
`<threat_model>` block; this document consolidates them and records verification
of each declared mitigation against the implemented tree.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| CLI argv → `session.ExplainCommandFile` / `session.QueryCommandFile` | Untrusted operand strings (path, ID, depth, `--kind`, `--cursor`) | Attacker-controllable strings |
| source file on disk → parser | Untrusted source bytes, bounded at `syntax.MaxSourceBytes` | Arbitrary file bytes |
| synthesized graph / joined facts → JSON stdout | Must respect the 64 KiB-plus-one output bound | Diagnostic prose, structured facts |
| cache directory on disk → `cache.Store.Get` | UNTRUSTED: hand-editable by a user or another process | Compiled artifacts, `meta.json` |
| spawned `clang` / platform probe → `ProbeClangIdentity`, `ProbeMachine` | Untrusted, potentially unbounded subprocess stdout/stderr | Tool version/identity bytes |
| declared input list → cache key | Soundness seam: anything undeclared is an escape by construction | Input digests |
| change-state file on disk → `LoadChangeState` | UNTRUSTED: locally editable/truncatable | Lane IDs, fixture-relative names |
| `risk_lanes.json` / `qlt02_budget_manifest.json` (embedded) → selector, gating | Repo-tracked, reviewed like code; the JSON diff IS the review surface | Lane declarations, ratified budgets |
| host environment → `MachineFacts` | Privacy seam: what about this host may enter a published identifier | Six declared facts only |
| raw timing samples → gate verdict | Honesty seam: a fabricated percentile would become a ratified budget | Wall-clock samples |
| new `/1` reporting fields → content identity | The D-06-32 seam a measurement field must never cross | Metrics, lane, stage fields |
| `lang` subprocess stdout → `cmd/lang-repair` driver | The driver's ONLY permitted input; bounded and strictly decoded | `lang.command/1` JSON |
| `replacement` string → source file splice | Tampering seam: an edit must not escape its declared span | Repair text |
| `cmd/lang-repair` package → `internal/` | Structural boundary D-06-28 enforces by build failure | Go import graph |
| script control list ↔ Go control set | Drift here is how a control silently stops being asserted | Control IDs, bounds |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-06-01 | Tampering | `protocol.Result.Finalize()` identity struct | high | mitigate | `TestMetricsAndLaneFieldsExcludedFromIdentity`, `TestIdentityFieldEnumerationIsExhaustive` — `internal/compiler/protocol/protocol_test.go` | closed |
| T-06-02 | Tampering | prior-phase frozen core/evidence/golden bytes | high | mitigate | `TestPreviousPhaseCoreBytesUnchanged`, `TestPreviousPhaseManifestIDsUnchanged`, `TestPreviousPhaseGoldenCUnchanged` — `internal/compiler/core/core_test.go` | closed |
| T-06-03 | Repudiation | half-landed 12-site schema bump | medium | mitigate | `TestLaneSchemaLiteralSiteCountIsPinned` — `session_phase6_pin_test.go` | closed |
| T-06-EXPLAIN-01 | Denial of Service | cause-DAG synthesis | high | mitigate | `ExplainMaxNodes` bound + `--depth`; `truncated:explain.*` codes — `protocol.go`, `session_phase6_explain.go` | closed |
| T-06-EXPLAIN-02 | Denial of Service | source read in `ExplainCommandFile` | medium | mitigate | `readBoundedFile(path, syntax.MaxSourceBytes)` — `session_phase6_explain.go` | closed |
| T-06-EXPLAIN-03 | Information Disclosure | `detail` prose in graph nodes | low | accept | AR-01 — no new disclosure surface; detail is already-printed compiler output | closed |
| T-06-EXPLAIN-04 | Spoofing | fabricated `availability` value | medium | mitigate | Closed `debugmap.Availability` vocabulary; `TestExplainSummarySchemaIsMinted` | closed |
| T-06-QUERY-01 | Denial of Service | unbounded fact list on a broad address | high | mitigate | `QueryMaxFactsPerPage` + `truncated:query.page_bound` — `protocol.go`, `session_phase6_query.go` | closed |
| T-06-QUERY-02 | Tampering | malformed/forged `--cursor` | medium | mitigate | Content-derived opaque cursor; malformed cursor is a usage error, never page-one fallback — `session_phase6_query.go` | closed |
| T-06-QUERY-03 | Spoofing | fabricated resolution for an unknown ID | high | mitigate | `TestQueryUnknownIDReportsNotCaptured` — `session_phase6_query_test.go` | closed |
| T-06-QUERY-04 | Information Disclosure | joined facts leaking beyond the addressed identity | low | accept | AR-02 — every vocabulary routes to an already-shipped lookup governing its own surface | closed |
| T-06-CACHE-01 | Tampering | hand-edited or partially-deleted cache directory | high | accept | AR-03 — D-06-13 hole 4; recorded in the 06-15 escape register, never hidden | closed |
| T-06-CACHE-02 | Denial of Service | unbounded `clang` probe stdout/stderr | high | mitigate | `MaxProbeBytes` (64 KiB+1) bounded writer + 5s `context.WithTimeout` in `runBoundedProbe` — `cache/probe.go` | closed |
| T-06-CACHE-03 | Spoofing | a Clang change that does not move its version string | high | mitigate | `clang_identity` digests the binary's own bytes; `TestClangDigestIsNotJustTheVersionString` — `cache/probe_test.go` | closed |
| T-06-CACHE-04 | Tampering | a verdict entering the cache | high | mitigate | `TestCacheExportedSurfaceStoresNoVerdict`, `TestCacheImportsStayIndependent` — `cache/cache_test.go` | closed |
| T-06-CACHE-05 | Denial of Service | unbounded artifact read from the cache directory | medium | mitigate | `io.LimitReader(file, MaxArtifactBytes+1)` + strict `meta.json` decode — `cache/cache.go:224` | closed |
| T-06-CACHE-06 | Tampering | undeclared environment altering an artifact | medium | accept | AR-04 — D-06-13 hole 1, recorded in the escape register | closed |
| T-06-RISK-01 | Tampering | edited change-state file skipping a lane | high | mitigate | Corrupt state widens to all lanes; deferral requires a matching declared-input hash; a not-run lane renders `deferred` — `session_phase6_risklanes.go` | closed |
| T-06-RISK-02 | Repudiation | a new lane silently escaping the declared table | high | mitigate | `control:risklanes.undeclared_lane` — `session_phase6_risklanes.go`, `scripts/verify-phase6.sh` | closed |
| T-06-RISK-03 | Tampering | a stale row masking a real gap | medium | mitigate | `control:risklanes.stale_lane_reference` cross-checked against `LiveLaneIDs()` | closed |
| T-06-RISK-04 | Denial of Service | unbounded state-file read | medium | mitigate | `readBoundedFile` at a declared cap + strict JSON decode — `session_phase6_risklanes.go` | closed |
| T-06-RISK-05 | Information Disclosure | state file recording paths outside the cache dir | low | accept | AR-05 — keys are fixture-relative names and lane IDs; no absolute host paths | closed |
| T-06-BUMP-01 | Tampering | already-published `/0` bytes | high | mitigate | Two-constant coexistence, `TestSchemaZeroConstantsStillExist`, three prior-phase byte pins re-run | closed |
| T-06-BUMP-02 | Tampering | a half-landed 12-site bump | high | mitigate | `TestLaneSchemaLiteralSiteCountIsPinned` | closed |
| T-06-BUMP-03 | Spoofing | free-text `cache_status` implying a cached verdict | high | mitigate | `ValidateLaneVocabularies` + closed accessors set-equal to `cache.CacheStatuses()` | closed |
| T-06-BUMP-04 | Tampering | a measurement field entering content identity | high | mitigate | 06-01 reflection identity-exclusion tests, updated and re-run | closed |
| T-06-VERIFY-01 | Tampering | a cache hit skipping an assertion | high | mitigate | `TestNativeDifferentialAssertionOutsideCacheBranch` (`go/ast` structural) — `session_phase6_verify_test.go:146` | closed |
| T-06-VERIFY-02 | Spoofing | a not-run lane rendering pass | high | mitigate | `addDeferredLane` hardcodes `StatusDeferred`; `TestDeferredLaneNeverRendersPass` | closed |
| T-06-VERIFY-03 | Repudiation | a lane silently absent from the Result | high | mitigate | `TestEveryLiveLaneIsAccountedFor` — selected+deferred ≡ `LiveLaneIDs()` | closed |
| T-06-VERIFY-04 | Denial of Service | unbounded expanded evidence trace | medium | mitigate | 64 KiB-plus-one output bound + `truncated:evidence.trace_bound` — `session_phase6_evidence.go` | closed |
| T-06-VERIFY-05 | Tampering | cached binary bytes written to a temp path and executed | medium | accept | AR-06 — same trust level as executing the freshly compiled artifact on the same host | closed |
| T-06-MEASURE-01 | Information Disclosure | host fingerprint in `machine_id` | high | mitigate | `TestMachineIDExcludesHostFingerprints` — source-level API absence + marshalled-facts substring absence | closed |
| T-06-MEASURE-02 | Denial of Service | unbounded probe stdout/stderr | high | mitigate | `MaxProbeBytes` bounded writer + 5s deadline in `runBoundedProbe` — `measure/machine.go:94` | closed |
| T-06-MEASURE-03 | Spoofing | a percentile over an incomplete sample set | high | mitigate | `Summary()` refuses empty/short/non-positive sample sets with typed errors — `measure/statistics.go` | closed |
| T-06-MEASURE-04 | Elevation of Privilege | a noisy metric silently becoming a blocking gate | high | mitigate | `Demote` has exactly one `VerdictBlocking` return position (structural assertion); only `recomputed_work` is gate-eligible | closed |
| T-06-MEASURE-05 | Repudiation | drift between the shell sampling loop and the Go statistics | medium | mitigate | `measure.WarmSampleCount` and thresholds pinned against the script — `session_phase6_test.go:153` | closed |
| T-06-BUDGET-01 | Tampering | editing a ratified number until the run is green | high | mitigate | `AuditQLT02BudgetManifest` cross-checks declared machine IDs against live probe; no code path writes the manifest | closed |
| T-06-BUDGET-02 | Spoofing | an undeclared machine claiming ratified results | high | mitigate | `RatificationMode` → `Ratified: false`; every lane becomes `not_ratified`; manifest access behind the branch, asserted structurally | closed |
| T-06-BUDGET-03 | Elevation of Privilege | a noisy wall-clock signal becoming a blocking gate | high | mitigate | `EvaluateBudget` blocks only for `recomputed_work`; agreement with `measure.Demote` asserted independently | closed |
| T-06-BUDGET-04 | Repudiation | a duplicate manifest row silently winning | medium | mitigate | `control:qlt02.duplicate_row` — `session_phase6_budget.go` | closed |
| T-06-BUDGET-05 | Information Disclosure | manifest rows encoding host identity | medium | mitigate | Rows key on `machine_id` only — six-fact short hash, no hostname/serial/MAC | closed |
| T-06-STAGE-01 | Tampering | an unconditional `time.Since` destabilizing pinned output | high | mitigate | `TimingObservationEnabled` single gate; `TestTimingEnvironmentIsReadInExactlyOnePlace`, `TestStageBreakdownAbsentWhenUnobserved` | closed |
| T-06-STAGE-02 | Tampering | stage timings entering content identity | high | mitigate | `TestStageBreakdownExcludedFromIdentity` + 06-01 reflection sentinel loop | closed |
| T-06-STAGE-03 | Spoofing | a fabricated peak-RSS number encoding the current host | medium | mitigate | `TestPeakRSSStaysUnavailable`, `TestNoGetrusageAnywhere` — `protocol_test.go` | closed |
| T-06-STAGE-04 | Denial of Service | a tracing runtime under the guise of stage attribution | medium | mitigate | Fixed five-name stage set; `TestStageRecorderFileDeclaresNoMutableStateOrGoroutines`, `TestStageRecorderRefusesUnknownStage` | closed |
| T-06-REPAIR-01 | Tampering | a coordinate shift changing a published diagnostic ID | high | mitigate | `span`/`replacement`/`applicability` excluded from `repairKinds`; `TestRepairSpanShiftDoesNotChangeDiagnosticID` | closed |
| T-06-REPAIR-02 | Elevation of Privilege | an incomplete or defaulted repair applied mechanically | high | mitigate | `DriverEligible` requires `MachineApplicable` AND span AND replacement; empty applicability normalizes to `Unspecified` | closed |
| T-06-REPAIR-03 | Tampering | `lang.diagnostic/0` bytes moving | high | mitigate | `TestDiagnosticZeroBytesUnchanged` — `diagnostic/diagnostic_test.go` | closed |
| T-06-REPAIR-04 | Tampering | a `replacement` escaping its span when applied | medium | mitigate | Enforced consumer-side in 06-13: `TestRepairDriverNeverOpensSourceOutsideSpan` + import-boundary tests | closed |
| T-06-INJECT-01 | Spoofing | an injector producing an unmutated fixture that passes | high | mitigate | `phase6.injector_target_missing` marker guard driven from `AllInjectors()`; `TestEveryInjectorRefusesWhenMarkerDisappears`, `TestInjectorMarkerCountGuardIsNotInert` | closed |
| T-06-INJECT-02 | Repudiation | the CI gate scored on derivation fixtures | high | mitigate | Disjoint `heldout_` / `derivation_` prefixes; `TestPhase6DefectCorpusIsHeldOut` | closed |
| T-06-INJECT-03 | Tampering | an injector producing more than one change | high | mitigate | `TestEveryInjectorProducesExactlyOneMechanicalChange` across all four source-granularity classes | closed |
| T-06-INJECT-04 | Tampering | the cleanup class drifting from the shipped runner | medium | mitigate | `TestCleanupInjectorReusesReleaseOmissionRunner` (`go/ast` delegation assertion) | closed |
| T-06-INJECT-05 | Repudiation | the stale-evidence class located by a source diff | medium | mitigate | `TestStaleEvidenceInjectorBreaksManifestBinding` — no source diff computed; locator is `ValidateEvidenceCommandFile` | closed |
| T-06-BOUNDARY-01 | Elevation of Privilege | the driver importing compiler internals | high | mitigate | `TestRepairDriverImportsStayOutsideInternal` (broad `/internal/` scan) + `TestImportBoundaryTestIsNotInert` positive control | closed |
| T-06-BOUNDARY-02 | Tampering | a `replacement` splice escaping its declared span | high | mitigate | Out-of-range spans refused, never clamped; `TestRepairDriverNeverOpensSourceOutsideSpan` | closed |
| T-06-BOUNDARY-03 | Spoofing | the driver scraping prose | high | mitigate | `message`/`detail` are not declared in the driver's local structs, so cannot be decoded (`repair.go:44-60`); `TestRepairDriverDecodesNoProseFields` | closed |
| T-06-BOUNDARY-04 | Spoofing | the degenerate delete-the-code repair passing the oracle | high | mitigate | Byte-identity with the pre-defect original required; `TestRepairOracleRejectsDeleteTheCode` positive control | closed |
| T-06-BOUNDARY-05 | Denial of Service | unbounded subprocess stdout or a hanging `lang` | medium | mitigate | `maxStdoutBytes` (1 MiB) bounded writer + `subprocessTimeout` 30s — `cmd/lang-repair/repair.go:38-41,158` | closed |
| T-06-BOUNDARY-06 | Tampering | the driver spawning a program other than `lang` | medium | mitigate | `TestRepairDriverSpawnsOnlyTheLangBinary` — every exec site uses the `--lang` path variable | closed |
| T-06-THEATER-01 | Spoofing | the driver scraping prose while appearing structure-driven | high | mitigate | `TestProseScrambleLeavesRepairBehaviourIdentical` + `TestVocabularyRemovalDrivesTheDriverRed` | closed |
| T-06-THEATER-02 | Spoofing | a guard that can only ever assert green | high | mitigate | `TestVocabularyRemovalGuardIsNotInert`, `TestProseScrambleFixtureKeepsStructuredFieldsIntact` | closed |
| T-06-THEATER-03 | Elevation of Privilege | a driver-side test hook becoming a second untested path | high | mitigate | Fixture-substitution subprocess, no test hook; `TestFixtureSubstitutionIsFaithful` | closed |
| T-06-THEATER-04 | Elevation of Privilege | the fallback oracle pulling `internal/compiler/reduce` into the driver | high | mitigate | Split oracle in `session_phase6_oracle_test.go`; import scan asserts no `cmd/lang-repair` file imports the reducer; `TestDifferentialFallbackOracleRunsInProcess` | closed |
| T-06-THEATER-05 | Tampering | a mutation accidentally altering a structured field | medium | mitigate | `TestMutateCaptureIdentityRoundTripIsByteStable`; a violated assertion fails, never skips | closed |
| T-06-GATE-01 | Repudiation | script and Go control sets drifting apart | high | mitigate | `TestPhase6RequiredControlsMatchScript` (bidirectional set equality) | closed |
| T-06-GATE-02 | Spoofing | an accepted escape presented as a solved control | high | mitigate | `TestPhase6EscapesAreNeverPresentedAsControls` (empty intersection) — `session_phase6_escapes_test.go` | closed |
| T-06-GATE-03 | Tampering | a prior gate script forked or invoked | medium | mitigate | `TestPhase6ScriptInvokesNoPriorGate`; prior corpora run with this phase's freshly built binary | closed |
| T-06-GATE-04 | Tampering | a duplicated bound in the script drifting from its Go constant | medium | mitigate | `TestPhase6BoundsMatchScript` — `session_phase6_test.go` | closed |
| T-06-GATE-05 | Spoofing | a corpus with no marker treated as a Phase 6 corpus | medium | mitigate | `isPhase6Corpus` requires an `os.Stat` marker hit; `TestPhase6CorpusDispatchRequiresMarker` — `cmd/lang/main_test.go` | closed |
| T-06-GATE-06 | Repudiation | fabricated values in the recorded agent exercise | medium | mitigate | D-06-30 honest-reporting instruction carried in the human-check body; recorded as manual, non-gating | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-01 | T-06-EXPLAIN-03 | `detail` prose in explain graph nodes is the diagnostic message the compiler already prints; no new disclosure surface is opened. | Phase 06 plan (06-02) | 2026-09-07 |
| AR-02 | T-06-QUERY-04 | Every query vocabulary routes to an already-shipped lookup that governs its own disclosure surface; the join exposes no new data. | Phase 06 plan (06-03) | 2026-09-07 |
| AR-03 | T-06-CACHE-01 | D-06-13 hole 4: the cache directory is hand-editable. No integrity check beyond content-hash lookup is claimed; a `meta.json` input-list mismatch is treated as absent, which bounds but does not close the hole. Declared in the 06-15 escape register and surfaced in gate output. | Phase 06 plan (06-04) | 2026-09-07 |
| AR-04 | T-06-CACHE-06 | D-06-13 hole 1: undeclared environment (locale, ulimit, filesystem case-sensitivity) may alter an artifact. Recorded in the 06-15 escape register. | Phase 06 plan (06-04) | 2026-09-07 |
| AR-05 | T-06-RISK-05 | Change-state keys are fixture-relative names and lane IDs; no absolute host paths are recorded. | Phase 06 plan (06-05) | 2026-09-07 |
| AR-06 | T-06-VERIFY-05 | Executing a cached compiler artifact is the same trust level as executing the freshly compiled one on the same host; the cache directory is already an accepted escape (AR-03). | Phase 06 plan (06-07) | 2026-09-07 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-09-07 | 71 | 71 | 0 | /gsd-secure-phase (ASVS L1, block_on: high) |

**Method (ASVS L1):** register consolidated from the 15 plan-time `<threat_model>`
blocks; each `mitigate` disposition verified by locating its named control
(test symbol, constant, or structural guard) in the implemented tree; each
`accept` disposition transcribed into the Accepted Risks Log above. No new
threat scanning was performed — the register was authored at plan time and is
treated as complete.

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-09-07
