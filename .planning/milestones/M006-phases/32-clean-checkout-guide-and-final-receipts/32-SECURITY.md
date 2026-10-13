---
phase: "32"
slug: "clean-checkout-guide-and-final-receipts"
status: verified
# threats_open = count of OPEN threats at or above workflow.security_block_on severity (the blocking gate)
threats_open: 0
asvs_level: 1
created: "2026-10-12"
---

# Phase 32 — Security

> Per-phase security contract: threat register, accepted risks, and audit trail.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| committed source -> formatter output | `schway format` output is a source of truth for reviewers | source text (integrity) |
| frozen golden -> allow-list / retired list | an allow-list or retired-path list can silently widen or hide what a frozen golden accepts | test expectations (integrity) |
| workflow file -> hosted runner | ci.yml env decides timeouts and whether native checks may skip | CI config (integrity) |
| shared test index -> meta-tests | a cached mutable index could leak one test's mutation into another | parsed test index (integrity) |
| live docs -> reader | a stale line could present deleted frontier as current | documentation claims |
| README text -> shell / guide checker | guide lines are executed by sh in tests and on hosted CI | executable shell text (integrity) |
| temp input path -> program stderr | an error message could echo a private path | local paths (low sensitivity) |
| git tree -> codedigest | the digest decides which receipts are current | content digest (integrity) |
| CLI dispatch -> receipts | a public route the receipts do not exercise could be presented as covered | coverage claims |
| hosted runner -> receipt log -> gate | logs are produced on runners, uploaded, and gated | receipt logs (untrusted until gated) |
| pushed HEAD / run headSha / origin -> freeze | a mismatched commit could bind receipts to unfrozen code | commit SHAs (integrity) |
| agent -> remote repository | push and dispatch act on shared state with user credentials | user credentials / remote state |
| downloaded artifact -> repository | hosted logs are untrusted bytes until gated | artifact bytes |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-32-01 | Tampering | internal/compiler/syntax/format.go | high | mitigate | `TestPhase32FormatVarBody` (internal/compiler/syntax/syntax_phase32_format_test.go) asserts that the semantic token sequence is preserved | closed |
| T-32-02 | Tampering | phase30FormatChanges | medium | mitigate | 3 dated entries in session_phase30_prechange_test.go, with a stale-entry control | closed |
| T-32-03 | Denial of Service | evidence-aggregate verify-phase6 step | high | mitigate | `GOFLAGS: -timeout=20m` (ci.yml:154), pinned by `TestPhase32HostedCIEnv` | closed |
| T-32-04 | Tampering | buildTestIndex cache | medium | mitigate | Deep copy per call; `TestPhase32TestIndexCacheIsolation` (session_phase32_ci_test.go) | closed |
| T-32-05 | Elevation of Privilege | hosted native skip | medium | mitigate | `SCHWAY_REQUIRE_NATIVE: "1"` on both jobs (ci.yml:57, :97) | closed |
| T-32-06 | Tampering | phase28RetiredEntries / phase30RetiredPaths | medium | mitigate | An observed retired path fails both gates (session_phase28_manifest_test.go, session_phase30_prechange_test.go); a tombstone scan also runs | closed |
| T-32-07 | Repudiation | dated historical records | low | mitigate | 32-03 acceptance keeps milestones/, research/ and LANGUAGE-MATURITY.md:180 unchanged; only a dated amendment is added | closed |
| T-32-08 | Information Disclosure | live docs | low | mitigate | `phase32LiveDocViolations` (session_phase32_retire_test.go), with armed controls | closed |
| T-32-09 | Tampering | README-driven sh execution | high | mitigate | `t.TempDir()` copy plus bounded `RunCLIErr`; dir and line passed as positional sh args; outputs confined to `.schway/` (main_phase32_guide_test.go:188, :204, :310) | closed |
| T-32-10 | Information Disclosure | app error rows | medium | mitigate | `ChecksumSecretSegment` absence asserted on stderr (main_phase32_guide_test.go, testsupport/checksum_oracle.go) | closed |
| T-32-11 | Spoofing | guide numbers | medium | mitigate | Cross-checked against checksum.expected.json and the independent oracle (main_phase32_guide_test.go, session_phase32_gate_test.go) | closed |
| T-32-12 | Elevation of Privilege | hosted native skip | medium | mitigate | `phase32NativePolicy`, pinned by `TestPhase32NativePolicy` (main_phase32_guide_test.go) | closed |
| T-32-13 | Tampering | codedigest path set | high | mitigate | `phase32DigestPaths`, one algorithm in sh and Go proven equal, armed controls (session_phase32_receipt_test.go, session_phase32_freeze_test.go) | closed |
| T-32-14 | Repudiation | stale receipt presented as current | high | mitigate | `stale_bound_claim` in phase32ReceiptProblems, with log hash, digest and host/program rules (session_phase32_receipt_test.go) | closed |
| T-32-15 | Spoofing | forged receipt log | medium | accept | Consistency-only gate; authenticity comes from the hosted run ID. Stated in scripts/check-phase32-receipts.sh:15-17 | closed |
| T-32-16 | Elevation of Privilege | unaudited public route | medium | mitigate | AST route audit `TestPhase32PublicRouteAudit` (cmd/schway/main_phase32_routes_test.go) | closed |
| T-32-17 | Tampering | check-phase32-receipts.sh | high | mitigate | Named rejections (digest_mismatch_, result_host_mismatch_, result_programs_mismatch_, row-count rules, etc.), armed in `TestPhase32ReceiptGate` (session_phase32_gate_test.go) | closed |
| T-32-18 | Spoofing | forged log | medium | accept | Consistency-only gate (header, scripts/check-phase32-receipts.sh:15-17); authenticity comes from the hosted run ID and artifact | closed |
| T-32-19 | Repudiation | skipped native control | high | mitigate | `run_control` fails on `--- SKIP:` / `no tests to run` (scripts/verify-phase32.sh:189-194); SCHWAY_REQUIRE_NATIVE on hosted jobs | closed |
| T-32-20 | Tampering | post-freeze code change | high | mitigate | `digest_path_changed_after_freeze` in scripts/check-phase32-freeze.sh, run in 32-10 and 32-11 | closed |
| T-32-21 | Denial of Service | hosted step hides failure | medium | mitigate | `shell: bash` (ci.yml:121, :134), `if-no-files-found: error` (ci.yml:131, :144), pinned by `TestPhase32CIHostMatrix` | closed |
| T-32-22 | Elevation of Privilege | git push / gh workflow run | high | mitigate | The agent never ran either; the user ran the paste-ready line (32-10-SUMMARY.md:55) | closed |
| T-32-23 | Tampering | dispatched revision | high | mitigate | check-phase32-freeze.sh re-armed live: a synthetic headSha was rejected with `sha_mismatch` (32-10-SUMMARY.md:48, :64) | closed |
| T-32-24 | Denial of Service | cancel-in-progress | medium | mitigate | Single dispatch (run 38315799626); Task 3 stop rule for duplicate runs | closed |
| T-32-25 | Tampering | downloaded logs | high | mitigate | Logs gated at FREEZE_DIGEST; `logSha256` recorded in 32-RECEIPT.json and re-verified by `TestPhase32ReceiptLivingClaims` | closed |
| T-32-26 | Repudiation | receipt presented after code drift | high | mitigate | `phase32BoundMarker` plus `stale_bound_claim` (session_phase32_receipt_test.go:172); freeze gate passed for run headSha and HEAD (32-11-SUMMARY.md:82) | closed |
| T-32-27 | Spoofing | receipt authenticity | medium | accept | Consistency-only by decision; authenticity is the recorded hosted RUN_ID 38315799626 and its artifacts | closed |
| T-32-28 | Tampering | phase32GuideStaticProblems | medium | mitigate | Five in-memory negative controls with anchor-present checks (main_phase32_guide_test.go) | closed |
| T-32-29 | Repudiation | not-covered reasons | low | mitigate | `bad_reason` requires a go/parser-found Test function (main_phase32_routes_test.go) | closed |
| T-32-30 | Spoofing | run headSha / origin tip | high | mitigate | `sha_mismatch` in check-phase32-freeze.sh, armed in `TestPhase32FreezeGate` and re-armed live | closed |
| T-32-31 | Tampering | freeze gate path list | medium | mitigate | `paths-match-digest` asserts that the script list equals phase32DigestPaths (session_phase32_freeze_test.go) | closed |
| T-32-SC | Tampering | npm/pip/cargo installs | high | mitigate | No package installs; go.mod/go.sum unchanged across the phase; actions/upload-artifact@v4 was already in use | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

---

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-32-01 | T-32-15 | The receipt gate checks consistency, not authenticity; authenticated receipts are deferred per CONTEXT. Authenticity comes from the hosted run ID. | plan 32-05 threat model | 2026-10-12 |
| AR-32-02 | T-32-18 | Same consistency-only scope for the receipts gate script; this is stated in its header. | plan 32-08 threat model | 2026-10-12 |
| AR-32-03 | T-32-27 | Receipt authenticity is the recorded hosted RUN_ID (38315799626) and its artifacts. | plan 32-11 threat model | 2026-10-12 |

*Accepted risks do not resurface in future audit runs.*

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-12 | 32 | 32 | 0 | /gsd-secure-phase (L1 grep verification, auditor short-circuited) |

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-12
