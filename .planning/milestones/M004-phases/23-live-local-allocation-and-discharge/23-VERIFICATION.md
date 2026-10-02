---
phase: 23-live-local-allocation-and-discharge
verified: 2026-10-02T14:51:56Z
status: passed
score: 11/11 must-haves verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-RESOURCE-DISCHARGE-CONTRACT.json
  - .planning/phases/23-live-local-allocation-and-discharge/23-VALIDATION.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-01-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-01-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-02-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-02-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-03-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-03-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-04-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-04-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-05-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-05-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-06-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-06-SUMMARY.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-07-PLAN.md
  - .planning/phases/23-live-local-allocation-and-discharge/23-07-SUMMARY.md
  - cmd/schway/main.go
  - cmd/schway/main_test.go
  - examples/phase23/README.md
  - examples/phase23/adapter.c
  - examples/phase23/adapter.h
  - examples/phase23/file_byte.bindings.json
  - examples/phase23/file_byte.expected.json
  - examples/phase23/file_byte.schway
  - internal/compiler/ability/ability.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_convention_absence_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/bindings.go
  - internal/compiler/native/native_app.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/native/phase23_observer_test.go
  - internal/compiler/native/testdata/phase23_observer.c
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_phase23_contract_test.go
  - internal/compiler/session/session_phase23_model_test.go
  - internal/compiler/session/session_phase6_test.go
  - internal/compiler/session/verification_groundedness_test.go
  - scripts/verify-phase23.sh
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase23/discard_owner.schway
covered_digest: "v2:sha256:64929304234ca066ccd5b063d55cb919cd82535a623162f678eb0ead1212ee0a"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 11/11 must-haves verified
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 23: Live Local Allocation and Discharge — Verification Report

**Phase Goal:** A developer can read a caller-selected file byte through a real allocation returned live to Schway and observe its generated local cleanup.
**Verified:** 2026-10-02T14:51:56Z
**Status:** passed
**Verification mode:** Initial-mode goal-backward verification. The prior report had no `gaps:` list, so I re-derived the five roadmap truths and merged the seven plans' additional must-haves instead of treating the prior PASS as evidence. The previous report's covered-input digest was stale after planning handoff edits; the Phase 23 goal and five success criteria are unchanged. The active `gsd_run query verification.fingerprint` produced the v2 digest above. This pass reran the Phase 23 focused evidence script and a named Phase 22 compatibility test; it did not replay plans or UAT.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | An opaque noncopyable resource receives bounded malloc storage, remains live through Schway-directed use, yields the independently expected byte, and is physically destroyed before exit. | ✓ VERIFIED | This verifier ran `sh scripts/verify-phase23.sh` at `7a8ed4748e824545025acbd711e540615954fee3`: observer tests passed for 0x41/0x42/0x43. Separate receipts require the same pointer through `malloc → use → free → outstanding=0`; hosted Ubuntu/macOS runs also passed at `692f791`. |
| 2 | Acquire, use, and infallible consuming release each use their own checked signature, operand/failure facts, and release pairing. | ✓ VERIFIED | Current focused gate passed `TestPhase23OperationABI` and the `TestPhase23OperationContractRefusalBeforeCSerialization` mutations. The three binding symbols each name their own prototype in `examples/phase23/file_byte.bindings.json` and `adapter.h`; independent peer mutation tests also passed. |
| 3 | Empty, maximum-size, oversized, malformed, and failed inputs obey the buffer/error contract; failed acquisition creates no owner and partial storage is cleaned. | ✓ VERIFIED | The current focused gate passed native partial-allocation fault injection and public cases for empty input, 1/4096-byte accepted bounds, 4097-byte and embedded-NUL rejection, empty/2-byte/non-regular/open-failed files, and 0x41/0x42. Failed adapter records have null owners and partial storage is freed. |
| 4 | Discarded owning success is rejected or immediately destroyed; independent acquisition-seeded validators reject missing, duplicate, wrong-resource, or fabricated cleanup even if every release is removed from candidate core. | ✓ VERIFIED | Current focused gate passed source/emitter discard refusals and all-release-deleted, duplicate, premature, wrong-resource, and fabricated-release mutations in core, origin, and path peers. Each peer derives the obligation from acquisition. |
| 5 | Normal completion and a real post-acquisition use/output failure clean each remaining owner once; independent physical observation rejects omitted or premature destruction despite plausible compiler events. | ✓ VERIFIED | Current focused gate passed public 0x41/0x42 normal runs, the real 0x43 typed-use failure, and all four reached physical controls. The observer is separately compiled and its receipt is checked independently of plausible semantic events. The 0x43 receipt puts release before the report boundary. |
| 6 | The bounded path token accepts 1–4096 bytes while the prior Phase 22 U64 application route remains available. | ✓ VERIFIED | `TestPhase23PublicFileByte` covers path bounds. A fresh named `TestPhase22IdentityApplicationBuildAndRunCLI` test passed on this checkout and checks the retained route prints `7`. |
| 7 | Unsupported copies, moved-from uses, escapes, cross-call transfers, and exits are refused before native serialization; discarded acquisition cannot hide an obligation. | ✓ VERIFIED | Current focused gate passed source refusal and emitter pre-serialization refusal tests for each form, including discard. Cross-call transfer remains refused in Phase 23 as declared. |
| 8 | The successor contract distinguishes consuming release, preserving borrow, and future transfer; interpreter/model evidence cannot claim host IO or physical cleanup. | ✓ VERIFIED | `23-RESOURCE-DISCHARGE-CONTRACT.json` declares transfer contract-only for Phase 24 and separates structural, model, and native evidence scopes. Current focused gate passed model-only acquire/use failure tests and contract mutation/schema tests. |
| 9 | Public documentation and independent expected answers cover 0x41/0x42, acquisition errors, and the 0x43 post-acquisition use error with scoped cleanup evidence. | ✓ VERIFIED | Current focused gate passed `TestPhase23ReadmeContract` and `TestPhase23ReadmeExpectedAnswersAreIndependentConstants`. `file_byte.expected.json` contains authored expected stdout, stderr, and exit codes. |
| 10 | One focused command is wired into both existing native CI host lanes and does not repeat full, race, vet, or sanitizer suites. | ✓ VERIFIED | `.github/workflows/ci.yml` calls `scripts/verify-phase23.sh` once in the `evidence-aggregate` Ubuntu/macOS matrix. The script runs five focused groups only. Hosted run [37005701631](https://github.com/szTheory/schway/actions/runs/37005701631) reports the Phase 23 step successful on both hosts at `692f791`. |
| 11 | Missing or skipped host evidence is kept incomplete rather than silently treated as a complete multi-host receipt. | ✓ VERIFIED | The script requires Clang and fails on a skipped or unmatched group; it prints `status=incomplete` before execution and reports `status=pass` only for the host it actually ran on. `23-VALIDATION.md` remains `in-progress` / `nyquist_compliant: false`, so its stale pending-host state is not a false pass. The newer hosted run above supplies both host receipts. |

**Score:** 11/11 must-haves verified (0 present, behavior-unverified).

### Seven-Plan Cross-Check

All seven plans and all seven summaries were read and mapped to the roadmap criteria plus plan-specific truths above. The plan artifacts/tests are current; historical `.lang` names in early plan snapshots reflect the later coordinated rename, while the implementation and current README use `.schway`.

| Plan | Current evidence cross-check | Status |
|---|---|---|
| 23-01 | Public retained app, 0x41/0x42 values, separate operation contracts, C prototype/layout checks | ✓ VERIFIED |
| 23-02 | Acquire-seeded core/origin/path obligations and reached lifecycle/contract mutations | ✓ VERIFIED |
| 23-03 | Discard/copy/moved-from/escape/call/exit refusals before C emission | ✓ VERIFIED |
| 23-04 | Bounded input, typed failures, no owner on failure, descriptor and partial-allocation cleanup | ✓ VERIFIED |
| 23-05 | Model-only outcomes and release/borrow/transfer contract | ✓ VERIFIED |
| 23-06 | Independent physical observer, use-error cleanup, and four reached destruction controls | ✓ VERIFIED |
| 23-07 | Public instructions/expected answers, focused script, and two-host CI wiring/receipts | ✓ VERIFIED |

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase23/file_byte.schway` | Public checked owner path | ✓ VERIFIED | Source acquires a `FileByteOwner`, borrows it for use, and generated local cleanup consumes the owner. |
| `examples/phase23/adapter.c`, `adapter.h`, `file_byte.bindings.json` | Bounded adapter and explicit ABI | ✓ VERIFIED | Adapter reads one regular-file byte plus EOF, returns malloc-backed storage, and exposes separate acquire/use/release functions with target layout assertions. |
| `internal/compiler/{check,corevalidate,originvalidate,pathoracle,cgen}` | Admission, independent derivation, and sole emission | ✓ VERIFIED | Source and emitter refusals plus acquire-seeded peer mutations passed in the focused gate. |
| `internal/compiler/native/testdata/phase23_observer.c` and `phase23_observer_test.go` | Physical lifetime evidence | ✓ VERIFIED | Separate observer tracks pointer identity, order, outstanding allocations, and reached invalid destructor attempts. |
| `examples/phase23/README.md`, `file_byte.expected.json` | Reproducible public witness | ✓ VERIFIED | Current documentation contract test passes and expected answers are fixed constants. |
| `scripts/verify-phase23.sh`, `.github/workflows/ci.yml` | Focused native host gate | ✓ VERIFIED | Gate passes locally and in both hosted matrix jobs; no expensive suites are duplicated in the script. |
| `23-RESOURCE-DISCHARGE-CONTRACT.json` | Successor ownership/evidence contract | ✓ VERIFIED | Contract-only transfer is explicitly separated from Phase 23 local release behavior. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/schway` | retained native app runner | CLI build/run dispatch | WIRED | `TestPhase23PublicFileByte` exercises caller-selected paths and expected results. |
| `.schway` source | checked foreign operation facts | source checker → validated core | WIRED | Focused source, operation ABI, and peer groups pass. |
| checked operations | `adapter.c` functions | manifest symbols and per-symbol prototypes | WIRED | Build tests compile the declared prototypes and native public tests invoke them. |
| successful acquire | peer owner obligation | core/origin/path acquisition-derived replay | WIRED | Removing all candidate releases still fails all three independent peers. |
| emitted cleanup | actual physical release | adapter hooks → separately compiled observer | WIRED | Native receipts require same-pointer use and release before normal return or error reporting. |
| CI evidence matrix | focused Phase 23 script | one step per Ubuntu/macOS job | WIRED | Hosted run 37005701631 records successful Phase 23 steps on both hosts. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| File-byte result | `owner.data`, then byte result | caller-selected file → `read` → `malloc` allocation → borrowed adapter dereference → Schway U64 output | Yes; current native tests compare independent 65/66 answers | ✓ FLOWING |
| Cleanup obligation | owner pointer | successful acquisition → local owner → checked consume operation → adapter `free` | Yes; separate observer matches pointer identity and zero outstanding allocations | ✓ FLOWING |
| Interpreter outcome | modeled owner/value | deterministic supplied operation outcomes | Does not represent host file IO or physical cleanup | ✓ MODEL-ONLY (correctly scoped) |

### Behavioral Spot-Checks

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Current Phase 23 focused gate | `gsd_run run-with-timeout 30 -- env GOCACHE=/tmp/schway-phase-handoff-gocache sh scripts/verify-phase23.sh` | This verifier: exit 0, 14 seconds, Darwin/arm64, Go 1.24.0, Apple Clang 21.0.0 targeting `arm64-apple-darwin25.6.0`, revision `7a8ed4748e824545025acbd711e540615954fee3`; five groups, no skips | ✓ PASS |
| Phase 22 route preserved | `gsd_run run-with-timeout 10 -- env GOCACHE=/tmp/schway-phase-handoff-gocache go test ./cmd/schway -run '^TestPhase22IdentityApplicationBuildAndRunCLI$' -count=1` | This verifier: exit 0; named test exercises the retained route and expected output `7` | ✓ PASS |
| Uncached repository regression | `GOCACHE=/tmp/schway-phase-handoff-gocache go test -count=1 ./...` | Orchestrator-run on this checkout: exit 0; includes `cmd/schway-repair` and compiler/session packages. Separate from the focused Phase 23 script. | ✓ PASS (orchestrator receipt) |
| Current two-host focused gate | Hosted CI run [37005701631](https://github.com/szTheory/schway/actions/runs/37005701631), head `692f791051ba671c49c68fdd2073229feb51b090` | GitHub reports `scripts/verify-phase23.sh` successful in both Ubuntu and macOS evidence-aggregate jobs; all checks jobs also succeeded. Current checkout differs from that tested source only in planning/report artifacts. | ✓ PASS (hosted receipt) |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase23.sh` | `sh scripts/verify-phase23.sh` on current checkout and through hosted `evidence-aggregate` matrix | This verifier's Darwin run exited 0 in 14 seconds with all five groups passing and no skips; hosted Ubuntu and macOS steps both succeeded at `692f791` | PASS |

### Requirements Coverage

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| FFI-03 | 01, 02, 03, 05, 07 | Each foreign operation uses its own checked signature, operand/failure facts, and release pairing | ✓ SATISFIED | Operation ABI probes, per-operation refusal mutations, independent peers, and public native build. |
| RES-04 | 01, 02, 04, 05, 06, 07 | Real bounded allocation remains live and supplies caller-selected byte through Schway use | ✓ SATISFIED | Public results and independent pointer lifecycle observer; hosted/current focused gate. |
| RES-07 | 01, 04, 06, 07 | Failed acquisition creates no owner and adapter cleans partial allocation under published bounds | ✓ SATISFIED | Native fault injection and public boundary cases pass. |
| RES-08 | 03, 07 | Discarded owning acquisition is refused or immediately consumed | ✓ SATISFIED | Checker/emitter discard refusal tests pass. |
| RES-09 | 02, 03, 06, 07 | Independent validation rejects missing or invalid cleanup from acquisition-derived obligation | ✓ SATISFIED | Core/origin/path mutations and physical observer controls pass. |

All requirements assigned to Phase 23 are listed above; no additional requirement is mapped to this phase outside the plans.

### Test Quality Audit

| Test File | Linked requirement | Active | Skipped | Circular | Assertion level | Verdict |
|---|---|---:|---:|---|---|---|
| `cmd/schway/main_test.go` | RES-04, RES-07 | Yes | 0 in run | No | Public input/output and failure values | ✓ ADEQUATE |
| `internal/compiler/native/native_app_test.go` | FFI-03, RES-04, RES-07 | Yes | 0 in run | No | Prototype/layout and acquisition failure behavior | ✓ ADEQUATE |
| `internal/compiler/native/phase23_observer_test.go` | RES-04, RES-09 | Yes | 0 in run | No | Ordered pointer lifecycle and reached physical mutations | ✓ ADEQUATE |
| `internal/compiler/corevalidate/corevalidate_test.go`, `originvalidate_test.go`, `pathoracle_test.go` | FFI-03, RES-09 | Yes | 0 | No | Reached positive/negative ownership mutations | ✓ ADEQUATE |
| `internal/compiler/session/session_phase23_contract_test.go`, `session_phase23_model_test.go` | FFI-03, RES-04, RES-07 | Yes | 0 | No | Contract/schema mutations and model/native evidence separation | ✓ ADEQUATE |

The ABI unit test contains a Clang-unavailable `t.Skip` fallback, but the focused script requires Clang before running and fails on any `--- SKIP:` output. Current local and hosted runs did not skip it. Test input files are generated by the harness; expected public answers come from checked-in authored constants, not from the system under test. Disabled tests linked to these requirements: 0. Circular expected-value generation: 0. Insufficient assertions: 0.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `internal/compiler/check/check.go` | 4644 | Comment contains “placeholder field” | Info | It describes an existing layout field; no placeholder implementation or user-visible stub. |
| `internal/compiler/native/testdata/phase23_observer.c` | 55 | `return NULL` | Info | Real allocator-failure path in the observer, not an empty implementation. |
| `internal/compiler/cgen/cgen_program.go`; test files | 79, 471; fixture/scanner definitions | Empty collection / placeholder marker matches | Info | Comments document current no-resource/no-attribute derivations; other matches are test-only fixtures/scanner tokens and do not feed user-visible behavior. |

No unreferenced `TBD`, `FIXME`, or `XXX` markers, user-visible placeholders, empty handlers, or stub data paths were found in the Phase 23 implementation files. The empty collections returned by adjacent compiler helpers are typed structural values, not rendered defaults or absent Phase 23 data.

### Human Verification Required

None. This phase’s acceptance criteria are objective CLI, ABI, model-scope, peer-mutation, and native lifetime behaviors. Current automated evidence covers both requested host lanes. No conversational UAT was replayed.

### Gaps Summary

No Phase 23 goal or plan must-have remains open. The current checkout passed the local focused gate and uncached repository suite; hosted run 37005701631 passed the Phase 23 aggregate on Ubuntu and macOS. `23-VALIDATION.md` still says the hosted Ubuntu receipt is pending; that note predates the hosted run and is stale. This verifier changed no implementation, roadmap, state, requirements, other phase report, plan, or summary.

### Decision Coverage

`gsd_run query check.decision-coverage-verify .planning/phases/23-live-local-allocation-and-discharge .planning/phases/23-live-local-allocation-and-discharge/23-CONTEXT.md` reports all 6 trackable context decisions honored, with none unhonored. This non-blocking check does not affect status.

---

_Verified: 2026-10-02T14:51:56Z_
_Verifier: the agent (gsd-verifier)_
