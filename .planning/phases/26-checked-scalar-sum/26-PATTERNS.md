# Phase 26: Checked Scalar Sum - Pattern Map

**Mapped:** 2026-10-02  
**Files analyzed:** 18 likely source, witness, and evidence files (new tests may be colocated)  
**Analogs found:** 17 / 18

Scope comes from `26-CONTEXT.md` and `26-RESEARCH.md`: mutable local U64/Bool scalars, pre-tested `while`, block `if/else`, checked `+`, `<`, bounded independent scalar CFG fixed points, unchanged refusals for resource/loan/provenance back-edge carries, plus a runnable `sum_to_n` through the current app route. Exact new filenames are planning decisions; existing test files below are pattern sources, not a mandate to split tests into those files.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/ast/ast.go` | model | transform | `internal/compiler/ast/ast.go` | role-match |
| `internal/compiler/syntax/{token.go,lexer.go,parser.go,format.go}` | utility | transform | `internal/compiler/syntax/parser.go`, `format.go` | role-match |
| `internal/compiler/check/check.go` | controller/checker | request-response | `internal/compiler/check/check.go` | role-match |
| `internal/compiler/core/core.go` | model | transform | `internal/compiler/core/core.go` | role-match |
| `internal/compiler/corevalidate/corevalidate.go` | validator | transform | same file; `corevalidate_cycle_peer_test.go` | role-match |
| `internal/compiler/originvalidate/originvalidate.go` | validator | transform | same file | role-match |
| `internal/compiler/pathoracle/pathoracle.go` | validator/oracle | transform | same file | partial (must remain distinct from loop proof) |
| `internal/compiler/interp/interp.go` | interpreter | request-response | same file | role-match |
| `internal/compiler/cgen/cgen_program.go` | serializer/backend | transform | same file | role-match |
| `internal/compiler/native/native_app.go`, `cmd/schway/main.go` | runner/route | request-response | same files | role-match |
| `examples/sum_to_n.schway` (or equivalent ordinary source witness) | fixture | request-response | `examples/checksum.schway` | partial (refusal fixture only) |
| `internal/compiler/{syntax,check,core,corevalidate,originvalidate,interp,cgen,native}/..._test.go` | tests | transform/request-response | phase-specific neighboring tests | role-match |
| `internal/compiler/session/session_phase26_test.go` (or equivalent) | integration test | request-response | `internal/compiler/session/session_phase20_test.go`, `session_app_verify_test.go` | role-match |

All named source analogs were confirmed tracked with `git ls-files`; no runtime or ignored mirror paths are used.

## Pattern Assignments

### Frontend: `ast.go`, syntax lexer/parser/formatter (model/utility, transform)

**Analogs:** `internal/compiler/ast/ast.go`; `internal/compiler/syntax/parser.go`; `internal/compiler/syntax/format.go`.

AST declarations retain source spans and separate syntax representation from semantics. `ast.go:69-76`:

```go
type FuncDecl struct {
    Name         string
    Parameter    Parameter
    ReturnOrigin *BorrowOrigin
    ReturnType   TypeRef
    Body         Body
    Span         diagnostic.Span
}
```

Parser entry (`parser.go:41-67`) preserves source/tokens and returns diagnostics alongside the parsed program:

```go
func Parse(source []byte) ParseResult {
    if len(source) > MaxSourceBytes {
        return ParseResult{Diagnostics: []diagnostic.Diagnostic{inputLimit(MaxSourceBytes)}}
    }
    tokens, diagnostics, truncated := lex(source)
    ...
    program := p.parseProgram()
    ...
    return ParseResult{Tree: Tree{Source: append([]byte(nil), source...), Tokens: tokens}, Program: program, Diagnostics: p.diagnostics}
}
```

Formatter (`format.go:8-26`) is token/CST based, retains comments as data, and emits a final newline. Extend tokenization and parser productions while preserving span attribution and stable recovery; add canonical spacing/indentation for `var`, `while`, `if/else`, `<`, and `+` in the existing formatter.

### `internal/compiler/check/check.go` (checker/lowering, request-response + CFG transform)

**Analog:** same file, especially `cfgBackEdgeDiagnostic` (`check.go:2804-2821`).

```go
func cfgBackEdgeDiagnostic(functionID, blockID string, span diagnostic.Span) diagnostic.Diagnostic {
    return diagnostic.Error(
        "check.cfg_back_edge", span,
        fmt.Sprintf("function %q's control-flow graph contains a cycle", functionID),
        diagnostic.Cause{Kind: "cycle_block", Detail: blockID},
    )
}
```

Keep diagnostics structured, stable, source-attributed, and carrying machine-readable causes. The existing cycle refusal is a boundary to refine for admitted scalar loops, not a check to delete. Preserve the checker’s existing producer role, source IDs/spans, and fail-closed behavior when its deterministic analysis bound is exceeded.

### `internal/compiler/core/core.go` (typed IR model, transform)

**Analog:** same file, operation inventory (`core.go:643-700`, `841`). New operations/control facts must be explicit operation kinds/facts and registered in the exhaustive inventories. Existing comments explain why semantically distinct operations receive distinct kinds (e.g. `OpConst`, `OpConstructPayload`) rather than conditionally populated fields. Add only the minimum representation needed for checked scalar operations and explicit CFG control; retain the invariant that source spellings remain in syntax while typed facts live in core.

### Independent validators: `corevalidate.go`, `originvalidate.go`, `pathoracle.go` (validators, transform)

**Analogs:** each target itself; `pathoracle/pathoracle.go:856-869` is a direct refusal precedent:

```go
// backEdgeError is returned when the declared successor relation contains a cycle.
// This package independently re-derives that rejection rather than trusting check.go.
type backEdgeError struct { functionID, blockID string }
func (e *backEdgeError) Code() string { return "pathoracle.cfg_back_edge" }
```

`corevalidate` and `originvalidate` have their own derivation and validation logic; do not import checker-produced loop acceptance facts as authority. Add bounded scalar analysis independently to the appropriate peers. Keep resource, loan, owner, and loan-derived provenance state outside admitted carries; preserve stable peer-specific refusal diagnostics. `pathoracle` currently refuses cycles by design and enumerates bounded acyclic paths: do not repurpose it as the scalar-loop proof. The analog found for the exact scalar fixed-point algorithm is **none**; this is new logic whose lattice and deterministic cap must be designed from the phase research.

### `internal/compiler/interp/interp.go` (interpreter, request-response)

**Analog:** same file, operation dispatch (`interp.go:1229-1245`) and U64 input parsing (`interp.go:606-613`).

```go
switch operation.Kind {
case core.OpConst:
    parsed, err := strconv.ParseUint(operation.ConstU64, 10, 64)
    if err != nil { return Execution{}, fmt.Errorf("operation %q has invalid canonical U64 constant", operation.ID) }
    top.values[operation.TargetID] = value{u64: parsed, isU64: true}
    top.idx++
case core.OpCopy:
    top.values[operation.TargetID] = sourceValue
    ...
}
```

Extend the existing execution state/dispatch to follow CFG edges and evaluate `while`/`if` conditions. Checked addition must detect overflow before writing the target and yield the specified program failure; it must not turn arithmetic failure into a compiler error or a catchable typed source failure.

### `internal/compiler/cgen/cgen_program.go` (native serializer, transform)

**Analog:** same file, `emitProgram` operation switch (`cgen_program.go:193-225`). Existing generation uses explicit operation-kind cases and `execution.Event` site identities. Emit structured CFG control and check U64 overflow explicitly before accepting a sum; C `uint64_t` addition alone wraps. For loop-repeated events, preserve static site identity plus a deterministic dynamic occurrence coordinate. Keep this as the sole production C serializer.

### App route: `native_app.go`, `cmd/schway/main.go` (runner/route, request-response)

**Analogs:** `native/native_app.go:292-306, 379-430`; `cmd/schway/main.go:176-225`.

```go
func (r Runner) RunApplication(parent context.Context, artifactPath, input string, stdout, stderr io.Writer) (RunOutcome, error) {
    outcome, _, err := r.runApplication(parent, artifactPath, input, "", stdout, stderr)
    return outcome, err
}
```

The runner validates bounded input, launches the retained artifact once, and streams stdout/stderr directly while keeping process outcome separate from tool errors and optional evidence capture. The CLI’s `runApplicationRun` parses `schway app run ... -- INPUT` and dispatches through that runner. Reuse this route for the ordinary source witness; keep app input above 1,000 a program-level failure with no stdout, and keep arithmetic overflow status distinct from evidence-capacity exhaustion.

### `examples/sum_to_n.schway` (source witness, request-response)

**Partial analog:** `examples/checksum.schway:28-36` plus refusal pin `session/session_phase20_test.go:14-47`. The checksum source has a loop-shaped idea, but is explicitly not supported semantics; copy neither its `loop` syntax nor its arithmetic helper calls. Use the phase-approved source shape: local mutable U64 counter/total, `while i < n`, assignments with checked `+`, and return/print through existing app conventions. No current ordinary source witness exercises scalar loops; this is a new witness.

### Tests and session evidence (test, transform/request-response)

**Analogs:** `internal/compiler/check/check_test.go` and `check_branch_test.go`; `corevalidate_cycle_peer_test.go`; `interp_test.go`; `cgen_program_test.go`; `native_app_test.go`; `session/session_phase20_test.go`.

Use focused package tests for parser/formatter, typed core invariants, checker refusals, independent validator mutation controls, interpreter behavior, and emitted C contract. Existing `session_phase20_test.go:14-47` shows a source fixture can be read and a stable refusal pinned; successful end-to-end output should instead follow the app route and separately compare each engine to fixed expected values. Include 0/10/1000, rejected 1001, direct near-max overflow, reached wrong-result and zero-iteration controls, and owner/resource/loan/provenance back-edge refusals. Do not use engine agreement alone as the oracle.

## Shared Patterns

### Independent derivation and explicit refusal

**Sources:** `check/check.go:2804-2821`; `pathoracle/pathoracle.go:856-869`; peer-specific validators. Each semantic consumer must derive/validate its own accepted facts. Preserve peer-specific diagnostics for unsupported ownership/resource/loan/provenance carries and fail closed at analysis bounds.

### Source spans, typed facts, exhaustive operation handling

**Sources:** `ast/ast.go:69-76`; `core/core.go:643-700`; `syntax/parser.go:41-67`. Carry source spans into diagnostics; register every new core kind in the operation inventories and downstream exhaustive switches (interpreter, validators, C emitter, evidence serializers).

### Program outcome separated from tool/evidence failure

**Sources:** `native/native_app.go:303-306, 331-357, 379-430`. Application stdout/stderr/status are caller-visible program behavior; build/runtime tool errors and evidence exhaustion are separate outcomes. Preserve that separation for checked overflow and repeated loop events.

## No Analog Found

| File/Concern | Role | Data Flow | Reason |
|---|---|---|---|
| Scalar U64/Bool CFG fixed-point transfer and convergence proof | checker/validator | transform | Current admitted CFG policy refuses cycles; no existing scalar-loop fixed-point implementation exists. |
| `sum_to_n` accepted ordinary source witness | fixture | request-response | `checksum.schway` is only a historical unsupported-frontier example. |

## Metadata

**Analog search scope:** `internal/compiler/{ast,syntax,check,core,corevalidate,originvalidate,pathoracle,interp,cgen,native,session}`, `cmd/schway`, and `examples/`.  
**Files scanned:** 18 primary files plus neighboring tests and source fixture.  
**Tracked-source check:** All analog paths named above are present in `git ls-files`.  
**Pattern extraction date:** 2026-10-02
