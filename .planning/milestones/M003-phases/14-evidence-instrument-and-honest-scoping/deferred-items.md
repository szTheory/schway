# Deferred Items — Phase 14

Out-of-scope discoveries found during plan execution, logged per the executor's
scope-boundary rule (do not auto-fix; log and continue).

## From 14-02 (diagnostic distinctness)

- **`TestSourceNeverSpawnsUnboundedProcesses` (`internal/compiler/native/native_test.go:616`) fails repo-wide `go test ./...`, pre-existing and unrelated to this plan's diff.**
  Confirmed to already fail at commit `7bd3e7d` (the tip of plan 14-01, before
  plan 14-02's first commit) via `git show 7bd3e7d:...` + isolated test run.
  Two violating files, neither touched by 14-02:
  - `.claude/skills/spike-findings-ai-lang/sources/005-native-ffi-provenance-cleanup/lab/lab.go`
    — a point-in-time spike-source snapshot committed at `58103f5`
    (`docs(spike-wrap-up): package 6 spike findings into project skill`), uses
    `exec.Command` without a context deadline and `CombinedOutput()`
    (unbounded buffer).
  - `internal/compiler/session/verification_groundedness_test.go` — added by
    plan 14-01 (`f71aba8`), the groundedness lint's static test-index
    resolver; uses `exec.Command` without a context deadline and `Output()`
    (unbounded buffer) in its `TestStaticTestIndexMatchesGoTestList` accuracy
    control (D-14-10), which intentionally shells out to real
    `go test ./... -list .` once.
  Neither file is in 14-02's `files_modified` (`internal/compiler/syntax/parser.go`,
  `internal/compiler/check/diagnostic_distinctness_test.go`,
  `testdata/distinctness/`). Per the scope-boundary rule this is logged, not
  fixed, by 14-02. All of 14-02's own scoped verify commands pass:
  `go test ./internal/compiler/syntax/... ./internal/compiler/diagnostic/... ./internal/compiler/check/... -count=1`
  is green, and `go test ./...` reports exactly this one pre-existing,
  unrelated failure — no new failure was introduced by 14-02.
  **Recommendation:** a follow-up plan should either add a context deadline +
  bounded writer to both call sites, or add a scoped exemption to
  `TestSourceNeverSpawnsUnboundedProcesses` for `.claude/skills/**/sources/**`
  (frozen point-in-time snapshots) and reconsider the groundedness lint's own
  accuracy-control shell-out.
  status: resolved
  **Resolution (verified 2026-09-18, at the close of Phase 14).** Both call
  sites are closed, and `unboundedSpawnAllowlist` is still empty — neither was
  exempted away:
  - `.claude/skills/.../sources/005-.../lab/lab.go` — `5318357`
    (`fix(14): restore the unbounded-spawn guard to green at the wave-1 gate`)
    widened the guard's throwaway-lab exclusion from live labs under
    `.planning/spikes/` to also cover labs archived verbatim into the
    spike-findings skill under `.claude/skills/*/sources/`. Per the guard's own
    doc comment the exclusion's rationale always covered them; only its path
    literal did not.
  - `internal/compiler/session/verification_groundedness_test.go` — `4f4d773`
    (`test(14-06): add failing test for the re-pinned frontier and corpus
    floors`) moved `TestStaticTestIndexMatchesGoTestList` to the project's
    bounded shape: `exec.CommandContext` rooted at
    `context.WithTimeout(context.Background(), staticIndexAccuracyControlTimeout)`
    with a size-capped reader, rather than bare `exec.Command` + `Output()`.

  Confirmed by running `go test ./internal/compiler/native/ -run
  'TestSourceNeverSpawnsUnboundedProcesses' -count=1` (ok, 0.382s) and a full
  `go test ./... -count=1` (exit 0, 25 packages) on the phase-completion tree.
