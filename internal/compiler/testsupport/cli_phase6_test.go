// Phase 6's CLI seam, exercised through the shipped binary.
//
// The Phase 6 session layer is already covered in process: session_phase6_
// explain_test.go and session_phase6_query_test.go call ExplainCommandFile /
// QueryCommandFile directly, and deliberately so -- they live in `package
// session` to reach unexported synthesizer internals. What neither can reach
// is the layer a user (or an agent) actually touches: argv parsing, the
// JSON-to-stdout / human-to-stderr routing in main.go's emit, exit codes, and
// the schema strings as they appear in real output bytes.
//
// That gap is why explain, query, the coordinated /1 schema bump, and `lang
// stats` were each confirmed by hand during Phase 6 UAT rather than by a test.
// This file closes it: every assertion here spawns the built binary and reads
// its bytes, so each one is a falsifier that fires in `go test ./...` and in
// scripts/verify-phase6.sh without a human driving a terminal.
//
// It is a sibling to cli_test.go (Phase 1-3's CLI suite) rather than an
// extension of it, mirroring the per-phase session_phase6_*.go convention.
package testsupport_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/measure"
	"github.com/codename-lang/lang/internal/compiler/protocol"
	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

// phase6CLI builds the binary for one test. It is deliberately per-test
// rather than cached across the file: BuildCLI builds into t.TempDir(), which
// Go removes when that test finishes, so a shared binary path would dangle for
// every test after the first. This matches cli_test.go's existing per-test
// BuildCLI usage; the Go build cache makes the repeat builds cheap.
func phase6CLI(t *testing.T) string {
	t.Helper()
	return testsupport.BuildCLI(t)
}

// decodePhase6Result unmarshals a --json invocation's stdout, failing with the
// raw bytes on error so a malformed document is legible rather than a bare
// "unexpected end of JSON input".
func decodePhase6Result(t *testing.T, result testsupport.CLIResult) protocol.Result {
	t.Helper()
	var decoded protocol.Result
	if err := json.Unmarshal(result.Stdout, &decoded); err != nil {
		t.Fatalf("decode CLI JSON: %v (stdout: %s) (stderr: %s)", err, result.Stdout, result.Stderr)
	}
	return decoded
}

// cliDiagnosticID derives a real diagnostic ID by running `check` and reading
// the first diagnostic, rather than hardcoding a content hash that a fixture
// or hashing change would silently invalidate. It pins the expected code so
// that such a change surfaces here, at its cause, instead of downstream as a
// confusing explain.diagnostic_not_found.
func cliDiagnosticID(t *testing.T, binary, source, wantCode string) string {
	t.Helper()
	run := testsupport.RunCLI(t, binary, nil, "--json", "check", source)
	decoded := decodePhase6Result(t, run)
	if len(decoded.Diagnostics) == 0 {
		t.Fatalf("check %s produced no diagnostics; this fixture is the ID source for the CLI phase 6 suite", source)
	}
	if got := decoded.Diagnostics[0].Code; got != wantCode {
		t.Fatalf("check %s first diagnostic code = %q, want %q -- fixture drifted", source, got, wantCode)
	}
	if decoded.Diagnostics[0].ID == "" {
		t.Fatalf("check %s first diagnostic carries no ID", source)
	}
	return decoded.Diagnostics[0].ID
}

// cliDebugMapEntry returns the first fully-available debug-map entry, the
// source of the core_id / operation_id / point_id addresses query resolves.
func cliDebugMapEntry(t *testing.T, binary, source string) protocol.DebugMapEntry {
	t.Helper()
	run := testsupport.RunCLI(t, binary, nil, "--json", "debug-map", source)
	decoded := decodePhase6Result(t, run)
	if decoded.DebugMap == nil {
		t.Fatalf("debug-map %s carried no debug_map projection", source)
	}
	for _, entry := range decoded.DebugMap.Entries {
		if entry.Availability == "available" && entry.CoreID != "" && entry.OperationID != "" && entry.PointID != "" {
			return entry
		}
	}
	t.Fatalf("debug-map %s produced no fully-joined available entry: %+v", source, decoded.DebugMap.Entries)
	return protocol.DebugMapEntry{}
}

// cliVerify runs a corpus and returns the decoded result, the source of real
// lane IDs and control identifiers.
func cliVerify(t *testing.T, binary, corpus string) protocol.Result {
	t.Helper()
	run := testsupport.RunCLI(t, binary, nil, "--json", "verify", corpus)
	if run.Exit != 0 {
		t.Fatalf("verify %s exited %d: %s", corpus, run.Exit, run.Stderr)
	}
	return decodePhase6Result(t, run)
}

// TestExplainCLIReturnsBoundedCauseDAG is the shipped-binary falsifier for
// DX-02's explain half: `lang explain` answers a stable diagnostic ID with a
// bounded, well-formed cause DAG under lang.explain/0 -- never a dump of the
// whole program.
func TestExplainCLIReturnsBoundedCauseDAG(t *testing.T) {
	binary := phase6CLI(t)
	source := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	id := cliDiagnosticID(t, binary, source, "ownership.use_after_move")

	run := testsupport.RunCLI(t, binary, nil, "--json", "explain", source, id)
	if run.Exit != 0 || len(run.Stderr) != 0 {
		t.Fatalf("explain CLI failed: exit=%d stderr=%s", run.Exit, run.Stderr)
	}
	decoded := decodePhase6Result(t, run)
	if decoded.Command != "explain" || decoded.Status != protocol.StatusPass {
		t.Fatalf("explain result: command=%q status=%q", decoded.Command, decoded.Status)
	}
	if decoded.Explain == nil {
		t.Fatal("explain result carried no explain projection")
	}
	summary := decoded.Explain

	// The schema is asserted both through the constant and as a literal in
	// the emitted bytes, so renaming the constant cannot mask a regression in
	// what agents actually parse.
	if summary.Schema != protocol.ExplainSchema {
		t.Fatalf("explain schema = %q, want %q", summary.Schema, protocol.ExplainSchema)
	}
	if !bytes.Contains(run.Stdout, []byte(`"schema":"lang.explain/0"`)) {
		t.Fatalf("explain stdout omitted the literal lang.explain/0 schema: %s", run.Stdout)
	}
	if summary.RootID != id {
		t.Fatalf("explain root_id = %q, want the requested %q", summary.RootID, id)
	}

	// Boundedness -- the "without dumping the whole program" half of the
	// deliverable. A DAG that grew with program size instead of stopping at
	// its budget would fail here.
	if len(summary.Nodes) < 1 || len(summary.Nodes) > protocol.ExplainMaxNodes {
		t.Fatalf("explain node count %d outside [1, %d]", len(summary.Nodes), protocol.ExplainMaxNodes)
	}
	if len(summary.Edges) >= len(summary.Nodes)*len(summary.Nodes) {
		t.Fatalf("explain emitted %d edges for %d nodes -- not a bounded DAG", len(summary.Edges), len(summary.Nodes))
	}

	// Well-formedness: closed edge vocabulary, closed availability
	// vocabulary, and no edge referencing a node that is not present.
	nodeIDs := make(map[string]bool, len(summary.Nodes))
	for _, node := range summary.Nodes {
		nodeIDs[node.ID] = true
		switch node.Availability {
		case "available", "optimized_out", "not_captured":
		default:
			t.Fatalf("explain node %q carries availability %q outside the closed vocabulary", node.ID, node.Availability)
		}
	}
	for _, edge := range summary.Edges {
		switch edge.Kind {
		case protocol.EdgeCausedBy, protocol.EdgeNarrows, protocol.EdgeSameBinding:
		default:
			t.Fatalf("explain edge kind %q outside the closed vocabulary", edge.Kind)
		}
		if !nodeIDs[edge.From] || !nodeIDs[edge.To] {
			t.Fatalf("explain edge %s -> %s references a node absent from the DAG", edge.From, edge.To)
		}
	}

	// A shallower depth can never widen the DAG.
	shallow := testsupport.RunCLI(t, binary, nil, "--json", "explain", source, id, "--depth=1")
	if shallow.Exit != 0 {
		t.Fatalf("explain --depth=1 exited %d: %s", shallow.Exit, shallow.Stderr)
	}
	shallowDecoded := decodePhase6Result(t, shallow)
	if shallowDecoded.Explain == nil {
		t.Fatal("explain --depth=1 carried no explain projection")
	}
	if len(shallowDecoded.Explain.Nodes) > len(summary.Nodes) {
		t.Fatalf("explain --depth=1 produced %d nodes, more than the default depth's %d",
			len(shallowDecoded.Explain.Nodes), len(summary.Nodes))
	}

	// D-06-02's no-persisted-store discipline, observable end to end: two
	// cold invocations must agree byte for byte.
	again := testsupport.RunCLI(t, binary, nil, "--json", "explain", source, id)
	if !bytes.Equal(run.Stdout, again.Stdout) {
		t.Fatal("two cold explain invocations produced different bytes")
	}
}

// phase6QueryAddress is one address plus the vocabulary label the CLI must
// report for it.
type phase6QueryAddress struct {
	name       string
	source     string
	address    string
	vocabulary string
}

// phase6ResolvableAddresses derives one genuinely resolvable address per
// stable ID vocabulary, every one of them read back out of a real command's
// output rather than hardcoded.
func phase6ResolvableAddresses(t *testing.T, binary string) []phase6QueryAddress {
	t.Helper()
	useAfterMove := testsupport.ProjectPath("testdata", "phase2", "use_after_move.lang")
	ownedTransfer := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	phase1Corpus := testsupport.ProjectPath("testdata", "phase1")

	entry := cliDebugMapEntry(t, binary, ownedTransfer)
	verified := cliVerify(t, binary, phase1Corpus)
	if len(verified.Lanes) == 0 {
		t.Fatal("verify testdata/phase1 produced no lanes to address")
	}

	// The control vocabulary answers against session.AllShippedControlIDs()
	// -- the same live-derived registry QLT-01's audit cross-checks -- which
	// is Phase 4's and Phase 5's control sets. A Phase 1 lane control such as
	// control:match.non_exhaustive is genuinely absent from it, so the
	// resolvable control address is drawn from the registry itself rather
	// than from an arbitrary lane's Controls list.
	shipped := session.AllShippedControlIDs()
	if len(shipped) == 0 {
		t.Fatal("session.AllShippedControlIDs() is empty; no control address to resolve")
	}
	control := shipped[0]

	evidenceRun := testsupport.RunCLI(t, binary, nil, "--json", "evidence", ownedTransfer)
	evidenceDecoded := decodePhase6Result(t, evidenceRun)
	if evidenceDecoded.Evidence == nil || evidenceDecoded.Evidence.ID == "" {
		t.Fatalf("evidence %s produced no evidence ID (exit=%d stderr=%s)",
			ownedTransfer, evidenceRun.Exit, evidenceRun.Stderr)
	}

	return []phase6QueryAddress{
		{"diagnostic", useAfterMove, cliDiagnosticID(t, binary, useAfterMove, "ownership.use_after_move"), "diagnostic"},
		{"core_id", ownedTransfer, entry.CoreID, "core_id"},
		{"operation_id", ownedTransfer, entry.OperationID, "operation_id"},
		{"point_id", ownedTransfer, entry.PointID, "point_id"},
		{"evidence", ownedTransfer, evidenceDecoded.Evidence.ID, "evidence"},
		{"control", ownedTransfer, control, "control"},
		{"lane", phase1Corpus, verified.Lanes[0].ID, "lane"},
	}
}

// TestQueryCLIResolvesEveryStableIDVocabulary is the shipped-binary falsifier
// for DX-02's query half: one joined addressing surface resolves an address
// drawn from each of the five stable ID vocabularies already in the tree.
func TestQueryCLIResolvesEveryStableIDVocabulary(t *testing.T) {
	binary := phase6CLI(t)
	for _, address := range phase6ResolvableAddresses(t, binary) {
		t.Run(address.name, func(t *testing.T) {
			run := testsupport.RunCLI(t, binary, nil, "--json", "query", address.source, address.address)
			if run.Exit != 0 || len(run.Stderr) != 0 {
				t.Fatalf("query %s exited %d: %s", address.address, run.Exit, run.Stderr)
			}
			decoded := decodePhase6Result(t, run)
			if decoded.Query == nil {
				t.Fatal("query result carried no query projection")
			}
			if decoded.Query.Schema != protocol.QuerySchema {
				t.Fatalf("query schema = %q, want %q", decoded.Query.Schema, protocol.QuerySchema)
			}
			if !bytes.Contains(run.Stdout, []byte(`"schema":"lang.query/0"`)) {
				t.Fatalf("query stdout omitted the literal lang.query/0 schema: %s", run.Stdout)
			}
			if len(decoded.Query.Facts) == 0 {
				t.Fatalf("query %s returned no facts at all", address.address)
			}
			if len(decoded.Query.Facts) > protocol.QueryMaxFactsPerPage {
				t.Fatalf("query returned %d facts, above the %d page bound",
					len(decoded.Query.Facts), protocol.QueryMaxFactsPerPage)
			}
			// Facts are returned in a stable sorted order, so the resolved
			// subject is not necessarily Facts[0] -- a diagnostic address
			// also yields one fact per cause, some of which are honestly
			// not_captured. Assert the vocabulary across every fact, and
			// require at least one genuinely available fact so the address
			// is proven resolved rather than merely well-formed.
			available := 0
			for _, fact := range decoded.Query.Facts {
				if fact.Vocabulary != address.vocabulary {
					t.Fatalf("query %s returned fact %q with vocabulary %q, want %q",
						address.address, fact.ID, fact.Vocabulary, address.vocabulary)
				}
				if fact.Vocabulary == "unrecognized" {
					t.Fatalf("query failed to recognize %q, an address the tree itself supplied", address.address)
				}
				if fact.Availability == "available" {
					available++
				}
			}
			if available == 0 {
				t.Fatalf("query %s resolved no available fact for an address taken from live output: %+v",
					address.address, decoded.Query.Facts)
			}
		})
	}
}

// TestQueryCLIMintsNoSixthVocabulary asserts that no label outside the closed
// CLI-observable set can leak out of a real invocation. The structural
// "exactly five dispatch vocabularies" pin lives in
// session_phase6_query_test.go#TestQueryMintsNoSixthVocabulary; this is its
// shipped-binary peer.
func TestQueryCLIMintsNoSixthVocabulary(t *testing.T) {
	binary := phase6CLI(t)
	// The five dispatch vocabularies as they surface at the CLI: debugmap
	// reports the finer core_id / operation_id / point_id label it actually
	// matched, plus the honest unrecognized sentinel.
	allowed := map[string]bool{
		"diagnostic": true, "core_id": true, "operation_id": true, "point_id": true,
		"evidence": true, "control": true, "lane": true, "unrecognized": true,
	}
	addresses := phase6ResolvableAddresses(t, binary)
	addresses = append(addresses, phase6QueryAddress{
		name:       "garbage",
		source:     testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang"),
		address:    "totally-unrecognized-address",
		vocabulary: "unrecognized",
	})
	observed := make(map[string]bool)
	for _, address := range addresses {
		run := testsupport.RunCLI(t, binary, nil, "--json", "query", address.source, address.address)
		if run.Exit != 0 {
			t.Fatalf("query %s exited %d: %s", address.address, run.Exit, run.Stderr)
		}
		decoded := decodePhase6Result(t, run)
		if decoded.Query == nil {
			t.Fatalf("query %s carried no query projection", address.address)
		}
		for _, fact := range decoded.Query.Facts {
			if !allowed[fact.Vocabulary] {
				t.Fatalf("query %s minted vocabulary %q outside the closed set", address.address, fact.Vocabulary)
			}
			observed[fact.Vocabulary] = true
		}
	}
	// Every one of the five vocabularies was actually exercised, so this test
	// cannot pass vacuously by resolving nothing.
	for _, required := range []string{"diagnostic", "core_id", "operation_id", "point_id", "evidence", "control", "lane", "unrecognized"} {
		if !observed[required] {
			t.Fatalf("the CLI query sweep never exercised the %q vocabulary", required)
		}
	}
}

// TestQueryCLIUnknownIDReportsNotCaptured is the shipped-binary falsifier for
// FND-04's honest-absence contract: an ID that resolves to nothing is answered
// with a not_captured fact -- never an error, and never a fabricated value.
func TestQueryCLIUnknownIDReportsNotCaptured(t *testing.T) {
	binary := phase6CLI(t)
	ownedTransfer := testsupport.ProjectPath("testdata", "phase2", "owned_transfer.lang")
	phase1Corpus := testsupport.ProjectPath("testdata", "phase1")

	for _, absent := range []phase6QueryAddress{
		{"diagnostic", ownedTransfer, "diagnostic:" + strings.Repeat("0", 24), "diagnostic"},
		{"operation_id", ownedTransfer, "s1:absent.module:fn:missing:op:999999", "operation_id"},
		{"point_id", ownedTransfer, "s1:absent.module:fn:missing:point:999999", "point_id"},
		{"evidence", ownedTransfer, strings.Repeat("0", 64), "evidence"},
		{"control", ownedTransfer, "control:never.shipped.control", "control"},
		{"lane", phase1Corpus, "lane:never-shipped-lane", "lane"},
	} {
		t.Run(absent.name, func(t *testing.T) {
			run := testsupport.RunCLI(t, binary, nil, "--json", "query", absent.source, absent.address)
			// Honest absence is data, not a failure: the process must
			// succeed and report no diagnostics.
			if run.Exit != 0 {
				t.Fatalf("query for an absent %s exited %d, want 0: %s", absent.name, run.Exit, run.Stderr)
			}
			decoded := decodePhase6Result(t, run)
			if decoded.Status != protocol.StatusPass {
				t.Fatalf("query for an absent %s reported status %q, want pass", absent.name, decoded.Status)
			}
			if len(decoded.Diagnostics) != 0 {
				t.Fatalf("query for an absent %s emitted %d diagnostics; absence is not an error",
					absent.name, len(decoded.Diagnostics))
			}
			if decoded.Query == nil || len(decoded.Query.Facts) != 1 {
				t.Fatalf("query for an absent %s did not return exactly one fact: %+v", absent.name, decoded.Query)
			}
			fact := decoded.Query.Facts[0]
			if fact.Availability != "not_captured" {
				t.Fatalf("query for an absent %s reported availability %q, want not_captured",
					absent.name, fact.Availability)
			}
			if fact.Vocabulary != absent.vocabulary {
				t.Fatalf("query for an absent %s reported vocabulary %q, want %q",
					absent.name, fact.Vocabulary, absent.vocabulary)
			}
			// Nothing may be invented in place of the missing value.
			if fact.ID != absent.address {
				t.Fatalf("query substituted %q for the requested absent address %q", fact.ID, absent.address)
			}
			if fact.Span != nil {
				t.Fatalf("query fabricated a span %+v for an address it could not resolve", fact.Span)
			}
			if fact.Detail != "" {
				t.Fatalf("query fabricated detail %q for an address it could not resolve", fact.Detail)
			}
		})
	}
}

// TestVerifyPhase1CLIEmitsCoordinatedSchemaBump is the shipped-binary
// falsifier for D-06-31's coordinated additive bump: a real `verify` emits
// lang.command/1 at the top level and lang.verify-lane/1 on every lane, with
// no /0 string left anywhere in the document.
func TestVerifyPhase1CLIEmitsCoordinatedSchemaBump(t *testing.T) {
	binary := phase6CLI(t)
	run := testsupport.RunCLI(t, binary, nil, "--json", "verify", testsupport.ProjectPath("testdata", "phase1"))
	if run.Exit != 0 || len(run.Stderr) != 0 {
		t.Fatalf("verify testdata/phase1 exited %d: %s", run.Exit, run.Stderr)
	}
	decoded := decodePhase6Result(t, run)
	if decoded.Schema != protocol.Schema1 {
		t.Fatalf("top-level schema = %q, want %q", decoded.Schema, protocol.Schema1)
	}
	if len(decoded.Lanes) != 5 {
		t.Fatalf("verify testdata/phase1 reported %d lanes, want 5", len(decoded.Lanes))
	}
	for _, lane := range decoded.Lanes {
		if lane.Schema != protocol.LaneSchema1 {
			t.Fatalf("lane %s schema = %q, want %q", lane.ID, lane.Schema, protocol.LaneSchema1)
		}
	}
	for _, literal := range []string{`"schema":"lang.command/1"`, `"schema":"lang.verify-lane/1"`} {
		if !bytes.Contains(run.Stdout, []byte(literal)) {
			t.Fatalf("verify stdout omitted %s", literal)
		}
	}
	// The bump is coordinated: a half-landed or partially reverted change
	// would leave a /0 string behind in real output.
	for _, stale := range []string{"lang.command/0", "lang.verify-lane/0"} {
		if bytes.Contains(run.Stdout, []byte(stale)) {
			t.Fatalf("verify stdout still carries %s -- the coordinated bump half-landed", stale)
		}
	}
}

// TestStatsCLIComputesGoStatistics pins `lang stats` -- the seam
// scripts/verify-phase6.sh's 20-sample loop pipes through so the shell never
// reimplements percentile arithmetic -- against measure.Samples.Summary(),
// its own source of truth. The command had no test coverage of any kind
// before this one, despite the gate depending on it.
func TestStatsCLIComputesGoStatistics(t *testing.T) {
	binary := phase6CLI(t)

	samples := make(measure.Samples, 0, measure.WarmSampleCount)
	var stdin bytes.Buffer
	for index := 0; index < measure.WarmSampleCount; index++ {
		// A deliberately unsorted, non-uniform set: a stats command that
		// forgot to sort would pass on a monotonic one.
		value := int64(1_000_000 + ((index*7)%measure.WarmSampleCount)*1_000)
		samples = append(samples, value)
		fmt.Fprintf(&stdin, "%d\n", value)
	}
	want, err := samples.Summary()
	if err != nil {
		t.Fatalf("measure summary: %v", err)
	}

	run := testsupport.RunCLIStdin(t, binary, nil, stdin.Bytes(), "stats")
	if run.Exit != 0 {
		t.Fatalf("stats exited %d: %s", run.Exit, run.Stderr)
	}
	var got struct {
		P50   int64   `json:"p50"`
		P95   int64   `json:"p95"`
		CoV   float64 `json:"cov"`
		Count int     `json:"count"`
	}
	if err := json.Unmarshal(run.Stdout, &got); err != nil {
		t.Fatalf("decode stats JSON: %v (stdout: %s)", err, run.Stdout)
	}
	if got.P50 != want.P50 || got.P95 != want.P95 || got.Count != want.Count {
		t.Fatalf("stats = {p50:%d p95:%d count:%d}, want {p50:%d p95:%d count:%d}",
			got.P50, got.P95, got.Count, want.P50, want.P95, want.Count)
	}
	if got.Count != measure.WarmSampleCount {
		t.Fatalf("stats count = %d, want WarmSampleCount %d", got.Count, measure.WarmSampleCount)
	}
	// main.go renders CoV with %g, so compare the parsed float against the
	// same rendering rather than the raw bytes.
	wantCoV, err := strconv.ParseFloat(fmt.Sprintf("%g", want.CoV), 64)
	if err != nil {
		t.Fatalf("parse rendered CoV: %v", err)
	}
	if got.CoV != wantCoV {
		t.Fatalf("stats cov = %v, want %v", got.CoV, wantCoV)
	}

	// The two refusal paths the shell gate relies on to fail loudly.
	malformed := testsupport.RunCLIStdin(t, binary, nil, []byte("12x\n"), "stats")
	if malformed.Exit != 64 {
		t.Fatalf("stats on a malformed sample exited %d, want 64 (usage)", malformed.Exit)
	}
	empty := testsupport.RunCLIStdin(t, binary, nil, []byte(""), "stats")
	if empty.Exit == 0 {
		t.Fatal("stats on empty stdin exited 0; an empty sample set must fail loudly")
	}
}
