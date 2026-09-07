---
phase: 04-fallible-resources-and-c-boundary
plan: "04"
subsystem: compiler
tags: [defect, panic, terminal-outcome, streaming-events, signal-adjudication, c17, mutation-testing]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "plan 01/02/03's OpForeignCall/OpFail/OpRelease trio, core.AllOperationKinds()/TerminatorKinds() six-site exhaustive-dispatch control, checkFallibleLinear/checkResourceLifecycle, the byte-frozen native/lang_foreign_resource.c + private header, the complete core.ForeignContract, verifyForeignCorpus's existing required controls"
provides:
  - "core.OpDefect: a real, reachable, abort-only terminal operation admissible in a match arm's terminal position (`defect \"<reason>\"`), carrying the required Reason field, joining AllOperationKinds()/TerminatorKinds()"
  - "The closed terminal-outcome axis (execution.OutcomeReturned/TypedFailure/Defect/Cancelled), with cancelled reserved and structurally unconstructible"
  - "corevalidate's independent OpFail-reached-only-from-an-err-edge control (core.fail_reached_without_err_edge/core.fail_edge_missing)"
  - "cgen's emitDefectSupport (_Noreturn lang_defect) and emitStreamingEventSupport/emitLinearForeignOutputSupport: the additive streaming event path for every foreign-acquiring function"
  - "native.ExpectDefect, signal-aware exit adjudication via ProcessState.Sys().(syscall.WaitStatus), and the native.terminal_record_absent hard-failure code distinct from cap truncation"
  - "session.DefectHasNoReleaseAfter and two new verifyForeignCorpus required controls: control:defect.no_release_on_defect, control:defect.signal_adjudicated"
affects: [04-05, 04-06, 04-07]

actuals:
  tokens: 20900
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "A defect terminator's core.LinearOperation carries no TargetID and reads its own arm's alias place purely to keep the 'every operation reads an initialized SourceID' invariant uniform across every OperationKind -- it never actually consumes the value."
    - "The streaming event path is a wholly NEW sibling emitter (emitStreamingEventSupport/emitLinearForeignOutputSupport), never a rewrite of the shared, frozen emitEventSupport -- selected only by emitLinearForeign, so every non-foreign-acquiring emitter's generated C stays byte-identical."
    - "A defect event needing a field the shared LANG_EVENT struct doesn't carry (the reason string, 'output') is assembled directly with lang_write_json_string calls rather than routed through lang_record_event/lang_write_events, so the frozen shared struct never has to grow a field only one caller needs."
    - "A nonzero child exit is adjudicated into exactly three distinct shapes (clean, nonzero-no-signal, signalled) via a small exitAdjudication value type, decided ONLY through ProcessState.Sys().(syscall.WaitStatus) -- never a hardcoded exit code."

key-files:
  created:
    - testdata/phase4/defect_terminal.lang
  modified:
    - internal/compiler/syntax/token.go
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/ast/ast.go
    - internal/compiler/core/core.go
    - internal/compiler/core/core_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/execution/execution.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "OpDefect is scoped to a match arm's terminal position this plan (checkBranch/analyzeArmBody), not to checkResourceLifecycle's straight-line resource-lifecycle shape -- the plan's own text requires only 'a witness reachable independently of every foreign decision', and defect-inside-a-resource-lifecycle-function would require checkResourceLifecycle/emitLinearForeign/corevalidate surgery this plan's files_modified does not include. Recorded as carried scope, not a narrowing that weakens any success criterion: the shipped witness needs no call surface at all, which is the stronger claim."
  - "The defect terminal record's own event (kind function.defected, carrying the required reason string) is assembled directly in cgen with individual lang_write_json_string calls rather than through lang_record_event/lang_write_events, because the shared LANG_EVENT struct emitEventSupport declares (frozen for every other emitter, D-04-23) has no field for it -- adding one there would move every existing committed generated-C golden."
  - "D-04-20's streaming event path is selected ONLY by emitLinearForeign (every function it handles already carries a foreign contract, checked at that function's own entry) -- not wired into emitBranch's defect emission, since an arm cannot carry a foreign acquisition this phase and its existing buffered write-then-abort sequence already keeps the terminal record the last write on that path."
  - "Task 3's 'aborting program' native-layer falsifiers are synthetic Go subprocess helpers (TestNativeHelperProcess's abort-mid-stream/abort-clean modes, self-signalling via syscall.Kill with GOTRACEBACK=crash so the OS reports a genuine WaitStatus.Signaled()), not the full compiled-C pipeline -- this narrowly and honestly tests native.go's own capture/adjudication layer (which is where D-04-24 and D-04-20's 'events survive an abort' guarantee actually live) without requiring a defect-producing foreign-acquiring fixture, which plan 04's own scope (see above) does not build."
  - "Deviation (Rule 1 -- bug): native.decodeExecution silently ignored its own `expect` parameter, always calling validateExecution(value, ExpectValue) regardless of what Runner.Run's caller declared. Fixed to thread `expect` through -- required for ExpectDefect (and the pre-existing ExpectTypedFailure) to mean anything end-to-end, not just when validateExecution is called directly in a test."

patterns-established:
  - "A three-way exit adjudication (clean / nonzero-no-signal / signalled) lives in a single small helper (adjudicateExit) that every Runner.Run() call site consults, rather than inlining WaitStatus type-assertions at each call site."
  - "A per-document mutation-kill (a hand-constructed execution.Execution literal) is the pattern for a control whose true production trigger (an interpreter bug) cannot be safely injected inside an automated test suite -- paired with a one-off, human-witnessed revert-and-fail demonstration against the real interpreter (recorded below) for the stronger claim."

requirements-completed: [SEM-03]

coverage:
  - id: D1
    description: "A real, reachable, deliberately terminal defect operation exists (core.OpDefect), admissible only in a match arm's terminal position, aborting via a generated _Noreturn lang_defect function whose every path ends in abort(), with no route into typed_failure"
    requirement: SEM-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_test.go#TestDefectTerminatorRoundTrips"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestDefectLowersToTerminalOutcome"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestCancelledOutcomeIsUnconstructible"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestNoErrorValueConstructorExists"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "An additive streaming event emitter writes each event at the point it occurs for every foreign-acquiring function, leaving every other emitter (and every Phase 1/2/3 golden) byte-identical; native.go reports a terminal-record-absence code distinct from cap truncation"
    requirement: SEM-03
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestStreamingEmitterWritesAtPointOfOccurrence"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestExistingEmittersAreByteIdentical"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestAbortingProgramStreamsEventsBeforeDying"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestTerminalRecordAbsenceIsHardFailure"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestTruncationAndAbsenceReportDistinctCodes"
        status: pass
    human_judgment: false
  - id: D3
    description: "A dying program is adjudicated by its real wait status (never a hardcoded exit code), and the rule that a defect path runs no cleanup is an enforced, mutation-killed control alongside a real interpreter/native signal-adjudication differential"
    requirement: SEM-03
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestAbortSignalAdjudicatedByWaitStatus"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestNonzeroExitIsDistinctFromSignal"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestDefectExpectationRejectsReturnedDocument"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestNoReleaseAfterDefect"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4DefectControls"
        status: pass
    human_judgment: false

duration: ~2h (single continuous session)
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 4: Fallible Resources and C Boundary — Terminal Defect and Signal-Aware Adjudication Summary

**A real `defect "<reason>"` terminator aborts the process root via a generated `_Noreturn lang_defect`, the terminal-outcome axis is closed with `cancelled` reserved and provably unconstructible, foreign-acquiring functions stream their events as they occur instead of buffering them, and a dying process is adjudicated by its true `WaitStatus` signal — never a hardcoded exit code — with the no-release-on-defect rule enforced as a mutation-killed control.**

## Performance

- **Duration:** ~2h (single continuous session)
- **Tasks:** 3/3 completed
- **Files modified:** 19 (1 created, 18 modified)

## Accomplishments

- `core.OpDefect` (`"defect"`) joins `core.AllOperationKinds()`/`core.TerminatorKinds()`; the six-site exhaustive-dispatch control (`TestAllOperationKindsHandledAtEverySite`) turned red before every site handled it and green after. It carries a required, non-empty `Reason` string and no `TargetID`.
- `defect "<reason>"` is a new terminator form admissible only in a match arm's terminal position: `TokenDefect` in the lexer/parser, `ast.LinearBody.DefectReason`, format-idempotent, checked in `analyzeArmBody`, lowered by `checkBranch` into an `OpDefect`-terminated block. `testdata/phase4/defect_terminal.lang` is the shipped witness — needs no call surface at all.
- The terminal-outcome axis is a closed named set declared once in `execution.go` (`OutcomeReturned`/`OutcomeTypedFailure`/`OutcomeDefect`/`OutcomeCancelled`), with `cancelled` reserved and asserted unconstructible by scanning `interp.go`/`cgen.go`/`native.go`'s own source for the literal — never by observing that no test happened to produce one.
- `corevalidate` independently refuses an `OpFail` reached from any predecessor other than a real `err` edge (`core.fail_edge_missing`/`core.fail_reached_without_err_edge`), proven by hand-corrupting a checked program's edge pattern and confirming validation now rejects it (`TestNoErrorValueConstructorExists`).
- `interp.runBranchArm` and `cgen.emitBranch` (via the new `emitDefectSupport`-generated `_Noreturn lang_defect` function, every path of which calls `abort()`) both execute the defect path: outcome kind `"defect"`, no value, no release, a `function.defected` event carrying the reason. `lang_defect`'s `_Noreturn` marker is the one and only exemption from the zero-attribute control (`control:foreign.no_unproven_attributes`), since it is a property of a function `cgen` itself emits.
- `cgen.emitStreamingEventSupport`/`emitLinearForeignOutputSupport` is a wholly new, additive event-emission path selected only by `emitLinearForeign`: the JSON events array opens before any operation executes and each terminal writer closes it as the last write on that path, so an aborting foreign-acquiring process never loses an event that occurred before it died. The shared, frozen `emitEventSupport` (and `emitLinear`/`emitBranch`, which still use it) is byte-for-byte untouched — every committed Phase 1/2/3 generated-C golden stays identical (`TestExistingEmittersAreByteIdentical`).
- `native.go` adjudicates a dying child through `ProcessState.Sys().(syscall.WaitStatus)` alone: clean exit, nonzero-without-signal (`native.run_failed`), and signalled (`native.run_signaled`) are three distinct verdicts — except a `SIGABRT` while the caller declared the new `native.ExpectDefect`, which is the expected abort-only shape and falls through to decode the aborting process's own terminal record. A stdout stream that ends before a terminal record finishes decoding is now the distinct `native.terminal_record_absent` hard failure, never confused with cap truncation (`native.run_stdout_truncated`) or an ordinary malformed document.
- `session.DefectHasNoReleaseAfter` backs `control:defect.no_release_on_defect`: scans a defect execution's event stream for any `resource.released` event, mutation-killed both by a hand-constructed document carrying one AND by a live revert-and-fail demonstration against the real interpreter (below). `control:defect.signal_adjudicated` compiles and runs the shipped defect witness through the real clang/exec toolchain under both terminal shapes its two arms produce. Both join `verifyForeignCorpus`'s required-control list with nonzero recomputed work.

## Task Commits

Each task was committed atomically (with one documented exception — see Deviations):

1. **Task 04-04-01: A reachable terminal defect and a closed terminal-outcome axis** — `bb0db47` (feat)
2. **Task 04-04-02: The additive streaming event emitter, with an absent terminal record as a hard failure** — `1aaf54e` (feat) — this commit also carries Task 1's `cgen.go` OpDefect-emission hunks (see Deviations)
3. **Task 04-04-03: Signal-aware process adjudication and the no-release-on-defect control** — `8743bfe` (feat) — this commit also carries Task 2's `native.go`/`native_test.go` terminal-record-absence hunks (see Deviations)

**Plan metadata:** committed alongside this SUMMARY.

## Files Created/Modified

- `internal/compiler/syntax/{token,lexer,parser,format}.go` — `TokenDefect`, `defect "<reason>"` parsing/formatting
- `internal/compiler/ast/ast.go` — `LinearBody.DefectReason`
- `internal/compiler/core/core.go` — `core.OpDefect`, `LinearOperation.Reason`, widened `AllOperationKinds()`/`TerminatorKinds()`
- `internal/compiler/execution/execution.go` — the closed terminal-outcome axis constants + `TerminalOutcomeKinds()`
- `internal/compiler/check/check.go` — `analyzeArmBody`'s defect branch
- `internal/compiler/corevalidate/corevalidate.go` — `OpDefect` cases in both replay switches, the OpFail-only-from-err-edge control in `blocksAndEdges`
- `internal/compiler/interp/interp.go` — `OpDefect` cases in `runBranchArm`/`runLinearBlocks`
- `internal/compiler/cgen/cgen.go` — `emitDefectSupport`/`functionHasDefect`, the `emitBranchOperations` `OpDefect` case, `emitStreamingEventSupport`/`emitLinearForeignOutputSupport`, `emitLinearForeign`'s streaming rewrite
- `internal/compiler/native/native.go` — `ExpectDefect`, `adjudicateExit`/`exitAdjudication`, `native.terminal_record_absent`, the `function.defected` event validation, the `decodeExecution` expect-threading fix
- `internal/compiler/session/session.go` — `DefectHasNoReleaseAfter`, `lane:defect-no-release`/`lane:defect-signal-adjudicated`
- Test files: `syntax_test.go`, `check_test.go`, `core_test.go`, `cgen_test.go`, `native_test.go`, `session_test.go`
- `testdata/phase4/defect_terminal.lang` — the shipped defect witness

## Decisions Made

See `key-decisions` in the frontmatter. The most consequential: `OpDefect` is scoped to match-arm reachability only this plan (not wired into `checkResourceLifecycle`'s straight-line resource shape), and the native-layer "aborting program" falsifiers are synthetic Go subprocess helpers rather than a full compiled-C pipeline, since plan 04's own scope does not build a defect-producing foreign-acquiring fixture.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `native.decodeExecution` ignored its own `expect` parameter**
- **Found during:** Task 3, while wiring `native.ExpectDefect` end-to-end through `Runner.Run()`
- **Issue:** `decodeExecution(stdout []byte, expect TerminalOutcome)` accepted an `expect` parameter but its body called `validateExecution(value, ExpectValue)` unconditionally — every caller's declared expectation (including the pre-existing `ExpectTypedFailure`) was silently discarded below the `Runner.Run()` level.
- **Fix:** Changed the call to `validateExecution(value, expect)`.
- **Files modified:** `internal/compiler/native/native.go`
- **Verification:** `TestDefectExpectationRejectsReturnedDocument`, `TestAbortSignalAdjudicatedByWaitStatus`, and the full existing `TestValidateExecutionTypedFailureDiscriminates`/`TestValidateExecutionStillRejectsOldGrounds` suite all green.
- **Committed in:** `8743bfe`

**2. [Rule 1 - Bug] `TestExecutionDecoderRejectsMalformedOutput`'s "malformed"/"truncated" cases needed their expected code updated**
- **Found during:** Task 2, after widening `decodeExecution` to distinguish an incomplete document from a complete-but-malformed one
- **Issue:** These two pre-existing test cases (an input ending mid-decode) asserted `native.invalid_execution`; per D-04-20 that shape is now the more precise `native.terminal_record_absent`.
- **Fix:** Updated the two cases' expected code, with a comment pointing at the new dedicated tests.
- **Files modified:** `internal/compiler/native/native_test.go`
- **Verification:** `TestExecutionDecoderRejectsMalformedOutput` green with the updated expectations.
- **Committed in:** `8743bfe`

### Process deviation (not a code defect)

**3. Per-task commits are not fully atomic in the git history — two commits each carry one extra task's hunks in a shared file.** `cgen.go` was edited in one continuous pass carrying both Task 1's `OpDefect` emission and Task 2's streaming emitter; both landed in commit `1aaf54e` (Task 2's commit). `native.go`/`native_test.go` were similarly edited in one continuous pass carrying both Task 2's terminal-record-absence work and Task 3's signal adjudication; both landed in commit `8743bfe` (Task 3's commit). Each commit message documents exactly which extra hunks it carries and why. No production behavior is affected — this is purely a git-history granularity note.

---

**Total deviations:** 2 auto-fixed (both Rule 1 — bugs found and fixed during implementation, before the affected commit landed) + 1 documented process deviation (commit granularity, no behavior impact). **Impact on plan:** Both code fixes were necessary for the new `ExpectDefect`/`terminal_record_absent` machinery to mean anything end-to-end; no scope creep.

## Issues Encountered

None beyond the two Rule-1 fixes above.

## Revert-and-Fail Demonstration (D-09)

### control:defect.no_release_on_defect (live interpreter mutation, in the working tree — file backed up, edited, tested, restored)

Backed up `internal/compiler/interp/interp.go`, injected a spurious `resource.released` event into `runBranchArm`'s `OpDefect` case (right after the honest `function.defected` event), ran the control:

```
$ go test ./internal/compiler/session/... -run TestVerifyPhase4DefectControls -v
=== RUN   TestVerifyPhase4DefectControls
    session_test.go:1445: Phase 4 defect verify failed: status=invalid diagnostics=[verify.control_missing [0:0]: control:defect.no_release_on_defect] lanes=[... {Schema:lang.verify-lane/0 ID:lane:defect-no-release Status:fail Controls:[] RecomputedWork:124 ElapsedNS:331500 PeakRSSStatus:unavailable PeakRSSBytes:0 OutputBytes:678}]
--- FAIL: TestVerifyPhase4DefectControls (0.31s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/session	0.514s
```

Restored `interp.go` from the backup (`diff` confirmed byte-identical to the pre-mutation state), re-ran the same command:

```
$ go test ./internal/compiler/session/... -run TestVerifyPhase4DefectControls -v
=== RUN   TestVerifyPhase4DefectControls
--- PASS: TestVerifyPhase4DefectControls (0.74s)
PASS
ok  	github.com/codename-lang/lang/internal/compiler/session	0.979s
```

Full `go test -count=1 ./...` was re-run clean after the restore, confirming no residual state from the injection.

## Verification Performed

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test -count=1 ./...` — pass
- `go test -race ./...` — pass
- `go vet ./...` — clean
- `git diff <phase-start>..HEAD -- go.mod go.sum testdata/phase1 testdata/phase2 testdata/phase3 native/lang_foreign_resource.c native/lang_foreign_resource_private.h` — empty
- `sh scripts/verify-phase3.sh` — exits 0
- All named tests per each task's `<verify>` block, run via `scripts/assert-go-tests.sh ./internal/compiler/...` — pass:
  - Task 1: `TestDefectTerminatorRoundTrips TestDefectLowersToTerminalOutcome TestCancelledOutcomeIsUnconstructible TestNoErrorValueConstructorExists TestAllOperationKindsHandledAtEverySite TestPreviousPhaseCoreBytesUnchanged`
  - Task 2: `TestStreamingEmitterWritesAtPointOfOccurrence TestAbortingProgramStreamsEventsBeforeDying TestTerminalRecordAbsenceIsHardFailure TestTruncationAndAbsenceReportDistinctCodes TestExistingEmittersAreByteIdentical`
  - Task 3: `TestAbortSignalAdjudicatedByWaitStatus TestNonzeroExitIsDistinctFromSignal TestDefectExpectationRejectsReturnedDocument TestNoReleaseAfterDefect TestVerifyPhase4DefectControls`
- `testdata/phase4/defect_terminal.lang` AND a hand-written, out-of-corpus defect program both passed `format --check`, `check`, and `run --engine=interpreter` through a freshly built `./cmd/lang` (manual CLI session, not committed as a corpus fixture).

## Known Stubs

None that block this plan's own success criteria.

**Documented narrowing (not a stub):** `OpDefect` is reachable only via a match arm's terminal position this plan — a resource-lifecycle-shaped function (`checkResourceLifecycle`/`emitLinearForeign`) cannot yet end in `defect`. This is a real scope limitation, not a placeholder: the shipped witness (needing no call surface at all) is the *stronger* structural claim SC3 asks for, and the plan's own `files_modified` list does not include the resource-lifecycle checker/emitter surgery a defect-in-foreign-function shape would require. A later plan wiring defect into resource-lifecycle functions should also revisit `control:defect.no_release_on_defect`'s and `control:defect.signal_adjudicated`'s current match-arm-only fixture.

**Documented narrowing (not a stub):** Task 3's `TestAbortingProgramStreamsEventsBeforeDying` and `TestAbortSignalAdjudicatedByWaitStatus` exercise `native.go`'s own capture/adjudication layer through a synthetic Go subprocess helper (self-signalling via `syscall.Kill`), not a full compiled-C-from-a-defect-fixture pipeline — a direct consequence of the narrowing above. `control:defect.signal_adjudicated` (session-level, in `verifyForeignCorpus`) DOES compile and run the real `defect_terminal.lang` fixture through real clang/exec, so the end-to-end real-C claim is still covered, just by a different test.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The terminal defect operation, the closed terminal-outcome axis, the streaming event path, signal-aware adjudication, and the two new required controls are all in place for the phase's remaining plans (the `nm -u` non-unwind control and the process-root `setjmp` landing pad are explicitly carried to plan 04-05 per 04-03-SUMMARY's own "Next Phase Readiness" note).
- Carried-forward debt, unchanged from plans 01/02/03: `originvalidate`/`pathoracle` still only walk `OpReturn` (D-04-29's widening to `{OpReturn, OpFail, OpDefect}` remains out of scope for plans 01-04, carried to plan 06 per the phase's own source-coverage audit — `TestAllOperationKindsHandledAtEverySite`'s own doc comment already documents this narrower "handled" bar for those two sites).
- New, plan-04-local narrowing (documented above): `OpDefect` is match-arm-only; the process-root `setjmp` pad's own no-release enforcement (D-04-18's other half) still needs a real foreign-acquisition-plus-nonlocal-exit fixture once plan 04-05 builds the pad.
- No blockers.

## Self-Check: PASSED

All key files verified present on disk; all three task commits (`bb0db47`, `1aaf54e`, `8743bfe`) verified present in `git log`; full test suite, race detector, vet, and `verify-phase3.sh` all green as of this summary.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*
