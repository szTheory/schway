---
phase: 04-fallible-resources-and-c-boundary
plan: "13"
subsystem: compiler-validation
tags: [go, check, corevalidate, cgen, code-injection, mutation-kill, tdd, ffi]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "04-12's foreign symbol identifier audit pattern (validCIdentifier/validForeignSymbol, the independent-layer posture, and the Mutation-Kill Register idiom)"
provides:
  - "check.validForeignPolicyValue and the check.foreign_policy_value_unsafe source-admission refusal in collectForeignSymbols, refusing a hostile allocator/unwind/nonlocal_exit (or any other policy key's value) before a core artifact is ever produced"
  - "corevalidate.foreign.policy_value_not_identifier (Allocator/Unwind/NonlocalExit identifier-shape audit) and corevalidate.foreign.contract_field_not_c_safe (commentSafe/validCTypeExpression/foreignContractFieldsCSafe audit of every remaining spliced field)"
  - "cgen's own independent commentSafeForeignField/validForeignCType/unsafeForeignContractField guard in singleForeignFunction, refusing before EmitForeignManifest/EmitForeignHeader/EmitForeignConformance ever splice a hostile field"
  - "TestForeignPolicyValueInjectionRefusedFromSource (end-to-end tracer), TestForeignPolicyValueUnsafeRefusedAtAdmission (check), TestForeignPolicyValueNotIdentifierRefused and TestForeignContractCommentSafetyRefused (corevalidate), TestForeignPolicyValueInjectionNeverReachesGeneratedC (cgen)"
  - "a fourth Mutation-Kill Register row proving all three audit layers are independently required and every predicate asserts SHAPE/comment-safety, not mere presence"
affects: [check, corevalidate, cgen, phase-04-verification]

actuals:
  tokens: 13309
  tasks: 4
  commits: 4

tech-stack:
  added: []
  patterns:
    - "A THIRD deliberate independent implementation of the C-identifier predicate (check.validForeignPolicyValue, alongside corevalidate.validCIdentifier and cgen.validForeignSymbol), extending 04-12's independence posture to the source-admission layer"
    - "commentSafe/commentSafeForeignField as a second predicate family (refuses */, /*, control bytes, non-ASCII) for fields that are legitimately non-identifier prose, paired with validCTypeExpression/validForeignCType for the one field spliced as a real C token sequence rather than into a comment"

key-files:
  created:
    - testdata/phase4/foreign_policy_value_injection.lang
  modified:
    - internal/compiler/check/check.go
    - internal/compiler/check/check_test.go
    - internal/compiler/native/native_test.go
    - internal/compiler/session/session_test.go
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "Full C-identifier shape (not comment-safety-only) for Allocator/Unwind/NonlocalExit (DD-04-13-01), matching Symbol's existing rule; every in-repo policy value (forbidden, possible, libc_malloc, a_different_allocator) is already a bare identifier, so nothing accepted today is newly refused."
  - "commentSafe/commentSafeForeignField refuse /* as well as */ because this project compiles generated C with warnings-as-errors, and an unclosed nested comment is a diagnosed condition, not merely a stylistic issue."
  - "validCTypeExpression/validForeignCType (space-separated C identifiers) exist ONLY for Layout.Fields[].CType, because its sole honest value (unsigned char) itself contains a space -- the plain identifier rule used for the three policy values would wrongly refuse it."
  - "check.foreign_policy_value_unsafe fires at DECLARATION time (inside collectForeignSymbols' policy loop) even though the sibling unwind/nonlocal_exit PRESENCE gate deliberately defers to checkFallibleLinear: a hostile VALUE is never legitimate for any symbol, called or not, so refusing it earliest is not in tension with that deferral, which is about an ABSENT key."
  - "All three new refusal codes pass a span or operation.ID, never the offending field's value, into diagnostic/error output -- echoing an attacker-controlled string into an output channel would be the same class of defect these checks exist to close; cgen's error names a field from a fixed vocabulary, a stronger guarantee than %q escaping."
  - "Session.go's phase4 negativeControls table is confirmed (by reading, not assumed) to be a curated SELECTED-control list of two named lanes, not an exhaustive-per-testdata/phase4-fixture list -- this plan adds no new lane, per its own explicit instruction to record rather than invent one."

requirements-completed: [FFI-01]

coverage:
  - id: D1
    description: "An ordinary Lang source declaring allocator: \"*/ int injected(void){return 1;} /*\" is refused at source admission by session.Check with check.foreign_policy_value_unsafe (span-bearing) and produces no checked function -- the honest-source path goes red before corevalidate or cgen ever run"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestForeignPolicyValueInjectionRefusedFromSource"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestForeignPolicyValueUnsafeRefusedAtAdmission"
        status: pass
    human_judgment: false
  - id: D2
    description: "corevalidate independently re-derives the same refusal purely from core.ForeignContract's three flat policy fields (foreign.policy_value_not_identifier), so a corrupted core.Program that skipped check.go's gate is still caught, without displacing the existing foreign.unwind_policy_undeclared code for an omitted policy"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestForeignPolicyValueNotIdentifierRefused"
        status: pass
    human_judgment: false
  - id: D3
    description: "corevalidate audits every REMAINING spliced contract string (foreign.contract_field_not_c_safe) and cgen's own singleForeignFunction guard independently refuses the whole hostile contract on the three EmitForeign* entry points that never call Validate, returning an empty string and naming only the offending field"
    requirement: "FFI-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestForeignContractCommentSafetyRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/cgen/cgen_test.go#TestForeignPolicyValueInjectionNeverReachesGeneratedC"
        status: pass
    human_judgment: false
  - id: D4
    description: "Mutation-kill demonstration: reverting any one of the three audit layers, or weakening any predicate to a non-empty/always-true check, turns a named test red, proving each layer is independently necessary and each predicate asserts shape, not mere presence"
    requirement: "FFI-01"
    verification:
      - kind: other
        ref: "verbatim mutation outputs captured below, via in-tree revert/run/restore (matching 04-12's recorded no-worktree discipline), never landed on the working branch"
        status: pass
    human_judgment: false

duration: 30min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 13: Foreign contract field audit at the C boundary, source-reachable path Summary

**Every `core.ForeignContract` string field `cgen` splices into generated C -- not just `Symbol` (04-12) -- is now validated at three independent layers, and the three source-reachable policy values (`allocator`, `unwind`, `nonlocal_exit`) are refused at SOURCE ADMISSION so an ordinary Lang author cannot inject a live top-level C function definition through a quoted foreign-policy value.**

## Performance

- **Duration:** ~30 min
- **Tasks:** 4
- **Files modified:** 10 (1 created, 9 modified)

## Accomplishments

- Added `check.validForeignPolicyValue` (a third deliberate byte-loop C-identifier implementation) and the `check.foreign_policy_value_unsafe` refusal inside `collectForeignSymbols`' policy loop, firing BEFORE the key-specific switch for every policy value regardless of key -- so a hostile `allocator`/`unwind`/`nonlocal_exit` value never reaches a core artifact at all.
- Added `testdata/phase4/foreign_policy_value_injection.lang`, an otherwise-honest copy of `foreign_acquire_one.lang` whose sole `allocator:` value is the comment-escaping payload, registered in both exhaustive corpus tables (`checkerCorpusVerdicts`, `phase4CorpusMatrix`).
- Added `TestForeignPolicyValueUnsafeRefusedAtAdmission` (check, 8 hostile shapes + 1 negative row) and the end-to-end tracer `TestForeignPolicyValueInjectionRefusedFromSource` (session), proving `session.Check` refuses the honest fixture with a span-bearing diagnostic and zero checked functions.
- Inserted `corevalidate`'s `foreign.policy_value_not_identifier` audit (re-deriving the same refusal purely from `core.ForeignContract`'s three flat fields, positioned after `foreign.unwind_policy_undeclared` and before `core.call_target_not_foreign`) and `foreign.contract_field_not_c_safe` (covering `Fails`, `InitializedState`, `Capture`, `Retention`, `Aliasing`, `Layout.ForeignTypeName`, and each `Layout.Fields[].Name`/`.CType`, positioned after `foreign.obligation_undeclared`), backed by new `commentSafe`/`validCTypeExpression`/`foreignContractFieldsCSafe` helpers.
- Gave `cgen` its own independent `commentSafeForeignField`/`validForeignCType`/`unsafeForeignContractField` guard in `singleForeignFunction`, refusing before `EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance` (none of which call `Validate`) ever splice a hostile field, naming only the offending field in the returned error.
- Added `TestForeignPolicyValueNotIdentifierRefused` and `TestForeignContractCommentSafetyRefused` (corevalidate) and `TestForeignPolicyValueInjectionNeverReachesGeneratedC` (cgen), each with positive controls and negative-result rows proving the audits test SHAPE, not membership in an allowlist.
- Measured (not assumed) the counted-work ripple: `acquireThreeSuccessChecks` moved 406 -> 409 (Task 2) -> 412 (Task 3), exactly matching the plan's predicted arithmetic (one new accepting-path `v.check` per `core.OpForeignCall`, three foreign calls in the fixture).
- Proved all three layers independently necessary with a four-mutation demonstration (in-tree revert/run/restore, matching 04-12's no-worktree discipline) and appended one row to the Mutation-Kill Register, owning plan `04-13`.
- Confirmed and recorded three explicit findings the plan required: `EmitForeignManifest` routes every field through JSON encoding (not a raw-splice site); `core.ForeignContract.Alias` is never spliced into generated C by any emitter; and `session.go`'s `negativeControls` table is a curated selected-lane list, not an exhaustive-per-fixture list, so this plan adds no new lane.

## Task Commits

1. **Task 1: End-to-end tracer -- an honest .lang source carrying the comment-escaping allocator value is refused by session.Check** -- `ee73eb3` (feat)
2. **Task 2: Audit the three policy values' identifier shape in corevalidate, with the counted-work pin moved deliberately** -- `0962f32` (feat)
3. **Task 3: Audit every remaining spliced contract string in corevalidate and give cgen its own independent peer guard** -- `0fbc002` (feat)
4. **Task 4: Mutation-kill all three audit layers and record the Mutation-Kill Register row** -- `ae62ae8` (docs)

_Note: Task 1 was declared `tracer`/`tdd="true"`. RED was captured verbatim in-session (test/fixture/registration landed first, run against unmodified production code, then the fix landed and tests re-run green), matching this project's established `type: execute` convention of a single `feat(04-13)` commit per task rather than a separate `test(...)`/`feat(...)` split (04-11/04-12-SUMMARY.md's identical precedent). See "TDD Gate Compliance" below._

## Files Created/Modified

- `testdata/phase4/foreign_policy_value_injection.lang` -- the honest-source fixture (new).
- `internal/compiler/check/check.go` -- `validForeignPolicyValue`; the `check.foreign_policy_value_unsafe` refusal inside `collectForeignSymbols`' policy loop, before the `switch policy.Key`.
- `internal/compiler/check/check_test.go` -- `TestForeignPolicyValueUnsafeRefusedAtAdmission`; the `checkerCorpusVerdicts` row for the new fixture.
- `internal/compiler/native/native_test.go` -- the `phase4CorpusMatrix()` `refused(...)` row for the new fixture.
- `internal/compiler/session/session_test.go` -- `TestForeignPolicyValueInjectionRefusedFromSource`.
- `internal/compiler/corevalidate/corevalidate.go` -- `commentSafe`, `validCTypeExpression`, `foreignContractFieldsCSafe`; the `foreign.policy_value_not_identifier` and `foreign.contract_field_not_c_safe` `v.check` calls in the `OpForeignCall` block.
- `internal/compiler/corevalidate/corevalidate_test.go` -- `TestForeignPolicyValueNotIdentifierRefused`, `TestForeignContractCommentSafetyRefused`; `acquireThreeSuccessChecks` pin moved 406 -> 412.
- `internal/compiler/cgen/cgen.go` -- `commentSafeForeignField`, `validForeignCType`, `unsafeForeignContractField` (placed above `foreignExternName`); the guard added to `singleForeignFunction`.
- `internal/compiler/cgen/cgen_test.go` -- `hostileForeignContractFields` shared table; `mutateForeignContractField`; `TestForeignPolicyValueInjectionNeverReachesGeneratedC`.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` -- one new Mutation-Kill Register row, owning plan `04-13`.

## Decisions Made

See `key-decisions` in the frontmatter for the five load-bearing decisions (DD-04-13-01's full-identifier scope, the `/*`-refusal rationale, `validCTypeExpression`'s narrow scope, the declaration-time firing point, the never-echo-the-value rule, and the confirmed reading of `session.go`'s control table).

## Confirmations From the Plan's `<planner_assumptions>` and Task Instructions

- **`core.ForeignContract.Alias` (planner_assumptions item 1c):** confirmed by reading `internal/compiler/cgen/cgen.go` -- grepping for `Alias` (excluding `Aliasing`) returns zero hits outside `core.go`'s own struct comment. `Alias` is never spliced into generated C by any emitter; it is compared only against the fixed `borrow`/`retain` vocabulary elsewhere in the compiler.
- **`EmitForeignManifest`'s routing (Task 3, confirmation (a)):** confirmed -- `foreignManifestDocument` (cgen.go) routes every field through `encoding/json`'s `json.Marshal`, so it is not a raw-splice site; only `EmitForeignHeader`'s comment block and `EmitForeignConformance`'s `_Static_assert` operands are.
- **Completeness of spliced-field coverage (Task 3, confirmation (c)):** after this plan, every `core.ForeignContract` string field `EmitForeignHeader`/`EmitForeignConformance` splices passes one of `validForeignSymbol` (Symbol, Allocator, Unwind, NonlocalExit, Layout.ForeignTypeName, Layout.Fields[].Name), `commentSafeForeignField` (Fails, InitializedState, Capture, Retention, Aliasing), or `validForeignCType` (Layout.Fields[].CType) inside cgen, and the parallel `validCIdentifier`/`commentSafe`/`validCTypeExpression` set inside corevalidate. No field is left unrouted.
- **`session.go`'s lane table (Task 1's read-finding):** confirmed by reading `verifyForeignCorpus`'s `negativeControls` slice (session.go:~2049-2057) -- it is a curated, SELECTED list of two named control lanes (`foreign_unwind_undeclared.lang`, `foreign_call_target_not_foreign.lang`), not an exhaustive per-`testdata/phase4`-fixture list. No test requires every fixture to own a lane. Per the plan's explicit instruction, this task adds NO new lane.
- **`scaleProgram`/counted-work series (Task 2/3's arithmetic note):** confirmed by reading -- `scaleProgram` declares no `core.OpForeignCall`, so `TestCoreValidationWorkSeries`'s exact-formula assertion is untouched; the uniform per-`OpForeignCall` increment (now +2 total, one per task) preserves `TestReleaseOrderValidationWorkSeries`'s strictly-increasing series. Both tests pass unmodified.
- **Formatter round-trip on the new fixture (Task 1):** confirmed via `TestShippedBinaryFourSubcommandCorpusMatrix` -- the payload lives entirely inside one `TokenString`, so `lang format --check` reaches a fixed point on it (exit 0, empty format diagnostic), exactly as registered in `phase4CorpusMatrix()`.

## Deviations from Plan

None -- plan executed exactly as written. The counted-work pin arithmetic (406 -> 409 -> 412) matched the plan's own predicted values exactly on measurement; no discrepancy to report.

## TDD RED Evidence (Task 1, captured verbatim)

Before landing `validForeignPolicyValue` and the `check.foreign_policy_value_unsafe` refusal, `TestForeignPolicyValueUnsafeRefusedAtAdmission` failed on all 8 hostile cases:

```
=== RUN   TestForeignPolicyValueUnsafeRefusedAtAdmission/allocator_carrying_the_comment-escaping_payload
    check_test.go:1499: expected check.foreign_policy_value_unsafe, got []
[... unwind, nonlocal_exit, semicolon-and-brace, embedded-space, leading-digit,
     non-ASCII-rune, and empty-string-literal cases: all "got []" ...]
--- PASS: TestForeignPolicyValueUnsafeRefusedAtAdmission/negative:_identifier-shaped_policy_values_are_not_refused
FAIL
```

The end-to-end tracer failed because `session.Check` returned zero diagnostics and a populated program for the honest fixture:

```
=== RUN   TestForeignPolicyValueInjectionRefusedFromSource
    session_test.go:439: expected at least one diagnostic, got none (program: {... Functions:[{ID:s1:phase4.foreign_policy_value_injection:fn:main ... ForeignContract:0x140001282c0 ...}]})
--- FAIL: TestForeignPolicyValueInjectionRefusedFromSource
```

The unfixed pipeline's `cgen.EmitForeignHeader` output for that program (captured via a throwaway test, deleted immediately after capture) -- the single strongest piece of evidence this plan produces:

```c
/* generated by Codename Lang; schema lang.c17/0 (foreign header, D-04-12a) */
#ifndef LANG_FOREIGN_LANG_LANG_RES_OPEN_H
#define LANG_FOREIGN_LANG_LANG_RES_OPEN_H

#include <stddef.h>

/* lang.foreign/0 obligations -- generated from the sidecar manifest;
 * see EmitForeignManifest. Never hand-edit this block: a hand-written
 * comment beside a generated JSON is a second source of truth that will
 * drift, which is exactly what D-04-12 forbids. */
/* symbol: lang_res_open */
/* allocator: */ int injected(void){return 1;} /* */
/* unwind: forbidden */
/* nonlocal_exit: forbidden */
/* fails: AcquireError */
/* initialized_state: fully */
/* capture: none (unchecked_obligation) */
/* retention: none (unchecked_obligation) */
/* aliasing: none (unchecked_obligation) */
/* layout.foreign_type_name: lang_foreign_resource_block */
/* layout.size: 1 */
/* layout.alignment: 1 */
/* layout.field: payload size=1 alignment=1 offset=0 */

typedef struct LANG_LANG_RES_OPEN_RESULT {
  unsigned char ok;
  unsigned char value;
} LANG_LANG_RES_OPEN_RESULT;

extern LANG_LANG_RES_OPEN_RESULT _LANG_lang_res_open(unsigned char argument);

_Static_assert(sizeof(LANG_LANG_RES_OPEN_RESULT) == 2, "LANG_LANG_RES_OPEN_RESULT must be a two-byte by-value ABI result");
_Static_assert(_Alignof(LANG_LANG_RES_OPEN_RESULT) == 1, "LANG_LANG_RES_OPEN_RESULT must have byte alignment");
_Static_assert(offsetof(LANG_LANG_RES_OPEN_RESULT, ok) == 0, "LANG_LANG_RES_OPEN_RESULT.ok must be the first field");
_Static_assert(offsetof(LANG_LANG_RES_OPEN_RESULT, value) == 1, "LANG_LANG_RES_OPEN_RESULT.value must follow ok");

#endif
```

The `/* allocator: */ int injected(void){return 1;} /* */` line -- the injected fragment's own `*/` closes the comment early, exposing `int injected(void){return 1;}` as live, uncommented top-level C source -- is the exact reproduction 04-VERIFICATION.md described.

Before landing Task 2's and Task 3's audits, the analogous corevalidate falsifiers (`TestForeignPolicyValueNotIdentifierRefused`, `TestForeignContractCommentSafetyRefused`) failed identically: every hostile subtest reported `Valid:true` (the unfixed validator silently accepted the corrupted contract), while the negative-result and ordering subtests already passed (confirming the test fixtures themselves were sound before the fix landed).

## Mutation-Kill Demonstration (Task 4)

All four mutations were applied via direct in-tree edits (matching 04-12's recorded no-worktree discipline for this session) and restored with `git checkout -- <file>` immediately after each capture; `git status --short` and `git diff --stat` confirmed byte-identical restoration after each mutation and after the whole demonstration (final `check.go` diffed byte-identical against its pre-mutation-A copy).

### Mutation A -- revert check.go's admission gate only (keep corevalidate audits and cgen guard)

Deleted `validForeignPolicyValue` and the `check.foreign_policy_value_unsafe` refusal from `check.go`.

```
=== RUN   TestForeignPolicyValueInjectionRefusedFromSource
    session_test.go:439: expected at least one diagnostic, got none (program: {... Functions:[{... ForeignContract:0x1400013a2c0 ...}]})
--- FAIL: TestForeignPolicyValueInjectionRefusedFromSource (0.00s)
FAIL
```

Red for the exact predicted reason: `session.Check` returns zero diagnostics for the honest fixture -- the reproduction 04-VERIFICATION.md described.

### Mutation B -- revert both corevalidate audits only (keep check.go's gate and cgen's guard)

Deleted the `foreign.policy_value_not_identifier` and `foreign.contract_field_not_c_safe` `v.check` calls from `corevalidate.go` (leaving their helpers and cgen's Task-3 guard untouched).

```
=== RUN   TestForeignPolicyValueNotIdentifierRefused/Allocator:_comment-terminator_payload
    corevalidate_test.go:643: expected foreign.policy_value_not_identifier, got {Valid:true Problems:[] Checks:131 ...}
[... all nine hostile field x shape cases, plus the empty-Allocator case: Valid:true ...]
--- FAIL: TestForeignPolicyValueNotIdentifierRefused (0.00s)

=== RUN   TestForeignContractCommentSafetyRefused/Fails:_comment-terminator_payload
    corevalidate_test.go:727: expected foreign.contract_field_not_c_safe, got {Valid:true Problems:[] Checks:131 ...}
[... every hostile field x payload case across Fails/InitializedState/Capture/Retention/Aliasing/
     ForeignTypeName/Fields[0].Name/CType: Valid:true ...]
--- FAIL: TestForeignContractCommentSafetyRefused (0.00s)
FAIL
```

Both go red even though `check.go`'s source-admission gate is fully present, because a corrupted `core.Program` constructed directly (never passing through `check.go`) is unaffected by it -- proving corevalidate's re-derivation is genuinely independent, not decoration riding on Task 1.

### Mutation C -- revert cgen's guard only (keep check.go's gate and both corevalidate audits)

Deleted the `unsafeForeignContractField` guard clause from `singleForeignFunction` (leaving the helper functions and both corevalidate checks untouched).

```
=== RUN   TestForeignPolicyValueInjectionNeverReachesGeneratedC/Allocator:_comment-terminator_payload
    cgen_test.go:433: EmitForeignManifest: expected an error, got:
        {"schema":"lang.foreign/0","symbol":"lang_res_open","allocator":"*/ int injected(void){return 1;} /*", ...}
[... all eleven hostile field cases: EmitForeignManifest succeeds and returns the manifest
     carrying the injected value verbatim ...]
--- FAIL: TestForeignPolicyValueInjectionNeverReachesGeneratedC (0.00s)
FAIL
```

A direct probe of `EmitForeignHeader` under this same mutation (the allocator case) captured the injected header text verbatim -- the single strongest piece of evidence this plan produces:

```c
/* symbol: lang_res_open */
/* allocator: */ int injected(void){return 1;} /* */
/* unwind: forbidden */
...
```

Red even with both other layers present, because `EmitForeignManifest`, `EmitForeignHeader` and `EmitForeignConformance` never call `corevalidate.Validate` -- proving cgen's own guard was genuinely required, not defensive decoration.

### Mutation D -- weaken every predicate to a non-empty/always-true check (the shape before this plan), in all three implementations

Weakened `check.validForeignPolicyValue` to `value != ""`; `corevalidate.validCIdentifier` and `validCTypeExpression` to `name != ""` / `true`; `corevalidate.commentSafe` to `true`; `cgen.validForeignSymbol` and `validForeignCType` to `symbol != ""` / `true`; `cgen.commentSafeForeignField` to `true`.

```
check package:
=== RUN   TestForeignPolicyValueUnsafeRefusedAtAdmission
[... 7 of 8 hostile subtests FAIL (comment-terminator x3, semicolon-and-brace, embedded-space,
     leading-digit, non-ASCII-rune); the empty-string-literal case still correctly refuses
     since "" != "" is false ...]
--- FAIL: TestForeignPolicyValueUnsafeRefusedAtAdmission (0.00s)

session package:
=== RUN   TestForeignPolicyValueInjectionRefusedFromSource
    session_test.go:439: expected at least one diagnostic, got none
--- FAIL: TestForeignPolicyValueInjectionRefusedFromSource (0.00s)

corevalidate package:
--- FAIL: TestForeignPolicyValueNotIdentifierRefused (0.00s)
--- FAIL: TestForeignContractCommentSafetyRefused (0.00s)
[52 subtest failures across both tests]

cgen package:
--- FAIL: TestForeignPolicyValueInjectionNeverReachesGeneratedC (0.00s)
[11 subtest failures, one per hostile field]
```

Every new test from Tasks 1, 2 and 3 goes red under the weakened predicates (except the emptiness-only cases, which a non-empty check still correctly catches by construction) -- proving the tests assert identifier SHAPE and comment-safety, not merely that some check function exists and returns a boolean.

After each mutation, the mutated file(s) were restored with `git checkout -- <file>` and confirmed byte-identical to the working branch (`git status --short` clean relative to this plan's own committed changes; `diff -q` against a pre-mutation-A backup copy of `check.go` confirmed exact restoration).

## Issues Encountered

None.

## Verification Results

- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestForeignPolicyValueUnsafeRefusedAtAdmission TestCheckerVerdictsUnchanged` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestForeignPolicyValueInjectionRefusedFromSource` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/native TestShippedBinaryFourSubcommandCorpusMatrix` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestForeignPolicyValueNotIdentifierRefused TestForeignSymbolNotIdentifierRefused TestForeignContractInternallyValidated TestForeignRefusalsAreIndependentlyDerived TestCoreValidationWorkSeries TestReleaseOrderValidationWorkSeries TestAcyclicChainsStillValidateUnderCycleGuard` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/corevalidate TestForeignContractCommentSafetyRefused TestForeignPolicyValueNotIdentifierRefused TestAcyclicChainsStillValidateUnderCycleGuard TestCoreValidationWorkSeries TestReleaseOrderValidationWorkSeries` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/cgen TestForeignPolicyValueInjectionNeverReachesGeneratedC TestForeignSymbolInjectionNeverReachesGeneratedC TestForeignEmittersRefuseNonIdentifierSymbolIndependently TestExistingEmittersAreByteIdentical TestGeneratedForeignHeaderNamesAreAllocated TestObligationCommentsAreGeneratedFromJSON TestConformanceUnitAssertsEveryField` -- pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./... -count=1` -- all packages pass (run after each task and again after Task 4's mutation restoration)
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` -- exit 0, no diagnostics
- `sh scripts/verify-phase4.sh` -- exit 0, all 32 lane entries `"status":"pass"`
- `env GOCACHE=/tmp/ai-lang-phase4-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestPhase4ReachabilityRecordIsComplete TestVerifyPhase4ControlsAndWork` -- pass
- `git diff --stat -- testdata/` -- empty after every task; every pre-existing golden byte-frozen
- `git diff --stat -- testdata/ scripts/ native/` -- empty after Task 4
- `git diff .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` -- zero deleted lines, one row appended, no hunk in the Reachability Register section
- `git status --short` after Task 4's mutation-kill work -- clean relative to this plan's own committed changes; no leftover reverted hunk

## TDD Gate Compliance

This is a `type: execute` plan whose Task 1 is `type="tracer" tdd="true"` and whose Tasks 2-3 are `type="auto" tdd="true"`, not a `type: tdd` plan, so plan-level RED/GREEN/REFACTOR commit-gate enforcement (a separate `test(...)` then `feat(...)` commit) does not structurally apply. Per-task, the plan's own ordering instruction ("write the test(s) first and observe them go red... then land the [fix] and observe green") was followed as an in-session verification discipline for all three TDD tasks: RED was genuinely captured verbatim for each (Task 1 above; Task 2's and Task 3's hostile-case failures shown identically in "TDD RED Evidence"), and only then was each production fix landed, with tests and fix committed together in each task's single `feat(04-13)` commit -- matching this project's established `type: execute` convention (04-11/04-12-SUMMARY.md's identical precedent). No gate violation: RED was genuinely observed and recorded before GREEN, per task.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- 04-VERIFICATION.md truth 2b's four `missing:` items are all closed: item 1 (corevalidate shape check for the three policy values, new code, before cgen) by Task 2; item 2 (source-admission diagnostic in `collectForeignSymbols`) by Task 1; item 3 (cgen's own independent refusal on the fields it is about to splice) by Task 3; item 4 (falsifiers in `corevalidate_test.go`/`cgen_test.go` for each field plus the end-to-end `.lang`-source falsifier) by Tasks 1, 2 and 3 together.
- Re-running `/gsd-verify-work` on Phase 04 should flip truth 2b from FAILED to verified, closing the last open must-have from the fourth verification round.
- This was the fifth-round gap-closure plan for Phase 4, and 04-13 was the only remaining incomplete plan. With this gap closed, Phase 4 is ready for the next re-verification pass.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*

## Self-Check: PASSED

- FOUND: .planning/phases/04-fallible-resources-and-c-boundary/04-13-SUMMARY.md
- FOUND: testdata/phase4/foreign_policy_value_injection.lang
- FOUND commit ee73eb3 (Task 1: feat)
- FOUND commit 0962f32 (Task 2: feat)
- FOUND commit 0fbc002 (Task 3: feat)
- FOUND commit ae62ae8 (Task 4: docs)
