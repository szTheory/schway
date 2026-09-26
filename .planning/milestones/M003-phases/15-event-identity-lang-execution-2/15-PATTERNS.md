# Phase 15: Event Identity (`lang.execution/2`) - Pattern Map

**Mapped:** 2026-09-19  
**Files analyzed:** 15 planned new/modified files  
**Analogs found:** 14 / 15

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/execution/execution.go` | model | transform | `internal/compiler/execution/execution.go` | exact extension |
| `internal/compiler/execution/invocation.go` | utility | transform | `internal/compiler/pathoracle/pathoracle.go` | role-match |
| `internal/compiler/execution/invocation_test.go` | test | transform | `internal/compiler/execution/execution_test.go` | role-match |
| `internal/compiler/executionpeer/executionpeer.go` | service | event-driven | `internal/compiler/pathoracle/pathoracle.go` | partial (independent bounded peer) |
| `internal/compiler/executionpeer/executionpeer_test.go` | test | event-driven | `internal/compiler/pathoracle/pathoracle_test.go` | role-match |
| `internal/compiler/interp/interp.go` | service | event-driven | `internal/compiler/interp/interp.go` | exact extension |
| `internal/compiler/interp/interp_test.go` | test | event-driven | `internal/compiler/interp/interp_test.go` | exact extension |
| `internal/compiler/cgen/cgen_program.go` | service | transform | `internal/compiler/cgen/cgen_program.go` | exact extension |
| `internal/compiler/cgen/cgen_program_test.go` | test | transform | `internal/compiler/cgen/cgen_program_test.go` | exact extension |
| `internal/compiler/cgen/cgen.go` | service | transform | `internal/compiler/cgen/cgen.go` | exact extension |
| `internal/compiler/native/native.go` | service | request-response | `internal/compiler/native/native.go` | exact extension |
| `internal/compiler/native/native_test.go` | test | request-response | `internal/compiler/native/native_test.go` | exact extension |
| `internal/compiler/session/session_phase5_compare.go` | service | transform | `internal/compiler/session/session_phase5_compare.go` | exact extension |
| `internal/compiler/session/session_phase5_compare_test.go` | test | transform | `internal/compiler/session/session_phase5_compare_test.go` | exact extension |
| `internal/compiler/session/session_phase11_differential_test.go` | test | request-response | `internal/compiler/session/session_phase11_differential_test.go` | exact extension |

`testdata/phase11/multi_function_diamond_call.lang` and `testdata/phase07/deep_diamond_acyclic.lang` are existing, tracked evidence inputs. The phase should consume them; no source-fixture change is implied by the locked decisions. Boundary programs may be built in Go test helpers unless a planner deliberately adds tracked fixtures.

## Pattern Assignments

### `internal/compiler/execution/execution.go` and `internal/compiler/execution/invocation.go` (model/utility, transform)

**Analogs:** `internal/compiler/execution/execution.go`; bounded parser/error precedent: `internal/compiler/pathoracle/pathoracle.go`.

**Model and legacy-byte pattern** (`execution.go:8-11,41-57,83-88`):

```go
const (
    Schema0 = "lang.execution/0"
    Schema1 = "lang.execution/1"
)

type Event struct {
    Schema string `json:"schema"`
    ID string `json:"id"`
    // ... legacy optional facts use omitempty
}

func CanonicalBytes(value Execution) ([]byte, error) { return json.Marshal(value) }
```

Add `Schema2` and `Invocation`/`CalleeFunctionID` with `omitempty` tags so direct `json.Marshal` leaves `/0` and `/1` bytes unchanged. Put canonical `E` escaping plus parse/format in `invocation.go`; producer packages call the formatter and the peer calls the parser, but neither imports the other.

**Stable fail-closed bound/error pattern** (`pathoracle.go:56-67,105-120`):

```go
const MaxPaths = 4096

type pathCapError struct { functionID string; limit int }
func (e *pathCapError) Error() string { ... }
func (e *pathCapError) Code() string { return "pathoracle.path_count_exceeded" }
```

Use the same typed-error + `Code()` shape for malformed invocation grammar where useful. Do not reuse `MaxPaths`: Phase 15 needs its own invocation-node bound and `cgen.invocation_path_table_exceeded` code.

### `internal/compiler/executionpeer/executionpeer.go` and tests (service/test, event-driven)

**Analog:** `internal/compiler/pathoracle/pathoracle.go:191-267,378-415`; entry semantics reference `internal/compiler/callgraph/callgraph.go:279-345`.

**Independent read-only index and bounded explicit walk pattern:**

```go
type blockIndex struct {
    byID map[string]core.Block
    operations map[string][]core.LinearOperation
}

func EnumeratePaths(..., cap int) ([][]string, error) {
    var results [][]string
    var walk func(id string, current []string) error
    walk = func(id string, current []string) error {
        if len(results) >= cap { return &pathCapError{...} }
        // deterministic declared-order traversal
    }
    ...
}
```

Create a dedicated `executionpeer` package that imports `core` and `execution`, never `interp` or `cgen`. Independently index functions/operations, resolve the entry with the documented in-degree/closure rule, unfold the invocation DAG up to 4096, then consume `Execution.Events` sequentially with an explicit activation stack. Return named errors containing violation class, event index, invocation, and expected parent/function. Keep static invocation-set equality in a separately named full-coverage control, not the general validator.

**Entry-resolution pattern** (`callgraph.go:285-344`): sort roots, resolve a unique root immediately, then select only a strictly largest reachable closure; otherwise fail closed. Re-derive this inside the peer rather than importing `callgraph` if structural independence requires no shared derivation seam.

### `internal/compiler/interp/interp.go` and tests (service/test, event-driven)

**Analog:** `internal/compiler/interp/interp.go:496-570,619-850`.

**Call admission and preorder seam** (`interp.go:787-825`):

```go
case core.OpCall:
    top.idx++
    result, err := partitionFrameForCall(program, top, operation)
    if err != nil { return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, err) }
    if result.frame == nil { ... }
    if len(stack) >= maxCallDepth() { ... return ... }
    stack = append(stack, *result.frame)
```

Thread an invocation string in `frame` (or an equivalently explicit frame-local field). After successful resolution and depth admission, append `function.called` with the caller frame's invocation, then set the child invocation and push it. This preserves strict preorder and prevents edges for resolution/depth failure. Preserve the immediate-match-arm special case while ensuring its callee event receives the child invocation.

**Uniform event construction pattern** (`interp.go:623-641,846-850`):

```go
func ownedEvent(function core.Function, operation core.LinearOperation, kind string) Event {
    return Event{Schema: execution.Schema1, ID: operation.ID + ":event", Kind: kind,
        FunctionID: function.ID, SourcePlace: operation.SourceID,
        TargetPlace: operation.TargetID, TypeID: operation.TypeID}
}
```

Refactor helpers to accept frame/invocation/schema context, then route every transition, terminal, immediate-return, leak, nonlocal, and depth event through it. Do not patch only `function.returned`; shared leaf transitions also collide under ID-only identity.

### `internal/compiler/cgen/cgen_program.go` and `internal/compiler/cgen/cgen.go` (service, transform)

**Analogs:** `internal/compiler/cgen/cgen_program.go:128-198,236-302,329-436`; C event support: `internal/compiler/cgen/cgen.go:1667-1713`.

**Preflight-before-emission pattern** (`cgen_program.go:128-164`):

```go
order, err := callgraph.Order(program)
if err != nil { return "", err }
entry, err := callgraph.EntryFunction(program)
if err != nil { return "", err }
// validate/collect all functions before strings.Builder serialization
```

Immediately after callgraph/entry validation, build a deterministic static invocation-node table. Append node one for the entry and check before every append; reject the attempted 4097th node with a typed error carrying entry, `limit: 4096`, `observed_at_least: 4097`, and the first overflow call context. Only allocate/render static tables after successful preflight.

**Threaded generated-function pattern** (`cgen_program.go:236-245,363-425`):

```go
for index := range functions {
    fmt.Fprintf(&out, "static %s %s(%s);\\n", typeNames[index], functionNames[index], typeNames[index])
}
...
fmt.Fprintf(out, "static %s %s(%s %s) {\\n", typeName, functionName, typeName, locals[parameter.ID])
...
emitCall(out, calleeTypeName, locals[target.ID], calleeName, locals[source.ID], operation)
```

Extend every program prototype, definition, entry call, and `OpCall` C call with an unsigned invocation-table index. For an `OpCall`, emit `function.called` before the generated child call and use a literal static parent-index × call-site child-index table, not a globally constant source-call index or runtime string building.

**Event-buffer/writer extension pattern** (`cgen.go:1667-1713`):

```c
typedef struct LANG_EVENT {
  const char *kind; const char *id; const char *function_id;
  const char *source_place; const char *target_place; const char *type_id;
} LANG_EVENT;
...
static int lang_record_event(const char *kind, const char *id, ...)
static int lang_write_events(void) { ... }
```

Extend this one support block with literal invocation and optional callee fields, and write `lang.execution/2` only for the multi-function event-identity path. Preserve legacy `/0` and `/1` writer bytes and JSON field order exactly.

### `internal/compiler/native/native.go` and tests (service/test, request-response)

**Analog:** `internal/compiler/native/native.go:553-622`.

**Schema-peek and version-gated validation pattern:**

```go
if value.Schema != execution.Schema0 && value.Schema != execution.Schema1 {
    return errors.New("unsupported execution schema")
}
seenIDs := make(map[string]struct{}, len(value.Events))
for index, event := range value.Events {
    if event.Schema != value.Schema || event.ID == "" || event.FunctionID == "" { ... }
    if _, duplicate := seenIDs[event.ID]; duplicate { ... }
    switch event.Kind { ... }
}
```

Add an explicit `/2` branch: require nonempty canonical invocation, require `callee_function_id` for `function.called`, use `(invocation, id)` as the duplicate key, and admit the new kind only in `/2`. Keep `/0`/`/1` ID-only uniqueness and all existing branches byte/semantic compatible; unknown schema and unclassified kinds remain fail-closed.

### `internal/compiler/session/session_phase5_compare.go`, its tests, and `session_phase11_differential_test.go` (service/test, transform/request-response)

**Analogs:** `internal/compiler/session/session_phase5_compare.go:64-143,172-283`; `internal/compiler/session/session_phase11_differential_test.go:300-340`.

**Ordered all-engine comparison and explicit field routing:**

```go
for index := 0; index < length; index++ {
    ...
    case left.Events[index] != right.Events[index]:
        return &Phase5EngineDisagreement{Axis: AxisEventOrder,
            OperationID: left.Events[index].ID, ...}
}

var Phase5ComparedComparisonFields = []string{
    "Execution.Events.Schema", "Execution.Events.ID", ...
}
```

Because event comparison is whole-struct equality, add both new fields to `Phase5ComparedComparisonFields`; retain the exhaustive reflection-routing test so neither can become silently ignored. Integrate `executionpeer` as an independent session gate for valid `/2` documents and retain faults in opposite directions: a corrupted producer must be refused by the intact peer, and a corrupted peer acceptance/derivation seam must make a deliberately bad document wrongly pass until restored.

**Existing inverted diamond control to flip, not delete:**

```go
seen := make(map[string]bool, len(interpreted.Events))
duplicated := false
for _, event := range interpreted.Events {
    if seen[event.ID] { duplicated = true }
    seen[event.ID] = true
}
if !duplicated { t.Fatalf(...) }
```

Replace this old debt assertion with a four-tier successful `/2` comparison that asserts distinct `(Invocation, ID)` pairs and preorder `function.called` edges across interpreter, `-O0`, `-O3`, and `-O3 -flto`. Add removal assertions against only the `function.called` projection, never the total event count.

## Shared Patterns

### Determinism and explicit traversal

**Sources:** `internal/compiler/callgraph/callgraph.go:295-345,449-510`; `internal/compiler/interp/interp.go:643-651`.

Use sorted roots / declared operation order plus explicit stacks. Never depend on map iteration for traversal, emitted table order, error selection, or event order.

### Fail-closed, named diagnostics

**Sources:** `internal/compiler/pathoracle/pathoracle.go:105-125`; `internal/compiler/native/native.go:553-590`.

Every malformed invocation, unknown `/2` kind/schema, duplicate pair, peer ownership failure, traversal exhaustion, and native preflight overflow needs a stable named refusal. A missing fact is never a permitting default.

### Legacy protocol isolation

**Sources:** `internal/compiler/execution/execution.go:41-57,83-88`; `internal/compiler/cgen/cgen_program.go:286-300`.

The shared struct is marshalled directly and the existing C writer emits literal fields. Version-gate additions and retain omission behavior; add canonical-byte tests before changing either producer.

### Tests as non-inertness proofs

**Sources:** `internal/compiler/pathoracle/pathoracle.go:69-103`; `internal/compiler/session/session_phase11_differential_test.go:300-340`.

Existing test seams are narrowly scoped and restored with cleanup. Follow that convention for producer/peer faults and cgen preflight bypass; test actual bad behavior when the seam is enabled, not merely that a test flag changes.

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| `internal/compiler/executionpeer/executionpeer.go` | service | event-driven | No existing peer validates an execution-event causal document; `pathoracle` supplies the closest independent, bounded traversal shape. |

## Metadata

**Analog search scope:** `internal/compiler/{execution,interp,cgen,native,session,callgraph,pathoracle}`, `testdata/{phase07,phase11}`  
**Files scanned:** 18  
**Tracked-source gate:** verified every named analog with `git ls-files`; no runtime/plugin mirror paths used.  
**Pattern extraction date:** 2026-09-19
