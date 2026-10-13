# Phase 33: Carried Debt, Bounded Capture Read, and Generic Freeze Gate - Pattern Map

**Mapped:** 2026-10-13
**Files analyzed:** 27 new/modified/deleted
**Analogs found:** 26 / 27 (one partial: the sparse-file mutant catcher)

All analog paths were checked with `git ls-files` (tracked source). Paths under `.planning/milestones/` are tracked archived docs, not runtime mirrors. Line numbers are from the working tree on 2026-10-13.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `scripts/check-freeze.sh` (new) | gate script | request-response | `scripts/check-phase32-freeze.sh` | exact (generalize) |
| `scripts/check-receipts.sh` (new) | gate script | request-response/transform | `scripts/check-phase32-receipts.sh` | exact (generalize) |
| `scripts/codedigest.sh` (mod: `--rev`, `--paths`, `--changed`) | utility script | transform | itself (`scripts/codedigest.sh`) + the git calls in `check-phase32-freeze.sh:42-45` | exact |
| `scripts/verify-phase32.sh` (mod: read manifest) | producer script | batch | itself, lines 21-70, 84-87, 112 | exact |
| `testdata/receipts/M006.manifest` (new) | config/data | file-I/O | `scripts/verify-phase32.sh:21,25-65` (rows copied verbatim) | exact |
| `scripts/check-phase32-{freeze,receipts}.sh` (delete) | gate script | - | n/a | n/a |
| `internal/compiler/session/session_freeze_gate_test.go` (rename of `session_phase32_freeze_test.go`) | test | request-response | `session_phase32_freeze_test.go` | exact |
| `internal/compiler/session/session_receipt_gate_test.go` (rename of `session_phase32_gate_test.go`) | test | request-response | `session_phase32_gate_test.go` | exact |
| differential test (`TestReceiptGateDifferential`, `TestFreezeGateDifferential`) | test | request-response | `runPhase32Gate` + `phase32FreezeRun` | role-match |
| `TestReceiptGateArchivedM006` + M006 freeze rows | test | file-I/O | `TestPhase32ReceiptLivingClaims` real-checkout + `phaseArtifactGlob` | role-match |
| `TestNoCheckPhase32Caller` | test (static) | file-I/O | `session_phase32_retire_test.go` (not read; see Metadata), `TestNoDuplicatedExitLiterals` (`internal/compiler/exitcode/scan_test.go:205`) | role-match |
| `session_phase32_receipt_test.go` (mod: generalize marker, extend doc set) | test | transform | itself, lines 204-405, 436-470 | exact |
| `internal/compiler/native/native_app.go` (mod: `readApplicationCapture`) | service | file-I/O | `readApplicationReceipt` (~649) | exact |
| `internal/compiler/native/native_phase33_capture_test.go` (new) | test | file-I/O | `TestPhase27OutputEvidenceExactLimit` (`native_app_test.go:587`) | exact |
| `internal/compiler/interp/interp.go` (mod: `isAbortLikeDefect` + 2 callers) | service | transform | itself (439-461; callers 319, 371) | exact |
| `internal/compiler/execution/defect_reasons.go` (new) | utility (constants) | n/a | `execution.MaxApplicationEvidenceBytes` constant in `execution.go:13` | partial |
| `executionpeer.go`, `cgen_program.go`, `session.go` (mod: use constants, in place) | service | transform | their own literals (see below) | exact |
| `internal/compiler/interp/interp_phase33_defect_test.go` (new) | test | transform | `interp_phase321_defect_test.go` + `core_internal_test.go` seam test | exact |
| `internal/compiler/native/native_phase321_defect_parity_test.go` or new `native_phase33_defect_parity_test.go` | test | request-response | `TestPhase321DefectParityWithLiveOwner` + `native_phase31_controls_test.go` helpers | exact |
| `cmd/schway/main_phase321_run_input_test.go` (mod, IN-02) | test | request-response | itself, line ~31 | exact |
| `internal/compiler/session/summary_coverage_test.go` (mod, kind pin) | test | transform | itself, entry switch lines 67-87; patch in commit `31d71e94` | exact |
| `.planning/milestones/M006-phases/28-*/28-VALIDATION.md` (new) | doc | n/a | `29-VALIDATION.md` (+ `32.1-VALIDATION.md` frontmatter) | exact |
| `.planning/milestones/M006-phases/28-*/28-SECURITY.md` (new) | doc | n/a | `29-SECURITY.md` | exact |
| `33-CARRIED-DEBT-DISPOSITION.md` (new) | doc | n/a | `32.1-REVIEW-DISPOSITION.md` | role-match |
| `32.1-REVIEW-DISPOSITION.md` (edit) | doc | n/a | itself | exact |
| `.planning/{LANGUAGE-MATURITY,PRODUCT-ROADMAP,PROJECT,MILESTONES,STATE}.md` (D-01 reopen) | doc | n/a | marker grammar in `session_phase32_receipt_test.go:207` | exact |
| `33-VALIDATION.md` (new) | doc | n/a | `32.1-VALIDATION.md` | exact |

## Pattern Assignments

### `scripts/check-freeze.sh` (gate script, request-response)

**Analog:** `scripts/check-phase32-freeze.sh` (47 lines, whole file read)

**Core (lines 11-47), the evaluation order to preserve:** reject() -> arity -> sha shape over ALL args -> SAME_AS == CANDIDATE -> `cat-file -e` both -> `merge-base --is-ancestor` -> changed paths -> pass line.
```sh
set -eu
phase32_freeze_paths="cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml"
reject() {
	printf 'phase32 freeze gate: status=fail reason=%s\n' "$1" >&2
	exit 1
}
[ "$#" -ge 2 ] || reject "usage"
for sha in "$@"; do
	case "$sha" in
	*[!0-9a-f]*) reject "bad_sha" ;;
	esac
	[ "${#sha}" -eq 40 ] || reject "bad_sha"
done
freeze=$1
candidate=$2
shift 2
for same in "$@"; do
	[ "$same" = "$candidate" ] || reject "sha_mismatch"
done
git cat-file -e "${freeze}^{commit}" 2>/dev/null || reject "unknown_commit"
git cat-file -e "${candidate}^{commit}" 2>/dev/null || reject "unknown_commit"
git merge-base --is-ancestor "$freeze" "$candidate" || reject "not_descendant"
changed=$(git log --format=%H "${freeze}..${candidate}" -- $phase32_freeze_paths ':(exclude)*.md') || reject "git_log_failed"
[ -z "$changed" ] || reject "digest_path_changed_after_freeze"
git diff --quiet "$freeze" "$candidate" -- $phase32_freeze_paths ':(exclude)*.md' || reject "digest_path_changed_after_freeze"
printf 'phase32 freeze gate: status=pass freeze=%s candidate=%s\n' "$freeze" "$candidate"
```

**Changes to apply (from CONTEXT D-04..D-08 and RESEARCH skeleton):**
- Read manifest first (`$1`), then arity floor 3. `printf '%s freeze gate: status=fail reason=%s\n' "$label" "$1" >&2` (label as data, never in the format position). Empty label: omit the leading space (RESEARCH Open Question 1).
- Add `root=$(git rev-parse --show-toplevel 2>/dev/null) || reject no_repo_root; cd "$root"` (WR-01; pattern from `scripts/codedigest.sh:23-25`). Make MANIFEST absolute before the `cd` (Pitfall 4).
- Replace lines 41-45 (inline path list, `:(exclude)*.md`) with `changed=$(sh "$root/scripts/codedigest.sh" --changed "$freeze" "$candidate" 2>/dev/null) || reject git_log_failed; [ -z "$changed" ] || reject digest_path_changed_after_freeze` (Pitfall 5: `2>/dev/null` keeps one stderr line).
- POSIX only: no `[[`, `local`, arrays (Ubuntu `sh` is dash).

---

### `scripts/check-receipts.sh` (gate script, request-response/transform)

**Analog:** `scripts/check-phase32-receipts.sh` (124 lines, whole file read)

**Header/reject/field helpers** (lines 19-53): copy verbatim, then parameterize the label.
```sh
reject() {
	printf 'phase32 receipt gate: status=fail reason=%s\n' "$1" >&2
	exit 1
}
[ "$#" -eq 2 ] || reject "usage_two_logs_required"
...
field() { # field LINE KEY -> value of KEY=value (values never contain spaces)
	printf '%s\n' "$1" | tr ' ' '\n' | sed -n "s/^$2=//p" | sed -n '1p'
}
```

**Digest source (lines 31-37), generalize env name:**
```sh
if [ -n "${PHASE32_EXPECTED_CODEDIGEST:-}" ]; then digest_source=env; else digest_source=computed; fi
expected_digest=${PHASE32_EXPECTED_CODEDIGEST:-$(sh scripts/codedigest.sh)}
[ -n "$expected_digest" ] || reject "expected_codedigest_unknown"
```
New: `EXPECTED_CODEDIGEST`; never recompute when set.

**Manifest read, replaces lines 39-45** (producer-file sed becomes manifest sed; `producer_missing` becomes `manifest_missing`):
```sh
mf_get() { sed -n "s/^$1=//p" "$manifest" | sed -n '1p'; }
label=$(mf_get label); want_programs=$(mf_get programs); controls=$(mf_get controls)
case_table=$(grep '^# case ' "$manifest" || true)
```

**Label-coupled hard-codes to replace (Pitfall 1), lines 62-69, 92, 96, 98, 103-105:**
```sh
[ "$(grep '^phase32 ' "$log" | sed -n '1p' | cut -c1-20)" = "phase32 host identit" ] || reject "identity_not_first_$log"
[ "$(grep '^phase32 ' "$log" | tail -n 1 | cut -c1-19)" = "phase32 host result" ] || reject "result_not_last_$log"
[ "$(field "$result" controls)" = 4 ] || reject "result_control_count_$log"
[ "$(grep -c '^phase32 control ' "$log" || true)" = 4 ] || reject "control_row_count_$log"
for control in guide route-audit codedigest receipt-rules; do
```
Replace with prefix tests on `"$label "`, and `controls=` count/names from the manifest. Keep `$log` untouched for reason suffix bytes; use an absolute read copy after `cd "$root"` (Pitfall 4).

**Case-row check (lines 97-102), keep intact:** the `while read -r _hash _case program key want_exit _stdout _stderr` loop with `grep -Fxc -- "$want" "$log"` must equal `1`, with `$label` in the expected string.

**Pass line (lines 123-124), keep field order:** `... status=pass hosts=%s,%s codedigest=%s programs=%s cases=%s digest_source=%s`.

---

### `scripts/codedigest.sh` (utility, transform)

**Analog:** itself (35 lines). Key lines to refactor:
```sh
27: listing=$(git -c core.quotePath=false ls-tree -r HEAD -- cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml) || fail "git_ls_tree_failed"
30: digest=$(printf '%s\n' "$listing" | grep -v '\.md$' | $hasher | cut -d' ' -f1)
34: [ "${#digest}" -eq 64 ] || fail "bad_digest"
```
- Hoist the path list into one variable; `--paths` prints it; `--rev REV` replaces the `HEAD` at line 27; `--changed A B` runs `git -c core.quotePath=false log --format= --name-only A..B -- $paths` and `git -c core.quotePath=false diff --name-only A B -- $paths`, then `grep -v '\.md$' | grep -v '^$' || true` (guard grep exit 1).
- Line 34: derive length from `printf '' | $hasher | cut -d' ' -f1` (assumption A1; test asserts equality with real digest length).
- Keep `fail()` shape and reason names. Parse flags before the hasher/root checks only where needed; the `--paths` branch should not require a hasher.

---

### `scripts/verify-phase32.sh` (producer, batch)

**Analog:** itself. Replace:
- line 21 `phase32_programs=...` and lines 25-65 `# phase32-cases-begin/end` block, by reading `testdata/receipts/M006.manifest`.
- line 84 `digest_paths="cmd internal ..."` -> `sh scripts/codedigest.sh --paths`.
- line 112: `case_table=$(sed -n '/^# phase32-cases-begin$/,/^# phase32-cases-end$/p' "$verify_root/scripts/verify-phase32.sh" | grep '^# case ' || true)` -> `grep '^# case ' "$verify_root/testdata/receipts/M006.manifest"`.
Keep the script name, the four `run_control` calls, and the test function names (`TestPhase32GuideCleanCheckout`, `TestPhase32PublicRouteAudit`, `TestPhase32CodeDigest`, `TestPhase32ReceiptLivingClaims`); archived logs name them (Pitfall 8). `.github/workflows/ci.yml` runs this script; leave the step alone.

---

### `testdata/receipts/M006.manifest` (data)

**Analog:** the producer's current block (copy `# case ...` rows byte-for-byte). Syntax per RESEARCH:
```
label=phase32
programs=checksum,adler_low,adler32_cli,adler32_fixture,byte_index_oob,sum_to_n,fizzbuzz
controls=guide,route-audit,codedigest,receipt-rules
# case checksum empty 0 1 -
...(39 rows)
```
Do not name it `*.schway` (corpus walkers). It is inside the digest paths (`testdata`), so it moves the digest. Pin: a test hashes it and checks every row appears in both archived logs; executor confirms rows equal `git show 9f8f878f:scripts/verify-phase32.sh` (RESEARCH says the script is unchanged since).

---

### `session_receipt_gate_test.go` / differential test (test, request-response)

**Analog:** `internal/compiler/session/session_phase32_gate_test.go`

**Literal log builder (lines 85-115), keep and add a label parameter for the second-label row:** `phase32ReceiptLogAt(t, host, digest, headsha)` builds a literal known-good log from `phase32GateCases` (lines 38-80), independent of the producer.

**Runner pattern (lines 117-135), adapt to run old and new on the same paths:**
```go
dir := t.TempDir()
paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
for index, content := range []string{first, second} {
	if content == "\x00missing" { continue }
	os.WriteFile(paths[index], []byte(content), 0o600)
}
environment := append(os.Environ(), "PHASE32_EXPECTED_CODEDIGEST="+expectedDigest)
result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-phase32-receipts.sh"), paths[0], paths[1])
```
New script: pass `EXPECTED_CODEDIGEST` and manifest arg. Compare `Exit`, `Stdout`, `Stderr` byte-equal; every catalog row also asserts the literal `reason=` string (Pitfall 7).

**Temp-repo fixture pattern:** `phase32GateTempRepo` (line ~137): copies `scripts/codedigest.sh` into a temp git repo; extend to copy the new scripts + manifest.

**Gaps to fill (RESEARCH catalog):** about 18 receipt reasons have no existing test (rows #5, 6, 9-15, 18-21, 23, 25-27 and gate-level `case_table_empty`, `programs_line_missing`, `expected_codedigest_unknown`, `usage_two_logs_required`). Manifest-variant rows write a temp manifest; for the old script, a temp tree with a variant `scripts/verify-phase32.sh`.

**Table retarget:** `TestPhase32ProducerCaseTable` (line 381) reads the producer; point it at the manifest.

---

### `session_freeze_gate_test.go` (test, request-response)

**Analog:** `session_phase32_freeze_test.go`

**Run helper (lines 13-23), keep; new arg shape:**
```go
full := append([]string{"-c", `cd "$0" && sh scripts/check-phase32-freeze.sh "$@"`, repo}, args...)
result, err := testsupport.RunCLIErr(context.Background(), "sh", nil, full...)
```
New: `sh scripts/check-freeze.sh MANIFEST "$@"`. Subdirectory row (Pitfall 3, WR-01): `testsupport.RunCLIErr(ctx, "sh", nil, "-c", `cd "$0/cmd" && sh ../scripts/check-freeze.sh "$@"`, repo, args...)` expects `digest_path_changed_after_freeze`.

**Fixture build (lines 60-110):** `build(t)` creates base/freeze/plan/late commits using `phase32Git`; write `scripts/check-freeze.sh`, `scripts/codedigest.sh`, and the manifest into the temp repo. `expectPass`/`expectReject` helpers (lines 36-48) reusable. The `md-only-change-after-freeze-accepted` subtest (line ~124+) is extended to assert `sh scripts/codedigest.sh --changed F C` is empty for `.md`-only and non-empty for `.go` (IN-03).

**Real-repo M006 rows:** full SHAs `0aac0a7b4801292b3f54f7ef57ed5fb83374c08e`, `9f8f878fc5328f953eac6ed04a716478ae986fab`, `31d71e9428eeb3aa89cd376b2a6fd5047bb4e69d`, `7ab2f48e28f2401f0545e9c414c496795651090b`. Start each with `git cat-file -e <sha>^{commit}`; fail loudly with the `git fetch --tags` remedy, never skip.

---

### `TestNoCheckPhase32Caller` (static test)

**Analogs:** `internal/compiler/session/session_phase32_retire_test.go` (exists, not read in detail by this mapper; the planner should open it first for the existing file-walk idiom) and `internal/compiler/exitcode/scan_test.go:205`. Build the needle by concatenation (`"check-" + "phase32"`) so the test file does not match itself. Per RESEARCH Finding 2, scan tracked non-`.planning/` files (use `git ls-files`), and also fail if `scripts/check-phase32-freeze.sh` or `...-receipts.sh` exists.

---

### `session_phase32_receipt_test.go` (living-claim generalization)

**Analog:** itself. Regex at line 207:
`^M006 final hosted receipts: reopened ([0-9]{4}-[0-9]{2}-[0-9]{2}) \(historical codedigest ([0-9a-f]{64})\)\.?$`. Bound: `^M006 final hosted receipts: bound to codedigest ([0-9a-f]{64}) \(run ([0-9]+)\)\.?$`. Line-parse trims `" \t>*-"` then `HasPrefix(line, "M006 final hosted receipts:")` (~253-259). Real-checkout reads docs at ~442. Parameterize the prefix by milestone label; port all 28 table rows; add the real-checkout row (archive + reopened markers + drifted digest passes). Do not rename `TestPhase32CodeDigest` or `TestPhase32ReceiptLivingClaims`. Helpers to reuse: `phase32CodeDigest` (:25), `phase32ShDigest` (:43), `phase32Git` (:52), `phaseArtifactGlob("32-*", "32-RECEIPT.json")` (`session_test.go:2579`).

---

### `internal/compiler/native/native_app.go` (`readApplicationCapture`, file-I/O)

**Analog:** `readApplicationReceipt` at lines 649-659.
```go
func readApplicationReceipt(path string) (BuildReceipt, error) {
	file, err := os.Open(path)
	if err != nil { ... }
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxApplicationReceiptBytes+1))
```
**Call site to change (lines 362-380):**
```go
data, readErr := os.ReadFile(capturePath)
limit := r.evidenceCapacityLimit()
switch {
case readErr == nil && len(data) > limit:
	report.CaptureStatus = EvidenceStatusCapacityExhausted
case readErr == nil:
	capture, captureErr := decodeApplicationCapture(data, limit)
```
New helper (RESEARCH): `func readApplicationCapture(path string, limit int) ([]byte, error)` with `os.Open` + `io.ReadAll(io.LimitReader(file, int64(limit)+1))`. Reorder so `limit` is computed before the read. Keep the `switch` unchanged. No `os.Stat`. A missing or read error must stay `EvidenceStatusIncomplete` (existing `phase22AssertIncompleteEvidence` rows).

---

### `native_phase33_capture_test.go` (test, file-I/O)

**Analog:** `TestPhase27OutputEvidenceExactLimit` (`native_app_test.go:587-601`)
```go
capture := phase22CompleteCapture(t)
body := fmt.Sprintf("printf 'app\\n'\nprintf %%s %s > \"$SCHWAY_APP_EVIDENCE_PATH\"\n", shellQuote(capture))
artifact := writePhase22EvidenceScript(t, body)
runner := DefaultRunner()
runner.evidenceLimit = len(capture)
reportPath := filepath.Join(t.TempDir(), "exact-capacity-report.json")
outcome, report, err := runner.RunApplicationWithEvidence(context.Background(), artifact, "7", reportPath, EvidenceEvents, &stdout, &stderr)
```
- exactly-limit: `evidenceLimit = len(capture)` (as above); limit+1: `len(capture)-1`. Package `native` (internal), no `t.Parallel`.
- Mutant catcher: child script creates the sparse file with POSIX `dd if=/dev/null of="$SCHWAY_APP_EVIDENCE_PATH" bs=1 seek=1073741824` (verified by RESEARCH), run via `RunApplicationWithEvidence` with the default limit (64 KiB, budget `16*limit` = 1 MiB; measure overhead once, assumption A3), `runtime.ReadMemStats` `TotalAlloc` delta. No direct analog for the allocation assertion (partial). Hand-run the `os.ReadFile` mutant once and paste failing output in the SUMMARY.
- Helpers: `writePhase22EvidenceScript` (:714), `phase22CompleteCapture`, `shellQuote`.

---

### `internal/compiler/interp/interp.go` (`isAbortLikeDefect`, transform)

**Analog:** itself, lines 439-461 (current predicate, scans every function, ignores `LiveResources` and `Output`) and callers at ~319 and ~371:
```go
if len(result.LiveResources) != 0 && !isAbortLikeDefect(program, result) {
	return Execution{}, fmt.Errorf("model execution reached a terminal outcome with live local resources %v", result.LiveResources)
}
```
Caller loops (lines ~296-304) already count `acquires` over `operation.Foreign.Mode == "acquire"`; also record `ownerOpID = operation.ID` there (exactly one by the `acquires != 1` guard). Keep the error text prefix. Look the op up with `findFunctionByID(program, final.FunctionID)` (`interp.go:1502`), not a global scan. Target shape: RESEARCH "Abort-like predicate shape (D-12)" code block (with seams `abortLikeAcceptLive`, `abortLikeReasonMatches`).

**Seam model:** `internal/compiler/core/core.go:509`
```go
var decoderRequireNonEmpty = func(value string) bool { return value != "" }
```
with the doc comment explaining the same-package-test-only discipline (core.go:500-508); copy that comment style for the two new seams.

---

### Reason constants refactor (`execution/defect_reasons.go`, `interp.go`, `executionpeer.go`, `cgen_program.go`, `session.go`)

**Analogs:** the existing literals (VERIFIED by RESEARCH): `interp.go:1705/1731/1754`, `executionpeer.go:392/409/426`, `cgen_program.go:233/238/243` and ~2725-2757, `session.go:535-539`. Values must be exactly `"U64 addition overflow"`, `"U64 remainder by zero"`, `"byte index out of bounds"`. Pattern for new file: a small `const` block in `package execution` beside `MaxApplicationEvidenceBytes` (execution.go:13).
**Constraint (RESEARCH Finding 3):** edit `session.go` one line for one line; `testdata/phase16/public-emitter-consumers.json` pins line numbers. Run `TestPhase16PublicEmitterConsumerInventory` after. Emitted C must stay byte-identical (`TestPhase272AcceptedCorpusByteIdentity`, `TestPhase273ConformanceCorpusByteIdentity`, `TestPhase28PrunedEmitMatchesGolden`).

---

### `interp_phase33_defect_test.go` (test, transform)

**Analog 1:** `interp_phase321_defect_test.go` (same package `interp`)
```go
func phase321Program(kinds map[string]core.OperationKind) core.Program {
	operations := []core.LinearOperation{}
	for id, kind := range kinds { operations = append(operations, core.LinearOperation{ID: id, Kind: kind}) }
	return core.Program{Functions: []core.Function{{Linear: &core.LinearBody{Operations: operations}}}}
}
func phase321Defect(finalID, output string) Execution {
	return Execution{
		Outcome: execution.Outcome{Kind: execution.OutcomeDefect},
		Events:  []Event{{Kind: "foreign.called", ID: "x:event"}, {Kind: "function.defected", ID: finalID, Output: output}},
	}
}
```
These builders have no function IDs and no `LiveResources`; the narrowed predicate needs `Function.ID`, event `FunctionID`, `Output`, `LiveResources: []string{ownerOpID}`. The existing accepted/refused tables (lines 29-64) must be ported to the 3-arg form (keep every refused row). Refusal rows to add: wrong reason, extra live `[owner, "x:op:9"]`, missing owner, wrong function, two-function same-op-ID falsifier.

**Analog 2 (mutant test), `core_internal_test.go:15-45`:**
```go
original := decoderRequireNonEmpty
decoderRequireNonEmpty = func(string) bool { return true }
defer func() { decoderRequireNonEmpty = original }()
```
Assert the pre-mutation refusal, then that the widened seam accepts a row production refuses, for each of `abortLikeAcceptLive` and `abortLikeReasonMatches`.

---

### File-route parity pin (`native_phase33_defect_parity_test.go`)

**Analog:** `native_phase321_defect_parity_test.go` (package `native_test`): `phase321InlineFixture` splices a callee between `fn adler32(` and `fn main(` (lines 17-32); `TestPhase321DefectParityWithLiveOwner` variants `{name, callee, caseID, prefix}` (lines 34-60), gated by `phase29RequireClang(t)`; expected stderr = `prefix + interpreter final Output + "\n"`. File-route build from `native_phase31_controls_test.go:26-125` helpers `phase31ControlCopy`, `phase31ControlMutateLibrary` (replaces `fn adler32`), `phase31ControlBuild`, `phase31ControlNative`, `phase31ControlInterp`.

---

### `cmd/schway/main_phase321_run_input_test.go` (IN-02)

**Analog:** itself, line ~31: `strings.Contains(stdout+stderr, dir) || strings.Contains(stdout+stderr, "n:")`. Replace the on-disk `examples/sum_to_n.schway` read with the inline fixture in RESEARCH (`fn main(zq_param_marker: U64) -> U64`), assert `zq_param_marker` and `dir` absent, plus control `strings.Contains(string(fixture), "zq_param_marker")`. Keep the final engine-JSON-equality assertion.

---

### `summary_coverage_test.go` (kind pin)

**Analog:** itself, entry switch at lines 67-87 (`case "human_judgment":`, `"rationale"`, `"ref"`, `"status"`). Apply the existing patch: `git show 31d71e94 -- internal/compiler/session/summary_coverage_test.go` (adds a `kinds` slice, `case "kind":`, enum switch over `unit, integration, e2e, automated_ui, manual_procedural, other`, and a seeded `bad-kind` control `kind: unit` -> `kind: manual`). Add the comment citing `gsd-core/bin/lib/coverage.cjs:65-67` (`VALID_KINDS`). Existing SUMMARYs use only `unit` and `manual_procedural`, so none turn red. New Phase 33 SUMMARYs are scanned (`summaryCoverageFirstPhase = 32`).

---

### Phase 28 `28-VALIDATION.md` (doc)

**Analog:** `.planning/milestones/M006-phases/29-byte-views-and-the-adler-32-library/29-VALIDATION.md`

**Frontmatter (lines 1-10), copy:**
```yaml
---
phase: "28"
slug: "exit-classes-verify-report-and-entry-pruning"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# Reconstructed after execution from the 28-01..28-08 PLAN/SUMMARY set (State B).
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-13"
---
```
Section order (grep): Test Infrastructure, Sampling Rate, `## Per-Requirement Verification Map` (NOT `Per-Task Verification Map`; that heading triggers grading), Wave 0 Requirements, Manual-Only Verifications, Validation Sign-Off, Validation Audit. Table columns: `Requirement | Plans | Behavior | Test Type | Automated Command | File Exists | Status`. Row content comes from the RESEARCH "Verified evidence inventory" (DX-17, DX-18, EVD-13, CR-01 rows with package + test names, all verified to exist). Say "retroactive/reconstructed" and "Hosted CI: not run for this revision".

### Phase 28 `28-SECURITY.md` (doc)

**Analog:** `29-SECURITY.md`. Frontmatter (`status: verified`, `threats_open: 0`, `asvs_level: 1`, `created`), then `## Trust Boundaries` (table `Boundary | Description | Data Crossing`), `## Threat Register` (table `Threat ID | Category | Component | Severity | Disposition | Mitigation | Status`, IDs like `T-28-01`), `## Accepted Risks Log`, `## Notes`, `## Security Audit Trail`, `## Sign-Off`. Sources: 28-REVIEW, 28-REVIEW-DISPOSITION, 28-VERIFICATION. Boundaries per RESEARCH: source refusal vs tool failure exit class (`errors.Join`), report path -> filesystem (stale report removal), relative path -> project root (G-28-3), entry pruning -> C byte identity. State "retroactive, no new tests".

### `33-VALIDATION.md` (doc)

**Analog:** `32.1-VALIDATION.md` frontmatter comment: `# status lifecycle: planned (plan-phase; TestPhase20ValidationLifecycle forbids committed drafts) → validated (...)`. Seed `status: planned`, never `draft`. Use `## Requirement → Test Map` heading.

### Disposition docs

**Analog:** `32.1-REVIEW-DISPOSITION.md` frontmatter keys `findings[].{id,severity,disposition,title}`, `open: 5`, `total: 5`, plus a table `Finding | Severity | Disposition | Source`. Vocabulary is `open|fixed|skipped|deferred`; use `fixed` (IN-01 with Source "Covered by TestPhase32PublicRouteAudit (no code change needed)"), update `open:` to 0, and replace the "re-take" paragraph with a one-line note that D-01/D-03 supersede it. `33-CARRIED-DEBT-DISPOSITION.md` is a new file (not named `*-REVIEW-DISPOSITION.md`): one row per item: finding, fixed/covered, witness test, plan.

### D-01 living-doc markers

Use exactly (RESEARCH Finding 1; no `, run N` suffix, own line):
`M006 final hosted receipts: reopened 2026-10-13 (historical codedigest b9253457705a224a0580d74896e8e5bd3821446a1d467626b7d611ca0d1c4b5c)` in `.planning/LANGUAGE-MATURITY.md:239` and `.planning/PRODUCT-ROADMAP.md:10`. Put `run 38335175306` in the next prose sentence. Do not start any line in PROJECT.md or MILESTONES.md with `M006 final hosted receipts:`. Re-run `TestPhase25LivingRoadmap` after editing.

## Shared Patterns

### Reason-line contract for gates
**Source:** `scripts/check-phase32-freeze.sh:15-18`, `check-phase32-receipts.sh:21-24`
**Apply to:** `check-freeze.sh`, `check-receipts.sh`
One stderr line `<label> <kind> gate: status=fail reason=<name>`, exit 1, empty stdout. `reject()` always calls `printf '%s ...' "$label" "$1" >&2`. Gates check consistency, not authenticity (keep the header comment, `check-phase32-receipts.sh:15-18`).

### Resolve repo root then `cd`
**Source:** `scripts/codedigest.sh:23-25`
**Apply to:** both new gates (WR-01)
```sh
root=$(git rev-parse --show-toplevel 2>/dev/null) || fail "no_repo_root"
[ -n "$root" ] || fail "no_repo_root"
cd "$root" || fail "no_repo_root"
```

### Subprocess in tests
**Source:** `testsupport.RunCLIErr(ctx, "sh", env, args...)` (all of `session_phase32_*_test.go`)
**Apply to:** every new test that spawns a process; there is a repo lint on raw `exec`. No `Dir` parameter, so use `sh -c 'cd "$0" && ...' dir args...`.

### Temp git repo fixtures
**Source:** `phase32Git`, `phase32ShDigest`, `phase32CodeDigest` (`session_phase32_receipt_test.go:25-60`), `build(t)` in the freeze test.
**Apply to:** differential, freeze, `.md`-only tests.

### Meta-tests to run after every plan
`TestPhase16PublicEmitterConsumerInventory`, `TestVerificationGroundedness`, `TestPhase20ValidationLifecycle`, `TestValidationRowGradesAreEarnedOverArchivedCorpus`, `TestSummaryCoverageGroundedness`, `TestPhase32ReceiptLivingClaims`, `TestPhase32CodeDigest`. Focused `go test -run` only (STATE).

## No Analog Found

| File | Role | Data Flow | Reason |
|---|---|---|---|
| Memory-budget assertion (`runtime.ReadMemStats` `TotalAlloc` delta) in `native_phase33_capture_test.go` | test | file-I/O | No existing allocation-bounded test; shape from RESEARCH "Budget numbers". The surrounding harness has analogs. |

## Metadata

**Analog search scope:** `scripts/`, `internal/compiler/{session,native,interp,core,execution}`, `cmd/schway`, `.planning/milestones/M006-phases/{28,29,32,32.1}`, `.planning/phases/33-*`
**Files read:** about 14 (scripts x4, native_app.go ranges, interp.go ranges, core.go/core_internal_test.go ranges, five test files in part, 29-VALIDATION/29-SECURITY/32.1-VALIDATION/32.1-REVIEW-DISPOSITION heads)
**Not read in detail:** `session_phase32_retire_test.go`, `session_phase32_ci_test.go`, `main_phase321_run_input_test.go`, `native_phase31_controls_test.go` (cited from RESEARCH's verified descriptions)
**Pattern extraction date:** 2026-10-13
