---
phase: 04-fallible-resources-and-c-boundary
plan: "08"
subsystem: compiler
tags: [corevalidate, gap-closure, mutation-kill, resource-lifecycle, terminal-block]

requires:
  - phase: 04-fallible-resources-and-c-boundary
    provides: "commit 16fb0c9's per-incoming-edge release-order rederivation in checkReleaseOrder (04-02/04-07 lineage), which this plan falsifies rather than re-implements"
provides:
  - "A falsifier for the merge-terminal-block release-order rederivation: four new tests proving the rederivation refuses a hand-corrupted divergent merge, accepts a legitimate agreeing merge, refuses an unreachable terminal block, and still accepts discard_because.lang's own legitimate merge"
  - "A new structural peer check, core.terminal_block_unreachable, in blocksAndEdges: every OpReturn/OpFail-terminated non-entry block must have at least one incoming edge, independent of checkReleaseOrder"
  - "Two new Mutation-Kill Register rows in 04-VALIDATION.md, both demonstrated with verbatim revert-and-fail output"
affects: [04-VERIFICATION.md gap 1, future corevalidate work touching terminal-block or release-order refusals]

actuals:
  tokens: 9800
  tasks: 3
  commits: 2

tech-stack:
  added: []
  patterns:
    - "Terminal-block entry identification via block.PointID == function.EntryPointID (not Blocks[0] position), matching check.go's own producer convention"
    - "Structural peer checks live beside the check they support (blocksAndEdges) rather than inside the rule they are independent of (checkReleaseOrder), so 'independent' is a line-number-verifiable fact, not a claim"

key-files:
  created: []
  modified:
    - internal/compiler/corevalidate/corevalidate.go
    - internal/compiler/corevalidate/corevalidate_test.go
    - .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md

key-decisions:
  - "Task 1 decision (auto-selected under auto-mode, gate=blocking default, recommended option first): declined 04-VERIFICATION.md's literal 'every terminal block has exactly one incoming edge' structural check (option B) and its stated-carve-out variant (option C), and shipped option A instead -- a weaker but sound peer check (core.terminal_block_unreachable: at-least-one incoming edge, non-entry) that keeps discard_because.lang's legitimate ok/err merge valid while still closing the independence gap 04-VERIFICATION.md identified. Recorded here as the human-decision artifact the plan requires; see the auto-selection note below."
  - "The rederivation's independence (D-04-07/D-12a) is now proven, not assumed: TestMergeTerminalBlockDivergentReleaseSetsRefused constructs the exact merge-terminal-block shape checkReleaseOrder used to skip before 16fb0c9, and the mutation-kill revert demonstrates the fix is load-bearing (Valid:true when reverted)."
  - "Entry-block identification for the new structural check uses block.PointID == function.EntryPointID, not linear.Blocks[0], per the read_first instruction to confirm rather than guess; check.go's own producers (checkBranch, checkForeignTracer, checkResourceLifecycle) already rely on this exact convention for the entry block's PointID."

patterns-established:
  - "A gap-closure plan's falsifier tests are written against the ALREADY-SHIPPED fix commit, then proven load-bearing via a throwaway detached-worktree revert of that exact commit's hunk -- never a fresh implementation the plan re-derives from scratch."

requirements-completed: [RES-01, SEM-03]

coverage:
  - id: D1
    description: "corevalidate's release-order rederivation is proven independent (not merely claimed) via a falsifier that constructs a merge terminal block with divergent completed-acquisition sets and is refused with core.release_order_mismatch"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestMergeTerminalBlockDivergentReleaseSetsRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestMergeTerminalBlockAgreeingChainsAccepted"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestLegitimateDiscardMergeStillValidates"
        status: pass
    human_judgment: false
  - id: D2
    description: "A new structural peer check (core.terminal_block_unreachable) refuses an unreachable non-entry terminal block, independent of checkReleaseOrder, without narrowing the language (discard_because.lang stays valid)"
    requirement: "RES-01"
    verification:
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestTerminalBlockUnreachableRefused"
        status: pass
      - kind: unit
        ref: "internal/compiler/corevalidate/corevalidate_test.go#TestTerminalBlockWithNoIncomingEdgeRefused"
        status: pass
    human_judgment: false
  - id: D3
    description: "Both production hunks are mutation-killed: reverting each in a throwaway detached worktree turns its named test red, with verbatim failing output recorded in this summary"
    verification:
      - kind: other
        ref: "manual git-worktree revert-and-run of 16fb0c9 and the blocksAndEdges hunk (see Mutation-Kill Demonstrations below)"
        status: pass
    human_judgment: false
  - id: D4
    description: "The whole Phase 4 gate stays green after the structural peer check is added"
    verification:
      - kind: other
        ref: "sh scripts/verify-phase4.sh (exit 0, all lanes pass, all required control IDs present)"
        status: pass
      - kind: other
        ref: "go test ./... and go vet ./... (both clean)"
        status: pass
    human_judgment: false

duration: 30min
completed: 2026-09-05
status: complete
---

# Phase 4 Plan 08: Gap-1 Falsifier and Terminal-Block-Reachability Peer Check Summary

**Four new corevalidate tests prove the merge-terminal-block release-order rederivation is genuinely independent (not vacuously passing), and a new `core.terminal_block_unreachable` structural peer check in `blocksAndEdges` closes 04-VERIFICATION.md gap 1's second missing item without narrowing the language.**

## Performance

- **Tasks:** 3 (1 checkpoint:decision, 1 tracer+tdd, 1 auto+tdd)
- **Files modified:** 3
- **Commits:** 2 (task 1 produced no code change — decision recorded here)

## Accomplishments

- **Task 1 (checkpoint:decision, auto-resolved):** Auto-mode was active (`workflow.auto_advance: true` in `.planning/config.json`, and the task carried the default `gate="blocking"`, not `gate="blocking-human"`), so per the executor's Rule 5 the first/recommended option was auto-selected and logged: `⚡ Auto-selected: A (decline "exactly one"; ship core.terminal_block_unreachable at-least-one/non-entry peer check)`. This is recorded here as the plan's required human-decision artifact under the label 04-VERIFICATION.md gap 1: 04-VERIFICATION.md's literal "every terminal block has exactly one incoming edge" structural check (option B) was declined — it was tried during the 04-REVIEW-FIX.md CR-01 pass and broke `discard_because.lang`'s legitimate ok/err merge (`TestValidatorRederivesReleaseOrder`, `TestReleaseOrderValidationWorkSeries` both went red). Option C's explicit carve-out was also declined — it restates the same carve-out 04-VERIFICATION.md objected to, just spelled out. Option A ships instead: independence from `check` is proven by Task 2's falsifier on the per-incoming-edge rederivation (not by the structural check), and structural well-formedness is proven by Task 3's new peer check (at-least-one incoming edge, non-entry) — weaker than "exactly one" but sound against `discard_because.lang`.
- **Task 2:** Added four tests to `internal/compiler/corevalidate/corevalidate_test.go` — `TestMergeTerminalBlockDivergentReleaseSetsRefused` (the falsifier: a hand-corrupted merge with divergent completed-acquisition sets is refused with `core.release_order_mismatch`), `TestMergeTerminalBlockAgreeingChainsAccepted` (sibling proof a second incoming edge isn't refused merely for existing), `TestTerminalBlockWithNoIncomingEdgeRefused` (accepts either refusal code so it's green before/after Task 3), and `TestLegitimateDiscardMergeStillValidates` (regression guard). All four passed on first implementation. Mutation-kill demonstration: reverting commit `16fb0c9`'s `checkReleaseOrder` hunk in a throwaway detached worktree turned `TestMergeTerminalBlockDivergentReleaseSetsRefused` red — the corrupted merge is silently accepted (`Valid:true`) once the per-incoming-edge rederivation is removed. Full output recorded below.
- **Task 3:** Extended `blocksAndEdges`'s existing terminal-block loop (previously `OpFail`-only) to also inspect `OpReturn`-terminated blocks, adding one refusal: `core.terminal_block_unreachable` for any terminal block that is not the function's entry block (identified via `block.PointID == function.EntryPointID`, confirmed by reading check.go's producers rather than assumed) and has zero incoming edges. No upper bound on incoming-edge count is imposed anywhere (that would be the declined "exactly one" form). Added `TestTerminalBlockUnreachableRefused`; mutation-kill demonstration confirms deleting the new refusal turns the test red (falls back to `core.release_order_indeterminate` instead, proving the specific new check — not some other check — is what the test depends on). Appended two rows to `04-VALIDATION.md`'s Mutation-Kill Register, both owning plan `04-08`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Decide whether to decline the literal "exactly one incoming edge" check** — no commit (decision-only checkpoint; auto-resolved to option A, recorded above and in this file)
2. **Task 2: The merge-terminal-block falsifier** — `8535bfd` (test)
3. **Task 3: The structural peer check in blocksAndEdges** — `3e8a4db` (feat)

_Note: Task 2 is `type="tracer" tdd="true"` but was executed and committed as a single test-only commit (adding falsifier tests against an already-shipped fix, not a fresh RED/GREEN/REFACTOR cycle against new production code); Task 3 similarly ships production code and its test in one `feat` commit since the new check and its test were designed together against the read-first material._

## Files Created/Modified

- `internal/compiler/corevalidate/corevalidate.go` — Adds the `core.terminal_block_unreachable` structural peer check to `blocksAndEdges` (Task 3); `checkReleaseOrder` itself was NOT modified by this plan (its independent rederivation, from commit `16fb0c9`, was already correct — this plan only proves it).
- `internal/compiler/corevalidate/corevalidate_test.go` — Adds five new tests: `TestMergeTerminalBlockDivergentReleaseSetsRefused`, `TestMergeTerminalBlockAgreeingChainsAccepted`, `TestTerminalBlockWithNoIncomingEdgeRefused`, `TestLegitimateDiscardMergeStillValidates` (Task 2), `TestTerminalBlockUnreachableRefused` (Task 3).
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` — Two new Mutation-Kill Register rows, owning plan `04-08`, both `✅ green`.

## Decisions Made

See Accomplishments/Task 1 above for the full disposition record. In short: option A (decline "exactly one", ship a weaker at-least-one/non-entry peer check) was selected over option B (ship "exactly one" and break `discard_because.lang`) and option C (ship "exactly one" with an explicit carve-out, restating rather than removing the objection).

## Mutation-Kill Demonstrations

### Revert 1: commit `16fb0c9`'s per-incoming-edge rederivation in `checkReleaseOrder`

Reverted in a throwaway detached worktree (`git worktree add --detach`, `git show 16fb0c9 -- internal/compiler/corevalidate/corevalidate.go | git apply -R`), then ran `TestMergeTerminalBlockDivergentReleaseSetsRefused`:

```
=== RUN   TestMergeTerminalBlockDivergentReleaseSetsRefused
    corevalidate_test.go:769: expected core.release_order_mismatch, got {Valid:true Problems:[] Checks:388 program:{Schema:lang.core/1 Module:phase4.acquire_three_success ModuleID:s1:phase4.acquire_three_success:module:phase4.acquire_three_success DataTypes:[{ID:s1:phase4.acquire_three_success:type:AcquireError Name:AcquireError Alternatives:[OpenFailed] Span:{Start:647 End:681}}] Functions:[{ID:s1:phase4.acquire_three_success:fn:main Name:main EntryPointID:s1:phase4.acquire_three_success:fn:main:point:entry ReturnPointID:s1:phase4.acquire_three_success:fn:main:point:return Parameter:{ID:s1:phase4.acquire_three_success:fn:main:place:0 Name:request Type:Byte} ReturnType:Byte Match:<nil> Linear:0x140001b66e0 PublicOrigin:<nil> ForeignContract:0x1400013a840 Span:{Start:683 End:838}}]}}
--- FAIL: TestMergeTerminalBlockDivergentReleaseSetsRefused (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/corevalidate	0.192s
FAIL
```

The corrupted merge silently validates (`Valid:true`) once the fix is reverted — proof the falsifier is load-bearing. The worktree was deleted immediately after; no revert reached `main`.

### Revert 2: the new `core.terminal_block_unreachable` refusal in `blocksAndEdges`

Reverted in a second throwaway detached worktree, then ran `TestTerminalBlockUnreachableRefused`:

```
=== RUN   TestTerminalBlockUnreachableRefused
    corevalidate_test.go:906: expected core.terminal_block_unreachable, got {Valid:false Problems:[{Code:core.release_order_indeterminate Detail:s1:phase4.acquire_three_success:fn:main:block:success}] Checks:359 program:{Schema:lang.core/1 Module:phase4.acquire_three_success ModuleID:s1:phase4.acquire_three_success:module:phase4.acquire_three_success DataTypes:[{ID:s1:phase4.acquire_three_success:type:AcquireError Name:AcquireError Alternatives:[OpenFailed] Span:{Start:647 End:681}}] Functions:[{ID:s1:phase4.acquire_three_success:fn:main Name:main EntryPointID:s1:phase4.acquire_three_success:fn:main:point:entry ReturnPointID:s1:phase4.acquire_three_success:fn:main:point:return Parameter:{ID:s1:phase4.acquire_three_success:fn:main:place:0 Name:request Type:Byte} ReturnType:Byte Match:<nil> Linear:0x14000216820 PublicOrigin:<nil> ForeignContract:0x1400012e840 Span:{Start:683 End:838}}]}}
--- FAIL: TestTerminalBlockUnreachableRefused (0.00s)
FAIL
FAIL	github.com/codename-lang/lang/internal/compiler/corevalidate	0.244s
FAIL
```

Without the new check, the artifact is still refused (`Valid:false`) but by a different, non-structural code (`core.release_order_indeterminate`, checkReleaseOrder's own pre-existing guard) — proving the test's exact-code assertion depends specifically on the new peer check, not on some other refusal that happens to also fire. The worktree was deleted immediately after; no revert reached `main`.

## Deviations from Plan

None - plan executed exactly as written. Task 1's checkpoint:decision was resolved via the executor's standard auto-mode Rule 5 (auto-select the first/recommended option under `gate="blocking"`), which is normal auto-mode flow per the checkpoint protocol, not a deviation.

## Issues Encountered

None. All four falsifier tests (Task 2) and the new structural-check test (Task 3) passed on first implementation; both mutation-kill reverts produced the expected red output on first attempt.

## Verification Results

- `env GOCACHE=/tmp/ai-lang-phase4-cache go test ./...` — exits 0, all packages pass
- `env GOCACHE=/tmp/ai-lang-phase4-cache go vet ./...` — exits 0, clean
- `sh scripts/verify-phase4.sh` — exits 0; all lanes `"status":"pass"`; all required control IDs present including `control:kind.exhaustive_dispatch`, `control:resource.release_order_transposed`, `control:resource.release_omitted`, `control:foreign.no_unproven_attributes`, `control:foreign.unwind_forbidden`, `control:defect.no_release_on_defect`, `control:defect.signal_adjudicated`, `control:foreign.nonlocal_exit_undetected`, `control:terminator.walk_incomplete`, `control:origin.foreign_origin_omitted`
- `git diff --stat .planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` — 2 insertions only, no deletions of existing rows

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- 04-VERIFICATION.md gap 1 is closed: item 1 ("treat an unexpected incoming-edge count as a hard refusal, not a skip") was already implemented by commit `16fb0c9` and is now falsified by this plan's tests; item 2 ("add a structural peer check independent of checkReleaseOrder") is implemented in the adapted form Task 1's recorded decision selects.
- Re-running `/gsd-verify-work` on Phase 04 should flip the 04-02 must_have "corevalidate independently rederives the expected release order... and compares" from `failed` to verified, on the evidence of `TestMergeTerminalBlockDivergentReleaseSetsRefused` (a test that constructs the previously-untested shape) and its mutation-kill revert (which turns it red) — not on the absence of a fixture that produces the shape.
- Gap 2 (04-VERIFICATION.md's `control:foreign.no_unproven_attributes` coverage gap) is out of scope for this plan; it is a separate gap-closure plan's responsibility per the phase's own gap-closure sequencing.

## Self-Check: PASSED

- `internal/compiler/corevalidate/corevalidate.go` exists and contains the new check — verified.
- `internal/compiler/corevalidate/corevalidate_test.go` exists and contains all five new tests — verified.
- `.planning/phases/04-fallible-resources-and-c-boundary/04-VALIDATION.md` exists and contains the two new register rows — verified.
- Commit `8535bfd` found in `git log --oneline --all`.
- Commit `3e8a4db` found in `git log --oneline --all`.
- All plan-level `<verification>` commands re-run and passing at summary time.

---
*Phase: 04-fallible-resources-and-c-boundary*
*Completed: 2026-09-05*
