#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase07.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestVerifyPhase7ControlsAndWork TestPhase7RequiredControlsMatchScript
sh scripts/assert-go-tests.sh --self-test ./internal/compiler/core TestAllOperationKindsHandledAtEverySite TestLinearProbeInputExercisesCallBasicFixture
sh scripts/assert-go-tests.sh --self-test ./cmd/lang TestPhase7CorpusDispatchRequiresMarker
go test ./...
go test -race ./...
go vet ./...
go build -o "$verify_tmp/lang" ./cmd/lang
go build -o "$verify_tmp/lang-repair" ./cmd/lang-repair
"$verify_tmp/lang" --json verify testdata/phase1 >"$verify_tmp/phase1.json"
"$verify_tmp/lang" --json verify testdata/phase2 >"$verify_tmp/phase2.json"
"$verify_tmp/lang" --json verify testdata/phase3 >"$verify_tmp/phase3.json"
"$verify_tmp/lang" --json verify testdata/phase4 >"$verify_tmp/phase4.json"

# The sanitizer lane's own ASAN_OPTIONS/UBSAN_OPTIONS are pinned explicitly
# on this invocation (D-05-13), byte-identical to
# native.ASanOptions/native.UBSanOptions -- carried forward unchanged from
# every prior phase's own sanitizer lane, which this script is a PEER of,
# never a fork or an extension of.
ASAN_OPTIONS='halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1' \
UBSAN_OPTIONS='halt_on_error=1:print_stacktrace=0' \
	"$verify_tmp/lang" --json verify testdata/phase5 >"$verify_tmp/phase5.json"
"$verify_tmp/lang" --json verify testdata/phase6 >"$verify_tmp/phase6.json"
"$verify_tmp/lang" --json verify testdata/phase07 >"$verify_tmp/phase07.json"

# Non-regression for Phase 1 through Phase 6 is proven by running THEIR OWN
# corpora with THIS phase's freshly built binary, never by invoking an
# older gate script (D-04-21/T-04-42, carried forward unchanged). Every
# prior phase's own gate script stays byte-frozen and is neither forked,
# extended, nor invoked here -- this script is a peer of them, not a
# descendant.
grep -q 'control:backend.runtime_causality' "$verify_tmp/phase2.json" || { echo "phase07 verify: phase2 backend runtime causality control missing" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase3.json" || { echo "phase07 verify: required Phase 3 control missing: $control" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase4.json" || { echo "phase07 verify: required Phase 4 control missing: $control" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase5.json" || { echo "phase07 verify: required Phase 5 control missing: $control" >&2; exit 1; }
done
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
	grep -q "$control" "$verify_tmp/phase6.json" || { echo "phase07 verify: required Phase 6 control missing: $control" >&2; exit 1; }
done

# Phase 07's own required-control set (plan 07-04). Every identifier
# listed here must also appear, verbatim, in
# internal/compiler/session/session_phase7.go's Phase7RequiredControls() --
# TestPhase7RequiredControlsMatchScript asserts the two sets are equal, so
# a control added to one and forgotten in the other fails that test.
for control in \
	control:kind.exhaustive_dispatch.phase07_in_process \
	control:kind.exhaustive_dispatch.phase07_lane \
	control:dispatch.recognized_not_executed \
	control:call.admission_body_blind \
	control:call.callable_refusal
do
	grep -q "$control" "$verify_tmp/phase07.json" || { echo "phase07 verify: required Phase 07 control missing: $control" >&2; exit 1; }
done

cat "$verify_tmp/phase1.json"
cat "$verify_tmp/phase2.json"
cat "$verify_tmp/phase3.json"
cat "$verify_tmp/phase4.json"
cat "$verify_tmp/phase5.json"
cat "$verify_tmp/phase6.json"
cat "$verify_tmp/phase07.json"
