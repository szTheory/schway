---
phase: 30-local-modules
plan: 09
subsystem: examples-and-phase-close
tags: [go, modules, adler32, runnable-gain, validation, transition-check]
requires:
  - phase: 30-local-modules
    provides: 30-00 gate and no-import golden; 30-04 modload.Load; 30-05 merged-program check; 30-07 module-qualified C names; 30-08 module-aware session routes
provides:
  - examples/phase30: adler32.schway (module adler32, export fn adler32) and adler32_cli.schway (imports it, owns the foreign block, data FixtureError and main)
  - TestPhase30AdlerModuleAppCLI (public build --manifest + app run, ten pinned answers, mutated-library control)
  - TestPhase30AdlerModuleInterpreter and TestPhase30LibraryIsSingleCopy (session)
  - register updates for the two new .schway files (phase28/272/273, LANGUAGE-MATURITY corpus sentence)
  - 30-VALIDATION.md finalized (validated) and 30-TRANSITION-CHECK.md
affects: [30-verification, phase-31]
actuals:
  tokens: 21000
  tasks: 3
  commits: 3
plan_head_before: 596520ba28d279cbea5f77281c60fbad60c869a4
plan_head_after: 0aa0d4ff6fb0b301886ad85bc301e990468c5633
commits: 3
tech-stack:
  added: []
  patterns:
    - "runnable gain through build --manifest + app run (native) and the session interpreter API (interp.RunWithFixtureBytes), answers from independent Phase 29 literals"
    - "armed control by mutating a TempDir copy of the library (var a = 1 to var a = 2) and requiring the pinned comparison to observe the change"
key-files:
  created:
    - examples/phase30/adler32.schway
    - examples/phase30/adler32_cli.schway
    - examples/phase30/README.md
    - internal/compiler/session/session_phase30_adler_test.go
    - cmd/schway/main_phase30_test.go
    - .planning/phases/30-local-modules/30-TRANSITION-CHECK.md
  modified:
    - internal/compiler/session/session_phase28_manifest_test.go
    - internal/compiler/session/session_phase272_identity_test.go
    - internal/compiler/session/session_phase273_identity_test.go
    - .planning/LANGUAGE-MATURITY.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/phases/30-local-modules/30-VALIDATION.md
key-decisions:
  - "The two .schway files are committed verbatim from the Phase 29 fixture, not run through `schway format`: the formatter has a pre-existing defect that rewrites the trailing `r` after `r = r + a` into `r = r + ar`, so formatting would break the library."
  - "Phase 28 manifest row for the importing entry records EntryShape unchecked because the byte-taking walker refuses imports; the module-aware route is pinned by the two Phase30 adler tests."
  - "TestPhase30LibraryIsSingleCopy lives in the session adler test file created in Task 1, so it was committed with Task 1 rather than Task 2."
requirements-completed: [MOD-03, MOD-04, MOD-05]
duration: 38min
completed: 2026-10-12
status: complete
coverage:
  - id: D1
    description: "The Adler-32 library is imported from its own module file and prints the ten pinned Phase 29 answers natively through build --manifest and app run, with unknown case exit 65"
    requirement: MOD-03
    verification:
      - kind: integration
        ref: "cmd/schway/main_phase30_test.go#TestPhase30AdlerModuleAppCLI"
        status: pass
    human_judgment: false
  - id: D2
    description: "The interpreter returns the same ten pinned answers for the imported library; the merged adler32 function belongs to module adler32"
    requirement: MOD-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_adler_test.go#TestPhase30AdlerModuleInterpreter"
        status: pass
    human_judgment: false
  - id: D3
    description: "The library exists as exactly one copy, and a copied library is detected"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_phase30_adler_test.go#TestPhase30LibraryIsSingleCopy"
        status: pass
    human_judgment: false
  - id: D4
    description: "Corpus registers, validation contract and transition record are current and grounded"
    verification:
      - kind: unit
        ref: "internal/compiler/session#Inventory|Groundedness|ValidationLifecycle meta set plus TestPhase28EntryExitClassManifest, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity, TestLanguageMaturityCountsAreCurrent"
        status: pass
    human_judgment: false
---

# Phase 30 Plan 09: Imported Adler-32 Library Summary

**The Phase 29 Adler-32 program now runs with its library in `examples/phase30/adler32.schway`, imported by the entry, printing the ten pinned answers natively (build --manifest + app run) and in the interpreter.**

## Performance

- **Duration:** about 38 min (includes a 3 min session package run and a 3 min native test)
- **Tasks:** 3
- **Files modified:** 12 (6 created, 6 modified)

## Accomplishments

- `adler32.schway` (`module adler32`, `export { fn adler32 }`, pure scalar) and `adler32_cli.schway` (imports it, keeps the foreign owner, `data FixtureError` and `main`); one definition of `fn adler32(` under examples/phase30.
- Native evidence, quoted from the focused run: `--- PASS: TestPhase30AdlerModuleAppCLI (1.21s)` and `--- PASS: TestPhase30AdlerModuleAppCLI/mutated-library (0.58s)`, with clang present and no SKIP line. The interpreter gives case 3 = 300286872 and all ten pinned values; the mutated library differs for all ten.
- Registers moved only for the two new files (phase28 rows, phase272/273 allow-lists for the library, corpus sentence 161 programs / 5,282 lines). `examples/phase29` and `testdata/phase29` are byte-unchanged and `TestPhase30PreChangeGolden` is green.
- 30-VALIDATION.md is `status: validated`, `nyquist_compliant: true`, `wave_0_complete: true`, its quick-run row names the real Phase30 packages, and every test it names exists; the meta set (`Inventory|Groundedness|ValidationLifecycle` plus the register tests) passes.
- 30-TRANSITION-CHECK.md records claims with evidence classes, the seed table (SEED-261008-z1g retained dormant with its witness and consumer), three next capabilities (Phase 31 first), runnable-gain status and the hosted-CI-at-gate note (EVD-15 stays in Phase 32). LANGUAGE-MATURITY and PRODUCT-ROADMAP carry dated amendments that state whole-program composition and the non-claims.

## Task Commits

1. **Task 1: library module, importing entry, interpreter answers, registers** - `c283344b` (feat)
2. **Task 2: native answers through the public route, README** - `535806a8` (test)
3. **Task 3: finalize validation and transition read** - `0aa0d4ff` (docs)

**Plan metadata:** the docs commit following this SUMMARY.

## Files Created/Modified

- `examples/phase30/adler32.schway`, `adler32_cli.schway`, `README.md` - the module pair and its documentation with non-claims.
- `internal/compiler/session/session_phase30_adler_test.go` - interpreter answers, single-copy guard, mutated-library and copied-library controls.
- `cmd/schway/main_phase30_test.go` - public CLI route, check, run-engine classes, mutated-library control.
- `session_phase28_manifest_test.go`, `session_phase272_identity_test.go`, `session_phase273_identity_test.go` - D-27 register rows.
- `.planning/LANGUAGE-MATURITY.md`, `.planning/PRODUCT-ROADMAP.md`, `30-VALIDATION.md`, `30-TRANSITION-CHECK.md` - living documents.

## Decisions Made

- The `.schway` files are verbatim copies of the Phase 29 text, not formatter output (see Deviations).
- `run --engine=interpreter|native` on the new entry gives exit 3 with `tool.run_failed` / `native.tool_failure`, identical to the Phase 29 fixture; the CLI test asserts both programs in the same subtest.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Did not commit `schway format` output**
- **Found during:** Task 1
- **Issue:** `schway format` is not idempotent-safe on this body: it rewrites the trailing expression `r` after `r = r + a` into `r = r + ar`, producing an invalid library (`syntax.expected_linear_result`). The Phase 29 fixture is not format-canonical either (`format.non_canonical`), a defect already carried since 29-02.
- **Fix:** committed the verbatim Phase 29 text (module header, import and export blocks added by hand). Recorded as a known gap in 30-TRANSITION-CHECK.md; not fixed (out of scope).
- **Files modified:** examples/phase30/adler32.schway, examples/phase30/adler32_cli.schway
- **Verification:** `schway check` on the entry passes; both engines give the pinned values.
- **Committed in:** c283344b

**2. Task boundary: TestPhase30LibraryIsSingleCopy committed with Task 1**
- It shares the session adler test file created in Task 1, so it landed in `c283344b`; Task 2 added the CLI test and README.

---

**Total deviations:** 1 auto-fixed (Rule 3), 1 bookkeeping note
**Impact on plan:** none on behavior; both files remain byte-faithful to the Phase 29 program.

## Issues Encountered

None beyond the formatter defect above. Requirements MOD-03/04/05 are listed in this plan's frontmatter; whether they are marked complete is decided by `requirements.ready-ids` and the orchestrator's phase completion, not here.

## Known Stubs

None.

## Threat Flags

None.

## Verification run

- `go test ./cmd/schway -count=1`: ok. `go test ./internal/compiler/session -count=1`: ok (195 s).
- Phase30 set across modload, syntax, diagnostic, protocol, debugmap, check, core, corevalidate, callgraph, interp, cgen, session, cmd/schway, cmd/schway-repair: all ok.
- Hosted full and race CI has not run; it runs at the phase gate and its run IDs go in 30-VERIFICATION.md.

## Next Phase Readiness

Phase 30 plans are complete; ready for phase verification. Phase 31 (second consumer, file adapter, MOD-06 module hashing) is the next capability.

## Self-Check: PASSED
