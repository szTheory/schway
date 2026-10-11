# Phase 28: Exit Classes, Verify Report, and Entry Pruning - Context

**Gathered:** 2026-10-10
**Status:** Ready for planning
**Method:** Advisor mode (minimal_decisive tier). Three parallel research passes (entry/pruning, exit-code table, N1 digest), each multi-lens with an adversarial pass, synthesized into one recommendation per area and auto-followed per the user's standing instruction (memory: fan-out-then-auto-follow). One research recommendation was overridden after code verification (D-03, noted inline).

<domain>
## Phase Boundary

Make exit classes uniform and named, let programs with unused functions run, and keep `app verify`'s named capacity reason at any n. Requirements DX-17, DX-18, EVD-13 (closes M005 debt W2, W3, N1; folds N3). No new language features, no byte views, no modules — those are Phases 29–31. No new dependencies.

</domain>

<decisions>
## Implementation Decisions

### Entry selection and pruning (DX-18)
- **D-01: Entry rule.** (1) If a function named exactly `main` exists, it is the entry. (2) Otherwise keep the existing adopted D-11-05 rule: the unique in-degree-zero root, or among several roots the one with the strictly largest reachable closure. (3) Otherwise (zero functions, no root, or a closure-size tie) refuse with `core.entry_ambiguous`. The refusal message lists the roots and says to add `fn main` to disambiguate (`Roots()` already exists). `main` precedence closes today's fail-open case where an orphan chain with a larger closure than `main` silently becomes the entry. — **Reversibility:** costly — every entry-resolution site and the corpus golden depend on it.
- **D-02: `entry_ambiguous` is a source refusal (exit 2)** from `check`, `build`, `run --engine=interpreter`, and `run --engine=native`. It is no longer operational (exit 3). The executionpeer.go:318 peer-document refusal keeps its code. Verify that it cannot surface for source programs.
- **D-03: Keep the closure-size tiebreak for programs without `main`.** OVERRIDE of the research recommendation to remove it. The tiebreak is the adopted, tested contract (D-11-05, `callgraph_entry_test.go:130`). Removing it would turn running programs without `main` into refusals, which contradicts DX-18's "runs instead of failing". `main` precedence (D-01) already removes the dangerous case. The planner's corpus audit must still list every file whose entry comes from the tiebreak rather than a unique root or `main`, so the dependency is visible.
- **D-04: Pipeline placement: check everything, then prune once.**
  - The checker (including Order/cycle refusal) runs on the full program, so dead code with errors is still refused. This is Rust/Go behavior, not Zig's lazy-analysis footgun.
  - Then one session helper, e.g. `resolveEntry(checked) (entry, pruned, err)`, calls `callgraph.EntryFunction` and a new `callgraph.PruneToReachable(entry)` (BFS over the existing `buildAdjacency`). It preserves the original declaration order of the surviving functions.
  - corevalidate, originvalidate, pathoracle, interp, cgen and app verify see only `pruned`. `check` calls the same helper, so it reports `entry_ambiguous` exactly like `build` and `run`.
  - All current `EntryFunction` call sites collapse onto the helper: session.go 511, 602, 1122, 1209, 1235, 1340, 1447; session_phase5_mismatch.go:393; session_app_verify.go (~263); cgen_program.go:639.
  - An inventory-style meta-test fails if `EntryFunction` is called outside the helper (and callgraph's own tests).
- **D-05: Byte identity.**
  - Before any code change, commit a corpus golden over all git-tracked `.schway` files: generated-C hash for every buildable program, plus the check/build/run class and code.
  - For a fully reachable program, `PruneToReachable` returns a deep-equal program (test-asserted).
  - Programs with unreachable functions change output by design and are listed explicitly in the golden diff.
  - `TestEntryFunctionEmitsUnreachableFunctions` (D-11-05 "unreachable functions are emitted") is deliberately rewritten to the pruning contract, with a dated note.
- **D-06: Phase 30 composition hook.** Later, the entry search runs on the composed whole program, but candidates come only from the root module. Library exports are never roots and unused library exports are pruned (still checked). Phase 28 only needs the helper shaped so this composes; nothing module-specific is built now.

### Exit-code table (DX-17, N3, W2, W3)
- **D-07: New leaf package `internal/compiler/exitcode`** (imports nothing). It is the single named table, consumed by protocol, session, native, cgen, `cmd/schway` and `cmd/schway-repair`. `protocol.ExitCode(status)` becomes a thin delegate. The table names domains explicitly:

  | Name | Value | Meaning | Domain |
  |---|---|---|---|
  | Success | 0 | pass | all |
  | Invalid | 2 | source refusal (checker, corevalidate, originvalidate, pathoracle, entry resolution) | compiler |
  | Operational | 3 | tool / IO / timeout failure | compiler, repair |
  | Mismatch | 4 | engine disagreement | compiler |
  | Usage | 64 | EX_USAGE | all |
  | AppInputRefused | 65 | EX_DATAERR: host pre-launch input refusal and generated program checked failure — never renumber | app-run |
  | EventCapacity | 69 | EX_UNAVAILABLE: generated-C event capacity — never renumber; `native.EventCapacityExitStatus` aliases it | app-run |
  | AppTimeout | 124 | timeout(1) convention | app-run |
  | AppLaunchError | 125 | launch/tool/invalid-outcome failure, env(1) convention | app-run |
  | SignalBase | 128 | `SignalExit(n)` = 128+n, POSIX shell convention | app-run |
  | Unrepaired | 1 | `schway-repair` found nothing applicable | repair |

  Do not add 126/127, because nothing emits them.
  — **Reversibility:** one-way — 65 and 69 are published contracts baked into generated C, `scripts/verify-phase27.sh` and the receipts scripts.
- **D-08: `app run` stays in its env/timeout-style domain.** It takes a built artifact, not source, so source refusals never reach it. The child's exit status passes through unchanged. A collision with host codes is inherent to passthrough, so it is documented rather than remapped; the `schway app run:` stderr prefix and `--report` disambiguate. Its 65 for `native.input_too_long` / `input_contains_nul` stays.
- **D-09: W2 fix — core refusal returns a diagnostic, not an error.**
  - Factor one `admitProgram(pruned)`: corevalidate, then originvalidate, then pathoracle, in `CheckCommandFile`'s existing order. It returns `[]diagnostic.Diagnostic` with a nil error, using the `peerRefusalDiagnostic` formatting and keeping the `core.peer_refusal_unnamed` fallback (never index `Problems[0]` unguarded; today's sites at session.go ~1116 and ~1335 do).
  - `BuildApplication`, `RunInterpreter*` and `runNative` all use it, so refusals map to `StatusInvalid` → 2.
  - The layers stay independent derivations (07-10 refusing-union rule); only formatting is shared. The `checkCommandPeerSeam` / `observeCheckCommandPeer` test seams stay.
- **D-10: W3 fix — admit before any tool runs.**
  - `runNative` calls `admitProgram` after check/prune and before cgen or any clang invocation, so `foreign_origin_omitted` exits 2 `core.foreign_origin_omitted` on native.
  - The interpreter path admits up front too, and the post-hoc `publishedOriginProblemFile` checks are dropped: fail-closed order, never run what will be refused.
  - Genuine cut-M004 lowering fixtures (`native_test.go` ~1507, ~1518) stay exit 3 `native.tool_failure`. Do not over-correct them.
- **D-11: Static no-duplicated-literals test is an AST scan (go/parser), not grep.**
  - Scope: non-test files in session, native, protocol, cgen, `cmd/schway`, and `cmd/schway-repair/main.go`. Exclude `repair.go` (applicability ranks 0–3 are not exit codes) and the exitcode package itself.
  - It flags INT literals in exit positions: `os.Exit` args; returns from exit-flow funcs; `exit*` consts/vars; `ExitCode`/`*ExitStatus` fields; comparisons or `128+x` against exit-named idents; switch cases on status/ExitCode.
  - It must ship with a self-test that feeds synthetic source for each position and asserts the detector fires (a negative control against passing vacuously).
  - A pin test asserts 65, 69 and the uniqueness of table values.
  - Generated-C `return 65` / `exit(65)` / `exit(69)` text is built from or test-asserted against the constants.
- **D-12: Enumerate pins before editing (PITFALLS 13).** The first plan re-greps and records every exit-code pin.
  - Expected to flip: `native_test.go` ~1524–1526 / ~1564, where native `foreign_origin_omitted` goes from 3 to 2; `cutAfterCheckRefusal` is deleted and the case uses `refused(...)`.
  - Preserve: the 65 pins in `cmd/schway/main_test.go` and `scripts/verify-phase27.sh` / `check-phase27-receipts.sh`; the 69 pins in native.go, native_test.go and `cgen_phase273_test.go:87-97`; the compiler-domain pins in testsupport `cli_test.go` / `cli_phase6_test.go`; and the `PATH=<empty>` native run, which is a genuine tool failure and stays 3.
  - Add pins for 124/125/128+n in `app run`, which are unpinned today.
  - Add a new corevalidate-refusal fixture asserting the same exit 2 and code across check/build/run on both engines.
  - Add a two-domain exit-code table to user docs (README or wiki).

### App verify report (EVD-13, N1)
- **D-13: Exhausted cases carry an outcome digest, not embedded interpreter evidence.** `ReplayCaseReport` gains `outcome_digest` (omitempty) and `Engines` is omitted for these cases. Fields: `engine` (interpreter), `outcome_kind`, `outcome_value`, `event_count`, `evidence_sha256`, `evidence_bytes`, `exhausted_level`. The sha256 preimage is the exact `json.Marshal` bytes of the omitted `execution.Execution`, test-asserted by re-marshal and re-hash. Pass and fail cases with complete evidence keep their full `Engines` map, unchanged.
- **D-14: Statuses unchanged.**
  - Exhausted with the interpreter outcome equal to the expected outcome → `unsupported`, with a diagnostic starting `native.event_capacity_exhausted` (`eventCapacityDiagnostic(level)` verbatim).
  - Interpreter outcome different from expected → `fail`.
  - Exhausted at one optimization level only → `fail` (WR-01 divergence) and still carries the digest.
  - `unsupported` never counts as success: `Verified=false`, exit 3, via existing `finalizeReplayReport`. Do not change that.
- **D-15: Schema bump to `lang.app-verification/2`.** Update consumers: the producer constant, `session_phase273_appverify_test.go` (rewrite the `len(Engines)==1` asserts to digest asserts — do not delete), `cmd/schway/main_test.go` if it pins the schema, and `examples/phase22/README.md` ~line 100. Re-grep `scripts/`, `.github/` and `examples/` for `lang.app-verification` and the `engines` key before landing. — **Reversibility:** costly — published report schema.
- **D-16: Static size budget.**
  - Add `MaxReplayCaseIDBytes = 128`, rejected in `DecodeReplayCases`, with non-printable IDs also rejected. Check that no existing cases file exceeds it.
  - `TestAppVerifyReportBudget`: 16 (`MaxReplayCaseCount`) exhausted cases with maximal IDs, inputs and values must encode to at most `MaxReplayCasesBytes/2`. It is self-checking, so adding a field that breaks the budget fails.
  - Keep the runtime oversize check as defense in depth, as a named error (e.g. `ErrApplicationVerificationReportOversize`) with size and limit. Exit 3, no partial file. A synthetic test hits it deliberately.
- **D-17: Required evidence:** sum_to_n n=14 passes; n=15, 172, 173 and 1000 are each `unsupported` with the named diagnostic and a written report; a 16-exhausted-case set names every case and stays within the limit; the one-level exhaustion test keeps asserting `fail`.

### Post-research amendments (2026-10-10, auto-followed per the research recommendations)
- **D-18: Admit the full program, then prune (refines D-04 and D-09).** `admitProgram` (corevalidate → originvalidate → pathoracle) runs on the full checked program. `resolveEntry` prunes only after admission, before interp, cgen and app verify. Dead code with a validator refusal, e.g. `multi_function_gate_corpus`, stays refused, so the publication gate never loosens. This follows D-04's own principle: dead code with errors is still refused.
- **D-19: Admission lives in the command-file layer only (refines D-10).**
  - Wire `admitProgram` into `check`, `build`, `RunInterpreterCommandFile` and `RunNativeCommandFile`. Split `runNative` into a check-resolve-admit step and a `runNativeChecked` step.
  - Internal differential lanes, `admitPublishedApplication` and the D-27.1-02 rule ("internal lanes do not run origin admission") stay unchanged. `TestExclusiveBorrowInterpreterNative` and the 12-fixture divergence table keep passing.
  - `executionpeer.resolveEntry` keeps its copy of the tiebreak. Add a property test that a pruned program has exactly one in-degree-zero root, and never hand an unpruned multi-root program to the peer.
- **D-20: Pins outside the D-12 list that also move:**
  - `TestPhase16ProductionPathsPreserveM004Refusal`: swap in a cut fixture that passes check, replacing `restrict_borrow.schway`.
  - The exact error-text asserts in `session_phase271_admission_test.go` (~169, 176, 195) become diagnostic asserts.
  - The `multi_function_zero_call` pins flip to "runs":
    - `TestEmitProgramZeroCallEdges`
    - the `ZeroCallEdges` subtest in `session_phase11_differential_test.go`
    - `payloadCorpusExpectedSkips`
    - both corpus goldens
  - `payload_borrow_interaction.schway` joins `knownAdmissionDivergences` as `core.entry_ambiguous`.
  - `callgraph`'s `Error()` text stays byte-identical; the "add fn main" hint goes only in the diagnostic message.
  - Reuse the existing corpus goldens (`testdata/phase273/conformance-corpus-hashes.json`, `testdata/phase272/accepted-corpus-hashes.json`) as the D-05 safety net, plus a thin `testdata/phase28` class manifest. New tests emit C through `Phase16ControlNativeC`, never `cgen.Emit*`.
- **D-21: `app verify` source refusals also exit 2, and the static scan covers every exit path (closes the DX-17 gap).**
  - `VerifyApplicationCases` runs the same check, admit and resolve steps. A source refusal gives exit 2 with the refusing code, while capacity and tool outcomes keep exit 3 and `unsupported` semantics.
  - The D-11 AST scan flags every integer literal passed to `os.Exit` or returned as an exit status from the cmd, session and native exit paths, including `0`. Production code uses `exitcode.Success`. A non-exit sentinel such as `-1` is out of scope.
  - cgen-emitted C exit strings are pinned by a string-literal scan test rather than edited.

- **D-22: `cmd/schway-repair` keeps a pinned mirror (refines D-07 and D-11; auto-followed after the plan check).**
  - Rule D-06-28 (`TestRepairDriverImportsStayOutsideInternal`) forbids the repair driver from importing `internal/`. That isolation is deliberate, and the project prefers a copy to a new dependency edge.
  - So `exitSuccess`, `exitUnrepaired`, `exitOperational` and `exitUsage` stay as named constants in `cmd/schway-repair/main.go`. A test in `internal/compiler/exitcode` parses that file and fails unless each value equals its `exitcode` counterpart (Success, Unrepaired, Operational, Usage).
  - The D-11 scan exempts only those four declarations and still flags every other exit-position literal in the file. The table stays the single source of truth; the mirror is test-pinned to it.
- **D-23: The corpus goldens stay frozen; allow-lists replace rewriting them (refines D-20).** The phase273 golden cannot be safely regenerated because of its sum_to_n reverse-substitution rule. So the `multi_function_zero_call` flip and the new Phase 28 example are recorded as explicit allow-list entries in the identity tests, and each entry names its reason. The test-visible result matches D-20.
- **D-24: The D-12 corevalidate fixture is `testdata/phase07/duplicate_function_name.schway` (refines D-12).** It passes the checker and is refused by the core-validator peer (`session_peer_gate_test.go`), so it is a genuine corevalidate refusal and no new `.schway` file is needed.

### Claude's Discretion
- Exact helper names (`resolveEntry`, `admitProgram`, `PruneToReachable`), file split, and plan/wave decomposition.
- Golden format: a per-file hash manifest or stored C text, whichever fits existing testdata conventions.
- Whether cgen formats exit literals from the constants or test-asserts them.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone scope and research
- `.planning/REQUIREMENTS.md` — DX-17, DX-18, EVD-13 (exact acceptance wording)
- `.planning/ROADMAP.md` — Phase 28 goal, success criteria 1–4, research flag
- `.planning/research/SUMMARY.md` — Resolved Decisions row 9 (carry-over debt), row 10 (phase order)
- `.planning/research/PITFALLS.md` — pitfall 13 (enumerate exit pins first), pitfall 1 (max-size and N1), pitfall 16 (debt growth)
- `AGENTS.md` — project rules (focused tests only, never run full suites locally, PRODUCT-ROADMAP amendment rule)

### M005 debt being closed
- `.planning/milestones/M005-MILESTONE-AUDIT.md` — N1 (line ~31), W1/W2/W3, N3 definitions
- `.planning/milestones/M005-phases/27.3-address-tech-debt-app-verify-capacity-abort-and-guide-clean/` — 27.3 plans and summaries (event-capacity 69 contract, WR-01 one-level divergence, app verify behavior)

### Code anchors
- `internal/compiler/callgraph/callgraph.go:279` — `EntryFunction` and the closure-size tiebreak
- `internal/compiler/callgraph/callgraph_entry_test.go:130` — D-11-05 contract test to rewrite
- `internal/compiler/protocol/protocol.go:681` — `ExitCode(status)`
- `cmd/schway/main.go` (~225–262 app run literals, ~156 runBuild, ~654) and `cmd/schway-repair/main.go:12-15`
- `internal/compiler/native/native.go:36-37` — event capacity code and 69
- `internal/compiler/session/session.go` — `CheckCommandFile` (~1062), `applicationAdmission` (~1170), `BuildApplication` (~1431), `runNative` (~1333), `RunNativeCommandFile` (~1604)
- `internal/compiler/session/session_app_verify.go` — `MaxReplayCasesBytes` (29), exhausted branches (~334–355), `writeApplicationVerificationReport` (~552)
- `internal/compiler/native/native_test.go:1507-1564` — cut and native-refusal exit pins

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `callgraph.buildAdjacency` and `reachableClosureSize`: BFS machinery for `PruneToReachable`.
- `entryAmbiguousError.Roots()`: witness data for the refusal message.
- `peerRefusalDiagnostic` and the `core.peer_refusal_unnamed` fallback: shared refusal formatting.
- `finalizeReplayReport`: already enforces "only pass counts".
- `replayDigest` helper in session: sha256 for the outcome digest.

### Established Patterns
- Independent validator layers (07-10 refusing union): share formatting, never reconcile derivations.
- Inventory/groundedness meta-tests in session: the style for the EntryFunction-call-site and exit-literal guards.
- Fail-closed: refuse rather than guess; no partial report files; atomic temp-file writes.

### Integration Points
- Hard-coded fixture shape matchers count functions and names and may shift if their fixtures contain dead functions: cgen_program.go (1159, 1214, 1341, 1374), interp.go (212–269), pathoracle.go:413. Run the focused phase24/phase5 fixtures before and after pruning.
- cgen reserves C `main` (cgen.go:171, 206): confirm source `main` mangling is unchanged.

</code_context>

<specifics>
## Specific Ideas

- Runnable gain to demonstrate: a library-style file with unused functions builds and runs; `schway app verify examples/sum_to_n.schway` at n = 1000 reports `unsupported` with `native.event_capacity_exhausted`.
- First plan should be the safety net: the exit-pin inventory plus the corpus golden, committed before any behavior change.

</specifics>

<deferred>
## Deferred Ideas

- Requiring an explicit `fn main` (or exported entry marker) for all programs: a deliberate language change for later, since 88 of 155 corpus files have no `main`.
- An evidence sidecar file so exhausted cases can be re-diffed: only if that becomes a requirement.
- Loop-aware event capacity covering sum_to_n's whole domain (D-27.3-03): already listed as Future in REQUIREMENTS.md.
- N4 and R1b: not scheduled (milestone decision).

</deferred>

---

*Phase: 28-exit-classes-verify-report-and-entry-pruning*
*Context gathered: 2026-10-10*
