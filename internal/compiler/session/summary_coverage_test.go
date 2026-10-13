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
	type entry struct {
		id, judgment, rationale string
		refs, statuses, kinds   []string
	}
	var entries []*entry
	for _, line := range lines[start:] {
		if line != "" && !strings.HasPrefix(line, " ") {
			break
		}
		trimmed := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- "))
		key, value, _ := strings.Cut(trimmed, ":")
		value = strings.Trim(strings.TrimSpace(value), `"`)
		if strings.HasPrefix(line, "  - id:") {
			entries = append(entries, &entry{id: value})
			continue
		}
		if len(entries) == 0 {
			continue
		}
		e := entries[len(entries)-1]
		switch key {
		case "human_judgment":
			e.judgment = value
		case "rationale":
			e.rationale = value
		case "ref":
			e.refs = append(e.refs, value)
		case "status":
			e.statuses = append(e.statuses, value)
		case "kind":
			e.kinds = append(e.kinds, value)
		}
	}
	if len(entries) == 0 {
		return []string{"coverage: block has no entries"}
	}
	var problems []string
	for _, e := range entries {
		// This enum mirrors gsd-core/bin/lib/coverage.cjs:65-67 (VALID_KINDS).
		for _, kind := range e.kinds {
			switch kind {
			case "unit", "integration", "e2e", "automated_ui", "manual_procedural", "other":
			default:
				problems = append(problems, e.id+": unknown verification kind "+kind)
			}
		}
		switch e.judgment {
		case "true":
			if e.rationale == "" {
				problems = append(problems, e.id+": human_judgment true without rationale")
			}
		case "false":
			if len(e.refs) == 0 {
				problems = append(problems, e.id+": no verification ref")
			}
			for _, s := range e.statuses {
				if s != "pass" {
					problems = append(problems, e.id+": verification status "+s)
				}
			}
			for _, ref := range e.refs {
				_, name, ok := strings.Cut(ref, "#")
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
	good := "---\ncoverage:\n  - id: D1\n    verification:\n      - kind: unit\n        ref: \"a_test.go#TestReal\"\n        status: pass\n    human_judgment: false\nstatus: complete\n---\n"
	if p := summaryCoverageProblems(good, known); len(p) != 0 {
		t.Fatalf("seeded good block rejected: %v", p)
	}
	for name, bad := range map[string]string{
		"missing":      "---\nstatus: complete\n---\n",
		"unknown-test": strings.Replace(good, "TestReal", "TestGhost", 1),
		"failing":      strings.Replace(good, "status: pass", "status: fail", 1),
		"no-rationale": strings.Replace(good, "human_judgment: false", "human_judgment: true", 1),
		"no-judgment":  strings.Replace(good, "    human_judgment: false\n", "", 1),
		"bad-kind":     strings.Replace(good, "kind: unit", "kind: manual", 1),
	} {
		if len(summaryCoverageProblems(bad, known)) == 0 {
			t.Fatalf("seeded %s control was accepted", name)
		}
	}

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
