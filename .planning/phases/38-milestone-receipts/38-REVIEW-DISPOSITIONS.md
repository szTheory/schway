# Phase 38 pre-freeze review dispositions

The initial review and security report are preserved as dated history. No freeze has been selected.

| Finding | Disposition | Required action |
| --- | --- | --- |
| CR-01 / T-38-01 | Fix | Real package and test JSON event handling, with bounded real-capture controls at both parsers |
| CR-02 | Fix | All descendants emitted by every selected root are included; codedigest's real child events are covered |
| CR-03 / T-38-01 | Fix | Full M007 inventory count and distinct rows, preserving 39 inherited rows |
| CR-04 | Fix | Resolve bounded in-repo preparation records, compare bytes/hash/source/digest and required dispositions; refuse absent/changed/stale records |
| WR-01 / T-38-04 | Fix | Exact integer identity and boolean input types, rejecting Python bool/float equality aliases |
| T-38-02 | Fix | Actual hosted run ID/attempt in producer identity and exact log/metadata equality controls |
| WR-02 | Reject factual premise; retain source-grounded count correction | Plan 01 Task 02 explicitly requires TestPhase35BitwiseOptimizationMatrix and optimization subtests; context D-07 also requires named optimization evidence. The parent's 16-root/341-event figures came from an interim executor update, before its final committed inventory. They were not a locked maximum. The reviewed snapshot has 17 roots/474 events, including the explicitly planned bitwise matrix. Preserve that control, inventory its complete emitted descendants and report the final measured count after fixes. Do not impose an arbitrary count threshold that drops required evidence. |

This is a correction of the supplied count premise, not an accepted security risk or a new scope decision. No source finding is closed until its focused control and independent re-review establish closure.

## Remediation and independent closure

Reached red controls: 6e827de2; production fixes: 02cfd716; real wrong-package test delta: 0d5d5144. CR-01..04, WR-01 and T-38-01/02/04 are independently closed. The later missing wrong-package-control warning is closed by reached refusals at both parsers. Current code review is clean, and security is SECURED 7/7 for Plan 01 preparation. The full bitwise matrix remains as explicitly planned; final measured inventory is 17 roots/480 events/106 process rows, with the six previously omitted codedigest descendants now present. No accepted risk, new dependency or remote authority follows from these dispositions.

## Final preparation remediation — 2026-10-16

Current independent review clean and security SECURED7/7 at 06e72dda0addd2744106e22ef35fec539337999b, codedigest f21194d7244778ac073269734e0103a6e1e0f6c59335e84305176401b465aadf. PF01–PF04 and the freeze-commit config boundary are closed by reached red/green controls (47500361/88b11413/efff39f9, ef935deb/991abac4, 87e1a6d9/06e72dda). Explicit workspace config data replaces a nonportable hardcoded baseline. Final mode is complete before freeze; Plan03 remains planning-only. All-parent NUL path history protects digest changes, final scope, and config in the freeze commit and descendants. No finding was waived. Current test receipts are credited in VALIDATION; named older review files retain their original scope.
