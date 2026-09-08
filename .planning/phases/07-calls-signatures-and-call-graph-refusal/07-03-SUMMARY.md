---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 03
subsystem: compiler-core
tags: [go, core, check, corevalidate, interp, cgen, syntax, call-graph]

requires:
  - phase: 07-02
    provides: "Callable as publication safety (D-04-03), independent corevalidate summary peer wired at both replay sites"
provides:
  - "core.LinearOperation.CalleeID (additive, omitempty) -- populated only on OpCall, holds the callee's function ID, never its name (D-07-29)"
  - "core.OpCall registered in AllOperationKinds(), recognized -- never faked -- at all six dispatch sites"
  - "core.CallCalleeUnresolved -- D-07-45's typed identity for an OpCall whose CalleeID names no declared function, distinct from the (07-06) cycle code"
  - "check.resolveCallBinding -- the call admission predicate (arity-1, argument-in-scope, callee resolution) shared by analyzeStraightLine and analyzeArmBody"
  - "check.verifyCallInvariants -- check's own independent post-build re-derivation of the three CalleeID invariants"
  - "the relocated foreign-bare-call refusal, now emitted by check with its pre-Phase-07 published code preserved verbatim (D-07-40)"
  - "testdata/phase07/call_basic.lang, testdata/phase07/call_from_both_match_arms.lang"
affects: [07-04, 07-05, 07-06, 07-07, 07-08]

actuals:
  tokens: 19300
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "One-token-lookahead formatter dispatch (syntax/format.go) -- a bare call's callee identifier must not end its own line the way a plain binding source does, mirroring the existing tryCall flag with a new callBinding flag"
    - "Shared unexported fault-injection seam consulted from two call sites in the same package (check.verifyCallInvariantsSeam gates both resolveCallBinding and verifyCallInvariants), vs. the cross-package split (corevalidate.disableCalleeResolutionCheckForTest via export_test.go) when the two independent derivation sites live in different packages"

key-files:
  created:
    - testdata/phase07/call_basic.lang
    - testdata/phase07/call_from_both_match_arms.lang
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/corevalidate/export_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/native/native_test.go

key-decisions:
  - "D-07-29/D-07-30 ratified as decided at the checkpoint (auto-approved, gate=blocking, default option): additive omitempty CalleeID field, function-ID-keyed (not name-keyed) node identity."
  - "The parser change (accepting a bare call's shape unconditionally) and the refusal's relocation to check (D-07-40) are inseparable within one green commit: removing the parser's unconditional refusal without an equivalent check-side refusal would have let testdata/phase4/fallible_call_unconsumed.lang silently pass, breaking a pinned test table entry (check_test.go:791) immediately. Task 1's commit therefore necessarily carries the full relocation and the two shipped-test edits (D-07-40), not just the tracer's happy path; Task 3 adds the remaining checker predicate tests and the match-arm enumeration proof on top."
  - "check.verifyCallInvariantsSeam and corevalidate.disableCalleeResolutionCheckForTest are two SEPARATE seam variables (not one shared value) gating the resolves-to-a-declared-function predicate, one per package, following 07-02's precedent that a single Go test cannot flip two packages' unexported vars in one function."

requirements-completed: [SEM-04, QLT-08]

coverage:
  - id: D1
    description: "core.LinearOperation.CalleeID added (additive+omitempty, D-07-29), core.OpCall registered in AllOperationKinds() (never TerminatorKinds()), core.CallCalleeUnresolved declared as a distinct typed identity (D-07-45)"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsRegistered"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestTerminatorKindsIsSubsetOfAll"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "A two-function Lang program (one calling the other) is admitted end to end and recognized -- never faked -- at all six dispatch sites (check, corevalidate, interp, cgen, pathoracle, originvalidate)"
    requirement: "SEM-04"
    verification:
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/call_basic.lang (exit 0, no error diagnostic)"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite"
        status: pass
    human_judgment: false
  - id: D3
    description: "interp and cgen each get their own explicit, dedicated OpCall arm returning a named recognized-but-unsupported error, never folded into a grouped copy/move/borrow case"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite (interp.ErrCallUnsupported exception path)"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen.go (grouped-case exclusion, build-verified: core.OpCall never appears alongside OpCopy/OpMove/OpBorrowShared/OpBorrowExclusive)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The three CalleeID invariants (empty on OpCall, non-empty on non-OpCall, unresolved callee) are derived independently at BOTH check and corevalidate, each with its own seeded-mutation seam"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestOpCallEmptyCalleeIDIsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestNonOpCallWithCalleeIDIsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestOpCallUnresolvedCalleeIsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestDisableCalleeResolutionCheckSeamSuppressesUnresolvedRefusal"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallInvariantsRefusesEmptyCalleeID"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallInvariantsRefusesNonCallWithCalleeID"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallInvariantsRefusesUnresolvedCallee"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallInvariantsSeamRestoresBothRefusals"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestUndeclaredCalleeRefusedNeverSilentlyDropped"
        status: pass
    human_judgment: false
  - id: D5
    description: "The relocated foreign bare-call refusal reports its original published code (syntax.fallible_call_not_consumed), now from check; a bare call to a declared Lang function is admitted; two shipped tests were edited deliberately (D-07-40); calls from both match arms are enumerated by Kind == core.OpCall across all operations, not by position"
    requirement: "SEM-04"
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_test.go#TestFallibleCallUnconsumedRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestFallibleCallUnconsumedRefusedAtCheckNotFormat"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestBareCallToDeclaredFunctionAdmitted"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestTwoArgumentCallRefusedAtCheckNotParse"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallArgumentNotInScopeRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestEveryBindingIsFallibleRejectsMixedCallBinding"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallFromBothMatchArmsEnumeratedByKind"
        status: pass
    human_judgment: false

duration: ~55min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 03: Calls, CalleeID, and Six-Site Recognition Summary

**`core.OpCall` carries a real `CalleeID` (the callee's function ID, never its name), is registered in `AllOperationKinds()`, and is recognized -- never faked -- at all six dispatch sites, with the three CalleeID invariants derived independently at both `check` and `corevalidate`.**

## Performance

- **Duration:** ~55 min
- **Tasks:** 3 (plus one checkpoint:decision, auto-approved under active auto-mode)
- **Files modified:** 16 (2 created, 14 modified)

## Accomplishments

- `core.LinearOperation.CalleeID` (additive, `omitempty`) added per the exact `OkEdgeID`/`Allocator`/`Reason` precedent: populated only on an `OpCall`, holds the callee's function ID, never its name. `core.OpCall` registered in `AllOperationKinds()` (never `TerminatorKinds()`); `TestPreviousPhaseCoreBytesUnchanged` and every pre-existing golden `.c` file stay byte-identical.
- `core.CallCalleeUnresolved` declared as D-07-45's own typed identity for an unresolvable callee -- distinct from the (07-06) cycle code, ready for `07-06`'s callgraph to reuse as an inert string constant.
- The parser's `linearBody` now accepts a bare call's shape (`ast.RHS{Kind:"call"}`) unconditionally instead of refusing it at parse time; admission relocates to `check` (D-07-01). Reused the existing `p.callArguments()` -- no second argument parser.
- `check.resolveCallBinding`, shared verbatim by `analyzeStraightLine` and `analyzeArmBody`: arity-1 (D-07-07 defers arity-N), argument-in-scope, and callee resolution to a declared Lang function (admitted, `CalleeID` set), a declared foreign symbol (refused with the pre-Phase-07 `syntax.fallible_call_not_consumed` code -- D-07-40 relocates the enforcement layer, not the code), or neither (refused with `core.CallCalleeUnresolved`, never silently dropped).
- `check.verifyCallInvariants`: a post-build pass over the whole assembled program, derived from `check`'s own function-ID table, independently re-deriving the same three CalleeID invariants `corevalidate` derives on its own side -- neither consults the other.
- `corevalidate`: `OpCall` replayed at both `replayStraightLine` and `replayBlocks`, modelled on `OpCopy`'s single-target shape; the CalleeID kind-exclusivity check runs once per operation (unconditionally, regardless of kind), and the resolves-to-a-declared-function check consults the program's own function ID set.
- `interp` and `cgen` each carry a dedicated, explicit `OpCall` arm in every one of their switches (three for `interp`, four grouped-case sites for `cgen`), returning a named "recognized, unsupported" error/`ErrCallUnsupported` rather than folding into a grouped copy/move/borrow case that would fake execution.
- Two independent D-07-41/D-07-42 fault-injection seams for the resolves-to-a-declared-function predicate: `check.verifyCallInvariantsSeam` (shared by `resolveCallBinding` and `verifyCallInvariants`, both in package `check`) and `corevalidate.disableCalleeResolutionCheckForTest` (exposed via `export_test.go`), following 07-02's cross-package split precedent.
- `testdata/phase07/call_basic.lang` (the tracer fixture) and `testdata/phase07/call_from_both_match_arms.lang` (D-07-28's Phase-12 blindness guard: both match arms call the same callee, and the checked program's two `core.OpCall` operations are found by scanning by `Kind`, never by block position -- still found with operation order reversed).
- Fixed a **pre-existing formatter bug** the new bare-call shape exposed: a call's callee identifier ended its own line prematurely (the rule a plain binding source follows), stranding the argument list on the next line garbled. Added a one-token lookahead to `syntax.Format` and a `callBinding` flag mirroring the existing `tryCall` shape.

## Task Commits

1. **Checkpoint: ratify CalleeID as an additive core schema field** -- auto-approved (`⚡ Auto-selected: Proceed as decided`); `AUTO_CFG=true`, gate defaulted to `blocking` (not `blocking-human`).
2. **Task 1: End-to-end two-function call -- CalleeID, OpCall, and recognition at all six sites** -- `7af419f` (feat)
3. **Task 2: The three CalleeID invariants and the unresolved-callee typed identity, at both derivation sites** -- `1cbe6d7` (test)
4. **Task 3: Relocate the foreign bare-call refusal, edit the two shipped tests deliberately, and prove match-arm enumeration** -- `5002913` (test)

## Files Created/Modified

- `internal/compiler/core/core.go` - `CalleeID` field, `OpCall` kind, `core.CallCalleeUnresolved`
- `internal/compiler/core/core_test.go` - registry count, six-site fixture list, interp exception path
- `internal/compiler/ast/ast.go` - `RHS.Kind == "call"` doc comment
- `internal/compiler/syntax/parser.go` - bare-call shape accepted unconditionally
- `internal/compiler/syntax/format.go` - one-token lookahead + `callBinding` flag (bug fix)
- `internal/compiler/syntax/syntax_test.go` - `TestFallibleCallUnconsumedRejected` relocated assertion (D-07-40)
- `internal/compiler/check/check.go` - `resolveCallBinding`, `verifyCallInvariants`, `functionIDs` table, threaded through `checkLinear`/`checkBranch`/`analyzeStraightLine`/`analyzeArmBody`
- `internal/compiler/check/check_test.go` - Task 2/3 tests; `analyzeStraightLine` call sites updated for the new signature
- `internal/compiler/corevalidate/corevalidate.go` - `OpCall` cases at both replay sites, CalleeID invariants, `disableCalleeResolutionCheckForTest` seam
- `internal/compiler/corevalidate/corevalidate_test.go` - Task 2 CalleeID invariant tests; `LinearWorkLimit`/pinned-count updates
- `internal/compiler/corevalidate/export_test.go` - `SetDisableCalleeResolutionCheckForTest`
- `internal/compiler/interp/interp.go` - `ErrCallUnsupported`, dedicated `OpCall` arms
- `internal/compiler/cgen/cgen.go` - dedicated `OpCall` arms at all four grouped-case sites
- `internal/compiler/native/native_test.go` - `fallible_call_unconsumed.lang`'s `format --check` expectation relocated (D-07-40); new native-level pin test
- `testdata/phase07/call_basic.lang` - the tracer fixture
- `testdata/phase07/call_from_both_match_arms.lang` - D-07-28's match-arm enumeration fixture

## Decisions Made

See `key-decisions` in frontmatter. The most consequential: **the parser change and the refusal relocation are inseparable within one green commit.** Task 1's action text scoped the relocation to Task 3, but removing the parser's unconditional bare-call refusal (required for the tracer's happy path) without an equivalent check-side refusal would have let `testdata/phase4/fallible_call_unconsumed.lang` silently pass admission -- breaking `check_test.go:791`'s pinned diagnostic-code table entry immediately, which the plan's own Task 1 `<verify>` (`go test ./internal/compiler/check/...`, unscoped by `-run`) would have caught. Task 1's commit therefore necessarily carries the full relocation logic and the two shipped-test edits (D-07-40) that Task 3's action text describes; Task 3 adds the remaining checker predicate tests (arity-2, argument-scope, admitted-Lang-callee) and the match-arm enumeration proof on top, plus one native-level test needed to satisfy the plan's own `-run` pattern against a real test name.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Formatter bug: a bare call's callee identifier ended its own line prematurely**
- **Found during:** Task 1, verifying `testdata/phase07/call_basic.lang` round-trips through `lang format`
- **Issue:** `syntax.Format`'s existing rule (a plain binding source identifier ends its own line right after `=`) fired for a bare call's callee identifier too, since both follow `TokenEqual` identically at the token level. The callee's own argument list then got stranded on the next line, garbled (`let result = identity\n(value) result`). This is a real, previously-latent bug: the token sequence already existed pre-Phase-07 for `testdata/phase4/fallible_call_unconsumed.lang` (a bare call, syntactically valid tokens even though semantically refused), but `format --check` never reached the real formatter for it because `syntax.Parse` used to return a diagnostic first, short-circuiting formatting entirely.
- **Fix:** Added a one-token lookahead to `Format()` and a `callBinding` flag (mirroring the existing `tryCall` flag) so a bare call's closing paren -- not its callee identifier -- ends the line.
- **Files modified:** `internal/compiler/syntax/format.go`
- **Verification:** Both new fixtures round-trip byte-identically through `lang format`; full `go test ./...` green.
- **Committed in:** `7af419f` (Task 1)

**2. [Rule 3 - Blocking] `native_test.go`'s `format --check` expectation for `fallible_call_unconsumed.lang` needed updating, not just the diagnostic-layer relocation**
- **Found during:** Task 1, running `go test ./internal/compiler/native/...`
- **Issue:** Once the parser no longer short-circuits on this fixture, `format --check` runs the real formatter, which correctly reports the fixture's pre-existing (never previously visible) non-canonical leading-comment-to-`module` spacing as `format.non_canonical` -- a genuine, pre-Phase-07 formatting fact this plan surfaces but does not alter.
- **Fix:** Updated the `refused(...)` table entry's `formatDiagnostic` argument from `syntax.fallible_call_not_consumed` to `format.non_canonical`, with a comment naming D-07-40.
- **Files modified:** `internal/compiler/native/native_test.go`
- **Verification:** `go test ./internal/compiler/native/...` green.
- **Committed in:** `7af419f` (Task 1)

**3. [Rule 3 - Blocking] `TestAllOperationKindsRegistered`'s declared count and `TestAllOperationKindsHandledAtEverySite`'s fixture list/interp exception**
- **Found during:** Task 1, running `go test ./internal/compiler/core/...`
- **Issue:** Registering a tenth `OperationKind` broke the hardcoded `declaredCount = 9` pin, and the six-site exhaustive-dispatch control requires every declared kind to be exercised by at least one corpus fixture -- `core.OpCall` had none. Additionally, that control's `interp.Run` step asserts no error for any fixture function, which `core.OpCall`'s intentionally-unsupported `ErrCallUnsupported` would otherwise fail.
- **Fix:** Bumped the declared count to 10; added `testdata/phase07/call_basic.lang` to the fixture list; added a documented exception so a function containing an `OpCall` returning `interp.ErrCallUnsupported` counts as "handled" (recognized, not faked) rather than a control failure.
- **Files modified:** `internal/compiler/core/core_test.go`
- **Verification:** `go test ./internal/compiler/core/...` green.
- **Committed in:** `7af419f` (Task 1)

**4. [Rule 1 - Bug, in scope of this plan's own change] `LinearWorkLimit` and a pinned corevalidate work-count constant needed updating**
- **Found during:** Task 1, running `go test ./internal/compiler/corevalidate/...`
- **Issue:** The new unconditional per-operation CalleeID kind-exclusivity check adds exactly one counted `v.check` per operation, moving `TestCoreValidationWorkSeries`'s exact-formula assertion and `TestAcyclicChainsStillValidateUnderCycleGuard`'s pinned constant, following the project's own established convention (see that test's pre-existing comment documenting three prior such moves from Phase 4).
- **Fix:** `LinearWorkLimit` moved from `16*facts+14` to `17*facts+14`; the pinned constant moved from 412 to 425 (measured from the built code, both with documented deltas).
- **Files modified:** `internal/compiler/corevalidate/corevalidate.go`, `internal/compiler/corevalidate/corevalidate_test.go`
- **Verification:** `go test ./internal/compiler/corevalidate/...` green.
- **Committed in:** `7af419f` (Task 1)

**5. [Rule 3 - Blocking] `testdata/phase07/call_from_both_match_arms.lang`'s callee could not be a bare-arm match function**
- **Found during:** Task 3, running the full `go test ./...` suite (corevalidate's corpus-wide summary-peer tests failed against the new fixture)
- **Issue:** A module cannot mix a match-bodied function with a straight-line one (`core.mixed_body_versions`), so the callee had to be match-bodied too -- but a bare-arm match function ALSO cannot share a program with an arm-bodied one: `corevalidate`'s per-function schema dispatch requires `lang.core/0` for a bare match, while the whole program's schema is bumped to `lang.core/1` the moment ANY function carries an arm body. My first draft used a bare-arm `helper`, which broke both `TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus` and `TestSummaryPeerCallableAgreesOnOriginOmittedClass` (07-02's corpus-wide walks, which pick up every `testdata/**/*.lang` fixture).
- **Fix:** Rewrote `helper` as an arm-bodied match function (trivial single-result arm bodies), matching `main`'s shape.
- **Files modified:** `testdata/phase07/call_from_both_match_arms.lang`
- **Verification:** `go test ./...` green, including the corevalidate summary-peer corpus tests.
- **Committed in:** `5002913` (Task 3)

**6. [Rule 3 - Blocking] The plan's own native `<verify>` command matched no real test name**
- **Found during:** Task 3, running the plan's exact verify command (`go test ... -run 'Unconsumed|Call|Fallible|MatchArms' -v`)
- **Issue:** `TestShippedBinaryFourSubcommandCorpusMatrix` -- the only relevant native test -- doesn't match that pattern at the top level (Go's `-run` with no `/` matches only the top-level test name; its subtests aren't reached), so the command reported "no tests to run" for the native package, which the plan's own `<fails_when>` clause treats as a failure.
- **Fix:** Added `TestFallibleCallUnconsumedRefusedAtCheckNotFormat`, a small native-level pin exercising this exact fixture's four-subcommand matrix entry via the same `phase4CorpusMatrix()`/`runShippedBinarySteps` machinery, whose name matches the plan's `-run` pattern.
- **Files modified:** `internal/compiler/native/native_test.go`
- **Verification:** The plan's exact verify command now runs and passes for all three packages.
- **Committed in:** `5002913` (Task 3)

---

**Total deviations:** 6 auto-fixed (1 bug in this plan's own new formatter path, 5 blocking issues required to keep `go test ./...` green at every task commit). **Impact:** All were necessary consequences of the plan's own scope (a new `OperationKind` registered, a new grammar shape accepted, a new corpus fixture added) rather than scope creep; none touch code or tests outside Phase 07's own surface.

## Known Stubs

None. `resolveCallBinding`, `verifyCallInvariants`, both derivation sites' invariant checks, and both fault-injection seams are fully wired and exercised. `interp`'s and `cgen`'s `OpCall` arms are intentionally-unsupported this phase (SEM-08's call-stack semantics are Phase 10; native multi-function emission is Phase 11) -- named as such per D-07-39, not silently stubbed.

## Threat Flags

None beyond what `07-03-PLAN.md`'s own `<threat_model>` already registered (T-07-14 through T-07-20, T-07-SC) -- all mitigated as designed; no new security-relevant surface was introduced outside that register.

## Issues Encountered

None beyond the deviations documented above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `core.OpCall`, `CalleeID`, and `core.CallCalleeUnresolved` are real, tested, and independently re-derived at both `check` and `corevalidate` -- `07-04`'s exhaustive-dispatch controls (which prove recognition, not execution) have real six-site data to assert against.
- `07-05`'s call-admission predicate (arity/signature matching against `Callable`) has `resolveCallBinding` to extend rather than replace.
- `07-06`'s callgraph has `core.CallCalleeUnresolved` to reuse as an inert string constant, `CalleeID` as its sole edge-bearing fact, and `testdata/phase07/call_from_both_match_arms.lang` as its Phase-12 blindness-guard precedent (scan by `Kind`, never by position).
- `07-08`'s `ClosureDigest` chaining arm (over real callees) has a real `CalleeID` to chain from.
- Ready for `07-04`.

## Self-Check: PASSED

- All key-files (created + modified) verified present on disk with `[ -f ]`.
- All three commits (`7af419f`, `1cbe6d7`, `5002913`) verified present via `git log --oneline --all`.
- Re-ran every task's `<acceptance_criteria>` and `<verify>` commands: all pass, including the plan-level `go test ./...`, `go vet ./...`, `go test -race ./internal/compiler/check/... ./internal/compiler/corevalidate/...`, and `TestPreviousPhaseCoreBytesUnchanged`.
- Both Stage 1 fixtures check clean via `go run ./cmd/lang --json check` and round-trip byte-identically through `lang format`.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*
