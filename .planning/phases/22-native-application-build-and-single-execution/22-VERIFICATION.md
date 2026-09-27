---
phase: 22-native-application-build-and-single-execution
verified: 2026-09-27T21:13:10Z
status: human_needed
score: 17/17 must-haves verified
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
  - internal/compiler/session/session_app_verify_test.go
covered_digest: "v1:sha256:fcc3fb354fc04434004590406518e41b743630e6fbd4c25721298873d8da8beb"
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Review the Phase 22 README as a developer following the documented build, run, evidence, replay, and local C workflows."
    expected: "A developer can understand the command sequence, input and report bounds, failure states, local C trust boundary, and limits of modeled outcomes without ambiguous or misleading wording."
    why_human: "Plan 22-03 records this documentation-clarity review as human judgment; implementation accuracy is checked in code, but clarity and usability require a reader. No Phase 22 UAT artifact records this review."
---

# Phase 22: Native Application Build and Single Execution Verification Report

**Phase Goal:** A developer can retain and execute a native application once on bounded real input, with application streams separate from compiler evidence.
**Verified:** 2026-09-27T21:13:10Z
**Status:** human_needed
**Re-verification:** No — the previous report had no `gaps:` section, so the verification workflow's initial-mode must-have rule applies. All truths were re-established against the current tree after the dated summary correction.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | From a relocated checkout, a developer can build a retained executable with declared local C sources, headers, symbols, ABI inputs, and runtime dependencies. Build has no application effects; invalid inputs fail clearly and relevant changes invalidate identity. | ✓ VERIFIED | `native.BuildApplication` resolves the manifest, compiles/links, and publishes retained artifact plus receipt; `TestPhase22RelocatedBindingsCLI` exercises a relocated checkout, `TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce` checks build launches zero app processes and run launches one, and bindings tests cover invalid paths/types/symbols and declared-input identity mutations. The receipt declares `platform-c-runtime` while correctly marking unknown host closure incomplete and non-cacheable. |
| 2 | Caller inputs `7` and `42` produce independently specified identity results; malformed or oversized input has a defined outcome. | ✓ VERIFIED | `identity.lang` returns its U64 input and `identity.expected.json` independently pins both values. `TestPhase22IdentityApplicationBuildAndRunCLI` exercises public build/run and malformed, signed, nonnumeric, overlong, and overflowing values; generated entry parsing bounds/rejects before entering the Lang body. |
| 3 | Each public app-run request launches the selected artifact once; streams and process outcomes are defined without execution-JSON parsing or hidden replay. | ✓ VERIFIED | CLI dispatch calls `native.RunApplication` or `RunApplicationWithEvidence`, converging on one `runApplication` child `Run`. `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, `TestPhase22ConcurrentRequestsLaunchIndependently`, `TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus`, and `TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute` exercise byte streams, child outcomes, independent launches, and replay separation. |
| 4 | Requested execution evidence is separate from app streams and distinguishes disabled, incomplete, and capacity-exhausted capture from verification success. | ✓ VERIFIED | `RunApplicationWithEvidence` captures through a private bounded file, validates after the child wait, and publishes a separate status-bearing report with `verified:false`. `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, `TestPhase22EvidenceMissingPartialAndCapacityControls`, and write-failure controls exercise the states and stream isolation. |
| 5 | Explicit replay uses independent expected answers and isolates declared model outcomes from actual host IO; ordinary app runs do not replay. | ✓ VERIFIED | `VerifyApplicationCasesFile` and explicit CLI `app verify` dispatch compare isolated source cases against independent expected values and interpreter/O0/O3 results. `TestPhase22AppVerifyIndependentIdentityCases`, wrong-answer and verifier-model controls, duplicate-key test, local-C/script refusal test, and ordinary-run routing test cover positive and negative behavior. Reports label modeled outcomes and set host-IO/physical-cleanup claims false. |
| 6 | Build publishes a retained executable without starting it; the executable remains usable after temporary build cleanup. | ✓ VERIFIED | `TestPhase22BuildRetainsRelocatableArtifactWithoutLaunching` and `TestPhase22BindingsBuildNeverLaunchesAndRunLaunchesOnce` exercise retention, relocation, cleanup, zero build launches, and successful run. |
| 7 | The same checked source returns exactly `7\n` and `42\n`; malformed and oversized input is refused before Lang body effects. | ✓ VERIFIED | Public CLI integration test uses pinned identity source; emitter and CLI tests cover canonical bounded U64 parsing and refusal cases. |
| 8 | An ordinary app request starts exactly one selected child, including stderr and unsuccessful exit cases. | ✓ VERIFIED | Named native stream/process and launch-counter tests passed. |
| 9 | Ordinary streams remain bytes, and exit, signal, timeout, and launch errors are distinguishable. | ✓ VERIFIED | `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes` plus CLI opaque-token/child-status test passed. |
| 10 | A relocated checkout builds with explicit local C source, header, symbol, ABI function type, and runtime dependency declaration. | ✓ VERIFIED | `identity.bindings.json`, `support.h`, `support.c`, and `BINDINGS.md` provide the declaration and contract; `TestPhase22RelocatedBindingsCLI` passes. |
| 11 | Declared local C inputs build without launching the app; retained artifact runs after temporary files are removed. | ✓ VERIFIED | Native bindings build/run counter test passed; native retained artifact test also exercises relocation and cleanup. |
| 12 | Missing/out-of-root inputs, incompatible header types, and unresolved symbols fail by declared input before publication. | ✓ VERIFIED | `TestPhase22BindingsRejectInvalidInputs`, closed-manifest/path controls, and actual compile/link controls passed. |
| 13 | Changes to Lang/C/header/manifest ABI or symbols, flags, compiler, target, and runtime declarations affect identity or are explicitly non-cacheable. | ✓ VERIFIED | `TestPhase22BuildIdentityDeclaredInputMutations` exercises the mutation matrix. Runtime closure is explicitly incomplete and receipts/reports are non-cacheable where host facts are unknown. |
| 14 | Requested app evidence is a separate sidecar with build/input identity and closed capture states; disabled, incomplete, and exhausted never mean verified success. | ✓ VERIFIED | Evidence tests exercise disabled/complete/incomplete/capacity-exhausted and publication failures; report implementation retains `verified:false`. |
| 15 | Same-run event capture does not create another app launch or invoke interpreter/O0/O3 replay. | ✓ VERIFIED | The one-child evidence integration test and `TestPhase22OrdinaryAppRunDoesNotEnterReplayRoute` passed; capture writes to a process-private channel. |
| 16 | Explicit identity replay checks isolated inputs against independent 7/42 results and interpreter/O0/O3 outcomes, with a separate verifier-only modeled outcome. | ✓ VERIFIED | Independent fixture and named CLI replay/model tests passed; ordinary source cases have no scripted foreign outcomes and local C is refused on replay route. |
| 17 | Missing, duplicate, unconsumed, or mismatched model outcomes fail; model success and process reclamation do not claim host IO or physical cleanup. | ✓ VERIFIED | Model negative controls and `TestDecodeReplayCasesRejectsDuplicateJSONKeys` passed; report labels/false claim fields are inspected by tests and documented in README. |

**Score:** 17/17 truths verified (0 present, behavior-unverified). The goal's automated and code-verifiable truths pass; a separate human documentation-clarity review remains, so the phase status is `human_needed`.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase22/identity.lang`, `identity.expected.json` | Checked scalar program and independent expected pairs | ✓ VERIFIED | Substantive source and literal expected pairs; consumed by CLI and replay checks. |
| `internal/compiler/session/session_app.go` (planned path) | Checked build path | ✓ VERIFIED (implementation-path substitution) | This exact file is absent. Current implementation is in `internal/compiler/session/session.go`: `BuildApplication` performs check, independent core validation, entry resolution, and `cgen.EmitApplication`; public CLI calls `BuildApplicationFile`. This matches the dated correction and is not a missing behavior. |
| `internal/compiler/native/native_app.go` | Retained build, receipt, one-child runner, evidence | ✓ VERIFIED | Substantive build/publish/integrity/run/evidence implementation. Public CLI calls the runner; named integration tests passed. |
| `internal/compiler/native/bindings.go` | Closed local-C manifest resolver | ✓ VERIFIED | Strict manifest/path/header/ABI resolution, content inventory and identity; wired through public build dispatch and native build. |
| `examples/phase22/identity.bindings.json`, `support.h`, `support.c`, `BINDINGS.md` | Declared relocatable C/ABI inputs and contract | ✓ VERIFIED | All exist with substantive declarations; actual build tests resolve and link the fixture, whose C function is intentionally not called by the Lang app. |
| `internal/compiler/session/session_app_verify.go` | Explicit replay service | ✓ VERIFIED | Closed case parsing and explicit interpreter/O0/O3 replay; called only by `lang app verify`. |
| `examples/phase22/identity.cases.json`, `examples/phase22/README.md` | Independent replay fixture and public contract | ✓ VERIFIED | Cases carry pinned expected answers and empty foreign scripts; README describes commands, limits, evidence states, C trust boundary, and modeled-world limits. Clarity remains subject to human review. |
| `cmd/lang/main_test.go`, `native_app_test.go`, `bindings_test.go`, `cgen_program_test.go`, `session_app_verify_test.go` | Reached positive and negative controls | ✓ VERIFIED | Relevant named tests passed in this verification run. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/lang/main.go` | `session/session.go` | `build` dispatch → `BuildApplicationFile` → `BuildApplication` | ✓ WIRED | Manual symbol trace confirms public CLI reaches Check/core validation/emission. Exact planned `session_app.go` target was corrected to existing `session.go`. |
| `session/session.go` | `cgen/cgen.go`, `cgen/cgen_program.go` | `BuildApplication` → `EmitApplication` → shared application emitter/body lowering | ✓ WIRED | Source calls and shared emitter implementation verified. |
| `cmd/lang/main.go` | `native/native_app.go` | app run dispatch → single retained runner | ✓ WIRED | Both ordinary and evidence modes call the same one-child run path. |
| `cmd/lang/main.go` | `native/bindings.go` | `--manifest` build option | ✓ WIRED | Manifest argument reaches session/native build and `ResolveBindings`. |
| `bindings.go` | `native_app.go` / build receipt | resolved files and digests | ✓ WIRED | Resolved inventory feeds compilation/link input and receipt identity. |
| `cgen_program.go` | `native_app.go` | private event capture | ✓ WIRED | Generated app selects private capture path only when parent supplies it; ordinary output remains separate. |
| `native_app.go` | evidence report | validate after child wait and atomic publication | ✓ WIRED | Status/report construction follows the single child wait and carries build/input/process identity. |
| `cmd/lang/main.go` | `session_app_verify.go` | explicit `app verify` dispatch | ✓ WIRED | CLI calls `VerifyApplicationCasesFile`; ordinary app run dispatch does not. |

The generic `verify.key-links` heuristic reports false negatives for indirect Go package references and the intentional planned-path substitution. Manual call tracing and passing integration tests resolve these links; no missing link was inferred from the heuristic alone.

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Identity application | argv token → parsed U64 → Lang result → stdout | Caller token is parsed by generated entry shell and passed to checked entry; result encoded as decimal bytes | Yes | ✓ FLOWING |
| Local C build | manifest paths → source/header bytes and ABI probe inputs | Root-confined local files are read, hashed, compiled and linked | Yes | ✓ FLOWING |
| App evidence | private capture file → validated execution document → sidecar | Same child writes bounded events; parent validates schema/status and publishes separate report | Yes | ✓ FLOWING |
| Replay report | fixture inputs/expected values → interpreter/O0/O3 outcomes | Case file plus independent literals feed explicit replay and comparisons | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Public identity, malformed U64, streams/process outcomes, local-C relocation/identity, evidence/replay controls | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/lang ./internal/compiler/native ./internal/compiler/cgen -run '^TestPhase22' -count=1` | Exit 0; cmd/lang, native, and cgen passed (0.822s, 8.420s, 0.011s). | ✓ PASS |
| Duplicate replay JSON keys rejected | `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestDecodeReplayCasesRejectsDuplicateJSONKeys$' -count=1` | Exit 0; named test passed (0.010s). | ✓ PASS |

No server or external service was started. Checks ran on macOS; Linux host evidence is unavailable and not claimed.

### Probe Execution

No declared probes or conventional `scripts/*/tests/probe-*.sh` files were found for this phase.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| APP-02 | 22-01, 22-02 | Retained documented build, no build effects, relocated executable and declared dependencies | ✓ SATISFIED | Relocation and launch-counter tests; receipt declares `platform-c-runtime`, with incomplete host closure/non-cacheable state documented. |
| APP-03 | 22-01 | Bounded caller input and independent scalar answers; malformed/oversize refusal | ✓ SATISFIED | Public identity CLI controls, expected fixture, and input refusal tests. |
| APP-04 | 22-01, 22-03 | Exactly one ordinary application launch; replay explicit | ✓ SATISFIED | Native launch counter, stream/process controls and ordinary-run replay exclusion. |
| APP-05 | 22-01 | Byte streams and defined exit/signal/timeout/launch behavior | ✓ SATISFIED | Named native and CLI stream/process tests passed. |
| APP-06 | 22-03 | Separate evidence with fail-closed capture states | ✓ SATISFIED | Disabled/complete/incomplete/capacity and publication-failure controls passed; evidence cannot claim verification. |
| FFI-02 | 22-02 | Declared local C/header/symbol/ABI/runtime inputs, relocation, errors and identity invalidation | ✓ SATISFIED | Manifest positive/negative, relocation, mutation-matrix and receipt checks passed. |
| EVD-11 | 22-03 | Independent replay, declared modeled outcomes, no host IO/cleanup overclaim | ✓ SATISFIED | Replay fixture, expected-value and model controls passed; reports bound the claim scope. |

`REQUIREMENTS.md` checkboxes and its Phase 22 traceability table mark these as complete. Its introductory sentence says “All requirements below are new pending work,” which conflicts with the completion checkboxes/table; this report relies on current implementation evidence rather than that stale sentence.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unreferenced TBD/FIXME/XXX debt marker, stub, or placeholder was found in the covered implementation and example files. | — | — |

The plan prohibitions are also respected: linked local C is not represented as proof of arbitrary C behavior; complete same-run capture is not represented as semantic verification; verifier-only modeled outcomes claim neither actual host IO nor physical cleanup.

### Human Verification Required

#### 1. Public README clarity

**Test:** Review `examples/phase22/README.md` as a developer following its build, run, evidence, replay, and local C workflows.
**Expected:** The command sequence, bounds, failure states, local C trust boundary, and modeled-outcome limits are clear and not misleading.
**Why human:** Plan 22-03 explicitly marks this README clarity check as human judgment. Its implementation accuracy is verifiable against source, but reader clarity cannot be established by source inspection or tests. No Phase 22 UAT artifact records this review.

### Gaps Summary

No automated truth, artifact behavior, key link, or requirement is missing. The former planned `session_app.go` artifact path is absent by the dated correction; the checked build entry and CLI wiring are present in `session.go`, and the replay service remains in `session_app_verify.go`. The report therefore retains a full automated score, while status remains `human_needed` until the explicit documentation-clarity check is recorded. Host SDK/linker/runtime closure remains incomplete and non-cacheable as declared; no Linux-host result is claimed.

---

_Verified: 2026-09-27T21:13:10Z_  
_Verifier: the agent (gsd-verifier)_
