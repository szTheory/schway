package evidence_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/corevalidate"
	"github.com/codename-lang/lang/internal/compiler/evidence"
	"github.com/codename-lang/lang/internal/compiler/execution"
	"github.com/codename-lang/lang/internal/compiler/testsupport"
)

func TestOwnedEvidenceBindings(t *testing.T) {
	product := ownedProduct(t)
	if golden := readPhase2(t, "evidence.golden.json"); !bytes.Equal(product.ManifestBytes, golden) {
		t.Fatalf("owned evidence golden differs:\n--- got ---\n%s\n--- want ---\n%s", product.ManifestBytes, golden)
	}
	manifest := product.Manifest
	if manifest.Schema != "lang.evidence/1" || manifest.CoreSchema != "lang.core/1" || manifest.ExecutionSchema != "lang.execution/1" || manifest.DiagnosticSchema != "lang.diagnostic/1" {
		t.Fatalf("owned evidence omitted concrete schemas: %+v", manifest)
	}
	if manifest.DigestClaim != "content-identity-only" || manifest.KnownEscape != corevalidate.KnownEscape {
		t.Fatalf("owned evidence overclaimed authority: %+v", manifest)
	}
	if len(manifest.ExecutionDigests) != 1 || len(product.Executions) != 1 {
		t.Fatalf("owned evidence omitted ordered execution binding: manifest=%+v executions=%+v", manifest, product.Executions)
	}
	encoded, err := execution.CanonicalBytes(product.Executions[0])
	if err != nil || manifest.ExecutionDigests[0] != evidence.ContentDigest(encoded) {
		t.Fatalf("execution digest mismatch: err=%v manifest=%+v execution=%s", err, manifest, encoded)
	}
	if err := evidence.Validate(manifest, readPhase2(t, "owned_transfer.lang"), ownedFacts()); err != nil {
		t.Fatalf("owned evidence did not validate: %v", err)
	}

	// Trust-boundary products must own ordered slices instead of aliasing caller
	// facts or the returned execution records.
	before := append([]byte(nil), product.ManifestBytes...)
	product.Manifest.Flags[0] = "-mutated"
	product.Executions[0].Events[0].ID = "event:mutated"
	if bytes.Equal(mustCanonical(t, product.Manifest), before) {
		t.Fatal("manifest mutation did not alter canonical bytes")
	}
	fresh := ownedProduct(t)
	if !bytes.Equal(fresh.ManifestBytes, before) {
		t.Fatal("returned owned slices aliased later evidence builds")
	}
}

func TestOwnedEvidenceMutationMatrix(t *testing.T) {
	product := ownedProduct(t)
	source := readPhase2(t, "owned_transfer.lang")
	tests := []struct {
		name string
		code string
		edit func(*evidence.Manifest)
	}{
		{"diagnostic schema", "evidence.diagnostic_schema_mismatch", func(v *evidence.Manifest) { v.DiagnosticSchema = "lang.diagnostic/0" }},
		{"execution digest", "evidence.execution_mismatch", func(v *evidence.Manifest) { v.ExecutionDigests[0] = staleDigest() }},
		{"escape", "evidence.escape_mismatch", func(v *evidence.Manifest) { v.KnownEscape = "detected:coordinated-source-core-lie" }},
		{"digest claim", "evidence.digest_claim_mismatch", func(v *evidence.Manifest) { v.DigestClaim = "translation-proof" }},
		{"core", "evidence.core_mismatch", func(v *evidence.Manifest) { v.CoreDigest = staleDigest() }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := cloneManifest(t, product.Manifest)
			test.edit(&mutated)
			if err := evidence.Validate(mutated, source, ownedFacts()); evidence.ErrorCode(err) != test.code {
				t.Fatalf("mutation code=%s want=%s", evidence.ErrorCode(err), test.code)
			}
		})
	}

	reordered := cloneManifest(t, product.Manifest)
	canonical := cloneExecution(t, product.Executions[0])
	canonical.Events[0], canonical.Events[1] = canonical.Events[1], canonical.Events[0]
	encoded, _ := execution.CanonicalBytes(canonical)
	reordered.ExecutionDigests[0] = evidence.ContentDigest(encoded)
	if err := evidence.Validate(reordered, source, ownedFacts()); evidence.ErrorCode(err) != "evidence.execution_mismatch" {
		t.Fatalf("event reorder code=%s", evidence.ErrorCode(err))
	}

	unknown := bytes.Replace(product.ManifestBytes, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1)
	if _, err := evidence.DecodeStrict(unknown); evidence.ErrorCode(err) != "evidence.invalid_json" {
		t.Fatalf("unknown field code=%s", evidence.ErrorCode(err))
	}
	trailing := append(append([]byte(nil), product.ManifestBytes...), []byte("{}")...)
	if _, err := evidence.DecodeStrict(trailing); evidence.ErrorCode(err) != "evidence.trailing_json" {
		t.Fatalf("trailing code=%s", evidence.ErrorCode(err))
	}
}

func TestPhase1EvidenceGoldenUnchanged(t *testing.T) {
	product := goldenProduct(t)
	if got, want := product.ManifestBytes, readGolden(t, "evidence.golden.json"); !bytes.Equal(got, want) {
		t.Fatalf("Phase 1 evidence golden changed:\ngot  %s\nwant %s", got, want)
	}
}

func ownedProduct(t testing.TB) evidence.Product {
	t.Helper()
	product, diagnostics, err := evidence.Build(readPhase2(t, "owned_transfer.lang"), ownedFacts())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build owned evidence: err=%v diagnostics=%+v", err, diagnostics)
	}
	return product
}

func ownedFacts() evidence.Facts {
	return evidence.Facts{CompilerIdentity: "codename-lang-stage0/go1.24-fixture", ClangIdentity: "clang-fixture 21.0.0", Target: "arm64-apple-darwin-fixture", Flags: append([]string(nil), evidence.DefaultFlags...), Policy: "phase2-owned-c17-v1"}
}

func readPhase2(t testing.TB, name string) []byte {
	t.Helper()
	value, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase2", name))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func mustCanonical(t testing.TB, manifest evidence.Manifest) []byte {
	t.Helper()
	value, err := evidence.CanonicalBytes(manifest)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func cloneManifest(t testing.TB, value evidence.Manifest) evidence.Manifest {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var cloned evidence.Manifest
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func cloneExecution(t testing.TB, value execution.Execution) execution.Execution {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var cloned execution.Execution
	if err := json.Unmarshal(encoded, &cloned); err != nil {
		t.Fatal(err)
	}
	return cloned
}

func TestCanonicalEvidence(t *testing.T) {
	product := goldenProduct(t)
	second := goldenProduct(t)
	if !bytes.Equal(product.ManifestBytes, second.ManifestBytes) {
		t.Fatalf("manifest bytes changed across identical builds:\n%s\n%s", product.ManifestBytes, second.ManifestBytes)
	}
	manifestGolden := readGolden(t, "evidence.golden.json")
	if !bytes.Equal(product.ManifestBytes, manifestGolden) {
		t.Fatalf("evidence golden differs:\n--- got ---\n%s\n--- want ---\n%s", product.ManifestBytes, manifestGolden)
	}
	cGolden := readGolden(t, "generated.golden.c")
	if !bytes.Equal(product.CSource, cGolden) {
		t.Fatalf("generated C golden differs:\n--- got ---\n%s\n--- want ---\n%s", product.CSource, cGolden)
	}
	decoded, err := evidence.DecodeStrict(product.ManifestBytes)
	if err != nil || !reflect.DeepEqual(decoded, product.Manifest) {
		t.Fatalf("strict round trip failed: err=%v decoded=%+v", err, decoded)
	}
}

func TestStaleManifest(t *testing.T) {
	product := goldenProduct(t)
	source := readGolden(t, "toggle.lang")
	staleSource := append([]byte("// changed review context\n"), source...)
	if err := evidence.Validate(product.Manifest, staleSource, goldenFacts()); evidence.ErrorCode(err) != "evidence.source_mismatch" {
		t.Fatalf("stale source code=%s want=evidence.source_mismatch", evidence.ErrorCode(err))
	}
}

func TestEvidenceMutationMatrix(t *testing.T) {
	product := goldenProduct(t)
	source := readGolden(t, "toggle.lang")
	tests := []struct {
		name string
		code string
		edit func(*evidence.Manifest)
	}{
		{"schema", "evidence.schema_mismatch", func(value *evidence.Manifest) { value.Schema = "lang.evidence/9" }},
		{"id algorithm", "evidence.id_algorithm_mismatch", func(value *evidence.Manifest) { value.IDAlgorithm = "other" }},
		{"source schema", "evidence.source_schema_mismatch", func(value *evidence.Manifest) { value.SourceSchema = "other" }},
		{"core schema", "evidence.core_schema_mismatch", func(value *evidence.Manifest) { value.CoreSchema = "other" }},
		{"execution schema", "evidence.execution_schema_mismatch", func(value *evidence.Manifest) { value.ExecutionSchema = "other" }},
		{"compiler", "evidence.compiler_mismatch", func(value *evidence.Manifest) { value.CompilerIdentity = "other" }},
		{"clang", "evidence.clang_mismatch", func(value *evidence.Manifest) { value.ClangIdentity = "other" }},
		{"target", "evidence.target_mismatch", func(value *evidence.Manifest) { value.Target = "other" }},
		{"flags changed", "evidence.flags_mismatch", func(value *evidence.Manifest) { value.Flags[0] = "-std=c99" }},
		{"flags reordered", "evidence.flags_mismatch", func(value *evidence.Manifest) { value.Flags[0], value.Flags[1] = value.Flags[1], value.Flags[0] }},
		{"policy", "evidence.policy_mismatch", func(value *evidence.Manifest) { value.Policy = "other" }},
		{"source digest", "evidence.source_mismatch", func(value *evidence.Manifest) { value.SourceDigest = staleDigest() }},
		{"core digest", "evidence.core_mismatch", func(value *evidence.Manifest) { value.CoreDigest = staleDigest() }},
		{"C digest", "evidence.c_mismatch", func(value *evidence.Manifest) { value.CDigest = staleDigest() }},
		{"manifest ID", "evidence.id_mismatch", func(value *evidence.Manifest) { value.ID = "evidence:stale" }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mutated := product.Manifest
			mutated.Flags = append([]string(nil), product.Manifest.Flags...)
			test.edit(&mutated)
			if err := evidence.Validate(mutated, source, goldenFacts()); evidence.ErrorCode(err) != test.code {
				t.Fatalf("mutation code=%s want=%s", evidence.ErrorCode(err), test.code)
			}
		})
	}

	unknown := bytes.Replace(product.ManifestBytes, []byte(`"schema":`), []byte(`"unknown":true,"schema":`), 1)
	if _, err := evidence.DecodeStrict(unknown); evidence.ErrorCode(err) != "evidence.invalid_json" {
		t.Fatalf("unknown field code=%s", evidence.ErrorCode(err))
	}
	trailing := append(append([]byte(nil), product.ManifestBytes...), []byte("{}")...)
	if _, err := evidence.DecodeStrict(trailing); evidence.ErrorCode(err) != "evidence.trailing_json" {
		t.Fatalf("trailing value code=%s", evidence.ErrorCode(err))
	}
}

func TestEvidenceRelocation(t *testing.T) {
	source := readGolden(t, "toggle.lang")
	var products []evidence.Product
	for index := 0; index < 2; index++ {
		path := filepath.Join(t.TempDir(), "relocated.lang")
		if err := os.WriteFile(path, source, 0o600); err != nil {
			t.Fatal(err)
		}
		relocated, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		product, diagnostics, err := evidence.Build(relocated, goldenFacts())
		if err != nil || len(diagnostics) != 0 {
			t.Fatalf("build relocated: err=%v diagnostics=%+v", err, diagnostics)
		}
		products = append(products, product)
	}
	if !bytes.Equal(products[0].ManifestBytes, products[1].ManifestBytes) || !bytes.Equal(products[0].CSource, products[1].CSource) {
		t.Fatal("temporary input location changed evidence bytes")
	}
}

func TestEvidenceCLI(t *testing.T) {
	binary := testsupport.BuildCLI(t)
	source := testsupport.ProjectPath("testdata", "phase1", "toggle.lang")
	emitted := testsupport.RunCLI(t, binary, nil, "evidence", source)
	if emitted.Exit != 0 || len(emitted.Stderr) != 0 {
		t.Fatalf("emit evidence: %+v", emitted)
	}
	manifestPath := filepath.Join(t.TempDir(), "evidence.json")
	if err := os.WriteFile(manifestPath, emitted.Stdout, 0o600); err != nil {
		t.Fatal(err)
	}
	validated := testsupport.RunCLI(t, binary, nil, "--json", "evidence", "--validate", manifestPath, source)
	if validated.Exit != 0 || len(validated.Stderr) != 0 || !bytes.Contains(validated.Stdout, []byte(`"status":"pass"`)) {
		t.Fatalf("validate evidence: %+v", validated)
	}

	staleSource := filepath.Join(t.TempDir(), "stale.lang")
	if err := os.WriteFile(staleSource, append([]byte("// stale\n"), readGolden(t, "toggle.lang")...), 0o600); err != nil {
		t.Fatal(err)
	}
	rejected := testsupport.RunCLI(t, binary, nil, "--json", "evidence", "--validate", manifestPath, staleSource)
	if rejected.Exit != 2 || !bytes.Contains(rejected.Stdout, []byte(`"code":"evidence.source_mismatch"`)) {
		t.Fatalf("stale evidence was not rejected: %+v", rejected)
	}
}

func FuzzEvidenceDecode(f *testing.F) {
	product := goldenProduct(f)
	f.Add(product.ManifestBytes)
	f.Add([]byte(`{"schema":"lang.evidence/0"}`))
	f.Add([]byte("not json"))
	f.Fuzz(func(t *testing.T, data []byte) {
		manifest, err := evidence.DecodeStrict(data)
		if err != nil {
			return
		}
		encoded, err := evidence.CanonicalBytes(manifest)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := evidence.DecodeStrict(encoded); err != nil {
			t.Fatalf("accepted manifest did not re-encode strictly: %v", err)
		}
	})
}

func goldenProduct(t testing.TB) evidence.Product {
	t.Helper()
	product, diagnostics, err := evidence.Build(readGolden(t, "toggle.lang"), goldenFacts())
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build golden: err=%v diagnostics=%+v", err, diagnostics)
	}
	return product
}

func goldenFacts() evidence.Facts {
	return evidence.Facts{
		CompilerIdentity: "codename-lang-stage0/go1.24-fixture",
		ClangIdentity:    "clang-fixture 21.0.0",
		Target:           "arm64-apple-darwin-fixture",
		Flags:            append([]string(nil), evidence.DefaultFlags...),
		Policy:           "phase1-pure-c17-v1",
	}
}

func readGolden(t testing.TB, name string) []byte {
	t.Helper()
	value, err := os.ReadFile(testsupport.ProjectPath("testdata", "phase1", name))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func staleDigest() string {
	return "sha256:0000000000000000000000000000000000000000000000000000000000000000"
}
