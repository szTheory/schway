# Phase 38: Milestone Receipts — Context

**Gathered:** 2026-10-16
**Status:** Ready for planning
**Decision authority:** Carry forward the user's explicit standing request for all relevant areas, specialist fan-out, adversarial synthesis, primary sources when useful, and automatic adoption of coherent recommendations. Three typed advisor researchers completed the bounded read-only fan-out. These are delegated decisions, not individually answered human questions. Persistent `workflow.auto_advance: true` advances to planning/execution. The commit-specific push/dispatch checkpoint remains required.

<domain>
## Phase Boundary

Close EVD-17 by binding hosted Ubuntu and macOS evidence for M007 and the earlier pinned native programs to the final audited M007 source. Finish the necessary manifest/producer/CI preparation before selecting the freeze. Reuse the Phase 33 generic freeze and receipt gates. The final ingest plan modifies only `.planning/`. Earlier completed plans, UAT, original goldens and receipts retain their scope and history. No new language surface, backend, dependency, signing system, or general evidence framework belongs here.
</domain>

<decisions>
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

### Agent discretion

Choose exact helper/file boundaries, manifest inventory encoding, bounded metadata projection and focused test arrangement during research/planning. Preserve the decisions above, existing gate refusal behavior and M006 default producer contract. Resolve routine implementation details autonomously. Remote authorization is reserved for the concrete D-10 checkpoint.
</decisions>

<canonical_refs>
## Canonical References

Downstream agents must read these relevant sections before planning or implementing.

### Ownership, audit and constraints

- `.planning/ROADMAP.md` — Phase 38 goal, three success criteria and explicit remote checkpoint.
- `.planning/REQUIREMENTS.md` — EVD-17 and unchanged ownership.
- `.planning/PROJECT.md` — verification preference, hosted blockers, current M007 constraints and dated closeout.
- `.planning/STATE.md` — current position and no-full-local-suite instruction.
- `.planning/M007-MILESTONE-AUDIT.md` — terminal closure passed; only EVD-17 remains open.
- `.planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-CLOSEOUT.md` and `.planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-VERIFICATION.md` — completed closure evidence and its limits.
- `.planning/phases/37.1-close-gap-legacy-terminal-outcomes/37.1-TRANSITION.md` — next capabilities and seed dispositions.
- `.planning/PRODUCT-ROADMAP.md` and `.planning/LANGUAGE-MATURITY.md` — current completion amendments and future consumer ownership.

### Reusable acceptance mechanisms and exact fixtures

- `scripts/codedigest.sh`, `scripts/check-freeze.sh`, `scripts/check-receipts.sh` — digest, ancestry, receipt consistency and their explicit non-authenticity boundary.
- `scripts/verify-phase32.sh` and `testdata/receipts/M006.manifest` — existing host producer and 39 earlier pinned rows.
- `.github/workflows/ci.yml` — existing two-host checks/evidence matrices and artifact upload.
- `internal/compiler/session/session_freeze_gate_test.go`, `internal/compiler/session/session_receipt_gate_test.go`, `internal/compiler/session/session_gate_differential_test.go` — generic gate acceptance/refusal controls and planning-only descendants.
- `examples/crc32/crc32.expected.json`, `examples/crc32/README.md`, `examples/crc32/file_bytes.bindings.json` — CRC consumers, independent answers, grammar and file cases.
- `testdata/phase36/helper-values.json`, `examples/helper_calls/`, `examples/bitwise.schway` — helper and bitwise witnesses; follow the exact fixture's selected entry/route.
- `cmd/schway/main_phase37_file_run_test.go` — named public-route grammar/input/outcome tests; do not infer a route from a fixture name.
- `testdata/phase371/compatibility-amendments.json` and the Phase 37.1 verification's named roots — scoped terminal repair controls and preserved compatibility history.

### Seed review and primary documentation

- `.planning/seeds/SEED-261008-z1g-bounded-symbol-and-caller-query.md`, `.planning/seeds/SEED-261008-trz-consumer-triggered-failure-observability.md`, `.planning/seeds/SEED-261008-z1s-compare-contract-and-source-forms.md` — trigger metadata, no promotion.
- `.planning/sources/DASHBIT-AI-ERA.md` — historical appraisal; no repeated essay study or new system follows.
- [GitHub workflow-run API](https://docs.github.com/en/rest/actions/workflow-runs), [workflow-job API](https://docs.github.com/en/rest/actions/workflow-jobs), [artifact API](https://docs.github.com/en/rest/actions/artifacts), and [rerun semantics](https://docs.github.com/en/actions/how-tos/manage-workflow-runs/re-run-workflows-and-jobs) — primary guidance on head/attempt/job/artifact identity. These inform provenance checks; no new external specification overrides the project contract.
</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable assets

The generic shell gates already take a manifest; the M006 producer builds actual public CLI applications, compares exact stdout/stderr/exit tuples, checks named non-skipped controls and records host/toolchain/source identity. Its current program loop assumes application arguments, so new conformance/file routes require explicit handling. Existing CRC/helper expected data supplies independent answers without deriving them from the engine being checked.

### Established patterns and integration

The codedigest includes scripts, tests, fixtures and CI, while `.planning` is outside it. Freeze acceptance permits planning-only descendants. The receipt gate validates complete row inventories but currently only requires `headsha` to be nonempty; exact GitHub run/log/freeze equality needs its own checked seam and negative control. CI currently uploads Phase 27/32 receipts, so full-suite success alone does not provide the missing M007 artifact. Artifact uploads use `always()` and retain failures; preserve that behavior.

No codebase maps, Phase 38 SPEC/context/plans/checkpoint, matching pending todo, advisor profile or packaged spike/sketch findings existed at entry. The three most recent contexts (37.1, 37, 36) carry the standing delegation, scoped-local-check rule and final hosted ownership.
</code_context>

<specifics>
## Evidence and Transition Review

This discussion inspected source and existing reports; it ran no compiler/native behavior suite and no hosted action. The freeze advisor freshly computed equal codedigests at audited `320f9b5` and entry `c248ffb`, and the freeze gate passed for that historical-to-entry pair; this does not select the final freeze or cover future preparation changes. Existing verification/audit receipts remain historical at their recorded source. GitHub primary docs were consulted for metadata and retry semantics.

| Next useful program | Current blocker | Smallest complete slice | Checker/guarantee changes | Evidence/debt | Owner/next action | Observation changing priority |
| --- | --- | --- | --- | --- | --- | --- |
| Audited crc32/crc_low/checksum and earlier pinned programs on Ubuntu/macOS | M007 producer coverage and final hosted binding | Prepare, audit, freeze, obtain and ingest one complete two-host run | No language rule; reached cases, exact provenance and source-drift rejection | EVD-17 Pending; fresh terminal audit; historical M006 logs | Phase 38 preparation, then commit-specific remote checkpoint | Missing case, host disagreement, regression or source drift postpones binding |
| Proposed M008 wc FILE / longest_line FILE sharing a module | Exact contract and justified argument/record/export laws | One bounded count/longest-line pair with shared library | Consumer-owned laws only if the program needs them | Living roadmaps; existing ByteView/modules/FileBytes | Next milestone language/checker owner specifies before promotion | Existing scalar interface suffices, or a smaller concrete consumer appears |
| Bounded fixed-schema key=value checker | Schema/duplicates/errors not specified | One bounded schema with literal positive/negative fixtures | Bounded parsing and attributable errors; representation only as needed | Living roadmaps and source appraisal | Future consumer owner supplies contract | Concrete JSON consumer provides a smaller justified slice |

All three seed triggers were scanned. Retain z1g dormant: `QueryCommandFile` and ordinary search suffice for locating CRC/helper/receipt consumers; no measured navigation gap. Retain trz dormant: `readApplicationCapture`/application replay and named controls explain current failures; no unexplained CRC/checksum failure. Retain z1s dormant: `syntax.Format` and checked CRC/helper sources have no contested form or actual reviewer experiment. Relevant source/consumer witnesses remain those in the Phase 37.1 transition table. No experiment, future requirement or broader query/observability/source-form system is promoted. Living documents' latest facts remain accurate; this phase context records the new preparation decision and unchanged priority.
</specifics>

<deferred>
## Deferred Ideas

Arity-N, scalar records/exported types, wc/longest_line, key=value/JSON, separate compilation, module-aware queries, generic observability, signed provenance and whole-loop performance studies retain their existing future consumer triggers. No new scope was requested. Native execution remains process execution, and codedigest/host receipts do not establish hermetic builds or physical cleanup beyond their named tests.
</deferred>

---

*Phase: 38-milestone-receipts*
*Context gathered: 2026-10-16*
