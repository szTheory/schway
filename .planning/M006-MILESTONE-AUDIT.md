---
milestone: M006
milestone_name: Bounded Byte Utility and Local Modules
audited: 2026-10-12T23:59:00Z
audited_head: c5703ce8
evidence_revision: 6c3042713fadab198d827cb259eb7705431055e1
status: tech_debt
scores:
  requirements: 16/16
  phases: 5/5
  integration: 9/15 seams connected clean; 6 non-blocking (1 deferred by decision, 5 warnings); 0 broken
  flows: 11/12 complete; 1 misclassified exit (run on parameterised main, tracked todo)
gaps:
  requirements: []
  integration: []
  flows: []
process_gaps:
  - phase: 28-exit-classes-verify-report-and-entry-pruning
    item: "No 28-VALIDATION.md while workflow.nyquist_validation=true"
    action: "/gsd-validate-phase 28 (optional before completion)"
  - phase: 28-exit-classes-verify-report-and-entry-pruning
    item: "No 28-SECURITY.md while workflow.security_enforcement=true"
    action: "/gsd-secure-phase 28 (optional before completion)"
  - phase: 29-byte-views-and-the-adler-32-library
    item: "29-REVIEW-DISPOSITION.md leaves WR-01..WR-04 and IN-01..IN-04 at disposition: open (never triaged; 29-VERIFICATION classifies all as non-blocking)"
    action: "Triage into 29-REVIEW-DISPOSITION.md (fixed/deferred/accepted)"
  - phase: 31-bounded-file-checksum-and-second-consumer
    item: "No 31-*-SUMMARY.md lists requirements-completed (APP-11, APP-12, MOD-06, EVD-14); 3-source matrix gives 'partial' until verified manually"
    action: "Verified manually in this audit (see Requirements). Summaries left untouched because they are covered_files of 31-VERIFICATION.md."
nyquist:
  compliant_phases: ["29", "30", "31", "32"]
  partial_phases: []
  not_validated_phases: []
  missing_phases: ["28"]
  overall: partial
tech_debt:
  - phase: cross-phase (interpreter engine leg)
    items:
      - "Interpreter parity for file-bytes programs is enforced only at Go level (interp.RunWithFileBytes, interp.go:329, called only from tests); no CLI route runs a file-bytes program on the interpreter, and the Phase 32 guide test ignores the `engines` field of checksum.expected.json. Deferred by decision in 31-CONTEXT.md <deferred> ('run --engine=X FILE -- INPUT ... every engine is met at Go level, as in Phases 29 and 30'). (APP-10, APP-11, EVD-14, BYTE-02)"
      - "session.RunInterpreterApplication (session.go:1181) has no production caller; used only by Phase 27 tests. Pre-M006, carried forward."
  - phase: cross-phase (CLI routes)
    items:
      - "`run --engine=interpreter|native` on a parameterised main (e.g. examples/checksum/checksum.schway) exits 3 with generic tool.run_failed / native.tool_failure, while the guide calls it unsupported (examples/checksum/README.md:146). Tracked: todos/pending/2026-10-12-run-parameterised-main-reports-operational-failure.md. (DX-17)"
      - "`schway evidence` (EvidenceCommandFile, session.go:1642) does not load modules and refuses import-bearing sources with module.imports_not_loaded (exit 2, named). Module digests still reach build receipts and application evidence. `interface` and `query` are similarly not module-aware. (MOD-03)"
      - "Usage text omits the stats route and build --manifest. Tracked: todos/pending/2026-10-12-usage-text-omits-stats-and-build-manifest.md. (DX-19)"
  - phase: 28-exit-classes-verify-report-and-entry-pruning
    items:
      - "Latent: executionpeer.resolveEntry (executionpeer.go:272) picks by largest reachable set and ignores `main`, unlike session resolveEntry (session_entry.go:34). Safe today because the peer only sees the pruned program (session.go:1327 before compare at :1395); diverges if any caller passes an unpruned program. (DX-18)"
      - "WR-01 deferred: line-number pins in the Phase 16 emitter inventory; mitigation is running session meta-tests after session.go edits."
  - phase: 29-byte-views-and-the-adler-32-library
    items:
      - "WR-01: fixture-owner admission advertised as test-only but user-reachable (FFI trust class; not a security boundary)."
      - "WR-02: `fn main(b: ByteView)` passes check but no engine runs it (needs a typed refusal)."
      - "WR-03: native aborts exit 65 on any defect with a live fixture owner; interpreter errors differently for non-OpBytesAt defects (closed fixture route only)."
      - "WR-04: law-disable switches and a C-source rewrite hook in production packages."
      - "IN-01..IN-04: unreachable fixture defense in check; no operand-type law for byte ops in check's scalar fixed point; IN-04 README note that run/evidence do not support the fixture owner."
  - phase: 30-local-modules
    items:
      - "IN-01 deferred: test seams in production packages (same class as 29 WR-04)."
      - "IN-02 deferred: test hygiene; fix when the file is next touched."
      - "Known, tracked: debugmap.Build panics on 5 committed sources (pinned in golden); explain/query unknown-ID limitation on import-bearing sources."
  - phase: 31-bounded-file-checksum-and-second-consumer
    items:
      - "IN-02 deferred: 65,536-byte cap enforced by both host byte routes, not inside RunWithFileBytes."
      - "IN-03 deferred: duplicated module-name trimming in two packages."
  - phase: 32-clean-checkout-guide-and-final-receipts
    items:
      - "WR-01, WR-02, IN-01..IN-05 deferred past the receipt freeze; all fail closed. Tracked: todos/pending/2026-10-12-phase32-review-hardening-after-m006.md."
---

# Milestone M006: Bounded Byte Utility and Local Modules

**Verdict: tech_debt.** All 16 requirements are satisfied, all 5 phases passed verification, and no critical gap blocks completion. The debt is non-blocking: one deliberately deferred engine route, four CLI-route warnings (three tracked as todos), untriaged Phase 29 review warnings, and missing Phase 28 Nyquist and security artifacts.

## Phases

| Phase | VERIFICATION | Score | UAT | SECURITY | VALIDATION |
|---|---|---|---|---|---|
| 28 Exit Classes, Verify Report, Entry Pruning | passed (re-verified after G-28-3) | 5/5 | yes | **missing** | **missing** |
| 29 Byte Views and the Adler-32 Library | passed | 5/5 | yes | yes | validated, compliant |
| 30 Local Modules | passed | 4/4 | yes | yes | validated, compliant |
| 31 Bounded File Checksum and Second Consumer | passed | 4/4 | yes | yes | validated, compliant |
| 32 Clean-Checkout Guide and Final Receipts | passed | 3/3 | yes (27/27 automated) | yes (threats_open 0) | validated, compliant |

## Requirements (3-source cross-reference)

| REQ | Phase | VERIFICATION | SUMMARY requirements-completed | REQUIREMENTS.md | Final |
|---|---|---|---|---|---|
| DX-17 | 28 | SATISFIED | 28-01, 28-02 | [x] | satisfied |
| DX-18 | 28 | SATISFIED | 28-01 | [x] | satisfied |
| EVD-13 | 28 | SATISFIED | 28-03 | [x] | satisfied |
| BYTE-01 | 29 | SATISFIED | 29-00..03, 10, 11 | [x] | satisfied |
| BYTE-02 | 29 | SATISFIED | 29-00, 01, 04, 07, 08, 10..12 | [x] | satisfied |
| BYTE-03 | 29 | SATISFIED | 29-03, 06, 09, 11 | [x] | satisfied |
| APP-10 | 29 | SATISFIED | 29-00, 01, 05..08, 10..12 | [x] | satisfied |
| MOD-03 | 30 | SATISFIED | 30-00, 01, 03, 04, 08, 09 | [x] | satisfied |
| MOD-04 | 30 | SATISFIED | 30-00, 05, 06, 08, 09 | [x] | satisfied |
| MOD-05 | 30 | SATISFIED | 30-00..09 | [x] | satisfied |
| APP-11 | 31 | SATISFIED | **missing** | [x] | satisfied (manual verification) |
| APP-12 | 31 | SATISFIED | **missing** | [x] | satisfied (manual verification) |
| MOD-06 | 31 | SATISFIED | **missing** | [x] | satisfied (manual verification) |
| EVD-14 | 31 | SATISFIED | **missing** | [x] | satisfied (manual verification) |
| EVD-15 | 32 | SATISFIED | 32-06, 08, 11 | [x] | satisfied |
| DX-19 | 32 | SATISFIED | 32-01, 07, 11 | [x] | satisfied |

Orphans: none. All 16 traceability rows appear in a phase VERIFICATION.md.

**Manual verification of the four Phase 31 "partial" rows** (SUMMARY metadata missing):
- `SCHWAY_REQUIRE_NATIVE=1 go test ./internal/compiler/... ./cmd/schway/... -run TestPhase31 -count=1 -v`: 25 top-level tests PASS, 0 SKIP, 0 FAIL.
- The integration checker ran the built CLI and observed: `app run` checksum on `hello` = 103547413 and on an empty file = 1; a 65,537-byte file → exit 65 `FileError.FileTooLarge`; a missing path or a directory → exit 65 `FileError.FileUnavailable`; `adler32_cli` builds with module rows and sha256 digests in its receipt.

## Integration (live at c5703ce8)

**Full suite:** `SCHWAY_REQUIRE_NATIVE=1 go test ./... -count=1 -timeout=30m` exits 0, with 27 packages ok and 0 FAIL. This includes `TestPhase32ReceiptLivingClaims`, which re-binds hosted run 38315799626 (headSha 6c304271) to the current code digest.

The integration checker (gsd-integration-checker) verified:
- **Connected (9):** modload → check/build/app verify/run-native; module digests → receipt and evidence; Phase 28 resolveEntry → every engine; DX-18 pruning → native build; checksum native route; Phase 32 freeze and codedigest path sets (they include `examples/`); public-route audit → test owners; exit table → CLI.
- **Non-blocking (6):**
  - interpreter file-bytes leg is test-only (deferred by decision in 31-CONTEXT);
  - `run` on a parameterised main exits 3 with a generic code (todo);
  - `evidence` is not module-aware;
  - latent divergence between the two entry resolvers (peer vs session);
  - orphaned `RunInterpreterApplication` (pre-M006).

  The checker rated the interpreter leg a BLOCKER. This audit downgrades it: APP-10 requires interpreter = native = pinned literals, which Phase 29 pins at Go level, and Phase 31 CONTEXT deliberately deferred a public interpreter input route because APP-11/APP-12 do not require it.

| E2E flow (built CLI) | Result |
|---|---|
| `check` modular entries (checksum, adler32_cli) | complete, exit 0 |
| `build --manifest` adler32_cli → receipt with module digests | complete |
| `app run` checksum on `hello`, empty file | complete (103547413, 1) |
| `app run` oversize 65,537 bytes | complete, exit 65 FileTooLarge |
| `app run` missing path, directory | complete, exit 65 FileUnavailable |
| `check` on a missing source | complete, exit 3 (IO) |
| missing import | complete, exit 2 `module.not_found` |
| module case mismatch | complete, exit 2 `module.case_mismatch` |
| type error in imported library | complete, diagnostic carries `unit=adler32` |
| `evidence` on a module program | named refusal, exit 2 `module.imports_not_loaded` |
| Phase 32 guide from a clean checkout | complete (TestPhase32GuideCleanCheckout) |
| `run` on parameterised main | **misclassified:** exit 3 generic, not a named refusal (todo) |

Not re-exercised by the checker: the module cycle, the MOD-06 mutation flip, and the BYTE-02 OOB on every engine and opt level. Each is pinned by tests in the green full suite (Phase 30, 31 and 29 suites respectively).

## Tech debt

See `tech_debt` in the frontmatter: 18 items across 7 areas. Three items already have pending todos. None blocks completion. The ones worth a follow-up phase:
1. A named refusal and exit 2 for `run` on a parameterised main (DX-17 consistency).
2. Triage of Phase 29 WR-02 and WR-03.
3. A module-aware `evidence`.
4. A public interpreter input route, if interpreter parity should be user-visible.

## Nyquist coverage

| Phase | VALIDATION.md | Compliant | Action |
|---|---|---|---|
| 28 | missing | n/a | `/gsd-validate-phase 28` |
| 29 | exists | true | none |
| 30 | exists | true | none |
| 31 | exists | true | none |
| 32 | exists | true | none |
