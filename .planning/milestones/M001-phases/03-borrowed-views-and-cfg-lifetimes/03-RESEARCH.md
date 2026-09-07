# Phase 3: Borrowed Views and CFG Lifetimes - Research

**Researched:** 2026-09-04
**Domain:** Static ownership/borrow checking — CFG liveness, exclusive/shared loan
conflicts, separately-compiled public borrow summaries
**Confidence:** MEDIUM-HIGH (strong in-repo evidence and validated prior-art spikes;
MEDIUM on anything requiring new surface syntax, which is Claude's discretion)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

- **D-01:** Preserve the debug-lineage seam described in
  `wiki/debug-evidence-symbolication-and-proof.md`. Source → typed-core →
  lowering → native identity must remain joinable; do not foreclose it.
- **D-02:** Treat that note's "Next bounded experiment" as a **bounded experiment**,
  not a build-out. It is an admission test for whether the seam earns its cost —
  not a commitment to production crash capture.
- **D-03:** Explicitly out of scope this phase: DWARF/CodeView emission, crash
  storage/upload/retention, symbol servers, minidump capture, proof/SMT tiers.
  The wiki note's own guidance is to defer these until concrete workloads justify
  them; honour that.
- **D-04:** The experiment must be bounded the way every other Phase 2 lane is —
  declared caps, counted work, stable schema, timeout, output ceiling — and must
  report `available` / `optimized_out` / `not_captured` honestly rather than
  inventing a value. If it cannot be bounded, it does not ship this phase.
- **D-05:** D-02-03 (checker is Θ(N²) in body size after transitive loan sets) is
  *this phase's* problem, not an afterthought: OWN-03's CFG/edge-specific liveness
  should remove the quadratic factor by construction. Make `recomputed_work` count
  the propagation so the metric stops understating real cost.
- **D-06:** D-02-05 (`__LANG_` violates C17 §7.1.3 reserved identifiers) has a
  deadline, not a priority: rename to `_LANG_` in a standalone commit **before any
  further C artifact is frozen**. Every new golden widens the freeze.
- **D-07:** D-02-09 (`Box`/`Pair` type-check but die spanless exit 3 on every
  engine) is the same taxonomy smell that made CR-01 a Phase 2 blocker. Prefer a
  causal compile-time diagnostic over widening the C backend.
- **D-08:** The remaining debt (D-02-01 spawn-guard strength, D-02-02
  `evidence.canonical_unstable` unreachable at the CLI, D-02-04 missing
  `native.timeout` falsifier, D-02-06 8 MiB CLI ceiling, D-02-07 literal-line
  causality seam, D-02-08 `protocol.Human` divergence) is real but not
  phase-shaped. Fold each into whatever plan already touches its file; do not
  create a debt-cleanup plan.
- **D-09:** A differential test is not evidence until reverting the production hunk
  makes it fail. **Mutation-kill every oracle.** Phase 02's ownership oracle
  encoded the same wrong law as production and the differential agreed on the wrong
  answer.
- **D-10:** Interrogate what inputs a green property test actually *reaches* before
  trusting it. Two Phase 02 blockers hid behind generators whose reachable input
  space omitted the hard case — coverage was never the problem; the production code
  was fully executed both times.
- **D-11:** Drive behaviour through the shipped binary on hand-written programs,
  not only the gate's own corpus. The gate only ever sees what ships with it.
- **D-12:** Preserve the independence of the two admission layers (checker and
  `corevalidate`). Where a change must touch both, say so explicitly and record
  that the differential cannot cross-check that specific row (as OV-02-01 does).
- **D-13:** Phase 1 goldens and schemas stay byte-identical. A Phase 2/3 golden may
  move only as a *causal* consequence of a deliberate semantic change, and the diff
  must be explained field-by-field.
- **D-14:** One global generated-C ordinary-identifier namespace; preferred
  source-derived names when unique, deterministic suffix only on real collision.
- **D-15:** Every input read and every spawned process stays bounded: deadline,
  independent stdout/stderr caps at max-plus-one, no `CombinedOutput`.

### Claude's Discretion

Plan decomposition, wave structure, how many plans, which fixture shapes exercise
CFG edges, and the concrete form of the debug-lineage experiment — provided it stays
bounded per D-04 and does not cross the D-03 scope fence.

### Deferred Ideas (OUT OF SCOPE)

- Full DWARF/CodeView emission, symbol bundles, crash capture/upload/retention,
  minidumps (D-03) — deferred until a concrete workload justifies the cost.
- Proof/SMT/refinement tiers (D-03) — deferred per the wiki note's own guidance.
- D2 dogfooding (a Lang-written corpus/evidence runner) — the compiler and harness
  remain Go; noted in Phase 02's decisions, not this phase's scope.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| OWN-03 | Shared and exclusive loans obey conflict rules and ordinary local loans end at proven CFG point/edge-specific last use. | Q1 (branching-syntax scoping), Q2 (backward worklist dataflow algorithm), Q3 (conflict matrix + fixture set), Q4 (independent bounded-path oracle) |
| OWN-04 | Public borrowed results record verified field/alternative origins and access mode without inspecting provider bodies downstream. | Q5 (`PublicOrigin` fact shape, staleness/omission/impossibility detection), Q6 (separate-compilation summary artifact) |
</phase_requirements>

## Summary

Phase 3 has a scope trap and this research resolves it first: **the current
surface language has no branching syntax at all.** `LinearBody` (source AST,
`internal/compiler/ast/ast.go:69-85`) is a flat list of `let`/`take`/`borrow`
bindings terminated by one `Result` name — no `if`, no loop, no nested block.
Match bodies (`internal/compiler/check/check.go:83-117`) are a single-level
exhaustive dispatch over closed-variant patterns where each arm immediately
yields an alternative value — there is no space *inside* an arm to place a
divergent loan use. Given this, OWN-03's "branch-specific last use ends a loan
on the correct CFG edge" cannot be demonstrated on the syntax that exists today
without adding *some* branching construct. Spike 002 already built and
validated the underlying analyzer (backward loan liveness over a small CFG,
independent bounded-path oracle) against a hand-built CFG notation, never
against real Lang source — so the spike's algorithm is validated but the
phase's surface-to-CFG lowering is still new work. Recommendation: add the
**smallest possible branching form** — a two-arm linear `if` over a `Byte`
condition, or (cheaper, reuses more) let a `match` arm's *value* position hold
a full linear body instead of a bare alternative name. Either gives >1 CFG
block with a real join. This is scoped as Claude's discretion per CONTEXT.md
but the research strongly favors extending `match` arms over inventing `if`,
because it reuses the existing exhaustiveness/arm/edge machinery in both
`check.go` and `corevalidate.go` rather than adding a second control-flow
surface concept.

The liveness algorithm itself is not open research: Spike 002 already selected
and validated **backward monotone dataflow over a small CFG with a worklist
fixpoint** (the standard "Kildall-style" finite-lattice dataflow used by every
production compiler's liveness pass, and the same shape Rust's NLL/Polonius use
for loan-region liveness) against an independent bounded-path-enumeration
oracle, and it removes the Θ(N²) transitive-loan-set cost recorded as D-02-03
by construction (finite per-block live-sets updated to a fixpoint, not a
transitive per-binding scan). Both admission layers (`check.go` and
`corevalidate.go`) currently duplicate an O(N²)-shaped transitive scan
(`discoverLoanLastUses` and `replay`'s `loansForPlace`, respectively) and both
need the CFG liveness treatment independently, per D-12.

OWN-04's public-origin shape is also not open research: Spike 003 validated
value-parameter origin paths, tagged alternatives, and independent
`copy`/`drop`/`share`/`send`/`escape` abilities as the public summary contract,
and Spike 004 validated an independent (source-blind) recompute-only
certificate checker as the right shape for a separate-compilation oracle, with
one explicitly accepted escape (`KnownEscape` — a coordinated frontend/summary
lie). Phase 3's job is to wire these two already-validated shapes into the
production `core`/`corevalidate` packages, not to re-derive them.

**Primary recommendation:** Extend `match` arms so an arm's value position may
be a full linear body (giving the CFG its first real branch/join), implement
loan liveness in `check.go` and `corevalidate.go` as independent backward
worklist dataflow over that per-function CFG (killing D-02-03 in both layers,
not just check.go), and add public origin paths + access mode as a new
optional field on `core.Function`/a new `core.PublicOrigin` fact validated by
a source-blind `corevalidate` pass extended with a compact recompute-only
certificate for cross-module (separate-compilation) checking — deliberately
NOT a second frontend, per Spike 004's "PARTIAL" finding. Treat the
debug-lineage experiment as a single new field on existing IDs, not a new
subsystem; scope it to steps 1, 2, and 4 of the wiki's six-step list only.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Surface branching construct (match-arm-with-body) | Frontend (parser/AST) | Typed core (CFG lowering) | Syntax must exist before a CFG can be built; `ast.go` currently has none |
| CFG construction (blocks/edges) | Typed core (`check.go`) | — | `check.go` already owns straight-line lowering; CFG lowering extends it, not a new package |
| Backward loan-liveness dataflow | Typed core (`check.go`) | Independent admission (`corevalidate.go`) | Both admission layers must independently derive liveness per D-12; neither may import the other |
| Bounded path oracle (differential) | Test/verification harness (`session.go` verify lane) | — | Per success criterion 1 and D-09, this must be a *third*, genuinely different algorithm, not reused production code |
| Public borrow origin summary (fields, alternatives, access mode) | Typed core (`core.Program`/new fact) | Ability derivation (`ability.go`) sibling | Origins ride alongside the sealed ability derivation, per Spike 003, but are a distinct fact, not an ability |
| Separate-compilation summary validation | Independent admission (`corevalidate.go` extension or sibling package) | — | Must stay source-blind and body-blind per Spike 004; cannot import the checker |
| Debug-lineage IDs (source→core→native) | Typed core (ID assignment) | Native backend (`cgen`) | IDs are assigned once, at lowering; native backend only threads them through |

## Package Legitimacy Audit

Not applicable this phase. Phase 3 is pure Go standard library and existing
in-repo packages (`internal/compiler/*`), matching Phase 1 and Phase 2. No new
external dependency is anticipated. If planning introduces a `go/analysis`-based
static-spawn-guard rewrite (D-05 fix target D-02-01), that would still use only
`golang.org/x/tools/go/analysis` if adopted — **flag for a `checkpoint:human-verify`**
if the planner selects that path; the standard-library-only alternative (a
proper `go/ast` walk without the `x/tools` dependency) avoids the question
entirely and is consistent with FND-01's "Go 1.24 and its standard library"
constraint. No packages are recommended or verified in this research; if the
planner chooses `x/tools`, it must run the Package Legitimacy Gate then.

## Q1 — CFG representation and the branching-syntax question (pivotal)

**Finding, verified by reading the source this session:**

- `ast.LinearBody` (`internal/compiler/ast/ast.go:69-73`): `Bindings []Binding;
  Result string; Span`. `Binding.RHS.Kind` is one of `"take"`, `"borrow"`, or
  the default (implicit copy) — confirmed by reading `check.go:249-318`'s
  switch on `binding.RHS.Kind`. There is no conditional, loop, or nested-block
  AST node anywhere in `ast.go`. `[VERIFIED: internal/compiler/ast/ast.go:69-85]`
  — quoted verbatim above.
- `checkLinear`/`analyzeStraightLine` (`internal/compiler/check/check.go:130-344`)
  process bindings in a single `for index, binding := range body.Bindings`
  loop — i.e., today's "CFG" is a single basic block, entry to return, no
  branches. `[VERIFIED: internal/compiler/check/check.go:239]` (loop header)
  — quoted: `for index, binding := range body.Bindings {`.
- Match bodies (`check.go:83-117`) have arms, but each arm's `Value` is a bare
  alternative *name* (`core.MatchArm{Pattern, Value string}`,
  `internal/compiler/core/core.go:111-116`), not a sub-body. There is no
  binding, borrow, or loan possible inside a match arm today.
  `[VERIFIED: internal/compiler/core/core.go:111-116]` — quoted:
  `type MatchArm struct { ID string; EdgeID string; Pattern string; Value
  string }`.
- Spike 002's CFG (`.planning/spikes/002-cfg-edge-last-use/README.md`) is a
  **hand-built Go notation of blocks/edges**, not derived from Lang source at
  all — the spike explicitly scopes out "native lowering" and works on
  synthetic block graphs. `[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:25-27]`

**Conclusion:** OWN-03 cannot be demonstrated on today's surface syntax. A
branching construct must be added this phase, or OWN-03's success criterion 2
("a branch-specific last use ends a loan on the correct CFG edge... omitting
that edge is detected") is unimplementable as stated. This is not avoidable by
reinterpreting "branch" as "match arm" under the *current* match semantics,
because arms have no body to hold a loan use today.

**Recommendation — smallest sufficient extension:** Give a match arm's value
position an optional full `LinearBody` instead of a bare alternative name
(`Arm.Value` becomes `oneof(bareName, LinearBody)` or the surface always
requires a body that trivially returns the bare name in the S0 case for
backward compatibility). This:
- reuses the existing exhaustiveness, arm identity (`ID`), and `EdgeID`
  machinery already present in both `check.go` and `corevalidate.go`
  (`match.go` sections, `internal/compiler/corevalidate/corevalidate.go:115-146`)
  instead of inventing a second surface control-flow concept (`if`);
- gives exactly the CFG shape Spike 002 validates: one entry block, N
  arm-successor edges, one join at return — a diamond/fan-out, matching the
  spike's "100 generated diamond products" gate;
  `[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:86]`
- keeps Phase 1's non-linear match fixtures byte-identical (D-13) because the
  bare-`Value`-name form stays legal — it lowers to a trivial one-node CFG.

**Rejected alternative:** adding `if`/loop syntax fresh. This duplicates
control-flow concepts (`match` vs `if`) in the surface grammar for no semantic
gain this phase, and Spike 002 explicitly validated loop-carried loans and
loop-exit edges too — so if the planner wants loop coverage specifically,
match-arm-bodies plus recursion (already excluded — no calls exist yet) cannot
give it. **Loops are out of reach this phase regardless of syntax choice**,
because there is no recursion or iteration construct in the language at all
yet (no function calls exist outside the one function parameter — confirmed by
absence of any `Call` AST node in `ast.go`). Recommend explicitly scoping OWN-03
this phase to **branch** edges (arm/join), not loop-exit edges, and recording
loop-carried loans as a documented Phase 4+ follow-on, consistent with Spike
002's own "known limitations" framing of loop-iteration dynamic loan identity
as a separate experiment. `[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:172-179]`

**This is the phase-split risk to flag explicitly:** ROADMAP success criterion
2 says "branch-specific last use... on the correct CFG edge" — this is
achievable with match-arm-bodies. It does **not** say "loop-specific" — good,
because loops are not achievable this phase without inventing a second new
control construct (loop) in addition to arm-bodies, which is a second unbounded
scope increase the CONTEXT.md discretion note does not appear to anticipate.
**Recommendation: scope OWN-03 to acyclic (branch-only) CFGs this phase.**

## Q2 — Liveness algorithm

**Recommendation:** backward monotone dataflow over a finite lattice (the set
of currently-live loan IDs per block boundary), computed to a fixpoint with a
worklist, exactly as Spike 002 implemented and validated
(`.planning/spikes/002-cfg-edge-last-use/README.md:106-112`, "Iteration 2 —
backward set liveness and edge materialization"). This is the standard
technique described in every production-compiler dataflow reference the spike
itself cites — Clang's dataflow-analysis framework documents the general
"finite-height lattice plus monotone transfer function reaches a fixpoint, and
a worklist avoids reprocessing unaffected blocks" law
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:36-37, citing https://clang.llvm.org/docs/DataFlowAnalysisIntro.html]`,
and Rust's NLL RFC frames the in-scope-loan set as exactly this kind of
fixed-point dataflow computation
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:33, citing https://rust-lang.github.io/rfcs/2094-nll.html]`.
This is not a novel formulation for this project to invent — it is the named,
decades-old technique (commonly attributed to Kildall's 1973 worklist
algorithm for dataflow analysis; `[ASSUMED]` — attribution from training
knowledge, not independently re-verified via web search this session because
no search provider is configured for this project (`.planning/config.json`:
all of `brave_search`/`exa_search`/`tavily_search`/`ref_search`/`firecrawl`/
`jina`/`perplexity` are `false`); the technique's *substance*, not its
historical name, is the load-bearing claim and that substance is
`[CITED]` from the two sources above).

**(a) Edge-specific:** the transfer function must be able to produce a
*different* live-out set per successor edge (not just per block), so that one
arm's use of a loan and another arm's non-use of it produce different loan-end
placements on the two edges out of the match point. Spike 002's endpoint
identities (`edge:entry:else:view`, `point:then:0:view`) demonstrate this is
representable with a stable identity scheme; Phase 3 should mirror that
`{block/edge}:{location}:{loan}` shape onto the real `core.Function` IDs
(e.g., `{functionID}:edge:{armIndex}:{loanID}`).
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:88-89]`

**(b) Removes Θ(N²):** today's cost (D-02-03) comes from `discoverLoanLastUses`
in `check.go` propagating a loan-index list forward through every subsequent
binding that derives from it (`internal/compiler/check/check.go:366-399`,
specifically the `for _, loanIndex := range loansForBinding[sourceBinding]`
loop inside the outer `for index, binding := range body.Bindings` loop —
`[VERIFIED: internal/compiler/check/check.go:370-388]`, quoted structure:
`for index, binding := range body.Bindings { ... for _, loanIndex := range
loansForBinding[sourceBinding] { ... } ... }`) — a chain of N reborrows each
carrying a growing loan list is O(N) work per binding, O(N²) total. The
**same shape exists independently in `corevalidate.go`**: `replay`'s
`carried := append([]string(nil), loansForPlace[operation.SourceID]...)`
inside `for index, operation := range operations` is the identical O(N²)
pattern. `[VERIFIED: internal/compiler/corevalidate/corevalidate.go:237-251]`,
quoted: `carried := append([]string(nil), loansForPlace[operation.SourceID]...)`
inside `for index, operation := range operations {`. **This is new information
this research surfaced: D-02-03/D-05 as recorded in `02-DEBT.md` names only
`check.go`; `corevalidate.go` has the identical asymptotic defect and must be
fixed too, independently, per D-12.** A backward worklist dataflow over
per-block live-sets does a bounded number of transfer evaluations per block
(Spike 002 measured exactly 1,501 transfer evaluations for a 1,501-block
graph — `[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:87, 152-153]`),
which is linear in program size for an acyclic CFG, removing the quadratic
factor by construction.

**(c) Bounded work / `recomputed_work`:** count one unit of work per transfer
function evaluation per block (matching the `check.work_limit` counted-work
convention already used elsewhere in `check.go`, e.g. `result.Work++` at
`internal/compiler/check/check.go:240` and `328`). D-05 explicitly requires
`recomputed_work` to count the propagation so the metric stops understating
real cost — the fix is mechanical once the dataflow is block-shaped: increment
work once per block visited by the worklist, not once per binding, so a
diamond that never revisits a block reports work proportional to blocks, not
to (blocks × bindings-per-block-derived-from-earlier-blocks).

## Q3 — Exclusive vs shared loan conflict rules

**Current state:** only shared loans exist. `core.OpBorrowShared` is the only
borrow operation kind (`internal/compiler/core/core.go:81`, quoted:
`OpBorrowShared OperationKind = "borrow_shared"`); there is no
`OpBorrowExclusive` and no `borrow mut` surface syntax anywhere in `ast.go`.
`[VERIFIED: internal/compiler/core/core.go:78-83]`. Today's only conflict rule
is move-vs-any-active-loan (`check.go:253-271`, `corevalidate.go:290-293`);
shared-vs-shared is implicitly always legal (no gate exists) and there is no
shared-vs-exclusive or exclusive-vs-exclusive rule to test because exclusive
loans do not exist yet.

**Required conflict matrix (standard borrow-checker law, matching Rust's model
which Spike 002/003 both take as their reference architecture):**

| Combination | Rule | Rationale |
|---|---|---|
| shared + shared (same owner, overlapping liveness) | Accept | Multiple simultaneous readers never observe a mutation |
| shared + exclusive (overlapping liveness) | Reject | An exclusive loan promises no concurrent observer; a live shared loan violates that |
| exclusive + exclusive (overlapping liveness) | Reject | Two exclusive loans would both promise sole access |
| loan (either kind) vs move (owner moved while loan live) | Reject | Already implemented for shared (`ownership.move_while_borrowed`); must extend identically to exclusive |
| shared or exclusive loan, non-overlapping liveness (sequential) | Accept | This is exactly what CFG-edge-specific last use exists to prove — a loan ended before the next one starts is not a conflict |

**Smallest fixture set exercising accept/reject for each combination** (7
fixtures, extending the existing `move_while_borrowed.lang` /
`reborrow_while_moved.lang` pattern):

1. `shared_shared_accept.lang` — two overlapping `borrow x` bindings both used
   after both exist; accept.
2. `shared_exclusive_reject.lang` — `borrow x` then `borrow mut x` while the
   first is still live; reject with a new `ownership.borrow_conflict` (or
   similarly named) diagnostic.
3. `exclusive_exclusive_reject.lang` — two overlapping `borrow mut x`; reject.
4. `exclusive_move_reject.lang` — `borrow mut x` then `take x` while live;
   reject (mirrors existing `move_while_borrowed.lang` but for exclusive).
5. `sequential_shared_then_exclusive_accept.lang` — `borrow x` fully used and
   dead (CFG last use passed) before `borrow mut x` begins; accept — this is
   the fixture that specifically exercises edge-specific liveness, not merely
   lexical scoping.
6. `branch_one_arm_shared_accept.lang` — a match-arm-body borrows on one arm
   only; the owner is mutated/moved after the join on the *other* arm's edge;
   accept (mirrors Spike 002's core disagreement-detection fixture).
7. `branch_one_arm_shared_reject.lang` — same shape, but the mutation is
   reachable from the *borrowing* arm's edge; reject.

Fixtures 6-7 directly encode success criterion 2's "omitting that edge is
detected" — the deliberate defect is "treat both edges as if the loan died
uniformly," which fixture 6 would wrongly reject and fixture 7 would wrongly
accept if edge-specificity were missing (matching Spike 002's own injected-fault
methodology: "one branch uses a shared view, the other does not, and a
mutation after the join is legal only if the unused edge ends the loan").
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:93-94]`

## Q4 — Independent bounded-path oracle for OWN-03

**Recommendation:** implement bounded explicit path enumeration over the
per-function CFG as a **separate package** under, e.g.,
`internal/compiler/pathoracle/` (name TBD by planner), consumed only by the
verification harness (`session.go`'s verify lane), never by `check.go` or
`corevalidate.go`. This mirrors Spike 002's own architecture exactly: "the
independent path oracle reuses the prior spike's linear checker and normalizer
[for the linear-segment portion]... a different algorithm but not a formal
proof or independent compiler."
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:180-182]` For this
production version, the oracle should independently walk every acyclic path
from entry to return (bounded because loops are out of scope this phase per
Q1), linearize each path into a straight-line ownership sequence, and check it
against the **existing, already-validated** straight-line last-use logic
(effectively today's pre-CFG `discoverLoanLastUses`, kept alive specifically
as the oracle's building block rather than deleted — a good reuse of an
already-correct piece once it's no longer on the production hot path).

**Genuinely different from production per D-09/D-12:** the oracle differs from
the block-dataflow production analyzer in mechanism (exhaustive path expansion
vs. finite-set fixpoint), not merely in file location — this satisfies D-09's
"a differential test is not evidence until reverting the production hunk makes
it fail" bar only if the oracle does NOT import or call any production
liveness function. Recommend the plan structure a task that literally deletes/
comments the production CFG-liveness call, runs the differential, and confirms
it fails, before oracle-writing is considered done (this is a direct
application of D-09's own text, and is exactly the check that Phase 02's own
retrospective (D-09 origin) says was skipped for the ownership oracle: "the
ownership oracle encoded the same wrong law as production and the differential
agreed on the wrong answer." `[VERIFIED: .planning/phases/02-owned-values-and-abilities/02-DEBT.md:181-182]`,
quoted verbatim in that source).

**How to mutation-kill it:** (1) revert the production CFG-liveness commit
locally and confirm the differential fails (D-09's explicit bar); (2) seed the
omitted-edge-endpoint fault from Q3 fixtures 6-7 and confirm the oracle
disagrees with a deliberately-broken production analyzer that always ends
loans at the join rather than per-edge; (3) run the same reorder/alpha-rename
metamorphic trials Spike 002 used (500 seeded trials preserving agreement)
`[CITED: .planning/spikes/002-cfg-edge-last-use/README.md:148-149]` against
the production `core.Function` shape, not just the spike's synthetic notation.

## Q5 — Public borrow origins (OWN-04)

**What a public borrowed result must record**, per Spike 003's validated
shape (`.planning/spikes/003-public-origins-generic-abilities/README.md:82-119`):

- **Origin path(s):** one or more `parameter[.field...]` paths naming which
  input(s) the returned borrow may have come from (`borrow(input)`,
  `borrow(left | right)` as a conservative union, or per-alternative origins
  on a tagged return like `Left(borrow(left) ...) | Right(borrow(right) ...)`).
- **Access mode**, carried **independently** of origin: `shared` or
  `exclusive` (`borrow(input)` vs `borrow mut(input)`).
- For a scoped/higher-ranked callback parameter: a **fresh quantified origin**
  that the type system proves cannot escape the one call
  (`for access fn(view: borrow(access) View[Byte]) -> R where R: escape`).

This must attach to `core.Function` as a new fact — recommend a
`PublicOrigin` struct on `core.Function` (sibling to `Match`/`Linear`, present
only when the function returns a borrowed type), not a field on `TypeFact`,
because origin is a per-function-signature fact about *which parameter* a
result derives from, not a property of the borrowed type itself. This is
`[ASSUMED]` design (not yet implemented in `core.go`, which currently has no
origin-related field at all — confirmed by reading the full 116-line file,
`[VERIFIED: internal/compiler/core/core.go:1-116]`) but follows directly from
Spike 003's validated contract.

**"Stale, omitted, or impossible" summary — what each looks like and how
detected**, directly reusing Spike 003's own producer-verification gates
(`.planning/spikes/003-public-origins-generic-abilities/README.md:122-125`,
"Producer verification rejects every deliberately dishonest summary"):

| Defect | Detection |
|---|---|
| **Omitted** — origin set understates real sources (e.g., `choose` declares only `left` but body can return from `right`) | Body-oracle recomputation (source-blind, like `corevalidate`) finds a body-derived origin not present in the declared set → reject at publish time, `core.origin_understated` (name TBD) |
| **Impossible** — declared exclusive but body only ever returns a shared loan, or vice versa | Same recomputation compares declared access mode to the body's actual `OpBorrowShared`/`OpBorrowExclusive` derivation → reject, `core.access_mismatch` |
| **Stale** — summary was generated against an old core artifact/digest and the implementation changed since | Content-digest binding exactly as Spike 004's certificate does: bind the summary to a SHA-256 digest of the typed-core artifact it was derived from; a consumer with a mismatched digest rejects before checking anything else, `evidence.core_mismatch`-shaped (mirrors the existing `evidence.core_mismatch` control already proven in `session.go:684`) |

**Interaction with sealed ability derivation:** origins and abilities are
independent facts checked by independent logic (Spike 003's design principle
#6: "`copy`, `drop`, `share`, `send`, and `escape` remain independent").
`[CITED: .planning/spikes/003-public-origins-generic-abilities/README.md:75-76]`
Concretely: a returned `View[Byte]` borrowed from a `Buffer` parameter has
`escape: false` on the view type (a borrowed view cannot outlive its origin)
regardless of what `Buffer`'s own abilities are — this is a *type-level*
ability fact on the view's synthetic type, derived structurally exactly as
`ability.go`'s `deriveStructural` already does for `Box`/`Pair`, not a new
mechanism. The origin path is orthogonal metadata riding alongside that
ability derivation on the same `core.Function`.

**Interaction with `lang.core/1`:** adding a `PublicOrigin` fact is an additive
field, which under `evidence.go`'s existing schema-branch pattern
(`if checked.Program.Schema == core.Schema1 { manifest.Schema = Schema1 ...
}`, `[VERIFIED: internal/compiler/evidence/evidence.go:215-219]`) can likely
stay on `lang.core/1` if the field is optional/`omitempty` and absent for
every existing Phase 1/2 program — see Q8 for the full schema-versioning
analysis.

## Q6 — Separate compilation (success criterion 4)

**Minimum viable form given one-module-per-file, no linker:** there is
currently no multi-file compilation at all — confirmed by the absence of any
import/module-resolution code path in the files read this session (`check.go`,
`corevalidate.go`, `core.go`, `session.go` all operate on a single
`ast.Program`/`core.Program` per invocation). Building an actual multi-file
linker or module resolver is out of scope per D-03's "no new subsystems" spirit
and is not named as in-scope by ROADMAP Phase 3 (which lists it as a success
*criterion about the summaries*, not about a build system).

**Recommendation:** simulate separate compilation the way Spike 003 and Spike
004 both already do — a body-blind **serialized interface artifact** (a
subset of `core.Program` containing only function signatures, `PublicOrigin`
facts, and ability facts, with `Linear`/`Match` bodies stripped) that a second,
independent process/test consumes without ever loading the producer's
`core.Program.Functions[].Linear`. This requires:

- **Schema:** a new `core.Interface` (or `core.PublicSummary`) type: `{Schema,
  ModuleID, CoreDigest string, Functions []core.FunctionSignature}` where
  `FunctionSignature` carries `{ID, Name, Parameter, ReturnType, PublicOrigin,
  Abilities}` and explicitly omits `Linear`/`Match`. `CoreDigest` binds the
  summary to the exact producing artifact, exactly as Spike 004's certificate
  binds to a SHA-256 digest of the typed-core artifact
  (`.planning/spikes/004-independent-certificate-checker/README.md:131-137`,
  "Iteration 1 — bind the statement, not just the proof").
- **Staying bounded:** the interface is linear in function/origin count (Spike
  003 measured a 1,000-origin union at ~79 KB, consumed in low single-digit
  milliseconds `[CITED: .planning/spikes/003-public-origins-generic-abilities/README.md:230-232]`)
  — cap origin-set size and function count exactly the way `MaxSourceBytes`
  already caps source (`internal/compiler/syntax` — referenced at
  `session.go:608,612` — `[VERIFIED: internal/compiler/session/session.go:608]`)
  and fail closed above the cap rather than truncate silently.
- **The "coordinated-frontend-lie limitation" retained explicitly:** success
  criterion 4 itself names this — Spike 004's own `KnownEscape` constant
  (`"escape:coordinated-source-core-lie"`,
  `[VERIFIED: internal/compiler/corevalidate/corevalidate.go:16-19]`, quoted:
  `const KnownEscape = "escape:coordinated-source-core-lie"`) is the exact
  production analog. Phase 3 should declare an equivalent
  `KnownEscape`-style constant on whatever package validates public-origin
  interfaces, and the verify harness should assert it appears in
  `result.ExpectedEscapes` (mirroring `session.go:582`,
  `result.ExpectedEscapes = []string{corevalidate.KnownEscape}`) rather than
  silently having no record of the limitation.

## Q7 — Debug-lineage bounded experiment (D-01..D-04)

The wiki's six steps
(`wiki/debug-evidence-symbolication-and-proof.md:163-170`):
1. Attach stable source/core IDs to one match arm, move, borrow, and return.
2. Emit a sidecar semantic map plus native debug info for O0 and O3.
3. Capture one deliberate panic/segfault and one ownership diagnostic.
4. Symbolize both to the same semantic IDs, with honest optimized-out fields.
5. Seed one stale-symbol, inlining, redaction, truncation, and wrong-build-ID
   defect; require each to fail closed.
6. Measure compile latency, binary/symbol size, capture time, symbolication
   time, output bytes, and AI root-cause/repair success.

**Recommendation — steps 1, 2, and 4 only, this phase:**

- **Step 1** is nearly free this phase: the CFG-liveness work already assigns
  stable IDs to every move/borrow/return/arm-edge (Q1-Q2). Extending those IDs
  through to a "semantic ID" side table costs almost nothing extra and is the
  literal seam D-01 says to preserve.
- **Step 2 (sidecar semantic map only, not native debug info):** emit a JSON
  side table mapping `{sourceSpan, coreID}` → `{operationID, pointID/edgeID}`.
  This is bounded, schema-stable, and reuses the existing evidence-manifest
  machinery pattern. Recommend **skipping the "native debug info for O0/O3"
  half of step 2 this phase** — DWARF/CodeView emission is explicitly D-03
  out-of-scope, and even a minimal "native debug info" would require touching
  `cgen` to emit `#line` directives or similar, which is real new surface on
  the native backend that the phase's actual requirements (OWN-03/OWN-04) do
  not need.
- **Step 4 (symbolize to the same IDs, honest `optimized_out`)** can be
  satisfied entirely from the *semantic* side (map a captured diagnostic or
  interpreter-trace event back to its `sourceSpan`/`coreID`) without ever
  touching native crash capture, satisfying D-04's "must report `available` /
  `optimized_out` / `not_captured` honestly."

**Explicitly rejected as out of scope under D-03, this phase:**
- **Step 3** ("capture one deliberate panic/segfault") — this requires an
  actual native crash-capture mechanism (even a minimal one touches signal
  handling / core-dump-adjacent code), which is the direct DWARF/crash-storage
  subsystem D-03 forbids. An **ownership diagnostic** capture (the other half
  of step 3) is already covered by step 4 using existing `diagnostic.Diagnostic`
  machinery — so step 3 can be satisfied for the diagnostic half only, and the
  panic/segfault half is rejected.
- **Step 5** (fault injection across stale-symbol/inlining/redaction/
  truncation/wrong-build-ID) presupposes the native debug-info artifacts that
  step 2's native half and step 3's crash half would have produced — since
  those are rejected, step 5 has no artifact to mutate this phase and is
  rejected wholesale.
- **Step 6** (compile/capture/symbolication latency measurement) is only
  meaningful once steps 2's native half and 3/5 exist — rejected for the same
  reason, though the *semantic*-side measurements (map emission cost, output
  bytes) can piggyback on the existing `recomputed_work`/`output_bytes`
  metrics infrastructure at no extra cost.

**This satisfies D-04's bounding requirement** because the reduced experiment
(semantic ID side table + honest-unavailable symbolization for diagnostics) has
a declared cap (bounded by existing `MaxSourceBytes`/function count), counted
work (existing `recomputed_work` convention), a stable schema (new
`lang.debug-map/0`-style schema, versioned like every other artifact), and no
new spawned process or file I/O beyond what the compiler already does — so it
introduces no new D-15 boundedness surface.

## Q8 — Schema versioning

**Current versioned schemas, verified by reading each source this session:**

| Schema | v0 | v1 | Source |
|---|---|---|---|
| `core` | `lang.core/0` (match bodies) | `lang.core/1` (linear bodies) | `[VERIFIED: internal/compiler/core/core.go:6-8]` |
| `diagnostic` | `lang.diagnostic/0` | `lang.diagnostic/1` | `[VERIFIED: internal/compiler/diagnostic/diagnostic.go:12-13]` |
| `evidence` | `lang.evidence/0` | `lang.evidence/1` | `[VERIFIED: internal/compiler/evidence/evidence.go:29-30]` |
| `execution` | (Schema1 referenced) `lang.execution/1` | — | `[VERIFIED: internal/compiler/evidence/evidence.go:218]`, quoted: `manifest.ExecutionSchema = execution.Schema1` |
| `protocol` (command) | `lang.command/0` | none yet | `[VERIFIED: internal/compiler/protocol/protocol.go:14]` |

**Recommendation:** loans/origins are an **additive** extension of the linear
body, exactly parallel to how `lang.core/1` was an additive extension over
`lang.core/0`'s match-only shape (`Match *Match` and `Linear *LinearBody` are
already both optional pointer fields on `core.Function`,
`[VERIFIED: internal/compiler/core/core.go:32-33]`). The existing schema-branch
pattern in `evidence.go` (`if checked.Program.Schema == core.Schema1 {...}`)
gives two viable paths:

1. **Keep `lang.core/1`, add optional fields** (`PublicOrigin *PublicOrigin` on
   `core.Function`, a `LoanID`/access-mode field on `core.LinearOperation`, a
   new `OpBorrowExclusive` operation kind). Every existing Phase 1/2 program
   omits these fields (`omitempty`), so serialized bytes for those programs
   are unchanged — satisfying D-13's "byte-identical" for Phase 1 and "moves
   only causally" for Phase 2. This is the recommended path: it avoids a
   schema-version bump entirely for a genuinely additive, opt-in extension,
   matching how `core.Schema1` itself introduced `LinearBody` as an additive
   optional field rather than bumping past it to a `/2`.
2. **Bump to `lang.core/2`** only if the CFG restructuring (Q1's match-arm-body
   change) requires *changing* an existing required field's meaning (e.g., if
   `MatchArm.Value` stops being a bare string and becomes a sum type). Given
   Q1's recommendation keeps the bare-name form legal as one variant, this can
   likely be modeled as an additive `Body *LinearBody` field alongside the
   existing `Value string` field (mutually exclusive, like `Match`/`Linear`
   are today) — **avoiding a version bump** is achievable and preferred, since
   `evidence.go`'s schema-branch logic (`if v.program.Schema == core.Schema1`
   gates, `[VERIFIED: internal/compiler/corevalidate/corevalidate.go:106,109]`)
   would otherwise need a third branch throughout `corevalidate.go`, growing
   the surface every consumer must handle.

**How Phase 1/2 artifacts stay byte-identical (D-13):** Go's `encoding/json`
omits `omitempty` fields entirely rather than emitting `null`/`{}`, so as long
as every new field added for Phase 3 is `omitempty` and Phase 1/2 fixtures
never populate it, the marshaled bytes for `testdata/phase1/*` and
`testdata/phase2/owned_transfer.golden.c`-equivalent JSON goldens are
unaffected. This must be verified per-field by the planner/executor exactly as
`02-OVERRIDES.md` did for OV-02-01 ("Verified key-by-key as parsed JSON;
`source_digest` and `c_digest` unchanged") — a byte-diff assertion in a test,
not an assumption.
`[VERIFIED: .planning/phases/02-owned-values-and-abilities/02-OVERRIDES.md:52]`

## Recommended approach (summary)

1. **Extend `match` arms to optionally carry a full linear body** in the value
   position, giving the CFG its first real branch/join, instead of adding a
   separate `if` construct. Scope OWN-03 to **acyclic** (branch-only) CFGs this
   phase; loops require a second new control construct and are out of reach.
2. **Implement backward worklist dataflow liveness independently in both
   `check.go` and `corevalidate.go`**, replacing each layer's own O(N²)
   transitive-loan-set logic (`discoverLoanLastUses` in `check.go`;
   `loansForPlace`/`replay` in `corevalidate.go` — both confirmed quadratic
   this session, only the first was previously recorded as debt). Count one
   work unit per block-transfer evaluation so `recomputed_work` stops
   understating cost (D-05).
3. **Add exclusive loans** (`OpBorrowExclusive`, `borrow mut` surface syntax)
   with the standard shared/shared-accept, shared/exclusive-reject,
   exclusive/exclusive-reject, loan-vs-move-reject conflict matrix; the 7
   fixtures in Q3 are the minimal accept/reject corpus.
4. **Build a genuinely separate bounded-path-enumeration oracle** in a new
   package that never calls production liveness code, and mutation-kill it by
   reverting the production dataflow commit and confirming the differential
   fails (D-09's explicit bar, which Phase 2's own ownership-oracle failure
   shows is not optional).
5. **Add a `PublicOrigin` fact** on `core.Function` (origin paths, access mode,
   optional fresh-quantifier marker) as an additive `omitempty` field, checked
   by a source-blind extension of `corevalidate.go` following Spike 003's
   producer-verification gates, and expose it as a body-stripped `core.Interface`
   summary bound to a content digest for the separate-compilation criterion,
   following Spike 004's compact-recompute-only certificate shape — explicitly
   retaining and surfacing the `coordinated-source-core-lie` escape as an
   `ExpectedEscapes` entry in the verify harness, not silently dropping it.
6. **Debug-lineage experiment: steps 1/2(semantic-only)/4 of the wiki's six**,
   explicitly rejecting steps 2's native-debug-info half, 3's panic/segfault
   half, 5, and 6 as requiring the DWARF/crash-capture subsystem D-03 forbids.
7. **No schema-version bump required** — model both CFG bodies and public
   origins as additive `omitempty` fields on `lang.core/1`, verified
   byte-identical against Phase 1/2 goldens per-field, exactly as OV-02-01 was
   verified.

## Runtime State Inventory

Not applicable — Phase 3 is not a rename/refactor/migration phase. No stored
data, live service config, OS-registered state, secrets, or build-artifact
renames are implicated. **Confirmed by category:**
- Stored data: None — the compiler has no datastore; all state is per-invocation.
- Live service config: None — there is no running service; `lang` is a CLI.
- OS-registered state: None.
- Secrets/env vars: None new; `LANG_OBSERVE_TIMING` (existing, `verify-phase2.sh:25`) is unaffected.
- Build artifacts: None renamed this phase; D-06's `__LANG_`→`_LANG_` rename (a
  carried Phase 2 debt item with a Phase 3 deadline) is a build-artifact-naming
  change and **should be scheduled as its own standalone commit before any new
  C golden is frozen**, per D-06's explicit text.

## Common Pitfalls

### Pitfall 1: Reusing the same dataflow implementation for both admission layers
**What goes wrong:** `check.go` and `corevalidate.go` end up calling a shared
helper function for CFG liveness, silently violating D-12.
**Why it happens:** the logic is genuinely identical in shape (both need
backward loan liveness over a CFG), so extracting a shared package feels like
good engineering.
**How to avoid:** write two independent implementations, even if structurally
similar, exactly as the two admission layers already independently implement
ability derivation (`ability.go`'s `deriveStructural` vs.
`corevalidate.go`'s free-function `deriveAbility` — confirmed as separately
coded, `[VERIFIED: internal/compiler/corevalidate/corevalidate.go:376-427]`
vs. `[VERIFIED: internal/compiler/ability/ability.go:96-173]`).
**Warning signs:** a new shared package under `internal/compiler/cfg/` or
similar imported by both `check` and `corevalidate`.

### Pitfall 2: Treating the bounded-path oracle as "the same algorithm, different code"
**What goes wrong:** the oracle is written to mirror the production dataflow's
per-block logic almost line-for-line, so a bug in the shared *mental model*
(not the code) escapes both.
**Why it happens:** it's the path of least resistance once the production
algorithm is well understood.
**How to avoid:** the oracle should be *exhaustive path enumeration*, a
different mechanism class entirely, not "dataflow but written twice." Apply
D-09's mutation-kill test as a gate before considering the oracle done.
**Warning signs:** the oracle package imports `core.LinearOperation`-shaped
intermediate structures identical to the production analyzer's internal state.

### Pitfall 3: Letting public-origin verification import the checker
**What goes wrong:** to check whether a declared origin is honest, it seems
natural to call `check.Program(...)` again on the producer's source.
**Why it happens:** the checker already has all the ownership facts computed.
**How to avoid:** origin verification must operate on the **typed-core
artifact only** (like `corevalidate.go` already does for everything else),
never on source or the checker's internal state — this is the entire point of
Spike 004's "source- and body-blind" design.
**Warning signs:** the new origin-verification code imports
`internal/compiler/check` or `internal/compiler/ast`.

### Pitfall 4: Silently widening `lang.core/1` in a way that breaks Phase 1 goldens
**What goes wrong:** a new required (non-`omitempty`) field on `core.Function`
or `core.LinearOperation` changes every existing golden's JSON bytes.
**Why it happens:** Go's `encoding/json` requires explicit `omitempty` tags;
forgetting one on a new field is silent until a byte-diff test catches it.
**How to avoid:** every new field added this phase must be `omitempty` and the
plan must include an explicit byte-diff assertion against
`testdata/phase1/*` and `testdata/phase2/*` goldens, mirroring how OV-02-01 was
verified "key-by-key as parsed JSON."
**Warning signs:** `go test` passes but a manual `git diff` on a golden file
shows unexpected bytes.

## Code Examples

### CFG-edge endpoint identity scheme (adapted from Spike 002's validated shape)
```text
// Source: .planning/spikes/002-cfg-edge-last-use/README.md:88-89 (CITED — spike notation, not production code)
edge:entry:else:view
point:then:0:view
```
Recommended production analog, following the existing `{functionID}:{kind}:{ordinal}` ID scheme already used throughout `check.go` (e.g. `functionID + ":point:entry"`, `[VERIFIED: internal/compiler/check/check.go:120]`):
```text
{functionID}:edge:{armIndex}:{loanID}
{functionID}:point:arm:{armIndex}:{operationIndex}:{loanID}
```

### Existing straight-line last-use pattern to preserve as the oracle's linear-segment building block
```go
// Source: internal/compiler/check/check.go:366-399 (VERIFIED — read this session)
func discoverLoanLastUses(parameterName string, body *ast.LinearBody) map[int]loanUse {
	uses := make(map[int]loanUse)
	visible := map[string]int{parameterName: -1}
	loansForBinding := make(map[int][]int)
	for index, binding := range body.Bindings {
		var inherited []int
		if sourceBinding, ok := visible[binding.RHS.Source]; ok {
			for _, loanIndex := range loansForBinding[sourceBinding] {
				if use, tracked := uses[loanIndex]; tracked {
					use.index = index
					use.span = binding.RHS.Span
					uses[loanIndex] = use
				}
				inherited = append(inherited, loanIndex)
			}
		}
		// ...
	}
	return uses
}
```
This function should be **kept** (not deleted) as the per-path linearization
step inside the new bounded-path oracle (Q4), since each enumerated path is a
straight-line sequence and this logic already correctly computes last-use on
straight lines — it is the CFG-wide use of it (calling it once per whole
function instead of once per path) that produces the O(N²) blowup and is
being replaced in the *production* analyzer.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|---|---|---|---|
| Transitive per-binding loan-set propagation (`discoverLoanLastUses`, `loansForPlace`) | Backward worklist dataflow over per-block live-sets | This phase (recommended) | Removes Θ(N²) blowup (D-02-03/D-05); makes `recomputed_work` honest |
| Shared-only loans | Shared + exclusive loans with a conflict matrix | This phase (OWN-03 requirement) | New `OpBorrowExclusive`, new surface syntax, new diagnostic code |
| No public borrow metadata | `PublicOrigin` fact + body-stripped `core.Interface` summary | This phase (OWN-04 requirement) | Enables separate compilation of borrow-returning functions |

**Deprecated/outdated:** none — this phase extends rather than replaces any
shipped Phase 1/2 behavior; the D-13 constraint requires exactly this framing.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|---|---|---|
| A1 | Extending `match` arms with an optional linear body (rather than adding `if`) is the right minimal branching construct | Q1 | If wrong, the planner may need a genuinely new `if`/loop grammar, doubling surface-syntax scope; this is explicitly flagged as Claude's Discretion in CONTEXT.md and should be confirmed with the user before locking |
| A2 | Kildall's worklist algorithm is the correct historical attribution for the backward-dataflow technique | Q2 | Cosmetic only — the technique's validity does not depend on correct historical attribution; the substantive claim is `[CITED]` from Spike 002's own sources |
| A3 | `PublicOrigin` should be a new struct field on `core.Function` rather than embedded in `TypeFact` | Q5 | If wrong, the planner reshapes where the fact lives, but the underlying origin/access-mode/escape-hatch contract from Spike 003 is unaffected |
| A4 | A schema-version bump to `lang.core/2` can be avoided by making CFG/origin fields additive `omitempty` fields | Q8 | If the `MatchArm.Value` sum-type change turns out non-additive in practice, a `/2` bump becomes necessary; low risk since Match/Linear already coexist as optional fields on Function |
| A5 | No external package (e.g. `golang.org/x/tools/go/analysis`) is needed to fix D-02-01's spawn-guard | Package Legitimacy Audit | If the planner chooses `go/analysis`, it must run the full Package Legitimacy Gate; flagged as a checkpoint either way |

**If this table is empty:** N/A — assumptions listed above.

## Open Questions

1. **Exact surface syntax for exclusive borrow and match-arm-bodies**
   - What we know: the semantic contract (`borrow mut`, arm-body CFG shape)
     is validated by Spikes 002/003.
   - What's unclear: exact keyword/token choices are explicitly marked
     "experimental" even in the validated spikes
     (`.planning/spikes/003-public-origins-generic-abilities/README.md:117`,
     "The spelling is experimental.").
   - Recommendation: leave concrete spelling to the planner/executor; do not
     lock syntax bikeshedding into this research.

2. **Whether the fresh-quantified-origin (higher-ranked callback) case belongs
   in Phase 3 at all**
   - What we know: Spike 003 validated one higher-ranked callback pattern.
   - What's unclear: ROADMAP Phase 3's four success criteria do not mention
     callbacks or higher-ranked access explicitly; only "field/alternative
     origins and access modes."
   - Recommendation: treat higher-ranked callback origins as optional/stretch
     for this phase, not a required success-criterion item — the four ROADMAP
     criteria are satisfiable without it.

3. **Is criterion 4's "separate compilation" satisfiable as a same-process
   simulation, or does it require actual separate `go test` / CLI invocations?**
   - What we know: Spikes 003/004 both simulate this in-process with a
     body-stripped struct passed to a separate package.
   - What's unclear: whether the plan-checker/verifier will accept an
     in-process simulation as satisfying "separate compilation," or require
     two literal separate `lang` CLI invocations (produce interface file,
     then consume it in a second process).
   - Recommendation: prefer two separate CLI invocations (matching how
     `verify-phase2.sh` already shells out to a built binary) since it is a
     stronger, more honest demonstration and costs little extra given the CLI
     already exists.

## Environment Availability

Not applicable — Phase 3 is pure Go compiler code with no new external tool,
service, or runtime dependency beyond what Phase 1/2 already require (Go
1.24 standard library, Clang for native backend — both already verified
present and used by the existing `verify-phase2.sh` gate).

## Validation Architecture

### Test Framework
| Property | Value |
|---|---|
| Framework | Go standard `testing` package (`go test ./...`), matching Phase 1/2 |
| Config file | none — Go's built-in test discovery; bounded gate script is `scripts/verify-phase2.sh` (to be extended to `verify-phase3.sh` or generalized) |
| Quick run command | `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/...` |
| Full suite command | `go test ./...` then `sh scripts/verify-phase2.sh`-equivalent for Phase 3 |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|---|---|---|---|---|
| OWN-03 | Shared/exclusive conflict accept/reject matrix (Q3's 7 fixtures) | unit + differential | `go test ./internal/compiler/check/... -run TestLoanConflict` | ❌ Wave 0 |
| OWN-03 | CFG-edge-specific last use, omitted-edge detection | property + differential vs. bounded-path oracle | `go test ./internal/compiler/check/... -run TestCFGLastUse` | ❌ Wave 0 |
| OWN-03 | `recomputed_work` no longer Θ(N²) | scale/regression | `go test ./internal/compiler/check/... -run TestLivenessWorkScale` | ❌ Wave 0 |
| OWN-04 | Public origin summary agrees with body oracle; dishonest summaries rejected | differential + mutation | `go test ./internal/compiler/corevalidate/... -run TestPublicOrigin` | ❌ Wave 0 |
| OWN-04 | Separate-compilation summary rejects stale/omitted/impossible; coordinated-lie escape retained | integration (CLI-level, two invocations per Open Question 3) | `go run ./cmd/lang -- <interface-export/import flow>` | ❌ Wave 0 |

### Sampling Rate
- **Per task commit:** targeted `go test ./internal/compiler/{check,corevalidate}/...`
- **Per wave merge:** `go test ./...`, `go test -race ./...`, `go vet ./...`
- **Phase gate:** full `verify-phase3.sh`-equivalent (extending `verify-phase2.sh`'s pattern) green before `/gsd-verify-work`

### Wave 0 Gaps
- [ ] New test files for loan-conflict matrix, CFG last-use differential,
      liveness work-scale regression (all under `internal/compiler/check/`)
- [ ] New test files for public-origin differential/mutation matrix (under
      `internal/compiler/corevalidate/` or a new sibling package)
- [ ] New bounded-path-oracle package with its own test suite, structured to
      allow the D-09 "revert-and-confirm-failure" mutation-kill check
- [ ] 7 new `.lang` fixtures under `testdata/phase3/` (Q3's matrix) plus a
      match-arm-body branch fixture pair (Q3 items 6-7)
- [ ] Extension of `scripts/verify-phase2.sh`'s pattern into a Phase 3
      equivalent, adding new required controls to the fail-closed set
      (mirroring the existing 9-item `requiredControls` list at
      `internal/compiler/session/session.go:690-700`)

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---|---|---|
| V2 Authentication | no | N/A — compiler has no auth surface |
| V3 Session Management | no | N/A |
| V4 Access Control | no | N/A |
| V5 Input Validation | yes | Existing bounded-input pattern: `MaxSourceBytes`, work limits, timeouts (`readBoundedFile`, `syntax.MaxSourceBytes` — `[VERIFIED: internal/compiler/session/session.go:608,612]`). New CFG/origin inputs must be bounded the same way — cap block count, path-enumeration count (bounded-path oracle), and origin-union size (Q6) |
| V6 Cryptography | yes | Content digests for stale-summary detection (Q5/Q6) must reuse the existing SHA-256 digest pattern already proven in `evidence.go`/Spike 004 — never hand-roll a new hash or comparison scheme |

### Known Threat Patterns for this stack
| Pattern | STRIDE | Standard Mitigation |
|---|---|---|
| Unbounded path enumeration in the bounded-path oracle (exponential blowup on adversarial CFGs) | Denial of Service | Explicit path-count cap with fail-closed rejection above the cap, exactly as the ability deriver already caps at `maxAbilityNodes = 4096` (`[VERIFIED: internal/compiler/ability/ability.go:14-16]`) |
| A dishonest public-origin summary understating borrow scope (the class OWN-04 exists to prevent) | Tampering | Body-oracle recomputation + content-digest binding (Q5), not trust-on-declaration |
| Coordinated frontend/summary lie (Spike 004's documented, accepted escape) | Tampering (accepted residual) | Explicitly declared as `ExpectedEscapes`, not silently ignored — must remain a documented, tested-for escape, not "solved" |
| Reintroducing the D-02-01 unbounded-spawn defeat pattern while touching `native`/build tooling incidentally this phase | Denial of Service | If the planner folds the D-02-01 fix into this phase's touched files (per D-08's "fold into whatever plan already touches its file"), use an AST-resolution check, not a substring scan |

## Sources

### Primary (HIGH confidence — read this session)
- `internal/compiler/ast/ast.go` — full 85 lines read; confirms no branching AST node exists
- `internal/compiler/check/check.go` — full 489 lines read; confirms straight-line liveness and its O(N²) shape
- `internal/compiler/corevalidate/corevalidate.go` — full 494 lines read; confirms independent O(N²) shape and source-blind design
- `internal/compiler/ability/ability.go` — full 182 lines read
- `internal/compiler/core/core.go` — full 116 lines read
- `internal/compiler/session/session.go` (lines 560-720) — verify-lane control-table pattern
- `scripts/verify-phase2.sh` — full 49 lines read
- `internal/compiler/evidence/evidence.go`, `evidence_test.go`, `diagnostic.go`, `protocol.go` (grep-targeted schema constants)
- `.planning/phases/02-owned-values-and-abilities/02-DEBT.md`, `02-OVERRIDES.md` — full reads
- `.planning/config.json` — confirms `nyquist_validation: true`, `security_enforcement: true`, all search providers `false`

### Secondary (MEDIUM confidence — cited from in-repo validated spike documents, which themselves cite official sources)
- `.planning/spikes/002-cfg-edge-last-use/README.md` — CFG liveness algorithm, edge-endpoint identity scheme, mutation-kill methodology; itself cites rustc-dev-guide, NLL RFC, Polonius, Clang dataflow docs
- `.planning/spikes/003-public-origins-generic-abilities/README.md` — public origin contract, access-mode independence, ability independence; itself cites Mojo, Rust elision/HRTB, Swift noncopyable/nonescapable/borrowing-iteration, Move abilities
- `.planning/spikes/004-independent-certificate-checker/README.md` — compact recompute-only certificate, `KnownEscape` shape; itself cites proof-carrying code, Lean, WebAssembly validation, Alive2/CompCert
- `wiki/ownership-evidence-roadmap.md`, `wiki/debug-evidence-symbolication-and-proof.md` — full reads

### Tertiary (LOW confidence — training knowledge, not verified this session)
- Attribution of the backward-worklist dataflow technique to "Kildall's algorithm" (A2) — no web search was run because no search provider is configured for this project.

## Metadata

**Confidence breakdown:**
- Standard stack: N/A — no new external packages
- Architecture (CFG/liveness/origins): HIGH — directly grounded in two VALIDATED and one PARTIAL in-repo spike plus this session's direct code reads
- Pitfalls: HIGH — derived directly from D-09/D-12's explicit text and Phase 2's own documented gate failures
- Surface syntax choice (Q1's match-arm-body recommendation): MEDIUM — sound reasoning but explicitly flagged as Claude's Discretion needing user confirmation

**Research date:** 2026-09-04
**Valid until:** Stable until Phase 3 code lands; re-verify code:line citations if any of `check.go`/`corevalidate.go`/`core.go`/`ast.go` change before planning starts.
