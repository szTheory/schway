# Plan 37.1-04 bounded evidence index

Captured 2026-10-16. This records implementation evidence for CR-06 and the
strict compatibility amendment gate. It is not phase closeout: code review,
security, validation, independent verification, report freshness, and the M007
audit remain downstream gates owned by the phase orchestrator.

## Behavioral evidence

The checked `examples/sum_to_n.schway` witness has no checked `OpCall` at the
defect point. `sum_to_n(1001)` now produces one complete structured native
defect document with the literal reason `input exceeds 1000`; the execution
peer accepts it at O0, O3, and O3-LTO. `sum_to_n(5)` still returns 15. The
application shell still has empty stdout, exit 65, and exact
`schway app: input exceeds 1000\n` stderr. The existing dynamic CR-05 entry and
helper cases remain covered, and Unit native conformance remains refused.

Defect reasons use counted emission for the cases where C string or JSON
string helpers would lose bytes. Literal assertions cover ASCII, U+0001+A,
NUL+A, NUL+Z, other control bytes, non-ASCII/U+FFFD, and backslash spellings
through no-call, dynamic-entry, and dynamic-helper witnesses. Native records
and application stderr preserve the exact bytes at O0/O3/O3-LTO. Ordinary
ASCII output retains its existing emission path. The focused sanitizer probe
was available on this host (address/heap-use-after-free smoke); the counted
reason path exited cleanly. This is a local run only.

The Task 1 semantic RED is preserved in `/tmp/phase371-04-01-go-red.log`
(SHA-256 `698d1e3a59d7137651f247a9caac68b45c1a0dfc2233f994e3d35da4fff0b5fe`)
and `/tmp/phase371-04-01-tap.log` (SHA-256
`e50314a1381dd7cd76433ee724a70018bcbab484e5e165baed935d534fe2f5de`). All
three optimization lanes reached the intended native semantic assertion:
the native conformance runner received application exit 65 instead of a
structured defect record. The reason-helper RED is
`/tmp/phase371-04-02-reason-red.log` (SHA-256
`ef10bdf9720289ed1d75986edb91d6e39a3dc92138cbd8c18a9d375c3bcefbf`); it
reached exact byte mismatches rather than a compile, tool, or skip failure.

## Frozen C amendments

The immutable census contains 325 source rows: 162 Phase 34 sources with four
route families and 163 Phase 36 sources with three route families. The final
measured impact is five exact `sum_to_n` route rows. All are bound to the same
source SHA-256
`6e3a5c992eb45a741ba17d55df3c378e681a7b151544d6625000cc23411f3aa6`, checked
core SHA-256
`b149a1761f25983ae17d817bc034bc03da8a3f12e5a699c236deb929a0055b36`, and
original C SHA-256
`163eb1d42ad269797fb0ea95be9252c643fb495e1f10d4bba372979186bed02c`.

| Frozen corpus | Exact route | Reviewed edit | Resulting C SHA-256 |
| --- | --- | --- | --- |
| Phase 34 | `control_native_c` | offset 10977; unique anchor offset 2, count 1 | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |
| Phase 34 | `emit` | offset 10977; unique anchor offset 2, count 1 | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |
| Phase 34 | `emit_native` | offset 10977; unique anchor offset 2, count 1 | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |
| Phase 36 | `emit` | offset 10977; unique anchor offset 2, count 1 | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |
| Phase 36 | `emit_native` | offset 10977; unique anchor offset 2, count 1 | `ffdaaf82e1fcd0137bd677f36085305c6c5f9049649f6c3345107ce72fd286a8` |

Each row in `testdata/phase371/compatibility-amendments.json` stores the exact
before/after byte strings and anchored replacement. The replacement changes
the no-call defect branch from application stderr plus exit 65 to the
structured `lang.execution/2` defect record and exit-0 transport. The final
actual-source comparator controls pass the exact amendments and reject
missing, duplicate, extra, stale source/core/old-C binding, wrong path, wrong
route, wrong expected digest, wrong or repeated anchor, and an unrelated
post-delta C-byte mutation. Original C corpora and historical source, ledger,
golden, and receipt snapshots were not changed. No path-wide exception,
normalization, regenerated oracle, or skip was added.

## Archived pre-resolution proof records

`371-PROHIBITION-PROJECTION.json` is the machine-readable current index. The
four earlier proof records, their source bindings, and the prior root-union
receipt are copied intact into `371-PROHIBITION-RESOLUTION.json` before the
live proof entries are refreshed. In those pre-resolution records, each bad
Go/TAP run reaches the named semantic failure, clean/default pass, and each
canonical producer is `green`, unflagged, and machine-proves `failFirst: true`.
The original unresolved projection remains unchanged and is checked against
the original plan bytes at its recorded Git `source_head`.

| Subject | Bad Go log SHA-256 | Clean / default Go log SHA-256 | Bad / clean / default TAP SHA-256 |
| --- | --- | --- | --- |
| Defect | `c16c1e19ecefff08141b3c7228ccd6ec2d58aed9454cad5d248597c87ba16e0b` | `77a7524a4fd842a2df75fc927b9ad57c0d681ab3f476d998934e83aa93b1219c` / `19f067a282854e2c1b0830addf695a602899270a78b4e337cba3c20e2cbae1fa` | `f59a8c6fdca173fd5039aa1e05badac780fcdf44ae2d0caaf9e4edeebe6732c9` / `c01c6db67a01e29995ea6385001edc46d8527a55040e2c567ea5b88839a5e186` / `394eeb52aa14ca1f5fb34e0d0b10827e832649662155400cd4173684ee076a18` |
| Typed failure | `53b77b71387e8f242a478aa4636e2280536783a5a7cc2e1899f0ed0a0ad734a5` | `a230e7661a717f827c391ef799c3a7f042a951a3974b3afa3033049bcdf3af10` / `909286be7a882ed9b669b83cbaac2db61392e212373ab6120c6ac8798727636b` | `35972a65c7062db0f2197dbace8efab59a8231f0e33aadd8cdb9f8e513c4dc41` / `86b1580c0db9937b3c4e68d868988b3beccb6c6a41f19e46fbab8b2daf2e80d9` / `3d0756a86453e7210cff3907ecaa92c80284655ee0229038c91aec4c6f457913` |
| Reason bytes | `90096731d443308e5c487d0c10cd7b70a8f5adc7fc2d5997896778f790fcbe1e` | `d1387e1b62ce7af8c7211e84b65c9d2434284214c892d5e6b77b919b6dc08926` / `bec61d5831f52462e1973b412259734ce9bef517328828910f5cddf8b435fe00` | `c8306f361ca381603ac8927cde2731b4a0294f32fe021c93125d0691071927b4` / `a197e57fe4bb653d8e14eac22fc8ce2ebd2b9a297057c8f8044166f40a036a71` / `da0244a4226d530dd8aaf4415f2f7073cf599dc2c60ecc4222b68bd4660a7af0` |
| Baseline amendment | `1a02ed203fa861b50723e6b02dae10b374ac8dc5c0eb19ed753b89e5ac3a42a5` | `a820d15b4506a6910b60ccc5cd684e226ca0e4cef0889241de72a885bfe6196b` / `a7a865a2ebd99df525e35b73bda5b0cefa87ca6cfce53a63dca18c6729c291e7` | `2e8f8024aca09b76280c33b13d5fa2d34ed65ca54fe5bd732ec311719f3d29b6` / `704662e82a3cfad78109aa2172cd192eec23639ac60ddbf9afeeb80788b218b4` / `544f6389e89930b0bf5ff2ca8bc7b359d8bb0a9e1345fa3f54a316d9da344532` |

The reason proof's exact original-source overlay binds cgen, parser, and core
validation. Its test log fails on the original reason helper's control/NUL
byte loss, the original parser's missing malformed-UTF-8 byte spans, and raw
core acceptance where `core.final_claim_mismatch` is expected. The overlay
manifest binds the three source replacements and records source files restored
unchanged after the run. The baseline proof overlay changes only the actual
compatibility comparator predicate; its source-bound Go root fails because an
unrelated extra C byte is accepted. Both controls are semantic failures.

Pre-resolution source bindings include these source and fixture digests:

| Bound input | SHA-256 |
| --- | --- |
| Plan 37.1-04 | `2f5f9c500aff40d5a00a23f8c37bf583ce29a5fa23b7466fedba9d8ccc307dea` |
| `371-prohibition-check.test.cjs` | `3e446dd90b9ee96de2138f0e0dcd8f804b372ec8a38aabcd260818618c808fa6` |
| `cgen_program.go` | `312dc57dc95ae61101dc211d0e10f8c46b0dfbee2e934f1be95d6c826d7ac5c2` |
| `cgen_phase371_legacy_defect_test.go` | `cdbf82cc785ffa3f24a212f30f77227ff6ec0b09cc3549463b383ad5101ab302` |
| `session_phase371_compatibility_test.go` | `83d8b4cd4b92a1b256babda066b4d40492e0726486efbb0cfca88d3832602f31` |
| `testdata/phase371/pre-emitter-census.json` | `518f32d2fcab58fce5719baf5d7248e3f032d48afc212bb535a463aa1a9c5f27` |
| `testdata/phase371/compatibility-amendments.json` | `2dd098ee0038f120d0670c3dd059b589834d610ebbc8f02b42f645e032235c78` |
| defect violation fixture | `cd29ad16fd8de29035afc52d647526dac8db50316106648d8b697abd6037dc30` |
| failure violation fixture | `86d278f9d0c8f146e6ace8452865206b7f1b43beb5e36c9a7409ed0015c30b00` |
| reason violation fixture | `df2dc168a2767d277d027e79693fa667cfa93ff7dc6a0089bed348eaa402bcf8` |
| baseline violation fixture | `8c6837dd4bd95592bf9988fbecdfb36a8a8aa30d86e8fc79da89adaa7400706f` |
| shared clean fixture | `6f1aa9862183245c6523b25829cee150ef4926953da22c5fbfec5ff83ee0e27a` |

## Final bounded root union

The pre-transition adapter's `--sample all` and `--prove all` observed 19
distinct roots across three packages, all run once with 253/253 observed
subtests passing and zero skips. Its original raw log is
`/tmp/schway371-prohibition-proof-TN3LOW/final-19-root-union.go.log`, SHA-256
`c24e4fc34fffc911d7f699202c8c1c1df5e274beb2aa97d82b2e11e6fc546e64`.
Package counts are executionpeer 6 roots / 93 subtests, session 10 / 70, and
cgen 3 / 90. The exact old union is retained inside the resolution archive;
the current post-transition union is in `final_root_union.roots` in the
projection.

The roots are `TestPhase371LegacyDefectClosesAllActiveFrames`,
`TestPhase371DefectTerminalControls`, `TestPhase33TerminalAttribution`,
`TestPhase36CallCausalAttributionControls`,
`TestPhase371ForeignFailureOpFailAttribution`,
`TestPhase371ForeignFailureAttributionControls`,
`TestPhase24ModelReplayIsDeterministicAndModelOnly`,
`TestPhase371DefectReasonUTF8Admission`, `TestPhase371PreEmitterRouteCensus`,
`TestPhase371CompatibilityAmendmentControls`,
`TestPhase34PreChangeCProgramGolden`,
`TestPhase34PreChangeCGoldenIsNotInert`,
`TestPhase36NoCallGeneratedCByteBaseline`,
`TestPhase36NoCallGeneratedCByteBaselineControls`,
`TestPhase37Phase34AdditionLedgerControls`,
`TestPhase37Phase36AdditionLedgerControls`,
`TestPhase371NoCallScalarDefectNativeConformance`,
`TestPhase371CountedDefectReasonBytes`, and
`TestPhase33ScalarExplicitDefectShells`.

## Current lifecycle resolution

The four live plan prohibitions now use the canonical test-tier state
`status: resolved` and `verification: test`. Their statements and flat check
descriptors are unchanged. The initial projection and its source prohibition
remain `unresolved` with null verification, bound to the original Plan 37.1-01
bytes (`source_head` `a80a290b059bd6b998f7a87ee7f5ef89ec212d58`, SHA-256
`14578da0811b83e0eb5b9511824eed17f79fc6996772beba81e323af2719513a`). The
resolution record preserves all four pre-transition proofs and their old
source bindings before replacing the current entries.

Fresh canonical producer runs executed after the plan transition for defect,
typed failure, reason bytes, and baseline amendment. Each returned
`green`, `flagged: false`, `tier: test`, and a passing fail-first proof against
its violation fixture. The source-bound bad Go and TAP lanes reached their
semantic assertions; clean/default lanes passed without skips. Current plan
hashes are `79bcfbfbe3171bc5c9c9383cc83fa14dbfcfdc97251c18a71fb5ac9207d71798`
(Plan 01), `ea2dd99835f02af8e58d7c298363e6772261da173a2155116d2b61a4dddea4d6`
(Plan 02), `62b55c7519f1d6cff46876544d0ad23134f74632cdbebcdcb96e000c67b7e0c8`
(Plan 03), and `2a0d8b81f107b81f678d0977798c413430cad37605370ba236440f323f5d6b00`
(Plan 04). The updated adapter binding is
`672a96a7cddc2a3cec5e72ed75a4b26ff7bc86d19bbd7ec93073bf5eead87552`.

| Subject | Current proof manifest SHA-256 | Canonical producer request SHA-256 | Producer verdict |
| --- | --- | --- | --- |
| Defect | `cdd1d3fa449382ca034c88e3c1583f96ac3f40caf626b8b159453feb5ec2b2d5` | `d120174f144c3831d897082c61e9234730109acdba0762eca6ed6fdde4b61452` | green, unflagged |
| Typed failure | `f8dd4b44847f1ea4699270a37a78be15bc448314fde5da0fc9e754282721f669` | `151c7acbdc8084a522f7b7edc22f97042d8b1b8962ea1def7546397196962661` | green, unflagged |
| Reason bytes | `ba79464785372e764c4a60f48b8cc23fefa7e97cbc84498676b96a374a016196` | `109cb6e427e276337a23e288a5a3d1b9c2788b0db3d23504cd33dea7e11e0316` | green, unflagged |
| Baseline amendment | `313e957202dff42d2ef36813f4538824e9b1ff9601011aab03a71321f901f4a0` | `3f97395cbfd4d4fb5d65ad5188dda52206973f764c291d116106ba49ac7326b9` | green, unflagged |

The fresh post-transition `--prove all` run again observed 19 roots across
three packages, 253/253 subtests passing, and zero skips. Its raw log is
`/tmp/schway371-prohibition-proof-d3Qy11/final-19-root-union.go.log`, SHA-256
`16ddd536635f2a1028fb8e2e435b6e2dad144424184d5383a9a72c91f5f6dda8`. The
post-transition `--sample all` also passed with the same 19-root inventory;
its manifest is `/tmp/schway371-prohibition-proof-6OFheN/sample-manifest.json`
(SHA-256 `1533899b3ae1ea2c0e5e500cca6d4d3b50dcd9fba01b4b4c398cdce61d4597ac`)
and its raw log is SHA-256
`a209e6f21c2ded30e2a371cce936e31e2c37f1fe5034da499471bd7996f64af3`.
Four temporary plan copies were rejected for an unsupported status, tier,
partial transition, and altered descriptor; the live plan bytes were then
re-read and accepted as resolved/test. The exact rejection reasons and copy
hashes are in the resolution record. It also re-parses each prior unresolved
plan at source-binding revision `8fe28e30f00104ae62d568db0bf95c349e1da34e`,
then checks that canonical serialization yields the same resolved/test
candidate with an unchanged descriptor. That record is SHA-256
`7782c739b780c9d5dcf8d3731105e0bff0b80e43c69874c284aca67108bc114d`.

## Transition disposition and limits

Source inspection confirms that `.planning/PRODUCT-ROADMAP.md`,
`.planning/LANGUAGE-MATURITY.md`, and the dated 37.1 context retain the current
capability order. This terminal implementation does not change it:

| Next capability | Blocker and smallest complete slice | Checker/guarantee impact; evidence and owner | Priority-changing observation |
| --- | --- | --- | --- |
| Audited `crc32`/`crc_low`/checksum and earlier pinned programs on Ubuntu/macOS | CR-03/04/06 closure, fresh gates, then one freeze with required same-source two-host receipts | Correct represented terminal/typed-error laws and reason bytes while retaining separate numeric, causal, and process proofs. Evidence: M007 audit, this phase context/probes, EVD-17, historical M006 receipts. Phase 37.1 peer/emitter owners hand off to Phase 38 receipt owner. | Reached regression, missing case, host disagreement, or source drift delays freeze. |
| Proposed M008 `wc FILE` / `longest_line FILE` over one module | Output contract and justified arity-N/scalar-record/exported-type needs | Specify bounded inputs and literal answers, then implement the pair. Checker changes only for consumer-required argument, record/export, and result laws. Evidence: current module/ByteView/FileBytes/CRC witnesses and future requirement landing. Next milestone language/checker owner. | Existing scalar results suffice, or a consumer needs fewer new laws. |
| Bounded fixed-schema `key=value` checking before JSON | Schema, duplicate, and malformed-input contract undecided | One schema, byte/line bounds, literal good/bad fixtures, bounded errors; add parser work/error attribution and representation only if required. Evidence: product roadmap and fixed file reader; no JSON admission. Future consumer owner specifies contract before promotion. | A concrete JSON consumer offers a smaller justified slice, or existing operations suffice. |

The relevant dormant seeds remain dormant with their source and consumer
witnesses: z1g (`QueryCommandFile` and CRC helper/failure-site navigation),
because ordinary search found the sites and no measured navigation gap;
trz (`readApplicationCapture` and modeled replay for CRC/checksum diagnosis),
because current sources, traces, and narrow probes explain the failures and
none is unexplained; and z1s (`syntax.Format` and checked CRC/helper sources),
because no source-form dispute or measured human-review gap is present. Their
existing named owners must capture their trigger observation before promotion.
The Dashbit appraisal remains historical; no experiment, requirement, or
future consumer is promoted.

The local evidence does not claim hosted authorization or receipts, Linux
execution, physical cleanup/destructor proof, a numeric peer guarantee,
human-readability results, or whole-loop latency distributions. EVD-17 remains
unchecked/Pending and Phase 38-owned; CALL-01/02 remain Phase 36-owned. The
original historical projection remains unresolved by design. The current
live projections are resolved/test on the basis of the recorded machine
proofs; independent code review, security, validation, verification freshness,
and the M007 audit remain downstream phase gates.
