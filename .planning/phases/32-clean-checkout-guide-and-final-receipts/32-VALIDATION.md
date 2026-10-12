---
phase: "32"
slug: "clean-checkout-guide-and-final-receipts"
# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (closure plan / validate-phase)
status: planned
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-12"
---

# Phase 32 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> The source of truth for each test is `32-RESEARCH.md` § Validation Architecture.
> Zero human UAT. The only human step is the single push-and-dispatch line in the checkpoint plan (PROJECT.md § Verification Operating Preference).

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test (Go 1.24 stdlib `testing`). The guide doc-test needs clang. A missing clang skips locally and fails on hosted CI (native policy). |
| **Config file** | none (go.mod only) |
| **Quick run command** | Run each touched package (syntax, session, cmd/schway) with `-run 'Phase32' -count=1`. |
| **Session meta set** | `go test ./internal/compiler/session -run 'Inventory\|Groundedness\|Reconciliation\|ValidationLifecycle\|TestPhase28EntryExitClassManifest\|TestPhase272AcceptedCorpusByteIdentity\|TestPhase273ConformanceCorpusByteIdentity\|TestPhase30PreChangeGolden\|TestLanguageMaturityCountsAreCurrent\|TestExitCodeDocsTable\|TestNoSuppressionOutlivesItsWitness\|TestPhase16PublicEmitterConsumerInventory' -count=1` |
| **Full suite command** | Hosted CI only: `go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest |
| **Estimated runtime** | About 5–90 s per focused package run. The guide doc-test builds `schway` and three native apps. |

---

## Sampling Rate

- **After every task commit:** run the focused `-run 'Phase32'` command for each package the task touched. Also run the session meta set whenever the task touches any of:
  - a `.schway` file;
  - a golden or allow-list;
  - a counted planning doc.
- **After every plan wave:** run the quick command and the session meta set.
- **Before the receipt checkpoint:** the full local `go test ./...` must be green, and the tree must be clean at the freeze commit.
- **Max feedback latency:** 90 s per focused run.

---

## Per-Task Verification Map

There is one row per success criterion or decision. Test names come from `32-RESEARCH.md`, and the planner may refine them. Every gate has an armed control that has been seen to fail.

| Ref | Requirement | Behavior | Test | Armed control | Status |
|-----|-------------|----------|------|---------------|--------|
| D-10 | (todo, process) | A `var` body ending in a bare identifier, number or `)` round-trips: reparse, fixed point, same tokens. Every other corpus file formats unchanged. | `TestPhase32FormatVarBody` | the pre-fix formatter output (`r = r + ar`) is rejected | ⬜ pending |
| D-10 / D-06 | (todo) | The Phase 30 golden holds with a dated format-change allow-list | `TestPhase30PreChangeGolden` | a stale allow-list entry fails | ⬜ pending |
| SC1 / D-01..D-03 | DX-19 | The README's marked blocks run in a tracked-file temp copy. Each row matches exit, stdout and stderr. Error rows hide the path. Outputs go only under `.schway/`. | `TestPhase32GuideCleanCheckout` | a mutated guide value fails | ⬜ pending |
| SC1 / D-04 | DX-19 | Anti-vacuity: block count, all three programs present, row count | `TestPhase32GuideParserNegativeControls` | a changed expected value fails; a renamed command fails | ⬜ pending |
| SC1 / D-01 | DX-19 | The guide's limit, case names and outputs agree with `checksum.expected.json` and the oracle rows | `TestPhase32GuideMatchesExpectedJSON` | a mutated JSON value fails | ⬜ pending |
| SC1 / D-04 | DX-19 | Hosted + missing clang fails; local + missing clang skips | `TestPhase32NativePolicy` | the policy function returns skip on hosted | ⬜ pending |
| SC2 / D-05 | DX-19 | The old path is absent and the replacement exists. No renamed copy by tombstone digest. Live-doc hygiene holds. | `TestPhase32ChecksumFrontierRetired` | old bytes in a tempdir are flagged | ⬜ pending |
| SC2 / D-06 | DX-19 | A retired path may be absent from the frozen Phase 28 and 30 goldens; a retired path observed again fails | `TestPhase28EntryExitClassManifest`, `TestPhase30PreChangeGoldenIsNotInert` | a retired path present again fails | ⬜ pending |
| SC2 / D-07 | DX-19 | LANGUAGE-MATURITY counts are re-derived | `TestLanguageMaturityCountsAreCurrent` | the existing count control | ⬜ pending |
| SC3 / D-08 | EVD-15 | The sh and Go codedigests agree. The digest moves on a code edit, but not on a `.md` or `.planning` edit. | `TestPhase32CodeDigest` | a mutated code file changes the digest | ⬜ pending |
| SC3 / D-09 | EVD-15 | The gate accepts a good pair and rejects digest, dirty-tree, missing-host and missing-program cases | `TestPhase32ReceiptGate` | each named mutation is rejected | ⬜ pending |
| SC3 / D-10a | EVD-15 | No receipt and no claim passes. A claim without a receipt fails. A current claim with a stale digest fails. A "reopened" claim passes. | `TestPhase32ReceiptLivingClaims` | a stale-digest current claim fails | ⬜ pending |
| SC3 / D-10b | EVD-15 | Every public route and entry program has a covered row or a "not covered because" row. The receipt program set equals the audit set. | `TestPhase32PublicRouteAudit` | an injected new route fails | ⬜ pending |
| SC3 / D-00 / D-08c | EVD-15 | Revision = codedigest (D-08). The run headSha, the pushed HEAD and origin are one commit with no digest-path change since the freeze; a `.planning`-only HEAD move is accepted | `TestPhase32FreezeGate` (`scripts/check-phase32-freeze.sh`) | a synthetic mismatched headSha, a mismatched origin and a post-freeze digest-path change are each rejected | ⬜ pending |
| SC3 | EVD-15 | The hosted macOS and Linux logs are bound to the final digest | `sh scripts/check-phase32-receipts.sh` at ingest, plus `TestPhase32ReceiptLivingClaims` | a digest mismatch is rejected | ⬜ pending |
| Process | (all) | Every Phase 32 SUMMARY has a resolvable `coverage:` block | `TestSummaryCoverageGroundedness` | the seeded bad controls | ⬜ pending |

---

## Wave 0 Requirements

- `internal/compiler/syntax/syntax_phase32_format_test.go`
- `cmd/schway/main_phase32_guide_test.go`
- `cmd/schway/main_phase32_routes_test.go`
- `internal/compiler/session/session_phase32_retire_test.go`
- `internal/compiler/session/session_phase32_receipt_test.go`
- `internal/compiler/session/session_phase32_freeze_test.go`
- `scripts/codedigest.sh`, `scripts/verify-phase32.sh`, `scripts/check-phase32-receipts.sh`, `scripts/check-phase32-freeze.sh`

No framework install is needed.

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Instructions |
|----------|-------------|------------|--------------|
| Push the frozen revision and dispatch hosted CI | EVD-15 | The auto-mode classifier blocks push and dispatch for the agent, so it needs explicit user approval | Paste `! git push -q origin phase26-validation && gh workflow run ci.yml --ref phase26-validation` |

Everything else is automated.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without an automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 90 s
- [ ] `nyquist_compliant: true` set in frontmatter at closure

**Approval:** pending
