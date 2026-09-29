# Phase 21 Scoped Emitted Multi-Function LTO Evidence

**Run date:** 2026-09-26 (UTC)  
**Command:** `go test -tags=phase21_lto_evidence ./internal/compiler/session -run '^TestPhase21EmittedMultiFunctionLTOComparison$' -count=1 -v`  
**Result:** PASS; all-pairs semantic equality across interpreter, `-O0`, `-O3`, and `-O3 -flto`.

## Recorded scope

| Field | Observed value |
|---|---|
| Fixture | `testdata/phase14/multi_function_match_refusal.schway` |
| Fixture SHA-256 | `a86d9e90be4866cb20626961fc3757604bea7ac68ffff27f4eab64cc51f3e556` |
| Route | Direct `cgen.EmitProgramNativeForTest` (`emitProgram(..., true)`) |
| Emitted C SHA-256 | `d096fca69195eb63b09566387690a7b40fdc9c364f55b0b4a24209a895ff6416` |
| Host | `darwin/arm64` |
| Clang executable | `/usr/bin/clang` |
| Clang version | `Apple clang version 21.0.0 (clang-2100.1.1.101)` |
| Common flags | `-std=c17 -Wall -Wextra -Werror -pedantic` |
| Lanes | Interpreter; Clang `-O0`; Clang `-O3`; Clang `-O3 -flto` |
| Comparator | Existing `session.Phase5CompareEngines`, all engine pairs |

The table uses the current repository path. The recorded run consumed the
same fixture bytes before the public source rename, when the file used the
`.lang` suffix.

The opt-in test obtains the checked program, directly emits the multi-function fixture, and reuses the established four-tier runner and semantic comparator. The existing independent negative control also passed:

```text
go test ./internal/compiler/session -run '^TestPhase16EmitterPortSemanticGuardIsNotInert$' -count=1 -v
--- PASS: TestPhase16EmitterPortSemanticGuardIsNotInert
```

## Evidence ceiling

This run establishes semantic agreement for this one fixture, Clang build, host, and four lanes. It does not measure optimizer activity or performance. It provides no evidence about resource discharge, foreign/by-pointer emitters, other hosts, or other toolchains. The comparison is build-tagged and excluded from the recurring untagged test suite; the existing full-suite CI continues to run the inexpensive structural guards.
