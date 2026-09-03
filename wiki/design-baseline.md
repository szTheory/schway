---
id: design-baseline
title: Converging design baseline
summary: The current coherent language-system shape, its confidence level, and the few decisions that still need experiments or author choice.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [baseline, convergence, architecture, roadmap]
related: [vision, author-intent-model, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, type-system-failure-semantics, syntax-cognitive-ergonomics, effects-and-capabilities, application-architecture-data, concurrency-memory, memory-reclamation-policy, networking-http-tls-security, runtime-profiles-dogfooding, resources-locks-caching, modules-architecture-live-development, performance-observability-delivery, data-asset-pipelines, ai-native-runtime-evals, native-low-level-profile, compiler-feedback-latency, compute-efficiency-constitution, capability-gauntlet, convergence-audit, convergence-work-program, implementation-path, open-questions]
---

# Converging design baseline

## Assessment

The design has reached **semantic convergence, not specification freeze**.
Adding UUIDs, modular monoliths, validation, contract tests, state machines, GUI,
and memory concerns did not require replacing the center. Each fits through the
same small set of mechanisms: nominal data, effects/capabilities, typed failure,
contracts, components, structured work, providers, and inspectable derivation.

That is the strongest cohesion signal so far. New ideas are increasingly
changing toolkit placement and policy rather than forcing new primitive
semantics.

## Provisional constitution

| Area | Current baseline | Confidence |
|---|---|---|
| Author/auditor | AI authors; expert humans review and diagnose | high intent |
| Optimization target | Minimum time/tokens/cost from intent to verified change | high synthesis |
| Compute doctrine | Budgeted abundance: spend for information/correctness/value; eliminate redundant or mis-scoped work across AI, compiler, CI, and runtime | high intent |
| Source | One formatter-owned human projection over a stable semantic model | high direction, surface open |
| Values/data | Immutable by default; ADTs; opaque nominal domain types; no general null | high |
| Behavior | Small consumer-owned structural ports; explicit public contracts | medium-high |
| Executable semantics | One strict kernel contract plus an adversarial differential probe suite | high direction, implementation required |
| Effects/DI | Typed authority rows plus statically resolved, non-resumable capability providers in v1 | medium-high; prototype required |
| Errors | `Option`/`Result`; cancellation separate; defect panic at smallest supervised boundary | medium-high |
| Assurance | Types first, then contracts/properties/model checks/proofs in progressive lanes | high direction, gates open |
| Architecture | Verified modular monolith default over a general component DAG | medium-high |
| Concurrency | Structured tasks, supervised actors, bounded streams | medium-high |
| Memory | Accepted ownership-first boundary; value/access surface over affine place/loan/origin semantics leads; stack/unique/arena baseline; explicit sharing and owner-confined managed graphs | high architecture, detailed checker/surface/storage experimental |
| Runtime | Compiled-first monotonic profiles: freestanding/native-core omit unused scheduler/GC services; application/service add a shared runtime; interpreter is the oracle; BEAM/Wasm are later conformance targets | high direction, implementation open |
| Native profile | V1 must support useful owned values, inferred local borrows, lexical resources, C interop, allocators, and checked release semantics | high intent, prototype required |
| Interop | Typed C boundary and Wasm component boundary; foreign claims remain marked | high direction |
| Trust flow | Typed safe sinks and boundary validation plus IR source/transform/sink summaries; targeted global/control-flow checks in stronger lanes; no Boolean production taint bit | medium-high direction, precision and ergonomics experimental |
| Operations | Typed causal runtime events; structured agent query/edit/verify protocol | high direction |
| Performance | Static cost explanations plus correlated benchmarks, load/fault tests, profiles, SLOs, and canary evidence | medium-high direction |
| Caching | Core purity/effect/version/authority hooks plus an official typed policy kit; no automatic or dedicated v1 cache syntax | medium-high direction, prevention coverage open |
| Dogfood | Corpus and standard modules first; Lang corpus/evidence runner leads; full compiler self-hosting later | high direction, workload selection measured |
| Modules/live use | Private-by-default explicit exports, acyclic component/module dependencies, no effectful top-level initialization, transactional live workspace | medium-high |
| AI systems | Models and tools are typed effects; agent graphs are projections; evals are stochastic evidence | medium-high direction |
| Toolkit | First-party JSON, HTTP, SQL/change, contracts, workflows, resilience, UI hosts | medium placement |
| Boundary data | Exact presence and number policy; boundary-specific unknown-field handling; decode/change/verified states; retry-safe transactions; bounded affine pool leases | medium-high direction, kit implementation open |
| Identity | Opaque `Id<Entity>`; UUIDv7 candidate service provider | medium |
| Evolution | Editions plus separate compiler/runtime/toolkit/protocol versions and migrations | high direction |
| Lifecycle | Versioned evidence, hermetic/content-addressed builds, impact testing, retraction without deletion, staged bootstrap, and tiered governance seams begin before stability | high direction, operating policy later |

## Small semantic center

```text
data + functions + modules
  + nominal identities and validated values
  + structural behavioral ports
  + effects/capabilities/providers
  + Result/Option/panic boundaries
  + contracts and evidence
  + tasks/actors/streams
  + components and versioned boundaries
```

JSON, SQL, HTTP, RPC, ORM-like changes, contract tests, state machines, sagas,
AI inference, GUI, containers, and telemetry exporters are built from or project
through that center. They do not each receive unrelated magic.

## Decisions that appear stable enough to prototype

- Algebraic data types and exhaustive pattern matching.
- Immutable values and pure guards by default.
- Nominal public/domain data plus structural behavioral ports.
- Named multi-role arguments with exact-name punning as an experiment.
- The formatter canonicalizes an exact unqualified `name: name` argument to
  `name:`; the parser accepts the expanded form and rejects ambiguous punning.
- Explicit public input/output/error/effect shapes with local inference.
- No arbitrary implicit conversion, truthiness, wildcard imports, or exception
  control flow.
- Typed effects/capabilities as the authority and substitution spine.
- Compile-time provider graph rather than service locator/reflection DI.
- No general multi-shot continuation handlers in v1; direct capability calls
  preserve owned resources and keep DI/runtime mechanics small.
- Strict left-to-right evaluation, including named arguments in source order.
- Public borrow/take modes and returned origins over local flow-sensitive
  inference; visible named-value transfer/exclusive access now leads the
  call-site experiment. No partial residual moves, escaping local borrows, or
  ordinary borrows across `await` in v1.
- Split correctness-significant `finish` from restricted automatic `release` so
  cleanup cannot erase a primary result or panic during unwinding.
- Structured task lifetime and supervised long-lived actors.
- Lexically owned resources with deterministic cleanup; ordinary lock guards
  cannot cross suspension points.
- No mandatory global tracing heap; implicit memory destruction cannot run
  arbitrary user effects; managed cyclic graphs are owner-confined and opt-in.
- Private-by-default modules, explicit exports/imports, acyclic dependency
  graphs, and effect-free top-level initialization.
- Canonical formatting and structured diagnostics/semantic queries.
- Opinionated verified modular monolith as the default service architecture.
- Official kits rather than framework-specific core grammar.
- Flat, independent, ordinary-code specifications.
- A fast affected edit lane distinct from exhaustive release verification.
- A typed transactional live workspace that records cells, effects, providers,
  results, and promotion into durable source/evidence.
- Typed ingress, nominal validation, destination-specific sinks, classified
  egress, and capability authorization; validity never silently becomes
  authority.

## Decisions still capable of changing the surface substantially

1. Braced conventional syntax versus indentation-oriented syntax; the uniform
   tree is likely an agent projection.
2. Plain effect calls versus a visible marker such as `Clock.now!()`.
3. Exact named-argument exceptions for unary algebraic/pipeline calls.
4. Whether text/Git or a semantic code database is authoritative.
5. How public effects are rendered when inferable.
6. Whether a declarative state machine becomes syntax or an editor projection.

These are now evaluation questions over a stable semantic corpus, not reasons
to keep ideating indefinitely.

## Decisions that need one strategic author choice

- Whether source text or a semantic code database is authoritative.
- Whether missing public evidence blocks ordinary `check` or only
  `verify`/`release`.
- Whether verified principal/tenant context may propagate implicitly beyond
  ingress or must become explicit application values immediately.
- Whether UUIDv7 is automatically bound in the service profile or selected by
  the application template.
- Whether GUI appears before or after the first production service proof.

The first vertical workload, shared-value/managed-region implementation,
syntax, backend, runtime safe-point strategy, and cache policy belong to
experiments rather than author preference. The current dogfood candidate is
evaluated by measured development dividend.

## Stop conditions for exploration

Begin the bounded Experiment 0/0b implementation plan now. End broad semantic
exploration and promote the experiments into a committed public-v0 roadmap
when:

1. one first workload is chosen for the now-settled native-capable v1 profile;
2. the assurance-gate philosophy is chosen;
3. the three source surfaces are rendered over the same 15–25 corpus programs;
4. the [semantic kernel contract](semantic-kernel-contract.md) and its
   [probe suite](semantic-kernel-probes.md) pass in one reference interpreter
   and one native lowering;
5. the agent protocol v0 can answer dependency, effect, failure, affected-test,
   and runtime-cause questions;
6. warm latency budgets are measured rather than aspirational;
7. every row in the [capability gauntlet](capability-gauntlet.md) maps to a
   named mechanism or explicit non-goal, and the chosen row has G3 criteria.

Further broad language research after those conditions should be demand-driven.
The next compounding dividend comes from executable evidence, not a longer
feature list.

The [language and ecosystem lifecycle backcast](language-ecosystem-lifecycle-backcast.md)
adds the governing reopen rule: a new concern changes the kernel only when
independent workloads, a blocking probe, unavoidable unsafe/global behavior,
failed budgets, cross-implementation ambiguity, or repeated adopter confusion
shows that it must. Otherwise it becomes a kit, provider, policy, or gauntlet
item.

The bounded path through the remaining uncertainty is specified in the
[residual uncertainty register](residual-uncertainty-register.md) and
[convergence work program](convergence-work-program.md). The Lang corpus/evidence
runner is its leading first-workload candidate, not a selected roadmap
commitment; it contains the streaming JSON/data workload.
