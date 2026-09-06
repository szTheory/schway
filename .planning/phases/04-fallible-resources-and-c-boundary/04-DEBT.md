---
phase: 04-fallible-resources-and-c-boundary
recorded: 2026-09-05
updated: 2026-09-05
code_head: df75205
status: accepted
disposition: carried-forward-by-design
items: 5
blocking: 0
---

# Phase 04: Debt Register

Recorded by plan 04-06 (Task 3), per D-04-26's instruction: the deferred half
of carried Phase 3 debt D-03-01, the deferred payload-carrying-alternative
work D-04-04 already named as debt rather than assumed away, and the accepted
residual limitations this phase inherits from its own decisions (D-04-CONTEXT
`<deferred>` §"Accepted residual limitations"). None of these are correctness
defects a program can observe executing today — each is recorded here, dated,
with an identifier, a severity, a source, and a landing phase, per the
standing rule that a deferred item is re-recorded rather than silently
dropped. Item D-04-32 was added by plan 04-07 (Task 3), the phase's closing
plan, recording a roadmap gap surfaced for Phase 5's own planning rather than
debt against this phase.

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-04-26 | 03-DEBT.md D-03-01 (carried), D-04-CONTEXT | D-05/D-02-03 | warning | Phase 5 | Retiring `discoverLoanLastUses` in favour of driving both admission and endpoint materialization from `loanLivenessFixpoint` is deliberately deferred |
| D-04-30 | D-04-CONTEXT `<deferred>`, D-04-04 | M002/Phase 6 | info | M002 or Phase 6 | A storable, matchable `Result` value and payload-carrying alternatives are deferred; the additive move is a sibling `alternative_details []Alternative omitempty` field, never a shape change to `Alternatives []string` |
| D-04-31 | D-04-CONTEXT `<deferred>` "Accepted residual limitations" | QLT-01, FFI-01 | warning | inherited, not scheduled | Four accepted residual limitations this phase inherits without engineering around: the coordinated three-way lie, two nonlocal-exit blind spots, the single-host single-record-shape fence, and the permanence of quarantine |
| D-04-32 | 04-CONTEXT.md `<specifics>` "Roadmap gap surfaced by this discussion", recorded 2026-09-05 by plan 04-07 (Task 3) | NAT-03, ROADMAP §Phase 5 | info | Phase 5 (planning) | NAT-03's "false no-alias facts" hostile mutation has no subject until Phase 5 itself first adds a proven alias-fact emission path, because D-04-13 means Phase 4 emits no optimizer-visible attributes to falsify. This is not a defect of Phase 4; it is a planning dependency for whoever plans Phase 5, recorded here per D-04-CONTEXT's own instruction that this gap "is not currently stated in ROADMAP.md §Phase 5 and should be added when Phase 5 is planned." |
| D-04-33 | 04-REVIEW.md WR-01 (fifth-round code review), 04-VERIFICATION.md (fifth round) | FFI-01 | warning | unscheduled — before any emitter splices `Alias` | `core.ForeignContract.Alias` is the one contract string field excluded from all three of plan 04-13's audit layers and from every falsifier table. It is safe today only because no emitter splices it into generated C; nothing pins that invariant, so a future emitter could reintroduce the closed injection class silently |

## Detail

### D-04-26 — `discoverLoanLastUses` retirement carries to Phase 5

**Disposition of D-03-01 (03-DEBT.md).** D-03-01 had two halves: (a) the
metric-honesty gap — `discoverLoanLastUses`'s own transitive-scan work was
never counted, understating a branch function's real cost — and (b) the
structural gap — `loanLivenessFixpoint`'s linear-cost dataflow never reaches
the admission-deciding code path; `discoverLoanLastUses`'s quadratic
propagation is still the sole law deciding accept/reject.

**This plan (Task 3, D-04-25) closes half (a) only.** `discoverLoanLastUses`
now returns its own recomputed-work count — one unit per binding the outer
loop visits, plus one for the final result-position check — added into both
`analyzeStraightLine`'s and `analyzeArmBody`'s `Work` totals, matching the
`blockLoanLiveness`/`result.Work++`-per-operation convention already
established. `TestLastUseDiscoveryWorkIsCounted` and
`TestLastUseDiscoveryWorkSeries` pin the counted growth; `TestCheckerVerdictsUnchanged`
pins that no accept/reject decision or diagnostic code moved across the
entire `testdata/phase1..4` corpus while this landed.

**Half (b) — retiring the quadratic derivation — is deliberately NOT
attempted this plan.** D-04-26 (04-CONTEXT.md) is explicit: *"performing that
surgery inside the same `check.go` that is simultaneously gaining
`OpForeignCall`, `OpFail`, `OpRelease`, and `OpDefect` is precisely the
compounding-wave-defect shape Phases 2 and 3 both hit."* `check.go` in this
phase already carries the coordinated foreign-call admission gate, the
reverse-order release materialization, the resource-lifecycle dispatch, and
the match-arm defect terminator — four new operation kinds and two new
control-flow shapes landing in the same file `discoverLoanLastUses` lives in.
Retiring it now would be a fifth simultaneous change to the same file's
admission-deciding core, exactly the shape that cost Phase 2 (D-02-03/D-05)
and Phase 3 (03-05's mid-phase gate) each a dedicated remediation round.

**Why not blocking.** Identical reasoning to D-03-01's own "why not blocking"
section: bounded by `MaxTokens`/parser limits, fails closed, no unbounded
path, and independently re-verified correct (not merely fast) by
`TestOwnershipSequenceExhaustive`'s ~113,164-case oracle differential and
`TestBranchSequenceExhaustive`'s per-arm equivalent. This is a residual
cost-shape gap, not a live vulnerability or a wrong verdict.

**Phase 5 fix.** Drive both admission and `LoanEndpoint` materialization from
`loanLivenessFixpoint`'s fixpoint directly, retiring `discoverLoanLastUses` as
a second, quadratic, admission-deciding law — the complete fix D-03-01/D-05
originally intended. Phase 5 is the first phase after this one whose
`check.go` work is NOT itself landing a new `OperationKind` in the same
plan, per D-08's "fold into whichever plan already touches the file" method
combined with D-04-26's compounding-wave avoidance.

### D-04-30 — payload-carrying alternatives remain deferred

**Disposition.** D-04-04 (one-way door, resolved `edge-based-now`) fixed
Phase 4's failure representation as a two-successor control-flow edge in
core, not a storable value: no `Result` type, no generics, no
payload-carrying alternatives, `core.DataType.Alternatives` (`[]string`)
untouched. This is recorded as debt at authoring time per D-04-CONTEXT's own
instruction, not discovered after the fact.

**Why not blocking.** Nothing in Phase 4's executable semantics needs a
payload-carrying alternative: the error payload rides the `err` edge as an
ordinary nullary ADT (D-04-05), and the ok payload is an ordinary place on
the ok successor block. No test, fixture, or requirement (SEM-03, RES-01,
FFI-01) needs a `Result` value to be storable, matchable, or passed across a
call boundary this phase — Phase 4 adds no Lang-to-Lang call surface at all
(D-04-01).

**Future fix.** When error *translation between layers* is actually needed
(M002, or a Phase 6 error-taxonomy requirement), add a sibling
`alternative_details []Alternative omitempty` field keyed by name — additive
to `core.DataType`, never a shape change to the existing `Alternatives
[]string` field D-13/D-04-23 already commit to leaving untouched. This is a
one-way door: a storable `Result` value bolted onto an edge-based core is a
genuine re-lowering of the failure representation, not an extension of it,
so the cost of NOT paying it now (rebuilding the whole `try`/`discard`
admission surface atop a value-based encoding) is fully accepted and named,
per D-04-04's own reversibility note.

### D-04-31 — accepted residual limitations this phase inherits

These four items are named in D-04-CONTEXT.md's `<deferred>` §"Accepted
residual limitations (record, do not engineer around)" and are re-recorded
here verbatim rather than left only in prose, per this task's own
instruction that the debt register must carry every accepted residual
limitation, not merely the two mechanically-derived items above.

1. **The coordinated three-way lie (QLT-01's documented escape class, one
   layer down).** An author who simultaneously edits the frozen foreign
   fixture, the Lang `foreign` declaration, and the conformance TU's own
   expectations passes Phase 4 on a wrong boundary. No control in this
   repository — Phase 1 through 4 — closes a coordinated multi-artifact lie;
   `originvalidate.KnownEscape` and `corevalidate.KnownEscape` already name
   the analogous coordinated frontend/summary lie as an accepted residual for
   the same reason. Not scheduled for closure; recorded as a standing,
   permanent property of independently-derived-but-jointly-editable
   artifacts.

2. **Two nonlocal-exit blind spots (D-04-17's pad).** The process-root
   `setjmp` pad cannot see a `longjmp` to a foreign-owned `jmp_buf`
   established below it (reachable only through foreign-invoked callbacks —
   a shape this phase's language cannot construct, since it has no
   calls-into-Lang), and foreign `exit()`/`_Exit()`/`raise()` is not
   observable at all (no pad seam exists for either). Both are named,
   fenced, non-provable claims about opaque foreign code, matching this
   project's existing quarantine posture (`wiki/native-and-low-level-profile.md`).

3. **The single-host single-record-shape fence (spike 005's Known
   Limitations, inherited unchanged).** One small by-value record on one
   host: unions, bit fields, vectors, variadics, packed/aligned records, and
   aggregate-passing thresholds remain open, as do ELF, x86-64, GCC, and
   non-Apple Clang. `core.ForeignContract.Layout`'s single-field
   `lang_foreign_resource_block` shape and `standardForeignLayout`'s
   "currently checker-derived, not a per-symbol declaration" comment are the
   in-tree evidence this fence is still exactly where spike 005 left it.

4. **Quarantine is permanent and non-discharging.** A green gate never
   proves the foreign implementation obeys its declared contract — the
   `lang.foreign/0` sidecar manifest's `unchecked_obligations` list
   (Capture/Retention/Aliasing, per `standardForeignObligations`'s own doc
   comment) is the honest record of exactly which obligations this phase
   asserts rather than proves. This is not a gap to close; it is the
   project's own adjudicated position on what a compiler can and cannot know
   about opaque C (`wiki/semantic-kernel-contract.md` §"Layout, validity,
   unsafe, and FFI").

**Why none of these are blocking.** All four are named, fenced, non-`omitempty`
recorded facts (the manifest's `unchecked_obligations` list, this debt
register) rather than silent gaps — the difference this project has
consistently drawn (D-04-CONTEXT, spike 005, spike 004) between an
*unproven* claim (acceptable, if named) and a *false* one (never acceptable).
No test, fixture, or control in this phase claims to close any of the four.

### D-04-32 — the Phase 5 roadmap gap this phase surfaces

**Disposition.** D-04-13 (04-CONTEXT.md) means Phase 4 emits zero
optimizer-visible attributes — there is nothing for NAT-03's "false no-alias
facts" hostile mutation (ROADMAP §Phase 5) to falsify, because Phase 4 never
emits a true alias fact for that mutation to corrupt. This was surfaced
during Phase 4's own discussion (04-CONTEXT.md `<specifics>`) as a gap not
yet reflected in ROADMAP.md §Phase 5, with an explicit instruction that it
"should be added when Phase 5 is planned."

**Why not blocking.** This is not a defect in Phase 4 — Phase 4 makes no
claim about alias-fact emission and its own required controls (control:
foreign.no_unproven_attributes) are satisfied precisely by emitting nothing
to falsify. It is a forward planning dependency: whoever plans Phase 5 must
first add a proven alias-fact emission path before NAT-03's mutation has a
subject to attack.

**Future fix.** Phase 5's own planning (`/gsd-plan-phase 5`) must read this
item and ROADMAP.md §Phase 5 together and sequence "add a proven alias-fact
emission path" before or alongside NAT-03's hostile-mutation work, not after.

---

### D-04-33 — `Alias` is outside the field audit, pinned by nothing

Plan 04-13 closed the fifth-round gap by auditing every `core.ForeignContract`
string field that reaches generated C, at three independent layers:
`check.validForeignPolicyValue` at source admission,
`corevalidate.foreignContractFieldsCSafe` / `validCIdentifier` on the core
artifact, and cgen's own `commentSafeForeignField` / `validForeignCType` /
`unsafeForeignContractField` peer guards on the exported `EmitForeign*` entry
points that never call `Validate`.

`Alias` is not in any of those predicates, and not in any falsifier table.

This is **not** a live injection channel: the fifth-round verification
confirmed by grep that `Alias` is referenced nowhere in `cgen.go` or
`corevalidate.go` — no emitter splices it, so there is no path from an
`Alias` value to emitted C text today. That is why it did not block the
phase.

What is missing is the *pin*. Every other field's safety is enforced by a
predicate plus a falsifier that goes red if the guard is weakened. `Alias`'s
safety rests only on the current absence of a splice site, and no test fails
if a future change adds one. The closed injection class could therefore
return silently through this single field.

**The cheap discharge**, for whoever next touches foreign emission: either add
`Alias` to `foreignContractFieldsCSafe` and `unsafeForeignContractField`
alongside its siblings (it costs one line each and rejects nothing accepted
today, since no honest `Alias` value exists yet), or add a test asserting that
no `EmitForeign*` output ever contains an `Alias` value — so that introducing
a splice site without an audit entry fails loudly.

Two INFO items from the same review are deliberately not tracked here: a
doc-comment/behavior mismatch in `validForeignPolicyValue`, and an
undocumented-but-safe splice of ADT alternative names in
`foreignFailureLiteral`. Neither bears on a security invariant.


## Status of carried Phase 3 debt

**D-03-01 is PARTIALLY closed by this plan.** Its metric-honesty half is
closed (see D-04-26's "Disposition" above); its structural half is
re-recorded here as D-04-26, carried to Phase 5. D-03-01's own row in
03-DEBT.md is left unedited — 03-DEBT.md is Phase 3's own historical record,
not rewritten retroactively; this file is the authoritative carry-forward
for Phase 4+.

**RES-01 and FFI-01 are dispositioned by plan 04-07.** Per D-04-CONTEXT's
`<flagged_assumptions>`, both were left `unresolved` by the edge probe;
04-07 closes them on the assumed reading recorded there — RES-01 as
cleanup-event agreement across three engines on the three-stage fixture
(falsified by the two release mutations, and now also by a genuinely
executed second-/third-stage typed-failure differential, task 04-07-02),
FFI-01 as the gate making foreign obligations inspectable and their
violations detectable, never as a proof the foreign implementation obeys
them (permanent, non-discharging quarantine, D-04-31 item 4). SEM-03 closes
on the same basis: all five Phase 4 path shapes, named explicitly, agree
across three engines. See `04-07-SUMMARY.md` for the full closure record.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Recorded: 2026-09-05 at `df75205` (plan 04-06, Task 3)*
