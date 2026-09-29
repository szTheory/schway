# Quick Task 260928-tzu: Prepare a Public-Ready Schway Repository — Context

<!-- schway-current:start -->
Current publication identity (2026-09-28): Schway uses the Go module `github.com/szTheory/schway`, the `schway` and `schway-repair` commands, `.schway` source files, `schway.*` and `schway:*` protocol identifiers, and `schway_` and `SCHWAY_` native ABI symbols. Preserve older spellings only where they document historical implementation evidence.
<!-- schway-current:end -->


**Gathered:** 2026-09-28
**Status:** Ready for planning

<domain>
## Task Boundary

Prepare the repository for its first public source push by completing the
Schway technical rename and scrubbing personal information from the publishable
worktree and reachable Git history. Keep the commit sequence and parent
topology intact. Do not publish unless the final privacy review is clean.

</domain>

<decisions>
## Implementation Decisions

### Public identity
- GitHub owner: `szTheory`.
- Canonical repository: `github.com/szTheory/schway`.
- The repository already exists as public and empty.
- The public source extension will be `.schway`.

### History and publication
- Rewrite a separate local copy so the existing checkout and its Git metadata
  remain recoverable.
- Preserve commit order, parent relationships, merge structure, milestone tags,
  and commit count for the publishable history. Rewritten commits will have new
  hashes.
- Promote the latest current Phase 23 tip to public `main`; retain milestone
  tags and omit local stash/tool checkpoint refs from publication.
- Preserve historical source semantics except for privacy redactions. Apply
  the coordinated technical rename to the current tree in a final migration
  commit rather than rewriting every historical snapshot into a different
  program.
- The user authorized the first push only when no PII is present.

### Privacy
- Do not upload personal information, secrets, credentials, or identifying
  local paths.
- Redact identifying absolute home-directory prefixes throughout the reachable
  history and current tree. Normalize personal commit display names to the
  public GitHub handle while retaining a matching GitHub noreply address.
- Treat ambiguous personal-data matches conservatively and block publication
  until classified or removed.

</decisions>

<specifics>
## Specific Ideas

- The existing Phase 23 GitHub Actions workflow provides the hosted Ubuntu
  evidence receipt after a clean source push.
- Current `main` is an ancestor of the current Phase 23 tip, so promoting the
  latest tip to public `main` retains that full project sequence.

</specifics>

<canonical_refs>
## Canonical References

- `.planning/PROJECT.md` — chosen name and migration boundary.
- `.planning/STATE.md` — Phase 23 hosted CI blocker and next action.
- `.github/workflows/ci.yml` — existing hosted `evidence-aggregate` matrix.
- `scripts/verify-phase23.sh` — focused Phase 23 receipt command.

</canonical_refs>
