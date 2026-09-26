# Phase 17: Return Type ≠ Parameter Type - Context

**Gathered:** 2026-09-22
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 17 removes the single-type-per-function invariant so a function may
declare a return type distinct from its parameter type. It must preserve sound
ownership facts through three independently-derived admission layers, execute
and lower through the sole surviving native-emission family, expose both
call-contract failures from source fixtures, and make `use_matching_argument`
a real, sealed, protocol-only repair class.

It does not add a new value form, a branch on a computed place, arity-N,
separate compilation, or DX-06/B1 blame closure.

</domain>

<decisions>
## Implementation Decisions

### Canonical two-type execution proof

- **D-17-01:** Use one checked-in, multi-function nominal-data
  `Resource -> Result` program as the canonical TYP-01 tracer. `classify`
  receives `Resource` and returns `Result`; `main` calls it and returns
  `Result`. This must check, interpret, lower, and agree across interpreter,
  `-O0`, `-O3`, and `-O3 -flto`. It exercises distinct AST/core type facts,
  call target resolution, generated-C prototypes/definitions/call sites, and
  entry output without duplicating the entire proof in the reverse direction.
- **D-17-02:** Do not represent this nominal `Resource` fixture as proof of
  tracked-resource ownership. A source-reachable `Buffer <-> Result` proof is
  not constructible on the current surface; claiming it would smuggle future
  payload/control-flow capability into this phase. A reverse-direction
  nominal regression is discretionary only if it catches a demonstrated
  directional defect.

### Source frontier diagnostics

- **D-17-03:** Create two minimal, parser-valid `.lang` frontier fixtures,
  each with one intended call-contract failure. The first isolates a
  caller/callee parameter mismatch and must move to
  `check.call_argument_type_mismatch`; the second has a matching argument but
  omits the callee return type from the caller's available facts and must move
  to `check.call_return_type_unrepresentable`. Pin the pre-widening refusal
  and assert that it moves; seam-level tests remain supplemental only.
- **D-17-04:** Make both diagnostics explain the contract distinction at the
  source call site using stable structured causes: callee, actual argument
  type, and declared parameter type for TYP-02; callee, declared return type,
  and the caller's available type facts for TYP-03. Do not let an incidental
  use-after-move or an earlier declaration error become the asserted proof.

### Independent return-ability derivation

- **D-17-05:** Derive parameter `Drops` and return `Fresh` locally and
  independently in `check`, `corevalidate`, and `originvalidate`, each from
  that layer's own inputs and type lookup. Sharing primitive ability rules is
  acceptable; sharing a return-contract helper, derived return facts, or a
  producer artifact is not. — **Reversibility:** costly — collapsing the
  derivations would invalidate TYP-04's independent-peer claim and require
  new evidence rather than a local refactor.
- **D-17-06:** Require return-only seeded mutations at each layer. A mutation
  must alter only the return-side ability/type lookup, leave parameter facts
  intact, and produce independently observable failure or divergence in that
  peer; a coordinated mutation must make the agreement gate fail. Extend
  import-boundary guards to cover any new production derivation files.

### Repair proof and permanent blame boundary

- **D-17-07:** Preserve Phase 13's sealed held-out call-argument fixtures as
  historical no-op/unrepairable controls. Create a new Phase 17
  derivation/held-out pair before repair emission: its held-out program has a
  real, uniquely in-scope parameter-typed alternative enabled by the widened
  return contract, structurally differs from derivation (at least the existing
  `main -> relay -> dispatch` depth pattern), and is sealed before the repair
  is implemented. — **Reversibility:** costly — changing a sealed fixture
  afterward would invalidate the anti-overfitting claim and require a new
  corpus and provenance record.
- **D-17-08:** The held-out proof must restore exact original source bytes,
  finish `repaired`, use exactly two real `lang --json check` subprocesses
  plus an independent clean re-check, and retain protocol-only checks for
  diagnosis, repair kind, span, replacement, and prose-scramble invariance.
  Add red controls for stripped repair fields/kind/span/replacement and a
  production-source scan forbidding held-out path references.
- **D-17-09:** Ratify D-13-02b as permanent until M006. Replace the obsolete
  `sameType`-based witness and P17 landing with an executable witness that
  proves no production `resolveBlame` path exists and all currently admitted
  signature fields are producer-verifiable. Its sole reopening condition is a
  user-declared contract field that the declaring function's own admission
  cannot verify (separate compilation); do not restate automatic DX-06 closure.

### Planning-input caveat

- **D-17-10:** S-009 has no recorded result. Treat it as a required planning
  input that calibrates appetite for later repair-surface expansion, not as a
  reason to weaken TYP-05 or as a blocker of the semantic type-widening work.

### the agent's Discretion

Exact nominal type names, fixture module names, helper names, test factoring,
and diagnostic prose are at the planner's discretion, provided the source
reachability, independent derivation, sealed-corpus, and structured-protocol
properties above remain true.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone and phase authority

- `.planning/ROADMAP.md` §Phase 17 — phase goal, TYP-01 through TYP-05,
  constructibility gate, scope exclusions, and S-009 planning-input note.
- `.planning/REQUIREMENTS.md` §Types — TYP-01 through TYP-05.
- `.planning/PROJECT.md` — M003 thesis, source-to-native evidence posture,
  corrected DX-06/DX-07 status, and explicit out-of-scope surface.
- `.planning/STATE.md` — phase position, emitter-before-widening ordering,
  and carried M003 constraints.
- `.planning/STANDING-VERDICTS.md` — dependency, optimizer, ownership, and
  evidence constraints that apply to every M003 phase.
- `.planning/LANGUAGE-MATURITY.md` — actual current surface and the limits
  that make a `Buffer <-> Result` source proof unconstructible today.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` §C1/C2 and P17 —
  authoritative widening inventory, sole-emitter rationale, constructibility
  precondition, and the permanent D-13-02b disposition.

### Prior decisions and evidence history

- `.planning/phases/14-evidence-instrument-and-honest-scoping/14-CONTEXT.md`
  — evidence grades, non-inertness controls, prose-free repair protocol, and
  the status of S-009.
- `.planning/phases/16-branch-match-emitter-port/16-CONTEXT.md` — the
  single surviving `emitProgram` lowering law that Phase 17 must extend once.
- `.planning/milestones/M002-phases/13-agent-loop-for-interprocedural-defects/PHASE-13-DEBT.md`
  — D-13-02b and D-13-10a provenance, existing seals, and prior no-op repair
  finding.

### Code and fixture anchors

- `internal/compiler/check/check.go` — three `sameType` admission heads,
  call binding, type-fact minting, and checker-side contract derivation.
- `internal/compiler/corevalidate/corevalidate.go` and
  `internal/compiler/originvalidate/originvalidate.go` — independent peer
  derivations and return-contract validation.
- `internal/compiler/core/core.go` — function signatures, type facts, and
  call-contract core schema.
- `internal/compiler/cgen/cgen_program.go` — surviving generated-C
  prototypes, definitions, call lowering, and output path.
- `internal/compiler/reduce/reduce.go` — the TypeID fallback that must not
  silently inherit the parameter fact after widening.
- `testdata/phase07/call_type_mismatch.lang` — established real-source
  argument-type refusal precedent.
- `testdata/phase13/heldout_call_argument_mismatch.lang` — historical
  sealed held-out topology to preserve as a control, not retrofit as success.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `core.Function`, `FunctionSignature`, and type-fact records already carry
  distinct parameter and return fields; Phase 17 must stop collapsing their
  facts into `functionID:type:0`.
- `resolveCallBinding` and its existing named diagnostic seams provide the
  integration point for real source reachability once callers can hold both
  parameter and callee-return facts.
- The Phase 13 derivation/held-out topology, SHA-256 sealing, injector, and
  repair-driver capture controls are reusable evidence machinery.
- `emitProgram` is the sole production C17 emitter after Phase 16, keeping the
  two-C-type model to one native family.

### Established Patterns

- A requirement needs a source fixture first, a pinned refusal, and a tested
  diagnostic movement; a seam alone proves only wiring.
- Peers independently derive facts and use structural import guards; agreement
  without a per-peer seeded fault is not evidence of independence.
- Repair consumes structured protocol data, never diagnostic prose; held-out
  corpus bytes are sealed and production code must not reference their paths.
- Named refusal codes, deterministic fixtures, and four-tier semantic
  comparison are preferred over permissive fallback or hand-waved agreement.

### Integration Points

- AST/core function type facts and all three `sameType` admission heads.
- Call target TypeID selection, match-arm return resolution, and reducer type
  fallback.
- Checker/corevalidate/originvalidate return contracts and their peer gates.
- `emitProgram` C declarations, function definitions, calls, entry input, and
  output rendering.
- `lang-repair` captures, injector/corpus seals, anti-theater checks, and the
  D-13 debt/unreachable-claims record.

</code_context>

<specifics>
## Specific Ideas

- The chosen tracer should resemble `classify(Resource) -> Result`, with
  `main` calling it, so Phase 18 can later consume a returned `Result` as a
  computed match scrutinee without Phase 17 claiming that feature now.
- “Sealed” is provenance protection against post-hoc alteration, not secrecy.
  Topology variation, source scans, alpha-rename controls, and protocol-field
  mutation controls remain necessary.

</specifics>

<deferred>
## Deferred Ideas

- A genuine `Buffer <-> Result` tracked-resource source proof — requires a
  future constructible surface; do not simulate it with a seam and call it
  Phase 17 evidence.
- Reverse-direction nominal two-type regression — add only if a real
  directional bug warrants its cost.
- DX-06/B1 blame wiring — M006, when separate compilation introduces a
  user-declared contract field the declaring function cannot verify.

</deferred>

---

*Phase: 17-Return Type ≠ Parameter Type*
*Context gathered: 2026-09-22*
