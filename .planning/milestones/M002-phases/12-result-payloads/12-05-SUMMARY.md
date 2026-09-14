---
phase: 12-result-payloads
plan: 05
subsystem: compiler-payload-representation
tags: [cgen, session, native, layout, mutation-testing, D-12-37, D-12-38, D-12-41, D-12-43]

# Dependency graph
requires:
  - phase: 12-result-payloads (plan 02)
    provides: OpConstructPayload/OpDestructurePayload, check.PayloadRecordLayout, testdata/phase12/payload_tracer.lang, the ratified published surface
  - phase: 12-result-payloads (plan 03)
    provides: testdata/phase12/payload_borrow_interaction.lang (the borrow-carrying payload fixture this plan's companion control depends on)
  - phase: 12-result-payloads (plan 04)
    provides: corpus characterization replay, originvalidate/pathoracle payload support
provides:
  - cgen.EmitPayloadConformance / session.PayloadLayoutMutationRunner (D-12-37's necessary-but-insufficient frozen-fixture C-layout control)
  - testdata/phase12/payload_layout_mismatch.golden.c (frozen, deliberately transposed; never a correct version committed)
  - cgen.SetPayloadSlotSwapForTest, a production-visible cross-package mutation seam, and TestPayloadSlotSwapMutationKilled (D-12-38's attempted decisive control)
  - The empirically-measured D-12-43 finding: D-12-38's value-divergence control is unconstructible against the current representation (escalated per D-12-41, recorded in PHASE-12-DEBT.md, not silently downgraded)
  - PHASE-12-DEBT.md's D-12-26 row (criterion 2's precise "one meaning" scope) and D-12-43 row (the escalated control-unconstructibility finding), plus D-12-24's confirmed-unchanged niche finding
  - 12-VALIDATION.md's completed Per-Task Verification Map (all five plans) and wave_0_complete: true
affects: []

# Actuals (#2632)
actuals:
  tokens: 33000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A payload-side conformance control (cgen.EmitPayloadConformance) clones EmitForeignConformance's shape exactly, but keys off check.PayloadRecordLayout(core.DataType) instead of a core.ForeignContract -- the same _Static_assert/offsetof generation logic, applied to a checker-derived fact with no foreign-symbol boundary at all"
    - "A purpose-built, single-byte-field probe DataType (session.PayloadProbeDataType, mirroring LayoutProbeContract's own two-one-byte-field design) is used for the frozen-fixture control instead of the production tracer's Outcome/Fault shape, specifically to avoid LANG_BUFFER's own real (padded, platform-dependent) size becoming an accidental confound in the transposition being tested"
    - "A mutation seam whose driving test lives in a DIFFERENT package than the mutated production code (cgen.SetPayloadSlotSwapForTest, driven from session_test) must be a production-visible function in a real .go file, never an export_test.go symbol -- export_test.go is only visible within the SAME package's own test binary (12-04-SUMMARY.md's own precedent, applied again here)"
    - "A control that cannot be constructed against the current architecture is proven so empirically (build the real mutation, run it, observe no divergence), then PINNED as a genuine passing regression test recording the absence -- mirroring TestC03PeerDeriveOriginFactsOpCallGapStillOpen's own gap-pinning precedent -- rather than silently downgraded to a weaker assertion or silently marked done"

key-files:
  created:
    - testdata/phase12/payload_layout_mismatch.golden.c
    - internal/compiler/session/session_payload_control_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/session/session.go
    - .planning/phases/12-result-payloads/PHASE-12-DEBT.md
    - .planning/phases/12-result-payloads/12-VALIDATION.md

key-decisions:
  - "session.PayloadProbeDataType() (two Byte-payload alternatives, both C type unsigned char) is a purpose-built fixture for Task 1's frozen-layout control, deliberately NOT reusing payload_tracer.lang's Outcome/Fault (Buffer/nullary) shape -- Buffer's real C struct size (padded via its size_t length field) would have made the transposition test depend on platform ABI details unrelated to what is being proven, exactly the reasoning LayoutProbeContract's own doc comment gives for its two-one-byte-field design."
  - "D-12-38's decisive control was built for real (a genuine, type-safe wrong-alternative-payload-slot write, seeded via a real production-code mutation, not a stub), then empirically PROVEN unconstructible against the current representation: a payload-carrying return's Outcome.Value is always the alternative's compile-time-known tag name on BOTH engines (cgen's returnLiteral is a literal string baked into the generated C at emission time; interp's value.String() resolves through the tag field), so a wrong-slot payload WRITE never reaches any of the five comparator axes. Escalated per D-12-41/D-11-36 as a defect in the criterion (PHASE-12-DEBT.md's new D-12-43 row), not silently downgraded to a compile-failure-only assertion."
  - "TestPayloadSlotSwapMutationKilled's companion beat drives payload_borrow_interaction.lang's own 'identity' function in isolation (a new isolateNativeFunction test helper extracts one function plus its transitively-needed DataTypes from the checked 2-function program), because native emission does not support multi-function branch bodies this phase (D-11-52/D-12-32) and the fixture deliberately declares two functions in one program."
  - "cgen.SetPayloadSlotSwapForTest is a production-visible function in cgen.go itself, not an export_test.go symbol, because TestPayloadSlotSwapMutationKilled lives in session_test (a different package) -- following interp.SetDisableEmptyTagSerializationForTest's D-07-42/D-12-38 cross-package precedent from 12-04-SUMMARY.md, not SetOpCallGroupedArmForTest's same-package export_test.go shape (which the plan's own read_first named as the shape to follow structurally, but which is unreachable across packages)."

requirements-completed: [RES-03]

coverage:
  - id: D1
    description: "D-12-37's frozen-fixture payload layout control: a committed, deliberately transposed C struct declaration is refused at compile time by the generated _Static_assert/offsetof pair, with its own doc comment naming D-12-38 as the control that covers its structural gap"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_payload_control_test.go#TestPayloadLayoutMutationRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_payload_control_test.go#TestPayloadLayoutMutationAttacksFrozenFixtureOnly"
        status: pass
    human_judgment: false
  - id: D2
    description: "D-12-38's decisive slot-swap mutation control was built and run for real; empirically measured to be unconstructible against the current representation, escalated as a defect in the criterion per D-12-41 rather than silently downgraded or skipped"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_payload_control_test.go#TestPayloadSlotSwapMutationKilled"
        status: pass
    human_judgment: false
  - id: D3
    description: "RES-03's niche-optimization clause remains recorded as an absence-of-applicable-input finding (uninstantiable), confirmed unchanged by this plan's two controls; criterion 2's 'one meaning' claim is stated at the strength it can actually be held (observable-behavior agreement, not byte-identical layout)"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
  - id: D4
    description: "The five-axis comparator gained no sixth axis and no new comparator field; session_phase5_compare.go is byte-unchanged"
    verification:
      - kind: unit
        ref: "git diff --quiet -- internal/compiler/session/session_phase5_compare.go"
        status: pass
    human_judgment: false
  - id: D5
    description: "All four pinned golden-C digests remain byte-identical after this plan's diff"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseGoldenCUnchanged"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-13
status: complete
---

# Phase 12 Plan 05: Criterion 2's Anti-Vacuity Controls Summary

**A real frozen-fixture C-layout control (D-12-37) plus a genuinely-built, genuinely-run wrong-slot mutation control (D-12-38) that was empirically proven unconstructible against the current representation and escalated as a recorded criterion defect (D-12-43), never a quiet downgrade — closing Phase 12's RES-03 requirement.**

## Performance

- **Duration:** ~55 min
- **Started:** 2026-09-13T03:15:00Z (approx.)
- **Completed:** 2026-09-13T04:10:00Z (approx.)
- **Tasks:** 3
- **Files modified:** 4 modified, 2 created

## Accomplishments

- `cgen.EmitPayloadConformance` and `session.PayloadLayoutMutationRunner` clone `EmitForeignConformance`/`LayoutMutationRunner`'s proven shape for `check.PayloadRecordLayout`, proving a deliberately transposed, frozen `testdata/phase12/payload_layout_mismatch.golden.c` is refused at compile time under the project's `-Werror` flag set — `TestPayloadLayoutMutationRefused` plus a companion `TestPayloadLayoutMutationAttacksFrozenFixtureOnly` proving the runner's verdict depends solely on the fixture's own content (an untransposed temp-file header compiles cleanly).
- `cgen.SetPayloadSlotSwapForTest` seeds a real, type-safe wrong-alternative-payload-slot write in `OpConstructPayload`'s codegen (correct tag, payload written into a *different* alternative's own struct field). `TestPayloadSlotSwapMutationKilled` drives this mutation against plan 02's tracer fixture and a companion, unmutated run of plan 03's `payload_borrow_interaction.lang` (isolated to its `identity` function, since native emission does not support multi-function branch bodies this phase).
- **Empirical finding:** the mutated beat produces **no disagreement on any of the five axes**. A payload-carrying return's `Outcome.Value` is, on both engines, always the alternative's own compile-time-known tag name — `cgen`'s `returnLiteral` is a literal baked into the generated C at emission time, never read back from the mutated struct; `interp`'s `value.String()` resolves through the tag field, never the payload bytes. This is not a flaw in the mutation's construction; it is a structural property of the current grammar (a bare-value match arm's result can only be an alternative name, never raw payload content).
- Per D-12-41/D-11-36, this is escalated as a **defect in the criterion** — recorded in `PHASE-12-DEBT.md`'s new `D-12-43` row — rather than silently downgraded to a weaker (e.g. compile-failure-only) assertion. `TestPayloadSlotSwapMutationKilled` pins the absence as a genuine passing regression test, mirroring `TestC03PeerDeriveOriginFactsOpCallGapStillOpen`'s own gap-pinning precedent.
- `PHASE-12-DEBT.md` gains `D-12-26` (criterion 2's "one meaning" claim is observable-behavior agreement, never byte-identical layout, because `interp` has no byte layout at all) and confirms `D-12-24`'s niche-uninstantiability finding unchanged by this plan's controls.
- `12-VALIDATION.md`'s Per-Task Verification Map is filled in for all five plans' real task IDs and `<automated>` commands; every Wave 0 item is checked off and `wave_0_complete: true` is set. `status`/`nyquist_compliant` are left untouched for `/gsd-validate-phase`.
- `session_phase5_compare.go` is byte-unchanged (no sixth axis, no new comparator field); all four pinned golden-C digests remain byte-identical; `go test ./...` exits 0; `go vet ./...` clean; every file this plan touched is `gofmt`-clean.

## Task Commits

Each task was committed atomically:

1. **Task 1: D-12-37's frozen-fixture layout control** - `66dd2ee` (feat)
2. **Task 2: D-12-38's decisive control — built, run, and escalated as D-12-43** - `b475d41` (test)
3. **Task 3: Record criterion 2's findings** - `5559929` (docs)

**Plan metadata:** (this commit, following SUMMARY creation)

## Files Created/Modified

- `internal/compiler/cgen/cgen.go` - `EmitPayloadConformance`, `payloadRecordHasCType`, `payloadSlotSwapForTest`/`SetPayloadSlotSwapForTest`, `wrongPayloadSlot`, and the mutation branch inside `OpConstructPayload`'s case in `emitBranchOperations`
- `internal/compiler/session/session.go` - `PayloadProbeDataType`, `PayloadLayoutMutationRunner`
- `testdata/phase12/payload_layout_mismatch.golden.c` - frozen, deliberately transposed fixture (D-12-37); no correct version ever committed
- `internal/compiler/session/session_payload_control_test.go` - `TestPayloadLayoutMutationRefused`, `TestPayloadLayoutMutationAttacksFrozenFixtureOnly`, `TestPayloadSlotSwapMutationKilled`, plus helpers `isolateNativeFunction`, `runFunctionOkExecutions`, `checkedProgram`
- `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` - `D-12-26` and `D-12-43` rows/sections added; `D-12-24` confirmed unchanged; `items:` count updated to 8
- `.planning/phases/12-result-payloads/12-VALIDATION.md` - Per-Task Verification Map completed for all 5 plans; Wave 0 checklist marked done; `wave_0_complete: true`

## Decisions Made

See `key-decisions` in the frontmatter for the full list. Summarized:

1. Task 1 uses a purpose-built, single-byte-field `PayloadProbeDataType` rather than the production tracer's Buffer/Fault shape, to avoid `LANG_BUFFER`'s real padded size becoming an accidental confound in the layout-transposition control.
2. Task 2's mutation seam is a production-visible `cgen.go` function (not `export_test.go`), because its driving test lives in a different package (`session_test`) — following 12-04's own cross-package precedent.
3. Task 2's companion beat isolates `payload_borrow_interaction.lang`'s `identity` function from its sibling `choose` function via a new test helper, since native emission cannot handle the fixture's two-function shape directly.
4. Task 2's central finding — the wrong-slot mutation is invisible to every comparator axis — was reached by actually building and running the mutation, not by inference alone, and is recorded as an escalated criterion defect per the plan's own pre-written contingency (D-12-41), never silently declared satisfied.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] `export_test.go`'s `func Set.*ForTest` grep count could not reach 3 as the plan's parenthetical implied**
- **Found during:** Task 2
- **Issue:** The plan's read_first named `cgen.SetOpCallGroupedArmForTest` (`export_test.go`) as "the exact... shape a mutation seam must follow", and the acceptance criterion greps `export_test.go` for `func Set.*ForTest` expecting "at least 2 (the existing seams plus the new one)". But `export_test.go` symbols are visible only within `cgen`'s own test binary — `session_test` (where `TestPayloadSlotSwapMutationKilled` must live, per the plan's own Task 2 action text) cannot see them, exactly as 12-04-SUMMARY.md's own Deviation #2 already discovered for `interp`'s equivalent seam.
- **Fix:** Followed 12-04's own precedent: `SetPayloadSlotSwapForTest` is a production-visible function in `cgen.go` itself (documented as a test-only no-op, never called from production code), not an `export_test.go` symbol. The literal acceptance criterion (`>=2` matches in `export_test.go`) is still satisfied — it was already true before this plan touched anything, since the two pre-existing seams (`SetOpCallGroupedArmForTest`, `SetCallBoundaryAttributeSetForTest`) are untouched.
- **Files modified:** `internal/compiler/cgen/cgen.go` (not `export_test.go`)
- **Verification:** `TestPayloadSlotSwapMutationKilled` passes from `session_test`; `grep -c 'func Set.*ForTest' internal/compiler/cgen/export_test.go` still prints 2.
- **Committed in:** `b475d41` (Task 2 commit)

**2. [Rule 1 - Bug, verification-command note] `grep -c 'axis:' session_phase5_compare.go` prints 11, not the plan's expected 5**
- **Found during:** Task 2's verification pass
- **Issue:** The plan's `<verify>` command expects exactly 5 occurrences of the substring `axis:` in `session_phase5_compare.go`, intending to prove no sixth axis was added. The file already contained 11 occurrences of that substring (in comments, not just the 5 `Axis*` constant declarations) BEFORE this plan touched it — confirmed via `git show HEAD:...| grep -c 'axis:'` against the original Phase 5 commit (`4ca3cdc`), predating this entire phase.
- **Fix:** Not a code fix — this is a pre-existing inaccuracy in the plan's own verification command, not a regression this plan introduced. The authoritative checks both pass: `git diff --quiet -- internal/compiler/session/session_phase5_compare.go` (file byte-unchanged) and `grep -n '^\tAxis' session_phase5_compare.go` (exactly 5 axis constant declarations, unchanged).
- **Files modified:** none
- **Verification:** `git diff --quiet` exits 0; exactly 5 `Axis*` constants confirmed by direct inspection.

---

**Total deviations:** 2 (1 Rule 3 auto-fix, 1 pre-existing verification-command inaccuracy noted rather than "fixed"). **Impact:** Both are necessary corrections for the plan's own stated intent to be reachable/verifiable at all; neither is scope creep, and neither touched code beyond what Task 2 already required.

## Known Stubs / Simplifications

- **D-12-38's decisive value-divergence control could not be made load-bearing this phase.** `TestPayloadSlotSwapMutationKilled`'s mutated beat proves, empirically, that the current architecture has no channel through which a wrong-slot payload write can reach any of the five comparator axes. This is recorded as `D-12-43` in `PHASE-12-DEBT.md`, reopening only if a future plan changes how a payload-carrying return's terminal value is derived (e.g. permitting a match arm to return bound payload content directly, or adding a payload-value-aware observation channel to the differential harness). No phase currently owns this reopening work.

These are flagged for whoever next extends the payload grammar or the Phase-5 comparator, not silently dropped.

## Issues Encountered

None beyond the two items recorded above under Deviations, both resolved within this plan's own task scope.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- RES-03 is now the last requirement this phase carries to completion; RES-02 was already marked Complete (plan 04/03).
- `PHASE-12-DEBT.md` carries eight items total, all attributed to a landing phase or closed; `D-12-43` is the one item this plan newly opened, and it is unowned — a future plan touching the payload grammar's match-arm-return shape should read it first.
- `12-VALIDATION.md`'s Per-Task Verification Map and Wave 0 checklist are complete for the whole phase; `status`/`nyquist_compliant` remain `draft`/`false` for `/gsd-validate-phase` to set.
- Phase 12 is complete after this plan (5 of 5 plans landed). Ready for `/gsd-verify-work 12` and `/gsd-validate-phase 12`.

## Self-Check: PASSED

- FOUND: internal/compiler/cgen/cgen.go
- FOUND: internal/compiler/session/session.go
- FOUND: testdata/phase12/payload_layout_mismatch.golden.c
- FOUND: internal/compiler/session/session_payload_control_test.go
- FOUND: .planning/phases/12-result-payloads/PHASE-12-DEBT.md
- FOUND: .planning/phases/12-result-payloads/12-VALIDATION.md
- FOUND commit: 66dd2ee (Task 1)
- FOUND commit: b475d41 (Task 2)
- FOUND commit: 5559929 (Task 3)
- `go test ./...` exits 0
- `go vet ./...` clean
- `gofmt -l` clean on every file this plan touched
- `go test ./internal/compiler/session/... -run 'TestPayloadLayoutMutationRefused' -count=1` exits 0
- `go test ./internal/compiler/session/... -run 'TestPayloadSlotSwapMutationKilled' -count=1 -race -shuffle=on` exits 0
- `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -count=1` exits 0
- `go test ./internal/compiler/core/... -run TestPreviousPhaseGoldenCUnchanged -count=1` exits 0 (four pinned digests byte-identical)
- `git diff --quiet -- internal/compiler/session/session_phase5_compare.go` exits 0 (byte-unchanged)
- `clang --version` resolves (Apple clang 21.0.0)

---
*Phase: 12-result-payloads*
*Completed: 2026-09-13*
