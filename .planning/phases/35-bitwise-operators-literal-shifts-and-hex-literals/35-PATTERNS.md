# Phase 35: Bitwise Operators, Literal Shifts, and Hex Literals - Pattern Map

**Mapped:** 2026-10-14  
**Files classified:** 27 source, test, fixture, and consumer groups  
**Tracked analogs found:** 27 / 27 file groups; the D-07 unused-Byte bridge contract has no exact analog.

## File Classification

| New/Modified File or Group | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/syntax/token.go` | utility | transform | `internal/compiler/syntax/token.go` | exact |
| `internal/compiler/syntax/lexer.go` | utility | transform | `internal/compiler/syntax/lexer.go` | exact |
| `internal/compiler/syntax/parser.go` | utility | transform | `internal/compiler/syntax/parser.go` | exact |
| `internal/compiler/syntax/format.go` | utility | transform | `internal/compiler/syntax/format.go` | exact |
| `internal/compiler/check/check.go` | service | transform | `internal/compiler/check/check.go#parseU64Literal` | exact |
| `internal/compiler/check/scalar.go` | service | transform | `internal/compiler/check/scalar.go#scalarBuilder.expression` | exact |
| `internal/compiler/core/core.go` | model | transform | `internal/compiler/core/core.go#AllOperationKinds` | exact |
| `internal/compiler/corevalidate/corevalidate.go` | service (independent validator) | transform | `internal/compiler/corevalidate/corevalidate.go#validateScalarFixedPoint` | exact |
| `internal/compiler/originvalidate/originvalidate.go` | service (independent validator) | transform | `internal/compiler/originvalidate/originvalidate.go#originTransferScalarOperation` | role-match; keep derivation independent |
| `internal/compiler/interp/interp.go` | service | transform | `internal/compiler/interp/interp.go` scalar operation switch | exact |
| `internal/compiler/pathoracle/pathoracle.go` | utility | transform | `internal/compiler/pathoracle/pathoracle.go` operation ancestry walk | role-match |
| `internal/compiler/cgen/cgen.go` | service | transform | `internal/compiler/cgen/cgen.go#Emit`, `#EmitNative`, `#EmitApplication` | role-match |
| `internal/compiler/cgen/cgen_program.go` | service | transform | `internal/compiler/cgen/cgen_program.go#hasScalarOperations` and scalar emitter | exact |
| `internal/compiler/executionpeer/executionpeer.go` | service | event-driven | `internal/compiler/executionpeer/executionpeer.go#Validate` | role-match; causal evidence only |
| `internal/compiler/surface/refusals.go` | config | request-response | `internal/compiler/surface/refusals.go#Refusals` | exact |
| `internal/compiler/surfacemanifest/manifest.go` | service | request-response | `internal/compiler/surfacemanifest/manifest.go#Manifest` | exact |
| `internal/compiler/core/core_test.go` | test | batch | `TestAllOperationKindsRegistered`, `TestAllOperationKindsHandledAtEverySite` | exact inventory pattern; not transfer proof |
| `internal/compiler/syntax/*phase35*_test.go` | test | transform | `internal/compiler/syntax/parser_phase19_test.go`; `parser_phase34_refusal_test.go` | role-match |
| `internal/compiler/check/*phase35*_test.go` | test | transform | `internal/compiler/check/check_phase19_test.go` | role-match |
| `internal/compiler/corevalidate/*phase35*_test.go` | test | transform | `internal/compiler/corevalidate/corevalidate_phase19_test.go` | role-match |
| `internal/compiler/originvalidate/*phase35*_test.go` | test | transform | `internal/compiler/originvalidate/originvalidate_phase29_test.go` | role-match |
| `internal/compiler/interp/*phase35*_test.go` | test | transform | `internal/compiler/interp/interp_test.go` | role-match |
| `internal/compiler/executionpeer/*phase35*_test.go` | test | event-driven | `internal/compiler/executionpeer/executionpeer.go#Validate` and package tests | role-match |
| `internal/compiler/cgen/*phase35*_test.go` | test | transform | `internal/compiler/cgen/cgen_program.go` and package tests | role-match |
| `internal/compiler/session/*phase35*_test.go` | test | request-response | `internal/compiler/session/session_phase34_prechange_test.go`; `session_phase34_surface_test.go` | role-match |
| `testdata/phase35/*` | config / evidence fixture | batch | `testdata/phase34/surface-witnesses.json`; `testdata/phase34/prechange-corpus.json` | role-match |
| `examples/checksum/adler32.schway` | service | transform | the existing Adler module itself | exact consumer |

The phase 35 selectors are named in RESEARCH.md, but their Go test filenames and phase 35 fixture names remain planner choices. Keep test inputs and expected vectors independently authored. Do not make a producer-generated manifest or let an agreement-only engine test stand in for expected answers.

## Pattern Assignments

### Frontend tokens, parsing, grouping, and formatting

**Analogs:** `internal/compiler/syntax/parser.go`, `lexer.go`, `token.go`, `format.go` (all tracked).

The existing scalar parser is Pratt-style: one precedence function, a recursive right parse at `precedence + 1`, then a small operator-to-AST mapping. Preserve this control flow and the current relative order of old tiers when inserting shift and bitwise precedence.

**Pratt pattern** (`parser.go`, lines 753-794):

```go
func scalarBinaryOperator(kind Kind) int {
    switch kind {
    case TokenLAngle, TokenEqualEqual:
        return 1
    case TokenPlus:
        return 2
    case TokenPercent:
        return 3
    default:
        return 0
    }
}

// In scalarBinaryExpr:
precedence := scalarBinaryOperator(operator.Kind)
if precedence < minPrecedence || precedence == 0 {
    return left
}
p.advance()
right := p.scalarBinaryExpr(precedence + 1)
left = &ast.Expr{Kind: kind, Left: left, Right: right,
    Span: diagnostic.Span{Start: left.Span.Start, End: right.Span.End}}
```

`parser.go` lines 798-822 owns the contiguous `||` refusal separately from `TokenPipe`. Extend admitted single-pipe handling without weakening that check. The lexer currently keeps `<` and `>` as individual tokens (`token.go` lines 43-51; `lexer.go` lines 240-274); if retaining that strategy, recognize shifts only from adjacent source spans in scalar-expression context so generic closers and comparisons remain intact. The token tree already retains source spans and token text.

Numeric ownership is deliberately lexical: `lexer.go` lines 54-73 consumes the complete candidate, emits one `syntax.malformed_numeric_literal` over the full span, and does not hand a valid prefix to the parser. Keep this ownership for invalid digits, suffixes, and separators. `format.go` lines 261-267 writes `TokenNumber.Text` unchanged; preserve accepted spelling during formatting and round trips. Its `isBinaryOperatorToken` at lines 55-58 is a separate formatter inventory that must learn the new operators.

Transparent groups should preserve the inner expression kind and literal provenance. Prefer extending the expression span while retaining the inner node over adding a runtime/core operation. The `parser.go` resource pattern is the named type caps in lines 69-92; add a distinct expression-depth counter/limit and manifest fact rather than reusing the type-depth label. Deep recovery should follow the iterative delimiter-stack routine at `parser.go` lines 1410-1459 so sibling diagnostics remain reachable.

### Numeric admission and producer lowering

**Analogs:** `internal/compiler/check/scalar.go#scalarBuilder.expression` and `internal/compiler/check/check.go#parseU64Literal` (both tracked).

`scalar.go` lowers a numeric leaf to a fresh `OpConst` with a canonical decimal payload, and evaluates binary operands left before right. Reuse that order for bitwise ops. The current remainder special case checks the direct right AST node for `Kind == "number"` (lines 644-668); the shift rule needs the corresponding check after transparent grouping, while still refusing names and computed expressions even when constant propagation knows their value.

**Numeric leaf and binary lowering** (`scalar.go`, lines 631-671):

```go
case "number":
    value, err := strconv.ParseUint(expr.Value, 0, 64)
    if err != nil {
        value, err = strconv.ParseUint(expr.Value, 10, 64)
    }
    if err != nil {
        b.refuse("type.invalid_u64", expr.Span, "invalid U64 literal")
        return "", ""
    }
    place := b.addPlace("literal", "U64", false)
    b.addOperation(core.LinearOperation{Kind: core.OpConst, TargetID: place.ID,
        ConstU64: strconv.FormatUint(value, 10), TypeID: b.types["U64"]})
```

The legacy route has its own policy at `check.go` lines 6809-6817: it strips underscores and recognizes lowercase `0x`/`0b`, while the scalar route uses base-zero then decimal fallback. Add uppercase `0X` narrowly to the legacy route. Keep the `010` and `0B10` route differences pinned; do not use this change to unify conversion policy.

### Core field contract and independent validation

**Analogs:** `internal/compiler/core/core.go`, `corevalidate/corevalidate.go`, and `originvalidate/originvalidate.go` (all tracked).

The core kind list is a closed inventory. Add distinct left/right shift kinds to both the declarations and `AllOperationKinds()`. The existing `ConstU64` JSON field is `omitempty` and documented as `OpConst`-only (`core.go` lines 969-972); change that comment and exclusivity contract to allow only `OpConst` and the two shifts. Keep old core bytes unchanged when the new kinds are absent.

Core validation already has separate field-kind checks and scalar fact paths. Its operation admission at `corevalidate.go` lines 1425-1438 rejects `ConstU64` outside `OpConst` and checks `RightID` against a closed allowlist; update both deliberately. Then add immediate validation and U64 transfer cases independently to both `validateScalarFixedPoint` (lines 4722-4850) and `validateStraightLineScalarFacts` (lines 4892-4948). Require canonical decimal `0..63`, empty `RightID`, initialized U64 source, and U64 target/type. Raw-core cases should cover every malformed field and immediate form.

Origin validation is a separate derivation, not a helper call into corevalidate. `originTransferScalarOperation` (`originvalidate.go` lines 418-468) has its own abstract state and operator cases. Add its own immediate/type/range checks and exact transfer. Also update `originHasScalarCFG` (lines 293-300), `phase27ArithmeticProblem` (1421-1449), and `PublishProblemsFor` (1330-1345): the current scalar publication trigger is remainder-specific, so a new pure op must reach origin admission/recomputation without relying on a remainder elsewhere in the function. Keep wrong-transfer controls observable through each peer's downstream obligation or a direct independent transfer assertion.

The core test `TestAllOperationKindsHandledAtEverySite` (`core_test.go` lines 1055-1073) is a useful exhaustive-inventory control. It proves dispatch reachability, not semantic transfer correctness; retain separate named shift contracts and transfer mutants for corevalidate and originvalidate.

### Interpreter, ancestry, C emission, and causal evidence

**Analogs:** `internal/compiler/interp/interp.go`, `pathoracle/pathoracle.go`, `cgen/cgen_program.go`, and `executionpeer/executionpeer.go` (tracked).

The interpreter switch around `interp.go` lines 1713-1764 shows the scalar operation shape: read initialized operands, fail closed if types/values are malformed, assign the target, and advance. New pure ops should have no defect event, reason, or runtime event. `pathoracle.go` lines 1048-1081 special-cases `OpConst` as a fresh loan-free root; add the shift kinds as ordinary U64 value operations without loan semantics.

`cgen_program.go` has multiple independent inventories. `schema2ExecutionDocumentSize` lines 220-245 reserves events for operations that may defect and has an explicit no-event arm for pure operations. `hasScalarOperations` lines 2547-2559 separately selects scalar C emission; the emission switch near 2731-2746 renders existing unsigned U64 comparisons/remainder. Add shifts to the scalar path and the no-event accounting arm. Emit a checked immediate with an unsigned C operand and a final exact-width cast, following the researched shape: `(uint64_t)((unsigned long long)lhs << count)` (analogous right shift). The static generated-C scan must count actual shift sites and reject dynamic and 64+ count controls. Check all exported `cgen` gates (`Emit`, `EmitNative`, `EmitApplication`) as research identifies distinct entry paths.

`executionpeer.Validate` (`executionpeer.go:72`) validates observed causal structure and takes a document, not an entry value. Its role is per-op event-free admission and rejection of fabricated step/defect events; it cannot prove numeric results. Under the live D-18 amendment, numeric transfer mutants belong to check/corevalidate/originvalidate and wrong-answer controls to interpreter/native; executionpeer controls remain causal-only. Do not describe causal admission as numeric replay or add operand/value events.

### Refusals, manifest, and evidence fixtures

**Analogs:** `internal/compiler/surface/refusals.go`, `internal/compiler/surfacemanifest/manifest.go`, `internal/compiler/session/session_phase34_surface_test.go`, `session_phase34_prechange_test.go` (tracked).

Add named shift-amount and expression-depth refusals to the existing catalog, then expose them through the existing `schway.surface/1` ten-field manifest. Do not add a JSON field or new discovery API. The Phase 34 witness test (`session_phase34_surface_test.go` lines 452-513) compares independent complete-source witnesses to the manifest and verifies expected status/code/hint; its controls around lines 516 onward mutate manifest claims and catalog rows. Extend that pattern with independently authored operator, precedence, literal-only, hex spelling, and depth witnesses plus the finite source/admission/limit census.

The immutable baseline analog is `session_phase34_prechange_test.go`: lines 66-84 enumerate tracked `.schway` files via `git ls-files`; lines 87-118 records checks and emitted/non-emitted outcomes; lines 285-308 checks the frozen corpus includes both outcomes; lines 310 onward arms a byte-drift control. Create a new Phase 35 corpus before compiler edits and leave Phase 34 goldens untouched. Audit `cgen.Emit`, `EmitNative`, and `EmitApplication`; Phase 34's corpus covers application and control-native routes, while RESEARCH.md flags direct `Emit` as another route to include or explicitly justify. Treat the Adler rewrite as a named source exception with unchanged answer expectations.

For value tests, use the Phase 19 syntax pattern at `parser_phase19_test.go` lines 14-41: assert source spelling and exact spans, then formatter idempotence and spelling retention. Use `corevalidate_phase19_test.go` lines 25-62 for table-driven forged-operation mutations. The Phase 35 value vectors need independently computed answers at counts 0, 31, 32, 63, high-bit/all-zero cases, wrong-result controls, and the focused native/UBSan lanes specified by RESEARCH.md. The existing Adler code in `examples/checksum/adler32.schway` lines 20-37 is the consumer to compact to `b << 16`; retain its final addition and answer tests.

## Shared Patterns

### Keep source, IR, and peer authority separate

The checker owns the source-literal rule. Corevalidate and originvalidate independently own raw immediate/type/range admission and transfer. Neither IR peer can prove that a raw immediate came from a source number token. Keep expected answer vectors independent from implementation and ensure transfer mutants reach a downstream observable sink.

### Pure operations remain event-free

Update each operation-kind inventory explicitly: core registration, core/origin transfer triggers, interpreter, pathoracle, cgen scalar selection, C event projection/capacity, and executionpeer causal admission. Do not add a defect reason or event for a defined bitwise operation or shift. Preserve the D-18 boundary: executionpeer proves causal shape, not numeric answers.

### Preserve source and wire compatibility

Keep original numeric token text in the lossless tree and formatter. Canonicalize shift immediates only when lowering to the existing decimal `ConstU64` field. Keep `RightID` empty for shifts and preserve `omitempty`, legacy identifiers, old goldens, `010` route behavior, and all no-feature generated-C bytes. Bound expression recursion separately from type recursion and recover iteratively.

### Keep the two runnable observations distinct

The D-07 bridge is a narrowly planned admission: one `unused: Byte` formal, U64/Bool local work, and no Byte use or conversion. Source checking and independent core/origin laws must refuse any use of that placeholder. Existing Phase 29 tests are only route analogs: interpreter conformance asserts an `Outcome.Value`, while native application asserts stdout bytes. For Phase 35 pin public conformance `Outcome.Value == "264"` on both engine routes, and separately pin application/interpreter-application output `"264\n"`. Do not add a U64 CLI default, new input flag, or change the U64 input-required rule. This bridge is planned; the Phase 29 fixtures do not prove it.

## No Analog Found

| File / Contract | Role | Data Flow | Reason |
|---|---|---|---|
| D-07 `fn main(unused: Byte) -> U64` bridge with no Byte use/conversion | service admission | request-response | Current scalar entry admission does not provide this exact unused-Byte/U64-local contract. The Phase 29 fixture tests are route/output analogs only; they do not prove the bridge. |

## Metadata

**Analog search scope:** `internal/compiler/{syntax,check,core,corevalidate,originvalidate,interp,pathoracle,cgen,executionpeer,surface,surfacemanifest,session,native}`; `examples/checksum`; `testdata/phase34`.  
**Tracked-source gate:** every existing analog path named above was verified with `git ls-files -- <path>`; no install/runtime mirror paths are used.  
**Pattern extraction date:** 2026-10-14.  
**Contract authority:** current `35-CONTEXT.md`, including the 2026-10-14 D-07 clarification and D-18 amendment. Earlier D-07/D-18 research text is provenance, not the live acceptance contract.
