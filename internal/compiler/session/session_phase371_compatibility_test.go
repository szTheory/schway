package session

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/szTheory/schway/internal/compiler/cgen"
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/corevalidate"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const (
	phase371CensusSchema    = "phase371.pre-emitter-census/1"
	phase371AmendmentSchema = "phase371.compatibility-amendments/1"
	phase371CensusPath      = "testdata/phase371/pre-emitter-census.json"
	phase371AmendmentPath   = "testdata/phase371/compatibility-amendments.json"
)

type phase371Census struct {
	Schema           string                  `json:"schema"`
	CapturedRevision string                  `json:"captured_revision"`
	CapturedFiles    []phase371FileBinding   `json:"captured_files"`
	Corpora          []phase371CorpusBinding `json:"corpora"`
	Entries          []phase371CensusEntry   `json:"entries"`
}

type phase371FileBinding struct {
	Path              string `json:"path"`
	SHA256            string `json:"revision_sha256,omitempty"`
	PresentAtRevision bool   `json:"present_at_revision"`
	WorkingSHA        string `json:"pre_emitter_worktree_sha256"`
}

type phase371CorpusBinding struct {
	Name     string `json:"name"`
	Revision string `json:"revision"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
	Rows     int    `json:"rows"`
}

type phase371CensusEntry struct {
	Corpus       string                 `json:"corpus"`
	Path         string                 `json:"path"`
	SourceB64    string                 `json:"source_base64"`
	SourceBytes  int                    `json:"source_bytes"`
	SourceSHA    string                 `json:"source_sha256"`
	CoreB64      string                 `json:"checked_core_base64,omitempty"`
	CoreBytes    int                    `json:"checked_core_bytes,omitempty"`
	CoreSchema   string                 `json:"checked_core_schema,omitempty"`
	CoreSHA      string                 `json:"checked_core_sha256,omitempty"`
	CoreChecked  bool                   `json:"independent_core_validation_ran"`
	CoreValid    bool                   `json:"independent_core_validation"`
	CoreProblems []corevalidate.Problem `json:"independent_core_problems,omitempty"`
	Routes       []phase371CensusRoute  `json:"routes"`
}

type phase371CensusRoute struct {
	Name            string `json:"route"`
	Outcome         string `json:"outcome"`
	Schema          string `json:"schema,omitempty"`
	Code            string `json:"code,omitempty"`
	Error           string `json:"error,omitempty"`
	Reason          string `json:"reason,omitempty"`
	CBase64         string `json:"c_base64,omitempty"`
	CBytes          int    `json:"c_bytes,omitempty"`
	CSHA            string `json:"c_sha256,omitempty"`
	FrozenOutcome   string `json:"frozen_outcome,omitempty"`
	FrozenCBase64   string `json:"frozen_c_base64,omitempty"`
	FrozenCBytes    int    `json:"frozen_c_bytes,omitempty"`
	FrozenCSHA      string `json:"frozen_c_sha256,omitempty"`
	PriorProjection string `json:"prior_projection,omitempty"`
}

type phase371AmendmentLedger struct {
	Schema  string              `json:"schema"`
	Entries []phase371Amendment `json:"entries"`
}

type phase371Amendment struct {
	Corpus          string             `json:"corpus"`
	Path            string             `json:"path"`
	Route           string             `json:"route"`
	SourceSHA       string             `json:"source_sha256"`
	CoreSHA         string             `json:"checked_core_sha256"`
	OldCSHA         string             `json:"old_c_sha256"`
	PriorProjection string             `json:"prior_projection,omitempty"`
	Edits           []phase371ByteEdit `json:"edits"`
	ExpectedCSHA    string             `json:"expected_c_sha256"`
}

type phase371ByteEdit struct {
	Offset       int    `json:"offset"`
	BeforeB64    string `json:"before_base64"`
	AfterB64     string `json:"after_base64"`
	AnchorB64    string `json:"anchor_base64"`
	AnchorOffset int    `json:"anchor_offset"`
	AnchorCount  int    `json:"anchor_count"`
}

type phase371ObservedRoute struct {
	Outcome string
	Schema  string
	Code    string
	Error   string
	Reason  string
	CSource string
}

type phase371FrozenPhase34 struct {
	Schema   string                     `json:"schema"`
	Revision string                     `json:"revision"`
	Entries  []phase371FrozenPhase34Row `json:"entries"`
}

type phase371FrozenPhase34Row struct {
	Path            string                     `json:"path"`
	EmitApplication phase371FrozenPhase34Route `json:"emit_application"`
	ControlNativeC  phase371FrozenPhase34Route `json:"control_native_c"`
}

type phase371FrozenPhase34Route struct {
	Outcome string `json:"outcome"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error,omitempty"`
	Reason  string `json:"reason,omitempty"`
	CSource string `json:"c_source,omitempty"`
}

type phase371ObservedPhase34Row struct {
	Path         string
	Source       []byte
	SourceSHA    string
	Core         []byte
	CoreSHA      string
	CoreSchema   string
	CoreChecked  bool
	CoreValid    bool
	CoreProblems []corevalidate.Problem
	Routes       []phase371CensusRoute
}

var (
	phase371ArtifactOnce sync.Once
	phase371CensusCache  phase371Census
	phase371LedgerCache  phase371AmendmentLedger
	phase371ArtifactErrs []string
	phase371ArtifactsOK  bool
)

func phase371Digest(data []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(data))
}

func phase371CensusFile() string {
	return testsupport.ProjectPath(filepath.FromSlash(phase371CensusPath))
}

func phase371AmendmentFile() string {
	return testsupport.ProjectPath(filepath.FromSlash(phase371AmendmentPath))
}

func phase371ReadJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return fmt.Errorf("decode %s: %w", path, err)
	}
	return nil
}

func phase371Artifacts() (phase371Census, phase371AmendmentLedger, []string, bool) {
	if _, err := os.Stat(phase371CensusFile()); err != nil {
		return phase371Census{}, phase371AmendmentLedger{}, nil, false
	}
	if _, err := os.Stat(phase371AmendmentFile()); err != nil {
		return phase371Census{}, phase371AmendmentLedger{}, nil, false
	}
	phase371ArtifactOnce.Do(func() {
		if err := phase371ReadJSON(phase371CensusFile(), &phase371CensusCache); err != nil {
			phase371ArtifactErrs = append(phase371ArtifactErrs, "census: "+err.Error())
			return
		}
		if err := phase371ReadJSON(phase371AmendmentFile(), &phase371LedgerCache); err != nil {
			phase371ArtifactErrs = append(phase371ArtifactErrs, "amendments: "+err.Error())
			return
		}
		phase371ArtifactErrs = append(phase371ArtifactErrs, phase371CensusProblems(phase371CensusCache)...)
		phase371ArtifactErrs = append(phase371ArtifactErrs, phase371AmendmentProblems(phase371CensusCache, phase371LedgerCache)...)
		phase371ArtifactsOK = true
	})
	return phase371CensusCache, phase371LedgerCache, append([]string(nil), phase371ArtifactErrs...), phase371ArtifactsOK
}

func phase371CensusProblems(census phase371Census) []string {
	var problems []string
	if census.Schema != phase371CensusSchema {
		problems = append(problems, "census schema changed")
	}
	if census.CapturedRevision == "" || !phase371GitObjectExists(census.CapturedRevision) {
		problems = append(problems, "census captured revision is missing or unavailable")
	} else if !phase371IsAncestor(census.CapturedRevision, "HEAD") {
		problems = append(problems, "census captured revision is not an ancestor of HEAD")
	}
	if len(census.Corpora) != 2 || census.Corpora[0].Name != "phase34" || census.Corpora[1].Name != "phase36" {
		problems = append(problems, "census must bind Phase34 and Phase36 corpora in order")
	}
	for _, binding := range census.Corpora {
		if binding.Revision == "" || binding.Path == "" || !phase371HexDigest(binding.SHA256) || binding.Rows <= 0 {
			problems = append(problems, binding.Name+": incomplete frozen corpus binding")
			continue
		}
		raw, err := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(binding.Path)))
		if err != nil || phase371Digest(raw) != binding.SHA256 {
			problems = append(problems, binding.Name+": frozen corpus bytes/hash changed")
		}
		var inventory struct {
			Entries []struct {
				Path string `json:"path"`
			} `json:"entries"`
		}
		if err == nil && json.Unmarshal(raw, &inventory) == nil {
			if len(inventory.Entries) != binding.Rows {
				problems = append(problems, binding.Name+": frozen corpus row count differs from its binding")
			}
			var found []string
			for _, entry := range census.Entries {
				if entry.Corpus == binding.Name {
					found = append(found, entry.Path)
				}
			}
			var expected []string
			for _, entry := range inventory.Entries {
				expected = append(expected, entry.Path)
			}
			if !phase36EqualStrings(found, expected) {
				problems = append(problems, binding.Name+": census path inventory differs from the frozen corpus")
			}
		}
	}
	fileSet := map[string]bool{}
	for _, binding := range census.CapturedFiles {
		if binding.Path == "" || fileSet[binding.Path] || !phase371HexDigest(binding.WorkingSHA) || (binding.PresentAtRevision && !phase371HexDigest(binding.SHA256)) || (!binding.PresentAtRevision && binding.SHA256 != "") {
			problems = append(problems, "captured file provenance has missing, duplicate, or malformed binding")
			continue
		}
		fileSet[binding.Path] = true
		captured, err := phase371GitShow(census.CapturedRevision, binding.Path)
		if (binding.PresentAtRevision && (err != nil || phase371Digest(captured) != binding.SHA256)) || (!binding.PresentAtRevision && err == nil) {
			problems = append(problems, binding.Path+": captured source hash does not match the recorded revision")
		}
	}
	for _, path := range []string{
		"internal/compiler/cgen/cgen_program.go",
		"internal/compiler/cgen/cgen.go",
		"internal/compiler/session/session_phase34_prechange_test.go",
		"internal/compiler/session/session_phase36_prechange_test.go",
		"internal/compiler/session/session_phase371_compatibility_test.go",
	} {
		if !fileSet[path] {
			problems = append(problems, path+": missing captured source/test provenance")
		}
	}
	if len(census.Entries) != 325 {
		problems = append(problems, fmt.Sprintf("census inventory has %d source rows, want 325", len(census.Entries)))
	}
	seen := map[string]bool{}
	lastKey := ""
	for _, entry := range census.Entries {
		key := entry.Corpus + "\x00" + entry.Path
		if entry.Corpus != "phase34" && entry.Corpus != "phase36" {
			problems = append(problems, key+": unknown source corpus")
		}
		if entry.Path == "" || seen[key] || (lastKey != "" && key < lastKey) {
			problems = append(problems, key+": census paths must be sorted and unique")
		}
		seen[key] = true
		lastKey = key
		source, sourceErr := base64.StdEncoding.DecodeString(entry.SourceB64)
		if sourceErr != nil || len(source) != entry.SourceBytes || phase371Digest(source) != entry.SourceSHA {
			problems = append(problems, key+": source bytes/hash binding is invalid")
		}
		if entry.CoreB64 == "" {
			if entry.CoreBytes != 0 || entry.CoreSHA != "" || entry.CoreSchema != "" || entry.CoreChecked || entry.CoreValid || len(entry.CoreProblems) != 0 {
				problems = append(problems, key+": absent core bytes carry a core binding")
			}
		} else {
			coreBytes, err := base64.StdEncoding.DecodeString(entry.CoreB64)
			if err != nil || len(coreBytes) != entry.CoreBytes || phase371Digest(coreBytes) != entry.CoreSHA || entry.CoreSchema == "" || !entry.CoreChecked || (entry.CoreValid == (len(entry.CoreProblems) != 0)) {
				problems = append(problems, key+": checked core bytes/hash/independent validation binding is invalid")
			}
		}
		wantRoutes := []string{"emit", "emit_native", "emit_application"}
		if entry.Corpus == "phase34" {
			wantRoutes = []string{"emit", "emit_native", "emit_application", "control_native_c"}
		}
		if len(entry.Routes) != len(wantRoutes) {
			problems = append(problems, key+": route inventory is incomplete")
			continue
		}
		for index, route := range entry.Routes {
			if route.Name != wantRoutes[index] {
				problems = append(problems, key+": route inventory/order mismatch")
			}
			if route.Outcome != "emitted" && route.Outcome != "refused" && route.Outcome != "inapplicable" {
				problems = append(problems, key+"/"+route.Name+": unknown route outcome")
			}
			if route.Outcome == "emitted" {
				cBytes, err := base64.StdEncoding.DecodeString(route.CBase64)
				if err != nil || len(cBytes) != route.CBytes || phase371Digest(cBytes) != route.CSHA || len(cBytes) == 0 {
					problems = append(problems, key+"/"+route.Name+": pre-emitter C bytes/hash binding is invalid")
				}
			} else if route.CBase64 != "" || route.CBytes != 0 || route.CSHA != "" {
				problems = append(problems, key+"/"+route.Name+": non-emitted route carries C bytes")
			}
			if route.FrozenCBase64 != "" {
				frozen, err := base64.StdEncoding.DecodeString(route.FrozenCBase64)
				if err != nil || len(frozen) != route.FrozenCBytes || phase371Digest(frozen) != route.FrozenCSHA || route.FrozenOutcome != "emitted" {
					problems = append(problems, key+"/"+route.Name+": original frozen C binding is invalid")
				}
			}
			if route.PriorProjection != "" && route.PriorProjection != "phase35-adler-exact-adapter" {
				problems = append(problems, key+"/"+route.Name+": unknown prior projection")
			}
			if route.PriorProjection == "phase35-adler-exact-adapter" {
				beforeBytes, beforeSHA, afterBytes, afterSHA, ok := phase371Phase35Projection(entry.Path, route.Name)
				if !ok || route.FrozenCBytes != beforeBytes || route.FrozenCSHA != beforeSHA || route.CBytes != afterBytes || route.CSHA != afterSHA {
					problems = append(problems, key+"/"+route.Name+": prior Phase35 projection is not bound to the exact frozen and projected C bytes")
				}
			} else if entry.Corpus == "phase34" && route.FrozenOutcome == "emitted" && route.Outcome == "emitted" && (route.FrozenCBytes != route.CBytes || route.FrozenCSHA != route.CSHA) {
				problems = append(problems, key+"/"+route.Name+": Phase34 frozen C bytes changed without a prior exact projection")
			}
		}
	}
	return problems
}

func phase371AmendmentProblems(census phase371Census, ledger phase371AmendmentLedger) []string {
	var problems []string
	if ledger.Schema != phase371AmendmentSchema {
		problems = append(problems, "amendment ledger schema changed")
	}
	rows := make(map[string]struct {
		entry phase371CensusEntry
		route phase371CensusRoute
	})
	for _, entry := range census.Entries {
		for _, route := range entry.Routes {
			rows[phase371RouteKey(entry.Corpus, entry.Path, route.Name)] = struct {
				entry phase371CensusEntry
				route phase371CensusRoute
			}{entry, route}
		}
	}
	seen := map[string]bool{}
	lastKey := ""
	for _, row := range ledger.Entries {
		key := phase371RouteKey(row.Corpus, row.Path, row.Route)
		if seen[key] {
			problems = append(problems, key+": duplicate amendment row")
		}
		if lastKey != "" && key < lastKey {
			problems = append(problems, key+": amendment rows must be sorted")
		}
		seen[key] = true
		lastKey = key
		bound, ok := rows[key]
		if !ok {
			problems = append(problems, key+": extra, wrong-path, or wrong-route amendment")
			continue
		}
		if bound.route.Outcome != "emitted" {
			problems = append(problems, key+": amendments require an emitted census route")
			continue
		}
		if row.SourceSHA != bound.entry.SourceSHA || row.CoreSHA != bound.entry.CoreSHA || row.OldCSHA != bound.route.CSHA || row.PriorProjection != bound.route.PriorProjection {
			problems = append(problems, key+": stale or wrong old source/core/C binding")
		}
		old, err := base64.StdEncoding.DecodeString(bound.route.CBase64)
		if err != nil {
			problems = append(problems, key+": invalid census C bytes")
			continue
		}
		amended, err := phase371ApplyEdits(old, row.Edits)
		if err != nil {
			problems = append(problems, key+": "+err.Error())
			continue
		}
		if !phase371HexDigest(row.ExpectedCSHA) || phase371Digest(amended) != row.ExpectedCSHA {
			problems = append(problems, key+": exact reviewed delta or expected post-delta digest mismatch")
		}
	}
	return problems
}

func phase371ApplyEdits(old []byte, edits []phase371ByteEdit) ([]byte, error) {
	if len(edits) == 0 {
		return nil, fmt.Errorf("amendment has no reviewed byte edits")
	}
	ordered := append([]phase371ByteEdit(nil), edits...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Offset > ordered[j].Offset })
	lastStart := len(old) + 1
	result := append([]byte(nil), old...)
	for _, edit := range ordered {
		before, beforeErr := base64.StdEncoding.DecodeString(edit.BeforeB64)
		after, afterErr := base64.StdEncoding.DecodeString(edit.AfterB64)
		anchor, anchorErr := base64.StdEncoding.DecodeString(edit.AnchorB64)
		if beforeErr != nil || afterErr != nil || anchorErr != nil || len(before) == 0 || len(after) == 0 || len(anchor) == 0 || bytes.Equal(before, after) {
			return nil, fmt.Errorf("invalid before/after/anchor bytes")
		}
		if edit.AnchorCount != 1 || edit.AnchorOffset < 0 || edit.AnchorOffset > len(anchor)-len(before) {
			return nil, fmt.Errorf("anchor must declare one exact match and contain the old bytes at its offset")
		}
		if !bytes.Equal(anchor[edit.AnchorOffset:edit.AnchorOffset+len(before)], before) {
			return nil, fmt.Errorf("anchor does not contain the declared old bytes at its offset")
		}
		if phase371CountOccurrences(old, anchor) != 1 {
			return nil, fmt.Errorf("anchor must occur exactly once in the bound old C bytes")
		}
		start := edit.Offset
		if start < 0 || start > len(result)-len(before) {
			return nil, fmt.Errorf("reviewed byte offset is outside the bound old C bytes")
		}
		end := start + len(before)
		anchorStart := start - edit.AnchorOffset
		if anchorStart < 0 || anchorStart > len(old)-len(anchor) || !bytes.Equal(old[anchorStart:anchorStart+len(anchor)], anchor) {
			return nil, fmt.Errorf("reviewed byte offset does not point to the unique declared anchor")
		}
		if end > lastStart {
			return nil, fmt.Errorf("reviewed byte edits overlap or are not independently addressable")
		}
		if !bytes.Equal(result[start:end], before) {
			return nil, fmt.Errorf("reviewed old bytes differ at the declared offset")
		}
		result = append(append(append([]byte(nil), result[:start]...), after...), result[end:]...)
		lastStart = start
	}
	return result, nil
}

func phase371CountOccurrences(data, pattern []byte) int {
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

func phase371RouteKey(corpus, path, route string) string {
	return corpus + "\x00" + path + "\x00" + route
}

func phase371HexDigest(value string) bool {
	return len(value) == 64 && strings.Trim(value, "0123456789abcdef") == ""
}

func phase371Phase35Projection(path, route string) (int, string, int, string, bool) {
	switch {
	case path == "examples/checksum/adler32.schway" && route == "control_native_c":
		return 44595, "d6581de8d80080d96fc762b74dcf1964e0eae438fd566cb6a0c68280f606765b", 19163, "8653dbec998b2d39975619912b400e465b299d9d60ef59bdb1397ed9991d8e6e", true
	case path == "examples/checksum/adler32_cli.schway" && route == "emit_application":
		return 19351, "0bf9cda02e6806c1aeadd4912fabdb46a1856e70ae2a69ba0a9071db86ebc350", 14911, "80df93af9f946d5db3fc520c412f7de31cec28efa7c658ce9a54dd6080950719", true
	case path == "examples/checksum/adler_low.schway" && route == "emit_application":
		return 21511, "34dcb4f59f20e2234ddda81654da483c1d9f9e89a18b2567d43b5fec5b6af234", 17071, "31b6410aa3ca9f43bda6db85b6aebd004d2abb30cea03df149be314753698495", true
	case path == "examples/checksum/checksum.schway" && route == "emit_application":
		return 19598, "9bb3babccd3d9a88b36bdedd148d0b548fe142128f2367e1a33f806d2f298751", 15158, "baef335d68fca5df6ed61c63fb6b6248ebbae04a38cc83937cdc19c4a0cdac9e", true
	default:
		return 0, "", 0, "", false
	}
}

func phase371GitObjectExists(revision string) bool {
	cmd := exec.Command("git", "cat-file", "-e", revision+"^{commit}")
	cmd.Dir = testsupport.ProjectPath(".")
	return cmd.Run() == nil
}

func phase371IsAncestor(older, newer string) bool {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", older, newer)
	cmd.Dir = testsupport.ProjectPath(".")
	return cmd.Run() == nil
}

func phase371GitShow(revision, path string) ([]byte, error) {
	cmd := exec.Command("git", "show", revision+":"+path)
	cmd.Dir = testsupport.ProjectPath(".")
	return cmd.Output()
}

func phase371ObserveComparison(problems *[]string, corpus, path, route, sourceSHA, coreSHA string, observed phase371ObservedRoute) bool {
	census, ledger, artifactProblems, loaded := phase371Artifacts()
	if !loaded {
		return false
	}
	for _, problem := range artifactProblems {
		*problems = append(*problems, "Phase371 census/amendment: "+problem)
	}
	var source *phase371CensusEntry
	var baseline *phase371CensusRoute
	for index := range census.Entries {
		entry := &census.Entries[index]
		if entry.Corpus != corpus || entry.Path != path {
			continue
		}
		source = entry
		for routeIndex := range entry.Routes {
			if entry.Routes[routeIndex].Name == route {
				baseline = &entry.Routes[routeIndex]
				break
			}
		}
		break
	}
	if source == nil || baseline == nil {
		*problems = append(*problems, corpus+"/"+path+"/"+route+": missing from complete Phase371 census")
		return true
	}
	if source.SourceSHA != sourceSHA || source.CoreSHA != coreSHA {
		*problems = append(*problems, corpus+"/"+path+"/"+route+": source or checked-core bytes changed since the pre-emitter census")
	}
	if baseline.Outcome != observed.Outcome {
		*problems = append(*problems, fmt.Sprintf("%s: %s outcome changed: got %q want %q", path, route, observed.Outcome, baseline.Outcome))
		return true
	}
	if baseline.Schema != observed.Schema {
		*problems = append(*problems, fmt.Sprintf("%s: %s schema changed", path, route))
	}
	if baseline.Outcome == "inapplicable" {
		if corpus == "phase36" && baseline.Reason != observed.Reason {
			*problems = append(*problems, fmt.Sprintf("%s: %s non-emission reason changed", path, route))
		}
		return true
	}
	if baseline.Outcome == "refused" {
		if baseline.Code != observed.Code || baseline.Error != observed.Error {
			*problems = append(*problems, fmt.Sprintf("%s: %s refusal changed", path, route))
		}
		return true
	}
	base, err := base64.StdEncoding.DecodeString(baseline.CBase64)
	if err != nil {
		*problems = append(*problems, corpus+"/"+path+"/"+route+": invalid census C bytes")
		return true
	}
	for _, amendment := range ledger.Entries {
		if phase371RouteKey(amendment.Corpus, amendment.Path, amendment.Route) != phase371RouteKey(corpus, path, route) {
			continue
		}
		base, err = phase371ApplyEdits(base, amendment.Edits)
		if err != nil {
			*problems = append(*problems, corpus+"/"+path+"/"+route+": "+err.Error())
			return true
		}
		break
	}
	got := []byte(observed.CSource)
	if !bytes.Equal(base, got) {
		index := phase36FirstByteDifference(base, got)
		*problems = append(*problems, fmt.Sprintf("%s: %s C byte mismatch at byte %d (got length %d, want length %d)", path, route, index, len(got), len(base)))
	}
	return true
}

func phase371Outcome(outcome string) string {
	if outcome == "not_emitted" {
		return "inapplicable"
	}
	return outcome
}

func phase371Observed36(result phase36Emission) phase371ObservedRoute {
	return phase371ObservedRoute{Outcome: phase371Outcome(result.Outcome), Schema: result.Schema, Code: result.Code, Error: result.Error, Reason: result.Reason, CSource: result.CSource}
}

func phase371CaptureCensus(t *testing.T) phase371Census {
	t.Helper()
	revisionOutput, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		t.Fatalf("capture current git revision: %v", err)
	}
	revision := strings.TrimSpace(string(revisionOutput))
	census := phase371Census{Schema: phase371CensusSchema, CapturedRevision: revision}
	for _, path := range []string{
		"internal/compiler/cgen/cgen_program.go",
		"internal/compiler/cgen/cgen.go",
		"internal/compiler/session/session_phase34_prechange_test.go",
		"internal/compiler/session/session_phase36_prechange_test.go",
		"internal/compiler/session/session_phase371_compatibility_test.go",
	} {
		current, readErr := os.ReadFile(testsupport.ProjectPath(filepath.FromSlash(path)))
		if readErr != nil {
			t.Fatalf("read pre-emitter captured file %s: %v", path, readErr)
		}
		captured, showErr := phase371GitShow(revision, path)
		present := showErr == nil
		if !present && path != "internal/compiler/session/session_phase371_compatibility_test.go" {
			t.Fatalf("read source at captured revision %s: %v", path, showErr)
		}
		if strings.Contains(path, "/cgen/") && !bytes.Equal(current, captured) {
			t.Fatalf("cgen source changed before route census: %s", path)
		}
		binding := phase371FileBinding{Path: path, PresentAtRevision: present, WorkingSHA: phase371Digest(current)}
		if present {
			binding.SHA256 = phase371Digest(captured)
		}
		census.CapturedFiles = append(census.CapturedFiles, binding)
	}

	phase34Path := testsupport.ProjectPath("testdata", "phase34", "prechange-corpus.json")
	phase34Raw, err := os.ReadFile(phase34Path)
	if err != nil {
		t.Fatal(err)
	}
	var phase34Corpus phase371FrozenPhase34
	if err := json.Unmarshal(phase34Raw, &phase34Corpus); err != nil {
		t.Fatalf("decode frozen Phase34 corpus: %v", err)
	}
	phase36Raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase36", "prechange-corpus.json"))
	if err != nil {
		t.Fatal(err)
	}
	phase36Corpus := phase36LoadPreChangeCorpus(t)
	census.Corpora = []phase371CorpusBinding{
		{Name: "phase34", Revision: phase34Corpus.Revision, Path: "testdata/phase34/prechange-corpus.json", SHA256: phase371Digest(phase34Raw), Rows: len(phase34Corpus.Entries)},
		{Name: "phase36", Revision: phase36Corpus.Revision, Path: "testdata/phase36/prechange-corpus.json", SHA256: phase371Digest(phase36Raw), Rows: len(phase36Corpus.Entries)},
	}
	for _, entry := range phase371ObservePhase34(t, phase34Corpus) {
		census.Entries = append(census.Entries, phase371Entry34(entry))
	}
	for _, entry := range phase36ObservePreChangeCorpus(t, phase36Corpus) {
		census.Entries = append(census.Entries, phase371Entry36(t, entry))
	}
	sort.Slice(census.Entries, func(i, j int) bool {
		if census.Entries[i].Corpus != census.Entries[j].Corpus {
			return census.Entries[i].Corpus < census.Entries[j].Corpus
		}
		return census.Entries[i].Path < census.Entries[j].Path
	})
	return census
}

func phase371IndependentCore(t *testing.T, coreBytes []byte, path string) (bool, []corevalidate.Problem) {
	t.Helper()
	var program core.Program
	if err := json.Unmarshal(coreBytes, &program); err != nil {
		t.Fatalf("decode independently checked core for %s: %v", path, err)
	}
	result := corevalidate.Validate(program)
	return result.Valid, result.Problems
}

func phase371ObservePhase34(t *testing.T, corpus phase371FrozenPhase34) []phase371ObservedPhase34Row {
	t.Helper()
	entries := make([]phase371ObservedPhase34Row, 0, len(corpus.Entries))
	for _, frozen := range corpus.Entries {
		path := testsupport.ProjectPath(filepath.FromSlash(frozen.Path))
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read current Phase34 source %s: %v", frozen.Path, err)
		}
		entry := phase371ObservedPhase34Row{Path: frozen.Path, Source: source, SourceSHA: phase371Digest(source)}
		checked, checkErr := CheckFile(path)
		if checkErr != nil || len(checked.Diagnostics) != 0 {
			reason := "CheckFile error: " + fmt.Sprint(checkErr)
			if checkErr == nil {
				codes := make([]string, 0, len(checked.Diagnostics))
				for _, diagnostic := range checked.Diagnostics {
					codes = append(codes, diagnostic.Code)
				}
				sort.Strings(codes)
				reason = "CheckFile diagnostics: " + strings.Join(codes, ",")
			}
			for _, name := range []string{"emit", "emit_native", "emit_application", "control_native_c"} {
				entry.Routes = append(entry.Routes, phase371CensusRoute{Name: name, Outcome: "inapplicable", Reason: reason})
			}
			entries = append(entries, phase371ApplyFrozen34(t, entry, frozen))
			continue
		}
		entry.Core, err = json.Marshal(checked.Program)
		if err != nil {
			t.Fatalf("marshal checked core for %s: %v", frozen.Path, err)
		}
		entry.CoreSHA = phase371Digest(entry.Core)
		entry.CoreSchema = checked.Program.Schema
		entry.CoreChecked = true
		entry.CoreValid, entry.CoreProblems = phase371IndependentCore(t, entry.Core, entry.Path)
		c, emitErr := cgen.Emit(checked.Program)
		entry.Routes = append(entry.Routes, phase371EmissionRoute("emit", c, emitErr))
		c, emitErr = cgen.EmitNative(checked.Program)
		entry.Routes = append(entry.Routes, phase371EmissionRoute("emit_native", c, emitErr))
		c, emitErr = cgen.EmitApplication(checked.Program)
		entry.Routes = append(entry.Routes, phase371EmissionRoute("emit_application", c, emitErr))
		c, emitErr = Phase16ControlNativeC(checked.Program, "phase34-prechange")
		entry.Routes = append(entry.Routes, phase371EmissionRoute("control_native_c", c, emitErr))
		entries = append(entries, phase371ApplyFrozen34(t, entry, frozen))
	}
	return entries
}

func phase371EmissionRoute(name, c string, err error) phase371CensusRoute {
	if err != nil {
		message := err.Error()
		code := ""
		if colon := strings.IndexByte(message, ':'); colon > 0 {
			code = message[:colon]
		}
		return phase371CensusRoute{Name: name, Outcome: "refused", Code: code, Error: message}
	}
	return phase371CensusRoute{
		Name: name, Outcome: "emitted", CBase64: base64.StdEncoding.EncodeToString([]byte(c)),
		CBytes: len(c), CSHA: phase371Digest([]byte(c)),
	}
}

func phase371ApplyFrozen34(t *testing.T, entry phase371ObservedPhase34Row, frozen phase371FrozenPhase34Row) phase371ObservedPhase34Row {
	t.Helper()
	if len(entry.Routes) != 4 {
		t.Fatalf("Phase34 route observation for %s has %d rows, want 4", entry.Path, len(entry.Routes))
	}
	for index, old := range []phase371FrozenPhase34Route{frozen.EmitApplication, frozen.ControlNativeC} {
		routeIndex := index + 2
		current := &entry.Routes[routeIndex]
		current.FrozenOutcome = phase371Outcome(old.Outcome)
		if old.Outcome == "emitted" {
			current.FrozenCBase64 = base64.StdEncoding.EncodeToString([]byte(old.CSource))
			current.FrozenCBytes = len(old.CSource)
			current.FrozenCSHA = phase371Digest([]byte(old.CSource))
		}
		if phase371HasPriorProjection(entry.Path, current.Name) {
			current.PriorProjection = "phase35-adler-exact-adapter"
		}
	}
	return entry
}

func phase371HasPriorProjection(path, route string) bool {
	if route == "control_native_c" && path == "examples/checksum/adler32.schway" {
		return true
	}
	if route != "emit_application" {
		return false
	}
	switch path {
	case "examples/checksum/adler32_cli.schway", "examples/checksum/adler_low.schway", "examples/checksum/checksum.schway":
		return true
	default:
		return false
	}
}

func phase371Entry34(entry phase371ObservedPhase34Row) phase371CensusEntry {
	return phase371CensusEntry{
		Corpus: "phase34", Path: entry.Path,
		SourceB64: base64.StdEncoding.EncodeToString(entry.Source), SourceBytes: len(entry.Source), SourceSHA: entry.SourceSHA,
		CoreB64: base64.StdEncoding.EncodeToString(entry.Core), CoreBytes: len(entry.Core), CoreSchema: entry.CoreSchema, CoreSHA: entry.CoreSHA, CoreChecked: entry.CoreChecked, CoreValid: entry.CoreValid, CoreProblems: entry.CoreProblems,
		Routes: entry.Routes,
	}
}

func phase371Entry36(t *testing.T, entry phase36PreChangeEntry) phase371CensusEntry {
	coreChecked, coreValid := false, false
	var coreProblems []corevalidate.Problem
	if len(entry.checkedCoreBytes) != 0 {
		coreChecked = true
		coreValid, coreProblems = phase371IndependentCore(t, entry.checkedCoreBytes, entry.Path)
	}
	return phase371CensusEntry{
		Corpus: "phase36", Path: entry.Path,
		SourceB64: base64.StdEncoding.EncodeToString([]byte(entry.Source)), SourceBytes: len(entry.Source), SourceSHA: strings.TrimPrefix(entry.SourceSHA, "sha256:"),
		CoreB64: base64.StdEncoding.EncodeToString(entry.checkedCoreBytes), CoreBytes: len(entry.checkedCoreBytes), CoreSchema: entry.CoreSchema, CoreSHA: strings.TrimPrefix(entry.CoreSHA, "sha256:"), CoreChecked: coreChecked, CoreValid: coreValid, CoreProblems: coreProblems,
		Routes: []phase371CensusRoute{phase371Route36("emit", entry.Emit), phase371Route36("emit_native", entry.EmitNative), phase371Route36("emit_application", entry.Application)},
	}
}

func phase371Route36(name string, current phase36Emission) phase371CensusRoute {
	route := phase371CensusRoute{Name: name, Outcome: phase371Outcome(current.Outcome), Schema: current.Schema, Code: current.Code, Error: current.Error, Reason: current.Reason}
	if current.Outcome == "emitted" {
		route.CBase64 = base64.StdEncoding.EncodeToString([]byte(current.CSource))
		route.CBytes = len(current.CSource)
		route.CSHA = phase371Digest([]byte(current.CSource))
		route.FrozenOutcome = "emitted"
		route.FrozenCBase64 = route.CBase64
		route.FrozenCBytes = route.CBytes
		route.FrozenCSHA = route.CSHA
	} else {
		route.FrozenOutcome = route.Outcome
	}
	return route
}

func TestPhase371PreEmitterRouteCensus(t *testing.T) {
	path := phase371CensusFile()
	if os.Getenv("GSD_PHASE371_CAPTURE") == "1" {
		if _, err := os.Stat(path); err == nil {
			t.Fatal("refusing to regenerate the immutable pre-emitter census")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		census := phase371CaptureCensus(t)
		encoded, err := json.MarshalIndent(census, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	var census phase371Census
	if err := phase371ReadJSON(path, &census); err != nil {
		t.Fatalf("read immutable pre-emitter census: %v", err)
	}
	if problems := phase371CensusProblems(census); len(problems) != 0 {
		t.Fatalf("pre-emitter census is incomplete or stale: %s", strings.Join(problems, "; "))
	}
	if len(census.Entries) != 325 {
		t.Fatalf("census has %d source rows, want 325", len(census.Entries))
	}
	var ledger phase371AmendmentLedger
	if err := phase371ReadJSON(phase371AmendmentFile(), &ledger); err != nil {
		t.Fatalf("read amendment ledger: %v", err)
	}
	if len(ledger.Entries) != 0 {
		t.Fatalf("initial amendment ledger must be empty, found %d rows", len(ledger.Entries))
	}
	if problems := phase371AmendmentProblems(census, ledger); len(problems) != 0 {
		t.Fatalf("empty amendment ledger is invalid: %s", strings.Join(problems, "; "))
	}
	phase371VerifyCurrentCensus(t, census)
}

func phase371VerifyCurrentCensus(t *testing.T, census phase371Census) {
	t.Helper()
	var problems []string
	phase34Raw, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase34", "prechange-corpus.json"))
	if err != nil {
		t.Fatal(err)
	}
	var frozen34 phase371FrozenPhase34
	if err := json.Unmarshal(phase34Raw, &frozen34); err != nil {
		t.Fatal(err)
	}
	for _, entry := range phase371ObservePhase34(t, frozen34) {
		phase371CompareCoreValidation(t, &problems, census, "phase34", entry.Path, entry.Core)
		for _, route := range entry.Routes {
			cBytes, _ := base64.StdEncoding.DecodeString(route.CBase64)
			observed := phase371ObservedRoute{Outcome: route.Outcome, Code: route.Code, Error: route.Error, Reason: route.Reason, CSource: string(cBytes)}
			if !phase371ObserveComparison(&problems, "phase34", entry.Path, route.Name, entry.SourceSHA, entry.CoreSHA, observed) {
				problems = append(problems, entry.Path+"/"+route.Name+": census comparator hooks were not loaded")
			}
		}
	}
	phase36Corpus := phase36LoadPreChangeCorpus(t)
	for _, entry := range phase36ObservePreChangeCorpus(t, phase36Corpus) {
		phase371CompareCoreValidation(t, &problems, census, "phase36", entry.Path, entry.checkedCoreBytes)
		for _, route := range phase36RouteAccessors {
			observed := phase371Observed36(*route.get(&entry))
			if !phase371ObserveComparison(&problems, "phase36", entry.Path, route.name, strings.TrimPrefix(entry.SourceSHA, "sha256:"), strings.TrimPrefix(entry.CoreSHA, "sha256:"), observed) {
				problems = append(problems, entry.Path+"/"+route.name+": census comparator hooks were not loaded")
			}
		}
	}
	if len(problems) != 0 {
		t.Fatalf("current observations differ from the frozen pre-emitter census: %s", strings.Join(problems, "; "))
	}
}

func phase371CompareCoreValidation(t *testing.T, problems *[]string, census phase371Census, corpus, path string, coreBytes []byte) {
	t.Helper()
	var expected *phase371CensusEntry
	for index := range census.Entries {
		if census.Entries[index].Corpus == corpus && census.Entries[index].Path == path {
			expected = &census.Entries[index]
			break
		}
	}
	if expected == nil {
		*problems = append(*problems, corpus+"/"+path+": missing core validation row")
		return
	}
	if len(coreBytes) == 0 {
		if expected.CoreChecked || expected.CoreSHA != "" {
			*problems = append(*problems, corpus+"/"+path+": checked core disappeared")
		}
		return
	}
	valid, details := phase371IndependentCore(t, coreBytes, path)
	if !expected.CoreChecked || expected.CoreValid != valid || !reflect.DeepEqual(expected.CoreProblems, details) {
		*problems = append(*problems, corpus+"/"+path+": independent core validation result changed")
	}
}

func TestPhase371CompatibilityAmendmentControls(t *testing.T) {
	var census phase371Census
	if err := phase371ReadJSON(phase371CensusFile(), &census); err != nil {
		t.Fatalf("read census: %v", err)
	}
	var ledger phase371AmendmentLedger
	if err := phase371ReadJSON(phase371AmendmentFile(), &ledger); err != nil {
		t.Fatalf("read amendment ledger: %v", err)
	}
	if len(ledger.Entries) != 0 {
		t.Fatalf("new amendment ledger must begin empty, found %d rows", len(ledger.Entries))
	}
	if problems := phase371AmendmentProblems(census, ledger); len(problems) != 0 {
		t.Fatalf("empty amendment ledger must pass: %v", problems)
	}

	old := []byte("prefix<old>middle<old>suffix")
	edit := phase371ByteEdit{
		Offset:       6,
		BeforeB64:    base64.StdEncoding.EncodeToString([]byte("<old>")),
		AfterB64:     base64.StdEncoding.EncodeToString([]byte("<new>")),
		AnchorB64:    base64.StdEncoding.EncodeToString([]byte("prefix<old>middle")),
		AnchorOffset: 6, AnchorCount: 1,
	}
	newBytes, err := phase371ApplyEdits(old, []phase371ByteEdit{edit})
	if err != nil || string(newBytes) != "prefix<new>middle<old>suffix" {
		t.Fatalf("exact byte edit projection failed: %q %v", newBytes, err)
	}
	synthetic := phase371Census{Schema: phase371CensusSchema, Entries: []phase371CensusEntry{{
		Corpus: "phase36", Path: "examples/control.schway",
		SourceSHA: phase371Digest([]byte("source")), CoreSHA: phase371Digest([]byte("core")),
		Routes: []phase371CensusRoute{{Name: "emit", Outcome: "emitted", CBase64: base64.StdEncoding.EncodeToString(old), CBytes: len(old), CSHA: phase371Digest(old)}},
	}}}
	base := synthetic.Entries[0].Routes[0]
	valid := phase371AmendmentLedger{Schema: phase371AmendmentSchema, Entries: []phase371Amendment{{
		Corpus: "phase36", Path: "examples/control.schway", Route: "emit",
		SourceSHA: synthetic.Entries[0].SourceSHA, CoreSHA: synthetic.Entries[0].CoreSHA, OldCSHA: base.CSHA,
		Edits: []phase371ByteEdit{edit}, ExpectedCSHA: phase371Digest(newBytes),
	}}}
	if problems := phase371AmendmentProblems(synthetic, valid); len(problems) != 0 {
		t.Fatalf("clean exact amendment was rejected: %v", problems)
	}
	controls := []struct {
		name   string
		mutate func(*phase371AmendmentLedger)
		want   string
	}{
		{"duplicate", func(l *phase371AmendmentLedger) { l.Entries = append(l.Entries, l.Entries[0]) }, "duplicate amendment row"},
		{"extra-path", func(l *phase371AmendmentLedger) { l.Entries[0].Path = "examples/other.schway" }, "extra, wrong-path, or wrong-route"},
		{"wrong-route", func(l *phase371AmendmentLedger) { l.Entries[0].Route = "emit_native" }, "extra, wrong-path, or wrong-route"},
		{"stale-source", func(l *phase371AmendmentLedger) { l.Entries[0].SourceSHA = strings.Repeat("0", 64) }, "stale or wrong old source/core/C binding"},
		{"stale-core", func(l *phase371AmendmentLedger) { l.Entries[0].CoreSHA = strings.Repeat("0", 64) }, "stale or wrong old source/core/C binding"},
		{"stale-old-c", func(l *phase371AmendmentLedger) { l.Entries[0].OldCSHA = strings.Repeat("0", 64) }, "stale or wrong old source/core/C binding"},
		{"wrong-prior-projection", func(l *phase371AmendmentLedger) { l.Entries[0].PriorProjection = "phase35-other" }, "stale or wrong old source/core/C binding"},
		{"wrong-delta", func(l *phase371AmendmentLedger) {
			l.Entries[0].Edits[0].AfterB64 = base64.StdEncoding.EncodeToString([]byte("<bad>"))
		}, "exact reviewed delta"},
		{"wrong-offset", func(l *phase371AmendmentLedger) { l.Entries[0].Edits[0].Offset++ }, "unique declared anchor"},
		{"ambiguous-anchor", func(l *phase371AmendmentLedger) {
			l.Entries[0].Edits[0].AnchorB64 = base64.StdEncoding.EncodeToString([]byte("<old>"))
			l.Entries[0].Edits[0].AnchorOffset = 0
		}, "exactly once"},
		{"wrong-anchor-count", func(l *phase371AmendmentLedger) { l.Entries[0].Edits[0].AnchorCount = 2 }, "declare one exact match"},
		{"unrelated-post-bytes", func(l *phase371AmendmentLedger) { l.Entries[0].ExpectedCSHA = phase371Digest([]byte("unrelated")) }, "exact reviewed delta"},
	}
	for _, control := range controls {
		t.Run(control.name, func(t *testing.T) {
			changed := phase371AmendmentLedger{Schema: valid.Schema, Entries: append([]phase371Amendment(nil), valid.Entries...)}
			changed.Entries[0].Edits = append([]phase371ByteEdit(nil), valid.Entries[0].Edits...)
			control.mutate(&changed)
			problems := phase371AmendmentProblems(synthetic, changed)
			if !phase371HasProblem(problems, control.want) {
				t.Fatalf("invalid amendment control did not reach %q: %v", control.want, problems)
			}
			if restored := phase371AmendmentProblems(synthetic, valid); len(restored) != 0 {
				t.Fatalf("restored exact amendment failed: %v", restored)
			}
		})
	}
	without := phase371AmendmentLedger{Schema: phase371AmendmentSchema}
	projected, found := phase371SyntheticProjection(synthetic, without, "phase36", "examples/control.schway", "emit")
	if !found || bytes.Equal(projected, newBytes) {
		t.Fatalf("missing-amendment control did not retain the old byte expectation: %q", projected)
	}
}

func phase371SyntheticProjection(census phase371Census, ledger phase371AmendmentLedger, corpus, path, route string) ([]byte, bool) {
	for _, entry := range census.Entries {
		if entry.Corpus != corpus || entry.Path != path {
			continue
		}
		for _, baseline := range entry.Routes {
			if baseline.Name != route {
				continue
			}
			base, err := base64.StdEncoding.DecodeString(baseline.CBase64)
			if err != nil {
				return nil, false
			}
			for _, amendment := range ledger.Entries {
				if phase371RouteKey(amendment.Corpus, amendment.Path, amendment.Route) != phase371RouteKey(corpus, path, route) {
					continue
				}
				projected, err := phase371ApplyEdits(base, amendment.Edits)
				return projected, err == nil
			}
			return base, true
		}
	}
	return nil, false
}

func phase371HasProblem(problems []string, want string) bool {
	for _, problem := range problems {
		if strings.Contains(problem, want) {
			return true
		}
	}
	return false
}
