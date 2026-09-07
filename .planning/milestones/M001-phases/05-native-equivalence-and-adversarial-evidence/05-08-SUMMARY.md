---
phase: 05-native-equivalence-and-adversarial-evidence
plan: 08
subsystem: testing
tags: [asan, ubsan, ffi, allocator, use-after-free, native, sanitizer]

requires:
  - phase: 05-native-equivalence-and-adversarial-evidence
    provides: "05-03's lane:native-sanitize (RunSanitized, ProbeSanitizerRuntime, classifySanitizerRun, SanitizerReport, pinned ASanOptions/UBSanOptions)"
  - phase: 05-native-equivalence-and-adversarial-evidence
    provides: "05-05's adversarial fixture convention and Phase5MilestoneCorpus union corpus"
provides:
  - "native/lang_foreign_arena.c: a third frozen foreign TU with a real second allocator identity, exposing a genuine ASan-detectable allocator-identity-mismatch defect"
  - "native/lang_foreign_retained.c: a fourth frozen foreign TU, a hostile call-count-dispatched variant subjecting FFI-003's retained-pointer lifetime shape and a genuine heap-use-after-free"
  - "session.VerifyPhase5SanitizeLane: the always-on sanitizer lane dispatch with three real controls plus a named, gate-visible expected escape"
affects: [05-09]

actuals:
  tokens: 11500
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Borrowing a C++ runtime entry point (Itanium-mangled operator new/delete, _Znwm/_ZdlPv) from a plain C17 translation unit via extern declaration plus -lc++, to reach an ASan allocation-type bucket (FROM_NEW) no purely-libc C allocator can produce -- verified empirically, never assumed"
    - "Call-count dispatch on a single declared foreign symbol (lang_foreign_nonlocal.c's own established convention) as the only mechanism for a multi-step, stateful adversarial fixture, since core.Function/cgen carry exactly one ForeignContract per function and only the FIRST declared symbol in a resource-lifecycle sequence is ever actually called in generated C"
    - "Exported test-only function-variable seam (Phase5RetainedPointerFixtureLoaderForTest) for a mutation-kill test that must prove the SUT itself goes red end-to-end, without widening a plan-specified function signature"

key-files:
  created:
    - native/lang_foreign_arena.c
    - native/lang_foreign_retained.c
    - internal/compiler/native/foreign_arena.go
    - internal/compiler/native/foreign_arena_test.go
    - internal/compiler/native/foreign_retained.go
    - internal/compiler/native/foreign_retained_test.go
    - internal/compiler/session/session_phase5_sanitize.go
    - internal/compiler/session/session_phase5_sanitize_test.go
    - testdata/phase5/allocator_mismatch.lang
    - testdata/phase5/retained_pointer.lang
  modified:
    - internal/compiler/native/foreign_nonlocal.go
    - internal/compiler/native/sanitize.go
    - internal/compiler/session/session_phase5_corpus_test.go

key-decisions:
  - "Falsified the plan's literal 'posix_memalign paired with plain free()' allocator-mismatch mechanism empirically before authoring the fixture: on this host (Apple clang 21, arm64), EVERY libc allocation function (malloc, calloc, realloc, posix_memalign, aligned_alloc, memalign, valloc) shares ASan's single FROM_MALLOC allocation-type bucket and is POSIX-compatible with free() -- confirmed with a direct clang+ASan probe before writing any project code, not assumed from the plan's prose. A genuine ASan alloc-dealloc-mismatch report requires the C++ operator-new family (FROM_NEW), reached from plain C via the Itanium-mangled _Znwm/_ZdlPv entry points declared extern and linked against -lc++ (added unconditionally to sanitize.go's compileSanitized link step -- harmless for every other sanitizer-lane build)."
  - "Collapsed the plan's literal three-distinct-symbol design (lang_retained_open/lang_retained_touch/lang_retained_stale) into ONE declared symbol (lang_retained_touch) with call-count dispatch, after dumping the generated C for the three-symbol draft and confirming cgen calls only the FIRST declared symbol for every step in a resource-lifecycle sequence -- core.Function carries exactly one ForeignContract, built from checkResourceLifecycle's first step only. This mirrors lang_foreign_nonlocal.c's own established precedent exactly."
  - "session_phase5_corpus_test.go's generic three-engine differential now explicitly skips both new fixtures (mirroring the existing typed_failure_truncated_stdout.lang skip precedent): D-05-10 requires detection to be ASan-only, never a bare native run, and allocator_mismatch.lang's frozen TU additionally needs -lc++, wired only into the sanitizer lane's own build."
  - "The plan named internal/compiler/native/symbols.go as ForeignSourcePathForSymbol's home file; the switch actually lives in foreign_nonlocal.go. Modified the correct file, documented here."

requirements-completed: [NAT-03]

coverage:
  - id: D1
    description: "A third frozen foreign TU (native/lang_foreign_arena.c) exposes a real second allocator identity and resolves by declared symbol through the existing ForeignSourcePathForSymbol machinery, with Phase 4's static release_allocator_mismatch refusal unchanged"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/native/foreign_arena_test.go#TestForeignArenaSymbolResolves"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/foreign_arena_test.go#TestExistingForeignSymbolsStillResolve"
        status: pass
    human_judgment: false
  - id: D2
    description: "A hostile frozen-TU variant (native/lang_foreign_retained.c) provides a retained-pointer / dangling-read subject reachable through a real Lang program, with the dangling read feeding an observable the optimizer cannot eliminate, and no test asserting a bare native run crashes"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/native/foreign_retained_test.go#TestForeignRetainedSymbolResolves"
        status: pass
      - kind: unit
        ref: "internal/compiler/native/foreign_retained_test.go#TestRetainedPointerFixtureIsUBUnderPlainRun"
        status: pass
    human_judgment: false
  - id: D3
    description: "VerifyPhase5SanitizeLane always runs the retained-pointer positive control, requires alloc-dealloc-mismatch for the allocator-mismatch fixture, and requires runtime error: for a UBSan-triggering fixture, every assertion keyed on SanitizerReportSignatures substring plus nonzero exit code"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestSanitizeLaneRetainedPointerAlwaysReports"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestSanitizeLaneAllocatorMismatchIsDetected"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestSanitizeLaneUBSanNoRecoverIsProven"
        status: pass
    human_judgment: false
  - id: D4
    description: "Sanitizer runtime absence is operational, never a pass, and the lane is proven non-inert: it goes red when the positive control's own fixture stops reporting"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestSanitizeLaneMissingRuntimeIsOperational"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestSanitizeLaneCleanPositiveControlFailsTheLane"
        status: pass
    human_judgment: false
  - id: D5
    description: "escape:callback-invocation-unsubjected is declared under ExpectedEscapes and never appears as a detected control"
    requirement: "NAT-03"
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase5_sanitize_test.go#TestCallbackEscapeIsDeclaredNeverDetected"
        status: pass
    human_judgment: false

duration: 95min
completed: 2026-09-06
status: complete
---

# Phase 5 Plan 8: Dynamic NAT-03 Subjects (Allocator Mismatch, Retained-Pointer UAF) Summary

**Two new frozen foreign TUs give the dynamic half of NAT-03 real subjects -- a genuine ASan-detectable allocator-identity mismatch reached via a direct C-callable C++ operator-new/delete pair, and a call-count-dispatched retained-pointer heap-use-after-free -- both wired into an always-on sanitizer lane whose own positive control is proven to make a clean run meaningful.**

## Performance

- **Duration:** ~95 min
- **Started:** 2026-09-06
- **Completed:** 2026-09-06
- **Tasks:** 3 completed
- **Files created:** 10
- **Files modified:** 3

## Accomplishments

- `native/lang_foreign_arena.c` (third frozen foreign TU, D-05-08): defines the correctly-paired `lang_arena_open_impl`/`lang_arena_close_impl` (posix_memalign + matching free -- a genuinely different physical allocator from `lang_foreign_resource.c`'s plain malloc, kept unused/documentary) and the exposed `_LANG_lang_arena_open`, which deliberately routes its release through the Itanium-mangled `operator delete` entry point (`_ZdlPv`) instead of the matching `free()`. Falsified the plan's literal "posix_memalign + free" mismatch mechanism empirically first: on this host every libc allocator shares ASan's FROM_MALLOC bucket and is `free()`-compatible by design; a genuine `alloc-dealloc-mismatch` report requires the C++ operator-new family, reached from plain C via `_Znwm`/`_ZdlPv` declared `extern` and linked with `-lc++` (added unconditionally to `sanitize.go`'s link step). `ForeignArenaSourcePath` joins `ForeignSourcePathForSymbol`'s resolution switch (which actually lives in `foreign_nonlocal.go`, a documented plan correction).
- `native/lang_foreign_retained.c` (fourth frozen foreign TU, D-05-06/D-05-09): ONE declared symbol (`_LANG_lang_retained_touch`) with call-count dispatch -- the exact convention `lang_foreign_nonlocal.c` already establishes -- because dumping the generated C for an earlier three-distinct-symbol draft proved `cgen` calls only the FIRST declared symbol for every step of a resource-lifecycle sequence (`core.Function` carries exactly one `ForeignContract`). Call 1 genuinely allocates and stashes a pointer in `static` storage (FFI-003, no calls-into-Lang); call 2 reads it live; call 3 frees the same pointer and immediately reads through it again, writing the dangling value into its own returned struct field -- unreachable to dead-store elimination since the reader (generated C) is a separate translation unit.
- `session.VerifyPhase5SanitizeLane` (`internal/compiler/session/session_phase5_sanitize.go`): probes sanitizer runtime availability (operational, never pass, on absence); ALWAYS runs the retained-pointer positive control and requires a genuine `heap-use-after-free` report; runs the allocator-mismatch fixture requiring `alloc-dealloc-mismatch`; runs an embedded UBSan-triggering fixture (signed integer overflow) requiring `runtime error:` with `-fno-sanitize-recover=all` proven effective. Every check keys on `SanitizerReportSignatures`' own substring plus an explicit nonzero-exit-code check. `escape:callback-invocation-unsubjected` is declared in `ExpectedEscapes` and asserted to never appear as a detected control.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add the second frozen foreign TU with a real second allocator identity** - `e7b0968` (feat)
2. **Task 2: Add the hostile retained-pointer / use-after-free frozen-TU variant** - `ca35b26` (feat)
3. **Task 3: Wire the three sanitizer controls with an always-on positive control and the named callback escape** - `b9811f4` (feat)

**Plan metadata:** committed separately below.

## Files Created/Modified

- `native/lang_foreign_arena.c` - Third frozen foreign TU: posix_memalign-based correct pairing plus the deliberately-defective exposed wrapper
- `native/lang_foreign_retained.c` - Fourth frozen foreign TU: call-count-dispatched retained-pointer/UAF fixture
- `internal/compiler/native/foreign_arena.go` - `ForeignArenaSourcePath`
- `internal/compiler/native/foreign_arena_test.go` - `TestForeignArenaSymbolResolves`, `TestExistingForeignSymbolsStillResolve`
- `internal/compiler/native/foreign_retained.go` - `ForeignRetainedSourcePath`
- `internal/compiler/native/foreign_retained_test.go` - `TestForeignRetainedSymbolResolves`, `TestRetainedPointerFixtureIsUBUnderPlainRun`
- `internal/compiler/native/foreign_nonlocal.go` - `ForeignSourcePathForSymbol` gains `lang_arena_open`/`lang_retained_touch` cases
- `internal/compiler/native/sanitize.go` - `compileSanitized`'s link step gains an unconditional `-lc++`
- `internal/compiler/session/session_phase5_sanitize.go` - `ControlSanitizeRetainedPointer`, `ControlSanitizeUseAfterFree`, `ControlSanitizeAllocatorMismatch`, `ControlSanitizeUBSanNoRecover`, `LaneNativeSanitize`, `EscapeCallbackInvocationUnsubjected`, `VerifyPhase5SanitizeLane`, `Phase5SanitizeResult`, `Phase5RetainedPointerFixtureLoaderForTest`
- `internal/compiler/session/session_phase5_sanitize_test.go` - `TestSanitizeLaneRetainedPointerAlwaysReports`, `TestSanitizeLaneAllocatorMismatchIsDetected`, `TestSanitizeLaneUBSanNoRecoverIsProven`, `TestSanitizeLaneMissingRuntimeIsOperational`, `TestSanitizeLaneCleanPositiveControlFailsTheLane`, `TestCallbackEscapeIsDeclaredNeverDetected`
- `internal/compiler/session/session_phase5_corpus_test.go` - `TestPhase5CorpusThreeEngineAgreement` skips the two new ASan-only fixtures
- `testdata/phase5/allocator_mismatch.lang` - Dynamic allocator-mismatch fixture
- `testdata/phase5/retained_pointer.lang` - Retained-pointer/use-after-free fixture

## Decisions Made

See `key-decisions` in frontmatter: the empirically-falsified allocator-mismatch mechanism (posix_memalign+free produces no report; operator-new/delete does, requiring `-lc++`), the collapse from three distinct foreign symbols to one call-count-dispatched symbol (cgen only calls the first declared symbol per resource-lifecycle sequence), the corpus-differential skip for both new ASan-only fixtures, and the `symbols.go` → `foreign_nonlocal.go` file correction.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Plan's literal allocator-mismatch mechanism does not exist on this host**
- **Found during:** Task 1, before writing any C source
- **Issue:** The plan's action text specified "routes its release through the plain-malloc wrapper's free" as the allocator-mismatch mechanism (posix_memalign allocate, plain `free()` release). Verified empirically with a direct clang+ASan probe (`ASAN_OPTIONS=alloc_dealloc_mismatch=1`) before writing project code: this produces exit 0, no report, on Apple clang 21/arm64, because every libc allocation function collapses into ASan's single FROM_MALLOC bucket and is `free()`-compatible by design.
- **Fix:** Used the Itanium-mangled `operator new`/`operator delete` entry points (`_Znwm`/`_ZdlPv`), callable directly from plain C via `extern` declaration and linked against `-lc++`, verified empirically to produce `ERROR: AddressSanitizer: alloc-dealloc-mismatch (malloc vs operator delete)` with a nonzero exit code.
- **Files modified:** `native/lang_foreign_arena.c`, `internal/compiler/native/sanitize.go`
- **Verification:** `TestSanitizeLaneAllocatorMismatchIsDetected` passes; manual probe transcript confirms the negative result for the plan's original mechanism.
- **Committed in:** `e7b0968` (Task 1 commit)

**2. [Rule 1 - Bug] Three-distinct-symbol resource-lifecycle sequence collapses to one symbol in generated C**
- **Found during:** Task 2, while authoring `testdata/phase5/retained_pointer.lang` with three separately-declared foreign symbols
- **Issue:** Dumping the generated C for the three-symbol draft showed `cgen` calling `_LANG_lang_retained_open` three times regardless of which of the three declared symbols each step named -- `core.Function` carries exactly one `ForeignContract`, built from `checkResourceLifecycle`'s FIRST step only, so the second and third symbols were never actually reachable from generated code.
- **Fix:** Collapsed to ONE declared symbol (`lang_retained_touch`) with `static` call-count dispatch, mirroring `lang_foreign_nonlocal.c`'s own established "second call behaves differently" convention.
- **Files modified:** `native/lang_foreign_retained.c`, `testdata/phase5/retained_pointer.lang`, `internal/compiler/native/foreign_nonlocal.go`
- **Verification:** `TestSanitizeLaneRetainedPointerAlwaysReports` and `TestForeignRetainedSymbolResolves` pass; the compiled generated C was dumped and inspected to confirm all three calls now dispatch through the single symbol.
- **Committed in:** `ca35b26` (Task 2 commit)

**3. [Rule 3 - Blocking] `testdata/phase5/allocator_mismatch.lang` breaks the existing Phase 5 milestone-corpus three-engine differential**
- **Found during:** Task 3's full-suite verification (`go test ./...`)
- **Issue:** Plan 05-05's `TestPhase5CorpusThreeEngineAgreement` globs every `testdata/phase5/*.lang` fixture and runs it through the PLAIN (non-sanitizer) native differential. `allocator_mismatch.lang`'s frozen TU requires `-lc++` (wired only into the sanitizer lane's build), so the plain differential's link step failed.
- **Fix:** Added an explicit skip for both `allocator_mismatch.lang` and `retained_pointer.lang` in that test, mirroring the existing `typed_failure_truncated_stdout.lang` skip precedent, documenting D-05-10's own requirement that detection for both fixtures is ASan-only, never a bare native run.
- **Files modified:** `internal/compiler/session/session_phase5_corpus_test.go`
- **Verification:** `go test ./...` and `go test -race ./internal/compiler/native/... ./internal/compiler/session/...` both green; `go vet ./...` clean.
- **Committed in:** `b9811f4` (Task 3 commit)

**4. [Rule 1 - Bug] Plan named the wrong file for `ForeignSourcePathForSymbol`**
- **Found during:** Task 1
- **Issue:** The plan's `<files>` list named `internal/compiler/native/symbols.go` as the home of `ForeignSourcePathForSymbol`'s switch; it actually lives in `foreign_nonlocal.go`.
- **Fix:** Modified the correct file.
- **Files modified:** `internal/compiler/native/foreign_nonlocal.go`
- **Verification:** `grep -n "func ForeignSourcePathForSymbol"` confirms the single definition site.
- **Committed in:** `e7b0968`, `ca35b26`

---

**Total deviations:** 4 auto-fixed (2 bugs discovered via empirical falsification before/during authoring, 1 blocking cross-plan test regression, 1 minor file-location correction).
**Impact on plan:** All four were necessary for the dynamic subjects to be genuine, reachable, and non-regressive. No scope creep -- the underlying design intent (a real second allocator identity; a real retained-pointer UAF; both ASan-detected) is preserved exactly, only the concrete mechanism differs from the plan's literal prose where that prose did not hold on this host/compiler.

## Issues Encountered

None beyond the four deviations above, all resolved before each task's own commit.

## User Setup Required

None - no external service configuration required. The sanitizer toolchain (clang with ASan/UBSan) and the C++ runtime (`libc++`, linked via `-lc++`) were already present on this host.

## Next Phase Readiness

- `lane:native-sanitize` now has real, reachable dynamic subjects for allocator mismatch and retained-pointer use-after-free, both proven via `session.VerifyPhase5SanitizeLane`'s own tests, ready for plan 05-09 to fold into `Phase5RequiredControls()`/`scripts/verify-phase5.sh`.
- `session_phase5_alias.go`'s `NAT03Mutations()` rows 6 and 7 (added by plan 05-07, run concurrently in the same wave) cite `testdata/phase5/allocator_mismatch.lang` and `testdata/phase5/retained_pointer.lang` at their final declared paths -- both now exist with the exact `adversarial-target` headers those rows expect.
- Full repo test suite (`go build ./...`, `go vet ./...`, `go test ./...`, `go test -race ./internal/compiler/native/... ./internal/compiler/session/...`) is green.
- No blockers.

---
*Phase: 05-native-equivalence-and-adversarial-evidence*
*Completed: 2026-09-06*

## Self-Check: PASSED

- FOUND: native/lang_foreign_arena.c
- FOUND: native/lang_foreign_retained.c
- FOUND: internal/compiler/native/foreign_arena.go
- FOUND: internal/compiler/native/foreign_arena_test.go
- FOUND: internal/compiler/native/foreign_retained.go
- FOUND: internal/compiler/native/foreign_retained_test.go
- FOUND: internal/compiler/session/session_phase5_sanitize.go
- FOUND: internal/compiler/session/session_phase5_sanitize_test.go
- FOUND: testdata/phase5/allocator_mismatch.lang
- FOUND: testdata/phase5/retained_pointer.lang
- FOUND: commit e7b0968 (Task 1)
- FOUND: commit ca35b26 (Task 2)
- FOUND: commit b9811f4 (Task 3)
- Plan-level `<verification>` re-run clean: named test list passes via `scripts/assert-go-tests.sh`, `go build ./...`/`go vet ./...`/`go test ./...` clean, `go test -race ./internal/compiler/native/... ./internal/compiler/session/...` clean, `grep -rn 'MallocScribble\|MALLOC_PERTURB_' internal/ native/ testdata/` returns no matches.
