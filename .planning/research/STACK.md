# Stack Research — M002 Interprocedural Semantic Spine

**Domain:** Compiler tooling — proving cross-function semantics for a from-scratch
AI-authored/human-audited language with a Go 1.24 stdlib-only host and a
readable-C17-through-Clang native path.
**Researched:** 2026-09-08
**Confidence:** MEDIUM-HIGH (primary sources for all libraries/flags; some
project-fit judgments are inference, labeled as such)

**Scope discipline:** this file covers only what M002 needs to build and prove
`OpCall`, call-graph construction, interprocedural loan liveness, and
interprocedural `-O3`/LTO equivalence. It does not revisit M001's validated
stack (Go 1.24, Clang, readable C17, the five-axis comparator, ASan+UBSan,
the HDD reducer) — those are load-bearing and out of re-research scope per
the milestone context.

Local environment verified: `go1.24.0 darwin/arm64`, `Apple clang version
21.0.0 (clang-2100.1.1.101)` — both already match what the recommendations
below assume.

---

## Recommended Stack

### Core additions (all zero new production dependencies)

| Technology | Version | Purpose | Why Recommended |
|---|---|---|---|
| `pgregory.net/rapid` | v1.3.0 (pkg.go.dev, confirmed current) | Property-based generators + shrinking for call-graph shapes, argument/loan combinations, cross-function borrow scenarios | Pure Go, stdlib-only per its own doc ("no dependencies outside the Go standard library"), integrates as a normal `go test` dependency — **but see "What NOT to Add" for why this is a `_test.go`-only addition, never imported by production code** |
| `go test -fuzz` (Go 1.24 stdlib, `testing` package) | Already present | Differential fuzzing of `check`/`corevalidate`/`interp`/native across call boundaries | Zero new dependency — it is the Go toolchain M001 already uses; corpus lives in-repo under `testdata/fuzz/` and is git-visible, matching the project's evidence-manifest philosophy |
| `golang.org/x/perf/cmd/benchstat` | latest via `go install golang.org/x/perf/cmd/...@latest` (module `golang.org/x/perf`, x/perf is unversioned/rolling — pin by commit or `go.sum` at whatever `go install` resolves at integration time) | A/B statistical comparison of compile-time and runtime cost pre/post interprocedural changes | Official Go sub-repo (golang.org/x team), not a third-party dependency in spirit; run as a **dev tool via `go run`/`go install`, not imported by module code** — keeps `go.mod` clean, matches M001's `internal/compiler/measure` p50/p95/CoV protocol by giving it a paired-comparison front end |
| `golang.org/x/tools/go/callgraph/{cha,rta,vta}` — **read, not imported** | current `golang.org/x/tools` (already an indirect toolchain dependency via `gopls`/`go vet` tooling, not a runtime import) | Reference algorithms for call-graph construction: CHA (sound on partial programs, conservative), RTA (dynamic reachability from roots), VTA (refines with type flow, still explicitly experimental) | See "Call-graph tooling" section — recommendation is to **study the algorithm shapes and read the source**, not import the package, because Lang's `OpCall` callable-set is closed and typed (no Go-style interface dispatch), so CHA/RTA/VTA solve a harder, more general problem than M002 has |

### Explicitly NOT added (see rationale table below)

SMT solvers (Z3/CVC5/Alloy), any mutation-testing framework (go-mutesting,
gremlins, ooze), gopter/testing-quick as alternatives to rapid, MSan, CFI,
libFuzzer-via-cgo, and any x/tools callgraph package as a runtime import.

---

## Detailed Findings by Question

### 1. Property-based testing, BMC/SMT, differential fuzzing

**Property-based testing — recommend `pgregory.net/rapid` v1.3.0.**
- **Verified (HIGH, primary source):** `pkg.go.dev/pgregory.net/rapid` states
  "No dependencies outside the Go standard library." Module path is
  `pgregory.net/rapid` (the GitHub mirror `flyingmutant/rapid` is the same
  project; the canonical import path is the vanity `pgregory.net/rapid`).
- **Why it earns its place over hand-rolled generators:** M001's own
  retrospective flags D-04-21's "Generator and Probe Reachability Register" —
  the project already learned that ad hoc generators silently fail to reach
  hard shapes. `rapid`'s integrated shrinker turns a generated interprocedural
  counterexample (e.g., a 4-function cycle with a shared loan crossing two
  calls) into a minimal reproduction automatically, which a copy-local
  generator would need to reimplement from scratch (shrinking is the genuinely
  hard part, not generation).
- **Why NOT `gopter`:** `gopter` (`leanovate/gopter`) is QuickCheck-style but
  has been effectively unmaintained (long gaps between releases, open issues
  with no response) and its shrinking model is weaker/more manual than
  rapid's integrated approach — **widely-reported convention (MEDIUM)**,
  consistent with rapid having become the de facto Go community recommendation
  in recent Go blog posts and conference talks, but I did not find a single
  authoritative "gopter is deprecated" statement, so treat this as a
  comparative judgment, not a fact about gopter's status.
- **Why NOT `testing/quick`:** it is in the Go stdlib but has been in a
  frozen, unmaintained state for years (no shrinking, minimal generator
  composition, the Go team has stated informally it is not a good model for
  new code) — do not build M002's cross-function properties on it.
- **Integration point:** new `_test.go` files under `internal/compiler/check`,
  `internal/compiler/corevalidate`, and a new `internal/compiler/callgraph`
  package. Rapid generates: (a) call-graph shapes (acyclic DAGs up to depth
  N, deliberately including cycles to exercise refusal), (b) loan-crossing
  argument/return combinations, (c) `Result`/payload alternative shapes for
  the new storable/matchable values. Cost: one new `go.mod` `require` line,
  test-only (does not touch the zero-external-dependency **production**
  build — `go build ./cmd/...` never imports it).
- **Cost:** near-zero. It's a single pure-Go module, MIT-licensed, no
  transitive dependencies per its own doc.

**BMC/SMT (Z3, CVC5, Alloy) — do not add for M002.**
- **Verified (HIGH):** `aclements/go-z3` and `mitchellh/go-z3` both bind Z3
  via **cgo**, requiring the Z3 C/C++ library to be built and present on the
  host (`pkg.go.dev/github.com/aclements/go-z3`: "you can't simply go get
  this library... Z3 must first be built"). This is a direct, structural
  violation of the project's stated constraint ("Prefer standard-library-only
  and shallow audited boundaries; every dependency must earn more than
  copy-local implementation") and of the harder fact that M001 shipped 6
  phases at **zero external dependencies** using pure Go.
- **Why it doesn't earn the cost here, specifically for M002's scope:**
  M002's interprocedural properties (loan liveness across calls, cycle
  refusal, callable⊆publishable) are **decidable by construction** with the
  project's existing fixpoint/reachability-closure techniques
  (`loanLivenessFixpoint`, `corevalidate`'s reachability-closure-plus-reduction)
  extended across a call graph — these are graph-reachability problems, not
  problems requiring general first-order SMT solving. SMT earns its keep when
  you need to discharge quantified arithmetic or complex boolean-combination
  side conditions that a fixpoint can't express; nothing in M002's charter
  (cycle refusal, bounded call stack, cross-function loan liveness, Result
  matchability) is in that category.
- **What SMT-style verification WOULD be for:** a later milestone doing
  numeric refinement types, bounds-proof elimination, or effect-row
  satisfiability might genuinely need it. Flag this in PITFALLS/roadmap as a
  deferred capability, not a rejected one — **inference (LOW-MEDIUM)**, my
  own architectural judgment based on what M002 charters, not a sourced claim.
- **Alloy:** a modeling language/tool (not embeddable in Go at all — it's a
  standalone Java-based analyzer). Worth using **offline, outside the repo,
  as a design-time model** of the call-graph cycle-refusal algorithm before
  implementing it, exactly the way M001's spikes in `.planning/spikes/`
  worked — never as a runtime or CI dependency. This is consistent with the
  project's own pattern of bounded workbenches that produce contracts, not
  binaries.

**Differential fuzzing — use Go 1.24 native fuzzing, do not add libFuzzer/cgo.**
- **Verified (HIGH, primary source, `go.dev/doc/security/fuzz/` and
  `go.dev/src/testing/fuzz.go`):** Go's native fuzzing (`go test -fuzz`,
  stable since Go 1.18) targets a single function of primitive-typed
  arguments (`func FuzzXxx(f *testing.F)`, then `f.Fuzz(func(t *testing.T,
  ...))` with parameter types restricted to strings, byte slices, and
  numeric/bool primitives — structs are not directly supported as fuzz
  target parameters). This is a real constraint for M002: you cannot fuzz
  "a call graph" directly as a struct.
- **How to use it despite that constraint:** fuzz a **serialized encoding**
  of a call graph / program fragment (the same pattern M001 already uses for
  its adversarial and bounded-enumeration corpora) — e.g. a byte-string DSL
  or small integer-vector encoding decoded into a call graph inside the fuzz
  target, then run the existing four engines (`check`, `corevalidate`,
  `interp`, native) differentially and fail on divergence. This is exactly
  M001's five-axis comparator pattern, extended with a Go-native fuzz driver
  instead of only a fixed corpus generator, so it is additive to existing
  infrastructure rather than a new subsystem.
- **Why NOT libFuzzer via cgo:** libFuzzer requires cgo and a C/C++
  coverage-instrumented build; Go's native fuzzer already gets
  coverage-guided mutation via the toolchain's own instrumentation
  (`go.dev/doc/security/fuzz/`), so cgo would add build complexity and an
  OS/toolchain-dependent C boundary for a capability the project already has
  natively. There is no interprocedural-specific reason to reach for
  libFuzzer — the target under fuzz is Lang programs (data), not the Go
  compiler binary's own C internals.
- **Integration point:** new `FuzzOpCall`, `FuzzLoanLiveness`,
  `FuzzResultMatch` targets under `internal/compiler/*` alongside existing
  tests; corpus seeds derived from Phase 3's loan-endpoint differential
  corpus, extended across function boundaries per M002's charter ("Cross-
  function rebuild of Phase 3's exhaustive loan-endpoint differentials").
- **Cost:** zero new dependency. CI cost: fuzzing runs are typically bounded
  by `-fuzztime` in CI (e.g. seconds-to-minutes per target) with longer runs
  as a separate nightly/scheduled lane — matches M001's existing "expensive
  evidence runs only when its distinct question is relevant" constraint.

### 2. Mutation testing for Go

**Recommend: do not adopt a mutation-testing framework as infrastructure for
M002. Continue M001's hand-authored mutation-kill discipline instead.**

- **Verified (MEDIUM, from GitHub project pages/issues, dated within the
  last research window):**
  - `go-mutesting` (`zimmski/go-mutesting`, with an active fork at
    `avito-tech/go-mutesting`): the original upstream is effectively
    inactive; the fork continues but is a community patch set, not a
    first-party-maintained tool.
  - `gremlins` (`go-gremlins/gremlins`): latest release v0.6.0 (December
    2025), still explicitly pre-1.0 ("0.x.x release, which doesn't guarantee
    backward compatibility"), the project's own docs state config flags can
    change between minor releases, and — most relevant to a compiler of
    M002's size — **"Gremlins doesn't work very well on very big Go
    modules... a run can take hours to complete."** The project is already
    ~59,150 lines of Go plus growing; a whole-module mutation run at that
    scale is a real CI-time risk, not a hypothetical one.
  - `ooze`: minimal evidence of active maintenance found in this research
    pass; treat as LOW confidence / not evaluated further given gremlins is
    the more visible active option and still has the scale problem above.
- **Why this matters more than "maturity" alone:** the project's own stated
  law (retrospective, Key Lesson 1) is **"a control you have never seen fail
  is a claim."** That law is about *your own differentials and controls*
  (e.g. `control:alias.false_no_alias`), each individually engineered and
  understood — not about achieving blanket automated mutation coverage of
  the whole codebase. A generic mutation-testing tool answers "does the test
  suite kill *some* random AST mutation," which is a different and weaker
  question than "does this specific interprocedural control
  (callable⊆publishable violation, a cross-function loan liveness escape, a
  call-graph cycle) fail exactly the way I designed it to." M001 got this
  right by hand (the `false_no_alias` engineered divergence); M002 should
  extend that same hand-authored pattern to interprocedural controls
  (a deliberately unsound cross-function loan-liveness mutant, a
  deliberately-broken cycle-refusal check) rather than delegate to a
  general-purpose mutator whose kill targets won't line up with the specific
  claims M002 needs to defend.
- **Where a mutation tool WOULD earn its place:** as a **periodic audit**
  (not a gate) run manually/nightly on a bounded package subset (e.g. just
  `internal/compiler/callgraph`) to catch *unknown-unknown* gaps in test
  coverage that hand-authored controls don't think to check. If adopted
  later, `gremlins` v0.6.0 is the better-maintained of the two, but scope it
  to a single small package given its own documented scaling limit, and
  treat its findings as leads to investigate, not gate failures — **this is
  an inference recommendation (MEDIUM), consistent with but not identical
  to the project's existing debt-register pattern (`*-DEBT.md`).**
- **Cost if deferred:** zero. Cost if adopted narrowly later: one dev-tool
  binary (`go install`), no `go.mod` production entry, CI minutes bounded by
  scoping to one package.

### 3. Call-graph / interprocedural analysis tooling

**Recommend: read `golang.org/x/tools/go/callgraph/{cha,rta,vta}` source as
reference; do not import any of them.**

- **Verified (HIGH, from pkg.go.dev source/docs):**
  - **CHA** (Class Hierarchy Analysis): computes the entire
    interface-implements relation ahead of time; conservative; explicitly
    "sound to run on partial programs, such as libraries without a main or
    test function." Cheapest, least precise.
  - **RTA** (Rapid Type Analysis): builds reachable-code and runtime-type
    sets incrementally from program roots (`main`/tests); needs a whole
    program (an entry point), not sound on partial programs.
  - **VTA** (Variable Type Analysis): refines an initial call graph using
    type flow; **the package itself documents that it is "in experimental
    phase and its interface is subject to change,"** with open TODOs around
    generic function bodies and instantiation wrappers.
- **Why not import, even though it's "free" (already in the Go toolchain
  ecosystem):** all three algorithms exist to solve Go's dispatch problem —
  resolving calls through **interfaces** and **first-class function values**
  where the callee set isn't syntactically fixed. Lang's M002 charter is
  narrower and *harder to get wrong in a good way*: `OpCall` dispatch is
  gated on **callable ⊆ publishable**, a closed, typed, checker-verified
  set — this is closer to whole-program monomorphic call resolution (like a
  first-order functional language without first-class functions yet) than
  to Go's dynamic-dispatch problem. Importing VTA in particular would mean
  depending on a package whose own maintainers call it experimental and
  unstable, to solve a problem M002 doesn't have (indirect/virtual dispatch).
  A hand-rolled direct call-graph builder over Lang's typed core is both
  simpler and exactly matched to what needs proving (cycle refusal,
  reachability for the bounded interpreter call stack).
- **What to actually borrow (design, not code):** SCC (strongly-connected-
  component) construction is the right algorithmic shape for cycle refusal
  — this is the same structure LLVM/MLIR use for their call-graph SCC pass
  (`llvm/Analysis/CallGraphSCCPass`, and MLIR's equivalent pass
  infrastructure processes an SCC bottom-up so each function's summary is
  available before its callers are analyzed). Read `x/tools/go/callgraph`'s
  `BuildCallGraph`/SCC helpers and read LLVM's `CallGraphSCC.cpp` structure
  for the traversal-order pattern (bottom-up SCC processing = the natural
  order for computing interprocedural loan-liveness summaries, since a
  callee's summary must exist before its caller's is computed, and a cycle
  must be refused rather than iterated to a fixpoint if M002's charter is
  "cycle refusal" rather than "recursive fixpoint convergence").
- **Rust's polonius/NLL interprocedural story — read, don't depend:**
  Rust's borrow checker (NLL, and its still-unshipped polonius successor)
  is **deliberately intraprocedural** — Rust does not attempt whole-program
  or cross-function borrow inference; instead it uses **function
  signatures** (explicit lifetime parameters) as the interprocedural
  contract, checking each function's body against its own signature and
  trusting (not re-deriving) callee signatures at call sites. This is
  directly relevant precedent for M002: Lang's own "declared public
  origins" pattern (already established in M001 — "Infer local borrow ends
  at CFG last use; declare public origins") is the same architectural
  choice Rust made, and D-03-02 (closing in M002) is precisely the
  interprocedural half of that same declared-origin contract. This is
  **verified as Rust's documented design stance (HIGH, rust-lang NLL RFC
  and polonius project documentation are the primary sources — not fetched
  in this pass but this is well-established, widely-published Rust
  compiler-team position)**, cited here as architectural precedent, not as
  a dependency recommendation — there is no Go/C-embeddable polonius
  library to depend on, and none should be sought.
- **Cost:** zero (reading, not importing).

### 4. Clang/LLVM interprocedural facts for `-O3`/LTO equivalence

This is where the project's own stated constraint bites hardest: "optimizer
attributes and FFI guarantees must derive from checked facts." M001 already
proved this pattern intraprocedurally (`restrict` emitted only on
checker-proven no-alias parameters, with `false_no_alias` as the engineered
negative control). M002 must extend exactly that discipline across calls.

- **Attributes that must be DERIVED, not asserted, once calls exist:**
  - **`noalias`** (parameter attribute): already earned intraprocedurally in
    M001 via checker-proven no-alias. Once `OpCall` exists, a `noalias`
    parameter's guarantee must additionally hold **at every call site that
    passes it through** — LLVM's own `noalias` semantics require both "not
    aliasing any other argument" AND "not captured prior to the call" (verified,
    LLVM LangRef / FunctionAttrs documentation). A borrow that is proven
    non-aliasing inside one function but passed into a callee that could
    capture it must not retain the `restrict`/`noalias` emission unless the
    callee's own signature is checked not to capture it — this is exactly
    where interprocedural loan liveness (M002's charter item) is the fact
    that licenses (or revokes) the attribute, not the other way around.
  - **`nocapture`**: LLVM's `FunctionAttrs` pass derives this via an
    interprocedural SCC walk ("Deduce nocapture attributes for the SCC" —
    verified from LLVM `FunctionAttrs.cpp` doxygen/source). Lang should
    derive its own equivalent fact (does a callee's core-IR ever let an
    escape-forbidden reference outlive the call) from the checker's
    existing escape/abilities system, and only then choose whether to
    *emit* `restrict`/pass a raw pointer vs. a defensively-copied one —
    never emit LLVM's `nocapture`/`noalias` C-level equivalents (`restrict`)
    speculatively.
  - **`readonly`/`readnone` (now spelled via `memory(...)` attributes in
    current LLVM, but `restrict`-adjacent C-level reasoning is the same):**
    relevant if M002's `Result`/payload work or call dispatch wants to let
    Clang treat a pure query call as side-effect-free for CSE/hoisting
    across calls — only justified if Lang's own effect/purity facts (which
    M002 does not claim to add — effects are explicitly Out of Scope) are
    established. **Conclusion: do not attempt to emit any purity-implying C
    annotation in M002** — there is no checked fact to derive it from yet,
    and asserting one without a checked basis is exactly the failure mode
    the project's own constraint forbids.
  - **`willreturn`/`mustprogress`**: relevant to whether Clang can hoist/
    reorder across a call at all. M002's bounded interpreter call stack and
    cycle refusal are actually the checked fact that would license this:
    if the call graph is proven acyclic/bounded (no unbounded recursion,
    which the charter's "cycle refusal" establishes), that is a genuine,
    checked basis for eventually asserting `willreturn`-equivalent
    reasoning — but the C code Lang emits doesn't need to *assert* this
    attribute explicitly for `-O3` correctness (Clang derives it itself for
    ordinary C functions when it can prove termination or when the function
    has no calls to `noreturn`/unbounded loops); flag as a future
    opportunity, not an M002 deliverable.
  - **`sret`/`byval`** (ABI-level, not optimization-level attributes):
    these already exist in M001's "one authoritative `core.ForeignContract`
    generating three lockstep-derived inspectable layers" for the C ABI —
    M002 doesn't change ABI-attribute derivation, it only adds more call
    sites that must go through the same lockstep-derived path. No new
    stack item — this is a call-site-count scaling concern for existing
    machinery, not a new tool.
  - **`nounwind`**: Lang's native path uses `setjmp`/`longjmp` (M001) for
    nonlocal exit, not C++ exceptions, so `nounwind` should already be safe
    to derive from "this function contains no C++-exception-raising
    construct" — verify this still holds once calls can transitively reach
    a `longjmp` landing pad; this is a real interprocedural fact to check
    (does a `nounwind`-eligible callee actually never reach a `longjmp`
    site transitively) rather than a new tool need.

- **`restrict` and TBAA across function boundaries (verified, HIGH,
  `llvm.org/docs/LinkTimeOptimization.html` and general LLVM alias-analysis
  docs):** C's `restrict` is a **per-translation-unit, per-function-
  signature** promise; without LTO, Clang cannot use a `restrict` promise
  made in one `.c` file to optimize a call from another `.c` file compiled
  separately, because the two are optimized independently. **This is the
  single most important LTO fact for M002:** if Lang's readable-C17 emission
  puts each Lang function (or module) in a separate translation unit, then
  interprocedural `restrict`/no-alias benefits **only materialize under
  `-flto`** — without LTO, cross-function calls are opaque at the C level
  regardless of what Lang's checker proved. This means M002's "Interprocedural
  `-O3` equivalence" charter item is not really about `-O3` alone — it's
  about proving equivalence **specifically at `-flto`**, because that's the
  only mode where the C compiler can even see across the Lang-level call
  boundary. `-O3` without `-flto` on separately-compiled TUs would silently
  under-test the interprocedural claim (equivalence would trivially hold
  because the C compiler isn't doing anything interprocedural at the C
  level either).
- **What `-flto` actually changes (verified, HIGH,
  `llvm.org/docs/LinkTimeOptimization.html`):** Full LTO makes all functions
  visible to a single whole-program optimization pass (enabling
  cross-TU inlining, constant propagation, and attribute propagation, i.e.,
  exactly the `noalias`/`nocapture` SCC-walk deduction described above);
  ThinLTO instead builds a global function-summary index and optimizes each
  TU mostly in parallel with targeted cross-module inlining decisions
  driven by that summary. **For M002:** stick with Full LTO (matches M001's
  existing "`-flto`-proven-non-inert `-O3` tier" — do not introduce ThinLTO,
  which trades some cross-module precision for parallel build speed the
  project doesn't need at its current scale, and which would add a second
  LTO mode to validate equivalence under, doubling the comparator's proof
  surface for no charter-required benefit).
- **How to prove an optimization tier is "non-inert" interprocedurally
  (this is the direct generalization of M001's `false_no_alias` control):**
  M002 needs at least one **engineered interprocedural divergence control**
  analogous to `control:alias.false_no_alias` — e.g. a deliberately-wrong
  cross-function no-alias claim (a loan proven live across a call when it
  isn't) that must produce an observable `interpreter == -O0 != -O3/-flto`
  divergence, the same way M001's did on Apple clang 21. Without this, "the
  interprocedural `-O3`/LTO tier is non-inert" is an assertion, not
  evidence — directly the shape of the project's own Key Lesson 1. This is
  an integration point into the existing five-axis comparator
  (`internal/compiler/...` wherever M001's comparator lives), not a new
  tool: extend its corpus and controls, don't build a second comparator.
- **Cost:** zero new dependency (this is compiler engineering work against
  the existing Clang toolchain M001 already targets); the cost is design/
  implementation effort in the call-graph and cgen packages, plus one or
  more new engineered negative-control fixtures.

### 5. Sanitizer coverage across calls

- **ASan/UBSan interprocedurally (verified, HIGH,
  `clang.llvm.org/docs/AddressSanitizer.html`):** the flags that matter
  once real function calls exist:
  - `-fsanitize-address-use-after-return=runtime` (Clang's default) or
    `=always` — becomes directly relevant once Lang has real call/return
    with locals whose addresses could theoretically be taken and escape a
    frame; M001's existing ASan+UBSan lane should confirm this is enabled
    (it is Clang's runtime-configurable default, verify it isn't disabled
    via `ASAN_OPTIONS`) since a call boundary is exactly where a
    use-after-return bug would surface.
  - `-fsanitize-address-use-after-scope` — relevant for the bounded
    interpreter call stack's frame lifetime, same reasoning.
  - No new *flags* are strictly required beyond what M001 already runs;
    what's new is the **corpus**: the existing ASan+UBSan lane needs
    interprocedural fixtures added (a call chain that would trigger
    stack-use-after-return / heap-use-after-free only when a value crosses
    a call boundary), not new sanitizer machinery.
- **MSan (MemorySanitizer) — feasible but not recommended for M002 scope.**
  **Verified (HIGH, `clang.llvm.org/docs/MemorySanitizer.html`):** MSan
  requires *all* linked code to be instrumented for full precision, but
  ships 70+ libc interceptors making uninstrumented libc usable with
  reduced (not zero) precision; achieving high-fidelity MSan typically
  requires an instrumented libc++ / custom-built runtime, which is
  real infrastructure cost. **Recommendation: defer MSan past M002.**
  Rationale: MSan detects uninitialized-memory reads, which is a
  *different* defect class than what M002's charter is proving (call-graph
  soundness, interprocedural loan liveness, `-O3`/LTO equivalence). ASan+
  UBSan already cover the memory-safety classes most relevant to loan/
  ownership bugs (use-after-free, use-after-return, UB from misused
  pointers). Add MSan as a candidate for a later milestone once
  uninitialized-payload semantics for the new `Result`/payload-carrying
  alternatives (M002's own charter item) becomes a specific suspected risk
  — at that point it's a scoped, justified addition, not a blanket "more
  sanitizers is better."
- **`-fsanitize=cfi` — not recommended for M002.** **Verified (HIGH,
  `clang.llvm.org/docs/ControlFlowIntegrity.html`):** most CFI schemes
  require `-flto` (already in use) plus `-fvisibility=hidden` and static
  linking of all TUs defining indirectly-called functions; forward-edge
  `cfi-icall` checks that an indirect call's target matches an expected
  signature set. **Why not now:** CFI is a *security hardening* mechanism
  (detecting attacker-controlled indirect-call hijacking at runtime), not a
  *correctness-proving* mechanism — it doesn't help prove interprocedural
  semantic equivalence, which is M002's actual charter. It's also KCFI (the
  non-LTO, kernel-oriented variant) that's lighter-weight; full CFI's
  `-fvisibility=hidden`+static-link requirement could interact awkwardly
  with M001's existing FFI/`nm -u` symbol-gating machinery in ways that
  would need separate validation. Revisit CFI when/if Lang gets a genuine
  indirect-call construct (function values, dynamic dispatch) that widens
  the attack surface CFI defends — `OpCall` in M002 is direct/typed
  dispatch per the charter, so there's no indirect-call surface for CFI to
  protect yet.
- **Stack-related checks for the bounded interpreter call stack:** this is
  a Lang-level invariant (the interpreter enforces its own bounded call
  depth), not something a C-level sanitizer proves — the relevant existing
  tool is `-fsanitize=address` with its stack-overflow-adjacent guard pages
  (ASan's default stack red-zoning) as a *backstop*, but the actual
  correctness claim (cycle refusal + bounded depth) must be proven by
  Lang's own call-graph analysis and interpreter, with ASan only catching
  the case where that proof was wrong and a real stack overflow occurs.
  No new tool — this is implementation work plus test fixtures (an
  adversarial deep/cyclic call program that must be *refused at check
  time*, and a negative control that removes the refusal to confirm ASan
  or a crash actually catches the resulting stack exhaustion).
- **Cost:** zero new dependency for ASan/UBSan corpus extension (uses
  existing lane). MSan and CFI: explicitly deferred, not costed for M002.

### 6. Benchmarking/measurement additions

- **Recommend `golang.org/x/perf/cmd/benchstat`, dev-tool only.**
  **Verified (HIGH, pkg.go.dev/golang.org/x/perf and
  pkg.go.dev/golang.org/x/perf/cmd/benchstat):** official golang.org/x
  sub-repo; computes median + confidence interval and A/B comparison across
  benchmark runs, recommending ≥10 runs per benchmark for statistical
  significance, default α=0.05.
  - **Why it earns its place over ad hoc comparison:** M001 already
    established a p50/p95/CoV measurement protocol
    (`internal/compiler/measure`) — `benchstat` is the natural paired-
    comparison front end for "did adding interprocedural loan liveness
    regress compile time" and "is native-call overhead measurably different
    from the intraprocedural baseline," without hand-rolling statistical
    comparison code. It consumes standard `go test -bench` output directly,
    so there's no new benchmark-authoring format to adopt.
  - **Integration point:** `go install golang.org/x/perf/cmd/benchstat@latest`
    as a documented dev-tool (like `goimports`/`staticcheck` typically are
    for Go projects) — **not** a `go.mod` `require`. Wire into the existing
    `internal/compiler/measure` reporting alongside `risk_lanes.json`/
    `qlt02_budget_manifest.json` outputs: before/after benchmark pairs for
    (a) call-graph construction time on the corpus, (b) interprocedural
    `check`/`corevalidate` admission time, (c) native runtime cost of calls
    at `-O0` vs `-O3`/LTO.
  - **What's new to benchmark that wasn't before:** call-graph construction
    cost (new — didn't exist pre-M002), interprocedural loan-liveness
    fixpoint cost (extends the existing intraprocedural benchmark), and
    native call-overhead at each optimization tier (new axis on the
    existing five-axis comparator).
  - **Cost:** zero new production dependency; one dev-tool install,
    documented in project tooling docs (e.g. a `tools.go`-style pattern or
    a README dev-setup note — Go's common "tool dependency" convention is a
    blank-import `tools.go` file under a build tag, which M001's zero-dep
    posture should probably still avoid for x/perf specifically since it's
    not needed at build time — keep it as a documented `go install` step,
    not a repo-tracked tool dependency).

---

## What NOT to Add — Summary Table

| Avoid (for M002) | Why | Use Instead |
|---|---|---|
| Z3/CVC5 (cgo-bound SMT, e.g. `aclements/go-z3`) | Requires building/linking an external C/C++ library (cgo), breaking the zero-external-dependency record; M002's interprocedural properties are graph-reachability problems solvable by extending existing fixpoint/reachability-closure techniques, not SMT-shaped problems | Extend `loanLivenessFixpoint`/`corevalidate`'s reachability-closure approach across the call graph |
| Alloy as a CI/build dependency | Standalone Java modeling tool, not embeddable; solves a different problem (exploratory model-checking) than build-time proof | Use offline, design-time only, the way `.planning/spikes/` already works — never wire into the pipeline |
| `gopter` or `testing/quick` for new property tests | `gopter`'s maintenance cadence and shrinking model are weaker than rapid's (comparative judgment, MEDIUM confidence); `testing/quick` is stdlib but frozen/unmaintained with no real shrinking | `pgregory.net/rapid` v1.3.0 |
| libFuzzer via cgo | Requires a C/C++ coverage-instrumented build and a cgo boundary; Go 1.24's native fuzzer already gives coverage-guided mutation for the actual fuzz target (Lang programs as data, not the compiler's own C internals) | `go test -fuzz` (Go 1.24 stdlib) |
| `go-mutesting` | Upstream (`zimmski/go-mutesting`) is inactive; only community forks remain active | If mutation testing is wanted later, use `gremlins`, scoped narrowly (see below) |
| `gremlins` (or any mutation tool) as a **CI gate** | Still pre-1.0 (v0.6.0), config/flags not stable across minor releases per its own docs, and its own maintainers document it doesn't scale to large modules ("a run can take hours to complete") — this project is already ~59K LOC and growing | Continue hand-authored, individually-engineered mutation-kill controls per differential (M001's `control:alias.false_no_alias` pattern), extended to interprocedural controls; use a mutation tool (if at all) as a narrow, non-gating, periodic audit on one small package |
| `golang.org/x/tools/go/callgraph/{cha,rta,vta}` as an **import** | Solves Go's dynamic-dispatch problem (interfaces, first-class function values); Lang's `OpCall` is closed/typed dispatch (callable⊆publishable), a simpler and different problem; VTA is explicitly documented as experimental/unstable by its own maintainers | Hand-rolled direct call-graph builder over Lang's typed core; read x/tools and LLVM's `CallGraphSCCPass` for SCC traversal-order design only |
| `-fsanitize=cfi` for M002 | Security-hardening mechanism (indirect-call-hijack detection), not a correctness-proving one; `OpCall` in M002 is direct/typed dispatch, so there's no indirect-call attack surface yet; adds `-fvisibility=hidden`+static-link constraints that would need separate validation against the existing FFI/`nm -u` symbol-gating machinery | Revisit only if/when Lang gains genuine indirect-call/function-value dispatch |
| MSan for M002 | Requires an instrumented libc++/custom runtime for full fidelity — real infrastructure cost; detects a different defect class (uninitialized memory) than M002's charter (call-graph soundness, interprocedural loan liveness, `-O3`/LTO equivalence) | ASan+UBSan lane, extended with interprocedural fixtures; revisit MSan if `Result`/payload uninitialized-read risk becomes concrete |
| ThinLTO as a second/replacement LTO mode | Trades cross-module precision for parallel build speed the project doesn't need yet; would double the equivalence-proof surface (two LTO modes to validate) for no charter-required benefit | Keep Full LTO (`-flto`), matching M001's existing tier |
| Any new numeric/statistics library for benchmark comparison | `benchstat` (official `golang.org/x/perf`) already does paired A/B comparison correctly (median + CI, configurable α) | `golang.org/x/perf/cmd/benchstat` as a dev tool |
| A second/separate comparator for interprocedural equivalence | Would duplicate M001's five-axis comparator infrastructure instead of extending it, doubling maintenance for the same underlying question ("do engines agree") | Extend the existing five-axis comparator's corpus and controls with call-crossing fixtures |

---

## Version Compatibility

| Package/tool | Compatible with | Notes |
|---|---|---|
| `pgregory.net/rapid` v1.3.0 | Go 1.24.0 (verified locally) | No dependencies, pure Go — no compatibility risk |
| `go test -fuzz` | Go 1.24.0 stdlib | Already the toolchain in use; no version action needed |
| `golang.org/x/perf/cmd/benchstat` | Go 1.24.0 | Installed via `go install`, not a module dependency — pin the resolved version in dev-tooling docs when first installed, since x/perf does not follow strict semver tagging discipline the way core modules do |
| Clang 21 (Apple clang-2100.1.1.101, locally verified) | `-flto` Full LTO, `-fsanitize=address,undefined`, `-fsanitize-address-use-after-return` | All flags referenced above are supported at this Clang version per current (not archived) `clang.llvm.org/docs/` pages fetched in this research pass |
| `golang.org/x/tools/go/callgraph/vta` | N/A — not a dependency | Cited only as reading material; its own docs mark it experimental, so do not pin/import even transitively |

---

## Sources

- `pkg.go.dev/pgregory.net/rapid` — confirmed v1.3.0, zero-dependency claim (HIGH, fetched 2026-09-08)
- `github.com/go-gremlins/gremlins` releases page — v0.6.0, Dec 2025, pre-1.0 status, scaling limitation (HIGH, fetched 2026-09-08)
- `github.com/zimmski/go-mutesting`, `github.com/avito-tech/go-mutesting` — upstream inactivity, active fork (MEDIUM, GitHub search results, fetched 2026-09-08)
- `pkg.go.dev/golang.org/x/tools/go/callgraph/{cha,rta,vta}` — algorithm docs, VTA experimental status (HIGH, fetched 2026-09-08)
- `go.dev/doc/security/fuzz/`, `go.dev/src/testing/fuzz.go` — Go native fuzzing target signature constraints (HIGH, fetched 2026-09-08)
- `pkg.go.dev/github.com/aclements/go-z3` — cgo/build requirement for Z3 bindings (HIGH, fetched 2026-09-08)
- `llvm.org/docs/LangRef.html`, LLVM `FunctionAttrs.cpp` doxygen — `noalias`/`nocapture` interprocedural deduction (HIGH, fetched 2026-09-08)
- `llvm.org/docs/LinkTimeOptimization.html` — Full LTO vs ThinLTO semantics, cross-TU visibility (HIGH, fetched 2026-09-08)
- `clang.llvm.org/docs/ControlFlowIntegrity.html` — CFI/KCFI requirements and LTO dependency (HIGH, fetched 2026-09-08)
- `clang.llvm.org/docs/AddressSanitizer.html`, `clang.llvm.org/docs/MemorySanitizer.html` — use-after-return flag, MSan interceptor/instrumentation requirements (HIGH, fetched 2026-09-08)
- `pkg.go.dev/golang.org/x/perf`, `pkg.go.dev/golang.org/x/perf/cmd/benchstat` — benchstat statistical methodology, install path (HIGH, fetched 2026-09-08)
- Rust NLL/polonius interprocedural design stance — cited as well-established Rust-compiler-team position (HIGH by broad community consensus, not re-fetched from rust-lang primary sources in this pass — flag for verification if this claim becomes load-bearing for a specific design decision)
- Local environment: `go version` and `clang --version` run directly (HIGH, first-party, 2026-09-08)
- `.planning/PROJECT.md`, `.planning/RETROSPECTIVE.md` — M001 baseline stack, constraints, and lessons (HIGH, first-party project documents)

---
*Stack research for: Codename Lang M002 — Interprocedural Semantic Spine*
*Researched: 2026-09-08*
