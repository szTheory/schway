---
phase: 04-fallible-resources-and-c-boundary
reviewed: 2026-09-05T00:00:00Z
depth: standard
files_reviewed: 37
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
findings:
  critical: 1
  warning: 1
  info: 1
  total: 3
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-09-05
**Depth:** standard
**Files Reviewed:** 37
**Status:** issues_found

## Summary

This is the fourth review round, following gap-closure plans 04-11 (interior-merge rederivation) and 04-12 (`ForeignContract.Symbol` C-identifier audit).

Both gaps the previous review round (`04-REVIEW.md`, round 3) identified are verified fixed:

- **Prior CR-01** (`checkReleaseOrder`'s backward walk silently collapsing multiple `"ok"` edges into the same interior block): `okEdgeInto` is now `map[string][]core.Edge` (`corevalidate.go:1222,1226`), and `rederive` (`corevalidate.go:1264-1326`) independently walks every candidate edge into an interior merge with its own copy of `visited`, requiring `sameReleaseHistory` agreement across all of them before trusting any one (`core.release_order_merge_mismatch` on disagreement). `TestInteriorMergeDivergentHistoriesRefused` (both edge-ordering subtests) and `TestInteriorMergeAgreeingHistoriesAccepted` in `corevalidate_test.go:878-994` correctly falsify both the refusal and acceptance paths, including the specific "last-writer-wins map" failure mode the old code had. This closes the gap.
- **Prior CR-02** (`ForeignContract.Symbol` reaching `cgen` unsanitized): `corevalidate.linear` now calls `validCIdentifier` (`corevalidate.go:316`, defined at `corevalidate.go:1639-1654`) immediately after the existing non-empty check, refusing with `foreign.symbol_not_identifier` before `cgen` ever sees a non-identifier `Symbol`. `cgen.go` additionally defines its own peer implementation `validForeignSymbol` (`cgen.go:789-804`) and applies it at `emitLinearForeign` (`cgen.go:326`) and inside `singleForeignFunction` (`cgen.go:1261`), which gates the three exported entry points (`EmitForeignManifest`, `EmitForeignHeader`, `EmitForeignConformance`) that never call `corevalidate.Validate`. The shared `hostileForeignSymbols` table (`cgen_test.go:212-223`) is exercised against both corevalidate's and cgen's independent guards, including the comment-terminator (`*/`) and semicolon-brace cases. This closes the gap for the `Symbol` field.

However, this round's full-scope pass surfaced a **new, more severe injection vector that the 04-12 fix does not cover**: `ForeignContract`'s *other* string-valued fields (`Allocator`, `Unwind`, `NonlocalExit`) are still spliced unsanitized into generated C **comments**, and unlike the already-fixed `Symbol` field, these fields are reachable directly from ordinary Lang source via a string-literal foreign policy value (`allocator: "..."`) — no corrupted `core.Program` is required to trigger it. This is classified as a new Critical finding below. One warning and one info item are carried forward from the prior review as still-unaddressed observations.

## Critical Issues

### CR-01: `ForeignContract.Allocator`/`Unwind`/`NonlocalExit` are never validated for C-comment safety, and are spliced unescaped into `EmitForeignHeader`'s generated comment block — a source-reachable code-injection vector, not merely a corrupted-`core.Program` one

**File:** `internal/compiler/cgen/cgen.go:1332-1340` (the unescaped `Fprintf` splice sites); `internal/compiler/check/check.go:1100-1111` (the admission path that copies the raw policy value with no shape validation); `internal/compiler/syntax/parser.go:227-240` (the string-literal policy-value grammar that admits the payload); `internal/compiler/corevalidate/corevalidate.go:327` (the only check ever applied to these fields — presence only)

**Issue:** The 04-12 fix (`validCIdentifier`/`validForeignSymbol`) closed the injection gap for `ForeignContract.Symbol` specifically, because `Symbol` is the field spliced into the `extern` declaration and call-expression callee. But `EmitForeignHeader` also splices four other flat string fields directly into C **comments**, with no validation anywhere in the pipeline beyond "non-empty":

```go
fmt.Fprintf(&out, "/* symbol: %s */\n", contract.Symbol)          // now guarded (validForeignSymbol)
fmt.Fprintf(&out, "/* allocator: %s */\n", contract.Allocator)    // UNGUARDED
fmt.Fprintf(&out, "/* unwind: %s */\n", contract.Unwind)          // UNGUARDED
fmt.Fprintf(&out, "/* nonlocal_exit: %s */\n", contract.NonlocalExit) // UNGUARDED
```

Unlike `Symbol` — which check.go always resolves from an identifier token (`p.identifier(...)` in `parser.go` for the symbol name) — `Allocator`, `Unwind`, and `NonlocalExit` are populated from an `ast.ForeignPolicy.Value`, which the parser (`parser.go:229-240`) accepts as *either* a bare identifier *or* a double-quoted string literal:

```go
value := p.peek()
isString := value.Kind == TokenString
if isString || value.Kind == TokenIdentifier {
    p.advance()
}
...
text := value.Text
if isString && len(text) >= 2 {
    text = text[1 : len(text)-1]
}
```

The lexer's string-literal rule (`lexer.go:58-82`) forbids only a literal newline and an unescaped closing quote inside the literal — every other byte, including `*` and `/`, is accepted verbatim, with no escape processing at all. So ordinary, syntactically valid Lang source such as:

```
foreign C {
    fn lang_res_open(x: Byte) -> Byte {
        allocator: "*/ int injected(void){return 1;} /*"
        unwind: forbidden
        nonlocal_exit: forbidden
        fails: SomeErr
    }
}
```

lexes and parses without a single diagnostic, is copied verbatim into `foreignSymbolInfo.Allocator` (`check.go:1107`) and then into `core.ForeignContract.Allocator` (`check.go:1337`), and `corevalidate.linear`'s only check on it is `Allocator != ""` (via the `foreign.obligation_undeclared`-adjacent non-empty check, `corevalidate.go:327` covers `Unwind`/`NonlocalExit`; `Allocator` has no shape check anywhere). `EmitForeignHeader` (called directly, and also by `EmitForeignConformance`, `cgen.go:1379`) then produces:

```c
/* allocator: */ int injected(void){return 1;} /* */
```

— the `*/` in the payload closes the comment early, the injected fragment becomes live top-level C source, and the trailing `/*` reopens a new comment that silently swallows the rest of the intended obligation block. I confirmed this is exploitable end-to-end with a standalone reproduction: calling `cgen.EmitForeignHeader` on a `core.Program` whose `ForeignContract.Allocator` is `"*/ int injected(void){return 1;} /*"` produces a header containing a live `int injected(void){return 1;}` function definition outside any comment.

This header is not merely descriptive output — `session.go:575` (`LayoutMutationRunner.Run`) and `session.go:2200-2225` route `EmitForeignHeader`'s and `EmitForeignConformance`'s output directly into `native.Runner.CompileConformanceUnit`, which invokes a real C compiler on the generated source (`native.go:207-216`). `cgen.ScanForBannedAttributes` (the one post-generation content scan in the pipeline, `session.go:2230`) only searches for a fixed list of optimizer-attribute tokens (`restrict`, `noalias`, etc.) and does not detect an escaped comment or injected function definition.

Critically, this does **not** require a hand-corrupted `core.Program` the way the fixed `Symbol` gap did — `Allocator`/`Unwind`/`NonlocalExit` reach this state through the ordinary, honest `check.Program` → `cgen.EmitForeignHeader`/`EmitForeignConformance` pipeline from source-level syntax the language already supports (a quoted foreign-policy value). Any Lang source file that declares a `foreign` block with a crafted `allocator`/`unwind`/`nonlocal_exit` string value can inject arbitrary top-level C source into a file this compiler then compiles.

**Fix:** Require every `ForeignContract` string field that is ever spliced into generated C text (comment or code) to be validated before emission, not just `Symbol`. Two complementary fixes:

1. At the source admission boundary (`check.go`), reject a policy value containing `*/`, or more conservatively require `allocator`/`unwind`/`nonlocal_exit` to always be identifier-shaped (matching the existing informal convention — `forbidden`, `libc_malloc`, etc. — none of the corpus fixtures actually need free-form string content for these three keys):
```go
// in collectForeignSymbols, alongside the existing switch:
if !validPolicyIdentifier(policy.Value) {
    return nil, []diagnostic.Diagnostic{diagnostic.Error("check.foreign_policy_value_unsafe", policy.Span, "policy value must be a plain identifier")}
}
```
2. Defense-in-depth at `corevalidate.linear`, alongside the existing `Symbol` check, apply the same `validCIdentifier` (or a comment-safe equivalent that at minimum forbids `*/`) to `Allocator`, `Unwind`, and `NonlocalExit` before `cgen` ever sees them — the same posture already adopted for `Symbol`.
3. As a last line of defense, `cgen.EmitForeignHeader` should itself refuse (or escape) any contract field it is about to splice into a comment if it contains `*/`, mirroring `validForeignSymbol`'s role as an independent peer guard for the entry points that skip `corevalidate.Validate`.

Add a falsifier alongside `hostileForeignSymbols` in `cgen_test.go`/`corevalidate_test.go` for each of `Allocator`, `Unwind`, and `NonlocalExit` containing a comment-terminator payload, asserting refusal (or, at minimum, that `EmitForeignHeader`'s output never contains the injected fragment) the same way `TestForeignSymbolInjectionNeverReachesGeneratedC` currently does for `Symbol` alone.

## Warnings

### WR-01: `TestReleaseOrderValidationWorkSeries`'s monotonic-work assertion still does not prove linearity, and does not model the new per-interior-merge recursive cost the 04-11 fix introduced

**File:** `internal/compiler/corevalidate/corevalidate_test.go:709-732` (carried forward; unchanged by 04-11/04-12)
**Issue:** Flagged in the prior review round and still unaddressed. The test's doc comment claims the rederivation cost is proven linear, but the test body only asserts a three-point series is strictly increasing, which is consistent with worse-than-linear growth. This is now additionally relevant because 04-11's interior-merge fix adds a *recursive*, per-candidate-edge branch to `rederive` (`corevalidate.go:1290-1323`) whose cost is proportional to the number of interior merge points times the fan-in at each — a shape no current fixture exercises and the monotonic assertion cannot distinguish from linear growth.
**Fix:** Either soften the doc comment to describe only what is actually checked, or strengthen the assertion to at least four points with a ratio bound consistent with linear growth, and add a fixture exercising a chain of interior merges to exercise the new recursive branch's cost.

## Info

### IN-01: No fixture exercises 3+ incoming `"ok"` edges into a single interior block (only 2-edge merges are tested)

**File:** `internal/compiler/corevalidate/corevalidate_test.go:878-994`
**Issue:** `TestInteriorMergeDivergentHistoriesRefused`/`TestInteriorMergeAgreeingHistoriesAccepted` both exercise exactly two incoming `"ok"` edges into the interior block. The 04-11 fix's pairwise-against-`candidates[0]` comparison (`corevalidate.go:1314-1320`) is transitively correct for a 3+-edge merge (if all agree with the first, all agree with each other), but this is not independently exercised by any fixture. Not a defect — carried forward as a coverage note, consistent with the "performance is out of scope" instruction for the analogous prior-round IN-01.
**Fix:** None required; consider a 3-edge interior-merge fixture if this code path is ever refactored.

---

_Reviewed: 2026-09-05T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
