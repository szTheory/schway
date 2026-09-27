# M004 Project Research Summary

**Project:** Codename Lang
**Milestone:** Native Emission Ownership and Resource Discharge
**Domain:** Native language, ownership checking, and evidence tooling
**Researched:** 2026-09-27
**Confidence:** HIGH in direction and observed blockers; MEDIUM in design and delivery size

## Executive Summary

Lang already has a Go-hosted compiler, independent interpreter, and readable C17 native
lowering. Its public native runner still supplies fixture inputs, executes optimization
tiers, and treats stdout as execution JSON. All three foreign/by-pointer families remain
refused. M004 should deliver a retained executable that runs once on caller input, then
a bounded file-byte utility whose malloc-backed allocation remains owned by Lang through
use, transfer, and generated cleanup. Keep Go 1.24, its standard library, and installed
Clang; the observed gaps do not justify another backend, dependency stack, or runtime.

Adopt four executable phases starting at 22: application execution, live local allocation,
ownership through calls/errors, and separate shared/exclusive read-copy pointer admissions.
Deliver the scalar application first and the first real resource application second.
Preserve completed Phase 21 as archived contract/retirement prework. Its successor must
distinguish release, which consumes an obligation, from transfer, which preserves it under
a new owner. Resolve D-16-11/12/13 through exact shapes in `emitProgram`; record a prospective
D-16-07 amendment allowing ordinary pointers without additional alias promises. Arithmetic
and loops stay outside M004; the following milestone completes iterative `sum_to_n` and
FizzBuzz with minimal text output, followed by byte libraries/JSON and a separate HTTP branch.

The main risk is reporting cleanup while storage was never owned, was freed early, or was
never freed. Derive obligations from successful acquisition; bind contracts per operation;
preserve dynamic identity through frames; independently observe physical allocation/use/free.
Pair expected application outputs with preserved-event destructor mutations and actual
macOS/Linux evidence for each pointer family. Refuse fallible implicit destructors,
unwinding, retention, mutation, and broader pointer shapes. Local C inputs are explicit build
authority and trusted foreign code; checked declarations do not prove their implementation.

## Key Findings

The four reports agree on a bounded application and ownership slice through the existing compiler.

### Recommended Stack

[STACK.md](STACK.md) identifies extensions to existing components, with no new dependency:

| Component | Recommendation and rationale |
|---|---|
| Go 1.24, standard library | Retain Stage 0 and zero third-party production dependencies. Existing parsing, checking, orchestration, and evidence mechanisms suffice. |
| C17, installed Clang | Extend the sole `emitProgram` authority; record actual compiler, target, flags, and source identities in receipts. |
| Audited local C adapter | Compile explicitly declared sources/headers/symbols/ABIs. Resolve paths reproducibly and invalidate builds when relevant inputs change. |
| Interpreter model | Use scripted/replayable foreign outcomes in explicit verification. It is not an implementation of host I/O. |
| Native observer, ASan/UBSan | Observe real lifetime before process exit; use sanitizers for their separate host-scoped questions. Event balance alone cannot prove destruction. |
| Existing planning documents | Use AGENTS, PRODUCT-ROADMAP, and LANGUAGE-MATURITY at workflow transitions; no background planner or new framework. |

Emit no added `restrict`, `noalias`, capture, or alignment assertion without separately
checked facts. Attribute manifests describe actual lowering. Keep target layout and portable
resource identities independent of current Apple arm64 host facts.

### Expected Features

[FEATURES.md](FEATURES.md) defines two public witnesses: caller-driven scalar execution,
then actual file content observed through a live Lang-owned buffer.

**Must have:** retained executable; bounded external input; ordinary stdout/stderr/exit;
separate optional evidence; build without application I/O and run exactly once; explicit C
build inputs without fixture-name dispatch; live noncopyable storage; acquisition-derived
obligations and per-operation contracts; transfer through calls; normal/typed-error cleanup;
separate shared/exclusive read-copy admissions; structured refusals and exact host evidence.

The adapter owns and closes its internal file descriptor. Lang owns the returned allocation;
this does not establish Lang-owned file handles. Maximum size, empty-buffer representation,
length/initialization, malformed input, and failed/partial acquisition need explicit contracts.

**Differentiators:** small canonical source with reviewable ownership; a held-out ownership
repair through supported diagnostics; bounded fresh-agent authoring evidence; and scoped
cold/warm measurements of the complete author/check/build/run/observe/repair loop.

**Defer:** arithmetic/comparison, continuation control flow, scalar loops, fixed text and U64
formatting to the next milestone's iterative `sum_to_n` and FizzBuzz. Add byte views/indexing,
arity/aggregates, local modules, and bounded JSON for real library consumers. HTTP has its own
byte/resource/timeout requirements. General strings, broad FFI, owning aggregates, fallible
implicit cleanup, async, separate compilation, package infrastructure, and a VM stay outside M004.

### Architecture Approach

[ARCHITECTURE.md](ARCHITECTURE.md) recommends shared checked function lowering with separate
application and verification entry shells. User effects execute once; explicit verification
may compare model/native tiers only in controlled disposable or replayable worlds.

| Component | Responsibility |
|---|---|
| Syntax/checker/abilities | Make selected acquire/use/release/transfer forms constructible; generalize foreign operands only as required by the witness. |
| Typed core | Bind each operation to its own signature, modes, release pair, and failure facts; represent noncopyable ownership and dynamic identity. |
| Independent validators | Re-derive acquisition/alias/path/discharge obligations in affected `corevalidate`, `originvalidate`, and `pathoracle` consumers. |
| Interpreter/execution peers | Model acquisition, transfer, errors, and release; distinguish activations and incomplete evidence. |
| `cgen.emitProgram` | Emit actual calls/cleanup; reject unsupported shapes before serialization. Both entry routes share body semantics. |
| CLI/session/native | Retain artifacts, accept bounded input/declared linkage, run once, separate streams/evidence, and record provenance. |
| Adapter/native observer | Return live storage, clean internal partial failures, provide infallible consuming release, and independently observe real destruction. |

Successful acquisition creates an obligation; failure transfers none. Borrow preserves
ownership; move/return transfer changes its owner without invoking a destructor. Release
consumes it. Ordinary return/error releases non-transferred local resources in reverse
successful-acquisition completion order. Defect/process termination retains the limited guarantee.
A missing terminal record cannot count as cleanup. Generic fallible implicit destructors remain refused.

Seed obligations from acquisition, never existing releases. Refuse discarded owning success
unless immediate consuming cleanup is implemented. Identity combines acquisition site and
activation and survives transfer; raw addresses are not portable semantic IDs. Refuse owning
process-entry results without an external receiver. Preserve resource-payload refusals and
D-10-C01/C02/C04 prerequisites unless a necessary narrow return form is explicitly admitted.

### Critical Pitfalls

1. **Harness success mistaken for application capability:** require retained artifacts, two caller inputs with independent expected results, observed single execution, and separate output/evidence channels.
2. **Events mistaken for physical lifetime:** observe content and actual destruction before process exit; omitted/premature/duplicate/wrong-resource release controls must fail even when compiler events remain plausible.
3. **Transfer mistaken for release:** publish a corrected successor contract; require live post-return use and path-complete obligations. Actual entry-to-error execution proves more than directly invoking a cleanup block.
4. **One pointer proof closes every family:** require separate shared/exclusive witnesses, actual pointer-parameter C, exact macOS/Linux receipts, and structural unsupported-neighbor refusals. Omitting `restrict` does not waive ownership.
5. **Assurance consumes the milestone or equality hides wrong results:** tie each phase to a runnable gain and independent expected output; prove mutations reached their target; add plans only for a named witness or safety obligation.

[PITFALLS.md](PITFALLS.md) also identifies stale prose, historical guards coupled to mutable
charters, obsolete debt dispositions, and duplicated expensive CI work. Preserve dated
records; these findings do not create a new M004 maintenance campaign.

## Implications for Roadmap

Recommend **four new phases, 22–25**. Phase 21 remains complete prework. Exact atomic
requirement IDs and ownership mapping belong in the next workflow step.

### Phase 22: Native Application Build and Single Execution

**Rationale:** establish the real-effects boundary before resource admission.
**Delivers:** retained scalar executable; two caller values with expected ordinary output;
bounded input/exit contracts; separate optional evidence; explicit local C build authority
and provenance. Establish the resource consumer's input route and checkout-independent execution.
**Addresses:** retained build, caller input, ordinary streams, explicit linkage, one execution.
**Avoids:** repeated effects, canned values/symbols, evidence parsed from application bytes,
and compile-time application I/O. Declare disabled/exhausted/incomplete evidence behavior.

### Phase 23: Live Local Allocation and Discharge

**Rationale:** deliver the resource witness immediately after the application boundary.
**Delivers:** a bounded file-byte application with a live returned allocation, Lang-directed
use, and generated release; failed acquisition and real later-operation/output failure.
Publish the successor contract and implement per-operation binding/acquisition obligations
across affected consumers. Settle size, empty-buffer, error, and source-form contracts first.
**Addresses:** owned file buffer, explicit adapter linkage, local error cleanup, first D-16-11
witness. Contracted adapter borrowed use does not itself reopen either pointer-helper family.
**Avoids:** allocation/free hidden inside one call, discard leaks, first-symbol binding,
synthetic failure-only evidence, and event-only cleanup. A preserved-event omitted destructor must fail.

### Phase 24: Ownership Transfer Through Calls and Errors

**Rationale:** extend already observable physical ownership across Lang frames.
**Delivers:** acquire in one frame, transfer through a call/return, use and release in the new
owner; repeated activations; multiple acquisitions/reverse cleanup; typed error propagation.
Complete bounded D-16-11 admission. Any narrow owning-return change explicitly covers all
affected operation consumers and evidence peers.
**Addresses:** moved-from refusal, transfer, dynamic identity, ordinary error cleanup.
**Avoids:** freeing on return transfer, owning copies, collapsed activations, and wrong-resource
release masked by matching events. Require real post-return use and final physical destruction.

### Phase 25: Separate Pointer Successors and Integrated Utility

**Rationale:** pointer admission builds on demonstrated lifetime and transfer.
**Delivers:** separate shared/exclusive read-copy helper witnesses using actual pointer C;
prospective D-16-07 amendment; individual D-16-12/13 dispositions; truthful attribute metadata;
family-specific native macOS/Linux evidence and refusal controls. Close with the documented
utility, held-out authoring/repair exercise, and scoped cold/warm observations.
**Addresses:** borrowed access, exact host evidence, predictable refusals, composed acceptance.
**Avoids:** by-value fallback, shared/exclusive proof substitution, stale `restrict` claims,
and broadening into mutation/forwarding/retention/unwind. Missing mandatory host evidence leaves its family incomplete.

### Phase Ordering and Research Flags

Application execution precedes real resources; local ownership precedes transfer; pointer
families follow established lifetime facts. D-16-11 spans 23–24; D-16-12 and D-16-13 have
distinct Phase 25 gates. Preserve NAT-09 lineage and retired emitters. Computation first would
reach algorithms sooner but defer the adopted resource commitment; honor resources now and
deliver computation next. All three families mean exact bounded admissions, not restoration
of every historical shape. If scope grows, narrow shapes before adding phases.

No additional milestone research round is needed. Phase 22 uses standard build/process
patterns. Phases 23–24 need bounded design decisions from this research: source constructibility,
buffer/error representation, any necessary owning payload, call transfer, and independent
discharge. Phase 25 needs exact fixtures and actual host evidence, not another ecosystem survey.
The following milestone needs a scalar-loop spike on fixed points, back-edge admission,
dynamic event occurrences, numeric definedness, and evidence budgets before implementation.

Each phase starts with a source/input/output witness and introduces its independent proofs
alongside the capability. A second consecutive enabling phase without runnable gain triggers
scope review. Assign expensive lanes one owner; preserve negative controls and rerun evidence
when its inputs or claims change.

### Current Three Capability Recommendations

| Program / priority | Blocker, smallest slice, and checker trigger | Evidence/debt, owner, and priority-change observation |
|---|---|---|
| 1. Caller-driven scalar executable | Canned input/JSON/replay; retained build and one execution; entry/process/evidence contract. | Two expected results and launch observation; Phase 22; revisit if resource input reveals a smaller reusable boundary. |
| 2. Bounded file-byte utility | Foreign refusal/first-symbol binding/freed-in-shim storage; return live allocation and use/release through Lang; acquisition/transfer/error obligations. | Physical lifetime controls, D-16-11, Phase 21 successor; Phases 23–24; narrow if growth lacks a witness or safety obligation. |
| 3. Iterative sum and FizzBuzz | Missing operators/control/loops/text; scalar iteration plus minimal output; numeric rules, CFG and event changes. | Cycle refusal/checksum frontier; following milestone; change only for a concrete consumer or measured constraint. |

This ranking does not defer M004's two pointer-family commitments.

## Confidence Assessment

| Area | Confidence | Notes |
|---|---|---|
| Stack | HIGH | Existing implementation and constraints agree; no latest-version or new host-pass claims. |
| Features | HIGH scope / MEDIUM sizing | Directly inspected blockers and user-adopted priorities; syntax and effort remain design work. |
| Architecture | MEDIUM | Concrete source/contract findings; proposed dynamic ownership and pointer successors are not yet admitted. |
| Pitfalls | HIGH observed risks | Specific mechanisms, historical receipts, and focused planning/mutation checks; proposed resource controls remain future evidence. |

### Gaps to Address

Phase 22 must settle input/evidence transport, signal/exit and capacity behavior, build inputs,
and declared runtime dependencies. Phase 23 fixes bounded representation, empty/failure state,
and any payload prerequisites. Phase 24 defines call/error ownership and independent dynamic
identity. Phase 25 requires actual native receipts on both hosts for each pointer family.
Independent physical observers and reached mutations are mandatory; process reclamation cannot
hide leaks. Explicit C inputs remain trusted build authority and external inputs remain bounded.
Historical timings cannot establish current end-to-end latency or a dependable p95.

**Provenance:** implementation inspection uses `d9bde05`; planning repairs are recorded in
`03d41d7`. FEATURES/ARCHITECTURE report inspection without new runtime tests; PITFALLS reports
focused mutation, debt, claims-view, and historical-owner guard checks. Those establish their
narrow claims, not new M004 production admission. This synthesis adds no runtime evidence.
Phase 21's one-TU pure-Lang Darwin LTO receipt does not prove linked adapter cleanup or other hosts.

## Sources

- [STACK](STACK.md), [FEATURES](FEATURES.md), [ARCHITECTURE](ARCHITECTURE.md), [PITFALLS](PITFALLS.md): complete findings, implementation anchors, alternatives, and scoped executed evidence.
- [PROJECT](../../PROJECT.md), [PRODUCT-ROADMAP](../../PRODUCT-ROADMAP.md), [LANGUAGE-MATURITY](../../LANGUAGE-MATURITY.md): current adopted scope and observed capability.
- [M003 requirements](../../milestones/M003-REQUIREMENTS.md), [Phase 16 debt](../../milestones/M003-phases/16-branch-match-emitter-port/PHASE-16-DEBT.md), [Phase 16 decisions](../../milestones/M003-phases/16-branch-match-emitter-port/16-CONTEXT.md): NAT-09 lineage and D-16-07/11/12/13.
- [Phase 21 contract](../../milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-RESOURCE-DISCHARGE-CONTRACT.json), [verification](../../milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-VERIFICATION.md), [LTO receipt](../../milestones/M004-phases/21-native-emission-ownership-and-resource-discharge-m004/21-LTO-EVIDENCE.md): archived prework and bounded evidence.
- [Clang toolchain](https://clang.llvm.org/docs/Toolchain.html), [LLVM attributes](https://llvm.org/docs/LangRef.html#parameter-attributes), [WG14 N1570](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf): build and optimizer/cleanup contracts. N1570 is a C11 draft, not a fetched C17 standard.
- [ASan](https://clang.llvm.org/docs/AddressSanitizer.html), [UBSan](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html), [LLVM testing](https://www.llvm.org/docs/TestingGuide.html): distinct bounded evidence tools.
- [Rust FFI](https://doc.rust-lang.org/nomicon/ffi.html#calling-foreign-functions), [no_std](https://doc.rust-lang.org/stable/embedded-book/intro/no-std.html), [NLL RFC](https://rust-lang.github.io/rfcs/2094-nll.html), [MIR RFC](https://rust-lang.github.io/rfcs/1211-mir.html): trust/runtime/CFG precedents, not imported Lang semantics.
- [RFC 8259](https://www.rfc-editor.org/rfc/rfc8259.html): later JSON format scope, independent of HTTP.
- [M003 synthesis](../M003/ADVERSARIAL-SYNTHESIS.md), [audit](../../milestones/M003-MILESTONE-AUDIT.md), [Phase 20 timings](../../milestones/M003-phases/20-nyquist-d-13-34-and-the-frontier-fixture/20-CLOSURE-TIMING.md): historical sizing and cost, not future forecasts.

---
*Ready for requirements and roadmap: yes; no new production admission claimed.*
