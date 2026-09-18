---
phase: 07-calls-signatures-and-call-graph-refusal
recorded: 2026-09-08
status: accepted
disposition: in-progress
items: 11
blocking: 0
---

# Phase 07 — Declared Deferred Scope (D-07-27)

**Written:** 2026-09-08, at *planning* time — not at phase end.
**Amended:** 2026-09-08, after cross-AI review, when the plan set was rewritten
from 5 plans to 8 and D-07-29..D-07-45 were locked.
**Amended again:** 2026-09-08, by 07-02, to bring this file's shape into the
mechanically-checked debt-register format `TestDebtRegistersAreWellFormed`
(`internal/compiler/session/session_test.go`) enforces on every `*-DEBT.md`
register, and to confirm D-07-33's narrowing entry now that the peer it
describes actually ships (`07-02-PLAN.md` Task 2).
Per Key Lesson 4: declare deferred scope in writing at the moment it is decided.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-03-02 | 07-01/07-02 (lineage: M001 D-03-02) | OWN-05, OWN-07, OWN-08, OWN-09, TRU-04, QLT-07 | warning | P09 | Interprocedural loan-*liveness* re-derivation in `corevalidate` does not ship this phase; only the independent signature-summary peer (D-07-20) and the call-graph cycle peer (D-07-19) do |
| D-07-33 | 07-02-PLAN.md (D-07-33, T-07-11) | SEM-06, QLT-08 | warning | P09 | The `corevalidate` summary peer's `Callable` re-derivation is narrowed to the `core.origin_omitted` class only; for `core.origin_understated`, `core.origin_access_mismatch`, and foreign-origin-omitted it can only ever falsely agree with the producer |
| D-07-46 | 07-09-PLAN.md (T-07-09-05) | SEM-05 | warning | UNOWNED(parameterized-shape-callable) | The call argument/return type boundary compares NOMINAL CONSTRUCTOR STRINGS only (`core.Parameter.Type`/`core.Function.ReturnType`, the `/1` `ParameterContract.Type`) — the core schema carries no type ARGUMENTS to compare, so a parameterized shape's arguments cannot be checked at all. Today only the arity-0 constructors `Byte` and `Buffer` are executable, so constructor equality and structural equality coincide and nothing unsound is admitted |
| D-07-47 | 07-09-PLAN.md | SEM-05 | info | UNOWNED(sametype-precondition-unreachable) | `check.call_return_type_unrepresentable` is UNREACHABLE from any legal source program while `sameType` forces every function's declared return type to equal its declared parameter type; it is mutation-killed through a seeded seam (`callReturnTypeDerivationSeam`) and through `corevalidate`'s synthetic input space, never through a `.lang` fixture |
| D-07-48 | 07-09-PLAN.md (probe edge 3, amends D-07-07) | SEM-04 | info | UNOWNED(arity-widens-past-1) | Argument ORDERING has no meaning at arity 1 and is therefore UNANSWERED, not answered, by 07-09's admission gate; when arity widens past 1, a positional argument/parameter correspondence rule must be specified before the widened gate can be called sound |
| D-07-49 | 07-10-PLAN.md (Task 3) | SEM-04, SEM-06 | info | P08 | PVG-04 / CR-02's defect (`computeLoanLastUses` has no `"call"` case) is now CLI-observable through the peer (`relay_escort_witness.lang` refuses at `lang check` as of 07-10) but is NOT fixed here and remains Phase 08 scope |
| D-07-50 | 07-10-PLAN.md (Task 3, WR-01) | SEM-06, QLT-08 | info | UNOWNED(needs-ratified-diagnostic-code) | WR-01's check-side half is deliberately carried: `check` still does not diagnose duplicate `fn` declarations, and `buildCalleeContracts` still silently resolves a call to the LAST declaration while the emitted `CalleeID` names an ID shared by both. Only the user-visible half (the peer's `core.duplicate_function_id` now reaching `lang check`) closed in 07-10 |
| D-07-51 | 07-10-PLAN.md (Task 3, WR-02) | QLT-08 | info | UNOWNED(needs-primary-span-ratification) | WR-02 is deliberately carried: `verifyCallableRefusal` still emits `core.callee_not_callable` with the whole calling function's span, even though `spanByOperationID` already exists and is already used for per-edge cycle spans; it also still returns on the first offending call |
| D-07-52 | 07-11-PLAN.md (T-07-11-06) | SEM-05 | info | UNOWNED(call-site-transfer-marker) | A-normal form has no `f(take y)` syntax, so 07-11's consume-on-call transfer is IMPLICIT in source: a reader cannot see at `identity(buffer)` that `buffer` is consumed there, the way `take buffer` shows it for a plain binding |
| D-07-53 | 07-12-PLAN.md (Task 3) | SEM-04, SEM-05 | info | UNOWNED(fails-schema-set-change) | `core.ForeignReach`'s worst-case join is defined field-by-field over TODAY's vocabulary only, and `FunctionSignature.Fails` is a single string that cannot express a union of two distinct error types |
| D-07-54 | 07-12-PLAN.md (Task 3) | SEM-05 | info | UNOWNED(check-side-paths-untouched) | IN-02 (`check`'s argument-type gate compares the caller's single `typeFact` rather than resolving `argument.place.TypeID`) and IN-03 (`check.Program` clears `result.Program` on a cyclic program but leaves `result.AliasFacts` populated) are deliberately carried, unchanged by this plan |

## Detail

### D-03-02 — interprocedural loan-*liveness* re-derivation

first-recorded: M002

**Interprocedural loan-liveness re-derivation in `corevalidate`.**

Phase 07 ships `corevalidate`'s independent **signature-summary** re-derivation
(D-07-20, `07-02-PLAN.md` Task 2, landed as `corevalidate`'s unexported
`derivePeerSignature`/`recordSummaryPeer`, exposed via `Result.PeerSignatures`)
and its independent **call-graph cycle** traversal (D-07-19, `07-07-PLAN.md`). It
does **not** ship an independent re-derivation of interprocedural *loan
liveness*.

**Lineage.** D-03-02, open past M001: an exported borrow-derived return with no
declared origin exports indistinguishable from a fully-owned return, in the
**interprocedural** half of the hazard. The single-function half closed in M001
Phase 3. The interprocedural half is owned by M002's `OpCall` charter
(D-05-32 / D-05-33).

**Closure gate, verbatim from ROADMAP.md Phase 09:**

> `corevalidate` independently re-derives the same interprocedural loan-liveness
> facts without sharing an implementation with `check`; a seeded endpoint-level
> fault makes the two peers diverge (OWN-07), and D-03-02 is closed — an
> exported borrow-derived return with no declared origin is refused in the
> interprocedural case, in **both** admission layers (OWN-08).

**Explicitly NOT part of this deferral.** The **signature-summary peer ships in
Phase 07** (`07-02-PLAN.md`, D-07-20). Deferring the liveness peer does not
defer the summary peer, and does not defer the seeded faults (D-07-24,
`07-02-PLAN.md` Task 3).

### D-07-33 — the peer's `Callable` re-derivation is NARROWED

first-recorded: M002

**This is stated plainly because an undeclared version of it is exactly the
failure mode this file exists to prevent.**

`Callable` **as published** is the full D-04-03 predicate: publication safety,
i.e. `originvalidate.PublishProblemsFor` returns no problems (D-07-31, D-07-32,
landed in `07-02-PLAN.md` Task 1). `ValidatePublished`/`PublishProblemsFor` can
refuse with any of four classes:

  - `core.origin_omitted`
  - `core.origin_understated`
  - `core.origin_access_mismatch`
  - foreign-origin-omitted (`checkForeignOriginOmitted`, code
    `core.foreign_origin_omitted`)

**In Phase 07, `corevalidate`'s independent peer (`peerCallable`,
`peerReturnDerivesFromBorrow`, in `corevalidate.go`) re-derives only the first
of those four.** For `core.origin_understated`, `core.origin_access_mismatch`,
and foreign-origin-omitted, **the Phase 07 peer can only ever falsely agree
with the producer.** It reports `Callable == true` unconditionally whenever a
public origin is declared (`function.PublicOrigin != nil`), never checking
whether that declaration is understated or access-mismatched, and its forward
borrow-propagation walk deliberately never crosses an `OpForeignCall` hop, so
it cannot detect a foreign-origin-omitted return either. It does not compute
those classes at all, so its agreement on them is not evidence of anything.
This is precisely the single-producer leak D-07-23 names —
`escape:coordinated-source-to-core-false-claim` promoted from origins to the
call contract — scoped and time-boxed rather than denied.
`TestPeerDoesNotRederiveNarrowedClasses`
(`corevalidate_summary_peer_test.go`) asserts this directly: it mutates an
honestly-declared origin into an understated/access-mismatched one (the same
technique `originvalidate_test.go`'s own falsifiers use, since check.go's
honest producer can never construct these declarations itself) and a foreign
declaration case, and shows the peer's `Callable` stays `true` — a **false**
agreement — in all three cases.

**Why narrowed.** A full peer-side origin recomputation is a second
implementation of `RecomputeOrigin` / `RecomputeOriginPerReturn`
(`originvalidate.go:234`, `:133`). That is Phase 09's size, not Phase 07's. This
is D-07-26's scope-cut trigger being pulled **deliberately, at planning time**,
rather than discovered mid-execution.

**What the phase does still guarantee for the narrowed classes:** the producer
computes them, they refuse publication, and `Callable` is false as published.
What is missing is only the *second, independent* derivation.

**Closure phase:** **Phase 09**, alongside the liveness peer, in the same plan
that builds the peer's own origin recomputation.

---

### D-07-46 — Constructor-string comparison is the disclosed granularity, not an oversight

first-recorded: M002

**Closes 07-VERIFICATION.md's single FAILED must-have truth and 07-REVIEW.md's
CR-01 (CRITICAL/BLOCKER).** `07-09-PLAN.md` lands the argument-type gate in
`check.resolveCallBinding` (new code `check.call_argument_type_mismatch`), a
genuinely independent peer in `corevalidate`'s `OpCall` replay (new codes
`core.CallArgumentTypeMismatch` / `core.CallReturnTypeMismatch`, sharing no
helper with `check`), and derives `OpCall`'s `TargetID.TypeID` from the
callee's own declared return contract (fail-closed via
`check.call_return_type_unrepresentable` when unresolvable) rather than
copying it from the caller's argument place. The
`FunctionSignature.Parameters[].Type -> call-site argument type check` key
link 07-VERIFICATION.md marked **NOT WIRED** is now wired and corpus-asserted.

**The retained limitation, disclosed rather than implied:** `core.Parameter.Type`
and `core.Function.ReturnType`, and the `/1` `ParameterContract.Type`, are
plain constructor strings carrying no type ARGUMENTS — the core schema has no
place to hold them. Type identity at the call boundary is therefore NOMINAL
CONSTRUCTOR-STRING EQUALITY, never `TypeID` equality (a `TypeID` is
per-function and can never match across two functions) and never a subtyping
or coercion relation (none exists). Today only the arity-0 constructors
`Byte` and `Buffer` are executable shapes (`check.executableShape`), so
constructor equality and structural equality coincide and nothing unsound is
admitted. The moment a parameterized shape (e.g. `Box[T]`) becomes callable,
this comparison is too weak: two calls with structurally different `Box`
payloads would compare equal on `Box` alone. Reopen this item when a
parameterized shape becomes callable — the schema will need to carry
structured parameter types, not just constructor names.

### D-07-47 — `check.call_return_type_unrepresentable` is unreachable from any legal source program

first-recorded: M002

The fail-closed half of deriving `OpCall`'s `TargetID.TypeID` from the
callee's declared return type cannot be reached by any program the parser
and `sameType` admit: `sameType` forces every function's declared return
type to equal its declared parameter type at the head of `checkLinear`,
`checkBranch`'s match path, and `checkFallibleLinear`, so a caller's own
single type fact always matches a resolvable callee's declared return type
whenever the argument-type gate has already been cleared. This is disclosed
so a reviewer does not mistake the absence of a `.lang` fixture for this code
as an absence of evidence: it is mutation-killed through
`callReturnTypeDerivationSeam` (`check_test.go`'s
`TestCallReturnTypeDerivationMutationKilled`) and through `corevalidate`'s
disjoint synthetic input space (`corevalidate_test.go`'s
`TestPeerRefusesCallReturnTypeMismatch`), never through a source fixture.
Reopens only if a future phase lets a function's declared return type differ
from its declared parameter type (see D-07-46's `Result`/payload-carrying
successor work), at which point a caller will need to synthesize a second
type fact it does not currently mint.

### D-07-48 — Argument-ordering is unanswered at arity 1, amending D-07-07

first-recorded: M002

07-09's admission gate compares exactly one argument against exactly one
declared parameter, because arity is fixed at 1 this phase (D-07-07). This
means the gate answers NOTHING about positional argument/parameter
correspondence — there is exactly one position, so no ordering question
exists to get right or wrong. This is recorded explicitly, rather than left
to be inferred from D-07-07's existing "arity-N calls are deferred" note,
because 07-09's own deterministic-edge probe (probe edge 3, SEM-04/ordering)
surfaced it directly: the moment arity widens past 1, a positional
argument/parameter correspondence rule (which caller argument maps to which
declared parameter, and in what order) must be specified and admitted
BEFORE the widened gate can be called sound — the `/1` schema's
`Parameters []ParameterContract` slice is already arity-ready (D-07-10), but
the CHECKER predicate and the corresponding peer re-derivation are not.

### D-07-49 — PVG-04 / CR-02 is now CLI-observable and is still Phase 08 scope

first-recorded: M002

`check.computeLoanLastUses` has no `"call"` case and reads `RHS.Source`,
which a call binding never populates (a call binding carries `RHS.Arguments`
instead) — so a loan whose last use is a call argument is computed dead and
`ownership.move_while_borrowed` never fires in `check`. `corevalidate`
independently catches it (`core.move_while_borrowed`), and as of 07-10 that
refusal reaches `lang check`: `relay_escort_witness.lang` now reports
`status: invalid` where it previously reported `status: pass`. The
CHECK-SIDE law is UNFIXED and remains ROADMAP.md Phase 08's stated subject
(*Interprocedural Loan Liveness in `check`*) — 07-10 makes the consequence
visible; it does not move the assignment, and `check.computeLoanLastUses`
gains no `"call"` case in 07-10 (`git diff --name-only --
internal/compiler/check/` is empty for this plan). Two known reproductions
Phase 08 inherits: `testdata/phase07/relay_escort_witness.lang` (this
disclosure) and CR-02's `lk5`-shaped program (07-REVIEW.md).

### D-07-50 — WR-01's check-side half deliberately carried

first-recorded: M002

`check` never diagnoses duplicate `fn` declarations: two declarations with
the same name are assigned the identical `semanticID` (a pure function of
module+kind+name, with no uniqueness check across declarations), and
`buildCalleeContracts` — a `map[string]calleeContract` keyed by declared
name, populated in declaration order — silently resolves a call to the LAST
declaration while the emitted `CalleeID` names an ID both functions share.
This means the callee identity recorded by the call graph and the
`ClosureDigest` chain is ambiguous whenever two `fn` declarations collide on
one name. The user-visible symptom is closed incidentally in 07-10 via the
peer's own `core.duplicate_function_id` reaching `lang check`
(`testdata/phase07/duplicate_function_name.lang`); the CHECK-SIDE half — a
span-bearing `name.duplicate_function` diagnostic at the second declaration
site, and closing the ambiguous-`CalleeID` hazard in `buildCalleeContracts`
itself — is not shipped. It needs its own ratified diagnostic code string
and its own mutation kill before it can land; not scheduled to a specific
phase.

### D-07-51 — WR-02 deliberately carried

first-recorded: M002

`verifyCallableRefusal` emits `core.callee_not_callable` with the whole
calling function's span even though `check.Program` already builds
`spanByOperationID` and `checkCallGraphAcyclic` already uses it for per-edge
`cycle_member` spans (07-06/07-07) — so a caller's diagnostic points at the
whole function rather than the specific call site. It also returns on the
first offending call, so a program with two uncallable targets surfaces
them one re-run at a time rather than both at once. Changing a published
diagnostic's `Primary` span is an agent-facing change (`lang-repair` and
`lang explain` both consume diagnostic spans) that deserves its own
ratification, not a drive-by inside 07-10, whose job is the peer-consult
channel, not a diagnostic-quality fix on a path 07-10 does not otherwise
touch.

### D-07-52 — the implicit call-site transfer, accepted and disclosed

first-recorded: M002

**Closes 07-VERIFICATION.md's PVG-01 / 07-REVIEW.md's CR-01 (CRITICAL/BLOCKER).**
`07-11-PLAN.md` lands the consume rule inside `check.resolveCallBinding`
(applying to both call paths through the shared resolver) and a genuinely
independent consume peer in `corevalidate`'s two `core.OpCall` replay arms
(deriving the argument's copy ability from the emitted core artifact's own
`core.TypeFact.Shape`, sharing no helper with `check`): a call transfers its
argument, a copyable argument (`Byte`) is copied, and a non-copyable argument
(`Buffer`) is moved — reachable through the EXISTING `ownership.use_after_move`
gate on the `check` side and the pre-existing generic `core.place_uninitialized`
gate on the `corevalidate` side. `testdata/phase07/call_argument_used_twice.lang`
is the standing negative control; `testdata/phase07/call_argument_used_once.lang`
pins the non-refusing direction.

**The retained limitation, disclosed rather than implied:** because A-normal
form has no `f(take y)` syntax (D-07-01: a call's argument must be the NAME of
an in-scope binding, never a nested `take`/`borrow`/`borrow mut` expression),
the transfer a call performs is **implicit** at the call site — a reader
cannot see, at `let first = identity(buffer)`, that `buffer` is consumed
there, the way `take buffer` makes an ordinary binding's transfer visible in
source text. This is accepted (T-07-11-06), not silently absorbed: the
rejected alternative — refuse every non-copyable call argument outright,
requiring an explicit call-site marker before it can be passed — was
considered and declined at the ratifying checkpoint, because A-normal form
has no such marker today, making `Buffer` **unpassable to any call** and
flipping `testdata/phase07/relay_escort_witness.lang` (the D-03-02 witness
this phase forbids disturbing) from clean to refused for an unrelated reason,
erasing the deferred divergence rather than deferring it. **Reopens** when a
call-site transfer marker (e.g. an explicit `f(take y)` form) is added to the
grammar, at which point the marker becomes required and the implicit-transfer
disclosure here is retired.

### D-07-53 — join granularity: three plain strings, no structured lattice

first-recorded: M002

`core.ForeignReach` carries three plain string policy fields (`Allocator`,
`Unwind`, `NonlocalExit`) with no structured lattice in the schema, so
07-12's worst-case join is defined field-by-field over TODAY's vocabulary:
`forbidden` is strictly more constraining than `permitted` and wins the
join in either argument order; a non-empty allocator name beats an empty
one; two DIFFERENT non-empty allocator names — which no vocabulary rule
orders — join to the declared, named `core.ForeignReachConflict` sentinel
rather than to an arbitrary pick.

`FunctionSignature.Fails` is a single string and therefore CANNOT express a
UNION of two distinct error types: a caller reaching two distinct fallible
callees publishes only ONE of them (the caller's own local value if
non-empty, else the first non-empty value encountered in sorted callee-ID
order — deterministic in production, never last-writer-wins). This is a
known imprecision in the DISCLOSING direction only: a caller is never
wrongly told it is infallible when it is fallible, but it may under-name
WHICH failure it can produce. **Reopens** when a third policy value beyond
`forbidden`/`permitted` lands, when a second allocator vocabulary is
introduced, or when `Fails` changes from a single string to a set in the
schema — at which point the join defined here must be re-specified, not
silently reused.

**Note (07-REVIEW.md IN-02, iteration 1 fix):** the producer/peer
agreement this debt's disclosure depends on (both sides folding over the
same sorted-callee-ID order — see `corevalidate.go`'s `peerJoinFails` doc
comment) is now test-caught, not just documented, by
`TestPeerFailsAgreesOnDisagreeingMultiCalleeJoin`
(`corevalidate_foreign_closure_test.go`), driven by the new fixture
`testdata/phase07/call_two_fallible_callees_disagree.lang`. This closes
the review finding that the coincidence was undefended by a dedicated
test; it does NOT close D-07-53 itself — `joinFails`/`peerJoinFails`
remain order-DEPENDENT within a single side (accumulator wins), and the
schema-level imprecision (`Fails` cannot express a union) is unchanged
and still reopens under the conditions stated above.

### D-07-54 — IN-02 and IN-03 deliberately carried

first-recorded: M002

**IN-02:** `check`'s argument-type admission gate compares the CALLING
function's single `typeFact` (one type fact per function, forced by
`sameType`) rather than resolving `argument.place.TypeID` directly. This is
sound only while `sameType` holds; `corevalidate`'s own independent
re-derivation already does the more robust per-place resolution, so the two
derivations would diverge in the PERMITTING direction on the `check` side
the moment a function is legally able to carry two distinct type facts.

**IN-03:** `check.Program` clears `result.Program` on a detected call-graph
cycle (D-07-14/D-07-15) but leaves `result.AliasFacts` populated — a
half-cleared result whose stated invariant is "nothing from a cyclic
program escapes admission". Latent today because no production consumer
reads `AliasFacts` from a refused result.

Both are on `check`-side paths this gap-closure run does not touch (its own
scope is the summary-signature join, `originvalidate`/`corevalidate`, and
the IN-01 index fix in the same pass); recorded here rather than silently
omitted, per 07-REVIEW.md's own INFO-level findings. Not scheduled to a
specific phase.

---

## Phase 07's own scope-cut trigger (D-07-26) — updated for the 8-plan set

If **Stage 0 + Stage 1** (`07-01` + `07-02` + `07-03` + `07-04` + `07-05`) exceed
**~2x their initial plan estimate**, Stage 2's cycle refusal (`07-06`, `07-07`)
renegotiates out to Phase 08.

**New consequence introduced by the post-review re-plan, recorded so it is not
discovered under pressure:** `07-08` (closure-digest chaining) **goes with them.**
D-07-38 orders digest chaining strictly behind cycle refusal — the chain
terminates only on a DAG — so cutting `07-06`/`07-07` necessarily cuts `07-08`.
If that cut is taken, `ClosureDigest` ships with its zero-callee base case only
(`07-01`), which is correct and complete for a corpus with no calls, and the
chaining arm plus its callee-changes-invalidates-caller regression move to
Phase 08 with the cycle refusal.

**Never cut under this trigger:**
- the Stage 0 `corevalidate` signature-summary peer (D-07-20)
- the seeded faults (D-07-24), including the bilateral one that must FAIL the gate

---

## Accepted, declared limitations of Phase 07 (declared, not implied)

- **`cgen`'s `OpCall` case is structurally unreachable this phase** (A-02).
  `cgen.Emit`/`EmitNative` hard-fail on `len(Functions) != 1` (`cgen.go:22,52`)
  and both exhaustive-dispatch controls gate the `cgen` site behind
  `len(program.Functions) == 1`. Phase 07 registers the case as forward hygiene
  for Phase 11 and claims **no** `cgen` runtime coverage for `OpCall`. Per
  **D-07-39** the arm is an explicit, dedicated "recognized, unsupported in
  Phase 07" error and is never folded into the grouped copy/move/borrow cases —
  folding it there would emit copy-like C and let a green control certify a stub.
- **`interp` recognizes but does not execute `OpCall`** (D-07-39). No call-stack
  semantics ship this phase; SEM-08's bounded call stack with its fixed
  documented ceiling is Phase 10. The exhaustive-dispatch controls assert
  **recognition**, never execution, and say so in their doc comments.
- **`restrict` at arity 1 is correct but never load-bearing** (D-07-04).
  Phase 11 must scope its `restrict` evidence to single-parameter derivation and
  name the two-parameter case as an M003 obligation, rather than implying
  two-pointer coverage it does not have.
- **No maximum call-graph depth is declared this phase** (D-07-17). Depth is
  unbounded provided the graph is acyclic; the distinguisher is on-stack (gray)
  re-entry, never visit count or depth. The fixed documented interpreter ceiling
  is SEM-08, Phase 10, informed by spike S-007.
- **No `Span` on `core.LinearOperation`** (D-07-35). Cycle-diagnostic spans are
  projected on the `check` side from operation IDs. The `corevalidate` peer emits
  the shared code string with **no spans** — its job is refusal, not diagnostics.
  Any future consumer wanting spans from the peer must add them deliberately.
- **No cross-module summary channel** (D-07-34). `check.Program(ast.Program)`
  keeps its signature; admission consults an in-process pre-body signature table.
  Cross-module summary consumption is Phase 09+.
- **No interprocedural fact is marked cacheable** (`07-08`). QLT-06 is a later
  requirement; Phase 07 builds the callee-changes-invalidates-caller regression
  and the closure-derived key it will need, and modifies `cache.Input` not at all.
- **Arity-N calls and multi-argument ALIASING** (D-07-07, CORRECTED — see
  below) are deferred. The `/1` schema is arity-ready (D-07-10), so the
  future change is a checker predicate plus an aliasing rule, not a `/2`.
  **Correction, not a supplement:** D-07-07's wording previously read as
  though it deferred "multi-argument loan interaction" while implying the
  SINGLE-argument case was already handled. It was not. From `07-03` through
  `07-09`, a call neither moved nor copy-checked its argument at all — the
  call boundary bypassed affine ownership entirely, so the identical
  non-copyable value could be passed to two separate calls and admitted by
  BOTH `check` and `corevalidate` (07-REVIEW.md **CR-01** / 07-VERIFICATION.md
  **PVG-01**). Reading D-07-07 as evidence the single-argument case was sound
  was the misreading this entry invited; it is retracted here. **`07-11`
  closes the single-argument case** with consume-on-call, ability-decided,
  refused independently by both layers (see D-07-52 below for the residual
  it accepts). What remains genuinely and ONLY deferred under D-07-07 is the
  MULTI-argument aliasing rule: with two or more arguments, consumption
  becomes an argument-ALIASING question (two arguments naming the same
  place; one shared and one exclusive loan of one owner) — a genuinely
  different rule from the single-argument consume-on-call law 07-11 ships,
  together with D-07-48's positional-correspondence obligation.
- **Labeled call-site arguments** (M001 D-02) are deferred, not revoked (D-07-06),
  and reinstated when arity widens past 1.
- **Keyed/signed summary digests** are out of scope. D-07-13: content digests
  detect staleness, never forgery — chaining them (`07-08`) widens what staleness
  they detect and adds no authenticity. Independent re-derivation is the forgery
  answer.

---

## Explicitly rejected, not deferred

Recorded here so a later phase does not re-open them believing they were merely
postponed.

- **Call-graph edges in the signature summary** — permanently rejected (D-07-11).
  It would make every consumer whole-program-aware and re-import the
  already-rejected Design A.
- **An export set in `core.Program` / `core.Function`** — rejected (D-07-31).
  `Callable` is publication safety, not export membership;
  `originvalidate.ValidatePublished` never reads an export list, so the schema
  never wanted the field. Suggested by a reviewer; declined on the evidence.
  Confirmed landed in `07-02-PLAN.md` Task 1: the string `Exports` never
  appears in `originvalidate.go`
  (`TestBuildInterfaceNeverConsultsExportList`).
- **The `export_callee` repair for the SEM-06 refusal** — rejected (D-07-31c).
  Exporting a function cannot fix an unsafe borrow-derived return; an agent
  applying the repair would re-submit an identically-refused program. Suggested
  by a reviewer; declined on the evidence. If a repair is ever offered for that
  code it must target the return-origin contract.
- **A depth or visit ceiling on the call-graph traversal** — rejected (D-07-17).
  It would convert a correctness property into a resource limit and silently
  refuse legal deep-but-acyclic programs. Only the emitted diagnostic is bounded.

---
*Written at planning time per D-07-27; amended after cross-AI review; amended
again by 07-02 for debt-register shape and D-07-33 confirmation.*
