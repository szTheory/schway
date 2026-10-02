# Project Research Summary

**Project:** Schway — M005 Practical Computation  
**Domain:** General-purpose systems programming language and native toolchain; bounded scalar computation  
**Researched:** 2026-10-02  
**Confidence:** MEDIUM

## Executive Summary

Schway is extending an existing checked source-to-native language spine so ordinary source can compute and produce useful bounded output. The research strongly recommends growing the current typed core, checker, independent validators, interpreter, C17 emitter, and application runner in place. `sum_to_n` should be the first runnable vertical slice; FizzBuzz should then prove that the same scalar loop and output path composes comparisons, remainder, branches, and exact text. Keep the source surface to a predicate-controlled loop and the conditional form these examples need, and preserve explicit refusal for owner/loan state carried across back edges.

The major contract choice still needing requirements-stage ratification is U64 overflow. The stack and architecture reports recommend explicit modulo-2^64 arithmetic; the feature and pitfalls reports recommend checked failure to avoid a plausible but incorrect sum. For the approved examples, checked overflow with a defined cross-engine error is the safer default recommendation, while explicit wrapping remains a coherent alternative only if the user wants bit-vector-style arithmetic. In either case, define the rule before lowering, guard dynamic division/remainder by zero before C, bound input/work/output, and use independently authored expected answers plus reached negative controls. No new dependency, backend, or general runtime is justified.

## Key Findings

### Recommended Stack

Retain Go 1.24 and its standard library for Stage 0, the existing interpreter as an independently implemented semantic engine, the single readable C17 emitter with installed Clang for native applications, and the current bounded application-output route. Implement small fixed-width arithmetic and decimal-output helpers locally. Keep macOS and Linux as explicit native execution lanes; target selection alone does not establish that an unavailable target sysroot or linker works.

**Core technologies:**
- Go 1.24 standard library: host compiler, analysis, interpreter, CLI, and evidence orchestration; no M005 package need exceeds a small local implementation.
- Typed Schway core plus independent validators and path oracle: portable semantic contract and protection against forged or inconsistent core facts.
- C17 `uint64_t` through installed Clang: existing production native route; preserve exact-width checks and avoid signed or implicitly promoted intermediates.
- Bounded local decimal writer and fixed byte literals: enough for `sum_to_n` and FizzBuzz without a string runtime or formatting dependency.

**Semantic decision requiring requirements/user ratification:**
- **Recommended default:** checked U64 add/subtract/multiply failure, with the same defined outcome in interpreter and native code. It avoids a silently wrapped `sum_to_n` result. The explicit user requirement is semantic equivalence, not yet a chosen overflow policy.
- **Alternative:** modulo-2^64 U64 arithmetic. This is portable and maps to Go/C unsigned arithmetic, but must be deliberately specified and independently tested; never inherit it accidentally.
- **Required whichever policy is selected:** U64-only first slice; compile-time zero divisor rejected; dynamic `/` or `%` by zero yields a defined checked failure and never reaches C UB. If division itself is outside the witness, admit only `%` needed by FizzBuzz.

### Expected Features

M005's confirmed product outcome is a caller-run `sum_to_n` and then FizzBuzz from ordinary Schway source with exact output. Research recommendations refine how narrowly to deliver it; they do not themselves select every semantic detail or numeric cap.

**Must have (table stakes):**
- Typed U64 addition, comparison/equality, Bool conditions, and the selected overflow contract.
- A narrow condition-controlled scalar loop with correct CFG fixed-point analysis; initially only scalar U64/Bool state crosses back edges.
- U64 remainder with a specified zero-divisor outcome for FizzBuzz.
- Bounded fixed literal output and decimal U64 formatting, with exact stdout/stderr/exit behavior on the ordinary app route.
- Independently pinned expected results for interpreter and native execution, plus boundary cases and negative controls that reach the changed assumption.

**Should have (differentiating):**
- Ship a working `sum_to_n` vertical slice before the later FizzBuzz composition, so the milestone gets an early runnable gain.
- Explicit input/work/output bounds, and failure before partial or misleading success where the contract can do so.
- Cross-consumer derivation in checker, core validator, origin validator, path oracle, interpreter, and C emitter, with separate evidence validation for repeated dynamic events if event-producing operations execute repeatedly.

**Defer:**
- General `for`/iterator forms, `break`/`continue`, labels, arbitrary jumps, nested-loop guarantees, and resource/loan carry across a loop back edge.
- Signed arithmetic, general division unless required, arbitrary precision, general strings/Unicode/formatting, generic collections, broad IO, and dynamic output allocation.
- A new backend, runtime, testing/property/SMT dependency, or third-party helper without a concrete consumer that beats a small audited local boundary.

### Architecture Approach

Preserve the route `.schway` source → lossless AST → checker-produced typed core and CFG → independent core/origin/path validation → interpreter and C17/Clang → session comparison against a separate expected oracle. Express `while` and `if/else` in readable source, lowering them into existing `core.Block`/`core.Edge` control flow. Fixed-point analysis belongs in the semantic validation path; loop eligibility must not be “proved” by simply walking cycles with the current acyclic path enumeration. Keep application streams distinct from execution evidence, and never treat evidence exhaustion as application semantics or successful verification.

**Major components:**
1. `syntax`/`ast`: only the operator, loop, conditional, and narrow output forms required by the witnesses, with stable formatting and source spans.
2. `check` and `core`: derive typed operations, Bool branch facts, CFG, scalar fixed points, and explicit loop-state refusals.
3. `corevalidate`, `originvalidate`, and `pathoracle`: independently rederive core/type/provenance and admitted-loop facts; bound analysis and fail closed on unsupported resource/loan cycles.
4. `interp` and `cgen`: independently implement the ratified U64/loop/output semantics; C uses exact-width unsigned values and defined operations only.
5. `executionpeer`, `execution`, `session`, and `native`: validate repeated-event identities where needed, keep reports separate from app output, and compare both engines to independent expected bytes/results.

Evidence/inspection status: researchers inspected source and planning documents. They ran no new tests or native checks. M003/M004 and Phase 25 receipts cited below are historical evidence for existing infrastructure, not proof of the new arithmetic or loop behavior.

### Critical Pitfalls

1. **Host C/Go behavior becomes the language definition** — ratify typed U64 semantics first; use exact-width operations, no implicit signed conversions or optimizer assumptions, and define zero-divisor behavior before emission.
2. **Cycle acceptance skips fixed-point or resource proofs** — use a finite monotone scalar worklist with deterministic/fail-closed bounds; continue refusing owners, resources, loans, and loan-derived values over back edges.
3. **Interpreter/native agreement masks shared error** — pin hand-derived outputs/boundary outcomes independently and include reached wrong-result, changed-assumption, skipped-iteration, and output-truncation controls.
4. **Input token bounds are mistaken for bounded execution/output** — requirements must choose accepted `n`, iteration/work limit, output byte cap, and failure/partial-write semantics. Keep tool evidence capacity separate from the user's program result.
5. **Existing cycle refusal is left stale or weakened** — replace the old positive frontier with runnable witnesses and keep focused, source-attributed refusal tests for resource/loan carries and unsupported shapes.

## Implications for Roadmap

M005 is already confirmed as one milestone, but the research supports staged runnable phases. These are phase-shaping recommendations for the roadmapper; semantics, exact caps, and requirements remain subject to the requirements confirmation gate.

### Phase 26: Scalar arithmetic and `sum_to_n`
**Rationale:** This is the smallest complete end-to-end computation and creates an early runnable gain before FizzBuzz. The overflow rule and accepted `n` must be ratified before implementation plans rely on them.  
**Delivers:** Typed U64 add and comparisons, Bool/conditional form, a predicate loop, bounded input/work, decimal output, exact app result for `0`, representative input (`10 → 55`), accepted maximum and rejected neighbor, plus overflow behavior.  
**Addresses:** Table-stakes scalar arithmetic, loops, first ordinary runnable witness, independent answers.  
**Avoids:** Signed C overflow, build-dependent behavior, unbounded work, a loop checker that merely removes `cfg_back_edge`, and any resource/loan back-edge widening.  
**Evidence shape:** Independently derived expected values; interpreter and native compared separately; transfer/exit mutation and resource/loan refusal control. Preserve source attribution and stable diagnostics.

### Phase 27: Remainder, fixed text, and FizzBuzz
**Rationale:** Reuse the admitted loop/checker/native/output path and add only the missing composition primitives.  
**Delivers:** U64 `%`, constant-zero rejection and dynamic-zero defined failure, fixed byte literals, bounded decimal/fixed output, and exact FizzBuzz lines/terminator through the ordinary app route.  
**Uses:** Existing typed core, CFG and fixed-point analysis, interpreter, C17/Clang, and app stream boundary.  
**Avoids:** General string/format runtime, unbounded whole-output allocation, locale-sensitive formatting, and deriving expected bytes from either execution engine.  
**Evidence shape:** 0/1/15 and selected cap boundaries, exact stdout bytes including final newline, stdout/stderr/status separation, output-cap failure, and reached wrong-remainder/last-line controls.

### Phase Ordering Rationale

- Land a runnable `sum_to_n` slice before adding FizzBuzz-only remainder and multiple fixed labels; this meets the “runnable gain per feature phase” direction and localizes failures.
- Keep one canonical typed core and CFG so all independent peers and both engines share a representational contract while rederiving facts independently.
- Treat fixed-point analysis bounds as tooling safeguards, not language termination semantics; separately choose a small accepted input/work/output limit.
- Preserve resource/loan loop refusals as named negative controls. A future named consumer is required before expanding ownership semantics over cycles.
- Do not create a new dependency, backend, or general runtime for the examples. Promote those only with an evidenced consumer and lower total maintenance/security cost than copy-local code.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 26 arithmetic/loop contract:** resolve checked versus wrapping U64 overflow, whether `-`, `*`, or `/` are actually needed, dynamic error representation, scalar lattice and fixed-point invariants, event identity impact, and exact input/work limits. Source peers are numerous and loop evidence is not yet established.
- **Phase 27 bounded output/evidence:** settle the byte ceiling, output failure and partial-write behavior, decimal conversion bound, and whether looped event-producing operations require a schema identity change. Inspect existing app and evidence separation in code before committing design.

Phases with standard patterns (targeted research can be skipped if phase discussion resolves details):
- Lowering structured `while`/`if` into blocks and edges and using local decimal conversion are established implementation patterns, but Schway's independent peers and evidence protocol still need phase-specific design checks.
- The existing app-run process/stream route and Go/C U64 primitive behavior are already available; reuse them rather than researching a new integration stack.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Existing host/native/interpreter route was inspected and has M004 precedent; official specs document relevant C/Go behavior. The preferred overflow policy conflicts across reports, so semantic selection is not high-confidence. |
| Features | MEDIUM | Confirmed scope comes from project context; exact input/output caps, overflow policy, and precise syntax remain open. Recommendations were cross-checked against official language references. |
| Architecture | HIGH for current integration points; MEDIUM for loop design | Current source topology and six-consumer pattern were inspected. New fixed-point and dynamic event identity details are recommendations that need implementation design and evidence. |
| Pitfalls | HIGH for source boundaries; MEDIUM for external comparisons | Source inspection is concrete. External sources are official primary references, but they describe differing language policies, reinforcing the need to ratify Schway's own. |

**Overall confidence:** MEDIUM

### Gaps to Address

- **Overflow policy conflict:** Prefer checked arithmetic for this milestone's sum witness; requirements confirmation must make checked vs explicit wrap a deliberate Schway decision. Do not implement until settled.
- **Input/work/output limits:** No user-facing maximum or output ceiling is currently established. Choose reproducible limits and exact rejection/failure behavior in requirements; avoid promising that a U64 parse bound alone bounds runtime.
- **Operator scope:** The reports recommend add/sub/mul in some places but the witness minimally needs addition, comparisons/equality, and remainder. Admit only operations required by examples unless independent requirements add a consumer.
- **Looped evidence event identity:** Existing `/2` event identities are based on invocation and static IDs; repeated execution may collide. Determine whether new looped operations produce events and whether a versioned dynamic occurrence identity is needed. Do not widen loop support while evidence can silently conflate occurrences.
- **Dynamic arithmetic error plumbing:** Confirm how defined overflow and zero-divisor failures map to interpreter results, native exit/status, diagnostics, and evidence. Compiler rejection, runtime failure, and tool evidence exhaustion must remain distinguishable.
- **No newly executed validation:** Research did not run tests. Historical M003/M004/Phase 25 receipts establish existing infrastructure only; Phase 26/27 plans need named new expected-answer and negative-control evidence.

## Sources

### Primary (HIGH confidence)
- Go Authors, [The Go Programming Language Specification: integer operators and overflow](https://go.dev/ref/spec#Integer_operators) — unsigned operations, arithmetic, and division behavior.
- WG14, [N1570 C11 committee draft](https://oifans.cn/docs/c11/n1570.html), §§6.2.5 and 6.5.5 — unsigned arithmetic and division/remainder facts used for native lowering; cited explicitly as a C11 committee draft, not the published C17 standard.
- LLVM Project, [Undefined Behavior Manual](https://llvm.org/docs/UndefinedBehavior.html) and [Clang UBSan checks](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html#available-checks) — optimizer assumptions and limits of sanitizer coverage.
- Clang, [Cross-compilation](https://clang.llvm.org/docs/CrossCompilation.html) and [Toolchain](https://clang.llvm.org/docs/Toolchain.html) documentation — target/sysroot/toolchain boundary.
- Rust Project, [integer operator overflow](https://doc.rust-lang.org/reference/expressions/operator-expr.html), [loop expressions](https://doc.rust-lang.org/reference/expressions/loop-expr.html), and [`u64` arithmetic APIs](https://doc.rust-lang.org/std/primitive.u64.html); Swift, [integer bounds and overflow](https://docs.swift.org/swift-book/LanguageGuide/TheBasics.html) — comparative evidence that languages expose different explicit policies, not direct Schway requirements.
- Detailed reports: [STACK.md](STACK.md), [FEATURES.md](FEATURES.md), [ARCHITECTURE.md](ARCHITECTURE.md), [PITFALLS.md](PITFALLS.md).

### Project source and historical evidence
- Current source inspected: `internal/compiler/check/check.go` (`check.cfg_back_edge`, `loanLivenessFixpoint`); `internal/compiler/pathoracle/pathoracle.go`; `internal/compiler/core/core.go`; validators, interpreter, C emitter, execution peers, session, and native app files named in the detailed reports.
- Current tests/source anchors inspected but not run: `internal/compiler/session/session_phase20_test.go:TestPhase20ChecksumFrontier`; Phase 19 U64 fixtures; `internal/compiler/native/native_app_test.go` Phase 22 stream/capacity cases.
- Historical receipts, not rerun in this research: M003 evidence audit, Phase 14 evidence fixes, Phase 15 native capacity resolution, M004 EVD-11 and Phase 25 wrong-result controls, and prior six-consumer/interpreter/native work. These do not validate M005 arithmetic or loops.
- Planning facts inspected: `.planning/PROJECT.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, and `.planning/REQUIREMENTS.md`.

---
*Research completed: 2026-10-02*  
*Ready for roadmap: yes, after requirements-stage ratification of the open semantic decisions above.*
