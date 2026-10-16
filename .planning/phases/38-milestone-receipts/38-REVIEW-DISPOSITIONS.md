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
