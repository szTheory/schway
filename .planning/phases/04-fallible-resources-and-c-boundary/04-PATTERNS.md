# Phase 4: Fallible Resources and C Boundary — Pattern Map

**Mapped:** 2026-09-04
**Files analyzed:** 15 new/modified-file areas from `04-CONTEXT.md`/`04-RESEARCH.md` (D-04-01..D-04-29) plus ROADMAP Phase 4 success criteria SC1-SC4
**Analogs found:** 15 / 15 — every area has an in-tree analog; 12 of 15 are the *same file, widened again* (Phase 4 extends exactly the packages Phase 3 already widened), 3 are genuinely new artifacts with only out-of-tree spike prior art
**Codebase:** ~11,700 lines Go across 23 packages (grew from 8,771 in Phase 3's mapping). All analog paths verified `git ls-files` tracked — no gitignored mirrors.

**Framing note, different from Phase 3's map:** Phase 4 is unusual in that most of its "new work" areas are widenings of files this project already widened once in Phase 3 (`check.go`, `corevalidate.go`, `interp.go`, `cgen.go`, `originvalidate.go`, `pathoracle.go`, `native.go`, `evidence.go`, `session.go`). For those, the closest analog is **the file's own most recent addition** — the shape a Phase 3 case/branch/field took is the shape a Phase 4 case/branch/field should take. Only the C boundary artifacts (frozen foreign TU, conformance TU, `core.ForeignContract`) and the `setjmp` pad are genuinely new, and for those the analog is spike 005's validated prior art, not any in-tree file.

---

## File Classification & Mapping Table

| New/Modified file | Role | Data flow | Closest analog | Match quality |
|---|---|---|---|---|
| `internal/compiler/core/core.go` (+4 `OperationKind`, `core.ForeignContract`, `AllOperationKinds()`) | model | CRUD (schema definition) | itself, `core.go:126-134` (`OperationKind` consts) + `:151-161` (additive-omitempty precedent) | exact |
| `internal/compiler/check/check.go` (foreign decl admission, `unwind`/`nonlocal_exit` no-default refusal, `OpForeignCall`/`OpFail`/`OpRelease`/`OpDefect` emission, reverse-order release materialization) | controller (frontend admission) | request-response (per-binding admission) + CRUD (emits ops) | itself, `checkLinear`/`analyzeStraightLine` seam `check.go:908-980`, `:1012-1160` (binding-kind switch) | exact |
| `internal/compiler/corevalidate/corevalidate.go` (independent re-derivation of release order, `call_target_not_foreign`/`foreign_origin_omitted` refusals, `OpFail`/`OpRelease`/`OpDefect`/`OpForeignCall` cases) | service (independent validator) | CRUD (replay) | itself, `replayStraightLine`/`replayBlocks` `corevalidate.go:730-935` (two switches, `default:` fail-closed) | exact |
| `internal/compiler/interp/interp.go` (two dispatch switches gain 4 cases; terminal-outcome axis) | service (semantic oracle) | event-driven (streamed events) | itself, `runLinear`/`runBranchArm` `interp.go:83-149` | exact |
| `internal/compiler/cgen/cgen.go` (two dispatch switches gain 4 cases; header/conformance-TU/manifest emission; streaming event emitter; `setjmp` pad + static ledger; zero-attribute emission) | service (C17 code generator) | transform + streaming (new emitter path) | itself, `emitLinear`/`emitBranchOperations` `cgen.go:199-292`, `:460-510`; `emitEventSupport` `cgen.go:310-363` | exact (dispatch), role-match (streaming emitter is additive-new) |
| `internal/compiler/native/native.go` (widened `validateExecution`, multi-TU compile, `nm -u` allowlist control, nonzero-exit/signal handling) | service (process-supervision harness) | request-response + file-I/O (process spawn) | itself, `Run`/`validateExecution`/`boundedWriter` `native.go:49-159`, `:243-282` | exact |
| `internal/compiler/originvalidate/originvalidate.go` (walk `{OpReturn,OpFail,OpDefect}`, `foreign_origin_omitted`, wire `ValidatePublished` into `check`/`run`) | service (independent validator) | CRUD (recompute-and-compare) | itself, `RecomputeOriginPerReturn`/`ValidatePublished` `originvalidate.go:94-252` | exact |
| `internal/compiler/pathoracle/pathoracle.go` (walk `{OpReturn,OpFail,OpDefect}` in `linearizePath`) | service (independent test oracle) | CRUD (path enumeration) | itself, `linearizePath` `pathoracle.go:237-283` | exact |
| `internal/compiler/evidence/evidence.go` (`foreign_digest` omitempty field on `Manifest`, `manifestID`) | model + service | CRUD (content-identity digest) | itself, `Manifest` struct `:49-68`, `build` `:207-224`, `manifestID` `:371-410` | exact |
| `internal/compiler/session/session.go` (new required controls: `control:kind.exhaustive_dispatch`, `control:foreign.no_unproven_attributes`, `control:defect.no_release_on_defect`, `control:foreign.unwind_forbidden`, `control:origin.foreign_origin_omitted`, mutation runners for layout + release-omission) | controller (verification gate) | CRUD (lane/control bookkeeping) | itself, `requiredControls` `:930-950`, `OwnedBackendMutationRunner` `:67-113`, `addLane` shapes `:671`/`:823`/`:965` | exact |
| `testdata/phase4/*.lang` fixtures (three-acquisition resource, foreign-decl-missing-unwind, nonlocal-exit probe) | test fixture | CRUD | `testdata/phase2/reborrow_while_moved.lang`, `testdata/phase3/borrowed_view.lang` | role-match |
| `native/` frozen foreign TU (hand-written, byte-frozen, wraps libc `malloc`/`free`) | config/fixture (frozen golden) | file-I/O | `testdata/phase2/owned_transfer.golden.c` (frozen golden precedent) | role-match |
| `lang_foreign_conformance.c` (generated conformance TU) | generated artifact | transform (compile-only, not linked) | no in-tree analog; spike 005 `native/nonlocal_exit.c` (structurally different purpose) + D-04-11's `_Static_assert` spec | no analog — new artifact class |
| process-root `setjmp` landing pad + static cleanup ledger (in `cgen`-emitted C) | generated artifact | event-driven (nonlocal-exit detection) | spike 005 `.planning/spikes/005-native-ffi-provenance-cleanup/native/nonlocal_exit.c` (full file, in `<code_context>` above) | exact (out-of-tree) |
| `scripts/verify-phase4.sh` (new gate script) | config | batch | `scripts/verify-phase3.sh` (full file, read above) | exact |

---

## Pattern Assignments

### 1. `internal/compiler/core/core.go` — new `OperationKind` values + `core.ForeignContract`

**Analog:** itself — the existing `OperationKind` const block and the `LinearBody`'s additive-`omitempty` fields.

**Current shape to extend** (`core.go:126-134`):
```go
type OperationKind string

const (
	OpCopy            OperationKind = "copy"
	OpMove            OperationKind = "move"
	OpBorrowShared    OperationKind = "borrow_shared"
	OpBorrowExclusive OperationKind = "borrow_exclusive"
	OpReturn          OperationKind = "return"
)
```
Add four new consts here (exact names Claude's Discretion — CONTEXT.md's illustrative set is `OpForeignCall`, `OpFail`, `OpRelease`, `OpDefect`). **Also add** a `core.AllOperationKinds() []OperationKind` registry function alongside this block — it does not exist yet and is the backbone of the new `control:kind.exhaustive_dispatch` required control (D-04-22). No existing function of this shape exists in `core.go`; write it as a small literal slice returning all nine consts, in declaration order, with a comment explaining it is the single table every dispatch site is tested against.

**Additive-omitempty precedent to copy verbatim** (`core.go:146-161`):
```go
type LinearBody struct {
	ID         string            `json:"id"`
	Types      []TypeFact        `json:"types"`
	Places     []Place           `json:"places"`
	Operations []LinearOperation `json:"operations"`
	// Blocks, Edges, and LoanEndpoints are additive omitempty Phase 3 facts:
	// a plain straight-line linear body (Phase 2 and earlier) never
	// populates them, so its serialized bytes are unchanged (D-13).
	Blocks        []Block        `json:"blocks,omitempty"`
	Edges         []Edge         `json:"edges,omitempty"`
	LoanEndpoints []LoanEndpoint `json:"loan_endpoints,omitempty"`
}
```
`core.ForeignContract` (a new sibling type, referenced from `Function` the same way `PublicOrigin *PublicOrigin` is — `core.go:40`) must follow exactly this shape: a pointer field, `omitempty`, with a comment stating which pre-Phase-4 functions never populate it (all of them). The `err`/`ok` edge shape for `OpFail`/`OpForeignCall`/`OpRelease` similarly should ride as new `omitempty` fields on `LinearOperation` (`core.go:136-144`) or a new sibling struct — not a change to `Alternatives []string` (D-04-04 forbids this).

**ID convention to inherit** (`03-PATTERNS.md` §3, still binding): function-local semantic ordinals, e.g. `<functionID>:op:<n>`. A `core.ForeignContract`'s own identity fields (allocator name, unwind policy) are declaration data, not ordinal-keyed facts — but any edge/place they reference must still resolve through the same `<functionID>:place:<n>` / `<functionID>:edge:<from>:<to>` conventions `corevalidate.go`'s referential-closure checks already enforce.

---

### 2. `internal/compiler/check/check.go` — foreign decl admission, fallible-op emission, reverse-order release materialization

**Analog:** itself — `checkLinear`/`analyzeStraightLine`'s existing seam and binding-kind switch.

**The seam a new admission gate slots into** (`check.go:908-980`, `checkLinear`):
```go
func checkLinear(module, functionID string, function ast.FuncDecl) (core.Function, []diagnostic.Diagnostic, int) {
	parameterType := coreType(function.Parameter.Type)
	if !sameType(function.ReturnType, function.Parameter.Type) {
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.return_mismatch", ...)}, typeNodeCount(parameterType)
	}
	// OWN-04 borrowed-view precedent for a syntactic-discriminant early gate:
	if function.ReturnOrigin != nil && function.ReturnOrigin.Path != function.Parameter.Name {
		causes := []diagnostic.Cause{...}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("origin.unknown_path", ...)}, typeNodeCount(parameterType)
	}
	...
	if !executableShape(parameterType) {
		causes := []diagnostic.Cause{{Kind: "type", Detail: typeID}, {Kind: "constructor", Detail: parameterType.Constructor}}
		return core.Function{}, []diagnostic.Diagnostic{diagnostic.ErrorWithRepairs(
			"check.unexecutable_shape", function.Parameter.Span, "...", causes,
			diagnostic.Repair{Kind: "use_executable_shape", Detail: "Byte or Buffer"},
		)}, typeNodeCount(parameterType)
	}
	...
}
```
D-04-16's "a foreign declaration missing `unwind`/`nonlocal_exit` policy is refused, no default" is exactly this shape: a syntactic-discriminant early gate returning `diagnostic.ErrorWithRepairs(...)` with causal spans, before any deeper analysis runs — copy the `!executableShape(parameterType)` gate's structure (causes = `{type, constructor}`-style detail pairs; one `Repair`). D-04-02's `call_target_not_foreign` refusal is the same shape again, gated on `OpForeignCall.Callee` resolving into the function-name→ID map `check.go` already builds for ordinary calls/bindings.

**Binding-kind emission switch to extend** (`check.go:1073-1159`, inside `analyzeStraightLine`):
```go
switch binding.RHS.Kind {
case "take":
	...
	kind = core.OpMove
	source.initialized = false
	source.movedAt = spanPointer(binding.RHS.Span)
	source.moveTargetID = target.ID
case "borrow":
	if !hasTypeAbility(typeFact, core.AbilityShare) {
		causes := []diagnostic.Cause{{Kind: "declared_here", Span: spanPointer(source.declared)}, ...}
		return fail(diagnostic.ErrorWithRepairs("ownership.borrow_requires_share", binding.RHS.Span, "...", causes,
			diagnostic.Repair{Kind: "use_take_instead"}))
	}
	if blocking := conflictingLoan(activeLoans[source.place.ID], "shared"); blocking != nil {
		return fail(borrowConflictDiagnostic(binding.RHS.Span, blocking, source.place.ID, source.place.TypeID))
	}
	kind = core.OpBorrowShared
	...
case "borrow_mut":
	... (identical shape, "exclusive")
default:
	...
}
```
New `binding.RHS.Kind` values for `acquire` (only reachable as the operand of `try`/`discard … because`, per D-04-06 — so the parser, not this switch, is where `let x = acquire …` must already be rejected) become new `case` arms here, each following the "gate → set `kind` → mutate loan/place bookkeeping" shape. The **materialized reverse-order `OpRelease`** (D-04-07) is new bookkeeping this switch does not have a precedent for inside a single function — model it on `expiringLoans[loan.lastUse] = append(...)` (`check.go:1129`, `:1158`): accumulate a per-failure-block ordered list of live acquisitions exactly the way loans accumulate per-expiry-index, then emit `OpRelease` operations into each failure block in the reverse of that accumulation order when the block is finalized (the `endLoans(index)` hook in `analyzeStraightLine`'s Phase 2/3 ancestor, described in `03-PATTERNS.md` §1, is the closest existing "fire on block exit" precedent, though Phase 4's fire point is a fail-edge rather than an operation index).

**Diagnostic emission shape to copy exactly** (cause-then-repair ordering, `check.go:1082-1092`):
```go
causes := []diagnostic.Cause{
	{Kind: "borrow_created_here", Span: spanPointer(blocking.borrowedAt)},
	{Kind: "borrow_used_later",   Span: spanPointer(blocking.lastUseSpan)},
	{Kind: "loan",  Detail: blocking.id},
	{Kind: "owner", Detail: source.place.ID},
	{Kind: "type",  Detail: source.place.TypeID},
}
return fail(diagnostic.ErrorWithRepairs(
	"ownership.move_while_borrowed", binding.RHS.Span,
	"cannot transfer ownership while a future-used shared loan is live", causes,
	diagnostic.Repair{Kind: "move_after_last_borrow_use"},
))
```
`core.call_target_not_foreign`, `core.foreign_origin_omitted`, and the missing-`unwind`-policy refusal should all use `diagnostic.ErrorWithRepairs` (repair-bearing → schema `/1`, per 03-PATTERNS I-4) with this exact cause-ordering convention: span-bearing causes first, then ID/name-bearing detail causes.

---

### 2b. `internal/compiler/corevalidate/corevalidate.go` — independent re-derivation + new refusals

**Analog:** itself — `replayStraightLine`/`replayBlocks`' two switches (`corevalidate.go:738-799`, `:870-935`), both ending in the fail-closed anchor.

```go
v.checks++ // dispatch one independently authorized transition
switch operation.Kind {
case core.OpCopy:
	if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
		return false
	}
	...
case core.OpBorrowShared:
	...
case core.OpMove:
	...
case core.OpReturn:
	if index != len(operations)-1 || returned || operation.TargetID != "" || ... {
		return v.check(false, "core.final_claim_mismatch", operation.ID)
	}
	returned = true
default:
	return v.check(false, "core.unknown_operation", string(operation.Kind))
}
```
**This `default:` is the mechanism that makes D-04-22 safe**: a new `OperationKind` a checker emits before `corevalidate` authorizes it produces `core.unknown_operation`, not silent admission. Add `case core.OpForeignCall`, `case core.OpFail`, `case core.OpRelease`, `case core.OpDefect` in **both** switches (straight-line and block-shaped), each following the `v.check(condition, code, detail)` idiom (`corevalidate.go:58-67`, quoted in full in `03-PATTERNS.md` §2 — copy exactly, do not accumulate beyond the first problem).

**Independent re-derivation for D-04-07's release order** — model on the loan-liveness re-derivation already present at `corevalidate.go:837-846` (`replayBlocks`'s `chain.carriedLoans`/`loanLastUse` accumulation): `corevalidate` must walk backward from each failure edge finding every `OpForeignCall` (acquisition) that dominates it and is not yet released, in reverse discovery order, and compare the resulting sequence against the `OpRelease` operations `check` materialized — **never reading `check`'s intermediate data structure**, exactly as `03-PATTERNS.md` §2/Pattern 3 states for the Phase 3 precedent (`corevalidate.recomputeLoanEndpoints` vs. `check.go`'s worklist fixpoint, `reflect.DeepEqual`-compared).

**`core.call_target_not_foreign` and `core.foreign_origin_omitted`** are new `v.check(...)` call sites, same idiom, independently derived from `check`'s (no shared helper, per D-12/D-12a).

---

### 3. `internal/compiler/interp/interp.go` — two dispatch switches gain 4 cases

**Analog:** itself, `runLinear`/`runBranchArm` (`interp.go:83-146`, both switches structurally identical, both ending `default: return Execution{}, fmt.Errorf(...)`):
```go
switch operation.Kind {
case core.OpCopy:
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.copied"))
case core.OpMove:
	delete(values, operation.SourceID)
	values[operation.TargetID] = value
	events = append(events, ownedEvent(function, operation, "value.transferred"))
case core.OpBorrowShared:
	...
case core.OpBorrowExclusive:
	...
case core.OpReturn:
	events = append(events, Event{...})
	return Execution{Schema: execution.Schema1, Outcome: Outcome{Kind: "returned", Value: value}, Events: events, LiveResources: []string{}}, nil
default:
	return Execution{}, fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
}
```
Add `case core.OpForeignCall`, `case core.OpFail`, `case core.OpRelease`, `case core.OpDefect` before each `default:`, in **both** `runLinear` and `runBranchArm`. `OpFail` and `OpDefect` are new **terminators** alongside `OpReturn` — each must produce an `Execution` whose `Outcome.Kind` is `"typed_failure"` or `"defect"` respectively (D-04-08's closed `value | typed_failure | defect` axis), not `"returned"`. `OpRelease` follows the `value.copied`/`value.transferred` non-terminal shape (append an event, continue the loop) but should emit a `resource.released` event kind. `Execution.LiveResources` (currently hardcoded `[]string{}` at every `OpReturn`, `interp.go:102`/`143`) must become a real accounting: a still-live resource at an `OpDefect` populates it, feeding D-04-17's `resource.leaked` requirement.

**05-PATTERNS I-5 still applies and is now sharper:** `interp` performs no ability/loan checks — it trusts the two admission layers. Adding `OpForeignCall` here means `interp` executes a foreign call by whatever mechanism the interpreter models (likely a fixed literal-outcome stub per the fixture's declared behavior, since the interpreter cannot actually call C) — this is Claude's Discretion but must not silently drop the operation from the trace, per I-5's warning.

---

### 4. `internal/compiler/cgen/cgen.go` — dispatch switches, streaming emitter, header/conformance-TU/manifest, `setjmp` pad

**Analog A — dispatch switches to extend** (`cgen.go:237-288`, `emitLinear`; `:460-510`+, `emitBranchOperations`):
```go
switch operation.Kind {
case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
	target, exists := places[operation.TargetID]
	...
	fmt.Fprintf(&out, "  %s %s = %s; /* %s: %s */%s\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
	fmt.Fprintf(&out, "  if (!lang_record_event(%s, %s, %s, %s, %s, %s)) return 74;\n", ...)
	declared[operation.TargetID] = true
case core.OpReturn:
	fmt.Fprintf(&out, "  if (!lang_record_event(...)) return 74; /* returned place: %s */\n", operation.ID)
	out.WriteString("  if (!lang_write_literal(\"{\\\"schema\\\":\\\"lang.execution/1\\\",\\\"outcome\\\":{\\\"kind\\\":\\\"returned\\\",...")
	...
default:
	return "", fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
}
```
`OpForeignCall` follows the transition-emission shape (declare a target, mark it declared) but emits a real C call expression to the `_LANG_`-namespaced extern declaration rather than a plain assignment. `OpFail`/`OpDefect` follow the `OpReturn` terminal shape but write `"typed_failure"`/`"defect"` outcome kinds instead of `"returned"`, using the **new streaming emitter** (see below), not the buffered `lang_write_literal`/`lang_write_events` calls this switch currently uses for `OpReturn`. `OpRelease` is a non-terminal transition emitting a `resource.released` event via `lang_record_event`.

**The `mutationMarker` precedent to reuse for the release-omission control** (`cgen.go:250-253`, and `session.go:81-90`):
```go
// The backend causality control (D-02-07) locates its mutation site by this
// stable generated marker instead of an exact source-derived line...
marker = " /* lang:mutation-site */"
```
D-04-25/Pitfall 2 requires a **separate** marker (e.g. `/* lang:release-site */`) at each emitted `OpRelease` call site, so `session.go`'s omitted-release mutation runner can locate and delete exactly one release emission — modeled directly on `session.OwnedBackendMutationRunner` (below), attacking `cgen`'s own output, never the frozen foreign TU.

**Analog B — the buffered emitter this phase must NOT reuse for defect/nonlocal-exit paths** (`cgen.go:310-363`, `emitEventSupport`, doc comment quoted in full):
```go
// emitEventSupport writes the LANG_EVENT macros, struct, bounded output
// writer, JSON-string escaper, and event recorder shared by every linear-
// shaped emitter (emitLinear and, from Phase 3, emitBranch).
func emitEventSupport(out *strings.Builder, capacity int) {
	fmt.Fprintf(out, "#define LANG_OUTPUT_LIMIT 65536u\n#define LANG_EVENT_CAPACITY %du\n\n", capacity)
	...
	out.WriteString("static int lang_record_event(...) {\n  if (lang_event_count >= LANG_EVENT_CAPACITY) return 0;\n  lang_events[lang_event_count++] = (LANG_EVENT){...};\n  return 1;\n}\n\n")
	out.WriteString("static int lang_write_events(void) {\n  size_t index;\n  for (index = 0u; index < lang_event_count; index++) {\n    ...\n  }\n  return 1;\n}\n\n")
}
```
This buffers (`lang_record_event` appends to `lang_events[]`; `lang_write_events` writes them all at the very end, called once from the `OpReturn` case). D-04-20 requires a **new, additive** function — e.g. `emitStreamingEventSupport` — that writes each event to stdout **as `lang_record_event` is called**, not into an array for later replay, so an `abort()`/`longjmp` mid-function still emits every event up to that point. Existing `emitEventSupport`/`emitLinear`/`emitBranch` stay byte-identical (D-13); the new path is invoked only from functions that declare `foreign` acquisitions.

**Analog C — the `_LANG_` namespace collision-avoidance precedent** (`cgen.go` around `:360`, per `03-PATTERNS.md` I-10): `cName`/`cLocal`/`newCNames.allocate` (`cgen.go:532-591`) is the existing deterministic-suffix-on-collision naming machinery — the generated `_LANG_`-namespaced foreign header (D-04-12a) and the conformance TU must allocate names through this same `cNames` machinery, not a second ad hoc namer.

**No in-tree analog — new artifact classes (spike 005 is the only prior art):**
- The conformance TU (`lang_foreign_conformance.c`) — no existing `cgen` function emits a compiled-but-not-linked TU. Design it as a new top-level `Emit*`-style function (sibling to `Emit`/`EmitNative`, `cgen.go:16`/`:37`) that writes `_Static_assert(sizeof(X) == N, "...")` / `_Alignof` / `offsetof` pairs — Pattern 4 in `04-RESEARCH.md` ("quarantine as an information-flow property") governs which two headers this one function alone may `#include`.
- The process-root `setjmp` pad + static ledger — copy spike 005's `native/nonlocal_exit.c` structure verbatim in spirit (already quoted in full in `04-RESEARCH.md`'s Code Examples section and reproduced in `<code_context>` above): `static jmp_buf`, `static int acquired`/`released`-style counters generalized into a ledger array, one `setjmp` call at `main`'s top, `longjmp`'s target reached only via a foreign callback. **D-04-18: the pad must not run releases** — this is the one place in the whole phase where copying the *release* pattern (reverse-order `OpRelease` calls) would be actively wrong; the pad only reads the static ledger to emit `resource.leaked` events, never calls a release function.

---

### 5. `internal/compiler/native/native.go` — widened `validateExecution`, multi-TU compile, `nm -u` control

**Analog:** itself. The load-bearing current hard-reject (`native.go:243-249`, quoted in full — this is the landmine every other Phase 4 native task depends on clearing first):
```go
func validateExecution(value execution.Execution) error {
	if value.Schema != execution.Schema0 && value.Schema != execution.Schema1 {
		return errors.New("unsupported execution schema")
	}
	if value.Outcome.Kind != "returned" || value.Outcome.Value == "" || value.Events == nil || value.LiveResources == nil || len(value.Events) == 0 || len(value.LiveResources) != 0 {
		return errors.New("execution document violates required contract")
	}
	...
}
```
Add an explicit expected-terminal-outcome parameter/axis (e.g. `validateExecution(value execution.Execution, expected TerminalOutcome)`), branching on `value.Outcome.Kind` in `{"returned", "typed_failure", "defect"}` rather than hard-requiring `"returned"`. **Per Pitfall 3, add a regression test asserting Phase 1-3 documents are still rejected on the old grounds** — i.e. a `"returned"`-expecting call must still reject a document with `len(LiveResources) != 0`, it must simply now *also* accept `"typed_failure"`/`"defect"` shapes when that is what's expected.

**Bounded-process spawn pattern to reuse for multi-TU compilation** (`native.go:68-92`, quoted in full in `03-PATTERNS.md` §9):
```go
ctx, cancel := context.WithTimeout(parent, r.Timeout); defer cancel()
started := time.Now()
command := r.commandContext(ctx, r.ClangPath, "-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic", optimization, sourcePath, "-o", binaryPath)
var compileStdout, compileStderr boundedWriter
command.Stdout = &compileStdout; command.Stderr = &compileStderr
err = command.Run()
if errors.Is(ctx.Err(), context.DeadlineExceeded) { return Result{}, &ToolError{Code: "native.timeout", Err: ctx.Err()} }
if compileStdout.overflowed() { return Result{}, streamError("native.compile_stdout_truncated") }
```
The conformance TU must be compiled as a **separate invocation** of exactly this pattern (own `command`, own `boundedWriter` pair, own timeout context) so its failure is distinguishable from the program's — do not merge it into one `clang` invocation with multiple source files, and do not reuse the same `boundedWriter` variables across the two compiles.

**`nm -u` allowlist control** — no existing precedent in `native.go` (this file has no symbol-table inspection today); model the *shape* on the existing tool-invocation pattern above (bounded, timed, `os/exec`), but the control itself is new: run `nm -u <binary>`, normalize underscore-prefix per Mach-O/ELF, diff against an allowlist, and — per D-04-19 — **must** return `operational`/`tool_missing` (not silently pass) when `exec.ErrNotFound`/`os.ErrNotExist` is returned, exactly like the existing `native.tool_missing` code path already handles a missing `clang` (`native.go:87-90`):
```go
if err != nil {
	code := "native.compile_failed"
	if errors.Is(err, exec.ErrNotFound) || errors.Is(err, os.ErrNotExist) {
		code = "native.tool_missing"
	}
	return Result{}, &ToolError{Code: code, Err: withStderr(err, compileStderr.bytes())}
}
```

**SIGABRT/signal handling (D-04-24)** — no existing precedent (every Phase 1-3 program returns 0 or a small nonzero code from `main`, never a signal). New code must call `command.Run()`, then type-assert `err.(*exec.ExitError).ProcessState.Sys().(syscall.WaitStatus)` and check `.Signaled() && .Signal() == syscall.SIGABRT` — **never** a hardcoded `134`. `boundedWriter`'s existing "always claims full write, never stalls the child" contract (`native.go:142-152`, quoted in full in `03-PATTERNS.md` §9) is unaffected and should be reused as-is for the aborting process's stdout/stderr.

---

### 6. `internal/compiler/originvalidate/originvalidate.go` — widen terminator walk, `foreign_origin_omitted`, wire into `check`/`run`

**Analog:** itself, `RecomputeOriginPerReturn` (`originvalidate.go:94-117`, the package's declared **sole** backward-walk site per its own doc comment):
```go
func RecomputeOriginPerReturn(function core.Function) []ReturnOrigin {
	if function.Linear == nil || len(function.Linear.Operations) == 0 { return nil }
	operations := function.Linear.Operations
	sourceOf := make(map[string]core.LinearOperation, len(operations))
	var returnOps []*core.LinearOperation
	for index := range operations {
		operation := operations[index]
		if operation.Kind == core.OpReturn {
			returnOps = append(returnOps, &operations[index])
			continue
		}
		sourceOf[operation.TargetID] = operation
	}
	...
}
```
**This is D-04-29's single highest-risk site.** Change `operation.Kind == core.OpReturn` to a set-membership test against `{core.OpReturn, core.OpFail, core.OpDefect}` — the loop already collects "terminator" operations into `returnOps` generically; only the discriminant condition needs to widen. `walkReturnOrigin` (`:122-151`) itself is terminator-agnostic (it walks backward from `returnOp.SourceID`) and needs no change beyond accepting the now-wider `returnOps` set. Required control (D-04-29): assert the walked terminator set equals `core.AllOperationKinds()` filtered to terminators; mutation-kill by deleting `OpFail` from this membership test and confirming the differential goes from green to red.

**`ValidatePublished` — the wiring point for WR-01/D-04-27** (`originvalidate.go:215-252`, quoted in full):
```go
func ValidatePublished(program core.Program) []Problem {
	for _, function := range program.Functions {
		recomputedPaths, recomputedAccess, ok := RecomputeOrigin(function)
		if function.PublicOrigin == nil {
			if ok {
				return []Problem{{Code: "core.origin_omitted", Detail: fmt.Sprintf(...)}}
			}
			continue
		}
		...
	}
	return nil
}
```
D-04-27 requires this function be called from `lang check`/`lang run` (currently only `interface export` calls it, per CONTEXT.md) — this is a CLI/session wiring change, not a change to `ValidatePublished` itself; find the `interface export` call site and add an equivalent call in the `check`/`run` command paths, using the same first-problem-only `[]Problem` return shape. D-04-28's `foreign_origin_omitted` is a new sibling check inside this same loop, same `Problem{Code: ..., Detail: ...}` shape, firing when an `OpForeignCall`'s declared borrow/retain contract implies a borrow-derived return that `PublicOrigin` doesn't declare.

---

### 7. `internal/compiler/pathoracle/pathoracle.go` — widen terminator walk

**Analog:** itself, `linearizePath` (`pathoracle.go:237-270`, quoted in full):
```go
for _, blockID := range blockIDs {
	operations := idx.operations[blockID]
	for opIndex, operation := range operations {
		work++
		if operation.Kind == core.OpReturn {
			sawReturn = true
		}
		...
	}
}
```
Same D-04-29 widening as `originvalidate`: change `operation.Kind == core.OpReturn` to the `{OpReturn, OpFail, OpDefect}` membership test. The `sawReturn`/"malformed CFG this oracle refuses to close by default" guard immediately below (`:271-281`) is exactly the "late terminal guard" `03-PATTERNS.md` §6 already warned is needed — do not let it silently accept a path whose only terminator is `OpFail`/`OpDefect` without renaming the `sawReturn` variable's meaning (or its comment becomes actively misleading).

---

### 8. `internal/compiler/evidence/evidence.go` — `foreign_digest` omitempty field

**Analog:** itself — `Manifest`'s existing additive field precedent, `DiagnosticSchema` (`evidence.go:49-68`, quoted in full):
```go
type Manifest struct {
	Schema           string   `json:"schema"`
	...
	DiagnosticSchema string   `json:"diagnostic_schema,omitempty"`
	...
	DigestClaim      string   `json:"digest_claim,omitempty"`
	KnownEscape      string   `json:"known_escape,omitempty"`
}
```
Add `ForeignDigest string \`json:"foreign_digest,omitempty"\`` as a new field, following the exact same pattern `DiagnosticSchema` used when it was added. The `build` function's guarded-upgrade block (`evidence.go:215-224`) is the wiring precedent:
```go
if checked.Program.Schema == core.Schema1 {
	manifest.Schema = Schema1
	manifest.CoreSchema = core.Schema1
	manifest.ExecutionSchema = execution.Schema1
	manifest.DiagnosticSchema = diagnostic.Schema1
	manifest.DigestClaim = DigestClaim
	manifest.KnownEscape = corevalidate.KnownEscape
	...
}
```
`ForeignDigest` should be set inside this same `if` block (or a new Phase-4-specific guard, if a `core.Schema2` bump proves necessary — **D-04-23 says it should not**: keep it additive at `Schema1`) only when the program actually declares a `foreign C {}` block, computed via the existing `digest(...)` helper (`evidence.go:212`, same one that produces `CoreDigest`/`CDigest`).

**`manifestID`'s two-identity-struct fork precedent** (`evidence.go:371-410`, quoted in full): the `/1` anonymous identity struct must gain `ForeignDigest` as a new trailing field (never inserted into the `/0` struct, which must never be touched — this is exactly the "the identity function forks rather than grows" rule `03-PATTERNS.md` §4(e) documents).

---

### 9. `internal/compiler/session/session.go` — new required controls, mutation runners

**Analog A — `OwnedBackendMutationRunner`, the shape for the two new mutation controls** (`session.go:67-113`, quoted in full):
```go
type OwnedBackendMutationRunner struct {
	runner        native.Runner
	mu            sync.Mutex
	optimizations []string
}
const mutationMarker = "/* lang:mutation-site */"
func (r *OwnedBackendMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
	lines := strings.Split(cSource, "\n")
	matched := -1
	for index, line := range lines {
		if strings.Contains(line, mutationMarker) {
			if matched != -1 {
				return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("... marker count is >1, want 1")}
			}
			matched = index
		}
	}
	if matched == -1 {
		return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("... marker count is 0, want 1")}
	}
	mutatedLine := lines[matched] + "\n  lang_value_delivered.bytes[0] ^= 0xffu; /* control: backend runtime causality */"
	...
	return r.runner.Run(ctx, mutated, optimization, inputs)
}
```
Per Pitfall 2, the **release-omission** mutation runner must be a sibling struct following this exact `matched := -1` / exact-one-marker-count / fail-closed shape, but locating `/* lang:release-site */` (cgen's new marker, §4 above) and *deleting* that line rather than mutating a value byte. The **layout** mutation must attack the **frozen foreign TU**, not the emitter's output — a different runner entirely, one that mutates the golden `.c` file's struct field order (not `cSource`), then recompiles the conformance TU and confirms `-Werror`/`_Static_assert` rejects it.

**Analog B — `requiredControls` fail-closed list to extend** (`session.go:930-940`, quoted in full):
```go
requiredControls := []string{
	"control:ownership.use_after_move",
	"control:ownership.move_while_borrowed",
	"control:ownership.transfer_requires_take",
	"control:ownership.move_while_reborrowed",
	"control:ability.forged_copy",
	"control:core.duplicate_operation_id",
	"control:interpreter-o0-o3-owned",
	"control:evidence.core_mismatch",
	"control:backend.runtime_causality",
}
for _, required := range requiredControls {
	if !hasControl(result.Lanes, required) {
		return fail(protocol.StatusInvalid, "verify.control_missing", required)
	}
}
for _, lane := range result.Lanes {
	if lane.RecomputedWork == 0 {
		return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID)
	}
}
```
Add (naming per CONTEXT.md's decisions, spellings Claude's Discretion): `control:kind.exhaustive_dispatch`, `control:foreign.no_unproven_attributes`, `control:defect.no_release_on_defect`, `control:foreign.unwind_forbidden`, `control:core.foreign_origin_omitted`, plus row-level release-order-transposition and release-omission control IDs from the three-acquisition fixture requirement — to whichever `requiredControls` list corresponds to the new `verifyForeignCorpus`-style dispatch (modeled on `verifyBorrowedCorpus`, `session.go:955` on, itself modeled on `verifyOwnedCorpus`). **The zero-work gate is copy-verbatim** — every new lane must carry nonzero `RecomputedWork`.

**Analog C — which `addLane` shape to copy.** `03-PATTERNS.md` I-1 already flagged that the Phase 2 `addLane` (`session.go:823`, hardcodes `Status: "pass"`, loses partial-work evidence on failure) is the wrong one to propagate; Phase 3's `verifyBorrowedCorpus` correctly copied the Phase 1 `VerifyCorpus.addLane` shape instead (`session.go:965`, explicit `status` parameter). **Phase 4's new lanes must do the same** — explicit status parameter, called on both the pass and fail paths.

---

## Shared Patterns (apply to every Phase 4 file)

### Fail-closed default arm
`corevalidate.go:797`/`:933` (`core.unknown_operation`), `interp.go:104`/`:145` (unknown-kind error), `cgen.go:286`/`:507` (unknown-kind error) — every switch on `core.OperationKind` ends in a `default:` that **errors**, never silently ignores. This is the exact mechanism D-04-22's `control:kind.exhaustive_dispatch` exploits: delete a `case` and the differential must go red because the `default:` fires.
**Apply to:** every new `case` added to any of the six dispatch sites — never add a case without confirming the sibling `default:` still exists and still errors.

### Independent re-derivation, never shared helpers (D-12/D-12a, unchanged since Phase 2/3)
`corevalidate.replayBlocks`'s loan-liveness re-derivation (`corevalidate.go:837-858`) computes the same fact `check.go` computes via a materially different mechanism (chain/index closure vs. iterative worklist), then compares. **Apply to:** D-04-07's release-order re-derivation (`corevalidate` must walk backward from failure edges, never read `check`'s materialized list) and D-04-02/D-04-28's foreign-call/origin refusals (independently derived in both `check` and `corevalidate`).

### Additive, omitempty core evolution (D-13/D-04-23)
Every new `core.go`/`evidence.go` field is `omitempty`, with a comment stating which pre-Phase-4 programs never populate it. **Apply to:** `core.ForeignContract`, the `err`/`ok` edge fields, `Manifest.ForeignDigest`, `emitted_attributes: []` (D-04-13 — deliberately a populated-empty field, not an omission, so it is NOT `omitempty` in spirit even though the JSON tag mechanics may still use it for pre-Phase-4 byte-identity; confirm this distinction explicitly in the plan).

### Determinism by explicit sort
`sort.Strings(loanIDs)` (`check.go:1080`) before selecting among multiple candidates. **Apply to:** any Phase 4 code that iterates a map of live acquisitions/loans to decide release order or which conflict to report first — map iteration order must never reach output.

### Diagnostic identity = repair-bearing → schema `/1`
`diagnostic.ErrorWithRepairs` stamps `/1` and folds sorted repair *kinds* into the ID hash; `diagnostic.Error` stamps `/0` (03-PATTERNS I-4). **Apply to:** every new Phase 4 diagnostic code (`core.call_target_not_foreign`, `core.foreign_origin_omitted`, the missing-unwind-policy refusal) — use `ErrorWithRepairs`, not `Error`, to land in the repair-bearing taxonomy per CONTEXT.md's Integration Points.

### Comment the law, not the code
Every non-obvious invariant in this repo carries a paragraph on what breaks if changed (`check.go:1098-1103`'s share-ability gate comment, `corevalidate.go` region documenting the fail-closed anchor, `originvalidate.go:82-93`'s "package's SOLE backward-walk site" doc comment). **Apply to:** D-04-18's "pad must not run releases" and D-04-13's "zero unproven attributes" are exactly this kind of invariant and must be commented at the emission site, not just in CONTEXT.md.

### Bounded process/IO (D-15/D-04-24, unchanged)
`native.boundedWriter`'s max-plus-one buffer + separate `total` counter (`native.go:137-155`, quoted in full above) — **apply to:** the conformance-TU compile's own stdout/stderr capture and the `nm -u` invocation's output capture; do not invent a third bounded-writer shape (03-PATTERNS I-2 already flags two divergent copies — a third would compound the debt).

---

## No Analog Found

| Work item | Role | Data flow | Reason |
|---|---|---|---|
| `core.ForeignContract`'s exact field layout (target layout, allocator identity, capture/retention, unwind obligations) | model | — | No in-tree analog; nearest structural precedent is `core.PublicOrigin` (a small pointer-typed sibling fact on `Function`, `core.go:40/44-54`) for *shape*, but the actual field set is FFI-01-specific and Claude's Discretion per CONTEXT.md. Design by first enumerating exactly what D-04-12 requires as fields, then verify each round-trips through all three inspectable layers (header comment, conformance TU, `lang.foreign/0` sidecar) without any layer inventing an un-sourced fact. |
| The generated conformance TU (`lang_foreign_conformance.c`) itself | generated artifact | transform | No `cgen` function emits a compiled-but-not-linked TU today; spike 005 has no conformance-TU analog either (it validated the concept narratively in its README, not as a generated artifact). Model the *emission machinery* (naming via `cNames`, `_Static_assert` construction) on `cgen`'s existing string-builder emitters, but the `_Static_assert`/layout-proof content itself has no in-tree precedent. |
| `lang.foreign/0` sidecar manifest's exact field layout | model | — | Sibling artifact to `evidence.Manifest`, but a **separate** schema/file per D-04-12(c) — no in-tree analog exists for a second content-addressed sidecar manifest alongside the main one. Follow `evidence.Manifest`'s structural conventions (flat fields, `sha256-v1` digests, a `manifestID`-style identity function) but design the field set fresh. |

---

## Metadata

**Analog search scope:** `internal/compiler/**` (23 packages), `scripts/**`, `testdata/**`, `.planning/spikes/005-native-ffi-provenance-cleanup/`, `.planning/phases/03-borrowed-views-and-cfg-lifetimes/03-PATTERNS.md`, `.planning/phases/04-fallible-resources-and-c-boundary/04-{CONTEXT,RESEARCH}.md`.
**Files read in full or by targeted range this session:** `core/core.go` (full), `interp/interp.go` (full), `native/native.go` (full), `corevalidate/corevalidate.go:730-938`, `cgen/cgen.go:199-402`, `originvalidate/originvalidate.go:76-255`, `pathoracle/pathoracle.go:237-306`, `evidence/evidence.go:1-70,200-270,370-410`, `session/session.go:60-115,671-950,928-958`, `check/check.go:908-1160`, `scripts/verify-phase3.sh` (full).
**Pattern extraction date:** 2026-09-04
