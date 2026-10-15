---
phase: "37"
slug: "crc-32-consumer-interpreter-file-route-and-guide"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-15"
---

# Phase 37 — Validation Strategy

This draft is seeded from the research validation architecture. The planner
must replace the provisional rows with exact task IDs, threat IDs, test owners
and selectors. A named test found in source is not a newly passing check.

## Test Infrastructure

| Property | Value |
|---|---|
| Framework | Go 1.24 standard-library testing; existing bounded CLI/native runners; Node built-ins only for canonical prohibition proof where required |
| Config file | `go.mod`, existing CI and `.planning/config.json` |
| Quick run command | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase32PublicRouteAudit|TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory)$' -count=1 -v -timeout=600s` |
| Focused interpreter baseline | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase31ChecksumInterpreter$' -count=1 -v -timeout=600s` |
| Focused native baseline | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase31CaseMatrix$' -count=1 -v -timeout=600s` |
| Full suite command | Existing recurring CI owns broad checks; local project-wide/whole-package suites and broad native matrices are excluded by context D-21 |
| Estimated runtime | Unmeasured. The explicit test timeout is a kill bound, not a latency target. Measure maximum-size CRC before selecting its focused process budget |

## Sampling Rate

- After each task edit, run the exact named owner selectors from that task.
- After each wave, run the affected focused consumer, reader, route and guide
  selectors; compare each engine with literal expected tuples before parity.
- Before phase verification, capture all named feature and affected regression
  owners against a fixed source revision with actual RUN/PASS counts, no missing
  roots, and explicit skip dispositions. Hosted binding belongs to Phase 38.
- Every negative command must name the expected failure direction; compile,
  unrelated tool failure, timeout or zero selected roots cannot count as RED.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---|---|---|---|---|---|---|---|---|---|
| Planner assigns | Pending | Pending | APP-13 | Planner assigns | Ordinary-source CRC with independent literal values and maximum case | integration | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | MOD-07 | Planner assigns | One library; both output and module identities change under library edit | mutation | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | APP-14 | Planner assigns | Opaque CLI grammar, full admission and exact outcome mapping | CLI integration | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | APP-14 | Planner assigns | Open descriptor, regular input, actual cap+1 read and correct close/error behavior | unit/fault | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | APP-13, APP-14 | Planner assigns | Disabled evidence avoids retention while computation and outcomes continue | runtime/fault | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | DX-24 | Planner assigns | Actual guide commands from tracked-only inputs, expected tuples and drift control | documentation integration | Exact new selector to be declared in plan | Wave 0 | pending |
| Planner assigns | Pending | Pending | APP-13, APP-14 | Planner assigns | Recalled prohibitions have real supported bad/clean/default producer proof | enforcement | Canonical prohibition-enforcement request and exact descriptor to be declared in plan | Wave 0 | pending |

## Wave 0 Requirements

- [ ] Ground the pre-change route/empty-input/C baseline before implementation.
- [ ] Add independently generated CRC rows and byte recipes/digests.
- [ ] Add exact public matrix and finite malformed/opaque argument controls.
- [ ] Add independent production reader tests and actual reached fault controls.
- [ ] Add evidence-disabled result/retention tests; keep prior capture controls.
- [ ] Add all five reached CRC/reuse mutations and module-digest control.
- [ ] Add tracked-only guide execution and guide/fixture drift controls.
- [ ] Declare and exercise supported canonical prohibition descriptors early.
- [ ] No framework or runtime dependency installation is required.

## Manual-Only Verifications

No planned Phase 37 behavior requires human-only UAT: deterministic acceptance
has executable proxies. Subjective readability and full-loop cold/warm latency
remain unmeasured. Linux reader behavior and final dual-host receipts remain
explicit host lanes, not local passes. A root-mode unreadability or unavailable
sanitizer skip must be named and cannot turn the whole matrix green by omission.

## Validation Sign-Off

- [ ] Every task has exact automated verification or an explicit Wave 0 owner.
- [ ] No three consecutive tasks lack meaningful automated verification.
- [ ] All new selector references exist and actually run at phase closeout.
- [ ] No watch flags, project-wide local suite or broad native matrix.
- [ ] Actual maximum-case cost observed; no timing promoted into a guarantee.
- [ ] Unreadable/sanitizer/host skips are individually disclosed.
- [ ] `nyquist_compliant: true` only after the audit establishes compliance.

**Approval:** pending implementation and validation audit.
