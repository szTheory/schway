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
#
# Completion witness (plan 14-11, EVD-02): after each pair's `go test`
# invocation returns, this script appends one additional JSON-line record
# marking that (package, pattern) pair finished, and after the whole loop
# appends one final record marking the batch finished with the pair count.
# Both sentinels carry an `Action` value ("record_pair_complete" /
# "record_batch_complete") that is disjoint from go test's own -json
# vocabulary (run/pause/cont/bench/pass/fail/skip/output), so the consumer's
# parser can never confuse a completion witness with a test result. This is
# what lets the consumer refuse a run record that did not finish producing
# what it claims to report, instead of silently degrading to a lower grade
# ceiling.
#
# Deliberately SEQUENTIAL, not parallel: a bounded-parallelism version was
# tried and measured to THRASH rather than speed up when several pairs share
# a `./internal/compiler/...`-shaped (dozens-of-packages) operand, each of
# which is already internally parallel across GOMAXPROCS -- running several
# such invocations concurrently oversubscribes the machine and can stall
# individual jobs at ~0% CPU for tens of seconds. A corpus-wide batch
# (100-200+ pairs) is therefore genuinely slow (measured: several minutes,
# not the "milliseconds" this phase's own threat model first assumed for
# run-record reads) -- an accepted, disclosed cost lane, not a bug to paper
# over with fragile backgrounding.
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

# json_escape backslash- and quote-escapes its one argument for embedding in
# a JSON string value. This is local string transformation for the
# sentinels THIS script writes, never a shell string that gets interpolated
# into a command and executed -- $pkg/$pattern stay discrete argv elements
# in every `go test` invocation below.
json_escape() {
  local s="$1"
  s=${s//\\/\\\\}
  s=${s//\"/\\\"}
  printf '%s' "$s"
}

pairs=0
while [ "$#" -ge 2 ]; do
  pkg="$1"
  pattern="$2"
  shift 2
  pairs=$((pairs + 1))
  go test "$pkg" -run "$pattern" -json -count=1 >> "$out" 2>>"${out}.stderr" || true
  pkg_json=$(json_escape "$pkg")
  pattern_json=$(json_escape "$pattern")
  printf '{"Action":"record_pair_complete","Package":"%s","Pattern":"%s"}\n' "$pkg_json" "$pattern_json" >> "$out"
done

printf '{"Action":"record_batch_complete","Pairs":%d}\n' "$pairs" >> "$out"

exit 0
