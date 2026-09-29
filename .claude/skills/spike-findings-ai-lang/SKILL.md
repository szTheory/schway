---
name: spike-findings-ai-lang
description: Implementation blueprint from spike experiments. Requirements, proven patterns, measured costs, and verified limits for building Schway's ownership kernel, loan-liveness dataflow, public origin summaries, certificate evidence, and native FFI contract. Auto-loaded during implementation work.
---

<context>
## Project: ai-lang (Schway)

**Idea: `ownership-kernel`** — turn the leading ownership semantics into a small
executable oracle before building a parser, compiler backend, runtime, or
user-facing syntax.

Six spikes were run against that idea. Four are VALIDATED (001, 002, 003, 006)
and two are PARTIAL (004, 005) — in both partial cases the *harness and contract*
are validated while the Lang-side claim is not. Spike 005's own conclusion marks
the end of disposable Go/C ownership workbenches: 001–005 define enough versioned
fixtures and failure modes to begin the real lossless frontend and portable typed
core, and the native harness should become a consumer of that core rather than
grow a parallel pseudo-language.

Read the reference file for the subsystem you are building. Read
`references/differential-verification-harness.md` before designing evidence for
any of them.

Spike sessions wrapped: 2026-09-17 (spikes 001–006, run 2026-09-03 – 2026-09-09)
</context>

<requirements>
## Requirements

Non-negotiable design decisions from the `ownership-kernel` idea. All six wrapped
spikes belong to this one idea, so the whole list applies; each reference file
repeats the subset it is accountable for.

### ownership-kernel

- Preserve affine ownership, shared/exclusive loans, explicit transfer,
  deterministic lexical release, and stable machine-readable diagnostics.
- Keep the spike dependency-free and cheap to run locally or in CI.
- Treat disagreement and minimized counterexamples as useful results.
- Do not infer final source syntax or production performance from kernel
  notation.
- Infer local loan endings across control flow without requiring source-level
  end markers on ordinary paths.
- Keep the production-shaped analysis polynomial while retaining bounded path
  expansion only as an independent oracle.
- Treat explicit source loan endings as outside the CFG-last-use slice rather
  than silently mixing manual and inferred policies.
- Export borrowed-result origins as verified value/field paths and tagged
  alternatives, with explicit shared/exclusive access in interface metadata.
- Keep `copy`, `drop`, `share`, `send`, and `escape` independent and derive their
  generic propagation structurally.
- Use higher-ranked fresh origins only at genuinely scoped callback/lending
  boundaries; do not force all borrowed results through callback-only APIs.
- Bind reusable ownership evidence to a canonical versioned typed-core subject.
  Use compact independent recomputation at package, cache, CI, and release
  boundaries; materialize full replay traces only on demand.
- Do not claim that a downstream certificate proves source-to-core correctness.
  A coordinated false typed-core statement remains inside the frontend trust
  boundary and must be attacked by different evidence.
- Treat target layout, initialization, allocator pairing, callback retention,
  address/provenance capture, cleanup, and unwind as explicit typed-core/FFI
  obligations. Derive optimizer attributes only from proven facts.
- Require unoptimized, optimized, semantic-oracle, and sanitizer evidence as
  distinct native lanes. Do not allow panic, unwind, or foreign nonlocal exit
  across an unaudited C ABI boundary.
- Memoize interprocedural summaries within a run: derive each function's summary
  once over the proven-acyclic call graph, never per call site. Cost must be
  declared against program operation count, not call-graph shape.
- Derive summaries from operations in program order. Any pass that reorders
  operations before summary derivation must restate the cost bound.
- Do not mark an interprocedural fact cacheable across runs before the
  invalidation fallout of a real callee edit is measured.
</requirements>

<findings_index>
## Feature Areas

| Area | Reference | Key Finding |
|------|-----------|-------------|
| Ownership kernel semantics | `references/ownership-kernel-semantics.md` | A finite lattice reaches loop fixed points without a solver; loan identities must be unique program-wide, and release evidence names the owning scope, not a depth |
| Loan liveness dataflow | `references/loan-liveness-dataflow.md` | A loan's inferred end can be a **control-flow edge**, not a source expression; and the interprocedural memo table is the mechanism, not an optimization — memoized is linear on every shape, the charitable unmemoized arm is 76x worse at 512 functions and doubling |
| Public interface: origins and abilities | `references/public-interface-origins-abilities.md` | Value-path origins plus tagged alternatives carry enough for a body-blind consumer; abilities stay independent through wrappers and erasure; summaries need immutable content-owned representations |
| Evidence and the trust boundary | `references/evidence-and-trust-boundary.md` | Compact recomputation (242 bytes) beats full replay (~1.0 MB) with no loss of soundness; a coordinated frontend lie escapes **by design**, and saying so is what keeps the claim honest |
| Native lowering and the FFI contract | `references/native-lowering-ffi-contract.md` | A false `restrict` promise is latent at `-O0` and changes the answer at `-O3`: optimizer attributes must derive from proven facts, never be asserted |
| Differential verification harness | `references/differential-verification-harness.md` | Independently written implementations plus armed fault injection; three of the most useful findings across all six spikes were defects in the harness itself |

## Source Files

Original spike source files are preserved in `sources/` for complete reference —
point-in-time snapshots at commit `d919e2f`. The canonical originals live at
`.planning/spikes/NNN-*/` in this repository; see `sources/README.md`.
</findings_index>

<metadata>
## Processed Spikes

- 001-ownership-kernel-workbench — VALIDATED
- 002-cfg-edge-last-use — VALIDATED
- 003-public-origins-generic-abilities — VALIDATED
- 004-independent-certificate-checker — PARTIAL
- 005-native-ffi-provenance-cleanup — PARTIAL
- 006-interprocedural-liveness-cost-scaling — VALIDATED
</metadata>
