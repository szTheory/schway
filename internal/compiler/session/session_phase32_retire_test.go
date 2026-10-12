package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32FrontierTombstone is the sha256 of the retired examples/checksum.schway
// (2026-10-12, Phase 32, DX-19, D-05). Git history keeps the bytes.
const phase32FrontierTombstone = "ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154"

func phase32SHA256(b []byte) string {
	digest := sha256.Sum256(b)
	return hex.EncodeToString(digest[:])
}

// phase32TombstoneMatches returns the .schway files under roots whose sha256
// equals tombstone.
func phase32TombstoneMatches(roots []string, tombstone string) ([]string, error) {
	var matches []string
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() || filepath.Ext(path) != ".schway" {
				return nil
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if phase32SHA256(data) == tombstone {
				matches = append(matches, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return matches, nil
}

const phase32OldFrontierPath = "examples/checksum.schway"

var (
	phase32HeadingRE = regexp.MustCompile(`^(#{1,6}) `)
	phase32DatedRE   = regexp.MustCompile(`^#{2,4} .*[0-9]{4}-[0-9]{2}-[0-9]{2}`)
)

// phase32LiveDocViolations reports lines of a live doc that name the retired
// frontier path as if current: not under a dated heading (level 2-4 with an ISO
// date; a heading of the same or higher level ends the dated region) and
// containing neither "retired" nor "amendment".
func phase32LiveDocViolations(name, text string) []string {
	var problems []string
	datedLevel := 0
	for i, line := range strings.Split(text, "\n") {
		if m := phase32HeadingRE.FindStringSubmatch(line); m != nil {
			level := len(m[1])
			if datedLevel != 0 && level <= datedLevel {
				datedLevel = 0
			}
			if datedLevel == 0 && phase32DatedRE.MatchString(line) {
				datedLevel = level
			}
			continue
		}
		if datedLevel != 0 || !strings.Contains(line, phase32OldFrontierPath) {
			continue
		}
		if strings.Contains(line, "retired") || strings.Contains(line, "amendment") {
			continue
		}
		problems = append(problems, fmt.Sprintf("%s:%d: names the retired frontier as current: %s", name, i+1, strings.TrimSpace(line)))
	}
	return problems
}

func TestPhase32ChecksumFrontierRetired(t *testing.T) {
	t.Run("old-path-absent", func(t *testing.T) {
		_, err := os.Stat(testsupport.ProjectPath("examples", "checksum.schway"))
		if !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("examples/checksum.schway must be retired (D-05), stat err = %v", err)
		}
	})
	t.Run("replacement-exists", func(t *testing.T) {
		info, err := os.Stat(testsupport.ProjectPath("examples", "checksum", "checksum.schway"))
		if err != nil {
			t.Fatal(err)
		}
		if !info.Mode().IsRegular() {
			t.Fatal("examples/checksum/checksum.schway is not a regular file")
		}
	})
	t.Run("no-renamed-copy", func(t *testing.T) {
		roots := []string{testsupport.ProjectPath("examples"), testsupport.ProjectPath("testdata")}
		matches, err := phase32TombstoneMatches(roots, phase32FrontierTombstone)
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 0 {
			t.Fatalf("retired frontier bytes reappeared at %v", matches)
		}
	})
	t.Run("armed-renamed-copy", func(t *testing.T) {
		dir := t.TempDir()
		payload := []byte("fn main() {}\n")
		if err := os.WriteFile(filepath.Join(dir, "renamed.schway"), payload, 0o644); err != nil {
			t.Fatal(err)
		}
		matches, err := phase32TombstoneMatches([]string{dir}, phase32SHA256(payload))
		if err != nil {
			t.Fatal(err)
		}
		if len(matches) != 1 {
			t.Fatalf("armed control: want exactly one match, got %v", matches)
		}
	})
	t.Run("live-doc-hygiene", func(t *testing.T) {
		for _, rel := range []string{".planning/PROJECT.md", ".planning/LANGUAGE-MATURITY.md", ".planning/PRODUCT-ROADMAP.md", "examples/checksum/README.md"} {
			data, err := os.ReadFile(testsupport.ProjectPath(strings.Split(rel, "/")...))
			if err != nil {
				t.Fatal(err)
			}
			for _, problem := range phase32LiveDocViolations(rel, string(data)) {
				t.Error(problem)
			}
		}
	})
	t.Run("armed-live-line", func(t *testing.T) {
		text := "## Live heading\n\nThe `" + phase32OldFrontierPath + "` is a refused integration target.\n"
		if got := phase32LiveDocViolations("armed.md", text); len(got) != 1 {
			t.Fatalf("armed control: want one violation, got %v", got)
		}
	})
	t.Run("dated-line-exempt", func(t *testing.T) {
		line := "The `" + phase32OldFrontierPath + "` is a refused integration target.\n"
		dated := "### M005 kickoff - 2026-10-02\n\n" + line
		if got := phase32LiveDocViolations("dated.md", dated); len(got) != 0 {
			t.Fatalf("a line under a dated heading must be exempt, got %v", got)
		}
		moved := "### M005 kickoff - 2026-10-02\n\nbody\n\n## Live heading\n\n" + line
		if got := phase32LiveDocViolations("moved.md", moved); len(got) != 1 {
			t.Fatalf("moving the line under an undated heading must violate, got %v", got)
		}
	})
}
