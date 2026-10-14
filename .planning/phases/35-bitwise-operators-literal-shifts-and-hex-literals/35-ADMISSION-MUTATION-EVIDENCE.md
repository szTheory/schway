# Phase 35 finite admission-mutation receipt — 2026-10-14

This scoped D-09/D-10/D-11 and 35-04 tasks 2/3 receipt records actual
implementation-source refusal-removal controls. It supplements the existing
15 arithmetic transfer controls in `35-MUTATION-EVIDENCE.md`; those controls
were unaffected and were not rerun here. No production source changed.

## Executed command and outcome

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$' -count=1 -v -timeout=600s
```

PASS, exit 0, `ok github.com/szTheory/schway/internal/compiler/session 9.566s`,
on Go 1.24.0, darwin/arm64, checkout based on
`c4138a34d62a667742577b95ff3bd840ce88bbe8`.
The tested harness and assertion tables are committed as `397bfb6d`.

The harness enforced exactly **26 controls: 12 corevalidate, 12 originvalidate,
2 check**. All 26 printed `MUTANT_KILLED`; the children failed at exactly 46
intended named assertions. All three peers printed `NO_OP_PASS` and
`RESTORED_PASS`. Each raw-IR peer ran and passed 34 named baseline rows in
`TestPhase35ShiftImmediateAdmission` and `TestPhase35ShiftTargetFreshness`;
check ran and passed nine named rows in `TestPhase35ShiftAdmission`.

Each serial child overlays a temporary copy of the actual owner implementation
using Go `-overlay`; checked-in implementation files stay untouched. Every
replacement checks its exact declared source occurrence count before running
the child. The 55 replacement sites below include binding repairs and combined
guard clauses; they are not 55 independent semantic laws. Coherent redundant
guards for the selected law are removed together. Normal Go compilation and
vet remain enabled. A compile/vet error, missing child, surviving mutant,
unexpected marker or unrelated failing row cannot satisfy the intended kill.

## Exact control and assertion inventory

Raw rows below qualify under `TestPhase35ShiftImmediateAdmission/` unless
otherwise stated. All raw controls require `ADMISSION_ACCEPTED:` on the named
row. Counts are exact source replacement sites, core/origin respectively.

| Independent law | Core sites | Origin sites | Required failing row(s) |
|---|---:|---:|---|
| Canonical and present immediate | 5 | 4 | `missing_immediate`, `leading_zero` |
| Immediate range 0 through 63 | 3 | 2 | `count_64` |
| Empty shift RightID | 2 | 1 | `extra_right_place` |
| U64 source | 3 | 2 | `wrong_source_type` |
| U64 operation result type | 4 | 1 | `wrong_result_type` |
| U64 result place | 3 | 2 | `wrong_result_place_type` |
| Initialized source | 3 | 1 | `uninitialized_U64_source` |
| Nonparameter, nonaliasing target | 3 | 2 | `parameter_target`, `source_aliases_target` |
| Fresh target | 2 | 1 | All eight `TestPhase35ShiftTargetFreshness` rows described below |
| Empty unrelated scalar fields | 1 | 1 | Core: `foreign_field`; origin: `payload_field` |
| Empty store/edge fields | 2 | 1 | `store_field`, `edge_field` |
| ConstU64 kind exclusivity | 2 | 1 | `unrelated_const_field` |

Freshness requires each combination of `shift_left_u64` / `shift_right_u64`
and `prior_constant` / `prior_copy` / `prior_mutable_store` / `prior_shift`,
qualified as `TestPhase35ShiftTargetFreshness/{kind}/{row}`. Both peers require
all eight assertions. Target-identity and result-place fixture edits rewire
later consumers to initialized U64 values so an abandoned result or a copy's
own type law does not mask the admission law. The origin payload witness
isolates the unrelated-field law; its foreign-field baseline refusal remains
checked, without disabling the separate local-owner lifecycle authority.

| Source check law | Sites | Required named assertion |
|---|---:|---|
| Amount must be a number AST leaf | 1 | `TestPhase35ShiftAdmission/name`: `SOURCE_ADMISSION_DIAGNOSTIC_MISMATCH:` |
| Amount range 0 through 63 | 2 | `TestPhase35ShiftAdmission/decimal_64`: `SOURCE_ADMISSION_ACCEPTED:` |

The number-leaf control must change the exact diagnostic contract. Every
other control must reach actual acceptance of the malformed raw/source row,
not an unrelated refusal. Marker tokens are matched exactly and attributed to
their named child `RUN`; the source acceptance marker cannot double-count as
the raw acceptance marker. Only intended failing rows and their aggregate
parent failures are allowed.

## Bound production snapshots

These SHA-256 hashes matched the debug-session baseline before and after the
run; the harness also rechecks live source hashes after every mutant and in
each restored control.

| Implementation source | SHA-256 |
|---|---|
| `internal/compiler/corevalidate/corevalidate.go` | `6c8892a0399ff817e86349fe80bd635429e969d63bd738c85ae3585c87c4466f` |
| `internal/compiler/originvalidate/originvalidate.go` | `583468465ff408e25b92854174f61546de65fd4a5444b24b33349c806e0e0fa6` |
| `internal/compiler/check/scalar.go` | `a2cacabc99cd3908cacdd11f50a3c29487576a8d3489a77d86f322c8cd6c0167` |

## Repair and independent reconfirmation

The pre-fix focused run reproduced 13 harness defects among 19 selected
controls. Cycle 1 repaired 12 of those defects; the remaining core result-place
mutant hit vet's `redundant and: true && true`. Cycle 2 combined those adjacent
clauses into one edit; the isolated result-place control passed before the
full matrix ran. No build failure was credited as a kill.

The old harness was then replayed through a temporary test-file overlay:

```sh
GOCACHE=/tmp/ai-lang-verification-gocache go test -overlay=/tmp/phase35-admission-reconfirm.json ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$/^check$/^source-range-0-through-63$' -count=1 -v -timeout=600s
GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session -run '^TestPhase35AdmissionRefusalMutationControls$/^check$/^source-range-0-through-63$' -count=1 -v -timeout=600s
```

The first command used the saved pre-repair harness and failed with the exact
`mutant produced 2 named assertion failures, want 1` bug (exit 1, 0.177s).
The current harness passed the same source-range control (exit 0, 0.187s).
This reconfirms the attribution repair without reverting any live files.
Temporary logs/overlays under `/tmp/phase35-admission-*` are local diagnostic
artifacts, not tracked fixtures or portable receipt prerequisites.

## Limits

This is finite law-level mutation evidence with explicit baseline boundary
neighbors (0/63 accepted; 64/max/overflow and malformed spellings refused).
It does not claim a Cartesian mutation of every field, every route, every
malformed spelling or every operation. Raw peers prove IR laws, not source
literal provenance. No broad test, separate vet/build, native/sanitizer,
hosted, or cross-platform lane ran here, and no phase completion is asserted.
No Stryker/Go mutation tool is configured for the harness parser; that extra
fix-site mutation signal was skipped explicitly. The old-parser overlay
reconfirmation and all 26 owner-source mutants are the executed controls.
