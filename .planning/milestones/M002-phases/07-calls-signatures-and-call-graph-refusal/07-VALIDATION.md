---
phase: "07"
slug: "calls-signatures-and-call-graph-refusal"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-08"
amended: "2026-09-08"
evidence_vocabulary: v1
graded_rows: 9
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Seeded from `07-RESEARCH.md` § Validation Architecture; **amended after
> cross-AI review**, when the plan set was rewritten from 5 plans to 8 and
> D-07-29..D-07-45 were locked.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go stdlib `testing` (`go test ./...`) — no external test framework in this repo, and none is to be added (zero-external-production-dependency record; `pgregory.net/rapid` is test-only and optional) |
| **Config file** | none — table-driven Go tests, fixtures under `testdata/phaseN/` |
| **Quick run command** | `go test ./internal/compiler/check/... ./internal/compiler/corevalidate/... ./internal/compiler/core/...` |
| **Full suite command** | `go test ./...` |
| **Race gate** | `go test -race ./...` — **promoted to a gate this phase** because `callgraph` is a production `check` dependency carrying fault-injection seams (D-07-42) |
| **Phase verify script** | `sh scripts/verify-phase7.sh` (created in `07-04`) |
| **Estimated runtime** | quick ~seconds; full suite dominated by native `-O0`/`-O3`/`-flto` lanes |

---

## Sampling Rate

- **After every task commit:** targeted package tests for the file(s) touched — `go test ./internal/compiler/<package>/...`
- **After every plan:** `go test ./...` plus `go vet ./...`; from `07-04` onward also `sh scripts/verify-phase7.sh`
- **After every plan that adds a fault-injection seam** (`07-01` .. `07-08`, all of them): `go test -race ./...`
- **Before `/gsd-verify-work`:** full suite green, race gate green, **both exhaustive-dispatch controls** green (the in-process `core_test.go` control and the phase-07-scoped CLI lane, each with its own literal fixture list — A-05), and `TestPhase7ControlsAreMutationKilled` green
- **Max feedback latency:** quick command, seconds

---

## Per-Task Verification Map

| Req ID | Behavior | Plan | Test Type | Automated Command | Grade | Non-inertness |
|---|---|---|---|---|---|---|
| SEM-05 | `lang.interface/1` strictly decoded; `/0` routed to a pinned struct; canonical non-self-referential digest preimage | 07-01 | unit | `go test ./internal/compiler/core/... -run 'DecodeInterface\|InterfaceV0'` | EXERCISED | — |
| SEM-05, SEM-06 | `Callable` = publication safety; producer/peer zero divergence over the M001 corpus at BOTH replay sites; five seeded faults | 07-02 | unit | `go test ./internal/compiler/corevalidate/... ./internal/compiler/originvalidate/... -run MutationMatrix` | EXERCISED | — |
| SEM-04 | `CalleeID` + `core.OpCall` recognized at all six dispatch sites; three `CalleeID` refusals at both sites; relocated foreign refusal | 07-03 | unit + CLI | `go test ./internal/compiler/{syntax,core,check,corevalidate,interp,cgen}/...` | WIRED | — |
| SEM-04 | Both exhaustive-dispatch controls green with their own phase-07 lists, asserting recognition not execution, each mutation-killed | 07-04 | unit + CLI | `go test ./internal/compiler/core/... -run TestAllOperationKindsHandledAtEverySite` and `sh scripts/verify-phase7.sh` | EXERCISED | — |
| SEM-06, SEM-05 | Pre-body signature table; call to a non-publishable callee refused; body-blindness falsifiable | 07-05 | unit + CLI | `go run ./cmd/lang --json check testdata/phase07/call_uncallable_callee.lang` | REACHABLE | — |
| SEM-07 | `callgraph` three-color DFS; deterministic cycle ID; 32-cause bound; diamond corpus; gray/self-edge mutations killed | 07-06 | unit + CLI | `go test ./internal/compiler/callgraph/... -run MutationMatrix` | EXERCISED | — |
| SEM-07, SEM-04 | Independent `corevalidate` cycle peer over synthetic artifacts; remaining corpora; phase-wide completeness by exact set equality | 07-07 | unit | `go test ./internal/compiler/session/... -run TestPhase7ControlsAreMutationKilled` | EXERCISED | — |
| SEM-05 | `ClosureDigest` chained over callee summary digests in reverse postorder over a proven DAG; callee-changes-invalidates-caller | 07-08 | unit | `go test ./internal/compiler/originvalidate/... -run 'CalleeChange\|Chain'` | EXERCISED | — |
| QLT-08 | Every control introduced this phase observed to fail against a seeded mutation **in the plan that introduced it** (D-07-41) | all | unit | `go test ./... -run 'MutationMatrix\|MutationKilled\|ControlsAreMutationKilled'` | EXERCISED | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Fixtures and infrastructure that do not exist and must be created before the
gates that depend on them. **Corrected after review** — one item previously
listed as missing was wrong.

- [ ] `testdata/phase07/call_basic.lang`, `call_from_both_match_arms.lang` (`07-03`)
- [ ] `testdata/phase07/clean_but_unpublishable.lang` — **EXTRACT, do not author.**
      The witness already exists as inline Go source at
      `check/check_exclusive_test.go:46-58` (module `owned.exclusive_borrow_clean`,
      `fn relay`), asserted again at `originvalidate_test.go:230`. Amendment A-03
      checked `testdata/` for a *file* and reported the wrong conclusion; **D-07-44**
      corrects it. It is the *right* witness because it fails publication with
      `core.origin_omitted` — D-07-31's real predicate. (`07-02`)
- [ ] `testdata/phase07/call_uncallable_callee.lang` (`07-05`)
- [ ] `testdata/phase07/relay_escort_witness.lang` — **author, and deliberately
      convert to A-normal form.** The D-04-03 narrative's
      `relay(borrow mut buffer)` is ungrammatical under D-07-01, so it cannot be
      copied verbatim; the equivalence argument goes in the fixture header
      (D-07-44). (`07-05`)
- [ ] `testdata/phase07/cycle_mutual.lang` (length 2), `cycle_self.lang` (length 1),
      **`deep_diamond_acyclic.lang`** (`07-06`). The diamond corpus is
      **load-bearing**: a chain-only corpus cannot kill the gray-versus-visited
      mutation, because reverse postorder never revisits on a chain.
- [ ] `testdata/phase07/cycle_indirect.lang` (length ≥ 3), `cycle_unreachable.lang`,
      `cycle_through_match_arm.lang`, `foreign_symbol_shadowing.lang` (`07-07`)
- [ ] **No `cycle_direct.lang`.** Fixtures are named by cycle length; the
      superseded name is retired (codex's naming finding).
- [ ] `internal/compiler/callgraph/callgraph.go` + `callgraph_test.go` — the package does not exist (`07-06`)
- [ ] `internal/compiler/session/session_phase7.go` + `_test.go` + `scripts/verify-phase7.sh` (`07-04`)
- [ ] `syntheticProgram(edges map[string][]string) core.Program` for the peer tests —
      **hand-built, never through the parser** (D-07-19). Building them through the
      parser is what would make the peer inert. Must produce programs that pass every
      structural validation running before the cycle peer, so a rejection is
      unambiguously the cycle. (`07-07`)
- [ ] A pinned `core.InterfaceV0` decode struct + frozen-bytes fixture proving no
      `lang.interface/0` byte moved, **plus `core.DecodeInterface` with schema-peek
      dispatch** — without the decoder, nothing routes a `/0` document to the pinned
      struct and no required-field claim is enforceable (D-07-36). (`07-01`)
- [ ] Framework install: **none.** Stdlib `testing` covers everything.

---

## Manual-Only Verifications

**None.** The previous revision listed the bilateral seeded fault (D-07-24
fault 3) as manual-only. That is now **automated**: the observable is the report
string `no divergence detected under bilateral fault` plus the gate-failed flag,
and the test asserts the failing report is produced rather than asserting the
sweep returns green. A passing assertion of a deliberately-failing gate is itself
the thing under test, and the report string makes it mechanically checkable.

The same pattern applies to the three later bilateral rows (`07-05` Test 6,
`07-07` Test 3, and the `Callable` bilateral in `07-02`).

---

## Determinism and Race Discipline (new this revision)

Two review findings raised the bar and both are enforced as validation policy,
not as per-task advice:

- **Determinism needs production ordering, not only repeated tests.** Map-backed
  builders can pass twice by luck. Root IDs, adjacency lists, emitted cycle
  members, digest callee pairs, and mutation-matrix rows are all explicitly
  sorted in production code, and every matrix is re-run under
  `-count=2 -shuffle=on` with identical per-control results required.
- **Fault seams are unexported** (D-07-42). `callgraph`, `check`, `corevalidate`,
  and `originvalidate` all run in production; exported mutable package globals
  there risk `-race` failures and test-order dependence. Every seam this phase
  adds is an injected predicate on a private configuration, `nil`-default,
  `defer`-restored, exercised by same-package tests — and `go test -race ./...`
  is a gate on every plan that adds one.

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify with a `<fails_when>` stating the failing direction
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references, with the A-03 correction applied
      (`exclusive_borrow_clean` is **extracted**, not authored)
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s for the quick command
- [ ] `go test -race ./...` green
- [ ] `TestPhase7ControlsAreMutationKilled` green by exact set equality against `Phase7RequiredControls()`
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
