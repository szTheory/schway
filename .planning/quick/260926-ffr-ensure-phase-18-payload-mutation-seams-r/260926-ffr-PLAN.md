---
id: 260926-ffr
phase: quick
plan: 260926-ffr
type: execute
mode: quick
status: planned
wave: 1
depends_on: []
files_modified:
  - internal/compiler/session/session_phase18_payload_test.go
autonomous: true
requirements:
  - CTL-03
must_haves:
  truths:
    - "Both Phase 18 wrong-slot mutation tests restore the package-global payload-slot seam even if their four-tier runner terminates the test goroutine."
    - "The existing eager restore after each successful runner call and the mutation anti-vacuity assertions remain intact."
    - "Both focused mutation controls still report an injected write and terminal-outcome disagreement."
  artifacts:
    - path: internal/compiler/session/session_phase18_payload_test.go
      provides: Deferred cleanup for both Phase 18 wrong-slot mutation tests
  key_links:
    - from: internal/compiler/session/session_phase18_payload_test.go
      to: internal/compiler/cgen/cgen.go
      via: Each SetPayloadSlotSwapForTest(true) restore closure is deferred before the fallible four-tier runner
---

<objective>
Close Phase 18 code review warning WR-01 by making both wrong-slot mutation controls restore their shared test seam on runner failure.

Purpose: Preserve deterministic test order and the CTL-03 mutation witness required by D-18-04 and D-18-05.
Output: A two-line test-only cleanup in `internal/compiler/session/session_phase18_payload_test.go`.
</objective>

<execution_context>
@~/.codex/gsd-core/workflows/execute-plan.md
@~/.codex/gsd-core/templates/summary.md
</execution_context>

<context>
@AGENTS.md
@.planning/STATE.md
@.planning/phases/18-branch-on-a-computed-value/18-CONTEXT.md
@.planning/phases/18-branch-on-a-computed-value/18-REVIEW.md
@internal/compiler/session/session_phase18_payload_test.go
@internal/compiler/cgen/cgen.go

WR-01 identifies `TestPhase18WrongSlotMutation` and `TestPhase18LongTagWrongSlotMutation`: each installs `SetPayloadSlotSwapForTest(true)`, then calls `phase11RunFourTiersWithSupplier`, which can call `t.Fatal` before the existing `restore()` line. The closure assigns the prior boolean value and is safe to call again. The seam contract itself says callers must defer restoration immediately.
</context>

<tasks>

<task type="auto">
  <name>Task 1: Defer both Phase 18 mutation-seam restorations</name>
  <files>internal/compiler/session/session_phase18_payload_test.go</files>
  <action>In `TestPhase18WrongSlotMutation` and `TestPhase18LongTagWrongSlotMutation`, insert `defer restore()` immediately after each `restore := cgen.SetPayloadSlotSwapForTest(true)` line, before `phase11RunFourTiersWithSupplier`. Retain each existing eager `restore()` following a successful runner call, so the active mutation interval stays as narrow as before; the deferred call covers `t.Fatal` and other early exits. Preserve the injected-write count checks and exact terminal-outcome disagreement assertions per D-18-04 and D-18-05. Change no production file, fixture, UAT record, or validation record.</action>
  <verify>
    <automated>GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase18WrongSlotMutation|TestPhase18LongTagWrongSlotMutation)$' -count=1</automated>
    <automated>git diff --check -- internal/compiler/session/session_phase18_payload_test.go</automated>
  </verify>
  <done>Each mutation test defers its restore closure on the line after seam installation while retaining its eager restore, and both named controls pass.</done>
</task>

</tasks>

<assumption_delta_decision>
The Phase 18 scan reported a pluralization cue in the roadmap sentence about “another tech_debt closeout.” Primary noun: the existing Phase 18 payload-slot mutation seam. Decision: no-change. This test cleanup introduces no second identity or representation.
</assumption_delta_decision>

<threat_model>
ASVS level 1; block on high. This test-only edit introduces no external input or production runtime boundary.

## Trust Boundaries

| Boundary | Description |
|----------|-------------|
| Test runner to package-global mutation seam | A fatal test exit can leave the cgen seam enabled for later tests in the same process. |

## STRIDE Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation Plan |
|-----------|----------|-----------|----------|-------------|-----------------|
| T-QFFR-01 | Tampering | Shared cgen payload-slot test seam | medium | mitigate | Defer both restore closures immediately after installation and retain existing eager restoration on success. |
</threat_model>

<source_audit>
| Source | Item | Task | Status |
|--------|------|------|--------|
| GOAL | Close WR-01 without changing Phase 18 behavior | 1 | COVERED |
| REQ | CTL-03 wrong-slot witness remains executable | 1 | COVERED |
| RESEARCH | Existing cgen seam contract requires immediate deferred restoration | 1 | COVERED |
| CONTEXT | D-18-04 and D-18-05 require automated mutation evidence | 1 | COVERED |
| CONTEXT | Other Phase 18 decisions and deferred ideas are outside this quick cleanup | — | EXCLUDED |
</source_audit>

<verification>
Run the two named Phase 18 mutation controls and scoped whitespace check. Inspect the diff to confirm exactly two added deferred calls in the declared test file.
</verification>

<success_criteria>
Both mutation controls pass, and an early test exit can no longer leave the package-global payload-slot seam enabled.
</success_criteria>

<output>
Create `.planning/quick/260926-ffr-ensure-phase-18-payload-mutation-seams-r/260926-ffr-SUMMARY.md` when done.
</output>
