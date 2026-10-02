---
phase: "26"
slug: "checked-scalar-sum"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-02"
---

# Phase 26 — Validation Strategy

> Draft evidence contract for execution planning. No Phase 26 checks have been run.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 standard `testing` package |
| **Config file** | Go module defaults; no separate test config identified |
| **Quick run command** | `go test ./internal/compiler/syntax ./internal/compiler/core ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/execution ./internal/compiler/executionpeer ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/native ./internal/compiler/session ./cmd/schway` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | Not measured; use focused package runs during implementation and record cold/warm timings if timing is reported. |

---

## Sampling Rate

- **After every task commit:** Run the smallest affected package tests plus the task's decisive semantic control.
- **After every plan wave:** Run the focused package command above, including the relevant interpreter and native app controls.
- **Before `$gsd-verify-work`:** Run `go test ./...` and the public `schway app run` witness/negative controls.
- **Max feedback latency:** Not yet measured; keep task feedback on focused packages and record observed cold/warm distributions during execution.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 26-01-T1 | 26-01 | 1 | U64-01, FLOW-01, FLOW-02, APP-07 | T-26-01, T-26-02 | Checked-in source reaches typed scalar CFG, independent peers, interpreter, C17, and retained app; `10` gives `55\n`. | End-to-end tracer | `go test ./internal/compiler/session -run '^TestPhase26TracerAppSum10$' -count=1` | New witness/test pending | ⬜ pending |
| 26-01-T2 | 26-01 | 1 | FLOW-01, FLOW-02, APP-07 | T-26-01, T-26-03 | Source recovery/spans, Bool typing, 0/10/1000, mutable-local limit, rejected 1001 without stdout. | Core/session boundary controls | `go test ./internal/compiler/core ./internal/compiler/session -run '^TestPhase26(Frontend|CoreInventory|AppBoundary|TracerAppSum10)$' -count=1` | Existing test homes; cases pending | ⬜ pending |
| 26-02-T1 | 26-02 | 2 | U64-01 | T-26-04 | Direct MAX+1 operation reaches checked failure; MAX+0 is exact; Go and C check before accepting a result. | Interpreter/C17 unit and native semantics | `go test ./internal/compiler/interp ./internal/compiler/cgen -run '^TestPhase26CheckedAdd(Interpreter|C17)$' -count=1` | New source/test pending | ⬜ pending |
| 26-02-T2 | 26-02 | 2 | U64-01, APP-07 | T-26-05 | Child overflow exit 65, bounded exact stderr, empty stdout, distinct from tool/capture errors. | Session/native integration | `go test ./internal/compiler/session ./internal/compiler/native -run '^TestPhase26(OverflowProcessOutcome|FailureChannels)$' -count=1` | Existing test homes; cases pending | ⬜ pending |
| 26-03-T1 | 26-03 | 2 | FLOW-01, FLOW-02 | T-26-06, T-26-07 | Each peer converges independently on scalar facts and refuses a test-forced deterministic analysis cap. | Checker/peer fixed-point mutation | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase26(FixedPoint|AnalysisExhaustion|PeerIndependence)$' -count=1` | Existing peer test homes; cases pending | ⬜ pending |
| 26-03-T2 | 26-03 | 2 | FLOW-02 | T-26-07, T-26-08 | Owner/resource/loan/provenance carries and unsupported event kinds are refused; scalar U64/Bool copy may repeat. Existing pathoracle cycle refusal remains an acyclic-oracle regression, not the loop proof. | Source/core negative controls | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run '^TestPhase26(BackEdgeAuthority|ScalarCopyCycleBoundary)$' -count=1` | Existing peer test homes; cases pending | ⬜ pending |
| 26-04-T1 | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-11 | Interpreter/native repeated-copy events use independent 0/1/2 ordinals with canonical one-shot compatibility. | Event encoding/engine tests | `go test ./internal/compiler/execution ./internal/compiler/interp ./internal/compiler/cgen -run '^TestPhase26(EventOccurrenceEncoding|InterpreterRepeatedCopy|NativeRepeatedCopy)$' -count=1` | Existing test homes; cases pending | ⬜ pending |
| 26-04-T2 | 26-04 | 3 | FLOW-01, FLOW-02, APP-07 | T-26-12 | Independent peer accepts 0/1/2 and rejects duplicate/gapped/reordered/wrong-invocation/wrong-site/forged non-scalar events. | Independent event peer | `go test ./internal/compiler/executionpeer -run '^TestPhase26PeerOccurrenceOrder$' -count=1` | Existing peer test home; cases pending | ⬜ pending |
| 26-05-T1 | 26-05 | 4 | U64-01, FLOW-01, APP-07 | T-26-09 | Each engine matches hand-pinned 0/55/500500; 1001 fails; reached wrong-result and skipped-iteration controls fail. | End-to-end expected-answer/mutation | `go test ./internal/compiler/session -run '^TestPhase26(ExactSumMatrix|ReachedWrongResult|SkippedIteration)$' -count=1` | Existing test home; cases pending | ⬜ pending |
| 26-05-T2 | 26-05 | 4 | U64-01, FLOW-02, APP-07 | T-26-10 | Exact public CLI 0/10/1000/1001/overflow streams/status; native app capture accepts complete 0/1/2 events and reports forced capacity separately from child result; frontend recovery remains stable. | Public command/app evidence | `go test ./cmd/schway ./internal/compiler/native ./internal/compiler/session ./internal/compiler/syntax -run '^TestPhase26(PublicAppCLI|EvidenceCapacity|ExactSumMatrix|OverflowProcessOutcome|FrontendRecovery|RepeatedCopyEvidence|RepeatedCopyCapacity)$' -count=1` | Existing test homes; cases pending | ⬜ pending |

*Task, plan, wave, and threat mappings are assigned. All status rows remain pending until execution produces evidence.*

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

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Review public overflow diagnostics and stream/status separation | U64-01 | Human check that bounded stderr is deterministic and understandable while stdout is empty and compiler/tool/evidence failures remain distinct. | Run the near-max source witness through each public engine route; inspect captured stdout, stderr, and process outcome against the pinned contract. |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < measured target
- [ ] `nyquist_compliant: true` set in frontmatter after validation

**Approval:** pending plan decomposition and execution evidence
