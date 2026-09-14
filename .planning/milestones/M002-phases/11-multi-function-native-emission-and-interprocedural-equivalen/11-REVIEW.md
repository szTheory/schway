---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
reviewed: 2026-09-12T00:00:00Z
depth: standard
files_reviewed: 39
files_reviewed_list:
  - internal/compiler/cache/cache_test.go
  - internal/compiler/cache/probe.go
  - internal/compiler/cache/probe_test.go
  - internal/compiler/callgraph/callgraph.go
  - internal/compiler/callgraph/callgraph_entry_test.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/cgen/cgen_program.go
  - internal/compiler/cgen/cgen_program_test.go
  - internal/compiler/cgen/export_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate_opcall_rewrite_spike_test.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_lto_test.go
  - internal/compiler/reduce/export_test.go
  - internal/compiler/reduce/mutationkill_test.go
  - internal/compiler/reduce/reduce.go
  - internal/compiler/reduce/reduce_multifunction_test.go
  - internal/compiler/reduce/reduce_test.go
  - internal/compiler/session/export_test.go
  - internal/compiler/session/qlt03_shape_register.go
  - internal/compiler/session/qlt03_shape_register.json
  - internal/compiler/session/qlt03_shape_register_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_phase11_differential_test.go
  - internal/compiler/session/session_phase11_gate.go
  - internal/compiler/session/session_phase11_gate_test.go
  - internal/compiler/session/session_phase5.go
  - internal/compiler/session/session_phase5_alias.go
  - internal/compiler/session/session_phase5_corpus.go
  - internal/compiler/session/session_phase5_mismatch.go
  - internal/compiler/session/session_phase5_mismatch_test.go
  - internal/compiler/session/session_phase6.go
  - internal/compiler/session/session_phase6_cache_hole_test.go
  - internal/compiler/session/session_phase6_escapes.go
  - internal/compiler/session/session_phase6_pin_test.go
  - internal/compiler/session/session_phase6_verify.go
  - internal/compiler/session/session_phase7.go
  - internal/compiler/session/session_qlt05_reverify_test.go
findings:
  critical: 0
  warning: 4
  info: 3
  total: 7
status: issues_found
---

# Phase 11: Code Review Report

**Reviewed:** 2026-09-12
**Depth:** standard
**Files Reviewed:** 39 (plus 8 testdata/phase11/*.lang fixtures)
**Status:** issues_found

## Summary

This phase adds multi-function native (C) emission (`cgen_program.go`), a
single-resolver call-graph entry point (`callgraph.EntryFunction`), widens
`reduce.Reduce` to multi-function seeds with two new whole-program moves,
adds an eighth cache input (cgen's own source) closing a real, previously
reproduced staleness hole, and adds a large amount of verification
machinery (the Phase 11 mid-phase gate, the QLT-03 shape register, a
four-tier interprocedural differential, and an engineered LTO-exclusive
negative control). I traced the actual diff against the pre-Phase-11 tree
(`cfd0ba22b^..HEAD`) rather than reviewing every line of the pre-existing
2000+-line `cgen.go`/`native.go` files, since most of their bulk predates
this phase.

The verification machinery is, on the whole, honestly built: the mid-phase
gate's two knowers are genuinely independent derivations, the QLT-03
register's audit is mechanically driven off a real corpus sweep (not
hand-classified data pretending to be mechanical), and the LTO composition
control (`TestCompositionOnlyLTODivergence`) is a real, host-verified
divergence with an explicit anti-vacuity assertion that fails the build if
the matrix comes back all-green — consistent with this codebase's own
documented calibration for that file.

I did find one substantive robustness gap in the new `reduce` package: the
whole-program `drop-orphan-function` move relies entirely on a
caller-supplied `Seed.EntryFunctionID` to know which function must never be
deleted, and neither `Seed` nor `Reduce` validates that this ID is
non-empty or names a real function in the program. Given this project's own
repeatedly-stated "ambiguity always resolves to run it, never to skip it,
never guess" philosophy (see e.g. `cache.InputsFor`'s doc comment, and
`callgraph.EntryFunction`'s own fail-closed refusal), this is a real
inconsistency: the single production caller happens to derive the ID
correctly via `callgraph.EntryFunction` with error handling, but the
package itself has no defense against a future or test caller getting it
wrong, and getting it wrong means silently deleting the actual program
entry point during reduction rather than refusing. The rest of the
findings are smaller robustness/cleanliness items.

## Warnings

### WR-01: `reduce.Reduce` never validates `Seed.EntryFunctionID`, so a wrong or empty value silently lets `dropOrphanFunction` delete the real entry function

**File:** `internal/compiler/reduce/reduce.go:202-254` (Reduce), `internal/compiler/reduce/reduce.go:320-347` (dropOrphanFunction)

**Issue:** `dropOrphanFunction` treats `entryFunctionID` (the package-level
copy of `Seed.EntryFunctionID`, set once at the top of `Reduce`,
reduce.go:208) as the one function ID it must never delete. Nothing checks
that `Seed.EntryFunctionID` is non-empty, nor that it actually names one of
`Seed.Program.Functions`. If a caller passes a `Seed` with an empty
`EntryFunctionID` (the zero value) or a stale/misspelled one, and the real
entry function happens to have zero in-edges (which is definitionally true
of a program's entry point — nothing else in the program calls `main`),
`dropOrphanFunction` will happily delete it once its own logic reaches
that pass, because no function in the program has `ID == entryFunctionID`
in that scenario, so no function is ever exempted.

This is a real usage hazard by construction, not merely hypothetical:
`TestReduceRejectsMultiFunctionSeed` (`reduce_test.go`) — a test this very
phase's diff touches — calls `reduce.Reduce(ctx, reduce.Seed{Program:
seed}, ...)` on a genuine multi-function seed with `EntryFunctionID` left
at its zero value, and only asserts `err == nil`; it never inspects
`result.Program.Functions` to confirm the (arbitrary, in this case)
"entry" function survived. The one production caller
(`session/session_phase5_mismatch.go:386-390`) does derive the ID
correctly via `callgraph.EntryFunction` with error handling, so no
production path is affected today — but the package's own exported
contract has no fail-closed guard here, in a codebase whose other new
Phase 11 surface (`callgraph.EntryFunction` itself) is built specifically
around never guessing at "which function is the program."

**Fix:** Validate `Seed.EntryFunctionID` at the top of `Reduce` (or in a
small `Seed.Validate()` the way `core.Program` validators elsewhere in this
codebase work): refuse with a named error if it is empty or does not match
any `Seed.Program.Functions[i].ID`, mirroring
`callgraph.entryAmbiguousError`'s own fail-closed shape rather than
silently proceeding. For example:

```go
func Reduce(ctx context.Context, seed Seed, interesting Predicate) (Result, error) {
    if interesting == nil {
        return Result{}, fmt.Errorf("reduce: interesting predicate must not be nil")
    }
    if len(seed.Program.Functions) > 1 {
        found := false
        for _, fn := range seed.Program.Functions {
            if fn.ID == seed.EntryFunctionID {
                found = true
                break
            }
        }
        if !found {
            return Result{}, fmt.Errorf("reduce: multi-function seed requires a valid Seed.EntryFunctionID, got %q", seed.EntryFunctionID)
        }
    }
    ...
```

### WR-02: `reduce.entryFunctionID` is unsynchronized package-level mutable state

**File:** `internal/compiler/reduce/reduce.go:136-146, 207-208`

**Issue:** `entryFunctionID` is a package-level `var`, written at the top of
every `Reduce` call and read from `dropOrphanFunction` deep inside the
move-application loop. The doc comment concedes this directly: "Reduce is
not re-entrant or concurrent by design ... so a single package-level value
set at the top of Reduce and read here is safe." That's an invariant
enforced by nothing — no mutex, no `sync/atomic`, no doc-level warning on
the exported `Reduce` function itself (the warning is buried on the
unexported `var`). If any future caller (or a test using `t.Parallel()`)
invokes `Reduce` concurrently from two goroutines, `entryFunctionID` from
one call can silently leak into another call's `dropOrphanFunction`
decisions — a data race `go test -race` would catch, but only if a
concurrent-call test is ever written, which none of the current tests are.

**Fix:** Thread `entryFunctionID` through as an explicit parameter to
`dropOrphanFunction` (it is the only one of the seven moves that needs
it) rather than a shared package var, e.g. change `Move.Apply`'s call
site for this one move, or wrap `dropOrphanFunction` in a closure captured
per-`Reduce`-call:
```go
moves := movesToApply()
// build a per-call closure so no package-level var is needed
```
At minimum, document the non-reentrancy requirement directly on the
exported `Reduce` function's doc comment, not only on the private `var`.

### WR-03: `emitBranchOperations`'s `OpCall` arm omits the double-target-write guard the sibling `OpCall` arms all have

**File:** `internal/compiler/cgen/cgen.go` (emitBranchOperations, OpCall case, ~line 1849-1866 per the diff)

**Issue:** `emitLinear`'s `OpCall` arm and `emitProgramFunction`'s `OpCall`
arm (`cgen_program.go:401-411`) both check
`target, exists := places[operation.TargetID]; if !exists ||
declared[operation.TargetID] { return error }` before emitting the call —
catching both an invalid target place and re-use of an already-declared
target. `emitBranchOperations`'s new `OpCall` arm only checks `!exists`:
```go
target, exists := places[operation.TargetID]
if !exists {
    return fmt.Errorf("operation %q has invalid target", operation.ID)
}
```
It has no `declared` map at all, so a hypothetical malformed/forged
`core.Program` with two `OpCall` operations both targeting the same place
inside a match arm would silently emit two C declarations of the same
local identifier (a C compile error at best, or — if names differ due to
allocator quirks — silently wrong code) instead of being refused the same
way the other two call sites are. This arm is documented as "provably
unreachable in production" (single-function branch bodies can never
legally contain an `OpCall`), which is the calibration reason this is a
WARNING and not a BLOCKER — but the whole point of `emitCall` being "the
single writer of a Lang-to-Lang call anywhere in cgen" (D-11-04) is
undermined if its three call sites don't apply the same input validation
before calling it, since a later change that makes this arm reachable
(e.g. multi-function branch bodies, explicitly named as a "documented
scope limit for a later phase" in `cgen_program.go`) would silently inherit
weaker validation than its siblings.

**Fix:** Add the same `declared[operation.TargetID]` check
`emitLinear`/`emitProgramFunction` use, for consistency and defense in
depth, even though it is currently unreachable:
```go
target, exists := places[operation.TargetID]
if !exists || declared[operation.TargetID] {
    return fmt.Errorf("operation %q has invalid target", operation.ID)
}
```
(This requires threading a `declared` map into `emitBranchOperations`,
which it does not currently maintain for any operation kind — worth
checking whether the grouped `OpCopy/OpMove/OpBorrow*` arm above it has
the same gap.)

### WR-04: Two structural helpers in `reduce.go` are dead code

**File:** `internal/compiler/reduce/reduce.go:1320-1351` (`removeIndices`, `scrubBlocks`)

**Issue:** `removeIndices` and `scrubBlocks` are defined but never called
anywhere in the package (production or test) as of this diff. They compile
silently (Go only errors on unused imports/locals, not unused
package-level functions), so this is easy to miss in review. Both look
like earlier iterations of `removeOpsByID`/`removeBlock`
(which ARE used) that were superseded but never deleted.

**Fix:** Delete `removeIndices` and `scrubBlocks` if genuinely unused, or
wire them in if they were meant to replace `removeOpsByID`/`removeBlock`.
`go vet`/`staticcheck -unused` (or `deadcode`) would catch this
automatically in CI.

## Info

### IN-01: Phase 11 mid-phase gate's fourth lane is a decorative always-pass conjunct

**File:** `internal/compiler/session/session_phase11_gate.go:217-227`

**Issue:** `report.ManifestEmptyAttributes = true` is set unconditionally,
and `lane:phase11-manifest-consistency` always reports
`protocol.StatusPass`. The doc comment is honest about this ("Asserted
here purely as a sanity cross-check; explicitly NOT the gate's second
knower"), so this is not a hidden vacuity — but it does mean
`Phase11GateReport.Result` carries one lane that can never contribute a
failure signal, which a future reader skimming lane counts (e.g. "4/4
lanes passed") could mistake for four independent live checks. Purely
informational since the intent is disclosed in the comment; no fix
required beyond, optionally, naming it something less lane-shaped (e.g.
folding it into the report struct's doc rather than `result.Lanes`) so a
lane-counting consumer doesn't inflate the gate's apparent conjunct count.

### IN-02: `native.go`'s widened `function.returned` shape check relies entirely on other event kinds' `isLast` exclusions to keep the terminal event actually terminal

**File:** `internal/compiler/native/native.go:592-615`

**Issue:** Pre-Phase-11, `function.returned` was required to be the last
event (`!isLast` was rejected). Phase 11 removes that requirement to allow
multiple `function.returned` events (one per popped call frame). The
document's actual last event is still constrained to be one of
`function.returned` / `function.failed` / `function.defected` only because
every OTHER event kind's own case explicitly rejects being last
(`value.copied`/`transferred`/`borrowed*`/`foreign.called`,
`resource.leaked`, `resource.released` all reject `isLast == true`). This
is correct today, but it is an implicit invariant spread across six
unrelated `case` arms rather than a single explicit assertion — a future
event kind added without also adding an `isLast` exclusion would silently
become a legal terminal event. Pre-existing design pattern, not introduced
by this diff, but Phase 11 is the change that removed the one arm
(`function.returned`) that used to make the invariant more visible.

**Fix (optional, low priority):** Add a single explicit check after the
loop: `lastEvent := value.Events[len(value.Events)-1]; if
!isTerminalEventKind(lastEvent.Kind) { return errors.New(...) }`, making
the "last event must be a terminal kind" rule a first-class assertion
instead of an emergent property of six other arms' exclusions.

### IN-03: `dropOrphanFunction`'s exemption is by `fn.ID == entryFunctionID` string equality only, with no defense against an accidentally-empty ID matching an accidentally-empty function ID

**File:** `internal/compiler/reduce/reduce.go:335-341`

**Issue:** Related to WR-01: even setting aside the missing top-level
validation, the exemption check itself (`if fn.ID == entryFunctionID {
continue }`) offers no diagnostic if `entryFunctionID` is `""` and,
hypothetically, a forged/corrupted `core.Program` contained a function
with an empty `ID` — that function would be (wrongly) treated as the
protected entry, while the real entry function (non-empty ID) would remain
eligible for deletion. This is a narrower restatement of WR-01's root
cause and would be closed by the same fix; listed separately only because
it is a second, more subtle way the missing validation can misfire beyond
the "no match at all" case in WR-01's example.

---

_Reviewed: 2026-09-12_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
