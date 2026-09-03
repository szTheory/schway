---
id: ownership-evidence-roadmap
title: Ownership evidence roadmap
summary: The executable ownership work completed so far, the ordered next experiments, their decision gates, and the point at which the project should move into a real frontend and native lowering.
type: strategy
status: active
confidence: high
created: 2026-09-03
updated: 2026-09-03
tags: [ownership, experiments, roadmap, testing, convergence]
related: [ownership-lifetime-decisive-study, ownership-experiment-readiness-audit, semantic-kernel-contract, semantic-kernel-probes, convergence-audit, convergence-work-program, implementation-path, research-ledger, open-questions]
---

# Ownership evidence roadmap

## Reader and action

This note is for the next contributor or agent choosing what to implement. After
reading it, they should be able to start the next bounded ownership experiment,
know what evidence already exists, and avoid either repeating broad design
research or mistaking a Go workbench for the production compiler.

## Current position

There is executable code, not only Markdown. It lives under the spike area
because both the notation and implementation are disposable experiment hosts:

| Spike | Evidence | Status |
|---|---|---|
| [ownership kernel workbench](../.planning/spikes/001-ownership-kernel-workbench/README.md) | linear moves/loans, lexical cleanup, scopes, branch joins, loop place states, 7,381 generated sequences, fault injection | validated within its bounded model |
| [CFG edge-specific last use](../.planning/spikes/002-cfg-edge-last-use/README.md) | backward loan liveness, point/edge endpoints, nested branches, loop exits, 100 generated CFG products, 500 seeded metamorphic trials, 1,501-block scale case | validated within local first-order loans |
| [public origins and generic abilities](../.planning/spikes/003-public-origins-generic-abilities/README.md) | body-blind package interfaces, value/field origins, tagged alternatives, access modes, scoped callbacks, five independent abilities, 24 fixtures, 66 generated calls, 500 seeded trials | validated within first-order separate-compilation summaries |
| [independent certificate checker](../.planning/spikes/004-independent-certificate-checker/README.md) | canonical typed-core binding, independently recomputed summaries/abilities/flows, 13-class mutation matrix, 500 canonicalization trials, 10,001-event scale comparison | partial: validates committed typed-core facts but cannot prove source-to-core truth |
| [native FFI, provenance, and cleanup](../.planning/spikes/005-native-ffi-provenance-cleanup/README.md) | separate C translation units, O0/O3 semantic comparison, layout/cleanup/allocator/retention faults, false no-alias optimization, ASan UAF, nonlocal exit | partial: native obligations validated on Apple arm64, but no Lang IR lowering exists yet |

Go 1.24 is the lab host because it is installed, fast, dependency-free here,
and suitable for explicit state machines. This is not a decision to implement
the production compiler in Go.

## What the testing strategy actually is

The project does not rely on a single example suite or on generated tests that
merely restate the implementation.

| Layer | Job | Current form |
|---|---|---|
| deterministic fixtures | pin named happy, error, boundary, and regression cases | versioned JSON corpora |
| differential implementations | expose implementation drift | separately coded local checker/oracles, normalizers, and public consumer/body oracle |
| bounded exhaustive generation | cover every short program in a declared alphabet | 7,381 linear programs and 100 CFG products |
| property/metamorphic checks | test invariants across representation changes | seeded block/edge/function/origin reorder and alpha-renaming trials |
| model/path oracle | compare a scalable abstraction with a slower precise reference | finite dataflow versus bounded path expansion |
| fault injection | prove the harness detects realistic unsoundness | move-during-loan, unsound join, omitted-edge, and omitted-public-origin defects |
| intermediate-artifact comparison | catch defects hidden by identical final outcomes | normalized programs, liveness sets, endpoint identities |
| scale/adversarial series | reveal cliffs and wasted work | explicit block/edge/transfer counts and 1,000-origin interface growth; richer distributions still pending |
| native differential evidence | catch ABI, optimizer, layout, and provenance mistakes | not started |

This is property- and model-based testing in substance. A library is not the
defining feature; declared invariants, generated input spaces, independent
oracles, reproducible seeds, and minimized counterexamples are. Add a library
only when shrinking or generation complexity beats a small inspectable local
implementation.

## Decisions now supported

1. Ordinary local loans should end automatically at proven last use. Source
   should not require routine `end_loan` ceremony.
2. Last-use facts belong to point-or-edge semantic identities. The IR and
   diagnostics preserve both even if source syntax exposes neither.
3. The fast checker starts as monotone intraprocedural dataflow over finite
   facts. Exhaustive paths remain a bounded oracle, never the production path.
4. Healthy agent output defaults to compact summaries. Full evidence remains
   addressable on mismatch or request.
5. Exported borrowed results use compiler-verified parameter/field origin paths,
   tagged alternatives, and a separate shared/exclusive access fact.
6. `copy`, `drop`, `share`, `send`, and `escape` remain independent and
   propagate structurally through generic wrappers and erasure.
7. A higher-ranked fresh origin is available for genuine callback lending;
   callback-only APIs are not the universal borrow mechanism.
8. These decisions do not settle variance, async lending, field-sensitive
   disjointness, native representation, or call-site spelling.
9. Compact independent recomputation is useful at package, shared-cache, CI,
   and release boundaries. Full event snapshots are mismatch-only evidence,
   not an ordinary edit-loop certificate.
10. A content-bound certificate cannot detect a coordinated false frontend
    statement. Do not market artifact integrity as source-to-core correctness.

## Ordered experiment queue

Run these in order. Each stage either produces an artifact used by the next or
can invalidate a high-blast-radius assumption.

### Completed: 003 — public origin summaries and generic abilities

**Question:** Can package interfaces express owned results, single- and
multi-source borrowed views, borrowing iterators/callbacks, and noncopyable
generic containers without user-facing lifetime algebra?

**Compare:** value-parameter origin paths, explicit region binders, and
restricted borrowing combinators. Preserve `copy`, `drop`, `share`, `send`,
and escape abilities independently.

**Evidence:** interface-only consumer checking; separate-compilation fixtures;
single/multi-source views; higher-ranked callback pressure cases; generic and
erased containers; deliberate dishonest summaries; diagnostic and summary-size
measurements.

**Gate:** select value-parameter origins only if consumers can check the common
corpus without bodies, higher-ranked failures are explicit rather than
miscompiled, and ordinary signatures remain calmer than explicit region forms.

**Result:** passed within the declared scope. The value-origin form preserved
all 13 direct APIs with one explicit fresh-origin binder; explicit regions
needed 13 binders, while callback-only encoding adapted 8 APIs. This is a
semantic-interface decision, not a frozen syntax verdict.

### Completed: 004 — independent certificate/model checker

**Question:** Can a much smaller checker validate ownership summaries and a
typed core trace without sharing the frontend's inference implementation?

**Evidence:** independently encoded rules, seeded false certificates, mutation
testing across parse/type/summary/flow defects, reduced counterexamples, and
measured certificate/check cost.

**Gate:** keep certificates only if they catch a distinct defect class without
becoming a second full compiler or slowing the ordinary edit loop.

**Result:** partial. A 363-line source/body-blind verifier caught 12 adversarial
mutations and performed 20,004 counted checks over 10,001 events. Compact
evidence remained 242 bytes; full replay evidence grew to roughly 1.0 MB and
was slower. The expected coordinated frontend/summary lie escaped. Retain a
compact validation manifest at trust crossings, generate forensic snapshots
on demand, and use different evidence for source-to-core and native lowering.

### Completed: 005 — native layout, ABI, provenance, and cleanup

**Question:** Do the surviving semantics lower correctly through one minimal
native backend and C boundary under optimization?

**Evidence:** interpreter/native differential events; layout and calling
convention fixtures; partial initialization; allocator pairing; callback
retention; unwind/cancellation; pointer provenance; sanitizers; and deliberately
wrong backend attributes.

**Gate:** no native-safety claim until hostile optimized cases agree. Backend
speed or code quality cannot compensate for one unsound ownership lowering.

**Result:** partial. Six semantic events agreed at `-O0` and `-O3`, and all
seven hostile cases were observed. Most decisively, the false `restrict`
contract returned `2` at `-O0` and `1` at `-O3`; linker-compatible field-order
drift also corrupted a by-value result. The harness now pins ABI, initialization,
allocator, retention, cleanup, sanitizer, and nonlocal-exit obligations, but it
cannot validate a Lang lowering until the real typed core exists.

### 006 — ownership surface and diagnostic comparison

**Question:** Which source projection best exposes transfer, exclusive access,
and borrowed origins to an AI author and human reviewer without Rust-like
ceremony?

**Compare:** always-visible named transfer/access, ambiguity-only markers, and
declaration-only modes over identical accepted IR.

**Evidence:** generation, malformed-edit recovery, semantic repair, token/tool
calls, diff churn, and timed source-only audit on the ownership corpus.

**Gate:** syntax follows semantics and whole-loop evidence; character count
alone never decides it.

### 007 — real ownership dogfood verticals

Implement the same semantics in at least a streaming parser/data tool and one
native resource/FFI utility. Add an async service slice only after the local and
public contracts are stable enough that suspension tests are meaningful.

**Gate:** two unlike workloads must avoid unsafe global workarounds, excessive
copies, distant diagnostics, and compile-time cliffs before the ownership slice
graduates into a production frontend roadmap.

## Per-spike execution checklist

- State one falsifiable Given/When/Then question.
- Name the competing mechanisms and the strongest objection to each.
- Preserve stable input, semantic, diagnostic, and measurement identities.
- Write deterministic happy, error, and boundary fixtures first.
- Add an oracle that does not reuse the mechanism being validated.
- Enumerate the smallest relevant state space, then add seeded properties and
  risk-selected interaction coverage.
- Inject at least one plausible defect and require a reduced counterexample.
- Measure analyzer work, output volume, memory when relevant, and wall time as
  a distribution once a real performance gate exists.
- Keep healthy output compact and mismatch evidence lossless.
- Record what the evidence establishes, what it does not, and which decision it
  changes.
- Stop when the gate is answered; do not expand the spike into the compiler.

## Reopening and stopping rules

Do not return to broad ownership ideation merely because another library or
domain has an ownership detail. Reopen the semantic center only when a reduced
counterexample demonstrates unsoundness, two unlike realistic workloads need
the same missing primitive, a public summary cannot support separate
compilation, a native lowering contradicts the interpreter, or measured checker
work violates the feedback budget.

After experiments 003–005 establish public contracts, independent checking,
and one honest native path, begin a real versioned frontend/core-IR repository.
Surface comparison and dogfood can then proceed against a semantic artifact
rather than extending disposable Go notation indefinitely.

## Current leverage assessment

The project has crossed the point where another broad ownership taxonomy or
disposable semantic workbench is likely to outperform building the actual
artifact. Five different boundaries now have
evidence: local state transitions, CFG last use, separately compiled public
origins/abilities, independently replayed committed typed-core facts, and real
native/FFI optimizer hazards. Spikes 004–005 locate the remaining trust gap
precisely: only a real source-to-core-to-native pipeline can now test whether
the intended relationships survive translation.

### Near term

1. Begin the real versioned lossless frontend and portable typed core. Import
   the semantic fixture concepts that survived Spikes 001–005, but freeze no
   user syntax beyond the minimum parser experiment.
2. Make the Spike 004 validator and Spike 005 native harness consumers of that
   typed core. Their negative controls become required regression lanes.
3. Implement one minimal native development lowering and C ABI subset before
   expanding effects, async, actors, packages, or official application kits.

### Mid term

1. Create the real frontend/core-IR repository with a lossless parser,
   deterministic formatter, typed interface summaries, interpreter, and one
   minimal native development backend.
2. Run Spike 006's source/diagnostic comparison against that artifact rather
   than hand-authored pseudo-syntax.
3. Dogfood the corpus/evidence runner and one native resource/FFI utility, then
   add the streaming parser/data slice if it increases semantic coverage.

### Long term

1. Add effects/providers, architecture policy, deterministic test worlds, and
   agent queries after the native ownership kernel is honest.
2. Add async tasks, actors, backpressure, and optional managed regions only
   through workloads that require them.
3. Grow first-party JSON/schema, HTTP/TLS, SQL/change, telemetry, data, GUI, and
   AI kits over stable language contracts rather than moving those domains into
   the kernel preemptively.

The efficiency rule is now concrete: spend broad research on irreversible
boundaries, but require the next unit of work to produce a counterexample,
measured comparison, or reusable versioned artifact. A new domain concern goes
into the gauntlet until it demonstrates a kernel-level contradiction.
