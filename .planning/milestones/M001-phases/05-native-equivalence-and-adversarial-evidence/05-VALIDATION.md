---
phase: "5"
slug: "native-equivalence-and-adversarial-evidence"
# status lifecycle: draft (seeded by plan-phase) → validated (set by validate-phase §6)
# audit-milestone §5.5 distinguishes NOT-VALIDATED (draft) from PARTIAL (validated + nyquist_compliant: false) (#2117)
status: validated
nyquist_compliant: false
wave_0_complete: false
created: "2026-09-05"
evidence_vocabulary: v1
graded_rows: 9
---

# Phase 5 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go's built-in `testing` package (`go test`) — no external test framework |
| **Config file** | none — `go.mod`/`go.sum` govern the build; `scripts/verify-phase5.sh` is the bounded phase gate, not a framework config |
| **Quick run command** | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... <TestName>...` |
| **Full suite command** | `go test ./... && go test -race ./... && go vet ./...` (the exact sequence `scripts/verify-phase4.sh` runs; `verify-phase5.sh` is its peer) |
| **Estimated runtime** | ~30s targeted · ~5-10 min full suite with `-race` |

---

## Sampling Rate

- **After every task commit:** Run `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/<changed-package>/... <relevant tests>`
- **After every plan wave:** Run `go test ./... && go test -race ./... && go vet ./...` plus `lang --json verify testdata/phase{1..5}` for all five corpora (non-regression by re-running prior corpora with the freshly built binary, per the Phase 4 precedent)
- **Mid-phase gate (D-05-40):** `scripts/verify-phase5.sh` must be green after the alias-fact/liveness work and before the reducer/QLT-01 work
- **Before `/gsd-verify-work`:** Full suite green + `scripts/verify-phase5.sh` green
- **Max feedback latency:** 60 seconds for the targeted command

---

## Per-Task Verification Map

> Task IDs are assigned by the planner; this table is the requirement→evidence contract the plans must satisfy. `validate-phase` fills exact task IDs.

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Grade | Non-inertness |
|---|---|---|---|---|---|---|---|---|---|---|
| TBD | TBD | 0 | NAT-02 | — | Differential corpus exists and is enumerable; no engine disagreement goes unreported | differential (execution) | `lang --json verify testdata/phase5` | ❌ W0 — `testdata/phase5/` does not exist | DEFINED | — |
| TBD | TBD | ≥1 | NAT-02 | — | Interpreter / `-O0` / `-O3` / `-O3+LTO` agree on terminal outcome, semantic-event order, live-resource state | differential (execution) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestPhase5CorpusEngineAgreement` | ❌ W0 | WIRED | — |
| TBD | TBD | ≥1 | NAT-03 | T-5-mutation | Each of the seven native hostile mutations is detected by its intended independent lane, not incidentally | mutation-kill (unit + differential) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... <Test*MutationIsDetected>` | ❌ W0 — mutation runners not written | WIRED | — |
| TBD | TBD | ≥1 | NAT-03 | T-5-sanitizer | ASan/UBSan lane catches retained-pointer, allocator-mismatch and use-after-free; its evidence is isolated from semantic-equivalence evidence | sanitizer (process exit + diagnostic substring) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... <TestSanitizerLane*>` | ❌ W0 — sanitizer build config absent from `native.go` | WIRED | — |
| TBD | TBD | ≥1 | INT-02 | — | On an injected mismatch the reducer emits a minimized source/core case plus a causal event trace | reducer (unit; 3 required mutation-kills: no-op reducer, too-loose predicate, non-determinism) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... <Test*Reducer*>` | ❌ W0 — reducer package does not exist | WIRED | — |
| TBD | TBD | ≥1 | INT-02 | — | `lang.mismatch/0` round-trips and binds to the evidence it minimizes | schema round-trip (unit) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... <TestMismatchDocumentRoundTrips>` | ❌ W0 | WIRED | — |
| TBD | TBD | ≥1 | QLT-01 | — | Every spike control has a live descendant or an explicit waiver; no stale `control:` reference survives | registry audit (unit) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/... TestQLT01RegistryComplete` | ❌ W0 — registry does not exist | EXERCISED | — |
| TBD | TBD | final | Phase-5-wide | — | `verify-phase5.sh`'s control set equals `Phase5RequiredControls()` — the gate cannot silently shrink | equality test (unit) | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/session TestPhase5RequiredControlsMatchScript TestVerifyPhase5ControlsAndWork` | ❌ W0 | EXERCISED | — |
| TBD | TBD | ≥1 | `discoverLoanLastUses` retirement | — | Zero divergence between the old scanner and `loanLivenessFixpoint` over the full bounded enumeration | shadow-mode divergence logging | `env GOCACHE=/tmp/ai-lang-phase5-cache sh scripts/assert-go-tests.sh ./internal/compiler/check TestOwnershipSequenceExhaustive TestBranchSequenceExhaustive` | ❌ W0 — shadow-mode harness not added | EXERCISED | — |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

- [ ] `testdata/phase5/` — hand-written adversarial subset (6–10 programs) plus the bounded enumerated closure generator — covers NAT-02
- [ ] New mutation-runner tests for the alias-fact, second-allocator and use-after-free hostile fixtures — covers NAT-03
- [ ] Sanitizer-lane build configuration in `native.go` plus its own test harness — covers NAT-03
- [ ] New reducer package plus its three required mutation-kill tests — covers INT-02
- [ ] `lang.mismatch/0` schema type plus round-trip test — covers INT-02
- [ ] QLT-01 registry file/table plus `TestQLT01RegistryComplete` — covers QLT-01
- [ ] `scripts/verify-phase5.sh` plus `TestPhase5RequiredControlsMatchScript` — covers the phase-wide gate
- [ ] Shadow-mode divergence-logging harness over `TestOwnershipSequenceExhaustive` / `TestBranchSequenceExhaustive` — covers the `discoverLoanLastUses` retirement precondition

---

## Manual-Only Verifications

| Behavior | Requirement | Why Manual | Test Instructions |
|----------|-------------|------------|-------------------|
| The coordinated source-to-core false claim remains an *escape*, not a solved problem | INT-02 (SC4) | Proving absence of a coordinated two-artifact lie is out of reach for this milestone; the phase records it rather than engineering around it | Read the Phase 5 evidence narrative and confirm the escape is named explicitly as a known blind spot, not claimed as covered |
| `alloc_dealloc_mismatch` defaults-off on macOS (D-05-08) | NAT-03 | The research falsification attempt was inconclusive (compatible allocator pairing); the claim stays `[CITED]`, not `[VERIFIED]` | If a mismatch fixture does not trip ASan on the dev host, confirm the flag is explicitly enabled in the sanitizer lane rather than assuming the default |

---

## Validation Sign-Off

- [ ] All tasks have `<automated>` verify or Wave 0 dependencies
- [ ] Sampling continuity: no 3 consecutive tasks without automated verify
- [ ] Wave 0 covers all MISSING references
- [ ] No watch-mode flags
- [ ] Feedback latency < 60s for the targeted command
- [ ] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
