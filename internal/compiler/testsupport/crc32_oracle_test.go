package testsupport

import (
	"crypto/sha256"
	"fmt"
	"hash/adler32"
	"hash/crc32"
	"strings"
	"testing"
)

const phase37CRC32GeneratorSHA256 = "61502ba4db0b2ee2b2208ef6dcc167cd7858e60a3fd9ad4142d8d2192606290f"

var phase37CRC32Pins = map[string]uint64{
	"empty":            0,
	"zero":             3523407757,
	"check":            3421780262,
	"highbit-80":       1069182125,
	"highbit-80ff":     1465410873,
	"all-bytes-00-ff":  688229491,
	"cap-minus-one-ff": 529201972,
	"cap-ff":           3735780942,
	"wikipedia":        2913648686,
}

var phase37CRC32RecipeDigests = map[string]string{
	"empty":            "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
	"zero":             "6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
	"check":            "15e2b0d3c33891ebb0f1ef609ec419420c20e320ce94c65fbc8c3312448eb225",
	"highbit-80":       "76be8b528d0075f7aae98d6fa57a6d3c83ae480a8469e668d7b0af968995ac71",
	"highbit-80ff":     "d87d01642f47a0d1901b1dfd2331c9def1bcfc69d8835c6cd911fe4165b804e4",
	"all-bytes-00-ff":  "40aff2e9d2d8922e47afd4648e6967497158785fbd1da870e7110266bf944880",
	"cap-minus-one-ff": "c810c1bb9815d632215ce08fda5afad7564a9c52c275422c7c92434232e6126b",
	"cap-ff":           "71189f7fb6aed638640078fba3a35fda6c39c8962e74dcc75935aac948da9063",
	"cap-plus-one-ff":  "ef9993e97e36b66b65b16a741da437aecddffa22185e176a49f8cfe9ecc7a07c",
	"wikipedia":        "d38b38a2dd476e045c299e8ee5d6466834456d97bd592a71746b423a6a05f386",
}

func TestPhase37CRC32OracleRows(t *testing.T) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := CRC32OracleRows()
	if err != nil {
		t.Fatal(err)
	}
	if problems := phase37CRC32FixtureProblems(fixture, rows); len(problems) != 0 {
		t.Fatalf("CRC32 oracle fixture problems: %v", problems)
	}
	if len(rows) != 10 {
		t.Fatalf("oracle recipes = %d, want 10", len(rows))
	}
	if len(fixture.Consumers) != 4 {
		t.Fatalf("consumer rows = %d, want 4", len(fixture.Consumers))
	}
	if len(fixture.Cases) != 27 {
		t.Fatalf("public matrix case rows = %d, want 27", len(fixture.Cases))
	}
	if len(fixture.ManifestControls) != 2 {
		t.Fatalf("manifest controls = %d, want missing and malformed", len(fixture.ManifestControls))
	}

	cases, err := CRC32CaseMatrix()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range cases {
		if len(row.Expected) != 4 {
			t.Errorf("case %s has %d consumer tuples, want 4", row.ID, len(row.Expected))
		}
	}
	if got := cases[2].Expected["crc_low"].Stdout; got != "14630\n" {
		t.Fatalf("check low16 tuple = %q, want 14630 newline", got)
	}
	capPlusOne := cases[8]
	for _, consumer := range []string{"crc32", "crc_low", "checksum", "adler_low"} {
		if got := capPlusOne.Expected[consumer]; got.Exit != 65 || got.Stdout != "" || got.Stderr != "FileError.FileTooLarge\n" {
			t.Errorf("cap-plus-one/%s = %+v", consumer, got)
		}
	}
}

func TestPhase37CRC32OracleDriftControls(t *testing.T) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := CRC32OracleRows()
	if err != nil {
		t.Fatal(err)
	}
	if problems := phase37CRC32FixtureProblems(fixture, rows); len(problems) != 0 {
		t.Fatalf("unmutated oracle rejected: %v", problems)
	}

	t.Run("literal-answer", func(t *testing.T) {
		mutated := append([]CRC32OracleRow(nil), rows...)
		index := phase37CRC32RecipeIndex(mutated, "check")
		if index < 0 || mutated[index].CRC32 == nil {
			t.Fatal("check recipe anchor is missing")
		}
		value := *mutated[index].CRC32 + 1
		mutated[index].CRC32 = &value
		if problems := phase37CRC32FixtureProblems(fixture, mutated); len(problems) == 0 {
			t.Fatal("changed check CRC literal escaped the independent control")
		}
	})

	t.Run("recipe-digest", func(t *testing.T) {
		mutated := append([]CRC32OracleRow(nil), rows...)
		index := phase37CRC32RecipeIndex(mutated, "cap-ff")
		if index < 0 {
			t.Fatal("cap recipe anchor is missing")
		}
		mutated[index].SHA256 = strings.Repeat("0", 64)
		if problems := phase37CRC32FixtureProblems(fixture, mutated); len(problems) == 0 {
			t.Fatal("changed cap recipe digest escaped the recomputation control")
		}
	})

	t.Run("generator-receipt", func(t *testing.T) {
		mutated := fixture
		mutated.Provenance.Command += " # altered"
		if problems := phase37CRC32FixtureProblems(mutated, rows); len(problems) == 0 {
			t.Fatal("changed Python generator command escaped its receipt pin")
		}
	})

	t.Run("failure-tuple", func(t *testing.T) {
		mutated := fixture
		mutated.Outcomes = make(map[string]CRC32Tuple, len(fixture.Outcomes))
		for name, outcome := range fixture.Outcomes {
			mutated.Outcomes[name] = outcome
		}
		bad := mutated.Outcomes["FileUnavailable"]
		bad.Stderr = "FileError.FileTooLarge\n"
		mutated.Outcomes["FileUnavailable"] = bad
		if problems := phase37CRC32FixtureProblems(mutated, rows); len(problems) == 0 {
			t.Fatal("changed public file-error tuple escaped its fixed control")
		}
	})
}

func phase37CRC32FixtureProblems(fixture crc32Fixture, rows []CRC32OracleRow) []string {
	var problems []string
	if fixture.Schema != "schway.crc32-cases/1" {
		problems = append(problems, fmt.Sprintf("schema %q", fixture.Schema))
	}
	if fixture.Limit != CRC32Limit {
		problems = append(problems, fmt.Sprintf("limit %d, want %d", fixture.Limit, CRC32Limit))
	}
	receipt := fixture.Provenance
	if receipt.Generator != "Python zlib" || receipt.Python != "3.14.4" || receipt.ZlibCompile != "1.2.12" || receipt.ZlibRuntime != "1.2.12" || receipt.GeneratedOn != "2026-10-15" {
		problems = append(problems, fmt.Sprintf("unexpected Python/zlib receipt %+v", receipt))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(receipt.Command))); got != phase37CRC32GeneratorSHA256 {
		problems = append(problems, fmt.Sprintf("generator command SHA256 %s", got))
	}
	problems = append(problems, phase37CRC32RowProblems(rows)...)
	problems = append(problems, phase37CRC32ConsumerProblems(fixture.Consumers)...)
	problems = append(problems, phase37CRC32OutcomeProblems(fixture.Outcomes)...)
	problems = append(problems, phase37CRC32CaseProblems(fixture.Cases)...)
	if len(fixture.ManifestControls) != 2 || fixture.ManifestControls[0].ID != "missing-manifest" || fixture.ManifestControls[0].Kind != "missing" || fixture.ManifestControls[1].ID != "bad-manifest" || fixture.ManifestControls[1].Kind != "malformed" {
		problems = append(problems, "missing or reordered missing/bad manifest controls")
	}
	return problems
}

func phase37CRC32RowProblems(rows []CRC32OracleRow) []string {
	var problems []string
	seen := make(map[string]bool, len(rows))
	for _, row := range rows {
		if seen[row.ID] {
			problems = append(problems, "duplicate recipe "+row.ID)
		}
		seen[row.ID] = true
		if row.Expression == "" {
			problems = append(problems, row.ID+": empty byte expression")
		}
		if row.Length != len(row.Input) {
			problems = append(problems, fmt.Sprintf("%s: recorded length %d, materialized %d", row.ID, row.Length, len(row.Input)))
		}
		digest := fmt.Sprintf("%x", sha256.Sum256(row.Input))
		if row.SHA256 != digest || phase37CRC32RecipeDigests[row.ID] != row.SHA256 {
			problems = append(problems, fmt.Sprintf("%s: SHA256 %s, recomputed %s", row.ID, row.SHA256, digest))
		}
		if row.ID == "cap-plus-one-ff" {
			if row.CRC32 != nil || row.Low16 != nil || row.Adler32 != nil || row.AdlerLow16 != nil {
				problems = append(problems, "oversize recipe has a numeric application answer")
			}
			if len(row.Input) != CRC32Limit+1 {
				problems = append(problems, fmt.Sprintf("oversize recipe length %d, want %d", len(row.Input), CRC32Limit+1))
			}
			continue
		}
		crcValue, lowValue, adlerValue, adlerLowValue := row.CRC32, row.Low16, row.Adler32, row.AdlerLow16
		if crcValue == nil || lowValue == nil || adlerValue == nil || adlerLowValue == nil {
			problems = append(problems, row.ID+": missing numeric answer")
			continue
		}
		gotCRC := uint64(crc32.ChecksumIEEE(row.Input))
		if *crcValue != gotCRC || *crcValue != phase37CRC32Pins[row.ID] {
			problems = append(problems, fmt.Sprintf("%s: CRC32 literal %d, IEEE %d", row.ID, *crcValue, gotCRC))
		}
		if *lowValue != *crcValue&65535 {
			problems = append(problems, fmt.Sprintf("%s: low16 %d, projection %d", row.ID, *lowValue, *crcValue&65535))
		}
		gotAdler := uint64(adler32.Checksum(row.Input))
		if *adlerValue != gotAdler || *adlerLowValue != gotAdler&65535 {
			problems = append(problems, fmt.Sprintf("%s: Adler literals %d/%d, Go %d/%d", row.ID, *adlerValue, *adlerLowValue, gotAdler, gotAdler&65535))
		}
	}
	for id := range phase37CRC32RecipeDigests {
		if !seen[id] {
			problems = append(problems, "missing recipe "+id)
		}
	}
	problems = append(problems, phase37CRC32LegacyAdlerProblems(rows)...)
	return problems
}

func phase37CRC32LegacyAdlerProblems(rows []CRC32OracleRow) []string {
	oldRows := ChecksumOracleRows()
	if errs := checksumOracleRowProblems(oldRows); len(errs) != 0 {
		return append([]string{"existing checksum oracle owner invalid"}, errs...)
	}
	byID := make(map[string]CRC32OracleRow, len(rows))
	for _, row := range rows {
		byID[row.ID] = row
	}
	oldToNew := map[string]string{
		"empty":      "empty",
		"0x00":       "zero",
		"Wikipedia":  "wikipedia",
		"65536x0xFF": "cap-ff",
		"0x80":       "highbit-80",
		"0x80 0xFF":  "highbit-80ff",
	}
	var problems []string
	for _, old := range oldRows {
		if old.Outcome != "returned" {
			continue
		}
		newID, ok := oldToNew[old.Name]
		if !ok {
			continue
		}
		row := byID[newID]
		want := row.Adler32
		if old.App == "adler_low" {
			want = row.AdlerLow16
		}
		if want == nil || *want != old.Want {
			problems = append(problems, fmt.Sprintf("legacy %s/%s literal %d disagrees with new row %s", old.App, old.Name, old.Want, newID))
		}
	}
	return problems
}

func phase37CRC32ConsumerProblems(consumers []CRC32Consumer) []string {
	want := map[string]CRC32Consumer{
		"crc32":     {ID: "crc32", Source: "examples/crc32/crc32.schway", Manifest: "examples/crc32/file_bytes.bindings.json", Answer: "crc32"},
		"crc_low":   {ID: "crc_low", Source: "examples/crc32/crc_low.schway", Manifest: "examples/crc32/file_bytes.bindings.json", Answer: "low16"},
		"checksum":  {ID: "checksum", Source: "examples/checksum/checksum.schway", Manifest: "examples/checksum/file_bytes.bindings.json", Answer: "adler32"},
		"adler_low": {ID: "adler_low", Source: "examples/checksum/adler_low.schway", Manifest: "examples/checksum/file_bytes.bindings.json", Answer: "adler_low16"},
	}
	var problems []string
	if len(consumers) != len(want) {
		problems = append(problems, fmt.Sprintf("consumer count %d, want %d", len(consumers), len(want)))
	}
	seen := make(map[string]bool, len(consumers))
	for _, consumer := range consumers {
		if seen[consumer.ID] {
			problems = append(problems, "duplicate consumer "+consumer.ID)
		}
		seen[consumer.ID] = true
		if consumer != want[consumer.ID] {
			problems = append(problems, fmt.Sprintf("consumer %s tuple %+v, want %+v", consumer.ID, consumer, want[consumer.ID]))
		}
	}
	return problems
}

func phase37CRC32OutcomeProblems(outcomes map[string]CRC32Tuple) []string {
	want := map[string]CRC32Tuple{
		"FileTooLarge":     {Exit: 65, Stderr: "FileError.FileTooLarge\n"},
		"FileUnavailable":  {Exit: 65, Stderr: "FileError.FileUnavailable\n"},
		"EmptyPath":        {Exit: 65},
		"InputTooLong":     {Exit: 65, Stderr: "schway run: input exceeds 4096 bytes\n"},
		"InputContainsNUL": {Exit: 65, Stderr: "schway run: input contains NUL\n"},
		"ManifestFailure":  {Exit: 3, Stderr: "schway run: native build failed\n"},
	}
	var problems []string
	if len(outcomes) != len(want) {
		problems = append(problems, fmt.Sprintf("outcome template count %d, want %d", len(outcomes), len(want)))
	}
	for name, expected := range want {
		if got, ok := outcomes[name]; !ok || got != expected {
			problems = append(problems, fmt.Sprintf("outcome %s = %+v, want %+v", name, got, expected))
		}
	}
	return problems
}

func phase37CRC32CaseProblems(cases []crc32CaseJSON) []string {
	want := map[string]crc32CaseJSON{
		"empty":               {ID: "empty", Kind: "file", Path: "empty.bin", Recipe: "empty"},
		"zero":                {ID: "zero", Kind: "file", Path: "zero.bin", Recipe: "zero"},
		"check":               {ID: "check", Kind: "file", Path: "check.bin", Recipe: "check"},
		"highbit-80":          {ID: "highbit-80", Kind: "file", Path: "highbit-80.bin", Recipe: "highbit-80"},
		"highbit-80ff":        {ID: "highbit-80ff", Kind: "file", Path: "highbit-80ff.bin", Recipe: "highbit-80ff"},
		"all-bytes-00-ff":     {ID: "all-bytes-00-ff", Kind: "file", Path: "all-bytes.bin", Recipe: "all-bytes-00-ff"},
		"cap-minus-one":       {ID: "cap-minus-one", Kind: "file", Path: "cap-minus-one.bin", Recipe: "cap-minus-one-ff"},
		"cap":                 {ID: "cap", Kind: "file", Path: "cap.bin", Recipe: "cap-ff"},
		"cap-plus-one":        {ID: "cap-plus-one", Kind: "file", Path: "cap-plus-one.bin", Recipe: "cap-plus-one-ff", Outcome: "FileTooLarge"},
		"wikipedia":           {ID: "wikipedia", Kind: "file", Path: "wikipedia.bin", Recipe: "wikipedia"},
		"missing":             {ID: "missing", Kind: "missing", Path: "missing.bin", Outcome: "FileUnavailable"},
		"directory":           {ID: "directory", Kind: "directory", Path: "directory", Outcome: "FileUnavailable"},
		"fifo":                {ID: "fifo", Kind: "fifo", Path: "no-writer.fifo", Outcome: "FileUnavailable"},
		"device":              {ID: "device", Kind: "device", Path: "/dev/null", Outcome: "FileUnavailable"},
		"unreadable":          {ID: "unreadable", Kind: "mode000", Path: "mode-000.bin", Recipe: "check", Outcome: "FileUnavailable"},
		"input-symlink":       {ID: "input-symlink", Kind: "symlink", Path: "input-link.bin", Target: "target.bin", Recipe: "check"},
		"opaque-json":         {ID: "opaque-json", Kind: "file", Path: "--json", Recipe: "check"},
		"opaque-depth":        {ID: "opaque-depth", Kind: "file", Path: "--depth=2", Recipe: "check"},
		"opaque-separator":    {ID: "opaque-separator", Kind: "file", Path: "--", Recipe: "check"},
		"opaque-leading-dash": {ID: "opaque-leading-dash", Kind: "file", Path: "-leading.bin", Recipe: "check"},
		"opaque-space":        {ID: "opaque-space", Kind: "file", Path: "space input.bin", Recipe: "check"},
		"opaque-unicode":      {ID: "opaque-unicode", Kind: "file", Path: "雪-λ.bin", Recipe: "check"},
		"opaque-shell":        {ID: "opaque-shell", Kind: "file", Path: "x ;$(literal).bin", Recipe: "check"},
		"empty-token":         {ID: "empty-token", Kind: "argument", Argument: "", Outcome: "EmptyPath"},
		"path-4096":           {ID: "path-4096", Kind: "argument", Repeat: "x", RepeatCount: 4096, Outcome: "FileUnavailable"},
		"path-4097":           {ID: "path-4097", Kind: "argument", Repeat: "x", RepeatCount: 4097, Outcome: "InputTooLong"},
		"nul-direct-dispatch": {ID: "nul-direct-dispatch", Kind: "direct-argument", Argument: "nul\x00path", Outcome: "InputContainsNUL", Execution: "direct-public-dispatch"},
	}
	var problems []string
	if len(cases) != len(want) {
		problems = append(problems, fmt.Sprintf("case row count %d, want %d", len(cases), len(want)))
	}
	seen := make(map[string]bool, len(cases))
	for _, row := range cases {
		if seen[row.ID] {
			problems = append(problems, "duplicate case "+row.ID)
		}
		seen[row.ID] = true
		if row != want[row.ID] {
			problems = append(problems, fmt.Sprintf("case %s = %+v, want %+v", row.ID, row, want[row.ID]))
		}
	}
	return problems
}

func phase37CRC32RecipeIndex(rows []CRC32OracleRow, id string) int {
	for index := range rows {
		if rows[index].ID == id {
			return index
		}
	}
	return -1
}
