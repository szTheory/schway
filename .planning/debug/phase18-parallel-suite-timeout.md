---
status: resolved
trigger: "Phase 18 UAT G-18-16: default-parallel full Go suite fails cache probe and native execution with 5-second deadlines; isolated and serialized runs pass"
created: 2026-09-25
updated: 2026-09-25
---

## Current Focus

hypothesis: Five-second subprocess deadlines were insufficient under package-level load; the original cache refusal's erased inner cause remains unknown.
bug_class: heisenbug-mandelbug
test: Run repeated default-parallel full suites and verify retained typed causes and bounded subprocess falsifiers.
expecting: Full suites pass under default package concurrency; any future Clang probe refusal exposes its underlying typed cause.
next_action: none — Plan 18-09 and its final evidence gate closed G-18-16.
reasoning_checkpoint:
  candidate_causes:
    - "code: fixed five-second subprocess timeout and lossy cache probe error mapping"
    - "environment: default Go package parallelism overloads this 18-core host during full suite"
    - "data: malformed fixture or missing temporary cgen files"
  and_gate: "Likely yes: a fixed timeout plus scheduling/resource contention co-occur; the cache error mapping obscures which probe failure happened. Confirm before collapsing."

## Symptoms

expected: `go test ./... -count=1` passes under default package parallelism.
actual: Two default-parallel full-suite runs failed in `internal/compiler/cache` and `internal/compiler/cgen`; isolated tests, serialized full suite, and `-p=4` affected-package run pass.
errors: `cache.input_undeclared` at `probe_test.go:86` (one run also line 156); `native.timeout: context deadline exceeded` after 5.10s at `cgen_payload_tracer_test.go:30`.
reproduction: `GOCACHE=/tmp/ai-lang-gocache go test ./... -count=1` under default package parallelism.
started: Observed during Phase 18 UAT on 2026-09-25; prior history unknown.

## Eliminated

## Evidence

- timestamp: 2026-09-25
  checked: Phase 18 UAT G-18-16
  found: Two full default-parallel runs fail; same focused tests pass alone, `-p=1` full suite passes, and `-p=4` affected packages pass.
  implication: Package-level load is a plausible amplifier, but the exact cause is not yet confirmed.
- timestamp: 2026-09-25
  checked: GSD debug knowledge base
  found: No prior resolution matches these cache probe/native subprocess timeouts.
  implication: Investigate the current paths directly.
- timestamp: 2026-09-25
  checked: `cache/probe_test.go:66-87`, `cache/probe.go:155-212,266-310`
  found: The failing line 86 is a valid synthetic `ArtifactSpec` passed to `InputsFor`; `ProbeClangIdentity` executes the fixture script under a five-second timeout. `InputsFor` maps any probe error, including `cache.probe_timeout`, to `cache.input_undeclared`.
  implication: The cache diagnostic does not reveal whether the subprocess timed out; fixture, path, or probe process failures are also possible.
- timestamp: 2026-09-25
  checked: `cgen/cgen_payload_tracer_test.go:23-30`, `session/session.go:1062-1150`, `native/native.go:150,459-598`
  found: The cgen test passes `native.DefaultRunner()` with a five-second timeout through `RunNativeFile`; native compiles and runs O0 then O3 with separate five-second subprocess deadlines, each returning `native.timeout` when exceeded.
  implication: The observed `context deadline exceeded` is an actual elapsed subprocess deadline, though the error does not identify compile versus run stage.
- timestamp: 2026-09-25
  checked: Three isolated repetitions of both exact failing tests with `go test -p=2 -count=3 -v`
  found: All pass. Cache test duration is 0.48-0.58s; cgen tracer is 0.93-1.52s.
  implication: Invalid fixtures and deterministic logic defects are less plausible; failure depends on broader process conditions.
- timestamp: 2026-09-25
  checked: Controlled cache probe experiment with identical valid `ArtifactSpec`, changing only fixture command from immediate echo to six-second delayed echo
  found: Immediate command passes. Delayed command makes `InputsFor` return `cache.input_undeclared` after 6.297s; direct `ProbeClangIdentity` on the delayed command returns `cache.probe_timeout` after 6.016s.
  implication: A five-second probe timeout is sufficient to reproduce the exact UAT cache diagnostic through lossy error mapping. It does not prove the original cache run took this path; a transient process-launch failure could produce the same public code.
- timestamp: 2026-09-25
  checked: Diagnosis boundary and narrow mitigation evidence
  found: Native `context deadline exceeded` is directly confirmed at `DefaultRunner`'s five-second bound. Cache's underlying error is erased. Prior `-p=4` success covered affected packages only, while the complete `-p=1` suite passed.
  implication: Controlled package parallelism is a candidate mitigation; a full `-p=4` suite is still required to validate it, and the cache root cause remains unconfirmed until the inner probe code is observed.

## Resolution

root_cause: "The native failure directly exceeded a five-second subprocess deadline under full-suite package load. The original cache failure was mapped to cache.input_undeclared before its inner cause was preserved, so its exact original cause remains unconfirmed."
fix: "Raised the finite default subprocess budget to 30 seconds while retaining caller deadlines and parent cancellation; cache refusals now retain typed inner probe failures."
verification: "Three independent default-parallel full suites, the -p=4 suite, the race suite, vet, build, and both Phase 18 evidence verifiers pass."
files_changed:
  - internal/compiler/cache/cache.go
  - internal/compiler/cache/probe.go
  - internal/compiler/cache/probe_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_test.go
  - internal/compiler/measure/machine.go
  - internal/compiler/measure/machine_test.go
