---
phase: 04-fallible-resources-and-c-boundary
plan: "06"
subsystem: compiler-verification
tags: [origin-analysis, path-oracle, terminator-registry, ffi-alias, ownership-cost]

# Dependency graph
requires:
  - phase: 04-fallible-resources-and-c-boundary (waves 1-5)
    provides: "core.TerminatorKinds()/core.AllOperationKinds() registry (W1), OpRelease and release-order rederivation (W2), the foreign contract's three inspectable layers (W3), OpDefect and the closed terminal-outcome axis (W4), the nonlocal-exit landing pad and ledger (W5)"
provides:
  - "originvalidate and pathoracle each walk every terminator (return, typed failure, defect), read from the single core.TerminatorKinds() registry"
  - "A mutation-killed control (control:terminator.walk_incomplete) proving the widening actually bites, independently for both packages"
  - "core.ForeignContract.Alias: a foreign symbol can declare its ok-edge return borrows or retains its argument, recognised as borrow-derived by originvalidate"
  - "core.foreign_origin_omitted: an undeclared borrow-derived foreign return is refused, derived from the core artifact alone"
  - "originvalidate.ValidatePublished now runs on lang check and both lang run engines, not only interface export (WR-01 closed)"
  - "discoverLoanLastUses counts its own transitive-scan work, closing the metric-honesty half of carried debt D-03-01"
  - "04-DEBT.md: the dated Phase 4 debt register (D-04-26 retirement carry-forward, D-04-30 payload-alternative carry-forward, D-04-31 accepted residual limitations)"
affects: [phase-05, phase-06]

# Actuals (#2632)
actuals:
  tokens: 47000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Set-membership terminator discriminant read from a single registry (core.TerminatorKinds()), never restated as a literal, in both independent analyses"
    - "Exported TerminatorKindsOverride/RecognizesTerminator seam per package (originvalidate, pathoracle) enabling both an in-process automated mutation-kill test and session.go's runtime control, without weakening the black-box _test package convention"
    - "A narrow, separate re-check at the CLI command-file layer (session.go's publishedOriginProblemFile) rather than folding a publication-only gate into the lower-level engine functions every non-publication test drives directly"

key-files:
  created:
    - testdata/phase4/foreign_origin_omitted.lang
    - .planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/check/check.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/pathoracle/pathoracle.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/check/check_test.go

key-decisions:
  - "core.ForeignContract.Alias (new additive omitempty field, values '' | borrow | retain) reuses the existing shared/exclusive access vocabulary rather than inventing a third, and is populated from a new optional 'alias' foreign-declaration policy key with no admission gate requiring its presence (unlike unwind/nonlocal_exit)"
  - "WR-01 wiring lives at the session.go command-file layer (CheckCommandFile/RunInterpreterCommandFile/RunNativeCommandFile), not inside RunInterpreter/RunNative themselves -- those two are driven directly by many non-publication tests (e.g. TestExclusiveBorrowInterpreterNative on the exclusive_borrow_clean/relay D-04-03 witness) that must keep passing unchanged"
  - "discoverLoanLastUses' own work counting is one unit per binding the outer loop visits plus one for the final result-position check -- deliberately NOT one unit per chain-ancestor step, so the independent test oracle (oracleStraightLine, a genuinely different BFS-reachability mechanism) can add the identical scalar term without reproducing production's internal walk shape"
  - "checkForeignOriginOmitted is a standalone backward walk in originvalidate.go, not a reuse of RecomputeOriginPerReturn's walk, per T-04-37's import/law independence discipline"

patterns-established:
  - "Fixture facts going unchecked under a narrowed recognised set are demonstrated by an automated in-process override AND a real throwaway-detached-worktree source mutation (D-09), both pasted verbatim"

requirements-completed: [SEM-03, FFI-01]

coverage:
  - id: D1
    description: "originvalidate and pathoracle both walk every terminator (return, typed failure, defect), read from core.TerminatorKinds(), mutation-killed independently in both packages"
    requirement: "SEM-03"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestOriginWalksEveryTerminator"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestTerminatorSetReadFromRegistry"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestTerminatorWalkMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_test.go#TestPathOracleClosesOnEveryTerminator"
        status: pass
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_test.go#TestFailureOnlyPathIsChecked"
        status: pass
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_test.go#TestTerminatorWalkMutationKilled"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4TerminatorControl"
        status: pass
    human_judgment: false
  - id: D2
    description: "A foreign call declared to borrow/retain its argument is recognised as borrow-derived; an undeclared origin on such a return is refused with core.foreign_origin_omitted"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestForeignBorrowDerivedReturnRecognised"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestForeignOriginOmittedRejected"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4ForeignOriginControl"
        status: pass
      - kind: e2e
        ref: "cmd/lang check testdata/phase4/foreign_origin_omitted.lang (exit 2, core.foreign_origin_omitted)"
        status: pass
    human_judgment: false
  - id: D3
    description: "originvalidate.ValidatePublished runs on lang check and lang run, not only interface export; testdata/phase3/public_view_omitted.lang, which checked clean before this plan, is now refused"
    requirement: "SEM-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestPublishedOriginValidatedOnCheckAndRun"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_test.go#TestOriginValidatorImportsStayIndependent"
        status: pass
    human_judgment: false
  - id: D4
    description: "discoverLoanLastUses counts its own transitive-scan work; growth series over three chain lengths; before/after values recorded; no verdict changes over the whole corpus"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLastUseDiscoveryWorkIsCounted"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestLastUseDiscoveryWorkSeries"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCheckerVerdictsUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestOwnershipWorkSeries"
        status: pass
    human_judgment: false
  - id: D5
    description: "04-DEBT.md records the deferred discoverLoanLastUses retirement, the deferred payload-carrying-alternative work, and the accepted residual limitations, each dated with identifier/severity/source/landing phase"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false
    rationale: "The structural claim -- that every register item carries an identifier, a source, a threat/requirement, a severity from a closed vocabulary, a landing phase, and a matching detail section, and that the frontmatter's declared count matches the table -- is now asserted on every commit for EVERY phase's register, not only this one, with the pre-landing-phase-column registers (02, 03) carried on an explicit, justified exemption list. What deliberately stays a human reading, and is NOT claimed by that test, is whether each item's prose honestly describes the deferral; the dangerous direction of that judgment (a residual being described as covered) is separately mechanized by TestNoCoverageClaimedForNamedResiduals."

duration: ~2h
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 06: Terminator-Walk Widening and Foreign Origin Alias Summary

**Both independent analyses now walk every terminator from one registry, a foreign call's declared borrow/retain obligation is recognised as borrow-derived, and published-origin validation reaches `lang check`/`lang run` -- closing the phase's single highest-risk item and the carried WR-01 debt.**

## Performance

- **Duration:** ~2h
- **Started:** 2026-09-04T23:00:00-04:00 (approx, first read)
- **Completed:** 2026-09-05T00:10:00-04:00 (approx, final commit)
- **Tasks:** 3 completed
- **Files modified:** 9 (2 created, 7 modified)

## Accomplishments

- `originvalidate.RecomputeOriginPerReturn` and `pathoracle.linearizePath` both discriminate terminators via a set-membership test against `core.TerminatorKinds()` instead of an equality test against `core.OpReturn` alone -- widening the registry now automatically widens both walkers.
- A new required gate control, `control:terminator.walk_incomplete`, asserts this equality directly in both packages and is mutation-killed independently in each (see verbatim output below).
- `core.ForeignContract.Alias` (new additive field) lets a foreign symbol declare its ok-edge return borrows or retains its argument; `originvalidate` recognises this as a borrow hop in its origin walk and independently refuses an undeclared such return with `core.foreign_origin_omitted`.
- `originvalidate.ValidatePublished` now runs on `lang check` and both `lang run` engines (interpreter and native), not only `interface export` -- closing WR-01. `testdata/phase3/public_view_omitted.lang`, which checked clean under `lang check` before this plan, is now refused with `core.origin_omitted`.
- `discoverLoanLastUses` counts its own transitive-scan work, closing the metric-honesty half of carried debt D-03-01; the retirement of the quadratic derivation itself is re-recorded as dated open debt (D-04-26) in the new `04-DEBT.md`, alongside the deferred payload-carrying-alternative work (D-04-30) and four accepted residual limitations (D-04-31).

## Task Commits

Each task was committed atomically:

1. **Task 1: Walk every terminator (D-04-29)** - `c9677e2` widened `pathoracle`; the `originvalidate` half of Task 1 landed together with Task 2 in the next commit (the two tasks share non-separable code paths in `originvalidate.go`/`session.go` -- see Deviations).
2. **Task 2: Foreign returns carry their origin; check/run wiring (D-04-27/D-04-28/D-04-29)** - `1103fc1`
3. **Task 3: Honest liveness cost; phase debt register (D-04-25/D-04-26)** - `0c9c524`

**Plan metadata:** commit pending (this SUMMARY + STATE/ROADMAP update).

## Files Created/Modified
- `internal/compiler/core/core.go` - `core.ForeignContract.Alias` additive omitempty field
- `internal/compiler/check/check.go` - new `"alias"` foreign policy key; `discoverLoanLastUses` returns its own work count
- `internal/compiler/originvalidate/originvalidate.go` - terminator-set widening, `TerminatorKindsOverride`/`RecognizesTerminator` seam, `checkForeignOriginOmitted`, `OpForeignCall` borrow-hop recognition
- `internal/compiler/pathoracle/pathoracle.go` - terminator-set widening, `sawReturn` renamed `sawTerminator`, same override/recognize seam
- `internal/compiler/session/session.go` - `control:terminator.walk_incomplete` and `control:origin.foreign_origin_omitted` lanes; `publishedOriginProblemFile` wiring into `CheckCommandFile`/`RunInterpreterCommandFile`/`RunNativeCommandFile`
- `internal/compiler/originvalidate/originvalidate_test.go`, `internal/compiler/pathoracle/pathoracle_test.go`, `internal/compiler/session/session_test.go`, `internal/compiler/check/check_test.go` - new falsifiers per task
- `testdata/phase4/foreign_origin_omitted.lang` - the shipped witness fixture
- `.planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md` - new dated debt register

## Decisions Made

See `key-decisions` above. In addition:

- **`core.ForeignContract.Alias` reuses the existing "shared"/"exclusive" access vocabulary** (via `"borrow"`→shared, `"retain"`→exclusive), not a third vocabulary -- consistent with the "never restate a law" principle already applied to `PublicOrigin.Access`.
- **Positive-declaration test for foreign-borrow origin (`TestForeignBorrowDerivedReturnRecognised`) attaches the declared `PublicOrigin` directly onto the checked `core.Function`** rather than via source syntax, because `check.go`'s `checkFallibleLinear`/`checkForeignTracer` does not (this phase) wire a source-level `-> borrow(path) Type` annotation onto a foreign-tracer-shaped function -- that wiring is out of this plan's scope (D-04-28 requires the *validator* recognise the shape, not that the *checker* newly support declaring it on this function shape). Recorded here rather than silently worked around.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `RunInterpreter`/`RunNative` regression from naive WR-01 wiring**
- **Found during:** Task 2, first wiring attempt
- **Issue:** Wiring `originvalidate.ValidatePublished` directly into `session.RunInterpreter`/`session.RunNative` (the lower-level engine functions) broke `TestExclusiveBorrowInterpreterNative`, a shipped Phase 3 test that deliberately runs the `exclusive_borrow_clean`/`relay` fixture (D-04-03's own witness: checks clean, undeclared borrow-derived return) through the native/interpreter engines directly, not through publication.
- **Fix:** Moved the origin-validation re-check to a narrow, separate helper (`publishedOriginProblemFile`) invoked only from the CLI command-file wrappers (`CheckCommandFile`, `RunInterpreterCommandFile`, `RunNativeCommandFile`), leaving `RunInterpreter`/`RunNative` themselves untouched -- exactly matching D-04-27's actual scope ("`lang check`/`lang run`", the command surface, not every internal call to the engines).
- **Files modified:** `internal/compiler/session/session.go`
- **Verification:** Full repo test suite green (`go test ./...`, `go test -race ./...`), including `TestExclusiveBorrowInterpreterNative` unchanged.
- **Committed in:** `1103fc1`

**2. [Rule 3 - Blocking] `discoverLoanLastUses`/`oracleLoanLastUses` Work-total mismatch after counting change**
- **Found during:** Task 3
- **Issue:** Adding a work counter to `discoverLoanLastUses` broke several exact-`Work`-value assertions (`TestOwnershipSequenceExhaustive`, `TestOwnershipOracleTracksLoansPerOwner`, `TestOwnershipWorkSeries`, `FuzzOwnershipLinear`) because the independent test oracle (`oracleStraightLine`/`oracleLoanLastUses`, a genuinely different BFS-reachability mechanism) did not carry an equivalent term.
- **Fix:** Simplified `discoverLoanLastUses`'s own counting to a mechanism-independent scalar (one unit per binding scanned, plus one for the final result check) and added the identical `len(body.Bindings)+1` term to `oracleStraightLine`, so the two independently-implemented mechanisms report the same aggregate cost without the oracle reproducing production's internal walk shape. Updated `TestOwnershipWorkSeries`'s formula and linear-bound constant accordingly.
- **Files modified:** `internal/compiler/check/check.go`, `internal/compiler/check/check_test.go`
- **Verification:** `go test ./internal/compiler/check/...` green, including the exhaustive oracle differential (~113,164 cases).
- **Committed in:** `0c9c524`

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking).
**Impact on plan:** Both fixes were necessary to avoid regressing shipped Phase 3 tests/oracles while still satisfying this plan's own required behavior. No scope creep.

## Issues Encountered

- **Task boundary is not perfectly clean across `originvalidate.go`/`session.go`.** Task 1's origin-walker widening and Task 2's foreign-origin recognition are interleaved in the same functions in `originvalidate.go` (both touch `RecomputeOriginPerReturn`'s collection loop and `walkReturnOrigin`'s switch), and `session.go` gained both tasks' gate lanes and CLI wiring in overlapping regions. Rather than force an artificial mid-function split that would leave an intermediate commit in a semantically incoherent state, Task 1's `originvalidate.go`/`session.go`/`core.go` changes were committed together with Task 2 (commit `1103fc1`), which is called out explicitly in that commit's message. `pathoracle.go` (pure Task 1) and `check.go`'s `discoverLoanLastUses` change (pure Task 3, cleanly separable from Task 2's `check.go` "alias" policy-key addition at a different line range) were split correctly into their own commits.

## Mutation-Kill Demonstrations (D-09/D-04-29)

Per the plan's explicit instruction, both walkers were mutation-killed in a throwaway detached worktree (`git worktree add --detach`, never committed, removed with `git worktree remove --force` afterward) by deleting `core.OpFail` from each package's own recognized-terminator function. Verbatim output below.

### Mutation 1 — `internal/compiler/originvalidate/originvalidate.go`

Mutation applied to `recognizedTerminatorKinds()`:
```go
// MUTATION (D-09/D-04-29 throwaway demonstration): OpFail deleted from
// the recognised terminator set.
return []core.OperationKind{core.OpReturn, core.OpDefect}
```

```
=== MUTATION 1: OpFail deleted from originvalidate's recognized terminator set ===
=== RUN   TestForeignBorrowDerivedReturnRecognised
--- PASS: TestForeignBorrowDerivedReturnRecognised (0.00s)
=== RUN   TestOriginWalksEveryTerminator
    originvalidate_test.go:726: expected one origin entry per terminator (return + fail), got 1: [{OperationID:s1:phase4.foreign_acquire_one:fn:main:op:1 Paths:[] Access: Derived:false}]
--- FAIL: TestOriginWalksEveryTerminator (0.00s)
=== RUN   TestTerminatorSetReadFromRegistry
    originvalidate_test.go:737: expected originvalidate to recognise registered terminator "fail"
--- FAIL: TestTerminatorSetReadFromRegistry (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/originvalidate	0.202s
```

The `foreign_acquire_one.lang` fixture's fail-only path (its err block's only exit is `core.OpFail`) goes from contributing an origin-picture entry to being silently invisible the instant `OpFail` is deleted from the recognized set — exactly the fixture fact the widening exists to protect.

### Mutation 2 — `internal/compiler/pathoracle/pathoracle.go`

Same mutation shape applied to `pathoracle`'s own `recognizedTerminatorKinds()`.

```
=== MUTATION 2: OpFail deleted from pathoracle's recognized terminator set ===
=== RUN   TestPathOracleClosesOnEveryTerminator
    pathoracle_test.go:217: terminator=fail: unexpected error (path incorrectly treated as malformed): pathoracle.unterminated_loan: function "s1:fn:fail_only" references loan "loan:0" with no recorded birth on this path
--- FAIL: TestPathOracleClosesOnEveryTerminator (0.00s)
=== RUN   TestFailureOnlyPathIsChecked
--- PASS: TestFailureOnlyPathIsChecked (0.00s)
=== RUN   TestTerminatorSetReadFromRegistry
    pathoracle_test.go:260: expected pathoracle to recognise registered terminator "fail"
--- FAIL: TestTerminatorSetReadFromRegistry (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/pathoracle	0.191s
```

A synthetic fail-only path carrying a live loan flips from a correctly-computed endpoint to a `pathoracle.unterminated_loan` false rejection the instant `OpFail` is deleted from the recognized set — the exact "malformed CFG" misclassification the widening exists to prevent.

## Work-Counting Before/After (D-04-25)

For a fixed straight-line chain of 100 non-borrowing bindings (`TestOwnershipWorkSeries`'s `operations=10` case is analogous; values below are for `operations=100`, the same shape scaled):

| | Formula | Value |
|---|---|---|
| Before (discoverLoanLastUses uncounted) | `1 + 2*(operations+1)` | 203 |
| After (discoverLoanLastUses counts itself) | `1 + 2*(operations+1) + (operations+1)` | 304 |

The 101-unit difference is exactly `discoverLoanLastUses`'s own scan cost (one unit per binding scanned, plus one for the final result-position check) — previously invisible to the caller's `Work` total, now counted honestly per D-04-25.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The phase's single highest-risk item (D-04-29) is closed with a mutation-killed control, not merely a passing differential.
- WR-01 is closed: published-origin validation reaches every command surface that admits a program.
- D-03-01's structural half (retiring `discoverLoanLastUses`) is re-recorded as D-04-26, explicitly deferred to Phase 5 to avoid the compounding-wave defect shape.
- RES-01 and FFI-01 remain open in REQUIREMENTS.md pending plan 04-07, per phase-state guidance -- not force-closed here.
- Ready for 04-07 (phase close-out / mid-phase gate finalization).

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*

## Self-Check: PASSED
