---
phase: 11-multi-function-native-emission-and-interprocedural-equivalen
verified: 2026-09-12T00:00:00Z
status: human_needed
score: 5/5 must-haves verified
behavior_unverified: 0
overrides_applied: 0
human_verification:
  - test: "Confirm WR-01 (internal/compiler/reduce/reduce.go): reduce.Reduce never validates Seed.EntryFunctionID before dropOrphanFunction relies on it to exempt the real entry point from deletion. Run `go test ./internal/compiler/reduce/... -run TestReduceRejectsMultiFunctionSeed -v` and inspect whether an empty/wrong EntryFunctionID on a multi-function seed can delete the real entry function."
    expected: "Either accept the current state (single production caller, session_phase5_mismatch.go:386-390, derives the ID correctly via callgraph.EntryFunction) as sufficient for M002, or require a follow-up plan implementing 11-REVIEW.md's suggested Seed.Validate() fail-closed guard before QLT-05 is treated as hardened against misuse, not just against its one current caller."
    why_human: "This is a judgment call about acceptable residual risk in a package whose own doc comments elsewhere in this codebase insist on 'never guess, always refuse' — the code review (11-REVIEW.md WR-01) already flagged this as a real, reproducible hazard by construction (not hypothetical), not merely stylistic, but it does not block any current production path today."
  - test: "Confirm the mid-phase gate's disclosed CLI-check divergence (11-MIDPHASE-GATE.md, 'Known divergence' section): `go run ./cmd/lang --json check testdata/phase11/multi_function_gate_corpus.lang` reports status:invalid (core.origin_omitted) even though session.Check + corevalidate.Validate (the path every Phase 11 test and production run site uses) accepts the identical program cleanly. Reproducing this locally: `go run ./cmd/lang --json check testdata/phase5/restrict_borrow.lang` shows the same pre-existing divergence on the already-shipped Phase 5 fixture."
    expected: "Decide whether this pre-existing check-command/session.Check split (originvalidate.ValidatePublished's unconditional origin-declaration strictness vs. session.Check's admission) needs its own tracked debt item/fix, independent of Phase 11, since Phase 11's own gate never depends on the CLI check path."
    why_human: "The phase's own mid-phase gate ratification explicitly flags this for human review rather than resolving it silently, and it predates Phase 11 (reproducible on an existing Phase 5 fixture) — a judgment call on priority, not a Phase 11 defect."
---

# Phase 11: Multi-Function Native Emission and Interprocedural Equivalence Verification Report

**Phase Goal:** A multi-function Lang program lowers to readable C17 and
executes identically under the interpreter, `-O0`, `-O3`, and `-O3 -flto` —
with the optimizer proven to have actually been given something to optimize.
**Verified:** 2026-09-12
**Status:** human_needed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth (ROADMAP success criterion) | Status | Evidence |
|---|---|---|---|
| 1 | `cgen` emits multi-function C17 (`Emit`/`EmitNative` no longer refuse `len(Functions) != 1`), and every alias/capture promise at a call boundary names the checked fact it derives from | ✓ VERIFIED | `cgen.go` `Emit`/`EmitNative` dispatch to `emitProgram` (`cgen_program.go`) for N-function programs, confirmed by direct read. Independent re-run of the roadmap's own scoping command: `awk '/^func /{f=$0} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')` = **26** matches today (down from the pre-phase 32/re-verified-29 baseline), with every remaining site accounted for in `11-GUARD-LEDGER.md` and traced by hand in this session — the 3-guard reduction (29→26) exactly matches `11-08`'s `reduce.go` widening (2 sites) plus `11-09`'s `foreignCallSequenceFor` widening (1 site), both independently confirmed by reading the current source (reduce.go now uses `== 1` for the single-function *projection* case only; `foreignCallSequenceFor` no longer early-returns on N != 1). NAT-05's attribute-naming requirement is satisfied by an explicit, disclosed **empty set** (D-11-10, ratified in `11-NAT07-EVIDENCE.md` item 4) — zero call-boundary alias attributes are emitted by construction in M002's language, with a generated comment recording why; this is honestly reported as "weakened by evidence," not a clean pass of the literal prose, and the mid-phase gate (`11-MIDPHASE-GATE.md`) independently proves the emptiness non-vacuous via a real mutation-kill (`TestPhase11GateMutationKill`, re-run this session, PASS, going red under `cgen.AttributesJustified`). |
| 2 | Interpreter, `-O0`, `-O3`, `-O3 -flto` produce equivalent semantic outcomes/events for the interprocedural corpus | ✓ VERIFIED | `TestPhase11InterproceduralDifferential` (`session_phase11_differential_test.go`) re-run this session: PASS across 5 named fixtures (`multi_function_entry_basic`, `multi_function_forward_callee`, `multi_function_unreachable`, `multi_function_gate_corpus`, `multi_function_relay_depth2`) plus `ZeroCallEdges`/`UnreachableFunction`/`ForwardDefinedCallee`/`DiamondSharedLeaf` subtests. `DivergingCallee` is an honest, named `t.Skip` (not silently absent), correctly attributed to D-11-52 (a diverging callee cannot be lowered to native in a multi-function program this phase because `emitMatch`, deliberately not generalized per D-11-02, is refused inside multi-function programs). `DiamondSharedLeaf` demonstrates the honest, disclosed D-11-51 event-identity collision (interpreter succeeds with a duplicate-ID document; all three native tiers structurally and consistently refuse it) rather than masking it. |
| 3 | **Gate (Pitfall 3):** an engineered composition-only negative control reproduces a divergence where interp/`-O0` agree and `-flto` disagrees; an all-green matrix is a lane failure; the red cell is re-measured against a recorded toolchain | ✓ VERIFIED | Re-ran `TestCompositionOnlyLTODivergence` directly this session (`go test ./internal/compiler/native/... -run TestCompositionOnlyLTODivergence -v`): PASS, matrix `restrict+write` diverges to `10` at `-O3 -flto` only, all other cells match their row's `-O0` reference (`none`=104 everywhere, `restrict-only`=10 everywhere, `restrict+write`=104 at `-O0`/`-O1`/`-O3`). Independently re-ran `clang --version` on this host: `Apple clang version 21.0.0 (clang-2100.1.1.101)`, byte-identical to the value recorded in `11-NAT07-EVIDENCE.md`. The test's own `t.Fatal` on an all-green matrix (`d-11-21`) is a real fail-closed assertion, confirmed by reading the assertion logic, not merely narrated. |
| 4 | **Gate (Pitfall 5):** a call-graph-shape reachability register names reached/provably-unreached shapes; HDD reducer output on multi-function programs is re-verified to reproduce the same property as its input | ✓ VERIFIED | QLT-03 register (`qlt03_shape_register.json` + `qlt03_shape_register.go`) re-run this session: `TestQLT03Register`, `TestQLT03AuditCanFail`, `TestQLT03RegisterRejectsOutOfSetMechanism`, `TestQLT03RegisterRejectsNonexistentFalsifierTest`, `TestQLT03RegisterIsDeterministicUnderRepeatAndParallel`, etc. — all PASS; register rows are discriminated `reached`/`unreachable` with a named mechanism (`generator` or `refused`) and a named falsifier test for each unreachable row, mechanically checked, not hand-classified narrative. QLT-05's reducer re-verification re-run this session: `TestQLT05GateIsNonVacuous` and `TestQLT05EmptyReductionFailsAntiVacuity` both PASS — the anti-vacuity gate is proven non-decorative on both sides (a real reduction that must apply `drop-orphan-function`, and a seed on which no move is eligible that must NOT trip the same assertion). `reduce.go` genuinely widened to multi-function seeds (two new whole-program moves, `dropCallSite`/`dropOrphanFunction`, live per Q-01 BRANCH A — not narrowed behind `RefusedShapes()`), confirmed by reading the source. **Caveat (routed to human verification, not a FAIL):** `11-REVIEW.md` WR-01 — `reduce.Reduce` never validates `Seed.EntryFunctionID`, so a caller passing an empty/wrong value on a multi-function seed can cause `dropOrphanFunction` to silently delete the real entry point instead of refusing. Confirmed by reading `reduce.go:202-254`/`320-347`: no validation exists. The one production caller derives the ID correctly, so no current path is affected, but this is a real, disclosed robustness gap in the reducer's own exported contract that a human should explicitly accept or route to a follow-up plan. |
| 5 | **Gate (Pitfall 6):** no interprocedural fact is cacheable until a callee-changes-invalidates-caller regression gates it; interprocedural cache keys derive from call-graph closure, not per-unit hashes | ✓ VERIFIED | QLT-06 splits honestly into 06a (Phase 07, mutation-killed by two independent knowers — `TestCalleeChangeInvalidatesCallerClosureDigest` and its `corevalidate` peer) and 06b (Phase 11, complete-by-abstention — nothing interprocedural is cached because `ArtifactSpec.FixtureSource` already hashes the whole program, which strictly dominates any closure key in a single-translation-unit language). Re-ran this session: `TestDeclaredInputNamesStableAndNoInterproceduralImport`, `TestCacheDirectImportGuardCanFail`, `TestCacheTransitiveImportGuardCanFail`, `TestNoClosureDigestInCache`, `TestCacheKeyIsIdempotent` — all PASS, confirming both the abstention claim and its own falsifiability. The real D-11-41 stale-`cgen`-cache hole (found by Q-02) is genuinely closed: `cache.DeclaredInputNames()` now has 8 entries (`cgen_source` appended), confirmed the fix is live, not merely documented. |

**Score:** 5/5 truths verified (0 present-but-behavior-unverified)

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `internal/compiler/cgen/cgen_program.go` | Multi-function `emitProgram`/`emitProgramFunction`/shared `emitCall` | ✓ VERIFIED | Exists, substantive, wired — `emitCall` is called from 3 sites (`cgen_program.go:410`, `cgen.go:358`, `cgen.go:1867`), confirmed one shared helper per the mid-phase gate's own mandate, not five independent copies. |
| `internal/compiler/callgraph/callgraph.go` (`EntryFunction`) | Single-resolver entry-function determination | ✓ VERIFIED | Used at all 3 CLI run sites (`session.go` `Phase4CheckedProgram`/`RunInterpreter`/`RunNative`), confirmed by reading `11-GUARD-LEDGER.md` rows 6-8 and cross-checked against current `session.go`. |
| `internal/compiler/session/session_phase11_gate.go` | Mid-phase gate (`VerifyPhase11ZeroAttributeGate`) | ✓ VERIFIED | `TestPhase11ZeroAttributeGate`, `TestPhase11GateMutationKill` re-run, PASS. |
| `internal/compiler/session/session_phase11_differential_test.go` | Four-tier interprocedural differential (NAT-06) | ✓ VERIFIED | Re-run, PASS (see truth 2). |
| `internal/compiler/native/native_lto_test.go` (`TestCompositionOnlyLTODivergence`) | Engineered composition-only LTO control (NAT-07) | ✓ VERIFIED | Re-run independently, red exactly at `-O3 -flto`, matches recorded evidence byte-for-byte. |
| `internal/compiler/session/qlt03_shape_register.go` + `.json` | Call-graph-shape reachability register (QLT-03) | ✓ VERIFIED | Re-run, PASS; rows discriminated reached/unreachable with named mechanism + falsifier. |
| `internal/compiler/reduce/reduce.go` (multi-function `Seed`, `dropCallSite`/`dropOrphanFunction`) | Multi-function HDD reducer (QLT-05) | ⚠️ VERIFIED WITH DISCLOSED GAP | Substantive and wired for the one production caller; `Seed.EntryFunctionID` unvalidated (WR-01) — see human verification. |
| `internal/compiler/cache/probe.go` (`DeclaredInputNames`, 8 entries) | Eighth cache input closing the `cgen`-staleness hole (QLT-06) | ✓ VERIFIED | Confirmed 8 declared names via `TestDeclaredInputNamesStableAndNoInterproceduralImport`. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `session.RunInterpreter`/`RunNative` | `callgraph.EntryFunction` | Entry resolution replacing `len(Functions)!=1` | ✓ WIRED | Confirmed by ledger rows 6-8, re-verified count consistent. |
| `cgen.Emit`/`EmitNative` | `emitProgram` | N-function dispatch | ✓ WIRED | Confirmed by direct read of `cgen.go`. |
| `emitLinear`/`emitProgramFunction`/`emitBranchOperations` | `emitCall` | Shared call-lowering helper | ✓ WIRED (with WR-03 caveat: `emitBranchOperations`'s arm omits the sibling `declared`-map guard before calling it — documented unreachable-in-production, not blocking) | Confirmed by grep of all 3 call sites. |
| `reduce.Reduce`'s `movesToApply` | `dropCallSite`/`dropOrphanFunction` | Q-01 BRANCH A live whole-program moves | ✓ WIRED | Confirmed by reading `reduce.go`; `TestQLT05GateIsNonVacuous` proves `AppliedMoves` actually contains `"drop-orphan-function"` on a real reduction. |
| `phase6ArtifactSpec` | `CgenSourceDigest` | Eighth cache input | ✓ WIRED | Confirmed via `TestDeclaredInputNamesStableAndNoInterproceduralImport`. |

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|---|---|---|---|---|
| NAT-04 | 11-03, 11-04 | `cgen` emits multi-function C17 | ✓ SATISFIED | Truth 1 |
| NAT-05 | 11-04 | Alias promise names checked fact | ✓ SATISFIED (weakened by evidence, disclosed) | Truth 1, `11-VERIFICATION-INPUTS.md` §1 |
| NAT-06 | 11-05 | Four-tier interprocedural equivalence | ✓ SATISFIED (weakened by evidence on the `-flto` inertness claim, disclosed) | Truth 2, `11-VERIFICATION-INPUTS.md` §2 |
| NAT-07 | 11-02 | `-flto` tier proven non-inert by an engineered control | ✓ SATISFIED | Truth 3 |
| QLT-03 | 11-06 | Call-graph-shape reachability register | ✓ SATISFIED | Truth 4 |
| QLT-05 | 11-01, 11-08, 11-09 | HDD reducer re-verified on multi-function programs | ✓ SATISFIED (ships full, per Q-01 BRANCH A; WR-01 disclosed robustness gap on unvalidated caller input) | Truth 4 |
| QLT-06 | 11-01, 11-07 | Interprocedural cache-soundness gate | ✓ SATISFIED (split 06a Complete / 06b Complete-by-abstention, disclosed) | Truth 5 |

**No orphaned requirements.** Cross-referenced against `.planning/REQUIREMENTS.md`: all 7 assigned Phase 11 requirement IDs (NAT-04..07, QLT-03, QLT-05, QLT-06) appear in at least one plan's `requirements:` frontmatter, and REQUIREMENTS.md's own Phase-11 traceability table lists the identical 7 IDs with no additional row mapped to Phase 11 that a plan does not claim.

### Anti-Patterns Found

No `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER` markers found in any non-test `.go` file touched by this phase's diff (`cfd0ba22b^..HEAD`), confirmed by direct grep this session.

Two code-review findings from `11-REVIEW.md` remain open (both disclosed, neither blocking a current production path):

| File | Finding | Severity | Impact |
|---|---|---|---|
| `internal/compiler/reduce/reduce.go` | WR-01: `Seed.EntryFunctionID` unvalidated — can silently delete real entry point on multi-function seed | Warning | Routed to human verification above |
| `internal/compiler/reduce/reduce.go` | WR-02: `entryFunctionID` package-level var, unsynchronized, doc-commented as non-reentrant but not enforced | Warning | Latent concurrency hazard, no current concurrent caller |
| `internal/compiler/cgen/cgen.go` | WR-03: `emitBranchOperations`'s `OpCall` arm omits the sibling double-target-write guard | Warning | Documented unreachable in production today |
| `internal/compiler/reduce/reduce.go` | WR-04: `removeIndices`/`scrubBlocks` dead code | Warning | Cosmetic |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Guard-count re-derivation matches ledger's claimed trajectory | `awk '/^func /{f=$0} /Functions\) != 1/{print FILENAME": "f}' $(find internal cmd -name '*.go' -not -name '*_test.go')` | 26 (down from 32 pre-phase / 29 at plan 11-05) | ✓ PASS |
| Composition-only LTO control is genuinely red on `-flto` only | `go test ./internal/compiler/native/... -run TestCompositionOnlyLTODivergence -v` | PASS, single diverging cell `restrict+write`/`-O3 -flto` | ✓ PASS |
| Toolchain identity matches recorded evidence | `clang --version` | `Apple clang version 21.0.0 (clang-2100.1.1.101)` — matches `11-NAT07-EVIDENCE.md` | ✓ PASS |
| Four-tier interprocedural differential | `go test ./internal/compiler/session/... -run TestPhase11InterproceduralDifferential -v` | PASS (5 fixtures + subtests), 1 honest skip | ✓ PASS |
| Mid-phase gate mutation-kill | `go test ./internal/compiler/session/... -run TestPhase11GateMutationKill -v` | PASS | ✓ PASS |
| QLT-03 register mechanical checks | `go test ./internal/compiler/session/... -run TestQLT03 -v` | PASS (all subtests) | ✓ PASS |
| QLT-05 anti-vacuity gate | `go test ./internal/compiler/session/... -run 'TestQLT05GateIsNonVacuous\|TestQLT05EmptyReductionFailsAntiVacuity' -v` | PASS both sides | ✓ PASS |
| QLT-06 abstention falsifiability | `go test ./internal/compiler/cache/... -run 'TestDeclaredInputNamesStableAndNoInterproceduralImport\|TestCacheDirectImportGuardCanFail\|TestCacheTransitiveImportGuardCanFail\|TestNoClosureDigestInCache\|TestCacheKeyIsIdempotent' -v` | PASS all 5 | ✓ PASS |
| Reduce EntryFunctionID validation (WR-01) | Read `reduce.go:202-254`, `320-347` | No validation present | ✗ CONFIRMS FINDING (routed to human verification, not a Phase 11 FAIL) |

Full `go build ./...` / `go test ./...` state independently verified by the orchestrator prior to this pass (exit 0, 23 packages ok, 0 FAIL) — not re-run in full here per spot-check discipline; targeted named tests above were re-run individually instead.

### Human Verification Required

See frontmatter `human_verification` block. Two items, both already surfaced by the phase's own artifacts (11-REVIEW.md WR-01, 11-MIDPHASE-GATE.md's disclosed CLI-check divergence) rather than newly discovered gaps — routed here because they are genuine judgment calls about acceptable residual risk, not missing work.

### Gaps Summary

No must-have truth failed. All 5 ROADMAP success criteria for Phase 11 are backed by re-run, passing tests and directly-read source, not merely SUMMARY.md narrative. The phase is unusually self-auditing: 4 of 7 requirements ship with an explicitly disclosed "weakened by evidence" framing (NAT-05's empty attribute set, NAT-06/criterion-2's LTO inertness-by-construction, QLT-06's complete-by-abstention split, QLT-05's one disclosed pre-existing debt item alongside a full, unnarrowed reducer), and two genuinely new debt items were discovered honestly during the phase itself (D-11-51 shared-leaf event-identity collision, D-11-52 diverging-callee inexpressibility) with tests that assert consistent, honest refusal rather than masking the gap.

The reason this report is `human_needed` rather than `passed` is a single independently-discovered code-review finding (WR-01) that is real, reproducible by construction, and not yet fixed or explicitly accepted by a human: `reduce.Reduce` trusts `Seed.EntryFunctionID` without validating it, so a future or misused caller (not today's one production caller) could silently delete the multi-function program's real entry point instead of refusing. This does not fail any of the phase's own success criteria as measured against its one production caller, but it is exactly the kind of "ambiguity resolved by guessing instead of refusing" pattern this codebase's own stated philosophy elsewhere refuses to accept, so it is surfaced for an explicit human decision rather than silently absorbed into a passing score.

---

_Verified: 2026-09-12_
_Verifier: Claude (gsd-verifier)_
