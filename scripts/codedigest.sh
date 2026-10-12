#!/bin/sh
# Phase 32 code digest (D-08, D-08a). Prints one line: the sha256 of the
# `git ls-tree -r HEAD` listing of the code paths below, with lines ending in
# `.md` removed. It reads HEAD, not the worktree, so a receipt binds to code
# and a .planning-only or prose commit does not move it. The Go twin is
# phase32CodeDigest (session_phase32_receipt_test.go); TestPhase32CodeDigest
# proves the two agree. This is consistency, not authenticity (T-32-15).
set -eu

fail() {
	printf 'codedigest: status=fail reason=%s\n' "$1" >&2
	exit 1
}

if command -v shasum >/dev/null 2>&1; then
	hasher="shasum -a 256"
elif command -v sha256sum >/dev/null 2>&1; then
	hasher="sha256sum"
else
	fail "unavailable_tool_sha256"
fi

listing=$(git -c core.quotePath=false ls-tree -r HEAD -- cmd internal examples native scripts testdata go.mod .github/workflows/ci.yml) || fail "git_ls_tree_failed"
# Command substitution strips trailing newlines; printf '%s\n' restores exactly one.
printf '%s\n' "$listing" | grep -v '\.md$' | $hasher | cut -d' ' -f1
