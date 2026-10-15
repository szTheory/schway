---
phase: "36"
slug: scalar-helper-calls
status: verified
threats_open: 0
asvs_level: 1
block_on: high
register_authored_at_plan_time: true
created: "2026-10-15"
---

# Phase 36 — Security

Final independent ASVS Level 1 re-audit returned SECURED at 2026-10-15T06:08:12Z: all24 unique plan threats closed and threats_open0. The earlier audit findings and their correction history are retained below. The two planned control gaps are resolved without production, dependency or frozen-fixture changes.

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation | Status |
|---|---|---|---|---|---|---|
| T-36-01 | Tampering | no-call C corpus | high | mitigate | Task 1 commits literal enumerated pre-edit bytes; Task 2 detects path/route drift, omission and false exemption. | closed |
| T-36-SC | Tampering | dependency installation | high | accept | No package installation. | closed |
| T-36-02 | Denial of service | recursive scalar call parser | high | mitigate | Task 1 shares the 64-depth counter and tests exact boundary plus recovery siblings. | closed |
| T-36-03 | Tampering | source call argument/result lowering | high | mitigate | Tasks 2/3 assert exact SourceID/TargetID/type/order and reach argument/result selection controls. | closed |
| T-36-04 | Elevation of privilege | module and scalar signature admission | high | mitigate | Task 3 reaches private/unimported/collision/cycle and forbidden-signature refusals through existing closed resolver. | closed |
| T-36-05 | Tampering | raw OpCall contract/freshness | high | mitigate | Task 1 directly validates malformed raw calls and reaches refusal-removal controls. | closed |
| T-36-06 | Tampering | origin scalar transfer | high | mitigate | Task 2 independently validates exact contracts, static freshness and unknown result state. | closed |
| T-36-07 | Elevation of privilege | scalar-call loan/back-edge context | high | mitigate | Task 3 retains source-loan accounting and refuses owner/loan/back-edge authority forgeries. | closed |
| T-36-08 | Spoofing | repeated/nested invocation identity | high | mitigate | Tasks 1/2 independently derive parent/site ordinals and reject gap/duplicate/reorder/parent/site/callee forgeries. | closed |
| T-36-09 | Tampering | terminal causal attribution | high | mitigate | Task 2 checks exact call/result fields, paired terminals, forged defects and post-defect rejection without numeric replay. | closed |
| T-36-10 | Denial of service | interpreter event/path/counter retention | high | mitigate | Task 3 checks bytes before allocation/append, portable ordinal ceiling, disabled recording and exact-plus-one controls. | closed |
| T-36-11 | Tampering | incomplete evidence versus actual interpreter result | high | mitigate | Task 3 returns actual result plus named evidence error and proves later real defect/output survives after capture stops. | closed |
| T-36-12 | Tampering | C evaluation and call result | high | mitigate | Tasks 1/3 require sequenced statements and literal dropped-argument/result controls. | closed |
| T-36-13 | Tampering | invocation pointer lifetime and ancestry | high | mitigate | Task 1 uses append-only bounded storage; Tasks 1/3 check post-return paths and stack-pointer mutation. | closed |
| T-36-14 | Denial of service | C event/path/ordinal storage | high | mitigate | Task 2 checks exact-plus-one guards before write/allocation/increment, including canonical ordinal exhaustion. | closed |
| T-36-15 | Tampering | capture/real process outcome separation | high | mitigate | Task 2 independently pins app streams/outcome and late 65 defect after optional recording stops. | closed |
| T-36-16 | Tampering | inherited emitter inventory | high | mitigate | Task 3 owns exact census adaptations and live missing/extra-site/preflight controls; old C corpus remains frozen. | closed |
| T-36-17 | Tampering | helper numeric/defect outcome | high | mitigate | Tasks 1/3 pin independent literals and kill dropped argument/result/overflow/remainder per engine at O0/O3/O3-LTO. | closed |
| T-36-18 | Tampering | byte argument evaluation and old callee bounds | high | mitigate | Task 2 proves caller OOB before entry separately from legacy ByteView callee bounds with exact owner/outcome. | closed |
| T-36-19 | Elevation of privilege | public input/capture route claims | high | mitigate | Task 2 preserves unused-Byte laws, U64 input refusal and actual app/capture separation. | closed |
| T-36-20 | Tampering | capability/authority discovery claims | high | mitigate | Task 1 checks literal complete-program pairs/contexts/tails and exact unchanged unsupported boundaries. | closed |
| T-36-21 | Tampering | manifest source/limit census | high | mitigate | Task 2 enumerates changed branches/consumers, preserves frozen fixture hashes and reaches missing/false claim/hint/restriction controls. | closed |
| T-36-22 | Tampering | final compatibility ledger/corpus | high | mitigate | Task 1 rejects changed old C, omission, false exemption and invalid exact amendment after integration. | closed |
| T-36-23 | Repudiation | evidence/guarantee publication | high | mitigate | Task 2 binds each claim to actual named owner/lanes and keeps numeric, causal, capture, historical and hosted facts distinct. | closed |

## Accepted Risks Log

| Risk ID | Threat Ref | Rationale | Accepted By | Date |
|---|---|---|---|---|
| R-36-SC | T-36-SC | All eight Phase36 plans explicitly selected accept for dependency installation: no packages are installed and no dependency or mandatory runtime is added; implementation uses the Go 1.24 standard library and installed Clang/C17 tools. This records the existing plan decision and does not accept a new implementation risk. | Phase36 plan-time decision authority; 36-CONTEXT D-01 and all eight PLAN threat registers | 2026-10-14 |

## Initial Open Findings — historical

| Threat | Missing declared control | Next action |
|---|---|---|
| T-36-13 | Actual native returned/reused stack-path pointer mutation. Parent-path substitution tests ancestry and cannot establish storage lifetime. | Extend TestPhase36EmitterCallInventoryControls with fresh clean/no-op/restored native positives and a named retained-path lifetime failure. |
| T-36-16 | Live removal and extra-site mutations of the emitted function/call/event/reservation census. | Rerun the actual region census against removed and inserted sites; require the named census mismatch and restored positives. |

## Independent Audit Evidence

The auditor inspected all eight plan registers and actual implementation, independently verified final45-root and compatibility15-root capture hashes/anchored results, and confirmed source equality with d117ca4058fbf6a727cb5aa634768cf6194bac34. Nyquist additionally ran32 inherited invocations. Eight repeated T-36-SC rows are one accepted-risk threat.

- T-36-01/22: session_phase36_prechange_test.go literal/inventory/ledger and drift/omission/false-exemption controls.
- T-36-02/03/04: parser_phase36_test.go bounded depth/recovery and check_phase36_scalar_calls_test.go ordered lowering, exact signatures, imports/cycles/source-intent controls.
- T-36-05/06/07: corevalidate_phase36_test.go, originvalidate_phase36_test.go, pathoracle_phase36_test.go and pathoracle_phase36_borrow_test.go independent freshness/unknown transfer/authority/loan controls.
- T-36-08/09/10/11: interp_phase36_test.go and executionpeer_phase36_test.go dynamic identity, terminal forgery, bounded retention and actual result separation.
- T-36-12/14/15/17/18/19: cgen_phase36_test.go, native_phase36_test.go, session_phase36_calls_test.go, session_phase36_byteview_test.go, session_phase36_peer_mutation_test.go and cmd/schway/main_phase36_test.go actual values/defects/order/capture/input/native lanes.
- T-36-20/21/23: manifest_test.go, session_phase36_surface_test.go, fixed Plan08 owner capture and36-ACCEPTANCE-EVIDENCE.md exact claims/census/evidence categories.

T-36-13: cgen.go append-only arena exists; cgen_phase36_test.go parent-pointer mutation is not the promised stack-lifetime control. T-36-16: positive census exists, but missing/extra-site adversarial controls are absent. These are missing promised checks, not accepted risks.

The source checker proves source intent, raw peers prove contract/authority, engines prove numbers/defects, and executionpeer proves causal attribution. CALL classification remains unclassified/unresolved and descriptorless prohibitions unverified as metadata contracts; no wired policy enforcement, full-suite, hosted, cross-host or human UAT claim.

## Security Audit 2026-10-15

| Metric | Count |
|---|---|
| Threats found | 24 |
| Closed | 22 |
| Open | 2 |

## Re-audit — 2026-10-15T05:59:22Z

T-36-13 is closed: cgen_phase36_test.go now runs the actual function-local stack-pointer escape/reuse mutant, requires returned value10 plus the named retained post-return path failure, and checks clean/no-op/restored native positives.

T-36-16 remains open. The first follow-up's six mutations operated on actual generated lines, but phase36CheckEmitterCallSiteCensus counted call comments. Required correction: enumerate actual callee invocations using checked call contracts, and reject both marker-preserving call removal and an unmarked extra actual call. Event/reservation controls stay live. No new risk is accepted.

The auditor independently verified the refreshed45-root capture at eddb535ae9fa10f3700143e02f7fc5c11106e8d5 and source equality; other22 closures including T-36-SC remain valid. Current disposition is23closed/1blocking open, superseding the earlier22/2 audit summary while preserving its history. No phase advancement.

## Security Audit 2026-10-15

| Metric | Count |
|---|---|
| Threats found | 24 |
| Closed | 23 |
| Open | 1 |

## Final independent re-audit — 2026-10-15T06:08:12Z

T-36-16 is closed: cgen_phase36_test.go independently derives expected callee symbols from checked call contracts, verifies definitions/call lines, strips comments/strings within the user-function body, and enumerates actual SCHWAY_ invocations separately from markers/events/reservations. Eight copied-emission mutations include marker-preserving actual-call removal and an unmarked extra actual call; each requires its named finding and restored positive. T-36-13 retains the actual returned-stack-pointer lifetime mutation with native value10 and the discriminating retained-path failure plus clean/no-op/restored positives.

The auditor independently verified current final45-root/12-package manifest /tmp/schway36-final-83c1f58a1720-jikc_g2_/owner-results.json, output hashes, anchored root records, all eight census subtests and source/test equality with83c1f58a1720770e77216ee82737efa3acdbeeaa. Other22 closures and the plan-time SC accepted risk remain valid. No unregistered flags.

Current open findings: none. All24 register rows are closed; threats_open0 at the high blocking threshold. The audit is source/receipt verification at ASVS1, not a full suite, hosted two-host receipt, human UAT, general security certification, or wired descriptor-policy claim.

## Sign-Off

- [x] All24 unique threats have verified or previously accepted dispositions.
- [x] T-36-SC accepted risk records the existing plan decision.
- [x] Both independently found adversarial-control gaps are resolved.
- [x] threats_open0; statusverified.

Approval: independent GSD security auditor verified2026-10-15.

## Security Audit 2026-10-15

| Metric | Count |
|---|---|
| Threats found | 24 |
| Closed | 24 |
| Open | 0 |

## Review-fix and regression provenance — 2026-10-15

Independent security confirmation at06:32:13Z retained24/24 closed, threats_open0 after WR-01: dynamic end-of-stream refusal metadata now uses the checked dynamic occurrence map; refusal conditions/classes and authority remain unchanged. The auditor independently verified45 current roots at3697ec2950487bb98723c06262d6d7e6bb2ae1d6. Subsequent edits are only tests/summary metadata: exact helper route inventory and baseline/mutation consumer composition. Their independent debug review and the current80-root regression sample are in36-REGRESSION.md. Production, dependencies and frozen expected data have no change since that independent confirmation. This paragraph records provenance; it does not invent another security audit or hosted claim.


## Successor verification refresh — 2026-10-15

The authored plan register still contains 24 unique threats, all closed, at ASVS L1/high. No new dependency, package installation or trust boundary is introduced by this refresh. The zero-open authored-L1 shortcut applies; no L2/L3 audit is claimed. Current source inspection confirmed typed single-evaluation lowering in `check/scalar.go:lowerScalarCall`, independent raw-call admission, per-parent/site ordinal attribution in `executionpeer.Validate`, pre-retention interpreter bounds, bounded `readApplicationCapture`, and preserved actual outcomes on optional capture failure. Fresh mapped roots (84/84) exercise these owners and the numeric/capture/causal mutation controls; both canonical prohibition proof gates are green. Sources and output hashes are bound in 36-REFRESH-RUN.json and 36-REFRESH-PROHIBITIONS.json. No threat was newly accepted or transferred.
