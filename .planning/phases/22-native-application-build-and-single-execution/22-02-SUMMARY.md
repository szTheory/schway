---
phase: 22-native-application-build-and-single-execution
plan: 02
subsystem: native
tags: [Go, C17, Clang, local-c, ABI, build-identity]
requires:
  - phase: 22-01
    provides: Retained application builds, checked U64 entry, receipt integrity and single-child execution
provides:
  - Closed local C manifests with confined input snapshots and isolated header/typedef symbol probes
  - Canonical input and artifact identities with explicit incomplete dependency closure
  - Relocatable public CLI fixture with mutation, stale-receipt and zero/one-launch controls
affects: [22-03, phase-23, FFI-02, APP-02]
actuals:
  tokens: 16904
  tasks: 2
  commits: 2
plan_head_before: 40dc5575b010de9d8c431d5408615d0c9426801a
tech-stack:
  added: []
  patterns:
    - Compile private snapshots of declared local inputs through fixed argv
    - Check every symbol in an isolated translation unit against its declared header
    - Separate canonical known-input identity from final artifact byte identity
key-files:
  created:
    - internal/compiler/native/bindings.go
    - internal/compiler/native/bindings_test.go
    - examples/phase22/identity.bindings.json
    - examples/phase22/support.h
    - examples/phase22/support.c
    - examples/phase22/BINDINGS.md
  modified:
    - cmd/lang/main.go
    - internal/compiler/session/session.go
    - internal/compiler/native/native_app.go
key-decisions:
  - "Use the existing session.go application entry from 22-01 rather than introduce a duplicate session_app.go."
  - "Compile each symbol probe separately so another declared header cannot supply a missing declaration."
  - "Keep macOS and Linux dependency closure incomplete and cacheability false; input_id is relocation-stable while build_id also binds final binary bytes."
patterns-established:
  - "ABI manifests grant trusted build authority without admitting Lang foreign calls."
  - "Input snapshots bind hashed bytes to compiled bytes after resolution."
requirements-completed: [APP-02, FFI-02]
coverage:
  - id: D1
    description: "Explicit local C inputs are confined, compiled and linked only after closed manifest and header/typedef validation."
    requirement: FFI-02
    verification:
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsBuild"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsRejectInvalidInputs"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsClosedJSON"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsIgnoreAmbientIncludePath"
        status: pass
    human_judgment: false
  - id: D2
    description: "Declared source, ABI and toolchain facts invalidate identity, and swapped artifacts or stale receipts are refused before launch."
    requirement: FFI-02
    verification:
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BuildIdentityDeclaredInputMutations"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BuildIdentityReceiptAndBinaryMismatch"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsCanonicalManifestAndSnapshot"
        status: pass
    human_judgment: false
  - id: D3
    description: "The checked-in identity fixture builds through the public CLI after relocation and runs after all source and intermediate files are removed."
    requirement: APP-02
    verification:
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22RelocatedBindingsCLI"
        status: pass
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsNestedHeadersAndSpacedPaths"
        status: pass
    human_judgment: false
  - id: D4
    description: "A local-C build causes zero application effects and its requested run launches exactly once."
    requirement: APP-02
    verification:
      - kind: integration
        ref: "internal/compiler/native/bindings_test.go#TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce"
        status: pass
    human_judgment: false
duration: 21 min
completed: 2026-09-27
status: complete
---

# Phase 22 Plan 02: Local C Build Authority and Identity Summary

**Relocatable local C manifests, isolated ABI probes, retained artifact identities, and executable input/receipt invalidation controls**

## Performance

- **Duration:** 21 min
- **Started:** 2026-09-27T18:19:59Z
- **Completed:** 2026-09-27T18:41:26Z
- **Tasks:** 2
- **Files modified:** 9 production/test/example files, plus this summary
- **Actuals:** 16,904 tokens measured as realized diff characters / 4, rounded up; 2 task commits measured from the persisted plan ledger before this metadata commit.

Three samples per lane, milliseconds, from the final focused test run. First-use
means a new artifact output path or first execution of its retained artifact;
OS caches were not flushed. These are process-cold observations, not cold-machine
distributions, performance thresholds, or cross-host evidence.

| Operation | First-use samples (ms) | Repeated samples (ms) |
|---|---|---|
| Manifest build | 269.597, 261.587, 247.653 | 262.262, 248.696, 240.669 |
| Run and observe exact stdout | 3.563, 3.364, 3.294 | 3.853, 3.280, 3.417 |

## Accomplishments

- Added `--manifest` through the CLI/session/native build path, a bounded closed `lang.local-c/1` parser, path/symlink confinement, private byte snapshots, and full local include checks using Clang dependencies.
- Added isolated C17 function-typedef probes and live link references, canonical manifest/inventory/toolchain identities, final artifact byte binding and receipt consistency checks. Unknown host closure remains incomplete and non-cacheable.
- Shipped the local C example and documented trust/closure contract. Public CLI relocation, exact `7`/`42` outputs after input deletion, declaration failures, identity mutations, stale receipts, and zero/one application launches are exercised.

## Task Commits

Each task was committed atomically with normal hooks:

1. **22-02-T1: Build with an explicit local C/ABI manifest** — `3fb87b5` (feat).
2. **22-02-T2: Prove relocation and build/evidence identity invalidation** — `b962491` (test, fixture/docs and probe correction).

## Files Created/Modified

- `internal/compiler/native/bindings.go` — Manifest loading/resolution, input snapshots, ABI probes, dependency confinement and identity construction.
- `internal/compiler/native/bindings_test.go` — Manifest negatives, public CLI relocation, input/receipt mutation, snapshot consistency, launch counts and bounded timing samples.
- `internal/compiler/native/native_app.go` — Optional manifest compilation/linkage, compiler fingerprint and receipt identity validation.
- `internal/compiler/session/session.go`, `cmd/lang/main.go` — Optional manifest passed through the existing checked application path.
- `examples/phase22/identity.bindings.json`, `support.h`, `support.c` — Explicit local C/ABI/runtime inputs; C is linked but uncalled by the Lang source.
- `examples/phase22/BINDINGS.md` — Public commands, bounds, path grammar, trusted-C boundary and macOS/Linux dependency-closure definition.

## Verification

Freshly executed on Darwin arm64 with Apple Clang 21.0.0, target
`arm64-apple-darwin25.6.0`:

- Task 1 command `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase22Bindings' -count=1 -v` passed with named tests, including missing/out-of-root paths, symlinks, undeclared/system-marked local headers, incompatible types and unresolved symbols.
- Final plan command `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase22(Bindings|BuildIdentity|Relocated)' -count=1 -v` passed all 11 matching top-level tests in 7.897s. No zero-test success was counted.
- Full changed-boundary suites passed: native 17.343s and CLI 0.375s in the final rerun; session 98.707s and cgen 3.503s in the broader run. The first broader run caught an unbounded-process test helper; native was rerun in full after correction. A prior narrow session invocation matched zero tests and is not counted as evidence.
- `go vet ./internal/compiler/native ./internal/compiler/session ./cmd/lang` passed with the sandbox-writable Go cache.
- `TestSourceNeverSpawnsUnboundedProcesses` passed after the test helper correction; `git diff --check` passed.
- Actual builds mutate Lang/C/header bytes and valid symbol/typedef declarations. The same production identity function has unit mutation controls for fixed flags, link argv, compiler fingerprint/version, target, host ABI, emitted C and runtime declarations; unsupported runtime declarations also fail real manifest resolution.
- Source inspection confirms the existing foreign-call/contract and by-pointer refusals remain in `cgen_program.go`. The cgen suite passed; this plan does not admit foreign Lang calls.

Linux was not executed here. Existing host CI remains the separate Linux lane;
this summary makes no Linux runtime or complete-provenance claim.

## Decisions Made

- Extended the actual 22-01 session entry in `session.go` while preserving existing callers via the optional manifest argument.
- Bound probes to each symbol's own header and required a function typedef, rejecting object and function-pointer typedefs through C17 function-designator conversion.
- Used separate `input_id` and `build_id`: relocation preserves canonical known inputs; executable bytes always participate in the latter. Output/compiler paths are recorded outside the hash fields.
- Kept legacy pure-source v1 receipts runnable. New identity-bearing receipts validate their internal hashes and artifact bytes; legacy receipts cannot supply the new provenance claim.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Planned session file did not exist**
- **Found during:** Task 1 read-first gate.
- **Issue:** 22-01 implemented the application session entry in `session.go`, not the planned `session_app.go`.
- **Fix:** Extended the existing entry rather than duplicating it.
- **Files modified:** `internal/compiler/session/session.go`.
- **Verification:** CLI/session suites and relocated public CLI build passed.
- **Committed in:** `3fb87b5`.

**2. [Rule 1 - Bug] A sibling header could mask a wrong symbol-header declaration**
- **Found during:** Task 2 declaration-control review.
- **Issue:** One probe including all headers could obtain the symbol declaration from a different header than its manifest entry named.
- **Fix:** Generated and compiled one isolated probe per symbol; added the wrong-declared-header negative control.
- **Files modified:** `internal/compiler/native/bindings.go`, `bindings_test.go`.
- **Verification:** `TestPhase22BindingsRejectInvalidInputs/wrong_declared_header` and the complete native suite passed.
- **Committed in:** `b962491`.

**3. [Rule 2 - Missing Critical] Bound test subprocesses and diagnostics**
- **Found during:** Task 2 broader package verification.
- **Issue:** The repository process guard rejected new CLI test invocations without context deadlines or independent bounded writers.
- **Fix:** Added a deadline-bound test command helper with bounded stdout/stderr.
- **Files modified:** `internal/compiler/native/bindings_test.go`.
- **Verification:** `TestSourceNeverSpawnsUnboundedProcesses` and full native/CLI suites passed.
- **Committed in:** `b962491`.

**Total deviations:** 3 auto-fixed (1 Rule 1, 1 Rule 2, 1 Rule 3).
**Impact:** Necessary integration/correctness corrections within the planned build boundary; no additional language semantics or dependencies.

## Issues Encountered

Relocated builds on this host produced different executable bytes with identical
canonical input IDs. The receipt honestly assigns different build IDs to those
bytes and retains incomplete/non-cacheable closure. No reproducible-binary or
complete host dependency closure claim is made. The retained binaries both
passed the independently specified output checks.

## User Setup Required

None - installed Clang is the existing prerequisite; no new packages or services.

## Next Phase Readiness

Ready for 22-03 evidence integration and Phase 23's explicit local C adapter
linkage. Evidence can reference `BuildReceipt.BuildID` and validate the adjacent
receipt/artifact; `Identity` exposes the known canonical inputs. A successful
link remains trusted build input evidence, not arbitrary C semantic proof.

The read-only requirements gate reports APP-02 and FFI-02 ready. The orchestrator
owns their shared requirement/status updates after integration. STATE.md,
ROADMAP.md, REQUIREMENTS.md and the shared living documents were not modified.
At the phase transition, refresh their now-historical application-boundary
claims using these receipts plus 22-03 evidence; foreign resources and pointers
remain refused until their owning phases.

## Self-Check: PASSED

- All six created code/test/example files exist.
- Task commits `3fb87b5` and `b962491` exist; the ledger measures two task commits.
- All task/plan automated verification commands executed and passed after fixes.
- No new stubs, skipped tests or unrun task verification remain.
- No unplanned trust boundary was introduced beyond the manifest/filesystem,
  trusted C/ABI and artifact identity boundaries in the plan threat model.
- Shared orchestrator planning artifacts are unchanged.

---
*Phase: 22-native-application-build-and-single-execution*
*Completed: 2026-09-27*
