---
phase: 18-branch-on-a-computed-value
plan: 10
subsystem: native-emission
tags: [c17, schema-2, terminal-outcome, payload, mutation]
requires:
  - phase: 18-05
    provides: computed terminal match and payload return path
  - phase: 18-09
    provides: four-tier session/program comparison and mutation evidence
provides:
  - Long accepted alternative names stream safely into schema-2 terminal outcomes
  - Long-tag wrong-slot mutation is detected at axis:terminal-outcome
affects: [phase-18-verification, schema-2-terminal-evidence]
actuals:
  tokens: 3200
  tasks: 2
  commits: 4
commits: 4
plan_head_before: 7231cc68377f0fdbcaaa94ee4942e14fe16542d1
tech-stack:
  added: []
  patterns: [bounded streaming JSON string-content writer]
key-files:
  created: [testdata/phase18/long_payload_return.lang]
  modified:
    - internal/compiler/cgen/cgen.go
    - internal/compiler/cgen/cgen_program.go
    - internal/compiler/session/session_phase18_payload_test.go
key-decisions:
  - "Stream the selected runtime tag and payload through the schema-2 output writer, retaining JSON escaping and output accounting."
  - "Declare only scratch buffers required by payload types present in the emitted branch."
patterns-established:
  - "Terminal tags use the schema-2 JSON string-content writer without tag-sized temporary storage."
requirements-completed: [CTL-03]
coverage:
  - id: D1
    description: "A 152-character Buffer alternative returns its exact tag and payload across interpreter, -O0, -O3, and -O3 -flto."
    requirement: CTL-03
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session -run '^TestPhase18LongPayloadPlaceReturn$' -count=1"
        status: pass
  - id: D2
    description: "Long-tag wrong-slot mutation remains observable as terminal-outcome disagreement, with ordinary payload and output-bound controls passing."
    requirement: CTL-03
    verification:
      - kind: integration
        ref: "go test ./internal/compiler/session ./internal/compiler/cgen -run '^(TestPhase18(LongPayloadPlaceReturn|LongTagWrongSlotMutation|PayloadPlaceReturn|WrongSlotMutation)|TestSchema2ExecutionOutputBoundIsPreflighted|TestProgramMatchPayloadLowering)$' -count=1"
        status: pass
duration: 8min
completed: 2026-09-26
status: complete
---

# Phase 18 Plan 10: Long Terminal Payload Summary

**Schema-2 terminal serialization now streams long alternative names and runtime payloads within the existing bounded output contract.**

## Performance

- **Duration:** 8 minutes
- **Started:** 2026-09-26T14:29:26Z
- **Completed:** 2026-09-26T14:34:25Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Added a checked source witness with an exactly 152-character `Buffer` alternative and verified its exact `name:01020304` outcome across four execution tiers.
- Extracted JSON string-content escaping into a schema-2 helper and streamed tag, colon, and runtime payload through `lang_write_bytes`, preserving output bounds and escaping.
- Added a long-tag wrong-slot mutation control that requires positive injection and exact `axis:terminal-outcome` disagreement; ordinary payload and output-bound controls pass.

## Task Commits

1. **Task 1 RED:** `a77acdc` (`test(18-10): add failing long payload return regression`)
2. **Task 1 GREEN:** `9995991` (`feat(18-10): stream long terminal payload values`)
3. **Task 1 Rule 1 correction:** `ec674d6` (`fix(18-10): declare only used payload scratch buffers`)
4. **Task 2:** `2113101` (`test(18-10): assert long-tag wrong-slot detection`)

## Files Created/Modified

- `testdata/phase18/long_payload_return.lang` - Long-name computed match and Buffer payload return witness.
- `internal/compiler/session/session_phase18_payload_test.go` - Four-tier long return and wrong-slot controls.
- `internal/compiler/cgen/cgen.go` - Schema-2 JSON content helper shared with the quote wrapper.
- `internal/compiler/cgen/cgen_program.go` - Bounded streaming of terminal tag and payload.

## Decisions Made

- Reused the schema-2 bounded byte writer for every terminal output byte; no tag-sized allocation or source name limit was introduced.
- Kept schema-1 support unchanged and limited the content-helper refactor to schema-2.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Blocking C17 compile issue] Moved terminal payload scratch declarations before switch labels**
- **Found during:** Task 1 GREEN verification
- **Issue:** C17 rejects a declaration directly after a `case` label under the installed Clang warning-as-error flags.
- **Fix:** Declared bounded payload scratch at function scope.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** Task 1 focused test passed.
- **Committed in:** `ec674d6`

**2. [Rule 1 - Bug] Avoided unused scratch variables for payload-free branches**
- **Found during:** Task 2 focused controls
- **Issue:** Unconditional Byte scratch declarations caused `-Werror=unused-variable` in generated programs without Byte payloads.
- **Fix:** Emit only the scratch declarations required by payload types in the branch.
- **Files modified:** `internal/compiler/cgen/cgen_program.go`
- **Verification:** Exact Task 2 focused command passed, including `TestProgramMatchPayloadLowering`.
- **Committed in:** `ec674d6`

**Total deviations:** 2 auto-fixed issues. Both were directly caused by the new writer; no scope was added.

## Issues Encountered

- The workspace sandbox initially denied GSD staging because Git could not write its index lock; the supported escalation enabled the requested GSD commits.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

CTL-03's long-name terminal serialization gap is closed. Existing Phase 18 UAT and validation evidence were not changed.

## Self-Check: PASSED

All four implementation artifacts and this summary exist; all four task commits are present in Git history.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-26*
