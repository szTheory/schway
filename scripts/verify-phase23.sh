#!/bin/sh
set -eu

require_tool() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "phase23 verify: required tool is missing: $1" >&2
		exit 1
	fi
}

for tool in go clang git uname date mktemp rm cat grep sed; do
	require_tool "$tool"
done

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase23.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM

host_os=$(uname -s)
host_arch=$(uname -m)
go_os=$(go env GOOS)
go_arch=$(go env GOARCH)
go_version=$(go version)
cschway_version=$(clang --version | sed -n '1p')
cschway_target=$(clang -dumpmachine)
revision=$(git rev-parse HEAD)
if [ -z "$host_os" ] || [ -z "$host_arch" ] || [ -z "$go_os" ] || [ -z "$go_arch" ]; then
	echo "phase23 verify: host identity is incomplete" >&2
	exit 1
fi
if [ -z "$cschway_version" ] || [ -z "$cschway_target" ]; then
	echo "phase23 verify: Clang version or target identity is unavailable" >&2
	exit 1
fi
if git diff --quiet && git diff --cached --quiet; then
	source_tree=clean
else
	source_tree=modified
fi

printf 'phase23 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s status=incomplete\n' \
	"$host_os" "$host_arch" "$go_os" "$go_arch" "$cschway_target" "$revision" "$source_tree"
printf 'compiler: %s\n' "$go_version"
printf 'cschway: %s\n' "$cschway_version"

run_phase23_tests() {
	name=$1
	pattern=$2
	shift 2
	log="$verify_tmp/$name.log"
	printf '\nphase23 verify: %s\n' "$name"
	if ! go test -v -count=1 -run "$pattern" "$@" >"$log" 2>&1; then
		cat "$log" >&2
		echo "phase23 verify: focused test group failed: $name" >&2
		exit 1
	fi
	cat "$log"
	if grep -Fq -- '--- SKIP:' "$log"; then
		echo "phase23 verify: native evidence was skipped in group: $name" >&2
		exit 1
	fi
	if grep -Fq 'testing: warning: no tests to run' "$log"; then
		echo "phase23 verify: no matching tests ran in group: $name" >&2
		exit 1
	fi
	if ! grep -Eq '^--- PASS: TestPhase23' "$log"; then
		echo "phase23 verify: group produced no passing Phase 23 test: $name" >&2
		exit 1
	fi
}

focused_start=$(date +%s)
run_phase23_tests source-admission '^TestPhase23(SourceRefusal|Discard|OperationContract)' \
	./internal/compiler/check ./internal/compiler/cgen
run_phase23_tests independent-peers '^TestPhase23' \
	./internal/compiler/corevalidate ./internal/compiler/originvalidate ./internal/compiler/pathoracle
run_phase23_tests model-only '^TestPhase23Model' \
	./internal/compiler/interp ./internal/compiler/session
run_phase23_tests native-observer '^TestPhase23(GeneratedRelease|OperationABI|Acquire|Observer|PhysicalDestructorControl)' \
	./internal/compiler/native ./internal/compiler/cgen
run_phase23_tests public-contract '^TestPhase23(Readme|Public|Contract|VerifierScript)' \
	./cmd/schway ./internal/compiler/session
focused_end=$(date +%s)
focused_elapsed=$((focused_end - focused_start))

printf '\nphase23 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s elapsed_seconds=%s status=pass\n' \
	"$host_os" "$host_arch" "$go_os" "$go_arch" "$cschway_target" "$revision" "$source_tree" "$focused_elapsed"
