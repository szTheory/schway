---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 03
subsystem: cgen
tags: [cgen, callgraph, session, native, multi-function, c17]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-01's Q-01/Q-02 spike verdicts and PHASE-11-DEBT.md; 11-02's NAT-07 LTO control and ratified roadmap amendments (D-11-09/10 zero-attribute, D-11-20/21/22)"
provides:
  - "callgraph.EntryFunction: the single program-entry resolver consumed by cgen's whole-program assembler and every session run site"
  - "cgen.emitProgram/cgen.emitCall: the whole-program C17 TU assembler and its sole Lang-to-Lang call writer, in new file cgen_program.go"
  - "session.go's three run sites and interpreterInputs resolved via EntryFunction instead of Functions[0]"
  - "testdata/phase11/ corpus: multi_function_entry_basic (the proven tracer), multi_function_forward_callee, multi_function_unreachable, multi_function_zero_call"
affects: [11-04, 11-05, 11-06, 11-07, 11-08, 11-09]

# Actuals (#2632)
actuals:
  tokens: 24000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "callgraph.EntryFunction's closure-size tie-break: when more than one in-degree-zero candidate exists, prefer the one whose transitive reachable closure is strictly larger than every other candidate's — a real structural fact (never Functions[0]-style guessing) that resolves a declared-but-uncalled function alongside a genuine entry, while a truly symmetric tie (including the all-isolated case) still refuses."
    - "Two-tier cNames allocation: one global cNames allocates every function's own C name in callgraph.Order's order; each function then gets a FRESH per-function cNames seeded with the reserved list plus every globally allocated name, so locals never carry a cross-function ordinal suffix."
    - "emitCall (cgen_program.go) is the one writer of a Lang-to-Lang C call expression; emitLinear's and emitBranchOperations' own OpCall arms (single-function paths, provably unreachable in production) route through it too, rather than forking a second copy of the call-emission logic."

key-files:
  created:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/cgen/cgen_program_test.go
    - internal/compiler/callgraph/callgraph_entry_test.go
    - testdata/phase11/multi_function_entry_basic.lang
    - testdata/phase11/multi_function_forward_callee.lang
    - testdata/phase11/multi_function_unreachable.lang
    - testdata/phase11/multi_function_zero_call.lang
  modified:
    - internal/compiler/callgraph/callgraph.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_names_test.go
    - internal/compiler/core/core.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session.go

key-decisions:
  - "EntryFunction's algorithm is in-degree-zero-root resolution with a closure-size tie-break, not pure in-degree alone — pure in-degree cannot satisfy the plan's own top-level must_have that a declared-but-uncalled function 'does not change the entry resolution' while also satisfying D-11-05's 'never guess' prohibition on a truly ambiguous program. The tie-break is a real structural fact (reachable-set size), never an arbitrary pick."
  - "multi_function_zero_call.lang's own test (TestEmitProgramZeroCallEdges) asserts a clean core.EntryAmbiguous refusal, not successful emission — the plan's literal 'assert emission succeeds' wording is unsatisfiable for a genuinely two-function, zero-call program under any non-guessing resolver (both candidates have an identically empty closure — an unbreakable tie). The refusal is the non-degenerate behavior; see Deviations."
  - "native.go's execution-document validator's single-FunctionID-across-all-events and function.returned-must-be-last checks were pre-existing single-function assumptions never previously exercised by a real multi-function native binary; both were widened (Rule 3, blocking-issue auto-fix) rather than routed around, since they are the exact validation Phase 11's own goal depends on."
  - "A declared-but-never-called function is kept alive against -Werror -Wunused-function by taking its own address in main ((void)LANG_ORPHAN;) — a harmless, side-effect-free reference, never a call — rather than suppressing the warning or excluding dead functions from emission."

patterns-established:
  - "Whole-program TU assembly owns includes/typedefs/support blocks/prototypes/main; per-function bodies write statements and events only, never their own event-support declarations."

requirements-completed: [NAT-04]

coverage:
  - id: D1
    description: "A two-function Lang program (multi_function_entry_basic.lang) compiles to C17, links, runs, and produces a lang.execution/1 document equal to interp.Run's, driven through session's own run path"
    requirement: "NAT-04"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgramEndToEndAgreesWithInterpreterAtO0"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgram"
        status: pass
    human_judgment: false
  - id: D2
    description: "callgraph.EntryFunction resolves the single program entry, refusing zero/many in-degree-zero roots by a named, byte-stable code, and agrees with the single ast.Export on 39 real pre-Phase-11 fixtures"
    verification:
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_entry_test.go#TestEntryFunctionAgreesWithSingleExport"
        status: pass
      - kind: unit
        ref: "internal/compiler/callgraph/callgraph_entry_test.go#TestEntryFunctionRefusesManyRoots"
        status: pass
    human_judgment: false
  - id: D3
    description: "Two-tier deterministic name allocation across functions: no cross-function ordinal suffix, no local/global collision, byte-identical re-emission"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_names_test.go#TestMultiFunctionNameAllocation"
        status: pass
    human_judgment: false
  - id: D4
    description: "Structural edge cases (zero-call, forward-reference, upstream recursion refusal, single-function byte-freeze) each pinned by a dedicated test"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgramZeroCallEdges"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgramForwardDefinedCallee"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgramRefusesRecursiveProgramsUpstream"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_program_test.go#TestEmitProgramSingleFunctionBytesUnchanged"
        status: pass
    human_judgment: false
  - id: D5
    description: "Emitted C for the tracer fixture is readable on manual inspection: one C function per Lang function, provenance comments present, no macro soup"
    verification: []
    human_judgment: true
    rationale: "11-VALIDATION.md names this a Manual-Only Verification; the emitted C was inspected during development (LANG_MAIN/LANG_IDENTITY, /* call: ... */ and /* returned place: ... */ provenance comments) but a human should independently confirm."

duration: 70min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 3: Multi-Function Native Emission Tracer Summary

**A two-function Lang program now checks, emits readable C17, compiles under `-O0`/`-O3`, links, runs, and produces an execution document byte-identical to the interpreter's — driven entirely through `session`'s own run path, with `callgraph.EntryFunction` as the single program-entry resolver both engines consult.**

## Performance

- **Duration:** ~70 min
- **Tasks:** 3 completed
- **Files modified:** 6 modified, 7 created

## Accomplishments

- **`callgraph.EntryFunction`** resolves a program's unique entry from its own call graph: in-degree-zero root analysis, with a closure-size tie-break (the candidate whose transitive reachable set is strictly larger wins) that lets a declared-but-uncalled function coexist with a genuine entry without either guessing or spuriously refusing. Zero or genuinely tied roots refuse with the new `core.EntryAmbiguous` code, sorted and byte-stable.
- **`cgen.emitProgram`** (new `cgen_program.go`) is the whole-program C17 assembler: one shared event buffer sized to the whole program, two-tier deterministic name allocation (D-11-08), every prototype before every definition (making forward references legal C17), and `cgen.emitCall` as the sole writer of a Lang-to-Lang call anywhere in the package. `Emit`/`EmitNative`'s old `len(Functions) != 1` hard error now dispatches to it.
- **`emitLinear`'s and `emitBranchOperations`' own `core.OpCall` arms** (single-function paths, provably unreachable in production since a one-function program can never legally contain a call) now route through the same `emitCall` helper instead of returning the old forward-hygiene error string — one writer, not two copies of the logic.
- **`session.go`'s three run sites** (`Phase4CheckedProgram`, `RunInterpreter`, `RunNative`) and `interpreterInputs` resolve the entry via `callgraph.EntryFunction`, with zero remaining `Functions[0]` indexing in any of the three — pinned by a source-scan test.
- **Two pre-existing single-function assumptions in `native.go`'s execution-document validator** were discovered and widened (Rule 3): every event previously had to share one `FunctionID`, and a `function.returned` event previously had to be the document's last event. Both assumptions were never exercised until a real multi-function native binary existed to violate them.
- **Four `testdata/phase11/` fixtures**, each proven end to end (checked, and — except the deliberately-ambiguous one — compiled/linked/run at both `-O0` and `-O3`, agreeing with the interpreter).

## Task Commits

1. **Task 1: End-to-end tracer** — `c8efe4e` (feat) — `EntryFunction`, `emitProgram`/`emitCall`, `Emit`/`EmitNative` dispatch, `session.go` run-site rewiring, `native.go` validator widening, all four `testdata/phase11/` fixtures, and `cgen_program_test.go`'s full test set.
2. **Task 2: EntryFunction's refusals and the no-`Functions[0]` invariant** — `9d82220` (test) — `callgraph_entry_test.go`.
3. **Task 3: Two-tier name allocation** — `db50d71` (test) — `cgen_names_test.go`'s `TestMultiFunctionNameAllocation`.

## Files Created/Modified

- `internal/compiler/cgen/cgen_program.go` — `emitProgram`, `emitCall`, `emitCallLookup`, `emitProgramFunction`
- `internal/compiler/cgen/cgen_program_test.go` — `TestEmitProgram`, `TestEmitProgramEndToEndAgreesWithInterpreterAtO0`, `TestEmitProgramZeroCallEdges`, `TestEmitProgramForwardDefinedCallee`, `TestEmitProgramRefusesRecursiveProgramsUpstream`, `TestEmitProgramSingleFunctionBytesUnchanged`
- `internal/compiler/callgraph/callgraph.go` — `EntryFunction`, `entryAmbiguousError`, `reachableClosureSize`, updated package doc
- `internal/compiler/callgraph/callgraph_entry_test.go` — the eight Task 2 tests
- `internal/compiler/cgen/cgen.go` — `Emit`/`EmitNative` dispatch fork; `emitLinear`/`emitBranchOperations` `OpCall` arms
- `internal/compiler/cgen/cgen_names_test.go` — `TestMultiFunctionNameAllocation`
- `internal/compiler/core/core.go` — `core.EntryAmbiguous`
- `internal/compiler/native/native.go` — `validateExecution` widened for multi-FunctionID, non-terminal `function.returned` events
- `internal/compiler/session/session.go` — `Phase4CheckedProgram`, `RunInterpreter`, `RunNative`, `interpreterInputs`
- `testdata/phase11/multi_function_entry_basic.lang`, `multi_function_forward_callee.lang`, `multi_function_unreachable.lang`, `multi_function_zero_call.lang`

## Decisions Made

See `key-decisions` above. The two load-bearing ones: (1) `EntryFunction`'s closure-size tie-break, needed to reconcile the plan's own internally-conflicting must_haves (a declared-but-uncalled function must not block resolution, yet the resolver must never guess), and (2) `multi_function_zero_call.lang`'s test asserting refusal rather than success, since a genuinely symmetric two-function zero-call program has no non-guessing resolution.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] `native.go`'s execution-document validator rejected legitimate multi-function documents**
- **Found during:** Task 1, first end-to-end run of the tracer fixture through `native.Runner.Run`
- **Issue:** `validateExecution` required every event in a document to share one `FunctionID` (the first event's) and required any `function.returned` event to be the document's last event. Both are true for every pre-Phase-11 single-function execution by construction, but a genuine multi-function call produces one `function.returned` event per popped frame (interp's own `terminalOutcome`/`runFrameStack` behavior since Phase 10, D-10-32) with per-frame `FunctionID`s — exactly what the tracer fixture's real native binary now emits.
- **Fix:** Widened both checks: events may carry distinct `FunctionID`s, and only the true `TargetPlace`-empty requirement remains on `function.returned`; it need no longer be last (the outermost frame's own event is still always the actual last one written, by construction of both `interp` and `emitProgram`'s own `main`).
- **Files modified:** `internal/compiler/native/native.go`
- **Verification:** `go test ./internal/compiler/native/...` green (including the full existing `TestValidateExecutionStillRejectsOldGrounds`-style suite); the tracer's own `TestEmitProgramEndToEndAgreesWithInterpreterAtO0` exercises the widened path directly.
- **Committed in:** `c8efe4e`

**2. [Rule 1 - Bug] `-Werror -Wunused-function` on a declared-but-never-called function**
- **Found during:** Task 1, compiling `multi_function_unreachable.lang`'s generated C
- **Issue:** D-11-05's own unreachable-function contract requires the orphan function to still be emitted as a real C definition, but a `static` function nothing calls fails this project's `-Werror` build.
- **Fix:** `main` now takes the address of every emitted function (`(void)LANG_NAME;`) — a harmless, side-effect-free reference, never a call.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** `TestEmitProgramEndToEndAgreesWithInterpreterAtO0`'s sibling coverage on `multi_function_unreachable.lang` compiles and links.
- **Committed in:** `c8efe4e`

### Architectural Note (Rule 4-adjacent — flagged, not silently resolved)

**3. `EntryFunction`'s algorithm and `multi_function_zero_call.lang`'s expected outcome diverge from the plan's literal text**

The plan's own must_haves contain a real internal tension: (a) "a function declared but never called ... does not change the entry resolution" (requires success on `multi_function_unreachable.lang`-shaped programs), and (b) "MUST NOT guess an entry function when zero or many in-degree-zero roots exist" (requires refusal on a genuinely ambiguous program). A pure in-degree-zero-root check cannot satisfy (a); a version that always resolves any tie would violate (b). The implemented algorithm — prefer the in-degree-zero candidate with a *strictly larger* transitive reachable closure, refuse only when that comparison is itself tied — satisfies both: it resolves `multi_function_unreachable.lang` (closure `{callee}` beats the orphan's empty closure) while still refusing a truly symmetric two-function zero-call program (both closures empty, an unbreakable tie).

Task 3's own literal acceptance text for `TestEmitProgramZeroCallEdges` ("two functions, ZERO OpCall operations... assert emission succeeds") is consequently unsatisfiable by any non-guessing resolver: with genuinely zero calls anywhere, every declared function has an in-degree of zero **and** an empty reachable closure, so no structural signal can distinguish "the entry" from "an uncalled orphan" without either adding a schema field (explicitly forbidden by D-11-06) or guessing (explicitly forbidden by D-11-05). The shipped test instead asserts a clean, named `core.EntryAmbiguous` refusal — never a crash, never a silently-wrong emission — which this executor reads as satisfying the task's own "must not take a degenerate path" framing honestly. **Flagged for human review**, since it changes Task 3's literal expected test outcome.

**4. `cgen.emitCall`'s call-site count is 4 (1 declaration + 3 calls: `emitLinear`, `emitBranchOperations`, `emitProgramFunction`), not the plan's literal 3 (1 declaration + 2 calls)**

The plan's D-11-04 acceptance criterion (`grep -c 'emitCall(' == 3`) implicitly requires `emitProgram`'s own per-function body loop to route through one of the two existing (`emitLinear`/`emitBranchOperations`) call sites rather than its own. Doing so without forking logic would require extracting `emitLinear`'s per-operation arm into a function shared with `emitProgram` — but `emitLinear`'s `OpReturn` arm writes a whole-TU JSON tail (single-function behavior, byte-frozen) while `emitProgram`'s own per-function `OpReturn` arm writes only an event and a bare `return` (multi-function behavior) — two genuinely different, both load-bearing behaviors that cannot share one arm without either breaking the byte-freeze or breaking multi-function correctness. This executor prioritized both `D-11-01`'s "byte-for-byte untouched" invariant (verified: `git diff` over `emitLinear`'s/`emitBranchOperations`' non-`OpCall` lines is empty, and `TestExistingEmittersAreByteIdentical`/`TestEmitProgramSingleFunctionBytesUnchanged` are green) and genuine multi-function correctness over the literal grep count. `emitCall` remains the single function that emits a Lang-to-Lang call's actual C text at every one of its three call sites — the substantive D-11-04 property — even though the mechanical count differs. **Flagged for human review.**

---

**Total deviations:** 2 auto-fixed (1 Rule 1, 1 Rule 3), 2 architectural notes flagged for human review (not silently resolved).
**Impact on plan:** None on correctness or scope — every deviation is documented with its reasoning, and the substantive goal (a two-function program running identically on both engines, with the six single-function emitters frozen) is met and independently verified (`go test ./...` green, `-O0`/`-O3` agreement, byte-freeze test passing).

## Issues Encountered

None beyond the two auto-fixes and two architectural notes above — all discovered during normal end-to-end verification, not left as open questions.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `callgraph.EntryFunction` is now the one resolver plans 11-04 through 11-09 should consult for "which function is this program" — no second convention should be introduced.
- `emitProgram`'s own scope this plan is deliberately narrow: straight-line bodies only (no `core.Match`, no `core.Block`-based control flow, no `core.ForeignContract`). Any later plan needing multi-function branch or foreign bodies will need to extend `emitProgram`, not fork it.
- The two `native.go` widenings (multi-`FunctionID` events, non-terminal `function.returned`) are now load-bearing for every subsequent multi-function native lane — worth a dedicated look during Phase 11's mid-phase gate (11-04) to confirm they don't mask a real defect on a shape this plan's corpus doesn't cover.
- Human review recommended for deviations 3 and 4 above before treating `EntryFunction`'s exact contract and `emitCall`'s call-site count as fully settled.
- No blockers to proceeding with plan 11-04.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

All created files verified present on disk (`internal/compiler/cgen/cgen_program.go`,
`cgen_program_test.go`, `internal/compiler/callgraph/callgraph_entry_test.go`, all four
`testdata/phase11/*.lang` fixtures). All three task commit hashes (`c8efe4e`, `9d82220`,
`db50d71`) verified present in `git log`. `go build ./...` clean. `go vet ./...` clean.
`go test ./...` green (re-run after all three commits). `go run ./cmd/lang --json check`
clean on all four `testdata/phase11/` fixtures. Emitted C for the tracer fixture manually
inspected: one C function per Lang function, `/* call: ... */`/`/* returned place: ... */`
provenance comments present, no macro soup.
