# Phase 16: Branch/Match Emitter Port - Context

**Gathered:** 2026-09-19
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 16 makes `emitProgram` the one native-emission law for every program it
admits: it ports match, branch, and ordinary linear lowering from the legacy
single-function family; removes the function-count routing fork; and deletes
`emitMatch`, `emitBranch`, and `emitLinear` in the same cutover commit.

The phase also records an explicit M004 cut for foreign lowering and any
by-pointer family not earned by the narrowly bounded `restrict` probe. It does
not introduce a resource ledger, caller-side alias-discharge protocol, new
ownership semantics, a new control-flow form, arity-N, or a general optimizer
attribute policy.

</domain>

<decisions>
## Implementation Decisions

### One emission law and atomic cutover

- **D-16-01:** Use a narrow capability port: move `emitMatch`, `emitBranch`
  (including its arm lowering), and ordinary `emitLinear` behavior into
  `emitProgram`; then make `emitProgram` the only production route for every
  admitted program. Do not rewrite all six legacy emitters and do not retain an
  N=1 production fork. — **Reversibility: one-way** — the intended deletion is
  an architectural authority cut; undoing it would recreate two independently
  maintained lowering laws and force Phase 17 to pay its two-type model twice.
- **D-16-02:** Establish three pre-deletion convergence gates, stop at a
  `checkpoint:decision`, then flip dispatch and delete `emitMatch`,
  `emitBranch`, and `emitLinear` in the same commit. Preserve
  `TestN1ConvergenceDifferential` as the characterization test; flip its
  expectations instead of deleting it.
- **D-16-03:** Preserve Phase 15's validation order: graph/entry validation,
  supported-shape validation, invocation/output preflight, then serialization.
  Adding match support must not change the diagnostic precedence of remaining
  unsupported shapes.
- **D-16-04:** Port match payload lowering through the existing checker-derived
  layout facts. Preserve tagged-struct representation and the `_Noreturn`
  defect exemption; do not independently re-derive payload layout or mint
  generic optimizer attributes.
- **D-16-05:** `live_resources` must be derived by the surviving emitter. A
  hardcoded empty JSON literal is not admissible after the broader port.

### Narrow `restrict` probe

- **D-16-06:** Run the optional probe before planning. It is a test-only,
  hand-written C17 microprogram that matches only the prospective emitted
  shape: one translation unit; a `static T f(T *restrict p)`; one pointer;
  read/copy of `*p` only; caller local access only before or after the call.
  Run it on macOS and Linux at `-O0`, `-O3`, and `-O3 -flto`, with the applicable
  sanitizer lane.
- **D-16-07:** Admit `emitLinearBorrowedByPointer` only if the normative and
  measured probe proves the exact shape and the emitter structurally refuses
  extensions: no pointee mutation, second pointer, pointer escape/forwarding,
  callback, foreign call, volatile/atomic access, or separate compilation.
  Otherwise cut it to M004 with the full discharge-pair design. This decision
  never generalizes `restrict` from the present read-only case to future richer
  aliasing.
- **D-16-08:** A successful probe removes only the former
  zero-optimizer-visible-attribute rationale. It does not establish that LTO
  is behaviorally non-inert; that evidence remains a distinct obligation.

### Compatibility and golden-C provenance

- **D-16-09:** Before deletion, require the legacy and `emitProgram` paths to
  emit byte-identical C for exactly `testdata/phase1/toggle.lang`,
  `testdata/phase2/owned_transfer.lang`, and
  `testdata/phase3/borrowed_view.lang`, for both applicable `Emit` and
  `EmitNative` modes. Keep the established four-tier semantic comparator as
  independent defense in depth. Do not require identity of explicitly cut
  foreign or unadmitted by-pointer shapes.
- **D-16-10:** Create a checked-in, machine-linked four-entry golden-change
  ledger adjacent to `previousPhaseGoldenCDigests`. Each entry names the path,
  old and new SHA-256, responsibility moved from legacy to `emitProgram`, exact
  structural reason, semantic witness, N=1 fixture, and review disposition.
  CI must fail for missing, duplicate, stale, or digest-map-mismatched entries.
  A digest proves integrity, not the rationale for a generated-C change.

### Formal M004 cut

- **D-16-11:** Formally amend `REQUIREMENTS.md` under D-10-60's amendment
  clause for `emitLinearForeign`, `emitLinearBorrowedByPointer`, and
  `emitLinearBorrowedByPointerPlain` unless D-16-07 admits the first
  by-pointer family. Refile each excluded family as an M004 debt row with an
  owner, real prerequisite, reopening condition, and honest `-flto` consequence.
  Never silently relabel a third deferral.

### the agent's Discretion

Exact helper names, test factoring, ledger encoding, and diagnostic prose are
at the planner's discretion, provided they preserve the required scope,
validation order, identity gates, structural refusal boundaries, and evidence
linkage above.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase and milestone authority

- `.planning/ROADMAP.md` §Phase 16 — scope, NAT-08/NAT-09 success criteria,
  same-commit deletion protocol, cut families, and optional probe.
- `.planning/REQUIREMENTS.md` §Native Emission — NAT-08 and NAT-09; amendment
  must preserve their explicit scope.
- `.planning/PROJECT.md` — M003 runtime posture, governing constructibility
  gate, and why Phase 16 precedes type widening.
- `.planning/STATE.md` — current phase position and carry-forward M003
  constraints.
- `.planning/STANDING-VERDICTS.md` — optimizer, LTO, ownership, dependency,
  and evidence constraints.
- `.planning/LANGUAGE-MATURITY.md` — actual language surface and exclusions.

### Authoritative prior analysis and debt

- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` §C1/C2 — measured emitter
  families, ordering decision, and port-then-delete rationale.
- `.planning/research/M003/EMISSION-AND-EVENT-IDENTITY.md` §Cluster A and
  §D-10-60 — emitter subset analysis, formal-cut mechanism, and debt discipline.
- `.planning/milestones/M002-phases/10-trusted-interprocedural-oracle/PHASE-10-DEBT.md`
  §D-10-60 — amendment clause governing a twice-deferred item.
- `.planning/milestones/M002-phases/11-multi-function-native-emission-and-interprocedural-equivalen/PHASE-11-DEBT.md`
  §D-11-02 — N=1 convergence differential and golden baseline obligations.
- `.planning/milestones/M002-phases/12-result-payloads/PHASE-12-DEBT.md` §D-12-36
  — carried emitter work and its disposition.
- `.planning/phases/15-event-identity-lang-execution-2/15-CONTEXT.md` —
  `emitProgram`'s `/2` invocation table and event semantics that the port must
  preserve.

### Implementation and evidence anchors

- `internal/compiler/cgen/cgen.go` — current N=1 dispatch and legacy emitter
  family.
- `internal/compiler/cgen/cgen_program.go` — surviving whole-program emitter,
  validation/preflight order, and `/2` native behavior.
- `internal/compiler/cgen/cgen_n1_convergence_test.go` — five-shape measured
  convergence characterization that must flip, not disappear.
- `internal/compiler/cgen/cgen_test.go` — frozen generated-C golden patterns.
- `internal/compiler/cgen/cgen_program_test.go` — whole-program golden and
  preflight test precedents.

### External normative references

- `https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf` §6.7.3.1 — C17
  `restrict` semantics used only to bound the probe.
- `https://llvm.org/docs/LinkTimeOptimization.html` — LTO behavior; not a
  substitute for project-specific non-inertness evidence.
- `https://llvm.org/docs/TestingGuide.html` — focused compiler-test practice.
- `https://reproducible-builds.org/docs/` — provenance distinction between a
  digest and the documented inputs/reasons that make it meaningful.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `internal/compiler/cgen/cgen_program.go`: the whole-program C17 emitter owns
  one translation unit, call lowering, static invocation paths, and
  `lang.execution/2` emission; it is the surviving integration point.
- `internal/compiler/cgen/cgen.go`: legacy `emitMatch`, `emitLinear`,
  `emitBranch`, `emitBranchOperations`, and the two public dispatch functions
  provide the behavior to characterize, port, and delete.
- `internal/compiler/cgen/cgen_n1_convergence_test.go`: direct unexported
  `emitProgram` access and five checked fixtures already form the cutover oracle.
- `internal/compiler/cgen/cgen_test.go` and `cgen_program_test.go`: exact-C
  goldens and deterministic, fixture-based test patterns.

### Established Patterns

- The checker remains the one source for semantic/payload layout facts;
  downstream peers derive independently rather than importing its artifacts.
- Evidence must be EXERCISED or better, with seeded-fault or structural
  non-inertness proofs; hashes alone cannot justify a re-baseline.
- Boundaries use named refusals and explicit, owned debt—not permissive
  fallbacks or silent deferred work.
- Phase 15's static occurrence-path table and schema `/2` event ordering are
  load-bearing contracts, not emitter-local implementation details.

### Integration Points

- `Emit` and `EmitNative` dispatch in `cgen.go`.
- `emitProgram` supported-shape admission, preflight, function emission, and
  generated JSON tail in `cgen_program.go`.
- Payload arm operations in `emitBranchOperations`.
- N=1 differential, golden-C digests, four-tier comparator fixtures, and debt
  register well-formedness tests.

</code_context>

<specifics>
## Specific Ideas

- Use an adversarial, evidence-backed port rather than a cosmetic refactor:
  byte identity establishes characterization of deterministic C output, while
  four-tier execution prevents a shared emitted-C defect from looking green.
- Treat `restrict` as a local C contract with an explicit source-shape fence;
  do not equate a Lang loan with a future-proof C aliasing promise.
- The golden ledger is a small provenance control: it records why each changed
  byte is allowed, not merely that its new hash was pinned.

</specifics>

<deferred>
## Deferred Ideas

None — the foreign resource ledger, richer by-pointer/discharge semantics, and
any remaining cut emitter family are explicitly routed to M004 rather than
being silently carried forward.

</deferred>

---

*Phase: 16-Branch/Match Emitter Port*
*Context gathered: 2026-09-19*
