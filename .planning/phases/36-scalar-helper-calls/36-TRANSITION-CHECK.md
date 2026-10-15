---
phase: 36-scalar-helper-calls
date: 2026-10-15
status: complete
next_phase: 37
---

# Phase 36 transition check

`phase.complete 36` ran once after canonical goal verification passed 32/32.
It completed 8/8 plans, closed CALL-01/CALL-02, and advanced state to Phase 37.
This run stops here under `--no-transition`; Phase 37 has not been started.
The current-position validator required a matching directory for the next
phase, so an empty `.gitkeep` scaffold was added for Phase 37. It contains no
context, plan or implementation work.

The completion query reported two apparent missing paths in Plan 05's summary:
a brace abbreviation for three existing RED receipts and the shortened path
to `examples/phase24/error.schway`. Both references were expanded to actual
files. This corrects publication metadata; it changes neither the historical
failure description nor compiler behavior. The verifier refreshes the report's
covered-input fingerprint after that spelling correction, without replaying
implementation, UAT or phase.complete.

## Evidence and limits

The final feature capture binds 45 roots/12 packages to source/test revision
`979357388171339005fd6dbd65aa760810ae9384`; the prior-phase regression capture
binds 80 compiled root occurrences/13 packages to
`b51043c6566f793893c8c4e143493e7ec3af5bcd`. The 52-file independent review is
clean with its warning fixed; the seven-file proof-bridge supplement is
reviewed separately. All 24 planned L1 threats are closed. Both canonical
prohibition producers independently proved bad-overlay RED, clean GREEN and
default GREEN. Actual commands, output hashes and limits remain in
`36-ACCEPTANCE-EVIDENCE.md`, `36-REGRESSION.md`,
`36-PROHIBITION-EVIDENCE.md` and `36-VERIFICATION.md`.

These are bounded local Darwin checks, not full suites or hosted receipts.
Source inspection, new test runs and historical receipts remain distinct.
Numeric results belong to engine tests; source argument/result intent belongs
to the checker; causal admission proves structural attribution. No subjective
readability result, complete-loop cold/warm distribution, full CRC library,
new backend or runtime dependency is claimed. The SPEC-less classifier's
unclassified rows remain historical planning metadata; explicit CALL contracts
were verified against actual source and behavior.

## Three next capabilities

| Program | Current blocker | Smallest complete slice | Checker/guarantee changes | Evidence/debt | Owner/next action | Priority changes if |
|---|---|---|---|---|---|---|
| File CRC-32 and a second importing consumer | Only the reflected-bit step helper exists; there is no complete file CRC library/second consumer | Compose ordinary-source CRC-32/ISO-HDLC with the bounded file route and a second module consumer | Reuse checked bitwise/call contracts; independently establish full CRC answers, file bounds and error cleanup | APP-13, MOD-07; helper answers in Phase 36, existing checksum/file witnesses; no new arithmetic assumed | Phase 37 discussion/planning, then one runnable vertical slice | Independent vectors disagree, or a named file consumer demonstrates a missing semantic operation |
| Interpreter file CLI and executable CRC guide | File-bytes parity is not exposed through the public interpreter CLI; the new CRC guide is absent | Run the same bounded file inputs through both public engines and execute every guide command | Preserve module-aware admission, required input, file outcome/cleanup distinctions and exact exit/stdout/stderr | APP-14, DX-24; current Go-level file/ByteView evidence and Phase 32 guide pattern | Phase 37, alongside the CRC consumer; keep user-facing commands and executable guide together | CLI failure tuples differ, or the guide cannot reproduce a pinned consumer result |
| Audited CRC/checksum acceptance on Ubuntu and macOS | No final M007 receipt is bound to the audited source | After Phase 37 and the milestone audit, bind exact two-host logs to one source/fixture revision | Recheck portability, non-skipped roots and independent answers without transferring local-host proof | EVD-17; historical M006 receipts stay historical | Phase 38 after `$gsd-audit-milestone` and any resulting gap closure | Host drift, source mismatch or platform-specific failure invalidates the proposed binding |

## Triggered seed review

Read the living roadmap/maturity records, seed trigger metadata and
`sources/DASHBIT-AI-ERA.md` at this transition. Review was limited to relevant
seeds and current witnesses; no repeated ecosystem study was needed.

| Seed | Disposition | Current witness and named consumer | Trigger observation | Next action |
|---|---|---|---|---|
| z1g | retain dormant | `session_phase6_query.go#QueryCommandFile`; a reviewer locating `crc_step` and its local/imported callsites can use ordinary source search | No measured lookup accuracy/material-cost gap was established | A Phase 37 reviewer must name and measure a failed lookup before a bounded symbol/caller experiment |
| trz | retain dormant | `native_app.go#readApplicationCapture`, `session_app_verify.go#VerifyApplicationCasesFile`; helper capture and later-defect controls explain the observed outcomes | No unexplained named failure remains; Phase 33 fixed the earlier unbounded allocation concern | Reconsider only for a concrete CRC/file failure that current capture, observers and replay cannot explain |
| z1s | retain dormant | `syntax/format.go#Format`; helper parser/formatter tests and the runnable importer | No real consumer/reviewer source-form dispute was established | Require paired forms, identical obligations and real review evidence before an experiment |

## Post-completion workflow

No Phase 36 handoff or matching pending todo required cleanup. Global learnings
are disabled. Graduation is enabled but there are no active LEARNINGS artifacts
in its window, so the fewer-than-five-items guard skips it. Flat workstream
mode excludes collision checks. PROJECT and STATE receive the current focus,
three decisions, remaining owners and session continuity. The temporary auto
chain flag is reset, preserving the user's effort routing and worktree setting.

The live roadmap and maturity documents receive a dated post-verification
amendment; earlier implementation-time pending statements remain historical.
Next command: `$gsd-discuss-phase 37`.
