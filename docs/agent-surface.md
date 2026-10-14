# Agent surface manifest

`schway surface --json` and `schway --json surface` emit the standalone
`schway.surface/1` document. Those are the only accepted argument forms. The
output is deterministic UTF-8 JSON with fixed field order, sorted arrays, and
one final newline. It reads no source file and reports no host, Clang, network,
Git, terminal, timing, metric, or command-envelope data. Invalid raw arguments
use the normal CLI usage error before global flag parsing can remove them.

The document describes byte-based half-open edit spans and 1-based line and
Unicode-scalar column positions. Its type, operator, statement, builtin, call,
restriction, limit, and refusal entries describe checked contexts in the
current compiler. They do not describe every form the parser can read or every
shape for which abilities can be derived. For example, `Box` and `Pair` can
have structural ability facts without being executable straight-line input
shapes. They are covered by the parse/ability restriction and are absent from
the admitted type list. `if/else` is advertised only in its witnessed scalar
CFG context. Bool facts do not grant general Bool literals or result tails.

Ordinary declarations take one parameter. Admitted ordinary calls also take
one argument in their checked linear binding and match-arm contexts, subject
to signature, acyclic-call-graph, type, ownership, and authority checks.
`byte_at` is a separate two-argument builtin; `byte_len`, `write_text`, and
`write_u64` each take one argument, while `write_newline` takes none. Scalar
CFG expressions do not admit helper calls. Foreign calls have their own narrow
declared policy and adapter contracts. Resource and match facts remain limited
to their checked acquisition, borrowing, branch, and discharge paths.

Limits identify their category, scope, unit, value, meaning, and applicable
command routes. Parser bounds, checker work, scalar fixed-point work, module
loading, native application streams, evidence inputs, and interface artifacts
are separate budgets. The maximum for a `FileBytes` input is selected by its
validated binding configuration; the manifest does not claim a universal
language-wide file-size cap.

The seven refusal hints are current diagnostics, not admission rules or
unchecked repair instructions. In particular, the manifest does not promise
bitwise operators, hexadecimal literals, dynamic shifts, or scalar-expression
helper calls. Use `schway check SOURCE` for the authoritative result for a
complete source program; the manifest is bounded discovery and early feedback.
