---
id: open-questions
title: Open questions and decision queue
summary: Prioritized questions that should be answered through conversation, research, or executable spikes.
type: ledger
status: active
confidence: high
created: 2026-09-02
updated: 2026-09-03
tags: [questions, decisions, experiments]
related: [vision, author-intent-model, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, design-atlas, compiler-feedback-latency, compute-efficiency-constitution, runtime-profiles-dogfooding, memory-reclamation-policy, networking-http-tls-security, convergence-work-program, domain-data-distribution, blind-spots-boundaries, performance-observability-delivery, resources-locks-caching, data-asset-pipelines, modules-architecture-live-development, ai-native-runtime-evals, native-low-level-profile, capability-gauntlet, convergence-audit, implementation-path, research-ledger]
---

# Open questions and decision queue

## How questions graduate

- **Conversation:** values, priorities, and acceptable tradeoffs only the project
  author can choose.
- **Research:** existing semantics, implementation experience, and known failure
  modes.
- **Spike:** feasibility, ergonomics, performance, and AI/human outcomes.
- **Decision:** chosen only after the relevant evidence is visible.

## Inferred defaults that do not need another author answer yet

The author resolved the former highest-impact fork on 2026-09-02: **usable
low-level ownership and resource programming belongs in the first
implementation**, accepting its compiler and design burden. The precise model
and native backend remain experiments, not value questions.

- Production is compiled-first with native AOT in v1; the interpreter is a
  semantic oracle and live-development engine.
- Runtime services are monotonic profiles. Freestanding/sequential programs do
  not pay for actors, scheduling, or tracing GC; applications/services may opt
  into the common runtime.
- The memory boundary is accepted: no universal tracing heap is required by
  the kernel; unique ownership, borrowing, resources, and arenas come first;
  managed cyclic graphs are explicit owner-confined regions. Precise RC/reuse
  and the managed-region algorithm remain measured implementation choices.
- Service scheduling starts with compiler safe points and bounded task budgets;
  blocking CPU/I/O work is classified and isolated. Arbitrary asynchronous
  preemption must earn its complexity.
- Cache correctness uses core purity/effect/version/authority semantics plus an
  official typed policy kit. It does not receive dedicated v1 syntax and is
  never silently automatic.
- The first Lang-written application candidate is the corpus/evidence runner,
  with streaming JSON/data as an internal stress workload. Full self-hosting is
  deliberately later.
- The component/capability DAG keeps a verified modular-monolith default.
- The first semantic prototype uses strict left-to-right evaluation and
  non-resumable statically resolved capability calls. General multi-shot effect
  handlers are not part of v1.
- The ownership prototype uses a value/access surface over affine place/loan/
  origin semantics. Public modes and returned origins are explicit; partial
  residual moves, escaping local borrows, and ordinary borrows across `await`
  are initially rejected. The deeper study now leads toward visible named-value
  transfer and exclusive access at call sites, but the laboratory still
  compares declaration-only and ambiguity-only alternatives.
- Automatic resource release is restricted and cannot replace a primary typed
  outcome; correctness-significant finish/commit remains an explicit consuming
  `Result` operation.
- Perl's source-to-sink insight is retained through typed boundaries and
  flow summaries, but a universal Boolean runtime taint bit is not the leading
  candidate. Validation, encoding, confidentiality, and authority stay
  separate.
- Long-term optionality comes from a versioned kernel/evidence contract,
  editions and migrations, narrow stable foreign boundaries, replaceable
  providers, hermetic content-addressed builds, and impact probes—not from
  implementing multiple runtimes, registries, collectors, or frameworks now.
- Broad semantic discovery has reached diminishing returns. New concerns reopen
  the kernel only under the evidence triggers in the
  [lifecycle backcast](language-ecosystem-lifecycle-backcast.md).

Except for the explicitly accepted memory boundary, these are trajectory-based
recommendations rather than accepted language decisions. Each still has an
explicit falsification experiment.

## Author choices that remain but do not block the next experiments

1. Is source text authoritative, or may a semantic code database be
   authoritative with text/Git as its canonical projection?
2. How strict should the default be when static assurance conflicts with the
   edit loop: reject, warn, generate an obligation, or permit a scoped waiver?
3. Which context is implicitly propagated: only trace/deadline/cancellation, or
   also verified principal and tenant—and where must those become explicit
   application/domain values?
4. Does “named arguments only” include unary calls, constructors, operators,
   callbacks, and pipelines, or is the intended rule “no ambiguous positional
   arguments”?
5. How visible should inferred types/effects be in checked-in source versus
   editor/agent views?
6. Which workload follows the toolchain dogfood proof: web service, durable AI
   workflow, multiplayer/game service, desktop tool, or two compared slices?
7. Should missing public evidence fail ordinary `check`, or only
   `verify`/`release` while `check` emits a precise obligation?
8. Should the service-profile `Id<Entity>` provider default to UUIDv7, or remain
   unbound until the application selects a storage/distribution policy?
9. Is a native desktop GUI part of the first production proof, or an official
   kit after the service/runtime thesis is established?

## Syntax experiments

- Conventional braces versus indentation-oriented keyword flow versus a
  uniform-tree machine projection over the same semantic model.
- Named-only calls versus named multi-argument calls with concise unary calls.
- Exact-name argument punning (`reservation:`) versus fully repeated labels and
  values.
- Plain effect calls (`Clock.now()`) versus marked calls (`Clock.now!()`).
- Explicit public types/effects versus a canonical rendered signature view.
- Pipeline placeholder syntax versus method-style chaining.
- Local specs inline, adjacent, folded, or rendered as a separate projection of
  the same semantic component.
- ASCII-only source versus optional Unicode aliases rendered canonically.
- First-class grammar notation versus official typed parsing library.
- Fixed typed literal forms for time/regex/SQL/bytes versus ordinary builders.
- Unicode identifier restriction and confusable-diagnostic policy.
- `snake_case` values/functions plus `PascalCase` types versus alternatives,
  measured on full repair/audit tasks rather than snippets.
- Pattern guards limited to pure expressions versus richer matching hooks.
- Syntax admission using the scorecard in [Feature coherence](feature-coherence.md).

## Semantic experiments

The broad semantic forks have been narrowed into falsifiable seams in the
[kernel contract](semantic-kernel-contract.md) and [probe
suite](semantic-kernel-probes.md):

- Non-resumable capability calls first; test whether one-shot resumable handlers
  later earn their ownership/runtime complexity.
- Declaration-site borrow/take modes with ambiguity-only explicit moves versus
  always-marked call sites.
- Local control-flow borrow inference and first-order public origin/access/
  ability summaries now have bounded executable support. Remaining composition
  risks are variance, field disjointness, nested lending, async suspension, and
  incremental invalidation rather than whether package summaries exist at all.
- The initial no-borrow-across-`await` restriction versus task-owned stable
  borrows.
- Restricted automatic release plus explicit fallible finish across real OS and
  foreign resource APIs.
- Ordinary isolated OOM versus explicit fallible allocator worlds and their
  effect on generic-library reuse.
- Stable aggregation of concurrent child failures after fail-fast cancellation.
- Exact semantic-event equivalence needed to validate optimization without
  freezing operational details.
- Strict versus relaxed floating behavior across targets and accelerators.
- Nominal public records with structural private records and consumer-owned
  behavioral ports, especially at public/security/foreign boundaries.
- Refinement types checked by SMT, executable contracts, or both in assurance
  tiers.
- Actor messages choosing copy, immutable share, or unique transfer through
  explicit cost evidence.
- Logical time passed as data versus `Clock` effect in each architecture layer.
- Typed execution context versus explicit parameters across each layer.
- Domain events and state machines as syntax, derivation, or ordinary ADTs only
  after the base kernel is executable.
- Small independently checked effect/type certificates versus one compiler
  trust boundary.
- `Input<Origin, T>` as source syntax versus an inferred projection of typed-IR
  provenance, and local versus targeted global/control-influence checks.

## Runtime experiments

- Reference-interpreter/native differential semantics from the first executable
  subset, including ownership, cleanup, OOM, panic, and C ABI behavior.
- Cranelift versus the simplest viable alternative for warm compile latency,
  runtime quality, platform reach, debug information, and implementation burden.
- BEAM code generation and typed OTP interop as a later conformance/backend path.
- Mapping structured task cancellation onto BEAM processes.
- Bounded mailbox and backpressure semantics.
- Precise RC/reuse versus one isolated managed-region implementation for
  explicit shared/cyclic graphs; global tracing is not a candidate fallback.
- Deterministic scheduling/replay boundaries.
- Whole-service deterministic worlds with network, disk, clock, random,
  scheduler, and process faults.
- Blocking foreign calls and scheduler isolation.
- Typed event ring-buffer overhead under realistic load.
- Typed causal event graph with request-wide-event projections.
- Classification-aware redaction, trust-boundary filtering, and focused debug leases.
- AoS/SoA views, arenas, and source-level layout explanations for native code.
- Container-aware memory and CPU budgets without pretending soft limits are hard.
- Durable workflow semantics: checkpointing, idempotency, compensation, and
  version migration.
- Runtime-local tables/caches and embedded durable logs without claiming a
  universal built-in distributed database.
- Stateful hot upgrade with mixed-version protocols, migration, downgrade, and
  rollback.
- Per-profile representation and release scheduling beyond the accepted
  ownership/resource kernel, including RC cascades, arena teardown, and
  per-owner managed-region tails.
- Actor reentrancy policy and mapping finite structured tasks onto BEAM.
- Native/web GUI host semantic overlap and explicit platform escape boundary.
- Native pointer provenance, aliasing, layout, initialization, unwind, and
  release-safety semantics lowered through interpreter and optimized backend.
- Inferred local borrows versus explicit ownership at stored, concurrent,
  foreign, and public boundaries.
- Realtime/interrupt restricted effect sets and bounded allocation.
- Lexical resource cleanup when body and cleanup both fail, including async
  cleanup and cancellation.
- Static lock ordering plus runtime dependency-graph validation; forbid ordinary
  guards across suspension and measure escape frequency.
- Cache key completeness, freshness/validation, invalidation, negative caching,
  capacity pressure, and stampede suppression.
- Batch/stream unification, event-time/watermark behavior, data quality,
  lineage, backfills, columnar interchange, and warehouse pushdown boundaries.
- Strict HTTP framing through real proxy chains, TLS provider/identity failure,
  0-RTT replay, SSRF/DNS rebinding, pool pressure, and provider upgrade/rollback.

## Toolchain experiments

- Stable AST/symbol identity across common edits.
- Transactional semantic edits and previewable repair plans.
- Minimal sufficient context bundles for coding agents.
- Diagnostic-delta scoring as an online agent reward.
- Query granularity and invalidation performance in a persistent compiler.
- Separate sound edit and exhaustive release lanes with explicit deferred/stale
  evidence.
- Worktree-isolated mutable compiler sessions sharing immutable content cache.
- Semantic diff rendering for human review.
- Hermetic content-addressed local builds before remote caching.
- Direct-dependency enforcement and dependency-cost explanations.
- Evidence manifests tying an artifact to tests, contracts, compiler/runtime,
  sources, and toolchain inputs.
- Directional semantic compatibility for APIs, effects, wire/storage schemas,
  actor state, and rolling deployments.
- Typed transactional live-workspace sessions, effect/provider recording, and
  promotion of discoveries into durable source/spec/fixture artifacts.
- Explicit module/export/import graphs, cycle diagnostics, and verified
  architecture conventions independent of folder layout.
- Correlated benchmark/load/fault/profile/SLO/canary evidence with separate
  purposes and bounded collection overhead.

## Evaluation experiments

- Multiple models to separate syntax effects from training-data familiarity.
- Hidden tests and injected architecture defects to prevent benchmark gaming.
- Human audit questions: “what can fail?”, “what has authority?”, “what changed?”,
  and “what happens under cancellation?”
- Mutation testing on changed code to evaluate oracle quality.
- Incident games measuring time to causal explanation.
- Token accounting across generation plus all repair rounds, not source length
  alone.
- Cost accounting for build CPU, cache transfer, runtime resources, telemetry,
  and model inference.
- AI eval judge calibration against expert labels, repeated sampling, confidence
  thresholds, and adversarial prompt/tool cases.
- Recorded replay versus live rerun of the same agent workflow.
- Capability-budget partitioning across agent handoffs and parallel children.
- Progression through the [capability gauntlet](capability-gauntlet.md) with no
  promotion based on happy-path execution alone.

## Governance questions

- Who may approve new syntax, core effects, unsafe powers, and official kits?
- What evidence and interaction analysis must a proposal contain?
- Can an experimental feature be removed cleanly, and is removal planned before
  adoption?
- How long are editions, runtime protocols, and official toolkit versions
  supported?
- What privacy-respecting ecosystem signals justify promotion into the official
  toolkit?
- How are security fixes allowed to override compatibility promises?

## Proposed next decision sequence

1. Begin the real lossless frontend and portable typed core. Carry the
   [bounded independent certificate result](../.planning/spikes/004-independent-certificate-checker/README.md)
   forward as a compact manifest and the [native contract
   harness](../.planning/spikes/005-native-ffi-provenance-cleanup/README.md) as
   its first lowering consumer. Retain both negative-control families.
2. Continue Stages 1–2 and the resource/async suites
   from the 64-workload [ownership and lifetime decisive
   study](ownership-and-lifetime-decisive-study.md), preserving the readiness
   audit's cross-cutting overlays.
3. Lower that same kernel through one minimal native path now; execute the blocking
   native/FFI ownership cases and establish differential outcome, event,
   initialization, provenance, and release evidence.
4. Complete the concurrency/graph and whole-program suites; select the
   ownership checker/surface contract from hard vetoes and measured evidence.
5. Preserve flow summaries in typed IR and run the 18 `FLOW-*` cases from the
   [trust boundary](trust-validation-and-information-flow.md).
6. Compare ordinary shared-value strategies and decide the implementation and
   admission threshold for the accepted optional managed actor/region.
7. Run the three-surface syntax and cognitive-ergonomics experiment over the
   surviving ownership model.
8. Run the Cranelift/QBE backend contest, adding LLVM only after the lowering
   harness is trustworthy.
9. Verify that the Lang corpus/evidence runner creates a development dividend;
   select it or the next-best native-capable G3 workload accordingly.
10. Set the assurance ladder: what blocks `check`, `verify`, and `release`.
11. Choose the context boundary: what propagates and what becomes explicit.
12. Define agent protocol v0 and stochastic evaluation evidence.
13. Ratify compiler/dev-build budgets from representative measurements.
14. Select the initial native backend from spike evidence; retain BEAM and Wasm
   as later conformance targets.
15. Define deterministic simulation and the first resilience/performance kits.
16. Build the service identity, JSON/schema, SQL/change, and contract slice.

The exhaustive dependency and decision-horizon view is in the [residual
uncertainty register](residual-uncertainty-register.md).
