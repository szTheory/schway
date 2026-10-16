# Phase 36 caller-shape repair and closeout — 2026-10-15

Phase 36 passes independent verification: 32/32 must-haves,24 artifacts,19 links,
26/26 decisions, zero unverified behavior and no human item. Canonical report
binds 94 files with digest
v3:sha256:5b64fbab5ee94a1a53fc96d493b27932d7934b0c5498b2eb93cfc380b84fb7de.
CALL-01/02 and all eight plans/21 tasks/seven waves remain complete.

## Repair and evidence

Independent review found CR-01: existing structured unused-Byte-to-U64 and
ByteView-to-Unit callers lowered helper calls but could fail native emission
unless an unrelated eligible caller enabled dynamic support. Commit d69380ed
repairs the bounded classifier and adds two regression tests. Byte formals
remain independently unused; helper signatures and peer/ownership laws are
unchanged. Commit 0e56a3be registers the new exact EmitNative consumer.
Original-source overlays fail both named emission assertions; repaired source
and independent closure pass. The Byte case proves literal 11 on interpreter
and native O0/O3/O3-LTO plus separate causal attribution. The ByteView Unit
case proves TU emission/linking and unrelated U64 entry result 7 newline, not
execution of that Unit body or a new ByteView entry/value-call route.

New local checks: 84 mapped+2 repair roots (86 names/roots) pass in 13 packages.
The final combined union passes 239 distinct names/251 root executions across16
packages, zero missing/fail/skip in 136.534seconds, with 179 input/16 log hashes.
The owner receipt binds 85 inputs/14 logs. Both mechanical prohibitions freshly
prove actual-source bad RED and clean/default GREEN with 22 bound inputs/two
raw results. The initial union's sole inventory failure is preserved and
resolved; only the final all-green union earns acceptance. Plan08 historical
capture producers were parsed read-only for owners and artifact checks;
original receipt revisions remain historical. See36-REFRESH2-RUN.json,
36-REFRESH2-REGRESSION.json,36-REFRESH2-PROHIBITIONS.json and36-REFRESH2-FIX.md.

The first verifier supported 32/32 but blocked on two future-phase plan-count
placeholders. Report7d007160 preserves that assessment. ROADMAP now explicitly
states Phase 38 Not planned /0 planned /Not started, preserving audit-before-
planning. Independent re-verification closed the metadata gap and audited
coverage omissions, adding direct runner/formatter/manifest/proof authorities
before copying the canonical 94-file fingerprint. See36-REFRESH2-METADATA-FIX.md
and36-VERIFICATION.md. No Phase 38 directory or plan was created.

## Remaining debt and limits

Current review has only WR-02, deferred stale core-validator explanatory
commentary; the runtime contract is correct, and review is not clean. CR-01
and historical WR-01 are fixed, canonical ledger zero open/one deferred of
three total. Phase 33 CR-03/04 remain separate deferred BLOCKER propagation false
rejections, owned by the execution-peer maintainer through mandatory M007 audit
before Phase 38 freeze. This report does not close or accept them. All 24 authored
L1/high threats retain closed dispositions; no wider certification is claimed.
UI is N/A. No new dependency/backend/schema/admission law or hosted receipt.
No full local project/package suite, Linux run, subjective human readability
or whole-loop cold/warm latency distribution is claimed. Frozen fixtures,
original plans/summaries, historical captures and completed UAT are preserved.

## Completion and next gate

phase.complete 36 ran exactly once: 8/8 plans, no warnings/conflicts/preservation
or indeterminate-freshness result. It retained canonical next-unimplemented38;
ROADMAP/requirements already matched and were not rewritten by completion.
Fresh status:34/36 passed,33/35/37 stale. Exact next command is
`$gsd-execute-phase 33`, then re-query after each refresh, audit/gap closure
including explicit CR-03/04 disposition, then Phase 38 planning/freeze/hosted
binding. The source repair and review records legitimately stale earlier
reports; passing regression never substitutes for their own fingerprints.

User config bytes and persistent auto-advance/worktree/model preferences are
preserved; ephemeral chain flag false. Global learnings is disabled; no recent
LEARNINGS item, matching phase todo, stale handoff or other active workstream
needs action. Post-completion transition begins after canonical completion
without calling it again. Only expected S004 for unplanned Phase 38 remains.
This invocation stops here; no automatic next phase, push or dispatch.



## Three useful next capabilities

| User-visible program | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence / debt | Owner / next action | Observation changing priority |
| --- | --- | --- | --- | --- | --- | --- |
| Audited crc32/crc_low/checksum and earlier pinned programs on Ubuntu/macOS | Freshness33/35/37, mandatory audit including propagation blockers, same-source host receipts. | Refresh current gates, disposition audit/gap work, freeze one revision, bind both hosts. | Preserve language laws; independently establish required propagation, actual host outcomes and source/fixture identity. | EVD-17; Phase 33 CR-03/04 ledger; current251-root gate; old host receipts historical. | Execution-peer maintainer and milestone auditor, then Phase 38receipt owner. | A required propagation failure, missing lane, host disagreement or source drift changes freeze priority. |
| Proposed M008 wc FILE /longest_line FILE sharing one module | Arity-N, scalar records/exported types and exact output contract need a complete named consumer. | Specify bounded literal cases and deliver the smallest counting/longest-line pair through normal ownership. | Add only needed argument/record/export/result laws across independent semantic consumers; retain resource/loan loop refusals. | Future requirements landing; current CRC/module and bounded FileBytes witnesses. | Next-milestone language/checker owner after M007. | A current-surface solution or another real consumer needs fewer new laws. |
| Bounded fixed-schema key=value config checker before JSON | Schema, duplicates and malformed-input/output contract undecided. | One bounded schema with independent good/bad literal cases, promoted normally. | Bound parser work/errors and add representation laws only as required. | Living roadmap order; current bounded reader and path-free outcomes. | Future consumer owner names the contract before implementation. | A concrete JSON consumer is smaller, or existing ByteView/scalar operations suffice. |

## Relevant seed dispositions

Living roadmaps, all three relevant seed trigger bodies and Dashbit appraisal
were read at entry and completion transition. Source anchors remain current;
archived decisions are preserved except these dated observations.

| Seed | Disposition | Current source witness / named consumer | Trigger assessment | Owner / next action |
| --- | --- | --- | --- | --- |
| z1g | retain dormant | session_phase6_query.go:140 QueryCommandFile; reviewer finding crc_step/crc32 definitions and callers. | No measured accuracy or material loop-cost gap against query plus ordinary search. | Query maintainer names a failed task before one bounded source-unit experiment. |
| trz | retain dormant | native_app.go:694 readApplicationCapture; checksum/CRC failure and terminal diagnosis. | Current bounded capture/replay and source explain findings; no unexplained failure. Old pre-read allocation concern was resolved in Phase 33 and remains historical. | Native/evidence owner demonstrates a named evidence gap before an optional local read-only experiment. |
| z1s | retain dormant | syntax/format.go:19 Format and check/scalar.go:937 lowerScalarCall; human audit of CRC helpers. | No named disputed source/contract form or human measurement. Formatter/AI inspection does not prove readability. | Language/checker owner and actual reviewer name a contested choice and independent obligations first. |

No future consumer or dormant experiment is promoted; only EVD-17 is active.

## Final closeout checks

After the final PROJECT/STATE/closeout changes, five exact session metadata/
report-grounding roots and the native living-roadmap root passed, zero missing,
fail or skip. Raw logs and SHA-256:

- /tmp/schway36-r2-closeout-session.log: 790d0c14ed62787234f3ae74eb29ffccff582fc9ed71456bc403becb2bfc3678
- /tmp/schway36-r2-closeout-native.log: 0b2d35b29bcd8dc63fe6b986762a09c7411e501334815fee7795ccfce39a6f5b

All 85 owner, 179 regression and 22 prohibition input bindings still match;
all 30 Go raw logs and both prohibition result hashes match. Canonical status
remains passed. User config matches the invocation snapshot, milestone lock
is absent and git diff --check is clean. Strict state validation reports only
the expected S004 for intentionally unplanned Phase 38.
