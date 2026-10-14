---
phase: 35-bitwise-operators-literal-shifts-and-hex-literals
plan: "08"
subsystem: testing
tags: [literal-shifts, adler32, evidence-grounding, compatibility]
requires:
  - phase: 35-07
    provides: [checked literal shifts, public operator witnesses, frozen emitter corpus]
provides:
  - Adler consumer using one bounded b << 16 operation with unchanged pinned answers
  - Exact operator and public/application projection guidance
  - Source-bound compatibility and scoped local evidence receipts
affects: [phase-35-verification, phase-36-scalar-helper-calls, phase-37-crc-consumer, phase-38-hosted-receipts]
actuals:
  tokens: 16768
  tasks: 3
  commits: 4
plan_head_before: 6270ddeab8dd40cb91d641b31aea50bbf9c42da0
plan_head_after: 6cb2ca5c
commits: 4
tech-stack:
  added: []
  patterns: [focused literal-answer consumers, source and test grounded evidence receipts]
key-files:
  created:
    - internal/compiler/session/session_phase35_adler_test.go
    - internal/compiler/session/session_phase35_closeout_test.go
    - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md
    - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md
  modified:
    - examples/checksum/adler32.schway
    - docs/agent-surface.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md
key-decisions:
  - "Bounded b makes b << 16 equivalent to the prior sixteen doublings; keep the final addition and all pinned expected values unchanged."
  - "Describe fresh local runs, source inspection, historical receipts, and incomplete lanes separately; executionpeer earns causal admission credit only."
  - "Keep the two descriptor-less plan prohibitions flagged for independent phase verification without inventing descriptors or human approval."
patterns-established:
  - "Numeric evidence uses independent literal answers and reached wrong-result controls."
  - "Operator guidance distinguishes public outcome values from application stdout and names unsupported routes."
requirements-completed: [BIT-01, BIT-02, BIT-03, BIT-04]
coverage:
  - id: D1
    description: "The Adler example uses b << 16 and preserves checksum/adler_low answers and cleanup on both engines."
    requirement: BIT-04
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase35_adler_test.go#TestPhase35ExistingAdlerAnswers
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_adler_test.go#TestPhase35AdlerWrongShiftControl
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_phase29_interp_fixture_test.go#TestPhase29InterpreterFixtureVectors
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_phase29_native_fixture_test.go#TestPhase29NativeFixtureVectors
        status: pass
      - kind: integration
        ref: internal/compiler/native/native_phase31_matrix_test.go#TestPhase31CaseMatrix
        status: pass
    human_judgment: false
  - id: D2
    description: "The operator guide pins precedence, shift bounds, numeric spelling routes, Byte boundary, and public versus application output."
    requirement: BIT-01
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase35_closeout_test.go#TestPhase35OperatorDocumentation
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_closeout_test.go#TestPhase35PublicDocumentValues
        status: pass
      - kind: integration
        ref: cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_phase35_values_test.go#TestPhase35FocusedValueAnswers
        status: pass
    human_judgment: false
  - id: D3
    description: "Compatibility deltas and evidence receipts are tied to source, schema, semantic IDs, actual tests, and explicit evidence limits."
    requirement: BIT-02
    verification:
      - kind: unit
        ref: internal/compiler/session/session_phase35_closeout_test.go#TestPhase35EvidenceGrounding
        status: pass
      - kind: unit
        ref: internal/compiler/session/summary_coverage_test.go#TestSummaryCoverageGroundedness
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeEmitterCorpus
        status: pass
      - kind: unit
        ref: internal/compiler/session/session_phase35_prechange_test.go#TestPhase35PreChangeCorpusControls
        status: pass
    human_judgment: false
duration: 16min
completed: 2026-10-14
status: complete
---

# Phase 35 Plan 08: Adler Shift and Evidence Contract Summary

**Adler now computes its upper checksum half with one bounded literal shift while preserving existing answers, and the operator and evidence contracts are pinned to fresh, scoped receipts.**

## Performance

- **Duration:** 16 minutes
- **Started:** 2026-10-14T21:32:33Z (first task commit; execution began no later than this point)
- **Completed:** 2026-10-14T21:49:14Z
- **Tasks:** 3
- **Files modified:** 9 plan-owned files

## Accomplishments

- Replaced exactly sixteen Adler doublings with `b << 16`, retained `r = r + a`, and kept the existing literal answers unchanged. `checksum` and `adler_low` each passed seven pinned rows through interpreter and native consumers; the copied shift-15 control reached the comparator and returned `150143896` against expected `300286872`.
- Documented exact precedence and grouping, literal shift range 0..63, logical-right and low-64-bit left-shift behavior, hex spellings and route policy, depth 64, the unused-Byte boundary, and separate public `264` versus application `264\n` projections.
- Added compatibility and evidence ledgers grounded in the current source, old/current core schemas and semantic IDs, exact C projections, named test executions, and explicit historical and incomplete lanes. Updated the living roadmap and maturity record with observed facts and three ordered next capabilities; dormant seeds z1g, trz, and z1s remain dormant with source witnesses and triggers.
- Kept the 35-07 historical count accurate at 33 fixture rows and 16 UBSan shift rows. This plan did not run O0/O3/O3-LTO, UBSan, hosted checks, or Phase 36.

## Task Commits

1. **Task 1: Adler consumer baseline and implementation** — `cff86361` (test), `83c01058` (feat)
2. **Task 2: operator guide and living transition facts** — `26380801` (docs)
3. **Task 3: compatibility and evidence grounding** — `6cb2ca5c` (test)

## Files Created/Modified

- `examples/checksum/adler32.schway` — one literal shift replaces the repeated doublings.
- `internal/compiler/session/session_phase35_adler_test.go` — independent answer rows and reached wrong-shift control.
- `internal/compiler/session/session_phase35_closeout_test.go` — documentation and evidence-grounding controls.
- `docs/agent-surface.md` — exact operator and output-route contract.
- `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md` — dated, source-bound transition observations and capability ordering.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-COMPATIBILITY-DELTA.md` — scoped old/current source, schema, ID, and emitter comparison.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-EVIDENCE.md` — named fresh selectors, controls, historical receipts, and limits.
- `.planning/phases/35-bitwise-operators-literal-shifts-and-hex-literals/35-VALIDATION.md` — task rows marked passed only after the selectors ran.

## Validation

The focused selectors and final scoped union passed with `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1 -v`, and a 600-second timeout. The native case matrix had no skips. Task 3's exact selector was rerun at `6cb2ca5c`; its five false-claim controls were reached and rejected, the frozen emitter corpus and controls passed, and the live Adler projections matched the recorded source/core/ID/C fingerprints. The summary-groundedness selector is rerun after this summary is added so it checks the final 43-summary set.

## Decisions Made

- Keep independent numeric answers and the wrong-shift comparator as numeric evidence; causal event admission does not establish numeric results.
- Preserve the old compatibility baseline and record the Adler source rewrite as the narrow declared exception; no baseline fixture was recaptured or broadly exempted.
- Keep the unsupported application route and no-U64-input-flag boundary explicit; interpreter file CLI parity remains owned by Phase 37.

## Deviations from Plan

Task 3 also refreshed the two plan-owned living documents with the final test-receipt revision so their current-run claims remained bound to the checked baseline. No files outside the nine plan-owned paths were changed. The user's configuration edit was preserved.

## Issues Encountered

Markdown line wrapping required documentation assertions to compare normalized clauses. The application emitter does not support this library-shaped Adler entry, so the compatibility receipt pins the unchanged refusal and compares the two supported C-emitting routes. These checks pass without broadening the app route.

## Known Stubs

None found in the plan-owned files.

## Threat Flags

None. The plan adds no network endpoint, authority path, schema, or new file-access surface.

## Next Phase Readiness

Plan 35-08 is complete. The parent workflow still owns phase-wide verification, security review, and code review; the two descriptor-less prohibitions remain flagged for that independent verification. Phase 36 work was not started. The living documents identify the next capability owners and reprioritization observations without claiming human readability, cold/warm loop distributions, or fresh hosted evidence.

## Self-Check: PASSED

- All nine plan-owned files and this summary exist; the four task commits are ancestors of the current plan head.
- `TestPhase35EvidenceGrounding` and `TestSummaryCoverageGroundedness` passed after adding this summary; the coverage scan checked 43 summaries.

---
*Phase: 35-bitwise-operators-literal-shifts-and-hex-literals*
*Completed: 2026-10-14*
