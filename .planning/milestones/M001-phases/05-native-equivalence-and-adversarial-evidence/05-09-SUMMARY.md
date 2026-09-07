---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 09
subsystem: testing
tags: [go, verification-gate, sanitizers, lto, mid-phase-gate, debt-register]

# Dependency graph
requires:
  - phase: 05-native-equivalence-and-adversarial-evidence (05-04, 05-05, 05-06, 05-07, 05-08)
    provides: "Phase 1-4 identity pins, Phase 5 corpus + LTO tier, comparator + field routing, alias-fact mutation table, sanitizer lane and fixtures"
provides:
  - "Phase5RequiredControls() — the ten-identifier control set for the Phase 5 gate"
  - "scripts/verify-phase5.sh — the Phase 5 verify entry point, a peer of verify-phase4.sh, wired into `lang verify`/`lang release` via cmd/lang/main.go"
  - "TestPhase5RequiredControlsMatchScript / TestPhase5CorpusBoundMatchesScript — bidirectional set-equality between the Go control set and the shell gate"
  - "D-05-22 closed for all six subjected NAT-03 rows (citation existence + the sanitizer row's claimed-axis proof)"
  - "The mid-phase gate (D-05-40): human-confirmed green, releasing waves 5-7 (reducer, QLT-01 registry)"
affects: [05-10, 05-11, 05-12, 05-13, 05-14]

actuals:
  tokens: 22000
  tasks: 3
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Phase gate scripts are hand-duplicated peers, never forked or helper-shared, with set-equality tests enforcing parity in both directions (established Phase 4, continued here)."
    - "A `tool_missing`/`sanitizer_inert` lane is surfaced as a named operational obligation in the verify JSON, never silently rendered as a pass."
    - "Cross-plan citation obligations (D-05-22 style `PENDING-05-0N` markers) are closed at the first plan ordered after both producing plans, with the closing plan limited to deleting the marker comments — no other field in the frozen file may change."

key-files:
  created:
    - internal/compiler/session/session_phase5.go
    - internal/compiler/session/session_phase5_test.go
    - scripts/verify-phase5.sh
    - .planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md
  modified:
    - internal/compiler/session/session_phase5_alias.go
    - cmd/lang/main.go

key-decisions:
  - "Phase5AssertMutationMovesAnAxis is a thin dispatcher added alongside 05-07's own switch, not an edit to it — keeps session_phase5_alias.go's diff to exactly the two PENDING-05-08 comment deletions."
  - "cmd/lang/main.go's runVerify dispatches testdata/phase5 to session.VerifyPhase5ControlsAndWork via an isPhase5Corpus marker check (restrict_borrow.lang), keeping session.go byte-frozen for the whole phase."
  - "D-05-41 (the gate's own control:interpreter-o0-o3-lto lane samples one adversarial fixture, not the full six-fixture subset, at the LTO tier) is accepted as recorded debt, not fixed now — go test's own TestPhase5CorpusThreeEngineAgreement and plan 05-06's TestLTOTierIsNotInert already cover the correctness gap the sample leaves; widening the lane is optional future work, landing phase 'Phase 5+'."
  - "05-DEBT.md's Items table was missing the 'Landing phase' column the standing TestDebtRegistersAreWellFormed test requires for any non-exempt register (only 02/03-DEBT.md are exempted as frozen prior art) — added during continuation close-out, since the gate was not actually green without it. Documented as a deviation below."

requirements-completed: [NAT-02, NAT-03]

coverage:
  - id: D1
    description: "Phase5RequiredControls() declares the ten-identifier Phase 5 control set; every control fires with nonzero RecomputedWork in VerifyPhase5ControlsAndWork"
    requirement: NAT-02
    verification:
      - kind: integration
        ref: "internal/compiler/session#TestVerifyPhase5ControlsAndWork"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestPhase5ControlsAllHaveNonzeroWork"
        status: pass
    human_judgment: false
  - id: D2
    description: "A tool_missing/sanitizer_inert sanitizer lane is reported as a named operational obligation, never a silent pass; the sanitizer lane is excluded from edit/check pipelines"
    requirement: NAT-02
    verification:
      - kind: integration
        ref: "internal/compiler/session#TestPhase5ToolMissingIsNamedObligationNotPass"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestSanitizeLaneNotInEditOrCheck"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-05-22 closed: all six Subjected:true NAT-03 rows cite a CorpusProgram that exists on disk (no row still PENDING-05-08), and the sanitizer allocator_mismatch row is proven to move its own claimed axis"
    requirement: NAT-03
    verification:
      - kind: integration
        ref: "internal/compiler/session#TestNAT03MutationsCiteExistingPrograms"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestNAT03SanitizerRowMovesItsClaimedAxis"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestNoNAT03RowRemainsPending"
        status: pass
    human_judgment: false
  - id: D4
    description: "scripts/verify-phase5.sh is a peer of verify-phase4.sh (never invokes it), re-runs testdata/phase1-5 with the freshly built binary, pins ASAN_OPTIONS/UBSAN_OPTIONS explicitly, and is bidirectionally set-equality-tested against Go for both the control set and the D-05-18 corpus bound"
    requirement: NAT-02
    verification:
      - kind: integration
        ref: "internal/compiler/session#TestPhase5RequiredControlsMatchScript"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestPhase5CorpusBoundMatchesScript"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestPhase5VerifierScriptContract"
        status: pass
      - kind: integration
        ref: "internal/compiler/session#TestPhase5SanitizerOptionsMatchScript"
        status: pass
    human_judgment: false
  - id: D5
    description: "The D-05-40 mid-phase gate itself is green and human-confirmed before waves 5-7 (reducer, QLT-01 registry) begin"
    requirement: NAT-02
    verification: []
    human_judgment: true
    rationale: "Gate release is a human sign-off on a written gate record (control status, NAT-03 arithmetic, toolchain identity, discoverLoanLastUses retirement, frozen-path diff-stat, and any newly discovered debt) — by design not something an automated check alone can approve."

duration: 21min active (task 1-2 execution) + human review hold + 15min continuation close-out
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 9: Mid-Phase Gate (D-05-40) Summary

**Phase5RequiredControls() plus scripts/verify-phase5.sh close D-05-22 for all six subjected NAT-03 rows and pass a human-confirmed mid-phase gate — nine lanes, ten controls, all nonzero work, releasing the reducer and QLT-01 registry waves.**

## Performance

- **Duration:** ~21 min active execution (Tasks 1-2, 02:01-02:23 local) + a human-review hold + ~15 min continuation close-out (gate re-verification and a debt-register format fix)
- **Started:** 2026-09-06T06:01:53Z
- **Completed:** 2026-09-06T14:06:08Z (final metadata commit follows)
- **Tasks:** 3 (2 auto + 1 checkpoint:human-verify gate)
- **Files modified:** 6 (2 created Go files, 1 created shell script, 1 created debt register, 1 modified alias file, 1 modified CLI entrypoint) + 1 debt-register fix during close-out

## Accomplishments

- `Phase5RequiredControls()` declares exactly the ten Phase 5 control identifiers as literal strings; `Phase5ExpectedEscapes()` returns `escape:callback-invocation-unsubjected`; `VerifyPhase5ControlsAndWork` peers `verifyForeignCorpus`'s addLane/fail shape and drives the alias-fact, comparator/field-routing, LTO-tier, and sanitizer lanes.
- `scripts/verify-phase5.sh` is a byte-level peer of `scripts/verify-phase4.sh` (never invokes it, never forks a shared helper): mktemp/trap/GOCACHE preamble, `assert-go-tests.sh --self-test`, `go test`/`-race`/`vet`/`build`, five `--json verify testdata/phaseN` invocations re-running Phases 1-4 with the freshly-built Phase 5 binary for non-regression, the Phase 5 control for-loop under its own anchor comment, and the D-05-18 corpus-bound anchor block. `ASAN_OPTIONS`/`UBSAN_OPTIONS` are pinned explicitly on the sanitizer invocation, byte-identical to `native.ASanOptions`/`native.UBSanOptions`.
- D-05-22 closed for the two rows plan 05-07 could not assert concurrently with plan 05-08: all six `Subjected: true` NAT-03 rows now cite a `CorpusProgram` that exists on disk, no row still carries the `PENDING-05-08` marker, and the `control:native.sanitize.allocator_mismatch` row is proven (via a new thin dispatcher, `Phase5AssertMutationMovesAnAxis`) to move its own claimed axis against plan 05-08's `allocator_mismatch.lang` fixture.
- The mid-phase gate itself: a full run of `sh scripts/verify-phase5.sh` plus `go test ./... && go test -race ./... && go vet ./...` is green, and the gate record below was reviewed and approved by the human.

## Task Commits

Each task was committed atomically:

1. **Task 1: Declare Phase5RequiredControls and the phase-5 verify entry point** - `bddad86` (feat)
2. **Task 2: Write scripts/verify-phase5.sh as a peer of verify-phase4.sh with a set-equality parity test** - `2383565` (feat)
3. **Gate record: mid-phase debt register** - `6703962` (docs) — written as part of holding the Task 3 checkpoint
4. **Close-out fix: add missing "Landing phase" column to 05-DEBT.md** - `e400c84` (fix) — found and fixed during this continuation, see Deviations

**Plan metadata:** committed as part of this SUMMARY's own commit (see below).

## Files Created/Modified

- `internal/compiler/session/session_phase5.go` (377 lines) - `Phase5RequiredControls`, `Phase5ExpectedEscapes`, `VerifyPhase5ControlsAndWork`, `Phase5AssertMutationMovesAnAxis`
- `internal/compiler/session/session_phase5_test.go` (412 lines) - the gate's own test suite (control-set parity, corpus-bound parity, script contract, sanitizer-option parity, NAT-03 citation/pending/axis tests, tool-missing/named-obligation test, edit/check exclusion test)
- `scripts/verify-phase5.sh` (110 lines) - the Phase 5 verify entry point, peer of `verify-phase4.sh`
- `internal/compiler/session/session_phase5_alias.go` - two `PENDING-05-08` marker-comment deletions only; no field value changed
- `cmd/lang/main.go` - `runVerify` dispatches `testdata/phase5` to `session.VerifyPhase5ControlsAndWork` via an `isPhase5Corpus` marker check; `session.go` stays byte-frozen
- `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md` (created at the gate, fixed during close-out) - D-05-41

## Decisions Made

See `key-decisions` in frontmatter: the `Phase5AssertMutationMovesAnAxis` thin-dispatcher pattern that preserves 05-07's frozen switch; the `cmd/lang/main.go` wiring that keeps `session.go` untouched; D-05-41 accepted as recorded debt rather than fixed now; and the close-out fix to `05-DEBT.md`'s missing column.

## Mid-Phase Gate Record (D-05-40)

**Gate verdict: PASS, human-confirmed.**

Ten required controls, nine lanes, all firing with nonzero `RecomputedWork` in the final `sh scripts/verify-phase5.sh` run (`testdata/phase5` JSON, `status: "pass"`):

| Lane | Control(s) | RecomputedWork |
|---|---|---|
| lane:foreign-no-unproven-attributes | control:foreign.no_unproven_attributes | 1 |
| lane:attribute-unjustified | control:core.attribute_unjustified | 1 |
| lane:alias-false-no-alias | control:alias.false_no_alias | 3 |
| lane:diagnostic-reject-program-id-equivalence | control:diagnostic.reject_program_id_equivalence | 2 |
| lane:compare-field-routing | control:compare.field_routing_unrouted | 18 |
| lane:interpreter-o0-o3-lto | control:interpreter-o0-o3-lto | 4 |
| lane:native-sanitize | control:native.sanitize.retained_pointer, control:native.sanitize.use_after_free | 1 |
| lane:native-sanitize | control:native.sanitize.allocator_mismatch | 1 |
| lane:native-sanitize | control:native.sanitize.ubsan_no_recover | 1 |

`expected_escapes: ["escape:callback-invocation-unsubjected"]` is declared and matched — not a silent pass.

**NAT-03 arithmetic:** six subjected rows (all citing an on-disk `CorpusProgram`, none still `PENDING-05-08`) plus one subsumed-with-named-escape row — **6/6 subjected, not 7/7** — matching the plan's required reading exactly. The sanitizer `allocator_mismatch` row is independently proven to move its own claimed axis (`TestNAT03SanitizerRowMovesItsClaimedAxis`) against plan 05-08's fixture, not merely to show *some* disagreement.

**Toolchain identity:** Apple clang 21.0.0 (clang-2100.1.1.101), target `arm64-apple-darwin25.6.0` — matching the identity recorded in 05-RESEARCH.md and re-confirmed live this session.

**`discoverLoanLastUses` retirement:** confirmed retired — `grep -rn "discoverLoanLastUses" --include="*.go" .` returns only comment references (no function definition, no call site). The retirement's shadow run (`loanLivenessFixpoint` widened to straight-line bodies, shadow-run against the old scan) recorded **230,692** admission-site comparisons with zero divergences before the old scan was deleted (05-02-SUMMARY.md), well above the plan's 113,000-case floor.

**Frozen-path diff-stat:** `git diff --stat` over `testdata/phase1`, `testdata/phase2`, `testdata/phase3`, `testdata/phase4`, `native/lang_foreign_resource.c`, `native/lang_foreign_nonlocal.c`, and `scripts/verify-phase4.sh` reports **no changes**. `internal/compiler/session/session.go` reports **no changes**. `session_phase5_alias.go`'s diff against its 05-07 state is exactly the two `PENDING-05-08` comment deletions — no `ControlID`, `CorpusProgram`, `ExpectedAxis`, `Subjected`, or `EscapeID` field changed.

**Debt discovered:** one item, `D-05-41` — the gate's own `control:interpreter-o0-o3-lto` lane samples one adversarial fixture (`inline_across_foreign.lang`) rather than the full six-fixture adversarial subset at the LTO tier. Not a correctness gap: `TestPhase5CorpusThreeEngineAgreement` already proves interpreter/-O0/-O3 agreement across the full corpus, and plan 05-06's `TestLTOTierIsNotInert` independently proves the LTO tier is non-inert. Recorded as `info` severity, landing phase "Phase 5+", accepted **as recorded debt** — not widened in this plan.

**Human approval:** the developer reviewed the full gate record and independently-verified evidence and approved without requested changes, explicitly confirming: all ten controls across nine lanes with nonzero work; the six-plus-one NAT-03 arithmetic; both file-scope deviations (the thin dispatcher and the `cmd/lang/main.go` wiring); and D-05-41 accepted as recorded, deliberately not widened now.

Waves 5-7 (reducer, plans 05-10/05-12; QLT-01 registry, plans 05-11/05-13) are released.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed 05-DEBT.md's missing "Landing phase" column**

- **Found during:** Continuation close-out, re-verifying the gate before writing this SUMMARY
- **Issue:** The standing test `TestDebtRegistersAreWellFormed` (added in Phase 4) requires every debt register not in a small frozen-prior-art exemption list (`02-DEBT.md`, `03-DEBT.md` only) to carry a "Landing phase" column. The `05-DEBT.md` committed at the gate (`6703962`) predates this discovery and omitted the column, which made `go test ./...` — and therefore `sh scripts/verify-phase5.sh` — actually fail. The gate as literally re-run at close-out time was red, not green, until this was fixed.
- **Fix:** Added a "Landing phase" column to the Items table header and separator, and a value of "Phase 5+" for `D-05-41`, matching the value implied by the item's own "Phase 5+ fix" prose and the `04-DEBT.md` convention (e.g. "Phase 5", "M002 or Phase 6").
- **Files modified:** `.planning/phases/05-native-equivalence-and-adversarial-evidence/05-DEBT.md`
- **Verification:** `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v` passes for all four registers (`02-DEBT.md`, `03-DEBT.md`, `04-DEBT.md`, `05-DEBT.md`); a full `sh scripts/verify-phase5.sh` re-run afterward exits 0 with `testdata/phase5` reporting `status: "pass"`.
- **Committed in:** `e400c84`

---

**Total deviations:** 1 auto-fixed (1 bug — Rule 1). No production Go code was touched; the fix was to a `.planning/` debt-register document's table shape.
**Impact on plan:** Necessary to make the gate's own re-verification actually pass end-to-end as the human's approval assumed. No scope creep — D-05-41's content and severity are unchanged, only the required column was added.

## Issues Encountered

None beyond the deviation above, which is documented there rather than here since it was a concrete auto-fixable defect, not an open problem.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The mid-phase gate (D-05-40) is green and human-confirmed. Plans 05-10 through 05-14 (reducer, waves 5-6; QLT-01 registry, waves 6-7) may now begin, consuming the settled interpreter/-O0/-O3(-LTO) differential and the alias-fact/attribute-justification work this gate proved.
- `D-05-41` is carried debt (info severity, landing phase "Phase 5+") — no blocker for subsequent plans, since none of them consume the LTO tier's per-fixture coverage specifically.
- `Phase5RequiredControls()` and `Phase5ExpectedEscapes()` are the extension points later plans use: plan 05-13 adds `escape:coordinated-source-to-core-false-claim`; plan 05-14 extends both lists with the reducer and registry controls.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*
