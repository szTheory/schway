---
phase: 18-branch-on-a-computed-value
plan: 09
subsystem: testing
tags: [go, clang, subprocess, validation-evidence, race]
plan_head_before: 517a0acb4b1fb6fddb150031cb67173cdd5c3e3d
commits: 4

requires:
  - phase: 18-08
    provides: Measured validation contract and the recorded CI disposition
provides:
  - Finite 30-second default subprocess budgets for Clang probes and native execution
  - Typed inner Clang probe failures preserved behind cache.input_undeclared
  - Seven independently captured post-fix full-suite, race, vet, and build receipts
affects: [phase-18-verification, CTL-01, CTL-02, CTL-03]

actuals:
  tokens: 1415953
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns:
    - Keep fail-closed public cache errors stable while exposing typed causes through Unwrap
    - Bind verification receipts to distinct captured output sections with byte ranges and SHA-256 digests

key-files:
  created:
    - .planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh
    - .planning/phases/18-branch-on-a-computed-value/18-09-RUNS.log
  modified:
    - internal/compiler/cache/cache.go
    - internal/compiler/cache/probe.go
    - internal/compiler/cache/probe_test.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_test.go
    - internal/compiler/measure/machine.go
    - internal/compiler/measure/machine_test.go
    - .planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md
    - testdata/phase16/validation-corpus-run-record.jsonl
    - testdata/phase16/validation-corpus-run-record.manifest.json

key-decisions:
  - Keep the existing cache.input_undeclared refusal code and Error() text while unwrapping its probe cause.
  - Use a finite 30-second default per subprocess; caller-provided runner timeouts and parent context deadlines remain authoritative.
  - Keep Plan 08's CI disposition unchanged because existing macOS and Linux full and race jobs cover the phase.
  - Regenerate the stale validation corpus from the live pair exporter and actual sequential test results.

patterns-established:
  - "Typed refusal: retain the stable outer refusal while exposing the originating typed probe error with errors.As."
  - "Evidence receipt: identify each run and bind its complete log section by offset, length, and SHA-256."

requirements-completed: [CTL-01, CTL-02, CTL-03]

coverage:
  - id: D1
    description: Cache probe failures remain fail-closed and inspectable; native and Clang subprocesses retain finite deadlines under package load.
    requirement: CTL-01
    verification:
      - kind: unit
        ref: "GOCACHE=/tmp/ai-lang-gocache go test ./internal/compiler/cache ./internal/compiler/native ./internal/compiler/cgen -run 'Test(CacheKeyCoversEveryDeclaredInput|ClangDigestProbeIsBounded|CacheProbeFailureCause|NativeTimeoutHasFalsifier|PayloadTracerThreeEngineAgreement)$' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: Three independent default-parallel full suites, capped-parallel full suite, race suite, vet, and build have matching captured receipts; Plan 08's CI disposition and workflow blob remain unchanged.
    requirement: CTL-02
    verification:
      - kind: integration
        ref: "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh --final"
        status: pass
      - kind: integration
        ref: "Plan 08 verify-phase18-validation-evidence.sh --final"
        status: pass
    human_judgment: false

duration: 63min
completed: 2026-09-25
status: complete
---

# Phase 18 Plan 09: Default-Parallel Compiler Evidence Summary

**Cache and native probes retain inspectable failures and finite 30-second subprocess budgets; seven full-suite and build lanes now pass with captured evidence.**

## Performance

- **Duration:** 63 minutes
- **Started:** 2026-09-25T17:48:50Z
- **Completed:** 2026-09-25T18:51:55Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments

- `InputsFor` keeps returning `cache.input_undeclared` after a failed Clang identity probe while preserving the typed timeout, launch, or command failure for `errors.As`.
- Cache probes and default native subprocesses use finite 30-second budgets. Explicit runner timeouts and parent-context cancellation remain effective; short timeout controls still prove fail-closed behavior.
- Captured seven passing post-fix lanes: three default-parallel full suites (282.723s, 232.755s, 275.610s), `-p=4` full suite (243.863s), race suite (198.482s), vet (0.425s), and build (0.252s).
- Plan 08's final evidence verifier passes, and `.github/workflows/ci.yml` retains the recorded blob and `not_added` disposition.
- Refreshed the stale validation-corpus run record from the current consumer-derived 33-pair set using actual sequential test results and completion witnesses.

## Task Commits

1. **Task 1 RED: Add cache probe cause controls** - `1c6884c` (`test`)
2. **Task 1 GREEN: Bound compiler subprocesses and retain probe causes** - `f2398f1` (`fix`)
3. **Task 2 deviation: Keep the machine probe bounded under race load** - `f8477e8` (`fix`)
4. **Task 2: Record repeated full-suite evidence** - `5cf3d3f` (`fix`)

**Plan metadata:** recorded in the final GSD close-out commit.

## Files Created/Modified

- `internal/compiler/cache/cache.go` - Added optional typed cause unwrapping without changing the stable code string.
- `internal/compiler/cache/probe.go` - Increased the finite cache probe ceiling and retained probe failures on refusal.
- `internal/compiler/cache/probe_test.go` - Added timeout, process failure, and unresolved-command controls through `InputsFor`.
- `internal/compiler/native/native.go` - Set a finite 30-second default per subprocess while preserving explicit `Runner.Timeout` values.
- `internal/compiler/native/native_test.go` - Kept compile/run hang falsifiers and pinned the default runner budget.
- `internal/compiler/measure/machine.go` and `machine_test.go` - Raised the bounded Clang machine probe ceiling and kept its timeout falsifier caller-controlled.
- `.planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh` - Added `--record` and `--final` modes for seven exact lanes and receipt validation.
- `.planning/phases/18-branch-on-a-computed-value/18-09-RUNS.log` - Captured full output for both the initial diagnostic attempt and the passing post-fix attempt.
- `.planning/phases/18-branch-on-a-computed-value/18-VALIDATION.md` - Appended separate `phase18-postfix` receipts; the original Plan 08 `text` block was preserved.
- `testdata/phase16/validation-corpus-run-record.jsonl` and `validation-corpus-run-record.manifest.json` - Rebound the existing corpus snapshot to the live validation rows and actual 33-pair run.

## Decisions Made

- Kept the public cache refusal code and string stable, with the underlying probe `Error` available through `Unwrap`.
- Retained finite subprocess limits at 30 seconds and preserved short caller-controlled timeout behavior.
- Used existing macOS/Linux full and race CI jobs for recurring coverage; no focused CI command was added.
- Regenerated the validation corpus after its current pair digest was shown stale against the live validation table.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Refreshed the stale validation-corpus snapshot**
- **Found during:** Task 2 full-suite capture
- **Issue:** The session test rejected the archived pair digest after Phase 18's validation table changed in the prior diagnostic commit.
- **Fix:** Exported the exact live pair list, ran the existing sequential producer for all 33 pairs, and updated the record and manifest from the generated bytes.
- **Files modified:** `testdata/phase16/validation-corpus-run-record.jsonl`, `testdata/phase16/validation-corpus-run-record.manifest.json`
- **Verification:** Corpus grade and completion-witness tests passed; all subsequent full and race suites passed.
- **Committed in:** `5cf3d3f`

**2. [Rule 3 - Blocking] Kept the machine probe falsifier green under race load**
- **Found during:** Task 2 initial race-suite capture
- **Issue:** The bounded 5-second Clang machine probe expired before its output-truncation control completed under `-race`.
- **Fix:** Raised the finite default probe limit to 30 seconds and changed the timeout falsifier to use an explicit 50-millisecond parent deadline.
- **Files modified:** `internal/compiler/measure/machine.go`, `internal/compiler/measure/machine_test.go`
- **Verification:** `GOCACHE=/tmp/ai-lang-gocache go test -race ./internal/compiler/measure -run '^TestMachineProbeIsBounded$' -count=1` passed; the full race lane passed.
- **Committed in:** `f8477e8`

**3. [Rule 1 - Bug] Corrected output-section marker validation**
- **Found during:** Task 2 final evidence check
- **Issue:** The final checker omitted the leading separator newline included in captured section offsets.
- **Fix:** Matched the complete logged header at the recorded section start.
- **Files modified:** `verify-phase18-postfix-evidence.sh`
- **Verification:** `bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh --final` passed.
- **Committed in:** `5cf3d3f`

**Total deviations:** 3 auto-fixed (two Rule 3 blockers and one Rule 1 checker bug).
**Impact on plan:** The fixes were required to establish the requested full-suite and race evidence. Plan 08's measured record and CI decision were preserved.

## Issues Encountered

- The first seven-lane attempt recorded the stale corpus failure, one race-load machine-probe timeout, and vet/build cache access denied under the default home cache. The recorder was updated to set `GOCACHE=/tmp/ai-lang-gocache` for vet/build while preserving their exact command arguments. Its failure receipts remain in the log; the latest complete attempt is the passing one checked by `--final`.
- The sandbox denied Git metadata writes on the initial commit attempt. The GSD task commits succeeded with repository metadata access; unrelated pre-existing workspace files were not staged.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

G-18-16 is closed with repeated default-parallel evidence, bounded subprocesses, inspectable cache probe causes, and the original Plan 08 CI contract intact. Phase 18 is ready for the orchestrator's phase verification.

---
*Phase: 18-branch-on-a-computed-value*
*Completed: 2026-09-25*

## Self-Check: PASSED

- Summary, evidence verifier, and run log exist.
- Task commits `1c6884c`, `f2398f1`, `f8477e8`, and `5cf3d3f` exist in Git history.
