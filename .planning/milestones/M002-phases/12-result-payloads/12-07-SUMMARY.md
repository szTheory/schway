---
phase: 12-result-payloads
plan: 07
subsystem: compiler-core
tags: [go, compiler, core-ir, cgen, interp, payload-types, cr-01, in-01, wr-02]

requires:
  - phase: 12-result-payloads
    provides: "check.duplicate_payload_type (plan 12-06's source-layer refusal), and session.PayloadProbeDataType (D-12-37) as proof a hand-built core.Program carrying a colliding shape exists in-tree"
provides:
  - "core.AlternativeNameForPayloadType: the single, ambiguity-detecting derivation of payload-type-to-alternative-name, replacing two independent engine-local reimplementations"
  - "core.LookupAlternativeDetail: the single alternative-detail lookup, replacing check's and cgen's separate reimplementations (IN-01 closed)"
  - "A committed tripwire (TestPayloadAlternativeResolutionHasExactlyOneDerivation) proving neither engine can silently re-grow a local copy"
  - "cgen.PayloadSlotSwapInjectedWriteCount: WR-02's anti-vacuity counter proving the D-12-38 fault-injection seam actually injected a wrong-slot write"
affects: ["12-08 (D-12-44 decision entry, requirements closure for RES-02/RES-03)"]

actuals:
  tokens: 6451
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Single shared derivation in core, consumed by check/cgen/interp, reported via Go error rather than a silent first-match (D-12-25 convergence discipline applied to a new fact)"
    - "Anti-vacuity counter for a fault-injection seam: increment only inside the seam's successful-target branch, reset on engage, read via a cross-package production-visible function (not export_test.go) following SetPayloadSlotSwapForTest's own precedent"

key-files:
  created:
    - internal/compiler/core/core_single_derivation_test.go
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/interp/interp.go
    - internal/compiler/check/check.go
    - internal/compiler/session/session_payload_control_test.go

key-decisions:
  - "PayloadSlotSwapInjectedWriteCount's reader was implemented as a production-visible function in cgen.go (mirroring SetPayloadSlotSwapForTest's own cross-package precedent), NOT as an export_test.go accessor as the plan's action text literally specified -- see Deviations."

requirements-completed: [RES-02, RES-03]

coverage:
  - id: D1
    description: "core.AlternativeNameForPayloadType is the single, ambiguity-detecting derivation of payload-type-to-alternative-name; it reports ambiguity (naming both colliding alternatives) rather than a first-match guess, and both cgen and interp read it exclusively."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_single_derivation_test.go#TestAlternativeNameForPayloadTypeAmbiguityRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_single_derivation_test.go#TestLookupAlternativeDetailMatchesCheckSemantics"
        status: pass
    human_judgment: false
  - id: D2
    description: "A committed tripwire fails if either engine re-grows a local alternativeNameForPayloadType declaration or stops reading the shared core helper; demonstrated red-then-green once during execution."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_single_derivation_test.go#TestPayloadAlternativeResolutionHasExactlyOneDerivation"
        status: pass
      - kind: manual_procedural
        ref: "Red-then-green transcript recorded below (stub reverted, never committed)"
        status: pass
    human_judgment: true
    rationale: "The red-then-green demonstration is an execution-time procedural proof (temporarily reintroducing a stub declaration), not a re-runnable committed test by construction -- verified manually and recorded verbatim below."
  - id: D3
    description: "IN-01's third independent derivation (cgen's emitBranch per-alternative detail loop) is closed: it now reads core.LookupAlternativeDetail, the same helper check delegates to."
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/cgen/... -count=1 (existing emitBranch-dependent tests, e.g. TestPayloadTracerThreeEngineAgreement)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The D-12-38 fault-injection seam can no longer pass vacuously: TestPayloadSlotSwapMutationKilled's mutated subtest asserts a wrong-slot write was actually injected (>=1), naming WR-02 on failure. D-12-43's recorded finding is preserved verbatim."
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_payload_control_test.go#TestPayloadSlotSwapMutationKilled/mutated"
        status: pass
    human_judgment: false
  - id: D5
    description: "core.LinearOperation gains no field; corevalidate.go is byte-unchanged; the six legacy cgen emitters are untouched beyond the two payload call sites; go test ./... is green across all 25 packages."
    requirement: "RES-03"
    verification:
      - kind: integration
        ref: "go test ./... (25 package result lines, zero FAIL)"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_convention_absence_test.go#TestConventionOverrideNotExpressibleInCore (linearOperationExpectedFields tripwire, unaffected)"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-13
status: complete
---

# Phase 12 Plan 07: CR-01 Engine-Layer Closure + WR-02 Summary

**`core.AlternativeNameForPayloadType`/`core.LookupAlternativeDetail` collapse three independently-reimplemented payload-alternative derivations into one ambiguity-reporting shared helper in `core`, backed by a reverted-fix tripwire, and `cgen.PayloadSlotSwapInjectedWriteCount` closes WR-02's vacuous-pass hole in the D-12-38 mutation-kill control.**

## Performance

- **Duration:** ~45 min
- **Tasks:** 3
- **Files modified:** 5 (core.go, cgen.go, interp.go, check.go, session_payload_control_test.go) + 1 created (core_single_derivation_test.go)

## Accomplishments

- Added `core.AlternativeNameForPayloadType(DataType, string) (string, error)` to `internal/compiler/core/core.go`: scans `AlternativeDetails`, returns the unique match, an explicit "no alternative declares this payload type" error for zero matches, an explicit ambiguity error naming BOTH colliding alternative names for two-or-more matches, and an error for an empty payload-type query (never a nullary match).
- Added `core.LookupAlternativeDetail(DataType, string) AlternativeDetail`, moved up from `check`'s own unexported copy, preserving byte-identical known-name/unknown-name semantics.
- Deleted `cgen.alternativeNameForPayloadType` and `interp.alternativeNameForPayloadType` (both file-local, both silent first-match). Rewired both `emitBranchOperations` payload call sites in `cgen` and `runFrameStack`'s `OpConstructPayload` case in `interp` to call the shared `core` helper, turning a non-nil error into an attributable `fmt.Errorf`/`return Execution{}, fmt.Errorf` naming the operation ID -- `interp`'s existing `(Execution, error)` channel needed no new mechanism, confirming the plan's flagged assumption.
- Closed IN-01: `emitBranch`'s per-alternative zero-value `detail` loop now calls `core.LookupAlternativeDetail` directly; `check.lookupAlternativeDetail` is now a one-line delegation to the same helper. All three consumers (`check`, `cgen`, `interp`) read one derivation.
- Created `internal/compiler/core/core_single_derivation_test.go` with `TestAlternativeNameForPayloadTypeAmbiguityRefused` (four subtests covering unique match, ambiguity, no-match, and empty-query), `TestLookupAlternativeDetailMatchesCheckSemantics` (known/unknown name), and `TestPayloadAlternativeResolutionHasExactlyOneDerivation` (Task 2's reverted-fix control, both negative "no local declaration" and positive "uses the shared helper" assertions per engine file).
- Manually demonstrated the tripwire's red-then-green transition (see below); reverted cleanly, `git diff --stat` on `cgen.go` from that demonstration was empty.
- Added `cgen.payloadSlotSwapInjectedWriteCount` (unexported package-level counter, incremented only inside `wrongPayloadSlot`'s successful-target branch, reset by `SetPayloadSlotSwapForTest` on engage) and `cgen.PayloadSlotSwapInjectedWriteCount()` (production-visible reader -- see Deviations for why this is not an `export_test.go` symbol). `TestPayloadSlotSwapMutationKilled`'s mutated subtest now asserts the injected count is >= 1 before comparing engines, naming WR-02 on failure. Observed on a real run: **2** wrong-slot writes injected. D-12-43's finding and the `companion_unmutated_agreement` subtest are unchanged.
- `go test ./...` green across all 25 packages; `git diff --stat` since the pre-plan commit names exactly the five files above plus the new test file -- no `testdata/**` golden, no `.planning/` file.

## Task Commits

1. **Task 1: One shared, ambiguity-detecting derivation in `core`; both engine copies deleted** - `e8b82b5` (feat)
2. **Task 2: A tripwire that fails if either engine's local derivation returns** - `bad615e` (test)
3. **Task 3: WR-02 — the fault-injection seam can no longer no-op silently** - `e07613b` (fix)

## Files Created/Modified

- `internal/compiler/core/core.go` - Added `AlternativeNameForPayloadType` and `LookupAlternativeDetail`
- `internal/compiler/core/core_single_derivation_test.go` - New: three test functions covering the shared derivation's behavior and the cross-engine tripwire
- `internal/compiler/cgen/cgen.go` - Deleted local `alternativeNameForPayloadType`; rewired both `emitBranchOperations` payload call sites and `emitBranch`'s detail loop to the shared `core` helpers; added WR-02's injection counter and its reader
- `internal/compiler/interp/interp.go` - Deleted local `alternativeNameForPayloadType`; rewired `OpConstructPayload` in `runFrameStack` to resolve the data type by name then call the shared `core` helper
- `internal/compiler/check/check.go` - `lookupAlternativeDetail` is now a one-line delegation to `core.LookupAlternativeDetail`
- `internal/compiler/session/session_payload_control_test.go` - Added WR-02's injected-write assertion inside `TestPayloadSlotSwapMutationKilled`'s mutated subtest; one sentence appended to the doc comment

## Ambiguity Error Text Shipped

`core.AlternativeNameForPayloadType` on the colliding `PayloadProbe`-shaped fixture (`First`/`Second`, both `Byte`):

> `data type "Colliding": payload type "Byte" is ambiguous between alternatives [First Second]`

No-match case:

> `data type "Outcome": no alternative declares payload type "LANG_NONEXISTENT"`

Empty-query case:

> `data type "Outcome": cannot resolve an alternative for the empty payload type`

## Which Error Channel `interp` Used

`interp`'s `OpConstructPayload` case sits inside `runFrameStack`, whose signature is already `(Execution, error)` and which already returns `Execution{}, fmt.Errorf(...)` for structurally impossible IR (confirming the plan's flagged assumption). No new error channel was introduced: a non-nil `core.AlternativeNameForPayloadType` error is wrapped as `return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, altErr)`, naming the operation ID exactly as the plan's read_first specified.

## Observed Red-Then-Green Demonstration (Task 2)

Procedure: appended a stub `func alternativeNameForPayloadType(dataType core.DataType, payloadType string) string { return "" }` to the end of `internal/compiler/cgen/cgen.go`, ran the tripwire, then restored the original file from a pre-edit backup and re-ran.

**RED** (stub present):
```
=== RUN   TestPayloadAlternativeResolutionHasExactlyOneDerivation/cgen:_no_local_declaration
    core_single_derivation_test.go:194: internal/compiler/cgen/cgen.go: found a local declaration of "alternativeNameForPayloadType" -- the shared core.AlternativeNameForPayloadType derivation must not be reimplemented locally (CR-01/D-12-25)
--- FAIL: TestPayloadAlternativeResolutionHasExactlyOneDerivation (0.00s)
    --- FAIL: TestPayloadAlternativeResolutionHasExactlyOneDerivation/cgen:_no_local_declaration (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/cgen:_uses_the_shared_core_helper (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/interp:_no_local_declaration (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/interp:_uses_the_shared_core_helper (0.00s)
FAIL
```

**GREEN** (stub removed, original restored):
```
--- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/cgen:_no_local_declaration (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/cgen:_uses_the_shared_core_helper (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/interp:_no_local_declaration (0.00s)
    --- PASS: TestPayloadAlternativeResolutionHasExactlyOneDerivation/interp:_uses_the_shared_core_helper (0.00s)
PASS
```

`git diff --stat internal/compiler/cgen/cgen.go` after restoring showed zero change — the demonstration was never committed.

Both directions of the tripwire were also confirmed by inspection of the test body per the plan's acceptance criteria: `engineSourceHasNoLocalResolverDeclaration` (negative) and `engineSourceUsesSharedHelper` (positive) are both called for each engine file inside `TestPayloadAlternativeResolutionHasExactlyOneDerivation`, so deleting payload resolution entirely (rather than sharing it) also turns the test red.

## Observed Injected-Write Count (Task 3)

`go test ./internal/compiler/session/... -run 'TestPayloadSlotSwapMutationKilled' -v -count=1`:

```
session_payload_control_test.go:250: WR-02: fault-injection seam injected 2 wrong-slot write(s)
```

Two writes were injected on a single run of the mutated beat (the tracer fixture's `identity` function is driven through the interpreter and native `-O0`/`-O3` engines via `runFunctionOkExecutions`; the counter reflects the total across whichever of those paths actually invoke `cgen`'s `emitBranchOperations` with the seam engaged). The seam is confirmed live, not vacuously no-opping, in the shipped tree.

## Decisions Made

- **PayloadSlotSwapInjectedWriteCount is production-visible, not an `export_test.go` symbol** (see Deviations below) — this mirrors `SetPayloadSlotSwapForTest`'s own established precedent and documented rationale for the identical cross-package constraint.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] WR-02's counter accessor could not live in `export_test.go` as specified**

- **Found during:** Task 3
- **Issue:** The plan's action text instructed adding the counter's test-only reader to `internal/compiler/cgen/export_test.go`. `TestPayloadSlotSwapMutationKilled` lives in `internal/compiler/session`'s `session_test` package, an EXTERNAL consumer that imports `cgen` normally (not `cgen_test`). Go's `_test.go` export trick — the mechanism `export_test.go`'s existing accessors rely on — compiles those symbols into `cgen`'s OWN test binary only; they are invisible to any other package's build, including `session_test`'s. `SetPayloadSlotSwapForTest`'s own doc comment in `cgen.go` already documents this exact constraint as the reason IT is production-visible rather than an `export_test.go` symbol. Following the plan literally would have produced a symbol `TestPayloadSlotSwapMutationKilled` could never reference, making the assertion impossible to compile.
- **Fix:** Added `cgen.PayloadSlotSwapInjectedWriteCount()` as a production-visible function in `cgen.go` (immediately after `SetPayloadSlotSwapForTest`), following that function's exact precedent and citing the same cross-package rationale in its own doc comment. No symbol was added to `export_test.go`.
- **Files modified:** `internal/compiler/cgen/cgen.go`
- **Verification:** `go build ./...` succeeds; `TestPayloadSlotSwapMutationKilled` compiles and asserts `cgen.PayloadSlotSwapInjectedWriteCount() >= 1` from the external `session_test` package.
- **Committed in:** `e07613b` (Task 3 commit)
- **Note on the literal acceptance criterion:** the plan's acceptance criteria for Task 3 states "`internal/compiler/cgen/export_test.go` exposes a test-only accessor for that counter." This criterion is NOT met literally — no symbol was added to that file — because meeting it literally would have shipped a non-functional accessor unreachable from the test that needs it. The counter itself, its production-visible reader, and the assertion are all present, tested, and green; only the FILE the reader lives in differs from the plan's literal text, for the reason `SetPayloadSlotSwapForTest`'s own precedent already establishes in this same file.

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The deviation is a file-placement correction required for the code to compile and the cross-package assertion to work at all; it does not change WR-02's behavior, scope, or the counter's semantics. No scope creep.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- CR-01 is fully closed: plan 12-06's source-layer refusal is the primary closure; this plan's shared, ambiguity-detecting `core` derivation is the CORE-layer backstop for a hand-built `core.Program`, and IN-01's third derivation is gone.
- WR-02 is closed: the D-12-38 fault-injection seam can no longer pass vacuously.
- `core.LinearOperation` is unchanged; `corevalidate.go` is byte-unchanged; the six legacy `cgen` emitters are untouched beyond the two payload call sites.
- `go test ./...` is green across all 25 packages.
- RES-02 and RES-03 are declared by plans 12-06, 12-07, AND 12-08 (shared IDs). Per `requirements.ready-ids`, they remain unmarked until plan 12-08 also produces a summary — this plan does NOT call `requirements.mark-complete` directly, consistent with the prior-wave guidance from 12-06.
- Plan 12-08 can proceed: D-12-44's decision entry and the shared requirements' final closure are its responsibility.

---
*Phase: 12-result-payloads*
*Completed: 2026-09-13*
