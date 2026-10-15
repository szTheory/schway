# Phase 36: Scalar Helper Calls — Pattern Map

**Mapped:** 2026-10-14  
**Files analyzed:** 26 anticipated targets (12 implementation files, 11 focused test files, 3 test-data files). Test and test-data names below are suggested from the Phase 36 selectors and local phase naming; RESEARCH.md does not pin every filename.  
**Closest tracked analogs:** verified with `git ls-files` for every existing source and test path named below.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match |
|---|---|---|---|---|
| `internal/compiler/ast/ast.go` | model | transform | same-file `Expr` / `RHS` | partial |
| `internal/compiler/syntax/parser.go` | utility (parser) | transform | same-file scalar expression parser | role |
| `internal/compiler/check/check.go` | service (checker) | transform | same-file signature table and call resolver | exact |
| `internal/compiler/check/scalar.go` | service (lowerer) | transform | same-file scalar expression builder | exact |
| `internal/compiler/corevalidate/corevalidate.go` | service (validator) | transform | same-file independent `OpCall` replay | exact |
| `internal/compiler/originvalidate/originvalidate.go` | service (validator) | transform | same-file scalar transfer and freshness pass | role |
| `internal/compiler/pathoracle/pathoracle.go` | service (oracle) | transform | same-file call-aware loan/path analysis | role |
| `internal/compiler/interp/interp.go` | service (interpreter) | event-driven | same-file call frames and execution events | role |
| `internal/compiler/executionpeer/executionpeer.go` | service (peer validator) | event-driven | same-file static call topology and event ordering | role |
| `internal/compiler/cgen/cgen_program.go` | service (C generator) | transform | same-file static call preflight and event capacity | role |
| `internal/compiler/cgen/cgen.go` | service (C emitter) | event-driven | same-file call emission and event recorder | role |
| `internal/compiler/surfacemanifest/manifest.go` | config (surface manifest) | transform | same-file call rules and restrictions | exact |
| `internal/compiler/syntax/parser_phase36_test.go` (suggested) | test | transform | `syntax/parser_phase34_recovery_test.go` | role |
| `internal/compiler/check/check_phase36_scalar_calls_test.go` (suggested) | test | transform | `check/check_phase35_numeric_test.go` | role |
| `internal/compiler/corevalidate/corevalidate_phase36_test.go` (suggested) | test | transform | `corevalidate/corevalidate_call_type_internal_test.go` | role |
| `internal/compiler/originvalidate/originvalidate_phase36_test.go` (suggested) | test | transform | `originvalidate/originvalidate_phase35_test.go` | role |
| `internal/compiler/pathoracle/pathoracle_phase36_test.go` (suggested) | test | transform | `pathoracle/pathoracle_phase35_test.go` | role |
| `internal/compiler/interp/interp_phase36_test.go` (suggested) | test | event-driven | `interp/interp_phase29_events_test.go` | role |
| `internal/compiler/executionpeer/executionpeer_phase36_test.go` (suggested) | test | event-driven | `executionpeer/executionpeer_phase35_test.go` | role |
| `internal/compiler/cgen/cgen_phase36_test.go` (suggested) | test | event-driven | `cgen/cgen_phase273_test.go` | role |
| `internal/compiler/session/session_phase36_calls_test.go` (suggested) | test | request-response | `session/session_phase29_interp_fixture_test.go` and `session/session_phase29_native_fixture_test.go` | role |
| `internal/compiler/session/session_phase36_prechange_test.go` (suggested) | test | batch | `session/session_phase34_prechange_test.go` | exact |
| `internal/compiler/session/session_phase36_surface_test.go` (suggested) | test | transform | `session/session_phase35_surface_test.go` | exact |
| `testdata/phase36/prechange-corpus.json` (suggested) | test | batch | `testdata/phase35/prechange-corpus.json` | exact |
| `testdata/phase36/surface-witnesses.json` (suggested) | test | transform | `testdata/phase35/surface-witnesses.json` | exact |
| `testdata/phase36/surface-census.json` (suggested) | test | transform | `testdata/phase35/surface-census.json` | exact |

## Pattern Assignments

### AST and parser: `ast/ast.go`, `syntax/parser.go`

There is no existing recursive scalar call expression. Extend the scalar expression node and retain spans/source identity using the existing AST conventions. Keep the new named call in scalar-expression parsing; preserve the existing 64-level expression bound and delimiter-aware recovery. Do not change legacy binding-call grammar.

**AST analog** — `internal/compiler/ast/ast.go:222-226`:
```go
type Expr struct {
    Kind        string // name, number, text, add, less, byte_len, byte_at
    Name, Value string
    Left, Right *Expr
    Span        diagnostic.Span
}
```

**Recovery-test analog** — `internal/compiler/syntax/parser_phase34_recovery_test.go:104-116`:
```go
parsed := syntax.Parse([]byte(draft.Source))
if !bytes.Equal(parsed.Tree.Bytes(), []byte(draft.Source)) {
    t.Errorf("%s recovery changed the lossless source bytes", draft.ID)
}
got := observePhase34Recovery(parsed)
if reason != "" {
    t.Errorf("%s recovery mismatch: diagnostics=%+v siblings=%v", draft.ID, got.Diagnostics, got.RecoveredSiblingNames)
}
```

### Source call resolution and ordered lowering: `check/check.go`, `check/scalar.go`

Use the existing closed module merge, signature lookup, exact arity/type checks, and stable callee IDs. Add scalar-expression call lowering to the scalar builder; lower argument work into ordered temporary operations before emitting the existing `OpCall`. Keep call binding facts independent from result-value propagation.

**Call-binding analog** — `internal/compiler/check/check.go:5094-5148`:
```go
if len(binding.RHS.Arguments) != 1 {
    return core.LinearOperation{}, core.Place{}, &diag
}
argumentName := binding.RHS.Arguments[0]
argument, ok := places[argumentName]
if !ok { return core.LinearOperation{}, core.Place{}, &diag }
if contract, isFunction := calleeContracts[binding.RHS.Callee]; isFunction {
    if !callArgumentTypeCheckSeam &&
        (argumentTypeFact.Shape.Constructor == "" || contract.ParameterType == "" || argumentTypeFact.Shape.Constructor != contract.ParameterType) {
        return core.LinearOperation{}, core.Place{}, &diag
    }
}
```

**Existing core contract:** `core.LinearOperation` carries `SourceID`, `TargetID`, `Kind`, and `TypeID`; `CalleeID` is the direct target. Reuse this shape from `internal/compiler/core/core.go:899-906,943-958`; do not add a call schema.

### Independent raw-call contract: `corevalidate/corevalidate.go`, `originvalidate/originvalidate.go`

Corevalidate resolves the callee from its own function set, checks visibility and signature, consumes an initialized compatible argument, and enforces a fresh target before marking it initialized. Keep these predicates inside corevalidate; do not call checker helpers.

**Corevalidate analog** — `internal/compiler/corevalidate/corevalidate.go:3078-3108`:
```go
if !v.check(operation.CalleeID != "", "core.callee_id_missing", operation.ID) { return false }
if !v.check(declaredFunctionIDs[operation.CalleeID], core.CallCalleeUnresolved, operation.CalleeID) { return false }
if callee, ok := functionByID[operation.CalleeID]; ok {
    if !v.checkVisibility(*function, callee) { return false }
    if !v.checkCallTypeContract(callee, source, operation, places, types) { return false }
}
if !v.consumeCallArgument(operation, source.TypeID, types, initialized) { return false }
if !v.targetMatches(function, index, operation, places, produced) { return false }
initialized[operation.TargetID] = true
```

Originvalidate owns a distinct scalar abstract transfer. A valid call result becomes initialized but value-unknown unless an existing local law proves otherwise. Its phase-35 freshness and mutation tests are the closest test pattern; add call-target freshness and loop-transfer controls there.

**Origin transfer analog** — `internal/compiler/originvalidate/originvalidate.go:491-513`:
```go
case core.OpRemainderU64:
    if right.initialized && right.known && right.u64 == 0 {
        *zeroOperationID = operation.ID
        return false
    }
    if !source.initialized || !right.initialized { return false }
    value := originScalarValue{initialized: true}
    if source.known && right.known {
        value.known, value.u64 = true, source.u64%right.u64
    }
    state[operation.TargetID] = value
```

### Loan/path boundaries: `pathoracle/pathoracle.go`

Preserve the oracle’s narrow provenance behavior: scalar/call result targets do not inherit local loan ancestry, while the call still counts as a source-loan use. Update only the scalar-kind/CFG trigger and explicit event allowances required by the new scalar calls; keep owner/back-edge restrictions intact.

**Path-oracle analog** — `internal/compiler/pathoracle/pathoracle.go:229-239`:
```go
if operation.Kind == core.OpCopy || operation.Kind == core.OpCall || isLoanFreeScalarValueOperation(operation.Kind) {
    placeLoans[operation.TargetID] = nil
} else {
    placeLoans[operation.TargetID] = inherited
}
```

### Dynamic execution identity: `interp/interp.go`, `executionpeer/executionpeer.go`, `cgen/cgen_program.go`, `cgen/cgen.go`

Keep static graph/cycle preflight bounded. At runtime, assign each child activation the next ordinal for its static call site within its dynamic parent; reset counters for a fresh parent frame. Event `Occurrence` continues to count repeated events within one activation. Interpreter, C capture, and executionpeer must derive the progression independently. Bound dynamic identity/path/event storage before counter wrap or allocation.

The current interpreter appends `InvocationSegment{OpCallID: operation.ID, Ordinal: 0}` at `internal/compiler/interp/interp.go:1393-1405`; the peer has the existing `MaxInvocations = 4096` static traversal bound at `internal/compiler/executionpeer/executionpeer.go:14-16`. The canonical formatter already accepts nonnegative ordinals in `internal/compiler/execution/invocation.go:10-49`; no schema edit is implied.

**C recorder analog** — `internal/compiler/cgen/cgen.go:771-776`:
```c
if (schway_event_count >= SCHWAY_EVENT_CAPACITY) { schway_event_overflow = 1; return 1; }
```

The conformance recorder branch uses `fputs("schway: event capacity exhausted\n", stderr); exit(69);` at the same source range. The generator chooses the capture-status or conformance branch from `applicationEvidence`.

The C path preflight and `invocationEventCapacity` in `internal/compiler/cgen/cgen_program.go:173-199,409-435` are the existing static bound to preserve. Extend the dynamic reservation without using 4096 as a semantic call-count limit.

### Native application capture: `native/native_app.go`

The runner already keeps the child process result in `outcome` and evidence status in `report.CaptureStatus`. Keep this distinction when repeated calls exhaust optional capture.

**Capture analog** — `internal/compiler/native/native_app.go:362-395`:
```go
if report.CaptureStatus == EvidenceStatusCapacityExhausted {
    return outcome, report, &ToolError{Code: "native.evidence_capacity_exhausted", Err: errors.New("application evidence exceeds the capture limit")}
}
return outcome, report, nil
```

### Surface and regression evidence: `surfacemanifest/manifest.go`, session tests

Update the public call/signature/context/refusal contract and the independent Phase 36 witness/census data. Follow the session Phase 35 pattern: a clean baseline, witness checks, source-branch census, and bidirectional controls. Freeze the no-call emitted-C corpus before compiler edits; retain emitted and non-emitted rows and arm byte-drift/omission/false-exemption controls.

**Manifest analog** — `internal/compiler/surfacemanifest/manifest.go:176-190`:
```go
Calls: CallRules{
    DeclarationArity: 1,
    OrdinaryArity: []int{1},
    ScalarExpressionCalls: false,
    Constraints: []string{"Ordinary calls require a checked signature and acyclic call graph."},
},
```

**Literal-C baseline analog** — `internal/compiler/session/session_phase34_prechange_test.go:384-407,415-440` compares every enumerated row, requires emitted and non-emitted outcomes, mutates an emitted C byte, and verifies restoration.  
**Manifest-control analog** — `internal/compiler/session/session_phase35_surface_test.go:55-85,88-118,120-156` checks independent witnesses/census and rejects both missing and invented claims.  
**Capacity analogs** — `internal/compiler/cgen/cgen_phase273_test.go:37-80` distinguishes conformance exit 69 from app capture overflow; `internal/compiler/native/native_phase33_capture_test.go:17-43` pins exact-limit and limit-plus-one behavior while preserving process exit.

## Shared Patterns

- **Call identity and visibility:** Resolve targets with the merged-program IDs and existing call signature table in `check/check.go`; preserve explicit export/private checks and the whole-program recursion refusal.
- **Argument order:** Materialize nested and sibling scalar work into explicit operations in source order before helper entry. An argument defect prevents the call event, later operands, and store.
- **Independent checking:** Checker lowering, corevalidate raw contract, originvalidate scalar transfer, each engine’s numeric result, and executionpeer causal structure are separate claims. Do not share a producer verdict or add numeric replay to executionpeer.
- **Freshness and loops:** A call result target is a fresh static temporary; writes to mutable locals remain `OpScalarStore`. Re-execution of one static operation in a loop does not make its target statically reusable.
- **Capture status:** Preserve conformance capacity exit 69 and app `capacity_exhausted` status separately from application streams and process outcome.

## No Exact Analog Found

| Behavior | Closest Existing Pattern | Planner Guidance |
|---|---|---|
| Recursive named-call scalar AST node and expression lowering | Existing recursive scalar operators plus ordinary binding calls | Add only one-argument call atoms; preserve current depth/recovery and lower to existing `OpCall`. |
| Per-parent/per-static-site dynamic activation ordinals | Ordinal-capable invocation grammar and current engines that emit ordinal zero | Introduce checked counters in each dynamic frame/activation; do not identify loop calls by event `Occurrence`. |

## Metadata

**Analog search scope:** `internal/compiler/{ast,syntax,check,corevalidate,originvalidate,pathoracle,interp,execution,executionpeer,cgen,native,surfacemanifest,session}`, `testdata/phase35`.  
**Tracked-source gate:** All existing analog source and test paths named above were confirmed by `git ls-files -- <path>`; no ignored runtime mirror paths are referenced.  
**Research boundary:** Phase 37 file CLI, complete file CRC, second importer, and guide are deferred; Phase 38 owns final two-host binding.
