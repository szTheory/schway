---
quick_id: 260924-tsl
status: planned
mode: validate
---
# Fix Phase 19 Full Suite Regressions

## Objective
Repair the three independently identified full-suite regressions while retaining their existing guard semantics.

## Must-haves
- `LANGUAGE-MATURITY.md` matches mechanically derived corpus and guard counts.
- All eight Phase 19 validation commands are pinned as owned R2b entries with Phase 20 landing ownership.
- Public emitter consumer inventory matches actual source call sites.
- Focused tests and `go test ./...` pass.

## Tasks
1. Derive and refresh maturity counts; run its focused session tests.
2. Add eight evidence-backed R2b entries and P20 ownership; run groundedness tests.
3. Reconcile the public emitter consumer registry against actual call sites; run inventory test.
4. Run `go test ./...`, inspect diff, update quick task summary and state.

## Verification
Use project test commands for the three affected areas, then `go test ./...`.
