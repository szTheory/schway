---
id: compute-efficiency-constitution
title: Compute-efficiency constitution
summary: Project-wide rules for spending compute deliberately across compiler development, AI iteration, CI, builds, runtime, telemetry, and evidence.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-02
tags: [performance, compiler, ci, ai, mechanical-sympathy]
related: [vision, author-intent-model, compiler-feedback-latency, performance-observability-delivery, convergence-work-program, implementation-path, research-ledger]
---

# Compute-efficiency constitution

## Reader and outcome

This note is for every language, compiler, runtime, toolkit, CI, and AI-system
contributor. After reading it, they should be able to design a feature, measure
its full resource cost, choose the cheapest sound feedback lane, and recognize
when spending more compute is justified.

## Doctrine: budgeted abundance

Compute is not intrinsically scarce and minimizing every CPU cycle is not the
goal. **Waste is unacceptable; valuable computation is welcome.** Spend more
when it materially improves correctness, uncertainty reduction, security,
runtime quality, or human/agent feedback. Avoid work that is duplicate,
mis-scoped, unactionable, prematurely optimized, or cheaper to derive from
already-known facts.

The project optimizes total cost from intent to verified, operated software:

```text
human wait + agent/model time + repair rounds
  + local/CI CPU + memory + disk + network
  + build/runtime/telemetry resources
  + defect, incident, migration, and maintenance cost
```

A slower analysis that prevents hours of repair can be efficient. A cache hit
that downloads more bytes than recomputation is not. A fast compiler producing
opaque diagnostics can lengthen the real loop.

## Optimization order

Apply these in order; later techniques must not conceal an earlier design
failure:

1. **Eliminate work.** Remove unnecessary passes, representations, dependencies,
   generated code, repeated context, and duplicate evidence.
2. **Narrow work soundly.** Use semantic dependency/impact graphs and explicit
   interfaces to check only what may be affected; widen when uncertain.
3. **Choose the right algorithm and representation.** Bound complexity; use
   compact IDs, interning, bitsets, arenas, locality-friendly layouts, streaming,
   batching, and bounded queues where the workload warrants them.
4. **Schedule work intelligently.** Parallelize independent work within CPU,
   memory, I/O, and thermal budgets; cancel superseded work; prioritize the
   answer currently blocking a human or agent.
5. **Reuse work economically.** Cache deterministic expensive results only when
   lookup, validation, storage, transfer, and invalidation cost less than
   recomputation.
6. **Improve generated/runtime code.** Specialize, vectorize, PGO, or use a
   heavier backend only after measured hot paths and compile-time costs justify
   them.

## Cost dimensions

Every benchmark or budget declares the dimensions it includes:

| Dimension | Typical failure hidden by one headline number |
|---|---|
| Wall latency | Parallel work consumes excessive CPU or blocks other interactive work. |
| CPU time | I/O, queueing, or remote execution dominates elapsed time. |
| Peak/resident memory | Fast workers multiply until the machine swaps or CI is killed. |
| Allocations and bytes moved | Throughput looks adequate until concurrency or data size grows. |
| Disk reads/writes and cache size | Incremental speed consumes unbounded local/CI storage. |
| Network bytes and requests | Remote cache or dependency fetch costs more than local work. |
| Artifact/debug-info size | Fast compilation produces slow linking, upload, startup, or deployment. |
| Energy/thermal behavior | Laptop iteration or embedded/runtime performance degrades under sustained work. |
| Model tokens/tool calls | Concise source still requires excessive context and repair iterations. |
| Human interruption/wait | Asynchronous throughput masks a poor interactive experience. |
| Correctness/evidence coverage | Speed came from skipping a soundness or policy obligation. |

## Budget manifest

Projects and the compiler itself use the same versioned policy model:

```text
budget development {
  parse_query.p95: 50.ms
  affected_check.p95: 200.ms
  first_diagnostic.p95: 100.ms
  native_run_ready.p95: 2.seconds

  compiler.peak_memory: 1.gib
  compiler.background_cpu: 2.cores
  cache.local: 5.gib
  cache.remote_download_when: cost(download) < cost(recompute)

  agent.context: 40_000.tokens
  agent.repair_rounds.p95: 2
}

budget ci_pull_request {
  wall.p95: 8.minutes
  cpu: 30.core_minutes
  peak_memory_per_job: 8.gib
  network: 2.gib
  retry_flake_rate: 0.1.percent
}
```

The numbers above are hypotheses, not defaults promised by the project. Every
reported result includes target, host class, corpus, edit/workload shape, cache
state, concurrency, samples, variance, and toolchain version.

Budgets have three responses:

- **prevent:** reject constructs with known catastrophic/unbounded behavior in
  restricted contexts;
- **explain:** show the responsible semantic nodes and cost delta;
- **gate:** block a configured verify/release lane after statistically credible
  regression, with an explicit reviewed exception when the value justifies it.

Do not make noisy microbenchmark thresholds fail every edit.

## Compiler invariants

### Work proportionality

- Parse incrementally and retain useful trees for incomplete source.
- Model analysis as deterministic queries with observed dependencies.
- Keep source, semantic-interface, layout/ABI, and code-generation fingerprints
  separate so a comment does not rebuild a binary and a layout change does.
- Consume compact exported interfaces rather than transitive source where
  possible.
- Avoid whole-program inference, global coherence searches, and implicit
  resolution whose cost or invalidation radius is difficult to bound.
- Put projection queries around large monolithic results so unchanged semantic
  slices stop invalidation propagation.
- Recompute when hashing/serialization/cache lookup is more expensive than the
  result; incremental machinery has overhead of its own.

Rust's incremental compiler documentation explicitly notes that fingerprinting
can be costly and that backend objects require special coarse-grained handling;
incrementality is an engineering tradeoff, not free speed. ([rustc incremental
compilation](https://rustc-dev-guide.rust-lang.org/queries/incremental-compilation-in-detail.html))

### Bounded language features

- Type/effect/ownership inference is local and bidirectional across explicit
  public boundaries.
- Trait/port resolution has deterministic candidate sets and depth/work limits.
- Generic specialization, inlining, constant evaluation, proof search, macro or
  derivation expansion, and optimizer passes have budgets and explanations.
- Recursive types, pattern matrices, deeply nested expressions, and pathological
  source inputs receive complexity guards.
- Compile-time extensions are deterministic, sandboxed, capability-limited, and
  separately cached; they cannot monopolize or crash the compiler service.
- Expensive release assurance does not silently enter the keystroke lane.

### Data structures and memory

- Measure before adopting a clever structure; asymptotic wins can lose at
  realistic sizes.
- Prefer dense indexes/IDs, contiguous storage, arenas for phase-local objects,
  interning of repeated immutable values, small-vector/inline storage where
  measured, and compact bitsets for dense set operations.
- Avoid pointer-rich object graphs on hot traversals, repeated string keys,
  accidental cloning, boxing per node, and retaining entire syntax/IR histories.
- Make ownership of compiler caches explicit; bound them and expose the value,
  age, bytes, hit cost, recompute cost, and eviction reason.
- Stream artifacts and diagnostics rather than materializing giant buffers when
  consumers can act incrementally.

### Scheduling

- The foreground query preempts speculative/background work.
- Edits cancel obsolete analysis and codegen at safe points.
- Parallelism is resource-aware: jobs declare expected CPU, memory, I/O, and
  exclusive resources; the scheduler avoids oversubscription.
- Parallel work shares immutable results by content identity but not mutable
  compiler sessions across worktrees.
- Deduplicate identical in-flight queries.
- Deterministic results must not depend on scheduling order.

## AI development-loop efficiency

The AI loop consumes the same semantic graph rather than repeatedly reading the
whole repository or parsing prose diagnostics:

```text
intent
  -> minimal semantic context bundle
  -> transactional edit
  -> diagnostic/evidence delta
  -> affected verification
  -> next context bundle only if needed
```

Rules:

- Return stable IDs, compact typed facts, and deltas; do not resend unchanged
  source, logs, or full test output.
- Provide the smallest sufficient context by default and explicit expansion
  queries when the model needs full fidelity.
- Run cheap discriminating checks before expensive broad ones.
- Persist failure reproductions, minimized inputs, and evidence so another
  agent does not rediscover them.
- Parallelize only independent uncertainty; do not ask several agents or jobs to
  duplicate the same exploration without a deliberate diversity experiment.
- Charge model tokens, tool calls, wall time, and repair rounds together. Short
  source that causes retries is not token-efficient.
- Summarize output only after retaining a lossless artifact address; compression
  must not erase provenance or uncertainty.

## CI and verification lanes

```text
keystroke: parse + names + local type/effect/ownership facts
save:      sound affected check + formatter + direct obligations
commit:    affected specs/properties/contracts + dev native build
PR:        hermetic clean/incremental samples + multi-target + security/policy
nightly:   broad fuzz/model/load/fault/performance matrices
release:   exhaustive policy + reproducibility + provenance + canary
```

- Local and CI execute the same action graph; CI supplies trusted credentials,
  scale, clean environments, and release authority.
- Order gates by expected information gain divided by cost and run independent
  gates concurrently only within resource budgets.
- Fail fast on deterministic structural errors, but preserve already-produced
  independent evidence.
- Test sharding accounts for setup, skew, database/port contention, and output;
  more shards are not always faster or cheaper.
- Flaky retries are measured defects. A retry may gather evidence but cannot
  silently turn red into green.
- Remote execution/caching includes queue, upload, download, decompression, and
  egress cost. Build locally when that is cheaper.
- Track clean and incremental builds separately because they represent distinct
  experiences and cache states. Bazel's guidance explicitly recommends this and
  records work proxies such as packages loaded, targets configured, actions
  created/executed, and artifacts produced. ([Bazel build performance](https://bazel.build/advanced/performance/build-performance-breakdown))

## Runtime mechanical sympathy

The language supplies inspectable defaults rather than claiming universal
zero-cost abstraction:

- predictable representation with queryable size, alignment, padding, boxing,
  pointer indirection, and cache-line footprint;
- streaming and bounded buffers by default for potentially large input;
- structured concurrency and bounded queues instead of thread/task explosion;
- explicit allocation domains and caller-owned arenas in hot/native paths;
- data-oriented layouts as derived representations behind stable domain APIs;
- batching only with visible latency, memory, cancellation, and partial-failure
  policy;
- nonblocking I/O separated from blocking foreign work;
- copies, retains/releases, scheduler hops, locks, syscalls, and network calls
  visible through static explanation and runtime profiles;
- specialization and vectorization reported with both successes and misses.

Backend-friendly IR matters: LLVM warns that small IR differences can strongly
affect generated code and documents super-linear behavior for some shapes.
([LLVM frontend performance tips](https://llvm.org/docs/Frontend/PerformanceTips.html))
The frontend therefore owns a backend-neutral defined IR and measures each
lowering rather than relying on backend folklore.

## Measurement and regression protocol

1. State the decision the measurement will change.
2. Choose representative corpus/workload and adversarial sizes.
3. Measure cold, warm, clean, incremental, successful, and failing paths that
   users actually experience.
4. Record hardware, OS, load, compiler/backend, configuration, samples, and raw
   artifacts.
5. Use repeated runs and distributions; isolate noise and report uncertainty.
   LLVM's own benchmarking guidance requires multiple runs and warns that low
   noise alone does not remove bias. ([LLVM benchmarking](https://llvm.org/docs/Benchmarking.html))
6. Compare wall, CPU, memory, I/O, network, work counts, output size, and
   correctness—not a single favorable metric.
7. Use a representative suite plus targeted microbenchmarks; neither substitutes
   for the other.
8. Bisect and profile before optimizing. The compiler emits query/pass timing,
   allocation, invalidation, cache, and critical-path data for itself.
9. Ratchet only stable metrics with a noise margin. Quarantine unstable metrics
   as observations, not gates.
10. Accept a regression only with its measured benefit, owner, scope, expiry or
    revisit trigger, and cheaper alternatives considered.

## Efficiency review checklist

Every material feature proposal answers:

- What work does it add in keystroke, check, build, verify, release, startup,
  steady-state, and incident paths?
- What are its time and space complexities, including adversarial input?
- Which data is traversed, allocated, copied, hashed, serialized, retained,
  uploaded, downloaded, and invalidated?
- Which semantic boundary contains its cost?
- Can an agent/human query why the work happened?
- Does it widen inference, invalidation, authority, specialization, or runtime
  fan-out?
- Is the cache/reuse path actually cheaper at realistic sizes?
- What happens under cancellation, memory pressure, partial failure, and stale
  evidence?
- Which benchmark and workload will falsify the efficiency claim?
- What value justifies spending the compute?

## Anti-patterns

- “Hardware is cheap” used to excuse repeated or unbounded work.
- “Performance first” used to optimize unmeasured cold code or obscure source.
- Hiding an expensive analysis behind incremental caching.
- Caching nondeterministic or incompletely keyed results.
- Global inference/coherence whose cost depends on the whole ecosystem.
- Monomorphizing every generic/effect/provider combination by default.
- Running release optimization, broad fuzzing, or all tests after every edit.
- Adding CI shards without measuring setup, contention, memory, and tail time.
- Remote cache downloads that exceed local recomputation.
- Unbounded telemetry/cardinality in order to understand performance.
- Parallelism that improves one job's wall time while degrading the machine and
  all other work.
- Summaries that hide which checks were skipped, stale, or inconclusive.
- A performance gate so noisy that contributors learn to rerun it.

## Primary evidence

- [rustc incremental compilation](https://rustc-dev-guide.rust-lang.org/queries/incremental-compilation-in-detail.html)
  documents dependency queries, red-green validation, stable fingerprints, and
  the real hashing/persistence/backend overheads.
- [rustc profiling](https://rustc-dev-guide.rust-lang.org/profiling.html) and
  [rustc-perf](https://github.com/rust-lang/rustc-perf) show compiler self-profile
  and per-change performance monitoring as ongoing infrastructure.
- [Go compiler export data](https://go.dev/src/cmd/compile/README) demonstrates
  compact, lazily decoded summaries for separate compilation and explains the
  deep-versus-shallow metadata tradeoff.
- [Tree-sitter](https://tree-sitter.github.io/tree-sitter/) demonstrates
  incremental, error-tolerant syntax trees designed for per-keystroke updates.
- [Bazel build performance](https://bazel.build/advanced/performance/build-performance-breakdown)
  distinguishes clean/incremental experiences and measures actual graph work,
  memory, CPU, I/O, and network behavior.
- [Build Systems à la Carte](https://www.microsoft.com/en-us/research/publication/build-systems-a-la-carte/)
  separates build-system mechanisms so their correctness and performance
  properties can be reasoned about deliberately.

All sources were last verified 2026-09-02. Numeric budgets and project rules
are synthesis to be tested, not claims established by those sources.
