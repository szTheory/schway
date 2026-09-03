---
id: author-intent-model
title: Author intent and predictive design model
summary: A confidence-labelled model of the author's design instincts, likely objections, and safeguards against overfitting to inferred preferences.
type: strategy
status: candidate
confidence: mixed
created: 2026-09-02
updated: 2026-09-03
tags: [intent, design, prediction, governance]
related: [vision, design-baseline, feature-coherence, ownership-lifetime-decisive-study, ownership-experiment-readiness-audit, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, runtime-profiles-dogfooding, capability-gauntlet, open-questions, research-ledger]
---

# Author intent and predictive design model

## Purpose

This note lets a future collaborator anticipate likely feedback without
pretending to know the author's mind. Explicit statements remain **intent**;
everything extrapolated from their trajectory is **synthesis** with a
confidence label. A prediction may guide a reversible prototype. It may not
silently settle a consequential design choice.

## The trajectory

The conversation has tightened in thirteen passes:

1. **AI and human legibility.** Begin with an expressive general-purpose
   language whose source is concise, elegant, canonical, and easy for AI to
   generate and humans to audit.
2. **Closed feedback loop.** Treat fast checking, hot reload, live runtime
   introspection, telemetry, CLI, and machine-readable protocols as part of the
   language—not editor accessories.
3. **Correctness by construction.** Pull functional design, effects, DI,
   testing, architecture boundaries, explicit failures, supervision, and
   concurrency safety into one coherent semantic model.
4. **Whole lifecycle.** Include dependency hygiene, configuration, security,
   CI, containers, distribution, upgrades, recovery, databases, contracts, and
   production operations in the design pressure.
5. **Practical ambition.** Borrow the strongest lessons from existing
   languages, ship incrementally using interop and a strangler strategy, but do
   not let the first backend foreclose native, embedded, GUI, or distributed
   work.
6. **Evidence and convergence.** Enumerate blind spots, pressure-test realistic
   software classes, explain the design accessibly, and stop expanding when
   executable evidence becomes more informative than more speculation.
7. **Native honesty now, not later.** Accept additional early compiler burden so
   ownership, resource programming, C boundaries, and useful native execution
   are present in the first implementation rather than protected only by future
   compatibility promises.
8. **Budgeted abundance.** Do not ration worthwhile computation, but make the
   compiler, AI workflow, CI, runtime, and telemetry mechanically sympathetic
   and intolerant of redundant or poorly scoped work from their first design.
9. **Compounding implementation.** Keep the production path compiled and the
   runtime proportional to the selected workload; borrow BEAM's isolation,
   preemption, local reclamation, supervision, and inspection where valuable;
   move high-leverage development tools into Lang early without forcing fragile
   full self-hosting.
10. **Boundary truth before framework convenience.** Revisit apparently mundane
    facilities—taint, JSON, validation, changes, transactions, and pools—to make
    sure their exact safety, evolution, overload, and runtime contracts fit the
    core before implementation makes them expensive to change.
11. **Lifecycle backcast before commitment.** Simulate adoption, incidents,
    ecosystem growth, compatibility, governance, succession, and legacy use;
    preserve hard-to-retrofit seams now while refusing to prebuild reversible
    future machinery.
12. **Ownership must survive reality before syntax freezes.** Decompose
    ownership into value authority, reference validity, reclamation, resource
    protocols, and concurrency isolation; compare genuinely different models
    across realistic success, failure, boundary, native, and lifecycle cases;
    then select the smallest coherent contract that passes hard vetoes.
13. **Bounded certainty, then autonomous evidence.** Run one last systematic
    unknowns audit, classify every remaining risk by horizon and trigger, then
    proceed without waiting through the smallest reversible experiment that can
    falsify the design.

The pattern is not “add every feature.” It is “find a small number of
mechanisms that make good engineering the path of least resistance across the
entire lifecycle.”

## Stable preference model

| Predicted preference | Confidence | Observable implication |
|---|---:|---|
| Constrained power over unconstrained convenience | high | Powerful features are welcome when authority, cost, and failure remain visible. |
| Local reasoning over hidden indirection | high | Avoid service locators, implicit globals, distant configuration, magical callbacks, and deep inheritance. |
| Fast, accurate feedback over maximal compile-time cleverness | high | Bound analysis, cache aggressively, and move expensive assurance into explicit lanes. |
| Deliberate compute over either stinginess or waste | high | Measure total wall/CPU/memory/disk/network/token cost and spend more only for visible correctness, evidence, or user value. |
| One coherent system over a bag of fashionable features | high | New syntax must compose with types, effects, failures, concurrency, tools, and upgrades. |
| Modern contextual defaults over universal dogma | high | UUIDv7, actors, or GC may be excellent profile defaults without becoming mandatory everywhere. |
| Practical interop and incremental adoption over purity | high | C, Wasm, host runtimes, and foreign packages are bridges with typed risk boundaries. |
| Modular monolith before distributed decomposition | high | Make ownership and consistency explicit in-process; preserve an honest extraction path. |
| Evidence over ceremony | high | Specs, contracts, traces, and evals must catch faults, not merely satisfy a policy count. |
| Calm canonical source over personal formatting freedom | high | The formatter owns layout and semantic tools explain hidden detail on demand. |
| Boring semantic core plus rich official kits | medium-high | JSON, HTTP, SQL, workflows, agents, and UI should mostly be derived libraries, not unrelated grammar. |
| Profiles over one semantic compromise for every workload | medium-high | Service, native, Wasm, and embedded profiles share values/contracts while admitting honest restrictions. |
| Foundational constraints must be executable early | high | If later low-level capability could overturn the IR, type system, or runtime assumptions, prove a useful subset in v1. |
| Runtime cost proportional to use over one universal runtime | high | Freestanding and sequential native programs should not pay for actors, schedulers, or tracing GC; service profiles may opt into them. |
| Compounding dogfood over symbolic self-hosting | high | Write corpus, evidence, diagnostics, cache, and build tools in Lang as soon as useful; migrate the compiler only when staged bootstrap is safe. |
| A stricter release lane than edit lane | medium-high | `check` stays sound and fast; `verify`/`release` enforce broader evidence policy. |
| Unsafe power should be acknowledged and contained, not hidden | medium-high | Unsafe modules need contracts, audits, tests, and visible transitive risk. |
| Human source and machine semantic graph are complementary | medium | Humans review canonical text; agents query stable identities, dependencies, obligations, and runtime facts. |
| Quiet values over invisible semantic cliffs | high | Ordinary reads stay terse, but named noncopyable transfer, exclusive access, escape, sharing, and fallible completion should be source-auditable unless experiments show the markers cost more than they prevent. |
| Bounded autonomous progress over either hesitation or runaway scope | high | Once risks have owners, defaults, probes, and stop rules, implement the smallest evidence-producing slice and leave an honest handoff without waiting for routine decisions. |

## Feedback this model predicts next

A proposal is likely to be rejected or revised if it has one of these traits:

- a pleasant surface hides allocation, blocking, retry, network, authority, or
  nondeterminism;
- an escape hatch quietly becomes the normal path;
- generated code satisfies a checklist while remaining semantically weak;
- release optimizations weaken safety without an explicit operation;
- a feature works on a snippet but creates poor diffs, upgrades, incident
  response, or Day Two operations;
- an official abstraction has a performance cliff that tools cannot explain;
- a compiler feature makes clean builds, cache misses, or CI unacceptably slow;
- “portable” means lowest-common-denominator behavior or undocumented backend
  divergence;
- a dependency or macro compromises reproducibility, provenance, or the trusted
  computing base;
- a versioning story covers source APIs but ignores schemas, durable state,
  protocols, data migrations, mixed-version deployment, or rollback;
- AI evals collapse stochastic evidence into a misleading green check;
- replay claims determinism while rerunning nondeterministic model or network
  calls;
- agents gain authority from text, remote annotations, or ambient credentials;
- low-level capability is promised without defining aliasing, layout, OOM,
  atomics, FFI, interrupts, and unsafe containment;
- a design is called general-purpose before representative workloads compile,
  run, fail, recover, upgrade, and remain inspectable.
- decoding, validation, authorization, persistence, or publication is described
  as one generic “safe” transition;
- strict parsing destroys forward-compatible data, or permissive parsing hides
  misspelled commands and lossy coercions;
- transaction retry can repeat a model, network, clock, random, email, payment,
  or other external effect;
- connection pooling is treated as invisible plumbing despite queueing,
  ownership, cancellation, fairness, and overload behavior;
- a typed capability is assumed to remain valid forever despite expiry,
  revocation, provider shutdown, or narrower delegated authority.
- “ownership” is treated as one borrow-checking choice while reclamation,
  resource completion, actor transfer, pointer provenance, and public origin
  contracts remain implicit;
- a local refactor silently introduces a deep clone, retain/release traffic,
  longer retention, or use-after-transfer semantics;
- an ergonomic ownership proposal is judged only on toy snippets rather than
  noncopyable generics, callbacks, suspension, partial initialization, graphs,
  FFI, and mixed-version public APIs.

## Newly illuminated blind spots

These concerns were not always named first, but follow directly from the
trajectory:

1. **Compiler trust and bootstrap.** Reproducible builds, a specified IR,
   conformance suites, deterministic output, and possibly a small independent
   certificate checker matter if compiler promises gate production.
2. **Supply-chain authority.** Package install scripts, procedural macros,
   compiler plugins, native libraries, and build scripts are effects and need
   capabilities, provenance, sandboxing, and budgets.
3. **Performance predictability.** “Fast” also means avoiding invisible
   allocation, copies, scheduler hops, boxing, reflection, and accidental
   quadratic compilation. Explanations matter as much as peak benchmarks.
4. **Evolution as a typed operation.** Source, wire schema, database schema,
   actor state, workflow history, capability sets, and telemetry contracts
   evolve in different directions and on different clocks.
5. **Stochastic evidence.** AI behavior needs distributions, confidence,
   adversarial datasets, judge calibration, and continuous evaluation rather
   than deterministic-test theatre.
6. **Information integrity.** Retrieved text, prompts, model output, tool output,
   and foreign metadata are untrusted data. They cannot create authority.
7. **Resource algebra.** Time, memory, tokens, money, tool calls, queue depth,
   and concurrency budgets must compose across child tasks and agents.
8. **Semantic accessibility.** A design that only its inventors can explain has
   failed. Every unusual keyword needs a compact mental model, queryable
   expansion, and precise diagnostic.
9. **Governance and removal.** Experimental syntax, unsafe powers, official
   packages, and runtime hooks need admission, deprecation, and removal rules.
10. **Adoption economics.** Installation, editor support, debugging, packaging,
    upgrades, hiring, interoperability, and migration determine whether the
    language earns real feedback.
11. **Cache semantics as correctness.** Keys, authority partitions, freshness,
    versions, invalidation order, fill concurrency, and capacity are part of the
    behavior being reviewed; automatic caching without those contracts will be
    rejected as hidden mutable state.
12. **Runtime proportionality.** “Compiled” and “runtime-backed” are independent
    axes. A credible general-purpose language needs a small sequential path and
    an earned application/service runtime rather than one accidental compromise.
13. **Boundary protocols are semantic transformations.** Parsing, typed decode,
    field presence, defaulting, canonicalization, validation, authorization,
    persistence, and publication establish different facts and require stable
    evidence identities.
14. **Authority has a lifecycle.** Capability scope prevents ambient access, but
    long-lived and delegated handles also need expiry, attenuation, revocation,
    shutdown, and typed failure behavior.
15. **Resource schedulers hide in ordinary libraries.** Database pools, HTTP
    clients, executors, model quotas, and caches all contain queues and fairness
    policy; the language should expose their bounded resource contract without
    putting each scheduler into grammar.
16. **Compatibility and strictness are directional.** Rejecting unknown command
    fields, tolerating additive response fields, preserving relay data, and
    signing canonical bytes are different boundary policies rather than one
    global “strict mode.”
17. **Reproducibility is not authenticity or freshness.** Lockfiles and
    deterministic output need verified build provenance, trusted update
    expectations, retraction, and rollback/freeze resistance around them.
18. **Success creates semantic pressure.** Compatibility, deprecation,
    ecosystem-impact testing, maintainer succession, and unsupported-kit
    retirement must be designed before every accidental behavior becomes a
    constituency.
19. **Model independence is part of AI-native design.** Canonical examples and
    semantic tools should help current agents, but compiler, build, evidence,
    and human audit must remain useful if a model or provider disappears.
20. **Research itself needs a stop rule.** General-purpose domain discovery is
    inexhaustible; only repeated workload evidence or a failed semantic probe
    earns new kernel complexity after convergence.
21. **Ownership needs two vocabularies, not two semantics.** Humans should see
    owned values, temporary access, transfer, sharing, and obligations. The
    compiler still needs precise places, loans, origins, initialization states,
    isolation domains, and provenance. Exposing all IR vocabulary is noisy;
    omitting it from the implementation makes native and diagnostic promises
    unverifiable.
22. **Unknowns need horizons.** “Not decided” is only safe when the project also
    records whether the fact must be fixed, preserved in IR, measured before
    syntax/native release, or deferred behind a concrete reopening trigger.

## Design response

The current response is a four-layer system:

```text
small semantic core
  -> profile-specific execution rules
  -> versioned official kits and providers
  -> evidence-driven project policies
```

- The core owns types, effects, failures, modules, components, tasks, resources,
  contracts, and stable semantic identities.
- Profiles define legal allocation, concurrency, runtime, ABI, and deployment
  behavior for service, native, Wasm, and embedded contexts.
- Official kits implement fast-changing domain affordances such as HTTP, SQL,
  workflows, AI tools/evals, and UI without freezing them into grammar.
- Project policy chooses budgets and release evidence without making ordinary
  editing unusably slow.

## Anti-overfitting rules

“Mental clone” is useful only as prediction under uncertainty.

- Preserve the original statement and the extrapolation separately.
- Attach confidence and a falsifying observation to consequential predictions.
- Prefer reversible prototypes when confidence is below high.
- Ask the author when a choice changes the semantic center, target scope, or
  irreversible implementation sequence.
- Do not optimize for the author's aesthetic preferences at the expense of
  evidence from representative users and workloads.
- Record surprises; they improve this model more than confirmations do.

## Pre-mortem

If the project fails despite good ideas, the most likely causes are scope
dispersion, a beautiful unimplemented design, slow compiler feedback, effect or
ownership notation that overwhelms ordinary code, unsafe/foreign boundaries
that invalidate guarantees, and insufficient real adopter pressure. The
[capability gauntlet](capability-gauntlet.md) and
[implementation path](implementation-path.md) exist to make those failures
visible early.
