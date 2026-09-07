# Phase 4: Fallible Resources and C Boundary - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-04
**Phase:** 4-fallible-resources-and-c-boundary
**Areas discussed:** Call surface, Failure representation, What the C boundary is, Panic & nonlocal exit
**Mode:** advisor (`minimal_decisive`), with four parallel research fan-outs at the
user's explicit request — each instructed to consider breadth and depth across
stakeholder-role lenses, ecosystem prior art with provenance, and an adversarial
pass, then synthesize one decisive recommendation.

---

## Area selection

The user selected all four presented gray areas, and directed that each be
researched by a fan-out subagent rather than decided inline.

| Option | Description | Selected |
|--------|-------------|----------|
| Call surface | SC1 needs a fallible two-step acquisition but no call construct exists | ✓ |
| Failure representation | SEM-03 wants `Result` propagation + ignored-result rules | ✓ |
| What the C boundary is | FFI-01 through "a real C call"; shim vs. libc; how obligations become inspectable | ✓ |
| Panic & nonlocal exit | SC3's two halves; build a panic construct or prove negatively | ✓ |

---

## Call surface

| Option | Description | Selected |
|--------|-------------|----------|
| `OpForeignCall` only | Foreign calls are the only callable surface; no Lang function ever calls another | ✓ |
| `OpCall` + `OpForeignCall` | Add the full call spine one phase earlier | |
| Lang-to-Lang only | Model the C boundary some other way | rejected outright |
| A builtin acquisition intrinsic not spelled as a call | Sidestep the call construct entirely | rejected outright |

**User's choice:** `OpForeignCall` only (D-04-01).

**Notes:** The decisive evidence was a constructed witness program under the
rejected option: `escort` calls `relay` — byte-for-byte the `exclusive_borrow_clean`
fixture 03-09 deliberately kept accepting with zero diagnostics — receives a value
`check` believes is owned with no live loan, moves the owner out from under it,
and returns the dangling alias. `lang check` accepts it. This is exactly the state
03-DEBT.md's D-03-02 described as "currently unreachable because there is no
consumer"; adding `OpCall` creates the consumer. No analogous program could be
constructed under the selected option, because a foreign function has no body for
a derivation to disagree with.

Secondary reasoning recorded: every Phase 3 exhaustive differential
(`TestOwnershipSequenceExhaustive` ~113k cases, `TestBranchSequenceExhaustive`
2,401 programs) is single-function by construction, and `checkBranch`'s per-arm
independence has no cross-function analogue — so shipping `OpCall` would leave
those tests green while covering a strictly smaller fraction of the admission
surface. That is the self-confirming shape WR-02 already cost a wave.

The lift condition was named rather than left implicit: **callable ⊆ publishable**,
reusing the proven 03-09/03-10 machinery rather than restoring the `check.go` gate
03-06 removed.

Rejected outright: modeling the boundary without generated C (nothing to inspect,
so SC2 has no subject) and a builtin acquisition intrinsic (carries no declared
allocator, capture, aliasing, or unwind policy, so FFI-01 has no subject and
attributes become decoration — the anti-pattern spike 005 iteration 4 caught).

---

## Failure representation

| Option | Description | Selected |
|--------|-------------|----------|
| Failure as a two-successor CFG edge; error payload is an existing nullary ADT on the `err` edge | No `Result` value, no generics, no payload alternatives; `try` / `discard … because` are the only consumers | ✓ |
| Monomorphic payload-carrying ADT (`\| Ok(Handle) \| Err(E)`) discriminated by `match` | Reuses existing match→CFG lowering; yields a real value | |
| Generic `Result<T,E>` in prelude | | rejected outright |
| Exceptions / unwinding failure | | rejected outright |
| C-style sentinel return codes at the Lang level | | rejected outright |
| `Result` value + `?` whose desugaring is a shared helper across both admission layers | | rejected outright |

**User's choice:** Failure as a CFG edge (D-04-04 … D-04-09).

**Notes:** The pivotal observation was that Phase 4 needs failure *propagation*,
not failure *data* — nothing in SC1 or SC4 requires a failure to be stored,
returned, passed, or algebraically combined, and every one of those would demand
payload-carrying alternatives plus eventually generics, both milestone
Out-of-Scope. Choosing the edge leaves `DataType.Alternatives []string` untouched,
so D-13's byte-identical goldens hold without a `lang.core/2`.

On the ignored-result rule, the ecosystem evidence was one-sided: Rust's
`#[must_use]` (RFC 1940) is a *lint* — suppressible, with documented coverage
holes (unused `==` escaped it because HIR sees binops as distinct from calls);
Go pushed the equivalent fully out of the compiler into `errcheck` and then
rejected its own `try` proposal; Zig made `!T` control flow and added `errdefer`
precisely because partial-acquisition rollback is the hard case, which is RES-01
verbatim. Midori's error model supplied the argument for the other half of
SEM-03 — separating recoverable errors from abandonment at the type-and-control-flow
level so a defect can never be caught as an error. The selected option inherits
that split for free: there is no `Err` value for a panic to become.

The sharpest single finding across the whole research set came from this area's
adversarial pass: **with two acquisitions the release set and the release order
coincide**, so a differential over the two-stage fixture SC1 literally asks for
stays green even if `cgen` derives its cleanup order from its own `goto`-ladder
structure rather than from the materialized `OpRelease` list. A three-acquisition
fixture plus a transposition mutation is therefore mandatory (recorded in
CONTEXT.md `<specifics>`).

The tradeoff was named explicitly: no storable, matchable failure value this
phase, and when error translation between layers is genuinely needed the
payload-alternative debt is paid at full price, because a `Result` value bolted
onto an edge-based core is a re-lowering rather than an extension.

---

## What the C boundary is

| Option | Description | Selected |
|--------|-------------|----------|
| Frozen hand-written foreign TU wrapping real libc `malloc`/`free`, independently declared, plus a generated conformance TU | Compiler never parses the foreign header; conformance TU is the single auditable meeting point | ✓ |
| Pure libc only, no shim | Maximal quarantine honesty | |
| In-repo shim with a shared header included by both sides | | rejected outright |
| In-repo shim whose C is generated from the same core program | | rejected outright |
| Header ingestion (bindgen / `translate-c` style) | | rejected outright |
| Obligations as prose comments only | | rejected outright |

**User's choice:** Hybrid frozen foreign TU over real libc, plus a generated
conformance TU (D-04-10 … D-04-12).

**Notes:** The reframe that settled the shim-vs-libc framing: quarantine is a
property of the **information flow, not the authorship**. The wiki's own words —
"a generated binding can create a candidate contract, never proof that the C
implementation obeys it" — make the test *whether the compiler is allowed to look
at the foreign declaration*, not who wrote it. A shim whose private header the
generated C includes destroys quarantine; a frozen TU the compiler never parsed
preserves it.

Pure libc was rejected on a decisive defect rather than a preference: `malloc`
has no record layout, so SC2's "target layout inspectable" has no subject at all,
and six of Phase 5's seven hostile mutations become non-injectable — leaving
Phase 5 nothing to attack.

The self-confirming-boundary objection was stated at full strength (one author
writes the `foreign` block, the fixture C, and the conformance asserts — the WR-02
shape) and answered structurally rather than by assurance: the conformance TU
makes alignment mechanical under `-Werror` rather than authorial; the two
mutation directions attack *different artifacts* (layout → the frozen fixture;
omitted release → the emitter's own output), so one aligned author cannot pass
both; the fixture is byte-frozen with a manifest digest; and the ABI, calling
convention, and allocator are the platform's.

The residual coordinated three-way lie was accepted and recorded rather than
engineered around — it is QLT-01's documented escape class one layer down, and
the alternative trades it for having nothing testable.

---

## Panic & nonlocal exit

| Option | Description | Selected |
|--------|-------------|----------|
| Minimal terminal `defect` (abort-only at process root, no catch/containment/unwinding) + structural non-unwind proof + root `setjmp` pad and static cleanup ledger | | ✓ |
| No panic construct; prove SC3's first half structurally and negatively only | | |
| Full panic + boundary shim that catches and translates | | rejected outright |
| Panic as a `Result` variant or error code | | rejected outright |
| Preventing (not detecting) foreign nonlocal exit | | rejected outright |
| Running structural release from the landing pad | | rejected outright |
| Unwind-section absence (`.eh_frame` / `__unwind_info`) as the control | | rejected outright |

**User's choice:** Minimal terminal `defect` plus detection-based nonlocal-exit
evidence (D-04-15 … D-04-20).

**Notes:** The wiki picked this without much room: `panic` means a defect,
ordinary code cannot catch it, "panic never unwinds across a foreign ABI … or
terminates the owning process boundary", and "an abort-only profile can host
panic only at its process/root boundary." Phase 4 has no isolation boundary to
contain a panic in and PROJECT.md forbids building one, so the abort-only profile
is the profile the contract names for exactly this situation.

The purely-negative option was rejected on this repo's own history rather than on
principle: a criterion whose evidence is "no construct exists that could do the
forbidden thing" has an empty reachable input space by construction — the maximal
instance of the failure mode that already cost 03-08, 03-09, and 03-10, and the
one standing rule D-10 exists to catch. Adding the `defect` terminator costs one
`OperationKind`, depends on none of the other three decisions, and converts the
claim into one with a reachable witness.

SC3's second half was adjudicated as **detection**, on the criterion's own wording
("cannot *silently* bypass"). Prevention against opaque C is unachievable and
claiming it would be an unproven attribute — the same class as the false
`restrict` that spike 005 iteration 4 showed changes the answer between `-O0` and
`-O3` on this host. Prevention is instead delivered statically at the declaration
layer: a foreign declaration that does not state its unwind/nonlocal-exit policy
is refused admission by both layers, with no default value.

Two constraints were verified empirically on this host (Apple clang 21, arm64)
rather than assumed:

- `-fno-exceptions -fno-asynchronous-unwind-tables -fno-unwind-tables` compiled
  clean under the repo's existing `-std=c17 -Wall -Wextra -Werror -pedantic`;
  `nm -u` returned exactly `_longjmp _printf _setjmp`; `otool -l | grep -c
  __unwind_info` returned `0`. Conclusion: `nm -u` with an underscore-normalized
  allowlist is the portable, discriminating control, and unwind-section absence
  is **not** — Darwin/arm64 mandates compact unwind for many function shapes, so
  a section assertion would pass or fail for the wrong reason.
- C17 §7.13.2.1 leaves non-`volatile` automatics changed since `setjmp`
  indeterminate after `longjmp`. The ledger must therefore live in `static`
  storage (as spike 005's `native/nonlocal_exit.c` already does), and the pad
  must emit evidence but **never** release — releasing from indeterminate handles
  converts a leak into a use-after-free.

A third hazard surfaced that is not a decision but a blocker: `native.Runner`
today rejects any nonzero exit and any stderr byte, so an aborting program cannot
be run at all, and today's buffered `lang_write_events()` means an aborting
process emits zero events. Both are recorded as integration points in CONTEXT.md.

---

## Cross-area conflict: optimizer-visible attributes

The four fan-outs agreed on everything material except one point. The call-surface
and panic fan-outs both proposed emitting `__attribute__((nothrow))` on foreign
declarations, derived from the admitted `unwind: forbidden` fact. The C-boundary
fan-out argued for emitting **zero** optimizer-visible attributes — including
`nothrow` — and making the refusal itself a fail-closed control.

| Option | Description | Selected |
|--------|-------------|----------|
| Emit nothing; the refusal is the control | `control:foreign.no_unproven_attributes` scans emitted C and the manifest's `emitted_attributes` and asserts both empty; mutation-kill injects `restrict` and the control must fail | ✓ |
| Emit `nothrow` only, derived from the admitted declaration fact | Makes the declared unwind fact load-bearing in the emitted C rather than only in a comment | |

**User's choice:** Emit nothing (D-04-13).

**Notes:** Phase 4 has no alias analysis, no capture analysis, and no proof
obligations to derive attributes from, and `nothrow` is an LLVM-visible promise
about a callee that cannot be inspected. The strongest honest Phase 4 statement
about the optimizer is a proven absence. `emitted_attributes: []` is kept as a
*field* rather than an omission so Phase 5 can assert on it. `_Noreturn` on the
generated `lang_defect` function is exempt (D-04-14) — it is not a claim about a
foreign callee but a property of a function cgen itself emits, every path of
which provably ends in `abort()`.

Consequence recorded for the roadmap: NAT-03's "false no-alias facts" hostile
mutation has **no subject** until Phase 5 first adds a proven alias-fact emission
path. That dependency is not currently stated in ROADMAP.md §Phase 5.

---

## Carried Phase 3 debt

| Option | Description | Selected |
|--------|-------------|----------|
| D-03-01: retire `discoverLoanLastUses` | The complete D-05 fix; drives both admission and endpoint materialization from `loanLivenessFixpoint` | deferred to Phase 5 |
| D-03-01: count its work honestly | Instrument the transitive scan per the `blockLoanLiveness` convention | ✓ |
| WR-01: wire `ValidatePublished` into `check`/`run` | | ✓ |
| Carry all of it to Phase 5 | | partially |

**User's choice:** "follow ur recs auto plz actually" — the user selected all four
options and then delegated the adjudication.

**Adjudication (Claude's, on delegation):** Take the cheap half of D-03-01 in
Phase 4 (D-04-25) and defer the retirement to Phase 5 (D-04-26) — performing that
surgery inside the same `check.go` that is simultaneously gaining `OpForeignCall`,
`OpFail`, `OpRelease`, and `OpDefect` is precisely the compounding-wave-defect
shape Phases 2 and 3 both hit. Close WR-01 this phase (D-04-27), because foreign
declarations become a second surface where an alias fact can be silently absent
and `originvalidate` must learn about foreign returns regardless (D-04-28) — folded
into whichever plan already touches that file, per D-08, rather than becoming its
own debt-cleanup plan.

---

## Claude's Discretion

Plan decomposition, wave structure, and plan count. The concrete spelling of the
`foreign` block and of `try` / `discard … because`. `OperationKind` naming — the
four fan-outs proposed overlapping sets (`OpForeignCall`,
`OpAcquire`/`OpResourceAcquire`, `OpRelease`/`OpResourceRelease`, `OpFail`,
`OpDiscard`, `OpDefect`) which must be reconciled into one minimal set at planning
time. The exact `lang.foreign/0` field layout, the fixture type shape, and the
diagnostic code names.

## Deferred Ideas

- Lang-to-Lang calls (`OpCall`) — Phase 5, gated on callable ⊆ publishable
- A storable, matchable `Result` value and payload-carrying alternatives — M002 or
  Phase 6, via an additive sibling `alternative_details` array
- Retiring `discoverLoanLastUses` — Phase 5
- Allocator-mismatch and stale-callback-retention detection — Phase 5 (NAT-03)
- ASan/UBSan lanes — Phase 5, isolated from semantic-equivalence evidence
- Callback registration, generation tokens, dynamic-library handles, `errno`
  translation, symbol versioning, `blocking`/`reentrant` enforcement, header
  ingestion — declared and refused-if-non-default, never exercised this phase
- Panic containment, isolation boundaries, `catch_unwind` analogues, cancellation
  shields/budgets/checks — Phase 6+
- Loop-carried loan liveness — still deferred; no loop construct is added
- DWARF/CodeView emission, crash capture, proof/SMT tiers — unchanged from D-03
