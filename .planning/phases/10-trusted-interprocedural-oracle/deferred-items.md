# Deferred Items — Phase 10

Out-of-scope discoveries found during plan execution, logged rather than fixed
(scope boundary: only auto-fix issues directly caused by the current task's
changes).

## From Plan 10-03

- **`internal/compiler/originvalidate/originvalidate_test.go`'s
  `transitiveImportsViolation` spawns `go list -deps` via a bare
  `exec.Command` and collects output via the unbounded `.Output()` merge —
  both forbidden by `internal/compiler/native/native_test.go`'s
  `TestSourceNeverSpawnsUnboundedProcesses` (D-02-01).** Landed in Plan
  10-02 (commit `7d7553b`), pre-dating this plan. Plan 10-03 fixed the
  identical pattern it introduced in
  `internal/compiler/pathoracle/pathoracle_test.go` (commit `9241354`,
  `exec.CommandContext` + a bounded stdout writer) but left
  originvalidate's own pre-existing instance untouched, since it was not
  introduced by this plan's changes. `go test ./internal/compiler/native/...
  -run TestSourceNeverSpawnsUnboundedProcesses` currently fails on `main`
  for this one reason. A future plan (or a small standalone fix) should
  apply the same `exec.CommandContext` + bounded-writer pattern to
  originvalidate's copy.

### RESOLVED at the Wave 2 post-merge gate

Fixed by the orchestrator during Phase 10's Wave 2 post-merge test gate
(`TestSourceNeverSpawnsUnboundedProcesses` was failing on `main`, blocking the
gate). `transitiveImportsViolation` in `originvalidate_test.go` now uses
`exec.CommandContext` with a 2-minute deadline and a bounded stdout writer,
matching the pattern `pathoracle_test.go` established in commit `9241354`.
`go test ./internal/compiler/native/... -run TestSourceNeverSpawnsUnboundedProcesses`
passes.

## From Plan 10-07

- **`corevalidate.peerDeriveOriginFacts` (`internal/compiler/corevalidate/corevalidate.go`)
  has no `core.OpCall` case, so ANY function declaring a borrow-returning
  `PublicOrigin` whose return value is sourced from forwarding a callee's
  own result (rather than a direct local `borrow`) is unconditionally
  reported as NOT `Callable` by `corevalidate`'s own independent origin
  peer -- regardless of whether the callee's own body actually returns a
  borrow, and regardless of what the calling function does afterward.**
  Discovered while attempting Task 1's ideal accept fixture (a genuine
  3-hop declared-borrow carry, forwarded leaf-to-f2-to-f1-to-caller with no
  conflicting use): `check.Program` admits that shape with ZERO
  diagnostics, but `session.CheckCommandFile`'s next consult,
  `corevalidate.Validate`, independently refuses it with
  `core.callee_not_callable` for the middle hop. Root cause confirmed by
  direct inspection of `peerDeriveOriginFacts`
  (`corevalidate.go:2324-2354`): its forward walk only ever extends
  `derived[...]` through `core.OpBorrowShared`/`core.OpBorrowExclusive`/
  `core.OpMove`/`core.OpCopy` -- there is no `core.OpCall` arm at all, so a
  call's own `TargetID` never enters `derived`, and `peerOriginContained`
  (which consults `peerDeriveOriginFacts` for any function with a declared
  `PublicOrigin`) always reports "not derived" for such a function.
  **This is NOT specific to the new depth-3 fixture: `relay_depth2_refuse.lang`
  (Phase 08, unmodified) is refused for the SAME independent reason today
  -- confirmed directly by removing its own `take buffer` conflict and
  re-running `corevalidate.Validate` in isolation, which still reports
  `core.callee_not_callable` for `relay`.** It has simply never been
  OBSERVABLE on that fixture, because `check`'s own
  `check.interprocedural_loan_liveness` refusal fires first and
  `CheckCommandFile`'s fixed check-then-peer precedence never reaches
  corevalidate for it. Structurally analogous to D-10-28 (a second
  detector missing a `core.OpCall` case) but for corevalidate's ORIGIN
  peer rather than its `usesParam` peer. Not fixed here: `corevalidate.go`
  is outside plan 10-07's `files_modified`, and widening
  `peerDeriveOriginFacts` to consult a callee's own declared/derived
  origin fact (mirroring `originvalidate.walkReturnOrigin`'s own
  `core.OpCall` case, or `corevalidate_peer_liveness.go`'s own
  `derivePeerLoanCarry` postorder-consult pattern) is a genuine,
  independent semantic change with its own blast radius, not a fixture
  fix. Reported per plan 10-07 Task 1's own escape-hatch instruction rather
  than silently falling back to an owned pass-through; see
  `testdata/phase10/relay_depth3_accept.lang`'s own header for the full
  empirical trail. A future plan (candidate: wherever D-10-28 lands, since
  both are "corevalidate's peer missing an OpCall case") should widen
  `peerDeriveOriginFacts` and add a companion mutation-kill test proving
  the new case load-bearing, mirroring `TestOpCallOriginWalkGateIsLoadBearing`'s
  own precedent for `originvalidate`'s equivalent case.

## From Plan 10-08

- **`corevalidate.Result.LoanEndpoints()` (plan 10-08's own Task 1
  accessor) is naturally INCOMPLETE for a branched (CFG-carrying) function
  when `corevalidate.Validate` refuses the program for a reason unrelated
  to that function's own loan-endpoint recomputation, discovered by
  `testdata/phase3/branch_one_arm_shared_reject.lang` while building
  plan 10-08 Task 2's three-way endpoint comparator.** `corevalidate.Validate`
  is a fail-fast, first-problem-wins replayer (`v.check` short-circuits
  `v.run()` the instant any problem is recorded), unlike `check`'s own
  `materializeLoanEndpoints` or `pathoracle.RecomputeEndpoints`, neither of
  which ever short-circuits on an unrelated problem elsewhere in the
  program. A branched function corevalidate refuses for a reason caught
  during an earlier structural or per-operation replay step -- before
  `v.loanEndpointsMatch` is ever reached for that function -- legitimately
  has an empty entry in `Result.LoanEndpoints`, diverging from `check`'s
  and `pathoracle`'s structurally-complete sets, for a reason that has
  nothing to do with any of the three peers' endpoint-derivation logic
  disagreeing. Not a bug in any of the three peers' own loan-endpoint
  algorithms. Plan 10-08's own three-way comparator (`session_peer_gate_test.go`'s
  `assertThreeWayEndpointAgreement`) is scoped around this: it skips the
  comparison only for a CFG-carrying function whose program corevalidate
  did not fully validate (`!validated.Valid`), stated in that function's
  own doc comment rather than papered over. Not fixed here (no code change
  is possible without altering `corevalidate.Validate`'s own fail-fast
  architecture, a genuine, independent semantic change outside this plan's
  scope) -- a future plan wanting a COMPLETE per-function endpoint record
  from a refused program would need `corevalidate` to keep validating past
  its first problem (or record endpoints eagerly, before any check that
  could short-circuit), which is a materially different validator
  architecture, not a narrow fix.
