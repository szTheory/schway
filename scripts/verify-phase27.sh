#!/bin/sh
# Phase 27 native host driver. Builds the public schway command from this
# checkout, builds both committed example applications, runs the declared
# input cases on THIS host only, and prints a provenance-bearing receipt.
# A local run is a one-host receipt; scripts/check-phase27-receipts.sh is the
# only gate that can accept a macOS + Linux pair. No row says pass until every
# case and control has finished.
set -eu

fail() {
	echo "phase27 verify: $1" >&2
	printf 'phase27 host result status=fail reason=%s\n' "$2"
	exit 1
}

for tool in go clang git uname mktemp rm cat sed cmp mkdir dirname printf; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		fail "required tool is missing: $tool" "unavailable_tool_$tool"
	fi
done

verify_root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
cd "$verify_root"
verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase27.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM

host_os=$(uname -s)
host_arch=$(uname -m)
go_host=$(go env GOOS)/$(go env GOARCH)
go_version=$(go version)
clang_version=$(clang --version | sed -n '1p')
clang_target=$(clang -dumpmachine)
revision=$(git rev-parse HEAD)
case "$host_os" in
	Darwin) host_name=macOS ;;
	Linux) host_name=Linux ;;
	*) fail "unsupported native evidence host: $host_os" "unsupported_host" ;;
esac
if [ -z "$host_arch" ] || [ -z "$go_host" ] || [ -z "$clang_version" ] || [ -z "$clang_target" ] || [ -z "$revision" ]; then
	fail "host/compiler identity is incomplete" "incomplete_identity"
fi
if [ -z "$(git status --porcelain)" ]; then source_tree=clean; else source_tree=modified; fi

sources="examples/fizzbuzz.schway examples/sum_to_n.schway"
fixtures="examples/phase27/fizzbuzz-15.out examples/phase27/fizzbuzz-1000.out"
case_inputs="0,1,15,1000,1001"

hash_list() {
	out=
	for path in $1; do
		[ -f "$path" ] || fail "declared input is missing: $path" "missing_input"
		out="$out${out:+,}$path=$(git hash-object "$path")"
	done
	printf '%s' "$out"
}
source_hashes=$(hash_list "$sources")
fixture_hashes=$(hash_list "$fixtures")

printf 'phase27 host identity host=%s arch=%s go=%s clang_target=%s revision=%s tree=%s sources=%s fixtures=%s inputs=fizzbuzz:%s;sum_to_n:%s\n' \
	"$host_name" "$host_arch" "$go_host" "$clang_target" "$revision" "$source_tree" \
	"$source_hashes" "$fixture_hashes" "$case_inputs" "$case_inputs"
printf 'phase27 toolchain go="%s" clang="%s"\n' "$go_version" "$clang_version"

schway="$verify_tmp/schway"
go build -o "$schway" ./cmd/schway || fail "unable to build the schway command" "build_failed"
for program in fizzbuzz sum_to_n; do
	"$schway" build "examples/$program.schway" --output "$verify_tmp/$program" >"$verify_tmp/build-$program.out" 2>"$verify_tmp/build-$program.err" ||
		{ cat "$verify_tmp/build-$program.err" >&2; fail "unable to build examples/$program.schway" "build_failed_$program"; }
done

case_count=0
# run_case PROGRAM INPUT EXPECTED_STDOUT_FILE EXPECTED_EXIT EXPECTED_STDERR_FILE
run_case() {
	program=$1; input=$2; want_out=$3; want_exit=$4; want_err=$5
	got_out="$verify_tmp/got-$program-$input.out"
	got_err="$verify_tmp/got-$program-$input.err"
	set +e
	"$schway" app run "$verify_tmp/$program" -- "$input" >"$got_out" 2>"$got_err"
	got_exit=$?
	set -e
	status=pass
	stdout=exact; stderr=exact
	cmp -s "$want_out" "$got_out" || { stdout=mismatch; status=fail; }
	cmp -s "$want_err" "$got_err" || { stderr=mismatch; status=fail; }
	[ "$got_exit" = "$want_exit" ] || status=fail
	printf 'phase27 case program=%s input=%s expected_exit=%s actual_exit=%s stdout=%s stderr=%s status=%s\n' \
		"$program" "$input" "$want_exit" "$got_exit" "$stdout" "$stderr" "$status"
	[ "$status" = pass ] || fail "case failed: program=$program input=$input" "case_failed_${program}_$input"
	case_count=$((case_count + 1))
}

: >"$verify_tmp/empty"
printf '1\n' >"$verify_tmp/one.out"
printf '0\n' >"$verify_tmp/sum-0.out"
printf '1\n' >"$verify_tmp/sum-1.out"
printf '120\n' >"$verify_tmp/sum-15.out"
printf '500500\n' >"$verify_tmp/sum-1000.out"
printf 'schway app: input exceeds 1000\n' >"$verify_tmp/over.err"

run_case fizzbuzz 0 "$verify_tmp/empty" 0 "$verify_tmp/empty"
run_case fizzbuzz 1 "$verify_tmp/one.out" 0 "$verify_tmp/empty"
run_case fizzbuzz 15 examples/phase27/fizzbuzz-15.out 0 "$verify_tmp/empty"
run_case fizzbuzz 1000 examples/phase27/fizzbuzz-1000.out 0 "$verify_tmp/empty"
run_case fizzbuzz 1001 "$verify_tmp/empty" 65 "$verify_tmp/over.err"
for n in 0 1 15 1000; do
	run_case sum_to_n "$n" "$verify_tmp/sum-$n.out" 0 "$verify_tmp/empty"
done
run_case sum_to_n 1001 "$verify_tmp/empty" 65 "$verify_tmp/over.err"

control_count=0
# run_control NAME PACKAGE REGEX TEST...
run_control() {
	name=$1; pkg=$2; regex=$3; shift 3
	log="$verify_tmp/control-$name.log"
	go test -v -count=1 -run "$regex" "$pkg" >"$log" 2>&1 ||
		{ cat "$log" >&2; fail "native control failed: $name" "control_failed_$name"; }
	if grep -Fq -- '--- SKIP:' "$log" || grep -Fq 'no tests to run' "$log"; then
		fail "native control skipped: $name" "control_skipped_$name"
	fi
	for test_name in "$@"; do
		grep -Fq -- "--- PASS: $test_name " "$log" || fail "native control did not pass: $test_name" "control_missing_$test_name"
	done
	printf 'phase27 control name=%s package=%s tests=%s status=pass\n' "$name" "$pkg" "$*"
	control_count=$((control_count + 1))
}
run_control dynamic-zero ./internal/compiler/session '^TestPhase27DynamicZero$' TestPhase27DynamicZero
run_control output-limit ./internal/compiler/cgen '^TestPhase27Output(ExactLimit|CrossingWrite)$' TestPhase27OutputExactLimit TestPhase27OutputCrossingWrite
run_control output-limit-public ./cmd/schway '^TestPhase27OutputPublicStatus$' TestPhase27OutputPublicStatus

[ "$case_count" = 10 ] || fail "incomplete case count: $case_count" "incomplete_cases"
printf 'phase27 host result host=%s revision=%s tree=%s cases=%s controls=%s status=pass\n' \
	"$host_name" "$revision" "$source_tree" "$case_count" "$control_count"
