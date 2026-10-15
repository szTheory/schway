# Phase 37 transition check — 2026-10-15

This transition update records current source inspection and newly executed
local evidence separately from historical receipts. Phase 37 implementation
and its focused local command capture are present; the independent parent
verifier, security audit, code review, and requirement acceptance remain
pending. This document does not mark Phase 37 complete or authorize Phase 38.

## Three scoped capabilities

| Program / capability | Current blocker | Smallest complete slice | Checker / guarantee changes triggered | Evidence and debt | Owner / next action | Observation that changes priority |
| --- | --- | --- | --- | --- | --- | --- |
| `crc32 FILE` and `crc_low FILE` using one ordinary-source CRC-32/ISO-HDLC module | The source consumer and second importer are implemented; APP-13/MOD-07 acceptance and the independent phase gates remain pending. MOD-07's classifier row remains flagged unresolved. | Parent goal-backward verification, security and review of the current implementation; keep requirement status pending until those gates accept it. No new implementation slice is currently identified. | No new checker feature is requested. Existing scalar-call, ByteView, module-admission and resource-discharge guarantees remain the contract. Independent numerical answers must stay separate from causal attribution. | [37-ACCEPTANCE-EVIDENCE.md](37-ACCEPTANCE-EVIDENCE.md), Plans 37-01/05/06 summaries, C14-C18 local owners, and the unchanged raw classifier history in [37-PROBE-RESOLUTION.md](37-PROBE-RESOLUTION.md). | Parent's phase verifier, security auditor, and review owner close their gates; the requirement owner resolves or preserves the flagged classification explicitly. | A gate finds a reached semantic defect, or a named file consumer needs a different bound, signature, or failure contract. |
| Public `schway run` file route on interpreter/native plus the executable CRC guide | The parser, reader, outcome boundary, and guide are implemented. The remaining blocker is APP-14/DX-24 parent acceptance; no new route defect appeared in the local capture. | Parent independently accepts exact route outcomes and the tracked-only guide evidence. Keep the guide's formatter and timeout limitations explicit. | No general foreign-execution or file-I/O framework is triggered. The interpreter retains its fixed 65,536-byte reader; native behavior remains binding-specific. | C19-C24 in [37-ACCEPTANCE-EVIDENCE.md](37-ACCEPTANCE-EVIDENCE.md); the guide executes 78/78 tuples from the tracked-only copy. Human readability and arbitrary cancellation parity are unmeasured/out of scope. | Parent verifier/reviewer accepts the app route and guide; preserve `(65, "", "")` for an empty path and `(0, "0\\n", "")` for an empty file. | An independent route tuple differs, a flag-shaped suffix stops being opaque, or a named program needs a different input contract. |
| Final audited CRC/checksum receipts on Ubuntu and macOS (Phase 38, EVD-17) | There is no current same-source, two-host hosted receipt; the milestone audit and any gap closure have not yet occurred. | After audit and gap closure, freeze one code and fixture revision, run both supported hosted lanes, require non-skipped roots, and bind both logs to that same revision. | No checker change follows from local success. The guarantee is same-source independent host evidence, not a new backend or deterministic executable bytes. | Phase37 local evidence is bounded to Darwin/arm64; earlier M006/Phase36 hosted receipts remain historical. See EVD-17 in [REQUIREMENTS.md](../../REQUIREMENTS.md). | Phase38 owner acts only after milestone audit/gap closure and parent handoff; no hosted dispatch is part of this plan. | Host disagreement, source/fixture drift, skipped required roots, or a newly found audit defect invalidates the freeze. |

Phase 37 includes a runnable gain: ordinary `crc32.schway` and
`crc_low.schway` programs plus a marked guide that builds and executes the
real consumers. This closes the prior enabling-only sequence concern; it does
not establish full-loop cold/warm latency.

## Source witnesses and evidence labels

**Source inspection:** `examples/crc32/crc32.schway` and
`examples/crc32/crc_low.schway` import the shared `lib.crc32` module;
`cmd/schway/main.go` owns the opaque path route; the interpreter's fixed
reader and application boundary are in
`internal/compiler/interp/filebytes_reader.go`,
`internal/compiler/interp/interp.go`, and
`internal/compiler/session/session_file_application.go`. The native adapter
and binding-specific limits remain in
`examples/crc32/adapter.c` and `internal/compiler/native/bindings.go`. The finite surface witness and the
additive Phase37 receipt preserve the historical Phase34 pinned bytes and the
existing ten-field wire schema.

**Newly executed local checks:** the fixed-source owner commands and hashes
are in [37-ACCEPTANCE-EVIDENCE.md](37-ACCEPTANCE-EVIDENCE.md). C01-C26 had 63
passing Go root invocations after the C07 map correction. The canonical
producer exercised its three current descriptors with bad/clean/default
subjects. The guide's actual marked commands reached 78/78 exact tuples from
2,691 `git ls-files` entries. The maximum-size interpreter read took 2.46
seconds in one Darwin/arm64 Go 1.24.0 run; this is one observation, not a
latency distribution. Focused ASan/UBSan and physical-cleanup owners passed
locally.

**Historical receipts:** the pre-addition baselines, earlier Phase37 wave
captures, and Phase35/36/M006 hosted results retain their original revisions
and scope. The recurring `.github/workflows/ci.yml` was inspected, not
dispatched. No Phase37 hosted receipt, local Linux execution, full project
suite, broad native matrix, human readability result, or cold/warm
generate/verify/run/observe/repair distribution is claimed.

The empty file and empty path are distinct: the empty file yields numeric CRC
zero with a newline; the empty path yields `(65, "", "")` before adapter
execution. CRC-32/ISO-HDLC is distinct from POSIX `cksum`/CRC-32/CKSUM and
CRC32C. The independent Python/zlib oracle is provenance for expected values,
not a compiler/runtime dependency. Native executable digests and full BuildID
may vary when module rows and InputID remain stable under the current
incomplete/non-cacheable receipt contract.

## Conditional seed dispositions

| Seed | Disposition | Current source witness and named consumer | Trigger observation | Next action |
| --- | --- | --- | --- | --- |
| z1g | retain dormant | `internal/compiler/session/session_phase6_query.go#QueryCommandFile`; the CRC helper and import callsites are available to ordinary source search and the guide runner. | No named reviewer/agent task measured an accuracy or material-cost gap in query plus search. | Require one measured failed lookup before scoping a source-unit symbol/caller query. |
| trz | retain dormant | `internal/compiler/native/native_app.go#readApplicationCapture`, `internal/compiler/session/session_app_verify.go#VerifyApplicationCasesFile`, and interpreter capture-retention/outcome roots. | The current capture, observer, and replay evidence explain the exercised CRC/file outcomes; no new unexplained consumer failure was named. | Review one actual unexplained failure against existing evidence before proposing optional diagnosis. |
| z1s | retain dormant | `internal/compiler/syntax/format.go#Format`; the CRC guide states the current formatter limitation. | No named consumer or human reviewer contested a concrete source-form or contract choice. | Require a real paired-form dispute and independent reviewer obligations before an experiment. |

The [Dashbit AI-era appraisal](../../sources/DASHBIT-AI-ERA.md) was consulted
against these current witnesses. Its positions on explicit obligations,
source-local queries, optional observability, and shared-library reuse do not
create a trigger without a named measured consumer. No seed was promoted,
rejected, or superseded.

## Deferred owners

Phase38 owns hosted evidence only after the milestone audit and gap closure.
M008 `wc`/`longest_line` remains the first later consumer requiring
arity-N/scalar-record/exported-type work; general arity expansion, bounded
key=value/JSON, module-aware query/evidence/interface work, arbitrary
interpreter cancellation, full large-file causal traces, and whole-loop
performance measurement remain deferred to their existing consumer triggers.
No new backend, general planning framework, or runtime dependency is proposed.

## Independent audit remediation closure —2026-10-15

The initial8plans missed D-02 helper placement and the explicit capture/Go-reader negative controls in37-04. Nine plans/21tasks are now summarized.37-09 fixes private eight-round byte-helper placement and adds3byte/3actualcall structural evidence plus5capture-growth/3reader reached mutants. Independent security re-audit31/31 closed (30mitigations,1existing accepted limitation), all repaired hashes checked. Earlier numeric/public/guide receipts stay historical; current affected receipts are in37-REMEDIATION-EVIDENCE.md. Validation restored on actual controls. Phase acceptance, requirement closure and hosted work remain separate; no new capability, dependency, linker flag or wire field.
