# Phase 28: Exit Classes, Verify Report, and Entry Pruning - Research

**Researched:** 2026-10-10
**Domain:** Go 1.24 stdlib-only compiler driver (session / cmd / native / cgen / callgraph); no new dependencies
**Confidence:** HIGH (all anchors read from source this session; two decision-consistency issues flagged in Open Questions)

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
D-01 .. D-17 in `28-CONTEXT.md` are LOCKED. Summary of the load-bearing ones (read the file for full text):
- D-01 entry rule: `main` if present; else unique in-degree-zero root, else strictly-largest closure; else refuse `core.entry_ambiguous` (message lists roots, says add `fn main`).
- D-02 `entry_ambiguous` is a source refusal (exit 2) from check, build, run (both engines). executionpeer.go:318 keeps its own code.
- D-03 keep the closure-size tiebreak for programs without `main` (override of research recommendation).
- D-04 check everything, then prune once: `resolveEntry(checked) (entry, pruned, err)` + `callgraph.PruneToReachable(entry)`; preserves declaration order; all `EntryFunction` call sites collapse onto the helper; inventory meta-test fails if `EntryFunction` is called outside the helper (and callgraph's own tests).
- D-05 corpus golden committed before any behavior change; fully reachable program => `PruneToReachable` deep-equal; `TestEntryFunctionEmitsUnreachableFunctions` deliberately rewritten with a dated note.
- D-06 helper shaped so Phase 30 can compose; nothing module-specific now.
- D-07 new leaf package `internal/compiler/exitcode` (imports nothing) with the named table (Success 0, Invalid 2, Operational 3, Mismatch 4, Usage 64, AppInputRefused 65, EventCapacity 69, AppTimeout 124, AppLaunchError 125, SignalBase 128, Unrepaired 1). 65 and 69 never renumbered. No 126/127.
- D-08 `app run` stays in its env/timeout domain; child status passes through; 65 for `native.input_too_long` / `input_contains_nul` stays.
- D-09 W2: `admitProgram(pruned)` = corevalidate, originvalidate, pathoracle in `CheckCommandFile` order, returns `[]diagnostic.Diagnostic` + nil error; keeps `core.peer_refusal_unnamed` fallback; test seams `checkCommandPeerSeam` / `observeCheckCommandPeer` stay.
- D-10 W3: `runNative` admits before any cgen/clang; interpreter admits up front; post-hoc `publishedOriginProblemFile` dropped; genuine cut-M004 lowering fixtures stay exit 3 `native.tool_failure`.
- D-11 static no-duplicated-literals AST scan (go/parser) over non-test files in session, native, protocol, cgen, cmd/schway, cmd/schway-repair/main.go (exclude repair.go and exitcode pkg); mandatory self-test with synthetic source per position; pin test for 65, 69 and table uniqueness; generated-C `return 65` / `exit(65)` / `exit(69)` built from or asserted against the constants.
- D-12 enumerate pins before editing; add pins for 124/125/128+n; add corevalidate-refusal fixture asserting same exit 2 + code on check/build/run on both engines; add two-domain exit-code table to user docs.
- D-13..D-17 verify report: `outcome_digest` for exhausted cases, `Engines` omitted for them; statuses unchanged; schema `lang.app-verification/2`; `MaxReplayCaseIDBytes = 128`; `TestAppVerifyReportBudget` (16 exhausted cases <= `MaxReplayCasesBytes/2`); named oversize error kept as defense in depth; required evidence n=14 pass, n=15/172/173/1000 unsupported, 16-case set.

### Claude's Discretion
- Exact helper names (`resolveEntry`, `admitProgram`, `PruneToReachable`), file split, plan/wave decomposition.
- Golden format (per-file hash manifest or stored C text) — existing testdata conventions below.
- Whether cgen formats exit literals from the constants or test-asserts them.

### Deferred Ideas (OUT OF SCOPE)
- Requiring an explicit `fn main` for all programs (language change).
- Evidence sidecar file for exhausted cases.
- Loop-aware event capacity (D-27.3-03; REQUIREMENTS Future).
- N4 and R1b.
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|------------------|
| DX-17 | Same exit class + refusing code for same source across check/build/run and both engines; one named table; closes W2, W3, N3 | Sections "Exit-code pin inventory", "admitProgram design", "Exitcode package + AST scan" |
| DX-18 | Unreachable functions pruned; run/build/check pass; ambiguous entry exit 2; byte-identical C for all-reachable programs | Sections "Entry call-site inventory", "Corpus audit", "Golden strategy" |
| EVD-13 | `app verify` keeps named capacity reason at any n; 16 exhausted cases within limit; never success | Section "App verify report" |
</phase_requirements>

## Summary

All three requirements are pure refactors/wiring inside existing packages; no library research is needed. The work is dominated by pin management: a large number of tests, two committed corpus goldens, and one CI-pinned workflow already encode today's behavior. The safety net DX-18 wants ("corpus golden committed before the change") **already exists in two forms**: `testdata/phase273/conformance-corpus-hashes.json` (EmitNative hash/error per accepted `.schway`) and `testdata/phase272/accepted-corpus-hashes.json` (core hash + EmitApplication hash/error), each guarded by a green identity test that writes goldens via an env var. Plan 1 should re-run those, add a thin Phase 28 behavior-class manifest (check / interpreter-run / admission class+code, entry source, dead-function list) and an exit-pin inventory, and commit before any code change.

Two facts change the shape of the work versus CONTEXT's assumptions. (1) `executionpeer` carries an **independent second copy** of the entry-resolution algorithm (`executionpeer.go:272` `resolveEntry`), so `main` precedence in `callgraph.EntryFunction` must not reach the peer through an unpruned multi-root program; the invariant "pruned program has exactly one in-degree-zero root, the entry" makes this safe in production paths. (2) The documented rule in `admitPublishedApplication` (D-27.1-02) says internal lanes (`RunNative`, `RunInterpreter`) must NOT run origin validation because `TestExclusiveBorrowInterpreterNative` and the 12-fixture divergence table rely on that; D-10 must therefore put admission in the **command-file layer** (`RunNativeCommandFile` / `RunInterpreterCommandFile`), not in the exported `RunNative` / `runNative` used by differential tests.

**Primary recommendation:** Wave 0 = pin inventory + Phase 28 class manifest + exitcode package and AST scan (behavior-neutral); Wave 1 = `PruneToReachable` + `main` precedence + `resolveEntry` + `admitProgram` and call-site collapse, with the listed pin flips; Wave 2 = verify-report digest + schema/2 + budget; Wave 3 = docs table, CI/guide script updates, receipts. Resolve the two Open Questions (dead-code origin gate, `check` regression on ambiguous-entry files) before Wave 1.

## Architectural Responsibility Map

| Capability | Primary Tier | Secondary Tier | Rationale |
|------------|-------------|----------------|-----------|
| Exit-code table (names/values) | `internal/compiler/exitcode` (leaf) | protocol (status->code), cmd/* (app-run domain), native (69), cgen (text) | Leaf with zero imports so every layer can consume it |
| Entry resolution + pruning | `callgraph` (algorithm: `EntryFunction`, `PruneToReachable`) | `session` (`resolveEntry` single call site) | callgraph owns adjacency; session owns pipeline placement |
| Admission (corevalidate -> originvalidate -> pathoracle) | `session` (`admitProgram`) | `CheckCommandFile`, build, run command-file wrappers | Independent validator layers stay separate; only orchestration + formatting shared |
| Status -> exit mapping | `protocol.ExitCode` (delegates to exitcode) | `cmd/schway` `emit()` | Already the single funnel for compiler-domain exits |
| App-run exit passthrough (124/125/128+n/65) | `cmd/schway` `runApplicationRun` | `native.RunOutcome` | Env/timeout-style domain; child status passthrough |
| Event-capacity contract (69) | `exitcode.EventCapacity` | `native` (alias `EventCapacityExitStatus`), `cgen` (C text) | Published contract baked into generated C |
| Verify report digest/budget | `session/session_app_verify.go` | `cmd/schway` exit mapping | Report schema owned by session |

## Standard Stack

Go 1.24.0 stdlib only (`go version` -> go1.24.0 darwin/arm64) [VERIFIED: `go version`, go.mod `go 1.24`]. No external packages are installed or recommended. `go/parser`, `go/ast`, `go/token`, `crypto/sha256`, `encoding/json` cover everything (D-11 AST scan, D-13 digest).

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `go/parser` AST scan | grep | Locked out by D-11 (grep misses positions, matches comments) |
| New hash-manifest golden | Reuse phase272/phase273 goldens | Reuse is cheaper; add only the class manifest (see Golden strategy) |

**Package Legitimacy Audit:** No external packages. Table omitted; `gsd-tools package-legitimacy` not applicable. Packages removed [SLOP]: none. Flagged [SUS]: none.

**Environment Availability**

| Dependency | Required By | Available | Version | Fallback |
|------------|------------|-----------|---------|----------|
| Go | everything | yes | 1.24.0 | — |
| clang | native/app-verify real-capacity tests, CLI matrix | yes | Apple clang 21.0.0 (`/usr/bin/clang`) | tests `t.Skip` when absent (CI fails loudly instead: ci.yml "Require clang") |
| Hosted CI (ubuntu+macos) | `go test ./...`, evidence-aggregate | not locally runnable | — | NEVER run full suite locally (AGENTS.md); focused runs only |

## Architecture Patterns

### System Architecture Diagram (target pipeline after D-04/D-09/D-10)

```
.schway source
   |
   v
session.Check(source)  --diagnostics--> exit 2 (checker code)          [full program, incl. Order/cycle]
   | ok
   v
resolveEntry(checked)  --entry_ambiguous--> exit 2 core.entry_ambiguous  [D-01/D-02; one callgraph.EntryFunction call site]
   |  returns (entry, pruned)         (PruneToReachable keeps decl order)
   v
admitProgram(pruned)   --corevalidate/originvalidate/pathoracle refusal--> diagnostics, nil err --> exit 2
   | admitted           (check stops here: reports pass)
   +--> build: cgen.EmitApplication(pruned) -> clang -> artifact          (tool failure -> exit 3)
   +--> run interpreter: interp.Run(pruned)                                (tool failure -> exit 3)
   +--> run native: Phase16ControlNativeC(pruned) -> clang O0/O3 -> compare (mismatch -> exit 4; tool -> exit 3)
   +--> app verify: interp + conformance C on pruned -> report schema /2   (not all pass -> exit 3)
exit code = exitcode table via protocol.ExitCode(status)
```

Order caveat: see Open Question 1 (whether corevalidate/originvalidate run on the full or pruned program).

### Recommended structure
```
internal/compiler/exitcode/exitcode.go        # consts + SignalExit(n) + pin helpers
internal/compiler/exitcode/exitcode_test.go   # pin (65, 69), uniqueness
internal/compiler/exitcode/scan_test.go       # AST scan + synthetic self-test (external pkg exitcode_test)
internal/compiler/callgraph/prune.go          # PruneToReachable (reuse buildAdjacency)
internal/compiler/session/session_entry.go    # resolveEntry, entryOf wrapper
internal/compiler/session/session_admit.go    # admitProgram
testdata/phase28/entry-exit-class-manifest.json
```

### Pattern: prune helper (must satisfy these invariants)
```go
// Source: derived from callgraph.go:279-345 (EntryFunction, reachableClosureSize, buildAdjacency)
// PruneToReachable(program, entryID): BFS from entryID over buildAdjacency, keep
// functions in original program.Functions order. Do NOT use the shape
// `len(x.Functions) != 1` (LANGUAGE-MATURITY census counts it) and do NOT
// index Functions[0] (TestSessionRunSitesDoNotIndexFunctionsZero).
// DataTypes / foreign decls / Module fields are copied unchanged.
```
Property to test: for a fully reachable program, `reflect.DeepEqual(PruneToReachable(p, e), p)`; for any program, `EntryFunction(pruned)` returns the same entry and `pruned` has exactly one in-degree-zero root.

### Anti-Patterns
- Calling `admitProgram` inside exported `RunNative`/`RunInterpreter` (breaks D-27.1-02 and `TestExclusiveBorrowInterpreterNative`; see below).
- New tests calling `cgen.Emit`/`cgen.EmitNative` directly: `session_phase16_emitter_inventory_test.go` tracks those calls against `testdata/phase16/public-emitter-consumers.json` [VERIFIED: session_phase16_emitter_inventory_test.go:12-80]. Use `session.Phase16ControlNativeC` in tests (as phase272/273 identity tests do).
- Changing `callgraphEntryAmbiguousError.Error()` text: two goldens pin it (below).

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Reachability | new graph walker | `buildAdjacency` + BFS like `reachableClosureSize` (callgraph.go:350) | Already handles unresolved callee refusal and deterministic order |
| Peer refusal formatting | new formatter | `peerRefusalDiagnostic` (session.go:839) + `originPeerRefusalDiagnostic` (session.go:849) | Keeps `core.peer_refusal_unnamed` fallback |
| sha256 hex | custom | `replayDigest` (session_app_verify.go) | Existing helper |
| Corpus golden writer | new tool | env-guarded `SCHWAY_PHASE273_WRITE_GOLDEN`-style pattern in `session_phase273_identity_test.go` | Established convention |
| AST scans | grep | pattern in `session_phase16_emitter_inventory_test.go` / `self_describing_docs_test.go` | Established convention |

## Entry call-site inventory (all non-test; D-04)

[VERIFIED: grep + Read this session] `callgraph.EntryFunction` callers:

| File:line | Function | Disposition |
|-----------|----------|-------------|
| session.go:511 | `ProjectExecutionSchema2` (exported; tests call with whole programs) | route through `entryOf` wrapper (program already pruned in pipeline) |
| session.go:602 | `Phase4CheckedProgram` (corpus lanes) | `resolveEntry` |
| session.go:1122 | `RunInterpreter` | `resolveEntry` (no origin admission, see below) |
| session.go:1209 | `RunInterpreterApplication` | `resolveEntry` + `admitProgram` |
| session.go:1235 | `interpreterInputs` (name pinned by `TestSessionRunSitesDoNotIndexFunctionsZero`, callgraph_entry_test.go:254) | accept already-resolved entry or call `entryOf`; keep the function name and no `Functions[0]` |
| session.go:1340 | `runNative` | `resolveEntry` |
| session.go:1447 | `BuildApplication` | `resolveEntry` |
| session_phase5_mismatch.go:393 | mismatch reducer seed | `entryOf` |
| session_app_verify.go:263 | per-case loop in `VerifyApplicationCases` (re-resolves every case) | hoist: resolve once before the loop |
| cgen/cgen_program.go:639 | `emitProgramWithShell` | **cannot import session**; keep (idempotent on pruned input) and allow-list in the meta-test |

Other mentions are test files or comments (core.go:759-773 constant `EntryAmbiguous = "core.entry_ambiguous"`; executionpeer.go:318 independent copy; reduce.go caller-supplied `EntryFunctionID`).

Meta-test: allow-list = callgraph pkg + `_test.go` + session `entryOf` (one call) + `cgen_program.go:639`. Mirror the AST style of `phase16PublicEmitterCalls`.

Current failure surface that DX-18 reclassifies [VERIFIED: session.go:1305-1316, cmd/schway/main.go:150-160]: interpreter entry error -> `RunInterpreterFile` err -> `tool.run_failed` StatusOperational (exit 3); `BuildApplication` err -> `native.build_failed` StatusOperational (exit 3); `CheckCommandFile` never calls EntryFunction (check passes). The `Error()` string is `callgraph.entry_ambiguous: ...` but `Code()` is `core.entry_ambiguous` (callgraph.go:218-227) — map via `callgraph.EntryAmbiguousError(err)` to `diagnostic.Error(ambiguous.Code(), diagnostic.Span{}, msg)`; put the "add fn main" hint in the diagnostic message, NOT in `Error()`.

### Corpus audit (D-03 / D-05 requirement)

Scratch test over all 155 git-tracked `.schway` files (temporary file, deleted) [VERIFIED: scratch go test this session]: 62 refused by checker, 1 refused by corevalidate, 90 accepted. Accepted programs by entry source:
- 61 single-function; all multi-function single-root programs have entry from the unique root (many also named `main`).
- **Tiebreak-without-`main` dependence: zero** among accepted programs that resolve today. Every multi-root program that resolves today contains a function named `main` (so D-01 `main` precedence gives the same entry).
- Multi-root (dead function) programs: `testdata/phase11/multi_function_unreachable.schway` (orphan), `testdata/phase11/multi_function_gate_corpus.schway` (touch), `testdata/phase13/heldout_shared_callee_twin_alpha.schway` and `..._mirror.schway` (5 fns, 2 roots), plus `testdata/phase11/multi_function_zero_call.schway` (tie, has `main`; refused today, **runs after D-01**).
- Ambiguous with no `main`: `testdata/phase12/payload_borrow_interaction.schway` (choose/identity, equal closures). Stays ambiguous; **`check` newly refuses it** (see Open Question 2).
So byte-identity is at risk for exactly 4 resolving files (gate_corpus, unreachable, twin_alpha, twin_mirror) plus 1 flip (zero_call).

## Admission design (D-09 / D-10)

[VERIFIED: session.go:1066-1106 CheckCommandFile] order today: Check diagnostics, then (unless `checkCommandPeerSeam`) `corevalidate.Validate` -> `peerRefusalDiagnostic`, then `originvalidate.ValidatePublished` -> `originPeerRefusalDiagnostic`, then `pathoracle.ValidateLocalOwnerPaths` -> `diagnostic.Error("core.pathoracle_refused", ...)`.

Where the five surfaces differ today (all must converge on StatusInvalid, same code):
| Surface | Core refusal | Origin refusal | Pathoracle |
|---------|--------------|----------------|------------|
| check (1066) | diagnostic (exit 2) | diagnostic with span | yes |
| build `BuildApplication` (1431) via `admitPublishedApplication` (1170) | **`fmt.Errorf("core validation failed: %s")` -> `native.build_failed` exit 3** | diagnostic (exit 2) | only for local-owner entries |
| `RunInterpreter` (1109) | `Problems[0].Code` **unguarded index** (1116) -> error -> exit 3 | n/a | no |
| `RunInterpreterCommandFile` (1305) | error -> `tool.run_failed` exit 3 | post-hoc `publishedOriginProblemFile` (1289, plain diagnostic without span) | no |
| `runNative` (1328) | **unguarded `Problems[0].Code`** (~1335) -> error | post-hoc in `RunNativeCommandFile` only if native err == nil | no |

Design that satisfies D-09/D-10 and the documented constraint:
1. `admitProgram(program core.Program, tree syntax.Tree) (admitted core.Program, diags []diagnostic.Diagnostic)`: corevalidate (guard `len(Problems)`), originvalidate, pathoracle; honors `checkCommandPeerSeam`/`observeCheckCommandPeer` for `check`.
2. `CheckCommandFile`, `BuildApplication`, `RunInterpreterApplication`, **`RunInterpreterCommandFile`, `RunNativeCommandFile`** call it. Keep exported `RunInterpreter`/`RunNative`/`runNative` (lane/test entry points) free of origin admission: `admitPublishedApplication`'s own comment says "Internal execution lanes (runNative, RunInterpreter, RunNativeCommandFile) deliberately do not call it ... Do not widen this gate without a new decision (D-27.1-02)" [VERIFIED: session.go:1160-1169] — and `RunNativeCommandFile` is named there, so D-10 is the "new decision" for the CLI layer only. Implement by splitting `runNative` into check/resolve + `runNativeChecked(ctx, checked, pruned, runner)`; the command-file wrapper does Check -> resolveEntry -> admitProgram -> `runNativeChecked`; `RunNative` does Check -> resolveEntry -> `runNativeChecked`.
3. Standardize on the span-carrying `originPeerRefusalDiagnostic`; `run --json` origin diagnostics gain span info (codes unchanged).
4. `foreign_origin_omitted` on native: admission precedes `Phase16ControlNativeC`/clang, so exit 2 `core.foreign_origin_omitted`. Cut-M004 fixtures (check clean) still reach cgen refusal -> exit 3 `native.tool_failure`.

`PATH=<empty>` native run remains a genuine tool failure (exit 3) because admission passes first.

### Fixtures for the new cross-surface assertion
- corevalidate refusal: `testdata/phase07/duplicate_function_name.schway` — checker clean, `corevalidate` code `core.duplicate_function_id` [VERIFIED: session_phase271_admission_test.go:136-141]. Today: check exit 2, build exit 3 (`native.build_failed`), interpreter exit 3 (`tool.run_failed`), native exit 3.
- origin refusal: `testdata/phase4/foreign_origin_omitted.schway` -> `core.foreign_origin_omitted`.

## Exit-code pin inventory (D-12; PITFALLS 13)

### Production literals (all must come from the table)
| File:line | Literal | Replacement |
|-----------|---------|-------------|
| protocol/protocol.go:681-697 | `func ExitCode(status string) int` returns 0/2/3/4/default 64 | delegate to exitcode |
| cmd/schway/main.go:25-28 | `exitSuccess = 0`, `exitInvalidSource = 2`, `exitOperational = 3`, `exitUsage = 64` | exitcode consts (`exitInvalidSource` is otherwise unused) |
| cmd/schway/main.go:242 | `return 65` (`native.input_too_long` / `input_contains_nul`) | `exitcode.AppInputRefused` |
| cmd/schway/main.go:245,257,260 | `return 125` | `exitcode.AppLaunchError` |
| cmd/schway/main.go:251 | `128 + outcome.SignalNumber` | `exitcode.SignalExit(n)` |
| cmd/schway/main.go:254 | `return 124` | `exitcode.AppTimeout` |
| cmd/schway-repair/main.go:12-15 | `exitSuccess 0, exitUnrepaired 1, exitOperational 3, exitUsage 64` | exitcode consts |
| native/native.go:37 | `EventCapacityExitStatus = 69` | `= exitcode.EventCapacity` (alias kept; `native.go:630` compares it) |
| native/native_app.go:424,432 | `ExitCode: 0`, `exitError.ExitCode()` | `exitcode.Success` for the literal |
| native/sanitize.go:226-287 | `exitCode == 0`, `exitCode := -1` (sanitizer child status, not a schway exit) | allow 0/-1 in scan or use `exitcode.Success` |
| cgen/cgen.go:735 | `exit(65)` inside C string (output limit) | build string from `strconv.Itoa(exitcode.AppInputRefused)` or assert |
| cgen/cgen.go:774 | `exit(69)` inside the long recorder string | split the string so `69` comes from `exitcode.EventCapacity`; output must be byte-identical (both goldens hash it) |
| cgen/cgen_program.go:1042,1045,1047,1074,1094,1097,1099,1200,1201,1203,1661,1663,1744,1760,1805,1807,1810,2690,2706,2754 | `return 65;` / `exit(65)` in emitted C | prefer one **string-literal scan test** over cgen sources asserting every `exit(N)`/`return N;` N is in {0, AppInputRefused, EventCapacity} — zero emitter edits, zero byte risk |
| session/session_phase5*.go | `report.ExitCode == 0` (sanitizer) | no change (value 0) |

The generated-C 69 string is also asserted in `cgen_phase273_test.go:87-97` (`TestPhase273CapacityContractLiteralsAgree`).

### Test pins that WILL flip (edit deliberately, with dated note)
| Pin | Today | After |
|-----|-------|-------|
| native/native_test.go:1522-1527 `cutAfterCheckRefusal`; used at 1564 for `foreign_origin_omitted.schway` | native exit 3 `native.tool_failure` | exit 2 `core.foreign_origin_omitted`; delete helper, use `refused(...)` |
| session/session_test.go:1818-1822 (native `foreign_origin_omitted`) | StatusOperational / `native.tool_failure` | StatusInvalid / `core.foreign_origin_omitted` |
| session/session_phase16_production_paths_test.go:20-37 `TestPhase16ProductionPathsPreserveM004Refusal` | `restrict_borrow.schway` native -> `native.tool_failure` | **flips**: `testdata/phase5/restrict_borrow.schway` is origin-refused by `check` (listed in `knownAdmissionDivergences` with `core.origin_omitted`, session_admission_divergence_test.go:75-88), so admission now yields StatusInvalid `core.origin_omitted`. `nonlocal_exit_probe.schway` (check-clean) stays `native.tool_failure`. **Not listed in CONTEXT D-12.** Replace the restrict_borrow entry with a check-clean cut fixture (e.g. `testdata/phase4/acquire_three_success.schway`) so the M004-refusal-preserved control keeps a native witness. |
| session/session_phase271_admission_test.go:169,176 (`corevalidate/build`, `corevalidate/interp`) | exact `err.Error() == "core validation failed: core.duplicate_function_id"`, no diagnostics | nil err + diagnostic with that code |
| session/session_phase271_admission_test.go:195 (`origin/interp` via `RunInterpreterApplication`) | err `"origin validation failed: core.origin_omitted"` | diagnostic (D-09 says RunInterpreter* use admitProgram) |
| session/session_phase271_admission_test.go:182,201,248 (`*/verify`) | verify errors `"source core validation failed"`, `"source origin validation failed: ..."` | **unchanged unless planner chooses** — D-09 does not list `VerifyApplicationCases`; app verify refusal stays exit 3. Record as a conscious exception (gap vs. DX-17 "same class"). |
| callgraph/callgraph_entry_test.go:130-158 `TestEntryFunctionEmitsUnreachableFunctions` | asserts `orphan` still present in checked program | rewrite to pruning contract (orphan absent from pruned, present in checked) + dated note (D-05) |
| cgen/cgen_program_test.go:1239-1281 `TestEmitProgramZeroCallEdges` and session/session_phase11_differential_test.go:357-396 `ZeroCallEdges` | assert `EntryAmbiguousError` for `multi_function_zero_call.schway` | `main` precedence resolves; rewrite to assert entry `main`, `unused` pruned, runs on all tiers; update the fixture header comment |
| session/session_payload_replay_test.go:186 `payloadCorpusExpectedSkips["testdata/phase11/multi_function_zero_call.schway"]` | listed as expected skip; test errors if a skip "no longer skips" (line ~300) | move to `payloadCorpusBaseline` with new digest (and `payloadCorpusPhase15Schema2Baseline` if schema-2) |
| session/session_admission_divergence_test.go `knownAdmissionDivergences` | sweep compares `Check+corevalidate` vs `CheckCommandFile` | (a) add `testdata/phase12/payload_borrow_interaction.schway` -> `core.entry_ambiguous` if `check` calls `resolveEntry` (Open Question 2); (b) `testdata/phase11/multi_function_gate_corpus.schway` entry ("core.origin_omitted", dead `touch`) **disappears** if origin validation runs on the pruned program (Open Question 1) — test fails on "RESOLVED" unless removed |
| testdata/phase273/conformance-corpus-hashes.json and testdata/phase272/accepted-corpus-hashes.json | pin `emitNativeError`/`emitApplicationError` text `callgraph.entry_ambiguous: program has 2 in-degree-zero root(s) ...` for zero_call and payload_borrow_interaction | zero_call entries change (hash or new error) in both; payload_borrow_interaction text unchanged iff `Error()` unchanged. If emitters keep running on the unpruned `checked.Program` (as these identity tests do), the 4 dead-function hashes do NOT change — but then zero_call's EmitNative now succeeds via `main` precedence inside `callgraph.EntryFunction`, so its golden entry must be regenerated once, listed in an allow-list |
| session/session_peer_gate_test.go:537 area | asserts check/interface peer refusal is StatusInvalid | unchanged (read to confirm) |

### Pins to PRESERVE (re-run before and after)
- 65: `cmd/schway/main_test.go` 105-112, 131, 147, 171, 191, 215, 234, 257, 263, 278, 334-376 (`TestPhase22*`, `TestPhase26PublicAppCLI`, `TestPhase27*`, `TestPhase23PublicFileByte`); `scripts/verify-phase27.sh:104,108` and `scripts/check-phase27-receipts.sh:99`; `examples/phase27/README.md:41,53`.
- 64: `cmd/schway/main_test.go:115`, `testsupport/cli_phase6_test.go:499`.
- 69: `native/native.go:630`, `cgen/cgen_phase273_test.go:87-97`, `native_test.go` (`TestPhase273RunnerEventCapacityAdjudication`), `session_phase273_appverify_test.go`.
- Compiler domain (2/3/4): `testsupport/cli_test.go:70,208,231`; `session_test.go:778` (`protocol.ExitCode(operational.Status) != 3`); `native_test.go:244` (check `foreign_origin_omitted` exit 2), `1507`/`1518` (cut fixtures exit 3 `native.tool_failure` — KEEP), `1155`-area `phase4CorpusMatrix`.
- `scripts/verify-phase25.sh` `65` hits are U64 *values* (0x41), not exit codes.
- No other script or `.github/workflows/ci.yml` step asserts a schway exit code; `ci.yml` runs `go test ./...`, `verify-phase27/24/25/23/6.sh`, and fixed `-run` lists (session_phase6_test.go `TestCIWorkflowRunsCurrentAggregateGate` pins workflow text). Adding tests does not require workflow edits; renaming `TestPhase27DynamicZero`, `TestPhase27OutputExactLimit|CrossingWrite`, `TestPhase27OutputPublicStatus`, `TestDecodeExecutionSchema2AdmissionSeam`, the named session seams in ci.yml, or `TestPhase26*` would break them.

### New pins required (D-12)
- 124/125/128+n: extract a pure `appRunExitCode(outcome native.RunOutcome, err error) int` from `runApplicationRun` (main.go:236-262) and table-test it (`RunTimedOut`, `RunLaunchError`, `RunSignaled`+signal, default invalid, `ToolError` input codes, plain err -> 125). A real subprocess timeout test would burn the 30 s `defaultSubprocessTimeout` (native.go:28). Add one real `app run /nonexistent -- 1` -> 125 CLI case.
- Cross-surface fixture test (below) with both engines.
- Exit-table docs test pinning each documented row to `exitcode` constants.

## Exitcode package + AST scan (D-07 / D-11)

Leaf package, zero imports; consts in D-07 table; `SignalExit(n int) int { return SignalBase + n }`. `protocol.ExitCode` keeps the `Status*` string switch but returns `exitcode.*`. `protocol` currently imports `cache`, `diagnostic`, `interp` — adding a leaf import is cycle-safe. The AST scan lives in `package exitcode_test` and walks files via `testsupport.ProjectPath`.

Detector spec (flag integer BasicLit values other than 0 and -1 unless the planner decides to also require `exitcode.Success`): argument of `os.Exit`; operand of `return` in `func ... int` in cmd/*/main.go and in functions named `ExitCode`/`*Exit*`; value of const/var/field whose name matches `(?i)exit`; either side of a comparison/`+` whose other side is an ident/selector matching `(?i)exit(code|status)?`; case clause values of a switch whose tag matches `(?i)(status|exitcode)`. Self-test: synthetic sources, one per position, asserting the detector fires, plus a negative control (clean source, zero hits) so the test cannot pass vacuously. Scope exclusions: `cmd/schway-repair/repair.go` (applicability ranks) and the exitcode package.

Current in-scope false-positive risk: `native/sanitize.go` sanitizer child exit codes (0/-1 only).

## App verify report (D-13..D-17)

[VERIFIED: session_app_verify.go] constants: `ApplicationVerifySchema = "lang.app-verification/1"`, `MaxReplayCasesBytes = 64 * 1024`, `MaxReplayCaseCount = 16`. `ReplayCaseReport.Engines map[string]execution.Execution \`json:"engines,omitempty"\`` is set in three places: asymmetric exhaustion (`Status=Fail`, ~line 300), both-level exhaustion (~line 312-322, `unsupported` iff interpreter outcome equals expected, else `fail`), and complete evidence (`interpreter`,`O0`,`O3`). `eventCapacityDiagnostic(level)` builds `native.CodeEventCapacityExhausted + ": generated C exceeded its static event capacity at " + level + "; outcome not established, not a mismatch"`. `writeApplicationVerificationReport` rejects `len(encoded)+1 > MaxReplayCasesBytes` with `fmt.Errorf("application verification report exceeds %d bytes", ...)` (the N1 failure). `finalizeReplayReport` sets `Verified=false` for any non-pass.

Measured this session (scratch test, deleted) with current code, sum_to_n: n=15 report 7,010 B, n=172 65,334 B, **n=173 65,706 B (> 65,536 -> N1)**, n=1000 373,353 B; each verify call ~0.3 s (clang; per-case recovery recompiles). So a 16-case exhausted set costs ~34 clang compiles (1 batch + 16 singles per level), roughly 10-15 s locally; keep the real-clang 16-case test to one test.

Design:
- `ReplayCaseReport.OutcomeDigest *ReplayOutcomeDigest \`json:"outcome_digest,omitempty"\`` with fields `engine` ("interpreter"), `outcome_kind`, `outcome_value`, `event_count` (= `len(interpreted.Events)`), `evidence_sha256` (= `replayDigest(json.Marshal(interpreted[reportIndex]))`; the preimage is the *post-`ProjectExecutionSchema2`* document held in `interpreted[reportIndex]`), `evidence_bytes`, `exhausted_level`.
- Replace the three `Engines = {"interpreter": ...}` assignments for exhausted cases (the asymmetric branch and the both-level branch) with the digest; leave `Engines` nil for them. Complete-evidence cases unchanged.
- Bump to `lang.app-verification/2`. Consumers to update [VERIFIED: grep]: `session_app_verify.go:28` and `:201`; `cmd/schway/main_test.go:551` (uses the constant, auto-follows); `examples/phase22/README.md:100` text; `session_phase273_appverify_test.go:60-80` (`phase273RequireUnsupportedCapacity`, rewrite `len(Engines)==1` -> digest assert), `:274-279` (one-level `fail` test, rewrite to digest); `.planning` docs are history. No script, workflow, or example JSON references the schema string or the `engines` key. Other `.Engines` readers (pass/fail complete cases) at `cmd/schway/main_test.go:562-582`, `session_phase27_evidence_test.go:275-276`, `session_phase273_appverify_test.go:123,150` stay valid.
- `MaxReplayCaseIDBytes = 128` rejected in `DecodeReplayCases` (currently only empty/duplicate IDs are checked, session_app_verify.go DecodeReplayCases loop). Existing case file IDs: `examples/phase22/identity.cases.json` ("identity-7", "identity-42") plus test-built IDs are short; add printable-ASCII check.
- Budget: per-case digest report is ~650-800 B (128 ID + 64 input_id + 64 sha + ~150 diagnostic + fields), so 16 cases ~13 KB, well under `MaxReplayCasesBytes/2` = 32,768.
- Keep runtime check, but make it a named `ErrApplicationVerificationReportOversize` (size+limit), still no partial file (temp-file + rename already atomic; the check precedes `os.CreateTemp`). Test it with a synthetic oversized report.
- CLI: `cmd/schway/main.go` verify path returns `exitOperational` for both `err != nil` and `!report.Verified` (lines ~296-304) — unchanged, switch to exitcode names.

## Common Pitfalls

### Pitfall 1: Moving admission into exported lanes
**What goes wrong:** `TestExclusiveBorrowInterpreterNative` and the 12-fixture divergence set break.
**How to avoid:** admission in command-file layer only (see Admission design).

### Pitfall 2: Independent peer disagrees with `main` precedence
`executionpeer.resolveEntry` (executionpeer.go:272-318) re-implements the tiebreak without `main` precedence. Safe only because pruned programs have one root. Add the property test; do not hand unpruned multi-root programs to `Phase5CompareProgramEngines`. `executionpeer_test.go:218-228` (`TestIndependentEntryResolution`) must keep passing: do not change the peer's behavior.

### Pitfall 3: Dead-code gate loosening (see Open Question 1)
Pruning before originvalidate silently accepts a dead function with an omitted origin (`multi_function_gate_corpus.schway`).

### Pitfall 4: Golden churn from `Error()` text
Two goldens embed `callgraph.entry_ambiguous: program has 2 in-degree-zero root(s)...`. Leave `Error()` unchanged.

### Pitfall 5: Census/shape meta-tests
`.planning/LANGUAGE-MATURITY.md:301` states "19 `len(Functions) != 1` guards across 6 files" and `self_describing_docs_test.go` re-derives it by AST. A prune helper written with that shape changes the count; the cgen/pathoracle exact-count shape matchers (cgen_program.go:1214,1341,1364; pathoracle.go:414,428; originvalidate.go:868,882) only become *more* likely to match after pruning (no corpus regression: examples/phase24/*.schway are fully reachable single-root).

### Pitfall 6: Output changes for the 4 dead-function programs
Pruned C drops the dead function and `SCHWAY_EVENT_CAPACITY` (sized to total operation count, cgen.go:668/719). That is by design; list the 4 files explicitly in the golden diff allow-list.

### Pitfall 7: `phase11CompareFourTiers` returns the program used to find `orphan`
session_phase11_differential_test.go `UnreachableFunction` subtest looks up `orphan` in the returned program and asserts it is absent from events; if that helper starts returning the pruned program, `orphanID == ""` fatals. Keep the helper on the checked program.

### Pitfall 8: Timeouts
CLI verify uses a 2-minute context (main.go ~`context.WithTimeout(..., 2*time.Minute)`); a 16-exhausted-case run is ~15 s locally, fine, but don't add more.

## Code Examples

Exhausted-case digest (shape):
```go
// Source: session_app_verify.go (both-level branch, ~line 312); interpreted[...] is already schema-2 projected
raw, _ := json.Marshal(interpreted[reportIndex])
report.Cases[reportIndex].OutcomeDigest = &ReplayOutcomeDigest{
    Engine: "interpreter", OutcomeKind: actual.Kind, OutcomeValue: actual.Value,
    EventCount: len(interpreted[reportIndex].Events),
    EvidenceSHA256: replayDigest(raw), EvidenceBytes: len(raw), ExhaustedLevel: level,
}
```
Test: re-marshal the in-memory execution and re-hash; assert equal.

## State of the Art

| Old | Current | Impact |
|-----|---------|--------|
| D-11-05 "unreachable functions are emitted" | prune to closure of entry (D-04/D-05, dated note) | `TestEntryFunctionEmitsUnreachableFunctions` rewritten |
| `Engines={"interpreter"}` for exhausted cases | `outcome_digest` (schema /2) | report size independent of n |

## Assumptions Log

| # | Claim | Section | Risk if Wrong |
|---|-------|---------|---------------|
| A1 | Allow literals 0 and -1 in the AST scan (flag all other ints) rather than also requiring `exitcode.Success` | Exitcode package | Planner/user may want 0 named too; small extra churn in native_app.go/sanitize.go |
| A2 | Per-case digest report is ~650-800 B (computed from field sizes, not measured on the new struct) | App verify | Budget test will self-check; headroom is ~2.5x |
| A3 | zero_call EmitNative golden entries (phase272/273) are regenerated once and allow-listed, rather than keeping error text | Pin table | If the user wants goldens frozen, `main` precedence must live above cgen instead of in `callgraph.EntryFunction` |
| A4 | Keeping `VerifyApplicationCases` refusals as errors (exit 3) is acceptable although DX-17 says "same class" | Pin table | A source refusal in `app verify` stays exit 3; needs explicit user sign-off or a one-line scope addition |

## Open Questions

1. **Does D-04 mean corevalidate/originvalidate see only the pruned program?** (needs user/planner confirmation before Wave 1)
   - Known: D-04 bullet 1 says dead code with errors is still refused (checker); bullet 3 and D-09 say corevalidate/originvalidate/pathoracle see only `pruned`. `testdata/phase11/multi_function_gate_corpus.schway` has a dead `touch` with an omitted origin: today `check` refuses it (`core.origin_omitted`, divergence-table entry) and `run` refuses it via `publishedOriginProblemFile`. Under the literal reading both pass and the publication gate is loosened for dead functions; the divergence-table entry becomes "RESOLVED" and fails.
   - Recommendation: run `admitProgram` on the full checked program, then prune before interp/cgen/pathoracle-for-local-owner/app-verify. It is the same code size, keeps the 12-fixture divergence table and the golden's check classes unchanged, and matches D-04's stated "dead code with errors is still refused" principle. If the literal reading is intended, remove the gate_corpus entry and list the check class flip in the Phase 28 manifest.

2. **`check` newly refuses ambiguous-entry files that pass today** (D-04 says check calls the helper).
   - `testdata/phase12/payload_borrow_interaction.schway` checks clean today and will exit 2 `core.entry_ambiguous`. It is used via `session.Check` (not `CheckCommandFile`) elsewhere (core_test.go:910, pathoracle_payload_test.go, session_payload_control_test.go:264), so only the divergence sweep and any CLI corpus matrices see it. Recommendation: add it to `knownAdmissionDivergences` with a "why", and record in the Phase 28 manifest.

3. **Hosted-gate wording**: no hosted gate pins `entry_ambiguous` or the report schema; `.github/workflows/ci.yml` needs no edit unless the planner adds a named Phase 28 control to the evidence-aggregate (optional).

## Validation Architecture

### Test Framework
| Property | Value |
|----------|-------|
| Framework | Go `testing`, go1.24.0 |
| Config file | none (go.mod only) |
| Quick run command | `go test ./internal/compiler/exitcode ./internal/compiler/callgraph -count=1` |
| Full suite command | hosted CI only (`go test ./...`); NEVER locally (AGENTS.md) |

### Phase Requirements -> Test Map
| Req | Behavior | Test Type | Automated Command | File Exists? |
|-----|----------|-----------|-------------------|-------------|
| DX-17 | table pin: 65, 69, uniqueness, SignalExit | unit | `go test ./internal/compiler/exitcode -run 'TestExitCodeTable' -count=1` | Wave 0 |
| DX-17 | AST scan finds no duplicate literals + synthetic self-test | unit/AST | `go test ./internal/compiler/exitcode -run 'TestNoDuplicatedExitLiterals\|TestExitScanDetectorFires' -count=1` | Wave 0 |
| DX-17 | cgen emitted-C `exit(N)`/`return N;` literals within table | unit | `go test ./internal/compiler/cgen -run 'TestEmittedCExitLiteralsMatchTable\|TestPhase273CapacityContractLiteralsAgree' -count=1` | Wave 0 / exists |
| DX-17 | same exit class + code: check/build/run(interp)/run(native) for `duplicate_function_name.schway` and `foreign_origin_omitted.schway` | CLI (shipped binary, clang) | `go test ./internal/compiler/native -run 'TestShippedBinaryFourSubcommandCorpusMatrix\|TestExitClassesUniformAcrossSurfaces' -count=1` | matrix exists (update); new test Wave 1 |
| DX-17 | tool/IO failure stays 3; mismatch 4; cut-M004 stays 3 | CLI | `go test ./internal/compiler/native -run 'TestShippedBinaryExercisesEveryPhase4Behavior' -count=1`; `go test ./internal/compiler/session -run 'TestPhase16ProductionPathsPreserveM004Refusal\|TestNativeToolFailureIsOperational' -count=1` | exist (edit) |
| DX-17 | app-run 124/125/128+n/65 | unit | `go test ./cmd/schway -run 'TestAppRunExitCode' -count=1` (pure func) + existing `-run 'TestPhase22\|TestPhase27PublicAppCLI'` | new + exist |
| DX-17 | docs table matches constants | unit | `go test ./internal/compiler/session -run 'TestExitCodeDocsTable' -count=1` | Wave 3 |
| DX-18 | `PruneToReachable`: deep-equal on all-reachable, order preserved, unique root, idempotent entry | unit | `go test ./internal/compiler/callgraph -run 'TestPrune\|TestEntryFunction' -count=1` | Wave 1 |
| DX-18 | `main` precedence, tie/no-main refusal, message lists roots | unit | `go test ./internal/compiler/callgraph -run 'TestEntryFunction' -count=1` | exists (edit 130-158) |
| DX-18 | EntryFunction only via helper (meta-test) | AST | `go test ./internal/compiler/session -run 'TestEntryFunctionCalledOnlyViaResolveEntry' -count=1` | Wave 1 |
| DX-18 | byte identity: pruned-pipeline C == phase272/273 goldens except the 5 allow-listed files | golden | `go test ./internal/compiler/session -run 'TestPhase273ConformanceCorpusByteIdentity\|TestPhase272\|TestPhase28PrunedEmitMatchesGolden' -count=1` | exist + new |
| DX-18 | unreachable-function program passes check, builds, runs both engines | CLI | `go test ./internal/compiler/native -run 'TestUnreachableFunctionProgramRuns' -count=1` (use multi_function_unreachable / zero_call) | Wave 1 |
| DX-18 | ambiguous entry exits 2 `core.entry_ambiguous` on check/build/run | CLI | same new test, fixture `payload_borrow_interaction.schway` | Wave 1 |
| DX-18 | payload replay + admission divergence + phase11 differential edits | unit | `go test ./internal/compiler/session -run 'TestPayloadCorpusCharacterizationReplay\|TestCLICheckAdmissionDivergenceIsExactlyKnown\|TestPhase11InterproceduralDifferential' -count=1` | exist (edit) |
| EVD-13 | n=14 pass; n=15/172/173/1000 unsupported with `native.event_capacity_exhausted` prefix, report written | integration (clang) | `go test ./internal/compiler/session -run 'TestPhase273AppVerify\|TestAppVerifyExhaustedCases' -count=1` | exist (edit) + new |
| EVD-13 | 16 exhausted cases named, within limit, never success | integration + static | `go test ./internal/compiler/session -run 'TestAppVerifyReportBudget\|TestAppVerifySixteenExhausted' -count=1` | Wave 2 |
| EVD-13 | digest preimage re-hash; oversize named error, no partial file; ID bound | unit | `go test ./internal/compiler/session -run 'TestAppVerifyOutcomeDigest\|TestAppVerifyReportOversize\|TestDecodeReplayCases' -count=1` | Wave 2 |
| EVD-13 | CLI exit 3 + report for unsupported | CLI | `go test ./cmd/schway -run 'TestPhase273AppVerifyCapacityCLI\|TestPhase22AppVerify' -count=1` | exist (extend n) |

### Sampling Rate
- Per task commit: the focused `-run` for the touched package (above).
- Per wave merge: all focused rows for that wave plus `go test ./internal/compiler/session -run 'TestPhase16|TestPhase271|TestPhase272|TestPhase273|TestSelfDescribing' -count=1`.
- Phase gate: hosted full + race CI green on both OSes (not run locally).

### Wave 0 Gaps
- [ ] `testdata/phase28/entry-exit-class-manifest.json` + generator test (env-guarded writer like `SCHWAY_PHASE273_WRITE_GOLDEN`); fields per file: check class/code, interpreter-run class/code, admission class/code, entry source (`main`/`unique`/`tiebreak`/`ambiguous`), dead-function IDs. Generate and commit BEFORE any behavior change (use Wave 0 numbers from the corpus audit as cross-check).
- [ ] Baseline run (green) of `TestPhase272*`, `TestPhase273ConformanceCorpusByteIdentity` at HEAD.
- [ ] `internal/compiler/exitcode` package, pin test, AST scan + self-test (behavior-neutral; add the scan with a temporary allow-list of today's offenders, then shrink it as call sites move).
- [ ] Exit-pin inventory file (the tables above) re-grepped at plan time.

## Security Domain

### Applicable ASVS Categories
| ASVS Category | Applies | Standard Control |
|---------------|---------|-----------------|
| V5 Input Validation | yes | `DecodeReplayCases` bounded decode + new case-ID length/printable check; existing `RejectDuplicateJSONKeys`, `DisallowUnknownFields` |
| V12 Files | yes | report written via temp file + `Chmod 0600` + rename (existing); keep "no partial file" on oversize |
| V2/V3/V4/V6 | no | no auth, sessions, access control, crypto (sha256 is an integrity digest only) |

### Known Threat Patterns
| Pattern | STRIDE | Standard Mitigation |
|---------|--------|---------------------|
| Report-size DoS / output flooding via large exhausted evidence | DoS | digest-not-evidence + static budget test + runtime named oversize error |
| Misleading success from unsupported cases | Spoofing/Repudiation | `finalizeReplayReport` unchanged; `Verified=false`, exit 3 |
| Exit-code confusion between domains (child status vs host) | Spoofing | documented two-domain table; `schway app run:` stderr prefix + `--report` |

## Sources

### Primary (HIGH — read this session)
- `internal/compiler/callgraph/callgraph.go:205-345`, `callgraph_entry_test.go:125-330`
- `internal/compiler/session/session.go` (1066-1106, 1109-1330, 1328-1470, 1586-1640), `session_app_verify.go` (whole file), `session_phase273_appverify_test.go`, `session_phase271_admission_test.go`, `session_admission_divergence_test.go`, `session_payload_replay_test.go`, `session_phase273_identity_test.go`, `session_phase16_production_paths_test.go`, `session_test.go:1805-1825,2095-2135`
- `cmd/schway/main.go`, `cmd/schway/main_test.go`, `cmd/schway-repair/main.go`
- `internal/compiler/protocol/protocol.go:681-697`, `native/native.go:28-45,630`, `native_test.go:195-330,1470-1620`
- `internal/compiler/cgen/cgen.go`, `cgen_program.go:585-700`, `cgen_phase273_test.go:80-110`, `cgen_program_test.go:1225-1290`
- `internal/compiler/executionpeer/executionpeer.go:270-322`
- `scripts/verify-phase27.sh`, `scripts/check-phase27-receipts.sh`, `.github/workflows/ci.yml`
- `testdata/phase272/accepted-corpus-hashes.json`, `testdata/phase273/conformance-corpus-hashes.json`
- Scratch runs (deleted): corpus entry audit over 155 files; n=15/172/173/1000 report sizes.

### Secondary / Tertiary
- None (no web research needed; no external libraries).

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH — stdlib only, confirmed.
- Architecture / call sites / pins: HIGH — every anchor read from source.
- Pitfalls: HIGH for items cited with file:line; MEDIUM for A2 (size estimate).
- Open Questions 1 and 2 are real decision-consistency gaps in CONTEXT, not research gaps.

**Research date:** 2026-10-10
**Valid until:** until Phase 28 execution begins (anchors are line-numbered against HEAD a3b5f77; re-grep at plan time).
