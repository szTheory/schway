---
phase: 15-event-identity-lang-execution-2
plan: 01
status: complete
---

# Phase 15 Plan 01 Summary

Published the foundational `lang.execution/2` model and admission contract.

## Delivered

- Added `Schema2`, optional `/2` event fields, and frozen `/0` and `/1`
  canonical-byte controls.
- Added canonical invocation format/parse support for
  `inv:entry:{E(entryID)}(/{E(opCallID)}#{ordinal})*`, with strict uppercase
  byte escaping and non-canonical spelling refusal.
- Added native `/2` validation: required canonical invocation, `(invocation,
  id)` uniqueness, closed event-kind admission, and caller-owned
  `function.called.callee_function_id` validation while retaining legacy
  ID-only uniqueness.
- Pinned the pre-producer shared-leaf diamond collision and native refusal.

## Verification

Passed:

`go test ./internal/compiler/execution ./internal/compiler/native ./internal/compiler/session -run 'TestInvocationGrammar|TestExecutionLegacyBytesFrozen|TestValidateExecutionSchema2|TestPhase15DiamondFrontierIsPinned' -count=1 -v`

`go build ./...`

`go vet ./internal/compiler/session/...`

The Go build cache was redirected to a writable temporary directory because
the host cache is outside this worktree sandbox.
