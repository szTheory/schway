---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 01
subsystem: compiler-codegen
tags: [cgen, corevalidate, core, native-equivalence, ownership, borrow-checking, ffi-security]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: emitLinearForeignOutputSupport's additive-sibling emitter shape, core.ForeignContract.Alias (D-04-28), the carried D-04-33 audit debt
provides:
  - A hard byte-identity pin (core bytes, evidence-manifest IDs, generated-C goldens) over every accepting Phase 1-4 fixture
  - selectsByPointerLowering + emitLinearBorrowedByPointer: the first working by-pointer C parameter lowering, selected by a structural loan-fact predicate, never a fixture-name allowlist
  - testdata/phase5/restrict_borrow.{lang,golden.c}: the tracer fixture proving the by-pointer path end-to-end (check, corevalidate, interpreter, -O0, -O3)
  - core.ForeignContract.Alias audited independently in both cgen and corevalidate (D-04-33 closed)
affects: [05-02, 05-03, 05-04, 05-05, 05-06, 05-07]

actuals:
  tokens: 9637
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Structural selection predicate over core.LinearOperation.SourceID/TargetID/Kind chains, gated additionally on function.PublicOrigin == nil to keep a declared borrow-return function (an already-shipped, structurally identical shape) on its own existing lowering path"
    - "New emitter as an additive sibling reached only through its own predicate, never called from or calling into emitLinear/emitBranch/emitLinearForeign (D-04-20's precedent, reused verbatim)"
    - "SHA-256 literal-table byte-identity pin over a fixed corpus, demonstrated live by seeding then reverting a one-byte drift"

key-files:
  created:
    - testdata/phase5/restrict_borrow.lang
    - testdata/phase5/restrict_borrow.golden.c
  modified:
    - internal/compiler/core/core_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/cgen/export_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go

key-decisions:
  - "selectsByPointerLowering additionally requires function.PublicOrigin == nil: testdata/phase3/public_view_mixed_access.lang has the identical exclusive-borrow-then-reborrow-to-terminator operation shape as the new tracer fixture, so pure operation-stream structure alone is not sufficient to keep the new path additive -- PublicOrigin is the one genuinely structural (not name-based) fact that distinguishes an already-shipped declared borrow-return function from this plan's new plain-owned-return by-pointer path."
  - "restrict_borrow.lang declares a PLAIN `-> Buffer` return (not `-> borrow(buffer) Buffer`): check.go admits a plain return of a borrow-derived place unchanged today (the documented D-03-02 gap), which is what lets the fixture terminate on a reborrow-derived place while still reading as an ordinary owned-return function for the by-pointer lowering's own purposes."
  - "emitLinearBorrowedByPointer discards lang_record_event's return status inside the new static function (casts to void) rather than threading a return-status/out-param pair through it: the emitted LANG_EVENT_CAPACITY always exactly matches the operation count, so the failure branch is structurally unreachable for this thin slice, and adding a status-return path would be unjustified complexity for a tracer."

patterns-established:
  - "A by-pointer C emitter takes its parameter as a genuine C pointer, dereferences exactly once at the loan's own creation site, and marks its signature line with a single stable comment (`/* lang:by-pointer-param */`) for a later mutation runner's fail-closed single-marker target."

requirements-completed: [NAT-03]

coverage:
  - id: D1
    description: "Phase 1-4 core bytes, generated-C goldens, and evidence-manifest IDs are pinned by test and unchanged"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "One Phase 5 fixture (a Buffer parameter exclusively borrowed for its full body) lowers by pointer through a new additive emitter and agrees across interpreter/-O0/-O3"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestPhase5ByPointerLoweringGolden"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestPhase5ByPointerLoweringIsAdditive"
        status: pass
      - kind: integration
        ref: "internal/compiler/cgen/cgen_test.go#TestPhase5ByPointerLoweringThreeEngineAgreement"
        status: pass
    human_judgment: false
  - id: D3
    description: "core.ForeignContract.Alias is audited independently in cgen and corevalidate with a passing regression test, and no EmitForeign* output ever contains an Alias value"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestEmitForeignNeverContainsAliasValue"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestCorevalidateRefusesUnsafeAliasField"
        status: pass
    human_judgment: false
  - id: D4
    description: "No restrict or other optimizer attribute is emitted anywhere by the new by-pointer path"
    verification:
      - kind: other
        ref: "grep -v '^\\s*//' internal/compiler/cgen/cgen.go | grep -c 'restrict' == 1 (the pre-existing BannedOptimizerAttributes literal only)"
        status: pass
    human_judgment: false

duration: ~110min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 1: Native Equivalence and Adversarial Evidence -- Tracer Summary

**Additive by-pointer C parameter lowering for exclusively-borrowed Buffer parameters, selected by a structural loan-fact predicate, plus a hard Phase 1-4 byte-identity tripwire and the closed D-04-33 Alias audit gap.**

## Performance

- **Duration:** ~110 min
- **Completed:** 2026-09-06
- **Tasks:** 3 completed
- **Files modified:** 6
- **Files created:** 2

## Accomplishments

- Widened `TestPreviousPhaseCoreBytesUnchanged`/`TestPreviousPhaseManifestIDsUnchanged` to cover `testdata/phase4`'s 7 accepting fixtures (previously stopped at Phase 3), and added `TestPreviousPhaseGoldenCUnchanged` — a SHA-256 literal-table pin over every committed `*.golden.c` under `testdata/phase1`-`phase4`. Live-demonstrated by seeding a one-byte change in `owned_transfer.golden.c` (test failed, naming the exact path and both digests) then restoring it.
- Built the first working by-pointer C parameter lowering: `selectsByPointerLowering` (a structural predicate over `core.LinearOperation` chains — never a fixture-name allowlist) and `emitLinearBorrowedByPointer` (a new additive-sibling emitter modeled on `emitLinearForeignOutputSupport`, never called from or calling into `emitLinear`/`emitBranch`/`emitLinearForeign`). The new `testdata/phase5/restrict_borrow.lang` fixture — a Buffer parameter exclusively borrowed for its whole body, reborrowed twice more, returned — lowers to `static LANG_BUFFER LANG_TOUCH(LANG_BUFFER *param) { /* lang:by-pointer-param */ ... }` and agrees byte-for-byte across the interpreter, `-O0`, and `-O3` (verified directly with `clang` before committing the golden, then via `session.RunNativeFile`'s own engine-agreement assertion).
- Closed the carried D-04-33 debt (D-05-36): `core.ForeignContract.Alias` is now audited independently in both `cgen.unsafeForeignContractField` and `corevalidate.foreignContractFieldsCSafe`, with a dedicated regression test in each package. Verified the corevalidate check is load-bearing by temporarily removing it and confirming the new test fails with the expected error before restoring.

## Task Commits

Each task was committed atomically:

1. **Task 1: Pin Phase 1-4 core bytes, generated-C goldens, and evidence-manifest IDs before anything else lands** - `4ca704b` (test)
2. **Task 2: End-to-end by-pointer parameter lowering — one exclusively-borrowed fixture, one path** - `05a1c55` (feat)
3. **Task 3: Discharge the carried D-04-33 Alias pin in both independent audit layers** - `952bebd` (fix)

_Task 2 is `type="tracer"`: executed and committed like `type="auto"`, then the tracer feedback gate ran before Task 3 (auto mode re-ran Task 2's `<verify>` end-to-end — all green — before expanding)._

## Files Created/Modified

- `internal/compiler/core/core_test.go` - Widened the Phase 1-3 byte-identity pin to Phase 4; added `TestPreviousPhaseGoldenCUnchanged`
- `internal/compiler/cgen/cgen.go` - `selectsByPointerLowering`, `emitLinearBorrowedByPointer`, `borrowByPointerMarker`; wired into `Emit`/`EmitNative`; added `Alias` to `unsafeForeignContractField`
- `internal/compiler/cgen/cgen_test.go` - `TestPhase5ByPointerLoweringGolden`, `TestPhase5ByPointerLoweringIsAdditive`, `TestPhase5ByPointerLoweringThreeEngineAgreement`, `TestEmitForeignNeverContainsAliasValue`
- `internal/compiler/cgen/export_test.go` - Exported `SelectsByPointerLowering` for the additive-corpus enumeration test
- `internal/compiler/corevalidate/corevalidate.go` - Added `Alias` to `foreignContractFieldsCSafe`
- `internal/compiler/corevalidate/corevalidate_test.go` - `TestCorevalidateRefusesUnsafeAliasField`; widened the existing `commentFields` table
- `testdata/phase5/restrict_borrow.lang` - New tracer fixture (created)
- `testdata/phase5/restrict_borrow.golden.c` - Committed golden for the by-pointer lowering (created)

## Decisions Made

See `key-decisions` in frontmatter: the `PublicOrigin == nil` gate (needed to keep `public_view_mixed_access.lang` on its existing path despite an identical operation-stream shape), the plain-owned-return fixture design (relying on the documented D-03-02 gap), and discarding `lang_record_event`'s status inside the new emitted function (structurally unreachable failure for this thin slice).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Tightened selectsByPointerLowering with a PublicOrigin gate**
- **Found during:** Task 2, while implementing `TestPhase5ByPointerLoweringIsAdditive`
- **Issue:** The initial structural predicate (exclusive borrow first, unbroken reborrow chain to `OpReturn`) also matched `testdata/phase3/public_view_mixed_access.lang`, an already-shipped Phase 3 fixture with the identical operation shape — the by-pointer path was not actually additive as first written.
- **Fix:** Added `function.PublicOrigin != nil` to the predicate's early-exit conditions (a declared `borrow(...)` return function is a distinct, already-shipped semantic category) and changed the new fixture's own return type to a plain `-> Buffer` to keep it selecting the new path.
- **Files modified:** `internal/compiler/cgen/cgen.go`, `testdata/phase5/restrict_borrow.lang`, `testdata/phase5/restrict_borrow.golden.c`
- **Verification:** `TestPhase5ByPointerLoweringIsAdditive` passes for all 23 enumerated Phase 1-4 accepting fixtures, including `public_view_mixed_access.lang`.
- **Committed in:** `05a1c55` (Task 2 commit — caught and fixed before commit, not a follow-up)

---

**Total deviations:** 1 auto-fixed (1 bug, caught during the task's own acceptance-criteria loop before committing)
**Impact on plan:** Necessary for the by-pointer path to genuinely satisfy its own additivity claim (D-05-02/D-05-39). No scope creep.

## Issues Encountered

`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`) fails on the unmodified tree with `04-DEBT.md: frontmatter declares items: 4 but the Items table holds 5 rows` — confirmed pre-existing (reproduced via `git stash` before any of this plan's changes) and unrelated to any file this plan touches. Out of scope per the deviation rules' scope boundary; not fixed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The Phase 1-4 byte-identity tripwire is live and will catch any later Phase 5 plan that accidentally perturbs a prior phase's core bytes, evidence-manifest ID, or generated-C golden.
- The by-pointer lowering architecture is proven end-to-end; 05-07's mutation runner has its fail-closed single-marker target (`/* lang:by-pointer-param */`) and the seed of D-05-05's engineered fixture (the reborrow chain) is already in place.
- The `Alias` field is now closed in both audit layers; no known carried debt remains from Phase 4 blocking this phase's `restrict`-emission work (D-05-01/D-05-03), which stays explicitly out of this plan's scope for a later plan to pick up.
- No blockers.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*

## Self-Check: PASSED

All 8 created/modified key files verified present on disk; all 3 task commit hashes (`4ca704b`, `05a1c55`, `952bebd`) verified in `git log`. Plan-level `<verification>` block re-run clean (all named tests pass; `go test ./... && go vet ./...` clean except the confirmed pre-existing `TestDebtRegistersAreWellFormed` failure; `git diff --stat testdata/phase1 testdata/phase2 testdata/phase3 testdata/phase4` reports no changes).
