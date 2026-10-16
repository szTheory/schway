<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

DATA_Q7M4P2LX_START
## Implementation Decisions

### Coverage and host producer

- **D-01:** Use one explicit M007 program/case/control manifest. Preserve all seven programs and 39 exact case rows from `testdata/receipts/M006.manifest`, including checksum/adler_low, the Phase 29 fixtures and OOB control, sum_to_n, and FizzBuzz. Add CRC-32, its second importer, bitwise and scalar helper witnesses. Do not substitute a successful workflow badge or a program-name list for executed cases.
- **D-02:** Reuse the maintained host-driver patterns and generic gates. Prefer an opt-in M007 mode of the existing driver with the M006 default contract preserved; add only the route handling the concrete new witnesses require. Do not clone the freeze/receipt gates or build a general runner framework. A small local helper is acceptable when it reduces duplicated acceptance logic. Standard library and installed tools suffice; no new dependency is justified.
- **D-03:** Pin CRC cases and independently generated answers to `examples/crc32/crc32.expected.json`. Account for all its CRC consumer cases, including boundaries, opaque input grammar and filesystem outcomes. NUL cannot be passed through process argv: record its direct-dispatch test as a named reached control, rather than inventing an external CLI case. Use the shared public-route matrix for interpreter/native parity. Preserve the earlier programs' pinned cases rather than expanding each into an unrelated new matrix.
- **D-04:** Cover all pinned helper answer/defect rows from `testdata/phase36/helper-values.json`, with local/imported calls and the public unused-Byte conformance witness on its actual supported route. Cover the bitwise result 264 and its alternate-association/refusal controls. A scalar application and a structured conformance result have different output contracts; validate each through its existing consumer, not a shared decimal-output assumption.
- **D-05:** Include focused Phase 37.1 regression controls for legacy/dynamic defect closure, checked `foreign.failed → OpFail` attribution, valid/invalid reason admission, counted native reason bytes, no-call structured defects and exact compatibility amendments. These are named controls in the hosted run, not a new corpus-wide legacy-program matrix. Preserve the distinction between causal admission, numerical answers, process tuples and physical cleanup.
- **D-06:** A required case/control must actually run and pass on each host. Empty selectors, skipped required subtests, unavailable native tools, missing case rows and partial runs fail acceptance. Environment-dependent permission/FIFO/device/symlink controls require explicit reached evidence or a stated blocker; a named skip is not a pass. Use exact selectors and verify expected test/subtest names, not only Go's exit status. The producer's completion inventory must agree with the manifest.
- **D-07:** Keep the existing hosted `checks` and `current evidence aggregate` jobs for both OSes, requiring their full and race lanes to succeed in the selected run. Add M007 artifacts to this established workflow. Default-build receipt tuples and the O0/O3/O3-LTO tests are separate claims; identify which named tests establish optimization coverage. Avoid an extra redundant broad job. Local preparation uses focused deterministic checks only: no full local suite, whole-package sweep, broad cross-program matrix or capture-pressure experiment.

### Freeze and authority

- **D-08:** Keep the codedigest boundary. The final freeze records a full commit SHA, full codedigest, manifest identity and audited preparation state. All changes to driver, manifests, tests, CI and other digest inputs land before freeze, with needed review, scoped verification freshness and a fresh milestone audit afterward. The current audit already closes CR-03/04/06, but does not audit future preparation edits. Never label the current discussion commit the final freeze in advance.
- **D-09:** Check both committed digest paths and their worktree cleanliness; the current generic freeze gate alone does not check uncommitted files. Preserve the pre-existing modified `.planning/config.json` byte-for-byte and exclude it from staging. Planning-only descendants remain allowed after freeze, and the actual final descendant must pass ancestry/no-code-drift checks. No digest-path edit after hosted execution may silently retain current receipts.
- **D-10:** Prepare all code, validation, audit and the concrete freeze record before asking for remote authorization. One checkpoint must name the actual full freeze SHA and codedigest, repository, destination branch/ref, workflow and inputs, and push scope. Inspect current remote/ref state before proposing the command; do not merge to main, force-push, publish tags, rewrite history or broaden the authorized action. The current branch is `phase26-validation`; local tracking information is not a freshly fetched remote fact. No push or dispatch is authorized by this discussion.
- **D-11:** Dispatch the selected ref only while it identifies the freeze commit; independently verify the resulting run's `head_sha` equals that commit and both raw host logs claim that same `headsha`. Codedigest equality supports later planning-only descendants, but does not waive exact provenance for the hosted run itself. Add an armed wrong-revision control at the binding seam. Reuse the generic gate's existing consistency checks rather than implying it authenticates hosted metadata.

### Acceptance, recovery and durable ingest

- **D-12:** Bind one wholly successful hosted run/attempt with both hosts and every required job, artifact, case and control present. Record repository, workflow path/event, run ID/attempt/URL, ref/head SHA and conclusions; required job IDs/conclusions; artifact names/IDs/digests/expiry; and raw log repository paths, byte counts and SHA-256. Verify source provenance separately from consistency checks. No cross-revision or cherry-picked cross-attempt pair is acceptable.
- **D-13:** Copy both raw host logs and bounded relevant provenance into `.planning/phases/38-milestone-receipts/` during the last plan. Re-gate those exact copied bytes. Artifact URLs alone are insufficient for durable evidence; hosted retention expires. Keep failure evidence labeled historical and preserve earlier receipts. Do not copy credentials, inherited environment dumps or irrelevant API user/profile data into the receipt.
- **D-14:** Permit one retry at the same freeze for a diagnosed transient infrastructure failure; record the failed run/attempt and diagnosis. Do not loop on deterministic test failures. A code fix reopens freeze, validation/audit and the commit-specific remote checkpoint. Use a fresh complete successful run, not partial job results assembled into a green pair. If hosted authority/access or required results are unavailable, leave EVD-17 pending and report the external blocker; do not create a fake implementation gap or subjective UAT.
- **D-15:** Keep the expected three-plan shape: (1) prepare/verify/audit/freeze, (2) commit-specific push/dispatch checkpoint and hosted observation, (3) `.planning`-only ingest and final binding. Bind receipts once in the last plan. EVD-17 remains Pending until both hosted evidence and final drift checks pass; no preceding plan or summary may complete it prematurely. Extra plans require a concrete safety obligation, not additional ceremony.
DATA_Q7M4P2LX_END

### the agent's Discretion

DATA_V3K8N1RW_START
Choose exact helper/file boundaries, manifest inventory encoding, bounded metadata projection and focused test arrangement during research/planning. Preserve the decisions above, existing gate refusal behavior and M006 default producer contract. Resolve routine implementation details autonomously. Remote authorization is reserved for the concrete D-10 checkpoint.
DATA_V3K8N1RW_END

### Deferred Ideas (OUT OF SCOPE)

DATA_C9L2F6TA_START
Arity-N, scalar records/exported types, wc/longest_line, key=value/JSON, separate compilation, module-aware queries, generic observability, signed provenance and whole-loop performance studies retain their existing future consumer triggers. No new scope was requested. Native execution remains process execution, and codedigest/host receipts do not establish hermetic builds or physical cleanup beyond their named tests.
DATA_C9L2F6TA_END
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| EVD-17 | DATA_H5N8Q2LA_START “Hosted ubuntu and macOS receipts for every M007 and earlier pinned program are bound once to the final M007 codedigest, after the milestone audit and in its last plan. No per-phase hosted receipt binding is required.” DATA_H5N8Q2LA_END [VERIFIED: .planning/REQUIREMENTS.md:84] | Producer, test owners, generic gates, hosted jobs, exact run/attempt binding and final copied-byte ingest are mapped below. Status is verbatim: DATA_K3W7P1VC_START “EVD-17 | Phase 38 | Pending” DATA_K3W7P1VC_END [VERIFIED: .planning/REQUIREMENTS.md:128]. |
</phase_requirements>

# Phase 38: Milestone Receipts — Research

**Researched:** 2026-10-16  
**Domain:** Native CLI evidence, hosted CI provenance, durable receipts  
**Confidence:** MEDIUM overall — repository contracts were inspected; remote authority and hosted results remain unobserved.

## Summary

Extend the existing host receipt producer through an opt-in M007 mode while preserving the no-argument M006 contract. One manifest should retain the inherited cases and add actual CRC, helper, bitwise, and focused Phase 37.1 controls. Reuse codedigest and receipt-consistency gates, then add narrow binding between the actual freeze commit, selected GitHub run/attempt, and both host logs. The existing receipt gate validates internal consistency and nonempty headsha; it does not authenticate a hosted revision. [VERIFIED: scripts/verify-phase32.sh:21-27,88-171; scripts/check-freeze.sh:1-3; scripts/check-receipts.sh:1-3,75-113; .planning/phases/38-milestone-receipts/38-CONTEXT.md:35-75]

Keep the three-plan sequence: prepare and focus-check, refresh audit and freeze; stop at a concrete commit-specific remote checkpoint; then ingest one wholly successful same-attempt Ubuntu/macOS run in a final planning-only plan. Copy both raw logs and bounded provenance into the phase directory and verify copied bytes before binding EVD-17. This research ran no tests or hosted actions. [VERIFIED: .planning/ROADMAP.md:346-364; .planning/REQUIREMENTS.md:84,128; 38-CONTEXT.md:62-83]

**Primary recommendation:** Preserve M006 defaults, use one explicit M007 manifest, require complete reached-case/control inventories on both hosts, retain existing full/race evidence jobs, and validate external provenance separately from receipt consistency.

## Architectural Responsibility Map

| Capability | Primary tier | Secondary tier | Rationale |
|------------|--------------|----------------|-----------|
| Public CLI and built-program tuples | CLI / backend | Native runtime | Host producer invokes applications and compares stdout/stderr/exit tuples. [VERIFIED: scripts/verify-phase32.sh:88-137] |
| Manifest, selector and completion inventory | Host driver | Test harness | Driver owns selected cases and controls; it must reject empty, skipped or partial evidence. [VERIFIED: scripts/verify-phase32.sh:109-171; 38-CONTEXT.md:51-57] |
| Digest, freeze and receipt consistency | Gate scripts | Git | Existing scripts own digest paths, ancestry, row and log consistency. [VERIFIED: scripts/codedigest.sh:48-64; scripts/check-freeze.sh:1-3; scripts/check-receipts.sh:1-113] |
| Hosted jobs and run provenance | GitHub Actions | Binding checker | Existing workflow owns two-host jobs/artifacts; new narrow seam binds run identity to freeze. [VERIFIED: .github/workflows/ci.yml:51-145; CITED: https://docs.github.com/en/rest/actions/workflow-runs] |
| Durable receipt and EVD-17 | Versioned .planning files | Receipt gate | Last plan copies raw logs and bounded metadata and validates retained bytes. [VERIFIED: 38-CONTEXT.md:67-83] |

## Standard Stack

### Core

| Tool | Version | Purpose |
|------|---------|---------|
| Go standard library | Project declares Go 1.24 | Compiler tests and structured data; standard-library-first is a project constraint. [VERIFIED: go.mod:1-3; AGENTS.md:1-32; 38-CONTEXT.md:38-43] |
| POSIX shell / Bash | Local shell; workflow uses Bash for receipt steps | Existing host producer and CI orchestration. [VERIFIED: scripts/verify-phase32.sh:1-27; .github/workflows/ci.yml:120-145] |
| Git and codedigest | Local Git 2.41.0 | Freeze ancestry, clean worktree and digest boundary. [VERIFIED: local probe; scripts/codedigest.sh:13,48-64; 38-CONTEXT.md:58-64] |
| Clang | Local Apple clang 21.0.1 | Native C build; CI provisions/checks Clang. [VERIFIED: local probe; AGENTS.md:1-32; .github/workflows/ci.yml:74-83,108-116] |

No package install is needed. [VERIFIED: 38-CONTEXT.md:38-43]

### Supporting

Official GitHub run, attempt-specific job, artifact, and rerun documentation informs the bounded provenance projection. Primary-document lookup and read-only remote/ref inspection are allowed before the D-10 checkpoint; remote writes remain gated. Local gh 2.101.0 exists, but auth and remote state were not probed. [CITED: https://docs.github.com/en/rest/actions/workflow-runs; https://docs.github.com/en/rest/actions/workflow-jobs; https://docs.github.com/en/rest/actions/artifacts; https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs?tool=cli; VERIFIED: local probe]

### Alternatives Considered

| Prefer | Avoid | Tradeoff |
|--------|-------|----------|
| Extend maintained driver in opt-in mode | New generic M007 runner | Avoid duplicating host identity, tuple and selector logic. [VERIFIED: scripts/verify-phase32.sh:49-68,109-171; 38-CONTEXT.md:38-43] |
| Exact run/attempt metadata plus raw logs | Green badge or artifact URL alone | Badge/URL does not provide durable per-case bytes bound to freeze. [CITED: GitHub API docs above; VERIFIED: 38-CONTEXT.md:35-43,65-75] |
| Existing SHA-256 and gates | Signing framework | Signing is outside scope. [VERIFIED: scripts/codedigest.sh:48-64; 38-CONTEXT.md:35-43,82-83] |

## Package Legitimacy Audit

Not applicable: no external package is added. [VERIFIED: 38-CONTEXT.md:38-43]

## Architecture Patterns

### Data flow

~~~mermaid
flowchart LR
  M[Explicit M007 manifest] --> D[Opt-in maintained driver]
  D --> C[Program tuples and named controls]
  C --> I[Complete observed inventory]
  I --> L[Per-host raw log and artifact]
  L --> W[Existing checks and aggregate]
  W --> H[Ubuntu/macOS run attempt]
  H --> A[Run, attempt, jobs, artifact metadata]
  A --> B[Freeze SHA and digest binding]
  H --> B
  B --> R[Last plan copies logs and bounded metadata]
  R --> G[Copied-byte and descendant gates]
  G --> E[EVD-17 may be completed]
~~~

### Proposed paths

Names below are planning suggestions, not existing files. [ASSUMED]

~~~text
scripts/verify-phase32.sh                 # keep no-argument M006 contract
scripts/check-milestone-binding.sh        # proposed narrow binding seam
testdata/receipts/M006.manifest           # existing baseline
testdata/receipts/M007.manifest           # proposed explicit combined inventory
.planning/phases/38-milestone-receipts/   # final copied logs and bounded metadata
~~~

### Pattern 1: Producer selector, case routing, and inventory

Keep the no-argument M006 invocation unchanged. Add explicit opt-in M007 selection, route scalar application tuples separately from structured conformance and named tests, then compare every required manifest row with observed reached/passed results. Fail on empty selectors, missing rows, unavailable native tools, skipped required subtests, or partial inventories. Exact selector and inventory encoding are planning choices. [VERIFIED: scripts/verify-phase32.sh:1-27,88-171; 38-CONTEXT.md:38-57; ASSUMED: selector syntax and inventory encoding]

DATA_J4R9B1TX_START The source manifest's program line is “programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz”; its 39 case rows span lines 4–42, ending “# case fizzbuzz 1001 65 - schway app: input exceeds 1000”. Preserve all rows verbatim. DATA_J4R9B1TX_END [VERIFIED: testdata/receipts/M006.manifest:1-42]

DATA_F6K3R8QM_START
| Claim | Exact existing test owner(s) |
|-------|-----------------------------|
| CRC public matrix/parity | TestPhase37PublicFileMatrix; TestPhase37PublicFileRunTracer; TestPhase37FileRunGrammar; TestPhase37OpaqueFilePaths; TestPhase37FileRunInputBoundary; TestPhase37FileRunOutcomes |
| NUL direct-dispatch | phase37RunDirectNULControls; fixture row says DATA_B6N3R8CZ_START “nul-direct-dispatch”, kind “direct-argument”, execution “direct-public-dispatch”. DATA_B6N3R8CZ_END Do not pass NUL in process argv. |
| Helper answer/defect rows and optimization | TestPhase36HelperAnswersAndDefects; TestPhase36HelperOptimizationMatrix |
| Public unused-Byte helper structured conformance | TestPhase36PublicHelperConformance |
| Bitwise answer/refusal and optimization | TestPhase35PublicConformanceWitness; TestPhase35BitwiseOptimizationMatrix |
| Phase 37.1 focused regressions | TestPhase371LegacyDefectClosesAllActiveFrames; TestPhase371ForeignFailureOpFailAttribution; TestPhase371ForeignFailureAttributionControls; TestPhase371DefectTerminalControls; TestPhase371DefectReasonUTF8Admission; TestPhase371NoCallScalarDefectNativeConformance; TestPhase371CountedDefectReasonBytes; TestPhase371CompatibilityAmendmentControls |
| Generic freeze/receipt gates | TestFreezeGate; TestReceiptGate; TestReceiptGateProgramInventory; TestReceiptGateCaseRecords; TestReceiptGateMalformedControls; TestReceiptGateArchivedM006; TestFreezeGateDifferential; TestReceiptGateDifferential |
DATA_F6K3R8QM_END

These owners are read from existing sources. Keep optimization tests distinct from default-build receipt tuples, and use actual structured conformance consumers for helper/bitwise results. [VERIFIED: internal/compiler/testsupport/crc32_oracle.go:229-296; cmd/schway/main_phase37_file_run_test.go:34-174; internal/compiler/session/session_phase36_calls_test.go:25-206; cmd/schway/main_phase36_test.go:20-85; cmd/schway/main_phase35_test.go:16-45; internal/compiler/session/session_phase35_values_test.go:57-117; internal/compiler/executionpeer/executionpeer_phase371_test.go:53-80,158-220,358-430,640-710; internal/compiler/session/session_phase371_defect_reason_test.go:110-200; internal/compiler/cgen/cgen_phase371_legacy_defect_test.go:22-180; internal/compiler/session/session_phase371_compatibility_test.go:985-1065; internal/compiler/session/session_freeze_gate_test.go:23-160; internal/compiler/session/session_receipt_gate_test.go:21-410,836-900; internal/compiler/session/session_gate_differential_test.go:35-180]

CRC fixture schema is DATA_R2Y8K5MF_START “schway.crc32-cases/1”; consumers are “crc32”, “crc_low”, “checksum”, “adler_low”; cases include “cap-plus-one”, “opaque-json”, “path-4097”, and “nul-direct-dispatch”. DATA_R2Y8K5MF_END [VERIFIED: examples/crc32/crc32.expected.json:2,12-17,38-65] The helper fixture schema is DATA_P9C4V7X2_START “schway.phase36.helper-values/1”. DATA_P9C4V7X2_END [VERIFIED: testdata/phase36/helper-values.json:2,9-24]

### Pattern 2: Bind provenance separately from consistency

Existing gates check digest/ancestry and receipt row/log consistency; the receipt checker only requires a nonempty headsha. Add a narrow binding check from selected run to freeze SHA and both log headsha values, including attempt, job and artifact membership. Include an armed wrong-revision negative control. [VERIFIED: scripts/check-freeze.sh:1-3; scripts/check-receipts.sh:1-113; 38-CONTEXT.md:58-75]

GitHub run metadata exposes SHA/attempt, jobs can be scoped to a run attempt, and artifacts expose workflow-run identity and digest. A rerun has a distinct attempt; never cherry-pick host evidence across attempts. [CITED: https://docs.github.com/en/rest/actions/workflow-runs; https://docs.github.com/en/rest/actions/workflow-jobs; https://docs.github.com/en/rest/actions/artifacts; https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs?tool=cli]

Retain only the bounded D-12 fields: repo, workflow/event, run ID/attempt/URL, ref/head SHA/conclusion, required job IDs/conclusions, artifact names/IDs/digests/expiry, and raw log paths/byte counts/SHA-256. Exclude credentials, environment dumps and unrelated API profile fields. [VERIFIED: 38-CONTEXT.md:65-75]

### Pattern 3: Keep evidence contracts separate

Default-build application tuples, interpreter/native public-route parity, O0/O3/O3-LTO tests, structured conformance, causal defect attribution, and physical cleanup answer different questions. TestPhase37CRC32PhysicalCleanup owns physical cleanup evidence; a successful process tuple is not a cleanup claim. [VERIFIED: scripts/verify-phase32.sh:109-137; cmd/schway/main_phase37_file_run_test.go:34-174; cmd/schway/main_phase36_test.go:20-85; internal/compiler/session/session_phase35_values_test.go:57-117; internal/compiler/native/native_phase37_filebytes_test.go:238-323; 38-CONTEXT.md:45-49]

### Anti-Patterns

- Green badge, program list or artifact URL in place of reached cases and durable bytes. [VERIFIED: 38-CONTEXT.md:35-43,65-75]
- Consistency mistaken for hosted provenance authentication. [VERIFIED: scripts/check-receipts.sh:1-3,75-113]
- Counting a skipped FIFO/device/symlink/unreadable row as a pass. [VERIFIED: internal/compiler/testsupport/crc32_oracle.go:229-296]
- Changing M006 defaults, cloning gates, or sharing an output contract across scalar apps and structured conformance. [VERIFIED: scripts/verify-phase32.sh:1-27,109-171; 38-CONTEXT.md:38-49]
- Receipt pipeline through tee without pipefail. Bash uses the final command status by default; enable pipefail while retaining upload-on-failure and missing-file failure. [CITED: https://www.gnu.org/software/bash/manual/html_node/Pipelines; VERIFIED: .github/workflows/ci.yml:120-145]
- Freezing before all digest inputs are committed and freshly audited, or keeping receipts after code changes. [VERIFIED: scripts/codedigest.sh:48-64; 38-CONTEXT.md:58-75]
- Completing EVD-17 before final same-attempt evidence and planning-only drift checks. [VERIFIED: 38-CONTEXT.md:65-83; .planning/REQUIREMENTS.md:84,128]

## Code Examples

Use the attempt-specific jobs endpoint for the selected run attempt; compare the run head SHA with the freeze SHA and each copied host log headsha before the final binding. [CITED: https://docs.github.com/en/rest/actions/workflow-runs; https://docs.github.com/en/rest/actions/workflow-jobs]

~~~text
GET /repos/{owner}/{repo}/actions/runs/{run_id}/attempts/{attempt_number}/jobs

run.head_sha == freeze.commit_sha
host_logs["ubuntu"].headsha == run.head_sha
host_logs["macos"].headsha == run.head_sha
~~~

The endpoint is documented by GitHub; the local comparisons are a proposed project check, not existing implementation. [CITED: https://docs.github.com/en/rest/actions/workflow-jobs; ASSUMED]

## Don't Hand-Roll

| Problem | Reuse | Why |
|---------|-------|-----|
| Digest/ancestry/receipt rules | scripts/codedigest.sh, scripts/check-freeze.sh, scripts/check-receipts.sh | Preserve existing scope/refusal behavior; add only clean-worktree and remote identity checks. [VERIFIED: cited scripts above] |
| CRC answers | examples/crc32/crc32.expected.json | Independent expected answers and case routes. [VERIFIED: examples/crc32/crc32.expected.json:4-28,38-70] |
| Helper answers | testdata/phase36/helper-values.json | Pinned answer/defect rows. [VERIFIED: testdata/phase36/helper-values.json:2,9-24] |
| Workflow provenance | GitHub run, attempt jobs and artifact APIs | They expose the identity fields needed for bounded binding. [CITED: GitHub REST docs above] |
| General runner framework | Producer-native manifest/observed inventory check | Only one concrete manifest and explicit skip/missing refusal are needed. [VERIFIED: 38-CONTEXT.md:38-57] |

**Key insight:** Accepted evidence needs three independent facts: reached cases/controls, exact source/attempt identity, and final copied-byte integrity. None substitutes for the others. [VERIFIED: 38-CONTEXT.md:35-83; scripts/check-receipts.sh:75-113]

## Common Pitfalls

1. **Consistency mistaken for provenance:** compare selected run head SHA, freeze SHA and both logs; test wrong revision. [VERIFIED: scripts/check-freeze.sh:1-3; scripts/check-receipts.sh:75-113]
2. **Skipped host cases:** the CRC materializer has environment skip paths. A required row without reached evidence blocks acceptance, regardless of package exit code. [VERIFIED: internal/compiler/testsupport/crc32_oracle.go:229-296]
3. **tee masks a failure:** use Bash pipefail and preserve upload-on-failure. [CITED: GNU Bash Pipelines; VERIFIED: .github/workflows/ci.yml:120-145]
4. **M006 contract regresses:** current manifest has DATA_D8L1M5QX_START “label=phase32” DATA_D8L1M5QX_END and the seven-program list quoted above; preserve its no-argument behavior. [VERIFIED: testdata/receipts/M006.manifest:1-3]
5. **Evidence claims are conflated:** route, structured conformance, optimization and cleanup need their named consumers. [VERIFIED: named tests above; 38-CONTEXT.md:45-49]
6. **Freeze/status advanced early:** finish code/CI/tests, focused validation, review and fresh audit before actual freeze; bind only in final planning-only plan. [VERIFIED: scripts/codedigest.sh:48-64; 38-CONTEXT.md:58-83]

## State of the Art

| Prior pattern | Required pattern | Impact |
|---------------|------------------|--------|
| Program list or green check | Reached manifest rows and controls | Empty, skipped, missing and partial evidence refuse. [VERIFIED: 38-CONTEXT.md:35-57] |
| Consistency treated as provenance | Exact run/attempt/freeze binding | Wrong-source run is rejected. [VERIFIED: scripts/check-receipts.sh:1-3,75-113; 38-CONTEXT.md:58-75] |
| Artifact URL as durable evidence | Copy raw logs and bounded metadata under .planning | Final gate checks retained bytes. [VERIFIED: 38-CONTEXT.md:67-83] |

Signing, hermetic builds, a new backend and a general evidence framework are out of scope. [VERIFIED: 38-CONTEXT.md:35-43,82-83]

## Project Constraints (from AGENTS.md)

- Preserve Go 1.24 standard-library bootstrap, readable C17/installed Clang, shallow dependencies, deterministic evidence lanes, and explicit security/resource boundaries. [VERIFIED: AGENTS.md:1-32]
- Make edits within GSD workflow entry points; direct edits outside workflow require user bypass. [VERIFIED: AGENTS.md:70-83]
- At phase/milestone transitions, compare PRODUCT-ROADMAP and LANGUAGE-MATURITY to source/refusal/test witnesses; scan relevant seed triggers, consult the Dashbit appraisal and record scoped dispositions. [VERIFIED: AGENTS.md:89-108]
- Surface three next capabilities with program, blocker, smallest slice, guarantee changes, evidence/debt, owner and reprioritizing observation. Update living docs only when facts change, inside workflow; preserve archived decisions except dated amendments. [VERIFIED: AGENTS.md:108-118]
- Prefer runnable gains; a second consecutive enabling phase triggers scope review. Plan growth needs a named witness/safety duty; keep negative controls tied to the capability and avoid invented completeness metrics or generic architecture. [VERIFIED: AGENTS.md:118-126]
- Requirement ownership remains in REQUIREMENTS.md and ROADMAP.md. [VERIFIED: AGENTS.md:126-128]

## Assumptions Log

| # | Claim | Risk if wrong |
|---|-------|---------------|
| A1 | Use an opt-in manifest selector while preserving no-argument M006. | Selector syntax is discretionary; default behavior is locked. |
| A2 | A narrow binding checker and wrong-revision test are the smallest external provenance seam. | Existing interface may already provide this seam. |
| A3 | New metadata/log filenames and bounded inventory encoding can be chosen in planning. | Names/schema are not locked. |
| A4 | A small producer-native inventory helper can refuse skips without becoming a general runner. | Shell-driver boundary may require a different local design. |
| A5 | Local probes say nothing about hosted behavior or authentication. | Hosted tools/results may block completion. |

## Open Questions

1. **RESOLVED — fail-closed execution policy, availability unobserved.** Can both hosted OSes execute all environment-dependent CRC rows without skips? Plan01's `TestPhase38ProducerInventoryControls` and `TestPhase38ManifestConsumerBindings` require the exact reached inventory, and Plan02 task38-02-02 rejects skipped/missing permission/FIFO/device/symlink controls. An unavailable required host row is a stated external blocker with EVD-17 Pending; this resolution does not assert either host can run it. [SOURCE: internal/compiler/testsupport/crc32_oracle.go:229-296; 38-01-PLAN.md; 38-02-PLAN.md]
2. **RESOLVED — concrete schema and checker selected.** The freeze record is `.planning/phases/38-milestone-receipts/38-FREEZE.json`, schema `schway.m007-freeze/1`; final run provenance is `38-HOST-RUN.json` in that directory, schema `schway.m007-hosted-run/1`. Plan01's Exact interfaces and ownership section fixes the bounded fields, full SHA/digest strings, exact integer identities, artifact/log distinctions and ingest_base_sha. `scripts/check-milestone-binding.sh` exposes freeze/run/final modes, using installed Python3 standard-library parsing and delegating the existing generic gates. `TestPhase38BindingControls` arms positive, wrong-revision/attempt/identity/bytes and precision refusals; `TestPhase38FreezeWorktreeControls` owns source/worktree/descendant controls in the new `internal/compiler/session/session_phase38_receipts_test.go`. These are planning choices; implementation and passing evidence remain future execution. [SOURCE: 38-CONTEXT.md D-12; 38-01-PLAN.md; 38-03-PLAN.md]
3. **RESOLVED — explicit authority and result gates, access unobserved.** Will authorized remote access and a complete run be available? Plan02 task38-02-01 performs read-only inspection and requires non-autoapproved human authorization naming the actual freeze SHA/digest/repository/ref/workflow/inputs/push scope. Task38-02-02 then requires one complete successful same-source run/attempt with both hosts and all required jobs/artifacts/controls; one diagnosed transient same-freeze retry is the limit. Missing access, approval or required results leaves EVD-17 Pending with its external blocker, and any code fix reopens audited freeze and approval. No current remote availability or human approval is asserted. [SOURCE: 38-CONTEXT.md D-10..D-15; 38-02-PLAN.md]

## Environment Availability

| Dependency | Required by | Available here | Version / fallback |
|------------|-------------|---------------|--------------------|
| Go | Focused checks and producer | Yes | go1.24.0 darwin/arm64 |
| Clang | Native build | Yes | Apple clang 21.0.1 |
| Shell/Bash | Producer and CI | Yes | /bin/sh and Bash; use Bash pipefail |
| Git | Freeze/digest | Yes | git 2.41.0 |
| GitHub CLI/API | Read-only inspection before checkpoint; approved writes afterward | CLI present; auth/remote unobserved | gh 2.101.0; unavailable authority/results mean pending blocker |
| Python | Existing fixture provenance and planned narrow binding JSON parser | Locally observed; hosted availability checked during execution | Python3 standard library and installed tool only; no package install; unavailable tool fails closed |

Fresh local probes only. No tests or compiler/native suite were run; these probes do not establish hosted availability or authentication. [VERIFIED: local command probes]

## Validation Architecture

Nyquist validation is enabled: DATA_C7F2W9NB_START “nyquist_validation”: true DATA_C7F2W9NB_END. [VERIFIED: .planning/config.json:26]

### Test Framework

| Property | Value |
|----------|-------|
| Framework | Go test; go.mod declares DATA_Z2F6J8Q1_START “go 1.24” DATA_Z2F6J8Q1_END. [VERIFIED: go.mod:1-3] |
| Local checks | Focused selectors only; proposed, not run. [ASSUMED] |
| Hosted full suite | DATA_A4P7L9WX_START “go test ./...” and “go test -race ./...” DATA_A4P7L9WX_END on both hosts. [VERIFIED: .github/workflows/ci.yml:74-83,104-116] |

### Requirement-to-test map

| ID | Behavior | Test/evidence |
|----|----------|---------------|
| EVD-17 | Every M007/inherited case/control runs on both hosts, and one complete attempt binds the audited freeze plus copied bytes | Existing full/race jobs, M007 producer/artifacts, run/attempt binding and final copied-byte gate. New producer, manifest and wrong-revision control are plan work. [VERIFIED: cited sources above; ASSUMED: additions] |

### Focused preparation checks

- Verify no-argument M006 selection is preserved and explicit M007 inventory contains every inherited and new row. [ASSUMED]
- Refuse empty selection, missing rows, partial inventory, unavailable native tools and skipped required subtests at the producer seam. [ASSUMED]
- Run existing freeze/receipt acceptance/refusal owners and the binding positive/wrong-revision controls. Existing roots include TestFreezeGate, TestReceiptGate, TestReceiptGateProgramInventory, TestReceiptGateCaseRecords, TestReceiptGateMalformedControls, TestReceiptGateArchivedM006, TestFreezeGateDifferential and TestReceiptGateDifferential. [VERIFIED: internal/compiler/session/session_freeze_gate_test.go:23-160; internal/compiler/session/session_receipt_gate_test.go:21-410,836-900; internal/compiler/session/session_gate_differential_test.go:35-180]
- Include named route/helper/bitwise/Phase 37.1 control owners from Pattern 1 in hosted producer. Do not run full local suite, whole-package sweep, broad matrix or capture-pressure experiment. [VERIFIED: 38-CONTEXT.md:45-49]

Proposed low-cost command groups for planning (not executed during research):

~~~sh
go test ./internal/compiler/session -run 'TestFreezeGate|TestReceiptGate' -count=1
go test ./cmd/schway -run 'TestPhase37PublicFileRun|TestPhase36PublicHelperConformance|TestPhase35PublicConformanceWitness' -count=1
go test ./internal/compiler/session -run 'TestPhase36Helper|TestPhase35BitwiseOptimizationMatrix' -count=1
go test ./internal/compiler/executionpeer -run 'TestPhase371' -count=1
go test ./internal/compiler/cgen -run 'TestPhase371NoCallScalarDefectNativeConformance|TestPhase371CountedDefectReasonBytes' -count=1
go test ./internal/compiler/native -run 'TestPhase37CRC32PhysicalCleanup' -count=1
go test ./internal/compiler/testsupport -run 'TestPhase37PublicFileMatrix' -count=1
~~~

These selectors target named owners only; new producer inventory and binding controls still need focused tests in their chosen packages. [ASSUMED]

### Hosted checks and Wave 0 gaps

Require current checks and current evidence aggregate on Ubuntu/macOS, full and race lanes, M007 artifacts, complete inventory, and exact run/attempt/source metadata. [VERIFIED: .github/workflows/ci.yml:51-145; 38-CONTEXT.md:45-75]

- [ ] M007 manifest and opt-in selector preserving M006 default. [ASSUMED]
- [ ] Observed inventory and required-skip refusal. [ASSUMED]
- [ ] Exact run/attempt/freeze binding and wrong-revision control. [ASSUMED]
- [ ] Per-host artifact wiring and pipefail. [ASSUMED]
- No framework or external package needed. [VERIFIED: 38-CONTEXT.md:38-43]

Preserve the preexisting modified .planning/config.json byte-for-byte; its observed SHA-256 was 148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521. Exclude it from staging. [VERIFIED: local hash probe and git status]

## Security Domain

Security enforcement is enabled at ASVS level 1. [VERIFIED: .planning/config.json:49-50]

| Category | Applies | Phase control |
|----------|---------|---------------|
| V2 Authentication | Remote authorization boundary | Stop at concrete D-10 checkpoint before push/dispatch. [VERIFIED: 38-CONTEXT.md:55-60] |
| V3 Session Management | No app session surface | This phase adds no app login/session behavior. [VERIFIED: .planning/ROADMAP.md:346-364] |
| V4 Access Control | Repo/workflow scope | Bind repository, workflow, ref, freeze SHA, attempt and required jobs; do not broaden action. [VERIFIED: 38-CONTEXT.md:55-75] |
| V5 Input Validation | Yes | Parse manifest/API fields as bounded data; reject malformed/missing/extra rows; never shell-evaluate metadata. [VERIFIED: 38-CONTEXT.md:51-75] |
| V6 Cryptography | Integrity only | Reuse SHA-256/codedigest; do not add signing or claim digest authenticity. [VERIFIED: scripts/codedigest.sh:48-64; 38-CONTEXT.md:58-83] |

Threats are wrong-run/revision binding, untrusted metadata reaching shell evaluation, skipped/partial evidence reported as complete, and artifact expiry. Use exact identity comparisons, constrained parsing, hard inventory checks, and retained raw bytes. This is not a general security audit or hermeticity claim. [VERIFIED: 38-CONTEXT.md:51-83]

## Transition priorities and seed dispositions

The current useful priorities are summarized below; Phase 38 owns only the first. [VERIFIED: .planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-TRANSITION.md:7-21; .planning/PRODUCT-ROADMAP.md:1131-1139; .planning/LANGUAGE-MATURITY.md:1021-1029]

DATA_M7Q2V9KC_START
| User-visible program/capability | Current blocker | Smallest complete slice | Checker / guarantee changes | Evidence / debt | Owner / next action | Observation that changes priority |
|---|---|---|---|---|---|---|
| Audited crc32/crc_low/checksum and earlier pinned programs on Ubuntu/macOS | M007 producer coverage and final hosted binding | Prepare, audit, freeze, obtain and ingest one complete two-host run | No language rule; require reached cases, exact provenance and source-drift rejection | EVD-17 Pending; fresh terminal audit; historical M006 logs | Phase 38 preparation, then commit-specific remote checkpoint | Missing case, host disagreement, regression or source drift postpones binding |
| Proposed M008 “wc FILE” / “longest_line FILE” sharing | Exact contract and justified argument/record/export laws unresolved | One bounded count/longest-line pair with a shared library | Consumer-owned laws only if the program needs them | Living roadmaps; existing ByteView/modules/FileBytes | Next milestone language/checker owner specifies before promotion | Existing scalar interface suffices, or a smaller concrete consumer appears |
| Bounded fixed-schema key=value checker | Schema, duplicate and error behavior unspecified | One bounded schema with literal positive/negative fixtures | Bounded parsing and attributable errors; representation only as needed | Living roadmaps and historical source appraisal | Future consumer owner supplies the contract | A concrete JSON consumer provides a smaller justified slice |
DATA_M7Q2V9KC_END

| Seed | Source witness / consumer | Disposition |
|------|---------------------------|-------------|
| z1g | QueryCommandFile and ordinary search locate current consumers | Retain dormant; no measured navigation gap |
| trz | readApplicationCapture, replay and named CRC/checksum controls explain failures | Retain dormant; no unexplained failure |
| z1s | syntax.Format and checked CRC/helper sources; no contested form or reviewer experiment | Retain dormant; no trigger |

Seed metadata and the historical Dashbit appraisal were inspected; no experiment or broader system is promoted. [VERIFIED: .planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-TRANSITION.md:13-21; .planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md:1-7; .planning/seeds/SEED-261008-trz-consumer-triggered-failure-observability.md:1-7; .planning/seeds/SEED-261008-z1s-compare-contract-and-source-forms.md:1-7; .planning/sources/DASHBIT-AI-ERA.md:11,49-70]

## Sources

### Primary — repository source of truth

Context, ROADMAP, REQUIREMENTS, audit and Phase 37.1 transition/closeout; AGENTS and living roadmap files; producer/digest/freeze/receipt scripts and CI workflow; M006 manifest, CRC/helper/compatibility fixtures; named owners in Pattern 1.

### Secondary — MEDIUM confidence

Official GitHub docs: [workflow runs](https://docs.github.com/en/rest/actions/workflow-runs), [workflow jobs](https://docs.github.com/en/rest/actions/workflow-jobs), [artifacts](https://docs.github.com/en/rest/actions/artifacts), [reruns](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs?tool=cli). GNU Bash: [Pipelines](https://www.gnu.org/software/bash/manual/html_node/Pipelines).

### Tertiary — LOW confidence

None used as recommendations. Discretionary choices are marked assumed.

## Metadata

**Confidence breakdown:** Standard stack HIGH (source inspected, no package); architecture HIGH for existing owners/gates and MEDIUM for proposed selector/binding seam; pitfalls HIGH for skip/gate/CI boundaries and MEDIUM for unobserved hosted behavior.

**Research date:** 2026-10-16  
**Valid until:** 2026-11-15 for repository architecture; re-check GitHub API details before implementation.

