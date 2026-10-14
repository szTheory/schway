---
phase: 34-agent-diagnostics-floor-and-surface-manifest
reviewed: 2026-10-14T01:46:39Z
depth: standard
files_reviewed: 46
files_reviewed_list:
  - cmd/schway-repair/repair_phase34_test.go
  - cmd/schway/main.go
  - cmd/schway/main_phase32_routes_test.go
  - cmd/schway/main_phase34_diagnostics_test.go
  - cmd/schway/main_phase34_routes_test.go
  - cmd/schway/main_phase34_surface_test.go
  - docs/agent-surface.md
  - docs/diagnostics.md
  - internal/compiler/ast/ast.go
  - internal/compiler/check/check.go
  - internal/compiler/check/check_phase34_refusal_test.go
  - internal/compiler/check/scalar.go
  - internal/compiler/diagnostic/diagnostic.go
  - internal/compiler/diagnostic/diagnostic_phase34_test.go
  - internal/compiler/protocol/presentation.go
  - internal/compiler/protocol/protocol.go
  - internal/compiler/protocol/protocol_phase34_test.go
  - internal/compiler/session/session.go
  - internal/compiler/session/session_app_verify.go
  - internal/compiler/session/session_modules.go
  - internal/compiler/session/session_phase30_inventory_test.go
  - internal/compiler/session/session_phase34_compat_test.go
  - internal/compiler/session/session_phase34_prechange_test.go
  - internal/compiler/session/session_phase34_presentation_test.go
  - internal/compiler/session/session_phase34_routes_test.go
  - internal/compiler/session/session_phase34_surface_test.go
  - internal/compiler/session/session_phase6_explain.go
  - internal/compiler/session/session_phase6_query.go
  - internal/compiler/session/session_presentation.go
  - internal/compiler/surface/refusals.go
  - internal/compiler/surfacemanifest/manifest.go
  - internal/compiler/surfacemanifest/manifest_test.go
  - internal/compiler/syntax/lexer.go
  - internal/compiler/syntax/parser.go
  - internal/compiler/syntax/parser_phase34_baseline_test.go
  - internal/compiler/syntax/parser_phase34_recovery_test.go
  - internal/compiler/syntax/parser_phase34_refusal_test.go
  - testdata/phase16/public-emitter-consumers.json
  - testdata/phase34/compat-deltas.json
  - testdata/phase34/draft-after.json
  - testdata/phase34/draft-baseline.json
  - testdata/phase34/drafts.json
  - testdata/phase34/identity-baseline.json
  - testdata/phase34/prechange-corpus.json
  - testdata/phase34/surface-census.json
  - testdata/phase34/surface-witnesses.json
findings:
  critical: 0
  warning: 3
  info: 0
  total: 3
status: issues_found
---

# Phase 34: Code Review Report

**Reviewed:** 2026-10-14T01:46:39Z  
**Depth:** standard  
**Files Reviewed:** 46  
**Status:** issues_found

## Summary

The diagnostics, presentation, recovery, and refusal paths reviewed cleanly under the configured scope. The current-surface manifest has inaccurate route applicability and a parser-limit scope; the independent surface tests repeat those claims without checking the actual route consumers. A narrow CLI probe confirmed both interface source routes parse and check malformed source while remaining absent from the published route lists.

## Narrative Findings (AI reviewer)

### WR-01: Source-checking routes are missing from limit applicability

**Severity:** WARNING  
**File:** `internal/compiler/surfacemanifest/manifest.go:111-112`

**Issue:** `parserRoutes` omits `interface core`, although `InterfaceCoreCommandFile` reads source and calls `Check` (`session.go:1815-1821`). `checkedRoutes` omits both `interface export` and `interface core`, which also call `Check` and therefore enforce `check.work` and the scalar analysis bound (`session.go:1758-1765,1815-1821`). It also omits `debug-map` for `scalar.analysis_work`, even though that command parses and invokes `check.Program` (`session.go:1919-1927`). The reached `surface --json` output exposed these omissions; malformed-source probes through `interface export` and `interface core` each returned `invalid` with `syntax.expected_linear_result`.

**Fix:** Add `interface core` to `parserRoutes`; add `interface export` and `interface core` to the checked routes; include `debug-map` in the scalar analysis route list. Keep `check.work` scoped to paths that actually enforce its global guard.

### WR-02: Linear-binding limit reports per-body scope while enforcing an aggregate

**Severity:** WARNING  
**File:** `internal/compiler/surfacemanifest/manifest.go:200`

**Issue:** The manifest says `parser.linear_bindings` is a maximum for “one parsed linear body,” but `programLimit` sums `len(function.Body.Linear.Bindings)` across every function in the parsed program before comparing the total with `maxLinearBindings` (`syntax/parser.go:151-160`). The scope and meaning therefore describe a different limit than the parser enforces. The literal expectations repeat the same per-body claim in `manifest_test.go:211` and `surface-witnesses.json:32`.

**Fix:** Describe the scope as the aggregate outer linear-body bindings across one parsed program and update both literal expectations. Keep the meaning precise about which AST binding lists are counted.

### WR-03: Surface limit witnesses do not independently check route behavior

**Severity:** WARNING  
**File:** `internal/compiler/session/session_phase34_surface_test.go:152-168,389-425`

**Issue:** Source witnesses are executed through `CheckCommandFile`, while limit route, scope, and value expectations are compared as literals against `surfacemanifest.Current()`. `manifest_test.go:237-239` and `surface-witnesses.json:29-38` repeat the same route and scope claims. No witness drives an affected CLI route or reaches the relevant parser/check bound, so these tests pass despite the route and scope mismatches above. This leaves the manifest-to-consumer direction of the drift check unverified.

**Fix:** Add independent route/limit witnesses that exercise the actual source-consuming commands and the relevant bounds, and compare their observed behavior to separately authored expected route applicability and limit facts. Preserve the existing manifest mutation controls.

---

_Reviewed: 2026-10-14T01:46:39Z_  
_Reviewer: the agent (gsd-code-reviewer)_  
_Depth: standard_
