---
phase: "27"
slug: "exact-fizzbuzz-and-cross-host-evidence"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-10-10"
---

# Phase 27 — Security

This audit checks the 26 plan-authored threats from `27-01-PLAN.md` through `27-10-PLAN.md` at configured ASVS level 1, with high severity blocking advancement. The register was authored at plan time and every threat had grep-depth evidence, so the L1 short-circuit applied and no auditor agent was spawned. Evidence is source inspection plus hosted execution; no tests or native programs were run locally for this audit. This verifies the declared mitigations; it does not claim the compiler has no other threats. The `27-09` and `27-10` summaries report no threat flags.

---

## Trust Boundaries

| Boundary | Description | Data Crossing |
|----------|-------------|---------------|
| Source to typed operations | Forged write operands or return modes could bypass checking | untrusted source text |
| Typed write to application stdout | Wrong entry selection or automatic Unit output could alter bytes | public program output |
| Source arithmetic to typed core | A zero fact or wrong operator type could escape checking | untrusted source arithmetic |
| Typed arithmetic to native C | A dynamic zero divisor could reach undefined C remainder | runtime values |
| Producer core to independent peers | A forged operation or zero fact could escape checking | core documents |
| Source literal to typed core | Dynamic or malformed text could bypass compile-time-only admission | literal bytes |
| Typed literal / input gate to C output | Wrong length, NUL truncation or oversized input could alter or prefix output | public program output |
| Child process to native runner/CLI | Capture or report errors could be mistaken for success | process outcome |
| Program outcome / event stream to evidence verdict | Agreeing engines or collapsed events could hide a wrong result | evidence documents |
| Git/source inputs and hosted logs to phase acceptance | Receipts for other bytes, revisions or failed hosts could be misattributed | CI receipts |
| Public guide to real toolchain | Stale commands could create a false reproducibility claim | documentation |

---

## Threat Register

| Threat ID | Category | Component | Severity | Disposition | Mitigation and evidence | Status |
|-----------|----------|-----------|----------|-------------|------------|--------|
| T-27-01 | Tampering | write operation admission | high | mitigate | `TestPhase27WriteAdmission` and `TestPhase27WritePeerAdmission` in `session/session_phase27_test.go` reject forged operands and return modes in checker and peers. | closed |
| T-27-02 | Information disclosure | application stdout | medium | mitigate | `TestPhase27Tracer*` in `session/session_phase27_test.go` pin exact bytes and early refusal with no Unit auto-print. | closed |
| T-27-03 | Tampering | C17 remainder lowering | high | mitigate | `cgen/cgen_program.go` `OpRemainderU64` emits the zero-divisor terminal defect before the C remainder expression; `executionpeer/executionpeer.go` checks `remainder_zero_*` attribution; `TestPhase27DynamicZero` in `session/session_phase27_remainder_test.go`. | closed |
| T-27-04 | Tampering | core/operator admission | high | mitigate | Static-zero and operand-type refusal tests in `session/session_phase27_remainder_test.go`; independent core check via `TestPhase27OverflowZeroControls` in `session/session_phase27_evidence_test.go`. | closed |
| T-27-05 | Tampering | literal text admission | high | mitigate | `TestPhase27LiteralWriteAdmission` in `session/session_phase27_test.go`; `strconv.Unquote` decoding in `native/native.go` and an independent re-derivation in `originvalidate/originvalidate.go`. | closed |
| T-27-06 | Denial of service | application byte writer | high | mitigate | `TestPhase27OutputEvidenceExactLimit` and `TestPhase27EvidenceCapacity` in `native/native_app_test.go` cover exact and crossing limits. | closed |
| T-27-07 | Tampering | process-outcome mapping | high | mitigate | `TestPhase27OutputEvidenceSeparation` in `session/session_phase27_output_test.go` asserts child status, stderr, stdout, tool error and evidence exhaustion separately. | closed |
| T-27-08 | Tampering | exact-answer adjudication | high | mitigate | `TestPhase27ExactAnswerMatrix`, `TestPhase27ReachedWrongRemainder` and `TestPhase27ReachedWrongArithmetic` in `session/session_phase27_evidence_test.go` use fixed answers per engine and reached source mutants. | closed |
| T-27-09 | Tampering | event evidence | high | mitigate | `TestPhase27SkippedIteration` and `TestPhase27RepeatedEventIdentity` in `session/session_phase27_evidence_test.go` inject omission, collision and reordering controls. | closed |
| T-27-10 | Repudiation | evidence-capacity verdict | medium | mitigate | `TestPhase27EvidenceCapacity` (`native/native_app_test.go`) and `TestPhase27OutputEvidenceSeparation` force incomplete or exhausted evidence to non-success with a separate child outcome. | closed |
| T-27-11 | Spoofing | host receipt identity | high | mitigate | `scripts/check-phase27-receipts.sh` accepts only `macOS`/`Linux` hosts and rejects `duplicate_host_*`; receipts record uname target and toolchain; `TestPhase27HostReceiptGate`. | closed |
| T-27-12 | Tampering | source/fixture identity | high | mitigate | The gate recomputes git blob hashes from the checkout and rejects `source_hash_mismatch`, `fixture_hash_mismatch`, `input_set_mismatch` and `wrong_revision`; `TestPhase27HostReceiptGate`. | closed |
| T-27-13 | Repudiation | CI execution logs | medium | mitigate | `.github/workflows/ci.yml` uploads each host's receipt log with `if: always()` and 14-day retention; `TestPhase27CIHostMatrix`. | closed |
| T-27-14 | Repudiation | host provenance | high | mitigate | `27-HOST-EVIDENCE.md` records run 38054848320, URL, SHA 6224f8b, host conclusions, input hashes and the two saved receipts `host-macos.receipt` and `host-linux.receipt`. | closed |
| T-27-15 | Tampering | documentation commands | medium | mitigate | `TestPhase27GuideContract` in `session/session_phase27_docs_test.go` checks the guide's links, tracked paths and claims and executes the pinned gate command. | closed |
| T-27-16 | Denial of service | FizzBuzz input/output | medium | mitigate | `cmd/schway/main_test.go` runs FizzBuzz at 0, 1, 15, exact 1000 lines (`fizzbuzz-1000.out`) and refuses 1001 with exit 65 and no stdout; the host receipts include input 1001 on both hosts. | closed |
| T-27-17 | Tampering | package/tool supply chain (27-01) | high | mitigate | `go.mod`/`go.sum` last changed at 21936ce, before Phase 27; no new module dependencies. | closed |
| T-27-18 | Tampering | package/tool supply chain (27-02) | high | mitigate | Same as T-27-17; native route stays on the existing C17/Clang backend. | closed |
| T-27-19 | Tampering | package/tool supply chain (27-03) | high | mitigate | Same as T-27-17. | closed |
| T-27-20 | Tampering | package/tool supply chain (27-04) | high | mitigate | Same as T-27-17. | closed |
| T-27-21 | Tampering | package/tool supply chain (27-05) | high | mitigate | Same as T-27-17; literal decoding uses the standard library `strconv`. | closed |
| T-27-22 | Tampering | package/tool supply chain (27-06) | high | mitigate | Same as T-27-17. | closed |
| T-27-23 | Tampering | package/tool supply chain (27-07) | high | mitigate | Same as T-27-17. | closed |
| T-27-24 | Tampering | package/tool supply chain (27-08) | high | mitigate | Same as T-27-17; Go test infrastructure only. | closed |
| T-27-25 | Tampering | package/tool supply chain (27-09) | high | mitigate | `ci.yml` uses only `actions/checkout@v4`, `actions/setup-go@v5` and `actions/upload-artifact@v4`; its only package install (`libc++-dev libc++abi-dev` via apt) predates Phase 27 (7b921d0, 2026-09-29). | closed |
| T-27-26 | Tampering | package/tool supply chain (27-10) | high | mitigate | The guide documents only the retained Go and Clang toolchain; no new install. | closed |

*Status: open · closed · open — below high threshold (non-blocking)*
*Severity: critical > high > medium > low — only open threats at or above workflow.security_block_on count toward threats_open*
*Disposition: mitigate (implementation required) · accept (documented risk) · transfer (third-party)*

Paths without a top-level directory are under `internal/compiler/`.

---

## Accepted Risks Log

No accepted risks.

Known limit, not a threat disposition: the receipt gate checks consistency between receipts and the checkout, not receipt authenticity; the hosted run record is the authenticity anchor (see `27-HOST-EVIDENCE.md` and Phase 27.1's `27.1-HOST-EVIDENCE.md`).

---

## Security Audit Trail

| Audit Date | Threats Total | Closed | Open | Run By |
|------------|---------------|--------|------|--------|
| 2026-10-10 | 26 | 26 | 0 | /gsd-secure-phase 27 (orchestrator, L1 short-circuit) |

Current-source hosted evidence: Phase 27.1 changed `session/session.go` and `session/session_app_verify.go` after 6224f8b. Run [38068888687](https://github.com/szTheory/schway/actions/runs/38068888687) at 7936dac passed the full and race suites and the Phase 27 receipt lanes on macOS and Ubuntu, and none of the controls listed above were removed.

---

## Sign-Off

- [x] All threats have a disposition (mitigate / accept / transfer)
- [x] Accepted risks documented in Accepted Risks Log
- [x] `threats_open: 0` confirmed
- [x] `status: verified` set in frontmatter

**Approval:** verified 2026-10-10 at the configured L1 depth.
