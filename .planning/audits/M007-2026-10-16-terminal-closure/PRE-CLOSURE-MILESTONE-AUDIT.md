---
milestone: M007
milestone_name: Bitwise Checksum and Composable Scalar Helpers
audited: 2026-10-16T13:09:50Z
audited_head: 502ab55edb501dbe2c6b3c101b45852bb7ec29bb
audit_stage: pre-receipt
status: gaps_found
scores:
  requirements: 18/19
  phases: 5/6
  integration: 22/26
  flows: 6/10
counts:
  integration_wired: 22
  integration_broken: 3
  integration_pending: 1
  flows_complete: 6
  flows_broken: 3
  flows_pending: 1
  summary_files: 38
  summary_matrix_initial_satisfied: 15
  summary_matrix_initial_partial: 3
  summary_matrix_initial_unsatisfied: 1
  manually_resolved_requirements: [CALL-01, CALL-02, DX-24]
gaps:
  requirements:
    - id: EVD-17
      status: unsatisfied
      phase: "38"
      claimed_by_plans: []
      completed_by_plans: []
      verification_status: missing
      orphan_kind: expected_unexecuted_owner
      evidence: "Phase 38 is intentionally unplanned until this audit and gap closure. No final M007 hosted Ubuntu/macOS receipts exist. Retain its existing owner; do not create a duplicate implementation gap phase."
  integration:
    - id: CR-03
      severity: BLOCKER
      from: legacy interpreter child-defect terminal
      to: executionpeer activation closure
      affected_requirements: [CALL-02, EVD-17]
      impact: "Carried semantic debt; indirect CALL-02 boundary and hard EVD-17 freeze prerequisite, not failure of the new dynamic helper acceptance."
      evidence: "Current peer closes only the child in static identity mode; the audit overlay reproduces executionpeer.preorder with an unterminated caller."
      owner: execution-peer maintainer
    - id: CR-04
      severity: BLOCKER
      from: interpreter OpReturn modeled typed failure
      to: executionpeer classifier and terminal attribution
      affected_requirements: [CALL-01, CALL-02, EVD-17]
      impact: "Carried semantic debt; indirect helper/peer boundary and hard freeze prerequisite. New U64/Bool helper acceptance remains green."
      evidence: "Interpreter emits function.failed for propagated OpReturn failure; peer accepts that event only for OpFail. Audit overlay reproduces executionpeer.unknown_kind."
      owner: execution-peer maintainer
    - id: CR-06
      severity: BLOCKER
      from: no-call native scalar OpDefect emission
      to: native structured defect document and peer attribution
      affected_requirements: [CALL-02, EVD-17]
      impact: "Carried semantic debt; indirect no-call compatibility boundary and hard freeze prerequisite. Frozen generated-C bytes currently pass."
      evidence: "Checker-admitted audit source produces application stderr and exit 65; native ExpectDefect reports native.run_failed. Application-shell control passes."
      owner: emitter maintainer
  flows:
    - id: CR-03
      name: static child defect attribution
      broken_at: executionpeer activation unwind
    - id: CR-04
      name: propagated modeled typed failure attribution
      broken_at: executionpeer OpReturn event classification
    - id: CR-06
      name: no-call native explicit defect
      broken_at: native structured terminal emission
orphaned_from_owner_verification: [EVD-17]
unintended_implemented_requirement_orphans: []
process_gaps:
  - phase: "36"
    item: "All eight summary requirements-completed fields remain empty although CALL-01/CALL-02 passed later independent phase gates."
    disposition: "Resolved manually for this audit using current verification and newly executed helper, dynamic peer, native defect and no-call byte controls. Preserve historical summaries."
  - phase: "37"
    item: "DX-24 is absent from summary requirements-completed despite explicit guide coverage and passed later independent phase gates."
    disposition: "Resolved manually using current verification and newly executed tracked-only guide. Preserve historical summaries."
nyquist:
  capability_active: true
  compliant_phases: ["33", "34", "35", "36", "37"]
  partial_phases: []
  not_validated_phases: []
  missing_phases: ["38"]
  overall: partial
  qualification: "All five implemented phases compliant. Phase 38 missing by design; validate it after execution, not before it exists."
tech_debt:
  - phase: 36-scalar-helper-calls
    items:
      - "WR-02: obsolete same-type raw-call ownership comment in corevalidate; implementation derives sourceTypeID correctly. Keep the existing next-semantic-edit owner; no runtime failure is implied."
evidence:
  integration_report: audits/M007-2026-10-16/INTEGRATION.md
  integration_data: audits/M007-2026-10-16/INTEGRATION.json
  requirements_cross_reference: audits/M007-2026-10-16/REQUIREMENTS.json
  summary_extraction: audits/M007-2026-10-16/SUMMARY-EXTRACTION.json
  command_and_hash_manifest: audits/M007-2026-10-16/EVIDENCE.json
  positive_commands: 12
  positive_root_executions: 24
  failing_diagnostic_commands: 2
next_action: '$gsd-phase --insert 37 "Close gap: legacy terminal outcomes"'
---

# M007 milestone audit

**Verdict: gaps_found.** Eighteen of nineteen requirements are satisfied at
their dated acceptance scope. All five implemented phases pass current
verification. Three carried critical defects still break terminal-evidence
flows and block the freeze. The remaining requirement, EVD-17, intentionally
belongs to the unplanned Phase 38 after this audit and gap closure.

This is the required pre-receipt audit, not milestone completion or permission
to freeze. No finding is waived, accepted as risk, or fixed by this report.

## Scope and evidence

The definition of done comes from PROJECT.md's M007 goal, ROADMAP.md Phases
33–38 and all nineteen REQUIREMENTS.md traceability rows. The implementation
arc is bounded capture and generic gates → agent diagnostics and discovery →
bitwise/hex operations → scalar helper composition → two ordinary-source CRC
consumers and public file routes → final same-source hosted receipts.

All five complete VERIFICATION.md files and all 38 SUMMARY frontmatters were
read/extracted. The audit also consulted current review dispositions,
validation/security records, the living product/maturity documents, the three
relevant dormant seeds and the local Dashbit source appraisal. Earlier
milestone archives remain historical; no archived decision was rewritten.

The specialized integration checker independently traced current source and
executed twelve bounded positive commands: **24/24 root executions passed**.
Two temporary diagnostic commands reproduced the three blockers. Their
nonzero results are failure evidence, never green validation. The parent
independently checked all 55 original source/test/overlay/log hashes, parsed
the named root results, and retained the raw evidence in Git-visible planning
artifacts. The [command and hash manifest](audits/M007-2026-10-16/EVIDENCE.json)
records exact selectors, original and durable log paths, and source hashes.

The peer probes submit handcrafted core/event pairs matching the documented
interpreter terminal forms; they do not establish source/core admission or
execute the interpreter. Existing review and current interpreter/peer source
inspection establish the mismatch. Closure must include paired real
interpreter/peer regressions. The CR-06 probe checks admitted source and an
actual native process; it is an always-failing diagnostic, so its observed
`native.run_failed` output, rather than its mere exit status, is the evidence.
An earlier checker-refused probe is retained and explicitly excluded.

## Phase verification and validation

| Phase | Plans / summaries | Current verification | Nyquist | Security / review |
| --- | --- | --- | --- | --- |
| 33 | 5/5 | passed, 60/60 truths | validated, compliant | verified L1, zero open threats; CR-03/04/06 deferred critical |
| 34 | 8/8 | passed, 38/38 truths | validated, compliant | verified L1, zero open threats; three review warnings fixed |
| 35 | 8/8 | passed, 38/38 truths | validated, compliant | verified L1, zero open threats; two review warnings fixed |
| 36 | 8/8 | passed, 32/32 truths | validated, compliant | verified L1, zero open threats; WR-02 comment debt deferred |
| 37 | 9/9 | passed, 14/14 distinct truths | validated, compliant | verified L1, zero open threats; review warning fixed |
| 38 | 0 planned | missing, expected future owner | missing, expected | not yet applicable |

The canonical progress query returned fresh `passed` for 33–37, with no
indeterminate checks. Phase 38 has no directory. Its missing verification is
an incomplete milestone condition, not an accidental loss of evidence.
Nyquist discovery is active via the `verify:post` validate-phase hook. No
retroactive validation is needed for 33–37; run `$gsd-validate-phase 38` only
after that phase is implemented if its normal execution gate has not already
done so. Historical pending prose in earlier task receipts is superseded by
the dated independent phase gates, without editing the original receipts.

## Requirements: three-source cross-reference

Every row was compared across the canonical traceability table and checkbox,
the owning verification's requirement table, and all summary completion
fields. [REQUIREMENTS.json](audits/M007-2026-10-16/REQUIREMENTS.json) retains
each claiming plan, completing summary, initial matrix verdict and final
manual disposition.

| Requirement | Behavior | Owner | Verification | Summary completion entries | Canonical | Final |
| --- | --- | --- | --- | --- | --- | --- |
| QLT-13 | bounded capture read before full-file allocation | 33 | satisfied | 33-01 | checked / Complete | satisfied |
| QLT-14 | carried freeze, owner-predicate and review debt | 33 | satisfied | 33-01/02/03/05 | checked / Complete | satisfied |
| QLT-15 | Phase 29 triage and Phase 28 evidence records | 33 | satisfied | 33-05 | checked / Complete | satisfied |
| EVD-16 | generic milestone freeze/receipt gates | 33 | satisfied | 33-03/04/05 | checked / Complete | satisfied |
| DX-20 | captured-source locations and snippets | 34 | satisfied | 34-01/02/03/08 | checked / Complete | satisfied |
| DX-21 | one diagnostic per recovery cause | 34 | satisfied | 34-01/04/08 | checked / Complete | satisfied |
| DX-22 | dedicated absent-form codes and hints | 34 | satisfied, shift sub-item closed by 35 | 34-01/05/08 | checked / Complete | satisfied |
| DX-23 | deterministic current surface manifest | 34 | satisfied | 34-01/06/07/08 | checked / Complete | satisfied |
| BIT-01 | independently checked pure bitwise operations | 35 | satisfied | 35-02/03/04/08 | checked / Complete | satisfied |
| BIT-02 | literal 0..63 shifts, defined values and refusals | 35 | satisfied | 35-02/03/04/08 | checked / Complete | satisfied |
| BIT-03 | hexadecimal admission, spelling and refusals | 35 | satisfied | 35-02/03/08 | checked / Complete | satisfied |
| BIT-04 | Adler shift with unchanged answers | 35 | satisfied | 35-08 | checked / Complete | satisfied |
| CALL-01 | local/imported one-argument helpers and defects | 36 | satisfied under dated D-18 | absent | checked / Complete | satisfied, manual |
| CALL-02 | independent call laws/controls and no-call C bytes | 36 | satisfied under dated clarification | absent | checked / Complete | satisfied, manual |
| APP-13 | ordinary bounded CRC-32/ISO-HDLC consumer | 37 | satisfied | 37-01 | checked / Complete | satisfied |
| MOD-07 | second importer of the same CRC module | 37 | satisfied | 37-01 | checked / Complete | satisfied |
| APP-14 | public interpreter/native file tuple parity | 37 | satisfied | 37-01 | checked / Complete | satisfied |
| DX-24 | executable clean-checkout CRC guide | 37 | satisfied | absent | checked / Complete | satisfied, manual |
| EVD-17 | final same-source Ubuntu/macOS receipts | 38 | missing | absent; no plans yet | unchecked / Pending | unsatisfied, expected pending |

The literal matrix initially yields 15 satisfied, three partial and one
unsatisfied. The three partials are manually resolved here, following the
matrix's verify-manually path and the M006 audit precedent. Their summaries
explicitly reserved completion for later independent gates; those gates have
since passed. Newly executed `TestPhase36HelperAnswersAndDefects`,
`TestPhase36CallCausalAttributionControls`,
`TestPhase33ScalarExplicitDefectShells`,
`TestPhase36NoCallGeneratedCByteBaseline` and
`TestPhase37GuideCleanCheckout` corroborate the relevant paths. The current
Phase 36 verification additionally owns independent raw core/origin and
mutation evidence. This is not credit from summaries alone, nor an assertion
that the legacy broken paths satisfy a wider helper contract.

No implemented requirement is accidentally orphaned. EVD-17 is the sole ID
without an owning phase verification: an expected future-owner orphan at this
audit point, kept unsatisfied. Its absence forces `gaps_found` even without the
three additional integration blockers. No requirement checkbox is changed.

## Cross-phase integration and end-to-end flows

The finite inventory contains **26 connections: 22 wired, three broken, one
pending**. It contains **ten flows: six complete, three broken, one pending**.
These are enumerated checks, not language-completeness or quality percentages.
The [full integration report](audits/M007-2026-10-16/INTEGRATION.md) maps all
nineteen requirements to their consumers and source anchors.

The six complete selected flows are bounded native capture, agent diagnostic
and repair, bitwise/Adler computation, local/imported scalar helpers, public
CRC file applications, and the tracked-only CRC guide. Their checks include
independent expected answers, reached algorithm/refusal mutations, both public
engines and error/boundary cases. Generic gates also accept the preserved M006
receipts locally. This makes no new hosted claim. Pure numeric derivation,
causal attribution and physical resource observation remain separate proofs.

| Blocker | Broken flow and current source | Reproduced observation | Requirement relationship / owner |
| --- | --- | --- | --- |
| CR-03 | legacy child defect → activation unwind, executionpeer.go:693–700 | `executionpeer.preorder`: caller activation has no terminal event | carried debt; indirect CALL-02, hard EVD-17 freeze prerequisite; execution-peer maintainer |
| CR-04 | modeled failure on OpReturn → event classification, interp.go:2829–2834 and executionpeer.go:738–740 | `executionpeer.unknown_kind` for `function.failed` | carried debt; indirect CALL-01/02, hard freeze prerequisite; execution-peer maintainer |
| CR-06 | no-call scalar OpDefect → native defect document, cgen_program.go:3008–3012 | admitted source gives app stderr/exit 65 and `native.run_failed`; app-shell control passes | carried debt; indirect CALL-02 compatibility boundary, hard freeze prerequisite; emitter maintainer |

Durable reproductions: [peer diagnostic](audits/M007-2026-10-16/evidence/peer-debt.log)
and [native no-call diagnostic](audits/M007-2026-10-16/evidence/cgen-nocall-debt.log).
The direct new helper acceptance remains green. CR-05's dynamic native defect
repair is closed and was rerun; it does not close CR-06. Passing frozen no-call
C bytes proves preservation, not correct native no-call terminal behavior.

The final receipt flow is pending because Phase 38 is deliberately last.
HTTP routes, auth protection and graphical UI are inapplicable to this local
compiler/CLI milestone; no such feature is inferred from the generic template.

## Scoped closure and routing

Insert **one terminal-outcome closure phase after 37**, provisionally 37.1,
grouping the three related defects. Its named runnable gain is that a legacy
child defect, a propagated typed failure and an admitted no-call explicit
defect each produce the correct terminal outcome and independently admissible
evidence. This is a specific safety obligation, not another feature-enabling
phase or a general evidence-framework expansion.

1. **Execution peer, CR-03/04:** close every active frame on a whole-execution
   defect independent of call-identity mode. Preserve ordinary one-frame
   returns. Derive the propagated OpReturn failure's actual source/type and
   outcome; preserve handled child failures and the existing OpFail path.
   Add real interpreter/peer positive traces plus wrong-source/type/outcome,
   forged terminal, post-terminal, handled-failure and dynamic-identity
   controls. Do not broaden classification based only on event spelling.
2. **Emitter, CR-06:** emit the structured no-call native defect terminal with
   exact reason and retain app-shell exit 65/stderr. Convert the diagnostic
   into a permanent meaningful regression at O0/O3/O3-LTO and validate the
   resulting terminal independently. Keep the dynamic CR-05 control green.
3. **Compatibility and acceptance:** before changing frozen no-call C,
   record a dated, source-grounded amendment resolving the semantic repair
   against the preservation obligation. Change only affected baseline rows
   when authorized by that explicit decision; preserve original history and
   unaffected bytes. Refresh affected verification evidence after real source
   changes, preserve completed UAT, and rerun this milestone audit. The new
   closure summary should record its requirement disposition after its gates,
   without rewriting earlier summaries.

Next command:

```text
$gsd-phase --insert 37 "Close gap: legacy terminal outcomes"
```

Then use discuss → plan → execute for the returned decimal phase. This audit
proposes the closure scope; it does not insert or execute it. EVD-17 retains
Phase 38 ownership and needs no duplicate gap phase. After closure and an
updated audit, plan Phase 38. Its final push/dispatch checkpoint must name the
actual freeze commit; this invocation authorizes no hosted action.

## Nonblocking debt and evidence limits

Phase 36 **WR-02** remains one comment-maintenance item: corevalidate's old
same-type explanation disagrees with its correct `sourceTypeID` derivation.
Retain its next-semantic-edit owner; an isolated comment edit is unnecessary
for this audit. Earlier fixed review items stay closed. Phase 37's previously
accepted T-37-07 limitation remains in its original security record; no new
risk acceptance is made here.

Current evidence is local Darwin arm64 with exact bounded selectors. No full
local project/package suite, current hosted/Linux result, human readability
study, cold/warm latency distribution, hermetic or deterministic native binary,
universal FileBytes binding limit, arbitrary divergent-interpreter timeout
parity or Darwin leak-detection result is claimed. Parser recovery and hex
formatter details also rely on the current phase reports; their exact roots
were not independently rerun by this audit. Earlier hosted and mutation
receipts retain their revisions and do not stand in for EVD-17.

## Three useful next capabilities

The living roadmap's order is unchanged. No status-only rewrite of its covered
files is needed; this audit owns the newly confirmed failure evidence and
closure route.

| User-visible program | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence / debt | Owner / next action | Observation changing priority |
| --- | --- | --- | --- | --- | --- | --- |
| Audited crc32/crc_low/checksum and earlier pinned programs on Ubuntu/macOS | Three terminal blockers, then final same-source receipts | Close the named legacy outcomes, audit again, freeze once and bind every required two-host case | Correct propagation/terminal laws; preserve independent numeric, causal and process proofs; explicit C compatibility amendment if needed | CR-03/04/06; EVD-17; this audit and historical M006 receipts | Peer/emitter closure owner, then Phase 38 receipt owner and freeze checkpoint | Any reached defect, missing required case, host disagreement or source drift delays freeze |
| Proposed M008 wc FILE / longest_line FILE sharing one module | Exact output contract plus justified arity-N/scalar-record/exported-type needs | Specify bounded inputs and literal answers, then the smallest complete counting/longest-line pair | Add only consumer-required argument, record/export and result laws across independent consumers | Future requirement landing; existing module/ByteView/FileBytes/CRC witnesses | Next-milestone language/checker owner after M007 | Existing scalar results suffice, or another consumer needs fewer new laws |
| Bounded fixed-schema key=value configuration checker before JSON | Schema, duplicates and malformed-input contract undecided | One schema, byte/line bounds, literal good/bad fixtures and bounded errors | Bound parser work and error attribution; add representation laws only when required | Product-roadmap order; current fixed file reader; no JSON admission claimed | Future consumer owner specifies contract before promotion | A concrete JSON consumer provides a smaller justified slice, or current operations suffice |

## Dormant seed dispositions

| Seed | Current source witness and named consumer | Disposition / trigger | Next owner action |
| --- | --- | --- | --- |
| z1g | QueryCommandFile in session_phase6_query.go; reviewer finding CRC helpers/direct calls | retain dormant; no measured answer-accuracy or navigation-cost gap | Query owner names a failed/costly task before one source-unit experiment |
| trz | readApplicationCapture and current file/capture controls; CRC/checksum failure diagnosis | retain dormant; existing source/events already explain all three blockers; original unbounded-read concern was repaired in 33 | Native/evidence owner demonstrates one unexplained named failure before optional read-only diagnosis |
| z1s | syntax.Format, check/scalar.go and CRC module/guide; actual source reviewer | retain dormant; no contested contract/inference/source-form choice or measured human review gap | Language owner and actual reviewer name a paired task and independent obligations |

The local Dashbit appraisal remains historical research. Its database/runtime/
proof-backend suggestions have no new consumer justification. No dormant seed
or future capability is promoted by this audit.
