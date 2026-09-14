---
phase: 09-peer-re-derivation-and-d-03-02-closure
plan: 05
subsystem: testing
tags: [ownership, call-graph, corevalidate, check, decode-validation, structural-absence]

# Dependency graph
requires:
  - phase: 09-peer-re-derivation-and-d-03-02-closure
    provides: "09-01's peerLoanCarryFact/chainPeerLoanCarry substrate and the D-09-01/D-09-02 OpCall consult that closed twin_a_accept.lang/relay_depth2_accept.lang's corevalidate divergence"
provides:
  - "Structural-absence proof that call-site ownership-convention override is not expressible in source (grammar/AST layer)"
  - "Structural-absence proof plus closed-set decode confirmation that call-site ownership-convention override is not expressible in a hostile core artifact (core layer)"
  - "Cross-peer agreement proof that check and corevalidate independently derive the same per-call-site move/borrow classification and post-call loan state"
affects: [09-10, phase-10-interp]

# Actuals (#2632)
actuals:
  tokens: 8664
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Structural-absence proof: explicit expected field-name list + reflection scan, rather than a substring heuristic alone, so an additive field lands in the test's own diff"
    - "Static source-text scan (go source, not AST) to pin a closed keyword/vocabulary set as a tripwire against silent future additions"
    - "Cross-peer agreement compared as DATA (classification struct equality) over independently-derived signatures, never by sharing a comparison implementation"
    - "Synthetic core.Program construction (never through the parser) for a peer-only negative-direction test case"

key-files:
  created:
    - internal/compiler/syntax/syntax_convention_override_test.go
    - internal/compiler/core/core_convention_absence_test.go
    - internal/compiler/check/check_call_transfer_agreement_test.go
  modified: []

key-decisions:
  - "Proved the source-layer and core-layer non-expressibility claims as two separate, separately-headed tests/files rather than one combined test, per D-09-35's 'corevalidate validates a core it did not produce' distinction"
  - "Did NOT mint a new fault-injection seam for the core layer (D-09-35): ParameterContract.Mode's existing closed-set decode check at DecodeInterface is the confirmed fail-closed control"
  - "Did NOT extract a shared classification helper for Phase 10's interp peer (D-09-38): the one comparison helper lives in check_call_transfer_agreement_test.go only"
  - "Excluded call_argument_used_twice.lang from the cross-peer sweep -- it is a negative control check itself refuses (ownership.use_after_move), not an admitted fixture"
  - "Built the negative-direction (differing Parameters[0].Mode/Return.Mode) test case as a synthetic core.Program rather than a .lang fixture, since no existing admitted fixture calls a borrow-returning function safely"

patterns-established:
  - "A per-call-site classification struct (callSiteClassification) compared by Go struct equality across two independently-derived core.FunctionSignature values is the reusable shape for any future cross-peer agreement test in this codebase"

requirements-completed: []  # OWN-05 stays Pending: also carried by plan 09-10, and must never be marked Complete this phase (D-09-37) -- Phase 09 ships only 2 of OWN-05's 3 named derivers (check, corevalidate); interp is Phase 10.

coverage:
  - id: D1
    description: "Call-site override of a callee's declared ownership convention is proven non-expressible in Lang source (structural absence across all three call-parse routes and every convention-shaped keyword)"
    requirement: "OWN-05"
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_convention_override_test.go#TestConventionOverrideNotExpressibleInSource"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/syntax_convention_override_test.go#TestCallArgumentGrammarAdmitsBareIdentifiersOnly"
        status: pass
    human_judgment: false
  - id: D2
    description: "Call-site override of a callee's declared ownership convention is proven non-expressible in a hostile/corrupted core artifact, and the existing ParameterContract/ReturnContract Mode decode control is confirmed fail-closed with no new seam"
    requirement: "OWN-05"
    verification:
      - kind: unit
        ref: "internal/compiler/core/core_convention_absence_test.go#TestConventionOverrideNotExpressibleInCore"
        status: pass
      - kind: unit
        ref: "internal/compiler/core/core_convention_absence_test.go#TestParameterModeDecodeIsClosedSet"
        status: pass
    human_judgment: false
  - id: D3
    description: "check and corevalidate independently derive the same per-call-site move/borrow classification and post-call loan state, over every admitted call-bearing fixture plus a negative-direction synthetic case"
    requirement: "OWN-05"
    verification:
      - kind: unit
        ref: "internal/compiler/check/check_call_transfer_agreement_test.go#TestCallSiteTransferClassificationAgreesAcrossPeers"
        status: pass
      - kind: unit
        ref: "internal/compiler/check/check_call_transfer_agreement_test.go#TestCallSiteTransferClassificationAgreesWhenModesDiffer"
        status: pass
    human_judgment: false

duration: 55min
completed: 2026-09-10
status: complete
---

# Phase 09 Plan 05: Ownership Transfer Has One Meaning — Source-Absence, Core-Absence, and Cross-Peer Agreement Summary

**Proved call-site ownership-convention override is structurally absent at both the source and core layers (as two separately-stated claims), and that `check`/`corevalidate` independently derive the same move-vs-borrow classification and post-call loan state, with no new seam and no shared helper minted for either proof.**

## Performance

- **Duration:** 55 min
- **Started:** 2026-09-10T21:50:00Z
- **Completed:** 2026-09-10T22:45:00Z
- **Tasks:** 3
- **Files modified:** 3 (all new test files)

## Accomplishments
- Source-layer structural-absence proof: no grammar production admits a `take`/`borrow`/`borrow mut` annotation on a call argument, in any of the three call-parse routes (bare call, `try` call, `discard ... because`), and `ast.RHS` declares no field a per-argument convention could occupy even if a future parser change admitted the token sequence. A static scan of `parser.go`'s `normalizeOwnershipTokens` pins the ownership-keyword vocabulary as a tripwire against a future keyword silently opening a hole.
- Core-layer structural-absence proof: `core.LinearOperation`'s complete field set is asserted against an explicit expected list (no field named mode/convention/ownership/transfer), and an unknown `convention_override` JSON key cannot survive a decode/round-trip into any populated field. Confirmed and extended coverage of the EXISTING `core.interface_invalid_mode` closed-set decode check with a five-value hostile table (empty, capitalized, plausible-looking, override-shaped, whitespace-padded) plus the three-value legal-admission control, for both `Parameters[i].Mode` and `Return.Mode`. No new seam was minted.
- Cross-peer agreement proof: over every admitted call-bearing fixture (straight-line and match-arm-bodied), `check`'s own admission-time signature consult (`buildCallSignatureTable`, sourced from `originvalidate.BuildInterface`) and `corevalidate`'s independently re-derived `PeerSignatures()` agree, call site by call site, on both the declared parameter mode and the return-derives-a-borrow bit. `PeerSiteCoverage()` was asserted per fixture shape (never inferred), and a synthetic `core.Program` (built directly, never through the parser) proved the agreement holds even for a callee whose `Return.Mode` genuinely differs from its `Parameters[0].Mode`, so the comparison is not vacuous.

## Task Commits

Each task was committed atomically:

1. **Task 1: Source-layer non-expressibility, proven as structural absence** - `2d5d822` (test)
2. **Task 2: Core-layer field absence plus the EXISTING closed-set decode control — and no new seam** - `d4c5f9c` (test)
3. **Task 3: One meaning for the per-call-site classification, derived independently by both peers** - `acde1cc` (test)

**Plan metadata:** (this commit, docs)

## Files Created/Modified
- `internal/compiler/syntax/syntax_convention_override_test.go` - Source-layer structural-absence proof (grammar refusal across all three call-parse routes and every convention-shaped keyword; `ast.RHS` field-shape assertion; a static scan pinning the ownership-keyword vocabulary)
- `internal/compiler/core/core_convention_absence_test.go` - Core-layer structural-absence proof (`core.LinearOperation`'s explicit field-name set; JSON round-trip check for an unknown key) plus confirmation/extension of the existing `Mode` closed-set decode control
- `internal/compiler/check/check_call_transfer_agreement_test.go` - Cross-peer per-call-site classification agreement sweep over admitted fixtures, plus a synthetic negative-direction case

## Decisions Made
- Stated the source-layer and core-layer non-expressibility claims as two entirely separate test files with their own header-comment arguments, per D-09-35's point that `corevalidate` validates a core it did not produce -- a hostile producer never crosses the parser, so the source-layer refusal proves nothing about that trust boundary.
- Confirmed and extended the EXISTING `ParameterContract.Mode`/`ReturnContract.Mode` closed-set decode check as the core-level fail-closed control instead of minting a new fault-injection seam (D-09-35): `core.LinearOperation` has no undecoded slot for a call-site override to hide in, so a seam here would be dead weight for a shape that does not exist.
- Kept the cross-peer comparison helper (`classifyFromSignature`, `callSiteClassification`) local to `check_call_transfer_agreement_test.go` rather than extracting it into a production file "ready for" Phase 10's `interp` peer (D-09-38) -- `check` and `corevalidate` already each read `signature.Parameters[0].Mode` independently, sharing nothing beyond the struct shape.
- Excluded `testdata/phase07/call_argument_used_twice.lang` from the cross-peer sweep: it is a negative control `check` itself refuses (`ownership.use_after_move`), not an admitted fixture, so it does not belong in a sweep whose precondition is "fixtures `check` admits."
- Built the negative-direction test case (a callee whose `Return.Mode` differs from `Parameters[0].Mode`) as a hand-built synthetic `core.Program`, since no existing admitted `.lang` fixture calls a borrow-returning function in a way `check` admits cleanly (every such fixture in the corpus is either refused by `check`'s own interprocedural liveness law or is one of the two `twin`/`relay` accept fixtures whose callee returns `owned`, not a borrow).

## Deviations from Plan

None - plan executed exactly as written. All three tasks' `<action>` and `<acceptance_criteria>` were followed; no production file was touched (`parser.go`, `token.go`, `internal/compiler/ast/`, and `core.go` are all unmodified, confirmed by `git status --short`).

## Issues Encountered
- The plan's Task 3 action text named `corevalidate`'s `peerLoanCarryFact` as the peer-side source of the "resulting loan state after call" fact, but that field is unexported with no `Result` accessor, and this plan's frontmatter forbids modifying any production file to add one. The comparison was instead built from the two exported, genuinely-independent readers of the SAME declared contract fact (`check`'s `callSignatureTable`, sourced from `originvalidate.BuildInterface`, versus `corevalidate`'s `PeerSignatures()`) -- exactly the pairing `buildCallSignatureTable`'s own doc comment in `check.go` names as "the two genuinely independent derivations." This satisfies D-09-36's shared-fact requirement (classification plus resulting loan-state bit, both derived from the callee's declared signature) without adding any new production surface.
- `call_argument_used_twice.lang` (initially included in the fixture sweep) turned out to be a negative control that `check` itself refuses; removed from the sweep after the first test run surfaced it, per the task's own precondition ("fixtures `check` admits").

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- OWN-05 remains `Pending` in REQUIREMENTS.md, as required: this plan and 09-01 together prove the `check`/`corevalidate` half, but OWN-05's own text names a third deriver (`interp`) that Phase 10 delivers (verified there by TRU-03, and in Phase 11 by NAT-06). Plan 09-10 owns the requirement-document split into OWN-05a (Phase 09)/OWN-05b (Phase 10) -- this plan does not pre-empt that.
- `go test ./...` and `go vet ./...` are green with these three new test files added; no other plan in Wave 2 is blocked by this one (only 09-01 was a dependency, already complete).

---
*Phase: 09-peer-re-derivation-and-d-03-02-closure*
*Completed: 2026-09-10*

## Self-Check: PASSED

All three created test files and this SUMMARY.md verified present on disk; all three task commit hashes (`2d5d822`, `d4c5f9c`, `acde1cc`) verified present in `git log`.
