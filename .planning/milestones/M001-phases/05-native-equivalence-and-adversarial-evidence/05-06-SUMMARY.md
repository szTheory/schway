---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 06
subsystem: compiler-native-differential
tags: [native, lto, clang, comparator, execution-document, fail-closed, reflection]

requires:
  - phase: 05-native-equivalence-and-adversarial-evidence
    provides: "05-01's Phase 1-4 byte-identity tripwire; 05-05's testdata/phase5/inline_across_foreign.lang fixture and session.Check/Phase4CheckedProgram/Phase4ThreeEngineDifferential glue"
provides:
  - "native.Runner.LTO: appends -flto to both the per-TU foreign compile and final link command lines, proven non-inert against real codegen"
  - "native.Runner.EnableCommandRecording/LastCommandLines/LastBinary: test-only command-line and binary recording, so -flto reaching clang is an assertion, never an inference from a successful build"
  - "session.Phase5CompareEngines / session.Phase5CompareDiagnosticIDs: a five-axis Phase 5 comparator (terminal-outcome, event-order, resource-ledger, exit-status-signal, diagnostic-id) as a peer of the frozen Phase4CompareThreeEngines"
  - "session.Phase5ComparedComparisonFields / Phase5ExcludedComparisonFields plus a reflection-driven fail-closed field-routing test over execution.Execution"
affects: [05-07, 05-09]

actuals:
  tokens: 10400
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Pointer-shared recorder field (commandRecorder) on a value-receiver Runner, so EnableCommandRecording's copy and the original share one mutable sink across value-copy method calls"
    - "Reflection-driven fail-closed field-routing table (compared+excluded union must exactly equal every reachable struct field), demonstrated red on a local mirror struct rather than a real schema change"
    - "New peer comparator file beside session.go, session.go itself untouched -- same peer-not-fork discipline this repository already applies to phase gate scripts"

key-files:
  created:
    - internal/compiler/native/native_lto_test.go
    - internal/compiler/session/session_phase5_compare.go
    - internal/compiler/session/session_phase5_compare_test.go
  modified:
    - internal/compiler/native/native.go
    - internal/compiler/execution/execution.go

key-decisions:
  - "LTO tests live in a new native_lto_test.go under `package native_test` (external), not the existing internal `package native` native_test.go the plan named -- mirrors this package's own existing native_conformance_test.go precedent for importing session/cgen (session imports native, so only the external test package variant avoids the internal-test cyclic-import shape)."
  - "TestLTOTierIsNotInert compares real compiled-binary bytes via a new Runner.LastBinary() accessor: Runner.Run deletes its temp directory before returning, so the only way to assert 'LTO changed codegen' rather than 'LTO merely built' is to capture the binary's bytes during the run itself, inside the same recorder EnableCommandRecording already introduces for command-line capture."
  - "axis:exit-status-signal is backed by two NEW, independently-set execution.Execution fields (ExitSignaled, ExitSignal) rather than a value derived from Outcome.Kind: a derived signature can never diverge while Outcome.Kind agrees, which would make 'name the diverging axis' fold this axis into axis:terminal-outcome and defeat the acceptance criterion that all five axes are independently seedable."
  - "Phase5ExcludedComparisonFields' two entries (AllocatorAddress, WallClockNanos) are new, execution.Execution fields never populated by interp/cgen/native production code: the pre-Phase-5 struct carries no address/timestamp/allocator/PRNG/thread-identity data at all (this project's place/type/event IDs are semantic ordinals, never addresses), so the exclusion list needed real struct-field targets before TestExcludedFieldsAreNeverRead could exercise anything -- an empty excluded list would make the fail-closed routing machinery present but untested on its excluded branch."
  - "axis:diagnostic-id is asserted via a separate function, Phase5CompareDiagnosticIDs(fixture string, diagnostics map[string]diagnostic.Diagnostic), rather than folded into Phase5CompareEngines' execution.Execution map: reject-programs never execute and therefore never produce an execution.Execution document, so the plan's own stated distinction ('under its own separate control ID... since reject-programs never execute') is honored by a distinct function signature, not merely a distinct axis label inside one shared comparator."
  - "TestRejectProgramDiagnosticIDsAgree exercises equivalence across two independent session.Check invocations on identical source bytes (not check-vs-corevalidate): corevalidate.Result's Problem type has no diagnostic ID field, so the only genuinely independent diagnostic-ID derivation available in this codebase today is repeated invocation of the deterministic, source-position-keyed check.Program path."
  - "Tasks 2 and 3 land as one commit (session_phase5_compare.go + its test), matching 05-03-SUMMARY.md's established precedent for tightly-coupled small files sharing one commit rather than three incremental ones."

patterns-established:
  - "A Runner field gated on itself (LTO bool) that appends a flag to an existing, otherwise-untouched argument-construction path -- copy for any future opt-in build-tier flag on native.Runner."
  - "assertStaticPrefix-by-path-filtering: filter os.PathSeparator-bearing arguments out of a recorded command line before comparing to a literal expected shape, since temp-directory paths are unavoidably non-deterministic across Run calls."

requirements-completed: [NAT-02]

coverage:
  - id: D1
    description: "-flto reaches both the per-TU foreign compile and the final link command lines when Runner.LTO is true, and the non-LTO path is byte-unchanged"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_lto_test.go#TestLTOFlagReachesCompileAndLink"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/native_lto_test.go#TestLTODisabledLeavesCommandLineUnchanged"
        status: pass
    human_judgment: false
  - id: D2
    description: "The LTO tier is proven non-inert: the -O3 LTO and non-LTO binaries for a cross-TU-inlining fixture genuinely differ in codegen"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/native/native_lto_test.go#TestLTOTierIsNotInert"
        status: pass
    human_judgment: false
  - id: D3
    description: "Five named comparison axes (terminal-outcome incl. ok payload/err-alternative, event-order, resource-ledger incl. released-count, exit-status-signal, diagnostic-id) are each independently seedable and correctly named on disagreement, and the frozen Phase4CompareThreeEngines is unregressed"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestPhase5ComparatorNamesTheDivergingAxis"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestPhase5ComparatorComparesOkPayloadAndErrAlternative"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestRejectProgramDiagnosticIDsAgree"
        status: pass
      - kind: integration
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestPhase4ComparatorUnchanged"
        status: pass
    human_judgment: false
  - id: D4
    description: "Every field reachable from execution.Execution is explicitly routed to compared or excluded, and an unrouted field is demonstrated to fail the build without a real schema change"
    requirement: "NAT-02"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestComparisonFieldRoutingIsExhaustive"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestUnroutedFieldFailsTheRoutingTest"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestExcludedFieldsAreNeverRead"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_compare_test.go#TestComparedAndExcludedFieldsHaveNoDuplicate"
        status: pass
    human_judgment: false

duration: ~75min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 6: Native Equivalence and Adversarial Evidence -- LTO Tier and Five-Axis Comparator Summary

**An `-flto`-proven-non-inert `-O3` build tier on `native.Runner`, plus a peer five-axis `session.Phase5CompareEngines`/`Phase5CompareDiagnosticIDs` comparator with a reflection-driven fail-closed field-routing test over `execution.Execution`.**

## Performance

- **Duration:** ~75 min
- **Completed:** 2026-09-06
- **Tasks:** 3 completed
- **Files created:** 3
- **Files modified:** 2

## Accomplishments

- Added `Runner.LTO` (native.go): when true, `-flto` is appended to BOTH the per-TU foreign compile argument list and the final link argument list, gated on the field alone so every existing `-O0`/`-O3` invocation with `LTO` false stays byte-for-byte unchanged (verified via `nonPathTokens` filtering out the unavoidably-nondeterministic temp-directory paths and comparing the remaining literal flag sequence). Added `EnableCommandRecording`/`LastCommandLines`/`LastBinary` -- a pointer-shared `commandRecorder` that survives Runner's value-copy method calls -- so a test can assert `-flto` reached clang directly, and `TestLTOTierIsNotInert` proves the `-O3` LTO and non-LTO binaries for `testdata/phase5/inline_across_foreign.lang` (05-05's own cross-TU-inlining fixture) genuinely differ in codegen bytes, not merely both compile successfully.
- Built `session.Phase5CompareEngines`/`Phase5CompareDiagnosticIDs` (`session_phase5_compare.go`, a new peer file -- `session.go` and the frozen `Phase4CompareThreeEngines` are untouched, confirmed by `git diff --stat` and `TestPhase4ComparatorUnchanged`). Five named axes, each an explicit branch: `axis:terminal-outcome` (closed value/typed_failure/defect set, including the ok payload and err-edge ADT alternative), `axis:event-order` (ordered event sequence with causal-role fields), `axis:resource-ledger` (live-resource set plus released-count), `axis:exit-status-signal` (D-05-20's addition for SC1's silent omission, backed by new independently-settable `ExitSignaled`/`ExitSignal` fields so it can diverge without `Outcome.Kind` also diverging), and `axis:diagnostic-id` (via the sibling `Phase5CompareDiagnosticIDs`, under its own `control:diagnostic.reject_program_id_equivalence`, since reject-programs never execute and therefore never produce an `execution.Execution` document).
- Implemented D-05-21's fail-closed field-routing contract: `Phase5ComparedComparisonFields`/`Phase5ExcludedComparisonFields` name every leaf field reachable from `execution.Execution` by exact struct-field path, and `TestComparisonFieldRoutingIsExhaustive` walks the real struct by reflection and requires the union to match exactly. `TestUnroutedFieldFailsTheRoutingTest` proves the red path on a local mirror struct (no real schema change needed), and `TestExcludedFieldsAreNeverRead` proves the two excluded fields (`AllocatorAddress`, `WallClockNanos` -- both newly added, unpopulated by any production engine) never produce a disagreement even when varied.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the -O3 -flto build configuration and prove the flag actually reaches both command lines** - `ad22a94` (feat)
2. **Task 2 + Task 3: Build the Phase 5 comparator with five compared axes, peering the Phase 4 comparator; make an unrouted comparison field fail the build** - `4ca3cdc` (feat)

_Tasks 2 and 3 build and test the same new file (`session_phase5_compare.go`) and were committed together, matching 05-03-SUMMARY.md's precedent for tightly-coupled small files._

## Files Created/Modified

- `internal/compiler/native/native.go` - `Runner.LTO`, `commandRecorder`, `EnableCommandRecording`/`LastCommandLines`/`LastBinary`/`recordCommandLine`/`recordBinary`; `-flto` wired into the foreign-compile and link argument construction
- `internal/compiler/native/native_lto_test.go` - `TestLTOFlagReachesCompileAndLink`, `TestLTODisabledLeavesCommandLineUnchanged`, `TestLTOTierIsNotInert` (new file, `package native_test`)
- `internal/compiler/session/session_phase5_compare.go` - `Phase5CompareEngines`, `Phase5CompareDiagnosticIDs`, `Phase5EngineDisagreement`, axis constants, `Phase5ComparedComparisonFields`/`Phase5ExcludedComparisonFields`, `reachableFieldPaths`/`unroutedFields`/`staleFields`
- `internal/compiler/session/session_phase5_compare_test.go` - `TestPhase5ComparatorNamesTheDivergingAxis`, `TestPhase5ComparatorComparesOkPayloadAndErrAlternative`, `TestRejectProgramDiagnosticIDsAgree`, `TestPhase4ComparatorUnchanged`, `TestComparisonFieldRoutingIsExhaustive`, `TestUnroutedFieldFailsTheRoutingTest`, `TestExcludedFieldsAreNeverRead`, `TestComparedAndExcludedFieldsHaveNoDuplicate`
- `internal/compiler/execution/execution.go` - Added `AllocatorAddress`, `WallClockNanos`, `ExitSignaled`, `ExitSignal` fields to `Execution` (all `omitempty`, unpopulated by any production engine)

## Decisions Made

See `key-decisions` in frontmatter: the external `native_test` package placement for the LTO tests, capturing binary bytes via a recorder rather than reusing Run's (deleted) temp directory, backing `axis:exit-status-signal` with independent fields rather than a Kind-derived projection, adding two genuinely-excluded fields to give the exclusion list real targets, splitting diagnostic-ID comparison into its own function/control ID, and the two-independent-check-invocations design for `TestRejectProgramDiagnosticIDsAgree`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing Critical] Added independent exit-status/diagnostic-ID/exclusion data execution.Execution lacked**
- **Found during:** Task 2, while designing `TestPhase5ComparatorNamesTheDivergingAxis`
- **Issue:** The plan's five axes require independently seeding a divergence per axis. `execution.Execution` as it existed (Schema/Outcome/Events/LiveResources) had no data that could diverge for `axis:exit-status-signal` without also diverging `axis:terminal-outcome` (a signal derived purely from `Outcome.Kind` can never disagree while `Outcome.Kind` agrees), and `Phase5ExcludedComparisonFields` had no real field to exclude (the pre-Phase-5 struct carries no address/timestamp/allocator/PRNG/thread-identity-shaped field at all).
- **Fix:** Added four new, `omitempty`, unpopulated-by-production-code fields to `execution.Execution`: `ExitSignaled`/`ExitSignal` (independent exit-status-signal axis data) and `AllocatorAddress`/`WallClockNanos` (genuine excluded-field targets).
- **Files modified:** `internal/compiler/execution/execution.go`, `internal/compiler/session/session_phase5_compare.go`
- **Verification:** `TestPhase5ComparatorNamesTheDivergingAxis`'s exit-status-signal subtest diverges the two new fields while `Outcome.Kind` is held identical, and correctly names `axis:exit-status-signal`, not `axis:terminal-outcome`. `TestExcludedFieldsAreNeverRead` varies both new excluded fields and asserts no disagreement.
- **Committed in:** `4ca3cdc` (Task 2/3 commit)

---

**Total deviations:** 1 auto-fixed (1 missing-critical addition, needed for the plan's own stated "five independently-seedable axes" and "genuine excluded-field target" requirements)
**Impact on plan:** Necessary for the comparator's own acceptance criteria to be honestly testable rather than vacuously true. No scope creep -- all four new fields are `omitempty` and untouched by every production engine (interp/cgen/native), so no existing execution document's JSON shape or byte-identity pin changes.

## Issues Encountered

`TestPhase5CorpusThreeEngineAgreement` (05-05's own test, unmodified by this plan) failed exactly once under `go test -race ./...` with `native.timeout: context deadline exceeded` on `acquire_three_fail_second.lang`'s `-O3` run -- a transient timeout under the CPU-contended, race-instrumented full-suite run (`native.Runner`'s default 5s timeout, unrelated to this plan's LTO/comparator changes). Re-ran `go test -race ./internal/compiler/session/... -run TestPhase5CorpusThreeEngineAgreement` in isolation immediately after: all 112 subtests passed cleanly in 127s. Confirmed out of scope per the deviation rules' scope boundary (pre-existing test, not touched by this plan's files) and not re-investigated further.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- `Runner.LTO` and its non-inertness proof are live; 05-07's own LTO-dependent work (mutation runners attacking the LTO tier, per 05-PATTERNS.md) has a real, tested flag to build on.
- `Phase5CompareEngines`/`Phase5CompareDiagnosticIDs` are exported and ready for 05-09's `scripts/verify-phase5.sh` wiring and for later plans to feed real `O3-LTO` engine executions once 05-07 lands the adversarial mutation runners.
- The field-routing test (`TestComparisonFieldRoutingIsExhaustive`) will fail the build the moment any future plan adds a field to `execution.Execution` without routing it -- exactly the fail-closed guarantee D-05-21 asked for.
- No blockers.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*

## Self-Check: PASSED

All 5 created/modified key files verified present on disk. Both task commit hashes (`ad22a94`, `4ca3cdc`) verified in `git log --oneline`. Plan-level `<verification>` block re-run clean: `sh scripts/assert-go-tests.sh ./internal/compiler/native/... TestLTOFlagReachesCompileAndLink TestLTODisabledLeavesCommandLineUnchanged TestLTOTierIsNotInert` passes; `sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestPhase5ComparatorNamesTheDivergingAxis TestPhase5ComparatorComparesOkPayloadAndErrAlternative TestRejectProgramDiagnosticIDsAgree TestPhase4ComparatorUnchanged TestPhase4CorpusThreeEngineAgreement` passes; `sh scripts/assert-go-tests.sh ./internal/compiler/session/... TestComparisonFieldRoutingIsExhaustive TestUnroutedFieldFailsTheRoutingTest TestExcludedFieldsAreNeverRead` passes; `go test ./...` is fully green; `go test -race ./...` is green after confirming one transient, pre-existing, unrelated timeout flake in `TestPhase5CorpusThreeEngineAgreement` reproduces clean in isolation; `go vet ./...` is clean; `git diff --stat internal/compiler/session/session.go` reports no changes.
