---
phase: "33"
slug: "carried-debt-bounded-capture-read-and-generic-freeze-gate"
status: verified
threats_open: 0
asvs_level: 1
register_authored_at_plan_time: true
created: "2026-10-13"
---

# Phase 33 — Security

The level-1 audit verifies the plan-authored register against current source, focused controls, and the committed task receipts. No new hosted receipt is bound. Gates check consistency, not authenticity.

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| child process -> host | the native app child writes the evidence capture file the host then reads | source, capture, manifest, log, or evidence data |
| living docs -> CI | doc markers are policed by TestPhase32ReceiptLivingClaims against the computed digest | source, capture, manifest, log, or evidence data |
| interpreter model -> native oracle | interpreter acceptance decides whether a live resource at a defect is a leak or an abort the OS reclaims | source, capture, manifest, log, or evidence data |
| cgen -> emitted C | reason strings become C string literals in the application shell | source, capture, manifest, log, or evidence data |
| caller cwd -> gate | the gate may be invoked from any directory; pathspecs must resolve against the repo root | source, capture, manifest, log, or evidence data |
| receipt log files -> gate | logs downloaded from hosted runs are untrusted text parsed by sh | source, capture, manifest, log, or evidence data |
| manifest -> gate | repo-controlled, digest-bound data interpolated into greps and output | source, capture, manifest, log, or evidence data |
| repo scripts -> CI | ci.yml runs verify-phase32.sh; a stale caller of a deleted gate would fail open or break the hosted run | source, capture, manifest, log, or evidence data |
| goldens -> refusal coverage | goldens are the only record of old-gate behavior after deletion | source, capture, manifest, log, or evidence data |
| planning documents -> meta-tests | enforced documents must stay grounded in real tests | source, capture, manifest, log, or evidence data |
| retroactive docs -> reviewers | backfilled evidence must not read as new evidence | source, capture, manifest, log, or evidence data |

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation and evidence | Status |
|-----------|----------|-----------|----------|-------------|-------------------------|--------|
| T-33-01 | D | native_app.go capture read | high | mitigate | readApplicationCapture reads at most limit+1 bytes from the opened handle; TestPhase33CaptureReadBounded sparse-1gib subtest bounds the TotalAlloc delta above an in-limit baseline below 16*limit and the seeded unbounded read was shown to fail | closed |
| T-33-02 | T | capture size check | medium | mitigate | no Stat-by-path pre-check; the size decision is made on bytes read from the same handle (CWE-367 avoided); acceptance greps for Stat inside the helper | closed |
| T-33-03 | R | living docs markers | medium | mitigate | docs-only reopen commit lands first; reopened marker grammar verified by TestPhase32ReceiptLivingClaims; archived receipts byte-unchanged (D-02) | closed |
| T-33-SC | T | npm/pip/cargo installs | low | accept | no package installs in this plan; go.mod and go.sum unchanged | closed |
| T-33-04 | R | interp.go isAbortLikeDefect | high | mitigate | live set must equal [ownerOpID], op resolved in final.FunctionID only, Output must equal the kind reason; D-13 row and two-function falsifier refused; TestPhase33AbortLikeWideningMutantKilled proves both seams bite | closed |
| T-33-05 | T | cgen_program.go emitted C text | medium | mitigate | constants concatenated into the same bytes; TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity, TestPhase28PrunedEmitMatchesGolden gate any drift | closed |
| T-33-06 | T | seam vars left widened | low | mitigate | seams unexported, swapped only in same-package tests and restored with defer; production defaults asserted after restore | closed |
| T-33-07 | S | check-freeze.sh / check-receipts.sh cwd handling (WR-01) | high | mitigate | resolve git toplevel, cd, reject no_repo_root; absolute read copies of MANIFEST and logs; TestFreezeGate/subdirectory and m006-real-repo-from-cmd rows reject the change the old gate falsely accepted | closed |
| T-33-08 | T | manifest fields in sh | medium | mitigate | label only as a printf %s argument, matched with grep -F or case patterns, no eval; field() style parsing kept | closed |
| T-33-09 | S | receipt digest source | high | mitigate | EXPECTED_CODEDIGEST never recomputed when set, digest_source printed; TestReceiptGateArchivedM006 digest-mismatch control | closed |
| T-33-10 | R | gate failure output | medium | mitigate | exactly one stderr line: codedigest --changed stderr discarded and mapped to git_log_failed; differential asserts byte-equal stderr | closed |
| T-33-11 | R | deletion of the Phase 32 gate scripts | high | mitigate | goldens captured from old-script output and checked equal to the new scripts in the same task before git rm; each golden carries a literal reason read by a human (RESEARCH Pitfall 7) | closed |
| T-33-12 | T | stale callers of the retired names | medium | mitigate | TestNoCheckPhase32Caller scans tracked non-.planning files with a concatenated needle and checks the files are absent, with an armed seeded-caller control | closed |
| T-33-13 | T | producer/manifest drift | medium | mitigate | one table: producer reads the manifest; TestPhase32ProducerCaseTable pins the manifest sha256 and grounds rows in both archived logs; TestPhase32PublicRouteAudit reads the same programs line | closed |
| T-33-14 | R | 28-VALIDATION.md / 28-SECURITY.md | medium | mitigate | "Reconstructed (State B)" and "retroactive, no new tests" stated; every command row grounded by TestVerificationGroundedness; Hosted CI not-run fact stated | closed |
| T-33-15 | R | 33-CARRIED-DEBT-DISPOSITION.md | medium | mitigate | each fixed/covered row names a test verified to exist with git grep; no row closes debt without a witness | closed |
| T-33-16 | S | living-claim check for future milestones | medium | mitigate | marker prefix built with regexp.QuoteMeta from the spec; other-label markers ignored by the M006 spec; M007 stale_bound_claim row armed | closed |

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---------|------------|-----------|-------------|------|
| AR-33-SC | T-33-SC | No installs or dependency changes; POSIX sh and the Go standard library retain the planned boundary. | Plan-authored disposition | 2026-10-13 |

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-13 | 17 | 17 | 0 | execute-phase level-1 source audit |

## Source Verification

`readApplicationCapture` reads from one opened handle through `io.LimitReader(limit+1)`; no path-based Stat pre-check occurs. Both routed interpreter callers pass the acquire operation ID; the final event function, singleton live-owner set, and exact per-kind reason are checked. The mutation seams are unexported and restored with defer in the tests.

Both gates anchor the repository root. Manifest labels appear as data arguments, with fixed-prefix comparisons and no eval. The receipt gate preserves an explicit EXPECTED_CODEDIGEST. The fixed goldens, retired-caller scan, shared manifest producer, and milestone-quoted regexes are present and exercised by the focused checks. Retroactive phase 28 documents state their evidence limits. Archived receipt/log files and dependency manifests remain unchanged.

## Sign-Off

- [x] All threats have a plan-authored disposition.
- [x] Mitigation patterns and named test witnesses inspected.
- [x] Accepted risk retains the original limited scope.
- [x] `threats_open: 0` confirmed.
- [x] `status: verified` set.

**Approval:** verified 2026-10-13

## Review Follow-Up — 2026-10-13

CR-01 exposed a malformed-control-list acceptance gap within T-33-08. A delimiter-only or whitespace-only list accepted arbitrary control rows before the fix; an empty list was already refused downstream. The gate now refuses missing and empty entries before inspecting logs. `TestReceiptGateMalformedControls` preserves six regression cases; the broader focused compatibility run passed (28.712s). No open blocking threat remains.

## Successor Refresh — 2026-10-13

The L1 authored register was rechecked after the Phase 34 successor changes. The security implementation boundaries are byte-unchanged since the prior clean review, and current source inspection plus the fresh 47-name focused controls support all 17 existing closures. `threats_open: 0` remains; this uses the L1 short-circuit without a new threat-discovery pass. See `33-REFRESH-AUDIT.md` for scope and execution details.

## Phase 35 successor refresh — 2026-10-14

All original plans remain complete; no implementation or UAT is replayed. The
Phase 35 final inherited regression refresh passed 94 distinct names / 98
package-qualified executions across eleven packages with no missing/fail/skip,
including this phase's mapped selectors. See `../35-bitwise-operators-literal-shifts-and-hex-literals/35-REGRESSION.md`.
This refresh labels the original execution receipts historical and recomputes
the current verification fingerprint through its canonical owner.

L1 authored-register review retains every closed mitigation: bounded native capture and root/digest shell gates are byte-unchanged; source presentation and byte identity/repair authority are byte-unchanged. Shift/recovery/manifest additions have passing named independent controls and the current Phase 35 review. No new installation or accepted-risk substitution. The configured L1 zero-open short-circuit applies; this is a current source/control applicability review, not a new hosted security run.


## Phase 37 successor L1 refresh — 2026-10-15

All five plans contain authored threat models; the register still has
17 closed dispositions, zero open, and ASVS level 1. Current source inspection
confirmed one opened-handle `LimitReader(limit+1)` capture read; singleton live
owner, terminal function/operation and exact reason matching; and root-anchored
POSIX gates preserving explicit EXPECTED_CODEDIGEST with data-only labels.
The successor's capture-disabled terminal path delegates to the same narrowed
owner/function/operation/reason predicate. All mapped capture, parity, narrowing,
widening-mutant, freeze/receipt and metadata controls passed in the fresh
47-root validation run. This refresh added only test/metadata corrections,
with no production or dependency change. The authored-register L1 short-circuit
applies; no new deep threat-discovery, hosted certification, or authenticity
guarantee is claimed. See 33-PHASE37-REFRESH-RUN.json for current parent-run
evidence, distinct from the independent goal verdict.

## Security Audit 2026-10-15

| Metric | Count |
|---|---|
| Threats found | 17 |
| Closed | 17 |
| Open | 0 |

## Review boundary follow-up — 2026-10-15

The earlier 47-root snapshot predates the two independent review fixes.
WR-02 adds a bounded literal program/case inventory check in the POSIX receipt
gate under existing T-33-08/T-33-13; default and archived M006 acceptance and
the fixed refusal catalog remain covered. WR-01 strengthens the coverage
metadata guard under T-33-14/T-33-15. Reached false-accept controls failed on
the original implementations and pass after repair. Both use existing
standard-library/POSIX facilities; no authority, dependency or hosted binding
is added. The L1 source review retains 17 closed authored dispositions and
zero open threats. The final verifier must include these changed inputs.

## Closure record-shape follow-up — 2026-10-15

The review's adjacent WR-03/WR-04 cases remain within the same authored
manifest/metadata boundaries. Case-record validation uses an explicit reason
so empty program data cannot mean no mismatch, and preserves the producer's
free-text stderr tail. Coverage metadata now checks required entry fields and
local canonical scalar/list shapes as well as per-verification fields. The
fixes add no dependency or new trust boundary; all 17 authored L1 dispositions
remain closed. Final source-bound controls and the independent goal verifier
must cover these latest inputs, not reuse the preceding snapshots as current.
