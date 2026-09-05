---
phase: 04-fallible-resources-and-c-boundary
plan: "07"
subsystem: verification-gate
tags: [go, cgen, interp, native, corevalidate, evidence, foreign-ffi, mutation-testing]

requires:
  - phase: 04-fallible-resources-and-c-boundary (plans 01-06)
    provides: every Phase 4 control, fixture, and lane this gate requires and re-asserts
provides:
  - scripts/verify-phase4.sh, the bounded Phase 4 gate requiring every control identifier and re-running Phase 1-3 corpora with one binary
  - session.Phase4RequiredControls()/Phase4ThreeEngineDifferential/Phase4CompareThreeEngines, the session-layer control list and differential ROADMAP SC4 needs
  - a fixed `lang run --engine=native` (session.runNativeInputs), previously wrong for any defect/typed-failure terminal outcome
  - the phase's final debt and escape register (04-DEBT.md, 04-VALIDATION.md), and closure of SEM-03/RES-01/FFI-01
affects: [phase-05-planning]

actuals:
  tokens: 210000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "session-layer required-control list cross-checked against the gate script's own text by a contract test (D-04-22)"
    - "structural core-artifact mutation plus a throwaway, non-frozen foreign test-double object to genuinely execute an otherwise-unreachable typed-failure path"
    - "narrow, test-only interpreter entry point (RunLinearBlockDirect) for exercising a block the public Run() entry point cannot reach"

key-files:
  created:
    - scripts/verify-phase4.sh
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/native/native_test.go
    - internal/compiler/evidence/evidence_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "session.Phase4RequiredControls() is the single canonical 12-identifier list; scripts/verify-phase4.sh's own text is cross-checked against it by TestPhase4RequiredControlsMatchScript rather than either side being hand-kept in sync"
  - "The second-stage/third-stage typed-failure paths are proven via a genuine three-engine EXECUTION, not a hand-constructed document: interp.RunLinearBlockDirect drives the interpreter's real OpRelease/OpFail case bodies directly on the failure block (Run's own entry point cannot reach it, since the interpreter always simulates foreign-call success), while a throwaway, non-frozen test-double foreign object makes the REAL generated C's own already-compilable-but-dead 'if (!result.ok)' branch genuinely execute at both optimization levels"
  - "cgen.go was explored for a structural fix (letting an OpForeignCall's ok edge lead to a Fail-terminated block) and reverted: corevalidate's own core.fail_reached_without_err_edge independently refuses exactly that shape, by design, so the only route to genuine execution is a real ok=0 at runtime, not a core-artifact reroute"
  - "session.runNativeInputs derives each input's native.Runner.Expect from the SAME interpreter verdict this function already computes as its own oracle for that input, fixing two real bugs `lang run --engine=native` had never been exercised against: a Match-shaped multi-arm function whose arms disagree on terminal-outcome kind, and any single-input program whose real outcome is a defect (including the already-committed nonlocal_exit_probe.lang fixture)"
  - "RES-01/FFI-01 close on the reading D-04-CONTEXT's own flagged_assumptions recorded: RES-01 as cross-engine cleanup-event agreement on return and typed-failure paths (the nonlocal-exit path deliberately reports leaks rather than releasing, per D-04-18, and stays a documented, non-overstated partiality); FFI-01 as obligations made inspectable and violations detectable, never as proof the foreign implementation obeys its contract (permanent, non-discharging quarantine)"

patterns-established:
  - "A gate script's required-control list lives in the session package as an exported function, not only as prose in the script or the validation doc, so a contract test can assert the two never drift apart"
  - "A structural core-artifact mutation is preferred over a hand-constructed execution document whenever the mutation can be fed through the SAME production interp/cgen/native machinery every other differential uses"

requirements-completed: [SEM-03, RES-01, FFI-01]

coverage:
  - id: D1
    description: "scripts/verify-phase4.sh: a bounded gate that builds the shipped binary once, verifies Phase 1-4 corpora with it, and requires every Phase 4 control identifier and expected escape by exact text, contract-tested against its own script text"
    requirement: "SEM-03"
    verification:
      - {kind: integration, ref: "internal/compiler/session/session_test.go#TestVerifyPhase4ControlsAndWork", status: pass}
      - {kind: integration, ref: "internal/compiler/session/session_test.go#TestVerifyPhase4CLI", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestPhase4VerifierScriptContract", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestPhase4RequiredControlsMatchScript", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestExpectedEscapesAreVisibleNotSolved", status: pass}
      - {kind: other, ref: "sh scripts/verify-phase4.sh", status: pass}
    human_judgment: false
  - id: D2
    description: "All three engines (interpreter, -O0, -O3) genuinely agree on terminal outcome, ordered events, and live-resource state across all five Phase 4 path shapes: success, second-stage typed failure, third-stage typed failure, defect, and nonlocal exit"
    requirement: "RES-01"
    verification:
      - {kind: integration, ref: "internal/compiler/session/session_test.go#TestPhase4CorpusThreeEngineAgreement", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestPhase4DifferentialNamesFirstDisagreement", status: pass}
      - {kind: unit, ref: "internal/compiler/evidence/evidence_test.go#TestForeignDigestMismatchRefused", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestReleaseOmissionMutationIsMismatch", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestReleaseTranspositionMutationIsMismatch", status: pass}
    human_judgment: false
  - id: D3
    description: "Every Phase 4 behavior is demonstrated on the shipped binary against a hand-written, out-of-corpus program; the debt and escape register is finalized and no test/comment/fixture claims coverage of a named residual"
    requirement: "FFI-01"
    verification:
      - {kind: integration, ref: "internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestNoCoverageClaimedForNamedResiduals", status: pass}
      - {kind: unit, ref: "internal/compiler/session/session_test.go#TestPhase4ReachabilityRecordIsComplete", status: pass}
    human_judgment: true
    rationale: "The debt register's honesty (D-04-32's wording, RES-01's recorded partiality) and the out-of-corpus fixtures' genuine novelty relative to the corpus are best judged by a human reader, not solely by the mechanical checks that back them"

duration: 55min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 07: Close the Phase Summary

**One bounded gate script requiring all twelve Phase 4 controls and three prior phases' non-regression, a genuinely-executed three-engine differential across all five Phase 4 path shapes (including two real `lang run --engine=native` bugs found and fixed along the way), and the phase's final debt/escape register.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-05T03:09:47Z
- **Completed:** 2026-09-05T04:03:34Z
- **Tasks:** 3 completed
- **Files modified:** 7 (1 created)

## Accomplishments

- `scripts/verify-phase4.sh`: builds the shipped binary once, verifies Phase 1, 2, 3, and 4 corpora with that one binary, requires all 12 Phase 4 control identifiers and all 3 expected escapes by exact text, and proves the three prior phases' non-regression by re-running their corpora rather than invoking their own gate scripts. Exits 0.
- `session.Phase4RequiredControls()` is the single, exported, session-layer source of truth for the 12 control identifiers; `TestPhase4RequiredControlsMatchScript` asserts it and the script's own text name the exact same set.
- A new `control:kind.exhaustive_dispatch` lane in `verifyForeignCorpus`, observable through the shipped CLI (not only a Go unit test), proving the D-04-22 six-site dispatch table's completeness for the four Phase-4-introduced operation kinds.
- All three engines (interpreter, `-O0`, `-O3`) genuinely agree — real executions, never hand-constructed documents — across all five Phase 4 path shapes named explicitly: success, second-stage typed failure, third-stage typed failure, defect, and nonlocal exit.
- Two real bugs in `lang run --engine=native` found by driving the shipped binary on out-of-corpus programs and fixed: (1) a Match-shaped multi-arm function whose arms produce different terminal-outcome kinds shared one wrong `Expect` value across all inputs; (2) any single-input program whose real outcome is a defect (including the already-committed `nonlocal_exit_probe.lang`) hit the same default-`Expect` bug. Both now resolve `Expect` per input from the interpreter's own verdict.
- `TestForeignDigestMismatchRefused`: a manifest whose `ForeignDigest` doesn't match its sidecar is refused.
- `TestShippedBinaryExercisesEveryPhase4Behavior`: every Phase 4 behavior demonstrated on a freshly built `./cmd/lang` against 8 hand-written, out-of-corpus programs (none under `testdata/`), with subcommands/exit codes/diagnostic codes recorded (see below, verbatim).
- `.planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md` finalized with D-04-32 (the Phase 5 roadmap gap: NAT-03's false-alias-fact mutation has no subject until Phase 5 first adds a proven alias-fact emission path).
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` finalized: every sign-off checkbox ticked from evidence this plan and its predecessors produced; `status: validated`, `nyquist_compliant: true`.
- SEM-03, RES-01, and FFI-01 closed in `REQUIREMENTS.md` (RES-01 with its documented nonlocal-exit-path partiality, never overstated; FFI-01 as inspectability/detectability, never as a proof of foreign conformance).

## Task Commits

Each task was committed atomically:

1. **Task 1: The Phase 4 gate, its required-control set, and its contract test (D-04-21)** - `4bafc34` (feat)
2. **Task 2: Three-engine agreement across every Phase 4 path (ROADMAP SC4)** - `4560d79` (feat)
3. **Task 3: Drive the shipped binary out of corpus and close the register (D-04-21 / D-11)** - `d69c3e2` (fix)

**Plan metadata:** commit pending (this SUMMARY + STATE/ROADMAP/REQUIREMENTS)

## Files Created/Modified

- `scripts/verify-phase4.sh` - the bounded Phase 4 gate script, a peer of `scripts/verify-phase3.sh`
- `internal/compiler/session/session.go` - `Phase4RequiredControls`, the `control:kind.exhaustive_dispatch` lane, `Phase4ThreeEngineDifferential`/`Phase4CompareThreeEngines`, `runNativeInputs` (the `Expect`-per-input fix), the three new expected-escape constants
- `internal/compiler/session/session_test.go` - `TestVerifyPhase4ControlsAndWork`, `TestVerifyPhase4CLI`, `TestPhase4VerifierScriptContract`, `TestPhase4RequiredControlsMatchScript`, `TestExpectedEscapesAreVisibleNotSolved`, `TestPhase4CorpusThreeEngineAgreement`, `TestPhase4DifferentialNamesFirstDisagreement`, `TestNoCoverageClaimedForNamedResiduals`, `TestPhase4ReachabilityRecordIsComplete`
- `internal/compiler/interp/interp.go` - `RunLinearBlockDirect`, a narrow test-only entry point for executing a specific block directly
- `internal/compiler/native/native_test.go` - `TestShippedBinaryExercisesEveryPhase4Behavior` and its 8 out-of-corpus fixtures
- `internal/compiler/evidence/evidence_test.go` - `TestForeignDigestMismatchRefused`
- `.planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md` - finalized, D-04-32 added
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` - finalized, all checkboxes ticked

## Decisions Made

See `key-decisions` in frontmatter. The most consequential: the second-/third-stage typed-failure differential is proven through **genuine execution** of the real, unmutated program's own compilable-but-dead `if (!result.ok)` branch (made live by a throwaway, non-frozen foreign test-double object), paired with a narrow interpreter entry point that drives the same block directly — not a hand-constructed execution document. An initial structural-mutation approach (retargeting an `OpForeignCall`'s ok edge to a Fail-terminated block) was tried and reverted after `corevalidate`'s own `core.fail_reached_without_err_edge` control correctly refused it — confirming the mutation was fighting an intentional, independent defense-in-depth invariant, not a gap.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `lang run --engine=native` rejected a Match-shaped multi-arm function whose arms disagree on terminal-outcome kind**
- **Found during:** Task 3, writing `TestShippedBinaryExercisesEveryPhase4Behavior`'s "Defect terminator" case
- **Issue:** `session.RunNative` called `runner.Run(ctx, cSource, opt, inputs)` once with ALL inputs sharing one `native.Runner.Expect` (defaulting to `ExpectValue`). For `defect_terminal.lang` (an ordinary-returning "Go" arm and a defect-terminating "Halt" arm), the "Halt" input's real SIGABRT was rejected as `native.run_signaled` (operational failure, exit 3) instead of decoding as a defect. Confirmed via `lang --json run --engine=native testdata/phase4/defect_terminal.lang` before the fix.
- **Fix:** Added `session.runNativeInputs`, which for a concrete `native.Runner` runs each input separately with `Expect` derived from that same input's own already-computed interpreter verdict (`expectForOutcomeKind`).
- **Files modified:** `internal/compiler/session/session.go`
- **Verification:** `lang --json run --engine=native testdata/phase4/defect_terminal.lang` now exits 0 and decodes both arms; full suite (`go test ./...`, `go test -race ./...`) unaffected.
- **Committed in:** `d69c3e2` (Task 3 commit)

**2. [Rule 1 - Bug] `lang run --engine=native` rejected the already-committed `nonlocal_exit_probe.lang` fixture**
- **Found during:** Task 3, writing the "Nonlocal-exit probe" out-of-corpus case
- **Issue:** Same root cause as #1, but for a SINGLE-input program whose real outcome is `defect` — the default `Expect` (`ExpectValue`) rejected the genuine SIGABRT. Confirmed via `lang --json run --engine=native testdata/phase4/nonlocal_exit_probe.lang` before the fix (`native.run_signaled`, exit 3) — this fixture has been in the committed corpus since plan 04-05 and had never been driven through this exact CLI command before.
- **Fix:** The same `runNativeInputs` fix in #1 covers this case (its per-input derivation applies regardless of input count).
- **Files modified:** `internal/compiler/session/session.go` (same fix as #1)
- **Verification:** `lang --json run --engine=native testdata/phase4/nonlocal_exit_probe.lang` now exits 0.
- **Committed in:** `d69c3e2` (Task 3 commit)

### Explored and reverted (not a deviation, recorded for audit)

An initial approach to the second-/third-stage typed-failure differential added a `cgen.go` case letting an `OpForeignCall`'s ok edge continue into an `OpFail`-terminated block, paired with a `session.ForceForeignCallErrEdge` structural mutation. `corevalidate`'s own `core.fail_reached_without_err_edge` control (D-04-09, defense-in-depth) correctly refused every construction of this shape — confirming it was fighting an intentional invariant, not exploiting a gap. Both were reverted (`cgen.go` is byte-identical to its state before this plan); the final approach (genuine execution via a test-double foreign object plus `interp.RunLinearBlockDirect`) is described in Decisions Made.

---

**Total deviations:** 2 auto-fixed (both Rule 1, both discovered by exactly the out-of-corpus/shipped-binary exercise D-04-21 mandates).
**Impact on plan:** Both fixes are necessary corrections to a real, reachable shipped-binary command; neither changes any currently-passing test's behavior (verified: `go test ./...` and `go test -race ./...` both green before and after). No scope creep — no new language feature, no new control, no touched frozen artifact.

## Issues Encountered

None beyond the explored-and-reverted approach above, which is not a defect — see "Explored and reverted."

## Out-of-Corpus Shipped-Binary Exercise (D-04-21, verbatim record)

Built once via `go build -o <tmp>/lang ./cmd/lang` (through `testsupport.BuildCLI`); each behavior below is a hand-written program written fresh to a temp directory (never under `testdata/`), driven through the actual subcommands, with exit code and (where applicable) diagnostic code recorded verbatim from the real run (`internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior`, `go test ./internal/compiler/native/... -run TestShippedBinaryExercisesEveryPhase4Behavior -v`):

| Behavior | Subcommand | Exit | Diagnostic code |
|---|---|---|---|
| Fallible foreign call through `try` | `format --check {path}` | 0 | — |
| | `check {path}` | 0 | — |
| | `run --engine=interpreter {path}` | 0 | — |
| | `run --engine=native {path}` | 0 | — |
| Three-stage acquisition with reverse-order release | `check {path}` | 0 | — |
| | `run --engine=interpreter {path}` | 0 | — |
| | `run --engine=native {path}` | 0 | — |
| `discard ... because` consumer | `format --check {path}` | 0 | — |
| | `check {path}` | 0 | — |
| | `run --engine=interpreter {path}` | 0 | — |
| | `run --engine=native {path}` | 0 | — |
| Defect terminator | `check {path}` | 0 | — |
| | `run --engine=interpreter {path}` | 0 | — |
| | `run --engine=native {path}` | 0 | — |
| Nonlocal-exit probe | `check {path}` | 0 | — |
| | `run --engine=interpreter {path}` | 0 | — |
| | `run --engine=native {path}` | 0 | — |
| Foreign-origin refusal | `--json check {path}` | 2 | `core.foreign_origin_omitted` |
| Policy-less foreign declaration refusal | `--json check {path}` | 2 | `foreign.unwind_policy_undeclared` |
| Lang-targeted call refusal | `--json check {path}` | 2 | `core.call_target_not_foreign` |

Every exercised source path was asserted (in-test) to NOT contain the substring `testdata` — none of the 8 programs lives in any corpus directory.

## Full Verification Results

```
$ sh scripts/verify-phase4.sh
[exit 0; all four corpora verified with one binary; all 12 Phase 4 control identifiers and 3 expected escapes present in the Phase 4 result; Phase 2's control:backend.runtime_causality and all 11 Phase 3 controls present in their own results]

$ go test ./...          # pass, all packages
$ go test -race ./...    # pass, all packages
$ go vet ./...           # clean
$ git diff <phase-start>..HEAD -- testdata/phase1 testdata/phase2 testdata/phase3   # empty
$ git diff <phase-start>..HEAD -- scripts/verify-phase1.sh scripts/verify-phase2.sh scripts/verify-phase3.sh native/lang_foreign_resource.c native/lang_foreign_resource_private.h native/lang_foreign_nonlocal.c go.mod   # empty
```

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 4 is complete: `sh scripts/verify-phase4.sh` exits 0 and asserts every control and escape this phase introduced. `SEM-03`, `RES-01`, and `FFI-01` are closed in `REQUIREMENTS.md`.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-DEBT.md` carries the phase's open debt into Phase 5's own planning: D-04-26 (retiring `discoverLoanLastUses`), D-04-30 (payload-carrying alternatives, deferred to M002/Phase 6), D-04-31 (four accepted residual limitations, permanent), and **D-04-32 (new): Phase 5's own planning must sequence a proven alias-fact emission path before or alongside NAT-03's false-no-alias-facts hostile mutation, which otherwise has no subject to attack.** Whoever plans Phase 5 should read D-04-32 and ROADMAP.md §Phase 5 together before starting.
- No blockers.

## Self-Check: PASSED

- `[ -f scripts/verify-phase4.sh ]` → FOUND
- `[ -f internal/compiler/interp/interp.go ]` → FOUND (RunLinearBlockDirect present)
- `git log --oneline --all --grep="04-07"` → 3 commits found (4bafc34, 4560d79, d69c3e2)
- Re-ran plan-level `<verification>`: all automated checks pass (see "Full Verification Results" above)
- Re-ran all task `<acceptance_criteria>`: pass

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*
