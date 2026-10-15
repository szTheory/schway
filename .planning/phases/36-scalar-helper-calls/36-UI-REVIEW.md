# Phase 36 — UI Review

**Audited:** 2026-10-15  
**Baseline:** Applicability check against the Phase 36 context and abstract six-pillar standards; no UI-SPEC.md exists.  
**Screenshots:** Not captured — this phase has no frontend surface or frontend files.  
**Interaction captures:** off (`workflow.ui_interaction_capture=false`)

## Applicability

Phase 36 implements scalar helper calls across the Schway parser, checker, validators, interpreter, native C emitter, and conformance evidence. The phase boundary describes this compiler capability and assigns the public interpreter file CLI and executable guide to Phase 37 ([36-CONTEXT.md:10-12](./36-CONTEXT.md#L10)). The eight plan inventories contain Go source/tests, `.schway` fixtures, JSON evidence, and planning records ([36-01-PLAN.md:7-10](./36-01-PLAN.md#L7), [36-02-PLAN.md:7-14](./36-02-PLAN.md#L7), [36-03-PLAN.md:7-13](./36-03-PLAN.md#L7), [36-04-PLAN.md:7-11](./36-04-PLAN.md#L7), [36-05-PLAN.md:7-16](./36-05-PLAN.md#L7), [36-06-PLAN.md:7-15](./36-06-PLAN.md#L7), [36-07-PLAN.md:7-15](./36-07-PLAN.md#L7), [36-08-PLAN.md:7-12](./36-08-PLAN.md#L7)).

The repository scan found no frontend source extensions, UI specification, component registry, or frontend build configuration. The canonical `check ui.safety-gate 36 --raw` result was `frontend: false`, `hasUiFiles: false`, `hasUiSpec: false`, `block: false`. Therefore all six pillars and the overall score are **not applicable**; numeric scores would imply a visual or interaction surface that this compiler-only phase does not have.

## Pillar Scores

| Pillar | Score | Key Finding |
|--------|-------|-------------|
| 1. Copywriting | N/A | No interface copy surface is in this phase's scope. |
| 2. Visuals | N/A | No rendered frontend components are in this phase's scope. |
| 3. Color | N/A | No UI styles, theme, or color tokens are present. |
| 4. Typography | N/A | No interface typography system is present. |
| 5. Spacing | N/A | No UI layout or spacing system is present. |
| 6. Experience Design | N/A | The phase changes compiler/runtime behavior, not a user interface flow. |

**Overall: N/A** — the six-pillar visual score does not apply to this phase.

## Top 3 Priority Fixes

Not applicable. No UI findings or UI fixes are supported by the phase scope; priority fixes: 0.

## Detailed Findings

### Pillar 1: Copywriting (N/A)

There are no screen labels, calls to action, empty states, or interface error messages to review. Phase 36's declared work is scalar-call language and compiler behavior, while the public interpreter file CLI is explicitly owned by Phase 37 ([36-CONTEXT.md:10-12](./36-CONTEXT.md#L10)). No copywriting defect is inferred from compiler diagnostics or source-language syntax.

### Pillar 2: Visuals (N/A)

The plan inventories cover parser/checker and validation code, runtime execution, and native emission ([36-02-PLAN.md:7-14](./36-02-PLAN.md#L7), [36-03-PLAN.md:7-13](./36-03-PLAN.md#L7), [36-04-PLAN.md:7-11](./36-04-PLAN.md#L7), [36-05-PLAN.md:7-16](./36-05-PLAN.md#L7)). The implemented source paths inspected are compiler functions such as `scalarExprAtom` in `internal/compiler/syntax/parser.go:1085`, `lowerScalarCall` in `internal/compiler/check/scalar.go:937`, and dynamic call selection in `internal/compiler/cgen/cgen_program.go:2622`; these do not render UI components. No visual hierarchy or icon-button claim can be scored.

### Pillar 3: Color (N/A)

No CSS, Sass, HTML, or frontend configuration appeared in the repository scan, and the phase's file inventories contain no visual styling files ([36-01-PLAN.md:7-10](./36-01-PLAN.md#L7), [36-06-PLAN.md:7-15](./36-06-PLAN.md#L7), [36-07-PLAN.md:7-15](./36-07-PLAN.md#L7)). There is no UI palette or accent usage to count.

### Pillar 4: Typography (N/A)

No frontend text styles, font declarations, or typography tokens are in scope. The implementation inventory is compiler/runtime source, tests, Schway fixtures, and evidence files ([36-02-PLAN.md:7-14](./36-02-PLAN.md#L7), [36-04-PLAN.md:7-11](./36-04-PLAN.md#L7), [36-08-PLAN.md:7-12](./36-08-PLAN.md#L7)); there is no UI font-size or font-weight distribution to assess.

### Pillar 5: Spacing (N/A)

There is no interface layout tree or spacing scale in this phase. The source and evidence inventories identify parser/checker, validators, runtime/native paths, test fixtures, and acceptance records, without frontend layout files ([36-02-PLAN.md:7-14](./36-02-PLAN.md#L7), [36-05-PLAN.md:7-16](./36-05-PLAN.md#L7), [36-08-PLAN.md:7-12](./36-08-PLAN.md#L7)). UI spacing consistency cannot be evaluated here.

### Pillar 6: Experience Design (N/A)

The changed interactions are scalar helper-call semantics and runtime/native execution evidence. The phase context excludes the public interpreter file CLI to Phase 37, and the code paths inspected include interpreter call handling at `internal/compiler/interp/interp.go:1677` and C-emitter call handling at `internal/compiler/cgen/cgen_program.go:2622`. There is no screen flow with loading, empty, error, disabled, or destructive-action states to audit.

## Files Audited

- `AGENTS.md`
- `.planning/phases/36-scalar-helper-calls/36-CONTEXT.md`
- `.planning/phases/36-scalar-helper-calls/36-01-PLAN.md` through `36-08-PLAN.md`
- `.planning/phases/36-scalar-helper-calls/36-01-SUMMARY.md` through `36-08-SUMMARY.md`
- Applicability source paths: `internal/compiler/syntax/parser.go`, `internal/compiler/check/scalar.go`, `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go`
- Repository inventory for frontend extensions/configuration and phase UI-SPEC/UI-REVIEW artifacts
