---
phase: 16-branch-match-emitter-port
recorded: 2026-09-19
status: accepted
disposition: D-10-60-m004-cut
items: 3
blocking: 0
---

# Phase 16 — M004 Emitter-Family Debt Register

**Written:** 2026-09-19, after Plan 16-05's human-selected `cut-m004`
disposition. The required Linux lane is unavailable; passed macOS evidence and
the structural refusal fence therefore do not authorize a production
by-pointer admission. This register is the explicit D-10-60 amendment record,
not a third-deferral euphemism: every excluded family is cut from M003 and
assigned to the M004 native-emission/ownership design workstream.

One-TU pure-Lang `emitProgram` output leaves `-flto` behaviorally inert: it
does not create an inter-translation-unit boundary or prove a useful LTO
optimization. That honest consequence is distinct from, and does not close,
the D-16-08 non-inertness evidence obligation.

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Grade | Witness | Item |
|---|---|---|---|---|---|---|---|
| D-16-11 | 16-CONTEXT.md D-16-11; NAT-09 | NAT-09, D-16-08 | warning | P21 | EXERCISED | probe:TestPhase21ResourceDischargeContract, probe:TestForeignLandingPadEmitterRemainsRefused, probe:TestForeignResourceLedgerEmitterRemainsRefused | `emitLinearForeign` is retired and foreign program shapes remain refused. The machine-checkable contract at `21-RESOURCE-DISCHARGE-CONTRACT.json` classifies modeled exits but admits no family and proves no runtime cleanup. Reopening still requires checked foreign discharge proof across every admitted exit and a native witness; current refusal remains. The scoped Phase 21 LTO run is semantic equality for one fixture/host, not proof of optimizer inactivity or foreign behavior. |
| D-16-12 | 16-05-SUMMARY.md cut-m004; D-16-07 | NAT-09, D-16-08 | warning | P21 | EXERCISED | probe:TestPhase21ResourceDischargeContract, probe:TestProgramBorrowedByPointerDisposition | `emitLinearBorrowedByPointer` is retired and by-pointer program shapes remain refused. The machine-checkable contract at `21-RESOURCE-DISCHARGE-CONTRACT.json` admits no family and proves no runtime cleanup. Reopening still requires the full discharge-pair semantics plus exact-shape macOS/Linux evidence and a live refusal fence; the scoped Phase 21 LTO sample does not substitute for that evidence. |
| D-16-13 | 16-05-SUMMARY.md cut-m004; D-16-11 | NAT-09, D-16-08 | warning | P21 | EXERCISED | probe:TestPhase21ResourceDischargeContract, probe:TestProgramBorrowedByPointerDisposition | `emitLinearBorrowedByPointerPlain` is retired and shared by-pointer program shapes remain refused. The machine-checkable contract at `21-RESOURCE-DISCHARGE-CONTRACT.json` admits no family and proves no runtime cleanup. Reopening still requires the family-specific shared-pointer alias/discharge proof, cross-host witness, and refusal fence; exclusive-family evidence cannot substitute. The scoped Phase 21 LTO sample makes no claim about this family. |

## Detail

### D-16-11 — foreign linear bodies

first-recorded: M003

M004 owner: Phase 21 — Native Emission Ownership and Resource Discharge.

Prerequisite: a checked resource ledger and discharge-pair design that covers
foreign-call blocks and makes cleanup obligations explicit at the C boundary.

Reopening condition: M004 names a checked foreign ownership model and a native
witness demonstrating cleanup across every admitted foreign exit path.

Witness: the current named refusal in `emitProgram` and NAT-09's amendment;
there is no M003 execution witness because this family is deliberately cut.

`-flto` evidence: `21-LTO-EVIDENCE.md` records semantic equality for one
emitted multi-function fixture on Darwin arm64 with Apple Clang 21.0.0. This
does not measure optimizer activity or establish behavior for foreign bodies.

### D-16-12 — exclusive borrowed-by-pointer bodies

first-recorded: M003

M004 owner: Phase 21 — Native Emission Ownership and Resource Discharge.

Prerequisite: the full discharge-pair design plus successful exact-shape
macOS and Linux evidence; Plan 16-05 records Linux as unavailable, so its
macOS results cannot satisfy this prerequisite.

Reopening condition: M004 proves the exact read/copy-only contract over both
hosts and re-establishes a live structural fence before any program-emitter
admission is proposed.

Witness: `16-05-SUMMARY.md` records `cut-m004`, the Linux-unavailable lane,
and the refusal-fence control; `TestProgramBorrowedByPointerDisposition` pins
the resulting whole-program refusal.

`-flto` evidence: the Phase 21 receipt is bounded to its named pure-Lang
fixture, compiler, and host. It does not establish optimizer activity or
authorize this pointer lowering.

### D-16-13 — shared borrowed-by-pointer plain bodies

first-recorded: M003

M004 owner: Phase 21 — Native Emission Ownership and Resource Discharge.

Prerequisite: a specified shared-pointer alias and discharge contract; the
exclusive `restrict` probe cannot substitute for this different family.

Reopening condition: M004 supplies a family-specific ownership/alias proof,
cross-host witness, and a refusal fence that rules out unsupported pointer
extensions before a program-emitter admission is proposed.

Witness: Plan 16-05's `cut-m004` decision explicitly keeps every by-pointer
family unavailable in M003; the `emitProgram` named refusal is the M003
behavioral witness.

`-flto` evidence: the Phase 21 receipt does not test this family and does not
prove optimizer activity or close D-16-08.
