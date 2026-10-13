---
phase: 34-agent-diagnostics-floor-and-surface-manifest
plan: "04"
subsystem: compiler
tags: [go, parser, diagnostics, recovery]

requires:
  - phase: 34-02
    provides: Phase 34 diagnostic spans and parser recovery foundations
provides:
  - Bounded delimiter-owned parser recovery that preserves caller closers and declaration siblings
  - Exact parser-token ownership for suppressing only matching generic lexer diagnostics
  - Exact current parser diagnostics, spans, cause extents, and sibling names for all twenty reconstructed drafts
affects: [34-05 parser refusals, 34-08 current corpus verification]

actuals:
  tokens: 27456.5
  tasks: 2
  commits: 5
plan_head_before: 4e29b7d0892becc25a7c3ca6f23a881303e88909
plan_head_after: 679893ffb3710eeb1cb5e077bbaeeac7d7f72abb

tech-stack:
  added: []
  patterns:
    - Per-parse exact token-span ownership for lexer suppression
    - Iterative delimiter-stack synchronization bounded by the existing token budget
key-files:
  created:
    - internal/compiler/syntax/parser_phase34_recovery_test.go
    - testdata/phase34/draft-after.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task1-tdd-red-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task1-tdd-green-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task1-tracer-final-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task2-tdd-red-evidence.json
    - .planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task2-tdd-green-evidence.json
  modified:
    - internal/compiler/syntax/parser.go

key-decisions:
  - "Recovery owns exact parser spans and caller boundaries; unrelated lexer diagnostics and independent missing closers remain visible."
  - "Keep the immutable before-state unchanged and record only measured current outputs in the separate after-state."

patterns-established:
  - "Delimiter recovery: consume matching nested pairs, stop at caller-owned closers or valid declaration siblings, and preserve the exact skipped byte extent."
  - "Test controls: mutate local observations to verify the corpus comparator's named cascade, sibling, and independent-cause failure classes."

requirements-completed: [DX-21]
coverage:
  - id: D1
    description: "Parser recovery reports exact causes across the twenty reconstructed drafts while retaining bytes, siblings, independent closers, and skipped-region extents."
    requirement: DX-21
    verification:
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryCorpus
        status: pass
      - kind: integration
        ref: internal/compiler/syntax/parser_skipped_region_test.go#TestSkippedRegionCauseSeparatesSpiralTrio
        status: pass
    human_judgment: false
  - id: D2
    description: "Deep malformed delimiter input terminates within existing bounds, retains the named error-cap marker, and reaches/restores all three recovery controls."
    requirement: DX-21
    verification:
      - kind: unit
        ref: internal/compiler/syntax/parser_phase34_recovery_test.go#TestPhase34RecoveryControls
        status: pass
      - kind: unit
        ref: internal/compiler/syntax/parser_skipped_region_test.go#TestValidProgramUnaffectedBySkippedRegionFix
        status: pass
    human_judgment: false

commits: 5
duration: 16min
completed: 2026-10-13
status: complete
---

# Phase 34 Plan 04: Agent Diagnostics Floor and Surface Manifest Summary

**Delimiter-owned recovery now reports parser causes at their exact spans, preserves siblings and lossless bytes, and records current outputs for all twenty reconstructed drafts.**

## Performance

- **Duration:** 16 min from the first recorded focused test run.
- **Started:** 2026-10-13T22:32:51Z
- **Completed:** 2026-10-13T22:48:47Z
- **Tasks:** 2
- **Files modified:** 13 plan output and evidence files.

## Accomplishments

- Added iterative delimiter-stack synchronization bounded by the parser's existing token budget. Recovery consumes matching nested pairs, preserves unmatched caller closers and valid declaration siblings, and retains exact skipped-region byte extents.
- Scoped diagnostic ownership to each `Parse`: only the generic `syntax.unexpected_byte` with the exact span claimed by a parser primary is removed. Independent lexer bytes, missing closers, and separately authored mistakes remain visible.
- Recorded exact current counts, codes, spans, cause extents, and recovered function names for all twenty reconstructed drafts in `draft-after.json`; the immutable baseline remains unchanged.
- Verified the rejected nested-paren draft preserves `Tree.Bytes` and `session.Check` returns diagnostics with an empty core program. The independent missing-closer fixture retains `expected_rparen@63:65` and `expected_rbrace@74:74`.
- Ran the named focused GREEN selector at the committed implementation: exit 0 in 1,030.8 ms, with six passing tests covering the corpus, controls, 4,096-level delimiter input, truncation marker, spiral identities/extents, and a valid parse.

## Task Commits

1. **Task 1 RED:** `423e9c0a` — pin one-cause sibling recovery.
2. **Task 1 implementation:** `228bb3b0` — suppress only the exact duplicate lexer report.
3. **Task 1 tracer:** `e6c43323` — record the post-GREEN tracer pass.
4. **Task 2 RED:** `644c7ada` — pin nested delimiter recovery cause.
5. **Task 2 implementation:** `679893ff` — recover locally at owned delimiters and record GREEN evidence.

The plan's measured commit count is five, from `4e29b7d0892becc25a7c3ca6f23a881303e88909` through `679893ffb3710eeb1cb5e077bbaeeac7d7f72abb` at summary creation.

## Files Created/Modified

- `internal/compiler/syntax/parser.go` — exact token and boundary ownership, delimiter-aware synchronization, and local recovery for malformed arguments and linear bodies.
- `internal/compiler/syntax/parser_phase34_recovery_test.go` — twenty-draft comparator, parse-gate check, explicit missing-closer assertion, recovery controls, nested-depth and cap checks.
- `testdata/phase34/draft-after.json` — measured current after-state with intentional-delta notes; no historical scratch counts are claimed.
- `.planning/phases/34-agent-diagnostics-floor-and-surface-manifest/34-04-task*-evidence.*` — raw Go JSON, JUnit, freshness metadata, reporter source, and inspected semantic assessments for RED, GREEN, and tracer runs.

`lexer.go` needed no change: its existing unknown-token spans provide the exact byte identity used by parser-side suppression.

## Decisions Made

- The malformed nested call's owning cause is `syntax.expected_call_argument@94:95`, where the parser first rejects the nested argument. The later closer and result errors were cascades from that point.
- Keep recovery controls in the external test as local observation mutations. They exercise the corpus comparator's named failure classifications without adding test-only production API or changing parser behavior.

## Deviations from Plan

None — the plan's recovery behavior, TDD gates, and focused verification were completed. The RED assertion was refined after source inspection to pin the actual parser-owned root at `94:95` before GREEN.

## Issues Encountered

The first RED draft expected the later `expected_linear_result@96:97` row to own the malformed call. Inspection showed the existing ownership-call path encounters the invalid nested argument first at `94:95`; the test now pins that root cause while retaining the one-cause requirement. No implementation scope changed.

## TDD Evidence

- Task 1 RED: `TestPhase34RecoveryCorpus` failed at the intended duplicate-diagnostic count assertion; `RED_EVIDENCE_OK`. Task 1 GREEN passed the same focused selector in 976.2 ms; the tracer rerun passed in 185.1 ms.
- Task 2 RED: `TestPhase34RecoveryCorpus` failed at the intended one-cause count assertion in 659.4 ms; `RED_EVIDENCE_OK`.
- Task 2 GREEN: the focused selector `^(TestPhase34RecoveryCorpus|TestPhase34RecoveryControls|TestSkippedRegionCauseSeparatesSpiralTrio|TestValidProgramUnaffectedBySkippedRegionFix)$` passed six named tests, exit 0, in 1,030.8 ms. The control logs show `cascade_count`, `missing_sibling`, and `missing_independent_cause` were each reached and normal output restored afterward.
- All Go calls used `GOCACHE=/tmp/ai-lang-verification-gocache` and focused selectors. No project-wide or native suite was run.

## Known Scope Limit

The authored corpus pins these reconstructed cases and recovery controls; it does not prove an exhaustive guarantee for every arbitrary malformed program. The plan's unresolved spec-less edge-probe assumption remains scoped for later planning rather than being claimed as complete here.

## User Setup Required

None — no external service configuration is required.

## Next Plan Readiness

Plan 34-05 can add its dedicated refusal diagnostics and update only its authorized after-state rows. This plan leaves those refusal codes unintroduced and preserves the immutable before-state.

## Self-Check: PASSED

The summary, parser, recovery test, and after-state files exist, and all five task commits are ancestors of the current HEAD. Required frontmatter markers for completion status, measured actuals, DX-21, and the plan-head ledger are present.

---
*Phase: 34-agent-diagnostics-floor-and-surface-manifest*
*Completed: 2026-10-13*
