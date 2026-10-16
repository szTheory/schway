#!/bin/sh
# Phase 32 native host driver (EVD-15, D-08b, D-09, D-09a). Builds the public
# schway command from this checkout, builds the receipt programs once each at
# the default build, runs the shared case table on THIS host only and prints a
# provenance-bearing log. A local run is a one-host observation, not a receipt:
# scripts/check-receipts.sh testdata/receipts/M006.manifest is the gate for a
# macOS + Linux pair and the manifest is the single source for cases and programs.
#
# The revision in every row is the codedigest (scripts/codedigest.sh), not a
# commit: headsha is recorded for traceability only.
#
# D-09a non-claim: case rows are taken at the default build only. The -O0, -O3
# and -O3+LTO matrix is bound by the hosted go test lanes
# (TestPhase29AdlerBothEngines, TestPhase29OOBApplicationLanes) in the same
# run, not by rows in this log.
#
# Case table grammar, one entry per line: `# case PROGRAM KEY EXIT STDOUT STDERR`.
# STDOUT is a decimal (compared as the decimal plus a newline), `-` for empty,
# or `@path` to cmp against a committed file. STDERR is `-` for empty, or the
# whole remainder of the line (compared as that text plus a newline).
set -eu

receipt_label=phase32
m007_mode=false
fail() {
	echo "$receipt_label verify: $1" >&2
	printf '%s host result status=fail reason=%s\n' "$receipt_label" "$2"
	exit 1
}

if [ "$#" -gt 0 ]; then
	if [ "$#" -eq 2 ] && [ "$1" = --milestone ] && [ "$2" = M007 ]; then
		receipt_label=phase38
		m007_mode=true
	elif [ "$#" -eq 4 ] && [ "$1" = --milestone ] && [ "$2" = M007 ] && [ "$3" = --probe ] && [ "$4" = crc32/check ]; then
		:
	else
		printf 'phase38 probe result status=fail reason=invalid_selector\n'
		exit 2
	fi
fi

if [ "$#" -eq 4 ]; then
	for tool in go clang mktemp rm dirname cmp printf; do
		command -v "$tool" >/dev/null 2>&1 || { printf 'phase38 probe result status=fail reason=unavailable_tool_%s\n' "$tool"; exit 1; }
	done
	verify_root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
	cd "$verify_root"
	probe_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase38-probe.XXXXXX") || { printf 'phase38 probe result status=fail reason=tempdir_failed\n'; exit 1; }
	trap 'rm -rf "$probe_tmp"' EXIT HUP INT TERM
	printf '123456789' >"$probe_tmp/check.bin"
	go build -o "$probe_tmp/schway" ./cmd/schway || { printf 'phase38 probe result status=fail reason=build_failed\n'; exit 1; }
	for engine in interpreter native; do
		set +e
		if [ "$engine" = native ]; then
			"$probe_tmp/schway" run --engine=native examples/crc32/crc32.schway --manifest examples/crc32/file_bytes.bindings.json -- "$probe_tmp/check.bin" >"$probe_tmp/$engine.out" 2>"$probe_tmp/$engine.err"
		else
			"$probe_tmp/schway" run --engine=interpreter examples/crc32/crc32.schway -- "$probe_tmp/check.bin" >"$probe_tmp/$engine.out" 2>"$probe_tmp/$engine.err"
		fi
		probe_status=$?
		set -e
		printf '3421780262\n' >"$probe_tmp/want.out"
		[ "$probe_status" -eq 0 ] && cmp -s "$probe_tmp/want.out" "$probe_tmp/$engine.out" && [ ! -s "$probe_tmp/$engine.err" ] || { printf 'phase38 probe result status=fail reason=crc_tuple_%s\n' "$engine"; exit 1; }
	done
	printf 'phase38 probe result status=pass\n'
	exit 0
fi

for tool in go clang git uname mktemp rm cat sed cmp mkdir dirname printf head tr; do
	if ! command -v "$tool" >/dev/null 2>&1; then
		fail "required tool is missing: $tool" "unavailable_tool_$tool"
	fi
done

verify_root=$(CDPATH= cd "$(dirname "$0")/.." && pwd -P)
cd "$verify_root"
receipt_manifest="$verify_root/testdata/receipts/M006.manifest"
[ "$m007_mode" = false ] || receipt_manifest="$verify_root/testdata/receipts/M007.manifest"
receipt_programs=$(sed -n 's/^programs=//p' "$receipt_manifest" | sed -n '1p')
[ -n "$receipt_programs" ] || fail "programs line is missing" "programs_line_missing"
verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase32.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM

digest_paths=$(sh scripts/codedigest.sh --paths | tr '\n' ' ')
tree_state() {
	# shellcheck disable=SC2086
	if [ -z "$(git status --porcelain -- $digest_paths)" ]; then printf 'clean'; else printf 'modified'; fi
}

host_os=$(uname -s)
host_arch=$(uname -m)
go_host=$(go env GOOS)/$(go env GOARCH)
go_version=$(go version)
clang_version=$(clang --version | sed -n '1p')
clang_target=$(clang -dumpmachine)
headsha=$(git rev-parse HEAD)
codedigest=$(sh scripts/codedigest.sh) || fail "codedigest failed" "codedigest_failed"
case "$host_os" in
	Darwin) host_name=macOS ;;
	Linux) host_name=Linux ;;
	*) fail "unsupported native evidence host: $host_os" "unsupported_host" ;;
esac
if [ -z "$host_arch" ] || [ -z "$go_host" ] || [ -z "$clang_version" ] || [ -z "$clang_target" ] || [ -z "$headsha" ] || [ -z "$codedigest" ]; then
	fail "host/compiler identity is incomplete" "incomplete_identity"
fi
tree_start=$(tree_state)

printf '%s host identity host=%s arch=%s go=%s clang_target=%s headsha=%s codedigest=%s tree=%s\n' \
	"$receipt_label" "$host_name" "$host_arch" "$go_host" "$clang_target" "$headsha" "$codedigest" "$tree_start"
printf '%s toolchain go="%s" clang="%s"\n' "$receipt_label" "$go_version" "$clang_version"

case_table=$(grep '^# case ' "$receipt_manifest" || true)
[ -n "$case_table" ] || fail "case table is empty" "case_table_empty"
table_count=$(printf '%s\n' "$case_table" | grep -c '^# case ')

schway="$verify_tmp/schway"
go build -o "$schway" ./cmd/schway || fail "unable to build the schway command" "build_failed"

build_program() { # build_program PROGRAM SOURCE [MANIFEST]
	program=$1; source=$2; manifest=${3:-}
	if [ -n "$manifest" ]; then
		"$schway" build "$source" --manifest "$manifest" --output "$verify_tmp/$program" >"$verify_tmp/build-$program.out" 2>"$verify_tmp/build-$program.err" ||
			{ cat "$verify_tmp/build-$program.err" >&2; fail "unable to build $source" "build_failed_$program"; }
	else
		"$schway" build "$source" --output "$verify_tmp/$program" >"$verify_tmp/build-$program.out" 2>"$verify_tmp/build-$program.err" ||
			{ cat "$verify_tmp/build-$program.err" >&2; fail "unable to build $source" "build_failed_$program"; }
	fi
}
build_program checksum examples/checksum/checksum.schway examples/checksum/file_bytes.bindings.json
build_program adler_low examples/checksum/adler_low.schway examples/checksum/file_bytes.bindings.json
build_program adler32_cli examples/checksum/adler32_cli.schway examples/phase29/fixture_bytes.bindings.json
build_program adler32_fixture examples/phase29/adler32_fixture.schway examples/phase29/fixture_bytes.bindings.json
build_program byte_index_oob examples/phase29/byte_index_oob.schway examples/phase29/fixture_bytes.bindings.json
build_program sum_to_n examples/sum_to_n.schway
build_program fizzbuzz examples/fizzbuzz.schway
if [ "$m007_mode" = true ]; then
	build_program crc32 examples/crc32/crc32.schway examples/crc32/file_bytes.bindings.json
fi

mkdir -p "$verify_tmp/in"
: >"$verify_tmp/in/empty"
printf '\000' >"$verify_tmp/in/zero"
printf '123456789' >"$verify_tmp/in/check"
printf 'Wikipedia' >"$verify_tmp/in/wikipedia"
head -c 65536 /dev/zero | LC_ALL=C tr '\000' '\377' >"$verify_tmp/in/ff65536"
head -c 65537 /dev/zero | LC_ALL=C tr '\000' '\377' >"$verify_tmp/in/ff65537"
mkdir "$verify_tmp/in/directory"

: >"$verify_tmp/empty"
case_count=0
# run_case PROGRAM KEY EXIT STDOUT STDERR
run_case() {
	program=$1; key=$2; want_exit=$3; want_out=$4; want_err=$5
	case "$program" in
		checksum|adler_low|crc32) arg="$verify_tmp/in/$key" ;;
		*) arg=$key ;;
	esac
	case "$want_out" in
		-) out_file="$verify_tmp/empty" ;;
		@*) out_file=${want_out#@} ;;
		*) out_file="$verify_tmp/want-$program-$key.out"; printf '%s\n' "$want_out" >"$out_file" ;;
	esac
	if [ "$want_err" = - ]; then
		err_file="$verify_tmp/empty"
	else
		err_file="$verify_tmp/want-$program-$key.err"; printf '%s\n' "$want_err" >"$err_file"
	fi
	got_out="$verify_tmp/got-$program-$key.out"
	got_err="$verify_tmp/got-$program-$key.err"
	set +e
	"$schway" app run "$verify_tmp/$program" -- "$arg" >"$got_out" 2>"$got_err"
	got_exit=$?
	set -e
	status=pass
	stdout=exact; stderr=exact
	cmp -s "$out_file" "$got_out" || { stdout=mismatch; status=fail; }
	cmp -s "$err_file" "$got_err" || { stderr=mismatch; status=fail; }
	[ "$got_exit" = "$want_exit" ] || status=fail
	printf '%s case program=%s input=%s expected_exit=%s actual_exit=%s stdout=%s stderr=%s status=%s\n' \
		"$receipt_label" "$program" "$key" "$want_exit" "$got_exit" "$stdout" "$stderr" "$status"
	[ "$status" = pass ] || fail "case failed: program=$program input=$key" "case_failed_${program}_$key"
	case_count=$((case_count + 1))
}

while read -r _hash _case program key want_exit want_out want_err; do
	run_case "$program" "$key" "$want_exit" "$want_out" "$want_err"
done <<EOF
$case_table
EOF

control_count=0
# run_control NAME PACKAGE REGEX TEST... (a skipped or empty control fails)
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
	printf '%s control name=%s package=%s tests=%s status=pass\n' "$receipt_label" "$name" "$pkg" "$*"
	control_count=$((control_count + 1))
}
run_control guide ./cmd/schway '^TestPhase32GuideCleanCheckout$' TestPhase32GuideCleanCheckout
run_control route-audit ./cmd/schway '^TestPhase32PublicRouteAudit$' TestPhase32PublicRouteAudit
run_control codedigest ./internal/compiler/session '^TestPhase32CodeDigest$' TestPhase32CodeDigest
run_control receipt-rules ./internal/compiler/session '^TestPhase32ReceiptLivingClaims$' TestPhase32ReceiptLivingClaims

tree_end=$(tree_state)
[ "$tree_end" = "$tree_start" ] || fail "tree moved during the run" "tree_moved"
[ "$case_count" = "$table_count" ] || fail "incomplete case count: $case_count of $table_count" "incomplete_cases"
printf '%s host result host=%s codedigest=%s tree=%s cases=%s controls=%s programs=%s status=pass\n' \
	"$receipt_label" "$host_name" "$codedigest" "$tree_end" "$case_count" "$control_count" "$receipt_programs"
