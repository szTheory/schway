---
phase: 05-native-equivalence-and-adversarial-evidence
recorded: 2026-09-06
code_head: 2383565
status: accepted
disposition: carried-from-mid-phase-gate-and-phase-close
items: 2
blocking: 0
---

# Phase 05: Mid-Phase Gate Debt Register

Recorded by 05-09's mandatory mid-phase gate (D-05-40), following the Phase 3
and Phase 4 precedent of writing discovered debt down as a dated register
entry rather than leaving it in prose. Waves 1-3 are otherwise clean:
`sh scripts/verify-phase5.sh` is green, every Phase 5 control fires with
nonzero `RecomputedWork`, the six/one NAT-03 arithmetic holds, `discoverLoanLastUses`
is confirmed retired, and `git diff --stat` over every frozen prior-phase
corpus file, frozen foreign TU, and prior gate script reports no changes.

## Items

| ID | Source | Threat/Req | Severity | Landing phase | Item |
|---|---|---|---|---|---|
| D-05-41 | 05-09 `VerifyPhase5ControlsAndWork` (`lane:interpreter-o0-o3-lto`) | D-05-19 | info | P05 | The runtime gate's own `control:interpreter-o0-o3-lto` lane drives the interpreter/-O0/-O3/-O3-LTO differential over exactly one adversarial fixture (`inline_across_foreign.lang`), not the full six-fixture adversarial subset D-05-18a defines |
| D-05-42 | 05-14 (Task 2 developer-confirmed decision, D-05-32/D-05-33) | OWN-03/OWN-04, D-03-02 | warning | UNOWNED(m002-broad-scope) | `OpCall`, interprocedural loan liveness in both admission layers, call-graph construction, cycle refusal, a bounded interpreter call stack, and the cross-function rebuild of Phase 3's exhaustive differentials are OUT of M001; D-03-02 remains open past the milestone on this basis |

## Detail

### D-05-41 — the gate's own LTO lane samples one adversarial fixture, not all six

first-recorded: M001

**Finding.** `VerifyPhase5ControlsAndWork`'s `phase5RunInterpreterO0O3LTOLane` (`internal/compiler/session/session_phase5.go`) proves `control:interpreter-o0-o3-lto` by running `testdata/phase5/inline_across_foreign.lang` through the interpreter and three native tiers (`-O0`, `-O3`, `-O3 -flto`) and comparing all four via `Phase5CompareEngines`. It does not also drive `dead_store_unused_acquire.lang`, `reorder_two_events.lang`, `tail_collapse_release_ladder.lang`, `typed_failure_truncated_stdout.lang`, or `defect_dies_by_signal.lang` through the same four-tier comparison inside this one gate function.

**Why this is not a correctness gap.** The full six-fixture adversarial subset, plus the enumerated closure, is already proven to agree across the interpreter, `-O0`, and `-O3` by `TestPhase5CorpusThreeEngineAgreement` (`session_phase5_corpus_test.go`, part of the ordinary `go test ./...`/`go test -race ./...` this gate also runs). Plan 05-06's own `TestLTOTierIsNotInert` independently proves the LTO tier is non-inert (distinct codegen between `-flto` and non-`-flto` builds) using its own fixture. `VerifyPhase5ControlsAndWork`'s own lane exists to prove `control:interpreter-o0-o3-lto` fires with real, non-zero work inside the gate's own control-and-work accounting — not to re-run the whole corpus at every tier, which would make `sh scripts/verify-phase5.sh`'s own runtime scale with the corpus size for a control the exhaustive `go test` suite already covers.

**What is real debt.** The one-fixture sample does not exercise `-flto`'s cross-TU inlining behavior against the other five adversarial shapes (dead-store elimination, event reordering, tail-collapse, truncated-stdout, and signal-terminated defect) inside the gate's own control-and-work JSON. A future toolchain regression that broke LTO agreement specifically for one of those five shapes (and not `inline_across_foreign.lang`) would still be caught by `TestPhase5CorpusThreeEngineAgreement` at `-O0`/`-O3`, but NOT by this gate's own `-O3-LTO` tier, since that test does not itself add an LTO tier.

**Phase 5+ fix.** Either (a) widen `phase5RunInterpreterO0O3LTOLane` to iterate `session.Phase5AdversarialFixtureFiles` at the `-O3-LTO` tier (cost: five more native builds and links per gate run), or (b) add an LTO tier to `TestPhase5CorpusThreeEngineAgreement`'s own adversarial-subset loop (moving the cost into the ordinary `go test` suite instead of the gate script's own runtime). Either closes the gap; neither is required before plans 05-10 through 05-14 begin, since both consume the settled interpreter/-O0/-O3 differential and the alias-fact/attribute-justification work this gate proves, not the LTO tier specifically.

### D-05-42 — `OpCall` and interprocedural equivalence deferred to M002; D-03-02 re-recorded as open (D-05-32/D-05-33)

first-recorded: M001

**Decision (developer-confirmed, 05-14 Task 2, `defer-to-m002`).** `OpCall`
(Lang-to-Lang calls), interprocedural loan liveness in BOTH admission layers,
call-graph construction, cycle refusal, a bounded interpreter call stack, and
the cross-function rebuild of Phase 3's exhaustive differentials are OUT of
M001 and become **M002's LEAD charter item** (see `.planning/ROADMAP.md`
§"M002 Charter (deferred from Phase 5, D-05-32/D-05-33)"). This decision is
rated **one-way for this milestone**: once M001 ships, it is not reversible
within it (D-05-32).

**Verbatim carry-forward (D-05-33's required wording):**

> D-03-02 remains open, unreachable-but-unfixed, lift condition unchanged
> (callable ⊆ publishable, D-04-03). M001 ships without Lang-to-Lang calls.
> Interprocedural equivalence is explicitly outside M001's proof scope.

**Accepted consequences, stated plainly (not softened):**

1. **M001 ships WITHOUT Lang-to-Lang calls.** No `OpCall` operation kind exists
   anywhere in the shipped M001 IR, checker, validator, interpreter, or cgen.
2. **Interprocedural `-O3` equivalence is outside M001's proof scope BY
   CONSTRUCTION**, since M001 ships without calls. This is not a gap in SC1's
   wording — it is a direct structural consequence of this deferral, and no
   phase in the current M001 roadmap closes it.
3. **D-03-02 remains OPEN past the milestone** (unreachable-but-unfixed; lift
   condition unchanged per D-04-03: callable ⊆ publishable). `03-DEBT.md`'s own
   historical closure record (`03-09`, 2026-09-04) is Phase 3's authoritative
   record of the SINGLE-FUNCTION fix that shipped (`originvalidate.ValidatePublished`
   refuses an undeclared borrow-derived exported return via `core.origin_omitted`)
   — that fix stands and is not reverted. What re-opens here is the INTERPROCEDURAL
   half of the same hazard class: with no `OpCall`, the callable-surface
   consequence of an omitted origin (a caller silently treating a borrow as an
   owned value across a call) is unreachable today. It becomes newly reachable
   the moment M002 lands `OpCall`, and D-04-03's lift condition (callable ⊆
   publishable) is the gate that must hold from that day one.

**Why not Phase 6.** Phase 6 is agent feedback and performance ratification,
not language surface — pulling `OpCall` there would be exactly the unowned,
undisciplined carry D-05-33 exists to prevent. M002 is where this work is
scheduled and owned.

**Not a widening of D-05-41.** D-05-41 (the gate's own LTO lane sampling one
of six adversarial fixtures) is unrelated in cause and scope to this entry
and is not touched by this decision.

## Named residual limitations (accepted, Phase 5 close)

Recorded as accepted residuals rather than open work — demonstrating an escape
or measuring a boundary is real evidence, but it does not close the
underlying hazard:

1. **The callback-invocation mechanism is unsubjected**
   (`escape:callback-invocation-unsubjected`) — declared, gate-visible, and
   asserted never-detected per D-05-07/D-05-30.
2. **The coordinated multi-artifact lie survives, one layer wider.** Phase 5
   makes the coordinated source-to-core false claim demonstrable
   (`escape:coordinated-source-to-core-false-claim`, plan 05-13) — strictly
   better evidence than the prose-only record it replaces — but demonstrating
   an escape is not closing it.
3. **Optimizer observability is host- and version-bound.** The `-O0`/`-O3`
   divergence fixtures were measured on Apple clang 21 / arm64; the compiler
   version is recorded in evidence and a lost divergence fails the lane, but
   the fixture's SHAPE remains an engineered artifact of this toolchain, not a
   portable law.
4. **A differential is existential, not universal.** Passing the milestone
   corpus proves agreement on the fixtures exercised, never over the full
   program space.
5. **Quarantine of the foreign boundary is permanent and non-discharging.**
   Spike 005 iteration 5 recorded a real cleanup bypass (`longjmp`) that
   produced NO sanitizer report; FFI-01's gate makes foreign obligations
   inspectable and their violations detectable, never a proof the foreign
   implementation obeys them.
6. **UBSan's `implicit-conversion` and `unsigned-integer-overflow` checks are
   a named escape, not a silent drop.** An `-O3`-level sanitizer tier is a
   named deferred obligation.

D-05-37's explicit criterion is also recorded here rather than left an assumed
corollary: **no prior phase proved that `-O3` code generation preserves the
CFG-precise last-use loan expiry semantics `check.go` gates on** — Phase 5's
native differential (`control:interpreter-o0-o3-lto` plus the full adversarial
subset's three-engine agreement) is the first to do so.

## Closure — carried Phase 3/4 debt closed by Phase 5

- **D-04-26 is CLOSED by plan 05-02.** `discoverLoanLastUses` was retired via
  the Rust-NLL-style shadow-mode migration D-05-34/D-05-35 required: both laws
  ran in shadow mode over `TestOwnershipSequenceExhaustive`'s ~113,164-case
  enumeration and `TestBranchSequenceExhaustive`'s per-arm equivalent with zero
  logged divergences, then `discoverLoanLastUses` was deleted and
  `loanLivenessFixpoint`/`core.LoanEndpoint` promoted to load-bearing across
  both `checkLinear`/`analyzeStraightLine` and `checkBranch`/`analyzeArmBody`.
  This closure was possible for the first time since D-04-26 was written only
  because D-05-32 kept `OpCall` out of Phase 5 (D-04-26's own stated
  precondition — "`check.go` is not itself landing a new `OperationKind`" —
  finally held).
- **D-03-01 is CLOSED by plan 05-02**, alongside D-04-26 (same shadow-mode
  retirement; D-03-01's metric-honesty half was already closed at D-04-26's
  recording, its structural half — the quadratic scanner remaining
  admission-deciding — is what this retirement removes).
- **D-04-33 is CLOSED by plan 05-01.** `core.ForeignContract.Alias` was folded
  into `corevalidate.foreignContractFieldsCSafe`/`validCIdentifier` and
  `cgen.unsafeForeignContractField` alongside its siblings, with a regression
  test asserting no `EmitForeign*` output ever contains an `Alias` value
  (D-05-36).
