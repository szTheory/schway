---
id: convergence-work-program
title: Experimental frontier and convergence work program
summary: The decision matrices, experiments, dependencies, and exit criteria for syntax, ownership ergonomics, native backend and runtime selection, dogfooding, and measured performance.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [convergence, syntax, ownership, backend, runtime, performance, dogfood]
related: [design-baseline, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, trust-validation-information-flow, boundary-data-validation-persistence, residual-uncertainty-register, convergence-audit, syntax-cognitive-ergonomics, native-low-level-profile, runtime-profiles-dogfooding, memory-reclamation-policy, networking-http-tls-security, compiler-feedback-latency, compute-efficiency-constitution, capability-gauntlet, implementation-path, open-questions, research-ledger]
---

# Experimental frontier and convergence work program

## Reader and outcome

This note is for the language-design and bootstrap team. After reading it, they
should be able to run the remaining experiments in dependency order, understand
which result changes which decision, and know when a frontier is sufficiently
settled to implement without pretending that every detail is final.

## Why the first proof is not selected yet

No first vertical is selected. The earlier streaming JSON cleaning CLI remains
a useful workload because it exercises bytes, Unicode, validation, streaming,
ownership, borrowed views, arenas, lexical resources, fallible allocation, C
interop, profiling, and data lineage in a small executable. The leading wrapper
is now a Lang-written corpus/evidence runner that uses this workload while also
accelerating conformance, performance, cache, and AI-feedback work.

Before selecting that runner—or another workload—we need to establish enough of
the following five frontiers so that the workload measures the intended
language rather than an accidental prototype:

```text
defined semantic kernel
    |
    +--> ownership/resource candidates --> syntax forms
    |             |
    |             +---------------> IR obligations
    |                                  |
    +--> evaluation corpus ----------> backend/runtime contest
                                           |
compute constitution --------------------> measured decision
                                           |
toolchain protocol ----------------------> compounding dogfood
```

We do not need to finish all language design first. We need candidate semantics,
two or three credible surfaces, a backend-neutral IR contract, and measurements
that can falsify each choice.

## Review every frontier through the same lenses

| Lens | Required question |
|---|---|
| language/type theory | Is the rule coherent, compositional, decidable, and locally explainable? |
| compiler implementation | What analysis, invalidation, lowering, diagnostic, and bootstrap burden does it add? |
| runtime/systems | What allocation, scheduling, synchronization, ABI, and failure machinery does it require? |
| AI generation/repair | Does it reduce ambiguity and repair rounds under malformed and nonlocal changes? |
| human audit | Can a reviewer locate authority, cost, ownership, failure, and state transition quickly? |
| software architecture | Does it preserve explicit boundaries without forcing one application methodology? |
| security/privacy | Can untrusted data or ambient authority exploit the mechanism or its diagnostics/cache? |
| performance/cost | What are the time, space, I/O, artifact, telemetry, and model-compute costs? |
| operations/evolution | Can it be observed, upgraded, rolled back, recovered, and run under version skew? |
| ecosystem/adoption | Can it interoperate, package, document, and migrate without a deep dependency burden? |

An option does not win by dominating one lens. Record the strongest objection
from every lens, the evidence that would change the recommendation, and the
migration seam if the recommendation later loses.

## Polished convergence register

| Area | Leading recommendation | Status | What can still overturn it |
|---|---|---|---|
| semantic center | values, ADTs, results, effects/capabilities, ownership/resources, tasks, components, evidence | stable enough to prototype | executable contradiction across two implementations |
| effects and DI | explicit public effect rows with local inference; statically resolved providers and lexical overrides | strong candidate | service corpus shows signature or resolution cost exceeds architecture/testing value |
| ownership behavior | value/access surface over affine places/loans/origins; local last use; verified value-path public origins; distinct abilities/obligations; targeted fresh callback scopes | executable strong candidate | native, async, variance, or 64-workload evidence produces unsoundness, compile cliffs, or excessive ceremony/copies |
| independent evidence | compact content-bound typed-core manifest with source/body-blind recomputation at trust crossings; full traces on mismatch | validated for bounded first-order ownership facts | schema growth recreates a compiler, or native evidence shows the checked facts are insufficient |
| ordinary storage | accepted unique/arena foundation with no mandatory global heap; compare precise RC/reuse and one isolated managed-region implementation | boundary accepted; representation open | graph, actor, UI, compiler, and release-tail workloads determine representation |
| syntax | braced control, indentation challenger, uniform-tree machine projection | open experiment | generation/recovery/audit/diff corpus |
| modules/architecture | private explicit modules and component DAG; verified modular-monolith preset | strong candidate | projects need frequent honest cycles or policy workarounds |
| native backend | interpreter oracle; Cranelift candidate, QBE control, LLVM later | open experiment | lowering, latency, runtime, debug, platform, and maintenance measurements |
| runtime | monotonic freestanding/native-core/application/service profiles | strong candidate | profile leakage or runtime ABI fragmentation |
| scheduling | safe-point budgets first; bounded blocking executors; async preemption only if needed | open experiment | fairness/tail-latency adversaries defeat safe points |
| caching | core semantic hooks plus first-party typed policy, no dedicated v1 grammar | strong candidate | repeated unsafe ceremony cannot be derived cleanly |
| performance | budgeted abundance with whole-loop, multi-resource evidence | stable governance | numeric thresholds await machines and workloads |
| dogfood | corpus/evidence runner before generic proof or compiler self-host | leading candidate | it fails to shorten later verified work |
| assurance gates | fast sound `check`, stronger `verify`/`release` | direction stable | exact blocking threshold remains author policy |
| trust/information flow | typed safe sinks, nominal validation, IR flow summaries, targeted stronger analysis; no universal runtime taint bit | strong candidate | boundary corpus shows unacceptable noise, gaps, or analysis cost |
| source authority | canonical text plus semantic graph tooling | unresolved value choice | author may choose semantic database authority later |

This register deliberately leaves algorithmic and ergonomic questions to
experiments while allowing implementation planning around stable seams.

## Frontier A — syntax and source ergonomics

### Recommendation to test

Use the existing conventional braced surface as the control, an indentation
surface as the human challenger, and a uniform tree only as the machine/debug
projection. Keep grammar context-free or close to it, reserve keywords, forbid
user-defined precedence in v1, and make formatting canonical and editioned.

Clarity at the use site should dominate raw character count. Swift's official
API guidelines explicitly prioritize clarity over brevity and use argument
labels when roles would otherwise be unclear. ([Swift API guidelines](https://www.swift.org/documentation/api-design-guidelines/))
Rust's formatting RFC demonstrates a stable default style with edition-aware
evolution instead of per-project dialects. ([Rust style evolution](https://rust-lang.github.io/rfcs/3338-style-evolution.html))

### Decision matrix

| Decision | Leading candidate | Serious alternative | Footgun to test | Decisive evidence |
|---|---|---|---|---|
| Grouping | Braces with formatter-controlled vertical rhythm | Significant indentation | Incomplete edits attach a block/call incorrectly | Error-recovery corpus and repair success |
| Statement termination | Newline/grammar, no optional semicolon dialect | Mandatory terminator | Automatic insertion surprises and diff noise | Incomplete/multiline parse corpus |
| Calls | Parenthesized calls; named roles for multi-role APIs | Whitespace/keyword calls | Call/declaration ambiguity | Parser recovery plus audit questions |
| Argument labels | Required when roles are distinguishable | Positional unary/algebraic exceptions | Labels repeat type/name without meaning | Use-site comprehension and token/repair cost |
| Exact-name punning | Formatter rewrites `x: x` to `x:` | Always expanded | Shadowing or qualification changes meaning | Rename/shadow corpus across models |
| Pipelines | One left-to-right operator over ordinary calls | Method chains or placeholder pipes | Hidden argument position and effect order | Data/service pipeline comprehension |
| Effect visibility | Declared `with { ... }`, plain call | Effect marker at call site | Effects hidden versus punctuation saturation | Effect-location tasks and diff churn |
| Ownership visibility | Public declaration modes plus visible named-value transfer/exclusive access | Declaration-only or ambiguity-only markers | Invisible moves/copies or Rust-like annotation spread | Ownership repair/audit tasks and copy profiles |
| Mutation | `var`/`mut` is explicit, narrow, exclusive | Pure update only in portable core | Aliasing or update boilerplate | Parser, buffer, state-machine corpus |
| Pattern matching | Exhaustive `match`, pure guards, compact payload punning | Visitor/method dispatch | Dense nested destructuring becomes visual noise | Failure/state-machine changes |
| Failure | `Result` plus postfix propagation operator | Keyword propagation | `?` hard to scan beside optionality | Error-path audit and edit corpus |
| Resources | `use resource = acquire()?` lexical binding | `with resource` block | Cleanup order or fallible close is hidden | Multi-resource/error/cancellation corpus |
| Modules | Explicit module/import/export blocks | File-implied modules | Renames and re-exports obscure public boundary | Semantic diff and cycle diagnostics |
| Generics/constraints | Explicit public bounds; local inference | Trait/implicit-heavy surface | Resolution errors become distant | Generic port/collection corpus |
| Literals | Few typed literal families with fixed grammar | Builders only | Sigil zoo and version-sensitive embedded DSLs | JSON/time/regex/bytes/SQL corpus |
| Lambdas | One explicit compact form | Placeholder shorthand | Nested placeholder numbering becomes write-only | Async/pipeline/callback corpus |
| Concurrency | Structured `task`/actor declarations | Library calls only | Detached work and hidden lifetime | Cancellation/supervision corpus |
| Specs | Flat ordinary-code `spec`; visible arrange/act/assert | Nested BDD DSL | Setup precedence and metaprogrammed tests | Mutation score and maintainer comprehension |
| Comments/docs | Stable attachment to semantic node | Pure byte-position comments | Formatter or move loses intent | Round-trip/move/semantic-edit corpus |
| Identifiers | ASCII-normalized identifiers initially; full Unicode text | Restricted Unicode scripts | Confusables and normalization instability | Security corpus and international review |
| Keywords | Small reserved set with contextual keywords only when recovery is strong | Minimal Lisp-like grammar | Contextual parsing harms diagnostics | Parser complexity and malformed corpus |
| Metaprogramming | Typed derivations/projections, no token macros v1 | Hygienic AST macros | Compile-time execution and dialect growth | Required toolkit implementations |

### Syntax anti-patterns

- Optimizing token count independently of repair rounds and semantic payload.
- Adding aliases or optional punctuation for “poetry” after choosing canonical
  source.
- User-defined operators, precedence, implicit conversions, or extension
  namespaces that expand the parse/name-resolution possibility space.
- Making effect, ownership, or failure punctuation carry several unrelated
  meanings.
- Evaluating only greenfield snippets rather than malformed edits, diffs,
  migrations, error messages, and large files.
- Letting the formatter resolve semantically ambiguous source.

### Syntax experiment

Render 25–30 identical semantic programs through the three surfaces. Seed each
with missing delimiters, incomplete matches, ownership errors, wrong providers,
argument reorder/addition, renames, merge conflicts, and AI-generated near
misses. Measure parse recovery, diagnostic locality, first-pass correctness,
repair rounds/tokens, diff size, semantic-edit stability, and timed human audit.
Tree-sitter demonstrates that per-keystroke incremental error-tolerant parsing
is feasible, but our grammar still has to earn that behavior. ([Tree-sitter](https://tree-sitter.github.io/tree-sitter/))

Freeze a v0 syntax only when one surface wins the declared whole-loop rubric;
retain the others as projections only if they round-trip the same semantic tree.

## Frontier B — ownership and resource ergonomics

### Recommendation to test

Separate five concerns that are often conflated:

1. **Value ownership:** who may consume, mutate, retain, or share a value.
2. **Storage strategy:** stack, inline, arena, unique heap, reference counted,
   managed region, static, or foreign storage.
3. **Resource protocol:** which values require exactly-once close/release and
   which operations such as commit/flush may fail.
4. **Reference validity:** which origin backs a view and where it may escape.
5. **Concurrency isolation:** what may move or be shared across owners.

The language-level candidate is a value/access surface over affine ownership:
owned immutable values, duplication and abandonment modeled separately, local
lexical borrow inference, explicit public origin/mode summaries, visible
consuming or exclusive access at semantic cliffs, noncopyable resources and
must-resolve obligations, and providers for allocation/storage policy. Storage
implementation must not leak into every domain signature. The complete design
and falsification protocol are in the [ownership and lifetime decisive
study](ownership-and-lifetime-decisive-study.md).

Swift now exposes `borrowing` and `consuming` parameters to control ownership
overhead while keeping ordinary call syntax unchanged, and requires explicit
copying within ownership-aware parameters. ([Swift declarations](https://docs.swift.org/swift-book/documentation/the-swift-programming-language/declarations/))
Cyclone showed that defaults and local inference can reduce region annotations,
but separate compilation still forces some boundary annotations. ([Cyclone
regions](https://www.cs.cornell.edu/Projects/cyclone/papers/cyclone-regions.pdf))

### Decision matrix

| Decision | Leading candidate | Main alternatives | Failure mode | Required probe |
|---|---|---|---|---|
| Default binding | Immutable owned value | Borrowed-by-default, implicit shared reference | Hidden lifetime or excessive copying | Collections/domain corpus |
| Duplicability | Type declares a duplicable ability; logical clone remains explicit | Everything copyable with optimizer elision | Semantic copy cost varies after refactor | Large buffer/string/record probes |
| Consumption | Public `take` mode plus visible transfer of a named noncopyable binding | Declaration-only or ambiguity-only marker | Use-after-move confusion, hidden clone, or sigil noise | Branch/loop/match/callback corpus |
| Shared immutable data | Candidate precise RC/reuse, interned/static, or managed provider | Universal tracing GC | Atomic RC traffic, cycles, pauses, runtime size | Trees/graphs/actors/data pipeline |
| Cycles | Disallow accidental strong cycles; explicit weak/managed graph | Cycle collector by default | Leaks or dual memory model | UI graph, actor registry, compiler graph |
| Borrow inference | Lexical, intra-procedural, non-escaping first | Whole-program lifetime inference | Distant diagnostics and compiler cost | Edit/invalidation benchmarks |
| Mutable borrow | One exclusive lexical capability | Interior mutability/locks by default | Alias races or wrapper proliferation | Parser, arena, cache, iterator |
| Returned borrow | Verified parameter/field origin path or tagged alternative; owning view when retention is required | Explicit region algebra everywhere or callback-only APIs | Fragile APIs, conservative union retention, or erased lifetime facts | Slices, iterators, parsers |
| Partial moves | Initially restrict or require explicit destructuring | Full field-sensitive moves | Complex drop flags and error messages | Records, variants, cleanup paths |
| Destruction order | Lexical reverse order, specified | Optimizer/runtime chosen | Lock/resource order surprises | Multiple dependent resources |
| Destructor power | compiler-known, non-failing structural release only | user-defined/throwing/async destructor | double failure, hidden authority, and latency spikes | deep value/resource probes |
| Fallible close | Explicit `commit`/`close` before automatic release fallback | Destructor reports errors | Lost durability errors | Disk-full and cancellation tests |
| Async resource | Structured `use async`/scope completion | Arbitrary async destructor | Cancellation deadlock or forgotten await | TLS/session/stream tests |
| Closures | Capture mode inferred and renderable; explicit on escape | Implicit shared capture | Lifetime extension and cycles | Tasks/callbacks/GUI handlers |
| Actor/task transfer | Unique move or immutable share; copy must be visible | Universal serialization/copy | Hidden bandwidth/allocation | Large messages and cancellation |
| FFI ownership | Generated candidate contract plus explicit audited adapter | Raw pointer signature only | Foreign retention/alias/unwind unsoundness | C callbacks and buffer ownership |
| Pin/address sensitivity | Stable-storage type at explicit boundary | Surface pinning in ordinary async code | Self-reference/DMA/intrusive unsoundness | Native future and intrusive list |
| OOM | Fallible allocator effect in reusable/native code; outer profile policy may abort | Universal exception or abort | Library cannot operate under budgets | Arena exhaustion and container limit |
| Unsafe scope | Module/component with transitive ledger and evidence | Lexical block only | Safe API rests on invisible claims | C parser/SIMD/atomic component |

### Storage strategy candidates

| Strategy | Strength | Cost/risk | Provisional placement |
|---|---|---|---|
| Unique ownership | deterministic, no runtime collector, strong mutation/transfer | graph sharing and ergonomics | universal resource/native foundation |
| Arenas/regions | fast allocation/free, locality, bounded lifetime | retention and lifetime granularity | explicit provider; compiler/pipeline/game hot paths |
| Precise RC with reuse | functional surface, prompt reclamation, possible in-place reuse | cycle handling and concurrent atomic overhead | serious ordinary immutable-value candidate |
| Tracing GC | cycles and graph-heavy application ergonomics | pauses/runtime/FFI/pinning/memory headroom | isolated managed application/service region, never a global kernel requirement |
| ARC/ordinary RC | deterministic reclamation and simple interop model | retain/release traffic and cycles | compare against precise RC/reuse |
| Manual/static | maximal control and no runtime | highest proof/unsafe burden | restricted embedded/unsafe boundary |

Perceus provides primary evidence that precise reference counting plus reuse can
support functional-but-in-place execution for cycle-free programs; it does not
establish that this is the right concurrent general-purpose default. ([Perceus](https://doi.org/10.1145/3453483.3454032))

### Ownership footguns to design out

- lifetime annotations whose meaning depends on compiler-internal region names;
- implicit expensive copies triggered by an unrelated later use;
- drop order that changes across syntactic refactoring;
- fallible work hidden in a destructor;
- unbounded RC cascades or arena reset work hidden on a latency-critical path;
- shared ownership cycles treated as a library footnote;
- holding borrows/locks across suspension without a stable protocol;
- callbacks that outlive foreign/library state;
- partial initialization/destruction paths implemented after the happy path;
- “unsafe” suppressing diagnostics instead of naming obligations;
- different overflow/bounds/cleanup semantics in optimized builds.

### Ownership experiment

Run the 64-workload laboratory across explicit-affine,
boundary-explicit-value-access, inference-heavy, and projection-constrained
checker/surface configurations. Cross the relevant workloads with unique/arena,
precise-RC/reuse, and owner-confined-managed storage. The suites cover local
control flow; views/collections/generics; resources/failure/suspension;
concurrency/graphs; native/FFI/representation; and whole-program ergonomics and
evolution.

Measure annotation count, repair rounds, diagnostic distance, copies, retains,
allocations, peak memory, release-queue depth, p99.9/max stalls, compile/query
work, runtime, unsafe surface, API migration, and source-only audit accuracy.
Apply the hard vetoes before comparing weighted outcomes. See the [ownership
and lifetime decisive study](ownership-and-lifetime-decisive-study.md) for the
complete protocol.

The reference interpreter must enforce moves, borrows, resource state, and drop
order dynamically so it can serve as a differential oracle for native lowering.

## Frontier C — backend and execution architecture

### Recommendation to test

Do not choose one engine for every lane. Specify one backend-neutral typed core
IR and compare this stack:

```text
reference interpreter  -- semantics, const/test/live evaluation
Cranelift candidate    -- fast native dev build and initial release
QBE control            -- very small AOT/C-ABI baseline
LLVM later candidate   -- optimized/HPC/GPU breadth when measured
Wasm/BEAM later        -- isolation/component and service conformance profiles
```

The interpreter is mandatory as a semantics oracle, not necessarily as a
performance tier. Bytecode earns admission only if it materially improves live
execution/startup/portability enough to justify another IR and verifier.

### Backend matrix

| Backend/path | Strength | Risk/limitation | Question it answers |
|---|---|---|---|
| Direct AST/HIR interpreter | simplest executable semantics, rich source mapping, deterministic tests | slow; may diverge from lowered semantics | Are language rules coherent and observable? |
| Compact bytecode VM | fast incremental load, portable live workspace, controlled runtime | another IR/runtime/GC and security surface | Does repeated interactive execution justify it? |
| Cranelift | library embedding, JIT/AOT, fast codegen, defined IR, production Wasmtime use | fewer targets/optimizations than LLVM; Rust dependency | Best v1 compile/run/value balance? |
| QBE | very small hackable backend, full C ABI, simple SSA, fast AOT | limited targets/features; process/text IL integration and optimization ceiling | How much backend is actually needed? |
| LLVM | target/optimizer/debug/sanitizer/PGO ecosystem and high ceiling | large dependency, compile/memory cost, subtle UB/IR contracts | Which proven workloads need its ceiling? |
| libgccjit | embedded GCC with JIT and AOT C APIs | documented experimental status and packaging/ecosystem costs | Useful fallback or comparison? |
| C source lowering | portability/bootstrap/debug familiarity | slow downstream C compile, semantic/ABI impedance, poor incremental control | Bootstrap escape, not default? |
| Custom native backend | exact semantics/control and potential minimal latency | enormous correctness/platform/debug burden | Only after external backends demonstrably block goals |

Cranelift states that it targets x86-64, AArch64, s390x, and riscv64, supports
JIT/AOT embedding, avoids undefined behavior in its IR, and trades aggressive
optimization for compiler speed and simplicity. ([Cranelift](https://cranelift.dev/))
QBE deliberately targets a much smaller scope and currently lists amd64,
arm64, and riscv64 with full C ABI support. ([QBE](https://c9x.me/compile/))
LLVM ORC supplies eager/lazy/concurrent JIT composition and is used for REPL and
debugger expression evaluation, but bringing LLVM into the edit lane remains a
measurement question. ([LLVM ORC](https://llvm.org/docs/ORCv2.html))

### IR/backend obligations

- Define integer, floating-point, overflow, bounds, initialization, provenance,
  aliasing, layout, atomics, panic/unwind, cancellation, effect order, and drop
  behavior independently of the backend.
- Validate every IR stage and retain stable semantic/source identities.
- Differentially execute interpreter and native outputs across optimization
  modes, targets, random programs, and adversarial reducers.
- Never attach a backend `noalias`, alignment, range, or non-null fact unless the
  frontend can prove its stronger contract.
- Keep development codegen per function/module where practical; avoid paying
  whole-program optimization/link cost for one edit.
- Emit usable debug/unwind/profile metadata and optimization remarks before
  calling a backend supported.
- Count code size, relocation/link, debug info, backend memory, and artifact
  caching—not codegen wall time alone.
- Pin and update backend versions deliberately; conformance must run across an
  upgrade before adoption.

### Backend experiment

Lower a deliberately small portable core—integers, floats, records, tagged
variants, calls, loops, checked indexing, results, owned buffers, lexical drop,
and C calls—to the interpreter, Cranelift, and QBE. LLVM is added only after the
semantic/lowering harness works. Measure:

- cold tool bootstrap and clean build;
- warm one-function edit to runnable artifact;
- frontend, codegen, link, and startup separately;
- CPU, peak memory, disk/cache, output/debug-info size;
- runtime across parser, allocation, branch, call, numeric, and C-boundary
  workloads;
- target/ABI/debugger/profiler/sanitizer coverage;
- implementation code and maintenance complexity;
- diagnostic/source-map quality;
- miscompile/conformance failures and reducer quality.

The winner may differ by lane. Selecting Cranelift for development does not
commit the language to Cranelift semantics or prohibit LLVM for specialized
release targets.

## Frontier D — measured performance and efficiency

### Recommendation to test

Adopt the [compute-efficiency constitution](compute-efficiency-constitution.md)
before implementing the compiler. Benchmark work performed, not only elapsed
time, and treat every feature as a cost across edit, clean build, CI, runtime,
telemetry, agent repair, and maintenance lanes.

### Measurement matrix

| Surface | Workloads | Required metrics | Common trap |
|---|---|---|---|
| Lexer/parser | clean, small edit, malformed/incomplete, huge/adversarial file | latency, bytes scanned, allocations, retained tree, recovery quality | fast valid parse with pathological error recovery |
| Name/type/effect/ownership | private body, public type/effect/layout, generic/provider change | queries executed, invalidation cone, wall/CPU/memory, diagnostics | headline warm result hides global edit |
| Dev codegen/link | one function/module, C boundary, debug build | codegen/link/startup, backend RSS, object/cache bytes | compiler time excludes linker/debug info |
| Clean build | 10K/100K/1M synthetic plus real corpus | critical path, work counts, CPU, memory, disk, parallel efficiency | synthetic modules lack real generic/effect shapes |
| Tests/evidence | success/failure, affected/full, fuzz/property/model | time to first useful failure, selected/omitted proof, flake, compute | minimum wall time by wasteful sharding |
| AI repair | seeded semantic defects and changes | first-pass success, rounds, tokens, tool calls, wall, hidden correctness | shortest syntax wins despite more retries |
| Runtime | CLI, parser, service, actor, stream, allocator, C call | throughput, distributions, CPU, RSS/peak, allocation/copy/RC, startup, energy | average latency and happy-path input only |
| Operations | telemetry off/default/focused/degraded | overhead, bytes/events/cardinality, dropped data, diagnosis time | “observability on” destabilizes workload |
| Build/cache/CI | local/remote hit/miss/stale, parallel worktrees | lookup/hash/serialize/upload/download, hit value, cache size, correctness | hit rate treated as value |

### Benchmark corpus shape

Use four layers:

1. **Micro:** lexer token, hash, lookup, allocation, call, effect dispatch,
   pattern match, retain/release, bounds check, C call.
2. **Meso:** parser, type-check module, JSON decode, tree transform, arena build,
   actor mailbox, stream stage, query invalidation.
3. **Macro:** compiler build, streaming data CLI, modular service, emulator or
   renderer slice.
4. **Journey:** AI makes a seeded change, receives evidence, repairs it, builds,
   tests, deploys a canary, and explains a failure.

Every result keeps raw samples and artifact/toolchain/workload identities.
Compare distributions and effect sizes, not isolated means. Low-noise machines,
repeated runs, and representative inputs are necessary; they do not eliminate
benchmark bias. ([LLVM benchmarking guidance](https://llvm.org/docs/Benchmarking.html))

### Performance governance

- Establish baselines before optimization work.
- Run tiny stable budget gates on PRs; larger/noisier suites report or run on
  dedicated machines/nightly.
- Use a change-point/regression triage process rather than a brittle universal
  percentage threshold.
- Profile compiler queries/passes/allocations before patching. Rust's compiler
  project uses self-profiling and continuous per-change benchmark
  infrastructure. ([rustc profiling](https://rustc-dev-guide.rust-lang.org/profiling.html),
  [rustc-perf](https://github.com/rust-lang/rustc-perf))
- Record deliberate regressions as a benefit/cost decision, not an exception
  without memory.
- Optimize compiler and generated programs independently; a backend may trade
  one for the other.
- Track complexity cliffs with adversarial size series, not one input size.

## Frontier E — runtime boundary and compounding dogfood

### Recommendation to test

Compile ahead of time for production, but use explicit monotonic runtime
profiles rather than equating compilation with either a mandatory VM or no
runtime services. The leading ladder is `freestanding` -> `native-core` ->
`application` -> `service`, with managed actor/regions opt-in only if graph-heavy
workloads justify them. See [Runtime profiles and
dogfooding](runtime-profiles-and-dogfooding.md).

| Decision | Leading candidate | Serious alternative | Decisive evidence |
|---|---|---|---|
| production execution | native AOT | VM/JIT default | startup, footprint, C ABI, runtime, deployment |
| sequential runtime | linked-on-demand small support | universal scheduler/collector | unused-service binary/startup/runtime cost |
| service scheduling | compiler safe points and bounded task budgets | arbitrary async preemption or voluntary yields | fairness, tail latency, overhead, FFI behavior |
| managed memory | no global heap; explicit actor/region experiment | one universal tracing heap or no managed graphs anywhere | graph ergonomics, release/collection tails, memory, cross-region cost |
| BEAM role | operational reference and later conformance target | first implementation backend | native/resource proof plus service semantic fit |
| first Lang tool | corpus/evidence runner | generic JSON cleaner or compiler frontend | development-loop time saved per implementation variable |
| self-hosting | progressive leaf tools, compiler late | immediate full compiler rewrite | bootstrap latency, repairability, reproducibility |
| host language | Go 1.24 selected for dependency-free Stage 0; Rust and OCaml remain later controls | select permanently by familiarity | frontend build loop, dependency surface, backend integration; preserve replacement seam |

The runtime is admitted service by service. A sequential CLI does not link an
actor scheduler or tracing collector. A service does not ask every library to
bring a competing executor. Foreign CPU and I/O work declares its scheduler
class and cannot silently monopolize ordinary work.

Dogfood starts with executable corpus programs, then a Lang-written
corpus/evidence runner, diagnostic/cache clients, formatter/build driver, and
selected compiler passes. Full self-hosting waits for reproducible staged builds
and a host fallback oracle.

## Workload candidates after frontier experiments

| First G3 candidate | Semantic coverage | Bootstrap cost | Blind spot |
|---|---:|---:|---|
| Lang corpus/evidence runner | very high native/resource/toolchain/cache coverage | medium | compiler protocol must exist first |
| Streaming JSON cleaning CLI | very high native/resource/data coverage | low-medium | weak compounding/toolchain and distributed/DI proof |
| Small HTTP/1.1 server | strong native I/O/parser/backpressure coverage | medium-high | TLS/OS details can swamp language questions |
| Candidate compiler/formatter itself | strongest AI/tooling feedback and dogfood | high | circular bootstrap and weak foreign/user workload |
| Image/audio asset processor | strong data layout/SIMD/C/asset coverage | medium | weak service/effect-provider breadth |
| Embedded KV slice | strongest storage/crash/concurrency pressure | very high | too many hard variables for first proof |
| Modular HTTP/SQL application | strongest effect/DI/architecture/operations proof | high | native memory semantics are less directly stressed |

The Lang corpus/evidence runner is now the leading recommendation because it
contains the useful JSON/data stress while compounding the project's own
conformance, benchmarking, provenance, cache, and AI-feedback capabilities. It
should not become a roadmap commitment until the ownership and backend
mini-spikes establish its prerequisites. The generic JSON CLI remains a
low-cost workload inside the runner; the HTTP/SQL application remains the
recommended next production-style proof.

## Dependency-ordered convergence sequence

### Experiment 0 — semantic contract

Implement the [semantic kernel contract](semantic-kernel-contract.md) and its
[probe suite](semantic-kernel-probes.md). Begin with an explicit kernel notation,
typed core, deterministic interpreter, stable diagnostics, and lossless evidence
schema. Make the interpreter enforce moved/released states, loans,
initialization, origins, and resource obligations dynamically. No backend or
pretty syntax owns these semantics.

The design work is complete enough to start. Exit requires an executable
ownership-aware core over all blocking probes, not more prose.

Five ownership workbenches have begun this gate: linear
moves/loans/releases, scopes, whole-value flow, CFG edge-specific last use, and
separately compiled public origin/access/ability summaries plus a compact
independent typed-core validator now execute with differential and mutation
evidence. The [public-summary
comparison](../.planning/spikes/003-public-origins-generic-abilities/README.md)
selects value paths plus targeted higher-ranked callback origins semantically,
without freezing source grammar. The [certificate
comparison](../.planning/spikes/004-independent-certificate-checker/README.md)
retains compact boundary recomputation but rejects any claim that it establishes
source-to-core truth. The [native contract
harness](../.planning/spikes/005-native-ffi-provenance-cleanup/README.md) now
pins ABI, cleanup, allocator, retention, no-alias, sanitizer, and nonlocal-exit
failures under a real C/Clang toolchain. Its partial verdict means the next
artifact must be the real frontend/core IR that produces this lowering; native
translation from Lang, variance, async, and the rest of the blocking probes
remain before Experiment 0 exits.

### Experiment 1 — ownership laboratory

Run the 64 ownership workloads across the candidate checker/surface and storage
models in dependency order. Begin with local values, views/generics, and
resources/suspension in the interpreter; then add one minimal native lowering
and the blocking native/FFI cases. Complete concurrency/graph and evolution
cases before selecting the contract.

Eliminate any model that cannot explain its decision locally, preserve
deterministic cleanup, express the native/FFI corpus safely, support separate
compilation, or meet a deliberately loose compiler/runtime budget. Exit
requires interpreter/native agreement for the blocking native subset plus a
recorded disposition for every workload.

### Experiment 1b — trust-flow skeleton

Preserve source, transformation, validation, encoding, sink, endorsement, and
declassification summaries in typed IR. Implement the 18 `FLOW-*` probes using
nominal safe sinks and local compositional checking. Prototype targeted global
and control-influence analysis as a separate budgeted query; do not add a
Boolean runtime taint bit or a full security lattice to the ordinary type
checker.

Exit requires useful source-to-sink diagnostics, conservative foreign-summary
behavior, and measured incremental cost. The complete candidate is in [Trust,
validation, and information flow](trust-validation-and-information-flow.md).

Exercise the same hooks with the bounded JSON/change/transaction/pool fixtures
from [Boundary data, validation, transactions, and
pools](boundary-data-validation-and-persistence.md). These are client programs
and kit contracts, not a new kernel experiment or a reason to delay native
lowering.

### Experiment 2 — syntax corpus

Render the surviving ownership/effect model through three surfaces and test
generation, recovery, repair, audit, and diffs. Select syntax from whole-loop
evidence.

### Experiment 3 — native lowering and runtime seam

Lower the portable kernel to interpreter, Cranelift, and QBE; add LLVM after the
harness is trustworthy. Compile one program without scheduler/GC services and
one with safe-point scheduling, then measure the incremental cost. Select
development and initial release paths separately.

### Experiment 4 — first vertical selection

Score the workload candidates using measured semantic coverage,
implementation variables, AI/human loop value, and adopter relevance. Promote
one to G3 with explicit acceptance thresholds. Prefer the corpus/evidence
runner only if it measurably shortens subsequent Lang development.

### Experiment 5 — growth and adversarial scale

Scale the selected design across module counts, public edits, generic/effect
shapes, malformed input, large data, concurrent worktrees, cold caches, and
resource pressure. Change query/data structures based on measured work.

## Exit criteria

The five frontiers are sufficiently converged for a v0 implementation plan
when:

1. an executable semantic contract covers the ownership/resource failure cases;
2. the typed IR preserves trust-flow summaries and the 18 flow probes pass
   without imposing ordinary production runtime metadata;
3. one ownership model wins the probe rubric with explicit unresolved edges;
4. one human syntax and one machine projection round-trip the corpus;
5. development and initial-release backend choices meet numeric budgets on the
   same portable kernel;
6. benchmark and AI-repair harnesses retain raw reproducible evidence;
7. the runtime profile seam proves that unused scheduler/GC services are not
   imposed on sequential native programs;
8. the first G3 workload and its success thresholds are selected;
9. every deferred choice has a trigger and no deferred choice can invalidate
   the first IR.

No choice is permanent. It is stable enough to build when alternatives have
been fairly represented, the winner has falsifiable evidence, migrations remain
possible, and further analysis has lower expected value than implementation.
