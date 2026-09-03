---
id: convergence-audit
title: Convergence audit and remaining proof obligations
summary: A cold assessment of the design's stable center, closed blind spots, residual risks, and the shortest path from exploration to executable evidence.
type: strategy
status: candidate
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [convergence, risks, decisions, evidence]
related: [design-baseline, author-intent-model, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, ownership-experiment-readiness-audit, trust-validation-information-flow, language-ecosystem-lifecycle-backcast, residual-uncertainty-register, compiler-feedback-latency, compute-efficiency-constitution, runtime-profiles-dogfooding, memory-reclamation-policy, networking-http-tls-security, capability-gauntlet, convergence-work-program, implementation-path, open-questions, research-ledger]
---

# Convergence audit and remaining proof obligations

## Verdict

The project has a coherent **candidate semantic and architectural contract**
and five bounded executable ownership evidence layers. It is not yet converged at
the full semantic, syntax, native backend, runtime-performance, or adoption
level.

The executable workbenches now cover the linear move/loan/release state
machine, whole-value joins and loop fixed points, CFG edge-specific last use,
body-blind public origin/access/ability summaries, and independent validation of
committed typed-core summaries/abilities/flows. That retires “the design is
only prose” for these slices and precisely exposes one residual trust boundary:
a downstream checker cannot prove the frontend emitted a truthful typed-core
statement. A real C/Clang harness now also exposes ABI, alias, provenance,
cleanup, and nonlocal-exit hazards, but no Lang IR produces that C yet. The work
does not establish variance, async, a real native lowering,
incremental compilation, or the human surface.

The language no longer needs another broad feature harvest before work starts.
Nearly every new concern now lands in one of six established mechanisms:

```text
typed values and failures
  + effects, capabilities, and providers
  + ownership, lexical resources, and structured lifetimes
  + modules, components, and verified dependency graphs
  + contracts, evidence, and stable semantic identities
  + profile-specific runtimes and independently versioned official kits
```

This is a point of diminishing returns for unconstrained ideation, not a reason
to stop research. Research should now answer a prototype decision or explain an
observed failure.

A full [language and ecosystem lifecycle
backcast](language-ecosystem-lifecycle-backcast.md) reached the same conclusion
after simulating bootstrap, first production, ecosystem growth, 1.0
compatibility, scale, compromise, governance, and legacy operation. It added
early lifecycle seams and probes, not another kernel feature.

## Largest remaining gaps

The largest risk is now the distance between an elegant paper design and an
executable semantic kernel. The gaps are ranked by blast radius, present
uncertainty, and how many later decisions their evidence unlocks.

| Rank | Gap | Why it is dangerous | Smallest decisive evidence |
|---:|---|---|---|
| 1 | build the real source-to-core-to-native spine | independent typed-core validation and native/FFI hazards are now executable separately, but no Lang IR connects them | implement lossless minimal source, portable typed core, interpreter, and one native lowering that consumes the existing manifests and hostile probes |
| 2 | ownership and reclamation ergonomics | a safe model can still create distant errors, copies, RC contention, or release tails | 64-workload ownership laboratory spanning local flow, generics, async/resources, graphs/concurrency, FFI/native, and evolution |
| 3 | compiler architecture and feedback latency | a correct language that answers slowly defeats the AI loop and CI/compute thesis | persistent query skeleton lowered through interpreter, Cranelift, and QBE with cold/warm/invalidation evidence |
| 4 | AI/human source and diagnostic interface | familiar-looking syntax can still cause more repair rounds or hide authority/cost | three source projections over identical semantics, malformed edits, hidden repairs, and timed audits |
| 5 | trust and information-flow precision | a Boolean taint or universal sanitizer creates false confidence, while full IFC can overwhelm source and compile latency | typed safe-sink corpus plus 18 local/global/control-flow probes and an instrumented native build |
| 6 | unsafe/FFI/security boundary | C, allocators, raw memory, parsers, networking, and providers can invalidate safe-core claims | audited C slice plus hostile parser/TLS-provider harness and unsafe-obligation ledger |
| 7 | application scheduler and managed-region isolation | service responsiveness, cancellation, blocking foreign work, and cyclic graphs remain unproven | safe-point scheduler plus one isolated graph region under burst, pressure, and p99.9 measurement |
| 8 | evolution, bootstrap, and production adoption | good semantics can still fail on upgrades, packaging, diagnostics, and ecosystem cost | reproducible bootstrap, fresh-machine install, version migration, rollback, and first external pilot |

TLS/HTTP is not a new core-semantics gap. It is a high-risk proof of the
existing resource/effect/provider/budget design. The first-party boundary is
now specified in [Networking, HTTP, and TLS security
boundary](networking-http-and-tls-security.md); its implementation follows the
native kernel rather than blocking that kernel.

### Biggest risk-reduction step

The design portion of the **semantic-kernel convergence wave** is now specified
in the [kernel contract](semantic-kernel-contract.md). The remaining work is
executable:

1. Keep the validated interface-only value-origin/access/ability summary as the
   input contract for the next experiments; do not freeze its sample syntax.
2. Make the compact validator and native harness consumers of the same real
   typed-core artifact; keep full replay snapshots as mismatch-only evidence and
   preserve both coordinated-frontend-lie and false-no-alias negative controls.
3. Lower the portable core through one minimal native path and execute the
   blocking native/FFI cases, retaining QBE as the complexity control and
   Cranelift as the leading candidate.
4. Complete the concurrency/graph and evolution suites while the host evidence
   runner compares values, failures, initialization, provenance, cleanup
   traces, memory/release distributions, diagnostics, and invalidation work.
5. Select or revise a semantic rule only from hard-veto evidence, a minimized
   counterexample, measured ergonomics failure, or backend-independent
   impossibility.

This wave settles more uncertainty per implementation variable than starting a
web framework, actor VM, self-hosted compiler, or syntax polish pass. It also
creates the protocol the first Lang-written corpus runner can later consume.

## What has genuinely settled

| Decision | Why it is stable |
|---|---|
| AI authors, expert humans audit | This is the optimization context for source, diagnostics, evidence, and operations. |
| Canonical formatter-owned source | The author explicitly accepts it; it shrinks irrelevant variation and diff noise. |
| ADTs, pattern matching, immutable defaults, no general null | These jointly expose states and failures to compiler, agent, and reviewer. |
| Named multi-role calls with canonical exact-name punning | Roles stay explicit while `reservation: reservation` becomes `reservation:` without ambiguity. |
| Typed effects/capabilities plus static providers | One mechanism covers authority, side effects, DI, testing, context, and instrumentation. |
| `Result` for expected failure; panic for defects at supervised boundaries | This prevents catch-all ambiguity while retaining fault isolation. |
| Strict sequencing and non-resumable v1 capabilities | Effects, moves, and cleanup have one observable order; DI does not capture or duplicate owned continuations. |
| Restricted automatic release plus explicit fallible finish | Lexical safety cannot erase a primary result or execute arbitrary destructor code. |
| Ownership and resource programming in v1 | Explicit author decision; the first IR and backend must prove it. |
| No mandatory global tracing heap | The author delegated the mechanism choice but explicitly prioritized smooth predictable execution; ownership/arenas are universal and managed graphs are opt-in islands. |
| Public borrowed-result summaries | Body-blind checking now works with verified parameter/field origin paths, tagged alternatives, access mode, independent abilities, and targeted fresh callback origins. |
| Structured tasks, supervised actors, bounded streams | These cover finite work, long-lived state, cancellation, and flow control without one concurrency primitive pretending to do everything. |
| Private-by-default modules and verified component DAG | Dependency and architecture meaning is compiler-visible; modular-monolith policy is a preset, not hard-coded DDD grammar. |
| Text plus semantic protocol | Humans review calm source; agents query lossless types, graphs, effects, obligations, costs, and live state. |
| Fast edit lane plus stronger verify/release lanes | Static soundness stays immediate while expensive evidence remains explicit and enforceable. |
| Small core plus official kits | Pipelines, caches, HTTP, SQL, AI, resilience, UI, and delivery gain turnkey support without freezing domain fashion into grammar. |

## Blind-spot closure matrix

| Concern | Placement | Important refusal |
|---|---|---|
| Monitoring, profiling, load, SLOs, delivery | Typed event/evidence substrate plus performance kit | Do not collapse distinct instruments into “automatic observability.” |
| Data cleaning, warehouses, streams, assets | Ordinary typed graph plus data/asset kit and runner providers | Do not put a universal pipeline DSL in core or promise exactly-once effects by name. |
| Caches | Typed provider/policy with clock, capacity, freshness, validation, and events | Do not make memoization invisible or infer invalidation from hope. |
| Files, sockets, locks, foreign handles | Owned lexical resources, deterministic cleanup, structured cancellation | Do not rely on finalizers or allow an ordinary guard across `await`. |
| DDD/onion/hexagonal layout | Component graph plus scale-sensitive project policy and conventional projection | Do not equate folder names with architecture or force aggregates into grammar. |
| REPL/live reload | Typed transactional live workspace over the compiler/runtime service | Do not invent a second dynamic language or claim external effects can be undone. |
| Imports/exports | Explicit private-by-default modules, no wildcard/side-effect imports, DAG dependencies | Do not permit initialization order to become hidden control flow. |
| CI/CD and provenance | One hermetic action/evidence graph from local check through canary | Do not let CI become a separate source of truth. |
| AI inference and workflows | Models/tools as effects; typed authority, budgets, traces, replay references, stochastic evals | Do not make one provider API or agent-graph notation a language primitive. |
| Untrusted input and information flow | Nominal validators and sink-specific types plus typed-IR flow summaries; targeted global/control checks in stronger lanes | Do not equate decoded, encoded, internal, authenticated, or model-produced data with universal trust. |
| Low-level programming | Native v1 subset, stable ABI/layout rules, unsafe contracts, allocators, ownership | Do not defer the hard semantic constraints while claiming future systems support. |
| TLS and HTTP | Independently versioned official kit over typed effects/resources and an audited provider | Do not write crypto, expose permissive verification, or let parser compatibility become ambiguity. |
| Ecosystem lifecycle | Versioned evidence, migrations, content identity, verified provenance, retraction, impact testing, staged bootstrap, and support ownership | Do not prebuild a registry or freeze compiler internals/ABI before demand. |

## Stakeholder objections and pre-emptive answers

| Lens | Likely concern | Design response | Evidence still required |
|---|---|---|---|
| Compiler engineer | Effects plus ownership plus refinements will destroy latency. | Separate bounded local analyses, stable fingerprints, persistent queries, progressive assurance, and a fast codegen tier. | Warm/cold edit benchmarks at scale and invalidation traces. |
| Systems programmer | Inferred ownership will become mysterious or unsound at FFI/async boundaries. | Infer only locally; render ownership; require explicit public, stored, concurrent, unsafe, and foreign contracts. | Native corpus with callbacks, arenas, aliasing, unwind, atomics, OOM, and async resources. |
| Application architect | Compiler-enforced architecture becomes ceremonial dogma. | General component DAG underneath; modular-monolith/DDD layout is a scale-sensitive policy with explicit escapes. | Small and large application trials plus escape-frequency audit. |
| SRE | Built-in telemetry creates cost, cardinality, and PII disasters. | Typed classification, budgets, compile-out levels, tail/focused sampling, backpressure, and self-observation. | Load test and incident game under telemetry degradation. |
| Data engineer | Batch/stream abstraction hides event-time and delivery semantics. | Make boundedness, event time, watermarks, state, replay, and sink guarantees explicit. | Late-data, backfill, transactional/idempotent sink, and recovery trials. |
| Security reviewer | Powerful agents, build plugins, FFI, and live inspection expand authority. | Capabilities, taint/classification, sandboxed builds, unsafe ledgers, expiring debug leases, provenance. | Threat-seeded tests and independent review of escape hatches. |
| AI evaluator | Concise syntax may only exploit model familiarity and tests may overfit. | Multi-model hidden tasks, mutation tests, semantic protocol comparison, repeated stochastic evals. | Baseline study across familiar languages and candidate projections. |
| Human maintainer | Machine-oriented metadata will make source oppressive. | One calm canonical human surface; details appear through semantic lenses on demand. | Timed audit studies on non-toy diffs and incidents. |
| Operator/adopter | A new language is too expensive to install, upgrade, debug, and staff. | Single tool, static artifacts, C/Wasm/BEAM bridges, semantic migrations, official kits, compatibility train. | Fresh-machine onboarding, upgrade/rollback, and production pilot. |

## Remaining unknowns: core versus experimental

### Must resolve during the first executable slice

1. The minimal ownership model: move/copy rules, local borrow inference,
   mutation, partial moves, structural release order and cost, cleanup failure,
   and async suspension.
2. The defined native semantics: integer overflow, initialization, pointer
   provenance, aliasing, panic/unwind, OOM, ABI/layout, and checked release
   behavior.
3. The smallest effect/provider model that composes with resource lifetimes and
   remains easy to explain.
4. The initial backend and development evaluator, chosen by measured compile
   latency and implementation burden—not peak benchmark reputation.
5. Stable syntax-tree/symbol identities and machine-readable diagnostics.
6. Numeric budgets for parse/check/dev-build/run on a declared machine and
   corpus.

### Can remain experiments after implementation begins

- braces versus indentation versus uniform-tree projection;
- visual effect-call marker;
- exact rendering of inferred public effects and borrows;
- dedicated state-machine or pipeline views;
- authoritative text versus semantic code store;
- UUIDv7 application-template default;
- first native GUI host;
- BEAM and Wasm backend order;
- native local-inference scheduling;
- precise RC/reuse details and the algorithm inside an optional isolated
  managed region; the absence of a mandatory global heap is no longer open.

## Leading candidate for the first vertical proof

A **Lang-written corpus and evidence runner** is now the highest-compounding
native-proof candidate. It should include a single-binary streaming JSON/data
workload rather than discard that useful stress case. If selected after the
ownership, syntax, and backend experiments, it should:

- stream a deliberately dirty, large input under a fixed memory budget;
- decode bytes and Unicode with adversarial limits;
- use owned buffers, borrowed views, an arena, lexical files, and fallible
  allocation;
- cross one audited C ABI boundary;
- express validation, quarantine, lineage, metrics, cancellation, and cleanup;
- run under the reference interpreter and native backend with differential
  checks;
- expose static allocation/copy/blocking explanations and runtime CPU,
  allocation, lock, I/O, and memory profiles;
- invoke and differentially compare interpreter/native corpus artifacts;
- build hermetically as one artifact with evidence and provenance;
- explain semantic/build cache keys, hits, misses, and invalidation;
- support a typed live session and machine-readable repair query.

It does not prove the service/actor thesis. It earns the native semantic
foundation while improving Lang's own conformance and feedback machinery. The
next modular-monolith HTTP/SQL/AI workflow can then test effects, DI,
architecture, concurrency, operations, and delivery without reopening memory
fundamentals.

The complete dependency-ordered inventory, including reversible later-horizon
questions, is maintained in the [residual uncertainty
register](residual-uncertainty-register.md).

## Exit from exploration

The design is ready for the bounded experiments in the
[convergence work program](convergence-work-program.md). The first vertical is
selected only after the ownership laboratory, syntax corpus, and small backend
contest establish that its stress points correspond to the intended language.
Stage 0 builds the semantic corpus, schemas, budgets, and competing renderings;
Stage 1 then implements the native-capable kernel. A design question reopens the
core only if two materially different prototypes expose the same contradiction.
