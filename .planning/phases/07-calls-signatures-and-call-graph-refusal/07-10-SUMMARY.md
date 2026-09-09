---
phase: 07-calls-signatures-and-call-graph-refusal
plan: 10
subsystem: compiler-session
tags: [go, call-admission, corevalidate, protocol, gap-closure]

requires:
  - phase: 07-01..07-09
    provides: check's call admission, call-graph refusal, closure-digest chaining, and argument/return type gates; corevalidate's independent peer re-derivations of each
provides:
  - "lang check consults corevalidate.Validate on every source it admits and reports the peer's refusal with the peer's own code, not silently"
  - "interface export / interface core report a peer refusal as protocol.StatusInvalid with the peer's own code, not tool.operation_failed/exit 3 with the code discarded"
  - "peerDivergenceExpected: a named, asserted, both-directions register of every check-admits/peer-refuses fixture in testdata/"
  - "two new mutation-killed controls (control:check.peer_consulted, control:interface.peer_refusal_is_invalid) wired into all three phase-07 registries"
  - "PHASE-07-DEBT.md D-07-49/50/51 recording the PVG-04/WR-01/WR-02 dispositions"
affects: [08-interprocedural-loan-liveness-in-check, 07-11-ownership-peer, 07-12-closure-derived-signature-peer]

actuals:
  tokens: 42000
  tasks: 3
  commits: 3

tech-stack:
  added: []
  patterns:
    - "Refusing union of two independent admission layers, fixed precedence order (check diagnostics, then corevalidate peer, then originvalidate on the peer-normalized program), asserted never reconciled"
    - "Shared diagnostic-FORMATTING helper (peerRefusalDiagnostic) across three call sites reporting the same already-computed peer verdict -- not a reconciliation of two derivations"
    - "Named, commented, both-directions-asserted divergence register (peerDivergenceExpected) as the mechanism that keeps an undeclared future divergence from passing silently"

key-files:
  created:
    - internal/compiler/session/session_peer_gate_test.go
    - testdata/phase07/duplicate_function_name.lang
  modified:
    - internal/compiler/session/session.go
    - internal/compiler/session/session_phase7.go
    - internal/compiler/session/session_phase7_test.go
    - internal/compiler/session/session_phase7_export_test.go
    - internal/compiler/corevalidate/corevalidate_summary_peer_test.go
    - scripts/verify-phase7.sh
    - .planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md

key-decisions:
  - "Checkpoint auto-ratified (auto_advance=true, gate not blocking-human): accepted the union rule, both CLI verdict flips (relay_escort_witness.lang, duplicate_function_name.lang), the interface-path protocol change, and core.peer_refusal_unnamed as the fallback constant name, verbatim as proposed."
  - "Extracted peerRefusalDiagnostic as shared formatting (not reconciliation) across CheckCommandFile, InterfaceExportCommandFile, and InterfaceCoreCommandFile -- reduces duplication of the fail-closed fallback logic across three call sites reporting one already-computed peer verdict."
  - "interfacePeerRefusalSeam guards both interface command paths with one seam (one fact, asserted twice), matching the plan's own stated design."

requirements-completed: [SEM-04, SEM-06, QLT-08]

coverage:
  - id: D1
    description: "lang check consults corevalidate.Validate and reports its refusal with the peer's own code; relay_escort_witness.lang and duplicate_function_name.lang both flip from status:pass to status:invalid at the CLI"
    requirement: SEM-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestCheckCommandSurfacesPeerRefusal"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestDuplicateFunctionDeclarationRefusedAtCLI"
        status: pass
      - kind: e2e
        ref: "go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang"
        status: pass
    human_judgment: false
  - id: D2
    description: "interface export / interface core report a peer refusal as protocol.StatusInvalid carrying the peer's own code, with a nil error and no written artifact -- never tool.operation_failed/exit 3 with the code discarded"
    requirement: SEM-06
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestInterfaceCommandsReportPeerRefusalAsInvalid"
        status: pass
    human_judgment: false
  - id: D3
    description: "an asserted, both-directions divergence register: the only check-admits/peer-refuses fixtures in the whole testdata/ corpus are the two declared ones"
    requirement: SEM-04
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestNoUndeclaredCheckPeerDivergenceAcrossCorpus"
        status: pass
    human_judgment: false
  - id: D4
    description: "two new controls (control:check.peer_consulted, control:interface.peer_refusal_is_invalid) mutation-killed in both directions and wired into Phase7RequiredControls(), scripts/verify-phase7.sh, and controlsWithRecordedMutationKill with exact-set-equality unweakened"
    requirement: QLT-08
    verification:
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestCheckCommandPeerConsultMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_peer_gate_test.go#TestInterfacePeerRefusalMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7ControlsAreMutationKilled"
        status: pass
      - kind: unit
        ref: "internal/compiler/session/session_phase7_test.go#TestPhase7RequiredControlsMatchScript"
        status: pass
    human_judgment: false

duration: 105min
completed: 2026-09-09
status: complete
---

# Phase 07 Plan 10: Peer-consulted `lang check` and honest `interface` refusals Summary

**`CheckCommandFile` now takes the refusing union of `check`'s own diagnostics and `corevalidate.Validate`'s independent verdict, closing 07-REVIEW.md CR-04 (PVG-03) — the gate users actually run no longer silently admits a program the independent peer refuses.**

## Performance

- **Duration:** ~105 min
- **Tasks:** 3 (plus one auto-ratified checkpoint)
- **Files modified:** 9 (2 created, 7 modified)

## Accomplishments

- `CheckCommandFile` consults `corevalidate.Validate` between `check`'s own diagnostics and `originvalidate.ValidatePublished`, in the exact precedence `InterfaceExportCommandFile` already used. `relay_escort_witness.lang` (the disclosed D-03-02 interprocedural divergence) and a new `duplicate_function_name.lang` fixture (WR-01's user-visible half) both flip from `status: pass` / exit 0 to `status: invalid` with the peer's own code (`core.move_while_borrowed`, `core.duplicate_function_id`). `check.Program` itself is byte-for-byte unchanged; `TestRelayEscortWitnessChecksCleanPendingInterproceduralLiveness` stays green.
- `InterfaceExportCommandFile` and `InterfaceCoreCommandFile` no longer discard a peer refusal behind `fmt.Errorf("core validation failed: …")` → `tool.operation_failed`/exit 3. Both now return `protocol.StatusInvalid` carrying the peer's own code, with a nil error and no output artifact written on a refusal.
- `peerDivergenceExpected` (session_peer_gate_test.go) asserts, in both directions, that the entire `testdata/` corpus has exactly two check-admits/peer-refuses fixtures — an undeclared new divergence, or a stale entry that stops diverging, fails the test.
- `control:check.peer_consulted` and `control:interface.peer_refusal_is_invalid` are wired into `Phase7RequiredControls()` (20 identifiers), `scripts/verify-phase7.sh`'s own control loop, and `controlsWithRecordedMutationKill`, each observed to fail under its own seeded seam in both directions.
- PHASE-07-DEBT.md records D-07-49 (PVG-04/CR-02 now CLI-observable, still Phase 08 scope), D-07-50 (WR-01's check-side half carried), and D-07-51 (WR-02 carried) — none moved or silently absorbed.

## Task Commits

Each task was committed atomically:

1. **Task 1: `lang check` consults the peer and reports its refusal** - `7b562dd` (feat)
2. **Task 2: interface paths report a peer refusal as invalid, not a tool failure** - `e25d63e` (feat)
3. **Task 3: mutation-kill both controls, wire into all three registries, record debt** - `045973a` (feat)

**Plan metadata:** commit pending (this SUMMARY + STATE.md + ROADMAP.md)

_Checkpoint (ratify the verdict-union rule): auto-approved under auto_advance mode — the default/first option was accepted verbatim, no separate commit._

## Files Created/Modified

- `internal/compiler/session/session.go` - `CheckCommandFile` consults `corevalidate.Validate`; `InterfaceExportCommandFile`/`InterfaceCoreCommandFile` report a peer refusal as `StatusInvalid`; new `peerRefusalDiagnostic` shared formatter, `peerRefusalUnnamedCode` fallback const, `checkCommandPeerSeam`/`interfacePeerRefusalSeam` fault-injection seams
- `internal/compiler/session/session_phase7.go` - `ControlCheckPeerConsulted`, `ControlInterfacePeerRefusalIsInvalid` consts; both appended to `Phase7RequiredControls()`
- `internal/compiler/session/session_phase7_test.go` - both controls added to `controlsWithRecordedMutationKill`
- `internal/compiler/session/session_phase7_export_test.go` - `PeerRefusalDiagnosticForTest`, `PeerRefusalUnnamedCodeForTest`, `SetCheckCommandPeerSeamForTest`, `SetInterfacePeerRefusalSeamForTest`
- `internal/compiler/session/session_peer_gate_test.go` (new) - `peerDivergenceExpected` register, corpus sweep, both command-path tests, both mutation-kill tests
- `internal/compiler/corevalidate/corevalidate_summary_peer_test.go` - generalized the single-module corpus exception into `isDeclaredPeerDivergentCorpusModule`, naming `duplicate_function_name.lang` as a second declared instance (deviation, see below)
- `scripts/verify-phase7.sh` - both new control identifiers added to the phase-07 required-control loop
- `testdata/phase07/duplicate_function_name.lang` (new) - two colliding `fn helper` declarations; standing negative control for WR-01's user-visible half
- `.planning/phases/07-calls-signatures-and-call-graph-refusal/PHASE-07-DEBT.md` - D-07-49, D-07-50, D-07-51 added (items: 5 → 8)

## Decisions Made

- Checkpoint auto-ratified under `auto_advance`: accepted the union rule, both CLI verdict flips, the `interface`-path protocol change, and `core.peer_refusal_unnamed` as proposed — no exemption, no override.
- `peerRefusalDiagnostic` extracted as one shared FORMATTER for three call sites reporting the SAME already-computed peer verdict. This is not a reconciliation of two independent derivations (forbidden by the plan's own coordinated-blindness prohibition): it reads nothing from `check`, and each of the three callers still independently decides whether to call it.
- One seam (`interfacePeerRefusalSeam`) guards both `interface` command paths, matching the plan's stated "one fact, asserted twice" design, rather than two separate seams.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug/plan-drift] `core validation failed` grep acceptance criterion was stale relative to the actual codebase**
- **Found during:** Task 2
- **Issue:** Task 2's acceptance criteria expected `grep -v '^\s*//' session.go | grep -c 'core validation failed'` to equal exactly `1` (only `RunInterpreter`'s site surviving) after both `interface` command sites were converted. In the actual tree, three additional non-`interface` sites already carried the same literal (`RunNative`, `DebugMapCommandFile`, `Phase4CheckedProgram`) — none named in the plan's `read_first` or files list, and none is an `interface` command path in scope for this plan.
- **Fix:** Left all three untouched (correctly out of scope: none is `InterfaceExportCommandFile` or `InterfaceCoreCommandFile`). Verified the real invariant instead: both `interface` command sites' `fmt.Errorf(...)` are gone, and the seam's own replacement error text uses a different literal so it doesn't reintroduce the discarded one.
- **Files modified:** none beyond the planned session.go edit
- **Verification:** post-Task-2 count is 4 (not 1), all four surviving sites independently confirmed out of scope
- **Committed in:** e25d63e (Task 2 commit, documented in the commit message)

**2. [Rule 1 - Bug, blocking] `duplicate_function_name.lang` broke three pre-existing corevalidate corpus-wide tests**
- **Found during:** Task 3 (surfaced during the full `sh scripts/verify-phase7.sh` re-run)
- **Issue:** `TestSummaryPeerStructuralFieldsMatchProducerAcrossCorpus`, `TestSummaryPeerClosureDigestMatchesProducerAcrossCorpus`, and `TestSummaryPeerCallableAgreesOnOriginOmittedClass` (internal/compiler/corevalidate) walk the whole `testdata/` corpus and assumed exactly ONE named checked-clean/corevalidate-invalid exception existed (`relay_escort_witness.lang`). The new fixture is a second, deliberately divergent one by design (that's what proves the CLI-observable refusal) and failed all three.
- **Fix:** Generalized the single-module constant into `isDeclaredPeerDivergentCorpusModule(module string) bool`, naming both `relayEscortWitnessModule` and the new `duplicateFunctionNameModule` explicitly — a second declared instance of the same existing pattern, not a broadened invariant.
- **Files modified:** internal/compiler/corevalidate/corevalidate_summary_peer_test.go (not in Task 3's declared files list, but required to keep `go test ./...` green)
- **Verification:** `go test ./internal/compiler/corevalidate/... -count=1` green; full `go test ./... -p 1` and `go test -race ./...` both green afterward
- **Committed in:** 045973a (Task 3 commit)
- **Note on the plan's own verification item 7** (`git diff --name-only -- internal/compiler/check/ internal/compiler/corevalidate/ internal/compiler/originvalidate/` empty): this fix touches `corevalidate_summary_peer_test.go`, so that literal grep now reports one path. The underlying invariant it protects — `corevalidate.go`'s own DERIVATION logic is untouched — holds: `git diff --name-only -- internal/compiler/corevalidate/corevalidate.go` is empty across all three tasks; only the test's own corpus-exception list was extended, following its own pre-existing pattern for the same reason.

---

**Total deviations:** 2 auto-fixed (1 plan-drift/stale-criterion, 1 blocking bug fix required to keep the build green).
**Impact on plan:** Both are necessary corrections; neither changes what the plan actually delivers. No scope creep — the corevalidate test change is a mechanical generalization of an existing named-exception pattern, not new derivation logic.

## Issues Encountered

None beyond the two deviations above (which were resolved inline, not left open).

## Known Stubs

None.

## Threat Flags

None — every new surface (the peer consult, the two seams, the divergence register) is inside this plan's own declared `<threat_model>`.

## User Setup Required

None - no external service configuration required.

## Verification Results (re-run live)

- `go build ./...` — clean.
- `go vet ./...` — clean.
- `go test ./... -p 1 -count=1` — 23/23 packages `ok`, zero `FAIL`.
- `go test -race ./...` — 23/23 packages `ok`, zero `FAIL`, zero `DATA RACE`.
- `sh scripts/verify-phase7.sh` — exits 0; `phase07.json` lists both `control:check.peer_consulted` and `control:interface.peer_refusal_is_invalid` with status `pass` (20 controls total). Three unrelated environment-contention flakes observed under the script's own concurrent load (`cache.input_undeclared`, `cgen`'s `native.timeout`, `measure.probe_timeout`) all passed in isolated re-runs — none touch `check`, `corevalidate`, or `session`, matching 07-09-SUMMARY.md's own documented precedent for this machine.
- `go run ./cmd/lang --json check testdata/phase07/relay_escort_witness.lang` — `status: invalid`, `core.move_while_borrowed`, non-zero exit.
- Every fixture in `testdata/phase07` other than the two declared divergence entries reports exactly the status/code 07-VERIFICATION.md's Behavioral Spot-Checks table recorded pre-07-10 (re-verified via direct CLI sweep of all 13 phase07 fixtures).
- `git diff --name-only -- internal/compiler/check/` is empty across all three task commits.
- `git diff --name-only -- internal/compiler/corevalidate/corevalidate.go internal/compiler/originvalidate/` is empty across all three task commits (only `corevalidate_summary_peer_test.go`, a test file, changed — see Deviation 2 above).
- `git diff --name-only -- go.mod go.sum` is empty.

## Next Phase Readiness

- CR-04 / PVG-03 is closed: the phase's two-peer safety argument is now backed on the gate users actually run.
- 07-11 (ownership peer, PVG-01/CR-01) and 07-12 (closure-derived signature peer, PVG-02/CR-03) will now have their refusals observable through `lang check` immediately, rather than against a channel that discarded them.
- PVG-04/CR-02 remains explicit Phase 08 scope (D-07-49); `check.computeLoanLastUses` is untouched.
- WR-01's check-side half and WR-02 remain explicitly carried (D-07-50, D-07-51), each needing its own future ratification.

---
*Phase: 07-calls-signatures-and-call-graph-refusal*
*Completed: 2026-09-09*

## Self-Check: PASSED

All key files confirmed present on disk (session_peer_gate_test.go, testdata/phase07/duplicate_function_name.lang, session.go, session_phase7.go); all three task commits (7b562dd, e25d63e, 045973a) confirmed in `git log --oneline --all`.
