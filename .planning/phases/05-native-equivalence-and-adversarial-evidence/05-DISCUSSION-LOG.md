# Phase 5: Native Equivalence and Adversarial Evidence - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-09-05
**Phase:** 5-native-equivalence-and-adversarial-evidence
**Mode:** advisor (USER-PROFILE.md present; calibration tier `minimal_decisive`;
`technical_background: true` so no plain-language reframing applied)
**Areas discussed:** Alias-fact emission, Callback retention subject, Allocator
mismatch + UAF, Sanitizer lane design, Equivalence corpus + axes, Minimization +
causal trace, QLT-01 spike control audit, Carried debt + `OpCall`

**Interaction shape:** the user selected all eight gray areas in one pass, supplied
a standing instruction to fan out across every relevant technical stakeholder lens
with an adversarial pass before synthesizing, and directed that the synthesized
recommendations be adopted automatically rather than returned for per-area
selection. Eight `gsd-advisor-researcher` agents ran in parallel; their outputs
were synthesized, adjudicated where they conflicted, and one was materially
corrected against the shipped tree (see "Adjudications" below).

---

## Alias-fact emission

| Option | Description | Selected |
|--------|-------------|----------|
| `restrict` on cgen-emitted function parameters only, derived from a proven exclusive borrow / unique ownership fact | Smallest attribute surface; C17 §6.7.3.1 semantics match a call-scoped exclusive loan exactly; reuses D-04-14's "attribute only on a function cgen owns end-to-end" boundary | ✓ |
| Broader attribute set now (`nonnull`, LLVM-level `noalias`, attributes on foreign extern declarations) | Closes more of NAT-03's theoretical surface at once | |

**Outcome:** D-05-01, D-05-03, D-05-04, D-05-05.
**Notes:** Rust's `&mut` → `noalias` history (rust-lang/rust#31681, PR #31545,
#54878, #84958 — disabled 2016, gated behind `-Zmutable-noalias` for years,
re-enabled on LLVM 12 in 2021, regressed once more) was the decisive argument for
stopping at C-source `restrict`. `control:foreign.no_unproven_attributes` is
narrowed rather than deleted, and gains a second obligation: every entry in
`emitted_attributes` must have a `corevalidate`-re-derived justification.

## Adjudications — where the research was overridden or corrected

**Alias-fact emission (materially corrected).** The research agent's
recommendation assumed cgen emits pointer parameters it could qualify. A check of
the shipped tree during synthesis found otherwise: `cgen.go:157` emits Lang
functions with **by-value** parameters, and `testdata/phase2/owned_transfer.golden.c`
confirms the only pointer parameters in emitted C belong to cgen's own runtime
helpers. A by-value struct parameter cannot be meaningfully `restrict`-qualified,
so the recommendation as returned had no subject either — the same defect it was
commissioned to fix, one layer down. Recorded as **D-05-02**: the by-pointer
lowering is a blocking precondition and a separate deliverable, built on D-04-20's
additive-emitter-path pattern, with an explicit escalation rule forbidding the
tempting substitutions (qualifying a by-value parameter, or qualifying a runtime
helper whose non-aliasing is a cgen convention rather than a checked fact).

**Sanitizer opt level vs. UAF observability (conflict resolved).** The
allocator/UAF research asked for a fixture shaped so `-O3` cannot dead-store
eliminate the dangling read; the sanitizer-lane research recommended building the
lane at `-O1`. Resolved in favour of `-O1` for the lane (D-05-12) with the
observability requirement retained at that level (D-05-09) — the two are
compatible and `-O1` carries the lower false-positive and blame-obscuring risk.

**Sanitizer determinism vs. DX (trade accepted).** `symbolize=0` /
`print_stacktrace=0` were adopted for evidence stability (D-05-13, D-05-16) at
the cost of a less readable first-line diagnostic; raw stderr is retained under
the existing bounded-writer caps so the detail is not lost.

---

## Callback retention subject

| Option | Description | Selected |
|--------|-------------|----------|
| Build a minimal synchronous single-shot callback seam | Gives NAT-03 a literal callback subject | |
| Reframe as a retained-pointer / escaped-borrow defect in a hostile frozen foreign TU, with the callback-invocation form recorded as a named escape | No new call surface, no new `OperationKind`, locates the defect at the actual sink | ✓ |

**Outcome:** D-05-06, D-05-07.
**Notes:** Two arguments decided it. First, cost symmetry with Phase 4: D-04-15
built the `defect` terminator because it cost one `OperationKind` and depended on
nothing; a callback seam is calls-into-Lang, reopening D-04-01 and creating the
consumer for D-03-02. Second, the cheap version is self-defeating — a synchronous
single-shot callback that fires before anything can go stale has an empty
reachable input space for staleness, re-creating D-10's failure mode while looking
closed. Ecosystem survey (GObject `GWeakRef`, JNI weak-global refs, N-API
`napi_ref`, Rust `Box::into_raw` + `catch_unwind`) put the defect at the retained
context pointer, not the invocation mechanism; ASan's documented blind spot
(google/sanitizers#652) is specifically deferred callback firing, so a real seam
would not have improved coverage. NAT-03 is discharged as six subjected plus one
subsumed-with-named-escape rather than a silent 7/7.

---

## Allocator mismatch + UAF

| Option | Description | Selected |
|--------|-------------|----------|
| Allocator mismatch: keep Phase 4's static contract refusal as the sole control | No new TU, no sanitizer dependency | |
| Allocator mismatch: add a second frozen foreign TU with a real second allocator, detected dynamically by ASan `alloc-dealloc-mismatch`, alongside the existing static refusal | Attacks a different artifact than the static check; reuses an existing load-bearing ASan diagnostic | ✓ |
| UAF: hostile variant of a frozen foreign TU returning a pointer into freed storage, detected by ASan quarantine/redzone | Attacks the foreign fixture, avoiding collision with the shipped release-omission controls | ✓ |
| UAF: mutate cgen's own release emission, or mutate core to defeat loan liveness | Would technically produce a UAF | |
| UAF: standalone hostile C file unrelated to the boundary | Simple to write | |

**Outcome:** D-05-08, D-05-09, D-05-10.
**Notes:** The rejected UAF options fail the "seven distinct questions"
constitution — mutating cgen's release emission collides with Phase 4's
already-shipped release-omitted/transposed controls, and a standalone hostile file
tests ASan rather than this compiler. The single most likely concrete false green
in the whole phase was identified here: `alloc_dealloc_mismatch` defaults to
**off on macOS**, this project's primary host.

---

## Sanitizer lane design

| Option | Description | Selected |
|--------|-------------|----------|
| Dedicated `sanitize` build: separate compile+link, separate binary, own lane ID, `-O1 -fsanitize=address,undefined -fno-sanitize-recover=all`, gated to verify/release | Structurally cannot contaminate the equivalence differential | ✓ |
| Same-binary sanitizer-tagged run reusing the differential lane infra with a taint flag | Less new code | |

**Outcome:** D-05-11 through D-05-17.
**Notes:** Isolation is enforced at the type level — the comparator's signature
does not accept a sanitizer-lane evidence value — because a boolean taint flag on
a shared table is one refactor away from laundering sanitizer noise as a semantic
mismatch. `-O1` chosen over both `-O0` (too far from shipped codegen) and `-O3`
(false positives, inlining obscures blame; no major sanitizer CI runs its primary
tier at full optimization). Spike 005 iteration 5 — a real cleanup bypass with no
sanitizer report — is the standing caution that a clean run is weak evidence,
which is why the lane carries an always-on positive control.

---

## Equivalence corpus + axes

| Option | Description | Selected |
|--------|-------------|----------|
| Widen the hand-written corpus only, keep `-O0`/`-O3`, no LTO | Minimal new machinery, fully auditable | |
| Hand-written adversarial subset + bounded enumerated closure over the current grammar, with a sampled `-O3 -flto` tier and an asserted exclusion list | Makes the `-O3` claim non-vacuous and the coverage claim falsifiable | ✓ |

**Outcome:** D-05-18 through D-05-22, and the `<specifics>` LTO finding.
**Notes:** The sharpest result of the whole fan-out. Without `-flto`, a
separately-compiled foreign TU is opaque at the call site, so `-O3` cannot legally
perturb the code paths SC1 is about — every current corpus program's only
observable effect crosses exactly that boundary, making today's differential green
by construction. Two axes SC1's wording silently omitted were added (exit
status/signal; reject-program diagnostic-ID equivalence under its own control),
and the exclusion list moves from wiki prose into an asserted, fail-closed
field-routing test. The wiki's 64-workload/24-probe superset was explicitly not
adopted as dishonest for a grammar with one parameter and no loops.

---

## Minimization + causal trace

| Option | Description | Selected |
|--------|-------------|----------|
| Source-text delta debugging (ddmin / C-Reduce style) | Well-known, general | |
| Deterministic structural reduction over the typed core with a fixed five-move alphabet, source reported as a projection of the reduced core | Respects the "never source offsets" identity discipline; candidates are valid by construction | ✓ |

**Outcome:** D-05-23 through D-05-27.
**Notes:** Reduction moves on core structure; the source case is a pretty-printed
projection of the same reduction run, never independently re-reduced. The
interestingness predicate pins axis, first-diverging operation identity, and
engine pair — "still fails" is explicitly rejected as the oracle, per C-Reduce's
own documented UB-drift failure, with the foreign boundary named as this
language's analogue of the UB trap. `bugpoint`'s pass-oblivious genericity was the
cited anti-pattern for not building a general engine. Three separate mutation
kills are required because a reducer has three distinct ways to be vacuous
(no-op, predicate too loose, non-deterministic).

---

## QLT-01 spike control audit

| Option | Description | Selected |
|--------|-------------|----------|
| Coverage-mapping document only | Cheap, human-readable | |
| Machine-readable control registry + executable audit + a *demonstrated* coordinated-lie escape | Matches the project's existing fail-closed duplicated-and-equality-tested precedent | ✓ |

**Outcome:** D-05-28 through D-05-31.
**Notes:** A mapping document alone is the DO-178C spreadsheet-RTM failure mode —
it decays silently and is caught at audit time, not commit time. Precedent cited
from LLVM's every-fixed-bug-gets-a-testcase rule, Rust's `//@ known-bug`
annotation (introduced specifically to replace a blanket `ignore-test` that hid
*why* a test was inert), and Chromium's disabled-test policy (bug link + OWNER
required, automated nagging past 180 days). "Not relevant" is admissible only as a
waived row with a falsifiable citation, owner, and phase. The coordinated
source-to-core lie must be **demonstrated** by a program that actually constructs
matching-but-false artifacts and passes as a named escape — asserting it in prose
is the weak version D-04-31 item 1 already has.

---

## Carried debt + `OpCall`

| Option | Description | Selected |
|--------|-------------|----------|
| `OpCall` out of Phase 5; retire `discoverLoanLastUses` here; fold the `Alias` pin into the foreign-emission plan | Matches the phase's evidence-shaped goal; satisfies D-04-26's precondition for the first time | ✓ |
| Pull `OpCall` into Phase 5 alongside everything else | Interprocedural corpus; closes D-03-02 outright | |

**Outcome:** D-05-32 through D-05-37.
**Notes:** Pulling `OpCall` in would land a new `OperationKind` at six dispatch
sites simultaneously with alias-fact emission, the first sanitizer lanes, the
first reducer, and the QLT-01 registry — the exact fingerprint of the failure that
cost Phase 2 a remediation round, Phase 3 a mid-phase gate plus three gap-closure
plans, and Phase 4 thirteen plans and five review rounds. It would also violate
D-04-26's own precondition. The auditor's strongest counter-argument was recorded
and answered rather than dismissed: SC1's `-O3` claim *is* thereby scoped to
intraprocedural programs, and the roadmap must say so explicitly, which is why
D-05-33 makes `OpCall` M002's lead charter item rather than leaving it unowned.
The Rust NLL migration (both borrow checkers in parallel migration mode for years,
old one deleted on convergence data rather than on a schedule) and GitHub's
Scientist supplied the retirement procedure in D-05-35. Phase 5 was identified as
the **last eligible slot in M001** for that retirement, since Phase 6 has no
language-surface work.

---

## Claude's Discretion

Plan decomposition, wave structure and plan count (subject to the pre-adopted
mid-phase gate and the early sequencing of the liveness retirement); the
enumeration bound constant; lane and control identifier spellings beyond those
named; the registry file format and location; `lang.mismatch/0` field encoding;
second-allocator symbol names; and the concrete shape of the by-pointer lowering.

## Deferred Ideas

`OpCall` and interprocedural equivalence (M002's lead charter item); a real
callback-registration seam; LSan/MSan/TSan; UBSan `implicit-conversion` and
`unsigned-integer-overflow`; an `-O3` sanitizer tier; the wiki's 64-workload
expansion and 24 blocking probes; payload-carrying alternatives; full code ports
of superseded spike fixtures; loops, generics, closures, threads, async.

## Scope creep redirected

None — the user selected only areas already inside the roadmap's Phase 5
boundary. The one genuine scope question raised (`OpCall`, deferred *to* Phase 5
by name in 04-CONTEXT.md but absent from the roadmap's Phase 5 goal and all four
success criteria) was adjudicated out and given an explicit home rather than left
to silence.
