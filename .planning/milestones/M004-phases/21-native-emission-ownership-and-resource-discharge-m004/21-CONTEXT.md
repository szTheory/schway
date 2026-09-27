# Phase 21: Native Emission Ownership and Resource Discharge (M004) - Context

**Gathered:** 2026-09-25
**Status:** Ready for planning

<domain>
## Phase Boundary

Design checked resource-discharge and foreign-boundary ownership contracts needed before any emitter family cut from M003 can be reconsidered. Make the contract's acceptance cases and refusal rules machine-checkable so later work can verify completeness in CI.

This phase owns design and evidence-boundary work for D-16-11 (`emitLinearForeign`), D-16-12 (`emitLinearBorrowedByPointer`), D-16-13 (`emitLinearBorrowedByPointerPlain`), residual legacy-emitter retirement/convergence (D-11-02/D-12-36), and D-14-45's measured one-translation-unit/`-flto` boundary. The owner assignment does not admit or reopen any emitter family. Preserve each family's own prerequisites, cross-host evidence, refusal fence, and one-TU/`-flto` limitation. Do not claim that `-flto` becomes behaviorally non-inert.

</domain>

<decisions>
## Implementation Decisions

### Machine-checkable resource-discharge contract
- **D-21-01:** Phase 21 must define machine-checkable acceptance cases and refusal rules for the resource-discharge and foreign-ownership contract. Reserve executable native witnesses and any production emitter admission for a later phase that satisfies the family-specific prerequisites. Contract-structure checks establish completeness and consistency; they do not prove runtime cleanup behavior.
- Keep the single-emission-law direction from Phase 16. Do not introduce a second/coexisting ownership or discharge law merely to support an emitter family.

### Foreign-boundary exit and cleanup classification
- **D-21-02:** The checked contract must explicitly classify modeled foreign exits, including normal return, error return, supported unwind, nonlocal transfer, defects/cancellation, and process termination. Require an explicit, checked discharge rule for each exit that is admitted; refuse any path whose cleanup behavior is opaque or unproved.
- Classification does not itself introduce new Lang exception, cancellation, or unwind semantics. State which classes are supported, refused, or outside cleanup guarantees, and require family-specific evidence before any emitter is reopened.

### D-14-45 one-TU / `-flto` evidence scope
- **D-21-03:** Preserve the current one-TU/no-`restrict`/foreign-refusal boundary, add one scoped compiler comparison of an emitted multi-function fixture to substantiate D-14-45, and retain a cheap structural guard as recurring CI evidence. Broader recurring host/toolchain matrices require demonstrated regression value.
- Reuse the existing semantic comparator and its independent controls; do not create a second comparator. Distinguish behavioral equivalence under `-O3` and `-O3 -flto` from evidence about optimization activity or performance. Do not infer that one-TU LTO performs no optimizations, and do not imply that this work proves future foreign or by-pointer emitter behavior.
- Keep the measured claim bounded to the tested fixture, compiler/toolchain, and host lanes. Any widened claim needs corresponding measured evidence.

### the agent's Discretion
The researcher and planner may choose the contract encoding, structural assertions, fixture details, and CI lane placement if they preserve the decisions above, the existing refusal boundary, and honest evidence grades. They may not use the design artifact as authorization to route a cut emitter into production.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase boundary and project constraints
- `.planning/ROADMAP.md` §Phase 21 — goal, ownership assignments, and the no-admission boundary.
- `.planning/PROJECT.md` §Verification Operating Preference and §Current State — shift-left verification preference and the project's current LTO/evidence claims.
- `.planning/REQUIREMENTS.md` §Native Emission / NAT-09 — formal M004 ownership and emitter-family cut; does not claim `-flto` becomes non-inert.
- `.planning/STATE.md` — current phase position, no-loop rule, and durable project constraints.
- `.planning/STANDING-VERDICTS.md` — no coexisting laws, evidence anti-patterns, and the existing LTO/toolchain posture.

### Prior emitter decisions and debt
- `.planning/phases/16-branch-match-emitter-port/16-CONTEXT.md` — locked emission-law, refusal, cross-host, and cut-family decisions.
- `.planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md` — exact prerequisites, reopening conditions, witnesses, and `-flto` consequences for D-16-11 through D-16-13.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` — M003 emitter-family evidence and rationale for cutting the families.
- `.planning/phases/14-evidence-instrument-and-honest-scoping/PHASE-14-DEBT.md` §D-14-45 — owned multi-function LTO-inertness evidence boundary and its named witness.

### Existing implementation and evidence seams
- `internal/compiler/cgen/cgen.go` — legacy linear/foreign/by-pointer paths and resource-ledger precedent; do not assume those paths are production-admissible.
- `internal/compiler/cgen/cgen_program.go` — current whole-program admission refusals, derived `live_resources`, and multi-function output path.
- `internal/compiler/native/foreign_resource.go` — current native foreign-resource support seam.

### External technical references
- [Rustonomicon: FFI and unwinding](https://doc.rust-lang.org/nomicon/ffi.html) — ABI-specific unwind behavior; precedent for explicit boundary classification, not Lang semantics.
- [Clang ThinLTO documentation](https://clang.llvm.org/docs/ThinLTO.html) — describes cross-module optimization; does not prove behavior for Lang's emitted fixture.
- [LLVM Testing Infrastructure Guide](https://www.llvm.org/docs/TestingGuide.html) — regression and whole-program test layering precedent.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `cgen.deriveProgramLiveResources` and `emitProgramLiveResourcesWriter` — current whole-program path derives and serializes `live_resources`; useful seams for verifying the contract representation.
- `cgen.resourceLedger` in the legacy emitter — runtime accounting precedent for acquired/released resources and a release-omission mutation. It is not permission to restore a second production lowering path.
- Existing multi-function native fixture/comparator and `TestLTOInertnessOnMultiFunctionEmission` — starting points for the scoped LTO evidence task; verify exactly what they execute before relying on their names.

### Established Patterns
- `emitProgram` validates supported shapes and refuses foreign and by-pointer bodies before C serialization. Preserve fail-closed behavior until a later family-specific admission gate.
- Ownership and semantic claims are independently re-derived across compiler layers; acceptance evidence should state whether it checks document structure, static derivation, or executed native behavior.
- Evidence that is deterministic and recurring should move to the appropriate CI lane when its regression value justifies its runtime and maintenance cost.

### Integration Points
- `internal/compiler/cgen` owns native shape admission, resource derivation, C generation, and current `-flto` fixture hooks.
- `internal/compiler/native` contains foreign-resource execution support that can inform the contract, while not replacing independent C-emitter ownership rules.
- `.planning/phases/16-branch-match-emitter-port/PHASE-16-DEBT.md` is the authoritative source for per-family reopening gates and should remain synchronized with the Phase 21 design outcome.

</code_context>

<specifics>
## Specific Ideas

- Prefer a machine-checkable contract with structural CI guards and explicit negative/refusal cases. Keep behavioral cleanup proof attached to a later authorized emitter implementation.
- For foreign calls, do not assume every foreign exit can be unwound safely. Classify exits, prove cleanup for admitted paths, and refuse unmodeled paths.
- For D-14-45, add one scoped emitted-fixture comparison while preserving the current claim boundary. Report compiler and host provenance, compare semantic behavior with the existing comparator, and keep performance/optimization claims separate.
- External precedent supports these boundaries: Rust documents ABI-specific unwind behavior; Clang describes LTO as cross-module optimization; LLVM separates regression tests from whole-program execution tests. These sources inform design choices but do not define Lang semantics.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 21-native-emission-ownership-and-resource-discharge-m004*
*Context gathered: 2026-09-25*
