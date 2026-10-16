# M007 cross-phase integration audit

**Audit point:** repository <repository-root>, branch phase26-validation, HEAD 502ab55edb501dbe2c6b3c101b45852bb7ec29bb.
**Scope:** Phases 33–38. Phase 38 is deliberately not planned or implemented yet; audit and scoped gap closure precede its hosted-receipt binding work.
**Worktree:** the only reported repository change is the pre-existing .planning/config.json modification. This audit wrote no repository files; temporary tests, overlays, logs, and this report are under /tmp.

## Integration result

The audit enumerated 26 expected wiring assertions: **22 WIRED, 3 BROKEN/BLOCKER, 1 PENDING**. The pending assertion is the intentionally deferred Phase 38/EVD-17 final hosted receipt binding. It is a milestone prerequisite, not an absent implementation that calls for a new gap phase.

The 3 blockers are still-current carried Phase 33 review findings, not new defects introduced by Phases 34–37:

- **CR-03:** executionpeer mishandles a legacy static child defect and leaves its caller activation open.
- **CR-04:** executionpeer rejects an interpreter function.failed event emitted for modeled typed failure on OpReturn.
- **CR-06:** a native no-call application scalar OpDefect produces app stderr/exit 65, which the native runner classifies as native.run_failed, not the structured defect outcome.

Newly executed targeted controls passed for the dynamic scalar-helper causal path, helper numeric/defect answers, and frozen no-call generated-C bytes. These green controls do not close the separate legacy terminal paths above. CR-06 needs an explicit dated compatibility decision and a narrow baseline amendment if fixing it changes frozen no-call C bytes; do not regenerate that baseline silently.

## Wiring inventory

HTTP/API routes and authentication are not part of this compiler/CLI milestone. The relevant public entrypoints are Schway CLI commands, source modules, and evidence scripts. There is no auth-sensitive route to audit.

| # | Expected connection | Evidence at current HEAD | Status |
|---|---|---|---|
| 1 | Captured native process output → bounded read before allocation → capacity result | internal/compiler/native/native_app.go:694-700 uses io.LimitReader(file, limit+1); focused 1 GiB sparse-file and boundary selector passed. | WIRED |
| 2 | Oversize capture → named capture-capacity outcome | Phase 33 native capture selector passed; current bounded read returns at most cap+1 for classification. | WIRED |
| 3 | Freeze checker → generic manifest/program list → generic codedigest | scripts/check-freeze.sh and scripts/codedigest.sh take generic inputs; TestFreezeGate passed. | WIRED |
| 4 | Receipt checker → manifest + two host logs → same digest/program inventory checks | scripts/check-receipts.sh consumes generic inputs; receipt-gate inventory and archived M006 selectors passed. This verifies the local gate contract, not hosted receipts. | WIRED |
| 5 | Historical producer/evidence → generic gate; phase-specific check callers retired; M006 historical records retained | Generic producer/gates are exercised; TestNoCheckPhase32Caller and archived-M006 selector passed. | WIRED |
| 6 | Captured source → diagnostic text/JSON presentation with source location and excerpt | internal/compiler/protocol/presentation.go:30-90; CLI diagnostic and imported-diagnostic selectors passed. | WIRED |
| 7 | Enriched diagnostic presentation → repair consumes authoritative spans without changing repair bytes | Repair compatibility selector passed; presentation tests confirm captured source; cmd/schway-repair JSON compatibility selector passed. | WIRED |
| 8 | Checker/admitted surface → versioned deterministic schway surface --json manifest | CLI delegates to surfacemanifest.Current at cmd/schway/main.go:140-158; surface witness selector passed. | WIRED |
| 9 | Phase 34 absent-form manifest → Phase 35 dynamic-shift refusal, closing DX-22/BIT-02 | internal/compiler/surfacemanifest/manifest.go:153+; Phase 35 surface witness and admission-refusal mutation controls passed. | WIRED |
| 10 | Bitwise U64 operations → independent checker/core/origin transfer rules and interpreter/native answers | check/scalar.go, corevalidate/corevalidate.go, originvalidate/originvalidate.go, interpreter and C lowering; bitwise optimization matrix and focused value controls passed. Executionpeer is only a causal/event-free witness, not numeric-result proof. | WIRED |
| 11 | Literal shift limits → checker/core/origin refusals, manifest registration, safe C lowering | Phase 35 surface witness and refusal controls passed; current manifest documents/refuses dynamic shifts. | WIRED |
| 12 | Hex literal parsing → exact source spelling retained by formatter and malformed/overflow refusal | Included in Phase 35 focused value/refusal selector; source spelling is retained through the formatter path. | WIRED |
| 13 | Adler-32 consumer → b << 16 optimization with existing exact outputs | examples/checksum/adler32.schway; Phase 35 Adler-answer selector passed. | WIRED |
| 14 | Local/imported scalar helper source → checking, all four U64/Bool argument/result pairs, and call lowering | Phase 36 helper-answer selector passed across the local/imported forms and all four pairs. | WIRED |
| 15 | Helper call → independent numeric engine answers, typed callee defect reasons, and dynamic causal peer attribution | Helper answer/defect selector, dynamic peer causal controls, and dynamic C defect shell controls passed. This does not validate the separate legacy CR-03/CR-04 traces below. | WIRED |
| 16 | Programs without calls → unchanged generated C bytes | TestPhase36NoCallGeneratedCByteBaseline passed. It freezes bytes, not no-call defect terminal semantics. | WIRED |
| 17 | Ordinary CRC library → first consumer calls one byte helper per byte and performs eight bit rounds | examples/crc32/lib/crc32.schway:1-35, examples/crc32/crc32.schway:3-5,46; single-copy/identity and public matrix selectors passed. | WIRED |
| 18 | Same ordinary CRC library → second consumer imports it without a copy and returns pinned answer | examples/crc32/crc_low.schway:4-6,53; public file matrix and CRC algorithm mutation controls passed. | WIRED |
| 19 | CRC adapter and independent oracle → public answer/control matrix | Shared adapter triple, Python zlib literal fixture, Go IEEE cross-check, and algorithm mutations are connected; public matrix and mutant selectors passed. | WIRED |
| 20 | Public schway run source-file route → module admission → interpreter/native engines → exact process tuple | Early source-file dispatch at cmd/schway/main.go:37-42; session route at :271-309; expected-first public file matrix passed. | WIRED |
| 21 | Public file route → one-descriptor, fixed 65,537-byte actual-read bound | internal/compiler/session/session_file_application.go:18-47, internal/compiler/interp/filebytes_reader.go:12,31-67; public matrix exercised 65,536/65,537 behavior on both engines. | WIRED |
| 22 | Tracked-only CRC guide → clean-checkout build/run path | TestPhase37GuideCleanCheckout passed. This establishes executable guide steps from tracked files; it makes no subjective readability claim. | WIRED |
| 23 | Static legacy child defect → terminal attribution closes child and caller correctly | executionpeer.go:693-700 pops only the child unless dynamic identity is set. Temp-overlay probe reproduced preorder refusal with caller still active. | **BROKEN — BLOCKER CR-03** |
| 24 | Modeled typed failure on OpReturn → executionpeer recognizes interpreter's function.failed event | Interpreter emits the typed-failure event at interp.go:2829-2834; peer only accepts OpFail at executionpeer.go:738-740. Temp-overlay probe reproduced executionpeer.unknown_kind. | **BROKEN — BLOCKER CR-04** |
| 25 | No-call native application OpDefect → structured defect terminal recognized by native runner | cgen_program.go:3008-3012 sends no-call applications through application-shell failure. Checker-admitted temp probe got exit 65/app stderr, then native.run_failed instead of ExpectDefect. | **BROKEN — BLOCKER CR-06** |
| 26 | M007 final manifest/digest → hosted Ubuntu and macOS receipts bound once after gap closure | Phase 38 directory/verification is absent by explicit ordering; no hosted dispatch or receipts were run. | **PENDING — EVD-17, expected** |

### Counts and finite scope

- Cross-phase integration assertions: 26 total; 22 WIRED, 3 BROKEN/BLOCKER, 1 expected PENDING.
- API/route coverage: 0 HTTP API routes in scope; 2 public CRC file consumers plus CLI surface/diagnostic/repair routes exercised; no orphaned M007 public route found in the inspected route inventory.
- Auth protection: not applicable to these local compiler/CLI flows.
- No whole-repository or whole-package suite was run.

## E2E flow inventory

The finite audit inventory contains 10 flows: **6 complete, 3 broken, 1 pending**.

| Flow | Trace | Status |
|---|---|---|
| Bounded native evidence capture | capture file → limit+1 reader → named oversize/capacity result → terminal/native result handling | Complete on focused bounded-read and outcome controls |
| Agent diagnostic and repair | source snapshot → checker diagnostic → text/JSON locations/snippet → repair span handling | Complete on CLI, captured-source and repair-compatibility selectors |
| Bitwise + Adler consumer | source operators/literals → checker/core/origin → interpreter/C → exact Adler outputs | Complete on focused matrix/value/Adler selectors |
| Scalar helper call | local/imported helper → checker/core/origin → interpreter/C → reasoned defects and causal peer | Complete on Phase 36 and dynamic peer/C controls; limited to declared U64/Bool helper paths |
| CRC source-file application | CLI source route → module admission/import → bounded file read → crc32 and crc_low → interpreter/native stdout/stderr/exit parity and oracle controls | Complete on public expected-first matrix and mutation controls |
| CRC guide | clean tracked checkout → documented build/run commands → executable consumer | Complete on clean-checkout guide selector |
| Static child defect attribution | child defect trace → peer activation unwind → terminal closure | Broken at executionpeer.go:693-700 (CR-03) |
| Modeled typed failure attribution | interpreter OpReturn typed failure → function.failed event → peer classifier | Broken at executionpeer.go:738-740 (CR-04) |
| Native no-call explicit defect | OpDefect → emitted C application shell → native runner terminal classification | Broken at cgen_program.go:3008-3012 (CR-06) |
| Final milestone freeze receipts | audited final M007 digest → Ubuntu/macOS receipts → generic receipt gate | Pending by design until Phase 38 |

## Blocker findings and smallest gap-closure scope

### CR-03 — legacy static child defect leaves caller active

**Current source:** internal/compiler/executionpeer/executionpeer.go:693-700. The full-stack unwind is guarded by dynamicCallIdentity; a legacy static child defect takes the single-frame pop and then fails preorder/terminal accounting while the caller remains active.

**New diagnostic evidence:** /tmp/schway-m007-audit/peer-debt.log, produced by temporary overlay /tmp/schway-m007-audit/peer_debt_test.go and /tmp/schway-m007-audit/peer-overlay.json. The targeted test intentionally exits nonzero and reports executionpeer.preorder ... activation has no terminal event for a valid child-defect trace. The same overlay repros CR-04. It is not a repository test or green phase gate.

**Smallest closure:** fix terminal attribution/unwind for the legacy child-defect trace; add a focused regression proving exactly the child and caller terminal behavior; preserve the dynamic repeated-activation causal test. A broader phase gate is unnecessary to establish this connection.

**Requirement impact:** carried semantic debt, not newly introduced by M007. Indirectly touches CALL-02's executionpeer integration surface; the new dynamic helper path passed its own selector.

### CR-04 — peer rejects modeled typed failure carried by OpReturn

**Current sources:** internal/compiler/interp/interp.go:2829-2834 emits function.failed for OpReturn with modeledFailure; internal/compiler/executionpeer/executionpeer.go:738-740 recognizes that event only for OpFail.

**New diagnostic evidence:** same overlay log as CR-03; valid typed-failure return is rejected as executionpeer.unknown_kind ... unclassified event kind or ID.

**Smallest closure:** allow peer classification of the typed-failure return form using the same reason/terminal attribution contract as the interpreter; add a paired interpreter/peer trace regression while preserving the existing OpFail path. Do not classify event presence alone as a numeric oracle.

**Requirement impact:** carried semantic debt with an indirect CALL-01/CALL-02 engine/peer integration concern. It is not a newly failing Phase 36 helper acceptance selector.

### CR-06 — no-call native OpDefect becomes generic process failure

**Current sources:** internal/compiler/cgen/cgen_program.go:3008-3012 routes application shells or programs without dynamic invocation through the app failure path; internal/compiler/native/native.go:634-639 rejects the nonzero native process as native.run_failed.

**New diagnostic evidence:** /tmp/schway-m007-audit/cgen-nocall-debt.log, through checker-admitted temporary overlay /tmp/schway-m007-audit/cgen_debt_test.go and /tmp/schway-m007-audit/cgen-overlay.json. Expected-failing ExpectDefect observes exit status 65 and schway app: no-call scalar stopped, returned as native.run_failed. /tmp/schway-m007-audit/cgen-probe-checker-refusal.log preserves an earlier inadmissible probe; it is not used as evidence for the finding.

**Smallest closure:** make the no-call native explicit-defect terminal structured and have the runner attribute the reason; pin the application-shell control. First make a dated compatibility decision: the Phase 36 no-call generated-C bytes are frozen and the current byte-baseline selector passes. If the fix requires changing those bytes, amend only the affected baseline with that dated rationale; never silently regenerate the fixture.

**Requirement impact:** carried debt, no direct M007 requirement failure demonstrated. It indirectly concerns CALL-02's no-call compatibility boundary, but the stated byte-identical generated-C criterion currently passes; it is not a callee-defect failure in the new U64/Bool helper path.

## Requirements integration map

“WIRED” means the current direct requirement path is supported by source tracing and the bounded selectors listed below. The issues column separates indirect carried debt and traceability metadata from direct behavior failure.

| Requirement | Integration path | Status | Issue / qualification |
|---|---|---|---|
| QLT-13 | native capture → bounded limit+1 read → oversize classification | WIRED | Newly executed capture bound selector passed. |
| QLT-14 | generic freeze working-directory behavior + owner/reason abort predicate + dispositions | WIRED | Generic gate and retired caller selector passed; no old phase checker caller found. |
| QLT-15 | prior triage/Phase 28 records → preserved evidence inventory | WIRED | Archived M006 receipt and generic program inventory selectors passed. |
| EVD-16 | manifest/program list → freeze/receipt scripts → codedigest | WIRED | Local generic freeze/receipt gate selectors passed. |
| DX-20 | captured source → CLI text/JSON presentation → file/line/col, excerpt/caret | WIRED | Diagnostic and imported-diagnostic selectors passed. |
| DX-21 | parser recovery → one recovery diagnostic, no cascade | WIRED | Phase 34 diagnostic selector passed. |
| DX-22 | absent-form catalog/manifest → named dynamic shift refusal added in Phase 35 | WIRED | Dynamic shift refusal mutation and manifest witnesses passed; this closes the Phase 34 pending item. |
| DX-23 | admitted/refused surface → versioned deterministic surface JSON | WIRED | Surface route/meta witness passed. |
| BIT-01 | independently derived value transfers → interpreter/native answers; executionpeer causal-only | WIRED | Bitwise optimization matrix and focused values passed. Peer is not credited with numeric result replay. |
| BIT-02 | literal shifts → values/refusals → manifest and safe C generation | WIRED | Surface witness/refusal controls passed. |
| BIT-03 | hex lexing/parsing → overflow/malformed refusal → formatter source spelling | WIRED | Focused value/refusal controls passed. |
| BIT-04 | Adler source uses b << 16 → historical pinned outputs | WIRED | Adler answers passed. |
| CALL-01 | local/imported U64/Bool helper pairs → both engines, reasoned defects, no recursion | WIRED | Phase 36 helper answers/defects passed. Existing ByteView callee-bound checks and caller-before-entry faults remain separate from U64/Bool helper defects. |
| CALL-02 | independent call value/defect checks + causal peer + unchanged no-call C bytes | WIRED | New helper and dynamic peer controls plus generated-C byte baseline passed. Indirect carried debt: CR-03/CR-04 break older peer trace forms; CR-06 breaks no-call native explicit-defect terminal attribution, not the stated C-byte criterion. |
| APP-13 | ordinary CRC module → capped file consumer → pinned CRC32/ISO-HDLC outcomes | WIRED | Public expected-first matrix, oracle and mutation controls passed. |
| MOD-07 | second consumer → same imported CRC source → pinned low result | WIRED | Shared-library import and public matrix/mutation selectors passed. |
| APP-14 | public source-file dispatch → interpreter/native → stdout/stderr/exit parity | WIRED | CLI matrix passed including boundary and error cases. |
| DX-24 | tracked-only guide commands → clean checkout execution | WIRED | Clean-checkout guide selector passed. Summary metadata omitted the ID; behavior and verification support it. |
| EVD-17 | final digest → both hosted receipt sets → one generic gate | PENDING (expected) | Phase 38 is intentionally after audit/gap closure. Milestone remains incomplete until this final binding is completed. |

**Requirements with no cross-phase wiring (self-contained or same-phase evidence):** QLT-13, QLT-14, QLT-15, DX-20, DX-21. They are locally bounded hardening/diagnostic outcomes; their evidence connects into the public CLI/evidence layer but does not consume a distinct feature-phase export. EVD-16 is integrated with the generic gate and prior milestone receipt records. The remaining implemented requirements have cross-phase links listed above.

**Metadata-only traceability note:** the three-source summary extraction recorded empty requirements_completed for every Phase 36 summary and no DX-24 completion entry in the Phase 37 summaries, although Phase 36/37 verification artifacts and current named behavior selectors support those requirements. Record as summary metadata partial; it is not evidence of a runtime break and no historical summary was rewritten.

## Newly executed checks and logs

All Go commands used GOCACHE=/tmp/ai-lang-verification-gocache, -count=1, bounded exact selectors and the listed timeout. The expected-failing overlay probes are diagnostic reproductions, not “tests passed.” Logs and exact commands:

| Result | Command | Log |
|---|---|---|
| PASS | go test ./internal/compiler/native -run '^TestPhase33CaptureReadBounded$' -count=1 -v -timeout=120s | /tmp/schway-m007-audit/capture-pass.log |
| PASS | go test ./internal/compiler/session -run '^(TestPhase35ExistingAdlerAnswers|TestPhase36HelperAnswersAndDefects)$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/adler-calls-pass.log |
| PASS | go test ./internal/compiler/session -run '^(TestFreezeGate|TestReceiptGateProgramInventory|TestReceiptGateArchivedM006|TestNoCheckPhase32Caller)$' -count=1 -v -timeout=180s | /tmp/schway-m007-audit/generic-gates-pass.log |
| PASS | go test ./cmd/schway -run '^(TestPhase34CheckDiagnosticPresentation|TestPhase34SurfaceRoute)$' -count=1 -v -timeout=180s | /tmp/schway-m007-audit/diagnostics-surface-pass.log |
| PASS | go test ./cmd/schway -run '^TestPhase37PublicFileMatrix$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/public-file-matrix-pass.log |
| PASS | go test ./cmd/schway -run '^TestPhase37GuideCleanCheckout$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/guide-pass.log |
| PASS | go test ./internal/compiler/session -run '^(TestPhase35BitwiseOptimizationMatrix|TestPhase35SurfaceManifestWitnesses|TestPhase36NoCallGeneratedCByteBaseline|TestPhase36SurfaceManifestCallWitnesses)$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/operators-surface-baseline-pass.log |
| PASS | go test ./internal/compiler/executionpeer -run '^TestPhase36CallCausalAttributionControls$' -count=1 -v -timeout=120s | /tmp/schway-m007-audit/dynamic-peer-control-pass.log |
| PASS | go test ./internal/compiler/cgen -run '^TestPhase33ScalarExplicitDefectShells$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/dynamic-defect-control-pass.log |
| PASS | go test ./internal/compiler/session -run '^(TestPhase35FocusedValueAnswers|TestPhase35AdmissionRefusalMutationControls|TestPhase37CRC32SingleCopyAndIdentity|TestPhase34ImportedDiagnosticPresentation|TestPhase34PresentationUsesCapturedSource)$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/semantics-import-and-diagnostics-pass.log |
| PASS | go test ./cmd/schway -run '^TestPhase37CRC32AlgorithmMutants$' -count=1 -v -timeout=600s | /tmp/schway-m007-audit/crc-mutants-pass.log |
| PASS | go test ./cmd/schway-repair -run '^TestPhase34RepairJSONCompatibility$' -count=1 -v -timeout=180s | /tmp/schway-m007-audit/repair-json-pass.log |
| EXPECTED FAILURE / reproduced CR-03 and CR-04 | go test -overlay=/tmp/schway-m007-audit/peer-overlay.json ./internal/compiler/executionpeer -run '^TestM007AuditResidualPeerDebt$' -count=1 -v -timeout=120s | /tmp/schway-m007-audit/peer-debt.log |
| EXPECTED FAILURE / reproduced CR-06 | go test -overlay=/tmp/schway-m007-audit/cgen-overlay.json ./internal/compiler/cgen -run '^TestM007AuditResidualNoCallNativeDefect$' -count=1 -v -timeout=120s | /tmp/schway-m007-audit/cgen-nocall-debt.log |

The source/test/log hashes and temporary overlay input hashes are in /tmp/schway-m007-audit/input-hashes.sha256. It includes the critical current source files, named phase tests, temporary diagnostic overlay tests, and logs. The hash list itself has SHA-256 29cd884239b2a90311a49289fab1fdc6ff764d23d34569cd8cb62737fad770e6.

## Evidence boundary

Historical Phase 33–37 verification records were read as submitted receipts. The focused checks above are newly executed on the audit HEAD and independently exercise important connections. Source inspection is labeled separately through the line anchors above. No full local project/package suite, hosted CI, Linux run, new hosted receipt, cold/warm performance distribution, or human readability study was performed. The go test output notes that an emitted artifact can vary in BuildID under the non-cacheable/incomplete-closure contract; no deterministic-binary claim is made.

## Orchestrator evidence qualification

The raw logs and original overlay inputs are retained in `evidence/`; `EVIDENCE.json` binds their hashes and exact commands. Peer diagnostics use handcrafted core/events, not an actual checker-admitted interpreter run. Their rejection agrees with source inspection and the pre-existing critical review; paired interpreter/peer closure regressions remain required. The CR-06 probe proves current native.run_failed via its output, but is an always-failing diagnostic, not a permanent regression test. The earlier checker-refused probe is excluded. Hex formatter/recovery claims also rely on the current phase verification records; those particular roots were not rerun by this audit.
