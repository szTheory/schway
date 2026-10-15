---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "08"
subsystem: testing
tags: [crc32, guide, tracked-source, evidence, transition]

# Dependency graph
requires:
  - phase: "37-05"
    provides: independent CRC vectors and expected-first public outcomes
  - phase: "37-06"
    provides: reached CRC mutations and module/build identity boundaries
  - phase: "37-07"
    provides: public file-run route, exit facts, and scoped surface evidence
provides:
  - tracked-only executable CRC-32 guide for both programs and public engines
  - source-bound C01-C26 focused owner capture with raw local logs and digests
  - current Phase37 transition and conditional-seed dispositions
affects: [phase-37-parent-acceptance, phase-38, DX-24, APP-13, MOD-07, APP-14]

# Actuals (#2632), chars/4 over realized plan-owned diff.
actuals:
  tokens: 22550
  tasks: 3
  commits: 4
plan_head_before: 6abff762d73acfc4b416c62fc005fd7faa4d0014
plan_head_after: 727842b243a372305b8b84e11b0d0f7dc7f95b08

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Execute marked guide fences from a git-ls-files-only private source copy.
    - Compare guide tuples and documentation drift against independent fixture rows.
key-files:
  created:
    - examples/crc32/README.md
    - cmd/schway/main_phase37_guide_test.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-ACCEPTANCE-EVIDENCE.md
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-TRANSITION-CHECK.md
  modified:
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-VALIDATION.md
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/STATE.md
    - .planning/ROADMAP.md
key-decisions:
  - "Keep Phase37 requirements pending until the parent verifier, security, and review gates accept them."
  - "Treat native executable digest and BuildID variance under the existing incomplete/non-cacheable receipt contract; make no deterministic-binary claim."
  - "Keep Phase38 hosted binding after milestone audit and gap closure."
patterns-established:
  - "A guide claim is exercised by running its actual marked commands from a tracked-only temporary copy."
  - "Evidence distinguishes independent numeric outcomes, causal structure, physical cleanup, source inspection, and historical receipts."
requirements-completed: []

# Objective guide and documentation acceptance; independent parent phase gates remain pending.
coverage:
  - id: D1
    description: "The executable CRC guide builds and runs both consumers on interpreter, one-shot native, and retained native routes."
    requirement: DX-24
    verification:
      - kind: integration
        ref: "C22: TestPhase37GuideCleanCheckout"
        status: pass
      - kind: integration
        ref: "C24: TestPhase37GuideCleanCheckout"
        status: pass
    human_judgment: false
  - id: D2
    description: "The documented tuples match the independent JSON fixture and armed parser, command, limit, error, and untracked-helper controls."
    requirement: DX-24
    verification:
      - kind: unit
        ref: "C23: TestPhase37GuideMatchesExpectedJSON and TestPhase37GuideParserNegativeControls"
        status: pass
    human_judgment: false
  - id: D3
    description: "The focused owner map and canonical prohibition evidence are bound to one code revision with exact logs and skip dispositions."
    verification:
      - kind: integration
        ref: "37-ACCEPTANCE-EVIDENCE.md C01-C26; C06 canonical producer"
        status: pass
    human_judgment: false

# Metrics
duration: 24min
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 08: Executable CRC-32 Guide and Local Evidence Summary

**A tracked-only clean checkout now runs the CRC guide and matches 78 exact tuples across both consumers, both engines, one-shot native, and retained native execution.**

## Performance

- **Duration:** 24 minutes
- **Started:** 2026-10-15T17:13:23Z
- **Completed:** 2026-10-15T17:37:10Z
- **Tasks:** 3
- **Plan-owned files modified:** 8

## Accomplishments

- Added a root-relative guide with Go 1.24/Clang/Python prerequisites, actual binary fixture setup, marked CLI commands, exact tuples and fixed errors, CRC variant distinctions, cap behavior, and scoped limitations.
- Added a closed Phase37 guide parser and tracked-only copier. Its clean-copy run hashed 2,691 tracked paths and passed all 78 tuple/route combinations. JSON agreement, malformed/duplicate/unknown fence, answer/limit/error drift, wrong runnable input, and untracked-helper controls pass.
- Captured every C01-C26 focused owner command at code revision `b900dfe31abb600a82b31ab6e8cc2a8b9b1e74ca`: 63 Go root invocations passed; C05 passed 3/3 Node tests; C06 refreshed the three canonical prohibition descriptors and their bad/clean/default subjects. Raw logs, hashes, source/fixture digests, and evidence boundaries are recorded in [37-ACCEPTANCE-EVIDENCE.md](37-ACCEPTANCE-EVIDENCE.md).
- Updated the validation map, added the dated transition check, and amended both living documents with current source facts, local evidence, historical-receipt boundaries, and three scoped next capabilities.
- Kept APP-13, MOD-07, APP-14, and DX-24 pending parent acceptance. Phase38 hosted receipts remain after audit and gap closure.

## Task Commits

1. **Task 1: Run the actual guide from tracked files only** — `78db2da4` (RED contract), `6c979c03` (guide and runner).
2. **Task 2: Arm guide drift and fixture agreement controls** — `b900dfe3` (unknown-block-kind control; shared parser/table controls also landed with the Task 1 runner).
3. **Task 3: Bind focused evidence and amend the living transition facts** — `727842b2` (evidence and transition documents).

## Files Created/Modified

- `examples/crc32/README.md` — executable build/run guide and exact output table.
- `cmd/schway/main_phase37_guide_test.go` — tracked-source copy, bounded marked-command execution, tuple assertions, and armed drift controls.
- `37-ACCEPTANCE-EVIDENCE.md` — fixed-source C01-C26 command table, log hashes, source/fixture digests, and claim limits.
- `37-TRANSITION-CHECK.md` — seven-column capability recommendations and seed dispositions.
- `37-VALIDATION.md` — corrected C07 selector and actual focused-owner status.
- `37-PROHIBITION-EVIDENCE.md` — fresh canonical producer evidence at the captured code revision.
- `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md` — dated source/evidence transition amendments.

## Decisions Made

- The guide uses only `git ls-files` inputs and executes the marked commands; the earlier Phase32 copier was not reused for this tracked-only claim.
- One maximum-size read is recorded as one local observation, not a timing distribution. No full-loop cold/warm latency or human readability claim is inferred.
- Native BuildID may vary when executable bytes differ while module rows and InputID remain stable under the existing incomplete/non-cacheable receipt contract.
- Phase38 remains the owner for hosted two-host binding after the milestone audit and any gap closure.

## Deviations from Plan

### Auto-fixed issues

**1. [Rule 1 - Bug] Parse Markdown tuple cells with code ticks**
- **Found during:** Task 1
- **Issue:** The first table parser treated surrounding Markdown backticks as part of an expected tuple.
- **Fix:** Normalize the tuple cell before parsing and comparing it with the inert JSON fixture.
- **Files modified:** `cmd/schway/main_phase37_guide_test.go`
- **Verification:** C22 and C24 passed the full 78-tuple guide run.
- **Committed in:** `6c979c03`

**2. [Rule 1 - Test control] Target an unknown run fence in the unknown-kind negative control**
- **Found during:** Task 2
- **Issue:** The first mutation changed the required setup marker, so the parser correctly failed earlier on a missing setup fence.
- **Fix:** Mutate one run marker so the parser succeeds and the closed-structure assertion rejects the unknown block kind.
- **Files modified:** `cmd/schway/main_phase37_guide_test.go`
- **Verification:** C23 passed the named `unknown-block-kind` control and the complete Phase32/Phase37 parser roots.
- **Committed in:** `b900dfe3`

**3. [Rule 3 - Blocking validation issue] Correct C07's nonexistent test selector**
- **Found during:** Task 3
- **Issue:** The planned `TestPhase37CaptureDisabledRetention` root did not exist; the interpreter package reported “no tests to run” while the session root passed.
- **Fix:** Updated C07 to select `TestPhase37CaptureRetentionControls`, then reran it with `TestPhase37InterpreterFileApplication`.
- **Files modified:** `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-VALIDATION.md`
- **Verification:** Corrected C07 reached and passed 2/2 intended package roots. The rejected initial log and corrected log are both preserved by path and digest.
- **Committed in:** `727842b2`

### Task-order overlap

Task 1 and Task 2 share one closed parser/copier test owner. The JSON and core
drift controls were added with the Task 1 runner because that harness needed
them to establish the tracer slice; the exact Task 2 command was then run, and
Task 2 added a separate unknown-kind control. This preserves the single-owner
design but means the full Task 2 RED/GREEN sequence was not split into its own
initial test commit.

**Total deviations:** 3 auto-fixed issues and 1 documented task-order overlap.  
**Impact on plan:** The selector typo was corrected before final evidence capture; all current exact owner commands pass. No feature scope or dependency was added.

## Issues Encountered

C07 initially exited successfully despite an empty package selection. It was
excluded from passing evidence, recorded as a validation-map defect, corrected,
and rerun. The final map and evidence receipt name only the reached owner roots.

## User Setup Required

None.

## Next Phase Readiness

The guide and focused evidence are ready for the parent-owned Phase37
verifier, security, review, and requirement gates. The raw probe/classifier
history remains intact; all four requirements are still pending acceptance.
Do not start Phase38 until the milestone audit and any gap closure finish.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*  
*Plan: 08*  
*Completed: 2026-10-15*

## Self-Check: PASSED

- All eight plan-owned source/evidence files and this summary exist.
- All four task commits are ancestors of the recorded plan head.
- The final focused capture and C07 correction are recorded in the acceptance and validation receipts.
