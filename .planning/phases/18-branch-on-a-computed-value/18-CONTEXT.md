# Phase 18: Branch on a Computed Value - Context

**Gathered:** 2026-09-24
**Status:** Ready for planning

<domain>
## Phase Boundary

Allow a source program to bind a computed `data` value and branch on that
in-scope place by generalizing the existing terminal `match` form. Deliver
CTL-01 through CTL-03, including a Result-returning callee, payload-place
returns, the seeded wrong-slot control, and a production-path rerun of the
S-010 borrow-across-branch shape. Preserve the existing five-axis semantic
comparison.

This phase does not add `if`/`else`, `Bool`, arithmetic, loops, aggregates,
arity-N calls, separate compilation, or a second control-flow law.
</domain>

<decisions>
## Implementation Decisions

### Computed scrutinee and source boundary

- **D-18-01:** Follow the ratified Phase 18 form: a linear body's terminal
  `match` may discriminate any in-scope place of a `data` type, including a
  value introduced by `let`; do not add `if`/`else` or a separate branch form.
  Keep the accepted type, arm, and body shape within the Phase 18 requirements.
- **D-18-02:** Start fixture-first. Check in the refused frontier `.lang`
  fixture and pin its current diagnostic before implementation; close by proving
  that diagnostic moved on the production source-to-core path.

### Ownership, calls, and native evidence

- **D-18-03:** S-010 is answered clean by spike 007 for the specified CFG
  topology: the existing `edge` and `point` endpoint kinds suffice and the
  mutation control kills an invented third kind. Treat that as permission to
  plan, not as production-path proof; satisfy Phase 18 success criterion 4 on
  production code and retain the declared liveness work bound.
- **D-18-04:** Preserve all three CTL requirements. In particular, exercise a
  Result-returning callee across the existing five comparator axes, return a
  destructured payload place, and make the seeded wrong-slot write observable
  at `axis:terminal-outcome`. Do not substitute seam-only or prose-only proof
  for source-level fixtures and executable controls.
- **D-18-05:** Automate objective acceptance by default: add focused source,
  integration, differential, smoke, and mutation-kill checks where each has
  recurring regression value; run recurring checks in CI when their
  maintenance and runtime cost are justified. Reserve human UAT for judgment
  or access the agent cannot supply. This carries forward the user's standing
  GSD preference recorded in `.planning/PROJECT.md`.

### the agent's Discretion

Parser/core representation, helper boundaries, execution ordering, fixture
names, and test factoring are open to the planner, provided the requirements,
independent admission checks, existing endpoint vocabulary, five-axis evidence,
and zero-external-production-dependency constraint are preserved. Prefer one
authoritative control-flow law and do not retain a parallel legacy law.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Phase, requirements, and project authority

- `.planning/ROADMAP.md` §Phase 18 and §Pre-Phase Spikes — goal, hard gate,
  contingency, success criteria, five-axis proof, production S-010 rerun, and
  fixture-first discipline.
- `.planning/REQUIREMENTS.md` §Control Flow — CTL-01, CTL-02, CTL-03 and the
  S-010 contingency.
- `.planning/PROJECT.md` — source-to-native core value, correctness and
  runtime constraints, and standing preference for automated verification and
  CI when recurring value justifies cost.
- `.planning/STATE.md` — settled Phase 18 planning position and M003 decisions.
- `.planning/STANDING-VERDICTS.md` — no external production dependencies,
  ownership and optimizer constraints, and evidence anti-patterns.
- `.planning/LANGUAGE-MATURITY.md` — actual source surface and explicit absence
  of `if`, arithmetic, loops, and computed-place branching before this phase.

### S-010 and authoritative analysis

- `.planning/spikes/007-loan-across-branch/README.md` — validated clean S-010
  result, exact existing endpoint kinds, focused test, and mutation witness.
- `.planning/spikes/MANIFEST.md` — registered spike verdict and scope.
- `.planning/research/M003/ADVERSARIAL-SYNTHESIS.md` §C3 and Phase 18 analysis
  — why matching a computed place is real control-flow work and why payload
  bytes, not arithmetic, make D-12-43 constructible.

### Prior decisions and implementation seams

- `.planning/phases/16-branch-match-emitter-port/16-CONTEXT.md` — sole
  `emitProgram` lowering law, validation ordering, and cut boundaries.
- `.planning/phases/17-return-type-parameter-type/17-CONTEXT.md` — distinct
  parameter/return types and source-reachable Result call proof.
- `.planning/EVIDENCE-RECONCILIATION.md` — generated evidence view and its
  regeneration contract.
- `.planning/UNREACHABLE-CLAIMS.md` — generated claims view and its
  regeneration contract.
- `.planning/MILESTONES.md` — milestone intent and archived decision history.
- `internal/compiler/syntax/parser.go` and `internal/compiler/ast/ast.go` —
  current function-body and match AST shapes.
- `internal/compiler/check/check.go` — current scrutinee restriction, branch
  construction, arm analysis, and loan endpoint/liveness pipeline.
- `internal/compiler/core/core.go`, `internal/compiler/corevalidate/`, and
  `internal/compiler/originvalidate/` — typed-core structure and independent
  admission peers.
- `internal/compiler/interp/` and `internal/compiler/cgen/cgen_program.go` —
  existing execution and sole native-emission paths.
- `internal/compiler/pathoracle/` — bounded path semantics and cycle refusal.
- `.github/workflows/ci.yml` — recurring macOS/Linux build and Go test lanes.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets

- `checkBranch` and `analyzeArmBody` already build match CFGs and analyze arm
  bodies; `loanLivenessFixpoint` and `materializeLoanEndpoints` implement the
  existing point/edge distinction.
- Phase 12 payload tests and Phase 16 `emitProgram` tests provide the nearest
  payload-layout and native-lowering precedents.
- The five-axis comparator, independent validators, and CI workflow already
  exist and should be extended rather than duplicated.

### Established Patterns

- Compiler acceptance is source-fixture-first with a pinned refusal that moves
  only after production code reaches the feature.
- The interpreter is the semantic oracle; `-O0`, `-O3`, and `-O3 -flto` are
  distinct native evidence tiers.
- Independent validators derive facts from their own inputs. Mutation controls
  must demonstrate that their intended guards can fail.

### Integration Points

- Syntax/parser and AST body variants feed `check` and the independent core
  validators; admitted core feeds interpreter, path oracle, and `emitProgram`.
- Match payload output and terminal execution evidence feed the existing
  schema-2 comparison and the D-12-43 wrong-slot witness.
- `.github/workflows/ci.yml` runs the project-wide build and test suite on
  macOS and Linux.
</code_context>

<specifics>
## Specific Ideas

S-010's clean spike result answers only the bounded CFG algorithm question.
The phase must still prove the same shape through production source parsing,
checking, execution, and native lowering.
</specifics>

<deferred>
## Deferred Ideas

- `if`/`else`, `Bool`, arithmetic operators, and loops remain outside M003 per
  the roadmap; do not pull them into this phase.
- Keep the source/API surface within the established single-argument and
  closed-data-alternative limits.
</deferred>

---

*Phase: 18-branch-on-a-computed-value*
*Context gathered: 2026-09-24*
