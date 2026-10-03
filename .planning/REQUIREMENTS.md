# Requirements: Schway M005

**Milestone:** Practical Computation
**Defined:** 2026-10-02
**Core Value:** Give an AI agent and a human reviewer the shortest reliable path from intent to sound, reproducible evidence without wasting iteration time or hiding runtime costs.

Research: [M005 synthesis](research/SUMMARY.md). These requirements follow the approved M005 scope and research recommendations. M004's acceptance record is preserved at [M004-REQUIREMENTS.md](milestones/M004-REQUIREMENTS.md).

## M005 Requirements

### U64 semantics

- [x] **U64-01**: A Schway program can add U64 values; an in-range result is exact, and overflow produces a defined checked failure with matching interpreter and native behavior.
- [ ] **U64-02**: U64 comparison, equality, Bool conditions, and remainder have one defined meaning across checking, independent validation, interpretation, and C17 emission; a constant zero remainder divisor is rejected and a dynamic zero divisor produces a defined checked failure.

### Scalar control flow

- [x] **FLOW-01**: A Schway program can use `if/else` and predicate-controlled scalar loops; the checker and independent validators derive admitted loop state through bounded CFG fixed-point analysis over U64 and Bool values.
- [x] **FLOW-02**: Values carrying ownership, resources, loans, or loan-derived provenance across a loop back edge remain explicitly refused with stable, source-attributed diagnostics; analysis bounds fail closed.

### Runnable applications

- [x] **APP-07**: A caller can run `sum_to_n` from ordinary Schway source through the public application route and obtain `0 → 0`, `10 → 55`, and `1,000 → 500,500`.
- [ ] **APP-08**: A caller can run FizzBuzz from ordinary Schway source and obtain exact output for the sequence `1…n`, including fixed labels, decimal values, line breaks, and the final newline.
- [ ] **APP-09**: Both M005 programs accept `0 ≤ n ≤ 1,000`; larger values fail without application output. Application writes remain bounded by the existing 65,536-byte ceiling, and exceeding that ceiling is a non-success outcome.

### Independent evidence and usability

- [ ] **EVD-12**: Interpreter and native results are each compared with independently pinned expected answers, including boundary cases and reached controls for wrong arithmetic, skipped loop iterations, overflow, zero remainder divisors, and repeated loop-event identity; incomplete or capacity-exhausted evidence cannot count as success, and native evidence covers macOS and Linux.
- [ ] **DX-16**: From a clean checkout, a developer can follow documented commands to build and run both M005 programs and locate their expected results and limits.

## Future Requirements

These capabilities are acknowledged but not part of M005 acceptance.

- General arithmetic beyond the addition and remainder needed by the approved programs, including subtraction, multiplication, and general division.
- `for`/iterator forms, `break`/`continue`, arbitrary jumps, and broader loop analysis.
- Ownership, resource, loan, or loan-derived values carried across loop back edges.
- General strings, Unicode APIs, dynamic formatting, collections, and broad IO.
- Additional backends, a runtime, or third-party dependencies without a named consumer that justifies their maintenance and security cost.

## Out of Scope

| Capability | Reason |
|---|---|
| New VM, LLVM backend, or service runtime | The existing interpreter and C17/Clang path support the selected programs. |
| General-purpose string/Unicode and collection runtime | Fixed text and bounded U64 decimal output cover FizzBuzz. |
| Resource and loan state across loop back edges | The initial scalar fixed-point analysis does not prove those obligations. |
| New third-party dependencies | Small local implementations fit the selected scope and preserve the shallow dependency tree. |

## Traceability

Each M005 requirement has one phase owner. Phase 26 implements the
comparison/Bool subset needed by its runnable sum; U64-02 as a complete
requirement closes in Phase 27 with remainder and zero-divisor behavior.
APP-09 and EVD-12 span both programs and close with Phase 27.

| Requirement | Phase | Status |
|---|---|---|
| U64-01 | Phase 26 | Complete |
| U64-02 | Phase 27 | Pending |
| FLOW-01 | Phase 26 | Complete |
| FLOW-02 | Phase 26 | Complete |
| APP-07 | Phase 26 | Complete |
| APP-08 | Phase 27 | Pending |
| APP-09 | Phase 27 | Pending |
| EVD-12 | Phase 27 | Pending |
| DX-16 | Phase 27 | Pending |

**Coverage:** 9/9 requirements assigned to exactly one phase.

---
*Requirements defined: 2026-10-02*
*Last updated: 2026-10-03 after Phase 26 verification; four Phase 26 requirements complete, five Phase 27 requirements pending*
