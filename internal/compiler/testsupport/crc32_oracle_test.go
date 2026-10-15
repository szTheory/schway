package testsupport

import (
	"encoding/json"
	"os"
	"testing"
)

func TestPhase37CRC32OracleRows(t *testing.T) {
	data, err := os.ReadFile(ProjectPath("examples", "crc32", "crc32.expected.json"))
	if err != nil {
		t.Fatalf("read independent CRC-32 oracle fixture: %v", err)
	}
	var fixture struct {
		Schema     string `json:"schema"`
		Limit      int    `json:"limit"`
		Provenance struct {
			Python        string `json:"python"`
			ZlibCompile   string `json:"zlib_compile"`
			ZlibRuntime   string `json:"zlib_runtime"`
			Command       string `json:"command"`
		} `json:"provenance"`
		Recipes []struct {
			ID      string  `json:"id"`
			CRC32   *uint64 `json:"crc32"`
			Low16   *uint64 `json:"low16"`
			SHA256  string  `json:"sha256"`
		} `json:"recipes"`
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatalf("decode independent CRC-32 oracle fixture: %v", err)
	}
	if fixture.Schema != "schway.crc32-cases/1" {
		t.Fatalf("schema = %q, want schway.crc32-cases/1", fixture.Schema)
	}
	if fixture.Limit != 65536 {
		t.Fatalf("limit = %d, want 65536", fixture.Limit)
	}
	if fixture.Provenance.Python != "3.14.4" || fixture.Provenance.ZlibCompile != "1.2.12" || fixture.Provenance.ZlibRuntime != "1.2.12" {
		t.Fatalf("Python/zlib receipt = %+v", fixture.Provenance)
	}
	if fixture.Provenance.Command == "" {
		t.Fatal("independent Python/zlib generation command is absent")
	}
	want := map[string]uint64{
		"empty":           0,
		"zero":            3523407757,
		"check":           3421780262,
		"highbit-80":      1069182125,
		"highbit-80ff":    1465410873,
		"all-bytes-00-ff": 688229491,
		"cap-ff":          3735780942,
	}
	seen := make(map[string]bool, len(fixture.Recipes))
	for _, recipe := range fixture.Recipes {
		if seen[recipe.ID] {
			t.Errorf("duplicate recipe ID %q", recipe.ID)
		}
		seen[recipe.ID] = true
		value, required := want[recipe.ID]
		if !required {
			continue
		}
		if recipe.CRC32 == nil || *recipe.CRC32 != value {
			t.Errorf("%s CRC32 = %v, want %d", recipe.ID, recipe.CRC32, value)
		}
		if recipe.Low16 == nil || *recipe.Low16 != value&65535 {
			t.Errorf("%s low16 = %v, want %d", recipe.ID, recipe.Low16, value&65535)
		}
		if recipe.SHA256 == "" {
			t.Errorf("%s recipe has no SHA256 digest", recipe.ID)
		}
	}
	for id := range want {
		if !seen[id] {
			t.Errorf("missing pinned CRC32 recipe %q", id)
		}
	}
}
