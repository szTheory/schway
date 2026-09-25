package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/session"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestPhase20ChecksumFrontier(t *testing.T) {
	path := testsupport.ProjectPath("examples", "checksum.lang")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := phase20SHA256(source); got != "ae95e550df67114c3464dda9cc24779f8a2efd69bd2b99dd0ccb2acc235cc154" {
		t.Fatalf("checksum fixture digest changed: got %s", got)
	}
	for _, witness := range []string{"checksum_read(path)", "loop buffer", "next(buffer)", "add(total, current_byte)", "checksum_print(total)"} {
		if !strings.Contains(string(source), witness) {
			t.Errorf("checksum intent witness %q missing from %s", witness, path)
		}
	}
	checked := session.Check(source)
	if len(checked.Diagnostics) == 0 {
		t.Fatal("checksum fixture unexpectedly passed production checking")
	}
	first := checked.Diagnostics[0]
	if first.Code != "syntax.expected_rbrace" || first.Primary.Start != 521 || first.Primary.End != 527 {
		t.Fatalf("current checksum refusal moved: got code=%q span=%+v", first.Code, first.Primary)
	}
}

func phase20SHA256(source []byte) string {
	digest := sha256.Sum256(source)
	return hex.EncodeToString(digest[:])
}
