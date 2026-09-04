---
phase: 03-borrowed-views-and-cfg-lifetimes
reviewed: 2026-09-04T00:00:00Z
depth: standard
files_reviewed: 60
files_reviewed_list:
  - cmd/lang/main.go
  - internal/compiler/ability/ability.go
  - internal/compiler/ability/ability_test.go
  - internal/compiler/ast/ast.go
  - internal/compiler/cgen/cgen.go
  - internal/compiler/cgen/cgen_names_test.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_branch_test.go
  - internal/compiler/check/check_exclusive_test.go
  - internal/compiler/check/check_origin_test.go
  - internal/compiler/check/check_test.go
  - internal/compiler/check/check_unexecutable_test.go
  - internal/compiler/core/core.go
  - internal/compiler/corevalidate/corevalidate.go
  - internal/compiler/corevalidate/corevalidate_branch_test.go
  - internal/compiler/corevalidate/corevalidate_endpoint_internal_test.go
  - internal/compiler/corevalidate/corevalidate_endpoint_test.go
  - internal/compiler/corevalidate/corevalidate_exclusive_test.go
  - internal/compiler/corevalidate/corevalidate_quadratic_test.go
  - internal/compiler/debugmap/debugmap.go
  - internal/compiler/debugmap/debugmap_test.go
  - internal/compiler/evidence/evidence_test.go
  - internal/compiler/interp/interp.go
  - internal/compiler/native/native.go
  - internal/compiler/native/native_test.go
  - internal/compiler/originvalidate/originvalidate.go
  - internal/compiler/originvalidate/originvalidate_test.go
  - internal/compiler/pathoracle/pathoracle.go
  - internal/compiler/pathoracle/pathoracle_test.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/protocol/protocol_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_borrow_conflict_test.go
  - internal/compiler/session/session_branch_test.go
  - internal/compiler/session/session_exclusive_test.go
  - internal/compiler/session/session_test.go
  - internal/compiler/syntax/format.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/syntax_test.go
  - internal/compiler/syntax/token.go
  - internal/compiler/testsupport/cli_test.go
  - internal/compiler/testsupport/testsupport.go
  - internal/compiler/testsupport/testsupport_internal_test.go
  - scripts/verify-phase3.sh
  - testdata/phase2/evidence.golden.json
  - testdata/phase2/owned_transfer.golden.c
  - testdata/phase3/borrowed_view.lang
  - testdata/phase3/branch_one_arm_shared_accept.lang
  - testdata/phase3/branch_one_arm_shared_reject.lang
  - testdata/phase3/branch_view.lang
  - testdata/phase3/exclusive_exclusive_reject.lang
  - testdata/phase3/exclusive_move_reject.lang
  - testdata/phase3/public_view.lang
  - testdata/phase3/public_view_impossible.lang
  - testdata/phase3/public_view_mixed_access.lang
  - testdata/phase3/public_view_omitted.lang
  - testdata/phase3/public_view_understated.lang
  - testdata/phase3/sequential_shared_then_exclusive_accept.lang
  - testdata/phase3/shared_exclusive_reject.lang
  - testdata/phase3/shared_shared_accept.lang
findings:
  critical: 1
  warning: 3
  info: 2
  total: 6
status: issues_found
---

# Phase 03: Code Review Report (re-review)

**Reviewed:** 2026-09-04T00:00:00Z
**Depth:** standard
**Files Reviewed:** 60 (42 Go/shell source+test files plus 18 `testdata/phase2`/`testdata/phase3` fixtures)
**Status:** issues_found

## Summary

This is a re-review of a phase that already went through one review cycle
(`03-REVIEW.md` at commit `acaa380`, one Critical finding — `CR-01`) and two
gap-closure plans (`03-08`, `03-09`, commits `e049c11..HEAD`) that both
touched `internal/compiler/originvalidate/originvalidate.go`. I first
verified the prior review's disposition, then focused the deepest scrutiny
on the newly changed code per the task's own instruction that it is "the
least-reviewed part of the phase."

**The original `CR-01` is fixed correctly.** `RecomputeOrigin`'s
`OpBorrowExclusive` branch is now guarded first-seen
(`originvalidate.go:88-91`), symmetric with the pre-existing
`OpBorrowShared` guard, and `public_view_mixed_access.lang` plus
`TestMixedAccessChainDerivesShared` / `TestMixedAccessChainRejectedAsAccessMismatch`
correctly pin the specific reborrow-chain shape the original finding
described. `03-09`'s omitted-origin gate (`core.origin_omitted`) is also
correctly wired for the single-arm, single-return shape it was built and
tested against (`public_view_omitted.lang`).

**However, both fixes share an unexamined assumption that does not hold in
general: `RecomputeOrigin` assumes a function's `core.LinearBody.Operations`
contains exactly one `OpReturn`.** That assumption is true for every
straight-line function (Linear-only, non-match) and for every fixture
either gap-closure plan added, but it is false for a match-arm-bodied
("branch-shaped") function — `check.go`'s own arm-body lowering appends one
`OpReturn` per arm into the *same* flat `Operations` slice (confirmed by
reading `check.go`'s branch-body constructor, lines ~200-365, and by
`corevalidate.go:810`'s own comment: "a branch-shaped function has one
return PER BLOCK"). `RecomputeOrigin` records only the *first* `OpReturn` it
encounters while scanning `Operations` in order and silently ignores every
other arm's return entirely. I confirmed this is not merely a theoretical
gap in the algorithm but a live, reachable defect through the actual
compiler pipeline and the shipped binary — see CR-01 (renumbered) below.
None of the phase's match/branch fixtures (`borrowed_view.lang`,
`branch_view.lang`, `branch_one_arm_shared_*.lang`) exercise this because
none of them return a live borrow from a non-first arm, and none of the new
origin fixtures (`public_view_mixed_access.lang`, `public_view_omitted.lang`)
are match-bodied, because `check.go`'s scope fence
(`ownership.match_borrowed_return_unsupported`) forbids a match function
from *declaring* an origin — but nothing forbids a match arm's body from
*returning a live borrow anyway*, which is exactly the shape `03-09`'s
`core.origin_omitted` gate exists to catch and, for this shape, does not.

Everything else previously flagged (`WR-01`, `WR-02`, `WR-03`, `IN-01`,
`IN-02`) is unchanged by the gap-closure plans and still applies as
originally described; I re-verified each against the current tree rather
than assuming carry-forward. D-03-01 and D-03-02 (D-03-02 now closed per
`03-DEBT.md`) remain out of scope for this review as before.

## Critical Issues

### CR-01: `originvalidate.RecomputeOrigin` only inspects a function's first `OpReturn`, so a match-arm-bodied function that returns a live borrow from any arm other than the first silently escapes both the access-mismatch and omitted-origin gates

**File:** `internal/compiler/originvalidate/originvalidate.go:59-98` (the return-selection and backward-walk loop), consumed unconditionally by `ValidatePublished` (`originvalidate.go:119-144`)

**Issue:** `RecomputeOrigin` builds `returnOp` from the first operation with
`Kind == core.OpReturn` encountered while iterating `function.Linear.Operations`
in index order, and every subsequent `OpReturn` is skipped (`continue`)
without ever being considered:

```go
for index := range operations {
    operation := operations[index]
    if operation.Kind == core.OpReturn {
        if returnOp == nil {
            returnOp = &operations[index]
        }
        continue
    }
    sourceOf[operation.TargetID] = operation
}
```

For a straight-line function this is correct — there is exactly one
`OpReturn`. For a match-arm-bodied ("branch") function, `check.go`'s arm
lowering (the constructor around lines 200-365, confirmed by
`corevalidate.go:810`'s explicit comment "a branch-shaped function has one
return PER BLOCK") appends one `OpReturn` per arm into the *same* flat
`function.Linear.Operations` slice, in arm order. `RecomputeOrigin` derives
its whole answer from arm 0's return chain alone and never looks at arm 1's
(or any later arm's).

`check.go` forbids a match-bodied function from *declaring* a
`ReturnOrigin` (`ownership.match_borrowed_return_unsupported`,
`check.go:83-87`), but this fence only blocks the declaration — it does not
stop an arm's body from returning a place that is a live borrow of the
parameter with no declaration at all. That is precisely the shape
`03-09`'s `core.origin_omitted` gate was built to close (for the
straight-line case). Because `RecomputeOrigin` only examines the first arm,
a function whose first arm returns owned and whose *second* arm returns a
live borrow evades the gate entirely, and a function whose first arm
returns one access mode and second arm returns a different one similarly
evades `core.origin_access_mismatch`-style detection (moot here since
origin can never be declared on a match function, but it means the
`origin_omitted` gate's "does this body actually leak a borrow" check is
simply wrong for every branch-shaped function with more than one arm).

I reproduced this end-to-end through the real pipeline and the shipped
binary, using an honest, unmutated fixture — no falsifying mutation was
needed, exactly the pattern that made the original `CR-01` and `D-03-02`
findings serious:

```
module owned.branch_origin_leak

export {
  type Switch
  fn choose
}

data Switch =
  | On
  | Off

fn choose(flag: Switch) -> Switch {
  match flag {
    On => {
      let moved = take flag
      moved
    }
    Off => {
      let view = borrow flag
      view
    }
  }
}
```

- `lang --json check` on this source: `"status":"pass"`, zero diagnostics.
- `lang --json interface export` on this source: `"status":"pass"`, exit 0,
  and the written summary's `choose` function carries **no** `public_origin`
  field at all — it is indistinguishable from a fully-owned function, even
  though the `Off` arm hands the caller a live, unreleased borrow of `flag`.
- Calling `originvalidate.ValidatePublished` directly on the validated
  `core.Program` for this exact source returns an **empty** problem slice.

This is a live escape of the publication guarantee `03-09` was written to
establish (ROADMAP SC4's "omitted" clause), reachable by an honest producer
writing ordinary match-arm code, not requiring any adversarial mutation —
the same severity class as the original `CR-01`, and specifically located
in the code both gap-closure plans touched without anyone tracing what
"first `OpReturn`" means once a function has more than one.

No fixture under `testdata/phase3/` exercises this: the phase's only
match/branch fixtures (`borrowed_view.lang`, `branch_view.lang`,
`branch_one_arm_shared_accept.lang`, `branch_one_arm_shared_reject.lang`)
all return an owned (`take`d) value from every arm, and the phase's only
origin fixtures (`public_view.lang`, `public_view_understated.lang`,
`public_view_impossible.lang`, `public_view_mixed_access.lang`,
`public_view_omitted.lang`) are all single-arm straight-line functions.
`scripts/verify-phase3.sh`'s nine required controls do not cover this shape
either.

**Fix:** `RecomputeOrigin` needs to reason over *every* `OpReturn` in the
function, not just the first, and combine the per-return answers
conservatively (e.g., if any return derives a borrow with no matching
declared origin, that is enough to trigger `core.origin_omitted`; if
returns disagree on access mode, that's at least as bad as a single
mismatching one and must not silently resolve to whichever arm happened to
be checked). A minimal fix:

```go
var returnOps []*core.LinearOperation
for index := range operations {
    operation := operations[index]
    if operation.Kind == core.OpReturn {
        returnOps = append(returnOps, &operations[index])
        continue
    }
    sourceOf[operation.TargetID] = operation
}
if len(returnOps) == 0 {
    return nil, "", false
}
// walk backward from EVERY returnOp, not just the first, and require
// derivedAccess (and the derived path set) to agree across all of them;
// disagreement or a mix of "owned" and "borrowed" arms must not resolve
// to a single confident answer that hides the disagreement.
```

Add a regression fixture pairing an owned-returning arm with a
borrow-returning arm (the `branch_origin_leak` shape demonstrated above,
or the mixed-access variant with two arms disagreeing on access mode), and
wire a new required control into `scripts/verify-phase3.sh` mirroring
`control:origin.omitted_summary` for the branch-bodied case specifically —
the existing `control:origin.omitted_summary` control only exercises the
single-arm shape and would not have caught this.

## Warnings

_(Carried forward from the prior review at `acaa380`; re-verified against
the current tree — none were touched by the `03-08`/`03-09` gap-closure
plans, and all still apply as originally described.)_

### WR-01: `check`-time programs carry an unverified `PublicOrigin` claim; only `interface export` verifies it

**File:** `internal/compiler/check/check.go:971-974`, `internal/compiler/session/session.go` (`Check`, `RunNative`, `RunInterpreter`)

**Issue:** `checkLinear` builds `core.PublicOrigin` directly from the AST's
parsed `ReturnOrigin` declaration with no body verification. `session.Check`,
`RunInterpreter`, and `RunNative` (the paths behind `lang check` and
`lang run`) never call `originvalidate.ValidatePublished`; only
`InterfaceExportCommandFile` does. A source file can declare a fictitious
origin and `lang check` / `lang run` will accept and execute it without
complaint until someone separately runs `lang interface export`. Still true
on the current tree — `PublicOrigin` remains unread by `corevalidate.go`,
`cgen.go`, and `interp.go`.

**Fix:** As previously suggested — add a doc comment on `core.PublicOrigin`
(or `check.Program`/`session.Check`) stating explicitly that the field is
unverified until `originvalidate.ValidatePublished` runs.

### WR-02: `corevalidate.cloneProgram` panics instead of returning an error

**File:** `internal/compiler/corevalidate/corevalidate.go:1140-1150` (confirmed unchanged: `panic(fmt.Sprintf("core validation clone: %v", err))` appears twice, lines 1143 and 1147)

**Issue:** `cloneProgram` (used by `Validate` on every call, including every
`check`/`run`/`interface`/`debug-map`/`verify` invocation) panics if
`json.Marshal`/`json.Unmarshal` fails, unlike the rest of the package which
is built entirely around "return a `Problem`, never crash." Not currently
reachable from adversarial input given today's callers, but the exported
function carries no documented precondition ruling it out for a future
caller.

**Fix:** Return an error / `core.validation_clone_failed` problem instead of
panicking.

### WR-03: `blockLoanLiveness`'s `uses` slice is recomputed and discarded on every worklist re-evaluation

**File:** `internal/compiler/check/check.go:475-501`, `593`

**Issue:** Not a correctness bug (unchanged assessment from the prior
review) — `blockLoanLiveness`'s `uses` return value is discarded inside
`loanLivenessFixpoint`'s worklist loop and is independently re-derived by
`materializeLoanEndpoints`. Purely a readability/maintenance note.

**Fix:** No functional change required; consider a comment at the call site
noting `uses` is intentionally discarded there.

## Info

### IN-01: `discoverLoanLastUses`'s admission logic is duplicated verbatim between `analyzeStraightLine` and `analyzeArmBody`

**File:** `internal/compiler/check/check.go:700, 1019` (and the ~140-line bodies that follow each)

**Issue:** Deliberate, documented scope decision (D-03-01), not an
oversight. Still worth flagging because the duplication is large enough
that a future fix to one admission rule has a real chance of being applied
to only one copy — unchanged from the prior review.

**Fix:** None required this phase.

### IN-02: `native.Runner.Run`'s stderr-nonempty rejection is the only signal distinguishing informational stderr from failure

**File:** `internal/compiler/native/native.go:118-120`

**Issue:** Deliberate and reasonable given today's generated C never writes
to stderr; a forward-compatibility note only, unchanged from the prior
review.

**Fix:** None required now.

---

_Reviewed: 2026-09-04T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
