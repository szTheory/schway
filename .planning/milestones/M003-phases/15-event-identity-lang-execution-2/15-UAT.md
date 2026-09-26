---
status: complete
phase: 15-event-identity-lang-execution-2
source: 15-01-SUMMARY.md, 15-02-SUMMARY.md, 15-03-SUMMARY.md, 15-04-SUMMARY.md, 15-05-SUMMARY.md, 15-06-SUMMARY.md, 15-07-SUMMARY.md
started: 2026-09-21T23:54:47Z
updated: 2026-09-22T13:00:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Schema 2 execution admission
expected: A multi-function execution document with canonical invocation identities is accepted, while a non-canonical invocation, duplicate activation identity, invalid event kind, or invalid caller-owned callee identity is refused with an actionable diagnostic.
result: pass
source: automated

### 2. Multi-function interpreter evidence
expected: Multi-function interpreter events carry canonical activation identity while single-function bytes remain /1.
result: pass
source: automated
coverage_id: D1

### 3. Interpreter call-edge admission
expected: Admitted calls produce one caller-owned preorder edge; rejected calls do not.
result: pass
source: automated
coverage_id: D2

### 4. Invocation membership peer validation
expected: Independent bounded invocation membership and entry resolution.
result: pass
source: automated
coverage_id: D1

### 5. Causal peer validation
expected: Observed causal structure, ownership, and preorder validation.
result: pass
source: automated
coverage_id: D2

### 6. Native invocation-table size
expected: The checked four-diamond source fixture unfolds to exactly 61 invocation nodes after graph and entry validation.
result: pass
source: automated
coverage_id: D1

### 7. Native invocation-table bound
expected: Native preflight admits 4096 nodes, refuses the attempted 4097th with stable context, and cannot be bypassed inertly.
result: pass
source: automated
coverage_id: D2

### 8. Native activation identity
expected: Static native invocation tables preserve distinct activation occurrences and thread indices through internal calls.
result: pass
source: automated
coverage_id: D1

### 9. Native call evidence
expected: Native multi-function execution writes interpreter-equivalent /2 invocation evidence and preorder call edges.
result: pass
source: automated
coverage_id: D2

### 10. Frozen legacy native bytes
expected: Legacy C event writers retain exact frozen output while schema selection is explicit.
result: pass
source: automated
coverage_id: D3

### 11. Schema 2 peer comparison gate
expected: The program-aware four-engine Schema 2 comparison refuses a corrupted engine document before it can report agreement, while /0 and /1 comparison behavior remains unchanged.
result: pass
source: automated

### 12. Four-tier shared-leaf diamond gate
expected: The unchanged shared-leaf diamond completes successfully across interpreter, O0, O3, and O3-LTO using Schema 2 evidence; the historical duplicate-pair collision control remains live.
result: pass
source: automated

## Summary

total: 12
passed: 12
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-15-1
  truth: "A multi-function execution document with canonical invocation identities is accepted, while a non-canonical invocation, duplicate activation identity, invalid event kind, or invalid caller-owned callee identity is refused with an actionable diagnostic."
  status: resolved
  resolved_by: 15-08-PLAN.md
  resolved_at: 2026-09-22
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 1
  root_cause: "Plan 01 omits machine-readable coverage metadata, causing verify-work to route already-passing CI tests to human UAT; its validator test also bypasses the JSON/ToolError seam and does not pin actionable diagnostic identity."
  artifacts:
    - path: ".planning/phases/15-event-identity-lang-execution-2/15-01-SUMMARY.md"
      issue: "Missing coverage block causes legacy UAT fallback."
    - path: "internal/compiler/native/native_test.go"
      issue: "Direct validator test lacks wire-boundary and diagnostic-identity assertions."
  missing:
    - "Coverage metadata for the automated schema-admission checks."
    - "A serialization-to-decode-to-ToolError seam test with refusal-specific diagnostics."
  debug_session: .planning/debug/phase-15-schema-admission-ci.md
- gap_id: G-15-11
  truth: "The program-aware four-engine Schema 2 comparison refuses a corrupted engine document before it can report agreement, while /0 and /1 comparison behavior remains unchanged."
  status: resolved
  resolved_by: 15-09-PLAN.md
  resolved_at: 2026-09-22
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 11
  root_cause: "Plan 06 omits machine-readable coverage metadata, causing human-UAT fallback; no direct wrapper regression exercises both /0 and /1 legacy behavior, and the named CI phase gate remains stale at Phase 6."
  artifacts:
    - path: ".planning/phases/15-event-identity-lang-execution-2/15-06-SUMMARY.md"
      issue: "Missing coverage block causes legacy UAT fallback."
    - path: "internal/compiler/session/session_phase5_compare_test.go"
      issue: "No direct /0 and /1 wrapper-preservation regression."
    - path: ".github/workflows/ci.yml"
      issue: "Named phase gate is stale at Phase 6."
  missing:
    - "Coverage metadata for the automated peer-gate checks."
    - "Direct /0 and /1 wrapper-preservation regression."
    - "A current durable CI aggregate seam."
  debug_session: .planning/debug/phase-15-peer-gate-ci.md
- gap_id: G-15-12
  truth: "The unchanged shared-leaf diamond completes successfully across interpreter, O0, O3, and O3-LTO using Schema 2 evidence; the historical duplicate-pair collision control remains live."
  status: resolved
  resolved_by: 15-09-PLAN.md
  resolved_at: 2026-09-22
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 12
  root_cause: "Plan 07 omits machine-readable coverage metadata, causing a healthy four-tier diamond gate already run by CI to route to human UAT; CI's named phase gate is also stale at Phase 6, obscuring current aggregate provenance."
  artifacts:
    - path: ".planning/phases/15-event-identity-lang-execution-2/15-07-SUMMARY.md"
      issue: "Missing coverage block causes legacy UAT fallback."
    - path: ".github/workflows/ci.yml"
      issue: "Named phase gate is stale at Phase 6."
    - path: "internal/compiler/session/session_phase11_differential_test.go"
      issue: "Healthy four-tier gate needs published automated coverage provenance."
  missing:
    - "Coverage metadata for the automated four-tier diamond and collision checks."
    - "A current durable CI aggregate seam."
  debug_session: .planning/debug/phase-15-diamond-gate-ci.md
