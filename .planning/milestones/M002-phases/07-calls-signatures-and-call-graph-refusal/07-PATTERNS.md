# Phase 07: Calls, Signatures, and Call-Graph Refusal - Pattern Map

**Mapped:** 2026-09-08
**Files analyzed:** 14 new/modified files
**Analogs found:** 13 / 14 (one new package modelled on `pathoracle`; one fixture family with no in-tree precedent for its *shape*, only its conventions)

All analog paths below were verified git-TRACKED (`git ls-files`). No gitignored
mirror paths appear in this document. All line numbers were read this session.

---

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `internal/compiler/callgraph/callgraph.go` (NEW) | service (whole-program derivation) | graph traversal / transform | `internal/compiler/pathoracle/pathoracle.go` | exact (posture + typed-error) |
| `internal/compiler/callgraph/callgraph_test.go` (NEW) | test | synthetic-input transform | `internal/compiler/pathoracle/pathoracle_test.go` | exact |
| `internal/compiler/core/core.go` (`Interface`/`FunctionSignature` → `/1`) | model / schema | data-shape | `core.go:148-171` self + `core.go:270-301` (`LinearOperation` additive fields) + `diagnostic.go:12-13` (`Schema`/`Schema1` pairing) | exact |
| `internal/compiler/core/core.go` (`AllOperationKinds()` + `OpCall` const) | config / registry | registry | `core.go:212-233, 244-252` (`OpForeignCall`/`OpRelease`/`OpDefect` registrations) | exact |
| `internal/compiler/originvalidate/originvalidate.go` (`BuildInterface` widened) | service (producer) | transform | `originvalidate.go:408-434` self | exact |
| `internal/compiler/corevalidate/corevalidate.go` (`case core.OpCall:` ×2) | middleware (validator) | request-response replay | `corevalidate.go:960-972` (straight-line `OpForeignCall`) and `:1157-1170` (blocks `OpForeignCall`); `:905-912`/`:946-956` for the simpler `OpCopy`/`OpMove` single-target shape | exact |
| `internal/compiler/corevalidate/corevalidate.go` (signature-summary peer) | service (peer re-derivation) | transform | `corevalidate.go:352-361` (`isDeclaredFunctionName` re-derivation with its "materially different mechanism" doc comment) | role-match |
| `internal/compiler/corevalidate/corevalidate.go` (own 3-color traversal peer) | service (peer) | graph traversal | `pathoracle.go:112-125` (`backEdgeError`) | role-match |
| `internal/compiler/syntax/parser.go` (new `ast.RHS.Kind` "call") | controller (parser) | streaming token consumption | `parser.go:456-465` (`tryCallBinding`) + `parser.go:404-421` (the refusal being relocated) | exact |
| `internal/compiler/ast/ast.go` (`RHS` doc for the new Kind) | model | data-shape | `ast.go:143-158` self | exact |
| `internal/compiler/check/check.go` (Lang-call admission + callgraph invocation) | service (producer) | request-response | `check.go:1312-1350` (`resolveForeignStep`, shape only — NOT the code path) + `check.go:1822-1825` (`places` scope map) | role-match |
| `internal/compiler/interp/interp.go` (`case core.OpCall:`) | service | event emission | `interp.go:97-107` (`OpForeignCall`, the documented-unreachable precedent) | exact |
| `internal/compiler/cgen/cgen.go` (`case core.OpCall:` ×4) | service | transform | `cgen.go:1723-1729` (`case core.OpForeignCall, core.OpFail:` known-but-unsupported arm) | exact |
| `internal/compiler/session/session.go` (phase-07 dispatch lane + `/1` export) | controller / config | request-response | `session.go:2501-2573` (`lane:kind-exhaustive-dispatch`) + `session.go:1023-1064` (`InterfaceExportCommandFile`) | exact |
| `internal/compiler/core/core_test.go` (fixture list + required kinds) | test | fixture sweep | `core_test.go:266-357` (`TestAllOperationKindsHandledAtEverySite`) | exact |
| `testdata/phase07/*.lang` (NEW corpus) | fixture | file-I/O | `testdata/phase4/foreign_acquire_one.lang`, `testdata/phase3/shared_shared_accept.lang` | role-match (no multi-function fixture exists in-tree) |

---

## Pattern Assignments

### `internal/compiler/callgraph/callgraph.go` (NEW package, service, graph traversal)

**Analog:** `internal/compiler/pathoracle/pathoracle.go`

**Package-doc posture** — copy this structure verbatim: what the package owns,
what it never imports, what it reads, and *why it is not obtainable from the
other derivations by renaming* (`pathoracle.go:1-29`):

```go
// Package pathoracle is Phase 3's third, genuinely different decision
// procedure for loan liveness (03-RESEARCH Q4, ROADMAP criterion 1). It is
// consumed only by tests and the verify harness (session.go) -- never by
// check, corevalidate, interp, or cgen -- and it reads nothing but a
// core.Function's own declared Blocks/Edges/Operations. It imports neither
// check nor corevalidate nor ast, and it calls no production liveness
// function.
//
// Three loan-endpoint derivations now coexist in this repository, and none
// can be obtained from either of the others by renaming ...
package pathoracle

import (
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/core"
)
```

**Deviation the planner must state explicitly in `callgraph`'s own doc comment:**
`pathoracle` is *test/verify-harness-only*; `callgraph` is imported by `check` in
production (D-07-14). The doc comment must say "consumed by `check` before it
returns a `core.Program`, and never by `corevalidate`, `interp`, or `cgen` —
`corevalidate` runs its own traversal" rather than copying pathoracle's
"consumed only by tests" claim.

**Typed-error shape** (`pathoracle.go:110-125` for `backEdgeError`;
`pathoracle.go:88-108` for `pathCapError`, which additionally supplies the
`errors.As`-style accessor):

```go
// backEdgeError is returned when the declared successor relation contains a
// cycle. OWN-03 is scoped to acyclic CFGs this phase (T-03-11); this
// package independently re-derives that rejection rather than trusting
// check.go's own loanLivenessFixpoint to have caught it first, so a
// corrupted or synthetic CFG that check.go never saw is still refused here.
type backEdgeError struct {
	functionID, blockID string
}

func (e *backEdgeError) Error() string {
	return fmt.Sprintf("pathoracle.cfg_back_edge: function %q block %q participates in a cycle", e.functionID, e.blockID)
}

func (e *backEdgeError) Code() string { return "pathoracle.cfg_back_edge" }

// PathCapError type-asserts err as the path-count-cap rejection, mirroring
// the errors.As convention used elsewhere in the repository (e.g.
// session.go's EngineMismatch).
func PathCapError(err error) (*pathCapError, bool) {
	e, ok := err.(*pathCapError)
	return e, ok
}
```

Copy shape: private struct + `Error() string` + `Code() string` returning
`"core.call_graph_cycle"` (D-07-15) + an exported `CycleError(err error)
(*cycleError, bool)` accessor. The struct must carry the witness path
(`members []string`) so the diagnostic's ordered `Causes` can be built without
re-deriving.

**Bounded-ceiling constant pattern** (`pathoracle.go:38-51`) — the shape to copy
for any Phase 07 bound (the 32-member diagnostic cap): a named exported const
whose doc comment states the *reachable* maximum today, why the constant sits
above it, and that it is fail-closed and never raised to make a test pass:

```go
// MaxPaths bounds the number of acyclic entry-to-return paths this package
// will enumerate for one function. Today's language caps a match at
// maxArmsPerMatch (64) arms ... so it never fires on any program the parser
// can produce today, while still being a genuine, finite, fail-closed
// ceiling against a future nested-branching CFG shape ...
const MaxPaths = 4096
```

**Fault-injection seam for the QLT-08 mutation** (`pathoracle.go:53-66`) — this
is the in-tree pattern for making a seeded mutation an *automated* test rather
than a detached-worktree demonstration. Phase 07's "swap the gray check for a
plain visited check" mutation should get exactly this shape:

```go
// TerminatorKindsOverride is a fault-injection seam for
// TestTerminatorWalkMutationKilled (D-09/D-04-29): production always closes
// a path on the full core.TerminatorKinds() set; the test temporarily
// narrows it ... to prove that a walker recognising fewer terminators really
// does misclassify a fixture path. nil (the always-true production default)
// means "use core.TerminatorKinds() unmodified".
var TerminatorKindsOverride func() []core.OperationKind

func recognizedTerminatorKinds() []core.OperationKind {
	if TerminatorKindsOverride != nil {
		return TerminatorKindsOverride()
	}
	return core.TerminatorKinds()
}
```

**Public entry-point doc style** (`pathoracle.go:343-354`) — one entry point,
doc naming the return triple and every error condition:

```go
// RecomputeEndpoints is the oracle's single entry point: independently
// decide loan endpoints for a checked function from the typed core alone.
// ... Returns the recomputed endpoints, the oracle's own counted work (one
// unit per operation inspected across every enumerated path), and an error
// if the function's CFG is cyclic, exceeds MaxPaths, or a linearized path
// fails the late terminal guard.
func RecomputeEndpoints(function core.Function) ([]core.LoanEndpoint, int, error) {
```

D-07-18's entry point should read
`func Order(program core.Program) (reversePostorder []string, err error)`.

---

### `internal/compiler/callgraph/callgraph_test.go` (NEW, test)

**Analog:** `internal/compiler/pathoracle/pathoracle_test.go`

**Static import-independence test** (`pathoracle_test.go:46-76`) — this is the
exact test D-07-24's "corevalidate must not import originvalidate" backing test
copies. Note the sibling names in the doc comment; Phase 07 adds a fourth to that
family:

```go
// TestOracleImportsStayIndependent reads pathoracle's own Go import list
// (parsed from source, never assumed) and fails if it imports check,
// corevalidate, or ast — the T-03-13 import-independence falsifier, in the
// style of corevalidate's own TestValidatorImportsStayIndependent and
// originvalidate's TestOriginValidateImportsNeitherCheckNorAst.
func TestOracleImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "pathoracle")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			forbidden := []string{"/compiler/check", "/compiler/corevalidate", "/compiler/ast", "/compiler/interp", "/compiler/cgen"}
			for _, bad := range forbidden {
				if strings.HasSuffix(path, bad) {
					t.Fatalf("%s imports %s, which pathoracle must never depend on", entry.Name(), path)
				}
			}
		}
	}
}
```

Sibling instances to keep in lockstep:
`internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go:74-100`
(`TestValidatorImportsStayIndependent`),
`internal/compiler/corevalidate/corevalidate_exclusive_test.go:205-215`
(`TestAttributeValidatorImportsStayIndependent`),
`internal/compiler/originvalidate/originvalidate_test.go:20-40`.
Phase 07's version must add `/compiler/originvalidate` and `/compiler/callgraph`
to `corevalidate`'s forbidden list (D-07-24's backing test).

**Synthetic `core.Program` builder helper** (`pathoracle_test.go:188-203`) — the
hand-built-artifact template D-07-19 requires; note it never touches
`syntax`/`check`:

```go
func failOnlyTerminatedFunction(terminatorOpID string, terminatorKind core.OperationKind) core.Function {
	return core.Function{
		ID: "s1:fn:fail_only", EntryPointID: "s1:fn:fail_only:point:entry",
		Parameter: core.Parameter{ID: "s1:fn:fail_only:place:0", Name: "value", Type: "Buffer"},
		Linear: &core.LinearBody{
			ID: "s1:fn:fail_only:linear",
			Operations: []core.LinearOperation{
				{ID: "op:0", Kind: core.OpBorrowShared, SourceID: "s1:fn:fail_only:place:0", TargetID: "place:1", LoanID: "loan:0"},
				{ID: terminatorOpID, Kind: terminatorKind, SourceID: "place:1"},
			},
			Blocks: []core.Block{
				{ID: "block:only", PointID: "s1:fn:fail_only:point:entry", OperationIDs: []string{"op:0", terminatorOpID}, Successors: nil},
			},
		},
	}
}
```

Phase 07's analog: `syntheticProgram(edges map[string][]string) core.Program`,
producing one `core.Function` per node with one `core.OpCall` operation per
outgoing edge. Note the ID convention (`"{functionID}:place:N"`,
`"{functionID}:point:entry"`, `"block:..."`) — reuse it.

**Fixture-loading helper** (`pathoracle_test.go:21-44`) — the parser-reachable
side, for the `check`-level integration tests:

```go
func readFixture(t *testing.T, name string) []byte {
	t.Helper()
	source, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase3", name))
	...
}

func checkedFunction(t *testing.T, fixture string) core.Function {
	t.Helper()
	source := readFixture(t, fixture)
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) != 0 {
		t.Fatalf("fixture %q failed to parse: %+v", fixture, parsed.Diagnostics)
	}
	result := check.Program(parsed.Program)
	...
}
```

**Mutation-kill test** — the QLT-08 naming convention and body shape
(`pathoracle_test.go:268-295`). This is the real per-control mutation-kill
precedent (research is correct that `session/qlt01_registry.json` is a distinct,
non-reusable M001 mechanism; no `qlt08` registry exists). Naming convention in
the tree: `Test<Subject>MutationKilled` for a single seeded fault,
`Test<Subject>MutationMatrix` for a table:

```go
// TestTerminatorWalkMutationKilled is D-09's automated mutation-kill
// falsifier for D-04-29 in pathoracle: narrowing the recognised terminator
// set (deleting core.OpFail, the exact mutation the throwaway-detached-
// worktree demonstration performs on the source) must make the oracle
// misclassify the fail-only fixture as an unterminated loan -- proving the
// widening actually bites.
func TestTerminatorWalkMutationKilled(t *testing.T) {
	function := failOnlyTerminatedFunction("op:1", core.OpFail)
	if _, _, err := pathoracle.RecomputeEndpoints(function); err != nil {
		t.Fatalf("unexpected error before mutation: %v", err)
	}

	original := pathoracle.TerminatorKindsOverride
	pathoracle.TerminatorKindsOverride = func() []core.OperationKind {
		return []core.OperationKind{core.OpReturn, core.OpDefect} // OpFail deleted
	}
	defer func() { pathoracle.TerminatorKindsOverride = original }()

	_, _, err := pathoracle.RecomputeEndpoints(function)
	if err == nil {
		t.Fatalf("mutation (deleting OpFail) had no observable effect: expected an unterminated-loan rejection")
	}
	type coded interface{ Code() string }
	code, ok := err.(coded)
	if !ok || code.Code() != "pathoracle.unterminated_loan" {
		t.Fatalf("want pathoracle.unterminated_loan after mutation, got: %v", err)
	}
}
```

Copy the four-beat structure exactly: (1) assert clean before mutation,
(2) save/override/`defer` restore the seam, (3) assert the mutation has an
**observable** effect, (4) assert the *specific* code. Other matrix-shaped
siblings for D-07-24's three-fault matrix:
`internal/compiler/corevalidate/corevalidate_test.go:21` (`TestOwnershipMutationMatrix`),
`:875` (`TestReleaseOrderMutationMatrix`),
`internal/compiler/session/session_phase5_alias_test.go:162`
(`TestAliasMutationWithoutDivergenceFailsTheLane` — **this is the exact analog for
D-07-24 fault 3, the bilateral fault that must make the gate FAIL**).

---

### `internal/compiler/core/core.go` — `lang.interface/1` (model, data-shape)

**Analog A — the versioned-const pairing** (`core.go:5-8`, mirrored at
`diagnostic.go:12-13` and `evidence.go:31`):

```go
const (
	Schema  = "lang.core/0"
	Schema1 = "lang.core/1"
)
```

Stage 0's edit follows this exactly — keep the existing single-const line
(`core.go:157`) and add the sibling:

```go
// InterfaceSchema versions the Interface artifact independently of the core
// schema it summarizes: adding a field here never moves a core.Program byte.
const InterfaceSchema = "lang.interface/0"
```

**Analog B — the additive-field doc style** (`core.go:270-301`). This is the
in-tree precedent D-07-08 argues *against* following for the required fields,
but it is still the precedent for the doc-comment discipline: every added field
names its plan, its decision ID, exactly which kind populates it, and states
what "absent" means so old serialized bytes are provably unchanged:

```go
	// OkEdgeID, ErrEdgeID, and ErrTargetID are Phase 4 additive omitempty
	// facts populated only on an OpForeignCall operation (D-04-04): the ok
	// edge continues at TargetID (an ordinary place, exactly like OpCopy's
	// target), while the err edge's own synthesized failure-ADT place is
	// ErrTargetID. Every pre-Phase-4 operation, and every operation kind
	// other than OpForeignCall, leaves all three empty.
	OkEdgeID    string `json:"ok_edge_id,omitempty"`
	ErrEdgeID   string `json:"err_edge_id,omitempty"`
	ErrTargetID string `json:"err_target_id,omitempty"`
	// ReleasesOperationID is Phase 4 plan 02's additive omitempty fact
	// (D-04-07): populated only on an OpRelease operation, it names the
	// OpForeignCall operation ID this release discharges ... Every
	// pre-plan-02 operation, and every operation kind other than OpRelease,
	// leaves this empty.
	ReleasesOperationID string `json:"releases_operation_id,omitempty"`
	// Allocator is Phase 4 plan 03's additive omitempty fact
	// (T-04-14/allocator-identity requirement): populated on an OpForeignCall
	// ... A release whose Allocator differs from its own acquisition's is
	// refused independently by check (at emission time) and by corevalidate
	// (by re-fetching the acquisition and comparing, never trusting check's
	// bookkeeping). ...
	Allocator string `json:"allocator,omitempty"`
```

Phase 07's `/1` doc comments must invert the last clause: **no `omitempty` on
any required field**, and each doc must say "absence is an error, refused by
both producer and peer" (D-07-09). D-07-13's integrity-not-authenticity note
belongs on `ClosureDigest`'s doc comment, in this same style.

**Analog C — the "structurally cannot reach a body" argument to preserve**
(`core.go:148-171`). The `/1` type must carry this claim forward verbatim in
spirit; do not drop it during the widening:

```go
// Interface is the Phase 3 OWN-04 body-stripped separate-compilation
// artifact (03-RESEARCH Q6): a subset of Program containing only module
// identity, a digest binding it to the exact core.Program it was derived
// from, and per-function signatures — explicitly no Linear or Match body.
// A consuming process decodes only this shape and can never reconstruct a
// provider body from it.
type Interface struct {
	Schema     string              `json:"schema"`
	ModuleID   string              `json:"module_id"`
	CoreDigest string              `json:"core_digest"`
	Functions  []FunctionSignature `json:"functions"`
}

// FunctionSignature is one function's body-stripped public surface ...
// It deliberately has no Linear/Match field at all — not merely an
// omitted one — so a consumer decoding this type structurally cannot reach a
// body even by accident.
type FunctionSignature struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Parameter    Parameter     `json:"parameter"`
	ReturnType   string        `json:"return_type"`
	PublicOrigin *PublicOrigin `json:"public_origin,omitempty"`
	Abilities    []Ability     `json:"abilities"`
}
```

**Analog D — `ForeignReach`'s three string fields** (planner discretion item;
the in-tree precedent for plain strings rather than an enumerated Go type is
`core.go:60-64`):

```go
type ForeignContract struct {
	Symbol       string `json:"symbol"`
	Allocator    string `json:"allocator"`
	Unwind       string `json:"unwind,omitempty"`
	NonlocalExit string `json:"nonlocal_exit,omitempty"`
	Fails        string `json:"fails,omitempty"`
```

Note `Fails` here is the vocabulary `FunctionSignature.Fails` reuses verbatim.

---

### `internal/compiler/core/core.go` — registering `core.OpCall` (config/registry)

**Analog:** `core.go:212-233` (the three Phase 4 kind constants) and
`core.go:244-252` (the registry).

```go
	// OpForeignCall is Phase 4's only call surface (D-04-01): a fallible call
	// to a declared `foreign C {}` symbol. It produces two successor edges
	// (OkEdgeID/ErrEdgeID below) rather than a single target place.
	OpForeignCall OperationKind = "foreign_call"
	// OpRelease is Phase 4 plan 02's resource-lifecycle operation (D-04-07):
	// a non-terminal transition discharging exactly one completed
	// OpForeignCall acquisition. It is never a terminator ...
	OpRelease OperationKind = "release"
```

```go
// AllOperationKinds returns every declared OperationKind, in declaration
// order. This is the single table every dispatch site (check, corevalidate,
// interp, cgen, pathoracle, originvalidate) is tested against (D-04-22): a
// constant added to the block above without also being added to this literal
// slice is exactly the defect this registry exists to catch --
// TestAllOperationKindsRegistered fails the moment the two counts diverge.
func AllOperationKinds() []OperationKind {
	return []OperationKind{OpCopy, OpMove, OpBorrowShared, OpBorrowExclusive, OpReturn, OpForeignCall, OpFail, OpRelease, OpDefect}
}
```

`OpCall` is **not** a terminator — do not touch `TerminatorKinds()`
(`core.go:256-258`). `OpCall`'s doc comment should follow `OpRelease`'s
"it is never a terminator" phrasing and state that it carries a single
`TargetID` and a callee identity, unlike `OpForeignCall`'s two edges.

---

### `internal/compiler/originvalidate/originvalidate.go` — widened `BuildInterface` (producer, transform)

**Analog:** itself, `originvalidate.go:408-434`. The exact code Stage 0 edits,
including the `:427` silent empty-abilities fallback D-07-23 names:

```go
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// BuildInterface strips every function body from program and binds the
// resulting summary to program's own content digest, reusing the same
// SHA-256 content-digest pattern already proven for evidence.CoreDigest
// (Spike 004's certificate binds the same way). Callers MUST run
// ValidatePublished against program first: BuildInterface packages what the
// producer already proved rather than re-deriving it.
func BuildInterface(program core.Program) (core.Interface, error) {
	coreBytes, err := json.Marshal(program)
	if err != nil {
		return core.Interface{}, err
	}
	summary := core.Interface{
		Schema: core.InterfaceSchema, ModuleID: program.ModuleID, CoreDigest: digest(coreBytes),
		Functions: make([]core.FunctionSignature, 0, len(program.Functions)),
	}
	for _, function := range program.Functions {
		abilities := []core.Ability{}
		if function.Linear != nil {
			for _, fact := range function.Linear.Types {
				if fact.ID == function.ID+":type:0" {   // <-- :427, the exact-ID lookup
					abilities = fact.Abilities
					break
				}
			}
		}                                                // <-- no else: silent empty-abilities fallback
		summary.Functions = append(summary.Functions, core.FunctionSignature{
			ID: function.ID, Name: function.Name, Parameter: function.Parameter, ReturnType: function.ReturnType,
			PublicOrigin: function.PublicOrigin, Abilities: abilities,
		})
	}
	return summary, nil
}
```

Notes for the planner:
- `digest()` (`:405-408`) is the existing `sha256:`-prefixed content-digest
  helper `ClosureDigest` must reuse — do **not** mint a second digest format.
- The `Schema: core.InterfaceSchema` literal at `:421` is one of the ~3 real
  `/1`-bump sites (A-04 confirmed `debugmap.go:54` is a comment).
- D-07-24 fault 1 seeds a fault *here*: force the `fact.ID == ...` comparison to
  never match. The seam should follow `pathoracle.TerminatorKindsOverride`'s
  package-level `var ...Override func(...)` shape above.

---

### `internal/compiler/corevalidate/corevalidate.go` — `case core.OpCall:` at BOTH replay sites

**Analog:** the two `OpForeignCall` cases, which are the D-07-21 proof that
these switches are deliberately duplicated rather than shared.

`replayStraightLine`, `corevalidate.go:960-972`:

```go
		case core.OpForeignCall:
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
			errTarget, errKnown := places[operation.ErrTargetID]
			if !v.check(errKnown && operation.ErrTargetID != operation.TargetID && operation.ErrTargetID != operation.SourceID && !produced[operation.ErrTargetID] && errTarget.TypeID != "", "core.invalid_target", operation.ErrTargetID) {
				return false
			}
			initialized[operation.ErrTargetID] = true
			produced[operation.ErrTargetID] = true
```

`replayBlocks`, `corevalidate.go:1157-1170` — byte-identical body for this kind,
confirming the "every new case lands twice, per engine" rule.

**The simpler shape `OpCall` should actually copy** — single target, no edges
(`corevalidate.go:905-912`, `OpCopy`):

```go
		case core.OpCopy:
			if !v.check(hasAbility(types[operation.TypeID], core.AbilityCopy), "core.ability.copy_denied", operation.TypeID) {
				return false
			}
			if !v.targetMatches(function, index, operation, places, produced) {
				return false
			}
			initialized[operation.TargetID] = true
			produced[operation.TargetID] = true
```

**Fail-closed `default:` arm** — present identically at `:1000-1002` and
`:1214-1216`; do not touch, it is what makes the registry edit break loudly:

```go
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
```

**The "documented but structurally unreachable, named so the six-site table
finds it" comment style** (`corevalidate.go:975-981`, `OpDefect` in
`replayStraightLine`) — reuse this exact justification style anywhere Phase 07
adds a case it cannot yet exercise:

```go
		case core.OpDefect:
			// A terminator alongside OpReturn/OpFail (D-04-15): never the
			// last operation of anything but its own straight-line body (this
			// case exists only so control:kind.exhaustive_dispatch's table
			// finds it handled here too -- a straight-line, block-less body
			// can never actually carry a defect this phase).
```

---

### `internal/compiler/corevalidate/corevalidate.go` — the independent peer(s)

**Analog:** `corevalidate.go:352-361` — the exact in-tree template for a peer
re-derivation whose doc comment names the *materially different mechanism*.
Copy this comment structure for both the Stage 0 signature-summary peer and the
Stage 2 traversal peer:

```go
			// D-04-02, independently derived: check.go refuses a callee that
			// resolves to a Lang function name at parse-resolution time (an
			// AST-level, pre-core fact this validator never sees). This
			// validator re-derives the same refusal by a materially
			// different mechanism -- a straight name collision scan against
			// every OTHER function this core.Program itself declares --
			// reading only the core artifact, never check's own name table.
			if !v.check(!isDeclaredFunctionName(v.program.Functions, function.ID, function.ForeignContract.Symbol), "core.call_target_not_foreign", operation.ID) {
				return false
			}
```

**Node-identity precedence rule** (`corevalidate.go:1654-1664`) — the function
the foreign-symbol-shadowing fixture pins; `callgraph` must reuse this exact
precedence rather than restating it:

```go
func isDeclaredFunctionName(functions []core.Function, excludeFunctionID, name string) bool {
	for _, candidate := range functions {
		if candidate.ID == excludeFunctionID {
			continue
		}
		if candidate.Name == name {
			return true
		}
	}
	return false
}
```

**Independent-rederivation-with-diagnostic pattern** (`check.go:1680-1703`,
`releaseAllocatorMismatch`) — two ordered causes plus a repair; the shape for
comparing a producer value to a peer value:

```go
		causes := []diagnostic.Cause{
			{Kind: "release", Detail: operation.ID},
			{Kind: "acquisition", Detail: acquisition.ID},
		}
		problem := diagnostic.ErrorWithRepairs(
			"foreign.release_allocator_mismatch", diagnostic.Span{},
			"a release's declared allocator differs from its acquisition's", causes,
			diagnostic.Repair{Kind: "match_acquisition_allocator"},
		)
```

---

### `internal/compiler/syntax/parser.go` + `internal/compiler/ast/ast.go` — the new `"call"` RHS kind

**Analog A — the code being relocated** (`parser.go:404-421`, inside
`linearBody`, which starts at `:383`):

```go
		source := p.identifier("syntax.expected_binding_source")
		if p.peek().Kind == TokenLParen {
			// D-04-06: a fallible foreign call is admissible only as the
			// operand of `try` (or `discard ... because`, not yet a parsed
			// construct this phase). A bare call in a binding right-hand
			// side is refused here, at parse time, so the core IR never has
			// to encode a fallible operation without a failure successor --
			// the parser diagnostic alone short-circuits before check.go
			// (and therefore corevalidate/interp/cgen) ever run.
			p.problem("syntax.fallible_call_not_consumed", source, "a fallible call must be the operand of `try`")
			_, end := p.callArguments()
			body.Bindings = append(body.Bindings, ast.Binding{
				Name: name.Text, RHS: ast.RHS{Kind: "call_unconsumed", Source: source.Text, Span: source.Span}, Span: spanFrom(bindingStart, source),
			})
			body.Span.End = end
			continue
		}
```

Phase 07 replaces the `p.problem(...)` call with a `Kind: "call"` binding that
carries `Callee` and `Arguments`, and relocates the refusal to `check`. Note the
existing `Kind: "call_unconsumed"` marker already flows to `check` — the
relocated refusal has a pre-existing carrier.

**Analog B — the exact constructor to model `"call"` on** (`parser.go:456-465`,
`tryCallBinding`):

```go
// tryCallBinding parses `try <callee>(<arg>, ...)` following a binding's `=`
// (D-04-06's sole admissible fallible-call consumer this phase).
func (p *parser) tryCallBinding(bindingStart, name Token) (ast.Binding, int) {
	tryToken := p.advance()
	callee := p.identifier("syntax.expected_foreign_callee")
	arguments, end := p.callArguments()
	rhs := ast.RHS{Kind: "try_call", Callee: callee.Text, Arguments: arguments, Span: spanFrom(tryToken, callee)}
	return ast.Binding{Name: name.Text, RHS: rhs, Span: diagnostic.Span{Start: bindingStart.Span.Start, End: end}}, end
}
```

`p.callArguments()` already exists and returns `([]string, int)` — reuse it, do
not write a second argument parser. D-07-01's arity-1 rule is a **checker**
predicate; the parser accepts what `callArguments` yields.

**Analog C — the `ast.RHS` doc-comment convention for kind-conditional fields**
(`ast.go:143-158`) — the exact place a new `"call"` kind is documented:

```go
type RHS struct {
	Kind   string
	Source string
	Span   diagnostic.Span
	// Callee and Arguments are populated when Kind == "try_call" or
	// Kind == "discard_call" (D-04-06): a fallible foreign call, admissible
	// only as the operand of `try` or `discard ... because`. Callee names
	// the foreign symbol; Arguments is its argument place names in source
	// order.
	Callee    string
	Arguments []string
	// Rationale is populated only when Kind == "discard_call": the required
	// non-empty string literal explaining why the call's failure is not
	// actionable here (D-04-06). Empty for every other Kind.
	Rationale string
}
```

Edit shape: extend the `Callee`/`Arguments` comment to `"call"` and record that
`"call"` never sets `Rationale`. There is a downstream kind-enumeration guard to
update at `check.go:1302-1309`:

```go
func everyBindingIsFallible(bindings []ast.Binding) bool {
	for _, binding := range bindings {
		if binding.RHS.Kind != "try_call" && binding.RHS.Kind != "discard_call" {
			return false
		}
	}
	return true
}
```

---

### `internal/compiler/check/check.go` — Lang-call admission + callgraph invocation

**Analog (shape only, NOT the code path):** `check.go:1312-1350`,
`resolveForeignStep`. Per RESEARCH Anti-Patterns and Open Question 2, this
function stays foreign-only; the new predicate is new code with the same shape:

```go
// resolveForeignStep independently resolves one step's declared foreign
// symbol and its failure-ADT type fact, shared by both checkForeignTracer and
// checkResourceLifecycle so the two admission gates (D-04-02/D-04-16) and the
// argument-shape rule stay in exact lockstep ...
func resolveForeignStep(binding ast.Binding, functionParameterName string, foreignSymbols map[string]foreignSymbolInfo, functionNames map[string]bool, dataTypes map[string]core.DataType) (foreignSymbolInfo, core.TypeRef, ability.Result, *diagnostic.Diagnostic) {
	if len(binding.RHS.Arguments) != maxForeignParametersPerSymbol || binding.RHS.Arguments[0] != functionParameterName {
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call argument must be the function's own parameter")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
	symbol, isForeign := foreignSymbols[binding.RHS.Callee]
	if !isForeign {
		if functionNames[binding.RHS.Callee] {
			causes := []diagnostic.Cause{{Kind: "callee", Detail: binding.RHS.Callee}}
			problem := diagnostic.ErrorWithRepairs(
				"core.call_target_not_foreign", binding.RHS.Span,
				"a fallible call's target must be a declared foreign symbol, not a Lang function", causes,
				diagnostic.Repair{Kind: "declare_foreign_symbol"},
			)
			return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
		}
		problem := diagnostic.Error("name.unknown", binding.RHS.Span, "foreign call target is unknown")
		return foreignSymbolInfo{}, core.TypeRef{}, ability.Result{}, &problem
	}
```

Copy the three-part structure exactly: (1) argument-shape predicate first,
(2) `foreignSymbols` lookup, (3) the `functionNames` branch producing a *specific*
diagnostic before the generic `name.unknown` fallback. Phase 07's new predicate
inverts branch 3 (a Lang name is now the *accept* case) and D-07-01 widens
predicate 1 from `== functionParameterName` to "resolves in `places`".
The `*diagnostic.Diagnostic` return convention (nil means OK) is the local idiom.

**The scope map D-07-01's rule consults** (`check.go:1822-1825`, inside
`analyzeStraightLine`; `analyzeArmBody` starts at `:784` and carries the same
map):

```go
	places := map[string]*placeState{
		parameterName: {place: result.Places[0], declared: parameterSpan, initialized: true},
	}
```

**Multi-cause + repair diagnostic construction** (`check.go:1710-1728`,
`missingForeignPolicyDiagnostic`) — the template for the relocated
`syntax.fallible_call_not_consumed` refusal and for SEM-06's callable refusal:

```go
	causes := []diagnostic.Cause{
		{Kind: "foreign_symbol", Detail: symbol.Name},
		{Kind: "missing_policy", Detail: missing},
	}
	problem := diagnostic.ErrorWithRepairs(
		"foreign.unwind_policy_undeclared", symbol.Span,
		"a foreign symbol must declare both an unwind and a nonlocal_exit policy, with no default", causes,
		diagnostic.Repair{Kind: "declare_unwind_policy"},
	)
	return &problem
```

**For `core.call_graph_cycle` use `diagnostic.Error`, not `ErrorWithRepairs`** —
D-07-15 declares the cycle explicitly non-repairable, and `Error` is the
variadic-causes constructor (`diagnostic.go:108-117`).

---

### `internal/compiler/interp/interp.go` — `case core.OpCall:` (documented-unreachable)

**Analog:** `interp.go:97-107` — verbatim the precedent CONTEXT.md cites:

```go
		case core.OpForeignCall:
			// No match arm can produce a foreign call this phase (checkBranch
			// does not admit `try` inside an arm body) -- this case exists
			// solely so control:kind.exhaustive_dispatch's six-site table
			// finds every kind handled at every site, per D-04-22.
			values[operation.TargetID] = value
			values[operation.ErrTargetID] = "err"
			events = append(events, Event{
				Schema: execution.Schema1, ID: operation.ID + ":event", Kind: "foreign.called", FunctionID: function.ID,
				SourcePlace: operation.SourceID, TargetPlace: operation.TargetID, TypeID: operation.TypeID,
			})
```

And the terser sibling at `interp.go:120-123`:

```go
		case core.OpRelease:
			// No match arm can produce a release this phase either -- named
			// here for the same six-site exhaustive-dispatch reason as
			// OpForeignCall above.
			events = append(events, ownedEvent(function, operation, "resource.released"))
```

Also note `ownedEvent(function, operation, "<verb>")` (used at `:83`, `:88`, etc.)
as the event-emission helper — `OpCall`'s interp case should emit through it, not
hand-build an `Event` unless it needs the two-place shape.
Sibling switches to update: `runLinearBlocks` and `runLinear` (`interp.go:221-318`,
`:409-428`).

---

### `internal/compiler/cgen/cgen.go` — `case core.OpCall:` (structurally unreachable)

**Analog:** `cgen.go:1723-1729` — the "known-but-unsupported, not unknown" arm:

```go
		case core.OpForeignCall, core.OpFail:
			// checkBranch never emits either kind inside a match arm body
			// this phase (no `try` support inside an arm) -- named here,
			// rather than falling into default:, so control:kind.exhaustive
			// _dispatch's six-site table finds a known-but-unsupported case
			// rather than an unknown one (D-04-22).
			return fmt.Errorf("operation %q: foreign calls inside a match arm body are not supported this phase", operation.ID)
		default:
			return fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
```

Per A-02, `OpCall`'s cgen arm should return this same "not supported this phase"
error and cite the `len(Functions) != 1` reachability argument
(`cgen.go:22,52`) in its comment. Four switch sites (verify live line numbers
before editing; research reports `:260, :467, :652, :1652` vs CONTEXT's
`:259, :466, :651`).

---

### `internal/compiler/session/session.go` — phase-07 lane + `/1` export

**Analog A — the CLI-observable dispatch lane** (`session.go:2501-2573`). Per
A-05 this is a *literal, phase-scoped* block; Phase 07 adds its own copy:

```go
	// Lane: control:kind.exhaustive_dispatch (D-04-22, task 04-07-01). This is
	// the session-layer, CLI-observable sibling of core_test.go's
	// TestAllOperationKindsHandledAtEverySite ... scoped to the four operation
	// kinds THIS phase introduced (OpForeignCall, OpRelease, OpFail, OpDefect)
	// -- the kinds Phase 1-3's own corpora cannot exercise ...
	laneStarted = time.Now()
	dispatchFixtures := []string{"foreign_acquire_one.lang", "acquire_three_success.lang", "defect_terminal.lang", "nonlocal_exit_probe.lang"}
	encounteredKinds := make(map[core.OperationKind]bool)
	dispatchWork := 0
	for _, fixtureName := range dispatchFixtures {
		fixtureSource, fixtureErr := readBoundedFile(filepath.Join(corpus, fixtureName), syntax.MaxSourceBytes)
		if fixtureErr != nil {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, 0, laneStarted)
			return fail(protocol.StatusOperational, "verify.fixture_missing", fixtureName)
		}
		fixtureChecked := Check(fixtureSource)
		...
		fixtureValidated := corevalidate.Validate(fixtureChecked.Program)
		...
		dispatchProgram := fixtureValidated.Program()
		for _, function := range dispatchProgram.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encounteredKinds[operation.Kind] = true
				}
				if function.Linear.ID != "" {
					if _, _, oracleErr := pathoracle.RecomputeEndpoints(function); oracleErr != nil { ... }
				}
			}
			_ = originvalidate.RecomputeOriginPerReturn(function)
			dispatchWork++
			...
		}
		if len(dispatchProgram.Functions) == 1 {
			if _, cgenErr := cgen.Emit(dispatchProgram); cgenErr != nil { ... }
		}
	}
	for _, kind := range []core.OperationKind{core.OpForeignCall, core.OpRelease, core.OpFail, core.OpDefect} {
		if !encounteredKinds[kind] {
			addLane("lane:kind-exhaustive-dispatch", "fail", nil, dispatchWork+1, 0, laneStarted)
			return fail(protocol.StatusInvalid, "verify.control_missing", "control:kind.exhaustive_dispatch (kind "+string(kind)+" never encountered)")
		}
	}
	addLane("lane:kind-exhaustive-dispatch", "pass", []string{"control:kind.exhaustive_dispatch"}, dispatchWork+1, 0, laneStarted)
```

Note `if len(dispatchProgram.Functions) == 1` at `:2559` — the gate confirming
`cgen` never runs on an `OpCall`-bearing program. Also note the lane-name /
control-name pairing and the `addLane(name, "pass", []string{"control:..."}, ...)`
convention — Phase 07's lane needs its own distinct lane name (the Phase 4 lane
must not be repurposed) plus registration in a `Phase7RequiredControls()`
mirroring `Phase4RequiredControls()` (`session.go:2576`) and
`session_phase5.go:29-55`'s control-name list.

**Analog B — the `/1` writer** (`session.go:1023-1064`, `InterfaceExportCommandFile`
— the real name; `RunInterfaceCommandFile` does not exist, per A-04):

```go
// InterfaceExportCommandFile is the producer side of OWN-04's separate
// compilation demonstration: it checks sourcePath, independently recomputes
// every declared PublicOrigin fact from the typed core alone
// (originvalidate.ValidatePublished — T-03-02/T-03-16), and only on success
// writes a body-stripped, digest-bound core.Interface summary to outPath.
func InterfaceExportCommandFile(sourcePath, outPath string) (protocol.Result, error) {
	...
	summary, err := originvalidate.BuildInterface(checked.Program)
	...
	summaryBytes, err := json.Marshal(summary)
	...
	if err := os.WriteFile(outPath, append(summaryBytes, '\n'), 0o600); err != nil { ... }
	result.Interface = interfaceProjection(summary.Schema, summary.ModuleID, summary.CoreDigest, summary.Functions)
	result.ExpectedEscapes = originvalidate.ExpectedEscapes()
	return completeCommand(result, started, len(summary.Functions)), nil
}
```

`interfaceProjection(...)` feeds `protocol.InterfaceSummary` /
`protocol.InterfaceFunctionAnswer` (`protocol.go:132-150`) — the "never from a
body field" consumer that must learn `/1` fields:

```go
// InterfaceFunctionAnswer is one function's body-blind origin answer, read
// directly from a core.Interface summary — never from a body field.
type InterfaceFunctionAnswer struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Paths  []string `json:"paths,omitempty"`
	Access string   `json:"access,omitempty"`
}
```

---

### `internal/compiler/core/core_test.go` — the in-process dispatch control

**Analog:** itself, `core_test.go:266-357`. The `fixtures` slice at `:272-284` is
the cumulative, hand-maintained list Phase 07 appends its two-function fixture to:

```go
func TestAllOperationKindsHandledAtEverySite(t *testing.T) {
	fixtures := []string{
		"testdata/phase1/toggle.lang",
		...
		"testdata/phase4/foreign_acquire_one.lang",
		"testdata/phase4/acquire_three_success.lang",
		"testdata/phase4/defect_terminal.lang",
	}
	encountered := make(map[core.OperationKind]bool)
	for _, path := range fixtures {
		source, err := os.ReadFile(testsupport.ProjectPath(splitPath(path)...))
		...
		checked := session.Check(source)
		...
		validated := corevalidate.Validate(program)
		...
		for _, function := range program.Functions {
			if function.Linear != nil {
				for _, operation := range function.Linear.Operations {
					encountered[operation.Kind] = true
				}
				if function.Linear.ID != "" {
					if _, _, err := pathoracle.RecomputeEndpoints(function); err != nil { ... }
				}
			}
			_ = originvalidate.RecomputeOriginPerReturn(function)
			...
		}
		// cgen site: Emit requires exactly one function.
		if len(program.Functions) == 1 {
			if _, err := cgen.Emit(program); err != nil { ... }
		}
	}
	for _, kind := range core.AllOperationKinds() {
		if !encountered[kind] {
			t.Fatalf("operation kind %q is never exercised by any corpus fixture in this control", kind)
		}
	}
}
```

The final loop is what makes registering `OpCall` fail loudly until a real
fixture produces it — this is the "impossible to forget" forcing function.
Note `linearProbeInput(function)` (`core_test.go:~245-263`) supplies interp
input per fixture; a two-function fixture likely needs a new arm there.

---

### Frozen-`/0` fixture discipline (Stage 0's `/1` bump)

**Analog A — the pinned-ID frozen-document test** (`protocol_test.go:169-192`).
This is the precedent D-07-08 requires Stage 0 to repeat for
`lang.interface/0` — a **pinned literal ID**, with the doc comment stating when
it was computed and what a failure means:

```go
// TestPhase1EvidenceBytesUnchangedAfterBump is Task 3's Behavior Test 4
// (D-06-31): a Result constructed with the frozen protocol.Schema ("/0")
// string, for a fixed deterministic Phase-1-shaped diagnostic, still
// finalizes to the exact same ID it did before 06-06's bump. The pinned
// literal below was computed at this commit against the tree as it stood
// immediately after Tasks 1 and 2 landed; a future change to Finalize()'s
// identity struct that would have perturbed already-published /0 documents
// fails here, proving /0 document bytes are reproducible, not merely
// retired.
func TestPhase1EvidenceBytesUnchangedAfterBump(t *testing.T) {
	result := protocol.Result{ Schema: protocol.Schema, ... }
	const pinnedID = "result:787da433102c86ea05c7a2f1"
	if got := result.Finalize().ID; got != pinnedID {
		t.Fatalf("a /0-schema Result's Finalize().ID moved: got %s, want pinned %s -- a /0 document's bytes are frozen once shipped and this change would have perturbed already-published output", got, pinnedID)
	}
}
```

Sibling: `TestSchemaZeroConstantsStillExist` (`protocol_test.go:160-167`) — the
one-liner proving the `/0` constant is still exported after the bump. Stage 0
needs the exact equivalent for `core.InterfaceSchema`.

**Analog B — the same pattern at diagnostic level** (`diagnostic_test.go:232-253`,
`TestDiagnosticZeroBytesUnchanged`) — note it pins the ID *and* asserts the
`/1`-only field stayed nil on the `/0` path:

```go
	// Pinned literal: the /0 identity struct (Schema, Code, Span, Causes) has
	// no RepairKinds field and must never gain one -- the /0 path is
	// untouched by the /1 Repair extension in this plan.
	const wantID = "diagnostic:d34396fcc8ea6530a1f04dc3"
	if d.ID != wantID {
		t.Fatalf("Error() ID = %q, want pinned %q -- a /0 ID change is a T-06-REPAIR-03 regression", d.ID, wantID)
	}
	if d.Repairs != nil {
		t.Fatalf("Error()-constructed diagnostic must have nil Repairs, got %v", d.Repairs)
	}
```

`InterfaceV0`'s test must do the same: decode a pinned `/0` byte string into the
pinned legacy struct, assert field-for-field equality, and assert the value is
**never admissible for a call**.

**Analog C — the corpus-wide byte pin** (`core_test.go:1-92` table +
`:94-119` test). The `pinnedFixtures` table (path, `CoreSHA256`, `ManifestID`)
is the shape Stage 0's "no `core.Program` byte moved" claim asserts against:

```go
// TestPreviousPhaseCoreBytesUnchanged pins every Phase 1-5 fixture's
// serialized core JSON to its exact byte value from before Phase 6
// (D-06-31/D-06-32, widened from the Phase 5 pin which stopped at Phase 4).
// It must be green before the coordinated lang.command and lang.verify-lane
// bump lands.
func TestPreviousPhaseCoreBytesUnchanged(t *testing.T) {
	for _, fixture := range pinnedFixtures {
		t.Run(fixture.Path, func(t *testing.T) {
			...
			sum := sha256.Sum256(product.CoreBytes)
			got := hex.EncodeToString(sum[:])
			if got != fixture.CoreSHA256 {
				t.Fatalf("core bytes moved for %s: got sha256 %s, want %s", fixture.Path, got, fixture.CoreSHA256)
			}
		})
	}
}
```

Stage 0 must extend this table to Phase 6 fixtures and keep it green — this is
the mechanical proof that a `/1` interface bump moved no `core.Program` byte
(the whole point of `InterfaceSchema`'s independent versioning).

---

### `testdata/phase07/*.lang` (NEW corpus)

**Analogs:** `testdata/phase4/foreign_acquire_one.lang` (header-comment
convention, `module <phase>.<name>`, `export { fn ... }`) and
`testdata/phase3/shared_shared_accept.lang` (the binding grammar D-07-01
reuses unchanged).

```
// Phase 4 tracer (D-04-01/D-04-04/D-04-05/D-04-10): one function declares a
// foreign C symbol, calls it fallibly through `try`, and returns its ok
// value. This proves the whole spine end to end -- checked, independently
// validated, interpreted, lowered to C17, and linked against the
// byte-frozen foreign translation unit -- identically at -O0 and -O3.

module phase4.foreign_acquire_one

export {
  fn main
}

foreign C {

  fn lang_res_open(request: Byte) -> Byte {
    unwind: forbidden
    nonlocal_exit: forbidden
    allocator: "libc_malloc"
    fails: AcquireError
  }
}

data AcquireError =
  | OpenFailed

fn main(request: Byte) -> Byte {
  let handle = try lang_res_open(request)
  handle
}
```

```
// Two overlapping shared loans on one owner are never a conflict: multiple
// simultaneous readers observe without mutation.

module owned.shared_shared_accept

export {
  fn relay
}

fn relay(buffer: Buffer) -> Buffer {
  let first = borrow buffer
  let second = borrow buffer
  let reviewed_first = borrow first
  let reviewed_second = borrow second
  let delivered = take buffer
  delivered
}
```

Fixture conventions to copy: a leading `//` comment naming the decision IDs and
what the fixture proves; `module <namespace>.<filename>`; an `export { fn ... }`
block; a goldens sibling named `<fixture>.golden.c` / `evidence.golden.json` only
where a phase actually pins output (`testdata/phase1/generated.golden.c`,
`testdata/phase2/owned_transfer.golden.c`, `testdata/phase4/foreign_layout_mismatch.golden.c`,
`testdata/phase5/restrict_borrow.golden.c`). **No golden `.c` for Phase 07** —
`cgen` never runs on a multi-function program.

**Important:** `shared_shared_accept.lang` already uses the name `relay` for a
*single-function* fixture in namespace `owned.`. Phase 07's two-function
`relay`/`escort` witness must use a distinct module namespace (`phase07.`) to
avoid confusion; the existing `relay` is not the D-04-03 witness.

---

## Shared Patterns

### Independent re-derivation (apply to: `callgraph`, `corevalidate` peer, `check`)
**Source:** `internal/compiler/pathoracle/pathoracle.go:110-118`,
`internal/compiler/corevalidate/corevalidate.go:352-361`
The doc comment must name (a) what the other derivation does, (b) the
*materially different mechanism* this one uses, (c) what input space it reads
that the other cannot. Never a shared helper. Never an import of the peer.

### Static import-independence falsifier (apply to: `callgraph`, `corevalidate`)
**Source:** `internal/compiler/pathoracle/pathoracle_test.go:46-76`
Parse the package's own `.go` files with `parser.ImportsOnly` and fail on any
forbidden suffix. Phase 07 adds `/compiler/originvalidate` to `corevalidate`'s
list (D-07-24) and creates `callgraph`'s own.

### Fault-injection seam + mutation-kill test (apply to: every new control, QLT-08)
**Source:** `internal/compiler/pathoracle/pathoracle.go:53-66` (seam),
`internal/compiler/pathoracle/pathoracle_test.go:268-295` (kill test),
`internal/compiler/session/session_phase5_alias_test.go:162` (bilateral/no-divergence
failure — D-07-24 fault 3's analog)
Package-level `var XOverride func(...)`, nil = production default, private
accessor reads it. Test: assert-clean → override → `defer` restore → assert
observable divergence → assert the *specific* code.

### Fail-closed `default:` arm (apply to: `corevalidate`, `interp`, `cgen`)
**Source:** `internal/compiler/corevalidate/corevalidate.go:1000-1002`, `:1214-1216`
```go
		default:
			return v.check(false, "core.unknown_operation", string(operation.Kind))
```
Never widen a `default:` to absorb a new kind; add the explicit case.

### Bounded output with a stable truncation code (apply to: `core.call_graph_cycle` causes)
**Source:** `internal/compiler/protocol/protocol.go:123-131`,
`internal/compiler/session/session_phase6_evidence.go:17-25`
```go
// TraceSummary is the expanded evidence-validation trace: one TraceEntry
// per bound input, bounded like every other untrusted-size projection in
// this project. Truncated carries the stable "truncated:evidence.trace_bound"
// code when the bound is hit, empty otherwise.
type TraceSummary struct {
	Entries   []TraceEntry `json:"entries"`
	Truncated string       `json:"truncated,omitempty"`
}
```
```go
const TruncatedEvidenceTraceBound = "truncated:evidence.trace_bound"
```
`truncated:core.call_cycle_bound` should be an exported named const in the same
style — declared once, never a string literal at the emission site. Note the
project's other codes for reference: `truncated:explain.depth`,
`truncated:explain.node_budget`, `truncated:query.page_bound`.

### Diagnostic identity (apply to: `core.call_graph_cycle`)
**Source:** `internal/compiler/diagnostic/diagnostic.go:96-117`
```go
type Cause struct {
	Kind   string `json:"kind"`
	Detail string `json:"detail,omitempty"`
	Span   *Span  `json:"span,omitempty"`
}

func Error(code string, span Span, message string, causes ...Cause) Diagnostic {
	identity := struct {
		Schema string
		Code   string
		Span   Span
		Causes []Cause
	}{Schema: Schema, Code: code, Span: span, Causes: causes}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	return Diagnostic{Schema: Schema, ID: "diagnostic:" + hex.EncodeToString(sum[:12]), Code: code, Severity: "error", Primary: span, Message: message, Causes: causes}
}
```
`Causes` is hashed **verbatim, in order** — this is why D-07-16's canonical
rotation is load-bearing. Contrast `ErrorWithRepairs` (`:123-151`), which sorts
repairs and reduces them to `Kind` only precisely because ordering and
non-semantic detail must not perturb an ID. Phase 07's rotation is the
`Causes`-side equivalent of that sort, done at the *construction* site because
`Error` deliberately does not sort. Use `Error` (no repairs) for the cycle.

### Content digest (apply to: `ClosureDigest`)
**Source:** `internal/compiler/originvalidate/originvalidate.go:405-408`
```go
func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}
```
Reuse this exact format string (`"sha256:" + full hex`, not the 12-byte prefix
`diagnostic`/`protocol` IDs use). Unkeyed — D-07-13's integrity-not-authenticity
note belongs in the doc comment.

---

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `testdata/phase07/` diamond/shared-leaf, self-call, unreachable-cycle, foreign-shadowing corpus | fixture | file-I/O | No multi-function `.lang` fixture exists anywhere in `testdata/` (verified). Conventions come from `testdata/phase4/foreign_acquire_one.lang`, but the *shape* is genuinely new. Author from scratch per RESEARCH Pitfall 2. |
| `testdata/phase07/` clean-but-uncallable SEM-06 negative control | fixture | file-I/O | RESEARCH Pitfall 3 / A-03: `testdata/phase3/exclusive_borrow_clean` does not exist; the existing `exclusive_*` phase3 fixtures are rejection fixtures, not clean-accept-but-uncallable. No analog. |
| QLT-08 per-control mutation-kill *registry* | config | — | Confirmed: `internal/compiler/session/qlt01_registry.json` + `qlt01.go` + `qlt01_test.go` is a distinct M001-era mechanism and is NOT reusable. The real per-control precedent is the `Test*MutationKilled` / `Test*MutationMatrix` **test-function** convention documented above — there is no registry file to extend. |

---

## Metadata

**Analog search scope:** `internal/compiler/{core,check,corevalidate,originvalidate,pathoracle,interp,cgen,syntax,ast,diagnostic,protocol,session,evidence,cache,debugmap}`, `testdata/phase{1..6}`
**Files scanned:** 22 source files + 6 test files + 3 fixtures (all reads targeted, non-overlapping)
**Pattern extraction date:** 2026-09-08
**Tracked-source gate:** all analog paths verified with `git ls-files`; repository has no capability-mirror layout, so no mirror-path substitution was needed.
