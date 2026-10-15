---
phase: "37"
status: complete
fixed: 1
skipped: 0
---

# Phase37 review fix

## Fixed Issues

### WR-01: Usage Syntax Suggests Interpreter Accepts a Manifest

The usage constant and its independent expected grammar now list interpreter without a manifest, and native with a required manifest. Parser authority/exit64 behavior is unchanged. Fresh exact command passed all3 roots and subtests with no skips:

`GOCACHE=/tmp/ai-lang-verification-gocache go test ./cmd/schway -run '^(TestPhase37FileRunGrammar|TestPhase37FileRunRouteInventoryControls|TestPhase34DiagnosticRouteInventory)$' -count=1 -v -timeout=600s`

Changed files: cmd/schway/main.go, cmd/schway/main_phase37_file_run_test.go. This small within-phase review correction adds no runtime behavior, feature, dependency or wire field.
