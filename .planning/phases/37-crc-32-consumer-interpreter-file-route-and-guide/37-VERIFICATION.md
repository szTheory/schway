---
phase: 37-crc-32-consumer-interpreter-file-route-and-guide
verified: 2026-10-16T21:20:20Z
status: passed
score: 14/14 must-haves verified
covered_files:
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-01-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-01-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-02-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-02-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-03-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-03-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-04-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-04-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-05-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-05-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-06-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-06-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-07-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-07-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-08-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-08-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-09-PLAN.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-09-SUMMARY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-371-REFRESH-2026-10-16-VERIFICATION-HISTORY.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-371-REFRESH-GATES.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-371-REFRESH-RUN.json"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REFRESH-CLOSEOUT.md"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REFRESH-COMPLETION.json"
  - ".planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-REFRESH-REGRESSION.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-03.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-04.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-371-REFRESH-REGRESSION-09.jsonl"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-REVIEW.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-SECURITY.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-COMPLETION-2026-10-16.json"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DISPOSITION-AMENDMENT-2026-10-16.md"
  - ".planning/phases/37.1-close-gap-legacy-terminal-outcomes/371-REFRESH-DOC-GUARDS-20261016.jsonl"
  - "cmd/schway/main.go"
  - "cmd/schway/main_phase37_crc32_controls_test.go"
  - "cmd/schway/main_phase37_file_run_test.go"
  - "cmd/schway/main_phase37_guide_test.go"
  - "cmd/schway/main_phase37_matrix_test.go"
  - "examples/crc32/README.md"
  - "examples/crc32/adapter.c"
  - "examples/crc32/adapter.h"
  - "examples/crc32/crc32.expected.json"
  - "examples/crc32/crc32.schway"
  - "examples/crc32/crc_low.schway"
  - "examples/crc32/file_bytes.bindings.json"
  - "examples/crc32/lib/crc32.schway"
  - "internal/compiler/cgen/cgen_exit_literals_test.go"
  - "internal/compiler/cgen/cgen_phase33_scalar_defect_test.go"
  - "internal/compiler/cgen/cgen_phase371_legacy_defect_test.go"
  - "internal/compiler/cgen/cgen_program.go"
  - "internal/compiler/corevalidate/corevalidate.go"
  - "internal/compiler/executionpeer/executionpeer.go"
  - "internal/compiler/executionpeer/executionpeer_phase33_terminal_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase36_test.go"
  - "internal/compiler/executionpeer/executionpeer_phase371_test.go"
  - "internal/compiler/interp/filebytes_reader.go"
  - "internal/compiler/interp/filebytes_reader_test.go"
  - "internal/compiler/native/native_phase37_filebytes_test.go"
  - "internal/compiler/session/session_file_application.go"
  - "internal/compiler/session/session_phase24_model_test.go"
  - "internal/compiler/session/session_phase34_prechange_test.go"
  - "internal/compiler/session/session_phase36_prechange_test.go"
  - "internal/compiler/session/session_phase371_compatibility_test.go"
  - "internal/compiler/session/session_phase371_defect_reason_test.go"
  - "internal/compiler/session/session_phase37_crc32_test.go"
  - "internal/compiler/session/session_phase37_file_application_test.go"
  - "internal/compiler/testsupport/crc32_oracle.go"
  - "internal/compiler/testsupport/crc32_oracle_test.go"
  - "testdata/phase16/public-emitter-consumers.json"
covered_digest: "v3:sha256:5baea201f9941cffbd390e82c2aed01595fc1446c9322fd1c9c53c6fd1cbd8b4"
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: passed
  previous_score: 14/14
  gaps_closed: []
  gaps_remaining: []
  regressions: []
decision_coverage:
  honored: 24
  total: 24
  not_honored: []
---

# Phase 37: CRC-32 Consumer, Interpreter File Route, and Guide Verification Report

**Phase Goal:** A developer can run `crc32` on one file and get exact CRC-32/ISO-HDLC from an ordinary-source module using bitwise operations and helper calls; a second app imports that module unchanged; both public CLI engines run file-bytes programs with identical results; and a clean checkout reproduces the workflow from the guide.

**Verified:** 2026-10-16T21:20:20Z
**Status:** passed
**Verification mode:** Initial-mode goal-backward refresh. A previous passed report existed without a `gaps:` section, so this is not a gap-closure pass. All nine plans are complete. Existing completed UAT and disposition records were preserved; no human readability approval is claimed.

**Evidence revision:** current checkout `20a519468839a0698e2c4e93fc51300509c6abbf` (Darwin arm64). The previous exact public matrix and guide receipts were recorded at `9258a7360222ff17a3267e1ccb873a0dfae79346`; their Phase 37 implementation and test inputs are unchanged. This refresh separately exercised current CLI, helper, emitter, and peer paths and records the limits below.

## Goal Achievement

The four roadmap success criteria, all nine plans' must-haves, and the four mapped requirements were checked against source, named tests, and current gate receipts. Repeated plan details are merged into 14 observable truths. Previous summaries were treated as pointers; current source and preserved exact test evidence determine the verdict.

### Observable Truths

| # | Truth | Status | Evidence |
|---|---|---|---|
| 1 | The shared ordinary-source module implements reflected CRC-32/ISO-HDLC with the required initial/final XOR, one byte loop, and one private `crc_byte` call per byte; the helper performs eight bitwise rounds. | VERIFIED | `examples/crc32/lib/crc32.schway`; current `TestPhase37CRC32ByteHelperCalls` observes one helper activation per input byte. |
| 2 | The pinned answers are exact, unsigned, and independently generated, including the empty, single-zero, canonical text, and 65,536-byte cases; 65,537 bytes is refused. | VERIFIED | `crc32.expected.json` records Python 3.14.4/zlib 1.2.12 provenance and recipes; Go IEEE cross-check is separate. Exact values: 0, 3523407757, 3421780262, and 3735780942. The prior public matrix covered the boundary and independent drift controls. |
| 3 | A second application imports the shared module unchanged and uses the same adapter/binding triple without copying the CRC algorithm. | VERIFIED | Both `.schway` consumers import `lib.crc32`; current `TestPhase37CRC32Consumers` checks the exact source census and byte-identical adapter, header, and manifest, and runs the native consumer. |
| 4 | Interpreter and native public CLI routes produce the same pinned file result and exact exit/stdout/stderr tuples for the declared cases. | VERIFIED | Current `TestPhase37PublicFileRunTracer` returns `3421780262\n` with empty stderr in both engines; current `TestPhase37FileRunOutcomes` checks typed failures and terminal exits. The full expected-first matrix passed in the preserved prior receipt; it was not rerun because its named root took 35.6 seconds. |
| 5 | Interpreter routing avoids Clang and native routing requires the manifest and uses a private one-shot build. | VERIFIED | Current CLI tracer uses a Clang-unavailable sentinel for interpreter mode and checks native temporary-build cleanup; `cmd/schway/main.go` dispatch and previous grammar/manifest controls agree. |
| 6 | The interpreter admits the full module closure and only the declared FileBytes owner path before executing file input. | VERIFIED | `session_file_application.go` validates policy, ABI, owner, and module closure before entry execution; named Phase 37 authority and file-application tests are preserved and source is unchanged. |
| 7 | Interpreter input is read from one opened regular descriptor with a 65,537-byte actual-read limit, stable identity, and once-only close/error handling. | VERIFIED | `filebytes_reader.go` enforces the cap-plus-one bound and error ordering. Prior named reader fault, bound, and mutation controls passed; this reader and its tests are unchanged in the refreshed source delta. |
| 8 | Capture-disabled execution retains real output while preventing evidence growth and preserving output order. | VERIFIED | Previous named capture-retention and capture-disabled outcome tests exercised the recorders, events, invocations, segments, and ordinal controls. The source paths are unchanged by the current delta. |
| 9 | Native cleanup is physically observed on CRC consumers with focused sanitizer evidence and host limits stated. | VERIFIED | Previous named CRC physical-cleanup, bound, and sanitizer tests observe the actual adapter resources. Current evidence remains local Darwin-specific and does not claim Darwin leak detection. |
| 10 | Algorithm mutations and library-only edits are causal on both consumers; module identity invalidation is checked separately. | VERIFIED | Previous named algorithm-mutant, single-copy/identity, and library dual-flip tests passed. Expected values remain independent literals rather than values derived from the implementation. |
| 11 | CLI grammar, opaque paths, route inventory, exit classes, and surface facts agree with the public contract. | VERIFIED | Current CLI outcome roots and the refreshed emitter/peer roots pass. `cgen` exit literals now include the tested terminal class 74; the initial test-only stale expectation was corrected and the named test rerun green. |
| 12 | The guide runs real commands from a clean tracked-file copy and checks the documented results and errors. | VERIFIED | `TestPhase37GuideCleanCheckout` previously passed, copying `git ls-files` content and comparing 78 tuples across two consumers and three routes. The guide, source, and test are unchanged. Current documentation guards passed 6/6. |
| 13 | Test-tier prohibitions have canonical bad, clean, and default evidence, with numeric and causal claims kept distinct. | VERIFIED | The Phase 37 canonical producer and current Phase 37.1 supplemental controls retain green bad/clean/default subjects, with no flagged prohibition; the refreshed security gate reports 13/13 supplemental controls closed and zero open threats. |
| 14 | Current evidence distinguishes fresh executions from historical receipts and preserves transition dispositions. | VERIFIED | Current gate receipts bind source/test inputs and raw logs. The existing transition disposition is retained; this report does not relabel deferred CR-03/04/06 or WR-02 obligations as resolved by Phase 37. |

**Score:** 14/14 truths verified; 0 behavior-unverified.

### Current Source Delta and Gate Evidence

The refreshed comparison began at `9258a736` and ends at `20a5194`. Paths shared with the prior report include the C emitter, execution peer, and Phase 34/36 prechange comparators. The changed emitter and peer paths have current named regression owners; no scalar payload, numeric outcome-value, or CFG-branch wire-authentication theorem is inferred from them.

The final bounded predecessor registry reports 250 package/test selectors, 251 root execution events, 240 distinct names, and 16 package groups. It reused 220 unchanged passing root receipts, reran five prior failures green, and ran 26 previously unrun roots green; the final gate has zero failed, missing, or skipped roots. Four current D18 proofs cover 37 inputs each; their 19-root union passed 254 subtests with no skips. Six documentation/lifecycle guards passed. The pressure selector `TestPhase33CaptureReadBounded` was excluded. This is a bounded registry, not the full historical 98-root acceptance inventory.

The initial exit-literal run failed only because its test expected `[0 65 69]` while current source correctly emitted `[0 65 69 74]`. The test expectation was updated and the named test passed on rerun. Both raw runs remain in the receipt; the initial red was not hidden or treated as a product defect.

Current bounded behavior checks run for this refresh:

- `GOCACHE=/tmp/ai-lang-verification-gocache-count1-v120 go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunOutcomes)$' -count=1 -json -timeout=10s` — PASS, 2 roots; 1.267s. Both engines return the canonical result, interpreter avoids Clang, and typed file failures preserve their exit and stream contract.
- `GOCACHE=/tmp/ai-lang-verification-gocache-count1-v120 go test ./internal/compiler/session -run '^(TestPhase37CRC32Consumers|TestPhase37CRC32ByteHelperCalls)$' -count=1 -json -timeout=10s` — PASS, 2 roots; 0.565s. The second native consumer uses the shared source/bindings, and the helper is called once per byte.
- Current cgen and executionpeer regression log groups passed `TestEmittedCExitLiteralsMatchTable`, `TestPhase33ScalarExplicitDefectShells`, and `TestPhase36CallCausalAttributionControls` (12 controls). Exact logs and hashes are in the current 37.1 completion receipt.

The prior full file matrix passed at the previous source revision and the Phase 37 application, reader, oracle, and matrix-test inputs remain unchanged. The full matrix and clean-checkout guide root were not rerun in this refresh: their recorded named-root durations were 35.6 and 15.0 seconds, respectively. Their existing receipts are preserved and not presented as newly executed checks.

## Required Artifacts

All nine Phase 37 plan artifact queries passed: 27/27 declarations. The substantive source files, module imports, binding ownership, named tests, fixture, and guide are connected as expected.

| Artifact | Expected | Status | Evidence |
|---|---|---|---|
| `examples/crc32/lib/crc32.schway` | Ordinary-source CRC implementation | VERIFIED | One shared byte loop and private bitwise helper. |
| `examples/crc32/crc32.schway`, `crc_low.schway` | Two consumers | VERIFIED | Both import `lib.crc32`; current consumer census forbids a copied loop. |
| `examples/crc32/adapter.c`, `adapter.h`, `file_bytes.bindings.json` | FileBytes adapter contract | VERIFIED | Current consumer test compares these files byte-for-byte. |
| `examples/crc32/crc32.expected.json` | Independent pinned answers | VERIFIED | Python/zlib generation provenance, recipes, digests, and literals. |
| `internal/compiler/session/session_file_application.go` | Interpreter admission boundary | VERIFIED | Admission and owner checks precede execution. |
| `internal/compiler/interp/filebytes_reader.go` | Bounded descriptor reader | VERIFIED | One regular descriptor; cap-plus-one read and cleanup paths. |
| `cmd/schway/main.go` | Public CLI routing | VERIFIED | Interpreter and native file forms reach intended engines. |
| `internal/compiler/testsupport/crc32_oracle.go`, `main_phase37_matrix_test.go` | Independent expected-first matrix | VERIFIED | Each engine is checked against expected tuples before parity. |
| `internal/compiler/native/native_phase37_filebytes_test.go` | Physical cleanup and sanitizer witnesses | VERIFIED | Actual adapter resource operations are observed. |
| `examples/crc32/README.md`, `main_phase37_guide_test.go` | Executable clean-checkout guide | VERIFIED | README fences run in a tracked-file temporary checkout. |
| Current cgen/executionpeer changes | Adjacent terminal-outcome repairs | VERIFIED | Current named emitter and peer tests pass; scope is limited to the tested defect/terminal paths. |

## Key Link Verification

The lexical query verified 4/23 plan links. I manually traced the remainder through imports, Go calls, fixtures, readers, and test owners; filename matching alone does not encode these relationships.

| Plan | Lexical query | Manual trace |
|---|---:|---|
| 37-01 | 0/4 | Both imports resolve to `lib.crc32`; current inventory owners read the source ledger; prohibition harness invokes its canonical producer. |
| 37-02 | 0/3 | Module admission precedes entry resolution; interpreter application reaches the fixed reader; capture is disabled before recorder growth. |
| 37-03 | 0/2 | CLI dispatch reaches interpreter mode; native mode builds and runs the manifest application once. |
| 37-04 | 0/2 | Fault tests reach the descriptor seam; physical observers reach C adapter open/read/free/close. |
| 37-05 | 0/2 | The matrix reads fixture IDs/literals; the oracle cross-checks them against Go IEEE. |
| 37-06 | 1/2 | Mutants target the shared library; reuse tests apply library-only edits to closure and receipt identities. |
| 37-07 | 1/3 | Route census resolves dispatch; exit-contract tests check docs; surface tests inspect manifest and consumers. |
| 37-08 | 1/3 | Guide tests execute marked README fences from tracked files; parser and expected-result controls check drift. |
| 37-09 | 1/2 | Helper-call test observes one call per byte; capture controls assert exact growth discriminators. |

## Data-Flow Trace (Level 4)

| Artifact | Data | Source | Real data | Status |
|---|---|---|---|---|
| Interpreter application | ByteView | Bytes read from opened input descriptor | Yes | FLOWING |
| Native FileBytes consumer | Bytes and length | Local C adapter reads supplied path into owned storage | Yes | FLOWING |
| CRC result | Accumulator/output | Ordinary-source algorithm consumes actual ByteView bytes | Yes | FLOWING |
| Expected rows | Fixed tuples | Python/zlib fixture cross-checked with Go IEEE | Yes, independent oracle | FLOWING |
| Clean-checkout guide | Source and test files | `git ls-files` copied into temporary checkout | Yes | FLOWING |

## Behavioral Spot-Checks

| Behavior | Command or evidence | Result | Status |
|---|---|---|---|
| Current public interpreter/native file run and typed outcomes | `go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunOutcomes)$' -count=1 -json -timeout=10s` | 2 roots pass; `3421780262\n`, empty stderr for both engines; exact typed failure tuples | PASS |
| Current shared consumer and helper activation | `go test ./internal/compiler/session -run '^(TestPhase37CRC32Consumers|TestPhase37CRC32ByteHelperCalls)$' -count=1 -json -timeout=10s` | 2 roots pass; shared module and one helper activation per byte | PASS |
| Changed emitter/peer outcomes | 37.1 regression logs 04 and 09 | Exit-literal and scalar-defect roots pass; 12 peer attribution controls pass | PASS |
| Full pinned public file matrix | Prior Phase 37 acceptance receipt at `9258a736`; current matrix inputs unchanged | Passed previously; not rerun in this refresh (35.6s recorded duration) | PASS (historical receipt) |
| Clean-checkout guide | Prior `TestPhase37GuideCleanCheckout` receipt; current guide inputs unchanged | Passed previously across 78 exact tuples; not rerun in this refresh (15.0s recorded duration) | PASS (historical receipt) |

## Probe Execution

No Phase 37 plan or summary declares a probe, and no conventional `scripts/*/tests/probe-*.sh` exists. The canonical prohibition producer has prior Phase 37 evidence; this refresh did not substitute another command for a declared probe.

## Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| APP-13 | 37-01, 37-05, 37-06, 37-09 | Ordinary-source CRC-32/ISO-HDLC consumer with pinned values and bounded file behavior | SATISFIED | Truths 1, 2, 4, and 10; independent oracle, helper activation, mutation controls, and public matrix evidence. |
| MOD-07 | 37-01, 37-06 | A second consumer imports the shared module and produces the pinned answer on both engines | SATISFIED | Truths 3 and 10; current shared-consumer test and preserved library-flip controls. |
| APP-14 | 37-02 through 37-08 | Public interpreter file route and native route produce identical pinned streams/status | SATISFIED | Truths 4–9 and 11–12; current CLI smoke and prior matrix, reader, outcome, cleanup, and guide evidence. |
| DX-24 | 37-08 | Executable clean-checkout CRC guide | SATISFIED | Truth 12; preserved clean-checkout test receipt and current documentation guards. |

The canonical phase mapping contains exactly `APP-13`, `MOD-07`, `APP-14`, and `DX-24`; no additional mapped requirement is orphaned.

## Decision Coverage

The canonical decision-coverage query returned 24 honored out of 24 trackable decisions, with no unhonored decisions.

## Test Quality Audit

The expected CRC values are independent literals generated by Python/zlib and cross-checked using Go's IEEE implementation. The public matrix compares each engine to expected tuples before comparing the engines. Mutation controls exercise wrong polynomial/recurrence/final XOR/bound behavior, and the shared-library change is checked against both consumers. The test scan found no disabled requirement-linked tests or circular expected-value path. Current gate logs retain the initial exit-literal assertion failure and its test-only correction; the rerun passed.

## Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| | | None | | |

The relevant source scan found no unreferenced TBD/FIXME/XXX markers, TODO/HACK/PLACEHOLDER comments, empty user-visible implementations, hardcoded empty user-visible data, or console-log-only implementation. C emitter empty collection values are documented semantic defaults, not stubs.

## Security and UI

The existing authored Phase 37 threat register has zero open threats. The supplemental Phase 37.1 register closes 13/13 scoped controls, with no accepted-risk text changed. The current independent eight-file delta review is clean. UI review is not applicable because this phase has no frontend, screen, or UI specification.

## Human Verification Required

None for the stated CLI/compiler goal. Existing completed UAT and disposition records remain preserved. No human readability approval or subjective readability result is claimed.

## Gaps Summary

No Phase 37 gap remains: the four roadmap criteria, 14 observable truths, 27 declared artifacts, 23 traced links, and four mapped requirements are satisfied. No overrides or deferred Phase 37 gaps were used.

Evidence is local Darwin arm64 only. This report does not claim hosted/Linux behavior, a full-project or whole-package run, a rerun of the full file matrix or guide root at the current revision, cold/warm latency distributions, a universal native-binding limit, arbitrary divergent-interpreter timeout parity, or Darwin leak detection. Phase 37.1's bounded regression inventory is not the full historical 98-root acceptance suite. The report also makes no human readability claim.

Phase 33 CR-03/CR-04/CR-06 and Phase 36 WR-02 remain separate disposition obligations; Phase 37 does not resolve or accept them. Phase 38 remains unstarted; its EVD-17 scope is separate from this verdict.

---

Verified: 2026-10-16T21:20:20Z
Verifier: the agent (gsd-verifier)
