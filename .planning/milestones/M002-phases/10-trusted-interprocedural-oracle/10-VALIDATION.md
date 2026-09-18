---
phase: "10"
slug: "trusted-interprocedural-oracle"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-11"
validated: "2026-09-11"
evidence_vocabulary: v1
graded_rows: 10
---

# Phase 10 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Derived from `10-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package (no third-party test framework in the repo) |
| **Config file** | none — plain `go test` |
| **Quick run command** | `go test ./internal/compiler/interp/... ./internal/compiler/originvalidate/... ./internal/compiler/pathoracle/... ./internal/compiler/corevalidate/...` |
| **Full suite command** | `go test ./...` (phase gate also runs `go test -race ./...`, matching `ci.yml:64-65`) |
| **Estimated runtime** | ~30 seconds quick / ~120 seconds full |

---

## Sampling Rate

- **After every task commit:** Run the targeted package test — `go test ./internal/compiler/<pkg>/...`
- **After every plan wave:** Run `go test ./...` (full suite, matching CI's `checks` job)
- **Before `/gsd-verify-work`:** `go test ./...` and `go test -race ./...` both green
- **Max feedback latency:** 120 seconds

---

## Per-Task Verification Map

Seeded from the RESEARCH requirement→test map. Task IDs are filled in by
`/gsd-validate-phase` once PLAN.md task numbering is final; every row below MUST
bind to at least one plan task.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 10-04 Task 2 | 10-04 | 2 | SEM-08 | — | Depth-exceeded refusal fires through the real pipeline on a genuine 129-function chain | integration | `go test ./internal/compiler/interp/... -run TestCallDepthExceeded -v` | ✅ | EXERCISED | — |
| 10-04 Task 3 | 10-04 | 2 | SEM-08 (Pitfall 4 gate) | — | Native-stack headroom probe is a distinct limit from the language-level call-depth bound | subprocess integration | `go test ./internal/compiler/interp/... -run TestNativeStackHeadroomIndependentOfCallDepth -v` | ✅ | EXERCISED | — |
| 10-05 Task 1 | 10-05 | 3 | SEM-09 | — | Drop/cleanup order observed via canonical bytes on normal return **and** every nonlocal exit | unit + differential | `go test ./internal/compiler/interp/... -run TestFrameDrainOrder -v` | ✅ | EXERCISED | — |
| 10-02 Task 2 | 10-02 | 2 | TRU-02 | — | `walkReturnOrigin` handles `case core.OpCall`; `twin_a_accept.lang` admits clean through the full CLI | unit + CLI gate | `go test ./internal/compiler/originvalidate/... -run TestOpCallOriginWalk -v` | ✅ | EXERCISED | — |
| 10-02 Task 3 | 10-02 | 2 | TRU-02 (import guard) | T-10-01 | `originvalidate` transitively imports neither `check` nor `corevalidate` | build/test guard | `go test ./internal/compiler/originvalidate/... -run 'Imports.*Independent' -v` | ✅ | EXERCISED | — |
| 10-03 Task 2 | 10-03 | 2 | TRU-03 | — | Composition splits the caller's endpoints per-path depending on which callee path is spliced | unit | `go test ./internal/compiler/pathoracle/... -run TestCompositionDiscriminatesPerPathBorrow -v` | ✅ | EXERCISED | — |
| 10-07 Task 3 | 10-07 | 5 | QLT-04 | — | Depth-3 composition corpus reaches the **declared** bound, checked bidirectionally | corpus + bidirectional gate | `go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -v` | ✅ | EXERCISED | — |
| 10-01 Task 3 | 10-01 | 1 | OWN-05b | T-10-02 | `interp` never reads `corevalidate.Result`'s ownership-bearing fields | guard | `go test ./internal/compiler/interp/... -run TestInterpDoesNotReadCorevalidateOwnershipFields -v` | ✅ | EXERCISED | — |
| 10-08 Task 2 | 10-08 | 6 | Criterion 4 | — | Three-way-on-refuse / four-way-on-accept differential with seeded faults per peer pair | differential + mutation-kill | `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v` | ✅ | EXERCISED | — |
| 10-09 Task 1/2 | 10-09 | 7 | Stability freeze | — | Golden corpus byte-for-byte and deterministic across repeated runs | golden + flake check | `go test ./internal/compiler/interp/... -run TestInterpOracleGoldenCorpus -v` and `-run 'TestInterpDeterministicAcrossRuns\|TestInterpOracleGoldenCorpus\|TestInterpOracleCorpusDeterministic' -count=10` | ✅ | EXERCISED | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

All ten rows bind to a real plan task with a real automated command, verified
green as of this plan's own commit (2026-09-11) — the condition
`nyquist_compliant: true` requires per this file's own frontmatter guidance.

---

## Wave 0 Requirements

- [x] `internal/compiler/interp/interp_test.go` — depth-exceeded through the real pipeline (D-10-24), drop-order seam (D-10-35), native-stack probe (D-10-42), OWN-05b mutant-kill pairs (D-10-41)
- [x] `internal/compiler/originvalidate/originvalidate_test.go` — `OpCall` widening test, `corevalidate` added to the forbidden-import list, transitive-scan upgrade, `TerminatorKindsOverride`-shaped seam (D-10-08)
- [x] `internal/compiler/pathoracle/pathoracle_test.go` — composition test, discriminating per-path-borrow fixture (D-10-14, required deliverable), transitive-guard upgrade
- [x] `testdata/phase10/` — depth-3 fixture pair varying which hop carries the borrow (D-10-48), the 129-function depth-exceeded fixture (or generator), `interp_oracle/` golden corpus
- [x] `internal/compiler/session/session_peer_gate_test.go` — extended to three/four peers; `peerDivergenceExpected` upgraded to an accountable struct carrying debt-register ID + landing phase (D-10-54)
- [x] Framework install: **none** — `testing` stdlib only

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 120s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** ratified at plan 10-09 (2026-09-11) — the Per-Task Verification
Map is fully bound, the structural coverage floor below is populated with no
blank cells, and `go test ./...` / `go test -race ./...` both exit 0.

---

## Structural Coverage Floor (D-10-58)

Per D-10-58: the floor is an ENUMERATED TABLE, never a percentage. Every
`core.OperationKind` is exercised through `Run` at each of the three
execution paths (`runBranchArm`, `runLinearBlocks`, `runLinear`) — or the
cell states `PROVABLY UNREACHABLE` with a structural reason, never left
blank — every refusal path is reached by a named test, and Mutation-Kill
Register rows exist for the call-stack, cross-frame, and ownership machinery
this phase and its predecessors added.

### 1. Operation-kind coverage matrix

Rows: every member of `core.AllOperationKinds()` (`core.go:675`). Columns:
`runBranchArm`, `runLinearBlocks`, `runLinear` — interp's three execution
paths (`interp.go`). All three paths share one operation-execution driver
(`runFrameStack`), so a kind's handling is IDENTICAL regardless of which
path's frame reached it; what differs across columns is whether a real (or,
where noted, a deliberately synthetic) program can construct a body of that
SHAPE containing that kind at all.

| Operation kind | `runBranchArm` | `runLinearBlocks` | `runLinear` |
|---|---|---|---|
| `core.OpCopy` | `interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus/single_frame_return` and `/defect_terminal` (real fixture, `defect_terminal.lang`'s "Go" arm) | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/block` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/flat`; also `interp_test.go#TestMoveAsCopyMutationKilled`'s mutated-arm sub-run |
| `core.OpMove` | `interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus/single_frame_return` (`defect_terminal.lang`'s `take flag`) | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/block` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/flat` |
| `core.OpBorrowShared` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/arm` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/block` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/flat` |
| `core.OpBorrowExclusive` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/arm` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/block` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/flat` |
| `core.OpReturn` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/arm`; `TestInterpOracleGoldenCorpus/single_frame_return` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/block`; `TestFrameDrainOrder/normal_return_across_one_boundary` | `interp_test.go#TestOperationKindCoverageAcrossAllThreePaths/flat`; `TestCallExecutesAcrossOneFrame` |
| `core.OpForeignCall` | `PROVABLY UNREACHABLE` — `checkFallibleLinear` (the sole producer of a fallible-call block) only ever runs for a Match-LESS function (`interp.go`'s own `newBlockFrame`/`runLinear` doc comments: "Every pre-Phase-4 linear (non-Match) function leaves Blocks empty -- only checkFallibleLinear ever populates it for a Match-less function"); a match-arm body can therefore never contain an `OpForeignCall` | `interp_test.go#TestFrameDrainOrder` (every subtest); `TestInterpOracleGoldenCorpus/cross_frame_nonlocal_landing_pad` | `PROVABLY UNREACHABLE` — `runLinear` dispatches to `runLinearBlocks` the instant `len(function.Linear.Blocks) > 0` (`interp.go:283-285`), and `OpForeignCall` cannot exist in a flat (`Blocks == nil`) body by the same `checkFallibleLinear` construction rule |
| `core.OpFail` | `PROVABLY UNREACHABLE` — same reasoning as `OpForeignCall`: `OpFail` is only ever emitted paired with a fallible call's err edge (`check.go:3080`/`3307`), which requires a block-based body a match arm can never have | `PROVABLY UNREACHABLE via the real pipeline` — `runLinearBlocks` (via `runFrameStack`'s `core.OpForeignCall` arm) unconditionally follows the OK edge and never visits the err block an `OpFail` lives in (this phase's own documented scope limit); the shape-agnostic switch dispatch proving `OpFail` executes correctly WHEN reached is already covered by the flat-shape probe (next column), never re-derived a second way here | `interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus/typed_failure` — a deliberately SYNTHETIC `core.Program` (`oracleTypedFailureProgram`), since no real check.go-emitted body can place `OpFail` in a flat body at all; this is the only way to observe the terminal shape, mirroring `moveAsCopyProbeProgram`'s established precedent |
| `core.OpRelease` | `PROVABLY UNREACHABLE` — `OpRelease` discharges an `OpForeignCall` acquisition (`core.go:546-551`); since `OpForeignCall` itself is provably unreachable in a match-arm body (row above), so is its paired release | `interp_test.go#TestFrameDrainOrder` (every subtest emitting a `resource.released`/`resource.leaked` event); `TestInterpOracleGoldenCorpus/cross_frame_nonlocal_landing_pad` | `PROVABLY UNREACHABLE` — same reasoning as `OpForeignCall`'s own `runLinear` cell: a flat body can never carry the block-based acquisition `OpRelease` discharges |
| `core.OpDefect` | `interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus/defect_terminal` (real fixture, `defect_terminal.lang`'s "Halt" arm — the only real construct producing `OpDefect`, D-04-15: "admissible only in a match arm's terminal position") | `PROVABLY UNREACHABLE` — `OpDefect` is match-arm-terminal-only by construction (D-04-15); a block-based non-match body never contains one | `PROVABLY UNREACHABLE` — same D-04-15 reasoning; a flat (Match-less) body never contains an `OpDefect` |
| `core.OpCall` | `interp_test.go#TestCallFromBothMatchArmsAcrossFrames` (real fixture, `call_from_both_match_arms.lang`) | `interp_test.go#TestFrameDrainOrder` (the drain builders' own `OpCall`); `TestInterpOracleGoldenCorpus/cross_frame_nonlocal_landing_pad` | `interp_test.go#TestCallExecutesAcrossOneFrame`; `TestMoveAsCopyMutationKilled`; `TestCallDepthExceeded` |

Every `PROVABLY UNREACHABLE` cell above cites a structural rule already
stated in this codebase's own doc comments or `core.go`'s own operation-kind
documentation (D-04-15, the `checkFallibleLinear` scoping rule, and
`runLinearBlocks`' own "always follows the ok edge" scope limit) — none is
a bare assertion. `TestOperationKindCoverageAcrossAllThreePaths` and
`TestCallFromBothMatchArmsAcrossFrames` were added by this plan's Task 3 to
close the real (not merely structural) gaps the matrix surfaced: before this
plan, `core.OpBorrowShared`/`core.OpBorrowExclusive` were untested at every
path, `core.OpCopy`/`core.OpMove` were each tested at only one of the three
reachable paths, and `core.OpCall` had never been driven through
`runBranchArm` at all.

### 2. Refusal-path table

Rows: the six refusals D-10-58's own action text names as the required
minimum. `interp.go` also contains several lower-level internal-consistency
guards (e.g. "block or body references unknown operation", "operation reads
uninitialized place", "checked match has no arm for input") that exist as
defense-in-depth against a malformed `core.Program`; those are out of this
table's declared scope (recorded here, not silently omitted) and remain open
follow-up if a future phase wants them enumerated too.

| Refusal | Test that reaches it | Notes |
|---|---|---|
| Depth-exceeded refusal (SEM-08) | `interp_test.go#TestCallDepthExceeded`; `TestCallDepthAtAndOverTheCap` | Named `Outcome`/`Event` data (`callDepthExceededDefectReason`), never a bare Go error |
| Foreign nonlocal-exit defect | `interp_test.go#TestFrameDrainOrder/foreign_nonlocal_landing_pad_across_multiple_frames`; `interp_oracle_golden_test.go#TestInterpOracleGoldenCorpus/cross_frame_nonlocal_landing_pad` | Named `Outcome`/`Event` data (`nonlocalExitDefectReason`) |
| Invalid body union | `interp_test.go#TestRunRefusesInvalidBodyUnion` | Documents a genuine finding: `Run`'s own `!function.HasClosedBody()` guard is `PROVABLY UNREACHABLE` via `Run()` in practice, because `corevalidate.Validate` (Run's unconditional first act) already refuses this exact shape as `core.invalid_body` before the guard is ever reached — confirmed directly, not assumed. The guard is defense-in-depth redundancy, not a live path |
| Absent function | `interp_test.go#TestRunRefusesAbsentFunction` | Plain Go error naming the missing function name |
| Unknown operation kind | `interp_test.go#TestRunFrameStackRefusesUnknownOperationKind` | Synthetic `core.LinearOperation` with a bogus `Kind` string — no real check.go emission path can produce one, since check.go only ever emits from `core.AllOperationKinds()`'s closed set |
| Corevalidate precondition failure | `interp_test.go#TestRunRefusesCorevalidateInvalidProgram` | An empty `core.Program{}` (no declared Schema) refused as `core.schema` before `Run` attempts to find or execute any function |

### 3. Mutation-Kill Register

Six seams this phase (and its own plan 10-02/10-08 predecessors, cross-
referenced here per this plan's own scope) added for the new call-stack,
cross-frame, and ownership machinery:

| Mutant | Seam | Killing test | Peer that alone catches it |
|---|---|---|---|
| Move-as-copy (skip the caller-side delete on a call argument) | `interp.go`'s `moveAsCopyForTest` (plan 10-01, D-10-41) | `interp_test.go#TestMoveAsCopyMutationKilled` | `interp` — `check`/`corevalidate` execute nothing, so they are structurally blind to a runtime move-vs-copy violation (D-10-36) |
| Frame-drain order flip (reverse the LIFO stack-traversal order) | `interp.go`'s `frameDrainOrderForTest` (plan 10-05, D-10-35) | `interp_test.go#TestFrameDrainOrder/frame_drain_order_seam_flips_canonical_bytes` | `interp` — proves the drain order is genuinely OBSERVED through `interp.CanonicalBytes`/`execution.Equal`, never merely computed and discarded |
| Call-depth cap disable | `interp.go`'s `maxCallDepthOverride` (plan 10-04, Pitfall-4) | `interp_test.go#TestNativeStackHeadroomIndependentOfCallDepth` (`cap_disabled` arm) | `interp` — with the cap disabled, an 800-function chain completes under a pinned 1 MiB host stack with no collapse, proving the language-level depth bound and the native-stack limit are structurally UNRELATED, not one limit wearing two names |
| Pathoracle seeded contract hop | `pathoracle_test.go`'s `forceContractHopForTest` (plan 10-03/10-08, D-10-14/D-10-55) | `pathoracle_test.go#TestCompositionDiscriminatesPerPathBorrow` | `pathoracle` — its own per-path composition detects the fault (and, per plan 10-08's own extension, `corevalidate.Result.LoanEndpoints()` stays byte-for-byte UNCHANGED under the same fault, proving `corevalidate` structurally blind to a per-path-only divergence) |
| `derivePeerSignature` parameter-mode mutant | `corevalidate.go`'s `parameterContractModeOverrideForTest` (plan 10-08, D-10-41 complement) | `corevalidate_endpoint_test.go#TestDerivePeerSignatureModeMutantPairing` | `corevalidate`/`originvalidate`'s declared-contract differential — while `interp`'s own canonical bytes stay unchanged, since interp never consults a `Mode` string at all (the complementary half of the move-as-copy pairing above: one peer catches what the other two structurally cannot, in each direction) |
| `originvalidate` `OpCall`-case no-op seam | `originvalidate.go`'s `disableOpCallOriginConsultForTest` (plan 10-02, D-10-08) | `originvalidate_test.go#TestOpCallOriginWalkGateIsLoadBearing` | `originvalidate`'s own `walkReturnOrigin` re-derivation — engaging the seam regresses `twin_a_accept.lang` from a clean admit to its pre-fix `core.origin_omitted` refusal through the FULL `session.CheckCommandFile` pipeline, proving the `OpCall` case DECIDES the outcome rather than merely existing syntactically |

No blank cells remain in any of the three tables above.
