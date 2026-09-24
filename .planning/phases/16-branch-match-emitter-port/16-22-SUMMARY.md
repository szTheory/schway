---
phase: 16-branch-match-emitter-port
plan: 22
type: execute
status: complete
completed: 2026-09-23
---

# Plan 16-22 Summary: Move Phase 11 N=2 Evidence Fixture

Moved the test-only N=2 source from broad `testdata/phase11/*.lang` discovery to `testdata/phase16/historical/phase11_gate_n_two.fixture`. Updated the Phase 11 test, frozen-evidence external test and manifest, consumer inventory, and provenance registry checks so every consumer follows the moved fixture and verifies its canonical evidence.

The fixture SHA-256 is unchanged: `4bfaf7c7b5116671163a0ca933bf05a8b8bab418c7cb4d5d1d1d30516fcc5bc5`.

## Verification

- Focused Phase 11 gate, frozen-evidence, consumer inventory, provenance mutation, CLI admission, payload replay, and language maturity tests passed.
- A second focused check of frozen evidence, provenance registry, and consumer inventory passed after extending registry digest validation to the relocated historical fixture.
- `rg` found no remaining source or data references to `multi_function_gate_n_two.lang`.

## Deviations

Updated the provenance registry's frozen-record scan to explicitly include the relocated historical fixture. This preserves digest, canonical-program, artifact, and exact-refusal checks after the fixture left the Phase 11 directory.
