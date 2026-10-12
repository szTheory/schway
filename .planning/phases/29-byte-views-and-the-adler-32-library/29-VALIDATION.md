---
phase: "29"
slug: "byte-views-and-the-adler-32-library"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# Reconstructed after execution from the 29-00..29-12 PLAN/SUMMARY set (State B).
status: validated
nyquist_compliant: true
wave_0_complete: true
created: "2026-10-11"
---

# Phase 29 — Validation Strategy

> Per-phase validation contract, reconstructed after execution.
> Sources: 29-00..29-12 PLAN and SUMMARY files, `29-RESEARCH.md` § Validation Architecture,
> `29-VERIFICATION.md`, and hosted run 38186068052.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing`, go1.24.0 |
| **Config file** | none (go.mod only) |
| **Quick run command** | `go test ./internal/compiler/syntax ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle ./internal/compiler/interp ./internal/compiler/cgen ./internal/compiler/executionpeer ./internal/compiler/session ./internal/compiler/native ./cmd/schway -run 'Phase29' -count=1` |
| **Full suite command** | hosted CI only (`go test ./...` and `go test -race -timeout=20m ./...` on ubuntu-latest and macos-latest); never run locally |
| **Estimated runtime** | ~10 seconds for every Phase 29 test across all eleven packages |

---

## Sampling Rate

- **After every task commit:** the focused `-run 'Phase29'` for each touched package, plus the session meta-test set when a `.schway` file, a `session.go` call site, or a planning document with a `go test` row changed.
- **After every plan wave:** every `Phase29` test for the wave's packages.
- **Before `/gsd-verify-work`:** hosted full and race CI green on both OSes; spike 008 strict step green.
- **Max feedback latency:** 60 seconds

---

## Per-Requirement Verification Map

| Requirement | Plans | Behavior | Test Type | Automated Command | File Exists | Status |
|-------------|-------|----------|-----------|-------------------|-------------|--------|
| BYTE-01 | 01, 02 | parse `byte_len`/`byte_at`, arity refusals, `data` parameter refused, formatter goldens | unit | `go test ./internal/compiler/syntax -run 'Phase29' -count=1` | ✅ | ✅ green |
| BYTE-01 | 02, 03 | reserved builtin names; byte reads never resolve as user calls; peers derive facts independently | peer | `go test ./internal/compiler/check -run 'TestPhase29ReservedBuiltinNames|TestPhase29ByteReadNotResolvedAsUserCall|TestPhase29PeerIndependence' -count=1` | ✅ | ✅ green |
| BYTE-01 | 03 | each consumer applies the byte-op laws independently | peer | `go test ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'Phase29' -count=1` | ✅ | ✅ green |
| BYTE-01 | 03 | op registry and exhaustive dispatch | unit | `go test ./internal/compiler/core -run 'TestAllOperationKindsRegistered|TestAllOperationKindsHandledAtEverySite|TestPhase26CoreInventory' -count=1` | ✅ | ✅ green |
| BYTE-01 | 04 | interpreter loop order, empty view, unsigned `0xFF` reads; cgen `const unsigned char` view | unit | `go test ./internal/compiler/interp ./internal/compiler/cgen -run 'Phase29' -count=1` | ✅ | ✅ green |
| BYTE-02 | 04, 07 | interpreter OOB defect at the boundary; cgen bounds check precedes the load | unit | `go test ./internal/compiler/interp ./internal/compiler/cgen -run 'TestPhase29InterpreterOOBBoundary|TestPhase29BoundsCheckPrecedesLoad' -count=1` | ✅ | ✅ green |
| BYTE-02 | 07, 08 | native `-O0`/`-O3`/`-O3`+LTO exit 65, exact stderr, empty stdout | integration | `go test ./internal/compiler/native -run 'TestPhase29OOBApplicationLanes' -count=1` | ✅ | ✅ green |
| BYTE-02 | 08, 10 | evidence peer and schema-2 projection carry the OOB reason | unit | `go test ./internal/compiler/executionpeer ./internal/compiler/session -run 'TestPhase29ExecutionPeerByteIndexDefect|TestPhase29OOBEvidenceProjection' -count=1` | ✅ | ✅ green |
| BYTE-03 | 03, 06 | admitted loop shape; neighbors refused with existing codes; seeded neighbor faults | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'TestPhase29NeighborRefusals|TestPhase29NeighborSeededFaults|TestPhase29LoopNeighborParseRefusals' -count=1` | ✅ | ✅ green |
| BYTE-03 | 09 | release or move of the owner while a view is live refused by each checker (D-22) | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/pathoracle -run 'TestPhase29ReleaseBeforeViewUseRefused|TestPhase29MoveOwnerWhileViewLive|TestPhase29OwnerLoopBlockRefused' -count=1` | ✅ | ✅ green |
| BYTE-03 | 04, 06 | no per-byte events, structurally and at run time | unit | `go test ./internal/compiler/check ./internal/compiler/interp -run 'TestPhase29NoLoopEventsStructure|TestPhase29NoLoopEventsInterpreter' -count=1` | ✅ | ✅ green |
| APP-10 | 05, 06 | fixture owner shape, lifecycle, abilities, origin matchers, pathoracle replay | unit | `go test ./internal/compiler/check ./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle -run 'Phase29Fixture|TestPhase29PathoracleFixtureReplay|TestPhase29ByteViewAbilityAgreement' -count=1` | ✅ | ✅ green |
| APP-10 | 05 | independent C and Go byte generators agree; unknown case refused | integration | `go test ./internal/compiler/native -run 'TestPhase29FixtureBytes' -count=1` | ✅ | ✅ green |
| APP-10 | 07, 10 | six pinned vectors on both engines; mutants killed | integration | `go test ./internal/compiler/native -run 'TestPhase29AdlerBothEngines|TestPhase29AdlerMutantsKilled' -count=1` | ✅ | ✅ green |
| APP-10 | 08, 10 | native and interpreter fixture runs through the session API; admission checks | integration | `go test ./internal/compiler/session -run 'TestPhase29NativeFixtureVectors|TestPhase29InterpreterFixture|TestPhase29FixtureAdmission|TestPhase29FixtureChecks' -count=1` | ✅ | ✅ green |
| APP-10 | 10 | `schway build --manifest` then `schway app run` end to end | integration | `go test ./cmd/schway -run 'TestPhase29FixtureAppCLI' -count=1` | ✅ | ✅ green |
| D-12/D-14 | 00 | Stage A gate: callable, interpreter and native proxy vectors | integration | `go test ./internal/compiler/session ./internal/compiler/cgen -run 'TestPhase29GateCallable|TestPhase29GateInterpreterVectors|TestPhase29GateNativeProxyVectors|TestPhase29NeighborBaseline' -count=1` | ✅ | ✅ green |
| D-12/D-14 | 12 | Stage B cap measurement through the real owner (non-strict locally) | spike | `go test ./internal/compiler/native -run 'TestSpike008ByteViewCap' -count=1 -v` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

Local evidence (2026-10-11, darwin/arm64): every `Phase29|Spike008` test across the eleven packages passed (228 PASS lines). Two cost readings skip locally by design: `TestPhase29GateNativeProxyCost` and `TestPhase29GateCostReading`.

---

## Wave 0 Requirements

- [x] Characterization pins of the neighbor-refusal codes before any parser or check change (`TestPhase29NeighborBaseline`)
- [x] `TestPhase29*` tests per package, with each seeded fault shown to fail
- [x] Test-only application-shell `-O3`/`-O3 -flto` compile-and-run helper (D-21, `TestPhase29OOBApplicationLanes`)
- [x] Stage A seams: Go-injected bytes for the interpreter and a gate-only native driver (D-14)

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| Hosted macOS and Linux cap timings, strict budget | D-12/D-13 | Needs hosted runners | Done: run 38186068052, `gh workflow run ci.yml --ref phase26-validation -f spike008_strict=true`; recorded in `.planning/spikes/008-byteview-cap/README.md` (cap kept at 65,536) |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-10-11

## Validation Audit

| Metric | Count |
|--------|-------|
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
