package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codename-lang/lang/internal/compiler/diagnostic"
	"github.com/codename-lang/lang/internal/compiler/interp"
)

const Schema = "lang.command/0"

const (
	StatusPass        = "pass"
	StatusInvalid     = "invalid"
	StatusOperational = "operational_failure"
	StatusMismatch    = "semantic_mismatch"
	StatusUsage       = "usage_error"
)

type Metrics struct {
	ElapsedNS      int64  `json:"elapsed_ns"`
	PeakRSSStatus  string `json:"peak_rss_status"`
	PeakRSSBytes   int64  `json:"peak_rss_bytes,omitempty"`
	OutputBytes    int    `json:"output_bytes"`
	RecomputedWork int    `json:"recomputed_work"`
}

type EvidenceSummary struct {
	Schema string `json:"schema"`
	ID     string `json:"id"`
	Digest string `json:"digest"`
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
}

type Result struct {
	Schema      string                  `json:"schema"`
	Command     string                  `json:"command"`
	Status      string                  `json:"status"`
	ID          string                  `json:"id"`
	ModuleID    string                  `json:"module_id,omitempty"`
	Formatted   string                  `json:"formatted,omitempty"`
	Diagnostics []diagnostic.Diagnostic `json:"diagnostics"`
	Executions  []interp.Execution      `json:"executions"`
	Evidence    *EvidenceSummary        `json:"evidence,omitempty"`
	Lanes       []Lane                  `json:"lanes"`
	Metrics     Metrics                 `json:"metrics"`
}

func New(command, status string) Result {
	return Result{
		Schema: Schema, Command: command, Status: status,
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
		LaneIDs          []string
	}{
		Schema: result.Schema, Command: result.Command, Status: result.Status,
		ModuleID:         result.ModuleID,
		DiagnosticIDs:    make([]string, 0, len(result.Diagnostics)),
		ExecutionDigests: make([]string, 0, len(result.Executions)), LaneIDs: make([]string, 0, len(result.Lanes)),
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

func Human(result Result) string {
	result = result.Finalize()
	for attempts := 0; attempts < 4; attempts++ {
		output := human(result)
		if result.Metrics.OutputBytes == len(output) {
			return output
		}
		result.Metrics.OutputBytes = len(output)
	}
	return human(result)
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
			fmt.Fprintf(&output, "%s %s input=%s output=%s\n", event.ID, event.Kind, event.Input, event.Output)
		}
	}
	if result.Evidence != nil {
		fmt.Fprintf(&output, "%s %s digest=%s\n", result.Evidence.ID, result.Evidence.Schema, result.Evidence.Digest)
	}
	for _, lane := range result.Lanes {
		fmt.Fprintf(&output, "%s %s %s work=%d\n", lane.ID, lane.Schema, lane.Status, lane.RecomputedWork)
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
