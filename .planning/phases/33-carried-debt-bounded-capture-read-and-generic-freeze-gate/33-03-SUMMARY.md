---
phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate
plan: 03
subsystem: testing
tags: [go, posix-sh, codedigest, receipts, freeze-gate, differential-tests]
requires:
  - phase: 33-01
    provides: reopened M006 receipt boundary and bounded capture baseline
provides:
  - A generic M006-manifest-driven freeze gate and receipt gate
  - One canonical digest path list with revision and changed-path modes
  - Differential evidence for legacy M006 gate behavior and archived receipt controls
affects: [phase-33, phase-38, evidence-gates, codedigest]
actuals:
  tokens: 13411
  tasks: 3
  commits: 8
plan_head_before: 0f83a781e271442f652ff0d575c66192a1e721da
plan_head_after: ff24ec1695ea9a3af4eee008f2c81b51cd0d0a44
tech-stack:
  added: []
  patterns:
    - Keep generic POSIX sh gates beside their phase-specific predecessors until differential outputs are frozen
    - Read receipt label, program list, controls, and case rows from a small manifest
    - Resolve repository-root paths before changing cwd while retaining caller-typed log paths in refusal reasons
key-files:
  created:
    - scripts/check-freeze.sh
    - scripts/check-receipts.sh
    - testdata/receipts/M006.manifest
    - internal/compiler/session/session_freeze_gate_test.go
    - internal/compiler/session/session_receipt_gate_test.go
    - internal/compiler/session/session_gate_differential_test.go
  modified:
    - scripts/codedigest.sh
key-decisions:
  - "Keep each generic gate's manifest reader local to its script, using POSIX sh and no shared runtime helper."
  - "Use EXPECTED_CODEDIGEST when supplied and report whether the gate used env or computed input."
  - "Derive the digest-width check from the selected hasher's empty-input output."
patterns-established:
  - "Gate refusal output is a single stable stderr line whose prefix comes from manifest data."
  - "Differential tests run old and generic gates over the same logs and temporary Git history."
requirements-completed: [EVD-16, QLT-14]
coverage:
  - id: D1
    description: The digest and generic freeze gate share repository-root paths, ignore Markdown-only changes, and reject changed or reverted code paths.
    requirement: QLT-14
    verification:
      - kind: integration
        ref: internal/compiler/session/session_freeze_gate_test.go#TestFreezeGate
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_gate_differential_test.go#TestFreezeGateDifferential
        status: pass
    human_judgment: false
  - id: D2
    description: The generic receipt gate matches the old M006 refusal catalog, derives label and controls from the manifest, and accepts archived M006 receipts with their recorded digest.
    requirement: EVD-16
    verification:
      - kind: integration
        ref: internal/compiler/session/session_receipt_gate_test.go#TestReceiptGateDifferential
        status: pass
      - kind: integration
        ref: internal/compiler/session/session_receipt_gate_test.go#TestReceiptGateArchivedM006
        status: pass
    human_judgment: false
duration: 20min
completed: 2026-10-13
status: complete
---

# Phase 33 Plan 03: Generic Freeze and Receipt Gates Summary

**The M006 evidence gates now use reusable POSIX shell scripts and a manifest, with digest path handling centralized in `codedigest.sh`.**

## Performance

- **Duration:** 20 min
- **Started:** 2026-10-13T17:05:00Z (recorded estimate; task start was not captured at launch)
- **Completed:** 2026-10-13T17:25:05Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments

- Extended `codedigest.sh` with `--paths`, `--rev`, and `--changed`; `.md` filtering and the 8-path set now have one source, and hash width is derived from the chosen hasher.
- Added the 39-row M006 manifest and root-aware generic freeze and receipt gates. The old phase-specific scripts remain available for the next plan's differential golden work.
- Added freeze and receipt parity tests, a second-label receipt fixture, archived macOS/Linux receipt acceptance with six armed controls, and explicit root, Markdown-only, revert, and relative-path regressions.

## Task Commits

1. **Task 1: Freeze slice and differential controls** — `aeac26a0` (`feat(33-03): add generic freeze gate and digest path contract`)
2. **Task 2: Manifest-driven receipt gate and differential catalog** — `2d2957c7` (`feat(33-03): add manifest-driven receipt gate`)
3. **Task 3: Archived M006 receipt controls and producer parity** — `77251dda` (`test(33-03): arm archived M006 receipt controls`)

Additional plan-acceptance follow-up commits:

- `fb3449ac` — explicit Markdown-only and reverted-range freeze controls.
- `1d21f4aa` — preserved receipt gate first-failure ordering and pinned manifest-prefixed usage.
- `7ac5bfcb` — exercised the real M006 freeze check from `cmd/` against project HEAD.
- `dc947061` — pinned label-less manifest, no-repository, and computed-digest paths.
- `ff24ec16` — compared digest width to the active hasher and covered non-ASCII Markdown filtering.

**Measured plan metadata:** 8 commits from `0f83a781e271442f652ff0d575c66192a1e721da` through `ff24ec1695ea9a3af4eee008f2c81b51cd0d0a44`.

## Files Created/Modified

- `scripts/codedigest.sh` — canonical digest paths, revision selection, changed-path reporting, and derived hash width.
- `testdata/receipts/M006.manifest` — label, programs, controls, and 39 case rows copied from the producer.
- `scripts/check-freeze.sh` — generic freeze consistency gate with repository-root and manifest handling.
- `scripts/check-receipts.sh` — generic two-host receipt gate with manifest-derived fields and optional caller-supplied digest.
- `internal/compiler/session/session_freeze_gate_test.go` — freeze behavior, root handling, M006 history, Markdown-only, and revert controls.
- `internal/compiler/session/session_receipt_gate_test.go` — generic receipt, archived receipt, second-label, and refusal tests.
- `internal/compiler/session/session_gate_differential_test.go` — old/new freeze gate parity matrix.

## Decisions Made

- The new scripts each carry their own manifest reader, as required by the generic-gate contract; neither script sources a shared helper.
- The manifest label is read before arity checks so usage errors retain the label prefix. Digest resolution precedes manifest-content checks, preserving the specified refusal order.
- Archived acceptance takes its expected digest from `32-RECEIPT.json` and uses temporary copies for mutations; archived files remain unchanged.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Preserved receipt refusal precedence**
- **Found during:** Task 2
- **Issue:** The first implementation checked manifest contents before the expected digest, which could change the first reported refusal when multiple inputs were invalid.
- **Fix:** Kept label loading before arity validation, then moved digest resolution ahead of manifest-content checks and added a byte-exact prefixed-usage regression.
- **Files modified:** `scripts/check-receipts.sh`, `internal/compiler/session/session_receipt_gate_test.go`
- **Verification:** The receipt differential catalog and focused manifest-usage test passed.
- **Committed in:** `1d21f4aa`

**Total deviations:** 1 auto-fixed bug. Explicit acceptance tests were added in follow-up commits where the first test pass did not yet cover every named plan case.
**Impact on plan:** No scope change; the follow-ups close plan-named acceptance criteria.

## TDD Gate Compliance

Task 2 followed the planned test-first cycle. Its initial known-good receipt test failed because `scripts/check-receipts.sh` did not exist; after the gate was implemented, that same test passed along with the full differential catalog. The phase config had `workflow.tdd_mode: false`, so no plan-level RED/GREEN commit gate was active.

## Verification

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run 'TestFreezeGate|TestPhase32CodeDigest|TestPhase32FreezeGate|TestReceiptGateDifferential|TestReceiptGate|TestPhase32ReceiptGate' -count=1` — passed after the implementation and named acceptance cases were complete (31.68s).
- `sh -n scripts/check-freeze.sh scripts/check-receipts.sh scripts/codedigest.sh` — passed.
- `TestReceiptGateDifferential` — all 35 differential subtests passed, including every old refusal reason and the three manifest variants.
- `TestReceiptGateArchivedM006` — archived pair, six mutations, and manifest-to-producer equality passed; archived receipt and logs remained unchanged.
- M006 manifest case rows matched `git show 9f8f878f:scripts/verify-phase32.sh` byte-for-byte.
- Shell syntax, manifest static checks, archived-artifact immutability, and `git diff --check` — passed. No full project suite, hosted receipt binding, CI workflow, push, or archived receipt/log edit was performed.

## Issues Encountered

- The sandbox initially rejected Git index and `.git` writes. The authorized escalation was approved, after which atomic commits and the plan ledger were created normally.
- The first Markdown-only fixture inherited an earlier code-change commit; the fixture was corrected to branch from the freeze commit and then passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

Plan 33-03 is complete. Plan 33-04 can use the generic script names while preserving the old scripts until the differential outputs have been frozen into goldens.

---
*Phase: 33-carried-debt-bounded-capture-read-and-generic-freeze-gate*
*Completed: 2026-10-13*

## Self-Check: PASSED

All seven plan-owned files exist, all eight measured plan commits are ancestors of `ff24ec16`, all 39 manifest rows match the archived producer table, and the archived receipt directory remained unchanged.
