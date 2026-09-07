package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/cache"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
)

// Schema is the frozen /0 record for lang.command. Its bytes must remain
// reproducible forever (D-06-31); Schema1 is the coordinated additive bump
// every new Result now carries, following diagnostic.go's/evidence.go's own
// two-constant coexistence shape.
const Schema = "lang.command/0"

// Schema1 is the coordinated lang.command/1 bump (D-06-31): protocol.New()
// now returns Schema1, while Schema ("lang.command/0") remains declared and
// bound to its original string so already-published /0 documents stay
// reproducible.
const Schema1 = "lang.command/1"

// LaneSchema is the frozen /0 record for lang.verify-lane. Its bytes must
// remain reproducible forever (D-06-31); LaneSchema1 is the coordinated
// additive bump landing at the same commit as Schema1.
const LaneSchema = "lang.verify-lane/0"

// LaneSchema1 is the coordinated lang.verify-lane/1 bump (D-06-31): every one
// of the 12 lane-schema composite literal sites across session.go,
// session_phase5.go, session_phase5_mismatch.go, and
// session_phase5_sanitize.go moves to this constant in the same commit as
// Schema1. LaneSchema ("lang.verify-lane/0") remains declared and bound to
// its original string so already-published /0 lane documents stay
// reproducible.
const LaneSchema1 = "lang.verify-lane/1"

const (
	StatusPass        = "pass"
	StatusInvalid     = "invalid"
	StatusOperational = "operational_failure"
	StatusMismatch    = "semantic_mismatch"
	StatusUsage       = "usage_error"
	// StatusDeferred is D-06-12's lane-status value for "not run this
	// invocation": a lane the selector deferred, or that this invocation
	// otherwise did not execute. It is structurally distinct from StatusPass
	// -- a lane that was not run must never render as pass (T-06-VERIFY-02).
	StatusDeferred = "deferred"
)

// LaneStatuses returns the closed lane-status vocabulary this project's
// verify paths may emit on a Lane's Status field. It exists so an
// exhaustiveness test can assert every emitted status is one of these
// values -- a status outside this set is, by construction, not a value any
// shipped lane may report.
func LaneStatuses() []string {
	return []string{StatusPass, StatusInvalid, StatusOperational, StatusMismatch, StatusUsage, StatusDeferred}
}

type Metrics struct {
	ElapsedNS      int64  `json:"elapsed_ns"`
	PeakRSSStatus  string `json:"peak_rss_status"`
	PeakRSSBytes   int64  `json:"peak_rss_bytes,omitempty"`
	OutputBytes    int    `json:"output_bytes"`
	RecomputedWork int    `json:"recomputed_work"`
	// CacheInputsReusedCount counts artifacts this invocation reused from
	// the local cache (D-06-12), in the same counted-unit currency as
	// RecomputedWork -- a plain int counter, no unit conversion between the
	// two. Never reaches Result.Finalize()'s identity struct (D-06-32).
	CacheInputsReusedCount int `json:"cache_inputs_reused_count,omitempty"`
}

type EvidenceSummary struct {
	Schema string `json:"schema"`
	ID     string `json:"id"`
	Digest string `json:"digest"`
}

// InterfaceFunctionAnswer is one function's body-blind origin answer, read
// directly from a core.Interface summary — never from a body field.
type InterfaceFunctionAnswer struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Paths  []string `json:"paths,omitempty"`
	Access string   `json:"access,omitempty"`
}

// InterfaceSummary is the `interface export`/`interface check` command
// projection: module identity, the digest binding the summary to its
// producing core artifact, and each function's body-blind origin answer.
type InterfaceSummary struct {
	Schema     string                    `json:"schema"`
	ModuleID   string                    `json:"module_id"`
	CoreDigest string                    `json:"core_digest"`
	Functions  []InterfaceFunctionAnswer `json:"functions"`
}

// DebugMapEntry is one joined lineage fact from the debugmap package's
// bounded debug-lineage experiment (D-01..D-04): a source span joined to a
// core ID and an operation/point identity, or an honest not_captured /
// optimized_out report when the compiler genuinely has no value.
type DebugMapEntry struct {
	ID           string `json:"id"`
	CoreID       string `json:"core_id,omitempty"`
	OperationID  string `json:"operation_id,omitempty"`
	PointID      string `json:"point_id,omitempty"`
	Kind         string `json:"kind,omitempty"`
	Availability string `json:"availability"`
}

// DebugMapSummary is the `debug-map` command projection.
type DebugMapSummary struct {
	Schema  string          `json:"schema"`
	Entries []DebugMapEntry `json:"entries"`
}

// ExplainSchema versions the `lang explain` command's cause-DAG projection
// independently of every other schema in the compiler. explain is a
// net-new capability (D-06-04): it gets a net-new /0 schema rather than
// folding into Schema ("lang.command/0"), and does not trigger that
// constant's coordinated /1 bump on its own account.
const ExplainSchema = "lang.explain/0"

// Edge kind vocabulary for ExplainSummary.Edges is closed to exactly these
// three values (D-06-03's Claude's-discretion edge typing).
const (
	EdgeCausedBy    = "caused_by"
	EdgeNarrows     = "narrows"
	EdgeSameBinding = "same_binding"
)

// ExplainDefaultDepth is the cause-DAG expansion depth used when `--depth`
// is not supplied (D-06-03).
const ExplainDefaultDepth = 3

// ExplainMaxNodes bounds the number of nodes a single ExplainSummary may
// contain, sized like debugmap.MaxEntries's order of magnitude (D-06-03):
// exceeding it stops expansion and reports a stable truncation code rather
// than growing the output.
const ExplainMaxNodes = 4096

// ExplainNode is one node in the synthesized cause DAG: either the
// diagnostic itself (the root, addressed by the diagnostic's own ID) or one
// of its Causes. Availability reuses the existing debugmap.Availability
// vocabulary (available/optimized_out/not_captured) rather than fabricating
// a value the compiler does not hold (D-06-02).
type ExplainNode struct {
	ID           string           `json:"id"`
	Kind         string           `json:"kind"`
	Detail       string           `json:"detail,omitempty"`
	Span         *diagnostic.Span `json:"span,omitempty"`
	Availability string           `json:"availability"`
}

// ExplainEdge is one typed relation between two ExplainNode IDs. Kind is one
// of EdgeCausedBy/EdgeNarrows/EdgeSameBinding — the vocabulary is closed.
type ExplainEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Kind string `json:"kind"`
}

// ExplainSummary is the `lang explain` command projection: a bounded,
// deterministic cause DAG synthesized fresh per cold invocation from a
// diagnostic's existing flat Causes list (D-06-02, D-06-04). Truncated
// carries a stable code ("truncated:explain.depth" or
// "truncated:explain.node_budget") when either bound is hit, and is empty
// otherwise.
type ExplainSummary struct {
	Schema    string        `json:"schema"`
	RootID    string        `json:"root_id"`
	Nodes     []ExplainNode `json:"nodes"`
	Edges     []ExplainEdge `json:"edges"`
	Truncated string        `json:"truncated,omitempty"`
}

// QuerySchema versions `lang query`'s joined-addressing-surface projection
// independently of every other schema in the compiler. query is a net-new
// capability (D-06-04): it gets a net-new /0 schema rather than folding into
// Schema ("lang.command/0"), and does not trigger that constant's
// coordinated /1 bump on its own account (D-06-31 names the actual
// coordinated bump separately).
const QuerySchema = "lang.query/0"

// QueryMaxFactsPerPage bounds how many QueryFact entries a single
// QuerySummary page may carry (D-06-03: query pages by cursor, never by
// depth). Deliberately smaller than ExplainMaxNodes/debugmap.MaxEntries --
// query's own bounding discipline is "smallest sufficient context by
// default" (wiki/compute-efficiency-constitution.md), so its default page is
// an order of magnitude below those two full-artifact caps rather than
// matching them.
const QueryMaxFactsPerPage = 64

// QueryFact is one resolved (or honestly absent) fact about the address a
// `lang query` call named. Vocabulary carries the fine-grained ID kind that
// was actually matched (e.g. "diagnostic", "operation_id", "core_id",
// "point_id", "evidence", "control", "lane", or the "unrecognized" sentinel
// when the address matched none of D-06-01's five vocabularies) --
// Availability reuses debugmap's existing honest three-value vocabulary
// rather than fabricating a new one.
type QueryFact struct {
	ID           string           `json:"id"`
	Kind         string           `json:"kind,omitempty"`
	Vocabulary   string           `json:"vocabulary"`
	Detail       string           `json:"detail,omitempty"`
	Span         *diagnostic.Span `json:"span,omitempty"`
	Availability string           `json:"availability"`
}

// QuerySummary is the `lang query` command projection: a bounded,
// cursor-paginated list of facts resolved against D-06-01's one joined
// addressing surface. NextCursor is set only when more facts remain past
// this page; Truncated carries the stable "truncated:query.page_bound" code
// in that case.
type QuerySummary struct {
	Schema     string      `json:"schema"`
	Query      string      `json:"query"`
	Facts      []QueryFact `json:"facts"`
	NextCursor string      `json:"next_cursor,omitempty"`
	Truncated  string      `json:"truncated,omitempty"`
}

// StageTiming is one named compilation stage's elapsed duration (D-06-21),
// appended to Lane.StageBreakdown. Like every field on StageTiming, it is
// gated by the LANG_OBSERVE_TIMING convention (protocol.go's completeCommand
// analog in session.go) and never populated unconditionally -- an
// unconditional time.Since would perturb a golden/pinned-JSON test.
type StageTiming struct {
	Stage     string `json:"stage"`
	ElapsedNS int64  `json:"elapsed_ns"`
}

type Lane struct {
	Schema         string   `json:"schema"`
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Controls       []string `json:"controls"`
	RecomputedWork int      `json:"recomputed_work"`
	ElapsedNS      int64    `json:"elapsed_ns"`
	PeakRSSStatus  string   `json:"peak_rss_status"`
	PeakRSSBytes   int64    `json:"peak_rss_bytes,omitempty"`
	OutputBytes    int      `json:"output_bytes"`

	// The seven /1 reporting fields (D-06-12, D-06-15, D-06-17, D-06-21).
	// All omitempty: an unaffected /0-era lane's JSON is a superset-by-
	// absence of what it was before this bump. None of these fields may
	// ever reach Result.Finalize()'s identity struct (D-06-32) -- see
	// TestMetricsAndLaneFieldsExcludedFromIdentity.

	// CacheStatus reports what THIS invocation skipped (or didn't) for this
	// lane, drawn from cache.CacheStatuses(). Deliberately NOT hit/miss:
	// those words would wrongly imply a verdict itself was cached, when in
	// fact the checker, five-axis comparator, and sanitizer classifier
	// always re-run fresh regardless of cache status (D-06-06, D-06-12).
	CacheStatus string `json:"cache_status,omitempty"`
	// SelectionReason names why this lane ran (or was deferred/widened) this
	// invocation, drawn from LaneSelectionReasons() (D-06-10, D-06-11).
	SelectionReason string `json:"selection_reason,omitempty"`
	// MachineID is the declared-machine identity this lane's measurements
	// were taken on (D-06-17), never a raw host fingerprint.
	MachineID string `json:"machine_id,omitempty"`
	// GateVerdict is this lane's ratification verdict, drawn from
	// LaneGateVerdicts() (D-06-14, D-06-18, D-06-22).
	GateVerdict string `json:"gate_verdict,omitempty"`
	// ColdOrWarm reports whether this lane's measurement was taken cold or
	// warm (D-06-19); exactly "cold" or "warm".
	ColdOrWarm string `json:"cold_or_warm,omitempty"`
	// StageBreakdown is a per-stage elapsed-time report (D-06-21).
	StageBreakdown []StageTiming `json:"stage_breakdown,omitempty"`
}

// LaneCacheStatuses returns the closed, four-value cache-status vocabulary a
// Lane's CacheStatus may report, drawn directly from cache.CacheStatuses()
// (D-06-12) so this package never maintains a second copy of that
// vocabulary.
func LaneCacheStatuses() []string {
	return cache.CacheStatuses()
}

// The closed SelectionReason vocabulary (D-06-10, D-06-11): a lane was
// selected, deferred, or its selection was widened out of an abundance of
// caution. These three literal words match session.ReasonSelected/
// ReasonDeferred/ReasonWidened's own values; protocol cannot import session
// (session imports protocol), so the values are declared independently here
// and must be kept in sync by inspection.
const (
	LaneSelectionSelected = "selected"
	LaneSelectionDeferred = "deferred"
	LaneSelectionWidened  = "widened"
)

// LaneSelectionReasons returns the closed, three-value selection-reason
// vocabulary a Lane's SelectionReason may report (D-06-10, D-06-11).
func LaneSelectionReasons() []string {
	return []string{LaneSelectionSelected, LaneSelectionDeferred, LaneSelectionWidened}
}

// The closed GateVerdict vocabulary (D-06-14, D-06-18, D-06-22).
const (
	LaneGateBlocking    = "blocking"
	LaneGateObserved    = "observed"
	LaneGateNotRatified = "not_ratified"
)

// LaneGateVerdicts returns the closed, three-value gate-verdict vocabulary a
// Lane's GateVerdict may report.
func LaneGateVerdicts() []string {
	return []string{LaneGateBlocking, LaneGateObserved, LaneGateNotRatified}
}

// The closed ColdOrWarm vocabulary (D-06-19).
const (
	LaneCold = "cold"
	LaneWarm = "warm"
)

// LaneColdOrWarmValues returns the closed, two-value cold/warm vocabulary a
// Lane's ColdOrWarm may report.
func LaneColdOrWarmValues() []string {
	return []string{LaneCold, LaneWarm}
}

// ValidateLaneVocabularies refuses a Lane whose CacheStatus, SelectionReason,
// GateVerdict, or ColdOrWarm holds any value outside its closed vocabulary.
// An empty value in any of these fields is valid (not yet populated is not
// the same as out-of-vocabulary) -- only a genuinely out-of-vocabulary,
// non-empty value is refused. This is what keeps a free-text cache_status
// from silently reintroducing hit/miss semantics (D-06-12) or any other
// field from carrying an unvocabularied value into a published /1 document.
func ValidateLaneVocabularies(lane Lane) error {
	if lane.CacheStatus != "" && !contains(LaneCacheStatuses(), lane.CacheStatus) {
		return fmt.Errorf("protocol: lane %q has out-of-vocabulary cache_status %q", lane.ID, lane.CacheStatus)
	}
	if lane.SelectionReason != "" && !selectionReasonInVocabulary(lane.SelectionReason) {
		return fmt.Errorf("protocol: lane %q has out-of-vocabulary selection_reason %q", lane.ID, lane.SelectionReason)
	}
	if lane.GateVerdict != "" && !contains(LaneGateVerdicts(), lane.GateVerdict) {
		return fmt.Errorf("protocol: lane %q has out-of-vocabulary gate_verdict %q", lane.ID, lane.GateVerdict)
	}
	if lane.ColdOrWarm != "" && !contains(LaneColdOrWarmValues(), lane.ColdOrWarm) {
		return fmt.Errorf("protocol: lane %q has out-of-vocabulary cold_or_warm %q", lane.ID, lane.ColdOrWarm)
	}
	return nil
}

// selectionReasonInVocabulary accepts either a bare vocabulary word
// ("selected", "deferred", "widened") or session.Selection.Reasons's own
// fuller detail form ("selected: pure_match matched", "deferred: no
// declared dependency", "widened: undeclared-input risk") -- the closed
// vocabulary governs the LEADING word, not the full free-text explanation
// that follows it (D-06-10/D-06-11's own reason strings always carry that
// detail; protocol cannot import session to compare against its exact
// constants, so this checks the same three-word prefix independently).
func selectionReasonInVocabulary(value string) bool {
	if contains(LaneSelectionReasons(), value) {
		return true
	}
	for _, vocabulary := range LaneSelectionReasons() {
		if strings.HasPrefix(value, vocabulary+":") {
			return true
		}
	}
	return false
}

func contains(vocabulary []string, value string) bool {
	for _, candidate := range vocabulary {
		if candidate == value {
			return true
		}
	}
	return false
}

type Result struct {
	Schema          string                  `json:"schema"`
	Command         string                  `json:"command"`
	Status          string                  `json:"status"`
	ID              string                  `json:"id"`
	ModuleID        string                  `json:"module_id,omitempty"`
	Formatted       string                  `json:"formatted,omitempty"`
	Diagnostics     []diagnostic.Diagnostic `json:"diagnostics"`
	Executions      []interp.Execution      `json:"executions"`
	Evidence        *EvidenceSummary        `json:"evidence,omitempty"`
	Interface       *InterfaceSummary       `json:"interface,omitempty"`
	DebugMap        *DebugMapSummary        `json:"debug_map,omitempty"`
	Explain         *ExplainSummary         `json:"explain,omitempty"`
	Query           *QuerySummary           `json:"query,omitempty"`
	Lanes           []Lane                  `json:"lanes"`
	ExpectedEscapes []string                `json:"expected_escapes,omitempty"`
	Metrics         Metrics                 `json:"metrics"`
}

func New(command, status string) Result {
	return Result{
		Schema: Schema1, Command: command, Status: status,
		Diagnostics: []diagnostic.Diagnostic{}, Executions: []interp.Execution{}, Lanes: []Lane{},
		Metrics: Metrics{PeakRSSStatus: "unavailable"},
	}
}

func (result Result) Finalize() Result {
	identity := struct {
		Schema           string
		Command          string
		Status           string
		ModuleID         string
		FormattedDigest  string
		DiagnosticIDs    []string
		ExecutionDigests []string
		EvidenceID       string
		InterfaceID      string
		DebugMapID       string
		ExplainID        string
		QueryID          string
		LaneIDs          []string
		ExpectedEscapes  []string `json:",omitempty"`
	}{
		Schema: result.Schema, Command: result.Command, Status: result.Status,
		ModuleID:         result.ModuleID,
		DiagnosticIDs:    make([]string, 0, len(result.Diagnostics)),
		ExecutionDigests: make([]string, 0, len(result.Executions)), LaneIDs: make([]string, 0, len(result.Lanes)),
		ExpectedEscapes: append([]string(nil), result.ExpectedEscapes...),
	}
	if result.Formatted != "" {
		sum := sha256.Sum256([]byte(result.Formatted))
		identity.FormattedDigest = hex.EncodeToString(sum[:])
	}
	for _, problem := range result.Diagnostics {
		identity.DiagnosticIDs = append(identity.DiagnosticIDs, problem.ID)
	}
	for _, execution := range result.Executions {
		encodedExecution, _ := json.Marshal(execution)
		executionSum := sha256.Sum256(encodedExecution)
		identity.ExecutionDigests = append(identity.ExecutionDigests, hex.EncodeToString(executionSum[:]))
	}
	if result.Evidence != nil {
		identity.EvidenceID = result.Evidence.ID
	}
	if result.Interface != nil {
		encodedInterface, _ := json.Marshal(result.Interface)
		interfaceSum := sha256.Sum256(encodedInterface)
		identity.InterfaceID = hex.EncodeToString(interfaceSum[:12])
	}
	if result.DebugMap != nil {
		encodedDebugMap, _ := json.Marshal(result.DebugMap)
		debugMapSum := sha256.Sum256(encodedDebugMap)
		identity.DebugMapID = hex.EncodeToString(debugMapSum[:12])
	}
	if result.Explain != nil {
		encodedExplain, _ := json.Marshal(result.Explain)
		explainSum := sha256.Sum256(encodedExplain)
		identity.ExplainID = hex.EncodeToString(explainSum[:12])
	}
	if result.Query != nil {
		encodedQuery, _ := json.Marshal(result.Query)
		querySum := sha256.Sum256(encodedQuery)
		identity.QueryID = hex.EncodeToString(querySum[:12])
	}
	for _, lane := range result.Lanes {
		identity.LaneIDs = append(identity.LaneIDs, lane.ID+":"+lane.Status)
	}
	encoded, _ := json.Marshal(identity)
	sum := sha256.Sum256(encoded)
	result.ID = "result:" + hex.EncodeToString(sum[:12])
	return result
}

func JSON(result Result) ([]byte, error) {
	result = result.Finalize()
	for attempts := 0; attempts < 4; attempts++ {
		encoded, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		outputBytes := len(encoded) + 1
		if result.Metrics.OutputBytes == outputBytes {
			return append(encoded, '\n'), nil
		}
		result.Metrics.OutputBytes = outputBytes
	}
	return nil, fmt.Errorf("command output size did not converge")
}

// Human renders result's human-readable projection, converging OutputBytes
// against the rendered length exactly like JSON does (D-02-08): both loops
// share the same shape (up to 4 attempts, OutputBytes fed back in), and
// both fail the same way -- an error, never a silently unconverged output
// bytes count -- when a result's projected size does not stabilize within
// that budget. Before this fix, Human unconditionally returned whatever the
// fourth attempt produced even if it had not converged, while JSON reported
// the same condition as an error; a caller comparing the two projections'
// self-reported output_bytes against their own actual lengths could
// silently observe a mismatch from the human side that the JSON side would
// have refused to serve.
func Human(result Result) (string, error) {
	result = result.Finalize()
	for attempts := 0; attempts < 4; attempts++ {
		output := human(result)
		if result.Metrics.OutputBytes == len(output) {
			return output, nil
		}
		result.Metrics.OutputBytes = len(output)
	}
	return "", fmt.Errorf("command output size did not converge")
}

func human(result Result) string {
	var output strings.Builder
	fmt.Fprintf(&output, "%s %s %s", result.ID, result.Command, result.Status)
	if result.ModuleID != "" {
		fmt.Fprintf(&output, " module=%s", result.ModuleID)
	}
	output.WriteByte('\n')
	for _, problem := range result.Diagnostics {
		fmt.Fprintf(&output, "%s %s [%d:%d]: %s\n", problem.ID, problem.Code, problem.Primary.Start, problem.Primary.End, problem.Message)
	}
	for _, execution := range result.Executions {
		for _, event := range execution.Events {
			fmt.Fprintf(&output, "%s %s", event.ID, event.Kind)
			if event.Input != "" || event.Output != "" {
				fmt.Fprintf(&output, " input=%s output=%s", event.Input, event.Output)
			}
			if event.SourcePlace != "" {
				fmt.Fprintf(&output, " source_place=%s", event.SourcePlace)
			}
			if event.TargetPlace != "" {
				fmt.Fprintf(&output, " target_place=%s", event.TargetPlace)
			}
			if event.TypeID != "" {
				fmt.Fprintf(&output, " type_id=%s", event.TypeID)
			}
			output.WriteByte('\n')
		}
	}
	if result.Evidence != nil {
		fmt.Fprintf(&output, "%s %s digest=%s\n", result.Evidence.ID, result.Evidence.Schema, result.Evidence.Digest)
	}
	if result.Interface != nil {
		fmt.Fprintf(&output, "%s module=%s core_digest=%s\n", result.Interface.Schema, result.Interface.ModuleID, result.Interface.CoreDigest)
		for _, function := range result.Interface.Functions {
			fmt.Fprintf(&output, "  %s %s paths=%v access=%s\n", function.ID, function.Name, function.Paths, function.Access)
		}
	}
	if result.DebugMap != nil {
		fmt.Fprintf(&output, "%s entries=%d\n", result.DebugMap.Schema, len(result.DebugMap.Entries))
		for _, entry := range result.DebugMap.Entries {
			fmt.Fprintf(&output, "  %s kind=%s availability=%s\n", entry.ID, entry.Kind, entry.Availability)
		}
	}
	if result.Explain != nil {
		fmt.Fprintf(&output, "%s root=%s nodes=%d edges=%d\n", result.Explain.Schema, result.Explain.RootID, len(result.Explain.Nodes), len(result.Explain.Edges))
		for _, node := range result.Explain.Nodes {
			fmt.Fprintf(&output, "  %s kind=%s availability=%s\n", node.ID, node.Kind, node.Availability)
		}
	}
	if result.Query != nil {
		fmt.Fprintf(&output, "%s query=%s facts=%d\n", result.Query.Schema, result.Query.Query, len(result.Query.Facts))
		for _, fact := range result.Query.Facts {
			fmt.Fprintf(&output, "  %s vocabulary=%s availability=%s\n", fact.ID, fact.Vocabulary, fact.Availability)
		}
	}
	for _, lane := range result.Lanes {
		fmt.Fprintf(&output, "%s %s %s work=%d\n", lane.ID, lane.Schema, lane.Status, lane.RecomputedWork)
	}
	for _, expected := range result.ExpectedEscapes {
		fmt.Fprintf(&output, "%s expected_escape\n", expected)
	}
	fmt.Fprintf(&output, "metrics elapsed_ns=%d peak_rss=%s output_bytes=%d recomputed_work=%d\n", result.Metrics.ElapsedNS, result.Metrics.PeakRSSStatus, result.Metrics.OutputBytes, result.Metrics.RecomputedWork)
	return output.String()
}

func ExitCode(status string) int {
	switch status {
	case StatusPass:
		return 0
	case StatusInvalid:
		return 2
	case StatusOperational:
		return 3
	case StatusMismatch:
		return 4
	default:
		return 64
	}
}
