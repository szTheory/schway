package session_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/diagnostic"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase35PreChangeSchema = "phase35.prechange-corpus/1"
)

type phase35PreChangeCorpus struct {
	Schema      string                  `json:"schema"`
	Revision    string                  `json:"revision"`
	CoreSchema  string                  `json:"core_schema"`
	CoreSchema1 string                  `json:"core_schema_1"`
	CGenSchema  string                  `json:"cgen_schema"`
	Entries     []phase35PreChangeEntry `json:"entries"`
}

type phase35PreChangeEntry struct {
	Path        string           `json:"path"`
	Source      string           `json:"source"`
	SourceSHA   string           `json:"source_sha256"`
	Check       phase35Admission `json:"check"`
	CoreSchema  string           `json:"checked_core_schema,omitempty"`
	CoreSHA     string           `json:"checked_core_sha256,omitempty"`
	SemanticIDs []string         `json:"semantic_ids,omitempty"`
	Emit        phase35Emission  `json:"emit"`
	EmitNative  phase35Emission  `json:"emit_native"`
	Application phase35Emission  `json:"emit_application"`
}

type phase35Admission struct {
	Outcome     string              `json:"outcome"`
	Error       string              `json:"error,omitempty"`
	Diagnostics []phase35Diagnostic `json:"diagnostics,omitempty"`
}

type phase35Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}

type phase35Emission struct {
	Outcome string `json:"outcome"`
	Schema  string `json:"schema"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	Reason  string `json:"reason,omitempty"`
	CSource string `json:"c_source,omitempty"`
}

type phase35NumericDeltaFile struct {
	PreChangeCorpusAmendment phase35PreChangeAmendment `json:"prechange_corpus_amendment"`
}

type phase35PreChangeAmendment struct {
	Date      string                     `json:"date"`
	Path      string                     `json:"path"`
	SourceSHA string                     `json:"source_sha256"`
	Reason    string                     `json:"reason"`
	Before    phase35PreChangeProjection `json:"before"`
	After     phase35PreChangeProjection `json:"after"`
}

type phase35PreChangeProjection struct {
	Check      phase35Admission `json:"check"`
	EmitReason string           `json:"emit_reason"`
}

func phase35PreChangePath() string {
	return testsupport.ProjectPath("testdata", "phase35", "prechange-corpus.json")
}

func TestPhase35PreChangeEmitterCorpus(t *testing.T) {
	corpus := phase35LoadPreChangeCorpus(t)
	observed := phase35ObservePreChangeCorpus(t, corpus)
	if problems := phase35ComparePreChangeCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("pre-change emitter corpus mismatch: %s", strings.Join(problems, "; "))
	}
	phase35RequireOutcomeCoverage(t, corpus)
}

func TestPhase35PreChangeCorpusControls(t *testing.T) {
	corpus := phase35LoadPreChangeCorpus(t)
	observed := phase35ObservePreChangeCorpus(t, corpus)
	if problems := phase35ComparePreChangeCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("clean pre-change observation must pass before mutation: %s", strings.Join(problems, "; "))
	}

	byteMutationCaught := false
	for index := range observed {
		for _, route := range phase35RouteAccessors {
			result := route.get(&observed[index])
			if result.Outcome != "emitted" || result.CSource == "" {
				continue
			}
			mutant := phase35CloneEntries(observed)
			mutated := route.get(&mutant[index])
			cBytes := []byte(mutated.CSource)
			cBytes[0] ^= 1
			mutated.CSource = string(cBytes)
			problems := phase35ComparePreChangeCorpus(corpus, mutant)
			if !phase35HasProblem(problems, observed[index].Path+": "+route.name+" C byte mismatch") {
				t.Fatalf("changed C byte was not attributed to %s: %v", route.name, problems)
			}
			if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored C observation did not pass: %s", strings.Join(clean, "; "))
			}
			byteMutationCaught = true
			break
		}
		if byteMutationCaught {
			break
		}
	}
	if !byteMutationCaught {
		t.Fatal("no emitted C row was available for the changed-byte control")
	}

	coreMutationIndex := phase35FirstCoreEntry(observed)
	if coreMutationIndex < 0 {
		t.Fatal("no checked core identity was available for the core/ID control")
	}
	coreMutant := phase35CloneEntries(observed)
	coreMutant[coreMutationIndex].CoreSHA += "mutated"
	if problems := phase35ComparePreChangeCorpus(corpus, coreMutant); !phase35HasProblem(problems, observed[coreMutationIndex].Path+": checked core identity changed") {
		t.Fatalf("changed core identity was not attributed: %v", problems)
	}
	if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
		t.Fatalf("restored core identity did not pass: %s", strings.Join(clean, "; "))
	}

	idMutant := phase35CloneEntries(observed)
	idMutant[coreMutationIndex].SemanticIDs[0] += ":mutated"
	if problems := phase35ComparePreChangeCorpus(corpus, idMutant); !phase35HasProblem(problems, observed[coreMutationIndex].Path+": semantic IDs changed") {
		t.Fatalf("changed semantic ID was not attributed: %v", problems)
	}
	if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
		t.Fatalf("restored semantic IDs did not pass: %s", strings.Join(clean, "; "))
	}

	statusMutationCaught := false
	for index := range observed {
		for _, route := range phase35RouteAccessors {
			if route.get(&observed[index]).Outcome != "not_emitted" {
				continue
			}
			mutant := phase35CloneEntries(observed)
			mutated := route.get(&mutant[index])
			mutated.Outcome = "emitted"
			mutated.CSource = "mutated emission"
			problems := phase35ComparePreChangeCorpus(corpus, mutant)
			if !phase35HasProblem(problems, observed[index].Path+": "+route.name+" outcome changed") {
				t.Fatalf("not-emitted-to-emitted change was not attributed: %v", problems)
			}
			if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored status observation did not pass: %s", strings.Join(clean, "; "))
			}
			statusMutationCaught = true
			break
		}
		if statusMutationCaught {
			break
		}
	}
	if !statusMutationCaught {
		t.Fatal("no not-emitted result was available for the status control")
	}

	missingSource := phase35CloneEntries(observed)
	missingPath := missingSource[len(missingSource)-1].Path
	missingSource = missingSource[:len(missingSource)-1]
	if problems := phase35ComparePreChangeCorpus(corpus, missingSource); !phase35HasProblem(problems, missingPath+": source row missing from observation") {
		t.Fatalf("omitted source row was not attributed: %v", problems)
	}
	if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
		t.Fatalf("restored source inventory did not pass: %s", strings.Join(clean, "; "))
	}

	emitterOmissionCaught := false
	for index := range observed {
		for _, route := range phase35RouteAccessors {
			if route.get(&observed[index]).Outcome == "" {
				continue
			}
			mutant := phase35CloneEntries(observed)
			*route.get(&mutant[index]) = phase35Emission{}
			problems := phase35ComparePreChangeCorpus(corpus, mutant)
			if !phase35HasProblem(problems, observed[index].Path+": "+route.name+" outcome changed") {
				t.Fatalf("omitted emitter row was not attributed: %v", problems)
			}
			if clean := phase35ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored emitter inventory did not pass: %s", strings.Join(clean, "; "))
			}
			emitterOmissionCaught = true
			break
		}
		if emitterOmissionCaught {
			break
		}
	}
	if !emitterOmissionCaught {
		t.Fatal("no emitter row was available for the omission control")
	}
}

type phase35RouteAccessor struct {
	name string
	get  func(*phase35PreChangeEntry) *phase35Emission
}

var phase35RouteAccessors = []phase35RouteAccessor{
	{"emit", func(entry *phase35PreChangeEntry) *phase35Emission { return &entry.Emit }},
	{"emit_native", func(entry *phase35PreChangeEntry) *phase35Emission { return &entry.EmitNative }},
	{"emit_application", func(entry *phase35PreChangeEntry) *phase35Emission { return &entry.Application }},
}

func phase35ObservePreChangeCorpus(t *testing.T, corpus phase35PreChangeCorpus) []phase35PreChangeEntry {
	t.Helper()
	if corpus.Schema != phase35PreChangeSchema {
		t.Fatalf("pre-change fixture schema is %q, want %q", corpus.Schema, phase35PreChangeSchema)
	}
	if corpus.CoreSchema != core.Schema || corpus.CoreSchema1 != core.Schema1 || corpus.CGenSchema != cgen.Schema {
		t.Fatalf("pre-change fixture schema identifiers changed: core=%q core1=%q cgen=%q", corpus.CoreSchema, corpus.CoreSchema1, corpus.CGenSchema)
	}
	root := testsupport.ProjectPath(".")
	phase35VerifyRevision(t, root, corpus)
	sourceRoot := t.TempDir()
	for _, entry := range corpus.Entries {
		if filepath.IsAbs(entry.Path) || filepath.Clean(entry.Path) != entry.Path || strings.HasPrefix(entry.Path, "..") {
			t.Fatalf("unsafe frozen source path %q", entry.Path)
		}
		path := filepath.Join(sourceRoot, filepath.FromSlash(entry.Path))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("create frozen source directory for %s: %v", entry.Path, err)
		}
		if err := os.WriteFile(path, []byte(entry.Source), 0o600); err != nil {
			t.Fatalf("materialize frozen source %s: %v", entry.Path, err)
		}
	}

	observed := make([]phase35PreChangeEntry, 0, len(corpus.Entries))
	for _, before := range corpus.Entries {
		path := filepath.Join(sourceRoot, filepath.FromSlash(before.Path))
		entry := phase35PreChangeEntry{Path: before.Path, Source: before.Source, SourceSHA: phase35Digest([]byte(before.Source))}
		checked, checkErr := session.CheckFile(path)
		entry.Check = phase35AdmissionResult(checked.Diagnostics, checkErr)
		if checkErr != nil {
			reason := "CheckFile error: " + checkErr.Error()
			entry.Emit = phase35NotEmitted(reason)
			entry.EmitNative = phase35NotEmitted(reason)
			entry.Application = phase35NotEmitted(reason)
		} else if len(checked.Diagnostics) != 0 {
			reason := "CheckFile diagnostics"
			if codes := phase35DiagnosticCodes(checked.Diagnostics); len(codes) != 0 {
				reason += ": " + strings.Join(codes, ",")
			}
			entry.Emit = phase35NotEmitted(reason)
			entry.EmitNative = phase35NotEmitted(reason)
			entry.Application = phase35NotEmitted(reason)
		} else {
			coreBytes, marshalErr := json.Marshal(checked.Program)
			if marshalErr != nil {
				t.Fatalf("marshal checked core for %s: %v", before.Path, marshalErr)
			}
			entry.CoreSchema = checked.Program.Schema
			entry.CoreSHA = phase35Digest(coreBytes)
			entry.SemanticIDs = phase35SemanticIDs(t, coreBytes)
			entry.Emit = phase35EmitResult(cgen.Emit(checked.Program))
			entry.EmitNative = phase35EmitResult(cgen.EmitNative(checked.Program))
			entry.Application = phase35EmitResult(cgen.EmitApplication(checked.Program))
		}
		observed = append(observed, entry)
	}
	return observed
}

func phase35VerifyRevision(t *testing.T, root string, corpus phase35PreChangeCorpus) {
	t.Helper()
	treeCommand := exec.Command("git", "ls-tree", "-r", "-z", "--name-only", corpus.Revision)
	treeCommand.Dir = root
	pathsOutput, err := treeCommand.Output()
	if err != nil {
		t.Fatalf("git ls-tree frozen revision %s: %v", corpus.Revision, err)
	}
	var revisionPaths []string
	for _, path := range bytes.Split(pathsOutput, []byte{0}) {
		if len(path) != 0 && strings.HasSuffix(string(path), ".schway") {
			revisionPaths = append(revisionPaths, filepath.ToSlash(string(path)))
		}
	}
	sort.Strings(revisionPaths)
	fixturePaths := make([]string, len(corpus.Entries))
	for index, entry := range corpus.Entries {
		fixturePaths[index] = entry.Path
	}
	if !equalStrings(revisionPaths, fixturePaths) {
		t.Fatalf("fixture inventory does not equal .schway inventory at %s: revision=%d fixture=%d", corpus.Revision, len(revisionPaths), len(fixturePaths))
	}
	for _, entry := range corpus.Entries {
		cmd := exec.Command("git", "show", corpus.Revision+":"+entry.Path)
		cmd.Dir = root
		source, err := cmd.Output()
		if err != nil {
			t.Fatalf("read frozen source %s at %s: %v", entry.Path, corpus.Revision, err)
		}
		if !bytes.Equal(source, []byte(entry.Source)) {
			t.Fatalf("fixture source bytes differ from frozen revision for %s", entry.Path)
		}
		if phase35Digest(source) != entry.SourceSHA {
			t.Fatalf("fixture source digest does not match frozen source for %s", entry.Path)
		}
	}
}

func phase35AdmissionResult(diagnostics []diagnostic.Diagnostic, err error) phase35Admission {
	if err != nil {
		return phase35Admission{Outcome: "error", Error: err.Error()}
	}
	if len(diagnostics) == 0 {
		return phase35Admission{Outcome: "admitted"}
	}
	rows := make([]phase35Diagnostic, 0, len(diagnostics))
	for _, item := range diagnostics {
		rows = append(rows, phase35Diagnostic{Code: item.Code, Message: item.Message, Start: item.Primary.Start, End: item.Primary.End})
	}
	return phase35Admission{Outcome: "diagnostics", Diagnostics: rows}
}

func phase35EmitResult(source string, err error) phase35Emission {
	row := phase35Emission{Schema: cgen.Schema}
	if err != nil {
		row.Outcome = "refused"
		row.Error = err.Error()
		row.Code = phase35ErrorCode(err.Error())
		return row
	}
	row.Outcome = "emitted"
	row.CSource = source
	return row
}

func phase35NotEmitted(reason string) phase35Emission {
	return phase35Emission{Outcome: "not_emitted", Schema: cgen.Schema, Reason: reason}
}

func phase35ComparePreChangeCorpus(want phase35PreChangeCorpus, observed []phase35PreChangeEntry) []string {
	var problems []string
	if want.Schema != phase35PreChangeSchema {
		problems = append(problems, fmt.Sprintf("fixture schema changed: got %q want %q", want.Schema, phase35PreChangeSchema))
	}
	paths := make([]string, len(observed))
	byPath := make(map[string]phase35PreChangeEntry, len(observed))
	for index, entry := range observed {
		paths[index] = entry.Path
		byPath[entry.Path] = entry
	}
	wantPaths := make([]string, len(want.Entries))
	for index := range want.Entries {
		wantPaths[index] = want.Entries[index].Path
	}
	if len(paths) != len(wantPaths) {
		problems = append(problems, fmt.Sprintf("inventory row count changed: got %d want %d", len(paths), len(wantPaths)))
	}
	if !equalStrings(paths, wantPaths) {
		for _, path := range wantPaths {
			if !phase35ContainsString(paths, path) {
				problems = append(problems, path+": source row missing from observation")
			}
		}
		for _, path := range paths {
			if !phase35ContainsString(wantPaths, path) {
				problems = append(problems, path+": unexpected source row in observation")
			}
		}
		if sort.StringsAreSorted(paths) == false {
			problems = append(problems, "observed source inventory is not in stable path order")
		}
	}
	for _, before := range want.Entries {
		after, ok := byPath[before.Path]
		if !ok {
			continue
		}
		if before.SourceSHA != after.SourceSHA || before.Source != after.Source {
			problems = append(problems, before.Path+": source bytes changed")
		}
		if !reflectPhase35Admission(before.Check, after.Check) {
			problems = append(problems, before.Path+": source check status or diagnostics changed")
		}
		if before.CoreSchema != after.CoreSchema {
			problems = append(problems, before.Path+": checked core schema changed")
		}
		if before.CoreSHA != after.CoreSHA {
			problems = append(problems, before.Path+": checked core identity changed")
		}
		if !equalStrings(before.SemanticIDs, after.SemanticIDs) {
			problems = append(problems, before.Path+": semantic IDs changed")
		}
		for _, route := range phase35RouteAccessors {
			wantEmission := *route.get(&before)
			gotEmission := *route.get(&after)
			if before.Path == "examples/sum_to_n.schway" &&
				(route.name == "emit" || route.name == "emit_native") &&
				wantEmission.Outcome == "emitted" && gotEmission.Outcome == "emitted" {
				if wantEmission.Schema != gotEmission.Schema {
					problems = append(problems, fmt.Sprintf("%s: %s schema changed", before.Path, route.name))
				}
				_, compatibilityProblems := phase371OlderCompare(
					"phase36",
					before.Path,
					route.name,
					strings.TrimPrefix(after.SourceSHA, "sha256:"),
					strings.TrimPrefix(after.CoreSHA, "sha256:"),
					[]byte(wantEmission.CSource),
					[]byte(gotEmission.CSource),
				)
				problems = append(problems, compatibilityProblems...)
			} else {
				phase35CompareEmission(&problems, before.Path, route.name, wantEmission, gotEmission)
			}
		}
	}
	return problems
}

func phase35CompareEmission(problems *[]string, path, route string, want, got phase35Emission) {
	if want.Outcome != got.Outcome {
		*problems = append(*problems, fmt.Sprintf("%s: %s outcome changed: got %q want %q", path, route, got.Outcome, want.Outcome))
		return
	}
	if want.Schema != got.Schema {
		*problems = append(*problems, fmt.Sprintf("%s: %s schema changed", path, route))
	}
	switch want.Outcome {
	case "emitted":
		if want.CSource != got.CSource {
			index := phase35FirstByteDifference([]byte(want.CSource), []byte(got.CSource))
			*problems = append(*problems, fmt.Sprintf("%s: %s C byte mismatch at byte %d (got length %d, want length %d)", path, route, index, len(got.CSource), len(want.CSource)))
		}
	case "refused":
		if want.Code != got.Code || want.Error != got.Error {
			*problems = append(*problems, fmt.Sprintf("%s: %s refusal changed: got %q want %q", path, route, got.Error, want.Error))
		}
	case "not_emitted":
		if want.Reason != got.Reason {
			*problems = append(*problems, fmt.Sprintf("%s: %s non-emission reason changed", path, route))
		}
	default:
		*problems = append(*problems, fmt.Sprintf("%s: %s has unknown baseline outcome %q", path, route, want.Outcome))
	}
}

func TestPhase35PreChangeNumericDeltaControls(t *testing.T) {
	corpus := phase35ReadPreChangeCorpus(t)
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-delta.json"))
	if err != nil {
		t.Fatalf("read numeric compatibility amendment: %v", err)
	}
	var file phase35NumericDeltaFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode numeric compatibility amendment: %v", err)
	}
	amendment := file.PreChangeCorpusAmendment
	requirePhase35AmendmentRestored(t, corpus, amendment)

	if problems := phase35NumericOwnershipAmendmentProblems(corpus, phase35PreChangeAmendment{}); len(problems) == 0 {
		t.Fatal("missing numeric delta was not detected")
	}
	requirePhase35AmendmentRestored(t, corpus, amendment)

	wrongBefore := amendment
	wrongBefore.Before.Check.Diagnostics = append([]phase35Diagnostic(nil), amendment.Before.Check.Diagnostics...)
	wrongBefore.Before.Check.Diagnostics[0].Start++
	if problems := phase35NumericOwnershipAmendmentProblems(corpus, wrongBefore); !phase35HasProblem(problems, "delta content differs from its exact locked projection") {
		t.Fatalf("wrong before-state was not attributed: %v", problems)
	}
	requirePhase35AmendmentRestored(t, corpus, amendment)

	wrongAfter := amendment
	wrongAfter.After.Check.Diagnostics = append([]phase35Diagnostic(nil), amendment.After.Check.Diagnostics...)
	wrongAfter.After.Check.Diagnostics[0].End++
	if problems := phase35NumericOwnershipAmendmentProblems(corpus, wrongAfter); !phase35HasProblem(problems, "delta content differs from its exact locked projection") {
		t.Fatalf("wrong after-state was not attributed: %v", problems)
	}
	requirePhase35AmendmentRestored(t, corpus, amendment)

	missingSource := corpus
	missingSource.Entries = append([]phase35PreChangeEntry(nil), corpus.Entries...)
	for index, entry := range missingSource.Entries {
		if entry.Path == amendment.Path {
			missingSource.Entries = append(missingSource.Entries[:index], missingSource.Entries[index+1:]...)
			break
		}
	}
	if problems := phase35NumericOwnershipAmendmentProblems(missingSource, amendment); !phase35HasProblem(problems, "pre-change corpus is missing amended source") {
		t.Fatalf("missing delta source row was not attributed: %v", problems)
	}
	requirePhase35AmendmentRestored(t, corpus, amendment)
}

func requirePhase35AmendmentRestored(t *testing.T, corpus phase35PreChangeCorpus, amendment phase35PreChangeAmendment) {
	t.Helper()
	if problems := phase35NumericOwnershipAmendmentProblems(corpus, amendment); len(problems) != 0 {
		t.Fatalf("restored numeric compatibility amendment did not pass: %v", problems)
	}
}

func phase35RequireOutcomeCoverage(t *testing.T, corpus phase35PreChangeCorpus) {
	t.Helper()
	if len(corpus.Entries) == 0 || !sort.SliceIsSorted(corpus.Entries, func(i, j int) bool { return corpus.Entries[i].Path < corpus.Entries[j].Path }) {
		t.Fatal("pre-change fixture must contain a non-empty sorted source inventory")
	}
	emitted := map[string]int{}
	nonEmitted := map[string]int{}
	checked := 0
	for _, entry := range corpus.Entries {
		if entry.Path == "" || entry.SourceSHA == "" || phase35Digest([]byte(entry.Source)) != entry.SourceSHA {
			t.Fatalf("source bytes/digest missing or invalid for %q", entry.Path)
		}
		if entry.Check.Outcome == "admitted" {
			checked++
			if entry.CoreSchema == "" || entry.CoreSHA == "" || len(entry.SemanticIDs) == 0 {
				t.Fatalf("admitted row lacks checked core identity/semantic IDs: %s", entry.Path)
			}
		} else if entry.CoreSchema != "" || entry.CoreSHA != "" || len(entry.SemanticIDs) != 0 {
			t.Fatalf("non-admitted row carries a core identity: %s", entry.Path)
		}
		for _, route := range phase35RouteAccessors {
			result := route.get(&entry)
			if result.Outcome == "" || result.Schema != cgen.Schema {
				t.Fatalf("%s has no explicit outcome/schema row: %s", route.name, entry.Path)
			}
			if entry.Check.Outcome != "admitted" && result.Outcome != "not_emitted" {
				t.Fatalf("%s crossed an emitter despite source refusal: %s", route.name, entry.Path)
			}
			if result.Outcome == "emitted" {
				emitted[route.name]++
				if result.CSource == "" {
					t.Fatalf("%s emitted an empty C program: %s", route.name, entry.Path)
				}
			} else {
				nonEmitted[route.name]++
			}
		}
	}
	if checked == 0 {
		t.Fatal("tracked source corpus has no checked success to exercise C emitters")
	}
	for _, route := range phase35RouteAccessors {
		if emitted[route.name] == 0 || nonEmitted[route.name] == 0 {
			t.Fatalf("%s needs emitted and explicit non-emitted/refused rows, emitted=%d non-emitted=%d", route.name, emitted[route.name], nonEmitted[route.name])
		}
	}
}

func phase35SemanticIDs(t *testing.T, coreBytes []byte) []string {
	t.Helper()
	var value any
	if err := json.Unmarshal(coreBytes, &value); err != nil {
		t.Fatalf("decode checked core identity: %v", err)
	}
	ids := make(map[string]struct{})
	var visit func(any)
	visit = func(node any) {
		switch typed := node.(type) {
		case map[string]any:
			for key, child := range typed {
				if key == "id" || strings.HasSuffix(key, "_id") {
					if id, ok := child.(string); ok && id != "" {
						ids[id] = struct{}{}
					}
				}
				visit(child)
			}
		case []any:
			for _, child := range typed {
				visit(child)
			}
		}
	}
	visit(value)
	result := make([]string, 0, len(ids))
	for id := range ids {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

func phase35LoadPreChangeCorpus(t *testing.T) phase35PreChangeCorpus {
	t.Helper()
	corpus := phase35ReadPreChangeCorpus(t)
	phase35ApplyNumericOwnershipAmendment(t, &corpus)
	return corpus
}

func phase35ReadPreChangeCorpus(t *testing.T) phase35PreChangeCorpus {
	t.Helper()
	raw, err := os.ReadFile(phase35PreChangePath())
	if err != nil {
		t.Fatalf("read immutable Phase 35 pre-change corpus: %v", err)
	}
	var corpus phase35PreChangeCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("decode immutable Phase 35 pre-change corpus: %v", err)
	}
	if corpus.Revision == "" {
		t.Fatal("pre-change corpus is missing its source revision")
	}
	return corpus
}

func phase35ApplyNumericOwnershipAmendment(t *testing.T, corpus *phase35PreChangeCorpus) {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase35", "numeric-delta.json"))
	if err != nil {
		t.Fatalf("read numeric compatibility amendment: %v", err)
	}
	var file phase35NumericDeltaFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode numeric compatibility amendment: %v", err)
	}
	amendment := file.PreChangeCorpusAmendment
	problems := phase35NumericOwnershipAmendmentProblems(*corpus, amendment)
	if len(problems) != 0 {
		t.Fatalf("numeric compatibility amendment does not match its frozen before-state: %s", strings.Join(problems, "; "))
	}
	for index := range corpus.Entries {
		if corpus.Entries[index].Path == amendment.Path {
			corpus.Entries[index].Check = amendment.After.Check
			afterEmission := phase35Emission{Outcome: "not_emitted", Schema: cgen.Schema, Reason: amendment.After.EmitReason}
			corpus.Entries[index].Emit = afterEmission
			corpus.Entries[index].EmitNative = afterEmission
			corpus.Entries[index].Application = afterEmission
			return
		}
	}
	t.Fatalf("validated numeric compatibility source disappeared from corpus: %s", amendment.Path)
}

func phase35NumericOwnershipAmendmentProblems(corpus phase35PreChangeCorpus, amendment phase35PreChangeAmendment) []string {
	malformedSpan := phase35Diagnostic{Code: "syntax.malformed_numeric_literal", Message: "malformed numeric literal", Start: 100, End: 105}
	duplicateSpan := phase35Diagnostic{Code: "syntax.expected_binding_source", Message: "expected `identifier`", Start: 100, End: 105}
	want := phase35PreChangeAmendment{
		Date:      "2026-10-14",
		Path:      "testdata/phase19/literal_malformed.schway",
		SourceSHA: "sha256:3dc93519d28272ba037adf9ef1fa1a213d5d4e6fdba23a8a80999bc78d39f8a6",
		Reason:    "Phase 35 exact-span malformed-number ownership removes the parser expectation duplicated at the lexer's TokenUnknown span. The frozen Phase 35 pre-change corpus remains unchanged; only this source-hash-bound live projection is amended.",
		Before: phase35PreChangeProjection{
			Check:      phase35Admission{Outcome: "diagnostics", Diagnostics: []phase35Diagnostic{malformedSpan, duplicateSpan}},
			EmitReason: "CheckFile diagnostics: syntax.malformed_numeric_literal,syntax.expected_binding_source",
		},
		After: phase35PreChangeProjection{
			Check:      phase35Admission{Outcome: "diagnostics", Diagnostics: []phase35Diagnostic{malformedSpan}},
			EmitReason: "CheckFile diagnostics: syntax.malformed_numeric_literal",
		},
	}
	var problems []string
	if !reflect.DeepEqual(amendment, want) {
		problems = append(problems, "delta content differs from its exact locked projection")
	}
	index := -1
	for candidate := range corpus.Entries {
		if corpus.Entries[candidate].Path == want.Path {
			if index >= 0 {
				problems = append(problems, "pre-change corpus contains duplicate amendment path "+want.Path)
				break
			}
			index = candidate
		}
	}
	if index < 0 {
		problems = append(problems, "pre-change corpus is missing amended source "+want.Path)
		return problems
	}
	entry := &corpus.Entries[index]
	if entry.SourceSHA != want.SourceSHA || phase35Digest([]byte(entry.Source)) != want.SourceSHA || !reflect.DeepEqual(entry.Check, want.Before.Check) {
		problems = append(problems, "frozen pre-change source or diagnostic before-state changed")
	}
	beforeEmission := phase35Emission{Outcome: "not_emitted", Schema: cgen.Schema, Reason: want.Before.EmitReason}
	for _, route := range phase35RouteAccessors {
		if !reflect.DeepEqual(*route.get(entry), beforeEmission) {
			problems = append(problems, "frozen pre-change "+route.name+" before-state changed")
		}
	}
	return problems
}

func phase35CloneEntries(entries []phase35PreChangeEntry) []phase35PreChangeEntry {
	clone := append([]phase35PreChangeEntry(nil), entries...)
	for index := range clone {
		clone[index].SemanticIDs = append([]string(nil), entries[index].SemanticIDs...)
		clone[index].Check.Diagnostics = append([]phase35Diagnostic(nil), entries[index].Check.Diagnostics...)
	}
	return clone
}

func phase35FirstCoreEntry(entries []phase35PreChangeEntry) int {
	for index, entry := range entries {
		if entry.CoreSHA != "" && len(entry.SemanticIDs) != 0 {
			return index
		}
	}
	return -1
}

func phase35HasProblem(problems []string, prefix string) bool {
	for _, problem := range problems {
		if strings.HasPrefix(problem, prefix) {
			return true
		}
	}
	return false
}

func phase35Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func phase35ErrorCode(message string) string {
	if end := strings.IndexByte(message, ':'); end > 0 {
		return message[:end]
	}
	return ""
}

func phase35DiagnosticCodes(diagnostics []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, item := range diagnostics {
		codes = append(codes, item.Code)
	}
	return codes
}

func phase35FirstByteDifference(left, right []byte) int {
	limit := len(left)
	if len(right) < limit {
		limit = len(right)
	}
	for index := 0; index < limit; index++ {
		if left[index] != right[index] {
			return index
		}
	}
	return limit
}

func reflectPhase35Admission(left, right phase35Admission) bool {
	if left.Outcome != right.Outcome || left.Error != right.Error || len(left.Diagnostics) != len(right.Diagnostics) {
		return false
	}
	for index := range left.Diagnostics {
		if left.Diagnostics[index] != right.Diagnostics[index] {
			return false
		}
	}
	return true
}

func phase35ContainsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
