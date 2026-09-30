#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/schway-phase6.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

run_step() {
	step_name=$1
	shift
	if "$@"; then
		return 0
	else
		step_status=$?
		printf 'phase6 script step: %s exit=%s\n' "$step_name" "$step_status" >&2
		return "$step_status"
	fi
}

run_step 'session assertion tests' sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestVerifyPhase6ControlsAndWork TestPhase6RequiredControlsMatchScript TestPhase6BoundsMatchScript TestPhase6VerifierScriptContract TestPhase6ScriptInvokesNoPriorGate TestPhase6SamplingLoopMatchesGoStatistics TestPhase6ExpectedEscapesAreDeclared TestPhase6EscapesAreNeverPresentedAsControls TestPhase6EscapeGrepsMatchScript
run_step 'command assertion tests' sh scripts/assert-go-tests.sh --self-test ./cmd/schway TestPhase6CorpusDispatchRequiresMarker
run_step 'go test ./...' go test ./...
run_step 'go test -race ./...' go test -race ./...
run_step 'go vet ./...' go vet ./...
run_step 'build schway' go build -o "$verify_tmp/schway" ./cmd/schway
run_step 'build schway-repair' go build -o "$verify_tmp/schway-repair" ./cmd/schway-repair

report_verify_failure() {
	corpus=$1
	result_path=$2
	verify_status=$3
	python3 - "$corpus" "$result_path" "$verify_status" <<'PY' >&2
import json, re, sys

corpus, result_path, verify_status = sys.argv[1:]
try:
    result = json.load(open(result_path, encoding="utf-8"))
except (OSError, json.JSONDecodeError):
    print(f"verify corpus: {corpus} exit={verify_status}, result JSON unavailable")
    raise SystemExit(0)

print(f"verify corpus: {corpus} exit={verify_status} status={result.get('status', 'unknown')}")
for lane in result.get("lanes", []):
    if lane.get("status") != "pass":
        print(f"verify lane: corpus={corpus} id={lane.get('id', 'unknown')} status={lane.get('status', 'unknown')} work={lane.get('recomputed_work', 0)}")
for diagnostic in result.get("diagnostics", []):
    message = str(diagnostic.get("message", ""))
    message = re.sub(r'/(?:Users|home)/[^/\s"\\]+', "[home]", message)
    message = re.sub(r'/(?:private/)?var/folders/[^/\s"\\]+/[^/\s"\\]+', "[temp]", message)
    message = re.sub(r'/tmp/[^/\s"\\]+', "[temp]", message)
    message = re.sub(r'(?i)[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}', "[email]", message)
    message = re.sub(r'(?<![A-Fa-f0-9])(?:\+?1[\s.-]?)?(?:\([2-9][0-9]{2}\)|[2-9][0-9]{2})[\s.-]?[2-9][0-9]{2}[\s.-]?[0-9]{4}(?![A-Fa-f0-9])', "[phone]", message)
    message = re.sub(r'(?<![0-9])[0-9]{3}-[0-9]{2}-[0-9]{4}(?![0-9])', "[ssn]", message)
    message = re.sub(r'\b(?:AKIA|ASIA)[0-9A-Z]{16}\b', "[aws-key]", message)
    message = re.sub(r'\b(?:gh[pousr]_[A-Za-z0-9]{20,}|github_pat_[A-Za-z0-9_]{20,})\b', "[token]", message)
    print(f"verify diagnostic: corpus={corpus} code={diagnostic.get('code', 'unknown')} message={message[:500]}")
PY
}

verify_corpus() {
	corpus=$1
	shift
	result_path="$verify_tmp/$corpus.json"
	if "$@" --json verify "testdata/$corpus" >"$result_path"; then
		return 0
	else
		verify_status=$?
		report_verify_failure "$corpus" "$result_path" "$verify_status"
		return "$verify_status"
	fi
}

verify_corpus phase1 "$verify_tmp/schway"
verify_corpus phase2 "$verify_tmp/schway"
verify_corpus phase3 "$verify_tmp/schway"
verify_corpus phase4 "$verify_tmp/schway"

# The sanitizer lane's own ASAN_OPTIONS/UBSAN_OPTIONS are pinned explicitly
# on this invocation (D-05-13), byte-identical to
# native.ASanOptions/native.UBSanOptions -- carried forward unchanged from
# the prior phase's own sanitizer lane, which this script is a PEER of,
# never a fork or an extension of.
verify_corpus phase5 env \
	ASAN_OPTIONS='halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1' \
	UBSAN_OPTIONS='halt_on_error=1:print_stacktrace=0' \
	"$verify_tmp/schway"
verify_corpus phase6 "$verify_tmp/schway"

# Non-regression for Phase 1 through Phase 5 is proven by running THEIR OWN
# corpora with THIS phase's freshly built binary, never by invoking an
# older gate script (D-04-21/T-04-42, carried forward unchanged). Every
# prior phase's own gate script stays byte-frozen and is neither forked,
# extended, nor invoked here -- this script is a peer of them, not a
# descendant.
grep -q 'control:backend.runtime_causality' "$verify_tmp/phase2.json" || { echo "phase6 verify: phase2 backend runtime causality control missing" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase3.json" || { echo "phase6 verify: required Phase 3 control missing: $control" >&2; exit 1; }
done
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
	control:origin.foreign_origin_omitted \
	control:defect.signal_adjudicated
do
	grep -q "$control" "$verify_tmp/phase4.json" || { echo "phase6 verify: required Phase 4 control missing: $control" >&2; exit 1; }
done
for control in \
	control:foreign.no_unproven_attributes \
	control:alias.false_no_alias \
	control:interpreter-o0-o3-lto \
	control:diagnostic.reject_program_id_equivalence \
	control:compare.field_routing_unrouted \
	control:native.sanitize.retained_pointer \
	control:native.sanitize.ubsan_no_recover \
	control:native.sanitize.allocator_mismatch \
	control:native.sanitize.use_after_free \
	control:core.attribute_unjustified \
	control:reduce.no_progress \
	control:reduce.predicate_too_loose \
	control:reduce.nondeterministic \
	control:qlt01.registry_incomplete \
	control:qlt01.stale_control_reference
do
	grep -q "$control" "$verify_tmp/phase5.json" || { echo "phase6 verify: required Phase 5 control missing: $control" >&2; exit 1; }
done

# Phase 6's own required-control set (plan 06-15). Every identifier listed
# here must also appear, verbatim, in
# internal/compiler/session/session_phase6.go's Phase6RequiredControls() --
# TestPhase6RequiredControlsMatchScript asserts the two sets are equal, so
# a control added to one and forgotten in the other fails that test.
for control in \
	control:risklanes.stale_lane_reference \
	control:risklanes.undeclared_lane \
	control:qlt02.manifest_empty \
	control:qlt02.unknown_machine \
	control:qlt02.duplicate_row \
	control:qlt02.ineligible_hard_gate \
	control:interpreter-o0-o3 \
	control:defect.match_injection \
	control:defect.move_injection \
	control:defect.borrow_injection \
	control:defect.cleanup_injection \
	control:defect.stale_evidence_injection
do
	grep -q "$control" "$verify_tmp/phase6.json" || { echo "phase6 verify: required Phase 6 control missing: $control" >&2; exit 1; }
done

# The expected escapes must appear under expected escapes and must never be
# claimed as a solved, detected control (D-06-13/D-06-29): declared,
# asserted visible, and never presented as covered.
grep -q 'escape:cache-undeclared-environment' "$verify_tmp/phase6.json" || { echo "phase6 verify: expected escape missing: escape:cache-undeclared-environment" >&2; exit 1; }
grep -q 'escape:cache-clang-version-string-stable' "$verify_tmp/phase6.json" || { echo "phase6 verify: expected escape missing: escape:cache-clang-version-string-stable" >&2; exit 1; }
grep -q 'escape:cache-nondeterministic-codegen' "$verify_tmp/phase6.json" || { echo "phase6 verify: expected escape missing: escape:cache-nondeterministic-codegen" >&2; exit 1; }
grep -q 'escape:cache-directory-hand-edited' "$verify_tmp/phase6.json" || { echo "phase6 verify: expected escape missing: escape:cache-directory-hand-edited" >&2; exit 1; }
grep -q 'escape:repair-heldout-corpus-residual-overfitting' "$verify_tmp/phase6.json" || { echo "phase6 verify: expected escape missing: escape:repair-heldout-corpus-residual-overfitting" >&2; exit 1; }

# Phase 6's own bound values, duplicated verbatim from Go (D-06-19's
# sample-count and CoV-demotion-threshold constants,
# internal/compiler/measure/statistics.go; D-06-03's explain/query bounds,
# internal/compiler/protocol/protocol.go) -- TestPhase6BoundsMatchScript
# asserts these values equal the Go constants, so this gate's own
# duplicated numbers cannot drift silently any more than a control set can.
phase6_warm_sample_count=20
phase6_cov_demotion_threshold=0.15
phase6_explain_default_depth=3
phase6_explain_max_nodes=4096
phase6_query_max_facts_per_page=64

# D-06-19's 20-warm-sample loop over this phase's own verify gate,
# SCHWAY_OBSERVE_TIMING-gated exactly like the Phase 2 gate's own observe()
# precedent. Unlike that shell-only precedent, the shell side here stays
# thin: it drives the binary WarmSampleCount times and collects raw
# elapsed_ns values, then pipes them through the shipped binary's own
# `stats` command (internal/compiler/measure.Samples.Summary(), the same
# Go statistics helper TestPhase6SamplingLoopMatchesGoStatistics pins),
# rather than reimplementing sorted-index percentile selection in sed.
observe() {
	name=$1; shift
	samples="$verify_tmp/$name.samples"; : >"$samples"
	index=0
	while [ "$index" -lt "$phase6_warm_sample_count" ]; do
		last=$(SCHWAY_OBSERVE_TIMING=1 "$verify_tmp/schway" --json "$@")
		elapsed=$(printf '%s\n' "$last" | sed -n 's/.*"elapsed_ns":\([0-9][0-9]*\).*/\1/p')
		[ -n "$elapsed" ] && [ "$elapsed" -gt 0 ] || { echo "phase6 verify: $name produced no timing" >&2; exit 1; }
		printf '%s\n' "$elapsed" >>"$samples"
		index=$((index + 1))
	done
	"$verify_tmp/schway" stats <"$samples" >"$verify_tmp/$name.stats.json"
}

observe phase6_verify_gate verify testdata/phase6

cat "$verify_tmp/phase1.json"
cat "$verify_tmp/phase2.json"
cat "$verify_tmp/phase3.json"
cat "$verify_tmp/phase4.json"
cat "$verify_tmp/phase5.json"
cat "$verify_tmp/phase6.json"
cat "$verify_tmp/phase6_verify_gate.stats.json"
