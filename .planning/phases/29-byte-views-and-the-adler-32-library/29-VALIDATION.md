---
phase: "29"
slug: "byte-views-and-the-adler-32-library"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# Planning-time draft only: dropped at plan commit (TestPhase20ValidationLifecycle forbids
# drafts; validate-phase recreates it after execution, as in Phase 28).
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-10-11"
---

# Phase 29 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.
> Test map source: `29-RESEARCH.md` § Validation Architecture.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing`, go1.24.0 |
| **Config file** | none (go.mod only) |
| **Quick run command** | `go test ./internal/compiler/<pkg> -run 'Phase29' -count=1` for each touched package |
| **Full suite command** | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest); never run locally |
| **Estimated runtime** | ~30 seconds per focused package run |

---

## Sampling Rate

- **After every task commit:** the focused `-run 'Phase29…'` for each touched package, plus the session meta-test set whenever a `.schway` file, a `session.go` call site, or a planning document with a `go test` row changed:
  `go test ./internal/compiler/session -run 'Inventory|Groundedness|Reconciliation|ValidationLifecycle|TestPhase28EntryExitClassManifest|TestPhase272AcceptedCorpusByteIdentity|TestPhase273ConformanceCorpusByteIdentity|TestLanguageMaturityCountsAreCurrent' -count=1`
- **After every plan wave:** all `Phase29` rows for the wave plus `go test ./internal/compiler/session -run 'TestPhase26|TestPhase27|TestPhase28' -count=1`
- **Before `/gsd-verify-work`:** hosted full and race CI green on both OSes; spike 008 hosted run IDs recorded
- **Max feedback latency:** 60 seconds

---

## Per-Requirement Verification Map

| Requirement | Behavior | Test Type | Automated Command | File Exists |
|-------------|----------|-----------|-------------------|-------------|
| BYTE-01 | parse `byte_len`/`byte_at`, arity refusals, reserved names, formatter goldens | unit | `go test ./internal/compiler/syntax -run 'Phase29' -count=1` | W0 |
| BYTE-01 | each consumer derives facts independently; `0x80`/`0xFF` read 128/255 | peer | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'Phase29' -count=1` | W0 |
| BYTE-01 | op registry and exhaustive dispatch | unit | `go test ./internal/compiler/core -run 'TestAllOperationKindsRegistered|TestAllOperationKindsHandledAtEverySite|TestPhase26CoreInventory' -count=1` | exists (edit) |
| BYTE-01 | interp and cgen semantics, unsigned widening, signed-char mutant killed | unit | `go test ./internal/compiler/interp ./internal/compiler/cgen -run 'Phase29' -count=1` | W0 |
| BYTE-02 | OOB: interpreter defect (D-19); native `-O0`/`-O3`/`-O3`+LTO exit 65, exact stderr, empty stdout (D-21 seam) | integration | `go test ./internal/compiler/native -run 'Phase29OOB' -count=1` | W0 |
| BYTE-02 | evidence peer and schema-2 projection carry the OOB reason | unit | `go test ./internal/compiler/executionpeer ./internal/compiler/session -run 'Phase29' -count=1` | W0 |
| BYTE-03 | admitted loop shape; neighbors refused with existing codes; FLOW-02 unchanged | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate -run 'TestBackEdgeRejected|TestPhase26|TestPhase29' -count=1` | exists + W0 |
| BYTE-03 | release-while-view-live refused independently (seeded faults) | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle -run 'Phase29Release' -count=1` | W0 |
| BYTE-03 | no per-byte events | unit | `go test ./internal/compiler/check ./internal/compiler/interp -run 'Phase29NoLoopEvents' -count=1` | W0 |
| APP-10 | six pinned vectors on both engines; mutants killed; C/Go generators agree | integration | `go test ./internal/compiler/session ./internal/compiler/native -run 'Phase29Adler' -count=1` | W0 |
| APP-10 | fixture owner end to end | integration | `go test ./internal/compiler/native -run 'Phase29Fixture' -count=1` | W0 |
| D-12/D-14 | cap measurement (env-gated spike harness) | spike | `SCHWAY_SPIKE008=1 go test ./internal/compiler/native -run 'Spike008' -count=1 -v` | W0 |

---

## Wave 0 Requirements

- [ ] Characterization pins of today's neighbor-refusal codes, committed before any parser/check change
- [ ] Baseline green of the three corpus byte-identity/manifest walkers at HEAD before edits
- [ ] `TestPhase29*` skeletons per package with each seeded fault proven red (no zero-match `-run` patterns)
- [ ] Test-only application-shell `-O3`/`-O3 -flto` compile-and-run helper (D-21)
- [ ] Stage A seams: Go-injected `[]byte` for the interpreter and a gate-only native driver (D-14)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Hosted macOS + Linux cap timings | D-12/D-13 | Needs hosted runners | Dispatch the existing CI job, record run IDs in `.planning/spikes/008-byteview-cap/README.md` |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
