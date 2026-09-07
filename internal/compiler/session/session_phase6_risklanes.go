package session

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/codename-lang/lang/internal/compiler/cache"

	_ "embed"
)

//go:embed risk_lanes.json
var riskLanesBytes []byte

// RiskLaneRow is one declared row of the change-risk selection table
// (D-06-10): which lanes exist for a given fixture kind, and which of
// cache.DeclaredInputNames() each lane's risk actually depends on. This is
// a checked-in, reviewed-like-code declaration, copying qlt01.go's
// embedded-registry shape verbatim -- never a derived dependency graph.
type RiskLaneRow struct {
	FixtureKind    string   `json:"fixture_kind"`
	LaneID         string   `json:"lane_id"`
	DeclaredInputs []string `json:"declared_inputs"`
	Rationale      string   `json:"rationale"`
}

// LoadRiskLaneRegistry parses the embedded risk_lanes.json registry.
func LoadRiskLaneRegistry() ([]RiskLaneRow, error) {
	return parseRiskLaneRegistry(riskLanesBytes)
}

// parseRiskLaneRegistry is LoadRiskLaneRegistry's core, exported indirectly
// so tests can exercise the malformed-bytes error path through the exact
// same code LoadRiskLaneRegistry uses, without needing to swap out the
// embedded risk_lanes.json (mirroring QLT01LaneFromRows's own test seam).
func parseRiskLaneRegistry(data []byte) ([]RiskLaneRow, error) {
	var rows []RiskLaneRow
	if err := json.Unmarshal(data, &rows); err != nil {
		return nil, fmt.Errorf("risklanes: failed to parse embedded registry: %w", err)
	}
	return rows, nil
}

// D-06-12's closed selection-reason vocabulary. These are the exact three
// prefixes 06-07 renders on protocol.Lane.selection_reason -- "selected: X
// matched" / "deferred: no declared dependency" / "widened: undeclared-input
// risk" -- so a future caller cannot silently invent a fourth reason shape.
const (
	ReasonSelected = "selected"
	ReasonDeferred = "deferred"
	ReasonWidened  = "widened"
)

// Selection is SelectLanes's report: which lanes were selected for this
// run, a reason for every lane in scope (selected, deferred, or widened --
// never absent), and the cache status this selection implies for the run
// as a whole. CacheStatus is only meaningfully populated on the widened
// path (cache.StatusNotCacheable); an ordinary matched/deferred selection
// leaves it at its zero value, since cache-status reporting for a
// cacheable run is cache.Consult's own responsibility (06-04), not this
// selector's.
type Selection struct {
	LaneIDs     []string
	Reasons     map[string]string
	CacheStatus cache.CacheStatus
}

// SelectLanes resolves which of kind's declared lanes should run this
// invocation, given the set of declared-input NAMES (drawn from
// cache.DeclaredInputNames()) that changed since the last recorded
// baseline. computeErr is non-nil when the fixture's declared-input set
// itself could not be fully computed this run (for example a failed Clang
// identity probe) -- distinct from an ordinary "nothing changed" empty
// changed list.
//
// D-06-11's conservative widening is fail-closed and unconditional: any
// fixture kind not exhaustively classified in the registry, OR any
// declared-input computation failure, selects ALL of kind's lanes (or,
// for an entirely unclassified kind, every live lane) with reason
// ReasonWidened and cache.StatusNotCacheable. There is NO code path from
// an unclassified kind or a failed computation to a deferred lane.
func SelectLanes(kind string, changed []string, computeErr error) (Selection, error) {
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		return Selection{}, err
	}
	return selectLanesFromRows(rows, LiveLaneIDs(), kind, changed, computeErr), nil
}

// selectLanesFromRows is SelectLanes's pure core, exported indirectly
// through SelectLanes but exercised directly by tests against an in-memory
// rows/liveLanes pair -- mirroring QLT01LaneFromRows's own test seam --
// without needing to swap out the embedded risk_lanes.json or LiveLaneIDs.
func selectLanesFromRows(rows []RiskLaneRow, liveLanes []string, kind string, changed []string, computeErr error) Selection {
	sel := Selection{Reasons: map[string]string{}}

	known := false
	for _, row := range rows {
		if row.FixtureKind == kind {
			known = true
			break
		}
	}

	if computeErr != nil || !known {
		var target []string
		if known {
			for _, row := range rows {
				if row.FixtureKind == kind {
					target = append(target, row.LaneID)
				}
			}
		} else {
			target = append(target, liveLanes...)
		}
		seenLane := map[string]bool{}
		for _, id := range target {
			if seenLane[id] {
				continue
			}
			seenLane[id] = true
			sel.LaneIDs = append(sel.LaneIDs, id)
			sel.Reasons[id] = ReasonWidened + ": undeclared-input risk"
		}
		sort.Strings(sel.LaneIDs)
		sel.CacheStatus = cache.StatusNotCacheable
		return sel
	}

	changedSet := make(map[string]bool, len(changed))
	for _, name := range changed {
		changedSet[name] = true
	}

	seenLane := map[string]bool{}
	for _, row := range rows {
		if row.FixtureKind != kind {
			continue
		}
		if seenLane[row.LaneID] {
			// A lane matched by two different rows is selected exactly
			// once, with the FIRST matching row's reason by registry
			// order -- a specified rule, so the tie has a defined winner.
			continue
		}
		seenLane[row.LaneID] = true

		matched := false
		for _, name := range row.DeclaredInputs {
			if changedSet[name] {
				matched = true
				break
			}
		}
		if matched {
			sel.LaneIDs = append(sel.LaneIDs, row.LaneID)
			sel.Reasons[row.LaneID] = fmt.Sprintf("%s: %s matched", ReasonSelected, kind)
		} else {
			sel.Reasons[row.LaneID] = ReasonDeferred + ": no declared dependency"
		}
	}
	sort.Strings(sel.LaneIDs)
	return sel
}

// LiveLaneIDs is the union of every shipped verify function's own lane
// identifiers, built from per-corpus-dispatch functions (mirroring
// AllShippedControlIDs()'s discipline: composed from named category
// functions, never a single hand-copied blob) so it cannot go stale
// independently of the code that actually emits these lane IDs.
// session_phase6_risklanes_test.go cross-checks every one of these strings
// against a literal grep of session.go/session_phase5.go's own addLane
// call sites, so a renamed or removed lane in either file cannot silently
// go undetected here.
func LiveLaneIDs() []string {
	seen := make(map[string]bool)
	var all []string
	add := func(ids []string) {
		for _, id := range ids {
			if !seen[id] {
				seen[id] = true
				all = append(all, id)
			}
		}
	}
	add(liveLanesPureMatch())
	add(liveLanesOwned())
	add(liveLanesBorrowed())
	add(liveLanesForeign())
	add(liveLanesPhase5Adversarial())
	sort.Strings(all)
	return all
}

// liveLanesPureMatch names every lane VerifyCorpus (session.go) emits.
func liveLanesPureMatch() []string {
	return []string{
		"lane:deterministic",
		"lane:syntax-properties",
		"lane:negative-controls",
		"lane:evidence-bindings",
		"lane:native-differential",
	}
}

// liveLanesOwned names every lane verifyOwnedCorpus (session.go) emits.
func liveLanesOwned() []string {
	return []string{
		"lane:owned-negative-controls",
		"lane:owned-core-controls",
		"lane:owned-native-differential",
		"lane:owned-backend-causality",
		"lane:owned-evidence-bindings",
	}
}

// liveLanesBorrowed names every lane verifyBorrowedCorpus and
// VerifyCorpusFile (session.go) emit for the borrowed corpus.
func liveLanesBorrowed() []string {
	return []string{
		"lane:borrowed-negative-controls",
		"lane:borrowed-origin-controls",
		"lane:borrowed-loan-endpoint-control",
		"lane:path-oracle-disagreement",
	}
}

// liveLanesForeign names every lane verifyForeignCorpus (session.go) emits.
func liveLanesForeign() []string {
	return []string{
		"lane:foreign-unwind-policy-undeclared",
		"lane:foreign-call-target-not-foreign",
		"lane:foreign-acquire-admitted",
		"lane:release-order-transposed",
		"lane:release-omitted",
		"lane:foreign-layout-mismatch",
		"lane:foreign-no-unproven-attributes",
		"lane:foreign-origin-omitted",
		"lane:foreign-unwind-forbidden",
	}
}

// liveLanesPhase5Adversarial names every lane
// VerifyPhase5ControlsAndWork (session_phase5.go and its siblings) emit.
func liveLanesPhase5Adversarial() []string {
	return []string{
		"lane:alias-false-no-alias",
		"lane:attribute-unjustified",
		"lane:coordinated-lie-escape",
		"lane:defect-no-release",
		"lane:defect-signal-adjudicated",
		"lane:diagnostic-reject-program-id-equivalence",
		"lane:interpreter-o0-o3-lto",
		"lane:kind-exhaustive-dispatch",
		"lane:mismatch-reduce",
		"lane:native-sanitize",
		"lane:native-sanitize-obligations",
		"lane:nonlocal-exit-undetected",
		"lane:qlt01-registry-audit",
		"lane:terminator-walk-complete",
		"lane:compare-field-routing",
	}
}
