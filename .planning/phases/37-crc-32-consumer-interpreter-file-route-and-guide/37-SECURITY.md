---
phase: "37"
slug: "crc-32-consumer-interpreter-file-route-and-guide"
status: verified
threats_open: 0
asvs_level: 1
created: "2026-10-15"
---

# Phase37 security

Independent authored-register audit at538e810e:29 mitigated closures, one missing high-severity control, and one pre-existing accepted limitation needing log persistence. Parent records that existing acceptance below, so30 rows are closed and T-37-14 remains blocking. No new risk is accepted. Evidence anchors below are audit-source anchors and may move during repair.

## Trust boundaries

Source/module admission to fixed interpreter file capability; opaque argv to descriptor IO; optional capture to actual outcomes; manifest-backed native build authority; independent tests/proofs to public receipts.

## Threat register

| Threat ID | Category | Component | Severity | Disposition | Mitigation / evidence | Status |
|---|---|---|---|---|---|---|
| T-37-01 | Tampering | CRC recurrence and shared module | high | mitigate | lib/crc32.schway; session_phase37_crc32_test.go; main_phase37_crc32_controls_test.go (numeric security obligation closed; separate D-02 structural defect open) | closed |
| T-37-02 | Elevation of privilege | prohibition adapter fixture selector | high | mitigate | 37-prohibition-check.test.cjs:57,68,92,157 | closed |
| T-37-03 | Repudiation | baseline/proof receipts | medium | mitigate | 37-BASELINE.md; 37-prohibition-check.test.cjs:114; 37-PROHIBITION-EVIDENCE.md | closed |
| T-37-30 | Tampering | additive tracked-source ledger | high | mitigate | session_phase34_prechange_test.go:170,212,684; session_phase36_prechange_test.go:233,459; original ledger hash unchanged | closed |
| T-37-SC | Tampering | package installs | high | mitigate | go.mod standard-library-only; no phase dependency changes; Node built-ins; adapter triple equality | closed |
| T-37-04 | Elevation of privilege | session file application | high | mitigate | session_file_application.go:19,26,30,33; interp.go:736,790,810 | closed |
| T-37-05 | Denial of service | fixed reader | high | mitigate | interp/filebytes_reader.go:27,52,62,70,93 | closed |
| T-37-06 | Tampering | capture-disabled frame stack | high | mitigate | interp.go:721,755,1943,2070,2310; session_phase37_file_application_test.go:49; interp_phase37_test.go:111,126 | closed |
| T-37-07 | Denial of service | arbitrary divergent source | medium | accept | Existing acceptance in37-02-PLAN and CONTEXT D-15; guide:125; native_app.go:425 process deadline. Accepted-risk entry below closes documentation bookkeeping. | closed |
| T-37-08 | Tampering | opaque prefix/suffix parser | high | mitigate | main.go:37,224; native_app.go:431; main_phase37_file_run_test.go | closed |
| T-37-09 | Elevation of privilege | interpreter authority | high | mitigate | session_admit.go:33; interp.go:790; session_phase37_authority_test.go:78,127,175,190 | closed |
| T-37-10 | Information disclosure | input/typed error rendering | high | mitigate | main.go:277,316; filebytes_reader.go:43; exact public tuple controls | closed |
| T-37-11 | Tampering | native one-shot outcome | high | mitigate | main.go:334,365; native/bindings.go:224,270,285; native_app.go:414,431 | closed |
| T-37-12 | Denial of service | special/oversized/changing input | high | mitigate | filebytes_reader.go:46,52,62,70; filebytes_reader_test.go:123,163,178,205,224 | closed |
| T-37-13 | Tampering | cleanup/error precedence | high | mitigate | filebytes_reader.go:92; examples/crc32/adapter.c:61,98,106,123; native_phase37_filebytes_test.go:224,238,297 | closed |
| T-37-14 | Tampering | disabled-capture retention | high | mitigate | Independent re-audit at3cb61cde: five real growth overlays and clean/actual-outcome discriminators in interp_phase37_test.go:55,136,158,184,204,241,316; all five markers reached exactly once. | closed |
| T-37-15 | Repudiation | sanitizer availability | medium | mitigate | native_phase37_filebytes_test.go:331,352,361,379; explicit sanitizer/host/leak boundaries | closed |
| T-37-16 | Tampering | circular or wrong oracle | high | mitigate | crc32.expected.json; crc32_oracle_test.go:12,84,145,170,202 | closed |
| T-37-17 | Information disclosure | error/path tuple matrix | high | mitigate | main_phase37_matrix_test.go:75,106,109,116,157 | closed |
| T-37-18 | Repudiation | skipped/missing matrix rows | high | mitigate | crc32_oracle_test.go:39,51,54,57; matrix_test.go:75,92; crc32_oracle.go:237,250,267 | closed |
| T-37-19 | Tampering | parallel CLI state | medium | mitigate | main_phase37_matrix_test.go:290,305,315,324,343 | closed |
| T-37-20 | Tampering | inert/unreached algorithm controls | high | mitigate | main_phase37_crc32_controls_test.go:37,84,88,202,217,256 | closed |
| T-37-21 | Tampering | duplicated/stale imported implementation | high | mitigate | session_phase37_reuse_test.go:67,94,116,125,171; main_phase37_crc32_controls_test.go:113; dated InputID amendment | closed |
| T-37-22 | Repudiation | mutated source restoration | medium | mitigate | main_phase37_crc32_controls_test.go:38,74,160,272; session_phase37_reuse_test.go:50,106 | closed |
| T-37-23 | Tampering | route ownership inventory | high | mitigate | main_phase37_inventory_test.go:24,46,97,198 | closed |
| T-37-24 | Information disclosure | misleading cap/authority facts | medium | mitigate | surfacemanifest/manifest.go:138,188,200; session_phase37_surface_test.go:68,178,237 | closed |
| T-37-25 | Tampering | public exit semantics | medium | mitigate | session_phase28_exitdocs_test.go:71,95,124; public outcome controls | closed |
| T-37-26 | Tampering | untracked guide dependency | high | mitigate | main_phase37_guide_test.go:410,529,813 | closed |
| T-37-27 | Tampering | guide/fixture command drift | high | mitigate | main_phase37_guide_test.go:100,237,566,610,701 | closed |
| T-37-28 | Repudiation | stale or inflated evidence | high | mitigate | 37-ACCEPTANCE-EVIDENCE.md; 37-prohibition-check.test.cjs:114; 37-PROHIBITION-EVIDENCE.md | closed |
| T-37-29 | Information disclosure | guide authority/error claims | medium | mitigate | examples/crc32/README.md:9,26,35,114,118,125; guide_test.go:742,762 | closed |

## Accepted risks log

| Risk ID | Threat ref | Rationale | Accepted by | Date |
|---|---|---|---|---|
| R37-01 | T-37-07 | Generic interpreter cancellation and arbitrary-divergence timeout parity are outside this phase. Only bounded pinned programs are claimed; native process deadline remains separate. | Existing37-02 plan / CONTEXT D-15 under user-delegated recommendations; this persists prior acceptance, not new approval |2026-10-15|

## Security audit trail

| Date | Total | Closed | Blocking open | Auditor |
|---|---|---|---|---|
|2026-10-15|31|30|1|Independent gsd-security-auditor; parent persisted existing accepted-risk entry|

## Required repair

37-09 closes T-37-14 with reached capture-growth controls, and separately corrects the D-02 byte-helper placement. Security remains blocked until the former is independently reaudited. Validation reopened; phase is incomplete.

## Independent repair re-audit —2026-10-15

At3cb61cde independent gsd-security-auditor returned SECURED31/31 (30 mitigated,1 existing accepted risk), threats_open0. T-37-14 now has actual recorder/event/invocation/segment/ordinal overlays with exact reached semantic failures, real successful/late-defect checks and meaningful unarmed targets. Reader wrong-bound/type/close mutations are reached at filebytes_reader_test.go:232,275. D-02 separately resolved: byte/helper loop placement and bounds pinned,3bytes produce3actualhelper calls (structural, not numeric or measured max count).

Fresh independent5interp+2session roots all passed with no skips; all5RETENTION and3READER markers exactly once. All11 remediation file hashes, canonical source/plan/fixture hashes, current ledger and copied binding equality matched. Earlier source anchors remain audit-history anchors; current T14/D02 anchors above supersede the blocked condition, not its history. Native artifact/BuildID variance and incomplete/noncacheable posture remain explicit; InputID/module checks passed. No new risk acceptance, dependencies or cross-host claim.

| Date | Total | Closed | Blocking open | Auditor |
|---|---|---|---|---|
|2026-10-15|31|31|0|Independent gsd-security-auditor repair re-audit|

Sign-off: authored L1 register verified2026-10-15; phase-goal acceptance remains separate.
