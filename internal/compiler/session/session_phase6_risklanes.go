package session

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

// FixtureInputs maps a declared-input NAME (drawn from
// cache.DeclaredInputNames()) to its current digest for one fixture
// instance -- the per-invocation input this file compares against the
// change-state file's last-seen record.
type FixtureInputs map[string]string

// changeStateFileName is the change-state file's basename. D-06-08:
// "changed" is measured against a gitignored LOCAL state file, never
// git diff and never a committed baseline, because a no-diff commit can
// still change effective risk through toolchain drift, and a committed
// baseline would import cross-machine staleness.
const changeStateFileName = "risklane-state.json"

// MaxChangeStateBytes bounds a read of the change-state file, going
// through the same readBoundedFile discipline session.go's own bounded
// reads use.
const MaxChangeStateBytes = 4 << 20

// ChangeState records the last-seen input hash per (fixture, lane) pair
// (D-06-08). Entries is keyed by laneChangeKey(fixtureID, laneID).
type ChangeState struct {
	Entries map[string]string `json:"entries"`
}

// DefaultChangeStatePath resolves the change-state file's default location
// under os.UserCacheDir() joined with "lang-verify" -- alongside the
// artifact cache from 06-04 -- NOT under the repository root.
func DefaultChangeStatePath() (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "lang-verify", changeStateFileName), nil
}

// LoadChangeState reads and strict-decodes path through the existing
// bounded-read discipline. ANY anomaly -- the file missing, unreadable,
// too large, or failing strict JSON decode -- returns an EMPTY state
// rather than an error: an empty state has no entries for any (fixture,
// lane) key, so it widens to "everything changed" by construction
// (D-06-11's rule applied to the change oracle), never a partially
// trusted parse of a corrupt prefix.
func LoadChangeState(path string) (ChangeState, error) {
	empty := ChangeState{Entries: map[string]string{}}

	data, err := readBoundedFile(path, MaxChangeStateBytes)
	if err != nil {
		return empty, nil
	}
	if len(data) > MaxChangeStateBytes {
		return empty, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var decoded ChangeState
	if err := decoder.Decode(&decoded); err != nil {
		return empty, nil
	}
	if decoded.Entries == nil {
		decoded.Entries = map[string]string{}
	}
	return decoded, nil
}

// Save writes s to path atomically (temp file in the same directory, then
// os.Rename), mirroring cache.Store.Put's write-then-rename discipline so
// a crashed run never leaves a half-written state file that a later
// LoadChangeState could misread as a legitimate (if odd) baseline.
func (s ChangeState) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries := s.Entries
	if entries == nil {
		entries = map[string]string{}
	}
	data, err := json.Marshal(ChangeState{Entries: entries})
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(dir, ".tmp-"+changeStateFileName+"-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		os.Remove(tempName)
		return err
	}
	if err := temp.Close(); err != nil {
		os.Remove(tempName)
		return err
	}
	return os.Rename(tempName, path)
}

// laneChangeKey is the (fixture, lane) composite key ChangeState.Entries
// is indexed by.
func laneChangeKey(fixtureID, laneID string) string {
	return fixtureID + "\x00" + laneID
}

// laneInputHash hashes the current digests of exactly declaredInputs (a
// lane's own declared_inputs list), in sorted-name order so the hash is
// independent of the registry's on-disk field ordering. ok is false when
// current is missing a digest for one of declaredInputs -- treated by the
// caller as widening, never as a partial hash over what happened to be
// present.
func laneInputHash(declaredInputs []string, current FixtureInputs) (hash string, ok bool) {
	names := append([]string(nil), declaredInputs...)
	sort.Strings(names)
	digest := sha256.New()
	for _, name := range names {
		value, present := current[name]
		if !present {
			return "", false
		}
		digest.Write([]byte(name))
		digest.Write([]byte{0})
		digest.Write([]byte(value))
		digest.Write([]byte{0})
	}
	return hex.EncodeToString(digest.Sum(nil)), true
}

// SelectLanesForFixture layers D-06-08's per-(fixture, lane) change-state
// oracle over the registry: for each of kind's declared lanes, it hashes
// the lane's own declared-input digests and compares against the
// last-seen hash recorded for (fixtureID, laneID) in the state file at
// statePath. A lane with no prior entry, or whose hash moved, is selected;
// everything else is deferred. computeErr and an unclassified kind still
// take SelectLanes's fail-closed widen path. cold is true when there was
// no usable prior state for this run at all (missing, corrupt, or truly
// the first invocation) -- reported honestly, never as a degraded mode.
func SelectLanesForFixture(kind, fixtureID string, current FixtureInputs, computeErr error, statePath string) (Selection, ChangeState, bool, error) {
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		return Selection{}, ChangeState{}, false, err
	}
	return selectLanesForFixtureFromRows(rows, LiveLaneIDs(), kind, fixtureID, current, computeErr, statePath)
}

// selectLanesForFixtureFromRows is SelectLanesForFixture's pure core,
// exercised directly by tests against an in-memory rows/liveLanes pair,
// mirroring selectLanesFromRows's own test seam.
func selectLanesForFixtureFromRows(rows []RiskLaneRow, liveLanes []string, kind, fixtureID string, current FixtureInputs, computeErr error, statePath string) (Selection, ChangeState, bool, error) {
	known := false
	for _, row := range rows {
		if row.FixtureKind == kind {
			known = true
			break
		}
	}

	if computeErr != nil || !known {
		sel := selectLanesFromRows(rows, liveLanes, kind, nil, computeErr)
		return sel, ChangeState{Entries: map[string]string{}}, true, nil
	}

	state, err := LoadChangeState(statePath)
	if err != nil {
		return Selection{}, ChangeState{}, false, err
	}

	sel := Selection{Reasons: map[string]string{}}
	next := ChangeState{Entries: map[string]string{}}
	for key, value := range state.Entries {
		next.Entries[key] = value
	}

	cold := true
	seenLane := map[string]bool{}
	for _, row := range rows {
		if row.FixtureKind != kind {
			continue
		}
		if seenLane[row.LaneID] {
			continue
		}
		seenLane[row.LaneID] = true

		hash, ok := laneInputHash(row.DeclaredInputs, current)
		key := laneChangeKey(fixtureID, row.LaneID)
		if !ok {
			sel.LaneIDs = append(sel.LaneIDs, row.LaneID)
			sel.Reasons[row.LaneID] = ReasonWidened + ": undeclared-input risk"
			continue
		}

		previous, seen := state.Entries[key]
		if seen {
			cold = false
		}
		if !seen || previous != hash {
			sel.LaneIDs = append(sel.LaneIDs, row.LaneID)
			sel.Reasons[row.LaneID] = fmt.Sprintf("%s: %s matched", ReasonSelected, kind)
		} else {
			sel.Reasons[row.LaneID] = ReasonDeferred + ": no declared dependency"
		}
		next.Entries[key] = hash
	}
	sort.Strings(sel.LaneIDs)
	return sel, next, cold, nil
}

// Control identifiers for the risk-lane registry audit (Task 3), following
// the qlt01.go control:qlt01.* naming convention.
const (
	// ControlRiskLaneStaleLaneReference fires when a registry row's
	// lane_id is not in the live lane-ID set produced by the shipped
	// verify functions -- a row referencing a removed or renamed lane
	// masking a real gap -- OR when a row names a declared input outside
	// cache.DeclaredInputNames(): both are the same species of stale
	// reference, a row pointing at something that does not exist.
	ControlRiskLaneStaleLaneReference = "control:risklanes.stale_lane_reference"

	// ControlRiskLaneUndeclaredLane fires when a live lane ID has no row
	// in the registry at all -- the audit's bidirectional half: a newly
	// added lane cannot silently escape declaration.
	ControlRiskLaneUndeclaredLane = "control:risklanes.undeclared_lane"
)

// LaneRiskLaneRegistryAudit is this audit's own lane identifier, following
// LaneQLT01RegistryAudit's naming convention.
const LaneRiskLaneRegistryAudit = "lane:risklanes-registry-audit"

// AuditRiskLaneRegistry cross-checks rows against liveLaneIDs and
// declaredInputs in BOTH directions (D-06-10): a declared row referencing
// a lane ID or declared-input name that does not exist fires
// control:risklanes.stale_lane_reference; a live lane with no row at all
// fires control:risklanes.undeclared_lane. An empty registry is a hard
// failure, not a vacuous pass -- an empty table passing a completeness
// audit would be exactly the false-green shape this whole phase is about.
func AuditRiskLaneRegistry(rows []RiskLaneRow, liveLaneIDs []string, declaredInputs []string) LaneResult {
	fired := map[string]bool{
		ControlRiskLaneStaleLaneReference: false,
		ControlRiskLaneUndeclaredLane:     false,
	}

	if len(rows) == 0 {
		return LaneResult{
			ID:             LaneRiskLaneRegistryAudit,
			Status:         "invalid",
			RecomputedWork: 0,
			Fired:          fired,
		}
	}

	liveSet := make(map[string]bool, len(liveLaneIDs))
	for _, id := range liveLaneIDs {
		liveSet[id] = true
	}
	declaredSet := make(map[string]bool, len(declaredInputs))
	for _, name := range declaredInputs {
		declaredSet[name] = true
	}

	declaredLanes := map[string]bool{}
	work := 0
	for _, row := range rows {
		work++ // one unit per row inspected
		if liveSet[row.LaneID] {
			declaredLanes[row.LaneID] = true
		} else {
			fired[ControlRiskLaneStaleLaneReference] = true
		}
		work++ // one unit for the lane-ID cross-check performed
		for _, name := range row.DeclaredInputs {
			work++ // one unit per declared-input cross-check performed
			if !declaredSet[name] {
				fired[ControlRiskLaneStaleLaneReference] = true
			}
		}
	}
	for _, id := range liveLaneIDs {
		work++ // one unit per bidirectional live-lane cross-check
		if !declaredLanes[id] {
			fired[ControlRiskLaneUndeclaredLane] = true
		}
	}

	status := "pass"
	var controls []string
	for _, control := range []string{ControlRiskLaneStaleLaneReference, ControlRiskLaneUndeclaredLane} {
		if fired[control] {
			status = "invalid"
		} else {
			controls = append(controls, control)
		}
	}

	return LaneResult{
		ID:             LaneRiskLaneRegistryAudit,
		Status:         status,
		Controls:       controls,
		RecomputedWork: work,
		Fired:          fired,
	}
}

// VerifyRiskLaneRegistry runs the risk-lane registry audit as a
// counted-work lane, mirroring VerifyQLT01Registry's shape exactly.
func VerifyRiskLaneRegistry(ctx context.Context) (LaneResult, error) {
	_ = ctx
	rows, err := LoadRiskLaneRegistry()
	if err != nil {
		return LaneResult{}, err
	}
	return AuditRiskLaneRegistry(rows, LiveLaneIDs(), cache.DeclaredInputNames()), nil
}
