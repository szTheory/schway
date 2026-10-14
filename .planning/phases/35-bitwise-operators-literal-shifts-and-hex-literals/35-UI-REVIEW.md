# Phase 35 — UI Review

**Audited:** 2026-10-14  
**Baseline:** Applicability review; no phase UI-SPEC exists  
**Canonical UI gate:** `check ui.safety-gate 35 --raw` returned `frontend: false`, `hasUiFiles: false`, `hasUiSpec: false`, `block: false`  
**Screenshots:** Not captured — this phase has no graphical frontend  
**Interaction captures:** off (`workflow.ui_interaction_capture` is false)  
**Assessment:** Source-only CLI copy and documentation review; no numeric 6-pillar score

---

## Applicability

This phase adds compiler, checker, native-emission, CLI, and documentation behavior. The canonical gate found no frontend files or UI specification. Visuals, color, typography, spacing, and graphical experience design are therefore **N/A**. Copywriting receives a bounded qualitative review below because the CLI has user-facing diagnostics and the phase adds operator documentation. No browser, screenshot, accessibility, or human usability evidence was produced.

| Area | Applicability |
|---|---|
| Copywriting | Qualitative source-only review of CLI diagnostics and `docs/agent-surface.md`; no numeric score |
| Visuals | N/A — no graphical surface |
| Color | N/A — no graphical surface |
| Typography | N/A — no graphical surface |
| Spacing | N/A — no graphical surface |
| Experience Design | N/A for graphical interactions; CLI outcomes and refusal contracts are inspected as source |

**Overall score:** N/A. A `/24` score would imply visual evidence that this phase does not contain.

## Source-only findings

### CLI outcome copy and projection

`TestPhase35PublicConformanceWitness` describes the public `--json run` result as an execution document and checks `outcome.value` against `264` for both interpreter and native routes. It also uses the alternate grouping as a discriminating `8` result. The source requires a passing `run` command, one returned execution, no defect event, one terminal return event, no events attributed to pure operators, and no live resources ([`main_phase35_test.go:16`](../../../cmd/schway/main_phase35_test.go:16), [`main_phase35_test.go:54`](../../../cmd/schway/main_phase35_test.go:54)).

The application path is separately specified as stdout `264\n`: the documentation makes the distinction explicit ([`agent-surface.md:75`](../../../docs/agent-surface.md:75)), and the session test source checks that output on interpreter and native application routes ([`session_phase35_values_test.go:22`](../../../internal/compiler/session/session_phase35_values_test.go:22)). These are source assertions inspected during this review, not newly executed tests.

### Existing U64 input refusal

The guide says `run` still has no default U64 value or new input flag, and assigns interpreter file-input parity to Phase 37 ([`agent-surface.md:80`](../../../docs/agent-surface.md:80)). The existing CLI regression source pins exit 64, `tool.run_input_required`, identical JSON from both engines, and omission of the source path and parameter name ([`main_phase321_run_input_test.go:10`](../../../cmd/schway/main_phase321_run_input_test.go:10)). The implementation uses a static message: “run supplies no input for this entry's parameter; use build, then app run ARTIFACT -- INPUT” ([`session_run_input.go:17`](../../../internal/compiler/session/session_run_input.go:17)). The Phase 35 source inspected here does not itself rerun that refusal case. Phase 35-07's summary records the existing input-required selector as passing historically; this review did not rerun it.

### Shift refusal language

The refusal catalog registers `surface.unsupported_shift_amount` with the exact hint: “Shift amounts must be integer literals from 0 to 63; write a literal in that range.” ([`refusals.go:31`](../../../internal/compiler/surface/refusals.go:31)). The checker test source requires that exact message for names, computed expressions, negative syntax, and counts at or above 64; it also requires a single diagnostic and no repair suggestion. Overflow remains `type.invalid_u64`, while malformed spelling remains `syntax.malformed_numeric_literal` ([`check_phase35_test.go:64`](../../../internal/compiler/check/check_phase35_test.go:64), [`check_phase35_test.go:91`](../../../internal/compiler/check/check_phase35_test.go:91), [`check_phase35_test.go:139`](../../../internal/compiler/check/check_phase35_test.go:139)). This wording gives a concrete bound and repair action without suggesting unsupported dynamic behavior.

### Documentation readability

`docs/agent-surface.md` uses a clear progression from manifest scope and limitations to operator semantics, examples, shift constraints, numeric-route compatibility, and CLI output distinctions ([`agent-surface.md:3`](../../../docs/agent-surface.md:3), [`agent-surface.md:50`](../../../docs/agent-surface.md:50)). The locked expression and alternate result make precedence concrete. The guide also states what the manifest and causal execution peer do not prove. Source inspection found no ambiguous copy defect in these sections; no human-reader assessment was performed.

## Evidence boundaries

- **Newly executed:** The canonical UI safety gate only; it returned no frontend, UI files, or UI-SPEC.
- **Source inspected:** The Phase 35 public CLI test, checker test, output-seam test, input-required regression test and implementation, refusal catalog, and `docs/agent-surface.md`.
- **Historical receipts:** Phase 35-07 and 35-08 summaries record their named test selectors as passing. Those receipts were not recreated in this review.
- **Not assessed:** Screenshots, browser interactions, accessibility, visual styling, or human usability. No tests were run.

## Priority fixes

No graphical UI fixes apply to this compiler/CLI phase. The source-only copy and documentation review found no actionable wording change within the requested scope.

## Files Audited

- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-PLAN.md` through `35-08-PLAN.md`
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-01-SUMMARY.md` through `35-08-SUMMARY.md`, and `35-CONTEXT.md`
- `cmd/schway/main_phase35_test.go`
- `internal/compiler/check/check_phase35_test.go`
- `internal/compiler/session/session_phase35_values_test.go`
- `internal/compiler/session/session_phase35_closeout_test.go`
- `cmd/schway/main_phase321_run_input_test.go`
- `internal/compiler/session/session_run_input.go`
- `internal/compiler/surface/refusals.go`
- `docs/agent-surface.md`
