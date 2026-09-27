# Deferred Items

- `GOCACHE=/tmp/ai-lang-verification-gocache go test ./internal/compiler/session` fails in the pre-existing `TestPhase16EmitterCutsAreAmendedAndOwned`: it reads the removed `.planning/REQUIREMENTS.md` instead of the archived M003 requirements. The same gap is recorded in `21-VERIFICATION.md` criterion 4. This is outside plan 21-06's validation-corpus and groundedness-reconciliation scope; a separate archive-path repair is needed.
- The regenerated corpus honestly records a failing `internal/compiler/cgen` result for `TestProgramBorrowedByPointerDisposition`, which belongs to pending plan 21-05. Re-run the producer after that plan's changes are integrated to refresh this observation.
