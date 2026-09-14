---
phase: 07-calls-signatures-and-call-graph-refusal
reviewed: 2026-09-09T00:00:00Z
depth: standard
files_reviewed: 44
files_reviewed_list:
  - cmd/lang/main.go
  - cmd/lang/main_test.go
  - internal/compiler/ast/ast.go
  - internal/compiler/callgraph/callgraph.go
  - internal/compiler/callgraph/callgraph_export_test.go
  - internal/compiler/callgraph/callgraph_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_exclusive_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_internal_test.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_call_type_internal_test.go
  - internal/compiler/corevalidate/corevalidate_closure_chain_mutation_test.go
  - internal/compiler/corevalidate/corevalidate_cycle_peer_test.go
  - internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go
  - internal/compiler/corevalidate/corevalidate_mutation_matrix_test.go
  - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/corevalidate/export_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/originvalidate/export_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_closure_chain_test.go
  - internal/compiler/originvalidate/originvalidate_internal_test.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_phase6_pin_test.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/session/session_phase7_export_test.go
  - internal/compiler/session/session_phase7_mutation_test.go
  - internal/compiler/session/session_phase7_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - scripts/verify-phase7.sh
findings:
  critical: 4
  warning: 3
  info: 3
  total: 10
status: issues_found
---

# Phase 07: Code Review Report

**Reviewed:** 2026-09-09T00:00:00Z
**Depth:** standard
**Files Reviewed:** 44
**Status:** issues_found

## Summary

**Prior CR-01 is genuinely closed.** I re-derived it empirically against a
freshly built binary rather than trusting the plan artifacts: the exact
program from the previous review (`identity(value: Byte) -> Byte` called with
a `Buffer`) is now refused with `check.call_argument_type_mismatch`, the
`OpCall` target's `TypeID` is derived from `calleeContract.ReturnType`
resolved against the caller's own type fact (`check.go:1885-1902`), and
`corevalidate.checkCallTypeContract` (`corevalidate.go:2119-2137`) is a
genuinely independent peer — it lives in `corevalidate`, is consulted from
both replay arms, reads `callee.Parameter.Type`/`callee.ReturnType` off this
program's own `functionByID`, resolves the argument/target constructors by
walking `places`/`types`, shares no symbol or constant with `check`, and its
import-independence is statically enforced
(`corevalidate_test.go:135`, `corevalidate_exclusive_test.go:216`). Both
sides fail closed on an empty constructor or contract string, both carry
isolated fault-injection seams, and the four new codes carry the ordered
`Causes` shapes their doc comments claim. `go vet ./...` and `go test ./...`
(including `-race`) are green. D-07-46's constructor-string granularity is a
disclosed cut and is not re-litigated here.

**The gap-closure was scoped to the TYPE half of the call contract only, and
the OWNERSHIP half of the same call site was never wired at all.** That is
where this review's findings are. `resolveCallBinding` produces a
`core.OpCall` whose `SourceID` is the argument place and then, at both call
sites that use it (`analyzeStraightLine` check.go:2519-2535 and
`analyzeArmBody` check.go:1346-1360), `continue`s **before** the binding
switch that enforces every ownership law in this package. The argument is
never move-marked, never copy-ability-checked, and never registered as a use
of any loan it carries. Three separately-reproducible admissions follow, each
confirmed against a built binary with a matched non-call control that IS
refused:

1. The same non-copyable `Buffer` is passed by value to two calls — accepted
   by check *and* by `corevalidate` (CR-01).
2. A `take` of an owner whose live loan's only later use is a call argument —
   accepted by check; `corevalidate` catches it, but `lang check` never runs
   `corevalidate`, so the CLI reports `pass`, exit 0 (CR-02, CR-04).
3. A caller of a fallible / foreign-reaching callee publishes
   `"fails": ""` and `"foreign": {}` in its own `lang.interface/1` summary,
   directly contradicting `core.FunctionSignature.Foreign`'s own documented
   "closure-derived worst-case foreign reach" — and both the producer and its
   peer share the blind spot (CR-03).

None of these appear in PHASE-07-DEBT.md. D-07-07 defers "multi-argument loan
interaction", which reads as a claim that the single-argument case IS handled;
it is not. Every reproduction below is a two-file A/B: the call-shaped program
and the structurally identical non-call program the checker already refuses.

## Critical Issues

### CR-01: A call neither moves nor copy-checks its argument — the same non-copyable `Buffer` can be passed by value to two calls

**File:** `internal/compiler/check/check.go:2519-2535` (`analyzeStraightLine`'s
call arm) and `internal/compiler/check/check.go:1346-1360` (`analyzeArmBody`'s
call arm); the bypassed law is `internal/compiler/check/check.go:1458-1470` /
`2635-2650` (`ownership.transfer_requires_take`) and the `take` arm's
move-marking at `2566-2570`.

**Issue:** Both call arms run `resolveCallBinding`, append the target place,
and `continue` — skipping the entire binding switch below, which is the only
place `ownership.transfer_requires_take` (non-copyable value used without
`take`), move-marking (`source.initialized = false`), and
`ownership.move_while_borrowed` are enforced. `resolveCallBinding` itself
checks arity, argument scope, prior-move state, callee resolution, and (since
07-09) types — but never ownership transfer.

`core.ParameterContract.Mode` is published unconditionally as `"owned"`
(`originvalidate.go:669`, "today's grammar has exactly one parameter form
(by-value)"), so a call transfers ownership of its argument to the callee.
`Buffer` does not carry `core.AbilityCopy`. Confirmed against a built binary:

```
module test.useafter
export { fn main }
fn identity(value: Buffer) -> Buffer { value }
fn main(buffer: Buffer) -> Buffer {
  let first = identity(buffer)
  let second = identity(buffer)
  second
}
```
→ `lang check`: **pass, exit 0**. `corevalidate.Validate`: **valid: true**.
The emitted core has two `kind=call` operations with the identical
`SourceID = ...fn:main:place:0`.

Matched control, same non-copyable value, non-call binding:
```
fn main(buffer: Buffer) -> Buffer {
  let view = borrow mut buffer
  let copy1 = buffer
  ...
}
```
→ refused: `ownership.transfer_requires_take` — "noncopyable binding requires
explicit take".

Under either reading of call-by-value this is unsound: if the parameter is a
copy, the copy-ability gate must fire (it does for every other binding kind);
if it is a move, use-after-move must fire. Both peers admit it, so the
two-derivation safety net gives zero protection here — the same structural
failure 07-VERIFICATION.md named for the type contract, repeated on the
ownership contract. `interp`/`cgen` refuse `OpCall` today (D-07-39), so this
is not yet a live double-free; it becomes one the moment `OpCall` is lowered
on top of core IR that is already admitted and digested.

**Fix:** In `resolveCallBinding`, after the argument resolves and the type gate
clears, apply the same ownership law the binding switch applies. Concretely,
mirror the `take`/default arms:

```go
// after the argument-type gate, before building the op:
if !hasTypeAbility(typeFact, core.AbilityCopy) {
    causes := []diagnostic.Cause{
        {Kind: "declared_here", Span: spanPointer(argument.declared)},
        {Kind: "missing_ability", Detail: missingAbilityDetail(typeFact, core.AbilityCopy)},
        {Kind: "place", Detail: argument.place.ID},
        {Kind: "type", Detail: argument.place.TypeID},
    }
    diag := diagnostic.Error("ownership.transfer_requires_take",
        binding.RHS.Span, "noncopyable call argument requires explicit take", causes...)
    return core.LinearOperation{}, core.Place{}, &diag
}
```
(or, if the intended semantics is a move, mark
`argument.initialized = false; argument.movedAt = spanPointer(binding.RHS.Span);
argument.moveTargetID = target.ID` and require an explicit `take` at the call
site). Add the identical, independently-implemented predicate to
`corevalidate`'s two `OpCall` replay arms — it already tracks
`initialized`/`produced` per place and has the type facts to read the copy
ability from. Ship a `testdata/phase07/call_argument_not_copyable.lang`
fixture and assert refusal in `check_test.go` and `corevalidate_test.go`, plus
a control lane entry with a mutation kill (a seam that skips the new gate must
make the fixture wrongly admit).

### CR-02: A call argument is invisible to the loan-liveness law — `ownership.move_while_borrowed` is silently defeated by wrapping the loan's last use in a call

**File:** `internal/compiler/check/check.go:2915` (`sourcePlaceID, ok :=
visible[binding.RHS.Source]`) and `internal/compiler/check/check.go:2926-2937`
(the binding-kind switch inside `computeLoanLastUses`, which has no `"call"`
case).

**Issue:** `computeLoanLastUses` is documented as "D-05-35(d)'s **sole**
liveness law … now the ONLY law computing loan expiry for admission anywhere
in this package". It rebuilds a synthetic operation stream from
`body.Bindings`, reading each binding's source from `binding.RHS.Source`. A
call binding stores nothing in `RHS.Source` — its argument is in
`RHS.Arguments[0]` (`parser.go:422-425`). So the lookup misses, the synthetic
source becomes `shadow:place:unknown:N` (check.go:2916-2924), and the switch's
missing `"call"` case leaves `kind = core.OpCopy`. The result: **a call never
counts as a use of anything.** Every loan whose live use is a call argument is
computed as already dead, and `ownership.move_while_borrowed` never fires.

Confirmed against a built binary. Offending program:
```
module test.lk5
export { fn main }
fn identity(value: Byte) -> Byte { value }
fn main(value: Byte) -> Byte {
  let view = borrow value
  let taken = take value
  let used = identity(view)   // loan used AFTER the move
  taken
}
```
→ `lang check`: **pass, exit 0**, zero diagnostics.

Matched control — byte-identical except the loan's post-move use is a plain
binding instead of a call:
```
fn main(value: Byte) -> Byte {
  let view = borrow value
  let taken = take value
  let used = view
  taken
}
```
→ refused: `ownership.move_while_borrowed` — "cannot transfer ownership while
a future-used shared loan is live".

`corevalidate` *does* catch the offending program
(`core.move_while_borrowed` on `...fn:main:op:1`), so the peer is sound here
— but see CR-04: `lang check` never consults it, so the user-facing verdict is
`pass`. Note also that this package now has **two** loan-endpoint laws that
disagree about calls: the real one (`aliasFactEndpoints`, check.go:2801,
running over real operations including `OpCall`) sees the call; the shadow one
(`computeLoanLastUses`) does not. That is precisely the divergence
D-05-35(b)/(c)'s shadow-mode migration discipline was built to forbid, and no
new divergence run was recorded when `"call"` was added as a binding kind.

**Fix:** Add the missing case and source resolution:

```go
sourceName := binding.RHS.Source
if binding.RHS.Kind == "call" && len(binding.RHS.Arguments) == 1 {
    sourceName = binding.RHS.Arguments[0]
}
sourcePlaceID, ok := visible[sourceName]
...
switch binding.RHS.Kind {
case "take":
    kind = core.OpMove
case "borrow":  ...
case "borrow_mut": ...
case "call":
    kind = core.OpCall   // a real use of its argument place
}
```
Then re-run the D-05-35(e) shadow-vs-real divergence comparison over
`TestOwnershipSequenceExhaustive`/`TestBranchSequenceExhaustive` extended with
call bindings, and add the `lk5`-shaped fixture above (plus its non-call
control) to `testdata/phase07/` with both asserted in `check_test.go`. This
defect is a strong mutation-kill candidate: the `case "call"` removal is
exactly the seeded fault the control lane should already have.

### CR-03: `FunctionSignature.Foreign` and `.Fails` are published as "closure-derived worst-case" but are computed purely locally — a caller of a fallible, libc-reaching callee publishes "infallible, no foreign reach"

**File:** `internal/compiler/originvalidate/originvalidate.go:686-696`
(producer) and `internal/compiler/corevalidate/corevalidate.go:1920-1928`
(peer — identical blind spot); contract violated at
`internal/compiler/core/core.go:217-221` and `core.go:287-291`.

**Issue:** `core.FunctionSignature.Foreign`'s own doc comment states it is
"this function's **closure-derived** worst-case foreign reach (R-01)", and
`core.ForeignReach`'s says the same. `Fails`'s says `""` means infallible.
Both are computed as `if function.ForeignContract != nil { ... }` — a strictly
local read, with no propagation across `OpCall` edges. Phase 07 is the phase
that introduced those edges and *did* chain `ClosureDigest` across them
(07-08, `BuildInterface`'s second pass), so the closure machinery is present
and simply was not applied to these two fields.

Confirmed against a built binary. `lang interface export` on:
```
module test.callfallible
foreign C { fn probe(request: Byte) -> Byte { unwind: forbidden
  nonlocal_exit: forbidden  allocator: "libc_malloc"  fails: ProbeError } }
data ProbeError = | ProbeFailed
fn tracer(value: Byte) -> Byte { let result = try probe(value)  result }
fn main(value: Byte) -> Byte { let out = tracer(value)  out }
```
produces:
```
tracer  fails='ProbeError'  foreign={allocator:libc_malloc, unwind:forbidden, nonlocal_exit:forbidden}  callable=true
main    fails=''            foreign={allocator:'', unwind:'', nonlocal_exit:''}    callable=true
```
`main` transitively reaches `libc_malloc` under `unwind: forbidden` and can
transitively fail with `ProbeError`, yet publishes the zero value that
`core.go:406-407` documents as "a LEGAL value meaning **no foreign reach**". A
body-blind consumer — exactly the consumer `lang.interface/1` exists for — acts
on a false claim. This is the `escape:coordinated-source-to-core-false-claim`
class the phase's own threat register names, and, unlike D-07-33, it is
undeclared: `PHASE-07-DEBT.md` contains no entry for `Foreign`/`Fails`
closure derivation, and `corevalidate`'s peer reproduces the producer's error
verbatim, so the bilateral check certifies the false claim rather than
catching it.

**Fix:** Propagate both fields over the call graph in `BuildInterface`'s
existing second pass — the pass already walks `chainOrder` (callee before
caller) precisely because the graph is a proven DAG at that point:

```go
for _, functionID := range chainOrder {
    ...
    for _, calleeID := range calleeIDsForClosureDigest(function) {
        callee := summary.Functions[indexByID[calleeID]]
        summary.Functions[index].Foreign = joinForeignReach(
            summary.Functions[index].Foreign, callee.Foreign)
        if summary.Functions[index].Fails == "" {
            summary.Functions[index].Fails = callee.Fails
        }
    }
    // ClosureDigest computed AFTER the join, so it covers the joined facts
}
```
with `joinForeignReach` defined as an explicit worst-case lattice join
(non-empty beats empty; `forbidden` vs `permitted` resolved to the strictly
more constraining value) rather than a last-writer-wins copy. Mirror the join
independently in `corevalidate.derivePeerSignature`. Alternatively, if
closure-derivation is genuinely out of Phase 07's scope, that is a
declaration-not-implication obligation: amend `core.go`'s doc comments to say
these fields are function-local this phase and add a `D-07-49` entry to
`PHASE-07-DEBT.md` — but a field whose published semantics contradicts its
declared semantics cannot ship silently either way.

### CR-04: `lang check` reports `pass` (exit 0) for programs `corevalidate` refuses — the peer's refusals never reach the user-facing gate

**File:** `internal/compiler/session/session.go:661-683` (`CheckCommandFile`)

**Issue:** `CheckCommandFile` runs `CheckFile` and then
`originvalidate.ValidatePublished`, but never `corevalidate.Validate`. Every
other consuming path does (`RunInterpreter` session.go:693,
`InterfaceExportCommandFile` :1041, `InterfaceCoreCommandFile` :1087). So
whenever check and its peer disagree, the primary CLI verdict is check's — and
a `corevalidate` refusal is invisible unless the user happens to run
`interface core`, where it surfaces not as an invalid-source diagnostic but as
`tool.operation_failed` / "unable to write core artifact", exit 3, with the
actual code (`fmt.Errorf("core validation failed: %s", ...)`) swallowed by
`cmd/lang/main.go:289`.

Confirmed on two independent programs:

- CR-02's `lk5.lang` — `lang check` **pass, exit 0**; `corevalidate.Validate`
  returns `core.move_while_borrowed` on `...fn:main:op:1`.
- Duplicate function declarations (see WR-01) — `lang check` **pass, exit 0**;
  `corevalidate.Validate` returns `core.duplicate_function_id`;
  `lang interface core` reports `tool.operation_failed`, exit 3, no code.

This is not merely a missing convenience: the phase's whole two-peer argument
is that a defect one derivation misses the other catches. That argument only
holds if the peer's verdict is actually surfaced on the gate users run.
Phase 07 makes it acute because `OpCall` is the first construct where the two
derivations can disagree on an ownership fact.

**Fix:** Run the peer on the `check` path and report its refusal as an invalid
source, not an operational failure:

```go
} else if validated := corevalidate.Validate(checked.Program); !validated.Valid {
    result.Status = protocol.StatusInvalid
    result.Diagnostics = []diagnostic.Diagnostic{
        diagnostic.Error(validated.Problems[0].Code, diagnostic.Span{}, validated.Problems[0].Detail)}
} else if problems := originvalidate.ValidatePublished(...); ...
```
Separately, in `InterfaceCoreCommandFile`/`InterfaceExportCommandFile`, stop
collapsing a validation refusal into `error` — return a
`protocol.StatusInvalid` result carrying `validated.Problems[0].Code` so
`cmd/lang/main.go` no longer reports a real source defect as
`tool.operation_failed`. Add a regression asserting that any program
`corevalidate` refuses also makes `lang check` exit non-zero.

## Warnings

### WR-01: Duplicate function declarations are never diagnosed; `buildCalleeContracts` silently resolves a call to the LAST declaration of a shadowed name

**File:** `internal/compiler/check/check.go:520-530`
(`buildCalleeContracts`), `internal/compiler/check/check.go:113`
(`functionID := semanticID(program.Module, "fn", function.Name)`)

**Issue:** Nothing in `check` refuses two `fn` declarations with the same name.
Both get the identical `semanticID`, so the emitted `core.Program` carries two
`core.Function` entries with the same `ID` (`corevalidate` refuses this with
`core.duplicate_function_id`, but see CR-04 — `lang check` still says `pass`).
The duplicate-name hazard predates Phase 07, but Phase 07 gives it new teeth:
`buildCalleeContracts` is a `map[string]calleeContract` keyed by function name
and populated in declaration order, so a call resolves to the *last*
declaration's signature while the emitted `CalleeID` names an ID shared by
both. Confirmed: given `fn helper(v: Byte) -> Byte` followed by
`fn helper(v: Buffer) -> Buffer`, a `Byte`-typed call to `helper` is refused
with `check.call_argument_type_mismatch` against the *second* declaration —
the first is silently unreachable, and the callee identity the call graph and
`ClosureDigest` chain record is ambiguous.

**Fix:** Refuse duplicates at their declaration site, before any contract table
is built:

```go
declared := make(map[string]diagnostic.Span, len(program.Funcs))
for _, function := range program.Funcs {
    if first, dup := declared[function.Name]; dup {
        result.Diagnostics = append(result.Diagnostics, diagnostic.Error(
            "name.duplicate_function", function.Span, "function name is already declared",
            diagnostic.Cause{Kind: "declared_here", Span: &first}))
        continue
    }
    declared[function.Name] = function.Span
}
```
Add a `testdata/phase07/duplicate_function_name.lang` fixture.

### WR-02: `core.callee_not_callable` reports the whole calling function's span, not the call site's, even though the call-site span is already in hand

**File:** `internal/compiler/check/check.go:675-683`

**Issue:** `verifyCallableRefusal` emits
`diagnostic.Error(core.CalleeNotCallable, function.Span, ...)` — the span of
the *entire calling function declaration*. `check.Program` already builds
`spanByOperationID` (check.go:139-146, threaded from `callSpans`) and
`checkCallGraphAcyclic` uses it to project a precise per-edge span for every
`cycle_member` cause. The single most actionable span for this refusal — the
call site — is available and unused. In a function with several calls, the
user is pointed at the function header and must guess which call was rejected.
`verifyCallableRefusal` also returns on the first offending call, so a program
with two uncallable targets surfaces them one re-run at a time (unchanged from
the previous review's WR-02; still worth deciding deliberately).

**Fix:** Thread `spanByOperationID` into `verifyCallableRefusal` and use
`spanByOperationID[operation.ID]` as the Primary span, falling back to
`function.Span` when absent, exactly as `checkCallGraphAcyclic` does.

### WR-03: `callArguments` has no length bound, unlike every other repeated-list parse in this package

**File:** `internal/compiler/syntax/parser.go:501-515`

**Issue:** Still open from the previous review. `callArguments` loops over
comma-separated identifiers with no cap, while every sibling list parse
declares one and refuses fail-closed above it: `maxArmsPerMatch` (parser.go:521),
`maxForeignSymbolsPerBlock`/`maxForeignPoliciesPerSymbol` (:180-183),
`maxDeclarations`/`maxFunctions`/`maxAlternatives`/`maxLinearBindings`
(:15-18). I verified the practical exposure is smaller than the previous review
implied — `MaxSourceBytes` (1 MiB) and `MaxTokens` (131072) bound the token
stream transitively, and no infinite-loop path exists (`expect` refuses to
advance only on EOF/`}`/declaration boundaries, each of which breaks the loop).
So this is a convention/defense-in-depth gap, not an exploitable DoS —
downgraded from the previous review's WARNING framing but still the one
unbounded repeated-list parse in the file.

**Fix:** Add `const maxCallArguments = 8` beside `maxArmsPerMatch` and refuse
with `syntax.call_argument_limit` above it, mirroring `matchExpr`'s shape
(parser.go:537-541).

## Info

### IN-01: `indexByID` is documented as indexing `summary.Functions` but is then used to index `program.Functions`

**File:** `internal/compiler/originvalidate/originvalidate.go:717` (comment and
population) vs. `originvalidate.go:744` (`function := program.Functions[index]`)

**Issue:** The map is built as `indexByID[function.ID] = len(summary.Functions)`
with a doc comment explaining it exists so the second pass can "mutate
`summary.Functions[index]` directly, by index". Line 744 then uses the same
index against `program.Functions`. It is correct today only because the first
loop appends exactly one signature per program function with no `continue` — an
invariant nothing enforces. Adding any skip to that loop silently
mis-associates every function after the skip with another function's callee
set and `ClosureDigest`.

**Fix:** Either keep a second `programIndexByID`, or (simpler) iterate
`chainOrder` against a `map[string]core.Function` built from
`program.Functions`, so the two slices are never index-coupled.

### IN-02: `check`'s argument-type gate compares the caller's function-level type fact, not the argument place's own type

**File:** `internal/compiler/check/check.go:1863`
(`typeFact.Shape.Constructor != contract.ParameterType`)

**Issue:** The gate compares `typeFact` — the calling function's single type
fact — rather than resolving `argument.place.TypeID`. This is sound *only*
because `sameType` forces one type fact per function today (`checkLinear`:1533,
`checkBranch` via a single `linear.Types` entry, `checkFallibleLinear`:1784).
The peer (`corevalidate.checkCallTypeContract`) does the more robust thing —
it resolves `types[source.TypeID]` from the argument place. The moment a
function carries two type facts, check's gate compares the wrong type while
the peer compares the right one, and the two derivations diverge silently in
the *permitting* direction on the check side. The doc comment explains the
constraint but the code does not encode it.

**Fix:** Resolve the argument's own type: pass the function's `[]core.TypeFact`
(or a `map[string]core.TypeFact`) into `resolveCallBinding` and compare
`types[argument.place.TypeID].Shape.Constructor`, keeping the current
`typeFact` only as the return-derivation source. Behaviour is identical today;
the invariant stops being load-bearing.

### IN-03: `check.Program` clears `result.Program` on a call-graph cycle but leaves `result.AliasFacts` populated

**File:** `internal/compiler/check/check.go:373` (`result.Program = core.Program{}`)

**Issue:** The cycle gate deliberately destroys the cyclic `core.Program` so it
"never [exists] returned, serialized, cached, interpreted, or lowered". The
alias facts derived from that same program (appended at check.go:158) survive
on the returned `CheckResult`. No production consumer reads `AliasFacts` today
(only `cgen`/`corevalidate` doc comments and `session_phase5_alias.go`
reference the concept), so this is latent rather than live — but it is a
half-cleared result whose stated invariant is "nothing from a cyclic program
escapes".

**Fix:** Add `result.AliasFacts = nil` alongside the program clear, and say so
in the comment.

---

_Reviewed: 2026-09-09T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
