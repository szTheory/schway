package session

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/execution"
)

// Axis identifiers (D-05-20): every axis Phase5CompareEngines and its
// diagnostic-ID sibling Phase5CompareDiagnosticIDs compare, each an
// explicitly named branch rather than a struct equality. SC1's wording
// silently omitted exit status/signal and reject-program diagnostic
// equivalence from its comparison scope; these five names make every axis
// -- including the two SC1 omitted -- an explicit, reviewable identifier
// rather than an implicit side effect of a whole-document comparison.
const (
	AxisTerminalOutcome  = "axis:terminal-outcome"
	AxisEventOrder       = "axis:event-order"
	AxisResourceLedger   = "axis:resource-ledger"
	AxisExitStatusSignal = "axis:exit-status-signal"
	AxisDiagnosticID     = "axis:diagnostic-id"
)

// ControlDiagnosticRejectProgramIDEquivalence is D-05-20's own control ID
// for reject-program diagnostic-ID equivalence, kept distinct from
// control:interpreter-o0-o3(-lto): a reject-program never executes, so its
// evidence is a diagnostic ID, never an execution document, and folding it
// into the execution-comparison control would misrepresent what was
// actually proven.
const ControlDiagnosticRejectProgramIDEquivalence = "control:diagnostic.reject_program_id_equivalence"

// Phase5EngineDisagreement names the fixture, the specific engine pair, the
// diverging axis, and the first diverging operation identity for a
// Phase5CompareEngines/Phase5CompareDiagnosticIDs disagreement (D-05-20).
// Naming the axis explicitly (rather than a generic Detail string, as
// Phase4EngineDisagreement does) is what lets a caller programmatically
// distinguish "the outcome payload diverged" from "the resource ledger
// diverged" without parsing prose.
type Phase5EngineDisagreement struct {
	Fixture     string
	EnginePair  string
	Axis        string
	OperationID string
	Detail      string
}

func (d *Phase5EngineDisagreement) Error() string {
	return fmt.Sprintf("phase5 engine disagreement: fixture=%s pair=%s axis=%s operation=%q detail=%s", d.Fixture, d.EnginePair, d.Axis, d.OperationID, d.Detail)
}

func countReleased(events []execution.Event) int {
	count := 0
	for _, event := range events {
		if event.Kind == "resource.released" {
			count++
		}
	}
	return count
}

// Phase5CompareEngines compares every pair in engines (keyed by engine
// name -- "interpreter", "O0", "O3", and "O3-LTO" when present, the last
// only for the adversarial subset) across the four execution-bearing axes
// and returns a *Phase5EngineDisagreement naming the fixture, engine pair,
// diverging axis, and first diverging operation identity on any
// disagreement. Reject-programs never reach this function -- see
// Phase5CompareDiagnosticIDs for axis:diagnostic-id.
func Phase5CompareEngines(fixture string, engines map[string]execution.Execution) error {
	names := make([]string, 0, len(engines))
	for name := range engines {
		names = append(names, name)
	}
	sort.Strings(names)
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			left, right := names[i], names[j]
			pair := left + "-vs-" + right
			if err := comparePhase5Pair(fixture, pair, engines[left], engines[right]); err != nil {
				return err
			}
		}
	}
	return nil
}

func comparePhase5Pair(fixture, pair string, left, right execution.Execution) error {
	// axis:terminal-outcome -- the closed value|typed_failure|defect set,
	// INCLUDING the ok payload and the err-edge ADT alternative. Schema is
	// compared here too: a schema mismatch makes the two documents
	// incomparable, which is itself a terminal-outcome-level disagreement.
	if left.Schema != right.Schema || left.Outcome.Kind != right.Outcome.Kind || left.Outcome.Value != right.Outcome.Value {
		return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisTerminalOutcome, Detail: fmt.Sprintf("schema=%s/%s outcome=%+v vs %+v", left.Schema, right.Schema, left.Outcome, right.Outcome)}
	}

	// axis:event-order -- semantic-event order with causal identity,
	// compared as an ordered sequence of (operation ordinal, event kind,
	// causal role). Event carries FunctionID/SourcePlace/TargetPlace/TypeID
	// as its causal-role facts alongside Kind, so a whole-struct equality at
	// each ordinal is exactly this comparison.
	length := len(left.Events)
	if len(right.Events) > length {
		length = len(right.Events)
	}
	for index := 0; index < length; index++ {
		switch {
		case index >= len(left.Events):
			return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisEventOrder, OperationID: fmt.Sprintf("%d", index), Detail: fmt.Sprintf("missing vs %+v", right.Events[index])}
		case index >= len(right.Events):
			return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisEventOrder, OperationID: fmt.Sprintf("%d", index), Detail: fmt.Sprintf("%+v vs missing", left.Events[index])}
		case left.Events[index] != right.Events[index]:
			return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisEventOrder, OperationID: left.Events[index].ID, Detail: fmt.Sprintf("%+v vs %+v", left.Events[index], right.Events[index])}
		}
	}

	// axis:resource-ledger -- the live-resource ledger, including
	// released-count (a live ledger that happens to match while the two
	// engines released a different NUMBER of resources along the way is
	// still a real divergence).
	leftLive := append([]string(nil), left.LiveResources...)
	rightLive := append([]string(nil), right.LiveResources...)
	sort.Strings(leftLive)
	sort.Strings(rightLive)
	if !reflect.DeepEqual(leftLive, rightLive) {
		return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisResourceLedger, Detail: fmt.Sprintf("live_resources: %+v vs %+v", leftLive, rightLive)}
	}
	if leftReleased, rightReleased := countReleased(left.Events), countReleased(right.Events); leftReleased != rightReleased {
		return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisResourceLedger, Detail: fmt.Sprintf("released-count: %d vs %d", leftReleased, rightReleased)}
	}

	// axis:exit-status-signal -- added by D-05-20 because SC1 silently
	// omitted it; abort/defect programs have no other observable that a
	// payload-shaped comparison would catch. Compared as its own
	// independent field pair, never derived from Outcome.Kind, so a
	// genuine divergence here is detectable even when Outcome.Kind agrees.
	if left.ExitSignaled != right.ExitSignaled || left.ExitSignal != right.ExitSignal {
		return &Phase5EngineDisagreement{Fixture: fixture, EnginePair: pair, Axis: AxisExitStatusSignal, Detail: fmt.Sprintf("signaled=%v/%v signal=%q/%q", left.ExitSignaled, right.ExitSignaled, left.ExitSignal, right.ExitSignal)}
	}

	return nil
}

// Phase5CompareDiagnosticIDs implements axis:diagnostic-id
// (control:diagnostic.reject_program_id_equivalence): for a reject-program,
// its diagnostic ID must be identical across every independent derivation
// named in diagnostics (keyed by derivation name). Reject-programs never
// execute, so this comparison never touches execution.Execution -- it is a
// distinct function under its own control ID, never folded into
// control:interpreter-o0-o3(-lto).
func Phase5CompareDiagnosticIDs(fixture string, diagnostics map[string]diagnostic.Diagnostic) error {
	names := make([]string, 0, len(diagnostics))
	for name := range diagnostics {
		names = append(names, name)
	}
	sort.Strings(names)
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			left, right := names[i], names[j]
			if diagnostics[left].ID != diagnostics[right].ID {
				return &Phase5EngineDisagreement{
					Fixture: fixture, EnginePair: left + "-vs-" + right, Axis: AxisDiagnosticID,
					Detail: fmt.Sprintf("%s vs %s", diagnostics[left].ID, diagnostics[right].ID),
				}
			}
		}
	}
	return nil
}

// Phase5ComparedComparisonFields lists every execution.Execution-reachable
// leaf field path the axes above actually read (D-05-21). Each entry is
// the exact struct field path, not a category label.
var Phase5ComparedComparisonFields = []string{
	"Execution.Schema",
	"Execution.Outcome.Kind",
	"Execution.Outcome.Value",
	"Execution.Events.Schema",
	"Execution.Events.ID",
	"Execution.Events.Kind",
	"Execution.Events.FunctionID",
	"Execution.Events.Input",
	"Execution.Events.Output",
	"Execution.Events.SourcePlace",
	"Execution.Events.TargetPlace",
	"Execution.Events.TypeID",
	"Execution.LiveResources",
	"Execution.ExitSignaled",
	"Execution.ExitSignal",
}

// Phase5ExcludedComparisonFields lists every execution.Execution-reachable
// leaf field path Phase5CompareEngines deliberately never reads (D-05-21):
// addresses, wall-clock time, allocator identity, hash/PRNG seed, thread
// identity. Each entry is the exact struct field path it excludes, not a
// category label -- and, per D-05-21, TestComparisonFieldRoutingIsExhaustive
// makes this list, not prose, the enforcement mechanism.
//
// execution.Execution's own pre-Phase-5 fields carry no address/timestamp/
// allocator/PRNG/thread-identity data (this project's place/type/event IDs
// are function-local semantic ordinals, never addresses or source
// offsets -- see STATE.md's Phase 2 decisions). AllocatorAddress and
// WallClockNanos are therefore new, deliberately unpopulated-by-production-
// code fields added by this plan specifically so this exclusion list has
// real struct-field targets to route to, rather than staying an empty slice
// that could never demonstrate the fail-closed routing test actually
// excludes anything (D-05-21 Rule 2: the routing machinery needs a genuine
// excluded field to prove it is load-bearing, not merely present).
var Phase5ExcludedComparisonFields = []string{
	"Execution.AllocatorAddress",
	"Execution.WallClockNanos",
}

// reachableFieldPaths walks, by reflection, every exported field of t and
// every struct type transitively reachable from it (through structs,
// pointers-to-struct, and slices-of-struct), returning the set of full leaf
// field paths. A "leaf" is any field whose (de-referenced/de-sliced) type is
// not itself a struct -- so a struct-typed field contributes its own
// children's paths, never a path of its own, matching the discipline that
// only actual DATA is routed to compared/excluded, not the containers
// holding it.
func reachableFieldPaths(t reflect.Type, prefix string) []string {
	var paths []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // unexported
			continue
		}
		path := prefix + "." + field.Name
		fieldType := field.Type
		for fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			paths = append(paths, reachableFieldPaths(fieldType, path)...)
			continue
		}
		paths = append(paths, path)
	}
	return paths
}

// unroutedFields returns every entry of actual that is present in neither
// compared nor excluded -- a NEW field that has not been routed anywhere,
// which must fail the build (D-05-21).
func unroutedFields(actual, compared, excluded []string) []string {
	routed := make(map[string]bool, len(compared)+len(excluded))
	for _, field := range compared {
		routed[field] = true
	}
	for _, field := range excluded {
		routed[field] = true
	}
	var unrouted []string
	for _, field := range actual {
		if !routed[field] {
			unrouted = append(unrouted, field)
		}
	}
	sort.Strings(unrouted)
	return unrouted
}

// staleFields returns every entry of the compared+excluded union that no
// longer names a real field of actual -- a routed field that was removed or
// renamed out from under the routing tables.
func staleFields(actual, compared, excluded []string) []string {
	present := make(map[string]bool, len(actual))
	for _, field := range actual {
		present[field] = true
	}
	var stale []string
	seen := make(map[string]bool)
	for _, field := range append(append([]string(nil), compared...), excluded...) {
		if !present[field] && !seen[field] {
			stale = append(stale, field)
			seen[field] = true
		}
	}
	sort.Strings(stale)
	return stale
}
