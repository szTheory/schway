# Phase 6: Agent Feedback and Performance Ratification - Research

**Researched:** 2026-09-06
**Domain:** Compiler-internal Go tooling — protocol/CLI extension, local build-artifact
caching, cost-budget ratification, agent-facing repair-loop verification
**Confidence:** HIGH (every claim below is grounded in a file/line read this session; no
external packages are involved, so there is no registry-lookup risk in this phase)

## Summary

Phase 6 is pure in-repo extension work: two new CLI commands (`explain`, `query`), a new
local-only artifact cache wired into the existing `verify` lane-running loop, a new checked-in
budget manifest with an executable audit, a repair-driver binary that talks to the shipped
`lang` binary over `--json` only, and one coordinated additive schema bump. There are **no new
external dependencies** — Go 1.24 stdlib only (`go.mod:3` `go 1.24`, no other `require` lines)
— so the Package Legitimacy Audit below is empty by construction.

The single most load-bearing fact this research turned up is that `protocol.Result.Finalize()`
(`internal/compiler/protocol/protocol.go:113-166`) builds its identity struct from an explicit,
closed field list that **already excludes `Metrics` entirely** and reduces every `Lane` to
`lane.ID+":"+lane.Status` (line 160). D-06-32's constraint ("no new metrics field may enter
`Result.Finalize()`'s identity struct") is therefore **already structurally satisfied** by every
field D-06-12/D-06-17/D-06-21 propose to add — `cache_status`, `selection_reason`, `machine_id`,
`stage_breakdown`, etc. all live on `Metrics`/`Lane`, and neither struct's non-ID/Status fields
reach `identity`. This is a verification target (write the test), not new plumbing.

The second load-bearing fact: `protocol.Lane{Schema: "lang.verify-lane/0", ...}` is a **hardcoded
string literal repeated at 12 call sites** across `session.go` and the `session_phase5*.go`
siblings (confirmed by grep, not assumed), not a single constant reference. The `/1` bump touches
all 12 sites plus every new lane this phase adds. `protocol.Result`'s schema, by contrast, is a
single `const Schema = "lang.command/0"` referenced through `protocol.New()|Finalize()` — bumping
it is a one-line change. Plan the two bumps as different-sized tasks.

Third: the "Phase 2 bounded gate's 20-sample warm distribution machinery" CONTEXT.md says D-06-19
reuses is **shell code, not Go code** — `scripts/verify-phase2.sh`'s `observe()` function (POSIX
`sh`, `sort`/`sed` over a 20-line samples file), gated on the `LANG_OBSERVE_TIMING=1` env var that
`session.go:1229` only honors there. There is no Go package computing p50/p95/CoV anywhere in the
tree today. D-06-19's CoV-based auto-demotion rule needs new Go code (testable) fed by these
samples, not a literal reuse of the shell loop — the shell loop is precedent for *shape*
(20 samples, first/last-half percentile-by-index), not an importable library.

Fourth: `debugmap.Resolve` (`internal/compiler/debugmap/debugmap.go:292-299`) is a **single-field
lookup by `OperationID` only** — it is not yet the five-vocabulary join D-06-01 describes. It is
real precedent for the *shape* (bounded map, honest `NotCaptured` absence, per-cold-invocation
build) but the join itself — `diagnostic:<hex>` / `core_id` / `operation_id` / `point_id` /
evidence ID or digest / `control:*` / `lane:*` — is new code this phase must write.

**Primary recommendation:** Treat this phase as four largely-independent workstreams gated by one
shared prerequisite (pin `protocol.Result`/`Lane`/`Metrics`/`diagnostic.Diagnostic` current byte
shapes with a `TestPreviousPhase*`-style freeze test, mirroring 05-01's own opening move) and one
shared final step (the coordinated `/1` bump, landed in its own commit touching all 12+ call
sites at once, verified by a script-vs-Go-constant equality test in the
`TestPhase5RequiredControlsMatchScript` mold).

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| `explain`/`query` addressing and cause-graph synthesis | Compiler frontend (`internal/compiler/session`, new `internal/compiler/query` or similar) | CLI (`cmd/lang/main.go`) | Pure in-process synthesis over already-produced core/diagnostic artifacts; CLI is a thin dispatch arm exactly like `debug-map` today |
| Change-risk lane selection + artifact cache | Compiler build/verify orchestration (`internal/compiler/session`, new `internal/compiler/cache`) | Filesystem (`$XDG_CACHE_HOME`) | The cache is a local content-addressed store the session package consults before invoking `native.Runner`; no network/service tier exists or is wanted (D-06-09) |
| Budget manifest + declared-machine ratification | Compiler verify orchestration (`internal/compiler/session`) | Shell (`scripts/verify-phase6.sh`) | Mirrors `qlt01_registry.json`/`qlt01.go` exactly: checked-in JSON registry + Go executable audit + shell-layer control-string assertions |
| Repair driver | Standalone `cmd/` binary, subprocess-only | none — explicitly forbidden from importing `internal/compiler/*` | D-06-28 makes this a structural boundary, not a convention; it is architecturally its own tier, siloed from the compiler internals by CI import-boundary lint |
| Schema versioning (`/1` bumps) | Compiler protocol/diagnostic packages (`internal/compiler/protocol`, `internal/compiler/diagnostic`) | Every session.go call site | Schema constants are the source of truth; but `Lane`'s schema is NOT centralized (12 literal call sites), so the "secondary tier" here is real migration work, not free |

## Standard Stack

### Core

No new external libraries. Go 1.24 stdlib only, matching every prior phase
(`go.mod:1-3`: `module github.com/codename-lang/lang` / `go 1.24`, no `require` block).

### Package Legitimacy Audit

**Not applicable — this phase installs zero external packages.** `go.mod` has no `require`
directives today and CONTEXT.md's decisions (D-06-06..D-06-32) name no third-party library at
any point; the cache, budget manifest, and repair driver are all hand-rolled per explicit
decision (`Don't Hand-Roll` section below explains why that is still correct policy). Skip the
Package Legitimacy Gate protocol for this phase; if a future task discovers a need for a stdlib-
adjacent helper, re-run the gate before adding it to `go.mod`.

## Architecture Patterns

### System Architecture Diagram

```
                     ┌─────────────────────────────────────────┐
                     │        cmd/lang/main.go (dispatch)       │
                     │  flat if/else on os.Args, extractJSON()  │
                     └───────┬───────────────┬─────────────────┘
                             │               │
                 existing arms│               │NEW: explain/query arms
                             │               │  (mirrors runDebugMap shape)
                             ▼               ▼
              ┌──────────────────────────────────────────┐
              │     internal/compiler/session (package)   │
              │  parse → check → corevalidate → (existing)│
              │  NEW: session.ExplainCommandFile(id)       │
              │  NEW: session.QueryCommandFile(id/pattern) │
              └───────┬─────────────────────┬──────────────┘
                      │                     │
       reads          │                     │ reads
       diagnostic.Cause[]                   debugmap.Map (extended:
       + span/binding correlation           5-vocabulary Resolve(),
       (per-cold-invocation synth,          not just OperationID)
        D-06-02)
                      │
                      ▼
              protocol.Result{Metrics, Lanes, ...} -- Finalize() -- JSON/Human


  ── verify path (separate flow) ──

  cmd `lang verify CORPUS`
        │
        ▼
  session.VerifyCorpus dispatch-by-marker-fixture
  (existing: foreign_acquire_one.lang / borrowed_view.lang / owned_transfer.lang / default)
        │
        ▼
  NEW: risk_lanes.json lookup (declared static table, D-06-10)
  → per (fixture × lane): consult NEW internal/compiler/cache package
       cache HIT (same declared inputs) → skip recompilation, reuse artifact on disk
       cache MISS/undeclared → recompile, always re-run checker/comparator/sanitizer FRESH
        │
        ▼
  protocol.Lane{cache_status, selection_reason, ...} appended (D-06-12)
        │
        ▼
  NEW: session/qlt02_budget_manifest.json cross-check against probed machine_id
  → ratified:true/false, gate_verdict on recomputed_work only (D-06-14/D-06-22)


  ── repair-driver path (fully separate binary) ──

  cmd/<repair-driver> (new binary)
        │  subprocess only, no internal/ import (D-06-28, CI-lint enforced)
        ▼
  invoke shipped `lang --json check|evidence ...` as a subprocess
        │  parses ONLY --json stdout
        ▼
  applies repair using span/replacement/applicability (D-06-24 additive Repair fields)
        │
        ▼
  reverify via another `lang --json check` subprocess call
```

### Recommended Project Structure
```
internal/compiler/
├── cache/                    # NEW (D-06-06..D-06-09): content-addressed artifact store
│   ├── cache.go              #   key computation (declared-input hash), get/put, meta.json
│   └── cache_test.go
├── session/
│   ├── session.go            # UNCHANGED except Lane literal /1 bump (12 sites)
│   ├── session_phase6_explain.go   # NEW, sibling convention (Phase 5 precedent)
│   ├── session_phase6_query.go     # NEW
│   ├── session_phase6_cache.go     # NEW: wires cache pkg into VerifyCorpus-adjacent path
│   ├── session_phase6_budget.go    # NEW: machine_id probe + budget manifest cross-check
│   ├── qlt02_budget_manifest.json  # NEW, checked-in, mirrors qlt01_registry.json
│   └── risk_lanes.json             # NEW, checked-in, mirrors qlt01_registry.json
├── diagnostic/
│   └── diagnostic.go         # Repair struct grows Span/Replacement/Applicability (additive)
├── protocol/
│   └── protocol.go           # Metrics/Lane grow fields; Schema constants /1 bump
cmd/
├── lang/main.go              # +explain, +query dispatch arms
└── lang-repair/              # NEW standalone binary (D-06-28), subprocess-only
    └── main.go
scripts/
└── verify-phase6.sh          # NEW, peer of verify-phase5.sh
```

### Pattern 1: Sibling-file + literal-registry-with-executable-audit
**What:** New JSON registry embedded via `//go:embed`, parsed into a typed row slice, audited by
an exported `AuditXxxRegistry([]Row, liveSet)` pure function, wired into a `VerifyXxxRegistry`
lane wrapper — exactly `qlt01.go`'s shape.
**When to use:** For both `risk_lanes.json` (D-06-10) and `qlt02_budget_manifest.json` (D-06-16).
**Example (verified shape, `internal/compiler/session/qlt01.go:12-49,246-253`):**
```go
//go:embed qlt01_registry.json
var qlt01RegistryBytes []byte

func LoadQLT01Registry() ([]QLT01Row, error) {
    var rows []QLT01Row
    if err := json.Unmarshal(qlt01RegistryBytes, &rows); err != nil {
        return nil, fmt.Errorf("qlt01: failed to parse embedded registry: %w", err)
    }
    return rows, nil
}
func VerifyQLT01Registry(ctx context.Context) (LaneResult, error) {
    rows, err := LoadQLT01Registry()
    if err != nil { return LaneResult{}, err }
    return QLT01LaneFromRows(rows, AllShippedControlIDs()), nil
}
```
The audit fires two control IDs (`ControlQLT01RegistryIncomplete`,
`ControlQLT01StaleControlReference`) cross-checked against a **live** set
(`AllShippedControlIDs()`, itself built from `Phase4RequiredControls()`/`Phase5RequiredControls()`
— never a hand-copied literal). `risk_lanes.json`'s audit should cross-check declared lane IDs
against the actual lane IDs `VerifyCorpus`/`verifyOwnedCorpus`/etc. produce; `qlt02_budget_manifest.json`'s
audit should cross-check declared `machine_id`s against a **live probe** output (D-06-16 says this
explicitly).

### Pattern 2: Marker-file corpus dispatch at the CLI layer
**What:** `cmd/lang/main.go`'s `isPhase5Corpus` (lines 152-164) recognizes a corpus by the
presence of a characteristic fixture file and dispatches to a phase-specific verify function,
keeping `session.go` itself untouched.
**When to use:** If `explain`/`query` or the repair exercise need their own held-out corpus
dispatch, follow this exact precedent rather than editing `VerifyCorpus`'s existing marker chain
(`foreign_acquire_one.lang` / `borrowed_view.lang` / `owned_transfer.lang`, `session.go:1241-1251`).

### Pattern 3: Per-cold-invocation synthesis, never a persisted graph
**What:** `debugmap.Build` (`internal/compiler/debugmap/debugmap.go:161-260`) reruns the full
join (astProgram × checkedProgram) on every call; nothing is cached or persisted across
invocations. `DebugMapCommandFile` (`session.go:1143-1195`) reruns parse→check→corevalidate from
scratch too.
**When to use:** `explain`'s cause-DAG synthesis (D-06-02) should follow this identically — no
graph store, computed fresh, and (per D-06-02's own explicit callout) needs a determinism test:
run twice on the same cold input, assert byte-identical output, the same shape as
`VerifyCorpus`'s own `lane:deterministic` (`session.go:1279-1288`, which already does exactly
this two-call-compare pattern for `RunInterpreter`).

### Anti-Patterns to Avoid
- **Fabricating peak-RSS or `cache_status: hit/miss`:** the project's own D-06-20 and D-06-12
  each forbid this by name; the existing 15 `PeakRSSStatus: "unavailable"` producer sites are the
  house style to imitate, not a gap to close.
- **A single shared `addLane` helper spanning multiple verify functions:** the existing code
  deliberately reimplements a local `addLane` closure inside each of `VerifyCorpus`,
  `verifyOwnedCorpus`, `verifyBorrowedCorpus`, `verifyForeignCorpus`, and
  `VerifyPhase5ControlsAndWork` independently (5 separate closures, confirmed by grep and reads)
  — this is consistent with the sibling-file/no-shared-mutation-point convention. Do not
  refactor them into one shared function as part of this phase; instead extend each local
  closure's `protocol.Lane{...}` literal with the new fields.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Content-addressed cache key | A bespoke hashing scheme from scratch with no declared-input transparency | Go's own `ActionID` model (declared-input list → single SHA-256), already the cited precedent (D-06-07) | `meta.json` recording the full declared-input list is the auditability mechanism; skipping it (e.g. hashing only file mtimes) reproduces ccache's own footgun class |
| Wall-clock statistics | A custom bootstrap/CoV statistical framework | The already-shipped 20-sample p50/p95 shape from `scripts/verify-phase2.sh`'s `observe()`, ported/extended in Go for CoV computation | The shell precedent already proves the sampling protocol works at this scale (single host, no CI fleet); building LNT/Talos-style tooling is explicitly rejected by D-06-15 |
| Host identity probing | A new subprocess/timeout wrapper | `evidence.runToolProbe`/`defaultFacts` (`internal/compiler/evidence/evidence.go:103-167`) — already bounded at `MaxToolProbeBytes = 64*1024` with a `5*time.Second` timeout | D-06-07/D-06-17 explicitly want this reused; it already handles the bounded-stdout/stderr and timeout mechanics correctly |

**Key insight:** every "don't hand-roll" item in this phase is actually "don't hand-roll a
*new mechanism* when an in-tree mechanism already solves the hard part (bounding, timeout,
declared-input transparency)" — there are no external-library alternatives to reject here, only
internal-reuse-vs-reinvent choices.

## Runtime State Inventory

Not applicable — this is a greenfield-within-phase-6 addition (new commands, new package, new
JSON registries), not a rename/refactor/migration. No existing stored data, live service config,
OS-registered state, secret/env var names, or build artifacts reference anything this phase
renames. Skip.

## Common Pitfalls

### Pitfall 1: Assuming `protocol.Lane`'s schema is a single edit point
**What goes wrong:** A plan that says "bump `protocol.Lane.Schema` to `/1`" as a one-line task
under-scopes the work by roughly 12x.
**Why it happens:** `protocol.Result`'s schema genuinely IS a single constant
(`protocol.Schema`), creating a false analogy.
**How to avoid:** grep-verified list of hardcoded `"lang.verify-lane/0"` literals: 12 occurrences
across `internal/compiler/session/session.go`, `session_phase5.go`, `session_phase5_mismatch.go`,
`session_phase5_sanitize.go`. Either (a) introduce a shared `const LaneSchema1 = "lang.verify-lane/1"`
and edit all 12 sites plus every new lane this phase adds in one commit, or (b) accept this as the
phase's largest single mechanical task and budget a dedicated plan/task for it.
**Warning signs:** a task estimate that treats the Lane bump as equal-effort to the Result bump.

### Pitfall 2: Treating `elapsed_ns` as always-populated
**What goes wrong:** New stage-timestamp code (D-06-21) or budget-observation code that reads
`Metrics.ElapsedNS` in a normal (non-`LANG_OBSERVE_TIMING=1`) invocation will silently see `0`.
**Why it happens:** `completeCommand` (`session.go:1225-1234`) **only** sets `ElapsedNS` when
`os.Getenv("LANG_OBSERVE_TIMING") == "1"` — deliberately, to keep ordinary command output
byte-reproducible. This is the existing mechanism protecting exactly the determinism property
D-06-32 cares about, and it is a real precedent for how `stage_breakdown` timestamps must be
gated the same way (not always-on).
**How to avoid:** any new per-stage timing (parse/check/lower/native-compile/link, D-06-21) must
route through the same `LANG_OBSERVE_TIMING`-gated path, or a documented equivalent, not an
unconditional `time.Since()`.
**Warning signs:** a golden/pinned JSON test that starts failing intermittently because a new
timing field is nonzero only sometimes.

### Pitfall 3: `debugmap.Resolve` is not yet the five-vocabulary join
**What goes wrong:** Assuming D-06-01's `query <id>` can call straight into
`debugmap.Resolve(built, id)` and "just work" for `diagnostic:<hex>` or `control:*` IDs.
**Why it happens:** `Resolve` (`debugmap.go:292-299`) only ever compares against
`entry.OperationID` — a single field, single vocabulary.
**How to avoid:** plan `query`'s join layer as new code: a dispatcher that recognizes an ID's
vocabulary by its prefix/shape (`diagnostic:`, `control:`, `lane:`, or debugmap's own
`{functionID}:op:{n}` / `:point:{n}` shapes, or an evidence ID/digest) and routes to the
appropriate existing lookup (`debugmap.Resolve`, a new diagnostic-ID index, a new evidence-ID
index), converging on one `NotCaptured`-shaped honest-absence answer when none match. This is a
genuinely new join layer, not a wrapper.
**Warning signs:** a task description that says "wire query to debugmap.Resolve" with no
mention of the other four vocabularies.

### Pitfall 4: The evidence-manifest identity struct already has a `ForeignDigest`-style precedent for additive fields — but the schema constant is separate from the Manifest struct
**What goes wrong:** Conflating `evidence.Schema0`/`Schema1` (`internal/compiler/evidence/evidence.go:28-31`)
with `protocol.Schema`/`diagnostic.Schema`/`diagnostic.Schema1` — they are four *independent*
version lineages (evidence, command/lane, diagnostic), each with its own frozen-`/0` obligation.
D-06-31's coordinated bump touches only `lang.command/0`→`/1` and `lang.verify-lane/0`→`/1`; it
explicitly does NOT touch `lang.diagnostic/0` or `lang.evidence/0` (those stay exactly as they
are; `lang.diagnostic/1` already exists as of Phase 2 and D-06-24 extends its `Repair` struct
additively without a further version bump — confirmed: `ErrorWithRepairs`'s identity struct
(`diagnostic.go:70-77`) derives `RepairKinds` from `repairs[i].Kind` only, so adding
`Span`/`Replacement`/`Applicability` fields to `Repair` cannot change any existing diagnostic's
`ID`).
**How to avoid:** keep the four schema lineages in four separate plan tasks; do not let one
"schema bump" task accidentally touch more than its named pair.

## Code Examples

### Existing tool-identity probe (reuse target for D-06-07/D-06-17)
```go
// Source: internal/compiler/evidence/evidence.go:146-167 (verified this session)
func runToolProbe(parent context.Context, command commandFactory, name string, arguments ...string) ([]byte, error) {
    ctx, cancel := context.WithTimeout(parent, 5*time.Second)
    defer cancel()
    cmd := command(ctx, name, arguments...)
    var stdout, stderr boundedProbeWriter // caps at MaxToolProbeBytes = 64*1024
    ...
}
```
`defaultFacts` (lines 107-127) already captures `clang --version` and `-dumpmachine` output as
strings — this is a **version-string** probe, not a **content-hash digest** of the Clang binary.
D-06-07 explicitly wants "a probed digest, not merely the `--version` string" — this is new work
layered on top of the existing bounded-subprocess mechanism, not something already done. Flag this
precisely: CONTEXT.md's phrase "the shipped tool-identity probe already does this kind of work"
is accurate for the *bounding/timeout* mechanism, not for the *digest* semantics D-06-07 needs.

### Existing 20-sample warm-distribution shape (precedent for D-06-19, shell not Go)
```sh
# Source: scripts/verify-phase2.sh (verified this session, lines ~14-38)
observe() {
	name=$1; shift
	samples="$verify_tmp/$name.samples"; : >"$samples"
	index=0
	while [ "$index" -lt 20 ]; do
		last=$(LANG_OBSERVE_TIMING=1 "$verify_tmp/lang" --json "$@")
		elapsed=$(printf '%s\n' "$last" | sed -n 's/.*"elapsed_ns":\([0-9][0-9]*\).*/\1/p')
		printf '%s\n' "$elapsed" >>"$samples"
		index=$((index + 1))
	done
	sort -n "$samples" >"$samples.sorted"
	p50=$(sed -n '10p' "$samples.sorted")
	p95=$(sed -n '19p' "$samples.sorted")
	...
}
```
D-06-19's CoV auto-demotion rule needs the samples array available to *Go* code (to be unit
tested per this project's own standing "mutation-kill every differential" rule) — recommend
either (a) a small Go helper invoked by the shell script that also emits the CoV and demotion
decision, or (b) moving the whole `observe`/ratify loop into a new Go-driven `verify --observe`
path that the shell script calls once instead of 20 times. Either is a genuinely new artifact;
neither is "just reuse the shell function."

### `Result.Finalize()`'s closed identity struct (verification target for D-06-32)
```go
// Source: internal/compiler/protocol/protocol.go:113-133 (verified this session)
func (result Result) Finalize() Result {
	identity := struct {
		Schema, Command, Status, ModuleID, FormattedDigest string
		DiagnosticIDs, ExecutionDigests                    []string
		EvidenceID, InterfaceID, DebugMapID                string
		LaneIDs                                            []string
		ExpectedEscapes                                     []string `json:",omitempty"`
	}{ /* ... */ }
	// Metrics is NEVER referenced here. Lanes contribute ONLY lane.ID+":"+lane.Status (line 160).
	...
}
```
This is the exact evidence that D-06-32 is already satisfied structurally by any field added to
`Metrics` or to `Lane` (beyond `ID`/`Status`) — write a test asserting this stays true (e.g. two
`Result`s differing only in a new `Metrics`/`Lane` field must produce identical `.ID`), rather
than adding new defensive plumbing.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| No cache of any kind — every `verify` invocation recompiles everything | Local content-addressed artifact cache (D-06-06..D-06-09), never caching verdicts | This phase | `verify` cost drops for unchanged declared-input sets; checker/comparator/sanitizer logic is unaffected — this is the phase's core soundness property |
| Wall-clock timing gated behind `LANG_OBSERVE_TIMING=1`, reported informally by shell scripts only | Ratified cold/warm distributions in a checked-in budget manifest, still observation-based for wall-clock, hard-gated only on `recomputed_work` | This phase | First ratified budgets in the project; establishes the machine-declaration discipline for any future second machine |
| `diagnostic.Repair{Kind, Detail}` — classification only | `Repair` gains optional `Span`/`Replacement`/`Applicability` — an applicable edit | This phase (additive on already-shipped `/1`) | Enables the repair driver; `kind` remains the only identity-bearing field |

**Deprecated/outdated:** nothing in this phase deprecates prior-phase surface; every change here
is additive per the project's standing "additive `/1` bump, `/0` bytes frozen" rule.

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Node-count budget and default `--depth=3` for `explain`, exact CoV threshold for auto-demotion, exact truncation code strings are all "Claude's Discretion" per CONTEXT.md — this research does not pick numbers, leaving them to the planner/executor to select by closest in-tree precedent (64 KiB-plus-one output bound, `MaxEntries = 4096` in debugmap) | Standard Stack / Code Examples | Low — CONTEXT.md explicitly delegates this choice; wrong-but-reasonable defaults are cheap to adjust since they are not identity-bearing |
| A2 | Whether the `risk_lanes.json` audit should be a new executable audit or an extension of the QLT-01 audit harness is left to planner discretion per CONTEXT.md; this research recommends a *new* audit (Pattern 1) because `qlt01.go`'s existing audit is scoped specifically to spike-control-descendant rows, a different row shape than a risk→lane declaration table | Architecture Patterns | Low — CONTEXT.md names this exact question as discretionary |

**If this table is empty:** N/A — two low-risk discretionary items are logged above; every
factual/mechanical claim elsewhere in this document is `[VERIFIED: <path:line>]` against a file
read this session.

## Open Questions

1. **Where does the CoV/demotion computation for D-06-19 live — Go package or shell script?**
   - What we know: the sampling loop precedent is 100% shell (`scripts/verify-phase2.sh`); no Go
     package computes percentiles or CoV anywhere in the tree today.
   - What's unclear: whether the planner wants a testable Go helper (this project's own standing
     "mutation-kill every differential" rule favors Go, since shell arithmetic is hard to unit
     test) or a shell-only implementation matching Phase 2 exactly.
   - Recommendation: put the sample-array → {p50, p95, CoV, demotion-verdict} computation in a
     small new Go function (table-testable with synthetic sample slices), and have
     `scripts/verify-phase6.sh` invoke the *binary* 20 times exactly like Phase 2's `observe()`
     but pipe the raw samples through the new Go computation instead of `sort`/`sed` percentile
     indexing. This keeps the shell script thin and the statistical logic testable.

2. **Exact shape of the repair driver's JSON-stream substitution hook for the two anti-theater
   guards (D-06-27.1/.2).**
   - What we know: the driver reads only `--json` stdout from a `lang` subprocess
     (`cmd/lang/main.go`'s `emit()` already writes canonical JSON to stdout on `--json`, confirmed
     at `main.go:232-255`). No existing test-hook mechanism intercepts/rewrites JSON between a
     spawned `lang` process and a Go test.
   - What's unclear: whether the prose-scramble/vocabulary-removal tests should (a) capture real
     `lang --json` output once, mutate the parsed JSON structure in Go, and feed the mutated bytes
     to the driver via a stdin/file substitution point the driver must therefore support, or
     (b) spawn a **fake** `lang`-shaped subprocess (a small test-only binary or script) that emits
     pre-mutated JSON, keeping the real driver's "spawn `lang` and read its stdout" code path
     completely unmodified.
   - Recommendation: prefer (b) — a fixture-substitution subprocess (e.g. `cat fixture.json`) is
     simpler, requires zero driver code changes to support a test hook (which would otherwise risk
     becoming a second, untested code path), and keeps the import-boundary lint (D-06-28) trivially
     satisfied since the driver never learns it's not talking to the real binary.

3. **Does `internal/compiler/reduce`'s `Reduce`/`SignatureFromDisagreement` machinery need new
   wiring for D-06-26's differential-behavior fallback, or can the repair oracle call it as-is?**
   - What we know: `reduce.Reduce(ctx, seed core.Program, interesting Predicate)` exists and is
     already wired from `session` (`session.SignatureFromDisagreement`, per STATE.md's Phase 5
     ledger). It operates on `core.Program`, not on repair-driver JSON.
   - What's unclear: the repair driver is subprocess-only and never imports `internal/compiler/*`
     (D-06-28) — so it cannot call `reduce.Reduce` directly. The differential-behavior fallback
     oracle (for defect classes where byte-identity is too strict) must live on the **CI-gate**
     side (in-process Go, which CAN import `reduce`), not inside the driver binary itself.
   - Recommendation: split D-06-26's oracle: byte-identity check happens driver-side (trivial,
     JSON/file comparison, no `internal/` import needed); the differential-behavior fallback
     check is a **separate, in-process Go test** that re-runs `reduce`/mismatch machinery over the
     driver's *produced files* after the driver has already exited — never inside the driver.

## Environment Availability

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go toolchain | All of Phase 6 | ✓ (existing project requirement) | 1.24 (`go.mod:3`) | — |
| Clang | Cache key input (D-06-07), native lanes | ✓ (existing project requirement, already probed by `evidence.DefaultFacts`) | probed at runtime via `clang --version` | — |
| `$XDG_CACHE_HOME` | Local artifact cache (D-06-09) | Environment-dependent; falls back to a conventional default (`~/.cache`) when unset — confirm at implementation time with `os.UserCacheDir()` | — | Use Go's `os.UserCacheDir()` (stdlib) rather than reading `$XDG_CACHE_HOME` directly, to get the existing cross-platform fallback for free |

**Missing dependencies with no fallback:** none identified.

**Missing dependencies with fallback:** `$XDG_CACHE_HOME` — use `os.UserCacheDir()`.

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go's built-in `testing` package (no third-party test framework anywhere in the tree) |
| Config file | none — `go test ./...` driven directly; `scripts/assert-go-tests.sh` wraps exact-test selection with a self-test-sentinel guard |
| Quick run command | `sh scripts/assert-go-tests.sh ./internal/compiler/session TestNameHere` |
| Full suite command | `go test ./... && go test -race ./... && go vet ./...` (exact sequence used by every `verify-phaseN.sh`) |

### Phase Requirements → Test Map
| Req ID | Behavior | Test Type | Automated Command | File Exists? |
|--------|----------|-----------|-------------------|-------------|
| FND-04 | Every command reports wall time/peak-mem-status/output bytes/cache status/affected work, no semantic-output change | unit + CLI | `go test ./internal/compiler/protocol/... ./internal/compiler/session/...` plus a new `TestMetricsFieldsExcludedFromIdentity` | ❌ Wave 0 — new test |
| DX-02 | `explain`/`query` bounded, stable-ID addressing, cause graph | unit + CLI golden | `go test ./internal/compiler/session/... -run TestExplain\|TestQuery` | ❌ Wave 0 — new package/tests |
| DX-03 | `verify` selects lanes by changed risk over local cache; `evidence` expands traces on demand | unit + CLI | `go test ./internal/compiler/cache/... ./internal/compiler/session/... -run TestCache\|TestRiskLane` | ❌ Wave 0 — new package/tests |
| DX-04 | Repair driver fixes match/move/borrow/cleanup/stale-evidence defects via protocol only | CLI/subprocess + anti-theater | `sh scripts/verify-phase6.sh` (new) driving `cmd/lang-repair` | ❌ Wave 0 — new binary + gate |
| QLT-02 | Ratified cold/warm distributions on declared machines; blocking rule on `recomputed_work` only | shell-driven sampling + unit | `sh scripts/verify-phase6.sh` (20-sample loop) + `go test ./internal/compiler/session/... -run TestBudget` | ❌ Wave 0 — new registry/tests |

### Sampling Rate
- **Per task commit:** `sh scripts/assert-go-tests.sh ./internal/compiler/<pkg> <ExactTestName>` for
  the specific test(s) the task adds/touches.
- **Per wave merge:** `go test ./... && go test -race ./... && go vet ./...` (exact sequence every
  `verify-phaseN.sh` already runs first).
- **Phase gate:** `sh scripts/verify-phase6.sh` full green, following the `verify-phase5.sh`
  shape (self-test sentinel → `go test ./...` → race → vet → build → per-phase-corpus JSON runs →
  required-control grep assertions → this phase's own 20-sample `observe()`-style loop).

### Wave 0 Gaps
- [ ] `internal/compiler/protocol/protocol_test.go` — add
  `TestMetricsAndLaneFieldsExcludedFromIdentity` (proves D-06-32 structurally, per the Code
  Examples section above) before any new `Metrics`/`Lane` field is added.
- [ ] `internal/compiler/cache/` — brand-new package, zero existing tests; needs its own
  `cache_test.go` from the first task that creates the package.
- [ ] `internal/compiler/session/session_phase6_*_test.go` — sibling test files per the existing
  Phase 5 convention, one per new surface (explain/query/cache-wiring/budget).
- [ ] `scripts/verify-phase6.sh` — does not exist yet; write following `verify-phase5.sh`'s exact
  shape (self-test sentinel test names, `GOCACHE` pinned to a disposable temp dir, per-corpus
  JSON captured then grepped for required `control:*` strings, expected-escapes grepped
  separately).
- [ ] A held-out `.lang` defect fixture corpus distinct from any fixtures used to hand-derive the
  driver's kind→edit mapping (D-06-29) — none exists yet; needs new `testdata/phase6/` (or
  similarly named) directory, structurally separate from whatever fixtures inform the driver's
  authored logic.

## Security Domain

`security_enforcement: true` in `.planning/config.json:48`.

### Applicable ASVS Categories

| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V2 Authentication | No | No auth surface in this phase — local CLI tooling only |
| V3 Session Management | No | N/A |
| V4 Access Control | No | N/A |
| V5 Input Validation | Yes | Every new input (cache `meta.json`, `risk_lanes.json`,
  `qlt02_budget_manifest.json`, repair-driver `--json` stream) must go through the project's
  existing bounded-read discipline: `readBoundedFile` (`session.go:1210-1217`, `io.LimitReader`
  at declared byte caps) and `encoding/json` strict decode (`evidence.DecodeStrict`-style,
  `evidence.go:317`) rather than ad hoc parsing. New JSON registries embedded via `//go:embed`
  are trusted (checked into the repo, reviewed like code); anything read from the untrusted cache
  directory or a repair-driver subprocess's stdout must be bounded and strictly decoded. |
| V6 Cryptography | Yes — content identity only | SHA-256 for cache keys and content digests,
  matching the project's own explicit stance ("SHA-256 is content identity, never proof",
  `debugmap.go` doc comment lines, `evidence.go:409-414` `digest`/`manifestID`) — never claim the
  cache key or digest as a security/integrity control beyond content-addressing. |

### Known Threat Patterns for this stack

| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| A hand-edited or partially-deleted cache directory silently treated as valid (D-06-13's own
  named hole #4) | Tampering | Documented as an accepted escape (per house style); no integrity
  check beyond content-hash lookup is claimed — this is explicit, recorded debt, not a gap to
  silently close |
| Command injection via a repair-driver `replacement` string applied to a `.lang` source file | Tampering | The driver reads `span`/`replacement` only from the binary's own `--json` output
  (never scrapes prose, D-06-28); the binary itself is the trust boundary producing that JSON —
  no new escape vector is introduced as long as the driver genuinely never opens a `.lang` file
  itself outside the span it was handed (the CI import-boundary lint is the enforcement
  mechanism, not a convention) |
| A subprocess-spawned `clang`/`lang` binary with an unbounded stdout/stderr hanging the cache-key
  computation | Denial of Service | Reuse `evidence.runToolProbe`'s existing bound
  (`MaxToolProbeBytes = 64*1024`, `5*time.Second` timeout) for the Clang-digest probe (D-06-07);
  do not write a second, unbounded probe path |

## Sources

### Primary (HIGH confidence — file/line read this session)
- `internal/compiler/protocol/protocol.go` — `Metrics`, `Lane`, `Result`, `Finalize()` identity
  struct (lines 24-30, 76-86, 88-166), `New()` (105-111), `JSON`/`Human`/`emit` output-bytes
  convergence loop (168-258).
- `internal/compiler/diagnostic/diagnostic.go` — `Cause`, `Repair`, `Diagnostic`, `Error()`,
  `ErrorWithRepairs()` (full file, 88 lines).
- `internal/compiler/debugmap/debugmap.go` — `Availability`, `Entry`, `Map`, `Build()`,
  `Resolve()` (full file, 300 lines).
- `internal/compiler/session/qlt01.go` — registry row shape, `AuditQLT01Registry`,
  `VerifyQLT01Registry`, `QLT01LaneFromRows` (full file, 295 lines) — the copy-target pattern for
  `risk_lanes.json`/`qlt02_budget_manifest.json`.
- `internal/compiler/session/session.go` — `VerifyCorpusFile`/`VerifyCorpus` dispatch
  (1241-1340), `BorrowedLoanEndpointControlLane`/`PathOracleDisagreementLane` (1866-1964),
  `completeCommand`/`LANG_OBSERVE_TIMING` gating (1210-1234), `DebugMapCommandFile` (1143-1195),
  `ValidateEvidenceCommandFile` (979-1002), `Phase4RequiredControls` (2004+), release-marker
  constant (128).
- `internal/compiler/session/session_phase5.go` — `Phase5RequiredControls`,
  `VerifyPhase5ControlsAndWork` lane-construction pattern (1-170+).
- `internal/compiler/evidence/evidence.go` — `Facts`, `defaultFacts`, `runToolProbe`,
  `boundedProbeWriter`, schema constants (1-170).
- `cmd/lang/main.go` — full dispatch shape, `extractJSON`, `emit`, `isPhase5Corpus` (full file,
  266 lines).
- `scripts/verify-phase5.sh`, `scripts/verify-phase2.sh`, `scripts/assert-go-tests.sh` — full
  read; the `observe()` 20-sample shell function, the self-test-sentinel `assert-go-tests.sh`
  pattern, the required-`control:*` grep-assertion pattern.
- `internal/compiler/core/core_test.go` — `TestPreviousPhaseCoreBytesUnchanged`,
  `TestPreviousPhaseManifestIDsUnchanged`, `TestPreviousPhaseGoldenCUnchanged` (lines 81-177) —
  the byte-freeze precedent for the coordinated schema bump.
- `go.mod` — confirms Go 1.24, stdlib-only, no `require` block.
- `.planning/config.json` — `security_enforcement: true` (line 48), `nyquist_validation: true`.
- `.planning/phases/06-agent-feedback-and-performance-ratification/06-CONTEXT.md` — all 32 locked
  decisions (verbatim, not re-researched).
- `.planning/REQUIREMENTS.md` — FND-04/DX-02/DX-03/DX-04/QLT-02 verbatim text.
- `.planning/STATE.md` — Phase 2 "20 warm distributions" ledger entry (confirms shell-not-Go
  origin), Phase 5 decision ledger (sibling-file convention, byte-freeze precedent).
- grep sweeps confirming: zero `"result:"`-ID pinned fixtures anywhere in the tree; 12 hardcoded
  `"lang.verify-lane/0"` literal call sites across `session.go`/`session_phase5*.go`; zero Go code
  computing p50/p95/CoV anywhere in the tree.

### Secondary (MEDIUM confidence)
None — every claim above was verified by direct file read this session; no WebSearch or Context7
lookups were needed since this phase touches no external library or public API surface.

### Tertiary (LOW confidence)
None.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — no external packages exist or are proposed; `go.mod` read directly.
- Architecture: HIGH — every referenced struct/function/file was read this session, not recalled
  from training data.
- Pitfalls: HIGH — each pitfall is backed by a direct grep/read (12-site Lane literal count,
  `LANG_OBSERVE_TIMING` gating, `debugmap.Resolve`'s single-field lookup, shell-only 20-sample
  loop).

**Research date:** 2026-09-06
**Valid until:** Stable — this is an in-repo, no-external-dependency phase; the only staleness
risk is the codebase itself changing before planning starts. Re-verify file/line citations if
more than a few commits land on `main` before `/gsd-plan-phase` runs.
