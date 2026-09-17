# Native Lowering and the FFI Contract

What the first real backend must preserve explicitly, established by running
candidate invariants through Apple Clang at `-O0` and `-O3` with hostile
mutations.

## Requirements

From the `ownership-kernel` idea (MANIFEST.md):

- Treat target layout, initialization, allocator pairing, callback retention,
  address/provenance capture, cleanup, and unwind as explicit typed-core/FFI
  obligations. Derive optimizer attributes only from proven facts.
- Require unoptimized, optimized, semantic-oracle, and sanitizer evidence as
  distinct native lanes. Do not allow panic, unwind, or foreign nonlocal exit
  across an unaudited C ABI boundary.

## How to Build It

The first real backend must preserve these facts explicitly:

1. **Target-specific layout and calling convention are versioned interface
   data.** The producer translation unit defines and returns `LangPacket`; the
   consumer independently declares its layout. Reordering fields in the consumer
   only **still links** and corrupts the returned field interpretation. Generated
   C headers need layout hashes and conformance tests — **linker success is not
   ABI evidence.**
2. **Track initialized fields and resources so every exit releases only live
   values, in specified reverse order.** Use explicit initialization flags and one
   cleanup block; all three paths (first-stage failure, second-stage failure,
   success) must acquire and release the same count in reverse. Removing the
   second release fails both the second-stage-error and success cases.
3. **Allocator identity travels with owning storage** unless a boundary proves a
   universal pairing. Keep the identity in a private header; check it at release.
4. **Callback retention, address capture, provenance capture, and read/write
   authority are separate FFI facts.** Not one pointer bit.
5. **Emit `noalias`, capture, `nofree`, lifetime, and unwind attributes only from
   proven typed-core obligations.** See What to Avoid — this is the highest-value
   observation in the spike.
6. **Panic, unwind, and nonlocal exit may not cross an unaudited C ABI
   boundary.**
7. **Keep sanitizer, unoptimized, optimized, and semantic-oracle lanes
   distinct.** They catch different defect classes and none subsumes another.
8. **Bind every artifact to source digest, compiler/toolchain identity, target,
   flags, and semantic evidence.**

**Harness shape:** a Go evidence runner independently *predicts* the semantic
event stream (it contains the expected events rather than deriving them from C
output), compiles separate C translation units at `-O0` and `-O3`, executes
them, and compares machine-readable events. Builds go in a private temporary
directory outside the source tree and are removed afterward.

**A dynamic generation token is a foreign-boundary escape hatch, not a borrow
mechanism.** The callback view carries an owner generation and is rejected after
that generation ends. Model it for unsafe/foreign callbacks only — ordinary Lang
borrows are statically checked and must not pay this cost.

## What to Avoid

- **Asserting optimizer attributes as unchecked performance decorations.** The
  no-alias probe compares an ordinary aliased function against one whose
  parameters are incorrectly marked `restrict`. Passing the same pointer for both
  violates the contract, and on the recorded host the function returns **`2` at
  `-O0` and `1` at `-O3`**. A frontend that emits such an attribute without proof
  creates **wrong code**, not a conservative performance loss. WG14's continuing
  work on `restrict` shows the source rule is subtle — it is not a casual hint.
- **Treating sanitizers as sufficient.** Clang documents ASan/UBSan as bug
  detectors, not production runtimes or formal proofs. They are a distinct
  dynamic lens; they cannot replace static ownership checks or differential
  optimization runs, and their coverage is path-dependent.
- **Assuming lexical cleanup survives foreign control transfer.** A valid
  `longjmp` leaves **one acquired resource and zero releases** — no memory access
  violation, no sanitizer report, cleanup simply skipped.
- **Deriving the oracle's expectations from the C program's output.** Then the
  comparison proves nothing.
- **Graduating the lowering claim from this spike.** Gate 9 is explicit: do not,
  until the same harness consumes real Lang typed core and backend output.
- **Growing the native harness into a parallel pseudo-language.** It should become
  a *consumer* of the real typed core.

## Constraints

- **Only Apple arm64 with Apple Clang 21.0.0 was executed.** Windows, ELF
  platforms, other architectures, other Clang/LLVM releases, GCC, and
  cross-compilation are all open.
- The layout fixture is **one small returned record**. Unions, bit fields,
  vectors, variadics, callbacks, atomics, packed/aligned records, and aggregate
  passing thresholds remain for the real ABI suite.
- ASan/UBSan **do not prove absence of UB**.
- `longjmp` is one representative nonlocal exit — not the complete C++ exception,
  panic, cancellation, signal, or process-termination matrix.
- **The `2`/`1` no-alias divergence is compiler- and target-sensitive.** The
  semantic lesson is stable; the exact observation is not guaranteed elsewhere.
- The host oracle and the C program are **manually kept aligned**. The real
  frontend must generate both lowering and manifest from versioned typed core,
  with an independent validator checking the relation.
- Baseline binaries were ~34 KB and compiled in well under 100 ms on warm runs.
  Observations, not budgets, and not cross-platform claims.

## Measured result (spike 005, PARTIAL)

Six baseline semantic events agree at both optimization levels · 7/7 hostile
cases observed · AddressSanitizer terminates the retained raw-pointer case with
a `heap-use-after-free` report (its unsuccessful exit **is** the passing result).

**Verdict: PARTIAL — the native contract and harness are validated; Lang native
lowering is not.** This is the stopping point for disposable Go/C ownership
workbenches; spikes 001–005 define enough versioned fixtures and failure modes to
begin the real lossless frontend and portable typed core.

## Origin

Synthesized from spikes: 005
Source files available in: `sources/005-native-ffi-provenance-cleanup/`
