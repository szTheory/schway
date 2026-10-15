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
	// This is the canonical SUMMARY indentation/scalar subset, not a general
	// YAML parser. Preserve the consumer's scalar types for required fields:
	// quoted booleans/null are strings; bare booleans/null and [] are not.
	stringValue := func(raw string) string {
		if len(raw) >= 2 && ((raw[0] == '"' && raw[len(raw)-1] == '"') || (raw[0] == '\'' && raw[len(raw)-1] == '\'')) {
			return raw[1 : len(raw)-1]
		}
		switch raw {
		case "", "true", "false", "null", "~", "[]":
			return ""
		}
		return raw
	}
	field := func(text string) (string, string, bool) {
		key, raw, ok := strings.Cut(text, ":")
		if !ok || key == "" {
			return "", "", false
		}
		for _, r := range key {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-') {
				return "", "", false
			}
		}
		return key, strings.TrimSpace(raw), true
	}
	type verification struct {
		kind, ref, status string
		mapping           bool
	}
	type entry struct {
		id, description, judgment, rationale       string
		mapping, hasVerification, verificationList bool
		verification                               []*verification
	}
	var entries []*entry
	inVerification, verificationBodyStarted := false, false
	for _, line := range lines[start:] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(line, " ") {
			break
		}
		indent := len(line) - len(strings.TrimLeft(line, " "))
		item := strings.HasPrefix(trimmed, "- ") || trimmed == "-"
		if item && (indent == 2 || indent == 6) {
			trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, "-"))
		}
		key, raw, mapping := field(trimmed)
		if item && mapping {
			_, inline, _ := strings.Cut(trimmed, ":")
			mapping = inline == "" || strings.HasPrefix(inline, " ")
		}
		entryStart := indent == 2 && item
		if entryStart {
			// Every item starts an entry, even if id is missing or not first.
			entries = append(entries, &entry{mapping: mapping || trimmed == ""})
			inVerification, verificationBodyStarted = false, false
		}
		if len(entries) == 0 {
			return []string{"coverage: block is not a list of entries"}
		}
		e := entries[len(entries)-1]
		if entryStart || indent == 4 {
			inVerification = false
			switch key {
			case "id":
				e.id = stringValue(raw)
			case "description":
				e.description = stringValue(raw)
			case "human_judgment":
				e.judgment = raw // Only bare true/false are booleans.
			case "rationale":
				e.rationale = stringValue(raw)
			case "verification":
				e.hasVerification = true
				e.verificationList = raw == "[]"
				e.verification = nil
				inVerification = raw == ""
				verificationBodyStarted = false
			}
			continue
		}
		if !inVerification {
			continue
		}
		if !verificationBodyStarted {
			verificationBodyStarted = true
			e.verificationList = indent == 6 && item
		}
		if !e.verificationList {
			continue
		}
		verificationStart := indent == 6 && item
		if verificationStart {
			e.verification = append(e.verification, &verification{mapping: mapping || trimmed == ""})
		}
		if len(e.verification) == 0 || !verificationStart && indent != 8 {
			continue
		}
		v := e.verification[len(e.verification)-1]
		switch key {
		case "kind":
			v.kind = stringValue(raw)
		case "ref":
			v.ref = stringValue(raw)
		case "status":
			v.status = stringValue(raw)
		}
	}

	if len(entries) == 0 {
		return []string{"coverage: block has no entries"}
	}
	var problems []string
	seenIDs := make(map[string]bool)
	for _, e := range entries {
		if !e.mapping {
			problems = append(problems, "coverage entry is not a mapping")
			continue
		}
		// Required entry fields mirror coverage.cjs validateEntry, including
		// uniqueness within this SUMMARY rather than across separate files.
		if strings.TrimSpace(e.id) == "" {
			problems = append(problems, "coverage entry has no non-empty string id")
		} else if seenIDs[e.id] {
			problems = append(problems, e.id+": duplicate coverage id")
		} else {
			seenIDs[e.id] = true
		}
		if strings.TrimSpace(e.description) == "" {
			problems = append(problems, e.id+": no non-empty string description")
		}
		if e.hasVerification && !e.verificationList {
			problems = append(problems, e.id+": verification is not a list")
		}
		for i, v := range e.verification {
			item := e.id + ": verification[" + strconv.Itoa(i) + "]"
			if !v.mapping {
				problems = append(problems, item+": verification item is not a mapping")
				continue
			}
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
			if strings.TrimSpace(e.rationale) == "" {
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
	good := "---\ncoverage:\n  - id: D1\n    description: A known test grounds the deliverable\n" + verification + "    human_judgment: false\nstatus: complete\n---\n"
	human := strings.Replace(good, "human_judgment: false", "human_judgment: true\n    rationale: Inspect the visual result", 1)
	entry := strings.TrimSuffix(strings.TrimPrefix(good, "---\ncoverage:\n"), "status: complete\n---\n")
	leadingEntry := func(extra string) string {
		return strings.Replace(good, "coverage:\n", "coverage:\n"+extra, 1)
	}
	valid := []struct{ name, markdown string }{
		{"automated", good},
		{"reordered-entry", strings.Replace(good, "  - id: D1\n    description: A known test grounds the deliverable", "  - description: A known test grounds the deliverable\n    id: D1", 1)},
		{"multiple-entries", leadingEntry(strings.Replace(entry, "id: D1", "id: D2", 1))},
		{"quoted-string-fields", strings.Replace(strings.Replace(strings.Replace(good, "id: D1", "id: \"false\"", 1), "description: A known test grounds the deliverable", "description: 'true'", 1), "ref: \"a_test.go#TestReal\"", "ref: 'null'", 1)},
		{"empty-coverage", "---\ncoverage: []\nstatus: complete\n---\n"},
		{"human", human},
		{"human-no-verification", strings.Replace(human, verification, "", 1)},
		{"human-empty-verification-list", strings.Replace(human, verification, "    verification: []\n", 1)},
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
		{"missing-id", strings.Replace(good, "  - id: D1\n", "  -\n", 1)},
		{"non-mapping-entry-head-without-space", strings.Replace(good, "- id: D1", "- id:D1", 1)},
		{"description-as-list-item", strings.Replace(good, "    description:", "    - description:", 1)},
		{"judgment-as-list-item", strings.Replace(good, "    human_judgment:", "    - human_judgment:", 1)},
		{"non-mapping-item-head-without-space", strings.Replace(good, "- kind: unit", "- kind:unit", 1)},
		{"status-as-list-item", strings.Replace(good, "        status:", "        - status:", 1)},
		{"empty-id", strings.Replace(good, "id: D1", "id: \"\"", 1)},
		{"whitespace-id", strings.Replace(good, "id: D1", "id: \"   \"", 1)},
		{"duplicate-id", leadingEntry(entry)},
		{"missing-description", strings.Replace(good, "    description: A known test grounds the deliverable\n", "", 1)},
		{"empty-description", strings.Replace(good, "description: A known test grounds the deliverable", "description: \"\"", 1)},
		{"whitespace-description", strings.Replace(good, "description: A known test grounds the deliverable", "description: \"   \"", 1)},
		{"missing-id-entry-starting-description", leadingEntry(strings.Replace(entry, "  - id: D1\n    description:", "  - description:", 1))},
		{"non-mapping-entry-before-valid", leadingEntry("  - null\n")},
		{"non-mapping-entry-after-valid", strings.Replace(good, "status: complete\n", "  - null\nstatus: complete\n", 1)},
		{"boolean-id", strings.Replace(good, "id: D1", "id: true", 1)},
		{"null-id", strings.Replace(good, "id: D1", "id: null", 1)},
		{"array-id", strings.Replace(good, "id: D1", "id: []", 1)},
		{"boolean-description", strings.Replace(good, "description: A known test grounds the deliverable", "description: true", 1)},
		{"null-description", strings.Replace(good, "description: A known test grounds the deliverable", "description: null", 1)},
		{"array-description", strings.Replace(good, "description: A known test grounds the deliverable", "description: []", 1)},
		{"quoted-judgment", strings.Replace(good, "human_judgment: false", "human_judgment: \"false\"", 1)},
		{"human-quoted-judgment", strings.Replace(human, "human_judgment: true", "human_judgment: 'true'", 1)},
		{"whitespace-rationale", strings.Replace(human, "rationale: Inspect the visual result", "rationale: \"   \"", 1)},
		{"boolean-rationale", strings.Replace(human, "rationale: Inspect the visual result", "rationale: true", 1)},
		{"null-rationale", strings.Replace(human, "rationale: Inspect the visual result", "rationale: null", 1)},
		{"array-rationale", strings.Replace(human, "rationale: Inspect the visual result", "rationale: []", 1)},
		{"human-scalar-verification", strings.Replace(human, verification, "    verification: manual\n", 1)},
		{"human-null-verification", strings.Replace(human, verification, "    verification: null\n", 1)},
		{"human-empty-verification-block", strings.Replace(human, verification, "    verification:\n", 1)},
		{"human-mapping-verification", strings.Replace(human, verification, "    verification:\n      ref: a_test.go#TestReal\n", 1)},
		{"non-list-verification-with-items", strings.Replace(good, "verification:\n", "verification: manual\n", 1)},
		{"non-mapping-item-with-fields", strings.Replace(good, "      - kind: unit\n", "      - null\n        kind: unit\n", 1)},
		{"non-mapping-item-before-valid", strings.Replace(good, "verification:\n", "verification:\n      - null\n", 1)},
		{"non-mapping-item-after-valid", strings.Replace(good, "        status: pass\n", "        status: pass\n      - null\n", 1)},
		{"boolean-ref", strings.Replace(good, "ref: \"a_test.go#TestReal\"", "ref: true", 1)},
		{"null-ref", strings.Replace(good, "ref: \"a_test.go#TestReal\"", "ref: null", 1)},
		{"array-ref", strings.Replace(good, "ref: \"a_test.go#TestReal\"", "ref: []", 1)},
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
