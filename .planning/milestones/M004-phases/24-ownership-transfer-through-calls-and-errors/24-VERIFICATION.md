---
phase: 24-ownership-transfer-through-calls-and-errors
verified: 2026-10-02T14:20:59Z
status: passed
score: 5/5 roadmap success criteria verified
covered_files:
  - .github/workflows/ci.yml
  - .planning/LANGUAGE-MATURITY.md
  - .planning/REQUIREMENTS.md
  - .planning/ROADMAP.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-01-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-01-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-02-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-02-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-PLAN.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-03-SUMMARY.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-CONTEXT.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-REVIEW.md
  - .planning/phases/24-ownership-transfer-through-calls-and-errors/24-VALIDATION.md
  - examples/phase24/README.md
  - examples/phase24/error.schway
  - examples/phase24/transfer.schway
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/interp/interp.go
  - internal/compiler/interp/interp_test.go
  - internal/compiler/native/native_app_test.go
  - internal/compiler/native/phase24_observer_test.go
  - internal/compiler/native/testdata/phase24_observer.c
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/session/session_phase24_error_test.go
  - internal/compiler/session/session_phase24_model_test.go
  - internal/compiler/session/session_phase24_transfer_test.go
  - scripts/verify-phase24.sh
  - testdata/phase16/public-emitter-consumers.json
covered_digest: "v2:sha256:7ccab4c18f0addd52c537560b887cc962f2e34d6e5194a5e89f3cadc0c33b7df"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 5/5
  gaps_closed: []
  gaps_remaining: []
  regressions: []
---

# Phase 24: Ownership Transfer Through Calls and Errors — Verification Report

**Phase Goal:** A developer can transfer a live resource through Schway calls and returns, use it under its new owner, and rely on exactly-once cleanup on admitted normal and typed-error paths.
**Verified:** 2026-10-02T14:20:59Z
**Status:** passed
**Re-verification:** Yes — current-checkout refresh; the prior report had no gaps, so the roadmap contract was re-established and all truths were checked.

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | An allocation acquired in one frame remains usable after ownership transfer through a Schway call and return; transfer invokes no destructor, borrow preserves ownership, and the final owner invokes the declared consuming destructor exactly once. | ✓ VERIFIED | Source witnesses are present in `examples/phase24/transfer.schway`; checker, three independent peers, interpreter, and `emitProgram` are connected. `TestPhase24EmitterTransfer`, `TestPhase24PositiveTransferNativeApplication`, and the focused Phase 24 aggregate exercise caller use and final release; the native observer independently matches the allocation, post-return use, and physical free. Hosted `scripts/verify-phase24.sh` passed on Ubuntu and macOS in [run 37005701631](https://github.com/szTheory/schway/actions/runs/37005701631). |
| 2 | Two activations of the same static acquisition site produce distinct semantic resource identities that survive transfer and are independently checked without using raw host addresses as portable IDs. | ✓ VERIFIED | `interp` keys activation identity semantically; `corevalidate`, `originvalidate`, and `pathoracle` independently derive ownership from acquisition/call facts. `TestPhase24ActivationPeerDerivesRepeatedAcquisitions`, `TestPhase24TransferPeerRejectsMutatedOwnership`, and `TestPhase24ObserverPublicLifecycleAndTypedError` cover repeated activations, mutation rejection, and private pointer matching. The hosted focused aggregate passed on both hosts. |
| 3 | Actual entry-to-success and entry-to-typed-error executions release every non-transferred caller/callee resource in reverse successful-acquisition completion order, including multiple acquisitions followed by a real later operation/output failure; failed acquisitions contribute no obligation. | ✓ VERIFIED | `examples/phase24/error.schway` contains the repeated acquisition and real 0x43 failure path. Checker and independent peer tests derive C,B,A cleanup and failed-acquire behavior. `TestPhase24NativeErrorApplication` and `TestPhase24ObserverPublicLifecycleAndTypedError` exercise the actual app; the observer checks physical C,B,A frees and zero outstanding allocations before typed-error exit. Focused hosted runs passed on Linux and macOS. |
| 4 | Copying ownership, using a moved-from owner, or returning an owning process-entry result without an external receiver is rejected before execution with source-attributed diagnostics. | ✓ VERIFIED | `TestPhase24SourceRefusal` covers owner-copy, stale-owner use, and entry-owner escape with source attribution. Emitter refusal cases verify unsupported core is rejected before C serialization. The named source-refusal and emitter-admission groups passed in the hosted aggregate on both hosts. |
| 5 | An observer independent of compiler release events proves allocation, use after acquisition/return, and destruction before exit. Reached omitted, premature, duplicate, and wrong-resource destruction controls all fail even with plausible reported events. | ✓ VERIFIED | `phase24_observer_test.go` and its instrumented C adapter privately track actual pointers while exposing semantic operation/activation IDs. `TestPhase24ObserverReachedPhysicalDestructorControls` reaches and rejects omitted, premature, duplicate, wrong-resource, and identity-collision cases. The hosted focused aggregate passed on Ubuntu and macOS; Phase 24 validation records the host-bound receipts. |

**Score:** 5/5 roadmap truths verified (0 behavior-unverified).

### Plan Cross-Check

All three Phase 24 plans and summaries were read. Their plan-specific must-haves merge into the five roadmap truths above: bounded helper transfer and source refusals; repeated activation identity, failed-acquire behavior, and typed-error cleanup; and independent physical observation, reached controls, model-only replay, and two-host runner. Required source witnesses, test files, observer, script, and CI wiring exist. The current source and test tree matches the hosted-tested implementation revision `692f791051ba671c49c68fdd2073229feb51b090`; changes after that revision in this checkout are planning records. No existing Phase 24 UAT file was found, and no human UAT was requested or replayed.

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `examples/phase24/transfer.schway`, `error.schway` | Ordinary success and post-acquisition typed-error source witnesses | ✓ VERIFIED | Both non-stub fixtures are consumed by source, model, emitter, native-app, and observer tests. |
| `internal/compiler/check/check.go`, `core/core.go` | Bounded ownership transfer and typed-error facts | ✓ VERIFIED | Source checks record call/return ownership and success/error cleanup edges; refusal tests pin unsupported cases. |
| `corevalidate`, `originvalidate`, `pathoracle` | Independent acquisition-derived validation | ✓ VERIFIED | Separate peers are called through session tests; mutation cases cover deleted, duplicated, mismatched, reversed, colliding, failed-acquire, and frame-confused obligations. |
| `interp/interp.go`, `cgen/cgen_program.go` | Model execution and sole production C serializer | ✓ VERIFIED | Named model, emitter, and native tests exercise transfer and typed-error flows; unsupported shapes fail before serialization. |
| `phase24_observer_test.go`, `testdata/phase24_observer.c` | Independent physical lifecycle proof and reached controls | ✓ VERIFIED | Actual allocation/use/free is observed independently of compiler release events. |
| `scripts/verify-phase24.sh`, `.github/workflows/ci.yml` | Focused Phase 24 evidence in existing two-host aggregate | ✓ VERIFIED | CI invokes the focused runner in Linux and macOS evidence jobs. Both passed in run 37005701631. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| Phase 24 source fixtures | `check` | Source admission to ownership operations | WIRED | Source tests cover transfer, error edges, and attributed refusals. |
| Checked acquisition/core facts | `corevalidate`, `originvalidate`, `pathoracle` | Independent obligation derivation | WIRED | Each peer validates valid input and adversarial ownership mutations independently. |
| Checked core and foreign contracts | Interpreter / `emitProgram` | Call-return and typed-error lowering | WIRED | Model, emitter, and native tests run the fixtures; serializer preflight rejects unsupported shapes. |
| Generated app and adapter | Physical observer | Actual allocation/use/free pointer instrumentation | WIRED | Observer confirms post-return use, exact destruction, and no outstanding allocation. |
| `scripts/verify-phase24.sh` | CI evidence aggregate | One named step in Ubuntu/macOS jobs | WIRED | Both focused steps passed in run 37005701631; the runner rejects skips and unmatched groups. |

### Data-Flow Trace (Level 4)

| Artifact | Data variable | Source | Produces real data | Status |
|---|---|---|---|---|
| Success app | byte → U64 result | Caller-created one-byte file through declared adapter, helper acquire/return, and caller borrow | Yes; native application asserts independently fixed 65/66 results | ✓ FLOWING |
| Typed-error app | `UseError.UnsupportedByte` and owner obligations | Caller path, three successful helper acquisitions, then actual 0x43 use failure | Yes; native app and physical observer exercise the error path and C,B,A cleanup | ✓ FLOWING |
| Physical cleanup receipt | allocation, semantic identity, use, and destruction | Actual malloc/use/free calls; host addresses remain private to observer | Yes; observer requires zero outstanding storage before exit | ✓ FLOWING |

### Behavioral Spot-Checks

No project commands were run locally, as required by the Phase 24 validation contract. Hosted `scripts/verify-phase24.sh` executes seven named groups and fails on unmatched tests or skips.

| Behavior | Command/evidence | Result | Status |
|---|---|---|---|
| Source refusal, peer mutations, model replay, emitter admission, native apps, observer controls, and public contract | `sh scripts/verify-phase24.sh` in [hosted run 37005701631](https://github.com/szTheory/schway/actions/runs/37005701631), hosted at Phase 25 implementation head `692f791051ba671c49c68fdd2073229feb51b090` | Passed on Linux and macOS; current checkout has no source, test, or CI changes after that tested revision | ✓ PASS |
| Current merged M004 regression evidence | [Hosted run 37007281171](https://github.com/szTheory/schway/actions/runs/37007281171) | Phase 24 focused step and full Go checks passed on both hosts; corroborates the earlier Phase 24-specific receipt | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| `scripts/verify-phase24.sh` | Hosted CI step in run 37005701631 | Passed on Ubuntu and macOS; not run locally under the validation contract | PASS |

### Requirements Coverage

| Requirement | Source plans | Description | Status | Evidence |
|---|---|---|---|---|
| RES-05 | 24-01, 24-02, 24-03 | Transfer preserves live ownership; borrow preserves the obligation; declared destructor consumes it once | ✓ SATISFIED | Source/core/emitter witnesses, native app, observer, and hosted focused groups. |
| RES-06 | 24-02, 24-03 | Normal and typed-error cleanup discharges non-transferred resources in reverse successful-acquisition order | ✓ SATISFIED | Independent peer mutation tests, actual 0x43 app, and C,B,A physical observer path. |
| OWN-10 | 24-01, 24-02 | Call/return transfer, new-owner use/release, and invalid owner-use refusals | ✓ SATISFIED | Source refusal, three independent peers, emitter refusal, and native transfer app. |
| OWN-11 | 24-02, 24-03 | Activation-qualified identities survive transfer and do not use host addresses as portable IDs | ✓ SATISFIED | Repeated activation peer tests and pointer-private physical observer. |
| OWN-12 | 24-02, 24-03 | Typed-error propagation cleans caller/callee obligations in declared order | ✓ SATISFIED | Native typed-error app, cleanup mutation peers, and C,B,A physical observer. |
| EVD-09 | 24-03 | Independent physical observer and reached destruction controls | ✓ SATISFIED | All five physical controls are run by the hosted focused aggregate on Linux and macOS. |

All six requirements mapped to Phase 24 are claimed by the plans; no additional Phase 24 requirement is orphaned in `REQUIREMENTS.md`.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| — | — | No unresolved TODO/FIXME/XXX markers or Phase 24 stubs found | — | Scanner matches are empty-slice helpers, a documented placeholder field, and a Clang skip guard for an unrelated Phase 23 probe. The Phase 24 runner requires Clang and rejects any skipped Phase 24 test group. |

### Test Quality Audit

| Test files | Linked requirements | Disabled Phase 24 tests | Circular expected values | Assertion strength | Verdict |
|---|---|---:|---|---|---|
| `check_test.go`, `session_phase24_*_test.go` | RES-05/06, OWN-10/11/12 | 0 | None found | Source diagnostics and independent positive/mutated-core behavior | PASS |
| `cgen_program_test.go`, `native_app_test.go`, `phase24_observer_test.go` | RES-05/06, OWN-10/11/12, EVD-09 | 0 | None found | Exact generated order, fixed expected app result/error, and physical pointer lifecycle | PASS |
| `session_phase24_model_test.go` | RES-05/06, OWN-11/12 | 0 | None found | Fixed expected outcomes and explicit `actual_host_io=false`, `physical_cleanup=false` | PASS |

No requirement-linked Phase 24 test is disabled or circular. A `t.Skip` in the shared native test file belongs to an unrelated Phase 23 Clang ABI probe; the hosted Phase 24 runner requires Clang and independently rejects skips in every Phase 24 group.

### Human Verification Required

None. All Phase 24 success criteria assert bounded compiler/native behavior and are covered by named tests plus the independent observer. There is no existing Phase 24 UAT artifact; no human interaction was inferred from the summaries.

### Decision Coverage

The decision-coverage query reports all 9 trackable Phase 24 context decisions honored; it is non-blocking.

### Gaps Summary

No gaps found. The prior report's implementation revision was stale relative to the current source tree; the Phase 24 focused aggregate was rerun in hosted CI at the Phase 25 implementation head and passed on both required hosts. Current checkout changes since that tested source revision are planning-only. The report separates source inspection from hosted execution evidence, and does not treat model replay or compiler events as physical cleanup proof.

---

_Verified: 2026-10-02T14:20:59Z_
_Verifier: the agent (gsd-verifier)_
