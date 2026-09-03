---
phase: 02-owned-values-and-abilities
fixed_at: 2026-09-03T23:36:27Z
review_path: .planning/phases/02-owned-values-and-abilities/02-REVIEW.md
iteration: 5
findings_in_scope: 3
fixed: 3
skipped: 0
status: all_fixed
---

# Phase 02: Post-Gate Fix Report

**Fixed at:** 2026-09-03T23:36:27Z
**Iteration:** 5 (compatibility closure after independent verification)

**Summary:**

- Independently found blockers in scope: 3
- Fixed: 3
- Skipped: 0

## Fixed Issues

### PG-01: Generated categories can collide in C's ordinary identifier namespace

**Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_test.go`, `internal/compiler/session/session.go`, `internal/compiler/session/session_test.go`, `testdata/phase1/generated.golden.c`, `testdata/phase1/evidence.golden.json`, `testdata/phase2/owned_transfer.golden.c`, `testdata/phase2/evidence.golden.json`
**Commit:** 91a6206
**Applied fix:** The interim ordinal-family change closed the whole generated ordinary namespace, including the valid-source counterexample `Thing`, alternative `thing`, and function `thing_LANG_THING`. Cross-category, Unicode, case, suffix, and shadow collisions executed successfully through interpreter, C17 `-O0`, and C17 `-O3`. Its unconditional golden migration was later rejected by independent verification and superseded by PG-03.

### PG-02: Evidence tool-identity probes buffer untrusted subprocess output without limits

**Files modified:** `internal/compiler/evidence/evidence.go`, `internal/compiler/evidence/toolprobe_test.go`
**Commit:** a8c14da
**Applied fix:** Both compiler-version and target-identity subprocesses now use timeout-bound commands with independently bounded stdout and stderr. Stable errors distinguish timeout, stdout overflow, stderr overflow, and ordinary tool failure. Helper-process tests flood or block each of the two probes separately.

### PG-03: Global collision repair changed frozen Phase 1 output

**Files modified:** `internal/compiler/cgen/cgen.go`, `internal/compiler/cgen/cgen_test.go`, `internal/compiler/session/session.go`, `internal/compiler/session/session_test.go`, and the Phase 1/2 C and evidence goldens
**Commit:** cbba405
**Applied fix:** Replaced unconditional ordinal renaming with one global allocator that keeps legacy source-derived C identifiers when unique and adds deterministic category/ordinal suffixes only on collision. The exact `Thing` / `thing` / `thing_LANG_THING` case and neighboring Unicode, case, suffix, and shadow cases remain green. Phase 1 and pre-fix Phase 2 golden bytes were restored exactly; a direct diff against `91a6206^` is empty.

## Verification

Verification ran in the main checkout because `.planning/config.json` sets `workflow.use_worktrees` to `false`.

- Focused cross-category collision native tests — pass
- Focused version/target stdout, stderr, and timeout tests — pass
- `go test ./...` — pass
- `go test -race ./...` — pass
- `go vet ./...` — pass
- `sh scripts/verify-phase2.sh` — pass; Phase 1 work 18, Phase 2 work 51, all eight negative controls present
- `git diff 91a6206^ -- testdata/phase1/generated.golden.c testdata/phase1/evidence.golden.json testdata/phase2/owned_transfer.golden.c testdata/phase2/evidence.golden.json --exit-code` — pass; frozen bytes restored
- `git diff --check` — pass

---

_Fixed: 2026-09-03T23:36:27Z_
_Fixer: Codex orchestration plus gsd-code-fixer_
_Iteration: 5_
