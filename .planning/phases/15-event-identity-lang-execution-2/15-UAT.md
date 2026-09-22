---
status: complete
phase: 15-event-identity-lang-execution-2
source: 15-01-SUMMARY.md, 15-02-SUMMARY.md, 15-03-SUMMARY.md, 15-04-SUMMARY.md, 15-05-SUMMARY.md, 15-06-SUMMARY.md, 15-07-SUMMARY.md
started: 2026-09-21T23:54:47Z
updated: 2026-09-22T09:15:27Z
---

## Current Test

[testing complete]

## Tests

### 1. Schema 2 execution admission
expected: A multi-function execution document with canonical invocation identities is accepted, while a non-canonical invocation, duplicate activation identity, invalid event kind, or invalid caller-owned callee identity is refused with an actionable diagnostic.
result: issue
reported: "integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
severity: major

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
result: issue
reported: "integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
severity: major

### 12. Four-tier shared-leaf diamond gate
expected: The unchanged shared-leaf diamond completes successfully across interpreter, O0, O3, and O3-LTO using Schema 2 evidence; the historical duplicate-pair collision control remains live.
result: issue
reported: "integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
severity: major

## Summary

total: 12
passed: 9
issues: 3
pending: 0
skipped: 0
blocked: 0

## Gaps

- gap_id: G-15-1
  truth: "A multi-function execution document with canonical invocation identities is accepted, while a non-canonical invocation, duplicate activation identity, invalid event kind, or invalid caller-owned callee identity is refused with an actionable diagnostic."
  status: failed
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 1
  artifacts: []
  missing: []
- gap_id: G-15-11
  truth: "The program-aware four-engine Schema 2 comparison refuses a corrupted engine document before it can report agreement, while /0 and /1 comparison behavior remains unchanged."
  status: failed
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 11
  artifacts: []
  missing: []
- gap_id: G-15-12
  truth: "The unchanged shared-leaf diamond completes successfully across interpreter, O0, O3, and O3-LTO using Schema 2 evidence; the historical duplicate-pair collision control remains live."
  status: failed
  reason: "User reported: integration/e2e/smoke/seam test... automate the world devops mindset shift left,  goal is 0 human verificaiton / uat required (even shift to CI ... iff recurring value there)"
  severity: major
  test: 12
  artifacts: []
  missing: []
