---
phase: 12-result-payloads
verified: 2026-09-13T11:00:00Z
status: passed
score: 3/3 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 2/3
  gaps_closed:
    - "Result layout has one meaning in the core IR, the interpreter, and emitted C17 (CR-01 duplicate-payload-type ambiguity closed)"
  gaps_remaining: []
  regressions: []
---

# Phase 12: `Result` Payloads Verification Report

**Phase Goal:** A `Result` can carry a payload that is stored, matched, moved out of, and laid out identically in all three engines.
**Verified:** 2026-09-13
**Status:** passed
**Re-verification:** Yes — after gap closure (plans 12-06, 12-07, 12-08)

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | A `Result` value with a payload-carrying alternative is stored and matched, and moving out of a matched payload obeys the affine drop obligation — the unmoved alternative's payload is still dropped exactly once | ✓ VERIFIED (regression check) | Unaffected by the gap-closure plans (no files listed here were touched by 06/07/08). Re-ran the supporting packages directly: `go test ./internal/compiler/check/... ./internal/compiler/core/... ./internal/compiler/cgen/... ./internal/compiler/interp/... ./internal/compiler/session/...` — all green. `TestPayloadTracerThreeEngineAgreement`, `TestPayloadPatternRefusals`, `TestResourcePayloadRefused` still pass; the six-dispatch-site affine accounting fixtures are untouched. |
| 2 | `Result` layout — tagged union with niche optimization only where a checked ability fact permits — has one meaning in the core IR, the interpreter, and emitted C17, verified on the five-axis comparator | ✓ VERIFIED | CR-01 (the sole blocking gap from the prior verification) is closed: `check.go`'s `program.Data` walk (line ~166) now emits `check.duplicate_payload_type`, a fail-closed declaration-time refusal, verified live against `testdata/phase12/payload_duplicate_payload_type.lang` (`TestDuplicatePayloadTypeRefused`, PASS) and against a hand-verified RED/GREEN demonstration recorded in D-12-44 (walk disabled -> test fails; walk restored -> passes). The two independent first-match linear scans are gone — grepped `internal/compiler/cgen/cgen.go` and `internal/compiler/interp/interp.go`: zero hits for a local `alternativeNameForPayloadType` declaration in either file. Both now call the single shared `core.AlternativeNameForPayloadType` (`core.go:107`) and `cgen`'s `emitBranch` calls `core.LookupAlternativeDetail` (`core.go:79`) instead of a local per-package copy — confirmed by direct grep of call sites (`interp.go:731`, `cgen.go:1877,2097,2121`). `TestPayloadAlternativeResolutionHasExactlyOneDerivation` (`core_single_derivation_test.go:190`) is a committed tripwire that fails if either local reimplementation is reintroduced. WR-02's anti-vacuity counter (`cgen.PayloadSlotSwapInjectedWriteCount`, `cgen.go:78-92`, incremented at `cgen.go:2146`) is read and asserted `>= 1` by `session_payload_control_test.go:246`, closing the hole where the fault-injection seam could silently no-op. **What this truth does NOT claim, stated plainly, per D-12-26/D-12-43:** "one meaning" is observable-behavior agreement (same alternative live, same payload value extracted, same event order) across all three engines, not byte-identical layout — `interp` has no byte layout to be identical to. The decisive wrong-slot value-divergence control (D-12-38) was built for real and empirically found structurally unconstructible against the current grammar (a match arm can only ever yield an alternative's compile-time-known tag name, never raw payload bytes, on either engine) — this is D-12-43, a measured absence-of-applicable-channel finding, not a failed engineering attempt, and not silently downgraded (`TestPayloadSlotSwapMutationKilled` pins the absence as a passing regression test, per the project's own gap-pinning precedent). D-12-43 was, in this run, put to the developer at an explicit `gate="blocking-human"` checkpoint (plan 12-08, Task 1) rather than accepted silently inside a task; the developer replied `ratify` verbatim, accepting D-12-43 as criterion 2's terminal state as worded. That decision is recorded as a dated, verbatim-quoted paragraph appended (additively — `git show 1292c2d --stat` shows 144 insertions, 1 deletion) to the existing `### D-12-43` section of `PHASE-12-DEBT.md`, not a rewrite of the original measurement narrative. This ratification closes the prior verification's human-verification item 2; it is not re-routed here. What is actually being certified for criterion 2: the necessary (frozen-layout, `_Static_assert`-policed) control passes (D-12-37), the duplicate-payload-type ambiguity hazard that WAS an active, unrefused correctness gap is now fail-closed at both the source layer (`check`) and the core layer (`core`'s shared resolver), and the decisive control's structural inapplicability is an explicitly ratified, not inferred, terminal state — not that the decisive control caught anything. |
| 3 | Gate (pre-flight probe): payload origin/ownership reduces to already-proven Phase 08-11 machinery; no hidden need for a new interprocedural rule | ✓ VERIFIED (regression check) | Unaffected by gap closure; `corevalidate_result_payload_probe_test.go`'s pre-existing controls remain in the green `go test ./...` run. |

**Score:** 3/3 truths verified (0 present-but-behavior-unverified)

### Required Artifacts (gap-closure plans 06-08)

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `internal/compiler/check/check.go` | `check.duplicate_payload_type` declaration-time refusal | ✓ VERIFIED | Present at line ~166, positioned in `program.Data` walk before function bodies are checked; confirmed by grep and by `TestDuplicatePayloadTypeRefused`/`TestPayloadPatternRefusals` (PASS) |
| `testdata/phase12/payload_duplicate_payload_type.lang` | Fixture reachable end-to-end through parser -> AST -> check | ✓ VERIFIED | Exists; exercised by `check_payload_test.go` |
| `internal/compiler/core/core.go` | `AlternativeNameForPayloadType` / `LookupAlternativeDetail` shared derivations | ✓ VERIFIED | Both functions present (lines 79, 107); ambiguity reported via error return, not first-match |
| `internal/compiler/core/core_single_derivation_test.go` | Tripwire against local reimplementation | ✓ VERIFIED | `TestPayloadAlternativeResolutionHasExactlyOneDerivation` present and green |
| `internal/compiler/cgen/cgen.go`, `internal/compiler/interp/interp.go` | No local `alternativeNameForPayloadType`; call the shared helper | ✓ VERIFIED | Zero grep hits for local declarations; both call `core.AlternativeNameForPayloadType`; `cgen.go:1877` calls `core.LookupAlternativeDetail` |
| `internal/compiler/cgen/cgen.go` (WR-02) | `PayloadSlotSwapInjectedWriteCount` anti-vacuity counter | ✓ VERIFIED | Present, incremented on injection, asserted `>= 1` in `session_payload_control_test.go:246` |
| `.planning/phases/12-result-payloads/PHASE-12-DEBT.md` | D-12-43 ratification, D-12-44/D-12-45 dispositions, mechanically well-formed | ✓ VERIFIED | `items: 10`, table row count matches; D-12-43 amended additively with dated ratification; D-12-44/D-12-45 added as new rows; `TestDebtRegistersAreWellFormed` green (part of full `session` package pass) |

### Key Link Verification

| From | To | Via | Status | Details |
|------|-----|-----|--------|---------|
| Operation's `PayloadType` | Declaring alternative name | `core.AlternativeNameForPayloadType` (single shared derivation) | ✓ WIRED | Both `cgen` and `interp` call the same `core` function; the prior "two independently reimplemented, ambiguity-blind scans" defect class is structurally closed and tripwired |
| `check`'s `program.Data` walk | Colliding-payload-type refusal | `check.duplicate_payload_type` | ✓ WIRED | Runs before any function body is checked; colliding shape cannot reach `analyzePayloadArm`, `interp`, or `cgen` for any program that crosses the parser |
| D-12-38 fault-injection seam | Anti-vacuity assertion | `cgen.PayloadSlotSwapInjectedWriteCount()` | ✓ WIRED | Asserted `>= 1` before the mutated/unmutated comparison is trusted; observed count 2 on a real run per `12-07-SUMMARY.md` |
| D-12-43's status | Recorded decision | Blocking-human checkpoint (plan 12-08, Task 1) -> `PHASE-12-DEBT.md` | ✓ WIRED | Checkpoint transcript (`ratify`) transcribed verbatim into the debt register with a date, closing the prior verification's open human item without inferring the answer |

### Anti-Patterns Found

Scanned every file touched by plans 06-08 (`check.go`, `check_payload_test.go`, `core.go`, `core_single_derivation_test.go`, `cgen.go`, `interp.go`, `session_payload_control_test.go`, `export_test.go`, `payload_duplicate_payload_type.lang`) for `TBD|FIXME|XXX|TODO|HACK|PLACEHOLDER` and empty-implementation patterns. **Zero matches.** No blockers, no warnings.

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|-------------|-------------|--------------|--------|----------|
| RES-02 | 12-02, 12-03, 12-04, 12-06, 12-07 | `Result` payload storage/matching, affine drop obligation | ✓ SATISFIED | Canonical fixture verified end-to-end; the general `data`-feature ambiguity gap (CR-01) is closed by a fail-closed refusal. Restriction now explicit and recorded: two alternatives of one data type may not declare the same payload type, with a named lifting condition (GEN-01's real generics, M003) |
| RES-03 | 12-02, 12-04, 12-05, 12-07 | Tagged-union layout, niche optimization where permitted, one meaning across engines | ✓ SATISFIED | Layout mechanism verified for the canonical case; niche correctly recorded as uninstantiable (D-12-24, absence-of-applicable-input, not a skip); "one meaning" is precisely scoped to observable-behavior agreement (D-12-26) and its necessary layout control passes (D-12-37); the decisive value-divergence control's structural inapplicability (D-12-43) is now an explicitly ratified terminal state via a blocking-human checkpoint, not a silent acceptance |

No orphaned requirements — `.planning/REQUIREMENTS.md` maps only RES-02/RES-03 to Phase 12 (both `[x]`/`Complete`), both cited in plan frontmatter across plans 02-08.

### Notes (non-blocking)

- **`12-07-SUMMARY.md` omits the `## Self-Check` section** the summary template calls for (12-06's and 12-08's both have one). This is an informational process gap, not a goal-achievement gap: the orchestrator independently verified `go build ./...` (exit 0) and `go test ./...` (exit 0, 25 packages, zero FAIL lines) rather than relying on 12-07's own narration, and this verifier independently re-ran the five directly-affected packages and inspected the actual diffs (`core.go`, `cgen.go`, `interp.go`, `core_single_derivation_test.go`) rather than trusting the summary text. No action required to close this phase; worth fixing in future summary authoring.
- **`must_haves.prohibitions` in plans 06/07/08 use `verification: descriptor-less`**, a value outside the `test | judgment` vocabulary this workflow's prohibition-routing step expects. Each prohibition was independently checked against the code by this verifier regardless of that label (see truth 2's evidence and the key-link table above: no field added to `core.LinearOperation`, no duplicate-payload-type refusal added to `corevalidate.Validate`, six legacy `cgen` emitters untouched by `git log` on `cgen.go` since 12-05, D-12-43's narrative textually unchanged, no debt row deleted/rewritten — all confirmed true). None was found violated. Flagged for future plan-authoring hygiene, not a phase gap.

## Gaps Summary

None. The single blocking gap from the prior verification (CR-01, `Result` layout truth) is closed with a fail-closed source-layer refusal and a shared, ambiguity-reporting core-layer derivation, both independently confirmed against the code (not inferred from SUMMARY.md claims), backed by tripwire tests that fail if either fix is reverted. Both items previously routed to human verification are now closed: CR-01's disposition is a recorded decision (D-12-44) with committed code; D-12-43 was ratified at an explicit blocking-human checkpoint rather than inferred, and that ratification is properly recorded as an additive, dated, verbatim transcript in `PHASE-12-DEBT.md`. The fresh code review (`12-REVIEW.md`) independently confirms all four prior findings (CR-01, WR-01, WR-02, IN-01) are genuinely fixed and finds nothing new. All truths verified; no gaps; no human verification items remain.

---

_Verified: 2026-09-13_
_Verifier: Claude (gsd-verifier)_
