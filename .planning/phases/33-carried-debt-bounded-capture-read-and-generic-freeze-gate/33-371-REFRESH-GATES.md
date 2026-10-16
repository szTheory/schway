# Phase 33 — D18 stale-report resume gates (2026-10-16)

This is the pre-verifier closeout record for the existing execute-phase resume route. 5/5 plans and summaries are present; init.execute-phase reports zero incomplete plans. The verifier owner reported the prior passed report stale because current covered inputs changed. The original report bytes are preserved at [33-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md](33-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md), SHA-256 `43a268e45bd603737e4f5cc1471c97f48b45195ba8132d3f79b9cd311d7bf23d`. The verifier has not run in this refresh.

**Phase goal:** The carried M006 debt is paid before any feature work: a child-written capture file can no longer force an unbounded host allocation, the open review warnings are closed or recorded with reasons, Phase 28 has its missing validation and security records, and later milestones reuse one freeze/receipt gate instead of copying phase-specific scripts.

**Requirements:** QLT-13, QLT-14, QLT-15, EVD-16  
**Phase path:** `.planning/phases/33-carried-debt-bounded-capture-read-and-generic-freeze-gate`  
**Current source:** `a7b21e292e289fd82a40d6c39af499e78aa54199` (the source delta begins after Phase 37's previous refresh HEAD `9258a7360222ff17a3267e1ccb873a0dfae79346`).

## Gate outcomes

| Gate | Outcome | Evidence and scope |
| --- | --- | --- |
| Nyquist validation | Existing plan map remains `validated`; no predecessor VALIDATION.md was rewritten. The D18 source delta has its own source-bound 19-root / 254-subtest current acceptance union, reused from [Phase 37.1 validation](../37.1-close-gap-legacy-terminal-outcomes/37.1-VALIDATION.md). The prior-phase registry refresh below stopped before its remaining selectors, so predecessor acceptance is not claimed fresh by this row. | [33-371-REFRESH-RUN.json](33-371-REFRESH-RUN.json), shared [regression receipt](../37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION.json) |
| Security | Existing authored `33-SECURITY.md` remains `verified`, ASVS L1, `threats_open: 0`. The current Phase 37.1 security register independently closes its 13 scoped controls; it supplements this register and does not replace it. No historical register or accepted-risk text was edited. | Existing `33-SECURITY.md`; [37.1-SECURITY.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-SECURITY.md) |
| UI applicability | Fresh `check ui.safety-gate 33 --raw`: `frontend=false`, `hasUiFiles=false`, `hasUiSpec=false`, `block=false`. | N/A; no frontend UI surface |
| Code review and disposition | Reused the current standard-depth review of the actual D18 source delta: 15 files, 0 critical / 0 warning / 0 info, status clean. Its three Phase 37.1 findings are all recorded fixed. Prior phase contracts were cross-checked against the delta and the exact overlap tests below; their historical REVIEW/DISPOSITION files were preserved. Current D18 evidence directly exercises the Phase33 terminal-attribution and scalar-defect shell roots in the 19-root proof union. Historical CR-03/04/06 disposition rows remain unchanged; the new behavior evidence is additive pending the independent verifier. | [37.1-REVIEW.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW.md), [37.1-REVIEW-DISPOSITION.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW-DISPOSITION.md) |
| Decimal parent-artifact close | No-op: this phase is non-decimal and init section_manifest includes only `regression-gate`, excluding `gap-closure-artifacts` and `partial-wave`. | Fresh init.execute-phase manifest |
| Prior-phase regression | **Halted before verifier dispatch on resumed registry group 14.** The original exit-74 mismatch is preserved in first-run history and resolved by test-only commit 02680cc8; the host exitcode table is unchanged. Reused groups 1–2 passed 12/12; rerun cgen passed 20/20; groups 4–13 passed 95/95. Session group 14 ran 98 roots: 93 passed, 5 failed, 0 skipped. The five failures are 'TestPhase16PublicEmitterConsumerInventory', 'TestPhase273ConformanceCorpusByteIdentity', 'TestPhase28PrunedEmitMatchesGolden', 'TestPhase35PreChangeEmitterCorpus', and 'TestPhase35PreChangeCorpusControls'. Groups 15–16 remain unrun (26 selectors). The bounded-capture pressure selector remains excluded. | [Shared regression receipt](../37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION.json) and raw logs 01–15 |

## Source delta and coverage boundary

The previous report covered 232 inputs with digest `v3:sha256:d548fa31d1ef2aa73da6159f1c57260a5ef319e6ea74d8ea3f38245ff0bd910f` and verified `2026-10-16T03:13:25Z` (`60/60 must-haves verified`). The current delta intersects these covered source/test inputs:

- `internal/compiler/cgen/cgen_program.go`
- `internal/compiler/corevalidate/corevalidate.go`
- `internal/compiler/executionpeer/executionpeer.go`
- `internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go`
- `internal/compiler/session/session_phase34_prechange_test.go`
- `internal/compiler/session/session_phase36_prechange_test.go`
- `internal/compiler/syntax/parser.go`

Other D18 implementation, test, and frozen compatibility inputs are new paths outside the old report's covered set. All 14 changed implementation/testdata paths are listed in the shared regression receipt. The current Phase 37.1 proof set binds the 37-path acceptance input map, four green/unflagged prohibition proofs, 53 raw evidence records, and the 19-root union. That evidence covers the D18 behavior slice, not every earlier phase acceptance criterion.

The original report copy is byte-identical to the pre-refresh canonical report; its historical `passed` status remains preserved there. Current `verification.status` is `stale` and routes back to execute-phase. No verifier report has been written, no UAT or summary was replayed, and no phase was marked complete.

**Next:** Parent-owned resolution of the exit-74 invariant in the active Phase 37.1 workflow, then resume only unresolved regression gates and dispatch independent verifiers after the regression gate passes.
