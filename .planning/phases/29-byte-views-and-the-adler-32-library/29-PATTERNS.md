# Phase 29: Byte Views and the Adler-32 Library - Pattern Map

**Mapped:** 2026-10-11
**Files analyzed:** 38 new/modified files (grouped below)
**Analogs found:** 36 / 38 (2 partial: ByteView struct ABI, spike 008 env-gated harness)

All analog paths are git-tracked source (`examples/phase23` verified with `git ls-files`; the rest are tracked Go sources under `internal/compiler/`). No `.gsd` capability mirrors are involved. Line numbers are from this session's reads; re-check if `session.go` or `check.go` move first.

Decisions D-14..D-21 are honored: parameter is named `bytes` (not `data`, a keyword); `fn main(case: U64) -> U64` with `try` acquire on `case`; ByteView abilities are `{drop}` only; interpreter OOB is a defect outcome (exit-65 is a projection); defect with live owner is abort-like and accepted only for this defect class; `-O3`/LTO application lanes are a test-only seam; spike 008 is two-stage.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/compiler/core/core.go` (add `OpBytesLen`, `OpBytesAt`; `AllOperationKinds`) | model/registry | transform | `OpRemainderU64` at core.go:704, :859 | exact |
| `internal/compiler/core/core_test.go` (22 -> 24; fixture list) | test | transform | `TestAllOperationKindsRegistered` core_test.go:799; `exhaustiveDispatchFixtures` :893 | exact |
| `internal/compiler/syntax/parser.go` (`scalarExprAtom`, `scalarLetStart`, `scalarBodyStart`) | parser | transform | `isWriteCallStart` + `scalarWriteStatement` parser.go:368, 440 | role-match |
| `internal/compiler/syntax/format.go` + goldens | utility | transform | existing comma/paren rendering format.go:372 | role-match |
| `internal/compiler/ast/ast.go` (`Expr` kind doc comment, reuse `Name/Left/Right`) | model | transform | `Expr{Kind:"remainder"}` parser.go:518 | exact |
| `internal/compiler/check/scalar.go` (admit one ByteView param; two new arms; lowering) | service | transform | `OpRemainderU64` arms scalar.go:217, 619 | exact |
| `internal/compiler/check/check.go` (reserved names, arity refusals; fixture-main dispatch + hand-built core) | service | request-response | `check.call_arity_unsupported` check.go:4200; `checkLocalFileByteEntry` :3961 | exact / role-match |
| `internal/compiler/ability/ability.go` (`case "ByteView"`) | model | transform | `case "FileByteOwner"` ability.go:157 | exact |
| `internal/compiler/corevalidate/corevalidate.go` (op law, scalar fixed point, replays, `deriveAbility`, release-vs-view law, lifecycle variant) | service | transform | `OpRemainderU64` sites 1237, 1386, 2597, 2928, 4303, 4439, 4539; `case "FileByteOwner"` :4698 | exact |
| `internal/compiler/originvalidate/originvalidate.go` (`originHasScalarCFG`, place gate, transfer arms, `phase27ArithmeticProblem`-style problem, owner matchers) | service | transform | `OpRemainderU64` sites 295, 307-319, 456, 1385-1425 | exact |
| `internal/compiler/pathoracle/pathoracle.go` (fixture-owner replay + release-vs-view law) | service | event-driven | `ValidateLocalOwnerPaths` pathoracle.go:100-135 | exact (name-pinned copy) |
| `internal/compiler/interp/interp.go` (`value.bytes`, two arms, OOB defect, owner binding with real bytes, defect-with-live-owner) | service | transform | `OpRemainderU64` arm interp.go:1406; `RunWithForeignOutcomes` :196-260 | exact / role-match |
| `internal/compiler/cgen/cgen_program.go` (`schema2ExecutionDocumentSize`, `hasScalarOperations`, `emitProgramScalarFunction` both shells, ByteView typedef, fixture-main emitter) | service | transform | `OpRemainderU64` sites 234, 2523, 2703; `emitProgramLocalFileByteFunction` :1773 | exact / role-match |
| `internal/compiler/cgen/cgen.go` (`linearInput` ByteView type name; `linearFixedNames`) | utility | transform | existing `SCHWAY_BUFFER` typedef handling | role-match |
| `internal/compiler/executionpeer/executionpeer.go` (checked-bytes-at defect arm) | service | transform | `checkedRemainderZero` executionpeer.go:368, 497 | exact |
| `internal/compiler/session/session.go` (`defectReasons` map; fixture-owner admission in `BuildApplication`; interpreter binding entry) | service | request-response | session.go:532; :1401-1425; `RunInterpreterApplication` :1153 | exact / role-match |
| `examples/phase29/adler32.schway` + vector fixtures (template) | fixture | request-response | `examples/phase23/file_byte.schway`; `examples/sum_to_n.schway` | role-match |
| `examples/phase29/fixture_bytes_adapter.c` / `.h` | adapter | file-I/O-free generator | `examples/phase23/adapter.c` / `adapter.h` | role-match |
| `examples/phase29/fixture_bytes.bindings.json` | config | config | `examples/phase23/file_byte.bindings.json` | exact |
| `examples/phase29/README.md` | doc | n/a | `examples/phase23/README.md` | role-match |
| New `*_phase29_test.go` per package (syntax, check, corevalidate, originvalidate, pathoracle, interp, cgen, executionpeer, session, native) | test | request-response | `session_phase26_overflow_test.go`; `session_phase23_model_test.go`; `native/phase23_observer_test.go` | role-match |
| `native` test-only application-shell `-O3`/`-O3 -flto` helper | test utility | batch | `applicationFlags` native_app.go:115; `check.scalarAnalysisBudgetForTest` scalar.go:32; `Runner.LTO` native.go:87 | role-match |
| `.planning/spikes/008-byteview-cap/README.md` + `MANIFEST.md` row | doc | batch | `.planning/spikes/007-loan-across-branch/README.md`; `CONVENTIONS.md` | exact |
| Spike 008 env-gated Go test (`SCHWAY_SPIKE008=1`) | test | batch | spike 007 reuses a production test; `native.Runner.Run` timing | partial |
| Register edits: `session_phase28_manifest_test.go:232`, `session_phase272_identity_test.go:34`, `session_phase273_identity_test.go`, `.planning/LANGUAGE-MATURITY.md` count, `session_phase16_emitter_inventory_test.go` | test register | config | Phase 28 `examples/phase28/unused_function.schway` entries | exact |
| `29-VALIDATION.md`, ROADMAP dated notes (SC1 param name; runnable-gain wording) | doc | n/a | prior phase VALIDATION docs | role-match |

## Pattern Assignments

### Core registry: `core/core.go`, `core/core_test.go` (model, transform)

**Analog:** the `OpRemainderU64` registration.

- Constant at `core.go:704`: `OpRemainderU64       OperationKind = "remainder_u64"`. Add `OpBytesLen OperationKind = "bytes_len"`, `OpBytesAt OperationKind = "bytes_at"` beside it.
- Registry literal `core.go:859` (22 entries today) must append both kinds.
- `core_test.go:799`: `const declaredCount = 22` -> `24`.
- `exhaustiveDispatchFixtures` (`core_test.go:893`) is a literal path list; add a scalar-only ByteView fixture (adler32 plus trivial `main`, no foreign block). Every kind must be produced by a fixture that goes through `session.Check`, corevalidate, interp, cgen, pathoracle, originvalidate.
- Do NOT add a `LinearOperation` field; reuse `SourceID` (view place), `RightID` (index place), `TargetID`.

---

### `syntax/parser.go` (parser, transform)

**Analog:** `isWriteCallStart` / `scalarWriteStatement` (name-recognized builtin with lookahead and arity problem).

**Name-recognition precedent** (`parser.go:368-373`):
```go
func (p *parser) isWriteCallStart(token Token) bool {
	if token.Kind != TokenIdentifier || p.peekNonTrivia(1).Kind != TokenLParen {
		return false
	}
	return token.Text == "write_text" || token.Text == "write_u64" || token.Text == "write_newline"
}
```

**Arity problem precedent** (`scalarWriteStatement`, parser.go ~440-455): `p.problem("syntax.write_u64_arity", p.peek(), "write_u64 requires one U64 operand")`. Copy the shape for `syntax.byte_len_arity` / `syntax.byte_at_arity` (codes are discretion).

**Atom to extend** (`parser.go:528-541`, bare-name today, no lookahead):
```go
func (p *parser) scalarExprAtom() *ast.Expr {
	leftToken := p.advance()
	switch leftToken.Kind {
	case TokenNumber: ...
	default:
		if leftToken.Kind != TokenIdentifier { p.problem("syntax.expected_scalar_operand", ...) }
		return &ast.Expr{Kind: "name", Name: leftToken.Text, Span: leftToken.Span}
```
Add, in the identifier branch, `if leftToken.Text == "byte_len" && p.peek().Kind == TokenLParen` (consume `(`, name, `)`) and the `byte_at` branch (`(` name `TokenComma` `p.scalarExpr()` `)`). Reuse `ast.Expr{Kind, Name, Left, Right}`; no new fields. Binary operator kinds are produced at parser.go:509-520 (`"add"`, `"less"`, `"equal"`, `"remainder"`) and consumed by name in `check/scalar.go:expression`.

**Gates to widen only if library bodies may begin with `let`** (`parser.go:342-366`): `scalarLetStart` ends `return next != TokenLParen && next != TokenDot` (that is what refuses `let x = byte_len(`); `scalarBodyStart` gates on `parameterType.Constructor == "U64"`. A `var`-first body is already routed to the scalar parser by `p.atAny(TokenVar, TokenIf, TokenWhile)`. Either widen both or document/test that library bodies start with `var`.

**Comma note:** the lexer yields `TokenUnknown` for `,`; `normalizeOwnershipTokens` rewrites it (parser ~140-152). No lexer change.

**Trap (Pitfall 1):** `data` is a keyword (lexer.go:10-26). Use `bytes`.

**Grammar trap:** no parentheses; `%` binds tighter than `+` (precedence `<`/`==`=1, `+`=2, `%`=3). One reduction per statement; scalar body must end in a bare identifier.

**Formatter** (`format.go`): token projection; `, ` at format.go:372. Expect no code change; add goldens (`byte_at(bytes, i)`; one space after comma). A `let x = byte_at(...) + 1` may break the line after `)` because `callBinding` is set.

---

### `check/scalar.go` (service, transform)

**Analog:** the `OpRemainderU64` arm (fixed-point) and lowering.

**Signature gate to widen** (`scalar.go:35-37`):
```go
if function.Parameter.Type.Constructor != "U64" || (function.ReturnType.Constructor != "U64" && function.ReturnType.Constructor != "Unit") {
	return core.Function{}, []diagnostic.Diagnostic{diagnostic.Error("type.scalar_signature", function.Span, "scalar CFG functions require U64 input and either U64 or Unit output")}, 1
}
```
Admit exactly one `ByteView` parameter; return stays U64/Unit. This keeps `-> ByteView`, `Buffer` params refused with `type.scalar_signature` (neighbor table).

**Type-fact order** (`scalar.go:40-49`) appends `U64`, `Bool`, `Unit` in a loop and the parameter place via `b.addPlace(function.Parameter.Name, "U64", false)` (:50). For a ByteView parameter, `:type:0` must be ByteView (published abilities read `function.ID+":type:0"`, originvalidate.go:2279). Keep the existing U64 order unchanged so no core hash moves.

**Fixed-point arm to copy** (`scalar.go:217-229`):
```go
case core.OpRemainderU64:
	right := state[op.RightID]
	if !source.initialized || !right.initialized {
		return work, "check.scalar_uninitialized", ""
	}
	...
	state[op.TargetID] = fact
```
`OpBytesLen`: result unknown U64 (`checkerScalarFact{initialized: true}`); `OpBytesAt`: same, index must be initialized; never `known` (value is dynamic). A `default` arm (scalar.go:261) otherwise yields `check.cfg_back_edge`.

**Lowering to copy** (`scalar.go:608-637`, "add/less/equal/remainder" case): `b.addPlace("scalar", resultType, false)`, `b.addOperation(core.LinearOperation{Kind: kind, SourceID: left, RightID: right, TargetID: place.ID, TypeID: b.types[resultType]})`, `b.initialized[place.ID] = true`. For `byte_at` build `SourceID` = view place, `RightID` = index place, result U64. Refuse a ByteView name used anywhere else with the existing codes (`type.scalar_operand`, `type.scalar_binding` scalar.go:424/453, `type.scalar_result` :79).

**`scalarAuthorityCategory`** (`scalar.go:280`): add `"ByteView"` next to `"Loan", "Borrowed", "View"`.

**Library statement shape** (flag (e)): `var`/assignment only (`OpScalarStore`, event-free); `let` lowers to `OpCopy` (evented, exhausts capacity).

---

### `check/check.go` (service)

**Reserved names / arity (D-04):** follow `check.call_arity_unsupported` at `check.go:4200`:
```go
diag := diagnostic.Error("check.call_arity_unsupported", binding.RHS.Span, "this phase supports only arity-1 calls", causes...)
```
Add named refusals for user `fn byte_len` / `fn byte_at` (reserved) where functions are collected.

**Fixture-owner main dispatch (D-05/D-15).** Analog: `checkFallibleLinear` dispatch and `checkLocalFileByteEntry`.

Dispatch (`check.go:3699-3706`) keys on the parameter type `PathToken` (`parameterType.Constructor == "PathToken"`). For `fn main(case: U64)` with a fixture owner, add a new branch keyed by the fixture owner's acquire symbol, before falling through to the generic try handling that otherwise refuses with `check.foreign_call_shape_unsupported`.

Foreign policies are already generic (`collectForeignSymbols`, check.go:3610-3660): `mode`, `abi`, `release`, `allocator`, `fails`, `alias` are all captured, so `foreign C { ... mode: "acquire" ... }` parses and collects with no change.

Hand-built core to copy (`checkLocalFileByteEntry`, check.go:3961-4020; shape-gate then literal assembly):
```go
if !foreignSignature(acquire, "PathToken", "FileByteOwner") || !foreignSignature(use, "FileByteOwner", "U64") || !foreignSignature(release, "FileByteOwner", "Unit") {
	return refusal("check.local_owner_signature_mismatch", ...)
}
if !localForeignPolicy(acquire, "acquire", "schway_file_byte_acquire_fn", "AcquireError") || acquire.Allocator != "libc_malloc" || acquire.Release != release.Name || ... {
	return refusal("check.local_owner_contract_invalid", ...)
}
...
operations := []core.LinearOperation{
	{ID: acquireID, PointID: functionID + ":point:linear:0", Kind: core.OpForeignCall, SourceID: places[0].ID, TargetID: places[1].ID, TypeID: types[1].ID, Allocator: acquire.Allocator, Foreign: localForeignContract(acquire)},
	{ID: functionID + ":op:1", ..., Kind: core.OpForeignCall, SourceID: places[1].ID, TargetID: places[2].ID, ..., Foreign: localForeignContract(use)},
	{ID: functionID + ":op:2", ..., Kind: core.OpRelease, SourceID: places[1].ID, TypeID: types[1].ID, ReleasesOperationID: acquireID, Allocator: release.Allocator, Foreign: localForeignContract(release)},
	{ID: functionID + ":op:3", ..., Kind: core.OpReturn, SourceID: places[2].ID, TypeID: types[2].ID},
}
```
For Phase 29 the list becomes `[acquire(case U64 param) -> FixtureBytes, borrow -> ByteView, OpCall adler32(view) -> U64, release, return]`. Recommendation from RESEARCH: construct the `OpCall` from the callee-contract table (`resolveCallBinding`, check.go:4189-4290) rather than forging the operation, so argument/return type gates still run. The release stays compiler-minted from the acquire's `release:` policy (not source syntax).

Names are hard-coded in Phase 23 (`PathToken`, `FileByteOwner`, `schway_file_byte_*`, `AcquireError`, `UseError`, `libc_malloc`). This is a new bespoke closed shape; do not edit the Phase 23/24/25 matchers.

**Call admission (flag (a)):** `buildCallSignatureTable` (check.go:2010) + `verifyCallableRefusal` (:2067) read `Callable`, which originvalidate computes (`originvalidate.go:2131`); see originvalidate section. `executableShape` (check.go:5869): leave ByteView OUT so a straight-line ByteView function stays `check.unexecutable_shape`.

---

### `ability/ability.go` and `corevalidate.deriveAbility` (model)

**Analog:** `case "FileByteOwner"` (`ability.go:157`):
```go
case "FileByteOwner":
	if len(shape.Arguments) != 0 {
		return derivedShape{}, fmt.Errorf("FileByteOwner takes no type arguments")
	}
	return derivedShape{
		set:       abilitySet{drop: true, share: true, send: true, escape: true},
		witnesses: map[core.Ability][]string{core.AbilityCopy: {"FileByteOwner"}},
	}, nil
```
For D-17 `ByteView` = `{drop}` only: `set: abilitySet{drop: true}` and witnesses for `copy`, `share`, `send`, `escape` each `{"ByteView"}`. Add a `case "FixtureBytes"` modeled on FileByteOwner (drop/share/send/escape, `copy` denied).

Independent twin at `corevalidate.go:4698-4705`:
```go
case "FileByteOwner":
	if len(shape.Arguments) != 0 { return false, nil, false }
	if requested == core.AbilityCopy { return false, []string{"FileByteOwner"}, true }
	return true, nil, true
```
Write the ByteView/FixtureBytes entries independently (the two derivations must agree; no shared helper).

---

### `corevalidate/corevalidate.go` (service, transform)

**Analog:** every `OpRemainderU64` site. Each is a distinct copy.

- **Structural per-op law** (`corevalidate.go:1237-1246`): the `core.scalar_operation_invalid` check on source/right/type constructors and absent store/edge fields. Add a `case core.OpBytesLen, core.OpBytesAt` with types `ByteView`/`U64`/`U64`, `OpBytesLen` has `RightID == ""`, `OpBytesAt` has a `U64` right, target is not the parameter, `StoreTargetID`/edges empty.
- **Kind-exclusive field law** (`corevalidate.go:1386`): `RightID` is legal only on add/less/equal/remainder; extend to `OpBytesAt` (else `core.scalar_field_kind_exclusive`).
- **Replay switches** `:2597`, `:2928`: add arms (otherwise `core.unknown_operation`).
- **Classifier** `hasScalarCFG` (`:4298-4308`): add the two kinds to the operation list (deliberate copy; comment at cgen_program.go:2508-2514 says copies are not shared).
- **Fixed point** `validateScalarFixedPoint` (`:4344-4352`): place gate `constructor == "U64" || "Bool"` must admit the one parameter ByteView place; arms at ~4429-4440 (remainder arm at 4439 is the template) and straight-line facts at ~4535-4545:
```go
case core.OpRemainderU64:
	right := state[operation.RightID]
	if !v.check(source.initialized && right.initialized, "core.place_uninitialized", operation.ID) { return false }
	...
	value := scalarAbstractValue{initialized: true}
```
- **Release-vs-view law (flag (c)):** `OpMove` is the only op compared to loan liveness today:
  `case core.OpMove: blockedUntil, hasLoan := ownerBlockedUntil[operation.SourceID]; if !v.check(!hasLoan || blockedUntil < index, "core.move_while_borrowed", ...)` (corevalidate.go:2653-2656, 3000-3003). Add a sibling law for `OpRelease` of owner place P while a view derived from P is read later, with its own code, proven by seeded fault only (releases are not user-spellable).
- **Lifecycle replay variant** (`:1930-1975`): current pin is `len(foreignModes) != 3 || foreignModes[0] != "acquire" || foreignModes[1] != "borrow" || foreignModes[2] != "consume"`. The fixture variant has the same three modes plus an `OpCall` between borrow and consume; copy this replay as a new closed shape.
- `peerCallable` (`:4029`): confirm it still returns true for a ByteView callee (no scalar gate there today).

---

### `originvalidate/originvalidate.go` (service, transform)

**Analog:** `OpRemainderU64` handling. This file contains the flag (a) blocker.

**Classifier** (`originvalidate.go:293-300`): add `OpBytesLen`, `OpBytesAt` to `originHasScalarCFG`'s long `||` chain.

**Place gate to teach** (`originvalidate.go:314-317`):
```go
for _, place := range linear.Places {
	if types[place.TypeID] != "U64" && types[place.TypeID] != "Bool" {
		return false, ""
	}
```
Admit exactly the single parameter place of type ByteView, else `adler32` is un-Callable and `check` emits `core.callee_not_callable` at every call (`PublishProblemsFor`, originvalidate.go:1325-1334, `Callable := len(PublishProblemsFor(...)) == 0` at :2131).

**Transfer arm** (`originvalidate.go:456-468`, `originTransferScalarOperation`):
```go
case core.OpRemainderU64:
	right := state[operation.RightID]
	...
	value := originScalarValue{initialized: true}
	if source.known && right.known { value.known, value.u64 = true, source.u64%right.u64 }
	state[operation.TargetID] = value
```
Add `OpBytesLen` / `OpBytesAt` arms producing unknown initialized U64 (require initialized source/index).

**Independent structural problem** (`phase27ArithmeticProblem`, `originvalidate.go:1403-1425`):
```go
if places[operation.SourceID] != "U64" || places[operation.RightID] != "U64" || ... types[operation.TypeID] != want || ... {
	return &Problem{Code: "core.scalar_operation_invalid", Detail: fmt.Sprintf("%s: invalid %s operation %q", function.ID, operation.Kind, operation.ID)}
}
```
Write a sibling for the two ops so this peer refuses a forged shape on its own. `hasRemainderOperation` (:1388) gates the fixed-point branch in `PublishProblemsFor`; the `adler32` body contains `%` so it already takes that branch.

`BuildInterface` hard-codes `Mode: "owned"` (originvalidate.go:2099); no Mode plumbing needed (RESEARCH). `interfaceParameterAbilities` reads `function.ID+":type:0"` (:2279).

---

### `pathoracle/pathoracle.go` (service, event-driven)

**Analog:** `ValidateLocalOwnerPaths` (`pathoracle.go:59-135`). Closed, name-pinned, positional:
```go
case "acquire":
	if operation.Kind != core.OpForeignCall || operation.SourceID != function.Parameter.ID || places[operation.SourceID] != "PathToken" || places[operation.TargetID] != "FileByteOwner" || ... contract.Symbol != "schway_file_byte_acquire" || ... {
		return fmt.Errorf("pathoracle.local_owner_acquire: operation %q is not a valid owner seed", operation.ID)
	}
	owners[operation.ID] = owner{place: operation.TargetID, contract: contract}
```
Write a new fixture-owner replay (acquire sourced from the U64 entry parameter; borrow yields `ByteView` into a place; an `OpCall` consumes the view; consume requires `borrowed && !released`). Add `pathoracle.fixture_release_before_view_use` (name is discretion) refusing a release ordered before the `OpCall` that reads the view; derive independently of corevalidate/check. Scalar loops are excluded by design (cyclic): confirm `RecomputeEndpoints` ignores the new kinds on acyclic bodies.

---

### `interp/interp.go` (service, transform)

**Analog (op arm + OOB defect):** the `OpRemainderU64` arm at `interp.go:1406-1420`:
```go
case core.OpRemainderU64:
	right, initialized := top.values[operation.RightID]
	if !initialized || !sourceValue.isU64 || !right.isU64 {
		return Execution{}, fmt.Errorf("operation %q has invalid remainder operands", operation.ID)
	}
	if right.u64 == 0 {
		const reason = "U64 remainder by zero"
		events = append(events, Event{
			Schema: top.eventSchema(), ID: operation.ID + ":event:defected", Kind: "function.defected",
			FunctionID: top.function.ID, Invocation: top.invocation, SourcePlace: operation.SourceID,
			TypeID: operation.TypeID, Output: reason,
		})
		return Execution{Schema: top.eventSchema(), Outcome: Outcome{Kind: execution.OutcomeDefect, Value: ""}, Events: events, LiveResources: liveResourceList(top.live, top.liveOrder)}, nil
	}
	top.values[operation.TargetID] = value{u64: sourceValue.u64 % right.u64, isU64: true}
	top.idx++
```
`OpBytesAt`: `Output: "byte index out of bounds"`; `OpBytesLen`: no event. Event-free on success (flag (e)). `OpScalarStore` arm (interp.go ~1422) is event-free; `OpCopy` (:1439) emits `value.copied` (do not use in the loop).

**value struct** (`interp.go:824-833`): add `bytes []byte` (+ tag/flag, e.g. `isBytes`). The view is read-only so sharing a slice header is fine; check any `value` comparisons with `==` (slices make the struct non-comparable).

**Owner binding / defect-with-live-owner (D-20).** Analog: `RunWithForeignOutcomes` (`interp.go:196-260`). It demands a `PathToken`-to-`U64` entry and rejects live resources at the terminal:
```go
if len(result.LiveResources) != 0 {
	return ModelResult{}, fmt.Errorf("model execution reached a terminal outcome with live local resources %v", result.LiveResources)
}
```
(`interp.go:253`). Add a new fixture-owner entry (U64 entry param; independent Go byte generator keyed by case id) and permit a defect terminal with live resources only when the final event is the byte-index defect (D-20); keep every other terminal rejecting. The existing acquire outcome parse is `strconv.ParseUint(outcome.Value, 10, 8)` (:311, one canonical byte); the new owner holds `[]byte` instead.

Stage A seam (D-14): a Go-injected `[]byte` for the interpreter used by the gate harness only.

---

### `cgen/cgen_program.go` (service, transform)

**Analog 1: evidence-size bound** (`cgen_program.go:228-239`). `OpAddChecked`/`OpRemainderU64` reserve a terminal `function.defected` event; comparisons/stores/consts `continue`:
```go
case core.OpRemainderU64:
	event.ID, event.Kind, event.Output = operation.ID+":event:defected", "function.defected", "U64 remainder by zero"
	event.TargetPlace = ""
case core.OpLessU64, core.OpEqualU64, core.OpScalarStore, core.OpBranch, ...:
	continue
```
Add `OpBytesLen` to the `continue` list; `OpBytesAt` reserves `"byte index out of bounds"` like remainder. A missing arm is a hard error.

**Analog 2: classifier** `hasScalarOperations` (`cgen_program.go:2515-2527`): add the two kinds to the `||` chain; keep `isStraightLineScalarBody` U64-only rules unchanged. Comment :2508-2514 explains the copy-not-share policy.

**Analog 3: emission, both shells** (`cgen_program.go:2703-2713`, the remainder arm is the template; application shell vs conformance shell):
```go
case core.OpRemainderU64:
	fmt.Fprintf(out, "  if (%s == UINT64_C(0)) {\n", locals[op.RightID])
	if shell == programApplicationShell {
		out.WriteString("    fputs(\"schway: U64 remainder by zero\\n\", stderr);\n    exit(65);\n")
	} else {
		defect := op
		defect.Reason = "U64 remainder by zero"
		emitProgramDefectTerminal(out, function, defect)
		out.WriteString("    exit(0);\n")
	}
	out.WriteString("  }\n")
	fmt.Fprintf(out, "  %s = (%s %% %s);\n", ...)
```
`OpBytesAt`: `if (idx >= view.length) { fputs("schway: byte index out of bounds\n", stderr); exit(65); }` then `dst = (uint64_t)view.data[idx];` with `const unsigned char *data`. Never `assert`, never elided. `OpBytesLen`: `dst = view.length;`. The local declaration loop in `emitProgramScalarFunction` types every non-parameter place `uint64_t`/`int`; the ByteView parameter needs `SCHWAY_BYTE_VIEW` (typedef gated on use, near `SCHWAY_BUFFER` ~cgen_program.go:905; names must stay in `SCHWAY_*`, `TestGeneratedIdentifierNamespacesStayConfined`). `exit(65)` is inside the scanned literal set `{0, 65, 69}` (`TestEmittedCExitLiteralsMatchTable`).

**Analog 4: fixture-main emitter** `emitProgramLocalFileByteFunction` (`cgen_program.go:1773-1805`):
```go
fmt.Fprintf(out, "  schway_file_byte_acquire_result %s = %s(%s);\n", acquireResultName, acquire.Foreign.Symbol, parameterName)
fmt.Fprintf(out, "  if (%s.status != 0) exit(65);\n", acquireResultName)
fmt.Fprintf(out, "  schway_file_byte_owner %s = %s.owner;\n", ownerName, acquireResultName)
fmt.Fprintf(out, "  if (%s.data == NULL || %s.length != UINT64_C(1)) { %s(%s); exit(65); }\n", ...)
fmt.Fprintf(out, "  %s(%s);\n", release.Foreign.Symbol, ownerName)
```
Emit a sibling that acquires on the `uint64_t case` parameter, builds `SCHWAY_BYTE_VIEW{data, length}` from the owner, calls the adler32 function, releases, records `function.returned`. Per D-20 the OOB path inside the callee `exit(65)`s without releasing. The TU includes the adapter via `#include "local/adapter.h"` only for local-owner programs (cgen_program.go ~890); extend that gate. Names via `newCNames(...).allocate(...)` as in the analog.

Application-flag baseline: `applicationFlags = {"-std=c17","-Wall","-Wextra","-Werror","-pedantic","-O0"}` (`native_app.go:115`).

---

### `executionpeer/executionpeer.go` and `session/session.go` (evidence projection)

**executionpeer** (`executionpeer.go:367-368`, `:497`):
```go
checkedOverflow := op.value.Kind == core.OpAddChecked && event.Kind == "function.defected"
checkedRemainderZero := op.value.Kind == core.OpRemainderU64 && event.Kind == "function.defected"
if (i.loopOperations[op.value.ID] || event.Occurrence > 0) && !i.scalarLoopCopies[...] && !checkedOverflow && !checkedRemainderZero { refuse "loop_event" }
```
Add `checkedBytesAt` (fixed reason `byte index out of bounds`, source-place attribution, final event, empty defect outcome) next to the others, and add `core.OpBytesAt` to the classifier at :497 (`... || op.value.Kind == core.OpRemainderU64`). Without this the peer rejects the OOB run.

**session.go `ProjectExecutionSchema2`** (`session.go:527-533`):
```go
case core.OpAddChecked:
	defectReasons[operation.ID+":event:defected"] = "U64 addition overflow"
case core.OpRemainderU64:
	defectReasons[operation.ID+":event:defected"] = "U64 remainder by zero"
```
Add `case core.OpBytesAt: ... = "byte index out of bounds"`.

**session.go admission** (`session.go:1401-1425`, copy for the fixture entry):
```go
localOwnerEntry := entry.Parameter.Type == "PathToken" && hasLocalForeignOperations(program)
u64Application := entry.Parameter.Type == "U64" && (entry.ReturnType == "U64" || entry.ReturnType == "Unit")
...
if localOwnerEntry {
	if err := pathoracle.ValidateLocalOwnerPaths(program); err != nil {...}
	if len(manifestPath) != 1 { return ..., errors.New("local-owner application requires one explicit binding manifest") }
	bindings, err := native.ResolveBindings(manifestPath[0])
	if err := validateLocalOperationBindings(program, bindings.Manifest); err != nil {...}
}
```
Fixture entries are `U64` entries with foreign operations, which bypass this block today (`u64Application` true, `localOwnerEntry` false). Add `fixtureOwnerEntry := u64Application && hasLocalForeignOperations(program)`, run the new pathoracle replay, and require the manifest. `validateLocalOperationBindings` (:1446-1470) is symbol-generic: no change. Interpreter side: `RunInterpreterApplication` (:1153-1180) accepts only linear `U64` entries and routes to `interp.RunApplication`; add the fixture-owner route (session API is the interpreter's runnable gain, D-16). Any edit that shifts an `EmitNative`/`Emit` call site needs `TestPhase16PublicEmitterConsumerInventory` registry update.

---

### `examples/phase29/` fixtures, adapter, manifest (fixture / adapter / config)

**Source analog:** `examples/phase23/file_byte.schway` (read in full):
```
foreign C {
  fn schway_file_byte_acquire(path: PathToken) -> FileByteOwner {
    unwind: forbidden
    nonlocal_exit: forbidden
    mode: "acquire"
    abi: schway_file_byte_acquire_fn
    allocator: "libc_malloc"
    release: "schway_file_byte_release"
    fails: AcquireError
  }
  fn schway_file_byte_use(owner: FileByteOwner) -> U64 { ... mode: "borrow" abi: schway_file_byte_use_fn fails: UseError }
  fn schway_file_byte_release(owner: FileByteOwner) -> Unit { ... mode: "consume" abi: schway_file_byte_release_fn allocator: "libc_malloc" }
}
data AcquireError = | AcquireFailed
fn main(path: PathToken) -> U64 {
  let owner = try schway_file_byte_acquire(path)
  let value = try schway_file_byte_use(owner)
  value
}
```
Phase 29 version (D-15, D-18): acquire takes `case: U64` -> `FixtureBytes`; borrow-mode function returns `ByteView`; `fn main(case: U64) -> U64 { let owner = try fixture_acquire(case); let view = try fixture_view(owner); let r = adler32(view); r }` (shape is discretion; release is compiler-inserted). `adler32(bytes: ByteView) -> U64` is the scalar library function (shape in RESEARCH "Library shape"). Do NOT edit `examples/phase23`; D-09 forbids reuse of `PathToken`/`FileByteOwner` names.

**Scalar loop shape analog:** `examples/sum_to_n.schway` (`var`/`while`, bare-identifier result).

**Adapter analog:** `examples/phase23/adapter.h` (struct owner `{unsigned char *data; uint64_t length;}`, `typedef ..._fn` signatures matching the `abi:` names, `_Static_assert` size/offset pins, `acquire_result{status, owner}`) and `adapter.c` (status enum, `SCHWAY_..._MALLOC`/`FREE` macros). Phase 29 adapter: independent C generator keyed by case id (empty, one `0x00`, `a`, `Wikipedia`, 4096x`0x00`, 5552x`0xFF`, OOB control, plus cap workloads 4096/16384/65536 x `0xFF`); reject unknown case ids with a status (V5); allocate `max(1,n)`; header typedefs `fixture_bytes_*_fn`.

**Manifest analog** (`examples/phase23/file_byte.bindings.json`, copy exactly with new names):
```json
{
  "schema": "schway.local-c/1",
  "sources": ["adapter.c"],
  "headers": ["adapter.h"],
  "include_dirs": ["."],
  "symbols": [
    {"name": "schway_file_byte_acquire", "header": "adapter.h", "function_type": "schway_file_byte_acquire_fn"},
    ...
  ],
  "runtime_dependencies": ["platform-c-runtime"]
}
```
Symbols must exactly cover the checked foreign operations (`validateLocalOperationBindings`).

---

### Tests (per-package `*_phase29_test.go`)

**End-to-end OOB/native-run analog:** `session/session_phase26_overflow_test.go:19-80` (`TestPhase26OverflowProcessOutcome`):
```go
interpreted, err := interp.Run(checked.Program, "main", "1")
if err != nil || interpreted.Outcome.Kind != execution.OutcomeDefect || interpreted.Outcome.Value != "" { ... }
event := interpreted.Events[len(interpreted.Events)-1]
if event.Kind != "function.defected" || event.Output != "U64 addition overflow" || event.SourcePlace == "" { ... }
interpretedEvidence, err := session.ProjectExecutionSchema2(checked.Program, interpreted)
executionpeer.Validate(checked.Program, interpretedEvidence)
cSource, err := cgen.EmitNative(checked.Program)
runner.Expect = native.ExpectDefect
nativeResult, err := runner.Run(context.Background(), cSource, "-O0", []string{"1"})
_, diagnostics, err := session.BuildApplicationFile(ctx, sourcePath, artifact, native.DefaultRunner())
outcome, err := native.RunApplication(ctx, artifact, "1", &stdout, &stderr)
// want: ExitCode == 65 && stdout.Len() == 0 && stderr.String() == "schway: U64 addition overflow\n"
```
Phase 29 interpreter assertion is the defect outcome + reason + empty output (D-19, Pitfall 9); native asserts exit 65 + exact stderr `schway: byte index out of bounds\n`. Mutation-kill analog: `TestPhase26WrongResultMutationKilled` (session_phase26_overflow_test.go:91): mutate emitted C text (signed-char: `unsigned char` -> `char`), must turn the `0x80`/`0xFF` fixture red.

**Fixture-owner end-to-end analogs:** `native/phase23_observer_test.go:162-260` (`newPhase23ObserverFixture`: copies `adapter.h`/`bindings.json`/`adapter.c` to a temp root, build via runner, run artifact) and `session/session_phase24_model_test.go:40` (`interp.RunWithForeignOutcomes(program, "main", ..., map[...]ForeignOutcome{...})`). Admission test pattern: `session/session_phase271_admission_test.go:213-238` controls table (add the Phase 29 fixture with `manifest: {"examples","phase29","<name>.bindings.json"}`).

**Seeded fault per peer / test-only seam analog:** `check/scalar.go:30-32` (`scalarAnalysisBudgetForTest`, unexported, nil in production; used by `check_test.go:5780`) and `check.go:35-55` (`SetPhase17ReturnLookupFaultForTest` returns a `restore` func). Use for the D-21 application-shell `-O3`/`-flto` seam in `native` (unexported override of `applicationFlags`, restored in test; do not change receipts or the production flag slice). `Runner.LTO` (`native.go:81-87`) already gates `-flto` on both compile and link.

**Registers to update** (each new committed `.schway` under `examples/` or `testdata/`):
- `session_phase28_manifest_test.go:232` `phase28NewlyAddedEntries` (entry per path: `Check`, `Build`, `InterpreterRun`, `NativeRun`, `EntryShape`, `Entry`; note `schway run --engine` exits 3 for U64 entries, copy the `examples/phase28/unused_function.schway` class values)
- `session_phase272_identity_test.go:34` `phase272NewlyAcceptedPaths`
- `session_phase273_identity_test.go` `phase273NewlyAcceptedPaths`
- `.planning/LANGUAGE-MATURITY.md` `**N .schway programs, N lines total**` sentence (`TestLanguageMaturityCountsAreCurrent`); do not add `len(Functions) != 1` guards
- Corpus byte-identity: existing core/C hashes must not move (new type-fact order only for ByteView functions).

**Avoid:** committing any `.schway` under `.planning/spikes/008-...` (Pitfall 6; generate into `t.TempDir()`).

---

### Spike 008: `.planning/spikes/008-byteview-cap/README.md` + MANIFEST row (doc, batch)

**Analog:** `.planning/spikes/007-loan-across-branch/README.md` frontmatter and sections:
```
---
spike: 007
idea: ownership-kernel
name: loan-across-branch
type: standard
validates: "..."
verdict: VALIDATED
related: [002-cfg-edge-last-use, 006-interprocedural-liveness-cost-scaling]
tags: [...]
---
# Spike 007: ...
## What This Validates / ## Research / ## How to Run / ## What to Expect
```
How-to-run analog (spike 007): `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/check -run '^TestEdgeSpecificLiveOut$' -count=1`. Use `SCHWAY_SPIKE008=1 go test ./internal/compiler/native -run 'Spike008' -count=1 -v` (and `-race`). Add one row to `.planning/spikes/MANIFEST.md`.

**CONVENTIONS to follow** (`.planning/spikes/CONVENTIONS.md`): stdlib only; "Include a deliberate fault-injection path proving that the harness detects semantic drift" (use the per-iteration `let` mutant that must trip capacity); "External tool invocations record exact toolchain, target, flags, content digest, compile time, and artifact size; temporary outputs stay outside the source tree"; "Separate semantic events from measurement fields such as elapsed wall time"; native lanes compare against an independently encoded oracle at `-O0` and optimized, sanitizers as an orthogonal lane.

Measurement primitives: `native.Runner.Run` returns `CompileTime`/`RunTime` separately; `Runner{LTO: true}`; deadlines `defaultSubprocessTimeout` 30 s (`native.go:30`), `RunCLITimeout = 2 * time.Minute` (`testsupport.go:37`); assert the checksum on every run (65,536x`0xFF` -> `2006388466`; 16,384 -> `2967061426`; 4,096 -> `2170679522`). Stage A (29-00) uses a harness-delivered `[]byte` and a gate-only native driver labeled "proxy, gate-only"; Stage B (last plan) goes through the real fixture owner with hosted run ids recorded. No wall time in golden files.

## Shared Patterns

### One op, one copy per consumer (no shared helper)
**Source:** comment at `cgen_program.go:2508-2514`: "the sets are deliberately copied, not shared (no shared package, D-27-07 copy-over-dep), so drift stays visible."
**Apply to:** check `scalar.go`, `corevalidate.hasScalarCFG`/`validateScalarFixedPoint`, `originvalidate.originHasScalarCFG`/`originTransferScalarOperation`, `cgen.hasScalarOperations`. Add a test per site and one seeded fault per peer.

### Checked-failure exit 65 with exact stderr (application shell) / defect terminal (conformance shell)
**Source:** `cgen_program.go:2690` (overflow) and `:2703-2713` (remainder), shown above.
**Apply to:** `OpBytesAt` in cgen. Interpreter analog is the `Event{Kind:"function.defected", Output: reason}` + `Outcome{Kind: execution.OutcomeDefect, Value: ""}` return at `interp.go:1410-1418`.

### No per-iteration events
**Source:** `OpScalarStore` event-free in interp (~:1422), cgen (plain assignment), evidence bound (`continue`, cgen_program.go:239); `OpCopy` evented (interp.go:1439, cgen_program.go:2685).
**Apply to:** library body (`var`/assignment only), new ops (no success events). Structural test: zero `OpCopy` in loop blocks; `let`-in-loop negative control must be seen to fail.

### Exit-code literals
**Source:** `internal/compiler/exitcode` single table; `cgen/cgen_exit_literals_test.go` (`TestEmittedCExitLiteralsMatchTable`) scans emitted C for exactly `{0, 65, 69}`.
**Apply to:** all new emitted C; no new exit status.

### Diagnostics tested by first code
**Source:** `phase28Classify` pins `result.Diagnostics[0].Code`; cascades follow the first parser error.
**Apply to:** every neighbor-refusal test (borrow/take in loop -> `syntax.expected_scalar_operand`; owner/loan param or `-> Buffer`/`-> ByteView` -> `type.scalar_signature`; copy/store/return view -> `type.scalar_binding`/`type.scalar_assignment`/`type.scalar_result`; loan param forged core -> `check.cfg_back_edge` category `loan`, existing `TestPhase26BackEdgeCauseCategories`; release-in-loop -> seeded fault only).

### Pinned oracle literals
**Source:** RESEARCH "Pinned vectors": empty `1`, `0x00` `65537`, `a` `6422626`, `Wikipedia` `300286872`, 4096x`0x00` `268435457`, 5552x`0xFF` `4052720524`, 65536x`0xFF` `2006388466`. Record producing command and version (Python 3.14.4 `zlib.adler32`, Go 1.24.0 `hash/adler32`) beside each; table test asserts the C and Go byte generators produce identical bytes.

### Meta-tests to run focused before hosted dispatch
`TestPhase28EntryExitClassManifest`, `TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestLanguageMaturityCountsAreCurrent`, `TestPhase16PublicEmitterConsumerInventory`, validation-lifecycle and groundedness tests (`29-VALIDATION.md` must not stay `status: draft`). Never run the full suite locally.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `SCHWAY_BYTE_VIEW` C struct passed by value as a function argument | cgen type | request-response | No existing generated C passes a two-member struct view as a function parameter; closest is `SCHWAY_BUFFER` (cgen_program.go ~905) and `schway_file_byte_owner` in `examples/phase23/adapter.h` (`{unsigned char *data; uint64_t length;}`). Use that struct layout and its `_Static_assert` pin style, with `const unsigned char *data`. |
| Env-gated spike harness `SCHWAY_SPIKE008=1` | test | batch | No existing env-gated measurement test; spike 007 reuses a production test, spikes 005/006 keep Go code beside the README with their own `go.mod`. Follow CONVENTIONS; skip with a named skip when unset (as the existing sanitizer lanes do). |
| Release-while-loan-live law (any layer) | validator law | transform | RESEARCH flag (c): no layer compares `OpRelease` to loan liveness (`OpMove` only, corevalidate.go:2653). Author new independent laws using the `core.move_while_borrowed` structure as a template. |
| Interpreter owner holding real bytes | interp | transform | `RunWithForeignOutcomes` is a one-byte modeled outcome (`interp.go:311`); no analog holds a byte slice. Closest structure: `base.modeledOutcomes` plumbing and `liveResourceList` handling. |

## Metadata

**Analog search scope:** `internal/compiler/{syntax,ast,core,check,ability,corevalidate,originvalidate,pathoracle,interp,cgen,executionpeer,session,native}`, `examples/{phase23,phase26,phase27,phase272,phase28,sum_to_n.schway}`, `.planning/spikes/`.
**Files scanned:** ~45 (targeted line-range reads; large files read by range only).
**Pattern extraction date:** 2026-10-11
