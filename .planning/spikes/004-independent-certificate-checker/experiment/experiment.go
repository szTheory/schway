package experiment

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"example.com/ai-lang/spike004/format"
	"example.com/ai-lang/spike004/producer"
	"example.com/ai-lang/spike004/verifier"
)

type EncodingResult struct {
	Mode             string `json:"mode"`
	ArtifactBytes    int    `json:"artifact_bytes"`
	CertificateBytes int    `json:"certificate_bytes"`
	Checks           int    `json:"checks"`
	P50Micros        int64  `json:"p50_micros"`
	P95Micros        int64  `json:"p95_micros"`
}

type MutationResult struct {
	ID                string `json:"id"`
	Class             string `json:"class"`
	Detected          bool   `json:"detected"`
	Code              string `json:"code,omitempty"`
	Expected          string `json:"expected"`
	ReducedEventCount int    `json:"reduced_event_count,omitempty"`
}

type Report struct {
	Schema            int              `json:"schema"`
	FixtureFunctions  int              `json:"fixture_functions"`
	FixtureTraces     int              `json:"fixture_traces"`
	FixtureEvents     int              `json:"fixture_events"`
	Mutations         []MutationResult `json:"mutations"`
	DetectedMutations int              `json:"detected_mutations"`
	EscapedMutations  int              `json:"escaped_mutations"`
	ScaleEvents       int              `json:"scale_events"`
	Encodings         []EncodingResult `json:"encodings"`
	RecommendedUse    string           `json:"recommended_use"`
	RejectedUse       string           `json:"rejected_use"`
	ExperimentVerdict string           `json:"experiment_verdict"`
}

func LoadFixture(path string) (format.Artifact, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return format.Artifact{}, err
	}
	return format.LoadArtifact(data)
}

func Run(artifact format.Artifact) (Report, error) {
	for _, mode := range []string{"recompute", "replay"} {
		certificate, err := producer.Build(artifact, mode)
		if err != nil {
			return Report{}, err
		}
		if result := verifier.Verify(artifact, certificate); !result.Valid {
			return Report{}, fmt.Errorf("valid %s certificate rejected: %+v", mode, result.Diagnostics)
		}
	}
	report := Report{
		Schema:            1,
		FixtureFunctions:  len(artifact.Functions),
		FixtureTraces:     len(artifact.Traces),
		FixtureEvents:     eventCount(artifact),
		RecommendedUse:    "package/cache/CI/release boundary validation bound to canonical typed-core digests",
		RejectedUse:       "claiming source-to-core correctness or paying replay-certificate cost on every keystroke",
		ExperimentVerdict: "PARTIAL",
	}
	report.Mutations = runMutations(artifact)
	for _, mutation := range report.Mutations {
		if mutation.Detected {
			report.DetectedMutations++
		} else {
			report.EscapedMutations++
		}
	}
	scale := ScaleArtifact(10001)
	report.ScaleEvents = eventCount(scale)
	for _, mode := range []string{"recompute", "replay"} {
		certificate, err := producer.Build(scale, mode)
		if err != nil {
			return Report{}, err
		}
		artifactJSON, _ := json.Marshal(scale)
		certificateJSON, _ := json.Marshal(certificate)
		var durations []int64
		checks := 0
		for iteration := 0; iteration < 21; iteration++ {
			started := time.Now()
			result := verifier.Verify(scale, certificate)
			durations = append(durations, time.Since(started).Microseconds())
			checks = result.Checks
			if !result.Valid {
				return Report{}, fmt.Errorf("scale verification failed: %+v", result.Diagnostics)
			}
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		report.Encodings = append(report.Encodings, EncodingResult{
			Mode: mode, ArtifactBytes: len(artifactJSON), CertificateBytes: len(certificateJSON),
			Checks: checks, P50Micros: durations[len(durations)/2], P95Micros: durations[len(durations)*95/100],
		})
	}
	return report, nil
}

func runMutations(artifact format.Artifact) []MutationResult {
	base, _ := producer.Build(artifact, "recompute")
	replay, _ := producer.Build(artifact, "replay")
	tests := []struct {
		id, class, expected string
		artifact            format.Artifact
		certificate         format.Certificate
	}{
		{"stale-artifact", "binding", "certificate.subject_mismatch", mutateModule(artifact), base},
		{"forged-digest", "binding", "certificate.subject_mismatch", artifact, mutateDigest(base)},
		{"omitted-certificate-origin", "summary", "certificate.summary_mismatch", artifact, mutateClaimOrigin(base)},
		{"omitted-published-origin", "summary", "artifact.summary_mismatch", mutatePublishedOrigin(artifact), base},
		{"borrowed-as-owned", "summary", "artifact.summary_mismatch", mutateOwnership(artifact), base},
		{"exclusive-as-shared", "summary", "artifact.summary_mismatch", mutateAccess(artifact), base},
		{"forged-copy-ability", "ability", "certificate.ability_mismatch", artifact, mutateAbility(base)},
		{"forged-snapshot", "flow-witness", "certificate.snapshot_mismatch", artifact, mutateSnapshot(replay)},
		{"forged-final-state", "flow-witness", "certificate.final_state_mismatch", artifact, mutateFinal(base)},
		{"core-use-after-move", "flow", "ownership.use_after_moved", mutateUseAfterMove(artifact), base},
		{"core-missing-loan-end", "flow", "ownership.borrow_conflict", mutateMissingEnd(artifact), base},
		{"duplicate-event-id", "identity", "artifact.duplicate_id", mutateDuplicateEvent(artifact), base},
		{"coordinated-frontend-summary-lie", "trust-boundary", "ESCAPES_BY_DESIGN", mutateCoordinatedLie(artifact), base},
	}
	var results []MutationResult
	for _, test := range tests {
		candidate := format.CloneArtifact(test.artifact)
		certificate := format.CloneCertificate(test.certificate)
		// All mutations except stale-artifact and forged-digest model a producer
		// that can update the binding. The checker must rely on semantics, not hash mismatch.
		if test.id != "stale-artifact" && test.id != "forged-digest" {
			certificate.ArtifactDigest = format.ArtifactDigest(candidate)
		}
		if test.id == "omitted-published-origin" {
			setSummary(&certificate, "choose", func(summary *format.Summary) { summary.Origins = []string{"left"} })
		}
		if test.id == "borrowed-as-owned" {
			setSummary(&certificate, "head", func(summary *format.Summary) {
				summary.Ownership = "owned"
				summary.Access = ""
				summary.Origins = nil
			})
		}
		if test.id == "exclusive-as-shared" {
			setSummary(&certificate, "head_mut", func(summary *format.Summary) { summary.Access = "shared" })
		}
		if test.id == "coordinated-frontend-summary-lie" {
			setSummary(&certificate, "head", func(summary *format.Summary) { summary.Origins = []string{"attacker"} })
		}
		result := verifier.Verify(candidate, certificate)
		code := firstCode(result)
		detected := !result.Valid
		reduced := 0
		switch test.id {
		case "core-use-after-move", "core-missing-loan-end":
			reduced = minimizeTrace(candidate, certificate, "shared_then_move", test.expected)
		case "duplicate-event-id":
			reduced = minimizeTrace(candidate, certificate, "exclusive_then_destroy", test.expected)
		}
		results = append(results, MutationResult{ID: test.id, Class: test.class, Detected: detected, Code: code, Expected: test.expected, ReducedEventCount: reduced})
	}
	return results
}

func mutateModule(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Module += ".changed"
	return a
}
func mutateDigest(c format.Certificate) format.Certificate {
	c = format.CloneCertificate(c)
	c.ArtifactDigest = "sha256:0000"
	return c
}
func mutateClaimOrigin(c format.Certificate) format.Certificate {
	c = format.CloneCertificate(c)
	setSummary(&c, "choose", func(summary *format.Summary) { summary.Origins = []string{"left"} })
	return c
}
func mutatePublishedOrigin(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Functions[1].Published.Origins = []string{"left"}
	return a
}
func mutateOwnership(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Functions[0].Published.Ownership = "owned"
	a.Functions[0].Published.Access = ""
	a.Functions[0].Published.Origins = nil
	return a
}
func mutateAccess(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Functions[2].Published.Access = "shared"
	return a
}
func mutateAbility(c format.Certificate) format.Certificate {
	c = format.CloneCertificate(c)
	c.Abilities[0].Abilities = append(c.Abilities[0].Abilities, "copy")
	return c
}
func mutateSnapshot(c format.Certificate) format.Certificate {
	c = format.CloneCertificate(c)
	c.Traces[0].Snapshots[0].Places[0].Ownership = "moved"
	return c
}

func mutateFinal(c format.Certificate) format.Certificate {
	c = format.CloneCertificate(c)
	c.Traces[0].FinalPlaces[0].Ownership = "owned"
	return c
}
func mutateUseAfterMove(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Traces[0].Events = append(a.Traces[0].Events, format.Event{ID: "e8", Op: "use", Place: "buffer"})
	return a
}
func mutateMissingEnd(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Traces[0].Events[2] = format.Event{ID: "e3", Op: "borrow_exclusive", Place: "buffer", Loan: "l3"}
	return a
}
func mutateDuplicateEvent(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Traces[1].Events[2].ID = "e6"
	return a
}
func mutateCoordinatedLie(a format.Artifact) format.Artifact {
	a = format.CloneArtifact(a)
	a.Functions[0].ReturnFact.Origins = []string{"attacker"}
	a.Functions[0].Published.Origins = []string{"attacker"}
	return a
}

func firstCode(result verifier.Result) string {
	if len(result.Diagnostics) == 0 {
		return ""
	}
	return result.Diagnostics[0].Code
}

func setSummary(certificate *format.Certificate, function string, update func(*format.Summary)) {
	for index := range certificate.Summaries {
		if certificate.Summaries[index].Function == function {
			update(&certificate.Summaries[index].Summary)
			return
		}
	}
}

func minimizeTrace(artifact format.Artifact, certificate format.Certificate, traceID, targetCode string) int {
	traceIndex := -1
	for index := range artifact.Traces {
		if artifact.Traces[index].ID == traceID {
			traceIndex = index
			break
		}
	}
	if traceIndex < 0 {
		return 0
	}
	changed := true
	for changed {
		changed = false
		for index := range artifact.Traces[traceIndex].Events {
			candidate := format.CloneArtifact(artifact)
			events := candidate.Traces[traceIndex].Events
			candidate.Traces[traceIndex].Events = append(append([]format.Event(nil), events[:index]...), events[index+1:]...)
			probe := format.CloneCertificate(certificate)
			probe.ArtifactDigest = format.ArtifactDigest(candidate)
			if hasCode(verifier.Verify(candidate, probe), targetCode) {
				artifact = candidate
				changed = true
				break
			}
		}
	}
	return len(artifact.Traces[traceIndex].Events)
}

func hasCode(result verifier.Result, code string) bool {
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func eventCount(artifact format.Artifact) int {
	total := 0
	for _, trace := range artifact.Traces {
		total += len(trace.Events)
	}
	return total
}

func ScaleArtifact(events int) format.Artifact {
	if events < 1 {
		events = 1
	}
	trace := format.Trace{ID: "scale", Places: []string{"buffer"}}
	for index := 0; index+1 < events; index += 2 {
		loan := fmt.Sprintf("l%05d", index/2)
		trace.Events = append(trace.Events,
			format.Event{ID: fmt.Sprintf("e%05d", index), Op: "borrow_shared", Place: "buffer", Loan: loan},
			format.Event{ID: fmt.Sprintf("e%05d", index+1), Op: "end_loan", Place: "buffer", Loan: loan},
		)
	}
	if len(trace.Events) < events {
		trace.Events = append(trace.Events, format.Event{ID: fmt.Sprintf("e%05d", len(trace.Events)), Op: "move", Place: "buffer"})
	}
	return format.Artifact{Schema: format.SchemaVersion, Module: "scale", Traces: []format.Trace{trace}}
}
