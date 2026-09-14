---
phase: 13-agent-loop-for-interprocedural-defects
plan: 04
subsystem: testing
tags: [defect-injection, held-out-corpus, mutation-testing, ownership-checker, interprocedural]

requires:
  - phase: 13-agent-loop-for-interprocedural-defects (plan 01)
    provides: the shipped interprocedural loan-liveness repair, the D-13-09b callIsLastUse backward gate, and the derivation_interprocedural_loan_defect.lang swap-fixable shape this plan's InterproceduralLoanInjector reproduces mechanically
provides:
  - Three new phase-13 interprocedural defect injectors (InterproceduralLoanInjector, FallibleConsumeInjector, CallArgumentTypeInjector) with fail-closed marker guards and not-inert guard-disabled twins, registered in AllInjectors
  - Seven new testdata/phase13 fixtures: D-13-28's twin pair (heldout_shared_callee_twin_alpha/mirror.lang, both halves), the fallible-consume held-out/derivation pair, the call-argument-mismatch held-out/derivation pair, and the ambiguous-match held-out sibling
  - testdata/phase13/HELDOUT.sha256, the sealed digest manifest for all five heldout_* fixtures, committed BEFORE plan 13-05 writes any repair against them
  - A topology-triple (function count, call-edge count, root-to-detection hop distance) re-derivation, independent of the callgraph package, proving the held-out/derivation split varies call-graph shape rather than identifiers
  - An empirical finding on D-13-02a: no B1-shaped interprocedural diagnostic is constructible at this language maturity (see below)
affects: ["13-05", "13-06"]

actuals:
  tokens: 13135
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Marker-driven mechanical defect injection (D-06-25/D-13-24) extended to interprocedural defect classes: line-swap, try-strip, and Byte/Buffer-toggle mutations"
    - "Topology-triple corpus disjointness (function count, call-edge count, root-to-detection hop distance), re-derived independently in the test rather than read from a production package -- replaces byte-inequality as the held-out/derivation distinctness oracle"

key-files:
  created:
    - internal/compiler/session/session_phase13_injectors.go
    - internal/compiler/session/session_phase13_injectors_test.go
    - testdata/phase13/heldout_shared_callee_twin_alpha.lang
    - testdata/phase13/heldout_shared_callee_twin_mirror.lang
    - testdata/phase13/heldout_fallible_call_unconsumed.lang
    - testdata/phase13/heldout_call_argument_mismatch.lang
    - testdata/phase13/heldout_call_argument_ambiguous.lang
    - testdata/phase13/derivation_fallible_call_unconsumed.lang
    - testdata/phase13/derivation_call_argument_mismatch.lang
    - testdata/phase13/HELDOUT.sha256
  modified:
    - internal/compiler/session/session_phase6_injectors.go
    - internal/compiler/session/session_phase6_injectors_test.go

key-decisions:
  - "Hop distance is defined as root-to-DETECTION call-graph depth (main -> ... -> the function containing the marked statement), not detection-to-fix distance in the literal blame sense -- because D-13-28's twin pair deliberately keeps the true-fix function's identity an open question pending 13-05/13-06's repair rules and D-13-02a's resolution. This is the axis every fixture genuinely varies and the one a topology-blind blame rule could overfit; documented explicitly as a discretion call in code comments."
  - "The twin pair's two halves differ by WHICH caller (alpha vs. beta) carries the marker and the depth-2 relay chain, not by which of caller/callee the eventual repair blames -- that deeper discrimination is deferred to 13-05/13-06 once wrap_call_in_try's sibling repairs exist and a driver test can actually apply them and observe re-check-clean. This plan's own obligation (D-13-26.3: clean baseline, exactly one diagnostic mutated) is met by both halves independently."
  - "TestCorpusTopologyDisjoint is scoped to the two classes with both a held-out and derivation member (fallible_call_unconsumed, call_argument_type_mismatch) -- the twin pair has no derivation counterpart in this plan's file list (D-13-28 asks only for both held-out halves), so it is covered by TestHeldoutBaselinesAreFailClosed's D-13-26.3 property instead, not the pairwise disjointness comparison."
  - "D-13-02a: empirically, no B1-shaped defect (blame belonging to the callee's own declared-fact violation, independent of any caller) is constructible in this language at this maturity -- see finding below. Recorded here per 13-CONTEXT.md's requirement that 13-06 be handed a grounded answer, not a guess."

requirements-completed: [DX-07]

coverage:
  - id: D1
    description: "Three new interprocedural defect injectors, fail-closed with not-inert guard-disabled twins, registered in AllInjectors"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestPhase13InjectorsProduceExactlyOneDefect"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestPhase13InjectorGuardsAreNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestEveryInjectorRefusesWhenMarkerDisappears"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase6_injectors_test.go#TestInjectorTargetChoiceIsSpecified"
        status: pass
    human_judgment: false
  - id: D2
    description: "Held-out/derivation corpus including D-13-28's twin pair (both halves), all checking clean unmutated and yielding exactly one diagnostic mutated"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestHeldoutBaselinesAreFailClosed"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang --json check testdata/phase13/heldout_shared_callee_twin_alpha.lang (diagnostics: [])"
        status: pass
      - kind: other
        ref: "go run ./cmd/lang --json check testdata/phase13/heldout_shared_callee_twin_mirror.lang (diagnostics: [])"
        status: pass
    human_judgment: false
  - id: D3
    description: "Topology-disjointness control: held-out/derivation triples differ by call-graph depth, not identifiers; the control is proven not inert against an alpha-renamed copy"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestCorpusTopologyDisjoint"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestCorpusTopologyGuardIsNotInert"
        status: pass
    human_judgment: false
  - id: D4
    description: "Held-out corpus sealed with HELDOUT.sha256, seal proven not inert via one-byte flip, corpus not misroutable by cmd/lang/main.go's existing dispatch"
    requirement: "DX-07"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestHeldoutCorpusSealed"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestHeldoutCorpusSealGuardIsNotInert"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase13_injectors_test.go#TestPhase13CorpusIsNotMisroutedByCorpusDispatch"
        status: pass
      - kind: other
        ref: "shasum -a 256 -c testdata/phase13/HELDOUT.sha256"
        status: pass
    human_judgment: false
  - id: D5
    description: "D-13-02a empirical finding on whether a B1-shaped defect is constructible at this language maturity"
    verification: []
    human_judgment: true
    rationale: "This is an analytical finding about the language's checker architecture (sameType's precondition placement), not a property a unit test asserts a pass/fail verdict for -- it is evidence for 13-06's own required decision, and needs a human/13-06-plan-author read of the argument, not an automated check."

duration: 60min
completed: 2026-09-13
status: complete
---

# Phase 13 Plan 04: Held-Out Corpus and Interprocedural Defect Injectors Summary

**Three marker-driven interprocedural defect injectors plus a seven-fixture held-out/derivation corpus (including D-13-28's twin pair) whose split is proven, by an independently re-derived topology triple, to vary call-graph depth rather than identifiers -- sealed with `HELDOUT.sha256` before any repair rule exists to tune against it.**

## Performance

- **Duration:** ~60 min
- **Base SHA:** `ef9a03eebb677a853e571a5def53822ab7711936`
- **Completed:** 2026-09-13
- **Tasks:** 3
- **Files modified:** 12 (2 modified, 10 created)
- **Commits:** 3 (`35a8647`, `5356724`, `47302f1`)

## Accomplishments

- Shipped `InterproceduralLoanInjector`, `FallibleConsumeInjector`, and `CallArgumentTypeInjector` in `internal/compiler/session/session_phase13_injectors.go`, each cloning the Phase 6 `Injector`/`markerGuard` shape exactly: a marker constant, a fail-closed `Inject`, and a never-called `*InjectSkippingGuard` twin (D-13-30a). Registered in `AllInjectors` so the existing `TestEveryInjectorRefusesWhenMarkerDisappears` covers all three automatically; `TestInjectorTargetChoiceIsSpecified` (hand-written, not registry-driven) was extended explicitly with three new ambiguous-input cases, as the plan's own contingency required.
- Authored the full seven-fixture `testdata/phase13` corpus: D-13-28's twin pair (both halves -- `heldout_shared_callee_twin_alpha.lang`, where the marked/mutated caller is `alpha`, and `heldout_shared_callee_twin_mirror.lang`, where it is `beta`), the fallible-consume held-out/derivation pair, and the call-argument-mismatch held-out/derivation pair plus its ambiguous sibling (a prior `let` bringing a second same-typed place into scope, per 13-RESEARCH.md Pitfall 3). Every fixture verified via `go run ./cmd/lang --json check` to check clean unmutated.
- Built an independent topology-triple re-derivation (`callEdges`/`bfsHopDistance`/`computeCorpusTopology` in the test file) that parses via `syntax.Parse` directly -- no import of the `callgraph` package -- and computes `(function count, call-edge count, root-to-detection hop distance)`. `TestCorpusTopologyDisjoint` asserts held-out and derivation triples differ and held-out hop distance is `>= 2`; `TestCorpusTopologyGuardIsNotInert` (D-13-30b) feeds an alpha-renamed copy of a derivation fixture through the same computation and confirms the rename leaves the triple unchanged -- i.e. would have tripped the disjointness fatal, proving the control is not ceremonial.
- Sealed the five `heldout_*.lang` fixtures with `testdata/phase13/HELDOUT.sha256` (sorted by path, `shasum -a 256 -c`-compatible) and `TestHeldoutCorpusSealed` (no bless/regenerate path anywhere). `TestHeldoutCorpusSealGuardIsNotInert` (D-13-30c) flips one byte in a temp copy and confirms the digest comparison reports a mismatch without touching the real fixture. `TestPhase13CorpusIsNotMisroutedByCorpusDispatch` confirms no file under `testdata/phase13` collides with `cmd/lang/main.go`'s existing `isPhase5/6/7Corpus` marker filenames.
- Produced the empirical D-13-02a finding below.

## D-13-02a finding

**Question posed by 13-CONTEXT.md:** is a B1-shaped defect (blame belonging to the callee's own declared-fact violation, independent of caller misuse) constructible at this language's current maturity? If yes, criterion 3's twin pair should exercise it and `resolveBlame` should be wired on that path. If no, that is a terminal finding about the phase's riskiest assumption.

**Finding: no B1-shaped interprocedural diagnostic is constructible at this maturity, and the reason is structural, not incidental.**

Evidence, read directly from `internal/compiler/check/check.go` this session:

- `sameType(function.ReturnType, function.Parameter.Type)` is checked at the HEAD of every function-body admission path -- `checkBranch` (check.go:255), and the straight-line/fallible counterparts (check.go:3148, check.go:3399) -- BEFORE that function's body is checked at all, and unconditionally, independent of whether the function is ever called. A function whose own declared return type does not equal its own declared parameter type is refused immediately with `type.return_mismatch`, Primary span the function's OWN declaration -- not a call site, and not gated on any caller existing.
- Because every Lang function has exactly one parameter and (per `sameType` enforcement) exactly one type fact (`LANGUAGE-MATURITY.md`, D-07-09), "the callee's own declared contract is self-contradictory" is not a state a callee can be in and still be admitted as a function at all. The self-consistency check runs as a precondition of EVERY subsequent pass, interprocedural passes included: `buildInterproceduralSummaries` and `resolveCallBinding` only ever see callees that have already survived `sameType`.
- The three D-13-09 Set A target codes (`check.interprocedural_loan_liveness`, `syntax.fallible_call_not_consumed`, `check.call_argument_type_mismatch`) are, per 13-RESEARCH.md's own verified table, all THREE among the original eight codes D-13-04 already confirmed are B2-shaped (Primary span lands on the CALLER). This plan's fixtures reuse those same three codes -- there is no new diagnostic code introduced here that could be B1-shaped even in principle.

**Consequence:** by the time any interprocedural diagnostic in this codebase could fire, the callee is already known-self-consistent (refused earlier, independently, with zero calls involved, if it were not). There is no reachable program shape where a call site's diagnostic is caused by the callee's own body contradicting its own declared signature -- that failure mode is caught and reported BEFORE the interprocedural pass exists to reach it. This is a genuine architectural property, not a corpus-authoring gap: no fixture, however constructed, can produce a B1-shaped diagnostic at this maturity, because the checker's own ordering makes the precondition for B1's existence (a self-inconsistent, still-admitted callee) unreachable.

**This settles D-13-02a's resolution (b):** no B1-shaped defect is constructible at this language maturity. 13-05/13-06 should NOT attempt to wire `resolveBlame` on a hunt for a B1-shaped twin-pair member; the twin pair this plan ships tests the repair-emission span choice (D-13-25's actual overfittable artifact), and that is DX-06's honestly-reachable evidence at this maturity, not a coincidental pass. Revisit D-13-02a if/when the language grows multi-parameter functions or per-function multiple type facts (LANGUAGE-MATURITY.md's "not started, not scheduled" items 2-4) -- only then could a callee's own body diverge from its declared contract while still surviving `sameType` in its current single-fact form.

## Task Commits

1. **Task 1: Marker-driven interprocedural injectors with fail-closed guards and not-inert twins** - `35a8647` (feat)
2. **Task 2: The corpus -- including D-13-28's twin pair, BOTH halves -- and the topology-disjointness control** - `5356724` (feat)
3. **Task 3: Seal the held-out corpus with `HELDOUT.sha256` and prove the seal is not inert** - `47302f1` (feat)

## Files Created/Modified

- `internal/compiler/session/session_phase13_injectors.go` - three new injectors (interprocedural_loan, fallible_consume, call_argument_type) plus their guard-disabled twins
- `internal/compiler/session/session_phase13_injectors_test.go` - Task 1/2/3 tests: injector correctness/not-inert, topology-triple disjointness/not-inert, held-out baselines, seal/not-inert, dispatch-collision guard
- `internal/compiler/session/session_phase6_injectors.go` - registered the three new injectors in `AllInjectors`
- `internal/compiler/session/session_phase6_injectors_test.go` - extended `markerAbsentInput` and `TestInjectorTargetChoiceIsSpecified` with the three new injector cases
- `testdata/phase13/heldout_shared_callee_twin_alpha.lang` / `heldout_shared_callee_twin_mirror.lang` - D-13-28's criterion-3 twin pair
- `testdata/phase13/heldout_fallible_call_unconsumed.lang` / `derivation_fallible_call_unconsumed.lang` - class 2 held-out/derivation pair
- `testdata/phase13/heldout_call_argument_mismatch.lang` / `derivation_call_argument_mismatch.lang` - class 3 held-out/derivation pair
- `testdata/phase13/heldout_call_argument_ambiguous.lang` - class 3's D-13-10 positive-refusal fixture
- `testdata/phase13/HELDOUT.sha256` - sealed digest manifest for the five held-out fixtures

## Decisions Made

See `key-decisions` in the frontmatter. Summarized: hop distance is root-to-DETECTION call-graph depth (not detection-to-fix in the literal blame sense, since the true-fix function's identity is still an open question the driver/repair plans resolve); the twin pair's two halves differ by which caller is marked/reachable, not yet by callee-vs-caller blame (deferred to 13-05/13-06); `TestCorpusTopologyDisjoint` is scoped to the two classes with a derivation counterpart; D-13-02a is resolved as "not constructible at this maturity" with the architectural evidence above.

## Deviations from Plan

### Auto-fixed Issues

None — Rules 1-3 did not trigger; no bugs, missing-critical gaps, or blockers were found requiring auto-fix during implementation.

### Interpretive discretion (not a Rule 1-4 deviation, recorded for transparency)

**1. Hop-distance metric scope and definition, and the twin pair's blame-discrimination deferral**
- **Found during:** Task 2, while constructing D-13-28's twin pair
- **Issue:** D-13-28's prose describes a twin pair where fixing the DETECTION site (a caller) leaves the OTHER caller equally broken (for the alpha half) versus a caller-only fix sufficing (for the mirror half) -- a distinction that requires actually applying a repair and re-checking, which does not exist until 13-05/13-06. Constructing THAT literal discrimination now, with only injectors and no repair rules, would require either an already-broken second caller in the "clean baseline" (violating D-13-26.3) or speculative repair logic out of this plan's scope.
- **Resolution:** Built both halves to satisfy this plan's OWN literal, testable obligations (D-13-26.3's clean-baseline/one-diagnostic property; D-13-26.2's hop distance >= 2, read as root-to-detection call-graph depth) while preserving the STRUCTURAL asymmetry (which caller is marked and reachable) the twin pair needs, and documented in file-header comments and this SUMMARY exactly which deeper claim is deferred to 13-05/13-06's driver test once the sibling repairs exist.
- **Files affected:** `testdata/phase13/heldout_shared_callee_twin_alpha.lang`, `testdata/phase13/heldout_shared_callee_twin_mirror.lang`, `internal/compiler/session/session_phase13_injectors_test.go` (topology triple's hop-distance doc comment)
- **Verification:** Both fixtures check clean unmutated and yield exactly one diagnostic when mutated (`TestHeldoutBaselinesAreFailClosed`); reasoning recorded in code comments and this SUMMARY for 13-05/13-06 to pick up.
- **Committed in:** `5356724`

---

**Total deviations:** 0 auto-fixed; 1 interpretive discretion documented.
**Impact on plan:** No scope creep or correctness risk. The discretion is a scoping clarification consistent with the plan's own explicit deferral of classes 2/3's repair-presence assertions to 13-05/13-06, applied symmetrically to the twin pair's deeper blame-discrimination claim.

## Issues Encountered

None — all `<verify>` blocks and acceptance criteria passed on first construction after the design phase; no repair loops were needed beyond normal fixture-construction iteration (validated incrementally via `go run ./cmd/lang --json check` per the plan's own mandated discipline).

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The held-out corpus is sealed and ready for plan 13-05 to write `wrap_call_in_try` and `use_matching_argument` without ever having seen `heldout_*` bytes influence their design (the `derivation_*` fixtures are the only ones plan 13-05 should consult by hand).
- 13-06 has a grounded, evidence-backed answer to D-13-02a ("not constructible at this maturity") rather than a guess, and should record it as the terminal finding 13-CONTEXT.md anticipated rather than re-investigating.
- 13-06's own driver test is where D-13-28's full whole-program-clean discrimination (does a caller-only fix leave the untouched twin still broken, does a callee fix clear both) gets its first real exercise -- this plan intentionally stopped short of that, per its own "do not yet assert repair presence" instruction.
- No blockers.

## Self-Check: PASSED

- `internal/compiler/session/session_phase13_injectors.go` — FOUND
- `internal/compiler/session/session_phase13_injectors_test.go` — FOUND
- `testdata/phase13/heldout_shared_callee_twin_alpha.lang` — FOUND
- `testdata/phase13/heldout_shared_callee_twin_mirror.lang` — FOUND
- `testdata/phase13/heldout_fallible_call_unconsumed.lang` — FOUND
- `testdata/phase13/heldout_call_argument_mismatch.lang` — FOUND
- `testdata/phase13/heldout_call_argument_ambiguous.lang` — FOUND
- `testdata/phase13/derivation_fallible_call_unconsumed.lang` — FOUND
- `testdata/phase13/derivation_call_argument_mismatch.lang` — FOUND
- `testdata/phase13/HELDOUT.sha256` — FOUND
- Commit `35a8647` — FOUND in `git log --oneline --all`
- Commit `5356724` — FOUND in `git log --oneline --all`
- Commit `47302f1` — FOUND in `git log --oneline --all`
- `go test ./internal/compiler/session/... -run 'TestPhase13Injectors|TestPhase13InjectorGuardsAreNotInert|TestEveryInjectorRefusesWhenMarkerDisappears|TestInjectorTargetChoiceIsSpecified|TestInjectorMarkerCountGuardIsNotInert|TestEveryInjectorProducesExactlyOneMechanicalChange' -v -count=1` — PASS (all subtests green)
- `go test ./internal/compiler/session/... -run 'TestCorpusTopologyDisjoint|TestCorpusTopologyGuardIsNotInert|TestHeldoutBaselinesAreFailClosed'` — PASS
- `go test ./internal/compiler/session/... -run 'TestHeldoutCorpusSealed|TestHeldoutCorpusSealGuardIsNotInert|TestPhase13CorpusIsNotMisroutedByCorpusDispatch'` — PASS
- `shasum -a 256 -c testdata/phase13/HELDOUT.sha256` — all OK
- `go build ./... && go vet ./internal/compiler/session/...` — clean
- `go test ./...` — PASS (full suite green, see below)

---
*Phase: 13-agent-loop-for-interprocedural-defects*
*Completed: 2026-09-13*
