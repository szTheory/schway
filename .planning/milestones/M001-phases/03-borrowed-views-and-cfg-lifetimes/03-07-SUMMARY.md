---
phase: 03-borrowed-views-and-cfg-lifetimes
plan: "07"
subsystem: compiler-frontend-tooling
tags: [debug-lineage, verify-gate, spawn-guard, native-timeout, cli-parity, phase-close]

requires:
  - phase: 03-borrowed-views-and-cfg-lifetimes
    provides: "03-01's CFG (core.Block/Edge/LoanEndpoint), 03-02's exclusive loans, 03-03's checkBranch dataflow, 03-04's corevalidate.recomputeLoanEndpoints and BorrowedLoanEndpointControlLane, 03-05's pathoracle and PathOracleDisagreementLane, 03-06's originvalidate package and public-origin fixtures — this plan wires all of it into one Phase 3 verify path and adds the debug-lineage side experiment"
provides:
  - "internal/compiler/debugmap: a bounded semantic debug-lineage side table (lang.debug-map/0) joining source span to core ID to operation/point identity, with honest available/not_captured reporting"
  - "`debug-map SRC [QUERY]` CLI subcommand"
  - "session.verifyBorrowedCorpus + scripts/verify-phase3.sh: the Phase 3 gate, requiring every Phase 3 control fail-closed and proving Phase 1/2 non-regression with the same once-built binary"
  - "testdata/phase3/borrowed_view.lang: the Phase 3 verify-corpus dispatch probe"
  - "D-02-01 (AST-resolving spawn guard), D-02-04 (native.timeout hang-mode falsifier), D-02-06 (tightened CLI stream ceiling), D-02-08 (Human/JSON convergence parity) closed"
affects: []

actuals:
  tokens: 19200
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "debugmap.Build reads two already-produced artifacts (ast.Program, core.Program) read-only and joins them positionally, rather than re-running or importing check — the join relies on check.go's own established 1:1 binding-to-operation correspondence (verified by reading check.go directly, including the one implicit alias-copy offset checkBranch prepends per arm)"
    - "verifyBorrowedCorpus composes 03-04's and 03-05's existing lane constructors (BorrowedLoanEndpointControlLane, PathOracleDisagreementLane) directly rather than re-implementing their mutation logic — the Phase 3 gate is an assembly of already-proven parts, not a fourth reimplementation"
    - "go/ast-resolving spawn-constructor and stream-capture detection (scanUnboundedSpawns), replacing a substring scan, entirely inside the existing test file — no production dependency added, matching D-15's no-new-dependency constraint"

key-files:
  created:
    - internal/compiler/debugmap/debugmap.go
    - internal/compiler/debugmap/debugmap_test.go
    - internal/compiler/testsupport/testsupport_internal_test.go
    - scripts/verify-phase3.sh
    - testdata/phase3/borrowed_view.lang
  modified:
    - cmd/lang/main.go
    - internal/compiler/session/session.go
    - internal/compiler/session/session_test.go
    - internal/compiler/testsupport/cli_test.go
    - internal/compiler/protocol/protocol.go
    - internal/compiler/protocol/protocol_test.go
    - internal/compiler/native/native_test.go
    - internal/compiler/testsupport/testsupport.go

key-decisions:
  - "borrowed_view.lang cannot also carry the OWN-04 'public borrowed return' case in the same file: check.Program's own core.mixed_body_versions gate rejects a module mixing a match-bodied function (the branch fixture) with a straight-line function (a public-view function), verified by reading check.go directly rather than assumed. verifyBorrowedCorpus exercises the three origin controls (understated/impossible/stale-summary) against the existing testdata/phase3/public_view*.lang fixtures instead, in the same lane."
  - "edge-last-use-omitted and loan-endpoint-mismatch are named as two laws in the plan's must_haves but implemented as one control (control:core.loan_endpoint_mismatch, via 03-04's already-shipped BorrowedLoanEndpointControlLane): they are the same law in this codebase's current implementation (an omitted edge-specific loan endpoint IS the loan-endpoint mismatch), and 03-04's own TestBorrowedLaneRecordsFailureStatus already asserts that lane emits exactly one control — adding a second control string to it would have broken an already-shipped, unrelated regression test for no correctness gain."
  - "debugmap.Build does not import or re-run check: it joins ast.Program's span-bearing bindings against core.Program's already-lowered operations by the SAME positional correspondence check.go's own analyzeStraightLine/analyzeArmBody establish (1:1 for straight-line; 1:1 after skipping checkBranch's one implicit per-arm alias-copy operation, discovered by a failing test before this was understood, not assumed from a prior summary read alone)."
  - "protocol.Human's signature changed from `string` to `(string, error)` to close D-02-08: both projections now share the identical convergence-or-error contract. Both call sites (cmd/lang/main.go, protocol_test.go) were updated in the same commit as the change."
  - "MaxCLIStreamBytes tightened from 8 MiB to 1 MiB (D-02-06), not to the literal 1,756 B high-water mark 02-DEBT.md measured for Phase 2's smaller test corpus — Phase 3's `--json verify testdata/phase3` documents and the boundary fixtures this repo now ships are larger than what D-02-06 originally measured, so 1 MiB keeps a real margin above the current measured maximum while still being a genuine ~2 orders of magnitude tighter ceiling than the prior 8 MiB, rather than an unmeasured round number."

requirements: [OWN-03, OWN-04]
requirements-completed: []

coverage:
  - id: D1
    description: "A bounded semantic debug map joins a source span to a core ID to an operation/point identity, covering a match arm, a move, a borrow, and a return, with honest available/not_captured reporting and fail-closed entry/output caps"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/debugmap#TestDebugMapJoinsSourceCoreAndOperation", status: pass}
      - {kind: unit, ref: "internal/compiler/debugmap#TestDebugMapReportsHonestAbsence", status: pass}
      - {kind: unit, ref: "internal/compiler/debugmap#TestDebugMapCapsFailClosed", status: pass}
      - {kind: unit, ref: "internal/compiler/debugmap#TestDebugMapIdentityUsesOrdinals", status: pass}
      - {kind: e2e, ref: "internal/compiler/testsupport#TestDebugMapCLIAvailability", status: pass}
    human_judgment: false
  - id: D2
    description: "The wiki note's four rejected experiment steps are named in the shipped package documentation"
    requirement: OWN-03
    verification: []
    human_judgment: true
    rationale: "Documentation-completeness is a review judgment, not a test assertion. debugmap.go's package doc comment names all four rejected steps (native debug info, panic/segfault capture, fault injection, latency measurement) by number and reason; a human should confirm the wording is auditable."
  - id: D3
    description: "sh scripts/verify-phase3.sh runs the shared Go suites once, builds the binary once, verifies Phase 1/2/3 with that one binary, and requires every Phase 3 control ID; it neither invokes verify-phase2.sh nor duplicates a shared suite"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/testsupport#TestPhase3VerifierScriptContract", status: pass}
      - {kind: integration, ref: "internal/compiler/session#TestVerifyPhase3ControlsAndWork", status: pass}
      - {kind: e2e, ref: "internal/compiler/testsupport#TestVerifyPhase3CLI", status: pass}
      - {kind: manual_procedural, ref: "sh scripts/verify-phase3.sh (exit 0, 7/7 Phase 3 controls, 20 warm samples per lane, Phase 1/2 controls also reported through the same binary)", status: pass}
    human_judgment: false
  - id: D4
    description: "Every Phase 3 verify lane records a status, carries nonzero recomputed work, and survives into the result even when the lane fails"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/session#TestBorrowedLaneRecordsFailureStatus", status: pass}
      - {kind: unit, ref: "internal/compiler/session#TestPathOracleLaneRecordsControl", status: pass}
      - {kind: integration, ref: "internal/compiler/session#TestVerifyPhase3ControlsAndWork", status: pass}
    human_judgment: false
  - id: D5
    description: "D-02-04/D-02-01/D-02-06 closed: native.timeout has a committed hang-mode falsifier, the AST-resolving spawn guard catches all four documented evasions and does not pass vacuously on zero sources, and the CLI stream ceiling boundary is exact"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/native#TestNativeTimeoutHasFalsifier", status: pass}
      - {kind: unit, ref: "internal/compiler/native#TestSourceNeverSpawnsUnboundedProcesses", status: pass}
      - {kind: unit, ref: "internal/compiler/native#TestSpawnGuardCatchesKnownEvasions", status: pass}
      - {kind: unit, ref: "internal/compiler/testsupport#TestCLIStreamCeilingBoundary", status: pass}
    human_judgment: false
  - id: D6
    description: "D-02-08 closed: the human and JSON projections share the identical convergence-or-error contract"
    requirement: OWN-03
    verification:
      - {kind: unit, ref: "internal/compiler/protocol#TestHumanJSONProjectionParity", status: pass}
    human_judgment: false

duration: 130min
completed: 2026-09-04
status: complete
---

# Phase 03 Plan 07: Bounded Debug-Lineage Experiment and the Phase 3 Gate Summary

**A bounded `lang.debug-map/0` side table joins source spans to typed-core identities with honest not_captured reporting, `scripts/verify-phase3.sh` becomes a real fail-closed gate requiring every Phase 3 control through one shipped binary that also proves Phase 1/2 non-regression, and D-02-01/D-02-04/D-02-06/D-02-08's four carried debt items close in the files this plan already touches — while D-03-01 and D-03-02 (the mid-phase gate's two open items) are carried forward intact, unresolved, and undisguised.**

## Performance

- **Duration:** ~130 min
- **Started:** 2026-09-04 (session start, after reading all prior plan summaries and 03-DEBT.md/02-DEBT.md)
- **Completed:** 2026-09-04
- **Tasks:** 3 completed
- **Files modified:** 13 (5 created, 8 modified)

## Accomplishments

- Built `internal/compiler/debugmap`, a new package implementing exactly steps 1, the semantic half of step 2, and step 4 of the wiki note's six-step debug-lineage experiment (D-01..D-04). `Build` reads an already-checked `ast.Program` and its resulting `core.Program` — never re-running or importing `check` — and joins each straight-line function's bindings, and each branch function's per-arm bindings (correctly skipping `checkBranch`'s one implicit per-arm alias-copy operation, a real correspondence bug caught by a failing test before any commit, not assumed from documentation), to their lowered `core.LinearOperation` identities. Every entry reports `available`; `Resolve` reports `not_captured` — never a guess — for an operation ID the map never produced. `MaxEntries`/`MaxOutputBytes` fail closed rather than truncate. The package doc comment names all four steps this phase explicitly rejects (native debug info, panic/segfault capture, fault injection, latency measurement) and why.
- Wired `debug-map SRC [QUERY]` through `cmd/lang/main.go` and `session.DebugMapCommandFile`, and `protocol.DebugMapSummary`/`DebugMapEntry` projection types, so the honest-absence behavior is provable through the shipped binary (`TestDebugMapCLIAvailability`), not only in-process.
- Wrote `testdata/phase3/borrowed_view.lang` — a single branch function combining a shared loan ending on its own arm's edge (before that arm's own later move) and a sequential exclusive loan in the sibling arm — as the Phase 3 verify-corpus dispatch probe. Discovered mid-session, by reading `check.Program` directly, that a module cannot mix a match-bodied function with a straight-line function (`core.mixed_body_versions`); the fixture therefore does not also carry a public borrowed return, and `verifyBorrowedCorpus` exercises those origin controls against the existing `public_view*.lang` fixtures instead, in the same lane.
- Implemented `session.verifyBorrowedCorpus`, dispatching in `VerifyCorpus` before the Phase 2 probe. It requires, fail-closed, with nonzero work on every lane and a failing lane still returned rather than dropped: the exclusive-conflict and exclusive-move negative controls (fresh checks against existing fixtures), the loan-endpoint-mismatch/edge-last-use-omitted law (reusing 03-04's `BorrowedLoanEndpointControlLane` unmodified — extending its control list would have broken its own already-shipped `TestBorrowedLaneRecordsFailureStatus`, which asserts exactly one control), the path-oracle-disagreement law (reusing 03-05's `PathOracleDisagreementLane` unmodified), and the three origin laws (understated, impossible, stale-summary), mutating honestly-checked `public_view*.lang` core artifacts the same way `originvalidate`'s own falsifiers do — `check.go`'s honest producer can never construct any of the three dishonest shapes itself.
- Wrote `scripts/verify-phase3.sh` as a peer of `verify-phase2.sh`, not an extension: it never invokes the Phase 2 script, runs each shared Go suite exactly once, builds the binary once, and re-verifies Phase 1 and Phase 2 with that same binary to prove non-regression, rather than delegating to the frozen prior script. `TestPhase3VerifierScriptContract` asserts both prohibitions and every required control ID and warm-observation token against the script's own text. Ran live: exit 0, all 7 Phase 3 controls present, 20 warm samples per lane, both expected escapes (`escape:coordinated-source-core-lie`, `escape:coordinated-frontend-summary-lie`) surfaced and never reported as detected controls.
- Closed **D-02-04**: added a hang mode to the native helper process and `TestNativeTimeoutHasFalsifier`, asserting both the compile and run deadlines fire `native.timeout` (not a compile/run-failure code) against a helper that blocks past a deliberately short test timeout. No production change was needed in `native.go` — the deadline mechanism itself was already correct; only its negative control was missing.
- Closed **D-02-01**: replaced the two-needle substring scan (`TestSourceNeverSpawnsUnboundedProcesses`) with `scanUnboundedSpawns`, a `go/ast`-resolving pass that identifies the actual spawn-constructor and stream-capture call shape rather than matching literal text. `TestSpawnGuardCatchesKnownEvasions` plants all four of D-02-01's documented evasions (a context-free constructor via `context.Background()`, a merged-output helper via `.CombinedOutput()`/`.Output()`, an unbounded pipe read via `.StdoutPipe()`/`.StderrPipe()`, and an aliased constructor via a bare `exec.Command`/`exec.CommandContext` value assigned to an identifier) in throwaway in-memory source files and confirms every one is caught, confirms a bare `exec.Command(` call is still caught, confirms an honestly bounded spawn (a real deadline context plus independently bounded writers) does not fire, and the scanner still fails rather than passing vacuously when it walks zero Go sources. No external Go dependency was added — only `go/ast`, `go/parser`, `go/token`, and `strconv` from the standard library. Re-run against the real repo, the new scanner still finds zero violations across every production spawn site.
- Closed **D-02-06**: tightened `MaxCLIStreamBytes` from 8 MiB to 1 MiB, with the measured rationale cited in the constant's own doc comment, and added `TestCLIStreamCeilingBoundary` asserting the boundary at exactly the limit, one byte over, and across a write split at the boundary.
- Closed **D-02-08**: `protocol.Human` now shares `protocol.JSON`'s exact convergence-or-error contract — both loops converge their own `output_bytes` metric against the actual rendered length before returning, and both fail with an error rather than silently serving an unconverged output on the fourth attempt. `TestHumanJSONProjectionParity` asserts both projections' self-reported `output_bytes` matches their own actual rendered length across a range of diagnostic counts.

## Task Commits

1. **Task 03-07-01 + 03-07-02 (interleaved, see Deviations): bounded debug-lineage experiment and the Phase 3 verify path** — `1bfadf1` (feat)
2. **Task 03-07-03: close carried Phase 2 debt (D-02-01, D-02-04, D-02-06)** — `6711022` (test)

**Plan metadata:** (this commit)

## Files Created/Modified

- `internal/compiler/debugmap/debugmap.go`, `debugmap_test.go` — the bounded debug-lineage side table
- `cmd/lang/main.go` — `debug-map` subcommand, `protocol.Human`'s new error return
- `internal/compiler/session/session.go` — `DebugMapCommandFile`, `verifyBorrowedCorpus`, `VerifyCorpus` dispatch
- `internal/compiler/session/session_test.go` — `TestVerifyPhase3ControlsAndWork`
- `internal/compiler/testsupport/cli_test.go` — `TestVerifyPhase3CLI`, `TestDebugMapCLIAvailability`, `TestPhase3VerifierScriptContract`
- `internal/compiler/protocol/protocol.go`, `protocol_test.go` — `DebugMapSummary`/`DebugMapEntry`, `Human`'s convergence-or-error fix, `TestHumanJSONProjectionParity`
- `internal/compiler/native/native_test.go` — hang-mode falsifier, AST-resolving spawn guard, evasion-planting test
- `internal/compiler/testsupport/testsupport.go`, `testsupport_internal_test.go` — tightened CLI stream ceiling, boundary test
- `scripts/verify-phase3.sh` — the Phase 3 gate
- `testdata/phase3/borrowed_view.lang` — the dispatch probe

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

### Documented Scope/Interpretation Deviations (not Rule 1-3 bug fixes)

**1. [Rule 4-style, documented] Tasks 1 and 2 landed in one commit, not two**
- **Plan text:** implied task-boundary commits.
- **Decision:** `cmd/lang/main.go`, `internal/compiler/session/session.go`, `internal/compiler/testsupport/cli_test.go`, and `internal/compiler/protocol/protocol.go` are each touched by both the debug-map CLI wiring (Task 1) and the Phase 3 verify path (Task 2). Splitting a single file's interleaved hunks across two commits after the fact would have required reconstructing intermediate states with no functional benefit — the same interleaving precedent 03-02's and 03-06's summaries document for their own single-file continuous edits.
- **Impact:** No functional difference; both tasks' own tests are fully present and passing in the resulting commit.

**2. [Rule 4-style, documented] borrowed_view.lang does not carry the public-borrowed-return case**
- **Plan text:** "exercising a branch with arm bodies, a shared loan ending on one edge, a sequential exclusive loan, and a public borrowed return, all in one canonical program."
- **Decision:** `check.Program`'s own `core.mixed_body_versions` gate (verified by reading `check.go` directly, not assumed) rejects a module that mixes a match-bodied function (the branch case) with a straight-line function (a public borrowed return, `-> borrow(path) Type`) — this is pre-existing, already-shipped behavior outside this plan's scope to change. `borrowed_view.lang` carries the branch/shared/exclusive cases only; `verifyBorrowedCorpus` exercises the three origin-summary controls against the existing `public_view*.lang` fixtures instead, in the same verify lane.
- **Impact:** All required controls are still enforced fail-closed by the gate; the "one canonical program" framing is not literally satisfied, but the phase's actual required-control set is.

**3. [Rule 4-style, documented] edge-last-use-omitted and loan-endpoint-mismatch share one control ID**
- **Plan text:** lists both as distinct required controls (`must_haves.truths`).
- **Decision:** Both are the same law under this codebase's current implementation — 03-04's `BorrowedLoanEndpointControlLane` already implements the omitted-edge-endpoint mutation and reports `control:core.loan_endpoint_mismatch`; adding a second control string to that lane's `Controls` slice would have broken its own already-shipped `TestBorrowedLaneRecordsFailureStatus`, which asserts `len(passLane.Controls) == 1`. This plan's `<verify>` block explicitly requires that exact test to still pass unchanged, so widening it was foreclosed by the plan's own regression requirement.
- **Impact:** No control coverage gap — the same mutation is what both named laws describe; only the control-ID cardinality differs from the plan's literal phrasing.

---

**Total deviations:** 0 auto-fixed bugs. 3 documented scope/interpretation decisions (Rule 4-style), made under the plan's own file-scope and regression constraints, not silently assumed correct.
**Impact on plan:** No scope creep. All required controls, all named test targets, and both gate-script prohibitions are satisfied; the deviations narrow implementation mechanics in ways forced by pre-existing, already-shipped code this plan was instructed not to widen into.

## Issues Encountered

- `debugmap.Build`'s first implementation assumed a naive 1:1 correspondence between `ast.LinearBody.Bindings[i]` and `core.LinearOperation[i]` for branch arm bodies, matching the already-verified straight-line case. `TestDebugMapJoinsSourceCoreAndOperation` failed immediately (a `move` kind never appeared; a `copy` appeared where a `borrow_shared` was expected) because `checkBranch` prepends one implicit per-arm alias-copy operation ahead of each arm's own bindings (03-01's per-arm aliasing decision). Fixed by reading `checkBranch` directly and skipping that one leading operation before joining. No broken version was committed.

## User Setup Required

None — no external service configuration required.

## Mutation-Kill / Non-Regression Evidence

- `git diff 7faf566..HEAD -- testdata/phase1/` is empty — no Phase 1 golden moved.
- `git diff 7faf566..HEAD -- go.mod go.sum` is empty — no new Go dependency.
- `sh scripts/verify-phase2.sh` exits 0 and reports all nine Phase 2 control IDs and both Phase 1 controls through a freshly built binary, unchanged.
- `sh scripts/verify-phase3.sh` exits 0 and reports all seven Phase 3 control IDs, both expected escapes, and 20 warm samples per lane through the same freshly built binary that also verified Phase 1 and Phase 2.
- `env GOCACHE=/tmp/ai-lang-phase3-cache go test ./...`, `go test -race ./...`, and `go vet ./...` all pass with zero findings.
- `TestSourceNeverSpawnsUnboundedProcesses` (the new AST-resolving scanner) re-run against the real repository finds zero violations across every production spawn site — the rewrite does not merely pass its own planted evasions, it still passes clean against the codebase it protects.
- `TestSpawnGuardCatchesKnownEvasions` catches all four D-02-01 evasions, planted fresh this session, and does not fire on an honestly bounded spawn shape.

## Known Stubs

None.

## Inherited Debt — carried forward intact, not fixed, per explicit instruction

**D-03-01** (03-05 mid-phase gate) and **D-03-02** (03-05 mid-phase gate) remain open, exactly as recorded in `03-DEBT.md`. This plan's own gate script (`scripts/verify-phase3.sh`) and its contract test do not assert or imply anything these two items contradict:

- The gate's `lane:path-oracle-disagreement` and `lane:borrowed-loan-endpoint-control` controls prove the CFG dataflow's own output agrees with two independent re-derivations (03-04's reachability-closure, 03-05's path oracle) — they do **not** claim the CFG dataflow governs admission. D-03-01's finding (the quadratic `discoverLoanLastUses`, not the linear CFG dataflow, is what actually decides every verdict) is unaffected and un-contradicted by anything this gate asserts.
- The gate does not add an `interface export` control asserting an exported borrow-derived return without a declared origin is rejected — D-03-02 is precisely the finding that no such rejection exists yet. Nothing in this plan's scope (Task 03-07-01/02/03's named files) touches `originvalidate.ValidatePublished`'s admission path, so widening into a fix was out of scope per the plan's own instruction not to chase D-03-01/D-03-02.

Both items are Phase 3 debt, distinct from the nine Phase 2 items already closed across this phase's plans (see the accounting table below). They are the phase verifier's to see next, not this plan's to resolve.

## Phase 2 Debt Closure Accounting (D-08)

| ID | Item | Closed by | Status |
|---|---|---|---|
| D-02-01 | Unbounded-spawn guard is a substring scan | **This plan** | Closed — AST-resolving scanner, all four evasions caught |
| D-02-02 | `evidence.canonical_unstable` unreachable at CLI | 03-06 | Closed |
| D-02-03 | Θ(N²) loan liveness (checker + validator) | 03-03 (validator-side partial), 03-04 (validator-side full) | Partially closed — see D-03-01 above; straight-line checker path remains quadratic and uncounted, carried as Phase 3 debt |
| D-02-04 | `native.timeout` has no falsifier | **This plan** | Closed — hang-mode falsifier for both deadlines |
| D-02-05 | `__LANG_` reserved-identifier violation | 03-01 | Closed |
| D-02-06 | 8 MiB CLI stream ceiling, never measured | **This plan** | Closed — tightened to 1 MiB with measured rationale |
| D-02-07 | Backend causality control coupled to one literal line | 03-01 | Closed |
| D-02-08 | `protocol.Human` diverges from `protocol.JSON` | **This plan** | Closed — identical convergence-or-error contract |
| D-02-09 | `Box`/`Pair` check but die spanless | 03-02 | Closed |

All nine Phase 2 carried debt items are now closed. Two new Phase 3 items (D-03-01, D-03-02) remain open, recorded above and in `03-DEBT.md`, for the phase verifier.

## Next Phase Readiness

- `scripts/verify-phase3.sh` is a real, runnable, fail-closed gate a human or the phase verifier can run directly: `sh scripts/verify-phase3.sh`.
- All nine Phase 2 carried debt items are closed. D-03-01 and D-03-02 are Phase 3 debt, explicitly not this plan's to fix, carried forward with full context for the phase verifier.
- Phase 1 and Phase 2 goldens and controls are unchanged; `go.mod` gained no dependency.
- No blockers for phase completion. REQUIREMENTS.md's OWN-03/OWN-04 completion is left to the orchestrator's `phase.complete` step, per this plan's explicit instruction not to mark them itself.

---
*Phase: 03-borrowed-views-and-cfg-lifetimes*
*Plan: 07*
*Completed: 2026-09-04*

## Self-Check: PASSED
