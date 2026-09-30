---
phase: 22-native-application-build-and-single-execution
verified: 2026-09-30T15:10:24Z
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
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_app_verify_test.go
covered_digest: "v1:sha256:f3f29f345cbc433d1d8426f7988af0815a27e32bf25f2116c862ae86b72060d3"
behavior_unverified: 0
overrides_applied: 0
decision_coverage:
  honored: 7
  total: 7
  not_honored: []
---

# Phase 22: Native Application Build and Single Execution — Verification Report

**Phase Goal:** A developer can retain and execute a native application once on bounded real input, with application streams separate from compiler evidence.
**Verified:** 2026-09-30T15:10:24Z
**Status:** passed
**Re-verification:** No. The existing report had no `gaps:` section, so this is an initial-mode refresh.

This refresh verifies the five current ROADMAP success criteria and cross-checks every plan must-have and prohibition. SUMMARY files were read for scope discovery only; they are not evidence for the verdict. No local tests were run. Current behavior-dependent evidence comes from hosted CI run `36707529870` at head `f991298b29779838a2b1a5c3cd5ac90aafcb84fc`, which passed the Linux and macOS checks and current evidence aggregate jobs. The only commit after that CI head changes `22-REVIEW.md`; the covered implementation and test files are unchanged. The optional validation-corpus receipt was skipped and is not claimed. CI workflow steps include `go vet ./...`, `go build ./...`, `go test ./...`, and `go test -race ./...` on both hosts.

The completed objective README contract UAT in `22-UAT.md` is preserved and not repeated. Its objective contract test is covered by the hosted full test runs; it does not claim subjective readability.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | From a relocated checkout, a developer can build a retained executable with declared local C sources, headers, symbols, ABI inputs, and runtime dependencies. Building performs no application effects; missing/incompatible inputs fail clearly and relevant input changes invalidate artifact/evidence identity. | ✓ VERIFIED | `session.BuildApplication` validates before `cgen.EmitApplication`; `native.BuildApplication` compiles to a staged path and publishes an adjacent receipt without running the output. `ResolveBindings` confines inputs and checks declarations; identity includes manifest, compiler, target, flags, source, C, header and ABI inputs. Relocation, refusal, mutation, and zero-build-launch tests are in the hosted Linux/macOS suite. |
| 2 | An ordinary scalar source program run with caller-selected inputs 7 and 42 produces independently specified corresponding results; malformed and oversized input has a defined outcome. | ✓ VERIFIED | `identity.schway` and `identity.expected.json` pin the witness/answers. The generated entry parses and bounds the opaque token before calling the checked body. Identity CLI, malformed-input, and unsupported-entry controls are in the hosted full test suite. |
| 3 | One public application-run request launches the selected artifact exactly once, with defined stdout, stderr, exit, and signal behavior; ordinary output and successful stderr are accepted without execution-JSON parsing or hidden interpretation/tier replay. | ✓ VERIFIED | Public `app run` calls `native.RunApplication`; evidence mode converges on the same runner. The runner verifies receipt/digest, creates one `exec.CommandContext`, assigns direct stream writers and waits once. Hosted tests cover opaque streams, nonzero/signal/timeout/launch outcomes, concurrency, evidence launch count, and ordinary-route replay exclusion. |
| 4 | A developer can request execution evidence through a separate channel and distinguish disabled, incomplete, and capacity-exhausted states from successful verification. | ✓ VERIFIED | `RunApplicationWithEvidence` captures to a private file from the same child, validates after wait, and atomically writes a separate sidecar with `Verified: false`; incomplete/capacity cases return errors. Capture isolation, failure and capacity tests are covered by hosted CI. |
| 5 | Explicit differential verification uses isolated or replayable inputs and independent expected answers; declared foreign outcomes are distinguished from actual host IO, and verification replay does not occur on the ordinary application route. | ✓ VERIFIED | `VerifyApplicationCases` compares interpreter and generated C O0/O3 outcomes against literal answers. Model cases use a separate adapter and report host-IO/physical-cleanup claims false; source replay refuses foreign scripts/local C. Hosted tests cover positive identity/model cases, wrong expected answers, malformed/missing/duplicate/unconsumed/mismatched outcomes, and ordinary-run separation. |

**Score:** 5/5 roadmap truths verified (0 present, behavior-unverified).

### Plan Must-Have Crosswalk

| Plan | Must-have / prohibition | Status | Evidence |
|---|---|---|---|
| 22-01 | Retained build, no build launch, relocation after cleanup; bounded identity inputs and refusal; one launch; byte streams and distinct outcomes | ✓ VERIFIED | Covered by roadmap truths 1–3; native/CLI/emitter controls named above and in hosted full suite. |
| 22-02 | Relocatable declared C/header/symbol/ABI/runtime inputs; fail before publication for bad paths/types/symbols; relevant mutations change identity or remain explicitly non-cacheable | ✓ VERIFIED | Covered by truth 1; resolver, compiler ABI probe, receipt construction and named binding tests. |
| 22-02 | Do not claim arbitrary local C's behavior or ownership is proven by its declaration | ✓ VERIFIED | `BINDINGS.md` defines trusted-C boundary; resolver only checks declared source/header/type/symbol compatibility. |
| 22-03 | Separate evidence status; capture does not create another run or replay; explicit isolated differential with model-only case; fail closed and do not claim host IO/cleanup | ✓ VERIFIED | Covered by truths 3–5; source paths and CI-covered capture/replay positive and negative tests. |
| 22-03 | Complete capture is not an independent semantic verdict; modeled foreign success is not source FFI, host IO, or physical cleanup | ✓ VERIFIED | Evidence report sets `verified:false`; model scope and claim flags are distinct and tested. |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `cmd/schway/main.go` | Public build, run, verify dispatch | ✓ VERIFIED | Substantive dispatch and handlers; build reaches session build, run reaches native one-child path, verify alone reaches replay. |
| `internal/compiler/session/session.go` | Checked build path | ✓ VERIFIED | `BuildApplication` performs check/core validation, resolves entry and emits the checked application shell. Planned `session_app.go` was consolidated into this existing file; public call site confirms wiring. |
| `internal/compiler/cgen/cgen.go`, `cgen_program.go` | Shared body lowering with application entry/output shell | ✓ VERIFIED | `EmitApplication` validates and delegates to shared `emitProgramWithShell`; separate shell preserves application output boundary. |
| `internal/compiler/native/native_app.go` | Retained build receipt and single-run/evidence path | ✓ VERIFIED | Staging, compiler invocation, publish/receipt validation, single child, direct streams and sidecar capture are implemented and wired. |
| `internal/compiler/native/bindings.go` | Closed local-C manifest resolution and identity | ✓ VERIFIED | JSON schema, path confinement, header/type/symbol checks, snapshots and content identity are implemented. |
| `internal/compiler/session/session_app_verify.go` | Explicit differential/model verifier | ✓ VERIFIED | Wired only through `app verify`; bounded inputs, independent outcomes and report scopes are substantive. |
| Phase 22 examples and docs | Runnable witness, independent cases, local-C inputs, public contract | ✓ VERIFIED | `.schway` identity witness, expected/case JSON, C/header/manifest and README/BINDINGS docs are present and consumed by tests and CLI paths. |

The previous report's planned `internal/compiler/session/session_app.go` path is absent by design; implementation resides in `session.go`. The old pre-rename names in the 2026-09-28 report are historical and were checked against the current `schway` command and `.schway` source.

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/schway/main.go` | `session/session.go` | `app build` dispatch → `BuildApplicationFile` | ✓ WIRED | Direct call; checked source path reaches validation/emission then native build. |
| `session/session.go` | `cgen/cgen.go` | checked program → `EmitApplication` | ✓ WIRED | Direct call; application emitter shares function lowering in `cgen_program.go`. |
| `cmd/schway/main.go` | `native/native_app.go` | `app run` → `RunApplication` or `RunApplicationWithEvidence` | ✓ WIRED | Both reach shared `runApplication`; there is no second ordinary run path. |
| `cmd/schway/main.go` | `native/bindings.go` | `--manifest` through session to `ResolveBindings` | ✓ WIRED | Session resolves manifest; native compilation and build identity consume resolved files and digests. |
| generated application shell | `native/native_app.go` | private event path when evidence explicitly requested | ✓ WIRED | Runner controls capture environment; ordinary run removes ambient authorization. |
| `native/native_app.go` | evidence report | wait → validate → atomic sidecar | ✓ WIRED | Report is distinct from stdout/stderr and incomplete/exhausted states fail closed. |
| `cmd/schway/main.go` | `session_app_verify.go` | explicit `app verify` dispatch | ✓ WIRED | Ordinary `app run` does not dispatch to replay. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Identity application | argv → parsed U64 → checked body → stdout | Caller token parsed in generated app shell; result emitted from the checked program | Yes | ✓ FLOWING |
| Local C build | manifest → C/header/type/symbol input snapshots | Confined filesystem reads; digests feed compile/link commands and receipt | Yes | ✓ FLOWING |
| App evidence | private capture bytes → validated event data → sidecar | Same child process writes only when explicit capture is enabled; parent validates afterward | Yes | ✓ FLOWING |
| Differential verification | isolated case input and pinned expected value → interpreter/O0/O3 comparison | Checked-in case fixture; local C/foreign scripts rejected on this route | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Build, app behavior, evidence, verifier controls and race-sensitive code on Ubuntu | Hosted GitHub Actions run `36707529870`, `checks (ubuntu-latest)` and `current evidence aggregate (ubuntu-latest)` at `f991298b29779838a2b1a5c3cd5ac90aafcb84fc` | Both jobs passed. Workflow runs `go vet ./...`, `go test ./...`, and `go test -race ./...` in the checks job. | ✓ PASS |
| Same suite and evidence aggregate on macOS | Same run, `checks (macos-latest)` and `current evidence aggregate (macos-latest)` | Both jobs passed. Workflow runs vet, full Go tests, and race tests in the checks job. | ✓ PASS |
| Current source correspondence | `git diff --name-status f991298b29779838a2b1a5c3cd5ac90aafcb84fc..HEAD` | Only `.planning/phases/22-native-application-build-and-single-execution/22-REVIEW.md` differs; no implementation/test files changed after CI head. | ✓ PASS |
| Optional validation-corpus receipt | Same hosted run | Skipped; no validation-corpus result is claimed. | ℹ️ SKIPPED |

No local tests or probes were executed in this refresh, per instruction. This refresh distinguishes the hosted receipt above from historical local test receipts recorded in the SUMMARY/UAT artifacts.

### Probe Execution

No phase-declared or conventional `scripts/*/tests/probe-*.sh` probes were found or implied by the Phase 22 plans. This is an application feature phase, not a migration/tooling phase.

### Requirements Coverage

| Requirement | Source plan | Description | Status | Evidence |
|---|---|---|---|---|
| APP-02 | 22-01, 22-02 | Retained native build, no build effects, relocation and declared runtime dependencies | ✓ SATISFIED | Build path, receipt, binding controls; current hosted Linux/macOS checks. |
| APP-03 | 22-01 | Bounded caller input, independent 7/42 outcomes and refusal | ✓ SATISFIED | Generated U64 entry and active CLI/emitter tests in hosted full suite. |
| APP-04 | 22-01, 22-03 | Exactly one ordinary app launch, with replay isolated to explicit verify | ✓ SATISFIED | Shared single-child runner and route separation controls in hosted suite. |
| APP-05 | 22-01 | Ordinary stdout/stderr and defined process outcomes | ✓ SATISFIED | Direct writer use and named stream/outcome controls in hosted suite. |
| APP-06 | 22-03 | Separate compiler evidence with fail-closed status states | ✓ SATISFIED | Evidence report code and capture failure/capacity controls in hosted suite. |
| FFI-02 | 22-02 | Declared local C inputs, relocation, compatibility failures and content identity | ✓ SATISFIED | Root-constrained resolver, ABI probe, identity tests and CI. |
| EVD-11 | 22-03 | Independent differential results and bounded modeled-world claims | ✓ SATISFIED | Replay service and positive/negative controls in hosted full suite. |

All seven plan requirement IDs are mapped to Phase 22 in `REQUIREMENTS.md`; no additional Phase 22-mapped requirement is orphaned.

### Decision Coverage

The current decision-coverage gate returned 7/7 trackable CONTEXT decisions honored, with no unhonored decisions. This is a non-blocking gate.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `cmd/schway/main.go` | 664 | Generic usage diagnostic still begins `usage: lang` after the public rename to `schway` | ⚠️ WARNING | The partial Phase 22 review's WR-01 is confirmed. This affects an invalid-command diagnostic, not the documented `schway app build/run` paths or their behavior; it does not fail a Phase 22 must-have. Follow-up: update the prefix and its contract assertion. |
| `scripts/verify-phase23.sh` | 34–38 | Review WR-02: clean-tree receipt checks tracked diffs but not untracked files | ⚠️ WARNING | This is outside Phase 22's implementation scope and does not affect Phase 22 goal evidence. Phase 23 evidence tooling should include untracked files before claiming a clean tree. |

No unreferenced `TBD`, `FIXME`, or `XXX` markers, placeholder implementations, or static-return stubs were found in the covered Phase 22 implementation and example files. The latest review is partial (5 of 443 manifest paths); its warnings are retained above and are not treated as a clean full-scope review.

### Human Verification Required

None. The requested objective README contract UAT is already complete and preserved in `22-UAT.md`; no genuinely unverifiable roadmap truth remains.

### Gaps Summary

No goal, roadmap success criterion, plan must-have, requirement, artifact, key link, or data flow is missing. The old report's evidence and paths were refreshed against the current Schway rename and current source. CI provides current cross-host behavior/regression evidence for the implementation files; the later commit changed only the review document. The skipped optional validation corpus is outside this phase's must-haves. Two warnings from the partial review remain actionable but do not block Phase 22's demonstrated goal.

---

_Verified: 2026-09-30T15:10:24Z_  
_Verifier: the agent (gsd-verifier)_
