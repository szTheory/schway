---
phase: 04-fallible-resources-and-c-boundary
verified: 2026-09-05T00:00:00Z
status: gaps_found
score: 5/7 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "`corevalidate` independently rederives the expected release order by walking backward from each failure edge over the block and edge graph, and compares — it never reads what `check` wrote and shares no helper with it (D-04-07, D-12a)." # 04-02 must_haves; underlies ROADMAP SC1
    status: failed
    reason: >
      Confirmed in code (not just per code review): `corevalidate.checkReleaseOrder`
      (internal/compiler/corevalidate/corevalidate.go:1232-1244) silently `continue`s
      past any terminal block (OpReturn- or OpFail-terminated) whose incoming-edge
      count is not exactly 1, instead of refusing the shape. Nothing else in
      `corevalidate` requires a terminal block to have exactly one incoming edge —
      `blocksAndEdges` (corevalidate.go:416-447) only requires an OpFail-terminated
      block to have "at least one" incoming edge and that every such edge carry
      Pattern "err"; it does not bound the count, and it does not apply to
      OpReturn-terminated blocks at all. So a merge-point terminal block (reachable
      from two acquisition chains with different completed-acquisition sets) would
      have its release order entirely unchecked by the one control D-04-07
      specifically requires to be an *independent* rederivation, while the checker's
      own generic per-operation replay would still pass. `check.go`'s own emission
      never happens to produce such a block today, so no shipped fixture exercises
      this path and no currently-compilable program is unsound — but the must_have
      as written ("independently rederives... and compares") is not true of the
      implementation without an unstated carve-out, and this is exactly the "control
      that could pass vacuously" failure mode this project's own commentary elsewhere
      (the `core.fail_reached_without_err_edge` comment, corevalidate.go:418-424)
      is written to guard against.
    artifacts:
      - path: internal/compiler/corevalidate/corevalidate.go
        issue: "checkReleaseOrder (line ~1242) skips validation for any terminal block whose incoming-edge count != 1, rather than refusing the shape; no other check bounds OpReturn-terminated (or OpFail-terminated) blocks to exactly one incoming edge"
    missing:
      - "Treat an unexpected incoming-edge count on a terminal block as a hard refusal in checkReleaseOrder, not a skip"
      - "Add a structural check (peer of the existing OpFail incoming-edge check in blocksAndEdges) that every OpReturn/OpFail-terminated block has exactly one incoming edge, independent of checkReleaseOrder itself"
  - truth: "One authoritative `core.ForeignContract` carries target layout, initialized state, allocator identity, capture and retention, aliasing, and unwind obligations, and all three inspectable layers are derived from it so no layer can invent a fact the contract does not carry (FFI-01, D-04-12, ROADMAP SC2)." # 04-03 must_haves; D-04-13's zero-attribute control is the enforcement mechanism for this inspectability claim
    status: failed
    reason: >
      Confirmed in code: `control:foreign.no_unproven_attributes` (D-04-13: "cgen
      emits zero optimizer-visible attributes... scanning all emitted C") is wired
      in session.go (lines 2183-2212) to call `cgen.ScanForBannedAttributes` on only
      four arguments: `cgen.Emit(positiveChecked.Program)`, its manifest,
      `cgen.Emit(releaseChecked.Program)`, and its manifest. `cgen.EmitForeignHeader`
      and `cgen.EmitForeignConformance` — two of D-04-12's three named inspectable
      layers, and the ones a human reviewer is most likely to actually read — are
      never passed to `ScanForBannedAttributes` anywhere in the repository (grep for
      `ScanForBannedAttributes` finds only the three session_test.go call sites,
      session.go:2204, none touching EmitForeignHeader/EmitForeignConformance
      output). A banned attribute injected into the generated header or the
      conformance unit's own added text would go undetected by the required
      control, even though the control's own required identifier
      (`control:foreign.no_unproven_attributes`) is asserted "pass" in every gate
      run. This is a control that names full coverage but does not scan the
      artifact it names.
    artifacts:
      - path: internal/compiler/session/session.go
        issue: "lane:foreign-no-unproven-attributes (line ~2178-2212) only scans cgen.Emit() and cgen.EmitForeignManifest() output, never cgen.EmitForeignHeader or cgen.EmitForeignConformance"
      - path: internal/compiler/cgen/cgen.go
        issue: "EmitForeignHeader (line ~1270) and EmitForeignConformance (line ~1327) are two of the three D-04-12 inspectable layers but are never fed into ScanForBannedAttributes in production or test code"
    missing:
      - "Add cgen.EmitForeignHeader(...) and cgen.EmitForeignConformance(...) output to the lane:foreign-no-unproven-attributes scan in session.go"
      - "Add a dedicated mutation-kill test injecting a banned token into EmitForeignHeader only (leaving emitLinearForeign's own extern line untouched) to prove the two artifacts are independently covered, since the existing 04-03 revert-and-fail demo happened to pass only because EmitForeignHeader's extern-declaration format string is byte-identical to emitLinearForeign's"
deferred: []
human_verification: []
---

# Phase 4: Fallible Resources and C Boundary Verification Report

**Phase Goal:** A noncopyable resource crosses one audited C boundary while partial
initialization, failure propagation, and cleanup remain defined.
**Verified:** 2026-09-05
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

Derived from ROADMAP.md success criteria and merged with PLAN frontmatter `must_haves` across all 7 plans (04-01 through 04-07).

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | SC1: A fallible two-step (in practice three-step) acquisition releases only initialized resources, exactly once, in reverse order on success and typed failure | ✓ VERIFIED (with a caveat, see #1a) | `check.go` materializes `OpRelease` into each failure block reverse of completed-acquisition order (D-04-07); `TestThreeAcquisitionReleaseOrder`, `TestPartialAcquisitionReleasesOnlyCompleted`, and the three-engine differential (`TestPhase4CorpusThreeEngineAgreement`) all pass on the shipped three-acquisition fixture family (`acquire_three_success.lang`, `acquire_three_fail_second.lang`, `acquire_three_fail_third.lang`). `sh scripts/verify-phase4.sh` reports `control:resource.release_order_transposed` and `control:resource.release_omitted` both pass with nonzero recomputed work (513 and 151 respectively). Mutation-kill register (04-VALIDATION.md) records both release mutations independently reverting production hunks and reproducing red. |
| 1a | The *independent* rederivation (`corevalidate.checkReleaseOrder`) proves release order without relying on `check`'s own bookkeeping, for every reachable terminal-block shape | ✗ FAILED | See gap 1 above. Confirmed by direct code read: the guard `if len(incoming) != 1 { continue }` (corevalidate.go:1242) silently skips validation for any terminal block that isn't a simple single-incoming-edge shape, and no other structural check bounds a return/fail-terminated block to exactly one incoming edge. Not reachable via any shipped fixture today (check.go never emits such a block), so no currently-compilable program is unsound, but the must_have as stated in 04-02's own frontmatter is not true of the implementation. |
| 2 | SC2: Generated C declarations and adapters make target layout, allocator identity, alias/capture, callback retention, and unwind policy inspectable | ✓ VERIFIED (existence/generation), see #2a for enforcement gap | `core.ForeignContract` (core.go) carries every named obligation; `EmitForeignHeader` generates the `_LANG_`-namespaced extern declaration, obligation comment block (generated from the same JSON, `TestObligationCommentsAreGeneratedFromJSON`), and self-layout `_Static_assert`s; `EmitForeignConformance` generates the separate, compiled-but-not-linked conformance TU with paired sizeof/_Alignof/offsetof assertions (`TestConformanceUnitAssertsEveryField`, `TestConformanceUnitCompilesSeparately`); `lang.foreign/0` sidecar manifest is digest-bound into `evidence.Manifest` via `foreign_digest` (omitempty). `control:foreign.layout_mismatch` passes with recomputed_work=1 (the layout-mutation mutation-kill row is green). |
| 2a | Zero optimizer-visible attributes are emitted, and the control scans all emitted C (D-04-13's own wording) | ✗ FAILED | See gap 2 above. Confirmed by direct code read: `lane:foreign-no-unproven-attributes` in session.go scans only `cgen.Emit()` (the compiled program) and `cgen.EmitForeignManifest()` output — never `cgen.EmitForeignHeader` or `cgen.EmitForeignConformance`, two of D-04-12's three named inspectable layers. `grep -rn ScanForBannedAttributes` across the repo confirms no call site anywhere passes header or conformance output to the scanner. A banned attribute injected into the generated header (the artifact a human reviewer is most likely to read) would not be caught by this required control despite the gate reporting `control:foreign.no_unproven_attributes: pass`. |
| 3 | SC3 first half: Panic cannot cross the ordinary non-unwinding C boundary | ✓ VERIFIED | `-fno-exceptions -fno-asynchronous-unwind-tables -fno-unwind-tables` compiled clean under `-Werror -pedantic`; `control:foreign.unwind_forbidden` (nm -u undefined-symbol allowlist, underscore-normalized) passes in `verify-phase4.sh` output with `expected_escapes` correctly excluding it from detected controls. `TestUndefinedSymbolAllowlistRejectsNewSymbol`, `TestMissingSymbolToolReportsOperational` both present. A foreign declaration lacking `unwind`/`nonlocal_exit` policy is refused independently by `check` and `corevalidate` with no default (`TestUnwindPolicyUndeclaredRejected`, verified by direct grep of both packages' independent refusal code). |
| 4 | SC3 second half: A foreign nonlocal exit cannot silently bypass Lang cleanup | ✓ VERIFIED | `native/lang_foreign_nonlocal.c` performs a real `longjmp`; `cgen.go` emits exactly one process-root `setjmp` landing pad plus a `static`-storage cleanup ledger (confirmed by direct grep: `setjmp.h`/`stdlib.h` includes, `lang_nonlocal_landing` jmp_buf). The pad emits `foreign.nonlocal_exit` plus one `resource.leaked` per still-live acquisition and terminates as a defect, running no release (`control:defect.no_release_on_defect` passes with recomputed_work=128; `DefectHasNoReleaseAfter` in session.go). D-04-18's "must not run releases" is honored — leaks are reported, not released, which is the deliberately PARTIAL disposition of RES-01 recorded plainly in 04-DEBT.md item 2 and in 04-DEBT.md's closing "Status of carried Phase 3 debt" section, not overstated anywhere as full release-on-nonlocal-exit. Two named, fenced blind spots (landing point below the pad; foreign process-exit) are recorded as `escape:nonlocal-exit-below-the-pad` and `escape:foreign-process-exit`, present in the gate's `expected_escapes` and never as a detected control — confirmed in the `verify-phase4.sh` JSON output captured above. |
| 5 | SC4: Interpreter and native executions agree on primary failure and cleanup events | ✓ VERIFIED | `TestPhase4CorpusThreeEngineAgreement` exists and is asserted in the Per-Task Verification Map (04-07-02) and the required-controls run; `sh scripts/verify-phase4.sh` (already run per "already established" and independently re-run here) exits 0 and reports every required control passing across all four lanes, including `lane:release-omitted`/`lane:release-order-transposed` (RES-01/SC1 path), `lane:nonlocal-exit-undetected` (SC3/SC4 path), and `lane:defect-signal-adjudicated`. `git diff <start>..HEAD -- testdata/phase1..3` empty, confirming no cross-engine regression was masked by a golden move. |

**Score:** 5/7 truths verified (2 of the 7 merged must-haves — the independent-rederivation half of SC1/RES-01, and the full-coverage half of SC2/FFI-01 — FAILED on direct code inspection, not merely per the code review's own claim).

### Deferred Items

None — both findings bear directly on this phase's own success criteria and are not deferred to a later phase in ROADMAP.md.

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/core/core.go` | `AllOperationKinds()`, `TerminatorKinds()`, `ForeignContract`, `OpForeignCall`/`OpRelease`/`OpFail`/`OpDefect` | ✓ VERIFIED | Present; six-site dispatch control (`control:kind.exhaustive_dispatch`) passes with recomputed_work=1524 |
| `internal/compiler/check/check.go` | foreign admission, reverse-order OpRelease materialization, discard-because | ✓ VERIFIED | Present; three-acquisition differential and discard fixture both green |
| `internal/compiler/corevalidate/corevalidate.go` | independent rederivation of release order and foreign refusals | ⚠️ PARTIAL | Foreign-call/unwind refusals independently derived and verified; release-order rederivation has the incoming-edge-count gap (gap 1) |
| `internal/compiler/cgen/cgen.go` | streaming emitter, landing pad, static ledger, zero-attribute emission, EmitForeignHeader/EmitForeignConformance | ⚠️ PARTIAL | All generation present and correct; the *scan* that is supposed to prove zero attributes across all three layers only covers one (gap 2) |
| `native/lang_foreign_resource.c`, `native/lang_foreign_resource_private.h`, `native/lang_foreign_nonlocal.c` | frozen, byte-identical foreign TUs | ✓ VERIFIED | `04-VALIDATION.md` sign-off confirms byte-identity; not independently re-hashed here since already covered by the "already established" pinning tests |
| `scripts/verify-phase4.sh` | bounded gate requiring every control by exact identifier | ✓ VERIFIED | Ran directly; exits 0; all 12 required negative controls present in output; 3 expected escapes present and never detected |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|----|--------|---------|
| `check.go` | `core.go` | ok/err edges, no stored Result value | ✓ WIRED | `OpFail` producer is the only route; grep confirms no `OpMakeErr`/`Err(...)` construct exists |
| `core.go` | `corevalidate.go` | independent release-order rederivation | ⚠️ PARTIAL | Wired for the single-incoming-edge case; unwired (skipped) for any other shape, per gap 1 |
| `cgen.go` | `session.go` | zero-attribute scan over "all emitted C" | ⚠️ PARTIAL | Wired for `Emit`/`EmitForeignManifest`; not wired for `EmitForeignHeader`/`EmitForeignConformance`, per gap 2 |
| `cgen.go` | `native.go` | streaming terminal record, last write, absence is hard failure | ✓ WIRED | `TestTerminalRecordAbsenceIsHardFailure`, `TestTruncationAndAbsenceReportDistinctCodes` present |
| `originvalidate`/`pathoracle` | `core.TerminatorKinds()` | walk every terminator, not just OpReturn | ✓ WIRED | `control:terminator.walk_incomplete` passes; D-04-29's registry-driven membership test confirmed present in session.go |

### Requirements Coverage

| Requirement | Source Plans | Description | Status | Evidence |
|---|---|---|---|---|
| SEM-03 | 04-01, 04-04, 04-06, 04-07 | Result propagation/ignored-result rules produce explicit typed control flow; panic/cancellation cannot be erased as ordinary errors | ✓ SATISFIED | OpFail-only typed-failure producer, defect terminator, cancelled reserved-unconstructible, all confirmed in code; REQUIREMENTS.md marks Complete, consistent with evidence found |
| RES-01 | 04-02, 04-05, 04-07 | Partially initialized noncopyable resources release exactly once in reverse completed-acquisition order on return and typed failure | ⚠️ PARTIALLY SATISFIED | Materialization, differential, and mutation-kill are correct and verified for the reachable shape; the *independent* validator half of this guarantee (D-04-07/D-04-12's explicit requirement) has the gap-1 hole. The nonlocal-exit path's deliberate non-release (reports leaks, D-04-18) is correctly scoped outside SC1/RES-01's own wording ("on return and typed failure") and is honestly recorded in 04-DEBT.md, not overstated. REQUIREMENTS.md's "Complete" marking is accurate for the requirement's literal text (return/typed-failure paths) but the phase's own must_have language for the independent proof is not fully met. |
| FFI-01 | 04-01, 04-03, 04-05, 04-06, 04-07 | Foreign contracts carry target layout, initialized state, allocator identity, capture/retention, aliasing, unwind obligations | ⚠️ PARTIALLY SATISFIED | Contract and all three generated layers exist and carry every obligation; the D-04-13 zero-attribute *enforcement* control (part of making obligations "inspectable" in a way that's actually checked) only covers 1 of 3 layers, per gap 2. |

No orphaned requirement IDs found — SEM-03, RES-01, FFI-01 all appear in at least one plan's `requirements` field and all three appear in REQUIREMENTS.md mapped to Phase 4.

### Anti-Patterns Found

None of TBD/FIXME/XXX/TODO/HACK/PLACEHOLDER found in the 24 files the code review already scanned (internal/compiler/{ast,cgen,check,core,corevalidate,evidence,execution,interp,native,originvalidate,pathoracle,session,syntax}, the two frozen C TUs, and verify-phase4.sh). No blocker-severity debt markers found.

### Behavioral Spot-Checks

Not re-run independently — `sh scripts/verify-phase4.sh` (Step 7c-equivalent probe) was re-executed directly in this verification session and confirmed exit 0 with every required control identifier present, matching the "already established" claim. This is the phase's own probe/gate script; running it directly (rather than trusting SUMMARY.md's narration of an earlier run) is the evidence basis for every ✓ VERIFIED item above that cites a `control:` identifier.

### Probe Execution

| Probe | Command | Result | Status |
|---|---|---|---|
| Phase 4 gate | `sh scripts/verify-phase4.sh` | exit 0; all lanes `"status":"pass"`; all 12 required control IDs present; 3 expected escapes present and never detected as controls | PASS |

### Human Verification Required

None. Both findings below are resolvable by direct code inspection (confirmed, not merely suspected) and do not require human judgment to adjudicate.

### Gaps Summary

Both critical findings from `04-REVIEW.md` (CR-01 and WR-01/CR-02) were independently confirmed against the current codebase, not accepted on the review's word alone:

1. **CR-01 confirmed.** `corevalidate.checkReleaseOrder` (corevalidate.go:1232-1260) silently skips its own independent release-order rederivation for any terminal block whose incoming-edge count isn't exactly 1. Direct read of `blocksAndEdges` (corevalidate.go:395-447) confirms no other structural check in the file bounds a return- or fail-terminated block to exactly one incoming edge — the existing check only requires "at least one" `err`-patterned edge into an `OpFail` block, and says nothing about `OpReturn`-terminated blocks or about upper-bounding the count. This is a genuine soundness gap in the control D-04-07/D-04-12 designed specifically to be independent of `check`'s own bookkeeping. It is not reachable through any program `check.go` currently emits (no fixture, corpus program, or the shipped binary's own compiler produces a merge-point terminal block today), so no currently-compilable Lang program has an actually-unverified release order — but the must_have as stated in 04-02-PLAN.md's own frontmatter ("independently rederives... and compares... never reads what check wrote") is not accurate to the implementation without noting this carve-out, and the gap is exactly the "control that could pass vacuously" class this project has repeatedly paid to close (03-08/03-09/03-10).

2. **WR-01/CR-02 confirmed.** `control:foreign.no_unproven_attributes` (session.go:2178-2212) is wired to scan only `cgen.Emit()` and `cgen.EmitForeignManifest()` output. A repo-wide grep for `ScanForBannedAttributes` confirms it is never called with `cgen.EmitForeignHeader` or `cgen.EmitForeignConformance` output anywhere — not in production code, not in `cgen_test.go`, not in `session_test.go`. Two of D-04-12's three named "inspectable layers" are therefore unscanned by the control whose entire job is to prove zero optimizer-visible attributes across "all emitted C." A banned attribute injected into the generated header (the artifact D-04-12 itself calls out as most likely to be read by a human reviewer) would pass this required control undetected, even though the gate reports it "pass."

Neither finding is a defect a currently-shipped Lang program can trigger — both are architectural gaps in the *independence* and *completeness* of validation machinery the phase's own decisions (D-04-07, D-04-12, D-04-13) explicitly required. Given that independence and completeness are the entire point of these two controls (the phase's own review brief and decisions repeatedly emphasize "a control that could pass vacuously" as the specific failure class to avoid), these are treated as gaps against the phase's own must_haves rather than passed on the strength of green tests that don't exercise the uncovered shape.

**This looks intentional in neither case** — both read as an oversight (a boundary condition not defended, and two of three generated artifacts not wired into an existing scanner call), not a deliberate scope decision recorded anywhere in 04-CONTEXT.md, 04-DEBT.md, or 04-VALIDATION.md. No override is suggested; recommend a small gap-closure plan addressing both (each is a narrowly-scoped fix: one added structural check plus a refusal-on-skip in corevalidate.go, and two additional scan arguments plus one new mutation-kill test in session.go/cgen.go).

---

_Verified: 2026-09-05_
_Verifier: Claude (gsd-verifier)_
