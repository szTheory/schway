---
phase: 20-nyquist-d-13-34-and-the-frontier-fixture
plan: "04"
subsystem: compiler
tags: [native, cache, content-addressing, timing]
requires:
  - phase: 05
    provides: Three-engine corpus and native execution runner
  - phase: 20-01
    provides: Current Phase 20 source fixture and validation surface
provides:
  - Content-addressed native executable reuse with fail-closed dependency discovery
  - Fresh execution and verdict checks on every cache hit
  - Three-pair timing harness with deterministic smoke tests
affects: [phase-20-plan-07, phase-20-plan-08, native-cache]
actuals:
  tokens: 15199
  tasks: 3
  commits: 1
tech-stack:
  added: []
  patterns: [dependency-byte manifest, artifact digest validation, injected paired-run harness]
key-files:
  created:
    - internal/compiler/session/session_phase20_cache_test.go
    - scripts/phase20-closure-timing.go
    - scripts/phase20-closure-timing_test.go
  modified:
    - internal/compiler/cache/cache.go
    - internal/compiler/cache/cache_test.go
    - internal/compiler/native/native.go
    - internal/compiler/native/native_test.go
    - internal/compiler/session/session_phase5_corpus_test.go
key-decisions:
  - "Only enable cache reuse when Darwin compiler, SDK, dependency, link, and target inputs are completely discoverable; unknown inputs compile fresh."
  - "Cache executable bytes only; each hit still launches native processes and decodes current execution output."
  - "Keep full-suite timing behind Plan 20-07 reconciliation and Plan 20-08's green preflight."
patterns-established:
  - "Hash ordered library-search candidate presence and bytes so a newly appearing higher-priority library forces a miss."
requirements-completed: []
coverage:
  - id: D1
    description: "The native artifact cache reuses only complete input identities and rejects corrupted artifacts."
    verification:
      - kind: unit
        ref: "go test ./internal/compiler/native ./internal/compiler/cache -run 'TestPhase20|TestCache|TestNative' -count=1"
        status: pass
    human_judgment: false
  - id: D2
    description: "The 112-program closure has cold misses and warm artifact hits while current executions remain fresh."
    verification:
      - kind: integration
        ref: "TestPhase5CorpusThreeEngineAgreement/enumerated-closure, LANG_PHASE5_CLOSURE_CACHE paired cold/warm runs"
        status: pass
    human_judgment: false
  - id: D3
    description: "The timing helper's pair ordering, cache roots, failure handling, and report formatting are smoke-tested without full-suite runs."
    verification:
      - kind: unit
        ref: "go test ./scripts -run '^TestPhase20ClosureTimingHarness$' -count=1"
        status: pass
    human_judgment: false
duration: "not reliably recorded across context continuation"
completed: 2026-09-25
status: complete
---

# Phase 20 Plan 04: content-addressed native closure summary

**The enumerated native closure reuses cached executables only when every discovered build input matches, while rerunning native execution and comparison on each pass.**

## Performance

- **Duration:** not reliably recorded across context continuation
- **Completed:** 2026-09-25
- **Tasks:** 3
- **Files modified:** 8

## Accomplishments

- Added a content-bound build cache for native executables. Its manifest includes emitted and foreign source bytes, transitive headers, compiler bytes/version, target, flags, SDK identity, link inputs, ordered search paths, and link-candidate presence. Unsupported or incomplete discovery bypasses reuse.
- Added artifact digest verification. Cached results never include executions or comparator outcomes; each cache hit restores a private executable and freshly runs and decodes it.
- Integrated cache counters into the complete 112-program closure and added a deterministic timing helper with a 30-minute cap, six full-suite sample limit, paired cache roots, and unit-level fake-run checks.

## Task Commits

- `fd89838` — native cache, closure integration, timing helper, and smoke tests.

## Verification

- Native/cache tests passed: `go test ./internal/compiler/native ./internal/compiler/cache -run 'TestPhase20|TestCache|TestNative' -count=1`.
- Closure integration passed on a paired cold/warm run over all 112 programs: cold `reused=0, recomputed=112`; warm `reused=112, recomputed=0`.
- Timing harness smoke test passed: `go test ./scripts -run '^TestPhase20ClosureTimingHarness$' -count=1`.
- Bounded subprocess and Phase 20 focused session tests passed.
- The full-suite timing protocol has not run. Plan 20-08 owns it after Plan 20-07 reconciles the tree and confirms a green preflight.

## Deviations from Plan

Plan 20-04 was revised to implement and smoke-check the timing harness in Wave 1 while deferring all six full-suite timing runs until Plan 20-08. This prevents the expected cross-plan red gates from contaminating the timing distribution.

## Next Phase Readiness

The cache and harness are ready. Plan 20-07 must refresh the source-derived archive, maturity, emitter, skip-witness, debt, and frontier records before Plan 20-08 runs the full suite and timing protocol.

---
*Phase: 20-nyquist-d-13-34-and-the-frontier-fixture*
*Completed: 2026-09-25*
