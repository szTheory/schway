---
created: 2026-10-12T00:00:00Z
title: Formatter merges a trailing result identifier into the preceding var assignment
area: syntax/formatter
severity: major
files:
  - internal/compiler/syntax/format.go
  - examples/phase29/adler32_fixture.schway
  - examples/phase30/adler32.schway
---

## Problem

`schway format` changes what a program means in scalar `var` statement bodies.
A body ending in `r = r + a` followed by the result expression `r` on its own
line is rewritten to `r = r + ar`. That produces a different, invalid program.
The formatter is meant to be lossless ("source authority is formatter-owned"),
so this violates its core contract.

This was first noted in the 29-02 SUMMARY ("formatter mishandles scalar `var`
statement bodies: missing newline before the result, `else` on its own line")
and carried in the 29 and 30 transition checks as a known gap. Because of it, the
Phase 29 Adler-32 fixture and the Phase 30 `adler32` library are committed
verbatim and are not format-canonical. Nothing tracked it as work until now.

## Solution

TBD. Likely the formatter's newline handling between the last assignment
statement and the trailing result expression in a `var`-body block. Before the
fix, add a format round-trip test over `examples/phase29/adler32_fixture.schway`
and `examples/phase30/adler32.schway`, asserting that the formatted output parses
and checks to the same core. Then format both files canonically. Consider
extending the existing generated-program and lossless lanes to cover var-bodies
ending in a bare result.
