---
phase: 38-milestone-receipts
slug: milestone-receipts
status: blocked
scope: plan-01-preparation-only
threats_open: 3
asvs_level: 1
source_sha: c2da3cf17ea2f7ef663d540138fbbe2d7cdbb14f
codedigest: 3a05bab4400428771cf9f5ee6eb7f3f4d2cbafea58b75fdc5f66764b6cea7d82
created: "2026-10-16"
---

# Phase 38 — Pre-freeze Security

This is the independent Plan 01 preparation audit required before freeze. Tasks 01/02 are committed; task 03 and Plans 02/03 remain incomplete. This is not a phase-wide or hosted security sign-off. The typed auditor returned OPEN_THREATS; no risk is accepted.

## Trust Boundaries

| Boundary | Required property |
| --- | --- |
| Manifest/test events to producer | Exact reached cases, tests and non-skipped completion |
| Metadata/logs to binding checker | Bounded inert data, strict identities and exact source/run/attempt binding |
| Git/worktree to freeze | No staged, unstaged, untracked or intervening committed digest changes |
| Producer to hosted artifact | Failure remains failure; partial evidence cannot pass |

## Threat Register

| Threat | Severity | Disposition | Status | Independent finding |
| --- | --- | --- | --- | --- |
| T-38-01 | high | mitigate | open | Both event parsers reject real package start and package-level pass. M007 completion compares 106 rows against inherited count 39. Synthetic fixtures miss the package events. |
| T-38-02 | high | mitigate | open | Producer identity omits run ID/attempt, and binding does not compare log run/attempt to selected metadata. |
| T-38-03 | high | mitigate | closed | FreezeWorktreeControls passed clean/planning-only, staged/unstaged/untracked and change-then-revert controls; generic ancestry/history checks are invoked. |
| T-38-04 | high | mitigate | open | Artifact run_id/attempt and log artifact_id lack strict integer checks; Python bool/float equality can admit malformed identity values. |
| T-38-05 | high | mitigate | closed | Actual extracted CI pipeline refusal, pipefail, attempt-named artifact, always-upload and missing-file checks exist; inventory gaps remain T-38-01. |
| T-38-06 | medium | mitigate | closed | Go control has 600s timeout, 660s capture deadline and 16MiB output bound; metadata has 1MiB bound. No pressure probe performed. |
| T-38-SC | high | mitigate | closed | No new package/dependency installation; existing CI package setup predates preparation. |

Plans 02/03 threat IDs remain pending and outside this scoped audit. The separate code review also found missing preparation-record byte/source verification; that finding must be fixed before freeze even though this register audit does not construct new threats.

## Executed Evidence

The security auditor ran the exact Plan 01 Task 02 eight-root focused union in 10.922s; all eight roots passed without skips. It then ran `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase38ManifestConsumerBindings$' -count=1 -json -timeout=30s`, which passed and emitted the real package events rejected by both parsers. The green synthetic union does not close the findings.

All five reviewed source hashes matched `/tmp/schway-phase38-pre-freeze-source.json`. The user config retained SHA-256 `148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521` and remained unstaged. No full producer, project/whole-package suite, hosted action, pressure experiment, implementation edit or package installation was performed by the auditor.

## Accepted Risks

None.

## Security Audit Trail

| Date | Scope | Closed | Open | Auditor |
| --- | --- | --- | --- | --- |
| 2026-10-16 | Plan 01 preparation at c2da3cf1 | 4/7 | 3/7 | independent gsd-security-auditor |

## Sign-Off

Blocked. Fix the three high threats and the actionable code review findings, add reached controls, then obtain independent re-review before freeze. EVD-17 remains Pending. No push or dispatch is authorized.
