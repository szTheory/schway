---
phase: 18-branch-on-a-computed-value
plan: 05
subsystem: compiler
tags: [computed-match, payloads, interpreter, native-c, differential-testing, mutation-testing]
requires:
  - phase: 18-03
    provides: Independent admission for computed scrutinees and payload provenance.
  - phase: 18-04
    provides: Typed computed-match results across interpreter and native execution.
provides:
  - Runtime payload values serialized in interpreter and native terminal outcomes.
  - Source-driven Phase 18 wrong-slot mutation evidence on the terminal-outcome axis.
affects: [phase-18-verification, CTL-03, payload-observability]
actuals:
  tokens: 5625
  tasks: 3
  commits: 0
tech-stack:
  added: []
  patterns: ["tag-plus-runtime-payload terminal serialization", "data-parameter input seeding aligned with native fixtures"]
key-files:
  created:
    - internal/compiler/session/session_phase18_payload_test.go
  modified:
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/interp/interp.go
    - internal/compiler/session/session_payload_control_test.go
    - internal/compiler/session/session_phase18_test.go
    - internal/compiler/session/witness_registry_test.go
    - testdata/phase16/public-emitter-consumers.json
    - testdata/phase18/payload_return.lang
    - .planning/LANGUAGE-MATURITY.md
key-decisions:
  - "Serialize payload-bearing terminal ADTs as the selected alternative plus their runtime payload, while retaining tag-only String() for branch dispatch."
  - "Seed interpreter data inputs with the same canonical Buffer, Byte, and nullary nested payload values used by native entry setup."
  - "Keep CTL-03 open until full Phase 18 verification, despite this plan's passing payload witness."
patterns-established:
  - "Mutation controls assert both a positive injected-write count and the exact comparator axis."
  - "Initialize constructed native ADT records so wrong-slot mutation outcomes are deterministic."
requirements-completed: []
coverage:
  - id: D1
    description: "A source-driven computed match carries a destructured Buffer payload into the returned Outcome value, serialized as Ok:01020304."
    requirement: CTL-03
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase18PayloadPlaceReturn' -count=1"
        status: pass
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase18' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The seeded wrong-slot write injects at least once and disagrees exactly on axis:terminal-outcome; an unmutated four-tier companion agrees."
    requirement: CTL-03
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run 'TestPhase18(PayloadWrongSlot|WrongSlotMutation)' -count=1"
        status: pass
      - kind: integration
        ref: "GOCACHE=/tmp/ai-lang-gocache go test ./..."
        status: pass
    human_judgment: false
metrics:
  duration: 24min
  completed_date: 2026-09-24
status: complete
plan_head_before: b808fc858c8d8fbb622c4dc1b00c66c15da2a0f9
commits: 0
---

# Phase 18 Plan 05: Payload Return Evidence Summary

**Computed-match terminal records now expose the runtime payload value, and the seeded wrong-slot mutation is caught at the terminal-outcome axis.**

## Performance

- **Duration:** approximately 24 minutes
- **Started:** 2026-09-24T17:43:00Z (approximate; the executor did not capture a monotonic start timestamp)
- **Completed:** 2026-09-24T18:07:17Z
- **Tasks:** 3 of 3
- **Files modified:** 8 tracked files and 1 created test file (excluding the pre-existing config edit)

## Accomplishments

- Extended interpreter input values to retain a selected data alternative and the same canonical runtime payload used by native entry setup. Branch dispatch remains tag-based.
- Added schema-2 native terminal serialization for returned payload-bearing ADTs. Buffer payloads serialize as lowercase hex; Byte and nested nullary data payloads preserve their runtime values.
- Expanded `payload_return.lang` to exercise a computed local, payload destructuring, reconstruction from the selected payload place, and runtime payload observation. All four execution tiers agree on `Ok:01020304` when unmutated.
- Strengthened the established mutation control and the D-12-43 witness to require positive injection and a disagreement exactly at `axis:terminal-outcome`.

## Task Commits

Task commits were deferred to the parent GSD commit helper as requested. The implementation was left uncommitted at handoff; current HEAD remains `b808fc8`.

## Files Created/Modified

- `internal/compiler/session/session_phase18_payload_test.go` — four-tier agreement, positive-injection, and exact-axis source-driven controls.
- `internal/compiler/interp/interp.go` — canonical data input initialization and returned tag-plus-payload serialization.
- `internal/compiler/cgen/cgen_program.go` — runtime terminal serialization and zero-initialized payload construction.
- `internal/compiler/session/session_phase18_test.go` — updated fixture-frontier assertions.
- `internal/compiler/session/session_payload_control_test.go` — now requires the existing wrong-slot control to kill at terminal outcome.
- `internal/compiler/session/witness_registry_test.go` — updates the established D-12-43 probe to assert constructibility and exact-axis divergence.
- `testdata/phase18/payload_return.lang` — computed payload return witness with distinct payload types.
- `testdata/phase16/public-emitter-consumers.json` — refreshed the shifted emitter call location.
- `.planning/LANGUAGE-MATURITY.md` — corrected the machine-checked corpus line count to 4,577.

## Decisions Made

- Keep `value.String()` tag-only for branch selection; a distinct terminal projection includes both the selected alternative and its runtime payload.
- Preserve D-12-43's existing probe identity for the archived evidence index while changing its assertion to require the now-constructible divergence.
- Keep phase requirements open until the complete Phase 18 verification pass.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing runtime evidence] Added runtime payload terminal serialization and aligned interpreter inputs.**
- **Found during:** Task 1
- **Issue:** Both engines previously serialized only the alternative tag, and interpreter data inputs did not carry the canonical payload that native `main` initializes.
- **Fix:** Added a terminal-only interpreter projection, canonical payload seeding for data inputs, and a native schema-2 terminal writer that reads the returned struct's selected payload slot.
- **Files modified:** `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen_program.go`
- **Verification:** Four-tier Phase 18 session coverage and existing payload tracer controls passed.

**2. [Rule 2 - Deterministic mutation outcome] Zero-initialized constructed native ADTs.**
- **Found during:** Task 3
- **Issue:** The seeded wrong-slot write left the selected payload field indeterminate, making the terminal read undefined and the mutation result nondeterministic.
- **Fix:** Initialize constructed ADT storage to zero before writing its selected tag and payload slot.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** The seeded write produces a stable terminal-outcome mismatch; the unmutated four-tier source route agrees.

**3. [Rule 3 - Blocking test integration] Updated dependent evidence tripwires and derived counts.**
- **Found during:** Overall verification
- **Issue:** Existing controls asserted that D-12-43 remained unconstructible, the corpus count was stale after fixture edits, and the emitter inventory referenced the pre-edit source line.
- **Fix:** Updated the established D-12-43 probe, refreshed the checked corpus line count and emitter call inventory.
- **Files modified:** `internal/compiler/session/session_payload_control_test.go`, `internal/compiler/session/witness_registry_test.go`, `testdata/phase16/public-emitter-consumers.json`, `.planning/LANGUAGE-MATURITY.md`
- **Verification:** Full `GOCACHE=/tmp/ai-lang-gocache go test ./...` passed.

**Total deviations:** 3 auto-fixed (2 Rule 2, 1 Rule 3). **Impact:** These changes make runtime payload comparison deterministic and update controls that pinned the previously open gap.

## Verification

- Passed `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/session -run 'TestPhase18' -count=1`.
- Passed the Phase 18 payload mutation, established payload mutation, D-12-43, maturity-count, corpus-digest, and emitter-inventory focused controls.
- Passed existing native payload compatibility controls: `GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cgen -run 'TestPayloadTracerThreeEngineAgreement|TestProgramMatchPayloadLowering' -count=1`.
- Passed the complete suite: `GOCACHE=/tmp/ai-lang-gocache go test ./...`.
- The mutation assertion checks injection count greater than zero and comparator axis exactly `axis:terminal-outcome`; the separate unmutated case agrees across interpreter, `-O0`, `-O3`, and `-O3 -flto`.

## Issues Encountered

- The first candidate fixture used two `Buffer` alternatives, which the checker correctly refuses as an ambiguous duplicate payload type. The final fixture uses `Ok(Buffer)` and `Again(Byte)`.
- The generic session schema-2 projection already preserves `Outcome.Value`; no change to `session.go` was needed.

## Next Phase Readiness

Plan 18-05's payload value and mutation evidence are complete. CTL-01 through CTL-03 remain open until full Phase 18 verification, as directed.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-24*

## Self-Check: PASSED

- The summary and source witness exist.
- The task was intentionally left uncommitted for the parent GSD helper; no commit hash is claimed here.
