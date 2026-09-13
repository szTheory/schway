---
phase: 12-result-payloads
verified: 2026-09-13T00:00:00Z
status: gaps_found
score: 2/3 must-haves verified
behavior_unverified: 0
overrides_applied: 0
gaps:
  - truth: "Result layout has one meaning in the core IR, the interpreter, and emitted C17"
    status: partial
    reason: "CR-01 (12-REVIEW.md, CRITICAL, unresolved in the shipped tree as of this verification): nothing in check.go, core.NewDataType, or corevalidate.go refuses two alternatives of the same data type declaring the same PayloadType. cgen.alternativeNameForPayloadType (cgen.go:1939) and interp.alternativeNameForPayloadType (interp.go:839) both resolve an OpConstructPayload/OpDestructurePayload operation's PayloadType back to a declaring alternative name via an ambiguous first-match linear scan over AlternativeDetails, rather than via the operation's own TargetID/pattern. session.PayloadProbeDataType (session.go:612-620) already constructs exactly this shape (First(Byte), Second(Byte)) as a legal core.DataType, confirming the shape is reachable and accepted by every validator. D-12-12 makes payload-carrying alternatives a GENERAL data feature (not special-cased to the canonical Result fixture), so this is a live, reachable correctness hazard in the delivered feature, not a hypothetical. It does not corrupt the canonical Result fixture itself (Ok(Buffer)/Err(Fault) declare distinct payload types), but it directly contradicts the phase's own stated design principle that every other new payload hazard (arity mismatch, missing binder, binder-on-nullary, resource payload) gets an explicit fail-closed refusal — this one silently mis-resolves instead of refusing."
    artifacts:
      - path: "internal/compiler/cgen/cgen.go"
        issue: "alternativeNameForPayloadType (line 1939) resolves PayloadType -> alternative name by first-match linear scan; ambiguous when two alternatives share a PayloadType"
      - path: "internal/compiler/interp/interp.go"
        issue: "alternativeNameForPayloadType (line 839) has the identical ambiguity, independently implemented"
      - path: "internal/compiler/check/check.go"
        issue: "no declaration-time refusal exists for duplicate PayloadType values across a data type's alternatives (verified: no check.duplicate_payload_type or equivalent diagnostic in the tree)"
      - path: "internal/compiler/core/core.go"
        issue: "NewDataType (line 54) enforces name-set equality only; does not check PayloadType uniqueness"
      - path: "internal/compiler/corevalidate/corevalidate.go"
        issue: "the corpus-wide AlternativeDetails re-assertion (lines ~296-306) checks name presence/uniqueness only, not PayloadType uniqueness"
    missing:
      - "A fail-closed check.* diagnostic refusing two alternatives of one data type from declaring the same non-empty PayloadType (the code review's suggested fix), OR an AlternativeName fact carried directly on the operation so resolution never round-trips through PayloadType at all."
deferred: []
human_verification:
  - test: "Decide whether CR-01 (duplicate PayloadType ambiguity) should block phase closure or be accepted as scoped debt with a landing plan, and whether D-12-43 (the decisive slot-swap value-divergence control being empirically unconstructible against the current representation) is an acceptable terminal finding for criterion 2 as currently worded."
    expected: "Either CR-01 gets a committed fix/refusal and a re-verification, or the developer explicitly ratifies it as accepted debt (with a landing-phase entry in PHASE-12-DEBT.md, following the project's own mechanism for every other named hazard this phase deferred) via a VERIFICATION.md override."
    why_human: "This is a judgment call about whether an un-refused, review-flagged CRITICAL correctness gap in a general language feature is acceptable to ship now versus must gate phase closure — the project's own documented pattern (every other new payload hazard gets an explicit refusal) argues against silently accepting it, but the specific Result fixture this phase's roadmap goal names is unaffected."
---

# Phase 12: `Result` Payloads Verification Report

**Phase Goal:** A `Result` can carry a payload that is stored, matched, moved out of, and laid out identically in all three engines.
**Verified:** 2026-09-13
**Status:** gaps_found
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | A `Result` value with a payload-carrying alternative is stored and matched, and moving out of a matched payload obeys the affine drop obligation — the unmoved alternative's payload is still dropped exactly once | ✓ VERIFIED | `TestPayloadTracerThreeEngineAgreement` (cgen, PASS); `TestPayloadPatternRefusals`, `TestResourcePayloadRefused` (check, PASS); `payload_drop_obligation.lang` / `payload_borrow_interaction.lang` fixtures exercise the affine accounting at all six `TestAllOperationKindsHandledAtEverySite` dispatch sites (core, PASS). D-12-29 correctly scopes "dropped exactly once" to a static by-function-end accounting property (no `OpDrop` exists), and that property is derived independently by `check`'s worklist and `corevalidate`'s set-propagation. |
| 2 | `Result` layout — tagged union with niche optimization only where a checked ability fact permits — has one meaning in the core IR, the interpreter, and emitted C17, verified on the five-axis comparator | ✗ FAILED (partial) | The canonical `Result` fixture (`Ok(Buffer)`/`Err(Fault)`, distinct payload types) is verified: `PayloadRecordLayout` is derived once in `check` and read by `interp`/`cgen` (D-12-25); `TestPayloadLayoutMutationRefused` kills a transposed/resized struct at compile time (D-12-37); niche optimization is correctly recorded as uninstantiable, not silently skipped (D-12-23/24). But the general claim ("one meaning ... in all three engines") is undermined by CR-01 (12-REVIEW.md, CRITICAL, confirmed still unresolved in the shipped tree): two alternatives sharing one `PayloadType` resolve ambiguously via first-match linear scan in both `cgen` and `interp`, with no refusal anywhere in the pipeline. See gap above. Separately, D-12-38's decisive wrong-slot value-divergence control was built and run for real but empirically found unconstructible against the current representation (D-12-43) — honestly disclosed as an escalated criterion defect per the project's own D-12-41/D-11-36 precedent, not silently downgraded, but it means the "verified on the five-axis comparator" clause is proven only for the necessary-but-insufficient frozen-layout control, not for the decisive value-divergence one. Routed to human verification below rather than silently passed. |
| 3 | Gate (pre-flight probe): payload origin/ownership reduces to already-proven Phase 08-11 machinery; no hidden need for a new interprocedural rule | ✓ VERIFIED | `TestC03ResultPayloadOriginAcrossOpCall` and its companions (`corevalidate_result_payload_probe_test.go`), committed at `97a7c03`/run 2026-09-12 before planning, returned BRANCH A: both refusals attributable to the already-catalogued D-10-C01 gap, closable by filling an existing empty switch arm — no new `OperationKind`, no new dispatch site. Verified pre-existing in tree, ran green. |

**Score:** 2/3 truths verified (0 present-but-behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/cgen/cgen_n1_convergence_test.go` | N=1 convergence differential, permanent | ✓ VERIFIED | Exists, runs, PASS (`TestN1ConvergenceDifferential`, 5 fixtures) |
| `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` | Mechanically-checked debt register | ✓ VERIFIED | 8 items, `TestDebtRegistersAreWellFormed` PASS |
| `testdata/phase12/payload_tracer.lang` + `cgen_payload_tracer_test.go` | End-to-end payload round-trip fixture | ✓ VERIFIED | `TestPayloadTracerThreeEngineAgreement` PASS |
| `internal/compiler/check/check_payload_test.go` + refusal fixtures | Three named pattern refusals + resource refusal | ✓ VERIFIED | `TestPayloadPatternRefusals`, `TestResourcePayloadRefused` PASS |
| `internal/compiler/session/session_payload_replay_test.go` | Corpus characterization replay | ✓ VERIFIED | `TestPayloadCorpusCharacterizationReplay` PASS across full pre-Phase-12 corpus (byte-identical execution documents) |
| `internal/compiler/pathoracle/pathoracle_payload_test.go` | Path enumeration termination on payload fixtures | ✓ VERIFIED | Present, part of green `go test ./...` |
| `internal/compiler/session/session_payload_control_test.go` | D-12-37 frozen-fixture + D-12-38 slot-swap controls | ⚠️ ORPHANED FROM ITS OWN GOAL | `TestPayloadLayoutMutationRefused` PASS (necessary control works); `TestPayloadSlotSwapMutationKilled` PASS but by design measures and pins an absence — the decisive control could not be made load-bearing (D-12-43), disclosed honestly rather than hidden |
| `testdata/phase12/payload_layout_mismatch.golden.c` | Frozen mismatched-layout fixture | ✓ VERIFIED | Present, exercised by `TestPayloadLayoutMutationRefused` |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| Parser binder | `ast.MatchArm.Binder` → `check.checkBranch` → `core.OpDestructurePayload` → `interp` value → `cgen` flat-struct field | Full chain | ✓ WIRED | Traced via `TestPayloadTracerThreeEngineAgreement` passing end-to-end |
| `check`'s payload `RecordLayout` deriver | `cgen`'s flat-struct emission | Single shared fact (D-12-25) | ✓ WIRED | `PayloadRecordLayout` computed once in `check`, read by both consumers per code inspection and D-12-25's design |
| `cgen_n1_convergence_test.go` | `emitProgram` + legacy `Emit` path | Same checked program, both driven | ✓ WIRED | Confirmed by test run; this measurement is what produced the RE-DEFER ratification (D-12-36) |
| Operation's `PayloadType` | Declaring alternative name | `alternativeNameForPayloadType` (cgen, interp) | ✗ NOT SAFELY WIRED | Ambiguous when two alternatives share a `PayloadType` — see CR-01 gap above. Both consumers derive from the *same flawed helper independently reimplemented*, so the phase's own convergence tests structurally cannot catch this class of bug (per code review). |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|--------------|--------|----------|
| RES-02 | 12-02, 12-03, 12-04 | `Result` payload storage/matching, affine drop obligation | ✓ SATISFIED (with caveat) | Canonical fixture verified end-to-end; general `data` feature has the CR-01 gap above for the shared-payload-type edge case |
| RES-03 | 12-02, 12-04, 12-05 | Tagged-union layout, niche optimization where permitted, one meaning across engines | ⚠️ PARTIAL | Layout mechanism verified for the canonical case; niche correctly recorded as uninstantiable; "one meaning" claim undermined by CR-01; decisive value-divergence control (D-12-38) empirically unconstructible, disclosed as D-12-43 |

No orphaned requirements found — `.planning/REQUIREMENTS.md` maps only RES-02/RES-03 to Phase 12, both are marked `[x]`/`Complete`, and both are cited in plan frontmatter (`requirements:` fields across plans 02-05).

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| `internal/compiler/cgen/cgen.go` | 1939-1946 | Ambiguous first-match resolution, no refusal (CR-01) | 🛑 Blocker | Silent mis-resolution of payload alternative when payload types collide; contradicts phase's own fail-closed design principle |
| `internal/compiler/interp/interp.go` | 839-848 | Same ambiguous pattern, independently reimplemented | 🛑 Blocker (same root cause as above) | Same as above |
| `internal/compiler/check/check.go` | 3372-3414 | `WR-01`: diagnostic code/message conflates pattern-side vs. construction-side binder source | ⚠️ Warning | Misleading diagnostic message, not a correctness bug |
| `internal/compiler/cgen/cgen.go` | 60-75, 2115-2137 | `WR-02`: fault-injection seam silently no-ops when no alternative-mismatch target exists | ⚠️ Warning | Mutation-kill test could pass vacuously if fixture shape changes; no assertion catches this |
| `internal/compiler/cgen/cgen.go` | 1841-1848 | `IN-01`: local reimplementation of `lookupAlternativeDetail` | ℹ️ Info | Minor duplication, not a bug |

No unreferenced `TBD`/`FIXME`/`XXX` markers found in phase-touched files.

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|----------|---------|--------|--------|
| N=1 convergence differential | `go test ./internal/compiler/cgen/... -run TestN1ConvergenceDifferential -v -count=1` | 5/5 subtests PASS | ✓ PASS |
| Payload tracer three-engine agreement | `go test ./internal/compiler/cgen/... -run TestPayloadTracer -v -count=1` | PASS | ✓ PASS |
| Payload pattern refusals | `go test ./internal/compiler/check/... -run TestPayloadPatternRefusals -v -count=1` | 4/4 subtests PASS | ✓ PASS |
| Resource payload refusal | `go test ./internal/compiler/check/... -run TestResourcePayloadRefused -v -count=1` | 2/2 subtests PASS | ✓ PASS |
| Exhaustive dispatch control | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite -v -count=1` | PASS | ✓ PASS |
| Corpus characterization replay | `go test ./internal/compiler/session/... -run TestPayloadCorpusCharacterizationReplay -v -count=1` | All non-skipped fixtures PASS, byte-identical | ✓ PASS |
| Debt registers well-formed | `go test ./internal/compiler/session/... -run TestDebtRegistersAreWellFormed -v -count=1` | PASS (PHASE-12-DEBT.md included) | ✓ PASS |
| Layout mutation control (D-12-37) | `go test ./internal/compiler/session/... -run TestPayloadLayoutMutationRefused -v -count=1` | PASS | ✓ PASS |
| Slot-swap mutation control (D-12-38) | `go test ./internal/compiler/session/... -run TestPayloadSlotSwapMutationKilled -v -count=1` | PASS (measures/pins absence, per D-12-43) | ✓ PASS (as an honesty-preserving regression test, not as a "bug caught" claim) |
| Full workspace suite (run once) | `go test ./...` | All 25 packages `ok` (2 have no test files) | ✓ PASS |
| Duplicate-payload-type ambiguity check exists | `grep -rn "duplicate_payload_type" internal/compiler/` | No match | ✗ FAIL — confirms CR-01 is unresolved |

### Probe Execution

No `scripts/*/tests/probe-*.sh` shape probes apply to this phase; the phase's own "pre-flight probe" (criterion 3) is a committed Go test (`TestC03ResultPayloadOriginAcrossOpCall`), verified above under Observable Truths #3 and re-run as part of the full suite.

## Human Verification Required

### 1. CR-01 disposition — accept as debt or require a fix before closing the phase

**Test:** Read `12-REVIEW.md`'s CR-01 finding and this report's corresponding gap. Decide whether the compiler should refuse `data` declarations where two alternatives share a `PayloadType` (the review's suggested fix), or whether this is acceptable to carry forward as recorded debt.
**Expected:** Either a committed fix (small, scoped diagnostic addition in `check.go`) followed by re-verification, or an explicit override/debt-register entry accepting the gap with a named landing condition, consistent with how every other payload hazard this phase introduced (arity mismatch, missing binder, binder-on-nullary, resource payload) was handled — each got an explicit fail-closed refusal, not silence.
**Why human:** This is a scope/risk judgment call the project's own developer has made for every other comparable hazard in this same phase; it should be made explicitly here too rather than left implicit by a passing test suite that cannot see the gap (per the review, the convergence tests structurally cannot catch this class of bug, since `interp` and `cgen` share the identical flawed derivation).

### 2. D-12-43 acceptance as criterion 2's terminal finding

**Test:** Confirm the developer accepts `PHASE-12-DEBT.md`'s D-12-43 entry (the decisive slot-swap value-divergence control is empirically unconstructible against the current representation) as the terminal state for criterion 2's "verified on the five-axis comparator" clause, rather than requiring further harness/grammar work before phase closure.
**Expected:** An explicit ratification (the pattern this phase used at D-12-31/D-12-36's checkpoint), or a decision to slip the reopening work named in D-12-43 to a future phase.
**Why human:** This mirrors D-12-23/D-11-09's "absence-of-applicable-input" pattern, which the project has previously accepted as a terminal state — but it was reached by empirical measurement inside this same plan (05) without an explicit blocking-human checkpoint recorded for this specific finding (unlike D-12-31's RE-DEFER, which did get one). A human sign-off closes that gap.

## Gaps Summary

Phase 12's implementation is broad, careful, and its own generated evidence (34 named decisions, a pre-registered pre-flight probe, two independently-built mutation-testing controls, a full corpus characterization replay) is unusually strong. The full test suite is green across all 25 packages, and every task-level automated check in `12-VALIDATION.md` passes. Both binding human checkpoint decisions cited in this verification's brief were honored in the code as claimed: D-12-31 was re-deferred (the six legacy `cgen` emitters remain, byte-untouched, confirmed by `TestN1ConvergenceDifferential` and the `PHASE-12-DEBT.md` D-12-36 entry), and the ratified payload surface (`| Ok(Buffer)`, `Ok(v) =>`, the six diagnostic codes) matches the shipped parser/checker exactly.

The one real gap is CR-01, the code review's sole CRITICAL finding: `cgen` and `interp` both resolve a payload operation's declaring alternative via an ambiguous first-match linear scan with no refusal anywhere in the pipeline for the colliding-payload-type case that triggers it. This was flagged by the project's own review process, remains unresolved in the tree as of this verification, and directly contradicts the phase's stated design principle that every other new payload hazard gets an explicit fail-closed refusal. It does not corrupt the canonical `Result` fixture (`Ok(Buffer)`/`Err(Fault)` have distinct payload types) but does undermine the generality of criterion 2's "one meaning ... in all three engines" claim for the `data` feature this phase actually shipped (D-12-12: general feature, not special-cased to `Result`).

A secondary, honestly-disclosed finding (D-12-43) shows the decisive value-divergence mutation control required by criterion 2 was built and run for real but is empirically unconstructible against the current representation — a legitimate instance of this project's own escalate-rather-than-downgrade pattern, but one that was not run through an explicit human checkpoint the way the phase's other major branch decision (D-12-31) was.

---

_Verified: 2026-09-13_
_Verifier: Claude (gsd-verifier)_
