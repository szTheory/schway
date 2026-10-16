# Phase 37 stale-report refresh gates

Captured 2026-10-16 on source HEAD `9258a7360222ff17a3267e1ccb873a0dfae79346`, Darwin arm64, Go 1.24.0, Node v26.10.0. This was a stale-report refresh inside the already active `$gsd-execute-phase 37` workflow. The phase has 9/9 plans complete, 21 tasks, and 6 waves. Gate order was validation, security, UI review, standard-depth code review, disposition, then regression. Configured validation/security/UI/code-review hooks are enabled; `workflow.code_review_depth` remains `standard` and user edits to `.planning/config.json` were preserved.

## Validation

The 28-entry task command registry and all 21 task rows were refreshed against current HEAD. C01–C23 ran directly; C26 ran directly; C27 and C28's non-overlapping roots ran as exact bounded selectors. C24's three roots are covered by C22+C23, C25's four roots by C26, C27's ten roots by C03/C04/C18/C26 plus its unique helper-call root, and C28's seven roots by C12 plus its three unique mutation roots. Those four mapped command IDs are marked as root-union coverage in the structured receipt; their full command strings were not presented as separate executions.

The directly executed Go owner roots totaled 60/60 PASS, with 0 failed, missing, or skipped roots. C05's Node TAP check passed 3/3 tests. C06's canonical prohibition proof passed all three descriptors in bad, clean, and default modes; each bad fixture reached its expected RED assertion, all clean/default controls were green, and canonical `check prohibition-enforcement` returned `green`, `flagged=false`, `failFirstProof=violation-fixture`. The independent command-registry guard also passed 2/2 roots and reached all 22 intended refusal controls without skips.

`37-REFRESH-RUN.json` records every C01–C28 command, the execution or root-union basis, root counts, raw-log paths and SHA-256 digests, test-reported durations, skips/missing/failures, and current input fingerprints. The pre-proof covered-file fingerprint was `v3:sha256:3e444641caf0b9ddd7f5a65dbcc1643ce364b6180a505b1825c12ab2a1d914e5`; after all gates it was `v3:sha256:4fc4eac54f1e0d7b4e6f1b583763a17ba9f4f7c633b5bf18b841ac4b02750835` across the same 78 covered files. C06 refreshed only the current-HEAD rows in `37-PROHIBITION-EVIDENCE.md`, changing that covered input's SHA-256 from `d69a477cb4a15025dde65630e717f8918d5ccb86d0147a84168f1007730205d8` to `e68b3c22a884be2d37fb0d194435af7342008f7a4464f8549cab6c20242f732f`. The stored verifier digest remains stale pending the parent-owned verifier run.

## Security and UI

The authored L1 threat register remains 31/31 closed: 30 mitigations and the already accepted T-37-07 limitation, with zero open threats. Its secure-phase short-circuit conditions remain satisfied. The refreshed canonical proof, current bounded owners, and affected regression roots exposed no new threat or mitigation gap, so `37-SECURITY.md` was left byte-identical (SHA-256 `d4dad33e5ce8b31a292be6844b0dc05d66bcfc237607d4be03ec76bb3c8fcccc`).

There is no frontend, UI specification, or screen surface in this phase. The existing six-pillar UI review remains N/A; `37-UI-REVIEW.md` was left byte-identical (SHA-256 `2677ecbc7adcbf218983d829f0f37e43e5d18aa989bf9a09d123725759571ae7`). No visual score, screenshot, or readability claim was added.

## Code review and disposition

Standard-depth review covered the relevant source delta since the prior Phase 37 verifier plus the current CRC, file-route, interpreter, native, and dynamic-call seams. The bounded delta comprised `cmd/schway/main.go`, `examples/crc32/README.md`, `examples/crc32/lib/crc32.schway`, `internal/compiler/cgen/cgen_program.go`, and `internal/compiler/executionpeer/executionpeer.go`; their current hashes and evidence owners are in `37-REFRESH-RUN.json`. The CLI grammar and guide changes agree with current route/guide owners; the CRC helper now has one helper call per byte and is checked by its helper-count, numeric matrix, mutation, identity, and ledger owners. The cgen and executionpeer deltas were checked against bounded cgen/native/peer regression selectors. No new actionable finding was identified.

Historical WR-01 remains fixed by the existing review fix; the existing disposition remains 0 open findings and was left byte-identical (SHA-256 `d0e3a46b6ced61d678cd1a7c45a2f2960e36ff4f0070a74c1b344a344f5adeac`). The deferred Phase 33 CR-03/04/06 obligations remain separate audit debt and are not counted as Phase 37 findings. This review does not claim legacy no-call native parity.

## Regression

The refresh reran the exact bounded owner inventory from `36-REFRESH3-REGRESSION.json` in 16 sequential package groups, using `GOCACHE=/tmp/ai-lang-verification-gocache`, `-count=1`, `-timeout=600s`, and Go JSON event logs. All 252 root executions passed; there were 251 named package selectors, 240 distinct test names, 0 failures, 0 skips, and 0 missing roots. `TestPhase35NumericRoutePolicy` ran twice in the checker package's internal/external test packages and is counted as two executions of one selector. Total elapsed time was 140.75 seconds; the session selector group took 122.539 seconds. Pre/post hashes for all 182 captured inputs had zero drift. Per-package commands, results, root-level events, logs, and hashes are in `37-REFRESH-REGRESSION.json`; raw logs are under `/tmp/schway37-refresh-regression/`.

No whole-package or project-wide local suite ran. The host is Darwin arm64; no Linux or hosted result is claimed. Subjective readability and cold/warm latency remain unmeasured. `37-VERIFICATION.md` remains stale until the parent-owned verifier refreshes it; this gate refresh does not claim Phase 37 complete.

## Receipts

- `37-REFRESH-RUN.json` — current command/root map, gate order, input fingerprints, review scope, and raw-log hashes.
- `37-REFRESH-PROHIBITIONS.json` — all three canonical bad/clean/default proof subjects and outcomes.
- `37-REFRESH-REGRESSION.json` — bounded 252-root regression result and per-input before/after hashes.
- `37-PROHIBITION-EVIDENCE.md` — refreshed by the required canonical C06 proof at current HEAD.
