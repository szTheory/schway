---
phase: 36-scalar-helper-calls
plan: "08"
subsystem: testing
tags: [compatibility, evidence, owner-capture, phase-transition]

requires:
  - phase: 36-07
    provides: Exact independent surface manifest and observed owner roots
provides:
  - Source-bound final compatibility and per-knower acceptance receipt
  - Dated evidence and next-capability amendments in both living documents
affects: [phase-36-audit, phase-37-planning, phase-38-receipts]

actuals:
  tokens: 6747
  tasks: 2
  commits: 2
  plan_head_before: 5474f16ecd3e80cf7cdafec73173b2af48c6556f
  plan_head_after: 92300103d4e64e14718e637c3c5e20281a364164

tech-stack:
  added: []
  patterns:
    - Owner-command logs bind argv, source HEAD, expected roots, exits, and output hashes
    - Current source inspection, fresh checks, and historical receipts remain separate

key-files:
  created:
    - .planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md
    - .planning/phases/36-scalar-helper-calls/36-prohibition-check.test.cjs
    - .planning/phases/36-scalar-helper-calls/prohibition-clean.json
    - .planning/phases/36-scalar-helper-calls/prohibition-outcome-violation.json
    - .planning/phases/36-scalar-helper-calls/prohibition-causal-violation.json
    - .planning/phases/36-scalar-helper-calls/36-PROHIBITION-EVIDENCE.md
  modified:
    - .planning/phases/36-scalar-helper-calls/36-08-PLAN.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md

key-decisions:
  - "Bind the final owner capture to compiler source HEAD d117ca4058fbf6a727cb5aa634768cf6194bac34; later commits changed planning documents only."
  - "Leave CALL-01 and CALL-02 pending for the phase-owned independent gates."

patterns-established:
  - "Acceptance rows cite actual test owners and separate numeric, causal, route, and capture claims."
  - "Living roadmap amendments distinguish current source, new focused checks, and historical receipts."

requirements-completed: []
coverage:
  - id: D1
    description: "Task 1 preserves the frozen no-new-call C and refusal inventory with exact new-fixture ledger and drift/omission/exemption controls."
    requirement: CALL-02
    verification:
      - kind: integration
        ref: "internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaseline"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase36_prechange_test.go#TestPhase36NoCallGeneratedCByteBaselineControls"
        status: pass
    human_judgment: false
  - id: D2
    description: "Task 2 records actual per-knower acceptance and the dated living-document capability/seed transition; the Python artifact checks own document grounding, while named Go owners prove only their respective runtime, raw-contract, causal, or compatibility claims."
    requirement: CALL-01
    verification:
      - kind: other
        ref: "36-08-PLAN.md Task 2 exact Python final owner/artifact command; historical accepted capture at 3697ec2950487bb98723c06262d6d7e6bb2ae1d6, /tmp/schway36-final-3697ec295048-safn6e49/owner-results.json, and 36-ACCEPTANCE-EVIDENCE.md"
        status: pass
    human_judgment: false
duration: "not recorded; first-to-last accepted owner capture was 7m 25s"
completed: 2026-10-15
status: complete
---

# Phase 36 Plan 08: Final compatibility and acceptance record Summary

**The integrated helper slice now has source-bound compatibility evidence and a per-knower acceptance record, with unresolved phase assumptions kept visible.**

## Performance

- **Duration:** The executor start marker was not preserved; accepted owner captures span 7m 25s.
- **Started:** Not recorded.
- **Completed:** 2026-10-15T05:30:20Z.
- **Tasks:** 2.
- **Files modified:** 3 task deliverables.
- **Actuals tokens:** 6,747, measured as 26,987 added characters divided by four across the receipt and two living-document amendments.

## Accomplishments

- The frozen compatibility controls passed 15/15 roots at source HEAD `5474f16e`; the tracer feedback rerun passed 15/15 roots at `d117ca40`.
- The amended final fixed-owner capture passed 45/45 expected roots across 12 packages at source HEAD `d117ca4058fbf6a727cb5aa634768cf6194bac34`. Every command exited zero, with no `FAIL` or `SKIP` counted as a pass. The accepted manifest is `/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json`; the receipt links its logs and hashes.
- The acceptance record maps each stable claim to observed test owners and keeps old callee ByteView bounds separate from caller argument bounds. It also keeps source lowering, raw-peer contracts, literal engine outcomes, causal attribution, and optional capture outcomes within their actual evidence boundaries.
- Both living documents received dated capability and dormant-seed dispositions. They distinguish current source inspection, fresh focused checks, and historical receipts, and make no hosted-evidence claim.

## Task Commits

1. **Task 1: Recheck integrated helper additions against immutable no-new-call C and old safety controls** - `d117ca40` (`docs(36-08): capture Phase 36 compatibility evidence`).
2. **Task 2: Record per-knower acceptance and append the living capability/seed transition** - `92300103` (`docs(36-08): record Phase 36 acceptance and transition`).

## Files Created/Modified

- `.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md` - source-bound frozen compatibility results, current owner captures, claim mapping, unresolved assumptions, and actual log links.
- `.planning/PRODUCT-ROADMAP.md` - dated capability, evidence, ownership, and seed transition.
- `.planning/LANGUAGE-MATURITY.md` - dated source and evidence maturity update.

## Decisions Made

- Keep final capture evidence bound to source HEAD `d117ca40`; commit `92300103` and later plan metadata commits affect planning documents only.
- Leave CALL-01 and CALL-02 unchecked and keep Phase 36 closure with the orchestrator's verification and audit gates.
- Preserve the two descriptorless prohibition assumptions as flagged and unverified; no implementation-level enforcement claim was added.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Documentation defect] Corrected acceptance-record references rejected by the fixed document checker**
- **Found during:** Task 2 (per-knower acceptance and living transition).
- **Issue:** An intermediate final capture found a duplicate dated heading and incorrect test-file paths in the receipt and living-document references.
- **Fix:** Removed the duplicate heading and corrected the paths to the actual owner test files.
- **Files modified:** `.planning/phases/36-scalar-helper-calls/36-ACCEPTANCE-EVIDENCE.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`.
- **Verification:** The amended fixed-owner command passed 45/45 roots and the final receipt/document checks; latest manifest: `/tmp/schway36-final-d117ca4058fb-evx1nc2r/owner-results.json`.
- **Committed in:** `92300103`.

**Total deviations:** 1 auto-fixed documentation defect.
**Impact on plan:** The correction made the named evidence references and dated transition satisfy the planned receipt contract; no production or test source changed.

## Issues Encountered

The first document-check attempts exposed receipt-reference errors and were corrected before the accepted final capture. No execution blocker remains. The run was local and focused; no whole-package suite, hosted dispatch, or push was performed.

## User Setup Required

None.

## Next Phase Readiness

The acceptance receipt and living-document amendments are ready for independent Phase 36 audit. Requirement status and phase closure remain pending the orchestrator-owned gates; this summary completes Plan 36-08 only. No Phase 37 work was started.

---
*Phase: 36-scalar-helper-calls*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary and all three task deliverables exist.
- Task commits `d117ca4058fbf6a727cb5aa634768cf6194bac34` and `92300103d4e64e14718e637c3c5e20281a364164` are ancestors of HEAD.
- `git diff --check` passed.

## Dated security-audit evidence refresh — 2026-10-15

The independent security audit added test-only commit `04ce6fa1`; it changes `internal/compiler/cgen/cgen_phase36_test.go` and no production code. `TestPhase36EmitterCallInventoryControls` now covers a returned stack-path reuse mutant that still returns 10 but fails the retained-child-path assertion, clean/no-op/restored native positives, and six region-scoped missing/extra controls for call sites, `function.called` events, and invocation reservations. The dated audit record in [36-05-SUMMARY.md](./36-05-SUMMARY.md) preserves its focused verification details.

I reran the exact fixed final command from the amended `36-08-PLAN.md` at source/test HEAD `eddb535ae9fa10f3700143e02f7fc5c11106e8d5`. It passed 45/45 roots across 12 packages, with every command exiting zero and no `FAIL` or `SKIP`. The current manifest and per-package logs are recorded in the acceptance receipt; the manifest is `/tmp/schway36-final-eddb535ae9fa-g7t84c5o/owner-results.json`. Earlier 15-root compatibility captures remain accurately bound to `5474f16e` and `d117ca40`; they were not rerun because the fixed final set freshly includes the compatibility baseline, controls, and affected cgen owner. The independent security review remains open pending re-audit, and Phase 36 plus CALL-01/CALL-02 remain pending their owner gates.

## Refresh self-check: PASSED

- The current manifest records HEAD `eddb535ae9fa10f3700143e02f7fc5c11106e8d5`, a clean source tree, 12 commands, 45 expected roots, and the new `TestPhase36EmitterCallInventoryControls` root.
- Every command exited zero; all 12 recorded log files exist and match their SHA-256 values in the manifest.
- The original plan task commits and the audit test commit remain ancestors of the tested HEAD.

## Dated emitted-call census correction — 2026-10-15

Test-only commit `50b90a30` corrected the T-36-16 census control in `internal/compiler/cgen/cgen_phase36_test.go#TestPhase36EmitterCallInventoryControls`. The control now counts actual emitted `SCHWAY_` invocations after removing comments and strings, and derives expected callees from checked `core.OpCall` contracts plus emitted definitions. It reaches both a marker-preserving removal of a real invocation and an unmarked extra invocation, alongside the existing six marker, event, and reservation controls. No production code, fixture, dependency, or schema changed.

I reran only the amended fixed final owner command at full source/test HEAD `83c1f58a1720770e77216ee82737efa3acdbeeaa`. It passed 45/45 unique expected roots across 12 owner groups; all exited zero, and the output contained no `FAIL` or `SKIP`. The actual manifest and log hashes are recorded in [36-ACCEPTANCE-EVIDENCE.md](./36-ACCEPTANCE-EVIDENCE.md); manifest: `/tmp/schway36-final-83c1f58a1720-jikc_g2_/owner-results.json`. The earlier `eddb535a` final capture remains an accurate record of its run, but predates the corrected T-36-16 census and does not evidence those corrected controls. Earlier 15-root compatibility captures remain historical at their recorded revisions and were not rerun. The independent security review remains open; this addendum does not clear security or complete Phase 36/CALL acceptance.

## Census-refresh self-check: PASSED

- The current manifest binds full HEAD `83c1f58a1720770e77216ee82737efa3acdbeeaa`, records a clean source tree, and contains 12 commands covering 45 unique expected roots, including `TestPhase36EmitterCallInventoryControls`.
- All 12 commands exited zero; each log exists and matches its recorded SHA-256.
- Test-only commit `50b90a30` and the amended-plan commit `83c1f58a` are ancestors of the source-bound capture revision.
- `git diff --check` passed before this documentation-only append; the final scoped documentation diff is checked before commit.

## Dated WR-01 review-fix capture — 2026-10-15

Code-review fix commit `61217af4` updates `internal/compiler/executionpeer/executionpeer.go` so end-of-stream refusals preserve the expected callee from independently observed dynamic call occurrences. Its two ordinal-1 truncation cases cover pending-child and open-activation boundaries and assert the refusal code, invocation context, and checked callee. The [review-fix report](./36-REVIEW-FIX.md) records the separate focused four-root PASS, including ordinal-zero compatibility. This fix changes neither the accepted source-call contract nor schema, dependencies, fixtures, or corpus.

I ran only the amended fixed final owner command at current source/test HEAD `3697ec2950487bb98723c06262d6d7e6bb2ae1d6`. It passed 45/45 unique roots across 12 owner groups, all commands exited zero, and there was no `FAIL` or `SKIP`. The manifest and log hashes are recorded in [36-ACCEPTANCE-EVIDENCE.md](./36-ACCEPTANCE-EVIDENCE.md); manifest: `/tmp/schway36-final-3697ec295048-safn6e49/owner-results.json`. Earlier captures remain bound to their recorded revisions. The security record reports 24 closed and zero open at source/test HEAD `83c1f58a`; this capture is not a re-audit of `61217af4`, and parent confirmation of the narrow T09 metadata repair remains pending.

## WR-01 capture self-check: PASSED

- The current manifest binds full HEAD `3697ec2950487bb98723c06262d6d7e6bb2ae1d6`, a clean source tree, 12 owner groups, and 45 unique roots including `TestPhase36CallCausalAttributionControls`.
- All 12 commands exited zero; all logs exist and match their recorded SHA-256 values.
- Review-fix commit `61217af4` is an ancestor of the capture revision; the separate four-root review-fix result is recorded in `36-REVIEW-FIX.md`.
- `git diff --check` passed before this documentation-only append; the final scoped documentation diff is checked before commit.

## Dated final current-source acceptance capture — 2026-10-15

I ran the exact fixed final owner command from the amended `36-08-PLAN.md` at full current HEAD `979357388171339005fd6dbd65aa760810ae9384`. It passed 45/45 unique expected roots across 12 owner groups; every command exited zero and the output contained no `FAIL` or `SKIP`. The complete manifest and per-package logs are recorded in [36-ACCEPTANCE-EVIDENCE.md](./36-ACCEPTANCE-EVIDENCE.md); manifest: `/tmp/schway36-final-979357388171-x4n9kf3c/owner-results.json`.

The bounded prior-phase regression receipt remains separately bound to source/test revision `b51043c6566f793893c8c4e143493e7ec3af5bcd`: it passed 80 root occurrences across 13 commands and is recorded in [36-REGRESSION.md](./36-REGRESSION.md), committed as `fbc19851`. Tracked source files are identical from `b51043c6` through this final capture, so the 80/13 sample and earlier 15-root compatibility captures were not rerun. The final 52-file code review is clean with zero open dispositions at `97935738`; this does not change the revision bindings of the earlier evidence. The preexisting frontmatter coverage block, added in `350340c0`, is retained unchanged.

## Current-source capture self-check: PASSED

- The final manifest binds full HEAD `979357388171339005fd6dbd65aa760810ae9384`, records a clean source tree, 12 command groups, and 45 unique expected roots.
- All commands exited zero; all 12 logs exist and match the manifest SHA-256 values.
- The separate `b51043c6` prior-phase regression manifest remains 80/13; source paths are identical through `97935738`, and no regression or compatibility rerun was performed.
- The final review disposition has zero open findings at `97935738`; this current owner capture remains a separate evidence lane.
- `git diff --check` passed before this documentation-only append; the exact scoped diff is checked before commit.

## Dated Task 2 prohibition proof supplement — 2026-10-15

The initial verifier had left two descriptorless scope prohibitions unverified. This bounded follow-up added one Node standard-library adapter and three closed JSON subjects, then wired both original prohibition statements through the existing `projectProhibitions` flat-scalar contract and canonical `check prohibition-enforcement` producer. Both requests returned `status=green`, `flagged=false`, `kind=node-test`, and `failFirstProof=violation-fixture`; the exact request JSON and raw producer output with SHA-256 are linked in [36-PROHIBITION-EVIDENCE.md](./36-PROHIBITION-EVIDENCE.md).

At tested full checkout HEAD `1aa48549963797b71bd7a0a6d7d5c314509cb139`, the outcome overlay made the exact existing native root fail on `capture exhaustion changed successful work` and `capture exhaustion hid the later defect`. The causal overlay made the exact execution-peer root fail on `causal peer should accept the structurally plausible false-success trace; numeric replay is not its claim`. The direct violation TAP captures each showed one intended failure and one unaffected pass. Direct clean-fixture and absent-subject default captures each showed both Node tests passing. The canonical producer independently repeated bad RED, clean GREEN, and default GREEN for both statements. Go 1.24.0 and Node v26.10.0 ran with `GOCACHE=/tmp/ai-lang-verification-gocache`; overlays and replacement Go files were temporary under `/tmp`, removed after each run, and tracked compiler/test source remained clean and byte-unchanged.

The first adapter draft failed closed because it did not normalize indentation on Go's nested `--- FAIL` lines and rejected the expected package-level `FAIL`. No plan descriptor was resolved on that attempt. The adapter now accepts only the exact existing root and assertion set, and invalid fixture, source anchor, timeout, output, skip, compile, or unrelated failure outcomes skip both Node tests so they cannot count as a non-vacuous RED or GREEN. The current descriptors resolve only the two tested scope prohibitions. CALL-01/CALL-02 remain unclassified/unresolved, and this supplement does not close the independent security, validation, verifier, or phase gates.

Direct raw TAP receipts:

- Outcome violation: [`debug-outcome-red.log`](/tmp/schway36-prohibition-proof-1aa485499637/debug-outcome-red.log), SHA-256 `1573d1d42f143bb68238c5e5580c5fb65405ada77970ad4105cd13562a511556` (exit 1).
- Causal violation: [`debug-causal-red.log`](/tmp/schway36-prohibition-proof-1aa485499637/debug-causal-red.log), SHA-256 `0f3be5e5e61c6577c2ea1291112cb3e8e07a00c6cb56a1864672fc9c9e14b5b2` (exit 1).
- Clean control: [`clean-control.tap`](/tmp/schway36-prohibition-proof-1aa485499637/clean-control.tap), SHA-256 `4c4ad3972bf2cf3a1c86e5b0a8ba698e243ba32bbe6d03547aeef3ee70600e1e` (exit 0, two passes).
- Default/restored control: [`default-control.tap`](/tmp/schway36-prohibition-proof-1aa485499637/default-control.tap), SHA-256 `6adb8460ab747f4b25ca57531acba78269e4ab344c072ad04200fbcf08d6f25a` (exit 0, two passes).

## Prohibition supplement self-check: PASSED

- The plan descriptors point to the shared adapter and the statement-specific violation fixtures, with the same clean fixture.
- Both canonical producer result files say green/unflagged and record machine-proven `violation-fixture` evidence.
- The adapter, three JSON subjects, evidence artifact, amended plan, and this summary are the only task-owned files in the supplement; production source and Go tests, acceptance/living docs, requirements, and phase state were not changed.
