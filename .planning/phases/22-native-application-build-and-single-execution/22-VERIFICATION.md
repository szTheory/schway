---
phase: 22-native-application-build-and-single-execution
verified: 2026-09-28T16:11:19Z
status: passed
score: 5/5 roadmap truths verified
covered_files:
  - .planning/REQUIREMENTS.md
  - .planning/phases/22-native-application-build-and-single-execution/22-01-PLAN.md
  - .planning/phases/22-native-application-build-and-single-execution/22-01-SUMMARY.md
  - .planning/phases/22-native-application-build-and-single-execution/22-02-PLAN.md
  - .planning/phases/22-native-application-build-and-single-execution/22-02-SUMMARY.md
  - .planning/phases/22-native-application-build-and-single-execution/22-03-PLAN.md
  - .planning/phases/22-native-application-build-and-single-execution/22-03-SUMMARY.md
  - cmd/lang/main.go
  - cmd/lang/main_test.go
  - examples/phase22/BINDINGS.md
  - examples/phase22/README.md
  - examples/phase22/identity.bindings.json
  - examples/phase22/identity.cases.json
  - examples/phase22/identity.expected.json
  - examples/phase22/identity.lang
  - examples/phase22/support.c
  - examples/phase22/support.h
  - examples/phase23/adapter.c
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/execution/execution.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/bindings_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_app_verify_test.go
covered_digest: "v1:sha256:c0055f165f24b8f352250692d878f1d228bfe41886662bd3920c88e74f87786f"
behavior_unverified: 0
overrides_applied: 0
---

# Phase 22: Native Application Build and Single Execution — Verification Report

**Phase Goal:** A developer can retain and execute a native application once on bounded real input, with application streams separate from compiler evidence.
**Verified:** 2026-09-28T16:11:19Z
**Status:** passed
**Re-verification:** No. The existing report had no `gaps:` section; this is an initial-mode refresh against the merged repository.

## Goal Achievement

The five roadmap success criteria are the contract for this verification. I checked their current source and wiring rather than treating the three plan summaries as proof. The prior completed UAT remains intact; its objective README contract is included among the freshly run CLI tests.

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | A relocated checkout can build and retain an executable from declared local C sources, headers, symbols, ABI inputs, and runtime dependencies; build does not launch the app; invalid inputs fail; relevant input changes affect identity. | ✓ VERIFIED | `BuildApplication` resolves the optional manifest before emitting/linking; `ResolveBindings` confines paths and validates declared inputs. Build receipt records input identity and honestly marks incomplete runtime closure non-cacheable. `TestPhase22RelocatedBindingsCLI`, `TestPhase22BindingsRejectInvalidInputs`, `TestPhase22BuildIdentityDeclaredInputMutations`, and launch-count tests all passed in the focused run. |
| 2 | Caller inputs 7 and 42 produce independent expected results, while malformed or oversized inputs have a defined refusal. | ✓ VERIFIED | `examples/phase22/identity.lang` and `identity.expected.json` provide the source and literal expectations. Generated entry parsing is bounded; `TestPhase22IdentityApplicationBuildAndRunCLI` and the README contract test passed. |
| 3 | One public app-run request launches once with defined streams and process outcomes, without parsing ordinary output as execution JSON or replaying tiers. | ✓ VERIFIED | CLI dispatch reaches `RunApplication` or `RunApplicationWithEvidence`; both converge on one child `Run`. `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, `TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus`, `TestPhase22ConcurrentRequestsLaunchIndependently`, and `TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute` passed. |
| 4 | Requested compiler evidence is separate from app streams and distinguishes disabled, incomplete, and capacity-exhausted capture from successful verification. | ✓ VERIFIED | `RunApplicationWithEvidence` captures through a bounded private file after launching the same single child, validates capture after wait, and publishes a separate `verified:false` report. `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22EvidenceMissingPartialAndCapacityControls`, and report-write failure controls passed. |
| 5 | Explicit differential verification uses isolated inputs and independent expected values, separates modeled foreign outcomes from actual host IO, and never runs on the ordinary app route. | ✓ VERIFIED | `app verify` alone calls `VerifyApplicationCasesFile`; cases compare interpreter/O0/O3 outcomes against checked-in 7/42 literals. Model-only outcome handling is isolated and reports false for host IO/physical cleanup claims. Positive/negative model controls and ordinary-route separation tests passed. |

**Score:** 5/5 roadmap truths verified (0 present, behavior-unverified).

## Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase22/identity.lang`, `identity.expected.json` | Checked scalar program and independent expected values | ✓ VERIFIED | Present, substantive, consumed by CLI and replay behavior. |
| `internal/compiler/session/session_app.go` (planned path) | Checked build path | ✓ VERIFIED — implementation path corrected | That exact file is absent. `BuildApplication` and `BuildApplicationFile` are implemented in `internal/compiler/session/session.go`; source traces and CLI integration tests verify the actual path. |
| `internal/compiler/native/native_app.go` | Retained build, receipt, single-child runner, evidence | ✓ VERIFIED | Substantive implementation wired from public CLI; focused native and CLI tests passed. |
| `internal/compiler/native/bindings.go` plus Phase 22 binding fixtures | Closed local C manifest resolution | ✓ VERIFIED | Manifest, C/header ABI fixture, and contract docs exist; relocation, invalid-input, and identity mutation tests passed. |
| `internal/compiler/session/session_app_verify.go`, `identity.cases.json` | Explicit independent replay service and fixture | ✓ VERIFIED | Wired only by explicit `app verify`; CLI tests pass independent answers and model controls. |
| `examples/phase22/README.md` | Public command and limitations contract | ✓ VERIFIED | Objective contract test passed; no subjective readability claim is made. |

## Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/lang/main.go` | `session/session.go` | build dispatch → `BuildApplicationFile` → check/core validation/emission | ✓ WIRED | Direct source trace; then `native.BuildApplication`. |
| `session/session.go` | `cgen/cgen_program.go` | checked build → `EmitApplication` → shared lowering | ✓ WIRED | Direct call and shared emitter source verified. |
| `cmd/lang/main.go` | `native/native_app.go` | app run dispatch → one-child runner | ✓ WIRED | Both normal and evidence modes reach the shared application run path. |
| `cmd/lang/main.go` | `native/bindings.go` | `--manifest` → session/native resolver | ✓ WIRED | Manifest argument flows to `ResolveBindings`; resolved input digests feed compile and receipt identity. |
| `cgen_program.go` | `native_app.go` | private event capture | ✓ WIRED | Generated application shell writes only when the runner supplies a private capture path. |
| `native_app.go` | evidence sidecar | wait → validate → atomically publish | ✓ WIRED | Status and build/input identity are built after the single child completes. |
| `cmd/lang/main.go` | `session_app_verify.go` | explicit `app verify` dispatch | ✓ WIRED | Ordinary app run does not enter this replay path. |

The generic `verify.key-links` checker returned false negatives because it checks direct textual target references and does not resolve Go package calls/indirection; its target-file result for the planned `session_app.go` is also superseded by the implementation path in `session.go`. Manual source call traces above plus passing reached integration tests verify the links; the false negatives are not being counted as green automated link checks.

## Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Identity app | argv token → bounded U64 → Lang result → stdout | Caller token parsed by generated application entry; decimal result emitted by the checked body | Yes | ✓ FLOWING |
| Local C build | manifest paths → source/header bytes and ABI inputs | Root-confined filesystem reads and content digests determine compilation/link inputs and receipt identity | Yes | ✓ FLOWING |
| App evidence | private capture file → validated capture → sidecar | Same child writes bounded events; parent validates status and publishes separately | Yes | ✓ FLOWING |
| Replay | fixture inputs and pinned expected values → interpreter/O0/O3 outcomes | Explicit case file and independent literal answers feed comparisons | Yes | ✓ FLOWING |

## Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| All Phase 22 CLI controls, including objective README contract, identity, one-run, evidence, replay, and model controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang -run '^TestPhase22' -count=1` | Exit 0; package passed in 1.379s. | ✓ PASS |
| Phase 22 native build, bindings, process outcomes, single launch, and evidence controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase22' -count=1` | Exit 0; package passed in 11.907s. | ✓ PASS |
| Session replay decoder/model controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase22|TestDecodeReplayCasesRejectsDuplicateJSONKeys)' -count=1` | Exit 0; package passed in 0.012s. | ✓ PASS |
| Application emitter and refusal controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/cgen -run '^TestPhase22' -count=1` | Exit 0; package passed in 0.014s. | ✓ PASS |
| Full current repository suite | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./...` | Passed on macOS after source commit `c50430d`; supplied as current merged-tree evidence, distinct from historical Phase 22 summary receipts. | ✓ PASS |
| Phase 23 adapter acquisition-failure cleanup, including the new empty-read/close-failure case | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^TestPhase23AcquireFailuresInitializeAndFreePartialAllocations$' -count=1 -v` | Exit 0; named test passed on the current tree (0.073s). The latest shared-test-file edit exercises the Phase 23 adapter; `native_app.go` and the Phase 22 public runner were unchanged by `c50430d`. | ✓ PASS |

No server or external service was started. All newly run checks and the supplied full-suite result are local macOS evidence. No hosted CI run, Linux host run, or Linux result is claimed. The Phase 23 close-failure amendment changes `examples/phase23/adapter.c` and its fault-injection case in `native_app_test.go`; the named current-tree test passed. Runtime dependency closure remains explicitly incomplete and non-cacheable when host facts are unknown.

## Test Quality Audit

| Test File | Linked requirement | Active | Skipped | Circular | Assertion level | Verdict |
|---|---|---:|---:|---|---|---|
| `cmd/lang/main_test.go` | APP-02–APP-06, EVD-11 | Yes | 0 skip markers found | No; independent checked-in answers and negative controls | Behavioral/value | PASS |
| `internal/compiler/native/native_app_test.go`, `bindings_test.go` | APP-02, APP-04–APP-06, FFI-02 | Yes | 0 skip markers found | No; process outputs and input mutations are asserted | Behavioral/value | PASS |
| `internal/compiler/cgen/cgen_program_test.go`, session replay tests | APP-03, EVD-11 | Yes | 0 skip markers found | No; independent fixtures and refusal controls | Behavioral/value | PASS |

**Disabled tests on requirements:** 0 found. **Circular patterns:** 0 found in reviewed requirement evidence. **Insufficient assertions:** 0 found.

## Decision Coverage

The decision-coverage gate returned 7/7 trackable CONTEXT.md decisions honored; no unhonored decisions.

## Probe Execution

No Phase 22-declared probes or conventional `scripts/*/tests/probe-*.sh` files were found.

## Requirements Coverage

| Requirement | Source plans | Status | Evidence |
|---|---|---|---|
| APP-02 | 22-01, 22-02 | ✓ SATISFIED | Retained build, zero build launches, relocation, and declared runtime dependency receipt behavior. |
| APP-03 | 22-01 | ✓ SATISFIED | Bounded identity input, literal 7/42 expectations, malformed/oversized controls. |
| APP-04 | 22-01, 22-03 | ✓ SATISFIED | One-child launch tests and ordinary-route replay exclusion. |
| APP-05 | 22-01 | ✓ SATISFIED | Byte streams and exit/signal/timeout/launch outcome controls. |
| APP-06 | 22-03 | ✓ SATISFIED | Separate evidence states and failure controls. |
| FFI-02 | 22-02 | ✓ SATISFIED | Local C declaration, ABI/input failures, relocation, content-bound identity. |
| EVD-11 | 22-03 | ✓ SATISFIED | Explicit independent replay and modeled outcome scope limits. |

All seven Phase 22 requirement IDs in the plans match the Phase 22 traceability table; no orphaned Phase 22 requirement was found.

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unreferenced debt markers, placeholders, or behavior stubs found in covered implementation and example files. Empty slice returns found by broad pattern scans are control/data-shape returns, not user-visible stubs. | — | — |

The plan prohibitions hold in code and tests: linking local C does not claim arbitrary C behavior; complete same-run capture does not claim semantic verification; modeled outcomes do not claim actual host IO or physical cleanup.

## Human Verification Required

None. Phase 22 has no outstanding human UAT: `22-UAT.md` is already `status: complete`, and the objective README contract test is part of the current passing CLI test group. The retired subjective readability criterion is not claimed as proven and is not a release gate.

## Gaps Summary

No roadmap truth, requirement, wired artifact, or data flow is missing. The absent planned `session_app.go` file is an implementation-path correction: the checked build path resides in `session.go`, and the actual public call chain is exercised by passing CLI integration tests. This refresh verifies the merged tree; plan-summary test claims remain historical receipts except where the current runs above independently reproduce them. The full repository test suite also passed after the Phase 23 merge. Host runtime closure remains explicitly non-cacheable when incomplete, and this report makes no Linux or subjective README readability claim.

---

_Verified: 2026-09-28T16:11:19Z_
_Verifier: the agent (gsd-verifier)_
