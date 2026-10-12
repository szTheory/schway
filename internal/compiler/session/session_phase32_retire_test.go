package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
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
}
