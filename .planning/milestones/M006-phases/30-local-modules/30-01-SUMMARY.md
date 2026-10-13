---
phase: 30-local-modules
plan: 01
subsystem: compiler-syntax
tags: [go, lexer, parser, formatter, ast, import, diagnostics, repairs]

requires:
  - phase: 30-local-modules
    provides: 30-00 gate verdict PROCEED and the no-import byte-identity golden
provides:
  - ast.ImportDecl, ImportItem, Unit, Program.Imports, Program.Units, FuncDecl.Module
  - TokenImport keyword, parser.importDecl, formatter import layout (source order kept)
  - syntax.expected_import_name, syntax.import_after_export, module.duplicate_import, module.imports_limit
  - exported syntax.MaxFunctions, MaxDeclarations, MaxImportsPerModule
  - test seams exceedsOffByOneForTest, admitModulePathFormForTest, sortImportsForTest
affects: [30-04, 30-05]

actuals:
  tokens: 10600
  tasks: 3
  commits: 4
plan_head_before: 5d16a9a44f22d0fce7e3dc704c91979381e43e2a
plan_head_after: 44866fe27eba1a37577d2d53ede7bc4a188103a8
commits: 4

tech-stack:
  added: []
  patterns:
    - "unexported test-only package variable seams, each with an armed control row that fails without it"
    - "repairs only where the edit provably loses nothing (subset-only delete, whole-export rotate)"

key-files:
  created:
    - internal/compiler/syntax/syntax_phase30_test.go
    - internal/compiler/syntax/syntax_phase30_internal_test.go
  modified:
    - internal/compiler/ast/ast.go
    - internal/compiler/syntax/token.go
    - internal/compiler/syntax/lexer.go
    - internal/compiler/syntax/parser.go
    - internal/compiler/syntax/format.go

key-decisions:
  - "A stray token in an import item list (comma, `as`, `*`) is refused with syntax.expected_import_name rather than a new code, keeping the plan's closed code set."
  - "Late imports are still appended to Program.Imports after syntax.import_after_export, so recovery and the duplicate/cap checks see them."
  - "The duplicate-import repair compares against the first declaration of that module only; the first is never deleted, so the repair cannot drop a name."

patterns-established:
  - "Per-form path-syntax seam: lexer joins `/` for slash and absolute; the parser accepts leading dots for dotdot."

requirements-completed: [MOD-03, MOD-05]

coverage:
  - id: D1
    description: "import a.b { fn x  fn y } parses into Program.Imports, formats to canonical bytes and re-formats identically"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_test.go#TestPhase30ImportGrammar"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_test.go#TestPhase30FormatImport"
        status: pass
    human_judgment: false
  - id: D2
    description: "Empty list, comma, alias, glob, import-as-identifier and qualified-call forms are refused at parse time"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_test.go#TestPhase30ImportRefusals"
        status: pass
    human_judgment: false
  - id: D3
    description: "Import-after-export and subset-duplicate repairs apply cleanly; no other import refusal carries a repair"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_test.go#TestPhase30ImportRepairs"
        status: pass
    human_judgment: false
  - id: D4
    description: "16 imports parse, 17 are refused module.imports_limit; the off-by-one seam fails the exact-16 row"
    requirement: MOD-05
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_internal_test.go#TestPhase30ImportsLimitParse"
        status: pass
    human_judgment: false
  - id: D5
    description: "`..x`, `/etc/x`, `a/b` and `\"a\"` are refused by lexer or parser; each of the first three has its own armed control row"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_internal_test.go#TestPhase30ImportPathSyntaxRefused"
        status: pass
    human_judgment: false
  - id: D6
    description: "Formatter keeps import source order; generated programs round-trip losslessly and idempotently; the sort seam is observed to fail the order test"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_internal_test.go#TestPhase30FormatKeepsImportOrder"
        status: pass
      - kind: unit
        ref: "internal/compiler/syntax/syntax_phase30_test.go#TestPhase30GeneratedImportRoundTrips"
        status: pass
    human_judgment: false
  - id: D7
    description: "No-import programs are byte-identical (format, check, C): the pre-change golden and the 272/273 identity tests stay green"
    requirement: MOD-03
    verification:
      - kind: unit
        ref: "internal/compiler/session#TestPhase30PreChangeGolden"
        status: pass
    human_judgment: false

duration: 22min
completed: 2026-10-12
status: complete
---

# Phase 30 Plan 01: Import syntax, refusals, repairs and formatter Summary

**`import a.b { fn x }` through lexer, parser, AST and formatter, with parse-time refusals, two deterministic repairs, a 16-per-file cap and a path-syntax refusal that has its own armed control per shape.**

## Performance

- **Duration:** about 22 min
- **Started:** 2026-10-12T14:46:00Z
- **Completed:** 2026-10-12T15:08:00Z
- **Tasks:** 3
- **Files modified:** 7

## Accomplishments
- The AST data model for 30-04 and 30-05: `ImportDecl`, `ImportItem`, `Unit`, `Program.Imports`, `Program.Units`, `FuncDecl.Module`. Nothing reads `Units` or `Module` yet.
- `import` is now a keyword; `importDecl` parses `import path { fn x  type T }` with whitespace-separated items. No qualified call, alias, glob or re-export form parses.
- Four new refusals: `syntax.expected_import_name`, `syntax.import_after_export` (one rotate repair), `module.duplicate_import` (one delete repair, only when the later items are a subset), `module.imports_limit` (17th declaration, duplicates counted, checked before any resolution).
- `MaxFunctions = 1024`, `MaxDeclarations = 4096` exported (values unchanged) for the merged-program caps in 30-04.
- Formatter writes one module per `import` and one item per line, keeps source order, and is idempotent and lossless over 200 generated programs plus a 20 s fuzz run with no findings.

## Task Commits

1. **Task 1: Tracer, one import declaration end to end** - `f3899d2` (feat)
2. **Task 2 RED: failing tests for refusals, repairs, cap and path syntax** - `604bcac` (test)
3. **Task 2 GREEN: refusals, repairs and the imports cap** - `884119e` (feat)
4. **Task 3: formatter lanes and the no-sort control** - `44866fe` (test)

**Plan metadata:** committed with this SUMMARY (docs).

## Files Created/Modified
- `internal/compiler/ast/ast.go` - import and unit data model
- `internal/compiler/syntax/token.go`, `lexer.go` - `TokenImport`; test-only `/` joining for the slash and absolute seam forms
- `internal/compiler/syntax/parser.go` - `importDecl`, refusals, repairs, `exceeds`, exported caps, `modulePathSpan`
- `internal/compiler/syntax/format.go` - import context, `sortImportsForTest` seam
- `internal/compiler/syntax/syntax_phase30_test.go` - grammar, refusals, repairs, format golden, generated round trips, fuzz target
- `internal/compiler/syntax/syntax_phase30_internal_test.go` - path-syntax controls, 16/17 cap, import-order control

## Decisions Made
See `key-decisions` above. `syntax_test.go` is untouched (`git diff` against the plan base is empty).

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] RED commit carries seam and constant scaffolding**
- **Found during:** Task 2
- **Issue:** The behavior tests reference `syntax.MaxFunctions`, `MaxImportsPerModule` and the two unexported seams, so without declarations the test package does not compile and the failure would be a build error, not a planned assertion (INVALID_RED).
- **Fix:** The RED commit renames `maxFunctions`/`maxDeclarations` to the exported names and declares `MaxImportsPerModule` and the two seam variables, with no behavior. All 11 other failing rows then fail on their planned assertions.
- **Files modified:** `internal/compiler/syntax/parser.go`
- **Commit:** `604bcac`

**2. [Rule 1 - Bug] Test used the wrong end offset for the rotate repair span**
- **Found during:** Task 2 GREEN
- **Issue:** The test expected the repair span to end at the index of the closing brace, but the span ends after it.
- **Fix:** Expected end is the brace index plus one.
- **Files modified:** `internal/compiler/syntax/syntax_phase30_test.go`
- **Commit:** `884119e`

**3. Additional test function.** Refusal rows live in `TestPhase30ImportRefusals` (new) rather than as rows of `TestPhase30ImportGrammar`; the grammar test keeps the tracer row. The plan's verify command for Task 2 does not name the new function, so its rows run in the broader `go test ./internal/compiler/syntax` pass.

**Total deviations:** 2 auto-fixed (1 blocking, 1 bug), 1 structural note.
**Impact on plan:** None on behavior; no scope creep.

## Issues Encountered

None.

## TDD Gate Compliance

The plan is `type: execute`, not `type: tdd`. Task 2 followed RED then GREEN (`604bcac` then `884119e`).

- **RED semantic assessment:** The target tests compile and run. They fail on the planned assertions: armed path forms still produce diagnostics (`form "dotdot" armed: diagnostics = [syntax.expected_module_name], want none`), 17 imports give no `module.imports_limit`, the empty list gives no `syntax.expected_import_name`, and the late and duplicate imports carry no diagnostics or repairs. Go test output is not one of the formats `gsd_run check tdd-red-evidence` accepts (TAP, JUnit, swift-testing, unittest), so no classifier verdict was produced; the assessment above is from the unmodified `go test -v` output.
- **Task 3:** the tests passed at once. The formatter behavior (source order, one item per line) was delivered by the Task 1 tracer, so there was no red state to reach. The no-sort control (`sortImportsForTest`) is what makes the order test non-vacuous: it flips the order and the check is observed to fail.

## User Setup Required

None - no external service configuration required.

## Known Stubs

None. `Program.Units` and `FuncDecl.Module` are intentionally unread data-model fields; plan 30-04 (merge) writes them and 30-05 (check) reads them.

## Threat Flags

None. The only new surface is the parser, covered by T-30-04, T-30-05 and T-30-06. The two path-form seams are unexported test variables defaulting to `""` and are asserted empty at the start of `TestPhase30ImportPathSyntaxRefused`.

## Next Phase Readiness

30-02 (diagnostic, protocol, debugmap, repair) runs next in the same wave and owns separate files. The new codes (`syntax.expected_import_name`, `syntax.import_after_export`, `module.duplicate_import`, `module.imports_limit`) and repair kinds (`move_import_above_export`, `delete_duplicate_import`) are ready for the protocol and repair registries that 30-02 handles. 30-04 can read `Program.Imports` and use `syntax.MaxFunctions`/`MaxDeclarations`.

## Self-Check: PASSED

- All created and modified files exist on disk.
- Commits `f3899d2`, `604bcac`, `884119e`, `44866fe` are ancestors of HEAD.
- Verification run: `go test ./internal/compiler/syntax -count=1` and the session meta set (Inventory, Groundedness, ValidationLifecycle, PreChangeGolden, 272/273 identity) all pass.

---
*Phase: 30-local-modules*
*Completed: 2026-10-12*
