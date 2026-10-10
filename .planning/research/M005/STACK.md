# Technology Stack

**Project:** Schway M005 Practical Computation  
**Researched:** 2026-10-02

## Recommended Stack

| Component | Recommendation | Purpose | Why |
|---|---|---|---|
| Stage 0 compiler | Keep Go 1.24 and the standard library | Parsing, checking, core IR, CLI, evidence, process orchestration | Existing toolchain already hosts all these responsibilities; the active dependency preference favors copy-local code when maintainable. |
| Native path | Keep readable C17 emitted by the sole production emitter and compiled by installed Clang | Production native artifacts | M004 and the project roadmap establish this path; no M005 requirement demands a new backend or runtime. |
| Semantic oracle | Extend the existing interpreter independently | Expected/differential outcomes for admitted scalar operations | It supplies an independent implementation path; interpreter/native agreement must be supplemented with pinned expected answers and negative controls. |
| Numeric model | Define Schway U64 as unsigned 64-bit arithmetic with wraparound for `+`, `-`, `*`; define comparisons as total over `0..2^64-1` | Predictable scalar computation | Go unsigned operators and C `uint64_t` operations provide an implementation match for modular arithmetic. Keep these semantics explicit in Schway rather than treating host-language behavior as the specification. |
| Division/remainder | Define unsigned quotient/remainder for nonzero divisor; make zero divisor a checked/runtime-defined Schway failure before native division | FizzBuzz remainder and future arithmetic | C division and remainder by zero are undefined behavior; an explicit guard is required if Schway promises a defined result. Go's runtime panic is an available precedent, not a requirement to expose Go panic behavior. |
| Output | Reuse existing bounded application stdout route; implement decimal U64 formatting and fixed text/byte output locally | `sum_to_n` and FizzBuzz output | These programs need only bounded scalar-to-decimal and fixed-message formatting. A dependency or general string runtime would exceed the demonstrated consumer. |
| Validation hosts | Native execution on macOS and Linux, with compiler/target/flags recorded | Portability evidence | Language semantics should not depend on current Apple arm64. Clang target selection alone does not supply target headers, libraries, sysroot, or linker; native host lanes provide honest initial portability evidence. |

**Concrete path:** retain Go stdlib → checked Schway core → interpreter and readable C17 → installed Clang. Admit only the scalar loop subset needed by `sum_to_n` and FizzBuzz, leaving resource/loan carries across loop back-edges refused until separately justified and proved. Add no package, runtime, backend, or framework for this slice.

## Arithmetic and C Definedness

- **Overflow is a language contract.** Recommend modulo-`2^64` for U64 add, subtract, and multiply. This gives deterministic behavior across interpreter and native execution and matches the unsigned arithmetic model of both Go and C. If the product instead chooses checked overflow, it must represent and report that outcome in both engines; do not rely on signed overflow, sanitizer behavior, or optimization flags to define it.
- **Emit unsigned C.** Lower Schway U64 to `uint64_t` from `<stdint.h>` and keep operands/results in that type through arithmetic and comparison. Avoid accidentally converting to signed C types, using unsuffixed constants that promote unexpectedly, or performing arithmetic in a narrower intermediate. Check representability of source constants before serialization.
- **Guard division and remainder.** In C17, `/` and `%` with a zero divisor have undefined behavior; the compiler must emit a Schway-defined failure/check edge before either operation. The checker may prove a divisor nonzero when it has a sound fact, but the native lowering and runtime contract must cover unproven or input-derived values. For unsigned operands, C quotient/remainder agree with the expected nonnegative arithmetic when the divisor is nonzero.
- **No host panic leakage.** The Go interpreter must implement Schway's explicit outcome, not expose Go panic as accidental language semantics. Keep checked errors/outcomes comparable with native execution.
- **Bounded formatting.** Format U64 using a small copy-local decimal conversion with explicit output-capacity handling; use bounded writes and fixed literals for FizzBuzz. Avoid `fmt` in generated C or an allocation-heavy dynamic string runtime. Go's standard library remains appropriate for compiler-side test/evidence orchestration, but it should not define target application behavior.

The Go specification defines unsigned arithmetic modulo `2^n` and runtime integer division by zero as a panic; its quotient/remainder relation is useful as a reference. WG14 N1570, a C11 committee draft commonly used for clause text (not itself the published C17 standard), defines unsigned arithmetic modulo the representable range and states division/remainder by zero is undefined. These external facts support implementation choices; the wrapping/failure policy above is a Schway semantic recommendation.

## Alternatives Considered

| Category | Recommended | Alternative | Why not (yet) |
|---|---|---|---|
| Backend | Existing C17 emitter + Clang | LLVM API/direct IR backend | New backend adds a second semantic/lowering path and maintenance surface without a M005 consumer or measured limitation. Reopen only for a demonstrated target, latency, diagnostics, or codegen constraint. |
| Interpreter arithmetic | Copy-local exact U64 operations | Add a big-integer or numeric dependency | U64 is fixed width and Go already provides `uint64`; a dependency does not solve semantic agreement or evidence obligations. |
| Overflow policy | Explicit modular U64 arithmetic | Trap/checked arithmetic | Checked arithmetic is viable if product semantics demand it, but it requires a first-class failure contract, checker/lowering parity, and defined overflow result evidence. Do not adopt by accident through signed C. |
| Text output | Bounded local decimal routine + existing output route | General-purpose string/formatting library in target | M005 needs only bounded decimal digits and literals; a dynamic string system adds representation, allocation, encoding and cleanup questions prematurely. |
| Test/runtime helper | Go stdlib and existing scripts | Third-party property, SMT, or cross toolchain dependency | No concrete need has been identified. Use deterministic boundary tables and existing mutation/differential infrastructure first; revisit if a named evidence gap persists. |
| Cross-platform strategy | Run native evidence on macOS and Linux | Assume one local Clang can target every OS | Cross-target Clang still needs target ABI headers, libraries, sysroot and linker. Cross-compilation is useful supplementary coverage but is not a substitute for executing on both supported hosts. |

## Standard Library and Dependency Decision

Use only the Go standard library for host-side implementation. Add no third-party dependency for scalar arithmetic, CFG iteration, or bounded output. The maintainability threshold is concrete: adopt a dependency only if a consumer-specific need cannot be implemented as a small, reviewable local component and the added abstraction reduces total maintenance rather than merely shortening code.

No installation changes are recommended. Keep `go.mod` at its current Go 1.24 baseline; keep Clang as the already-required external native compiler. Record the actual Go and Clang versions in phase evidence, not as a claim that a particular release is current.

## M005 Evidence Constraints

1. Pin U64 edge cases: zero, one, max, wrap at `2^64`, comparison boundaries, division/remainder by one, and zero-divisor behavior.
2. Exercise those cases independently through interpreter and native paths; compare each to explicit expected outcomes. Include a negative control that would catch a signed intermediate or omitted zero guard.
3. Prove scalar loop fixed points and termination/bounds for the admitted source forms. Keep resource/loan loop carries at refusal boundaries; scalar CFG admission does not imply cleanup/borrow correctness on cycles.
4. Validate exact bounded output including zero, max decimal formatting, capacity exhaustion, and FizzBuzz boundary values; preserve application stdout/stderr/exit behavior.
5. Run native receipts on macOS and Linux. Record source/build identity, host and target, compiler version, flags, expected result, and skipped lanes. Do not describe a cross-compile as runtime validation.

## Sources

- [Go specification: integer operators and overflow](https://go.dev/ref/spec#Integer_operators) — primary language semantics; queried 2026-10-02.
- [WG14 N1570 C committee draft, §6.2.5 and §6.5.5](https://oifans.cn/docs/c11/n1570.html) — unsigned arithmetic and integer division/remainder clauses. N1570 is a C11 committee draft, used here for stable clause text rather than labeled as the published C17 standard.
- [Clang cross-compilation documentation](https://clang.llvm.org/docs/CrossCompilation.html) — target selection and target-specific toolchain/sysroot dependencies; queried 2026-10-02.
- [Clang complete toolchain documentation](https://clang.llvm.org/docs/Toolchain.html) — compiler, runtime, assembler, and linker boundary; queried 2026-10-02.
- Repository inspection: `go.mod`, `.planning/research/M004/STACK.md`, `.planning/PRODUCT-ROADMAP.md`, `.planning/LANGUAGE-MATURITY.md`, and `.planning/REQUIREMENTS.md`. These are project facts, distinct from external language-standard claims.
