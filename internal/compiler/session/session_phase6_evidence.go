package session

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/evidence"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/syntax"
)

// MaxEvidenceTraceBytes bounds an expanded evidence trace's JSON encoding,
// mirroring the project's existing 64 KiB-plus-one output discipline
// (evidence.MaxToolProbeBytes, native.MaxStreamBytes). Exceeding it stops
// adding entries and reports the stable "truncated:evidence.trace_bound"
// code, rather than growing the trace unboundedly.
const MaxEvidenceTraceBytes = 64 * 1024

// TruncatedEvidenceTraceBound is the stable truncation code an expanded
// evidence trace reports when MaxEvidenceTraceBytes is exceeded, parallel to
// explain's "truncated:explain.depth"/"truncated:explain.node_budget" and
// query's "truncated:query.page_bound".
const TruncatedEvidenceTraceBound = "truncated:evidence.trace_bound"

// phase6BoundTraceEntries appends entries into a bounded result, stopping
// (and reporting TruncatedEvidenceTraceBound) the moment the accumulated
// JSON encoding would exceed MaxEvidenceTraceBytes -- a stable code, never
// silent truncation.
func phase6BoundTraceEntries(entries []protocol.TraceEntry) ([]protocol.TraceEntry, string) {
	bounded := make([]protocol.TraceEntry, 0, len(entries))
	size := 2 // "[]"
	for _, entry := range entries {
		encoded, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		additional := len(encoded) + 1 // +1 for the separating comma/bracket
		if size+additional > MaxEvidenceTraceBytes+1 {
			return bounded, TruncatedEvidenceTraceBound
		}
		size += additional
		bounded = append(bounded, entry)
	}
	return bounded, ""
}

// phase6EvidenceTrace independently recomputes evidence.Build(source, facts)
// and reports, for every digest-bearing field the manifest carries, the
// recorded value alongside the value recomputed fresh THIS invocation --
// never merely "match" or "mismatch": the actual two digests and which
// input they bound, so a reviewer can see exactly what diverged.
func phase6EvidenceTrace(source []byte, facts evidence.Facts, manifest evidence.Manifest) (*protocol.TraceSummary, error) {
	expected, diagnostics, err := evidence.Build(source, facts)
	if err != nil {
		return nil, err
	}
	if len(diagnostics) > 0 {
		return nil, fmt.Errorf("phase6.evidence_trace_source_invalid")
	}

	entries := []protocol.TraceEntry{
		{Input: "source_digest", Recorded: manifest.SourceDigest, Recomputed: expected.Manifest.SourceDigest},
		{Input: "core_digest", Recorded: manifest.CoreDigest, Recomputed: expected.Manifest.CoreDigest},
		{Input: "c_digest", Recorded: manifest.CDigest, Recomputed: expected.Manifest.CDigest},
		{Input: "id", Recorded: manifest.ID, Recomputed: expected.Manifest.ID},
	}
	if expected.Manifest.Schema == evidence.Schema1 {
		entries = append(entries, protocol.TraceEntry{Input: "foreign_digest", Recorded: manifest.ForeignDigest, Recomputed: expected.Manifest.ForeignDigest})
	}
	for index, recomputed := range expected.Manifest.ExecutionDigests {
		recorded := ""
		if index < len(manifest.ExecutionDigests) {
			recorded = manifest.ExecutionDigests[index]
		}
		entries = append(entries, protocol.TraceEntry{Input: fmt.Sprintf("execution_digest[%d]", index), Recorded: recorded, Recomputed: recomputed})
	}
	for index := range entries {
		entries[index].Match = entries[index].Recorded == entries[index].Recomputed
	}

	bounded, truncated := phase6BoundTraceEntries(entries)
	return &protocol.TraceSummary{Entries: bounded, Truncated: truncated}, nil
}

// ValidateEvidenceExpanded is a sibling of the shipped
// ValidateEvidenceCommandFile (session.go), not an edit to it: the default
// projection is compact (the existing EvidenceSummary only), and expansion
// -- attaching a TraceSummary -- triggers on either a validation failure or
// the explicit expand argument. Expansion is a projection over the SAME
// verdict Validate already computed, never a re-derivation that could
// itself disagree (TestExpansionNeverChangesTheVerdict): the Status and
// Diagnostics returned here are identical whether expand is true or false.
func ValidateEvidenceExpanded(ctx context.Context, manifestPath, sourcePath string, expand bool) protocol.Result {
	manifestBytes, err := readBoundedFile(manifestPath, evidence.MaxManifestBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read evidence manifest")
	}
	source, err := readBoundedFile(sourcePath, syntax.MaxSourceBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "tool.read_failed", "unable to read source")
	}
	manifest, err := evidence.DecodeStrict(manifestBytes)
	if err != nil {
		return commandProblem("evidence", protocol.StatusInvalid, evidence.ErrorCode(err), "evidence manifest is not valid")
	}
	facts, err := evidence.DefaultFacts(ctx, "clang")
	if err != nil {
		return commandProblem("evidence", protocol.StatusOperational, "evidence.tool_failure", "unable to inspect native toolchain")
	}

	validateErr := evidence.Validate(manifest, source, facts)

	result := protocol.New("evidence", protocol.StatusPass)
	summary := &protocol.EvidenceSummary{Schema: manifest.Schema, ID: manifest.ID, Digest: evidence.ContentDigest(manifestBytes)}
	result.Evidence = summary

	if validateErr != nil {
		result.Status = protocol.StatusInvalid
		result.Diagnostics = []diagnostic.Diagnostic{diagnostic.Error(evidence.ErrorCode(validateErr), diagnostic.Span{}, "evidence manifest does not match recomputed facts")}
	}

	if expand || validateErr != nil {
		if trace, traceErr := phase6EvidenceTrace(source, facts, manifest); traceErr == nil {
			summary.Trace = trace
		}
	}

	return result.Finalize()
}
