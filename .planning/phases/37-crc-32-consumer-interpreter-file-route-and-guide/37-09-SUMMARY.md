---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
plan: "09"
subsystem: testing
tags: [crc32, interpreter, mutation-testing, Go]
requires:
  - phase: 37-08
    provides: Phase 37 source and guide with prior focused owner receipts
provides:
  - Once-per-byte CRC helper placement with small-trace activation evidence
  - Reached interpreter capture-growth and file-reader negative controls
  - Repaired-source owner receipts for independent parent audit
affects: [phase-37-acceptance, phase-38]
actuals:
  tokens: 13005
  tasks: 2
  commits: 3
commits: 3
plan_head_before: c4697ec05660c510236011f442d252f73236c4d5
plan_head_after: c611acdcfa4dec959fb5b4d2850c8aed90a4e833
tech-stack:
  added: []
  patterns:
    - Bounded Go source overlays with exact semantic failure markers
    - Structural helper activation evidence kept separate from numeric truth
key-files:
  created:
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REMEDIATION-EVIDENCE.md
  modified:
    - examples/crc32/lib/crc32.schway
    - testdata/phase37/new-fixture-ledger.json
    - internal/compiler/session/session_phase37_crc32_test.go
    - internal/compiler/session/session_phase37_reuse_test.go
    - cmd/schway/main_phase37_crc32_controls_test.go
    - internal/compiler/interp/interp_phase37_test.go
    - internal/compiler/interp/filebytes_reader_test.go
    - .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md
key-decisions:
  - Keep the byte-loop activation count as structural evidence; numeric CRC vectors remain independent.
  - Require clean baseline assertions in standalone target roots and named semantic failures in armed overlay children.
patterns-established:
  - Mutation overlays must prove child root, exact target root, and actual named behavior failure.
requirements-completed: [] # APP-13, MOD-07, and APP-14 remain pending parent phase acceptance.
coverage:
  - id: D1
    description: CRC-32 calls its private eight-round byte helper once for each input byte.
    requirement: APP-13
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase37_reuse_test.go#TestPhase37CRC32ByteHelperCalls
        status: pass
    human_judgment: true
    rationale: Parent phase acceptance and requirement closure remain pending; this plan records focused evidence only.
  - id: D2
    description: Capture-growth and Go file-reader defects are rejected by reached semantic assertions.
    requirement: APP-14
    verification:
      - kind: unit
        ref: internal/compiler/interp/interp_phase37_test.go#TestPhase37CaptureRetentionControls
        status: pass
      - kind: unit
        ref: internal/compiler/interp/filebytes_reader_test.go#TestPhase37FileReaderMutationControls
        status: pass
    human_judgment: true
    rationale: Parent phase acceptance and requirement closure remain pending; this plan records focused evidence only.
duration: 21min
completed: 2026-10-15
status: complete
---

# Phase 37 Plan 09: Bounded CRC and Interpreter Remediation Summary

**CRC-32 now calls its private eight-round byte helper once per byte, and interpreter capture and reader defects have reached bounded semantic controls.**

## Performance

- **Duration:** 21 min
- **Started:** 2026-10-15T18:03:00Z
- **Completed:** 2026-10-15T18:24:22Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- Moved the eight-bit-round loop inside `crc_byte`; the outer `crc32` loop now calls it exactly once per byte. A three-byte interpreter trace observes three actual helper activations, while independent numeric vectors continue to establish CRC truth.
- Added five reached capture-growth mutations and three reached Go-reader mutations. Each overlay runs only its named root, checks the pinned project root, rejects compile or setup errors as evidence, and must fail at its matching semantic assertion.
- Kept the standalone target roots meaningful when unarmed: capture retention checks successful and late-defect runs; the reader target checks cap-plus-one, nonregular input, and close-error behavior.
- Refreshed the required affected C03/C04/C06/C07/C08/C12/C13/C14/C15/C16/C17/C18/C22/C23/C24/C26 receipts on repaired code. Exact commands, logs, source hashes, and observed root results are in [37-REMEDIATION-EVIDENCE.md](./37-REMEDIATION-EVIDENCE.md).

## Task Commits

1. **Task 1 RED: Pin CRC helper placement and activation count** — `fa27e67e` (`test`)
2. **Task 1 GREEN: Call the CRC byte helper once per input byte** — `8d0d7c6a` (`fix`)
3. **Task 2: Exercise retention and reader mutants** — `c611acdc` (`test`)

The final plan metadata commit records this summary and the sequential state/roadmap updates.

## Files Created/Modified

- `examples/crc32/lib/crc32.schway` — places eight rounds inside the private byte helper.
- `internal/compiler/session/session_phase37_reuse_test.go` and `session_phase37_crc32_test.go` — pin loop placement, helper call count, and current library hashes.
- `cmd/schway/main_phase37_crc32_controls_test.go` — updates only current finite algorithm mutation anchors and hash bindings.
- `testdata/phase37/new-fixture-ledger.json` — refreshes the Phase 37 library binding without changing historical rows.
- `internal/compiler/interp/interp_phase37_test.go` — adds per-growth retention mutations and clean baseline target assertions.
- `internal/compiler/interp/filebytes_reader_test.go` — adds wrong-bound, missing-regular-check, and ignored-close-error mutations plus clean baseline target assertions.
- `37-REMEDIATION-EVIDENCE.md` — records repair hashes, exact commands, logs, reached mutations, and receipt limits.
- `37-PROHIBITION-EVIDENCE.md` — refreshed by the required C06 proof command for current source and test bytes.

## Decisions Made

- Kept the three-byte helper activation count as structural evidence and the CRC numeric vectors as independent correctness evidence.
- Used bounded temporary Go overlays with exact unique source anchors. A mutant counts as rejected only when the actual child root reaches its named semantic failure.
- Preserved the public interpreter API and disabled-capture production path; only the CRC library implementation changed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking Evidence] Refreshed the current canonical prohibition receipt**

- **Found during:** Task 2, after the required C06 `--prove` run.
- **Issue:** The proof receipt binds current test and source digests, so it changed when rerun against the repaired revision.
- **Fix:** Committed the generated `37-PROHIBITION-EVIDENCE.md` refresh alongside the plan-local remediation receipt; historical receipts and corpora remain unchanged.
- **Files modified:** `.planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-PROHIBITION-EVIDENCE.md`
- **Verification:** C06 passed for all three projected descriptors; the refreshed file hash is recorded in `37-REMEDIATION-EVIDENCE.md`.
- **Committed in:** `c611acdc`.

**Total deviations:** 1 auto-fixed evidence update.
**Impact on plan:** Required to bind the canonical proof to repaired inputs; no feature scope or production API was added.

## Issues Encountered

Two initial reader overlay candidates did not compile because replacing their production checks left locals unused. Those runs were rejected as evidence. The mutants were adjusted to preserve compilation, then rerun until each child reached the specific wrong-result or wrong-count assertion documented in the evidence receipt.

Canonical state handlers also refreshed the generated `.planning/state.json` projection. Per parent direction, it remains unstaged and unreverted for phase-closeout reconciliation; `.planning/config.json` remained byte-identical.

## User Setup Required

None — no external service or configuration is required.

## Next Phase Readiness

The bounded remediation and focused receipts are ready for independent parent audit. Phase 37 requirement closure, security closure, validation acceptance, and phase verification remain pending; Phase 38 has not started.

---
*Phase: 37-crc-32-consumer-interpreter-file-route-and-guide*
*Completed: 2026-10-15*

## Self-Check: PASSED

- Summary, remediation evidence, and refreshed prohibition receipt exist.
- All three task commits are ancestors of the recorded plan head.
- The plan commit count is measured as 3 from `plan_head_before` to `plan_head_after`.
