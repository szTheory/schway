---
phase: "01"
title: Canonical Pure Spine implementation research
status: complete
created: 2026-09-03
sources_verified: 2026-09-03
---

# Phase 1 Research: Canonical Pure Spine

## Recommendation

Implement a deliberately narrow handwritten lexer/parser in Go over a lossless
green-tree-like CST, then project typed AST/core structures rather than attaching
semantic meaning to syntax nodes. Keep the first end-to-end path in one process:

```text
bytes -> tokens/trivia -> lossless CST -> AST -> resolved typed core
      -> interpreter outcome
      -> readable C17 -> clang -O0/-O3 outcomes
      -> normalized equivalence + evidence manifest
```

Use immutable value structs and explicit IDs at stage boundaries. Avoid a generic
compiler framework, parser generator, plugin API, daemon, or persistent cache in
Phase 1. Those are justified only by measured pressure after the first full path.

## Findings That Change the Plan

### Lossless syntax without importing a framework

rust-analyzer documents two properties worth copying, not its implementation:
parsing remains lossless even for invalid input, and converting source to a
syntax tree and back is a total identity operation. Its rowan-based tree stores
syntax independently from typed AST views. Lang can preserve the same boundary
with compact Go nodes containing kind, byte span, token/trivia children, and
explicit error nodes. ([rust-analyzer syntax](https://rust-analyzer.github.io/book/contributing/syntax.html),
[rust-analyzer guide](https://rust-analyzer.github.io/book/contributing/guide.html))

For S1, use a small recursive-descent parser with recovery sets at declarations,
braces, and match arms. A parser generator would introduce a dependency and still
leave CST shape, recovery, formatting, spans, and diagnostic causality to design.

### Canonical source and canonical evidence are different contracts

Source formatting preserves comments and semantic order. Evidence encoding drops
incidental formatting and binds versioned semantic facts. Do not use one
canonicalizer for both.

RFC 8785 exists because hashing/signing JSON requires invariant bytes, but full
JCS includes recursive property sorting, UTF-16 property-name ordering, and exact
ECMAScript/IEEE-754 number serialization. Phase 1 does not need that full
interoperability surface. Use a schema-specific canonical encoder over ordered Go
structs with no map or floating-point fields, UTF-8 output, and golden bytes; name
the schema `lang.evidence/0`, not “JCS.” Add a compatibility decision before any
external consumer relies on it. ([RFC 8785](https://www.rfc-editor.org/rfc/rfc8785.html))

### Deterministic properties first; fuzzing as a bounded discovery lane

Go's standard toolchain runs fuzz seed corpora during ordinary `go test` and can
run coverage-guided mutation with `go test -fuzz`. The official guidance says
targets should be fast, deterministic, and independent of persistent global
state; failures are minimized and saved as regression seeds. This fits lexer,
parser, formatter, and manifest decoders. Ordinary CI runs the seeds; scheduled
or release verification gets a fixed `-fuzztime`, never an unbounded default.
([Go fuzzing](https://go.dev/doc/security/fuzz/),
[testing package](https://pkg.go.dev/testing))

### Machine output should stream stable events, not decorated prose

Go's `test2json` demonstrates a useful precedent: newline-separated typed events
support live consumption without unnecessary buffering, and timestamps are
optional because they harm deterministic comparison. Lang S1 can emit one JSON
document for bounded commands; preserve an event-envelope seam so later long
verification runs can use JSON Lines without replacing the schema.
([Go test2json](https://go.dev/cmd/test2json/))

### Profiling and sanitizers answer separate questions

Go diagnostics warn that profiling modes can interfere with each other; collect
CPU and memory evidence in isolated runs and measure their overhead. Clang's
AddressSanitizer and UndefinedBehaviorSanitizer instrument different native
defect classes and require linking through the compiler driver. Neither proves
semantic equivalence, so keep them out of the normal S1 match oracle and run them
as separate evidence lanes. ([Go diagnostics](https://go.dev/doc/diagnostics),
[Clang AddressSanitizer](https://clang.llvm.org/docs/AddressSanitizer.html),
[Clang UndefinedBehaviorSanitizer](https://clang.llvm.org/docs/UndefinedBehaviorSanitizer.html))

## Concrete Architecture

| Area | Initial path | Contract |
|------|--------------|----------|
| Command | `cmd/lang` | thin argument/stdout/stderr/exit-code adapter |
| Orchestration | `internal/compiler/session` | one request produces diagnostics, core, outcome, metrics, and evidence |
| Syntax | `internal/compiler/syntax` | tokens include trivia; CST round-trips all bytes; AST is a typed projection |
| Core | `internal/compiler/core` | immutable nominal definitions/expressions with stable IDs and spans |
| Checking | `internal/compiler/check` | name/type/exhaustiveness results; stable diagnostic codes and cause facts |
| Interpreter | `internal/compiler/interp` | deterministic semantic outcome/events; no wall clock or host addresses |
| C backend | `internal/compiler/cgen` | readable C17 plus build request; no semantic decisions |
| Native runner | `internal/compiler/native` | explicit Clang path/target/flags/temp directory and normalized result |
| Evidence | `internal/compiler/evidence` | schema-specific canonical bytes, SHA-256 bindings, strict validation |
| Fixtures | `testdata/phase1` | positive toggle, comment stability, incomplete match, stale manifest |

Keep these packages acyclic in the direction shown. `core` must not import
`syntax`; source spans are neutral value types. `interp` and `cgen` consume core
and do not import each other. `session` is the only composition root.

## Stable Identity Strategy

Use two identity classes:

- Content IDs: `sha256:<hex>` over explicitly versioned canonical source/core/
  evidence bytes. These bind artifacts and cache inputs.
- Semantic IDs: deterministic namespace + declaration path + local structural
  ordinal for S1, serialized as opaque strings. They support diagnostic/event
  correlation but are not promised stable across arbitrary refactors yet.

Do not hash Go memory layouts, map iteration, temp paths, absolute paths, wall
times, native addresses, or human wording. Put schema version and ID algorithm
name next to every serialized ID so replacement remains possible.

## Error Recovery and Diagnostics

Lexer errors consume at least one byte. Parser recovery stops at the nearest
declaration keyword, `}`, or match-arm boundary and records an error node; it
never skips beyond the enclosing declaration. Diagnostic records contain:

- `schema`, `code`, `severity`, `primary_span`, `message_key`;
- ordered `causes` with stable kinds/spans/symbol IDs;
- bounded `related` facts and repair candidates; and
- no stack trace, absolute temp path, or raw Go error in normal invalid-source
  output.

S1 diagnostic codes begin with `syntax.*`, `name.*`, `type.*`,
`match.non_exhaustive`, `native.*`, and `evidence.*`. Exit status is a separate
command contract; never infer machine state from English text.

## Native Development Path

Generate one translation unit with a tiny versioned support header embedded by
the compiler. Compile in a fresh explicit temporary directory using `clang`
directly, capture argv/tool version/target/optimization in evidence, and normalize
the program output into the same semantic event schema as the interpreter.

The generated C must be deterministic for identical core+target input. Readable
names may include stable semantic-ID suffixes. Compile both `-O0` and `-O3` from
the same C bytes. Do not add `restrict`, undefined overflow assumptions, host
layout guesses, or runtime allocation in S1.

## Pitfalls and Preventive Gates

| Pitfall | Early signal | Prevention |
|---------|--------------|------------|
| Syntax nodes become semantic state | checker mutates CST or backend reads tokens | immutable AST/core projection and import DAG test |
| Formatter hides parser loss | reparse succeeds but comments/tokens change | byte identity CST test plus semantic-tree round trip |
| Recovery loops or swallows file | zero progress or distant diagnostics | progress assertion and declaration-bound recovery tests |
| “Canonical JSON” overclaims JCS | maps/floats appear without rules | schema-specific encoder and rejection test for unsupported shapes |
| IDs include incidental data | temp path or wording changes digest | cross-directory and wording-mutation golden tests |
| C backend defines meaning | interpreter altered to match generated C | interpreter expected outcomes authored independently first |
| Optimization drift | O0 alone passes | always compare O0 and O3 for native fixtures |
| Tool absence becomes source error | no Clang yields invalid-program status | distinct tool-failure exit code and structured `native.tool_missing` |
| Metrics perturb semantics | timestamps appear in equivalence bytes | metrics envelope excluded from semantic digest |
| Default loop grows without bound | fuzz/sanitizers run on every check | named cost lanes and fixed budgets |

## Validation Architecture

### Fast lane after every task

- `go test ./...` with fuzz seeds running as ordinary tests;
- deterministic golden source/core/diagnostic/evidence bytes;
- no network, no unbounded fuzzing, no sanitizer build.

### Phase lane

- `go test -race ./...`;
- `go vet ./...`;
- interpreter/`-O0`/`-O3` differential fixture run;
- CLI human/JSON/exit-code contract tests;
- 20-run latency/RSS/output observation with machine/toolchain metadata.

### Discovery/release lane

- time-bounded lexer/parser/formatter/manifest fuzz targets;
- focused mutation controls for missing match arm and stale manifest;
- ASan/UBSan only once unsafe/native-boundary behavior enters later phases.

### Requirement-to-evidence map

| Requirements | Primary evidence |
|--------------|------------------|
| FND-01, DX-01 | clean-checkout build and black-box CLI tests |
| FND-02, SEM-01 | deterministic core/event/diagnostic identity goldens |
| FND-03 | manifest round trip plus stale source/core/toolchain mutations |
| SYN-01..04 | CST identity, format idempotence, semantic round trip, recovery corpus |
| SEM-02 | exhaustive positive fixture and missing-arm negative control |
| INT-01, NAT-01 | interpreter/C generation unit tests plus O0/O3 black-box run |

No manual-only behavior is required in Phase 1. Human readability is reviewed
through golden artifacts, but pass/fail derives from explicit structural and
behavioral predicates.

## Sources and Provenance

External claims above were verified 2026-09-03 against primary/official sources.
Project-specific semantic and native conclusions derive from the canonical refs
in `01-CONTEXT.md` and executable Spikes 001–005. This research adds implementation
selection; it does not promote provisional syntax or performance thresholds into
a permanent language promise.
