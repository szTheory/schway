---
phase: 07-calls-signatures-and-call-graph-refusal
reviewed: 2026-09-09T00:00:00Z
depth: standard
files_reviewed: 52
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
  - testdata/phase07/call_basic.lang
  - testdata/phase07/call_from_both_match_arms.lang
  - testdata/phase07/call_uncallable_callee.lang
  - testdata/phase07/clean_but_unpublishable.lang
  - testdata/phase07/cycle_indirect.lang
  - testdata/phase07/cycle_mutual.lang
  - testdata/phase07/cycle_self.lang
  - testdata/phase07/cycle_through_match_arm.lang
  - testdata/phase07/cycle_unreachable.lang
  - testdata/phase07/deep_diamond_acyclic.lang
  - testdata/phase07/foreign_symbol_shadowing.lang
  - testdata/phase07/relay_escort_witness.lang
findings:
  critical: 1
  warning: 2
  info: 1
  total: 4
status: issues_found
---

# Phase 07: Code Review Report

**Reviewed:** 2026-09-09T00:00:00Z
**Depth:** standard
**Files Reviewed:** 52
**Status:** issues_found

## Summary

Phase 07 adds Lang-to-Lang call expressions (`core.OpCall`), a call-graph
acyclicity refusal (`callgraph` package, three-color iterative DFS), a
publication-safety "callable" predicate consulted at every call site
(`verifyCallableRefusal`/`peerCallable`), and closure-digest chaining in
`lang.interface/1`. The engineering around the *refusal* machinery itself —
cycle detection, canonical witness selection, body-blindness at the call
admission site, bounded diagnostics, dual independent derivations in
`check`/`callgraph` vs. `corevalidate` — is careful and heavily
test-instrumented (fault-injection seams, mutation-kill tests, bilateral
divergence tests).

However, one significant correctness gap survived all of that
instrumentation: **a call's argument type is never checked against the
callee's declared parameter type**, in either of the two independent
derivations (`check.resolveCallBinding` and
`corevalidate`'s `OpCall` replay case). This was confirmed empirically — a
program that calls `identity(value: Byte) -> Byte` with a `Buffer` argument
from a `Buffer`-parametered caller is admitted with zero diagnostics. Because
neither `interp` nor `cgen` execute `OpCall` yet this phase (D-07-39), this
cannot yet cause a runtime memory-safety incident, but it is a genuine
soundness hole in the type checker of a project whose entire premise is
"refuse what cannot be proven safe" — and it will become live the moment a
later phase lowers `OpCall` to a real call. This is a BLOCKER.

Two smaller issues are also reported: an unbounded call-argument list at
parse time (inconsistent with this codebase's otherwise-universal
fail-closed list bounds), and a duplicated boolean-flag-count symmetry gap
in `main.go`'s flag parsing that is asymmetric with the codebase's own
stated conventions. Neither is severe.

## Critical Issues

### CR-01: Call admission never checks the argument type against the callee's declared parameter type

**File:** `internal/compiler/check/check.go:1737-1791` (`resolveCallBinding`), mirrored in `internal/compiler/corevalidate/corevalidate.go:1458-1484` (the `core.OpCall` replay case) and `internal/compiler/corevalidate/corevalidate.go:2254-2262` (`targetMatches`)

**Issue:** `resolveCallBinding` resolves a call's callee purely by name
(`functionIDs[binding.RHS.Callee]`) and admits the call as long as the
argument name resolves to an initialized place in the caller's own scope.
It never compares the argument's declared type to the callee function's own
`Parameter.Type`. The resulting `core.OpCall` operation's `TargetID` is
given `TypeID: argument.place.TypeID` — the *caller's* type, not a
re-derivation of the callee's actual return type. `corevalidate`'s
independent replay of `OpCall` (`targetMatches`) only checks that the
target place's `TypeID` matches the *operation's own* `TypeID` (which check
already set to the caller's argument type) — it never cross-references
`functionByID[operation.CalleeID].Parameter.Type` either. Both of the
project's "independent dual derivations" have the identical blind spot, so
the intended two-peer safety net does not catch this.

Confirmed empirically: the following program is accepted by `check.Program`
with zero diagnostics:

```
module test.call_type_mismatch
export { fn main }
fn identity(value: Byte) -> Byte {
  value
}
fn main(buffer: Buffer) -> Buffer {
  let result = identity(buffer)
  result
}
```

No fixture in `testdata/phase07/` exercises a type-mismatched call, and no
test in `check_test.go` (`TestBareCallToDeclaredFunctionAdmitted`,
`TestCallToDeclaredLangCalleeStillAdmitted`, etc.) asserts a mismatched-type
call is refused — every "admitted" test fixture happens to have caller and
callee agree on `Byte`/`Buffer`, which masks the gap.

Because `interp.Run` and `cgen.Emit`/`EmitNative` both explicitly refuse to
execute or lower `core.OpCall` this phase (D-07-39/A-02), this specific
defect cannot yet corrupt a running program or generated C. But it means
the checker — the one place this project's "refusal" guarantee is supposed
to live — currently *type-checks nothing* about a call's argument, which
directly contradicts the project's stated purpose and will become an active
memory-safety/type-safety hole the moment a future phase implements real
`OpCall` execution/lowering on top of today's admitted (but unsound) core
IR.

**Fix:** In `resolveCallBinding`, once `calleeID` resolves, look up the
callee's own declared parameter type (available either from the
in-progress `program.Funcs`/`functionIDs`-adjacent table, or by threading a
`calleeParameterType map[string]core.TypeRef` alongside `functionIDs`) and
refuse with a new causal diagnostic (e.g. `check.call_argument_type_mismatch`)
when `argument.place`'s type constructor does not match. Set the `OpCall`
operation's `TargetID.TypeID` from the callee's own return-type fact, not
the caller's argument type. Mirror the identical check independently in
`corevalidate`'s `OpCall` case (comparing `operation.TypeID`/target type
against `functionByID[operation.CalleeID].Parameter.Type`), so the two
derivations stay independent per this phase's existing convention. Add a
`testdata/phase07/call_type_mismatch.lang` fixture (calling a `Byte`
function with a `Buffer` argument) and assert refusal in both `check_test.go`
and `corevalidate_test.go`.

## Warnings

### WR-01: `callArguments` has no bound on argument-list length, unlike every other repeated-list parse in this codebase

**File:** `internal/compiler/syntax/parser.go:501-515`

**Issue:** `callArguments` loops reading comma-separated identifiers with no
upper bound, in contrast to `matchExpr`'s `maxArmsPerMatch` (parser.go:521,
537) and `check.go`'s `maxForeignBlocksPerProgram`/
`maxForeignSymbolsPerBlockCheck`/`maxForeignPoliciesPerSymbolCheck` — all of
which exist specifically as "fail-closed defense-in-depth against a
CFG-shape/declaration-surface denial-of-service", per this codebase's own
documented convention (see `maxBlocksPerFunction`'s doc comment: "an
unreachable gate must be named as such, not assumed correct" and "bounded
generously... rather than an arbitrary round number"). `resolveCallBinding`
does refuse arity != 1 at check time (`check.call_arity_unsupported`), so
this is not exploitable to defeat arity admission — but a syntactically
huge argument list (e.g. thousands of identifiers) is parsed and fully
materialized into an `ast.Binding` before check ever gets a chance to
reject it, which is exactly the class of unbounded-parse-time-allocation
surface this codebase otherwise treats as a fail-closed bound target.

**Fix:** Add a `maxCallArguments` constant (e.g. 8, well above the arity-1
admission ceiling this phase actually allows) and refuse with a
`syntax.call_argument_limit` diagnostic when exceeded, mirroring
`maxArmsPerMatch`'s shape.

### WR-02: `verifyCallableRefusal`'s single-callee-refusal short-circuit reports only the first offending call, not every offending call

**File:** `internal/compiler/check/check.go:543-597`

**Issue:** `verifyCallableRefusal` returns on the *first* `core.OpCall`
operation whose callee is not callable (`return &diag`), scanning
`functions` and each function's `Linear.Operations` in declaration order.
This matches the existing single-diagnostic-per-gate pattern used
elsewhere in this file (e.g. `verifyCallInvariants`), so it's consistent
with local convention and not a bug per se, but it does mean a program with
two independently-uncallable call targets only ever surfaces one of them at
a time — a user fixing the first-reported callee will re-run the checker
only to hit the second. This is a minor UX/quality concern worth noting
given the project's general emphasis on causal, complete diagnostics
elsewhere (e.g. `checkCallGraphAcyclic`'s bounded-but-complete
`cycle_member` cause list).

**Fix:** Optional — not required for correctness. If diagnostic
completeness is valued here as it is for cycle causes, consider collecting
every uncallable-callee operation into one diagnostic's causes rather than
returning on the first, or documenting explicitly that this is deliberately
single-shot (fail-fast) like `verifyCallInvariants`.

## Info

### IN-01: `checkFallibleLinear`'s foreign-call path and `resolveCallBinding`'s Lang-call path independently duplicate the "argument must be an initialized in-scope place" check with subtly different diagnostics

**File:** `internal/compiler/check/check.go:1737-1791` vs. `internal/compiler/check/check.go:1807-1811` (`resolveForeignStep`)

**Issue:** `resolveCallBinding` (Lang-to-Lang calls) refuses an
out-of-scope argument with `name.unknown` and a moved argument with
`ownership.use_after_move`, while `resolveForeignStep` (foreign calls)
refuses an argument that isn't literally the function's own parameter name
with a flat `name.unknown` (it doesn't distinguish "not the parameter" from
"moved" at all, since Phase 4's foreign-call shape restricts the argument
to always be the parameter). This divergence is scope-appropriate (Phase 07
generalizes argument resolution to arbitrary in-scope places; Phase 4 never
did), so this is not a defect, just worth a comment near
`resolveCallBinding` cross-referencing why the two admission paths differ
in strictness, for a future reader trying to unify them.

**Fix:** Optional — add a one-line comment noting the scope difference is
intentional (Phase 07 arity-1-over-any-place vs. Phase 4
arity-1-over-the-parameter-only) to save a future reader from assuming this
is drift.

---

_Reviewed: 2026-09-09T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
