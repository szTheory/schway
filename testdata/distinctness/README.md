# `testdata/distinctness/`

Corpus for the diagnostic-distinctness gate (DX-08,
`internal/compiler/check/diagnostic_distinctness_test.go`). Every member here
is a **parse-failing** whole-source-file fixture: none of them parses, all are
refused by `lang --json check` with a non-zero exit and at least one
diagnostic.

## Corpus distinctness predicate

Two corpus members are structurally distinct iff their non-trivia
**token-kind sequences** (as computed by `internal/compiler/syntax.Lex` alone)
differ. This predicate is total (defined on every byte string, including ones
that never parse), rename-immune by construction (every identifier lexes to
`TokenIdentifier`, so the corpus cannot be padded with `foo`/`bar` renames),
and independent of the instrument under test (it never touches `parseProgram`
or `check`).

## Members

- `spiral_full.schway`, `spiral_narrow.schway`, `spiral_bare.schway` — the spiral
  trio (`if v { v } else { v }`, `if v { v }`, `if v`), sharing an identical
  `module spiral` / `export {}` preamble so `primary_span` lands at the exact
  same byte offset in all three. Before the fix these three collide on one
  diagnostic ID and one `result:` ID (see `collision_control.json`). After the
  fix they yield three distinct diagnostic-ID sets and three distinct result
  IDs.
- `for_range.schway` — a for-style iteration form over a range; not in the
  language.
- `arithmetic.schway` — an arithmetic expression; not in the language (no
  numeric-literal lexing exists yet).
- `unclosed_brace.schway` — a function declaration whose body brace is never
  closed (`syntax.expected_rbrace`, not the declaration-recovery arm).
- `stray_semicolon.schway` — a bare `;` at declaration level.
- `lone_else.schway` — an `else` block with no leading conditional.
- `empty_file.schway` — a zero-byte source file.
- `one_token_hard_case.schway` — a valid leading `fn` declaration followed by
  two bare identifiers at EOF. The declaration-recovery cause discards exactly
  one token here (the same byte-width as `spiral_bare.schway`'s one-token
  region), making this the adversarial near-collision case: same skipped
  region size as the third spiral member, different diagnostic content, so
  distinctness must come from more than region length alone.
- `dangling_pipe.schway` — a `data` declaration with zero alternatives
  (`syntax.expected_alternative`, a different parser arm entirely).

## `collision_control.json` — frozen, never regenerated

This file is a **historical artifact**. It records, byte-for-byte, the
pre-fix three-way collision on the spiral trio: identical diagnostic ID and
identical `result:` ID across `spiral_full.schway`, `spiral_narrow.schway`, and
`spiral_bare.schway`, captured from the tree **before** the `skipped_region`
cause fix landed in `internal/compiler/syntax/parser.go`
(`git rev-parse HEAD` at capture time is recorded inside the file).

**This file must never be regenerated, and nothing in this tree may offer a
path to regenerate it.** `TestDiagnosticDistinctnessGuardIsNotInert`
(`internal/compiler/check/diagnostic_distinctness_test.go`) reads it and
asserts a *property* over it — the metric reports the collided value (1/3),
not the healthy value (3/3) — never a golden-match replay of the fixtures
themselves. If the spiral trio's live behavior ever changes, this frozen
capture stays exactly as captured; only a brand-new, separately-dated capture
file (never a rewrite of this one) could replace it, and doing so requires the
same recorded-decision discipline as removing a corpus member (below).

## Corpus membership is additive-only

Removing a member from this directory to raise the distinctness metric is a
reviewable deletion in the diff and requires a recorded decision — the same
discipline as re-pinning a published diagnostic ID. A near-duplicate addition
(e.g. an identifier-renamed copy of an existing member) is refused by
`TestDistinctnessCorpusMembersAreStructurallyDistinct`, which fails the
**corpus**, not the compiler, when two members share a non-trivia token-kind
sequence.
