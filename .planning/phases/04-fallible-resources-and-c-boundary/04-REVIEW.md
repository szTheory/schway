---
phase: 04-fallible-resources-and-c-boundary
reviewed: 2026-09-05T20:51:14Z
depth: standard
files_reviewed: 48
files_reviewed_list:
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/evidence/evidence.go
  - internal/compiler/evidence/evidence_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interptestdirect/interptestdirect.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/native/foreign_resource.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_conformance_test.go
  - internal/compiler/native/native_test.go
  - internal/compiler/native/symbols.go
  - internal/compiler/native/symbols_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - native/lang_foreign_nonlocal.c
  - native/lang_foreign_resource.c
  - native/lang_foreign_resource_private.h
  - scripts/verify-phase4.sh
  - .github/workflows/ci.yml
  - .gitignore
  - testdata/phase4/acquire_three_fail_second.lang
  - testdata/phase4/acquire_three_fail_third.lang
  - testdata/phase4/acquire_three_success.lang
  - testdata/phase4/defect_terminal.lang
  - testdata/phase4/discard_because.lang
  - testdata/phase4/fallible_call_unconsumed.lang
  - testdata/phase4/foreign_acquire_one.lang
  - testdata/phase4/foreign_call_target_not_foreign.lang
  - testdata/phase4/foreign_layout_mismatch.golden.c
  - testdata/phase4/foreign_origin_omitted.lang
  - testdata/phase4/foreign_unwind_undeclared.lang
  - testdata/phase4/nonlocal_exit_probe.lang
findings:
  critical: 2
  warning: 1
  info: 1
  total: 4
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-09-05
**Depth:** standard
**Files Reviewed:** 48
**Status:** issues_found

## Summary

This is a re-review after the third round of gap closure (plan 04-10, commits f875062..97a73a8), which added a visited-set cycle guard to `checkReleaseOrder`'s `rederive` backward walk in `internal/compiler/corevalidate/corevalidate.go` and two falsifiers (`TestCyclicOkEdgeChainRefusedNotHung`, `TestAcyclicChainsStillValidateUnderCycleGuard`) in `corevalidate_test.go`.

The prior review's CR-01 (unbounded loop / DoS hang on a cyclic `"ok"`-edge chain) is verified fixed: the `visited` map at `corevalidate.go:1252-1258` correctly refuses re-entry into an already-visited block with `core.release_order_cyclic` rather than looping, the guard is allocated fresh per call to `rederive` (no cross-call state leakage), and the two new falsifiers correctly exercise both the refusal path (with a bounded-wall-clock assertion proving it does not hang) and the accepting path (proving the guard costs nothing on every existing acyclic fixture, pinned at an unchanged `Checks` count). This closes the gap cleanly.

However, this pass found two new critical issues that survived (or were newly exposed by) the full-scope review, plus one warning and one info item carried forward as still-open observations:

1. `checkReleaseOrder`'s own fix for merge points only covers a terminal block's *own direct* incoming edges (the CR-01-from-the-prior-round fix, `blocksAndEdges`/the per-incoming-edge loop). The single-edge `okEdgeInto` map used to hop backward through *interior* (non-terminal) blocks still silently collapses to one arbitrary edge when more than one `"ok"`-pattern edge targets the same interior block — a distinct, untested soundness gap from the fixed hang.
2. `corevalidate` never checks that `core.ForeignContract.Symbol` is a syntactically safe C identifier before `cgen.Emit`/`cgen.EmitNative` splice it, unsanitized, directly into generated C source as an `extern` declaration and as a call-expression callee — a code-injection gap in the one boundary explicitly designed to defend against a corrupted, non-source-derived `core.Program`.

## Critical Issues

### CR-01: `checkReleaseOrder`'s backward walk silently collapses multiple incoming `"ok"` edges into the SAME interior block, unlike the already-fixed handling for a terminal block's own incoming edges

**File:** `internal/compiler/corevalidate/corevalidate.go:1205-1213` (the `okEdgeInto` map construction) and `:1260-1271` (`rederive`'s use of it)
**Issue:**

```go
edgesByTo := make(map[string][]core.Edge, len(linear.Edges))
okEdgeInto := make(map[string]core.Edge, len(linear.Edges))
for _, edge := range linear.Edges {
    edgesByTo[edge.ToBlockID] = append(edgesByTo[edge.ToBlockID], edge)
    if edge.Pattern == "ok" {
        okEdgeInto[edge.ToBlockID] = edge   // last "ok" edge into this block wins; any earlier one is silently discarded
    }
}
...
edge, ok := okEdgeInto[currentBlockID]   // rederive's backward hop through an INTERIOR block
if !ok {
    break
}
currentBlockID = edge.FromBlockID
```

The 04-08/04-10 fix correctly closed the gap for a *terminal* block's own incoming edges: the outer loop in `checkReleaseOrder` (corevalidate.go:1284-1319) fetches `incoming := edgesByTo[block.ID]` — every edge, from `edgesByTo`, not just one — and walks `rederive` independently for each one, comparing every result against the block's single fixed release list (`TestMergeTerminalBlockDivergentReleaseSetsRefused` proves this). But `rederive`'s own backward hop through every *interior* block it passes through on the way back to the entry uses `okEdgeInto`, a `map[string]core.Edge` (singular, not a slice) keyed by `ToBlockID` that keeps only the last `"ok"`-pattern edge seen for a given target block during construction. If a hand-corrupted `core.Program` declares two `"ok"`-pattern edges from two different blocks A and A′ into the SAME interior (non-terminal) block T — where A and A′ represent genuinely different completed-acquisition histories — `okEdgeInto[T]` silently keeps only one of them (whichever appears last in `linear.Edges`' iteration order), and the backward walk from any terminal block reachable through T only ever rederives the history through the kept edge. The path through the discarded edge is never independently rederived or compared at all, so a corrupted program whose T-via-A′ path acquires/releases a different set of resources than T-via-A silently passes as long as the *kept* path happens to agree with the terminal block's declared release list.

This is materially the same class of gap CR-01 in the prior review closed for terminal blocks (silently skipping a merge point's disagreeing path instead of independently checking every path into it), just one level removed — it is the identical `okEdgeInto` map that motivated `blocksAndEdges`'/the CR-01 fix's own explanatory comment ("What this rederivation cannot tolerate is SKIPPING the check for a merge point... checking against every incoming edge, not skipping the block, is what closes the gap") but that fix was applied only at the outer per-terminal-block loop, never inside `rederive`'s own interior-block hop. No fixture or falsifier in `corevalidate_test.go` constructs two `"ok"` edges converging on a *non-terminal* block (`TestMergeTerminalBlockDivergentReleaseSetsRefused`/`TestMergeTerminalBlockAgreeingChainsAccepted` both add the second edge directly into the terminal `success` block, never into an interior `step:N` block), so this gap is untested and unfixed.

check.go's own honest lowering (`checkResourceLifecycle`) never produces two `"ok"`-pattern edges into the same block (each step's single successor is either the next step's block or the success block, and a `discard`'s two edges into the same target use patterns `"ok"` and `"err"`, never `"ok"`/`"ok"`), so this cannot be triggered by the compiler's own pipeline today. It is exactly the shape corevalidate's own stated purpose — defending against a corrupted or adversarially hand-constructed `core.Program`, independent of whether check.go could have produced it — exists to catch, and it does not.

**Fix:** Change `okEdgeInto` to `map[string][]core.Edge` and, inside `rederive`, when more than one `"ok"` edge targets the current block, walk each one independently (recursively) and require they all rederive an identical accumulated release set before continuing — the same "every path must agree" principle already applied one level out for terminal blocks:

```go
okEdgesInto := make(map[string][]core.Edge, len(linear.Edges))
for _, edge := range linear.Edges {
    edgesByTo[edge.ToBlockID] = append(edgesByTo[edge.ToBlockID], edge)
    if edge.Pattern == "ok" {
        okEdgesInto[edge.ToBlockID] = append(okEdgesInto[edge.ToBlockID], edge)
    }
}
...
candidates := okEdgesInto[currentBlockID]
if len(candidates) == 0 {
    break
}
if len(candidates) > 1 {
    // every incoming "ok" edge into this interior block must independently
    // rederive the SAME accumulated history before the walk can trust any one of them
    var first []core.LinearOperation
    for i, candidate := range candidates {
        sub, ok := rederive(candidate) // requires refactoring rederive to be safely re-entrant / iterative-with-explicit-stack
        if !ok {
            return nil, false
        }
        if i == 0 {
            first = sub
        } else if !reflect.DeepEqual(first, sub) {
            v.check(false, "core.release_order_ambiguous_merge", currentBlockID)
            return nil, false
        }
    }
    return append(expected, first...), true
}
edge := candidates[0]
currentBlockID = edge.FromBlockID
```

Add a falsifier constructing two `"ok"`-pattern edges into a shared interior `step:N` block from two blocks with different completed-acquisition histories, asserting `Validate` refuses it (rather than silently accepting via the arbitrarily-kept edge), plus a companion accepting-path test where both incoming edges genuinely agree.

### CR-02: `corevalidate` never validates `ForeignContract.Symbol` as a safe C identifier before `cgen` splices it unsanitized into generated C source — a code-injection gap in the validator's own stated adversarial threat model

**File:** `internal/compiler/corevalidate/corevalidate.go:311` (the only check ever applied to `Symbol`); `internal/compiler/cgen/cgen.go:352-373`, `:441`, `:778`, `:1278-1308`
**Issue:** `corevalidate.linear` requires only that `Symbol` is non-empty:

```go
if !v.check(function.ForeignContract != nil && function.ForeignContract.Symbol != "", "core.foreign_contract_missing", operation.ID) {
    return false
}
```

No check anywhere in `corevalidate.go` requires `Symbol` to match a valid C identifier (e.g. `^[A-Za-z_][A-Za-z0-9_]*$`). `cgen.go` then splices this string directly, unescaped and unvalidated, into generated C source as an identifier in three places:

```go
func foreignExternName(symbol string) string { return "_LANG_" + symbol }
...
symbolC := foreignExternName(function.ForeignContract.Symbol)
resultType := symbolC + "_result"
fmt.Fprintf(&out, "extern %s %s(unsigned char argument);\n\n", resultType, symbolC)
...
fmt.Fprintf(&out, "  %s %s = %s(%s); /* %s */\n", resultType, resultLocal, symbolC, locals[lastOp.SourceID], lastOp.ID)
```

Unlike every other identifier `cgen.go` emits — type names, function names, parameter names, place names — which are routed through `cName`/`cLocal` (both of which strip every character outside `[A-Za-z0-9_]` and replace it with `_`, per their own documented invariant), `Symbol` reaches the output completely raw. A `core.Program` with `ForeignContract.Symbol` set to something like `"real_symbol); } int LANG_pwned(void) { system(\"id\"); return 0"` would pass `corevalidate.Validate` (the only gate `cgen.Emit`/`cgen.EmitNative` run before generating C text — see `cgen.go:16-20` and `:40-45`) and then have that entire string spliced verbatim into the `extern` declaration and the call-expression line, breaking out of the intended single declaration/call and injecting arbitrary top-level C source (including a full function definition with an arbitrary body) into the file this compiler is about to compile and, in `session.RunNative`, execute.

`corevalidate.go`'s own package doc and inline comments repeatedly state its entire reason for existing is to stay defined against "a corrupted or adversarially hand-constructed `core.Program`" reaching `cgen`/`interp`, independent of whether `check.go`'s own parser-level identifier restrictions would ever produce such a `Symbol` (in the honest pipeline, `Symbol` originates from a lexer identifier token and is therefore always identifier-shaped — but `cgen.Emit`/`EmitNative` are exported functions that accept any `core.Program`, and `corevalidate.Validate` is the one documented, source-blind gate standing between an arbitrary caller-supplied `core.Program` and C source generation). No test in `cgen_test.go` or `corevalidate_test.go` constructs a `Symbol` containing a non-identifier character, so this gap is untested.

**Fix:** Add an identifier-shape check to `corevalidate.linear`'s existing `ForeignContract` validation block, alongside the existing non-empty check:

```go
var validCIdentifier = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
...
if !v.check(validCIdentifier.MatchString(function.ForeignContract.Symbol), "foreign.symbol_not_identifier", operation.ID) {
    return false
}
```

placed immediately after the existing `core.foreign_contract_missing` check at corevalidate.go:300-302, so a `Symbol` that is non-empty but not identifier-shaped is refused before `cgen` ever sees it. Add a falsifier in `corevalidate_test.go` that mutates a valid foreign-call program's `ForeignContract.Symbol` to contain a semicolon/parenthesis/newline and asserts refusal with the new code, alongside a regression test in `cgen_test.go` proving `cgen.Emit` is never reached (or, if reached directly bypassing corevalidate, that it independently refuses) with such a `Symbol`.

## Warnings

### WR-01: `TestReleaseOrderValidationWorkSeries`'s monotonic-work assertion does not by itself prove linearity, and does not model the per-incoming-edge cost the 04-08/04-10 fixes introduced

**File:** `internal/compiler/corevalidate/corevalidate_test.go:709-732`
**Issue:** The test's doc comment claims it "proves the validator's release-order rederivation cost is linear in the number of blocks plus edges plus operations," but the test body only asserts the three-point `series` is strictly increasing. A strictly increasing series is consistent with linear, quadratic, or worse growth — it does not distinguish "linear" from "quadratic in incoming-edge-count per terminal block" (the per-edge `rederive` loop introduced by the 04-08 fix) or "quadratic in interior-merge fan-in" (were CR-01 above to be fixed with the recursive-per-edge approach sketched there). This was already flagged in the prior review pass and remains unaddressed.
**Fix:** Either soften the doc comment to describe what is actually checked ("proves the counted work does not regress to a constant, i.e. it grows with fixture size") or strengthen the assertion to fit at least four points and check a ratio bound consistent with linear growth.

## Info

### IN-01: `checkReleaseOrder`'s per-incoming-edge loop makes a single terminal block's check cost proportional to `incoming-edge-count × chain-depth`, not modeled by `LinearWorkLimit`

**File:** `internal/compiler/corevalidate/corevalidate.go:1284-1319`
**Issue:** Carried forward from the prior review as a documentation-only note, not a defect: no current fixture exercises a terminal block reached by 3+ incoming edges, so `LinearWorkLimit`'s facts-based formula and `TestReleaseOrderValidationWorkSeries`'s monotonic check do not currently model this cost dimension. Performance is out of scope for this review.
**Fix:** None required; consider a fixture with 3+ incoming edges into one terminal block if `LinearWorkLimit` is ever asserted against that shape in the future.

---

_Reviewed: 2026-09-05T20:51:14Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
