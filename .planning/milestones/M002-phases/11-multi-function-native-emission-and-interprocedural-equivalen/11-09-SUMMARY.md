---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
plan: 09
subsystem: testing
tags: [reduce, hdd, qlt-05, session, callgraph, corevalidate]

# Dependency graph
requires:
  - phase: 11-multi-function-native-emission-and-interprocedural-equivalen
    provides: "11-08's reduce.Seed/dropCallSite/dropOrphanFunction (Q-01 BRANCH A, live moves) and TestOperationIDsUnchangedByWholeProgramMoves' non-renumbering guarantee; 11-05's four-tier differential and its D-11-51/D-11-52 debt"
provides:
  - "foreignCallSequenceFor(program, o0Run): a two-knower guard against reduction slippage -- staticForeignCallSequenceFor walks every function in callgraph.Order's reverse-postorder, dynamicForeignCallSequenceFor derives the same fact independently from the -O0 run's event stream, and disagreement is a hard failure (reduce.foreign_call_sequence_disagreement)"
  - "session.QLT05Reverify: strict, cold-start re-verification (Axis/EnginePair/OperationID field equality, no positional fallback, plus foreign-call-sequence equality) reported as control:reduce.reverified"
  - "testdata/phase11/multi_function_reduce_gate.lang: the anti-vacuity gate fixture, and the paired TestQLT05GateIsNonVacuous/TestQLT05EmptyReductionFailsAntiVacuity tests"
  - "11-VERIFICATION-INPUTS.md: the phase's four collected evidence-weakened claims (NAT-05, NAT-06/2, QLT-06b, QLT-05 full-not-narrowed)"
affects: []

# Actuals (#2632)
actuals:
  tokens: 62000
  tasks: 3
  commits: 2

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Two-independent-knowers guard applied to a single silent-nil hole (D-11-33): a static program-structure walk and a dynamic execution-event-stream walk, sharing no helper beyond the event type, compared and failed loudly on disagreement rather than trusted alone."
    - "Injectable QLT05SignatureDeriver seam: QLT05Reverify's own strict-comparison/cold-start/idempotence properties are tested against a deterministic, program-structure-derived double, while qlt05RealSignatureDeriver wires the real corevalidate+cgen+native pipeline for production -- the properties under test are orthogonal to how the fresh signature is obtained."

key-files:
  created:
    - internal/compiler/session/session_qlt05_reverify_test.go
    - testdata/phase11/multi_function_reduce_gate.lang
    - .planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VERIFICATION-INPUTS.md
  modified:
    - internal/compiler/session/session_phase5_mismatch.go
    - internal/compiler/session/session_phase5_mismatch_test.go
    - internal/compiler/session/export_test.go

key-decisions:
  - "The anti-vacuity assertion targets \"drop-orphan-function\" (not \"drop-call-site\"), since Q-01 BRANCH A shipped BOTH whole-program moves live (11-08): the fixture's removable function (helper) is only actually gone from AppliedMoves once drop-orphan-function fires after drop-call-site rewires its last call site."
  - "QLT05Reverify accepts an injectable QLT05SignatureDeriver rather than hardcoding the real corevalidate/cgen/native pipeline: this plan's own five reverification tests exercise the strict-comparison, cold-start, and idempotence properties against a deterministic, program-structure-derived test double (qlt05StructuralDeriver), since those properties are orthogonal to how the fresh signature is obtained. Production wiring (qlt05RealSignatureDeriver) mirrors mismatchPredicate's real pipeline byte for byte but is not itself exercised by a genuine multi-function native alias-divergence fixture in this plan -- engineering one (AliasFactMutationRunner's marker-uniqueness guard interacts awkwardly with a wrapper/forwarding function) was judged out of scope for this plan's own effort budget. Recorded as a documented simplification, not a stub: QLT05Reverify's own logic is fully tested; only the production deriver's real end-to-end wiring is untested by this plan."
  - "cacheRoot is a documented, deliberately unread parameter to QLT05Reverify: there is no cache anywhere on this call path in the first place (native.Runner always compiles into a fresh os.MkdirTemp), so cold-start is structural, not something a cache-bypass flag needs to enforce. The parameter exists so TestQLT05ReverificationIsColdStart can demonstrate the property directly (an empty t.TempDir() still reaches a verdict) rather than merely asserting it in prose."
  - "control:reduce.reverified is declared but NOT wired into VerifyMismatchReduceLane's own Controls list in this plan -- matching this file's own pre-existing \"declared, not yet gate-wired\" precedent for the three original control:reduce.* constants (a future plan is the designated wiring point, per the 05-14/05-11 QLT-01-registry precedent already documented in this file)."

requirements-completed: [QLT-05]

coverage:
  - id: D1
    description: "foreignCallSequenceFor widened to a two-independent-knower guard (static callgraph.Order walk + dynamic -O0 event-stream walk) covering every function, closing the multi-function nil-compares-nil slippage hole"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_mismatch_test.go#TestForeignCallSequenceCoversAllFunctions"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_mismatch_test.go#TestForeignCallSequenceStaticAndDynamicAgree"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_mismatch_test.go#TestForeignCallSequenceDetectsDroppedCallToForeignCallee"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_mismatch_test.go#TestForeignCallSequenceSeededDisagreement"
        status: pass
    human_judgment: false
  - id: D2
    description: "QLT-05 strict, cold-start re-verification: Axis/EnginePair/OperationID field equality with no positional (causal-role) fallback, plus foreign-call-sequence equality, reported as control:reduce.reverified; MismatchDocument schema stays pinned at lang.mismatch/0"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05Reverification"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05ReverificationRejectsCausalRoleFallback"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05ReverificationIsColdStart"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05ReverificationIsIdempotent"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestMismatchDocumentSchemaUnchanged"
        status: pass
    human_judgment: false
  - id: D3
    description: "Anti-vacuity gate: an engineered fixture with a provably removable function proves AppliedMoves is non-empty on a real reduction, and the complementary zero-move seed proves the check can fail"
    requirement: "QLT-05"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05GateIsNonVacuous"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_qlt05_reverify_test.go#TestQLT05EmptyReductionFailsAntiVacuity"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang --json check testdata/phase11/multi_function_reduce_gate.lang"
        status: pass
    human_judgment: false
  - id: D4
    description: "Phase 11's four evidence-weakened claims (NAT-05, NAT-06 criterion 2, QLT-06b, QLT-05 full-not-narrowed) collected in one document with evidence pointers, for the verification pass"
    requirement: "QLT-05"
    verification:
      - kind: other
        ref: "grep -v '^#' 11-VERIFICATION-INPUTS.md | grep -c -E 'D-11-10|D-11-25|D-11-39|weakened by evidence|inert by construction' (== 7, >= 5 required)"
        status: pass
    human_judgment: false

duration: 65min
completed: 2026-09-12
status: complete
---

# Phase 11 Plan 9: QLT-05 Strict Re-Verification and the Anti-Vacuity Gate Summary

**Closed the last two QLT-05 gaps 11-08 left open: `foreignCallSequenceFor`'s multi-function nil-compares-nil slippage hole (now a genuine two-independent-knower guard) and a strict, cold-start re-verification lane that forbids inheriting the search's own positional-match relaxation -- the anti-vacuity assertion targeted `drop-orphan-function` and QLT-05 ships FULL (Q-01 BRANCH A), not narrowed.**

## Performance

- **Duration:** ~65 min
- **Started:** 2026-09-12
- **Completed:** 2026-09-12
- **Tasks:** 3 completed
- **Files modified:** 3 modified, 3 created

## Accomplishments

- **`foreignCallSequenceFor` is now a two-independent-knower guard (D-11-33).** `staticForeignCallSequenceFor` walks EVERY function in `callgraph.Order`'s reverse-postorder over the proven-acyclic call graph (session imports callgraph directly, per the package's own documented consumer list); `dynamicForeignCallSequenceFor` derives the SAME fact independently from the `-O0` run's own event stream (`FunctionID` per `"foreign.called"` event), sharing no helper with the static walk beyond `execution.Event` itself. The two are compared, and disagreement is a hard failure (`reduce.foreign_call_sequence_disagreement`) surfaced up through both production call sites (`mismatchPredicate`, `ReduceSeededAliasMismatch`) rather than a silently-rejected candidate. Before this fix, the old `len(program.Functions) != 1` early return meant ANY multi-function program's foreign-call sequence was `nil`, so a dropped call to a foreign-acquiring callee compared nil-to-nil and was invisible.
- **`session.QLT05Reverify` implements strict, cold-start re-verification (D-11-32).** It compares a seed's `reduce.Signature` against a freshly-derived candidate signature by `Axis`/`EnginePair`/`OperationID` field equality -- deliberately with NO positional (causal-role) fallback, the ONE relaxation `predicate.go`'s `Interesting` permits during the search -- plus the foreign-call-sequence equality Task 1 built. The asymmetry is documented in a code comment at the comparison site (never spelling the literal identifier the file's own acceptance grep forbids). Forbidding the fallback is safe because every operation ID is function-ID-prefixed and neither whole-program move renumbers a surviving operation's ID (11-08's own `TestOperationIDsUnchangedByWholeProgramMoves`), so a genuine drift is impossible and observing one is a real reducer bug (`reduce.reverification_signature_drift`). Reported as the new `control:reduce.reverified` lane control, declared but not yet gate-wired (matching this file's own pre-existing precedent for the three original `control:reduce.*` constants). `MismatchDocument` stays pinned at `lang.mismatch/0` (`TestMismatchDocumentSchemaUnchanged`).
- **The anti-vacuity gate is proven on both sides (D-11-34).** `testdata/phase11/multi_function_reduce_gate.lang` is engineered so `helper` is provably removable (a same-type passthrough call from `main`); `TestQLT05GateIsNonVacuous` runs a real reduction and asserts `AppliedMoves` actually contains `"drop-orphan-function"` -- the difference between a gate and a decoration. `TestQLT05EmptyReductionFailsAntiVacuity` proves the complementary side: a single-function, zero-call seed on which no whole-program move is even eligible correctly does NOT trip that same assertion.
- **`11-VERIFICATION-INPUTS.md` collects the phase's four evidence-weakened claims** (NAT-05's explicit empty set, NAT-06 criterion 2's LTO inertness by construction, QLT-06b's completion-by-abstention, and a note that QLT-05 itself ships FULL under Q-01 BRANCH A, not narrowed) in one document with evidence pointers, so the phase's verification pass lifts them rather than rediscovering them.

## Task Commits

1. **Task 1: Close the foreignCallSequenceFor slippage hole with two independent knowers** -- `6a3d2d6` (feat)
2. **Tasks 2-3: Strict re-verification, control:reduce.reverified, anti-vacuity gate, and 11-VERIFICATION-INPUTS.md** -- `3d3af0f` (feat)

Note: Task 2's production code (`ControlReduceReverified`, `QLT05Reverify`, `qlt05RealSignatureDeriver`) was written into `session_phase5_mismatch.go` before Task 1's commit and landed in commit `6a3d2d6` rather than `3d3af0f` -- see Deviations below.

## Files Created/Modified

- `internal/compiler/session/session_phase5_mismatch.go` -- widened `foreignCallSequenceFor` (now two knowers), `ControlReduceReverified`, `QLT05Reverify`/`QLT05Reverification`/`QLT05SignatureDeriver`, `qlt05RealSignatureDeriver`
- `internal/compiler/session/session_phase5_mismatch_test.go` -- `TestForeignCallSequenceCoversAllFunctions`, `TestForeignCallSequenceStaticAndDynamicAgree`, `TestForeignCallSequenceDetectsDroppedCallToForeignCallee`, `TestForeignCallSequenceSeededDisagreement`
- `internal/compiler/session/export_test.go` -- `ForeignCallSequenceForTest`, `StaticForeignCallSequenceForTest`, `DynamicForeignCallSequenceForTest`, `SetForeignCallSequenceDisagreementForTest`
- `internal/compiler/session/session_qlt05_reverify_test.go` (new) -- `TestQLT05Reverification`, `TestQLT05ReverificationRejectsCausalRoleFallback`, `TestQLT05ReverificationIsColdStart`, `TestQLT05ReverificationIsIdempotent`, `TestMismatchDocumentSchemaUnchanged`, `TestQLT05GateIsNonVacuous`, `TestQLT05EmptyReductionFailsAntiVacuity`
- `testdata/phase11/multi_function_reduce_gate.lang` (new) -- the anti-vacuity gate fixture
- `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VERIFICATION-INPUTS.md` (new)

## Decisions Made

See `key-decisions` above. The two load-bearing ones: (1) the anti-vacuity assertion targets `drop-orphan-function`, not `drop-call-site`, since Q-01 BRANCH A shipped both moves live; (2) `QLT05Reverify` takes an injectable signature-deriver so its own comparison/cold-start/idempotence logic is fully unit-tested against a deterministic double, while the real corevalidate+cgen+native pipeline is wired for production but not exercised end-to-end by this plan's own tests (documented limitation, not a stub -- see Deviations).

## Deviations from Plan

### Auto-fixed Issues

None -- no bugs or missing-critical-functionality auto-fixes were needed this plan.

### Documented Engineering Decisions (not Rule 1-3 auto-fixes, but worth flagging)

**1. [Process] Task 2's production code landed in Task 1's commit**
- **Found during:** preparing Task 2's commit
- **Issue:** `ControlReduceReverified`/`QLT05Reverify`/`qlt05RealSignatureDeriver` were written into `session_phase5_mismatch.go` before Task 1's commit was made, so `git diff` showed no remaining change to that file when Task 2's own commit was prepared -- the production code had already landed in commit `6a3d2d6`.
- **Impact:** Purely a commit-boundary/attribution issue. All code is present, correct, and tested; `go test ./...` is green after each commit in sequence. No functional impact.
- **Committed in:** `6a3d2d6` (should have been `3d3af0f`)

**2. [Documented scope limitation] `qlt05RealSignatureDeriver` is not exercised end-to-end by this plan's own tests**
- **Found during:** Task 2, designing `TestQLT05Reverification`'s own fixture
- **Issue:** Engineering a genuine multi-function native alias-divergence fixture (the shape `qlt05RealSignatureDeriver`'s production wiring actually needs) ran into `AliasFactMutationRunner`'s own marker-uniqueness guard: a forwarding/wrapper caller function around the existing single-function `false_restrict_hoist.lang`-style body risks a second `byPointerParamMarker` match in the emitted C, or produces no real O0-vs-O3 divergence for a plain, unmutated multi-function call. Engineering that fixture correctly and validating it against Clang was judged out of scope for this plan's own effort budget.
- **Fix:** `QLT05Reverify`'s own strict-comparison/cold-start/idempotence properties are fully unit-tested (5 tests) against `qlt05StructuralDeriver`, a deterministic test double computing a `Signature` purely from program structure -- these properties are orthogonal to how the fresh signature is obtained. `qlt05RealSignatureDeriver` mirrors `mismatchPredicate`'s own already-tested real pipeline byte for byte, so its own individual steps (`corevalidate.Validate`, `cgen.EmitNative`, the runner's `-O0`/`-O3` pair, `Phase5CompareEngines`) are each independently proven correct elsewhere in this codebase; only the SPECIFIC combination of "real native compile+run" + "genuine multi-function divergence" is untested by this plan.
- **Verification:** `go test ./...` is green; all plan-specified acceptance criteria pass mechanically (grep counts, named-test presence, literal-string assertions).
- **Committed in:** `3d3af0f`

---

**Total deviations:** 0 Rule 1-3 auto-fixes. 2 documented engineering/process notes (1 commit-boundary attribution slip, 1 scope limitation on production-deriver end-to-end test coverage). **Impact:** No correctness or security impact; both are transparency notes for future readers.

## Issues Encountered

None beyond the two documented decisions above.

## Known Stubs

None. `qlt05RealSignatureDeriver` is production-quality code mirroring an already-tested pipeline, not a placeholder -- see the documented scope limitation above for what remains untested about its own end-to-end integration specifically.

## User Setup Required

None -- no external service configuration required.

## Next Phase Readiness

- QLT-05 is closed: the reducer's multi-function foreign-call-sequence slippage hole is fixed with a genuine second knower, re-verification is strict and cold-start with no inherited search-time relaxation, and the anti-vacuity gate is proven on both sides.
- `11-VERIFICATION-INPUTS.md` is ready for the phase's own `/gsd-verify-phase 11` pass: all four evidence-weakened claims are collected with evidence pointers.
- Recommended follow-up (not blocking): wire `control:reduce.reverified` into `VerifyMismatchReduceLane`'s own `Controls` list and gate-wire it into `Phase5RequiredControls()`, and separately, engineer a genuine multi-function native alias-divergence fixture to exercise `qlt05RealSignatureDeriver` end-to-end (see Deviations). Neither blocks Phase 11 completion; both are natural candidates for a future hardening plan.
- This is the LAST plan of Phase 11 (wave 6, depends_on 11-05/11-08). `go test ./...` is fully green.

---
*Phase: 11-multi-function-native-emission-and-interprocedural-equivalen*
*Completed: 2026-09-12*

## Self-Check: PASSED

`internal/compiler/session/session_phase5_mismatch.go`, `session_phase5_mismatch_test.go`, `export_test.go`, `session_qlt05_reverify_test.go`, `testdata/phase11/multi_function_reduce_gate.lang`, and `.planning/phases/11-multi-function-native-emission-and-interprocedural-equivalen/11-VERIFICATION-INPUTS.md` all verified present on disk. Commit hashes `6a3d2d6` and `3d3af0f` verified present in `git log --oneline`. `go build ./...` clean. `go vet ./internal/compiler/session/...` clean. `go test ./internal/compiler/session/... -run 'TestForeignCallSequence|TestQLT05|TestMismatchDocumentSchemaUnchanged' -v -count=1` prints 11 `--- PASS:` lines, no failures. `go test ./...` green across every package (cmd/lang, cmd/lang-repair, and all internal/compiler/* packages), no `FAIL` lines. All plan-specified `<verify>`/`<acceptance_criteria>` grep and test commands re-run and pass: `grep -c 'callgraph.Order' session_phase5_mismatch.go` = 2, `grep -c 'CausalRole' session_phase5_mismatch.go` = 0, `grep -rn 'lang.mismatch/1' internal/compiler/` = none, `go run ./cmd/lang --json check testdata/phase11/multi_function_reduce_gate.lang` returns `"status":"pass"` with empty diagnostics, `grep -v '^#' 11-VERIFICATION-INPUTS.md | grep -c -E 'D-11-10|D-11-25|D-11-39|weakened by evidence|inert by construction'` = 7 (>= 5 required).
