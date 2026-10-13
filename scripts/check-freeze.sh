#!/bin/sh
# Generic freeze consistency gate. This checks consistency, not authenticity.
# Usage: check-freeze.sh MANIFEST FREEZE CANDIDATE [SAME_AS ...]
set -eu

label=
reject() {
	if [ -n "$label" ]; then prefix="$label "; else prefix=; fi
	printf '%sfreeze gate: status=fail reason=%s\n' "$prefix" "$1" >&2
	exit 1
}

[ "$#" -gt 0 ] || reject usage
case $1 in /*) manifest=$1 ;; *) manifest=$PWD/$1 ;; esac
[ -r "$manifest" ] || reject manifest_missing
mf_get() { sed -n "s/^$1=//p" "$manifest" | sed -n '1p'; }
label=$(mf_get label)
[ -n "$label" ] || reject manifest_missing
shift
[ "$#" -ge 2 ] || reject usage

root=$(git rev-parse --show-toplevel 2>/dev/null) || reject no_repo_root
[ -n "$root" ] || reject no_repo_root
cd "$root" || reject no_repo_root

for sha in "$@"; do
	case "$sha" in *[!0-9a-f]*) reject bad_sha ;; esac
	[ "${#sha}" -eq 40 ] || reject bad_sha
done

freeze=$1
candidate=$2
shift 2
for same in "$@"; do
	[ "$same" = "$candidate" ] || reject sha_mismatch
done

git cat-file -e "${freeze}^{commit}" 2>/dev/null || reject unknown_commit
git cat-file -e "${candidate}^{commit}" 2>/dev/null || reject unknown_commit
git merge-base --is-ancestor "$freeze" "$candidate" || reject not_descendant

changed=$(sh "$root/scripts/codedigest.sh" --changed "$freeze" "$candidate" 2>/dev/null) || reject git_log_failed
[ -z "$changed" ] || reject digest_path_changed_after_freeze
printf '%s freeze gate: status=pass freeze=%s candidate=%s\n' "$label" "$freeze" "$candidate"
