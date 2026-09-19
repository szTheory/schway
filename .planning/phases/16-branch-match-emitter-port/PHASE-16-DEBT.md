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
| D-16-11 | 16-CONTEXT.md D-16-11; NAT-09 | NAT-09, D-16-08 | warning | UNOWNED(m004-native-emission-design) | DEFINED | n/a | `emitLinearForeign` is cut from M003. M004 must design resource discharge and foreign-boundary ownership before this family can enter `emitProgram`; one-TU pure-Lang output leaves `-flto` behaviorally inert. |
| D-16-12 | 16-05-SUMMARY.md cut-m004; D-16-07 | NAT-09, D-16-08 | warning | UNOWNED(m004-native-emission-design) | DEFINED | n/a | `emitLinearBorrowedByPointer` is cut from M003. M004 must provide discharge-pair semantics and required macOS/Linux evidence before reopening; one-TU pure-Lang output leaves `-flto` behaviorally inert. |
| D-16-13 | 16-05-SUMMARY.md cut-m004; D-16-11 | NAT-09, D-16-08 | warning | UNOWNED(m004-native-emission-design) | DEFINED | n/a | `emitLinearBorrowedByPointerPlain` is cut from M003. M004 must prove its shared-pointer/alias contract and discharge behavior before reopening; one-TU pure-Lang output leaves `-flto` behaviorally inert. |

## Detail

### D-16-11 — foreign linear bodies

first-recorded: M003

M004 owner: native-emission/ownership design workstream.

Prerequisite: a checked resource ledger and discharge-pair design that covers
foreign-call blocks and makes cleanup obligations explicit at the C boundary.

Reopening condition: M004 names a checked foreign ownership model and a native
witness demonstrating cleanup across every admitted foreign exit path.

Witness: the current named refusal in `emitProgram` and NAT-09's amendment;
there is no M003 execution witness because this family is deliberately cut.

`-flto` consequence: a foreign body is refused, while every admitted pure-Lang
program is one translation unit; `-flto` is behaviorally inert for that
one-TU subset and this row does not claim otherwise.

### D-16-12 — exclusive borrowed-by-pointer bodies

first-recorded: M003

M004 owner: native-emission/ownership design workstream.

Prerequisite: the full discharge-pair design plus successful exact-shape
macOS and Linux evidence; Plan 16-05 records Linux as unavailable, so its
macOS results cannot satisfy this prerequisite.

Reopening condition: M004 proves the exact read/copy-only contract over both
hosts and re-establishes a live structural fence before any program-emitter
admission is proposed.

Witness: `16-05-SUMMARY.md` records `cut-m004`, the Linux-unavailable lane,
and the refusal-fence control; `TestProgramBorrowedByPointerDisposition` pins
the resulting whole-program refusal.

`-flto` consequence: the probe is one translation unit and cannot establish
behaviorally non-inert LTO; pure-Lang one-TU emission remains inert evidence,
not authorization for this pointer lowering.

### D-16-13 — shared borrowed-by-pointer plain bodies

first-recorded: M003

M004 owner: native-emission/ownership design workstream.

Prerequisite: a specified shared-pointer alias and discharge contract; the
exclusive `restrict` probe cannot substitute for this different family.

Reopening condition: M004 supplies a family-specific ownership/alias proof,
cross-host witness, and a refusal fence that rules out unsupported pointer
extensions before a program-emitter admission is proposed.

Witness: Plan 16-05's `cut-m004` decision explicitly keeps every by-pointer
family unavailable in M003; the `emitProgram` named refusal is the M003
behavioral witness.

`-flto` consequence: pure-Lang program emission is one translation unit, so
the current `-flto` lane is behaviorally inert and does not prove a future
shared-pointer optimization or close D-16-08.
