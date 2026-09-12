---
status: complete
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
source: [11-VERIFICATION.md]
started: 2026-09-12T10:19:31Z
updated: 2026-09-12T11:48:00Z
---

## Current Test

[testing complete]

## Tests

### 1. Decide on WR-01 — reduce.Reduce trusts Seed.EntryFunctionID without validating it
expected: Accept the residual risk for M002, or require a follow-up plan adding a fail-closed Seed.Validate() guard. Reproduce with: go test ./internal/compiler/reduce/... -run TestReduceRejectsMultiFunctionSeed -v — then inspect reduce.go:202-254 and :320-347, where dropOrphanFunction relies on EntryFunctionID to exempt the real entry point from deletion.
result: pass
source: automated
resolution: shift-left — the human decision was eliminated rather than made. The fail-closed guard 11-REVIEW.md suggested is implemented (internal/compiler/reduce/seed_validate.go, Seed.Validate, called at the top of Reduce), so the hazard is now refused by the compiler rather than adjudicated by a reviewer. Covered by internal/compiler/reduce/seed_validate_test.go.
verification:
  - go test ./internal/compiler/reduce/... -run TestReduceRefusesMultiFunctionSeedWithEmptyEntryID
  - go test ./internal/compiler/reduce/... -run TestReduceRefusesMultiFunctionSeedWithUnknownEntryID
  - go test ./internal/compiler/reduce/... -run TestReduceAcceptsSingleFunctionSeedWithoutEntryID
  - go test ./internal/compiler/reduce/... -run TestSeedEntryHazardIsReal
  - go test ./internal/compiler/reduce/... -run TestSeedEntryRefusalMessageIsByteStable
  - go test ./internal/compiler/reduce/... -run TestSeedEntryRefusalDispatchesLikeCallgraphRefusal

### 2. Decide on the mid-phase gate's disclosed CLI-check divergence
expected: Decide whether the pre-existing check-command vs session.Check split needs its own tracked debt item, independent of Phase 11. Reproduce with: go run ./cmd/lang --json check testdata/phase11/multi_function_gate_corpus.lang (reports status:invalid / core.origin_omitted) versus session.Check + corevalidate.Validate, which accept the identical program cleanly. The same divergence reproduces on the already-shipped Phase 5 fixture testdata/phase5/restrict_borrow.lang, confirming it predates Phase 11.
result: pass
source: automated
resolution: shift-left — the question is now re-answered by the test suite on every run instead of once in prose. A corpus-wide sweep (internal/compiler/session/session_admission_divergence_test.go) drives every committed .lang fixture through BOTH admission surfaces and pins the exact divergence set. It answers the gate's question with evidence: the divergence spans 10 fixtures across phases 3, 4, 5, 7, 10 and 11 — it is neither new nor Phase 11's, but the designed consequence of CheckCommandFile applying originvalidate.ValidatePublished (D-04-27/WR-01) as a third admission layer the internal path does not. No debt row added: a tracked debt item is read once, whereas this is checked forever, in CI, on both hosts.
verification:
  - go test ./internal/compiler/session/ -run TestCLICheckAdmissionDivergenceIsExactlyKnown
  - go test ./internal/compiler/session/ -run TestKnownAdmissionDivergencesIsSortedAndComplete

## Summary

total: 2
passed: 2
issues: 0
pending: 0
skipped: 0
blocked: 0

## Gaps

[none]

## Resolution Note — human verification eliminated, not waived

Both items entered this session as *decisions asked of a human*. Neither was
answered as such. Per the operator's direction ("shift left, automate the
world, goal is 0 human verification/UAT required — even roll it into CI iff
there's recurring value there"), each was converted into a mechanical check
that re-decides it on every run:

| Was | Now | Recurring value |
|---|---|---|
| "Is reduce's unvalidated EntryFunctionID an acceptable residual risk?" | It is refused. `Seed.Validate` fails closed; 6 tests, including an anti-vacuity control proving the deletion hazard is real. | The guard cannot silently rot: `TestSeedEntryHazardIsReal` fails if the hazard it protects against ever stops existing, forcing the guard to be re-justified rather than kept as decoration. |
| "Does the CLI-check divergence need a tracked debt item?" | It is pinned. A corpus-wide sweep asserts the divergence set exactly, failing in BOTH directions. | Catches the two admission surfaces drifting further apart (new divergence) AND silently converging (a resolved one), so reconciliation becomes a reviewed edit instead of an unnoticed change. |

Both land in CI with no workflow edit: `.github/workflows/ci.yml` already runs
`go vet ./...`, `go test ./...` and `go test -race ./...` on ubuntu-latest and
macos-latest for every push and pull request.

**Mutation-checked.** The divergence test was verified to fail in both
directions before being accepted: removing a real entry from the table
produces `UNDECLARED admission divergence`, and adding a phantom entry
produces `RESOLVED admission divergence` plus a missing-fixture failure.
A test that cannot fail would have been worse than the human question it
replaced.

**Full suite green at resolution:** `go vet ./...`, `go build ./...`,
`go test ./...`, `go test -race ./...`, and `sh scripts/verify-phase6.sh`
(the phase gate CI runs) all pass.
