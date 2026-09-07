---
phase: 01-canonical-pure-spine
plan: "03"
subsystem: evidence
tags: [protocol, evidence, sha256, cli, verification, performance]
requires: [01-01, 01-02]
provides:
  - "Versioned human/JSON command fact model with stable semantic IDs"
  - "Strict content/tool/policy-bound evidence manifests"
  - "Bounded offline Phase 1 verifier with negative controls and observations"
affects: [ownership, native, ci, agent-feedback, performance]
actuals:
  tasks: 3
  commits: 5
tech-stack:
  added: ["SHA-256 content bindings", "strict encoding/json decoder", "POSIX verification script"]
  patterns: ["one fact model with multiple projections", "metrics outside semantic identity", "fail-closed named verification lanes"]
key-files:
  created: [internal/compiler/protocol/protocol.go, internal/compiler/evidence/evidence.go, internal/compiler/evidence/evidence_test.go, scripts/verify-phase1.sh, testdata/phase1/evidence.golden.json, testdata/phase1/generated.golden.c]
  modified: [cmd/lang/main.go, internal/compiler/session/session.go, internal/compiler/core/core.go, internal/compiler/check/check.go, .planning/phases/01-canonical-pure-spine/01-VALIDATION.md]
key-decisions:
  - "Result IDs bind semantic outcomes and facts but exclude timing, RSS, and output-size observations."
  - "Evidence manifests establish mutual content consistency, not source-to-core derivation proof or signer authority."
  - "Every verifier lane must perform nonzero work and all required negative-control IDs must be observed before green."
  - "Typed core exposes explicit deterministic entry, match, return, and match-edge IDs before ownership CFG work begins."
patterns-established:
  - "Exit 0 pass, 2 invalid input, 3 operational/tool failure, 4 semantic mismatch, 64 usage."
  - "Default command projections remain reproducible; the verifier owns explicit wall-time observations."
requirements-completed: [FND-02, FND-03, DX-01]
coverage:
  - id: D1
    description: "Human and JSON output derive from one command record and share result, diagnostic, evidence, and event identities."
    requirement: DX-01
    verification:
      - kind: black-box
        ref: "internal/compiler/testsupport/cli_test.go#TestCLIOutputContract, TestCLIExitTaxonomy, and TestHumanJSONIdentityParity"
        status: pass
    human_judgment: false
  - id: D2
    description: "Strict evidence rejects every required field mutation, stale content, unknown/trailing JSON, and path leakage."
    requirement: FND-03
    verification:
      - kind: mutation
        ref: "internal/compiler/evidence/evidence_test.go#TestEvidenceMutationMatrix, TestStaleManifest, and TestEvidenceRelocation"
        status: pass
    human_judgment: false
  - id: D3
    description: "The bounded offline verifier performs five named lanes, observes three required controls, and rejects forced engine and manifest faults."
    requirement: FND-02
    verification:
      - kind: end-to-end
        ref: "scripts/verify-phase1.sh and internal/compiler/session/session_test.go#TestVerifyMutationControls"
        status: pass
    human_judgment: false
duration: 16min
completed: 2026-09-03
status: complete
---

# Phase 1 Plan 3: Structured Evidence and Verification Summary

**The real compiler slice now returns stable machine facts, binds its products to strict evidence, and reproduces every Phase 1 claim through one bounded offline command.**

## Performance

- **Duration:** 16 min
- **Started:** 2026-09-03T19:13:24Z
- **Completed:** 2026-09-03T19:29:31Z
- **Tasks:** 3
- **Implementation/audit commits:** 5

## Accomplishments

- Unified human and JSON projections over `lang.command/0`, stable diagnostic/event/result identities, deterministic stream separation, and the explicit exit taxonomy.
- Added `lang.evidence/0` manifests binding canonical source, typed core, generated C, schemas, compiler/Clang/target, ordered flags, and policy.
- Shipped five fail-closed verification lanes with missing-arm, stale-evidence, and interpreter/O0/O3 controls plus nonzero work and operational measurements.
- Measured 20 warm samples for each main path and retained the results and machine/tool provenance in `01-VALIDATION.md`.

## Task Commits

1. **Task 1: Unified command protocol** — `7b89610`
2. **Task 2: Strict reproducible evidence** — `5b0c405`
3. **Task 3: Bounded phase verifier** — `a7f562a`
4. **Completion audit: semantic result identity** — `7989e00`
5. **Completion audit: typed-core point/edge identity** — `be5ac66`

## Decisions Made

- Raw formatter and evidence artifacts remain directly pipeable in human/default mode; `--json` exposes the common command envelope.
- Metrics are observable but cannot change semantic result/evidence identity.
- The verifier recomputes all Phase 1 work and contains no persistent green-result cache.
- Peak RSS remains explicitly unavailable in the sandbox rather than being invented or recorded as zero.

## Deviations from Plan

- The completion audit added two small contract-hardening commits. One made result identity sensitive to semantic outcomes; the other made CFG point/edge identities explicit before Phase 2. Both closed plan requirements rather than expanding language semantics.

## Issues Encountered

- `/usr/bin/time -l` could not read the host clock-rate sysctl inside the sandbox, so RSS measurement was recorded as unavailable with a Phase 6 remeasurement obligation.
- The first command-timing projection made concurrent read-only output nondeterministic. Timing was moved to the explicit verifier observation path while ordinary command output retained stable semantic IDs and work counts.

## User Setup Required

None. Go 1.24 and Clang were already present; the verification script uses a disposable build cache and no network.

## Next Phase Readiness

Phase 2 can now add affine ownership transfer and independent abilities to a real typed core with stable point/edge IDs, an interpreter oracle, native differential execution, strict evidence bindings, and a fast fail-closed feedback loop.

## Self-Check: PASSED

- Fresh `sh scripts/verify-phase1.sh` completed tests, race detection, vet, and all five verifier lanes with exit 0.
- Forced native disagreement maps to exit 4 and forced stale evidence fails nonzero.
- All three Phase 1 plans and every mapped Phase 1 requirement have executable evidence.

---
*Phase: 01-canonical-pure-spine*
*Completed: 2026-09-03*
