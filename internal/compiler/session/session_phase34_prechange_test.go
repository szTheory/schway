package session_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/protocol"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase34PreChangeSchema = "phase34.prechange-corpus/1"
	phase34PreChangeRows   = 162
)

type phase34PreChangeCorpus struct {
	Schema   string                  `json:"schema"`
	Revision string                  `json:"revision"`
	Entries  []phase34PreChangeEntry `json:"entries"`
}

type phase34PreChangeEntry struct {
	Path             string           `json:"path"`
	CheckFile        phase34Admission `json:"check_file"`
	CheckCommandFile phase34Admission `json:"check_command_file"`
	EmitApplication  phase34Emission  `json:"emit_application"`
	ControlNativeC   phase34Emission  `json:"control_native_c"`
}

type phase34Admission struct {
	Outcome     string              `json:"outcome"`
	Status      string              `json:"status,omitempty"`
	Error       string              `json:"error,omitempty"`
	Diagnostics []phase34Diagnostic `json:"diagnostics,omitempty"`
}

type phase34Emission struct {
	Outcome string `json:"outcome"`
	Reason  string `json:"reason,omitempty"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	CSource string `json:"c_source,omitempty"`
}

type phase34Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}

func phase34PreChangePath() string {
	return testsupport.ProjectPath("testdata", "phase34", "prechange-corpus.json")
}

func phase34TrackedSchwayPaths(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--", "examples", "testdata")
	cmd.Dir = testsupport.ProjectPath(".")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files examples testdata: %v", err)
	}
	var paths []string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasSuffix(line, ".schway") {
			paths = append(paths, filepath.ToSlash(line))
		}
	}
	sort.Strings(paths)
	if len(paths) != phase34PreChangeRows {
		t.Fatalf("tracked Schway inventory has %d rows, want frozen 162-row inventory", len(paths))
	}
	return paths
}

func phase34ObserveCorpus(t *testing.T) []phase34PreChangeEntry {
	t.Helper()
	paths := phase34TrackedSchwayPaths(t)
	entries := make([]phase34PreChangeEntry, 0, len(paths))
	for _, relative := range paths {
		path := testsupport.ProjectPath(filepath.FromSlash(relative))
		entry := phase34PreChangeEntry{Path: relative}
		checked, checkErr := session.CheckFile(path)
		entry.CheckFile = phase34CheckFileAdmission(checked.Diagnostics, checkErr)
		commandResult, commandErr := session.CheckCommandFile(path)
		entry.CheckCommandFile = phase34CommandAdmission(commandResult, commandErr)

		if checkErr != nil {
			entry.EmitApplication = phase34NotEmitted("CheckFile error: " + checkErr.Error())
			entry.ControlNativeC = phase34NotEmitted("CheckFile error: " + checkErr.Error())
		} else if len(checked.Diagnostics) != 0 {
			codes := phase34DiagnosticCodes(checked.Diagnostics)
			reason := "CheckFile diagnostics"
			if len(codes) > 0 {
				reason += ": " + strings.Join(codes, ",")
			}
			entry.EmitApplication = phase34NotEmitted(reason)
			entry.ControlNativeC = phase34NotEmitted(reason)
		} else {
			application, applicationErr := cgen.EmitApplication(checked.Program)
			entry.EmitApplication = phase34EmissionResult(application, applicationErr)
			control, controlErr := session.Phase16ControlNativeC(checked.Program, "phase34-prechange")
			entry.ControlNativeC = phase34EmissionResult(control, controlErr)
		}
		entries = append(entries, entry)
	}
	return entries
}

func phase34CheckFileAdmission(diagnostics []diagnostic.Diagnostic, err error) phase34Admission {
	if err != nil {
		return phase34Admission{Outcome: "error", Error: err.Error()}
	}
	if len(diagnostics) != 0 {
		return phase34Admission{Outcome: "diagnostics", Diagnostics: phase34DiagnosticRecords(diagnostics)}
	}
	return phase34Admission{Outcome: "admitted"}
}

func phase34CommandAdmission(result protocol.Result, err error) phase34Admission {
	if err != nil {
		return phase34Admission{Outcome: "error", Error: err.Error()}
	}
	return phase34Admission{Outcome: result.Status, Status: result.Status, Diagnostics: phase34DiagnosticRecords(result.Diagnostics)}
}

func phase34NotEmitted(reason string) phase34Emission {
	return phase34Emission{Outcome: "not_emitted", Reason: reason}
}

func phase34EmissionResult(source string, err error) phase34Emission {
	if err != nil {
		return phase34Emission{Outcome: "refused", Code: phase34ErrorCode(err.Error()), Error: err.Error()}
	}
	return phase34Emission{Outcome: "emitted", CSource: source}
}

func phase34ErrorCode(message string) string {
	if end := strings.IndexByte(message, ':'); end > 0 {
		return message[:end]
	}
	return ""
}

func phase34CompareCorpus(want phase34PreChangeCorpus, observed []phase34PreChangeEntry) []string {
	var problems []string
	if want.Schema != phase34PreChangeSchema {
		problems = append(problems, fmt.Sprintf("schema: got %q want %q", want.Schema, phase34PreChangeSchema))
	}
	paths := make([]string, len(observed))
	for i := range observed {
		paths[i] = observed[i].Path
	}
	if len(paths) != phase34PreChangeRows || len(want.Entries) != phase34PreChangeRows {
		problems = append(problems, fmt.Sprintf("inventory count: baseline=%d observed=%d want=%d", len(want.Entries), len(paths), phase34PreChangeRows))
	}
	if !equalStrings(paths, phase34EntryPaths(want.Entries)) {
		problems = append(problems, "tracked inventory path set or stable order changed")
	}
	if !sort.StringsAreSorted(paths) {
		problems = append(problems, "observed inventory is not in stable path order")
	}
	byPath := make(map[string]phase34PreChangeEntry, len(observed))
	for _, entry := range observed {
		byPath[entry.Path] = entry
	}
	for _, before := range want.Entries {
		after, ok := byPath[before.Path]
		if !ok {
			continue
		}
		// A previously accepted input must remain accepted. Failed diagnostic
		// presentation and recovery are historical evidence, not C identity.
		if before.CheckFile.Outcome == "admitted" && after.CheckFile.Outcome != "admitted" {
			problems = append(problems, before.Path+": previously successful CheckFile admission changed")
		}
		if before.CheckCommandFile.Status == "pass" && after.CheckCommandFile.Status != "pass" {
			problems = append(problems, before.Path+": previously successful CheckCommandFile admission changed")
		}
		phase34CompareEmission(&problems, before.Path, "emit_application", before.EmitApplication, after.EmitApplication)
		phase34CompareEmission(&problems, before.Path, "control_native_c", before.ControlNativeC, after.ControlNativeC)
	}
	return problems
}

func phase34CompareEmission(problems *[]string, path, name string, want, got phase34Emission) {
	if want.Outcome == "not_emitted" {
		// The emitter was not reached because CheckFile rejected the source. A
		// later diagnostic recovery can change this observation without changing
		// the identity of any successful C emission.
		return
	}
	if want.Outcome != got.Outcome {
		*problems = append(*problems, fmt.Sprintf("%s: %s outcome changed: got %q want %q", path, name, got.Outcome, want.Outcome))
		return
	}
	switch want.Outcome {
	case "emitted":
		if !bytes.Equal([]byte(want.CSource), []byte(got.CSource)) {
			index := firstByteDifference([]byte(want.CSource), []byte(got.CSource))
			*problems = append(*problems, fmt.Sprintf("%s: %s C byte mismatch at byte %d (got length %d, want length %d)", path, name, index, len(got.CSource), len(want.CSource)))
		}
	case "refused":
		if got.Code != want.Code || got.Error != want.Error {
			*problems = append(*problems, fmt.Sprintf("%s: %s refusal changed: got code %q error %q want code %q error %q", path, name, got.Code, got.Error, want.Code, want.Error))
		}
	default:
		*problems = append(*problems, fmt.Sprintf("%s: %s has unknown baseline outcome %q", path, name, want.Outcome))
	}
}

func phase34EntryPaths(entries []phase34PreChangeEntry) []string {
	paths := make([]string, len(entries))
	for i := range entries {
		paths[i] = entries[i].Path
	}
	return paths
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func firstByteDifference(left, right []byte) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for i := 0; i < limit; i++ {
		if left[i] != right[i] {
			return i
		}
	}
	return limit
}

func phase34LoadCorpus(t *testing.T) phase34PreChangeCorpus {
	t.Helper()
	raw, err := os.ReadFile(phase34PreChangePath())
	if err != nil {
		t.Fatal(err)
	}
	var corpus phase34PreChangeCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("decode pre-change corpus: %v", err)
	}
	return corpus
}

func phase34DiagnosticCodes(diagnostics []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		codes = append(codes, diagnostic.Code)
	}
	return codes
}

func phase34DiagnosticRecords(diagnostics []diagnostic.Diagnostic) []phase34Diagnostic {
	records := make([]phase34Diagnostic, 0, len(diagnostics))
	for _, item := range diagnostics {
		records = append(records, phase34Diagnostic{Code: item.Code, Message: item.Message, Start: item.Primary.Start, End: item.Primary.End})
	}
	return records
}

func TestPhase34PreChangeCProgramGolden(t *testing.T) {
	observed := phase34ObserveCorpus(t)
	corpus := phase34LoadCorpus(t)
	problems := phase34CompareCorpus(corpus, observed)
	if len(problems) != 0 {
		t.Fatalf("pre-change C corpus mismatch: %s", strings.Join(problems, "; "))
	}
	emitted := map[string]int{"emit_application": 0, "control_native_c": 0}
	nonEmitted := map[string]int{"emit_application": 0, "control_native_c": 0}
	for _, entry := range corpus.Entries {
		for name, result := range map[string]phase34Emission{"emit_application": entry.EmitApplication, "control_native_c": entry.ControlNativeC} {
			if result.Outcome == "emitted" {
				emitted[name]++
			} else {
				nonEmitted[name]++
			}
		}
	}
	for _, name := range []string{"emit_application", "control_native_c"} {
		if emitted[name] == 0 || nonEmitted[name] == 0 {
			t.Fatalf("%s outcomes need emitted and non-emitted rows; emitted=%d non-emitted=%d", name, emitted[name], nonEmitted[name])
		}
	}
}

func TestPhase34PreChangeCGoldenIsNotInert(t *testing.T) {
	corpus := phase34LoadCorpus(t)
	observed := phase34ObserveCorpus(t)
	if problems := phase34CompareCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("clean pre-change observation must pass before mutation: %s", strings.Join(problems, "; "))
	}
	for index := range observed {
		for _, emitter := range []struct {
			name string
			get  func(*phase34PreChangeEntry) *phase34Emission
		}{
			{"emit_application", func(entry *phase34PreChangeEntry) *phase34Emission { return &entry.EmitApplication }},
			{"control_native_c", func(entry *phase34PreChangeEntry) *phase34Emission { return &entry.ControlNativeC }},
		} {
			result := emitter.get(&observed[index])
			if result.Outcome != "emitted" || result.CSource == "" {
				continue
			}
			original := result.CSource
			mutated := []byte(original)
			mutated[0] ^= 0x01
			result.CSource = string(mutated)
			problems := phase34CompareCorpus(corpus, observed)
			result.CSource = original
			if clean := phase34CompareCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored observation did not pass: %s", strings.Join(clean, "; "))
			}
			needle := observed[index].Path + ": " + emitter.name + " C byte mismatch"
			for _, problem := range problems {
				if strings.HasPrefix(problem, needle) && strings.Contains(problem, "byte ") {
					return
				}
			}
			t.Fatalf("reached mutation was not named by the comparator; got %v", problems)
		}
	}
	t.Fatal("no successful literal C emission was available for the comparator mutation")
}
