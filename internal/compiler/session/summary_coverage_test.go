package session_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// summaryCoverageFirstPhase is the first phase whose SUMMARYs must carry a
// `coverage:` block (PROJECT.md § Verification Operating Preference, settled
// 2026-10-12 after Phase 31 UAT). With the block, `/gsd-verify-work`
// classifies each deliverable deterministically and auto-passes it from its
// named tests; without it the agent re-derives checkpoints from prose. Phases
// before 32 are grandfathered, not rewritten.
const summaryCoverageFirstPhase = 32

// summaryCoverageProblems returns why a SUMMARY's coverage block cannot be
// auto-classified, or nil when every entry is either test-backed
// (human_judgment: false, at least one passing ref naming a known test) or an
// explicit human checkpoint with a rationale. knownTests nil skips name
// resolution.
func summaryCoverageProblems(markdown string, knownTests map[string]bool) []string {
	if !strings.HasPrefix(markdown, "---\n") {
		return []string{"no frontmatter"}
	}
	end := strings.Index(markdown[4:], "\n---")
	if end < 0 {
		return []string{"unterminated frontmatter"}
	}
	lines := strings.Split(markdown[4:4+end], "\n")
	start := -1
	for i, line := range lines {
		if line == "coverage: []" {
			return nil
		}
		if line == "coverage:" {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return []string{"missing coverage: block"}
	}
	type verification struct {
		kind, ref, status string
	}
	type entry struct {
		id, judgment, rationale string
		verification            []*verification
	}
	var entries []*entry
	inVerification := false
	for _, line := range lines[start:] {
		if line != "" && !strings.HasPrefix(line, " ") {
			break
		}
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		key, value, _ := strings.Cut(trimmed, ":")
		value = strings.Trim(strings.TrimSpace(value), `"`)
		if strings.HasPrefix(line, "  - id:") {
			entries = append(entries, &entry{id: value})
			inVerification = false
			continue
		}
		if len(entries) == 0 {
			continue
		}
		e := entries[len(entries)-1]
		if strings.HasPrefix(line, "    verification:") {
			inVerification = true
			continue
		}
		if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "      ") {
			inVerification = false
		}
		// The SUMMARY subset uses six-space verification list items and
		// eight-space continuation fields. Start a record for any item key,
		// so an item beginning with ref or status still has its own fields.
		if inVerification {
			if strings.HasPrefix(line, "      - ") || line == "      -" {
				e.verification = append(e.verification, &verification{})
			}
			if len(e.verification) != 0 {
				v := e.verification[len(e.verification)-1]
				switch key {
				case "kind":
					v.kind = value
				case "ref":
					v.ref = value
				case "status":
					v.status = value
				}
			}
			continue
		}
		switch key {
		case "human_judgment":
			e.judgment = value
		case "rationale":
			e.rationale = value
		}
	}
	if len(entries) == 0 {
		return []string{"coverage: block has no entries"}
	}
	var problems []string
	for _, e := range entries {
		for i, v := range e.verification {
			item := e.id + ": verification[" + strconv.Itoa(i) + "]"
			// This enum mirrors gsd-core/bin/lib/coverage.cjs:65-67 (VALID_KINDS).
			switch v.kind {
			case "unit", "integration", "e2e", "automated_ui", "manual_procedural", "other":
			default:
				problems = append(problems, item+": unknown verification kind "+v.kind)
			}
			// Schema-valid human checkpoints may record fail or unknown;
			// automated entries additionally require pass below.
			switch v.status {
			case "pass", "fail", "unknown":
			default:
				problems = append(problems, item+": unknown verification status "+v.status)
			}
			if strings.TrimSpace(v.ref) == "" {
				problems = append(problems, item+": no verification ref")
			}
		}
		switch e.judgment {
		case "true":
			if e.rationale == "" {
				problems = append(problems, e.id+": human_judgment true without rationale")
			}
		case "false":
			if len(e.verification) == 0 {
				problems = append(problems, e.id+": no verification ref")
			}
			for _, v := range e.verification {
				if v.status != "pass" {
					problems = append(problems, e.id+": verification status "+v.status)
				}
				_, name, ok := strings.Cut(v.ref, "#")
				if !ok || !strings.HasPrefix(name, "Test") {
					continue // a command invocation, not a test name
				}
				name, _, _ = strings.Cut(name, "/")
				if knownTests != nil && !knownTests[name] {
					problems = append(problems, e.id+": unknown test "+name)
				}
			}
		default:
			problems = append(problems, e.id+": human_judgment must be true or false")
		}
	}
	return problems
}

func TestSummaryCoverageGroundedness(t *testing.T) {
	known := map[string]bool{"TestReal": true}
	verification := "    verification:\n      - kind: unit\n        ref: \"a_test.go#TestReal\"\n        status: pass\n"
	good := "---\ncoverage:\n  - id: D1\n" + verification + "    human_judgment: false\nstatus: complete\n---\n"
	human := strings.Replace(good, "human_judgment: false", "human_judgment: true\n    rationale: Inspect the visual result", 1)
	valid := []struct{ name, markdown string }{
		{"automated", good},
		{"empty-coverage", "---\ncoverage: []\nstatus: complete\n---\n"},
		{"human", human},
		{"human-no-verification", strings.Replace(human, verification, "", 1)},
		{"human-fail", strings.Replace(human, "status: pass", "status: fail", 1)},
		{"human-unknown", strings.Replace(human, "status: pass", "status: unknown", 1)},
		{"reordered-item", strings.Replace(good, verification, "    verification:\n      - ref: a_test.go#TestReal\n        status: pass\n        kind: unit\n", 1)},
		{"multiple-items", strings.Replace(good, "        status: pass\n", "        status: pass\n      - ref: go test ./...\n        status: pass\n        kind: other\n", 1)},
	}
	for _, control := range valid {
		t.Run("valid/"+control.name, func(t *testing.T) {
			if p := summaryCoverageProblems(control.markdown, known); len(p) != 0 {
				t.Fatalf("seeded valid block rejected: %v", p)
			}
		})
	}
	invalid := []struct{ name, markdown string }{
		{"missing", "---\nstatus: complete\n---\n"},
		{"unknown-test", strings.Replace(good, "TestReal", "TestGhost", 1)},
		{"failing", strings.Replace(good, "status: pass", "status: fail", 1)},
		{"unknown-status", strings.Replace(good, "status: pass", "status: unknown", 1)},
		{"no-rationale", strings.Replace(good, "human_judgment: false", "human_judgment: true", 1)},
		{"no-judgment", strings.Replace(good, "    human_judgment: false\n", "", 1)},
		{"bad-kind", strings.Replace(good, "kind: unit", "kind: manual", 1)},
		{"missing-kind", strings.Replace(good, "      - kind: unit\n        ref:", "      - ref:", 1)},
		{"missing-status", strings.Replace(good, "        status: pass\n", "", 1)},
		{"missing-ref", strings.Replace(good, "        ref: \"a_test.go#TestReal\"\n", "", 1)},
		{"empty-kind", strings.Replace(good, "kind: unit", "kind: \"\"", 1)},
		{"empty-status", strings.Replace(good, "status: pass", "status: \"\"", 1)},
		{"empty-ref", strings.Replace(good, "ref: \"a_test.go#TestReal\"", "ref: \"\"", 1)},
		{"invalid-status", strings.Replace(good, "status: pass", "status: pending", 1)},
		{"no-verification", strings.Replace(good, verification, "", 1)},
		// These three malformed items have equal aggregate kind/ref/status
		// counts. Fields must belong to the same item, not merely the entry.
		{"split-malformed-items", strings.Replace(good, verification, "    verification:\n      - kind: unit\n      - ref: a_test.go#TestReal\n      - status: pass\n", 1)},
		{"missing-ref-with-valid-item", strings.Replace(good, "        status: pass\n", "        status: pass\n      - kind: unit\n        status: pass\n", 1)},
		{"human-missing-kind", strings.Replace(human, "      - kind: unit\n        ref:", "      - ref:", 1)},
		{"human-missing-status", strings.Replace(human, "        status: pass\n", "", 1)},
		{"human-missing-ref", strings.Replace(human, "        ref: \"a_test.go#TestReal\"\n", "", 1)},
		{"human-invalid-status", strings.Replace(human, "status: pass", "status: pending", 1)},
	}
	for _, control := range invalid {
		t.Run("invalid/"+control.name, func(t *testing.T) {
			if len(summaryCoverageProblems(control.markdown, known)) == 0 {
				t.Fatal("seeded invalid block was accepted")
			}
		})
	}
	t.Logf("summary coverage controls: %d valid and %d invalid reached", len(valid), len(invalid))

	index := buildTestIndex(t)
	tests := make(map[string]bool)
	for _, names := range index.byImportPath {
		for name := range names {
			tests[name] = true
		}
	}
	summaries, err := phaseArtifactGlob("*", "*-SUMMARY.md")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, path := range summaries {
		prefix, _, _ := strings.Cut(filepath.Base(filepath.Dir(path)), "-")
		phase, err := strconv.ParseFloat(prefix, 64)
		if err != nil || phase < summaryCoverageFirstPhase {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		checked++
		if p := summaryCoverageProblems(string(data), tests); len(p) != 0 {
			t.Errorf("%s: %s", path, strings.Join(p, "; "))
		}
	}
	t.Logf("summary coverage: %d SUMMARYs from phase %d on checked", checked, summaryCoverageFirstPhase)
}
