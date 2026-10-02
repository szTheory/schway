---
phase: 22-native-application-build-and-single-execution
verified: 2026-10-02T14:02:46Z
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
  - .planning/phases/22-native-application-build-and-single-execution/22-CONTEXT.md
  - .planning/phases/22-native-application-build-and-single-execution/22-REVIEW.md
  - .planning/phases/22-native-application-build-and-single-execution/22-UAT.md
  - cmd/schway/main.go
  - cmd/schway/main_test.go
  - examples/phase22/BINDINGS.md
  - examples/phase22/README.md
  - examples/phase22/identity.bindings.json
  - examples/phase22/identity.cases.json
  - examples/phase22/identity.expected.json
  - examples/phase22/identity.schway
  - examples/phase22/support.c
  - examples/phase22/support.h
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/bindings_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
covered_digest: "v1:sha256:f3d1f675cbb445aa770ee29849d67ee0ac29881a84248c17e6b424c4cbf7def2"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 7
  total: 7
  not_honored: []
---

# Phase 22: Native Application Build and Single Execution — Verification Report

**Phase Goal:** A developer can retain and execute a native application once on bounded real input, with application streams separate from compiler evidence.
**Verified:** 2026-10-02T14:02:46Z
**Status:** passed
**Re-verification:** No. The existing report had no `gaps:` section, so this is an initial-mode refresh.

This refresh checked all five roadmap success criteria against current source and ran named, criterion-relevant tests from this checkout on Darwin 25.6.0 arm64 at revision `704199cce861dc715a98b099ef31394afda37062`. The prior report's hosted CI receipt is not used as current evidence. No full test suite or cross-host check was run. The objective README contract UAT in `22-UAT.md` remains complete at 1/1 from 2026-09-27; it was neither modified nor replayed. That automated contract does not claim subjective readability.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | From a relocated checkout, a developer can build a retained executable with declared local C sources, headers, symbols, ABI inputs, and runtime dependencies. Building performs no application effects; missing/incompatible inputs fail clearly and relevant input changes invalidate artifact/evidence identity. | ✓ VERIFIED | `native.BuildApplication` stages and compiles without starting the artifact; `ResolveBindings` snapshots root-confined declared inputs and checks headers, ABI typedefs, and symbols; receipt identity includes source, C/header/manifest, ABI, compiler, target, flags, runtime declaration, and executable. `TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce`, `TestPhase22RelocatedBindingsCLI`, `TestPhase22BindingsRejectInvalidInputs`, and `TestPhase22BuildIdentityDeclaredInputMutations` passed. The checked-in contract explicitly reports host dependency closure as incomplete and non-cacheable. |
| 2 | An ordinary scalar source program run with caller-selected inputs `7` and `42` produces independently specified corresponding results; malformed and oversized input has a defined outcome. | ✓ VERIFIED | `examples/phase22/identity.schway` and `identity.expected.json` provide the source and independent cases. The generated application entry bounds and parses the token before invoking the checked body. `TestPhase22IdentityApplicationBuildAndRunCLI` passed the 7/42 cases and empty, signed, nonnumeric, overflowing, and oversized refusals. |
| 3 | One public application-run request launches the selected artifact exactly once, with defined stdout, stderr, exit, and signal behavior; ordinary output and successful stderr are accepted without execution-JSON parsing or hidden interpretation/tier replay. | ✓ VERIFIED | `cmd/schway/main.go` routes ordinary `app run` to the retained native runner. `RunApplication` validates the receipt/digest, creates one child, forwards streams directly, and waits once. `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, `TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce`, and `TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute` passed controls for raw streams, exit, signal, timeout, launch/receipt errors, transport refusal, and one launch. |
| 4 | A developer can request execution evidence through a separate channel and distinguish disabled, incomplete, and capacity-exhausted states from successful verification. | ✓ VERIFIED | `RunApplicationWithEvidence` captures through a private file from the same process, validates after wait, and writes a separate report with `verified:false`. `TestPhase22EvidenceDisabledCompleteAndStreamIsolation` and `TestPhase22EvidenceMissingPartialAndCapacityControls` passed, covering disabled/complete/incomplete/capacity states, unchanged application bytes, and one launch. |
| 5 | Explicit differential verification uses isolated or replayable inputs and independent expected answers; declared foreign outcomes are distinguished from actual host IO, and verification replay does not occur on the ordinary application route. | ✓ VERIFIED | The explicit `app verify` route calls `VerifyApplicationCases`, compares interpreter and generated C O0/O3 outcomes against literal case expectations, and keeps model-only scripted outcomes in a separate adapter. Reports mark actual host IO and physical cleanup false. `TestPhase22AppVerifyIndependentIdentityCases`, `TestPhase22AppVerifyModelOnlyOutcomes`, and `TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute` passed; model controls covered missing, duplicate, unconsumed, and mismatched outcomes. |

**Score:** 5/5 truths verified (0 present, behavior-unverified).

### Plan Must-Have Crosswalk

| Plan | Must-have / prohibition | Status | Evidence |
|---|---|---|---|
| 22-01 | Retained build without app launch; relocation; bounded identity inputs and refusal; exactly one launch; byte streams and defined process outcomes | ✓ VERIFIED | Current-checkout tests listed under truths 1–3 passed. The plan's historical `identity.lang` and `session_app.go` paths have since been renamed/consolidated to `identity.schway` and `session.go`; current CLI wiring and implementation were inspected. |
| 22-02 | Relocatable declared C/header/symbol/ABI/runtime inputs; clear failure before publication; relevant mutation changes identity or stays explicitly non-cacheable | ✓ VERIFIED | Binding resolution, relocation, negative declarations, and identity mutations passed their named tests. Documentation and receipt code state that host closure remains incomplete and cacheability is false. |
| 22-02 | A local-C declaration must not imply arbitrary C behavior, ownership, or resource behavior is proven | ✓ VERIFIED | `BINDINGS.md` describes trusted C build authority and the limits of compile/link/type checks; no arbitrary C behavior or ownership result is claimed. |
| 22-03 | Separate evidence states; no second run or replay for capture; explicit isolated differential with model-only case; fail closed without host IO/cleanup claims | ✓ VERIFIED | Evidence, identity replay, and model outcome tests passed. The report schema sets evidence `verified:false`; verification reports set `actual_host_io:false` and `physical_cleanup:false`. |
| 22-03 | Complete capture is not an independent semantic verdict; modeled foreign success is not source FFI, actual host IO, or physical cleanup | ✓ VERIFIED | Report construction and model adapter are separate; prohibited claims are explicitly false in code/docs and negative controls pass. |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `cmd/schway/main.go`, `cmd/schway/main_test.go` | Public build/run/verify dispatch and CLI controls | ✓ VERIFIED | Build reaches session validation/emission; ordinary run reaches the one-child runner; only explicit verify reaches replay. |
| `internal/compiler/session/session.go` | Checked application build and manifest path | ✓ VERIFIED | Checks source and core, resolves supported entry, validates local binding declarations, emits once, then builds. |
| `internal/compiler/cgen/cgen.go`, `cgen_program.go` | Shared checked body lowering with application shell | ✓ VERIFIED | `EmitApplication` delegates to the shared program emitter with its application shell; emitter sharing test passed. |
| `internal/compiler/native/native_app.go` | Retained build receipt, one-run path, evidence channel | ✓ VERIFIED | Staging/publication, receipt and executable validation, direct streams, one child, private capture, bounded status report are implemented. |
| `internal/compiler/native/bindings.go` | Closed local-C resolver and content identity | ✓ VERIFIED | Strict schema, path/symlink confinement, snapshot, include checks, ABI probe, and identity inventory are implemented and exercised. |
| `internal/compiler/session/session_app_verify.go` | Explicit differential/model verifier | ✓ VERIFIED | Bounded cases, independent expected values, interpreter/O0/O3 comparison, and verifier-only model outcomes are implemented and exercised. |
| `examples/phase22/*` | Identity witness, case/expected data, local-C inputs, public contract | ✓ VERIFIED | Current `.schway` witness and JSON fixtures are consumed by the CLI tests; README and BINDINGS describe limits and commands. |

The phase plans predate the public rename. Their historical `cmd/lang`, `.lang`, and `session_app.go` paths are not present; the corresponding current artifacts above are wired into `cmd/schway`, `.schway`, and `session.go`. This is confirmed by direct call-site inspection and named tests, rather than treating path absence alone as failure.

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/schway/main.go` | `session/session.go` | `build` dispatch → `BuildApplicationFile` | ✓ WIRED | Direct call enters source/core checks, entry validation, application emission, and native build. |
| `session/session.go` | `cgen/cgen.go` | checked core → `EmitApplication` | ✓ WIRED | Direct call; emitter shares checked body lowering. |
| `cmd/schway/main.go` | `native/native_app.go` | ordinary `app run` → `RunApplication` / `RunApplicationWithEvidence` | ✓ WIRED | Both use the shared one-child runner; evidence capture does not launch another child. |
| `cmd/schway/main.go` | `native/bindings.go` | `--manifest` → session resolution | ✓ WIRED | Session resolves the manifest; compiler inputs and receipt identity consume the resolved snapshots. |
| generated application shell | `native/native_app.go` | explicit private event path | ✓ WIRED | Parent selects private capture only when events are requested; ordinary streams remain direct. |
| `native/native_app.go` | evidence report | child wait → validation → sidecar write | ✓ WIRED | Report contains process/capture status and cannot claim verified semantics. |
| `cmd/schway/main.go` | `session_app_verify.go` | explicit `app verify` dispatch | ✓ WIRED | Ordinary `app run` does not enter replay; named separation test passed. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Identity application | argv → parsed U64 → checked body → stdout | Caller token parsed by generated shell; checked identity body returns the scalar | Yes | ✓ FLOWING |
| Local C build | manifest → source/header/ABI snapshots → commands/receipt | Confined file reads and content digests | Yes | ✓ FLOWING |
| App evidence | private capture bytes → validated execution → sidecar | Same child process writes only when explicit capture is enabled | Yes | ✓ FLOWING |
| Differential verification | case input and pinned expected value → interpreter/O0/O3 comparison | Checked-in case fixture and the checked source | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

All commands below ran on the current Darwin arm64 checkout; each named test passed.

| Behavior | Command | Result | Status |
|---|---|---|---|
| Relocation, zero build launches, one run launch | `go test ./internal/compiler/native -run '^TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce$' -count=1 -v` | PASS, 0.60s | ✓ PASS |
| Identity answers and malformed/oversized input | `go test ./cmd/schway -run '^TestPhase22IdentityApplicationBuildAndRunCLI$' -count=1 -v` | PASS, 7/42 and six refusal inputs | ✓ PASS |
| Streams and process outcomes | `go test ./internal/compiler/native -run '^TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes$' -count=1 -v` | PASS, stream/exit/signal/timeout/launch/receipt/transport controls | ✓ PASS |
| Evidence success and fail-closed states | `go test ./internal/compiler/native -run '^TestPhase22EvidenceDisabledCompleteAndStreamIsolation$' -count=1 -v`; `go test ./internal/compiler/native -run '^TestPhase22EvidenceMissingPartialAndCapacityControls$' -count=1 -v` | Both PASS | ✓ PASS |
| Differential answers and model boundaries | `go test ./cmd/schway -run '^TestPhase22AppVerifyIndependentIdentityCases$' -count=1 -v`; `go test ./cmd/schway -run '^TestPhase22AppVerifyModelOnlyOutcomes$' -count=1 -v` | Both PASS, including four malformed model controls | ✓ PASS |
| Ordinary run avoids replay | `go test ./cmd/schway -run '^TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute$' -count=1 -v` | PASS | ✓ PASS |
| Binding refusals and input identity changes | `go test ./internal/compiler/native -run '^TestPhase22BindingsRejectInvalidInputs$' -count=1 -v`; `go test ./internal/compiler/native -run '^TestPhase22BuildIdentityDeclaredInputMutations$' -count=1 -v` | Both PASS | ✓ PASS |
| Emitter body sharing | `go test ./internal/compiler/cgen -run '^TestPhase22ApplicationEmitterSharesBodyAndSeparatesOutputShell$' -count=1 -v` | PASS | ✓ PASS |

Commands used the sandbox-writable cache override `GOCACHE=/tmp/ai-lang-verification-gocache`. No full suite, CI run, or Linux run is claimed. The existing UAT receipt is retained as historical completed acceptance evidence and was not replayed.

### Probe Execution

No Phase 22 plan declares a probe, and this is not a migration/tooling phase. No probe was required or run.

### Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| APP-02 | 22-01, 22-02 | Retained native build, no build effects, relocation, declared runtime dependency | ✓ SATISFIED | Build/relocation/launch tests passed; build receipt and docs honestly mark host closure incomplete and non-cacheable. |
| APP-03 | 22-01 | Bounded caller input, independent 7/42 results, defined refusal | ✓ SATISFIED | Identity CLI test passed success and invalid input cases. |
| APP-04 | 22-01, 22-03 | Exactly one ordinary launch, replay only on explicit verify | ✓ SATISFIED | Launch-count and ordinary-route separation tests passed. |
| APP-05 | 22-01 | Defined stdout/stderr and process outcomes | ✓ SATISFIED | Named stream/outcome test passed. |
| APP-06 | 22-03 | Separate execution evidence and fail-closed statuses | ✓ SATISFIED | Named success and incomplete/capacity evidence tests passed. |
| FFI-02 | 22-02 | Declared local C inputs, relocation, compatibility failures and content identity | ✓ SATISFIED | Relocation, negative declarations, and identity mutation tests passed. |
| EVD-11 | 22-03 | Independent differential outcomes and bounded modeled-world claims | ✓ SATISFIED | Identity replay and model-only positive/negative tests passed; reports disclaim host IO and physical cleanup. |

All seven plan requirement IDs are mapped to Phase 22 in `REQUIREMENTS.md`; no additional Phase 22-mapped requirement is orphaned.

### Decision Coverage

The decision coverage query returned 7/7 trackable CONTEXT decisions honored, with no unhonored decisions. This gate is non-blocking.

### Test Quality Audit

| Test File | Linked Req | Active | Skipped | Circular | Assertion Level | Verdict |
|---|---|---:|---:|---|---|---|
| `cmd/schway/main_test.go` | APP-02–06, EVD-11 | Yes | 0 found | None found | Behavioral/value | PASS |
| `internal/compiler/native/bindings_test.go` | APP-02, FFI-02 | Yes | 0 found | None found | Behavioral/value | PASS |
| `internal/compiler/native/native_app_test.go` | APP-02, APP-04–06 | Yes | 0 found | None found | Behavioral/value | PASS |
| `internal/compiler/cgen/cgen_program_test.go` | APP-03, APP-06 | Yes | 0 found | None found | Behavioral/value | PASS |

The independent 7/42 expectations are checked-in literals; the model adapter's expected outcome is explicit test input. No disabled requirement-linked tests or system-generated expected-value files were found in the inspected test surface. No test-quality blocker was found.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No Phase 22 blocker or stub found | — | The grep matches for `return []string{}` are documented empty derivations for shapes the emitter admits; generated C `return NULL` is the expected protocol exit, not a static API response. Neither affects the Phase 22 truths. |

No unreferenced `TBD`, `FIXME`, or `XXX` debt markers, placeholder implementations, or unpopulated user-visible data paths were found in the current Phase 22 implementation and fixtures.

### Human Verification Required

None. The objective README contract UAT is complete at 1/1 and was preserved without replay. No remaining success criterion requires subjective or external human observation.

### Gaps Summary

No roadmap truth, plan must-have, requirement, artifact, key link, or data flow is missing. The current source and named behavior tests establish the five roadmap criteria. The retained build deliberately declares `platform-c-runtime` while reporting unknown host SDK/runtime closure as incomplete and non-cacheable; the phase does not claim complete toolchain provenance. The previous report's old hosted-CI source-revision claim has been removed because it was not current-checkout evidence.

---

_Verified: 2026-10-02T14:02:46Z_
_Verifier: the agent (gsd-verifier)_
