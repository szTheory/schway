package format

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const SchemaVersion = 1

var AbilityOrder = []string{"copy", "drop", "share", "send", "escape"}

type TypeExpr struct {
	Name string     `json:"name"`
	Args []TypeExpr `json:"args,omitempty"`
}

type TypeRule struct {
	Name        string           `json:"name"`
	Parameters  int              `json:"parameters,omitempty"`
	Base        []string         `json:"base,omitempty"`
	Conditional map[string][]int `json:"conditional,omitempty"`
}

type Summary struct {
	Ownership string   `json:"ownership"`
	Access    string   `json:"access,omitempty"`
	Origins   []string `json:"origins,omitempty"`
	Abilities []string `json:"abilities,omitempty"`
}

type FunctionFact struct {
	ID         string  `json:"id"`
	ReturnFact Summary `json:"return_fact"`
	Published  Summary `json:"published"`
}

type Event struct {
	ID    string `json:"id"`
	Op    string `json:"op"`
	Place string `json:"place"`
	Loan  string `json:"loan,omitempty"`
}

type Trace struct {
	ID     string   `json:"id"`
	Places []string `json:"places"`
	Events []Event  `json:"events"`
}

type Artifact struct {
	Schema          int            `json:"schema"`
	Module          string         `json:"module"`
	Types           []TypeRule     `json:"types"`
	AbilityRequests []TypeExpr     `json:"ability_requests"`
	Functions       []FunctionFact `json:"functions"`
	Traces          []Trace        `json:"traces"`
}

type PlaceState struct {
	Place     string `json:"place"`
	Ownership string `json:"ownership"`
}

type LoanState struct {
	ID     string `json:"id"`
	Place  string `json:"place"`
	Access string `json:"access"`
}

type Snapshot struct {
	EventID string       `json:"event_id"`
	Places  []PlaceState `json:"places"`
	Loans   []LoanState  `json:"loans,omitempty"`
}

type SummaryClaim struct {
	Function string  `json:"function"`
	Summary  Summary `json:"summary"`
}

type AbilityClaim struct {
	Type      TypeExpr `json:"type"`
	Abilities []string `json:"abilities"`
}

type TraceClaim struct {
	Trace       string       `json:"trace"`
	FinalPlaces []PlaceState `json:"final_places"`
	Snapshots   []Snapshot   `json:"snapshots,omitempty"`
}

type Certificate struct {
	Schema         int            `json:"schema"`
	Mode           string         `json:"mode"`
	ArtifactDigest string         `json:"artifact_digest"`
	Summaries      []SummaryClaim `json:"summaries"`
	Abilities      []AbilityClaim `json:"abilities"`
	Traces         []TraceClaim   `json:"traces"`
}

type Diagnostic struct {
	Code    string `json:"code"`
	Subject string `json:"subject,omitempty"`
	Event   string `json:"event,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func (d Diagnostic) Error() string {
	return fmt.Sprintf("%s: %s", d.Code, d.Detail)
}

func LoadArtifact(data []byte) (Artifact, error) {
	var artifact Artifact
	if err := json.Unmarshal(data, &artifact); err != nil {
		return Artifact{}, err
	}
	return artifact, nil
}

func CloneArtifact(artifact Artifact) Artifact {
	data, _ := json.Marshal(artifact)
	var clone Artifact
	_ = json.Unmarshal(data, &clone)
	return clone
}

func CloneCertificate(certificate Certificate) Certificate {
	data, _ := json.Marshal(certificate)
	var clone Certificate
	_ = json.Unmarshal(data, &clone)
	return clone
}

func CanonicalArtifact(artifact Artifact) Artifact {
	artifact = CloneArtifact(artifact)
	sort.Slice(artifact.Types, func(i, j int) bool { return artifact.Types[i].Name < artifact.Types[j].Name })
	for i := range artifact.Types {
		sort.Strings(artifact.Types[i].Base)
		for ability := range artifact.Types[i].Conditional {
			sort.Ints(artifact.Types[i].Conditional[ability])
		}
	}
	sort.Slice(artifact.AbilityRequests, func(i, j int) bool {
		return TypeKey(artifact.AbilityRequests[i]) < TypeKey(artifact.AbilityRequests[j])
	})
	sort.Slice(artifact.Functions, func(i, j int) bool { return artifact.Functions[i].ID < artifact.Functions[j].ID })
	for i := range artifact.Functions {
		canonicalizeSummary(&artifact.Functions[i].ReturnFact)
		canonicalizeSummary(&artifact.Functions[i].Published)
	}
	sort.Slice(artifact.Traces, func(i, j int) bool { return artifact.Traces[i].ID < artifact.Traces[j].ID })
	for i := range artifact.Traces {
		sort.Strings(artifact.Traces[i].Places)
	}
	return artifact
}

func ArtifactDigest(artifact Artifact) string {
	data, _ := json.Marshal(CanonicalArtifact(artifact))
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func TypeKey(expr TypeExpr) string {
	key := expr.Name
	if len(expr.Args) == 0 {
		return key
	}
	key += "["
	for index, arg := range expr.Args {
		if index > 0 {
			key += ","
		}
		key += TypeKey(arg)
	}
	return key + "]"
}

func CanonicalAbilities(values []string) []string {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	result := make([]string, 0, len(set))
	for _, ability := range AbilityOrder {
		if set[ability] {
			result = append(result, ability)
			delete(set, ability)
		}
	}
	var rest []string
	for ability := range set {
		rest = append(rest, ability)
	}
	sort.Strings(rest)
	return append(result, rest...)
}

func CanonicalSummary(summary Summary) Summary {
	canonicalizeSummary(&summary)
	return summary
}

func canonicalizeSummary(summary *Summary) {
	sort.Strings(summary.Origins)
	summary.Abilities = CanonicalAbilities(summary.Abilities)
}
