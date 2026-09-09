---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 05
subsystem: compiler-core
tags: [go, check, corevalidate, originvalidate, call-admission, mutation-testing, diagnostic]

requires:
  - phase: 07-02
    provides: "Callable as publication safety (originvalidate.PublishProblemsFor, D-07-31/D-07-32), corevalidate's independent peerCallable re-derivation (D-07-33), testdata/phase07/clean_but_unpublishable.lang"
  - phase: 07-03
    provides: "core.OpCall with CalleeID resolved to a declared Lang function's ID; check.resolveCallBinding; the functionIDs table"
provides:
  - "check.Program builds an immutable per-program callSignatureTable (reusing core.FunctionSignature, which structurally carries no Linear/Match field) via originvalidate.BuildInterface, and consults it -- and only it -- to refuse a call to a non-callable callee with core.callee_not_callable (diagnostic.Error, no repairs, D-07-31c)"
  - "core.CalleeNotCallable ('core.callee_not_callable'), the checkpoint-ratified SEM-06 refusal code"
  - "corevalidate's OWN, independently-implemented SEM-06 refusal at both replay sites, using peerCallable (07-02) on the callee's own core.Function -- never check's table"
  - "testdata/phase07/call_uncallable_callee.lang -- SEM-06's negative control (a call to a checks-clean-but-unpublishable callee)"
  - "testdata/phase07/relay_escort_witness.lang -- D-04-03's decisive research witness, converted to A-normal form (D-07-44), which checks clean under Phase 07's rules while corevalidate independently refuses it with core.move_while_borrowed -- the interprocedural half of D-03-02, left open until Phase 08/09"
  - "The body-blindness control (verifyCallableRefusalBodyReadSeam/calleeBodyReadObserved), making SEM-05's structural claim falsifiable"
  - "control:call.admission_body_blind and control:call.callable_refusal, added to session.Phase7RequiredControls() and scripts/verify-phase7.sh"
affects: [07-06, 07-07, 07-08, 09]

actuals:
  tokens: 14700
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Reusing core.FunctionSignature (lang.interface/1) verbatim as an admission-time signature-table entry type, rather than minting a parallel type, so the body-blindness property (no Linear/Match field) is inherited from an already-declared invariant instead of re-asserted"
    - "A post-build second admission pass (verifyCallableRefusal), mirroring verifyCallInvariants' own precedent, for a predicate (Callable) that requires every function's body to already be individually lowered before it can be computed at all -- 'before any body admission' read as 'before the ONE decision the table exists to gate', not 'before any per-function lowering'"
    - "A hand-built synthetic core.Program (bypassing check.Program entirely) as the only way to hand corevalidate an admitted-call-to-non-callable-callee shape once check itself refuses that shape at the source"
    - "A named, single exception to a corpus-wide parity invariant (TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus), paired with a decisive test proving the divergence itself, rather than silently loosening the invariant"

key-files:
  created:
    - testdata/phase07/call_uncallable_callee.lang
    - testdata/phase07/relay_escort_witness.lang
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/core/core.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - internal/compiler/session/session_phase7.go
    - scripts/verify-phase7.sh

key-decisions:
  - "Checkpoint ratified core.callee_not_callable, constructed with diagnostic.Error (never ErrorWithRepairs) and no repairs, exactly as proposed and defaulted -- auto-approved under active auto-mode (gate=blocking, first/recommended option)."
  - "The signature table's Callable field is populated by calling originvalidate.BuildInterface/PublishProblemsFor directly -- consuming the established single source of truth for lang.interface/1's Callable predicate, not re-deriving it a second time inside check. The two GENUINELY independent derivations of the D-04-03 fact this phase's threat register requires are check's admission-time table consult versus corevalidate's own, separately-implemented peerCallable re-derivation -- never check versus originvalidate, which would just be the producer read twice."
  - "'Before any body admission runs' (D-07-34) is read as 'before the ONE admission decision the table exists to gate' (the callable consult in verifyCallableRefusal), not literally before every function's own per-function lowering: Callable is a predicate over a CHECKED body (it recomputes return origin from core.LinearOperation facts), so a signature that precedes every function's own admission cannot exist when the signature's own Callable bit is derived from that admission's output. The table is built once, immediately before its one consumer runs, after every function has been individually lowered."
  - "relay_escort_witness.lang's relay and escort BOTH declare a correct PublicOrigin (making both Callable) rather than leaving relay undeclared: an undeclared relay is refused by THIS PLAN's own SEM-06 arm before the interprocedural D-03-02 hazard the witness exists to demonstrate is ever reached. Declaring the origin correctly is what lets the call be admitted and the intended finding surface."
  - "corevalidate's move_while_borrowed refusal on relay_escort_witness.lang is a genuine, real divergence from check (which admits it) -- resolved by adding ONE named exception to the corpus-wide parity test plus a decisive test proving the divergence itself (TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed), never by silently loosening the invariant for every fixture."
  - "Test 5's corevalidate-side refusal is proven against a HAND-BUILT synthetic core.Program (syntheticCallToNonCallableCalleeProgram), not a .lang fixture through session.Check, because check itself now refuses this exact shape at the source -- there is no other way to hand corevalidate an admitted call to a non-callable callee."
  - "control:call.admission_body_blind and control:call.callable_refusal are declared, not implied, as CLI-lane-observable: the phase07 lane's own two fixtures (both Callable) genuinely exercise the same production admission arm with zero divergence, but neither fixture contains a refused call -- the refusing branch is proven only by in-process Go tests, exactly like control:dispatch.recognized_not_executed's own precedent from 07-04."

patterns-established:
  - "A post-build second pass consulting a just-built table, for any predicate whose input requires every function to already be individually checked -- the same shape verifyCallInvariants (07-03) and verifyCallableRefusal (this plan) both use."

requirements-completed: [SEM-06, SEM-05, QLT-08]

coverage:
  - id: D1
    description: "check.Program builds an immutable callSignatureTable via originvalidate.BuildInterface after every function is individually admitted, and refuses a call to a non-callable callee with core.callee_not_callable (diagnostic.Error, no repairs), asserted end to end against testdata/phase07/call_uncallable_callee.lang"
    requirement: "SEM-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallToNonCallableCalleeRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestBuildCallSignatureTableCallableMatchesPublishProblemsFor"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase07/call_uncallable_callee.lang (exactly one core.callee_not_callable diagnostic, no repairs)"
        status: pass
    human_judgment: false
  - id: D2
    description: "The signature table entry type (core.FunctionSignature) structurally carries no Linear/Match field, and the table is a same-package read-only accessor (lookup only) -- proven by reflection and by AST-parsing check.go's own method set"
    requirement: "SEM-05"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallSignatureTableEntryCarriesNoBodyReachableField"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallSignatureTableHasOnlyLookupMethod"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallSignatureTableBuiltBeforeCallableAdmissionRuns"
        status: pass
    human_judgment: false
  - id: D3
    description: "The body-blindness control is falsifiable, not merely asserted: across every testdata/phase07 fixture containing a call, admission consults only the table (zero body reads); engaging the fault seam makes the SAME instrumentation observe a body read, proving the control would have gone red"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallAdmissionBodyBlindControl"
        status: pass
    human_judgment: false
  - id: D4
    description: "The permitting-Callable seam wrongly admits call_uncallable_callee.lang; the production default refuses it; a synthetic gap (unpopulated table entry) is refused, never admitted"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallableRefusalSeamAdmitsUncallableCallee"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallableRefusalRefusesUnpopulatedTableEntry"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallableRefusalAcceptsPopulatedCallableEntry"
        status: pass
    human_judgment: false
  - id: D5
    description: "corevalidate independently refuses the same admitted-call-to-non-callable-callee shape on a hand-built synthetic core.Program, via its own narrowed peerCallable re-derivation; the two refusals (check's and corevalidate's) are disabled independently, each with the other still refusing; the bilateral case reports a false agreement and fails the gate"
    requirement: "QLT-08"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestVerifyCallableRefusalSeamCheckDisabledCorevalidateStillRefuses"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go#TestStage0SummaryMutationMatrix (fault6_callable_refusal_peer_disabled)"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go#TestStage0SummaryMutationMatrix (fault7_bilateral_callable_refusal)"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... -run 'BodyBlind|MutationMatrix' -count=2 -shuffle=on (identical per-control kill results)"
        status: pass
      - kind: unit
        ref: "go test -race ./internal/compiler/check/... ./internal/compiler/corevalidate/... (green)"
        status: pass
    human_judgment: false
  - id: D6
    description: "testdata/phase07/relay_escort_witness.lang: D-04-03's research witness converted to A-normal form (D-07-01/D-07-44), parses with zero diagnostics, every call argument is a bare identifier, and its check outcome (clean) is asserted explicitly as the interprocedural D-03-02 finding rather than a bare boolean -- with corevalidate's own independent core.move_while_borrowed refusal on the identical program asserted as a decisive, named divergence"
    requirement: "SEM-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessParsesCleanly"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessCallArgumentsAreANormalForm"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestRelayEscortWitnessBothFunctionsAreCallable"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_summary_peer_test.go#TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed"
        status: pass
    human_judgment: true
    rationale: "Whether the D-03-02 interprocedural finding is correctly characterized (a genuine open gap rather than a masked bug) benefits from human review of the reasoning, even though every mechanical assertion passes."

duration: ~140min
completed: 2026-09-08
status: complete
---

# Phase 07 Plan 05: Pre-Body Signature Table and SEM-06 Callee-Callable Refusal Summary

**A call is now admitted only when its callee's Callable bit (publication safety, D-04-03) is true, consulted through an immutable post-build signature table that structurally cannot reach a callee body -- refused otherwise with `core.callee_not_callable`, independently re-derived a second time by corevalidate, with the interprocedural loan-liveness gap this exposes (D-03-02) surfaced as a named, tested finding rather than papered over.**

## Performance

- **Duration:** ~140 min
- **Tasks:** 3 (plus one checkpoint:decision, auto-approved under active auto-mode)
- **Files modified:** 10 (2 created, 8 modified)

## Accomplishments

- **Task 1 (tracer):** `check.Program` builds an immutable `callSignatureTable` -- reusing `core.FunctionSignature` verbatim (which structurally carries no `Linear`/`Match` field, so a caller admission path holding it cannot reach a callee body even by following a pointer) -- via `originvalidate.BuildInterface`, after every function in the program has been individually admitted. A new post-build admission arm, `verifyCallableRefusal`, walks every `core.OpCall` and refuses with the checkpoint-ratified `core.callee_not_callable` when the callee's table entry is absent or `Callable` is false: `diagnostic.Error`, no repairs (D-07-31c). `testdata/phase07/call_uncallable_callee.lang` is the negative control -- a callee shaped exactly like `clean_but_unpublishable.lang` (checks clean, fails publication with `core.origin_omitted`), refused end to end with exactly one diagnostic.
- **Task 2:** `testdata/phase07/relay_escort_witness.lang` converts D-04-03's decisive research witness to A-normal form (D-07-01 made the recorded nested `relay(borrow mut buffer)` ungrammatical): bind the exclusive borrow first, pass that binding to the callee, then move the owner. Both `relay` and `escort` declare a correct `PublicOrigin` (both `Callable`), so the call is admitted under this plan's own SEM-06 arm -- and the witness still checks **clean**: `check`'s loan-liveness law (`computeLoanLastUses`) builds its use-chain from each binding's `RHS.Source`, but a `"call"` binding carries `RHS.Arguments`, so the call is invisible to that chain as a use of its argument, and the exclusive loan on `buffer` is treated as ending at its own creation point. `take buffer` is wrongly admitted while `aliased` remains live -- the INTERPROCEDURAL half of D-03-02, asserted explicitly, left open until Phase 08/09. `corevalidate`'s own, independently-implemented replay DOES catch this (`core.move_while_borrowed`) -- a genuine cross-peer divergence, resolved with one named exception to the corpus-wide parity test plus a decisive test proving the divergence itself.
- **Task 3:** The body-blindness control (`verifyCallableRefusalBodyReadSeam`/`calleeBodyReadObserved`) makes SEM-05's structural claim falsifiable: production admission touches only the table for a callee, proven across every `testdata/phase07` fixture with a call; engaging the fault seam makes the same instrumentation observe a body read. `corevalidate` now independently refuses the identical SEM-06 shape at both replay sites, using `peerCallable` (07-02) on the callee's own `core.Function` -- never check's table -- proven on a hand-built synthetic `core.Program` (check itself now refuses the shape at the source, so no `.lang` fixture can reach corevalidate with it). The two refusals are disabled independently (fault6/fault7 in the existing `TestStage0SummaryMutationMatrix`), with the bilateral case reporting a false agreement and failing the gate. `control:call.admission_body_blind` and `control:call.callable_refusal` join `session.Phase7RequiredControls()` and `scripts/verify-phase7.sh`'s mirror list.

## Task Commits

1. **Checkpoint: ratify the SEM-06 refusal code and confirm the rejected repair** -- auto-approved (`⚡ Auto-selected: Accept core.callee_not_callable with diagnostic.Error and no repairs.`); `AUTO_CFG=true`, gate defaulted to `blocking`, plan's own `<default>` is the first/recommended option.
2. **Task 1: The pre-body signature pass, and one call to a non-publishable callee refused through it** - `6bf99da` (feat)
3. **Task 2: The A-normal-form relay/escort dangling-alias witness** - `ccdab6c` (test)
4. **Task 3: The body-blindness control and its mutation kill** - `0c0b00e` (test)

## Files Created/Modified

- `internal/compiler/check/check.go` - `callSignatureTable`, `buildCallSignatureTable`, `verifyCallableRefusal`, the body-blindness fault seam, all wired into `Program()`
- `internal/compiler/check/check_test.go` - the full test suite for Tasks 1-3 (signature table, refusal, body-blindness, cross-package divergence tests)
- `internal/compiler/core/core.go` - `core.CalleeNotCallable` (`"core.callee_not_callable"`)
- `internal/compiler/corevalidate/corevalidate.go` - the SEM-06 `peerCallable`-based refusal at both `replayStraightLine`/`replayBlocks`, `functionByID` at both sites
- `internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go` - fault6/fault7 subtests, `syntheticCallToNonCallableCalleeProgram`
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` - the named `relayEscortWitnessModule` exception, `TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed`
- `internal/compiler/session/session_phase7.go` - `ControlCallAdmissionBodyBlind`, `ControlCallCallableRefusal`, added to `Phase7RequiredControls()`
- `scripts/verify-phase7.sh` - mirrored the two new control identifiers
- `testdata/phase07/call_uncallable_callee.lang` - SEM-06's negative control (new)
- `testdata/phase07/relay_escort_witness.lang` - the A-normal-form D-04-03 witness (new)

## Decisions Made

See `key-decisions` in frontmatter. The two most consequential: **(1)** "before any body admission runs" (D-07-34) cannot mean "before every function's own per-function lowering" without contradiction -- `Callable` is a predicate over a CHECKED body, so it necessarily postdates every function's own admission. It is read instead as "before the ONE decision this table exists to gate," matching `verifyCallInvariants`' own established post-build-pass precedent from 07-03. **(2)** `relay_escort_witness.lang`'s check-clean outcome, once both functions correctly declare their origin, revealed a genuine, previously-undiscovered divergence between `check` (admits) and `corevalidate` (refuses with `core.move_while_borrowed`) -- resolved as a named, tested exception rather than silently loosened, and used as the trigger for Task 3's own corevalidate-side SEM-06 refusal (which did NOT previously exist before this plan).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug/interaction] `verifyCallInvariantsSeam` needed to also suppress the new Callable admission arm**
- **Found during:** Task 1, running the full `go test ./...` suite
- **Issue:** `TestVerifyCallInvariantsSeamRestoresBothRefusals` (07-03) engages `verifyCallInvariantsSeam` to admit a call to an UNDECLARED name via a synthetic non-ID `CalleeID`. The new `verifyCallableRefusal` (this plan) ran regardless and correctly found that synthetic ID absent from the table, refusing it with `core.callee_not_callable` -- a DIFFERENT, newer refusal than the one that seam was written to suppress, breaking the pre-existing test's "the seam suppresses the source-level refusal too" expectation.
- **Fix:** Gated the new admission arm's whole block on `!verifyCallInvariantsSeam` too, with a doc comment explaining the shared-failure rationale (D-07-42).
- **Files modified:** `internal/compiler/check/check.go`
- **Verification:** `go test ./internal/compiler/check/...` green, including the pre-existing test.
- **Committed in:** `6bf99da` (Task 1)

**2. [Rule 1 - Bug, real divergence discovered] corevalidate independently refuses `relay_escort_witness.lang` with `core.move_while_borrowed`, while check admits it**
- **Found during:** Task 2, running `go test ./...` after authoring the fixture
- **Issue:** `TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus` (07-02) asserts every check-clean corpus fixture also `corevalidate.Validate`s cleanly. The new fixture is a genuine, deliberate exception to that invariant (the interprocedural D-03-02 finding this plan's Task 2 explicitly anticipated) -- corevalidate's own, independently-implemented loan tracking DOES treat the call as a use, catching what check's shadow-model liveness misses.
- **Fix:** Added the one named `relayEscortWitnessModule` exception to that test's loop, plus a decisive, dedicated test (`TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed`) proving the divergence itself rather than silently skipping it.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_summary_peer_test.go`
- **Verification:** `go test ./internal/compiler/corevalidate/...` green; the new test asserts the exact divergence.
- **Committed in:** `ccdab6c` (Task 2)

**3. [Rule 2 - Missing critical functionality, required by Task 3's own action text] `internal/compiler/session/session_phase7.go` and `scripts/verify-phase7.sh` needed edits, not listed in `files_modified`**
- **Found during:** Task 3, implementing "add them to `Phase7RequiredControls()`" literally
- **Issue:** The plan's `files_modified` list omits `internal/compiler/session/session_phase7.go` and `scripts/verify-phase7.sh`, but Task 3's own action text explicitly requires `control:call.admission_body_blind`/`control:call.callable_refusal` to be added to `Phase7RequiredControls()` -- and T-07-24's established set-equality precedent (07-04) requires the script's mirror list to move in lockstep, or `TestPhase7RequiredControlsMatchScript` fails immediately.
- **Fix:** Added both control identifiers to both files, with doc comments declaring honestly what the CLI lane's own fixtures do and do not directly exercise for each (mirroring 07-04's own A-02 declaration style for cgen).
- **Files modified:** `internal/compiler/session/session_phase7.go`, `scripts/verify-phase7.sh`
- **Verification:** `TestPhase7RequiredControlsMatchScript` passes; `lang --json verify testdata/phase07` includes both new controls in its lane output.
- **Committed in:** `0c0b00e` (Task 3)

---

**Total deviations:** 3 (1 interaction bug in this plan's own new code, 1 genuine divergence discovered and honestly documented rather than hidden, 1 missing-file addition required by the task's own literal text). **Impact:** All necessary for correctness or for satisfying acceptance criteria the plan itself specified. No scope creep beyond Task 1-3's own asks.

## Known Stubs

None. The signature table, both admission arms (check's and corevalidate's), both fault seams, and all instrumentation are fully wired and exercised.

## Threat Flags

None beyond what `07-05-PLAN.md`'s own `<threat_model>` already registered (T-07-27 through T-07-32, T-07-SC) -- all mitigated as designed:
- T-07-27 (elevation of privilege via `Callable`): mitigated by the fail-closed refusal (absent table entry refuses, never admits) and the stable ratified code.
- T-07-28 (information disclosure via a callee-body read): mitigated by `core.FunctionSignature`'s structural absence of a body-reachable field, plus the falsifiable body-blindness control.
- T-07-29 (a repair that does not repair): mitigated -- `diagnostic.Error`, no repairs, verified directly.
- T-07-30 (the wrong predicate enforced by a green gate): mitigated -- the negative control is a publication failure, never an export omission.
- T-07-31 (`Callable` derived once): mitigated -- check's table consult and corevalidate's own `peerCallable` are genuinely independent, disabled independently, with the bilateral case failing the gate.
- T-07-32 (production code reachable through a fault seam): mitigated -- every new seam is unexported, restored via explicit reset/defer, `go test -race` green.

## Issues Encountered

**Pre-existing environmental flakiness in `cache`/`measure`/`cgen` packages under heavy parallel `go test -race ./...` load, unrelated to this plan.** `sh scripts/verify-phase7.sh`'s own `go test -race ./...` step intermittently failed on `TestCacheKeyCoversEveryDeclaredInput`/`TestMutationRunnerSourceHashIsADeclaredInput` (`cache.input_undeclared`), `TestMachineIDExcludesHostFingerprints`/`TestMachineProbeIsBounded` (`measure.probe_timeout`), and `TestPhase5ByPointerLoweringThreeEngineAgreement` (`native.timeout`) -- all three are fixed-wall-clock-timeout subprocess-spawn probes in packages this plan never touches. Every one passes cleanly when run in isolation or under lighter concurrent load; the flakiness tracks system resource contention, not this plan's diff. Verified independently: `go test ./...` (non-race) full-suite green; `go test -race ./internal/compiler/check/... ./internal/compiler/corevalidate/...` (the two packages this plan touches) green; `go vet ./...` clean; every `lang --json verify testdata/phaseN` CLI invocation (phase1 through phase07) passes with status `"pass"`, and `testdata/phase07`'s own JSON output carries both new control identifiers verbatim -- the exact facts `scripts/verify-phase7.sh` would have asserted had the unrelated `-race` step not been contended out.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SEM-06's refusal and SEM-05's body-blindness claim are each independently derived twice (check + corevalidate) and each observed to fail under a seeded mutation, with the bilateral case failing the gate rather than passing it -- ROADMAP Phase 07 success criterion 2 is met in full.
- `07-06`'s call-graph work inherits `Phase7RequiredControls()` (now 5 entries) as the pattern to extend, and `core.CalleeNotCallable`/`core.CallCalleeUnresolved` as the two adjacent, already-distinct codes it must not collide with when minting its own cycle-refusal code.
- Phase 08/09 has one concrete, tested, named finding to close: `testdata/phase07/relay_escort_witness.lang` demonstrates the interprocedural half of D-03-02 -- a call masks loan liveness across the function boundary in `check`, though `corevalidate` already, incidentally, catches this one shape. Closing it means making `check`'s own admission agree with `corevalidate`'s, never the reverse.
- Ready for `07-06`.

## Self-Check: PASSED

- All key-files (created + modified) verified present on disk with `[ -f ]`.
- All three task commits (`6bf99da`, `ccdab6c`, `0c0b00e`) verified present via `git log --oneline --all`.
- Re-ran every task's `<verify>` and `<acceptance_criteria>` commands: all pass, including `go run ./cmd/lang --json check testdata/phase07/call_uncallable_callee.lang`, the full `-run` pattern matches for each task, `go test ./... && go vet ./...`, `go test -race ./internal/compiler/check/... ./internal/compiler/corevalidate/...`, and `-count=2 -shuffle=on` producing identical per-control kill results.
- `go build -o /tmp/lang-verify ./cmd/lang && /tmp/lang-verify --json verify testdata/phase07` confirmed `"status":"pass"` with both new control identifiers present in the lane's `"controls"` array; the same binary against `testdata/phase1` through `testdata/phase6` all report `"status":"pass"`.
- The one plan-level verify command not observed green end-to-end as literally written (`sh scripts/verify-phase7.sh`, whose own `go test -race ./...` step hit unrelated environmental flakiness) is documented above under "Issues Encountered" with the equivalent passing evidence, gathered piece by piece, that covers everything the script itself would have asserted.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-08*
