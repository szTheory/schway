---
phase: 32-clean-checkout-guide-and-final-receipts
plan: 06
subsystem: cmd/schway route audit
tags: [evd-15, public-route-audit, go-parser, receipts]
requires: [32-03]
provides:
  - TestPhase32PublicRouteAudit with phase32RouteRows covering every public route and example entry program
  - phase32EnumerateRoutes and phase32EnumerateEntries (go/parser and syntax.Parse enumeration)
  - pending todo recording the usage-text omissions
affects: [32-08, 32-09]
key-files:
  created:
    - cmd/schway/main_phase32_routes_test.go
    - .planning/todos/pending/2026-10-12-usage-text-omits-stats-and-build-manifest.md
  modified: []
key-decisions:
  - "Covered program set is exactly {checksum, adler_low, adler32_cli, adler32_fixture, byte_index_oob, sum_to_n, fizzbuzz}; the two phase29 programs are covered rows (D-09a)"
  - "Usage omissions (stats, build --manifest) are recorded in phase32UsageOmissions and a todo; main.go is not edited (D-10d)"
requirements-completed: [EVD-15]
status: complete
commits: 2
plan_head_before: d7ee938e5314b22d598cef7e779cde10d1491fe2
plan_head_after: f7a03f7cd4dc6ccd44221fbf3417a2e788feecdd
verification_role: proposal
actuals:
  tokens: 5000
  tasks: 2
  commits: 2
coverage:
  - id: D-10b / D-10d
    description: "Every public CLI route enumerated by AST from run, runApplication and runApplicationRun, and every examples fn main program, has a covered row naming a receipt program or a not-covered row naming an existing test; a new route, deleted row, stale row, invented test name and unknown program each fail, and the usage text is cross-checked"
    requirement: "EVD-15"
    verification:
      - kind: unit
        ref: "cmd/schway/main_phase32_routes_test.go#TestPhase32PublicRouteAudit"
        status: pass
    human_judgment: false
---

# Phase 32 Plan 06: Public-route audit Summary

An AST-driven audit test enumerates the 22 public CLI route keys and 14 example entry programs and requires each to have a covered or test-owned row, so the final receipts name exactly what they cover.

## Accomplishments

- `TestPhase32PublicRouteAudit` reads `cmd/schway/main.go`, walks `run`, `runApplication` and `runApplicationRun` with go/parser, and builds route keys from `args[N]` literal comparisons, `strings.HasPrefix` literals and nested switch cases. It also finds `examples/**/*.schway` files declaring `fn main` through `syntax.Parse`.
- Covered rows name the seven D-09a receipt programs; the audit asserts the covered union equals that set. Every other route or entry has a `not covered because` row whose reason names a Test function found by a go/parser scan of `cmd/` and `internal/`.
- Armed subtests: `uncovered-route`, `deleted-row`, `stale-row`, `bad-reason`, `unknown-program`, `usage-cross-check`.
- The usage text is cross-checked: every subcommand it names is a route, and the routes it omits equal `phase32UsageOmissions` (`stats`, `build --manifest` in both orders). A pending todo records the gap; `main.go` is unchanged.
- Scope non-claim in the file header: only `cmd/schway` routes and `examples/**` entries are audited; `testdata/` fixtures are outside the receipt claim.

## Task Commits

1. Task 1 (tracer): `f2e50c87` test(32-06): public-route audit for the final receipts
2. Task 2: `f7a03f7c` test(32-06): arm the route audit and record the usage-text gap

The armed controls were written together with the tracer, so Task 1's commit already carries them; Task 2's commit adds the todo.

## Deviations from Plan

**Armed subtests landed in the Task 1 commit** rather than Task 2. Written in one pass, the file was committed once the whole test passed. No behavior differs from the plan. Added an `unknown-program` control beyond the four named.

No auto-fixes needed.

## Verification

- `go test ./cmd/schway -run '^TestPhase32PublicRouteAudit$' -count=1 -v`: PASS with six subtests.
- `go test ./cmd/schway -count=1`: ok.
- Session meta set from 32-VALIDATION.md plus `TestEntryFunctionCalledOnlyViaResolveEntry` and `TestValidationRowGradesAreEarnedOverArchivedCorpus`: ok.
- `grep -cE 'callgraph\.EntryFunction|cgen\.'` on the new file prints 0; `main.go` unchanged against the plan base.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- FOUND: cmd/schway/main_phase32_routes_test.go
- FOUND: .planning/todos/pending/2026-10-12-usage-text-omits-stats-and-build-manifest.md
- FOUND commits: f2e50c87, f7a03f7c
