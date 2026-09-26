---
phase: 14-evidence-instrument-and-honest-scoping
plan: 13
subsystem: testing
tags: [go-build-constraints, suppression-guard, evidence-instrument, witness-grammar]

requires:
  - phase: 14-evidence-instrument-and-honest-scoping
    provides: "plan 14-07's scanSuppressionSurfaces/suppressionProblems suppression-witness guard, and the D-14-25 four-surface enumeration it left with an inert build-constraint branch"
provides:
  - "buildConstraintAllowlist (darwin/linux/amd64/arm64/cgo) and buildConstraintTerms, a term extractor for //go:build lines"
  - "suppressionProblems' missing build-constraint branch: an unallowlisted term with no resolvable citation is now a named problem"
  - "scanSuppressionSurfaces' textual //go:build pass reordered ahead of buildCtx.MatchFile, so a host-excluded file cannot hide its own constraint from this guard"
  - "TestSuppressionWitnessGuardIsNotInert's fourth seeded-fault subtest (build constraint outside the allowlist)"
  - "D-14-25's doc text rewritten to state the enforced policy and allowlist membership in place"
affects: [session-suppression-guard, evd-04-closure]

actuals:
  tokens: 3528
  tasks: 2
  commits: 4

tech-stack:
  added: []
  patterns:
    - "Term-extraction, not constraint-evaluation: buildConstraintTerms asks 'which terms appear', never 'does this build' -- go/build.Context.MatchFile owns the evaluation question"
    - "Textual pre-pass ahead of MatchFile: a surface that a constraint could hide from its own checker must be scanned before the constraint is evaluated as a gate"

key-files:
  created: []
  modified:
    - internal/compiler/session/witness_registry_test.go

key-decisions:
  - "buildConstraintAllowlist seeded with darwin, linux, amd64, arm64, cgo per AGENTS.md's stated macOS/Linux host priorities plus Go's own cgo tag -- any other term needs a citation"
  - "MatchFile's early return moved to gate only the AST-based passes (comments, string literals, Skip calls); the textual //go:build pass now runs on every .go file's raw bytes unconditionally"
  - "The allowlist escape-hatch citation fixture places its //go:build-prefixed line AFTER the package clause (not at the real directive position), because a genuine D-XX-NN citation contains a hyphen that is not valid Go build-tag syntax and would make go/parser refuse to parse a real leading directive -- the textual scanner has no such position requirement and still detects it"

requirements-completed: [EVD-04]

coverage:
  - id: D1
    description: "suppressionProblems raises a named problem for a //go:build term outside buildConstraintAllowlist with no resolvable citation"
    requirement: "EVD-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestBuildConstraintsOutsideTheAllowlistAreRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestSuppressionWitnessGuardIsNotInert/build_constraint_outside_the_allowlist"
        status: pass
    human_judgment: false
  - id: D2
    description: "A .go file whose own //go:build constraint excludes it from the current host build is still enumerated as a build-constraint site, and an unallowlisted term in it is still reported"
    requirement: "EVD-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile"
        status: pass
    human_judgment: false
  - id: D3
    description: "The live module stays clean under the new branch, and the module-wide scan still finds the pre-existing Skip/comment/string-literal surfaces"
    requirement: "EVD-04"
    verification:
      - kind: unit
        ref: "internal/compiler/session/witness_registry_test.go#TestNoSuppressionOutlivesItsWitness"
        status: pass
    human_judgment: false

duration: 15min
completed: 2026-09-18
status: complete
---

# Phase 14 Plan 13: Build-Constraint Suppression Guard Summary

**`suppressionProblems` now checks the `//go:build` surface it always collected but never gated -- a declared `buildConstraintAllowlist`, a moved `MatchFile` gate so host-excluded files can't hide their own constraint, and a fourth seeded fault proving the branch is live.**

## Performance

- **Duration:** ~15 min
- **Started:** 2026-09-18T14:00:00Z (approx.)
- **Completed:** 2026-09-18T14:08:51Z (last task commit)
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Closed the structurally inert `//go:build` branch of the EVD-04 suppression-witness guard: `buildConstraintAllowlist` (term -> allow reason) plus `buildConstraintTerms` (a stdlib-only, boolean-syntax-aware term extractor) wired into `suppressionProblems`, which now raises a named problem for any constraint term outside the allowlist with no resolvable citation on the same line.
- Moved the textual `//go:build` scan in `scanSuppressionSurfaces` ahead of `buildCtx.MatchFile`, so a file cannot use its own constraint to become invisible to the guard that checks constraints (closes T-14-13-02); `MatchFile` now gates only the AST-based passes (comments, string literals, Skip calls).
- `TestSuppressionWitnessGuardIsNotInert` gained a fourth seeded-fault subtest (three fixtures: allowlisted green, unallowlisted red, cited-unallowlisted green), and a dedicated tracer/RED test each prove the two new load-bearing behaviors independently.
- Rewrote the D-14-25 doc comment above `scanSuppressionSurfaces` to state, in the tree, exactly what the code now enforces and the allowlist's membership -- no claim the code doesn't back.

## Task Commits

Each task was committed atomically:

1. **Task 1: One unallowlisted build constraint, end to end, goes red** - `6854b63` (feat)
2. **Task 2: The scan sees constrained-out files, and the new branch carries its own seeded fault**
   - RED - `84deafc` (test)
   - GREEN - `f81d303` (feat)

**Plan metadata:** (this commit) - docs: complete plan

## Files Created/Modified
- `internal/compiler/session/witness_registry_test.go` - added `buildConstraintAllowlist`/`buildConstraintTerms`, wired the missing build-constraint branch into `suppressionProblems`, reordered `scanSuppressionSurfaces`'s `MatchFile` gate, added `TestBuildConstraintsOutsideTheAllowlistAreRefused`, `TestBuildConstraintSurfaceSeenEvenWhenHostExcludesFile`, the fourth `TestSuppressionWitnessGuardIsNotInert` subtest, and rewrote the D-14-25 doc comment

## Decisions Made
- `buildConstraintAllowlist` membership: `darwin`, `linux` (AGENTS.md's stated host priorities), `amd64`, `arm64` (portable GOARCH targets, current host included), and `cgo` (Go's own build tag, not a portability decision). Any other term needs a citation.
- The citation escape-hatch fixture places its `//go:build`-prefixed text after the package clause rather than at the real leading-directive position, since a real `D-XX-NN` citation contains a hyphen that is invalid Go build-tag syntax and would make `go/parser` refuse to parse a genuine leading directive carrying it. The textual scanner enumerates any line prefixed `//go:build` regardless of position, so the fixture still exercises the identical site kind and citation-resolution code path `suppressionProblems` runs against a real leading directive.

## Deviations from Plan

None - plan executed exactly as written. Both tasks landed with the exact task shapes, fixture styles (interpreted strings, escaped newlines, no backtick raw strings, no incidental citation-shaped substrings), and verification commands the plan specified.

## Issues Encountered

- The plan's Task 2 citation-escape-hatch fixture ("an unallowlisted term with a resolvable citation on the same line") initially failed with a `go/parser` error (`parsing //go:build line: invalid syntax`) when the `D-14-25` citation was appended as trailing text on a genuine leading `//go:build` directive line -- `go/parser`'s own build-constraint syntax validator rejects the hyphenated citation as an invalid build-tag token when the comment sits in the real directive position. Resolved by moving that fixture's `//go:build`-prefixed line to after the package clause, where `go/parser` no longer treats it as an official directive (skipping validation) while `scanSuppressionSurfaces`' own position-agnostic textual scan still detects and evaluates it identically. Documented inline in the test as a comment explaining why.

## Manual Red Proofs (plan-required, `<verify>`/acceptance criteria)

Both perturbations were applied to real files under version control, observed failing `TestNoSuppressionOutlivesItsWitness`, then deleted and reverted; `git status --short` was confirmed clean after each.

**Proof 1 -- an unallowlisted term in a real package (host-matching constraint):**

File added: `internal/compiler/testsupport/tmp_manual_red_proof.go`
```go
//go:build darwin || linux || windows

package testsupport

func unallowlistedProbeConstant() {}
```

`go test ./internal/compiler/session/ -run 'TestNoSuppressionOutlivesItsWitness$' -count=1 -v` failure output (verbatim):
```
    witness_registry_test.go:562: 1 suppression problem(s):
        internal/compiler/testsupport/tmp_manual_red_proof.go:1 (build-constraint): constraint term "windows" is outside buildConstraintAllowlist and carries no resolvable citation
--- FAIL: TestNoSuppressionOutlivesItsWitness (0.23s)
```

File deleted; `git status --short` confirmed clean.

**Proof 2 -- an unallowlisted term in a constraint that EXCLUDES the current host (proves the `MatchFile` reordering is load-bearing):**

Host at the time: `GOOS=darwin GOARCH=arm64`.

File added: `internal/compiler/testsupport/tmp_manual_red_proof2.go`
```go
//go:build plan9

package testsupport

func excludedProbeConstant() {}
```

`go test ./internal/compiler/session/ -run 'TestNoSuppressionOutlivesItsWitness$' -count=1 -v` failure output (verbatim):
```
    witness_registry_test.go:562: 1 suppression problem(s):
        internal/compiler/testsupport/tmp_manual_red_proof2.go:1 (build-constraint): constraint term "plan9" is outside buildConstraintAllowlist and carries no resolvable citation
--- FAIL: TestNoSuppressionOutlivesItsWitness (0.23s)
```

The guard named the file and term even though `plan9` excludes it from the current `darwin`/`arm64` host -- proving the file could no longer hide its own constraint from the guard, which was exactly the pre-existing hole. File deleted; `git status --short` confirmed clean.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- EVD-04 closed: the `//go:build` suppression surface is now checked, not merely collected. All four `scanSuppressionSurfaces` surfaces (build-constraint, comment, string-literal, skip) are now gated by `suppressionProblems`.
- `go test ./... -count=1` exits 0 (full corpus, 227.573s in package `session`).
- No new debt rows or open blockers introduced by this plan.

---
*Phase: 14-evidence-instrument-and-honest-scoping*
*Completed: 2026-09-18*

## Self-Check: PASSED
