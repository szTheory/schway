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
