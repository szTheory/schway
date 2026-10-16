# Phase 36 refreshed closeout — 2026-10-15

Phase 36 is accepted after independent verification: 32/32 must-haves (four roadmap criteria, 26 plan truths and two mechanical prohibitions), 24 artifacts, and 26/26 context decisions. No implementation gap or human item remains. The report binds 62 inputs with canonical digest v3:sha256:4f46995f362e3f195d977d7498baf71be8ab33c2cde97ae60e898cacd9a30db0. CALL-01/02 remain complete. Eight plans, 21 tasks, seven waves and completed UAT are preserved.

## Changes and evidence

Only planning/evidence records changed. Source inspection reconfirmed one U64/Bool argument, four parameter/result pairs, computed arguments evaluated once, imported calls, independently checked contracts, per-parent/site dynamic invocation identity and actual application outcomes preserved after optional recording stops. New helper callee defects are overflow/remainder; byte-bounds uses the documented legacy ByteView callee regression plus the separate caller-before-entry fault. Causal acceptance does not prove numeric answers or source argument/result intent.

The first independent report verified all 32 truths but returned gaps_found for four stale completed-phase plan counters in ROADMAP.md. Those entries now match actual plans, summaries and the existing progress table: 34=8/8, 35=8/8, 36=8/8, 37=9/9. Phase38's unplanned placeholders are preserved. The first report and repair are retained in d60ae944; independent re-verification closed the gap and produced the final passed report in a8788260. See 36-REFRESH-METADATA-FIX.md.

Newly executed focused checks: the mapped owner union passed 84 exact roots / 13 packages, zero missing/fail/skip, at source 3d647e44. It covers all 21 task rows. The 19 direct Go commands and two plan08 wrapper owner lists were resolved into exact tests. Original Python capture wrappers were not replayed because they append to source-bound historical receipts; their fixed owners ran freshly and their artifact assertions were reused read-only with original historical revision preserved. 36-REFRESH-RUN.json binds 82 inputs and 13 logs.

The required prior-phase regression union passed 167 distinct names / 179 package-qualified root executions in 16 packages at 534fac4d, zero missing/fail/skip. 36-REFRESH-REGRESSION.json binds 172 current inputs and 16 logs. Duplicate internal/external names retain multiplicity. Commands are one-shot, exact, uncached, JSON-captured and bounded through canonical normalization/timeout. The mapped helper O0/O3/O3-LTO/UBSan and real-source mutation controls ran. Both existing prohibition descriptors independently returned canonical green after actual-source bad RED, clean GREEN and default GREEN, recorded in 36-REFRESH-PROHIBITIONS.json; the verifier reran both and checked input/log hashes, source behavior and rendered-report groundedness itself.

Historical receipts: original Phase36 acceptance/prohibition/regression captures and all earlier hosted receipts retain their stated revisions and limits. No whole-package/project local suite, new hosted run, Linux run, hosted binding, human readability result or cold/warm latency distribution is claimed. Production compiler/runtime, examples, frozen C/source/identity fixtures, wire contracts and dependencies are byte-identical throughout this refresh. No dependency or backend was added.

## Review and gate disposition

The independent standard review inspected 53 explicit source/fixture paths. It found one non-runtime documentation warning, WR-02: an older core-validator explanatory block describes obsolete same-type calls while the implementation and nearby current comments correctly derive argument ownership from sourceTypeID and check parameter/result contracts separately. It remains explicitly deferred to the core-validator maintainer's next semantic edit. The review is not clean and the warning is not fixed. Historical WR-01 remains fixed; the canonical disposition has zero open of two total findings. Validation has one found/resolved metadata gap, none escalated. All 24 authored L1/high threats remain closed; no L2/L3 claim. UI applicability is N/A with frontend/hasUiFiles/hasUiSpec/block false. Automatic UI hook reused the existing review.

## Completion and next gate

phase.complete36 ran exactly once with 8/8 plans and zero warnings, no conflict/preservation/indeterminate result. It retained next unimplemented phase38 in canonical state.json. Fresh init.progress reports 34/36 passed, 33/35/37 stale. The new review/disposition bytes are covered inputs of the old33/35 reports: only those Phase36 gate records changed, not their implementation. A passing regression does not replace their stale fingerprints. Do not edit predecessor digests or erase the current finding.

Exact next command: `$gsd-execute-phase 33` for gates/verifier only, then re-query progress (currently35, then37). Preserve completed plans/UAT; complete M007 audit and gap closure before Phase38 planning, push/dispatch checkpoint or hosted binding. S004 only says the intentionally unplanned Phase38 directory is absent; no placeholder directory or manual canonical state projection rewrite is warranted. Current completion/routing facts are in 36-REFRESH-COMPLETION.json.

User config bytes, persistent auto-advance/worktree preferences and false ephemeral chain flag are preserved. Global learnings is disabled; there are no LEARNINGS items, matching pending phase todos, stale phase handoffs or other active workstreams to process. This invocation stops after Phase36 closeout.

## Three useful next capabilities

| User-visible program | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence / debt | Owner / next action | Observation changing priority |
| --- | --- | --- | --- | --- | --- | --- |
| Audited crc32/crc_low/checksum and earlier pinned programs on Ubuntu/macOS | Report freshness for33/35/37, milestone audit and same-source host receipts. | Refresh gates, close audit gaps, freeze one revision and bind both hosts once. | Preserve current language laws; independently establish host outcomes and source/fixture consistency. | EVD-17; current34/36 reports and fresh179-root regression; original M006 receipts historical. | Existing verifier owners, milestone auditor, then Phase38 with its explicit push/dispatch checkpoint. | Reached audit defect, skipped required case, host disagreement or source drift changes freeze priority. |
| Proposed M008 wc FILE / longest_line FILE sharing one module | Arity-N, scalar records/exported types and exact output contract need a named complete consumer. | Specify bounded literal inputs/outputs and deliver the smallest counting/longest-line pair through normal milestone ownership. | Add only required arity/record/export/result laws across independent semantic consumers; preserve resource/loan loop refusals. | Future requirements landing; existing CRC/module and bounded FileBytes witnesses. | Next-milestone language/checker owner after M007 acceptance. | A current-surface solution or another real consumer needs fewer new laws. |
| Bounded fixed-schema key=value config checker before JSON | Schema, duplicate-key policy and malformed-input/output contract are unsettled. | Name one schema and byte/line limits with independent good/bad cases, then promote through normal ownership. | Bound parser work/errors and add representation laws only as required by the consumer. | Living roadmap order; current bounded reader and path-free outcomes. | Future consumer owner specifies contract before implementation. | A concrete JSON consumer is a smaller justified slice, or existing ByteView/scalar operations suffice. |

No future consumer or dormant experiment is promoted into active requirements; EVD-17 remains the only active M007 requirement.

## Relevant seed dispositions

The living roadmaps, seed trigger metadata/bodies and Dashbit appraisal were consulted for this transition. Archived decisions stay historical except for these dated observations.

| Seed | Disposition | Current source witness / named consumer | Trigger assessment | Owner / next action |
| --- | --- | --- | --- | --- |
| z1g bounded symbol/caller query | retain dormant | session_phase6_query.go:QueryCommandFile; reviewer navigating crc32/crc_step helper definitions and calls. | No measured exact-answer or whole-loop cost gap against current query plus search. | Query maintainer names a failed task and measures current workflow before experiment/promotion. |
| trz consumer failure observability | retain dormant | native_app.go:readApplicationCapture and capture-outcome separation; crc32/checksum failure diagnosis. | Current capture, explicit errors and replay explain tested cases; no unexplained named failure. Appraisal's old pre-read allocation concern was resolved in Phase33 and remains historical. | Native/evidence owner names a failure beyond current evidence before an optional local read-only experiment. |
| z1s compare contract/source forms | retain dormant | syntax/format.go:Format, check/scalar.go:lowerScalarCall; human audit of scalar CRC helpers. | No concrete disputed contract/inference/canonical-form choice or human measurement. Formatter and AI inspection do not establish human readability. | Language/checker owner and real reviewer name a contested choice and independent obligations before comparison. |

## Final closeout checks

After all current-state and living-document edits, five exact session roots passed: report-grounding frontier and empty classes, Phase32 living claims, Phase34 transition documents and Phase35 operator documentation. The native Phase25 living-roadmap root also passed. No root was missing, failed or skipped. Raw outputs are `/tmp/schway36-closeout-final-session.log` and `/tmp/schway36-closeout-final-native.log`. Both refresh receipts retained matching 82/172 input and 13/16 log hashes. Canonical Phase36 status stayed passed, user config bytes matched the invocation snapshot, milestone lock was absent and git diff --check passed.
