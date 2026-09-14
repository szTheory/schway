# Phase 12: `Result` Payloads - Pattern Map

**Mapped:** 2026-09-12
**Files analyzed:** 13 (all modified, not created — this phase is pure extension of existing packages; no new packages, per RESEARCH.md's "Recommended Project Structure")
**Analogs found:** 13 / 13 (every file's analog is itself — this is a same-file dispatch-site-addition pattern, not a new-file-from-different-file pattern)

**Special note on this phase's shape:** Unlike a typical phase that creates new
files, Phase 12 is defined by D-12-05 as *"~16-20 independent edits across the
six literal dispatch sites"* inside already-existing files. There are no new
`.go` files in `internal/compiler/*` (only new `testdata/phase12/*.lang` /
`*.golden.c` fixtures and one or two new `_test.go` files). Consequently the
"analog" for each touched production file is **the existing sibling
`OperationKind` case arm in that same file** (how `OpForeignCall`/`OpCall`
were added in Phases 4/7), not a different file. This PATTERNS.md maps each
touched file to the precedent arm/commit-shape it must copy.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog (same-file precedent) | Match Quality |
|---|---|---|---|---|
| `internal/compiler/core/core.go` (`OperationKind` consts, `AllOperationKinds()`, `DataType`, `LinearOperation`) | model (core IR) | transform | `OpCall`'s own addition (D-07-29): new const + `AllOperationKinds()` entry + additive omitempty fields (`CalleeID`) | exact |
| `internal/compiler/core/core_test.go` (`exhaustiveDispatchFixtures`, `TestAllOperationKindsHandledAtEverySite`) | test | batch | the existing fixture-list append pattern used when `OpCall` was added (`testdata/phase07/call_basic.lang` entries) | exact |
| `internal/compiler/ast/ast.go` (`Alternative` payload-type field, `MatchArm` binder field) | model (AST) | transform | `MatchArm.BlockID`'s own additive-field precedent (Phase 3 extension, doc-commented) | exact |
| `internal/compiler/syntax/parser.go` (`dataDecl`, `matchExpr`) | parser | request-response (parse) | the existing `TokenPipe`-loop alternative parse (`dataDecl` line 282) and the pattern-then-arrow parse (`matchExpr` line 523) | exact |
| `internal/compiler/check/check.go` (`checkBranch`, arm-entry `OpCopy` emission, `standardForeignLayout`) | service (semantic pass) | transform | `checkBranch`'s own existing arm-entry `aliasOp := core.LinearOperation{Kind: core.OpCopy, ...}` (line 1653) — the exact site the payload move-on-bind op is added beside | exact |
| `internal/compiler/corevalidate/corevalidate.go` (independent re-derivation switch, `peerDeriveOriginFacts`) | service (independent validator) | transform | the existing `switch operation.Kind { case core.OpCopy: ... case core.OpMove: ... }` arms at line 1667/1709 and 1924/1966 | exact |
| `internal/compiler/interp/interp.go` (`values` map widening, operation-kind switch) | service (semantic oracle / interpreter) | event-driven | the existing `switch operation.Kind { case core.OpCopy: ...; case core.OpMove: ... }` at line 628/632, and `newFlatFrame`/`newBlockFrame`/`newArmFrame` (lines 357/374/427) | exact |
| `internal/compiler/cgen/cgen.go` (`emitMatch`, the four `OpCopy/OpMove/OpBorrowShared/OpBorrowExclusive` case arms) | service (native codegen) | transform | the existing shared case arm `case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:` (lines 273, 582, 770, 1773) | exact |
| `internal/compiler/cgen/cgen_program.go` (`emitProgram`, D-11-52 refusal lift) | service (native codegen, multi-function) | transform | `emitProgram` (`cgen_program.go:128`) and the refusal string at `:152` — inherited D-11-02/D-11-52 work, not a payload-specific analog | role-match |
| `internal/compiler/pathoracle/pathoracle.go` (terminator/path walk) | service (independent validator) | transform | the existing `operation.Kind == core.OpBorrowShared \|\| operation.Kind == core.OpBorrowExclusive` membership test (line 324) | role-match (body-blind by design, per D-04-29) |
| `internal/compiler/originvalidate/originvalidate.go` (`walkReturnOrigin`'s `switch operation.Kind`) | service (independent validator) | transform | the existing `switch operation.Kind { case core.OpBorrowExclusive: ... case core.OpCall: ... }` at line 243-270 | exact |
| `internal/compiler/session/session.go` (new `PayloadLayoutMutationRunner` or equivalent, D-12-37) | service (mutation-control runner) | batch | `LayoutMutationRunner` / `control:foreign.layout_mismatch` (lines 565-587) | exact |
| `internal/compiler/session/session_phase5_compare.go` | service (differential comparator) | transform | **NO CHANGE** — `comparePhase5Pair`'s existing `Outcome.Value` comparison (line 89-96) already covers payload values (D-12-39) | n/a — explicitly do-not-modify |

## Pattern Assignments

### `internal/compiler/core/core.go` (model, transform)

**Analog:** `OpCall`'s own addition (D-07-29), same file.

**Existing `OperationKind` const block** (`core.go:530-564`):
```go
const (
	OpCopy            OperationKind = "copy"
	OpMove            OperationKind = "move"
	OpBorrowShared    OperationKind = "borrow_shared"
	OpBorrowExclusive OperationKind = "borrow_exclusive"
	OpReturn          OperationKind = "return"
	// OpForeignCall is Phase 4's only call surface (D-04-01): ...
	OpForeignCall OperationKind = "foreign_call"
	// OpFail is a terminator (see TerminatorKinds) ...
	OpFail OperationKind = "fail"
	// OpRelease is Phase 4 plan 02's resource-lifecycle operation (D-04-07): ...
	OpRelease OperationKind = "release"
	// OpDefect is a terminator (see TerminatorKinds) ...
	OpDefect OperationKind = "defect"
	// OpCall is Phase 07's Lang-to-Lang call surface (D-07-29): unlike
	// OpForeignCall's two successor edges, an OpCall carries a single
	// TargetID (exactly like OpCopy) plus a CalleeID naming the resolved
	// callee.
	OpCall OperationKind = "call"
)
```
Copy this exact doc-comment shape for the two new kinds — a comment stating
which existing kind the new one resembles structurally (TargetID shape) and
which precedent it deviates from (D-12-05's rejection of riding `OpCopy`).

**`AllOperationKinds()`** (`core.go:704`):
```go
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect, OpCall}
}
```
Append the two new kinds (e.g. `OpConstructPayload`, `OpDestructurePayload`) here. This single line is what forces `TestAllOperationKindsHandledAtEverySite` to demand real fixture coverage — this is the anti-vacuity mechanism D-12-05 argues for.

**`DataType`** (`core.go:22-27`) — additive sibling field, following D-12-09's constructor-invariant requirement:
```go
type DataType struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Alternatives []string        `json:"alternatives"`
	Span         diagnostic.Span `json:"span"`
}
```
Add a `omitempty` payload-detail list here (e.g. `AlternativeDetails []AlternativeDetail `json:"alternative_details,omitempty"``), keyed by name, matching `Alternatives`. Per D-12-09, this struct must gain **exactly one** exported constructor function enforcing name-set equality between `Alternatives` and the new list — do not let call sites build `DataType{}` literals directly once payload alternatives exist (existing raw-literal call site to update: `check.go:80`, `dataType := core.DataType{ID: ..., Alternatives: alternatives, ...}`).

**`LinearOperation`** additive-omitempty precedent to copy verbatim in shape (`core.go:714-762`, the `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` and `CalleeID` fields):
```go
// OkEdgeID, ErrEdgeID, and ErrTargetID are Phase 4 additive omitempty
// facts populated only on an OpForeignCall operation (D-04-04): ...
// Every pre-Phase-4 operation, and every operation kind other than
// OpForeignCall, leaves all three empty.
OkEdgeID    string `json:"ok_edge_id,omitempty"`
ErrEdgeID   string `json:"err_edge_id,omitempty"`
ErrTargetID string `json:"err_target_id,omitempty"`
...
// CalleeID is Phase 07's additive omitempty fact (D-07-29): populated
// only on an OpCall operation, it holds the resolved callee's function
// ID -- never its name. Every pre-Phase-07 operation, and every
// operation kind other than OpCall, leaves this empty, so no
// pre-Phase-07 core artifact moves a byte.
CalleeID string `json:"callee_id,omitempty"`
```
Add `PayloadType string `json:"payload_type,omitempty"`` (D-12-07, bare string like `ReturnType`/`Parameter.Type`) and `PayloadTargetID string `json:"payload_target_id,omitempty"`` (D-12-10, following the `ErrTargetID` naming precedent exactly), each with the identical "populated only on Op<New>, every other kind leaves this empty" doc-comment shape.

---

### `internal/compiler/core/core_test.go` (test, batch)

**Analog:** the existing fixture-list append + the control's own doc comment (`core_test.go:267-300`).

```go
// TestAllOperationKindsHandledAtEverySite is the control:kind.exhaustive_dispatch
// table (D-04-22): for every existing OperationKind, drive real programs
// containing that kind through check (session.Check), corevalidate,
// interp, cgen, pathoracle, and originvalidate, and assert none of them
// reject or crash. Every declared kind must be exercised by at least one
// fixture, so a kind that no corpus program ever produces cannot silently
// pass this control by omission.
//
// pathoracle and originvalidate do not switch exhaustively on
// core.OperationKind today (they walk only terminators); "handled" for these
// two sites means the walk completes without error on a program containing
// the kind ...
var exhaustiveDispatchFixtures = []string{
	"testdata/phase1/toggle.lang",
	...
	"testdata/phase07/call_basic.lang",
	"testdata/phase07/call_from_both_match_arms.lang",
}
```
Append new `testdata/phase12/*.lang` fixture path(s) exercising `OpConstructPayload`/`OpDestructurePayload`. Per Pitfall 1 (RESEARCH.md), do **not** author only a trivial happy-path construct-then-match fixture — interact with an existing borrow/call shape, per D-12-02's "darkest corner, not average" discipline.

---

### `internal/compiler/ast/ast.go` (model, transform)

**Analog:** `MatchArm.BlockID`'s own additive-field precedent, and the doc comment that falsifies the naive "extends `Value`" claim (D-12-13) — read this exactly, it is load-bearing for why the parser work is real:

```go
// Source: internal/compiler/ast/ast.go:53-62, :111-124 (verified)
type DataDecl struct {
	Name         string
	Alternatives []Alternative
	Span         diagnostic.Span
}

type Alternative struct {
	Name string
	Span diagnostic.Span
}

type MatchArm struct {
	Pattern string
	Value   string
	// Body is the Phase 3 extension: an arm's value position may hold a full
	// linear body instead of a bare alternative name. Exactly one of Value
	// and Body is populated ...
	Body *LinearBody
	Span diagnostic.Span
}
func (a MatchArm) HasClosedVariant() bool { return (a.Value != "") != (a.Body != nil) }
```
`Alternative` gets an additive `PayloadType string` field (empty for nullary alternatives — mirrors `DataType.PayloadType`'s bare-string convention). `MatchArm` gets an additive `Binder string` field (empty when the arm binds nothing), following the exact "additive field, empty means old behavior" shape `Body` itself established relative to `Value`.

---

### `internal/compiler/syntax/parser.go` (parser, request-response)

**Analog:** the existing `dataDecl` alternative-parse loop and `matchExpr`'s pattern-then-arrow parse (both same file, adjacent sites).

**Alternative-declaration parse to extend** (`parser.go:282-296`):
```go
func (p *parser) dataDecl() ast.DataDecl {
	start := p.expect(TokenData, "syntax.expected_data")
	name := p.identifier("syntax.expected_type_name")
	p.expect(TokenEqual, "syntax.expected_equal")
	decl := ast.DataDecl{Name: name.Text, Span: spanFrom(start, name)}
	for p.accept(TokenPipe) {
		alternative := p.identifier("syntax.expected_alternative")
		decl.Alternatives = append(decl.Alternatives, ast.Alternative{Name: alternative.Text, Span: alternative.Span})
		decl.Span.End = alternative.Span.End
	}
	if len(decl.Alternatives) == 0 {
		p.problem("syntax.expected_alternative", p.peek(), "data declaration needs at least one alternative")
	}
	return decl
}
```
Insert an optional `( identifier )` production immediately after `alternative := p.identifier(...)` for `| Ok(Buffer)` (D-12-11).

**Match-arm pattern parse to extend** (`parser.go:523-556`, confirms D-12-13's finding — this is genuinely new work, not an extension of an existing binder slot):
```go
func (p *parser) matchExpr() ast.MatchExpr {
	start := p.expect(TokenMatch, "syntax.expected_match")
	scrutinee := p.identifier("syntax.expected_scrutinee")
	p.expect(TokenLBrace, "syntax.expected_lbrace")
	expression := ast.MatchExpr{Scrutinee: scrutinee.Text, Span: spanFrom(start, scrutinee)}
	for !p.atAny(TokenRBrace, TokenData, TokenFn, TokenEOF) {
		startPosition := p.position
		pattern := p.identifier("syntax.expected_pattern")
		if !p.accept(TokenFatArrow) {
			p.problem("syntax.expected_fat_arrow", p.peek(), "expected `=>`")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF)
			continue
		}
		if len(expression.Arms) >= maxArmsPerMatch {
			p.problem("syntax.arm_limit", pattern, "match exceeds the declared arm limit")
			p.recoverUntil(TokenIdentifier, TokenRBrace, TokenData, TokenFn, TokenEOF)
			p.assertProgressOrBoundary(startPosition, TokenRBrace, TokenData, TokenFn, TokenEOF)
			continue
		}
		// ... bare-name-or-body-arm dispatch (existing TokenLBrace / value.identifier branches)
	}
```
Insert the optional `( identifier )` binder production immediately after `pattern := p.identifier("syntax.expected_pattern")`, before the `p.accept(TokenFatArrow)` check, with new named refusals (`syntax.*` namespace, matching `syntax.expected_pattern`/`syntax.arm_limit`'s naming shape) for malformed forms.

---

### `internal/compiler/check/check.go` (service, transform)

**Analog:** `checkBranch`'s own existing arm-entry copy operation — this is literally the site the new move-on-bind op is added beside.

**Existing arm-entry `OpCopy` emission** (`check.go:1653-1660`):
```go
armBlockID := fmt.Sprintf("%s:block:arm:%d", functionID, index)
aliasOpID := fmt.Sprintf("%s:op:%d", functionID, nextIndex)
aliasPointID := fmt.Sprintf("%s:point:linear:%d", functionID, nextIndex)
aliasPlaceID := fmt.Sprintf("%s:place:%d", functionID, nextIndex+1)
aliasOp := core.LinearOperation{
	ID: aliasOpID, PointID: aliasPointID, Kind: core.OpCopy, SourceID: parameterID, TargetID: aliasPlaceID, TypeID: typeID,
}
linear.Places = append(linear.Places, core.Place{ID: aliasPlaceID, Name: function.Parameter.Name, TypeID: typeID})
linear.Operations = append(linear.Operations, aliasOp)
armOperationIDs := []string{aliasOpID}
armOps := []core.LinearOperation{aliasOp}
nextIndex++
```
Per D-12-14, when the arm's pattern carries a binder (`arm.Binder != ""`), emit a second op here — `Kind: core.OpDestructurePayload` (or the chosen name), `SourceID: aliasPlaceID` (the just-minted arm-scoped alias), `TargetID:` a freshly-minted place ID following the SAME `nextIndex`-based minting authority (D-12-10 — never a second minting authority) — mint a fresh `core.Place` for it exactly as `aliasPlaceID` is minted above, then append to `armOps`.

**Existing named-refusal error shape to copy for D-12-15's three new refusals** (`check.go:1631-1637` area, same function):
```go
if !contains(dataType.Alternatives, arm.Pattern) {
	diagnostics = append(diagnostics, diagnostic.Error("match.unreachable", arm.Span, "pattern is not an alternative of the scrutinee type"))
	continue
}
if arm.Body == nil {
	diagnostics = append(diagnostics, diagnostic.Error(
		"core.mixed_arm_forms", arm.Span,
		"a match with any arm body requires every arm to carry a body this phase",
	))
	continue
}
```
Use this exact `diagnostic.Error("<namespace>.<code>", span, "<message>")` + `continue` shape for arity-mismatch, binder-on-nullary, and missing-binder-for-payload refusals — namespace them `check.*` per RESEARCH.md's Open Question 1 recommendation (planner records the exact strings).

**Shared layout fact to extend, not reinvent** (`check.go:3118-3125`, D-12-25):
```go
func standardForeignLayout() *core.RecordLayout {
	return &core.RecordLayout{
		Size: 1, Alignment: 1, ForeignTypeName: "lang_foreign_resource_block",
		Fields: []core.LayoutField{
			{Name: "payload", Size: 1, Alignment: 1, Offset: 0, CType: "unsigned char"},
		},
	}
}
```
Add a sibling function deriving the payload struct's `RecordLayout` (tag field + one field per alternative, D-12-22) the same way — this becomes the single shared fact `corevalidate`/`originvalidate`/`pathoracle`/`interp`/`cgen` all read (D-12-25).

**Raw `DataType{}` literal to migrate to the new constructor (D-12-09)** (`check.go:80`):
```go
dataType := core.DataType{ID: semanticID(program.Module, "type", declaration.Name), Name: declaration.Name, Alternatives: alternatives, Span: declaration.Span}
```

---

### `internal/compiler/corevalidate/corevalidate.go` (service, transform)

**Analog:** the existing per-operation-kind switch's `OpCopy`/`OpMove` arms (two separate switches in this file — one for per-place ownership tracking, one for a different peer).

**First switch** (`corevalidate.go:1665-1713`):
```go
switch operation.Kind {
case core.OpCopy:
	if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
		return false
	}
	if !v.targetMatches(function, index, operation, places, produced) {
		return false
	}
	initialized[operation.TargetID] = true
	produced[operation.TargetID] = true
...
case core.OpMove:
	blockedUntil, hasLoan := ownerBlockedUntil[operation.SourceID]
	if !v.check(!hasLoan || blockedUntil < index, "core.move_while_borrowed", operation.ID) {
		return false
	}
	if !v.targetMatches(function, index, operation, places, produced) {
		return false
	}
	initialized[operation.SourceID] = false
	initialized[operation.TargetID] = true
	produced[operation.TargetID] = true
```
Add `case core.OpConstructPayload:` and `case core.OpDestructurePayload:` arms here with the analogous `initialized`/`produced` bookkeeping — destructure clears the source's payload-slot liveness and initializes the fresh target place (mirrors `OpMove`'s `initialized[operation.SourceID] = false`), enforcing D-12-29's "dropped exactly once" static accounting property independently of `check`'s worklist.

**Second switch, `peerDeriveOriginFacts`** (`corevalidate.go:2407-2432`) — **DO NOT** add a `case core.OpCall:` arm here (Pitfall 3 / D-12-28's explicit exclusion); this switch stays as-is except for whatever payload-kind arms are needed for the payload-origin re-derivation itself:
```go
func peerDeriveOriginFacts(function *core.Function) peerOriginFact {
	...
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared:
			...
		case core.OpBorrowExclusive:
			...
		case core.OpMove, core.OpCopy:
			if paramTrace[operation.SourceID] {
				paramTrace[operation.TargetID] = true
			}
			if mode := derived[operation.SourceID]; mode != "" && derived[operation.TargetID] == "" {
				derived[operation.TargetID] = mode
			}
		}
	}
```
If a payload destructure needs to propagate `derived`/`paramTrace` through the new op (so a bound payload's origin is still traceable), add `core.OpDestructurePayload` to the **existing** `case core.OpMove, core.OpCopy:` arm's kind list (propagate-through, same semantics) — this is filling a case list with an already-modeled semantics, not a new interprocedural rule, so it does not trip D-12-28's guard.

---

### `internal/compiler/interp/interp.go` (service, event-driven)

**Analog:** the existing `switch operation.Kind` at the interpreter's main dispatch loop (`interp.go:626-676`).

```go
switch operation.Kind {
case core.OpCopy:
	top.values[operation.TargetID] = value
	events = append(events, ownedEvent(top.function, operation, "value.copied"))
	top.idx++
case core.OpMove:
	delete(top.values, operation.SourceID)
	top.values[operation.TargetID] = value
	events = append(events, ownedEvent(top.function, operation, "value.transferred"))
	top.idx++
case core.OpBorrowShared:
	top.values[operation.TargetID] = value
	events = append(events, ownedEvent(top.function, operation, "value.borrowed"))
	top.idx++
```
Add `case core.OpConstructPayload:` (build a `value{tag: <alternative name>, payload: <sourced string>}`) and `case core.OpDestructurePayload:` (mirror `OpMove`'s `delete` + assign shape, extracting `.payload` into the fresh target place) here, each emitting its own named event (`ownedEvent(top.function, operation, "value.payload_constructed")` / `"value.payload_destructured"`, following the existing `"value.copied"`/`"value.transferred"` naming convention).

**Value-map widening sites** (`interp.go:357-364`, `374-397`, `427-431` — representative of ~17 total per D-12-17):
```go
func newFlatFrame(function core.Function, values map[string]string) frame {
	...
	return frame{
		function: function, values: values, live: map[string]bool{}, operations: ops, ids: ids,
		placeTypes: placeTypeIndex(function), types: typeFactIndex(function),
	}
}
func newBlockFrame(function core.Function, values map[string]string, startBlockID string) frame { ... }
func newArmFrame(function core.Function, values map[string]string, blockID string) frame {
	f := newBlockFrame(function, values, blockID)
	f.singleBlockOnly = true
	return f
}
```
Change every `map[string]string` parameter/field to `map[string]value` where `type value struct{ tag, payload string }` (D-12-17). The three literal frame seeds (`map[string]string{function.Parameter.ID: input}` at lines ~206, ~220, ~287) become `map[string]value{function.Parameter.ID: {tag: "", payload: input}}` — the `tag: ""` zero-value is what makes scalar values evidence-invisible by construction (D-12-18).

---

### `internal/compiler/cgen/cgen.go` (service, transform)

**Analog:** the four repeated shared-kind case arms (this file has the SAME arm literally repeated at four emitter sites, per the D-12-33 ledger).

**Representative site** (`cgen.go:270-290`, `emitLinear`):
```go
for _, operation := range function.Linear.Operations {
	source := places[operation.SourceID]
	switch operation.Kind {
	case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
		target, exists := places[operation.TargetID]
		if !exists || declared[operation.TargetID] {
			return "", fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		label := "copy"
		marker := ""
		if operation.Kind == core.OpMove {
			label = "authority transfer"
			marker = " /* lang:mutation-site */"
		} else if operation.Kind == core.OpBorrowShared {
			label = "shared borrow representation"
		...
```
This exact shared-case-arm pattern repeats at `cgen.go:582`, `:770`, `:1773` (the four sites `emitLinear`/`emitLinearBorrowedByPointer`/`emitLinearBorrowedByPointerPlain`/`emitBranch` region, per D-12-33's re-verified ledger). Each needs its own `case core.OpConstructPayload:`/`case core.OpDestructurePayload:` arm emitting the flat-struct field write/read (D-12-22) against the shared `RecordLayout` fact `check.go`'s new function derives — **but only after** the D-12-31 six-emitter deletion lands, since `emitMatch` (`cgen.go:152`) is one of the six scheduled for deletion and D-12-32 names it as the exact collision point. Per D-12-31/D-12-35, write these `case` arms exactly once, against `emitProgram`'s single surviving family, never against the doomed six.

---

### `internal/compiler/cgen/cgen_program.go` (service, transform — inherited debt closure, NOT a RES-* deliverable)

**Analog:** `emitProgram` itself and the D-11-52 refusal string, same file — D-12-33's re-verified ledger:

```go
// emitProgram :128
// D-11-52 refusal ("multi-function branch bodies are not supported...") :152
// emitProgramFunction :339
```
Sequence per D-12-31: (i) land the N=1 convergence differential test green; (ii) lift the refusal at `cgen_program.go:152` (`"function %q: multi-function branch bodies are not supported by native emission this phase"`); (iii) delete the six `cgen.go` emitters (ledger below), re-pin the four golden-C digests **in a diff containing nothing else** (Pitfall 5); (iv) only then write payload `case` arms.

**D-12-33's verified emitter deletion ledger** (all in `cgen.go` unless noted):
| Emitter | Verified line |
|---|---|
| `emitMatch` | `cgen.go:152` |
| `emitLinear` | `cgen.go:234` |
| `emitLinearBorrowedByPointer` | `cgen.go:532` |
| `emitLinearBorrowedByPointerPlain` | `cgen.go:726` |
| `emitLinearForeign` | `cgen.go:870` |
| `emitBranch` | `cgen.go:1674` |

---

### `internal/compiler/pathoracle/pathoracle.go` (service, transform — body-blind by design)

**Analog:** the existing membership test at line 324 — this file never switches exhaustively on kind (D-04-29, confirmed in `core_test.go`'s own comment).

```go
// Source: internal/compiler/pathoracle/pathoracle.go:320-326
if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
	...
}
```
No new case arm is required here for the exhaustive-dispatch control to pass ("handled" for `pathoracle` means the walk completes without error per `core_test.go:267`'s own comment) — but per D-12-02's probe methodology, verify a payload-carrying fixture's path enumeration still terminates and produces correct endpoints via `BuildCalleeLookup`/`RecomputeEndpoints` (cited in the accepted probe, D-12-04a).

---

### `internal/compiler/originvalidate/originvalidate.go` (service, transform)

**Analog:** `walkReturnOrigin`'s own `switch operation.Kind` (`originvalidate.go:243-270`) — the exact site D-12-28 says must NOT gain a `case core.OpCall:` this phase, but DOES need to understand payload ops for origin propagation through a destructure.

```go
switch operation.Kind {
case core.OpBorrowExclusive:
	if derivedAccess == "" {
		derivedAccess = "exclusive"
	}
case core.OpBorrowShared:
	if derivedAccess == "" {
		derivedAccess = "shared"
	}
case core.OpForeignCall:
	...
case core.OpCall:
	// D-10-01/D-10-07: consult the CALLEE's DECLARED return contract ...
```
If a payload's origin must walk backward through an `OpDestructurePayload` (matching's fresh place) to find the original borrow, add a `case core.OpDestructurePayload:` arm here that continues the walk through `SourceID` — this is walking through a NEW kind's OWN semantics, not filling `core.OpCall`'s gap, so it does not trip D-12-28's guard (the distinction Pitfall 3 names).

---

### `internal/compiler/session/session.go` (service, batch — anti-vacuity control)

**Analog:** `LayoutMutationRunner` / `control:foreign.layout_mismatch` (`session.go:565-604`), the exact pattern D-12-37 clones.

```go
type LayoutMutationRunner struct {
	Runner      native.Runner
	Contract    *core.ForeignContract
	FixturePath string
}

// Run assembles a minimal single-function core.Program carrying only
// r.Contract, generates its conformance unit against r.FixturePath, and
// compiles it as its own separate, bounded invocation. A nil error means the
// fixture at FixturePath conforms to r.Contract's declared layout; a
// *native.ToolError with code "native.conformance_failed" means it was
// refused at compile time.
func (r LayoutMutationRunner) Run(ctx context.Context) error {
	program := core.Program{
		Schema: core.Schema1, Module: "phase4.layout_probe", ModuleID: "phase4.layout_probe",
		Functions: []core.Function{{
			ID: "phase4.layout_probe:fn:probe", Name: "probe",
			EntryPointID: "phase4.layout_probe:fn:probe:point:entry", ReturnPointID: "phase4.layout_probe:fn:probe:point:return",
			ForeignContract: r.Contract,
		}},
	}
	...
}
```
Clone this shape into a `PayloadLayoutMutationRunner` (or equivalent name): a frozen `testdata/phase12/*.golden.c` declaring the payload struct's alternative slots transposed/resized relative to the checker-derived `RecordLayout`, killed by the generated `_Static_assert`/`offsetof` pair under `-Werror` (D-12-37). This is **necessary but not sufficient** — pair it with D-12-38's reverted-production-hunk `cgen` mutation (see the `cgen.go` section above), which is the decisive control, run against the **existing** `session_phase5_compare.go:89-96` `Outcome.Value` axis with **no modification to that file**.

---

## Shared Patterns

### Additive-omitempty, byte-freeze discipline
**Source:** `internal/compiler/core/core.go:722-762` (the `OkEdgeID`/`ErrEdgeID`/`ErrTargetID` and `CalleeID` doc comments)
**Apply to:** every new field on `DataType`, `LinearOperation`, `Alternative`, `MatchArm` — each must be `omitempty`/pointer/empty-string and carry the exact doc-comment shape: "populated only on Op<X>; every pre-Phase-12 operation, and every operation kind other than Op<X>, leaves this empty." This is what keeps `core_test.go:156-159`'s four pinned golden-C digests byte-identical (D-12-08, D-12-18, Pitfall 5).

### Named fail-closed refusal shape
**Source:** `internal/compiler/check/check.go:1628-1637` (`diagnostic.Error("<namespace>.<code>", span, "<message>")` + `continue`)
**Apply to:** all of D-12-15's three new refusals (arity mismatch, binder-on-nullary, missing-binder-for-payload) and D-12-27's resource-payload refusal. Namespace as `check.*` (semantic, needs type info) or `syntax.*` (grammar-level), matching existing `core.callee_not_callable` / `syntax.expected_pattern` / `syntax.arm_limit` shape.

### Exhaustive-dispatch anti-vacuity control
**Source:** `internal/compiler/core/core_test.go:267-337` (`TestAllOperationKindsHandledAtEverySite`, `exhaustiveDispatchFixtures`, `runExhaustiveDispatchControl`)
**Apply to:** every plan touching `AllOperationKinds()` — new fixtures MUST be appended to `exhaustiveDispatchFixtures`, and must stress the new kinds interacting with existing borrow/call machinery (Pitfall 1), never only a trivial happy path.

### Independent peer re-derivation (four admission peers)
**Source:** `internal/compiler/corevalidate/corevalidate.go` (two switches, lines 1667/1709/1924/1966/2432), `internal/compiler/originvalidate/originvalidate.go:243-270`, `internal/compiler/pathoracle/pathoracle.go:320-326`
**Apply to:** any plan adding payload-kind semantics to independent validators. Each peer's arm must be independently written (never copy-pasted from `check`'s own emission logic) per D-09-02's "independence by derivation method plus import boundary" — the deliberate narrow exception is *layout only* (D-12-25), never *judgement* logic.

### Shared checker-derived `RecordLayout` fact
**Source:** `internal/compiler/check/check.go:3118-3125` (`standardForeignLayout`), `internal/compiler/core/core.go:111-127`-area (`RecordLayout`/`LayoutField` — re-verify exact line before use, `ARCHITECTURE.md` cites :79-130 for the type block)
**Apply to:** the payload struct layout derivation — one function, one shared fact, read identically by `check` (deriving), the four peers (consuming), and `cgen` (emitting). Never re-derive layout independently in `cgen` or a peer (D-12-25's deliberate exception to independence).

### Five-axis differential comparator — extend nothing
**Source:** `internal/compiler/session/session_phase5_compare.go:20-24` (axis constants), `:89-96` (`comparePhase5Pair`'s `Outcome.Value` comparison, "INCLUDING the ok payload and the err-edge ADT alternative")
**Apply to:** D-12-38's decisive anti-vacuity control and any other payload-value differential check. Do **not** add a sixth axis or a new comparator field (D-12-40, D-11-42's standing constraint) — the payload-value channel already exists.

## No Analog Found

None. Every touched file already contains the precedent shape (an existing `OperationKind` case arm, an existing additive-omitempty field, or an existing mutation-control runner) that the new work extends — this is the expected shape for a phase explicitly scoped as "extend the six literal dispatch sites," per D-12-05/D-04-22.

## Metadata

**Analog search scope:** `internal/compiler/{core,ast,syntax,check,corevalidate,interp,cgen,pathoracle,originvalidate,session}` — the exact package set RESEARCH.md's "Recommended Project Structure" names, plus `core_test.go`/`session_test.go` for control-shape precedent.
**Files scanned:** 13 (all git-tracked, verified via `git ls-files`)
**Pattern extraction date:** 2026-09-12
