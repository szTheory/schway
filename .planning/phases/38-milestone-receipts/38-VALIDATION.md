---
phase: "38"
slug: "milestone-receipts"
status: validated
nyquist_compliant: false
wave_0_complete: true
created: "2026-10-16"
---
<!-- phase38-preparation/1 kind=validation source_sha=06e72dda0addd2744106e22ef35fec539337999b codedigest=f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf disposition=focused -->

# Phase 38 — Validation Strategy

This contract has been reconciled against executed Plan 01 preparation. The eight new roots exist and their focused positive/negative controls passed. Hosted execution, actual remote approval and final ingest remain pending, so nyquist_compliant stays false and no EVD-17 completion is claimed. The original planning-only assertions below retain their dated scope; the execution amendment records the current evidence.

## Test Infrastructure

| Property | Value |
| --- | --- |
| Framework | Go 1.24 test, existing shell gates and hosted GitHub Actions |
| Config | go.mod; .github/workflows/ci.yml |
| Existing focused baseline | `TestFreezeGateDifferential`, `TestReceiptGateDifferential`, `TestPhase32CodeDigest` in the tracked session test package; included once in command B below |
| Local policy | Exact bounded selectors; no full local suite, whole-package sweep, broad native matrix or pressure capture |
| Hosted policy | Required checks and evidence aggregate on Ubuntu/macOS; existing full and race lanes plus complete M007 producer inventory |
| Runtime | Measure actual focused runs; do not invent a latency estimate |

## Sampling Rate

After each implementation task, run its focused behavioral controls and verify expected non-skipped roots/subtests were reached. After a preparation wave, run the union of changed acceptance seams once. Before freeze, complete relevant review/security/freshness and the new audit. After ingest, re-gate copied logs/provenance and the actual planning-only descendant. Full/race suites run hosted; a local observation cannot replace either host.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat refs / acceptance | Test type | Command/evidence | Exists | Status |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 38-01-01 | 01 | 1 | EVD-17 prerequisite | T-38-01/02/04/05: real CRC path, partial-probe refusal, wrong-revision controls | tracer seam | A | created and exercised | pass: bounded tracer |
| 38-01-02 | 01 | 1 | EVD-17 prerequisite | T-38-01..06: exact consumer inventory, skip/empty/duplicate/precision/history/pipeline refusals | focused seam | B | created and exercised | pass: focused acceptance controls |
| 38-01-03 | 01 | 1 | EVD-17 prerequisite | T-38-03: source-bound review/security/freshness/new audit before actual freeze | lifecycle + gate | C; independent pre-freeze review/security and fresh milestone audit | gate/schema/records created by Plan01 | audit/review passed; actual freeze preflight follows metadata commit |
| 38-02-01 | 02 | 2 | EVD-17 prerequisite | T-38-07: actual full freeze/repo/ref/workflow/input/push-scope human approval | authority checkpoint | C establishes readiness only; actual explicit user response is mandatory | future concrete freeze/approval | pending; cannot autoapprove |
| 38-02-02 | 02 | 2 | EVD-17 prerequisite | T-38-08..11: authenticated one-attempt head/job/step/artifact/log identity; one transient retry max | hosted integration | D + authenticated API/archive digest observations | source seam from01; actual run unknown | pending external execution |
| 38-03-01 | 03 | 3 | EVD-17 final acceptance | T-38-12/14/15: exact copied bytes, run/attempt/source/ingest scope | copied-byte gate | E | source seam from01; retained bytes unknown | pending execution |
| 38-03-02 | 03 | 3 | EVD-17 final acceptance | T-38-16: bounded claims, current consumer/seed transition | gate + claim review | E; transparent source/receipt/claim links and explicit flagged prohibition | future transition/bytes | pending execution |
| 38-03-03 | 03 | 3 | EVD-17 completion | T-38-13/14: actual postcommit HEAD and planning-only ingest; no premature completion | lifecycle + gate | E after actual last planning commit, repeat after later lifecycle commits | future final descendant | pending execution |

Commands assume repository root. Every automated command in the plans is immediately followed by an observable `fails_when`; missing PASS/success markers, no-tests-to-run, skips or nonzero exit fail. A synthetic seam fixture never becomes a hosted receipt.

**A — real bounded tracer:** `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase38ReceiptTracer$' -count=1 -v -timeout=60s`

**B — acceptance union:** `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase38ProducerInventoryControls|TestPhase38ManifestConsumerBindings|TestPhase38BindingControls|TestPhase38FreezeWorktreeControls|TestPhase38FinalCandidateControls|TestPhase38FreezeCommitConfigControls|TestPhase38WorkflowFailureControls|TestFreezeGateDifferential|TestReceiptGateDifferential|TestPhase32CodeDigest)$' -count=1 -v -timeout=120s`

**C — actual freeze preflight:** `sh scripts/check-milestone-binding.sh freeze .planning/phases/38-milestone-receipts/38-FREEZE.json`

**D — observed staged bytes:** `sh scripts/check-milestone-binding.sh run .planning/phases/38-milestone-receipts/38-FREEZE.json "$PHASE38_STAGE/run.json" "$PHASE38_STAGE/receipt-ubuntu-latest.log" "$PHASE38_STAGE/receipt-macos-latest.log"` where PHASE38_STAGE is the actual observed private `/tmp/schway-phase38-RUNID-ATTEMPT` directory recorded in HOST-OBSERVATION.

**E — retained bytes and actual descendant:** `sh scripts/check-milestone-binding.sh final .planning/phases/38-milestone-receipts/38-FREEZE.json .planning/phases/38-milestone-receipts/38-HOST-RUN.json .planning/phases/38-milestone-receipts/38-HOST-UBUNTU.log .planning/phases/38-milestone-receipts/38-HOST-MACOS.log`

E delegates existing generic receipt/freeze gates, checks current committed digest and clean worktree, validates metadata/log identities and hashes, and enforces all ingest changes since retained ingest_base_sha are under .planning. Its success includes the actual full candidate HEAD. It must run again after the final ingest/summary/lifecycle commit. No anticipated/self-referential final SHA is substituted.

## Wave 0 Requirements

- [x] `internal/compiler/session/session_phase38_receipts_test.go`: TestPhase38ReceiptTracer exercises real `--milestone M007 --probe crc32/check` interpreter/native tuple and partial-probe refusal.
- [x] TestPhase38ProducerInventoryControls and TestPhase38ManifestConsumerBindings exercise real producer parsing plus exact fixture/consumer/subtest joins, missing/duplicate/wrong-result/empty-selector/skip/partial inventory and unavailable-tool refusals.
- [x] TestPhase38BindingControls proves wrong freeze/run/log SHA, mixed attempts, duplicate hosts, missing/failed jobs/steps/artifacts, altered bytes, malformed metadata and non-integer/overflow identity refusals; both host orders pass.
- [x] TestPhase38FreezeWorktreeControls proves staged/unstaged/untracked digest edits and change-then-revert history refuse while a true planning-only descendant passes; config invariant remains fixed.
- [x] TestPhase38WorkflowFailureControls executes the extracted real Bash producer/tee failure seam and verifies attempt-named artifacts, always-upload and missing-file failure.
- [x] No new framework/dependency installation.

All eight new roots assert expected non-skipped children internally; the focused execution and independent reviews below verify that obligation. Reuse existing generic gate refusal goldens; do not clone their implementation. The driver reuses exact existing CRC matrix, helper, bitwise and scoped Phase37.1 consumers. Five native/code preparation files form one safety obligation, then three dependent plans preserve the locked freeze → authority → planning-only ingest sequence.

## Revision 1 bookkeeping dispositions

- Research questions1..3 are explicitly RESOLVED in RESEARCH: exact freeze/run JSON files/schemas and the narrow checker/test owner are selected; unobserved host environment/access are resolved by fail-closed execution gates, not availability claims. Any absent required result or approval leaves EVD-17 Pending.
- Scope counts are explicit: Plan01 has five production/fixture/CI paths plus five bounded review/security/validation/freeze/audit outputs, with at most five paths per task. Plan03 has twelve planning-only paths divided4/3/5 across retained evidence, mandated transition claims and canonical lifecycle tasks. D-08/D-13/D-15 and AGENTS supply the named safety/output obligations; no extra plan, subsystem or broad local execution is added.
- T-38-01..T-38-16 remain unique. T-38-SC is one reserved shared phase-level package-install obligation repeated identically in all plans under the canonical exemption, not three independent threats; no package install is planned.
- PATTERNS now classifies the actual new `internal/compiler/session/session_phase38_receipts_test.go` and names tracked `session_gate_differential_test.go` only as its existing analog/helper owner. No source/test file is changed by this planning revision.

## Hosted claim registry

The explicit M007 manifest preserves all seven programs/39 M006 rows verbatim. New CRC cases are bound to `TestPhase37PublicFileMatrix/CASE/CONSUMER/ENGINE`; NUL is `.../nul-direct-dispatch/CONSUMER/ENGINE/direct-public-dispatch`. All 15 helper rows bind to the actual selected `TestPhase36HelperAnswersAndDefects/CASE` consumers, including the five single-step inline-source rows; fixture source labels are not executable entry selectors. Public unused-Byte helper and bitwise witnesses use their structured JSON consumers, not decimal-output assumptions. Optimization evidence comes from the named O0/O3/O3-LTO subtests under TestPhase36HelperOptimizationMatrix, TestPhase35BitwiseOptimizationMatrix and the scoped terminal tests, separate from default process tuples. Physical cleanup claims retain their named existing consumer boundary.

Both hosts require the existing checks jobs (full and race steps), existing current evidence aggregate jobs, complete reached case/control inventory, native tools and exact attempt artifacts. Hosted-only broad work must never be moved into local preflight. Missing environment-dependent FIFO/device/permission/symlink evidence is a blocker, even if the test process exits0.

## Planning hook dispositions

All planner contributions in `/tmp/schway-phase38-plan-pre-hooks.json` were applied inline by semantics. Canonical API coverage detector was run on the actual Phase38 roadmap section plus all three plan bodies: `detected:false`, `signals:[]`, not skipped. Thus no API coverage checkpoint/matrix is asserted or fabricated. The phase uses bounded hosted run/job/artifact observation; it does not add an external API product. If later scope changes cause the canonical detector to return true, produce COVERAGE.md with explicit capability decisions before sealing; skipped payloads remain skips, never false verdicts.

Canonical assumption-delta detector returned `detected:true`, pluralization `second` from the existing CRC importer. Plan01 records primary noun=manifest-backed executed program/case/control identity, decision=no-change, because plural programs/consumers already exist. No schema/ORM path occurs in scope, CONTEXT or RESEARCH, so the schema-push contribution adds no database task. Security uses configured ASVS1 and blocks high/critical. Codebase-drift precheck skipped for no STRUCTURE.md; context-drift returned no stale artifacts. Canonical `check ui.plan-gate 38` returned frontend:false, block:false, hasFrontendEvidence:false, hasUiSpec:false. No package install, AI integration or UI artifact is needed. Research and pattern-mapper outputs were consumed. Graph context is disabled, explicitly reported by graphify status; none was invented.

## Four-source coverage audit

| Source | ID | Required outcome/constraint | Plan/task | Status |
|---|---|---|---|---|
| GOAL | Phase38 | Every native M007/earlier pinned claim bound once after audit to final-source two-host evidence | 01 preparation;02 hosted;03 final binding | COVERED |
| REQ | EVD-17 | Same-source two-host receipts in last planning-only plan after audit | all plans;completion only38-03-03 | COVERED |
| RESEARCH | responsibility map | Driver owns inventory; generic gates own consistency; hosted source provenance separate | 38-01-01/02;38-02-02 | COVERED |
| RESEARCH | producer/fixture routes | Preserve M006; consume CRC matrix, helper source selections and structured outputs correctly | 38-01-01/02 | COVERED |
| RESEARCH | skips/pipeline/tools | Reject non-reached/partial evidence and tee-masked failures; no native skip | 38-01-02;38-02-02 | COVERED |
| RESEARCH | actual freeze/worktree | Committed preparation + review/security/freshness/new audit; clean digest worktree/history | 38-01-03;38-03-03 | COVERED |
| RESEARCH | bounded provenance | Full run/attempt/job/artifact/log identities and raw durable bytes | 38-02-02;38-03-01 | COVERED |
| RESEARCH | authority/retry | Read-only inspection allowed; explicit actual-write approval; one transient same-freeze retry | 38-02-01/02 | COVERED |
| RESEARCH | no extra platform | Standard library/installed tools; no package/backend/framework/signing/hermetic claim | 01 interfaces;03 claim boundary | COVERED |
| CONTEXT | D-01 | Explicit inherited/new program/case/control manifest | 38-01-01/02 | COVERED |
| CONTEXT | D-02 | Maintained opt-in driver, unchanged M006 default/generic gates | 38-01-01/02 | COVERED |
| CONTEXT | D-03 | CRC oracle/matrix/boundaries/opaque paths/filesystems; NUL direct route | 38-01-02 | COVERED |
| CONTEXT | D-04 | All helper rows, exact selected entries, structured helper/bitwise contracts/refusals | 38-01-02 | COVERED |
| CONTEXT | D-05 | Focused named Phase37.1 terminal/reason/compatibility controls | 38-01-02 hosted registry | COVERED |
| CONTEXT | D-06 | Every required root/subtest/case reached; no skipped/partial acceptance | 38-01-02;38-02-02 | COVERED |
| CONTEXT | D-07 | Existing hosted checks/full/race/aggregate, distinct optimization claims; bounded local checks | 38-01-01/02;38-02-02 | COVERED |
| CONTEXT | D-08 | Preparation committed/reviewed/verified/audited before actual freeze | 38-01-03 | COVERED |
| CONTEXT | D-09 | Committed history + worktree clean; preserve config; actual final descendant | 38-01-02/03;38-03-01/03 | COVERED |
| CONTEXT | D-10 | Actual freeze-specific non-autoapproved remote scope | 38-01-03;38-02-01 | COVERED |
| CONTEXT | D-11 | Ref/run/log exact freeze SHA equality + armed wrong-revision control | 38-01-01/02;38-02-02 | COVERED |
| CONTEXT | D-12 | One wholly successful run/attempt; bounded jobs/artifacts/raw identities | 38-01-02;38-02-02;38-03-01 | COVERED |
| CONTEXT | D-13 | Durable copied raw logs/metadata; re-gate copied bytes; historical failures/private field limits | 38-03-01 | COVERED |
| CONTEXT | D-14 | One transient retry; code fixes reopen freeze/checkpoint; blockers keep pending | 38-02-02;38-03-03 | COVERED |
| CONTEXT | D-15 | Exactly3 dependent plans; only final planning-only ingest completes EVD-17 | 01/02 empty completion arrays;38-03-03 | COVERED |

The five explicit edge rows are all lifted into must_haves truths (boundary/ordering/precision in01;adjacency/empty in02) and real gate controls. Prohibition precision dropped routine edge-covered candidates and canon security concerns (referred to security review); one bespoke scope-transparency item remains descriptor-less via projectProhibitions in03, flagged-unverified rather than auto-dismissed or assigned invented checks. Deferred language/features/seed experiments are exclusions, not missing phase coverage.

## Transition check at planning

Source inspection agrees with the living documents' latest Phase37.1 completion amendments: CR-03/04/06 closed, EVD-17 Pending, no current two-host freeze bound. Existing QueryCommandFile/search is the z1g witness/CRC-helper reviewer consumer; readApplicationCapture/replay is the trz witness/CRC-checksum diagnosis consumer; syntax.Format and checked CRC/helper sources are the z1s witness/actual reviewer consumer. All three seeds retain dormant status: no measured navigation gap, unexplained failure or contested source form was found. The Dashbit appraisal remains historical. Context38's complete three-capability table remains current: same-source audited CRC/checksum/earlier receipts (Phase38 owner), proposed wc/longest_line module pair (next-milestone contract owner), then fixed-schema key=value checker (future schema owner). Plan03 rechecks those triggers and records the changed receipt facts; no deferred feature is implemented here.

## Manual-Only Verification

Only authorization to push the named freeze and dispatch the named workflow requires the user. The checkpoint must name the full actual commit/digest, repository/ref, workflow/inputs and push scope. All objectively checkable behavior, raw bytes and hosted provenance are agent-verified. Missing external access/results leave EVD-17 pending; they do not create subjective UAT.

## Validation Sign-Off

- [x] Every planned task has an exact automated command immediately followed by observable failing directions; the authority checkpoint additionally requires actual human approval.
- [ ] No three consecutive tasks lack objective automated evidence.
- [x] Wave 0 controls exercise real producer/parser and binding seams, including reached review defects.
- [ ] No watch mode, hidden broad local suite or skipped required case accepted.
- [ ] Both hosts and required full/race/evidence lanes pass at the selected freeze.
- [ ] Copied logs and metadata pass binding and final drift checks.
- [ ] `nyquist_compliant: true` only after execution evidence establishes it.

**Approval:** preparation validated; hosted/authority/ingest obligations pending. This is partial phase validation, not phase completion.

## Preparation validation amendment — 2026-10-16

Source: `0d5d51440e66e47c53f1703bef8b521eb78e389a`; codedigest `b978317a7057bbd35a8ef0aff425bcd54dadb9ff364b28e0cac42d88f31466fd`. The actual manifest has 106 distinct process rows (39 inherited, 52 CRC and 15 helper), 17 control roots and 480 exact event names. This is a declaration of required hosted coverage, not a claim that the complete producer ran locally.

The real CRC tracer, producer inventory/consumer joins, binding precision/provenance, source/history/preparation records and CI failure controls passed. Independent security ran the eight-root acceptance union at the production-identical 02cfd716 snapshot (17.156s, no skips) and fresh wrong-package real-event controls at 0d5d5144 (5.059s, six children, no skips); code review is clean after the same test delta. Phase 33's 33-38-REFRESH-RUN.json records a further current-source ten-root receipt union (117 run/pass events), tracer (6), and isolated groundedness guard (1), all without failure/skip. The Phase 33 report is fresh/passed 60/60. Previous broad or capture-pressure receipts remain historical and were not rerun.

Independent review caught package lifecycle/selector mismatches, missing codedigest descendants, 39-versus-106 accounting, unchecked preparation bytes, identity type aliases and missing log attempt fields. Reached red/green and wrong-package controls now cover those seams. Original review/security bytes remain in named history artifacts. The generic gates remain consistency checks; future authenticated GitHub observations establish provenance separately.

The first six-document-guard run failed only because this validation document still had status:draft. It did not report a behavior failure. This reconciliation sets status:validated and wave_0_complete:true while preserving nyquist_compliant:false and all external/ingest tasks pending. Re-run the documentation union before freezing. No full local suite, complete producer, race lane, broad native matrix or pressure experiment is claimed.

## Final gate remediation — 2026-10-16

Current reviewed source: `06e72dda0addd2744106e22ef35fec539337999b`, codedigest `f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf`. The clean-checkout config baseline is explicit freeze data; this workspace's original config remains unchanged and unstaged. Final mode is fully implemented before freeze and requires `ingest_base_sha` while ordinary run mode permits that field optionally. It checks actual HEAD, freeze/base ancestry, every parent diff (including merge-only create/remove), planning-only ingest history/net/worktree paths, and protected config changes in the freeze commit and after it.

Reached regression controls first reproduced, then rejected: hardcoded config baseline, final-mode stub, committed config and restore, merge-only non-planning changes, merge-only digest changes, and config inclusion in ordinary or merge freeze commits. Executor's post-commit three-root freeze/final union passed in 12.991s. Independent code review is clean; independent security is SECURED 7/7. Each independently ran only the new TestPhase38FreezeCommitConfigControls (security 1.623s); earlier unions remain credited to their actual revisions. Phase33 refresh ran that new root in 2.277s and recomputed canonical v3 freshness, reusing its unchanged historical controls explicitly. No full producer, broad local native matrix, full/race or pressure test was run.

The manifest remains 106 cases, 17 control roots and 480 event names. Actual milestone audit and freeze selection remain the next preparation gates; hosted authority/results and final copied-byte acceptance remain pending.

The new pre-freeze milestone audit at06e72dda joins18/19 requirements with EVD-17 alone pending; no additional implementation blocker. See M007-MILESTONE-AUDIT.md and its bounded integration/three-source inputs. Actual freeze preflight results belong to38-01-SUMMARY.md after the metadata commit.
