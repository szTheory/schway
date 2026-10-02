# Domain Pitfalls

**Domain:** Practical computation in a small compiler and native toolchain (M005: bounded `sum_to_n`, then FizzBuzz)
**Researched:** 2026-10-02
**Overall confidence:** HIGH for current Schway boundaries (source inspection); MEDIUM for external semantic comparisons (official primary docs cross-checked; provider confidence tier is MEDIUM).

## Critical Pitfalls

### Pitfall 1: Schway arithmetic accidentally inherits C promotions or undefined behavior
**What goes wrong:** A seemingly direct C expression can use a different width or signedness than Schway intended. C applies integer promotions/usual arithmetic conversions; signed overflow, integer divide-by-zero, and unrepresentable quotient/remainder cases can be undefined. A C compiler may then optimize based on assumptions Schway did not intend. Merely spelling a temporary `uint64_t` does not make every mixed-type expression correct.
**Why it happens:** The interpreter is hosted in Go, whose unsigned arithmetic wraps modulo 2^n and whose signed overflow is defined; C and Go therefore cannot serve as each other's semantic definition. LLVM's `nsw` overflow marker yields poison, which can feed optimizer transformations.
**Consequences:** Interpreter/native disagreement, target-dependent results, or optimized native behavior that differs from an O0-looking result. FizzBuzz modulo tests may appear correct for small inputs while edge inputs violate the intended contract.
**Prevention:** Before admitting operators, define typed Schway rules for `U64` add/subtract/multiply/divide/remainder/comparison and overflow/divide-by-zero outcomes. Lower exact-width operations with explicitly typed `uint64_t` operands and results; make any required checks explicit. Do not add `nsw`/`nuw`, signed temporaries, or unchecked mixed-width C expressions unless the checker establishes the corresponding fact. Pick one overflow policy and make it build-mode independent.
**Detection:** Test carry/borrow and maximum-value boundaries in the interpreter and in generated C at baseline and optimized levels. Include zero-divisor rejection and remainder cases. Require an independently authored expected result and a control that changes the arithmetic rule or emitted operand type and is rejected.
**Schway evidence:** Current source inspection finds U64 constants, not arithmetic (`internal/compiler/session/session_phase19_test.go`, `internal/compiler/interp/interp_phase19_test.go`, `internal/compiler/cgen/cgen_program_test.go`). `internal/compiler/cgen/cgen_program.go` already checks exact-width U64 support; retain that portability check. No checks were run for this research task.
**External evidence (MEDIUM):** WG14's C23 draft specifies the usual arithmetic conversions and division/remainder behavior in §§6.3.1.8 and 6.5.5; the [Go specification](https://go.dev/ref/spec#Integer_overflow) defines unsigned wrap and deterministic signed overflow; [LLVM's UB manual](https://llvm.org/docs/UndefinedBehavior.html) explains optimizer constraints and `nsw` poison.

### Pitfall 2: “Loop support” lands without correct loop-carried facts, termination bounds, or resource refusals
**What goes wrong:** A compiler accepts a source loop but computes the wrong fixed point for values or loans at the back edge. Typical misses include a stale initial value, a branch that skips or repeats an iteration, an exit state that is joined incorrectly, and ownership/loan state that appears discharged on one pass but remains live on another. Separately, a syntactically valid loop may not terminate or may take an impractical number of iterations.
**Why it happens:** A loop is a cyclic CFG, so one-pass traversal is not enough for dataflow. Existing acyclic path enumeration cannot simply be reused as if it represented an unbounded iteration. Resource identity and last-use facts that are straightforward in acyclic paths need an explicit loop contract.
**Consequences:** Off-by-one sums, a FizzBuzz row missing/duplicated at a boundary, compiler nontermination, unsound borrow/ownership admission, or a command that appears hung on a bounded U64 input.
**Prevention:** Implement scalar CFG fixed-point analysis with a finite lattice, deterministic worklist ordering, and a derived fail-closed iteration/work bound. Initially refuse loans and live resources crossing loop back edges; do not weaken the refusal merely to get the resource-free examples compiling. Define the app-level input/iteration bound and its failure behavior so accepted programs have an operationally bounded run.
**Detection:** Cover zero, one, the exact loop bound, just-below/at/just-above boundaries, and a deliberately altered induction update/exit predicate. Mutation controls must prove that changed back-edge state or skipped iteration is detected. Add a resource/loan-carrying loop negative control that reaches the refusal; retain the existing refused loop fixture as a moved-frontier check.
**Schway evidence:** `internal/compiler/check/check.go` contains `check.cfg_back_edge` refusal; `internal/compiler/pathoracle/pathoracle.go` independently refuses CFG back edges; `internal/compiler/session/session_phase20_test.go:TestPhase20ChecksumFrontier` pins the provisional checksum loop refusal. Existing acyclic loan worklist precedent is `loanLivenessFixpoint` in `check.go` with controls in `check/check_test.go`, but that is not evidence of loop-carried ownership correctness. M005 kickoff in `.planning/LANGUAGE-MATURITY.md` explicitly says no project checks were run.

### Pitfall 3: Agreement between engines is mistaken for semantic correctness
**What goes wrong:** The interpreter and native output match because both implement the same mistaken operator, loop bound, overflow rule, or expected-output generator. A comparison harness can certify shared error, especially when the reference answer is generated from the interpreter or fixture under test.
**Why it happens:** Differential agreement answers “do these implementations agree?” not “is this the specified answer?” Reusing a producer to compute the expectation defeats the independent oracle.
**Consequences:** A regression is certified as conformance, then becomes much harder to correct once programs rely on the accidental behavior.
**Prevention:** For each runnable program, store hand-derived expected outputs for selected boundary inputs or compute them in a genuinely separate, simple reference implementation. Compare interpreter and native O0/optimized outputs to that oracle, and compare process status/stderr separately. Keep oracle semantics explicit (including overflow and zero-divisor behavior).
**Detection:** Seed wrong-result, wrong-operator, incorrect initial-state, skipped-final-iteration, and wrong-exit-status mutations. Require each declared control to reach the mutated path and fail for the intended axis. A test that only checks engine equality is insufficient.
**Schway evidence:** M003's evidence policy grades exercised behavior and seeded-fault proofs; Phase 14 corrected multiple “green because wired” defects. M004's `EVD-11` likewise requires independent expected answers. Existing examples include `TestPhase22AppVerifyIndependentIdentityCases` and explicit Phase 25 wrong-result controls, but neither proves new arithmetic/loop behavior.

## Moderate Pitfalls

### Pitfall 4: Remainder, division, or comparison signs diverge across engines
**What goes wrong:** `%` for negative operands, quotient rounding, zero divisors, and signed/unsigned comparisons differ from the eventual Schway rule. C truncates integer division toward zero and leaves a zero divisor undefined; a remainder expression has the same divisor restriction. Go and other languages make their own choices. FizzBuzz's positive constants do not exercise all those edges.
**Prevention:** Keep M005's first integer domain explicitly `U64` if that is all the program needs. Specify `/` and `%` by type and specify divisor zero as a checked Schway outcome, never as emitted C UB. If signed arithmetic is not needed by the selected witness, refuse it rather than silently importing host behavior.
**Detection:** Add a source-level zero-divisor negative control and boundary operands. If signed types enter scope, include negative dividend/divisor cases and `MIN / -1` / corresponding remainder behavior.
**Phase/topic:** Phase 26 arithmetic and checker contract; defer signed arithmetic unless a named consumer needs it.
**External evidence (MEDIUM):** [WG14 N3088, §§6.5.5–6.5.6](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n3088.pdf); [Rust Reference integer-overflow section](https://doc.rust-lang.org/reference/behavior-not-considered-unsafe.html#integer-overflow) documents a different build-sensitive policy, illustrating why Schway must state its own.

### Pitfall 5: C promotion, literal, and format-string details make a portable U64 implementation host-specific
**What goes wrong:** A literal or intermediate is treated as signed or narrower than intended, or a decimal formatter uses a mismatched C variadic format. One host happens to have the expected widths and reports correct output; another target or compiler warns, truncates, or misbehaves.
**Prevention:** Carry the Schway type through core validation and emission; use `<stdint.h>` exact-width types and matching constants/formatting strategy, or a small local decimal writer that has its own boundary proof. Do not depend on host `int` width, signed casts, implicit promotions, or unverified libc formatting assumptions. Preserve the existing exact-width target rejection.
**Detection:** Native evidence should name target and compiler on macOS and Linux, test `0`, the maximum accepted value, and decimal digit transitions. Inspect generated C and compile with warnings enabled; sanitizer success alone is not a portability argument.
**Phase/topic:** Phase 26 typed core, C emission, decimal output, and native host matrix.

### Pitfall 6: Bounded U64 input is confused with bounded execution or successful evidence
**What goes wrong:** The application input parser bounds a token, but an accepted loop still performs billions of iterations. Separately, evidence capture can hit a capacity limit, and the result is mistaken for an application failure or success. Compiler/report output truncation can hide missing output or a missing final event.
**Why it happens:** `schway app run` and the older `schway run` have distinct roles: the app route launches once and preserves ordinary process streams; the conformance route is synthetic-input replay. Event/evidence capacity is a separate bounded channel from stdout/stderr. The repository has historically fixed a mismatch where admission used a 4,096-occurrence bound but native execution output still used a smaller bound.
**Prevention:** Put an explicit iteration maximum on the accepted `sum_to_n` contract (and define rejection before execution if exceeded). Use the app route for the user-visible program. Preserve the separation between application stdout/stderr/status and evidence status; `capacity_exhausted`, incomplete, disabled, write failure, and truncated streams are not semantic program results and cannot be treated as verified success.
**Detection:** Exercise the highest accepted input and one rejected input; test ordinary output at and around documented limits; confirm that output truncation is an error and evidence exhaustion leaves `verified=false`. Keep exact expected program output independent from execution-report JSON.
**Schway evidence:** Phase 22 tests `TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, `TestPhase22EvidenceDisabledCompleteAndStreamIsolation`, and the app evidence capacity controls in `internal/compiler/native/native_app_test.go`. The historical `phase15-native-capacity` resolved record is a concrete warning that admission bounds and runtime-output bounds can diverge.

### Pitfall 7: Output is judged by a prefix, formatting convention, or CLI wrapper rather than the program contract
**What goes wrong:** FizzBuzz is semantically “working” but has wrong separators, newline policy, final newline, labels, decimal conversion, missing last item, or truncated output. Tests parse/normalize output too generously and pass. A stream cap or evidence envelope contaminates ordinary output.
**Prevention:** Freeze byte-for-byte expected stdout, stderr, and exit behavior for `sum_to_n` and FizzBuzz. Define which command accepts input and how invalid/oversized values fail. Keep evidence/report output on its documented separate channel. Prefer a narrow bounded output primitive and explicit decimal conversion over a general string/runtime subsystem.
**Detection:** Compare exact bytes, including empty and final-line cases; use at least one negative control for dropped final newline/line and one for premature truncation. Test both the retained app command and compiler evidence path according to their distinct contracts.
**Phase/topic:** Phase 26 product route and output/decimal-format slice.

### Pitfall 8: A sanitizer lane is treated as the arithmetic specification
**What goes wrong:** No UBSan finding is interpreted as proof that arithmetic is correct. But Clang's `undefined` group excludes unsigned overflow by default, while the `integer` group adds unsigned overflow and suspicious implicit conversions. Sanitizers also only witness exercised paths and their behavior can vary by lane.
**Prevention:** Declare semantics in Schway and independently compare answers. Use UBSan and, where supported, integer sanitizers as additional defect detectors; do not infer wrapping, checked overflow, or correctness from a quiet sanitizer run. Do not make acceptance depend on toolchain-specific diagnostic wording.
**Detection:** Include edge inputs that actually execute overflow and zero-divisor branches, and ensure the selected sanitizer flags cover the intended check. Keep a semantic expected-answer check even when sanitizer lanes pass.
**Phase/topic:** Phase 26 native optimizer/sanitizer evidence planning.
**External evidence (MEDIUM):** [Clang UBSan documentation](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks) explicitly distinguishes unsigned overflow and implicit conversion checks from the default `undefined` group.

## Minor Pitfalls

### Pitfall 9: Enabling a dependency or runtime to solve a tiny first arithmetic slice
**What goes wrong:** A big-integer, formatting, testing, or runtime dependency becomes an unaudited semantic boundary and complicates bootstrap/portability for a bounded U64 witness.
**Prevention:** Keep the existing Go 1.24 standard-library compiler/interpreter and readable C17/Clang path. Use a small copy-local checked arithmetic or decimal formatting helper only if the chosen implementation needs one; record a concrete consumer before introducing a dependency.
**Priority change:** Reconsider only when a named program needs arbitrary precision, locale/Unicode formatting, or a portability primitive that cannot be implemented and tested locally at lower lifetime cost.

### Pitfall 10: The old refusal test is retained but no longer proves the new frontier
**What goes wrong:** `TestPhase20ChecksumFrontier` continues passing after `loop` parses, because it checks an obsolete diagnostic, or it is weakened to accept any failure. That neither demonstrates computation nor pins the intentionally refused shapes.
**Prevention:** Move the positive witness to a real app command with exact expected results; separately rewrite refusal controls to target concrete unsupported cases (resource/loan across back edge, unsupported signed operation, out-of-bound input). Assert stable diagnostic identity and source attribution, not just “some error.”
**Phase/topic:** Phase 26 parser/checker frontier and refusal diagnostics.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| Phase 26: typed U64 arithmetic and comparisons | C/Go semantics drift, C promotions, optimizer assumptions | Ratify semantics first; exact-width lowering; independent boundary oracle; O0 and optimized evidence plus reached mutations |
| Phase 26: scalar loop fixed point | Off-by-one, stale loop-carried values, analyzer nontermination | Finite monotone lattice, deterministic bounded worklist, exact 0/1/max cases, mutated induction/exit control |
| Phase 26: ownership/loan interaction | Existing acyclic proofs accidentally applied across back edges; leaks or unsound loans | Explicitly refuse live resources/loans across back edges; test that refusal is reached; only widen after independent fixed-point ownership evidence |
| Phase 26: `sum_to_n` public route | Bounded token mistaken for practical execution bound; model route mistaken for app behavior | State accepted N maximum, exact error behavior; use one-launch `schway app run`; report stdout/stderr/status independently |
| Phase 26: FizzBuzz output and remainder | Happy-path-only `%` test; formatting or last-row truncation | Byte-exact expected stdout and correct result for zero and boundary values; test invalid divisor behavior and line-ending cases |
| Native evidence/portability | macOS arm64 green is assumed to prove Linux; UBSan is assumed exhaustive | Name both hosts/compiler/flags; preserve per-lane manifests; distinguish historical receipts from newly executed checks; do not run tests as part of this research task |

## What Would Change Priority

The current product scale is intentionally a small, bounded native language slice: Phase 26 should not acquire generic collections, a new backend, a runtime, a generalized effect system, or broad resource semantics to run two scalar programs. Raise priority for loop-carried resource/loan support only when a named runnable consumer requires it and a concrete ownership witness defines safe iteration/exit behavior. Raise priority for signed arithmetic, overflow modes, arbitrary precision, or broader text support only when the accepted program contract demonstrates that requirement. A new dependency is justified only if a concrete use case plus portability evidence beats a small audited local boundary.

This report records source inspection of the current tree and historical receipts cited by the living documents; it does not claim new test or native execution evidence.

## Sources

- Schway source/refusal witnesses: `internal/compiler/check/check.go` (`check.cfg_back_edge`, `loanLivenessFixpoint`), `internal/compiler/pathoracle/pathoracle.go` (`pathoracle.cfg_back_edge`), `internal/compiler/session/session_phase20_test.go` (`TestPhase20ChecksumFrontier`), `internal/compiler/session/session_phase19_test.go`, `internal/compiler/interp/interp_phase19_test.go`, and `internal/compiler/cgen/cgen_program_test.go` (`TestPhase19U64NativeExactWidthAndOutput`).
- Schway evidence and output boundaries: `internal/compiler/native/native_app_test.go` (`TestPhase22RunApplicationPreservesStreamsAndProcessOutcomes`, evidence stream/capacity controls), `.planning/milestones/M003-MILESTONE-AUDIT.md` (historical evidence-instrument findings), `.planning/debug/resolved/phase15-native-capacity.md` (historical admission/output-bound mismatch), `.planning/REQUIREMENTS.md` (M004 EVD-11 separation and NAT-09 refusal ownership), and `.planning/LANGUAGE-MATURITY.md` (2026-10-02 M005 kickoff and current frontier).
- ISO/IEC JTC 1/SC 22/WG 14, [N3088: ISO C23 draft](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n3088.pdf), §§6.3.1.8, 6.5.5–6.5.8.
- The Go Authors, [The Go Programming Language Specification: Integer overflow](https://go.dev/ref/spec#Integer_overflow).
- The Rust Project, [Rust Reference: Behavior not considered unsafe, integer overflow](https://doc.rust-lang.org/reference/behavior-not-considered-unsafe.html#integer-overflow).
- LLVM Project, [LLVM IR Undefined Behavior Manual](https://llvm.org/docs/UndefinedBehavior.html).
- LLVM Project, [Clang UndefinedBehaviorSanitizer](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html).
