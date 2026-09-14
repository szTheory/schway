---
phase: 08-interprocedural-loan-liveness-in-check
plan: 03
subsystem: check
tags: [ownership, loan-liveness, interprocedural, testdata, corpus, disclosure, go]

# Dependency graph
requires:
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 01
    provides: "interproceduralSummaryTable/buildInterproceduralSummaries scaffold, checkInterproceduralLoanLiveness/interproceduralLoanLivenessDiagnostic, the forward (D-08-07) direction"
  - phase: 08-interprocedural-loan-liveness-in-check
    plan: 02
    provides: "deriveFunctionUsesParam, the backward blockLoanLiveness OpCall gate (D-08-08), the two-direction diagnostic classification"
provides:
  - "Nine real .lang fixtures under testdata/phase08/: the adversarial corpus success criterion 1 is adjudicated from"
  - "readPhase08Fixture, TestInterproceduralTwinPairsDifferOnlyInTheCallee: the mechanical twin-discipline harness (T-08-09)"
  - "TestInterproceduralLivenessTwinPatternA/TestInterproceduralLivenessTwinPatternBRealFixtures, TestInterproceduralLivenessRelayDepth2, TestInterproceduralLivenessNegativeControl, TestInterproceduralLivenessMatchArmRegression"
  - "interproceduralConsultObserved + TestInterproceduralDisclosedFieldSet: criterion 4's refusal half, machine-asserted"
  - "peerDivergenceExpected entries for the two new ACCEPT-twin fixtures corevalidate's still-intraprocedural peer diverges on"
affects: [08-04, 08-05, 08-06, 09-peer-re-derivation-and-d-03-02-closure]

# Actuals (#2632)
actuals:
  tokens: 11464
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "A twin pair's discipline is enforced by parsing BOTH real .lang members and structurally comparing the CALLER's own ast.LinearBody (binding-by-binding), never by review alone -- a caller-varying pair fails a test outright."
    - "A composition-depth-2 real fixture needs the MIDDLE function to locally borrow its own parameter before forwarding to the leaf, not merely forward the bare parameter -- originvalidate's origin-recomputation walk is transparent through an OpCall (continues to the call's own SourceID) but needs a LOCAL borrow hop somewhere in the chain back to the function's own parameter, or the middle function is unpublishable/uncallable (core.origin_understated), an unrelated, pre-existing gate."
    - "A negative control that must vary a callee's Fails/ForeignReach while holding its declared return contract and body fixed cannot make the SAME function both foreign-fallible and locally-borrow-returning (checkFallibleLinear's single-try-binding-or-all-fallible shape excludes mixing with an ordinary borrow op) -- route the Fails/ForeignReach variance through a SEPARATE leaf the liveness-critical callee calls with an ordinary (non-try) binding, inheriting Fails/ForeignReach transitively (07-12's joinFails/joinForeignReach) while its OWN declared signature never changes."

key-files:
  created:
    - testdata/phase08/twin_a_refuse.lang
    - testdata/phase08/twin_a_accept.lang
    - testdata/phase08/twin_b_refuse.lang
    - testdata/phase08/twin_b_accept.lang
    - testdata/phase08/relay_depth2_refuse.lang
    - testdata/phase08/relay_depth2_accept.lang
    - testdata/phase08/negative_control_fails.lang
    - testdata/phase08/negative_control_infallible.lang
    - testdata/phase08/match_arm_call.lang
  modified:
    - internal/compiler/check/check_test.go
    - internal/compiler/check/check.go
    - internal/compiler/session/session_peer_gate_test.go

key-decisions:
  - "Pattern A's ACCEPT twin (twin_a_accept.lang) and the depth-2 ACCEPT twin (relay_depth2_accept.lang) are verified clean at the check.Program() level only, not through the full `lang check` CLI: corevalidate's own loan-liveness peer is still intraprocedural this phase (unchanged, explicitly Phase 09's charter) and unconditionally treats an OpCall's target as inheriting every loan its argument carries, so it independently refuses BOTH fixtures via core.move_while_borrowed regardless of the callee's declared contract. Registered in session's own peerDivergenceExpected map (07-10's established mechanism for exactly this class of finding), matching the pattern 08-01 already used for relay_escort_witness.lang."
  - "Pattern B's real .lang twin pair (twin_b_refuse.lang/twin_b_accept.lang) cannot demonstrate differing end-to-end verdicts through the CLI: the caller's required `borrow; move; call` shape is refused INTRAPROCEDURALLY (ownership.move_while_borrowed) for BOTH members alike by computeLoanLastUses' summary-blind AST-shadow admission path, before check's own interprocedural pass ever runs -- 08-02's own SUMMARY explicitly flagged this as an open item for whichever plan landed this fixture pair. The twin discipline (byte-identical caller, callee differing only in UsesParam-relevant behavior) is still verified structurally; the CONTRACT-DRIVEN differentiation itself is proven at the checked-core level by 08-02's own TestInterproceduralLivenessTwinPatternB, extended here to also exercise and document the real fixtures' true (identical) end-to-end verdict."
  - "The negative control's Fails/ForeignReach variance lives on a SEPARATE leaf function the liveness-critical relay calls (ordinary, non-try binding), not on relay itself -- checkFallibleLinear's shape rules make it structurally impossible for the SAME function to be both foreign-fallible and locally-borrow-returning. relay's OWN declared signature and body stay byte-identical between the fails/infallible members; only its transitively-INHERITED Fails/ForeignReach (07-12's own closure-join) differs."

patterns-established:
  - "A fixture-header convention for this phase's adversarial corpus: state which pattern the fixture belongs to, which member of its twin it is, the single thing that differs, the exact expected diagnostic code (or 'no diagnostics'), and any known end-to-end CLI residual with its full technical explanation -- mirroring testdata/phase07/relay_escort_witness.lang's own header depth."

requirements-completed: [OWN-06]

coverage:
  - id: D1
    description: "Pattern A's real twin pair (ReturnsBorrowOfParam) refuses/admits contract-driven at the check.Program level, with the twin discipline mechanically enforced"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessTwinPatternA"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralTwinPairsDifferOnlyInTheCallee"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase08/twin_a_refuse.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "Pattern B's real twin pair exists, parses cleanly, and its TRUE end-to-end verdict (identical intraprocedural refusal on both members, a documented pre-existing limitation) is asserted rather than assumed; the contract-driven backward gate itself stays proven at the checked-core level"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessTwinPatternBRealFixtures"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessTwinPatternB"
        status: pass
    human_judgment: false
  - id: D3
    description: "Composition depth >= 2: check refuses a depth-2 relay chain whose loan state diverges across the boundary and admits its safe twin, with the caller's admission proven to consult RELAY's own summary entry, never leaf's"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessRelayDepth2"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase08/relay_depth2_refuse.lang"
        status: pass
    human_judgment: false
  - id: D4
    description: "A negative-control pair differing only in Fails/ForeignReach resolves identically in both members (verdict AND summary-entry field equality), proving the law ignores fields it must not read"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessNegativeControl"
        status: pass
    human_judgment: false
  - id: D5
    description: "A call appearing in both match arms is covered by a regression fixture, verdict matching the straight-line equivalent (recommended, not gate-blocking)"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralLivenessMatchArmRegression"
        status: pass
    human_judgment: false
  - id: D6
    description: "The exact set of callee-signature fields the derivation consults is machine-asserted as a subset of {return.mode, parameters[0].mode}, and the diagnostic's disclosed field is proven to have actually been consulted, for every fixture in the corpus plus relay_escort_witness.lang"
    requirement: "OWN-06"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_test.go#TestInterproceduralDisclosedFieldSet"
        status: pass
    human_judgment: false
  - id: D7
    description: "No existing ownership.*/check.*/core.* verdict moved; the whole pre-existing suite plus go vet stay green (excluding the pre-existing, unrelated session-package spike-registry gap)"
    verification:
      - kind: unit
        ref: "go test ./... && go vet ./..."
        status: pass
    human_judgment: false

# Metrics
duration: ~30min
completed: 2026-09-09
status: complete
---

# Phase 08 Plan 03: The Adversarial Composition-Depth Corpus Summary

**Nine real `.lang` fixtures now adjudicate success criterion 1 through the shipped parser and checker -- two contract-driven twin pairs, a depth-2 relay chain, a Fails/ForeignReach negative control, and a match-arm regression -- plus a mechanical twin-discipline harness and a machine-asserted, closed two-field consulted-signature set that makes the shipped diagnostic's own disclosure claim TRUE, not merely well-formed.**

## Performance

- **Duration:** ~30 min
- **Started:** 2026-09-09T21:23:14-04:00 (previous plan-doc commit)
- **Completed:** 2026-09-09T21:52:35-04:00
- **Tasks:** 3 completed (all `type="auto"`)
- **Files modified:** 12 (9 new fixtures, 3 modified Go files)

## Accomplishments

- Landed `twin_a_refuse.lang`/`twin_a_accept.lang` (Pattern A, `ReturnsBorrowOfParam`) and `twin_b_refuse.lang`/`twin_b_accept.lang` (Pattern B, `UsesParam`): the two twin pairs that prove the interprocedural derivation genuinely reads the CALLEE's declared/derived contract, never the caller's shape -- `TestInterproceduralTwinPairsDifferOnlyInTheCallee` parses both members of each pair and mechanically asserts the caller's own `ast.LinearBody` is structurally identical, so a caller-varying (decorative) pair can no longer be introduced without failing a test (T-08-09's mitigation).
- Landed `relay_depth2_refuse.lang`/`relay_depth2_accept.lang` (D-08-28.2): a genuine caller -> relay -> leaf composition-depth-2 chain. `TestInterproceduralLivenessRelayDepth2` proves the caller's admission consults RELAY's own interprocedural summary entry -- never leaf's -- closing the gap a depth-1-only corpus cannot close (it cannot distinguish "reads a declared signature" from "works for one hop").
- Landed `negative_control_fails.lang`/`negative_control_infallible.lang` (D-08-28.4): two programs whose liveness-critical callee (`relay`) differs ONLY in its transitively-inherited `Fails`/`ForeignReach` (routed through a separate leaf, since a single function cannot be both foreign-fallible and locally-borrow-returning). `TestInterproceduralLivenessNegativeControl` asserts both verdict equality AND field-for-field `interproceduralSummary` equality for `relay` between the two programs.
- Landed `match_arm_call.lang` (D-08-28.5, recommended, not gate-blocking) and `TestInterproceduralLivenessMatchArmRegression`, restating D-07-28's both-arms-call shape under this phase's own namespace and recording that its verdict matches an equivalent straight-line program's.
- Added `interproceduralConsultObserved` (a nil-in-production hook mirroring `callSignatureTableLookupObserved`'s own shape) at `buildInterproceduralSummaries`' exactly two field reads, and `TestInterproceduralDisclosedFieldSet`, table-driven over the whole `testdata/phase08` corpus plus `relay_escort_witness.lang`: asserts the consulted field set is always a subset of `{"return.mode", "parameters[0].mode"}`, and that every emitted diagnostic's third cause names a field that was actually recorded as consulted for that same callee -- criterion 4's refusal half, machine-asserted rather than merely documented (D-08-26).

## Task Commits

Each task was committed atomically:

1. **Task 1: The two twin pairs, and the harness that enforces the twin discipline** - `c1cea1f` (feat)
2. **Task 2: Composition depth >= 2, the negative control, and the match-arm regression** - `786550e` (feat)
3. **Task 3: Criterion 4's refusal half -- assert the exact consulted field set** - `7ad156a` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE.md/ROADMAP.md/REQUIREMENTS.md)

_Note: `workflow.tdd_mode` is `false` for this project (config.json); no task in this plan was marked `tdd="true"`. Fixtures and their driving tests were developed and verified together, and each task's own `<verify>` block was run and confirmed green before that task's single commit._

## Files Created/Modified

- `testdata/phase08/twin_a_refuse.lang` / `twin_a_accept.lang` - Pattern A's real twin pair
- `testdata/phase08/twin_b_refuse.lang` / `twin_b_accept.lang` - Pattern B's real twin pair
- `testdata/phase08/relay_depth2_refuse.lang` / `relay_depth2_accept.lang` - composition depth-2 chain
- `testdata/phase08/negative_control_fails.lang` / `negative_control_infallible.lang` - the Fails/ForeignReach negative control
- `testdata/phase08/match_arm_call.lang` - the both-arms-call regression
- `internal/compiler/check/check_test.go` - `readPhase08Fixture`, `callerLinearBody`, `stringSlicesEqual`, `assertIdenticalCallerBodies`, `TestInterproceduralTwinPairsDifferOnlyInTheCallee`, `TestInterproceduralLivenessTwinPatternA`, `TestInterproceduralLivenessTwinPatternBRealFixtures`, `functionIDByName`, `TestInterproceduralLivenessRelayDepth2`, `TestInterproceduralLivenessNegativeControl`, `TestInterproceduralLivenessMatchArmRegression`, `splitCalleeFieldDetail`, `TestInterproceduralDisclosedFieldSet`
- `internal/compiler/check/check.go` - `interproceduralConsultObserved` (declaration + the two call sites inside `buildInterproceduralSummaries`)
- `internal/compiler/session/session_peer_gate_test.go` - two new `peerDivergenceExpected` entries (`twin_a_accept.lang`, `relay_depth2_accept.lang`)

## Decisions Made

See `key-decisions` in frontmatter for the three load-bearing decisions this plan made while landing the corpus: (1) the ACCEPT twins' known corevalidate residual, registered rather than worked around; (2) Pattern B's real fixtures' TRUE (identical, intraprocedural) end-to-end verdict, asserted honestly rather than forced to match the plan's literal prose; (3) routing the negative control's Fails/ForeignReach variance through a separate leaf rather than the liveness-critical callee itself.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] `twin_a_accept.lang` and `relay_depth2_accept.lang` trip an existing corpus-wide divergence register**
- **Found during:** Task 1 and Task 2, while verifying the plan's own `<verify>` CLI commands against the full `lang check` pipeline
- **Issue:** `TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` (session package, 07-10) walks every `.lang` file under `testdata/` and asserts the exact set of check-admits/corevalidate-refuses divergences equals a hand-maintained register. Both new ACCEPT-twin fixtures are genuine, expected divergences (see Decisions above) and tripped this test as UNDECLARED.
- **Fix:** Added both fixtures to `peerDivergenceExpected` with `core.move_while_borrowed` and a comment explaining the mechanism, following the exact precedent 08-01 already established for `relay_escort_witness.lang` in the same map.
- **Files modified:** `internal/compiler/session/session_peer_gate_test.go`
- **Verification:** `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus` passes; full `go test ./...` shows only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure.
- **Committed in:** `7ad156a`

---

**Total deviations:** 1 auto-fixed (1 Rule 1 -- a genuine, mechanically-detected consequence of landing fixtures whose accepted verdict is check-only, not architectural, no scope creep beyond the fix itself: it is exactly the maintenance action the existing register was designed for).
**Impact on plan:** Necessary to keep the pre-existing corpus-wide divergence gate green; does not touch corevalidate's production code (explicitly out of this plan's scope and Phase 09's own charter).

### Plan-text vs. verified-reality notes (not Rule 1-3 auto-fixes -- documented for transparency, no code path changed to "fix" these)

**A. `twin_a_accept.lang`'s CLI verification is check-level, not full-CLI.** The plan's `<verify>` block only exercises `twin_a_refuse.lang` through the CLI (matching what was verified); the acceptance criteria's "zero error diagnostics" claim for `twin_a_accept.lang` is satisfied at the `check.Program()` level (proven directly, `TestInterproceduralLivenessTwinPatternA`), not through the full CLI (which additionally surfaces corevalidate's documented residual, item 1 above). No plan text explicitly required the full-CLI claim for the accept twin, and the CLI command for it was not part of the plan's own `<verify>` block.

**B. Pattern B's real fixtures do not demonstrate a differing CLI verdict.** The plan's action text describes extending `TestInterproceduralLivenessTwinPatternB` "to drive the real fixtures" with the refusing/accepting split. Empirically verified (both via direct CLI runs and via `Program()` in-test) that BOTH `twin_b_refuse.lang` and `twin_b_accept.lang` are refused identically at the INTRAPROCEDURAL layer (`ownership.move_while_borrowed`), before check's interprocedural pass ever runs -- exactly the residual 08-02-SUMMARY.md's own "Next Phase Readiness" section flagged for this plan to account for. `TestInterproceduralLivenessTwinPatternBRealFixtures` asserts this TRUE, verified outcome rather than a verdict split the language's current intraprocedural admission path cannot yet produce for this real source shape. The plan's own acceptance criteria for Task 1 do not require a CLI/Program() divergence check for `twin_b_accept.lang` specifically (only structural/parse checks), so this is not a failed acceptance criterion -- it is an honest resolution of prose that assumed a capability (summary-aware AST-shadow admission) this plan's `files_modified` (test/fixture files only, no `check.go` production changes to `computeLoanLastUses`) does not include.

## Issues Encountered

None beyond the deviation and notes above -- both were surfaced by running the plan's own verification commands and existing tests, not by unrelated exploration.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- The full nine-fixture adversarial corpus for success criterion 1 exists in real `.lang` source, each shape carrying its own "or the gate is decorative" argument in a fixture header and/or test doc comment.
- The twin discipline is enforced by a test (`TestInterproceduralTwinPairsDifferOnlyInTheCallee`), not by review alone.
- Success criterion 4's refusal half is machine-asserted (`TestInterproceduralDisclosedFieldSet`); its accepted-program half remains declared debt exactly as `PHASE-08-DEBT.md`'s D-08-26 already states (no shipped runtime artifact this phase; `protocol.ExplainSummary` needs a diagnostic to expand).
- **For Phase 09 (peer re-derivation and D-03-02 closure):** this plan's own corpus now names, in `testdata/phase08/twin_a_accept.lang`'s and `relay_depth2_accept.lang`'s headers plus `peerDivergenceExpected`'s comments, the EXACT mechanism (`loanChainIndex.carriedLoans`' unconditional `parent[TargetID] = SourceID` for every `OpCall`) that keeps corevalidate's own loan-liveness re-derivation intraprocedural. This is concrete, fixture-backed input for whichever 09 plan extends corevalidate to consult callee signatures the same way `check` now does.
- No blockers for 08-04 through 08-06.

## Self-Check: PASSED

- FOUND: testdata/phase08/twin_a_refuse.lang
- FOUND: testdata/phase08/twin_a_accept.lang
- FOUND: testdata/phase08/twin_b_refuse.lang
- FOUND: testdata/phase08/twin_b_accept.lang
- FOUND: testdata/phase08/relay_depth2_refuse.lang
- FOUND: testdata/phase08/relay_depth2_accept.lang
- FOUND: testdata/phase08/negative_control_fails.lang
- FOUND: testdata/phase08/negative_control_infallible.lang
- FOUND: testdata/phase08/match_arm_call.lang
- FOUND: internal/compiler/check/check_test.go
- FOUND: internal/compiler/check/check.go
- FOUND: internal/compiler/session/session_peer_gate_test.go
- FOUND commit: c1cea1f
- FOUND commit: 786550e
- FOUND commit: 7ad156a
- Re-ran acceptance criteria for all three tasks: `go test ./internal/compiler/check/... -run 'TwinPattern|TwinPairsDifferOnlyInTheCallee|RelayDepth2|NegativeControl|MatchArmRegression|InterproceduralDisclosedFieldSet' -v` -- all PASS
- Re-ran `go test ./internal/compiler/check/... -count=2 -shuffle=on` -- PASS
- Re-ran `go run ./cmd/lang --json check testdata/phase08/twin_a_refuse.lang` -- emits `check.interprocedural_loan_liveness` -- PASS
- Re-ran `go run ./cmd/lang --json check testdata/phase08/relay_depth2_refuse.lang` -- emits `check.interprocedural_loan_liveness` -- PASS
- Re-ran plan-level `<verification>`: `go test ./internal/compiler/check/...` exits 0; `go test ./... && go vet ./...` -- only the pre-existing, unrelated `TestQLT01RegistryCoversAllFiveSpikes` failure (confirmed present before this phase began) -- PASS

---
*Phase: 08-interprocedural-loan-liveness-in-check*
*Completed: 2026-09-09*
