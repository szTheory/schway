# Phase 01 Pattern Map

## Scope

There is no production compiler code yet. These are experimental analogs to
mine for evidence patterns, not modules to import or extend.

| New production area | Closest analog | Reuse | Do not inherit |
|---------------------|----------------|-------|----------------|
| `internal/compiler/core` | `.planning/spikes/001-ownership-kernel-workbench/ownership/model.go` | explicit immutable facts and enums | ownership-only API and spike naming |
| `internal/compiler/check` | `.planning/spikes/003-public-origins-generic-abilities/origins/checker.go` | deterministic findings and independent expected outcomes | frontend/body assumptions |
| `internal/compiler/evidence` | `.planning/spikes/004-independent-certificate-checker/format/model.go` and `verifier/verifier.go` | strict schema validation, digest binding, mutation matrix | replay-sized default certificates |
| `internal/compiler/native` | `.planning/spikes/005-native-ffi-provenance-cleanup/lab/lab.go` | explicit Clang commands, O0/O3 normalization, temp isolation | hand-authored C as language semantics |
| `cmd/lang` | spike `cmd/*/main.go` programs | standard-library CLI, deterministic report | one-off report-only command surface |

## Required Dependency Direction

```text
cmd/lang -> session
session -> syntax, check, interp, cgen/native, evidence
check -> ast/core
interp -> core
cgen -> core
native -> generated C + tool request
evidence -> exported schema values only
core -/-> syntax
interp -/-> cgen/native
```

`session` is the composition root. Cross-stage communication uses explicit
results; no package-global registry, logger, clock, cache, or mutable compiler
singleton is introduced.

## Test Pattern

Keep flat table/AAA tests close to their package. Keep black-box language
fixtures in `testdata/phase1` with stable IDs and expected outcomes. Generated
goldens must be reviewable text. Seeded property cases run in normal `go test`;
time-bounded fuzzing is opt-in verification evidence.
