---
phase: 10-trusted-interprocedural-oracle
plan: 04
subsystem: interp
tags: [call-depth, stack-safety, subprocess-probe, mutation-testing, denial-of-service]

# Dependency graph
requires:
  - phase: 10-trusted-interprocedural-oracle
    provides: "plan 10-01's explicit heap []frame call stack in interp.go, whose O(1)-Go-stack property is exactly what this plan's subprocess probe measures"
provides:
  - "MaxCallDepth = 128, a genuinely reachable call-depth ceiling declared BELOW the real structural maximum (1024, syntax's maxFunctions combined with callgraph's cycle refusal), the deliberate inverse of pathoracle.MaxPaths' above-maximum direction"
  - "The depth-exceeded refusal as a typed, comparable Outcome/Event (callDepthExceededDefectReason), never a bare Go error -- reachable from interp.CanonicalBytes, byte-identical across repeated runs, with an outcome kind drawn from execution.TerminalOutcomeKinds()'s closed set"
  - "generateCallDepthChainSource/generateAndCheckCallDepthChain: a parameterized N-link .lang chain generator driven through the REAL pipeline (syntax.Parse -> check.Program -> corevalidate.Validate), documented in testdata/phase10/call_depth_chain_generator.md, reused by all three tasks"
  - "TestNativeStackHeadroomIndependentOfCallDepth: a subprocess probe (os/exec re-exec, runtime/debug.SetMaxStack) directly observing that the host-stack limit and the language-level call-depth bound are structurally unrelated -- SEM-08's Pitfall-4 gate, Roadmap criterion 2"
affects: [11]

# Actuals (#2632)
actuals:
  tokens: 6588
  tasks: 3
  commits: 3

# Tech tracking
tech-stack:
  added: []
  patterns:
    - "Inverted-direction sibling constants (D-10-23): MaxCallDepth (128, deliberately BELOW its real 1024 structural ceiling so it genuinely fires) documented explicitly as the opposite of pathoracle.MaxPaths (4096, deliberately ABOVE its real 64 reachable maximum so it never fires) -- the rationale comment states the inversion so a reader pattern-matching the two constants does not draw the wrong conclusion"
    - "A depth-exceeded refusal modeled exactly like the existing core.OpDefect arm: a typed Outcome/Event serializable into CanonicalBytes, never a Go error -- comparable execution DATA for Phase 11's five-axis comparator"
    - "A single generator helper (generateCallDepthChainSource/generateAndCheckCallDepthChain), parameterized by chain length N, reused across all three tasks' tests instead of three separately hand-copied fixture builders"
    - "First subprocess-probe machinery in this repository: os/exec re-exec of the test binary itself guarded by an env var, runtime/debug.SetMaxStack pinning a small deterministic host ceiling in the child, the parent asserting only on exit status and two independently bounded output writers (never .Output()/.CombinedOutput()/.StdoutPipe(), per this repo's own TestSourceNeverSpawnsUnboundedProcesses guard)"
    - "A named, source-committed probe-depth multiplier (probeCallDepthMultiplier) expressed against a reduced in-test cap (probeReducedCallDepthCap) rather than the production MaxCallDepth, kept well under the parser's maxFunctions ceiling"

key-files:
  created:
    - testdata/phase10/call_depth_chain_generator.md
  modified:
    - internal/compiler/interp/interp.go
    - internal/compiler/interp/interp_test.go

key-decisions:
  - "The depth check runs AFTER partitionFrameForCall resolves the callee, immediately before the actual stack push -- never before -- so a bare match-arm callee (which needs no frame at all) is never wrongly refused for depth it would not have consumed"
  - "Depth is measured as total frame-stack size (base frame counts as depth 1): a chain of exactly MaxCallDepth functions therefore peaks at MaxCallDepth frames and returns normally, while MaxCallDepth+1 functions requires one push beyond the ceiling and refuses -- this is the exact semantics documented in testdata/phase10/call_depth_chain_generator.md so a future reader does not have to re-derive it from the enforcement site"
  - "The subprocess probe's cap-disabled arm and cap-enabled arm both drive the IDENTICAL probeChainDepth (800) chain under the identical pinned 1 MiB host stack ceiling -- only the seam differs (override raises the cap vs. production MaxCallDepth governs) -- so the two arms are a true minimal-pair comparison rather than two differently-shaped experiments"
  - "probeChainDepth (800 = 8 x 100) stays comfortably under syntax's maxFunctions (1024) by expressing the multiplier against a reduced in-test cap (8) rather than the production MaxCallDepth (128), since 100x128 would have exceeded the parser's own ceiling"

patterns-established:
  - "A same-package unexported override seam (maxCallDepthOverride) that mirrors an existing exported precedent's discipline (TerminatorKindsOverride) but is deliberately kept unexported when the plan's own must_haves require it, rather than copying the precedent's visibility verbatim"

requirements-completed: [SEM-08]

coverage:
  - id: D1
    description: "MaxCallDepth = 128 is declared with its inversion relative to pathoracle.MaxPaths stated explicitly in source, and exceeding it produces a named refusal as comparable Outcome/Event data reachable from interp.CanonicalBytes"
    requirement: SEM-08
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestCallDepthAtAndOverTheCap"
        status: pass
      - kind: other
        ref: "grep -c 'MaxCallDepth = 128' internal/compiler/interp/interp.go"
        status: pass
    human_judgment: false
  - id: D2
    description: "A genuine 129-function chain, generated and driven through syntax.Parse -> check.Program -> corevalidate.Validate -> interp.Run (never a hand-built core.Program), reaches the named depth-exceeded refusal"
    requirement: SEM-08
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestCallDepthExceeded"
        status: pass
      - kind: other
        ref: "grep -c 'testsupport' internal/compiler/interp/interp_test.go (returns 0)"
        status: pass
    human_judgment: false
  - id: D3
    description: "A subprocess probe directly observes that the host-stack limit and the language-level call-depth bound are structurally unrelated limits, with its threshold constants fixed in source before the first passing run"
    requirement: SEM-08
    verification:
      - kind: unit
        ref: "internal/compiler/interp/interp_test.go#TestNativeStackHeadroomIndependentOfCallDepth"
        status: pass
      - kind: other
        ref: "go test -race ./internal/compiler/interp/... (no DATA RACE); go test ./internal/compiler/interp/... -count=2 (exits 0)"
        status: pass
    human_judgment: true
    rationale: "The must_haves flagged this edge as UNRESOLVED at plan-authoring time: 'the deterministic edge probe could not classify SEM-08's edge category... a reviewer must confirm no further edge exists' beyond the at/over-the-cap boundary this plan covers. A human reviewer should confirm no other call-depth edge case is left unconsidered."

duration: 45min
completed: 2026-09-11
status: complete
---

# Phase 10 Plan 4: Trusted Interprocedural Oracle — Call-Depth Ceiling and the Pitfall-4 Probe Summary

`interp.MaxCallDepth = 128` gives SEM-08 a genuinely reachable, fail-closed call-depth ceiling — declared deliberately BELOW the real 1024-function structural maximum, the exact inverse of how `pathoracle.MaxPaths` is set ABOVE its own — with the refusal modeled as comparable `Outcome` data and a subprocess probe directly proving the host-stack limit and the language-level bound are two structurally unrelated limits.

## Performance

- **Duration:** 45 min
- **Started:** 2026-09-11 (see git commit timestamps)
- **Completed:** 2026-09-11
- **Tasks:** 3 completed
- **Files modified:** 3 (2 modified, 1 created)

## Accomplishments

- `MaxCallDepth = 128` is declared in `interp.go` with a rationale comment that states its inversion relative to `pathoracle.MaxPaths` explicitly in source: `MaxPaths` (4096) sits well above its real reachable maximum (64) so it never fires today, while `MaxCallDepth` sits below the real structural ceiling (1024, `syntax/parser.go`'s `maxFunctions` combined with `callgraph`'s cycle refusal making every admitted call graph a DAG) so it genuinely fires on a program the compiler legitimately admits. The cap is enforced in `runFrameStack`'s `core.OpCall` arm, checked immediately before the actual stack push (after `partitionFrameForCall` resolves the callee, so a bare match-arm callee that needs no frame is never wrongly refused).
- The depth-exceeded refusal is modeled exactly like the existing `core.OpDefect` arm: a typed `Outcome{Kind: execution.OutcomeDefect}` / `Event` pair carrying the stable named reason `callDepthExceededDefectReason`, reachable from `interp.CanonicalBytes`, with an outcome kind drawn from `execution.TerminalOutcomeKinds()`'s closed set and byte-identical `CanonicalBytes` across two runs of the same over-depth program — comparable execution DATA, never a bare Go error, for Phase 11's five-axis comparator.
- `generateCallDepthChainSource`/`generateAndCheckCallDepthChain` generate a genuine N-link `.lang` chain (`link0` calls `link1` ... calls `link(N-1)`, the base case) and drive it through the REAL pipeline — `syntax.Parse` → `check.Program` → `corevalidate.Validate` — asserting a clean result at each stage, never a hand-built `core.Program`. The generator's full contract (module shape, per-link template, entry point, depth semantics, and the generate-vs-commit rationale) is recorded in `testdata/phase10/call_depth_chain_generator.md`, reused by `TestCallDepthAtAndOverTheCap` (N = 128 and N = 129), `TestCallDepthExceeded` (N = 129, asserting each pipeline stage clean before the interpreter refuses), and Task 3's probe.
- `TestNativeStackHeadroomIndependentOfCallDepth` is a genuine SUBPROCESS probe (never an in-process headroom-ratio measurement, which is rejected as this gate's evidence) — the first subprocess/`SetMaxStack`/`TestMain` re-exec machinery anywhere in this repository. The parent re-execs the test binary via `exec.CommandContext` with a real deadline, guarded by an env var, capturing stdout/stderr through independently bounded writers (never `.Output()`/`.CombinedOutput()`/`.StdoutPipe()`, per this repo's own `TestSourceNeverSpawnsUnboundedProcesses` guard). The child pins `runtime/debug.SetMaxStack` to a small, fixed 1 MiB ceiling and drives an identical 800-function chain under two arms: cap-disabled (the `maxCallDepthOverride` seam lifts the cap; the chain completes with no host collapse) and cap-enabled (the cap stays at `MaxCallDepth`; the named refusal fires first). Because `interp`'s call stack is an explicit heap `[]frame` slice (D-10-21), the result is the honest, stronger finding this plan requires: language call depth consumes O(1) host stack, so the two limits are structurally unrelated, never one limit wearing two names.
- `maxCallDepthOverride` is an unexported, nil-default override seam consulted through `maxCallDepth()`, following `pathoracle.TerminatorKindsOverride`'s discipline but deliberately kept unexported (per the plan's own must_have) since only a same-package test needs to raise or disable the cap for the subprocess probe.

## Task Commits

Each task was committed atomically:

1. **Task 1: MaxCallDepth, the inverted rationale, and the depth-exceeded Outcome** - `e5a27e4` (feat)
2. **Task 2: The 129-function chain, generated as real source and run through the real pipeline** - `b496113` (test)
3. **Task 3: The Pitfall-4 subprocess probe — two structurally unrelated limits** - `b7b8885` (test)

## Files Created/Modified

- `internal/compiler/interp/interp.go` — new `MaxCallDepth` const with its inversion rationale, `callDepthExceededDefectReason`, `maxCallDepthOverride`/`maxCallDepth()` seam, and the depth-check enforcement inserted into `runFrameStack`'s `core.OpCall` arm.
- `internal/compiler/interp/interp_test.go` — `generateCallDepthChainSource`, `generateAndCheckCallDepthChain`, `TestCallDepthAtAndOverTheCap`, `TestCallDepthExceeded`, `probeBoundedWriter`, `runCallDepthProbeSubprocess`, `runCallDepthProbeChild`, `TestNativeStackHeadroomIndependentOfCallDepth`, and the probe's named constants (`probeReducedCallDepthCap`, `probeCallDepthMultiplier`, `probeChainDepth`, `probeMaxStackBytes`, etc.).
- `testdata/phase10/call_depth_chain_generator.md` — new committed contract documenting the generator's module shape, per-link template, entry point, depth semantics, and generate-vs-commit rationale.

## Decisions Made

See `key-decisions` in frontmatter.

## Deviations from Plan

None — plan executed exactly as written. Every acceptance criterion in all three tasks passed on the first implementation attempt with no fix-up commits required.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- SEM-08's full "bounded call stack with a named refusal" requirement (deferred from plan 10-01, which proved only the O(1)-Go-stack shape making this refusal reachable at all) is now complete: a genuinely reachable ceiling, a comparable refusal, a real-pipeline-admitted over-depth program, and a subprocess-observed structural-independence proof.
- The depth-exceeded `Outcome`/`Event` shape (`callDepthExceededDefectReason`, drawn from `execution.TerminalOutcomeKinds()`) is now available as comparable execution data for Phase 11's five-axis comparator when diffing against native's own, structurally different limit-hit behavior.
- The must_haves' own flagged assumption remains UNRESOLVED, not dismissed: "the deterministic edge probe could not classify SEM-08's edge category... a reviewer must confirm no further edge exists" beyond the exactly-at-cap / one-over-cap boundary this plan covers. A human reviewer should confirm no other call-depth edge case is unconsidered.
- The pre-existing `originvalidate` unbounded-spawn defect (logged in `deferred-items.md` from plan 10-03) is unchanged by this plan — `go test ./internal/compiler/native/... -run TestSourceNeverSpawnsUnboundedProcesses` still fails on `main` for that one pre-existing reason only; this plan's own new subprocess-spawning test code does not appear in its violation list.
- Ready for plan 10-05.

## Self-Check: PASSED

- `internal/compiler/interp/interp.go` — FOUND
- `internal/compiler/interp/interp_test.go` — FOUND
- `testdata/phase10/call_depth_chain_generator.md` — FOUND
- Commit `e5a27e4` — FOUND (`git log --oneline --all`)
- Commit `b496113` — FOUND
- Commit `b7b8885` — FOUND
- `go test ./internal/compiler/interp/... -v` — PASS (all named: TestMoveAsCopyMutationKilled, TestCallExecutesAcrossOneFrame, TestInterpDoesNotReadCorevalidateOwnershipFields, TestCallDepthAtAndOverTheCap, TestCallDepthExceeded, TestNativeStackHeadroomIndependentOfCallDepth)
- `go test -race ./internal/compiler/interp/...` — PASS, no DATA RACE
- `go test ./internal/compiler/interp/... -count=2` — PASS
- `go build ./... && go vet ./...` — PASS
- `grep -c 'MaxCallDepth = 128' internal/compiler/interp/interp.go` — prints `1`
- `go test ./internal/compiler/...` — all pass except the pre-existing, previously-logged `TestSourceNeverSpawnsUnboundedProcesses` failure in `internal/compiler/native` (originvalidate's unbounded spawn, unrelated to this plan)

---
*Phase: 10-trusted-interprocedural-oracle*
*Completed: 2026-09-11*
