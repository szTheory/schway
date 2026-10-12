#!/bin/sh
# Phase 32 freeze binding gate (D-00, D-08c).
# Usage: check-phase32-freeze.sh FREEZE_COMMIT CANDIDATE [SAME_AS ...]
#
# Binds a candidate commit (a pushed HEAD, a hosted run headSha, an origin tip)
# to the frozen code. Revision = codedigest (D-08): commits after the freeze
# that touch only .planning or prose (outside the digest paths) are accepted;
# any change to a digest path is rejected. Every SAME_AS must equal CANDIDATE.
# This is consistency, not authenticity.
set -eu

phase32_freeze_paths="cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml"

reject() {
	printf 'phase32 freeze gate: status=fail reason=%s\n' "$1" >&2
	exit 1
}

[ "$#" -ge 2 ] || reject "usage"

for sha in "$@"; do
	case "$sha" in
	*[!0-9a-f]*) reject "bad_sha" ;;
	esac
	[ "${#sha}" -eq 40 ] || reject "bad_sha"
done

freeze=$1
candidate=$2
shift 2
for same in "$@"; do
	[ "$same" = "$candidate" ] || reject "sha_mismatch"
done

git cat-file -e "${freeze}^{commit}" 2>/dev/null || reject "unknown_commit"
git cat-file -e "${candidate}^{commit}" 2>/dev/null || reject "unknown_commit"

git merge-base --is-ancestor "$freeze" "$candidate" || reject "not_descendant"

# shellcheck disable=SC2086
changed=$(git log --format=%H "${freeze}..${candidate}" -- $phase32_freeze_paths) || reject "git_log_failed"
[ -z "$changed" ] || reject "digest_path_changed_after_freeze"
# shellcheck disable=SC2086
git diff --quiet "$freeze" "$candidate" -- $phase32_freeze_paths || reject "digest_path_changed_after_freeze"

printf 'phase32 freeze gate: status=pass freeze=%s candidate=%s\n' "$freeze" "$candidate"
