---
phase: 05-native-equivalence-and-adversarial-evidence
reviewed: 2026-09-06T18:57:55Z
depth: standard
files_reviewed: 53
files_reviewed_list:
  - cmd/lang/main.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_exclusive_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/core/core.go
  - internal/compiler/core/core_test.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_exclusive_test.go
  - internal/compiler/corevalidate/corevalidate_test.go
  - internal/compiler/execution/execution.go
  - internal/compiler/native/foreign_arena.go
  - internal/compiler/native/foreign_arena_test.go
  - internal/compiler/native/foreign_nonlocal.go
  - internal/compiler/native/foreign_retained.go
  - internal/compiler/native/foreign_retained_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/native/sanitize.go
  - internal/compiler/native/sanitize_test.go
  - internal/compiler/reduce/export_test.go
  - internal/compiler/reduce/mismatch.go
  - internal/compiler/reduce/mismatch_test.go
  - internal/compiler/reduce/mutationkill_test.go
  - internal/compiler/reduce/predicate.go
  - internal/compiler/reduce/predicate_test.go
  - internal/compiler/reduce/reduce.go
  - internal/compiler/reduce/reduce_test.go
  - internal/compiler/reduce/testdata/mismatch.golden.json
  - internal/compiler/session/qlt01.go
  - internal/compiler/session/qlt01_registry.json
  - internal/compiler/session/qlt01_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_alias_test.go
  - internal/compiler/session/session_phase5_compare.go
  - internal/compiler/session/session_phase5_compare_test.go
  - internal/compiler/session/session_phase5_corpus.go
  - internal/compiler/session/session_phase5_corpus_test.go
  - internal/compiler/session/session_phase5_escapes.go
  - internal/compiler/session/session_phase5_escapes_test.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_mismatch_test.go
  - internal/compiler/session/session_phase5_sanitize.go
  - internal/compiler/session/session_phase5_sanitize_test.go
  - internal/compiler/session/session_phase5_test.go
  - native/lang_foreign_arena.c
  - native/lang_foreign_retained.c
  - scripts/verify-phase5.sh
  - testdata/phase5/coordinated_lie.core.json
  - testdata/phase5/restrict_borrow.golden.c
findings:
  critical: 0
  warning: 2
  info: 1
  total: 3
status: issues_found
---

# Phase 5: Code Review Report

**Reviewed:** 2026-09-06T18:57:55Z
**Depth:** standard
**Files Reviewed:** 53
**Status:** issues_found

## Summary

This review covers the phase 5 diff (`4ca704b^..HEAD`) that adds the by-pointer C
lowering, the alias-fact/`restrict` justification machinery (`check`/`cgen`/
`corevalidate`), the retired `discoverLoanLastUses` liveness law, the isolated
ASan/UBSan sanitizer lane, two new hostile frozen foreign translation units, the
five-axis native/interpreter comparator, the core-level reducer plus
`lang.mismatch/0`, the QLT-01 control registry, the coordinated
source-to-core-false-claim escape demonstration, and the Phase 5 gate
(`scripts/verify-phase5.sh` / `Phase5RequiredControls`).

I read all 14 plan SUMMARYs before flagging anything, and cross-checked every
place that looked suspicious against its recorded design rationale before
concluding it was a genuine defect. Several things that look like bugs on
first read are deliberate, documented, and test-proven:

- The apparent double-counting of `len(body.Bindings)+1` in
  `analyzeStraightLine` (`internal/compiler/check/check.go`) is correct: a
  later, unrelated loop (`for index, binding := range body.Bindings { result.Work++ ... }`)
  supplies the third term `TestOwnershipWorkSeries` actually pins against
  (`wantWork := 1 + 2*(operations+1) + (operations+1)`). Verified by reading
  the full function body and re-running `TestOwnershipWorkSeries` — not a
  defect.
- `lang_foreign_retained.c`'s heap-use-after-free, `lang_foreign_arena.c`'s
  `operator new`/`operator delete` mismatch, and `coordinated_lie.core.json`'s
  disagreement with its own `.lang` source are all intentional adversarial
  subjects per the domain context and are not flagged here.
- The `Status: "fail"` string literal in `session_phase5_escapes.go` matches
  this codebase's own pre-existing `Lane.Status` convention (dozens of
  existing call sites in `session.go` use the same literal); it is not a new
  inconsistency.

The overall implementation is careful about the D-12 "independent derivation"
discipline (check/cgen/corevalidate never share helpers), fail-closed
sanitizer/tool-missing posture, and bounded subprocess I/O. The findings below
are narrower: two real warnings and two minor quality items, none of which
threaten correctness of the shipped gate.

## Warnings

### WR-01: Two structural helper functions in `reduce.go` are unreachable dead code

**File:** `internal/compiler/reduce/reduce.go:937-968`
**Issue:** `removeIndices` and `scrubBlocks` are fully-implemented package-level
functions with no call site anywhere in the `reduce` package (production or
test):

```
$ grep -rn "removeIndices\|scrubBlocks" internal/compiler/reduce/
internal/compiler/reduce/reduce.go:937:func removeIndices(ops []core.LinearOperation, indices []int) []core.LinearOperation {
internal/compiler/reduce/reduce.go:951:func scrubBlocks(blocks []core.Block, opIDs []string) []core.Block {
```
Go permits unused package-level functions (unlike unused locals/imports), so
this compiles silently. It reads as leftover scaffolding from an earlier
draft of `dropUnmatchedArm`/`dropOffpathForeignStage` (both of which ended up
using `removeOpsByID`/`removeBlock`/`removeEdgesTouching` instead) that was
never deleted. In a package whose own stated design goal is "no dead seam
left behind" (05-02-SUMMARY.md's own phrase, applied elsewhere in this phase
to the shadow-liveness apparatus and the reducer's own vacuity-kill seams),
this is a straightforward inconsistency with the project's own house rule.
**Fix:** Delete both functions (and confirm no future plan's helper depends on
them via `grep`).

### WR-02: `qlt01.go`'s registry-audit lane omits the two control IDs from `Controls` on failure, producing a redundant secondary diagnostic

**File:** `internal/compiler/session/qlt01.go:279-294`
**Issue:** `QLT01LaneFromRows` only populates `firedControls` (and therefore
`LaneResult.Controls`) in the success branch:

```go
status := "pass"
var firedControls []string
if len(failures) != 0 {
    status = "invalid"
} else {
    firedControls = []string{ControlQLT01RegistryIncomplete, ControlQLT01StaleControlReference}
}
```

When the audit finds a real violation (`len(failures) != 0`), `Controls`
stays `nil`. Back in `session_phase5.go`'s
`VerifyPhase5ControlsAndWork`, the final required-control sweep
(`for _, required := range Phase5RequiredControls() { if !hasControl(result.Lanes, required) { ... } }`)
will then ALSO report both `control:qlt01.registry_incomplete` and
`control:qlt01.stale_control_reference` as `verify.control_missing`, on top of
the lane's own `status: invalid` and its `QLT01AuditFailure` diagnostics. This
is not a false negative (the gate still correctly fails), but it means a
single real registry defect is reported as three or more overlapping
diagnostics with two different vocabularies (`control:qlt01.*` appearing both
as an audit-failure `Control` field and, separately, as a
`verify.control_missing` diagnostic naming the same identifier), which will
read confusingly to whoever triages a red gate. Every other Phase 5 lane in
`session_phase5.go` (e.g. the attribute-justification lanes, the alias lane)
reports its control ID in `Controls` on both its positive and negative
assertions (see `addLane("lane:attribute-unjustified", protocol.StatusPass, []string{"control:core.attribute_unjustified"}, ...)`), so this lane's asymmetry is
inconsistent with the pattern the rest of the file establishes.
**Fix:** Populate `Controls` with whichever of the two identifiers actually
fired (from the `fired` map already computed) regardless of overall pass/fail,
so `hasControl` finds the control via the lane itself rather than falling
through to the generic missing-control diagnostic.

## Info

### IN-01: `mismatchEvidenceID` and other content-digest constructions swallow `json.Marshal` errors silently

**File:** `internal/compiler/session/session_phase5_mismatch.go:135-148`
**Issue:**
```go
encoded, err := json.Marshal(identity)
if err != nil {
    return ""
}
```
`identity` embeds a `core.Program` value; if a future change to `core.Program`
introduces a field that cannot marshal (e.g. a `chan`, `func`, or an
unexported cyclic type reached through an interface), this silently returns
an empty `evidence_id` on a document that otherwise reports success — the
`EvidenceID` field would be a syntactically valid-but-empty string rather than
a build failure. Given `core.Program` is exhaustively JSON-tagged elsewhere in
this codebase, this is unlikely to trigger in practice, but the failure mode
is silent rather than fail-closed, which is inconsistent with this phase's
otherwise strict "never a silent pass" posture (`NewMismatchDocument` itself,
by contrast, correctly propagates a `json.Marshal` failure as an error for the
`reduced_core` field one function away in `mismatch.go`).
**Fix:** Either propagate the marshal error out of `ReduceSeededAliasMismatch`
(mirroring `NewMismatchDocument`'s own `reducedCoreBytes` handling immediately
downstream), or document explicitly why an empty `evidence_id` is an
acceptable degraded mode here.

---

_Reviewed: 2026-09-06T18:57:55Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
