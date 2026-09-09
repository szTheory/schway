---
phase: 07-calls-signatures-and-call-graph-refusal
verified: 2026-09-09T00:00:00Z
status: passed
score: 4/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 3/4
  gaps_closed:
    - "A call is admitted only when it can be proven safe from the callee's signature — the signature summary is a sound gate, not merely a body-blindness mechanism"
  gaps_remaining: []
  regressions: []
---

# Phase 07: Calls, Signatures, and Call-Graph Refusal — Verification Report

**Phase Goal:** A Lang function can call another Lang function, admitted from
the callee's signature alone, and a program whose calls form a cycle is
refused by name instead of hanging.
**Verified:** 2026-09-09
**Status:** passed
**Re-verification:** Yes — after gap-closure plan 07-09

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | `lang check` admits a two-function program where one calls another; `core.OpCall` is handled at all six dispatch sites and both exhaustive-dispatch controls are green | ✓ VERIFIED (regression check) | `go run ./cmd/lang --json check testdata/phase07/call_basic.lang` → `status: pass`. Full `lang verify testdata/phase07` lane `kind-exhaustive-dispatch-phase07` → `status: pass`, 18 controls listed (16 prior + 2 new from 07-09), all `pass`. |
| 2 | No caller admission path reads a callee body — the signature summary is the only input — **and it is a sound gate**: a call whose argument type does not match the callee's declared parameter type is refused, not admitted | ✓ VERIFIED (gap closed) | Independently re-derived: `go run ./cmd/lang --json check testdata/phase07/call_type_mismatch.lang` → `status: invalid`, code `check.call_argument_type_mismatch`, causes `callee`/`argument_type=Buffer`/`declared_parameter_type=Byte`, exactly as ratified. Read `check.go:1852-1905` directly: the gate compares `typeFact.Shape.Constructor` (caller) against `calleeContracts[...].ParameterType` (callee's declared parameter type, from a pre-body AST-derived table) before ever constructing `core.LinearOperation`. Body-blindness still holds — the new `calleeContract` table is built from `ast.FuncDecl.Parameter.Type`/`ReturnType` only, never a body. `TargetID.TypeID` (and the operation's own `TypeID`) is now `typeFact.ID` resolved against the callee's declared `ReturnType`, not `argument.place.TypeID` (`grep -c 'TypeID: argument.place.TypeID' check.go` = 0 in the production path). `corevalidate`'s independent peer (`checkCallTypeContract`, `corevalidate.go`) re-derives the same two facts from `places`/type-facts/`functionByID` — confirmed by reading the source and by `grep -rn 'compiler/check' internal/compiler/corevalidate/` returning zero real imports (only forbidden-import-list *strings* in tests). All named tests (`TestCallArgumentTypeMismatchRefused`, `TestCallTargetTypeDerivedFromCalleeReturn`, `TestOpCallTargetTypeIDUnchangedAcrossAcceptingCorpus`, `TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus`, `TestPeerRefusesCallArgumentTypeMismatch`, `TestPeerRefusesCallReturnTypeMismatch`, `TestCallTypePeersIndependentOfCheck`) re-run in this session and pass. |
| 3 | Direct, mutual, and indirect call cycles are each refused with a named refusal code and never hang; depth alone does not trigger refusal; traversal is worklist-based, not native recursion | ✓ VERIFIED (regression check) | Re-ran the full cycle corpus in this session: `cycle_self`, `cycle_mutual`, `cycle_indirect`, `cycle_unreachable`, `cycle_through_match_arm`, `foreign_symbol_shadowing` → all `core.call_graph_cycle`, unchanged by 07-09. `deep_diamond_acyclic.lang` → `status: pass`, unchanged. |
| 4 | Every new interprocedural control introduced this phase has been observed to fail against a seeded mutation in the plan that introduced it (QLT-08), including the two new 07-09 controls | ✓ VERIFIED, with a disclosed judgment call (see below) | `control:call.argument_type_matches_parameter` and `control:call.target_type_from_callee_return` both appear in `session.Phase7RequiredControls()` (session_phase7.go:58-59), verbatim in `scripts/verify-phase7.sh` (lines 129-130), and in `controlsWithRecordedMutationKill` (session_phase7_test.go:165-166). `TestPhase7ControlsAreMutationKilled` and `TestPhase7RequiredControlsMatchScript` re-run green in the full suite. Each control's mutation-kill test (`TestCallArgumentTypeCheckMutationKilled`, `TestCallReturnTypeDerivationMutationKilled`, `TestCheckCallTypeContractArgumentPeerSeamKilled`, `TestCheckCallTypeContractReturnPeerSeamKilled`) re-run and pass. |

**Score:** 4/4 truths verified (up from 3/4 in the prior verification).

### Judgment Call: Does the direct-API mutation-kill technique satisfy QLT-08?

**The deviation.** 07-09-SUMMARY.md discloses (Rule-1, auto-fixed) that this
language's `sameType` invariant forces every function's declared return type
to equal its declared parameter type. Consequently, for any real `.lang`
fixture or any full-pipeline-constructible synthetic `core.Program`, the
argument-type comparison and the return-type comparison are the mathematically
identical boolean — disabling only one of the two new seams on a real fixture
still gets caught by the *other*, still-active gate, so neither seam can be
observed failing in isolation through an end-to-end fixture. The executor
instead proved each control's kill by calling the unexported production
predicate directly (`resolveCallBinding` in `check`; `checkCallTypeContract` in
`corevalidate`) with hand-built inputs that deliberately decouple the two
facts — a shape no real source program or full-pipeline synthetic can
construct.

**This verification's independent read of the tests** (`check_test.go:3348-3433`,
reproduced above) confirms the claim precisely: these tests call the actual
production function with the actual production seam variable
(`callArgumentTypeCheckSeam`, `callReturnTypeDerivationSeam`) — not a mock,
not a re-implementation — and observe the real gate change its verdict from
admit to refuse and back. The only artificial element is the *input*
(a `calleeContract` whose `ParameterType` and `ReturnType` differ, which the
language surface cannot express on any legally-typed callee), not the
predicate under test.

**Judgment: this satisfies QLT-08's substance, not merely its letter.** QLT-08
requires "observed to fail against a seeded mutation" — it does not mandate
that the seeding vehicle be an end-to-end `.lang` fixture, and the same plan's
own D-07-47 (the `check.call_return_type_unrepresentable` code) is already
disclosed as reachable only through a seeded seam and never through a fixture,
for the identical structural reason. Requiring fixture-only mutation kills here
would be requiring a mathematical impossibility given this language's `sameType`
rule, not a stronger form of evidence — it would not catch a bug that this
technique misses. This is a legitimate, disclosed, same-real-code
fault-injection technique, not a weakened control. It is recorded here rather
than passed over silently, per this task's instruction.

### Required Artifacts (07-09 delta only — see prior verification for the rest, all still holding)

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/core/core.go` | `CallArgumentTypeMismatch`, `CallReturnTypeMismatch` peer codes | ✓ VERIFIED | Present with doc comments distinguishing from `core.type_mismatch` |
| `internal/compiler/check/check.go` | `calleeContract`, `buildCalleeContracts`, argument-type gate, callee-return-derived `TargetID.TypeID` | ✓ VERIFIED | Read in full; matches must_haves exactly (see Truth 2 evidence) |
| `internal/compiler/corevalidate/corevalidate.go` | Independent `checkCallTypeContract` peer, no import of `check` | ✓ VERIFIED | `grep -rn 'compiler/check' internal/compiler/corevalidate/` → zero real imports |
| `testdata/phase07/call_type_mismatch.lang` | Standing negative-control fixture | ✓ VERIFIED | Exists, refused at both layers, verdict confirmed live in this session |
| `session_phase7.go`, `scripts/verify-phase7.sh`, `session_phase7_test.go` | Two new control identifiers wired into all three | ✓ VERIFIED | Confirmed via grep and a live `sh scripts/verify-phase7.sh` run in this session, both controls listed `pass` |
| `PHASE-07-DEBT.md` | Discloses retained limitation | ✓ VERIFIED | D-07-46 (nominal constructor-string comparison), D-07-47 (unreachable return-refusal code), D-07-48 (arity-N ordering unanswered) all present and well-formed |

### Key Link Verification (delta)

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| `FunctionSignature.Parameters[].Type` | call-site argument type check | `calleeContract.ParameterType` consulted in `resolveCallBinding` before `OpCall` construction | ✓ WIRED (was NOT WIRED) | `TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus` re-run and passes; source read confirms the gate fires before construction, not after |
| `check`-produced `core.Program` | `corevalidate` replay | independent re-derivation from `places`/type-facts/`functionByID`, no shared helper with `check` | ✓ WIRED, independence confirmed | Zero real cross-package imports; different input mechanism (AST-derived vs core-artifact-derived) per source read |

### Behavioral Spot-Checks (re-run live in this session)

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Type-mismatched call now refused (the prior FAIL) | `lang --json check testdata/phase07/call_type_mismatch.lang` | `status: invalid`, `check.call_argument_type_mismatch`, correct ordered causes | ✓ PASS (was ✗ FAIL) |
| Two-function call still admitted | `lang --json check testdata/phase07/call_basic.lang` | `status: pass` | ✓ PASS |
| All six cycle fixtures still refused unchanged | `lang --json check` on each | all `core.call_graph_cycle` | ✓ PASS |
| Deep acyclic diamond still admitted | `lang --json check testdata/phase07/deep_diamond_acyclic.lang` | `status: pass` | ✓ PASS |
| Callability/publication refusals unchanged | `call_uncallable_callee.lang`, `clean_but_unpublishable.lang` | `core.callee_not_callable`, `core.origin_omitted` | ✓ PASS |
| `relay_escort_witness.lang` divergence untouched | `lang --json check testdata/phase07/relay_escort_witness.lang` | `status: pass` (check admits; corevalidate's independent refusal is asserted separately in its own test, per the disclosed, deferred D-03-02 divergence) | ✓ PASS (unchanged, deferred to Phase 08/09) |
| Named unit/integration tests | `go test ./internal/compiler/check/... -run 'TestCallArgumentTypeMismatchRefused\|TestCallTargetTypeDerivedFromCalleeReturn\|TestOpCallTargetTypeIDUnchangedAcrossAcceptingCorpus\|TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus\|TestCallArgumentTypeCheckMutationKilled\|TestCallReturnTypeDerivationMutationKilled' -v` | 6/6 PASS | ✓ PASS |
| Named corevalidate tests | `go test ./internal/compiler/corevalidate/... -run 'TestPeerRefusesCallArgumentTypeMismatch\|TestPeerRefusesCallReturnTypeMismatch\|TestCallTypePeersIndependentOfCheck\|TestCallTypePeerMutationMatrix\|TestCheckCallTypeContractArgumentPeerSeamKilled\|TestCheckCallTypeContractReturnPeerSeamKilled' -v` | 6/6 PASS | ✓ PASS |
| Full `sh scripts/verify-phase7.sh` (includes full `go test ./...`) | run live, full output captured | `status: pass` on every `verify` JSON emitted, `kind-exhaustive-dispatch-phase07` lane lists both new controls `pass`, zero `FAIL` lines in the whole transcript | ✓ PASS |
| Golden/core artifacts untouched | `git status --porcelain testdata/phase1/evidence.golden.json testdata/phase2/evidence.golden.json testdata/phase5/coordinated_lie.core.json` | empty; `git log` on each shows no commit from this plan | ✓ PASS |
| `relay_escort_witness.lang` parity exception unwidened | `grep -n relay_escort_witness internal/compiler/corevalidate/*.go` | Same named test/exception present, unchanged shape | ✓ PASS |

**Note on transient flakes:** an earlier isolated run of `sh scripts/verify-phase7.sh` (run concurrently with other background test invocations in this same verification session) showed unrelated flakes in `internal/compiler/cache` (`cache.input_undeclared`), `internal/compiler/cgen` (`native.timeout`), and `internal/compiler/measure` (`measure.probe_timeout`) — none in `check`, `corevalidate`, or `session`, and none touching this plan's files. A subsequent clean, non-concurrent full run (captured in full above) showed zero `FAIL` lines across all packages, consistent with 07-09-SUMMARY.md's own account of environmental contention under concurrent test execution on this machine.

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|-----------------|-------------|--------|----------|
| SEM-04 | 07-03, 07-04, 07-07 | `OpCall` real at all six dispatch sites, both exhaustive controls green | ✓ SATISFIED | Unchanged by 07-09; re-confirmed live |
| SEM-05 | 07-01, 07-02, 07-05, 07-08, 07-09 | Digest-bound signature summary carries everything a caller needs for admission; no admission reads a callee body; **the signature is actually consulted for argument-type soundness** | ✓ SATISFIED (gap closed) | The `Parameters[].Type` field is now read and enforced at the one call site meant to consult it; body-blindness preserved by construction (new contract table is AST-derived, pre-body) |
| SEM-06 | 07-02, 07-05 | Call admitted only when callee callable ⊆ publishable; stable refusal code | ✓ SATISFIED | `call_uncallable_callee.lang` → `core.callee_not_callable`, unchanged |
| SEM-07 | 07-06, 07-07 | Call graph constructed, direct/mutual/indirect cycles refused by name, never a hang | ✓ SATISFIED | Full cycle corpus re-confirmed unchanged |
| QLT-08 | all 9 plans | Every new interprocedural control mutation-killed in its introducing plan | ✓ SATISFIED, with disclosed judgment call | Direct-API fault-injection for the two 07-09 controls judged sufficient — see "Judgment Call" section above |

No orphaned requirements: all five IDs mapped to Phase 07 in REQUIREMENTS.md (lines 15-25, 89) appear in at least one plan's `requirements` frontmatter field, including 07-09's `[SEM-05, QLT-08]`. REQUIREMENTS.md marks all five `[x]` complete.

### Anti-Patterns Found

No `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER` markers found in any file 07-09 modified. The prior verification's single defect (CR-01) is closed with a real, tested fix — not a stub, not a suppressed warning.

### Human Verification Required

None. All 07-09 claims were independently reproduced programmatically in this
session: the fixture verdict, the source-level derivation, the peer's
independence (via import grep and source read), the control wiring, and the
full test/verify-script runs. The QLT-08 mutation-kill-technique question is a
judgment call, not a fact requiring human observation, and is resolved above
with explicit reasoning rather than deferred.

### Gaps Summary

None remaining. The single FAILED must-have truth from the prior verification
— "a call is admitted only when it can be proven safe from the callee's
signature" — is now demonstrably true: `check.resolveCallBinding` refuses a
type-mismatched call before constructing `core.OpCall`, derives the call
result's type from the callee's own declared return contract (fail-closed when
unresolvable), and `corevalidate` independently re-derives both facts from a
materially different input path with no shared helper. The
`FunctionSignature.Parameters[].Type -> call-site argument type check` key
link marked NOT WIRED is now wired and corpus-asserted
(`TestSignatureParameterTypeMatchesAdmissionContractAcrossCorpus`). No
previously-admitted or previously-refused fixture changed verdict; no
committed golden or core artifact was modified; the `relay_escort_witness.lang`
divergence remains disclosed and deferred to Phase 08/09, untouched.

The retained limitation (nominal constructor-string comparison, too weak once
a parameterized shape becomes callable) is disclosed in PHASE-07-DEBT.md
(D-07-46) rather than left implicit, alongside two related disclosures
(D-07-47, D-07-48). This is appropriately scoped future work, not a gap in
this phase's own goal.

Phase 07 is fully closed. All 5 requirement IDs (SEM-04, SEM-05, SEM-06,
SEM-07, QLT-08) are satisfied. Ready to proceed to Phase 08.

---

_Verified: 2026-09-09_
_Verifier: Claude (gsd-verifier)_
