package session

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/check"
	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/debugmap"
	"github.com/codename-lang/lang/internal/compiler/native"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/syntax"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// This file is deliberately `package session` (internal), matching
// session_phase6_explain_test.go's own established precedent: Task 3's
// pagination/sort tests need to construct exact synthetic QueryFact shapes
// (many facts sharing sort keys, engineered tie cases) that no real fixture
// happens to produce, and the sort/pagination helpers are intentionally
// unexported. Schema minting (Task 1) and the every-real-vocabulary sweep
// (Task 2) still drive QueryCommandFile end-to-end against real fixtures.

// --- Task 1: schema minting + one vocabulary end-to-end -------------------

func TestQuerySummarySchemaIsMinted(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	operationID := firstDebugMapOperationID(t, path)

	result, err := QueryCommandFile(path, operationID, QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if result.Query == nil {
		t.Fatalf("result.Query is nil")
	}
	if result.Query.Schema != protocol.QuerySchema {
		t.Fatalf("schema = %q, want %q", result.Query.Schema, protocol.QuerySchema)
	}
	if len(result.Query.Facts) != 1 {
		t.Fatalf("Facts = %+v, want exactly one resolved fact", result.Query.Facts)
	}
	if result.Query.Facts[0].Availability != string(debugmap.Available) {
		t.Fatalf("Facts[0].Availability = %q, want available", result.Query.Facts[0].Availability)
	}
	if result.Query.Facts[0].Vocabulary != "operation_id" {
		t.Fatalf("Facts[0].Vocabulary = %q, want operation_id", result.Query.Facts[0].Vocabulary)
	}

	// Two Results differing only in Query must have different Result.ID
	// (Finalize folds QueryID into identity).
	finalized := result.Finalize()
	withoutQuery := result
	withoutQuery.Query = nil
	if finalized.ID == withoutQuery.Finalize().ID {
		t.Fatalf("Query did not perturb Result.ID: %s", finalized.ID)
	}
	if finalized.Command != "query" {
		t.Fatalf("Command = %q, want query", finalized.Command)
	}
}

// TestQueryUnknownIDReportsNotCaptured proves the honest-absence contract
// (FND-04): an operation ID the map never produced is one not_captured
// fact, never an error and never a fabricated span.
func TestQueryUnknownIDReportsNotCaptured(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	result, err := QueryCommandFile(path, "s1:owned.transfer:fn:relay:op:999999", QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if result.Status != protocol.StatusPass {
		t.Fatalf("Status = %q, want pass (absence is honest data, not an error)", result.Status)
	}
	if len(result.Query.Facts) != 1 {
		t.Fatalf("Facts = %+v, want exactly one fact", result.Query.Facts)
	}
	fact := result.Query.Facts[0]
	if fact.Availability != string(debugmap.NotCaptured) {
		t.Fatalf("Availability = %q, want not_captured", fact.Availability)
	}
	if fact.Vocabulary != "operation_id" {
		t.Fatalf("Vocabulary = %q, want operation_id (recognized shape, absent value)", fact.Vocabulary)
	}
	if fact.Span != nil {
		t.Fatalf("Span = %+v, want nil for an absent fact (never a fabricated span)", fact.Span)
	}

	// A genuinely unrecognized address (no vocabulary shape at all) reports
	// the "unrecognized" sentinel, never an error.
	unrecognized, err := QueryCommandFile(path, "totally-unrecognized-address", QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile (unrecognized): %v", err)
	}
	if unrecognized.Status != protocol.StatusPass {
		t.Fatalf("unrecognized Status = %q, want pass", unrecognized.Status)
	}
	if len(unrecognized.Query.Facts) != 1 || unrecognized.Query.Facts[0].Vocabulary != QueryVocabularyUnrecognized {
		t.Fatalf("unrecognized Facts = %+v, want one fact with vocabulary=unrecognized", unrecognized.Query.Facts)
	}
	if unrecognized.Query.Facts[0].Availability != string(debugmap.NotCaptured) {
		t.Fatalf("unrecognized Availability = %q, want not_captured", unrecognized.Query.Facts[0].Availability)
	}
}

// --- Task 2: the five-vocabulary join dispatcher ---------------------------

func TestQueryResolvesEveryStableIDVocabulary(t *testing.T) {
	t.Run("diagnostic", func(t *testing.T) {
		path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
		diagID := firstDiagnosticID(t, path)
		result, err := QueryCommandFile(path, diagID, QueryOptions{})
		if err != nil {
			t.Fatalf("QueryCommandFile: %v", err)
		}
		if len(result.Query.Facts) == 0 {
			t.Fatalf("no facts resolved for a real diagnostic ID")
		}
		if result.Query.Facts[0].Vocabulary != QueryVocabularyDiagnostic {
			t.Fatalf("Vocabulary = %q, want diagnostic", result.Query.Facts[0].Vocabulary)
		}
	})

	t.Run("debugmap operation_id/core_id/point_id", func(t *testing.T) {
		path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
		built := buildDebugMap(t, path)
		if len(built.Entries) == 0 {
			t.Fatalf("debugmap.Build produced no entries for %s", path)
		}
		entry := built.Entries[0]

		for _, testCase := range []struct {
			name       string
			address    string
			vocabulary string
		}{
			{"operation_id", entry.OperationID, "operation_id"},
			{"core_id", entry.CoreID, "core_id"},
			{"point_id", entry.PointID, "point_id"},
		} {
			t.Run(testCase.name, func(t *testing.T) {
				result, err := QueryCommandFile(path, testCase.address, QueryOptions{})
				if err != nil {
					t.Fatalf("QueryCommandFile: %v", err)
				}
				if len(result.Query.Facts) != 1 {
					t.Fatalf("Facts = %+v, want exactly one", result.Query.Facts)
				}
				if result.Query.Facts[0].Vocabulary != testCase.vocabulary {
					t.Fatalf("Vocabulary = %q, want %q", result.Query.Facts[0].Vocabulary, testCase.vocabulary)
				}
				if result.Query.Facts[0].Availability != string(debugmap.Available) {
					t.Fatalf("Availability = %q, want available", result.Query.Facts[0].Availability)
				}
			})
		}
	})

	t.Run("evidence", func(t *testing.T) {
		path := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
		_, evidenceResult, err := EvidenceCommandFile(context.Background(), path)
		if err != nil {
			t.Fatalf("EvidenceCommandFile: %v", err)
		}
		if evidenceResult.Evidence == nil {
			t.Fatalf("no evidence manifest produced for %s", path)
		}
		result, err := QueryCommandFile(path, evidenceResult.Evidence.ID, QueryOptions{})
		if err != nil {
			t.Fatalf("QueryCommandFile: %v", err)
		}
		if len(result.Query.Facts) != 1 || result.Query.Facts[0].Vocabulary != QueryVocabularyEvidence {
			t.Fatalf("Facts = %+v, want one evidence fact", result.Query.Facts)
		}
		if result.Query.Facts[0].Availability != string(debugmap.Available) {
			t.Fatalf("Availability = %q, want available", result.Query.Facts[0].Availability)
		}
	})

	t.Run("control", func(t *testing.T) {
		controls := AllShippedControlIDs()
		if len(controls) == 0 {
			t.Fatalf("AllShippedControlIDs() is empty")
		}
		result, err := QueryCommandFile("", controls[0], QueryOptions{})
		if err != nil {
			t.Fatalf("QueryCommandFile: %v", err)
		}
		if len(result.Query.Facts) != 1 || result.Query.Facts[0].Vocabulary != QueryVocabularyControl {
			t.Fatalf("Facts = %+v, want one control fact", result.Query.Facts)
		}
		if result.Query.Facts[0].Availability != string(debugmap.Available) {
			t.Fatalf("Availability = %q, want available", result.Query.Facts[0].Availability)
		}
	})

	t.Run("lane", func(t *testing.T) {
		corpus := testsupport.ProjectPath("testdata", "phase1")
		verifyResult := VerifyCorpusFile(context.Background(), corpus, native.DefaultRunner())
		if len(verifyResult.Lanes) == 0 {
			t.Fatalf("verify produced no lanes for %s", corpus)
		}
		laneID := verifyResult.Lanes[0].ID
		result, err := QueryCommandFile(corpus, laneID, QueryOptions{})
		if err != nil {
			t.Fatalf("QueryCommandFile: %v", err)
		}
		if len(result.Query.Facts) != 1 || result.Query.Facts[0].Vocabulary != QueryVocabularyLane {
			t.Fatalf("Facts = %+v, want one lane fact", result.Query.Facts)
		}
		if result.Query.Facts[0].Availability != string(debugmap.Available) {
			t.Fatalf("Availability = %q, want available", result.Query.Facts[0].Availability)
		}
	})
}

// TestQueryMintsNoSixthVocabulary pins D-06-01's closed vocabulary set: the
// dispatch table this file's classifier routes through has exactly five
// entries, each one derivable from a shipped constant, and the two can
// never independently drift because QueryVocabularies() is read directly
// off the same map classifyQueryVocabulary/QueryCommandFile dispatch
// through.
func TestQueryMintsNoSixthVocabulary(t *testing.T) {
	vocabularies := QueryVocabularies()
	if len(vocabularies) != 5 {
		t.Fatalf("QueryVocabularies() = %v, want exactly 5 entries", vocabularies)
	}
	want := map[string]bool{
		QueryVocabularyDiagnostic: true, QueryVocabularyDebugMap: true, QueryVocabularyEvidence: true,
		QueryVocabularyControl: true, QueryVocabularyLane: true,
	}
	for _, vocabulary := range vocabularies {
		if !want[vocabulary] {
			t.Fatalf("unexpected vocabulary %q outside the closed D-06-01 set", vocabulary)
		}
		if _, ok := queryVocabularyDispatch[vocabulary]; !ok {
			t.Fatalf("vocabulary %q has no matching dispatcher branch", vocabulary)
		}
	}
	// Every dispatch-table entry must also be named by QueryVocabularies() --
	// an entry present in the map but absent from the slice would itself be
	// an undetected sixth vocabulary.
	seen := map[string]bool{}
	for _, vocabulary := range vocabularies {
		seen[vocabulary] = true
	}
	for name := range queryVocabularyDispatch {
		if !seen[name] {
			t.Fatalf("dispatch table entry %q is not named by QueryVocabularies() -- unrouted/undeclared vocabulary", name)
		}
	}
}

// TestQueryKindFilterVocabularyIsClosed covers Test 7: --kind filters the
// fact list, and an unrecognized kind is a usage error, never a silent
// no-op.
func TestQueryKindFilterVocabularyIsClosed(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	diagID := firstDiagnosticID(t, path)

	unfiltered, err := QueryCommandFile(path, diagID, QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if len(unfiltered.Query.Facts) < 2 {
		t.Fatalf("fixture produced too few facts (%d) to prove filtering narrows the list", len(unfiltered.Query.Facts))
	}

	filtered, err := QueryCommandFile(path, diagID, QueryOptions{Kind: "ownership"})
	if err != nil {
		t.Fatalf("QueryCommandFile (kind=ownership): %v", err)
	}
	if len(filtered.Query.Facts) == 0 {
		t.Fatalf("--kind=ownership matched nothing on a fixture with place/transfer_target causes")
	}
	if len(filtered.Query.Facts) >= len(unfiltered.Query.Facts) {
		t.Fatalf("--kind=ownership did not narrow the fact list: filtered=%d unfiltered=%d", len(filtered.Query.Facts), len(unfiltered.Query.Facts))
	}
	for _, fact := range filtered.Query.Facts {
		if !queryOwnershipKinds[fact.Kind] {
			t.Fatalf("fact %+v leaked through --kind=ownership filter", fact)
		}
	}

	for _, kind := range QueryKinds() {
		result, err := QueryCommandFile(path, diagID, QueryOptions{Kind: kind})
		if err != nil {
			t.Fatalf("QueryCommandFile (kind=%s): %v", kind, err)
		}
		if result.Status == protocol.StatusUsage {
			t.Fatalf("kind=%s (a declared valid kind) was rejected as a usage error", kind)
		}
	}

	unknownKind, err := QueryCommandFile(path, diagID, QueryOptions{Kind: "nonsense"})
	if err != nil {
		t.Fatalf("QueryCommandFile (kind=nonsense): %v", err)
	}
	if unknownKind.Status != protocol.StatusUsage {
		t.Fatalf("Status = %q, want usage_error for an unrecognized --kind", unknownKind.Status)
	}
	if len(unknownKind.Diagnostics) != 1 || unknownKind.Diagnostics[0].Code != "tool.query_unknown_kind" {
		t.Fatalf("Diagnostics = %+v, want one entry with code tool.query_unknown_kind", unknownKind.Diagnostics)
	}
}

// --- Task 3: cursor pagination, bounding, stable order on ties ------------

func syntheticQueryFacts(count int) []protocol.QueryFact {
	facts := make([]protocol.QueryFact, count)
	for index := range facts {
		facts[index] = protocol.QueryFact{
			ID: fmt.Sprintf("synthetic:%03d", index), Kind: "synthetic",
			Vocabulary: "control", Availability: string(debugmap.Available),
		}
	}
	return facts
}

func TestQueryCursorPaginationIsBounded(t *testing.T) {
	facts := syntheticQueryFacts(protocol.QueryMaxFactsPerPage + 10)
	sortQueryFacts(facts)

	page, nextCursor, truncated, err := paginateQueryFacts(facts, "")
	if err != nil {
		t.Fatalf("paginateQueryFacts: %v", err)
	}
	if len(page) != protocol.QueryMaxFactsPerPage {
		t.Fatalf("len(page) = %d, want exactly QueryMaxFactsPerPage (%d)", len(page), protocol.QueryMaxFactsPerPage)
	}
	if nextCursor == "" {
		t.Fatalf("nextCursor is empty, want a cursor for the remaining facts")
	}
	if truncated != "truncated:query.page_bound" {
		t.Fatalf("truncated = %q, want truncated:query.page_bound", truncated)
	}
}

func TestQueryPagesConcatenateExactlyOnce(t *testing.T) {
	facts := syntheticQueryFacts(protocol.QueryMaxFactsPerPage*3 + 7)
	sortQueryFacts(facts)

	seen := map[string]bool{}
	var walked []protocol.QueryFact
	cursor := ""
	for pages := 0; pages < 20; pages++ {
		page, nextCursor, _, err := paginateQueryFacts(facts, cursor)
		if err != nil {
			t.Fatalf("paginateQueryFacts: %v", err)
		}
		for _, fact := range page {
			if seen[fact.ID] {
				t.Fatalf("fact %s emitted on more than one page", fact.ID)
			}
			seen[fact.ID] = true
			walked = append(walked, fact)
		}
		if nextCursor == "" {
			break
		}
		cursor = nextCursor
	}
	if len(walked) != len(facts) {
		t.Fatalf("walked %d facts across all pages, want exactly %d (no gaps)", len(walked), len(facts))
	}
	for index, fact := range walked {
		if fact.ID != facts[index].ID {
			t.Fatalf("concatenated page order diverges from the unpaginated list at index %d: got %s want %s", index, fact.ID, facts[index].ID)
		}
	}
}

func TestQueryMalformedCursorIsUsageError(t *testing.T) {
	facts := syntheticQueryFacts(5)
	sortQueryFacts(facts)

	if _, _, _, err := paginateQueryFacts(facts, "not-a-real-cursor"); err == nil {
		t.Fatalf("paginateQueryFacts accepted a garbage cursor without error")
	}

	path := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	operationID := firstDebugMapOperationID(t, path)
	result, err := QueryCommandFile(path, operationID, QueryOptions{Cursor: "garbage-cursor-value"})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	if result.Status != protocol.StatusUsage {
		t.Fatalf("Status = %q, want usage_error for a malformed cursor", result.Status)
	}
	if len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "tool.query_malformed_cursor" {
		t.Fatalf("Diagnostics = %+v, want one entry with code tool.query_malformed_cursor", result.Diagnostics)
	}
}

// TestQueryResultOrderIsStableOnTies proves sortQueryFacts is load-bearing:
// two facts tied on (vocabulary, span) compare by ID last, giving one
// specified, stable order rather than leaving ties to map/slice iteration
// order (FND-04 ordering edge). This was manually verified during this
// plan's execution: temporarily replacing sortQueryFacts's final ID
// comparison with an unstable/no-op comparator reproduces out-of-order
// output for this exact fixture and this test goes red -- recorded in
// 06-03-SUMMARY.md.
func TestQueryResultOrderIsStableOnTies(t *testing.T) {
	facts := []protocol.QueryFact{
		{ID: "z-last", Vocabulary: "control", Availability: string(debugmap.Available)},
		{ID: "a-first", Vocabulary: "control", Availability: string(debugmap.Available)},
		{ID: "m-middle", Vocabulary: "control", Availability: string(debugmap.Available)},
	}
	sortQueryFacts(facts)
	want := []string{"a-first", "m-middle", "z-last"}
	for index, fact := range facts {
		if fact.ID != want[index] {
			t.Fatalf("order = %v, want %v (ties on vocabulary+span must break on ID)", factIDs(facts), want)
		}
	}

	// Cross-invocation stability: sorting the same tied set twice must not
	// depend on map iteration or any other nondeterministic input.
	again := []protocol.QueryFact{
		{ID: "z-last", Vocabulary: "control", Availability: string(debugmap.Available)},
		{ID: "a-first", Vocabulary: "control", Availability: string(debugmap.Available)},
		{ID: "m-middle", Vocabulary: "control", Availability: string(debugmap.Available)},
	}
	sortQueryFacts(again)
	for index := range facts {
		if facts[index].ID != again[index].ID {
			t.Fatalf("sortQueryFacts produced different order across two calls on identical input: %v vs %v", factIDs(facts), factIDs(again))
		}
	}
}

func factIDs(facts []protocol.QueryFact) []string {
	ids := make([]string, len(facts))
	for index, fact := range facts {
		ids[index] = fact.ID
	}
	return ids
}

// TestQueryDepthDoesNotPaginate covers Test 4: --depth changes join
// traversal depth only and must never affect pagination/bounding.
func TestQueryDepthDoesNotPaginate(t *testing.T) {
	path := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	diagID := firstDiagnosticID(t, path)

	withoutDepth, err := QueryCommandFile(path, diagID, QueryOptions{})
	if err != nil {
		t.Fatalf("QueryCommandFile: %v", err)
	}
	withDepth, err := QueryCommandFile(path, diagID, QueryOptions{Depth: 1})
	if err != nil {
		t.Fatalf("QueryCommandFile (depth=1): %v", err)
	}
	firstBytes, _ := json.Marshal(withoutDepth.Query)
	secondBytes, _ := json.Marshal(withDepth.Query)
	if string(firstBytes) != string(secondBytes) {
		t.Fatalf("--depth changed query's output; D-06-03 requires query to page by cursor only:\nwithout depth: %s\nwith depth=1:  %s", firstBytes, secondBytes)
	}
}

// --- shared test helpers ----------------------------------------------------

// firstDebugMapOperationID returns a real operation ID debug-map reports
// for path, following runDebugMap's own construction exactly.
func firstDebugMapOperationID(t *testing.T, path string) string {
	t.Helper()
	built := buildDebugMap(t, path)
	if len(built.Entries) == 0 {
		t.Fatalf("debugmap.Build produced no entries for %s", path)
	}
	return built.Entries[0].OperationID
}

func buildDebugMap(t *testing.T, path string) debugmap.Map {
	t.Helper()
	source, err := readBoundedFile(path, syntax.MaxSourceBytes)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	parsed := syntax.Parse(source)
	if len(parsed.Diagnostics) > 0 {
		t.Fatalf("%s failed to parse: %+v", path, parsed.Diagnostics)
	}
	checked := check.Program(parsed.Program)
	if len(checked.Diagnostics) > 0 {
		t.Fatalf("%s failed to check: %+v", path, checked.Diagnostics)
	}
	validated := corevalidate.Validate(checked.Program)
	if !validated.Valid {
		t.Fatalf("%s failed core validation: %+v", path, validated.Problems)
	}
	built, _, err := debugmap.Build(context.Background(), parsed.Program, validated.Program())
	if err != nil {
		t.Fatalf("debugmap.Build(%s): %v", path, err)
	}
	return built
}
