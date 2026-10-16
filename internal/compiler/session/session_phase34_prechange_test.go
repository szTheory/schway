package session_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
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
	phase35AdlerSourceSHA  = "16490db4e53cb9db287727a8e5918745aa412659e799dd2fd53b38805bb3bef1"
	phase35AdlerOutputSHA  = "8653dbec998b2d39975619912b400e465b299d9d60ef59bdb1397ed9991d8e6e"
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
	Emit             phase34Emission  `json:"-"`
	EmitNative       phase34Emission  `json:"-"`
	sourceBytes      []byte           `json:"-"`
	sourceSHA        string           `json:"-"`
	checkedCoreBytes []byte           `json:"-"`
	coreSHA          string           `json:"-"`
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

type phase34Phase371Census struct {
	Schema  string                 `json:"schema"`
	Entries []phase34Phase371Entry `json:"entries"`
}

type phase34Phase371Entry struct {
	Corpus    string                 `json:"corpus"`
	Path      string                 `json:"path"`
	SourceSHA string                 `json:"source_sha256"`
	CoreSHA   string                 `json:"checked_core_sha256"`
	Routes    []phase34Phase371Route `json:"routes"`
}

type phase34Phase371Route struct {
	Name            string `json:"route"`
	Outcome         string `json:"outcome"`
	Code            string `json:"code,omitempty"`
	Error           string `json:"error,omitempty"`
	Reason          string `json:"reason,omitempty"`
	CBase64         string `json:"c_base64,omitempty"`
	CBytes          int    `json:"c_bytes,omitempty"`
	CSHA            string `json:"c_sha256,omitempty"`
	FrozenOutcome   string `json:"frozen_outcome,omitempty"`
	FrozenCBytes    int    `json:"frozen_c_bytes,omitempty"`
	FrozenCSHA      string `json:"frozen_c_sha256,omitempty"`
	PriorProjection string `json:"prior_projection,omitempty"`
}

type phase34Phase371Ledger struct {
	Schema  string                     `json:"schema"`
	Entries []phase34Phase371Amendment `json:"entries"`
}

type phase34Phase371Amendment struct {
	Corpus          string                    `json:"corpus"`
	Path            string                    `json:"path"`
	Route           string                    `json:"route"`
	SourceSHA       string                    `json:"source_sha256"`
	CoreSHA         string                    `json:"checked_core_sha256"`
	OldCSHA         string                    `json:"old_c_sha256"`
	PriorProjection string                    `json:"prior_projection,omitempty"`
	Edits           []phase34Phase371ByteEdit `json:"edits"`
	ExpectedCSHA    string                    `json:"expected_c_sha256"`
}

type phase34Phase371ByteEdit struct {
	Offset       int    `json:"offset"`
	BeforeB64    string `json:"before_base64"`
	AfterB64     string `json:"after_base64"`
	AnchorB64    string `json:"anchor_base64"`
	AnchorOffset int    `json:"anchor_offset"`
	AnchorCount  int    `json:"anchor_count"`
}

var (
	phase34Phase371Once sync.Once
	phase34Phase371Rows map[string]struct {
		entry phase34Phase371Entry
		route phase34Phase371Route
	}
	phase34Phase371Amendments map[string]phase34Phase371Amendment
	phase34Phase371Problems   []string
	phase34Phase371Loaded     bool
)

func phase34PreChangePath() string {
	return testsupport.ProjectPath("testdata", "phase34", "prechange-corpus.json")
}

func phase34TrackedSchwayPaths(t *testing.T) []string {
	t.Helper()
	tracked := phase34CurrentTrackedSchwayPaths(t)
	corpus := phase34LoadCorpus(t)
	paths := phase34EntryPaths(corpus.Entries)
	if len(paths) != phase34PreChangeRows {
		t.Fatalf("frozen pre-change corpus has %d paths, want %d", len(paths), phase34PreChangeRows)
	}
	for _, path := range paths {
		if !phase34Contains(tracked, path) {
			t.Fatalf("frozen pre-change Schway source %q is no longer tracked", path)
		}
	}
	var additions []string
	for _, path := range tracked {
		if !phase34Contains(paths, path) {
			additions = append(additions, path)
		}
	}
	ledger := phase34LoadNewFixtureLedger(t)
	phase37Ledger := phase34LoadPhase37FixtureLedger(t)
	if problems := phase34AdditionProblems(paths, ledger, phase37Ledger, additions); len(problems) != 0 {
		t.Fatalf("tracked source inventory additions: %s", strings.Join(problems, "; "))
	}
	delta, err := os.ReadFile(testsupport.ProjectPath(".planning", "phases", "35-bitwise-operators-literal-shifts-and-hex-literals", "35-COMPATIBILITY-DELTA.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(delta), "`examples/bitwise.schway`") {
		t.Fatal("Phase 35 compatibility delta does not declare the additive tracked witness")
	}
	return paths
}

func phase34CurrentTrackedSchwayPaths(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--", "examples", "testdata")
	cmd.Dir = testsupport.ProjectPath(".")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-files examples testdata: %v", err)
	}
	tracked := make(map[string]bool)
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasSuffix(line, ".schway") {
			tracked[filepath.ToSlash(line)] = true
		}
	}
	var paths []string
	for path := range tracked {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

// Compose only the Phase35 additive witness and the exact three Phase36 ledger
// rows. The current git inventory is independent of that ledger; it must close
// in both directions. No old row gains an exemption from its frozen C checks.
type phase34NewFixtureLedger struct {
	Schema  string                 `json:"schema"`
	Entries []phase34NewFixtureRow `json:"entries"`
}

type phase34NewFixtureRow struct {
	Path      string `json:"path"`
	SourceSHA string `json:"source_sha256"`
	Reason    string `json:"reason"`
}

type phase34Phase37FixtureLedger struct {
	Schema  string                 `json:"schema"`
	Entries []phase34NewFixtureRow `json:"entries"`
}

func phase34LoadNewFixtureLedger(t *testing.T) phase34NewFixtureLedger {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "new-fixture-ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger phase34NewFixtureLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode Phase36 fixture ledger: %v", err)
	}
	return ledger
}

func phase34LoadPhase37FixtureLedger(t *testing.T) phase34Phase37FixtureLedger {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase37", "new-fixture-ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger phase34Phase37FixtureLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode Phase37 fixture ledger: %v", err)
	}
	return ledger
}

func phase34AdditionProblems(frozen []string, ledger phase34NewFixtureLedger, phase37 phase34Phase37FixtureLedger, additions []string) []string {
	want := []string{
		"examples/helper_calls/crc_step.schway",
		"examples/helper_calls/importer.schway",
		"examples/helper_calls/public.schway",
	}
	var problems []string
	if ledger.Schema != "phase36.new-fixture-ledger/1" {
		problems = append(problems, "fixture ledger schema changed")
	}
	var ledgerPaths []string
	seen := map[string]bool{}
	for _, row := range ledger.Entries {
		ledgerPaths = append(ledgerPaths, row.Path)
		if phase34Contains(frozen, row.Path) || row.Path == "examples/bitwise.schway" {
			problems = append(problems, row.Path+": old corpus source cannot be exempted")
		}
		if seen[row.Path] {
			problems = append(problems, "ledger paths must be sorted and unique")
		}
		seen[row.Path] = true
		if !phase34Contains(want, row.Path) {
			// Never read a ledger-supplied path outside the finite new set.
			problems = append(problems, row.Path+": unsafe new-fixture ledger path")
			continue
		}
		if row.SourceSHA == "" || strings.TrimSpace(row.Reason) == "" {
			problems = append(problems, row.Path+": requires path, source hash, and reason")
		}
		if len(row.SourceSHA) != len("sha256:")+64 || !strings.HasPrefix(row.SourceSHA, "sha256:") || strings.Trim(row.SourceSHA[len("sha256:"):], "0123456789abcdef") != "" {
			problems = append(problems, row.Path+": source hash is malformed")
		}
		source, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(row.Path)))
		if err != nil {
			problems = append(problems, row.Path+": source is missing from the working tree")
		} else if row.SourceSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) {
			problems = append(problems, row.Path+": source hash differs from the working tree")
		}
	}
	if !equalStrings(ledgerPaths, want) {
		problems = append(problems, "Phase36 ledger must contain exactly the three declared helper paths in order")
	}
	wantPhase37 := []string{
		"examples/crc32/crc32.schway",
		"examples/crc32/crc_low.schway",
		"examples/crc32/lib/crc32.schway",
	}
	if phase37.Schema != "phase37.new-fixture-ledger/1" {
		problems = append(problems, "Phase37 fixture ledger schema changed")
	}
	phase37Paths := make([]string, 0, len(phase37.Entries))
	phase37Seen := map[string]bool{}
	for _, row := range phase37.Entries {
		phase37Paths = append(phase37Paths, row.Path)
		if phase34Contains(frozen, row.Path) || row.Path == "examples/bitwise.schway" || phase34Contains(want, row.Path) {
			problems = append(problems, row.Path+": old corpus source cannot be exempted")
		}
		if phase37Seen[row.Path] {
			problems = append(problems, "Phase37 ledger paths must be sorted and unique")
		}
		phase37Seen[row.Path] = true
		if !phase34Contains(wantPhase37, row.Path) {
			problems = append(problems, row.Path+": unsafe Phase37 new-fixture ledger path")
			continue
		}
		if strings.TrimSpace(row.Reason) == "" {
			problems = append(problems, row.Path+": Phase37 ledger requires a nonblank reason")
		}
		if len(row.SourceSHA) != len("sha256:")+64 || !strings.HasPrefix(row.SourceSHA, "sha256:") || strings.Trim(row.SourceSHA[len("sha256:"):], "0123456789abcdef") != "" {
			problems = append(problems, row.Path+": Phase37 source hash is malformed")
		}
		source, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(row.Path)))
		if err != nil {
			problems = append(problems, row.Path+": Phase37 source is missing from the working tree")
		} else if row.SourceSHA != fmt.Sprintf("sha256:%x", sha256.Sum256(source)) {
			problems = append(problems, row.Path+": Phase37 source hash differs from the working tree")
		}
	}
	if !equalStrings(phase37Paths, wantPhase37) {
		problems = append(problems, "Phase37 ledger must contain exactly the three CRC source paths in order")
	}
	allowed := append([]string{"examples/bitwise.schway"}, ledgerPaths...)
	allowed = append(allowed, phase37Paths...)
	for _, path := range additions {
		if !phase34Contains(allowed, path) {
			problems = append(problems, path+": tracked source is neither frozen nor explicitly ledgered")
		}
	}
	for _, path := range allowed {
		if !phase34Contains(additions, path) {
			problems = append(problems, path+": frozen or ledgered source is missing from tracked inventory")
		}
	}
	return problems
}

func phase34ObserveCorpus(t *testing.T) []phase34PreChangeEntry {
	t.Helper()
	paths := phase34TrackedSchwayPaths(t)
	entries := make([]phase34PreChangeEntry, 0, len(paths))
	for _, relative := range paths {
		path := testsupport.ProjectPath(filepath.FromSlash(relative))
		entry := phase34PreChangeEntry{Path: relative}
		var sourceErr error
		entry.sourceBytes, sourceErr = os.ReadFile(path)
		if sourceErr != nil {
			t.Fatalf("read source bytes for %s: %v", relative, sourceErr)
		}
		entry.sourceSHA = phase34Digest(entry.sourceBytes)
		checked, checkErr := session.CheckFile(path)
		entry.CheckFile = phase34CheckFileAdmission(checked.Diagnostics, checkErr)
		commandResult, commandErr := session.CheckCommandFile(path)
		entry.CheckCommandFile = phase34CommandAdmission(commandResult, commandErr)

		if checkErr != nil {
			entry.EmitApplication = phase34NotEmitted("CheckFile error: " + checkErr.Error())
			entry.ControlNativeC = phase34NotEmitted("CheckFile error: " + checkErr.Error())
			entry.Emit = phase34NotEmitted("CheckFile error: " + checkErr.Error())
			entry.EmitNative = phase34NotEmitted("CheckFile error: " + checkErr.Error())
		} else if len(checked.Diagnostics) != 0 {
			codes := phase34DiagnosticCodes(checked.Diagnostics)
			reason := "CheckFile diagnostics"
			if len(codes) > 0 {
				reason += ": " + strings.Join(codes, ",")
			}
			entry.EmitApplication = phase34NotEmitted(reason)
			entry.ControlNativeC = phase34NotEmitted(reason)
			entry.Emit = phase34NotEmitted(reason)
			entry.EmitNative = phase34NotEmitted(reason)
		} else {
			var marshalErr error
			entry.checkedCoreBytes, marshalErr = json.Marshal(checked.Program)
			if marshalErr != nil {
				t.Fatalf("marshal checked core for %s: %v", relative, marshalErr)
			}
			entry.coreSHA = phase34Digest(entry.checkedCoreBytes)
			emitted, emitErr := cgen.Emit(checked.Program)
			entry.Emit = phase34EmissionResult(emitted, emitErr)
			native, nativeErr := cgen.EmitNative(checked.Program)
			entry.EmitNative = phase34EmissionResult(native, nativeErr)
			application, applicationErr := cgen.EmitApplication(checked.Program)
			entry.EmitApplication = phase34EmissionResult(application, applicationErr)
			control, controlErr := session.Phase16ControlNativeC(checked.Program, "phase34-prechange")
			entry.ControlNativeC = phase34EmissionResult(control, controlErr)
		}
		entries = append(entries, entry)
	}
	return entries
}

func phase34Digest(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
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
		phase34CompareEmission(&problems, before.Path, "emit_application", before.EmitApplication, after.EmitApplication, after.sourceSHA, after.coreSHA)
		phase34CompareEmission(&problems, before.Path, "control_native_c", before.ControlNativeC, after.ControlNativeC, after.sourceSHA, after.coreSHA)
		phase34Phase371Projection(&problems, before.Path, "emit", after.sourceSHA, after.coreSHA, after.Emit)
		phase34Phase371Projection(&problems, before.Path, "emit_native", after.sourceSHA, after.coreSHA, after.EmitNative)
	}
	return problems
}

func phase34Phase371Projection(problems *[]string, path, route, sourceSHA, coreSHA string, got phase34Emission) bool {
	censusPath := testsupport.ProjectPath("testdata", "phase371", "pre-emitter-census.json")
	if _, err := os.Stat(censusPath); err != nil {
		return false
	}
	phase34Phase371Once.Do(func() {
		phase34Phase371Rows = map[string]struct {
			entry phase34Phase371Entry
			route phase34Phase371Route
		}{}
		phase34Phase371Amendments = map[string]phase34Phase371Amendment{}
		var census phase34Phase371Census
		if err := phase34ReadJSON(censusPath, &census); err != nil {
			phase34Phase371Problems = append(phase34Phase371Problems, "census: "+err.Error())
			return
		}
		if census.Schema != "phase371.pre-emitter-census/1" {
			phase34Phase371Problems = append(phase34Phase371Problems, "census schema changed")
		}
		ledgerPath := testsupport.ProjectPath("testdata", "phase371", "compatibility-amendments.json")
		var ledger phase34Phase371Ledger
		if err := phase34ReadJSON(ledgerPath, &ledger); err != nil {
			phase34Phase371Problems = append(phase34Phase371Problems, "amendments: "+err.Error())
			return
		}
		if ledger.Schema != "phase371.compatibility-amendments/1" {
			phase34Phase371Problems = append(phase34Phase371Problems, "amendment ledger schema changed")
		}
		for _, entry := range census.Entries {
			if entry.Corpus != "phase34" {
				continue
			}
			for _, row := range entry.Routes {
				key := entry.Path + "\x00" + row.Name
				phase34Phase371Rows[key] = struct {
					entry phase34Phase371Entry
					route phase34Phase371Route
				}{entry, row}
			}
		}
		seen := map[string]bool{}
		last := ""
		for _, amendment := range ledger.Entries {
			if amendment.Corpus != "phase34" {
				continue
			}
			key := amendment.Corpus + "\x00" + amendment.Path + "\x00" + amendment.Route
			if seen[key] {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": duplicate amendment row")
			}
			if last != "" && key < last {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": amendment rows must be sorted")
			}
			seen[key] = true
			last = key
			bound, ok := phase34Phase371Rows[amendment.Path+"\x00"+amendment.Route]
			if !ok {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": extra, wrong-path, or wrong-route amendment")
				continue
			}
			if bound.route.Outcome != "emitted" || amendment.SourceSHA != bound.entry.SourceSHA || amendment.CoreSHA != bound.entry.CoreSHA || amendment.OldCSHA != bound.route.CSHA || amendment.PriorProjection != bound.route.PriorProjection {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": stale or wrong old source/core/C binding")
			}
			old, err := base64.StdEncoding.DecodeString(bound.route.CBase64)
			if err != nil {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": invalid census C bytes")
				continue
			}
			updated, err := phase34Phase371Apply(old, amendment.Edits)
			if err != nil || phase34Digest(updated) != amendment.ExpectedCSHA {
				phase34Phase371Problems = append(phase34Phase371Problems, key+": exact reviewed delta or expected post-delta digest mismatch")
			}
			phase34Phase371Amendments[bound.entry.Path+"\x00"+bound.route.Name] = amendment
		}
		phase34Phase371Loaded = true
	})
	phase34Phase371Loaded = true
	for _, problem := range phase34Phase371Problems {
		*problems = append(*problems, "Phase371 census/amendment: "+problem)
	}
	key := path + "\x00" + route
	bound, ok := phase34Phase371Rows[key]
	if !ok {
		*problems = append(*problems, path+"/"+route+": missing from complete Phase371 census")
		return true
	}
	if bound.entry.SourceSHA != sourceSHA || bound.entry.CoreSHA != coreSHA {
		*problems = append(*problems, path+"/"+route+": source or checked-core bytes changed since the pre-emitter census")
	}
	gotOutcome := got.Outcome
	if gotOutcome == "not_emitted" {
		gotOutcome = "inapplicable"
	}
	if bound.route.Outcome != gotOutcome {
		*problems = append(*problems, fmt.Sprintf("%s: %s outcome changed: got %q want %q", path, route, gotOutcome, bound.route.Outcome))
		return true
	}
	if bound.route.Outcome == "inapplicable" {
		return true
	}
	if bound.route.Outcome == "refused" {
		if bound.route.Code != got.Code || bound.route.Error != got.Error {
			*problems = append(*problems, fmt.Sprintf("%s: %s refusal changed", path, route))
		}
		return true
	}
	if bound.route.PriorProjection != "" {
		binding, exists := phase34AdlerRewriteCRow(path, route)
		if !exists || bound.route.CBytes != binding.outputBytes || bound.route.CSHA != binding.outputSHA || bound.route.FrozenCBytes != binding.beforeBytes || bound.route.FrozenCSHA != binding.beforeSHA || !phase34AdlerCompatibilityDeltaBound() {
			*problems = append(*problems, path+": Phase35 exact prior projection no longer binds the Phase371 census")
		}
	}
	want, err := base64.StdEncoding.DecodeString(bound.route.CBase64)
	if err != nil {
		*problems = append(*problems, path+"/"+route+": invalid census C bytes")
		return true
	}
	if amendment, amended := phase34Phase371Amendments[key]; amended {
		want, err = phase34Phase371Apply(want, amendment.Edits)
		if err != nil {
			*problems = append(*problems, path+"/"+route+": "+err.Error())
			return true
		}
	}
	if !bytes.Equal(want, []byte(got.CSource)) {
		index := firstByteDifference(want, []byte(got.CSource))
		*problems = append(*problems, fmt.Sprintf("%s: %s C byte mismatch at byte %d (got length %d, want length %d)", path, route, index, len(got.CSource), len(want)))
	}
	return true
}

func phase34ReadJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func phase34Phase371Apply(old []byte, edits []phase34Phase371ByteEdit) ([]byte, error) {
	if len(edits) == 0 {
		return nil, fmt.Errorf("amendment has no reviewed byte edits")
	}
	ordered := append([]phase34Phase371ByteEdit(nil), edits...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Offset > ordered[j].Offset })
	lastStart := len(old) + 1
	result := append([]byte(nil), old...)
	for _, edit := range ordered {
		before, e1 := base64.StdEncoding.DecodeString(edit.BeforeB64)
		after, e2 := base64.StdEncoding.DecodeString(edit.AfterB64)
		anchor, e3 := base64.StdEncoding.DecodeString(edit.AnchorB64)
		if e1 != nil || e2 != nil || e3 != nil || len(before) == 0 || len(after) == 0 || len(anchor) == 0 || bytes.Equal(before, after) {
			return nil, fmt.Errorf("invalid before/after/anchor bytes")
		}
		if edit.AnchorCount != 1 || edit.AnchorOffset < 0 || edit.AnchorOffset > len(anchor)-len(before) || !bytes.Equal(anchor[edit.AnchorOffset:edit.AnchorOffset+len(before)], before) {
			return nil, fmt.Errorf("anchor must declare one exact match and contain old bytes at its offset")
		}
		if phase34CountOccurrences(old, anchor) != 1 {
			return nil, fmt.Errorf("anchor must occur exactly once in bound old C bytes")
		}
		start, end := edit.Offset, edit.Offset+len(before)
		if start < 0 || start > len(result)-len(before) {
			return nil, fmt.Errorf("reviewed byte offset is outside bound old C bytes")
		}
		anchorStart := start - edit.AnchorOffset
		if anchorStart < 0 || anchorStart > len(old)-len(anchor) || !bytes.Equal(old[anchorStart:anchorStart+len(anchor)], anchor) {
			return nil, fmt.Errorf("reviewed byte offset does not point to the unique declared anchor")
		}
		if end > lastStart || !bytes.Equal(result[start:end], before) {
			return nil, fmt.Errorf("reviewed byte edits overlap or old bytes differ at the declared offset")
		}
		result = append(append(append([]byte(nil), result[:start]...), after...), result[end:]...)
		lastStart = start
	}
	return result, nil
}

func phase34CountOccurrences(data, pattern []byte) int {
	if len(pattern) == 0 {
		return 0
	}
	count := 0
	for offset := 0; offset+len(pattern) <= len(data); offset++ {
		if bytes.Equal(data[offset:offset+len(pattern)], pattern) {
			count++
		}
	}
	return count
}

func phase34CompareEmission(problems *[]string, path, name string, want, got phase34Emission, sourceSHA, coreSHA string) {
	if phase34Phase371Projection(problems, path, name, sourceSHA, coreSHA, got) {
		return
	}
	if want.Outcome != got.Outcome {
		*problems = append(*problems, fmt.Sprintf("%s: %s outcome changed: got %q want %q", path, name, got.Outcome, want.Outcome))
		return
	}
	if binding, ok := phase34AdlerRewriteCRow(path, name); want.Outcome == "emitted" && ok {
		// This exact-row adapter accepts only the four frozen Phase 34 outputs
		// whose imported Adler implementation changed in D-20. Each source and
		// current output is hash-bound; the imported implementation hash and
		// its direct C projection are also required in the Phase 35 delta.
		source, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if err != nil {
			*problems = append(*problems, fmt.Sprintf("%s: read exact Adler adapter source: %v", path, err))
			return
		}
		if fmt.Sprintf("%x", sha256.Sum256(source)) != binding.sourceSHA {
			*problems = append(*problems, fmt.Sprintf("%s: exact Adler adapter source hash changed", path))
		}
		beforeDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(want.CSource)))
		if len(want.CSource) != binding.beforeBytes || beforeDigest != binding.beforeSHA {
			*problems = append(*problems, fmt.Sprintf("%s: frozen %s C bytes changed (got %d/%s want %d/%s)", path, name, len(want.CSource), beforeDigest, binding.beforeBytes, binding.beforeSHA))
		}
		gotDigest := fmt.Sprintf("%x", sha256.Sum256([]byte(got.CSource)))
		if len(got.CSource) != binding.outputBytes || gotDigest != binding.outputSHA {
			*problems = append(*problems, fmt.Sprintf("%s: %s C byte mismatch in exact Adler adapter (got %d/%s want %d/%s)", path, name, len(got.CSource), gotDigest, binding.outputBytes, binding.outputSHA))
		}
		if !phase34AdlerCompatibilityDeltaBound() {
			*problems = append(*problems, "Phase 35 Adler source/direct-output hashes no longer bind the compatibility adapter")
		}
		return
	}
	if want.Outcome == "not_emitted" {
		// Preserve the explicit non-emission outcome while leaving its failed
		// diagnostic reason, code and presentation as historical facts rather
		// than treating them as C identity.
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

type phase34AdlerCBinding struct {
	sourceSHA   string
	beforeBytes int
	beforeSHA   string
	outputBytes int
	outputSHA   string
}

func phase34AdlerRewriteCRow(path, emitter string) (phase34AdlerCBinding, bool) {
	if emitter == "control_native_c" && path == "examples/checksum/adler32.schway" {
		return phase34AdlerCBinding{sourceSHA: phase35AdlerSourceSHA, beforeBytes: 44595, beforeSHA: "d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b", outputBytes: 19163, outputSHA: phase35AdlerOutputSHA}, true
	}
	if emitter != "emit_application" {
		return phase34AdlerCBinding{}, false
	}
	switch path {
	case "examples/checksum/adler32_cli.schway":
		return phase34AdlerCBinding{sourceSHA: "62478d7b2ab4e777f6adfeab1c9a4e79c230178d4e1c9ca9a41f8c45fad6aa7f", beforeBytes: 19351, beforeSHA: "0bf9cda02e6806c1aeadd4912fabdb46a1856e70ae2a69ba0a9071db86ebc350", outputBytes: 14911, outputSHA: "80df93af9f946d5db3fc520c412f7de31cec28efa7c658ce9a54dd6080950719"}, true
	case "examples/checksum/adler_low.schway":
		return phase34AdlerCBinding{sourceSHA: "42a38540c3c282ab7089ee8f2ccc63c23d039879e19f9ddf5c6ffff9bcfe6987", beforeBytes: 21511, beforeSHA: "34dcb4f59f20e2234ddda81654da483c1d9f9e89a18b2567d43b5fec5b6af234", outputBytes: 17071, outputSHA: "31b6410aa3ca9f43bda6db85b6aebd004d2abb30cea03df149be314753698495"}, true
	case "examples/checksum/checksum.schway":
		return phase34AdlerCBinding{sourceSHA: "da61026ce0f4a6296603e359244d526bef31db230d7339fe04608ae6e5216518", beforeBytes: 19598, beforeSHA: "9bb3babccd3d9a88b36bdedd148d0b548fe142128f2367e1a33f806d2f298751", outputBytes: 15158, outputSHA: "baef335d68fca5df6ed61c63fb6b6248ebbae04a38cc83937cdc19c4a0cdac9e"}, true
	default:
		return phase34AdlerCBinding{}, false
	}
}

func phase34AdlerCompatibilityDeltaBound() bool {
	delta, err := os.ReadFile(testsupport.ProjectPath(".planning", "phases", "35-bitwise-operators-literal-shifts-and-hex-literals", "35-COMPATIBILITY-DELTA.md"))
	if err != nil {
		return false
	}
	text := string(delta)
	sourceRow := "| source_sha256 | `sha256:317ac13972a53eccd04a7ab84cc914aff755e77b881db6fcaf77fa5240ebcb4b` | `sha256:" + phase35AdlerSourceSHA + "` |"
	outputRow := "| `emit` | emitted | 44595 | `sha256:d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b` | emitted | 19163 | `sha256:" + phase35AdlerOutputSHA + "` |"
	return strings.Contains(text, sourceRow) && strings.Contains(text, outputRow)
}

func phase34Contains(paths []string, target string) bool {
	for _, path := range paths {
		if path == target {
			return true
		}
	}
	return false
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
	frozen := phase34EntryPaths(corpus.Entries)
	ledger := phase34LoadNewFixtureLedger(t)
	phase37Ledger := phase34LoadPhase37FixtureLedger(t)
	additions := []string{"examples/bitwise.schway", "examples/crc32/crc32.schway", "examples/crc32/crc_low.schway", "examples/crc32/lib/crc32.schway", "examples/helper_calls/crc_step.schway", "examples/helper_calls/importer.schway", "examples/helper_calls/public.schway"}
	for _, control := range []struct {
		name   string
		mutate func(*phase34NewFixtureLedger, *[]string)
		want   string
	}{
		{"schema", func(l *phase34NewFixtureLedger, _ *[]string) { l.Schema = "invalid" }, "fixture ledger schema"},
		{"missing-row", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries = l.Entries[1:] }, "tracked source is neither frozen nor explicitly ledgered"},
		{"extra-row", func(l *phase34NewFixtureLedger, _ *[]string) {
			l.Entries = append(l.Entries, phase34NewFixtureRow{Path: "examples/undeclared.schway", SourceSHA: l.Entries[0].SourceSHA, Reason: "control"})
		}, "exactly the three declared helper paths"},
		{"malformed-hash", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[0].SourceSHA = "bad" }, "source hash is malformed"},
		{"wrong-hash", func(l *phase34NewFixtureLedger, _ *[]string) {
			l.Entries[0].SourceSHA = "sha256:" + strings.Repeat("0", 64)
		}, "source hash differs from the working tree"},
		{"missing-reason", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[0].Reason = " " }, "requires path, source hash, and reason"},
		{"old-exemption", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[0].Path = frozen[0] }, "old corpus source cannot be exempted"},
		{"phase35-exemption", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[0].Path = "examples/bitwise.schway" }, "old corpus source cannot be exempted"},
		{"unsafe-path", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[0].Path = "../outside.schway" }, "unsafe new-fixture ledger path"},
		{"duplicate-row", func(l *phase34NewFixtureLedger, _ *[]string) { l.Entries[1] = l.Entries[0] }, "paths must be sorted and unique"},
		{"missing-source", func(_ *phase34NewFixtureLedger, a *[]string) { *a = (*a)[:len(*a)-1] }, "source is missing from tracked inventory"},
		{"missing-phase35-source", func(_ *phase34NewFixtureLedger, a *[]string) { *a = (*a)[1:] }, "examples/bitwise.schway: frozen or ledgered source is missing"},
		{"extra-source", func(_ *phase34NewFixtureLedger, a *[]string) { *a = append(*a, "examples/undeclared.schway") }, "tracked source is neither frozen nor explicitly ledgered"},
	} {
		t.Run("additions/"+control.name, func(t *testing.T) {
			if problems := phase34AdditionProblems(frozen, ledger, phase37Ledger, additions); len(problems) != 0 {
				t.Fatalf("clean addition closure: %v", problems)
			}
			changed := phase34NewFixtureLedger{Schema: ledger.Schema, Entries: append([]phase34NewFixtureRow(nil), ledger.Entries...)}
			changedAdditions := append([]string(nil), additions...)
			control.mutate(&changed, &changedAdditions)
			problems := phase34AdditionProblems(frozen, changed, phase37Ledger, changedAdditions)
			if !strings.Contains(strings.Join(problems, "; "), control.want) {
				t.Fatalf("addition control did not reach %q: %v", control.want, problems)
			}
			if problems := phase34AdditionProblems(frozen, ledger, phase37Ledger, additions); len(problems) != 0 {
				t.Fatalf("restored addition closure: %v", problems)
			}
		})
	}
	observed := phase34ObserveCorpus(t)
	if problems := phase34CompareCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("clean pre-change observation must pass before mutation: %s", strings.Join(problems, "; "))
	}
	byteMutationCaught := false
byteMutation:
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
					byteMutationCaught = true
					break byteMutation
				}
			}
			t.Fatalf("reached mutation was not named by the comparator; got %v", problems)
		}
	}
	if !byteMutationCaught {
		t.Fatal("no successful literal C emission was available for the comparator mutation")
	}

	statusMutationCaught := false
statusMutation:
	for index := range observed {
		for _, emitter := range []struct {
			name string
			get  func(*phase34PreChangeEntry) *phase34Emission
		}{
			{"emit_application", func(entry *phase34PreChangeEntry) *phase34Emission { return &entry.EmitApplication }},
			{"control_native_c", func(entry *phase34PreChangeEntry) *phase34Emission { return &entry.ControlNativeC }},
		} {
			result := emitter.get(&observed[index])
			if result.Outcome != "not_emitted" {
				continue
			}
			original := result.Outcome
			result.Outcome = "emitted"
			problems := phase34CompareCorpus(corpus, observed)
			result.Outcome = original
			if clean := phase34CompareCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored non-emitting observation did not pass: %s", strings.Join(clean, "; "))
			}
			needle := observed[index].Path + ": " + emitter.name + " outcome changed"
			for _, problem := range problems {
				if strings.HasPrefix(problem, needle) {
					statusMutationCaught = true
					break statusMutation
				}
			}
			t.Fatalf("non-emission mutation was not named by the comparator; got %v", problems)
		}
	}
	if !statusMutationCaught {
		t.Fatal("no non-emitting result was available for the explicit outcome mutation")
	}
}

func TestPhase37Phase34AdditionLedgerControls(t *testing.T) {
	corpus := phase34LoadCorpus(t)
	frozen := phase34EntryPaths(corpus.Entries)
	phase36 := phase34LoadNewFixtureLedger(t)
	phase37 := phase34LoadPhase37FixtureLedger(t)
	phase34TrackedSchwayPaths(t) // The clean control uses the actual staged git inventory.
	var additions []string
	for _, path := range phase34CurrentTrackedSchwayPaths(t) {
		if !phase34Contains(frozen, path) {
			additions = append(additions, path)
		}
	}
	if problems := phase34AdditionProblems(frozen, phase36, phase37, additions); len(problems) != 0 {
		t.Fatalf("clean Phase37 addition closure: %v", problems)
	}
	for _, control := range []struct {
		name   string
		mutate func(*phase34Phase37FixtureLedger, *[]string)
		want   string
	}{
		{"omitted-ledger-row", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries = l.Entries[:len(l.Entries)-1] }, "tracked source is neither frozen nor explicitly ledgered"},
		{"omitted-tracked-source", func(_ *phase34Phase37FixtureLedger, paths *[]string) { *paths = (*paths)[1:] }, "frozen or ledgered source is missing from tracked inventory"},
		{"unknown-extra-tracked-source", func(_ *phase34Phase37FixtureLedger, paths *[]string) {
			*paths = append(*paths, "examples/phase37/undeclared.schway")
		}, "tracked source is neither frozen nor explicitly ledgered"},
		{"malformed-hash", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[0].SourceSHA = "bad" }, "Phase37 source hash is malformed"},
		{"wrong-hash", func(l *phase34Phase37FixtureLedger, _ *[]string) {
			l.Entries[0].SourceSHA = "sha256:" + strings.Repeat("0", 64)
		}, "Phase37 source hash differs from the working tree"},
		{"blank-reason", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[0].Reason = "  \t" }, "Phase37 ledger requires a nonblank reason"},
		{"duplicate-row", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[1] = l.Entries[0] }, "Phase37 ledger paths must be sorted and unique"},
		{"unsorted-rows", func(l *phase34Phase37FixtureLedger, _ *[]string) {
			l.Entries[0], l.Entries[1] = l.Entries[1], l.Entries[0]
		}, "Phase37 ledger must contain exactly the three CRC source paths in order"},
		{"unsafe-path", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = "../outside.schway" }, "unsafe Phase37 new-fixture ledger path"},
		{"old-exemption", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = frozen[0] }, "old corpus source cannot be exempted"},
		{"phase35-exemption", func(l *phase34Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = "examples/bitwise.schway" }, "old corpus source cannot be exempted"},
		{"phase36-exemption", func(l *phase34Phase37FixtureLedger, _ *[]string) {
			l.Entries[0].Path = "examples/helper_calls/crc_step.schway"
		}, "old corpus source cannot be exempted"},
	} {
		t.Run(control.name, func(t *testing.T) {
			changed := phase34Phase37FixtureLedger{Schema: phase37.Schema, Entries: append([]phase34NewFixtureRow(nil), phase37.Entries...)}
			changedAdditions := append([]string(nil), additions...)
			control.mutate(&changed, &changedAdditions)
			problems := phase34AdditionProblems(frozen, phase36, changed, changedAdditions)
			if !strings.Contains(strings.Join(problems, "; "), control.want) {
				t.Fatalf("control did not reach %q: %v", control.want, problems)
			}
			if clean := phase34AdditionProblems(frozen, phase36, phase37, additions); len(clean) != 0 {
				t.Fatalf("restored Phase37 addition closure: %v", clean)
			}
		})
	}
}
