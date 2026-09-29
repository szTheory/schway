package session_test

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase20ChecksumFrontier(t *testing.T) {
	path := testsupport.ProjectPath("examples", "checksum.schway")
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
	// This is a historical reconstruction, not a contemporaneous M003-open
	// fixture pin. The baseline checker at d21db90 refused these exact bytes
	// at the numeric literal; current checking reaches the unsupported loop.
	const historicalRevision = "d21db90e67750bb19976c4206f4c23c68cd06207"
	historical := struct {
		code       string
		start, end int
	}{code: "syntax.unexpected_byte", start: 512, end: 513}
	if first.Code == historical.code && first.Primary.Start == historical.start && first.Primary.End == historical.end {
		t.Fatalf("checksum refusal did not move from M003-open %s (%s) at %d-%d", historicalRevision, historical.code, historical.start, historical.end)
	}
}

func phase20SHA256(source []byte) string {
	digest := sha256.Sum256(source)
	return hex.EncodeToString(digest[:])
}
