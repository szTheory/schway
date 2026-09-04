---
phase: 03-borrowed-views-and-cfg-lifetimes
reviewed: 2026-09-04T00:00:00Z
depth: standard
files_reviewed: 42
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
findings:
  critical: 1
  warning: 3
  info: 2
  total: 6
status: issues_found
---

# Phase 03: Code Review Report

**Reviewed:** 2026-09-04T00:00:00Z
**Depth:** standard
**Files Reviewed:** 42 (plus 12 `testdata/phase3/*.lang` and `testdata/phase2/*` fixtures)
**Status:** issues_found

## Summary

This phase's headline architectural claim — three mechanically independent
derivations of loan liveness (`check.go`'s `loanLivenessFixpoint`,
`corevalidate.go`'s `recomputeLoanEndpoints`, and `pathoracle.go`'s
`RecomputeEndpoints`) — holds up under inspection. Each uses a genuinely
different algorithm (backward worklist fixpoint vs. static reachability
closure vs. exhaustive path enumeration), none imports another's package or
calls its helpers, and the loan-endpoint/loan-liveness surface is the one
place I looked hardest for the "self-confirming oracle" failure mode called
out for this review. I did not find that failure mode in the liveness
machinery itself.

I did find it, in effect, in a smaller and less obvious corner: OWN-04's
origin/access proof. `check.go` writes a function's `PublicOrigin.Access`
straight from the parsed declaration with zero verification against the
body; the only place that ever recomputes access from the body
(`originvalidate.RecomputeOrigin`) contains a real bug that lets it silently
report a stronger access level than the body actually establishes for a
reborrow chain that changes access partway through (a shared reborrow of an
exclusively-borrowed place). That is exactly the kind of defect this
review's mandate is calibrated to catch, and it is untested — no fixture or
unit test in this phase exercises a mixed shared/exclusive reborrow chain
through `RecomputeOrigin`.

Unbounded-work discipline is otherwise good: every untrusted read in
`session.go` uses the max-plus-one `io.LimitReader` pattern, `native.Runner`
enforces a context deadline on both the compile and run subprocess and
captures stdout/stderr into independently bounded writers, `pathoracle`
declares and enforces `MaxPaths`, and `debugmap.Build` checks its context
deadline and two independent output caps. See Warnings/Info for smaller gaps
(an unguarded `panic` in `corevalidate.cloneProgram`, and the already-known
scope note that `check`-time origin declarations are not validated until
`interface export`, which I flag only insofar as it is not documented next
to the code that makes the trust boundary real).

D-03-01 and D-03-02 (recorded in 03-DEBT.md) are known and out of scope for
this review; they are not repeated below.

## Critical Issues

### CR-01: `originvalidate.RecomputeOrigin` can report a stronger access mode than the body actually grants for a reborrow chain that changes access

**File:** `internal/compiler/originvalidate/originvalidate.go:75-100`

**Issue:** `RecomputeOrigin` walks backward from the function's `OpReturn`
operation toward the parameter, following the `TargetID -> SourceID` chain,
and derives the effective access mode:

```go
case core.OpBorrowExclusive:
    derivedAccess = "exclusive"
case core.OpBorrowShared:
    if derivedAccess == "" {
        derivedAccess = "shared"
    }
}
```

The `OpBorrowShared` branch is guarded ("only set the access mode the first
time it is seen"), correctly making the *closest-to-return* borrow the one
that determines the answer. The `OpBorrowExclusive` branch has no such
guard: it unconditionally overwrites `derivedAccess` every time an exclusive
borrow is encountered anywhere along the chain, including one that is
*further from the return value* than an already-recorded shared borrow.

Consider a straight-line body that is legal under this phase's grammar and
checker (each `let` binding's RHS source is just an identifier, so chaining
is unrestricted, and `checkLinear`/`analyzeStraightLine` place no rule
against reborrowing a borrow-derived place with a different access mode):

```
fn view(x: Buffer) -> borrow mut(x) Buffer {
  let y = borrow mut x
  let z = borrow y
  z
}
```

`z` is a *shared* reborrow of the exclusively-borrowed `y` — the caller who
receives `z` only ever has read access. Walking backward from the `Return
z` operation: the first hop found is `BorrowShared y->z`, which correctly
sets `derivedAccess = "shared"`. The walk continues to `BorrowExclusive
x->y`, which then unconditionally overwrites `derivedAccess` to
`"exclusive"` — the wrong answer. `RecomputeOrigin` reports `"exclusive"`,
which matches the function's declared `borrow mut(x)` origin, so
`originvalidate.ValidatePublished` (and `originvalidate.CheckSummary`'s
whole reason for existing) accepts a function that *claims write access to
the caller* while the value it actually hands back is only ever a
shared/read-only reborrow.

This is the exact class of bug OWN-04 exists to catch (`check.go`'s own
`PublicOrigin` construction is copied verbatim from the parsed declaration
with no body verification at all — see `check.go:971-974` — so
`originvalidate.RecomputeOrigin` is the *only* place in the entire
compiler that ever checks a declared origin's access mode against the
body). A bug in it defeats the load-bearing guarantee, not a peripheral one.

No test in `originvalidate_test.go`, `check_origin_test.go`, or any
`testdata/phase3/public_view*.lang` fixture exercises a chain that mixes
`borrow` and `borrow mut` across more than one hop — every fixture's origin
chain is a single borrow directly off the parameter — so this is a live,
unguarded gap, not a documented residual (it is materially different from
D-03-02, which is about a *missing* declared origin, not a *wrong* access
mode reported as matching for a real declared one).

**Fix:** Make the exclusive case symmetric with the shared case — the
access mode should be decided by the *first* borrow encountered while
walking backward from the return (the hop nearest the returned place), not
overwritten by anything further up the chain:

```go
switch operation.Kind {
case core.OpBorrowExclusive:
    if derivedAccess == "" {
        derivedAccess = "exclusive"
    }
case core.OpBorrowShared:
    if derivedAccess == "" {
        derivedAccess = "shared"
    }
}
```

Add a regression fixture/test with a chain like the one above (`borrow mut`
then `borrow`, and the reverse `borrow` then `borrow mut` if that ordering
is reachable) asserting `RecomputeOrigin` returns `"shared"` for the first
case, so `ValidatePublished` correctly rejects a `borrow mut(x)` declaration
against it with `core.origin_access_mismatch`.

## Warnings

### WR-01: `check`-time programs carry an unverified `PublicOrigin` claim; only `interface export` verifies it

**File:** `internal/compiler/check/check.go:971-974`, `internal/compiler/session/session.go:125-142` (`Check`), `internal/compiler/session/session.go:283-336` (`RunNative`), `internal/compiler/session/session.go:204-230` (`RunInterpreter`)

**Issue:** `checkLinear` builds `core.PublicOrigin` directly from the AST's
parsed `ReturnOrigin` declaration with no verification that the function
body actually returns a value derived from a borrow of that access mode
(confirmed by grep: `PublicOrigin` is never read by `corevalidate.go`,
`cgen.go`, or `interp.go` — it is inert everywhere except
`originvalidate.go`). `session.Check`, `RunInterpreter`, and `RunNative` —
the paths behind `lang check` and `lang run` — never call
`originvalidate.ValidatePublished`. Only `InterfaceExportCommandFile` does.
This means a source file can declare a completely fictitious origin (e.g.
`-> borrow mut(x) Buffer` on a function whose body never borrows `x` at
all, only moves and returns an owned value) and `lang check` / `lang run`
will accept and execute it without complaint; the lie is caught only if and
when someone separately runs `lang interface export`.

Given `PublicOrigin` currently has no runtime consumer this is likely an
intentional "declaring doesn't mean anything until you publish it" design,
but it is exactly the kind of split that quietly breaks when a future
change (e.g. `debugmap` or a future native ABI consumer) starts trusting
`PublicOrigin` off a plain `checked.Program` — nothing in `check.go` or
`session.Check` documents that `PublicOrigin` is unverified there.

**Fix:** At minimum, add a doc comment on `core.PublicOrigin` (or on
`check.Program`/`session.Check`) stating explicitly that the field is an
unverified declaration until `originvalidate.ValidatePublished` runs, so a
future caller does not accidentally trust it off a bare `check` result.
Consider whether `lang check` itself should run `ValidatePublished` when a
function declares a `PublicOrigin`, since silently accepting an impossible
origin declaration through the primary `check`/`run` path is a surprising
UX even if nothing downstream currently acts on the lie.

### WR-02: `corevalidate.cloneProgram` panics instead of returning an error

**File:** `internal/compiler/corevalidate/corevalidate.go:1140-1150`

**Issue:** `cloneProgram` (used by `Validate` on every call, including every
`check`/`run`/`interface`/`debug-map`/`verify` invocation) calls `panic` if
`json.Marshal`/`json.Unmarshal` fails. In today's callers this is reached
only with well-formed, already-Go-typed `core.Program` values built by
`check.Program`, so it is not currently reachable from adversarial input —
but `corevalidate.Validate` is an exported package function with no
documented precondition that its input must already be internally
consistent Go data (as opposed to, say, a value freshly deserialized from
untrusted JSON by a future caller). A crash-on-panic in a compiler library
function is a harsher failure mode than the rest of this package, which is
built entirely around "return a `Problem`, never crash" (see the extensive
`carriedLoans` doc comment explicitly calling out avoiding recursion so it
"must never crash... on adversarial input").

**Fix:** Return an error (or a `Result` with a `core.validation_clone_failed`
problem) instead of panicking, so `corevalidate.Validate` stays crash-free
under the same adversarial-input posture the rest of the file documents for
itself.

### WR-03: `blockLoanLiveness`'s `uses` slice grows unboundedly across worklist re-evaluations without being reset when discarded

**File:** `internal/compiler/check/check.go:475-501`, `593` (call site inside `loanLivenessFixpoint`)

**Issue:** Inside `loanLivenessFixpoint`'s worklist loop, `blockLoanLiveness`
is called once per block per re-evaluation and constructs a fresh `uses`
slice each call, which is discarded (only `newLiveIn` and `transferWork`
are consulted) unless the block's live-in set changed. This is not a
correctness bug (the discarded `uses` is genuinely unused there — it is
only consulted later by `materializeLoanEndpoints`'s own separate,
single-pass call), but it does mean every worklist re-evaluation of a block
re-walks and re-allocates a `uses` slice purely for a value the fixpoint
loop throws away. Given the documented per-function work-counting
discipline (D-05) this is counted correctly in `work`, so it is not an
unbounded-work correctness issue, but it is worth a second look since a
reader skimming the fixpoint loop for its live-in-only fixpoint state may
reasonably assume `blockLoanLiveness`'s return values are all consumed
there.

**Fix:** No functional change required; consider a comment at the call site
(`check.go:593`) noting that `uses` is intentionally discarded here and is
re-derived independently by `materializeLoanEndpoints`, to save the next
reader the trip through both functions to confirm nothing is silently
dropped.

## Info

### IN-01: `discoverLoanLastUses`'s uncounted work is the pre-existing debt item, but its result is now used by two different code paths that could silently drift

**File:** `internal/compiler/check/check.go:700, 1019`

**Issue:** Purely informational, not a new finding: `discoverLoanLastUses`
(governs admission in both `analyzeStraightLine` and `analyzeArmBody`, per
D-03-01) is called from two places with the same signature and the same
loop-driven mutation pattern duplicated verbatim between
`analyzeStraightLine` and `analyzeArmBody` (compare `check.go:700-720` with
`check.go:1019-1026`, and the two ~140-line bodies that follow them). This
is called out in the package doc comment as a deliberate scope decision
(not touching the straight-line path this phase) rather than an oversight,
so it is not a defect — flagging only because the duplication is large
enough (two near-identical ~150-line functions) that a future fix to one
admission rule has a real chance of being applied to only one copy.

**Fix:** None required this phase; consider unifying the two admission
functions in a later phase once the straight-line path is deliberately
rewired onto the CFG dataflow (tracked separately from D-03-01).

### IN-02: `native.Runner.Run`'s stderr-nonempty rejection makes clang warnings-as-errors the only way informational stderr output is distinguished from failure

**File:** `internal/compiler/native/native.go:118-120`

**Issue:** After a successful run, any non-empty stderr from the compiled
program is treated as a hard `native.run_stderr` failure:
`if len(runStderr.bytes()) != 0 { return ..., "native.run_stderr" ... }`.
This is a deliberate, and reasonable, "the generated program must be
silent on stderr" contract given the compiler's own generated C never
writes to stderr — not a bug — but it means any future addition to
`cgen`'s generated C that writes a warning to stderr (e.g. under a future
`-Wextra`-driven runtime assertion) would silently start failing every
native run with an opaque `native.run_stderr` code rather than a
targeted diagnostic. Purely a forward-compatibility note, not a defect in
this phase's code.

**Fix:** None required now; if a future phase adds any stderr-writing
runtime path to generated C, revisit this gate.

---

_Reviewed: 2026-09-04T00:00:00Z_
_Reviewer: Claude (gsd-code-reviewer)_
_Depth: standard_
