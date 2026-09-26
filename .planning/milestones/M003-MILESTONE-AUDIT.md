---
milestone: M003
milestone_name: Computation and Honest Instruments
audited: 2026-09-26T19:19:10Z
status: tech_debt
scores:
  requirements: "33/33 satisfied; 0 unsatisfied; 0 orphaned"
  phases: "7/7 passed; 85/85 plans complete"
  integration: "6/6 milestone paths wired; 0 broken; 2 non-blocking hardening warnings"
  flows: "6/6 complete; 0 broken"
gaps:
  requirements: []
  integration: []
  flows: []
partial:
  requirements: []
tech_debt:
  - phase: "14, 17, 18"
    items:
      - "Nyquist validation is partial: each VALIDATION.md is marked validated but has nyquist_compliant: false and wave_0_complete: false."
  - phase: "16"
    items:
      - "The private cgen.emitBranchOperations helper has no call sites; public Emit and EmitNative both route through emitProgram."
  - phase: "19"
    items:
      - "19-REVIEW.md retains WR-01: uppercase radix prefixes lex but fail checking."
      - "19-REVIEW.md retains WR-02: interpreter input accepts a leading plus that generated native input rejects."
  - phase: "20"
    items:
      - "The live PRC-02 population test pins four current open-unowned IDs, while the five-item cap predicate is tested separately rather than applied to that live set."
      - "QLT-12 was manually verified from Plan 20-04, Plan 20-08, and 20-CLOSURE-TIMING.md because no SUMMARY requirements-completed frontmatter lists it."
      - "Four open-unowned items remain within the PRC-02 limit: D-10-C04, D-12-43, D-14-46, D-14-47."
nyquist:
  compliant_phases: ["15", "16", "19", "20"]
  partial_phases: ["14", "17", "18"]
  not_validated_phases: []
  missing_phases: []
  overall: partial
scope:
  included_phases: ["14", "15", "16", "17", "18", "19", "20"]
  excluded_phase: "21 — assigned to M004 in ROADMAP.md"
closeout_preflight:
  open_items: 10
  debug_sessions: 8
  quick_tasks: 2
  uat_gaps: 0
  verification_gaps: 0
---

# M003 — Computation and Honest Instruments — Milestone Audit

**Audited:** 2026-09-26  
**Scope:** M003 Phases 14–20  
**Status:** tech_debt — all 33 requirements are satisfied and all seven phase verifications pass. No critical requirement, integration, or flow gaps remain. Nyquist coverage and several non-blocking hardening items remain for review before or during M004.

## Scope and Milestone Intent

ROADMAP.md defines M003 as Phases 14–20. The milestone proves that Lang can create a value through computation and carry it through checking, execution, and native emission, while its evidence tools distinguish exercised behavior from work that is merely wired.

The generic phase scan includes the Phase 21 directory and init.milestone-op reports eight directories. Phase 21 is explicitly assigned to M004 in ROADMAP.md, so it is excluded here. Its verification fingerprint remains stale; its UAT is already complete at 7/7. Reconcile it when M004 is formally opened, without replaying completed UAT.

## 1. Requirements Coverage — Three-Source Cross-Reference

All 33 M003 requirement IDs were compared across REQUIREMENTS.md traceability and checkboxes, their phase VERIFICATION.md tables, and SUMMARY.md requirements-completed frontmatter. Every requirement has a passing phase verification. Thirty-two appear in summary frontmatter. QLT-12 is the sole frontmatter omission; its Plan 20-04 and Plan 20-08 artifacts, closure timing record, and Phase 20 verification were manually checked and support the requirement. No requirement is orphaned from phase verification.

| Phase | Requirements | Verification | Summary frontmatter | Audit result |
|---|---|---|---|---|
| 14 | EVD-01–08, DX-08, DX-09, PRC-01 | 11/11 satisfied; passed | Present across Plans 14-01–14-13 | Satisfied |
| 15 | OBS-01–04, NAT-10 | 5/5 satisfied; passed | Present across Phase 15 summaries | Satisfied |
| 16 | NAT-08, NAT-09 | 2/2 satisfied; passed | Present across Phase 16 summaries | Satisfied |
| 17 | TYP-01–05 | 5/5 satisfied; passed | Present across Phase 17 summaries | Satisfied |
| 18 | CTL-01–03 | 3/3 satisfied; passed | Present across Phase 18 summaries | Satisfied |
| 19 | VAL-01–03 | 3/3 satisfied; passed | Present across Phase 19 summaries | Satisfied |
| 20 | QLT-10, QLT-11, QLT-12, PRC-02 | 4/4 pass | QLT-10, QLT-11, PRC-02 listed; QLT-12 manually verified | Satisfied |

The audit found QLT-10, QLT-12, and PRC-02 checked off as pending in REQUIREMENTS.md despite Phase 20's passing verification. Those three rows are now marked [x] / Complete. QLT-12's missing summary frontmatter remains disclosed as documentation debt; no historical summary was rewritten.

S-010 is a Phase 18 entry gate, not one of the 33 M003 requirements. It is recorded complete and is excluded from the requirement count.

| Requirement group | Result | Evidence |
|---|---|---|
| Phase 14 evidence, diagnostic, and ownership controls | 11/11 satisfied | [Phase 14 verification](M003-phases/14-evidence-instrument-and-honest-scoping/14-VERIFICATION.md) |
| Phase 15 event identity and diamond | 5/5 satisfied | [Phase 15 verification](M003-phases/15-event-identity-lang-execution-2/15-VERIFICATION.md) |
| Phase 16 emitter law and explicit cuts | 2/2 satisfied | [Phase 16 verification](M003-phases/16-branch-match-emitter-port/16-VERIFICATION.md) |
| Phase 17 distinct return types | 5/5 satisfied | [Phase 17 verification](M003-phases/17-return-type-parameter-type/17-VERIFICATION.md) |
| Phase 18 computed branches and payloads | 3/3 satisfied | [Phase 18 verification](M003-phases/18-branch-on-a-computed-value/18-VERIFICATION.md) |
| Phase 19 numeric literals and OpConst | 3/3 satisfied | [Phase 19 verification](M003-phases/19-numeric-literals-and-opconst/19-VERIFICATION.md) |
| Phase 20 archive, frontier, cache, timing, and debt cap | 4/4 satisfied | [Phase 20 verification](M003-phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-VERIFICATION.md) |

**Totals:** 33 satisfied, 0 partial after manual review, 0 unsatisfied, 0 orphaned. The fail gate does not fire.

## 2. Phase Verifications

| Phase | Name | Plans | Verification score | Status |
|---|---|---:|---:|---|
| 14 | Evidence Instrument and Honest Scoping | 13 | 11/11 requirements | passed |
| 15 | Event Identity (lang.execution/2) | 10 | 7/7 must-haves | passed |
| 16 | Branch/Match Emitter Port | 26 | 4/4 must-haves | passed |
| 17 | Return Type ≠ Parameter Type | 9 | 5/5 must-haves | passed |
| 18 | Branch on a Computed Value | 10 | 30/30 must-haves | passed |
| 19 | Numeric Literals and OpConst | 7 | 4/4 must-haves | passed |
| 20 | Nyquist, D-13-34, and the Frontier Fixture | 10 | 4/4 requirements | passed |

All seven reports exist and pass. All 85 plans have completed summaries. Updating REQUIREMENTS.md made four covered-file fingerprints stale, so the canonical Phase 15, 16, 18, and 19 verification reports were refreshed. No plans or UAT were replayed, and the full suite was not rerun; the Phase 16 and 19 refreshes also ran focused checks.

The post-refresh GSD resolver reports Phases 14–20 complete with passing verification. It still sees Phase 21 as stale, but ROADMAP.md assigns Phase 21 to M004, so that result does not change M003 scope or its next command.

## 3. Cross-Phase Integration

The GSD integration checker traced six representative milestone connections. All six reach a production or recurring-evidence consumer; no expected compiler flow is broken. API-route and authentication checks are not applicable to this Go compiler and command-line toolchain.

| Connection | Requirements | Result | Evidence path |
|---|---|---|---|
| Phase 14 evidence and debt machinery → Phase 20 reconciliation and CI | EVD-01–08, PRC-01, QLT-10 | WIRED | Groundedness/witness scans and archive reconciliation run in the Go suite; CI runs go test ./... and go test -race ./... on Linux and macOS. |
| Phase 15 event identity → Phase 16 public native emission | OBS-01–04, NAT-10 | WIRED | Shared-leaf diamond is compared across interpreter and three native tiers; the independent peer re-derives occurrence paths. |
| Phase 17 return facts → checker, validators, interpreter, and C emitter | TYP-01–05 | WIRED | Distinct return types reach four-tier execution; mismatch/refusal and mutation controls cover negative behavior. |
| Phase 18 computed branch and payload → execution tiers | CTL-01–03 | WIRED | A callee-produced match scrutinee and destructured payload reach interpreter/native outcomes with wrong-slot mutation controls. |
| Phase 19 OpConst → semantic dispatch and native execution | VAL-01–03 | WIRED | Literal source reaches parser/checker/core, independent consumers, interpreter, and -O0, -O3, and -O3 -flto. |
| Phase 20 cache and evidence → enumerated closure and archive | QLT-10–12, PRC-02 | WIRED, with warnings below | The 112-program closure records 112 cold recomputations and 112 warm reuses while executing and comparing current results. |

### Non-Blocking Integration Warnings

- **PRC-02 enforcement:** the live population test pins the current source-derived set, and the separate cap test covers five and six items. The live population test does not call the cap predicate. Phase 20 verification records four current open-unowned items, within the limit of five. The requirement is met for the recorded population, but the two guards are not joined into one assertion. See session_test.go around lines 3220–3263.
- **NAT-08 dead helper:** cgen.emitBranchOperations is defined but has no call sites. Public Emit and EmitNative route through emitProgram, so this is an orphaned private helper rather than a second production emitter. See internal/compiler/cgen/cgen.go:729 and its call-site search.
- **QLT-12 summary metadata:** no phase summary's requirements-completed frontmatter lists QLT-12. The requirement is supported by the cache behavior in Plan 20-04, the paired suite/closure timing in Plan 20-08, and 20-CLOSURE-TIMING.md; it was manually checked for this audit.

### End-to-End Flows

| Flow | Result | Representative evidence |
|---|---|---|
| Numeric literal → parser/checker/core → interpreter/native tiers | Complete | TestPhase19FourTierLiteral; malformed and overflow refusals |
| Distinct return/parameter types → admission → interpreter/native | Complete | TestPhase17TwoTypeFourTierDifferential and emitter mutation control |
| Callee Result → computed match → terminal outcome | Complete | TestPhase18ResultComputedMatch |
| Destructured payload → returned value → negative mutation | Complete | TestPhase18WrongSlotMutation and long-tag counterpart |
| Shared-leaf invocation → /2 event identity → independent peer | Complete | Phase 15 diamond and collision-negative controls |
| Evidence/debt source → Phase 20 archive/cache/timing reconciliation | Complete | Groundedness, validation lifecycle, closure cache and timing evidence |

## 4. Nyquist Coverage Discovery

The validate-phase step hook is active. All seven M003 phases have a VALIDATION.md; none is missing or remains status: draft.

| Phase | Validation state | Nyquist | Audit class |
|---|---|---:|---|
| 14 | validated | false | PARTIAL |
| 15 | complete | true | COMPLIANT |
| 16 | validated | true | COMPLIANT |
| 17 | validated | false | PARTIAL |
| 18 | validated | false | PARTIAL |
| 19 | validated | true | COMPLIANT |
| 20 | complete | true | COMPLIANT |

Overall Nyquist status is **partial**. Phases 14, 17, and 18 are coverage follow-ups; this discovery does not invalidate their passing phase verifications or auto-run $gsd-validate-phase. Phases 15 and 20 retain the legacy status value complete rather than validated, but both have nyquist_compliant: true and wave_0_complete: true; they are treated as compliant based on those affirmative flags and their passing phase evidence.

## 5. Tech Debt and Deferred Items

- **Phases 14, 17, and 18:** Nyquist coverage is partial as listed above.
- **Phase 16:** remove or otherwise account for the unused private emitBranchOperations helper if the one-emission-law cleanup is revisited.
- **Phase 19:** 19-REVIEW.md retains two advisory input-parsing findings: uppercase radix prefixes parse but fail checking, and interpreter U64 input accepts a leading plus that generated native input rejects. The Phase 19 verifier marks these advisory and confirms that they do not invalidate its acceptance criteria.
- **Phase 20:** join the current debt-population assertion to the five-item cap predicate. The current verified set is four: D-10-C04, D-12-43, D-14-46, and D-14-47. D-13-34 closed after Plan 07's five-item checkpoint. The requirement is satisfied; this is a guard-strengthening item.
- **Phase 20:** add QLT-12 to appropriate summary frontmatter in a future documentation repair if historical plan summaries are intentionally revised; the requirement was manually verified here and is checked complete.
- **M004 carry-forward:** the emitter families explicitly cut by Phase 16 remain refused with M004 ownership as recorded in NAT-09; this is an intentional scope boundary, not an M003 blocker.

## 6. Closeout Readiness

The read-only audit-open --json preflight on 2026-09-26 found 10 open GSD artifacts: eight debug sessions (seven diagnosed, one unknown) and two quick tasks (one missing status, one unknown). It found zero UAT gaps and zero verification gaps. These artifacts are outside the M003 requirement score, but $gsd-complete-milestone M003 will display them and require a resolve, acknowledge, or cancel choice before archiving. This audit did not delete or acknowledge them.

## Result and Next Command

M003 has met its requirements and phase gates with no critical blockers. The audit status is tech_debt because Nyquist coverage is partial and the non-blocking integration/documentation warnings above remain.

**Recommended next command:** $gsd-complete-milestone M003

That command moves M003 into archive/closeout and does not replay Phases 14–20. Its pre-close artifact gate will present the 10 open items noted above. If the user prefers to close validation coverage first, the exact validation commands are $gsd-validate-phase 14, $gsd-validate-phase 17, and $gsd-validate-phase 18; the milestone audit itself did not run them.
