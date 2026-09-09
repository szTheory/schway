---
phase: 07-calls-signatures-and-call-graph-refusal
verified: 2026-09-09T00:00:00Z
status: passed
score: 4/4 must-haves verified
behavior_unverified: 0
overrides_applied: 0
re_verification:
  previous_status: gaps_found
  previous_score: 4/4 (must-haves), 0/3 (post-verification gaps)
  gaps_closed:
    - "PVG-03 (CR-04): lang check never ran corevalidate.Validate — CheckCommandFile now consults it, InterfaceExportCommandFile/InterfaceCoreCommandFile now report a peer refusal as protocol.StatusInvalid instead of tool.operation_failed/exit 3"
    - "PVG-01 (CR-01): a call neither moved nor copy-checked its non-copyable argument — check.resolveCallBinding now consumes the argument (copy if AbilityCopy, move otherwise), independently re-derived by corevalidate's consumeCallArgument on both OpCall replay arms"
    - "PVG-02 (CR-03): core.FunctionSignature.Foreign/.Fails were computed locally, not closure-derived — originvalidate.BuildInterface now joins Foreign/Fails with every already-joined callee via an explicit worst-case lattice join before ClosureDigest, independently re-derived by corevalidate's own postorder join"
  gaps_remaining: []
  regressions: []
---

# Phase 07: Calls, Signatures, and Call-Graph Refusal — Verification Report

**Phase Goal:** `OpCall` becomes real at all six dispatch sites; cycles are
refused, never hung. (Full phase intent: a Lang function can call another
Lang function, admitted from the callee's signature alone, and a program
whose calls form a cycle is refused by name instead of hanging.)

**Verified:** 2026-09-09
**Status:** passed
**Re-verification:** Yes — after gap-closure plans 07-10, 07-11, 07-12, closing
post-verification gaps PVG-01/02/03 recorded by the prior 07-VERIFICATION.md.

## Central Question Answered

**Are PVG-01, PVG-02, and PVG-03 genuinely closed, and did closing them
regress anything in 07-01..07-09?**

Yes to both. Each PVG was independently re-tested live in this session — not
by reading the SUMMARYs — against the actual CLI, and the full existing
`testdata/phase07` corpus was swept to confirm no verdict changed except the
two deliberately-declared new divergences (`relay_escort_witness.lang`,
`duplicate_function_name.lang`, both from 07-10, both previously disclosed as
intended flips). `go build ./...`, `go vet ./...`, `go test ./... -p 1`, and a
full live run of `sh scripts/verify-phase7.sh` (which itself runs `go test
./...` and `go test -race ./...` from a clean cache) all completed with zero
failures, exit code 0, and all 24 phase-07 controls `pass` in
`kind-exhaustive-dispatch-phase07`.

## Goal Achievement

### Observable Truths (PVG re-tests)

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | **PVG-01 closed:** a call now consumes its non-copyable argument — the same `Buffer` passed to two separate calls is refused, not admitted | ✓ VERIFIED | Live run: `go run ./cmd/lang --json check testdata/phase07/call_argument_used_twice.lang` → `status: invalid`, `ownership.use_after_move`, exit 1. `call_argument_used_once.lang` (A/B control, non-refusing direction) → `status: pass`, exit 0. Source read: `check.resolveCallBinding` sets all four `placeState` fields on a non-copyable argument, mirroring the `take` arm; `corevalidate.consumeCallArgument` independently re-derives ability from the emitted `core.Program`'s own `types[operation.TypeID].Shape`, called from both `OpCall` replay arms, sharing no helper with `check`. |
| 2 | **PVG-02 closed:** `core.FunctionSignature.Foreign`/`.Fails` are closure-derived — a caller of a fallible, libc-reaching callee publishes the real reach, not the empty/zero value | ✓ VERIFIED | Live run: `go run ./cmd/lang --json interface export testdata/phase07/call_fallible_foreign_reach.lang /tmp/phase07_foreign_reach.json` → both `tracer` and `main` publish identical `foreign: {allocator: "libc_malloc", unwind: "forbidden", nonlocal_exit: "forbidden"}` and `fails: "ProbeError"`. Source read: `originvalidate.BuildInterface`'s `chainOrder` loop runs `joinForeignReach`/`joinFails` before `computeClosureDigest`; `corevalidate`'s `chainPeerClosureDigests` runs its own independently-written `peerJoinForeignReach`/`peerJoinFails` over its own `peerPostorder`/`peerAdjacency`, sharing no helper with `originvalidate` (only the `core.ForeignReachConflict` schema sentinel is shared). |
| 3 | **PVG-03 closed:** `lang check` now consults `corevalidate.Validate` and reports its refusal with the peer's own code, instead of discarding it | ✓ VERIFIED | Live run: `grep -n corevalidate.Validate internal/compiler/session/session.go` shows `CheckCommandFile` (the function `lang check` dispatches to) calling `corevalidate.Validate(checked.Program)` at line 732, in the documented precedence (check's own diagnostics, then the peer, then `originvalidate.ValidatePublished`). `relay_escort_witness.lang` → `status: invalid`, `core.move_while_borrowed`, exit 1 (was `status: pass` before 07-10). `duplicate_function_name.lang` → `status: invalid`, `core.duplicate_function_id`, exit 1 (new fixture). Both are declared, intended divergence flips, not accidental over-refusal. |
| 4 | **No regression:** every fixture in `testdata/phase07` unaffected by the three closures reports exactly the status/code the prior verification's Behavioral Spot-Checks table recorded | ✓ VERIFIED | Live CLI sweep of all 17 phase07 fixtures in this session: `call_basic` → pass; `call_type_mismatch` → `check.call_argument_type_mismatch`; `deep_diamond_acyclic` → pass; `call_uncallable_callee` → `core.callee_not_callable`; `clean_but_unpublishable` → `core.origin_omitted`; `foreign_symbol_shadowing`, `cycle_self`, `cycle_mutual`, `cycle_indirect`, `cycle_through_match_arm`, `cycle_unreachable` → all `core.call_graph_cycle`; `call_from_both_match_arms` → pass. No previously-admitted fixture is now refused for an unrelated reason; no previously-refused fixture is now admitted. |

**Score:** 4/4 truths verified.

### Requirements Coverage

| Requirement | Source Plan(s) | Description | Status | Evidence |
|-------------|-----------------|-------------|--------|----------|
| SEM-04 | 07-03, 07-04, 07-07, 07-10, 07-11 | `OpCall` real at all six dispatch sites, both exhaustive controls green | ✓ SATISFIED | REQUIREMENTS.md marks `[x]` complete; unchanged by 07-10..12 except the CLI now surfacing the peer; re-confirmed live via `verify-phase7.sh`'s 24-control `kind-exhaustive-dispatch-phase07` lane, all `pass`. |
| SEM-05 | 07-01, 07-02, 07-05, 07-08, 07-09, 07-11, 07-12 | Digest-bound signature summary carries everything a caller needs for admission; no admission reads a callee body; the signature is actually consulted for argument-type soundness AND ownership AND foreign-reach | ✓ SATISFIED, gaps closed | The type half closed in 07-09 (prior verification). The ownership half (PVG-01) and the `Foreign`/`Fails` closure-derivation half (PVG-02) close here, independently re-tested above. |
| SEM-06 | 07-02, 07-05, 07-10 | Call admitted only when callee callable ⊆ publishable; stable refusal code | ✓ SATISFIED | `call_uncallable_callee.lang` → `core.callee_not_callable`, unchanged; `interface` command paths now report a peer refusal as `StatusInvalid` rather than discarding it (PVG-03/07-10). |
| SEM-07 | 07-06, 07-07 | Call graph constructed, direct/mutual/indirect cycles refused by name, never a hang | ✓ SATISFIED | Full cycle corpus re-confirmed unchanged live in this session. |
| QLT-08 | all 12 plans | Every new interprocedural control mutation-killed in its introducing plan | ✓ SATISFIED | 24 controls total in `Phase7RequiredControls()`/`scripts/verify-phase7.sh`/`controlsWithRecordedMutationKill`, all `pass` in the live run, including the 8 controls added across 07-09..07-12 (2 each). |

No orphaned requirements: all five IDs (SEM-04, SEM-05, SEM-06, SEM-07,
QLT-08) map to Phase 07 in REQUIREMENTS.md and appear `[x]` complete there
(lines 15-25, 89, 163-166, 186).

### Live Verification Runs (this session)

| Check | Command | Result |
|-------|---------|--------|
| Build | `go build ./...` | clean |
| Vet | `go vet ./...` | clean |
| Full test suite | `go test ./... -p 1` | 23/23 packages `ok`, zero `FAIL` |
| Full gate script (build/vet/test/-race/verify all corpora) | `sh scripts/verify-phase7.sh` | exit code 0; all `verify` JSON emitted `status: pass`; `lane:kind-exhaustive-dispatch-phase07` lists 24 controls, all `pass` |
| PVG-01 refusing direction | `lang --json check call_argument_used_twice.lang` | `status: invalid`, `ownership.use_after_move` |
| PVG-01 non-refusing direction | `lang --json check call_argument_used_once.lang` | `status: pass` |
| PVG-02 | `lang --json interface export call_fallible_foreign_reach.lang` | `main` and `tracer` both publish identical non-empty `foreign`/`fails` |
| PVG-03 | `lang --json check relay_escort_witness.lang`, `duplicate_function_name.lang` | both `status: invalid` via peer-sourced codes (`core.move_while_borrowed`, `core.duplicate_function_id`) |
| Regression sweep | `lang --json check` on all 17 `testdata/phase07/*.lang` | every fixture matches the documented/expected verdict, no unexplained flip |
| Anti-pattern scan | `grep -nE 'TBD|FIXME|XXX'` on all files touched by 07-10/11/12 | none found |

### Anti-Patterns Found

None. No `TBD`/`FIXME`/`XXX`/`TODO`/`HACK`/`PLACEHOLDER` markers in any file
modified by the gap-closure plans. `07-REVIEW.md` (the fresh gap-closure code
review) found 0 critical, 1 warning (WR-01), 2 info — none rise to blocker
level.

### Carried Warning (not a gap — noted for awareness)

**WR-01** (`07-REVIEW.md`): `corevalidate.consumeCallArgument` derives
copy-ability from `types[operation.TypeID].Shape` (the call's target/return
type) rather than `types[source.TypeID].Shape` (the argument's own type).
This is safe today only because a separate, independently-checked invariant
(`sameType`, D-07-09) forces the two to agree for every legally-typed
program — the generic `source.TypeID==operation.TypeID` law isn't
`OpCall`-specific and would not by itself catch a future divergence. This is
disclosed, not silently absorbed, and does not admit an unsound program
today. Not a Phase 07 blocker; the reviewer's suggested fix (derive from
`source.TypeID` directly, or add an explicit `OpCall`-specific assertion) is
appropriately deferred as a hardening item, not required for this phase's
goal.

### Debt Register

`PHASE-07-DEBT.md` is well-formed (11 items, `TestDebtRegistersAreWellFormed`
passing), correctly records D-07-49 (PVG-04/CR-02, explicitly assigned to
Phase 08, out of scope here), D-07-50/51 (WR-01/WR-02 from the round-1
review, check-side halves deliberately carried), D-07-52 (the accepted
implicit call-site transfer residual from PVG-01's closure), D-07-53 (join
granularity limits from PVG-02's closure), and D-07-54 (IN-02/IN-03 carried).
D-07-07 is corrected, not merely supplemented, to no longer imply the
single-argument ownership case was already handled before 07-11.

### Human Verification Required

None. Every claim above was independently reproduced by running the actual
CLI and reading the actual source in this session — not by trusting
SUMMARY.md or REVIEW.md narration.

### Gaps Summary

None remaining. All three post-verification gaps (PVG-01, PVG-02, PVG-03)
from the prior 07-VERIFICATION.md are closed and independently re-verified.
PVG-04 (CR-02, `computeLoanLastUses` has no `"call"` case) was explicitly out
of scope for this re-verification per the task's own instructions — it is
Phase 08's stated subject and remains correctly assigned there in
`PHASE-07-DEBT.md` D-07-49, unmodified by 07-10/11/12. No regression was
found anywhere in the existing `testdata/phase07` corpus, and the full
`go test ./... -p 1`, `go test -race ./...` (via `verify-phase7.sh`), and
`go vet ./...` are all green.

**Phase 07 is complete.** All 5 requirement IDs (SEM-04, SEM-05, SEM-06,
SEM-07, QLT-08) are satisfied, all must-haves verified, and no open
post-verification gap remains.

---

_Verified: 2026-09-09_
_Verifier: Claude (gsd-verifier)_
