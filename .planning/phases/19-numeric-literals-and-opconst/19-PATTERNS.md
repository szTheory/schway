# Phase 19: Numeric Literals and `OpConst` - Pattern Map

**Mapped:** 2026-09-24  
**Files analyzed:** 23 anticipated source and test files across 11 implementation seams  
**Analogs found:** 23 / 23 (closest analogs for implementation; test files should extend owning package suites)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/syntax/token.go` | model/config | transform | same file's `Token` model | exact |
| `internal/compiler/syntax/lexer.go` | utility | transform | same file's string/identifier scanning | exact |
| `internal/compiler/syntax/parser.go` | utility | transform | `linearBody` binding parser | exact |
| `internal/compiler/syntax/format.go` | utility | transform | same file's token formatter | exact |
| `internal/compiler/ast/ast.go` | model | transform | `Binding` / `RHS` | exact |
| `internal/compiler/check/check.go` | service | transform | `checkLinear` binding lowering | role-match |
| `internal/compiler/core/core.go` | model | transform | `OpConstructPayload` declaration and registry | role-match |
| `internal/compiler/corevalidate/corevalidate.go` | service | transform | `replayStraightLine` operation switch | role-match |
| `internal/compiler/interp/interp.go` | service | request-response | scalar `value` and `String`; operation execution switch | role-match |
| `internal/compiler/cgen/cgen_program.go` | service | transform | `emitProgramFunction` operation emission | role-match |
| `internal/compiler/pathoracle/pathoracle.go` | service | transform | `linearizePath` operation traversal | role-match |
| `internal/compiler/originvalidate/originvalidate.go` | service | transform | `RecomputeOriginPerReturn` operation traversal | role-match |
| `internal/compiler/core/core_test.go` | test | transform | exhaustive dispatch fixtures/control | exact |
| `internal/compiler/syntax/syntax_test.go` | test | transform | lexer/formatter tests | exact |
| `internal/compiler/syntax/parser_phase19_test.go` (anticipated) | test | transform | `parser_phase18_test.go` | role-match |
| `internal/compiler/check/check_phase19_test.go` (anticipated) | test | transform | `check_phase18_test.go` | role-match |
| `internal/compiler/corevalidate/corevalidate_phase19_test.go` (anticipated) | test | transform | `corevalidate_phase18_test.go` | role-match |
| `internal/compiler/interp/interp_test.go` | test | request-response | interpreter fixture tests | exact |
| `internal/compiler/cgen/cgen_program_test.go` (anticipated) | test | transform | existing emitter tests | role-match |
| `internal/compiler/pathoracle/pathoracle_test.go` | test | transform | path recomputation tests | exact |
| `internal/compiler/originvalidate/originvalidate_test.go` | test | transform | origin recomputation tests | exact |
| `internal/compiler/session/session_phase19_test.go` (anticipated) | test | request-response | `session_phase11_differential_test.go` | role-match |
| `internal/compiler/session/session_phase5_compare.go` | utility | transform | `Phase5CompareProgramEngines` | exact (reuse; no new comparator) |

The test paths suffixed `phase19` are suggested by neighboring phase-specific suites, not locked filenames. Follow repository naming where the concrete plan identifies an existing test file to extend instead.

## Pattern Assignments

### `internal/compiler/syntax/token.go` and `lexer.go` (tokenization, transform)

**Analogs:** `internal/compiler/syntax/token.go`, `internal/compiler/syntax/lexer.go` (tracked).

The token already carries exact spelling and source span; add a distinct numeric kind and preserve `Text` verbatim. Lexer scanning is byte-offset based and emits a span for every token. The current identifier path is:

```go
if unicode.IsLetter(r) || r == '_' {
    offset += size
    for offset < len(source) {
        next, nextSize := utf8.DecodeRune(source[offset:])
        if !(unicode.IsLetter(next) || unicode.IsDigit(next) || next == '_') {
            break
        }
        offset += nextSize
    }
    text := string(source[start:offset])
    kind := TokenIdentifier
    if keyword, ok := keywords[text]; ok {
        kind = keyword
    }
    tokens = append(tokens, Token{Kind: kind, Text: text, Span: diagnostic.Span{Start: start, End: offset}})
    continue
}
```

Malformed numeric forms should produce source diagnostics with the full offending span, following the lexer’s fixed diagnostic pattern (e.g. `syntax.unterminated_string` at `lexer.go:73-76`). Do not let separators, suffix tails, or invalid radix digits get silently split into an admitted numeric token plus unrelated tokens.

### `internal/compiler/syntax/parser.go` and `ast/ast.go` (literal source form, transform)

**Analogs:** `internal/compiler/syntax/parser.go:390-455`, `internal/compiler/ast/ast.go:165-192` (tracked).

`linearBody` parses `let name = ...`, recognizes special RHS forms, otherwise builds an `ast.RHS` with kind, source and span. Keep literal spelling/span in the source AST so checker diagnostics remain localized; parsing alone should not decide range/type admission. The normal binding construction is:

```go
source := p.identifier("syntax.expected_binding_source")
body.Bindings = append(body.Bindings, ast.Binding{
    Name: name.Text, RHS: ast.RHS{Kind: kind, Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
})
body.Span.End = source.Span.End
```

`Binding` and `RHS` are deliberately small tagged source structures. Numeric literals need a distinguishable RHS kind/operand; do not overload a place-name RHS, since later checker logic interprets `Source` as a bound place.

### `internal/compiler/syntax/format.go` (formatter, transform)

**Analog:** `internal/compiler/syntax/format.go:8-26, 66-80, 144-159` (tracked).

Formatter is a token/CST projection: it removes whitespace tokens, retains comments, and writes semantic `token.Text`. For numeric tokens, preserve spelling and let layout canonicalization handle surrounding whitespace.

```go
for _, token := range tree.Tokens {
    if token.Kind == TokenEOF || token.Kind == TokenWhitespace {
        continue
    }
    tokens = append(tokens, token)
}
```

The `TokenIdentifier` arm writes `token.Text` directly (`format.go:144-159`). Follow that spelling-preserving behavior, while checking whether the new token needs a binding-boundary newline just as identifiers currently do.

### `internal/compiler/check/check.go` (checker and core lowering, transform)

**Analog:** `internal/compiler/check/check.go` `checkLinear` and binding lowering (tracked; use `checkLinear` as the closest admission/lowering path).

The checker owns source-level refusals and assigns semantic facts before operations enter executable core. The `checkLinear` path resolves parameter/return types and returns span-bearing `diagnostic.Error(...)` values on refusals. Numeric literal admission should independently enforce valid magnitude and U64 type here (or in a checker-owned helper called here); overflow must not reach a core operation.

For each binding, build the target place through existing monotonically indexed place/operation ID conventions and append a new `core.OpConst` carrying the admitted canonical decimal fact. Enforce normal function-boundary type rules; no untyped/context-inferred constant path should be created. Consult the existing `OpConstructPayload` emission in check.go (around lines 2300-2400) as a precedent for making a new kind rather than disguising value production as `OpCopy`.

### `internal/compiler/core/core.go` (core model and registry, transform)

**Analog:** `internal/compiler/core/core.go:663-680, 814-822, 831-840` (tracked).

`OpConstructPayload` is explicitly distinct because it introduces new value semantics and must be visible to exhaustive dispatch. `AllOperationKinds()` is a maintained registry consumed by completeness controls:

```go
func AllOperationKinds() []OperationKind {
    return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall, OpConstructPayload, OpDestructurePayload}
}
```

Add `OpConst` to the declared kinds and this registry, and define kind-exclusive fields/invariants on `LinearOperation` as needed. Keep a canonical semantic value separate from source spelling; source spelling belongs to the syntax/AST path.

### `internal/compiler/corevalidate/corevalidate.go` (independent admission peer, transform)

**Analog:** `internal/compiler/corevalidate/corevalidate.go:2138-2250` (`replayStraightLine`, tracked).

The peer independently validates universal source/type/initialization facts before its operation-kind switch and updates its own place state in each arm. `OpConst` must validate its target uniqueness/type and constant payload representation, then independently mark its target initialized and produced. Do not call checker logic or reuse checker-produced validation verdicts.

The existing `OpConstructPayload` arm at lines 2205-2222 is a concrete new-kind pattern: document why its target rules differ and update source/target liveness itself. Keep unused fields empty for `OpConst` under kind-exclusive-field checks.

### `internal/compiler/interp/interp.go` (interpreter execution and projection, request-response)

**Analog:** `internal/compiler/interp/interp.go:432-435, 466-479, 867-1008` (tracked).

Scalar values use the existing `{tag, payload string}` representation; scalar serialization returns payload verbatim when no tag is set:

```go
type value struct {
    tag     string
    payload string
}

func (v value) String() string {
    if v.tag != "" {
        return v.tag
    }
    return v.payload
}
```

The numeric `OpConst` execution arm should place canonical decimal digits in `payload`; do not replace this shared projection or encode the source radix. Follow the existing operation switch's target declaration/state updates and ensure execution records any event required by the established event protocol.

### `internal/compiler/cgen/cgen_program.go` (native lowering, transform)

**Analog:** `internal/compiler/cgen/cgen_program.go:472-490, 642-680, 1297-1405` (tracked).

The surviving program emitter writes its shared C include block in `emitProgram`, then each function dispatches operations in `emitProgramFunction`. A new binding-producing operation belongs in both relevant straight-line and branch operation emitters if validation permits both; current checker/source linear bindings and match-arm bodies share executable paths. Use the actual core literal value, not source text, as the emission input.

For U64, emit `<stdint.h>` support and an exact-width constant form (`uint64_t` plus `UINT64_C(...)` where supported); missing exact-width support must fail closed. Never use `long`, host pointer size, a signed intermediate, or a narrowing cast. Preserve emitter errors as `fmt.Errorf` with operation identity, consistent with invalid-target/unresolved-callee errors in the current switch.

### `internal/compiler/pathoracle/pathoracle.go` (independent traversal, transform)

**Analog:** `internal/compiler/pathoracle/pathoracle.go:281-329` (tracked).

`linearizePath` walks operations in deterministic order, follows source-place loan inheritance, and only creates loan state for borrow kinds. It does not need literal-value evaluation. Ensure `OpConst` does not get misclassified as a terminator/borrow and that its produced target is accounted for wherever this oracle derives source/target relationships. The exhaustive dispatch control calls `RecomputeEndpoints` on a real program containing every required kind.

### `internal/compiler/originvalidate/originvalidate.go` (independent origin analysis, transform)

**Analog:** `internal/compiler/originvalidate/originvalidate.go:280-330` (tracked).

`RecomputeOriginPerReturn` walks from return source to producer, using a kind switch to infer borrow access and then following `SourceID`. A constant is an origin root: ensure traversal stops or yields the expected non-borrow origin rather than following an empty/irrelevant source place. Preserve independent logic; this peer must not depend on checker results to interpret the operation.

### `internal/compiler/core/core_test.go` (exhaustive-dispatch test, transform)

**Analog:** `internal/compiler/core/core_test.go:551-703` (tracked).

The control requires each declared kind to occur in a checked fixture and drives the fixture through `session.Check`, `corevalidate`, `pathoracle`, `originvalidate`, `interp`, and `cgen`. Add a literal-bearing fixture and include it in `exhaustiveDispatchFixtures`; add `OpConst` to `AllOperationKinds()`. The control has explicit anti-vacuity logic:

```go
for _, kind := range requiredKinds {
    if !encountered[kind] {
        return fmt.Errorf("operation kind %q is never exercised by any corpus fixture in this control", kind)
    }
}
```

Retain the existing mutation controls proving omission of a required kind/site is detected. Note `pathoracle` and `originvalidate` currently mean “walk completes without error” in this control, not an exhaustive kind switch.

### `internal/compiler/session/session_phase11_differential_test.go` and `session_phase5_compare.go` (four-tier evidence, request-response)

**Analogs:** `internal/compiler/session/session_phase11_differential_test.go:144-209`, `internal/compiler/session/session_phase5_compare.go:106-180` (tracked).

Reuse `phase11RunFourTiers`/`phase11RunFourTiersWithSupplier` to run the interpreter and native C at `-O0`, `-O3`, and `-O3 -flto`. Reuse `Phase5CompareProgramEngines`, which validates each schema-2 document before its all-pairs comparator. The existing fixture helper explicitly calls:

```go
engines := phase11RunFourTiers(t, ctx, program, entryName, input)
if err := session.Phase5CompareProgramEngines(fixture, program, engines); err != nil {
    t.Fatalf("%s: four-tier disagreement: %v", fixture, err)
}
```

Create/use a fixture whose returned scalar depends on the literal. Do not add a parallel comparator or settle for compiling C without observing the numeric return. Also replay established scalar projection goldens (Phase 12) before blessing any golden movement.

### Package tests (syntax, checker, corevalidate, interp, cgen, pathoracle, originvalidate)

Use nearby package tests as test-layout analogs: `syntax/syntax_test.go` and `parser_phase18_test.go`; `check/check_phase18_test.go`; `corevalidate/corevalidate_phase18_test.go`; `interp/interp_test.go`; and existing `cgen`, `pathoracle`, and `originvalidate` test suites. Tests should distinguish malformed spelling/range diagnostics, max/max+1 and radix equivalence, preserved formatted token spelling, semantic core admission, interpreter decimal projection, and exact-width native output. Keep exhaustive dispatch, golden replay, mutation controls, and four-tier differential as distinct claims.

## Shared Patterns

### Source spelling vs semantic value

**Sources:** `syntax.Token{Kind, Text, Span}` (`syntax/token.go:47-51`), `Format` (`syntax/format.go:10-26`), interpreter `value` (`interp/interp.go:432-435`).

Keep original spelling on tokens/AST for formatting and source spans; derive and check one U64 semantic value before core admission; project runtime scalars as canonical decimal strings.

### Independent peers and dispatch completeness

**Sources:** `core.AllOperationKinds` (`core/core.go:814-822`) and `TestAllOperationKindsHandledAtEverySite` (`core/core_test.go:551-703`).

`check`, `corevalidate`, `interp`, `cgen`, `pathoracle`, and `originvalidate` retain distinct responsibilities. Register the kind, implement the operation in all six, add a real fixture, and preserve controls that fail when the kind is absent or a site stops handling it.

### Error handling and fail-closed representation

**Sources:** source diagnostics in `syntax/lexer.go:73-76`; emitter error returns in `cgen/cgen_program.go:1330-1375`; C include block in `cgen/cgen_program.go:642-646`.

Malformed/out-of-range literals must be refused before executable core with a span-bearing diagnostic. Backend target support must use a proven exact-width C type and surface unsupported targets as compilation errors, never silently narrowing.

### Differential evidence

**Sources:** `session_phase11_differential_test.go:152-209`; `session_phase5_compare.go:110-180`.

Use the established interpreter plus three native tiers and existing all-pairs/five-axis comparator. The literal must affect the observable return value. Keep scalar-golden replay separate from engine agreement.

## No Analog Found

No production analog exists for numeric token grammar, numeric AST operands, checker-produced `OpConst`, or U64 C target gating because numeric values are new to the language. Use the adjacent patterns above while observing D-19-01 through D-19-05.

| File/Concern | Role | Data Flow | Reason |
|---|---|---|---|
| Numeric literal grammar and checked radix conversion | utility | transform | No numeric lexer/parser semantics exist yet. |
| `OpConst` constant operand/value contract | model | transform | No operation currently introduces a source-written scalar value. |
| C exact-width U64 capability gate | service | transform | Existing emitter has no U64 target representation to reuse. |

## Metadata

**Analog search scope:** `internal/compiler/{syntax,ast,check,core,corevalidate,interp,cgen,pathoracle,originvalidate,session}`  
**Tracked-source check:** Every analog path named above is present in `git ls-files`; no ignored runtime/plugin mirrors used.  
**Pattern extraction date:** 2026-09-24
