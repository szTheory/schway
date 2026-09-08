# Feature Research: Interprocedural Semantic Spine (M002)

**Domain:** Systems programming language — affine ownership, borrow checking,
call semantics, sum types, cross-tier equivalence proof
**Researched:** 2026-09-08
**Confidence:** MEDIUM-HIGH (primary sources for Rust/Ada/SPARK/LLVM tooling are
strong; Val/Hylo and Austral are pre-1.0 projects whose public docs are the
primary source, so confidence there is MEDIUM; anti-feature judgments are the
researcher's inference calibrated against M001's decisions, marked as such)

---

## 0. Reading This Document

Each of the five target-feature areas below is analyzed through five lenses:
**language designer** (what must the type system express), **compiler
implementer** (what must the checker/codegen actually do), **verification
engineer** (what must be independently provable), **AI agent author** (what
must be inferable/checkable without full-program context so an agent can move
fast), **human reviewer** (what must be legible on read-through). Findings are
tagged for provenance:

- **(V)** verified from a primary source (spec, RFC, compiler source, paper) — URL given
- **(C)** widely reported convention across multiple implementations, no single primary source
- **(I)** researcher inference, calibrated against M001's existing architecture

---

## 1. Function Calls Under Affine/Linear Ownership

### 1.1 Call-site move vs. borrow semantics — how real languages express it

**Rust** has no first-class "convention" keyword at call sites; ownership
transfer is inferred from the parameter's declared type: `T` moves (unless
`T: Copy`), `&T` shares, `&mut T` borrows exclusively. The caller's obligation
(move vs. borrow) is entirely a function of the *callee's declared signature*,
which is why Rust's borrow checker can treat a call as a black box governed by
its signature rather than needing to inline the callee. **(C)**, consistent
with the standard NLL design (see §2).

**Val/Hylo** makes the convention explicit and mandatory in the signature:
four parameter-passing conventions — `let` (borrowed, immutable, cannot
escape the call), `inout` (exclusive, in-place mutation, caller marks `&` at
the call site), `sink` (ownership transfer, callee may let it escape), `set`
(callee initializes an uninitialized binding). **(V)**
https://docs.hylo-lang.org/language-tour/functions-and-methods (accessed
2026-09-08). This is the single most directly transferable finding for
Codename Lang: Val/Hylo proves a *purely local, no-whole-program-inference*
convention system is workable — every convention is legible from the
signature alone, which is exactly the "declared, not inferred, at boundaries"
posture Codename Lang already committed to for public origins (see PROJECT.md
"Infer local borrow ends at CFG last use; declare public origins").

**Austral** takes the linear-types-as-capabilities route: a linear value is
*consumed* by being passed to a function (ownership transfer is the only
option unless the function explicitly returns it back, a discipline called
"linearity threading"), and non-linear (free) types pass by value/copy.
**(V)** https://austral-lang.org/tutorial/linear-types (accessed 2026-09-08).
Austral has no separate "borrow" call convention in its core linear-type
system — borrowing is a derived, checker-verified feature layered on top,
which keeps the calling convention itself trivial to typecheck at a call site
(cost: no interprocedural aliasing at all unless the borrow layer is used).

**C++** has no ownership type system; move-vs-copy at a call site is
determined by overload resolution on rvalue/lvalue reference qualifiers
(`T&&` vs `T&` vs `const T&`), decided per call site, not declared as part of
a single canonical signature contract — which is precisely why C++ needs
separate move-constructor/copy-constructor overloads and why "moved-from but
not reset" is a perennial correctness pitfall. **(C)** — this is the anti-
pattern Codename Lang's M001 affine-values design (single canonical ability
per type, not overload-selected) already avoids; M002 must preserve that: one
call, one ownership discipline per parameter position, not overload-resolved.

**Swift** ownership annotations (`borrowing`, `consuming`, `inout`) were
added post-hoc in Swift 5.9 to an ARC-based language and are opt-in
performance annotations, not the load-bearing safety mechanism (ARC + COW
still make owning wrong non-fatal, just slower). Not a strong precedent for
Codename Lang since Codename Lang's ownership *is* the safety mechanism.
**(C)**.

**Zig** has no ownership types at all; the language relies entirely on
convention and `defer`. Not applicable as ownership precedent, but relevant
as an anti-pattern: Zig's calling convention decisions are the *reason* Zig
projects need external tools (e.g. GeneralPurposeAllocator's leak detector)
to catch what a type system would catch statically. **(I)**.

### 1.2 Table stakes / differentiators / anti-features for §1

| Item | Category | Why | Complexity | Depends on |
|---|---|---|---|---|
| Per-parameter ownership convention derived from callee's declared type (move for owned params, borrow for reference params) with no overload-based dispatch | Table stakes | Every serious ownership language (Rust, Val/Hylo, Austral) makes the callee signature the sole source of truth at the call site; anything else (C++ overload resolution) reintroduces the "moved-from but still readable" class of bug M001's affine `drop`/`copy`/`share`/`send` abilities exist to prevent | MEDIUM — this is "just" applying M001's existing per-type abilities to argument positions in `OpCall`, but doing it at all 6 dispatch sites named in PROJECT.md is where the real work is | M001's ability system (`copy`/`drop`/`share`/`send`/`escape`), CFG per-function |
| Declared, not inferred, ownership contract at the function boundary (parameter and return conventions are part of the signature, checkable without seeing the callee body) | Table stakes | This is the entire reason NLL/Polonius can treat calls as black boxes (§2); Val/Hylo proves this scales to a from-scratch design | MEDIUM | Extends M001's existing "declare public origins" precedent from borrow origins to the full call contract |
| "callable ⊆ publishable" gate: the type/ability surface a function may accept or return through its *signature* must be a subset of what that function is permitted to expose across a compilation-unit boundary | Differentiator | Named directly in PROJECT.md (D-04-03); this is not a common named gate in any of the surveyed languages under that name, but it is structurally what Rust's `pub` visibility + auto-trait leakage rules and Austral's module export lists both approximate piecemeal — Codename Lang's version is likely to be a first-class, uniformly-enforced typing judgment rather than a visibility-modifier side effect. This is exactly where Codename Lang can be distinctively legible: an AI agent or reviewer can ask "is this callable?" as one static predicate instead of chasing `pub(crate)` visibility rules through modules | HIGH — needs precise definition of "publishable" as a set (which abilities, which origin declarations) and a checker pass that rejects any `OpCall` target whose signature exceeds it | M001's typed core, ability system, `originvalidate`'s public-origin machinery |
| Call-site override of the callee's declared convention (e.g., forcing a copy instead of a move, or vice versa, via caller-side annotation) | Anti-feature | C++'s move/copy overload-resolution ambiguity is the cautionary tale: caller-chosen conventions bifurcate the meaning of a call, making the callee signature no longer the single source of truth, defeating the "callable is a black box governed by its signature" property that makes NLL/Val's approach tractable | — | — |
| Implicit reference-counting fallback for arguments that don't cleanly move or borrow (Swift ARC-style "just retain it") | Anti-feature | Contradicts M001's stated "no mandatory global tracing heap or service runtime" constraint outright; would silently reintroduce a managed-heap-shaped feature through the argument-passing back door | — | — |

---

## 2. Interprocedural Borrow/Loan Liveness

### 2.1 How Rust actually does this — signatures as summaries, not whole-program analysis

Rust's non-lexical lifetimes (NLL) and its successor Polonius are both
**per-function** analyses. A function call is not inlined or re-analyzed
against the callee's body; the callee's *signature* — its declared lifetime
parameters and how they relate its borrowed inputs to its borrowed outputs —
is the entire interprocedural contract the caller's borrow checker consults.
**(C)**, corroborated by rustc's own architecture description: NLL computes
region/lifetime inference and loan-liveness per function body, using elided
or explicit lifetime parameters on other functions' signatures as opaque
facts rather than re-deriving them
(https://doc.rust-lang.org/nightly/nightly-rustc/src/rustc_borrowck/nll.rs.html,
accessed 2026-09-08). Polonius reformulates this as a set of Datalog-style
input facts (loans, origins, subset relations) derived per function and
consumed by a solver, but the boundary is unchanged: origins/lifetimes on a
signature are the summary a caller trusts without inlining
(https://github.com/rust-lang/polonius, accessed 2026-09-08; the design
explicitly frames "origins as sets of loans" as the mechanism, which is a
strict formalization but not a change in the whole-program-vs-modular
boundary). This is a direct match to what PROJECT.md already committed to:
"declare public origins" for the interprocedural half, verified independently
by `originvalidate`.

**Two-phase borrows** matter here because calls are exactly where they bite:
a mutable borrow taken to pass as an `&mut` argument (e.g. `x.push(...)`
where an argument expression also reads `x`) is split into a *reservation*
point (borrow exists, but only conflicts with other *mutable* uses) and an
*activation* point (borrow becomes fully exclusive, at the point the callee
actually begins executing with it) — implemented as two separate dataflow
analyses over the MIR, one computing reservations and one computing
activation-following uses. **(V)**
https://rustc-dev-guide.rust-lang.org/borrow_check/two_phase_borrows.html
(accessed 2026-09-08). Without this, common call patterns like
`vec.push(vec.len())` are rejected even though they are safe — a widely-cited
Rust ergonomics gap that was specifically about *call sites*. This is directly
relevant to `OpCall`: M001's CFG-last-use loan liveness will need an analogous
reservation/activation split at call sites where an argument expression
itself reads through an exclusive loan being formed for another argument in
the same call, or Codename Lang will inherit Rust's pre-2018 ergonomic
failure mode.

**What breaks when you infer instead of declare across boundaries**: Rust
tried whole-program region inference in its earliest (pre-1.0) borrow checker
design and abandoned it for exactly the scalability/modularity reason
Codename Lang's PROJECT.md already names — per-function declared contracts
let each function be checked once, independent of its callers, and let
separate compilation and incremental recompilation work. **(I)**, inferred
from the general and well-documented modular-typechecking argument, not a
specific citation found. The concrete failure modes of full inference across
function boundaries: (a) checking a function requires knowing every caller's
context, breaking separate compilation; (b) a change to a leaf function's
body (not its signature) can change what's accepted at every call site
transitively, breaking incremental caching — this directly threatens M001's
"content-bound evidence manifests" and "verdict-free cache" if not guarded
against; (c) error messages point at call sites arbitrarily far from the
actual constraint violation, which is a poor fit for the agent-facing
`lang explain` cause-DAG design already shipped.

### 2.2 Table stakes / differentiators / anti-features for §2

| Item | Category | Why | Complexity | Depends on |
|---|---|---|---|---|
| Declared origin parameters/relations on every publishable function signature that borrows or returns a borrow, checked interprocedurally by treating the *signature* (not the body) as the caller's fact | Table stakes | This is the entire Rust NLL/Polonius precedent and matches M001's existing intraprocedural "declare public origins" design exactly — extending it to calls is not a new idea, just a new consumer (`OpCall`) of an existing mechanism | HIGH — closes D-03-02 per PROJECT.md; must land in both `check` and `corevalidate` (independent re-derivation, per M001's established architectural rule) | M001's public-origin declaration and `originvalidate`; CFG-last-use loan liveness fixpoint |
| Reservation/activation split (two-phase-borrow-equivalent) for loans formed to satisfy one argument position while another argument expression in the same call reads the same binding | Table stakes if Codename Lang wants ergonomic multi-argument calls involving self-referential expressions; otherwise a deliberate, documented restriction | Without it, straightforward patterns (`f(&mut x, x.field)`-shaped calls) are rejected even when safe, which is exactly the ergonomic failure Rust fixed post-1.0 at real cost; better to decide this deliberately in M002 than patch it in a later milestone under the "coexisting laws" anti-pattern the retrospective calls out (Key Lesson 2) | MEDIUM-HIGH — a second dataflow pass, or an explicit deferred-loan-formation rule at call lowering | M001's `loanLivenessFixpoint` (extend, do not duplicate — retrospective Key Lesson 2 applies directly) |
| Independent re-derivation of interprocedural loan liveness in `corevalidate`, separate algorithm from `check`'s, per M001's established peer architecture | Table stakes for this project specifically | M001's core architectural rule ("independent re-derivation at every trust crossing... the producer's claim is never the consumer's evidence") — a PROJECT.md requirement, not optional here | HIGH — retrospective flags this exact peer-pair pattern (D-02-03/D-03-01) as the single most expensive kind of mistake in M001: fixing one of two peers is not fixing the item | `corevalidate`'s existing reachability-closure re-derivation strategy |
| Signature-as-cache-key: a callee's *body* changing without its declared signature changing must not force re-verification of unrelated callers | Differentiator | Directly protects M001's "content-bound evidence manifests... verdict-free cache" and "changed-risk lane selection" investments from being undermined by interprocedural analysis; this is the concrete mechanism that keeps M002 additive rather than a scalability regression on the fast feedback loop that is this project's stated core value | MEDIUM — mostly a cache-key/dependency-tracking design question layered on the existing evidence manifest system | `internal/compiler/cache`, `risk_lanes.json` changed-risk lane selection |
| Whole-program / cross-module inference of borrow relationships (inferring what a function's origin contract "must have been" from its body, propagated to all callers, no declaration required) | Anti-feature | This is precisely the design Rust abandoned; breaks separate compilation, breaks incremental caching (directly threatens the shipped evidence-manifest/cache architecture), produces call-site-displaced error messages that are a poor fit for the cause-DAG diagnostic design already shipped in Phase 6 | — | — |
| A second, more "precise" loan-liveness law introduced for the interprocedural case that coexists with the intraprocedural `loanLivenessFixpoint` rather than being the same law applied at a new site | Anti-feature (self-imposed, from this project's own retrospective) | Retrospective Key Lesson 2, verbatim: "When a phase introduces a better derivation, retiring the old one is part of that phase, not a follow-up. Two coexisting laws is a defect with a delayed fuse." M001 already paid this cost once (03-03 through 05-02, ~2 phases carrying a duplicate law) | — | — |

---

## 3. Call-Graph Construction

### 3.1 Direct/indirect calls, recursion, cycle handling

For a language at Codename Lang's stated maturity (no reflection, no macros,
no full generics/dynamic dispatch yet per PROJECT.md's Out of Scope list),
call-graph construction is close to fully static: every `OpCall` target is
either a direct, statically-resolved function reference, or — if the
language has any first-class function values at this stage — an indirect
call through a value whose possible targets require either (a) a closed
enumeration derivable from the type system, or (b) conservative refusal.
Given PROJECT.md's explicit exclusion of "reflection, macros," and the
absence of any stated first-class-function-value feature in the M002 charter,
the realistic scope is **direct calls only** for M002, which substantially
simplifies call-graph construction relative to Rust (trait objects, `dyn`)
or C++ (virtual dispatch) — this should be stated as an explicit, deliberate
scope boundary in the roadmap, not discovered mid-phase. **(I)**, based on
reading PROJECT.md's scope list; flag this as a question for the roadmap
phase, since it is the single biggest complexity lever in this whole
capability area.

**Recursion / SCC handling — what languages actually refuse:**
- **MISRA C Rule 17.2** (formerly Rule 70) bans direct or indirect recursion
  outright in required guidelines for safety-critical C, precisely because
  "the presence of (mutually) recursive functions considerably complicates
  the task of bounding the maximum stack size needed to run the code," and
  the rule is checkable as a decidable approximation via call-graph cycle
  detection (with a caveat that calls through function pointers cannot be
  soundly excluded without flagging them separately). **(V)**
  https://arxiv.org/pdf/2212.13933 "Coding Guidelines and Undecidability"
  (accessed 2026-09-08); cross-referenced by
  https://pvs-studio.com/en/docs/warnings/v2565/ (accessed 2026-09-08).
- **Ada/SPARK** does not ban recursion outright but requires — for the
  "Platinum" (highest) SPARK assurance level — that recursion be either
  eliminated or *provably bounded* (a decreasing metric proof obligation,
  similar to termination proofs in other verification systems), and provides
  `GNATstack` as a dedicated whole-program static stack-usage analyzer that
  computes worst-case stack consumption per subprogram and flags any call
  graph it cannot bound (e.g., unbounded recursion, calls through unresolved
  function pointers) as a hard failure requiring manual annotation or
  restructuring. **(V)**
  https://www.adacore.com/blog/from-ada-to-platinum-spark-a-case-study-for-reusable-bounded-stacks
  (accessed 2026-09-08) — concrete worked figures given (e.g., 5648 bytes
  worst-case stack for a specific call under 32-bit RISC-V).
- **seL4** and comparable formally-verified kernels are widely reported to
  avoid unbounded recursion entirely in the verified core and to size stacks
  from static call-graph-depth analysis rather than runtime guard pages
  alone, consistent with the general embedded/safety-critical pattern above
  — **(C)**, general knowledge, no specific primary citation retrieved in
  this pass; treat as convention-level evidence, not verified.
- **Embedded Rust** (`no_std`, no heap) commonly uses external tools
  (`cargo-call-stack`) that build a static call graph from LLVM IR and reject
  or flag any cycle (recursion) or indirect call it cannot resolve, because
  the guarantee is only sound for an acyclic, fully-resolved graph — this is
  convention-level community tooling, not a language guarantee. **(C)**.

**What this means for Codename Lang**: PROJECT.md's M002 charter already
says "cycle refusal" for the call graph and "bounded interpreter call stack"
— this puts Codename Lang in the MISRA-C/embedded-Rust camp (refuse
recursion entirely, rather than the Ada/SPARK camp of "prove a bound"), which
is the *simpler, more conservative* choice and is consistent with a young
kernel-oracle project that has explicitly deferred "full generics" and other
scope. This should be named as a deliberate choice with the Ada/SPARK
alternative (bounded-not-refused recursion) recorded as a plausible future
differentiator, not a gap.

### 3.2 Table stakes / differentiators / anti-features for §3

| Item | Category | Why | Complexity | Depends on |
|---|---|---|---|---|
| Static, whole-corpus call graph built from all `OpCall` sites, direct calls only (no first-class function values/dynamic dispatch in scope for M002) | Table stakes | Required before any interprocedural liveness, equivalence, or stack-bound work is possible; PROJECT.md names this explicitly | MEDIUM given direct-calls-only scope; would be HIGH if indirect calls are silently in scope — this must be pinned down as a roadmap decision | New subsystem; consumes M001's per-function CFGs |
| Cycle detection (SCC computation) with hard refusal of any detected recursion (direct or mutual), matching MISRA-C Rule 17.2 and embedded-Rust convention | Table stakes, per PROJECT.md's explicit "cycle refusal" charter | Matches the project's existing "fail closed on ambiguity" pattern (retrospective: "the lane selector refuses drift; graph walks carry visited-set guards and named refusal codes rather than hanging") — this is not a new pattern for this codebase, it is applying an established one to a new graph | MEDIUM — standard Tarjan/Kosaraju SCC, well-trodden; the discipline is making refusal a *named, diagnosable* verdict (per the diagnostic-as-API architecture) rather than a bare compiler error | Graph-walk visited-set discipline already established (per retrospective) |
| Bounded interpreter call stack with a fixed, documented depth ceiling, enforced independent of host OS stack limits, producing a diagnosable "stack bound exceeded" verdict rather than a host segfault | Table stakes | PROJECT.md names this explicitly; a language whose interpreter oracle can silently SIGSEGV instead of producing a structured diagnostic breaks the "stable, bounded, machine-readable cause graphs" requirement already validated in M001 | LOW-MEDIUM — mechanical once the call graph exists (no recursion means depth is graph-diameter-bounded, not runtime-input-dependent) | `lang.diagnostic` API, deterministic interpreter oracle |
| Statically provable stack bound per function (report worst-case stack bytes/frames per callable, GNATstack-style), consumable by the evidence manifest | Differentiator | Since recursion is refused outright (not merely bounded), the call graph is a DAG and a per-function/per-callable worst-case stack bound is a *cheap, exact* computation — Codename Lang could report this as a first-class evidence-manifest fact essentially for free, which is a real distinguishing capability relative to languages that only prove this at the "safety-critical subset" tier (Ada/SPARK Platinum) rather than by default | MEDIUM — arithmetic over the already-acyclic call graph plus per-function frame-size facts `cgen` likely already computes for C emission | `cgen`'s C17 emission (frame layout), call graph, evidence manifest |
| Allowing *bounded* recursion via a decreasing-metric proof obligation (Ada/SPARK Platinum style) instead of refusing all recursion | Differentiator, deliberately deferred | Legitimate and more expressive than blanket refusal, but adds a termination-proof obligation system this project has no scaffolding for yet (no proof-obligation infrastructure named anywhere in M001's shipped surface) — correctly out of scope for M002, worth flagging for a later milestone once the kernel oracle is trustworthy, per this project's own stated risk posture | HIGH | Would need a new proof-obligation subsystem; not a near-term item |
| Indirect/dynamic dispatch calls (function pointers, closures, trait-object-style dispatch) in the M002 call graph | Anti-feature for M002 specifically | PROJECT.md's Out of Scope list already excludes "full generics, subtyping, variance" and there is no first-class function-value feature named in the M002 charter; admitting indirect calls now would force either unsound call-graph approximation or premature design of a dispatch mechanism, both of which match the exact failure pattern the retrospective calls out ("a green test whose reachable input space omitted the hard case") if done partially | — | — |
| Silent depth-based stack growth (dynamically resizing interpreter stack with no fixed, documented ceiling) | Anti-feature | Defeats the entire purpose of a "bounded interpreter call stack" as a deterministic, evidence-bearing guarantee; matches the general pattern this project explicitly avoids (unbounded/undiagnosed failure modes) | — | — |

---

## 4. `Result`/Sum Types With Payloads

### 4.1 Representation, exhaustiveness, ownership interaction, propagation

**Representation.** The dominant real-world representation for a
payload-carrying sum type is a tagged union: a discriminant plus the union of
each variant's payload layout, sized to the largest variant (Rust enums,
Swift enums, C++ `std::variant`, Ada discriminated records all converge on
this). **(C)**, broadly convergent design across all four ecosystems.
Rust additionally performs **niche-filling optimization**: when a variant's
payload type has an in-memory bit pattern that is provably unreachable for a
valid value of that type (e.g., a `NonZeroUsize` can never be `0`, a
reference can never be null), the compiler reuses that invalid bit pattern as
the discriminant instead of adding a separate tag byte/word — this is why
`Option<&T>` and `Option<NonZeroU8>` are the same size as `&T`/`u8`
respectively rather than one word larger. **(V)**
https://rust-lang.github.io/rfcs/2195-really-tagged-unions.html and
https://medium.com/@zty0826/the-optimization-of-enum-layout-in-rust-cc0e0f34df55
(accessed 2026-09-08) — the compiler computes both a naive tagged layout and
a niche-filling layout and picks the smaller sound one. This directly
interacts with M001's readable-C17 codegen goal: a `Result<T, E>` whose `Ok`
payload type has a compiler-provable niche (Codename Lang's ownership system
already proves non-null/non-zero facts for some types via its ability
system) can be emitted as a plain struct with no explicit tag field in the
generated C, which is both a real size/perf win and a legibility win for
human review of the emitted C (fewer synthetic tag fields cluttering the
readable-C17 output that M001 explicitly optimized for readability).

**RVO/sret and return-by-value payloads.** When a payload-carrying return
type is larger than a register (or generally, per the platform ABI), C/C++
and Rust's C-ABI-compatible codegen both lower it to the caller passing a
hidden pointer to pre-allocated storage (`sret` in LLVM terms) that the
callee writes into directly, avoiding a copy. **(C)**, standard ABI
convention documented across LLVM/C ABI references; not specific to any one
citation retrieved this pass, but foundational and uncontested. This is a
direct consequence of M001's "readable C17" commitment: `Result` payload
returns must lower to either register-return (small payload, no niche
issues) or explicit out-parameter (`sret`-equivalent) in the emitted C, and
the choice must be stable and derivable independent of optimization level
(a claim the five-axis interpreter/`-O0`/`-O3` comparator can directly
verify once `Result` exists).

**Exhaustive matching.** Rust, Swift, and OCaml/ML-family all require match
exhaustiveness as a compile error, not a lint, and this is table stakes for
any language claiming sum types are a first-class safety feature — a
non-exhaustive match on a `Result` is precisely the class of silent
error-swallowing bug that untyped/unchecked error handling (C's `errno`,
unchecked exceptions) is infamous for. **(C)**, universal convention.

**Ownership interaction — moving out of a matched payload.** This is the
genuinely hard part and the one most directly gated by M001's existing
affine-ownership machinery. In Rust, `match`-ing on an owned `Result<T, E>`
and binding a payload variable in an arm *moves* that payload out of the
matched value, which is why the original scrutinee cannot be used again
afterward (whole-value affine move applies transitively through the match).
Binding by reference (`match &result { Ok(ref x) => ... }` or ergonomic
default-binding-modes since Rust 2018) instead borrows the payload without
consuming the enclosing value. **(C)**, standard Rust semantics, uncited
here as it is textbook-level and consistent across every Rust reference.
For Codename Lang specifically, this is not a new primitive — it is the
existing affine `move`/`share`/`drop` ability system applied to a *matched
payload binding site* rather than a plain `let` binding, meaning the checker
work is "extend the existing per-binding ability derivation to match arms,"
not "invent a new ownership discipline for sum types." **(I)**, inference
from M001's documented architecture (per-value abilities already derived
independently through aggregates and generics per the Key Decisions table).

**Error propagation operator.** Rust's `?` operator is the standard
precedent: applied to a `Result<T, E>`, it evaluates to `T` on `Ok`,
short-circuits and returns from the enclosing function on `Err`, and
performs an implicit `From`-based error-type conversion on the propagated
error via the `Try` trait mechanism. **(V)**
https://github.com/rust-lang/rfcs/blob/master/text/1859-try-trait.md and
https://doc.rust-lang.org/reference/expressions/operator-expr.html (accessed
2026-09-08). Two design questions this raises for Codename Lang, both
directly relevant to "affine drop obligations": (a) does propagation of the
`Err` arm correctly run drop obligations for every other live owned binding
in the enclosing function's scope at the early-return point (this is exactly
the kind of nonlocal-exit path M001's own debt register already names —
D-04-31's "two nonlocal-exit blind spots" — so a `Result` propagation
operator is a *third* nonlocal-exit shape that must reuse, not duplicate,
whatever cleanup-ordering machinery resource acquisition's three-stage
cleanup already established); (b) does the language need `From`-style
implicit error conversion at all, or is that exactly the kind of "ambient"
convenience that this project's constitution ("no generic `untaint` or
ambient authority escape") would reject as too magic for a calm canonical
form meant to be read, not typed quickly.

### 4.2 Table stakes / differentiators / anti-features for §4

| Item | Category | Why | Complexity | Depends on |
|---|---|---|---|---|
| Tagged-union representation for `Result`/sum types with a well-defined, documented layout algorithm (tag + largest-variant-sized payload storage) | Table stakes | Universal across every surveyed language; the emitted C17 must have one deterministic, explainable layout rule | MEDIUM | `cgen`'s existing struct/aggregate lowering |
| Exhaustive match-arm checking as a hard compile error (not a lint), verified independently in `corevalidate` per M001's peer-re-derivation rule | Table stakes | Universal convention (Rust/Swift/OCaml); non-exhaustive matching is the single most common source of silent error-swallowing in languages without it (unchecked C `errno`, uncaught exceptions) | MEDIUM | Existing exhaustiveness-adjacent machinery if any from M001's typed core; independent re-derivation per M001's architectural rule |
| Moving-out-of-a-matched-payload as the standard, default binding semantics (matching an owned scrutinee transfers ownership of the bound payload; matching a borrowed scrutinee borrows the payload) | Table stakes | Direct extension of M001's per-value ability system to a new binding site; the alternative (payloads always implicitly copied/cloned out of a match) would silently reintroduce hidden costs — a category of bug this project's abilities system exists specifically to make visible | MEDIUM-HIGH — this is genuinely new checker surface (match-arm binding sites as a new kind of ownership-transfer point), not a copy of existing logic | M001's affine ability derivation, CFG-last-use liveness (extended to cover match-arm-bound bindings) |
| Error-propagation operator that runs the same drop-obligation cleanup machinery as every other nonlocal exit path in the function (resource cleanup, foreign-boundary unwind, etc.) rather than a bespoke cleanup path | Table stakes for correctness once a propagation operator exists at all | Directly reuses M001's three-stage-acquisition reverse-order cleanup and setjmp/longjmp nonlocal-exit landing pad rather than adding a fourth, independently-reasoned-about early-exit mechanism — avoiding a fourth "coexisting law" per the retrospective's Key Lesson 2 | HIGH — the actual hard part of `Result`, more than the type representation itself | M001's resource acquisition/cleanup machinery, nonlocal-exit landing pad, D-04-31's two known nonlocal-exit blind spots (must be closed here, not compounded) |
| Niche-filling layout optimization for `Result`/payload types where the ownership/ability system already proves a payload's bit pattern excludes a discriminant-usable value (e.g., a proven-non-null owned reference) | Differentiator | This is where Codename Lang's *existing* checked-fact infrastructure (abilities, `originvalidate`) gives it something Rust's niche optimization gets from ad hoc compiler-internal knowledge of specific stdlib types — Codename Lang can make niche eligibility a *derived, explainable* fact from already-checked ownership properties rather than a closed set of compiler-special-cased types, which is both leaner and more transparently auditable in generated C | MEDIUM-HIGH, and should be sequenced *after* the plain tagged-union representation ships and is proven correct at all three tiers (interp/-O0/-O3) — optimizing layout before proving the unoptimized layout is semantically sound repeats the exact "green test, narrow reachable space" failure pattern the retrospective warns about three separate times | Tagged-union representation (above), M001's ability system, five-axis comparator |
| Implicit `From`-based error-type conversion at the propagation operator (Rust's `?`-plus-`Try`-trait auto-conversion) | Anti-feature, at minimum for M002 | Directly in tension with this project's own stated constitution — "no generic `untaint` or ambient authority escape" — an implicit type-converting control-flow operator is exactly the kind of "magic" that undermines "calm canonical form" legibility for a human reviewer scanning generated diffs; Rust itself accepts real debuggability cost for this convenience (a `?` can silently call an arbitrary `From::from`). Recommend requiring an explicit conversion call at the propagation site instead, at least until this project has evidence it wants implicit conversions anywhere | — | — |
| A single monomorphic `Result<T, E>` type baked into the language as a hardcoded pair, versus a general payload-carrying sum-type mechanism that `Result` is merely the standard instance of | Design question, not strictly anti-feature, but worth flagging | PROJECT.md's Out of Scope explicitly excludes "full generics" for now — this means M002's `Result` is very likely to land as a built-in, non-generic-parameterized construct (or minimally parameterized) rather than a library-definable sum type, which is a reasonable, deliberate scope cut consistent with the rest of M001/M002's posture, but should be stated explicitly in the roadmap rather than discovered as an ambiguity mid-phase | MEDIUM either way; the two paths diverge significantly in implementation cost | Interacts with the "Out of Scope: full generics" boundary already declared |

---

## 5. Interprocedural Equivalence Testing

### 5.1 What "the optimizer did not change observable semantics" means across function boundaries

**Translation validation (Alive2).** Alive2 proves, per compiled function,
that an LLVM optimization pass refines (preserves or narrows, in the
undefined-behavior-permitting sense) the original function's semantics,
using an SMT solver over a precise model of LLVM IR including undefined
behavior — but its scope is explicitly **intraprocedural**: `alive-tv`
validates a pipeline "without any interprocedural passes," and its authors
state the SMT encoding is precise enough for "all of LLVM's intra-procedural
memory optimizations" specifically, not interprocedural ones. **(V)**
https://users.cs.utah.edu/~regehr/alive2-pldi21.pdf and
https://github.com/AliveToolkit/alive2 (accessed 2026-09-08). This is an
important, somewhat surprising finding for M002: even the state-of-the-art
formal translation-validation tool for LLVM has not solved general
interprocedural equivalence proof — it is genuinely still an open, hard
problem at the SMT-formal-proof tier, which recalibrates what "interprocedural
equivalence" can realistically mean for M002's five-axis comparator. It means
"observable-behavior equivalence" for Codename Lang almost certainly must
stay at the **differential-testing** tier (comparing concrete executions
across interpreter/-O0/-O3/native, which M001 already does intraprocedurally)
rather than attempting a formal cross-function proof — which is exactly
consistent with PROJECT.md's stated M002 item, "Interprocedural `-O3`/LTO
equivalence on the five-axis comparator," i.e., extending the *existing
empirical* comparator across call boundaries, not building a new formal
verifier.

**Random/generative differential testing at scale.** Csmith is the
foundational random C-program generator used to find compiler miscompilation
by comparing outputs of the same program compiled by multiple
compilers/optimization levels. **(C)**, well-established, widely cited tool
in compiler testing literature. YARPGen improves on this by statically
tracking value-range/overflow constraints during generation to avoid
accidentally exercising undefined behavior (which would make a divergence a
false positive rather than a real compiler bug), and explicitly steers
generation toward patterns likely to exercise optimization-sensitive code
paths; it has found 220+ real bugs across GCC/LLVM/ICC. **(V)** per this
pass's search results (arxiv survey material,
https://arxiv.org/pdf/2507.06584, accessed 2026-09-08, citing YARPGen's
bug-count figures; treat the bug-count figure itself as **(C)**-tier
secondary reporting rather than a primary YARPGen paper citation, since the
primary Intel paper was not directly retrieved this pass). **EMI
(Equivalence Modulo Inputs)**, originated by the "Orion" framework, is a
distinct technique: rather than comparing across compilers, it generates
*program variants* (e.g., by deleting code paths proven unreachable for a
specific concrete input) that are provably behaviorally equivalent *for that
input* to the original program, then checks that a single compiler produces
the same output for both variants — a form of metamorphic testing that
finds miscompilation bugs a Csmith-style cross-compiler comparison would
miss (since it needs only one compiler, it catches bugs that are consistent
across compilers but still wrong relative to the source program's own
semantics). **(V)** https://fm.csl.sri.com/SSFT14/PLDI14-Orion.pdf (accessed
2026-09-08).

**How this composes for M001/M002's existing five-axis comparator.** M001
already runs interpreter/-O0/-O3 differential comparison intraprocedurally
with a mutation-killed control
(`control:alias.false_no_alias`, a real engineered `restrict`-attribute
divergence proven detectable). The natural, low-risk extension for M002 is
not a new proof technique but the same technique applied to a new axis of
program shape: call-graph-shaped programs (direct calls, at minimum one
level of call depth, later multi-level and cross-SCC-boundary-adjacent
shapes even though recursion itself is refused) fed through the same
comparator, with new *interprocedural-specific* negative controls analogous
to `false_no_alias` — for example, an engineered false claim that two
distinct call sites' loans do not interfere, mutation-killed the same way
M001's intraprocedural controls were. **(I)**, direct extrapolation from
M001's documented pattern, which the retrospective explicitly endorses as
the project's standing rule ("every differential gets a mutation-kill").

### 5.2 Table stakes / differentiators / anti-features for §5

| Item | Category | Why | Complexity | Depends on |
|---|---|---|---|---|
| Extend the existing interpreter/-O0/-O3 five-axis comparator to call-graph-shaped programs (multi-function corpora exercising `OpCall`), still empirical/differential, not a formal proof | Table stakes | Directly named in PROJECT.md; consistent with what even Alive2 (the SOTA formal tool) does NOT attempt interprocedurally — differential testing is the correct, achievable tier for this project's maturity | MEDIUM-HIGH — mechanically similar to M001's existing comparator but needs a call-graph-aware corpus generator | M001's five-axis comparator, ASan/UBSan lane, HDD reducer (all should extend, not be replaced) |
| At least one engineered, mutation-killed negative control specifically for interprocedural miscompilation (e.g., a deliberately-false cross-call alias/no-alias claim proven to produce a real interpreter-vs-`-O3` divergence, mirroring M001's `false_no_alias`) | Table stakes, per this project's own standing rule | Retrospective states this as a hard rule: "A control that has never been made to fail is a claim, not evidence" — this is not optional for a project with this rule already codified in its own process | MEDIUM | M001's `restrict`-emission pattern, its `false_no_alias` control as the template to replicate at the interprocedural level |
| A call-graph-aware corpus generator analogous to a minimal Csmith/YARPGen for this project's scope — i.e., randomly generating multi-function programs with varied call shapes, ownership-transfer argument patterns, and `Result` payload flows, constrained to avoid Codename Lang's own defined-UB-equivalent space so divergences are real bugs not false positives (the YARPGen lesson) | Differentiator | This project already has adversarial and bounded-enumeration corpora (per PROJECT.md) at the intraprocedural level; building the interprocedural analog with YARPGen's constraint-aware generation discipline (rather than pure Csmith-style unconstrained randomness) would materially reduce false-positive divergence noise, which matters enormously for an AI-agent-facing tool where a flaky differential is worse than a slow one | HIGH — real generator-construction effort, and the "Generator and Probe Reachability Register" pattern from D-04-21 should be reused so reachability gaps are declared, not discovered late | M001's existing adversarial + bounded-enumeration corpora, D-04-21's reachability register pattern |
| EMI-style variant generation (pruning provably-input-irrelevant call subtrees and checking the pruned/unpruned variants agree) as a *supplementary* technique layered onto the existing comparator | Differentiator, and a lower-priority one for M002 specifically | Genuinely catches a different bug class than cross-tier comparison alone, but it is additive machinery on top of a comparator that does not yet exist interprocedurally at all — sequencing this before the base interprocedural comparator ships would be premature optimization of the evidence pipeline before the pipeline exists | HIGH; explicitly a "later milestone" candidate, not M002 | The base interprocedural five-axis comparator (must exist first) |
| Attempting a formal (SMT-based, Alive2-style) proof of interprocedural equivalence for M002 | Anti-feature for M002 specifically, not forever | Even Alive2 — a mature, well-funded, LLVM-specific formal tool — has not solved this problem interprocedurally; attempting it now would be a multi-year research-grade detour completely disproportionate to a milestone whose own charter is "prove Lang-to-Lang calls exist safely," and it would violate this project's own established risk-budget discipline (the retrospective explicitly credits "deferring `OpCall`... under load" as a correct call for exactly this kind of reason) | — | — |
| Treating a *single* -O3 run with no engineered negative control as sufficient interprocedural equivalence evidence | Anti-feature | Directly contradicts this project's own standing rule (see above); would be the kind of "green test whose reachable input space omitted the hard case" the retrospective names as the single most expensive repeated mistake across three separate M001 phases | — | — |

---

## 6. Breadth Pass — Highest-Leverage Capability Areas Beyond M002's Charter

Given the project's stated intent — AI agents author, humans audit;
deterministic reproducible evidence; usable low-level ownership; calm
canonical form; fast feedback loop — the following candidates were evaluated
for "what comes after the interprocedural spine is proven." Each is rated on
leverage (impact on the stated core value) and readiness (whether M002's
output is a prerequisite).

### Ranked candidates

**1. Generic functions/types restricted to ownership-ability bounds (not full
parametric polymorphism/variance).** *High leverage, not premature for the
milestone immediately after M002.* Real-world sum types (`Result<T, E>`)
and containers are close to unusable as a general, ergonomic feature without
*some* form of generics — M002 is likely to ship a hardcoded, non-generic
`Result` (per §4.2's design-question row), and a monomorphic special case is
a sensible, deliberate near-term cut but not a permanent one. The narrow
form worth prioritizing is bounded generics over ability constraints (e.g.,
"any `T` with `copy`") rather than Rust-style full trait-bound generics with
variance/subtyping, which PROJECT.md explicitly and correctly continues to
exclude for now. Directly unblocks a general `Result<T, E>`/`Option<T>`
rather than a bespoke type, and unblocks reusable containers. **Rank: 1
(highest leverage, natural sequel to M002's `Result` work).**

**2. A minimal effect/capability surface for I/O and non-determinism sources,
scoped far short of "general effect rows."** *Medium-high leverage, real
prematurity risk.* PROJECT.md explicitly defers "general effect rows,
resumable handlers" — correctly, for now. But *some* mechanism for marking
which functions can perform observable I/O or read non-deterministic state
(clock, environment, randomness) is close to required before this language
can be used for real production programs beyond pure-computation corpora,
and it interacts directly with the deterministic-interpreter-oracle
guarantee (an oracle is only meaningful for programs whose I/O surface is
declared and controllable). Recommend treating this as a research spike, not
a milestone commitment yet — it is genuinely adjacent to "general effect
rows" territory and this project's own retrospective explicitly warns
against underscoping this kind of thing (Phase 2-4's repeated "green test
omitted the hard case" pattern is exactly what happens when a
capability-shaped feature is scoped too small on the first attempt).
**Rank: 2, but flagged premature to commit to yet — spike first.**

**3. A minimal, deterministic standard library surface (containers, string
handling, basic I/O) built entirely on the shipped ownership/ability
primitives, with its own evidence manifests.** *High leverage, low novelty
risk, good sequencing fit.* This is the highest-leverage *non-research* item:
it does not require new semantic machinery, it consumes what M001/M002 build
(ownership, `Result`, calls), and it is the actual gating dependency for any
agent being able to write more than toy programs. Foundational languages
that under-invest here (early Zig, early Rust pre-1.0) paid a long "everyone
hand-rolls their own Vec" tax. **Rank: 3, and arguably could be pulled
earlier than generics if `Result` monomorphism is accepted a bit longer** —
worth an explicit roadmap tradeoff discussion, not decided here.

**4. Bounded/proven-terminating recursion (Ada/SPARK Platinum style) as an
alternative to blanket refusal, gated behind an explicit termination-metric
proof obligation.** *Medium leverage, clearly premature.* Named in §3 as a
differentiator this project could plausibly own well, given the ownership
system already produces checked facts a termination metric could consume —
but it requires a proof-obligation subsystem this project has zero
scaffolding for today, and PROJECT.md's own Out of Scope list ("full
generics... reflection, macros") signals a general "no new proof machinery
before the kernel oracle is trustworthy" posture that this would violate.
**Rank: 4, explicitly premature — revisit only after the kernel oracle
(M001+M002's independent-re-derivation architecture) has a track record
across at least one more milestone.**

**5. A compiler-service/incremental-compilation contract (stable API for a
long-running compiler process, as a prerequisite for any future LSP/agent-
loop tooling) — WITHOUT building the LSP or MCP server itself.** *Medium
leverage, timing-sensitive but not this milestone.* PROJECT.md explicitly
excludes "LSP, MCP server" as later-milestone items "built on the stable
compiler service contract" — implying the *contract* (not the protocol
servers) is recognized as a real, separate, valuable near-future target.
Given this project's core value is explicitly about tightening the
generate-verify-run-repair loop for an AI agent, a stable incremental
compiler-service boundary is arguably higher-leverage than either generics
or a stdlib for that specific goal — but M002's interprocedural work
(call-graph construction, signature-as-cache-key discipline named in §2.2)
is a direct and necessary prerequisite, since without stable interprocedural
caching semantics an incremental service has nothing sound to build on.
**Rank: 5 — correctly sequenced after M002 by PROJECT.md's own scope
language, and this research affirms that sequencing is right, not premature
but not yet ready either.**

**6. Formal/SMT-assisted equivalence proof for a bounded subset of the
language (Alive2-style, intraprocedural first).** *Lower near-term leverage,
high long-term differentiation potential, clearly premature now.* Given
§5's finding that even LLVM's own Alive2 has not solved this
interprocedurally, and that this project's differential-testing-plus-
mutation-kill discipline is already both rigorous and appropriately scoped
to its maturity, investing in SMT-based proof machinery before the empirical
comparator has matured across two milestones would be a significant,
premature detour — but it is worth naming as a plausible multi-milestone-out
differentiator precisely because this project's typed core and independent
re-derivation discipline are unusually well-suited to eventually hosting
this kind of proof (more so than most languages, which retrofit formal
verification onto an existing untyped-adjacent IR). **Rank: 6 — name it,
do not schedule it.**

**Honest assessment of what NOT to propose:** async/await, actors,
schedulers, package registry/remote cache, and full generics-with-variance
are all explicitly and correctly out of scope per PROJECT.md, and nothing in
this research surfaces evidence to challenge those exclusions — if anything,
the Rust/Ada/Austral precedent survey reinforces that ownership-and-calls
correctness (M002's actual charter) is the load-bearing prerequisite every
one of those later features would need anyway.

---

## Sources

**Primary (verified, (V)-tagged claims above):**
- Rust two-phase borrows — https://rustc-dev-guide.rust-lang.org/borrow_check/two_phase_borrows.html (accessed 2026-09-08)
- Rust NLL architecture — https://doc.rust-lang.org/nightly/nightly-rustc/src/rustc_borrowck/nll.rs.html (accessed 2026-09-08)
- Polonius design — https://github.com/rust-lang/polonius (accessed 2026-09-08)
- Rust `Try`/`?` operator RFC — https://github.com/rust-lang/rfcs/blob/master/text/1859-try-trait.md (accessed 2026-09-08)
- Rust operator expressions reference — https://doc.rust-lang.org/reference/expressions/operator-expr.html (accessed 2026-09-08)
- Rust tagged-union/niche RFC — https://rust-lang.github.io/rfcs/2195-really-tagged-unions.html (accessed 2026-09-08)
- Rust enum layout optimization — https://medium.com/@zty0826/the-optimization-of-enum-layout-in-rust-cc0e0f34df55 (accessed 2026-09-08)
- Val/Hylo parameter conventions — https://docs.hylo-lang.org/language-tour/functions-and-methods (accessed 2026-09-08)
- Austral linear types tutorial — https://austral-lang.org/tutorial/linear-types (accessed 2026-09-08)
- AdaCore Platinum SPARK bounded stacks case study (GNATstack figures) — https://www.adacore.com/blog/from-ada-to-platinum-spark-a-case-study-for-reusable-bounded-stacks (accessed 2026-09-08)
- MISRA C recursion rule and undecidability analysis — https://arxiv.org/pdf/2212.13933 (accessed 2026-09-08)
- Alive2 paper (interprocedural scope limitation) — https://users.cs.utah.edu/~regehr/alive2-pldi21.pdf and https://github.com/AliveToolkit/alive2 (accessed 2026-09-08)
- EMI/Orion original framework paper — https://fm.csl.sri.com/SSFT14/PLDI14-Orion.pdf (accessed 2026-09-08)

**Convention-level (C)-tagged, cross-referenced but no single primary citation:**
- C++ overload-resolution-based move/copy semantics (general C++ reference knowledge)
- Swift `borrowing`/`consuming`/`inout` as post-hoc ARC-era annotations (general Swift 5.9 release knowledge)
- seL4 static stack sizing convention (general formal-verification-kernel knowledge, not independently re-verified this pass)
- `cargo-call-stack` embedded-Rust convention (general embedded-Rust ecosystem knowledge)
- Csmith/YARPGen bug-finding conventions — https://arxiv.org/pdf/2507.06584 (secondary reporting, accessed 2026-09-08)
- ABI-level RVO/`sret` return convention (standard LLVM/C ABI convention, uncontested but no single citation retrieved this pass)

**Project-internal (required reading):**
- ~/projects/ai-lang/.planning/PROJECT.md
- ~/projects/ai-lang/.planning/RETROSPECTIVE.md

---
*Feature research for: Codename Lang, M002 Interprocedural Semantic Spine*
*Researched: 2026-09-08*
