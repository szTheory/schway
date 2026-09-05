---
phase: 04-fallible-resources-and-c-boundary
plan: "05"
subsystem: compiler
tags: [ffi, nonlocal-exit, setjmp, longjmp, undefined-symbols, nm, mutation-testing, c17]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "plan 01-04's OpForeignCall/OpFail/OpRelease/OpDefect trio, checkResourceLifecycle's reverse-order release materialization, the byte-frozen native/lang_foreign_resource.c + private header, the closed terminal-outcome axis, the streaming event emitter, signal-aware exit adjudication, verifyForeignCorpus's existing required controls"
provides:
  - "cgen.emitNonlocalPad: D-04-17's ONE process-root setjmp landing pad per foreign-acquiring generated program, reading a static-storage resource ledger and running NO release (D-04-18)"
  - "native/lang_foreign_nonlocal.c: the second frozen foreign translation unit whose symbol acquires normally on its first call and performs a real longjmp on its second, giving the pad a reachable witness"
  - "interp.runLinearBlocks' shared, documented nonlocal-exit-on-second-call convention, kept in byte-for-byte agreement with the native pad's own event sequence and terminal record"
  - "internal/compiler/native/symbols.go: nm -u undefined-symbol listing, Mach-O/ELF normalization, and CheckUndefinedSymbolAllowlist against a checked-in positive allowlist with an honest SymbolsOperational status when the tool is absent"
  - "session.verifyForeignCorpus's two new required controls: control:foreign.unwind_forbidden and control:foreign.nonlocal_exit_undetected, the latter mutation-killed two different ways"
  - "native.ForeignSourcePathForSymbol: session.RunNative's foreign-TU auto-wiring now resolves by declared symbol instead of hardcoding the first frozen TU"
affects: [04-06, 04-07]

actuals:
  tokens: 19372
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "The process-root landing pad is installed unconditionally, once, inside emitLinearForeign's own preamble -- 'exactly one pad' is a structural property of the call site, not a runtime count, so it never needs a per-function policy check to decide whether to install it."
    - "Detection convention for a foreign nonlocal exit the interpreter cannot literally perform: a foreign contract whose declared nonlocal_exit policy is not 'forbidden' is understood, by documented shared convention with the frozen TU's own static call counter, to perform its nonlocal exit on its SECOND call within one function execution -- never its first, so at least one acquisition is already live."
    - "A generated artifact mutation-kill runner (NonlocalPadOmissionMutationRunner / NonlocalLedgerOmissionMutationRunner) deletes a marker-bracketed SPAN of generated C, not a single line, when the deleted code cannot be a single line without leaving an unmatched brace."
    - "Foreign-TU auto-wiring resolves by declared symbol name (native.ForeignSourcePathForSymbol), not by hardcoding the sole existing TU -- the moment a second frozen TU exists, hardcoding the first one silently breaks linking any program declaring the second."

key-files:
  created:
    - native/lang_foreign_nonlocal.c
    - testdata/phase4/nonlocal_exit_probe.lang
    - internal/compiler/native/foreign_nonlocal.go
    - internal/compiler/native/symbols.go
    - internal/compiler/native/symbols_test.go
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "The nonlocal-exit probe fixture uses ONE declared foreign symbol for both the acquisition and the nonlocal-exit trigger (not two different symbols), because the current codebase model is strictly one symbol per foreign-acquiring function -- cgen.go's emitLinearForeign resolves the extern call name once, from the function-level ForeignContract, and calls it uniformly at every OpForeignCall site. The frozen TU's own static call counter (first call acquires, second call longjmps) and the interpreter's per-function call counter are the shared, documented convention that makes this work without any check.go/corevalidate.go changes."
  - "The pad's own function.defected event is assembled manually with lang_write_json_string/lang_write_literal calls (emitManualDefectEvent), exactly mirroring emitBranchOperations' existing OpDefect case, because the shared LANG_EVENT struct has no 'output' field and adding one there would move every other emitter's generated C (D-04-23)."
  - "LiveResources and each resource.leaked event's SourcePlace use the acquisition's TARGET PLACE id (matching cgen's existing resourceLedger convention), not the operation id liveResourceList uses for every OTHER terminator -- a narrower, path-scoped convention (liveResourcePlaces) chosen specifically so execution.Equal's canonical-bytes comparison agrees between the two engines on this new path, without touching the pre-existing convention any other terminator relies on."
  - "session.RunNative's foreign-source auto-wiring, previously hardcoded to native.ForeignResourceSourcePath() for any foreign-contract function, now resolves via native.ForeignSourcePathForSymbol(contract.Symbol) -- a Rule 1 bug fix required the moment a second frozen TU exists, or `lang run --engine=native` would try to link the nonlocal-exit probe against the wrong translation unit and fail at compile time rather than at the (separately accepted) run-time expectation mismatch."

patterns-established:
  - "A landing-pad/ledger-population marker pair (padInstallMarker/padEndMarker, ledgerPopulateMarker) duplicated verbatim between cgen.go (the producer) and session.go (the mutation-runner consumer) -- the same duplication shape releaseMarker already established for lang:release-site."

requirements-completed: []

coverage:
  - id: D1
    description: "One process-root setjmp landing pad, reading a static-storage resource ledger, emits one foreign.nonlocal_exit event plus one resource.leaked event per still-live acquisition, then a function.defected terminal record, then aborts -- with a second frozen foreign TU giving it a reachable witness through the real generated pipeline, and the interpreter agreeing byte-for-byte via a shared documented convention"
    requirement: RES-01
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestExactlyOneLandingPadIsInstalled"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestLedgerIsStaticStorage"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestNonlocalExitEmitsLeakPerLiveAcquisition"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestPadRunsNoRelease"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestNonlocalExitProbeInterpreterNative"
        status: pass
    human_judgment: false
  - id: D2
    description: "An nm -u undefined-symbol allowlist control: a positive, rationale-carrying list where any new undefined symbol fails until a human adds it, an honest operational status when the listing tool is absent, and no dependence on unwind-section presence"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/native/symbols_test.go#TestUndefinedSymbolAllowlistRejectsNewSymbol"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/symbols_test.go#TestUndefinedSymbolNormalizationHandlesBothFormats"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/symbols_test.go#TestMissingSymbolToolReportsOperational"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/symbols_test.go#TestUnwindControlDoesNotInspectSections"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4UnwindControl"
        status: pass
    human_judgment: false
  - id: D3
    description: "control:foreign.nonlocal_exit_undetected is mutation-killed two different ways (pad-span omission; ledger-population omission surfacing specifically as a leak-count disagreement), and the detection mechanism's reachable input space -- including two named, never-claimed-covered blind spots -- is recorded rather than implied covered"
    requirement: RES-01
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestNonlocalExitDetectionIsMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestLeakCountMatchesLiveAcquisitions"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestNonlocalExitReachabilityIsRecorded"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestBlindSpotsAreNamedNotClaimed"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4NonlocalExitControl"
        status: pass
    human_judgment: false

duration: ~2h (single continuous session)
completed: 2026-09-04
status: complete
---

# Phase 4 Plan 5: Fallible Resources and C Boundary — Nonlocal-Exit Landing Pad and Undefined-Symbol Allowlist Summary

**A real setjmp/longjmp process-root landing pad detects a foreign nonlocal exit through a second frozen foreign translation unit and reports every still-live acquisition as a leak (never a release), an `nm -u` positive allowlist gates every undefined symbol a linked binary pulls in with an honest tool-missing status, and both new controls are mutation-killed rather than merely asserted.**

## Performance

- **Duration:** ~2h (single continuous session)
- **Started:** 2026-09-04 (session start)
- **Completed:** 2026-09-04
- **Tasks:** 3/3 completed
- **Files modified:** 11 (5 created, 6 modified)

## Accomplishments

- `cgen.emitNonlocalPad` installs exactly ONE `setjmp()` call at the top of every foreign-acquiring generated program's `main()`, immediately after the streamed events array opens. Verified across one-, two-, and three-acquisition fixtures (`TestExactlyOneLandingPadIsInstalled`). The pad's body reads only the ALREADY-static-storage resource ledger (`TestLedgerIsStaticStorage`), emits `foreign.nonlocal_exit`, one `resource.leaked` per still-live acquisition, a manually-assembled `function.defected` terminal event, then calls `abort()` — and never a release (`TestPadRunsNoRelease`).
- `native/lang_foreign_nonlocal.c` is the second byte-frozen foreign translation unit: its single symbol acquires a real one-byte `malloc`-backed block on its first call and performs a genuine `longjmp` back into Lang's generated `lang_nonlocal_landing` on its second — giving the pad a reachable witness through the real compiled-and-linked pipeline, not a hand-written C harness.
- `testdata/phase4/nonlocal_exit_probe.lang` is the shipped witness: it acquires once (becomes live), then calls the same declared symbol again, which never returns. The interpreter's `runLinearBlocks` models the identical shared convention (a foreign contract whose `nonlocal_exit` policy is not `"forbidden"` triggers on its second call) and agrees byte-for-byte with the real compiled binary's own event stream and terminal record (`TestNonlocalExitProbeInterpreterNative`, via `execution.Equal`'s canonical-bytes comparison).
- `internal/compiler/native/symbols.go` adds `nm -u` undefined-symbol listing (its own bounded, timed `os/exec` invocation), Mach-O/ELF leading-underscore normalization, and `CheckUndefinedSymbolAllowlist` against a checked-in POSITIVE allowlist (`AllowedUndefinedSymbols`) where every entry carries a written rationale. A missing `nm` reports `SymbolsOperational`, never `SymbolsPass` (`TestMissingSymbolToolReportsOperational`). The control never inspects unwind-section presence (`TestUnwindControlDoesNotInspectSections`, which confirms the project's own tracer binary genuinely carries Apple's mandatory compact-unwind section yet the control still passes on symbol membership alone).
- `control:foreign.unwind_forbidden` and `control:foreign.nonlocal_exit_undetected` are both registered in `session.verifyForeignCorpus` with nonzero recomputed work (`TestVerifyPhase4UnwindControl`, `TestVerifyPhase4NonlocalExitControl`). The nonlocal-exit control is mutation-killed TWICE, attacking the emitter's own generated C: once by deleting the pad's entire generated span (`TestNonlocalExitDetectionIsMutationKilled`), and once by dropping one ledger-population site so the leak count silently UNDERSTATES the true live set — checked specifically as a leak-count disagreement, not merely "some difference" (`TestLeakCountMatchesLiveAcquisitions`).
- The probe fixture and `emitNonlocalPad` both carry a REACHABILITY comment naming what the mechanism reaches and, per D-10, what it does NOT (`TestNonlocalExitReachabilityIsRecorded`): a landing point established BELOW the pad by a foreign-invoked callback, and a foreign call that terminates the process directly (`exit()`/`_exit()`). Both are named as accepted residual limitations in the `lang.foreign/0` sidecar's `unchecked_obligations` list and asserted, by scanning the fixture/cgen.go/the frozen TU, never claimed as covered anywhere (`TestBlindSpotsAreNamedNotClaimed`).

## Task Commits

Each task was committed atomically:

1. **Task 04-05-01: The process-root landing pad and its static-storage cleanup ledger** — `59e1587` (feat)
2. **Task 04-05-02: The undefined-symbol allowlist, with a tool-missing status that never passes** — `ae83e3b` (feat)
3. **Task 04-05-03: Prove detection is falsifiable and that criterion 3 is not satisfied vacuously** — `03e0f1f` (feat) — this commit also carries the `session.RunNative` foreign-source-by-symbol fix (see Deviations)

**Plan metadata:** committed alongside this SUMMARY.

## Files Created/Modified

- `native/lang_foreign_nonlocal.c` — the second frozen foreign translation unit (call-count-based acquire/longjmp)
- `testdata/phase4/nonlocal_exit_probe.lang` — the shipped witness, with its REACHABILITY comment
- `internal/compiler/native/foreign_nonlocal.go` — `ForeignNonlocalSourcePath`, `ForeignSourcePathForSymbol`
- `internal/compiler/native/symbols.go` — `nm -u` listing, normalization, `CheckUndefinedSymbolAllowlist`, `Runner.CompileOnly`
- `internal/compiler/native/symbols_test.go` — the five task-02 tests
- `internal/compiler/cgen/cgen.go` — `emitNonlocalPad`, `emitManualDefectEvent`, `resourceLedger.emitLeakEvents`, the pad/ledger markers, extended `uncheckedForeignObligations`
- `internal/compiler/cgen/cgen_test.go` — pad-installation, ledger-storage, reachability, and blind-spot tests
- `internal/compiler/interp/interp.go` — the shared nonlocal-exit-on-second-call convention in `runLinearBlocks`, `liveResourcePlaces`
- `internal/compiler/native/native.go` — `foreign.nonlocal_exit`/`resource.leaked` event-kind validation cases
- `internal/compiler/session/session.go` — the two new required-control lanes, the two mutation runners, the `RunNative` symbol-by-source fix
- `internal/compiler/session/session_test.go` — the task-01/03 differential, mutation-kill, and required-control tests

## Decisions Made

See `key-decisions` in the frontmatter. The most consequential: the probe uses ONE foreign symbol whose real C behavior differs by call count (rather than two distinct symbols), because the codebase's current model is one symbol per foreign-acquiring function — this is a real, structural constraint of `cgen.go`'s `emitLinearForeign`, not a plan-05 shortcut, and it made the whole design tractable without touching `check.go`/`corevalidate.go` at all.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `session.RunNative`'s foreign-source auto-wiring hardcoded the first frozen TU**
- **Found during:** Task 3, while running the nonlocal-exit probe and a non-corpus probe through a freshly built `./cmd/lang` per the plan's own verification requirement
- **Issue:** `RunNative` always appended `native.ForeignResourceSourcePath()` for ANY foreign-contract function, regardless of which symbol it actually declared. Once `native/lang_foreign_nonlocal.c` existed as a second frozen TU, `lang run --engine=native testdata/phase4/nonlocal_exit_probe.lang` failed at COMPILE time (`native.compile_failed`, an undefined-symbol link error) rather than reaching the run at all.
- **Fix:** Added `native.ForeignSourcePathForSymbol(symbol)`, resolving the correct frozen TU by the function's own declared `ForeignContract.Symbol`; `RunNative` now calls it instead of hardcoding the first TU.
- **Files modified:** `internal/compiler/native/foreign_nonlocal.go`, `internal/compiler/session/session.go`
- **Verification:** `lang run --engine=native testdata/phase4/foreign_acquire_one.lang` still passes cleanly; `lang run --engine=native testdata/phase4/nonlocal_exit_probe.lang` now compiles and links (see Verification Performed below for its remaining, separately-accepted run-time shape). Full `go test -count=1 ./...` green after the fix.
- **Committed in:** `03e0f1f` (part of Task 3's commit; see Task Commits above)

---

**Total deviations:** 1 auto-fixed (Rule 1 — a bug found and fixed during implementation, before the affected commit landed). **Impact on plan:** Necessary for the plan's own "shipped-binary run... through a freshly built ./cmd/lang" verification requirement to mean anything for a second frozen TU; no scope creep.

## Issues Encountered

None beyond the one Rule-1 fix above.

## Revert-and-Fail Demonstrations (D-09)

### control:foreign.nonlocal_exit_undetected — mutation 1: pad-span omission (deleting the entire generated setjmp/pad block)

```
=== RUN   TestZZDemoPadOmission
    zzdemo1_test.go:21: pad-omission mutation result: err=native.run_signaled: process terminated by signal segmentation fault: signal: segmentation fault result={Optimization: Pairs:[] CompileTime:0s RunTime:0s OutputBytes:0}
--- FAIL: TestZZDemoPadOmission (0.35s)
FAIL
```

With the pad removed, `lang_nonlocal_landing` is declared but never established via `setjmp()`; the frozen TU's second call still performs its `longjmp` into it, which is undefined behavior against an unestablished jump buffer — on this host it segfaults rather than producing any execution document at all. Detection is unambiguously absent: no `foreign.nonlocal_exit` event, no `resource.leaked` event, no clean `defect` outcome — nothing.

### control:foreign.nonlocal_exit_undetected — mutation 2: ledger-population omission (leak-count disagreement, not a crash)

```
=== RUN   TestZZDemoLedgerOmission
    zzdemo2_test.go:33: leak-count disagreement: golden=1 mutated=0 mutatedErr=<nil> mutatedExecution={Optimization:-O0 Pairs:[{Input:7 Execution:{Schema:lang.execution/1 Outcome:{Kind:defect Value:} Events:[{...Kind:foreign.called...} {...Kind:foreign.nonlocal_exit...} {...Kind:function.defected...Output:foreign nonlocal exit detected at process-root landing pad...}] LiveResources:[]}}] CompileTime:49.239875ms RunTime:156.796375ms OutputBytes:995}
--- FAIL: TestZZDemoLedgerOmission (0.46s)
FAIL
```

Golden run: 1 `resource.leaked` event, `LiveResources` has 1 entry. Mutated run (one `lang_resource_live[0] = 1;` population line deleted): the program still compiles, links, and cleanly reaches the pad and aborts — but the pad's own leak accounting silently reads the never-flipped `lang_resource_live[0]` as still 0, so ZERO `resource.leaked` events are emitted and `LiveResources` is empty. This is exactly the "leak count understates the live set" failure mode D-04-21 requires be caught SPECIFICALLY as a leak-count disagreement, not an incidental difference — confirmed by the diagnosis above (a silent understatement, not a crash, unlike mutation 1).

Both demonstrations were run as throwaway, uncommitted Go test files (`internal/compiler/session/zzdemo{1,2}_test.go`), deleted immediately after capturing output; nothing from either demonstration reached any commit. `go test -count=1 ./...` re-ran clean after both were removed.

## Verification Performed

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test -count=1 ./...` — pass
- `go test -race ./...` — pass
- `go vet ./...` — clean
- `git diff HEAD -- go.mod go.sum testdata/phase1 testdata/phase2 testdata/phase3 native/lang_foreign_resource.c native/lang_foreign_resource_private.h` — empty (measured against the pre-plan-05 tree)
- `sh scripts/verify-phase3.sh` — exits 0
- All named tests per each task's `<verify>` block, run via `scripts/assert-go-tests.sh ./internal/compiler/...` — pass:
  - Task 1: `TestExactlyOneLandingPadIsInstalled TestLedgerIsStaticStorage TestNonlocalExitEmitsLeakPerLiveAcquisition TestPadRunsNoRelease TestNonlocalExitProbeInterpreterNative`
  - Task 2: `TestUndefinedSymbolAllowlistRejectsNewSymbol TestUndefinedSymbolNormalizationHandlesBothFormats TestMissingSymbolToolReportsOperational TestUnwindControlDoesNotInspectSections TestVerifyPhase4UnwindControl`
  - Task 3: `TestNonlocalExitDetectionIsMutationKilled TestLeakCountMatchesLiveAcquisitions TestNonlocalExitReachabilityIsRecorded TestBlindSpotsAreNamedNotClaimed TestVerifyPhase4NonlocalExitControl`
- Shipped-binary runs through a freshly built `./cmd/lang` (`format --check`, `check`, `run --engine=interpreter`, `run --engine=native`):

  **`testdata/phase4/nonlocal_exit_probe.lang`** (the shipped corpus fixture):
  ```
  format --check: pass, exit=0
  check:          pass, exit=0
  run --engine=interpreter: pass, exit=0 (foreign.called, foreign.nonlocal_exit, resource.leaked, function.defected; outcome=defect)
  run --engine=native:      operational_failure, exit=3, code=native.run_signaled
  ```

  **A hand-written, non-corpus probe** (`/tmp/noncorpus_nonlocal.lang`, same shape, different module name, never committed):
  ```
  format --check: pass, exit=0
  check:          pass, exit=0
  run --engine=interpreter: pass, exit=0 (identical event sequence to the corpus fixture)
  run --engine=native:      operational_failure, exit=3, code=native.run_signaled
  ```

  **Honest note on the native-engine CLI result:** `lang run --engine=native` always calls `session.RunNative`, which has no per-fixture way to declare "this program's outcome is `defect`" — it always validates against the default `ExpectValue` ("returned") contract. A real `SIGABRT` (this plan's own correct, intended behavior) is therefore reported as `native.run_signaled` by the generic CLI path, exactly like the PRE-EXISTING `testdata/phase4/defect_terminal.lang` fixture from plan 04 already does through the same CLI command (verified side by side above the fix). This is not a regression this plan introduces — it is an existing, accepted limitation of the generic `run --engine=native` CLI command's fixed expectation, unrelated to and unaffected by this plan's `session.RunNative` symbol-resolution fix (which only fixed the earlier, and worse, COMPILE-time link failure). `control:defect.signal_adjudicated` and `control:foreign.nonlocal_exit_undetected` (both in `session.verifyForeignCorpus`) exercise the correct `Expect`-aware path and pass cleanly — the CLI-level narrowing is cosmetic to the gate's own correctness.

## Known Stubs

None that block this plan's own success criteria.

**Documented narrowing (not a stub):** the interpreter's nonlocal-exit-on-second-call convention (`interp.go`'s `runLinearBlocks`) is a Claude's-Discretion modeling choice specific to this plan's single-probe-fixture shape, not a general-purpose simulation of arbitrary nonlocal-exit timing. It is exercised and proven correct only for exactly the shape `nonlocal_exit_probe.lang` has (two calls to one symbol, second call exits). A later plan adding a differently-shaped nonlocal-exit fixture (e.g., the exit occurring on the first call, or after three acquisitions) would need to extend this convention, not assume it generalizes.

**Documented narrowing (not a stub):** `interp.go`'s `liveResourcePlaces` (place-id-based `LiveResources`) is intentionally a DIFFERENT identifier convention than `liveResourceList` (operation-id-based, used by every other terminator: `OpReturn`/`OpFail`/arm-level `OpDefect`). This is scoped narrowly to the nonlocal-exit path this plan adds and chosen specifically so `execution.Equal`'s literal byte comparison agrees with `cgen`'s existing `resourceLedger` place-id convention — it does not retroactively change what any pre-existing terminator reports.

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The process-root nonlocal-exit landing pad, the undefined-symbol allowlist, and both new required controls (`control:foreign.unwind_forbidden`, `control:foreign.nonlocal_exit_undetected`) are in place, mutation-killed, and wired into `session.verifyForeignCorpus`.
- **RES-01 stays `PARTIAL`, as directed by phase state:** return and typed-failure paths release (closed in plan 02); the nonlocal-exit path reports leaks rather than releasing, per D-04-18's forbidding of release from indeterminate state. This plan does not force RES-01 (or FFI-01) closed in `REQUIREMENTS.md` — both remain declared by sibling plans in this phase per the standing shared-ID gate, and no tension requiring escalation arose: the flagged assumption in the plan's own `<flagged_assumptions>` ("if RES-01 is read as requiring cleanup on every path including this one, it cannot be satisfied without introducing the use-after-free D-04-18 exists to prevent") did not surface as a real conflict — the honest-leak-report reading is exactly what this plan delivers, and no reviewer signal contradicted it.
- Carried-forward debt, unchanged from plans 01-04: `originvalidate`/`pathoracle` still only walk `OpReturn` (D-04-29's widening to `{OpReturn, OpFail, OpDefect}` remains carried to plan 06).
- New, plan-05-local narrowing (documented above): the interpreter's nonlocal-exit convention and `liveResourcePlaces`' place-id scheme are both scoped to this plan's single fixture shape; a later plan adding a differently-shaped nonlocal-exit fixture should revisit both.
- No blockers.

## Self-Check: PASSED

All key files verified present on disk; all three task commits (`59e1587`, `ae83e3b`, `03e0f1f`) verified present in `git log`; full test suite, race detector, vet, and `verify-phase3.sh` all green as of this summary.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-04*
