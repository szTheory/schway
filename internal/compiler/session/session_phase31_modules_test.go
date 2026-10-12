package session_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/modload"
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

	t.Run("comment-only-library-edit", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		pristine := phase31ModulesCopyChecksum(t)
		edited := phase31ModulesCopyChecksum(t)
		library := filepath.Join(edited, "adler32.schway")
		data, err := os.ReadFile(library)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(library, append([]byte("// phase31 comment-only control\n"), data...), 0o600); err != nil {
			t.Fatal(err)
		}
		_, a := phase31ModulesBuild(t, filepath.Join(pristine, "checksum.schway"), filepath.Join(pristine, "file_bytes.bindings.json"))
		_, b := phase31ModulesBuild(t, filepath.Join(edited, "checksum.schway"), filepath.Join(edited, "file_bytes.bindings.json"))
		if a.Modules[0].Digest == b.Modules[0].Digest {
			t.Fatal("adler32 row digest did not change")
		}
		if a.BuildID == b.BuildID || a.InputID == b.InputID {
			t.Fatal("BuildID/InputID did not change (InputID excludes the executable digest, so the module row alone moved it)")
		}
		if a.EmittedCDigest != b.EmittedCDigest || a.SourceDigest != b.SourceDigest || a.Modules[1] != b.Modules[1] {
			t.Fatalf("emitted C / source / entry row moved: %+v vs %+v", a.Modules, b.Modules)
		}
	})

	t.Run("import-order", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		const entryModule = "main"
		ab := phase31ModulesWriteOrder(t, "import alpha {\n  fn f_alpha\n}\nimport beta {\n  fn f_beta\n}\n")
		ba := phase31ModulesWriteOrder(t, "import beta {\n  fn f_beta\n}\nimport alpha {\n  fn f_alpha\n}\n")
		_, ra := phase31ModulesBuild(t, filepath.Join(ab, "main.schway"))
		_, rb := phase31ModulesBuild(t, filepath.Join(ba, "main.schway"))
		libraries := func(r native.BuildReceipt) []native.ModuleDigest {
			var out []native.ModuleDigest
			for _, row := range r.Modules {
				if row.Module != entryModule {
					out = append(out, row)
				}
			}
			return out
		}
		if len(libraries(ra)) != 2 || !reflect.DeepEqual(libraries(ra), libraries(rb)) {
			t.Fatalf("library rows differ: %+v vs %+v", libraries(ra), libraries(rb))
		}
		if ra.EmittedCDigest != rb.EmittedCDigest {
			t.Fatal("EmittedCDigest differs between import orders")
		}
		// Reproducibility control: rebuild the SAME dir. If this host's clang
		// does not reproduce its own executable (observed on darwin: the
		// digest moves between two builds of identical input), the
		// cross-order ExecutableDigest comparison is meaningless and the
		// InputID (the product's own ExecutableDigest-free ID) is compared
		// instead. ExecutableDigest is never zeroed here.
		_, again := phase31ModulesBuild(t, filepath.Join(ab, "main.schway"))
		reproducible := again.ExecutableDigest == ra.ExecutableDigest
		if reproducible && ra.ExecutableDigest != rb.ExecutableDigest {
			t.Fatalf("ExecutableDigest differs between import orders on a reproducible toolchain: %s vs %s", ra.ExecutableDigest, rb.ExecutableDigest)
		}
		if !reproducible {
			t.Logf("toolchain does not reproduce executables on this host; comparing InputID instead of ID")
		}
		// D-21: the entry bytes differ between the two dirs, so the BuildID
		// itself is not compared; only the entry-derived fields are zeroed.
		normalize := func(r native.BuildReceipt) string {
			identity := *r.Identity
			identity.SourceDigest = ""
			identity.Modules = append([]native.ModuleDigest(nil), identity.Modules...)
			for i := range identity.Modules {
				if identity.Modules[i].Module == entryModule {
					identity.Modules[i].Digest = ""
				}
			}
			if !reproducible {
				return identity.InputID()
			}
			return identity.ID()
		}
		if normalize(ra) != normalize(rb) {
			t.Fatal("normalized identities differ between import orders")
		}
		set, diagnostics, err := modload.Load(filepath.Join(ab, "main.schway"))
		if err != nil || len(diagnostics) > 0 {
			t.Fatalf("load: %v %+v", err, diagnostics)
		}
		if len(set.Sources) != 3 {
			t.Fatalf("sources = %d, want 3", len(set.Sources))
		}
		reversed := set
		reversed.Sources = []modload.Source{set.Sources[0]}
		for i := len(set.Sources) - 1; i >= 1; i-- {
			reversed.Sources = append(reversed.Sources, set.Sources[i])
		}
		if !reflect.DeepEqual(session.ModuleDigestsForTest(set), session.ModuleDigestsForTest(reversed)) {
			t.Fatal("ModuleDigestsForTest depends on Sources order")
		}
	})

	t.Run("no-import", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		artifact, _ := phase31ModulesBuild(t, testsupport.ProjectPath("examples", "sum_to_n.schway"))
		data, err := os.ReadFile(artifact + ".schway-build.json")
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), `"modules"`) {
			t.Fatalf("no-import receipt carries a modules key: %s", data)
		}
	})

	t.Run("evidence", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		dir := testsupport.ProjectPath("examples", "checksum")
		artifact, receipt := phase31ModulesBuild(t, filepath.Join(dir, "checksum.schway"), filepath.Join(dir, "file_bytes.bindings.json"))
		wikipedia := filepath.Join(t.TempDir(), "wikipedia.txt")
		if err := os.WriteFile(wikipedia, []byte("Wikipedia"), 0o600); err != nil {
			t.Fatal(err)
		}
		var stdout, stderr bytes.Buffer
		_, report, err := native.RunApplicationWithEvidence(context.Background(), artifact, wikipedia, filepath.Join(t.TempDir(), "report.json"), native.EvidenceEvents, &stdout, &stderr)
		if err != nil {
			t.Fatalf("run: %v", err)
		}
		if len(report.Modules) != 2 || !reflect.DeepEqual(report.Modules, receipt.Modules) {
			t.Fatalf("report modules %+v != receipt modules %+v", report.Modules, receipt.Modules)
		}
	})

	t.Run("two-apps-same-library", func(t *testing.T) {
		phase31ModulesNeedClang(t)
		ab := phase31ModulesWriteOrder(t, "import alpha {\n  fn f_alpha\n}\nimport beta {\n  fn f_beta\n}\n")
		other := phase31ModulesWriteOrder(t, "import beta {\n  fn f_beta\n}\nimport alpha {\n  fn f_alpha\n}\n")
		_, ra := phase31ModulesBuild(t, filepath.Join(ab, "main.schway"))
		_, rb := phase31ModulesBuild(t, filepath.Join(other, "main.schway"))
		for _, name := range []string{"alpha", "beta"} {
			if phase31ModulesRow(ra, name) != phase31ModulesRow(rb, name) || phase31ModulesRow(ra, name).Digest == "" {
				t.Fatalf("%s rows differ or are missing", name)
			}
		}
	})
}

func phase31ModulesRow(r native.BuildReceipt, name string) native.ModuleDigest {
	for _, row := range r.Modules {
		if row.Module == name {
			return row
		}
	}
	return native.ModuleDigest{}
}

func phase31ModulesCopyChecksum(t *testing.T) string {
	t.Helper()
	src := testsupport.ProjectPath("examples", "checksum")
	dst := t.TempDir()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

func phase31ModulesWriteOrder(t *testing.T, imports string) string {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"alpha.schway": "module alpha\n\nexport {\n  fn f_alpha\n}\n\nfn f_alpha(n: U64) -> U64 {\n  var r = n\n  r = r + 1\n  r\n}\n",
		"beta.schway":  "module beta\n\nexport {\n  fn f_beta\n}\n\nfn f_beta(n: U64) -> U64 {\n  var r = n\n  r = r + 2\n  r\n}\n",
		"main.schway":  "module main\n\n" + imports + "\nexport {\n  fn main\n}\n\nfn main(n: U64) -> U64 {\n  let a = f_alpha(n)\n  let b = f_beta(a)\n  b\n}\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}
