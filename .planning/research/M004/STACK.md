# M004 Stack Research

**Date:** 2026-09-27. **Repository baseline:** `d9bde05`.
**Scope:** Native application execution and bounded resource discharge.
Current repository observations are source inspection; proposed gates are not
claims that the functionality has shipped.

## Recommendation

Retain Go 1.24 and its standard library for Stage 0, readable C17 through the
installed Clang for native execution, and the independent interpreter for
bounded semantic comparison. No production dependency, VM, LLVM binding,
global tracing heap, SMT solver, package manager, or service runtime is needed
for this milestone. Installed tool versions belong in actual evidence receipts;
this document does not assert that a particular tool is the latest release.

| Component | Keep / change | Reason and integration |
|---|---|---|
| `go.mod` / Stage 0 | Keep Go 1.24, zero third-party dependencies | Existing parser, checker, IR, process orchestration and evidence tooling are sufficient |
| `cgen.emitProgram` | Extend the sole production emitter | Preserve the one-emitter law; do not restore retired bodies to bypass refusal |
| Clang / C17 | Keep installed compiler; record target/compiler/flags | C ABI must be exercised per family; safe Lang operations must not lower to C undefined behavior |
| `cmd/lang`, `session`, `native` | Add a real build/application route | Current temp executables and strict execution-JSON output are conformance infrastructure; application IO must execute once |
| Build description | Explicit local source/symbol/ABI inputs | Source-relative paths survive checkout relocation; no magic fixture-symbol lookup, ambient search or implicit downloads |
| Resource adapter | Small audited C source supplied explicitly | Return a malloc-backed bounded buffer live; adapter owns internal file operations, Lang owns/free-dispatches the allocation |
| Interpreter/evidence | Explicit isolated/replayable conformance lane | Do not transparently interpret and repeat actual user side effects at O0/O3 |
| ASan/UBSan and host lanes | Use existing infrastructure where each claim needs it | Physical release witness and host/ABI evidence are distinct from serialized event agreement |
| Planning automation | AGENTS + PRODUCT-ROADMAP + LANGUAGE-MATURITY | Existing GSD transition work is enough; no new registry, executable or global skill |

## Application boundary

Build a retained executable with explicit inputs and artifact paths; application
run executes that artifact once. Define bounded argv/input, stdout/stderr, exit
status and compiler evidence channels before enabling side-effectful foreign
code. Preserve existing conformance commands/contracts where needed; their
differential replay must use disposable or reproducible inputs.

The chosen resource program reads a caller-selected file via an audited adapter,
returns the allocation live, accesses it through Lang, and invokes generated
cleanup. General byte-array source syntax, dynamic strings, loops and fallible
implicit close are not prerequisites. The exact entry/input representation is
a Phase 22 design decision, constrained by this visible behavior.

## ABI and optimizer policy

Ordinary C pointers can represent the bounded shared/exclusive read-copy shapes.
Emit no extra `restrict`, `noalias`, capture/alignment or similar promise without
a separately checked fact and applicable evidence. Update manifests from what
is actually emitted. Keep the source borrow rules independent of optional
optimization attributes.

The prospective D-16-07 amendment removes an unnecessary optimization
prerequisite; it does not waive per-family pointer ABI witnesses, macOS/Linux
evidence, or structural refusals of escape, mutation, forwarding and callbacks.
See [LLVM's attribute contract](https://llvm.org/docs/LangRef.html#parameter-attributes)
and [WG14 N1570 §6.7.3.1](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n1570.pdf).
N1570 is a C11 committee draft, not a fetched C17 standard.

## Evidence tools and limits

[Clang AddressSanitizer](https://clang.llvm.org/docs/AddressSanitizer.html)
provides memory-error detection, and
[UndefinedBehaviorSanitizer](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html)
checks selected undefined operations. Neither proves resource protocol
correctness. Leak-detection availability is host/toolchain scoped; independently
observed outstanding allocations and destructor calls remain necessary.

Use expected answers in addition to interpreter/native agreement. A mutation
that omits the real destructor while leaving events unchanged must fail an
independent observer. Record skipped/unavailable lanes without closing their
claims. The [LLVM testing guide](https://www.llvm.org/docs/TestingGuide.html)
is a useful precedent for separating regression tests from whole-program
execution evidence, not a replacement for Lang's acceptance criteria.

## Following milestone implications

FizzBuzz requires arithmetic, comparison, scalar iteration and minimal output;
none requires a new backend. Choose arithmetic semantics explicitly, then lower
them without accidental C undefined behavior. For loops, study CFG fixed-point
analysis and dynamic identities before replacing the existing cycle refusal.
[Rust NLL RFC 2094](https://rust-lang.github.io/rfcs/2094-nll.html) and
[MIR RFC 1211](https://rust-lang.github.io/rfcs/1211-mir.html) provide design
precedents, not semantics to import wholesale. Initial scalar loops may retain
refusal of resources/loans across back edges.

## Reopen conditions

Reconsider the backend or dependencies only when a runnable consumer exposes a
measured target, latency, ABI, optimization, security or maintenance limitation
that the present shallow boundaries cannot reasonably address. Do not turn a
desire for usable programs into an unrelated compiler/runtime migration.
