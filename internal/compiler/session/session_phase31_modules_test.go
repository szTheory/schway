package session_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func phase31ModulesDigest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func phase31ModulesNeedClang(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}
}

// phase31ModulesBuild builds entry with an optional manifest and returns the
// artifact path and the receipt read back from disk.
func phase31ModulesBuild(t *testing.T, entry string, manifest ...string) (string, native.BuildReceipt) {
	t.Helper()
	artifact := filepath.Join(t.TempDir(), "app")
	receipt, diagnostics, err := session.BuildApplicationFile(context.Background(), entry, artifact, native.DefaultRunner(), manifest...)
	if err != nil || len(diagnostics) > 0 {
		t.Fatalf("build %s: err=%v diagnostics=%+v", entry, err, diagnostics)
	}
	data, err := os.ReadFile(artifact + ".schway-build.json")
	if err != nil {
		t.Fatal(err)
	}
	var disk native.BuildReceipt
	if err := json.Unmarshal(data, &disk); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(disk.Modules, receipt.Modules) {
		t.Fatalf("returned receipt modules %+v differ from disk %+v", receipt.Modules, disk.Modules)
	}
	return artifact, disk
}

func TestPhase31ModuleDigests(t *testing.T) {
	t.Run("checksum-receipt", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		dir := testsupport.ProjectPath("examples", "checksum")
		_, receipt := phase31ModulesBuild(t, filepath.Join(dir, "checksum.schway"), filepath.Join(dir, "file_bytes.bindings.json"))
		library, err := os.ReadFile(filepath.Join(dir, "adler32.schway"))
		if err != nil {
			t.Fatal(err)
		}
		want := []native.ModuleDigest{
			{Module: "adler32", Path: "adler32.schway", Digest: phase31ModulesDigest(library)},
			{Module: "examples.checksum.checksum", Path: "checksum.schway", Digest: receipt.SourceDigest},
		}
		if !reflect.DeepEqual(receipt.Modules, want) {
			t.Fatalf("modules = %+v, want %+v", receipt.Modules, want)
		}
		if !reflect.DeepEqual(receipt.Identity.Modules, receipt.Modules) {
			t.Fatalf("identity modules %+v != receipt modules %+v", receipt.Identity.Modules, receipt.Modules)
		}
	})
}
