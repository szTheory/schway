---
id: debug-evidence-symbolication-proof
title: Debug evidence, symbolication, crash artifacts, and proof posture
summary: Preserve semantic identity from source through optimized native code so humans and agents can diagnose failures from bounded, privacy-aware evidence.
type: design
status: candidate
confidence: medium
created: 2026-09-03
updated: 2026-09-03
tags: [debugging, symbols, crash, provenance, proof]
related: [semantic-kernel-contract, semantic-kernel-probes, ai-native-runtime-evals, context-telemetry-security, native-low-level-profile, runtime-profiles-dogfooding, research-ledger]
---

# Debug evidence, symbolication, crash artifacts, and proof posture

## Conclusion

Debuggability should be an end-to-end compiler contract, not a release-build
afterthought. Every source construct, typed-core fact, generated operation,
semantic event, machine-code range, diagnostic, profile sample, and crash frame
should be joinable through stable semantic identities.

The leading design is two complementary evidence planes:

1. **Semantic evidence**: versioned source/core IDs, ownership/effect/resource
   transitions, causal diagnostics, and deterministic interpreter/native traces.
2. **Physical evidence**: target-native DWARF or CodeView, build IDs, symbol
   bundles, machine addresses, inline stacks, registers, and bounded crash dumps.

Neither plane should pretend to reconstruct facts that optimization removed.
Agent-facing output must report `available`, `optimized_out`, `redacted`,
`not_captured`, or `unsupported` instead of inventing a value.

## What is already being dogfooded

The project has reached D0 and part of D1 of the dogfood ladder:

- Lang source fixtures drive the real formatter, checker, interpreter, C17
  emitter, Clang O0/O3 executions, evidence manifests, and verification gates.
- Compiler work, output bytes, stable diagnostic/control IDs, semantic events,
  and expected escapes are already machine-readable.
- Independent reviews used those facts to find a non-causal backend result,
  ownership identity errors, boundedness gaps, name collisions, and a golden
  compatibility regression.

This is already an AI-development dividend: failures are narrower and
reproducible than prose-only logs. It is not yet full AI-native dogfooding.
The compiler and harness are still written in Go, queries are mostly batch
commands, and the D2 Lang-written corpus/evidence runner does not exist.

## Required identity chain

```text
source span + syntax ID
  -> typed-core declaration/place/operation ID
  -> lowering/optimization lineage
  -> native address range + build ID
  -> semantic event / diagnostic / profile sample / crash frame
```

Identity is semantic and versioned; byte offsets are projections. A formatter,
macro expansion, inlining, monomorphization, or code motion may map one source
identity to many physical ranges or many source identities to one range. The
mapping therefore needs explicit one-to-many lineage rather than a single
`file:line` field.

LLVM's debug model maps source AST objects to IR metadata and lowers to DWARF
or CodeView; it also warns that optimized variable recovery is inherently more
complex than line mapping. Assignment tracking retains source-assignment facts
longer through optimization. ([LLVM source-level debugging](https://llvm.org/docs/SourceLevelDebugging.html),
[LLVM assignment tracking](https://llvm.org/docs/AssignmentTracking.html))

## Artifact profiles

| Profile | Required artifacts | Default policy |
|---|---|---|
| local debug | embedded line/type/variable info, semantic maps, assertions | rich and immediate |
| optimized debug | line/inline/call-site maps, build ID, semantic IDs, honest unavailable values | optimized code with symbols retained locally |
| release | stripped binary plus separately stored symbol bundle and manifest | reproducible, content-addressed, access-controlled |
| crash capture | bounded minidump/core subset plus classified context and build ID | local persistence first; upload requires policy/consent |
| replay/evidence | semantic trace, nondeterministic inputs, toolchain and artifact digests | separate from physical dump; explicit completeness |

DWARF 5 supports call-site information and split/supplementary debug objects,
which fits small release binaries plus separately retained symbols.
([DWARF 5](https://dwarfstd.org/dwarf5std.html)) Minidumps can retain threads,
modules, register contexts, and selected stack memory for later symbolication;
they still require matching debug information and are sensitive data.
([Crashpad dump tool](https://chromium.googlesource.com/crashpad/crashpad/%2Bshow/HEAD/tools/generate_dump.md),
[Chromium minidump guide](https://chromium.googlesource.com/chromium/src/%2B/HEAD/docs/linux/debugging_minidump.md))

## Shift-left requirements

- Emit location/lineage facts from typed core before the first lowering; do not
  reverse-engineer them from generated C later.
- Make every optimization declare how it preserves, merges, duplicates, or
  intentionally drops semantic/debug identities.
- Differentially test O0/O3 stack traces, inline frames, source locations, and
  ownership/resource event attribution on a small golden corpus.
- Validate symbol bundles against build IDs and complete compiler/target flags;
  never symbolize with a merely similar binary.
- Treat symbol generation and symbolication as bounded compiler lanes with
  stable schemas, work counts, timeouts, and output caps.
- Minimize crash cases while preserving the same semantic failure signature.
- Redact secrets, PII, raw prompts, tenant data, and arbitrary stack memory by
  classification policy before upload; keep local crash persistence recoverable.
- Report capture loss explicitly. A truncated dump or missing symbol server is
  evidence about observability failure, not evidence that the program was safe.

## Agent-facing crash envelope

```text
record CrashEvidence {
  schema: CrashEvidenceSchema
  crash_id: CrashId
  build: BuildIdentity
  target: TargetIdentity
  signal: CrashSignal
  semantic_frames: List<SemanticFrame>
  native_frames: List<NativeFrame>
  last_events: BoundedList<SemanticEvent>
  unavailable: List<UnavailableFact>
  redactions: List<RedactionFact>
  artifacts: List<ContentDigest>
}
```

Normal output should be a compact causal summary. Exact frames, source slices,
IR lineage, and dump references expand by stable ID on request. This is the
same summary/detail pattern already proven useful by the compiler evidence
workbenches.

## Proof posture

Do not turn the general-purpose language into a proof assistant in v1. The
high-leverage lesson is architectural: a rich, convenient elaborator may
produce a small explicit artifact checked by a smaller trusted kernel. Lean
uses this separation and also records interactive metadata in side tables;
its documentation is explicit that a checked proof does not establish that the
formal statement matches human intent. ([Lean elaboration and kernel](https://lean-lang.org/doc/reference/latest/Elaboration-and-Compilation/),
[Lean proof validation](https://lean-lang.org/doc/reference/latest/ValidatingProofs/))

Apply that lesson selectively:

- keep the typed core and independent validators small and versioned;
- permit optional proof/certificate producers for ownership, effects, bounds,
  protocols, or selected refinements;
- publish the theorem/claim, assumptions, axioms, tool identity, and residual
  trust boundary alongside the certificate;
- use translation validation and differential/native evidence where proving the
  entire evolving compiler would be too expensive;
- introduce SMT/refinement/proof tiers only after concrete workloads show that
  ordinary types, effects, contracts, properties, and validators are inadequate.

The anti-pattern is a powerful proof surface whose elaboration latency,
diagnostics, or hidden axioms degrade the default edit loop. Proof strength is
an assurance tier; fast ordinary checking remains the default.

## Next bounded experiment

During Phase 3 planning, add a debug-lineage micro-slice rather than building a
crash platform:

1. Attach stable source/core IDs to one match arm, move, borrow, and return.
2. Emit a sidecar semantic map plus native debug info for O0 and O3.
3. Capture one deliberate panic/segfault and one ownership diagnostic.
4. Symbolize both to the same semantic IDs, with honest optimized-out fields.
5. Seed one stale-symbol, inlining, redaction, truncation, and wrong-build-ID
   defect; require each to fail closed.
6. Measure compile latency, binary/symbol size, capture time, symbolication time,
   output bytes, and AI root-cause/repair success.

This preserves the necessary IR seam now while deferring production crash
storage, upload, retention, and symbol-server infrastructure until it earns its
cost.
