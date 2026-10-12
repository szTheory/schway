---
status: complete
phase: 30-local-modules
source: [30-00-SUMMARY.md, 30-01-SUMMARY.md, 30-02-SUMMARY.md, 30-03-SUMMARY.md, 30-04-SUMMARY.md, 30-05-SUMMARY.md, 30-06-SUMMARY.md, 30-07-SUMMARY.md, 30-08-SUMMARY.md, 30-09-SUMMARY.md]
started: 2026-10-12T12:00:00Z
updated: 2026-10-12T17:11:59.711Z
---

## Current Test

[testing complete]

## Tests

### 1. Imported Adler-32 library prints the reference checksums
expected: From the repo root, build `examples/phase30/adler32_cli.schway` with `--manifest examples/phase29/fixture_bytes.bindings.json`. `schway app run <artifact> -- N` for N = 0..9 prints the examples/phase29/README.md table values (case 3 → 300286872, case 6 → 33620352), with exit 0 and empty stderr. Case 10 prints nothing and exits 65.
result: pass
source: automated
evidence: TestPhase30AdlerModuleAppCLI (with a mutated-library control), TestPhase30AdlerModuleInterpreter, TestPhase30LibraryIsSingleCopy. Agent CLI run on 2026-10-12: all 10 cases equal the README table, case 10 printed nothing and exited 65.

### 2. A private library function cannot be imported
expected: The entry imports `fn helper` from a module whose export list names only `fn api`. `schway check` exits 2 with `module.not_exported`: "module lib does not export \"helper\"". Importing the exported `api` checks clean (`check pass`), and the built app prints 42 for input 41.
result: pass
source: automated
evidence: TestPhase30UATRefusalTranscriptsCLI (public `check`, accepted sibling per row), TestPhase30PrivateNameRefused, TestPhase30CorevalidateVisibilityPeer. Agent CLI run: private import → exit 2 `module.not_exported`. Control → `check pass`, and `app run -- 41` → 42, exit 0.

### 3. The module graph is refused with named codes
expected: Each of these exits 2 with its own code: an a↔b cycle gives `module.import_cycle` "a -> b -> a"; a missing module gives `module.not_found` naming `nothere.schway`; a header that doesn't match the path gives `module.path_mismatch`; a symlinked module file gives `module.symlink`; `import ..lib` gives `syntax.expected_module_name`.
result: pass
source: automated
evidence: TestPhase30UATRefusalTranscriptsCLI (public `check`, accepted sibling per row), TestPhase30ImportCycles, TestPhase30GraphRefusals, TestPhase30SymlinkAndEscapeRefused, TestPhase30ImportPathSyntaxRefused. Agent CLI run: each case gave exactly the expected code, exit 2, and no host path in the message.

### 4. An error inside an imported file names that file
expected: An undefined call inside `lib.schway` is reported with `unit=lib` on the diagnostic line, so it is clearly not the entry file.
result: pass
source: automated
evidence: TestPhase30UATRefusalTranscriptsCLI (public `check`, accepted sibling per row), TestPhase30ImportedFileDiagnosticNamesFile, TestPhase30UnitProjection, TestPhase30RepairDriverDeclinesCrossUnit. Agent CLI run: `core.call_callee_unresolved unit=lib`, exit 2.

### 5. Limit errors are reported by name (CR-01 fix)
expected: A 10-deep import chain gives `module.depth_limit` ("exceeds the limit of 8 modules below the entry"), and an entry importing 16 modules gives `module.count_limit` ("exceeds the limit of 16 modules, the entry included"). Neither gives the misleading `module.imports_not_loaded`.
result: pass
source: automated
evidence: TestPhase30UATRefusalTranscriptsCLI (public `check`, accepted sibling per row), TestPhase30CapsExactNAndNPlusOne, TestPhase30CapPrecedence, CR-01 route tests (11d64746). Agent CLI run: both limits gave the named code, exit 2.

### 6. Programs without imports are unchanged
expected: Every committed no-import source produces byte-identical output to the snapshot taken before any Phase 30 code changed.
result: pass
source: automated
evidence: TestPhase30PreChangeGolden, TestPhase30PreChangeGoldenIsNotInert, TestPhase272AcceptedCorpusByteIdentity, TestPhase273ConformanceCorpusByteIdentity. The full `go test ./...` is green and the focused Phase30 run was re-run green on 2026-10-12.

## Summary

total: 6
passed: 6
issues: 0
pending: 0
skipped: 0
blocked: 0

## Automation Note

Every SUMMARY coverage block was classified `all_auto_covered`. The agent ran all six checkpoints through the public CLI, and the `go test ./...` suite pins each one in CI. No human checkpoint remained (PROJECT.md § Verification Operating Preference). TestPhase30UATRefusalTranscriptsCLI (cmd/schway) pins the CLI transcripts for checkpoints 2-5 through the binary's `check` route. It was armed by swapping the depth-limit row's refused program for its accepted sibling: the test failed, and the change was reverted. One observation does not count against this phase: `schway run --engine=...` with a parameterised `main` fails as an operational failure on single-file programs too, and the built-artifact route (`build` + `app run`) is the supported input path.

## Gaps

[none]
