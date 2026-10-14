---
phase: "35"
slug: bitwise-operators-literal-shifts-and-hex-literals
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-14"
---

# Phase 35 — Validation Strategy

> Planning contract, not executed feature evidence. The planner fills the per-task map from final plans and preserves exact named package selectors. Research test names marked proposed are Wave 0 obligations until they exist and run.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 standard-library testing; existing Clang native runner |
| Config | go.mod; no added test dependency |
| Existing quick baseline | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase34BoolNameBaseline|TestPhase34IdentityBaseline)$' -count=1 -v` |
| Phase gate | Union of final PLAN automated named selectors; no local whole-package/project suite or broad native matrix |
| Runtime | Unmeasured at planning; normalized commands must have a bounded 600-second timeout, with observed runtime recorded during execution |

## Sampling Rate

- Before compiler edits: commit new immutable enumerated C corpus and numeric-route policy baselines; require the reached byte-drift control.
- Per task: run the focused owner selector after the last edit, inspect exit status and require each selected test name in verbose/JSON run output.
- Per wave: corresponding final-plan selectors, including inherited relevant controls.
- Before phase completion: all final scoped requirement/decision/security controls freshly passed; no matcher with zero test executions may count as evidence.
- Existing recurring CI owns broad lanes; Phase 38 owns hosted two-host codedigest binding. Local host evidence cannot close missing host receipts.

## Per-Task Verification Map

Plan/wave IDs and exact new selectors are populated by the planner before plan checking. These rows describe obligations, not passing receipts.

| Requirement | Behavior | Type | Existing evidence / Wave 0 obligation | Status |
|---|---|---|---|---|
| BIT-01 | U64 ^/&/\| answers, precedence, independent value transfers, causal event-free controls | unit, mutation, native differential | New per-op value-transfer/engine mutants and causal fabricated-event controls; O0/O3/LTO vectors | pending W0 |
| BIT-02 | Number-leaf shift count, canonical immediate 0..63, raw core fields, unsigned C, no dynamic/64+ shifts | unit, mutation, focused sanitizer | New check/core/origin admission/range controls; static site count and dynamic/64+ C mutants; focused UBSan | pending W0 |
| BIT-03 | Existing spellings/policies, uppercase legacy hex, overflow, exact malformed ownership and lossless/idempotent format | unit, golden | Existing Phase19 token/range/format cases plus new 010/0B route policy and depth 64/65/nested-generic cases | pending W0 |
| BIT-04 | Adler shift preserves literal outputs and module/cleanup behavior | integration, mutation | Unchanged TestPhase29InterpreterFixtureVectors and TestPhase29NativeFixtureVectors in session; TestPhase31CaseMatrix in native; new reached wrong-Adler-shift control | pending |
| BIT-01/02 runnable gain | Public conformance outcome 264 through unused-Byte bridge; Byte use refused; U64 input-required unchanged | CLI integration, admission mutation | New unused-Byte source/core/origin controls and both public engines; bare 264 newline separately through existing application seams | pending W0 |

The research report's example places `TestPhase31CaseMatrix` under session; actual source is `internal/compiler/native/native_phase31_matrix_test.go`. Plans must use the grounded native package, not copy the mistaken command. Test names not yet present must be clearly new artifacts, never historical passing tests.

## Wave 0 Requirements

- [ ] Immutable baseline corpus with emitting/non-emitting cases, all selected emitters and reached changed-byte control, committed before compiler changes.
- [ ] Numeric compatibility (`010`, `0B10`, Bool names), grouped grammar, exact lexical ownership, expression depth/recovery and token/generic formatter controls.
- [ ] Independent source/core/origin immediate-only shape/range and unused-Byte bridge controls; scalar transfer fault sinks.
- [ ] Interpreter/native literal vectors, per-op causal event-free/forged events, evidence capacity/projection inventories and generated shift static scan mutants.
- [ ] Focused O0/O3/LTO and UBSan feature checks, manifest admission/refusal/limit census and bidirectional controls.
- [ ] Existing Adler/checksum/adler_low literal answers unchanged and reached wrong-shift control.
- [ ] Exact task map, threat IDs, real test paths/names and fails_when bindings filled by final plans.

## Manual-Only Verifications

None intended: the scoped contracts have deterministic proxies. Human readability and whole-loop cold/warm latency remain unmeasured and are not acceptance claims.

## Validation Sign-Off

- [ ] Every task has an executable named verify or explicit Wave 0 dependency.
- [ ] No three consecutive implementation tasks lack verification.
- [ ] Every expected test name was executed, all fault controls reached/restored.
- [ ] No watch mode, whole-project local test, push/dispatch or premature hosted binding.
- [ ] Validation lifecycle changes to validated/compliant only from fresh execution evidence.

**Approval:** pending execution evidence; planning artifact authorized by the user-delegated auto-advance workflow.
