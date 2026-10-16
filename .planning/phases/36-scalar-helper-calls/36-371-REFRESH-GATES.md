# Phase 36 — D18 stale-report resume gates (2026-10-16)

This is the pre-verifier closeout record for the existing execute-phase resume route. 8/8 plans and summaries are present; init.execute-phase reports zero incomplete plans. The verifier owner reported the prior passed report stale because current covered inputs changed. The original report bytes are preserved at [36-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md](36-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md), SHA-256 `111846425602ab2dcb9b208c54f44a69bf4e4f45e42cfbd303286e5cbf826fa4`. The verifier has not run in this refresh.

**Phase goal:** A scalar function can be factored into small helpers: it calls another scalar function with one U64 or Bool argument and uses the result, including calls into imported modules, with the same defect behavior on both engines and every peer re-deriving the call independently.

**Requirements:** CALL-01, CALL-02  
**Phase path:** `.planning/phases/36-scalar-helper-calls`  
**Current source:** `a7b21e292e289fd82a40d6c39af499e78aa54199` (the source delta begins after Phase 37's previous refresh HEAD `9258a7360222ff17a3267e1ccb873a0dfae79346`).

## Gate outcomes

| Gate | Outcome | Evidence and scope |
| --- | --- | --- |
| Nyquist validation | Existing plan map remains `validated`; no predecessor VALIDATION.md was rewritten. The D18 source delta has its own source-bound 19-root / 254-subtest current acceptance union, reused from [Phase 37.1 validation](../37.1-close-gap-legacy-terminal-outcomes/37.1-VALIDATION.md). The prior-phase registry refresh below stopped before its remaining selectors, so predecessor acceptance is not claimed fresh by this row. | [36-371-REFRESH-RUN.json](36-371-REFRESH-RUN.json), shared [regression receipt](../37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION.json) |
| Security | Existing authored `36-SECURITY.md` remains `verified`, ASVS L1, `threats_open: 0`. The current Phase 37.1 security register independently closes its 13 scoped controls; it supplements this register and does not replace it. No historical register or accepted-risk text was edited. | Existing `36-SECURITY.md`; [37.1-SECURITY.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-SECURITY.md) |
| UI applicability | Fresh `check ui.safety-gate 36 --raw`: `frontend=false`, `hasUiFiles=false`, `hasUiSpec=false`, `block=false`. | N/A; no frontend UI surface |
| Code review and disposition | Reused the current standard-depth review of the actual D18 source delta: 15 files, 0 critical / 0 warning / 0 info, status clean. Its three Phase 37.1 findings are all recorded fixed. Prior phase contracts were cross-checked against the delta and the exact overlap tests below; their historical REVIEW/DISPOSITION files were preserved. Current D18 evidence reruns the Phase36 no-call generated-C byte baseline and controls, causal-attribution control, and emitter helper roots. WR-02 remains historically recorded as deferred although the source comment was corrected in Phase37.1. | [37.1-REVIEW.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW.md), [37.1-REVIEW-DISPOSITION.md](../37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW-DISPOSITION.md) |
| Decimal parent-artifact close | No-op: this phase is non-decimal and init section_manifest includes only `regression-gate`, excluding `gap-closure-artifacts` and `partial-wave`. | Fresh init.execute-phase manifest |
| Prior-phase regression | **Aborted before verifier dispatch.** The one deduplicated registry run attempted 32 roots from the first three of 16 groups: 31 passed, 1 failed, 0 skipped, 0 missing. `TestEmittedCExitLiteralsMatchTable` observed exit literal 74 in addition to 0/65/69 and failed at `cgen_exit_literals_test.go:70`. The run halted there; 219 further roots were not run. `TestPhase33CaptureReadBounded` was excluded under the large-input prohibition. | [Shared regression receipt](../37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION.json) and raw JSON logs |

## Source delta and coverage boundary

The previous report covered 102 inputs with digest `v3:sha256:c90fbf587b476354621ea29e16323bb1d73741c6e287b39b0e28c656ddaf3dc0` and verified `2026-10-16T04:07:18Z` (`32/32 must-haves verified`). The current delta intersects these covered source/test inputs:

- `internal/compiler/cgen/cgen_program.go`
- `internal/compiler/corevalidate/corevalidate.go`
- `internal/compiler/executionpeer/executionpeer.go`
- `internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go`
- `internal/compiler/session/session_phase36_prechange_test.go`
- `internal/compiler/syntax/parser.go`

Other D18 implementation, test, and frozen compatibility inputs are new paths outside the old report's covered set. All 14 changed implementation/testdata paths are listed in the shared regression receipt. The current Phase 37.1 proof set binds the 37-path acceptance input map, four green/unflagged prohibition proofs, 53 raw evidence records, and the 19-root union. That evidence covers the D18 behavior slice, not every earlier phase acceptance criterion.

The original report copy is byte-identical to the pre-refresh canonical report; its historical `passed` status remains preserved there. Current `verification.status` is `stale` and routes back to execute-phase. No verifier report has been written, no UAT or summary was replayed, and no phase was marked complete.

**Next:** Parent-owned resolution of the exit-74 invariant in the active Phase 37.1 workflow, then resume only unresolved regression gates and dispatch independent verifiers after the regression gate passes.
