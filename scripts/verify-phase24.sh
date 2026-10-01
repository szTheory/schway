#!/bin/sh
set -eu

require_tool() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "phase24 verify: required tool is missing: $1" >&2
		exit 1
	fi
}

for tool in go clang git uname date mktemp rm cat grep sed; do
	require_tool "$tool"
done

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase24.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM

host_os=$(uname -s)
host_arch=$(uname -m)
go_os=$(go env GOOS)
go_arch=$(go env GOARCH)
go_version=$(go version)
clang_version=$(clang --version | sed -n '1p')
clang_target=$(clang -dumpmachine)
revision=$(git rev-parse HEAD)
if [ -z "$host_os" ] || [ -z "$host_arch" ] || [ -z "$go_os" ] || [ -z "$go_arch" ]; then
	echo "phase24 verify: host identity is incomplete" >&2
	exit 1
fi
if [ -z "$clang_version" ] || [ -z "$clang_target" ]; then
	echo "phase24 verify: Clang version or target identity is unavailable" >&2
	exit 1
fi
if [ -z "$(git status --porcelain)" ]; then
	source_tree=clean
else
	source_tree=modified
fi

printf 'phase24 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s status=incomplete\n' \
	"$host_os" "$host_arch" "$go_os" "$go_arch" "$clang_target" "$revision" "$source_tree"
printf 'compiler: %s\n' "$go_version"
printf 'clang: %s\n' "$clang_version"

run_phase24_tests() {
	name=$1
	pattern=$2
	shift 2
	log="$verify_tmp/$name.log"
	printf '\nphase24 verify: %s\n' "$name"
	if ! go test -v -count=1 -run "$pattern" "$@" >"$log" 2>&1; then
		cat "$log" >&2
		echo "phase24 verify: focused test group failed: $name" >&2
		exit 1
	fi
	cat "$log"
	if grep -Fq -- '--- SKIP:' "$log"; then
		echo "phase24 verify: native evidence was skipped in group: $name" >&2
		exit 1
	fi
	if grep -Fq 'testing: warning: no tests to run' "$log"; then
		echo "phase24 verify: no matching tests ran in group: $name" >&2
		exit 1
	fi
	if ! grep -Eq '^--- PASS: TestPhase24' "$log"; then
		echo "phase24 verify: group produced no passing Phase 24 test: $name" >&2
		exit 1
	fi
}

focused_start=$(date +%s)
run_phase24_tests source-refusal '^TestPhase24(SourceRefusal|ErrorSource|RepeatedHelperSource|SourceTransfer)' \
	./internal/compiler/check
run_phase24_tests independent-peer '^TestPhase24(TransferPeer|ActivationPeer|CleanupPeer)' \
	./internal/compiler/session
run_phase24_tests model-only '^TestPhase24(RepeatedHelperTypedError|ModelReplay)' \
	./internal/compiler/interp ./internal/compiler/session
run_phase24_tests emitter-admission '^TestPhase24Emitter' \
	./internal/compiler/cgen
run_phase24_tests native-public-app '^TestPhase24(PositiveTransferNativeApplication|NativeErrorApplication)' \
	./internal/compiler/native
run_phase24_tests native-observer '^TestPhase24Observer' \
	./internal/compiler/native
run_phase24_tests public-contract '^TestPhase24ReadmeAndVerifierScriptContract$' \
	./internal/compiler/session
focused_end=$(date +%s)
focused_elapsed=$((focused_end - focused_start))

printf '\nphase24 host receipt: host=%s/%s go=%s/%s target=%s revision=%s tree=%s elapsed_seconds=%s status=pass\n' \
	"$host_os" "$host_arch" "$go_os" "$go_arch" "$clang_target" "$revision" "$source_tree" "$focused_elapsed"
