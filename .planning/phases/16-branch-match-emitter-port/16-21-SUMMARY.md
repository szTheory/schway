---
phase: 16-branch-match-emitter-port
plan: 21
subsystem: testing
tags: [go, compiler, frozen-evidence, provenance]
requires:
  - phase: 16
    plan: 16
    provides: Phase 11 N=1/N=2 frozen C artifacts bound to fixture, canonical-program, artifact, and refusal records
provides:
  - Test-only Phase 11 gate with exact refusal-first frozen evidence for N=1/N=2 by-pointer fixtures
  - Source-derived emitter inventory rows bound to Phase 11 fixture and provenance records
  - Mutation controls for fixture, program, artifact, refusal, classification, witness, and evidence mapping substitutions
affects: [phase-16-verification, M004-by-pointer-admission]
actuals:
  tokens: 14490
  tasks: 2
  commits: 4
tech-stack:
  added: []
  patterns:
    - Exact public refusal is checked before a test-only loader supplies provenance-validated historical C.
    - Frozen evidence consumers carry explicit fixture identities through inventory and mutation checks.
key-files:
  created: []
  modified:
    - internal/compiler/session/session_phase11_gate.go (deleted)
    - internal/compiler/session/session_phase11_gate_test.go
    - internal/compiler/session/session_phase16_emitter_inventory_test.go
    - internal/compiler/session/witness_registry_test.go
    - testdata/phase16/public-emitter-consumers.json
    - .planning/phases/16-branch-match-emitter-port/16-RESEARCH.md
key-decisions:
  - "Keep Phase 11 assurance test-only; by-pointer checks first require the exact cut-m004 refusal and then use frozen C for differential comparison."
  - "Keep zero-call entry ambiguity typed as callgraph entry ambiguity, without a by-pointer frozen-evidence mapping."
patterns-established:
  - "Map each by-pointer refusal consumer to exact committed fixture records, canonical-program digest, artifact digest, and refusal witness."
requirements-completed: [NAT-08, NAT-09]
coverage:
  - id: D1
    description: "Phase 11 N=1/N=2 refusal-first frozen evidence preserves the gate floors and four-tier behavior comparison."
    requirement: NAT-09
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestPhase11ZeroAttributeGate|TestPhase11GateIsNonVacuous|TestPhase11GateFailsAtNZero|TestPhase11GateCountsAdjacentWouldCarryFunctions|TestPhase11GateMutationKill|TestPhase11SuppressionIsDiffLocal)$' -count=1 -v"
        status: pass
    human_judgment: false
  - id: D2
    description: "Emitter inventory and frozen-evidence provenance reject fixture, canonical program, artifact, refusal, classification, witness, and mapping faults."
    requirement: NAT-08
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/session -run '^(TestPhase16PublicEmitterConsumerInventory|TestPhase16EmitterInventoryMutationControls|TestPhase16EmitterInventoryRefusalWitnessesResolve|TestPhase16M004ProvenanceRegistryRejectsFaults)$' -count=1 -v"
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-09-24
status: complete
plan_head_before: 24d6376f6bdaef8d0f1e26ee275d541b0108f3a6
commits: 4
---

# Phase 16 Plan 21: Refusal-First Phase 11 Evidence Summary

**Phase 11’s by-pointer assurance now proves the exact cut-m004 refusal before loading fixture-bound frozen C, while inventory and mutation controls validate the full provenance chain.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-09-24T00:43:38Z
- **Completed:** 2026-09-24T01:03:39Z
- **Tasks:** 2
- **Files modified:** 6 task files (one deleted)

## Accomplishments

- Moved the Phase 11 gate and independent structural derivation into test-only code and deleted the production helper.
- Bound N=1/N=2 by-pointer checks to exact named refusals before loading the corresponding frozen C artifacts; four-tier semantic comparison uses only those refusal-gated bytes.
- Preserved the live N=0 falsification case, structural floors, non-vacuity checks, would-carry counts, and gate lanes.
- Reworked seeded `restrict` mutation and diff-local controls to add one token at the function declaration seam in frozen, attribute-free C and require the banned-attribute scanner to detect it.
- Bound inventory rows to fixture, canonical program, artifact, exact refusal, classification, and witness records; retained zero-call ambiguity as a separate typed refusal.
- Resolved the research questions: Linux evidence is unnecessary for Phase 16’s no-admission cut and remains mandatory before any future by-pointer admission; no production by-pointer fence is enabled in Phase 16.

## Task Commits

1. **Task 1: Move Phase 11 gate verification to refusal-first, test-only evidence** - `1b5fa64` (test)
2. **Task 1: Record the declared production helper deletion** - `e191f5a` (refactor)
3. **Task 1: Clarify frozen evidence mutation controls** - `9526d2c` (test)
4. **Task 2: Bind Phase 11 inventory rows to frozen provenance** - `75336b1` (test)

**Plan metadata:** pending GSD close-out commit.

## Files Created/Modified

- `internal/compiler/session/session_phase11_gate.go` - Deleted production gate helper and lowering calls.
- `internal/compiler/session/session_phase11_gate_test.go` - Added test-only gate derivation and refusal-first frozen evidence use.
- `internal/compiler/session/session_phase16_emitter_inventory_test.go` - Added evidence-fixture row data and typed-refusal classification checks.
- `internal/compiler/session/witness_registry_test.go` - Validates canonical-program digests, exact refusals, registry mappings, and mutation faults.
- `testdata/phase16/public-emitter-consumers.json` - Refreshed the source-derived inventory and Phase 11 witnesses.
- `.planning/phases/16-branch-match-emitter-port/16-RESEARCH.md` - Resolved the Linux-evidence and production-fence questions.

## Decisions Made

- Keep Phase 11 assurance test-only and treat frozen artifacts solely as historical evidence after exact current refusal.
- Keep the ambiguous-entry refusal distinct from by-pointer frozen-evidence consumers.

## Deviations from Plan

None in implementation scope. GSD’s explicit `--files` commit skipped the deleted file, so its dedicated `--files-removed` option was used to record the declared deletion. Task 1 consequently has separate commits for the test migration, deletion, and wording/assertion cleanup.

## Issues Encountered

- Both plan-focused verification commands pass.
- Full `GOCACHE=/tmp/ai-lang-gocache go test ./...` was attempted on macOS and failed in `internal/compiler/session` with these findings:
  - `TestLanguageMaturityCountsAreCurrent`: `LANGUAGE-MATURITY.md:85` reports 133 programs/4,478 lines; re-derived values are 134/4,508.
  - `TestSelfDescribingDocsGuardIsNotInert/unmodified_maturity_doc_copy_passes` and `.../rewritten_re-verify_command_does_not_change_the_verdict`: the same stale corpus totals cause the copied-document controls to fail.
  - `TestValidationRowGradesAreEarnedOverArchivedCorpus`: checked-in validation corpus pair digest does not match the current requested corpus.
  - `TestCLICheckAdmissionDivergenceIsExactlyKnown`: `testdata/phase11/multi_function_gate_n_two.lang` is accepted by `session.Check+corevalidate.Validate` but refused by CLI `check` with `core.origin_omitted`.
  - `TestPayloadCorpusCharacterizationReplay/testdata/phase11/multi_function_gate_n_two.lang`: the fixture is missing from `payloadCorpusBaseline`; observed digest is `9e89e54a220c9e344fa82fe71c411426713941c31174b64c4138b008eab70f68`.
  - `TestLaneSchemaLiteralSiteCountIsPinned`: expected 18 lane-schema literals including the deleted production helper; 17 remain.
- The root executor will address fixture-index/validation/maturity refresh and the removed production-literal pin in a follow-on gap plan. `go test -race ./...` and Linux CI were not run in this macOS executor.
- Go’s default build cache was outside the writable workspace, so requested test runs used `GOCACHE=/tmp/ai-lang-gocache`.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

The Phase 11 gate and provenance inventory meet this plan’s focused acceptance tests. The full-suite findings above remain queued for the root executor’s follow-on gap plan before Phase 16 closeout.

## Self-Check: PASSED

- Summary file exists.
- Task commits `1b5fa64`, `e191f5a`, `9526d2c`, and `75336b1` exist.
- Plan-focused verification commands passed; full-suite failures are recorded above for follow-on work.

---
*Phase: 16-branch-match-emitter-port*
*Completed: 2026-09-24*
