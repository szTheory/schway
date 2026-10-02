# Roadmap: Schway

## Milestones

- ✅ **M001 — Source-to-Native Semantic Spine** — Phases 1–6 (shipped 2026-09-07) — [archive](milestones/M001-ROADMAP.md)
- ✅ **M002 — Interprocedural Semantic Spine** — Phases 07–13 (shipped 2026-09-14) — [archive](milestones/M002-ROADMAP.md)
- ✅ **M003 — Computation and Honest Instruments** — Phases 14–20 (shipped 2026-09-26; audit: tech debt) — [archive](milestones/M003-ROADMAP.md)
- ✅ **M004 — Native Emission Ownership and Resource Discharge** — Phases 22–25 implementation and hosted evidence are complete; the final report-only handoff and next milestone are recorded below

## M004 Goal

A developer can build a retained native executable, run it once on caller input,
and observe a bounded file byte through an allocation that remains owned by Schway
across use, calls, errors, and exactly-once destruction or ownership transfer.
Shared and exclusive read-copy pointers have separate bounded admissions.

Scope: [REQUIREMENTS](REQUIREMENTS.md). Adopted decisions and inspected boundaries:
[research synthesis](research/M004/SUMMARY.md). Future order:
[PRODUCT-ROADMAP](PRODUCT-ROADMAP.md); observed capability:
[LANGUAGE-MATURITY](LANGUAGE-MATURITY.md). Phases 22–25 are complete, and the
M004 requirements are closed. Phase 22's objective README contract UAT is
complete (1/1); subjective readability is not claimed. When a completed phase
has a stale covered-input report and zero incomplete plans, `$gsd-execute-phase
N` resumes at verification gates without replaying plans. Preserve completed
UAT, re-query after the refresh, and use `$gsd-new-milestone` when the live
resolver permits that transition.

## Completed Historical Prework

**Phase 21 — Native Emission Ownership and Resource Discharge:** complete and
archived under [M004 phase artifacts](milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/).
Its six plans, seven UAT cases, and six passing verification truths describe its
recorded revision. No new requirement belongs to Phase 21; do not replay it.
Its design ownership and emitter retirement remain in force. It did not admit
the three refused native families or prove physical resource destruction.

Dated kickoff amendments affect archived Phase 16/21 report fingerprints;
those reports remain historical evidence and completed UAT remains preserved.
The successor contract explicitly corrects the old return wording:
**release consumes an obligation; transfer preserves it under its new owner;
borrow leaves ownership unchanged.** M003 remains the shipped predecessor.

## Phases

Sequential IDs continue after archived Phase 21. Every new phase delivers a
runnable source/input/output witness, with executable plans and summaries kept
in its phase directory.

- [x] **Phase 22: Native Application Build and Single Execution** — A retained scalar application accepts caller input and runs once with ordinary streams. (3/3 implementation plans complete; objective README contract passed; subjective readability not claimed) (completed 2026-09-27)
- [x] **Phase 23: Live Local Allocation and Discharge** — A bounded file-byte application uses and releases a real Schway-owned allocation. (completed 2026-09-30)
- [x] **Phase 24: Ownership Transfer Through Calls and Errors** — Live resources survive ownership transfer and discharge across frames and typed errors. (completed 2026-10-01)
- [x] **Phase 25: Separate Pointer Successors and Integrated Utility** — Shared/exclusive pointer helpers have distinct native proof and the complete utility is reproducible. (completed 2026-10-02)

## Phase Details

### Phase 22: Native Application Build and Single Execution

**Goal**: A developer can retain and execute a native application once on bounded real input, with application streams separate from compiler evidence.
**Depends on**: Completed Phase 21 prework; no new implementation phase
**Requirements**: APP-02, APP-03, APP-04, APP-05, APP-06, FFI-02, EVD-11
**Success Criteria** (what must be TRUE):

1. From a relocated checkout, a developer can build a retained executable with declared local C sources, headers, symbols, ABI inputs, and runtime dependencies. Building performs no application effects; missing/incompatible inputs fail clearly and relevant input changes invalidate artifact/evidence identity.
2. An ordinary scalar source program run with caller-selected inputs `7` and `42` produces independently specified corresponding results; malformed and oversized input has a defined outcome.
3. One public application-run request launches the selected artifact exactly once, with defined stdout, stderr, exit, and signal behavior; ordinary output and successful stderr are accepted without execution-JSON parsing or hidden interpretation/tier replay.
4. A developer can request execution evidence through a separate channel and distinguish disabled, incomplete, and capacity-exhausted states from successful verification.
5. Explicit differential verification uses isolated or replayable inputs and independent expected answers; declared foreign outcomes are distinguished from actual host IO, and verification replay does not occur on the ordinary application route.

**Plans**: 3/3 plans complete
**Wave 1**

- [x] 22-01-PLAN.md — Retained identity application with bounded input and one launch

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 22-02-PLAN.md — Relocatable local C linkage and content-bound build identity

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 22-03-PLAN.md — Separate execution evidence and explicit differential replay

**Runnable witness**: Check in a scalar identity source with input/output pairs
`7 → 7` and `42 → 42`, then build/run its retained artifact through documented
public commands. Fix output encoding and malformed/oversize behavior before
implementation. Observe zero application launches during build and one during run.

**Planning boundary**: Settle input/evidence transport, the resource consumer's
input needs, subprocess outcomes, and local build identity. Declaring/linking C
inputs does not admit foreign resource calls. EVD-11's replay contract applies
to later foreign outcomes; extend its fixtures without reassigning ownership.

### Phase 23: Live Local Allocation and Discharge

**Goal**: A developer can read a caller-selected file byte through a real allocation returned live to Schway and observe its generated local cleanup.
**Depends on**: Phase 22
**Requirements**: FFI-03, RES-04, RES-07, RES-08, RES-09
**Success Criteria** (what must be TRUE):

1. A source-constructible opaque noncopyable resource receives a bounded malloc-backed allocation from the explicitly linked adapter, stays live after acquisition returns, supplies the independently expected byte through Schway-directed use, and is physically destroyed before application exit.
2. Distinct acquire, use, and infallible consuming release calls operate on admitted local values using their own checked signature, operand modes, acquisition/failure facts, and release pairing; none inherits an unrelated first-symbol contract.
3. Empty, maximum-size, oversized, malformed, and failed inputs obey the published buffer/initialization/error contract; failed acquisition creates no Schway owner, and the adapter cleans its own partial acquisition.
4. Discarded owning success is rejected or immediately destroyed. Independent acquisition-derived validation rejects missing, duplicate, wrong-resource, or fabricated cleanup even when every release operation is removed from candidate core.
5. Native runs exercise normal completion and a real later operation/output failure after successful acquisition. Generated local cleanup consumes each remaining obligation exactly once; independent physical observation rejects reached omitted or premature destruction despite plausible compiler events.

**Plans**: 7/7 plans executed

Plans:
**Wave 1**

- [x] 23-01-PLAN.md — Public live-owner tracer and three operation ABIs

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 23-02-PLAN.md — Acquisition-seeded independent peer validation and reached mutations
- [x] 23-03-PLAN.md — Source and emitter ownership refusal, including discarded acquisition
- [x] 23-04-PLAN.md — Bounded acquisition and partial-allocation cleanup
- [x] 23-05-PLAN.md — Model-only outcomes and successor discharge contract

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 23-06-PLAN.md — Independent native lifetime observer and reached controls

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 23-07-PLAN.md — Documented witness and macOS/Linux focused CI gate

**Execution waves**: Wave 1: 23-01 public tracer. Wave 2: 23-02 peer proof, 23-03 source/emitter refusal, 23-04 acquisition boundaries, and 23-05 model/contract in parallel. Wave 3: 23-06 physical observer after peer, refusal, and acquisition proof. Wave 4: 23-07 documented host evidence after model and observer. The extra plan isolates RES-08/RES-09 and D-23-05 safety controls without narrowing the public tracer.

**Runnable witness**: A file-byte source consumes two supplied one-byte files
containing `0x41` and `0x42` and reports the corresponding expected byte. Add a
failed-acquisition input and a reached post-acquisition failure. Select source
form, output encoding, maximum size, empty representation, and failure contract
before implementation; reuse Phase 22's public route.

**Admission obligations**: Publish Phase 21's corrected successor contract.
Local release/error cleanup and physical controls are mandatory here although
RES-05/06 and EVD-09 finish in Phase 24. Require foreign-family native macOS/Linux
receipts and replayable modeled outcomes now (EVD-10/11). The adapter owns its
file descriptor; Schway owns the returned allocation. Contracted adapter use does
not reopen either pointer-helper family. Any necessary narrow owning return
form covers all affected consumers; owning aggregates remain refused.

### Phase 24: Ownership Transfer Through Calls and Errors

**Goal**: A developer can transfer a live resource through Schway calls and returns, use it under its new owner, and rely on exactly-once cleanup on admitted normal and typed-error paths.
**Depends on**: Phase 23
**Requirements**: RES-05, RES-06, OWN-10, OWN-11, OWN-12, EVD-09
**Success Criteria** (what must be TRUE):

1. An allocation acquired in one frame remains usable after ownership transfer through a Schway call and return; transfer invokes no destructor, borrow preserves ownership, and the final owner invokes the declared consuming destructor exactly once.
2. Two activations of the same static acquisition site produce distinct semantic resource identities that survive transfer and are independently checked without using raw host addresses as portable IDs.
3. Actual entry-to-success and entry-to-typed-error executions release every non-transferred caller/callee resource in reverse successful-acquisition completion order, including multiple acquisitions followed by a real later operation/output failure; failed acquisitions contribute no obligation.
4. Copying ownership, using a moved-from owner, or returning an owning process-entry result without an external receiver is rejected before execution with source-attributed diagnostics.
5. An observer independent of compiler release events proves allocation, use after acquisition/return, and destruction before exit. Reached omitted, premature, duplicate, and wrong-resource destruction controls all fail even with plausible reported events.

**Plans**: 3/3 plans complete
**Wave 1**

- [x] 24-01-PLAN.md — Transfer a live owner through a helper and return

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 24-02-PLAN.md — Preserve repeated activation identity and typed-error cleanup

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 24-03-PLAN.md — Prove physical cleanup and bind dual-host evidence

**Runnable witness**: Extend the file-byte utility so a helper acquires the
allocation, transfers it through a call/return, and the receiving owner reports
`0x41` or `0x42` before cleanup. A separate ordinary entry exercises repeated
helper activations and a later typed error after multiple successful acquisitions;
fix its expected result/error and destruction order before admission.

**Admission obligations**: Complete the bounded ordinary foreign successor
(D-16-11), extending native host receipts and EVD-11's declared outcome model.
Preserve independent derivation across affected `check`, `corevalidate`,
`originvalidate`, `pathoracle`, `interp`, and `cgen` consumers. Wider owning
payloads retain their refusal/debt prerequisites unless this witness requires
a narrow form. Defect/process termination stays outside cleanup guarantees.

### Phase 25: Separate Pointer Successors and Integrated Utility

**Goal**: A developer can use separately checked shared and exclusive read-copy pointer helpers in the documented native utility, with reproducible evidence for each admitted family.
**Depends on**: Phase 24
**Requirements**: NAT-11, NAT-12, NAT-13, EVD-10, DX-14, DX-15
**Success Criteria** (what must be TRUE):

1. A source consumer invokes a bounded shared plain by-pointer read/copy helper through actual pointer-parameter C from the sole production emitter; its positive native witness succeeds and independent checking rejects incompatible access and escape.
2. A distinct source consumer invokes a bounded exclusive borrowed-by-pointer read/copy helper through actual pointer-parameter C; its own positive witness succeeds and independent checking rejects conflicting access and escape.
3. Emitted C and attribute manifests agree and add no `restrict`, `noalias`, capture, or alignment promise. Unsupported mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer shapes are structurally refused before C serialization.
4. Each foreign/shared/exclusive family has reproducible macOS and Linux native receipts naming source/build inputs, compiler, target, flags, expected answers, and applicable optimizer/sanitizer lanes. Missing host/lane evidence leaves the corresponding claim incomplete.
5. From a clean checkout, a developer follows documented public commands to build/run the utility on the two supplied files and a specified failure, gets independent expected results, and locates explicit C bindings and lifetime evidence. Moved-from use, discarded ownership, pointer escape, and unsupported cleanup/exit examples produce stable structured diagnostics with source attribution and actionable boundary explanations.

**Plans**: 7/7 plans complete and summarized

Plans:
**Wave 1**

- [x] 25-01-PLAN.md — Admit the shared read/copy pointer ABI

**Wave 2** *(blocked on Wave 1 completion)*

- [x] 25-02-PLAN.md — Independently validate borrow families and live paths

**Wave 3** *(blocked on Wave 2 completion)*

- [x] 25-03-PLAN.md — Admit exclusive helper and integrated caller composition

**Wave 4** *(blocked on Wave 3 completion)*

- [x] 25-04-PLAN.md — Complete exclusive lowering, model behavior, and diagnostics

**Wave 5** *(blocked on Wave 4 completion)*

- [x] 25-05-PLAN.md — Complete production lowering and independent owner-result peer proofs

**Wave 6** *(blocked on Wave 5 completion)*

- [x] 25-06-PLAN.md — Publish utility instructions, dual-host evidence gate, and living-document closeout (hosted matrix passed: run 36971855722)

**Wave 7** *(blocked on Wave 6 completion)*

- [x] 25-07-PLAN.md — Reconcile Phase 25 validation, security, product-roadmap, and maturity records with the successful hosted dual-host receipt

**Phase status at Plan 25-07 kickoff (historical):** implementation plans were
complete. Hosted run 36971855722 had passed the macOS/arm64 and Linux/x86_64
foreign/shared/exclusive matrix across baseline, optimized, and ASan+UBSan
lanes at one PR merge revision. At that point, the four tracked evidence and
living-language records still described EVD-10 as pending; Plan 25-07 was
assigned to reconcile them before goal-backward re-verification.

**Phase 25 closeout amendment — 2026-10-02:** Plan 25-07 reconciled the current
validation, security, product-roadmap, and language-maturity records to hosted
run 36971855722. The goal-backward verifier passes 6/6 must-have truths and
5/5 roadmap criteria; the seven plans and all 24 requirements are complete.
The independent cold-read and automated README evidence-link checks are
recorded in `25-VERIFICATION.md`. PR #1 has merged, and final-head hosted CI
passed as recorded in STATE.md. The hosted matrix above remains tied to its
recorded revision.

**Runnable witness**: Separate shared and exclusive helper sources each read/copy
the known input byte with an exact expected answer. The final utility composes
the admitted resource/call path with these bounded accesses and reproduces the
two-file and failure behavior through the clean-checkout commands.

**Admission obligations**: Apply the prospective D-16-07 amendment for ordinary
pointers with no added alias attributes; attribute-bearing variants require
separate future proof. A by-value fallback cannot satisfy either pointer family.
Introduce each family's independent checks, negative controls, actual host
receipts, and applicable sanitizer lanes when it lands. EVD-10 closes the full
coverage matrix; it does not postpone Phase 23/24 evidence.

## NAT-09 Successor Ownership

Phase 21 retains archived design ownership. These are bounded runtime successors
to the M003 cut, not claims that every historical shape is restored.

| Historical family | New requirements / delivery owner | Exact admitted subset | Evidence and residual refusal |
|---|---|---|---|
| D-16-11 ordinary foreign | FFI-03, RES-04/07/08/09: Phase 23; RES-05/06, OWN-10–12: Phase 24 | Live bounded allocation, contracted use, infallible consuming release, call/return transfer, typed errors | EVD-09 physical controls complete in 24; EVD-10/11 extend on each admission. Refuse unwind, callbacks, nonlocal transfer, generic partial initialization, fallible implicit destructors, and unsupported owning payloads. |
| D-16-12 exclusive borrowed by-pointer | NAT-12, NAT-13: Phase 25 | Exclusive read/copy helper with ordinary C pointer | Separate expected-answer, conflict/escape controls and macOS/Linux receipts. Refuse mutation, forwarding, retention, wider shapes, and unsupported attributes. |
| D-16-13 shared plain by-pointer | NAT-11, NAT-13: Phase 25 | Shared read/copy helper with ordinary C pointer | Separate expected-answer, incompatible-access/escape controls and macOS/Linux receipts. Refuse forwarding, retention, wider shapes, and unsupported attributes. |

## Delivery Rules and Future Direction

Before implementation, check in each source/input/output frontier and refusal
diagnostic. Acceptance requires the intended result, independent validation,
and reached negative controls; engine agreement alone is insufficient. Introduce
guarantees alongside admitted behavior and keep modeled foreign outcomes distinct
from actual host IO and physical destruction. C inputs remain trusted/audited.

Every feature phase ends with a public runnable gain. A second consecutive
enabling phase triggers scope review; extra plans require a named witness or
safety obligation. Assign expensive lanes one owner and rerun when their inputs
or claims change. Review both living documents at planning and completion.

M004 excludes arithmetic, loops, recursion, owning aggregates, fallible implicit
destructors, unwind/nonlocal exits, retained pointers, and broad mutation.
The following milestone targets `sum_to_n` and FizzBuzz with scalar arithmetic,
comparison, iteration, and minimal output; its loop spike resolves fixed points,
dynamic occurrences, and bounded evidence before loop planning. Byte libraries,
JSON, and the independent HTTP branch follow concrete consumer needs.

## Progress

**Execution order:** 22 → 23 → 24 → 25. Phase 21 is complete historical prework
outside this new delivery count. All 24 requirements have one owner; none is met
by roadmap creation. Future phase plan counts remain TBD until those phases are
planned.

| Phase | Plans Complete | Status | Completed |
|---|---|---|---|
| 22. Native Application Build and Single Execution | 3/3 | Complete    | 2026-09-27 |
| 23. Live Local Allocation and Discharge | 7/7 | Complete    | 2026-09-30 |
| 24. Ownership Transfer Through Calls and Errors | 3/3 | Complete    | 2026-10-01 |
| 25. Separate Pointer Successors and Integrated Utility | 7/7 | Complete    | 2026-10-02 |

## Next Action

Phase 25 was the final M004 implementation phase. PR #1 is merged; the final
PR head and squash commit are recorded in STATE.md. Phase 22 UAT is complete
(1/1), and Phases 23–25 require no human UAT. When a completed phase report is
stale, `$gsd-execute-phase <N>` resumes only at its verification gates; it does
not replay summarized plans or completed UAT. After the Phase 23 and Phase 25
reports pass, the exact next command is `$gsd-new-milestone` to shape the
practical-computation milestone (`sum_to_n`, then FizzBuzz). Preserve completed
UAT and do not reopen Phase 25 implementation.
