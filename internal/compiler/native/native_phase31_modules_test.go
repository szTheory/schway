package native_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/native"
	"github.com/szTheory/schway/internal/compiler/session"
	"github.com/szTheory/schway/internal/compiler/testsupport"
)

// phase31ModulesIdentityMirror is BuildIdentityInputs minus Modules.
type phase31ModulesIdentityMirror struct {
	SourceDigest        string                 `json:"source_digest"`
	EmittedCDigest      string                 `json:"emitted_c_digest"`
	ManifestDigest      string                 `json:"manifest_digest"`
	Inventory           []native.BindingInput  `json:"inventory"`
	Symbols             []native.BindingSymbol `json:"symbols"`
	IncludedHeaders     []string               `json:"included_headers"`
	ABIProbeDigest      string                 `json:"abi_probe_digest"`
	CompilerDigest      string                 `json:"compiler_digest"`
	CompilerVersion     string                 `json:"compiler_version"`
	Target              string                 `json:"target"`
	HostABI             string                 `json:"host_abi"`
	Flags               []string               `json:"flags"`
	Commands            [][]string             `json:"commands"`
	RuntimeDependencies []string               `json:"runtime_dependencies"`
	ExecutableDigest    string                 `json:"executable_digest"`
}

// phase31ModulesIdentityNoOmit carries Modules WITHOUT omitempty.
type phase31ModulesIdentityNoOmit struct {
	phase31ModulesIdentityMirror
	Modules []native.ModuleDigest `json:"modules"`
}

func phase31ModulesBuildChecksum(t *testing.T) (string, native.BuildReceipt) {
	t.Helper()
	if _, err := exec.LookPath("clang"); err != nil {
		t.Skipf("env:clang toolchain unavailable on this host: %v", err)
	}
	dir := testsupport.ProjectPath("examples", "checksum")
	artifact := filepath.Join(t.TempDir(), "app")
	receipt, diagnostics, err := session.BuildApplicationFile(context.Background(), filepath.Join(dir, "checksum.schway"), artifact, native.DefaultRunner(), filepath.Join(dir, "file_bytes.bindings.json"))
	if err != nil || len(diagnostics) > 0 {
		t.Fatalf("build: err=%v diagnostics=%+v", err, diagnostics)
	}
	return artifact, receipt
}

func TestPhase31ModuleDigests(t *testing.T) {
	t.Run("omitempty-invariance", func(t *testing.T) {
		in := native.BuildIdentityInputs{
			SourceDigest: "a", EmittedCDigest: "b", CompilerDigest: "c", CompilerVersion: "v", Target: "t", HostABI: "h",
			Flags: []string{"-O0"}, Commands: [][]string{{"clang"}}, RuntimeDependencies: []string{"platform-c-runtime"}, ExecutableDigest: "e",
		}
		got, _ := json.Marshal(in)
		mirror := phase31ModulesIdentityMirror{
			SourceDigest: in.SourceDigest, EmittedCDigest: in.EmittedCDigest, CompilerDigest: in.CompilerDigest,
			CompilerVersion: in.CompilerVersion, Target: in.Target, HostABI: in.HostABI, Flags: in.Flags,
			Commands: in.Commands, RuntimeDependencies: in.RuntimeDependencies, ExecutableDigest: in.ExecutableDigest,
		}
		want, _ := json.Marshal(mirror)
		if !bytes.Equal(got, want) {
			t.Fatalf("nil Modules changed bytes:\n got %s\nwant %s", got, want)
		}
		control, _ := json.Marshal(phase31ModulesIdentityNoOmit{phase31ModulesIdentityMirror: mirror})
		if bytes.Equal(control, want) {
			t.Fatal("control without omitempty did not differ; the invariance check is vacuous")
		}
	})

	t.Run("receipt-validation", func(t *testing.T) {
		artifact, receipt := phase31ModulesBuildChecksum(t)
		receiptPath := artifact + ".schway-build.json"
		original, err := os.ReadFile(receiptPath)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := native.DefaultRunner().RunApplication(context.Background(), artifact, "x", &bytes.Buffer{}, &bytes.Buffer{}); err != nil && strings.Contains(err.Error(), "receipt") {
			t.Fatalf("untampered receipt refused: %v", err)
		}
		cases := []struct {
			name   string
			mutate func(r *native.BuildReceipt)
			want   string
		}{
			{"one-row", func(r *native.BuildReceipt) { r.Modules = r.Modules[1:] }, "at least one import"},
			{"unsorted", func(r *native.BuildReceipt) { r.Modules[0], r.Modules[1] = r.Modules[1], r.Modules[0] }, "unsorted or duplicated"},
			{"duplicate", func(r *native.BuildReceipt) { r.Modules[1] = r.Modules[0] }, "unsorted or duplicated"},
			{"bad-digest", func(r *native.BuildReceipt) { r.Modules[0].Digest = "xyz" }, "invalid name or digest"},
			{"absolute-path", func(r *native.BuildReceipt) { r.Modules[0].Path = "/etc/passwd" }, "relative"},
			{"dotdot-path", func(r *native.BuildReceipt) { r.Modules[0].Path = "../adler32.schway" }, ".."},
			{"empty-path", func(r *native.BuildReceipt) { r.Modules[0].Path = "" }, "relative"},
			{"no-entry-link", func(r *native.BuildReceipt) { r.Modules[1].Digest = r.Modules[0].Digest }, "entry source digest"},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				tampered := receipt
				tampered.Modules = append([]native.ModuleDigest(nil), receipt.Modules...)
				tc.mutate(&tampered)
				identity := *receipt.Identity
				identity.Modules = tampered.Modules
				tampered.Identity = &identity
				tampered.BuildID, tampered.InputID = identity.ID(), identity.InputID()
				writePhase31ModulesReceipt(t, receiptPath, tampered)
				_, err := native.DefaultRunner().RunApplication(context.Background(), artifact, "x", &bytes.Buffer{}, &bytes.Buffer{})
				if err == nil || !strings.Contains(err.Error(), "native.receipt_invalid") || !strings.Contains(err.Error(), tc.want) {
					t.Fatalf("err = %v, want receipt_invalid containing %q", err, tc.want)
				}
			})
		}
		t.Run("identity-differs", func(t *testing.T) {
			tampered := receipt
			tampered.Modules = append([]native.ModuleDigest(nil), receipt.Modules...)
			tampered.Modules[0].Path = "other.schway"
			writePhase31ModulesReceipt(t, receiptPath, tampered)
			_, err := native.DefaultRunner().RunApplication(context.Background(), artifact, "x", &bytes.Buffer{}, &bytes.Buffer{})
			if err == nil || !strings.Contains(err.Error(), "native.receipt_invalid") {
				t.Fatalf("err = %v, want receipt_invalid", err)
			}
		})
		if err := os.WriteFile(receiptPath, original, 0o600); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("receipt-size", func(t *testing.T) {
		artifact, _ := phase31ModulesBuildChecksum(t)
		info, err := os.Stat(artifact + ".schway-build.json")
		if err != nil {
			t.Fatal(err)
		}
		if info.Size() >= 64*1024 {
			t.Fatalf("receipt size %d >= 64 KiB", info.Size())
		}
	})
}

func writePhase31ModulesReceipt(t *testing.T, path string, receipt native.BuildReceipt) {
	t.Helper()
	data, err := json.Marshal(receipt)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}
