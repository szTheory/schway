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

Parser limits include both source-producing interface routes. Scalar analysis
also applies to the checked paths of evidence, explain, query, debug-map, and
verify. The global checker-work guard applies to debug-map only with imports
and to query only through its delegated lane/corpus queries. The linear-binding
parser guard counts the aggregate outer function body binding lists in one
program; it does not count nested statement bindings. Earlier token or source
bounds can refuse an input before that guard is reached.

The refusal hints are current diagnostics, not admission rules or unchecked
repair instructions. The manifest describes the current U64 bitwise operators
and literal shifts; it does not admit dynamic shifts or scalar-expression
helper calls. Use `schway check SOURCE` for the authoritative result for a
complete source program; the manifest is bounded discovery and early feedback.

## Scalar operator contract

Precedence is, from loosest to tightest, comparisons (`<`, `==`), bitwise OR
(`|`), XOR (`^`), AND (`&`), shifts (`<<`, `>>`), addition (`+`), then
remainder (`%`). Operators at the same tier associate left to right; `<` and
`==` share one tier.

The written expression `x ^ (y << 3) & 0xFF` means
`x ^ ((y << 3) & 0xFF)`. With `x=0x100` and `y=0x21`, it returns decimal
`264`; the alternate `(x ^ (y << 3)) & 0xFF` returns `8`. Numbers are
displayed in decimal; hexadecimal output is not added.

Shift amounts must be integer literals from `0` through `63`. Parentheses
around a numeric literal preserve that literal; names and computed expressions
remain unsupported, even when their value is known. Left shifts keep the low
64 bits; right shifts are logical. Scalar expression nesting is separately
capped at 64 levels, including nested parentheses and recursively nested
`byte_at` indices; this is not the type-nesting limit.

Hexadecimal source tokens accept lowercase and uppercase `0x`/`0X` in scalar
expressions and preserve their spelling when formatted. Numeric conversion
remains route-specific: scalar `010` is `8` and `0B10` is `2`; the legacy route
reads `010` as `10` and refuses `0B10`. Phase 35 adds uppercase `0X` to the
legacy route without making the two routes identical.

`examples/bitwise.schway` uses one entry parameter named `unused: Byte` so the
existing Byte fixture route can reach the U64/Bool witness. The parameter is
not read, copied, stored, or converted. Public conformance returns an execution
document with `outcome.value` equal to `264` on both engines. The U64 application
seam separately emits `264\n`; this is not the public conformance projection.
The existing U64 run input requirement remains in force: there is no default
U64 value or new input flag. File input through the interpreter CLI belongs to
Phase 37.

`executionpeer` checks causal event admission and fabricated-event refusal; it
does not prove numeric answers. Numeric correctness uses independently authored
literal answers and reached wrong-result controls. The public answer anchors
are `cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness` and
`internal/compiler/session/session_phase35_values_test.go#TestPhase35FocusedValueAnswers`;
the Adler consumer anchor is
`internal/compiler/session/session_phase35_adler_test.go#TestPhase35ExistingAdlerAnswers`.
