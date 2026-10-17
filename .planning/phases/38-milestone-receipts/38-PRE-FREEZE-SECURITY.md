---
phase: 38-milestone-receipts
slug: milestone-receipts
status: verified
scope: plan-01-preparation-and-implemented-final-gates
threats_open: 0
asvs_level: 1
source_sha: 06e72dda0addd2744106e22ef35fec539337999b
codedigest: f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf
created: "2026-10-16"
---
<!-- phase38-preparation/1 kind=security source_sha=06e72dda0addd2744106e22ef35fec539337999b codedigest=f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf disposition=secured -->

# Phase 38 — Pre-freeze Security

The independent gsd-security-auditor returned SECURED, 7/7 threats closed, at this exact source. High/critical findings block under ASVS1. All dispositions are mitigate; no risk was accepted. This is preparation and implemented-gate clearance; hosted retrieval and final retained-byte execution remain pending.

| Threat | Severity | Status | Verified mitigation |
| --- | --- | --- | --- |
| T-38-01 | high | closed | Producer, manifest, exact-event inventory unchanged; prior source-bound consumer and wrong-package controls apply |
| T-38-02 | high | closed | Exact run/log/attempt/job/artifact identity, actual candidate HEAD and ingest ancestry |
| T-38-03 | high | closed | Config protected in freeze commit and subsequent history; all-parent merge-aware digest and planning scope; current bytes/index/worktree checked |
| T-38-04 | high | closed | Strict bounded JSON, duplicate/type/integer checks, argv boundaries, NUL Git paths and validated parent SHAs |
| T-38-05 | high | closed | CI pipefail/upload transport unchanged; new final refusals require successful receipt validation first |
| T-38-06 | medium | closed | Existing output/capture limits unchanged; only bounded local roots executed |
| T-38-SC | high | closed | No installation or dependency addition |

Fresh independent execution: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase38FreezeCommitConfigControls$' -count=1 -v -timeout=60s` passed in 1.623s. Ordinary and two-parent merge freeze controls reached without skips; both freeze and final mode refused config_history_changed. All five hashes match /tmp/schway-phase38-pre-freeze-06e72dda-source.json; workflow, producer and manifest are unchanged. User config SHA remains 148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521, unstaged.

Historical independent checks: 02cfd716 eight-root union 17.156s and CRC probe; 0d5d5144 wrong-package controls 5.059s; efff39f9 freeze/final union 7.824s; 991abac4 freeze/final union 12.709s with 28 children. The efff/991 verdicts were superseded for newly identified history and freeze-boundary defects, now repaired and independently rechecked. These earlier executions were not rerun or relabeled current. Original reports remain in named history artifacts. The auditor wrote no files; parent persists this structured verdict under the single-writer contract.

No unregistered flags within scope. Actual freeze selection, authenticated hosted evidence, retained-byte final execution and EVD-17 remain pending.
