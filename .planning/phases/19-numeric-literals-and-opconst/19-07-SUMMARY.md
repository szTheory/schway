---
phase: 19-numeric-literals-and-opconst
plan: 07
subsystem: compiler-testing
tags: [u64, opconst, exhaustive-dispatch, differential-testing, scalar-goldens]
requires:
  - phase: 19-05
    provides: Independently admitted canonical OpConst and U64 core facts
  - phase: 19-06
    provides: Interpreter and native execution for exact U64 constants
provides:
  - OpConst registered and exercised by in-process and CLI-observable six-consumer controls
  - Four-tier, five-axis checks for decimal, zero, maximum, hexadecimal, and binary literals
  - Scalar execution and frozen generated-C baseline review
affects: [phase-19-verification, compiler-regression-controls]
actuals:
  tokens: 5929
  tasks: 2
  commits: 2
tech-stack:
  added: []
  patterns: [Operation-kind registry drives exhaustive dispatch; execution equality is paired with an explicit expected-value assertion]
key-files:
  created:
    - .planning/phases/19-numeric-literals-and-opconst/19-SCALAR-REVIEW.md
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/core/core_convention_absence_test.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase19_test.go
    - testdata/phase16/public-emitter-consumers.json
key-decisions:
  - "Use the existing schema-2 comparison projection for the interpreter before comparing all four execution tiers."
  - "Keep the stale Phase 11 maturity and reconciliation findings as regression debt; they are outside this plan's scope."
patterns-established:
  - "Exhaustive dispatch fixtures assert OpConst's canonical source-free root facts and observed value."
  - "Four-tier equality must be paired with an expected scalar value to detect shared wrong results."
requirements-completed: [VAL-01, VAL-02, VAL-03]
coverage:
  - id: D1
    description: OpConst is registered and a real literal fixture is driven through all six consumers in both exhaustive controls.
    requirement: VAL-02
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/core ./internal/compiler/session -run 'Test(AllOperationKinds|Phase7DispatchControlsMutationKilled|Phase19Dispatch)' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Decimal, zero, max U64, hexadecimal, and binary literals return exact decimal values on all four tiers and compare across five axes.
    requirement: VAL-03
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session ./internal/compiler/core -run 'TestPhase19(FourTier|LiteralRun|WrongResult|Dispatch)|TestPayloadCorpusCharacterizationReplay|TestAllOperationKindsHandledAtEverySite' -count=1"
        status: pass
    human_judgment: false
  - id: D3
    description: Public interpreter and native run commands return the literal value; scalar and frozen C baselines remain unchanged.
    requirement: VAL-01
    verification:
      - kind: integration
        ref: "TestPhase19LiteralRun; TestPayloadCorpusCharacterizationReplay; TestPreviousPhaseGoldenCUnchanged; TestLegacyEmitterEvidence"
        status: pass
    human_judgment: false
duration: 17min
completed: 2026-09-25
status: complete
plan_head_before: ea6e4f12341acc684f75474c25364ed3cfbf181b
---

# Phase 19 Plan 07: Exhaustive OpConst and Four-Tier Evidence Summary

**OpConst is registered and proven through both exhaustive controls, while five literal forms return exact U64 values across interpreter and three native optimization tiers.**

## Performance

- **Duration:** 17 min
- **Started:** 2026-09-25T00:49:00Z
- **Completed:** 2026-09-25T01:06:28Z
- **Tasks:** 2/2
- **Files modified or created:** 7

## Accomplishments

- Registered `OpConst` in `AllOperationKinds()` and extended the in-process and CLI-observable dispatch controls with the real Phase 19 literal fixture.
- Added operation-specific checks for canonical constant facts, interpreter output, C lowering, path-oracle loan roots, and independently recomputed return origin.
- Compared decimal `42`, zero, maximum U64, hexadecimal `0x2A`, and binary `0b10_1010` across interpreter, `-O0`, `-O3`, and `-O3 -flto`; each returned the expected decimal string and passed the five-axis all-pairs comparator.
- Added a shared-wrong-result negative control and public interpreter/native command smoke tests.
- Replayed 62 scalar digest baselines and 35 expected skips; reviewed the four frozen generated-C SHA-256 baselines. No golden changed. See [19-SCALAR-REVIEW.md](./19-SCALAR-REVIEW.md).

## Task Commits

1. **Task 1: Register OpConst and make both exhaustive controls mutation-sensitive** — `881a784` (`feat`).
2. **Task 2: Compare four tiers and review scalar evidence** — `c9232a7` (`test`).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Integration regression] Updated the `LinearOperation` structural field inventory**
- **Found during:** Task 2 full-suite verification.
- **Issue:** The complete-field inventory still expected the pre-`ConstU64` shape and failed after Phase 19 added the semantic constant field.
- **Fix:** Added `ConstU64` to the expected field sequence.
- **Files modified:** `internal/compiler/core/core_convention_absence_test.go`.
- **Verification:** The core package passed in the subsequent full-suite run.
- **Committed in:** `c9232a7`.

**2. [Rule 1 - Integration regression] Refreshed moved emitter call locations**
- **Found during:** Task 2 full-suite verification.
- **Issue:** Adding dispatch evidence moved three existing `cgen.Emit` source locations, leaving their line-numbered registry rows stale.
- **Fix:** Updated those three rows in the Phase 16 public-emitter registry. Its remaining failures concern unchanged `cgen_program_test.go` locations and are recorded below as baseline debt.
- **Files modified:** `testdata/phase16/public-emitter-consumers.json`.
- **Verification:** Targeted inventory now reports only the pre-existing `cgen_program_test.go` drift; no stale rows remain for files changed by this plan.
- **Committed in:** `c9232a7`.

**Total deviations:** 2 auto-fixed integration issues.
**Impact on plan:** Both fixes keep current structural and source-inventory controls aligned with Phase 19 evidence; no product scope changed.

## Verification

- Focused dispatch, four-tier, public-run, wrong-result, and scalar replay tests — PASS.
- Four frozen generated-C digest checks and the Phase 16 frozen-emitter ledger — PASS.
- `go test ./...` — FAILS on existing repository regression debt: stale `.planning/LANGUAGE-MATURITY.md` corpus and guard totals; unreconciled Phase 19 validation-frontier entries; and stale public-emitter inventory locations in unchanged `internal/compiler/cgen/cgen_program_test.go`. The same full-suite session run also reports the previously recorded Phase 11 maturity/reconciliation failures. No Phase 19 focused test failed. The convention-field inventory failure found on the first full-suite run was fixed and passes now.

## Issues Encountered

The sandbox initially denied Go's default build cache under `~/Library/Caches` and Git index writes under `.git`. Setting `GOCACHE=/private/tmp/go-build-ai-lang` resolved test execution. Repository commits succeeded through the authorized GSD helper with repository-write escalation; this confirmed the Git failure was a Codex sandbox boundary, not a worktree or GSD defect.

The GSD roadmap progress updater has previously returned `missing_phase_details` for Phase 19. Per orchestration guidance, ROADMAP was not edited directly.

## Next Phase Readiness

All Plan 19-07 tasks and VAL-01 through VAL-03 evidence are complete. Phase 19 full-suite debt remains visible for later reconciliation; it does not affect the passing focused Phase 19 controls.

## Self-Check: PASSED

- Summary file exists at the required path.
- Task commits `881a784` and `c9232a7` exist in Git history.
- The plan ledger records base `ea6e4f12341acc684f75474c25364ed3cfbf181b`; two task commits are measured between that base and current HEAD.

---
*Phase: 19-numeric-literals-and-opconst*
*Completed: 2026-09-25*
