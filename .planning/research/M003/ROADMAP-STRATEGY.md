# M003+ Research: Roadmap Strategy and Milestone Gameplan

**Researched:** 2026-09-17
**Mode:** Feasibility + sequencing (multi-milestone)
**Overall confidence:** HIGH on M003 content and ordering; MEDIUM on the M004–M006 arc; LOW–MEDIUM on the AI-author premise (§ Adversarial pass).

---

## Verdict

**M003 should be "Computation" — the milestone in which a Lang program produces a value that was not already in its input.** Concretely: retire the single-function guard inventory and the event-identity chain first (pure subtraction, forced by the project's own D-10-60 no-third-deferral rule); then lift `sameType(ReturnType, Parameter.Type)` so a function's return type may differ from its parameter type (this closes debt Cluster C — D-13-02b, D-13-10a, and Phase 13's twin-pair sub-finding — automatically, and makes the already-built `resolveBlame` reachable); then land value literals and arithmetic/comparison operators as new `core.OperationKind`s at all six dispatch sites; then prove those new semantics survive Clang at `-O0`/`-O3`/`-flto` at the one place the native path is genuinely likely to disagree with the interpreter oracle — **integer overflow, division by zero, and C's integer-promotion rules**; then close the Nyquist debt for 07/08/11/12/13 as a *gate condition*, repair D-13-34, and install a standing "frontier fixture." Five phases, ~40–45 plans, explicitly smaller than M002. **Arity-N, iteration, `if`, strings, arrays, generics, modules, and effects are all out.**

The reason this is the highest-leverage cut: arithmetic is the first language feature whose *semantics are genuinely contestable between the interpreter and C*. Every M001/M002 equivalence claim was about plumbing — move a value, call a function, destructure a payload. `uint8_t a * b` promotes to `int` in C17 and signed overflow is undefined behaviour; that is the first real subject the five-axis comparator has ever had that was not hand-engineered as a negative control. M003 is where the assurance stack stops being ahead of the language and starts being *used by* it.

### The M003–M006 arc

| # | Charter | Thesis (one line) | The gate that becomes meaningful | Proves what the prior could not | Phases |
|---|---------|-------------------|----------------------------------|---------------------------------|--------|
| **M003** | **Computation — Values, Operators, and Contracts** | A program creates a value instead of relocating one, and a function's contract can finally say something. | Interpreter and all three native tiers agree on arithmetic *at the overflow / division / promotion boundary*, with an engineered divergence proving the tier non-inert — and a B1 contract-violation diagnostic is constructible. | M002 could not construct a program whose output differed from its input, so no comparator axis ever adjudicated a computed value (see D-12-43). | 5 |
| **M004** | **Iteration — Bounded Repetition and Back Edges** | The CFG grows a back edge, and every analysis that assumed acyclicity is either widened or replaced. | A loop's loan liveness is derived across a back edge by two structurally opposite peers that agree, and a non-terminating program is refused *by name*, never hung. | M002 and M003 both run on acyclic CFGs; `pathoracle.cfg_back_edge` (`pathoracle.go:243`) refuses every cycle by construction. Iteration is the first genuine architectural one-way door. | 5–6 |
| **M005** | **Aggregates and Arity — Indexed Data** | Functions take more than one argument and values have addressable interiors. | An element-level wrong-slot write is *observable* by a comparator axis — i.e. **D-12-43 finally becomes constructible**, not ratified-unconstructible. | Neither M003 nor M004 gives the harness a payload-value-aware observation channel; the D-12-43 finding names exactly this as its unblocking trigger. | 6 |
| **M006** | **Modules and Separate Compilation** | Two units compile apart, link, and carry reusable evidence; the language gets a place to put a stdlib. | Editing a callee invalidates exactly the right cached interprocedural facts — no stale verdict, no over-invalidation — across two separate `lang` invocations. | M001 demonstrated three separate CLI invocations for origins only; the content-bound cache has never been tested against a real cross-unit callee edit (STANDING-VERDICTS names this as the Rust-ThinLTO / Bazel precedent risk). | 6 |

**The first real program lands at the end of M005** (§ The first real program). It is nameable today and should be checked in *now* as a refused fixture.

---

## The central tension: 70% assurance vs 10% surface — strength or trap?

This is the most important question in this document, so both cases get their strongest form before the position.

### The case that it is a strength

**1. The historical direction of travel runs the other way, and it is expensive.** Rust's original AST-based borrow checker determined lifetimes from lexical scope; it had to be *redone* against MIR as non-lexical lifetimes because scope-based analysis was not flexible enough for the language that had grown around it ([rustc_borrowck docs](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_borrowck/index.html), [PR 45538](https://github.com/rust-lang/rust/pull/45538)). That was a multi-year rewrite of the central safety analysis *after* the surface was fixed. C++ never retrofitted ownership at all. Adding `+` to a sound affine core is a weekend by comparison with retrofitting affinity onto a language that already has `+`.

**2. The assurance stack here is architectural, not feature-specific.** The load-bearing assets are: independent re-derivation at trust crossings by *structurally opposite* algorithms (`check` backward-memoized vs `corevalidate` forward set-propagation), a deterministic interpreter as the semantic oracle, differential execution across four optimization tiers with engineered negative controls, and mutation-killing every control. None of those are properties of arity-1 `Byte → Byte` functions. They are properties of the *build discipline*, and they transfer to any surface.

**3. Verification-first language projects are the precedent, and they grew surface slowly on purpose.** CompCert's verified core is Clight, a deliberately restricted C subset, and the full C surface accreted over many years on top of a 20-pass, 11-intermediate-language pipeline ([Clight semantics, Leroy](https://xavierleroy.org/publi/Clight.pdf); [CompCert structure](https://www.absint.com/compcert/structure.htm)). CakeML compiles high-level features away incrementally through 12 intermediate languages precisely so each pass can be verified at the right level of detail ([CakeML](https://cakeml.org/)). Neither started expressive. Both ended up with real surface.

**4. Deferral has a good track record *in this project*.** Deferring `OpCall` out of M001 was ratified as correct: M002 then spent 7 phases and 61 plans on exactly that one addition. The judgment that a large structural change deserves its own milestone has been tested once and held.

### The case that it is a trap

**1. The ratio is an artifact of a moving denominator.** "~70% of the assurance stack" means 70% of the assurance needed for a language with one type per function, exactly one parameter (`core.Function.Parameter` is singular; `check.go:3432` fixes arity at 1 and `check.go:3449` refuses anything else as `check.call_arity_unsupported`), and **no numeric literals at all** (`grep -rn 'IntLiteral\|NumberLiteral\|ByteLiteral' internal/compiler/syntax/` returns nothing — you cannot write `1`). Against the language that must eventually exist, the honest number is closer to 30%. The denominator grows faster than the numerator, and the project's own standard says each new `OperationKind` costs a **minimum of 8–10 independent edits** across six dispatch sites plus two exhaustive-dispatch controls (STANDING-VERDICTS, "load-bearing technical facts"). If assurance cost is superlinear in surface, "hard part first" is backwards.

**2. The trap has already fired three times, in M002, and it is written down.** Not hypothetically — three of M002's own headline deliverables were blocked by the surface being too small:
- **D-12-43**: the decisive wrong-slot value-divergence control was found *unconstructible against the current representation* and escalated as a criterion defect. The comparator cannot see a value corruption because the grammar never lets a value be seen.
- **D-13-02b**: `resolveBlame` is built, tested, compile-time exhaustive — and has zero call sites, because no B1-shaped diagnostic is constructible while `sameType(ReturnType, Parameter.Type)` holds (`check.go:255`, `:3148`, `:3399`).
- **D-13-10a**: the third repair class was withdrawn empirically because, with one type per function, the "correct" argument is always the one already passed.

That is the assurance machinery producing three unfalsifiable claims in a single milestone. A smart outside reviewer would call that the trap firing at small scale.

**3. Specific mechanisms are known not to survive.** `pathoracle` does bounded exhaustive path enumeration and refuses any CFG cycle by name (`pathoracle.cfg_back_edge`, `pathoracle.go:243`). Iteration does not make that expensive — it makes it **impossible**. One of the three independent peers that gives the project its authority has to be redesigned or replaced the moment loops land. Similarly, 32 `len(Functions) != 1` guards across 6 files in 3 packages had to be individually re-adjudicated in Phase 11 (29 WIDENED/KEPT dispositions) just to run two functions.

**4. Cost evidence is not improving.** M001: 62 plans / 5 days. M002: 61 plans / 6 days — for exactly one new operation kind. The retrospective states this plainly: "per-plan cost did not fall." A curve that is flat while the feature delta shrinks is a curve that will be steep when the feature delta grows.

### Position

**It is a strength — at roughly 70/30 — but it is one milestone away from becoming a trap, and the trap has a measurable tell.**

The reason I land on strength is a distinction the debt register makes visible and that the critique elides: **in all three M002 failures the machinery was correct and the surface was missing.** Nothing in `loanLivenessFixpoint`, the five-axis comparator, the HDD reducer, or the peer architecture has been shown *wrong*. Three things have been shown *unexercised*. Those are very different failure modes. "Built right, not yet loadable" is repaired by adding surface; "built against the wrong model" is repaired by rewriting. The project is in the first state, and the fix for the first state is precisely what M003 does.

But the bet is falsifiable, and it should be stated as a trigger rather than a belief:

> **Trap-confirmation trigger.** If M003's arithmetic phases require a *redesign* — not an extension — of `loanLivenessFixpoint`, the five-axis comparator, or `corevalidate`'s forward propagation, then the assurance stack was calibrated against a toy and M004 must be re-chartered as an assurance-refactor milestone before any further surface is added.

And one already-known one-way door should be named now rather than discovered: **`pathoracle`'s back-edge refusal does not survive M004.** Budget a pathoracle successor (bounded unrolling to a declared depth, or a loop-summary oracle) as a pre-M004 spike, not as a mid-phase surprise. That is exactly the shape of S-006, which the project ran correctly for interprocedural liveness cost.

Finally, an instrument, because the project already owns one (`internal/compiler/measure`, growth-exponent fits, the QLT-02 budget manifest): **track edits-per-new-`OperationKind` and plans-per-surface-feature across M003–M005.** If that number rises materially, the superlinearity critique is winning and the strategy should change. Today it is 8–10 edits/kind by the project's own count. Publish it.

---

## M003 charter

**Name:** **M003 — Computation: Values, Operators, and Contracts**

**Thesis:** A Lang program produces a value that was not present in its input, and a function's contract can finally say something a caller could violate.

**The gate that becomes meaningful:** *Can a computed value be adjudicated?* Every gate before this one asked "can this claim be independently checked?" about a fact (a loan endpoint, a call-graph edge, a drop obligation). M003's gate asks it about a **value**: interpreter, `-O0`, `-O3`, and `-O3 -flto` must agree on the result of arithmetic including at overflow, division by zero, and the C integer-promotion boundary — with an engineered divergence (a naive `uint8_t` emission that diverges only at `-O3`) proving the tier is not inert. This is the same shape as `control:alias.false_no_alias` from M001 Phase 5, and it is the first time the comparator has a subject that arises from the *language* rather than from a hand-built hazard.

**What M003 proves that M002 could not:**

> **M001 proved one meaning survives lowering. M002 proved it survives a function boundary. M003 proves it survives *computation* — the first value Lang creates rather than moves — at exactly the boundary where "safe code has defined behavior" stops being free.**

M002 structurally could not make this claim: with no literals, one type per function, and arity 1, every admitted program's output was a relocation or a re-tagging of its input. D-12-43 is the formal record of that limit — a control the harness could not construct because no value was ever observable.

**Rough phase cut (5 phases, ~40–45 plans, target 4–5 days):**

| Phase | Name | Gate |
|-------|------|------|
| **14** | **Retire the Single-Function Inventory and Own Event Identity** | A non-test scan for `len(Functions) != 1` returns zero hits in `cgen`, `session`, and `reduce`; shared-leaf diamond call graphs produce distinct event identities (D-11-51 → D-12-21 closed). |
| **15** | **Signature Widening: Return Type ≠ Parameter Type** | A B1-shaped interprocedural contract-violation diagnostic is constructible; `resolveBlame` has real call sites; `use_matching_argument` produces a repair that is not byte-identical to its input. DX-06 and DX-07 flip from Partial to Complete. |
| **16** | **Literals and Arithmetic Operation Kinds** | The new `OperationKind`s are *recognized, not faked* at all six dispatch sites, with both exhaustive-dispatch controls green and a seeded mutation observed to fail each. (The Phase 07 pattern, exactly.) |
| **17** | **Defined Arithmetic Under Optimization** | Five-axis agreement across interpreter/`-O0`/`-O3`/`-flto` on an overflow/division/promotion corpus, with an engineered `-O3`-only divergence proving non-inertness, and a checked-in refusal or defined wrap for every UB-reachable operation. |
| **18** | **Validation Debt, Evidence Repair, and the Frontier Fixture** | `/gsd-validate-phase` reconciled for 07/08/11/12/13 **as a gate condition**; D-13-34's alpha-renamed held-out pairs structurally distinct under the same predicate that found the hole; the frontier fixture harness green. |

**Explicit non-goals for M003 (write these down at the moment of deciding, per M001 Lesson 4):**

- **Arity-N.** Deferred to M005. The `lang.interface/1` schema already carries `Parameters []ParameterContract` (D-07-10 took the capacity early), so the deferral is cheap to reverse.
- **Iteration of any kind.** Deferred to M004. It breaks `pathoracle` structurally.
- **`if`/`else`.** Deferred *permanently* — see § Dependency reasoning. `match` over a two-alternative `data` type is `if`, and a second control-flow law violates the project's own Key Lesson 2.
- **Strings, arrays, collections, indexing.** M005.
- **Recursion, generics, modules, separate compilation, effects, async.** Unchanged from PROJECT.md Out of Scope.
- **A second comparator.** STANDING-VERDICTS: extend the corpus and the controls, do not duplicate the infrastructure.

---

## M004 / M005 / M006 charters

### M004 — Iteration: Bounded Repetition and Back Edges (5–6 phases)

**Thesis:** The CFG grows its first back edge, and every analysis that silently assumed acyclicity is widened, replaced, or explicitly retired.

**Gate:** A loop's loan liveness is derived across a back edge by two structurally opposite peers that agree (with bidirectional seeded faults), and a non-terminating program is refused *by name* under a declared bound, never hung — the same fail-closed discipline `callgraph`'s cycle refusal and `loanLivenessFixpoint`'s derived iteration bound already use.

**Proves what M004's predecessors could not:** every program admitted through M003 runs on an acyclic CFG. `pathoracle` refuses cycles by construction. M004 is the first time the trust architecture faces a fixpoint it cannot exhaustively enumerate.

**Out:** iterators/generators, `for ... in` over collections (no collections yet), recursion (still refused), unbounded loops with no declared budget.

**Mandatory pre-phase spike (hard entry gate on Phase 1 planning, the S-006 pattern):** *What replaces `pathoracle`?* Bounded unrolling to a declared depth, a loop-summary oracle, or retirement of pathoracle as a peer with a named replacement. Answer this before planning, not during.

### M005 — Aggregates and Arity: Indexed Data (6 phases)

**Thesis:** Functions take more than one argument, and values have addressable interiors that the harness can observe.

**Gate:** **D-12-43 becomes constructible.** An element-level wrong-slot write produces a real, observed divergence on a comparator axis — closing, on evidence, the one M002 criterion that was ratified as unconstructible. Secondary gate: an owned aggregate passed alongside a borrow of one of its elements is adjudicated identically by both admission peers.

**Proves what M004 could not:** arity-N is the second real load test of interprocedural loan liveness (multiple simultaneous parameter loans, per-parameter public origins), and interior addressability is the first time the evidence layer can see *into* a value.

**Out:** partial moves (PROJECT.md Out of Scope — an element cannot be moved out of an aggregate), dynamic allocation beyond the existing explicit-resource model, generic collections, strings as a distinct type (Buffer-of-Byte with an encoding convention is enough).

### M006 — Modules and Separate Compilation (6 phases)

**Thesis:** Two units compile apart, link, and carry reusable evidence; the language gets a home for a standard library.

**Gate:** Editing a callee invalidates exactly the right cached interprocedural facts across two separate `lang` invocations — no stale verdict served, and no over-invalidation that would destroy the feedback budget. Measured against the ratified p50/p95/CoV protocol, not asserted.

**Proves what M005 could not:** M001 demonstrated separate compilation for published origins only, as three CLI invocations. The content-bound cache has never been exercised against a real cross-unit callee edit; STANDING-VERDICTS names Rust's ThinLTO import-map invalidation and Bazel/Buck cross-module bugs as the live precedents.

**Out:** a package registry, a remote cache, an LSP, version resolution. STANDING-VERDICTS explicitly warns against bundling modules into a milestone already extending five-plus-consumer machinery — which is why this gets its own milestone and comes *after* the surface stabilizes, not before.

---

## Dependency reasoning: why this order

Taking the six candidate M003 ingredients in turn, by what each blocks or unblocks rather than by preference.

**(c) Retire the six single-function emitters (D-11-02 → D-12-36) — M003, Phase 14, FIRST.**
Two independent reasons it goes first, neither of them "because the rule says so."
1. *It is pure subtraction and it is ready now.* Phase 11 already shipped the multi-function coverage the deletion needs and recorded a WIDENED/KEPT disposition for all 29 re-verified guards. Nothing in M003 blocks it.
2. *Ordering avoids doing the work twice.* Every later M003 phase adds fixtures and emission paths. Adding a new `OperationKind` to `cgen` and `session` while 32 single-function guards are still in place means writing each new path against the guard and then rewriting it after deletion. Subtraction before addition.
The D-10-60 rule is a third reason, and it is decisive if the first two were close: a third deferral would break a rule the project wrote specifically to prevent this exact pattern, and M002's Key Lesson 2 already records that *writing the rule did not prevent the behaviour*. Either land it or retire D-10-60 in writing.

**(d) Own event identity (D-11-51 → D-12-21) — M003, Phase 14, bundled with (c).**
Same files, same layer. D-11-51 (shared-leaf diamond call graphs collide on event identity) and D-12-21 (blocked on D-11-51) are both `session`/`interp`/`cgen` event-plumbing defects against NAT-06. Bundling is correct here for the reason M002's Phase 11 bundling was correct: the phase gate is about the *event stream being a trustworthy identity space*, and both items are about that one claim. Separating them would produce two phases whose gates are the same sentence. Additional urgency: arithmetic makes event collisions *more* observable (more distinct computed values flowing through the same leaf), so fixing it before M003's later phases prevents a class of false comparator divergence.

**(a) Widen the type system so return type ≠ parameter type — M003, Phase 15, SECOND.**
This has the best leverage-to-cost ratio in the entire candidate list and it has a hard ordering constraint nobody has stated yet:

- *It closes three debt items for free.* D-13-02b, D-13-10a, and Phase 13's twin-pair sub-finding are one root cause (the audit says so explicitly: "these are not independent debts — they close together, automatically"). The blame resolver and its exhaustiveness guard are already built and waiting. This is the cheapest debt retirement available.
- *It is a prerequisite for comparison operators to be useful at a signature.* `<`, `==` produce a Bool-shaped value from Byte operands. A *local* comparison inside a `Byte → Byte` function works without the widening; a function `fn is_small(v: Byte) -> Ordering` does not. Landing comparison operators before the widening creates a language in which a predicate cannot be a function — a half-state that would need re-proving later at six dispatch sites.
- *Therefore it must precede the operator phases, not follow them.* Fixtures written after the widening are written in the final language; fixtures written before it are written twice.

**(b) Control flow + arithmetic — SPLIT THREE WAYS.**
This candidate is not one item and treating it as one is the single biggest scoping error available in M003.
- **Literals + arithmetic + comparison operators → M003 Phases 16–17.** Semantically local (Byte is copy, no ownership interaction), but semantically *contestable* against C — which is what makes them the right M003 subject. Two phases, not one, for the same reason M002 separated Phase 07 (recognition) from Phase 11 (execution/equivalence): a "recognized at six sites" gate must not be allowed to imply an "agrees under `-O3`" answer it never tested. That separation is what caught Phase 07's requirement-vs-traceability defect.
- **Iteration → M004.** Hard dependency: `pathoracle.cfg_back_edge` refuses every CFG cycle. This is not a cost issue, it is a structural refusal in one of the three independent peers that give the project its authority. Iteration cannot be bundled into a milestone that is also landing two new operation kinds and a type-system widening — that is the exact fingerprint (a new kind at six sites *alongside* a structural change to a peer) that the M001→M002 `OpCall` deferral was ratified for.
- **`if`/`else` → never.** Recommend an explicit anti-feature entry in STANDING-VERDICTS. `match` already provides branching with a real CFG (M001 Phase 3 gave arm bodies entry/arm/join blocks and real edges), and a two-alternative `data` type is a boolean. Adding `if` creates a second control-flow law that must be independently re-derived at six dispatch sites and kept in agreement with `match` forever — precisely the "two coexisting laws is a defect with a delayed fuse" failure M001 Key Lesson 2 records. It buys zero expressiveness. Reject it, in writing, now.

**(e) Close the Nyquist validation debt for 07/08/11/12/13 — M003, Phase 18, LAST, and as a GATE.**
Two project rules collide here and the collision has to be resolved explicitly rather than by picking one.
- STANDING-VERDICTS process anti-pattern: *"Don't Nyquist-validate a surface a milestone is about to change."* M003 changes exactly that surface — `cgen` and `session` guards (Phase 14), `check` admission (Phase 15), all six dispatch sites (Phase 16).
- M001 Lesson 5 and M002 Lesson 5, confirmed twice: *"validation that is not gated slips"* — compliant for exactly the phases where it was a gate condition (09, 10), not for the other five.
**Resolution:** validate at the *end* of M003, not the start, and make it a phase gate rather than a cleanup task. Rows M003 touched get validated against the *new* surface; rows M003 did not touch get validated as-is. Any row that M004 will change gets carried as *disclosed* debt with a named landing phase — not as an unowned item. Doing this at the start of M003 would validate five phases' worth of surface and then invalidate most of it within the same milestone.

**(f) Repair M001's alpha-renamed held-out corpus (D-13-34) — M003, Phase 18, one plan.**
Cheap, backward-pointing, and it weakens a claim that is *already shipped*. It cannot be deferred on the grounds that it blocks nothing, because "blocks nothing" is exactly why it will be deferred forever. One plan, in the phase that is already doing evidence hygiene, using the same structural-distinctness predicate that found the hole. Its real value is regression-shaped: the new predicate becomes a standing control over every future held-out split, including M003's own.

**What this ordering deliberately does not do:** it does not parallelize. Phases 14→15→16→17 are a strict chain and each one's gate would be meaningless before its predecessor closed. That is the M002 phase-cutting rule applied honestly ("cut where a gate becomes meaningful, not where implementation could parallelize"), and it is why the phase count is 5 rather than 7 — there is no fifth and sixth independent gate hiding in here.

---

## Scope discipline: right size for M003, and the overscoping risk

**Recommended size: 5 phases, 40–45 plans, 4–5 days.** Explicitly *smaller* than M002 (7 / 61 / 6) and smaller than M001 (6 / 62 / 5).

**Why smaller, when M003 adds more surface than M002 did?** Because the observed cost driver is not feature size, it is coordination breadth. M002 added exactly one `OperationKind` and cost 61 plans; the retrospective's own conclusion is that "the peer architecture ... is now the steady-state cost: every semantic change lands at six dispatch sites plus two exhaustive-dispatch controls." M003 adds two operation kinds (literals, binary operators) plus one type-system widening plus one deletion. On the 8–10-edits-per-kind standard that is roughly 20–25 coordinated edits of new work plus a subtraction pass plus an admission change. 40–45 plans is generous for that. **If M003's plan count is heading past 50, something has been smuggled in** — most likely arity-N or iteration, the two things this document says to keep out.

**The concrete overscoping risk, given the `tech_debt` closeout:**

M002 closed with **10 open, unowned debt items out of 77**, which is *why* the audit status is `tech_debt` rather than `passed`. The risk is not that M003 inherits ten items — M003's Phases 14, 15, and 18 retire eight of them by design (D-11-02, D-11-27, D-12-36, D-11-51, D-12-21, D-13-02b, D-13-10a, D-13-34), leaving D-10-C04 and D-12-43, the latter with a named future trigger in M005. The risk is **the accrual rate turning positive across two consecutive milestones.** M001 carried one known item past close (D-03-02); M002 carried ten. If M003 carries fifteen, the debt registers stop functioning as a control and become a wall, and the `tech_debt` closeout becomes the normal closeout — at which point the project has lost its single best honesty mechanism.

Three concrete WIP limits, stated as gate conditions rather than intentions (because M002 Key Lesson 2 records that a rule without an owner and a trigger does not prevent the pattern it names):

1. **M003 may not close with more than 5 open unowned debt items.** Enforced at the milestone audit, not discovered there.
2. **Every debt item names an owning phase at the moment it is recorded.** M002 Lesson 3: "flagged for human review is not a work item." D-12-21 blocked on D-11-51 with neither having an owner for two phases.
3. **`deferral_count >= 2` is a hard stop, mechanically.** D-10-60 already says this in prose and was violated anyway. Make it a field in a machine-checked register (§ Preserving high-signal findings) so it fails a test instead of a conscience.

**The secondary risk, from M001's retrospective: plan-count drift under review pressure** (3, 7, 10, 13, 14, 15 across M001's six phases — "later phases absorbed gap-closure work as plan count rather than as scope renegotiation"). M003's Phase 17 (arithmetic under optimization) is the phase most likely to drift, because divergence hunting is open-ended. Pre-register its corpus size and control count *before* the phase opens, the way D-09-40's Nyquist threshold was pre-registered before the count was taken. That mechanism already worked once in this project; reuse it.

---

## The first real program: concrete target, and which milestone lands it

**The honest answer: not in M003, and pretending otherwise would repeat the roadmap-vocabulary trap that LANGUAGE-MATURITY.md exists to prevent.**

Today the language cannot express a program at all. The complete keyword set is `module export fn data type let mut match try take borrow discard because defect foreign`; there are no numeric literals in the lexer; functions take exactly one parameter of the same type as their return. The 115-file, 4,189-line `.lang` corpus contains zero programs — every file is an admission-rule fixture, and the largest (193 lines) is mostly comment.

### The named target

```
examples/checksum.lang  —  lang run examples/checksum.lang input.bin
```

**A byte-checksum utility.** Open a file through the existing `foreign C` boundary, read it into a `Buffer`, walk the buffer accumulating a rolling 8-bit sum (or a table-free CRC-8), release the handle through the existing three-stage fallible-acquisition machinery, and print the result through a foreign `putchar`/`printf`. Forty lines. Auditable by a human in one screen. Genuinely useful.

Chosen over the obvious alternatives for specific reasons:
- **Not FizzBuzz** (LANGUAGE-MATURITY's own calibration point), because it needs string literals and string output, which drags in a whole value category for a program that does nothing. FizzBuzz is a *later* trophy, not the first one.
- **Not hello-world**, because it is already reachable today through a `foreign C` declaration and proves nothing new.
- **Checksum** because its ingredient list is exactly the M003→M005 arc and nothing else: arithmetic and literals (M003), a bounded loop (M004), multi-argument functions and indexed buffer access (M005), foreign resource acquisition (already shipped, M001 Phase 4). It requires no strings, no collections, no generics, no allocation, no recursion, and no effects. It is the smallest program that is a program.

**It lands at the close of M005.** At the observed cadence (5–6 days per milestone) that is roughly three weeks out. Say that number out loud in the roadmap rather than letting "M003: Computation" imply usability.

### Work backward — and make the target executable now

Check `examples/checksum.lang` into the repository **in M003 Phase 18**, as a **frontier fixture**: a program that is *refused*, with a test that pins the exact diagnostic that refuses it.

```
M003 close:  refused at `check.call_arity_unsupported` (or the loop token) — pinned
M004 close:  refusal moves past the loop to the indexing/arity boundary  — pinned
M005 close:  test flips to: checks, interprets, natively runs, and all
             four engines agree on the checksum of a frozen input file
```

This is the project's own idiom — an executable escape, a mutation-killed control, a test that fails in both directions — applied to the roadmap itself. It converts "how far along is the language" from a question that requires reading LANGUAGE-MATURITY.md into a test that runs on every CI invocation. It also makes overscoping visible: if a milestone claims progress but the frontier fixture's pinned diagnostic did not move, the claim is about assurance, not surface.

A second, smaller frontier fixture is worth adding for M003 alone, so the milestone has its own visible trophy: **`examples/luhn_digit.lang`** — a single function that doubles a digit and folds the overflow, i.e. one Luhn step. It is expressible in exactly the M003 language (literals, `+`, comparison, `match` on a two-alternative `data` type, return type ≠ parameter type) and it is a real algorithm fragment rather than a fixture.

---

## Preserving high-signal findings across milestones (mechanism recommendation)

**Is what exists enough? No — and the failure mode is already documented twice.**

What exists is strong and unusual: `STANDING-VERDICTS.md` (project-level verdicts that outlive milestone-scoped research), `LANGUAGE-MATURITY.md` (the calibration file, with its own re-verify commands and its own staleness triggers), seven phase debt registers with a closed severity vocabulary and a shape-asserting test, `.planning/spikes/` with a conventions file, and a 44-note `wiki/` design corpus with an explicit provenance vocabulary (Intent / Research / Synthesis / Candidate / Decision / Unresolved). Most funded language projects have none of this.

The hole is structural, not effort-related: **all of it is documents, and this project has two milestones of evidence that documents slip.** Nyquist validation slipped in M001, was named as Lesson 5, and slipped identically in M002. D-10-60 was written to prevent a third deferral and a third deferral was proposed anyway. The D-09-51 verdict flip was "flagged for human review" and sat unreviewed through three phases. The project's own best pattern is the answer to its own problem — Phase 11 eliminated two human-judgment items with tests rather than adjudicating them, on the reasoning that "a debt note is read once; a test runs on every CI invocation."

Five recommendations, in leverage order.

**1. Make `LANGUAGE-MATURITY.md` machine-checked. (Highest leverage, ~1 plan.)**
The file already carries its own re-verify commands. Promote them to a Go test that pins the current keyword set, the corpus file count and line count, the `len(Functions) != 1` guard count, the arity limit, and the presence/absence of arithmetic tokens — and fails when any of them changes without the file's snapshot being updated. This is precisely the self-invalidating reflection-test pattern already shipped for `protocol.Metrics` identity exclusion. It converts the project's most important calibration document from a thing someone must remember to update into a thing that cannot silently rot.

**2. Collapse seven phase debt registers into one queryable register.**
77 items across 7 files is *why* 10 went unowned: there is no single surface to query "what is open and who owns it." Promote to `.planning/DEBT.md` (or better, a JSON register with a schema test, matching `qlt01_registry.json`'s precedent) with fields `id, severity, opened_in, requirements, owning_phase, deferral_count, unblocking_trigger, state`. Two mechanical gates: **`owning_phase` may not be empty** (M002 Lesson 3), and **`deferral_count >= 2` fails the build** (mechanizes D-10-60, which prose did not). Phase registers stay as the authoring surface; the aggregate is generated and tested.

**3. Add `.planning/UNREACHABLE-CLAIMS.md` — the genuinely missing artifact.**
D-12-43, D-13-02b, and D-13-10a are a new *kind* of finding the project invented and has no home for: a criterion that is correct but **not constructible at the current language maturity**, escalated rather than downgraded. This is arguably the project's most valuable original methodological contribution, and right now those three rows live in phase debt files that get archived at milestone close. Each row should carry: the claim, the invariant that makes it unreachable, and **the trigger that would make it constructible**. Then every milestone charter can mechanically answer "which unreachable claims does this milestone make reachable?" — which is, not coincidentally, how M003's Phase 15 gate and M005's headline gate were derived in this document.

**4. Defuse `wiki/example-tour.md` mechanically.**
LANGUAGE-MATURITY calls it the "second trap, same shape" — it shows effect rows, `?` propagation, `defer ... unless`, generics, and named arguments, none of whose tokens exist in the lexer, and it reads like a language description. A README disclaimer has already failed to prevent the trap once. Add a test that every identifier-shaped token in `example-tour.md` either appears in `syntax/token.go` or is listed in an explicit `## Not yet implemented` block in the file itself. Cheap, mechanical, permanently kills a recurring calibration error.

**5. The frontier fixture (§ The first real program) is the sixth mechanism**, and it is the one that preserves the *roadmap's* intent rather than its findings: a pinned, executable statement of exactly how far the language is from being usable, that cannot drift from the truth.

**What should NOT change:** `STANDING-VERDICTS.md` is working — it is doing exactly what it claims (dependency verdicts, anti-features, load-bearing facts, reopen only on new evidence) and its "load-bearing technical facts" section saved real re-derivation work in preparing this document. Add two entries (`if` as a rejected anti-feature; `pathoracle` back-edge refusal as a named M004 one-way door) and otherwise leave it alone. Likewise `.planning/spikes/CONVENTIONS.md`, which correctly rejected S-008 as a category error.

---

## Role-lens disagreements

Where the lenses actually disagree, and how each disagreement resolves.

**Compiler lead vs. Product manager — on Phase 17 (arithmetic under optimization).**
The PM wants it cut or merged into Phase 16: it ships no user-visible capability, it is the phase most likely to drift open-ended, and users cannot tell whether `-O3` agrees with the interpreter. The compiler lead says Phase 17 *is* the milestone: it is the first time in the project's life that the five-axis comparator has a subject arising from the language rather than from a hand-engineered hazard, and C's integer-promotion and signed-overflow UB is where a naive emitter silently produces a different program. **Resolution: the lead wins, decisively.** Cutting Phase 17 makes M003 a feature milestone in a project whose entire differentiation is that features arrive with evidence. But the PM's drift concern is legitimate and is addressed by pre-registering Phase 17's corpus and control count before the phase opens.

**Software architect vs. Risk/program manager — on bundling arity-N into M003.**
The architect wants arity-N landed in M003 alongside the signature widening: both touch `check` admission, `core.Function`, and the interface schema (which already reserves `Parameters []ParameterContract`), so doing them together pays the six-dispatch-site coordination cost once instead of twice, and it is a reversibility argument — arity is a one-way door for the C calling convention. The risk PM says arity-N is a second load test of interprocedural loan liveness (simultaneous parameter loans, per-parameter public origins) and roughly doubles M003. **Resolution: the risk PM wins, but the architect's point is recorded as a real cost.** The precedent is the ratified M001→M002 `OpCall` deferral, which cost a second pass at six sites and was judged correct anyway. The mitigating fact is that D-07-10 already took the schema capacity, so the second pass is cheaper than the first was.

**Developer experience vs. Skeptical outside reviewer — on `if` and `for`.**
DX says a language without `if` is a science project: every practitioner expects it, and telling an author to write `match` over a two-alternative `data` type is the kind of purity that makes a language unpleasant. The skeptical reviewer says every keyword is a keyword that must be independently re-derived at six dispatch sites forever, `match` already provides branching with a real CFG, and the *actual* author is an AI agent for whom `match cond { True -> ..., False -> ... }` costs nothing. **Resolution: reviewer wins on `if`** (it buys zero expressiveness and creates a second control-flow law — Key Lesson 2). **DX wins on `for`, deferred**: iteration is genuinely load-bearing, which is why it gets a whole milestone rather than being dismissed.

**Product manager vs. Compiler lead — on who the user is.**
The PM observes that there are zero users, that the AI-author premise has never been tested against an agent that is not the development agent, and that this makes "minimum lovable surface" reasoning unfalsifiable — the real product today is the evidence protocol, and the roadmap should optimize for that. The lead says an evidence protocol with no programs to be evidence *about* is a tautology. **Resolution: both are right, and the disagreement is resolvable cheaply** — see the fresh-agent eval spike in the adversarial response below. This is the one lens disagreement that should change the plan: it adds a spike to M003.

**Risk/program manager vs. everyone — on the `tech_debt` closeout.**
The risk PM alone treats M002's closeout status as a leading indicator rather than a bookkeeping detail, and would spend an entire phase on debt before adding any surface. **Partially adopted:** Phase 14 is a debt phase and it goes first. Not fully adopted: Phase 18's validation work is scheduled at the *end* because validating a surface M003 is about to change is a documented anti-pattern.

---

## Prior art: how other languages sequenced, and what they regret

**Rust — cut aggressively pre-1.0, and rewrote the safety analysis anyway.** Typestate was removed in 0.4; green threads and the entire runtime were removed in late 2014, months before 1.0, via [RFC 230](https://github.com/nox/rust-rfcs/blob/master/text/0230-remove-runtime.md) — the abstraction had been made so heavyweight in the attempt to work across 1:1 and N:M threading that the green threads "barely even qualified as green anymore." The relevant lesson for this project is not the cutting, which Lang already does well; it is that Rust's *borrow checker itself* had to be redone. The original AST-based checker derived lifetimes from lexical scope, which was easy to implement but insufficiently flexible; non-lexical lifetimes required re-implementing the analysis against MIR as a liveness-and-constraint problem over program points ([rustc_borrowck](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_borrowck/index.html), [PR 45538](https://github.com/rust-lang/rust/pull/45538)). Lang's `loanLivenessFixpoint` is already the MIR-shaped design — a CFG liveness fixpoint over a typed core — which means Lang starts where Rust ended up. That is genuine evidence for the strength case. The counter-evidence is that Rust needed the expressive surface to *discover* that lexical scope was wrong.

**Zig — staged the compiler, and accepted named regressions to get there.** The stage1 C++ bootstrap compiler was retired via `-fstage1` removal ([PR 13383](https://github.com/ziglang/zig/pull/13383)) after 0.10.0 defaulted to the self-hosted compiler. What is instructive is the honesty of the transition: stage2 *regressed* relative to stage1 on comptime assembly, async/await/suspend/resume were temporarily regressed in master, and O(1) bootstrapping was [explicitly allowed to regress](https://github.com/ziglang/zig/issues/6378) with a stated plan to restore it before 1.0. Zig wrote down what it was giving up and when it would come back. That is the same discipline as Lang's disclosed-debt registers — and Zig's experience is a caution that "restore before 1.0" items have a way of outliving several releases.

**Swift — deferred the one-way door until the enforcement it depended on was complete.** ABI stability was deferred repeatedly and finally landed in Swift 5, and it landed *with* full runtime enforcement of exclusive access to memory, enabled by default in release builds ([Swift 5 Exclusivity Enforcement](https://www.swift.org/blog/swift-5-exclusivity/), [Swift 5 Released](https://www.swift.org/blog/swift-5-released/)). The reasoning is directly transferable: failing to fully enforce exclusivity would have had an unpredictable impact on ABI stability, because binaries built without full enforcement might work in one release and misbehave in the next. Read across: **do not freeze a one-way door before the analysis that makes it sound is complete.** For Lang, the door is the C calling convention that arity-N fixes, and the analysis is interprocedural loan liveness across multiple simultaneous parameter loans. Another argument for arity-N in M005, after iteration has stress-tested the liveness machinery.

**Go — deferred the big feature for a decade and defends it.** Generics shipped in Go 1.18 in March 2022, thirteen years after the language; Cox called it "the most significant change to Go since Go 1," and the stated reasoning for the wait was that a poorly designed generics system would be worse than none ([Generics — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-generics-overview.md); [research!rsc: The Generic Dilemma](https://research.swtch.com/generic)). The cost is real and should not be glossed: a decade of `interface{}`, code generation, and duplicated container libraries. But the relevant point for Lang is that LANGUAGE-MATURITY ranks ability-bounded generics as the **#1 post-M002 feature**, and Go is the strongest available evidence that #1-ranked is not the same as next. Generics are correctly absent from M003–M006.

**CompCert and CakeML — the closest structural analogues, and the most encouraging.** CompCert's verified core is Clight, a deliberately restricted C subset (pointer arithmetic, structs, unions, loops, structured switch, no `goto`), and the surface grew outward over years on a pipeline of 20 passes and 11 intermediate languages ([Clight, Leroy](https://xavierleroy.org/publi/Clight.pdf); [CompCert structure](https://www.absint.com/compcert/structure.htm); [compcert.org](https://compcert.org/)). CakeML likewise compiles high-level features away incrementally through 12 intermediate languages, specifically so each pass can be verified at the appropriate level of semantic detail, and eventually bootstrapped itself inside HOL4 across five architectures ([cakeml.org](https://cakeml.org/), [CakeML: a verified implementation of ML](https://dl.acm.org/doi/10.1145/2535838.2535841)). Both spent years with a surface that a working programmer would have called a toy, and both ended up with real, usable, verified compilers. **This is the single best external evidence that Lang's ratio is a strength.** The caveat is scale: both are multi-institution, multi-year, funded efforts, and both were verifying a *fixed, pre-existing* source language rather than co-designing one.

**Unison — the cautionary tale for unusual authoring models.** Unison stores code in a content-addressed database rather than text files; a structured editor was experimented with early and set aside in favour of core language work ([What's cool about Unison?](https://jaredforsyth.com/posts/whats-cool-about-unison/), [LWN](https://lwn.net/Articles/978955/)). The recurring adoption critique is that abandoning text-and-Git means abandoning an enormous surrounding tooling ecosystem, and that doing so requires a killer use case forcing enough to overcome it. **Lang partially escapes this**: PROJECT.md's source-authority constraint keeps text Git-friendly and formatter-owned, so Lang pays none of Unison's tooling-abandonment cost. But the deeper lesson still bites — "AI-authored, human-audited" is a thesis, not yet a forcing function.

**The "build one to throw away" literature, and Brooks's own retraction.** Brooks's *Mythical Man-Month* advice — "plan to throw one away; you will, anyhow" — is usually cited in support of prototypes, but Brooks himself later wrote that "this I now perceive to be wrong, not because it is too radical, but because it is too simplistic," on the grounds that it assumes a waterfall model, and argued for iterative development in *The Design of Design* ([second-system effect](https://en.wikipedia.org/wiki/Second-system_effect), [Brooks quotations](https://en.wikiquote.org/wiki/Fred_Brooks), [Tim Bray](https://www.tbray.org/ongoing/When/200x/2008/08/22/Build-One-to-Throw-Away)). This matters because the natural critique of Lang — "you built the verifier before the language; throw it away and start over" — invokes advice its own author repudiated. The *second-system* warning is the one that actually applies: the danger is the designer, now confident, incorporating every previously deferred improvement and generalization. Lang's five `.planning/spikes/` workbenches were the throwaway systems, and their surviving contracts are production code. The live risk is second-system scope inflation in M003, which is what the WIP limits above exist to contain.

**Formal-methods ROI — the contrarian note that most applies here.** Hillel Wayne's business case for formal methods is that specification-based design verification finds fundamental design problems fast and cheap, and that the main adoption barrier is pedagogical rather than technical ([The Business Case for Formal Methods](https://www.hillelwayne.com/post/business-case-formal-methods/), [Pragmatic Engineer interview](https://newsletter.pragmaticengineer.com/p/formal-methods-with-hillel-wayne)). But the part that should give this project pause is his contrarian read on AI: LLMs are "extremely bad" at writing properties, the bottleneck is automation of *intent*, not of proof, and lightweight approaches like property-based testing are the practical middle ground. Lang's entire agent-facing thesis is that precise structured verifier feedback makes AI authoring work. That is a bet *adjacent to* the thing Wayne says LLMs are bad at, and it has never been tested here against an agent other than the one building the compiler.

---

## Adversarial pass: the strongest outside critique of this direction, and an honest response

### The critique, stated as strongly as I can make it

**1. Two milestones, ~102,000 lines of Go across 25 packages, 721 commits, and the language cannot add two numbers or take two arguments.** Not "does not yet have a standard library" — cannot express the integer `1`. `core.Function.Parameter` is singular. `check.go:255/:3148/:3399` require a function's return type to equal its parameter type. The entire `.lang` corpus is 115 files and 4,189 lines, none of which is a program; the largest file is 193 lines and mostly comment. By any measure an uninvested engineer would apply, this is a verification framework with a language-shaped test input, not a programming language.

**2. The 70/10 ratio is not a fact about the world, it is a fact about the denominator.** 70% of the assurance needed *for this surface*. Each new operation kind costs 8–10 coordinated edits across six dispatch sites plus two exhaustive-dispatch controls, by the project's own D-04-22 standard. If assurance obligations grow superlinearly in surface — which six mutually-independent re-deriving consumers strongly suggests — then "do the hard part first" has the causation backwards: the hard part is not the ownership theory, it is *maintaining six independent re-derivations of an expanding semantics*, and that cost is entirely in front of the project.

**3. The trap is not hypothetical; it fired three times in M002 and is written in the audit.** D-12-43: a decisive control the harness cannot construct. D-13-02b: a blame resolver built, tested, exhaustive, with zero call sites. D-13-10a: a repair class withdrawn because all types are identical. Three of M002's headline deliverables were unreachable *for lack of language*. A reviewer is entitled to say: you built elaborate machinery, and when you finally tried to point it at something, it had nothing to point at.

**4. The AI-authoring premise is unfalsified and the market evidence points the other way.** Nobody has written Lang except the agent building Lang. The claim that AI authors need precise verifier feedback more than they need expressive syntax is plausible and untested. The single largest determinant of LLM code quality is training-corpus size, and a new language has a corpus of zero. No diagnostics protocol, however well-designed, compensates for that. And per Wayne, LLMs are specifically weak at exactly the specification-authoring skill this design leans on.

**5. The process signals are trending the wrong way.** M002 closed `tech_debt`, not `passed`. Ten unowned debt items, up from one carried past M001. A self-imposed no-third-deferral rule was written and then nearly violated by the very next phase. Nyquist compliance: 3/6, then 2/7. Sixty-one plans in six days is a velocity sustained partly by not closing things.

**6. Nobody asked for this.** Unison's lesson is that a language demanding ecosystem change needs a forcing function. "AI-authored, human-audited production software" is a hypothesis about the future of software authorship, not a customer. Three more milestones gets a checksum utility.

### Honest response

**Points 1, 3, and 4 land. I concede them.**

On **(1)**: the critique is right about the framing and the framing is the project's actual liability. "Language" is the wrong noun for what exists; "a semantic-evidence substrate that will eventually carry a language" is right. PROJECT.md and LANGUAGE-MATURITY.md already say this — LANGUAGE-MATURITY exists specifically to "stop re-discovering the gap between how sophisticated the verification stack sounds and how little the language can currently express." The problem is that the *roadmap vocabulary* keeps undoing that honesty, milestone after milestone. This is precisely why this document recommends a frontier fixture and a machine-checked maturity snapshot: not because the team is dishonest, but because the documents that tell the truth are the ones that go stale.

On **(3)**: yes, three times, in one milestone. But the *shape* of all three failures deserves weight, and the critique elides it. In every case the machinery was built correctly and the surface was missing. Not one M002 assurance mechanism has been shown to be *wrong*; three have been shown to be *unexercised*. "Built right, not yet loadable" is repaired by adding surface; "built against the wrong model" is repaired by rewriting. The project is in the first state. That distinction is the entire bet — and it is falsifiable, which is why this document attaches a trigger to it: **if M003's arithmetic work requires redesigning `loanLivenessFixpoint`, the comparator, or `corevalidate`'s forward propagation rather than extending them, the bet is lost and M004 becomes an assurance-refactor milestone.** If a reviewer will accept no evidence that would change their mind, the critique is not falsifiable either.

On **(4)**: fully conceded, and it is the most important unaddressed risk in the project. The premise has never been tested. It is also **cheap to test**, and the project has never run the experiment. Recommendation, added to M003 as a bounded spike under existing spike conventions:

> **S-009 — Fresh-agent authoring probe.** Give an agent with no project context the shipped `lang` binary, the `--json check` protocol, `lang explain`, and one sentence of intent. Have it author `examples/luhn_digit.lang` (M003's frontier fixture) using only the structured diagnostic channel. Measure iterations-to-green, and measure the same task with the diagnostic *messages* scrambled to lorem ipsum while the structured `repairs[]/kind/span/replacement` channel is preserved — the falsifiability control M001's Phase 6 already built and proved bites. **Falsifiable question:** does the structured channel carry the authoring signal, or is the agent succeeding on prose it could have gotten from any language? Stop when answered; it is a workbench, not production code.

That is the single highest-leverage cheap experiment available to this project right now, and it is more important than any individual M003 phase. If it comes back negative, the roadmap should change materially — toward corpus, examples, and tooling, and away from surface.

**Point 2 is the one I partially rebut, and it is measurable rather than arguable.** The superlinearity worry is real, but the project owns the instrument to settle it: `internal/compiler/measure`, the growth-exponent fitting already used for `buildInterproceduralSummaries` (a ≤1.2 exponent against operation count), and the QLT-02 budget manifest. The correct response is not argument but instrumentation: track **edits-per-new-`OperationKind`** and **plans-per-surface-feature** across M003–M005 and publish the curve. The current baseline is 8–10 edits/kind. Phase 12 landed *two* new operation kinds in a single 8-plan phase, which is weak early evidence against superlinearity for semantically-local additions. If the number rises, the critique wins on evidence.

**Point 5 is conceded and is why the scope section exists.** The WIP limits, the single machine-checked debt register, the mechanized `deferral_count >= 2` stop, and Nyquist-as-a-gate are direct responses. The accrual rate, not the absolute count, is the thing to watch — and M003 as chartered retires eight of the ten inherited items.

**Point 6 I partially rebut.** Lang does not pay Unison's tooling-abandonment cost: text stays Git-friendly and formatter-owned by constraint, and the toolchain is a conventional CLI over conventional files. But the substance — that there is no forcing function and no customer — is true, and the honest framing is that this is a research project with unusually good engineering hygiene, not a product with a market. That is a perfectly respectable thing to be. It just should not be described as the other thing.

**One thing the critique misses, in the project's favour.** The most unusual asset here is not the verification stack; it is the *institutional honesty machinery* — escalating a defect in the criterion rather than downgrading the assertion (D-12-43, D-13-10a); rejecting an integration checker's green because the tree disagreed; reporting that M001's own shipped held-out corpus was alpha-renamed rather than quietly fixing the fixtures. Very few projects at any funding level find and publish holes in their own already-shipped evidence. Whatever happens to the language, that discipline is transferable and worth preserving deliberately — which is the deeper reason § Preserving high-signal findings recommends mechanizing it rather than trusting it to documents.

---

## Confidence + what would change my mind

| Area | Confidence | Basis |
|------|-----------|-------|
| M003 content and internal ordering | **HIGH** | Grounded directly in the tree: `core.Function.Parameter` singular, the three `sameType` sites, `check.call_arity_unsupported`, the absent numeric-literal tokens, `pathoracle.cfg_back_edge`, and the audit's own three-cluster analysis. The ordering follows from those facts, not from taste. |
| M003 sizing (5 phases / 40–45 plans) | **MEDIUM-HIGH** | Extrapolated from two data points (62/5d, 61/6d) and the 8–10-edits-per-kind standard. Two points is a line, not a curve. |
| "Strength, not trap" at 70/30 | **MEDIUM** | Strong external precedent (CompCert, CakeML, Rust's AST→MIR rewrite) but the three M002 unreachability findings are real counter-evidence and the position rests on a distinction (unexercised vs. wrong) that M003 will test for the first time. |
| M004 as iteration, with a pathoracle successor spike | **MEDIUM-HIGH** | The back-edge refusal is verified in code; that iteration breaks pathoracle is a fact. That iteration is the right *next* milestone is a judgment. |
| M005 / M006 charters | **MEDIUM** | Coherent and dependency-justified, but three milestones out on a project whose surface has never been load-tested. Expect these to change. |
| Checksum as the first real program, landing end of M005 | **MEDIUM-HIGH** | The ingredient list is mechanically derived from the milestone arc. The date follows from the arc and inherits its uncertainty. |
| The AI-authoring premise | **LOW-MEDIUM** | Untested against any agent but the development agent. This is the weakest link in the whole project and S-009 is the cheapest available test. |

**What would change my mind:**

1. **Arithmetic forces a redesign, not an extension.** If M003 Phase 16 or 17 requires reworking `loanLivenessFixpoint`, the five-axis comparator, or `corevalidate`'s forward propagation, the trap hypothesis is confirmed: re-charter M004 as assurance refactoring and stop adding surface until the peers are stable against the new semantics.
2. **A pathoracle successor turns out to be cheap.** If a pre-M004 spike shows bounded loop unrolling to a declared depth preserves pathoracle's authority at acceptable cost, pull iteration forward into M003 as a sixth phase and land the first real program a full milestone earlier. This is the most likely upside surprise.
3. **S-009 comes back negative.** If a fresh agent cannot author a 15-line Lang function from the structured protocol alone, the core value proposition is wrong and the roadmap should pivot toward corpus, examples, and authoring tooling over language surface.
4. **Edits-per-operation-kind rises materially.** If M003's two operation kinds cost substantially more than 8–10 coordinated edits each, superlinearity is real and the peer architecture needs a cheaper extension mechanism (a generated dispatch table, a single derivation with mechanical peer generation) before any further surface lands.
5. **M003 closes with more unowned debt than it inherited.** Two consecutive `tech_debt` closeouts with a rising accrual rate would mean the debt registers have stopped being a control. At that point a full debt-only milestone is the correct response, however unsatisfying.
6. **A real external user appears.** Any actual adopter's first request should outrank every judgment in this document.

---

## Sources

**Project files read (all claims about the project are grounded in these):**
`.planning/PROJECT.md` · `.planning/LANGUAGE-MATURITY.md` · `.planning/STANDING-VERDICTS.md` · `.planning/RETROSPECTIVE.md` · `.planning/ROADMAP.md` · `.planning/MILESTONES.md` · `.planning/milestones/M001-ROADMAP.md` · `.planning/milestones/M002-ROADMAP.md` · `.planning/milestones/M002-MILESTONE-AUDIT.md` · `.planning/spikes/MANIFEST.md` · `wiki/README.md` (corpus index)

**Code verified directly (not taken from documentation):**
`internal/compiler/core/core.go:127-142` (`Function.Parameter`, singular) · `core.go:289-310` (`FunctionSignature.Parameters` slice, D-07-10 capacity note) · `internal/compiler/check/check.go:255,:3148,:3399` (`sameType(ReturnType, Parameter.Type)`) · `check.go:3430-3449` (arity fixed at 1, `check.call_arity_unsupported`) · `internal/compiler/pathoracle/pathoracle.go:130-143,:243` (`pathoracle.cfg_back_edge`) · `internal/compiler/syntax/token.go` (keyword set; no arithmetic or numeric-literal tokens) · `testdata/phase07/call_basic.lang` · corpus census: 115 `.lang` files, 4,189 lines, 193-line maximum

**External:**
- [RFC 230: Remove runtime (Rust green threads)](https://github.com/nox/rust-rfcs/blob/master/text/0230-remove-runtime.md) — HIGH
- [rustc_borrowck documentation (MIR borrowck / NLL)](https://doc.rust-lang.org/nightly/nightly-rustc/rustc_borrowck/index.html) — HIGH
- [rust-lang/rust PR 45538 — enable NLL in the MIR borrow checker](https://github.com/rust-lang/rust/pull/45538) — HIGH
- [ziglang/zig PR 13383 — remove `-fstage1`](https://github.com/ziglang/zig/pull/13383) — HIGH
- [ziglang/zig issue 6378 — allow O(1) bootstrapping to temporarily regress](https://github.com/ziglang/zig/issues/6378) — HIGH
- [Zig 0.10.0 release notes](https://ziglang.org/download/0.10.0/release-notes.html) — HIGH
- [Swift 5 Exclusivity Enforcement](https://www.swift.org/blog/swift-5-exclusivity/) — HIGH
- [Swift 5 Released (ABI stability)](https://www.swift.org/blog/swift-5-released/) — HIGH
- [Go 2 draft: Generics — Problem Overview](https://go.googlesource.com/proposal/+/master/design/go2draft-generics-overview.md) — HIGH
- [research!rsc: The Generic Dilemma](https://research.swtch.com/generic) — HIGH
- [Leroy, Mechanized semantics for the Clight subset of C](https://xavierleroy.org/publi/Clight.pdf) — HIGH
- [The structure of CompCert (AbsInt)](https://www.absint.com/compcert/structure.htm) · [compcert.org](https://compcert.org/) — HIGH
- [CakeML project site](https://cakeml.org/) · [CakeML: a verified implementation of ML (POPL'14)](https://dl.acm.org/doi/10.1145/2535838.2535841) — HIGH
- [What's cool about Unison? (Jared Forsyth)](https://jaredforsyth.com/posts/whats-cool-about-unison/) · [Programming in Unison (LWN)](https://lwn.net/Articles/978955/) — MEDIUM
- [Second-system effect (Wikipedia)](https://en.wikipedia.org/wiki/Second-system_effect) · [Fred Brooks quotations (Wikiquote)](https://en.wikiquote.org/wiki/Fred_Brooks) · [Tim Bray, Build One to Throw Away](https://www.tbray.org/ongoing/When/200x/2008/08/22/Build-One-to-Throw-Away) — MEDIUM
- [Hillel Wayne, The Business Case for Formal Methods](https://www.hillelwayne.com/post/business-case-formal-methods/) · [Pragmatic Engineer: Formal methods with Hillel Wayne](https://newsletter.pragmaticengineer.com/p/formal-methods-with-hillel-wayne) — MEDIUM-HIGH

---
*Written 2026-09-17 for M003 planning. Not committed.*
