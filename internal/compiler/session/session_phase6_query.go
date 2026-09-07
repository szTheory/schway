package session

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// The five stable ID vocabularies D-06-01 joins, plus the "unrecognized"
// sentinel a genuinely unmatched address reports (never a sixth real
// vocabulary -- see QueryVocabularies/TestQueryMintsNoSixthVocabulary).
const (
	QueryVocabularyDiagnostic   = "diagnostic"
	QueryVocabularyDebugMap     = "debugmap"
	QueryVocabularyEvidence     = "evidence"
	QueryVocabularyControl      = "control"
	QueryVocabularyLane         = "lane"
	QueryVocabularyUnrecognized = "unrecognized"
)

// QueryVocabularies returns the closed set of five vocabulary names the
// join dispatcher recognizes (D-06-01). It is built from this file's own
// dispatch table rather than a second hand-copied literal, so the two
// cannot independently drift -- TestQueryMintsNoSixthVocabulary asserts the
// dispatch table's key set equals exactly this slice.
func QueryVocabularies() []string {
	names := make([]string, 0, len(queryVocabularyDispatch))
	for name := range queryVocabularyDispatch {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// QueryKinds returns D-06-05's closed --kind vocabulary.
func QueryKinds() []string {
	return []string{"symbol", "type", "ownership", "dependency", "test"}
}

var queryKindVocabulary = map[string]bool{
	"symbol": true, "type": true, "ownership": true, "dependency": true, "test": true,
}

// queryVocabularyResolver resolves address against SRC (path) for one
// vocabulary category, returning every fact it found (at least one honest
// fact -- available or not_captured -- for every recognized address) plus
// the work it recomputed.
type queryVocabularyResolver func(path, address string) ([]protocol.QueryFact, int, error)

// queryVocabularyDispatch is the single source of truth for "which of the
// five vocabularies handles this address" -- classifyQueryVocabulary picks
// a key from this map's key set, and TestQueryMintsNoSixthVocabulary proves
// the two never drift apart.
var queryVocabularyDispatch = map[string]queryVocabularyResolver{
	QueryVocabularyDiagnostic: resolveQueryDiagnosticFacts,
	QueryVocabularyDebugMap:   resolveQueryDebugMapFacts,
	QueryVocabularyEvidence:   resolveQueryEvidenceFacts,
	QueryVocabularyControl: func(_, address string) ([]protocol.QueryFact, int, error) {
		facts, work := resolveQueryControlFacts(address)
		return facts, work, nil
	},
	QueryVocabularyLane: func(path, address string) ([]protocol.QueryFact, int, error) {
		facts, work := resolveQueryLaneFacts(path, address)
		return facts, work, nil
	},
}

// classifyQueryVocabulary recognizes an address's vocabulary by prefix or
// shape alone (D-06-01): "diagnostic:", "control:", "lane:" are unambiguous
// prefixes; "evidence:" or a bare 64-hex content digest identify the
// evidence vocabulary; everything else is handed to the debugmap resolver,
// which itself decides among core_id/operation_id/point_id or reports the
// address as genuinely unrecognized (see resolveQueryDebugMapFacts).
func classifyQueryVocabulary(address string) string {
	switch {
	case strings.HasPrefix(address, "diagnostic:"):
		return QueryVocabularyDiagnostic
	case strings.HasPrefix(address, "control:"):
		return QueryVocabularyControl
	case strings.HasPrefix(address, "lane:"):
		return QueryVocabularyLane
	case strings.HasPrefix(address, "evidence:") || isQueryHexDigest(address):
		return QueryVocabularyEvidence
	default:
		return QueryVocabularyDebugMap
	}
}

func isQueryHexDigest(address string) bool {
	if len(address) != 64 {
		return false
	}
	for _, r := range address {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}

// QueryOptions carries `lang query`'s optional flags (D-06-05). Depth is
// captured but does not affect pagination (D-06-03: query pages by cursor,
// never by depth) -- it is reserved for a future join-traversal-depth
// consumer; none of the five vocabularies this plan resolves needs more
// than one join hop, so it is currently inert.
type QueryOptions struct {
	Kind   string
	Cursor string
	Depth  int
}

// QueryCommandFile is the CLI seam for `lang query SRC ID_OR_PATTERN
// [--kind=K] [--cursor=C] [--json]` (D-06-01/D-06-05): it re-derives every
// fact from SRC on this cold invocation alone (D-06-02's no-persisted-store
// discipline extends to query, matching explain/debug-map), routes address
// through the closed five-vocabulary dispatcher, optionally filters by
// --kind, and returns a bounded, cursor-paginated page under lang.query/0.
func QueryCommandFile(path, address string, options QueryOptions) (protocol.Result, error) {
	started := time.Now()
	result := protocol.New("query", protocol.StatusPass)

	if options.Kind != "" && !queryKindVocabulary[options.Kind] {
		result.Status = protocol.StatusUsage
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("tool.query_unknown_kind", diagnostic.Span{}, "unrecognized --kind value")}
		return completeCommand(result, started, 1), nil
	}

	category := classifyQueryVocabulary(address)
	resolver, ok := queryVocabularyDispatch[category]
	var facts []protocol.QueryFact
	var work int
	var err error
	if ok {
		facts, work, err = resolver(path, address)
	} else {
		// Structurally unreachable: classifyQueryVocabulary only ever
		// returns keys present in queryVocabularyDispatch. Kept as an
		// explicit fail-closed branch rather than a silent fallthrough.
		facts, work = []protocol.QueryFact{unrecognizedQueryFact(address)}, 1
	}
	if err != nil {
		return protocol.Result{}, err
	}

	if options.Kind != "" {
		facts = filterQueryFactsByKind(facts, options.Kind)
	}
	sortQueryFacts(facts)

	page, nextCursor, truncated, cursorErr := paginateQueryFacts(facts, options.Cursor)
	if cursorErr != nil {
		result.Status = protocol.StatusUsage
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error("tool.query_malformed_cursor", diagnostic.Span{}, "cursor does not correspond to a valid page boundary")}
		return completeCommand(result, started, work), nil
	}

	result.Query = &protocol.QuerySummary{
		Schema: protocol.QuerySchema, Query: address, Facts: page, NextCursor: nextCursor, Truncated: truncated,
	}
	return completeCommand(result, started, work+len(page)), nil
}

func unrecognizedQueryFact(address string) protocol.QueryFact {
	return protocol.QueryFact{ID: address, Kind: "unrecognized", Vocabulary: QueryVocabularyUnrecognized, Availability: string(debugmap.NotCaptured)}
}

// --- diagnostic vocabulary ------------------------------------------------

// resolveQueryDiagnosticFacts re-derives SRC's diagnostics exactly like
// `explain` and flattens the requested diagnostic (root fact) plus its flat
// Causes list into QueryFacts -- no graph, no edges (that is `explain`'s
// job); query answers "what facts exist at this address," not "how are they
// related."
func resolveQueryDiagnosticFacts(path, address string) ([]protocol.QueryFact, int, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return nil, 0, err
	}
	parsed := syntax.Parse(source)
	diagnostics := parsed.Diagnostics
	work := 1
	if len(diagnostics) == 0 {
		checked := check.Program(parsed.Program)
		diagnostics = checked.Diagnostics
		work = checked.Work
	}
	target, found := findExplainDiagnostic(diagnostics, address)
	if !found {
		return []protocol.QueryFact{{ID: address, Kind: "diagnostic", Vocabulary: QueryVocabularyDiagnostic, Availability: string(debugmap.NotCaptured)}}, work, nil
	}
	primary := target.Primary
	facts := []protocol.QueryFact{{
		ID: target.ID, Kind: target.Code, Detail: target.Message, Span: &primary,
		Vocabulary: QueryVocabularyDiagnostic, Availability: string(debugmap.Available),
	}}
	for index, cause := range target.Causes {
		availability := string(debugmap.Available)
		if cause.Span == nil {
			availability = string(debugmap.NotCaptured)
		}
		facts = append(facts, protocol.QueryFact{
			ID: fmt.Sprintf("%s:cause:%d", target.ID, index), Kind: cause.Kind, Detail: cause.Detail,
			Span: cause.Span, Vocabulary: QueryVocabularyDiagnostic, Availability: availability,
		})
	}
	return facts, work + len(target.Causes), nil
}

// --- debugmap vocabulary (core_id / operation_id / point_id) -------------

// resolveQueryDebugMapFacts is the catch-all branch: it builds SRC's debug
// map exactly like `debug-map` and then decides, by address shape, which of
// the three debugmap-joined ID fields to resolve against. An address
// containing ":op:" or ":point:" is always routed here (even when absent,
// it is an honestly not_captured operation_id/point_id fact, matching
// debug-map's own honest-absence report); a bare address is checked against
// every entry's CoreID, and only reported as genuinely QueryVocabularyUnrecognized
// when it matches neither shape nor any real core ID -- core_id has no
// stable prefix of its own to recognize by shape alone.
func resolveQueryDebugMapFacts(path, address string) ([]protocol.QueryFact, int, error) {
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		return nil, 0, err
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		return []protocol.QueryFact{unrecognizedQueryFact(address)}, 1, nil
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		return []protocol.QueryFact{unrecognizedQueryFact(address)}, checked.Work, nil
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		return nil, 0, fmt.Errorf("core validation failed: %s", validated.Problems[0].Code)
	}
	program := validated.Program()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	built, work, buildErr := debugmap.Build(ctx, parsed.Program, program)
	if buildErr != nil {
		return nil, work, buildErr
	}

	switch {
	case strings.Contains(address, ":op:"):
		entry := debugmap.Resolve(built, address)
		return []protocol.QueryFact{debugMapEntryToFact(entry, "operation_id", address)}, work + 1, nil
	case strings.Contains(address, ":point:"):
		entry := resolveDebugMapByPointID(built, address)
		return []protocol.QueryFact{debugMapEntryToFact(entry, "point_id", address)}, work + 1, nil
	default:
		entry, found := resolveDebugMapByCoreID(built, address)
		if !found {
			return []protocol.QueryFact{unrecognizedQueryFact(address)}, work, nil
		}
		return []protocol.QueryFact{debugMapEntryToFact(entry, "core_id", address)}, work + 1, nil
	}
}

// resolveDebugMapByPointID and resolveDebugMapByCoreID extend
// debugmap.Resolve's single-field (OperationID) lookup shape to the two
// other joined fields ENTIRELY in this session-layer file -- debugmap.Resolve
// itself is not modified, per the plan's explicit instruction.
func resolveDebugMapByPointID(built debugmap.Map, pointID string) debugmap.Entry {
	for _, entry := range built.Entries {
		if entry.PointID == pointID {
			return entry
		}
	}
	return debugmap.Entry{PointID: pointID, Availability: debugmap.NotCaptured}
}

func resolveDebugMapByCoreID(built debugmap.Map, coreID string) (debugmap.Entry, bool) {
	for _, entry := range built.Entries {
		if entry.CoreID == coreID {
			return entry, true
		}
	}
	return debugmap.Entry{}, false
}

func debugMapEntryToFact(entry debugmap.Entry, vocabulary, address string) protocol.QueryFact {
	id := entry.ID
	if id == "" {
		id = address
	}
	var span *diagnostic.Span
	if entry.Availability == debugmap.Available {
		spanCopy := entry.SourceSpan
		span = &spanCopy
	}
	return protocol.QueryFact{ID: id, Kind: entry.Kind, Vocabulary: vocabulary, Span: span, Availability: string(entry.Availability)}
}

// --- control vocabulary ---------------------------------------------------

// resolveQueryControlFacts answers against AllShippedControlIDs() -- the
// same live-derived set QLT-01's registry audit already cross-checks
// against -- never a hand-copied literal (D-06-01).
func resolveQueryControlFacts(address string) ([]protocol.QueryFact, int) {
	for _, control := range AllShippedControlIDs() {
		if control == address {
			return []protocol.QueryFact{{ID: control, Kind: "control", Vocabulary: QueryVocabularyControl, Availability: string(debugmap.Available)}}, 1
		}
	}
	return []protocol.QueryFact{{ID: address, Kind: "control", Vocabulary: QueryVocabularyControl, Availability: string(debugmap.NotCaptured)}}, 1
}

// --- lane vocabulary -------------------------------------------------------

// resolveQueryLaneFacts answers against the live lane-ID set a real `verify`
// run over path produces (D-06-01) -- never a hand-copied literal. path is
// therefore expected to be a corpus directory in this vocabulary, exactly
// like `lang verify CORPUS`'s own SRC operand.
func resolveQueryLaneFacts(path, address string) ([]protocol.QueryFact, int) {
	result := VerifyCorpusFile(context.Background(), path, native.DefaultRunner())
	work := result.Metrics.RecomputedWork + 1
	for _, lane := range result.Lanes {
		if lane.ID == address {
			return []protocol.QueryFact{{ID: lane.ID, Kind: "lane", Vocabulary: QueryVocabularyLane, Availability: string(debugmap.Available)}}, work
		}
	}
	return []protocol.QueryFact{{ID: address, Kind: "lane", Vocabulary: QueryVocabularyLane, Availability: string(debugmap.NotCaptured)}}, work
}

// --- evidence vocabulary ---------------------------------------------------

// resolveQueryEvidenceFacts reuses EvidenceCommandFile (the already-shipped
// `lang evidence` seam) rather than re-deriving the manifest a second way,
// and matches address against either the manifest's own ID or its content
// digest (D-06-01: "evidence IDs and content digests").
func resolveQueryEvidenceFacts(path, address string) ([]protocol.QueryFact, int, error) {
	_, result, err := EvidenceCommandFile(context.Background(), path)
	if err != nil {
		return nil, 0, err
	}
	work := result.Metrics.RecomputedWork
	if result.Status != protocol.StatusPass || result.Evidence == nil {
		return []protocol.QueryFact{{ID: address, Kind: "evidence", Vocabulary: QueryVocabularyEvidence, Availability: string(debugmap.NotCaptured)}}, work, nil
	}
	digest := strings.TrimPrefix(result.Evidence.Digest, "sha256:")
	if result.Evidence.ID == address || result.Evidence.Digest == address || digest == address {
		return []protocol.QueryFact{{
			ID: result.Evidence.ID, Kind: "evidence", Detail: result.Evidence.Digest,
			Vocabulary: QueryVocabularyEvidence, Availability: string(debugmap.Available),
		}}, work, nil
	}
	return []protocol.QueryFact{{ID: address, Kind: "evidence", Vocabulary: QueryVocabularyEvidence, Availability: string(debugmap.NotCaptured)}}, work, nil
}

// --- --kind filtering (D-06-05) -------------------------------------------

// queryOwnershipKinds are the diagnostic.Cause Kind values explainCorrelationKinds
// already recognizes as binding/place identities, plus the debugmap
// operation-kind vocabulary kindFor produces -- both genuinely describe
// ownership facts.
var queryOwnershipKinds = map[string]bool{
	"place": true, "owner": true, "loan": true, "transfer_target": true,
	"move": true, "borrow_shared": true, "borrow_exclusive": true, "copy": true,
}

func isQueryDebugMapVocabulary(vocabulary string) bool {
	switch vocabulary {
	case "core_id", "operation_id", "point_id":
		return true
	}
	return false
}

// queryFactMatchesKind maps a QueryFact onto D-06-05's five declared --kind
// values. This mapping is Claude's Discretion (the grammar names the five
// kinds but not their semantics): "type" reuses the real diagnostic.Cause
// Kind value check.go already emits for type-mismatch causes; "ownership"
// reuses the same binding-identity Kind values explain's own correlation
// table recognizes, widened to the debugmap operation-kind vocabulary;
// "dependency" is the source-to-core join facts (debugmap's three ID
// fields); "test" is verify's lane facts; "symbol" is a resolved evidence
// identity (a compiled build artifact's own name).
func queryFactMatchesKind(fact protocol.QueryFact, kind string) bool {
	switch kind {
	case "symbol":
		return fact.Vocabulary == QueryVocabularyEvidence
	case "type":
		return fact.Kind == "type"
	case "ownership":
		return queryOwnershipKinds[fact.Kind]
	case "dependency":
		return isQueryDebugMapVocabulary(fact.Vocabulary)
	case "test":
		return fact.Vocabulary == QueryVocabularyLane
	default:
		return false
	}
}

func filterQueryFactsByKind(facts []protocol.QueryFact, kind string) []protocol.QueryFact {
	filtered := make([]protocol.QueryFact, 0, len(facts))
	for _, fact := range facts {
		if queryFactMatchesKind(fact, kind) {
			filtered = append(filtered, fact)
		}
	}
	return filtered
}

// --- sorting and cursor pagination (D-06-03) ------------------------------

// sortQueryFacts orders facts by (vocabulary, span.start, span.end, id) so
// equal-comparing facts have one specified, stable order across cold
// invocations and across page boundaries (FND-04 ordering edge) -- the same
// discipline buildExplainGraph's own final sort already proves out.
func sortQueryFacts(facts []protocol.QueryFact) {
	sort.SliceStable(facts, func(left, right int) bool {
		if facts[left].Vocabulary != facts[right].Vocabulary {
			return facts[left].Vocabulary < facts[right].Vocabulary
		}
		leftStart, leftEnd := querySpanBounds(facts[left].Span)
		rightStart, rightEnd := querySpanBounds(facts[right].Span)
		if leftStart != rightStart {
			return leftStart < rightStart
		}
		if leftEnd != rightEnd {
			return leftEnd < rightEnd
		}
		return facts[left].ID < facts[right].ID
	})
}

func querySpanBounds(span *diagnostic.Span) (int, int) {
	if span == nil {
		return 0, 0
	}
	return span.Start, span.End
}

// queryCursorKey is the sort key a cursor is content-derived over -- the
// LAST fact emitted on the previous page, never a raw offset (D-06-03): an
// unstable underlying order cannot silently skip or duplicate a position
// expressed this way, only a genuinely reordered corpus can.
type queryCursorKey struct {
	Vocabulary string `json:"vocabulary"`
	Start      int    `json:"start"`
	End        int    `json:"end"`
	ID         string `json:"id"`
}

// ErrQueryMalformedCursor is returned by decodeQueryCursor for any cursor
// that does not decode to a valid, checksum-matching key -- a usage error,
// never a silent first-page fallback (D-06-03).
var ErrQueryMalformedCursor = errors.New("query.malformed_cursor")

func encodeQueryCursor(fact protocol.QueryFact) string {
	start, end := querySpanBounds(fact.Span)
	key := queryCursorKey{Vocabulary: fact.Vocabulary, Start: start, End: end, ID: fact.ID}
	payload, _ := json.Marshal(key)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:6]) + "." + base64.RawURLEncoding.EncodeToString(payload)
}

func decodeQueryCursor(cursor string) (queryCursorKey, error) {
	parts := strings.SplitN(cursor, ".", 2)
	if len(parts) != 2 {
		return queryCursorKey{}, ErrQueryMalformedCursor
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return queryCursorKey{}, ErrQueryMalformedCursor
	}
	sum := sha256.Sum256(payload)
	if hex.EncodeToString(sum[:6]) != parts[0] {
		return queryCursorKey{}, ErrQueryMalformedCursor
	}
	var key queryCursorKey
	if err := json.Unmarshal(payload, &key); err != nil {
		return queryCursorKey{}, ErrQueryMalformedCursor
	}
	return key, nil
}

// queryFactAfterKey reports whether fact sorts strictly after key in
// sortQueryFacts's own ordering -- the predicate paginateQueryFacts binary
// searches on to find the first fact of the next page.
func queryFactAfterKey(fact protocol.QueryFact, key queryCursorKey) bool {
	if fact.Vocabulary != key.Vocabulary {
		return fact.Vocabulary > key.Vocabulary
	}
	start, end := querySpanBounds(fact.Span)
	if start != key.Start {
		return start > key.Start
	}
	if end != key.End {
		return end > key.End
	}
	return fact.ID > key.ID
}

// paginateQueryFacts pages sortQueryFacts-ordered facts by an opaque
// content-derived cursor (D-06-03). A malformed cursor is a usage error
// (non-nil error return), never a silent page one. Concatenating every page
// this function returns, driven by each page's own NextCursor in turn,
// reproduces facts exactly once with no gaps and no duplicates
// (TestQueryPagesConcatenateExactlyOnce).
func paginateQueryFacts(facts []protocol.QueryFact, cursor string) ([]protocol.QueryFact, string, string, error) {
	start := 0
	if cursor != "" {
		key, err := decodeQueryCursor(cursor)
		if err != nil {
			return nil, "", "", err
		}
		start = sort.Search(len(facts), func(index int) bool {
			return queryFactAfterKey(facts[index], key)
		})
	}
	end := start + protocol.QueryMaxFactsPerPage
	truncated := end < len(facts)
	if !truncated {
		end = len(facts)
	}
	page := append([]protocol.QueryFact(nil), facts[start:end]...)
	nextCursor := ""
	truncatedCode := ""
	if truncated && len(page) > 0 {
		nextCursor = encodeQueryCursor(page[len(page)-1])
		truncatedCode = "truncated:query.page_bound"
	}
	return page, nextCursor, truncatedCode, nil
}
