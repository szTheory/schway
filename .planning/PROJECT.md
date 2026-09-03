# Codename Lang

## What This Is

Codename Lang is a general-purpose programming language and toolchain for
AI-authored, human-audited production software. It aims to combine fast,
structured feedback with a small coherent semantic core, usable low-level
ownership and resource control, deterministic evidence, and a calm canonical
source form that remains readable when humans mostly review rather than type it.

The first milestone is deliberately narrower than the full vision: prove one
real source-to-native semantic spine before expanding into effects, services,
agents, databases, GUIs, or an ecosystem.

## Core Value

Give an AI agent and a human reviewer the shortest reliable path from intent to
sound, reproducible evidence without wasting iteration time or hiding runtime
costs.

## Requirements

### Validated

(None yet — five executable spikes validate design hypotheses, but no production
language slice has shipped.)

### Active

- [ ] A clean checkout can format, check, interpret, and natively run a small
      canonical Lang program through one production command.
- [ ] One portable typed core gives ownership, borrowing, cleanup, and foreign
      boundaries the same meaning in the checker, interpreter, and native path.
- [ ] Interpreter, native `-O0`, and native `-O3` executions produce equivalent
      semantic outcomes and events for the milestone corpus.
- [ ] Invalid programs fail early with stable, bounded, machine-readable cause
      graphs and calm human projections.
- [ ] Compact content-bound evidence manifests make successful work reusable
      without treating frontend claims as proofs of source correctness.
- [ ] The default edit loop remains fast, deterministic, offline-capable, and
      measurable; expensive evidence runs only when its distinct question is
      relevant.

### Out of Scope

- General effect rows, resumable handlers, and runtime dependency injection —
  preserve typed-core seams, then design from a working ownership spine.
- Async/await, actors, schedulers, backpressure, supervision, or managed cyclic
  heaps — defer until sequential resource semantics survive native lowering.
- Package registry, remote cache, LSP, MCP server, web/SQL/TLS/GUI kits, AI
  inference, and distributed workflows — later milestones built on the stable
  compiler service contract.
- Full generics, subtyping, variance, partial moves, borrow across suspension,
  reflection, macros, and user-defined precedence — they would broaden the
  search space before the kernel oracle is trustworthy.
- A permanent compiler host or final optimizing backend decision — Go 1.24 and
  readable C17/Clang are reversible Stage 0 choices.

## Context

The design corpus in `wiki/` records the author's intent, primary-source
research, synthesis, decisions, examples, and open questions with explicit
provenance. Five bounded workbenches in `.planning/spikes/` have already tested
local affine ownership, CFG edge-specific loan liveness, separate-compilation
origins and generic abilities, independent typed-core certificate validation,
and native C ABI/provenance/cleanup hazards.

Those experiments converged on an immutable portable core, deterministic
interpreter oracle, compact trust-crossing validation, and conservative native
contracts. Their Go and C APIs are disposable evidence, not production language
architecture. The immediate risk is continuing to elaborate shadow models
instead of connecting the surviving contracts through a real frontend and
backend.

Canonical planning inputs:

- `wiki/real-frontend-core-ir-entry-plan.md`
- `wiki/semantic-kernel-contract.md`
- `wiki/semantic-kernel-probes.md`
- `wiki/ownership-evidence-roadmap.md`
- `wiki/compiler-and-feedback-latency.md`
- `wiki/compute-efficiency-constitution.md`
- `wiki/residual-uncertainty-register.md`
- `.planning/spikes/MANIFEST.md`

## Constraints

- **Feedback latency**: Optimize for the whole generate → verify → run → observe
  → repair loop; report cold and warm distributions rather than one flattering
  number.
- **Correctness**: Safe code has defined behavior; optimizer attributes and FFI
  guarantees must derive from checked facts.
- **Low-level capability**: Usable ownership, borrowing, deterministic cleanup,
  target layout, and C interop are milestone requirements, not post-v1 add-ons.
- **Runtime posture**: No mandatory global tracing heap or service runtime; the
  first native slice uses stack/inline values and explicit resources.
- **Bootstrap**: Go 1.24 standard library hosts Stage 0; readable C17 emitted to
  installed Clang is the development native path.
- **Dependencies**: Prefer standard-library-only and shallow audited boundaries;
  every dependency must earn more than copy-local implementation.
- **Source authority**: Text remains Git-friendly and formatter-owned; parsing is
  lossless enough to preserve comments and produce stable recovery.
- **Evidence**: Deterministic fixtures, negative controls, properties, mutation,
  differential execution, and sanitizers answer distinct questions and have
  explicit cost lanes.
- **Portability**: macOS and Linux are the initial host priorities; wire contracts
  and target facts must not accidentally encode the current Apple arm64 host.
- **Security**: Untrusted input, unsafe operations, FFI, secrets, and build
  authority stay explicit; no generic `untaint` or ambient authority escape.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Build M001 as vertical source-to-native slices | Cross-layer feedback reveals semantic mistakes sooner than disconnected compiler layers | — Pending |
| Use Go 1.24 stdlib for Stage 0 | Fast builds, current spike evidence, simple offline onboarding, and low dependency depth | — Pending |
| Emit readable C17 through Clang first | Exposes real ABI and optimizer behavior while keeping backend replacement reversible | — Pending |
| Make the deterministic interpreter the semantic oracle | Prevents C output or optimizer accidents from defining the language | — Pending |
| Use affine values with separate copy/drop/share/send/escape abilities | Avoids conflating distinct semantic and cost properties | — Pending |
| Infer local borrow ends at CFG last use; declare public origins | Preserves local ergonomics without hiding public ABI relationships | — Pending |
| Validate compact typed-core facts independently at trust crossings | Gives cheap cache/release evidence without claiming source-to-core proof | — Pending |
| Keep syntax provisional but canonical | Enables real parser/formatter evidence without freezing decorative choices early | — Pending |
| Treat structured diagnostics and evidence as a versioned product API | AI effectiveness depends more on precise verifier feedback than exotic syntax | — Pending |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition:**
1. Move disproved active requirements to Out of Scope with evidence.
2. Move verified requirements to Validated with a phase reference.
3. Add requirements exposed by reduced counterexamples.
4. Record decisions and their measured outcomes.
5. Check that the description and core value have not drifted.

**After each milestone:**
1. Review all sections against shipped behavior.
2. Reconfirm the core value.
3. Audit deferred scope and preservation seams.
4. Update context with adopter, performance, and correctness evidence.

---
*Last updated: 2026-09-03 after formal GSD initialization*
