---
status: complete
phase: 28-exit-classes-verify-report-and-entry-pruning
source: [28-01-SUMMARY.md, 28-02-SUMMARY.md, 28-03-SUMMARY.md, 28-04-SUMMARY.md, 28-05-SUMMARY.md, 28-06-SUMMARY.md, 28-07-SUMMARY.md]
started: 2026-10-11T03:21:51Z
updated: 2026-10-11T18:50:36.609Z
---

## Current Test

[testing complete]

## Tests

### 1. Uncalled function no longer blocks check, build and run
expected: From the repo root, build the CLI (`go build -o /tmp/schway ./cmd/schway`). `/tmp/schway check examples/phase28/unused_function.schway` exits 0, `/tmp/schway build examples/phase28/unused_function.schway --output /tmp/uf` exits 0, and `/tmp/schway app run /tmp/uf -- 7` prints 7 and exits 0.
result: pass

### 2. Ambiguous entry is a source refusal
expected: `/tmp/schway check testdata/phase12/payload_borrow_interaction.schway` and `/tmp/schway run --engine=interpreter testdata/phase12/payload_borrow_interaction.schway` each exit 2 with first diagnostic `core.entry_ambiguous`, naming the roots `choose` and `identity` and suggesting adding `fn main`.
result: pass

### 3. Same source gets the same exit class on every surface
expected: For `testdata/phase07/duplicate_function_name.schway` (relative path, from the repo root), `check`, `build ... --output /tmp/dup`, `run --engine=interpreter` and `run --engine=native` all exit 2 with `core.duplicate_function_id`, and the build leaves no /tmp/dup file.
result: issue
reported: "Accepted inline result: `schway run --engine=native testdata/phase07/duplicate_function_name.schway` (relative path) exits 3 native.tool_failure; the absolute path exits 2 core.duplicate_function_id. check, build (no artifact left) and run --engine=interpreter all exit 2 correctly. TestExitClassesUniformAcrossSurfaces uses absolute paths only, so it misses this."
severity: major

### 4. App verify writes a schema /2 report
expected: `/tmp/schway app verify examples/phase22/identity.schway --cases examples/phase22/identity.cases.json --report /tmp/r.json` exits 0, and /tmp/r.json starts with `"schema":"lang.app-verification/2"` and `"status":"pass"`.
result: pass

### 5. App verify source refusal exits 2 and still writes the report
expected: `/tmp/schway app verify testdata/phase12/payload_borrow_interaction.schway --cases examples/phase22/identity.cases.json --report /tmp/r2.json` exits 2, prints `schway app verify: source refused: core.entry_ambiguous` on stderr, and /tmp/r2.json exists as a fail report.
result: pass

### 6. README publishes the exit-code table
expected: README.md has an "Exit codes" section with two tables. The compiler table lists Success 0, Unrepaired 1, Invalid 2, Operational 3, Mismatch 4 and Usage 64. The `schway app run` table lists AppInputRefused 65, EventCapacity 69, AppTimeout 124 and AppLaunchError 125, and the section notes that the child's exit status passes through.
result: pass

## Summary

total: 6
passed: 5
issues: 1
pending: 0
skipped: 0
blocked: 0

## Gaps

```yaml
- gap_id: G-28-3
  truth: "For testdata/phase07/duplicate_function_name.schway (relative path, from the repo root), check, build, run --engine=interpreter and run --engine=native all exit 2 with core.duplicate_function_id, and the build leaves no output file."
  status: failed
  reason: "User reported: run --engine=native with a relative path exits 3 native.tool_failure instead of 2 core.duplicate_function_id; absolute path is correct"
  severity: major
  test: 3
  root_cause: "runNativeCommand (internal/compiler/session/session.go:1556-1559) calls filepath.Rel(nat03ProjectRoot(), path) before Check(source) at :1560. nat03ProjectRoot (session_phase5_alias.go:29-31) is an absolute build-time path from runtime.Caller; filepath.Rel errors when one side is relative, and RunNativeCommandFile (:1588-1601) maps that error to native.tool_failure / exit 3. Check never runs. The computed value feeds Phase16ControlNativeC's fixture argument (session_phase5.go:26), which ignores it. Scope: every relative `run --engine=native` fails, valid programs included (call_basic.schway relative -> exit 3, absolute -> exit 0)."
  artifacts:
    - path: internal/compiler/session/session.go
      issue: "runNativeCommand (1551-1577) does fatal filepath.Rel before Check; RunNativeFile (1536-1546) has the same latent pattern"
    - path: internal/compiler/session/session_phase5_alias.go
      issue: "nat03ProjectRoot is a build-time runtime.Caller path used on the shipped CLI path; breaks for -trimpath or relocated binaries"
    - path: internal/compiler/native/native_phase28_test.go
      issue: "TestExitClassesUniformAcrossSurfaces (106-150) uses only absolute testsupport.ProjectPath inputs"
  missing:
    - "Remove the fatal filepath.Rel from runNativeCommand (and RunNativeFile) or make it non-fatal after Check; fixture is unused"
    - "Drop the runtime dependency on nat03ProjectRoot from the CLI production path"
    - "Add a relative-path row of all four surfaces to TestExitClassesUniformAcrossSurfaces (child cwd is internal/compiler/native: use ../../../testdata/... or a Dir-aware RunCLI variant)"
    - "Add a positive control: relative run --engine=native testdata/phase07/call_basic.schway exits 0"
  debug_session: "inline diagnosis 2026-10-11; other surfaces (build, app run, app verify, evidence, debug-map) checked clean with relative paths"
```
