---
milestone: M007
milestone_name: Bitwise Checksum and Composable Scalar Helpers
audited: "2026-10-16T21:40:06.234614+00:00"
audited_head: 320f9b556b68f42d0395f8da23135cfcd63232b5
audit_stage: post-terminal-closure-pre-receipt
status: gaps_found
terminal_closure: passed
scores:
  requirements: 18/19
  phases_at_audit: 5/7
  integration: 10/10
  scoped_runtime_flows: 6/6
gaps:
  requirements:
    - id: EVD-17
      status: unsatisfied
      phase: "38"
      verification_status: missing
      orphan_kind: expected_unexecuted_owner
      evidence: Final same-source Ubuntu/macOS receipts remain Phase38-owned and unbound.
  integration: []
  flows: []
terminal_dispositions:
  CR-03: closed
  CR-04: closed_corrected_checked_source_OpFail_target
  CR-06: closed
new_blocking_regressions: []
manual_joins_resolved: [CALL-01, CALL-02, DX-24]
orphaned_from_owner_verification: [EVD-17]
unintended_implemented_requirement_orphans: []
nyquist:
  compliant_phases: ["33", "34", "35", "36", "37", "37.1"]
  partial_phases: []
  not_validated_phases: []
  missing_phases: ["38"]
  overall: partial_expected_future_owner
next_action: "$gsd-discuss-phase 38"
---

# M007 audit after terminal closure

**Terminal closure passes: CR-03/04/06 are closed and no new blocking regression was found.** Milestone status remains `gaps_found` solely because EVD-17 intentionally awaits Phase38 hosted receipts. This audit does not bind a freeze commit or authorize push/dispatch.

The independent source audit inspected ten explicit component connections (six cross-phase, four within-phase) and six scoped runtime flows; all are wired. Final hosted binding is a separate pending flow. Denominators describe this finite inventory, not compiler completeness. The original pre-closure audit remains byte-identical at [PRE-CLOSURE-MILESTONE-AUDIT.md](audits/M007-2026-10-16-terminal-closure/PRE-CLOSURE-MILESTONE-AUDIT.md).

## Terminal closure

| Obligation | Disposition | Source and evidence boundary |
| --- | --- | --- |
| CR-03 | Closed | Correctly attributed child defects close all active invocations in legacy and dynamic modes. Ordinary/failed child completion closes one. Causal frame closure does not prove physical cleanup. |
| CR-04 | Closed against corrected admitted target | Checked Phase24 source produces indexed foreign.failed followed by OpFail, including compatible caller error transfer/root attribution. Contract/typefact/place/activation controls are independent. The historical synthetic OpReturn probe did not establish checked-source admission; arbitrary OpReturn failures remain unclaimed. Literal model values are separate from causal evidence. |
| CR-06 | Closed | Checked no-call scalar defects produce structured native documents with process0 and peer acceptance; application65 and exact stderr bytes retain their contract. Static counted output preserves valid control/NUL reasons; malformed reason UTF-8 is refused before JSON cloning. Dynamic controls stay green; native Unit remains refused. |

The census covers325source rows (Phase34:162×4routes; Phase36:163×3routes). Exactly five existing sum_to_n route keys carry reviewed old-source/core/C-bound byte edits. Frozen corpora/goldens are unchanged. Older Phase273/28/35 comparisons reuse that same strict projection; unrelated-byte/stale/wrong-route controls reject. No broad normalization or regenerated oracle is used. WR-02's obsolete ownership comment was corrected during the necessary raw-core semantic edit; its previous disposition remains historical.

## Current evidence and limits

The four canonical terminal proofs are green/unflagged. Each uses the same exact suite in its bad/clean/default controls; all12receipts, all37selected bindings per proof, and53raw evidence records were independently checked. The final union records19roots/254subtests/zeroskips. These records and local raw pointers are in the Phase37.1 projection/resolution; raw receipts remain machine-local under /tmp.

The bounded predecessor regression accounts for250package/test selectors and251root events across16groups:220unchanged passing receipts reused, five former failures rerun green, and26previously unrun roots passed. TestPhase33CaptureReadBounded is explicitly excluded from this refresh; its earlier evidence remains historical with unchanged source. Six documentation/lifecycle/groundedness guards pass. Original-source expected RED and clean/default controls remain distinct from green acceptance. The independent audit inspected source/evidence and ran no tests itself.

No full-project/whole-package suite, current Linux/hosted result, physical destructor proof, scalar-payload/numeric Outcome.Value authentication, executed CFG-branch theorem, hermetic build, readability study or cold/warm latency distribution is claimed. Prior two-host/full-matrix receipts retain their revisions and limitations.

## Verification and requirements

Phases33–37 now all have fresh canonical passed reports, including Phase34's scoped user-delegated wording acceptance. The integration snapshot predates the final Phase34 UAT owner update; [PARENT-AGGREGATION.json](audits/M007-2026-10-16-terminal-closure/PARENT-AGGREGATION.json) records the later five passed owner queries. The user asked the assistant to recommend and follow a choice; the existing operator-specific diagnostic was retained. This is delegated editorial acceptance, not an independent human-readability measurement.

Phase37.1's first verifier confirmed the three behavior outcomes but correctly left predecessor freshness and this audit pending. Those administrative inputs are now available for its final independent owner recheck; the phase count above is the audit-time snapshot (five passed, Phase37.1 final recheck next, Phase38 unstarted), not a new source gap. Phase37.1 owns no canonical requirement IDs.

All19requirements were joined across canonical checkboxes/traceability, owning verification tables, and42SDK-extracted summaries. Fifteen join directly; CALL-01/02 and DX-24 have missing historical summary-completion fields, manually resolved against current owner verification, validation and independently checked consumers. Historical summaries and canonical requirements remain unchanged. No implemented requirement is unintentionally orphaned.

| Requirement | Owner | Final disposition | Join |
| --- | --- | --- | --- |
| QLT-13 | 33 | satisfied | Three-source join |
| QLT-14 | 33 | satisfied | Three-source join |
| QLT-15 | 33 | satisfied | Three-source join |
| EVD-16 | 33 | satisfied | Three-source join |
| DX-20 | 34 | satisfied | Three-source join |
| DX-21 | 34 | satisfied | Three-source join |
| DX-22 | 34 | satisfied | Three-source join |
| DX-23 | 34 | satisfied | Three-source join |
| BIT-01 | 35 | satisfied | Three-source join |
| BIT-02 | 35 | satisfied | Three-source join |
| BIT-03 | 35 | satisfied | Three-source join |
| BIT-04 | 35 | satisfied | Three-source join |
| CALL-01 | 36 | satisfied | Manual source-backed join; summary history preserved |
| CALL-02 | 36 | satisfied | Manual source-backed join; summary history preserved |
| APP-13 | 37 | satisfied | Three-source join |
| MOD-07 | 37 | satisfied | Three-source join |
| APP-14 | 37 | satisfied | Three-source join |
| DX-24 | 37 | satisfied | Manual source-backed join; summary history preserved |
| EVD-17 | 38 | unsatisfied | Expected future receipt owner |

## Audit records and next gate

[INTEGRATION.md](audits/M007-2026-10-16-terminal-closure/INTEGRATION.md) and its JSON define the inspected links/flows and terminal dispositions. REQUIREMENTS.json preserves the initial three-source matrix; SUMMARY-EXTRACTION.json preserves all42SDK outputs; EVIDENCE.json separates source inspection, newly executed checks, reused receipts and historical limits. PARENT-AGGREGATION.json records the final manual joins and owner statuses.

Finish Phase37.1's independent verifier/phase-close owner, then discuss/plan Phase38. EVD-17 stays unchecked/Pending until the final same-source Ubuntu/macOS receipts. Any later push/hosted dispatch checkpoint must name the actual freeze commit. No new phase, backend, dependency or dormant experiment was promoted by this audit.
