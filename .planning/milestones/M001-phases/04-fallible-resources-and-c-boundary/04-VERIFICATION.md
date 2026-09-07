---
phase: 04-fallible-resources-and-c-boundary
verified: 2026-09-05T23:45:00Z
status: passed
score: 8/8 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 7/8
  gaps_closed:
    - "Every core.ForeignContract string field cgen ever splices into generated C text — not just Symbol — must be validated or escaped before EmitForeignHeader/EmitForeignConformance emit it (FFI-01) — closed by plan 04-13."
  gaps_remaining: []
  regressions: []
deferred: []
human_verification: []
---

# Phase 4: Fallible Resources and C Boundary Verification Report

**Phase Goal:** A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined.
**Verified:** 2026-09-05
**Status:** passed
**Re-verification:** Yes — fifth round, after gap-closure plan 04-13 (FFI-01 sibling-field
injection audit)

## Goal Achievement

### Observable Truths

The fourth round scored 7/8, with exactly one failed truth (2b): `Allocator`, `Unwind`, and
`NonlocalExit` — and, as this round's fresh code review additionally surfaced, `Fails`,
`InitializedState`, `Capture`, `Retention`, `Aliasing`, `Layout.ForeignTypeName`, and each
`Layout.Fields[].Name`/`.CType` — were spliced into generated C with no shape validation, and
the gap was reachable from ordinary, honest Lang source (a quoted foreign-policy value), not
merely a corrupted `core.Program`. Plan 04-13 claims to close this with three independent audit
layers. This verification did not accept that claim on the SUMMARY's word: it read the
production code directly, ran every named falsifier itself, and independently reproduced the
end-to-end injection scenario in its own throwaway test (removed after use, working tree left
clean) to confirm the honest-source path is now refused.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: A fallible acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure | ✓ VERIFIED | Regression spot-check: unchanged since round 4; not touched by 04-13. |
| 1a | Independent rederivation proves release order at TERMINAL merge points | ✓ VERIFIED | Regression spot-check: unchanged; not touched by 04-13. |
| 1b | `rederive`'s backward walk stays defined (no hang) against a cyclic "ok"-edge chain | ✓ VERIFIED | Re-ran `TestCyclicOkEdgeChainRefusedNotHung` and `TestAcyclicChainsStillValidateUnderCycleGuard` in this session — both PASS. |
| 1c | `checkReleaseOrder`'s rederivation must not silently collapse two disagreeing histories converging on an INTERIOR block | ✓ VERIFIED | Re-ran `TestInteriorMergeDivergentHistoriesRefused` (both subtests) and `TestInteriorMergeAgreeingHistoriesAccepted` in this session — PASS. Not touched by 04-13. |
| 2 | SC2: Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable | ✓ VERIFIED | Regression spot-check: unchanged; `core.ForeignContract` field set unchanged, all fields still emitted. |
| 2a | `core.ForeignContract.Symbol` is validated as a syntactically safe C identifier before cgen splices it into generated C | ✓ VERIFIED | Re-ran `TestForeignSymbolNotIdentifierRefused`, `TestForeignSymbolInjectionNeverReachesGeneratedC`, `TestForeignEmittersRefuseNonIdentifierSymbolIndependently` in this session — all PASS. |
| 2b | Every OTHER `core.ForeignContract` string field spliced into generated C must be validated the same way `Symbol` now is, at the source-admission, corevalidate, and cgen layers, closing the injection class independently of any single layer | ✓ VERIFIED | **Fixed by plan 04-13.** Confirmed by direct code read: `check.validForeignPolicyValue` (check.go:1095-1110) plus the `check.foreign_policy_value_unsafe` refusal fired inside `collectForeignSymbols`'s policy loop (check.go:~1149) before the key-specific switch, applied to every declared policy value; `corevalidate.go:349` adds `foreign.policy_value_not_identifier` (Allocator/Unwind/NonlocalExit identifier-shape, positioned after the existing `foreign.unwind_policy_undeclared` non-empty check so an omitted policy keeps its own code) and `corevalidate.go:386` adds `foreign.contract_field_not_c_safe` via `foreignContractFieldsCSafe` (corevalidate.go:1765-1781, covering Fails/InitializedState/Capture/Retention/Aliasing/Layout.ForeignTypeName/Layout.Fields[].Name+CType); `cgen.go`'s `singleForeignFunction` independently calls `unsafeForeignContractField` (cgen.go:873-912) at cgen.go:1382, gating all three exported `EmitForeign*` entry points that never call `Validate`. I independently re-ran, myself, all four named falsifiers — `TestForeignPolicyValueUnsafeRefusedAtAdmission` (check, 8 hostile shapes + 1 negative), `TestForeignPolicyValueInjectionRefusedFromSource` (session, end-to-end tracer), `TestForeignPolicyValueNotIdentifierRefused` and `TestForeignContractCommentSafetyRefused` (corevalidate), `TestForeignPolicyValueInjectionNeverReachesGeneratedC` (cgen) — all PASS. I additionally reproduced the scenario independently in a throwaway test against `session.Check` on `testdata/phase4/foreign_policy_value_injection.lang` (the exact fixture and payload the prior round's reproduction used): result carries `Diagnostics[0].Code == "check.foreign_policy_value_unsafe"` and `len(result.Program.Functions) == 0` — the honest-source path that previously produced a live injected C function definition with zero diagnostics is now refused before a checked program is even produced. Test file removed after use; `git status --short` confirms a clean tree (only pre-existing unrelated `.planning/state.json`/`milestone.lock` changes). |
| 3 | SC3 first half: Panic cannot cross the ordinary non-unwinding C boundary | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-13. |
| 4 | SC3 second half: A foreign nonlocal exit cannot silently bypass Lang cleanup | ✓ VERIFIED | Regression spot-check: unchanged, not touched by 04-13. |
| 5 | SC4: Interpreter and native executions agree on primary failure and cleanup events | ✓ VERIFIED | Not touched by 04-13; established by orchestrator this round: `go test ./... -count=1` and `go vet ./...` both exit 0, and all 23 named prior-phase (01/02/03) regression targets pass. |

**Score:** 8/8 truths verified. The one remaining gap from round 4 is closed, independently
confirmed by direct code read, by re-running every named falsifier, and by an independent
end-to-end reproduction of the exact previously-successful injection scenario.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` | Source-admission audit of every declared foreign policy value | ✓ VERIFIED | `validForeignPolicyValue` (byte-loop identifier predicate) and the `check.foreign_policy_value_unsafe` refusal confirmed present, applied uniformly to every policy key's value before the key-specific switch. |
| `internal/compiler/corevalidate/corevalidate.go` | Independent, source-blind shape audit of every spliced `ForeignContract` field | ✓ VERIFIED | `foreign.policy_value_not_identifier` (Allocator/Unwind/NonlocalExit) and `foreign.contract_field_not_c_safe` (`foreignContractFieldsCSafe`, remaining fields) both confirmed present and correctly ordered relative to existing checks. |
| `internal/compiler/cgen/cgen.go` | All emitted C text (declarations and comments) sanitized, independent of `corevalidate.Validate` having run | ✓ VERIFIED | `unsafeForeignContractField` (cgen.go:873-912) confirmed gating `singleForeignFunction` (cgen.go:1382), the single resolution point for `EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance`. |
| `internal/compiler/corevalidate/corevalidate_test.go` / `internal/compiler/cgen/cgen_test.go` / `internal/compiler/check/check_test.go` / `internal/compiler/session/session_test.go` | Falsifiers for every adversarial shape claimed defended, plus an end-to-end tracer | ✓ VERIFIED | All five named tests re-run in this session, all PASS, including the comment-terminator payload across every field and the honest-source end-to-end tracer. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `ast.ForeignPolicy.Value` (source text) | `check.collectForeignSymbols` | `validForeignPolicyValue`, span-bearing refusal | ✓ WIRED | Confirmed check.go:1095-1149; end-to-end tracer test passes. |
| `core.ForeignContract.Allocator`/`Unwind`/`NonlocalExit` | `corevalidate.linear` | `foreign.policy_value_not_identifier`, after `foreign.unwind_policy_undeclared` | ✓ WIRED | Confirmed corevalidate.go:349; ordering case in `TestForeignPolicyValueNotIdentifierRefused` passes, proving the existing code is not displaced. |
| `core.ForeignContract`'s remaining spliced fields | `corevalidate.linear` | `foreign.contract_field_not_c_safe` | ✓ WIRED | Confirmed corevalidate.go:386, `foreignContractFieldsCSafe` at corevalidate.go:1765; `TestForeignContractCommentSafetyRefused` passes, including the CType-specific negative case (`unsigned char` accepted). |
| `cgen.singleForeignFunction` | `EmitForeignManifest`/`EmitForeignHeader`/`EmitForeignConformance` | `unsafeForeignContractField`, independent of `Validate` | ✓ WIRED | Confirmed cgen.go:1382; `TestForeignPolicyValueInjectionNeverReachesGeneratedC` passes for all 11 field cases. |
| Ordinary Lang source (`allocator: "*/ ... /*"`) | `session.Check` | `check.foreign_policy_value_unsafe` refusal, zero checked functions | ✓ WIRED (now refused) | Independently reproduced in this session: `session.Check` on `testdata/phase4/foreign_policy_value_injection.lang` returns the expected diagnostic and zero functions — the honest-source path that previously escaped is now closed. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Honest fixture refused at source admission (own reproduction, not SUMMARY's word) | ad hoc test against `session.Check` on `testdata/phase4/foreign_policy_value_injection.lang`, run and removed in this session | `Diagnostics[0].Code == "check.foreign_policy_value_unsafe"`, `len(Program.Functions) == 0` | ✓ PASS |
| Source-admission falsifier (8 hostile shapes + 1 negative) | `go test ./internal/compiler/check -run '^TestForeignPolicyValueUnsafeRefusedAtAdmission$' -v` | PASS, all 9 subtests | ✓ PASS |
| End-to-end tracer | `go test ./internal/compiler/session -run '^TestForeignPolicyValueInjectionRefusedFromSource$' -v` | PASS | ✓ PASS |
| corevalidate identifier-shape and comment-safety audits | `go test ./internal/compiler/corevalidate -run '^(TestForeignPolicyValueNotIdentifierRefused|TestForeignContractCommentSafetyRefused)$' -v` | PASS, all subtests (52 total across both) | ✓ PASS |
| cgen independent peer guard | `go test ./internal/compiler/cgen -run '^TestForeignPolicyValueInjectionNeverReachesGeneratedC$' -v` | PASS, all 11 field cases | ✓ PASS |
| Round-4 regression: interior-merge and cyclic-guard | `go test ./internal/compiler/corevalidate -run '^(TestInteriorMergeDivergentHistoriesRefused|TestInteriorMergeAgreeingHistoriesAccepted|TestCyclicOkEdgeChainRefusedNotHung|TestAcyclicChainsStillValidateUnderCycleGuard)$' -v` | PASS | ✓ PASS |
| Round-4 regression: Symbol audit | `go test ./internal/compiler/corevalidate -run '^TestForeignSymbolNotIdentifierRefused$' -v` and `go test ./internal/compiler/cgen -run '^(TestForeignSymbolInjectionNeverReachesGeneratedC|TestForeignEmittersRefuseNonIdentifierSymbolIndependently|TestExistingEmittersAreByteIdentical)$' -v` | PASS | ✓ PASS |
| Accepting path unchanged | `git diff --stat -- testdata/` | Empty (no pre-existing fixture/golden byte moved) | ✓ PASS |
| Full build/test/vet (already established this round by the orchestrator) | `go test ./... -count=1`; `go vet ./...`; 23 named prior-phase regression targets | Exit 0; exit 0; all pass | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|----------------|--------------|--------|----------|
| SEM-03 | 04-01, 04-02, 04-04, 04-06, 04-07, 04-08 | `Result` propagation and ignored-result rules produce explicit control flow | ✓ SATISFIED | Truths 1/3/4 verified above; not implicated by 04-13. **Note:** REQUIREMENTS.md's status table (line 141) and checkbox (line 42) still show `[ ]`/"Gaps Found" for SEM-03, a pre-existing tracking lag noted (but not caused) by the fourth round's verification and not updated since — this verification's own findings do not contradict SEM-03's satisfaction; the table entry should be reconciled but that is a documentation task, not an open code gap. |
| RES-01 | 04-02, 04-05, 04-07, 04-08, 04-10, 04-11 | Partially initialized noncopyable resources release exactly once in reverse order | ✓ SATISFIED | Truths 1/1a/1b/1c verified above, closed by 04-11 and independently re-confirmed with passing named tests this round; not implicated by 04-13. **Note:** REQUIREMENTS.md's status table (line 146) and checkbox (line 52) still show `[ ]`/"Gaps Found" for RES-01, unreconciled since the `8a35f1f` revert (round 3) despite 04-11 closing it and round 4 confirming it — same documentation-lag observation as SEM-03, not an open code gap. |
| FFI-01 | 04-01, 04-03, 04-05, 04-06, 04-07, 04-09, 04-12, 04-13 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, unwind obligations, all inspectable and non-inventable | ✓ SATISFIED | Truths 2/2a/2b all verified above. `Symbol` injection (04-12) and the sibling-field injection (04-13, this round) are both closed and independently confirmed. REQUIREMENTS.md's checkbox (line 54) and status table (line 147) already correctly show `[x]`/"Complete", set by 04-13's own completion commit `d38ba81` — consistent with this verification's finding. |

No orphaned requirements: SEM-03, RES-01, FFI-01 are the only IDs REQUIREMENTS.md maps to
Phase 4, and all three appear in at least one plan's `requirements:` frontmatter (04-01
through 04-13 collectively cover all three).

### Anti-Patterns Found

None newly introduced by plan 04-13's changes (`validForeignPolicyValue`, `foreign.policy_value_not_identifier`,
`foreign.contract_field_not_c_safe`, `commentSafeForeignField`/`validForeignCType`/`unsafeForeignContractField`)
— scoped, minimal, matches existing file idiom; no TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER markers in the touched
region; no value is echoed into any diagnostic or error string (confirmed by direct read — refusals pass only
`policy.Span`, `operation.ID`, or a fixed field-name string, never the field's value).

One WARNING-level finding from this round's fresh code review (04-REVIEW.md WR-01), independently confirmed by
this verification: `core.ForeignContract.Alias` (distinct from `Aliasing`, which is covered) is excluded from
both `corevalidate.foreignContractFieldsCSafe` and `cgen.unsafeForeignContractField`, and from both packages'
falsifier tables (`hostileForeignContractFields` in cgen_test.go and its corevalidate equivalent). I confirmed by
grep that this is safe TODAY — `.Alias` (not `.Aliasing`) has zero references in either `cgen.go` or
`corevalidate.go`; it is consumed only by `originvalidate.go` against a fixed `borrow`/`retain` vocabulary and is
never spliced into generated C by any emitter. My own judgment, weighing this honestly against the phase goal:
this is **not a genuine open gap against "one audited C boundary"** — the phase goal and FFI-01 concern the
boundary actually crossed, and `Alias` does not cross it today, so there is no live injection channel this round
leaves open. It IS a real defense-in-depth gap: nothing pins the invariant that `Alias` must never be spliced
without first being added to both audit tables, so a future contributor adding an `/* alias: %s */` comment line
to `EmitForeignHeader` (a natural-looking addition, since every other contract field already gets one) would
silently reopen exactly the injection class this round closed, and no falsifier here would catch it. This belongs
in `04-DEBT.md` as a forward-looking guard-rail item (either add `Alias` to both audit tables preemptively, or add
an explicit "deliberately unaudited, must be added here first if ever spliced" doc-comment plus a table entry that
asserts the invariant), not as a blocker on this phase's goal, because the phase goal is about the boundary that
exists today, not a hypothetical future emitter change. Recording it here so it is not lost, but it does not move
this round's verdict.

### Human Verification Required

None. All truths are confirmed by direct code inspection, by re-running every named falsifier in
this session, and by an independent end-to-end reproduction of the previously-successful injection
scenario (removed after use, working tree left clean) — no ambiguity requiring human judgment.

### Gaps Summary

None remaining. Round 4 closed the interior-merge collapse (RES-01, plan 04-11) and the `Symbol`
injection (FFI-01, plan 04-12). This round's plan 04-13 closes the last open gap: every remaining
`core.ForeignContract` string field `cgen` splices into generated C is now validated at three
independent layers (source admission in `check.go`, shape/comment-safety re-derivation in
`corevalidate.go`, and cgen's own independent peer guard in `singleForeignFunction`), and the
source-reachable policy values (`allocator`, `unwind`, `nonlocal_exit`) are refused at source
admission before a core artifact is ever produced — closing the honest-source path this
verification's fourth round demonstrated was open, not merely the corrupted-`core.Program` case.

I did not accept 04-13-SUMMARY.md's claim on its word: I independently confirmed the three audit
layers exist at the code locations claimed, that they are correctly ordered relative to existing
checks (so no prior refusal is displaced), that all five named falsifiers pass when I ran them
myself, and I reproduced the exact previously-successful injection scenario against
`session.Check` in my own throwaway test — confirming it is now refused with the expected
diagnostic code and zero checked functions, then removed the test and confirmed a clean working
tree.

One WARNING (04-REVIEW.md WR-01, `core.ForeignContract.Alias` unaudited but also unspliced) is
judged to be a defense-in-depth documentation/DEBT item, not an open gap against this phase's own
goal — see "Anti-Patterns Found" above for the full reasoning.

Phase 4's goal — "a noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined" — is achieved: all 8
must-haves across SEM-03, RES-01, and FFI-01 are verified, with no remaining code-level gap.

---

_Verified: 2026-09-05T23:45:00Z_
_Verifier: Claude (gsd-verifier)_
