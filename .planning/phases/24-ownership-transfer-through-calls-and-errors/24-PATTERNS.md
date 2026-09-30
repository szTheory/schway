# Phase 24: Ownership Transfer Through Calls and Errors - Pattern Map

**Mapped:** 2026-09-30  
**Files analyzed:** 14 planned change surfaces  
**Analogs found:** 14 / 14 (several are intentionally narrow predecessors, not implementations of Phase 24 behavior)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/check/check.go` | checker / lowering | request-response, control-flow transform | same file: `checkResourceLifecycle` | role-match |
| `internal/compiler/core/core.go` | model / IR | transform | `LinearOperation`, `ForeignOperationContract` definitions | exact |
| `internal/compiler/corevalidate/corevalidate.go` | validator | transform / path validation | `checkReleaseOrder`, `peerCalleeFrameDrained` | exact for independent validation, local-only |
| `internal/compiler/originvalidate/originvalidate.go` | validator | transform / call-graph derivation | `localOwnerOriginProblem`, `BuildCalleeOriginFacts` | role-match |
| `internal/compiler/pathoracle/pathoracle.go` | validator / oracle | path validation | `ValidateLocalOwnerPaths` | exact for owner proof shape, local-only |
| `internal/compiler/interp/interp.go` | interpreter | request-response / call-frame lifecycle | `frame`, `partitionFrameForCall`, `runFrameStack` | exact for dynamic calls, role-match for owner transfer |
| `internal/compiler/cgen/cgen_program.go` | emitter | transform / native serialization | `emitProgramWithShell`, local file-byte emitter | exact authority, narrow shape |
| `internal/compiler/native/phase23_observer_test.go` | evidence test | file I/O / native process | Phase 23 observer fixture and mutation controls | exact |
| `internal/compiler/native/native_app_test.go` | evidence test | file I/O / app build-run | generated release/use witness | exact |
| `internal/compiler/session/session_phase23_model_test.go` | evidence test / replay fixture | deterministic model request-response | Phase 23 model-only evidence vocabulary | exact |
| `examples/phase23/file_byte.schway` | example program | file I/O via foreign calls | current source witness | exact |
| `examples/phase23/adapter.c` | foreign adapter | file I/O / allocation | bounded acquire/use/release adapter | exact |
| `examples/phase23/README.md` | documentation | evidence/reporting | public app and evidence-boundary contract | exact |
| `scripts/verify-phase23.sh`, `.github/workflows/ci.yml` | config / evidence runner | batch | focused Phase 23 groups and host lane | exact |

No project-local `.codex/skills/` or `.agents/skills/` indexes were present. All analog paths below were confirmed tracked with `git ls-files`; no runtime mirrors are cited.

## Pattern Assignments

### `internal/compiler/check/check.go` (checker / lowering, control-flow transform)

**Analog:** `internal/compiler/check/check.go`

`checkFallibleLinear` dispatches bounded shapes and rejects all others at lines 3648-3695. `checkResourceLifecycle` at lines 4548-4721 is the closest lowering template: it emits operation facts, explicit blocks/edges, then materializes reverse cleanup from a forward completion list.

**Core pattern** (lines 4652-4708):

```go
var completed []resourceStep
releaseOps := func(from []resourceStep) []core.LinearOperation {
    ops := make([]core.LinearOperation, 0, len(from))
    for j := len(from) - 1; j >= 0; j-- {
        acquired := from[j]
        ops = append(ops, core.LinearOperation{
            Kind: core.OpRelease, SourceID: acquired.okPlaceID,
            ReleasesOperationID: acquired.callOpID,
            Allocator: acquired.symbol.Allocator,
        })
    }
    return ops
}
```

**Guidance:** Preserve deterministic accumulation rather than map iteration. For Phase 24 derive cleanup from each return/error edge's currently owned identities, distinguish a failed acquire (no resource) from completed acquisitions, and represent move/return as owner-place changes without synthesizing release/reacquire. Keep source-attributed refusal in checker admission for copy, moved-from use, and owner result at entry.

### `internal/compiler/core/core.go` (IR model, transform)

**Analog:** `ForeignOperationContract` lines 204-218 and `OperationKind` / `LinearOperation` lines 643-678, 851-930.

**Existing contract excerpt** (lines 204-218):

```go
type ForeignOperationContract struct {
    Symbol, ABIType, Mode string
    ParameterType, ResultType string
    Fails, Allocator, Release string
    Unwind, NonlocalExit string
}
```

`LinearOperation` already encodes call identity and owned source/target places; `OpMove`, `OpReturn`, `OpFail`, and `OpCall` are established operation kinds. Add only the activation/transfer facts required for stable semantic identity and independent derivation; keep the per-operation foreign contract attached to the operation and avoid host pointer addresses.

### `internal/compiler/corevalidate/corevalidate.go` (independent validator, paths)

**Analogs:** `checkReleaseOrder` lines 3357-3460; `peerCalleeFrameDrained` lines 3256-3342.

**Reverse cleanup pattern:** `checkReleaseOrder` independently walks backward from terminal blocks over declared edges; its expected ordering comes from graph history, not checker-supplied release order. `peerCalleeFrameDrained` seeds obligation discovery from acquisition operations and accepts a value returned through a pure move/copy chain as a transfer out of that frame.

**Guidance:** Extend this independent derivation across calls and activations. Do not make release operations the sole premise for whether an acquisition must be discharged. Validate transfer receiver and caller/callee responsibilities separately, and retain mutation-resistant path checks.

### `internal/compiler/originvalidate/originvalidate.go` (independent validator, call graph)

**Analog:** `BuildCalleeOriginFacts` lines 161-194, `OpCall` derivation lines 311-372, and `localOwnerOriginProblem` lines 565-675.

`OpCall` origin derivation resolves callee facts by `CalleeID` rather than trusting a caller-provided result fact. `localOwnerOriginProblem` is a source-blind operation-contract check; lines 597-655 track owner identity by acquisition operation ID and owner place in the current local shape.

**Guidance:** Re-derive only the ownership/origin facts this peer owns from program bodies and declared callee contracts. The local owner map is a useful shape, but Phase 24 must qualify resource keys by activation and preserve identity across `OpMove`/`OpCall`/`OpReturn`; do not trust checker event/release lists.

### `internal/compiler/pathoracle/pathoracle.go` (path oracle, request-response)

**Analog:** `ValidateLocalOwnerPaths` lines 59-145.

**Owner state pattern** (lines 82-90, 111-134):

```go
type owner struct {
    place string
    result string
    contract *core.ForeignOperationContract
    borrowed bool
    released bool
}
owners := map[string]owner{}
ownerByPlace := map[string]string{}
```

The current validator refuses blocks/edges and insists on exactly one acquire/borrow/consume sequence. For Phase 24 use the explicit block/edge graph and dynamic activation-qualified resource identity; keep this oracle independently implemented from `check` and `corevalidate`.

### `internal/compiler/interp/interp.go` (interpreter, dynamic calls)

**Analog:** `frame` lines 534-590; `childInvocation` / `configureChild` lines 782-799; `partitionFrameForCall` lines 817-897; call-stack execution at `runFrameStack` (around line 690 onward).

**Activation pattern** (lines 782-799):

```go
segments := append(append([]execution.InvocationSegment{}, f.segments...),
    execution.InvocationSegment{OpCallID: operation.ID, Ordinal: 0})
invocation, err := execution.FormatInvocation(f.entryID, segments)
```

Frames already carry invocation segments, a caller return target, their own place namespace, and per-frame resource bookkeeping. `partitionFrameForCall` removes non-copyable arguments from the caller and seeds the callee parameter. Reuse this activation path for resource identity, transfer the same identity to the receiver, and drain only obligations still owned by the exiting frame on admitted typed-error edges. Keep `ModelResult.ActualHostIO` and `PhysicalCleanup` false.

### `internal/compiler/cgen/cgen_program.go` (native emitter, serialization)

**Analog:** `emitProgramWithShell` lines 574-640 and `validateLocalFileByteFunction` / `emitProgramLocalFileByteFunction` lines 1058-1177.

`emitProgram` is the sole serializer; whole-program callgraph ordering and unsupported-shape refusals happen before serialization. The local emitter validates all three operation contracts and then emits a fixed acquire/use/release sequence. Extend this preflight and lowering path rather than adding another serializer. Keep `emitProgram` the production authority, validate every operation's foreign contract, and preserve fail-closed refusal for unadmitted shapes. Native generated events must not be treated as proof of physical destruction.

### `internal/compiler/native/phase23_observer_test.go` (evidence test, file I/O)

**Analog:** public lifecycle table lines 31-77 and reached mutation table lines 79-160.

The observer records allocation/use/free by actual pointer identity and checks `outstanding=0`; mutations cover omitted, premature, duplicate, and wrong-resource release while checking compiler events remain plausible. Extend fixtures to repeated activations and reverse-order typed-error cleanup; add identity-collision/reached controls per D-24-08. Expected receipts should assert the full allocation/use/free sequence and error boundary.

### `internal/compiler/native/native_app_test.go` (evidence test, app request-response)

**Analog:** `TestPhase23GeneratedReleaseFollowsBorrowedUse`, lines 28-70, and neighboring binding/acquisition tests.

This test compiles checked core to the ordinary app shell and checks generated ordering plus same-owner use/release arguments. Extend the public application witness so a helper returns the owner, caller uses/releases it, and 0x41/0x42 still independently yield 65/66.

### `internal/compiler/session/session_phase23_model_test.go` (model evidence test)

**Analog:** lines 16-49.

The test builds independent replay expectations and explicitly asserts model-only vocabulary (`actual_host_io=false`, `physical_cleanup=false`). Add deterministic cases for admitted call/return and typed-error operation shapes without turning modeled outcomes into claims of host IO or physical cleanup.

### `examples/phase23/file_byte.schway` and `examples/phase23/adapter.c` (example / foreign adapter)

**Analogs:** source contract lines 7-44; adapter acquire failure/success lines 62-128, borrowed use lines 131-143, and release lines 145-147.

Retain the one-byte input contract and existing operation-local ABI declarations. Split or extend source entries only within the same bounded application boundary. The C adapter is already explicit about allocation failure (no owner result) and consuming release; no new runtime or general FFI is implied.

### `examples/phase23/README.md`, `scripts/verify-phase23.sh`, `.github/workflows/ci.yml` (docs / evidence config)

README lines 25-56 separates independent expected answers, public error behavior, physical observer claims, and model-only claims. The verifier lines 45-81 partitions source admission, independent peers, model-only replay, native observer, and public contract groups; CI lines 80-103 runs it on Ubuntu and macOS.

Extend those same evidence groups and host lanes for transfer, repeated activation, and typed-error cleanup. Preserve the checkout's no-local-test boundary; this map does not execute any checks. Avoid adding a duplicate expensive suite lane.

## Shared Patterns

### Activation-qualified identity

**Sources:** `internal/compiler/interp/interp.go` lines 782-799; `internal/compiler/core/core.go` lines 851-930.  
**Apply to:** checker/core facts, validators, interpreter, emitter evidence, observer mapping.

Build identity from static acquisition operation plus deterministic dynamic invocation path (or an equivalently stable shared activation key). Owner places are mutable locations, not resource identities. Never publish a host pointer as the portable semantic ID.

### Independent ownership derivation

**Sources:** `internal/compiler/corevalidate/corevalidate.go` lines 3357-3460; `internal/compiler/originvalidate/originvalidate.go` lines 565-675; `internal/compiler/pathoracle/pathoracle.go` lines 59-145.  
**Apply to:** all peer validators.

Each peer re-derives its obligations from core contracts, operations, paths, and calls. Do not share the checker's computed cleanup schedule as proof or erase the obligation when a candidate release is removed.

### Cleanup edge scope and order

**Sources:** `internal/compiler/check/check.go` lines 4652-4708; `internal/compiler/corevalidate/corevalidate.go` lines 3357-3460.  
**Apply to:** normal return and typed `OpFail` handling only.

Track successful acquisition completion dynamically; release still-owned resources in reverse completion order. Do not interpret defects, process termination, unwind, cancellation, callbacks, or nonlocal transfer as guaranteed cleanup paths.

### Evidence boundary

**Sources:** `internal/compiler/interp/interp.go` lines 120-170 (model result contract); `internal/compiler/native/phase23_observer_test.go` lines 31-160; `examples/phase23/README.md` lines 39-56.  
**Apply to:** session/model reports, native observer, docs, CI receipts.

Interpreter results prove deterministic modeled semantics only. Physical cleanup requires the independent native observer on the actual generated binary and host, including zero outstanding allocations before normal or typed-error exit.

## No Analog Found

No exact existing implementation supports owning transfer through a helper plus repeated acquisition activations followed by typed-error cleanup. Phase 24 is the first admitted shape crossing this boundary. Use the narrow Phase 23 local-owner patterns above as components, while preserving existing refusal defaults until each checker, peer, interpreter, and emitter path agrees on the expanded contract.

## Metadata

**Analog search scope:** `internal/compiler/{check,core,corevalidate,originvalidate,pathoracle,interp,cgen,native,session}`, `examples/phase23`, `scripts`, `.github/workflows`  
**Files scanned:** 14 principal implementation/evidence paths (bounded to closest predecessor seams)  
**Pattern extraction date:** 2026-09-30
