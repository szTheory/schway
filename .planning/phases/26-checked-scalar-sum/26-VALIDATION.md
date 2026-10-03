---
phase: "26"
slug: "checked-scalar-sum"
status: in_progress
nyquist_compliant: false
wave_0_complete: true
created: "2026-10-02"
---

# Phase 26 — Validation Strategy

> Tests and native runs execute only in hosted GitHub Actions under project policy. No local Go tests or program runs are permitted.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 standard `testing` package |
| **Config file** | Go module defaults; no separate test config identified |
| **Quick run command** | `gh workflow run ci.yml --ref phase26-validation -f phase26_focused=true` |
| **Full suite command** | `gh workflow run ci.yml --ref phase26-validation` |
| **Observed runtime** | Focused run 37092337472: Ubuntu about 2m18s and macOS about 3m14s; cold/warm distribution not measured. |

---

## Sampling Rate

- **After every task commit:** Run the affected package tests and decisive controls in focused hosted CI.
- **After every plan wave:** Run the focused hosted workflow, including interpreter, native app, and CLI controls.
- **Before `$gsd-verify-work`:** Run the full hosted workflow and public CLI witness/negative controls.
- **Max feedback latency:** See hosted timings above; cold/warm distributions remain unmeasured.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 26-01-T1 | 26-01 | 1 | U64-01, FLOW-01, FLOW-02, APP-07 | T-26-01, T-26-02 | Checked-in source reaches typed scalar CFG, independent peers, interpreter, C17, and retained app; `10` gives `55\n`. | End-to-end tracer | `go test ./internal/compiler/session -run '^TestPhase26TracerAppSum10$' -count=1` | Existing witness/test | ✅ hosted |
| 26-01-T2a | 26-01 | 1 | FLOW-01, FLOW-02, APP-07 | T-26-01, T-26-03 | Core operation inventory for the scalar loop. | Core boundary | `go test ./internal/compiler/core -run '^TestPhase26CoreInventory$' -count=1` | Existing test home | ✅ hosted |
| 26-01-T2b | 26-01 | 1 | FLOW-01, FLOW-02, APP-07 | T-26-01, T-26-03 | Source recovery and bounded app inputs. | Session controls | `go test ./internal/compiler/session -run '^TestPhase26Frontend$' -count=1` | Existing test home | ✅ hosted |
| 26-02-T1a | 26-02 | 2 | U64-01 | T-26-04 | Interpreter MAX+1 checked failure. | Interpreter test | `go test ./internal/compiler/interp -run '^TestPhase26CheckedAddInterpreter$' -count=1` | Existing test home | ✅ hosted |
| 26-02-T1b | 26-02 | 2 | U64-01 | T-26-04 | C17 MAX+1 checked failure. | Cgen test | `go test ./internal/compiler/cgen -run '^TestPhase26CheckedAddC17$' -count=1` | Existing test home | ✅ hosted |
| 26-02-T2a | 26-02 | 2 | U64-01, APP-07 | T-26-05 | Child overflow exit and exact streams. | Session integration | `go test ./internal/compiler/session -run '^TestPhase26OverflowProcessOutcome$' -count=1` | Existing test home | ✅ hosted |
| 26-02-T2b | 26-02 | 2 | U64-01, APP-07 | T-26-05 | Tool/capture errors distinct from child status. | Native integration | `go test ./internal/compiler/native -run '^TestPhase26FailureChannels$' -count=1` | Existing test home | ✅ hosted |
| 26-03-T1a | 26-03 | 2 | FLOW-01, FLOW-02 | T-26-06, T-26-07 | Checker fixed point. | Checker test | `go test ./internal/compiler/check -run '^TestPhase26FixedPoint$' -count=1` | Existing test home | ✅ hosted |
| 26-03-T1b | 26-03 | 2 | FLOW-01, FLOW-02 | T-26-06, T-26-07 | Core peer analysis exhaustion. | Core peer | `go test ./internal/compiler/corevalidate -run '^TestPhase26AnalysisExhaustion$' -count=1` | Existing test home | ✅ hosted |
| 26-03-T1c | 26-03 | 2 | FLOW-01, FLOW-02 | T-26-06, T-26-07 | Origin peer analysis exhaustion. | Origin peer | `go test ./internal/compiler/originvalidate -run '^TestPhase26AnalysisExhaustion$' -count=1` | Existing test home | ✅ hosted |
| 26-03-T2a | 26-03 | 2 | FLOW-02 | T-26-07, T-26-08 | Unsupported back-edge authority and event cases are refused. | Checker control | `go test ./internal/compiler/check -run '^TestPhase26BackEdgeAuthority$' -count=1` | Existing test home | ✅ hosted |
| 26-03-T2b | 26-03 | 2 | FLOW-02 | T-26-07, T-26-08 | Existing composition cycle guard remains closed. | Path oracle control | `go test ./internal/compiler/pathoracle -run '^TestCompositionCycleGuardFailsClosed$' -count=1` | Existing test home | ✅ hosted |
| 26-04-T1a | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-11 | Occurrence encoding. | Execution test | `go test ./internal/compiler/execution -run '^TestPhase26EventOccurrenceEncoding$' -count=1` | Existing test home | ✅ hosted |
| 26-04-T1b | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-11 | Interpreter repeated-copy ordinals. | Interpreter test | `go test ./internal/compiler/interp -run '^TestPhase26InterpreterRepeatedCopy$' -count=1` | Existing test home | ✅ hosted |
| 26-04-T1c | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-11 | Native repeated-copy ordinals. | Cgen test | `go test ./internal/compiler/cgen -run '^TestPhase26NativeRepeatedCopy$' -count=1` | Existing test home | ✅ hosted |
| 26-04-T2 | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-12 | Independent peer accepts 0/1/2 and rejects duplicate/gapped/reordered/wrong-invocation/wrong-site/forged non-scalar events. | Independent event peer | `go test ./internal/compiler/executionpeer -run '^TestPhase26PeerOccurrenceOrder$' -count=1` | Existing peer test home | ✅ hosted |
| 26-05-T1 | 26-05 | 4 | U64-01, FLOW-01, APP-07 | T-26-09 | Each engine matches hand-pinned 0/55/500500; 1001 fails; reached wrong-result and skipped-iteration controls fail. | End-to-end expected-answer/mutation | `go test ./internal/compiler/session -run '^TestPhase26ExactSumMatrix$' -count=1` | Existing test home | ✅ hosted in 37092337472 |
| 26-05-T2a | 26-05 | 4 | U64-01, APP-07 | T-26-10 | Exact public CLI streams/status for sum matrix and overflow. | Public command | `go test ./cmd/schway -run '^TestPhase26PublicAppCLI$' -count=1` | Existing test home | ✅ hosted in 37093111718 |
| 26-05-T2b | 26-05 | 4 | FLOW-02, APP-07 | T-26-10 | Forced capacity outcome retains actual child streams/status and Verified=false. | Native app capture | `go test ./internal/compiler/native -run '^TestPhase26EvidenceCapacity$' -count=1` | Existing test home | ✅ hosted in 37093111718 |
| 26-05-T2c | 26-05 | 4 | FLOW-01, FLOW-02, APP-07 | T-26-10 | Source-authored immutable copy event ordinals 0/1/2 with independent peers. | Source-to-native evidence | `go test ./internal/compiler/session -run '^TestPhase26SourceRepeatedCopyEvidence$' -count=1` | Existing test home | ✅ hosted in 37093111718 |
| 26-05-T2d | 26-05 | 4 | FLOW-01 | T-26-10 | Malformed source spans, recovery, and comment roundtrip. | Syntax recovery | `go test ./internal/compiler/syntax -run '^TestPhase26FrontendRecovery$' -count=1` | Existing test home | ✅ hosted in 37093111718 |

*Plan 01–04 rows refer to completed hosted evidence. The Plan 05 controls passed on Ubuntu and macOS in 37092337472 and 37093111718; phase validation now awaits the hosted corpus receipt and full gate.*

---

## Wave 0 Requirements

- [ ] Check in the ordinary `sum_to_n` Schway source witness before broadening cyclic-CFG admission.
- [ ] Add syntax/formatter round trips and malformed-input/source-span coverage for `var`, `if/else`, `<`, checked `+`, and pre-tested `while`.
- [ ] Add independent fixed-point, deterministic budget-exhaustion, and scalar assignment/branch tests for checker, core validator, and origin validator.
- [ ] Add stable refusal controls for owner, resource, loan, and loan-derived provenance back-edge cases in each applicable validator.
- [ ] Pin hand-derived interpreter and native outcomes independently for `n = 0`, `10`, `1000`, rejected `1001`, and a direct near-`U64::MAX` overflow witness.
- [ ] Add reached wrong-result and skipped-iteration controls; emit 0/1/2 occurrence ordinals for repeated scalar-copy events, independently reject forged sequences, and force evidence capacity exhaustion separately.

## Threat Coverage

| Threat IDs | Plan / wave | Decisive control |
|---|---|---|
| T-26-01, T-26-02, T-26-03 | 26-01 / 1 | Bounded source/analysis and typed tracer, source defect before output. |
| T-26-04, T-26-05 | 26-02 / 2 | Direct checked overflow; child outcome separated from tool/evidence failure. |
| T-26-06, T-26-07, T-26-08 | 26-03 / 2 | Deterministic analysis cap, independent back-edge authority and unsupported-event refusals. |
| T-26-11, T-26-12 | 26-04 / 3 | Dynamic occurrence encoding, bounded producer signals, independent sequence validation. |
| T-26-09, T-26-10 | 26-05 / 4 | Hand-pinned expected answers, reached mutations, exact public CLI matrix, native app capture/capacity outcome separation. |

---

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < measured target
- [ ] `nyquist_compliant: true` set in frontmatter after validation

**Approval:** pending the corrected Plan 05 hosted run, validation-corpus receipt, and full hosted gate.
