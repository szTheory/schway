---
id: ownership-experiment-readiness-audit
title: Ownership experiment readiness audit
summary: The final bounded unknowns pass: what must be fixed before implementation, what the first workbench must measure, which seams must remain open, and when further ideation should stop.
type: strategy
status: candidate
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [ownership, readiness, unknowns, experiment, convergence]
related: [ownership-lifetime-decisive-study, semantic-kernel-contract, semantic-kernel-probes, residual-uncertainty-register, convergence-audit, convergence-work-program, implementation-path, author-intent-model, research-ledger, open-questions]
---

# Ownership experiment readiness audit

## Reader and outcome

This audit is for the person implementing the first ownership workbench. After
reading it, they should be able to start without silently inventing a language
rule, know which uncertainty the workbench is meant to resolve, and avoid
turning an experiment into an accidental production compiler.

## Verdict

**Ready for a bounded executable experiment. No further general ownership
survey is justified before it begins.**

The final pass found no missing ownership family and no contradiction that
requires another paper design. It did find six cross-cutting seams that the 64
base workloads must exercise explicitly:

1. type abstraction: variance, higher-ranked access, erasure, and
   noncopyable generics;
2. termination: normal exit, error, panic, cancellation, abort, process exit,
   and power loss do not share one cleanup guarantee;
3. identity: value identity, resource identity, object address, equality, and
   hashing must not collapse into one concept;
4. long-lived storage: globals, statics, thread/task locals, memory maps, and
   shared memory have different initialization and shutdown laws;
5. concurrency reclamation: unique transfer is not a complete answer for
   lock-free structures, weak handles, ABA, epochs, or hazard ownership;
6. toolchain translation: generated code, optimized-away values, public
   summaries, debug views, and backend attributes can invalidate a sound
   source-level story.

These are overlays and seams, not six new universal language features. Their
discovery changes the experiment matrix and IR preservation requirements; it
does not overturn the leading value/access surface over affine ownership IR.

## The completeness criterion

Convergence does not mean that every future ownership decision is answered.
That is impossible before real programs and implementations exist. It means:

- every high-blast-radius question has a conservative starting law;
- every hard-to-retrofit fact has a place in the typed IR or evidence schema;
- every disputed ergonomic or performance claim has a falsifiable probe;
- every deferred mechanism has an admission trigger;
- no experiment result can be reported as success without checking its hard
  vetoes;
- new concerns can be classified without adding another semantic center.

The project now meets this criterion on paper. The next uncertainty reducer is
execution.

## Unknown lifecycle

Every unresolved ownership question belongs to exactly one horizon.

| Horizon | Meaning | Required treatment |
|---|---|---|
| Fix before workbench | The oracle would be ambiguous without it | choose a conservative rule and encode it in fixtures |
| Preserve before workbench | A retrofit would replace IR or evidence identities | carry the fact even if v0 rejects the advanced case |
| Resolve in laboratory | Competing answers are implementable and evidence can distinguish them | run the declared probe and record distributions |
| Resolve before native claim | Interpreter evidence cannot establish ABI, optimizer, or hardware behavior | require one native lowering and hostile boundary evidence |
| Resolve before public syntax | Semantics can execute while source spelling remains open | compare projections only after semantic survivors exist |
| Defer with trigger | Reversible or outside the first credible workload | name the event that reopens it |

“Unknown” without a horizon, owner, default, and probe is an unmanaged risk.

## Rules fixed before implementation

These are sufficient to build the workbench. Exact surface keywords are not
fixed.

### Values, places, and evaluation

- A value occupies one initialized place at a time unless its type is
  duplicable or it is behind an explicit shared representation.
- Place projections include fields, variants, indices, and dereferences. The
  workbench starts whole-value and field-sensitive; dynamic-index separation
  requires a checked projection operation.
- Evaluation order is strict and specified. Move, loan activation, effect,
  failure, and release events therefore have one observable order.
- Joining control-flow paths computes a conservative ownership state. A place
  used or released after the join must be initialized on every incoming live
  path.
- V0 rejects a residual aggregate after moving out one resource-bearing field.
  Whole-value consuming destructure is allowed.

### Access and origins

- Shared access permits reads and other shared access.
- Exclusive access permits mutation and conflicts with every other live access.
- Local access ends at last use when control-flow analysis proves it.
- A borrowed view carries an origin set and cannot escape the permitted extent.
- Public returned views name their source relationship. No package consumer
  needs a dependency implementation body to validate a call.
- Generic containers, callbacks, ports, and erased values preserve ownership
  modes and abilities rather than assuming copyability.

### Transfer, identity, and address

- Transfer invalidates the old place but does not promise that the physical
  address stayed stable.
- Value equality and hashing depend on declared value semantics, never on a
  transient storage address.
- Unique resources may have stable logical identity. Copying a handle does not
  duplicate the underlying authority unless the type explicitly defines shared
  ownership.
- Address stability is a separate property provided by a named owner. Moving
  the stable handle may not move the referent.
- Integer addresses do not recreate bounds, initialization, provenance, or
  permission.

### Resources and termination

- Language-mediated scope exit—return, expected error, contained panic, or
  cancellation—runs compiler-known structural release in reverse acquisition
  order.
- Meaningful fallible completion remains an explicit consuming operation.
- Body failure remains primary when fallback release also reports a problem;
  both remain inspectable.
- Abort, uncontained process termination, `SIGKILL`, host failure, and power
  loss do not promise in-process cleanup. Durability and recovery require
  protocols such as atomic replacement, journaling, idempotency, or an
  external supervisor.
- Static/global storage does not receive an implicit process-shutdown protocol.
  V0 permits constant/static data and explicitly initialized owned services;
  effectful global initialization and destructor-order dependence are rejected.

### Concurrency and unsafe boundaries

- Crossing an isolation boundary requires unique transfer or a declared
  immutable/thread-safe shared representation.
- An ordinary borrow cannot cross detachment, actor send, foreign retention, or
  suspension in the first checker.
- Shared mutable memory requires an explicit synchronization/protocol type.
- Unsafe authority is granular: initialization, bounds, provenance, alias,
  layout, atomic ordering, retention, unwind, and thread rules remain distinct
  obligations.
- A backend attribute such as `noalias`, `nonnull`, `dereferenceable`, or
  `captures(none)` may be emitted only from a written frontend proof rule.

## Facts that must be preserved now

The first implementation may conservatively reject these cases, but its IR and
evidence must leave room for them.

| Seam | Preserve now | Conservative v0 behavior |
|---|---|---|
| multi-source returned views | origin sets and path-sensitive result variants | keep every possible origin live |
| higher-ranked access | binder/scoping form in function types | permit only simple first-order public origins initially |
| variance | declared/inferred variance facts for origins and type parameters | invariant for exclusive or mutable views |
| noncopyable generics | ability constraints on parameters, associated types, and erased boxes | reject an operation whose required ability is absent |
| yielding access | temporary projection/yield ownership in callable IR | no generally escaping yielded borrow |
| stable storage | stable-owner identity distinct from handle place | safe builders only; no arbitrary self-reference |
| weak/remote handles | liveness check and upgrade failure in the type | no silent weak dereference |
| managed owner region | owner identity and cross-boundary snapshot/share operations | no direct managed reference escape |
| dynamic exclusivity | check event and failure kind | enabled only in interpreter/sanitizer-selected boundaries |
| concurrent reclamation | atomic ownership state and reclamation-domain identity | official lock-based structures first |
| memory mapping/shared memory | external storage owner, visibility, persistence, and invalidation | unsafe/provider boundary until specified |
| address spaces | ordinary, foreign, MMIO, GPU/device distinction | no implicit pointer conversion between spaces |
| globals/task locals | initialization owner, access effect, and teardown policy | constant data or explicitly scoped service owner |
| public ABI | parameter mode, result origin, abilities, layout, and release protocol | no stable native ABI promise yet |
| generated/derived code | semantic origin and ownership summary of generated declarations | generated code receives the same checks and provenance |
| optimized debug state | stable semantic value/place IDs independent of machine location | explicitly report unavailable physical values |

Rust's reference demonstrates why variance follows the position of lifetime and
type parameters rather than being an aesthetic annotation. Swift's
noncopyable-generic evolution demonstrates how quickly copyability assumptions
spread through option/result, protocols, existentials, and standard-library
APIs. ([Rust subtyping and variance](https://doc.rust-lang.org/reference/subtyping.html),
[Swift noncopyable generics](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0427-noncopyable-generics.md),
[Swift noncopyable standard-library primitives](https://github.com/swiftlang/swift-evolution/blob/main/proposals/0437-noncopyable-stdlib-primitives.md))

## Questions resolved by the laboratory

These should not be answered by another prose vote.

| Question | Leading candidate | Evidence that decides it |
|---|---|---|
| named transfer marker | visible for a named noncopyable value | source-only audit, AI repair, diff churn, parse recovery |
| exclusive-call marker | visible when mutation authority is granted | alias-error repair and human cost audit |
| private mode inference | infer locally and render on request | predictability, diagnostic locality, public-summary stability |
| origin syntax | value-parameter relationships rather than named lifetime variables | multi-source, iterator, closure, callback, and generic corpus |
| field sensitivity | fields yes; dynamic indices through checked split/projection | false rejection, checker work, diagnostic complexity |
| partial moves | whole-consuming destructure only in v0 | wrapper/copy burden on resource and parser aggregates |
| structured task lending | specialized lexical child loan | zero-copy benefit against cancellation/exclusivity complexity |
| flow-sensitive isolation | disconnected-region transfer behind calm diagnostics | graph handoff ergonomics and incremental-check cost |
| dynamic exclusivity | narrow interpreter/sanitizer fallback | safe-program acceptance against runtime failure/cost |
| must-resolve obligation | only for semantically invalid abandonment | branching/error-path ceremony versus caught protocol defects |
| immutable sharing | precise RC/reuse candidate | retain traffic, contention, release tails, code size, memory |
| managed cyclic region | one owner-confined provider | graph ergonomics and latency against arenas/indices/weak edges |
| checker architecture | cheap screen followed by precise local analysis | adversarial complexity and warm-edit distributions |
| human diagnostic | value timeline and violated relationship | timed comprehension and correct repair |
| machine diagnostic | stable cause graph and semantic edits | multi-model first-pass/repair/token/tool-call outcomes |

## Questions requiring native evidence

Interpreter success is insufficient for:

- exact layout, padding, niches, unions, alignment, and calling convention;
- pointer provenance, capture, alias, and lifetime claims passed to an optimizer;
- unwinding, `longjmp`-like control, callbacks, reentrancy, and dynamic unload;
- uninitialized storage and partial construction under optimization;
- volatile/MMIO, atomics, weak memory, and interrupt-visible state;
- allocator pairing, memory maps, DMA, and foreign retention;
- stable address/self-reference and stack movement;
- process/thread termination and cleanup that the host runtime may skip;
- sanitizer coverage and interpreter/native disagreement.

LLVM now distinguishes pointer address capture from provenance capture and
assigns strong behavioral meaning to lifetime and `noalias` facts. That makes
backend lowering a proof obligation, not a mechanical type translation.
([LLVM language reference](https://llvm.org/docs/LangRef.html))

## Cross-cutting adversarial overlays

The 64 workloads remain the base corpus. Each is not multiplied blindly by
every configuration. Instead, generated cases apply these orthogonal overlays:

| Overlay | Representative levels |
|---|---|
| control | straight line, branch join, loop fixed point, early return, panic |
| value shape | scalar, nested aggregate, variant, recursive, erased/generic |
| access | own, shared, exclusive, transfer, explicit share, weak handle |
| lifetime | local, returned view, stored closure, suspended task, foreign retention |
| termination | success, expected error, cancellation, contained panic, abort simulation |
| storage | inline/stack, unique heap, arena, immutable share, managed owner, foreign |
| isolation | one task, lexical child, detached task, actor, shared memory, process boundary |
| compilation | interpreter, debug native, optimized native, sanitizer/evidence mode |
| evolution | private body edit, public summary edit, mixed package version, ABI adapter |
| observer | compiler diagnostic, AI repair, source-only human audit, runtime telemetry |

Use exhaustive enumeration for the small ownership state machine. Use
pairwise covering arrays across all overlays, stronger three- to six-way
coverage for combinations selected by the threat model, and targeted full
cross-products for known cliffs such as borrow × cancellation × cleanup ×
foreign retention. Property generation and mutation remain separate layers.

NIST's combinatorial-testing work shows how covering arrays exercise all
selected t-way interactions without executing the infeasible Cartesian
product. It does not determine Lang's required interaction strength; that is
selected from consequence and observed defects. ([NIST ACTS](https://csrc.nist.gov/Projects/Automated-Combinatorial-Testing-for-Software/),
[NIST practical combinatorial testing](https://www.nist.gov/publications/practical-combinatorial-testing-beyond-pairwise))

## Newly explicit pressure cases

These are overlays or variants of existing workloads, not new language
features:

1. a returned view through a covariant immutable container and an invariant
   exclusive container;
2. a higher-ranked callback that may borrow any caller-owned value only for the
   call;
3. a noncopyable value behind a structural port/existential and through
   `Option`/`Result`;
4. a value used as a map key before and after moving its owning handle;
5. a stable referent whose movable handle is transferred between tasks;
6. static data, explicitly initialized service state, task-local state, and
   shutdown without destructor-order dependence;
7. contained panic versus abort/power-loss simulation during resource update;
8. mapped-file invalidation, shared-memory visibility, and external truncation;
9. ABA/reuse in a lock-free stack compared with an official locked structure;
10. weak-handle upgrade racing owner termination without resurrection;
11. C `setjmp`/`longjmp`-like escape and a callback racing library unload;
12. generated binding or derived codec that lies about retention/copyability;
13. debug inspection of moved or optimized-away source values;
14. hostile generic/control-flow growth intended to trigger checker blow-up;
15. stale cached ownership summary after a public mode/origin change;
16. independent checker and native backend agreeing because both consume one
    incorrect frontend summary.

## Specialist ownership of risks

| Accountability owner | Owns the decision/evidence | Cannot delegate away |
|---|---|---|
| language semanticist | value/place/access/origin/ability laws | ambiguous observable behavior |
| type-system engineer | generic, variance, higher-ranked, and erasure rules | nontermination or unsound substitution |
| checker engineer | local algorithms, summaries, incremental invalidation | adversarial complexity and distant diagnostics |
| runtime engineer | release, sharing, managed owner, and termination behavior | hidden latency or global services |
| native/backend engineer | layout, alias, provenance, unwind, atomics | unjustified optimizer attributes |
| FFI maintainer | allocation pairing, retention, callback, unload, thread contract | treating a header as proof |
| concurrency engineer | transfer, sharing, structured lending, reclamation | data races, ABA, deadlock, leaked tasks |
| library designer | noncopyable containers, iterators, ports, projections | requiring unsafe for ordinary composition |
| realtime/performance owner | allocation/release/checker distributions | averages that hide tails |
| security owner | unsafe proof surface and authority separation | laundering trust through ownership |
| AI-eval owner | diagnostic context, repair actions, model diversity | benchmark overfitting or token-only scoring |
| human reviewer | source-only audit tasks and diff clarity | IDE-only correctness stories |
| evolution owner | source/API/ABI/summary migration and mixed versions | accidental permanent compatibility |
| verification owner | independent oracles, reducers, negative tests | shared-implementation false agreement |

Each hard-veto result has exactly one accountable owner even when several
specialists contribute evidence.

## Decision gates

### Gate 0 — workbench readiness

Pass when fixtures have stable IDs, the state/event vocabulary is written, and
every implemented rule points to a candidate law. This document completes the
paper portion; the workbench must make it executable.

### Gate 1 — dynamic oracle coherence

Pass when a deterministic interpreter executes positive programs and produces
the expected concrete violation for deliberately invalid programs. Exhaustively
enumerate the small state machine. No static checker claim is made yet.

### Gate 2 — minimal static agreement

Pass when the checker accepts/rejects the Stage 1–2 corpus, distinguishes a
law violation from conservative incompleteness, and agrees with the dynamic
oracle on reduced invalid executions.

### Gate 3 — native semantic agreement

Pass when one unoptimized and optimized native lowering agrees on results,
initialization, release, and declared semantic events for the blocking native
subset. Sanitizer and hostile FFI cases must run.

### Gate 4 — ergonomic selection

Pass when the surviving semantics are rendered through alternative source
surfaces and measured with multiple AI models and source-only human reviewers.
Only here may transfer/exclusive markers become syntax decisions.

### Gate 5 — storage admission

Pass unique/arena as the baseline first. Admit precise RC/reuse or a managed
owner only after the relevant graph/sharing workloads demonstrate a whole-loop
win and acceptable tail behavior.

## First bounded experiment

Build an **ownership kernel workbench**, not a language frontend.

It contains:

- a tiny typed instruction form for bindings, moves, loans, uses, mutation,
  release, branches, and lexical scopes;
- a deterministic dynamic oracle with stable allocation/place/loan/resource
  identities and JSON event output;
- a minimal static checker for definite initialization, whole-value moves,
  shared/exclusive loans, last-use loan ends, and lexical release;
- fixtures containing positive, negative, error-path, and boundary cases;
- exhaustive short-sequence generation for the state machine;
- a differential runner comparing checker decisions with dynamic violations;
- measurement of wall time, analyzed operations, diagnostics, and reduced
  counterexamples.

It deliberately omits:

- final source syntax and parser recovery;
- general generics, effects, actors, async lowering, or managed memory;
- production allocation or garbage collection;
- a native backend until Gates 1–2 pass;
- claims about human/AI ergonomics from kernel notation.

The workbench is successful even if it invalidates the leading model. Its job
is to turn disagreement into a small reproducible program.

### First execution result

The [first ownership kernel workbench](../.planning/spikes/001-ownership-kernel-workbench/README.md)
now implements the linear slice, acyclic branch joins, and a finite ownership
loop fixed point. Its final run passed 28 linear/scope fixtures, 12
control-flow fixtures, 7,381 exhaustive short programs, and 64 generated
branch products with no unexpected disagreement. A deliberately faulty
checker was detected after 159 programs and reduced to a four-operation
move-while-borrowed case; a separate unsound maybe-initialized join was also
detected against exhaustive path expansion.

The spike also discovered and corrected an ambiguity in reused loan evidence
identities. Loan IDs are now unique for a program. This is evidence that the
workbench shape can expose specification-support defects rather than merely
confirming happy paths.

An independently implemented loan-end normalizer then found a second support
defect: an explicit `end_loan` before a borrow had incorrectly suppressed the
checker's inferred endpoint. Result-only comparison had hidden it because the
program failed on the early end first. The two normalizers now agree over all
fixtures and 7,381 generated programs, although agreement is still not a
formal proof of their shared intended law.

Gate 1 passes for the bounded linear state machine. Gate 2 remains partial:
whole-value branches, explicit branch-crossing loans, zero-or-more loops, and
nested lexical resource scopes now execute. Control-flow last-use inference,
wider place sensitivity, and an independent formal model of loan-end analysis
remain. No native or source-syntax claim has advanced.

## Deferred mechanisms and reopening triggers

| Deferred mechanism | Reopen only when |
|---|---|
| general lifetime-parameter syntax | value-parameter origins fail two unrelated public APIs |
| partial residual moves | two realistic workloads require them without tolerable replacement/destructure patterns |
| general borrow across `await` | structured task ownership causes measured copies or ceremony in two zero-copy workloads |
| universal flow-sensitive region transfer | nominal unique transfer materially harms graph handoff |
| dynamic checked references everywhere | static/projection models reject important safe programs without a local redesign |
| universal tracing GC | owner-confined managed graphs cannot support required workloads |
| lock-free standard structures | measured contention/latency justifies their proof and reclamation burden |
| arbitrary effectful globals | explicit service ownership cannot express a required platform integration |
| stable native ABI | a real external consumer needs independent compiler/runtime upgrades |
| self-hosting | Lang can build and debug its own conformance/evidence tools reliably |

## Stop rule

The final fan-out stops here because:

1. every newly found concern mapped to an existing semantic axis;
2. the six missing seams can be preserved or tested without selecting another
   language family;
3. every foundational unknown has a conservative default and decision gate;
4. further argument cannot provide checker complexity, diagnostic, runtime,
   native, or human/AI evidence;
5. the bounded workbench is smaller and more reversible than another design
   cycle.

Reopen broad exploration only if the workbench produces an irreducible
counterexample that crosses the established axes, or two independent realistic
workloads require the same missing primitive. Otherwise, add a fixture,
overlay, library design, profile rule, or later proposal.

## Evidence limits

- Existing-language semantics demonstrate feasible mechanisms and recurring
  interactions; they do not prove Lang's synthesis is ergonomic or fast.
- Covering arrays reduce combinatorial explosion; they do not replace targeted
  threat cases, exhaustive small-state exploration, fuzzing, model checking, or
  native execution.
- Interpreter/checker agreement can share a mistaken specification. Native
  differential execution, hostile tests, independent validation, and a small
  formal core remain separate evidence.
- Deterministic language cleanup cannot guarantee recovery from process loss or
  power failure. Durable correctness belongs to explicit external protocols.
