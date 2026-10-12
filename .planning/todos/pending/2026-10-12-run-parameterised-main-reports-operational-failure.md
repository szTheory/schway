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

The current class is pinned deliberately. The Phase 28 exit-class manifest
(`internal/compiler/session/session_phase28_manifest_test.go`, e.g.
`examples/phase28/unused_function.schway`) pins `InterpreterRun` as exit 3
`tool.run_failed` and `NativeRun` as exit 3 `native.tool_failure`. Every
admitted `main` takes a parameter, and `fn main()` does not parse, so `run`
cannot execute any admitted program today. This is a DX decision, not a
regression: changing it means re-pinning that manifest.

## Expected

The program is refused before any engine starts, with a named code under the
DX-17 single exit-class table: either a usage error that points to
`build` + `app run -- INPUT`, or a source refusal for a `main` arity that `run`
does not support.

## Solution sketch

Either give `run` an INPUT argument (as `app run -- INPUT` has), or refuse
before engine dispatch with a named code that points to `build` + `app run`.
Re-pin the Phase 28 manifest rows in the same change. Armed control: a source
refusal on `run` still exits 2 with its check code.
