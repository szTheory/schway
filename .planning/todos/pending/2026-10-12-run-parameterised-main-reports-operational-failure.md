---
created: 2026-10-12T00:00:00Z
title: run --engine on a parameterised main reports an operational failure instead of a usage/source refusal
area: cmd/schway run
severity: minor
files:
  - cmd/schway/main.go
---

## Problem

`schway run --engine=interpreter FILE` and `--engine=native` on a program whose
`main` takes a parameter (`fn main(n: U64) -> U64`) exit 3 with
`tool.run_failed` / `native.tool_failure`. That is the operational exit class,
but nothing failed operationally: `run` takes no INPUT, so `main` has no
argument to bind. This happens on single-file programs as well, so it predates
Phase 30. Observed during the Phase 30 UAT on 2026-10-12.

## Expected

The program is refused before any engine starts, with a named code under the
DX-17 single exit-class table: either a usage error that points to
`build` + `app run -- INPUT`, or a source refusal for a `main` arity that `run`
does not support.

## Solution sketch

Admit `main`'s arity before engine dispatch on the `run` route. Pin it with a
cmd/schway test that covers both engines and has an armed control: a zero-input
`main` still runs.
