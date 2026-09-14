# Phase 10: Trusted Interprocedural Oracle - Pattern Map

**Mapped:** 2026-09-11
**Files analyzed:** 14 (create/modify)
**Analogs found:** 14 / 14 (every file has an exact or role-match in-repo analog; this phase is explicitly a shipped-precedent extension exercise, not new design)

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/originvalidate/originvalidate.go` (widen + `OpCall` case) | service (static re-deriver) | transform (backward walk over checked core) | `internal/compiler/corevalidate/corevalidate.go` (`buildLoanChainIndex`, `peerLoanCarry` threading) | exact — D-10-03 names this the shipped precedent to mirror |
| `internal/compiler/originvalidate/originvalidate_test.go` (guard hardening + OpCall tests + D-10-08 gate) | test | request-response (CLI-level) + unit | itself (existing guard tests at lines 33/57) + `internal/compiler/pathoracle/pathoracle_test.go:51` (`TestOracleImportsStayIndependent`, already has a 5-item forbidden list) | exact for the widen tests; role-match for transitive-guard shape |
| `internal/compiler/pathoracle/pathoracle.go` (or sibling file — composition + depth cap) | service (static re-deriver, enumerator) | transform (recursive path enumeration) | itself — `EnumeratePaths`/`linearizePath` (existing), `pathCapError` (lines 87-110) | exact — D-10-09 requires reusing pathoracle's own mechanism recursively, not a new one |
| `internal/compiler/pathoracle/pathoracle_test.go` (discriminating fixture + transitive guard) | test | unit + CLI-level | itself; `originvalidate_test.go`'s guard-test shape | exact |
| `internal/compiler/interp/interp.go` (`[]frame` stack, `MaxCallDepth`, 3 `OpCall` arms, frame-partition helper, drop ordering) | service (deterministic interpreter / oracle) | event-driven (CanonicalBytes/Execution emission) | `internal/compiler/callgraph/callgraph.go` (`Order`'s iterative explicit-stack DFS, lines 283-400) for the stack shape; itself (existing single-frame `Run`, `runBranchArm`/`runLinearBlocks`/`runLinear`) for the execution-path shape | exact for stack shape; exact for execution-path shape (extending, not replacing) |
| `internal/compiler/interp/interp_test.go` (depth-exceeded, drop-order, native-stack probe, mutant-kill pairs, OWN-05b guard) | test | integration (full pipeline) + subprocess + unit | itself — `checkedCallBasicProgram` (`interp_test.go:30`), `opCallGroupedArmForTest`/`TestOpCallGroupedArmMutationKilled` (existing sole test) | exact for fixture-through-pipeline shape; no in-repo analog for the subprocess probe (see "No Analog Found") |
| `internal/compiler/corevalidate/corevalidate.go` (export `recomputeLoanEndpoints` as `RecomputeEndpoints`-shaped accessor; add per-signature drop-before-pop invariant) | service | CRUD-like (compute-once, expose) | itself — `pathoracle.RecomputeEndpoints` (already exported, same shape) is the sibling to match; `derivePeerSignature`'s existing per-function-declaration check shape (`corevalidate.go:2163` area) for the new invariant | exact |
| `internal/compiler/session/session_peer_gate_test.go` (extend to 3/4-way, `peerDivergenceExpected` → accountable struct) | test | batch (corpus walk) | itself — `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (lines 226-279), `peerDivergenceExpected` (line 23) | exact — D-10-51 requires extending, not duplicating |
| `testdata/phase10/relay_depth3_accept.lang` / `relay_depth3_refuse.lang` (borrow varied per hop) | test fixture | file I/O (static `.lang` source) | `testdata/phase08/relay_depth2_accept.lang` (structure to extend, shape to NOT repeat — it's owned pass-through, D-10-48 forbids copying that shape) | role-match, deliberately divergent content |
| `testdata/phase10/` 129-function call-depth-exceeded fixture | test fixture | file I/O | `testdata/phase07/` (or wherever `checkedCallBasicProgram`'s source fixture lives) — the "genuine `.lang` source through the real pipeline" pattern (D-10-24) | role-match |
| `testdata/phase10/interp_oracle/` golden corpus | test fixture (golden) | file I/O (byte-for-byte comparison) | no existing golden-corpus directory in this repo — new pattern this phase introduces (see "No Analog Found") | none — new |
| `internal/compiler/check/check.go` (doc-comment-only correction to `deriveFunctionUsesParam`, D-10-29) | service (doc fix, no logic change) | — | itself (lines 700-741) | exact (self-modification, comment only) |
| `testdata/phase08/twin_b_accept.lang` (header comment correction, D-10-29) | test fixture | file I/O | itself | exact (self-modification, comment only) |
| `.planning/phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md` | doc / debt register | — | `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/PHASE-09-DEBT.md` (D-09-44's mechanically-checked format, verified by `TestDebtRegistersAreWellFormed`, `session/session_test.go:2613-2671`) | exact |

## Pattern Assignments

### `internal/compiler/originvalidate/originvalidate.go` (service, transform)

**Analog:** `internal/compiler/corevalidate/corevalidate.go` (`buildLoanChainIndex`, lines 1185-1213) — the D-09-03 fix for the structurally identical cross-function-fact-threading problem.

**Imports pattern** (`originvalidate.go:1-19`, unchanged — the new code must NOT add to this list beyond what's already there):
```go
// Package originvalidate independently re-derives and verifies a function's
// declared public borrow origin (OWN-04) from the typed-core artifact alone.
// It intentionally does not know source or reuse checker code: it imports
// neither internal/compiler/check nor internal/compiler/ast...
package originvalidate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/callgraph"
	"github.com/codename-lang/lang/internal/compiler/core"
)
```
D-10-19: `callgraph` stays imported (defensible-by-review asymmetry, do not touch). `corevalidate` must NEVER appear here — that is exactly what the widened guard (D-10-16) checks for.

**Narrow-map threading pattern to mirror** (`corevalidate.go:1185-1213`):
```go
// buildLoanChainIndex builds the parent-pointer chain used by
// carriedLoans. loanCarry is Phase 09's own peer-derived interprocedural
// loan-liveness fact (D-09-03): before this unconditionally sets
// idx.parent[operation.TargetID] = operation.SourceID for an OpCall, it
// first consults loanCarry[operation.CalleeID]...
func buildLoanChainIndex(operations []core.LinearOperation, checks *int, loanCarry map[string]peerLoanCarryFact) *loanChainIndex {
	idx := &loanChainIndex{ /* ... */ }
	for _, operation := range operations {
		if operation.Kind == core.OpBorrowShared || operation.Kind == core.OpBorrowExclusive {
			idx.bornAt[operation.TargetID] = operation.LoanID
		}
		if operation.Kind == core.OpCall && !disablePeerLoanCarryConsultForTest && !forcePeerLoanCarryTrueForTest && !loanCarry[operation.CalleeID].ReturnsBorrowOfParam {
			continue // callee's own declared contract does not carry
		}
		if operation.TargetID != "" {
			idx.parent[operation.TargetID] = operation.SourceID
		}
	}
	return idx
}
```
Mirror shape for the new widening: a `map[string]<minimal fact struct>` built once in `ValidatePublished`, threaded as an explicit parameter through `RecomputeOriginPerReturn(function, calleeContracts) → walkReturnOrigin(function, sourceOf, returnOp, calleeContracts)`. **Never** `core.Program` (D-10-02).

**Core pattern — the switch to extend** (`originvalidate.go:171-218`, `walkReturnOrigin`):
```go
func walkReturnOrigin(function core.Function, sourceOf map[string]core.LinearOperation, returnOp *core.LinearOperation) ReturnOrigin {
	current := returnOp.SourceID
	visited := make(map[string]bool)
	derivedAccess := ""
	for current != function.Parameter.ID {
		if visited[current] { return ReturnOrigin{OperationID: returnOp.ID} }
		visited[current] = true
		operation, exists := sourceOf[current]
		if !exists { return ReturnOrigin{OperationID: returnOp.ID} }
		switch operation.Kind {
		case core.OpBorrowExclusive:
			if derivedAccess == "" { derivedAccess = "exclusive" }
		case core.OpBorrowShared:
			if derivedAccess == "" { derivedAccess = "shared" }
		case core.OpForeignCall:
			// D-04-28: a foreign declaration is itself a signature carrying
			// origin and access facts...
			if derivedAccess == "" && function.ForeignContract != nil {
				switch function.ForeignContract.Alias {
				case "borrow": derivedAccess = "shared"
				case "retain": derivedAccess = "exclusive"
				}
			}
		// NEW: case core.OpCall: consult calleeContracts[operation.CalleeID]
		}
		current = operation.SourceID
	}
	if derivedAccess == "" { return ReturnOrigin{OperationID: returnOp.ID} }
	return ReturnOrigin{OperationID: returnOp.ID, Paths: []string{function.Parameter.Name}, Access: derivedAccess, Derived: true}
}
```
Per D-10-07: the `case core.OpCall` mirrors `OpForeignCall`'s "read a declared contract at the hop" shape but is NOT exact — `OpForeignCall` reads `function.ForeignContract` (current function, zero lookup); `OpCall` must look up a DIFFERENT function via the new map. State this in the doc comment on the new case.

**Value-type discipline (D-10-04):** the map's value type carries only `Access`/derived-ness — never the full `core.PublicOrigin` shape (which additionally has `Paths []string`, unneeded here). Pair with a same-package signature-text assertion test that fails if `walkReturnOrigin` ever gains a `core.Function`- or `core.Program`-shaped parameter.

**Sequencing (D-10-06):** one plan, two commits — commit 1 widens the signature and builds the map with the `OpCall` case absent/inert (suite green); commit 2 adds the real consult and flips `testdata/phase08/twin_a_accept.lang`'s fixture (D-10-08's gate).

---

### `internal/compiler/originvalidate/originvalidate_test.go` (test)

**Analog:** itself (existing guard tests) + `pathoracle_test.go:51-75`'s already-5-item forbidden list as the target shape.

**Existing guard pattern to extend** (`originvalidate_test.go:33-79`, both `TestOriginValidatorImportsStayIndependent` and `TestOriginValidateImportsNeitherCheckNorAst`):
```go
func TestOriginValidatorImportsStayIndependent(t *testing.T) {
	dir := testsupport.ProjectPath("internal", "compiler", "originvalidate")
	entries, err := os.ReadDir(dir)
	// ...
	fileSet := token.NewFileSet()
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
		// ...
		for _, imported := range file.Imports {
			path := strings.Trim(imported.Path.Value, `"`)
			if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
				t.Fatalf("%s imports %s, which originvalidate (including the foreign-origin path) must never depend on", entry.Name(), path)
			}
		}
	}
}
```
**D-10-16 gap to close:** add `"/compiler/corevalidate"` to BOTH forbidden lists (currently forbids only `check`/`ast`). Reference `pathoracle_test.go:67` for the target shape — pathoracle's own list is already `[]string{"/compiler/check", "/compiler/corevalidate", "/compiler/ast", "/compiler/interp", "/compiler/cgen"}`.

**D-10-17 hardening:** supplement the direct `go/parser` scan (keep it — the redundancy is intentional per D-10-18) with a `go list -deps` transitive check via `os/exec`, asserting none of the transitive dependency set matches the forbidden suffixes. Apply the identical addition to `pathoracle_test.go`'s `TestOracleImportsStayIndependent`.

**D-10-08 gate seam pattern to copy** (`pathoracle.go:59-64`'s `TerminatorKindsOverride`, already established in a sibling package):
```go
var TerminatorKindsOverride func() []core.OperationKind
func recognizedTerminatorKinds() []core.OperationKind {
	if TerminatorKindsOverride != nil { return TerminatorKindsOverride() }
	return core.TerminatorKinds()
}
```
Build an analogous nil-default seam that forces the new `OpCall` case to no-op, demonstrating `twin_a_accept.lang` regresses to the old wrong refusal when the seam is engaged — proving the case is load-bearing, not merely present syntactically.

**Full-CLI gate assertion pattern:** `session.CheckCommandFile` (see `session_peer_gate_test.go`'s usage) is the pipeline the D-10-08 test must run through (fixed precedence check → corevalidate → originvalidate), not a direct package-level call.

---

### `internal/compiler/pathoracle/pathoracle.go` (or sibling file) (service, transform)

**Analog:** itself — package doc (lines 1-28), `EnumeratePaths`/`linearizePath`, `MaxPaths` + `pathCapError` (lines 33-110).

**Package doc pattern to extend, precisely** (`pathoracle.go:1-28`):
```go
// Package pathoracle is Phase 3's third, genuinely different decision
// procedure for loan liveness...
//   - This package enumerates every DISTINCT acyclic entry-to-return path
//     as an explicit, concrete block sequence, replays each one forward in
//     isolation (never converging or closing anything), and only combines
//     the independent per-path answers at the very end...
package pathoracle
```
D-10-11: the doc addition for `OpCall` composition must state the precise property — "never converges, never closes a relation, only enumerates and replays concrete paths" — and that composition preserves this under call-crossing while a contract hop (the pattern `check`/`corevalidate` already use at `OpCall`) would not, because it collapses a callee's per-path distinctions into one summary BEFORE replay.

**Cap-plus-typed-error idiom to mirror exactly** (`pathoracle.go:33-46, 87-110`):
```go
// MaxPaths bounds the number of acyclic entry-to-return paths this package
// will enumerate for one function... MaxPaths is set well above that...
const MaxPaths = 4096

type pathCapError struct {
	functionID string
	limit      int
}
func (e *pathCapError) Error() string {
	return fmt.Sprintf("pathoracle.path_count_exceeded: function %q exceeds the declared cap of %d acyclic entry-to-return paths", e.functionID, e.limit)
}
func (e *pathCapError) Code() string { return "pathoracle.path_count_exceeded" }
func PathCapError(err error) (*pathCapError, bool) {
	e, ok := err.(*pathCapError)
	return e, ok
}
```
D-10-12: add a SEPARATE composition-depth cap with its own `*compositionDepthCapError`-shaped type and `Code()`, mirroring this exactly — never folding into `MaxPaths`. Unlike `MaxPaths` (deliberately set above the real reachable maximum, 4096 vs 64), the new composition-depth cap sits AT the declared depth (3, per D-10-46) — a genuinely reachable/declared bound, not a decorative ceiling.

**Composition mechanism (D-10-09):** recurse into the callee via pathoracle's OWN `EnumeratePaths`/`linearizePath`, splice the callee's enumerated concrete paths into the caller's path sequence, then run the EXISTING forward replay over the spliced sequence. No new convergence/fixpoint machinery — this is the existing enumerator closing over itself, recursively.

**Cycle guard (D-10-15):** composition needs its OWN mutual-recursion guard across the call edge (distinct from `EnumeratePaths`'s existing intra-function back-edge handling). Document this as fail-closed defence for a corrupted core artifact — `callgraph` already refuses cycles before this runs, so state that this guard is unreachable in a correctly-produced artifact, mirroring the doc-comment discipline `pathoracle.go`'s existing `backEdgeError` section already uses for intra-function cycles.

---

### `internal/compiler/pathoracle/pathoracle_test.go` (test)

**Analog:** itself — `TestOracleImportsStayIndependent` (lines 51-75), already the most complete forbidden-list guard in the repo (5 entries: check/corevalidate/ast/interp/cgen).

**Existing forbidden-list shape** (`pathoracle_test.go:67-70`):
```go
forbidden := []string{"/compiler/check", "/compiler/corevalidate", "/compiler/ast", "/compiler/interp", "/compiler/cgen"}
for _, bad := range forbidden {
	if strings.HasSuffix(path, bad) {
		t.Fatalf("%s imports %s, which pathoracle must never depend on", entry.Name(), path)
	}
}
```
D-10-17: only the mechanism needs hardening here (add the `go list -deps` transitive check) — the forbidden list itself is already correct; do not add entries.

**D-10-14's required discriminating fixture** — construct a callee with two distinct concrete paths where only one borrows a parameter into its return; assert the caller's composed paths correctly split into an edge endpoint (divergent) vs a point endpoint (non-divergent) depending on which callee path is spliced. Pattern this test using the existing `checkedFunction(t, fixture)` helper (`pathoracle_test.go:29-42`) which already runs fixtures through the real `syntax.Parse → check.Program` pipeline — do not hand-build a `core.Function`.

---

### `internal/compiler/interp/interp.go` (service, event-driven)

**Analog for the frame stack shape:** `internal/compiler/callgraph/callgraph.go:277-292, 283-378` (`stackFrame` type + `Order`'s iterative DFS).

**Iterative explicit-stack pattern to mirror** (`callgraph.go:277-292`):
```go
// stackFrame is one entry of the DFS's own explicit stack: the node's own
// ID and the index of the next adjacency-list child still to visit. Using
// an explicit stack of these, rather than a Go call stack, is what makes
// the traversal cost no Go stack (D-07-17/D-07-18).
type stackFrame struct {
	id             string
	nextChildIndex int
}
func Order(program core.Program) ([]string, error) {
	// ...
	stack := []stackFrame{{id: root}}
	for len(stack) > 0 {
		top := &stack[len(stack)-1]
		// push: stack = append(stack, stackFrame{id: child})
		// pop:  stack = stack[:len(stack)-1]
	}
}
```
`interp`'s `[]frame` stack follows this shape but each element carries execution state (`values map[string]string`, live-resource tracking) rather than mere traversal position. D-10-21/22: this must be an explicit heap-allocated slice with a `for {}` driver loop — Go native recursion is architecturally disqualified (Go's stack-exhaustion `throw` cannot be `recover()`'d).

**Analog for the three `OpCall` arms to replace:** itself, all three currently stub-return `ErrCallUnsupported`.

**Current stub shape (all three arms today)** — `interp.go:156-158` (`runBranchArm`):
```go
case core.OpCall:
	// D-07-39: recognized, never faked. See ErrCallUnsupported.
	return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, ErrCallUnsupported)
```
— `interp.go:341-343` (`runLinearBlocks`): identical shape.
— `interp.go:454-464` (`runLinear`): identical shape, plus the existing `opCallGroupedArmForTest` mutation-kill seam:
```go
case core.OpCall:
	if opCallGroupedArmForTest {
		// D-07-42 Test 3 mutation: fold into OpCopy's grouped behaviour...
		values[operation.TargetID] = value
		events = append(events, ownedEvent(function, operation, "value.copied"))
		continue
	}
	return Execution{}, fmt.Errorf("operation %q: %w", operation.ID, ErrCallUnsupported)
```
All three must gain real push-frame-and-continue behavior, sharing ONE small in-package frame-partition helper (D-10-39 — permitted because all three arms are the same peer/package/derivation method, D-09-38 forbids only cross-peer helpers).

**`Run`'s existing single-frame entry to become the stack base** (`interp.go:37-50`):
```go
func Run(program core.Program, functionName, input string) (Execution, error) {
	validated := corevalidate.Validate(program) // fail-closed precondition; OWN-05b guard:
	                                              // never read validated's ownership-bearing fields
	if !validated.Valid {
		return Execution{}, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program = validated.Program()
	function, ok := findFunction(program, functionName)
	// ...
}
```
D-10-36: keep the `corevalidate.Validate` import/call (fail-closed posture), add a guard test asserting `interp` never calls `validated.PeerSignatures()`/`validated.PeerSiteCoverage()` (grep confirms zero hits today — convert "happens not to" into "cannot without a red test").

**`MaxCallDepth` constant — the inversion to write into the rationale comment (D-10-23):**
```go
// MaxPaths (pathoracle.go:46) sits ABOVE its real reachable maximum (64) so
// it never fires today. MaxCallDepth = 128 must sit BELOW the real
// structural ceiling (1024, from syntax/parser.go:16's maxFunctions,
// combined with callgraph's cycle refusal making every admitted call graph
// a DAG) so that it DOES fire on a legitimately generated program —
// otherwise the refusal is decorative. This direction is deliberately
// OPPOSITE pathoracle.MaxPaths; a reader who pattern-matches the two will
// draw the wrong conclusion.
const MaxCallDepth = 128
```

**Refusal-as-Outcome pattern (D-10-25):** model the depth-exceeded refusal like `OpDefect`'s existing Outcome-producing arm (`interp.go`, `case core.OpDefect:` in each execution path) — a typed `Outcome`/`Event` serializable into `CanonicalBytes`, never a bare `error`.

**Determinism discipline to assert explicitly (D-10-26):** `liveResourceList(live, liveOrder)` — the existing ordered-slice-not-map-range pattern already used for single-frame output — must be the model for any new per-frame drain accounting; write an explicit test proving new code never ranges a map for output ordering.

---

### `internal/compiler/interp/interp_test.go` (test)

**Analog:** itself — `checkedCallBasicProgram` (`interp_test.go:30`) is the ONLY existing test function in the file (97 lines total today — confirms D-10-58's coverage-floor finding) and is the full-pipeline fixture pattern the depth-exceeded test must follow (D-10-24): build a genuine `.lang` source, run `syntax.Parse → check.Program → corevalidate.Validate → interp.Run` — never a hand-built `core.Program` bypassing `corevalidate.Validate`.

**Seam pattern to copy for `frameDrainOrderForTest`** (D-10-35) — mirror `opCallGroupedArmForTest`'s existing shape (`interp.go:23-31`):
```go
var opCallGroupedArmForTest = false
```
Unexported, false/nil by default, flipped only by the same-package test, asserting divergence via `interp.CanonicalBytes`/`execution.Equal` — never `frame.liveResources` or map contents directly.

**Subprocess probe (D-10-42/43/44) — NO in-repo analog exists** (confirmed: `grep -rn "os/exec" internal/` returns zero hits repo-wide). Use Go's standard `TestHelperProcess` idiom (re-exec the test binary env-guarded) + `runtime/debug.SetMaxStack` in the child for a deterministic host ceiling; parent asserts exit status/stderr. This is new machinery for this repo — document it as such in the test's own doc comment, since there is no shipped precedent to cite.

---

### `internal/compiler/corevalidate/corevalidate.go` (service, CRUD-like accessor)

**Analog:** `internal/compiler/pathoracle/pathoracle.go`'s already-exported `RecomputeEndpoints` (same shape, sibling package).

**Existing unexported method to promote** (`corevalidate.go:1360-1366`, `(*validator).recomputeLoanEndpoints`):
```go
func (v *validator) recomputeLoanEndpoints(function *core.Function) []core.LoanEndpoint {
	linear := function.Linear
	if len(linear.Blocks) == 0 { return nil }
	// ... blockReach-based closure, mathematically forced to agree with the
	// producer on an acyclic graph.
}
```
D-10-52: add an exported `Result`-level accessor (mirroring `pathoracle.RecomputeEndpoints`'s already-exported signature shape) so criterion 4's differential can consume `core.LoanEndpoint` from all three static peers uniformly.

**Existing per-function-declaration check shape to mirror for the D-10-33 drop-order invariant** (`corevalidate.go:2155-2170`, the `derivePeerSignature` area):
```go
parameterContract := core.ParameterContract{
	ID: function.Parameter.ID, Name: function.Parameter.Name, Type: function.Parameter.Type,
	// D-07-01: today's grammar has exactly one parameter form.
	Mode:  "owned",
	Drops: hasDropAbility && !peerParameterEscapesOwned(function),
}
```
This is the existing "checked once per function declaration, not re-derived by the caller per call site" shape (D-10-33 requires the same discipline for the new "callee's frame drained before pop" invariant).

---

### `internal/compiler/session/session_peer_gate_test.go` (test, batch)

**Analog:** itself — `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (lines 226-279), `peerDivergenceExpected` (line 23).

**Existing bare-map register to upgrade** (`session_peer_gate_test.go:23`):
```go
var peerDivergenceExpected = map[string]string{
	// path -> Problems[0].Code, with inline comments naming the debt item
	// and closure status per entry (existing informal discipline)
}
```
D-10-54: upgrade each entry to a struct carrying a debt-register ID and landing phase, with a companion test asserting every ID resolves to a still-open `PHASE-NN-DEBT.md` entry.

**Existing corpus-walk pattern to extend, not duplicate** (`session_peer_gate_test.go:226-279`):
```go
func TestNoUndeclaredCheckPeerDivergenceAcrossCorpus(t *testing.T) {
	root := testsupport.ProjectPath("testdata")
	found := make(map[string]string)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		// ...
		checked, checkErr := session.CheckFile(path)
		if len(checked.Diagnostics) > 0 { return nil }
		validated := corevalidate.Validate(checked.Program)
		if validated.Valid { return nil }
		// ...
	})
	for path, wantCode := range peerDivergenceExpected { /* both-directions assert */ }
	for path, gotCode := range found { /* undeclared-divergence assert */ }
}
```
D-10-51: extend this exact walker to also compute `pathoracle`'s endpoint set and (on accept) `interp`'s metamorphic execution comparison — never build a second harness. Bill it everywhere as "three-way on refuse, four-way on accept" (D-10-53), never unqualified "four-way."

**Seeded-fault pattern (D-10-55), per D-09-25's precedent** — an unexported, package-scoped seam inside `pathoracle` (and later `interp`'s frame teardown) reachable only by that package's own test, plus a companion assertion that the OTHER peers' independently-derived endpoint sets stay byte-for-byte unchanged while only the seeded peer disagrees.

---

## Shared Patterns

### Narrow-map threading for cross-function widening
**Source:** `internal/compiler/corevalidate/corevalidate.go:1185-1213` (`buildLoanChainIndex` + `peerLoanCarry` param)
**Apply to:** `originvalidate.go`'s `walkReturnOrigin` widening (D-10-02/03) and `pathoracle.go`'s `RecomputeEndpoints` widening (D-10-13) — both independently decided but deliberately the same shape: a precomputed `map[calleeID]<minimal fact>` built once at the top-level entry point, threaded as an explicit parameter, never `core.Program`.

### Iterative explicit-stack traversal (never native Go recursion)
**Source:** `internal/compiler/callgraph/callgraph.go:277-292, 283-378` (`stackFrame`, `Order`)
**Apply to:** `interp.go`'s new `[]frame` call stack (D-10-21/22) — the only in-repo precedent for "traverse without consuming Go stack," and its own doc comment (lines 25-26) states the rationale to reuse verbatim.

### Typed, capped, fail-closed refusal (never truncation)
**Source:** `internal/compiler/pathoracle/pathoracle.go:33-46, 87-110` (`MaxPaths`, `pathCapError`)
**Apply to:** every new bound this phase introduces — `interp.MaxCallDepth` (with D-10-23's inverted-direction rationale) and `pathoracle`'s new composition-depth cap (D-10-12, its own typed error, never folded into `MaxPaths`).

### Nil-default unexported injection seam
**Source:** `internal/compiler/pathoracle/pathoracle.go:59-64` (`TerminatorKindsOverride`); `internal/compiler/interp/interp.go:23-31` (`opCallGroupedArmForTest`); `internal/compiler/check/check.go:701` area (`deriveFunctionUsesParamObserved`)
**Apply to:** D-10-08's gate seam, D-10-35's `frameDrainOrderForTest`, D-10-41's mutant pairing, D-10-42/43's `maxCallDepthOverride`, D-10-55's seeded-fault seams — every fault-injection point this phase adds. Never a production-mutable exported global.

### Direct-import guard scan (`go/parser`, `ImportsOnly` mode)
**Source:** `internal/compiler/originvalidate/originvalidate_test.go:33-79`; `internal/compiler/pathoracle/pathoracle_test.go:51-75` (the more complete 5-item forbidden list); `internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go:95-112`
**Apply to:** every guard extension this phase makes. Pattern: `os.ReadDir` the package dir, skip `_test.go` files, `parser.ParseFile(..., parser.ImportsOnly)`, check each import's path suffix against a forbidden list, `t.Fatalf` on match. D-10-18: do NOT consolidate these into one shared table — the existing redundancy (4+ near-identical guards) is a real defense, not duplication debt.

### Corpus-walk differential extension
**Source:** `internal/compiler/session/session_peer_gate_test.go:226-279` (`TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`)
**Apply to:** criterion 4's three-way-on-refuse/four-way-on-accept differential (D-10-51) — extend this exact function/walker, never build a parallel harness.

## No Analog Found

Files/subsystems with no close match in the codebase (planner should use RESEARCH.md's Go-stdlib guidance instead):

| File/Subsystem | Role | Data Flow | Reason |
|---|---|---|---|
| Pitfall-4 subprocess stack probe (new test in `interp_test.go`) | test | subprocess integration | No `os/exec`, `SetMaxStack`, or `TestMain` pattern exists anywhere in this repo today (`[VERIFIED]` via `grep -rn "os/exec" internal/` = zero hits). Use Go's standard `TestHelperProcess` idiom (stdlib-documented, not repo-precedented) plus `runtime/debug.SetMaxStack` (D-10-42). |
| `testdata/phase10/interp_oracle/` golden corpus + regeneration procedure | test fixture (golden) | file I/O, byte-for-byte | No existing golden-corpus directory/discipline in this repo. Model on the general shape of other `testdata/phaseNN/` fixture directories for naming, but the golden-byte-freeze + reviewed-regeneration-commit discipline (D-10-57) is new this phase. |
| `pathoracle`'s call-composition function (new code, not merely a widened signature) | service | transform (recursive) | This is D-10-09's one irreversible new design commitment — there is no existing cross-function composition anywhere in `pathoracle`; the analog is `pathoracle`'s own `EnumeratePaths` applied recursively to itself, not a separate package's precedent. |

## Metadata

**Analog search scope:** `internal/compiler/{originvalidate,pathoracle,interp,corevalidate,callgraph,check,session}`, `testdata/{phase07,phase08}/`, `.planning/phases/09-peer-re-derivation-and-d-03-02-closure/`
**Files scanned:** 14 target files + 9 analog source files read directly this session (originvalidate.go, originvalidate_test.go, pathoracle.go, pathoracle_test.go, interp.go, corevalidate.go, corevalidate_peer_liveness.go, callgraph.go, session_peer_gate_test.go, check.go, corevalidate_endpoint_internal_test.go)
**Pattern extraction date:** 2026-09-11
