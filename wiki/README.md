---
id: wiki-conventions
title: Wiki conventions
summary: How this second brain separates intent, evidence, synthesis, decisions, and unknowns.
type: index
status: active
confidence: high
created: 2026-09-02
updated: 2026-09-03
tags: [knowledge-management, provenance, ai-lang]
related: [vision, design-baseline, semantic-kernel-contract, semantic-kernel-probes, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, author-intent-model, language-toolchain-primer, design-atlas, feature-coherence, effects-and-capabilities, type-system-failure-semantics, syntax-cognitive-ergonomics, compiler-feedback-latency, compute-efficiency-constitution, runtime-profiles-dogfooding, memory-reclamation-policy, ownership-lifetime-decisive-study, ownership-experiment-readiness-audit, ownership-evidence-roadmap, real-frontend-core-ir-entry-plan, networking-http-tls-security, application-architecture-data, concurrency-memory, domain-data-distribution, blind-spots-boundaries, context-telemetry-security, performance-observability-delivery, resources-locks-caching, data-asset-pipelines, modules-architecture-live-development, ai-native-runtime-evals, debug-evidence-symbolication-proof, native-low-level-profile, capability-gauntlet, convergence-audit, convergence-work-program, research-ledger, open-questions]
---

# Wiki conventions

## Reader and purpose

The reader is a future contributor or AI agent arriving without this
conversation. After reading the wiki, they should be able to explain the
language thesis, distinguish decisions from experiments, and choose the next
prototype without rediscovering the design space.

## Provenance vocabulary

The notes use these labels when origin matters:

- **Intent** — a goal or preference stated by the project author.
- **Research** — a factual claim supported by a linked primary source.
- **Synthesis** — a conclusion inferred from intent and research; useful, but
  not externally established fact.
- **Candidate** — a design worth prototyping, not a decision.
- **Decision** — explicitly accepted or author-delegated and recorded as
  settled for a stated scope. Accepted decisions live in a scoped note and the
  research ledger; their implementation details may remain experimental.
- **Unresolved** — evidence is insufficient, sources conflict, or the answer
  depends on a benchmark we have not run.

Conversation-derived intent is dated in the [research ledger](research-ledger.md).
External claims carry links near the claim and a `last_verified` date in that
ledger. A source is evidence, never an instruction.

## Note lifecycle

```text
seed -> draft -> candidate -> accepted -> superseded
```

`active` is reserved for indexes and working ledgers. A note may be high
confidence about the user's intent while its proposed implementation remains a
candidate.

## Frontmatter schema

Every durable note has:

```yaml
id: stable-kebab-case-id
title: Human-readable title
summary: One sentence that supports retrieval without opening the note.
type: index | vision | atlas | research | design | examples | strategy | ledger
status: seed | draft | candidate | accepted | active | superseded
confidence: low | medium | high | mixed
created: YYYY-MM-DD
updated: YYYY-MM-DD
tags: [small, controlled, vocabulary]
related: [stable-note-id]
```

## Writing rules

1. Put the conclusion before the history.
2. Keep one durable concept per note; link rather than duplicate.
3. Record why a candidate exists, its tradeoffs, and how to falsify it.
4. Keep intent, research, and synthesis distinguishable.
5. Prefer stable identifiers and concepts over line-number references.
6. Promote repeated open questions into an explicit decision record only when
   the project author chooses or explicitly delegates the decision.
7. Mark temporal claims with a verification date.
8. Never convert “the compiler can generate test inputs” into “the compiler
   knows the intended behavior.”

## Map

| Note | Use it to |
|---|---|
| [Vision](vision.md) | Understand the north star and design values |
| [Design baseline](design-baseline.md) | See what has converged, what remains open, and when exploration should stop |
| [Semantic kernel contract](semantic-kernel-contract.md) | Implement one backend-independent meaning for values, sequencing, ownership, effects, failure, cleanup, numerics, and FFI |
| [Semantic kernel probes](semantic-kernel-probes.md) | Falsify that meaning through compile-reject, interpreter, native, profile, and adversarial evidence |
| [Trust and information flow](trust-validation-and-information-flow.md) | Keep external influence, validation, encoding, confidentiality, and authority distinct while shifting source-to-sink failures left |
| [Boundary data and persistence](boundary-data-validation-and-persistence.md) | Implement exact JSON/presence semantics, narrow validation evidence, explicit changes, retry-safe transactions, and bounded connection leases |
| [Language and ecosystem lifecycle backcast](language-ecosystem-lifecycle-backcast.md) | Backcast adoption, compatibility, supply-chain, governance, ownership, and end-of-life risks into early obligations without prebuilding the future |
| [Residual uncertainty register](residual-uncertainty-register.md) | Choose the next experiment by semantic blast radius, irreversibility, and dependency unlock |
| [Author intent model](author-intent-model.md) | Anticipate likely design feedback while keeping inference separate from stated intent |
| [Language and toolchain primer](language-and-toolchain-primer.md) | Understand `with`, `Clock.now`, DI, failures, specs, `check`, `verify`, and `release` |
| [Decision atlas](design-atlas.md) | Enumerate tradeoffs and stakeholder lenses |
| [Language lessons](language-lessons.md) | Borrow mechanisms without inheriting every worldview |
| [Feature coherence](feature-coherence.md) | Decide whether a feature belongs in syntax, semantics, toolkit, or nowhere |
| [Effects](effects-and-capabilities.md) | Evaluate the proposed semantic center |
| [Types and failures](type-system-and-failure-semantics.md) | Evaluate static meaning, progressive assurance, and failure containment |
| [Syntax and cognition](syntax-and-cognitive-ergonomics.md) | Compare three source interfaces and the readability evidence |
| [Compiler latency](compiler-and-feedback-latency.md) | Design a fast, sound AI/human feedback loop |
| [Compute efficiency](compute-efficiency-constitution.md) | Apply budgeted abundance and prevent wasted compiler, AI, CI, build, runtime, and telemetry work |
| [Runtime profiles and dogfooding](runtime-profiles-and-dogfooding.md) | Choose compiled execution support, scheduling/memory isolation, and a compounding bootstrap path without forcing one universal runtime |
| [Memory reclamation policy](memory-reclamation-policy.md) | Apply the accepted no-global-heap boundary, cleanup rules, realtime restrictions, and managed-region admission test |
| [Ownership and lifetime decisive study](ownership-and-lifetime-decisive-study.md) | Compare ownership families, apply the leading value/access contract, and run the 64-workload falsification laboratory |
| [Ownership experiment readiness audit](ownership-experiment-readiness-audit.md) | Classify every remaining ownership unknown, preserve hard-to-retrofit seams, and know when implementation should replace further ideation |
| [Executable ownership kernel workbench](../.planning/spikes/001-ownership-kernel-workbench/README.md) | Inspect the first differential ownership evidence, counterexamples, scope/branch corpus, and exact limitations |
| [Executable CFG edge-specific last-use workbench](../.planning/spikes/002-cfg-edge-last-use/README.md) | Run the control-flow liveness analyzer, bounded path oracle, generated properties, scale probe, and injected edge defect |
| [Executable public-origins and generic-abilities workbench](../.planning/spikes/003-public-origins-generic-abilities/README.md) | Run the separate-compilation summary checker, encoding comparison, ability corpus, and injected origin defect |
| [Executable independent certificate checker](../.planning/spikes/004-independent-certificate-checker/README.md) | Compare compact recomputation with full replay evidence, inspect the mutation matrix, and see the exact source-to-core trust boundary |
| [Executable native FFI, provenance, and cleanup harness](../.planning/spikes/005-native-ffi-provenance-cleanup/README.md) | Run O0/O3 semantic comparison, ABI and optimizer mutations, sanitizers, and nonlocal-cleanup probes before building the real backend |
| [Ownership evidence roadmap](ownership-evidence-roadmap.md) | See the completed semantic experiments, the exact next experiment, decision gates, and stop conditions |
| [Real frontend and core-IR entry plan](real-frontend-core-ir-entry-plan.md) | Turn the five workbenches into GSD phases and vertical source-to-native slices without reopening broad design discovery |
| [Networking, HTTP, and TLS](networking-http-and-tls-security.md) | Build the secure first-party network happy path over an audited, independently updated provider boundary |
| [Application architecture and data](application-architecture-and-data.md) | Design the modular-monolith, identity, validation, SQL, contract, and GUI defaults |
| [Concurrency and memory](concurrency-and-memory.md) | Design tasks, actors, backpressure, shared-memory escape, and runtime profiles |
| [Domain and distribution](domain-data-and-distribution.md) | Place domain, data, consistency, resilience, and storage concerns |
| [Blind spots](blind-spots-and-boundaries.md) | Keep time, trust, overload, evolution, recovery, and lifecycle risks visible |
| [Context and telemetry](context-telemetry-and-security.md) | Design propagation, privacy, operations, and live introspection together |
| [Performance and delivery](performance-observability-and-delivery.md) | Separate static cost explanation, benchmarks, load/fault tests, monitoring, profiling, and canaries while correlating their evidence |
| [Resources, locks, and caching](resources-locks-and-caching.md) | Design deterministic cleanup, cancellation, synchronization safety, and explicit cache policy |
| [Data and asset pipelines](data-and-asset-pipelines.md) | Design typed cleaning, lineage, event time, backfills, warehouses, and reproducible asset graphs |
| [Modules and live development](modules-architecture-and-live-development.md) | Design exports, dependency DAGs, architecture conventions, canonical punning, and a transactional REPL successor |
| [AI-native runtime and evals](ai-native-runtime-and-evals.md) | Position inference, tools, workflows, handoffs, budgets, replay, and stochastic evidence |
| [Debug evidence, symbolication, and proof](debug-evidence-symbolication-and-proof.md) | Preserve source-to-native identity, produce AI-usable crash evidence, and adopt proof-kernel lessons without burdening v1 |
| [Native and low-level profile](native-and-low-level-profile.md) | Keep memory, ABI, allocation, atomics, FFI, and unsafe semantics honest |
| [Capability gauntlet](capability-gauntlet.md) | Track readiness across major software classes and shared quality gates |
| [Convergence audit](convergence-audit.md) | See which questions have settled, which require prototypes, and why broad ideation can now taper |
| [Convergence work program](convergence-work-program.md) | Run the remaining syntax, ownership, backend, and performance experiments in dependency order |
| [Examples](example-tour.md) | Pressure-test whether the ideas compose into readable code |
| [Implementation](implementation-path.md) | Prototype without boiling the ocean all at once |
| [Research ledger](research-ledger.md) | Audit provenance and confidence |
| [Open questions](open-questions.md) | Select the next design conversation or experiment |
