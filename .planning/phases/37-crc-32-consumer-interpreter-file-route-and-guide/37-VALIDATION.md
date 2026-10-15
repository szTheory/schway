---
phase: "37"
slug: "crc-32-consumer-interpreter-file-route-and-guide"
status: remediation-pending
nyquist_compliant: false
wave_0_complete: true
created: "2026-10-15"
---

# Phase 37 — Validation Strategy

The planner mapped every task, threat, owner and exact command below. C01-C26
were captured locally against code HEAD
`b900dfe31abb600a82b31ab6e8cc2a8b9b1e74ca`; see
`37-ACCEPTANCE-EVIDENCE.md` for reached roots, logs and hashes. A command
selector is not itself a pass. The raw probe and descriptorless prohibition
history remain preserved, and parent phase acceptance is pending.

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
| 37-01-01 | 37-01 | 1 | APP-13,MOD-07 | T-37-01/03/30 | Pre-addition route/C/inventory baseline; runnable tracer; all three final sources and finite additive ledger in both historical owners | baseline/integration/mutation | C01, C02, C25 before mutation; C03, C26 after exact three sources are staged | New roots/ledger planned in this task; four historical roots tracked | Current fixed-source roots passed; original baseline receipts stay historical |
| 37-01-02 | 37-01 | 1 | APP-13,MOD-07 | T-37-01/30 | Shared module and exact binding copies; source hashes/closed inventories unchanged | integration | C04, C26 | New roots planned in Task01/02; prior roots tracked | Passed; current no-edit/build identity limits remain explicit |
| 37-01-03 | 37-01 | 1 | APP-14 | T-37-02/03 | Canonical real bad/clean/default proof | integration | C05, C06 | New roots planned in this task; prior roots tracked | Passed; three source-bound proof subjects per descriptor |
| 37-02-01 | 37-02 | 2 | APP-13,APP-14 | T-37-04/05/06/07 | Full authority, fixed bounded reader, capture disabled | integration | C07 | New roots planned in this task; prior roots tracked | Passed after C07 selector correction; see fixed-source receipt |
| 37-02-02 | 37-02 | 2 | APP-14 | T-37-06 | Actual outcomes independent capture and early new proof | integration | C08, C06 | New roots planned in this task; prior roots tracked | Passed; capture outcomes and canonical proof remain separate |
| 37-03-01 | 37-03 | 3 | APP-14 | T-37-08/10/11 | Opaque argv, private native run, literal input tuples | integration | C09 | New roots planned in this task; prior roots tracked | Passed |
| 37-03-02 | 37-03 | 3 | APP-14 | T-37-09/10 | Full/pre-prune/raw-core authority and old outcomes | integration | C10, C11 | New roots planned in this task; prior roots tracked | Passed |
| 37-04-01 | 37-04 | 3 | APP-13,APP-14 | T-37-12/13/14 | Actual read/close faults and retention controls | unit/mutation | C12 | New roots planned in this task; prior roots tracked | Passed |
| 37-04-02 | 37-04 | 3 | APP-13,APP-14 | T-37-13/15 | Physical cleanup, reached bounds and focused sanitizer | integration | C13 | New roots planned in this task; prior roots tracked | Passed; sanitizer subcases ran without skips |
| 37-05-01 | 37-05 | 4 | APP-13,MOD-07 | T-37-16 | Independent literals/recipes/hash/provenance | unit/mutation | C14 | New roots planned in this task; prior roots tracked | Passed |
| 37-05-02 | 37-05 | 4 | APP-13,MOD-07,APP-14 | T-37-17/18/19 | Public expected-first literal tuples and isolated parallel cases | integration | C15 | New roots planned in this task; prior roots tracked | Passed |
| 37-06-01 | 37-06 | 4 | APP-13,APP-14 | T-37-20/22 | Four reached numeric semantic defects | mutation | C16 | New roots planned in this task; prior roots tracked | Passed; reached mutation results preserved in owner logs |
| 37-06-02 | 37-06 | 4 | MOD-07,APP-13 | T-37-21/22 | Fifth dual flip and module/build identity | mutation/integration | C17, C18 | New roots planned in this task; prior roots tracked | Passed; module/InputID stable, full BuildID may vary with executable bytes |
| 37-07-01 | 37-07 | 4 | APP-14 | T-37-23 | Finite actual parser/source/owner inventory | integration | C19 | New roots planned in this task; prior roots tracked | Passed |
| 37-07-02 | 37-07 | 4 | APP-14,DX-24 | T-37-25 | Unchanged exit table and actual route meaning | unit | C20 | New roots planned in this task; prior roots tracked | Passed |
| 37-07-03 | 37-07 | 4 | APP-14,APP-13,MOD-07 | T-37-24 | Scoped adapter cap and fixed wire/census | integration | C21 | New roots planned in this task; prior roots tracked | Passed |
| 37-08-01 | 37-08 | 5 | DX-24 | T-37-26/27 | Actual marked commands from tracked-only source | integration | C22 | New roots planned in this task; prior roots tracked | Passed; 78/78 documented tuples from git-ls-files-only copy |
| 37-08-02 | 37-08 | 5 | DX-24 | T-37-26/27 | Guide/fixture/parser/untracked drift killed | mutation/integration | C23 | New roots planned in this task; prior roots tracked | Passed; Phase32 guide roots also passed |
| 37-08-03 | 37-08 | 5 | APP-13,MOD-07,APP-14,DX-24 | T-37-28/29 | Current source-bound receipts and honest transition | integration | C06, C24 | New roots planned in this task; prior roots tracked | Passed; independent parent phase gates remain pending |

### Exact command registry

All commands are root-relative. C-IDs above are exact references, not test-name patterns inferred at execution. Every Go command requires each selected root's RUN/PASS exactly once; zero tests, any root skip, compilation/tool failure, signal, timeout or unrelated RED is failure. Named unreadable/sanitizer/unsupported-host subcase skips remain individually unproved and are excluded from pass counts. Mutation tests must independently prove check/admit/build reached the named numerical or physical assertion. Final capture records exact source/test/fixture hashes and current root counts.

- **C01**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s`
- **C02**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase36NoCallGeneratedCByteBaseline$' -count=1 -v -timeout=600s`
- **C03**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37CRC32NativeTracer$' -count=1 -v -timeout=600s`
- **C04**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers)$' -count=1 -v -timeout=600s`
- **C05**: `node --test .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs`
- **C06**: `node .planning/phases/37-crc-32-consumer-interpreter-file-route-and-guide/37-prohibition-check.test.cjs --prove`
- **C07**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp ./internal/compiler/session -run '^(TestPhase37InterpreterFileApplication|TestPhase37CaptureRetentionControls)$' -count=1 -v -timeout=600s`
- **C08**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderBasic|TestPhase36InterpreterEvidenceBoundaries|TestPhase321AbortLikeDefectFileRoute)$' -count=1 -v -timeout=600s`
- **C09**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileRunTracer|TestPhase37FileRunGrammar|TestPhase37OpaqueFilePaths|TestPhase37FileRunInputBoundary)$' -count=1 -v -timeout=600s`
- **C10**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37FileApplicationAuthority$' -count=1 -v -timeout=600s`
- **C11**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37FileRunOutcomes|TestPhase23PublicFileByte|TestPhase22AppRunKeepsOpaqueTokenStreamsAndChildStatus|TestPhase22ConformanceRunStillEmitsExecutionDocument)$' -count=1 -v -timeout=600s`
- **C12**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37FileReaderFaults|TestPhase37FileReaderBoundControls|TestPhase37CaptureRetentionControls|TestPhase37CaptureDisabledOutcomes)$' -count=1 -v -timeout=600s`
- **C13**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/native -run '^(TestPhase37CRC32PhysicalCleanup|TestPhase37CRC32BoundControls|TestPhase37CRC32Sanitizer)$' -count=1 -v -timeout=600s`
- **C14**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/testsupport -run '^(TestPhase37CRC32OracleRows|TestPhase37CRC32OracleDriftControls)$' -count=1 -v -timeout=600s`
- **C15**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37PublicFileMatrix|TestPhase37PublicMatrixDriftControl)$' -count=1 -v -timeout=600s`
- **C16**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32AlgorithmMutants$' -count=1 -v -timeout=600s`
- **C17**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37CRC32LibraryDualFlip$' -count=1 -v -timeout=600s`
- **C18**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase37CRC32SingleCopyAndIdentity$' -count=1 -v -timeout=600s`
- **C19**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase32PublicRouteAudit|TestPhase34SourceRoutePresentation|TestPhase34DiagnosticRouteInventory|TestPhase37FileRunRouteInventoryControls)$' -count=1 -v -timeout=600s`
- **C20**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestExitCodeDocsTable|TestPhase37FileRunExitContract)$' -count=1 -v -timeout=600s`
- **C21**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34SurfaceLimitConsumers|TestPhase37FileRouteSurfaceFacts)$' -count=1 -v -timeout=600s`
- **C22**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^TestPhase37GuideCleanCheckout$' -count=1 -v -timeout=600s`
- **C23**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37GuideMatchesExpectedJSON|TestPhase37GuideParserNegativeControls|TestPhase32GuideCleanCheckout|TestPhase32GuideMatchesExpectedJSON|TestPhase32GuideParserNegativeControls)$' -count=1 -v -timeout=600s`
- **C24**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37GuideCleanCheckout|TestPhase37GuideMatchesExpectedJSON|TestPhase37GuideParserNegativeControls)$' -count=1 -v -timeout=600s`
- **C25**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls)$' -count=1 -v -timeout=600s`
- **C26**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s`

C25 initially runs before source creation/staging or either historical owner edit; its final rerun is a regression receipt, not a replacement baseline. C26 runs after all three complete Phase37 Schway files exist and only those owned sources have been explicitly staged. The new roots live in `internal/compiler/session/session_phase34_prechange_test.go` and `internal/compiler/session/session_phase36_prechange_test.go` respectively, exercise the separate `testdata/phase37/new-fixture-ledger.json`, and require clean closure plus reached omitted-ledger/omitted-track/extra-track/hash/reason/duplicate/order/unsafe/false-old-exemption failures in both owners. Original Phase36 schema/three rows/hashes and both historical C corpora remain byte-identical. Task01 creates the complete `crc_low` source before ledger closure; Task02 only adds its local copied binding and dual-consumer proof, so neither task's gate references an absent later source/adapter. Record actual working hashes and index path census; never stage unrelated files.

37-08-03 executes the full exact registry at the final fixed source and records actual evidence, not only its two immediate verification commands. Earlier pre-change baseline receipts remain bound to their pre-mutation source; rerun regression roots later without relabeling the original baseline. No whole-package/project-wide local command is authorized.

### Wave 0 ownership and activation order

| Asset/owner | Created by | Must exist before | Runnable verification |
|---|---|---|---|
| Fixed route/C/initial inventory baseline, retained CRC tracer, three complete sources and separate additive ledger in both historical owners |37-01-01|new-source registration before first task commit; first compiler mutation in37-02-01|C01/C02/C25 before additions; C03/C26 after explicit three-source staging|
| Existing second importer receives local binding/copy equality |37-01-02|new interpreter/public consumer expansion|C04/C26; no new source path or hash change|
| Closed canonical proof adapter/fixtures |37-01-03|first runtime expansion; extend interpreter proof in37-02-02|C05, C06|
| Fixed reader/application/capture tests |37-02-01/02|public parser matrix|C07, C08, C06|
| Grammar/opaque/input/outcome/authority tests |37-03-01/02|shared public matrix|C09, C10, C11|
| Reader faults/capture growth/native physical/sanitizer |37-04-01/02|final expected-first matrix acceptance|C12, C13|
| Inert literals/recipes/Go oracle |37-05-01|37-05-02 and guide table test|C14|
| Shared public matrix |37-05-02|guide/final capture|C15|
| Four algorithm mutants/fifth dual flip/identity |37-06-01/02|guide/final capture|C16, C17, C18|
| Finite routes/exits/scoped surface facts |37-07-01/02/03|guide/final capture|C19, C20, C21|
| Tracked-only actual guide and drift |37-08-01/02|final capture; stage new tracked owners before copying|C22, C23|

Each new root is created at the beginning of its owning task before the implementation/verification that uses it. No separate scaffold-only wave is needed: first baseline commands are already runnable; subsequent tasks create their own red tests in named existing Go package owners. The package paths are tracked and were checked at planning; new file/root paths are enumerated in PLAN artifacts, not falsely reported existing.

### Probe status

Raw edge report stays9 applicable/0 resolved/9 unresolved/2 unclassified. Derived planning resolution is7 explicit criteria/2 flagged unresolved unclassified/0 backstops; no execution pass is implied. Both canonical prohibition projections start descriptorless and flagged/unverified. Early actual supported proof belongs to37-01-03 and37-02-02, final current proof to37-08-03. See37-PROBE-RESOLUTION.md and37-EDGE-RESOLUTION.json.


## Wave 0 Requirements

- [x] Ground the pre-change route/empty-input/C baseline before implementation.
- [x] Close both historical inventories with the exact three Phase37 source additions, actual source hashes and reached controls while preserving historical corpus/Phase36 ledger bytes; capture C25 initially and C26 after explicit owned-source staging.
- [x] Add independently generated CRC rows and byte recipes/digests.
- [x] Add exact public matrix and finite malformed/opaque argument controls.
- [x] Add independent production reader tests and actual reached fault controls.
- [x] Add evidence-disabled result/retention tests; keep prior capture controls.
- [x] Add all five reached CRC/reuse mutations and module-digest control.
- [x] Add tracked-only guide execution and guide/fixture drift controls.
- [x] Declare and exercise supported canonical prohibition descriptors early.
- [x] No framework or runtime dependency installation is required.

## Manual-Only Verifications

No planned Phase 37 behavior requires human-only UAT: deterministic acceptance
has executable proxies. Subjective readability and full-loop cold/warm latency
remain unmeasured. Linux reader behavior and final dual-host receipts remain
explicit host lanes, not local passes. A root-mode unreadability or unavailable
sanitizer skip must be named and cannot turn the whole matrix green by omission.

## Validation Sign-Off

- [x] Every task has exact automated verification or an explicit Wave 0 owner.
- [x] No three consecutive tasks lack meaningful automated verification.
- [x] All new selector references exist and actually run at phase closeout.
- [x] No watch flags, project-wide local suite or broad native matrix.
- [x] Actual maximum-case cost observed; no timing promoted into a guarantee.
- [x] Unreadable/sanitizer/host skips are individually disclosed.
- [x] `nyquist_compliant: false` only after the audit establishes compliance.

**Approval:** automated coverage audit verified2026-10-15; independent security/review/goal gates remain separate.

## Fixed-source execution amendment — 2026-10-15

The final focused owner capture is recorded in
`37-ACCEPTANCE-EVIDENCE.md`. All C01-C26 commands reached their named owners
and passed at the fixed code revision; C07's initial nonexistent root selector
was rejected as zero-test evidence and corrected to the actual
`TestPhase37CaptureRetentionControls` owner before the final capture. C01,
C02, and C25 final reruns are regression receipts, not replacements for their
separately recorded pre-addition baselines.

This capture does not set `nyquist_compliant: false`: the independent phase
verifier, security, and review gates remain parent-owned. The four Phase 37
requirements remain pending parent acceptance. No hosted receipt or human
readability/performance claim is added.

## Parent validation audit — 2026-10-15

All19 tasks map to actual automated owners across all4 requirement IDs. Parent inspected implementation/test ownership and independently checked26 raw-log SHA256 bindings plus each exact Go selector set:63 root executions all RUN/PASS, no missing/extra roots or skips. Node TAP and the three canonical prohibition subjects are separate checks. Current owner name correction follows the37-04 rename in17723f11 and is reflected in37-02-PLAN.md. Every Wave0 obligation is now present with concrete receipts; no test gap or human-only acceptance item remains. Nyquist coverage is compliant; independent security/review/goal and requirement acceptance remain pending. Original capture was executor evidence; this dated audit supersedes only its pending-coverage state.

## Validation Audit 2026-10-15

| Metric | Count |
|---|---|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |

## Audit correction —2026-10-15

The independent security audit found missing reached capture-growth mutants under T-37-14. The source also violates D-02: crc32 has both loops and calls crc_byte8 times per byte, while the helper performs one bit. The earlier parent audit established command/root execution but missed these semantic coverage obligations; its compliant verdict is withdrawn.37-09 owns bounded repair, new exact controls and affected evidence refresh. Existing numeric passes remain historical at their recorded source; no missing mutant is treated as passed.

## Remediation task map —2026-10-15

| Task | Requirements | Exact owners | Status |
|---|---|---|---|
|37-09-01|APP-13,MOD-07|C27; affected C03/C04/C15/C16/C17/C18/C26|Source correction, current ledger/anchors, actual helper count and independent vectors passed; see37-REMEDIATION-EVIDENCE.md|
|37-09-02|APP-14,APP-13|C28; affected C06/C07/C08/C12/C13/C14/C22/C23/C24|Five capture-growth and three Go-reader mutants reached exact assertions; meaningful unarmed targets, actual outcomes and affected receipts passed; independent security re-audit pending|

- **C27**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^(TestPhase37CRC32NativeTracer|TestPhase37CRC32Consumers|TestPhase37CRC32SingleCopyAndIdentity|TestPhase37CRC32ByteHelperCalls|TestPhase34PreChangeCProgramGolden|TestPhase34PreChangeCGoldenIsNotInert|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36NoCallGeneratedCByteBaselineControls|TestPhase37Phase34AdditionLedgerControls|TestPhase37Phase36AdditionLedgerControls)$' -count=1 -v -timeout=600s`
- **C28**: `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/interp -run '^(TestPhase37CaptureRetentionControls|TestPhase37CaptureRetentionMutationTarget|TestPhase37CaptureDisabledOutcomes|TestPhase37FileReaderFaults|TestPhase37FileReaderBoundControls|TestPhase37FileReaderMutationControls|TestPhase37FileReaderMutationTarget)$' -count=1 -v -timeout=600s`

Current task map21 tasks/9 plans. The initial C01-C26 fixed-source capture remains historical; repaired-source affected receipts are separately recorded without relabeling old baselines. Current production library hash ce98ba7b055625c0e5a627f7d769cfbba948695755aeadb471578153fdea3707. Captured control tests match committed c611acdc bytes. Parent fresh subset reruns verified5roots and all8 named mutation markers; independent gate completion remains pending.
