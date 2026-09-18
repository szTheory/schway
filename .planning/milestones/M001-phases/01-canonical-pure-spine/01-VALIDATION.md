---
phase: "01"
slug: "canonical-pure-spine"
status: complete
nyquist_compliant: true
wave_0_complete: true
created: "2026-09-03"
evidence_vocabulary: v1
graded_rows: 8
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go 1.24 `testing`, `testing/quick`, native fuzzing, black-box `os/exec` |
| **Config file** | `go.mod` (created in Wave 1) |
| **Quick run command** | `go test ./...` |
| **Full suite command** | `go test -race ./... && go vet ./... && go run ./cmd/lang verify testdata/phase1` |
| **Estimated runtime** | Quick < 10 seconds; full < 30 seconds on the declared development Mac |

## Sampling Rate

- **After every task commit:** Run `go test ./...`
- **After every plan wave:** Run `go test -race ./... && go vet ./...`
- **Before `$gsd-verify-work`:** Full suite and black-box CLI contract must be green
- **Max feedback latency:** 10 seconds for the quick suite

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| 01-01-01 | 01 | 1 | FND-01, DX-01 | T-01-01 | CLI rejects invalid commands and separates tool/source failures | tracer | `go test ./... -run TestCLIToggleTracer` | ✅ | EXERCISED | — |
| 01-01-02 | 01 | 1 | SYN-01, SEM-01, SEM-02, INT-01 | T-01-02 | Invalid/missing match alternatives cannot execute | integration | `go test ./... -run 'TestTogglePipeline|TestNonExhaustiveMatch'` | ✅ | EXERCISED | — |
| 01-01-03 | 01 | 1 | NAT-01 | T-01-03 | Native tool invocation is explicit and generated C has defined S1 behavior | differential | `go test ./... -run TestNativeToggleO0O3` | ✅ | EXERCISED | — |
| 01-02-01 | 02 | 2 | SYN-02, SYN-04 | T-01-04 | Formatter cannot silently delete/reorder user source | property | `go test ./... -run 'TestCSTRoundTrip|TestFormatIdempotent'` | ✅ | EXERCISED | — |
| 01-02-02 | 02 | 2 | SYN-03 | Malformed input makes bounded parser progress | property/fuzz seeds | `go test ./... -run 'TestRecovery|FuzzParseFormat'` | ✅ | DEFINED | — |
| 01-03-01 | 03 | 3 | FND-02, FND-03 | Evidence bytes exclude incidental paths/timing and reject stale bindings | contract | `go test ./... -run 'TestCanonicalEvidence|TestStaleManifest'` | ✅ | DEFINED | — |
| 01-03-02 | 03 | 3 | DX-01 | JSON and human projections share IDs; stdout/stderr/exit codes are stable | black-box | `go test ./... -run TestCLIOutputContract` | ✅ | DEFINED | — |
| 01-03-03 | 03 | 3 | FND-01..03, SYN-01..04, SEM-01..02, INT-01, NAT-01, DX-01 | T-01-08 | Full corpus cannot pass when an engine or manifest disagrees | end-to-end | `go run ./cmd/lang verify testdata/phase1` | ✅ | REACHABLE | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠ flaky*

## Wave 0 Requirements

- [x] `go.mod` — establishes the dependency-free Stage 0 module.
- [x] `internal/compiler/testsupport` — shared in-process and black-box helpers.
- [x] `testdata/phase1/toggle.lang` — positive canonical tracer fixture.
- [x] `testdata/phase1/non_exhaustive.lang` — seeded static-rejection control.
- [x] `testdata/phase1/comments.lang` — lossless CST/formatter fixture.

Wave 1's tracer task creates these before its implementation goes green.

## Manual-Only Verifications

All Phase 1 pass/fail behaviors have automated verification. Generated C and
human diagnostics are committed as golden-readable artifacts, but readability
alone does not decide phase completion.

## Executed observations

Measured 2026-09-03 on Darwin 25.6.0 arm64 with Go 1.24.0 and Apple Clang
21.0.0. Each direct command used one already-built binary, one fixture, serial
execution, and 20 warm process-level samples. Values include process startup.

| Command | warm p50 | warm p95 | min–max | JSON bytes | recomputed work |
|---------|---------:|---------:|--------:|-----------:|----------------:|
| `format` | 2.453 ms | 3.089 ms | 2.121–3.752 ms | 470 | 1 |
| `check` | 2.120 ms | 2.603 ms | 1.880–2.662 ms | 299 | 1 |
| interpreter run | 1.995 ms | 2.291 ms | 1.826–2.321 ms | 823 | 2 |
| native O0/O3 run | 353.967 ms | 381.572 ms | 333.884–388.972 ms | 823 | 6 |
| full `verify` | 373.058 ms | 400.813 ms | 333.779–424.643 ms | 1,428 | 18 |

The repository gate (`sh scripts/verify-phase1.sh`) completed in 11.9 seconds
with tests, race detection, vet, and all five verifier lanes green. These
observations are below D-12's provisional 100 ms warm frontend/interpreter and
1 second native investigation thresholds on this declared machine. They are
not portable release budgets. Peak RSS is recorded as `unavailable`, not zero:
the sandbox denied the host `sysctl kern.clockrate` query required by
`/usr/bin/time -l`; Phase 6 must repeat RSS measurement on an unrestricted
declared runner before ratification.

## Validation Sign-Off

- [x] All tasks have `<automated>` verification and an explicit failure signal.
- [x] Sampling continuity: no three consecutive tasks without automated verification.
- [x] Wave 0 covers all missing paths.
- [x] No watch-mode or unbounded fuzz flags appear in default commands.
- [x] Quick feedback latency remains below 10 seconds.
- [x] `nyquist_compliant: true` is set after implementation evidence exists.

**Approval:** current Phase 1 evidence green; portable performance budgets remain deferred to Phase 6.
