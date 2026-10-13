#!/bin/sh
# Phase 32 code digest (D-08, D-08a). Prints one line: the sha256 of the
# `git ls-tree -r HEAD` listing of the code paths below, with lines ending in
# `.md` removed. The Go twin is phase32CodeDigest; this is consistency, not
# authenticity (T-32-15).
set -eu

fail() {
	printf 'codedigest: status=fail reason=%s\n' "$1" >&2
	exit 1
}

digest_paths="cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml"
mode=digest
rev=HEAD
change_a=
change_b=
while [ "$#" -gt 0 ]; do
	case "$1" in
	--paths) mode=paths; shift ;;
	--rev)
		[ "$#" -ge 2 ] || fail usage
		rev=$2; shift 2 ;;
	--changed)
		[ "$#" -ge 3 ] || fail usage
		mode=changed; change_a=$2; change_b=$3; shift 3 ;;
	*) fail usage ;;
	esac
done

if [ "$mode" = paths ]; then
	printf '%s\n' $digest_paths
	exit 0
fi

if command -v shasum >/dev/null 2>&1; then
	hasher="shasum -a 256"
elif command -v sha256sum >/dev/null 2>&1; then
	hasher="sha256sum"
else
	fail "unavailable_tool_sha256"
fi

root=$(git rev-parse --show-toplevel 2>/dev/null) || fail "no_repo_root"
[ -n "$root" ] || fail "no_repo_root"
cd "$root" || fail "no_repo_root"

if [ "$mode" = changed ]; then
	log_names=$(git -c core.quotePath=false log --format= --name-only "$change_a..$change_b" -- $digest_paths) || fail "git_log_failed"
	diff_names=$(git -c core.quotePath=false diff --name-only "$change_a" "$change_b" -- $digest_paths) || fail "git_log_failed"
	changed=$(printf '%s\n%s\n' "$log_names" "$diff_names" | { grep -v '\.md$' || [ "$?" -eq 1 ]; } | { grep -v '^$' || [ "$?" -eq 1 ]; }) || fail "git_log_failed"
	[ -n "$changed" ] && printf '%s\n' "$changed"
	exit 0
fi

listing=$(git -c core.quotePath=false ls-tree -r "$rev" -- $digest_paths) || fail "git_ls_tree_failed"
[ -n "$listing" ] || fail "empty_listing"
# Command substitution strips trailing newlines; printf '%s\n' restores exactly one.
digest=$(printf '%s\n' "$listing" | grep -v '\.md$' | $hasher | cut -d' ' -f1)
case "$digest" in
*[!0-9a-f]*) fail "bad_digest" ;;
esac
empty_hash=$(printf '' | $hasher | cut -d' ' -f1)
[ "${#digest}" -eq "${#empty_hash}" ] || fail "bad_digest"
printf '%s\n' "$digest"
