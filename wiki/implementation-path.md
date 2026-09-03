---
id: implementation-path
title: Implementation and adoption path
summary: A strangler-fig roadmap that validates the differentiating ideas before committing to a new runtime and ecosystem.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [roadmap, compiler, runtime, ecosystem, strangler-fig]
related: [vision, author-intent-model, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, design-atlas, feature-coherence, effects-and-capabilities, compiler-feedback-latency, compute-efficiency-constitution, runtime-profiles-dogfooding, memory-reclamation-policy, networking-http-tls-security, domain-data-distribution, blind-spots-boundaries, context-telemetry-security, performance-observability-delivery, resources-locks-caching, data-asset-pipelines, modules-architecture-live-development, ai-native-runtime-evals, native-low-level-profile, capability-gauntlet, convergence-audit, convergence-work-program, example-tour, open-questions]
---

# Implementation and adoption path

## Strategy

Do not begin by building a parser, optimizing compiler, distributed runtime,
package registry, ORM, AI inference engine, and standard library simultaneously.
Begin with the claims that differentiate this language and could be false:

1. Typed effects can improve AI correctness and architecture without creating
   intolerable signature noise.
2. Structured compiler/runtime feedback can reduce agent repair rounds and
   tokens materially.
3. Ownership, lexical resources, tasks, actors, and effects can share one useful
   native-capable semantic model without forcing borrow ceremony into ordinary
   application code.
4. Local executable specifications can improve evidence without overwhelming
   source readability.
5. A compile-time provider graph can remove DI plumbing without recreating a
   reflective container or an incomprehensible implicit-resolution system.
6. Typed causal events and classified context can improve diagnosis within
   explicit privacy, trust, and runtime budgets.
7. Sound affected checks and high-leverage diagnostics can remain fast enough
   to sustain an AI repair loop as projects grow.
8. Deterministic effect worlds can reproduce meaningful distributed failures
   without implying that all production nondeterminism is simulated.
9. A compiler-verified modular monolith can prevent architectural drift while
   remaining easier to build and operate than premature services.
10. Flat, local specs and strict typed boundary validation can improve evidence
    without creating a second metaprogrammed test or web language.
11. Static cost/resource explanations and one correlated evidence graph can
    shift performance, operations, profiling, load behavior, and delivery left
    without pretending that any one instrument replaces the others.
12. Explicit modules and component policy can enforce useful architecture while
    filesystem conventions remain a replaceable projection rather than dogma.

Everything else should initially live off the land.

All stages follow the [compute-efficiency constitution](compute-efficiency-constitution.md):
compute may be spent generously for useful evidence, but redundant analysis,
mis-scoped CI, uneconomic caching, pathological algorithms, and opaque resource
cost are treated as defects.

## First artifact: an evaluation corpus

The corpus begins with the [semantic kernel probe
suite](semantic-kernel-probes.md): 24 blocking laws and an extended adversarial
matrix derived from the [kernel contract](semantic-kernel-contract.md). Its
stable evidence schema becomes the first host-side tool that Lang can later
dogfood.

Before choosing syntax, create 20–30 representative programs selected from the
[capability gauntlet](capability-gauntlet.md):

- pure domain transformations;
- opaque identity generation and UUID/database representation boundaries;
- validation and typed errors;
- HTTP plus database service;
- multi-tenant request with authentication and context propagation;
- payment workflow with compensation;
- bounded parallel fan-out;
- stateful supervised actor;
- distributed retry and idempotency;
- bounded streams, backpressure, connection pools, and overload;
- event-sourced state plus projection migration and replay;
- deterministic network/storage/process fault simulation;
- parser/binary protocol;
- state-machine transition coverage and grammar ambiguity/resource limits;
- desktop GUI model/update/view with native and web host adapters;
- streaming model inference with cost limits;
- bounded agent tools, handoffs, adversarial evals, and recorded replay;
- C library integration;
- allocation-sensitive native loop;
- pointer/layout/atomic/FFI conformance probes;
- lexical files, sockets, arenas, foreign handles, cleanup failure,
  cancellation, ranked locks, and await-with-lock rejection;
- cache freshness, invalidation, negative entries, memory pressure, and
  stampede suppression;
- batch cleaning, lineage, event-time streams, backfills, warehouse boundaries,
  and a reproducible asset graph;
- benchmark, load/fault, profile, SLO, canary, and rollback evidence;
- private exports, dependency-cycle diagnostics, architecture projections, and
  transactional live-workspace sessions;
- restricted embedded task;
- schema and package migration;
- rolling stateful upgrade, rollback, backup, and restore;
- production incident diagnosis;
- telemetry overload, PII leak, and untrusted-context incidents;
- large AI-authored refactor;
- Git-worktree-parallel AI changes sharing immutable cache state safely.

Implement comparable slices in Gleam, Elixir, Rust, Go, Koka or Unison where
practical, and the candidate notation. Measure:

| Dimension | Example measure |
|---|---|
| Agent correctness | compile/test success, hidden-test success, semantic defects |
| Repair efficiency | iterations, tool calls, generated tokens, wall time |
| Context efficiency | tokens needed to represent the relevant semantic slice |
| Human audit | time to answer dependency/effect/failure questions accurately |
| Diagnostic quality | localization, repair applicability, false positives |
| Build loop | cold and incremental parse/check/test latency |
| Invalidations | queries/modules/tests rebuilt by representative edit shape |
| Runtime | throughput, p50/p99 latency, allocations, RSS, startup, binary size |
| Operability | time/evidence needed to explain injected failures |
| Ecosystem cost | dependencies, build scripts, licenses, vulnerabilities, cache size |

This prevents optimizing for aesthetically pleasing toy examples or one model's
training distribution.

## Bootstrap architecture

```text
canonical source
    |
lossless syntax tree ---- formatter / semantic diff / editor views
    |
resolved high-level IR --- symbol + component + dependency graph
    |
typed/effect IR ---------- contracts + change impact + agent queries
    |                       flow summaries + source/sink diagnostics
    |
portable core IR --------- reference interpreter / native / BEAM / Wasm backends
    |
typed runtime events ----- tests / traces / profiles / replay / OTel export
```

The compiler should be a persistent query service from the beginning. CLI,
editor, CI, and agent protocol are clients of the same incremental state rather
than four tools that repeat analysis.

See [Compiler and feedback latency](compiler-and-feedback-latency.md) for the
candidate service-level budgets and architecture.

## Dogfood before self-hosting

Move compounding leaf tools into Lang as soon as their required semantic slice
is trustworthy, while retaining a host-language replay oracle. The order is:
executable corpus programs, pure standard modules, corpus/evidence runner,
diagnostic and cache/invalidation clients, formatter/build driver, selected
compiler passes, then the compiler frontend. This earns daily feedback without
making an immature compiler the only tool capable of repairing itself.

The complete profile and bootstrap rationale is in [Runtime profiles and
dogfooding](runtime-profiles-and-dogfooding.md).

## Candidate staged roadmap

### Stage 0 — executable design

- Freeze no syntax.
- Build the evaluation corpus and rubric.
- Define typed JSON schemas for diagnostics, symbol queries, semantic diffs,
  evidence bundles, and runtime events.
- Version those fixtures from their first revision and include source edition,
  kernel/protocol version, target/profile, provider identity, and policy identity
  where relevant; this preserves evolution without promising stability yet.
- Define typed-IR summaries for input sources, transformations, validators,
  encoders, sinks, endorsements, declassifications, and foreign uncertainty;
  run the 18 trust-flow probes without requiring a production taint runtime.
- Mock the agent protocol against hand-authored program graphs.
- Compare conventional-braced, indentation-oriented, and uniform-tree source
  interfaces over one semantic model.
- Run the ownership laboratory and the small interpreter/Cranelift/QBE lowering
  contest defined by the convergence work program.
- Preserve point and edge identities for inferred local loan endings; the
  bounded CFG spike has validated this IR requirement before parser work.
- Preserve the validated interface-only contract: compiler-verified
  parameter/field origin paths, tagged alternatives, shared/exclusive access,
  independent generic abilities, and targeted fresh callback origins. Keep its
  sample spelling provisional until the surface experiment.
- Carry a compact content-bound validation manifest for those typed-core facts
  across package, cache, CI, release, and native-lowering boundaries. Recompute
  cheap ownership rules independently; emit full state traces only on mismatch,
  and never claim the manifest proves frontend derivation.
- Establish benchmark machines, corpus sizes, edit shapes, and cold/warm
  latency dashboards before adding expensive semantics.
- Exercise clean, empty-cache, offline/vendored, and previous-stage recovery
  paths before self-hosting or registry work can hide bootstrap assumptions.
- Add the `LIFE-*` scenarios from the
  [lifecycle backcast](language-ecosystem-lifecycle-backcast.md) as fixture and
  metadata obligations without building a registry or mature update system.
- Assign every gauntlet row a mechanism/non-goal, select the first workload from
  measured coverage/cost, and define its G3 acceptance.

Exit: demonstrate that the proposed feedback lets an agent diagnose and repair
representative defects more reliably than terminal-text baselines; select an
ownership model, v0 source surface, initial native lowering, and first G3
workload with recorded evidence.

### Stage 1 — native-capable semantic kernel

- Lossless parser and deterministic formatter.
- Stable symbol/node identities that survive formatting and nearby edits.
- ADTs, records, functions, explicit modules, named calls, pattern matching,
  results, owned values, lexical borrows, moves, and deterministic resources.
- Private-by-default exports, explicit imports, cycle detection, and no
  effectful top-level initialization.
- Reference interpreter for defined semantics plus a deliberately small native
  lowering; spike Cranelift first because its stated priorities include fast
  compilation, security, and relative simplicity.
- A stable C-compatible layout/ABI subset, unsafe boundary contracts, explicit
  allocators/arenas, fallible allocation, and checked release behavior.
- Minimal non-resumable capability calls/providers for files, clocks, allocation,
  and process I/O so resourceful programs do not rely on ambient globals; Stage
  2 generalizes these into effect rows, handlers, scopes, and architecture.
- Structured diagnostics with machine-applicable fixes.
- LSP plus agent query protocol.
- Persistent query graph, exported semantic fingerprints, invalidation
  explanations, and a low-optimization native development evaluator.

Exit: pure domain programs plus the selected native G3 proof parse,
type/effect/ownership-check, format idempotently, run in the interpreter and
native backend, cross a small C boundary, and support transactional AST edits.
The native proof must demonstrate a file/socket-like resource, an arena,
cleanup on every exit path, bounds/OOM behavior, and no hidden release-mode
safety weakening. The Lang corpus/evidence runner is the leading candidate,
not a commitment; its fixture set includes the streaming JSON/data stress case.
Representative warm checks and dev builds must remain within
measured budgets or the stage does not graduate.

### Stage 2 — effects, contracts, and architecture

- Effect rows, capabilities, lexical one-shot handlers.
- Statically resolved providers, lifetimes, cleanup, and lexical overrides.
- Narrow typed execution context with propagation/trust rules.
- Explicit public effects with private inference.
- Component DAG and configurable architecture policies.
- Opinionated verified-modular-monolith policy, private component storage/API
  checks, and extraction-readiness diagnostics.
- Examples, properties, generated boundary inputs, shrinking, and coverage
  obligations in one test runner.
- Change-impact and affected-test selection.
- Deterministic test worlds for clock, randomness, scheduler, network, storage,
  and process failure, with seed/replay evidence.
- Typed transactional live workspace with recorded providers, effects, results,
  and promotion into ordinary source/spec/fixture artifacts.

Exit: service-domain examples prove the central effect/assurance/architecture
thesis before the native runtime grows a full async service scheduler.

### Stage 3 — native application and service runtime

- Add structured async I/O, cancellation, task scopes, supervised actors,
  bounded mailboxes/streams, scheduler and blocking isolation to the native
  runtime.
- Integrate lexical connections, pools, deadlines, backpressure, overload,
  ranked-lock diagnostics, and runtime dependency-cycle detection.
- Integrate runtime tree, task, scheduler, allocation, lock, queue, and resource
  inspection.
- Project typed causal events into request summaries, traces, metrics, and
  redacted live-agent views under declared budgets.
- Package as a native executable and minimal container image.

Exit: deploy a fault-injected production-style service and diagnose it entirely
through structured evidence.

The native runtime proves the chosen v1 scope honestly. It should still borrow
OTP's supervision lessons and leave actor semantics precise enough to test on a
later BEAM backend; it must not casually recreate transparent distributed
actors or unbounded mailboxes.

### Stage 4 — package and official toolkit

- Hermetic resolver, full lock graph, checksums, vendoring, license/SBOM data.
- Sandboxed build plugins with no network by default.
- Content-addressed local build cache; remote cache protocol later.
- Official kits for data/encoding, testing, telemetry, HTTP, SQL interfaces,
  actors/workflows, resilience/backpressure, versioned schemas, local caches,
  event sourcing/CQRS templates, and model capabilities.
- Official performance kit for benchmarks, load/fault scenarios, profiles,
  SLOs, canaries, and artifact-linked evidence.
- Official data/asset kit for typed cleaning, dataset contracts, lineage,
  batch/stream event time, backfills, columnar/Arrow interop, warehouses, and
  reproducible assets.
- Typed cache policies covering key completeness, freshness, validation,
  invalidation, stampede suppression, capacity, and deterministic testing.
- Strict JSON/schema validation, Ecto-like changes/repository interfaces,
  consumer contracts, state-machine/grammar projections, and UUIDv7 identity
  providers.
- First-party templates are ordinary typed packages, not a second templating
  language.
- AI toolkit artifacts include provider-neutral model capabilities, typed tool
  authority, run budgets, classified traces, durable handoffs, and calibrated
  stochastic evals.

Exit: a new service needs one tool and no third-party framework to become
tested, observable, packaged, and deployable.

### Stage 4b — application hosts

- Incubate a pure model/event/update/view UI kit.
- Implement one native desktop host and one web host without claiming pixel or
  behavior identity.
- Integrate accessibility, text/IME, focus, async command cancellation, and UI
  event inspection into the evidence model.

Exit: one non-toy tool ships through both hosts while retaining explicit
platform differences and a usable native escape hatch.

### Stage 5 — BEAM and Wasm component profiles

- Compile the portable subset to a BEAM-compatible representation.
- Map actors, links, supervision, mailboxes, and service lifecycle to OTP while
  explicitly reporting semantic/profile differences.
- Generate safe Erlang/Elixir/Gleam boundary adapters and package OTP releases.
- WIT-compatible imports/exports and capability-based hosting.
- Wasm plugins for compiler derivations and untrusted extensions.
- Browser/edge/server component deployment.

Exit: the same conformance and service corpus runs across native, BEAM, and
Wasm where each profile claims support; compose components across languages
without raw shared memory.

### Stage 6 — native depth and optimization

- Add optimizing tiers and only then evaluate LLVM or another backend where
  measured throughput/platform gaps justify the extra compile-time complexity.
- Deepen C import/export, pointer provenance, aliasing, initialization, integer,
  panic/unwind, atomic, vectorization, ownership, and blocking contracts.
- Region/arena/allocator APIs, fallible OOM, ownership inference, explicit
  boundary borrows, and transitive unsafe accounting.
- Checked semantics remain checked across release optimization; explicitly
  unchecked operations require local safety contracts.
- Static binaries and cross-compilation.

Exit: run allocation-sensitive, C-heavy, server, emulator, storage, and
renderer slices competitively without making borrow syntax mandatory in
ordinary application code.

### Stage 7 — restricted embedded profile

- No-runtime subset, fixed/region allocation, restricted effects, panic policy.
- Cross toolchains and hardware-in-loop test protocol.
- Explicit real-time guarantees only where measured and proven.

Exit: the embedded slice retains language semantics and produces predictable
artifacts; unsupported constructs fail statically.

### Stage 8 — self-host selectively

Self-hosting is not proof of usefulness. Move formatter, package tool, compiler
passes, and runtime components into the language only when it improves dogfood,
reliability, or contributor velocity without bootstrapping fragility.

## Live-off-the-land choices

| Need | Initial dependency | Possible later ownership trigger |
|---|---|---|
| Native code generation | Cranelift candidate, measured against alternatives | Backend size, compile latency, platform support, or control becomes limiting |
| Service/runtime lessons | Native runtime plus OTP semantics/conformance corpus | Add BEAM backend where mature scheduling, isolation, or ecosystem interop produces measured value |
| Portable isolation | Wasm component ecosystem | Only replace adapters; retain standards compatibility |
| Telemetry export | OpenTelemetry | Keep internal event model; own exporters only for missing semantics |
| C parsing/bindings | Clang-based tooling or Zig-style translation | Own only the stable contract layer, not a C frontend without cause |
| Package storage | Existing object storage and transparency mechanisms | Dedicated registry after ecosystem demand |
| Model inference | Provider/local-runtime adapters | Own scheduling/cache layers when workload evidence supports it |

Replacing a dependency because it exists is not simplification. Replacement is
justified by measured control, security, performance, stability, or semantic
needs.

## Dependency policy

Candidate defaults:

- Imports may use only declared direct dependencies.
- The lockfile records the complete graph, source digest, license, build
  capabilities, minimum toolchain, and provenance.
- Resolution and fetching are separate from building; builds have no network.
- Build plugins run sandboxed with declared inputs and outputs.
- One command explains why any transitive dependency or feature is present.
- Dependency additions show estimated compile, binary, vulnerability, and
  capability impact before acceptance.
- Projects can set depth, count, license, build-script, and size budgets.
- Vendoring is supported and verified against the lock graph.
- Tiny copied implementations retain source/license provenance and become
  owned code with tests; copying is not a license or maintenance escape.

Go's module checksum database demonstrates immutable content verification and
vendoring; Cargo demonstrates a rich resolver but also documents feature
unification and duplicate-build complexity. ([Go modules](https://go.dev/ref/mod),
[Cargo features](https://doc.rust-lang.org/stable/cargo/reference/features.html))

## Toolkit lifecycle

```text
community -> incubator -> official kit -> stable core interface
                         \-> community again when superseded
```

- **Core:** tiny, universally required semantic/runtime interfaces with a long
  compatibility promise.
- **Official kits:** batteries maintained with the toolchain but independently
  versioned.
- **Incubator:** promising APIs gathering production evidence; no stability
  promise hidden behind a “standard” label.
- **Community:** open experimentation with the same provenance and capability
  metadata.

An ORM, web toolkit, and model client belong in official kits, not syntax.
Stable schemas and capability interfaces may eventually move into core while
implementations continue to evolve independently.

## Versioning model

Version separately:

- source edition;
- compiler implementation;
- core language specification;
- portable IR;
- runtime ABI/protocol;
- agent protocol;
- official toolkit packages.

Release them on a tested compatibility train. New editions are opt-in,
cross-edition components link together, and migrations are previewable semantic
edits followed by verification. Rust editions provide a strong precedent for
surface evolution without ecosystem splits. ([Rust editions](https://doc.rust-lang.org/edition-guide/editions/))

Semantic version declarations communicate intent; they do not prove
compatibility. The toolchain compares public types, effects, capabilities,
contracts, wire/storage schemas, and runtime protocols and reports source,
binary, behavior, and rollout compatibility separately. Deleted wire identities
remain reserved. Stateful upgrades include checked forward, coexistence,
rollback, and downgrade paths.

## Build and CI model

The local command and CI command execute the same hermetic action graph. CI adds
scale and policy, not a new truth source.

```text
ai verify --changed       # fastest sound affected slice
ai verify --workspace     # full local evidence
ai release --target linux # hermetic artifact, SBOM, provenance, signatures
```

Inputs, tools, environment, and outputs are declared and content-addressed.
Cache lookup should account for network transfer cost; downloading a large hit
can be slower than rebuilding it. Reproducibility is a prerequisite for shared
caches, not a side effect of using one. ([Bazel caching](https://bazel.build/remote/caching),
[hermeticity](https://bazel.build/concepts/hermeticity))

Containers are an output target, not the development environment's semantic
foundation. Native artifacts, OTP releases, Wasm components, and OCI images all
derive from the same release graph.

Git worktrees get isolated mutable compiler sessions while sharing only
immutable content-addressed artifacts. Paths in artifacts and diagnostics are
repository-relative so concurrent agents do not poison one another's caches or
produce checkout-specific diffs.

## Risk register

| Risk | Early warning | Mitigation |
|---|---|---|
| Effect system is elegant but noisy | Public signatures expand; users mask effects | Compare effect renderings and abstraction boundaries in corpus |
| “Perfect architecture” becomes dogma | Tiny programs require boilerplate | General graph policy with scale-sensitive presets |
| Compiler becomes the bottleneck | Edit feedback exceeds 100–200 ms | Persistent queries, dependency fingerprints, performance budgets from day one |
| Assurance features block iteration | Proof search/fuzzing runs on every save | Sound fast lane, explicit deferred evidence, small checker, release lane |
| Runtime scope explodes | Native, BEAM, and Wasm semantics diverge | Small specified portable subset; explicit profile restrictions; shared conformance suite |
| Built-in telemetry is too expensive | Allocation/latency/cardinality regressions | Typed event budgets, sampling, focused capture, compile-out profile |
| Provider/context inference becomes magic | Ambiguous resolution, hidden authority, leaked request state | Composition-only synthesis, explicit rendering, lifetime checks, no global overrides |
| Automatic propagation leaks identity or secrets | Sensitive context reaches sinks or third parties | Deny-by-default propagation, classification, trust-boundary conversion, seeded security tests |
| Generated tests create false confidence | High coverage, low mutation score, production defects | Oracle provenance, mutation checks, unresolved-obligation reporting |
| Agent protocol becomes vendor-specific | Only one model/tool can use it | Versioned documented schemas; CLI and SDK clients; conformance fixtures |
| Standard toolkit fossilizes | Core release required for ordinary API evolution | Independently versioned official kits and incubation |
| Escape hatches dominate | Growing suppressions/unsafe/foreign surface | Queryable budgets, expiry/rationale, component isolation |
| Training-data scarcity hurts models | Candidate language underperforms familiar syntax | Retrieval of canonical examples, constrained edits, compiler feedback, benchmark multiple models |
| Domain toolkit becomes architecture dogma | Small services inherit CQRS/event-store machinery | General semantic core; opt-in declarative kits with operational-cost reports |
| Built-in storage is mistaken for durability | Data cannot be restored or survive partitions | Explicit storage contracts, recovery manifests, and tested restore evidence |

## Recommended first four spikes

1. **Native semantic kernel:** implement Stages 1–3 and the first three suites
   from the [ownership and lifetime decisive
   study](ownership-and-lifetime-decisive-study.md), plus the blocking native/
   FFI cases and a small portable IR in the reference interpreter and one
   minimal native lowering. Use the
   evidence to select the first G3 workload; the Lang corpus/evidence runner is
   the leading candidate and includes the streaming JSON/data case. Compare warm
   build time, query work, runtime, diagnostics, binary size, and implementation
   burden. Run the broader backend contest after the semantic harness is sound.
2. **Effect/provider ergonomics:** model five corpus programs in Koka, Roc,
   Unison, Gleam, and two candidate syntaxes. Compare explicit capability parameters,
   compile-time providers, request scopes, and overrides; measure signature noise
   and diagnostic explainability.
3. **Agent evidence loop:** implement a fake compiler server that returns symbol,
   graph, diagnostic-delta, and transactional-edit schemas. Compare agent repair
   performance with ordinary text errors.
4. **Runtime mapping:** implement one supervised actor service in the candidate
   native runtime and Elixir/Gleam; define the portable subset without claiming
   location transparency.

Compiler latency is measured in all three spikes. It is not deferred to a
fourth performance project.
