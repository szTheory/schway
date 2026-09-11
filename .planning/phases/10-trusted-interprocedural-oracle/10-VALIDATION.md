---
phase: "10"
slug: "trusted-interprocedural-oracle"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-11"
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

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| TBD | TBD | TBD | SEM-08 | — | Depth-exceeded refusal fires through the real pipeline on a genuine 129-function chain | integration | `go test ./internal/compiler/interp/... -run TestCallDepthExceeded -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | SEM-08 (Pitfall 4 gate) | — | Native-stack headroom probe is a distinct limit from the language-level call-depth bound | subprocess integration | `go test ./internal/compiler/interp/... -run TestNativeStackHeadroomIndependentOfCallDepth -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | SEM-09 | — | Drop/cleanup order observed via canonical bytes on normal return **and** every nonlocal exit | unit + differential | `go test ./internal/compiler/interp/... -run TestFrameDrainOrder -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | TRU-02 | — | `walkReturnOrigin` handles `case core.OpCall`; `twin_a_accept.lang` admits clean through the full CLI | unit + CLI gate | `go test ./internal/compiler/originvalidate/... -run TestOpCallOriginWalk -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | TRU-02 (import guard) | T-10-01 | `originvalidate` transitively imports neither `check` nor `corevalidate` | build/test guard | `go test ./internal/compiler/originvalidate/... -run 'Imports.*Independent' -v` | ✅ (needs extension) | ⬜ pending |
| TBD | TBD | TBD | TRU-03 | — | Composition splits the caller's endpoints per-path depending on which callee path is spliced | unit | `go test ./internal/compiler/pathoracle/... -run TestCompositionDiscriminatesPerPathBorrow -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | QLT-04 | — | Depth-3 composition corpus reaches the **declared** bound, checked bidirectionally | corpus + bidirectional gate | `go test ./internal/compiler/session/... -run TestCompositionDepthCorpusReachesDeclaredBound -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | OWN-05b | T-10-02 | `interp` never reads `corevalidate.Result`'s ownership-bearing fields | guard | `go test ./internal/compiler/interp/... -run TestInterpDoesNotReadCorevalidateOwnershipFields -v` | ❌ W0 | ⬜ pending |
| TBD | TBD | TBD | Criterion 4 | — | Three-way-on-refuse / four-way-on-accept differential with seeded faults per peer pair | differential + mutation-kill | `go test ./internal/compiler/session/... -run TestNoUndeclaredCheckPeerDivergenceAcrossCorpus -v` | ✅ (needs extension) | ⬜ pending |
| TBD | TBD | TBD | Stability freeze | — | Golden corpus byte-for-byte and deterministic across repeated runs | golden + flake check | `go test ./internal/compiler/interp/... -run TestInterpOracleGoldenCorpus -v` and `-run TestInterpDeterministicAcrossRuns -count=10` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `internal/compiler/interp/interp_test.go` — depth-exceeded through the real pipeline (D-10-24), drop-order seam (D-10-35), native-stack probe (D-10-42), OWN-05b mutant-kill pairs (D-10-41)
- [ ] `internal/compiler/originvalidate/originvalidate_test.go` — `OpCall` widening test, `corevalidate` added to the forbidden-import list, transitive-scan upgrade, `TerminatorKindsOverride`-shaped seam (D-10-08)
- [ ] `internal/compiler/pathoracle/pathoracle_test.go` — composition test, discriminating per-path-borrow fixture (D-10-14, required deliverable), transitive-guard upgrade
- [ ] `testdata/phase10/` — depth-3 fixture pair varying which hop carries the borrow (D-10-48), the 129-function depth-exceeded fixture (or generator), `interp_oracle/` golden corpus
- [ ] `internal/compiler/session/session_peer_gate_test.go` — extended to three/four peers; `peerDivergenceExpected` upgraded to an accountable struct carrying debt-register ID + landing phase (D-10-54)
- [ ] Framework install: **none** — `testing` stdlib only

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| — | — | — | — |

*All phase behaviors have automated verification.*

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 120s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
