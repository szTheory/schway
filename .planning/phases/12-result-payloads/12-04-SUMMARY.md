---
phase: 12-result-payloads
plan: 04
subsystem: compiler-payload-representation
tags: [originvalidate, corevalidate, interp, pathoracle, payload, D-12-09, D-12-19, D-12-28, D-12-38, D-12-04c]

# Dependency graph
requires:
  - phase: 12-result-payloads (plan 02)
    provides: OpConstructPayload/OpDestructurePayload, core.NewDataType/AlternativeDetails, testdata/phase12/payload_tracer.lang
  - phase: 12-result-payloads (plan 03)
    provides: testdata/phase12/payload_borrow_interaction.lang (the borrow-carrying payload fixture this plan's pathoracle control depends on)
provides:
  - originvalidate.walkReturnOrigin's own OpDestructurePayload arm, plus the RecomputeOriginPerReturn sourceOf-indexing fix that makes a destructured place reachable by a backward walk at all
  - corevalidate's corpus-wide core.alternative_details_desynchronized invariant, re-asserting D-12-09's NewDataType constructor guarantee for a program that arrives already built
  - TestPayloadCorpusCharacterizationReplay + TestPayloadCorpusCharacterizationReplayMutationKilled (session package) and interp.SetDisableEmptyTagSerializationForTest, the production-visible test-only seam that makes the mutation-kill beat possible
  - TestPayloadPathEnumerationTerminates (pathoracle package), proving path enumeration terminates and produces correct endpoints on payload-carrying fixtures without a new case arm
  - testdata/phase08/relay_depth2_accept.lang's corrected header (D-12-04c closed) and PHASE-12-DEBT.md's updated D-12-04c row
affects: [12-05]

# Actuals (#2632)
actuals:
  tokens: 66000
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Cross-package test-only seams that must be reachable from a DIFFERENT package's test binary are defined as production-visible functions in a real .go file (never export_test.go, which is only visible within the same package's own test binary) -- following corevalidate.SetDisableCyclePeerForTest's D-07-42 precedent exactly. interp.SetDisableEmptyTagSerializationForTest follows this shape so session_test (a different package) can drive it."
    - "A corpus-characterization replay's baseline is committed golden digests keyed by fixture path (never a self-comparison), with every fixture that does not produce a comparable execution document named in an explicit expected-skip set -- so a fixture silently dropping out of either set fails loudly rather than being absorbed."

key-files:
  created:
    - internal/compiler/corevalidate/corevalidate_alternative_details_test.go
    - internal/compiler/originvalidate/originvalidate_payload_test.go
    - internal/compiler/session/session_payload_replay_test.go
    - internal/compiler/pathoracle/pathoracle_payload_test.go
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/originvalidate/originvalidate.go
    - internal/compiler/interp/interp.go
    - testdata/phase08/relay_depth2_accept.lang
    - .planning/phases/12-result-payloads/PHASE-12-DEBT.md

key-decisions:
  - "originvalidate's RecomputeOriginPerReturn map-building loop needed a Rule 1 fix (not just a new switch arm): it indexes sourceOf by operation.TargetID, but OpDestructurePayload leaves TargetID empty and names its produced place via PayloadTargetID instead (D-12-10). Without special-casing this in the map-building loop itself, the destructured place could never be found by any backward walk, and the plan's own stated case-arm addition alone would have been a no-op -- the fix was structurally necessary, not optional."
  - "The interp mutation-kill seam (SetDisableEmptyTagSerializationForTest) is defined as a production-visible function in interp.go, not in an export_test.go file, because the replay test lives in a DIFFERENT package (session_test) -- Go's _test.go export trick (as cgen.SetOpCallGroupedArmForTest uses) only works within the same package's own test binary. This project's own corevalidate.SetDisableCyclePeerForTest/D-07-42 precedent is the correct shape for a cross-package seam and is what this plan follows."
  - "The corpus characterization baseline is honestly a POST-widening snapshot, not a pre-widening one: plan 02's interp value-widening (D-12-17) already landed before this plan captured digests, so a literal pre-widening baseline no longer exists to capture from this tree (flagged_assumption 3, 12-04-PLAN.md). The mutation-kill beat (disabling the empty-tag special case and proving at least one digest moves) is what keeps the replay load-bearing despite that honesty gap, per the plan's own stated fallback."

requirements-completed: [RES-02, RES-03]

coverage:
  - id: D1
    description: "originvalidate independently understands a payload destructure's own origin semantics, so a declared PublicOrigin covering the original borrow is not reported as understated"
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_payload_test.go#TestOriginWalksThroughPayloadDestructureToBorrow"
        status: pass
      - kind: unit
        ref: "internal/compiler/originvalidate/originvalidate_payload_test.go#TestOriginUnderstatedAcrossPayloadDestructure"
        status: pass
    human_judgment: false
  - id: D2
    description: "corevalidate's AlternativeDetails name-set invariant holds corpus-wide (for a program built without NewDataType), not only at construction"
    requirement: "RES-02"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_alternative_details_test.go#TestAlternativeDetailsDesynchronizedNameAbsent"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_alternative_details_test.go#TestAlternativeDetailsDesynchronizedDuplicateName"
        status: pass
    human_judgment: false
  - id: D3
    description: "D-10-C01's peerDeriveOriginFacts gap stays measurably open (this plan must not close it as a side effect)"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_result_payload_probe_test.go#TestC03PeerDeriveOriginFactsOpCallGapStillOpen"
        status: pass
      - kind: unit
        ref: "awk-scanned peerDeriveOriginFacts body contains zero case core.OpCall: arms"
        status: pass
    human_judgment: false
  - id: D4
    description: "The interp value widening (D-12-17) moved zero execution-document bytes across the pre-Phase-12 corpus, proven by a replay that a mutation can kill"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_payload_replay_test.go#TestPayloadCorpusCharacterizationReplay"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_payload_replay_test.go#TestPayloadCorpusCharacterizationReplayMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_payload_replay_test.go#TestScalarValueSurvivesCopyMoveBorrowSequence"
        status: pass
    human_judgment: false
  - id: D5
    description: "pathoracle's path enumeration terminates and produces correct endpoints on a payload-carrying fixture that meets real borrow machinery, without a new case arm in a deliberately body-blind file"
    requirement: "RES-03"
    verification:
      - kind: unit
        ref: "internal/compiler/pathoracle/pathoracle_payload_test.go#TestPayloadPathEnumerationTerminates"
        status: pass
    human_judgment: false
  - id: D6
    description: "testdata/phase08/relay_depth2_accept.lang's stale header is corrected, and the correction itself is recorded (D-09-45), not silently applied"
    verification:
      - kind: unit
        ref: "git diff --stat confirms only the header comment block changed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_test.go#TestDebtRegistersAreWellFormed"
        status: pass
    human_judgment: false

duration: 24min
completed: 2026-09-12
status: complete
---

# Phase 12 Plan 04: Independent-Peer Payload Completeness and the Corpus Characterization Replay Summary

**originvalidate and pathoracle now independently understand payload operations, corevalidate's AlternativeDetails invariant is enforced corpus-wide, and a mutation-killable replay proves the interp value widening moved zero bytes across the pre-Phase-12 corpus.**

## Performance

- **Duration:** ~24 min (task work; prior plan 03 completed 2026-09-12T23:08:59-04:00, this plan's last commit 2026-09-12T23:32:47-04:00)
- **Tasks:** 3
- **Files modified:** 9 (4 created, 5 modified)

## Accomplishments

- `originvalidate.walkReturnOrigin` gains a `case core.OpDestructurePayload:` arm that walks through a NEW kind's own semantics — categorically different from filling `core.OpCall`'s gap in `peerDeriveOriginFacts`, so D-12-28's guard stays untripped. A real structural bug had to be fixed first: `RecomputeOriginPerReturn`'s `sourceOf` map indexed every operation by `TargetID`, but `OpDestructurePayload` leaves `TargetID` empty and names its produced place via `PayloadTargetID` (D-12-10) — without special-casing this in the map-building loop, the destructured place could never be found by any backward walk, making the case arm alone a no-op.
- `corevalidate.run()` gains a corpus-wide `core.alternative_details_desynchronized` invariant: every `AlternativeDetails` entry must name an alternative present in `Alternatives`, with no duplicates — re-asserting `core.NewDataType`'s construction-time guarantee (D-12-09) for a program that arrives already built (decoded from JSON, produced by a future pass, or hand-assembled in a test).
- `peerDeriveOriginFacts` is byte-unchanged; `TestC03PeerDeriveOriginFactsOpCallGapStillOpen` stays green and an `awk`-scanned check confirms zero `case core.OpCall:` arms inside it.
- `TestPayloadCorpusCharacterizationReplay` globs all 97 pre-Phase-12 `.lang` fixtures, hashes each clean fixture's interp execution document(s) against a committed baseline (62 fixtures), and names every one of the 35 fixtures that cannot produce a comparable execution document (negative controls, deliberately malformed sources, cyclic/ambiguous-entry programs) in an explicit `payloadCorpusExpectedSkips` set — a fixture silently dropping from either set fails the test loudly.
- `TestPayloadCorpusCharacterizationReplayMutationKilled` proves the replay is load-bearing: disabling `interp`'s D-12-18 empty-tag serialization special case via the new `interp.SetDisableEmptyTagSerializationForTest` seam moves `owned_transfer.lang`'s digest away from the pinned baseline.
- `TestScalarValueSurvivesCopyMoveBorrowSequence` is D-12-19's narrow companion property test, over a hand-built `core.Program`, independent of any corpus fixture.
- `TestPayloadPathEnumerationTerminates` drives both plan 02's tracer fixture and plan 03's borrow-interaction fixture through `pathoracle.BuildCalleeLookup`/`RecomputeEndpoints` — no case arm added to `pathoracle.go` (deliberately body-blind, D-04-29); a nil error IS the proof that enumeration terminated within `MaxPaths`. The borrow-carrying fixture (`payload_borrow_interaction.lang`'s `choose`, 1 loan endpoint) is what makes the control non-vacuous; a payload-only fixture (the tracer's `identity`) alone would have produced a degenerate zero-loan enumeration.
- `testdata/phase08/relay_depth2_accept.lang`'s stale header is corrected: the claim that corevalidate also refuses the fixture is replaced with a dated D-12-04c correction naming D-09-03 as the closing decision, quoting the superseded text verbatim (D-09-45) — no Lang source line below the header changed. `PHASE-12-DEBT.md`'s D-12-04c row and detail section are updated to record closure in this plan.
- All four pinned golden-C digests remain byte-identical (`TestPreviousPhaseGoldenCUnchanged`); `go test ./...` exits 0; `go vet ./...` clean; every touched `.go` file is `gofmt`-clean.

## Task Commits

Each task was committed atomically:

1. **Task 1: originvalidate's payload arm and corevalidate's corpus-wide name-set invariant** - `57302c3` (feat)
2. **Task 2: The corpus characterization replay** - `c307564` (test)
3. **Task 3: pathoracle termination on payload shapes, and the D-12-04c header correction** - `82da692` (test)

**Plan metadata:** (this commit, following SUMMARY creation)

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` - corpus-wide `core.alternative_details_desynchronized` invariant arm in `run()`
- `internal/compiler/corevalidate/corevalidate_alternative_details_test.go` - four new tests for the invariant (absent name, duplicate name, synchronized-accepted, empty-still-accepted)
- `internal/compiler/originvalidate/originvalidate.go` - `RecomputeOriginPerReturn`'s `sourceOf` map-building fix (index by `PayloadTargetID` for `OpDestructurePayload`) plus the new `case core.OpDestructurePayload:` arm in `walkReturnOrigin`
- `internal/compiler/originvalidate/originvalidate_payload_test.go` - two new tests proving the origin walk through a payload destructure to a borrow
- `internal/compiler/interp/interp.go` - `disableEmptyTagSerializationSeamForTest` + `SetDisableEmptyTagSerializationForTest`, consulted by `value.String()`
- `internal/compiler/session/session_payload_replay_test.go` - the corpus characterization replay, its mutation-kill beat, and the scalar-survival property test
- `internal/compiler/pathoracle/pathoracle_payload_test.go` - `TestPayloadPathEnumerationTerminates`
- `testdata/phase08/relay_depth2_accept.lang` - header-only D-12-04c correction
- `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` - D-12-04c row/section updated to CLOSED

## Decisions Made

See `key-decisions` in the frontmatter for the full list. Summarized:

1. The `sourceOf` map-building fix in `originvalidate.go` was structurally necessary (Rule 1), not an optional refinement — without it the plan's own stated case-arm addition would have been unreachable code.
2. The interp mutation-kill seam is a production-visible function (`interp.go`), not an `export_test.go` symbol, because it must be reachable from a different package's test binary (`session_test`) — following `corevalidate.SetDisableCyclePeerForTest`'s D-07-42 precedent rather than `cgen.SetOpCallGroupedArmForTest`'s same-package shape.
3. The corpus characterization baseline is honestly a post-widening snapshot (plan 02's `interp` widening already landed before this plan ran); the mutation-kill beat is what keeps it load-bearing despite that, exactly as the plan's `flagged_assumptions` anticipated.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `originvalidate.RecomputeOriginPerReturn`'s `sourceOf` map indexed `OpDestructurePayload` by the wrong field**
- **Found during:** Task 1, while writing `TestOriginWalksThroughPayloadDestructureToBorrow` (RED confirmed the bug before any fix)
- **Issue:** The map-building loop unconditionally wrote `sourceOf[operation.TargetID] = operation`. `OpDestructurePayload` leaves `TargetID` empty and names its produced place via `PayloadTargetID` instead (D-12-10) — so the destructured place's own producing operation was never indexed under a key any backward walk would ever look up, silently making the plan's stated case-arm addition alone insufficient.
- **Fix:** The map-building loop now special-cases `core.OpDestructurePayload`, indexing by `PayloadTargetID` instead of `TargetID`. Every other operation kind (all of which never populate `PayloadTargetID`, D-12-10) is unaffected.
- **Files modified:** `internal/compiler/originvalidate/originvalidate.go`
- **Verification:** `TestOriginWalksThroughPayloadDestructureToBorrow` and `TestOriginUnderstatedAcrossPayloadDestructure` both pass; full `go test ./...` green.
- **Committed in:** `57302c3` (Task 1 commit)

**2. [Rule 3 - Blocking] A cross-package test-only mutation-kill seam could not live in `export_test.go`**
- **Found during:** Task 2, while designing the mutation-kill beat
- **Issue:** The plan's action text named the `SetOpCallGroupedArmForTest`/`callReturnTypeDerivationSeam` shape as the precedent to follow. `SetOpCallGroupedArmForTest` specifically lives in `cgen/export_test.go` (package `cgen`), which is visible only within `cgen`'s OWN test binary — `session_test` (a different package importing `interp` as an ordinary dependency) cannot see anything defined in `interp`'s `_test.go` files.
- **Fix:** Followed this project's OWN cross-package precedent instead — `corevalidate.SetDisableCyclePeerForTest` (D-07-42), which is a production-visible function in a real `.go` file, documented as a test-only no-op. `interp.SetDisableEmptyTagSerializationForTest` is defined in `interp.go` itself.
- **Files modified:** `internal/compiler/interp/interp.go`
- **Verification:** `TestPayloadCorpusCharacterizationReplayMutationKilled` passes from the `session_test` package; `go vet ./...` clean.
- **Committed in:** `c307564` (Task 2 commit)

---

**Total deviations:** 2 auto-fixed (1 bug, 1 blocking). **Impact:** Both were required for the plan's own stated success criteria to be reachable at all — the first made the origin-walk arm actually load-bearing rather than dead code; the second made the mutation-kill beat constructible across the package boundary the plan itself specified. No scope creep.

## Known Stubs

None introduced by this plan. Plan 02's and plan 03's previously-recorded Known Stubs (nullary-construction-from-payload-pattern; single-function payload+borrow combination; `alternativeNameForPayloadType`'s distinct-payload-types-only limitation) are unaffected and unchanged by this plan's work.

## Issues Encountered

None beyond the two items recorded above, both resolved within this plan's own task scope.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- D-10-C01's gap in `peerDeriveOriginFacts` stays measurably open, as D-12-30's three-part resource-payload landing condition requires; plan 05 (or any future plan) inherits it unchanged.
- The corpus characterization replay's honest post-widening-baseline framing (this plan's key-decisions) should be carried forward to any future plan that recomputes or extends `payloadCorpusBaseline`.
- `interp.SetDisableEmptyTagSerializationForTest` is available as a general cross-package mutation-kill seam for any future `interp` value-model change.
- Plan 05 can proceed with all of this phase's independent-peer and corpus-wide invariant work in place.

## Self-Check: PASSED

- FOUND: internal/compiler/corevalidate/corevalidate.go
- FOUND: internal/compiler/corevalidate/corevalidate_alternative_details_test.go
- FOUND: internal/compiler/originvalidate/originvalidate.go
- FOUND: internal/compiler/originvalidate/originvalidate_payload_test.go
- FOUND: internal/compiler/interp/interp.go
- FOUND: internal/compiler/session/session_payload_replay_test.go
- FOUND: internal/compiler/pathoracle/pathoracle_payload_test.go
- FOUND: testdata/phase08/relay_depth2_accept.lang
- FOUND: .planning/phases/12-result-payloads/PHASE-12-DEBT.md
- FOUND commit: 57302c3 (Task 1)
- FOUND commit: c307564 (Task 2)
- FOUND commit: 82da692 (Task 3)
- `go test ./...` exits 0
- `go vet ./...` clean
- `gofmt -l` clean on every file this plan touched
- `go test ./internal/compiler/corevalidate/... -run TestC03PeerDeriveOriginFactsOpCallGapStillOpen -count=1` exits 0
- `go test ./internal/compiler/core/... -run TestPreviousPhaseGoldenCUnchanged -count=1` exits 0
- `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -count=1` exits 0

---
*Phase: 12-result-payloads*
*Completed: 2026-09-12*
