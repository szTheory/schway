---
verification_role: proposal
---

# Phase 32: Clean-Checkout Guide and Final Receipts - Pattern Map

**Mapped:** 2026-10-12
**Files analyzed:** 21 (13 new, 8 modified, 1 deleted)
**Analogs found:** 20 / 21 (one role-match only: `scripts/codedigest.sh`)

All analog paths below were checked with `git ls-files --error-unmatch` and are tracked source. None is a gitignored mirror.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `internal/compiler/syntax/format.go` (mod) | utility (formatter) | transform | itself, loop at lines 34-42 | exact (in-place) |
| `internal/compiler/syntax/syntax_phase32_format_test.go` (new) | test | transform | `syntax_phase29_test.go` `TestPhase29FormatByteReads`, `syntax_test.go` `TestGeneratedLinearRoundTrips` | exact |
| `internal/compiler/session/session_phase30_prechange_test.go` (mod: `phase30FormatChanges`, `phase30RetiredPaths`, controls) | test (frozen-golden gate) | batch/compare | itself (`phase30NewCommittedPaths`, `phase30Compare`, `...IsNotInert`) | exact |
| `internal/compiler/session/session_phase28_manifest_test.go` (mod: `phase28Compare` extraction, `phase28RetiredEntries`, control) | test (frozen-golden gate) | batch/compare | `session_phase30_prechange_test.go` `phase30Compare` | role-match |
| `internal/compiler/session/session_phase32_retire_test.go` (new) | test | file-I/O + doc scan | `session_phase20_test.go` `TestPhase20ChecksumFrontier` (replaced), `phase20SHA256` | exact |
| `internal/compiler/session/session_phase20_test.go` (mod: delete test) | test | n/a | n/a | n/a |
| `examples/checksum.schway` (delete) | example | n/a | n/a | n/a |
| `examples/checksum/README.md` (mod) | guide (executable spec) | request-response (CLI) | `examples/phase27/README.md` clean-checkout block | exact |
| `cmd/schway/main_phase32_guide_test.go` (new) | test | request-response (subprocess) | `session_phase273_docs_test.go` (parse and target lint) + `main_phase31_test.go` (route probes) | role-match (merge of two) |
| `cmd/schway/main_phase32_routes_test.go` (new) | test | transform (AST enumeration) | `cmd/schway/main_phase31_test.go` (package main, `run()` access) | partial |
| `scripts/codedigest.sh` (new) | utility (script) | transform | `scripts/check-phase27-receipts.sh` `hash_list` | role-match |
| `scripts/verify-phase32.sh` (new) | script (producer) | batch | `scripts/verify-phase27.sh` | exact |
| `scripts/check-phase32-receipts.sh` (new) | script (gate) | batch/validate | `scripts/check-phase27-receipts.sh` | exact |
| `internal/compiler/session/session_phase32_receipt_test.go` (new) | test | batch/validate | `session_phase27_host_test.go` + `native/phase25_utility_test.go` `TestPhase25LivingRoadmap` | exact |
| `.github/workflows/ci.yml` (mod) | config | event-driven | itself, `evidence-aggregate` phase27 step | exact |
| `.planning/phases/32-.../32-RECEIPT.json` + logs (ingest) | data | file-I/O | `examples/phase27` committed receipts (see Phase 27 precedent in RESEARCH Finding 4) | role-match |
| `.planning/PROJECT.md`, `LANGUAGE-MATURITY.md`, `PRODUCT-ROADMAP.md` (mod) | docs | n/a | dated amendment convention | n/a |

## Pattern Assignments

### `internal/compiler/syntax/format.go` (utility, transform) - plan 32-01

**Analog:** the existing loop in `Format`, lines 34-42. Add a second guard in the same style immediately before `f.token(token, next)`.

**Existing guard to mirror** (lines 34-42):
```go
		if token.Kind == TokenIdentifier && next == TokenEqual && index > 0 && f.lineOpen && f.insideFunctionBody() {
			previous := tokens[index-1]
			if previous.Span.End <= token.Span.Start && bytes.Contains(tree.Source[previous.Span.End:token.Span.Start], []byte("\n")) {
				f.newline()
			}
		}
		f.token(token, next)
```

**Added guard** (RESEARCH Finding 1; `f.context() != "match"` is mandatory, otherwise 9 phase12/18 outputs regress):
```go
if token.Kind == TokenIdentifier && index > 0 && f.lineOpen && f.insideFunctionBody() && f.context() != "match" {
	previous := tokens[index-1]
	if (previous.Kind == TokenIdentifier || previous.Kind == TokenNumber || previous.Kind == TokenRParen) &&
		previous.Span.End <= token.Span.Start && bytes.Contains(tree.Source[previous.Span.End:token.Span.Start], []byte("\n")) {
		f.newline()
	}
}
```
Check that `f.context()` exists on `formatter` (grep before using; the prototype used it). Do not touch `}` / `else {` placement. Do not reformat committed sources.

---

### `internal/compiler/syntax/syntax_phase32_format_test.go` (test, transform)

**Analog A:** `internal/compiler/syntax/syntax_phase29_test.go` (package `syntax_test`).

**Imports and source helper** (lines 1-14):
```go
package syntax_test

import (
	"bytes"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

func phase29Source(body string) []byte {
	return []byte("module phase29.surface\n\nexport {\n  fn probe\n}\n\nfn probe(bytes: ByteView) -> U64 {\n" + body + "}\n")
}
```
Reuse `phase29Source` directly (same package) for the var-body table.

**Table plus idempotence pattern** (`TestPhase29FormatByteReads`, lines 75-99):
```go
for _, test := range []struct{ name, input, want string }{ ... } {
	t.Run(test.name, func(t *testing.T) {
		first := session.Format(phase29Source(test.input))
		if len(first.Diagnostics) != 0 { t.Fatalf("format diagnostics: %+v", first.Diagnostics) }
		if want := string(phase29Source(test.want)); string(first.Canonical) != want { t.Fatalf("Format = %q, want %q", first.Canonical, want) }
		second := session.Format(first.Canonical)
		if !bytes.Equal(first.Canonical, second.Canonical) { t.Fatalf("not idempotent: %q vs %q", first.Canonical, second.Canonical) }
	})
}
```
Table rows needed: `var r = r + a` newline `r` (identifier), `r = r % 7` newline `r` (number), `var x = byte_at(bytes, n)` newline `x` (call paren). Source inputs must be on separate lines (`"  var r = 1\n  r = r + a\n  r\n"`).

**Analog B (property body to copy for an optional `generatedVarProgram`):** `syntax_test.go` `TestGeneratedLinearRoundTrips` lines 601-650. Per-case checks: bytes preserved by CST, no parse diagnostics, canonical reparses, `Format` fixed point, `semanticTokens` equality, `withoutSpans(Program)` equality, `comments` equality, and `diagnosticCodes(session.Check(...))` equal before and after (comment at lines 641-643: spans move on reindent, so compare codes only).

**Corpus guard** (RESEARCH Pitfall 2): read the 5 committed sources from `testsupport.ProjectPath(...)` and format them as in-memory copies; also assert that the formatted output for every other corpus file is unchanged. No write to the repo.

---

### `internal/compiler/session/session_phase30_prechange_test.go` (test, frozen-golden compare) - plans 32-01 and 32-02

**Analog:** itself. Extend the existing allow-list pattern.

**Allow-list shape to copy** (lines 61-73):
```go
var phase30NewCommittedPaths = map[string]bool{
	// 2026-10-12 (Phase 30, D-27): ... the golden stays frozen (D-15).
	"examples/checksum/adler32.schway":     true,
	...
}
```
Add:
```go
// 2026-10-12 (Phase 32, DX-19, D-06): retired path; may be absent, failure if observed again.
var phase30RetiredPaths = map[string]bool{"examples/checksum.schway": true}
// 2026-10-12 (Phase 32, plan 32-01, D-10c): formatter fix changes format_sha256 for these 3 frozen entries.
var phase30FormatChanges = map[string]string{ /* path: NEW sha, computed AFTER the real fix */ }
```

**Compare function edit points** (`phase30Compare`, lines 237-275). In the golden loop at 246-254 (the `!ok` branch), skip when `phase30RetiredPaths[want.Path]`. In the field loop at 255-267, substitute the allow-listed `format_sha256` and report `stale phase30FormatChanges entry equals the golden` when equal (mirrors `phase28ClassChanges` at `session_phase28_manifest_test.go:591-594`). In the observed loop (269-273) add:
```go
if phase30RetiredPaths[entry.Path] {
	problems = append(problems, entry.Path+": retired path observed again")
}
```

**Negative control to extend** (`TestPhase30PreChangeGoldenIsNotInert`, lines 326-348): it asserts a clean baseline via `phase30Compare(golden, observed)`, then copies `golden.Entries`, flips `target.CheckSHA256[0]`, and requires non-empty problems. Add sub-controls in the same shape: (1) append a synthetic observed entry with a retired path and require a "retired path observed again" problem; (2) set a `phase30FormatChanges`-style entry equal to the golden and require the stale problem. These need `phase30Compare` to accept the maps as parameters or use package variables swapped in the test (keep it simple: package vars saved and restored with `t.Cleanup`).

Do not regenerate `testdata/phase30/prechange-corpus.json` (header: "never regenerated").

---

### `internal/compiler/session/session_phase28_manifest_test.go` (test, frozen-golden compare) - plan 32-02

**Analog:** `phase30Compare` (above) for the extracted-helper shape; the current inline loop is the source to extract.

**Allow-list pattern** (lines 230-244): `phase28NewlyAddedEntries` is a `map[string]phase28ManifestEntry` with dated comments. Add `phase28RetiredEntries = map[string]bool{"examples/checksum.schway": true}` with a dated comment alongside it.

**Inline compare to extract into `phase28Compare(golden phase28Manifest, observed []phase28ManifestEntry) []string`** (lines 572-608):
```go
for _, got := range observed {
	want, inGolden := goldenByPath[got.Path]
	added, isAdded := phase28NewlyAddedEntries[got.Path]
	changed, isChanged := phase28ClassChanges[got.Path]
	switch {
	case !inGolden && !isAdded:
		t.Errorf("%s: absent from the golden and not in phase28NewlyAddedEntries: %+v", got.Path, got)
	...
}
for path := range goldenByPath {
	if _, ok := observedByPath[path]; !ok {
		t.Errorf("%s: in the golden at %s but no longer observed", path, golden.Revision)
	}
}
```
Convert each `t.Errorf` to `problems = append(problems, fmt.Sprintf(...))`, change the final loop to `if phase28RetiredEntries[path] { continue }`, and add "retired entry observed again" for an observed retired path. `TestPhase28EntryExitClassManifest` (line 533) then calls `phase28Compare` and `t.Error`s each problem. Keep the write-golden branch (lines 536-553) unchanged. Add a new `TestPhase28EntryExitClassManifestIsNotInert` modelled on `TestPhase30PreChangeGoldenIsNotInert` (it does not exist today; CONTEXT's "extend" is wrong for Phase 28). Sort check at 569-571 stays outside the helper or moves with it. Note: `reflect.DeepEqual` and `sort` are already imported.

---

### `internal/compiler/session/session_phase32_retire_test.go` (test, file-I/O) - plan 32-02

**Analog:** `session_phase20_test.go` lines 1-60 (the test being replaced).

**Imports to copy** (lines 3-12):
```go
import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)
```

**Core pattern:**
```go
path := testsupport.ProjectPath("examples", "checksum.schway")
source, err := os.ReadFile(path)
...
if got := phase20SHA256(source); got != "ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154" { ... }
```
The new test inverts this: `os.Stat(path)` must return `errors.Is(err, os.ErrNotExist)`; `examples/checksum/checksum.schway` must exist; `filepath.Walk` over `examples/` and `testdata/` for `.schway` and hash each with `phase20SHA256` (tombstone digest constant moves to the new file). `phase20SHA256` (lines 56-59) lives in `session_phase20_test.go`: when deleting `TestPhase20ChecksumFrontier`, keep this helper (or move it); grep for other users first.

**Armed control:** write the tombstone bytes into `t.TempDir()` is not possible without the bytes (file deleted). Use a synthetic file whose digest is injected: make the scanner take a `tombstone string` parameter, run it on a temp file and pass that file's own digest. Same idea as the walker controls in `session_phase28_manifest_test.go:464-531` (`phase28ObserveCorpus` walks with `filepath.Walk`).

**Doc-hygiene scan:** "live" means lines not under a heading matching `^#{2,4} .*\d{4}-\d{2}-\d{2}`. Closest style: `phase25CheckEvidenceIndexLinks` in `native/phase25_utility_test.go:226-279` (line scan with a heading regex and an armed broken-input control at 186-190). Files: `PROJECT.md`, `LANGUAGE-MATURITY.md`, `PRODUCT-ROADMAP.md`, `examples/checksum/README.md`. Control: a historical line moved into a live section must be flagged; a line under a dated heading must be exempt.

---

### `examples/checksum/README.md` (guide, mod) - plan 32-03

**Analog:** `examples/phase27/README.md` "Clean-checkout run" block (as consumed by `phase273CleanCheckoutBlock`, which cuts on `"\n## Clean-checkout run\n"` then the next `"\n```sh\n"`). Phase 32 replaces that with marked fences:

````
```sh doc-test=clean-checkout
go build -o ./schway ./cmd/schway
mkdir -p .schway/checksum
./schway build examples/checksum/checksum.schway --manifest examples/checksum/file_bytes.bindings.json --output .schway/checksum/checksum
```
````
Rules: all `--output` and `app run` targets under `.schway/`; `mkdir -p` before the first `./schway build`; table row keys are the case names `empty zero wikipedia ff65536 ff65537 missing directory`; one `Earlier programs` pointer to `examples/phase27/README.md`; line 140 note becomes a pointer to the PROJECT.md amendment; add one known-limits line for `schway run` on a parameterised main linking the pending todo.

---

### `cmd/schway/main_phase32_guide_test.go` (test, request-response) - plan 32-03

**Analog A (parse and output lint, copy locally; `cmd/schway` cannot import `session_test`):** `session_phase273_docs_test.go`.

**Output-target checker to copy** (lines 51-84): `phase273GuideOutputTargets` splits lines with `strings.Fields` and switches on `-o`, `--output` (adds `next+".schway-build.json"`), `>`/`>>`, `--output=`, `-o=`, and `app run <artifact>`; `phase273TargetAllowed` does `path.Clean` then accepts only `.schway/` prefix or `schway`:
```go
func phase273TargetAllowed(target string) bool {
	cleaned := path.Clean(target)
	return strings.HasPrefix(cleaned, ".schway/") || cleaned == "schway"
}
```
Keep its bypass-control test (`TestPhase273GuideTargetCheckerRefusesBypasses`, lines 192-206) as the model for the parser negative controls: in-memory mutated guide strings, each must produce a failure.

**Block extractor style** (lines 32-46): `strings.Cut` on a heading, then the fence. Replace with a line scan over the regex `^```sh doc-test=([a-z0-9-]+)((?: [a-z]+=[a-z0-9_]+)*)$` through the closing fence, joining `\` continuations (RESEARCH Code Examples).

**Pre-fix negative-control pattern** (lines 15-26, 180-187): a const holding a known-bad block plus a test asserting each expected violation is reported.

**Analog B (route and helper conventions):** `main_phase31_test.go`.

**Imports and clang skip** (lines 1-12, 49-52):
```go
import ( "os"; "os/exec"; "path/filepath"; "strconv"; "strings"; "testing"
	"github.com/szTheory/schway/internal/compiler/testsupport" )
...
if _, err := exec.LookPath("clang"); err != nil {
	t.Skip("env:clang toolchain unavailable on this host")
}
```
Replace the bare skip with `phase32NativePolicy(err, os.Getenv)` (fail on `SCHWAY_REQUIRE_NATIVE=1` or `GITHUB_ACTIONS=true`, otherwise skip). Pure function plus `TestPhase32NativePolicy`.

**Error-row assertion pattern** (lines 66-72):
```go
missing := filepath.Join(t.TempDir(), "phase31-secret-segment", "absent.bin")
code, stdout, stderr := captureLangRun(t, []string{"app", "run", artifact, "--", missing})
if code != 65 || stdout != "" || stderr != "FileError.FileUnavailable\n" || strings.Contains(stderr, "phase31-secret-segment") { ... }
```
Phase 32 does the same through the built binary. Use `testsupport.ChecksumSecretSegment` for the segment and `testsupport.ChecksumCaseMatrix()` (`Name`, `Kind`, `Content`, `Exit`, `Stdout`, `Stderr`) to materialise inputs by case name. `testsupport.MaterializeChecksumCase(t, dir, c)` already creates the file, missing, and directory cases under the secret segment.

**Subprocess spawn:** `testsupport.RunCLIErr(ctx, "sh", os.Environ(), "-c", `cd "$0" && eval "$1"`, copyDir, line)` (`testsupport.go:159`; bounded streams, 2-minute `RunCLITimeout`, passes `scanUnboundedSpawns`). Substitute only the `PATH` token, single-quoted, never in build lines. Do not use `-trimpath`; keep the copy alive for the whole test. `captureLangRun` (`main_test.go:717`) is in-process and not suitable for the guide run, only for the JSON cross-check helpers if needed.

**Copy helper analog:** `phase30CopyExampleDir` (`main_phase30_test.go:15-36`) copies a flat dir with `os.ReadDir`/`os.WriteFile(..., 0o600)`. The Phase 32 copy needs a recursive copy driven by `git -C <root> ls-files -co --exclude-standard` (via `RunCLIErr`), skipping `ENOENT` and, optionally, `.planning`; fall back to a `filepath.Walk` that skips `.git` and `.schway`.

**JSON cross-check:** decode `examples/checksum/checksum.expected.json` (keys `schema`, `limit`, `error_exit`, `outcomes`, `cases[name,kind,exit,stdout,stderr,outcome,...]`); copy the struct from `internal/compiler/testsupport/checksum_matrix_test.go` `TestPhase31ExpectedMatrixJSON` (do not import; test files are not importable).

---

### `cmd/schway/main_phase32_routes_test.go` (test, AST transform) - plan 32-04

**Analog:** `main_phase31_test.go` for the `package main` setup. No test elsewhere in the repo enumerates routes by AST, so the enumeration itself comes from RESEARCH Finding 5 / D-10d.

**Enumeration shape:** parse `main.go` with `go/parser`; walk `run`, `runApplication`, `runApplicationRun`; collect string literals in `==` binary expressions whose other operand is `args[<int>]`; add `"--..."` literals from the dispatch and `extract*` helpers. Cross-check with `usageResult()` tokens (it omits `stats` and `build --manifest`; record both as audit rows). Entry programs: `filepath.Walk` `examples/` for `*.schway`, `syntax.Parse` and look for a function named `main`. Do not call `callgraph.EntryFunction` (restricted by `TestEntryFunctionCalledOnlyViaResolveEntry`) or any `cgen` emitter (`TestPhase16PublicEmitterConsumerInventory`). Table rows: `covered` (program name) or `not covered because <real reason>`; test fails on a route absent from the table and on a table row with no observed route.

---

### `scripts/codedigest.sh` (script, transform) - plan 32-04

**Analog (role-match):** `hash_list` in `check-phase27-receipts.sh` lines 37-44 (POSIX sh helper that fails by name).

**Recipe** (CONTEXT D-08a, RESEARCH Finding 4):
```sh
git -c core.quotePath=false ls-tree -r HEAD -- cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml \
  | grep -v '\.md$' | { shasum -a 256 2>/dev/null || sha256sum; } | cut -d' ' -f1
```
Add a `command -v` probe with a named failure if neither hasher exists. Go equivalent: same git command through `RunCLIErr`, drop lines with `strings.HasSuffix(line, ".md")`, rejoin with `"\n"` plus trailing newline, `sha256`, hex. The `ls-tree` reads HEAD, so `tree=clean` must be required separately.

---

### `scripts/verify-phase32.sh` (script, batch producer) - plan 32-04

**Analog:** `scripts/verify-phase27.sh` (copy and extend; do not edit the original).

**Header and failure idiom** (lines 8-14, 16-20):
```sh
set -eu
fail() {
	echo "phase27 verify: $1" >&2
	printf 'phase27 host result status=fail reason=%s\n' "$2"
	exit 1
}
for tool in go clang git uname mktemp rm cat sed cmp mkdir dirname printf; do
	if ! command -v "$tool" >/dev/null 2>&1; then fail "required tool is missing: $tool" "unavailable_tool_$tool"; fi
done
```

**Identity capture** (lines 22-42): `verify_root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)`, `mktemp -d` plus `trap ... EXIT HUP INT TERM`, `host_name` from `uname -s` (`Darwin`->macOS, `Linux`->Linux), `go env GOOS/GOARCH`, `clang -dumpmachine`, tree state from `git status --porcelain`. Replace `revision=` with `headsha=` and add `codedigest=$(sh scripts/codedigest.sh)`. Improve: record tree at start and end.

**`run_case` function** (lines 71-90) is the unit to copy: runs `"$schway" app run "$verify_tmp/$program" -- "$input"`, `cmp -s` stdout and stderr files, compares exit, prints
`phase27 case program=%s input=%s expected_exit=%s actual_exit=%s stdout=%s stderr=%s status=%s`. Change the prefix to `phase32`. For `checksum` and `adler_low`, use `build ... --manifest ... --output`.

**`run_control`** (lines 112-125): `go test -v -count=1 -run "$regex" "$pkg"`, fail on `--- SKIP:` or `no tests to run`, require `--- PASS: <name> ` for each named test. Controls for Phase 32: `TestPhase32GuideCleanCheckout`, `TestPhase32PublicRouteAudit`, etc. (only after the tests exist).

**Final row** (lines 130-132): case-count check, then `phase32 host result host=%s codedigest=%s tree=%s cases=%s controls=%s programs=checksum,adler_low,adler32_cli,sum_to_n,fizzbuzz status=pass`. Never reuse the `phase27 ` prefix (the phase27 gate treats unknown `phase27 ` rows as forgery).

---

### `scripts/check-phase32-receipts.sh` (script, validate) - plan 32-04

**Analog:** `scripts/check-phase27-receipts.sh` (copy and adapt).

**Header to keep, including the non-claim** (lines 12-15): "This gate checks consistency, not authenticity ... Authenticity must come from the hosted run record itself."

**Reject idiom and expected value** (lines 20-23, 34-35):
```sh
reject() { printf 'phase27 receipt gate: status=fail reason=%s\n' "$1" >&2; exit 1; }
expected_revision=${PHASE27_EXPECTED_REVISION:-$(git rev-parse HEAD)}
```
Phase 32: `expected_digest=${PHASE32_EXPECTED_CODEDIGEST:-$(sh scripts/codedigest.sh)}`.

**Field parser** (lines 49-51): `field LINE KEY` via `tr ' ' '\n' | sed -n "s/^$2=//p" | sed -n '1p'` (values contain no spaces).

**`check_log` structure** (lines 53-109) to follow in order: file exists and non-empty; reject any `status=(fail|incomplete|unavailable|unobserved)`; exactly one identity and one result row; identity first and result last; reject unknown row kinds with
`grep '^phase27 ' "$log" | grep -Ev '^phase27 (host identity|toolchain|case|control|host result) '`; toolchain row malformed check; host in `macOS|Linux`; digest equals expected (`digest_mismatch_$log`); `tree=clean`; result row agrees; exact-text case rows via `grep -Fxc -- "$want"` for every program and input; control rows regexed. Add named reasons for D-09: `digest_mismatch`, `log_digest_differs`, `tree_not_clean`, `host_missing`/`duplicate_host`, `program_missing_<name>`. End with the two-log combine (lines 111-115): `host_a != host_b` else `duplicate_host_$host_a`.

---

### `internal/compiler/session/session_phase32_receipt_test.go` (test, validate) - plan 32-04

**Analog A:** `session_phase27_host_test.go` for the gate mutation table and the literal known-good log builder.

**Independent known-good builder** (lines 26-52): `phase27ReceiptLog(t, host)` builds a literal log and measures only the file hashes, so the gate cannot pass by echoing its producer. Copy for `phase32ReceiptLog`.

**Gate runner** (lines 54-76):
```go
environment := append(os.Environ(), "PHASE27_EXPECTED_REVISION="+phase27GateRevision)
result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-phase27-receipts.sh"), paths[0], paths[1])
```
Use `PHASE32_EXPECTED_CODEDIGEST` and the new script. The `"\x00missing"` sentinel for an absent log is part of the pattern.

**Mutation table** (lines 133-167): a `mutate(log, old, replacement)` helper that fatals if the anchor is absent, a `without(prefix)` line dropper, and a table `{name, first, other, reason}` driving subtests that require `status=fail`, the named reason and no `status=pass`. Needed rows: mutated digest, flipped `tree=modified`, dropped host (one log plus empty), dropped program rows, swapped `programs=` list, unknown `phase32` row, altered case row. Also a mutated-code-file control (edit a byte in a temp clone, recompute digest, expect move; edit only `.md` or `.planning`, expect no move).

**CI text pin pattern:** `TestPhase27CIHostMatrix` (lines 173-210) slices the job with `strings.Index(workflow, "\n  evidence-aggregate:\n")` to `"\n  phase26-focused:\n"` and checks `strings.Contains` for each required fragment plus `continue-on-error` absence. Add `TestPhase32CIHostMatrix` the same way for the new step and upload names. Leave the phase27 tests untouched.

**Analog B (living-claim rule):** `native/phase25_utility_test.go` `TestPhase25LivingRoadmap` lines 293-306:
```go
for file, text := range map[string]string{"PRODUCT-ROADMAP.md": product, "LANGUAGE-MATURITY.md": maturity} {
	for _, want := range []string{"Phase 25", ..., "Historical receipts", ...} {
		if !strings.Contains(text, want) { t.Errorf("%s does not distinguish/include %q", file, want) }
	}
}
if strings.Contains(product, "Phase 25 passed on Linux and macOS") || strings.Contains(maturity, "...") {
	t.Fatal("living documents claim hosted Phase 25 evidence without a bound receipt")
}
```
It uses `phase25ReadRepoFile(t, ".planning", "PRODUCT-ROADMAP.md")`. D-10a is the stronger form: a fixed marker line (`M006 final hosted receipts: bound to codedigest <hex>` or `... reopened <date>`). Absent marker passes (no `t.Skip`); present marker requires `32-RECEIPT.json`, recomputed log hashes, and equal digests; stale digest plus "bound" fails; "reopened" passes. Include an in-memory control on doc text, like the broken-link control at `phase25_utility_test.go:186-190`.

**Meta-test registration:** each test must be a plain top-level `func TestX(t *testing.T)` (static index equality with `go test -list`); `32-*-SUMMARY.md` needs a `coverage:` block (see Shared Patterns).

---

### `.github/workflows/ci.yml` (config, event-driven) - plan 32-04

**Analog:** the phase27 step and upload in `evidence-aggregate` (lines 95-125).

**Step to duplicate** (verbatim shape):
```yaml
      - name: scripts/verify-phase27.sh
        shell: bash
        run: |
          mkdir -p "$RUNNER_TEMP/phase27"
          sh scripts/verify-phase27.sh 2>&1 | tee "$RUNNER_TEMP/phase27/receipt-${{ matrix.os }}.log"
      - name: Upload Phase 27 host receipt
        if: ${{ always() }}
        uses: actions/upload-artifact@v4
        with:
          name: phase27-receipt-${{ matrix.os }}-${{ github.sha }}
          path: ${{ runner.temp }}/phase27/receipt-${{ matrix.os }}.log
          if-no-files-found: error
          retention-days: 14
```
Add the `phase32` pair directly after it. Keep `fetch-depth: 0` (line 101), the `Require clang` step (lines 106-107), `os: [ubuntu-latest, macos-latest]`, no `continue-on-error`. Optionally add `env: SCHWAY_REQUIRE_NATIVE: "1"` to the `go test` steps (D-04a). Top-level `concurrency` has `cancel-in-progress: true` (line 47-49): dispatch once, on the branch ref. Every edit here is inside the codedigest, so land them before the freeze.

---

### Files with no real analog

| File | Role | Data Flow | Reason |
|---|---|---|---|
| route enumeration by Go AST in `main_phase32_routes_test.go` | test | transform | The repo has no AST-based route enumerator; use `go/parser` per RESEARCH Finding 5. Scan style can borrow from `scanUnboundedSpawns` in `internal/compiler/native/native_test.go`. |
| `32-RECEIPT.json` | data | file-I/O | No JSON receipt exists; M005 used plain-text `.receipt` logs. Use the schema in RESEARCH Finding 4 (`schway.m006-receipt/1`). |

## Shared Patterns

### Dated allow-list against a frozen golden, never a re-pin
**Source:** `session_phase30_prechange_test.go:61-73` (`phase30NewCommittedPaths`), `session_phase28_manifest_test.go:230-244` and `:591-594` (`phase28NewlyAddedEntries`, stale-entry check on `phase28ClassChanges`).
**Apply to:** `phase30RetiredPaths`, `phase30FormatChanges`, `phase28RetiredEntries`.
Every entry carries a dated comment naming the phase and decision, and has an armed stale/observed-again control.

### Armed negative control for every gate
**Source:** `TestPhase30PreChangeGoldenIsNotInert` (lines 326-348); `TestPhase273GuideTargetCheckerRefusesBypasses`; `TestPhase27HostReceiptGate` mutation table.
**Apply to:** guide parser, retirement scan, receipt gate, living-claim rule, route audit.
Run the checker on a clean baseline first (`phase30Compare` must return no problems), then on one in-memory mutation that must fail.

### Bounded subprocess spawn only through testsupport
**Source:** `testsupport.go:159-200` (`RunCLIErr` -> `runCLI`: `exec.CommandContext`, `WithTimeout(RunCLITimeout)`, `boundedWriter` on both streams).
**Apply to:** every Go test that runs `git`, `sh`, `go build` or the built binary. `scanUnboundedSpawns` fails any raw `exec.Command`. `exec.LookPath("clang")` is allowed.

### Output targets confined to `.schway/`
**Source:** `session_phase273_docs_test.go:51-84` (`phase273GuideOutputTargets`, `phase273TargetAllowed`); `.gitignore` ignores only `/.schway/`, `/schway`, `/.gsd/`.
**Apply to:** README blocks, `verify-phase32.sh` temp outputs (use `mktemp -d` as at `verify-phase27.sh:24-25`).

### Receipt row grammar and named rejection
**Source:** `verify-phase27.sh:59-62,86-87,131-132` (producer rows) and `check-phase27-receipts.sh:20-23,49-51,69-70` (reject, field parser, unknown-row rejection).
**Apply to:** `verify-phase32.sh`, `check-phase32-receipts.sh`, `phase32ReceiptLog`. Use the `phase32 ` prefix only.

### Skip policy for clang
**Source:** `main_phase31_test.go:50-52`.
**Apply to:** all new native-dependent tests, through the single pure `phase32NativePolicy`.

### SUMMARY `coverage:` and planning-doc conventions
**Source:** `summary_coverage_test.go` (`TestSummaryCoverageGroundedness`), `verification_groundedness_test.go`, 31-VALIDATION.md.
**Apply to:** every `32-NN-PLAN.md` (`verification_role: proposal` in frontmatter), `32-VALIDATION.md` (must be `validated`, never `draft`; any `go test -run` span must resolve to existing tests, so pin commands only after the tests land), every `32-NN-SUMMARY.md` (`coverage:` entries with `ref: "<file>#<TestName>"`, `status: pass`; the push-checkpoint plan is the only `human_judgment: true` with a rationale).

### Dated historical records are never rewritten
**Source:** CONTEXT code_context "Established Patterns"; AGENTS.md "Living language roadmap".
**Apply to:** PROJECT.md amendment (dated 2026-10-12, Phase 32, DX-19), live-row pointers only (`LANGUAGE-MATURITY.md:297,304,412`, `PRODUCT-ROADMAP.md:272,298`, README:140). Leave phase dirs, `milestones/`, `research/` and `LANGUAGE-MATURITY.md:180` untouched. Update the counts at `LANGUAGE-MATURITY.md:304` (163->162 programs, 5,389->5,352 lines) for `TestLanguageMaturityCountsAreCurrent`.

## Ordering constraints for the planner

1. The `Format` change and `phase30FormatChanges` must land together (hashes computed after the final fix). 32-01 edits `session_phase30_prechange_test.go` before 32-02 does.
2. 32-02 and 32-03 both touch `examples/checksum/README.md:140`; give the line to exactly one plan.
3. All files under `cmd internal native examples scripts testdata go.mod .github/workflows/ci.yml` (including the receipt test, route audit and scripts) must be final before the freeze. The ingest plan writes only `.planning/**`.
4. No new `.schway` file may be added in the retirement plan (counts assume 162 programs).

## Metadata

**Analog search scope:** `internal/compiler/{syntax,session,native,testsupport}`, `cmd/schway`, `scripts/`, `.github/workflows/ci.yml`, `examples/phase27`, `examples/checksum`
**Files read for excerpts:** 11 (plus the CONTEXT and RESEARCH inputs)
**Pattern extraction date:** 2026-10-12

## PATTERN MAPPING COMPLETE
