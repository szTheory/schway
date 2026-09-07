---
phase: 04-fallible-resources-and-c-boundary
plan: "01"
subsystem: compiler
tags: [ffi, c17, foreign-call, typed-failure, core-ir, corevalidate, interp, cgen, native]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: block/edge CFG shape (core.Block/core.Edge), checkBranch's admit-a-CFG-shaped-linear-function precedent, the six-dispatch-site convention for a new core.OperationKind
provides:
  - "core.AllOperationKinds()/TerminatorKinds() registry and the table-driven six-site exhaustive-dispatch control"
  - "core.OpForeignCall/core.OpFail, core.ForeignContract, and additive LinearOperation edge fields (OkEdgeID/ErrEdgeID/ErrTargetID)"
  - "the `foreign C { }` declaration surface and the `try <callee>(<args>)` fallible-call form"
  - "checkFallibleLinear: a fully separate check.go path lowering one fallible foreign call into a three-block, two-edge CFG"
  - "independent corevalidate re-derivation of the foreign-call shape plus D-04-16/D-04-02's two admission refusals"
  - "runLinearBlocks (interp) and emitLinearForeign (cgen): ok/err-edge-aware execution and C17 generation"
  - "native.Runner.ForeignSources: a second, separately-compiled-and-linked translation unit"
  - "native/lang_foreign_resource.c + private header: the byte-frozen foreign translation unit wrapping libc malloc/free"
affects: [04-02, 04-03, 04-04]

actuals:
  tokens: 29740
  tasks: 4
  commits: 3

tech-stack:
  added: []
  patterns:
    - "A straight-line (Match-less) linear function can now carry Blocks/Edges of its own (not only a match-arm body); runLinear/emitLinear/corevalidate's linear() dispatch on len(Blocks) > 0 to a block-walking path, byte-identical for every function that leaves Blocks empty."
    - "A completely separate check.go lowering function (checkFallibleLinear) for a new syntactic shape, rather than generalizing checkLinear/analyzeStraightLine — zero risk to the Phase 1-3 straight-line path."
    - "corevalidate re-derives an admission refusal from a materially different fact (a core.Function.Name collision scan) than check's own AST-level resolution, per D-12a."

key-files:
  created:
    - native/lang_foreign_resource.c
    - native/lang_foreign_resource_private.h
    - internal/compiler/native/foreign_resource.go
    - internal/compiler/core/core_test.go
    - testdata/phase4/foreign_acquire_one.lang
    - testdata/phase4/fallible_call_unconsumed.lang
    - testdata/phase4/foreign_unwind_undeclared.lang
    - testdata/phase4/foreign_call_target_not_foreign.lang
  modified:
    - internal/compiler/core/core.go
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go
    - internal/compiler/syntax/token.go
    - internal/compiler/syntax/syntax_test.go
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/native/native.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go

key-decisions:
  - "D-04-04 checkpoint (Task 04-01-02) resolved: edge-based-now — typed failure is a two-successor control-flow edge in core, not a storable Result value; core.DataType.Alternatives is untouched."
  - "The tracer's ok payload and the function's own parameter are both typed Byte (not a richer opaque Handle type) — a deliberate narrowing so the existing Byte scalar writer/reader in cgen/interp needs no generalization this plan; a real opaque resource-handle type is future work."
  - "The declared failure ADT's error VALUE on the err edge is always the type's first declared alternative, in both interp and cgen — a documented narrowing (no case-analysis syntax exists yet to pick a specific cause), not a hidden inconsistency."
  - "checkFallibleLinear supports exactly one shape this plan: a straight-line function whose sole binding is the try-call, immediately returned. Anything richer (ordinary bindings before/after, multiple foreign calls) is refused with check.foreign_call_shape_unsupported rather than mishandled."

patterns-established:
  - "Six-site exhaustive dispatch for a new core.OperationKind: check (emission), corevalidate (both replay switches), interp (both dispatch switches), cgen (both dispatch switches), pathoracle/originvalidate (terminator-membership tests, unaffected by a non-terminal or not-yet-terminator-recognized kind)."
  - "A frozen foreign translation unit lives at repo-root native/, resolved at both test time (testsupport.ProjectPath) and native-run time (native.ForeignResourceSourcePath, a runtime.Caller-based analog) — never a CWD-relative assumption."

requirements-completed: [SEM-03, FFI-01]

coverage:
  - id: D1
    description: "core.AllOperationKinds()/TerminatorKinds() registry; six-site exhaustive-dispatch control green on the five pre-Phase-4 kinds"
    requirement: SEM-03
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsRegistered"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestAllOperationKindsHandledAtEverySite"
        status: pass
    human_judgment: false
  - id: D2
    description: "Pre-Phase-4 core bytes and evidence manifest IDs pinned byte-for-byte against a phase-start golden"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseCoreBytesUnchanged"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_test.go#TestPreviousPhaseManifestIDsUnchanged"
        status: pass
    human_judgment: false
  - id: D3
    description: "native.validateExecution carries an explicit expected-terminal-outcome axis; every original Phase 1-3 rejection ground still rejects"
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestValidateExecutionStillRejectsOldGrounds"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_test.go#TestValidateExecutionTypedFailureDiscriminates"
        status: pass
    human_judgment: false
  - id: D4
    description: "`foreign C { }` block and `try <callee>(<args>)` parse losslessly, format to a fixed point, and reparse to the same semantic token projection; a bare fallible call is refused at parse time"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_test.go#TestForeignCallRoundTrips"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/syntax_test.go#TestFallibleCallUnconsumedRejected"
        status: pass
    human_judgment: false
  - id: D5
    description: "One Lang program declaring a foreign C symbol, calling it fallibly through try, checks into a three-block/two-edge CFG with a populated ForeignContract"
    requirement: SEM-03
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestForeignCallLowersToOkAndErrEdges"
        status: pass
    human_judgment: false
  - id: D6
    description: "Interpreter and Clang-built native code (-O0 and -O3) agree on terminal outcome, ordered events, and live-resource state for the tracer fixture"
    requirement: FFI-01
    verification:
      - kind: e2e
        ref: "internal/compiler/session/session_test.go#TestForeignCallInterpreterNative"
        status: pass
    human_judgment: false
  - id: D7
    description: "Two inadmissible foreign shapes (missing unwind/nonlocal_exit policy; call target resolving to a Lang function) are refused independently by check and corevalidate, and visible to the session gate as required negative controls"
    requirement: FFI-01
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestUnwindPolicyUndeclaredRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestCallTargetNotForeignRejected"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestForeignRefusalsAreIndependentlyDerived"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase4ForeignControls"
        status: pass
    human_judgment: false
  - id: D8
    description: "The shipped ./cmd/lang binary runs the tracer fixture and one hand-written, non-corpus program cleanly through format --check, check, run --engine=interpreter, and run --engine=native"
    requirement: FFI-01
    verification:
      - kind: integration
        ref: "internal/compiler/native/native_test.go#TestShippedBinaryFourSubcommandCorpusMatrix"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/native_test.go#TestShippedBinaryExercisesEveryPhase4Behavior"
        status: pass
      - kind: other
        ref: ".github/workflows/ci.yml (checks + phase-4 gate, ubuntu-latest and macos-latest)"
        status: pass
    human_judgment: false
    rationale: "Originally driven manually against the built ./cmd/lang binary during this session (per D-04-21's 'drive the shipped binary' method), with the four exit codes for both programs recorded verbatim in this summary. That manual act was retired in favour of two in-repo assertions: the corpus matrix test drives all four subcommands over EVERY testdata/phase4 fixture against a freshly built binary and fails if a fixture has no recorded outcome, and the out-of-corpus test does the same for the hand-written program. Both run in CI on Linux and macOS. The verbatim table below is kept as the historical evidence trail, not as the live control."

duration: ~5h (single continuous session)
completed: 2026-09-04
status: complete
---

# Phase 4 Plan 1: Fallible Resources and C Boundary — Tracer Summary

**One Lang program declares a foreign C symbol, calls it fallibly through `try`, and returns its ok value — checked, independently validated, interpreted, lowered to C17, and linked against a byte-frozen foreign translation unit, agreeing bit-for-bit between the interpreter and Clang-built native code at -O0 and -O3.**

## Performance

- **Duration:** ~5h (single continuous session, no wall-clock breaks tracked)
- **Tasks:** 4/4 completed (including one checkpoint resolution)
- **Files modified:** 25 (8 created, 17 modified)

## Accomplishments

- `core.AllOperationKinds()`/`core.TerminatorKinds()` registry, backing a real table-driven `control:kind.exhaustive_dispatch` equivalent (`TestAllOperationKindsHandledAtEverySite`) that drives every existing corpus fixture plus the new tracer through check/corevalidate/interp/cgen/pathoracle/originvalidate.
- Every Phase 1/2/3 fixture's core JSON bytes and evidence manifest ID pinned against a phase-start golden, proven unmoved by this plan's additive changes.
- `native.validateExecution` widened with an explicit `value | typed_failure` expected-terminal-outcome axis (the closed `defect` member is declared but not yet constructible), with a regression test proving every original Phase 1-3 rejection ground still rejects.
- A `foreign C { }` declaration surface (lexer/parser/formatter/ast) and the `try <callee>(<args>)` fallible-call form; a bare call with no `try` is refused at parse time with a span-bearing diagnostic, so the core IR never has to encode a fallible operation with no failure successor.
- Two new core operation kinds, `OpForeignCall` and `OpFail`, dispatched at all six required sites, plus `core.ForeignContract` and three additive `LinearOperation` fields (`OkEdgeID`/`ErrEdgeID`/`ErrTargetID`). `core.DataType.Alternatives` is untouched.
- `check.go`'s `checkFallibleLinear` — a fully separate lowering path from `checkLinear`/`analyzeStraightLine` — turns the tracer's one supported shape into a three-block, two-edge CFG.
- `corevalidate` independently authorizes the new shape in both replay switches and independently re-derives both admission refusals (missing unwind/nonlocal_exit policy; a call target colliding with a declared Lang function name).
- `interp`'s new `runLinearBlocks` and `cgen`'s new `emitLinearForeign` give both engines a genuine ok/err-edge-aware execution path; the interpreter always simulates success (a documented discretionary stub), while the generated C makes a real extern call to the byte-frozen foreign translation unit and forks on the real runtime result.
- `native/lang_foreign_resource.c` + private header: the hand-written, byte-frozen foreign translation unit wrapping real libc `malloc`/`free`, compiled and linked as its own separate `os/exec` invocation (`native.Runner.ForeignSources`).
- Proven end-to-end through the real toolchain: `session.RunNative` on `testdata/phase4/foreign_acquire_one.lang` succeeds at both `-O0` and `-O3`, agreeing with the interpreter on outcome, events, and live resources.

## Task Commits

Each task was committed atomically:

1. **Task 04-01-01: Pin byte identity, add the kind registry, clear the native terminal-outcome landmine** — `d529905` (feat)
2. **Task 04-01-02: Confirm the one-way door (checkpoint:decision)** — resolved `edge-based-now`; no code commit (decision recorded here)
3. **Task 04-01-03: End-to-end fallible foreign call across the frozen C boundary** — `62a443d` (feat)
4. **Task 04-01-04: Refuse missing foreign policy and non-foreign call targets** — `7df0211` (feat)

## Checkpoint Resolution (Task 04-01-02)

**Decision:** D-04-04 — typed failure representation.
**Option chosen:** `edge-based-now` — typed failure lands as a two-successor control-flow edge in core; the error payload rides the `err` edge as an ordinary existing nullary ADT, the ok payload as an ordinary place on the ok successor block. No `Result` type, no generics, no payload-carrying alternatives; `core.DataType.Alternatives` is unchanged. The additive escape hatch (`alternative_details []Alternative omitempty`, keyed by name) stays deferred debt.

This choice is exercised structurally throughout Task 04-01-03: `core.OpForeignCall`'s `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` fields, `core.DataType.Alternatives` remaining `[]string`, and the err edge carrying a place of the declared failure ADT's type rather than a constructed error value.

## Files Created/Modified

- `internal/compiler/core/core.go` — `AllOperationKinds()`/`TerminatorKinds()`, `OpForeignCall`/`OpFail`, `core.ForeignContract`, additive `LinearOperation` edge fields
- `internal/compiler/core/core_test.go` — pinning goldens, registry tests, six-site exhaustive-dispatch control
- `internal/compiler/native/native.go` — `TerminalOutcome` axis, `Runner.ForeignSources` (separate compile+link)
- `internal/compiler/native/foreign_resource.go` — `ForeignResourceSourcePath()` (build-time-relative frozen-TU path resolution)
- `internal/compiler/ast/ast.go` — `ForeignBlock`/`ForeignSymbol`/`ForeignPolicy`, `RHS.Callee`/`RHS.Arguments`
- `internal/compiler/syntax/{lexer,parser,format,token}.go` — `foreign`/`try` keywords, string literals, `foreign C {}` grammar, `try` call grammar, bare-call rejection, format support
- `internal/compiler/check/check.go` — `checkFallibleLinear`, `collectForeignSymbols`, `missingForeignPolicyDiagnostic`, the caps
- `internal/compiler/corevalidate/corevalidate.go` — `OpForeignCall`/`OpFail` cases in both replay switches, independent policy/call-target refusals
- `internal/compiler/interp/interp.go` — `runLinearBlocks`, `OpForeignCall`/`OpFail` cases in both switches
- `internal/compiler/cgen/cgen.go` — `emitLinearForeign`, `OpForeignCall`/`OpFail` handling in both switches
- `internal/compiler/session/session.go` — `ForeignSources` wiring in `RunNative`, `verifyForeignCorpus` gate dispatch
- `native/lang_foreign_resource.c` + `native/lang_foreign_resource_private.h` — the frozen foreign translation unit
- `testdata/phase4/*.lang` — four new fixtures (tracer, bare-call rejection, missing-policy rejection, call-target rejection)

## Decisions Made

- D-04-04 checkpoint resolved `edge-based-now` (see above).
- The tracer's ok payload and the function's own parameter are both typed `Byte`, not a richer opaque `Handle` type — deliberate, so cgen/interp's existing Byte scalar writer needs no generalization this plan. A real opaque resource-handle representation is explicitly future work (plan 02/03).
- The declared failure ADT's error value on the err edge is always the type's first declared alternative in both engines — a documented narrowing (no case-analysis syntax exists yet to pick a specific failure cause), not a hidden inconsistency between interpreter and native.
- `checkFallibleLinear` supports exactly one shape this plan (a straight-line function whose sole binding is the try-call, immediately returned); anything richer is refused with `check.foreign_call_shape_unsupported` rather than silently mishandled.

## Deviations from Plan

None — plan executed as written, within the narrowing choices Claude's Discretion explicitly permitted (concrete `OperationKind` spellings, `foreign`/`try` surface syntax, `core.ForeignContract` field layout, the error-value-selection narrowing).

## Shipped-Binary Proof (D-04-21)

Built `./cmd/lang` via `go build -o /tmp/lang_bin ./cmd/lang`. Ran all four subcommands against the corpus fixture and one hand-written, non-corpus program:

**Corpus fixture — `testdata/phase4/foreign_acquire_one.lang`:**
| Subcommand | Exit code |
|---|---|
| `format --check` | 0 |
| `check` | 0 |
| `run --engine=interpreter` | 0 |
| `run --engine=native` | 0 |

**Non-corpus hand-written program** (`module handwritten.probe`, a `foreign C {}` block calling `lang_res_open` through `try`, never checked into any corpus):
| Subcommand | Exit code |
|---|---|
| `format --check` | 0 |
| `check` | 0 |
| `run --engine=interpreter` | 0 |
| `run --engine=native` | 0 |

`run --engine=native` for both programs compiled and linked the generated C against the byte-frozen `native/lang_foreign_resource.c` via two separate `clang` invocations (per D-04-10) at both `-O0` and `-O3`, and produced execution documents identical to the interpreter's.

## Verification Performed

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...` — pass
- `go test -race ./...` — pass
- `go vet ./...` — clean
- `git diff <phase-start>..HEAD -- testdata/phase1 testdata/phase2 testdata/phase3` — empty
- `go.mod`/`go.sum` — no new requirement
- `sh scripts/verify-phase3.sh` — exits 0; `scripts/verify-phase3.sh` byte-identical to phase-start
- No file under `internal/compiler/` references `native/lang_foreign_resource_private.h` (grepped; only `native/lang_foreign_resource.c` includes it)

## Issues Encountered

None blocking. One formatter subtlety was found and fixed during development: the pre-existing "a return-type identifier immediately follows a borrow-origin annotation's closing paren" spacing rule in `format.go` also fired (incorrectly) after a `try`-call's own closing paren once that paren was given its own newline; guarded on whether the line was already open before the identifier, so the borrow-origin case is unaffected and the try-call case no longer emits a stray leading space. This was investigation, not a plan deviation.

## Known Stubs

None that block this plan's own success criteria. Documented narrowings (not stubs): the ok payload/parameter both being `Byte` rather than an opaque `Handle`, and the always-first-alternative failure-ADT value — both are explicitly future-plan scope per the plan's own `<flagged_assumptions>` and `<source_coverage_audit>` (FFI-01 marked PARTIAL, full contract lands in 04-03; RES-01/release ordering lands in 04-02).

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness

- The six-site exhaustive-dispatch registry, the `foreign C {}` surface, `core.ForeignContract`, and the checkFallibleLinear/corevalidate/interp/cgen four-package wiring are all in place for plan 02 (RES-01's three-acquisition release-ordering fixture) and plan 03 (FFI-01's fuller three-layer contract and conformance TU) to extend directly.
- `core.TerminatorKinds()` currently returns `{OpReturn, OpFail}`; a later plan's defect terminator will extend it again, and `pathoracle`/`originvalidate` still only walk `OpReturn` — D-04-29's widening to `{OpReturn, OpFail, OpDefect}` is explicitly out of this plan's scope and remains open debt for a later plan in this phase.
- No blockers.

## Self-Check: PASSED

All key files verified present on disk; all three task commits (`d529905`, `62a443d`, `7df0211`) verified present in `git log`.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-04*
