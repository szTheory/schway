# Public Repository Verification

Prepared on 2026-09-28 for the initial source publication to `szTheory/schway`.

## Current tree

- The staged identifier gate passed. Legacy module, CLI, extension, protocol, ABI, and path counts are all zero.
- The current tree has 145 tracked `.schway` source files, `cmd/schway`, `cmd/schway-repair`, and module path `github.com/szTheory/schway`.
- All nine active planning documents have one marked current-guidance region and zero legacy identifier matches.
- The existing workflow remains configured to run `evidence-aggregate` on `ubuntu-latest`.

## History topology

- `audit/commit-map.tsv` maps the 1,538 original commits one-to-one. `audit/topology-report.json` records each ordered parent mapping, the 1 root, 18 merges, all three milestone tag targets, and the mapped Phase 23 tip. These artifacts are private to the original checkout and are not included in the public clone.
- The rewritten sequence preserves every original parent list. The first additive commit records the current-tree Schway migration and planning changes, directly after the mapped Phase 23 tip.
- Public refs are limited to `main`, `M001`, `M002`, and `vM003`. Stash and Codex checkpoint refs are absent.

## Privacy and secret scan

- `audit/scan-summary.json` records the redacted full-ref and tracked-tree scan, candidate-family counts, false-positive classifications, and redaction totals. It is private to the original checkout and contains no candidate values or matching lines.
- The scan found zero user-specific home or temporary-directory paths, contact tokens in blobs/messages, owner-name tokens, SSN-like candidates, private-key headers, AWS/GitHub credential patterns, confirmed candidates, or unclassified candidates.
- The original history contained 536 co-author identity footer lines, 1,143 home-path occurrences, and 264 macOS temporary-path occurrences; these were removed. Commit and tag display names are `szTheory`; only the account-matching GitHub noreply metadata remains.
- The phone-shaped matcher produced 80 false positives: POSIX/version references (18), test execution timestamps (50), standards/documentation references (9), timing-fixture values (2), and a saved shell diagnostic line number (1). None was classified as a contact number.
- Gitleaks reported 8 `generic-api-key` matches in secret-classification tests. All are synthetic test values; there are zero confirmed or unclassified secret findings.

The final redacted scan is rerun against the verification document and every public ref immediately before pushing. The hosted Ubuntu receipt is pending the initial push; no project tests were run locally.
