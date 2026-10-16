---
phase: 38-milestone-receipts
slug: milestone-receipts
status: verified
scope: plan-01-preparation-only
threats_open: 0
asvs_level: 1
source_sha: 02cfd716b0d53b390b558f1d6bf9ed96098816a5
codedigest: d4f72f9a81ca85874b99c9505412b8710937ad2d23934afe8655e4b072c168cd
created: "2026-10-16"
---
<!-- phase38-preparation/1 kind=security source_sha=02cfd716b0d53b390b558f1d6bf9ed96098816a5 codedigest=d4f72f9a81ca85874b99c9505412b8710937ad2d23934afe8655e4b072c168cd disposition=secured -->

# Phase 38 — Pre-freeze Security

The independent gsd-security-auditor returned SECURED for the seven Plan 01 preparation threats at the exact source above. This does not clear future Plan 02/03 actions or establish hosted results. Original blocking audit bytes remain in the c2da3cf1 history file. No risk was accepted.

| Threat | Severity | Disposition | Status | Verified mitigation |
| --- | --- | --- | --- | --- |
| T-38-01 | high | mitigate | closed | Real package/test event handling, exact event joins, real producer capture, 106-case accounting and skipped/partial/duplicate refusals; 17 roots/480 events/39 inherited rows |
| T-38-02 | high | mitigate | closed | Workflow supplies actual run ID/attempt; producer emits both; binding compares exact log/source/run/attempt identities |
| T-38-03 | high | mitigate | closed | Committed ancestry/history, clean staged/unstaged/untracked digest paths, config invariant and bounded real preparation-file hash/source/disposition checks |
| T-38-04 | high | mitigate | closed | Bounded strict JSON, duplicate-key refusal, exact integer identities and boolean inputs; argv subprocess boundaries |
| T-38-05 | high | mitigate | closed | Real extracted pipefail failure control, attempt-named artifacts, always-upload and missing-file checks; diagnostic capture cannot complete a host receipt |
| T-38-06 | medium | mitigate | closed | 600s Go timeout, 660s capture deadline, 16MiB output bound, 60s package resolution and bounded metadata |
| T-38-SC | high | mitigate | closed | No new package installation or dependency change |

The auditor independently reran the eight-root Task 02 focused union: PASS in 17.156s, all eight roots reached, no skips. It included actual Go JSON capture through both parsers, identity aliases, run/attempt mismatches, preparation refusals and source-history controls. The standalone `GOCACHE=/tmp/ai-lang-verification-gocache sh scripts/verify-phase32.sh --milestone M007 --probe crc32/check` also passed. All five hashes matched `/tmp/schway-phase38-pre-freeze-02cfd716-source.json`; config remained unchanged and unstaged. No full producer/full/race/whole-package suite, hosted run, pressure experiment or implementation edit was performed by the auditor.

The code review separately requests explicit wrong-package rejection controls. Those tests and any subsequent digest change require refreshed source binding before the actual freeze. Plans 02/03 and EVD-17 remain pending. Source-bound preparation clearance is not final freeze readiness.
