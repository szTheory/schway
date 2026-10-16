# M007 Cross-Phase Integration Audit — Terminal Closure

Audit head: `e732b8e7d09e6ff796d04c959033bca8a0f34af2` (2026-10-16). Scope is Phases 33, 34, 35, 36, 37, and 37.1; Phase 38 is not started. This is a fresh, independent wiring audit after the predecessor verification refreshes. It is not a milestone-completion declaration.

## Integration result

**No new cross-phase blocker or broken runtime flow was found.** The six exercised user/compiler flows below trace through their producer, consumer, and outcome evidence. CR-03, CR-04, and CR-06 are each WIRED and closed within the evidence limits recorded here. EVD-17 remains pending with its assigned Phase 38 owner. The Phase 37.1 verification report still records `gaps_found` at its snapshot; its predecessor-report-refresh condition has since been met, and the final Phase 37.1 verification/closeout remains a parent-owned gate.

| Area | Result |
|---|---:|
| Expected end-to-end component connections inspected | 10 WIRED, 0 BROKEN (6 cross-phase, 4 within-phase) |
| Scoped CLI/compiler end-to-end flows | 6 WIRED, 0 BROKEN |
| Phase-local requirements with no later feature consumer | 3 identified; self-contained |
| HTTP API routes / authenticated web areas | 0 / 0 (not applicable to this Go CLI/compiler) |
| Requirement rows | 15 WIRED, 3 PARTIAL metadata traceability, 1 PENDING (EVD-17) |

“PARTIAL” for CALL-01, CALL-02, and DX-24 means their requirements are verified technically but absent from canonical summary frontmatter. It is a documentation cross-reference gap, not a broken compiler connection. The three Phase 36 requirement rows were manually joined to the verifier and validation; DX-24 was manually joined to Phase 37’s verifier and validation. See `REQUIREMENTS.json` and `SUMMARY-EXTRACTION.json`.

## Wiring map

| Provider → consumer | Verified path | Status |
|---|---|---|
| Phase 33 bounded capture → native evidence-capture path | Capture reader caps reads at limit+1 before full allocation; oversized input reaches the named capacity result; the unbounded-read mutant is rejected by the named control. This obligation is phase-local and has no later feature consumer. | WIRED, self-contained |
| Phase 33 generic freeze/receipt gate → later GSD evidence workflow | Generic gate consumes milestone program list and codedigest; later M007 evidence/proof workflows use source-bound inputs and named gates without copying a Phase-specific freeze script. EVD-17’s final hosted receipts are still owned by not-started Phase 38. | WIRED; terminal hosted binding pending |
| Phase 34 parser/checker diagnostics → CLI presentation | Parser recovery and checked diagnostics flow to text `file:line:col`/snippet/caret and versioned JSON spans with byte offsets plus line/column. Syntax-error, unsupported-surface, malformed-input, and source-privacy controls are named in the phase verification. | WIRED |
| Phase 34 surface manifest → Phase 35/36 admissions | `schway surface --json` is a versioned deterministic manifest; Phase 35 and 36 add admitted operators/call rules and corresponding unsupported codes. Manifest tests compare declared surface with actual admission and reject missing/extra entries. | WIRED |
| Phase 34 unsupported dynamic shift → Phase 35 BIT-02 | Phase 34 owns the dedicated diagnostic/manifest boundary; Phase 35 registers the shift refusal and rejects nonliteral or out-of-range amounts at checker/core/origin boundaries. This closes the explicit DX-22 deferred item. | WIRED |
| Phase 35 U64 bit operations and hex literals → Phase 36/37 source consumers | Syntax and literal ownership lower through check, core validation, origin validation, interpreter, and C/native paths. Phase 37’s `lib.crc32` uses `>>`, `&`, `^`, and literal values; the Phase 35 checksum consumer also uses `b << 16`. Independent literal expected answers and transfer-mutant controls cover value peers; the execution peer checks causal admission only. | WIRED |
| Phase 36 local/imported scalar calls → Phase 37 consumers | Calls lower to `OpCall`, are checked by the independent validation layers, and execute through interpreter/native backends. `examples/crc32/crc32.schway` imports `lib.crc32`; `crc_low.schway` imports and calls the same implementation. Tests assert shared module/adapter use and helper activations; there is no copied CRC implementation. | WIRED |
| Phase 37 CLI file route → checked application/session → FileBytes adapter | `schway run` selects interpreter/native application flow; session checks the source and module closure, admits the program, validates owner paths and manifest ABI, then invokes fixed file-byte read semantics. The adapter opens a regular file, bounds it at 65,536 bytes, and releases it; the program borrows a ByteView and passes it to shared `crc32`. | WIRED |
| Phase 37 CRC output → interpreter/native user result | The bit-at-a-time CRC module feeds bare decimal output. Pinned empty, NUL, standard vector, 65,536-byte, oversize, missing/unreadable cases are connected to oracle controls and CLI parity evidence. The second consumer proves imported reuse on both engines. | WIRED |
| Phase 37.1 producer traces → independent execution peer | Checked-source interpreter traces and explicit native terminal records are passed to `executionpeer.Validate`; actual indexed operation/type/route, activation-local provenance, caller edge, and root terminal are independently reconstructed. Model-only literal answers are separately tested and are not presented as peer numeric replay. | WIRED |

The inspected cross-phase graph contains no orphan API consumer or uncalled phase export on these paths. This repository has no HTTP route tree or authenticated web pages: its relevant public surface is the `schway` CLI and compiler library, so API-coverage and web-auth metrics are N/A rather than zero-feature defects.

## End-to-end flow inventory

| Flow | Trace | Outcome |
|---|---|---|
| Authoring diagnostics | CLI check → parser/recovery → checker → text or JSON diagnostic projection; Phase 34 manifest declares admitted and rejected forms; Phase 35 dynamic shift refusal is surfaced by the owned diagnostic. | WIRED. The user delegated acceptance of the existing Phase 34 diagnostic wording; this is not an independent readability measurement. Current report status remains `human_needed` until its owner refreshes it. |
| Bitwise checksum computation | Source literals/operators → typed core and independent validators → interpreter/native execution → independently authored expected answers; Adler32 uses the admitted `b << 16` form. | WIRED. Causal peer evidence does not prove scalar numeric results. |
| Scalar helper execution | Local/imported scalar call → `OpCall` contract validation → interpreter/native call/return/defect behavior → call-graph refusal and mutation controls. | WIRED. Phase 37 exercises imported calls with CRC32. |
| CRC file application | CLI grammar/path → checked source/module closure → interpreter reader or validated native application build → bounded FileBytes acquisition/ByteView → shared CRC helper → decimal stdout; invalid, missing, unreadable, and oversize paths are covered. | WIRED. The current evidence is host-local and is not a two-host receipt. |
| Checked typed failure | Checked `examples/phase24/error.schway` foreign failure → indexed `OpForeignCall` and checked contract/type facts → child `OpFail` → caller’s checked error edge → root `OpFail` → peer attribution. | WIRED for this actual checked-source route. The old synthetic `OpReturn` failed probe is not source evidence and does not support a universal claim about arbitrary raw-core returns. |
| No-call scalar defect reporting | Checked `examples/sum_to_n.schway` → native conformance shell at O0/O3/O3-LTO → structured terminal consumed by peer; application shell separately yields exit 65, empty stdout, exact reason stderr. Counted UTF-8/control/NUL reason bytes are preserved; malformed UTF-8 is refused before JSON cloning. | WIRED for the bounded admitted no-call defect. Unit native conformance remains refused. |
| Final hosted evidence binding | Generic freeze/receipt machinery → hosted ubuntu/macOS receipt set bound to final M007 digest. | PENDING by design: EVD-17 belongs to Phase 38, which has not started. This is an expected later owner gate, not an orphan or new regression in Phases 33–37.1. |

The six completed flows are the scoped compiler/user flows above; the hosted evidence binding is a separate pending milestone-terminal obligation. These counts describe the named audit inventory and are not general codebase coverage percentages.

## Terminal dispositions

| Audit obligation | Disposition | Evidence boundary |
|---|---|---|
| CR-03 — legacy and dynamic child defect closes all active invocations | CLOSED / WIRED | Actual checked-source interpreter traces exercise both invocation styles, ordinary one-frame return/failure, forged terminals, and late events after whole-execution closure. The claim is causal stack closure; it does not claim destructors, synthetic unwind records, or physical resource cleanup. |
| CR-04 — modeled foreign failure is attributed to checked route | CLOSED / WIRED | Positive evidence is `examples/phase24/error.schway`: checked `foreign.failed` at indexed `OpForeignCall`, source/type/contract facts, precise caller error route, activation-local provenance, and root `OpFail`. Literal model outcomes are independent expected inputs. No scalar payload or missing `Outcome.Value` authentication, numerical replay, or executed-CFG-branch proof is claimed. |
| CR-06 — no-call scalar defect is admitted only in structured native conformance | CLOSED / WIRED | Bounded checked `sum_to_n` case is emitted and run at three optimization settings, with peer-accepted terminal record. Application exit/stdout/stderr tuple remains separately asserted; Unit conformance remains refused. This does not admit or repair arbitrary raw `OpReturn` failures. |

For defect reasons, valid UTF-8 including control and NUL bytes is counted and preserved through the native reason route. Malformed reason bytes are refused at source and independent raw-core admission before JSON clone. The core semantic change corrects the WR-02 comment at the necessary edit site while dated amendments preserve earlier evidence and history.

## Requirement integration map

See `REQUIREMENTS.json` for the machine-readable three-source join of requirement ownership/checkboxes, verifier rows, and SDK-extracted summary fields. The requirements with no later feature consumer are QLT-13 and QLT-15 (phase-local carried-debt obligations) and QLT-14’s phase-local review-debt triage; QLT-14’s WR-02 explanation is later corrected as a narrow source-semantics clarification. These are self-contained obligations, not missing integrations. EVD-16 is a workflow foundation. EVD-17 is the single pending Phase 38 requirement.

## Verification state and limits

Current predecessor verification snapshots are Phase 33 passed 60/60, Phase 34 human_needed 38/38 technical, Phase 35 passed 38/38, Phase 36 passed 32/32, and Phase 37 passed 14/14. Their Nyquist validation frontmatter is `validated`, `nyquist_compliant: true`, and `wave_0_complete: true`. Phase 38 has no directory/validation because it is not started. Phase 37.1’s recorded 19/22 `gaps_found` report lists three workflow truths that share affected-report freshness and this fresh M007 audit; it is not a product-flow failure. The predecessor refreshes are now recorded; this audit supplies the requested M007 disposition, while parent-owned final Phase 37.1 verification and status refresh remain outstanding.

The Phase 37.1 bounded regression refresh reports 250 selectors, 251 root execution events, 240 distinct names across 16 groups: 220 prior green root receipts reused, 5 formerly failing roots rerun green, and 26 previously unrun roots passed; zero fail, skip, or missing. The capture-pressure selector was expressly excluded. Separately, four current terminal-proof subjects each bind 37 source paths; the same finite union covers 19 roots and 254 passing subtests with zero skips. These are finite local evidence lanes, not a whole-project/whole-package rerun, a complete replay of all historical roots, a pressure or stress result, a Linux/hosted/latency claim, or a hermetic clean-checkout run. Historical summaries, UAT, corpora, goldens, and pre-closure audit were preserved.

This audit itself performed source/evidence inspection, SDK summary extraction, repository status/head checks, and hash/integrity comparisons described in `EVIDENCE.json`; it did not run tests. Host and tool snapshot: Darwin arm64, macOS 26.6.2, Go 1.24.0, Apple clang 21.0.1, Node 26.10.0.
