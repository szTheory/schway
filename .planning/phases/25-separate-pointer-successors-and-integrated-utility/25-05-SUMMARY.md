---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "05"
subsystem: compiler/native
tags: [Go, C, ownership, pointer-lowering, pathoracle, originvalidate]

requires:
  - phase: 25-04
    provides: "The bounded transfer utility and Phase 24 cleanup behavior used by the integrated caller."
provides:
  - "An integrated shared-copy then exclusive-copy utility with independent path and origin proofs."
  - "Exact shared U64 pointer ABI admission and fail-closed application lowering."
  - "Native actual-C evidence for results 65/66, typed 0x43 failure ordering, and reached wrong-result controls; source controls cover conflicts and escapes."
affects: [25-06, cgen, pathoracle, originvalidate, native-utility]

actuals:
  tokens: 20343
  tasks: 1
  commits: 11
plan_head_before: 66895e88de36b0bcefe6619024c119dacab38079

tech-stack:
  added: []
  patterns:
    - "One checked pointer ABI fact feeds declaration, call, body, and manifest emission."
    - "Independent peers recognize the same exact owner-result chain and preserve a bounded legacy route."

key-files:
  created:
    - internal/compiler/native/phase25_utility_test.go
    - internal/compiler/pathoracle/pathoracle_pointer_successor_test.go
    - internal/compiler/originvalidate/originvalidate_pointer_successor_test.go
  modified:
    - examples/phase24/transfer.schway
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/cgen/cgen_pointer_successor_test.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/originvalidate/originvalidate.go

key-decisions:
  - "Admit only the exact shared U64 borrow-copy-return ABI shape through one checked fact."
  - "Keep the Phase 24 direct-result route limited to its two-function, four-operation caller while admitting the exact Phase 25 shared-to-exclusive result chain."
  - "Fail malformed pointer-lowering candidates before serialization, while preserving PublicOrigin borrowed views and Match callback refusal behavior."

patterns-established:
  - "Use actual emitted-C execution for each pointer family and keep wrong-result, conflict, and escape controls distinct."
  - "Retain source-compatible Phase 24 behavior through an explicit direct-shape regression fixture."

requirements-completed: []
coverage:
  - id: D1
    description: "The integrated utility executes through native C, returns the pinned successful values, reaches wrong-result controls, and preserves the typed error before helper calls."
    verification:
      - kind: integration
        ref: "internal/compiler/native: TestPhase25UtilityNative and TestPhase25NativeFamilyWrongResult"
        status: pass
      - kind: unit
        ref: "internal/compiler/native: TestPhase25UtilityFamilyControls (separate shared/exclusive conflict and escape source controls)"
        status: pass
      - kind: integration
        ref: "POSIX sh clean-checkout CLI smoke: 0x41=65, 0x42=66, 0x43=UseError.UnsupportedByte (exit 65)"
        status: pass
    human_judgment: false
  - id: D2
    description: "C generation, path replay, and origin validation admit the same bounded owner-result chain and refuse malformed extensions."
    verification:
      - kind: unit
        ref: "internal/compiler/cgen: TestPhase25SharedPointerCopyABI, TestPhase25ExclusivePointerRefusal, TestPhase5ByPointerLoweringIsAdditive"
        status: pass
      - kind: unit
        ref: "internal/compiler/pathoracle and internal/compiler/originvalidate: TestPhase25UtilityOwnerTransfer"
        status: pass
    human_judgment: false

duration: 219min
completed: 2026-10-01
status: complete
---

# Phase 25 Plan 05: Integrated Pointer Utility Summary

**The integrated shared-to-exclusive pointer utility now has fail-closed C lowering and matching independent owner-transfer proofs.**

## Performance

- **Duration:** 3h 39m
- **Started:** 2026-10-01T17:34:21-04:00
- **Completed:** 2026-10-01T21:13:12-04:00
- **Tasks:** 1
- **Files modified:** 10

## Accomplishments

- Extended `examples/phase24/transfer.schway` through shared copy, exclusive copy, and return. The native utility tests execute emitted C for both families, prove 65/66 results, reach both wrong-result controls, and prove the 0x43 typed use error occurs before either pointer helper. Separate source controls cover conflict and escape for each family.
- Added exact shared U64 copy ABI and integrated caller admission in cgen. The checked ABI fact drives the pointer declaration, call-site address, scalar copy body, and manifest. The direct Phase 24 lowering route stays pinned to its original two-function/four-operation shape.
- Added independent pathoracle and originvalidate proofs for the exact shared-copy then exclusive-copy result chain, including swapped, arbitrary, missing, and tampered-chain refusals.
- Added a pre-serialization refusal for malformed pointer-lowering candidates after the strict selectors reject their broken chains. The guard preserves public borrowed views (`PublicOrigin`) and leaves `Match` eligible for callback refusal.

## Evidence status

Source inspection confirms the admission predicates remain shape-specific and the public-origin carve-out remains outside the new candidate refusal. The focused local receipts below were executed on the current host. They are not hosted macOS/Linux Phase 25 receipts.

The exact Plan05 checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^(TestPhase24EmitterTransferRefusesUnsupportedCoreBeforeSerialization|TestPhase25SharedPointerCopyABI|TestPhase25ExclusivePointerRefusal)$' ./internal/compiler/cgen`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase25UtilityOwnerTransfer' ./internal/compiler/pathoracle ./internal/compiler/originvalidate`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase25(Utility|NativeFamilyWrongResult)' ./internal/compiler/native`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^TestPhase5ByPointerLoweringIsAdditive$|^TestPhase25ExclusivePointerRefusal$' ./internal/compiler/cgen`
- `git diff --check`

A clean-checkout command sequence was exercised with POSIX `sh`: the application built with explicit source and manifest inputs; inputs `0x41` and `0x42` returned 65 and 66, and `0x43` reported `UseError.UnsupportedByte` and exited 65 before pointer helpers. The generated binary, inputs, and manifest were removed after the smoke run.

The historical hosted Phase 24 receipt remains as recorded in STATE.md; it establishes the prior transfer/cleanup behavior only. Plan05 produced no hosted Phase 25 host/lane receipt. Plan06 owns the host-by-family evidence index, CI integration, README, and living-document reassessment; applicable Phase 25 requirements therefore remain open.

An optional broader changed-package sweep was not clean and is recorded as an observed limitation, not a passed gate. It flagged `native.TestSourceNeverSpawnsUnboundedProcesses` on the Plan06-owned pointer test's unbounded subprocess call, and the Plan06-owned `TestPhase25EvidenceScript`, `TestPhase25EvidenceIndex`, and `TestPhase25LivingRoadmap` checks before that plan's artifacts were complete. The same sweep also surfaced existing Phase 24 observer, pathoracle terminator, and originvalidate mixed-access cases outside this Plan05 acceptance gate. These were not altered or claimed as passing here.

## Task Commits

Plan05 implementation/test commits:

1. **Native utility RED evidence** — `f41185e` (test)
2. **Shared U64 ABI RED evidence** — `1f66430` (test)
3. **Integrated utility and peer refusal coverage** — `aa7541b` (test)
4. **Integrated pointer successor utility** — `ec32dbd` (feat)
5. **Exclusive pointer refusal controls** — `a44a30f` (test)
6. **Malformed candidate fail-closed guard** — `e18886c` (fix)

The measured plan ledger contains 11 commits total, including five plan/scope amendment commits. The closeout metadata commit is made after this summary is written.

## Files Created/Modified

- `examples/phase24/transfer.schway` — integrated shared/exclusive copy utility.
- `internal/compiler/native/phase25_utility_test.go` — actual-C positive, wrong-result, conflict, escape, and error-order controls.
- `internal/compiler/cgen/` — checked ABI integration, caller admission, compatibility fixture, and refusal/public-view regressions.
- `internal/compiler/pathoracle/` and `internal/compiler/originvalidate/` — independent result-chain recognition and mutation tests.

## Decisions Made

- Keep pointer ABI admission and caller lowering narrowly tied to checked operation chains, deriving all C-facing details from one fact.
- Preserve the shipped Phase 24 direct-result shape and the additive `PublicOrigin` borrowed-view path explicitly.
- Refuse malformed candidates before C serialization even when a broken chain no longer matches the stricter structural selector.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Refuse malformed exclusive pointer candidates before serialization**
- **Found during:** Task 1 (package-sweep review)
- **Issue:** Forwarding, retention, and wider-pointer bodies failed the exact selectors but could fall through to generic C serialization.
- **Fix:** Added a narrow first-borrow candidate guard in production emission; it exempts existing `PublicOrigin` borrowed views and does not exempt `Match` callback candidates.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_program.go`
- **Verification:** `TestPhase25ExclusivePointerRefusal` and `TestPhase5ByPointerLoweringIsAdditive` passed with the full focused Plan05 gates.
- **Committed in:** `e18886c`

**Total deviations:** 1 auto-fixed (Rule 1 - Bug). The remaining scope was adjusted by the orchestrator to move README, evidence automation, CI, and living-document closeout to Plan06.

## Issues Encountered

The direct Phase 24 emitter test helper initially followed the now-integrated public fixture and could no longer pin the historical operation indexes. It was changed to synthesize the explicit two-function/four-operation direct fixture, then the exact compatibility and refusal controls passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The compiler, native utility, and independent peer implementation for Plan05 are complete. Plan06 must finish README/evidence-script/CI integration, re-check the living roadmap and maturity recommendations against current sources, and obtain any available hosted host/lane receipts. Phase 25 requirements remain open until their applicable acceptance evidence is complete.

---
*Phase: 25-separate-pointer-successors-and-integrated-utility*
*Completed: 2026-10-01*

## Self-Check: PASSED

- Summary file exists at the required phase path.
- All six listed Plan05 implementation/test commits are present in git history.
- No stub or placeholder patterns were found in the Plan05-owned files.
