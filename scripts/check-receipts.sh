#!/bin/sh
# Generic two-host receipt consistency gate. This checks consistency, not authenticity.
# Usage: check-receipts.sh MANIFEST LOG_A LOG_B
set -eu

label=
reject() {
	if [ -n "$label" ]; then prefix="$label "; else prefix=; fi
	printf '%sreceipt gate: status=fail reason=%s\n' "$prefix" "$1" >&2
	exit 1
}

[ "$#" -gt 0 ] || reject usage_two_logs_required
case $1 in /*) manifest=$1 ;; *) manifest=$PWD/$1 ;; esac
[ -r "$manifest" ] || reject manifest_missing
mf_get() { sed -n "s/^$1=//p" "$manifest" | sed -n '1p'; }
label=$(mf_get label)
[ -n "$label" ] || reject manifest_missing
want_programs=$(mf_get programs)
controls=$(mf_get controls)
case_table=$(grep '^# case ' "$manifest" || true)
[ -n "$case_table" ] || reject case_table_empty
[ -n "$want_programs" ] || reject programs_line_missing
shift
[ "$#" -eq 2 ] || reject usage_two_logs_required

typed_a=$1
typed_b=$2
case $typed_a in /*) log_a=$typed_a ;; *) log_a=$PWD/$typed_a ;; esac
case $typed_b in /*) log_b=$typed_b ;; *) log_b=$PWD/$typed_b ;; esac
root=$(git rev-parse --show-toplevel 2>/dev/null) || reject no_repo_root
[ -n "$root" ] || reject no_repo_root
cd "$root" || reject no_repo_root

if [ -n "${EXPECTED_CODEDIGEST:-}" ]; then
	digest_source=env
	expected_digest=$EXPECTED_CODEDIGEST
else
	digest_source=computed
	expected_digest=$(sh "$root/scripts/codedigest.sh") || reject expected_codedigest_unknown
fi
[ -n "$expected_digest" ] || reject expected_codedigest_unknown
table_count=$(printf '%s\n' "$case_table" | grep -c '^# case ')
control_count=$(printf '%s' "$controls" | awk -F, '{print NF}')

prefix_count() { awk -v p="$1" 'index($0,p)==1 { n++ } END { print n+0 }' "$2"; }
prefix_first() { awk -v p="$1" 'index($0,p)==1 { print; exit }' "$2"; }
field() { printf '%s\n' "$1" | tr ' ' '\n' | sed -n "s/^$2=//p" | sed -n '1p'; }
identity_digest() { field "$(prefix_first "$label host identity " "$1")" codedigest; }

check_log() {
	log=$1
	typed=$2
	[ -f "$log" ] || reject "host_missing_$typed"
	[ -s "$log" ] || reject "log_empty_$typed"
	if grep -Eq 'status=(fail|incomplete|unavailable|unobserved)|unavailable' "$log"; then reject "failed_or_unavailable_row_$typed"; fi
	identity_count=$(prefix_count "$label host identity " "$log")
	result_count=$(prefix_count "$label host result " "$log")
	[ "$identity_count" = 1 ] || reject "identity_row_count_${identity_count}_$typed"
	[ "$result_count" = 1 ] || reject "result_row_count_${result_count}_$typed"
	first=$(awk -v p="$label " 'index($0,p)==1 { print; exit }' "$log")
	case "$first" in "$label host identity "*) ;; *) reject "identity_not_first_$typed" ;; esac
	last=$(awk -v p="$label " 'index($0,p)==1 { row=$0 } END { print row }' "$log")
	case "$last" in "$label host result "*) ;; *) reject "result_not_last_$typed" ;; esac
	unknown_rows=$(awk -v p="$label " 'index($0,p)==1 { rest=substr($0,length(p)+1); if (rest !~ /^(host identity |toolchain |case |control |host result )/) print }' "$log")
	[ -z "$unknown_rows" ] || reject "unknown_${label}_row_$typed"
	toolchain_valid=$(awk -v p="$label toolchain " 'index($0,p)==1 && substr($0,length(p)+1) ~ /^go="[^"]+" clang="[^"]+"$/ { n++ } END { print n+0 }' "$log")
	[ "$(prefix_count "$label toolchain " "$log")" = 1 ] || reject "toolchain_row_count_$typed"
	[ "$toolchain_valid" = 1 ] || reject "toolchain_row_malformed_$typed"
	identity=$(prefix_first "$label host identity " "$log")
	result=$(awk -v p="$label host result " 'index($0,p)==1 { print; exit }' "$log")
	host=$(field "$identity" host)
	case "$host" in macOS|Linux) ;; *) reject "unknown_host_$typed" ;; esac
	[ -n "$(field "$identity" arch)" ] || reject "arch_missing_$typed"
	[ -n "$(field "$identity" clang_target)" ] || reject "target_missing_$typed"
	[ -n "$(field "$identity" go)" ] || reject "go_host_missing_$typed"
	[ -n "$(field "$identity" headsha)" ] || reject "headsha_missing_$typed"
	[ "$(field "$identity" codedigest)" = "$expected_digest" ] || reject "digest_mismatch_$typed"
	[ "$(field "$identity" tree)" = clean ] || reject "tree_not_clean_$typed"
	[ "$(field "$result" host)" = "$host" ] || reject "result_host_mismatch_$typed"
	[ "$(field "$result" codedigest)" = "$expected_digest" ] || reject "result_digest_mismatch_$typed"
	[ "$(field "$result" tree)" = clean ] || reject "result_tree_not_clean_$typed"
	[ "$(field "$result" status)" = pass ] || reject "result_not_pass_$typed"
	[ "$(field "$result" programs)" = "$want_programs" ] || reject "result_programs_mismatch_$typed"
	for program in $(printf '%s\n' "$want_programs" | tr ',' ' '); do
		printf '%s\n' "$case_table" | grep -Fq "# case $program " || reject "program_without_table_entry_$program"
		[ "$(prefix_count "$label case program=$program " "$log")" -ge 1 ] || reject "program_missing_${program}_$typed"
	done
	[ "$(field "$result" controls)" = "$control_count" ] || reject "result_control_count_$typed"
	[ "$(field "$result" cases)" = "$table_count" ] || reject "result_case_count_$typed"
	[ "$(prefix_count "$label case " "$log")" = "$table_count" ] || reject "case_row_count_$typed"
	while read -r _hash _case program key want_exit _stdout _stderr; do
		want="$label case program=$program input=$key expected_exit=$want_exit actual_exit=$want_exit stdout=exact stderr=exact status=pass"
		[ "$(grep -Fxc -- "$want" "$log" || true)" = 1 ] || reject "case_missing_or_altered_${program}_${key}_$typed"
	done <<EOF
$case_table
EOF
	[ "$(prefix_count "$label control " "$log")" = "$control_count" ] || reject "control_row_count_$typed"
	for control in $(printf '%s' "$controls" | tr ',' ' '); do
		valid_count=$(awk -v p="$label control name=$control package=" 'index($0,p)==1 && $0 ~ / package=[^ ]+ tests=[^ ]+( [^ ]+)* status=pass$/ { n++ } END { print n+0 }' "$log")
		[ "$valid_count" = 1 ] || reject "control_missing_or_altered_${control}_$typed"
	done
	printf '%s' "$host"
}

if [ -f "$log_a" ] && [ -f "$log_b" ]; then
	digest_a=$(identity_digest "$log_a" || true)
	digest_b=$(identity_digest "$log_b" || true)
	if [ -n "$digest_a" ] && [ -n "$digest_b" ] && [ "$digest_a" != "$digest_b" ]; then reject log_digest_differs; fi
fi
host_a=$(check_log "$log_a" "$typed_a")
host_b=$(check_log "$log_b" "$typed_b")
[ "$host_a" != "$host_b" ] || reject "duplicate_host_$host_a"
printf '%s receipt gate: status=pass hosts=%s,%s codedigest=%s programs=%s cases=%s digest_source=%s\n' \
	"$label" "$host_a" "$host_b" "$expected_digest" "$want_programs" "$table_count" "$digest_source"
