---
status: complete
phase: 18-branch-on-a-computed-value
source: [18-01-SUMMARY.md, 18-02-SUMMARY.md, 18-03-SUMMARY.md, 18-04-SUMMARY.md, 18-05-SUMMARY.md, 18-06-SUMMARY.md, 18-07-SUMMARY.md, 18-08-SUMMARY.md, 18-09-SUMMARY.md]
started: 2026-09-25T16:36:27Z
updated: 2026-09-25T19:32:21Z
---

## Current Test

[testing complete]

## Tests

### 1. Computed terminal-match parser refusal
expected: The parser/check entry point pins the computed terminal-match refusal at its intended diagnostic boundary.
result: pass
source: automated
coverage_id: D1

### 2. Result and payload source witnesses
expected: The independent Result-returning and payload-place source witnesses reach the computed-match parser frontier.
result: pass
source: automated
coverage_id: D2

### 3. Production one-arm loan witness
expected: The production source fixture reaches the pre-match borrow used in exactly one branch arm.
result: pass
source: automated
coverage_id: D3

### 4. Computed source-to-native differential
expected: Production parsing/checking, independent admission, interpreter, emitted native C, and execution agreement all accept the computed terminal match.
result: pass
source: automated
coverage_id: D1

### 5. Computed-match round trip and refusals
expected: Generated round trips preserve the terminal match, while out-of-scope, shadowed, and non-data scrutinees are refused.
result: pass
source: automated
coverage_id: D2

### 6. Independent core admission
expected: Core admission accepts a valid computed prefix and rejects arm-local, forged, or wrong-type scrutinees.
result: pass
source: automated
coverage_id: D1

### 7. Independent origin admission
expected: Origin admission derives payload returns through computed aliases and refuses understated or sibling-arm origins.
result: pass
source: automated
coverage_id: D2

### 8. Result-returning computed call
expected: A computed Result returned by a callee is admitted and returns Accepted through interpreter and native tiers.
result: pass
source: automated
coverage_id: D1

### 9. Destructured payload return
expected: The source-driven computed match returns the destructured Buffer payload as Ok:01020304.
result: pass
source: automated
coverage_id: D1

### 10. Wrong-slot mutation control
expected: The injected wrong-slot write is observed and diverges exactly at axis:terminal-outcome; the unmutated companion agrees.
result: pass
source: automated
coverage_id: D2

### 11. One-arm live loan endpoints
expected: The production loan has a point endpoint in its using arm and an edge endpoint in the unused sibling arm.
result: pass
source: automated
coverage_id: D1

### 12. Loan peer and execution agreement
expected: The accepted loan source passes independent peers and agrees across interpreter and all three native optimization tiers for both alternatives.
result: pass
source: automated
coverage_id: D2

### 13. Cold/warm verification evidence
expected: All required verification lanes have internally consistent cold/warm distributions and provenance.
result: pass
source: automated
coverage_id: D1

### 14. CI cost and host-lane disposition
expected: The focused CI disposition is justified by measured cost and existing macOS/Linux full and race lanes remain configured.
result: pass
source: automated
coverage_id: D2

### 15. Fresh phase validation
expected: Phase 18 focused checks, vet, build, serialized full suite, race suite, and both evidence-verifier modes pass.
result: pass
source: automated

### 16. Default-parallel full-suite reliability
expected: `go test ./... -count=1` completes without cache-probe or native-execution timeouts.
result: pass
reported: "Two default-parallel full-suite runs failed in internal/compiler/cache with cache.input_undeclared at probe_test.go:86 (one run also failed line 156) and internal/compiler/cgen with native.timeout: context deadline exceeded at cgen_payload_tracer_test.go:30. The same focused tests pass alone; complete `go test -p=1 ./... -count=1` and `go test -p=4 ./... -count=1` runs pass."
resolution: "Plan 18-09 retained typed Clang probe causes and finite subprocess deadlines, then passed three independent default-parallel full-suite runs, the capped-parallel and race suites, vet, and build. The current evidence verifier validates the captured receipts and Plan 08 CI disposition."
source: automated
verification:
  - "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-postfix-evidence.sh --final"
  - "bash .planning/phases/18-branch-on-a-computed-value/verify-phase18-validation-evidence.sh --final"

## Summary

total: 16
passed: 16
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-18-16
  truth: "The default-parallel full Go suite completes without cache-probe or native-execution timeouts."
  status: resolved
  reason: "Repeated `go test ./... -count=1` runs exceed 5-second subprocess deadlines under package-level parallel load; serial execution and isolated tests pass."
  severity: blocker
  test: 16
  root_cause: "The cgen failure is confirmed to exceed native.DefaultRunner's five-second subprocess timeout under full-suite package load. InputsFor erases the cache Clang-probe error as cache.input_undeclared, so the underlying cache failure is not confirmed; a controlled six-second probe reproduces that public diagnostic. Both complete suites pass when package concurrency is capped at four or one."
  artifacts:
    - path: internal/compiler/native/native.go
      issue: "DefaultRunner applies a five-second timeout to Clang and native execution subprocesses."
    - path: internal/compiler/cache/probe.go
      issue: "InputsFor maps all Clang probe failures to cache.input_undeclared, hiding timeout versus other probe failures."
    - path: internal/compiler/cache/probe_test.go
      issue: "The failing cache acceptance tests only report the mapped code."
    - path: internal/compiler/cgen/cgen_payload_tracer_test.go
      issue: "The tracer acceptance test fails when a native subprocess exceeds the runner deadline."
  missing:
    - "Make full-suite CI reliable under normal package parallelism while preserving bounded, fail-closed subprocess execution and the measured Phase 18 CI disposition."
  debug_session: ".planning/debug/phase18-parallel-suite-timeout.md"
  resolved_by: 18-09-PLAN.md
  resolved_at: 2026-09-25
