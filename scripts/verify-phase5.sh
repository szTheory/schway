#!/bin/sh
set -eu

verify_tmp=$(mktemp -d "${TMPDIR:-/tmp}/codename-lang-phase5.XXXXXX")
trap 'rm -rf "$verify_tmp"' EXIT HUP INT TERM
export GOCACHE="$verify_tmp/go-cache"

sh scripts/assert-go-tests.sh --self-test ./internal/compiler/session TestTogglePipeline TestVerifyPhase5ControlsAndWork TestPhase5RequiredControlsMatchScript TestPhase5CorpusBoundMatchesScript TestPhase5VerifierScriptContract TestPhase5SanitizerOptionsMatchScript
go test ./...
go test -race ./...
go vet ./...
go build -o "$verify_tmp/lang" ./cmd/lang
"$verify_tmp/lang" --json verify testdata/phase1 >"$verify_tmp/phase1.json"
"$verify_tmp/lang" --json verify testdata/phase2 >"$verify_tmp/phase2.json"
"$verify_tmp/lang" --json verify testdata/phase3 >"$verify_tmp/phase3.json"
"$verify_tmp/lang" --json verify testdata/phase4 >"$verify_tmp/phase4.json"

# The sanitizer lane's own ASAN_OPTIONS/UBSAN_OPTIONS are pinned explicitly
# on this invocation (D-05-13), never inherited from the calling shell's
# environment -- byte-identical to native.ASanOptions/native.UBSanOptions,
# asserted by TestPhase5SanitizerOptionsMatchScript. The Go binary also
# pins these internally when it spawns the sanitizer binary (sanitize.go);
# this is a second, script-layer pin so a future change removing the
# internal one still fails loudly here rather than silently inheriting a
# hostile or merely-absent environment (T-05-35).
ASAN_OPTIONS='halt_on_error=1:abort_on_error=1:symbolize=0:detect_leaks=0:detect_odr_violation=0:alloc_dealloc_mismatch=1' \
UBSAN_OPTIONS='halt_on_error=1:print_stacktrace=0' \
	"$verify_tmp/lang" --json verify testdata/phase5 >"$verify_tmp/phase5.json"

# Non-regression for Phase 1 through Phase 4 is proven by running THEIR OWN
# corpora with THIS phase's freshly built binary, never by invoking an
# older gate script (D-04-21/T-04-42, carried forward unchanged). The
# prior phase's own gate script stays byte-frozen and is neither forked,
# extended, nor invoked here -- this script is a peer of it, not an
# extension.
grep -q 'control:backend.runtime_causality' "$verify_tmp/phase2.json" || { echo "phase5 verify: phase2 backend runtime causality control missing" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase3.json" || { echo "phase5 verify: required Phase 3 control missing: $control" >&2; exit 1; }
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
	grep -q "$control" "$verify_tmp/phase4.json" || { echo "phase5 verify: required Phase 4 control missing: $control" >&2; exit 1; }
done

# Phase 5's own required-control set (D-05-17). Every identifier listed
# here must also appear, verbatim, in
# internal/compiler/session/session_phase5.go's Phase5RequiredControls() --
# TestPhase5RequiredControlsMatchScript asserts the two sets are equal, so
# a control added to one and forgotten in the other fails that test.
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
	control:core.attribute_unjustified
do
	grep -q "$control" "$verify_tmp/phase5.json" || { echo "phase5 verify: required Phase 5 control missing: $control" >&2; exit 1; }
done

# The expected escape must appear under expected escapes and must never be
# claimed as a solved, detected control (D-05-07): declared, asserted
# visible, and never presented as covered.
grep -q 'escape:callback-invocation-unsubjected' "$verify_tmp/phase5.json" || { echo "phase5 verify: expected escape missing: escape:callback-invocation-unsubjected" >&2; exit 1; }

# Phase 5's own corpus bound (D-05-18). Duplicated verbatim from
# internal/compiler/session/session_phase5_corpus.go's own
# Phase5CorpusBoundVersion/Phase5EnumerationMaxDepth/
# Phase5EnumerationMaxStatements constants -- TestPhase5CorpusBoundMatchesScript
# asserts these three values equal the Go constants, so the enumerated
# closure's own scope cannot drift silently any more than a control set
# can.
phase5_corpus_bound_version=1
phase5_enumeration_max_depth=3
phase5_enumeration_max_statements=4

cat "$verify_tmp/phase1.json"
cat "$verify_tmp/phase2.json"
cat "$verify_tmp/phase3.json"
cat "$verify_tmp/phase4.json"
cat "$verify_tmp/phase5.json"
