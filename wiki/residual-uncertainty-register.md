---
id: residual-uncertainty-register
title: Residual uncertainty and risk-reduction register
summary: A dependency-ordered map of what remains unknown, why it matters, which evidence can resolve it, and when the decision must be made.
type: strategy
status: active
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [uncertainty, convergence, risks, experiments, roadmap]
related: [design-baseline, semantic-kernel-contract, semantic-kernel-probes, ownership-lifetime-decisive-study, trust-validation-information-flow, boundary-data-validation-persistence, language-ecosystem-lifecycle-backcast, convergence-audit, convergence-work-program, implementation-path, capability-gauntlet, compiler-feedback-latency, memory-reclamation-policy, native-low-level-profile, open-questions, research-ledger]
---

# Residual uncertainty and risk-reduction register

## Reader and outcome

This is for the project author, language designers, and implementation agents
choosing what to investigate next. After reading it, they should be able to
distinguish an unresolved value choice from a design hypothesis or measurement,
avoid reopening settled foundations without evidence, and select the next work
that retires the most upstream risk.

## Executive assessment

The project has explored enough breadth to begin implementation. The remaining
unknowns are no longer an unbounded universe of missed language features. They
fall into three classes:

1. **Semantic falsification:** does the candidate meaning remain coherent when
   executable, concurrent, resource-constrained, optimized, and foreign?
2. **Measured selection:** which ownership, backend, syntax, runtime, and
   analysis implementation best meets the stated budgets?
3. **Product policy:** where should defaults and gates sit for actual adopters?

The first class can invalidate the IR and therefore goes first. The second can
change implementation architecture and follows immediately. The third is
important but remains reversible behind stable interfaces until real users and
workloads exist.

The largest present risk is not a missing concept. Five bounded workbenches now
cover local ownership, CFG last use, public summaries, independent typed-core
validation, and real native/FFI hazards. The risk is continuing to extend those
disposable models instead of connecting their surviving contracts through one
real lossless frontend, portable typed core, interpreter, and native lowering.

## Final extrapolation sweep

Three additional passes—first over boundary data and persistence, then over the
author's likely next objections, and finally over the simulated lifetime of a
successful language ecosystem—did not reveal another semantic-center feature.
They produced the following placement results:

| Concern surfaced | Preserve now | Settle later |
|---|---|---|
| JSON presence, precision, duplicate fields, canonicalization | exact ADTs, limits, schema/version identities, flow summaries | codec representation and API ergonomics |
| Zod/Pydantic-like validation and Ecto-like changes | nominal constructors, contracts, evidence identity | error accumulation, derivation, form/API projections |
| transaction plans and retries | effect restriction, affine transaction, cancellation/result taxonomy | SQL isolation mapping, outbox and retry policy |
| connection pooling | affine lease, bounds, deadlines, causal events | scheduling/fairness/topology implementation |
| capability expiry/revocation | scoped capability lifetime and typed operation failure | distributed revocation provider and policy |
| decimal/money, calendars, MIME/content handling | nominal types and explicit codecs already suffice | independently versioned data/network kits |
| OS signals, process startup/fork, dynamic loading | unsafe/profile/FFI hooks already exist | per-platform native kit and conformance corpus |
| N+1/query-plan/index drift | query identity and cost/evidence hooks | database-specific analysis and release policy |
| package compromise/retraction | content identity, declared build effects, provenance fields | registry/update service and key operations |
| source and ecosystem evolution | versioned semantics/evidence, migration identities, narrow stable boundaries | edition cadence and compatibility promises |
| bootstrap and maintainer succession | interpreter oracle, staged-build fixtures, owned support surfaces | self-hosting schedule and governance bodies |
| model/provider churn | open semantic/agent protocols and compiler-owned evidence | provider adapters, training, and retrieval strategy |

The resulting [boundary contract](boundary-data-validation-and-persistence.md)
adds adversarial kit probes, while the
[lifecycle backcast](language-ecosystem-lifecycle-backcast.md) adds compatibility,
supply-chain, adoption, and end-of-life probes. Neither reorders the
semantic-kernel wave. A further unconstrained survey is now more likely to
rename toolkit backlog than change the core.

## Ranking method

Each frontier is ordered using five questions:

| Factor | High-risk signal |
|---|---|
| Semantic blast radius | A different answer changes most programs or every backend |
| Irreversibility | A later change breaks source, ABI, storage, wire, or ecosystem contracts |
| Dependency unlock | Evidence resolves several downstream choices |
| Uncertainty | Competing credible designs remain and no local evidence exists |
| Evidence efficiency | A small probe can disprove a large assumption |

Performance, security, AI accuracy, and developer ergonomics are evaluated
inside every frontier rather than deferred into separate cleanup phases.

## Decision horizons

```text
NOW: preserve in core IR or prove semantic law
  -> BEFORE SURFACE: measure ergonomics and diagnostic representation
    -> BEFORE NATIVE G3: select backend, memory, FFI, and dogfood workload
      -> BEFORE SERVICE G3: select scheduler, managed graphs, I/O, telemetry
        -> BEFORE 1.0: freeze evolution, packaging, compatibility, governance
```

A later-horizon concern can still require an early hook. Trust flow, stable
semantic identity, effect summaries, versioned schemas, and runtime event IDs
are examples: their final policy can wait, but omitting them from IR creates an
expensive retrofit.

## Frontier 0 — executable semantic kernel

**Priority:** immediate; highest blast radius.

| Unknown | Leading posture | What can overturn it | Decisive evidence |
|---|---|---|---|
| ownership transfer and borrow inference | value/access surface, public modes/origins, local last-use inference, visible semantic cliffs | annotations become distant, markers dominate, or copies hide | 64 ownership workloads with repair/audit/compile/runtime metrics |
| partial moves | reject in v1 | whole-value workarounds are pervasive | parser, resource aggregate, state-machine corpus |
| borrow across suspension | reject ordinary local borrow | async zero-copy requires excessive copying | socket/parser/task-owned-storage experiment |
| cleanup and fallible close | restricted `release`, explicit `finish` | real OS/foreign APIs cannot adapt cleanly | file, socket, TLS, lock, transaction, callback probes |
| cancellation and concurrent failures | sticky control with stable aggregate | cleanup latency or error volume is unacceptable | bounded fan-out and cascading-failure schedules |
| panic, OOM, and stack failure | distinct configured isolation outcomes | reusable code fragments across profiles | interpreter/native/profile matrix |
| arithmetic and floats | checked ordinary integers; explicit float policy | target cost or reproducibility is unacceptable | cross-target numeric and optimizer corpus |
| effects/providers | static non-resumable capabilities | generators/control abstractions demand handlers | provider corpus followed by one-shot-handler comparison |
| semantic event oracle | causal meaning, not operational timing | it blocks valid optimization or misses defects | generated differential optimization cases |
| text and bytes | UTF-8 `Text`, explicit units and normalization | editor/OS/FFI workloads require other views | parser/editor/filesystem/grapheme corpus |
| atomics | explicit five-order low-level model | diagnostics remain too subtle | litmus tests plus protocol-specific API comparison |

**Exit:** the 24 blocking probes agree in a reference interpreter and one native
lowering; every discrepancy is reduced to a minimal program.

## Frontier 0b — trust and information-flow hook

**Priority:** preserve now; policy strength measured later.

| Unknown | Leading posture | What can overturn it | Decisive evidence |
|---|---|---|---|
| surface representation | `Input<Origin, T>` or equivalent IR projection | wrapper noise harms ordinary parsing | boundary corpus rendered with explicit and inferred forms |
| safe-sink coverage | typed structured APIs first | too many important sinks remain stringly | SQL/HTML/process/path/network/model fixtures |
| propagation precision | local, field-sensitive declared summaries | false positives or laundering dominate | partial aggregates, collections, closures, callbacks |
| implicit/control flows | targeted at authority and executable sinks | material leaks occur outside selected policies | high-integrity decision and confidentiality probes |
| confidentiality form | secret/nominal wrappers plus classification metadata | composition becomes a parallel type system | telemetry/storage/service corpus |
| validation freshness | evidence names policy/world version or is transaction-local | wrapper/version noise obscures ordinary domain code | mutable-policy, entitlement, uniqueness, and stale-command cases |
| global analysis lane | policy-scoped `verify`/`release` | edit feedback or security coverage is poor | incremental source-to-sink benchmarks |
| runtime shadow labels | adversarial builds only | native/dynamic behavior cannot otherwise be tested | FFI/plugin/reflection differential runs |
| dependency summaries | derived where possible; foreign claims are untrusted | ecosystem friction prevents coverage | safe-adapter and package-summary trial |

**Exit:** the 18 `FLOW-*` probes have stable diagnostics and no general
`untaint()` operation; the IR can preserve summaries without paying ordinary
runtime metadata cost.

## Frontier 1 — compiler architecture and feedback economics

**Priority:** begin with Frontier 0, because a slow checker invalidates the AI
loop.

| Unknown | Leading posture | Evidence |
|---|---|---|
| host language | Go 1.24 selected for reversible Stage 0 bootstrap; Rust/OCaml remain later comparators | measured parser/query/type/lowering tracer slice can reopen before stable compiler commitments |
| query granularity | persistent incremental semantic service | edit-shape invalidation and peak-memory traces |
| stable identities | content/structure-aware definition and expression IDs | move/rename/format/refactor survival corpus |
| public fingerprints | types/effects/ownership/contracts separated from bodies | downstream invalidation after private/public edits |
| diagnostic protocol | structured cause graph plus concise human projection | hidden multi-model repair and timed human audit |
| independent checker | compact typed-core recomputation at trust crossings; full traces on mismatch | schema-growth gate as real effects/native facts arrive; Spike 004 resolved the bounded slice |
| specialization | restrained monomorphization with explainable sharing | generic corpus build time, size, and runtime |
| cache persistence | content-addressed immutable artifacts | cold/warm/worktree correctness and transfer economics |
| solver/proof budget | explicit higher assurance lanes | refinements under timeout, cache, and failure diagnostics |

**Exit:** budgets are distributions measured on declared machines and corpora,
not aspirational single numbers; an affected edit returns sound useful feedback
without recompiling the world.

## Frontier 2 — human and AI source interface

**Priority:** after enough semantics execute to avoid testing decorative syntax.

| Unknown | Candidate set | Required comparison |
|---|---|---|
| block structure | braces, indentation, uniform-tree projection | malformed recovery, generation, diff, scan, audit |
| call labels | named multi-role, concise unary, exact-name punning | reorder/refactor safety and token/visual cost |
| ownership visibility | declaration-only, all call sites, ambiguity-only move marker | owner/copy audit and model repair |
| effect visibility | signature only, call marker, semantic overlay | authority audit versus noise/churn |
| inferred fact visibility | source annotations versus editor/agent lenses | cold review without opening tooling |
| pipeline notation | pipe, method chain, explicit nested call | data-flow reading and precedence errors |
| source authority | canonical text versus semantic store with text projection | Git interoperability, merge, tooling, recovery |
| specs | inline/adjacent projection while preserving flat AAA tests | maintenance, mutation score, failure diagnosis |
| identifiers | ASCII rules, confusable policy, casing convention | scan, model generation, international use |

**Exit:** one canonical human source and one machine projection round-trip the
same semantic model and outperform credible alternatives on whole-loop tasks.

## Frontier 3 — native representation, backend, and foreign boundary

**Priority:** before selecting the first G3 workload.

| Unknown | Leading posture | Evidence |
|---|---|---|
| development backend | Cranelift leading, QBE complexity control | same kernel compile/runtime/debug/size measurements |
| release backend | defer LLVM until a measured gap | optimized numeric/FFI/throughput corpus |
| shared immutable values | precise RC/reuse is a candidate | trees, large messages, concurrency, release tails |
| cycles | explicit owner-confined managed region | compiler/GUI/actor graph comparison with indices/weak refs |
| arenas and allocators | explicit governed providers | parser/data/game workloads and OOM behavior |
| layout/ABI | explicit stable boundary contracts | C structs, callbacks, varargs, bitfields, alignment |
| pointer provenance/aliasing | safe core stricter than backend IR | interpreter, sanitizer, optimizer adversaries |
| panic/unwind | never cross ordinary foreign ABI | C/Rust/C++ shim and process-boundary matrix |
| dynamic libraries | handle owns symbol/callback lifetimes | unload/reload/version-skew probes |
| SIMD/GPU | explicit target/cost contracts later | rendering/inference kernels without semantic drift |

**Exit:** a nontrivial native program crosses an audited C boundary and agrees
with the interpreter under optimization, sanitizers, OOM, cancellation, and
resource release.

## Frontier 4 — application and service runtime

**Priority:** after the native-core contract survives.

| Unknown | Leading posture | Evidence |
|---|---|---|
| scheduling | cooperative compiler safe points first | CPU loop, FFI, allocation, burst, p99.9 fairness |
| blocking classification | isolate declared CPU/I/O foreign work | scheduler watchdog and saturation tests |
| tasks versus actors | finite structured tasks; actors own long-lived state | service, stream, GUI, workflow examples |
| mailboxes and streams | bounded by default with explicit overflow policy | overload, backpressure, cancellation |
| supervision | explicit restart intensity and state recovery | repeated defect and dependency failure |
| deterministic worlds | effects replace time/random/network/storage/process | seeded failures and recorded replay honesty |
| managed graphs | one owner-confined region service | pause/headroom/pinning/cross-owner tests |
| telemetry substrate | structural causal events with budgets | disabled/degraded/overload/privacy load tests |
| live inspection | scoped authenticated debug capabilities | production incident games and PII controls |
| hot reload/upgrade | state/schema/protocol-aware transaction | mixed-version, rollback, downgrade, in-flight work |

**Exit:** a fault-injected production-style service stays bounded, explains its
failures through structured evidence, and can be upgraded or rolled back.

## Frontier 5 — architecture, data, operations, and official kits

**Priority:** interfaces early; concrete breadth after core/runtime evidence.

| Unknown | Leading posture | Evidence |
|---|---|---|
| component policy | general DAG with modular-monolith preset | small/large app refactor and waiver counts |
| execution/security/business context | implicit only when structurally safe | auth/tenant propagation threat cases |
| HTTP/TLS | audited provider behind strict semantic API | hostile framing, identity, replay, proxy-chain tests |
| SQL/change model | Ecto-like explicit changes and repository ports | migrations, constraints, query plans, transactions |
| boundary codec contract | exact JSON plus derived versioned schemas | duplicate/presence/unknown/precision/canonicalization probes |
| transaction/pool contract | effect-restricted retry and affine bounded leases | conflict, cancellation, saturation, fairness, and commit-unknown probes |
| caching | explicit typed key/freshness/authority policy | stale fill, stampede, tenant, version, poisoning |
| workflows/sagas | durable event history plus explicit compensation | crash/resume/version migration/duplicate delivery |
| data pipelines | typed graph with explicit time/state/lineage | late data, backfill, warehouse pushdown |
| asset pipelines | hermetic content graph | reproducibility, invalidation, platform tools |
| contract testing | versioned provider/consumer evidence | compatible and breaking evolution cases |
| UI hosts | shared pure model where honest, native/web adapters | accessibility, input/IME, platform escapes |
| AI kit | `Model` and tool capabilities, budgets, replay, evals | provider changes and adversarial agent tasks |

**Exit:** the toolkit removes repeated unsafe plumbing without introducing a
second hidden language, a mandatory architecture style, or deep dependency
trees.

## Frontier 6 — packaging, evolution, and adoption

**Priority:** design hooks early; freeze only after external use.

| Unknown | Leading posture | Evidence |
|---|---|---|
| package resolution | direct dependencies, lock graph, hermetic builds | conflict, feature, vendoring, offline cases |
| build authority | sandboxed declared inputs/outputs, no network | malicious build/plugin fixtures |
| version axes | source edition, spec, compiler, IR, ABI, protocol, kits separate | compatibility train and rollback rehearsal |
| distribution | one signed tool and reproducible artifacts | clean macOS/Linux/Windows onboarding |
| migrations | previewable semantic edits plus evidence | multi-edition package and state/schema upgrades |
| governance | evidence gate and removal path per feature | real proposals, deprecations, security exception |
| standard-library boundary | small stable core plus independently versioned kits | update cadence, security patch, ecosystem feedback |
| interop adoption | C first, Wasm components, BEAM later | strangler-fig migration and mixed-language debugging |
| external users | maintained exemplar and pilot | onboarding time, escapes, incidents, upgrades |

**Exit:** an external team can install, understand, deploy, observe, upgrade,
rollback, and remove Lang components without relying on project authors.

## Frontier 7 — AI-development loop

**Priority:** protocol skeleton with compiler; production automation later.

| Unknown | Leading posture | Evidence |
|---|---|---|
| context bundles | minimal semantic slice plus on-demand queries | full-source versus retrieval repair trials |
| transactional edits | previewable typed operations with rollback | concurrent worktrees and stale-state tests |
| feedback reward | diagnostic/evidence delta, not compiler-green alone | hidden defects and wrong-fix rate |
| stochastic evals | versioned datasets, repeated samples, calibrated graders | model/provider drift and expert comparison |
| tool safety | low-integrity proposals plus capability authorization | NIST/AgentDojo-style hijacking tasks |
| handoffs | explicit goal/context/authority/budget/evidence envelope | delegation amplification and cancellation |
| autonomous release | risk gates, provenance, canary, rollback | seeded security/operations incidents |
| model neutrality | stable protocol independent of provider | several models and local/remote inference |

**Exit:** AI assistance improves verified-change time and defect rate without
increasing authority escapes, hidden deferred evidence, or total compute waste.

## Cross-cutting failure modes

These can invalidate several frontiers at once:

- **Semantic cross-product explosion:** each feature needs exceptions for
  effects, ownership, async, cleanup, FFI, and profiles.
- **Analysis creep:** a local rule becomes global inference and destroys edit
  latency or diagnostic locality.
- **Runtime leakage:** libraries accidentally require the scheduler, collector,
  telemetry, TLS, or dynamic loader in restricted profiles.
- **False security:** validation, taint, types, sandboxing, or proofs are treated
  as evidence of unstated business intent.
- **Evidence theater:** generated tests prove their own implementation or stale
  cached evidence is rendered as current.
- **Authority laundering:** text, metadata, structural conformance, dependency
  code, or a model response creates permission.
- **Policy fossilization:** volatile HTTP, TLS, Unicode, AI, or deployment
  practice is frozen into language grammar.
- **Escape normalization:** `unsafe`, waiver, raw string sink, or arbitrary
  dependency becomes the ordinary path.
- **Cost displacement:** a “fast” runtime pushes work into compilation, CI,
  memory, cache transfer, telemetry, or human review.
- **Dogfood distortion:** the compiler toolchain becomes the only workload and
  hides ordinary application problems.

## Convergence wave sequence

The shortest risk-reduction path is now:

1. Create the real versioned lossless frontend and typed-core spine, importing
   only the stable identities, origins, abilities, ownership events, evidence
   binding, and native obligations established by Spikes 001–005.
2. Implement the semantic kernel and 24 blocking probes in the reference
   interpreter.
3. Add the 18 trust-flow probes using nominal safe sinks and local summaries.
4. Add the JSON/change/transaction/pool fixtures as kit clients of that core;
   do not wait for a database-aware compiler.
5. Make the existing native harness consume the same core through the minimal
   native path and compare outcomes, events, release, and flow diagnostics.
6. Run ownership/reclamation and compiler-invalidation measurements.
7. Render three source interfaces and run AI repair plus human audit trials.
8. Select the backend and Lang corpus/evidence-runner vertical from the
   measurements.
9. Add service runtime capabilities only after the native core survives.

This sequence intentionally makes trust-flow support cheap to preserve but
does not make full information-flow control a prerequisite for the first
interpreter.

## Author choices versus delegated experiments

No new author choice blocks the next wave. The following value decisions will
matter before a public v0, but reasonable reversible defaults can carry the
experiments:

- text/Git remains authoritative until a semantic-store prototype proves
  materially better;
- `check` rejects sound local violations and reports expensive assurance as
  explicit obligations; `verify`/`release` enforce configured policy;
- only execution context propagates structurally by default; tenant/principal
  become explicit application inputs after ingress;
- named multi-role arguments allow concise unary operations and exact-name
  punning until corpus evidence disagrees;
- UUIDv7 remains a service-kit provider choice, not a core default;
- GUI follows the first native/service proof unless a real adopter changes the
  priority.

These are inferred defaults, not immutable decisions.

## Stop rule

Do not run another broad feature survey merely because an unknown remains. A
frontier is ready to build when:

- credible alternatives and their strongest failure modes are recorded;
- the choice cannot invalidate an earlier unimplemented hook;
- one bounded experiment can now produce more information than more prose;
- the decision has a migration seam or an explicit compatibility cost;
- success and failure evidence is machine-readable;
- another research pass is unlikely to change the candidate set.

By this rule, the next unknown-reduction work is implementation, measurement,
and counterexample reduction—not more unconstrained enumeration.
