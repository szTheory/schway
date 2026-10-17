---
phase: 38-milestone-receipts
slug: milestone-receipts
status: verified
scope: plan-01-preparation-only
threats_open: 0
asvs_level: 1
source_sha: 0d5d51440e66e47c53f1703bef8b521eb78e389a
codedigest: b978317a7057bbd35a8ef0aff425bcd54dadb9ff364b28e0cac42d88f31466fd
created: "2026-10-16"
---
<!-- phase38-preparation/1 kind=security source_sha=0d5d51440e66e47c53f1703bef8b521eb78e389a codedigest=b978317a7057bbd35a8ef0aff425bcd54dadb9ff364b28e0cac42d88f31466fd disposition=secured -->

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

At the earlier production-identical 02cfd716 snapshot, the auditor independently ran the eight-root Task 02 focused union: PASS in 17.156s, all eight roots reached, no skips. That union and the standalone CRC probe are historical executions supporting unchanged production bytes. It included actual Go JSON capture through both parsers, identity aliases, run/attempt mismatches, preparation refusals and source-history controls. The standalone `GOCACHE=/tmp/ai-lang-verification-gocache sh scripts/verify-phase32.sh --milestone M007 --probe crc32/check` also passed. The earlier five hashes matched `/tmp/schway-phase38-pre-freeze-02cfd716-source.json`; config remained unchanged and unstaged. No full producer/full/race/whole-package suite, hosted run, pressure experiment or implementation edit was performed by the auditor.

Fresh delta audit at 0d5d5144: SECURED 7/7; `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase38ProducerInventoryControls$' -count=1 -v -timeout=60s` passed in 5.059s, all six children reached without skips. Both wrong-package mutations reached the real producer/binder refusals while preserving test run/pass identities. The four production files are unchanged from the prior secured review; all five current hashes match `/tmp/schway-phase38-pre-freeze-0d5d5144-source.json`. Code delta review is clean. Any later digest change reopens source binding before freeze. Plans 02/03 and EVD-17 remain pending. Source-bound preparation clearance is not final freeze readiness.
