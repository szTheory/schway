# Phase 23: Live Local Allocation and Discharge - Pattern Map

**Mapped:** 2026-09-27  
**Files analyzed:** 30 existing or proposed paths / path groups  
**Analogs found:** 29 / 30 (the observer has no direct implementation analog; several matches are partial or limited)

## File Classification

The proposed Phase 23 filenames below are placeholders where CONTEXT.md and
RESEARCH.md leave exact names to planning. Every named existing analog was
checked with `git ls-files -- <path>` and is tracked source. No ignored
`.gsd/capabilities/` mirror paths are used. `AGENTS.md` is present; it says
there are no project skills, and neither `.codex/skills/` nor
`.agents/skills/` contains a project skill index.

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/compiler/ast/ast.go` | model | transform | same file's `RHS`/`LinearBody` declarations | role-match |
| `internal/compiler/syntax/parser.go` | parser | transform | same file's foreign declarations and binding parser | exact |
| `testdata/phase23/*.lang` | fixture | transform | `testdata/phase4/foreign_acquire_one.lang` and `acquire_three_success.lang` | role-match |
| `internal/compiler/syntax/syntax_test.go` | test | request-response | `TestForeignCallRoundTrips` | exact |
| `internal/compiler/check/check.go` | checker | transform | `checkFallibleLinear`, `checkResourceLifecycle` | role-match; old restricted shape |
| `internal/compiler/check/check_test.go` | test | request-response | foreign lowering and refusal tests | role-match |
| `internal/compiler/core/core.go` | model | transform | `ForeignContract`, `OperationKind`, `LinearOperation` | role-match; operation facts need widening |
| `internal/compiler/corevalidate/corevalidate.go` | validator | transform | foreign contract checks and `checkReleaseOrder` | role-match; existing obligation discovery is insufficient |
| `internal/compiler/corevalidate/corevalidate_test.go` | test | transform | `TestReleaseOrderMutationMatrix` | role-match |
| `internal/compiler/originvalidate/originvalidate.go` | validator | transform | `OpForeignCall` origin derivation | role-match |
| `internal/compiler/originvalidate/originvalidate_test.go` | test | transform | independent validator fixture and import guards | role-match |
| `internal/compiler/pathoracle/pathoracle.go` | utility / validator | transform | bounded `EnumeratePaths` and path replay | role-match |
| `internal/compiler/pathoracle/pathoracle_test.go` | test | transform | `checkedFunction` fixture admission helper | role-match |
| `internal/compiler/interp/interp.go` | interpreter | request-response | `Run`, `OpForeignCall`, `OpRelease` | role-match; model only |
| `internal/compiler/session/session_app_verify.go` | verifier | request-response | source replay refusal and verifier-model case | role-match; preserve the evidence boundary |
| `internal/compiler/cgen/cgen_program.go` | emitter | transform | `emitProgram` and application entry writer | role-match; sole whole-program emitter |
| `internal/compiler/cgen/cgen_program_test.go` | test | transform | whole-program emission tests and pre-serialization refusal assertions | role-match |
| `internal/compiler/session/session.go` | service / verifier | request-response | `BuildApplication` and release mutation seams | role-match |
| `internal/compiler/native/native_app.go` | service / process runner | file-I/O / request-response | `BuildApplication`, `RunApplication` | exact for retained application route |
| `internal/compiler/native/bindings.go` | config / build adapter | file-I/O / transform | `BindingManifest`, `ResolveBindings`, `compileBindings` | exact for explicit local C inputs |
| `cmd/lang/main.go` | CLI controller | request-response | `runApplicationRun` | exact for one-token direct-argv run |
| `examples/phase23/file_byte.lang` (proposed) | example / application | file-I/O / request-response | `examples/phase22/identity.lang` plus Phase 4 foreign fixture syntax | partial |
| `examples/phase23/adapter.c` and `.h` (proposed) | adapter | file-I/O | `native/lang_foreign_resource.c` and its private header | partial only; legacy shim frees before returning |
| `examples/phase23/*.bindings.json` (proposed) | config | transform | `examples/phase22/identity.bindings.json` | exact manifest shape |
| `examples/phase23/README.md` and independent expected-output JSON (proposed) | docs / fixture | request-response | Phase 22 README and `identity.expected.json` | exact |
| `internal/compiler/native/phase23_observer_test.c` (proposed; location open) | test observer | event-driven / file-I/O | no physical-lifetime observer exists; use native subprocess tests only for harness conventions | no direct analog |
| `internal/compiler/{check,corevalidate,originvalidate,pathoracle,interp,cgen,native,session}/*_test.go` | tests | transform / request-response | corresponding Phase 4, Phase 21, and Phase 22 tests listed below | role-match |
| `internal/compiler/session/session_phase23_contract_test.go` and phase contract JSON (proposed) | test / config | transform | Phase 21 discharge contract and contract test | exact schema-test pattern |
| `scripts/verify-phase23.sh` (proposed) | verification script | batch | `scripts/verify-phase6.sh` or `verify-phase4.sh` | role-match |
| `.github/workflows/ci.yml` | config | batch | current `checks` and `evidence-aggregate` host matrices | exact |

## Pattern Assignments

### AST and source parser

**Files:** `internal/compiler/ast/ast.go`, `internal/compiler/syntax/parser.go`,
and source fixtures/tests.  
**Analogs:** `ast.RHS`, `ast.LinearBody`, the existing fallible-call binding
parser, `testdata/phase4/foreign_acquire_one.lang`, and
`TestForeignCallRoundTrips`.

The AST keeps source form narrow and unresolved callee identity for checker
admission. Add only the path-token/owner/use forms Phase 23 requires; do not
make an arbitrary string or path API. Existing binding records retain argument
places, spans, and source order:

```go
// internal/compiler/ast/ast.go:171-190
type RHS struct {
    Kind      string
    Source    string
    Span      diagnostic.Span
    Callee    string
    Arguments []string
    ArgumentSpans []diagnostic.Span
}
```

The parser currently makes ordinary call identity a check-time fact and records
the full binding span. Follow that lossless boundary when adding the new syntax:

```go
// internal/compiler/syntax/parser.go:428-462
source := p.identifier("syntax.expected_binding_source")
if p.peek().Kind == TokenLParen {
    arguments, argumentSpans, end := p.callArguments()
    body.Bindings = append(body.Bindings, ast.Binding{
        Name: name.Text,
        RHS: ast.RHS{Kind: "call", Callee: source.Text,
            Arguments: arguments, ArgumentSpans: argumentSpans, Span: source.Span},
        Span: diagnostic.Span{Start: bindingStart.Span.Start, End: end},
    })
    body.Span.End = end
    continue
}
```

Phase 4 fixture syntax already demonstrates declared C operations, failure
types, and `try` call sites (`testdata/phase4/foreign_acquire_one.lang:13-28`).
Its `Byte` handle contract is historical and must not define the new pointer
owner representation. Add source admission cases beside existing phase fixtures,
then follow `internal/compiler/syntax/syntax_test.go:988-1035` for parse/format/
reparse semantic-token stability.

### Checker lowering and operation-specific facts

**Files:** `internal/compiler/check/check.go` and `check_test.go`.  
**Analogs:** `checkFallibleLinear`, `resolveForeignStep`,
`checkResourceLifecycle`; tests `TestForeignCallLowersToOkAndErrEdges`,
`TestForeignContractCarriesEveryObligation`, and the Phase 4 release-order tests
in `internal/compiler/check/check_test.go`.

The current checker dispatch accepts either one immediately returned fallible
call or the old all-foreign-call lifecycle shape. This is the right locality
for a new exact-shape admission predicate, but not code to extend by simply
adding another first-symbol field:

```go
// internal/compiler/check/check.go:3674-3684
body := function.Body.Linear
if len(body.Bindings) == 1 && body.Bindings[0].RHS.Kind == "try_call" && body.Result == body.Bindings[0].Name {
	return checkForeignTracer(functionID, function, parameterType, derived, typeID, returnFact, work, body.Bindings[0], foreignSymbols, functionNames, dataTypes)
}
if len(body.Bindings) > 0 && body.Result == function.Parameter.Name && everyBindingIsFallible(body.Bindings) {
	return checkResourceLifecycle(functionID, function, parameterType, derived, typeID, returnFact, work, foreignSymbols, functionNames, dataTypes)
}
return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error(
	"check.foreign_call_shape_unsupported", body.Span,
	"this phase supports only a single fallible foreign call immediately returned, or a sequence of try/discard foreign calls whose result is the function's own parameter",
)}, work
```

The resolver is currently restricted to the function's parameter. Phase 23
needs place resolution from the source scope and operation-specific ABI facts:

```go
// internal/compiler/check/check.go:3971-3989
if len(binding.RHS.Arguments) != maxForeignParametersPerSymbol ||
    binding.RHS.Arguments[0] != functionParameterName {
    problem := diagnostic.Error("name.unknown", binding.RHS.Span,
        "foreign call argument must be the function's own parameter")
    return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
}
symbol, isForeign := foreignSymbols[binding.RHS.Callee]
if !isForeign { /* distinguish Lang-function collision from unknown */ }
```

The legacy lowering shows stable IDs, explicit ok/error edges, and reverse-order
cleanup generation. It also demonstrates what must change: it records each
operation's allocator but later attaches one `ForeignContract` from `infos[0]`
to the whole function (`check.go:4513-4518, 4585-4602`). Keep its explicit block
and operation construction style while putting the checked acquire/use/release
declarations on their respective core operations.

### Core schema and independent validation

**Files:** `internal/compiler/core/core.go`,
`internal/compiler/corevalidate/corevalidate.go` and tests.  
**Analogs:** `ForeignContract`, `LinearOperation`, foreign checks in
`Validate`, and `checkReleaseOrder`.

`ForeignContract` is presently function-scoped; `LinearOperation` already has
operation identity, error edges, release-to-acquire identity, and allocator
pairing. Extend the operation record and update JSON/cloning/structural
invariants together, rather than reusing the function's first contract:

```go
// internal/compiler/core/core.go:835-868 (selected existing fields)
type LinearOperation struct {
	ID       string        `json:"id"`
	PointID  string        `json:"point_id"`
	Kind     OperationKind `json:"kind"`
	SourceID string        `json:"source_id"`
	TargetID string        `json:"target_id,omitempty"`
	TypeID   string        `json:"type_id"`
	OkEdgeID            string `json:"ok_edge_id,omitempty"`
	ErrEdgeID           string `json:"err_edge_id,omitempty"`
	ErrTargetID         string `json:"err_target_id,omitempty"`
	ReleasesOperationID string `json:"releases_operation_id,omitempty"`
	Allocator           string `json:"allocator,omitempty"`
}
```

`corevalidate.Validate` clones input before checking and returns the checked
copy. Each new per-operation declaration must be checked from this core-only
view, with C-identifier and ABI/layout safety independently re-derived at the
operation that uses it (`corevalidate.go:205-216, 1215-1313`).

**Important gap for Phase 23:** current `checkReleaseOrder` builds its tracked
acquisition set by scanning the release records that still exist. Deleting all
releases can therefore erase the very obligation it should validate. Do not
copy that discovery rule; seed owners from successful acquire operations and
prove all terminal paths discharge each seeded identity once.

```go
// internal/compiler/corevalidate/corevalidate.go:3248-3253 (legacy seed)
tracked := make(map[string]bool, len(linear.Operations))
for _, operation := range linear.Operations {
    if operation.Kind == core.OpRelease && operation.ReleasesOperationID != "" {
        tracked[operation.ReleasesOperationID] = true
    }
}
```

Extend the existing independent negative controls in
`corevalidate/corevalidate_test.go:1035-1175`. It already mutates allocator
mismatch, order, dropped, and duplicate releases. Add acquisition-seeded
all-releases-removed, wrong-owner, premature, fabricated, and post-use-error
cases, and assert that the intended mutation was reached.

### Origin peer, path oracle, interpreter, and model replay

**Files:** `originvalidate.go`, `pathoracle.go`, `interp.go`, and
`session_app_verify.go`, with their adjacent tests.  
**Analogs:** existing foreign origin derivation, bounded path expansion,
interpreter release/call modeling, and replay refusal rules.

`originvalidate` is deliberately source-blind and currently derives a foreign
return origin from the function contract alias. For Phase 23, inspect the
operation declaration and source place independently:

```go
// internal/compiler/originvalidate/originvalidate.go:321-337
case core.OpForeignCall:
    if derivedAccess == "" && function.ForeignContract != nil {
        switch function.ForeignContract.Alias {
        case "borrow": derivedAccess = "shared"
        case "retain": derivedAccess = "exclusive"
        }
    }
```

The `pathoracle` enumerator preserves declared successor order, refuses back
edges, and caps paths. Retain this bounded independent path mechanism while
replaying owner state, including the error edge after a successful acquire:

```go
// internal/compiler/pathoracle/pathoracle.go:237-267
func EnumeratePaths(functionID, entryBlockID string, idx blockIndex, cap int) ([][]string, error) {
	var results [][]string
	visiting := map[string]bool{}
	var walk func(id string, current []string) error
	walk = func(id string, current []string) error {
		if visiting[id] {
			return &backEdgeError{functionID: functionID, blockID: id}
		}
		visiting[id] = true
		defer delete(visiting, id)
		current = append(current, id)

		block, ok := idx.byID[id]
		if !ok || len(block.Successors) == 0 {
			if len(results) >= cap {
				return &pathCapError{functionID: functionID, limit: cap}
			}
			results = append(results, append([]string(nil), current...))
			return nil
		}
		for _, successor := range block.Successors {
			if err := walk(successor, current); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(entryBlockID, nil); err != nil {
		return nil, err
	}
	return results, nil
}
```

The interpreter validates core before execution and its current foreign-call
path fabricates modeled values; release only updates the model's live set and
emits a semantic event (`interp.go:166-180, 958-990`). New deterministic acquire
success/failure and use success/error outcomes belong here. A use error after
success must leave the modeled owner live until release. This remains model
evidence only.

Keep `session_app_verify.go`'s separation: source replay rejects foreign
outcomes and local C, while verifier-model cases only compare declared values
(`session_app_verify.go:241-267, 360-402`). Do not let replay execute the new
adapter or claim host IO/free; the public retained app and native observer own
those evidence claims.

### Whole-program C emitter

**Files:** `internal/compiler/cgen/cgen_program.go` and
`cgen_program_test.go`.  
**Analogs:** `emitProgramWithShell`, `emitProgramBranchOperation`, the app main
writer, and current refusal tests.

`emitProgramWithShell` is the whole-program emission gate and currently refuses
all foreign contracts and by-pointer lowering before serialization. Keep it the
sole native serialization authority and narrow the newly admitted shape there.
Preserve refusals for unsupported ownership/FFI shapes:

```go
// internal/compiler/cgen/cgen_program.go:583-611 (selected refusal gates)
if shell == programApplicationShell && (entry.Match != nil || entry.Linear == nil || entry.Parameter.Type != "U64" || entry.ReturnType != "U64") {
	return "", fmt.Errorf("application entry %q: only a linear U64-to-U64 entry is supported", entry.ID)
}
if function.Match == nil && (selectsByPointerLowering(function, function.Linear) || selectsByPointerLoweringSharedOnly(function, function.Linear)) {
	return "", fmt.Errorf("function %q: by-pointer bodies are not supported by whole-program native emission this phase", function.ID)
}
if function.ForeignContract != nil {
	return "", fmt.Errorf("function %q: multi-function foreign contracts are not supported by native emission this phase", function.ID)
}
```

The current writer accepts one bounded argv token, parses it as U64, invokes the
entry once, and emits decimal U64 (`cgen_program.go:887-904, 1001-1005`). Extend
only the entry representation and preserve the U64 route. Emit cleanup before
the bounded ordinary failure diagnostic on typed-use error; compiler release
events or generated text are not physical cleanup proof. `emitProgramBranchOperation`
currently rejects `OpForeignCall` / `OpFail` (`cgen_program.go:1435-1438`), so
add explicit checked operation lowering there or in the shared linear emitter,
not a second C backend.

### Session, retained native app, CLI, and manifest

**Files:** `session.go`, `native_app.go`, `bindings.go`, `cmd/lang/main.go`,
and app/manifest integration tests.  
**Analogs:** `BuildApplication`, `BuildApplicationFile`, `RunApplication`,
`ResolveBindings`, `compileBindings`, and `runApplicationRun`.

The session route checks source, validates the independent core, emits, then
hands source/C/manifest to the native builder. Extend this seam rather than
adding a parallel build path (`session.go:1142-1170`). The runner compiles
manifest sources as objects and links them with generated `program.c`; its
receipt records manifest digest, input inventory, compiler, target, flags, and
host ABI (`native_app.go:202-225`).

The app runner passes one opaque argument directly to one process, checks the
receipt/digest, bounds the token, and uses `exec.CommandContext` rather than
shell interpolation:

```go
// internal/compiler/native/native_app.go:379-416 (selected transport checks)
if len(input) > MaxApplicationArgumentBytes {
	return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.input_too_long", Err: fmt.Errorf("argument is %d bytes; limit is %d", len(input), MaxApplicationArgumentBytes)}
}
if actualDigest != receipt.ExecutableDigest {
	return RunOutcome{}, BuildReceipt{}, &ToolError{Code: "native.artifact_digest_mismatch", Err: errors.New("executable bytes do not match the build receipt")}
}
runCtx, cancel := context.WithTimeout(parent, timeout)
defer cancel()
command := r.commandContext(runCtx, artifactPath, input)
command.Stdout, command.Stderr = stdout, stderr
err = command.Run()
```

`BindingManifest` is a closed build-authority record, separate from Lang foreign
authority. Keep local sources/headers/include dirs and typed symbol typedefs;
resolved paths must remain under the manifest root and exact bytes are
snapshotted (`bindings.go:30-45, 224-282`). The compiler emits one binding probe
per declared symbol, compiles those units and listed C sources, then links them
(`bindings.go:515-569`). Add per-operation prototype conformance while keeping
the manifest closed and path-confined.

The CLI currently enforces one token after `--` and calls the retained runner
once. Preserve its option parsing and direct argument route; the app's
generated entry currently enforces its own 4096-byte bound:

```go
// cmd/lang/main.go:183-237
if separator < 0 || separator+2 != len(args) { return exitUsage }
...
outcome, err = native.RunApplication(context.Background(), args[2],
    args[separator+1], os.Stdout, os.Stderr)
```

`cmd/lang/main_test.go:67-117` is the public application test pattern: read an
independent expected-output file, build with the real CLI, run multiple cases,
and assert output, stderr, error exit, argument bound, and exact token count.
`native/bindings_test.go:46-72, 74-110` supplies temporary manifest fixtures,
real Clang build/run, inventory checks, and path/symlink rejection controls.

### C adapter, native observer, and reached controls

**Proposed files:** adapter `.c/.h` beside the new example, a separately built
native observer translation unit, and public-process evidence tests.  
**Analog:** `native/lang_foreign_resource.c` plus its private header; native
process outcome tests in `native_app_test.go`.

The historical C fixture allocates, serializes the address through `uint64_t`,
and frees before returning (`native/lang_foreign_resource.c:20-32, 49-60`). It
is useful only as a small C17/allocator-style example. It is explicitly not a
live owner, uses a target-width-unsafe integer handle, and must not be reused as
the new utility or observer.

```c
/* native/lang_foreign_resource.c:20-32 — historical only */
lang_foreign_resource_block *block = malloc(sizeof(lang_foreign_resource_block));
if (block == NULL) return 0;
block->payload = request;
*out_handle = (uint64_t)(uintptr_t)block;
...
free((void *)(uintptr_t)handle);
```

There is no tracked implementation that observes a real pointer's allocation,
later use, and matching free independently of compiler events. Design the
observer as a new test-only C translation unit. It should record pointer
identity and lifecycle order without double-freeing or dereferencing a
violating pointer, and it must assert omitted, premature, duplicate, and
wrong-resource attempts were reached. `native_app_test.go:319-390` is a
subprocess stream/exit/timeout harness analog, while
`session.go:131-148, 271-287` shows marker-based release mutation mechanics.
Neither existing compiler event ledgers nor that mutation seam can establish
physical cleanup.

### Contract artifact, documentation, CI, and verification script

**Files:** successor phase resource-discharge contract JSON and its test,
`examples/phase23/README.md`, independent expected-output JSON,
`scripts/verify-phase23.sh`, and `.github/workflows/ci.yml`.  
**Analogs:** Phase 21 contract/test, Phase 22 docs/expected fixture, and current
CI host/evidence aggregate.

The Phase 21 contract is tracked and has a strict typed test: decode the JSON,
check schema and required exits/families, then mutate in-memory copies to show
missing, duplicate, unknown, or overclaiming records are rejected
(`session_phase21_contract_test.go:13-100, 133-195`). Add a successor and leave
the archived Phase 21 artifact untouched. State the ownership transition
rules explicitly: borrow preserves, local release consumes, transfer would
preserve under a new owner but is not implemented here.

Use the Phase 22 README for clean-checkout commands and a clear boundary around
what linking and event reports prove (`examples/phase22/README.md:7-47, 52-60`).
Its `identity.expected.json` uses authored expected stdout values rather than
copying a run (`examples/phase22/identity.expected.json:1-7`). The new example
should document `0x41 -> 65`, `0x42 -> 66`, empty/oversized acquisition errors,
and `0x43`'s real post-acquisition typed use error.

For recurring CI, the current checks and evidence-aggregate jobs already run on
both Ubuntu and macOS; the aggregate requires Clang and invokes a focused phase
script (`.github/workflows/ci.yml:35-79`). Add only the focused Phase 23 check
to that host matrix if its maintenance/runtime cost is justified; do not add
another full vet/build/test/race pass.

`verify-phase4.sh`/`verify-phase6.sh` are peer gate scripts, not templates to
fork wholesale. Reuse the portable shell conventions (`set -eu`, a trapped
temporary directory, private Go cache) from `scripts/verify-phase6.sh:1-18`,
and add only Phase 23's independent-app and native-observer checks. Keep exact
macOS/Linux receipts separate; an unavailable host stays incomplete.

## Shared Patterns

### Independent compiler facts

The checker may construct per-operation ownership and ABI facts, but
`corevalidate`, `originvalidate`, and `pathoracle` must derive their own
obligations from core operations/edges. In particular, successful acquire is
the source of the cleanup obligation even if every release is removed.

### One C emission authority and closed build inputs

Keep `cgen.EmitApplication` / `emitProgramWithShell` as the only application
serializer. Keep external C in the closed local manifest; its declarations and
layout must be conformance-compiled against each checked operation signature.
Manifest membership permits compilation/linking, not a semantic guarantee that
arbitrary C obeys its contract.

### Model, event record, and physical behavior remain separate

`interp.Run` models declared foreign outcomes; application compiler events
record generated semantic intent; the separate native observer establishes
real malloc/use/free ordering. Do not let one report claim another evidence
class.

### Public argument and error route

Preserve one bounded opaque argv token through `lang app run` and generated
entry code. The C adapter owns/closes its descriptor, the Lang local owner is
only the returned allocation, and typed failures use the existing nonzero
application outcome with bounded diagnostics.

## No Direct Analog Found

| File/behavior | Role | Data Flow | Reason |
|---------------|------|-----------|--------|
| New test-only native observer translation unit | observer | event-driven / file-I/O | No existing source independently watches a pointer survive adapter return, get used while live, and be freed through generated cleanup. |
| New live-allocation adapter | adapter | file-I/O | The historical resource shim frees in C before return and exposes a pointer through `uint64_t`; it cannot witness a live pointer-bearing Lang owner. |
| Acquisition-seeded release proof | validator | transform | Existing `corevalidate.checkReleaseOrder` discovers tracked owners from existing release records. Phase 23 must change the source of obligations to successful acquire operations. |

## Metadata

**Analog search scope:** `internal/compiler/{ast,syntax,check,core,corevalidate,originvalidate,pathoracle,interp,cgen,session,native}`, `cmd/lang`, `native`, `examples/phase22`, `testdata/phase4`, `scripts`, and `.github/workflows`.  
**Files scanned:** 30 principal implementation, fixture, test, and workflow paths/groups; targeted sibling tests and fixtures were also inspected.  
**Pattern extraction date:** 2026-09-27  
**Verification status:** source inspection only; no tests or implementation commands were run.
