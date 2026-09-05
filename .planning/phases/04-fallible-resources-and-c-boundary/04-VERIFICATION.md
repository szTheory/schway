---
phase: 04-fallible-resources-and-c-boundary
verified: 2026-09-05T00:00:00Z
status: gaps_found
score: 6/7 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 5/7
  gaps_closed:
    - "`corevalidate` independently rederives the expected release order by walking backward from each failure edge over the block and edge graph, and compares — it never reads what `check` wrote and shares no helper with it (D-04-07, D-12a)."
    - "One authoritative `core.ForeignContract` carries target layout, initialized state, allocator identity, capture and retention, aliasing, and unwind obligations, and all three inspectable layers are derived from it so no layer can invent a fact the contract does not carry (FFI-01, D-04-12, ROADMAP SC2)."
  gaps_remaining: []
  regressions: []
gaps:
  - truth: "corevalidate's independent rederivation walks (in particular checkReleaseOrder's rederive backward walk) are source-blind and must stay defined against a corrupted/adversarial core.Program, matching the guard already applied to every other backward/forward graph walk in the same file (blockReach, loanChainIndex.carriedLoans)."
    status: failed
    reason: >
      Newly surfaced by this run's fresh code review (04-REVIEW.md CR-01) and independently
      confirmed here by direct code read, not accepted on the review's word alone.
      `checkReleaseOrder`'s `rederive` closure (internal/compiler/corevalidate/corevalidate.go,
      the `for { ... }` loop around lines 1230-1250, walking `okEdgeInto[currentBlockID]`
      backward) has no visited-set / cycle guard. `okEdgeInto` is built directly from
      `linear.Edges` with no acyclicity check performed anywhere earlier in `Validate` or
      `blocksAndEdges` (confirmed by direct read: `blocksAndEdges`, corevalidate.go:376-476,
      checks edge/block ID uniqueness and referential closure but never checks the block
      graph is acyclic). A hand-corrupted `core.Program` with a cyclic chain of "ok"-pattern
      edges feeding into a terminal block's incoming edge causes `rederive`'s loop to never
      hit its only `break` (the map lookup `okEdgeInto[currentBlockID]` keeps succeeding
      forever around the cycle), hanging validation indefinitely. This is precisely the
      adversarial-input class the same file explicitly defends against in its sibling walks:
      `loanChainIndex.carriedLoans` (corevalidate.go ~525-545) and `blockReach`
      (corevalidate.go ~584-596) both carry visited-set guards with doc comments stating in
      nearly identical language that "a source-blind validator must stay defined against"
      exactly this shape of cyclic/corrupted input. `rederive` is the one backward-graph
      traversal in the file that lacks the equivalent guard, and it is reachable from
      `Validate` on any branch-shaped function (via `replayBlocks` → `checkReleaseOrder`),
      not gated behind any earlier acyclicity check. Confirmed absent: a repo-wide grep of
      `corevalidate_test.go` for "cyclic"/"Cyclic" returns zero matches — no falsifier
      constructs a cyclic ok-edge chain, so this gap is untested. 04-08's gap-closure plan
      (which added the merge-terminal-block falsifiers and the core.terminal_block_unreachable
      peer check) did not touch `rederive` itself and did not add cycle protection; its own
      SUMMARY confirms `checkReleaseOrder itself was NOT modified by this plan`.
      corevalidate's entire reason for existing — per its own documented design and the
      RES-01/D-04-07 must_have's own wording ("independently rederives... and compares") —
      is to stay defined against a corrupted or adversarially hand-constructed core.Program,
      not merely one honestly produced by check.go today. An unbounded hang is a real
      denial-of-service against any caller (fuzzing, future untrusted-core-artifact path)
      that runs an adversarial program through validation, and it undermines the soundness
      claim the independent rederivation exists to provide.
    artifacts:
      - path: internal/compiler/corevalidate/corevalidate.go
        issue: "rederive (the backward-walk closure inside checkReleaseOrder, ~lines 1230-1250) has no visited-set guard against a cyclic \"ok\"-edge chain, unlike blockReach and loanChainIndex.carriedLoans in the same file"
    missing:
      - "Add a visited-set to rederive (the same shape carriedLoans and blockReach already use) and treat a re-visited block during the backward walk as a hard refusal (e.g. a new core.release_order_cyclic code), not a silent truncation or infinite loop"
      - "Add a falsifier test constructing a two-block \"ok\"-edge cycle feeding into a terminal block's incoming edge, asserting the validator returns promptly (does not hang) and refuses with the new code"
deferred: []
human_verification: []
---

# Phase 4: Fallible Resources and C Boundary Verification Report

**Phase Goal:** A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined.
**Verified:** 2026-09-05
**Status:** gaps_found
**Re-verification:** Yes — after gap closure (plans 04-08, 04-09)

## Goal Achievement

### Observable Truths

Re-verified against the previous VERIFICATION.md's 7 merged must-haves. The two
previously-FAILED items were checked at full depth (exists, substantive, wired, plus
direct code/test re-derivation); the five previously-VERIFIED items were spot-checked
for regression. One new truth (source-blind robustness of the independent rederivation)
is added because this run's fresh code review surfaced a confirmed, unaddressed critical
finding directly bearing on RES-01/D-04-07's own wording.

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: A fallible two-step (in practice three-step) acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure | ✓ VERIFIED | Unchanged since prior verification. `check.go` materializes `OpRelease` reverse of completed-acquisition order; `TestThreeAcquisitionReleaseOrder`, `TestPartialAcquisitionReleasesOnlyCompleted`, `TestPhase4CorpusThreeEngineAgreement` all re-run and pass. `scripts/verify-phase4.sh` re-run directly this session: `lane:release-order-transposed` and `lane:release-omitted` both `pass` with recomputed_work 525 and 151. |
| 1a | The *independent* rederivation (`corevalidate.checkReleaseOrder`) proves release order without relying on `check`'s own bookkeeping, for every reachable terminal-block shape (gap 1 closed by plan 04-08) | ✓ VERIFIED | Direct code read confirms `checkReleaseOrder` (corevalidate.go ~1252-1298) no longer skips a terminal block whose incoming-edge count != 1: the guard is now `len(incoming) > 0` (refusal, not skip) and the walk iterates `for _, edge := range incoming { expected := rederive(edge); ... }`, comparing every incoming edge's independently-rederived expectation against the single fixed `actual` release list — closing item 1 of the prior gap. Item 2 (a structural peer check independent of checkReleaseOrder) is closed in an adapted, human-decision-recorded form: `core.terminal_block_unreachable` in `blocksAndEdges` (corevalidate.go ~416-447) refuses any non-entry OpReturn/OpFail-terminated block with zero incoming edges — an at-least-one/non-entry check, not the literal "exactly one" the prior gap's `missing:` list first suggested. That literal form was tried (04-REVIEW-FIX.md's CR-01 pass) and broke `discard_because.lang`'s legitimate two-incoming-edge merge; the auto-selected weaker check still closes the independence gap without narrowing the language. Falsifier tests re-run directly this session and pass: `TestMergeTerminalBlockDivergentReleaseSetsRefused` (constructs a hand-corrupted divergent merge, refused with `core.release_order_mismatch`), `TestMergeTerminalBlockAgreeingChainsAccepted`, `TestTerminalBlockUnreachableRefused`, `TestLegitimateDiscardMergeStillValidates`. Mutation-kill reverts recorded in 04-08-SUMMARY.md reproduce red on both hunks. **Judgment on the decision:** the auto-selected weaker check satisfies the prior gap's *substance* (an independent structural check now exists, and the merge-point vacuous-skip is closed by the per-edge walk) even though it does not literally implement missing-item 2's "exactly one" wording — that literal wording was infeasible against the language's own legitimate `discard...because` merge shape, so the substitution is judged sound, not a scope reduction. |
| 2 | SC2: Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable | ✓ VERIFIED | Unchanged since prior verification. `core.ForeignContract` carries every named obligation; `EmitForeignHeader`/`EmitForeignConformance` re-confirmed present and correct; `control:foreign.layout_mismatch` re-run, passes with recomputed_work=1. |
| 2a | Zero optimizer-visible attributes are emitted, and the control scans all emitted C (D-04-13's own wording) — gap 2 closed by plan 04-09 | ✓ VERIFIED | Direct code read of session.go confirms `ScanForBannedAttributes` is now called with all eight arguments: `tracerCSource, tracerManifest, tracerHeader, tracerConformance, releaseCSource, releaseManifest, releaseHeader, releaseConformance` (session.go ~2230) — `EmitForeignHeader` and `EmitForeignConformance` output for both fixtures is included, closing the coverage gap. `addLane` records `recomputed_work=8` (two fixtures × four artifacts), matching `TestAttributeScanLaneCoversEveryInspectableLayer`'s pin, re-run and passing. `TestAttributeInjectionIntoConformanceOnlyMakesControlFail` (constructs the previously-untested conformance-only-region injection, isolating the region after the embedded header via `strings.Index`) re-run and passes. `scripts/verify-phase4.sh` re-run directly this session: `lane:foreign-no-unproven-attributes` reports `"status":"pass","recomputed_work":8`. |
| 3 | SC3 first half: Panic cannot cross the ordinary non-unwinding C boundary | ✓ VERIFIED | Unchanged since prior verification; not re-derived at full depth (no code in this path was touched by 04-08/04-09), spot-checked via the re-run gate: `lane:foreign-unwind-forbidden` passes. |
| 4 | SC3 second half: A foreign nonlocal exit cannot silently bypass Lang cleanup | ✓ VERIFIED | Unchanged since prior verification; spot-checked via re-run gate: `lane:defect-no-release`, `lane:nonlocal-exit-undetected` both pass. |
| 5 | SC4: Interpreter and native executions agree on primary failure and cleanup events | ✓ VERIFIED | Unchanged since prior verification; `sh scripts/verify-phase4.sh` re-run directly this session, exits 0, all 14 lanes `"status":"pass"`, all required control identifiers present, `go test ./...` and `go vet ./...` both clean. |
| 6 | corevalidate's independent rederivation walks are source-blind and must stay defined (never hang) against a corrupted/adversarial core.Program, matching the cycle-guard discipline already applied to every other backward/forward graph walk in the same file | ✗ FAILED | New finding this run, surfaced by 04-REVIEW.md's CR-01 and independently confirmed by direct code read (not accepted on the review's word alone): `checkReleaseOrder`'s `rederive` closure has no visited-set guard against a cyclic "ok"-edge chain among non-terminal blocks, unlike `blockReach` and `loanChainIndex.carriedLoans` in the same file, both of which carry near-identical doc comments stating a source-blind validator "must stay defined against" exactly this adversarial shape. Confirmed absent from `blocksAndEdges`: no earlier acyclicity check on the declared block/edge graph. Confirmed untested: zero "cyclic"/"Cyclic" matches in `corevalidate_test.go`. Not introduced by 04-08 (which added a different check, `core.terminal_block_unreachable`, and did not modify `rederive`) — this is a pre-existing gap in code central to RES-01's own "independently rederives... and compares" wording, now surfaced by review and not yet closed by any gap-closure plan. |

**Score:** 6/7 truths verified (both originally-reported gaps genuinely closed on direct code inspection and passing named tests; one new gap surfaced by this run's fresh code review and independently confirmed).

### Deferred Items

None — the new finding (truth 6) bears directly on RES-01's own must-have wording (an independent, source-blind rederivation) and is not deferred to a later phase in ROADMAP.md.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/corevalidate/corevalidate.go` | independent rederivation of release order and foreign refusals, source-blind against corrupted input | ⚠️ PARTIAL | The vacuous-skip gap (prior gap 1) is closed: `checkReleaseOrder` now refuses (not skips) any terminal block with zero incoming edges and walks every incoming edge independently; `core.terminal_block_unreachable` adds a structural peer check. New gap: `rederive`'s backward walk has no cycle guard, unlike its siblings `blockReach`/`carriedLoans` — see truth 6. |
| `internal/compiler/session/session.go` | zero-attribute scan over all emitted C, including EmitForeignHeader/EmitForeignConformance | ✓ VERIFIED | `lane:foreign-no-unproven-attributes` now scans 8 arguments (both fixtures × compiled program, manifest, header, conformance); `TestAttributeScanLaneCoversEveryInspectableLayer` pins the count so a dropped argument turns a named test red. |
| `internal/compiler/corevalidate/corevalidate_test.go` | falsifiers proving the merge-terminal-block rederivation and structural peer check are load-bearing | ✓ VERIFIED | `TestMergeTerminalBlockDivergentReleaseSetsRefused`, `TestMergeTerminalBlockAgreeingChainsAccepted`, `TestTerminalBlockWithNoIncomingEdgeRefused`, `TestLegitimateDiscardMergeStillValidates`, `TestTerminalBlockUnreachableRefused` all present and passing; mutation-kill reverts recorded. No falsifier for the cycle gap (truth 6) exists. |
| `internal/compiler/session/session_test.go` | falsifiers proving the conformance-layer scan and artifact-count pin are load-bearing | ✓ VERIFIED | `TestAttributeInjectionIntoConformanceOnlyMakesControlFail`, `TestAttributeScanLaneCoversEveryInspectableLayer` both present and passing; mutation-kill reverts recorded in 04-09-SUMMARY.md. |
| `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` | Mutation-Kill Register rows for both gap-closure plans | ✓ VERIFIED | Four new rows present (two per plan), each recording verbatim revert-and-fail output. |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `core.go` | `corevalidate.go` | independent release-order rederivation, refusal not skip, for every reachable terminal-block shape | ✓ WIRED | Confirmed by direct code read: `len(incoming) > 0` refusal replaces the prior `!= 1` skip; every incoming edge is walked via `for _, edge := range incoming`. |
| `corevalidate.go` (rederive) | itself, recursively via `okEdgeInto` | acyclicity of the "ok"-edge chain | ✗ NOT_WIRED | No guard prevents infinite recursion/looping on a cyclic chain; see truth 6/gap. |
| `cgen.go` (EmitForeignHeader/EmitForeignConformance) | `session.go` (ScanForBannedAttributes) | zero-attribute scan over "all emitted C" | ✓ WIRED | Confirmed 8-argument call site; `recomputed_work=8` in the live gate run. |

### Data-Flow Trace (Level 4)

Not applicable — this phase's must-haves concern compiler-internal validation and code generation, not rendered UI data.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| Gap-1 falsifier tests pass | `go test ./internal/compiler/corevalidate/... -run 'TestMergeTerminalBlockDivergentReleaseSetsRefused\|TestTerminalBlockUnreachableRefused\|TestLegitimateDiscardMergeStillValidates' -v` | All 3 named tests PASS | ✓ PASS |
| Full corevalidate + session suites pass | `go test ./internal/compiler/corevalidate/... ./internal/compiler/session/...` | both `ok` | ✓ PASS |
| Full workspace suite | `go test ./...` (run once, via background gate script) | all packages `ok` | ✓ PASS |

### Probe Execution

| Probe | Command | Result | Status |
|-------|---------|--------|--------|
| Phase 4 gate | `sh scripts/verify-phase4.sh` | exit 0; all 14 lanes `"status":"pass"`; `lane:foreign-no-unproven-attributes` recomputed_work=8; all required control IDs present; 3 expected escapes present and never detected | PASS |

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|---|---|---|---|---|
| SEM-03 | 04-01, 04-04, 04-06, 04-07, 04-08 | Result propagation/ignored-result rules produce explicit typed control flow; panic/cancellation cannot be erased as ordinary errors | ✓ SATISFIED | OpFail-only typed-failure producer, defect terminator, cancelled reserved-unconstructible, all confirmed in code; REQUIREMENTS.md marks Complete, consistent with evidence found. Not affected by the new gap. |
| RES-01 | 04-02, 04-05, 04-07, 04-08 | Partially initialized noncopyable resources release exactly once in reverse completed-acquisition order on return and typed failure | ⚠️ PARTIALLY SATISFIED | The vacuous-skip gap (prior gap 1) is closed. The requirement's literal text (release order on return/typed-failure) holds for every reachable shape today. But the *independent, source-blind* rederivation this requirement's own decisions (D-04-07/D-12a) require is not robust against a corrupted core.Program — `rederive`'s unguarded cycle can hang the validator, which is exactly the adversarial-input class independent validation exists to defend against, per the file's own documented design principle. REQUIREMENTS.md's "Complete" marking is accurate for the requirement's literal text on honestly-produced programs, but the phase's own independence/robustness intent is not fully met. |
| FFI-01 | 04-01, 04-03, 04-05, 04-06, 04-07, 04-09 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, unwind obligations | ✓ SATISFIED | Contract and all three generated layers exist and carry every obligation; the D-04-13 zero-attribute enforcement control now covers all three inspectable layers (prior gap 2 closed), pinned at recomputed_work=8 so a future regression cannot silently shrink coverage. |

No orphaned requirement IDs found — SEM-03, RES-01, FFI-01 all appear in at least one plan's `requirements` field (04-08 and 04-09 added to their respective owning requirements) and all three appear in REQUIREMENTS.md mapped to Phase 4.

### Anti-Patterns Found

None of TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER found in the files touched by 04-08/04-09 (`corevalidate.go`, `corevalidate_test.go`, `session_test.go`) or elsewhere in the phase's file set per 04-REVIEW.md's own scan (48 files). No blocker-severity debt markers found. CR-01 (the cycle-guard gap) is a correctness/robustness defect, not a debt marker — recorded as a gap above, not an anti-pattern.

### Human Verification Required

None. The one remaining gap (CR-01, the missing cycle guard in `rederive`) is resolvable by direct code inspection and does not require human judgment to adjudicate — it is a concrete, well-scoped fix (add a visited-set, refuse on re-visit, add a falsifier) with a precedent pattern already present twice in the same file.

### Gaps Summary

Both previously-reported gaps are genuinely closed, confirmed by direct code inspection (not the summaries' claims alone):

1. **Prior gap 1 (RES-01/D-04-07) — CLOSED.** `checkReleaseOrder` no longer skips validation for a terminal block with an unexpected incoming-edge count; it now refuses (`core.release_order_indeterminate`) any terminal block with zero incoming edges and walks every incoming edge independently, comparing each rederived expectation against the fixed actual release list. The prior gap's second missing item (a structural peer check) is closed in an adapted form — `core.terminal_block_unreachable` (at-least-one/non-entry) rather than the literally-worded "exactly one" — a substitution judged sound because the literal form was tried and broke a legitimate language construct (`discard...because`'s two-incoming-edge merge), and the adapted form still closes the independence gap the missing item was about. This decision was auto-selected under auto-mode (recorded in 04-08-SUMMARY.md) rather than confirmed by a human; it is judged here as satisfying the gap's substance, but is flagged for awareness since it was not a human-reviewed decision.

2. **Prior gap 2 (FFI-01/D-04-12/D-04-13) — CLOSED.** `ScanForBannedAttributes` is now called with all eight arguments across both fixtures, including `EmitForeignHeader` and `EmitForeignConformance` output for each. A dedicated falsifier (`TestAttributeInjectionIntoConformanceOnlyMakesControlFail`) proves the conformance unit's own added text (not just its embedded header) is independently scanned, and a count-pin test (`TestAttributeScanLaneCoversEveryInspectableLayer`) ensures a future dropped argument cannot silently shrink coverage back down.

One new gap was found and confirmed independently this run:

3. **New finding (CR-01) — NOT CLOSED.** `corevalidate.checkReleaseOrder`'s backward-walk closure `rederive` has no cycle/visited-set guard, unlike the file's two other backward/forward graph walks (`blockReach`, `loanChainIndex.carriedLoans`), which both explicitly defend against cyclic/corrupted input with documented rationale ("a source-blind validator must stay defined against" adversarial input). A hand-corrupted `core.Program` with a cyclic "ok"-edge chain feeding a terminal block's incoming edge would hang `rederive`'s loop forever — an unbounded denial-of-service in the exact mechanism RES-01/D-04-07 requires to be independently sound. This gap pre-dates 04-08 (which added a different, unrelated check and did not touch `rederive`) and is untested (no cyclic-edge falsifier exists). It was not addressed by either gap-closure plan (04-08, 04-09), since neither plan's scope included it — it was only surfaced by this run's fresh code review pass.

**This looks like a genuine oversight, not a deliberate scope decision** — no record of it in 04-CONTEXT.md, 04-DEBT.md, or 04-VALIDATION.md, and the fix pattern (a visited-set, exactly matching two sibling walks already in the same file) is narrow and well-precedented. No override is suggested; recommend a small, targeted gap-closure plan: add the visited-set to `rederive`, propagate a `core.release_order_cyclic` refusal at both call sites, and add one falsifier constructing a two-block cyclic "ok"-edge chain.

---

_Verified: 2026-09-05_
_Verifier: Claude (gsd-verifier)_
