# Phase 6: Agent Feedback and Performance Ratification - Pattern Map

**Mapped:** 2026-09-06
**Files analyzed:** ~20 new/modified files (from 06-CONTEXT.md D-06-01..D-06-32 and 06-RESEARCH.md)
**Analogs found:** 15 / 20 (5 explicitly greenfield — see "No Analog Found")

All analog paths below were confirmed git-tracked (`git ls-files`), not gitignored mirrors —
this repo has no `.gsd/capabilities/` mirror layer, so every path cited is the real tracked source.

## File Classification

| New/Modified File | Role | Data Flow | Closest Analog | Match Quality |
|---|---|---|---|---|
| `cmd/lang/main.go` (+`explain`, +`query` arms) | controller (CLI dispatch) | request-response | `cmd/lang/main.go`'s existing `debug-map`/`interface`/`evidence --validate` arms (same file, same function `run`) | exact |
| `internal/compiler/protocol/protocol.go` (`lang.explain/0`, `lang.query/0` structs) | model/schema | transform | `protocol.DebugMapSummary`/`InterfaceSummary`/`EvidenceSummary` + `Result` fields (same file) | exact |
| new `internal/compiler/session/session_phase6_explain.go` — `ExplainCommandFile` | service (command orchestration) | transform (synthesize DAG) | `session.DebugMapCommandFile` (`session.go:1143-1195`) | exact (per-cold-invocation synthesis shape) |
| new `internal/compiler/session/session_phase6_query.go` — `QueryCommandFile` + ID-vocabulary dispatcher | service | request-response (bounded, paginated) | `debugmap.Resolve` (`debugmap.go:292-299`) for the lookup shape; genuinely new join logic (see Pitfall below) | role-match (join layer itself is new) |
| `internal/compiler/diagnostic/diagnostic.go` (`Repair` gains `Span`/`Replacement`/`Applicability`) | model | transform | same file's existing `Repair`/`ErrorWithRepairs` identity-bearing-field split (`diagnostic.go:27-30,58-83`) | exact |
| new `internal/compiler/session/session_phase6_cause.go` (or similar) — cause-DAG synthesis over `diagnostic.Causes` | service | transform (graph synthesis) | `debugmap.Build` (`debugmap.go:161-260`, per-cold-invocation join) + `session.go:1279-1288`'s two-call determinism-compare pattern for `lane:deterministic` | role-match |
| new `internal/compiler/cache/cache.go` + `cache_test.go` | service (new leaf package) | file-I/O + CRUD (content-addressed store) | **No direct analog** — closest structural precedent is `internal/compiler/reduce` as a dependency-free leaf package with its own `_test.go` sibling (imports only `core`), not a functional analog | no analog (greenfield) |
| `internal/compiler/session/risk_lanes.json` + `session_phase6_risklanes.go` + audit | config (checked-in registry) + service (audit) | CRUD (lookup) | `internal/compiler/session/qlt01_registry.json` + `qlt01.go` (full file, 294 lines) | exact |
| `internal/compiler/session/qlt02_budget_manifest.json` + `session_phase6_budget.go` + audit | config + service | CRUD (lookup) + event-driven (probe cross-check) | `internal/compiler/session/qlt01_registry.json` + `qlt01.go` (same as above) | exact |
| `protocol.Lane` (+`cache_status`,`selection_reason`,`machine_id`,`gate_verdict`,`cold_or_warm`,`stage_breakdown`) / `protocol.Metrics` (+`cache_inputs_reused_count`) | model | transform | same file's existing `Lane`/`Metrics` structs (`protocol.go:24-30,76-86`) | exact |
| 12-site `"lang.verify-lane/0"` → `/1` literal bump | migration (mechanical, cross-file) | transform | itself — see full site list below | exact (already enumerated) |
| new Go statistics helper (p50/p95/CoV/demotion) | utility | batch/transform | **No Go analog exists** — `scripts/verify-phase2.sh`'s `observe()` shell function (full text below) is the behavioral spec, not an importable library | no analog (greenfield; shell is the spec) |
| host-identity probe for `machine_id` | utility | request-response (bounded subprocess) | `evidence.runToolProbe`/`defaultFacts` (`evidence.go:103-167`) | exact |
| 5 defect injectors (match/move/borrow/cleanup/stale-evidence) | test fixture + service | event-driven (mutation) | **cleanup**: `session.ReleaseOmissionMutationRunner` (`session.go:128-179`) reused verbatim. **match/move/borrow**: new `.lang` fixtures — no runner analog, only the marker-mutation-kill *shape* from `AliasFactMutationRunner` (`session_phase5_alias.go:44-127,211-350`). **stale-evidence**: `session.ValidateEvidenceCommandFile` (`session.go:979-1002`) is the locator/repair path, already shipped | exact (cleanup) / role-match (others) |
| new `cmd/lang-repair/main.go` (repair driver binary) | controller (standalone binary) | request-response (subprocess-only) | `cmd/lang/main.go`'s own subprocess-free dispatch shape is the *style* precedent; there is no existing subprocess-spawning CLI in the tree to copy the "spawn `lang --json` and parse stdout only" mechanic from — this exact mechanic is new | role-match (style only) |
| import-boundary test for `cmd/lang-repair` (D-06-28) | test | transform (static analysis over AST) | `corevalidate.TestValidatorImportsStayIndependent` (`corevalidate_endpoint_internal_test.go:74-101`) and `corevalidate.TestAttributeValidatorImportsStayIndependent` (`corevalidate_exclusive_test.go:205-...`) | exact |
| `scripts/verify-phase6.sh` | test/config (shell gate script) | batch | `scripts/verify-phase5.sh` (full file, 116 lines) for skeleton; `scripts/verify-phase2.sh` (49 lines) for the `observe()` sampling loop | exact |
| new held-out `.lang` defect fixture corpus (`testdata/phase6/` or similar) | test fixture | file-I/O | `testdata/phase5/` layout (11 fixture files) + `cmd/lang/main.go`'s `isPhase5Corpus` marker-file dispatch (`main.go:152-164`) | exact |

## Pattern Assignments

### `cmd/lang/main.go` — new `explain`/`query` arms (controller, request-response)

**Analog:** same file's existing arms.

**Dispatch pattern to copy** (`main.go:38-64`, add two more `if` blocks in this exact shape):
```go
if len(args) == 2 && args[0] == "check" {
    return runCheck(args[1], jsonMode)
}
...
if len(args) == 2 && args[0] == "debug-map" {
    return runDebugMap(args[1], "", jsonMode)
}
if len(args) == 3 && args[0] == "debug-map" {
    return runDebugMap(args[1], args[2], jsonMode)
}
```
`query` needs the flag-bearing variant `interface export|core|check` uses (fixed positional args)
— but since `query` takes optional `--kind=`/`--depth=`/`--cursor=`/`--json` flags beyond the
addressing argument, follow `extractJSON`'s pattern (`main.go:216-230`) of stripping recognized
flags from `args` before the fixed-arity dispatch runs, rather than inventing a new flag parser.

**Handler pattern to copy** (`main.go:203-214`, `runDebugMap`):
```go
func runDebugMap(source, query string, jsonMode bool) int {
    result, err := session.DebugMapCommandFile(source, query)
    if err != nil {
        return emit(problemResult("debug-map", protocol.StatusOperational, "tool.read_failed", "unable to read input"), jsonMode, false)
    }
    return emit(result, jsonMode, false)
}
```
`runExplain`/`runQuery` should follow this exact three-line shape: call into `session.*CommandFile`,
map any error through `problemResult(command, protocol.StatusOperational, "tool.read_failed", ...)`,
otherwise `emit(result, jsonMode, false)`.

**Usage string** must be extended (`main.go:263-265`, `usageResult`) with the new command grammar
from D-06-05, same string-concatenation style.

---

### `internal/compiler/protocol/protocol.go` — `lang.explain/0`, `lang.query/0` (model, transform)

**Analog:** `DebugMapSummary`/`DebugMapEntry` (`protocol.go:57-74`) and `EvidenceSummary`
(`protocol.go:32-36`) — both are `Schema string` + typed payload fields, attached to `Result` as
an `omitempty` pointer field, and folded into `Finalize()`'s identity struct via a content-hash
of the marshalled summary (mirrors `InterfaceID`/`DebugMapID` at `protocol.go:149-158`):
```go
// Source: protocol.go:70-74, the exact shape to copy for a new /0 summary struct
type DebugMapSummary struct {
    Schema  string          `json:"schema"`
    Entries []DebugMapEntry `json:"entries"`
}
```
```go
// Source: protocol.go:154-158, the exact identity-folding shape for the new summary
if result.DebugMap != nil {
    encodedDebugMap, _ := json.Marshal(result.DebugMap)
    debugMapSum := sha256.Sum256(encodedDebugMap)
    identity.DebugMapID = hex.EncodeToString(debugMapSum[:12])
}
```
`ExplainSummary`/`QuerySummary` follow this identically: new struct, new `*ExplainSummary`/
`*QuerySummary` field on `Result` (omitempty), new `ExplainID`/`QueryID` in `Finalize()`'s
anonymous identity struct, new rendering block in `human()` (`protocol.go:244-249` is the
`DebugMap` render-block template to copy).

**D-06-32 verification target** — write a test proving `Finalize()`'s identity struct
(`protocol.go:113-166`) still excludes `Metrics` entirely and reduces `Lane` to
`lane.ID+":"+lane.Status` (line 160) after every Phase 6 field addition. This is a NEW test, not
new plumbing — the struct already satisfies the constraint by construction.

---

### `session_phase6_explain.go` — `ExplainCommandFile` (service, transform)

**Analog:** `session.DebugMapCommandFile` (`session.go:1143-1195`, full function reproduced above
in research) — copy its exact shape: `readBoundedFile` → `syntax.Parse` → `check.Program` →
`corevalidate.Validate` → build the new artifact (here: cause DAG) with a `context.WithTimeout`
guard → `completeCommand(result, started, work)`. The 2-second `context.WithTimeout` at
`session.go:1168` is the bounding precedent for `explain`'s own per-invocation deadline.

**Determinism test pattern to copy** — `session.go:1279-1288`'s two-call-compare shape for
`lane:deterministic` (run twice, assert byte-identical) is exactly what D-06-02's "determinism
across cold invocations is a testable obligation" needs; do not invent a new comparison harness.

**Error-code routing pattern** — `debugmap.Error{Code string}` (`debugmap.go:90-103`) is the
`{Code}`-only typed-failure shape every new package in this phase should reuse for its own
fail-closed errors (cache miss, budget-manifest audit failure, etc.), matching
`evidence.ValidationError`/`originvalidate.Error`.

---

### `session_phase6_query.go` — `QueryCommandFile` + 5-vocabulary join (service, request-response)

**Analog for the lookup primitive:** `debugmap.Resolve` (`debugmap.go:292-299`):
```go
func Resolve(built Map, operationID string) Entry {
    for _, entry := range built.Entries {
        if entry.OperationID == operationID {
            return entry
        }
    }
    return Entry{OperationID: operationID, Availability: NotCaptured}
}
```
Copy the "loop, match, else return an honest absence value" shape — but **this is a single-field
lookup, not the five-vocabulary join D-06-01 needs.** Pitfall (confirmed by RESEARCH): do not plan
"wire query to debugmap.Resolve" as if it already does the join. The new dispatcher must recognize
an ID's vocabulary by prefix/shape (`diagnostic:`, `control:`, `lane:`, debugmap's own
`{functionID}:op:{n}`/`:point:{n}`, or an evidence ID/digest) and route to the right existing
lookup, converging on one `NotCaptured`-shaped answer — genuinely new join code.

**Honest-absence vocabulary to reuse:** `debugmap.Availability` (`debugmap.go:58-77`) —
`Available`/`OptimizedOut`/`NotCaptured` — is the exact three-value enum `query`/`explain` should
reuse rather than minting a new one.

---

### `diagnostic.go` — `Repair` additive fields (model, transform)

**Analog:** the file's own existing identity-bearing/non-identity-bearing split:
```go
// Source: diagnostic.go:58-83, ErrorWithRepairs — the exact pattern:
// only repairs[i].Kind enters the identity struct; Detail (prose) does not.
repairKinds := make([]string, len(repairs))
for index, repair := range repairs {
    repairKinds[index] = repair.Kind
}
identity := struct {
    Schema      string
    Code        string
    Span        Span
    Causes      []Cause
    RepairKinds []string
}{Schema: Schema1, Code: code, Span: span, Causes: causes, RepairKinds: repairKinds}
```
Adding `Span *Span`, `Replacement string`, `Applicability string` to the `Repair` struct
(`diagnostic.go:27-30`) follows this exactly: extend the struct fields as `omitempty`/pointer,
do NOT add them to `identity`'s `RepairKinds` derivation — `RepairKinds` must keep deriving from
`repair.Kind` only. `Schema1 = "lang.diagnostic/1"` (line 13) is already the frozen additive
version; no new schema constant is needed (confirmed by RESEARCH Pitfall 4).

---

### `internal/compiler/session/qlt01.go` — the registry-audit pattern (reused twice: `risk_lanes.json` and `qlt02_budget_manifest.json`)

**Full pattern, extracted in full since RESEARCH flags this as the single most reusable asset:**

Embed + parse (`qlt01.go:9-13,42-49`):
```go
import (
    ...
    _ "embed"
)

//go:embed qlt01_registry.json
var qlt01RegistryBytes []byte

func LoadQLT01Registry() ([]QLT01Row, error) {
    var rows []QLT01Row
    if err := json.Unmarshal(qlt01RegistryBytes, &rows); err != nil {
        return nil, fmt.Errorf("qlt01: failed to parse embedded registry: %w", err)
    }
    return rows, nil
}
```
Row shape + exactly-one-of union (`qlt01.go:32-40`):
```go
type QLT01Row struct {
    SpikeID          string               `json:"spike_id"`
    ControlID        string               `json:"control_id"`
    ControlMechanism string               `json:"control_mechanism"`
    LiveDescendant   *QLT01LiveDescendant `json:"live_descendant"`
    Waived           *QLT01Waiver         `json:"waived"`
}
```
Live cross-check against a **live-derived** set, never a hand-copied literal (`qlt01.go:51-71`):
```go
func AllShippedControlIDs() []string {
    seen := make(map[string]bool)
    var all []string
    for _, control := range Phase4RequiredControls() { ... }
    for _, control := range Phase5RequiredControls() { ... }
    return all
}
```
Audit function returning typed failures (`qlt01.go:96-189`, `AuditQLT01Registry`) — the
neither/both-required-field check, the stale-reference check against the live set, the
on-disk-fixture-existence check (`os.Stat`), and the empty-registry-is-a-hard-failure check are
all directly transferable to `risk_lanes.json` (cross-check declared lane IDs against actual lane
IDs `VerifyCorpus`/`verifyOwnedCorpus`/etc. produce) and `qlt02_budget_manifest.json` (cross-check
declared `machine_id`s against a live probe output, per D-06-16's own explicit instruction).

Lane wrapper (`qlt01.go:241-294`, `VerifyQLT01Registry`/`QLT01LaneFromRows`) — one work unit per
row inspected plus one per cross-check performed, `LaneResult{ID, Status, Controls,
RecomputedWork, Fired, ExpectedEscapes}` independent of `protocol.Lane` so the file has no
dependency on how the caller wires it in. Copy this exact independence — do not import
`protocol` into the new registry files themselves.

---

### 12-site `"lang.verify-lane/0"` → `/1` bump — full enumerated site list (migration)

Confirmed by grep this session — do not re-derive, use this list directly:

| # | File | Line |
|---|------|------|
| 1 | `internal/compiler/session/session.go` | 1255 |
| 2 | `internal/compiler/session/session.go` | 1406 |
| 3 | `internal/compiler/session/session.go` | 1549 |
| 4 | `internal/compiler/session/session.go` | 1870 |
| 5 | `internal/compiler/session/session.go` | 1889 |
| 6 | `internal/compiler/session/session.go` | 1916 |
| 7 | `internal/compiler/session/session.go` | 1960 |
| 8 | `internal/compiler/session/session.go` | 2035 |
| 9 | `internal/compiler/session/session_phase5.go` | 113 |
| 10 | `internal/compiler/session/session_phase5_mismatch.go` | 301 |
| 11 | `internal/compiler/session/session_phase5_mismatch.go` | 314 |
| 12 | `internal/compiler/session/session_phase5_sanitize.go` | 146 |

Every one of these is a `protocol.Lane{Schema: "lang.verify-lane/0", ...}` composite literal
inside one of 5 independently-implemented `addLane` closures (`session.go`'s `VerifyCorpus`,
`verifyOwnedCorpus`, `verifyBorrowedCorpus`, `verifyForeignCorpus`, and
`VerifyPhase5ControlsAndWork`) — per RESEARCH's Anti-Pattern note, **do not** refactor these into
one shared helper as part of this bump; edit each literal in place. Recommend introducing
`const LaneSchema1 = "lang.verify-lane/1"` in `protocol.go` beside `Schema`/`Schema1`-style
diagnostic constants, then mechanically replacing all 12 literals plus every new lane this phase
adds. Budget this as its own dedicated task — RESEARCH explicitly warns against sizing it equal
to the one-line `protocol.Schema` (`lang.command/0"`) bump, which is genuinely a single constant.

**Precedent for how the actual `/0`→`/1` migration was done before:** `lang.diagnostic/0`→`/1`
(`diagnostic.go:11-14`, two co-existing constants `Schema`/`Schema1`, old callers keep using
`Error()`/`Schema`, new callers use `ErrorWithRepairs()`/`Schema1`) and `lang.evidence/0`→`/1`
(`evidence.go:28-31`, same two-constant co-existence, `ForeignDigest` added as a trailing
`omitempty` field on the `/1` identity struct only, never inserted into `/0`). Both are the
precedent for "coordinate an additive bump without breaking `/0` byte-identity" — copy this
two-constant-coexistence shape for `protocol.Schema`/`LaneSchema1` too.

---

### Host-identity probe for `machine_id` (utility, request-response)

**Analog, full text** (`evidence.go:146-167`, `runToolProbe`):
```go
func runToolProbe(parent context.Context, command commandFactory, name string, arguments ...string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cmd := command(ctx, name, arguments...)
    var stdout, stderr boundedProbeWriter
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()
    if errors.Is(ctx.Err(), context.DeadlineExceeded) {
        return nil, &ToolProbeError{Code: "evidence.tool_timeout", Err: ctx.Err()}
    }
    if stdout.total > MaxToolProbeBytes {
        return nil, &ToolProbeError{Code: "evidence.tool_stdout_truncated", Err: fmt.Errorf("tool stdout exceeded %d bytes", MaxToolProbeBytes)}
    }
    ...
}
```
`MaxToolProbeBytes = 64 * 1024` (`evidence.go:36`) is the exact 64 KiB-plus-one bound to reuse
for the Clang-digest probe (D-06-07) and the `machine_id` probe (D-06-17) — do not write a second
bounding mechanism. `defaultFacts` (`evidence.go:107-127`) is the composition precedent: probe
`clang --version` and `-dumpmachine`, trim, assemble into a `Facts`-shaped struct. `machine_id`'s
`{os, arch, cpu_model, logical_cores, go_version, clang_version}` hash follows the same shape but
adds `runtime.GOOS`/`runtime.GOARCH`/`runtime.NumCPU()`/`runtime.Version()` (all stdlib, no probe
needed for those) alongside one `runToolProbe` call for Clang identity.

**Important distinction (RESEARCH Code Examples):** `defaultFacts` captures a **version string**,
not a **content-hash digest**. D-06-07 wants a probed digest of the Clang binary itself — this is
new work layered on the existing bounded-subprocess mechanism (hash the binary's own bytes or a
richer probe output), not something already done.

---

### Defect injectors (D-06-25)

**cleanup — reuse verbatim:** `session.ReleaseOmissionMutationRunner` (`session.go:128-179`,
full text):
```go
const releaseMarker = "/* lang:release-site */"

type ReleaseOmissionMutationRunner struct {
    runner        native.Runner
    mu            sync.Mutex
    optimizations []string
}

func (r *ReleaseOmissionMutationRunner) Run(ctx context.Context, cSource, optimization string, inputs []string) (native.Result, error) {
    lines := strings.Split(cSource, "\n")
    matched := -1
    for index, line := range lines {
        if strings.Contains(line, releaseMarker) {
            matched = index
        }
    }
    if matched == -1 {
        return native.Result{}, &native.ToolError{Code: "native.backend_control_invalid", Err: fmt.Errorf("release mutation marker count is 0, want at least 1")}
    }
    mutated := strings.Join(append(append([]string(nil), lines[:matched]...), lines[matched+1:]...), "\n")
    ...
}
```
This is exactly "targets a missing-`OpRelease` defect" — D-06-25 says use it directly, and this
IS its full implementation, ready to drive as-is.

**Fail-closed marker-count guard — the model for D-06-27's third anti-theater guard:** the
`matched == -1` branch above (refuse when the marker disappears, `native.backend_control_invalid`)
is the exact shape; `AliasFactMutationRunner` (`session_phase5_alias.go:44-127`) documents the same
discipline in its own doc comment ("refuses to run when it does not appear on..."). Copy this
refuse-rather-than-silently-pass shape for every injector's own marker-mutation-kill test.

**match/move/borrow:** no existing runner analog — these need new held-out `.lang` fixtures at
source granularity (remove an arm / inject use-after-move / inject a loan conflict). The
*fixture-corpus layout* analog is `testdata/phase5/` (11 `.lang`/`.json`/`.c` files, flat
directory) plus `cmd/lang/main.go`'s `isPhase5Corpus` (`main.go:152-164`):
```go
func isPhase5Corpus(corpus string) bool {
    _, err := os.Stat(filepath.Join(corpus, "restrict_borrow.lang"))
    return err == nil
}
```
A new `testdata/phase6/` corpus (or similarly named) with its own characteristic marker file and
a parallel `isPhase6Corpus`-shaped dispatch keeps `session.go` untouched, per the sibling-file
convention.

**stale-evidence:** already-shipped surface — `session.ValidateEvidenceCommandFile`
(`session.go:979-1002`, full text read this session) is both the locator (its `evidence.Validate`
call reports the mismatch) and, combined with `session.EvidenceCommandFile`, the repair path
(recapture a manifest that binds). No new session code needed for the locator/repair halves; only
the injector (re-touch source after manifest capture) and the driver-side classification logic
are new.

---

### Import-boundary test for `cmd/lang-repair` (D-06-28)

**Analog, full text** (`corevalidate_endpoint_internal_test.go:74-101`,
`TestValidatorImportsStayIndependent`):
```go
func TestValidatorImportsStayIndependent(t *testing.T) {
    dir := testsupport.ProjectPath("internal", "compiler", "corevalidate")
    entries, err := os.ReadDir(dir)
    if err != nil {
        t.Fatal(err)
    }
    fileSet := token.NewFileSet()
    for _, entry := range entries {
        if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
            continue
        }
        file, err := parser.ParseFile(fileSet, filepath.Join(dir, entry.Name()), nil, parser.ImportsOnly)
        if err != nil {
            t.Fatal(err)
        }
        for _, imported := range file.Imports {
            path := strings.Trim(imported.Path.Value, `"`)
            if strings.HasSuffix(path, "/compiler/check") || strings.HasSuffix(path, "/compiler/ast") {
                t.Fatalf("%s imports %s, which corevalidate must never depend on", entry.Name(), path)
            }
        }
    }
}
```
The mechanism is `go/parser.ParseFile(..., parser.ImportsOnly)` over every non-test `.go` file in
a directory, asserting no import path matches a forbidden suffix. `TestOracleImportsStayIndependent`
(`pathoracle/pathoracle_test.go:46-...`) and `TestOriginValidatorImportsStayIndependent`
(`originvalidate/originvalidate_test.go:26-...`) and
`TestAttributeValidatorImportsStayIndependent` (`corevalidate_exclusive_test.go:205-...`) are three
more instances of the identical shape — this is a well-established, four-times-repeated pattern in
this tree, not a one-off. D-06-28's new test should live beside `cmd/lang-repair/main.go` (a
`cmd/lang-repair/import_boundary_test.go` or similar) and assert the forbidden suffix is
`"/compiler/"` (anything under `internal/compiler/`) rather than a single named package — the
broadest form of the existing check, since the constraint here is "never import `internal/`" not
"never import one specific sibling package."

---

### `scripts/verify-phase6.sh` (test/config, batch)

**Analog, full skeleton to copy** — `scripts/verify-phase5.sh` (116 lines, read in full):
1. `mktemp -d`, `trap ... EXIT HUP INT TERM`, `export GOCACHE="$verify_tmp/go-cache"` (lines 1-6)
2. Self-test sentinel: `sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session <TestNames...>` (line 8) — list every new Phase 6 test name here
3. `go test ./...`, `go test -race ./...`, `go vet ./...`, `go build -o "$verify_tmp/lang" ./cmd/lang` (lines 9-12)
4. Per-prior-phase JSON capture: `"$verify_tmp/lang" --json verify testdata/phaseN >"$verify_tmp/phaseN.json"` for phase1..phase5 (lines 13-28), including the pinned `ASAN_OPTIONS`/`UBSAN_OPTIONS` env vars for the phase5 lane specifically (lines 18-28) — add a `testdata/phase6` capture in the same shape
5. Required-control grep assertions, one `for control in ...; do grep -q "$control" ...; done` block per prior phase (lines 36-93) — add Phase 6's own `for control in ...` block, sourced from a new `Phase6RequiredControls()` Go function mirroring `Phase5RequiredControls()`
6. Expected-escape greps (lines 95-99) — reuse this exact `grep -q 'escape:...'` shape for any new declared escape this phase records (e.g. the cache soundness holes from D-06-13, if surfaced as `expected_escapes`)
7. `cat` every phase's JSON at the end (lines 112-116)

**20-sample observation loop to port** — `scripts/verify-phase2.sh`'s `observe()` (lines 17-40,
full text already in RESEARCH, reproduced here for completeness):
```sh
observe() {
	name=$1; shift
	samples="$verify_tmp/$name.samples"; : >"$samples"
	last=; index=0
	while [ "$index" -lt 20 ]; do
		last=$(LANG_OBSERVE_TIMING=1 "$verify_tmp/lang" --json "$@")
		elapsed=$(printf '%s\n' "$last" | sed -n 's/.*"elapsed_ns":\([0-9][0-9]*\).*/\1/p')
		[ -n "$elapsed" ] && [ "$elapsed" -gt 0 ] || { echo "... produced no timing" >&2; exit 1; }
		printf '%s\n' "$elapsed" >>"$samples"
		index=$((index + 1))
	done
	sort -n "$samples" >"$samples.sorted"
	p50=$(sed -n '10p' "$samples.sorted")
	p95=$(sed -n '19p' "$samples.sorted")
	...
}
```
This is the exact 20-sample, `LANG_OBSERVE_TIMING=1`-gated shape D-06-19 must match — per the
Open Questions in RESEARCH, recommend porting the samples→{p50,p95,CoV,demotion} computation into
a small testable Go function, with `verify-phase6.sh` still driving the binary 20 times exactly
like this shell loop but piping raw samples through the new Go helper instead of `sort`/`sed`
percentile indexing.

**`LANG_OBSERVE_TIMING` gating precedent** (`session.go:1225-1234`, `completeCommand`):
```go
func completeCommand(result protocol.Result, started time.Time, work int) protocol.Result {
    if os.Getenv("LANG_OBSERVE_TIMING") == "1" {
        result.Metrics.ElapsedNS = time.Since(started).Nanoseconds()
    }
    result.Metrics.RecomputedWork = work
    return result.Finalize()
}
```
Any new per-stage timing (D-06-21's parse/check/lower/native-compile/link timestamps) must route
through this same gate, or a documented equivalent — never an unconditional `time.Since()`, or a
golden/pinned-JSON test will intermittently fail exactly as RESEARCH's Pitfall 2 warns.

## Shared Patterns

### Sibling-file convention (applies to every new session.go-adjacent surface)
**Source:** Phase 5 kept every new surface in `session_phase5*.go` siblings, leaving `session.go`
itself untouched except for the 12-site Lane bump. Phase 6 should do the same:
`session_phase6_explain.go`, `session_phase6_query.go`, `session_phase6_cache.go`,
`session_phase6_budget.go`, `session_phase6_risklanes.go` (or fold risk-lanes into the cache
sibling), one file per new surface. `cmd/lang/main.go`'s `isPhase5Corpus` (`main.go:152-164`) is
the CLI-layer half of this convention — a new `isPhase6Corpus` should live there too, not inside
`session.VerifyCorpus`.
**Apply to:** every new `.go` file this phase adds under `internal/compiler/session/`.

### `{Code}`-only typed error (applies to every new fail-closed error type)
**Source:** `debugmap.Error` (`debugmap.go:90-103`), `evidence.ValidationError`
(`evidence.go:87-101`), `originvalidate.Error` (same shape).
```go
type Error struct{ Code string }
func (e *Error) Error() string { return e.Code }
```
**Apply to:** the new `cache` package's errors, the risk-lanes/budget-manifest audit failures,
and any new query/explain-layer errors.

### Bounded subprocess (applies to any new probe)
**Source:** `evidence.runToolProbe`/`boundedProbeWriter` (`evidence.go:129-167`),
`MaxToolProbeBytes = 64*1024`, `5*time.Second` timeout.
**Apply to:** the Clang-digest probe (D-06-07), the `machine_id` probe (D-06-17). Do not write a
second bounding mechanism for either.

### Honest unavailability (applies to every new metrics/availability field)
**Source:** `debugmap.NotCaptured`/`OptimizedOut` (`debugmap.go:58-77`), the 15 hardcoded
`PeakRSSStatus: "unavailable"` producer sites across `session.go`.
**Apply to:** `cache_status` (never `hit`/`miss`, per D-06-12 explicitly), `peak_rss_status`
(stays `"unavailable"`, D-06-20), `machine_id`/`ratified` when undeclared (D-06-18).

### Checked-in-JSON-registry + live-cross-check audit (applies twice: risk lanes, budget manifest)
**Source:** the full `qlt01.go` pattern extracted in full above. Reuse the exact shape:
`//go:embed`, typed row slice, `AuditXxxRegistry(rows, liveSet)` pure function, `VerifyXxxRegistry`
lane wrapper. Do not build a second registry-loading mechanism.

### Fail-closed marker-count guard (applies to every new mutation runner and injector)
**Source:** `ReleaseOmissionMutationRunner`'s `matched == -1` refusal (`session.go:165-167`);
documented explicitly in `AliasFactMutationRunner`'s own comment
(`session_phase5_alias.go:44-73`). **Apply to:** all five D-06-25 injectors' own marker-kill test
(D-06-27.3).

## No Analog Found

Files/areas with no close existing match — planner should treat these as genuinely new design
surface, using RESEARCH.md's Architecture Patterns section rather than an in-tree copy target:

| File | Role | Data Flow | Reason |
|------|------|-----------|--------|
| `internal/compiler/cache/cache.go` + `cache_test.go` | service (new leaf package) | file-I/O, CRUD | No content-addressed cache exists anywhere in the tree today (confirmed by RESEARCH). Closest *structural* precedent is `internal/compiler/reduce` as a dependency-free leaf package with its own test file (imports only `core`), but it shares no functional shape with a cache — it is cited only for "how a brand-new package is scaffolded and tested in this repo," not for cache logic itself. |
| new Go statistics helper (p50/p95/CoV/demotion) | utility | batch/transform | Zero Go code anywhere computes percentiles or coefficient of variation today (confirmed by grep). `scripts/verify-phase2.sh`'s `observe()` (full text above) is the *behavioral spec* the new Go code must reproduce (20 samples, index-based p50/p95), not an importable library — it is shell, and the new code is Go. |
| `cmd/lang-repair/main.go` (the subprocess-spawning core logic itself, not its import-boundary test) | controller (standalone binary) | request-response, subprocess-only | No existing binary in this tree spawns another binary and parses only its `--json` stdout. `cmd/lang/main.go`'s own dispatch style (flat matching, `problemResult`, `emit`) is a reasonable stylistic template for the driver's *own* CLI surface, but the "spawn `lang`, read stdout, apply repair, respawn `lang`" control flow itself has no precedent to copy. |
| match/move/borrow defect fixtures (the `.lang` source files and their mutation logic, as opposed to the corpus-dispatch mechanism) | test fixture | file-I/O | No existing mutation runner targets match-exhaustiveness, use-after-move, or borrow-conflict defects at source granularity — the existing NAT-03 runners (`OwnedBackendMutationRunner`, `LayoutMutationRunner`, etc.) all mutate generated C, not `.lang` source. Only the corpus-layout and marker-mutation-kill *shapes* transfer (see Pattern Assignments above); the fixtures and their injection logic are new content. |
| Prose-scramble / vocabulary-removal test harness (D-06-27.1/.2) | test | transform (JSON mutation + subprocess substitution) | No existing test-hook mechanism intercepts or rewrites JSON between a spawned subprocess and a Go test anywhere in the tree. RESEARCH's Open Question #2 recommends a fixture-substitution subprocess (e.g. a `cat fixture.json`-shaped stand-in for `lang`) as the simplest new mechanism, since it requires zero changes to the real driver's code path — but this stand-in itself has no in-tree precedent. |

## Metadata

**Analog search scope:** `internal/compiler/protocol`, `internal/compiler/diagnostic`,
`internal/compiler/debugmap`, `internal/compiler/session` (incl. all `session_phase5*.go`
siblings and `qlt01.go`), `internal/compiler/evidence`, `internal/compiler/reduce`,
`internal/compiler/corevalidate`, `internal/compiler/pathoracle`, `internal/compiler/originvalidate`,
`cmd/lang`, `scripts/`, `testdata/`.
**Files scanned:** ~25 read in full or by targeted section this session; grep sweeps confirmed the
12-site Lane-schema literal count and the absence of any Go percentile/CoV code.
**Pattern extraction date:** 2026-09-06
