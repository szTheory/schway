package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// Phase 31 D-01: the Adler-32 library and the Phase 30 CLI moved into
// examples/checksum/. These pins keep the moved bytes from drifting.

const phase31LibrarySHA256 = "317ac13972a53eccd04a7ab84cc914aff755e77b881db6fcaf77fa5240ebcb4b"

func phase31MoveDigest(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

func TestPhase31MovedLibraryIdentity(t *testing.T) {
	lib, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "adler32.schway"))
	if err != nil {
		t.Fatalf("read moved library: %v", err)
	}
	if got := phase31MoveDigest(lib); got != phase31LibrarySHA256 {
		t.Fatalf("moved library sha256 = %s, want %s", got, phase31LibrarySHA256)
	}

	cli, err := os.ReadFile(testsupport.ProjectPath("examples", "checksum", "adler32_cli.schway"))
	if err != nil {
		t.Fatalf("read moved CLI: %v", err)
	}
	header := ""
	for _, line := range strings.Split(string(cli), "\n") {
		if strings.HasPrefix(line, "module ") {
			header = line
			break
		}
	}
	if header != "module examples.checksum.adler32_cli" {
		t.Fatalf("moved CLI header = %q", header)
	}

	if _, err := os.Stat(testsupport.ProjectPath("examples", "phase30")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("old examples/phase30 directory must not exist, stat err = %v", err)
	}

	t.Run("non-inert", func(t *testing.T) {
		mutated := append([]byte(nil), lib...)
		mutated[len(mutated)/2] ^= 0x01
		if phase31MoveDigest(mutated) == phase31LibrarySHA256 {
			t.Fatal("a flipped byte did not change the digest; the control is inert")
		}
	})
}
