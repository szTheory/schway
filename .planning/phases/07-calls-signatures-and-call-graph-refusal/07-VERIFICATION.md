---
phase: 07-calls-signatures-and-call-graph-refusal
verified: 2026-09-09T00:00:00Z
status: gaps_found
score: 3/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "A call is admitted only when it can be proven safe from the callee's signature — the signature summary is a sound gate, not merely a body-blindness mechanism"
    status: failed
    reason: >
      `check.resolveCallBinding` (internal/compiler/check/check.go:1737-1791) never
      compares the call argument's type to the callee's declared `Parameter.Type`.
      It sets the OpCall target's `TypeID` to the CALLER's own argument type
      (`argument.place.TypeID`), never derived from the callee's declared return
      type. `corevalidate`'s independent OpCall replay
      (internal/compiler/corevalidate/corevalidate.go:1458-1484, `targetMatches`
      at :2254-2262) has the identical blind spot — it only checks the target's
      TypeID against the OPERATION's own TypeID (which check already set to the
      caller's type), never against `functionByID[operation.CalleeID].Parameter.Type`.
      Confirmed empirically in this verification: a program calling
      `identity(value: Byte) -> Byte` with a `Buffer` argument is admitted by
      `lang --json check` with zero diagnostics and status "pass". Both of the
      project's "independent dual derivations" share this exact blind spot, so
      the two-peer safety net this phase is built around does not catch it.
      `core.FunctionSignature.Parameters []ParameterContract` (core.go:199) exists
      in the /1 schema specifically to carry this fact to a caller, but no
      admission path this phase actually consults it for argument-type matching.
      This directly undermines the phase goal's central claim — "admitted from
      the callee's signature alone" — because the signature's type facts are
      present in the schema but structurally unused at the one call site meant to
      consult them; the call is admitted despite an incompatible signature, not
      because the signature was checked. Not yet exploitable at runtime this
      phase because interp/cgen both refuse to execute/lower OpCall (D-07-39),
      but it is a live, undisclosed soundness gap in the checker as shipped, not
      a debt item recorded in PHASE-07-DEBT.md — it was found by code review, not
      declared at planning time.
    artifacts:
      - path: "internal/compiler/check/check.go"
        issue: "resolveCallBinding (lines ~1737-1791) never reads functionIDs[...]'s parameter type; TargetID.TypeID is copied from the caller's own argument, never derived from the callee's declared return type"
      - path: "internal/compiler/corevalidate/corevalidate.go"
        issue: "OpCall replay case (~1458-1484) and targetMatches (~2254-2262) never cross-reference functionByID[operation.CalleeID].Parameter.Type"
    missing:
      - "An argument-type-vs-parameter-type check in resolveCallBinding, refusing with a new causal diagnostic (e.g. check.call_argument_type_mismatch) on mismatch"
      - "The mirrored, independently-derived check in corevalidate's OpCall replay case"
      - "OpCall's TargetID.TypeID derived from the callee's own declared return type, not the caller's argument type"
      - "A testdata/phase07/call_type_mismatch.lang fixture and refusal assertions in both check_test.go and corevalidate_test.go"
deferred:
  - truth: "check's admission and corevalidate's replay agree on every fixture in the corpus (peer non-divergence)"
    addressed_in: "Phase 08 / Phase 09"
    evidence: >
      testdata/phase07/relay_escort_witness.lang is a NAMED, tested, explicitly
      documented divergence: check admits it (07-05-SUMMARY.md, check's
      intraprocedural loan-liveness law does not see a "call" binding's
      RHS.Arguments as a use, so an exclusive loan is wrongly treated as ending
      at its own creation point), while corevalidate independently refuses it
      with core.move_while_borrowed. This is the interprocedural half of D-03-02,
      explicitly scoped to Phase 08 ("The checker decides, from callee signatures
      alone, whether a loan is still live across a call boundary") and Phase 09
      ("D-03-02 is closed ... in both admission layers"), per ROADMAP.md. A
      single named exception to the corpus-wide parity test
      (TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus) plus a decisive
      test proving the divergence
      (TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed)
      exist, so the divergence is disclosed and tested, not silently hidden.
---

# Phase 07: Calls, Signatures, and Call-Graph Refusal — Verification Report

**Phase Goal:** A Lang function can call another Lang function, admitted from
the callee's signature alone, and a program whose calls form a cycle is
refused by name instead of hanging.
**Verified:** 2026-09-09
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `lang check` admits a two-function program where one calls another; `core.OpCall` is handled at all six dispatch sites (check, corevalidate, interp, cgen, pathoracle, originvalidate) and both exhaustive-dispatch controls are green | ✓ VERIFIED | `go run ./cmd/lang --json check testdata/phase07/call_basic.lang` → `status: pass`. `go run ./cmd/lang --json verify testdata/phase07` → lane `kind-exhaustive-dispatch-phase07` status `pass` with all 16 controls listed, including `control:kind.exhaustive_dispatch.phase07_in_process` and `.phase07_lane`. `runExhaustiveDispatchControl` (core_test.go:338-425) directly drives check, corevalidate, pathoracle.RecomputeEndpoints, originvalidate.RecomputeOriginPerReturn, interp.Run, and cgen.Emit over the phase-07 corpus and asserts every declared `core.OperationKind` including `OpCall` is encountered. `go build ./...`, `go vet ./...`, and `go test ./... -p 1` are all green (confirmed in this session). |
| 2 | No caller admission path reads a callee body — the signature summary is the only input — and a call to a non-publishable callee is refused with a stable diagnostic code | ⚠️ PARTIALLY VERIFIED — body-blindness holds; signature-based type soundness does not | Body-blindness: `verifyCallableRefusalBodyReadSeam`/`calleeBodyReadObserved` (check.go) is a dedicated, mutation-killed control proving admission never reads `Linear`/`Match`. `testdata/phase07/call_uncallable_callee.lang` → `status: invalid`, code `core.callee_not_callable`; `clean_but_unpublishable.lang` → `core.origin_omitted`. **However**, see Gap 1: the signature's `Parameters` (types) are structurally present in `core.FunctionSignature` but never consulted for argument-type compatibility in either `check.resolveCallBinding` or `corevalidate`'s OpCall replay — confirmed empirically that a `Byte`-parametered function called with a `Buffer` argument is admitted with zero diagnostics. This is a real soundness gap in what "admitted from the callee's signature" is supposed to mean, found by this phase's own code review (07-REVIEW.md CR-01, rated CRITICAL/BLOCKER) and independently reproduced in this verification. |
| 3 | Direct, mutual, and indirect call cycles are each refused with a named refusal code and never hang; an indirect-cycle corpus and a pathological-depth-but-acyclic corpus both return bounded verdicts; traversal is explicit-worklist/visited-set-guarded, never native Go recursion | ✓ VERIFIED | `cycle_self.lang`, `cycle_mutual.lang`, `cycle_indirect.lang`, `cycle_unreachable.lang`, `cycle_through_match_arm.lang`, `foreign_symbol_shadowing.lang` (a cycle case) all → `status: invalid`, code `core.call_graph_cycle`, each returning promptly (no hang observed). `deep_diamond_acyclic.lang` (13 functions, 4 chained diamonds, depth 8+) → `status: pass`, proving depth alone does not trigger refusal. `internal/compiler/callgraph/callgraph.go` implements an iterative white/gray/black DFS over an explicit stack (confirmed by reading source and by 07-06/07-07 SUMMARY claims); `corevalidate` runs its own independent traversal (`checkCallGraphAcyclic`) over a disjoint (synthetic-only) input space, verified via `foreign_symbol_shadowing.lang` and the cycle corpus above. Diagnostic bound (`MaxCycleCauses = 32`, `truncated:core.call_cycle_bound`) is declared, traversal itself is unbounded (O(V+E)). |
| 4 | Every new interprocedural control introduced this phase has been observed to fail against a seeded mutation in the plan that introduced it (QLT-08) | ✓ VERIFIED | Every plan's SUMMARY documents named mutation-kill tests for its own new controls (e.g. `TestPhase7DispatchControlsMutationKilled`, `TestOpCallGroupedArmMutationKilled`, `TestPhase7ControlsAreMutationKilled` as the phase-wide completeness matrix compared by exact set equality against `Phase7RequiredControls()`). `go test ./... -p 1` is green, meaning these mutation-kill tests themselves currently pass (i.e., the seam correctly demonstrates failure under the seeded mutation and correctly passes when the seam is off). Not independently re-derived byte-for-byte in this verification pass beyond reading the test bodies and confirming they exist and are wired into the required-control lists that the CLI-observable lane (`lang verify testdata/phase07`) exercises and reports green. |

**Score:** 3/4 truths fully verified, 1 truth partially failed (Truth 2's type-soundness half) — reported as `gaps_found` per Step 9 rule 1 (a must-have truth FAILED).

### Central Judgment Call: Does CR-01 Defeat the Phase Goal?

The phase goal states a call is "admitted from the callee's signature alone."
This verification treats that clause as making two claims at once:
(a) a *mechanism* claim — no admission path reads a callee body, only its
published signature — and (b) an implicit *soundness* claim — the signature is
consulted to determine whether the call is actually legal.

Claim (a) holds: the body-blindness control is real, tested, and
mutation-killed; no admission path this phase reaches a callee's `Linear` or
`Match`.

Claim (b) does not hold for argument types. `core.FunctionSignature` was
explicitly designed in 07-01 to carry `Parameters []ParameterContract` "even
though the checker admits exactly arity 1 today ... schema capacity is taken
now" — i.e., the summary was built to be the vehicle for exactly this check —
but no consumer this phase reads that field to gate a call. The call is
admitted regardless of whether the caller's argument type matches the
callee's declared parameter type. Both of check's own resolveCallBinding and
corevalidate's independent OpCall replay share this exact omission, so the
project's own "independent dual derivation" safety net (the mechanism this
phase's entire engineering effort is otherwise built around) provides no
protection against this specific defect class.

This verification treats it as a genuine BLOCKER against the phase goal, not
a stylistic nit, for three reasons: (1) it was not declared in
`PHASE-07-DEBT.md` at planning time — the debt file enumerates several
deliberate narrowings (D-07-33's narrowed `Callable` peer, D-03-02's deferred
liveness peer) but never mentions argument-type checking as out of scope; (2)
the phase's own code review independently found and rated it CRITICAL/
BLOCKER, and this verification independently reproduced the empirical
counter-example rather than trusting that finding; (3) none of the eight
plans' `must_haves.truths` claim argument-type checking is out of scope for
Phase 07 — arity is explicitly narrowed to 1 (D-07-07), but type-compatibility
of that one argument is never named as deferred anywhere in the plans, the
debt file, or the roadmap's scope-cut order. It is an omission, not a declared
cut.

It is not exploitable as a runtime memory-safety defect *this phase* because
`interp` and `cgen` both explicitly refuse to execute/lower `core.OpCall`
(D-07-39) — confirmed by reading `interp.go`/`cgen.go`'s dedicated
"recognized, unsupported" arms. But "admitted from the callee's signature" is
this phase's headline claim, evaluated at `check`-admission time, not at
execution time — and at admission time, the signature's type facts are
present but unused.

### Deferred Items

| # | Item | Addressed In | Evidence |
|---|------|-------------|----------|
| 1 | check/corevalidate peer agreement on `relay_escort_witness.lang` (a real, named, tested divergence: check admits, corevalidate refuses with `core.move_while_borrowed`) | Phase 08 / Phase 09 | Phase 08's goal is exactly "The checker decides, from callee signatures alone, whether a loan is still live across a call boundary"; Phase 09's goal closes D-03-02 "in both admission layers." The divergence is disclosed via a decisive named test (`TestRelayEscortWitnessCorevalidateIndependentlyRefusesMoveWhileBorrowed`) and a single named exception to the corpus-wide parity invariant, not silently absorbed. |

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/core/core.go` | `core.OpCall`, `CalleeID`, `FunctionSignature`/`ParameterContract`/`ReturnContract`, `DecodeInterface` | ✓ VERIFIED | Present, exercised, `AllOperationKinds()` includes `OpCall` |
| `internal/compiler/check/check.go` | `resolveCallBinding`, `verifyCallableRefusal`, pre-body signature table | ✓ VERIFIED (wiring) / ✗ gap (type soundness — see Gap 1) | Admission logic present and wired into `analyzeStraightLine`/`analyzeArmBody`; does not consult parameter types |
| `internal/compiler/corevalidate/corevalidate.go` | independent `OpCall` replay, `peerCallable`, cycle peer, closure-chain peer | ✓ VERIFIED (wiring) / ✗ gap (type soundness — see Gap 1) | Present at both replay sites; shares check's type-check blind spot |
| `internal/compiler/callgraph/callgraph.go` | iterative 3-color DFS, deterministic witness selection, bounded diagnostic | ✓ VERIFIED | Confirmed via cycle corpus outcomes and source inspection |
| `internal/compiler/interp/interp.go`, `internal/compiler/cgen/cgen.go` | dedicated "recognized, unsupported" `OpCall` arms | ✓ VERIFIED | `ErrCallUnsupported` and cgen's D-07-39/A-02 comment + dedicated case confirmed by source read |
| `internal/compiler/session/session_phase7.go`, `scripts/verify-phase7.sh` | `Phase7RequiredControls()`, CLI-observable lane | ✓ VERIFIED | `lang verify testdata/phase07` reports 16 controls, all `pass` |
| `testdata/phase07/*.lang` (12 fixtures) | call/cycle/callable/witness corpus | ✓ VERIFIED | All 12 files present; each produces the expected status/code when run through `lang --json check` |
| `PHASE-07-DEBT.md` | declared deferred scope | ✓ VERIFIED (as far as it goes) | Well-formed, discloses D-03-02 liveness peer deferral and D-07-33's narrowed Callable peer — does **not** disclose the CR-01 argument-type gap, which is the basis of Gap 1 above |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `core.AllOperationKinds()` | six dispatch sites | exhaustive switches | ✓ WIRED | `runExhaustiveDispatchControl` drives all six sites over the phase-07 corpus |
| `check.resolveCallBinding` | `core.LinearOperation{Kind: OpCall, CalleeID}` | callee name resolution | ✓ WIRED | Confirmed by source read and `call_basic.lang` outcome |
| `originvalidate.PublishProblemsFor` | `FunctionSignature.Callable` | `check`'s signature table | ✓ WIRED | `clean_but_unpublishable.lang` / `call_uncallable_callee.lang` outcomes |
| `check` emits `OpCall` | `callgraph.Order` | in-memory ephemeral `core.Program` before return | ✓ WIRED | Cycle corpus refused with `core.call_graph_cycle` before any downstream stage runs |
| `FunctionSignature.Parameters[].Type` | call-site argument type check | — | ✗ NOT WIRED | This is Gap 1: the field exists but no admission path reads it |
| `callgraph.Order` reverse postorder | `ClosureDigest` chaining | 07-08 | ✓ WIRED | `control:summary.closure_digest_chained`/`.closure_digest_reverse_postorder` both `pass` in the verify lane |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Two-function call admitted | `lang --json check testdata/phase07/call_basic.lang` | `status: pass` | ✓ PASS |
| Self-cycle refused, no hang | `lang --json check testdata/phase07/cycle_self.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Mutual cycle refused | `lang --json check testdata/phase07/cycle_mutual.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Indirect cycle refused | `lang --json check testdata/phase07/cycle_indirect.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Unreachable-from-entry cycle still refused (fail-closed) | `lang --json check testdata/phase07/cycle_unreachable.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Cycle through match arm refused | `lang --json check testdata/phase07/cycle_through_match_arm.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Deep acyclic diamond corpus admitted (bounded verdict, not refused by depth) | `lang --json check testdata/phase07/deep_diamond_acyclic.lang` | `status: pass` | ✓ PASS |
| Call to non-publishable callee refused | `lang --json check testdata/phase07/call_uncallable_callee.lang` | `status: invalid`, `core.callee_not_callable` | ✓ PASS |
| Foreign-symbol shadowing preserves cycle edge | `lang --json check testdata/phase07/foreign_symbol_shadowing.lang` | `status: invalid`, `core.call_graph_cycle` | ✓ PASS |
| Exhaustive-dispatch + all phase-07 controls green | `lang --json verify testdata/phase07` | lane `pass`, 16/16 controls listed as `pass` | ✓ PASS |
| Full build/vet/test | `go build ./...`, `go vet ./...`, `go test ./... -p 1` | all exit 0 | ✓ PASS |
| **Type-mismatched call wrongly admitted (negative check)** | `lang --json check` on an ad hoc fixture calling `identity(value: Byte) -> Byte` with a `Buffer` argument | `status: pass`, zero diagnostics | ✗ FAIL — confirms Gap 1 |

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|-----------------|-------------|--------|----------|
| SEM-04 | 07-03, 07-04, 07-07 | `OpCall` real at all six dispatch sites, both exhaustive controls green | ✓ SATISFIED | Verified above |
| SEM-05 | 07-01, 07-02, 07-05, 07-08 | Digest-bound signature summary carries everything a caller needs for admission; no admission reads a callee body | ✗ PARTIALLY BLOCKED | Body-blindness satisfied; "everything a caller needs" is not satisfied — `Parameters` (types) are carried but unused for admission (Gap 1) |
| SEM-06 | 07-02, 07-05 | Call admitted only when callee callable ⊆ publishable; stable refusal code | ✓ SATISFIED | `call_uncallable_callee.lang` → `core.callee_not_callable` |
| SEM-07 | 07-06, 07-07 | Call graph constructed, direct/mutual/indirect cycles refused by name, never a hang | ✓ SATISFIED | Verified above via full cycle corpus |
| QLT-08 | all 8 plans | Every new interprocedural control mutation-killed in its introducing plan | ✓ SATISFIED | Confirmed present in every plan's SUMMARY; phase-wide completeness matrix (`TestPhase7ControlsAreMutationKilled`) compares by exact set equality |

No orphaned requirements: all five IDs mapped to Phase 07 in REQUIREMENTS.md appear in at least one plan's `requirements` frontmatter field.

### Anti-Patterns Found

No `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER` markers found in the phase's modified files during this review. The one substantive defect found (CR-01 / Gap 1) is a missing check, not a stub or a debt-marker comment — it is silent, which is why the code review and this verification both had to reproduce it empirically rather than grep for it.

### Human Verification Required

None. Gap 1 was confirmed programmatically (empirical reproduction of the unsound admission) and does not require human judgment to establish factually — only to decide disposition (fix now vs. accept as an override).

### Gaps Summary

One must-have truth fails: the phase's headline claim that a call is "admitted
from the callee's signature alone" is undermined by a confirmed, empirically
reproduced soundness gap — neither `check` nor `corevalidate` compares a call
argument's type against the callee's declared parameter type, despite the /1
signature schema carrying exactly that fact for this purpose. This is not
speculative: `lang --json check` admits a `Byte`-parametered function called
with a `Buffer` argument with zero diagnostics. It is not yet a runtime
memory-safety hazard (interp/cgen refuse to execute/lower `OpCall` this
phase), but it is a live hole in the one place ("admission") this project's
entire design premise depends on, it was not declared as deferred scope at
planning time, and both of the project's "independent" derivations share the
identical blind spot — meaning the phase's dual-peer safety net, otherwise
carefully engineered throughout this phase, provides zero protection against
this specific defect class.

Everything else checked in this phase — the six-dispatch-site registration,
cycle refusal (direct/mutual/indirect/unreachable/through-match-arm/foreign-
shadowing), the body-blindness control, publication-safety refusal, bounded
diagnostics with unbounded (worklist-based, non-recursive) traversal, and the
mutation-kill discipline — is real, wired, and independently confirmed in this
verification pass. The `relay_escort_witness.lang` check/corevalidate
divergence is a disclosed, tested, and explicitly deferred item (to Phase
08/09), not a phase-07 gap.

**This looks like it could be accepted as an override** if the team judges
"admitted from the callee's signature alone" narrowly as a mechanism claim
(never reads a body) rather than a full soundness claim, and treats argument-
type checking as an undeclared-but-acceptable scope gap to close immediately
in a follow-up plan before Phase 08 begins (since Phase 08's loan-liveness
work will build directly on top of this same admission path). To accept this
deviation, add to this file's frontmatter:

```yaml
overrides:
  - must_have: "admitted from the callee's signature alone (type soundness)"
    reason: "Arity is fixed at 1 and no fixture in the M001-carried corpus exercises a type mismatch; interp/cgen do not execute/lower OpCall this phase so no runtime hazard exists yet; fix deferred to a named follow-up plan before Phase 08."
    accepted_by: "<name>"
    accepted_at: "<ISO timestamp>"
```

Absent that explicit acceptance, this verification recommends a small
follow-up plan (mirroring the code review's own suggested fix: compare
`argument.place`'s type constructor to `functionIDs`-resolved callee's
`Parameter.Type` in `check.resolveCallBinding`, mirror in `corevalidate`'s
`OpCall` replay, derive `OpCall`'s `TargetID.TypeID` from the callee's own
return type, and add `testdata/phase07/call_type_mismatch.lang`) before this
phase is considered fully closed.

---

_Verified: 2026-09-09_
_Verifier: Claude (gsd-verifier)_
