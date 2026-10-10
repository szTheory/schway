# Phase 27: Exact FizzBuzz and Cross-Host Evidence - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-10-03
**Phase:** 27-exact-fizzbuzz-and-cross-host-evidence
**Areas discussed:** Source output API, output limit semantics, remainder zero failure, application entry output contract, clean-checkout documentation

The user asked for a broad, adversarial fan-out across relevant technical and stakeholder roles, with tradeoffs, examples, failure modes, established patterns, and useful ecosystem comparisons. The discussion synthesized security, product, compiler/checker, runtime, C17, interpreter, operations, evidence, UX, and documentation perspectives. The user then explicitly adopted all recommendations. Their standing preference is to favor small copy-local implementations and shallow dependency trees, adding dependencies only when a concrete need justifies them.

---

## Source output API

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit bounded writes | Literal-only text, decimal U64, and LF writes; minimal surface for FizzBuzz | ✓ |
| General strings and formatting | Dynamic string values, interpolation, or a formatting abstraction | |
| Broader IO or dependency | Add a generalized IO layer or package for output | |

**User's choice:** “adopt all” — adopt the recommendation for the bounded explicit write surface.
**Notes:** Use `write_text` for compile-time literals, `write_u64` for decimal values, and `write_newline` for LF. Keep general strings, formatting, broad IO, and new dependencies outside Phase 27. Official Rust/Go write APIs were reviewed as comparative examples; the project-specific surface remains intentionally smaller.

---

## Output limit semantics

| Option | Description | Selected |
|--------|-------------|----------|
| Streaming cap | Check each write chunk against 65,536 bytes; a failed later write may leave a prior prefix visible | ✓ |
| Atomic buffered output | Buffer all output and publish only if the whole result fits | |
| Remove or raise the cap | Avoid defining boundary behavior by expanding/removing the existing limit | |

**User's choice:** “adopt all” — adopt the streaming, bounded-write recommendation.
**Notes:** Exceeding the shared output ceiling is a non-success outcome; do not promise stdout rollback. Keep the public `n ≤ 1,000` bound, test exact-limit success and crossing failure with a lowered test-only limit, and treat exit outcome as authoritative if a prefix was already written.

---

## Remainder zero failure

| Option | Description | Selected |
|--------|-------------|----------|
| Defined checked remainder | Add U64 equality/remainder; reject known zero statically and check dynamic zero before native `%` | ✓ |
| Engine-specific/native behavior | Let a dynamic zero reach engine-specific behavior or C undefined behavior | |
| General arithmetic expansion | Add division and unrelated arithmetic along with remainder | |

**User's choice:** “adopt all” — adopt the narrow defined-semantics recommendation.
**Notes:** Use stable reason `U64 remainder by zero`; application failure is status 65 with `schway: U64 remainder by zero\n`. Match interpreter/native behavior. A witness that fails before output has empty stdout; prior output is not undone if a later operation fails. The official C17 draft's zero-remainder rule and Rust/Go checked or runtime behavior were reviewed as comparative references, not as normative Schway specifications.

---

## Application entry output contract

| Option | Description | Selected |
|--------|-------------|----------|
| Explicit no-result entry | Permit Unit/no-result for writing apps and preserve U64 auto-print for `sum_to_n` | ✓ |
| Infer no auto-print from writes | Suppress automatic output if a function happens to write | |
| Make all output explicit | Replace existing U64 return-and-print behavior for every app | |

**User's choice:** “adopt all” — adopt the explicit Unit/no-result entry recommendation.
**Notes:** Do not infer entry behavior from side effects. Keep the existing U64 return contract stable and coordinate checker, interpreter, native app path, and public app behavior.

---

## Clean-checkout documentation

| Option | Description | Selected |
|--------|-------------|----------|
| Phase-specific guide/index | One Phase 27 guide linked from root README, with both programs' commands, results, limits, failures, and host provenance | ✓ |
| Scatter documentation | Duplicate command and expected-output details across multiple existing docs | |
| FizzBuzz only | Document the new witness without making both M005 programs easy to reproduce | |

**User's choice:** “adopt all” — adopt one linked Phase 27 guide/evidence index.
**Notes:** Keep one source of truth for commands; distinguish newly produced macOS/Linux receipts from historical evidence and record an unavailable host as unavailable, not passing.

---

## the agent's Discretion

- Exact source syntax and internal op encoding for the bounded writes, with the literal/decimal/LF scope and cross-engine agreement preserved.
- Focused test-only mechanism for lowering the cap without changing the public output limit.
- Fixture and guide organization consistent with current repository patterns and independent expected answers.

## Deferred Ideas

- General strings, Unicode, dynamic formatting, collections, and broad IO.
- General division and arithmetic beyond the exact FizzBuzz need.
- Buffered/transactional output or rollback guarantees.
- Resource or loan state across loop back edges.
- New dependency or backend without a named consumer and justified need.
