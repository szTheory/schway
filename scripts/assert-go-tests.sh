#!/bin/sh
set -eu

selection() {
	package=$1
	shift
	[ "$#" -gt 0 ] || {
		echo "assert-go-tests: at least one exact test or fuzz name is required" >&2
		return 2
	}

	listed=$(go test "$package" -list .)
	pattern=
	for requested do
		case "$requested" in
			*[!A-Za-z0-9_]*|'')
				echo "assert-go-tests: invalid exact Go target: $requested" >&2
				return 2
				;;
		esac
		found=false
		while IFS= read -r discovered; do
			if [ "$discovered" = "$requested" ]; then
				found=true
				break
			fi
		done <<EOF
$listed
EOF
		if [ "$found" != true ]; then
			echo "assert-go-tests: target not discovered: $requested" >&2
			return 1
		fi
		if [ -z "$pattern" ]; then
			pattern=$requested
		else
			pattern="$pattern|$requested"
		fi
	done

	go test "$package" -run "^($pattern)$" -count=1
}

if [ "${1:-}" = "--self-test" ]; then
	shift
	[ "$#" -ge 2 ] || {
		echo "usage: $0 --self-test PACKAGE KNOWN [POSITIVE...]" >&2
		exit 2
	}
	package=$1
	known=$2
	shift 2
	if selection "$package" TestCodenameLangSelectionGuardMustNotExist >/dev/null 2>&1; then
		echo "assert-go-tests: nonexistent sentinel was accepted" >&2
		exit 1
	fi
	selection "$package" "$known"
	if [ "$#" -gt 0 ]; then
		selection "$package" "$@"
	fi
	exit 0
fi

[ "$#" -ge 2 ] || {
	echo "usage: $0 PACKAGE TEST_OR_FUZZ [TEST_OR_FUZZ...]" >&2
	exit 2
}
selection "$@"
