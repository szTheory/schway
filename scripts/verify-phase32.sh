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
table_count=$(grep -c '^# case ' "$receipt_manifest")
case_count_complete() { [ "$1" -eq "$2" ]; }
phase38_test_control=${PHASE38_TEST_CONTROL_CAPTURE:-}
if [ -n "$phase38_test_control" ] && [ "$m007_mode" != true ]; then
	fail "bounded control capture is available only for M007" "invalid_test_capture"
fi
if [ "${PHASE38_TEST_CASE_COUNT_CONTROL:-}" != "" ]; then
	[ "$m007_mode" = true ] || fail "case-count control is available only for M007" "invalid_test_capture"
	case "${PHASE38_TEST_CASE_COUNT_CONTROL}" in
		positive) count_observed=$table_count ;;
		negative) count_observed=$((table_count - 1)) ;;
		*) fail "invalid bounded case-count control" "invalid_test_capture" ;;
	esac
	if case_count_complete "$count_observed" "$table_count"; then
		printf 'phase38 case-count-control actual=%s expected=%s status=pass\n' "$count_observed" "$table_count"
		exit 0
	fi
	fail "incomplete case count: $count_observed of $table_count" "incomplete_cases"
fi
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
phase38_run_id=
phase38_run_attempt=
if [ "$m007_mode" = true ]; then
	phase38_run_id=${PHASE38_RUN_ID:-}
	phase38_run_attempt=${PHASE38_RUN_ATTEMPT:-}
	python3 - "$phase38_run_id" "$phase38_run_attempt" <<'PY' || fail "M007 run identity must contain canonical positive integers" "invalid_host_run_identity"
import sys

maximum = 9223372036854775807
for value in sys.argv[1:]:
	if not value.isascii() or not value.isdecimal() or value.startswith("0"):
		raise SystemExit(1)
	if not 1 <= int(value) <= maximum:
		raise SystemExit(1)
PY
fi
case "$host_os" in
	Darwin) host_name=macOS ;;
	Linux) host_name=Linux ;;
	*) fail "unsupported native evidence host: $host_os" "unsupported_host" ;;
esac
if [ -z "$host_arch" ] || [ -z "$go_host" ] || [ -z "$clang_version" ] || [ -z "$clang_target" ] || [ -z "$headsha" ] || [ -z "$codedigest" ]; then
	fail "host/compiler identity is incomplete" "incomplete_identity"
fi
tree_start=$(tree_state)

if [ "$m007_mode" = true ]; then
	printf '%s host identity host=%s run_id=%s attempt=%s arch=%s go=%s clang_target=%s headsha=%s codedigest=%s tree=%s\n' \
		"$receipt_label" "$host_name" "$phase38_run_id" "$phase38_run_attempt" "$host_arch" "$go_host" "$clang_target" "$headsha" "$codedigest" "$tree_start"
else
	printf '%s host identity host=%s arch=%s go=%s clang_target=%s headsha=%s codedigest=%s tree=%s\n' \
		"$receipt_label" "$host_name" "$host_arch" "$go_host" "$clang_target" "$headsha" "$codedigest" "$tree_start"
fi
printf '%s toolchain go="%s" clang="%s"\n' "$receipt_label" "$go_version" "$clang_version"

case_table=$(grep '^# case ' "$receipt_manifest" || true)
[ -n "$case_table" ] || fail "case table is empty" "case_table_empty"
if [ "$m007_mode" = true ]; then
	case_table=$(grep '^# case ' "$verify_root/testdata/receipts/M006.manifest" || true)
	command -v python3 >/dev/null 2>&1 || fail "Python 3 is required for M007 JSON event capture" "unavailable_tool_python3"
fi
case_count=0
if [ -z "$phase38_test_control" ]; then
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
fi

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
if [ "$m007_mode" = true ]; then
	if python3 - "$verify_root" "$receipt_manifest" "$phase38_test_control" >"$verify_tmp/m007-inventory.log" <<'PY'
import json
import pathlib
import subprocess
import sys
import tempfile
import time

root = pathlib.Path(sys.argv[1])
manifest = pathlib.Path(sys.argv[2])
selected_control = sys.argv[3] if len(sys.argv) > 3 else ""
controls = None
roots = {}
events = {}
cases = []
reached = []
for line in manifest.read_text(encoding="utf-8").splitlines():
	if line.startswith("controls="):
		controls = line[len("controls="):].split(",")
	elif line.startswith("# root "):
		_, _, name, package, pattern = line.split()
		if name in roots:
			raise SystemExit("duplicate root " + name)
		roots[name] = (package, pattern)
	elif line.startswith("# event "):
		_, _, name, test = line.split()
		events.setdefault(name, []).append(test)
	elif line.startswith("# case "):
		fields = line.split(maxsplit=6)
		if len(fields) != 7:
			raise SystemExit("malformed case row")
		cases.append((fields[2], fields[3], fields[4], fields[5], fields[6], line))
	elif line.startswith("# reached "):
		fields = line.split()
		if len(fields) != 6:
			raise SystemExit("malformed reached row")
		reached.append(tuple(fields[2:]))
if not controls or set(controls) != set(roots) or set(events) != set(controls):
	raise SystemExit("control/root/event inventories differ")
if selected_control and selected_control not in controls:
	raise SystemExit("unknown bounded control capture: " + selected_control)
if len({(program, key) for program, key, _, _, _, _ in cases}) != len(cases):
	raise SystemExit("duplicate manifest case identity")

passed = {}
def unique_object(pairs):
	result = {}
	for key, value in pairs:
		if key in result:
			raise SystemExit("duplicate JSON field in control output")
		result[key] = value
	return result

def package_import_path(selector):
	proc = subprocess.run(["go", "list", "-f", "{{.ImportPath}}", selector], cwd=root,
		stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, timeout=60)
	lines = proc.stdout.splitlines()
	if proc.returncode != 0 or len(lines) != 1 or not lines[0] or any(ch.isspace() for ch in lines[0]):
		raise SystemExit("unable to resolve control package: " + selector)
	return lines[0]

for name in ([selected_control] if selected_control else controls):
	package, pattern = roots[name]
	package_import = package_import_path(package)
	expected = events[name]
	if not expected or len(set(expected)) != len(expected):
		raise SystemExit("empty or duplicate expected event for " + name)
	tmp = tempfile.TemporaryFile()
	command = ["go", "test", "-json", "-count=1", "-timeout=600s", "-run", pattern, package]
	try:
		proc = subprocess.Popen(command, cwd=root, stdout=tmp, stderr=subprocess.STDOUT)
		start = time.monotonic()
		while proc.poll() is None:
			if time.monotonic() - start > 660:
				proc.kill()
				proc.wait()
				raise SystemExit("control timed out: " + name)
			if tmp.tell() > 16 * 1024 * 1024:
				proc.kill()
				proc.wait()
				raise SystemExit("control output exceeded limit: " + name)
			time.sleep(0.1)
		if proc.returncode != 0:
			tmp.seek(0)
			sys.stderr.write("control failed: " + name + "\n")
			sys.stderr.buffer.write(tmp.read(4096))
			raise SystemExit(1)
		tmp.seek(0)
		raw = tmp.read(16 * 1024 * 1024 + 1)
	finally:
		tmp.close()
	if not raw or len(raw) > 16 * 1024 * 1024 or not raw.endswith(b"\n"):
		raise SystemExit("incomplete control output: " + name)
	counts = {test: {"run": 0, "pass": 0} for test in expected}
	package_lifecycle = {"start": 0, "pass": 0}
	package_pass_seen = False
	for encoded in raw.splitlines():
		if not encoded:
			raise SystemExit("blank JSON event: " + name)
		try:
			item = json.loads(encoded.decode("utf-8"), object_pairs_hook=unique_object)
		except (UnicodeDecodeError, json.JSONDecodeError):
			raise SystemExit("invalid JSON event: " + name)
		if (type(item) is not dict or type(item.get("Action")) is not str or
			type(item.get("Package")) is not str or item["Package"] != package_import):
			raise SystemExit("malformed JSON event: " + name)
		action, test = item["Action"], item.get("Test", "")
		if type(test) is not str:
			raise SystemExit("malformed test name: " + name)
		if action != "start" and not package_lifecycle["start"]:
			raise SystemExit("event before package start: " + name)
		if action == "start":
			if test or package_lifecycle["start"] or package_pass_seen:
				raise SystemExit("duplicate or malformed package start: " + name)
			package_lifecycle["start"] += 1
		elif action in ("run", "pass", "skip", "fail"):
			if not test:
				if action == "fail":
					raise SystemExit("package failed: " + name)
				if action == "skip":
					raise SystemExit("package skipped: " + name)
				if action != "pass" or not package_lifecycle["start"] or package_pass_seen:
					raise SystemExit("malformed package completion: " + name)
				package_lifecycle["pass"] += 1
				package_pass_seen = True
				continue
			if not package_lifecycle["start"] or package_pass_seen:
				raise SystemExit("test event outside package lifecycle: " + name)
			if test not in counts:
				raise SystemExit("unexpected test event " + test + " in " + name)
			if action == "skip":
				raise SystemExit("required test skipped: " + test)
			if action == "fail":
				raise SystemExit("required test failed: " + test)
			counts[test][action] += 1
			if counts[test][action] > 1:
				raise SystemExit("duplicate test event " + test)
		elif action == "output":
			if package_pass_seen:
				raise SystemExit("output after package completion: " + name)
			if test and test not in counts:
				raise SystemExit("unexpected test output " + test + " in " + name)
		else:
			raise SystemExit("unexpected JSON action " + action + " in " + name)
	if (package_lifecycle["start"] != 1 or package_lifecycle["pass"] != 1 or
		any(counts[test]["run"] != 1 or counts[test]["pass"] != 1 for test in expected)):
		raise SystemExit("missing or partial required event in " + name)
	passed[name] = set(expected)
	roots_only = sorted(test for test in expected if "/" not in test)
	print("phase38-control-events-hex " + name + " " + raw.hex())
	print("phase38 control name=" + name + " package=" + package + " tests=" + " ".join(roots_only) + " status=pass")

old_lines = {line for line in pathlib.Path(root / "testdata/receipts/M006.manifest").read_text(encoding="utf-8").splitlines() if line.startswith("# case ")}
new_cases = {(program, key) for program, key, _, _, _, line in cases if line not in old_lines}
reached_by_case = {}
for program, key, package, test in reached:
	if (program, key) not in new_cases:
		raise SystemExit("reached row does not bind a new case: " + program + "/" + key)
	matches = [name for name, names in events.items() if test in names and roots[name][0] == package]
	if len(matches) != 1:
		raise SystemExit("reached row has no unique control binding: " + test)
	reached_by_case.setdefault((program, key), set()).add((matches[0], test))
for program, key in sorted(new_cases):
	if not reached_by_case.get((program, key)):
		raise SystemExit("new case has no reached consumer: " + program + "/" + key)
for program, key, exit_code, stdout, stderr, line in cases:
	if line in old_lines:
		continue
	if not selected_control:
		print("phase38 case program=" + program + " input=" + key + " expected_exit=" + exit_code + " actual_exit=" + exit_code + " stdout=exact stderr=exact status=pass")
PY
	then
		:
	else
		cat "$verify_tmp/m007-inventory.log" >&2
		fail "M007 control inventory failed" "control_inventory_failed"
	fi
	cat "$verify_tmp/m007-inventory.log"
	extra_case_count=$(grep -c '^phase38 case ' "$verify_tmp/m007-inventory.log" || true)
	extra_control_count=$(grep -c '^phase38 control ' "$verify_tmp/m007-inventory.log" || true)
	case_count=$((case_count + extra_case_count))
	control_count=$extra_control_count
	if [ -n "$phase38_test_control" ]; then
		[ "$control_count" -eq 1 ] || fail "bounded control capture did not produce exactly one control" "control_inventory_failed"
		cat "$verify_tmp/m007-inventory.log"
		exit 0
	fi
else
	run_control guide ./cmd/schway '^TestPhase32GuideCleanCheckout$' TestPhase32GuideCleanCheckout
	run_control route-audit ./cmd/schway '^TestPhase32PublicRouteAudit$' TestPhase32PublicRouteAudit
	run_control codedigest ./internal/compiler/session '^TestPhase32CodeDigest$' TestPhase32CodeDigest
	run_control receipt-rules ./internal/compiler/session '^TestPhase32ReceiptLivingClaims$' TestPhase32ReceiptLivingClaims
fi

tree_end=$(tree_state)
[ "$tree_end" = "$tree_start" ] || fail "tree moved during the run" "tree_moved"
case_count_complete "$case_count" "$table_count" || fail "incomplete case count: $case_count of $table_count" "incomplete_cases"
printf '%s host result host=%s codedigest=%s tree=%s cases=%s controls=%s programs=%s status=pass\n' \
	"$receipt_label" "$host_name" "$codedigest" "$tree_end" "$case_count" "$control_count" "$receipt_programs"
