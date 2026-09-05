#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase4.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestOwnedBackendMutationIsMismatch TestVerifyPhase2ControlsAndWork TestVerifyPhase3ControlsAndWork TestVerifyPhase4ControlsAndWork
go test ./...
go test -race ./...
go vet ./...
go build -o "$verify_tmp/lang" ./cmd/lang
"$verify_tmp/lang" --json verify testdata/phase1 >"$verify_tmp/phase1.json"
"$verify_tmp/lang" --json verify testdata/phase2 >"$verify_tmp/phase2.json"
"$verify_tmp/lang" --json verify testdata/phase3 >"$verify_tmp/phase3.json"
"$verify_tmp/lang" --json verify testdata/phase4 >"$verify_tmp/phase4.json"

# Non-regression for Phase 1, Phase 2, and Phase 3 is proven by running
# THEIR OWN corpora with THIS phase's freshly built binary, never by
# invoking an older gate script (D-04-21/T-04-42). The prior phase's own
# gate script stays byte-frozen and is neither forked, extended, nor
# invoked here -- this script is a peer of it, not an extension.
grep -q 'control:backend.runtime_causality' "$verify_tmp/phase2.json" || { echo "phase4 verify: phase2 backend runtime causality control missing" >&2; exit 1; }
for control in \
	control:ownership.exclusive_conflict \
	control:ownership.exclusive_move \
	control:core.loan_endpoint_mismatch \
	control:cfg.path_oracle_disagreement \
	control:origin.understated_summary \
	control:origin.impossible_summary \
	control:origin.stale_summary \
	control:origin.omitted_summary \
	control:origin.mixed_access_chain \
	control:origin.multi_arm_omitted \
	control:origin.multi_arm_access_conflict
do
	grep -q "$control" "$verify_tmp/phase3.json" || { echo "phase4 verify: required Phase 3 control missing: $control" >&2; exit 1; }
done

# Phase 4's own required-control set (D-04-22/T-04-41). Every identifier
# listed here must also appear, verbatim, in
# internal/compiler/session/session.go's Phase4RequiredControls() --
# TestPhase4RequiredControlsMatchScript asserts the two sets are equal, so a
# control added to one and forgotten in the other fails that test.
for control in \
	control:kind.exhaustive_dispatch \
	control:foreign.call_target_not_foreign \
	control:foreign.unwind_policy_undeclared \
	control:resource.release_omitted \
	control:resource.release_order_transposed \
	control:foreign.layout_mismatch \
	control:foreign.no_unproven_attributes \
	control:defect.no_release_on_defect \
	control:foreign.unwind_forbidden \
	control:foreign.nonlocal_exit_undetected \
	control:terminator.walk_incomplete \
	control:origin.foreign_origin_omitted
do
	grep -q "$control" "$verify_tmp/phase4.json" || { echo "phase4 verify: required Phase 4 control missing: $control" >&2; exit 1; }
done

# The expected escapes must appear under expected escapes and must never be
# claimed as a solved, detected control (D-04-31/T-04-07/T-04-33): declared,
# asserted visible, and never presented as covered.
for escape in \
	escape:coordinated-foreign-boundary-lie \
	escape:nonlocal-exit-below-the-pad \
	escape:foreign-process-exit
do
	grep -q "$escape" "$verify_tmp/phase4.json" || { echo "phase4 verify: expected escape missing: $escape" >&2; exit 1; }
done

# Duplicated from the prior phase's own warm-observation helper rather than
# factored into a shared scripts/observe.sh -- this is the same duplication
# rationale the prior phase's own text already recorded, applied again here:
# the prior phase's gate script is byte-frozen and must not be touched to
# extract a shared helper, and a fourth near-identical copy is small enough
# that duplication stays honest and auditable rather than silently
# drifting -- recorded here, and in this plan's SUMMARY, so it cannot drift
# unnoticed.
observe() {
	name=$1
	shift
	samples="$verify_tmp/$name.samples"
	: >"$samples"
	last=
	index=0
	while [ "$index" -lt 20 ]; do
		last=$(LANG_OBSERVE_TIMING=1 "$verify_tmp/lang" --json "$@")
		elapsed=$(printf '%s\n' "$last" | sed -n 's/.*"elapsed_ns":\([0-9][0-9]*\).*/\1/p')
		[ -n "$elapsed" ] && [ "$elapsed" -gt 0 ] || { echo "phase4 verify: $name produced no timing" >&2; exit 1; }
		printf '%s\n' "$elapsed" >>"$samples"
		index=$((index + 1))
	done
	sort -n "$samples" >"$samples.sorted"
	minimum=$(sed -n '1p' "$samples.sorted")
	p50=$(sed -n '10p' "$samples.sorted")
	p95=$(sed -n '19p' "$samples.sorted")
	maximum=$(sed -n '20p' "$samples.sorted")
	bytes=$(printf '%s\n' "$last" | wc -c | tr -d ' ')
	work=$(printf '%s\n' "$last" | sed -n 's/.*"recomputed_work":\([0-9][0-9]*\).*/\1/p')
	[ -n "$work" ] && [ "$work" -gt 0 ] || { echo "phase4 verify: $name produced zero work" >&2; exit 1; }
	printf '%s warm_samples=20 p50_ns=%s p95_ns=%s min_ns=%s max_ns=%s output_bytes=%s work=%s peak_rss=unavailable\n' "$name" "$p50" "$p95" "$minimum" "$maximum" "$bytes" "$work"
}

observe format format --check testdata/phase4/foreign_acquire_one.lang
observe check check testdata/phase4/foreign_acquire_one.lang
observe interpreter run --engine=interpreter testdata/phase4/foreign_acquire_one.lang
observe native run --engine=native testdata/phase4/foreign_acquire_one.lang
observe full_verify verify testdata/phase4

cat "$verify_tmp/phase1.json"
cat "$verify_tmp/phase2.json"
cat "$verify_tmp/phase3.json"
cat "$verify_tmp/phase4.json"
