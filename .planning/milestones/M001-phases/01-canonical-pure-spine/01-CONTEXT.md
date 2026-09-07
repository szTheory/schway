# Phase 1: Canonical Pure Spine - Context

**Gathered:** 2026-09-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Deliver the smallest honest production compiler vertical: one formatter-owned
Lang source file is losslessly parsed, lowered to a nominal typed core, checked,
interpreted, emitted as readable C17, built by Clang at `-O0` and `-O3`, and
compared through stable semantic outcomes and a compact content-bound manifest.

This phase proves the architecture and protocol with a pure closed-variant
transformation. It does not implement ownership, borrowing, resources, effects,
general-purpose grammar, packages, or a runtime.

</domain>

<decisions>
## Implementation Decisions

### Provisional Source Surface
- **D-01:** Use the conventional braced “semantic prose” candidate for S1:
  `module`, explicit `export { ... }`, `data`, `fn`, braces, and exhaustive
  `match`. This is robust under malformed edits, familiar to humans/models, and
  remains a projection over the core rather than its specification.
- **D-02:** Parameters are always labeled at call sites; exact-name punning is
  canonical when local and parameter names match. No positional calls,
  user-defined operators/precedence, macros, sigils, implicit conversions, or
  alternate block syntaxes enter S1.
- **D-03:** The formatter owns one output. It preserves comments through a
  lossless CST, is idempotent, and never reorders expressions or named arguments.

The representative source shape is:

```text
module examples.switch

export {
  type Switch
  fn toggle
}

data Switch =
  | Off
  | On

fn toggle(state: Switch) -> Switch {
  match state {
    Off => On
    On => Off
  }
}
```

### Repository and Compiler Boundaries
- **D-04:** Use one Go module at the repository root with production entry point
  `cmd/lang` and private compiler packages under `internal/compiler/`. Keep
  fixtures under `testdata/`; do not move or import the disposable Go spikes.
- **D-05:** Name the executable `lang` while the language remains codenamed
  Lang. Package boundaries follow semantic stages—syntax, core, checking,
  interpretation, C emission, evidence—but S1 may combine packages when a split
  has no independent contract yet.
- **D-06:** Core and evidence wire shapes are explicit versioned structs with
  canonical JSON tests. Internal Go data structures are not accidentally exposed
  as the protocol.

### Command and Diagnostic Contract
- **D-07:** S1 exposes `lang format`, `lang check`,
  `lang run --engine=interpreter|native`, and `lang evidence`. Human output is
  the default; `--json` selects the versioned machine projection.
- **D-08:** Successful machine output goes to stdout. Human diagnostics and
  operational failure messages go to stderr. Exit codes distinguish success,
  invalid source, tool failure, and semantic-engine mismatch.
- **D-09:** Every diagnostic has a stable code, primary span, concise cause
  chain, and bounded related facts. Wording may improve without changing the
  identity. S1 must include a missing-match-alternative negative control.

### Evidence and Cost Discipline
- **D-10:** The interpreter defines semantic outcome; C17 and optimization are
  implementations checked against it. Wall time, addresses, temporary paths,
  and compiler process details are operational facts excluded from equivalence.
- **D-11:** Default tests include deterministic fixtures, formatter/CST
  properties, canonical-JSON properties, stale-manifest mutation, and
  interpreter/`-O0`/`-O3` differential execution. Each lane must name the defect
  class it can detect.
- **D-12:** Phase 1 records cold/warm latency, peak RSS, output bytes, and
  recomputed-work counts. Provisional investigation thresholds are warm
  format/check/interpreter p95 under 100 ms, native p95 under 1 s, and compiler
  peak RSS under 128 MiB on the current declared Mac. These are guardrails, not
  cross-machine release promises; Phase 6 ratifies portable budgets.
- **D-13:** Cache only successful hermetic products keyed by source/core/schema,
  compiler, target, flags, and policy. S1 may implement the manifest before a
  persistent cache; no cached green result can suppress a changed negative
  control.

### the agent's Discretion
- Exact punctuation inside the selected braced surface, provided the example
  round-trips canonically and no new semantic feature is introduced.
- The smallest useful Go package split and internal representation details.
- Benchmark harness mechanics, diagnostic prose, and C symbol spelling.
- Whether the first CLI build is invoked by `go run`, `go build`, or a small
  repository task wrapper, as long as clean offline reproduction is one command.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone and semantics
- `wiki/real-frontend-core-ir-entry-plan.md` — authoritative vertical outcome,
  S1 boundary, bootstrap choice, gates, and stop conditions.
- `wiki/semantic-kernel-contract.md` — evaluation, nominal ADTs, matching,
  outcomes, event oracle, and facts that must remain backend-independent.
- `wiki/semantic-kernel-probes.md` — probe/evidence records, blocking negative
  controls, and differential comparison contract.

### Syntax and feedback
- `wiki/syntax-and-cognitive-ergonomics.md` — braced surface candidate,
  canonical formatting, named arguments, and alternatives considered.
- `wiki/example-tour.md` — representative module/export/data/function/match
  forms; examples are candidates, not a frozen grammar.
- `wiki/compiler-and-feedback-latency.md` — incremental identities,
  diagnostics, cache boundaries, and latency measurement principles.
- `wiki/compute-efficiency-constitution.md` — measure useful feedback, avoid
  redundant lanes, and account for compute across AI, CI, and runtime.

### Prior evidence
- `.planning/spikes/MANIFEST.md` — validated experimental claims and explicit
  limitations across all five spikes.
- `.planning/spikes/004-independent-certificate-checker/README.md` — compact
  trust-crossing validation and the coordinated frontend-lie limitation.
- `.planning/spikes/005-native-ffi-provenance-cleanup/README.md` — native
  harness pattern, O0/O3 comparison, sanitizer separation, and host toolchain.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- Spike fixture concepts and mutation operators: migrate claims and expected
  outcomes into production-owned fixtures; do not import spike packages.
- Spike 004 canonical hashing/validation lessons: reuse the contract, then
  implement a deliberately smaller S1 manifest over real core facts.
- Spike 005 native harness mechanics: reuse the observed Clang invocation and
  normalized event comparison pattern without making its hand-written C the
  backend.

### Established Patterns
- Go 1.24 standard-library-only modules with deterministic JSON fixtures,
  seeded generation, bounded exhaustive cases, race/vet/coverage lanes, and
  small command programs.
- Every spike documents the seeded defect, the independent oracle, performance
  observations, and its exact blind spot.
- Wiki notes separate intent, research, synthesis, decisions, candidates, and
  unresolved questions through stable frontmatter and ledger references.

### Integration Points
- New production code starts at repository-root `go.mod`, `cmd/lang`,
  `internal/compiler/`, and `testdata/`.
- The root `README.md` and wiki index should link to the working compiler only
  after S1 actually runs.
- `.planning/spikes/` remains read-only experimental evidence during Phase 1.

</code_context>

<specifics>
## Specific Ideas

- The first source should be visually calm enough to review directly but small
  enough that parser recovery, canonical formatting, typed lowering, two engines,
  and evidence all land together.
- Stable machine identities and compact output matter more than minimizing raw
  source tokens.
- A compile-success claim is insufficient: the deliberately incomplete match,
  stale manifest, and engine divergence controls must demonstrably fail.

</specifics>

<deferred>
## Deferred Ideas

- Ownership transfer and abilities — Phase 2.
- Borrowed views, public origins, and CFG edge-specific last use — Phase 3.
- `Result`, partial resource initialization/cleanup, and real C boundary
  obligations — Phase 4.
- Full adversarial native/sanitizer matrix — Phase 5.
- Query/explain/verify repair loop and ratified performance budgets — Phase 6.
- Effects, concurrency/runtime, trust policy, ecosystem kits, and applications —
  post-M001 milestones.

</deferred>

---

*Phase: 01-canonical-pure-spine*
*Context gathered: 2026-09-03*
