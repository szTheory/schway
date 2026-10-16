# Phase 37.1 predecessor-debt disposition amendment — 2026-10-16

This dated note records implementation facts observed during the stale-report gate run. It is additive: historical M007 audit, Phase 33/36 review reports and accepted-risk statements remain unchanged until the parent-owned independent verifiers and fresh milestone audit update them.

| Historical item | Current implementation evidence | Current disposition |
| --- | --- | --- |
| CR-03 — legacy child defect left active callers without whole-execution closure | executionpeer closes all active frames on the reached legacy child-defect trace. The fresh source-bound D18 proof set has four green/unflagged subjects and a 19-root union (254 subtests, no skips); named peer roots include TestPhase371LegacyDefectClosesAllActiveFrames and terminal-attribution controls. | Implementation fixed; independent phase verification and milestone-audit disposition remain pending. |
| CR-04 — checked modeled foreign failure lacked propagated typed-failure attribution | A checked foreign.failed at indexed OpForeignCall flows through the generated error block's OpFail; peer roots cover source/type/outcome and child-to-caller error-edge attribution. | Implementation fixed; independent phase verification and milestone-audit disposition remain pending. |
| CR-06 — scalar no-call OpDefect lacked structured native terminal | No-call native emission carries the structured defect document, preserves application exit 65, and checks counted reason bytes. The fresh 19-root union includes the named no-call conformance root; bounded comparator controls remain separate. | Implementation fixed; independent phase verification and milestone-audit disposition remain pending. |
| WR-02 — obsolete same-type raw-call ownership comment | The semantic core-validator edit corrected the explanatory comment to describe ownership from sourceTypeID and independent parameter/result checks. | Comment corrected within the semantic edit. Historical Phase 36 disposition/review remains unchanged; parent review/audit will decide its final bookkeeping. |

No historical acceptance risk was rewritten and no hosted or two-host claim is made. See 371-REFRESH-COMPLETION-2026-10-16.json for exact current bindings, logs, hashes, counts and coverage limits.
