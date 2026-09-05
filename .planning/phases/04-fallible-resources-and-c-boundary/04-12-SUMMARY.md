---
phase: 04-fallible-resources-and-c-boundary
plan: "12"
subsystem: compiler-validation
tags: [go, corevalidate, cgen, code-injection, mutation-kill, tdd, ffi]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "04-03's core.ForeignContract with its Symbol field, and cgen's existing emitLinearForeign/singleForeignFunction emission sites"
provides:
  - "corevalidate.validCIdentifier and the foreign.symbol_not_identifier refusal in the ForeignContract validation block, immediately after core.foreign_contract_missing"
  - "cgen's own independent validForeignSymbol guards in emitLinearForeign and singleForeignFunction, refusing without depending on corevalidate.Validate having run"
  - "TestForeignSymbolNotIdentifierRefused (corevalidate), TestForeignSymbolInjectionNeverReachesGeneratedC and TestForeignEmittersRefuseNonIdentifierSymbolIndependently (cgen)"
  - "a Mutation-Kill Register row proving both audit layers are independently required and both predicates assert identifier SHAPE, not mere presence"
  - "a named, unfixed finding: EmitForeignHeader's obligation comment block splices Allocator/Unwind/NonlocalExit/Fails/InitializedState/Capture/Retention/Aliasing raw into C comments by the same unsanitized-splice pattern Symbol used before this plan"
affects: [cgen, corevalidate, phase-04-verification]

actuals:
  tokens: 5234
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "A DELIBERATE second implementation of the same predicate in two independent packages (corevalidate.validCIdentifier / cgen.validForeignSymbol), so a validator-bypassing consumer still refuses on its own terms -- the same independence posture corevalidate already takes toward check.go"

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "validCIdentifier and validForeignSymbol are hand-rolled byte loops (matching cName/cLocal's existing idiom), not a regexp import -- corevalidate.go's import block is unchanged (git diff shows zero hunks inside `import (`)."
  - "The audit passes operation.ID as Problem.Detail, never the Symbol itself, so an attacker-controlled string carrying newlines or quotes never reaches diagnostic JSON; cgen's two refusal messages use %q for the same reason."
  - "Found and recorded (not fixed, explicitly out of scope): EmitForeignHeader's obligation comment block splices Allocator, Unwind, NonlocalExit, Fails, InitializedState, Capture, Retention and Aliasing raw into C comments via the same fmt.Fprintf(\"%s\", ...) pattern Symbol's own comment used before this plan. See 'Additional Finding (Out of Scope)' below."
  - "Updated TestAcyclicChainsStillValidateUnderCycleGuard's pinned counted-work constant from 403 to 406 (mechanical +1 v.check per of the fixture's three OpForeignCall operations); LinearWorkLimit and the exact-formula TestCoreValidationWorkSeries are unaffected since scaleProgram declares no OpForeignCall."

requirements-completed: [FFI-01]

coverage:
  - id: D1
    description: "corevalidate audits core.ForeignContract.Symbol's SHAPE (not merely presence) with foreign.symbol_not_identifier, positioned immediately after core.foreign_contract_missing and before every other foreign check"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestForeignSymbolNotIdentifierRefused"
        status: pass
    human_judgment: false
  - id: D2
    description: "cgen.Emit and cgen.EmitNative are never reached with a non-identifier Symbol (corevalidate.Validate refuses first), returning an empty generated string"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestForeignSymbolInjectionNeverReachesGeneratedC"
        status: pass
    human_judgment: false
  - id: D3
    description: "cgen's own independent validForeignSymbol guard refuses a hostile Symbol on EmitForeignManifest, EmitForeignHeader and EmitForeignConformance -- the three exported entry points that never call corevalidate.Validate"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestForeignEmittersRefuseNonIdentifierSymbolIndependently"
        status: pass
    human_judgment: false
  - id: D4
    description: "Mutation-kill demonstration: reverting either audit layer, or weakening either predicate to a non-empty check, turns a named test red in a throwaway detached worktree that never touched the working branch"
    requirement: "FFI-01"
    verification:
      - kind: other
        ref: "verbatim mutation outputs captured below, in throwaway detached git worktrees, never landed on the working branch"
        status: pass
    human_judgment: false

duration: 45min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 12: Foreign symbol identifier audit at the C boundary Summary

**`corevalidate` now audits `core.ForeignContract.Symbol`'s identifier SHAPE (not just its presence) with `foreign.symbol_not_identifier`, and `cgen` independently refuses the same hostile input on the three exported `EmitForeign*` entry points that never call `Validate` -- closing 04-VERIFICATION.md gap 2's arbitrary-C-injection hole at the phase's "audited" C boundary.**

## Performance

- **Duration:** 45 min
- **Tasks:** 3
- **Files modified:** 5 (`corevalidate.go`, `corevalidate_test.go`, `cgen.go`, `cgen_test.go`, `04-VALIDATION.md`)

## Accomplishments

- Added `validCIdentifier` to `corevalidate.go` and inserted the `foreign.symbol_not_identifier` refusal into the existing `OpForeignCall` validation block, immediately after `core.foreign_contract_missing` and before `foreign.unwind_policy_undeclared` -- so the audit fires before every downstream foreign check and before any emitter runs.
- Added `TestForeignSymbolNotIdentifierRefused`: seven hostile-Symbol subtests (semicolon/brace, parenthesis, embedded newline, leading digit, embedded space, comment terminator, non-ASCII rune), a negative-result subtest proving an identifier-shaped-but-unknown symbol is not refused by this code, and an end-to-end subtest proving `cgen.Emit` refuses with an empty generated string.
- Added `validForeignSymbol` to `cgen.go` -- a deliberate independent second implementation of the same predicate -- and extended `emitLinearForeign`'s existing guard plus `singleForeignFunction` (which backs `EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance`, none of which calls `Validate`) so all five entry points refuse a hostile Symbol.
- Added `TestForeignSymbolInjectionNeverReachesGeneratedC` and `TestForeignEmittersRefuseNonIdentifierSymbolIndependently`, both driven from a shared hostile-Symbol table, each with a positive control proving the unmutated program still emits through all five entry points.
- Proved both layers with a three-mutation-kill demonstration in throwaway detached git worktrees (never landed on the working branch): reverting either audit alone, and weakening either predicate to a presence-only check, each turn a named test red for a distinct, predicted reason.
- Appended one row to the Mutation-Kill Register in `04-VALIDATION.md`, owning plan `04-12`.
- Found and recorded (explicitly not fixed, out of this plan's scope) that `EmitForeignHeader`'s obligation comment block splices eight other `core.ForeignContract` string fields raw into C comments by the same unsanitized-splice pattern Symbol used before this plan.

## Task Commits

1. **Task 1: Audit the Symbol's identifier shape in corevalidate and falsify the injection end-to-end** -- `b77332a` (feat)
2. **Task 2: Give cgen its own independent refusal on the three entry points that never call Validate** -- `6b55866` (feat)
3. **Task 3: Mutation-kill both audit layers and record the Mutation-Kill Register row** -- `42485b0` (docs)

_Note: Task 1 was declared `tdd="true"`. RED was observed via a throwaway detached git worktree at the pre-Task-1 commit (`e8b1850`), running only the new test file against the unmodified production code, then deleted; the production fix and its test landed together in the single `feat(04-12)` commit, matching this project's established `type="execute"` convention (see 04-11-SUMMARY.md's identical rationale) rather than a separate `test(...)`/`feat(...)` commit split. See "TDD Gate Compliance" below._

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` -- `validCIdentifier` helper; `foreign.symbol_not_identifier` refusal inserted into the `OpForeignCall` block between `core.foreign_contract_missing` and `foreign.unwind_policy_undeclared`.
- `internal/compiler/corevalidate/corevalidate_test.go` -- `TestForeignSymbolNotIdentifierRefused`; `cgen` import added; `TestAcyclicChainsStillValidateUnderCycleGuard`'s pinned constant updated 403 -> 406.
- `internal/compiler/cgen/cgen.go` -- `validForeignSymbol` helper (placed immediately above `foreignExternName`); guard added to `emitLinearForeign` and `singleForeignFunction`.
- `internal/compiler/cgen/cgen_test.go` -- `hostileForeignSymbols` shared table; `TestForeignSymbolInjectionNeverReachesGeneratedC`; `TestForeignEmittersRefuseNonIdentifierSymbolIndependently`.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` -- one new Mutation-Kill Register row, owning plan `04-12`.

## Decisions Made

- **Hand-rolled byte loops, not regexp.** Both `validCIdentifier` and `validForeignSymbol` implement `^[A-Za-z_][A-Za-z0-9_]*$` as explicit byte loops (iterating bytes, not runes, so a multibyte rune is rejected by its individual bytes), matching `cName`/`cLocal`'s existing idiom. `corevalidate.go`'s import block is byte-unchanged (confirmed via `git diff`, no hunk inside `import (`).
- **`operation.ID`, never the Symbol, as `Problem.Detail`.** `Problem.Detail` is serialized into diagnostic JSON; echoing an attacker-controlled string containing newlines or quotes into that channel would be the same class of defect the check exists to close. `cgen`'s two refusal messages use `%q` for the identical reason.
- **`validForeignSymbol` is a deliberate second implementation, not a shared helper.** `corevalidate.validCIdentifier` is unexported; `cgen` must be able to refuse a hostile Symbol without depending on `corevalidate.Validate` having run first, matching corevalidate's own independence posture toward `check.go` throughout this phase.
- **Mechanically updated a pinned counted-work constant.** `TestAcyclicChainsStillValidateUnderCycleGuard` pins `acquire_three_success.lang`'s exact `Checks` count; this plan's one new accepting-path `v.check` per `core.OpForeignCall` moves it from 403 to 406 (the fixture declares three foreign calls). `LinearWorkLimit` and the exact-formula `TestCoreValidationWorkSeries`/`TestReleaseOrderValidationWorkSeries` are confirmed unaffected (`scaleProgram` declares no `OpForeignCall`).

## Confirmations From the Plan's `<planner_assumptions>` and Task Instructions

**`scaleProgram` and the counted-work series (Task 1's counted-work note):** confirmed by reading -- `scaleProgram` (corevalidate_test.go:325) declares no `core.OpForeignCall`, so `TestCoreValidationWorkSeries`'s exact-formula assertion is untouched by the new check. `TestReleaseOrderValidationWorkSeries` (corevalidate_test.go:709) uses `discard_because.lang` (0 acquisitions, +0), `foreign_acquire_one.lang` (1 acquisition, +1), and `acquire_three_success.lang` (3 acquisitions, +3) -- the uniform per-`OpForeignCall` increment preserves the series' strict monotonicity, and both tests pass unmodified.

**Import cycle check (Task 1's "if importing cgen from corevalidate_test would create an import cycle" contingency):** the import does NOT create a cycle. `corevalidate_test` is package `corevalidate_test` (external test package); `cgen` imports `corevalidate` (production package), not `corevalidate_test`. `go build ./...` and the full test suite confirm no cycle -- the end-to-end subtest stayed in `corevalidate_test.go` as originally planned; no move to `cgen_test.go` was needed.

## Additional Finding (Out of Scope)

Per Task 2's explicit instruction to confirm-and-record (not fix) whether any other `core.ForeignContract` string field reaches generated C by the same raw-splice route:

**Confirmed: `EmitForeignHeader`'s obligation comment block (`internal/compiler/cgen/cgen.go`, the `fmt.Fprintf(&out, "/* ... : %s */\n", ...)` sequence at approximately lines 1333-1340, immediately following the now-guarded `symbol:` comment line) splices EIGHT other `core.ForeignContract` string fields raw into C comments, unsanitized, by the exact same pattern `Symbol`'s comment used before this plan:**

- `contract.Allocator` -- `/* allocator: %s */`
- `contract.Unwind` -- `/* unwind: %s */`
- `contract.NonlocalExit` -- `/* nonlocal_exit: %s */`
- `contract.Fails` -- `/* fails: %s */`
- `contract.InitializedState` -- `/* initialized_state: %s */`
- `contract.Capture` -- `/* capture: %s (unchecked_obligation) */`
- `contract.Retention` -- `/* retention: %s (unchecked_obligation) */`
- `contract.Aliasing` -- `/* aliasing: %s (unchecked_obligation) */`

`corevalidate` only checks these fields for non-empty presence (`Unwind != "" && NonlocalExit != ""` at corevalidate.go's `foreign.unwind_policy_undeclared` check; `InitializedState != "" && Capture != "" && Retention != "" && Aliasing != ""` at `foreign.obligation_undeclared`) -- never their shape. `Allocator` and `Fails` reach `EmitForeignManifest`'s JSON output through `encoding/json` (safe), but ALSO reach `EmitForeignHeader`'s comment block raw (NOT safe) -- the plan's own out-of-scope assumption ("Allocator and Fails reach generated C only through already sanitized or JSON-encoded paths") is INCOMPLETE for the header comment path specifically, though correct for the manifest.

On the honest compiler pipeline, `InitializedState`/`Capture`/`Retention`/`Aliasing` are compiler-derived fixed constants (`standardForeignObligations()` in check.go always returns `"fully", "none", "none", "none"`) never influenced by source text, so they cannot carry a hostile value via any Lang source program today -- only via a corrupted `core.Program` artifact, the same threat class Symbol was defended against. `Unwind`/`NonlocalExit`/`Allocator` ARE parsed from source text (`unwind: forbidden`, `allocator: "libc_malloc"`) with no vocabulary restriction beyond "an identifier or string token" (`internal/compiler/syntax/parser.go` lines 220-241) -- meaning a real Lang source file, not merely a corrupted artifact, could in principle declare `unwind: "*/int evil(void){}/*"` and have it reach the header comment unescaped. This was NOT verified against the actual string-literal lexer's escaping rules (out of scope for this plan to investigate further), and NO fix was attempted here per the plan's explicit instruction not to expand scope. This finding is recorded for the verifier and the next planning round.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug/mechanical test update] Updated a pinned counted-work constant**
- **Found during:** Task 1, running the full compiler test suite after landing the audit
- **Issue:** `TestAcyclicChainsStillValidateUnderCycleGuard` pins `acquire_three_success.lang`'s exact validator `Checks` count at `403`. Adding one accepting-path `v.check` per `core.OpForeignCall` (the fixture declares three) moved the true count to `406`, exactly as the plan's own "Counted-work note" anticipated and instructed confirming rather than assuming.
- **Fix:** Updated the pinned constant from `403` to `406` with a comment explaining the mechanical cause.
- **Files modified:** `internal/compiler/corevalidate/corevalidate_test.go`
- **Verification:** `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./internal/compiler/... -count=1` passes; `LinearWorkLimit` and the exact-formula `TestCoreValidationWorkSeries`/`TestReleaseOrderValidationWorkSeries` confirmed unaffected.
- **Committed in:** `b77332a` (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (mechanical test-constant update, directly caused by this task's own change and explicitly anticipated by the plan). **Impact:** None beyond the anticipated ripple; no scope creep, no behavior change on any accepting path.

## TDD RED Evidence (Task 1, captured verbatim)

Captured by copying only the new `corevalidate_test.go` into a throwaway detached git worktree (`git worktree add --detach`) at the pre-fix commit `e8b1850`, running the test, then deleting the worktree with `git worktree remove --force`:

```
=== RUN   TestForeignSymbolNotIdentifierRefused
=== RUN   TestForeignSymbolNotIdentifierRefused/semicolon_and_brace_closing_the_extern_and_opening_a_new_definition
    corevalidate_test.go:569: expected foreign.symbol_not_identifier, got {Valid:true Problems:[] Checks:130 ...}
[... all seven hostile subtests report Valid:true -- the unfixed validator ACCEPTS every corrupted program ...]
--- FAIL: TestForeignSymbolNotIdentifierRefused (0.00s)
    --- FAIL: .../semicolon_and_brace_closing_the_extern_and_opening_a_new_definition (0.00s)
    --- FAIL: .../parenthesis-bearing_fragment (0.00s)
    --- FAIL: .../embedded_newline (0.00s)
    --- FAIL: .../leading_digit (0.00s)
    --- FAIL: .../embedded_space (0.00s)
    --- FAIL: .../comment_terminator_escaping_the_header_comment (0.00s)
    --- FAIL: .../non-ASCII_rune (0.00s)
    --- PASS: .../identifier-shaped_but_unknown_symbol_is_not_refused_by_this_check (0.00s)
    --- FAIL: TestForeignSymbolNotIdentifierRefused/end-to-end:_cgen.Emit_refuses_before_generating_any_C (0.00s)
FAIL
```

The end-to-end subtest's failure captured the injected C verbatim -- `cgen.Emit`, unfixed, spliced the hostile Symbol directly into generated C:

```c
typedef struct _LANG_lang_res_open;}
int injected(void){return 0;}//_result {
  unsigned char ok;
  unsigned char value;
} _LANG_lang_res_open;}
int injected(void){return 0;}//_result;

extern _LANG_lang_res_open;}
int injected(void){return 0;}//_result _LANG_lang_res_open;}
int injected(void){return 0;}//(unsigned char argument);
```

The `int injected(void){return 0;}` fragment appears as genuine, unescaped top-level C source -- direct proof the injection channel was open before the fix.

## Mutation-Kill Demonstration (Task 3)

All four mutations were applied and reverted inside throwaway detached `git worktree`s, deleted immediately after each capture (`git worktree remove --force`); `git status --short` and `diff -q` against the working branch confirmed byte-identical restoration after each mutation and after the whole demonstration.

### Mutation A -- revert corevalidate's audit only, keep cgen's guards

Deleted the `foreign.symbol_not_identifier` check and its comment block from `corevalidate.go` (leaving `validCIdentifier` itself, and cgen's independent guards from Task 2, untouched).

```
=== RUN   TestForeignSymbolNotIdentifierRefused/semicolon_and_brace_closing_the_extern_and_opening_a_new_definition
    corevalidate_test.go:569: expected foreign.symbol_not_identifier, got {Valid:true Problems:[] Checks:130 ...}
[... all seven hostile subtests: Valid:true -- the unfixed validator accepts the corrupted program ...]
=== RUN   TestForeignSymbolNotIdentifierRefused/end-to-end:_cgen.Emit_refuses_before_generating_any_C
    corevalidate_test.go:594: expected error to name foreign.symbol_not_identifier, got foreign symbol "lang_res_open;}\nint injected(void){return 0;}//" is not a C identifier
--- FAIL: TestForeignSymbolNotIdentifierRefused (0.00s)
FAIL
```

All seven hostile subtests go red because the reverted validator silently ACCEPTS the corrupted program (`Valid:true`). The end-to-end subtest is also red, but for the expected DIFFERENT reason predicted by defense-in-depth: Task 2's independent `cgen` guard (still present in this mutation) still refuses the call, just with its own message text ("is not a C identifier") rather than the code string the test asserts by name -- proving cgen's guard is a genuinely independent second layer, not decoration riding on corevalidate.

### Mutation B -- revert cgen's two guards only, keep corevalidate's audit

Removed the guard clauses from `emitLinearForeign` and `singleForeignFunction` (leaving `validForeignSymbol` itself, and corevalidate's audit from Task 1, untouched).

```
=== RUN   TestForeignEmittersRefuseNonIdentifierSymbolIndependently/semicolon_and_brace_closing_the_extern_and_opening_a_new_definition
    cgen_test.go:309: EmitForeignManifest: expected an error, got:
        {"schema":"lang.foreign/0","symbol":"lang_res_open;}\nint injected(void){return 0;}//","allocator":"libc_malloc", ...}
[... all seven hostile subtests: EmitForeignManifest succeeds and returns the manifest carrying the injected Symbol verbatim ...]
--- FAIL: TestForeignEmittersRefuseNonIdentifierSymbolIndependently (0.00s)
FAIL
```

All seven subtests go red even though corevalidate's audit is fully present, because `EmitForeignManifest`, `EmitForeignHeader` and `EmitForeignConformance` never call `corevalidate.Validate` -- proving the second, independent layer was genuinely required, not defensive decoration.

A direct probe of `EmitForeignHeader` under this same mutation (isolated because the table-driven test's `t.Fatalf` on the manifest step stops before reaching the header assertion) captured the injected header text verbatim:

```c
/* symbol: lang_res_open*/int injected(void){return 0;}/* */
...
extern LANG_LANG_RES_OPEN__INT_INJECTED_VOID__RETURN_0_____RESULT _LANG_lang_res_open*/int injected(void){return 0;}/*(unsigned char argument);
```

The `/* symbol: ... */` comment's own terminator (`*/`) is escaped early by the injected fragment, closing the comment and exposing `int injected(void){return 0;}` as live, uncommented top-level C source -- the single strongest piece of evidence this plan produces that the injection channel was real and reachable.

### Mutation C -- weaken both predicates to a non-empty check (the shape before this plan)

Replaced both `validCIdentifier` and `validForeignSymbol` bodies with `return name != ""` / `return symbol != ""`.

corevalidate:
```
=== RUN   TestForeignSymbolNotIdentifierRefused
[... all seven hostile subtests FAIL: the weakened predicate accepts every non-empty hostile string ...]
--- FAIL: TestForeignSymbolNotIdentifierRefused (0.00s)
FAIL
```

cgen:
```
=== RUN   TestForeignSymbolInjectionNeverReachesGeneratedC
[... all seven hostile subtests FAIL ...]
=== RUN   TestForeignEmittersRefuseNonIdentifierSymbolIndependently
[... all seven hostile subtests FAIL: EmitForeignManifest succeeds with the injected Symbol ...]
FAIL
```

Both layers' falsifiers go red under the weakened predicate in both packages -- proving the tests assert identifier SHAPE, not merely that some check function exists and returns a boolean.

After each mutation, the mutated file was restored (`git checkout --`) and confirmed byte-identical to the working branch via `diff -q` before proceeding to the next mutation; the worktree was deleted after the full demonstration.

## Issues Encountered

None.

## Verification Results

- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestForeignSymbolNotIdentifierRefused TestForeignContractInternallyValidated TestForeignRefusalsAreIndependentlyDerived TestCoreValidationWorkSeries TestReleaseOrderValidationWorkSeries` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen TestForeignSymbolInjectionNeverReachesGeneratedC TestForeignEmittersRefuseNonIdentifierSymbolIndependently TestExistingEmittersAreByteIdentical TestGeneratedForeignHeaderNamesAreAllocated TestObligationCommentsAreGeneratedFromJSON TestConformanceUnitAssertsEveryField` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./... -count=1` -- all packages pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` -- exit 0, no diagnostics
- `sh scripts/verify-phase4.sh` -- exit 0, every one of 32 lane entries `"status":"pass"` across all four `command:verify` invocations (deterministic, owned, borrowed, foreign/resource lanes)
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestPhase4ReachabilityRecordIsComplete TestVerifyPhase4ControlsAndWork` -- pass
- `git diff --stat -- testdata/` -- empty; every golden byte-frozen
- `git diff .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` -- zero deleted lines, one row appended
- `git diff --stat scripts/verify-phase4.sh internal/compiler/session/session.go internal/compiler/check/check.go` -- empty
- `git status --short` after all mutation-kill work -- clean relative to this plan's own committed changes; no leftover worktree directory

## TDD Gate Compliance

This is a `type: execute` plan with two `tdd="true"` tasks (Task 1 and, implicitly, Task 2's TDD-ordered instruction), not a `type: tdd` plan, so plan-level RED/GREEN/REFACTOR commit-gate enforcement (a separate `test(...)` then `feat(...)` commit) does not apply structurally. Per-task, the plan's own ordering instruction ("write the test first and observe it go red... then land the audit and observe green") was followed as an in-session verification discipline for both tasks: RED was genuinely observed in a throwaway detached worktree against unmodified production code (captured verbatim above for Task 1; captured verbatim in Mutation B's demonstration for Task 2's `EmitForeign*` entry points specifically, which are red regardless of Task 1), and only then was each production fix landed -- with tests and fix committed together in each task's single `feat(04-12)` commit, matching this project's established `type: execute` convention (04-11-SUMMARY.md's identical precedent). No gate violation: RED was genuinely observed and recorded before GREEN, per task.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- 04-VERIFICATION.md gap 2's three `missing:` items are closed: item 1 by the `foreign.symbol_not_identifier` audit placed immediately after the non-empty check; item 2 by `TestForeignSymbolNotIdentifierRefused`; item 3 by the two cgen regression tests proving `Emit`/`EmitNative` are never reached with such a Symbol and that the `EmitForeign*` entry points independently refuse it.
- Re-running `/gsd-verify-work` on Phase 04 should flip truth 2a from FAILED to verified and both `core.ForeignContract.Symbol` key links from NOT WIRED / unsafely-wired to WIRED.
- The Additional Finding above (Allocator/Unwind/NonlocalExit/Fails/InitializedState/Capture/Retention/Aliasing splicing raw into `EmitForeignHeader`'s comment block) is an adjacent, unfixed instance of the same defect class and should be triaged in the next planning round -- particularly `Unwind`/`NonlocalExit`/`Allocator`, which are reachable from real Lang source text, not only from a corrupted core artifact.
- This was 04-12, the second and final gap-closure plan queued after 04-07's phase-close approval (alongside sibling plan 04-11, which closed gap 1). With both gaps closed, Phase 4 is ready for re-verification.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*

## Self-Check: PASSED

- FOUND: .planning/phases/04-fallible-resources-and-c-boundary/04-12-SUMMARY.md
- FOUND commit b77332a (Task 1: feat)
- FOUND commit 6b55866 (Task 2: feat)
- FOUND commit 42485b0 (Task 3: docs)
