#!/bin/sh
# Phase 27 two-host receipt gate. Usage: check-phase27-receipts.sh LOG_A LOG_B
#
# Accepts exactly one macOS and one Linux receipt produced by
# scripts/verify-phase27.sh, both bound to one revision and to the declared
# source/fixture hashes and input set. Anything else (a missing, duplicate,
# unavailable, failed, dirty-tree, wrong-revision, altered-hash, incomplete, or
# malformed receipt) exits non-zero with a named reason. The expected revision
# defaults to this checkout's HEAD (override: PHASE27_EXPECTED_REVISION) and the
# expected hashes are recomputed from this checkout, never copied from a log.
set -eu

reject() {
	printf 'phase27 receipt gate: status=fail reason=%s\n' "$1" >&2
	exit 1
}

[ "$#" -eq 2 ] || reject "usage_two_logs_required"

root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
cd "$root"

expected_revision=${PHASE27_EXPECTED_REVISION:-$(git rev-parse HEAD)}
[ -n "$expected_revision" ] || reject "expected_revision_unknown"

hash_list() {
	out=
	for path in $1; do
		[ -f "$path" ] || reject "declared_input_missing_$path"
		out="$out${out:+,}$path=$(git hash-object "$path")"
	done
	printf '%s' "$out"
}
expected_sources=$(hash_list "examples/fizzbuzz.schway examples/sum_to_n.schway")
expected_fixtures=$(hash_list "examples/phase27/fizzbuzz-15.out examples/phase27/fizzbuzz-1000.out")
expected_inputs="fizzbuzz:0,1,15,1000,1001;sum_to_n:0,1,15,1000,1001"

field() { # field LINE KEY -> value of KEY=value (values never contain spaces)
	printf '%s\n' "$1" | tr ' ' '\n' | sed -n "s/^$2=//p" | sed -n '1p'
}

check_log() { # check_log LOG -> prints host name on success
	log=$1
	[ -f "$log" ] || reject "log_missing_$log"
	[ -s "$log" ] || reject "log_empty_$log"
	if grep -Eq 'status=(fail|incomplete|unavailable|unobserved)|unavailable' "$log"; then
		reject "failed_or_unavailable_row_$log"
	fi
	identity_count=$(grep -c '^phase27 host identity ' "$log" || true)
	result_count=$(grep -c '^phase27 host result ' "$log" || true)
	[ "$identity_count" = 1 ] || reject "identity_row_count_${identity_count}_$log"
	[ "$result_count" = 1 ] || reject "result_row_count_${result_count}_$log"
	[ "$(grep '^phase27 ' "$log" | sed -n '1p' | cut -c1-20)" = "phase27 host identit" ] || reject "identity_not_first_$log"
	[ "$(grep '^phase27 ' "$log" | tail -n 1 | cut -c1-19)" = "phase27 host result" ] || reject "result_not_last_$log"

	identity=$(grep '^phase27 host identity ' "$log")
	result=$(grep '^phase27 host result ' "$log")
	host=$(field "$identity" host)
	case "$host" in macOS|Linux) ;; *) reject "unknown_host_$log" ;; esac
	[ -n "$(field "$identity" arch)" ] || reject "arch_missing_$log"
	[ -n "$(field "$identity" clang_target)" ] || reject "target_missing_$log"
	[ -n "$(field "$identity" go)" ] || reject "go_host_missing_$log"
	[ "$(field "$identity" revision)" = "$expected_revision" ] || reject "wrong_revision_$log"
	[ "$(field "$identity" tree)" = clean ] || reject "tree_not_clean_$log"
	[ "$(field "$identity" sources)" = "$expected_sources" ] || reject "source_hash_mismatch_$log"
	[ "$(field "$identity" fixtures)" = "$expected_fixtures" ] || reject "fixture_hash_mismatch_$log"
	[ "$(field "$identity" inputs)" = "$expected_inputs" ] || reject "input_set_mismatch_$log"
	[ "$(field "$result" host)" = "$host" ] || reject "result_host_mismatch_$log"
	[ "$(field "$result" revision)" = "$expected_revision" ] || reject "result_revision_mismatch_$log"
	[ "$(field "$result" tree)" = clean ] || reject "result_tree_not_clean_$log"
	[ "$(field "$result" cases)" = 10 ] || reject "result_case_count_$log"
	[ "$(field "$result" controls)" = 3 ] || reject "result_control_count_$log"
	[ "$(field "$result" status)" = pass ] || reject "result_not_pass_$log"

	[ "$(grep -c '^phase27 case ' "$log" || true)" = 10 ] || reject "case_row_count_$log"
	for program in fizzbuzz sum_to_n; do
		for input in 0 1 15 1000 1001; do
			exit_code=0
			[ "$input" = 1001 ] && exit_code=65
			want="phase27 case program=$program input=$input expected_exit=$exit_code actual_exit=$exit_code stdout=exact stderr=exact status=pass"
			[ "$(grep -Fxc -- "$want" "$log" || true)" = 1 ] || reject "case_missing_or_altered_${program}_${input}_$log"
		done
	done
	[ "$(grep -c '^phase27 control ' "$log" || true)" = 3 ] || reject "control_row_count_$log"
	for control in dynamic-zero output-limit output-limit-public; do
		[ "$(grep -Ec "^phase27 control name=$control package=[^ ]+ tests=[^ ]+( [^ ]+)* status=pass\$" "$log" || true)" = 1 ] || reject "control_missing_or_altered_${control}_$log"
	done
	printf '%s' "$host"
}

host_a=$(check_log "$1")
host_b=$(check_log "$2")
[ "$host_a" != "$host_b" ] || reject "duplicate_host_$host_a"
printf 'phase27 receipt gate: status=pass hosts=%s,%s revision=%s sources=%s fixtures=%s\n' \
	"$host_a" "$host_b" "$expected_revision" "$expected_sources" "$expected_fixtures"
