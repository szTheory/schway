#!/bin/sh
set -eu

require_tool() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "phase25 verify: required tool is missing: $1" >&2
		exit 1
	fi
}

for tool in go clang git uname mktemp rm cat grep sed awk sort mkdir dirname env; do
	require_tool "$tool"
done
if [ ! -x /usr/bin/time ]; then
	echo "phase25 verify: required portable timer is missing: /usr/bin/time" >&2
	exit 1
fi

verify_root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
cd "$verify_root"
verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase25.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM

host_os=$(uname -s)
host_arch=$(uname -m)
go_os=$(go env GOOS)
go_arch=$(go env GOARCH)
go_version=$(go version)
clang_version=$(clang --version | sed -n '1p')
clang_target=$(clang -dumpmachine)
revision=$(git rev-parse HEAD)
case "$host_os" in
	Darwin) host_name=macOS ;;
	Linux) host_name=Linux ;;
	*) echo "phase25 verify: unsupported native evidence host: $host_os" >&2; exit 1 ;;
esac
if [ -z "$host_arch" ] || [ -z "$go_os" ] || [ -z "$go_arch" ] || [ -z "$clang_version" ] || [ -z "$clang_target" ]; then
	echo "phase25 verify: host/compiler identity is incomplete" >&2
	exit 1
fi
if [ -z "$(git status --porcelain)" ]; then
	source_tree=clean
else
	source_tree=modified
fi

source_inputs="examples/phase24/transfer.schway examples/phase23/file_byte.bindings.json examples/phase23/adapter.c examples/phase23/adapter.h testdata/phase25/shared_copy_accept.schway testdata/phase25/exclusive_copy_accept.schway internal/compiler/check/check.go internal/compiler/session/session.go internal/compiler/cgen/cgen.go internal/compiler/cgen/cgen_program.go internal/compiler/native/native_app.go internal/compiler/native/phase25_utility_test.go go.mod"
input_hashes=
for input in $source_inputs; do
	if [ ! -f "$input" ]; then
		echo "phase25 verify: declared build input is missing: $input" >&2
		exit 1
	fi
	input_hash=$(git hash-object "$input")
	input_hashes="$input_hashes $input=$input_hash"
done

printf 'phase25 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s status=incomplete\n' \
	"$host_name" "$host_arch" "$go_os" "$go_arch" "$clang_target" "$revision" "$source_tree"
printf 'compiler: %s\n' "$go_version"
printf 'clang: %s\n' "$clang_version"
printf 'phase25 build identity: revision=%s tree=%s target=%s source_inputs="%s" input_git_blob_hashes="%s"\n' \
	"$revision" "$source_tree" "$clang_target" "$source_inputs" "$input_hashes"

timings="$verify_tmp/timings.csv"
: >"$timings"

run_lane_sample() {
	lane=$1
	sample_kind=$2
	cache_path=$3
	log="$verify_tmp/$lane-$sample_kind.log"
	time_log="$verify_tmp/$lane-$sample_kind.time"
	case "$lane" in
		baseline) flags='-std=c17 -Wall -Wextra -Werror -pedantic -O0' ;;
		optimized) flags='-std=c17 -Wall -Wextra -Werror -pedantic -O2' ;;
		sanitizer) flags='-std=c17 -O1 -g -fno-omit-frame-pointer -fsanitize=address,undefined -fno-sanitize-recover=all -lc++' ;;
		*) echo "phase25 verify: unknown lane: $lane" >&2; exit 1 ;;
	esac
	printf '\nphase25 verify: lane=%s sample=%s flags=%s\n' "$lane" "$sample_kind" "$flags"
	if ! /usr/bin/time -p -o "$time_log" env GOCACHE="$cache_path" SCHWAY_PHASE25_LANE="$lane" \
		go test -v -count=1 -run '^TestPhase25UtilityNative$' ./internal/compiler/native >"$log" 2>&1; then
		cat "$log" >&2
		echo "phase25 verify: native utility test failed: lane=$lane sample=$sample_kind" >&2
		exit 1
	fi
	grep -F 'phase25 native receipt' "$log"
	grep -F -- '--- PASS: TestPhase25UtilityNative' "$log"
	if grep -Fq -- '--- SKIP:' "$log"; then
		echo "phase25 verify: native evidence was skipped: lane=$lane sample=$sample_kind" >&2
		exit 1
	fi
	if grep -Fq 'testing: warning: no tests to run' "$log" || ! grep -Fq -- '--- PASS: TestPhase25UtilityNative' "$log"; then
		echo "phase25 verify: no passing native utility witness: lane=$lane sample=$sample_kind" >&2
		exit 1
	fi
	for family in foreign shared exclusive; do
		if ! grep -Fq "phase25 native receipt family=$family lane=$lane host=$go_os/$go_arch input=0x41 expected=\"65\" actual=\"65\" status=pass" "$log" ||
		   ! grep -Fq "phase25 native receipt family=$family lane=$lane host=$go_os/$go_arch input=0x42 expected=\"66\" actual=\"66\" status=pass" "$log"; then
			echo "phase25 verify: family receipt missing or mismatched: family=$family lane=$lane" >&2
			exit 1
		fi
		printf 'phase25 receipt family=%s host=%s/%s lane=%s flags="%s" compiler="%s" target=%s revision=%s expected=65,66 actual=65,66 status=pass\n' \
			"$family" "$host_name" "$host_arch" "$lane" "$flags" "$clang_version" "$clang_target" "$revision"
	done
	if ! grep -Fq "phase25 native receipt family=utility-error lane=$lane host=$go_os/$go_arch input=0x43 expected=UseError.UnsupportedByte" "$log" ||
	   ! grep -Fq 'helper_calls=0 status=pass' "$log"; then
		echo "phase25 verify: typed 0x43 error-before-helpers receipt is absent: lane=$lane" >&2
		exit 1
	fi
	seconds=$(awk '$1 == "real" { print $2 }' "$time_log")
	if [ -z "$seconds" ]; then
		echo "phase25 verify: timer did not report elapsed seconds: lane=$lane sample=$sample_kind" >&2
		exit 1
	fi
	printf '%s,%s,%s\n' "$sample_kind" "$lane" "$seconds" >>"$timings"
	cache_state=empty_go_build_cache
	if [ "$sample_kind" = warm ]; then cache_state=reused_go_build_cache; fi
	printf 'phase25 feedback sample scope=TestPhase25UtilityNative class=%s cache_state=%s lane=%s seconds=%s status=pass\n' "$sample_kind" "$cache_state" "$lane" "$seconds"
}

# Each optimization lane gets its own initially empty Go build cache. The
# following sample reuses that cache, giving cold/warm timings for the same
# build/run/observe test without conflating O0, O2, and sanitizer flags.
for lane in baseline optimized sanitizer; do
	cache_path="$verify_tmp/gocache-$lane"
	mkdir "$cache_path"
	run_lane_sample "$lane" cold "$cache_path"
	run_lane_sample "$lane" warm "$cache_path"
done

print_distribution() {
	sample_kind=$1
	values=$(awk -F, -v kind="$sample_kind" '$1 == kind { print $3 }' "$timings" | sort -n)
	if [ -z "$values" ]; then
		echo "phase25 verify: no $sample_kind timing samples recorded" >&2
		exit 1
	fi
	printf '%s\n' "$values" | awk -v kind="$sample_kind" '
		{ value[NR]=$1 }
		END {
			if (NR % 2) median=value[(NR+1)/2]; else median=(value[NR/2]+value[NR/2+1])/2
			printf "phase25 feedback distribution scope=TestPhase25UtilityNative class=%s lanes=baseline,optimized,sanitizer n=%d min_s=%.3f median_s=%.3f max_s=%.3f status=pass\n", kind, NR, value[1], median, value[NR]
		}'
}
print_distribution cold
print_distribution warm

printf '\nphase25 verify: native reached pointer controls\n'
control_log="$verify_tmp/controls.log"
if ! GOCACHE="$verify_tmp/gocache-baseline" go test -v -count=1 -run '^TestPhase25(NativeFamilyWrongResult|UtilityFamilyControls|UtilityPointerManifests)$' ./internal/compiler/native >"$control_log" 2>&1; then
	cat "$control_log" >&2
	echo "phase25 verify: pointer wrong-result/conflict/escape control group failed" >&2
	exit 1
fi
grep -E 'phase25 reached wrong-result control|--- PASS: TestPhase25(NativeFamilyWrongResult|UtilityFamilyControls|UtilityPointerManifests)' "$control_log"
if grep -Fq -- '--- SKIP:' "$control_log" || grep -Fq 'testing: warning: no tests to run' "$control_log" ||
	! grep -Fq -- '--- PASS: TestPhase25NativeFamilyWrongResult' "$control_log" ||
	! grep -Fq -- '--- PASS: TestPhase25UtilityFamilyControls' "$control_log" ||
	! grep -Fq -- '--- PASS: TestPhase25UtilityPointerManifests' "$control_log"; then
	echo "phase25 verify: pointer control group skipped or incomplete" >&2
	exit 1
fi
for control in 'family=shared helper expected=65 actual=99 status=pass' 'family=exclusive helper expected=65 actual=98 status=pass' \
	'TestPhase25UtilityFamilyControls/shared_conflict' 'TestPhase25UtilityFamilyControls/exclusive_conflict' \
	'TestPhase25UtilityFamilyControls/shared_escape' 'TestPhase25UtilityFamilyControls/exclusive_escape' \
	'TestPhase25UtilityPointerManifests/shared' 'TestPhase25UtilityPointerManifests/exclusive'; do
	if ! grep -Fq -- "$control" "$control_log"; then
		echo "phase25 verify: required reached pointer control receipt is missing: $control" >&2
		exit 1
	fi
done
printf 'phase25 control receipt family=shared reached_wrong_result=99 conflict=ownership.borrow_conflict escape=core.origin_omitted pointer_manifest=const_uint64_t* status=pass\n'
printf 'phase25 control receipt family=exclusive reached_wrong_result=98 conflict=ownership.borrow_conflict escape=core.origin_omitted pointer_manifest=uint64_t* status=pass\n'

printf '\nphase25 verify: required-host matrix rows not produced by this native process remain incomplete\n'
for required_host in macOS Linux; do
	if [ "$required_host" != "$host_name" ]; then
		for family in foreign shared exclusive; do
			for lane in baseline optimized sanitizer; do
				printf 'phase25 receipt family=%s host=%s lane=%s expected=65,66 actual=unobserved status=incomplete\n' "$family" "$required_host" "$lane"
			done
		done
	fi
done
printf 'phase25 evidence matrix: families=foreign,shared,exclusive hosts=macOS,Linux lanes=baseline,optimized,sanitizer local_host=%s hosted_receipts=not_observed status=incomplete\n' "$host_name"
printf '\nphase25 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s status=pass\n' \
	"$host_name" "$host_arch" "$go_os" "$go_arch" "$clang_target" "$revision" "$source_tree"
