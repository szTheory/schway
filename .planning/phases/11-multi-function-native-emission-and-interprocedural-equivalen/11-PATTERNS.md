# Phase 11: Multi-Function Native Emission and Interprocedural Equivalence - Pattern Map

**Mapped:** 2026-09-11
**Files analyzed:** 13 new/modified production files + 7 new/modified test files
**Analogs found:** 13 / 13 (all files have a concrete, tracked in-repo analog; this
phase is exclusively additive/generalizing extension of existing packages, so
every new file's closest analog is a sibling file in the same package)

All analog paths below were verified tracked (`git ls-files`) before being named.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/cgen/cgen_program.go` (new) | service (compiler backend, TU assembler) | transform (core.Program -> C source) | `internal/compiler/cgen/cgen.go` (`emitLinear`/`emitLinearForeign`, the whole-TU writers) | role-match (same role, adjacent data flow — whole-TU assembly vs. single-function body) |
| `internal/compiler/cgen/cgen.go` (modified: `Emit`/`EmitNative` dispatch fork) | service | transform | itself, `Emit`/`EmitNative` at lines 29-90 | exact (additive dispatch inside the same function, D-04-20 pattern) |
| `internal/compiler/callgraph/callgraph.go` (modified: add `EntryFunction`) | utility (graph derivation) | transform | itself, `buildAdjacency`(215)/`Order`(294) | exact |
| `internal/compiler/session/session.go` (modified: 3 run sites switch to `EntryFunction`) | controller/driver | request-response | itself, `RunInterpreter`/`RunNative`/`interpreterInputs` (759, 783, 872-886) | exact |
| `internal/compiler/session/qlt03_shape_register.go` (new) | service (audit/register loader) | batch (mechanical enumeration + JSON load) | `internal/compiler/session/qlt01.go` (registry loader skeleton) | exact |
| `internal/compiler/session/qlt03_shape_register.json` (new data artifact) | config/data | batch | `internal/compiler/session/qlt01_registry.json` | exact |
| `internal/compiler/session/session_phase11_gate_test.go` (new) | test | request-response (assert-conjunction) | `internal/compiler/session/session_phase6_verify.go`'s `verifyPhase6NativeDifferentialLane` (245-370) as the lane shape to imitate; `session_phase5.go`'s `VerifyPhase5ControlsAndWork` for gate-conjunction structure | role-match |
| `internal/compiler/reduce/reduce.go` (modified: `Seed{Program,EntryFunctionID}`, two new moves, per-function loops, derived `MaxReductionAttempts`) | utility (HDD reducer) | transform / batch | itself, `Reduce`(116-119)/`ProjectSource`(636-639) and the five existing moves (184, 214, 266, 484, 584) | exact |
| `internal/compiler/reduce/reduce_multifunction_test.go` (new) | test | transform | `internal/compiler/reduce/reduce_test.go` (existing single-function move tests — not read this pass, but same package/test convention) | role-match |
| `internal/compiler/session/session_phase5_mismatch.go` (modified: `foreignCallSequenceFor` widened + dynamic second knower) | service | event-driven (drift detection over event stream) | itself, `foreignCallSequenceFor`(72-75)/`mismatchPredicate`(161-204) | exact |
| `internal/compiler/cache/probe.go` (modified: `DeclaredInputNames()` gains 8th entry) | config/utility | CRUD (declared-input registry) | itself, `DeclaredInputNames()`(38-48) | exact |
| `internal/compiler/cache/probe_test.go` (modified/extended) | test | CRUD | `internal/compiler/originvalidate/originvalidate_test.go`'s import-guard test (see Shared Patterns) as the structural-test-shape analog; `cache/probe_test.go` itself for the declared-names assertion shape | role-match |
| `internal/compiler/native/native_lto_test.go` (modified: 4×3 hand-written-C matrix) | test | batch (matrix of compile/run/compare) | itself, existing `-flto` reach assertions (72-97) | exact |
| `internal/compiler/corevalidate/corevalidate.go` (conditionally modified: `peerDeriveOriginFacts` gains `core.OpCall` case, ONLY if Q-01 passes) | service (independent re-deriver) | transform | itself, `peerDeriveOriginFacts`(2407-2450), the existing `OpBorrowShared`/`OpBorrowExclusive`/`OpMove`/`OpCopy` arms | exact |
| `PHASE-11-DEBT.md` (new, at phase open per D-11-02) | doc | — | `.planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` | exact |

## Pattern Assignments

### `internal/compiler/cgen/cgen_program.go` (new — the whole-program TU assembler)

**Analog:** `internal/compiler/cgen/cgen.go` (existing single-function whole-TU
emitters: `emitLinear` ~234, `emitLinearForeign` ~793, `emitBranch` ~1597)

**Dispatch-fork pattern to hook `emitProgram` in** — copy this shape exactly,
do not touch the six functions below the fork (D-04-20 additive-sibling
precedent, verified verbatim):
```go
// Source: internal/compiler/cgen/cgen.go:29-42 (Emit) — EmitNative (61-...) is the
// same shape and both forks land in this phase.
func Emit(program core.Program) (string, error) {
	validated := corevalidate.Validate(program)
	if !validated.Valid {
		return "", fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	if len(program.Functions) != 1 {
		return "", fmt.Errorf("C emitter expects one function")
		// ^ D-11-01 replaces this line's error with: return emitProgram(program)
	}
	function := program.Functions[0]
	if function.Match != nil && function.Linear != nil {
		return emitBranch(program, function)
	}
	// ... (five more single-function branches, byte-for-byte untouched)
}
```

**Includes/typedef/support-block pattern to reuse in the TU assembler**
(the six existing emitters interleave these per-TU; `emitProgram` centralizes
them once per program, per D-11-03):
```go
// Source: internal/compiler/cgen/cgen.go (emitLinearForeign's TU preamble, ~line 261)
out.WriteString("#include <stddef.h>\n#include <stdio.h>\n#include <string.h>\n\n")
if function.Parameter.Type == "Buffer" {
	out.WriteString("typedef struct LANG_BUFFER {\n  unsigned char bytes[4];\n  size_t length;\n} LANG_BUFFER;\n\n")
}
emitLinearOutputSupport(&out, function, typeName)
out.WriteString("int main(int argc, char **argv) {\n")
```
`emitProgram` owns exactly one call to each of `emitEventSupport`,
`emitStreamingEventSupport`, `emitDefectSupport`, `resourceLedger`, and exactly
one `main` — called, never forked, per D-11-01/D-11-03.

**Body-emitter operation-loop pattern** (per-operation `switch` on
`core.OperationKind`, with an inline provenance comment on every
value-producing line — extend this exact shape for the new `emitCall`
call sites):
```go
// Source: internal/compiler/cgen/cgen.go:270-349 (emitLinear's operation loop)
for _, operation := range function.Linear.Operations {
	source := places[operation.SourceID]
	switch operation.Kind {
	case core.OpCopy, core.OpMove, core.OpBorrowShared, core.OpBorrowExclusive:
		target, exists := places[operation.TargetID]
		if !exists || declared[operation.TargetID] {
			return "", fmt.Errorf("operation %q has invalid target", operation.ID)
		}
		// ... label/marker/eventKind selection, then:
		fmt.Fprintf(&out, "  %s %s = %s; /* %s: %s */%s\n", typeName, locals[target.ID], locals[source.ID], label, operation.ID, marker)
	case core.OpReturn:
		// ... (event + lang.execution/1 JSON emission)
	case core.OpCall:
		// THIS is the exact arm emitCall must replace — see next excerpt.
	default:
		return "", fmt.Errorf("operation %q has unknown kind %q", operation.ID, operation.Kind)
	}
}
```

**The exact arm `emitCall` replaces (forward-hygiene error, D-07-39/A-02)** —
this is the site, verbatim, that the new helper must supersede at BOTH of
D-11-04's two call sites (`emitLinear`'s loop above and `emitBranchOperations`,
its block-bodied sibling):
```go
// Source: internal/compiler/cgen/cgen.go:339-345, verified this session
case core.OpCall:
	// D-07-39/A-02: OpCall is registered but not lowered by native
	// emission this phase. Emit/EmitNative hard-fail on
	// len(program.Functions) != 1 (cgen.go:22,52) before this arm
	// could ever run -- no legal OpCall-bearing program has exactly
	// one function -- so this is forward hygiene for Phase 11's
	// multi-function C emission, not a live gap.
	return "", fmt.Errorf("operation %q: Lang-to-Lang calls are not supported by native emission this phase", operation.ID)
```
Note the file ALSO has a test-only fault-injection seam
(`opCallGroupedArmForTest`, lines 17-27, 316-330) guarding against a stub-fold
regression — `emitCall`'s own test suite should add an equivalent seam proving
a stubbed/folded call is caught, following the same unexported-var,
`export_test.go`-only-wrapper convention.

**Error handling pattern:** every emission function returns `("", error)` on
malformed IR (missing target, unknown kind) — `emitCall` and `emitProgram`
must follow this convention, never panic.

---

### `internal/compiler/callgraph/callgraph.go` (add `EntryFunction`)

**Analog:** itself — `buildAdjacency` (215-...) and `Order` (294-...)

**Adjacency-building pattern to reuse for entry resolution** (in-degree-zero
root = no adjacency list has this node as a target):
```go
// Source: internal/compiler/callgraph/callgraph.go:215-267 (buildAdjacency), verified
func buildAdjacency(program core.Program) (map[string][]string, map[[2]string]string, error) {
	declared := make(map[string]bool, len(program.Functions))
	for _, function := range program.Functions {
		declared[function.ID] = true
	}
	// ... dedup edges via edgeSeen map, refuse (not drop) an unresolved CalleeID:
	if !declared[operation.CalleeID] {
		return nil, nil, &unresolvedCalleeError{
			functionID: function.ID, operationID: operation.ID, calleeID: operation.CalleeID,
		}
	}
	// adjacency lists sorted by callee ID for determinism (load-bearing)
}
```
`EntryFunction(program)` should build the same adjacency map, compute in-degree
per node by inverting it, and apply the SAME "refuse, don't guess" discipline
seen in `unresolvedCalleeError` — zero or multiple in-degree-zero roots is a
**named fail-closed refusal** (D-07-45's precedent: "a dropped edge is how a
cycle escapes detection" generalizes to "a guessed root is how a
program-identity mismatch escapes detection").

**Package-doc import-boundary convention** (the doc comment IS the enforced
contract — mirror this style when documenting `EntryFunction`'s new consumer,
`cgen`):
```go
// Source: internal/compiler/callgraph/callgraph.go:1-9, verified
// Package callgraph is Phase 07 Stage 2a's whole-program call-graph
// acyclicity check (SEM-07, D-07-14/D-07-17/D-07-18). It is consumed by
// check, in production, immediately before check.Program returns a
// core.Program -- never by corevalidate, ast, interp, or cgen.
```
**This doc comment is FALSE the moment `cgen`/`session`/`reduce` (D-11-05,
D-11-29) start calling `EntryFunction`** — the plan MUST update this comment
in the same diff that adds the new consumers, and MUST NOT let `reduce` import
this package (D-11-29 explicitly keeps `reduce` out of `callgraph`'s consumer
set; only `cgen` and `session` join `check`).

**Determinism pattern:** iterative three-color DFS, never native recursion,
sorted-by-ID root/adjacency ordering — `EntryFunction`'s root search must sort
candidate roots by function ID before erroring on "many roots" so the error
message names them in a byte-stable order.

---

### `internal/compiler/session/session.go` (3 run sites switch to `EntryFunction`)

**Analog:** itself

**Site 1 — `RunInterpreter`-style guard, verified verbatim:**
```go
// Source: internal/compiler/session/session.go:759 area (interp path)
if len(checked.Program.Functions) != 1 { /* refuse */ }
interp.Run(checked.Program, checked.Program.Functions[0].Name, input)
```
Replace `checked.Program.Functions[0].Name` with
`callgraph.EntryFunction(checked.Program).Name` (propagating its error into
the same refusal path); widen the guard to permit N>1 while still requiring
`callgraph.EntryFunction` to resolve successfully.

**Site 2 — `interpreterInputs`:** same `Functions[0]` substitution; the
function's existing single-`Byte`/`Buffer`-parameter input-synthesis logic is
unchanged — only which function it reads the parameter from changes.

**Site 3 — `RunNative`:** identical substitution, feeding the resolved entry
into both `interp.Run` (oracle) and `cgen.EmitNative` (binary), which is
exactly the mechanism D-11-05 requires ("the oracle and the binary can never
disagree about which function *is* the program").

---

### `internal/compiler/session/qlt03_shape_register.go` + `.json` (new)

**Analog:** `internal/compiler/session/qlt01.go` (the skeleton D-11-43 explicitly
says to reuse — "same skeleton, different row key, never merge the files")

**Embed + load pattern, copy near-verbatim (rename types/embed target):**
```go
// Source: internal/compiler/session/qlt01.go:1-46, verified
package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	_ "embed"
)

//go:embed qlt01_registry.json
var qlt01RegistryBytes []byte

type QLT01LiveDescendant struct {
	Fixture   string `json:"fixture"`
	ControlID string `json:"control_id"`
}

type QLT01Waiver struct {
	Reason   string `json:"reason"`
	Citation string `json:"citation"`
	Owner    string `json:"owner"`
	Phase    string `json:"phase"`
}

type QLT01Row struct {
	SpikeID          string               `json:"spike_id"`
	ControlID        string               `json:"control_id"`
	ControlMechanism string               `json:"control_mechanism"`
	LiveDescendant   *QLT01LiveDescendant `json:"live_descendant"`
	Waived           *QLT01Waiver         `json:"waived"`
}

func LoadQLT01Registry() ([]QLT01Row, error) {
	var rows []QLT01Row
	if err := json.Unmarshal(qlt01RegistryBytes, &rows); err != nil {
		return nil, fmt.Errorf("qlt01: failed to parse embedded registry: %w", err)
	}
	return rows, nil
}
```
For QLT-03: row key is the axis-product cell (edge topology, argument mode,
callee body shape, return-origin class, composition-depth bucket), NOT
`spike_id` — this is exactly why D-11-43 forbids extending `qlt01_registry.json`
("its row key is `spike_id` over an unrelated universe"). Reuse the
discriminated-union idea (`LiveDescendant *X` XOR `Waived *X`) as
`Reached *ReachedEvidence` XOR `Unreachable *UnreachableProof`, and the
placeholder-ban discipline (`Waiver.Reason`/`Citation` must never be free
text) as `UnreachableProof.Mechanism` being one of the five closed enum values
(`grammar`/`refused`/`generator`/`bounded_exhaustive`/`gap`) per D-11-46.

**JSON artifact shape, copy the discriminated-row style verbatim (rename
fields):**
```json
{
  "spike_id": "001",
  "control_id": "spike-001-move-while-borrowed-fault",
  "control_mechanism": "a checker that permits transferring ownership ...",
  "live_descendant": null,
  "waived": {
    "reason": "the spike's ad hoc differential-search counterexample has no control:-registry identifier of its own; ...",
    "citation": "internal/compiler/check/check.go's \"ownership.move_while_borrowed\" diagnostic ..., both proven zero-divergence by check_test.go's TestOwnershipSequenceExhaustive",
    "owner": "compiler-frontend",
    "phase": "05"
  }
}
```
Every `waived`/`unreachable`-shaped row cites a real test name (here
`TestOwnershipSequenceExhaustive`) — QLT-03's rows must do the same (D-11-44:
"every hand-written row names a falsifier test that goes red if the claim
stops being true").

---

### `internal/compiler/reduce/reduce.go` (Seed type, two new moves, per-function loops)

**Analog:** itself

**The exact guard `Reduce` currently hard-fails with, verbatim (the site
`Seed{Program, EntryFunctionID}` replaces):**
```go
// Source: internal/compiler/reduce/reduce.go:116-119, verified
func Reduce(ctx context.Context, seed core.Program, interesting Predicate) (Result, error) {
	if len(seed.Functions) != 1 {
		return Result{}, fmt.Errorf("reduce: seed must carry exactly one function, got %d", len(seed.Functions))
	}
```

**The exact guard `ProjectSource` currently hard-fails with, verbatim
(returns a STRING, not an error — preserve this convention for any new
refusal message):**
```go
// Source: internal/compiler/reduce/reduce.go:636-639, verified
func ProjectSource(program core.Program) string {
	if len(program.Functions) != 1 {
		return unsupportedProjection(fmt.Sprintf("expected exactly one function, got %d", len(program.Functions)))
	}
```
`unsupportedProjection` (line 686-688) wraps as a
`// reduce: projection unsupported for this program shape -- ...` C comment —
new `RefusedShapes()` entries (`reduce.seed_shape_unsupported`,
`reduce.projection_unsupported`, `reduce.reverification_signature_drift`)
should route through this same string-wrapping helper, not a new one.

**Per-function indexing sites to convert to loops** (five sites, mechanical):
`reduce.go:184, 214, 266, 484, 584` each do `fn := p.Functions[0]` — confirmed
pattern to replicate per function in `callgraph.Order`'s reverse-postorder
(NOT native iteration order), never importing `callgraph` itself (see import
constraint below).

**Import boundary to preserve — verified current import block:**
```go
// Source: internal/compiler/reduce/reduce.go:16-23 (approx), verified via research
// imports: context, encoding/json, fmt, strings, core — NO callgraph import.
```
`reduce`'s package doc and `callgraph`'s own doc comment
(`internal/compiler/callgraph/callgraph.go:1-9`) both currently claim
`callgraph` is consumed by `check` alone. D-11-29 requires `reduce` to walk in
reverse-postorder WITHOUT importing `callgraph` — either `Seed` carries a
pre-computed order (caller-supplied, matching D-11-30's "entry function is
caller-supplied, never minted" philosophy) or `reduce` re-derives its own
minimal topological order locally. Do not add `import ".../callgraph"` to
`reduce.go`; if the plan does, `callgraph.go`'s doc comment AND
`TestImportsStayIndependent`-style guard (see Shared Patterns) must be updated
together, and this is new scope the plan should flag explicitly.

---

### `internal/compiler/session/session_phase5_mismatch.go` (`foreignCallSequenceFor` widening)

**Analog:** itself

The exact site returning `nil` for multi-function programs today
(`session_phase5_mismatch.go:72-75`, per RESEARCH.md's verified citation) must
widen to walk every function via reverse-postorder (same ordering-source
constraint as `reduce` above — this file lives in `session`, which MAY import
`callgraph`, so this is the natural place to call `callgraph.Order` directly)
and add a second, independently-derived sequence from the `-O0` run's event
stream (`FunctionID`-tagged events already exist per the research). This is
the "two independent knowers" pattern (see Shared Patterns) applied to
silent-slippage detection (D-11-33).

---

### `internal/compiler/cache/probe.go` (`DeclaredInputNames()` 8th entry)

**Analog:** itself, verbatim current state:

```go
// Source: internal/compiler/cache/probe.go:38-48, verified
func DeclaredInputNames() []string {
	return []string{
		"fixture_source",
		"build_flags",
		"clang_identity",
		"runtime_identity",
		"foreign_translation_unit",
		"mutation_runner_source",
		"go_toolchain",
	}
}
```
Add exactly one new entry (e.g. `"cgen_source"`) naming the digest of
`internal/compiler/cgen/*.go`'s own source, per D-11-41. Follow the doc
comment's own numbered-hole convention (lines above the function,
"D-06-13 records four such holes") — append a fifth numbered item disclosing
this was a fifth hole, now closed, rather than silently deleting/renumbering
the existing four.

**Package-doc dependency-free-leaf constraint to preserve, verified verbatim:**
```go
// Source: internal/compiler/cache/probe.go:15-19
// MaxProbeBytes bounds a single spawned tool probe's stdout/stderr,
// mirroring evidence.MaxToolProbeBytes's 64 KiB-plus-one bounding mechanism
// (D-06-07). cache is a dependency-free leaf package (see cache.go's
// package doc) and therefore does not import evidence; this is a
// deliberate, minimal duplicate of the same bounding shape -- not a shared
// helper, and never a second, weaker bounding mechanism.
```
This is the load-bearing evidence for D-11-39b/QLT-06b's structural test: the
new `cgen_source` input must be computed by hashing files directly (`os`/
`crypto/sha256`, already imported) — **never** by importing `core` or
`corevalidate` into `cache` to ask them for a digest. Confirmed current
imports: `bytes, context, crypto/sha256, encoding/hex, errors, os, os/exec,
runtime, time` — no `core`/`originvalidate`.

---

### `internal/compiler/native/native_lto_test.go` (NAT-07's 4×3 matrix)

**Analog:** itself — existing `-flto` reach assertions at lines 72-97
(`native.Runner`'s single `program.c`, `Options.LTO` appended to both compile
and link argument lists per `native.go:57-66`). Extend
`AliasFactMutationRunner`'s shape (`session_phase5_alias.go:46,67,351`,
`injectRestrictIntoSignature` ~147) — the false `restrict` injection on a
callee TU's by-pointer parameter, the aliasing write in a third TU, caller
binding probe==primary — as a hand-written-C matrix, not a Lang-source-driven
one (per D-11-24, this must ship BEFORE `cgen` writes multi-source output).

---

### `internal/compiler/corevalidate/corevalidate.go` (`peerDeriveOriginFacts` — conditional on Q-01)

**Analog:** itself, verbatim current gap:

```go
// Source: internal/compiler/corevalidate/corevalidate.go:2407-2450, verified this session
func peerDeriveOriginFacts(function *core.Function) peerOriginFact {
	if function.Linear == nil {
		return peerOriginFact{}
	}
	paramTrace := map[string]bool{function.Parameter.ID: true}
	derived := make(map[string]string)
	for _, operation := range function.Linear.Operations {
		switch operation.Kind {
		case core.OpBorrowShared:
			// ...
		case core.OpBorrowExclusive:
			// ...
		case core.OpMove, core.OpCopy:
			// ...
		}
		// NOTE: no `case core.OpCall:` arm exists in this switch.
	}
	// ... OpReturn scan
}
```
If Q-01 passes, add a `case core.OpCall:` arm following the exact structural
shape of the existing `OpMove, core.OpCopy` arm (forwarding `derived[TargetID]
= derived[SourceID]`-style propagation) — do not invent a new derivation
style. The seam-style precedent for any fault-injection test this gains is
`disablePeerOriginContainmentForTest` (corevalidate.go:2452-2460).

---

## Shared Patterns

### Additive Sibling Over Signature Change (D-04-20)

**Source:** `internal/compiler/cgen/cgen.go:29-42` (`Emit`'s dispatch), same
shape at `EmitNative`
**Apply to:** `cgen_program.go`'s `emitProgram` fork, `reduce.go`'s two new
moves (added to the `Moves` slice, never replacing an existing move), and
`cache/probe.go`'s eighth declared-input entry (appended, not replacing the
existing seven — `TestDeclaredInputNamesStableAndNoInterproceduralImport`
must assert the first seven are byte-identical and ordered).
```go
if len(program.Functions) != 1 {
	return emitProgram(program) // additive dispatch branch
}
// ... six existing single-function paths, untouched
```

### Two Independent Knowers, Never One Derivation Read Twice

**Source (peer-derivation precedent):** `corevalidate.peerDeriveOriginFacts`
(corevalidate.go:2407) is `corevalidate`'s OWN independent re-derivation,
never calling into `callgraph` or `check` — verified by `callgraph.go`'s own
doc comment naming its consumer set as `check` alone (never `corevalidate`).
**Source (mid-phase gate precedent, D-11-14):**
```
knower 1: cgen.ScanForBannedAttributes(emittedArtifacts...) == nil   // reads OUTPUT BYTES
knower 2: check-derived count of functions whose by-pointer lowering
          WOULD have selected restrict, N >= 1                       // reads the PROGRAM
assert: knower1_empty AND knower2_N>=1 AND structural_floors AND (-O0 == interpreter)
```
**Apply to:** the mid-phase gate test (`session_phase11_gate_test.go`), the
reducer's `foreignCallSequenceFor` silent-slippage guard (static
reverse-postorder walk vs. dynamic `-O0` event-stream derivation, D-11-33),
and QLT-06b's structural test (declared-input-count assertion is a SEPARATE
knower from "does `cache` actually import `core`" — both must be asserted,
neither substitutes for the other).

### Structural Import Guards (`go list -deps` + direct AST scan)

**Source:** `internal/compiler/originvalidate/originvalidate_test.go` —
this is THE analog to copy verbatim for any new Phase 11 import-boundary
test (notably `cache` must not import `core`/`originvalidate`, D-11-39b).

Forbidden-import list declaration:
```go
// Source: internal/compiler/originvalidate/originvalidate_test.go:44, verified
var originValidateForbiddenImports = []string{"/compiler/check", "/compiler/ast", "/compiler/corevalidate"}
```

Direct-scan test (parses only the package's own `.go` files' import blocks
via `go/parser` with `ImportsOnly`):
```go
// Source: internal/compiler/originvalidate/originvalidate_test.go (directImportViolation, ~lines 50-81)
fileSet := token.NewFileSet()
for _, entry := range entries {
	if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
		continue
	}
	file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
	// ... for each import, HasSuffix match against forbidden list
}
```

Transitive-scan test (hardens the direct scan via `go list -deps`, bounded
and timed per this repo's own spawn-safety lint):
```go
// Source: internal/compiler/originvalidate/originvalidate_test.go:142-158, verified
func transitiveImportsViolation(t *testing.T, forbidden []string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), goListDepsTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "list", "-deps", "github.com/codename-lang/lang/internal/compiler/originvalidate")
	cmd.Dir = testsupport.ProjectPath()
	stdout := &boundedGoListWriter{}
	cmd.Stdout = stdout
	if err := cmd.Run(); err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	if stdout.overflowed {
		t.Fatalf("go list -deps produced more than %d bytes; dependency listing is implausibly large", maxGoListDepsOutputBytes)
	}
	deps := strings.Split(strings.TrimSpace(stdout.buffer.String()), "\n")
	return transitiveImportViolation(deps, forbidden)
}
```
Both scans must run (direct scan is NOT replaced by the transitive one — this
repo's own comment states "hardened, not replaced"). Bounds:
`maxGoListDepsOutputBytes = 1 << 20`, `goListDepsTimeout = 2 * time.Minute`.

**Negative control (the anti-vacuity pattern, D-11-18's precedent):**
```go
// Source: internal/compiler/originvalidate/originvalidate_test.go:209-218, verified
func TestTransitiveImportsGuardCanFail(t *testing.T) {
	synthetic := []string{
		"github.com/codename-lang/lang/internal/compiler/originvalidate",
		"github.com/codename-lang/lang/internal/compiler/core",
		"github.com/codename-lang/lang/internal/compiler/corevalidate",
	}
	if got := transitiveImportViolation(synthetic, originValidateForbiddenImports); got == "" {
		t.Fatal("expected the synthetic dependency list's forbidden corevalidate entry to be flagged")
	}
}
```
**Apply to:** the new `cache`-does-not-import-`core`/`originvalidate` test
(D-11-39b/QLT-06b), and (if a plan ever risks it) a `reduce`-does-not-import-
`callgraph` guard extending D-11-29's constraint mechanically rather than by
review alone — copy this exact three-part shape (forbidden-list var, direct
scan, transitive scan, negative control) for either.

### Verification Lane Shape (`session` phase-N verify lanes)

**Source:** `internal/compiler/session/session_phase6_verify.go:245-370`
(`verifyPhase6NativeDifferentialLane`) — the concrete analog for a new
interprocedural lane (the mid-phase gate lane, or any NAT-06 multi-function
differential lane):
```go
// Source: internal/compiler/session/session_phase6_verify.go:245-370, verified this session (abridged)
func verifyPhase6NativeDifferentialLane(ctx context.Context, source []byte, runner native.Runner, store *cache.Store) (protocol.Lane, cache.Outcome, error) {
	laneStarted := time.Now()
	recorder := &StageRecorder{}
	// parse -> check -> corevalidate.Validate, unconditionally, BEFORE cache is consulted
	// (checker/validator can never be skipped by a cache outcome)
	...
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) != 0 || len(checked.Program.Functions) != 1 {
		return protocol.Lane{}, cache.Outcome{}, fmt.Errorf("phase6.fixture_invalid: checker rejected the native-differential fixture")
	}
	validated := corevalidate.Validate(checked.Program)
	...
	cSource, err := cgen.EmitNative(program)
	...
	spec := phase6ArtifactSpec(source, clangPath)
	outcome, err := cache.Consult(ctx, store, spec)
	var binary []byte
	if outcome.Status == cache.StatusArtifactReused {
		binary = outcome.Artifact
	} else {
		compiled, compileErr := phase6CompileBinary(ctx, clangPath, cSource, phase6BuildFlags, runner.Timeout, recorder)
		...
		if outcome.Status == cache.StatusArtifactRecomputed {
			store.Put(outcome.Key, binary)
		}
	}
	// comparator ALWAYS re-runs fresh here, unconditionally, regardless of cache hit/miss
	for _, input := range inputs {
		interpreted, _ := interp.Run(program, program.Functions[0].Name, input)
		actual, _ := phase6RunCompiledBinary(ctx, binaryPath, input, runner.Timeout)
		if !execution.Equal(interpreted, actual) {
			return protocol.Lane{Status: protocol.StatusMismatch, ...}, outcome, nil
		}
	}
	return protocol.Lane{Status: protocol.StatusPass, Controls: []string{"control:interpreter-o0-o3"}, ...}, outcome, nil
}
```
**Apply to:** any new Phase 11 lane function. Key structural rules this
example enforces that a new lane must preserve: (1) check+corevalidate run
BEFORE cache consult, unconditionally; (2) the equivalence comparator reruns
fresh after cache hit OR miss — a cache hit changes only whether `clang` ran;
(3) `program.Functions[0].Name` here is exactly one of the sites D-11-05
requires switching to `callgraph.EntryFunction(program).Name`.

### JSON Register Artifact (committed audit data)

**Source:** `internal/compiler/session/qlt01_registry.json` (sibling:
`qlt02_budget_manifest.json`) — the analog for `qlt03_shape_register.json`.
Discriminated-union row shape (exactly one of two nullable sub-objects
populated per row), free-text-citation ban (every `waived`/`unreachable` row
names a real, existing test by name), embedded via Go's `//go:embed` directive
rather than read from disk at runtime (keeps the artifact content-addressed
with the binary). See the full JSON excerpt under
`qlt03_shape_register.go`'s Pattern Assignment above.

## No Analog Found

None — every file in this phase's scope generalizes or sits directly beside
an existing, previously-verified analog in the same package. This is a direct
consequence of D-11-01 through D-11-50's additive-sibling design discipline:
the phase deliberately avoids introducing any wholly new architectural
concept (no new package, no new external dependency, no new schema field on
`core.Program`).

## Fixture/Testdata Layout

`testdata/phaseNN/*.lang` — one flat directory per phase, `.lang` source files
named descriptively (`call_basic.lang`, `deep_diamond_acyclic.lang`,
`cycle_unreachable.lang`, `call_argument_used_twice.lang`), consumed by
package tests via `testsupport.ProjectPath("testdata", "phaseNN", "name.lang")`-style
helpers. Phase 11's new interprocedural corpus follows the SAME flat
convention in a new `testdata/phase11/` directory (verified sibling
directories: `testdata/phase07/`, `testdata/phase08/`, `testdata/phase10/`
already exist with this exact shape — `testdata/phase11/` does not exist yet
and must be created following the identical naming style, e.g.
`multi_function_entry_basic.lang`, `multi_function_diamond_call.lang`).
Golden files pair `<name>.lang` with `<name>.golden.c` in the SAME directory
when a fixture pins exact emitted C bytes (analog:
`testdata/phase5/restrict_borrow.lang` + `.golden.c`).

## Metadata

**Analog search scope:** `internal/compiler/{cgen,callgraph,session,reduce,
cache,corevalidate,native,originvalidate,testsupport}`, `testdata/phase{07,08,
10,5}`, `.planning/phases/10-trusted-interprocedural-oracle/`
**Files scanned:** ~14 read/grepped this session (cgen.go, callgraph.go,
reduce.go [via RESEARCH.md verified excerpts], session.go,
session_phase6_verify.go, session_phase5_mismatch.go [via RESEARCH.md],
qlt01.go, qlt01_registry.json, cache/probe.go,
originvalidate/originvalidate_test.go, corevalidate.go [via RESEARCH.md],
native.go/native_lto_test.go [via RESEARCH.md])
**Pattern extraction date:** 2026-09-11
