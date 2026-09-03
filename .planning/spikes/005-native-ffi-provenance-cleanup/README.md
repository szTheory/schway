---
spike: 005
idea: ownership-kernel
name: native-ffi-provenance-cleanup
type: standard
validates: "Given the candidate ownership and resource invariants expressed through a minimal native C path, when compiled at O0/O3 and crossed through independently declared C boundaries, then semantic events agree and hostile ABI, alias, provenance, allocator, cleanup, sanitizer, and nonlocal-exit defects are observable"
verdict: PARTIAL
related: [001, 002, 003, 004]
tags: [native, ffi, abi, provenance, cleanup, optimization, sanitizers]
---

# Spike 005: Native FFI, provenance, and cleanup

## What this validates

This spike moves the ownership work out of an abstract Go state machine and
through a real native toolchain. A Go evidence runner independently predicts a
small semantic event stream, compiles separate C translation units with Apple
Clang at `-O0` and `-O3`, executes them, and compares their machine-readable
events.

The native cases cover:

- a record returned by value across independently declared translation units;
- reverse-order cleanup after first-stage failure, second-stage failure, and
  success;
- allocator identity carried with an allocation and checked at release;
- a retained borrowed callback view rejected after its owner generation ends;
- alias-safe code versus an unsound `restrict`/no-alias promise under
  optimization;
- a concrete heap use-after-free under AddressSanitizer and UBSan; and
- a foreign `longjmp` that bypasses ordinary cleanup.

The result is partial because these are hand-authored C lowering analogues. No
Lang typed core or backend exists yet, so the experiment establishes native
contract obligations and a reusable differential harness—not the correctness
of a Lang-to-native lowering.

## Research

LLVM treats `noalias`, pointer capture, provenance capture, `nofree`, lifetime,
and unwind behavior as semantically meaningful optimizer promises. A frontend
that emits one of these attributes without proof can therefore create wrong
code rather than a conservative performance loss. WG14's continuing work on
`restrict` also shows how subtle the source rule is; it is not a casual hint.

Clang documents AddressSanitizer and UndefinedBehaviorSanitizer as bug
detectors, not production runtimes or formal proofs. They provide a distinct
dynamic lens but cannot replace static ownership or differential optimization
checks. Apple's arm64 guidance makes the target ABI part of the contract rather
than an implementation detail.

- [LLVM language reference: noalias, capture, provenance, and unwind](https://llvm.org/docs/LangRef.html)
- [WG14 restrict-semantics proposal](https://www.open-std.org/jtc1/sc22/wg14/www/docs/n3234.htm)
- [Clang AddressSanitizer](https://clang.llvm.org/docs/AddressSanitizer.html)
- [Clang UndefinedBehaviorSanitizer](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html)
- [Apple arm64 ABI guidance](https://developer.apple.com/documentation/xcode/writing-arm64-code-for-apple-platforms)
- [Alive2 bounded translation validation](https://web.ist.utl.pt/nuno.lopes/pubs.php?id=alive2-pldi21)

| Mechanism | What it catches | What it misses | Disposition |
|---|---|---|---|
| interpreter/native event comparison | lowering changes to values, cleanup, and outcomes | shared wrong expectations; unobserved UB | mandatory from first backend |
| `-O0`/`-O3` differential run | optimizer-sensitive false contracts | defects stable at both levels | mandatory hostile lane |
| ASan/UBSan | concrete memory and UB executions | unexecuted paths and semantic API lies | routine verify/release evidence, not core semantics |
| runtime generation token at dynamic callback boundary | stale retained foreign view in a debuggable way | statically known ordinary borrows should never pay this cost | sanitizer/foreign escape hatch only |
| content-bound evidence manifest | exact source/toolchain/flag provenance | compiler truth, signer authority, target equivalence | retain from Spike 004 |
| formally verified or translation-validated backend | stronger lowering relation | high early implementation cost and specification trust | later assurance tier after IR stabilizes |

## Experiment shape

The producer translation unit defines and returns `LangPacket`; the consumer
translation unit independently declares its layout. The hostile layout build
reorders fields in the consumer only. Linkage succeeds, but the semantic event
shows corrupted fields.

Resources use explicit initialization flags and one cleanup block. Three paths
must acquire and release the same number of resources in reverse order. The
hostile build removes the second release.

Allocator identity lives in a private header. A release with the wrong
allocator is rejected while a matching release succeeds. The hostile build
removes the identity check.

The callback view carries an owner generation. The normal boundary rejects use
after the generation changes. This dynamic token is a model for unsafe or
foreign callbacks only; ordinary Lang borrows should remain statically checked.

The optimizer probe compares an ordinary aliased pointer function with a
version whose parameters are incorrectly marked `restrict`. On the current
Clang/arm64 host, the false contract is latent at `-O0` and changes the result at
`-O3`.

## Gates

1. Native events match the independent host oracle at both `-O0` and `-O3`.
2. Separate C translation units agree on by-value layout and calling
   convention; a field-order mutation must be observable.
3. Every partial-initialization and success path releases exactly what it
   acquired in reverse order; an omitted release must fail.
4. Allocator mismatch and stale callback retention must be rejected.
5. The deliberately false no-alias contract must be exposed by an optimized
   differential run.
6. A concrete retained raw pointer must be rejected by a sanitizer build.
7. A foreign nonlocal exit must demonstrate that lexical cleanup cannot be
   promised across arbitrary C control transfer.
8. The report records compiler version, target, source digest, flags, compile
   time, binary size, and stable semantic events.
9. Do not graduate the lowering claim until the same harness consumes real Lang
   typed core and backend output.

## How to run

From this directory on a host with Clang:

```sh
env GOCACHE=/private/tmp/ai-lang-spike005-go-cache go test -count=1 ./...
env GOCACHE=/private/tmp/ai-lang-spike005-go-cache go test -race -count=1 ./...
env GOCACHE=/private/tmp/ai-lang-spike005-go-cache go test -coverpkg=./... ./...
env GOCACHE=/private/tmp/ai-lang-spike005-go-cache go vet ./...
env GOCACHE=/private/tmp/ai-lang-spike005-go-cache go run ./cmd/native-lab -pretty=false
```

The runner uses a private temporary build directory and removes it afterward.
The AddressSanitizer program is expected to terminate unsuccessfully; detection
of its `heap-use-after-free` report is a passing experiment result.

## What to expect

On the recorded Apple Clang 21.0.0 arm64 host:

- six baseline semantic events agree at both optimization levels;
- seven of seven hostile cases are observed;
- normal aliased code returns `2` at `-O0` and `-O3`;
- the falsely restricted version returns `2` at `-O0` but `1` at `-O3`;
- AddressSanitizer terminates the retained raw-pointer case with a heap-use-
  after-free report;
- `longjmp` leaves one acquired resource and zero releases; and
- the two baseline binaries are roughly 34 KB and compile in well under 100 ms
  in observed warm host runs.

Those timings and sizes are observations, not product budgets or cross-platform
claims.

## Investigation trail

### Iteration 1 — establish an optimization-independent baseline

The initial native program emits values and resource events as JSON lines. The
Go runner contains the expected events rather than deriving them from C output.
Both `-O0` and `-O3` produced the same six-event result.

### Iteration 2 — break the ABI without breaking the linker

The producer and consumer declare the record in separate translation units.
Changing only the consumer field order still links, but corrupts the returned
field interpretation. This is exactly why generated C headers and layout hashes
need conformance tests; linker success is not ABI evidence.

### Iteration 3 — make resource ownership visible

Partial initialization records which resources are live, then unwinds one
explicit cleanup path in reverse order. Removing release of the second resource
fails both the second-stage error and success cases. Allocator pairing and
callback owner generations expose two more obligations that a raw pointer type
does not carry.

### Iteration 4 — prove optimizer attributes can change meaning

The no-alias mutation is the highest-value observation. Passing the same pointer
for both `restrict` parameters violates the contract. The current compiler
returns the intuitive value at `-O0` and a different value at `-O3`. Lang must
derive alias/capture attributes from checked IR facts; unsafe authors and FFI
metadata must not assert them as unchecked performance decorations.

### Iteration 5 — add orthogonal dynamic evidence

ASan catches an actual heap use after free. A valid `longjmp` demonstrates a
different problem: nonlocal foreign control transfer can skip cleanup without a
memory access violation. Sanitizers, cleanup-event comparison, and ABI/optimizer
differential tests therefore cover different defect classes.

## Results

**Verdict: PARTIAL. The native contract and harness are validated; Lang native
lowering is not.**

The first real backend must preserve these facts explicitly:

- target-specific layout and calling convention are versioned interface data;
- initialized fields/resources are tracked so every exit releases only live
  values in specified reverse order;
- allocator identity travels with owning storage unless a boundary proves a
  universal pairing;
- callback retention, address capture, provenance capture, and read/write
  authority are separate FFI facts;
- no-alias, capture, `nofree`, lifetime, and unwind attributes are generated
  only from proven typed-core obligations;
- panic/unwind/nonlocal exit may not cross an unaudited C ABI boundary;
- sanitizer, unoptimized, optimized, and semantic-oracle lanes remain distinct;
  and
- every artifact binds source digest, compiler/toolchain identity, target,
  flags, and semantic evidence.

This is the stopping point for disposable Go/C ownership workbenches. Spikes
001–005 now define enough versioned fixtures and failure modes to begin the real
lossless frontend and portable typed core. The native harness should become a
consumer of that core; it should not grow a parallel pseudo-language.

## Known limitations

- Only Apple arm64 with Apple Clang 21 was executed. Windows, ELF platforms,
  other architectures, other Clang/LLVM releases, GCC, and cross-compilation
  remain open.
- The layout fixture is one small returned record. Unions, bit fields, vectors,
  variadics, callbacks, atomics, packed/aligned records, and aggregate passing
  thresholds remain for the real ABI suite.
- Generation-token checking models a dynamic foreign boundary. It is not a
  proposal to add runtime checks to every borrow.
- ASan/UBSan coverage is path-dependent and does not prove absence of UB.
- `longjmp` is a representative nonlocal exit, not the complete C++ exception,
  panic, cancellation, signal, or process-termination matrix.
- The no-alias divergence is compiler/target-sensitive. The semantic lesson is
  stable; the exact `2/1` observation is not guaranteed on every toolchain.
- The host oracle and C program are manually kept aligned. The real frontend
  must generate both lowering and manifest from versioned typed core while an
  independent validator checks the relation.
