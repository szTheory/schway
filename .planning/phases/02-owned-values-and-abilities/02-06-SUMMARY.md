---
phase: 02-owned-values-and-abilities
plan: "06"
subsystem: compiler
tags: [go, evidence, content-binding, verification, warm-distributions, tdd]

requires:
  - phase: 02-owned-values-and-abilities
    plan: "05"
    provides: Strict bounded owned native executions and full interpreter/O0/O3 semantic equality
provides:
  - Versioned owned evidence binding independently validated core, readable C17, and ordered execution content
  - Concrete diagnostic/core/execution schemas with Phase 1 evidence byte preservation
  - Seven exact Phase 2 negative controls with nonzero work and a separately named expected escape
  - One non-duplicating Phase 1+2 gate with 20-sample warm timing distributions
affects: [phase-2-verification, phase-3-planning, evidence, feedback-latency]

actuals:
  tokens: 9209
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns: [feature-specific evidence schemas, independently admitted content binding, expected-escape separation, single-pass phase gate, warm distribution reporting]

key-files:
  created:
    - testdata/phase2/evidence.golden.json
    - scripts/verify-phase2.sh
  modified:
    - internal/compiler/evidence/evidence.go
    - internal/compiler/evidence/evidence_test.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go

key-decisions:
  - "Phase 1 evidence remains byte-identical on lang.evidence/0; owned content selects lang.evidence/1 with concrete diagnostic/core/execution schema identities."
  - "SHA-256 is labeled content-identity-only, and escape:coordinated-source-core-lie remains an expected limitation rather than a detected mutation."
  - "The Phase 2 gate owns shared Go test, race, and vet work exactly once and measures shipped commands through one prebuilt CLI."

patterns-established:
  - "Evidence admission: independently validate and content-own typed core before hashing core, C, and ordered execution facts."
  - "Bounded phase verification: exact controls, nonzero work, expected escapes, and warm distributions share one offline command."

requirements-progressed: [OWN-01, OWN-02]

coverage:
  - id: D1
    description: "Owned evidence binds canonical source, independently validated core, C17, concrete schemas, tool/target/flags/policy, and ordered execution facts without claiming translation, signing, or freshness authority."
    requirement: OWN-01
    verification:
      - kind: integration
        ref: "internal/compiler/evidence/evidence_test.go#TestOwnedEvidenceBindings,TestOwnedEvidenceMutationMatrix,TestPhase1EvidenceGoldenUnchanged"
        status: pass
      - kind: integration
        ref: "internal/compiler/protocol/protocol_test.go#TestOwnershipProjectionIdentityParity"
        status: pass
    human_judgment: false
  - id: D2
    description: "The bounded Phase 1+2 gate observes all seven exact controls with nonzero work, separates the expected escape, avoids nested shared suites, and reports five 20-sample warm distributions."
    requirement: OWN-02
    verification:
      - kind: e2e
        ref: "sh scripts/verify-phase2.sh"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_test.go#TestVerifyPhase2ControlsAndWork"
        status: pass
      - kind: integration
        ref: "internal/compiler/testsupport/cli_test.go#TestVerifyPhase2CLI,TestPhase2VerifierScriptContract"
        status: pass
    human_judgment: false

duration: 10min
completed: 2026-09-03
status: complete
---

# Phase 02 Plan 06: Owned Evidence and Phase Gate Summary

**Independently admitted owned core, C17, and ordered executions now produce honest content-only evidence, reproduced by one bounded Phase 1+2 gate with exact controls and warm distributions.**

## Performance

- **Duration:** 10 min
- **Started:** 2026-09-03T21:33:48Z
- **Completed:** 2026-09-03T21:44:01Z
- **Tasks:** 2
- **Files modified:** 9

## Accomplishments

- Added `lang.evidence/1` for owned programs with concrete `/1` diagnostic/core/execution schemas, canonical source/core/C digests, ordered execution digests, and content-owned slices after independent core validation.
- Preserved the Phase 1 evidence golden byte-for-byte on `lang.evidence/0` and kept command projection compatibility on `lang.command/0`.
- Added one offline Phase 2 script that self-tests exact Go target discovery, runs shared test/race/vet work once, verifies both shipped corpora, and fails closed on seven named controls or zero-work lanes.
- Reported 20 warm observations for format, check, interpreter, native, and full verify with p50/p95/min/max, output bytes, work, and honest unavailable RSS.

## Task Commits

Each task was committed with a TDD RED contract followed by its GREEN implementation:

1. **Task 02-06-01 RED: owned evidence contracts** - `2031e75` (test)
2. **Task 02-06-01 GREEN: validated owned content binding** - `a5d9710` (feat)
3. **Task 02-06-02 RED: bounded phase-gate contracts** - `ccb8e5e` (test)
4. **Task 02-06-02 GREEN: Phase 1+2 verification gate** - `ccf36b1` (feat)

## Files Created/Modified

- `internal/compiler/evidence/evidence.go` - Feature-specific evidence schemas, independent core admission, ordered execution binding, strict mutation codes, and content-only assurance metadata.
- `internal/compiler/evidence/evidence_test.go` - Owned bindings, one-boundary mutations, event reorder, alias ownership, strict JSON, and Phase 1 golden preservation.
- `internal/compiler/protocol/protocol.go` - Ownership-aware human event projection and explicit expected-escape records without changing command `/0`.
- `internal/compiler/protocol/protocol_test.go` - Human/JSON ownership identity parity.
- `internal/compiler/session/session.go` - Owned evidence summaries, Phase 2 control lanes, expected-escape separation, nonzero work, and opt-in timing observations.
- `internal/compiler/session/session_test.go` - Phase 2 gate control/work contract.
- `internal/compiler/testsupport/cli_test.go` - Shipped CLI and nonduplicating script contracts.
- `testdata/phase2/evidence.golden.json` - Deterministic reviewable owned evidence fixture.
- `scripts/verify-phase2.sh` - Single bounded offline phase gate and five warm distributions.

## Decisions Made

- Evidence schema selection follows the admitted core feature: legacy match artifacts remain `/0`; owned linear artifacts select `/1` and state their diagnostic/core/execution schemas explicitly.
- Execution order is identity-bearing through canonical per-execution digests; event reorder is rejected as `evidence.execution_mismatch`.
- `content-identity-only` is a machine-readable claim boundary. The coordinated source/core lie remains `escape:coordinated-source-core-lie` and never appears in detected controls.
- Timing observations are opt-in operational metrics excluded from semantic identity; the ordinary deterministic command projection remains stable.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The first expected-escape human projection patch landed in result finalization rather than rendering; the compile failure exposed it immediately and the loop was moved to the human renderer before verification.

## Verification Evidence

- Exact Plan 02-06 evidence selector passed all five requested targets, including the Phase 1 golden and named coordinated escape.
- `env GOCACHE=/tmp/ai-lang-phase2-cache go test ./...` passed across every package.
- `sh scripts/verify-phase2.sh` passed `go test ./...`, `go test -race ./...`, and `go vet ./...` exactly once, then passed shipped verification for both Phase 1 and Phase 2.
- The Phase 2 result recorded all seven required controls, four nonzero-work lanes totaling 49 work units, and the separate expected escape.
- Warm observations (20 each): format p50/p95 `49,125/153,667 ns`; check `63,250/94,333 ns`; interpreter `185,542/200,334 ns`; native `381,998,583/456,140,959 ns`; full verify `427,166,125/542,596,959 ns`. RSS remained honestly unavailable; these observations do not ratify an SLO.

## Known Stubs

None.

## Threat Flags

None. Evidence admission/content claims and the bounded local verification gate implement the planned T-02-08 and T-02-01..07 mitigations without introducing a new trust boundary.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- All six Phase 2 plans are executed and ready for independent goal-backward phase verification.
- `OWN-01` and `OWN-02` remain pending until the independent verifier/orchestrator accepts the phase; no requirement-completion claim is made here.
- No executor blocker remains.

## Self-Check: PASSED

All nine changed implementation/test/golden/script files and all four RED/GREEN commits were found. The full Phase 2 gate passed, this summary uses `requirements-progressed: [OWN-01, OWN-02]`, and neither `REQUIREMENTS.md` nor `.planning/milestone.lock` was modified.

---
*Phase: 02-owned-values-and-abilities*
*Completed: 2026-09-03*
