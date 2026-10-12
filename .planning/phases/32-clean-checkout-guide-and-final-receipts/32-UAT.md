---
status: complete
phase: 32-clean-checkout-guide-and-final-receipts
source: [32-01-SUMMARY.md, 32-02-SUMMARY.md, 32-03-SUMMARY.md, 32-04-SUMMARY.md, 32-05-SUMMARY.md, 32-06-SUMMARY.md, 32-07-SUMMARY.md, 32-08-SUMMARY.md, 32-09-SUMMARY.md, 32-10-SUMMARY.md, 32-11-SUMMARY.md]
started: 2026-10-12T22:55:27Z
updated: 2026-10-12T22:55:32.826Z
---

## Current Test

[testing complete]

## Tests

### 1. D-10 (32-01)
expected: schway format keeps a trailing var-body result identifier on its own line; output reparses, is a fixed point, and preserves semantic tokens, checked on three terminal shapes and five committed sources as in-memory copies
result: pass
source: automated
coverage_id: D-10
covered_by: internal/compiler/syntax/syntax_phase32_format_test.go#TestPhase32FormatVarBody

### 2. D-10c (32-01)
expected: The frozen Phase 30 pre-change golden passes through a dated phase30FormatChanges allow-list whose stale-entry control is observed to fail
result: pass
source: automated
coverage_id: D-10c
covered_by: internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGolden, internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGoldenIsNotInert

### 3. ci-flake-timeout (32-02)
expected: The hosted verify-phase6 step carries GOFLAGS -timeout=20m, with an armed control that fails when the line is removed
result: pass
source: automated
coverage_id: ci-flake-timeout
covered_by: internal/compiler/session/session_phase32_ci_test.go#TestPhase32HostedCIEnv

### 4. D-04a (32-02)
expected: checks and evidence-aggregate both set SCHWAY_REQUIRE_NATIVE 1, with an armed control
result: pass
source: automated
coverage_id: D-04a
covered_by: internal/compiler/session/session_phase32_ci_test.go#TestPhase32HostedCIEnv

### 5. test-index-cache (32-02)
expected: The static test index is parsed once per process, each caller gets an isolated deep copy equal to a fresh parse and to go test -list
result: pass
source: automated
coverage_id: test-index-cache
covered_by: internal/compiler/session/session_phase32_ci_test.go#TestPhase32TestIndexCacheIsolation, internal/compiler/session/verification_groundedness_test.go#TestStaticTestIndexMatchesGoTestList

### 6. D-05 (32-03)
expected: examples/checksum.schway is deleted, the replacement exists, no renamed copy of the tombstone digest survives, and the armed control detects one
result: pass
source: automated
coverage_id: D-05
covered_by: internal/compiler/session/session_phase32_retire_test.go#TestPhase32ChecksumFrontierRetired

### 7. D-06 (32-03)
expected: Both frozen goldens stay byte-identical; the retired path may be absent and is a failure if observed again, proven by armed controls
result: pass
source: automated
coverage_id: D-06
covered_by: internal/compiler/session/session_phase28_manifest_test.go#TestPhase28EntryExitClassManifestIsNotInert, internal/compiler/session/session_phase30_prechange_test.go#TestPhase30PreChangeGoldenIsNotInert

### 8. D-07 (32-03)
expected: Dated PROJECT.md amendment plus live pointers; the live-doc scan passes and its armed and dated-exempt controls fire
result: pass
source: automated
coverage_id: D-07
covered_by: internal/compiler/session/session_phase32_retire_test.go#TestPhase32ChecksumFrontierRetired

### 9. D-07a (32-03)
expected: The LANGUAGE-MATURITY.md corpus census reads 162 programs and 5,352 lines
result: pass
source: automated
coverage_id: D-07a
covered_by: internal/compiler/session/self_describing_docs_test.go#TestLanguageMaturityCountsAreCurrent

### 10. SC1-D-01-D-02 (32-04)
expected: The checksum guide runs line by line from a git ls-files clean-checkout copy (go build, three native builds, one app run per table row for checksum, adler_low and adler32_cli) and every result matches the README cell
result: pass
source: automated
coverage_id: SC1-D-01-D-02
covered_by: cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout

### 11. D-01-D-03 (32-04)
expected: Limit, error exit, case names, stdout/stderr, adler_low = checksum mod 65536 and fixture numbers in the README agree with checksum.expected.json, the independent oracle rows and Phase29FixtureExpected; armed-json control observed to fail
result: pass
source: automated
coverage_id: D-01-D-03
covered_by: cmd/schway/main_phase32_guide_test.go#TestPhase32GuideMatchesExpectedJSON

### 12. D-01a (32-04)
expected: Every guide output target and sidecar is under .schway/ and mkdir -p precedes the first build
result: pass
source: automated
coverage_id: D-01a
covered_by: cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout

### 13. D-08-D-08a (32-05)
expected: codedigest.sh and the Go digest agree on the real checkout and in temp repos; a code edit moves the digest while a .md edit and a .planning edit do not
result: pass
source: automated
coverage_id: D-08-D-08a
covered_by: internal/compiler/session/session_phase32_receipt_test.go#TestPhase32CodeDigest

### 14. D-09-D-10a-D-08c (32-05)
expected: Receipt validator with named reasons and armed in-memory controls; reopen rule for bound versus reopened markers; passes with no receipt and no marker without t.Skip
result: pass
source: automated
coverage_id: D-09-D-10a-D-08c
covered_by: internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims

### 15. D-10b / D-10d (32-06)
expected: Every public CLI route enumerated by AST from run, runApplication and runApplicationRun, and every examples fn main program, has a covered row naming a receipt program or a not-covered row naming an existing test; a new route, deleted row, stale row, invented test name and unknown program each fail, and the usage text is cross-checked
result: pass
source: automated
coverage_id: D-10b / D-10d
covered_by: cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit

### 16. D-04 / D-01a (32-07)
expected: The guide checker rejects a changed value, a renamed command, a dropped program fence, a repo-root output target and a dropped table row on in-memory copies; the real README passes the static check before anything executes
result: pass
source: automated
coverage_id: D-04 / D-01a
covered_by: cmd/schway/main_phase32_guide_test.go#TestPhase32GuideParserNegativeControls, cmd/schway/main_phase32_guide_test.go#TestPhase32GuideCleanCheckout

### 17. D-04a (32-07)
expected: phase32NativePolicy fails a missing clang under GITHUB_ACTIONS=true or SCHWAY_REQUIRE_NATIVE=1, skips otherwise, and never fails when clang is present
result: pass
source: automated
coverage_id: D-04a
covered_by: cmd/schway/main_phase32_guide_test.go#TestPhase32NativePolicy

### 18. D-08b / D-09 (32-08)
expected: The two-host gate accepts an independent literal macOS + Linux pair and rejects a mutated digest, a code edit, a dirty tree, a dropped host, a duplicate host, dropped program and OOB rows, a swapped programs= list, an unknown row, altered cases and controls, a misplaced result and differing log digests; a .planning-only HEAD move keeps a receipt current
result: pass
source: automated
coverage_id: D-08b / D-09
covered_by: internal/compiler/session/session_phase32_gate_test.go#TestPhase32ReceiptGate

### 19. EVD-15 programs (32-08)
expected: The producer case table holds 39 rows over seven programs, each agreeing with checksum.expected.json, the ChecksumOracleRows adler_low values, Phase29FixtureExpected, the BYTE-02 stderr literal and the Phase 27 fixtures
result: pass
source: automated
coverage_id: EVD-15 programs
covered_by: internal/compiler/session/session_phase32_gate_test.go#TestPhase32ProducerCaseTable

### 20. D-10b (32-08)
expected: The route audit's covered program set equals the producer's phase32_programs= line, and a producer line with fizzbuzz removed is reported
result: pass
source: automated
coverage_id: D-10b
covered_by: cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit

### 21. D-00-D-08c-freeze-binding (32-09)
expected: check-phase32-freeze.sh rejects a mismatched headSha, a mismatched origin tip, a post-freeze digest-path change, a non-descendant and a malformed sha, and accepts a planning-only HEAD move with an unchanged codedigest; its path list equals phase32DigestPaths
result: pass
source: automated
coverage_id: D-00-D-08c-freeze-binding
covered_by: internal/compiler/session/session_phase32_freeze_test.go#TestPhase32FreezeGate

### 22. D-08b-ci-wiring (32-09)
expected: evidence-aggregate runs verify-phase32.sh after the Phase 27 upload and uploads phase32-receipt-os-sha with if-no-files-found: error, with an armed control
result: pass
source: automated
coverage_id: D-08b-ci-wiring
covered_by: internal/compiler/session/session_phase32_ci_test.go#TestPhase32CIHostMatrix

### 23. D-00-D-08c-run-bound-to-freeze (32-10)
expected: The dispatched run's headSha, the pushed HEAD and the origin tip are one commit with no digest-path change since FREEZE_COMMIT; a synthetic headSha is rejected with sha_mismatch
result: pass
source: automated
coverage_id: D-00-D-08c-run-bound-to-freeze
covered_by: internal/compiler/session/session_phase32_freeze_test.go#TestPhase32FreezeGate

### 24. SC3-EVD-15-receipt (32-11)
expected: The committed hosted macOS and Linux logs and 32-RECEIPT.json are accepted by the receipt gate at the frozen digest, and the living claims are enforced against the real receipt
result: pass
source: automated
coverage_id: SC3-EVD-15-receipt
covered_by: internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims, internal/compiler/session/session_phase32_gate_test.go#TestPhase32ReceiptGate

### 25. D-10a-bound-markers (32-11)
expected: The three living docs carry the bound marker at the receipt digest; a later code change must flip them to reopened
result: pass
source: automated
coverage_id: D-10a-bound-markers
covered_by: internal/compiler/session/session_phase32_receipt_test.go#TestPhase32ReceiptLivingClaims

### 26. D-08c-freeze-binding (32-11)
expected: The run headSha and the post-ingest HEAD are bound to the freeze; the .planning-only HEAD move is accepted and a synthetic sha is rejected
result: pass
source: automated
coverage_id: D-08c-freeze-binding
covered_by: internal/compiler/session/session_phase32_freeze_test.go#TestPhase32FreezeGate

### 27. D-11-coverage-groundedness (32-11)
expected: Every Phase 32 SUMMARY coverage block resolves
result: pass
source: automated
coverage_id: D-11-coverage-groundedness
covered_by: internal/compiler/session/summary_coverage_test.go#TestSummaryCoverageGroundedness

## Summary

total: 27
passed: 27
issues: 0
pending: 0
skipped: 0
blocked: 0

## Automation Evidence

All 27 deliverables were classified `mode: coverage`, `all_auto_covered: true` by `uat.classify-coverage`. On 2026-10-12 the agent re-ran all 21 distinct covering tests with `go test -count=1 -run` in ./cmd/schway, ./internal/compiler/session and ./internal/compiler/syntax: 21 PASS, 0 SKIP, 0 FAIL. Every test is pinned in recurring CI via `go test ./...` (.github/workflows/ci.yml:82). The hosted receipt run is 38315799626 (32-RECEIPT.json). Human checkpoints: 0. Nothing lacks a repeatable proxy (PROJECT.md § Verification Operating Preference).

## Gaps

[none]
