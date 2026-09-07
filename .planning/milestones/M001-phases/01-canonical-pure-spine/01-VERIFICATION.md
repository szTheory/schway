---
phase: 01-canonical-pure-spine
status: passed
verified: 2026-09-03
verifier: inline-goal-backward-audit
---

# Phase 1 Verification

## Verdict

**PASSED.** The Phase 1 goal is achieved for the deliberately narrow S1 program:
one canonical nominal ADT transformation travels from source through lossless
syntax and typed core to deterministic interpretation and readable C17, and the
native O0/O3 outcomes agree.

Fresh gate: `sh scripts/verify-phase1.sh` exited 0 after running all Go tests,
the race detector, vet, and `lang verify testdata/phase1`. The verifier reported
five passing lanes and 18 recomputed work units.

## Goal-backward evidence

| Required truth | Evidence | Result |
|---|---|---|
| Clean offline Stage 0 build and root commands | `scripts/verify-phase1.sh`, CLI black-box tests | Pass |
| Lossless canonical syntax and bounded incomplete-edit recovery | CST identity, idempotence, comment, invalid UTF-8, recovery, fuzz-seed, and 1,000-case generated tests | Pass |
| Stable semantic schemas and identities | core declaration/node/point/edge IDs, diagnostic/event/result identity tests | Pass |
| Static exhaustive matching | positive toggle plus `match.non_exhaustive` control | Pass |
| Deterministic semantic oracle | repeated interpreter results and semantic events | Pass |
| Readable native path | committed C17 golden and isolated shell-free Clang runner | Pass |
| Optimizer agreement | interpreter/O0/O3 differential lane and forced-mismatch control | Pass |
| Strict current evidence | golden manifest, every-field mutation matrix, stale/unknown/trailing/relocation controls | Pass |
| Agent/human command contract | common `lang.command/0`, JSON canonicality, identity parity, stream and exit tests | Pass |
| Bounded useful feedback | five nonzero-work lanes, compact output, 20-run observations | Pass |

## Requirement disposition

FND-01..03, SYN-01..04, SEM-01..02, INT-01, NAT-01, and DX-01 are covered.
SYN-01 wording was narrowed to the actual roadmap goal—one named-parameter
match function—because immutable local bindings and calls belong under the
ownership slice rather than this pure tracer.

## Named limitations retained

- Evidence binds compiler products but cannot prove a coordinated frontend made
  the correct source-to-core claim.
- Syntax remains provisional and intentionally tiny.
- Peak RSS could not be observed in the sandbox and is not reported as zero.
- There is no incremental compiler/cache, sanitizer lane, cross-target matrix,
  ownership, resources, effects, concurrency, or runtime in Phase 1.

These are roadmap obligations, not hidden Phase 1 failures.
