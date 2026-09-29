---
quick_task: 260928-rta
status: complete
---

# Record Schway as the chosen public language name

## Goal

Make **Schway** the canonical project and language name in current identity and
planning entry points. Preserve the selection and the remaining technical
identifier migration boundary in durable planning context so future GSD work
does not reopen the naming decision or strand it halfway through a release
rename.

## Tasks

1. Update the README, project identity, and generated project context to call
   the language Schway. Update active M004 planning document titles and current
   status summaries to use that identity where the name is meant. Preserve
   historical milestone records and verification receipts.
2. Record the dated naming decision, the accepted unrelated same-name business
   result, the no-reopen rule, and a concrete pre-public-release migration list
   for technical identifiers. Keep Phase 23's CI blocker and next action intact.

## Scope boundary

This documentation task does not rename Go import paths, CLI commands, source
extensions, wire schemas, or historical fixtures. Those identifiers are listed
as one coordinated migration slice to complete after the public repository
owner/path is chosen and before a public release.

## Verification

Manually inspect the edited identity and planning documents, confirm the Phase
23 hosted Ubuntu receipt remains the active blocker, and confirm no historical
phase archive or implementation identifiers were changed. No test suite is
needed for this documentation-only task.
