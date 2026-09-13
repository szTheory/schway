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

## Current Milestone: M002 Interprocedural Semantic Spine

**Goal:** Land Lang-to-Lang calls and prove that every semantic guarantee M001
established intraprocedurally still holds across function boundaries.

**Target features:**
- `OpCall` as a real `OperationKind` at all six dispatch sites, gated on
  callable ⊆ publishable (D-04-03)
- Call-graph construction with cycle refusal and a bounded interpreter call stack
- Interprocedural loan liveness in both admission layers (`check` and
  `corevalidate`) — closes D-03-02
- Cross-function rebuild of Phase 3's exhaustive loan-endpoint differentials
- Interprocedural `-O3`/LTO equivalence on the five-axis comparator
- Storable/matchable `Result` values and payload-carrying alternatives (D-04-30)
- Close the Nyquist validation debt carried from Phases 3, 5, and 6

## Requirements

### Validated

- ✓ A clean checkout can format, check, interpret, and natively run a small
  canonical Lang program through one production command — M001 (Phase 1;
  `lang format` / `lang check` / `lang run --engine={interp,native}`).
- ✓ One portable typed core gives ownership, borrowing, cleanup, and foreign
  boundaries the same meaning in the checker, interpreter, and native path —
  M001 (Phases 2-4; `check`, `corevalidate`, `originvalidate`, `interp`, `cgen`
  independently derive the same facts).
- ✓ Interpreter, native `-O0`, and native `-O3` executions produce equivalent
  semantic outcomes and events for the milestone corpus — M001 (Phase 5;
  five-axis comparator, `-flto`-proven-non-inert `-O3` tier, ASan/UBSan lane,
  adversarial + bounded-enumeration corpora). Scope note: M001 ships without
  Lang-to-Lang calls, so this holds intraprocedurally by construction.
- ✓ Invalid programs fail early with stable, bounded, machine-readable cause
  graphs and calm human projections — M001 (Phases 1-3, 6; `lang.diagnostic`,
  `lang explain` cause DAG, rustfix-style `diagnostic.Repair` applicability).
- ✓ Compact content-bound evidence manifests make successful work reusable
  without treating frontend claims as proofs of source correctness — M001
  (Phases 1-2, 6; `evidence --validate`, `internal/compiler/cache` which is
  structurally incapable of holding a verdict).
- ✓ The default edit loop remains fast, deterministic, offline-capable, and
  measurable; expensive evidence runs only when its distinct question is
  relevant — M001 (Phase 6; `risk_lanes.json` changed-risk lane selection,
  `internal/compiler/measure` p50/p95/CoV protocol, `qlt02_budget_manifest.json`).

- ✓ Storable/matchable `Result` values with payload-carrying alternatives —
  M002 (Phase 12; `OpConstructPayload`/`OpDestructurePayload` with a single
  shared `core.AlternativeNameForPayloadType` derivation read by `check`,
  `interp`, and `cgen`, and affine drop obligation held on the unmoved
  alternative). Closes D-04-30, deferred from M001 Phase 4. Scope note: a
  `data` declaration whose alternatives collide on payload type is refused
  (`check.duplicate_payload_type`) rather than resolved — see D-12-44.

### Active

- [ ] Lang-to-Lang calls: `OpCall` as a real `OperationKind` at all six dispatch
      sites, gated on callable ⊆ publishable (D-04-03).
- [ ] Interprocedural loan liveness in both admission layers (`check` and
      `corevalidate`), rebuilt cross-function — closes D-03-02, the one debt
      item deliberately left open past M001.
- [ ] Call-graph construction with cycle refusal and a bounded interpreter call
      stack.
- [ ] Cross-function rebuild of Phase 3's exhaustive loan-endpoint differentials.
- [ ] Interprocedural `-O3` equivalence, which M001 could not claim because it
      ships without calls.

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
  readable C17/Clang remain reversible Stage 0 choices; M001 gave no reason to
  revisit either.

## Context

**Shipped M001 — Source-to-Native Semantic Spine (2026-09-07).** 6 phases,
62 plans, 174 tasks over 5 days; ~59,150 lines of Go plus ~916 lines of C/H
across `cmd/`, `internal/`, `native/`, `scripts/`, and `testdata/`.

What exists now is a real vertical compiler: lossless canonical frontend →
typed core with a per-function CFG → an affine ownership and borrow checker →
an independently re-implemented core validator (`corevalidate`) and published-
origin validator (`originvalidate`) → deterministic interpreter oracle →
readable C17 through Clang at `-O0`/`-O3`/`-flto`, with an ASan+UBSan lane, a
core-level HDD reducer, and a content-bound evidence/cache layer. Phase 6 added
the agent-facing surface: `lang explain`'s bounded cause DAG, rustfix-style
repair applicability, changed-risk lane selection, and a measured feedback
budget manifest.

**The one structural gap, stated plainly: M001 ships without Lang-to-Lang
calls.** `OpCall` was deliberately deferred rather than landed alongside
alias-fact emission, the first sanitizer lanes, the reducer, and the QLT-01
registry. Interprocedural `-O3` equivalence is therefore outside M001's proof
scope by construction, and Phase 3's D-03-02 (an exported borrow-derived return
with no declared origin, in its interprocedural half) is the single debt item
knowingly carried past the milestone. Both are M002's lead charter.

Remaining tech debt is registered per phase in `*-DEBT.md` and summarized in
`.planning/milestones/M001-MILESTONE-AUDIT.md`: nine Phase 2 items (mostly info-
level, four already closed in Phase 3), Phase 4's D-04-30 deferred `Result`
payloads and D-04-31's four accepted residual limitations (the coordinated
three-way lie, two nonlocal-exit blind spots, the single-host single-record-shape
fence, and the permanence of quarantine). Nyquist validation is compliant for
Phases 1, 2, and 4; Phases 3, 5, and 6 are not-validated.

The design corpus in `wiki/` records intent, primary-source research, synthesis,
decisions, examples, and open questions with explicit provenance. Five bounded
workbenches in `.planning/spikes/` tested local affine ownership, CFG
edge-specific loan liveness, separate-compilation origins and generic abilities,
independent typed-core certificate validation, and native C ABI/provenance/
cleanup hazards; their surviving contracts are now production code, and their
hazards are reconciled row-by-row in `qlt01_registry.json`.

Canonical planning inputs:

- `.planning/LANGUAGE-MATURITY.md` — where the language actually stands
  (expressiveness vs assurance) versus how the roadmap sounds. **Read this
  before answering any "how far along are we" question.**
- `.planning/STANDING-VERDICTS.md` — dependency verdicts, anti-features, and
  load-bearing facts already researched. Reopen only on new evidence.
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
| Build M001 as vertical source-to-native slices | Cross-layer feedback reveals semantic mistakes sooner than disconnected compiler layers | ✓ Good — every phase surfaced consumer gaps (e.g. 03-02's native execution-document validator) that a layered build would have hidden until integration |
| Use Go 1.24 stdlib for Stage 0 | Fast builds, current spike evidence, simple offline onboarding, and low dependency depth | ✓ Good — zero external deps through 6 phases; `go/ast`/`go/parser` doubled as a structural-test substrate |
| Emit readable C17 through Clang first | Exposes real ABI and optimizer behavior while keeping backend replacement reversible | ✓ Good — real `-O3`/LTO/`restrict` divergences and ASan findings only exist because the backend is real |
| Make the deterministic interpreter the semantic oracle | Prevents C output or optimizer accidents from defining the language | ✓ Good — the five-axis comparator's authority rests on it; it caught two real `--engine=native` bugs in Phase 4 |
| Use affine values with separate copy/drop/share/send/escape abilities | Avoids conflating distinct semantic and cost properties | ✓ Good — abilities derived independently through aggregates and generics; one accepted override (OV-02-01) |
| Infer local borrow ends at CFG last use; declare public origins | Preserves local ergonomics without hiding public ABI relationships | ✓ Good intraprocedurally — `loanLivenessFixpoint` became the sole liveness law after a 230,692-comparison shadow run; ⚠️ Revisit for the interprocedural half (D-03-02) |
| Validate compact typed-core facts independently at trust crossings | Gives cheap cache/release evidence without claiming source-to-core proof | ✓ Good — and its known escape is executable, not hypothetical (`escape:coordinated-source-to-core-false-claim`) |
| Keep syntax provisional but canonical | Enables real parser/formatter evidence without freezing decorative choices early | ✓ Good — losslessness and idempotence held through generated programs and malformed AI edits |
| Treat structured diagnostics and evidence as a versioned product API | AI effectiveness depends more on precise verifier feedback than exotic syntax | ✓ Good — the `/0`→`/1` schema bump landed across 12 literal sites with `/0` bytes provably frozen; `cmd/lang-repair` repairs 5 defect classes through the JSON protocol alone |
| Defer `OpCall` and interprocedural equivalence out of M001 | Landing a new `OperationKind` at six dispatch sites alongside alias facts, sanitizers, the reducer, and the QLT-01 registry matched the fingerprint of the failures that cost Phases 2-4 extra remediation rounds | ⚠️ Revisit — correct for M001's risk budget, but it is the reason M001 ships without Lang-to-Lang calls; M002's lead charter |
| Ratify D-12-43: accept that the decisive wrong-slot value-divergence control is unconstructible at this language maturity | The hazard is independently policed at compile time by D-12-37, and the current grammar never lets a match arm's result expose raw payload bytes — so a wrong-slot write is invisible to every comparator axis by construction, not by weak testing. Escalated as a defect in the criterion rather than a downgraded assertion (D-12-41/D-11-36 precedent) | ⚠️ Revisit when the grammar widens to expose bound payload content, or when a payload-value-aware harness observation channel exists; ratified at a blocking-human checkpoint on 2026-09-13, no owning phase named |

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
*Last updated: 2026-09-13 after Phase 12*
