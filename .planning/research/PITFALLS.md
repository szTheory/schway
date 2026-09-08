# Pitfalls Research: Interprocedural Semantic Spine (M002)

**Domain:** Compiler engineering — adding Lang-to-Lang calls (`OpCall`), call
graphs, interprocedural loan liveness, interprocedural `-O3`/LTO equivalence,
and payload-carrying `Result` values to an existing intraprocedural typed-core
compiler (Go 1.24 host, affine ownership + borrow checker, independent
`corevalidate`/`originvalidate` re-derivation, deterministic interpreter
oracle, C17→Clang `-O0`/`-O3`/`-flto`, ASan/UBSan, HDD reducer, content-bound
cache).

**Researched:** 2026-09-08
**Confidence:** MEDIUM-HIGH — the general compiler-literature classes below
are verified from primary sources (issue trackers, papers, specs); the
project-specific mapping (Section 8 / "Process/Anti-Patterns") is HIGH
confidence because it is inferred directly from this project's own recorded
architecture and the M001 retrospective, not from external precedent.

## Critical Pitfalls

### Pitfall 1: Cross-function borrow/lifetime unsoundness at the call boundary

**What goes wrong:**
A function is proven sound in isolation — its parameters' loans are correctly
tracked within its own CFG — but the *composition* of two functions leaks a
loan the composed system does not know about. Concretely: a function returns
a borrow derived from a parameter without the origin being correctly threaded
through the call site; or a callee's exclusive loan on a parameter is treated
as compatible with a caller-side shared loan still live across the call.
Rust's own history shows this is not a hypothetical: variance and higher-ranked
lifetime bugs recur specifically at *composition* boundaries (closures,
trait-object return positions, associated types) rather than within a single
function body. Examples verified from the tracker: rust-lang/rust#84366
(`'static` closures/futures with non-`'static` return type are unsound — an
`I-unsound` labeled issue, i.e. accepted, safe code, no `unsafe`, compiles,
UB at runtime); rust-lang/rust#106431 (implied bounds/well-formedness of
references treats contravariant lifetimes the same as covariant — a variance
confusion at a function-signature boundary); rust-lang/rust#71550
(unsoundness from variance of trait objects with respect to associated
types); rust-lang/rust#123241 (async closures reject valid captures when one
lifetime is invariant — the opposite failure mode, false rejection from
overly conservative variance, which is the "fail closed" cost of getting
variance interprocedurally right). Every one of these is exactly the failure
class Codename Lang is opening the door to with `OpCall`: a *return value's
origin* now depends on facts declared or inferred at another function's
boundary, and variance/subtyping-shaped reasoning about "is this loan still
compatible across the call" is new load-bearing logic that did not exist
intraprocedurally.

**Why it happens:**
Intraprocedural analysis proves facts about values *within one CFG*, where
the analysis has full visibility of every use. The moment a call crosses a
function boundary, the callee's obligations must be *summarized* — declared
origins, ability requirements, loan liveness at entry/exit — and the summary
is necessarily an approximation of the callee's real behavior. Unsoundness
appears exactly where the summary is looser than reality (missing constraint,
so the caller accepts something unsafe) or where the summary and the actual
callee body silently diverge over time (someone changes the callee's body
without updating what the summary asserts). Rust's unsoundness backlog is
dominated by exactly this pattern: a local, textually-scoped check (variance
of a single signature) that turns out to interact unsoundly with a caller
context the checker never modeled together with it.

**How to avoid:**
- Do not infer interprocedural loan facts implicitly from callee bodies the
  way M001 infers local borrow ends at CFG last use. Require declared
  callee-side contracts (parameter loan requirements, return-origin
  declarations) at every `OpCall` site, checked at the call site against the
  declared contract — never against a re-walk of the callee's CFG. This
  mirrors D-03-02's already-decided direction (declared public origins,
  verified not trusted) and extends it to loans, not just origins.
  Make the transition point in the codebase where "local inference" stops and
  "declared contract" starts a literal `OperationKind` boundary so it is
  never ambiguous which rule applied to which value.
- Treat every declared interprocedural contract as data that both `check` and
  `corevalidate` re-derive independently from the *same* typed-core artifact
  (never from each other's output) — see Pitfall 2.
- Build the adversarial corpus for this milestone around the exact shape of
  the cited Rust issues: a callee returning a borrow derived from a
  caller-supplied parameter under two callers with different loan states at
  the call site (one where the loan is dead, one where it's live);
  cross-function variance-shaped cases (a caller passing a "more permissive"
  loan where the callee declares a "less permissive" one, and vice versa).
- Do not resolve D-03-02 ("an exported borrow-derived return with no declared
  origin, in its interprocedural half") by pattern-matching the intraprocedural
  fix. The M001 retrospective explicitly named this the "revisit" item; the
  contract must be re-derived for the call-boundary case, not copied.

**Warning signs:**
- Any place `check` computes a call site's admissibility by looking *into*
  the callee's CFG rather than at its declared contract. This is the
  intraprocedural instinct leaking into new code and is the single highest-
  risk code smell for this milestone.
- A test corpus for interprocedural loans where every counterexample is
  single-call-depth. Real unsoundness in the cited issues appears at
  composition of ≥2 boundaries (closure capturing a reference, returned from
  a function, called from another function) — Codename Lang's analog is
  call chains of depth ≥2 with divergent loan states, not just one call.
- A green interprocedural test suite with zero rejected-by-design negative
  cases mirroring the Rust issues above.

**Phase to address:**
The phase introducing declared callee contracts and call-site admissibility
checking (interprocedural loan liveness in `check`) — this must be scoped
before or together with `OpCall` dispatch-site gating, not after. Make the
negative-control corpus (composition-depth-≥2 divergent-loan-state cases) a
gate condition of that phase, not a follow-up.

---

### Pitfall 2: `check` and `corevalidate` silently drift on cross-function facts

**What goes wrong:**
M001's peer architecture — `check` (iterative worklist) and `corevalidate`
(reachability-closure-plus-reduction) independently re-deriving the same
loan-endpoint facts — is exactly the kind of dual-admission-layer design
whose entire value proposition is "if both agree, the fact is real." M001's
own retrospective already recorded the concrete failure mode this produces
when scope is uneven: D-02-03's quadratic-cost fix "landed in the wrong
half" — the validator half closed, the admission-deciding path did not — and
Lesson 3 states it plainly: "fixing one of N independent peers is not fixing
the item." Interprocedural facts multiply the number of places this can
happen, because now there are *two* new cross-cutting facts (call-graph
membership, interprocedural loan liveness) each needing independent
re-derivation in both `check` and `corevalidate`, and the two derivations
are architecturally required to differ in algorithm (worklist vs.
reachability-closure) — which is precisely what makes them valuable, and
precisely what makes them prone to independently-introduced bugs that happen
to agree with each other on the visible test corpus while disagreeing on an
untested shape.
Precedent for this class of drift: LLVM's IR verifier checks *shape*
invariants (types, dominance, attribute well-formedness) but does not, and
architecturally cannot, verify that individual optimization passes preserve
*semantic* invariants across a transformation — passes have repeatedly
produced IR that passes the verifier but is semantically wrong (this is
exactly why LLVM ships `-verify-each` as a debugging aid rather than a
soundness proof: passing the verifier is a necessary, not sufficient,
condition). CompCert's own design responds to this exact risk by choosing,
pass-by-pass, between a theorem-proved transformation and a translation-
validated one (verified from primary sources: CompCert's frontend/backend
passes are proved in Coq for semantic preservation; select passes such as SSA
construction/destruction and instruction scheduling instead use translation
validation — checking each individual compilation's output against its input
— because proving the general transformation was judged too costly). The
lesson generalizes: whichever of `check`/`corevalidate` is cheaper to extend
first for interprocedural facts will be tempting to treat as "the" fix, with
the other layer's extension deferred — reproducing D-02-03/D-03-01 exactly.

**Why it happens:**
Two independent implementations of the same fact are, by construction, two
separate places where a scope decision, an edge case, or an off-by-one can be
made *differently* — and because they usually implement the underlying rule
correctly on the common-case corpus, they agree there and only diverge on the
input shapes neither implementer thought to write a test for. Under time
pressure, a scope-narrowing decision ("handle this for direct calls, defer
recursive/cyclic calls") tends to get made once, in the layer being actively
worked on, and not mirrored to the peer layer — especially because the peer
layer's re-implementation is a distinct piece of work with its own review
cycle, and "keep both peers in lockstep scope" is not something either
layer's diff makes visible by itself.

**How to avoid:**
- For every new interprocedural fact (call-graph reachability, interprocedural
  loan liveness, cross-function origin), write the scope statement once, as a
  structured artifact (mirroring `*-DEBT.md`'s discipline), and require both
  `check` and `corevalidate`'s implementation PRs to cite it and assert
  their scope matches it — a mechanical, checkable statement, not prose.
  Roadmap-level: land the `check` and `corevalidate` halves of each
  interprocedural fact in the *same* plan/phase, never split across phases
  the way D-02-03 was.
  A shadow-run comparison (the technique 05-02 used for 230,692 admission-site
  comparisons before deleting the old intraprocedural liveness law) should be
  run for every new cross-function fact at the point both derivations exist,
  specifically diffed over call-graph shapes neither peer's author wrote by
  hand: recursion, mutual recursion, diamond call graphs, and deep call
  chains — not just the shapes each implementer thought to test.
- Route interprocedural facts through the *same* typed-core artifact both
  layers already read from (never let one layer consume the other's output),
  preserving the "shared inert records only" rule from M001's established
  pattern.
- Track, as a gated milestone-close check, that every peer-layer pair of
  functions handling a given interprocedural fact has equal test coverage
  (same corpus fed to both), so scope drift produces a visible coverage gap
  rather than a silent one.

**Warning signs:**
- A fix or feature-add PR that touches `check`'s interprocedural logic
  without a corresponding diff in `corevalidate` (or vice versa) in the same
  commit/plan.
- Two admission layers agreeing on 100% of the milestone's own generated
  corpus but never having been run against each other on hand-adversarial,
  out-of-corpus call graphs (the M001 lesson: "out-of-corpus proofs caught
  what in-corpus tests could not" — this applies doubly to peer-agreement
  claims, since a generator built by one mental model tends to also miss the
  shapes that break peer agreement).
- Debt items describing an interprocedural fix landing in only one of the two
  layers (the D-02-03/D-03-01 shape recurring for a new fact).

**Phase to address:**
The phase(s) introducing call-graph construction and interprocedural loan
liveness. Make "both layers, same plan, shadow-run diffed on adversarial call
graphs before either ships" an explicit gate condition, analogous to how
mutation-kill was retrofitted as a rule after M001 learned it the hard way —
except this time write it into the plan up front rather than relearning it.

---

### Pitfall 3: An `-O3`/LTO tier that looks proven interprocedurally but is inert

**What goes wrong:**
M001 already solved this problem intraprocedurally: `restrict` is emitted
only on checker-proven no-alias parameters, and the engineered
`control:alias.false_no_alias` negative control proved the comparator can
actually detect a real `interpreter == -O0 (2) != -O3 (7)` divergence on
Apple Clang 21. The interprocedural extension of this problem is harder and
has a much larger literature of real miscompiles behind it. The specific
failure mode — a proof tier that *looks* like it exercises the interesting
optimizer behavior but structurally cannot, because the negative control was
never re-derived for the new dimension — is exactly what M001's Lesson 1
("a control you have never seen fail is a claim") warns about, applied now to
LTO/interprocedural inlining specifically. Concrete precedent: LLVM issue
llvm/llvm-project#122537 (TBAA/GVN miscompile from incorrectly replacing a
pointer with `undef`) is dated 2025 and specifically involves a function that
was *not* originally inlined becoming inlined after simplification passes —
exactly the "a call that was opaque yesterday is transparent today" shift
that interprocedural `-O3`/LTO introduces. Rust's own multi-year `noalias`
saga (rust-lang/rust#31681, rust-lang/rust#54878) is directly on point: Rust
emitted `noalias` on `&mut` parameters, LLVM inlined functions carrying that
attribute, and the *scoped `!noalias` metadata LLVM would need to preserve
the guarantee after inlining was never emitted* — so the attribute was
correct locally and silently wrong once two functions were combined by the
optimizer, for years, until it was disabled. The mechanism generalizes
directly: `restrict`/no-alias facts proven true *of a single function's
parameters* do not automatically remain true *of the values after inlining
composes two functions' provenance*, because pointer provenance itself does
not automatically compose across a call the way scalar facts do — this is
precisely the problem the C provenance TS (WG14 N3005, "A Provenance-aware
Memory Object Model for C") exists to formalize: provenance must be tracked
through "exposure" and "synthesis" rules across the whole program, not
assumed to survive composition by default.

**Why it happens:**
An intraprocedural no-alias/`restrict` proof and an interprocedural one are
different theorems even when they use the same checker machinery, because
the interprocedural version must additionally prove the fact survives
*composition* — inlining, cross-TU optimization, and any transformation LTO
enables that was not reachable at `-O0`/no-LTO. It is easy to extend the
attribute-emission code to also emit interprocedural `restrict`/no-alias
without extending the *negative control* to a shape that can only fail
interprocedurally (an aliasing violation only detectable after two functions
are combined) — leaving a proof tier that runs, passes, and asserts nothing
new.

**How to avoid:**
- Before trusting the interprocedural `-O3`/LTO equivalence claim, construct
  a deliberately-broken interprocedural aliasing/`restrict` case analogous to
  `false_no_alias`, but requiring cross-function composition to manifest
  (e.g., two functions independently and locally alias-clean, whose
  composition after inlining is not) and prove the five-axis comparator
  actually distinguishes `interpreter`/`-O0`/`-O3`(non-LTO)/`-O3+flto` on it,
  exactly as 05-02 proved for the intraprocedural case.
  Do not consider interprocedural `-O3` equivalence "proven" for this
  milestone until that control exists and has been shown to fail red.
- Treat "the function got inlined and stopped being opaque to the optimizer"
  as a first-class differential axis, not an incidental detail: the corpus
  should include the same source-level program compiled once in a shape
  Clang inlines and once in a shape it does not (e.g., via a size/complexity
  threshold or `noinline`-equivalent), and require the comparator to be
  identical either way — this directly targets the class both llvm/llvm-
  project#122537 and Rust's `noalias` history hit.
- Where the checker proves an interprocedural no-alias fact, prefer emitting
  it as LLVM/Clang scoped alias-analysis metadata attached at the call site
  (mirroring what Rust's fix direction eventually required — emitting real
  scoped `!noalias` metadata rather than relying on the attribute alone to
  survive inlining) rather than trusting the C-level attribute to propagate
  correctly through Clang's own inliner unassisted; verify this with the
  comparator, don't assume it from the C standard's `restrict` semantics.

**Warning signs:**
- The interprocedural `-O3`/LTO equivalence claim ships with all-green
  results but no engineered adversarial subject specific to the
  interprocedural dimension (mirrors the exact gap M001 closed for the
  intraprocedural case in Phase 5 — do not let M002 reopen the gap one level
  up).
- Any code path that emits `restrict`/no-alias attributes at call sites based
  purely on the checker's intraprocedural-style per-parameter reasoning,
  without a distinct interprocedural provenance/aliasing judgment.
- Corpus programs that are all small enough, or shaped such that Clang's
  inliner behaves identically whether or not LTO/`-O3` is on — i.e., the
  interprocedural optimizer surface is never actually exercised differently
  from the intraprocedural one.

**Phase to address:**
The phase claiming interprocedural `-O3`/LTO equivalence. Require the
engineered interprocedural negative control (and its proof of failing red)
as a gate condition before the "equivalence holds" requirement is marked
validated — mirroring how M001 gated the intraprocedural claim on
`false_no_alias`.

---

### Pitfall 4: Fixpoint non-termination, SCC mishandling, and stack blowup in interprocedural analysis

**What goes wrong:**
Three related failure shapes appear specifically once analysis must handle
recursive/cyclic call structure and unbounded call depth, none of which
exist in a purely intraprocedural analysis:
1. **Non-termination or wrong-answer fixpoints on recursive/mutually
   recursive functions.** An interprocedural liveness/loan analysis is
   naturally formulated as a fixpoint over the call graph; if recursive
   cycles (self-recursion or mutual recursion through an SCC) are not
   detected and special-cased, the naive worklist either loops forever or
   converges to an unsound (too-permissive) answer because the analysis
   assumed acyclicity when computing per-function summaries.
2. **Combinatorial/exponential blowup from context sensitivity.** If the
   analysis is context-sensitive (tracks per-call-site or per-argument-shape
   facts rather than one summary per function), the number of contexts to
   analyze can grow exponentially with call-graph depth and fan-out —
   well-documented in the interprocedural dataflow-analysis literature (the
   IFDS/IDE family exists specifically to make context-sensitive
   interprocedural analysis polynomial rather than exponential by
   construction; naive approaches without that structure are the ones that
   blow up).
3. **Native stack overflow in the checker/interpreter itself** when checking
   or interpreting deeply nested or recursive call chains, because Go's
   default goroutine stack grows but is not unbounded, and a recursive-
   descent implementation of call-graph traversal, loan-liveness fixpoint, or
   the interpreter's own call-stack walking can overflow on adversarial input
   long before any language-level "unbounded interpreter call stack" limit is
   reached in the interpreter proper.

**Why it happens:**
Intraprocedural CFG analysis is naturally acyclic in the "call" dimension —
loops within one function are handled by dataflow fixpoints over a bounded
CFG, a well-trodden problem. The call graph, by contrast, is not guaranteed
acyclic (Codename Lang's own roadmap explicitly names "cycle refusal," which
is exactly the right instinct — but refusal has to be detected correctly,
including through indirect/mutual cycles, not just direct self-calls) and is
not guaranteed bounded in depth for algorithms that walk it structurally
(as opposed to bounding it via the interpreter's declared call-stack limit).
Analysis authors who have only ever worked intraprocedurally tend to port a
"walk it structurally" mental model to the call graph without first asking
"is this graph guaranteed acyclic and bounded here, or do I need an explicit
visited-set/SCC/worklist algorithm."

**How to avoid:**
- Compute strongly-connected components (SCCs) of the call graph explicitly
  as a first step, and require every recursive-cycle-containing SCC to be
  refused per the roadmap's own "cycle refusal" requirement — but verify
  refusal triggers on *indirect* mutual recursion (A calls B calls A through
  an intermediate function or through a stored/matched `Result` alternative)
  and not only on direct self-recursion, since the M002 scope also adds
  `Result` payload matching, which is a plausible place for a recursive call
  to be structurally hidden from a naive direct-call-only cycle check.
  Codename Lang's own established pattern ("graph walks carry visited-set
  guards and named refusal codes rather than hanging") already states the
  right discipline — the M002-specific risk is applying it consistently to
  *every* new graph walk introduced this milestone (call-graph construction,
  interprocedural liveness fixpoint, cross-function origin propagation),
  not just the first one written.
- For interprocedural loan liveness, prefer function-level summaries
  (one fact per function, computed once the call graph's SCCs are ordered
  topologically, summaries for a whole SCC computed together via one
  bounded fixpoint) over unbounded per-call-site context sensitivity, unless
  a specific requirement demands the latter — this keeps the analysis
  polynomial rather than risking exponential blowup, consistent with the
  project's compute-efficiency constitution.
- Convert the checker's and interpreter's call-graph/call-stack traversal
  code from native recursion to an explicit worklist/stack data structure
  wherever traversal depth is attacker/generator-controlled (i.e., driven by
  program structure rather than a bounded internal loop), so a pathological
  input triggers the intended "bounded interpreter call stack" refusal
  rather than a Go runtime stack overflow that bypasses the language-level
  diagnostic entirely.
- Add an explicit fixpoint-iteration-count assertion (fail closed past N
  iterations with a named refusal code) to every new interprocedural
  fixpoint, mirroring the existing "fail closed on ambiguity" pattern, so a
  future bug that reintroduces non-termination fails loudly in test rather
  than hanging CI.

**Warning signs:**
- Any new call-graph or interprocedural-fact traversal implemented as plain
  native recursion without a visited-set or explicit stack.
- A cycle-refusal test corpus that only contains direct two-function
  self-recursion (A calls A, or A calls B calls A) and no deeper/indirect
  cycles, no cycles routed through `Result` alternative matching, and no
  cycles that only become visible after some other transformation.
- No test that intentionally constructs a pathologically deep (but acyclic)
  call chain to check whether checker/interpreter code overflows the Go
  stack before the language-level bounded-call-stack limit fires.

**Phase to address:**
The phase introducing call-graph construction and cycle refusal owns SCC
detection and traversal-safety; the phase introducing interprocedural loan
liveness owns fixpoint termination and complexity bounding. Both should gate
on an adversarial corpus containing indirect cycles, cycles through `Result`
matching, and pathologically deep acyclic chains — not just the direct-cycle
case the roadmap language names explicitly.

---

### Pitfall 5: The intraprocedural random-program generator systematically under-covers cross-function cases, and the reducer may not preserve interprocedural properties

**What goes wrong:**
Two related testing traps:
1. **Generator under-coverage.** A property/fuzz generator built and tuned
   for M001's intraprocedural corpus encodes, in its production rules and
   probability weights, an implicit model of "what a program looks like"
   that was shaped entirely by single-function semantics. Naively adding an
   `OpCall` production rule to the same generator produces call graphs that
   are shallow, rarely cyclic, rarely mutually recursive, and rarely exercise
   divergent loan states at a call site across different callers — precisely
   the shapes Pitfall 1 and Pitfall 4 depend on for their counterexamples —
   because the generator was never weighted to explore call-graph shape as
   an independent dimension. This is the exact failure class M001's own
   D-04-21 register was built to catch ("each row must name a shape the
   generator does *not* reach") and the retrospective's explicit lesson
   ("interrogate what a property test actually reaches, not what it
   nominally covers"; "drive the shipped binary on hand-written programs...
   out-of-corpus proofs caught what in-corpus tests could not"). For M002
   this generalizes concretely: call-graph shape (depth, fan-out, SCC
   structure, diamond patterns, argument-loan-state variation across call
   sites to the same callee) needs its own reachability register entries,
   independent from the existing arm/loan-endpoint register.
2. **Combinatorial call-shape explosion.** The naive way to be thorough is
   to cross every callee loan-contract shape with every caller loan-state
   shape with every call-graph topology — which is combinatorially
   explosive and cannot be exhaustively run, unlike M001's exhaustive
   loan-endpoint differentials (which the roadmap explicitly wants "rebuilt"
   cross-function — note the word "rebuilt," implying exhaustiveness was a
   real intraprocedural property that needs a genuinely new, not merely
   reused, bounding strategy interprocedurally).
3. **HDD reducer correctness across function boundaries.** M001's core-level
   HDD reducer was built and interestingness-tested against single-function
   or flat programs. Delta-debugging a multi-function program risks removing
   an *unrelated* function (or inlining/merging call sites) in a way that
   accidentally changes which interprocedural fact is exercised, producing a
   "reduced" counterexample that (a) no longer reproduces the original bug,
   or worse (b) reproduces a *different, unrelated* bug — silently
   invalidating the reduction. Delta debugging is well known in the general
   literature to require domain-aware reduction passes (e.g., "inline this
   call" or "remove this function only if no live call references it") to
   remain both sound and minimal on structured, cross-referencing programs;
   a reducer built for single-function programs has no reason to already
   have such passes.

**Why it happens:**
Test infrastructure encodes assumptions about the shape of the system it was
built against. A generator's production-rule weights and a reducer's
interestingness/minimization passes are both, in effect, compiled artifacts
of "what mattered when this was written" — and neither self-updates when the
system grows a new structural dimension (call graphs) that the original
design never had to represent.

**How to avoid:**
- Do not extend the existing generator in place and assume coverage
  transfers. Explicitly design call-graph shape as an independent generation
  dimension (depth, branching factor, presence/absence of recursion,
  diamond/shared-callee patterns, per-call-site argument loan-state
  variation) with its own weight knobs, and add a Generator and Probe
  Reachability Register section (mirroring D-04-21) enumerating shapes the
  call-graph generator does *not* reach, reviewed the same way the existing
  register is.
- Bound the call-shape combinatorics honestly and in writing: pick a small
  number of named axes (e.g., call-graph topology class × loan-contract
  shape class × recursion class) and enumerate their cross product
  exhaustively within a stated small bound (matching M001's "exhaustive loan-
  endpoint differentials," now literally rebuilt for the cross-function case
  per the roadmap's own language), while explicitly declaring anything beyond
  that bound out of exhaustive scope and covered only by weighted random
  generation — do not let "exhaustive" silently become "as exhaustive as the
  intraprocedural version happened to be, extended ad hoc."
  Continue to supplement with hand-written, out-of-corpus multi-function
  programs specifically, since M001 already proved these catch what the
  generator/corpus cannot.
- Before trusting the reducer on multi-function counterexamples, add
  reduction passes aware of function boundaries and call edges (e.g.,
  "delete an unreferenced function," "inline a trivial callee," "collapse a
  call chain") and add a differential test that the reducer's *output*, on a
  representative interprocedural counterexample, still reproduces the exact
  same failing property (not merely "some" failing property) as the
  unreduced input — this is the mutation-kill discipline applied to the
  reducer itself, and should be budgeted into the plan that extends the
  reducer for multi-function programs, per Lesson 1.

**Warning signs:**
- The interprocedural reachability register (or its absence) — if call-graph
  shape isn't tracked as its own dimension with named unreached shapes, this
  pitfall is already active.
- Reducer output on a multi-function bug report that, when re-run, no longer
  fails the same assertion/property it was reduced from (a correctness bug
  in the reducer, not the compiler).
- "Exhaustive" interprocedural differential claims with no explicit
  stated bound on call-graph depth/fan-out — if the bound isn't written
  down, it isn't known to be honest.

**Phase to address:**
The phase rebuilding Phase 3's exhaustive loan-endpoint differentials
cross-function owns the honest bounding statement and the reachability
register extension. Any phase extending the HDD reducer for multi-function
input owns the reduced-output re-verification mutation-kill, budgeted into
that same plan rather than deferred.

---

### Pitfall 6: Content-bound per-unit cache invalidation misses whole-program (call-graph) facts

**What goes wrong:**
A cache keyed on a single unit's content hash (M001's evidence cache is
explicitly "structurally incapable of holding a verdict," keyed presumably
per compilation unit/function/file) is a poor fit for a fact that is
*inherently whole-program*: the call graph, and any interprocedural fact
derived from it (loan liveness across calls, interprocedural `-O3`
equivalence). If function A's cache entry is keyed only on A's own content
hash, and function B (which A calls) changes in a way that alters A's
interprocedural loan-liveness or aliasing facts without changing A's own
source text, a naive per-unit cache will serve A's stale interprocedural
verdict — a classic "unit changed" vs. "unit's transitive dependency
changed" invalidation miss. This is a well-documented class in real
incremental-compilation and incremental-build systems: Rust's own
incremental compilation system layers a dependency graph precisely to avoid
this trap (a change to a callee must invalidate the caller's cached
codegen/analysis units that depend on it, not just the callee's own unit),
and this dependency-graph machinery is a large and historically bug-prone
part of `rustc` (ThinLTO's incremental mode specifically persists an
"import map" across sessions expressly to re-invalidate a module when
something it imported from changed — evidence that "did the thing I directly
import from change" is a distinct, necessary invalidation signal beyond "did
my own content change"). Bazel and Buck's cross-module invalidation bugs are
the build-system analog of the same root cause: a target's build graph
dependency edges must be complete and precise, or a downstream target is
served stale output when an upstream dependency it doesn't textually
reference (but does semantically depend on, e.g. through a transitively
included header or an ABI-relevant type) changes.

**Why it happens:**
A per-unit content hash is sufficient for a purely intraprocedural fact by
definition — if the fact only depends on the unit's own text, the unit's own
hash is a sound cache key. The moment a fact (call-graph membership,
interprocedural loan liveness, cross-function `-O3` equivalence) depends on
*another* unit's content or on the shape of the call graph itself, the cache
key must include that dependency — and it is easy to ship a cache extension
that adds the *new fact* to the existing per-unit-keyed cache infrastructure
without also extending the *key* to cover its true dependency set, because
the cache machinery itself doesn't change shape, only what's stored in it.

**How to avoid:**
- Treat the call graph itself as a whole-program artifact with its own
  content-bound key derived from the hash set of every function reachable
  from the entry set (or the relevant compilation unit's transitive callee
  closure), and make any interprocedural fact's cache key include that
  call-graph-closure hash, not just the local function's own hash.
- Where any cached fact for function F depends on a fact about a function G
  that F calls (directly or transitively), the cache key for F's entry must
  be a function of G's cache key (or content hash), recursively — mirroring
  the necessity Rust's dependency-graph and ThinLTO import-map machinery
  encode, so that changing G invalidates every cached interprocedural fact
  that transitively depended on it.
  Prefer a design where the interprocedural cache key is *structurally*
  computed from the call graph's closure (so it is correct by construction)
  over one where each new interprocedural fact's author has to remember to
  add every transitive dependency to the key by hand — the latter is exactly
  the shape of bug class Bazel/Buck have repeatedly hit.
- Given the project's existing "fail closed on ambiguity... cache resolves to
  `not_cacheable`" pattern, extend that same discipline explicitly to
  interprocedural facts: if the call-graph-closure hash cannot be computed
  cheaply/soundly for some reason (e.g., during cycle refusal, or when a
  callee is only reachable through an indirect/dynamic path the current
  design doesn't fully resolve), the interprocedural cache entry must resolve
  to `not_cacheable` rather than silently keying on the caller's local hash
  alone.
- Add a differential regression test: cache function A, change only callee B
  (not A's own text), re-request A's cached interprocedural verdict, and
  assert it was invalidated/recomputed — the direct interprocedural analog
  of M001's existing cache-correctness discipline, and a concrete mutation
  to kill per Lesson 1.

**Warning signs:**
- Any interprocedural fact stored under the existing per-unit cache key
  format without an explicit review of what that key covers.
- No test that changes a callee only and asserts the caller's cached
  interprocedural fact is invalidated.
- A "call-graph closure hash" concept that exists in design docs but is not
  actually the literal cache key used for interprocedural entries.

**Phase to address:**
Whichever phase first makes an interprocedural fact cacheable (likely the
interprocedural `-O3`/LTO equivalence phase, since that is the most
expensive-to-recompute interprocedural evidence and therefore the most
tempting to cache). Gate on the callee-changes-invalidates-caller regression
test before any interprocedural fact is added to the cache's cacheable set.

---

### Pitfall 7: Interprocedural diagnostics blame the wrong function, and the bounded cause DAG grows unbounded across function boundaries

**What goes wrong:**
M001's `lang explain` bounded cause DAG and rustfix-style repair
applicability were built and proven against causes that live inside one
function's CFG. Once a rejected program's root cause is "callee G's declared
loan contract conflicts with caller F's loan state at the call site," there
are now (at least) two plausible places to locate the blamed span — the call
site in F, or the declaration in G — and picking the wrong one produces a
diagnostic that sends an AI repair loop to edit the *innocent* function.
This is a well-known cross-function diagnostics problem: Rust's own borrow
checker (NLL/Polonius-era) has spent years of incremental UX work specifically
on making cross-function/cross-closure borrow errors point at the right
span and explain the right relationship (e.g., distinguishing "the argument
you passed doesn't live long enough" from "the function you called requires
a longer lifetime than you declared") — this is a widely-reported, ongoing
area of rustc diagnostics investment, not a solved problem with one known
fix. For an AI-authored/AI-repaired system specifically, a misattributed
interprocedural diagnostic is worse than for a human: a human reviewer
skims the call graph and often self-corrects; an automated repair loop
trusts the blamed span and edits it directly, so a wrong blame produces a
wrong edit with no human-in-the-loop sanity check unless one is
deliberately kept. Separately, the cause DAG's boundedness guarantee was
presumably sized/tested against intraprocedural cause chains; a cross-
function cause chain (loan conflict in F traces back through G's contract,
which itself traces back to H's return value, etc.) has strictly more nodes
per genuine root cause than an intraprocedural one, and "bounded" needs to be
re-verified as a real bound under call-chain depth, not silently inherited
from the intraprocedural measurement.

**Why it happens:**
A diagnostic-generation system built against single-function causes
naturally locates blame at "the span within this CFG where the violated
invariant was last true" — a well-defined, unambiguous concept
intraprocedurally. Once the invariant's violation depends on a fact declared
in a different function (a contract mismatch), there is a genuine modeling
choice about which span to blame, and naive extension of the existing
diagnostic code tends to default to whichever span the code happened to be
looking at when it detected the violation (usually the call site, since
that's where admission-checking runs) — which is not always the more
actionable or more correct location, and is not a decision anyone made on
purpose.

**How to avoid:**
- Treat "which function's declared contract is at fault" as an explicit,
  separately-computed diagnostic decision, not an accident of where the
  check happens to run — mirroring the distinction rustc's diagnostics
  eventually had to make between "the caller passed something insufficient"
  and "the callee demanded something excessive." Where genuinely ambiguous
  (the contract itself may be wrong, or the call site may be wrong), surface
  both candidate spans rather than guessing one, since a repair loop
  presented with two candidates and asked to pick is safer than one
  presented with a single wrong answer stated as fact.
  Since Codename Lang's declared-origin philosophy already favors "declared,
  verified not trusted" contracts (per Pitfall 1's recommendation), the
  diagnostic should default to blaming whichever side violates its own
  *declared* contract, since that is the actionable, agent-editable
  surface — not whichever side the checker's control flow happened to visit
  last.
- Explicitly re-measure the cause DAG's boundedness claim against
  interprocedural cause chains of realistic call-chain depth (not just the
  intraprocedural depth M001 validated), and add a named refusal/truncation
  code (consistent with the existing "fail closed... named refusal codes"
  pattern) for cause chains that would otherwise exceed the bound, rather
  than silently truncating in a way that could drop the real root cause.
- Build a small adversarial diagnostic corpus specifically for cross-function
  blame: cases where the "obvious" call-site blame is wrong and the true fix
  is in the callee's contract, and vice versa; verify `lang explain` and
  `cmd/lang-repair` land on the correct, minimal fix on both, not just one.

**Warning signs:**
- Diagnostic code for interprocedural rejections that only ever emits a span
  inside the calling function, never inside the callee's declared contract
  (or vice versa) — a one-sided default is the tell.
- No adversarial cases in the diagnostic test corpus where the "correct" fix
  location is the less-obvious of the two candidate functions.
- `lang-repair` (or an equivalent agent-facing consumer) applying a
  cross-function repair that doesn't actually resolve the reported violation
  when re-checked — the sharpest possible signal that blame attribution is
  wrong, and worth a dedicated regression test (repair, then re-check,
  assert clean) for every interprocedural diagnostic class.

**Phase to address:**
The phase(s) surfacing interprocedural rejections through `lang explain` and
`cmd/lang-repair`. Gate on the repair-then-re-check regression test and the
cause-DAG boundedness re-measurement before claiming interprocedural
diagnostics are complete.

---

## Technical Debt Patterns

| Shortcut | Immediate Benefit | Long-term Cost | When Acceptable |
|----------|-------------------|-----------------|------------------|
| Infer interprocedural loan facts by re-walking callee CFGs instead of requiring declared contracts | Faster to land, no new declaration syntax needed | Reproduces exactly the class of Rust unsoundness bugs cited in Pitfall 1; couples caller correctness to callee's current implementation rather than its declared contract | Never for the shipped semantic core; acceptable only as a throwaway spike explicitly marked non-production, per the project's own spike precedent |
| Extend only `check` (not `corevalidate`) for a new interprocedural fact in one phase, defer the peer | Smaller single-phase diff, faster to green | Reproduces D-02-03/D-03-01 exactly — a fix that isn't a fix; the milestone's stated goal ("every guarantee M001 established... still holds across function boundaries") is falsified by construction until closed | Never — land both peers in the same plan, per Pitfall 2 |
| Ship interprocedural `-O3`/LTO equivalence without an engineered composition-only negative control | Faster claim of "proven," reuses the existing comparator unchanged | An inert proof tier per Pitfall 3 — indistinguishable from real coverage until an actual interprocedural miscompile occurs in production | Never for the milestone's headline claim; acceptable only as an interim/marked-provisional status while the control is built |
| Reuse the intraprocedural generator with an `OpCall` production rule bolted on, skip a dedicated reachability register for call-graph shape | Fast corpus growth, no new tooling | Systematic under-coverage of exactly the shapes (recursion, diamonds, divergent loan states) most likely to hide real bugs, per Pitfall 5 | Acceptable as a first pass only if paired, in the same phase, with an explicit reachability register naming what it doesn't reach |
| Key interprocedural cache entries on the caller's own content hash alone | No cache-key redesign needed | Silent stale-verdict serving when an unchanged caller's callee changes, per Pitfall 6 | Never; if key redesign is deferred, mark the interprocedural fact `not_cacheable` explicitly rather than caching it incorrectly |

## Integration Gotchas

Cross-boundary integration points specific to this milestone's architecture
(not external services, but internal trust/derivation boundaries — the
project's own analog of "integration"):

| Integration | Common Mistake | Correct Approach |
|--------------|-----------------|-------------------|
| `check` ↔ `corevalidate` on interprocedural facts | Treat validator agreement on the existing corpus as proof of soundness | Require shadow-run diffing on adversarial call-graph shapes (recursion, diamonds, deep chains) before trusting agreement, per Pitfall 2 |
| Interpreter call stack ↔ checker-declared bounded call depth | Assume the interpreter's declared bound is the only thing standing between a deep call chain and a crash | Verify Go's own native recursion in checker/analysis code (not just the interpreter) is converted to explicit stacks wherever call depth is program-controlled, per Pitfall 4 |
| `restrict`/no-alias attribute emission ↔ Clang's inliner | Assume a checker-proven no-alias fact automatically survives inlining the way it does for a single function at `-O0` | Verify with an engineered composition-only negative control, and prefer emitting scoped alias metadata at call sites over trusting the C-level attribute alone, per Pitfall 3 |
| Content-bound cache ↔ call-graph-shaped facts | Add interprocedural facts to the existing per-unit cache format unchanged | Derive interprocedural cache keys from the call-graph closure hash, structurally, per Pitfall 6 |
| `lang explain`/`lang-repair` ↔ cross-function causes | Default diagnostic blame to whichever function the checker's control flow visits when it detects the violation | Explicitly decide caller-vs-callee blame based on which declared contract is violated; verify with repair-then-re-check regression tests, per Pitfall 7 |

## Performance Traps

| Trap | Symptoms | Prevention | When It Breaks |
|------|----------|------------|-----------------|
| Context-sensitive interprocedural analysis without a bounding structure (no IFDS/IDE-style polynomial formulation) | Analysis time grows sharply with call-graph depth/fan-out on real corpus programs, not just adversarial ones | Prefer per-function summaries computed once per SCC (topological order) over unbounded per-call-site context sensitivity, per Pitfall 4 | Becomes visible once call graphs exceed a handful of functions with shared callees (diamond shapes) — likely well within this milestone's own corpus size, not a distant-future concern |
| Exhaustive interprocedural differential testing scaled naively from the intraprocedural exhaustive approach | Test/CI time explodes combinatorially as call-shape axes are crossed | Name axes explicitly and bound the cross product in writing; supplement with weighted random + hand-written cases beyond the bound, per Pitfall 5 | Breaks immediately if "rebuild Phase 3's exhaustive differentials cross-function" is read as "same exhaustiveness strategy, one more dimension" rather than a genuinely new bounding decision |

## Security Mistakes

| Mistake | Risk | Prevention |
|---------|------|------------|
| Treating a checker-proven interprocedural loan/alias fact as sufficient grounds for `restrict`/no-alias codegen without a composition-specific proof | Optimizer exploits a fact that was true of each function in isolation but false of their composition, producing silently miscompiled code with the class of consequences CWE-562-style dangling-reference and aliasing bugs share (memory corruption, UB) — the interprocedural analog of the Rust `noalias`+inlining CVE-adjacent saga (rust-lang/rust#31681/#54878) | Engineered composition-only negative control per Pitfall 3, gating any interprocedural `restrict`/no-alias codegen decision |
| Widening the checker's/`check`'s implicit trust surface at call sites (accepting a callee's re-walked body as sufficient evidence instead of a declared, independently-validated contract) | Reproduces the exact shape of accepted Rust unsoundness issues (Pitfall 1): safe, checker-accepted code that is unsound at runtime, silently, with no `unsafe` marker to flag it for review | Require declared, `corevalidate`-verified contracts at every call boundary; never derive caller-side admissibility from a live re-walk of the callee, per Pitfall 1 |
| Cache serving a stale interprocedural verdict as if it were a fresh proof | A cached "sound" verdict for a caller is served after its callee changed in a way that breaks the interprocedural fact — the cache becomes an unwitting false-claim generator, adjacent in spirit to the project's own accepted `escape:coordinated-source-to-core-false-claim` but *unintended and undeclared* rather than accepted and executable | Call-graph-closure-derived cache keys and the callee-changes-invalidates-caller regression test, per Pitfall 6 |

## "Looks Done But Isn't" Checklist

- [ ] **Interprocedural loan liveness:** Often missing the composition-depth-≥2
      adversarial corpus (Pitfall 1) — verify the corpus contains call chains
      of depth ≥2 with divergent loan states across callers, not just
      single-call cases.
- [ ] **`check`/`corevalidate` agreement on interprocedural facts:** Often
      missing a shadow-run diff specifically on adversarial (not just
      in-corpus) call-graph shapes (Pitfall 2) — verify a shadow-run exists
      and was run on recursion, diamonds, and deep chains, not just the
      milestone's own generated corpus.
- [ ] **Interprocedural `-O3`/LTO equivalence:** Often missing an engineered
      composition-only negative control that fails red before the fix, the
      way `false_no_alias` did intraprocedurally (Pitfall 3) — verify such a
      control exists and has been shown to fail.
- [ ] **Call-graph cycle refusal:** Often missing indirect/mutual-recursion
      cases and cycles routed through `Result` alternative matching, testing
      only direct self-recursion (Pitfall 4) — verify the refusal corpus
      covers indirect cycles.
- [ ] **Interprocedural test corpus:** Often missing a dedicated call-graph-
      shape reachability register distinct from the existing loan-endpoint
      register (Pitfall 5) — verify one exists and names shapes not reached.
- [ ] **HDD reducer on multi-function programs:** Often missing a
      re-verification that reduced output still reproduces the *same*
      property as the unreduced input (Pitfall 5) — verify this regression
      test exists.
- [ ] **Interprocedural evidence caching:** Often missing a callee-changes-
      invalidates-caller regression test (Pitfall 6) — verify it exists
      before any interprocedural fact is marked cacheable.
- [ ] **Interprocedural diagnostics:** Often missing adversarial cases where
      the "obvious" blamed span is the wrong one (Pitfall 7) — verify
      `lang-repair`'s repair-then-re-check regression test covers
      cross-function misattribution cases specifically.
- [ ] **D-03-02 closure:** Often "closed" by extending the intraprocedural
      fix pattern rather than genuinely re-deriving the interprocedural
      contract (Pitfall 1) — verify the fix is accompanied by its own
      negative-control corpus, not just a scope extension of the existing one.

## Recovery Strategies

| Pitfall | Recovery Cost | Recovery Steps |
|---------|----------------|------------------|
| Cross-function unsoundness discovered post-ship (Pitfall 1) | HIGH | Treat as a soundness incident: quarantine the affected `OpCall` shape (mirroring the existing quarantine pattern for accepted residual limitations), add the counterexample as a permanent negative control, re-derive the contract rather than patching around the symptom |
| `check`/`corevalidate` drift discovered late (Pitfall 2) | MEDIUM-HIGH | Shadow-run both layers over the full adversarial call-graph corpus immediately, characterize every divergence before fixing any (fixing the first found tends to mask siblings — "fixing one of N peers is not fixing the item"), fix all divergent sites in one coordinated plan |
| Inert `-O3`/LTO interprocedural proof tier discovered late (Pitfall 3) | MEDIUM | Build the missing composition-only negative control retroactively, run it against the shipped comparator, treat any pass as evidence the tier needs strengthening (not evidence of soundness) before re-claiming the equivalence requirement is met |
| Fixpoint non-termination/stack overflow found in production input (Pitfall 4) | LOW-MEDIUM | Add the triggering call-graph shape to the SCC/cycle-refusal test corpus, convert the offending traversal to an explicit worklist if it was native recursion, add the fixpoint-iteration-count fail-closed guard |
| Generator/reducer under-coverage discovered via an out-of-corpus hand-written program (Pitfall 5) | LOW | Add the discovered shape to the reachability register as a newly-reached row (mirroring D-04-21's discipline), extend generator weights, add the case to the exhaustive-bound corpus if it falls within the stated axes |
| Stale interprocedural cache verdict served (Pitfall 6) | MEDIUM | Treat as a cache-correctness incident (same severity class as a false claim escaping): audit the cache key's dependency closure computation, add the missing edge, add the regression test that would have caught it, and invalidate/rebuild any cache entries that could have served the stale verdict |
| Misattributed interprocedural diagnostic caused a bad automated repair (Pitfall 7) | MEDIUM | Add the case to the diagnostic adversarial corpus, correct the blame-decision logic to check which declared contract was actually violated, add a repair-then-re-check regression test for the specific misattribution shape |

## Pitfall-to-Phase Mapping

| Pitfall | Prevention Phase | Verification |
|---------|-------------------|----------------|
| 1. Cross-function borrow/lifetime unsoundness | Phase introducing declared callee contracts + interprocedural loan liveness in `check` | Composition-depth-≥2 divergent-loan-state adversarial corpus passes; D-03-02 closed with its own negative-control corpus, not a scope extension |
| 2. `check`/`corevalidate` drift | Same phase(s) as Pitfall 1, both layers landed together | Shadow-run diff over recursion/diamond/deep-chain call graphs shows zero divergence before either layer ships |
| 3. Inert interprocedural `-O3`/LTO proof tier | Phase claiming interprocedural `-O3`/LTO equivalence | Engineered composition-only negative control exists and is shown to fail red pre-fix, mirroring `false_no_alias` |
| 4. Fixpoint non-termination / SCC mishandling / stack blowup | Phase introducing call-graph construction + cycle refusal; phase introducing interprocedural liveness fixpoint | Indirect-cycle and pathological-depth-but-acyclic corpus passes; fixpoint has an explicit iteration-count fail-closed guard |
| 5. Generator under-coverage / reducer correctness on multi-function programs | Phase rebuilding Phase 3's exhaustive differentials cross-function; any phase extending the HDD reducer | Call-graph-shape reachability register exists with named unreached rows; reducer output re-verified to reproduce the same property as input |
| 6. Cache invalidation misses whole-program facts | First phase making any interprocedural fact cacheable (likely the `-O3`/LTO equivalence phase) | Callee-changes-invalidates-caller regression test passes before the fact is marked cacheable |
| 7. Diagnostics blame the wrong function | Phase surfacing interprocedural rejections through `lang explain`/`cmd/lang-repair` | Repair-then-re-check regression test covers cases where the non-obvious function is the correct fix location |

## Sources

**Verified from primary source (issue trackers, specs, papers):**
- rust-lang/rust#84366 — `'static` closures/futures with non-`'static` return
  type unsound (`I-unsound` label): https://github.com/rust-lang/rust/issues/84366
- rust-lang/rust#106431 — implied bounds treat contravariant lifetimes as
  covariant: https://github.com/rust-lang/rust/issues/106431
- rust-lang/rust#71550 — unsoundness from variance of trait objects wrt
  associated types: https://github.com/rust-lang/rust/issues/71550
- rust-lang/rust#123241 — async closures reject valid captures under
  invariant lifetimes: https://github.com/rust-lang/rust/issues/123241
- rust-lang/rust#31681 — undo `&mut` `noalias` workaround pending LLVM fix:
  https://github.com/rust-lang/rust/issues/31681
- rust-lang/rust#54878 — `-Zmutable-noalias` default tracking issue:
  https://github.com/rust-lang/rust/issues/54878
- llvm/llvm-project#122537 — TBAA/GVN miscompile from incorrect pointer→undef
  replacement, triggered after inlining: https://github.com/llvm/llvm-project/issues/122537
- llvm/llvm-project#98978 — BasicAA incorrect alias analysis causing
  SLP-vectorizer miscompile: https://github.com/llvm/llvm-project/issues/98978
- Swift SR-8546 / swiftlang/swift#51064 — exclusivity gap: nested function
  with implicitly captured inout parameter, assert-build-only diagnosis:
  https://github.com/swiftlang/swift/issues/51064
- Swift 5 Exclusivity Enforcement (official): static diagnostics do not cover
  escaping closures/class properties/globals, dynamic checks required:
  https://www.swift.org/blog/swift-5-exclusivity/
- WG14 N3005 — "A Provenance-aware Memory Object Model for C" (Gustedt,
  Sewell, Memarian, Gomes, Uecker), ISO/IEC TS 6010:2023:
  https://www.open-std.org/jtc1/sc22/wg14/www/docs/n3005.pdf
- CompCert structure — pass-by-pass choice between Coq-proved transformation
  and translation validation (e.g., SSA construction): https://www.absint.com/compcert/structure.htm ;
  Leroy, "CompCert – A Formally Verified Optimizing Compiler" (ERTS 2016):
  https://xavierleroy.org/publi/erts2016_compcert.pdf
- seL4 binary verification layer and trusted-compiler-assumption limits;
  bitfield-generation distrust of GCC: https://sel4.systems/Research/pdfs/comprehensive-formal-verification-os-microkernel.pdf ,
  https://cacm.acm.org/research/sel4-formal-verification-of-an-operating-system-kernel/
- CWE-562 — Return of Stack Variable Address: https://cwe.mitre.org/data/definitions/562.html

**Widely-reported / secondary but consistent across sources:**
- Rust ThinLTO + incremental compilation import-map re-invalidation
  mechanism (rustc dev guide / RFC 1298 / PR #53673) as precedent for
  "changed transitive dependency, not just own content, must invalidate
  cache": https://rust-lang.github.io/rfcs/1298-incremental-compilation.html ,
  https://github.com/rust-lang/rust/pull/53673
- IFDS/IDE interprocedural dataflow framework as the standard answer to
  naive context-sensitive analysis blowup — widely cited in program-analysis
  literature (Reps, Horwitz, Sagiv 1995, "Precise Interprocedural Dataflow
  Analysis via Graph Reachability"); not independently re-verified this
  session, flagged as inference from established literature rather than a
  freshly fetched primary source.

**This project's own primary sources (authoritative for project-specific
claims):**
- `~/projects/ai-lang/.planning/PROJECT.md`
- `~/projects/ai-lang/.planning/RETROSPECTIVE.md`

**My inference (not externally sourced):** The mapping in Section
"Pitfall-to-Phase Mapping" and all "warning signs" tied to this project's
specific architecture (`check`/`corevalidate` peer structure, the evidence
cache's `not_cacheable` fail-closed pattern, the HDD reducer, `lang explain`'s
bounded cause DAG) are derived by applying the externally-verified failure
classes above to this project's documented M001 architecture and its own
recorded lessons — these are reasoned extrapolations, not independently
verified external precedents for *this* codebase.

---
*Pitfalls research for: Codename Lang M002 — Interprocedural Semantic Spine*
*Researched: 2026-09-08*
