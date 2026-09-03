---
phase: "01"
slug: "canonical-pure-spine"
status: draft
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-03"
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

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FND-01, DX-01 | T-01-01 | CLI rejects invalid commands and separates tool/source failures | tracer | `go test ./... -run TestCLIToggleTracer` | ❌ W0 | ⬜ pending |
| 01-01-02 | 01 | 1 | SYN-01, SEM-01, SEM-02, INT-01 | T-01-02 | Invalid/missing match alternatives cannot execute | integration | `go test ./... -run 'TestTogglePipeline|TestNonExhaustiveMatch'` | ❌ W0 | ⬜ pending |
| 01-01-03 | 01 | 1 | NAT-01 | T-01-03 | Native tool invocation is explicit and generated C has defined S1 behavior | differential | `go test ./... -run TestNativeToggleO0O3` | ❌ W0 | ⬜ pending |
| 01-02-01 | 02 | 2 | SYN-02, SYN-04 | T-01-04 | Formatter cannot silently delete/reorder user source | property | `go test ./... -run 'TestCSTRoundTrip|TestFormatIdempotent'` | ❌ W0 | ⬜ pending |
| 01-02-02 | 02 | 2 | SYN-03 | T-01-05 | Malformed input makes bounded parser progress | property/fuzz seeds | `go test ./... -run 'TestRecovery|FuzzParseFormat'` | ❌ W0 | ⬜ pending |
| 01-03-01 | 03 | 3 | FND-02, FND-03 | T-01-06 | Evidence bytes exclude incidental paths/timing and reject stale bindings | contract | `go test ./... -run 'TestCanonicalEvidence|TestStaleManifest'` | ❌ W0 | ⬜ pending |
| 01-03-02 | 03 | 3 | DX-01 | T-01-07 | JSON and human projections share IDs; stdout/stderr/exit codes are stable | black-box | `go test ./... -run TestCLIOutputContract` | ❌ W0 | ⬜ pending |
| 01-03-03 | 03 | 3 | FND-01..03, SYN-01..04, SEM-01..02, INT-01, NAT-01, DX-01 | T-01-08 | Full corpus cannot pass when an engine or manifest disagrees | end-to-end | `go run ./cmd/lang verify testdata/phase1` | ❌ W0 | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠ flaky*

## Wave 0 Requirements

- [ ] `go.mod` — establishes the dependency-free Stage 0 module.
- [ ] `internal/compiler/testsupport` — shared in-process and black-box helpers.
- [ ] `testdata/phase1/toggle.lang` — positive canonical tracer fixture.
- [ ] `testdata/phase1/non_exhaustive.lang` — seeded static-rejection control.
- [ ] `testdata/phase1/comments.lang` — lossless CST/formatter fixture.

Wave 1's tracer task creates these before its implementation goes green.

## Manual-Only Verifications

All Phase 1 pass/fail behaviors have automated verification. Generated C and
human diagnostics are committed as golden-readable artifacts, but readability
alone does not decide phase completion.

## Validation Sign-Off

- [ ] All tasks have `<automated>` verification and an explicit failure signal.
- [ ] Sampling continuity: no three consecutive tasks without automated verification.
- [ ] Wave 0 covers all missing paths.
- [ ] No watch-mode or unbounded fuzz flags appear in default commands.
- [ ] Quick feedback latency remains below 10 seconds.
- [ ] `nyquist_compliant: true` is set after implementation evidence exists.

**Approval:** pending execution
