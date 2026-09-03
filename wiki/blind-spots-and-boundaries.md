---
id: blind-spots-boundaries
title: Blind spots and system boundaries
summary: A risk atlas for concerns that can invalidate an otherwise elegant AI-first language design.
type: atlas
status: active
confidence: medium
created: 2026-09-02
updated: 2026-09-03
tags: [blind-spots, risk, boundaries, security, evolution]
related: [vision, design-atlas, trust-validation-information-flow, boundary-data-validation-persistence, residual-uncertainty-register, compiler-feedback-latency, domain-data-distribution, implementation-path, research-ledger]
---

# Blind spots and system boundaries

## Purpose

The project has a strong semantic center. Its greatest remaining risks are now
outside pretty syntax: time, change, overload, trust, recovery, compiler
latency, and the mismatch between what tools can prove and what production can
do. This atlas keeps those concerns visible without promoting each one into a
language feature.

## Concentric scope

```text
language semantics
  effects, types, errors, contracts, tasks, resources, versioned data
    service runtime profile
      actors, supervision, backpressure, deterministic worlds, event substrate
        official kits
          HTTP, SQL, schemas, resilience, workflows, telemetry, deployment
            external systems
              databases, queues, load balancers, cloud, model providers
```

The inner ring contains durable meaning. Outer rings may be turnkey and
first-party without becoming grammar. This is the mechanism for “boiling the
ocean” analytically while implementing a narrow, falsifiable core.

## Blind-spot inventory

| Blind spot | Failure if ignored | Best initial home |
|---|---|---|
| Compiler latency | Correctness loop becomes too slow for agents and CI | Compiler architecture and budgets |
| Trusted computing base | Rich guarantees depend on a large, unauditable implementation | Small core checker, conformance suite |
| Temporal semantics | Wall time, monotonic time, ordering, expiry, and replay get conflated | Core time types plus `Clock` effect |
| Schema/state evolution | Rolling deploys corrupt or strand durable data | Versioned schema semantics and migration tools |
| Partial failure | Local-looking calls hide timeout, duplication, and uncertainty | Explicit remote effects and resilience kit |
| Overload and liveness | Retries, queues, mailboxes, and pools amplify failure | Bounded runtime primitives and budgets |
| Deterministic simulation | Rare distributed failures cannot be reproduced | Replaceable effects plus simulated world |
| Storage and recovery | “Durable” state lacks restore and failure evidence | Provider contracts and deployment policy |
| Information flow | PII, secrets, and tenant data reach logs or sinks | Classified values, sink policy, tests |
| Boundary representation | Parsing, defaulting, number conversion, or canonicalization silently changes meaning | Versioned codecs, exact presence, boundary probes |
| Validation freshness | A once-valid fact is used after policy or world state changes | Versioned evidence and transactional recheck |
| Transaction retry | Retried closure duplicates external effects or uses new time/randomness | Effect-restricted retry scope and outbox |
| Pool behavior | Hidden queues leak resources, starve tenants, or amplify overload | Affine leases, bounds, deadlines, causal timings |
| Authority lifecycle | Long-lived handles outlive delegation, principal, or provider policy | Scoped, attenuated, expiring, revocable capabilities |
| Query opacity | N+1 access, cardinality mistakes, or plan drift appears only in production | Typed cardinality plus plan/budget evidence |
| Content interpretation | Declared media type, bytes, extension, and sniffed content disagree | Explicit media/decoder policy and bounded inspection |
| Supply-chain/build authority | Dependencies execute code during build or ship hidden risk | Hermetic sandboxed package/build system |
| Numeric semantics | Overflow, NaN, rounding, and platform drift corrupt invariants | Explicit core numeric policy |
| Text and identity | Unicode confusables, normalization, locale, and collation cause bugs | Core text policy and safe identifiers |
| Resource termination | Programs are memory-safe but exhaust CPU, memory, handles, or money | Budgets, structured lifetimes, cancellation |
| Foreign code | FFI invalidates purity, memory, blocking, and replay claims | Audited boundary contracts |
| Hot upgrade | New code interprets old state/protocols incorrectly | Compatibility and state-migration tooling |
| Governance | Features and kits accrete without removal or evidence | Proposal gates, editions, deprecation policy |
| Autonomous development | Agents optimize proxies, exceed authority, or hide uncertainty | Evidence protocol, bounded actions, rollback |
| Accessibility of evidence | Only compiler experts can interpret failures | Human and machine projections of one model |
| Economic/ecological cost | Fast-looking workflows consume excessive CI, telemetry, or inference | End-to-end resource accounting |

## Security that can genuinely shift left

Security is not one `Safe` type. High-leverage candidates are:

- memory-safe defaults, initialized values, bounds checks, and explicit
  arithmetic overflow policy;
- no ambient filesystem, network, process, environment, clock, random, secret,
  or foreign-code authority;
- destination- and protocol-scoped network capabilities where practical;
- opaque secret values that cannot be formatted, serialized, compared, or
  exported without explicit operations;
- nominal validated values such as `SafeHtml`, `SqlParameter`, `NormalizedPath`,
  `VerifiedPrincipal`, and tenant-scoped IDs;
- classification and declassification attached to values and semantic sinks;
- provenance/integrity tracked separately from semantic validity,
  destination-specific encoding, confidentiality, and authority;
- typed query/encoding APIs instead of string construction;
- structured concurrency, bounded allocation/queues, deadlines, and cancellation
  as denial-of-service controls;
- sandboxed dependency builds with declared capabilities and provenance;
- `unsafe` and FFI only in narrow, reasoned, auditable scopes;
- constant-time cryptographic operations only through vetted implementations,
  never a compiler's optimistic inference;
- architecture policy separating untrusted parsing, domain decisions, secrets,
  and egress.

The compiler can prove conformance to a declared model. It cannot establish that
an authorization policy expresses the business's intent, that a foreign
library is bug-free, or that infrastructure matches its deployment manifest.
Generated attack cases, fuzzing, dependency review, runtime containment, and
operational response remain required.

The detailed boundary treatment is in [Boundary data, validation,
transactions, and pools](boundary-data-validation-and-persistence.md). These
concerns reuse the core rather than justify JSON, SQL, or ORM grammar.

## Time is more than `Clock.now()`

`Clock.now()` is an effect and is supplied through the provider graph. The type
system should distinguish:

- wall-clock instants used for human/calendar meaning;
- monotonic durations used for timeouts and latency;
- logical/causal versions used for distributed ordering;
- simulated time used for deterministic tests;
- event occurrence, observation, processing, and persistence times.

Time zones, calendars, leap behavior, parsing, and locale belong in an official
data kit backed by a versioned database. Storing a local datetime without zone
or interpretation should require an explicit domain reason.

## Numeric and text policy

These mundane choices create serious cross-platform and security failures:

- integer sizes are explicit; overflow is checked by default and wrapping,
  saturating, or trapping arithmetic is named;
- money and quantities use decimal/fixed-point or dimensioned domain types,
  never an accidental binary float;
- floating-point equality, NaN, nondeterministic reductions, and target-specific
  behavior are visible in reproducibility evidence;
- source identifiers use one normalized representation and flag confusables;
- bytes, Unicode scalar sequences, grapheme clusters, and display width are not
  interchangeable `String` operations;
- case folding, collation, and locale-dependent formatting are explicit.

These should be decided before a stable wire format or package namespace exists.

## Small trusted checker and evidence certificates

A feature-rich compiler is difficult to trust as one monolith. Explore a small
checker for typed/effect IR, capability containment, architecture edges,
resource lifetimes, and proof/evidence manifests. Compiler passes emit an
artifact plus enough certificate data for the checker to validate the important
claims independently.

Lean's architecture—automation produces proof terms checked by a smaller
kernel—is inspiration for the trust shape, not a proposal to turn all programs
into theorem-proving exercises.
([Lean reference](https://lean-lang.org/doc/reference/latest/))

The checker itself needs fuzzing, differential tests, reproducible builds,
formalization where cost-effective, and versioned semantics. “Small” is not a
synonym for correct.

## AI and autonomous-loop boundary

“Dark software factory” is not yet a stable technical term with one authoritative
specification. The useful design target is an autonomous change loop whose
authority and evidence are explicit:

```text
intent -> scoped plan -> transactional semantic edits -> affected evidence
       -> policy gate -> reversible artifact -> monitored rollout -> rollback
```

The protocol should provide:

- machine-readable goal, constraints, and unresolved assumptions;
- capability-scoped actions and resource budgets;
- stable semantic IDs and transactional edits;
- diagnostic deltas and counterexamples;
- an evidence manifest identifying exactly what passed, failed, or was deferred;
- approval gates based on risk, not an unconditional autonomous mode;
- signed provenance, deployment observation, and rollback handles;
- resistance to reward hacking: hidden tests, mutation tests, independent
  checkers, and production invariants.

The language improves this loop by making program truth queryable. It should not
embed a particular planning framework, milestone vocabulary, model vendor, or
agent harness into source semantics. GSD-like research/plan/implement systems
consume the protocol and evidence graph as external clients.

## Lifecycle test matrix

Evaluate more than greenfield compilation:

| Life stage | Required evidence |
|---|---|
| Minute zero | One signed tool installs or runs portably; no toolchain assembly |
| Day one | Format, check, test, run, inspect, package, and containerize work from one project contract |
| Team growth | Modules, ownership, architecture policy, dependency review, and CI scale predictably |
| Production | Health, traces, profiles, live queries, overload behavior, and redaction are bounded and useful |
| Incident | Failure state persists; causal evidence and safe introspection identify uncertainty |
| Upgrade | Semantic compatibility, migrations, mixed versions, rollback, and deprecations are checked |
| Legacy adoption | C/Wasm/BEAM boundaries and generated adapters permit incremental strangling |
| Disaster | Restore is tested against declared RPO/RTO and compatible code/schema versions |
| Retirement | Data export/deletion, dependency removal, and archived evidence are supported |

## Representative stress-test programs

The corpus should expand in rings rather than attempt four production systems at
once:

1. **Checkout service:** effects, domain types, HTTP/SQL, auth, telemetry,
   retries, compensation, upgrades, and fault injection.
2. **Durable agent workflow:** model capability, nondeterminism, budgets,
   checkpointing, tool authority, replay, and human approval.
3. **Network service or small broker:** bounded streams, backpressure, pooling,
   load shedding, protocol parsing, and deterministic simulation.
4. **Embedded controller:** restricted allocation/effects, C interop, timing,
   hardware simulation, and predictable artifacts.
5. **Rendering or emulator core:** data layout, numeric determinism, tight loops,
   resources, native interop, profiling, and fuzzing.
6. **Storage engine slice:** checksums, crash consistency, recovery, schemas,
   and file effects—only after the language can test these claims honestly.

Each starts as a vertical slice, graduates to a maintained open-source exemplar,
and gathers real adopter evidence. Feature proposals must improve measured
outcomes on this corpus rather than merely make a showcase snippet attractive.

## Focus guardrail

The first implementation still has one thesis: **explicit effects plus fast,
structured evidence improve AI-authored, human-audited production software**.
Domain kits, databases, renderers, embedded targets, and autonomous factories
are pressure tests and later rings. They must not delay the frontend,
effect/provider prototype, agent protocol, or representative service proof.
