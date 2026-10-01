# Requirements: Schway M004

**Milestone:** Native Emission Ownership and Resource Discharge
**Defined:** 2026-09-27
**Core value:** Give an AI agent and a human reviewer the shortest reliable path
from intent to sound, reproducible evidence without hiding runtime costs.

The user authorized adopting the second specialist review automatically.
Research: [M004 synthesis](research/M004/SUMMARY.md). Living direction:
[PRODUCT-ROADMAP](PRODUCT-ROADMAP.md). All requirements below are new pending
work; completed archived Phase 21 is prework, not evidence of runtime admission.
IDs continue existing categories, including reserved historical future IDs.

## M004 Requirements

### Native application execution

- [x] **APP-02**: A developer can build a retained native executable from an admitted Schway source using a documented public command; building does not execute application effects, and the artifact runs outside the compiler's temporary build directory with its runtime dependencies declared.
- [x] **APP-03**: A developer can supply bounded input through the public application route; two caller-selected scalar inputs produce independently specified results, and malformed or oversized input is rejected with a defined outcome.
- [x] **APP-04**: One application-run request launches the selected native artifact exactly once, without prior interpretation or hidden optimization-tier replay of application effects.
- [x] **APP-05**: An application has defined stdout, stderr, and process-exit behavior; ordinary output and successful stderr are not parsed as or rejected for violating compiler execution JSON.
- [x] **APP-06**: A developer can obtain compiler execution evidence separately from application streams, with explicit disabled, incomplete, and capacity-exhausted states that cannot be reported as successful verification.

### Explicit foreign boundaries

- [x] **FFI-02**: A developer declares the local C sources/headers, symbols, and ABI inputs required by a build; path resolution survives checkout relocation, missing/incompatible inputs fail clearly, and relevant input changes invalidate the artifact/evidence identity without fixture-symbol lookup or ambient repository paths.
- [x] **FFI-03**: Each admitted foreign operation resolves its own checked signature, operand modes, acquisition/failure behavior, and release pairing; distinct acquire/use/release operations can consume admitted local values without inheriting the function's first foreign symbol contract.

### Live local resource ownership

- [x] **RES-04**: A source-constructible opaque noncopyable resource can receive a real bounded malloc-backed buffer from an explicitly linked C adapter, remain live after the adapter returns, and supply a byte determined by a caller-selected file through Schway-directed use.
- [x] **RES-05**: Generated release invokes the resource's declared infallible consuming destructor exactly once; a borrow preserves the owner's obligation, and transfer preserves the live resource under its new owner without calling the destructor.
- [x] **RES-06**: On admitted normal and typed-error exits, every successfully acquired locally owned resource that is not transferred is released in reverse successful-acquisition completion order, including an actual operation/output failure after acquisition.
- [x] **RES-07**: A failed foreign acquisition creates no Schway-owned resource; the bounded adapter contract specifies maximum size, empty input, initialization/length, failure representation, and cleanup of its own partial acquisition before exposing a result.
- [x] **RES-08**: Discarding an owning acquisition cannot erase its obligation: the admitted form either performs immediate consuming cleanup or is rejected before C serialization.
- [x] **RES-09**: Independent validation derives obligations from successful acquisitions and follows each admitted path; missing, duplicate, wrong-resource, or fabricated cleanup is rejected without relying on the presence of an existing release operation to discover the obligation.

### Ownership across Schway calls

- [x] **OWN-10**: A live resource can move through an admitted Schway call and return, remain usable by its new owner, and be released there exactly once; copying or using the moved-from owner is rejected, and an owning process-entry result is refused without an external receiver.
- [x] **OWN-11**: Acquisitions from the same static site in distinct callee activations have distinct semantic identities, preserved across transfers and independently checked without treating raw host addresses as portable identities.
- [x] **OWN-12**: Typed error propagation across Schway calls discharges remaining caller/callee obligations according to the declared cleanup order, including multiple successful acquisitions followed by a real later failure; an actual entry-to-error path exercises the cleanup.

### Bounded pointer successors

- [ ] **NAT-11**: The sole production emitter admits a bounded shared plain by-pointer read/copy helper with actual pointer-parameter C, a source-level consumer, independent shared-borrow checking, and its own positive and incompatible-access/escape negative witnesses.
- [ ] **NAT-12**: The sole production emitter admits a bounded exclusive borrowed-by-pointer read/copy helper with actual pointer-parameter C, a source-level consumer, independent exclusive-borrow checking, and its own positive and conflicting-access/escape negative witnesses.
- [ ] **NAT-13**: Bounded pointer lowering and its manifests agree on emitted attributes and add no `restrict`, `noalias`, capture, or alignment promise unsupported by checked facts; unsupported mutation, forwarding, retention, callbacks, nonlocal exits, and wider pointer shapes remain structurally refused before C serialization.

### Decisive native evidence

- [x] **EVD-09**: An observer independent of compiler release events establishes actual allocation, post-acquisition use, and destruction before process exit; reached controls for omitted, premature, duplicate, and wrong-resource destruction fail even when reported events remain plausible.
- [ ] **EVD-10**: Every admitted foreign/shared/exclusive family has its own reproducible macOS and Linux native receipt identifying source/build inputs, compiler, target, flags, expected result, and applicable optimizer/sanitizer lanes; missing host or lane evidence keeps that claim incomplete.
- [x] **EVD-11**: Explicit differential verification uses isolated or replayable inputs and declared foreign outcomes, compares against independent expected answers, and never claims that modeled foreign success proves actual host IO or that process reclamation proves cleanup.

### Developer usability

- [ ] **DX-14**: From a clean checkout, a developer can follow documented commands to build and run the bounded file-byte utility against two supplied files, observe the expected differing result and an admitted failure, and locate its explicit C bindings and resource-lifetime evidence.
- [ ] **DX-15**: Unsupported ownership/resource/pointer uses produce stable structured diagnostics with source attribution and an actionable boundary explanation; examples cover moved-from use, discarded ownership, pointer escape, and unsupported cleanup/exit behavior.

## Scope and acceptance rules

- The C adapter owns its internal file descriptor and returns the allocation
  live. The claim is Schway-owned buffer cleanup, not Schway-owned file close. Do
  not hide acquire/use/release of the claimed allocation inside one C call.
- Publish a successor to Phase 21's contract-only artifact: **release consumes,
  transfer preserves, borrow leaves ownership unchanged**. Preserve the archive;
  its old return wording is explicitly corrected, not silently treated as
  executable truth. Defects/process termination remain outside cleanup guarantees.
- Acquisition, transfer and error handling must remain coherent across the
  affected `check`, `corevalidate`, `originvalidate`, `pathoracle`, `interp`, and
  `cgen` consumers. New peers must not merely import the producer's conclusion.
- Check in the selected source/input/output frontier before implementation.
  Each phase finishes with an ordinary public route and expected answer.
  A moved parser refusal or cross-engine agreement alone is insufficient.
- EVD-09 controls are introduced as their corresponding behavior is admitted;
  physical proof is not deferred until final integration. EVD-10 is the final
  whole-milestone coverage owner, with family evidence required at admission.
  EVD-11 is the initial replay boundary, extended with each new foreign outcome.
- Foreign C implementations and their declarations remain trusted/audited build
  inputs. This milestone cannot prove arbitrary C obeys its contracts. File/input
  bounds are explicit. No implicit network fetch, package resolution, or broad
  sandboxing framework is added to implement local linkage.
- Optimizer attributes are optional commitments beyond pointer passing.
  D-16-07 is prospectively amended for conservative pointer lowering; the
  attribute-bearing variant remains gated on separate proof.
- Existing archived verification retains its revision scope. Do not rerun
  completed Phase 21 UAT or relabel it runtime proof. D-12-43's dated closure
  is a planning repair, not a new M004 implementation requirement.

## NAT-09 Successor Ownership

The archived M003 cut and Phase 21 design ownership remain intact. New runtime
successors map as follows; ROADMAP assigns exact phase owners.

| Historical family | Current requirements | Admitted subset | Required evidence / residual refusal |
|---|---|---|---|
| D-16-11 ordinary foreign | FFI-03, RES-04–09, OWN-10–12 | Live bounded allocation, contracted use, infallible release, admitted transfer and typed errors | EVD-09/10/11; no unwind, callbacks, nonlocal transfer, generic partial initialization or fallible implicit destructor |
| D-16-12 exclusive borrowed by-pointer | NAT-12, NAT-13 | Exclusive read/copy helper with ordinary C pointer | Separate native pointer witness on both hosts; no mutation, forwarding, escape/retention or unsupported alias promise |
| D-16-13 shared plain by-pointer | NAT-11, NAT-13 | Shared read/copy helper with ordinary C pointer | Separate native pointer witness on both hosts; no incompatible access, forwarding, escape/retention or unsupported alias promise |

## Future Requirements

These are capability candidates, not active M004 acceptance or assigned future
milestone numbers. Promote at the next milestone boundary with fresh evidence.

- Defined U64 arithmetic/remainder, comparison/Bool, continuation branching,
  scalar iteration and minimal text/decimal output: `sum_to_n` and FizzBuzz.
- Scalar-loop spike: CFG fixed points, back-edge loan/ownership boundaries,
  dynamic event occurrences, and bounded oracle exhaustion distinct from
  application semantics. Initially refuse resource/loan carries if necessary.
- Byte views, bounds/indexing, arity-N, small aggregates and local modules when
  reusable file/JSON library consumers require them; explicit malformed-input,
  depth/size and allocation/error policies.
- A bounded HTTP consumer independent of JSON, with protocol framing, timeouts,
  resource/close behavior, and a selected TLS posture where relevant.
- Separate compilation and user-declared contracts, including DX-06 / D-13-02b;
  richer generics/effects/concurrency and ecosystem after concrete demand.

## Out of Scope

| Capability | Reason |
|---|---|
| Arithmetic, loops, recursion in M004 | Keep native ownership bounded; practical computation is the following milestone |
| New VM, LLVM backend, tracing heap or service runtime | Existing native toolchain suffices for the selected programs |
| Full String runtime, Unicode API, generic collections, owning aggregates | Not needed by the bounded buffer and scalar entry witness |
| Fallible implicit destructors, direct owning file/socket close | Require consumed/retained/unknown state and error-precedence contracts beyond infallible free |
| Unwinding, longjmp/nonlocal exits, cancellation, callbacks, retained pointers | Change cleanup/lifetime assumptions; continue explicit refusal |
| Broad pointer mutation/forwarding and alias optimizer attributes | Exact read/copy shapes are the committed successors; broader proof is separate |
| JSON, HTTP, remote packages, LSP, effects/concurrency | Future capability horizons; avoid hiding ecosystem work inside this milestone |
| Broad historical assurance cleanup / CI redesign | Carry named debt; perform only changes required by the selected capability |

## Traceability

Each of the 24 requirements has exactly one completion owner in Phases 22–25.
Cross-phase admission/regression duties do not create duplicate ownership.
RES-05/06 and EVD-09 finish in Phase 24; local cleanup and physical controls
are mandatory at Phase 23 admission. EVD-10 finishes the coverage matrix in
Phase 25; each family's native host evidence is required when it lands.
EVD-11 starts in Phase 22 and extends with every admitted foreign outcome.
No new requirement maps to completed historical Phase 21. All remain Pending.

| Requirement | Phase | Status |
|---|---|---|
| APP-02 | Phase 22 | Complete |
| APP-03 | Phase 22 | Complete |
| APP-04 | Phase 22 | Complete |
| APP-05 | Phase 22 | Complete |
| APP-06 | Phase 22 | Complete |
| FFI-02 | Phase 22 | Complete |
| FFI-03 | Phase 23 | Complete |
| RES-04 | Phase 23 | Complete |
| RES-05 | Phase 24 | Complete |
| RES-06 | Phase 24 | Complete |
| RES-07 | Phase 23 | Complete |
| RES-08 | Phase 23 | Complete |
| RES-09 | Phase 23 | Complete |
| OWN-10 | Phase 24 | Complete |
| OWN-11 | Phase 24 | Complete |
| OWN-12 | Phase 24 | Complete |
| NAT-11 | Phase 25 | Gaps Found |
| NAT-12 | Phase 25 | Pending |
| NAT-13 | Phase 25 | Pending |
| EVD-09 | Phase 24 | Complete |
| EVD-10 | Phase 25 | Pending |
| EVD-11 | Phase 22 | Complete |
| DX-14 | Phase 25 | Pending |
| DX-15 | Phase 25 | Pending |

Coverage: 24/24 requirements mapped; zero orphans, zero duplicate owners.

---
*Last updated: 2026-09-27 during M004 roadmap creation.*
