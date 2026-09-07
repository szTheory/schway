---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 11
subsystem: testing
tags: [go, traceability-registry, spikes, fail-closed-audit, json, embed]

# Dependency graph
requires:
  - phase: 05-native-equivalence-and-adversarial-evidence (05-09)
    provides: "Phase5RequiredControls()/Phase4RequiredControls() as the shipped required-control sets this registry cross-checks; the D-05-40 mid-phase gate release for waves 5-7"
provides:
  - "internal/compiler/session/qlt01_registry.json — 22-row machine-readable QLT-01 control registry extracted from spikes 001-005"
  - "session.LoadQLT01Registry / session.AllShippedControlIDs / session.AuditQLT01Registry — the loader and fail-closed completeness audit (D-05-29)"
  - "session.VerifyQLT01Registry / session.QLT01LaneFromRows / session.LaneQLT01RegistryAudit — the counted-work lane plan 05-14 wires into the gate"
affects: [05-13, 05-14]

actuals:
  tokens: 32000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "A go:embed JSON data file plus a Go loader and a fail-closed completeness audit, mirroring the existing Phase 4/5 required-control-set duplication-plus-set-equality pattern (Phase4RequiredControls/TestPhase4RequiredControlsMatchScript), applied here to registry rows instead of control-ID strings."
    - "A waiver's citation names a concrete, grep-able production mechanism or test (a function, constant, or test name), never a vague 'superseded' claim — the same falsifiability discipline this project already applies to escape: identifiers."

key-files:
  created:
    - internal/compiler/session/qlt01_registry.json
    - internal/compiler/session/qlt01.go
    - internal/compiler/session/qlt01_test.go
  modified: []

key-decisions:
  - "Only Phase 4/5 control:-prefixed identifiers (AllShippedControlIDs(), built from Phase4RequiredControls()+Phase5RequiredControls()) are valid live_descendant.control_id citations, since those are the only identifiers the registry's own audit can cross-check for staleness. Every ownership-kernel hazard from spikes 001-004 predates that native/FFI-focused control vocabulary and has no control:-prefixed identifier of its own, so all 13 of their rows are legitimately waived with citations to the concrete production mechanism or test that supersedes each hazard (TestOwnershipSequenceExhaustive, TestBranchSequenceExhaustive, D-05-34/D-05-35's discoverLoanLastUses retirement, D-12's independence rule, Box/Pair's execution-admission-gate rejection, originvalidate.KnownEscape, evidence.Manifest's SHA-256 identity) rather than to a vague 'not relevant'."
  - "Spike 005 iteration 5's longjmp/no-sanitizer-report row is recorded as a LIVE descendant citing Phase 4's control:foreign.nonlocal_exit_undetected (not a Phase 5 sanitizer control, and not waived) — its control_mechanism explicitly names the no-sanitizer-report property, per the plan's own instruction that getting this row wrong would be the exact overclaim the registry exists to prevent."
  - "control:native.sanitize.ubsan_no_recover's live_descendant.fixture cites internal/compiler/session/session_phase5_sanitize.go (the file containing the inline ubsanTriggerFixtureSource Go constant) rather than a testdata path, since the UBSan positive control is an inline string, not a checked-in fixture file — task 2's fixture-exists-on-disk audit condition (c) still requires a real path, and this is the honest one."
  - "Tasks 1-3 were implemented as one cohesive design pass but split into three sequential commits along the plan's own task boundaries (registry file + minimal loader/tests first, then the audit, then the lane), matching each task's own <verify> command rather than committing the whole file in one shot — see Deviations."

requirements-completed: [QLT-01]

coverage:
  - id: D1
    description: "Every negative control and reduced counterexample from spikes 001-005 is extracted into qlt01_registry.json, with a falsifiable control_mechanism and exactly one of live_descendant or waived per row"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestQLT01RegistryParses"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestQLT01RegistryCoversAllFiveSpikes"
        status: pass
    human_judgment: false
  - id: D2
    description: "TestQLT01RegistryComplete fails on an unaccounted row (neither disposition), a dual disposition, a stale control: reference, a missing fixture, or an empty registry"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestQLT01RegistryComplete"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestQLT01AuditGoesRedOnStaleControl"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestQLT01AuditGoesRedOnDualDisposition"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestQLT01AuditGoesRedOnEmptyRegistry"
        status: pass
    human_judgment: false
  - id: D3
    description: "The registry audit runs as a counted-work lane (lane:qlt01-registry-audit) that cannot report nonzero work on an empty registry, and never reports zero work with a pass"
    requirement: QLT-01
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestQLT01LaneCountsWork"
        status: pass
      - kind: unit
        ref: "internal/compiler/session#TestQLT01LaneEmptyRegistryReportsZeroWorkAndFails"
        status: pass
    human_judgment: false
  - id: D4
    description: "Extraction covers control mechanism and identity from all five spikes without regressing session.go, session_phase5.go, or full go test/vet/verify-phase5.sh"
    requirement: QLT-01
    verification:
      - kind: integration
        ref: "sh scripts/verify-phase5.sh (full run, exit 0, all lanes pass)"
        status: pass
      - kind: integration
        ref: "go test ./... && go vet ./... (full repo)"
        status: pass
    human_judgment: true
    rationale: "The 22-row extraction's fidelity to the five spike READMEs (whether a given hazard is honestly waived vs. force-fit as a live descendant) is a documentary-accuracy judgment plan 05-14's human review gate should confirm alongside the waiver-set summary below, not something an automated check alone can certify."

duration: ~45min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 11: QLT-01 Control Registry Extraction Summary

**A 22-row machine-readable registry (`qlt01_registry.json`) extracts every negative control and reduced counterexample from spikes 001-005, gated by a fail-closed completeness audit and a counted-work lane — 9 rows cite live Phase 4/5 `control:` identifiers, 13 are honestly waived with falsifiable citations to the production mechanisms that supersede each ownership-kernel-era hazard.**

## Performance

- **Duration:** ~45 min
- **Started:** 2026-09-06T14:50:00Z (approx, first file read)
- **Completed:** 2026-09-06T15:14:00Z
- **Tasks:** 3 (all `type="auto"`)
- **Files modified:** 3 (all newly created)

## Accomplishments

- `internal/compiler/session/qlt01_registry.json` — one row per negative control / reduced counterexample recorded across spikes 001-005's READMEs and (for spike 005) its actual `native/`/`lab/` C and Go sources, walked in full per the plan's `<read_first>`. 22 rows total: 3 (spike 001), 3 (002), 4 (003), 3 (004), 9 (005).
- `session.LoadQLT01Registry`/`session.AllShippedControlIDs`/`session.AuditQLT01Registry` implement D-05-29's fail-closed completeness audit: neither/both disposition, a stale `control:` reference, a missing fixture, or an empty registry are all distinct hard failures naming the offending `spike_id`/`control_id`.
- `session.VerifyQLT01Registry`/`session.QLT01LaneFromRows`/`session.LaneQLT01RegistryAudit` run the audit as a counted-work lane (one unit per row inspected plus one per control-ID cross-check), proven unable to report nonzero work on an empty registry.
- `session.go` and `session_phase5.go` are untouched (verified by `git diff --stat`); `Phase5RequiredControls()` was deliberately NOT extended — plan 05-14 owns that update together with the shell gate.

## Task Commits

Each task was committed atomically:

1. **Task 1: Extract control mechanism and identity from all five spikes into the registry file** - `c2a3761` (feat)
2. **Task 2: Load the registry in Go and add the fail-closed completeness audit** - `85c0b6d` (feat)
3. **Task 3: Wire the registry audit into the Phase 5 lane with counted work** - `3aefa84` (feat)

**Plan metadata:** committed as part of this SUMMARY's own commit (see below).

## Files Created/Modified

- `internal/compiler/session/qlt01_registry.json` (22 rows) — the QLT-01 registry
- `internal/compiler/session/qlt01.go` — `QLT01Row`/`QLT01LiveDescendant`/`QLT01Waiver`, `LoadQLT01Registry`, `AllShippedControlIDs`, `AuditQLT01Registry`, `VerifyQLT01Registry`, `QLT01LaneFromRows`, `LaneQLT01RegistryAudit`, `ControlQLT01RegistryIncomplete`, `ControlQLT01StaleControlReference`
- `internal/compiler/session/qlt01_test.go` — `TestQLT01RegistryParses`, `TestQLT01RegistryCoversAllFiveSpikes`, `TestQLT01RegistryComplete`, `TestQLT01AuditGoesRedOnStaleControl`, `TestQLT01AuditGoesRedOnDualDisposition`, `TestQLT01AuditGoesRedOnEmptyRegistry`, `TestQLT01LaneCountsWork`, `TestQLT01LaneEmptyRegistryReportsZeroWorkAndFails`

## Registry Summary (for plan 05-14's human review gate)

**22 rows: 9 live descendants, 13 waived.**

| Spike | Live descendants | Waived |
|---|---|---|
| 001 (ownership-kernel-workbench) | 0 | 3 |
| 002 (cfg-edge-last-use) | 0 | 3 |
| 003 (public-origins-generic-abilities) | 0 | 4 |
| 004 (independent-certificate-checker) | 0 | 3 |
| 005 (native-ffi-provenance-cleanup) | 9 | 0 |

**Waived rows and their citations** (as recorded in the registry):

1. `001 / spike-001-move-while-borrowed-fault` — `check.go`'s `ownership.move_while_borrowed` + `corevalidate.go`'s `core.move_while_borrowed`, proven by `TestOwnershipSequenceExhaustive`.
2. `001 / spike-001-unsound-branch-join-fault` — `corevalidate.go`'s `core.loan_endpoint_mismatch`, proven by `TestBranchSequenceExhaustive`.
3. `001 / spike-001-manual-inferred-loan-end-divergence` — 05-09-SUMMARY.md's recorded `discoverLoanLastUses` retirement (D-05-34/D-05-35), zero source-level loan-ending syntax in this language.
4. `002 / spike-002-omitted-edge-endpoints-fault` — `check.go`'s `loanLivenessFixpoint` (D-05-34/D-05-35), the sole liveness law as of 05-02-SUMMARY.md.
5. `002 / spike-002-malformed-cfg-graph` — `corevalidate.go`'s `core.terminal_block_unreachable` and sibling structural predicates.
6. `002 / spike-002-oracle-masking-defense` — D-05-35's explicit divergence-logging (not silent-repair) discipline.
7. `003 / spike-003-under-declared-origin-fault` — Box/Pair rejected at `check.go`'s execution-admission gate (Phase 03-06 decision); `RecomputeOriginPerReturn`/`ValidatePublished` cover the single-parameter case that DOES exist.
8. `003 / spike-003-ability-propagation-leak` — Buffer's independent `share` grant (OV-02-01); M001 ships no generic containers.
9. `003 / spike-003-shallow-copy-interface-aliasing` — D-12/D-05-39's independence rule (no shared helpers between `check`/`corevalidate`).
10. `003 / spike-003-higher-ranked-origin-escape` — D-05-32/D-05-33, `OpCall` and closures out of M001's scope.
11. `004 / spike-004-coordinated-frontend-summary-lie` — `originvalidate.KnownEscape = "escape:coordinated-frontend-summary-lie"` (matches the spike's own escape name verbatim); D-05-30 schedules the reachability demonstration for plan 05-13.
12. `004 / spike-004-certificate-mutation-coverage` — D-12/D-05-39 plus the spike's own README on where an independent validator belongs (trust crossings M001 has not built).
13. `004 / spike-004-digest-reorder-sensitivity` — `evidence.Manifest`'s "SHA-256 is content identity only" (Phase 2 decision); no certificate-replay consumer exists in M001 to test reorder sensitivity against.

**Live-descendant rows** (all citing real, currently-shipped `control:` identifiers, cross-checked against `AllShippedControlIDs()`):

1. `005 / spike-005-iter2-layout-mismatch` → `control:foreign.layout_mismatch` (`testdata/phase4/foreign_layout_mismatch.golden.c`)
2. `005 / spike-005-iter3-release-omitted` → `control:resource.release_omitted` (`testdata/phase4/acquire_three_success.lang`)
3. `005 / spike-005-iter3-release-order-transposed` → `control:resource.release_order_transposed` (`testdata/phase4/acquire_three_success.lang`)
4. `005 / spike-005-iter3-allocator-identity-mismatch` → `control:native.sanitize.allocator_mismatch` (`testdata/phase5/allocator_mismatch.lang`)
5. `005 / spike-005-iter3-stale-callback-retention` → `control:native.sanitize.retained_pointer` (`testdata/phase5/retained_pointer.lang`)
6. `005 / spike-005-iter4-false-restrict-alias` → `control:alias.false_no_alias` (`testdata/phase5/false_restrict_hoist.lang`)
7. `005 / spike-005-iter5-heap-use-after-free` → `control:native.sanitize.use_after_free` (`testdata/phase5/retained_pointer.lang`)
8. `005 / spike-005-iter5-ubsan-no-recover` → `control:native.sanitize.ubsan_no_recover` (`internal/compiler/session/session_phase5_sanitize.go`, the file holding the inline UBSan positive-control constant)
9. `005 / spike-005-iter5-nonlocal-exit-no-sanitizer-report` → `control:foreign.nonlocal_exit_undetected` (`testdata/phase4/nonlocal_exit_probe.lang`) — the load-bearing row naming the no-sanitizer-report property per the plan's explicit instruction

## Decisions Made

See `key-decisions` in frontmatter: the control-ID vocabulary boundary that makes all of spikes 001-004 waived and all of spike 005 (mostly) live-descendant; the spike-005-iteration-5 row's citation to Phase 4's nonlocal-exit control rather than a waiver; the UBSan fixture-path citation to a Go source file instead of a testdata path; and the three-way task-boundary commit split.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking issue] Task 1's own `<verify>` command required Go test functions not scoped to Task 1's declared `<files>`**

- **Found during:** Task 1, before writing the registry
- **Issue:** Task 1's `<files>` lists only `qlt01_registry.json`, but its `<verify>` runs `TestQLT01RegistryParses` and `TestQLT01RegistryCoversAllFiveSpikes` — both of which require `session.LoadQLT01Registry` and a Go test file, neither of which exists until (per the plan's own task 2 description) `qlt01.go`/`qlt01_test.go` are created. Running Task 1's verify command against only the JSON file would fail with "no test files".
- **Fix:** Created a minimal `qlt01.go` (embed directive, the three row/descendant/waiver types, `LoadQLT01Registry` only) and a minimal `qlt01_test.go` (the two named tests plus their `discoverSpikeIDs` helper) as part of Task 1's commit, then incrementally added Task 2's audit machinery and Task 3's lane machinery to the SAME two files in their own commits — so each task's own `<verify>` command runs against exactly the code that commit introduces, and the final state matches Task 2/3's file-scope declarations (`qlt01.go`, `qlt01_test.go`) exactly.
- **Files modified:** `internal/compiler/session/qlt01.go`, `internal/compiler/session/qlt01_test.go` (both created incrementally across all three task commits)
- **Verification:** Each task's own `<verify>` command was run and passed against that task's commit in isolation (see Task Commits above); the full three-commit sequence was re-verified together via `sh scripts/verify-phase5.sh` (exit 0) and `go test ./... && go vet ./...` (green) after Task 3.
- **Committed in:** `c2a3761`, `85c0b6d`, `3aefa84`

---

**Total deviations:** 1 auto-fixed (1 blocking issue — Rule 3). No production code outside this plan's own new files was touched.
**Impact on plan:** None on scope or content — this only affected which task's commit introduced which lines within the two Go files the plan itself scopes to tasks 2 and 3; the committed end state matches the plan's file list exactly.

## Issues Encountered

None beyond the deviation above.

## User Setup Required

None — no external service configuration required.

## Known Stubs

None. Every registry row resolves to either a real, verified live descendant (fixture exists on disk, control ID exists in the shipped set, cross-checked by `TestQLT01RegistryComplete`) or a waiver with a citation to specific, grep-able production code, a test, or a state-file decision record — none of the forbidden free-text placeholders appear anywhere in the registry.

## Next Phase Readiness

- The registry, loader, audit, and lane are all in place and independently green. `Phase5RequiredControls()` and `scripts/verify-phase5.sh` are untouched, exactly as the plan requires — plan 05-14 is the designated integration point for adding `LaneQLT01RegistryAudit`'s two controls to the shipped required-control set alongside the reducer's own controls and the shell gate's parity block.
- Plan 05-13 should additionally close `spike-004-coordinated-frontend-summary-lie`'s forward reference: this row already cites the shipped `originvalidate.KnownEscape` constant; once plan 05-13 lands `escape:coordinated-source-to-core-false-claim`'s adversarial-program demonstration (D-05-30), no registry change is required here since the citation already points at the correct shipped constant.
- The waiver set above is ready for plan 05-14's human review gate to confirm as a whole per the plan's own instruction ("so the human gate in plan 05-14 can review the waiver set as a whole rather than row by row").

## Self-Check: PASSED

- `internal/compiler/session/qlt01_registry.json` — FOUND
- `internal/compiler/session/qlt01.go` — FOUND
- `internal/compiler/session/qlt01_test.go` — FOUND
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-11-SUMMARY.md` — FOUND
- Commits `c2a3761`, `85c0b6d`, `3aefa84` — all present in `git log --oneline --all`
- All plan-level `<verification>` commands re-run and green: `sh scripts/verify-phase5.sh` (exit 0), `go test ./... && go vet ./...` (green)

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*
