# Phase 36 — D18 stale-report resume gates (2026-10-16)

**Goal:** A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Requirements:** CALL-01, CALL-02. **Phase path:** .planning/phases/36-scalar-helper-calls. Fresh init.execute-phase reports 8/8 plans complete and zero incomplete plans; its manifest includes regression-gate and excludes gap-closure-artifacts and partial-wave.

The prior report was reported stale after source inputs changed. Its original bytes remain in .planning/phases/36-scalar-helper-calls/36-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md, SHA-256 111846425602ab2dcb9b208c54f44a69bf4e4f45e42cfbd303286e5cbf826fa4; historical status was passed. No verifier ran during this refresh. Source delta begins at Phase 37 refresh head 9258a7360222ff17a3267e1ccb873a0dfae79346; current checkout head is 20a519468839a0698e2c4e93fc51300509c6abbf. Prior report intersecting covered paths: internal/compiler/cgen/cgen_program.go, internal/compiler/corevalidate/corevalidate.go, internal/compiler/executionpeer/executionpeer.go, internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go, internal/compiler/session/session_phase36_prechange_test.go, internal/compiler/syntax/parser.go.

## Current pre-verifier gate outcomes

- Validation: Current D18 evidence is source-bound to 37 paths: four green, unflagged proof subjects and a 19-root union (254 subtests; no skips). The finite predecessor registry passed 250 package/test selectors, 251 root execution events, 16 groups, 240 distinct names, no missing/fail/skip. TestPhase33CaptureReadBounded is excluded. This scoped sampler is not the full historical acceptance inventory. Existing phase VALIDATION files were preserved.
- Security: Existing authored phase register remains verified with zero open threats. The additional Phase 37.1 register closes 13/13 scoped controls; it supplements each older register. No accepted-risk text changed.
- UI: Fresh check ui.safety-gate 36 --raw reports frontend=false, hasUiFiles=false, hasUiSpec=false, block=false; not applicable.
- Review/disposition: The prior Phase 37.1 standard review covered 15 files and was clean at its recorded review point, with its three adapter findings fixed. The later proof-adapter/source-evolution update and exact inventory-coordinate edit are not covered by that review. Parent independent review remains pending. The dated CR-03/04/06 and WR-02 amendment is at ../37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DISPOSITION-AMENDMENT-2026-10-16.md; historical review/audit records remain unchanged.
- Parent artifacts: No-op. The fresh execute-phase manifest excludes gap-closure and partial-wave inputs.
- Regression: Complete for the fixed bounded registry. The 220 unchanged passing root receipts were reused with their prior log/source basis; the five formerly failing roots were rerun and now pass; all 26 previously unrun roots in groups 15–16 pass. No broad 98-root session rerun occurred. The inventory RED is preserved: actual AST locations 782/784 versus stale registry rows 775/777. Only those two registry coordinates were updated; classifications and refusal witness remain strict. Exact GREEN rerun passed inventory and refusal-witness roots 2/2 with no skips. The central completion receipt records exact commands, root inventory, current bindings, and per-log hashes.

Six post-artifact session documentation/lifecycle/groundedness guards passed with count=1 and no skips; their exact root names and log SHA-256 are recorded in the central completion receipt. The original verification report and history copy are unchanged. Current verification remains stale until the parent-owned independent verifier runs. No phase was completed and no hosted or two-host claim is made.

## Parent review gate completion — 2026-10-16

Independent standard review of the eight-file regression delta is clean with zero findings; prior unchanged production scope retains its clean15-file review. Current proofs/37inputmaps/all12canonicalreceipts were independently checked. The preserved three earlier findings remain fixed in Phase37.1 disposition; there are no new findings to triage. The earlier pending-review line above is superseded by this dated gate result.

Final security delta recheck at f68674bb retained Phase37.1 controls13/13closed/zeroopen, supplementing the phase existing verified authored register. Independent delta review clean. These settle the parent review/security checkpoints; independent goal verifier is now the next step.
