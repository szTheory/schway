---
phase: "36"
slug: scalar-helper-calls
status: draft
threats_open: 2
asvs_level: 1
block_on: high
register_authored_at_plan_time: true
created: "2026-10-15"
---

# Phase 36 — Security

Independent ASVS Level 1 audit returned OPEN_THREATS at 2026-10-15T05:42:10Z: 22 of 24 unique plan threats closed; two high-severity planned adversarial controls remain open. The bounded arena and positive census exist; the audit demonstrated no production exploit. Phase advancement remains blocked until the controls are implemented and independently re-audited.

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
| T-36-13 | Tampering | invocation pointer lifetime and ancestry | high | mitigate | Task 1 uses append-only bounded storage; Tasks 1/3 check post-return paths and stack-pointer mutation. | open |
| T-36-14 | Denial of service | C event/path/ordinal storage | high | mitigate | Task 2 checks exact-plus-one guards before write/allocation/increment, including canonical ordinal exhaustion. | closed |
| T-36-15 | Tampering | capture/real process outcome separation | high | mitigate | Task 2 independently pins app streams/outcome and late 65 defect after optional recording stops. | closed |
| T-36-16 | Tampering | inherited emitter inventory | high | mitigate | Task 3 owns exact census adaptations and live missing/extra-site/preflight controls; old C corpus remains frozen. | open |
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

## Open Findings

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
