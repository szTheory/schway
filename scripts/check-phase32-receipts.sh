#!/bin/sh
# Phase 32 two-host receipt gate. Usage: check-phase32-receipts.sh LOG_A LOG_B
#
# Accepts exactly one macOS and one Linux log produced by
# scripts/verify-phase32.sh, both bound to one codedigest (the revision in the
# SC3 sense, D-08) and to the producer's case table and program list, which
# this gate reads from the producer so the two share one table. Anything else
# (a missing, duplicate, failed, dirty-tree, wrong-digest, incomplete or
# malformed log) exits non-zero with a named reason. The expected digest
# defaults to `sh scripts/codedigest.sh` (override:
# PHASE32_EXPECTED_CODEDIGEST); it is recomputed from this checkout, never
# copied from a log. headsha is traceability only, so a .planning-only HEAD
# move keeps a receipt current.
#
# This gate checks consistency, not authenticity: it validates the shape and
# the digest binding of the logs, but nothing here binds a log to a particular
# hosted CI run, runner, or artifact. Authenticity must come from the hosted
# run record itself (the recorded run ID in the receipt).
set -eu

reject() {
	printf 'phase32 receipt gate: status=fail reason=%s\n' "$1" >&2
	exit 1
}

[ "$#" -eq 2 ] || reject "usage_two_logs_required"

root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
cd "$root"

expected_digest=${PHASE32_EXPECTED_CODEDIGEST:-$(sh scripts/codedigest.sh)}
[ -n "$expected_digest" ] || reject "expected_codedigest_unknown"

producer="$root/scripts/verify-phase32.sh"
[ -f "$producer" ] || reject "producer_missing"
case_table=$(sed -n '/^# phase32-cases-begin$/,/^# phase32-cases-end$/p' "$producer" | grep '^# case ' || true)
[ -n "$case_table" ] || reject "case_table_empty"
table_count=$(printf '%s\n' "$case_table" | grep -c '^# case ')
want_programs=$(sed -n 's/^phase32_programs=//p' "$producer" | sed -n '1p')
[ -n "$want_programs" ] || reject "programs_line_missing"

field() { # field LINE KEY -> value of KEY=value (values never contain spaces)
	printf '%s\n' "$1" | tr ' ' '\n' | sed -n "s/^$2=//p" | sed -n '1p'
}

identity_digest() { # identity_digest LOG -> codedigest of the identity row, empty when absent
	grep '^phase32 host identity ' "$1" | sed -n '1p' | tr ' ' '\n' | sed -n 's/^codedigest=//p' | sed -n '1p'
}

check_log() { # check_log LOG -> prints host name on success
	log=$1
	[ -f "$log" ] || reject "host_missing_$log"
	[ -s "$log" ] || reject "log_empty_$log"
	if grep -Eq 'status=(fail|incomplete|unavailable|unobserved)|unavailable' "$log"; then
		reject "failed_or_unavailable_row_$log"
	fi
	identity_count=$(grep -c '^phase32 host identity ' "$log" || true)
	result_count=$(grep -c '^phase32 host result ' "$log" || true)
	[ "$identity_count" = 1 ] || reject "identity_row_count_${identity_count}_$log"
	[ "$result_count" = 1 ] || reject "result_row_count_${result_count}_$log"
	[ "$(grep '^phase32 ' "$log" | sed -n '1p' | cut -c1-20)" = "phase32 host identit" ] || reject "identity_not_first_$log"
	[ "$(grep '^phase32 ' "$log" | tail -n 1 | cut -c1-19)" = "phase32 host result" ] || reject "result_not_last_$log"

	unknown_rows=$(grep '^phase32 ' "$log" | grep -Ev '^phase32 (host identity|toolchain|case|control|host result) ' || true)
	[ -z "$unknown_rows" ] || reject "unknown_phase32_row_$log"
	[ "$(grep -c '^phase32 toolchain ' "$log" || true)" = 1 ] || reject "toolchain_row_count_$log"
	[ "$(grep -Ec '^phase32 toolchain go="[^"]+" clang="[^"]+"$' "$log" || true)" = 1 ] || reject "toolchain_row_malformed_$log"

	identity=$(grep '^phase32 host identity ' "$log")
	result=$(grep '^phase32 host result ' "$log")
	host=$(field "$identity" host)
	case "$host" in macOS|Linux) ;; *) reject "unknown_host_$log" ;; esac
	[ -n "$(field "$identity" arch)" ] || reject "arch_missing_$log"
	[ -n "$(field "$identity" clang_target)" ] || reject "target_missing_$log"
	[ -n "$(field "$identity" go)" ] || reject "go_host_missing_$log"
	[ -n "$(field "$identity" headsha)" ] || reject "headsha_missing_$log"
	[ "$(field "$identity" codedigest)" = "$expected_digest" ] || reject "digest_mismatch_$log"
	[ "$(field "$identity" tree)" = clean ] || reject "tree_not_clean_$log"
	[ "$(field "$result" host)" = "$host" ] || reject "result_host_mismatch_$log"
	[ "$(field "$result" codedigest)" = "$expected_digest" ] || reject "result_digest_mismatch_$log"
	[ "$(field "$result" tree)" = clean ] || reject "result_tree_not_clean_$log"
	[ "$(field "$result" status)" = pass ] || reject "result_not_pass_$log"
	[ "$(field "$result" programs)" = "$want_programs" ] || reject "result_programs_mismatch_$log"

	for program in $(printf '%s\n' "$want_programs" | tr ',' ' '); do
		printf '%s\n' "$case_table" | grep -q "^# case $program " || reject "program_without_table_entry_$program"
		[ "$(grep -c "^phase32 case program=$program " "$log" || true)" -ge 1 ] || reject "program_missing_${program}_$log"
	done
	[ "$(field "$result" controls)" = 4 ] || reject "result_control_count_$log"
	[ "$(field "$result" cases)" = "$table_count" ] || reject "result_case_count_$log"
	[ "$(grep -c '^phase32 case ' "$log" || true)" = "$table_count" ] || reject "case_row_count_$log"
	while read -r _hash _case program key want_exit _stdout _stderr; do
		want="phase32 case program=$program input=$key expected_exit=$want_exit actual_exit=$want_exit stdout=exact stderr=exact status=pass"
		[ "$(grep -Fxc -- "$want" "$log" || true)" = 1 ] || reject "case_missing_or_altered_${program}_${key}_$log"
	done <<EOF
$case_table
EOF
	[ "$(grep -c '^phase32 control ' "$log" || true)" = 4 ] || reject "control_row_count_$log"
	for control in guide route-audit codedigest receipt-rules; do
		[ "$(grep -Ec "^phase32 control name=$control package=[^ ]+ tests=[^ ]+( [^ ]+)* status=pass\$" "$log" || true)" = 1 ] || reject "control_missing_or_altered_${control}_$log"
	done
	printf '%s' "$host"
}

# Cross-log digest comparison runs first, so two logs at different digests are
# reported as such rather than as a mismatch against the expected digest.
if [ -f "$1" ] && [ -f "$2" ]; then
	digest_a=$(identity_digest "$1" || true)
	digest_b=$(identity_digest "$2" || true)
	if [ -n "$digest_a" ] && [ -n "$digest_b" ] && [ "$digest_a" != "$digest_b" ]; then
		reject "log_digest_differs"
	fi
fi

host_a=$(check_log "$1")
host_b=$(check_log "$2")
[ "$host_a" != "$host_b" ] || reject "duplicate_host_$host_a"
printf 'phase32 receipt gate: status=pass hosts=%s,%s codedigest=%s programs=%s cases=%s\n' \
	"$host_a" "$host_b" "$expected_digest" "$want_programs" "$table_count"
