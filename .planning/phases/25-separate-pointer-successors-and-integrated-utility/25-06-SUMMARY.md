---
phase: 25-separate-pointer-successors-and-integrated-utility
plan: "06"
subsystem: compiler/native
tags: [Go, C, ownership, evidence, CI, documentation]

requires:
  - phase: 25-05
    provides: "The integrated shared/exclusive pointer utility and independent owner-result proofs."
provides:
  - "A clean-checkout README flow and indexed Phase 25 host/family/lane evidence gate."
  - "One invocation of the Phase 25 evidence gate in the existing dual-host CI aggregate."
  - "Bounded subprocesses in the existing pointer-successor native controls."
  - "Refreshed living-roadmap, maturity, validation, and groundedness records based on observed implementation and checks."
affects: [25-verification, CI, native-evidence, language-roadmap]

actuals:
  tasks: 1
  commits: 1
plan_head_before: f00cdf843bb3b7ed93948a47d3b2acd4af3085af

tech-stack:
  added: []
  patterns:
    - "Native evidence is indexed by family, host, and lane; absent native-host results remain incomplete."
    - "Compiler and application subprocesses use context deadlines and independently bounded output."
    - "Cross-phase regression findings are repaired at their semantic owner and pinned in the existing verification ledger."

key-files:
  created:
    - scripts/verify-phase25.sh
  modified:
    - examples/phase24/README.md
    - .github/workflows/ci.yml
    - internal/compiler/native/phase25_utility_test.go
    - internal/compiler/native/phase25_pointer_successor_test.go
    - internal/compiler/interp/interp.go
    - internal/compiler/check/check_test.go
    - internal/compiler/originvalidate/originvalidate_test.go
    - internal/compiler/pathoracle/pathoracle_test.go
    - internal/compiler/session/session_admission_divergence_test.go
    - internal/compiler/session/verification_groundedness_test.go
    - .planning/phases/25-separate-pointer-successors-and-integrated-utility/25-VALIDATION.md
    - .planning/PRODUCT-ROADMAP.md
    - .planning/LANGUAGE-MATURITY.md
    - .planning/ROADMAP.md

key-decisions:
  - "The local Phase 25 script records successful macOS lanes while explicitly leaving Linux host rows incomplete; local or cross-compiled evidence cannot close EVD-10."
  - "A repository-wide regression gate is required in addition to the focused Plan06 command because integrated pointer admission altered prior Phase 24 and synthetic-fixture expectations."
  - "The reverse shared-to-exclusive reborrow still recomputes its closest access mode as exclusive, while publication independently rejects it as core.borrow_conflict."

requirements-completed: []
coverage:
  - id: D1
    description: "The clean-checkout utility and local family/lane evidence reproduce the success and typed-error behavior with bounded native controls."
    verification:
      - kind: integration
        ref: "sh scripts/verify-phase25.sh on macOS/arm64: foreign/shared/exclusive lanes returned 65/66; typed 0x43 failure occurred before helper calls; both wrong-result controls were reached."
        status: pass
      - kind: contract
        ref: "README, CI wiring, bounded-process and living-roadmap contract tests."
        status: pass
    human_judgment: false
  - id: D2
    description: "Cross-phase regressions and the current local test/build evidence are checked."
    verification:
      - kind: regression
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 ./..."
        status: pass
      - kind: regression
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go test -race -count=1 ./..."
        status: pass
      - kind: static
        ref: "GOCACHE=/tmp/ai-lang-verification-gocache go vet ./...; GOCACHE=/tmp/ai-lang-verification-gocache go build ./..."
        status: pass
  - id: D3
    description: "EVD-10 has complete native receipts for foreign/shared/exclusive families across macOS/Linux and all applicable lanes."
    verification:
      - kind: hosted
        ref: "The current local process produced macOS/arm64 receipts only; Linux and hosted dual-host receipts remain unobserved and incomplete."
        status: incomplete
    human_judgment: false

duration: "~111m"
completed: 2026-10-01
status: complete
---

# Phase 25 Plan 06: Utility Evidence and Closeout Summary

**The clean-checkout utility and fail-closed evidence workflow are implemented, and all local repository gates pass; hosted dual-host receipts remain the final Phase 25 evidence gap.**

## Accomplishments

- Updated the Phase 24 example README with explicit source and binding-manifest inputs, zsh/POSIX-safe commands, expected 65/66 results, typed 0x43 failure behavior, and an evidence index.
- Added `scripts/verify-phase25.sh` and one invocation in the existing `evidence-aggregate` workflow. The script identifies the source/build inputs, compiler, target, flags, expected answers, and native host status for foreign/shared/exclusive families across baseline, optimized, and ASan+UBSan lanes. It reports missing Linux rows as incomplete.
- Bounded compiler and application subprocesses with context deadlines and independent stdout/stderr caps. Native family output retains exact 65/66, typed-error, and reached wrong-result receipts.
- Refreshed PRODUCT-ROADMAP and LANGUAGE-MATURITY from current source boundaries, named tests, local receipts, and historical receipts. Both documents carry the three ranked next capabilities with their runnable programs, blockers, smallest slices, checker changes, evidence/debt, owners, and reprioritization observations.
- Corrected the live corpus and guard census to 151 `.schway` programs, 4,929 lines, and 52 including-test single-function guards.
- Fixed the Phase 24 typed-error interpreter regression so a modeled failure still reaches `OpFail` after cleanup. Updated the stale transfer-order expectation for the new helper calls, repaired the malformed path-oracle fixture, aligned the mixed-access test with independent conflict rejection, and recorded the two intentional Phase 25 admission divergences.
- Marked local Nyquist validation complete while preserving EVD-10 as incomplete until hosted macOS/Linux family-by-lane receipts exist.

## Evidence

The exact Plan06 evidence command passed:

```text
GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 -run '^(TestSourceNeverSpawnsUnboundedProcesses|TestPhase25(EvidenceScript|EvidenceIndex|LivingRoadmap))$' ./internal/compiler/native
sh scripts/verify-phase25.sh
```

On macOS/arm64 with Go 1.24.0 and Apple Clang 21.0.0, all three native families produced 65/66 under `-O0`, `-O2`, and ASan+UBSan. The typed 0x43 control reported `UseError.UnsupportedByte` with zero helper calls; shared and exclusive wrong-result controls were reached; conflict, escape, and pointer-manifest controls passed. Cold Go-cache min/median/max was 2.44/2.72/2.73s; warm-cache min/median/max was 0.27/0.43/0.53s. The script correctly marked the nine Linux family/lane rows incomplete.

The full local checks passed:

- `GOCACHE=/tmp/ai-lang-verification-gocache go test -count=1 ./...`
- `GOCACHE=/tmp/ai-lang-verification-gocache go test -race -count=1 ./...`
- `GOCACHE=/tmp/ai-lang-verification-gocache go vet ./...`
- `GOCACHE=/tmp/ai-lang-verification-gocache go build ./...`
- Focused groundedness, validation lifecycle, maturity census, README, evidence-script, CI-workflow, and living-roadmap contract tests.
- `git diff --check`

The first repository-wide test run found several regressions/stale expectations. They were repaired and the full normal and race suites then passed. No hosted Phase 25 receipt is claimed: this local checkout can produce macOS evidence only, and the CI workflow must run on its native macOS and Linux hosts before EVD-10 closes.

## Remaining Phase Evidence

Plan06 implementation and local validation are complete. Phase 25 remains open until the existing hosted CI aggregate records all foreign/shared/exclusive family results for applicable macOS and Linux lanes. A local run, replay, or cross-compilation does not substitute for those native results.
