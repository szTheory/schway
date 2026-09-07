# Phase 2: Owned Values and Abilities - Pattern Map

**Mapped:** 2026-09-03
**Files analyzed:** 25 new/modified implementation and verification targets
**Analogs found:** 25 / 25
**Scope:** `OWN-01`, `OWN-02`; straight-line ownership only

## Mapping Posture

Phase 2 should extend the tracked production spine under `internal/compiler`, not
copy a spike wholesale. The spikes are executable semantic precedents; the Phase
1 packages are the production conventions for package boundaries, stable facts,
strict decoding, orchestration, and verification.

The five strongest tracked analog families are:

1. `internal/compiler/{syntax,ast,core,check,interp,cgen,native,session}` — the
   source-to-native vertical slice to extend.
2. `.planning/spikes/001-ownership-kernel-workbench/ownership` — independently
   implemented affine state transitions, cause-rich diagnostics, exhaustive
   short programs, and a load-bearing move-during-loan counterexample.
3. `.planning/spikes/003-public-origins-generic-abilities/origins` — independent
   per-ability structural derivation through generic shapes.
4. `.planning/spikes/004-independent-certificate-checker/{verifier,experiment}`
   — source-blind validation, mutation matrices, stable diagnostics, and counted
   linear work.
5. `internal/compiler/{evidence,protocol}` plus `scripts/verify-phase1.sh` — strict
   JSON, content binding, one human/JSON fact model, fail-closed lanes, and the
   bounded offline gate.

All named analogs were checked with `git ls-files`; no generated, ignored, or
runtime-mirror path is used here.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/syntax/token.go` | model | transform | `internal/compiler/syntax/token.go` | exact extension |
| `internal/compiler/syntax/lexer.go` | utility | transform | `internal/compiler/syntax/lexer.go` | exact extension |
| `internal/compiler/syntax/parser.go` | service | transform | `internal/compiler/syntax/parser.go` | exact extension |
| `internal/compiler/syntax/format.go` | utility | transform | `internal/compiler/syntax/format.go` | exact extension |
| `internal/compiler/syntax/syntax_test.go` | test | transform | `internal/compiler/syntax/syntax_test.go` | exact extension |
| `internal/compiler/ast/ast.go` | model | transform | `internal/compiler/ast/ast.go` | exact extension |
| `internal/compiler/core/core.go` | model | transform | `internal/compiler/core/core.go` | exact extension |
| `internal/compiler/ability/ability.go` | service | transform | `.planning/spikes/003-public-origins-generic-abilities/origins/abilities.go` | role/data-flow exact |
| `internal/compiler/ability/ability_test.go` | test | transform | `.planning/spikes/003-public-origins-generic-abilities/origins/origins_test.go` | role/data-flow exact |
| `internal/compiler/check/check.go` | service | transform | `.planning/spikes/001-ownership-kernel-workbench/ownership/checker.go` | behavior exact, boundary differs |
| `internal/compiler/check/check_test.go` | test | transform | `.planning/spikes/001-ownership-kernel-workbench/ownership/ownership_test.go` | role/data-flow exact |
| `internal/compiler/corevalidate/corevalidate.go` | service | transform | `.planning/spikes/004-independent-certificate-checker/verifier/verifier.go` | role/data-flow exact |
| `internal/compiler/corevalidate/corevalidate_test.go` | test | transform | `.planning/spikes/004-independent-certificate-checker/experiment/experiment_test.go` | role/data-flow exact |
| `internal/compiler/diagnostic/diagnostic.go` | model | request-response | `.planning/spikes/001-ownership-kernel-workbench/ownership/model.go` | role match |
| `internal/compiler/execution/execution.go` | model | event-driven | `internal/compiler/interp/interp.go` | extraction exact |
| `internal/compiler/interp/interp.go` | service | event-driven | `.planning/spikes/001-ownership-kernel-workbench/ownership/oracle.go` | role/data-flow exact |
| `internal/compiler/cgen/cgen.go` | service | transform | `internal/compiler/cgen/cgen.go` | exact extension |
| `internal/compiler/native/native.go` | service | request-response | `internal/compiler/evidence/evidence.go` | strict-decoder role match |
| `internal/compiler/native/native_test.go` | test | request-response | `internal/compiler/evidence/evidence_test.go` | strict-input role match |
| `internal/compiler/protocol/protocol.go` | model | request-response | `internal/compiler/protocol/protocol.go` | exact extension |
| `internal/compiler/evidence/evidence.go` | service | transform | `internal/compiler/evidence/evidence.go` | exact extension |
| `internal/compiler/evidence/evidence_test.go` | test | transform | `internal/compiler/evidence/evidence_test.go` | exact extension |
| `internal/compiler/session/session.go` | service/provider | request-response | `internal/compiler/session/session.go` | exact extension |
| `internal/compiler/session/session_test.go` | test | request-response/event-driven | `internal/compiler/session/session_test.go` | exact extension |
| `testdata/phase2/*.lang`, `scripts/verify-phase2.sh` | test/config | file-I/O/batch | `testdata/phase1/*`, `scripts/verify-phase1.sh` | exact extension |

## Pattern Assignments

### Frontend: `syntax/*` and `ast/ast.go`

**Primary analogs:** `internal/compiler/syntax/parser.go`,
`internal/compiler/syntax/format.go`, `internal/compiler/syntax/syntax_test.go`

Preserve the lossless boundary: tokens retain all bytes and the AST is only a
typed projection. Add real tokens for `let`, `take`, and `borrow`, plus `<`, `>`,
and `,`; do not detect these forms by identifier text after lexing.

**Closed parser dispatch pattern** (`internal/compiler/syntax/parser.go:35-50`):

```go
for p.peek().Kind != TokenEOF {
	switch p.peek().Kind {
	case TokenData:
		program.Data = append(program.Data, p.dataDecl())
	case TokenFn:
		program.Funcs = append(program.Funcs, p.funcDecl())
	default:
		p.problem("syntax.expected_declaration", p.peek(), "expected `data` or `fn` declaration")
		p.recoverUntil(TokenData, TokenFn, TokenEOF)
	}
}
```

Use the same closed dispatch inside a function body: select `match` versus
`linear` from the first non-trivia token and require exactly one AST body variant.
Do not represent a linear body as a special empty match.

**Bounded progress/recovery pattern** (`internal/compiler/syntax/parser.go:227-247`):

```go
func (p *parser) recoverUntil(boundaries ...Kind) {
	if p.atAny(boundaries...) { return }
	start := p.position
	for !p.atAny(boundaries...) { p.advance() }
	if p.position <= start { panic("parser recovery made no progress") }
}

func (p *parser) assertProgressOrBoundary(start int, boundaries ...Kind) {
	if p.position > start || p.atAny(boundaries...) { return }
	panic("parser branch made no progress")
}
```

Every ownership/type-argument recovery loop must use this advance-or-boundary
contract. Add explicit type depth and node counters at the `typeRef` parser seam;
return bounded diagnostics instead of allowing unbounded recursion.

**Formatter-owned token projection** (`internal/compiler/syntax/format.go:8-18`):

```go
func Format(tree Tree) []byte {
	f := formatter{contexts: make([]string, 0, 3)}
	for _, token := range tree.Tokens {
		if token.Kind == TokenEOF || token.Kind == TokenWhitespace { continue }
		f.token(token)
	}
	return append(bytes.TrimRight([]byte(f.out.String()), " \t\r\n"), '\n')
}
```

Extend token rendering in place. Canonical linear bodies should render one
binding per line with two-space indentation; type applications must not insert
spaces before `<` or `>`, and must use one space after commas.

**Property-test pattern** (`internal/compiler/syntax/syntax_test.go:138-168`):

```go
f.Fuzz(func(t *testing.T, source []byte) {
	parsed := syntax.Parse(source)
	if !bytes.Equal(parsed.Tree.Bytes(), source) { t.Fatal("CST round trip changed fuzz input") }
	if len(parsed.Diagnostics) > 21 { t.Fatalf("unbounded diagnostics: %d", len(parsed.Diagnostics)) }
	if len(parsed.Diagnostics) != 0 { return }
	first := syntax.Format(parsed.Tree)
	formatted := syntax.Parse(first)
	second := syntax.Format(formatted.Tree)
	if !bytes.Equal(first, second) { t.Fatalf("formatter is not idempotent") }
})
```

Add all Phase 2 fixtures as seeds and extend span erasure for the closed body
union. Preserve the 20-primary-plus-truncation cap.

### Typed facts: `core/core.go` and `execution/execution.go`

**Primary analogs:** `internal/compiler/core/core.go:7-51`,
`internal/compiler/interp/interp.go:10-31`

Keep core as inert serializable records with explicit IDs and JSON tags. Add
`TypeRef`, ordered ability facts/witnesses, places, loans, and linear operations;
use slices in stable order on the wire, not maps. Function bodies become a
validated closed union of match or linear records.

**Stable point/edge fact pattern** (`internal/compiler/core/core.go:22-49`):

```go
type Function struct {
	ID            string `json:"id"`
	EntryPointID  string `json:"entry_point_id"`
	ReturnPointID string `json:"return_point_id"`
	// ... explicit body fact
}

type MatchArm struct {
	ID     string `json:"id"`
	EdgeID string `json:"edge_id"`
}
```

Linear operation IDs should similarly be `function ID + source-order ordinal`,
with explicit point IDs. Do not derive semantic identity from byte offsets.

Extract the execution records from `interp` without behavior:

**Current inert fact shape** (`internal/compiler/interp/interp.go:10-31`):

```go
type Execution struct {
	Schema        string   `json:"schema"`
	Outcome       Outcome  `json:"outcome"`
	Events        []Event  `json:"events"`
	LiveResources []string `json:"live_resources"`
}
```

Move these records and canonical encoding into `internal/compiler/execution`.
Both interpreter and native decoder depend on that package; `execution` must not
import either engine. Preserve `lang.execution/0` for Phase 1 and introduce the
owned-body schema deliberately rather than silently changing the old wire form.

### Ability engine: `ability/ability.go` and tests

**Primary analog:**
`.planning/spikes/003-public-origins-generic-abilities/origins/abilities.go`

Port the per-ability structural fold, not the open registry or public-origin
machinery. The real compiler owns sealed roots `Byte` and `Buffer` and sealed
constructors `Box` and `Pair`.

**Independent ordered derivation** (`.../origins/abilities.go:36-74`):

```go
func (r Registry) Has(expr TypeExpr, ability Ability, substitutions map[string]TypeExpr) (bool, error) {
	// derive one requested ability through the shape
}

func (r Registry) Abilities(expr TypeExpr, substitutions map[string]TypeExpr) ([]Ability, error) {
	var abilities []Ability
	for _, ability := range AbilityOrder {
		has, err := r.Has(expr, ability, substitutions)
		if err != nil { return nil, err }
		if has { abilities = append(abilities, ability) }
	}
	return abilities, nil
}
```

Required divergence: return a deterministic negative witness path for every
missing ability and memoize by canonical structural key. Enforce acyclic sealed
shapes plus parser/type-node caps. Do not add ability implications, user-authored
positive roots, traits, recursive aliases, or a general constraint solver.

**Exhaustive independence test**
(`.planning/spikes/003-public-origins-generic-abilities/origins/origins_test.go:109-141`):

```go
for mask := 0; mask < 1<<len(AbilityOrder); mask++ {
	// construct every atomic mask
	for _, ability := range AbilityOrder {
		// compare production derivation with an independently coded oracle
	}
}
```

Extend this pattern to all 32 leaf masks through `Box` and all 1,024 left/right
mask pairs through `Pair`. The test oracle must not call the production fold,
canonical-key helper, or witness-path helper.

### Ownership lowering: `check/check.go`, `check/check_test.go`

**Primary analog:**
`.planning/spikes/001-ownership-kernel-workbench/ownership/checker.go`

Retain `check.Program(ast.Program) Result` as the frontend/core seam, but factor
small private helpers for type resolution, ability materialization, straight-line
last-use discovery, and forward ownership lowering. The checker decides whether
a bare RHS is `copy`; only `take` can lower a noncopyable transfer.

**Independent state-machine posture** (`.../ownership/checker.go:29-55`):

```go
// Check statically interprets every instruction. It intentionally does not
// call the oracle transition implementation.
func Check(source Program, options CheckerOptions) Result {
	result := Result{Program: source.ID, Valid: true, Events: []Event{}}
	places := map[string]*checkerPlace{}
	loans := map[string]*checkerLoan{}
	// local record/reject helpers keep causes attached at the decision point
}
```

**Move legality and transition** (`.../ownership/checker.go:205-227`):

```go
case Move:
	place, failure := lookup(operation)
	if failure != nil { return *failure }
	blocked := place.writeLoan != "" || len(place.readLoans) != 0
	if blocked {
		return reject(operation, "ownership.move_while_borrowed", "cannot transfer ownership while a loan is live")
	}
	// initialize target, then mark source moved
```

Required divergence: Phase 2 has only shared loans, straight-line precomputed
last uses, and no release/scope/exit/resource semantics. Preserve distinct source
spans for declaration, borrow, move, later loan use, and invalid read. Count
operations and type nodes explicitly for scale assertions.

**Counterexample pattern**
(`.planning/spikes/001-ownership-kernel-workbench/ownership/ownership_test.go:122-146`):

```go
func TestFaultInjectionProducesMinimalCounterexample(t *testing.T) {
	result := FindMismatch(3, CheckerOptions{AllowMoveWhileShared: true})
	// The later loan read is load-bearing: without it, last-use inference ends
	// the unused loan immediately and moving the owner is legal.
	if len(result.MinimalInput.Operations) != 4 { /* fail */ }
}
```

Port the four-step semantic shape—declare, shared borrow, attempted move, later
loan read—into the source fixture. Also exhaust the reduced short-operation
alphabet and compare intermediate operation/last-use facts, not only pass/fail.

### Independent core validator: `corevalidate/*`

**Primary analog:**
`.planning/spikes/004-independent-certificate-checker/verifier/verifier.go`

The package may import `core`, `ability` only for shared inert names/types if
necessary, and `diagnostic`; it must not import `ast`, `check`, `interp`, or
`cgen`. Prefer duplicating the small ability and transition laws over sharing
authorization logic.

**Source-blind entry and counted checks** (`.../verifier/verifier.go:12-50`):

```go
type Result struct {
	Valid       bool                `json:"valid"`
	Diagnostics []format.Diagnostic `json:"diagnostics,omitempty"`
	Checks      int                 `json:"checks"`
}

// Verify is body- and source-blind. It validates declared typed-core facts
// without importing or calling the producer implementation.
func Verify(artifact format.Artifact, certificate format.Certificate) Result {
	checker := checker{artifact: artifact, certificate: certificate}
	checker.run()
	return Result{Valid: len(checker.diagnostics) == 0, Diagnostics: checker.diagnostics, Checks: checker.checks}
}
```

**Unique-ID validation** (`.../verifier/verifier.go:53-85`) and **independent
replay** (`.../verifier/verifier.go:263-325`) are the direct patterns for checking
operation/place/type/loan IDs and linear move/loan facts.

Required mutations: forged `copy`, forged positive ability, omitted move,
duplicate operation ID, unknown source place, wrong type reference, and reordered
event/final claim. Keep `escape:coordinated-source-core-lie` explicit; internal
consistency is not source-translation proof.

**Mutation/complexity tests**
(`.planning/spikes/004-independent-certificate-checker/experiment/experiment_test.go:38-51,123-137`):

```go
for _, mutation := range report.Mutations {
	if mutation.Expected != "ESCAPES_BY_DESIGN" && mutation.Code != mutation.Expected { /* fail */ }
}

for _, events := range []int{101, 1001, 10001} {
	result := verifier.Verify(artifact, certificate)
	if result.Checks > 2*events+8 { /* fail: not linear */ }
}
```

### Diagnostics: `diagnostic/diagnostic.go`

**Primary analogs:** `internal/compiler/diagnostic/diagnostic.go`,
`.planning/spikes/001-ownership-kernel-workbench/ownership/model.go:54-75,104-128`

Add a sorted semantic `repairs` array and deliberately bump to
`lang.diagnostic/1`. Keep message prose out of identity; include stable semantic
causes and repair kinds in the ID input.

**Current identity boundary** (`internal/compiler/diagnostic/diagnostic.go:33-42`):

```go
identity := struct {
	Schema string
	Code   string
	Span   Span
	Causes []Cause
}{Schema: Schema, Code: code, Span: span, Causes: causes}
encoded, _ := json.Marshal(identity)
```

**Repair vocabulary precedent** (`.../ownership/model.go:104-121`):

```go
case "ownership.move_while_borrowed":
	return []Repair{{Kind: "move_after_last_use", Effect: "..."}}
case "ownership.use_after_move":
	return []Repair{{Kind: "use_move_target", Effect: "..."}}
```

Use the narrower research names (`use_transfer_target`,
`move_use_before_transfer`, `move_after_last_borrow_use`, `insert_take`, etc.).
Never suggest granting `copy` to a unique value as a generic repair.

### Interpreter, C17, and native decoder

**Primary analogs:** `internal/compiler/interp/interp.go`,
`internal/compiler/cgen/cgen.go`, `internal/compiler/native/native.go`,
`internal/compiler/evidence/evidence.go`

The interpreter executes validated core operations with its own place/loan
state. It must not call the static checker or core validator transition helper.
The C emitter consumes explicit core operations and emits one bounded canonical
JSON `execution.Execution`; it must not infer copy versus move from source.

**Readable deterministic C builder** (`internal/compiler/cgen/cgen.go:22-54`):

```go
var out strings.Builder
out.WriteString("/* generated by Codename Lang; schema lang.c17/0 */\n")
// emit explicit checked declarations and operations in stable order
return out.String(), nil
```

For the owned tracer, use a fixed inline `LANG_BUFFER` representation. An
ordinary C assignment may implement a language move; the proof concerns
authority/liveness and event order, not zero machine copies. Emit no `restrict`,
`noalias`, heap allocator, cleanup, or ABI claim in this phase.

**Shell-free bounded process pattern** (`internal/compiler/native/native.go:61-95`):

```go
ctx, cancel := context.WithTimeout(parent, r.Timeout)
defer cancel()
command := exec.CommandContext(ctx, r.ClangPath,
	"-std=c17", "-Wall", "-Wextra", "-Werror", "-pedantic",
	optimization, sourcePath, "-o", binaryPath)
```

Preserve literal argv, temporary-directory isolation, timeouts, and distinct
tool error codes.

**Strict JSON decoder pattern** (`internal/compiler/evidence/evidence.go:144-155`):

```go
decoder := json.NewDecoder(bytes.NewReader(data))
decoder.DisallowUnknownFields()
if err := decoder.Decode(&value); err != nil { /* invalid */ }
var trailing any
if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) { /* trailing */ }
```

Wrap the native reader with an explicit 64 KiB maximum and reject oversized,
unknown-field, and trailing-value output. Native results should hold decoded
executions rather than `Input`/`Output` string pairs.

### Protocol, evidence, and session orchestration

**Primary analogs:** `internal/compiler/protocol/protocol.go`,
`internal/compiler/evidence/evidence.go`, `internal/compiler/session/session.go`

Keep `lang.command/0` and the exit taxonomy. Move protocol's execution import
from `interp` to the inert `execution` package. Human rendering needs stable
ownership event fields and should remain a projection of the same result record.

**Semantic result identity pattern** (`internal/compiler/protocol/protocol.go:72-110`):

```go
for _, execution := range result.Executions {
	encodedExecution, _ := json.Marshal(execution)
	executionSum := sha256.Sum256(encodedExecution)
	identity.ExecutionDigests = append(identity.ExecutionDigests, hex.EncodeToString(executionSum[:]))
}
```

Do not include timing, RSS, physical addresses, or C process details in semantic
identity.

In `session.Check`, preserve parse-before-check fail closed behavior. After
checking, run `corevalidate` before any engine. In `RunNative`, compare complete
canonical executions rather than Phase 1's output-only check:

**Current seam to replace** (`internal/compiler/session/session.go:188-207`):

```go
cSource, err := cgen.Emit(checked.Program)
o0, err := runner.Run(ctx, cSource, "-O0", inputs)
o3, err := runner.Run(ctx, cSource, "-O3", inputs)
for index, expected := range interpreted {
	for _, actual := range []native.Result{o0, o3} {
		if actual.Pairs[index].Output != expected.Outcome.Value { /* mismatch */ }
	}
}
```

Replace string comparison with canonical `Outcome + ordered Events +
LiveResources` comparison and retain optimization/input in mismatch context.

Evidence should select core/execution schemas from the admitted body feature,
bind owned-core bytes and generated C, and validate core before accepting the
binding. Preserve the trust-boundary comment:

**Honest evidence claim** (`internal/compiler/evidence/evidence.go:97-101`):

```go
// This manifest proves that these compiler products are mutually bound. It
// does not independently prove that a coordinated frontend translated user
// intent into the correct core.
```

Do not rewrite Phase 1 goldens just because Phase 2 exists. Feature-specific
`lang.core/1` and `lang.execution/1` artifacts coexist with Phase 1 `/0`.

### Fixtures, tests, and `scripts/verify-phase2.sh`

**Primary analogs:** `internal/compiler/session/session_test.go`,
`internal/compiler/evidence/evidence_test.go`, `scripts/verify-phase1.sh`

Add canonical fixtures:

- `owned_transfer.lang`
- `implicit_copy.lang`
- `use_after_move.lang`
- `move_while_borrowed.lang`
- `implicit_noncopy.lang`

**End-to-end test shape** (`internal/compiler/session/session_test.go:39-51`):

```go
result, diagnostics, err := session.RunNativeFile(context.Background(), path, native.DefaultRunner())
if err != nil || len(diagnostics) != 0 { t.Fatalf("native run failed") }
// assert readable C plus interpreter/O0/O3 semantic equivalence
```

**Mutation table shape** (`internal/compiler/evidence/evidence_test.go:43-76`):

```go
tests := []struct {
	name string
	code string
	edit func(*evidence.Manifest)
}{ /* one represented boundary per row */ }
for _, test := range tests { t.Run(test.name, func(t *testing.T) { /* mutate and assert exact code */ }) }
```

**Bounded gate shape** (`scripts/verify-phase1.sh:1-11`):

```sh
set -eu
verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase1.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"
go test ./...
go test -race ./...
go vet ./...
go run ./cmd/lang verify testdata/phase1
```

Phase 2's script should execute the shared Go suite once, race once, vet once,
then explicitly verify both Phase 1 and Phase 2 corpora. Do not invoke the Phase
1 script from it because that would duplicate Go test/race/vet work.

The Phase 2 verifier fails closed unless these exact controls are observed:

```text
control:ownership.use_after_move
control:ownership.move_while_borrowed
control:ownership.transfer_requires_take
control:ability.forged_copy
control:core.duplicate_operation_id
control:interpreter-o0-o3-owned
control:evidence.core_mismatch
```

Record `escape:coordinated-source-core-lie` separately as a named expected trust
boundary, never as a detected control. Every lane performs nonzero counted work.

## Shared Patterns

### Dependency Direction

```text
syntax -> ast -> check -> core
                   |       |
                   v       v
                ability  corevalidate
                           |
             +-------------+-------------+
             v                           v
           interp                      cgen -> native
             \                           /
              +------ execution --------+

session -> all orchestration packages
protocol/evidence -> inert facts, never engine internals
```

`corevalidate`, `interp`, and `check` must not share the state-transition
implementation. `execution` owns data only. `native` does not parse prose.

### Stable Identity

- Semantic IDs derive from module/function identity and source-order semantic
  ordinals, never byte offsets, object addresses, map iteration, temp paths, or
  timings.
- Diagnostic IDs include schema, code, semantic causes, primary span, and
  repair kinds; message prose stays display-only.
- Ordered semantic events are identity-bearing; event reorder must change the
  digest and fail differential evidence.

### Error and Authority Boundaries

- Parse/type/ownership rejection is invalid input (exit 2).
- Missing Clang, timeout, malformed native output, and filesystem failure are
  operational (exit 3).
- Interpreter/O0/O3 disagreement is semantic mismatch (exit 4).
- No invalid or unvalidated core reaches interpreter or C generation.
- No source declaration can grant a positive primitive ability in Phase 2.

### Performance and Boundedness

- Count inspected operations and type nodes; assert linear work at 10, 100,
  1,000, and 10,000 operations.
- Memoize structural ability derivation by canonical type key.
- Cap type depth/nodes and native output bytes.
- Keep coverage-guided fuzzing out of the default gate; ordinary `go test` runs
  deterministic seeds, exhaustive short sequences, and fixed-seed properties.

## Divergences to Avoid

| Tempting copy | Why it must not be copied |
|---|---|
| All of Spike 001's scopes, resources, exclusive loans, release, panic/cancel paths | These belong to Phases 3–4 and would blur OWN-01. |
| Spike 003's open registry, public origins, callbacks, substitutions | Phase 2 needs sealed shapes and independent abilities only. |
| Spike 004's full certificate modes and public-summary checking | Phase 2 needs compact core consistency validation, not a certificate subsystem. |
| Phase 1 native `Pair{Input, Output}` | It loses ownership event order and moved-place facts. |
| Shared transition helpers across checker/interpreter/validator | Agreement would become tautological and mutation controls weaker. |
| Global core/execution schema bump | It would create unrelated Phase 1 golden churn. |
| Heap-backed buffer or C alias attributes | It prematurely claims allocation, cleanup, provenance, or optimizer facts. |
| General generics/traits/recursive aliases | It turns five structural folds into an open solver before evidence demands one. |

## No Analog Found

None. Every Phase 2 target has either an exact production seam or a tracked
executable spike analog. New package names (`ability`, `corevalidate`, and
`execution`) are new boundaries, but their internal patterns are established.

## Metadata

**Analog search scope:** `cmd/lang`, `internal/compiler`, `scripts`,
`testdata/phase1`, and tracked Spikes 001, 003, and 004

**Files scanned closely:** 31 source/test/planning files

**Pattern extraction date:** 2026-09-03

**Planner note:** Split work vertically enough that the first plan produces a
real `Byte` copy and `Buffer` transfer through check/interpreter/native, then add
independent derivation/validation pressure, then close with full controls,
evidence, and the Phase 2 gate. Avoid a frontend-only or checker-only wave.
