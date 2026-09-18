#!/bin/sh
set -eu

# assert-reconciliation-touched.sh -- the D-14-14 coupling rule (plan
# 14-10 Task 2): any commit range that touches an archived validation or
# verification document must also touch .planning/EVIDENCE-RECONCILIATION.md.
# The point is stated plainly: editing the archive until it's green then
# requires writing down, by name and date, that you did it -- the
# reconciliation view (and the PHASE-14-DEBT.md reconciliation entries it
# is derived from) is that record. This script never edits anything; it
# only reports.
#
# Usage: assert-reconciliation-touched.sh COMMIT_RANGE
#
# Exit 0 -- coupled: either no archived VALIDATION/VERIFICATION document
#           changed in the range, or the reconciliation view changed
#           alongside it.
# Exit 1 -- uncoupled: an archived VALIDATION/VERIFICATION document
#           changed in the range without the reconciliation view changing.
#
# These are the script's only two defined verdicts. Any other exit status
# (a bad range, a missing git repository, an unreadable diff) is an
# operational error and must never be read as a silent pass.

[ "$#" -eq 1 ] || {
	echo "usage: $0 COMMIT_RANGE" >&2
	exit 2
}
range=$1

changed=$(git diff --name-only "$range" --)

archive_changed=$(printf '%s\n' "$changed" | grep -E '^\.planning/milestones/[^/]+-phases/.*-(VALIDATION|VERIFICATION)\.md$' || true)
view_changed=$(printf '%s\n' "$changed" | grep -F -x '.planning/EVIDENCE-RECONCILIATION.md' || true)

if [ -n "$archive_changed" ] && [ -z "$view_changed" ]; then
	echo "assert-reconciliation-touched: archived evidence document(s) changed in $range without touching .planning/EVIDENCE-RECONCILIATION.md:" >&2
	printf '%s\n' "$archive_changed" >&2
	exit 1
fi

exit 0
