package session_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase32GateDigest is the codedigest the literal logs are bound to; the gate
// runner passes it as PHASE32_EXPECTED_CODEDIGEST.
var phase32GateDigest = strings.Repeat("ab", 32)

const phase32GateHeadSha = "0123456789abcdef0123456789abcdef01234567"

// phase32GateCase is one entry of the shared case table in the notation of
// scripts/verify-phase32.sh: Stdout and Stderr are "-" for empty.
type phase32GateCase struct {
	Program, Key string
	Exit         int
	Stdout       string
	Stderr       string
}

// phase32GateCases is the literal case list the known-good logs are built
// from. It is written out here, independent of the producer's table;
// TestPhase32ProducerCaseTable proves the two agree.
var phase32GateCases = []phase32GateCase{
	{"checksum", "wikipedia", 0, "300286872", "-"},
}

func phase32ReceiptLog(t *testing.T, host, digest string) string {
	t.Helper()
	return phase32ReceiptLogAt(t, host, digest, phase32GateHeadSha)
}

// phase32ReceiptLogAt builds a literal known-good log for one host. Nothing
// here is read from scripts/verify-phase32.sh, so the gate cannot pass by
// echoing its own producer.
func phase32ReceiptLogAt(t *testing.T, host, digest, headsha string) string {
	t.Helper()
	var lines []string
	lines = append(lines, "phase32 host identity host="+host+" arch=x86_64 go=linux/amd64 clang_target=x86_64-pc-linux-gnu headsha="+headsha+" codedigest="+digest+" tree=clean")
	lines = append(lines, `phase32 toolchain go="go version go1.24.0" clang="clang version"`)
	for _, c := range phase32GateCases {
		lines = append(lines, fmt.Sprintf("phase32 case program=%s input=%s expected_exit=%d actual_exit=%d stdout=exact stderr=exact status=pass", c.Program, c.Key, c.Exit, c.Exit))
	}
	lines = append(lines, fmt.Sprintf("phase32 host result host=%s codedigest=%s tree=clean cases=%d controls=0 programs=%s status=pass",
		host, digest, len(phase32GateCases), strings.Join(phase32ReceiptPrograms, ",")))
	return strings.Join(lines, "\n") + "\n"
}

func runPhase32Gate(t *testing.T, expectedDigest, first, second string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "a.log"), filepath.Join(dir, "b.log")}
	for index, content := range []string{first, second} {
		if content == "\x00missing" {
			continue
		}
		if err := os.WriteFile(paths[index], []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	environment := append(os.Environ(), "PHASE32_EXPECTED_CODEDIGEST="+expectedDigest)
	result, err := testsupport.RunCLIErr(context.Background(), "sh", environment, testsupport.ProjectPath("scripts", "check-phase32-receipts.sh"), paths[0], paths[1])
	if err != nil {
		return "", err
	}
	output := string(result.Stdout) + string(result.Stderr)
	if result.Exit != 0 {
		return output, errors.New("gate rejected the receipt pair")
	}
	return output, nil
}

func TestPhase32ReceiptGate(t *testing.T) {
	mac := phase32ReceiptLog(t, "macOS", phase32GateDigest)
	linux := phase32ReceiptLog(t, "Linux", phase32GateDigest)

	t.Run("known-good", func(t *testing.T) {
		output, err := runPhase32Gate(t, phase32GateDigest, mac, linux)
		if err != nil || !strings.Contains(output, "status=pass hosts=macOS,Linux") {
			t.Fatalf("known-good pair rejected: err=%v output=%s", err, output)
		}
	})

	t.Run("digest-mismatch", func(t *testing.T) {
		other := strings.Repeat("cd", 32)
		output, err := runPhase32Gate(t, phase32GateDigest, phase32ReceiptLog(t, "macOS", other), phase32ReceiptLog(t, "Linux", other))
		if err == nil || !strings.Contains(output, "status=fail") || !strings.Contains(output, "digest_mismatch") {
			t.Fatalf("mutated digest was not rejected with digest_mismatch: err=%v output=%s", err, output)
		}
		if strings.Contains(output, "status=pass") {
			t.Fatalf("rejected pair also reported pass: %s", output)
		}
	})
}
