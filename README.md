# ai-lang

An exploration of a general-purpose programming language for AI-authored,
human-audited software.

The current thesis is not that an AI-native language should use the fewest
characters. It should give an agent the shortest reliable path from intent to
evidence:

```text
generate -> verify -> run -> observe -> repair -> audit
```

The project is still in design exploration, with bounded semantic experiments
under `.planning/spikes`. Syntax and mechanisms shown here are candidate
designs, not settled specifications.

## Start here

- [Wiki conventions](wiki/README.md)
- [Vision and design principles](wiki/vision.md)
- [Converging design baseline](wiki/design-baseline.md)
- [Semantic kernel contract](wiki/semantic-kernel-contract.md)
- [Semantic kernel adversarial probe suite](wiki/semantic-kernel-probes.md)
- [Trust, validation, and information-flow boundary](wiki/trust-validation-and-information-flow.md)
- [Boundary data, validation, transactions, and pools](wiki/boundary-data-validation-and-persistence.md)
- [Language and ecosystem lifecycle backcast](wiki/language-ecosystem-lifecycle-backcast.md)
- [Ownership and lifetime decisive study](wiki/ownership-and-lifetime-decisive-study.md)
- [Ownership experiment readiness audit](wiki/ownership-experiment-readiness-audit.md)
- [Executable ownership kernel workbench](.planning/spikes/001-ownership-kernel-workbench/README.md)
- [Executable CFG edge-specific last-use workbench](.planning/spikes/002-cfg-edge-last-use/README.md)
- [Executable public-origins and generic-abilities workbench](.planning/spikes/003-public-origins-generic-abilities/README.md)
- [Executable independent ownership-certificate checker](.planning/spikes/004-independent-certificate-checker/README.md)
- [Executable native FFI, provenance, and cleanup harness](.planning/spikes/005-native-ffi-provenance-cleanup/README.md)
- [Ownership evidence roadmap](wiki/ownership-evidence-roadmap.md)
- [Real frontend and core-IR entry plan](wiki/real-frontend-core-ir-entry-plan.md)
- [Residual uncertainty and risk-reduction register](wiki/residual-uncertainty-register.md)
- [Author intent and predictive design model](wiki/author-intent-model.md)
- [Candidate language and toolchain primer](wiki/language-and-toolchain-primer.md)
- [Design-decision atlas](wiki/design-atlas.md)
- [Language lessons](wiki/language-lessons.md)
- [Feature coherence and admission framework](wiki/feature-coherence.md)
- [Effects and capabilities](wiki/effects-and-capabilities.md)
- [Type system and failure semantics](wiki/type-system-and-failure-semantics.md)
- [Syntax and cognitive ergonomics](wiki/syntax-and-cognitive-ergonomics.md)
- [Compiler and feedback latency](wiki/compiler-and-feedback-latency.md)
- [Compute-efficiency constitution](wiki/compute-efficiency-constitution.md)
- [Runtime profiles, scheduling, memory isolation, and dogfooding](wiki/runtime-profiles-and-dogfooding.md)
- [Memory reclamation and latency policy](wiki/memory-reclamation-policy.md)
- [Networking, HTTP, and TLS security boundary](wiki/networking-http-and-tls-security.md)
- [Application architecture and data boundaries](wiki/application-architecture-and-data.md)
- [Concurrency and memory model](wiki/concurrency-and-memory.md)
- [Domain, data, and distributed-system semantics](wiki/domain-data-and-distribution.md)
- [Blind spots and system boundaries](wiki/blind-spots-and-boundaries.md)
- [Context, telemetry, privacy, and operations](wiki/context-telemetry-and-security.md)
- [Performance, observability, profiling, and delivery](wiki/performance-observability-and-delivery.md)
- [Resources, locks, synchronization, and caching](wiki/resources-locks-and-caching.md)
- [Data, warehouse, and asset pipelines](wiki/data-and-asset-pipelines.md)
- [Modules, architecture conventions, and live development](wiki/modules-architecture-and-live-development.md)
- [AI-native runtime, agents, and evaluation](wiki/ai-native-runtime-and-evals.md)
- [Native and low-level programming profile](wiki/native-and-low-level-profile.md)
- [General-purpose capability gauntlet](wiki/capability-gauntlet.md)
- [Convergence audit](wiki/convergence-audit.md)
- [Experimental frontier and convergence work program](wiki/convergence-work-program.md)
- [Example language tour](wiki/example-tour.md)
- [Implementation path](wiki/implementation-path.md)
- [Research ledger](wiki/research-ledger.md)
- [Open questions](wiki/open-questions.md)
