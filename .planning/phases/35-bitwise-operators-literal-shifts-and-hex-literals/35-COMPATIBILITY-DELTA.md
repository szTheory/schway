# Phase 35 compatibility delta — 2026-10-14

This is a dated, narrow after-state ledger. It does not rewrite the immutable
Phase 35 pre-change fixtures or broaden their exceptions. The source/core/C
baseline remains `testdata/phase35/prechange-corpus.json`, captured at
`1fd69337d980c4156914d47669a520fc8b04cc96` with core schemas `lang.core/0`
and `lang.core/1` and C schema `lang.c17/0`.

## Declared language and resource deltas

| Area | Frozen before-state | Current after-state | Evidence |
|---|---|---|---|
| Scalar operators | Grouped scalar `^`, `&`, `|`, `<<`, and `>>` were not admitted by the source parser. | Five exact-width U64 operators are registered; shifts require integer literals 0..63, left shift retains low 64 bits, and right shift is logical. | `35-02-SUMMARY.md`, `35-03-SUMMARY.md`, `35-04-SUMMARY.md`; checker/core/origin and interpreter/native witnesses in those summaries. |
| Numeric routes | Scalar `010` is 8 and scalar `0B10` is 2; legacy `010` is 10 and legacy `0B10` is refused. Legacy uppercase `0X` was refused. | Preserve those route-specific facts and admit uppercase `0X` on the legacy route only. No general numeric normalization or hex output is added. | `testdata/phase35/numeric-policy-baseline.json`, `testdata/phase35/numeric-delta.json`, `35-02-SUMMARY.md`. |
| Scalar expression depth | Frozen nested `byte_at` index rows at depths 63, 64, 65, and 128 all parsed, checked, and emitted. | Top-level expression starts at depth zero; each group or recursive index enters once. Groups, indices, and mixed expressions admit 63/64 and refuse 65/128 with `surface.unsupported_expression_depth`. The old successful C rows remain historical. | `numeric-delta.json` binds old revision `c4a2eec1ab47a6e5eabc7ff7ff9093e21854aa4b` to the separate current depth result. |
| Malformed numeric ownership | The frozen `0x_FF` projection had a lexer diagnostic and a parser expectation at the same full-token span: legacy 93..98, scalar 92..97. Phase 19's `literal_malformed.schway` had two diagnostics at 100..105. | Keep lexer ownership and suppress only the duplicate parser expectation at that exact `TokenUnknown` span; the Phase 19 row becomes one lexer diagnostic at 100..105. Other spans and unknown-token diagnostics remain active. | Source-hash-bound `diagnostic_ownership` and `prechange_corpus_amendment` in `numeric-delta.json`; `TestPhase35PreChangeNumericDeltaControls`. |
| New public source | `examples/bitwise.schway` did not exist in the frozen source inventory. | Add the named unused-Byte public conformance witness; it returns 264 and its alternate grouping returns 8 on both engines. | `cmd/schway/main_phase35_test.go#TestPhase35PublicConformanceWitness`; `session_phase35_values_test.go#TestPhase35FocusedValueAnswers`. |

The pre-change source fixture for Adler remains untouched. The one declared
existing-source exception is `examples/checksum/adler32.schway`: sixteen
`r = r + r` doublings become one `r = b << 16`; the bound `b <= 65520` makes
that exact, and the final `r = r + a` remains. This is a source-shape change,
not a blanket exemption for old sources, schemas, identifiers, numeric
policies, or C goldens.

| Adler projection fact | Frozen source/core | Current source/core |
|---|---|---|
| source_bytes | 571 | 385 |
| source_sha256 | `sha256:317ac13972a53eccd04a7ab84cc914aff755e77b881db6fcaf77fa5240ebcb4b` | `sha256:16490db4e53cb9db287727a8e5918745aa412659e799dd2fd53b38805bb3bef1` |
| source expression count | sixteen `r = r + r`; one final `r = r + a` | one `r = b << 16`; one final `r = r + a` |
| checked_core_schema | `lang.core/1` | `lang.core/1` |
| checked_core_sha256 | `sha256:5d1a35122edfaad5182b54b599959052b3a961e44f636982707a58f6b7d3b317` | `sha256:c3059df2d0744916a7a14f97c2fcc24efa893e512a5c28ad2746d433e330d448` |
| checked_core_semantic_id_count | 168 | 91 |
| checked_core_semantic_ids_sha256 | `sha256:3b87999dea365594280fd909e06c81e921514e787ad31e67d626f9f2c2b7562b` | `sha256:a17c5d92315cebc962362a2f8a6131548eb8ff277411fb5d7d29a0e0b80b95dc` |
| cgen_schema | `lang.c17/0` | `lang.c17/0` |

| Emitter route | Before outcome | Before bytes | Before C sha256 | After outcome | After bytes | After C sha256 |
|---|---|---:|---|---|---:|---|
| `emit` | emitted | 44595 | `sha256:d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b` | emitted | 19163 | `sha256:8653dbec998b2d39975619912b400e465b299d9d60ef59bdb1397ed9991d8e6e` |
| `emit_native` | emitted | 44595 | `sha256:d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b` | emitted | 19163 | `sha256:8653dbec998b2d39975619912b400e465b299d9d60ef59bdb1397ed9991d8e6e` |
| `emit_application` | refused (`application entry "s1`) | — | — | refused (`application entry "s1`) | — | — |

The application refusal is unchanged because this library's `ByteView -> U64`
function is not the supported linear U64 application-entry shape. The
current-C values above are recomputed from the live Adler source and compared
to the frozen C bytes by `TestPhase35EvidenceGrounding`; the frozen fixture is
never recaptured. `TestPhase35PreChangeEmitterCorpus` and its changed-byte,
identity, status, omission, and restoration controls continue to compare the
original snapshot for every recorded row. The deliberate Adler source rewrite
does not suppress or rewrite any other baseline comparison.
