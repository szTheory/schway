# Phase 38: Milestone Receipts - Pattern Map

**Mapped:** 2026-10-16  
**Files analyzed:** 6  
**Analogs found:** 5 / 6

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|-------------------|------|-----------|----------------|---------------|
| `scripts/verify-phase32.sh` (modify) | utility / host driver | batch | `scripts/verify-phase32.sh` | exact |
| `scripts/check-milestone-binding.sh` (new) | utility / gate | batch | `scripts/check-phase27-receipts.sh`; `scripts/check-receipts.sh` | role-match |
| `testdata/receipts/M007.manifest` (new) | config / fixture | batch | `testdata/receipts/M006.manifest` | exact |
| `.github/workflows/ci.yml` (modify) | config / workflow | event-driven, batch | `.github/workflows/ci.yml` | exact |
| `internal/compiler/session/session_gate_differential_test.go` (modify; inferred test owner) | test | request-response | `internal/compiler/session/session_gate_differential_test.go` | exact |
| `.planning/phases/38-milestone-receipts/` (new copied host logs and bounded run metadata; final plan) | evidence artifact | batch | None tracked | none |

The final artifact basenames and metadata encoding remain planning choices. The final plan should copy both raw host logs and bounded run/attempt provenance, then check the copied bytes. A hosted log is run-specific evidence, so no persisted receipt artifact is used as a source-code analog.

## Pattern Assignments

### `scripts/verify-phase32.sh` (utility / host driver, batch)

**Analog:** `scripts/verify-phase32.sh` (tracked; modifying the maintained driver is the intended fit).

Keep its no-argument M006 contract and add an explicit opt-in M007 selector. Existing setup selects the pinned M006 manifest and emits host/source identity (`scripts/verify-phase32.sh:37-68`):

```sh
receipt_manifest="$verify_root/testdata/receipts/M006.manifest"
receipt_programs=$(sed -n 's/^programs=//p' "$receipt_manifest" | sed -n '1p')
[ -n "$receipt_programs" ] || fail "programs line is missing" "programs_line_missing"
```

The current driver compares exact process tuples and makes missing results fatal (`scripts/verify-phase32.sh:129-137`):

```sh
status=pass
stdout=exact; stderr=exact
cmp -s "$out_file" "$got_out" || { stdout=mismatch; status=fail; }
cmp -s "$err_file" "$got_err" || { stderr=mismatch; status=fail; }
[ "$got_exit" = "$want_exit" ] || status=fail
printf 'phase32 case program=%s input=%s expected_exit=%s actual_exit=%s stdout=%s stderr=%s status=%s\n' \
  "$program" "$key" "$want_exit" "$got_exit" "$stdout" "$stderr" "$status"
[ "$status" = pass ] || fail "case failed: program=$program input=$key" "case_failed_${program}_$key"
```

The named-control runner also checks that the selected Go test actually ran, passed, and did not skip (`scripts/verify-phase32.sh:148-160`). Extend that reached-test inventory for the M007 CRC, helper, bitwise, and Phase 37.1 owners from RESEARCH; keep scalar application tuples and structured conformance outputs on their existing consumers. A skipped required subtest, empty selector, missing row, or partial inventory must fail. Do not change the M006 default or replace per-case evidence with a program list.

### `scripts/check-milestone-binding.sh` (utility / gate, batch)

**Analog:** `scripts/check-phase27-receipts.sh` for a small POSIX shell receipt gate; reuse `scripts/check-receipts.sh` for generic row consistency.

The Phase 27 gate provides a small refusal pattern with named failures and explicit expected-revision comparison (`scripts/check-phase27-receipts.sh:14-29,79-85`):

```sh
set -eu
reject() {
  printf 'phase27 receipt gate: status=fail reason=%s\n' "$1" >&2
  exit 1
}
[ "$#" -eq 2 ] || reject "usage_two_logs_required"
```

Use this as a shell-boundary example only. The new seam must compare the selected run's full `head_sha` to the freeze SHA and each host log's `headsha` to that same SHA, while retaining run **attempt** identity and required job/artifact membership. Add the armed wrong-revision control. Parse bounded metadata as data; do not `source` or `eval` it.

`check-receipts.sh` is explicitly a consistency check, not provenance authentication (`scripts/check-receipts.sh:1-3`). It only requires a nonempty log `headsha` (`scripts/check-receipts.sh:82`); preserve its generic inventory and host checks, and put run/attempt authentication in the narrow new seam.

### `testdata/receipts/M007.manifest` (config / fixture, batch)

**Analog:** `testdata/receipts/M006.manifest` (tracked; exact manifest grammar and inherited baseline).

The current manifest header declares label, program set, and required controls, followed by exact tuple rows (`testdata/receipts/M006.manifest:1-5`):

```text
label=phase32
programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz
controls=guide,route-audit,codedigest,receipt-rules
# case checksum empty 0 1 -
```

Build M007 by retaining all 39 M006 case rows verbatim and adding the explicit CRC/helper/bitwise programs, cases, and named reached controls. Keep the M006 file itself unchanged. The producer's observed inventory and the generic receipt gate must agree with every declared row.

### `.github/workflows/ci.yml` (config / workflow, event-driven batch)

**Analog:** the existing two-host checks matrix and Phase 27/32 artifact steps in `.github/workflows/ci.yml` (tracked).

The Phase 32 upload runs after the producer and retains the artifact even when earlier steps fail (`.github/workflows/ci.yml:136-145`):

```yaml
- name: Upload Phase 32 host receipt
  if: ${{ always() }}
  uses: actions/upload-artifact@v4
  with:
    name: phase32-receipt-${{ matrix.os }}-${{ github.sha }}
    path: ${{ runner.temp }}/phase32/receipt-${{ matrix.os }}.log
    if-no-files-found: error
    retention-days: 14
```

Add the M007 receipt artifact to the existing Ubuntu/macOS checks matrix and retain upload-on-failure plus missing-file failure. The current receipt pipeline uses `tee`; enable Bash `pipefail` in the producer step so `tee` cannot hide a failed producer. Keep the existing full and race lanes as required evidence; do not add a redundant broad job.

### `internal/compiler/session/session_gate_differential_test.go` (test, request-response)

**Analog:** the existing differential tests for freeze and receipt shell gates (tracked).

Use its `phase33RunAtEnv` helper to execute a shell gate in an isolated working directory and capture exit/stdout/stderr (`internal/compiler/session/session_gate_differential_test.go:21-33`):

```go
func phase33RunAtEnv(t *testing.T, cwd, script string, environment []string, args ...string) phase33CommandResult {
    t.Helper()
    full := append([]string{"-c", `cd "$0" && script=$1 && shift && exec sh "$script" "$@"`, cwd, script}, args...)
    result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, full...)
    if err != nil && result.Exit == 0 {
        t.Fatalf("spawn %s: %v", script, err)
    }
    return phase33CommandResult{result.Exit, string(result.Stdout), string(result.Stderr)}
}
```

Add focused positive and wrong-revision cases for the binding seam using the existing table-driven gate style (`TestReceiptGateDifferential`, line 114 onward). Keep the fixture self-contained and assert the named refusal reason, not only a nonzero status.

### Final copied logs and bounded metadata (evidence artifacts, batch)

**Analog:** None in tracked source for persisted hosted logs/receipt documents. The final plan should choose exact filenames under `.planning/phases/38-milestone-receipts/`.

Record only the D-12 run fields, required job IDs/conclusions, artifact names/IDs/digests/expiry, and copied log paths/byte counts/SHA-256. Verify source provenance separately from generic receipt consistency and re-gate the exact copied bytes. Artifact URLs alone do not preserve durable evidence; do not mix hosts from separate runs or attempts.

## Shared Patterns

### Keep freeze, receipt consistency, and hosted provenance separate

**Sources:** `scripts/check-freeze.sh:1-3,38-43`; `scripts/check-receipts.sh:1-3,75-113`; new `scripts/check-milestone-binding.sh`.  
**Apply to:** binding checker, final copied receipt, and EVD-17 completion gate.

The freeze gate owns ancestry and codedigest-path drift. The receipt gate owns declared case/control inventories, host identity consistency, and codedigest agreement. Neither authenticates a hosted run or attempt. The new checker should make only the exact run/head/freeze/log bindings and attempt-specific completeness claims required by D-11/D-12. Also check worktree cleanliness around freeze preparation because the generic freeze gate covers committed ancestry, not uncommitted files.

### Manifest-driven, reached evidence

**Sources:** `scripts/verify-phase32.sh:71-76,148-170`; `testdata/receipts/M006.manifest:1-42`.  
**Apply to:** M007 manifest and host driver.

The manifest is the single source of required programs, exact cases, and controls. Verify exact selected test/subtest names and reject skipped, empty, missing, and partial inventory. Preserve the separation between default-build stdout/stderr/exit tuples, structured conformance, optimization test evidence, causal defect attribution, and physical-cleanup evidence.

### Hosted logs and artifact retention

**Source:** `.github/workflows/ci.yml:120-145`.  
**Apply to:** M007 CI step and the last-plan `.planning` ingest.

Keep both OSes on the existing jobs, upload raw host logs with `always()`, fail when a log is absent, and set `pipefail` for the `tee` pipeline. The final planning-only descendant may add evidence files under `.planning`; those copied bytes must still match the selected run and freeze.

### Workspace invariant

Preserve `.planning/config.json` byte-for-byte at SHA-256 `148ccc4c56db8516788c264216ae277d9860d43e00b92de299d8ccff2fd59521`; exclude it from staging.

## No Analog Found

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `.planning/phases/38-milestone-receipts/` copied host logs and bounded metadata | evidence artifact | batch | These are newly generated bytes for the selected hosted run; there is no reusable tracked source analog. |

## Metadata

**Analog search scope:** `scripts/`, `testdata/receipts/`, `.github/workflows/`, and `internal/compiler/session/`.  
**Tracked source analog paths named:** 7.  
**Pattern extraction date:** 2026-10-16.  
**Tracked-source gate:** every named source analog above was verified with `git ls-files`; no ignored Phase 27/32 receipt mirror path is emitted.
