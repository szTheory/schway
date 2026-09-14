---
phase: 10-trusted-interprocedural-oracle
plan: 07
subsystem: pathoracle
tags: [composition-depth, bidirectional-gate, corevalidate-defect, mutation-testing, qlt-04]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-03's MaxCompositionDepth placeholder (8) and composeCall/composeCarriesOwnLoan/composeLinearizeCaller machinery, whose declared final depth and bidirectional gate this plan supplies; plan 10-06's mid-phase gate verdict (DID NOT FIRE, 0.19x), which cleared this plan to execute at full declared scope (depth 3, not the contingency-cut depth 2)"
provides:
  - "MaxCompositionDepth's final declared value (3) with a rationale comment stating the adopted definition (D-10-45), the depth-2 necessity floor citing D-09-49 Q2 (D-10-46), the sufficiency-margin reason for 3 (D-10-46), the missing-fixture finding (D-10-47), and the directional contrast with MaxPaths (D-10-12)"
  - "The product-space table (ParameterContract.Mode's D-07-01 named exclusion, ReturnContract.Mode's 3 values, LoanEndpoint.Kind's 2, verdict's 2) stated in shipped source on both the constant's own comment and the gate test's doc comment (D-10-49)"
  - "The depth-3 fixture pair (relay_depth3_accept.lang, relay_depth3_refuse.lang), the first depth-3 corpus members in this tree, varying which hop carries the borrow (D-10-48)"
  - "TestCompositionDepthCorpusReachesDeclaredBound (D-10-50): a bidirectional gate that fails if the corpus falls short of the declared depth AND (observed to fire via a simulated raise) if the constant is raised without extending the corpus, computed via a third, independent re-implementation of composition depth operating purely on core.LinearOperation"
  - "A previously-undocumented, empirically-confirmed finding: corevalidate.peerDeriveOriginFacts has no core.OpCall case, so any function declaring a borrow-returning PublicOrigin sourced from forwarding a callee's result is refused as not-Callable by corevalidate independently of check -- affecting relay_depth2_refuse.lang too, never observed there because check's own refusal fires first"
affects: [10-08, 11]

# Actuals (#2632)
actuals:
  tokens: 6800
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A THIRD independent re-implementation of D-10-45's composition-depth definition (functionCarryDepth, session_composition_depth_test.go), deliberately never calling into package pathoracle, so the gate and the machinery it verifies cannot share a bug"
    - "A gate's evidence fixture need not itself be end-to-end admitted: pathoracle's own composition depth is a property of the checked core.Program structure, so a check-refused fixture (relay_depth3_refuse.lang) still legitimately supplies the corpus-walk's depth-3 evidence"
    - "Bidirectional-gate logic factored into one reused function (compositionDepthGate) called with two different requirement values (the real declared constant, and a hypothetically-raised local value), so direction 2 is proven by the SAME logic direction 1 uses, not two independently-written assertions that could drift apart"

key-files:
  created:
    - testdata/phase10/relay_depth3_accept.lang
    - testdata/phase10/relay_depth3_refuse.lang
    - internal/compiler/session/session_composition_depth_test.go
  modified:
    - internal/compiler/pathoracle/pathoracle_compose.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - .planning/phases/10-trusted-interprocedural-oracle/deferred-items.md

key-decisions:
  - "The literal ideal accept fixture (a genuine 3-hop DECLARED-BORROW carry, forwarded through every hop, admitted end-to-end) is impossible today: corevalidate's own peerDeriveOriginFacts has no core.OpCall case, so it refuses ANY borrow-returning function whose return is sourced from forwarding a call's result, independent of check and independent of the caller's own behavior. Per Task 1's own escape hatch, this was reported as a finding (deferred-items.md) rather than silently falling back to an owned pass-through."
  - "relay_depth3_accept.lang uses the nearest expressible shape instead: a genuine borrow (not owned pass-through) created and used across one call boundary, then taken, inside an overall depth-3 call chain -- passing the full CLI, but its OWN loan crosses only 1 hop rather than 3."
  - "relay_depth3_refuse.lang carries the ideal shape instead: a declared borrow forwarded through BOTH middle hops (f2 and f1), refused via check.interprocedural_loan_liveness (caller's own conflicting take, generalizing relay_depth2_refuse.lang's exact pattern one level deeper). Its checked-but-refused core.Program is what supplies TestCompositionDepthCorpusReachesDeclaredBound's own depth-3 evidence, since pathoracle's own definition of composition depth is check/corevalidate-verdict-agnostic by design."
  - "TestCompositionDepthCorpusReachesDeclaredBound treats a fixture as corpus-eligible once session.CheckFile produces a structurally real core.Program (parse succeeded, check.Program built), regardless of check's own admission diagnostics -- deliberately broader than the sibling peer-divergence gate's own check-admitted-only walk, because this gate's subject (pathoracle's structural composition depth) is a property of the checked core, not of what check/corevalidate separately conclude."

patterns-established:
  - "Bidirectional gate logic factored into one reusable function called twice with different requirement values, so direction 2 is proven by reusing direction 1's own code path rather than reasoned about separately."

requirements-completed: [QLT-04]

coverage:
  - id: D1
    description: "A depth-3 fixture pair exists, varying which hop carries the borrow, with empirically-verified verdicts written into self-explaining headers"
    requirement: QLT-04
    verification:
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase10/relay_depth3_accept.lang"
        status: pass
      - kind: integration
        ref: "go run ./cmd/lang --json check testdata/phase10/relay_depth3_refuse.lang"
        status: pass
      - kind: unit
        ref: "go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v"
        status: pass
    human_judgment: true
    rationale: "The ideal shape (a single loan literally crossing all 3 hops while being end-to-end admitted) is not achievable given the newly-discovered corevalidate defect documented in deferred-items.md; a reviewer should confirm the nearest-expressible-shape substitution and its documentation are an honest, sufficient discharge of the task's own escape hatch, not a quiet weakening of the corpus."
  - id: D2
    description: "The composition depth is declared at 3 with a rationale a reviewer can check, and the product space's one collapsed dimension is named in source"
    requirement: QLT-04
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/pathoracle/... -v"
        status: pass
      - kind: other
        ref: "grep -c 'D-07-01' internal/compiler/pathoracle/pathoracle_compose.go internal/compiler/pathoracle/pathoracle_test.go"
        status: pass
    human_judgment: false
  - id: D3
    description: "The declared composition depth is bidirectionally gated against the corpus, with both directions observed to fire and no order dependence"
    requirement: QLT-04
    verification:
      - kind: unit
        ref: "session_composition_depth_test.go#TestCompositionDepthCorpusReachesDeclaredBound"
        status: pass
      - kind: other
        ref: "go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -count=2 -shuffle=on"
        status: pass
    human_judgment: false

duration: 65min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 7: QLT-04's Declared Composition Depth Summary

Declared `pathoracle.MaxCompositionDepth = 3` with a full rationale, built the tree's first depth-3 fixture pair, and shipped a bidirectional gate proven to fail in both directions -- discovering and documenting, rather than silently working around, a previously-unobserved corevalidate defect that makes a literal end-to-end-admitted 3-hop declared-borrow carry currently unrepresentable.

## Performance

- **Duration:** 65 min
- **Started:** 2026-09-11
- **Completed:** 2026-09-11
- **Tasks:** 3 completed
- **Files modified:** 6 (3 created, 2 modified in `pathoracle`, 1 modified deferred-items.md)

## Accomplishments

- `MaxCompositionDepth` set to its final declared value, 3, with a rationale comment stating the adopted definition (D-10-45, matching `corevalidate_peer_liveness.go:64-77`'s own "chain of d+1 functions" prose), why not less than 2 (D-10-46, citing D-09-49 Q2), why 3 and not merely 2 (D-10-46, CBMC's unwind-plus-assertion framing), what forced it (D-10-47, the missing depth-3 fixture), and the directional contrast with `MaxPaths` (D-10-12: `MaxPaths` sits above its real reachable maximum so it never fires; `MaxCompositionDepth` sits AT the declared depth so it genuinely does).
- The product-space table (D-10-49) -- `ParameterContract.Mode`'s cardinality-1 named exclusion (D-07-01), `ReturnContract.Mode`'s 3 values, `LoanEndpoint.Kind`'s 2, verdict's 2 -- is stated in shipped source in both the constant's own comment and `pathoracle_test.go`'s doc comment, never only in a planning document.
- `testdata/phase10/relay_depth3_accept.lang` and `relay_depth3_refuse.lang` are the tree's first depth-3 fixtures (4 functions, 3 `OpCall` hops each). `relay_depth3_refuse.lang` carries a declared borrow through BOTH middle hops (`f2` and `f1`), refused via `check.interprocedural_loan_liveness` naming `f1` -- generalizing `relay_depth2_refuse.lang`'s single-middle-hop pattern one level deeper, genuinely exercising "a borrow surviving two hops" for the first time in this tree. `relay_depth3_accept.lang` carries a genuine borrow (not `relay_depth2_accept.lang`'s owned pass-through) across one call boundary before being taken, inside an overall depth-3 call chain, passing the full CLI cleanly.
- **Finding, reported per Task 1's own escape hatch rather than silently substituted:** the literal ideal accept fixture -- the SAME declared-borrow-forwarded-through-every-hop shape as the refuse twin, but with no conflicting caller-side use so the full CLI would admit it -- does not exist in this compiler today. `check.Program` admits that shape with zero diagnostics, but `corevalidate.Validate` (consulted next in `session.CheckCommandFile`'s fixed precedence) independently refuses it with `core.callee_not_callable`, because `corevalidate.go`'s `peerDeriveOriginFacts` has no `core.OpCall` case at all -- it only ever extends its `derived` set through `OpBorrowShared`/`OpBorrowExclusive`/`OpMove`/`OpCopy`, so a call's own result never enters it. Confirmed this is NOT new to this plan: `relay_depth2_refuse.lang` (Phase 08, unmodified) is refused by `corevalidate.Validate` for the identical independent reason once its own conflicting `take` is removed -- it has simply never been observable, because `check`'s own refusal fires first. Documented in full in `relay_depth3_accept.lang`'s own header and in `deferred-items.md`, structurally analogous to D-10-28 (a second detector missing an `OpCall` case), not fixed here since `corevalidate.go` is outside this plan's `files_modified`.
- `TestCompositionDepthCorpusReachesDeclaredBound` (`internal/compiler/session/session_composition_depth_test.go`) is D-10-50's bidirectional gate. `functionCarryDepth` is a third, independent re-implementation of D-10-45's own definition -- never a call into package `pathoracle` -- operating purely on `core.LinearOperation`, so the gate and the composition machinery it verifies cannot share a bug. Direction 1 walks `testdata/` and finds `relay_depth3_refuse.lang`'s own checked-but-refused `core.Program` genuinely reaching depth 3 (pathoracle's own composition-depth definition is check/corevalidate-verdict-agnostic by design, so a check-refused fixture still legitimately supplies this evidence). Direction 2 reuses the IDENTICAL `compositionDepthGate` logic against a hypothetically-raised local requirement (never editing the production constant) and asserts it reports a shortfall -- proving direction 2 has been observed to fire, not merely reasoned about. Passes under `-count=2 -shuffle=on` (no package-level state).

## Task Commits

Each task was committed atomically:

1. **Task 1: The depth-3 fixture pair, varying which hop carries the borrow** - `a384b7c` (test)
2. **Task 2: Declare the composition depth — the constant and its rationale** - `a032a7a` (feat)
3. **Task 3: The bidirectional gate — fail if the corpus falls short AND if the constant outruns it** - `81e109a` (test)

**Plan metadata:** (this commit) `docs(10-07): complete composition-depth declaration plan`

## Files Created/Modified

- `testdata/phase10/relay_depth3_accept.lang` — the nearest-expressible accepting depth-3 fixture, with a header documenting the empirically-confirmed corevalidate finding in full.
- `testdata/phase10/relay_depth3_refuse.lang` — the ideal depth-3 shape (declared borrow through both middle hops), refused via `check.interprocedural_loan_liveness`.
- `internal/compiler/pathoracle/pathoracle_compose.go` — `MaxCompositionDepth` set to 3, full rationale comment, product-space table.
- `internal/compiler/pathoracle/pathoracle_test.go` — product-space table cross-referenced from `TestCompositionDepthCapRejects`'s own doc comment, citing D-07-01.
- `internal/compiler/session/session_composition_depth_test.go` — new file: `functionCarryDepth`, `maxCompositionDepthAcrossCorpus`, `compositionDepthGate`, `TestCompositionDepthCorpusReachesDeclaredBound`.
- `.planning/phases/10-trusted-interprocedural-oracle/deferred-items.md` — new entry recording the corevalidate `peerDeriveOriginFacts` finding in full.

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Stale core.go line citations in the product-space table**
- **Found during:** Task 2, writing the product-space table comment
- **Issue:** The plan's own action text cited `core.go:191-198` for `ParameterContract.Mode` and `core.go:236-284` for `ReturnContract.Mode`; the actual current file has these types at `245-256` and `270-285` respectively (mirroring D-10-40's own already-recorded stale-citation precedent for a different mechanism).
- **Fix:** Cited the verified current line numbers instead, with an inline note explaining the correction (mirroring D-10-40's own convention for handling a stale citation).
- **Files modified:** `internal/compiler/pathoracle/pathoracle_compose.go`
- **Verification:** `grep -n "type ParameterContract" -A 12 internal/compiler/core/core.go` and `grep -n "type ReturnContract" -A 15 internal/compiler/core/core.go` confirm the corrected line ranges.
- **Committed in:** `a032a7a`

---

**Total deviations:** 1 auto-fixed (Rule 1). No architectural changes; the corevalidate finding (see below) was reported per the task's own explicit escape-hatch instruction, not treated as an auto-fixable deviation, since fixing `corevalidate.go` is outside this plan's file scope and carries its own independent blast radius.

**Impact on plan:** Necessary line-citation correction; no scope creep.

## Issues Encountered

**Empirically discovered, not an "issue" caused by this plan's own changes:** `corevalidate.peerDeriveOriginFacts` (`internal/compiler/corevalidate/corevalidate.go:2324-2354`) has no `core.OpCall` case, making any borrow-returning function whose return derives from forwarding a call's result unconditionally refused as not-Callable by corevalidate, independent of check and independent of the caller's own behavior. This is a genuine, previously-undocumented pre-existing defect (confirmed to also affect `relay_depth2_refuse.lang`, unmodified since Phase 08, once its own conflicting `take` is removed) rather than something introduced by this plan. Resolved per the plan's own explicit escape hatch: reported in full in both `relay_depth3_accept.lang`'s header and `deferred-items.md`, with the nearest expressible accepting shape substituted for the fixture named `relay_depth3_accept.lang`, and the ideal shape carried instead by `relay_depth3_refuse.lang` (whose checked-but-refused core structure supplies the bidirectional gate's own depth-3 evidence). Not fixed: `corevalidate.go` is outside this plan's `files_modified`, and the fix is a genuine, independent semantic change (mirroring D-10-28's own precedent for a sibling peer) that a future plan should take up with its own mutation-kill test.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- QLT-04 is closed: composition depth is declared at 3, stated in shipped source rather than implied by the corpus, with a bidirectional gate proven to fail in both directions.
- Plan 10-08's criterion-4 differential can build on this depth-3 corpus and the composition machinery unchanged.
- The newly-discovered `corevalidate.peerDeriveOriginFacts` gap (no `core.OpCall` case) is named, unassigned debt, structurally analogous to D-10-28 — a candidate for the same future landing phase (Phase 11 or M003) that eventually takes up D-10-28, since both are "corevalidate's peer missing an OpCall case." Not formally registered as a `D-10-NN` row in `PHASE-10-DEBT.md` by this executor (debt-ID assignment is normally a planning-time act); a future planning pass should promote `deferred-items.md`'s own entry into a numbered debt row if it is not resolved first.
- Ready for plan 10-08.

## Self-Check: PASSED

- `testdata/phase10/relay_depth3_accept.lang` — FOUND
- `testdata/phase10/relay_depth3_refuse.lang` — FOUND
- `internal/compiler/session/session_composition_depth_test.go` — FOUND
- Commit `a384b7c` — FOUND (`git log --oneline --all`)
- Commit `a032a7a` — FOUND
- Commit `81e109a` — FOUND
- `go run ./cmd/lang --json check testdata/phase10/relay_depth3_accept.lang` — exit 0, status "pass", empty diagnostics
- `go run ./cmd/lang --json check testdata/phase10/relay_depth3_refuse.lang` — exit non-zero, status "invalid", one diagnostic (`check.interprocedural_loan_liveness`)
- `grep -c 'QLT-04' testdata/phase10/relay_depth3_accept.lang testdata/phase10/relay_depth3_refuse.lang` — 2 and 2
- `go test ./internal/compiler/pathoracle/... -v` — PASS (13/13 tests)
- `grep -c 'D-07-01' internal/compiler/pathoracle/pathoracle_compose.go internal/compiler/pathoracle/pathoracle_test.go` — 1 and 1
- `go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -v` — PASS (both subtests)
- `go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -count=2 -shuffle=on` — PASS
- `go test ./internal/compiler/session/... ./internal/compiler/pathoracle/... ./internal/compiler/corevalidate/... ./internal/compiler/check/...` — PASS
- `go test ./internal/compiler/...` (full suite, pre-existing before this plan) — PASS, 0 failures
- `go build ./... && go vet ./...` — PASS

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*
