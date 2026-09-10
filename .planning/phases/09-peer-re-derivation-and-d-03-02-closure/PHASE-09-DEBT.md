---
phase: 09-peer-re-derivation-and-d-03-02-closure
recorded: 2026-09-10
status: accepted
disposition: planning-time
items: 14
blocking: 0
---

# Phase 09 — Declared Deferred Scope, Reversals, and Carried Resolutions (D-09-44)

**Written:** 2026-09-10, at *planning* time — not at phase end.

Per RETROSPECTIVE Key Lesson 4: declare deferred scope in writing at the moment
it is decided. Shape follows the mechanically-checked debt-register format
`TestDebtRegistersAreWellFormed` (`internal/compiler/session/session_test.go`)
enforces on every `*-DEBT.md` register: an `items:` count matching the `## Items`
table, one `### <ID>` detail section per row, a severity from the closed
vocabulary (blocker, warning, info), and a non-empty landing phase.

Three of the rows below are **resolutions of inherited Phase 08 debt**
(D-08-26 → D-09-29, D-08-27 → D-09-14, D-08-40 → D-09-03). Two are **formal
reversals of written Phase 08 commitments** (D-09-08 reverses D-08-41's
disposition; D-09-31 supersedes D-08-21's promotion commitment). Both reversals
name the superseded text so no future reader has to discover the change.

---

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-09-03 | 09-CONTEXT.md (D-09-03), resolves PHASE-08-DEBT.md D-08-40 | OWN-07, TRU-04 | info | Phase 09 — plan 09-01 (required assertion, not a side effect) | RESOLUTION of D-08-40. `buildLoanChainIndex` (`corevalidate.go:1133`) propagates `parent[TargetID] = SourceID` through every `core.OpCall` with no callee-signature consultation. Landing the peer's contract-aware consult must retire BOTH `peerDivergenceExpected` entries — `testdata/phase08/twin_a_accept.lang` and `testdata/phase08/relay_depth2_accept.lang` (`session_peer_gate_test.go:56-58`). Retiring both is a required assertion of this phase |
| D-09-08 | 09-CONTEXT.md (D-09-08), REVERSES PHASE-08-DEBT.md D-08-41 | OWN-09 | warning | Phase 09 — plan 09-08 (the deletion), gated by plan 09-07's mid-phase gate | REVERSAL. D-08-41 adjudicated `computeLoanLastUses`' summary-blindness as "an accepted, permanent, disclosed scope limitation… Landing phase: Not scheduled." That disposition is SUPERSEDED: `computeLoanLastUses` (`check.go:3743`) and its `shadow:place:*` synthesis are DELETED in Phase 09, because a summary-blind decision point actively masks the law the milestone exists to prove (it refuses both Pattern B twins identically before the interprocedural pass runs). Reversibility: costly — restoring the shadow path would mean reconstructing the `shadow:place:*` synthesis and re-threading two call sites through lowering. Deliberately so |
| D-09-13 | 09-CONTEXT.md (D-09-13), 09-RESEARCH.md Pitfall 4 | OWN-09, TRU-04 | warning | Phase 09 — plan 09-06 authors the gate; plan 09-08 must keep it green | Moving `ownership.move_while_borrowed` / `ownership.borrow_conflict` emission from per-function lowering to the post-assembly pass changes WHICH diagnostic is reported first for a program with errors in two functions. `Code + Span + Causes` fold into a SHA-256 diagnostic identity (`diagnostic/diagnostic.go:96-112`), so an emission-order change moves a published, agent-facing diagnostic ID for an ordering reason rather than a content reason. Requires an explicit ordering-stability assertion, never a corpus replay |
| D-09-14 | 09-CONTEXT.md (D-09-14), resolves PHASE-08-DEBT.md D-08-27 | OWN-09 | info | Phase 09 — plan 09-09 (document change) | RESOLUTION of D-08-27. The OWN-09 Phase-08-vs-09 document conflict closes in favour of `REQUIREMENTS.md:173`'s Phase 09 mapping, which stands as operative. Phase 08 is shipped (`4ab5c65`) and reopening it is not viable. OWN-09's OWN REQUIREMENT TEXT is what gets corrected to say Phase 09, exactly as Phase 08's mid-phase gate anticipated |
| D-09-21 | 09-CONTEXT.md (D-09-21), PHASE-07-DEBT.md D-07-33 | OWN-08, SEM-06 | warning | Phase 09 — plan 09-03 executes; fallback lands in this register with a named landing phase if triggered | DECLARED FALLBACK, with its trigger fixed in advance. If EXECUTION — not plan-time analysis — discovers the peer's `OpForeignCall` origin-omitted class is genuinely unbounded (it touches multi-hop foreign chains spike 005 did not clear), then and only then: split D-07-33 into two rows, close the two body-only classes (`core.origin_understated`, `core.origin_access_mismatch`) in Phase 09, cut foreign-origin-omitted peer re-derivation with a NAMED landing phase, and amend REQUIREMENTS.md to carve that class explicitly out of OWN-08's scope. OWN-08 is on the never-cut list, so narrowing any slice of its supporting peer requires explicit written sign-off, never a silent scope read |
| D-09-29 | 09-CONTEXT.md (D-09-29/D-09-30), resolves PHASE-08-DEBT.md D-08-26 | OWN-06 (Phase 08 success criterion 4) | info | Resolved at Phase 09 planning — no vehicle, by design; not deferred | RESOLUTION of D-08-26. The accepted-side consulted field set is a COMPILE-TIME CONSTANT: the consultation loop (`check.go:594-609`) fires exactly `return.mode` and `parameters[0].mode` for every function with a declared signature, corpus-wide, already proven a closed set by `TestInterproceduralDisclosedFieldSet` (`check_test.go:4898`). A runtime record would reprint a constant on every run at real `qlt02`-gated cost with zero incremental per-program information. All three Phase 08 vehicles stay ruled out; the fourth candidate (an opt-in flag mirroring `evidence --validate`'s `Trace *TraceSummary`) is buildable and still rejected on the constant-payload argument. D-08-26's "declared debt" framing is REPLACED with "resolved — no vehicle, by design" so no future phase reopens it as unfinished. Phase 09's positive obligation instead is D-09-30: `corevalidate` gets its own independently-written closed-field-set test and a cross-peer test asserting the two sets are identical |
| D-09-31 | 09-CONTEXT.md (D-09-31/D-09-32/D-09-33), SUPERSEDES 08-CONTEXT.md D-08-21 | OWN-07, TRU-04 | warning | Superseded at Phase 09 planning — closed, not deferred; reopen only under the condition named below | SUPERSESSION. Phase 08 committed in writing to promoting `check.interprocedural_loan_liveness` to `core.interprocedural_loan_liveness` in Phase 09 "at the moment `corevalidate` independently re-derives the same fact and both peers must agree on-code." That trigger is RETIRED as the promotion criterion because it was never achievable — D-08-17's own reasoning (the peer computes liveness through a reachability closure, not a worklist, so it cannot fail the same way) is general to the two mechanisms, not scoped to the iteration bound. LOCKED: `check.interprocedural_loan_liveness` stays `check.*`; `corevalidate` keeps `core.move_while_borrowed` with its own existing causes; neither is renamed, aliased, or merged. Reversibility: one-way if reversed later — promotion would move every existing diagnostic ID for the code, and diagnostic codes are published agent-facing API consumed by `lang-repair` and `lang explain` |
| D-09-37 | 09-CONTEXT.md (D-09-37), REQUIREMENTS.md:169 | OWN-05 | warning | Phase 09 — plan 09-09 records the split; OWN-05b lands in Phase 10 | OWN-05's text names THREE derivers (`check`, `corevalidate`, `interp`); Phase 09 delivers two. OWN-05 must NOT be flipped to Complete at Phase 09's end — that would be exactly the requirement-vs-code overclaim the debt registers exist to catch. Disposition chosen at planning time: SPLIT into OWN-05a (Phase 09: `check` + `corevalidate`) and OWN-05b (Phase 10: `interp`, verified there by TRU-03 / NAT-06), rather than a single partial row, so the traceability table's one-requirement-one-phase discipline survives |
| D-09-40a | 09-CONTEXT.md (D-09-40/D-09-40a), inventory run 2026-09-10 | QLT-07 | info | Phase 09 — plan 09-09 (closure section); threshold pre-registered before planning began | QLT-07's committed-vs-stretch status was decided against a threshold FIXED BEFORE the count was taken: committed iff the inventory requires ≤ 1 additional plan and opens zero packages or files OWN-07/OWN-08 do not already touch. Inventory result: all 33 tests named across the 12 loan-liveness-scoped rows (03-03/03-04/03-05 only, excluding OWN-04's 03-06/03-07 rows) already exist and pass. VERDICT: QLT-07 is COMMITTED. The debt is a ratification/documentation debt, not a test-authorship debt; the work touches only `.planning/` documents and ZERO production files. Scope-cut order item 2 is NOT triggered. Caveat preserved: this covers only the loan-liveness subset — `03-VALIDATION.md`'s own `nyquist_compliant: false` stays false (D-09-41), its OWN-04 rows are outside the claim, and its loop-carried-liveness clause is structurally un-closable while the language has no loops |
| D-09-43 | 09-CONTEXT.md (D-09-43), ROADMAP.md M002 scope-cut order | all six phase requirements | info | Phase 09 — adjudicated at plan 09-07's mid-phase gate | DECLARED SCOPE-CUT TRIGGER. If the peer-derivation + D-07-33-closure work exceeds ~2x its initial plan estimate, the cut order is (1) QLT-07 per D-09-40's threshold, then (2) the `OpForeignCall` slice of D-07-33 per D-09-21's named fallback. NEVER the criterion-1 differential corpus (D-09-23), never the seeded fault and its companion assertion (D-09-25), never the `computeLoanLastUses` deletion (D-09-08), and never the two `peerDivergenceExpected` retirements (D-09-03). Any cut is a deferral with a named landing phase, never a silent drop |
| D-09-45 | 09-CONTEXT.md (D-09-45), 09-RESEARCH.md Pitfall 5 | QLT-01, OWN-07 | info | Phase 09 — plans 09-01 and 09-09 record the corrections; the spike-registry row closes opportunistically in 09-09 | STALE PLANNING-DOCUMENT REFERENCES, corrected here rather than left to mislead. (a) `08-CONTEXT.md`'s D-08-09 cites `computeLoanLastUses` at `check.go:2979`; it is at `check.go:3743`. (b) `08-CONTEXT.md`'s D-08-03 asserts `callgraph.Order` is callee-before-caller; it is caller-before-callee (already D-08-42). (c) An earlier revision of D-09-03 wrote the two retiring fixtures under `testdata/phase07/`; the shipped map keys are under `testdata/phase08/`. (d) D-08-43's `.planning/spikes` registry gap (spike 006 has a directory and no registry row, so `TestQLT01RegistryCoversAllFiveSpikes` fails) is still open and still un-owned; it is cheap to close while Phase 09 touches the spike table under D-09-39 and is recommended, not required |
| D-09-46 | 09-CONTEXT.md (D-09-46), 09-RESEARCH.md Pitfall 3 | TRU-04 | info | Phase 09 — plan 09-02 Task 1 (Wave 0 prerequisite, blocks TRU-04's vehicle) | CORPUS-VISIBILITY PREREQUISITE. `generateCallGraphCorpus` is unexported inside `internal/compiler/check/costcorpus_test.go:50`, so Go's package-visibility rules make it structurally uncallable from any other package's test binary. D-09-23's "same generator, disjoint consumers" cannot be satisfied until it is made reachable. Resolution chosen at planning time: RELOCATE it (not duplicate it) to `internal/compiler/testsupport`, which imports only stdlib today and therefore cannot become a route by which `session` or `corevalidate` gains a production import of `check` |
| D-09-50 | 09-PLAN planning pass 2026-09-10 (new finding, not in 09-CONTEXT.md or 09-RESEARCH.md) | TRU-04 | warning | Phase 09 — plan 09-01 hosts the synthetic-shape differential; plan 09-07's gate reviews the split | D-09-23 prescribes feeding the synthetic call-graph shapes into `session_peer_gate_test.go`'s existing corpus-wide gate. VERIFIED AT PLANNING TIME THAT THIS IS NOT POSSIBLE AS WRITTEN: `check`'s only exported entry point is `check.Program(ast.Program)` (`check.go:46`), so no package outside `check` can compute a `check`-side verdict for a synthetically constructed `core.Program`. The `.lang` corpus gate therefore stays exactly where it is (and is where the two D-09-03 retirements are proven), while the SYNTHETIC-shape differential must live inside package `check`, which alone can reach both the unexported post-assembly pass and `corevalidate.Validate`. This is a split of VEHICLE, never of TRUTH: the synthetic differential must reuse the same both-directions exactness discipline and cross-reference `peerDivergenceExpected`'s doc comment, and must never define "divergence" a second way |
| D-09-51 | 09-01 execution pass 2026-09-10 (new finding, not in 09-CONTEXT.md or 09-RESEARCH.md) | OWN-07, TRU-04 | warning | Discovered plan 09-01; NOT landed there — out of that plan's file scope (`originvalidate.go` untouched); no landing phase yet named | NEWLY DISCOVERED, PREVIOUSLY-MASKED DEFECT in a THIRD validator. Once `buildLoanChainIndex`'s peer-derived OpCall consult (D-09-03) lands and `corevalidate.Validate` correctly stops refusing `testdata/phase08/twin_a_accept.lang` and `testdata/phase08/relay_depth2_accept.lang`, running either fixture through the FULL `lang check` CLI (`session.CheckCommandFile`, which additionally consults `originvalidate.ValidatePublished` after `check` and `corevalidate` both pass) surfaces a DIFFERENT, PRE-EXISTING refusal: `core.origin_omitted`. Root cause verified directly: `originvalidate.walkReturnOrigin` (`originvalidate.go:171-218`) has no `case core.OpCall` in its backward-walk switch, so it treats a call boundary as fully transparent, walking straight through `operation.SourceID` into the CALL'S ARGUMENT's own provenance — it never consults the CALLEE's own declared return contract the way `check`'s `derivePlaceLoans`/`corevalidate`'s new `derivePeerLoanCarry` both now do. For `twin_a_accept.lang`, `escortee` declares a plain owned return (`-> Buffer`), so `escort`'s own return is genuinely a fresh owned value with no live alias risk — but `originvalidate` still reports it as borrow-derived (walking through the call to the pre-call exclusive borrow) and refuses the exported, origin-undeclared `escort` with `core.origin_omitted`. This is NOT a defect introduced by this plan's change: confirmed directly (`check.Program` alone returns zero diagnostics for this fixture; `originvalidate.ValidatePublished(checked.Program)` independently returns the `core.origin_omitted` problem regardless of what corevalidate does) that this refusal has been latent since Phase 07/08, simply unreachable through the full CLI because `corevalidate`'s own (now-fixed) unconditional-propagation bug always refused FIRST in `CheckCommandFile`'s fixed precedence (check, then corevalidate, then originvalidate) — this is the exact "one validator's own defect masks another's" shape D-09-08 names for `check`, discovered here one layer further down the pipeline than any Phase 08/09 planning document anticipated. `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (the actual mechanically-enforced gate for D-09-03's retirement) is UNAFFECTED and passes cleanly: it compares `check.Program` diagnostics against `corevalidate.Validate` only, never invoking `originvalidate.ValidatePublished` at all. A real fix requires threading callee-return-contract lookups through `RecomputeOriginPerReturn`/`walkReturnOrigin`'s signatures (currently `func(function core.Function) ...`, no whole-`core.Program` access) — a cross-cutting signature change touching `BuildInterface`, `ValidatePublished`, and every existing `originvalidate_test.go` call site, judged out of bounds for this plan's declared `files_modified` and Rule 4 territory (significant structural modification), not a bounded inline fix. Landing phase and vehicle: not yet decided; needs its own scoped plan or a Phase 09 mid-phase gate (09-07/09-08) agenda item |

## Detail

### D-09-03 — the two peer-divergence fixtures, and why retiring them is an assertion

`buildLoanChainIndex` (`corevalidate.go:1133-1148`) today writes
`idx.parent[operation.TargetID] = operation.SourceID` for **every** operation
with a non-empty `TargetID`, regardless of `Kind`. Across a `core.OpCall` that
is wrong: a callee that does not return a borrow of its parameter hands back a
fresh, owned identity, and the caller's loan chain must break at the call
boundary.

The consequence is live and fixture-backed today:
`session_peer_gate_test.go:56-58` pins two accepted-program divergences where
`check` correctly admits under Phase 08's interprocedural liveness law and
`corevalidate` refuses via `core.move_while_borrowed`.

Both entries must be **deleted** by the plan that lands the consult, and
`TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` proves the deletion is honest
in both directions: a stale entry that no longer diverges fails the test just as
loudly as an undeclared new divergence. That two-directional exactness is why
this is a required assertion rather than a hoped-for side effect.

The one `phase07`-prefixed entry in that map,
`testdata/phase07/duplicate_function_name.lang` → `core.duplicate_function_id`,
is **unrelated** to this phase and must **not** be retired.

### D-09-08 — `computeLoanLastUses` is deleted, reversing D-08-41

D-08-41's recorded disposition — "an accepted, permanent, disclosed scope
limitation… Landing phase: Not scheduled" — is superseded, and the reason is the
harm D-08-41 itself documented. The Pattern B twin pair
(`twin_b_refuse.lang` / `twin_b_accept.lang`) cannot demonstrate a differing
end-to-end CLI verdict, because the summary-blind shadow path refuses **both**
members identically with `ownership.move_while_borrowed` before the
interprocedural pass ever runs. A summary-blind decision point is actively
**masking** the law the milestone exists to prove.

Making it summary-aware instead would leave two call sites into one algorithm —
exactly the coexistence OWN-09 forbids. Re-scoping "the intraprocedural law" to
mean something narrower is goalpost-moving with no basis in the code.

**What is deleted is a call site and its scaffolding, not an algorithm.** There
was never a second loan-liveness law: `computeLoanLastUses` calls the *same*
`loanLivenessFixpoint`, fed synthetic `shadow:place:*` operations built from
`ast.LinearBody` before any `core.Function` exists, and a permanent zero-value
`interproceduralSummaryTable{}`. One algorithm, two callers, one of which is
deliberately summary-blind. Any plan that reads OWN-09 as "delete a law" will
either find nothing to delete or delete the wrong thing.

Cross-ecosystem precedent: rustc's `-Zborrowck=migrate` two-law period was
explicitly temporary and ended at an edition boundary. The status-quo option
here is the documented "strangler fig that never finishes."

**Ordering is a safety constraint, not a preference** (D-09-10): the peer lands
fully first, the zero-divergence differential goes green, and only then does the
deletion land in the same commit that flips lowering to defer to the single pass.

### D-09-13 — diagnostic order and identity is the real risk of the restructure

Today a per-function early exit can report a different *first* error than a
whole-program late pass would. `Code + Span + Causes` marshal and SHA-256-fold
into a diagnostic `ID` (`internal/compiler/diagnostic/diagnostic.go:96-112`), so
emission-order changes are observable in a published, agent-facing identifier.
`LoanFinalUse` facts also move from pre-emission to post-lowering, touching
`checkLinear`/`checkBranch`'s emission contract.

The mitigation is an explicit ordering-stability assertion, authored and green
**before** the deletion lands: iterate `program.Functions` in declaration order
(verified a slice, `core/core.go:19`, so its order is deterministic) and assert
on a synthetic two-function program with a known error in each function that the
reported diagnostic matches source declaration order.

This risk has a concrete within-function form as well, recorded as D-09-49's
first open question: if the inline `move_while_borrowed` gate currently
short-circuits a binding loop, deferring it lets lowering proceed past the move
and may ALSO trip `ownership.use_after_move`, producing two errors where one
existed. The plan set settles this by enumeration over the existing corpus, not
by assumption.

### D-09-14 — OWN-09's document conflict closes in favour of the mapping table

`REQUIREMENTS.md:173`'s Phase 09 mapping stands as operative. Phase 08 is
shipped (`4ab5c65`) and reopening it to relitigate OWN-09 is not viable.
OWN-09's own requirement text — "The intraprocedural loan-liveness law is
retired in the same phase the interprocedural law lands" — is what gets
corrected, exactly as Phase 08's mid-phase gate anticipated ("Phase 09's own
planning is where OWN-09's text should be corrected to match"). This closes
D-08-27, which the Phase 08 gate reviewed and deliberately carried unresolved.

### D-09-21 — the `OpForeignCall` fallback and the trigger that fires it

D-07-33's own debt row already records "Landing phase: Phase 09" and "in the
same plan that builds the peer's own origin recomputation." It was committed in
writing at Phase 07 planning time per Key Lesson 4; it was never proposed for
cutting, which is why REQUIREMENTS.md's cut order does not list it. It needs
**execution**, or a **new** cut decision with its own justification.

The literal OWN-08 reading — "no declared origin" maps exactly to
`core.origin_omitted`, the one class the peer already covers, therefore OWN-08
closes without touching the other three — is **rejected**. `Callable` is defined
project-wide as the full D-04-03 predicate (`PublishProblemsFor` returns no
problems across all four classes). Declaring OWN-08 closed while three of four
refusal classes stay provably vacuous on the peer side is Key Lesson 3's failure
mode verbatim, and it would be D-03-02 repeating its own M001 history — deferred
under a closure claim, for a second consecutive milestone, on the same item.

Plan-time analysis says the work is bounded: `checkForeignOriginOmitted`
(`originvalidate.go:282-345`) is a ~55-line backward walk scoped to one
function's own `ForeignContract.Alias` — proof by existence that a narrow,
function-local foreign-crossing check is possible without spike 005's completion.
The peer mirrors that **narrowness in spirit, never in code**.

The fallback above fires only on an execution-time discovery, and it is written
here now so that discovering it under load cannot be mistaken for deciding it.

### D-09-29 — the accepted-program disclosure is resolved as "no vehicle, by design"

Verified in tree: the accepted-side consulted field set is a **compile-time
constant**. The consultation loop (`check.go:594-609`) fires exactly
`return.mode` and `parameters[0].mode` for every function with a declared
signature, corpus-wide, recorded through the `interproceduralConsultObserved`
seam (`check.go:641`) and already proven a closed set by the shipped
`TestInterproceduralDisclosedFieldSet` (`check_test.go:4898`).

A runtime record would therefore reprint a constant on every run at real
`qlt02`-gated cost, with zero incremental per-program information — the opposite
of the refusal-side disclosure, which genuinely varies (direction, which callee).
This is a structural property, not a budget excuse, which is why the framing
changes from "declared debt" to "resolved."

Phase 09's positive obligation in its place is D-09-30: `corevalidate` gets its
**own independently-written** corpus-wide closed-field-set test, and a cross-peer
test asserts the two sets are **identical**. Matching consulted *field sets* is a
contract-conformance check ("both peers read only what SEM-05's body-blind
`ParameterContract.Mode` / `ReturnContract.Mode` legally permit"); it says
nothing about worklist-vs-closure, traversal order, or internal data structures.
Far from weakening independence, it is strictly stronger than verdict-agreement
alone: two independent derivations can agree on accept/reject while one secretly
reads a third field, and verdict comparison never catches that.

### D-09-31 — the promotion commitment is superseded; the codes stay divergent

D-08-17 argued the peer "computes liveness through a reachability closure, not a
worklist, so it has a differently-shaped ceiling and **cannot fail the same
way**. A shared `core.*` code would assert an agreement the two mechanisms are
structurally incapable of having." That reasoning is general to the two
mechanisms, not scoped to the iteration bound.

The corpus already demonstrates it: `corevalidate` refuses the relay-escort
witness via `core.move_while_borrowed` (`corevalidate.go:1551`) while `check`
refuses the identical program via `check.interprocedural_loan_liveness`
(`check.go:963-966`) — and `session_peer_gate_test.go` **celebrates** this as
"BOTH sides refuse independently, via different codes, for the same program."

Promoting would require either forcing `corevalidate` to synthesize `check`'s
fixed three-role cause template
(`borrow_created_here` / `loan_extended_by_call` / `callee_return_contract`), for
which a reachability closure has no natural "the call that extended the loan" —
reverse-engineering the worklist's shape into the peer, the literal anti-pattern
this phase exists to prevent — or accepting the same code with different causes,
which (per the identity fold) produces **different diagnostic IDs for the same
code depending on which layer refused**: a strictly worse dispatch contract for
`lang-repair` than two distinct stable codes. `Repairs: nil` stays on `check`'s
side, independently justified by D-08-25, and is never imposed on the peer.

Divergent-codes-per-peer is already shipped, not novel:
`check.call_argument_type_mismatch` / `core.CallArgumentTypeMismatch` (D-07-46)
are a deliberate different-code pair for the same defect.
`core.call_graph_cycle` is the *shared*-code precedent and is justified by
something absent here: a shared, peer-agnostic witness type
(`callgraph.CycleError`) available to both layers from day one. Cross-mechanism
sameness is expressed as a documented relationship, never as a merged identifier.

**Reopen only if** the two mechanisms are ever proven to refuse for the identical
underlying fact with an identical natural cause shape.

### D-09-37 — OWN-05 completes partially, and the split is recorded now

OWN-05's text names three derivers; Phase 09 delivers two. This is the same class
of conflict as D-08-27, and it gets the same treatment **now rather than later**.

Disposition chosen at planning time: **split** into

- **OWN-05a** (Phase 09): ownership transfer at a call site has one meaning,
  derived independently by `check` and `corevalidate`; call-site override of a
  callee's declared convention is not expressible in source and is fail-closed at
  the core layer.
- **OWN-05b** (Phase 10): the same fact derived independently by `interp`,
  verified there by TRU-03 and in Phase 11 by NAT-06.

A split is preferred over a single "partial" row because REQUIREMENTS.md's
traceability table asserts every requirement maps to exactly one phase, and a
row that maps to two phases quietly breaks the coverage arithmetic the table
publishes.

Phase 09 must **not** prepare for Phase 10 by extracting a shared "convention
classification" helper (D-09-38). `check` and `corevalidate` already each read
`signature.Parameters[0].Mode` independently, sharing nothing beyond the struct
shape. A helper "ready for `interp`" is precisely ARCHITECTURE §4's "single point
of failure wearing two names."

### D-09-40a — QLT-07's pre-registered threshold, and the COMMITTED verdict

The threshold was fixed **before** the count was taken, which is what keeps this
gate non-decorative and distinguishes it from D-03-02's defer-under-load pattern:

> QLT-07 is committed (in scope) if and only if the inventory requires ≤ 1
> additional plan beyond what Phase 09 already needs for
> OWN-05/07/08/09/TRU-04, AND opens zero packages or files that OWN-07 /
> OWN-08's work in `check` / `corevalidate` does not already touch.

The inventory (run 2026-09-10, before any `09-PLAN.md` was drafted) enumerated
only the loan-liveness-scoped rows of `03-VALIDATION.md` — the 03-03 / 03-04 /
03-05 rows, explicitly excluding 03-06 / 03-07's OWN-04 origin rows, which
QLT-07's text does not claim. Finding: **all 33 tests named across those 12 rows
already exist in the tree and pass**
(`go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle`
— all three packages `ok`).

Both bounds are satisfied with margin: ≤ 1 additional plan (one
`09-VALIDATION.md` section plus one pointer line), and zero production files
touched. **QLT-07 is committed.** No stretch re-declaration is written and the
scope-cut order's item 2 is not triggered.

`03-VALIDATION.md` is never amended in place (D-09-41). Its
`nyquist_compliant: false`, its unticked checkboxes, and its recorded acyclic-CFG
scope limitation stay exactly as written — its own text defers loop-carried loan
liveness to "Phase 4-or-later" and the language still has no loops, so that
clause is structurally un-closable today and the document's boolean can never
legitimately flip to `true`. The only permitted edit is a single non-mutating
pointer line.

### D-09-43 — the scope-cut trigger, declared before the phase begins

The milestone declares a 2x trigger for Phases 08 and 09. Adopted for Phase 09
with an explicit, ordered cut list and an explicit never-cut list, both above.

Any cut is a **deferral with a named landing phase**, never a silent drop, and
the decision is adjudicated at plan 09-07's mid-phase gate where the elapsed
cost against the initial estimate is actually visible.

### D-09-45 — four stale references, corrected rather than left to mislead

Recorded per this project's own stale-reference discipline: a planning document
that is quietly wrong costs a future reader a re-derivation they cannot know they
need.

(a) and (b) are corrections to `08-CONTEXT.md` (a line number and a call-graph
order direction). (c) is a correction to an earlier revision of D-09-03's own
fixture paths — the shipped map keys are under `testdata/phase08/`, and a `grep`
for `testdata/phase07/twin_a_accept.lang` returning nothing is the warning sign.

(d) is `TestQLT01RegistryCoversAllFiveSpikes` failing today with "spike 006 has a
directory under .planning/spikes but no registry row cites it" — a pre-existing,
un-owned registry-maintenance gap, the one known pre-existing test failure in the
tree. Phase 09 touches the spike table anyway under D-09-39 (S-008's replacement
must be recorded as an explicit process amendment, not a silent substitution), so
adding spike 006's row is cheap and clears the failure. Recommended, not required.

### D-09-46 — the corpus generator must be made reachable before TRU-04's vehicle works

`generateCallGraphCorpus` (`internal/compiler/check/costcorpus_test.go:50`) is
unexported and lives in a `_test.go` file inside package `check`. Unexported
symbols in `_test.go` files are invisible outside their own package, full stop —
external-test-package status does not change this.

Resolution chosen at planning time: **relocate**, do not duplicate. Duplicating
it would reintroduce the exact two-derivations-one-truth risk this phase exists
to eliminate, this time for the test fixtures. The destination is
`internal/compiler/testsupport`, whose current imports are stdlib only, so it
cannot become a route by which `session` or `corevalidate` gains a production
import of `check`. `session_peer_gate_test.go` already imports `testsupport`
(`session_peer_gate_test.go:12`), so the consumer side costs nothing.

`TestCostCorpusIsNotParsed`'s structural scan must follow the generator to its
new file, or the "never generated as `.lang` source through the real parser"
guarantee (D-08-34) silently stops being checked.

### D-09-50 — the synthetic differential's vehicle splits from the `.lang` corpus gate

D-09-23 prescribes feeding Phase 08's synthetic call-graph shapes into
`session_peer_gate_test.go`'s existing corpus-wide gate, and forbids standing up
a second, independently-truthed harness.

Verified at planning time that the literal prescription is not achievable:
package `check` exports exactly one admission entry point,
`check.Program(program ast.Program) Result` (`check.go:46`). There is no exported
route by which any other package can compute a `check`-side verdict for a
synthetically constructed `core.Program` — and the synthetic corpus is
deliberately built directly against `core`'s types, never parsed (D-08-34), so
there is no `ast.Program` to feed.

Disposition: the `.lang` corpus gate stays exactly where it is and is where
D-09-03's two retirements are proven. The **synthetic-shape** differential lives
inside package `check`, which alone can reach both the unexported post-assembly
pass and `corevalidate.Validate` (`check_test.go:20` already imports it).

This is a split of **vehicle**, never of **truth**, and the distinction is
load-bearing: the synthetic differential must apply the same both-directions
exactness discipline (an undeclared divergence fails; a stale declared entry that
no longer diverges also fails), must cross-reference `peerDivergenceExpected`'s
own doc comment, and must never define "divergence" a second way. If a future
reader finds two different definitions of divergence in this tree, this row is
where the intent was recorded and the intent was violated.

### D-09-51 — `originvalidate`'s own OpCall-transparent origin walk, unmasked by D-09-03's fix

`originvalidate.walkReturnOrigin` (`originvalidate.go:171-218`) walks backward
from a function's own `OpReturn` through a `sourceOf` map built from every
operation's `TargetID`, regardless of kind. Its `switch operation.Kind` names
`OpBorrowExclusive`, `OpBorrowShared`, and `OpForeignCall` explicitly, but has
no `case core.OpCall` at all — an `OpCall` simply falls through to
`current = operation.SourceID`, exactly like an ordinary pass-through hop,
never consulting the CALLEE's own declared return contract.

Verified directly (`internal/compiler/session`, ad hoc probe against
`testdata/phase08/twin_a_accept.lang`, removed after verification):
`check.Program` returns zero diagnostics; `originvalidate.ValidatePublished`
called independently on that same checked program returns exactly one
problem, `core.origin_omitted`, for `escort`. The same result reproduces for
`testdata/phase08/relay_depth2_accept.lang`'s `caller`. Both are the exact two
fixtures D-09-03 retires from `peerDivergenceExpected` — this plan's fix is
correct and complete for corevalidate's own loan-liveness re-derivation; the
CLI-level refusal that remains afterward is a wholly different validator's
wholly different, pre-existing bug, simply never observable before because
corevalidate's own defect refused first.

Confirmed NOT a regression introduced by this plan: `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus`
(`session_peer_gate_test.go`) — the actual mechanically-enforced gate behind
D-09-03's retirement — computes its verdict from `check.Program` and
`corevalidate.Validate` only. It never calls `originvalidate.ValidatePublished`,
so it is blind to this finding by construction and passes cleanly once the two
entries are removed, exactly as D-09-03 requires.

The plan's own literal `<verify>`/`<acceptance_criteria>` text additionally
asserts `go run ./cmd/lang --json check testdata/phase08/twin_a_accept.lang`
(the FULL CLI, which does call `originvalidate.ValidatePublished` after
`corevalidate.Validate` passes, per `session.CheckCommandFile`'s fixed
precedence) reports a clean check. That literal claim is NOT satisfied after
this plan: the full CLI now reports `core.origin_omitted` instead of
`core.move_while_borrowed` for both fixtures — a different validator, a
different (pre-existing) root cause, and objectively closer to correct (the
corevalidate-side divergence this phase exists to close is genuinely gone),
but not a clean check. Recorded here rather than silently declared satisfied.

Not fixed in this plan: a real fix requires `RecomputeOriginPerReturn`/
`walkReturnOrigin`/`RecomputeOrigin` to gain access to the whole
`core.Program` (or a precomputed callee-return-mode table), so an `OpCall`
hop can consult its callee's own declared origin/return contract the same
way `check`'s `derivePlaceLoans` and this plan's own `derivePeerLoanCarry`
already do — a signature change to originvalidate's own exported API,
touching `BuildInterface`, `ValidatePublished`, and every existing
`originvalidate_test.go` call site that constructs these calls directly
against a bare `core.Function`. `internal/compiler/originvalidate/originvalidate.go`
is not in this plan's `files_modified`, and the blast radius (an exported-API
signature change across a validator with its own extensive same-package test
suite) is judged Rule 4 territory (significant structural modification), not
a bounded inline fix available to an executor mid-task.

---

*Register written at Phase 09 planning time, 2026-09-10.*
*D-09-51 appended during plan 09-01 execution, 2026-09-10 (execution-time
finding, not a planning-time item).*
*Shape validated by `TestDebtRegistersAreWellFormed`
(`internal/compiler/session/session_test.go`).*
