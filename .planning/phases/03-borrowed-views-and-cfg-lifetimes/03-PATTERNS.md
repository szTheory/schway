# Phase 3: Borrowed Views and CFG Lifetimes — Pattern Map

**Mapped:** 2026-09-03
**Files analyzed:** 9 new-work areas from `03-CONTEXT.md` (D-01..D-15) + ROADMAP Phase 3 criteria 1–4
**Analogs found:** 9 / 9 (all exact or role-match; no area lacks an in-tree analog)
**Codebase:** 8,771 lines Go across 22 packages — small enough that every analog below was read in full.

All analog paths verified git-tracked (`git ls-files`). No gitignored mirrors.

---

## Mapping Table

| New work | Closest analog | file:line | What to copy | What to change |
|---|---|---|---|---|
| CFG liveness pass (OWN-03) | `analyzeStraightLine` + `discoverLoanLastUses` | `internal/compiler/check/check.go:190-344`, `:366-399` | The `ownershipSupport` return struct, the counted `result.Work++` per step, the `fail(...)` closure, the `endLoans(index)` expiry hook, cause+repair diagnostic construction | Replace the single forward `for index, binding := range body.Bindings` walk with a block/edge worklist; `loanUse{index,span}` becomes an endpoint (`point:` or `edge:`); count propagation in `Work` (D-05) |
| Exclusive loan / access-mode admission in the independent layer | `validator.replay` | `internal/compiler/corevalidate/corevalidate.go:226-310` | `v.checks++` before every dispatch, the `switch operation.Kind` authorization table, `v.check(ok, code, detail)` fail-closed idiom, first-problem-only accumulation | Add `core.OpBorrowExclusive` case + conflict rule; keep the loan-liveness derivation **structurally different** from the checker's (D-12) |
| New core fact/record (loan origin, access mode) | `core.Place` / `core.TypeFact` / `core.LinearOperation` | `internal/compiler/core/core.go:63-93` | Flat struct, `json:` tags, `omitempty` only on optional fields, ID as `fmt.Sprintf("%s:<kind>:%d", functionID, ordinal)` | Every new ID must be a **function-local semantic ordinal**, never a source offset; add a matching `core.<kind>_order` validator rule |
| New schema version `/2` | Phase 2's `/1` introduction | `core.go:5-8`, `execution.go:8-11`, `diagnostic.go:11-14`, `evidence.go:29-37`, `check.go:64`, `corevalidate.go:70/106/109`, `evidence.go:207-245`, `evidence.go:371-410` | Add constants **alongside**; upgrade at exactly one admission point; add a *new* identity struct in `manifestID` rather than editing an existing one | Phase 3 upgrade point is the new CFG/origin admission, not `check.go:64`; add a third `manifestID` branch; extend the body↔schema cross-lock |
| New negative control | `control:ownership.move_while_reborrowed` | fixture `testdata/phase2/reborrow_while_moved.lang`; table `session.go:596-601`; lane `session.go:602-621`; required set `session.go:690-700`; zero-work gate `session.go:706-710`; CLI `cli_test.go:243-261`; unit `session_test.go:731-761`; script `scripts/verify-phase2.sh:15` | The five-place wiring: fixture → control table row → lane with nonzero work → `requiredControls` entry → CLI substring assertion. All five or the control is not real. | New corpus dir `testdata/phase3/`; new lane IDs; a `verifyBorrowedCorpus` dispatch alongside `verifyOwnedCorpus` (`session.go:426-428`) |
| New independent test oracle | `oracleLoanLastUses` | `internal/compiler/check/check_test.go:254-306` (+ doc comment `:254-263`) | Materialize the derivation relation as an explicit edge set, close by order-independent fixed-point, reduce afterwards — a method production never uses. Plus `oracleResolveBinding` (`:410-423`) resolving names *backward* where production resolves forward. | CFG oracle = bounded path expansion (spike 002's method), not a second worklist. Mutation-kill it (D-09) before it counts. |
| New exhaustive / property test | `TestOwnershipSequenceExhaustive` (`check_test.go:14-97`), `TestPairAllAbilityMasks` (`ability_test.go:56-91`), `TestGeneratedLinearRoundTrips` + `generatedLinearProgram` (`syntax_test.go:353-459`) | Full alphabet enumeration to a bounded length; `semanticTokens` token-identity assertion (`syntax_test.go:540-549`); fuzz seeds naming the combination fixtures never reach (`syntax_test.go:266-269`) | Generators must reach **branching** shapes; write down, in a comment, what the generator does *not* reach (D-10) |
| New phase gate script | `scripts/verify-phase2.sh` + `scripts/assert-go-tests.sh` | `verify-phase2.sh:1-49`, `assert-go-tests.sh:4-42` | build-once to `$verify_tmp/lang`; `--self-test` exact-target discovery; `observe()` 20-sample warm loop with p50/p95/min/max/bytes/work; nonzero-work assertion | **Add `scripts/verify-phase3.sh`**, do not extend phase2 — see §8 |
| Bounded process / IO | `native.boundedWriter` (`native.go:137-159`), `evidence.boundedProbeWriter` (`evidence.go:121-159`), `readBoundedFile` (`session.go:394-401`) | max-plus-one buffer, `total` counter independent of buffer, separate stdout/stderr writers, `context.WithTimeout` + `exec.CommandContext`, never `CombinedOutput` | Prefer the `native.boundedWriter` shape (it has `overflowed()`); do not add a **third** copy — see Inconsistency I-2 |

---

## Per-Area Detail

### 1. A new checker analysis pass (OWN-03 CFG liveness)

**Analog:** `internal/compiler/check/check.go`

**Structure of the current straight-line pass.** `Program` (`:20-128`) walks functions; for a linear body it delegates to `checkLinear` (`:130-158`), which builds the `core.LinearBody` skeleton and then calls the analysis at a **single seam**:

```go
// check.go:147-152
support := analyzeStraightLine(functionID, function.Parameter.Name, function.Parameter.Span, linear.Types[0], function.Body.Linear)
if support.Diagnostic != nil {
    return core.Function{}, []diagnostic.Diagnostic{*support.Diagnostic}, support.Work
}
linear.Places = support.Places
linear.Operations = support.Operations
```

**This is the seam a CFG pass slots into.** `analyzeStraightLine` is a pure function `(ids, name, span, TypeFact, *ast.LinearBody) -> ownershipSupport`. A CFG pass keeps that signature and that return type; nothing above it changes. `ownershipSupport` (`:172-180`) already carries `Places`, `Operations`, `LoanFinalUses`, `States`, `Work`, `DiagnosticCode`, `Diagnostic` — Phase 3 extends the struct (loan endpoints keyed by point-or-edge) rather than replacing it.

**Counted work accounting.** Two contributions only:

```go
// check.go:195 — seed
Work: typeNodeCount(typeFact.Shape) + len(body.Bindings) + 1,
// check.go:240 — one per binding
result.Work++
// check.go:328 — one for the return claim
result.Work++
```

`TestOwnershipWorkSeries` (`check_test.go:142-158`) pins `wantWork == 1 + 2*(operations+1)` and asserts `got.Work <= 2*len(got.Operations)+1`. **D-05 requires breaking this bound deliberately**: `discoverLoanLastUses` (`:366-399`) propagates transitive loan sets and counts *nothing*, which is exactly why `recomputed_work` understates the Θ(N²) cost (D-02-03). The CFG pass must count worklist transfer evaluations and re-insertions — the spike already models this (`Analysis.TransferEvaluations`, `Analysis.WorklistReinsertions`, `.planning/spikes/002-cfg-edge-last-use/cfg/model.go:62-63`). Update `TestOwnershipWorkSeries`'s bound in the same commit, with the new bound justified.

**`placeState` bookkeeping** (`:182-188`): `place`, `declared` span, `initialized` flag, `movedAt *Span`, `moveTargetID`. Keyed by *source name* in `places map[string]*placeState` (`:205-207`) so shadowing works by rebinding (`:320`). A CFG pass needs this per-block-entry, joined at merges — the `initialized` bit becomes a lattice element, and `movedAt`/`moveTargetID` become the *witness* selected at the join (pick a deterministic one, or the diagnostic loses determinism).

**`loanState` bookkeeping** (`:351-357`): `id`, `ownerID`, `borrowedAt`, `lastUse int`, `lastUseSpan`. Two indices maintained in lockstep:

```go
// check.go:208-219
activeLoans := make(map[string]map[string]*loanState)   // ownerID -> loanID -> loan
expiringLoans := make(map[int][]*loanState)             // operation index -> loans ending there
endLoans := func(index int) { ... delete from activeLoans ... }
```

`endLoans(index)` is called after each operation (`:325`) and after the return (`:341`). **This is the hook that generalizes to edges**: `expiringLoans` becomes keyed by endpoint identity (`point:<block>:<n>:<loan>` / `edge:<from>:<to>:<loan>`, spike 002's spelling), and `endLoans` fires on block exit *per successor edge*.

**Diagnostic emission with causes + repairs.** Three sites, all the same shape (`:225-238` `useAfterMove`, `:260-270` `move_while_borrowed`, `:283-293` `borrow_requires_share`, `:306-317` `transfer_requires_take`). The canonical form:

```go
// check.go:260-270
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

Copy exactly: span-bearing causes first, then ID-bearing detail causes (`place`/`loan`/`owner`/`type`), then repairs. Determinism comes from sorting loan IDs and taking the first (`:255-259`) — a CFG pass must sort blocking loans by ID the same way or two runs disagree.

`check_test.go:65-71` asserts the *cause-kind sequence* for the borrow gate. Any Phase 3 diagnostic should get the same style of assertion.

**Repair-bearing = schema `/1`.** `diagnostic.ErrorWithRepairs` (`diagnostic.go:58-83`) stamps `Schema1` and folds sorted *repair kinds* (not details, not prose) into the ID hash. `diagnostic.Error` (`:43-53`) stamps `/0`. Phase 3 diagnostics that carry repairs must use `ErrorWithRepairs`; plain `Error` silently lands in `/0` (see Inconsistency I-4).

---

### 2. A new independent validator rule (`corevalidate`)

**Analog:** `internal/compiler/corevalidate/corevalidate.go`

**`replay` structure** (`:226-310`) is two passes:

*Pass 1 — derive loan liveness* (`:237-258`):

```go
for index, operation := range operations {
    v.checks++ // inspect each operation once while finding final loan uses
    carried := append([]string(nil), loansForPlace[operation.SourceID]...)
    for _, loanID := range carried { loanLastUse[loanID] = index }
    if operation.Kind == core.OpBorrowShared {
        loanOwner[operation.LoanID] = operation.SourceID
        loanLastUse[operation.LoanID] = index
        carried = append(carried, operation.LoanID)
    }
    if operation.Kind != core.OpReturn && len(carried) > 0 { loansForPlace[operation.TargetID] = carried }
}
ownerBlockedUntil := ...  // :252-258, one v.checks++ per declared loan
```

*Pass 2 — authorize each transition* (`:261-308`): source type match, `initialized` check, then

```go
v.checks++ // dispatch one independently authorized transition
switch operation.Kind {
case core.OpCopy:          // requires AbilityCopy   :271-279
case core.OpBorrowShared:  // requires AbilityShare  :280-288
case core.OpMove:          // blockedUntil < index   :289-299
case core.OpReturn:        // last, no target        :300-304
default: return v.check(false, "core.unknown_operation", string(operation.Kind))  // :305-306
}
```

The `default:` arm is the fail-closed anchor: **a new operation kind is rejected until it is explicitly authorized here.** That is the mechanism that makes adding `core.OpBorrowExclusive` safe — a checker that emits it before the validator learns it produces a hard `core.unknown_operation`, not silent admission.

**`v.check` idiom** (`:58-67`):

```go
func (v *validator) check(ok bool, code, detail string) bool {
    v.checks++
    if ok { return true }
    if len(v.problems) == 0 { v.problems = append(v.problems, Problem{Code: code, Detail: detail}) }
    return false
}
```

Every caller is `if !v.check(...) { return false }`. Only the **first** problem is recorded, so `Problems[0].Code` is the stable assertion target used everywhere (`session.go:642`, `session.go:187`, `evidence.go:196`). Copy this exactly; do not accumulate.

**Work counting.** `v.checks++` fires in `check` (`:59`), `unique`/`uniqueType`/`uniquePlace` (`:430/448/462`), `derive` (`v.checks += nodes`, `:324`), and the two explicit `v.checks++` in `replay` (`:238`, `:254`, `:269`). `LinearWorkLimit(facts) = 16*facts + 13` (`:24`) is the declared bound for the canonical scale shape — **a new per-operation check changes this constant and its comment must be re-derived, not bumped.**

**Independent re-derivation + `reflect.DeepEqual`.** This is the load-bearing independence mechanism (`:163-169`):

```go
abilities, witnesses, ok := v.derive(fact.Shape, 0)
if !v.check(reflect.DeepEqual(fact.Abilities, abilities) && reflect.DeepEqual(fact.NegativeWitnesses, witnesses), "core.ability_mismatch", fact.ID) { return false }
```

`v.derive`/`deriveAbility` (`:322-427`) is a **second, separately written** implementation of the same law as `ability.combineStructural` — note the duplicated OV-02-01 Buffer-grants-share prose at `corevalidate.go:390-392` and `ability.go:120-122`. `control:ability.forged_copy` (`session.go:636-645`) is the falsifier.

**Phase 3 obligation (D-12):** an access-mode or origin fact must be re-derived here by a materially different route from the checker's. Where that is impossible (as with OV-02-01), say so explicitly and record that the differential cannot cross-check that row.

---

### 3. A new core fact/record type

**Analog:** `internal/compiler/core/core.go`

```go
// core.go:70-74
type Place struct {
    ID     string `json:"id"`
    Name   string `json:"name"`
    TypeID string `json:"type_id"`
}
// core.go:85-93
type LinearOperation struct {
    ID       string        `json:"id"`
    PointID  string        `json:"point_id"`
    Kind     OperationKind `json:"kind"`
    SourceID string        `json:"source_id"`
    TargetID string        `json:"target_id,omitempty"`
    LoanID   string        `json:"loan_id,omitempty"`
    TypeID   string        `json:"type_id"`
}
// core.go:63-68
type TypeFact struct {
    ID string; Shape TypeRef; Abilities []Ability; NegativeWitnesses []AbilityWitness
}
```

**ID conventions — function-local semantic ordinals, never source offsets:**

| Record | Producer | Format |
|---|---|---|
| module / type / fn | `semanticID` `check.go:480` | `s1:<module>:<kind>:<name>` |
| type fact | `check.go:139` | `<functionID>:type:<n>` |
| place | `check.go:140`, `:248` | `<functionID>:place:<n>` |
| operation | `check.go:322` | `<functionID>:op:<n>` |
| point | `check.go:322` | `<functionID>:point:linear:<n>` |
| loan | `check.go:201`, `:297` | `<functionID>:loan:<n>` |
| entry/return point | `check.go:154` | `<functionID>:point:entry` / `:point:return` |
| match arm / edge | `check.go:100-101` | `<functionID>:arm:<n>` / `<matchID>:edge:<pattern>` |

Note `MatchArm.EdgeID` (`core.go:112`) — **an edge identity already exists in core**, keyed by a semantic discriminator (`pattern`), not a number. A CFG edge ID should follow that precedent: `<functionID>:edge:<fromBlock>:<toBlock>`.

**What adding a loan-origin or access-mode record must satisfy — five obligations:**

1. **Ordinal-positional validation.** Every ordinal-bearing slice has a `core.<kind>_order` rule: `core.type_order` (`corevalidate.go:160`), `core.place_order` (`:177`), `core.operation_order` (`:201`). A new slice needs its own.
2. **Uniqueness.** A `v.unique(...)`/`v.uniqueXxx(...)` pass with a `core.duplicate_<kind>_id` code (`:429-473`).
3. **Referential closure.** Every ID referenced must resolve into a map built earlier in `linear` (`:204-213`) — `core.unknown_place`, `core.unknown_type`, `core.unknown_loan`.
4. **JSON clone-round-trip stability.** `cloneProgram` (`:484-494`) marshals and unmarshals; `Result.Program()` (`:40`) returns a clone. Any field that does not survive `json.Marshal`→`Unmarshal` identically breaks `reflect.DeepEqual`-based controls. Slices must be non-nil-initialized where the checker emits them (`check.go:146`, `ability.go:82`) or `[]` vs `null` diverges.
5. **Evidence digest impact.** `CoreDigest` (`evidence.go:212`) hashes `json.Marshal(checked.Program)`. Adding a field with `omitempty` and leaving it empty on Phase 1/2 programs is what keeps `testdata/phase1/evidence.golden.json` and `testdata/phase2/evidence.golden.json` byte-identical (D-13). **Without `omitempty`, both goldens move.**

---

### 4. A new schema version

**How Phase 2 introduced `/1` alongside a frozen `/0` — the exact mechanism, in five parts:**

**(a) Constants added, never mutated.**

```go
core.go:5-8        Schema = "lang.core/0";      Schema1 = "lang.core/1"
execution.go:8-11  Schema0 = "lang.execution/0"; Schema1 = "lang.execution/1"
diagnostic.go:11-14 Schema = "lang.diagnostic/0"; Schema1 = "lang.diagnostic/1"
evidence.go:29-31  Schema0 = "lang.evidence/0";  Schema1 = "lang.evidence/1"; Schema = Schema0
```

**(b) One upgrade point, gated on the feature actually being admitted.** `check.Program` initialises `Schema: core.Schema` (`check.go:21`) and upgrades at exactly one line:

```go
// check.go:59-67 — inside the linear branch, and only after diagnostics are empty
if len(diagnostics) == 0 {
    result.Program.Schema = core.Schema1
    result.Program.Functions = append(result.Program.Functions, checked)
}
```

A match-only program never executes line 64, so its core bytes are literally the Phase 1 bytes. `core.mixed_body_versions` (`check.go:30-41`) forbids the mixed case outright — **that guard is what makes "one schema per program" true**, and it is the thing Phase 3 must extend or deliberately relax.

**(c) A fail-closed body↔schema cross-lock in the independent layer.**

```go
// corevalidate.go:70   accept either
if !v.check(v.program.Schema == core.Schema || v.program.Schema == core.Schema1, "core.schema", ...)
// corevalidate.go:106  linear body REQUIRES /1
if !v.check(v.program.Schema == core.Schema1, "core.schema", "linear body requires lang.core/1") || !v.linear(function)
// corevalidate.go:109  match body REQUIRES /0
} else if !v.check(v.program.Schema == core.Schema, "core.schema", "match body requires lang.core/0") || !v.match(...)
```

Bidirectional: a `/1` program with a match body is rejected as firmly as a `/0` program with a linear body.

**(d) Evidence upgrades as a whole block, guarded on the core schema.** `evidence.build` writes `/0` defaults (`:207-213`) then:

```go
// evidence.go:215-224
if checked.Program.Schema == core.Schema1 {
    manifest.Schema = Schema1
    manifest.CoreSchema = core.Schema1
    manifest.ExecutionSchema = execution.Schema1
    manifest.DiagnosticSchema = diagnostic.Schema1
    manifest.DigestClaim = DigestClaim
    manifest.KnownEscape = corevalidate.KnownEscape
    if manifest.Policy == "phase1-pure-c17-v1" { manifest.Policy = "phase2-owned-c17-v1" }
    ...
}
```

Note the policy rename is itself guarded, and the `/1`-only manifest fields all carry `json:",omitempty"` (`evidence.go:56,65,66,67`) so a `/0` manifest serialises byte-identically to Phase 1.

**(e) The identity function forks rather than grows.** `manifestID` (`evidence.go:371-410`) has **two separate anonymous identity structs**: `/1` at `:373-385`, `/0` at `:387-406`. The `/0` struct was not touched when `/1` was added — that is precisely why Phase 1 evidence IDs are byte-frozen. `evidence.Validate` mirrors this: shared checks at `:314-328`, `/1`-only checks appended conditionally at `:329-335` and `:344-346`.

**Phase 3 prescription.** Add `core.Schema2 = "lang.core/2"` (and peers) as new constants. Upgrade at exactly one new admission point. Add a **third** identity struct in `manifestID` — do not edit either existing one. Extend the `corevalidate` cross-lock with the `/2`-requires-CFG-body rule. Every new manifest/core field gets `omitempty`. If a Phase 2 golden must move, D-13 requires a field-by-field explanation in the plan.

---

### 5. A new negative control — full trace of `control:ownership.move_while_reborrowed`

This is the template. Five wiring points; all five are required.

**(i) Fixture** — `testdata/phase2/reborrow_while_moved.lang`:

```
// A reborrow keeps the original loan live: loan liveness is transitive.
module owned.reborrow_while_moved
export { fn relay }
fn relay(code: Byte) -> Byte {
  let view = borrow code
  let review = borrow view      // reborrow — the hop that expires the loan early if liveness is not transitive
  let delivered = take code
  let observed = review
  delivered
}
```

Note the leading comment states the *law under test*, and the fixture is the minimal shape that distinguishes it. Fixtures must be canonical (`syntax_test.go:168-171` asserts `Format(Parse(src)) == src` for named phase-2 fixtures) and stay under `syntax.MaxSourceBytes`.

**(ii) Control table + lane** — `session.go:596-621`:

```go
ownershipControls := []struct{ file, code, control string }{
    {"use_after_move.lang",        "ownership.use_after_move",        "control:ownership.use_after_move"},
    {"move_while_borrowed.lang",   "ownership.move_while_borrowed",   "control:ownership.move_while_borrowed"},
    {"implicit_noncopy.lang",      "ownership.transfer_requires_take","control:ownership.transfer_requires_take"},
    {"reborrow_while_moved.lang",  "ownership.move_while_borrowed",   "control:ownership.move_while_reborrowed"},
}
```

Note row 4: **the control name differs from the diagnostic code** — the control names the *law*, the code names the *diagnostic*. Copy that. The loop reads each fixture bounded, enforces the byte ceiling, and asserts exactly one diagnostic of the exact code:

```go
// session.go:617-619
checked := Check(source)
if len(checked.Diagnostics) != 1 || checked.Diagnostics[0].Code != control.code {
    return fail(protocol.StatusInvalid, "verify.control_missing", control.control)
}
...
addLane("lane:owned-negative-controls", ownershipControlNames, len(ownershipControls), ownershipBytes, laneStarted)
```

`len(...) != 1` is load-bearing: it forbids the control passing incidentally alongside another diagnostic.

**(iii) `requiredControls` fail-closed set** — `session.go:690-705` (now nine entries), plus the zero-work gate:

```go
// session.go:706-710
for _, lane := range result.Lanes {
    if lane.RecomputedWork == 0 { return fail(protocol.StatusInvalid, "verify.zero_work", lane.ID) }
}
```

**(iv) CLI assertion through the shipped binary** — `cli_test.go:243-261` (`TestVerifyPhase2CLI`) builds the binary, runs `--json verify testdata/phase2`, and substring-asserts all nine control IDs in stdout. Mirrored in-process by `session_test.go:731-761` (`TestVerifyPhase2ControlsAndWork`), which additionally requires `lane.Status == "pass"`, `RecomputedWork != 0`, `PeakRSSStatus == "unavailable"` per lane.

**(v) Script gate** — `verify-phase2.sh:15`:

```sh
grep -q 'control:backend.runtime_causality' "$verify_tmp/phase2.json" || { echo "..."; exit 1; }
```

and `cli_test.go:312-326` (`TestPhase2VerifierScriptContract`) asserts the script's own text contains the required tokens — i.e. **the gate script is itself gated by a test.**

**Phase 3 additions:** new corpus `testdata/phase3/`, a `verifyBorrowedCorpus` selected by the same fixture-probe dispatch as `verifyOwnedCorpus` (`session.go:426-428`), plus new controls. Candidates implied by the ROADMAP criteria: `control:ownership.exclusive_conflict`, `control:ownership.edge_last_use_omitted` (criterion 2's "omitting that edge is detected" — model on spike 002's `OmitEdgeEnds` fault injection, `.../002-cfg-edge-last-use/cfg/model.go:46-50`), `control:origin.stale_summary`, `control:origin.impossible_summary`.

---

### 6. A new independent test oracle (what D-09 requires)

**Analog:** `oracleLoanLastUses`, `internal/compiler/check/check_test.go:254-306`.

The doc comment at `:254-263` *is* the pattern statement and should be copied verbatim in spirit:

> Production streams forward over the bindings once, carrying an inherited loan association from each source binding to the binding it produces, and never materializes the derivation relation. The oracle does the reverse: it first materializes the whole derivation relation as an explicit edge set, then closes each loan under that relation by order-independent fixed-point iteration (repeat until nothing new becomes reachable), and only afterwards reduces the reachable set to its maximum ordinal. Neither derivation can be obtained from the other by renaming, so a transitivity error in one cannot be mirrored by the other.

Concretely the oracle differs on **four independent axes**:

| Axis | Production (`check.go`) | Oracle (`check_test.go`) |
|---|---|---|
| Derivation relation | implicit, streamed (`:366-399`) | materialized `edges map[int][]int` (`:265-271`) |
| Closure | single forward pass | order-independent fixed point (`:280-294`) |
| Reduction | `use.index` updated in place | max-of-reachable computed last (`:295-299`) |
| Name resolution | forward environment `visible map[string]int` (`:368`) | backward nearest-prior scan `oracleResolveBinding` (`:410-423`) |

**Why it counts as independent:** `oracleStraightLine` (`:308-408`) rebuilds places, transitions, IDs, loan expiry, and state snapshots without calling *any* production helper except pure formatters — even `typeNodeCount` is duplicated as `typeNodeCountOracle` (`:459-465`). Comparison is whole-struct: `assertSupportEqual` (`:467-475`) `reflect.DeepEqual`s `Operations`, `LoanFinalUses`, `States`, plus `Work` and `DiagnosticCode`.

**How it is mutation-killed.** Three layers:

1. **Exhaustive drive** — 36-symbol alphabet × lengths 0..3 = 50,653 programs, twice (share-granting and share-withholding facts), `check_test.go:14-46`.
2. **Named witness** — the four-step reborrow witness at `:73-96` asserts the specific answer (`LoanFinalUses[0].OperationIndex == 2`) and then *removes the later read* and asserts the program becomes legal. That pair is the mutation kill: a checker that ignores the later use passes the first half and fails the second.
3. **Fuzz** — `FuzzOwnershipLinear` (`:160-173`) with seeds `{}`, `{2,1,3}`, `{1,0}`.

**The Phase 2 failure this exists to prevent (D-09):** WR-02. The original oracle encoded the *same* wrong law as production — non-transitive loan liveness — so the differential agreed on the wrong answer. The fix (`3d9493a`) had to change both, which is why the doc comment now explicitly argues non-derivability.

**Phase 3 requirement:** the CFG oracle must not be a second worklist. Use **bounded path expansion** — enumerate paths in the bounded CFG, lower each to a linear program, and reuse the already-independent straight-line oracle. That is exactly spike 002's validated method (`.planning/spikes/002-cfg-edge-last-use/README.md`, "Iteration 3 — independent path oracle and masking defense"), including its warning that a late terminal guard is needed so the linear normalizer cannot silently repair a missing CFG endpoint. **Then revert the production hunk and confirm the differential fails.** Not evidence until that is demonstrated.

---

### 7. A new exhaustive / property test

**Analog A — exhaustive mask enumeration** (`internal/compiler/ability/ability_test.go`):

- `fabricatedSet(mask)` / `observedMask(set)` (`:16-41`) — a bijection between an integer and the private domain type, so enumeration is a plain integer loop.
- `TestBoxAllAbilityMasks` (`:43-54`) — all 32 masks; also asserts the input was **not mutated**.
- `TestPairAllAbilityMasks` (`:56-91`) — all 32×32 = 1,024 pairs against an independently recomputed bitwise conjunction, then explicitly **falsifies four rival laws** (first-child, union, forward implication, reverse implication) using asymmetric cases `{1,0},{0,1},{5,18},{18,5}`. This is the strongest pattern in the repo: it is not enough that the law holds; the plausible wrong laws must be shown *not* to hold.
- `TestArbitraryMasksRemainTestPrivate` (`:201-234`) — parses `ability.go`, `check.go`, `core.go`, `ast.go` with `go/parser` and fails if production exports anything named `*mask*`/`*combine*`. An architectural invariant enforced as a test.

**What it reaches / does not reach.** `TestShareIsUniversallyGrantedAfterBufferShare` (`:175-199`) enumerates `{Byte,Buffer}` closed under `Box`/`Pair` to depth 4 (≥100 shapes) and records as an executable fact that **no source-reachable type withholds `share`** — therefore `ownership.borrow_requires_share` (`check.go:282-293`) is *unreachable from source today*. That gate is only tested via the synthetic `nonShareableTypeFact()` (`check_test.go:225-231`), whose comment says so. Phase 3's exclusive-borrow rules will very likely be in the same position; **name it in a comment the way this one does, or D-10 is violated.**

**Analog B — generated program property** (`internal/compiler/syntax/syntax_test.go`):

- `generatedLinearProgram(caseID)` (`:424-459`) — `{5 declared types} × {"", "take ", "borrow "} × {0..3 bindings}`, decoded from the case ID by successive modulus, with separator variation (space/tab, `caseID%2`) and comment placements at `caseID%3`, `%7`, and — critically — `%11` for **a comment trailing the declaration header** (`:443-445`), the one placement that defeated the brace classifier.
- `TestGeneratedLinearRoundTrips` (`:353-411`) — `quick.Config{MaxCount: 1000, Rand: rand.New(rand.NewSource(0x11ea40))}`, fixed seed, plus `if count != config.MaxCount` to catch a generator that silently stops.
- Assertions per case: byte-lossless CST, zero diagnostics, canonical reparse, formatting fixed point, `semanticTokens` identity, span-stripped AST identity, comment identity, and diagnostic-**code** sequence stability across canonicalisation (spans legitimately move — `:393-400`).
- `semanticTokens` (`:540-549`) projects every non-whitespace non-EOF token as `kind:text`. Comment: "Fusing two identifiers into one, or dropping one, changes this projection even when the result still parses." That is the assertion that catches formatter defects a reparse cannot.

**What these generators historically did NOT reach** (from `02-DEBT.md` "Process debt", lines 175-182) — three misses, all reachability, none coverage:

1. The syntax generator emitted only Phase 1 **match** programs, so the round-trip property could never reach the linear surface → missed CR-02. Fixed by adding `generatedLinearProgram` as a separate module shape (a linear body cannot share a module with a match body — `core.mixed_body_versions`, so it could not be an extra function).
2. The generator then emitted comments only *before* `module` and *before* a binding, never trailing a header → missed the CR-02 regression **past 1,000 generated cases**. Fixed by the `caseID%11` branch.
3. The ownership oracle encoded the same law as production → missed WR-02.

The corresponding manual patch is the fuzz seed comment at `syntax_test.go:264-268`: two hand-written seeds for "a generic type combined with a binding: the combination the shipped fixtures never reach."

**Phase 3 prescription.** `linearGeneratedKinds` (`:418`) must gain exclusive-borrow spellings, and the generator must reach **branching** bodies — a `{0..3 bindings}` linear generator can never produce a CFG edge, which is exactly the failure mode (1) above repeating. Before trusting any green Phase 3 property test, write down the reachable input space in a comment.

---

### 8. A new phase gate script — **recommendation: add `scripts/verify-phase3.sh`**

**Analog:** `scripts/verify-phase2.sh` (49 lines) + `scripts/assert-go-tests.sh` (68 lines).

**Structure of `verify-phase2.sh`:**

```sh
:4-6   verify_tmp=$(mktemp -d ...); trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM; export GOCACHE="$verify_tmp/go-cache"
:8     sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork
:9-11  go test ./... ; go test -race ./... ; go vet ./...
:12    go build -o "$verify_tmp/lang" ./cmd/lang        # build ONCE
:13-14 "$verify_tmp/lang" --json verify testdata/phase1 > phase1.json
       "$verify_tmp/lang" --json verify testdata/phase2 > phase2.json
:15    grep -q 'control:backend.runtime_causality' phase2.json || exit 1
:17-40 observe() { 20 warm samples; sed-extract elapsed_ns and recomputed_work; fail on zero; print p50/p95/min/max/output_bytes/work/peak_rss }
:42-46 observe format|check|interpreter|native|full_verify
:48-49 cat phase1.json phase2.json
```

**`assert-go-tests.sh` — exact fail-closed target discovery** (`:4-42`):

- rejects any name containing a non-`[A-Za-z0-9_]` character (`:16-21`) — no regex injection into `-run`;
- lists real targets via `go test "$package" -list .` (`:12`) and requires an **exact string match** (`:22-33`), erroring `target not discovered` otherwise — so a renamed test fails the gate instead of silently matching nothing;
- builds an anchored alternation `-run "^(A|B|C)$" -count=1` (`:41`);
- `--self-test` mode (`:44-62`) first asserts a **nonexistent sentinel** `TestCodenameLangSelectionGuardMustNotExist` is rejected, proving the selector itself is not vacuous.

**Bounded warm observation** (`verify-phase2.sh:17-40`): exactly 20 samples, `sort -n` then index 1/10/19/20 for min/p50/p95/max, `LANG_OBSERVE_TIMING=1` to enable per-command `elapsed_ns` (gated in `session.go:409-418` so ordinary output does not churn), and two fail-closed assertions — no timing (`:27`) and zero work (`:38`).

**Recommendation: a new `scripts/verify-phase3.sh`, not an extension of phase2.**

Reasoning, from how Phase 1 → Phase 2 was actually done:

1. `verify-phase1.sh` was **left in place unmodified** (11 lines) and `verify-phase2.sh` was added as a peer. Phase 1's gate still runs standalone.
2. `verify-phase2.sh` **does not invoke** `verify-phase1.sh` — and that is enforced: `cli_test.go:318` fails if the script text contains `"verify-phase1.sh"` or has more than one `go test ./...` / `go test -race ./...` / `go vet ./...`. Nesting the gates would double every suite run.
3. Instead of nesting, phase 2 **absorbed phase 1's corpus assertion**: line 13 runs `verify testdata/phase1` with the same binary. Non-regression is proven by running the older corpus, not by running the older script.

So: `verify-phase3.sh` runs the shared Go suites **once**, builds **once**, verifies `testdata/phase1`, `testdata/phase2`, and `testdata/phase3` with that one binary, greps the new Phase 3 control IDs, and observes the Phase 3 fixture. Add `TestPhase3VerifierScriptContract` mirroring `cli_test.go:312-326`, including the `strings.Contains(text, "verify-phase2.sh")` prohibition. `verify-phase2.sh` stays byte-stable (D-13 in spirit).

Caveat: the `observe()` function is now about to be copied a third time. Either factor it into `scripts/observe.sh` sourced by both, or accept the duplication explicitly in the plan — do not let it drift silently.

---

### 9. Bounded process / IO (D-15)

**Spawn pattern** — `native.Runner.Run` (`native.go:49-128`):

```go
:68-69   ctx, cancel := context.WithTimeout(parent, r.Timeout); defer cancel()   // default 5s, :47
:71      command := r.commandContext(ctx, r.ClangPath, "-std=c17","-Wall","-Wextra","-Werror","-pedantic", optimization, ...)
:72-74   var compileStdout, compileStderr boundedWriter                          // INDEPENDENT streams
         command.Stdout = &compileStdout; command.Stderr = &compileStderr        // never CombinedOutput
:77-79   if errors.Is(ctx.Err(), context.DeadlineExceeded) { return native.timeout }
:80-85   if compileStdout.overflowed() { ... } ; if compileStderr.overflowed() { ... }
:97      runCtx, runCancel := context.WithTimeout(parent, r.Timeout)             // fresh deadline PER INPUT
:103     runCancel()                                                             // explicit, not deferred in a loop
```

Deadline is checked **before** the error, so a timeout is reported as `native.timeout` rather than `native.compile_failed`. Order matters; copy it.

**Max-plus-one bounded writer** — `native.boundedWriter` (`:137-155`):

```go
func (w *boundedWriter) Write(data []byte) (int, error) {
    w.total += len(data)                                   // total counts EVERYTHING, unbounded counter
    remaining := MaxStreamBytes + 1 - w.buffer.Len()       // buffer holds at most limit+1
    if remaining > 0 { if remaining > len(data) { remaining = len(data) }; _, _ = w.buffer.Write(data[:remaining]) }
    return len(data), nil                                  // always claims full write: never stalls the child
}
func (w *boundedWriter) overflowed() bool { return w.total > MaxStreamBytes }
```

`MaxStreamBytes = 64 * 1024` (`:18`). The limit+1 buffer is what makes "exactly at the limit" distinguishable from "one over" without reading unbounded input.

**Tool-probe variant** — `evidence.boundedProbeWriter` (`:121-136`) + `runToolProbe` (`:138-159`): same law, `MaxToolProbeBytes = 64 * 1024` (`:36`), hard 5s deadline (`:139`), distinct codes `evidence.tool_timeout` / `evidence.tool_stdout_truncated` / `evidence.tool_stderr_truncated` / `evidence.tool_failed`, and the returned bytes are **copied out** (`:158`) so the caller cannot retain the writer's buffer.

**Bounded file read** — `session.readBoundedFile` (`:394-401`):

```go
return io.ReadAll(io.LimitReader(file, int64(limit)+1))
```

Same max-plus-one. Callers in `VerifyCorpus`/`verifyOwnedCorpus` then check `len(source) > syntax.MaxSourceBytes` and emit `verify.fixture_input_limit` (`session.go:452-454, 476-478, 505-507, 612-614, 627-629`). Boundary-tested at exactly `1<<20` and `1<<20 + 1` in `cli_test.go:73-100` and `:263-310`.

**Related debt to fold in (D-08):** D-02-04 (add a hang mode to `TestNativeHelperProcess` so `native.timeout` gets a falsifier), D-02-06 (tighten `MaxCLIStreamBytes` 8 MiB → `1<<20`, cite the measured 1,756 B max), D-02-01 (replace the two-needle substring spawn guard with a `go/analysis` pass). Each folds into whatever Phase 3 plan already touches its file — do not create a debt-cleanup plan.

---

## Shared Patterns (apply to every Phase 3 file)

**Fail-closed default arm.** `corevalidate.go:305-306` (`core.unknown_operation`), `interpreterInputs` returning `nil,false` (`session.go:216-217`), `ability.deriveAt` default (`ability.go:139-141`), `deriveAbility` default (`corevalidate.go:424-426`). Unknown input is rejected, never defaulted.

**Determinism by explicit sort.** `sort.Strings(missing)` (`check.go:111`), `sort.Strings(loanIDs)` before selecting the blocking loan (`check.go:258`), `sort.Strings` in both snapshot functions (`check.go:408,415` / `check_test.go:432,437`), repair sort in `ErrorWithRepairs` (`diagnostic.go:60-65`). Map iteration order must never reach an output.

**Non-nil slice initialisation.** `Operations: []core.LinearOperation{}` (`check.go:145`), `Granted: []core.Ability{}` (`ability.go:82`), `Diagnostics/Executions/Lanes` in `protocol.New` (`protocol.go:68`). `[]` vs `null` in JSON breaks digests and `reflect.DeepEqual`.

**Identity = structural hash of a named struct.** `diagnostic.Error`/`ErrorWithRepairs` (`diagnostic.go:43-83`), `protocol.Result.Finalize` (`protocol.go:73-112`), `evidence.manifestID` (`evidence.go:371-410`). Prose and display detail are excluded from identity; kinds, IDs and digests are included. Truncated to 12 bytes hex with a typed prefix (`diagnostic:`, `result:`, `evidence:`).

**Comment the law, not the code.** Every non-obvious invariant in this repo has a paragraph explaining *what breaks if you change it* — `check.go:276-281`, `check.go:359-365`, `corevalidate.go:231-234`, `ability.go:43-45`, `cgen.go:370-374`, `check_test.go:254-263`, `session.go:62-65`. Phase 3 code that skips this will read as foreign.

**Never overclaim.** `corevalidate.KnownEscape = "escape:coordinated-source-core-lie"` (`corevalidate.go:19`) is surfaced in `Result.ExpectedEscapes` (`session.go:582`), the manifest (`evidence.go:221`), and asserted at the CLI (`cli_test.go:253`). `DigestClaim = "content-identity-only"` (`evidence.go:34`). A Phase 3 origin summary must state its own escape the same way.

---

## Inconsistencies — do not propagate the wrong one

**I-1 — Two `addLane` shapes; the Phase 2 one cannot record a failing lane.**
`VerifyCorpus.addLane` (`session.go:431-439`) takes a `status` parameter and is called with `"fail"` before returning (`:462`, `:473`, `:483`, `:502`, `:510`, `:518`, `:523`, `:531`, `:537`, `:546`, `:550`, `:558`). `verifyOwnedCorpus.addLane` (`session.go:583-587`) **hardcodes `Status: "pass"`** and every failure path calls `fail(...)` directly *without adding a lane* — so a Phase 2 verify failure emits no lane record at all, losing the partial-work evidence Phase 1's path preserves. **Copy the Phase 1 (`VerifyCorpus`) shape for Phase 3**, not the Phase 2 one.

**I-2 — Two near-duplicate bounded writers with divergent interfaces.**
`native.boundedWriter` (`native.go:137-155`) exposes `overflowed()`; `evidence.boundedProbeWriter` (`evidence.go:121-136`) does not, so its callers compare `stdout.total > MaxToolProbeBytes` inline (`evidence.go:149-153`). The `Write` bodies also differ in branch nesting (equivalent today, but hand-maintained twice). A Phase 3 third copy would make three. Factor into one internal package, or pick `native.boundedWriter`'s shape and say so.

**I-3 — `readBoundedFile` callers check the limit inconsistently.**
`VerifyCorpus`/`verifyOwnedCorpus` always follow the read with an explicit `len(source) > syntax.MaxSourceBytes` check (`session.go:452, 476, 505, 612, 627`). `CheckFile` (`:153-159`), `FormatFile` (`:122-128`), `RunInterpreterFile` (`:226-232`), `RunNativeFile` (`:307-313`), and `EvidenceCommandFile` (`:345-348`) do **not** — they rely on `syntax.Parse` emitting `syntax.input_limit` downstream. That works for source, but `ValidateEvidenceCommandFile` reads a manifest with `evidence.MaxManifestBytes` (`:370`) and relies on `DecodeStrict`'s own check (`evidence.go:290-292`). New Phase 3 read sites must pick the explicit-check form; the implicit form only works because a specific downstream happens to re-check.

**I-4 — Diagnostic schema is per-constructor, not per-phase or per-program.**
`diagnostic.Error` always stamps `/0`; `ErrorWithRepairs` always stamps `/1`. So a Phase 2 linear program still emits `/0` diagnostics from `checkLinear` (`check.go:133, 137`) and `analyzeStraightLine` (`:243, 331`) for `type.return_mismatch`, `type.unknown`, and `name.unknown`. `TestHumanJSONMixedDiagnosticVersionParity` (`cli_test.go:198-220`) acknowledges the mix. The real rule is **repair-bearing → `/1`**. A Phase 3 diagnostic written with `Error` silently lands in `/0` and carries no repairs — which is *not* what "joins the existing repair-bearing `lang.diagnostic/1` taxonomy" (03-CONTEXT Integration Points) means.

**I-5 — `interp` is a third authorization site that authorizes nothing.**
`interp.runLinear` (`interp.go:54-78`) switches on `core.OpCopy`/`OpMove`/`OpBorrowShared`/`OpReturn` and emits events, but performs **no ability or loan checks** — it trusts the two admission layers. D-12 says "the two admission layers"; the interpreter is a third consumer of `OperationKind` that must be updated for any new kind or it will silently drop operations from the execution trace (and the O0/O3 differential compares traces). Adding `core.OpBorrowExclusive` requires edits in `check`, `corevalidate`, `interp`, **and** `cgen` — say so explicitly in the plan (D-12).

**I-6 — `interp.Schema` no longer means "the schema".**
`interp.go:11` declares `const Schema = execution.Schema0`, but `runLinear` hardcodes `execution.Schema1` inline (`:75, :78, :88`). `evidence.build` sets `ExecutionSchema: interp.Schema` (`evidence.go:209`) then overwrites it (`:218`). The constant is now misleading. Do not add `interp.Schema2`; read the schema off the program.

**I-7 — Two names for the same closed-variant law.**
`ast.Body.HasClosedVariant()` (`ast.go:55`) and `core.Function.HasClosedBody()` (`core.go:102`) express the same XOR at two layers. Adding a third body variant (a CFG body) means both, plus the `core.mixed_body_versions` guard (`check.go:30-41`) which currently assumes exactly two.

**I-8 — Checker and validator count work by different rules.**
`corevalidate.replay` counts its loan-derivation scan (`v.checks++` at `:238`, `:254`); `check.discoverLoanLastUses` (`:366-399`) counts nothing. That asymmetry *is* D-02-03/D-05: `recomputed_work` understates the checker's real Θ(N²) cost while the validator's `Checks` is honest. Fix the checker side; do not copy the checker's convention into the new pass.

**I-9 — `verify-phase1.sh` uses `go run`, `verify-phase2.sh` builds once.**
`verify-phase1.sh:11` is `go run ./cmd/lang verify testdata/phase1`; `verify-phase2.sh:12-14` builds to `$verify_tmp/lang` and reuses it for both corpora and all five `observe` runs. Phase 3 must copy the phase 2 build-once form (a `go run` inside a 20-sample warm loop measures the Go toolchain, not the compiler).

**I-10 — `__LANG_` is live in the collision path (D-06, deadline not priority).**
`cgen.go:360` builds `preferred + "__LANG_" + ...`, a C17 §7.1.3 reserved spelling. No golden contains it today. D-06 requires renaming to `_LANG_` in a **standalone commit before any further C artifact is frozen** — so it must land before the first Phase 3 golden `.c` file, not after.

---

## No Analog Found

| Work item | Role | Data flow | Reason |
|---|---|---|---|
| Debug-lineage bounded experiment (D-01..D-04) | instrumentation | transform | No in-tree analog for source↔native identity joining. Nearest constraints, not patterns: the ID conventions in §3 (the seam must use function-local ordinals, *not* byte offsets — 03-CONTEXT "Established Patterns"), the lane/work/timeout/output-ceiling contract in §5+§8 (D-04), and `wiki/debug-evidence-symbolication-and-proof.md`. If it cannot be expressed as a bounded lane with declared caps, counted work, a stable schema, a timeout, and an honest `available`/`optimized_out`/`not_captured` report, it does not ship this phase (D-04). |
| CFG block/edge core representation | model | — | No `core.Block`/`core.Edge` exists. Closest prior art is out-of-tree: `.planning/spikes/002-cfg-edge-last-use/cfg/model.go:5-15` (`Block{ID, Operations, Successors}`, `Program{ID, Entry, Blocks}`) and `:36-44` (`Endpoint{ID, Kind, Loan, Block, AfterOperation, From, To}`). Import the *shape*, but re-key IDs to §3's function-local ordinal convention (`MatchArm.EdgeID`, `core.go:112`, is the in-tree precedent for a semantically-keyed edge). |
| Public origin summary record (OWN-04) | model | — | No in-tree analog. Prior art `.planning/spikes/003-public-origins-generic-abilities/origins/model.go:26-60` (`Parameter{Mode}`, `ReturnCase{Access, Origins}`, `Function{TypeParams, Params, Returns, Callbacks, Body}`). Must satisfy all five obligations in §3 and be re-derived independently in `corevalidate` per §2. |

---

## Metadata

**Analog search scope:** `internal/compiler/**` (all 22 packages, 8,771 lines), `scripts/**`, `testdata/**`, `.planning/spikes/002`, `.planning/spikes/003`, `.planning/phases/02-*/02-DEBT.md`, `.planning/ROADMAP.md`, `.planning/REQUIREMENTS.md`.
**Files read in full:** `check/check.go`, `check/check_test.go`, `core/core.go`, `corevalidate/corevalidate.go`, `session/session.go`, `ability/ability.go`, `ability/ability_test.go`, `syntax/syntax_test.go`, `native/native.go`, `evidence/evidence.go`, `testsupport/cli_test.go`, `diagnostic/diagnostic.go`, `ast/ast.go`, `execution/execution.go`, `protocol/protocol.go`, `syntax/token.go`, `scripts/verify-phase{1,2}.sh`, `scripts/assert-go-tests.sh`, all `testdata/phase2/*.lang`.
**Pattern extraction date:** 2026-09-03
