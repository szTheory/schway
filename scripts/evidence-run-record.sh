#!/usr/bin/env bash
# evidence-run-record.sh is a PRODUCER, not a blessing button. Given an
# output path and one or more <package> <pattern> pairs, it runs
# `go test <package> -run <pattern> -json -count=1` for each pair and
# appends the raw JSON-lines output to the output file. It records whatever
# the suite actually did -- pass, fail, or skip -- and asserts nothing
# itself. internal/compiler/session/evidence_grade_test.go is the CONSUMER
# that turns this record into a derived grade ceiling; a missing or empty
# record there caps every row's ceiling below EXERCISED (fail-closed), never
# a silent pass here.
#
# Usage: evidence-run-record.sh <output-path> <package> <pattern> [<package> <pattern> ...]
#
# A failing test inside one pair does not abort the batch: this script's job
# is to RECORD what happened, including a genuine failure, not to require a
# clean run before it will write anything.
set -uo pipefail

if [ "$#" -lt 3 ]; then
  echo "usage: evidence-run-record.sh <output-path> <package> <pattern> [<package> <pattern> ...]" >&2
  exit 2
fi

out="$1"
shift

if [ $(( $# % 2 )) -ne 0 ]; then
  echo "evidence-run-record.sh: arguments after <output-path> must be <package> <pattern> pairs" >&2
  exit 2
fi

: > "$out"

while [ "$#" -ge 2 ]; do
  pkg="$1"
  pattern="$2"
  shift 2
  go test "$pkg" -run "$pattern" -json -count=1 >> "$out" 2>>"${out}.stderr" || true
done

exit 0
