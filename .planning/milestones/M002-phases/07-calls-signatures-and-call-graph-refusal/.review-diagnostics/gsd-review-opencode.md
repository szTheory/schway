I’m checking the cited source files against the plan claims before I review the plans.
I’ve confirmed the key code paths the plans talk about. I’m checking the test scaffolding and isolation patterns next.
I’ve found the main test and independence hooks. I’m checking one structural assumption now: whether the core artifact actually carries the export information the later plans want to re-derive from.
The core artifact itself doesn’t carry an export set, so I’m checking whether the plan’s wording matches some other in-tree representation or if that’s a real gap.
I found a concrete mismatch: the core IR doesn’t keep an export set. I’m confirming whether the checker still retains that fact anywhere, or whether the plan is assuming a field that doesn’t exist.
I have the likely blocker: the planned peer needs export metadata, but the core IR doesn’t obviously carry it. I’m grabbing the AST/source-side representation to separate “needs a new core field” from “already present somewhere else.”
I’m verifying whether exports are already threaded through the checker. If they aren’t, that confirms the plan needs a structural addition, not just a re-derivation.
**Summary**
The phase 07 plan is well-structured: it uses explicit gates, phase-scoped control lists, mutation-kill seams, and independent derivations in the right places. The biggest problem is in 07-03: it assumes `Callable` can be re-derived from a core-only artifact, but the current core IR does not preserve publishability/export data, so that independence claim is not implementable as written.

**Strengths**
- Good reuse of proven patterns: the mutation seam shape matches `pathoracle`’s test-only override pattern and four-beat mutation test style (`pathoracle.go:51-66`, `pathoracle_test.go:268-295`).
- Good avoidance of false coverage: the plan explicitly recognizes `cgen`’s `len(Functions) == 1` gate, so it does not pretend `OpCall` emission is exercised when it cannot be (`core_test.go:347-352`, `session.go:2560-2565`).
- Good control drift protection: the phase-07 lane/script approach mirrors the Phase 5 set-equality pattern (`session_phase5.go:20-27`, `session_phase5_test.go:291-295`).
- Good diagnostic discipline: the plan accounts for cause-order identity and bounded truncation with the same kind of load-bearing semantics already used elsewhere (`diagnostic.go:108-117`, `protocol.go:123-129`, `session_phase6_evidence.go:21-25`).

**Concerns**
- **HIGH**: 07-03’s `Callable` peer cannot be implemented from the current core shape alone. Source exports live in `ast.Program.Exports` (`ast.go:5-9, 47-51`), but `check.Program` lowers into `core.Program` with only `Schema`, `Module`, `ModuleID`, `DataTypes`, and `Functions` (`check.go:44-125`, `core.go:10-16`). `core.Function` also has no publication marker (`core.go:25-47`). The plan needs a durable core-side publishability fact or the peer independence claim breaks.
- **MEDIUM**: The plan introduces a lot of phase-scoped lists, scripts, and tests across 07-01 to 07-05. That is correct shape, but it is high drift risk because the repo already relies on strict duplication checks to keep these lists aligned (`session_phase5.go:20-27`, `session_phase5_test.go:291-295`).
- **MEDIUM**: The cycle-refusal work depends on two independent graph walkers, shadowing precedence, and several mutation seams. That is the right design, but it is easy to accidentally share logic or miss a root-order edge case unless the synthetic fixtures are extremely strict.
- **LOW**: Some acceptance around `cgen` is documentation-only by design, not behavioral. That is fine, but the plan should keep saying that explicitly so a green control is not mistaken for real multi-function codegen coverage.

**Suggestions**
- Add a concrete core-level publishability marker before 07-03, or revise `Callable` so it is derived from a fact already preserved in `core.Program`.
- Keep the phase-07 control lists and verify scripts under strict set-equality tests, like Phase 5, to prevent silent drift.
- Make the synthetic fixtures the primary truth for 07-04/07-05, especially the diamond/shared-leaf corpus and the shadowing case.
- Add one explicit note in the plan summary that `cgen`’s `OpCall` arm is hygiene-only this phase, not runtime coverage.

**Risk Assessment**
**HIGH**. The overall structure is strong, but the `Callable` independence assumption is currently unsatisfied by the core shape, and that fact sits on the critical path for the later call-admission and cycle-refusal phases.
