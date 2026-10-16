package session

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
	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase36PreChangeSchema = "phase36.prechange-corpus/1"
const phase36LedgerSchema = "phase36.new-fixture-ledger/1"
const phase37LedgerSchema = "phase37.new-fixture-ledger/1"

var phase36ExpectedLedger = map[string]string{
	"examples/helper_calls/crc_step.schway": "sha256:0242b3a253a80d8818310897220494722eda3029e12f67baa34cfeb529a50d0f",
	"examples/helper_calls/importer.schway": "sha256:da4db4e265aae96bf90ab17c5c4b064500115a93bb5095a192334be53a0bc5ab",
	"examples/helper_calls/public.schway":   "sha256:ef608e10f62d79e731bbf3e6a2320fa567914c2e5e48e13807cd05535d7a4f38",
}

var phase37ExpectedLedgerPaths = []string{
	"examples/crc32/crc32.schway",
	"examples/crc32/crc_low.schway",
	"examples/crc32/lib/crc32.schway",
}

type phase36PreChangeCorpus struct {
	Schema      string                  `json:"schema"`
	Revision    string                  `json:"revision"`
	CoreSchema  string                  `json:"core_schema"`
	CoreSchema1 string                  `json:"core_schema_1"`
	CGenSchema  string                  `json:"cgen_schema"`
	Entries     []phase36PreChangeEntry `json:"entries"`
}

type phase36PreChangeEntry struct {
	Path             string                `json:"path"`
	Source           string                `json:"source"`
	SourceSHA        string                `json:"source_sha256"`
	Check            phase36Admission      `json:"check"`
	Modules          []phase36ModuleDigest `json:"modules,omitempty"`
	CoreSchema       string                `json:"checked_core_schema,omitempty"`
	CoreSHA          string                `json:"checked_core_sha256,omitempty"`
	SemanticIDs      []string              `json:"semantic_ids,omitempty"`
	Emit             phase36Emission       `json:"emit"`
	EmitNative       phase36Emission       `json:"emit_native"`
	Application      phase36Emission       `json:"emit_application"`
	checkedCoreBytes []byte                `json:"-"`
}

type phase36ModuleDigest struct {
	Module     string `json:"module"`
	Path       string `json:"path"`
	SourcePath string `json:"source_path"`
	Digest     string `json:"digest"`
}

type phase36Admission struct {
	Outcome     string              `json:"outcome"`
	Error       string              `json:"error,omitempty"`
	Diagnostics []phase36Diagnostic `json:"diagnostics,omitempty"`
}

type phase36Diagnostic struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Start   int    `json:"start"`
	End     int    `json:"end"`
}

type phase36Emission struct {
	Outcome string `json:"outcome"`
	Schema  string `json:"schema"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	Reason  string `json:"reason,omitempty"`
	CSource string `json:"c_source,omitempty"`
}

type phase36FixtureLedger struct {
	Schema  string                    `json:"schema"`
	Entries []phase36FixtureLedgerRow `json:"entries"`
}

type phase36Phase37FixtureLedger struct {
	Schema  string                    `json:"schema"`
	Entries []phase36FixtureLedgerRow `json:"entries"`
}

type phase36FixtureLedgerRow struct {
	Path      string `json:"path"`
	SourceSHA string `json:"source_sha256"`
	Reason    string `json:"reason"`
}

type phase36RouteAccessor struct {
	name string
	get  func(*phase36PreChangeEntry) *phase36Emission
}

var phase36RouteAccessors = []phase36RouteAccessor{
	{"emit", func(entry *phase36PreChangeEntry) *phase36Emission { return &entry.Emit }},
	{"emit_native", func(entry *phase36PreChangeEntry) *phase36Emission { return &entry.EmitNative }},
	{"emit_application", func(entry *phase36PreChangeEntry) *phase36Emission { return &entry.Application }},
}

func TestPhase36NoCallGeneratedCByteBaseline(t *testing.T) {
	corpus := phase36LoadPreChangeCorpus(t)
	ledger := phase36LoadFixtureLedger(t)
	phase37Ledger := phase36LoadPhase37FixtureLedger(t)
	phase36VerifyRevision(t, testsupport.ProjectPath("."), corpus)
	if problems := phase36InventoryProblems(corpus, ledger, phase37Ledger, phase36TrackedSources(t)); len(problems) != 0 {
		t.Fatalf("tracked source inventory is not closed: %s", strings.Join(problems, "; "))
	}
	observed := phase36ObservePreChangeCorpus(t, corpus)
	if problems := phase36ComparePreChangeCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("pre-change emitter corpus mismatch: %s", strings.Join(problems, "; "))
	}
	phase36RequireOutcomeCoverage(t, corpus)
	phase36RequireModuleCoverage(t, corpus)
}

func TestPhase36NoCallGeneratedCByteBaselineControls(t *testing.T) {
	corpus := phase36LoadPreChangeCorpus(t)
	ledger := phase36LoadFixtureLedger(t)
	phase37Ledger := phase36LoadPhase37FixtureLedger(t)
	phase36VerifyRevision(t, testsupport.ProjectPath("."), corpus)
	if problems := phase36InventoryProblems(corpus, ledger, phase37Ledger, phase36TrackedSources(t)); len(problems) != 0 {
		t.Fatalf("tracked source inventory is not closed before controls: %s", strings.Join(problems, "; "))
	}
	observed := phase36ObservePreChangeCorpus(t, corpus)
	if problems := phase36ComparePreChangeCorpus(corpus, observed); len(problems) != 0 {
		t.Fatalf("clean pre-change observation must pass before mutation: %s", strings.Join(problems, "; "))
	}
	if problems := phase36ComparePreChangeCorpus(corpus, phase36CloneEntries(observed)); len(problems) != 0 {
		t.Fatalf("no-op observation must pass: %s", strings.Join(problems, "; "))
	}

	byteMutationCaught := false
	for index := range observed {
		for _, route := range phase36RouteAccessors {
			result := route.get(&observed[index])
			if result.Outcome != "emitted" || result.CSource == "" {
				continue
			}
			mutant := phase36CloneEntries(observed)
			mutated := route.get(&mutant[index])
			cBytes := []byte(mutated.CSource)
			cBytes[0] ^= 1
			mutated.CSource = string(cBytes)
			problems := phase36ComparePreChangeCorpus(corpus, mutant)
			want := observed[index].Path + ": " + route.name + " C byte mismatch"
			if !phase36HasProblem(problems, want) {
				t.Fatalf("changed C byte was not attributed to %s: %v", want, problems)
			}
			if clean := phase36ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
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

	missing := phase36CloneEntries(observed)
	missingPath := missing[len(missing)-1].Path
	missing = missing[:len(missing)-1]
	if problems := phase36ComparePreChangeCorpus(corpus, missing); !phase36HasProblem(problems, missingPath+": source row missing from observation") {
		t.Fatalf("omitted source row was not attributed to %s: %v", missingPath, problems)
	}
	if clean := phase36ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
		t.Fatalf("restored source inventory did not pass: %s", strings.Join(clean, "; "))
	}

	falseExemptionCaught := false
	for index := range observed {
		for _, route := range phase36RouteAccessors {
			if route.get(&observed[index]).Outcome != "emitted" {
				continue
			}
			mutant := phase36CloneEntries(observed)
			*route.get(&mutant[index]) = phase36NotEmitted("misclassified as exempt")
			problems := phase36ComparePreChangeCorpus(corpus, mutant)
			want := observed[index].Path + ": " + route.name + " outcome changed"
			if !phase36HasProblem(problems, want) {
				t.Fatalf("false route exemption was not attributed to %s: %v", want, problems)
			}
			if clean := phase36ComparePreChangeCorpus(corpus, observed); len(clean) != 0 {
				t.Fatalf("restored route observation did not pass: %s", strings.Join(clean, "; "))
			}
			falseExemptionCaught = true
			break
		}
		if falseExemptionCaught {
			break
		}
	}
	if !falseExemptionCaught {
		t.Fatal("no admitted emitting route was available for the false-exemption control")
	}

	old := corpus.Entries[0]
	mutatedLedger := phase36FixtureLedger{Schema: phase36LedgerSchema, Entries: []phase36FixtureLedgerRow{{Path: old.Path, SourceSHA: old.SourceSHA, Reason: "control: attempt to exempt an old source"}}}
	if problems := phase36InventoryProblems(corpus, mutatedLedger, phase37Ledger, phase36TrackedSources(t)); !phase36HasProblem(problems, old.Path+": old corpus source cannot be exempted by the new-fixture ledger") {
		t.Fatalf("old source false exemption was not rejected: %v", problems)
	}
	unsafeLedger := phase36FixtureLedger{Schema: phase36LedgerSchema, Entries: []phase36FixtureLedgerRow{{Path: "../outside.schway", SourceSHA: "sha256:bad", Reason: "control: traversal"}}}
	if problems := phase36InventoryProblems(corpus, unsafeLedger, phase37Ledger, phase36TrackedSources(t)); !phase36HasProblem(problems, "../outside.schway: unsafe new-fixture ledger path") {
		t.Fatalf("unsafe ledger path was not rejected before file access: %v", problems)
	}
	if problems := phase36InventoryProblems(corpus, ledger, phase37Ledger, phase36TrackedSources(t)); len(problems) != 0 {
		t.Fatalf("restored fixture ledger did not pass: %s", strings.Join(problems, "; "))
	}
}

func TestPhase37Phase36AdditionLedgerControls(t *testing.T) {
	corpus := phase36LoadPreChangeCorpus(t)
	phase36 := phase36LoadFixtureLedger(t)
	phase37 := phase36LoadPhase37FixtureLedger(t)
	tracked := phase36TrackedSources(t)
	if problems := phase36InventoryProblems(corpus, phase36, phase37, tracked); len(problems) != 0 {
		t.Fatalf("clean composed source inventory: %v", problems)
	}
	for _, control := range []struct {
		name   string
		mutate func(*phase36Phase37FixtureLedger, *[]string)
		want   string
	}{
		{"omitted-ledger-row", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries = l.Entries[:len(l.Entries)-1] }, "Phase37 ledger must contain exactly the three CRC source paths in order"},
		{"omitted-tracked-source", func(_ *phase36Phase37FixtureLedger, paths *[]string) {
			for index, path := range *paths {
				if path == phase37ExpectedLedgerPaths[0] {
					*paths = append((*paths)[:index], (*paths)[index+1:]...)
					return
				}
			}
		}, phase37ExpectedLedgerPaths[0] + ": frozen or ledgered source is missing from tracked inventory"},
		{"unknown-extra-tracked-source", func(_ *phase36Phase37FixtureLedger, paths *[]string) {
			*paths = append(*paths, "examples/phase37/undeclared.schway")
			sort.Strings(*paths)
		}, "examples/phase37/undeclared.schway: tracked source is neither frozen nor explicitly ledgered"},
		{"malformed-hash", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[0].SourceSHA = "bad" }, phase37ExpectedLedgerPaths[0] + ": Phase37 source hash is malformed"},
		{"wrong-hash", func(l *phase36Phase37FixtureLedger, _ *[]string) {
			l.Entries[0].SourceSHA = "sha256:" + strings.Repeat("0", 64)
		}, phase37ExpectedLedgerPaths[0] + ": Phase37 source hash differs from the working tree"},
		{"blank-reason", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[0].Reason = " \t" }, phase37ExpectedLedgerPaths[0] + ": Phase37 ledger requires a nonblank reason"},
		{"duplicate-row", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[1] = l.Entries[0] }, "Phase37 ledger paths must be sorted and unique"},
		{"unsorted-rows", func(l *phase36Phase37FixtureLedger, _ *[]string) {
			l.Entries[0], l.Entries[1] = l.Entries[1], l.Entries[0]
		}, "Phase37 ledger paths must be sorted and unique"},
		{"unsafe-path", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = "../outside.schway" }, "../outside.schway: unsafe Phase37 new-fixture ledger path"},
		{"old-exemption", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = corpus.Entries[0].Path }, corpus.Entries[0].Path + ": old corpus source cannot be exempted by the Phase37 ledger"},
		{"phase35-exemption", func(l *phase36Phase37FixtureLedger, _ *[]string) { l.Entries[0].Path = "examples/bitwise.schway" }, "examples/bitwise.schway: old corpus source cannot be exempted by the Phase37 ledger"},
		{"phase36-exemption", func(l *phase36Phase37FixtureLedger, _ *[]string) {
			l.Entries[0].Path = "examples/helper_calls/crc_step.schway"
		}, "examples/helper_calls/crc_step.schway: Phase36 source cannot be exempted by the Phase37 ledger"},
	} {
		t.Run(control.name, func(t *testing.T) {
			changed := phase36Phase37FixtureLedger{Schema: phase37.Schema, Entries: append([]phase36FixtureLedgerRow(nil), phase37.Entries...)}
			changedTracked := append([]string(nil), tracked...)
			control.mutate(&changed, &changedTracked)
			problems := phase36InventoryProblems(corpus, phase36, changed, changedTracked)
			if !strings.Contains(strings.Join(problems, "; "), control.want) {
				t.Fatalf("control did not reach %q: %v", control.want, problems)
			}
			if clean := phase36InventoryProblems(corpus, phase36, phase37, tracked); len(clean) != 0 {
				t.Fatalf("restored composed source inventory: %v", clean)
			}
		})
	}
}

func phase36ObservePreChangeCorpus(t *testing.T, corpus phase36PreChangeCorpus) []phase36PreChangeEntry {
	t.Helper()
	if corpus.Schema != phase36PreChangeSchema {
		t.Fatalf("pre-change fixture schema is %q, want %q", corpus.Schema, phase36PreChangeSchema)
	}
	if corpus.CoreSchema != core.Schema || corpus.CoreSchema1 != core.Schema1 || corpus.CGenSchema != cgen.Schema {
		t.Fatalf("pre-change fixture schema identifiers changed: core=%q core1=%q cgen=%q", corpus.CoreSchema, corpus.CoreSchema1, corpus.CGenSchema)
	}
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

	observed := make([]phase36PreChangeEntry, 0, len(corpus.Entries))
	for _, before := range corpus.Entries {
		path := filepath.Join(sourceRoot, filepath.FromSlash(before.Path))
		checked, _, modules, checkErr := checkSourcePathModules(path)
		entry := phase36PreChangeEntry{Path: before.Path, Source: before.Source, SourceSHA: phase36Digest([]byte(before.Source))}
		entry.Check = phase36AdmissionResult(checked.Diagnostics, checkErr)
		entry.Modules = phase36ModuleRows(before.Path, modules)
		if checkErr != nil {
			reason := "CheckFile error: " + checkErr.Error()
			entry.Emit = phase36NotEmitted(reason)
			entry.EmitNative = phase36NotEmitted(reason)
			entry.Application = phase36NotEmitted(reason)
		} else if len(checked.Diagnostics) != 0 {
			reason := "CheckFile diagnostics"
			if codes := phase36DiagnosticCodes(checked.Diagnostics); len(codes) != 0 {
				reason += ": " + strings.Join(codes, ",")
			}
			entry.Emit = phase36NotEmitted(reason)
			entry.EmitNative = phase36NotEmitted(reason)
			entry.Application = phase36NotEmitted(reason)
		} else {
			coreBytes, marshalErr := json.Marshal(checked.Program)
			if marshalErr != nil {
				t.Fatalf("marshal checked core for %s: %v", before.Path, marshalErr)
			}
			entry.CoreSchema = checked.Program.Schema
			entry.checkedCoreBytes = append([]byte(nil), coreBytes...)
			entry.CoreSHA = phase36Digest(coreBytes)
			entry.SemanticIDs = phase36SemanticIDs(coreBytes)
			entry.Emit = phase36EmitResult(cgen.Emit(checked.Program))
			entry.EmitNative = phase36EmitResult(cgen.EmitNative(checked.Program))
			entry.Application = phase36EmitResult(cgen.EmitApplication(checked.Program))
		}
		observed = append(observed, entry)
	}
	return observed
}

func phase36VerifyRevision(t *testing.T, root string, corpus phase36PreChangeCorpus) {
	t.Helper()
	tree := exec.Command("git", "ls-tree", "-r", "-z", "--name-only", corpus.Revision)
	tree.Dir = root
	pathsOutput, err := tree.Output()
	if err != nil {
		t.Fatalf("git ls-tree frozen revision %s: %v", corpus.Revision, err)
	}
	var paths []string
	for _, path := range bytes.Split(pathsOutput, []byte{0}) {
		if len(path) != 0 && strings.HasSuffix(string(path), ".schway") {
			paths = append(paths, filepath.ToSlash(string(path)))
		}
	}
	sort.Strings(paths)
	fixturePaths := make([]string, len(corpus.Entries))
	for index, entry := range corpus.Entries {
		fixturePaths[index] = entry.Path
	}
	if !phase36EqualStrings(paths, fixturePaths) {
		t.Fatalf("fixture inventory does not equal .schway inventory at %s: revision=%d fixture=%d", corpus.Revision, len(paths), len(fixturePaths))
	}
	for _, entry := range corpus.Entries {
		cmd := exec.Command("git", "show", corpus.Revision+":"+entry.Path)
		cmd.Dir = root
		source, err := cmd.Output()
		if err != nil {
			t.Fatalf("read frozen source %s at %s: %v", entry.Path, corpus.Revision, err)
		}
		if !bytes.Equal(source, []byte(entry.Source)) || phase36Digest(source) != entry.SourceSHA {
			t.Fatalf("fixture source bytes/hash differ from frozen revision for %s", entry.Path)
		}
	}
}

func phase36TrackedSources(t *testing.T) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "-z", "--", "*.schway")
	cmd.Dir = testsupport.ProjectPath(".")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("list current tracked Schway sources: %v", err)
	}
	var paths []string
	for _, path := range bytes.Split(output, []byte{0}) {
		if len(path) != 0 {
			paths = append(paths, filepath.ToSlash(string(path)))
		}
	}
	sort.Strings(paths)
	return paths
}

func phase36InventoryProblems(corpus phase36PreChangeCorpus, ledger phase36FixtureLedger, phase37 phase36Phase37FixtureLedger, tracked []string) []string {
	var problems []string
	if ledger.Schema != phase36LedgerSchema {
		problems = append(problems, fmt.Sprintf("fixture ledger schema is %q, want %q", ledger.Schema, phase36LedgerSchema))
	}
	if phase37.Schema != phase37LedgerSchema {
		problems = append(problems, fmt.Sprintf("Phase37 fixture ledger schema is %q, want %q", phase37.Schema, phase37LedgerSchema))
	}
	base := make(map[string]phase36PreChangeEntry, len(corpus.Entries))
	for _, entry := range corpus.Entries {
		if _, duplicate := base[entry.Path]; duplicate {
			problems = append(problems, entry.Path+": duplicate old corpus source row")
		}
		base[entry.Path] = entry
	}
	phase36Allowed := []string{
		"examples/helper_calls/crc_step.schway",
		"examples/helper_calls/importer.schway",
		"examples/helper_calls/public.schway",
	}
	ledgerPaths := make([]string, 0, len(ledger.Entries))
	for _, row := range ledger.Entries {
		ledgerPaths = append(ledgerPaths, row.Path)
		if _, old := base[row.Path]; old {
			problems = append(problems, row.Path+": old corpus source cannot be exempted by the new-fixture ledger")
			continue
		}
		if !phase36ContainsString(phase36Allowed, row.Path) {
			problems = append(problems, row.Path+": unsafe new-fixture ledger path")
			continue
		}
		cleanPath := filepath.Clean(filepath.FromSlash(row.Path))
		if row.Path == "" || filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || filepath.ToSlash(cleanPath) != row.Path {
			problems = append(problems, row.Path+": unsafe new-fixture ledger path")
			continue
		}
		if row.Path == "" || row.SourceSHA == "" || strings.TrimSpace(row.Reason) == "" {
			problems = append(problems, row.Path+": new-fixture ledger row requires path, source hash, and reason")
		}
		if len(row.SourceSHA) != len("sha256:")+64 || !strings.HasPrefix(row.SourceSHA, "sha256:") || strings.Trim(row.SourceSHA[len("sha256:"):], "0123456789abcdef") != "" {
			problems = append(problems, row.Path+": new-fixture ledger source hash is malformed")
		} else if row.SourceSHA != phase36ExpectedLedger[row.Path] {
			problems = append(problems, row.Path+": Phase36 fixture ledger pinned source hash changed")
		}
		if source, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(row.Path))); err != nil {
			problems = append(problems, row.Path+": new-fixture ledger path is missing from the working tree")
		} else if phase36Digest(source) != row.SourceSHA {
			problems = append(problems, row.Path+": new-fixture ledger source hash differs from the working tree")
		}
	}
	if !sort.StringsAreSorted(ledgerPaths) || phase36HasDuplicates(ledgerPaths) {
		problems = append(problems, "new-fixture ledger paths must be sorted and unique")
	}
	if !phase36EqualStrings(ledgerPaths, phase36Allowed) {
		problems = append(problems, "Phase36 ledger must contain exactly its three declared helper paths in order")
	}

	phase37Paths := make([]string, 0, len(phase37.Entries))
	phase37Seen := make(map[string]bool, len(phase37.Entries))
	for _, row := range phase37.Entries {
		phase37Paths = append(phase37Paths, row.Path)
		if _, old := base[row.Path]; old {
			problems = append(problems, row.Path+": old corpus source cannot be exempted by the Phase37 ledger")
			continue
		}
		if phase36ContainsString(phase36Allowed, row.Path) {
			problems = append(problems, row.Path+": Phase36 source cannot be exempted by the Phase37 ledger")
			continue
		}
		if !phase36ContainsString(phase37ExpectedLedgerPaths, row.Path) {
			problems = append(problems, row.Path+": unsafe Phase37 new-fixture ledger path")
			continue
		}
		cleanPath := filepath.Clean(filepath.FromSlash(row.Path))
		if row.Path == "" || filepath.IsAbs(cleanPath) || cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) || filepath.ToSlash(cleanPath) != row.Path {
			problems = append(problems, row.Path+": unsafe Phase37 new-fixture ledger path")
			continue
		}
		if phase37Seen[row.Path] {
			problems = append(problems, "Phase37 ledger paths must be sorted and unique")
		}
		phase37Seen[row.Path] = true
		if strings.TrimSpace(row.Reason) == "" {
			problems = append(problems, row.Path+": Phase37 ledger requires a nonblank reason")
		}
		if len(row.SourceSHA) != len("sha256:")+64 || !strings.HasPrefix(row.SourceSHA, "sha256:") || strings.Trim(row.SourceSHA[len("sha256:"):], "0123456789abcdef") != "" {
			problems = append(problems, row.Path+": Phase37 source hash is malformed")
		}
		if source, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(row.Path))); err != nil {
			problems = append(problems, row.Path+": Phase37 source is missing from the working tree")
		} else if phase36Digest(source) != row.SourceSHA {
			problems = append(problems, row.Path+": Phase37 source hash differs from the working tree")
		}
	}
	if !sort.StringsAreSorted(phase37Paths) || phase36HasDuplicates(phase37Paths) {
		problems = append(problems, "Phase37 ledger paths must be sorted and unique")
	}
	if !phase36EqualStrings(phase37Paths, phase37ExpectedLedgerPaths) {
		problems = append(problems, "Phase37 ledger must contain exactly the three CRC source paths in order")
	}

	allowed := make(map[string]bool, len(base)+len(phase36Allowed)+len(phase37ExpectedLedgerPaths))
	for path := range base {
		allowed[path] = true
	}
	for _, path := range phase36Allowed {
		allowed[path] = true
	}
	for _, path := range phase37ExpectedLedgerPaths {
		allowed[path] = true
	}
	if !sort.StringsAreSorted(tracked) || phase36HasDuplicates(tracked) {
		problems = append(problems, "tracked source inventory must be sorted and unique")
	}
	for _, path := range tracked {
		if !allowed[path] {
			problems = append(problems, path+": tracked source is neither frozen nor explicitly ledgered")
		}
	}
	trackedSet := make(map[string]bool, len(tracked))
	for _, path := range tracked {
		trackedSet[path] = true
	}
	for path := range allowed {
		if !trackedSet[path] {
			problems = append(problems, path+": frozen or ledgered source is missing from tracked inventory")
		}
	}
	return problems
}

func phase36ComparePreChangeCorpus(want phase36PreChangeCorpus, observed []phase36PreChangeEntry) []string {
	var problems []string
	if want.Schema != phase36PreChangeSchema {
		problems = append(problems, fmt.Sprintf("fixture schema changed: got %q want %q", want.Schema, phase36PreChangeSchema))
	}
	paths := make([]string, len(observed))
	byPath := make(map[string]phase36PreChangeEntry, len(observed))
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
	if !phase36EqualStrings(paths, wantPaths) {
		for _, path := range wantPaths {
			if !phase36ContainsString(paths, path) {
				problems = append(problems, path+": source row missing from observation")
			}
		}
		for _, path := range paths {
			if !phase36ContainsString(wantPaths, path) {
				problems = append(problems, path+": unexpected source row in observation")
			}
		}
		if !sort.StringsAreSorted(paths) {
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
		if !reflect.DeepEqual(before.Check, after.Check) {
			problems = append(problems, before.Path+": source check status or diagnostics changed")
		}
		if !reflect.DeepEqual(before.Modules, after.Modules) {
			problems = append(problems, before.Path+": imported module inventory or digest changed")
		}
		if before.CoreSchema != after.CoreSchema {
			problems = append(problems, before.Path+": checked core schema changed")
		}
		if before.CoreSHA != after.CoreSHA {
			problems = append(problems, before.Path+": checked core identity changed")
		}
		if !phase36EqualStrings(before.SemanticIDs, after.SemanticIDs) {
			problems = append(problems, before.Path+": semantic IDs changed")
		}
		for _, route := range phase36RouteAccessors {
			got := *route.get(&after)
			handled := phase371ObserveComparison(
				&problems,
				"phase36",
				before.Path,
				route.name,
				strings.TrimPrefix(after.SourceSHA, "sha256:"),
				strings.TrimPrefix(after.CoreSHA, "sha256:"),
				phase371Observed36(got),
			)
			if !handled {
				phase36CompareEmission(&problems, before.Path, route.name, *route.get(&before), got)
			}
		}
	}
	return problems
}

func phase36RequireModuleCoverage(t *testing.T, corpus phase36PreChangeCorpus) {
	t.Helper()
	byPath := make(map[string]phase36PreChangeEntry, len(corpus.Entries))
	for _, entry := range corpus.Entries {
		byPath[entry.Path] = entry
	}
	for _, entry := range corpus.Entries {
		for _, module := range entry.Modules {
			if module.Module == "" || module.Path == "" || module.SourcePath == "" || module.Digest == "" {
				t.Fatalf("module route is missing identity/path/digest for %s: %+v", entry.Path, module)
			}
			source, ok := byPath[module.SourcePath]
			if !ok {
				t.Fatalf("module source is not an exact tracked corpus row for %s: %s", entry.Path, module.SourcePath)
			}
			if module.Digest != strings.TrimPrefix(source.SourceSHA, "sha256:") {
				t.Fatalf("module digest does not bind exact imported source %s from %s", module.SourcePath, entry.Path)
			}
		}
	}
}

func phase36CompareEmission(problems *[]string, path, route string, want, got phase36Emission) {
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
			index := phase36FirstByteDifference([]byte(want.CSource), []byte(got.CSource))
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

func phase36RequireOutcomeCoverage(t *testing.T, corpus phase36PreChangeCorpus) {
	t.Helper()
	if len(corpus.Entries) == 0 || !sort.SliceIsSorted(corpus.Entries, func(i, j int) bool { return corpus.Entries[i].Path < corpus.Entries[j].Path }) {
		t.Fatal("pre-change fixture must contain a non-empty sorted source inventory")
	}
	emitted := map[string]int{}
	nonEmitted := map[string]int{}
	checked := 0
	for _, entry := range corpus.Entries {
		if entry.Path == "" || entry.SourceSHA == "" || phase36Digest([]byte(entry.Source)) != entry.SourceSHA {
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
		for _, route := range phase36RouteAccessors {
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
	for _, route := range phase36RouteAccessors {
		if emitted[route.name] == 0 || nonEmitted[route.name] == 0 {
			t.Fatalf("%s needs emitted and explicit non-emitting/refused rows, emitted=%d non-emitted=%d", route.name, emitted[route.name], nonEmitted[route.name])
		}
	}
}

func phase36LoadPreChangeCorpus(t *testing.T) phase36PreChangeCorpus {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "prechange-corpus.json"))
	if err != nil {
		t.Fatalf("read Phase 36 pre-change corpus: %v", err)
	}
	var corpus phase36PreChangeCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("decode Phase 36 pre-change corpus: %v", err)
	}
	if corpus.Revision == "" {
		t.Fatal("pre-change corpus is missing its source revision")
	}
	return corpus
}

func phase36LoadFixtureLedger(t *testing.T) phase36FixtureLedger {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "new-fixture-ledger.json"))
	if err != nil {
		t.Fatalf("read Phase 36 intentional new-fixture ledger: %v", err)
	}
	var ledger phase36FixtureLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode Phase 36 intentional new-fixture ledger: %v", err)
	}
	return ledger
}

func phase36LoadPhase37FixtureLedger(t *testing.T) phase36Phase37FixtureLedger {
	t.Helper()
	raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase37", "new-fixture-ledger.json"))
	if err != nil {
		t.Fatalf("read Phase37 additive source ledger: %v", err)
	}
	var ledger phase36Phase37FixtureLedger
	if err := json.Unmarshal(raw, &ledger); err != nil {
		t.Fatalf("decode Phase37 additive source ledger: %v", err)
	}
	return ledger
}

func phase36AdmissionResult(diagnostics []diagnostic.Diagnostic, err error) phase36Admission {
	if err != nil {
		return phase36Admission{Outcome: "error", Error: err.Error()}
	}
	if len(diagnostics) == 0 {
		return phase36Admission{Outcome: "admitted"}
	}
	rows := make([]phase36Diagnostic, 0, len(diagnostics))
	for _, item := range diagnostics {
		rows = append(rows, phase36Diagnostic{Code: item.Code, Message: item.Message, Start: item.Primary.Start, End: item.Primary.End})
	}
	return phase36Admission{Outcome: "diagnostics", Diagnostics: rows}
}

func phase36EmitResult(source string, err error) phase36Emission {
	row := phase36Emission{Schema: cgen.Schema}
	if err != nil {
		row.Outcome = "refused"
		row.Error = err.Error()
		row.Code = phase36ErrorCode(err.Error())
		return row
	}
	row.Outcome = "emitted"
	row.CSource = source
	return row
}

func phase36NotEmitted(reason string) phase36Emission {
	return phase36Emission{Outcome: "not_emitted", Schema: cgen.Schema, Reason: reason}
}

func phase36ModuleRows(entryPath string, modules []native.ModuleDigest) []phase36ModuleDigest {
	if len(modules) == 0 {
		return nil
	}
	rows := make([]phase36ModuleDigest, 0, len(modules))
	for _, module := range modules {
		rel := filepath.Clean(filepath.FromSlash(module.Path))
		if !filepath.IsAbs(rel) && (filepath.Dir(filepath.FromSlash(entryPath)) == "." || !strings.HasPrefix(filepath.ToSlash(rel), filepath.ToSlash(filepath.Dir(filepath.FromSlash(entryPath))+string(filepath.Separator)))) {
			rel = filepath.Join(filepath.Dir(filepath.FromSlash(entryPath)), rel)
		}
		rows = append(rows, phase36ModuleDigest{Module: module.Module, Path: filepath.ToSlash(module.Path), SourcePath: filepath.ToSlash(filepath.Clean(rel)), Digest: module.Digest})
	}
	return rows
}

func phase36SemanticIDs(coreBytes []byte) []string {
	var value any
	if err := json.Unmarshal(coreBytes, &value); err != nil {
		return nil
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

func phase36DiagnosticCodes(diagnostics []diagnostic.Diagnostic) []string {
	codes := make([]string, 0, len(diagnostics))
	for _, item := range diagnostics {
		codes = append(codes, item.Code)
	}
	return codes
}

func phase36CloneEntries(entries []phase36PreChangeEntry) []phase36PreChangeEntry {
	clone := append([]phase36PreChangeEntry(nil), entries...)
	for index := range clone {
		clone[index].SemanticIDs = append([]string(nil), entries[index].SemanticIDs...)
		clone[index].Check.Diagnostics = append([]phase36Diagnostic(nil), entries[index].Check.Diagnostics...)
		clone[index].Modules = append([]phase36ModuleDigest(nil), entries[index].Modules...)
	}
	return clone
}

func phase36HasProblem(problems []string, prefix string) bool {
	for _, problem := range problems {
		if strings.HasPrefix(problem, prefix) {
			return true
		}
	}
	return false
}

func phase36Digest(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func phase36ErrorCode(message string) string {
	if end := strings.IndexByte(message, ':'); end > 0 {
		return message[:end]
	}
	return ""
}

func phase36EqualStrings(left, right []string) bool {
	return reflect.DeepEqual(left, right)
}

func phase36ContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func phase36HasDuplicates(values []string) bool {
	for index := 1; index < len(values); index++ {
		if values[index] == values[index-1] {
			return true
		}
	}
	return false
}

func phase36FirstByteDifference(left, right []byte) int {
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
