package testsupport

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// CRC32Limit is the sample adapter's cap. FileBytes is otherwise
// binding-dependent, so this is test fixture data rather than a language-wide
// compiler limit.
const CRC32Limit = 65536

type CRC32OracleRow struct {
	ID         string
	Expression string
	Input      []byte
	Length     int
	SHA256     string
	CRC32      *uint64
	Low16      *uint64
	Adler32    *uint64
	AdlerLow16 *uint64
}

type CRC32Receipt struct {
	Generator   string `json:"generator"`
	Python      string `json:"python"`
	ZlibCompile string `json:"zlib_compile"`
	ZlibRuntime string `json:"zlib_runtime"`
	GeneratedOn string `json:"generated_on"`
	Command     string `json:"command"`
}

type CRC32Consumer struct {
	ID       string `json:"id"`
	Source   string `json:"source"`
	Manifest string `json:"manifest"`
	Answer   string `json:"answer"`
}

type CRC32Tuple struct {
	Exit    int    `json:"exit"`
	Stdout  string `json:"stdout"`
	Stderr  string `json:"stderr"`
	Outcome string `json:"outcome"`
}

type CRC32Case struct {
	ID          string
	Kind        string
	Path        string
	Target      string
	Recipe      string
	Outcome     string
	Execution   string
	Argument    string
	Repeat      string
	RepeatCount int
	Expected    map[string]CRC32Tuple
}

type CRC32ManifestControl struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

type crc32Fixture struct {
	Schema           string                 `json:"schema"`
	Limit            int                    `json:"limit"`
	Provenance       CRC32Receipt           `json:"provenance"`
	Consumers        []CRC32Consumer        `json:"consumers"`
	Recipes          []crc32RecipeJSON      `json:"recipes"`
	Outcomes         map[string]CRC32Tuple  `json:"outcomes"`
	Cases            []crc32CaseJSON        `json:"cases"`
	ManifestControls []CRC32ManifestControl `json:"manifest_controls"`
}

type crc32RecipeJSON struct {
	ID         string  `json:"id"`
	Expression string  `json:"expression"`
	Length     int     `json:"length"`
	SHA256     string  `json:"sha256"`
	CRC32      *uint64 `json:"crc32"`
	Low16      *uint64 `json:"low16"`
	Adler32    *uint64 `json:"adler32"`
	AdlerLow16 *uint64 `json:"adler_low16"`
}

type crc32CaseJSON struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	Target      string `json:"target"`
	Recipe      string `json:"recipe"`
	Outcome     string `json:"outcome"`
	Execution   string `json:"execution"`
	Argument    string `json:"argument"`
	Repeat      string `json:"repeat"`
	RepeatCount int    `json:"repeat_count"`
}

func loadCRC32Fixture() (crc32Fixture, error) {
	data, err := os.ReadFile(ProjectPath("examples", "crc32", "crc32.expected.json"))
	if err != nil {
		return crc32Fixture{}, err
	}
	var fixture crc32Fixture
	if err := json.Unmarshal(data, &fixture); err != nil {
		return crc32Fixture{}, fmt.Errorf("decode CRC-32 fixture: %w", err)
	}
	return fixture, nil
}

// CRC32OracleReceipt returns the exact Python/zlib generator receipt kept with
// the inert oracle rows.
func CRC32OracleReceipt() (CRC32Receipt, error) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		return CRC32Receipt{}, err
	}
	return fixture.Provenance, nil
}

// CRC32OracleRows materializes only the finite byte recipes recorded in the
// JSON fixture. Numeric answers remain the fixture's independent literals.
func CRC32OracleRows() ([]CRC32OracleRow, error) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		return nil, err
	}
	rows := make([]CRC32OracleRow, 0, len(fixture.Recipes))
	for _, recipe := range fixture.Recipes {
		input, err := crc32RecipeBytes(recipe.ID)
		if err != nil {
			return nil, err
		}
		rows = append(rows, CRC32OracleRow{
			ID:         recipe.ID,
			Expression: recipe.Expression,
			Input:      input,
			Length:     recipe.Length,
			SHA256:     recipe.SHA256,
			CRC32:      recipe.CRC32,
			Low16:      recipe.Low16,
			Adler32:    recipe.Adler32,
			AdlerLow16: recipe.AdlerLow16,
		})
	}
	return rows, nil
}

// CRC32Consumers returns the fixed public entries and their explicit native
// manifests. Interpreter invocations intentionally omit the manifest.
func CRC32Consumers() ([]CRC32Consumer, error) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		return nil, err
	}
	return append([]CRC32Consumer(nil), fixture.Consumers...), nil
}

// CRC32CaseMatrix expands each finite case into an exact expected tuple for
// every consumer. The caller runs each tuple independently on both engines.
func CRC32CaseMatrix() ([]CRC32Case, error) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		return nil, err
	}
	recipes := make(map[string]crc32RecipeJSON, len(fixture.Recipes))
	for _, recipe := range fixture.Recipes {
		recipes[recipe.ID] = recipe
	}
	rows := make([]CRC32Case, 0, len(fixture.Cases))
	for _, item := range fixture.Cases {
		row := CRC32Case{
			ID: item.ID, Kind: item.Kind, Path: item.Path, Target: item.Target,
			Recipe: item.Recipe, Outcome: item.Outcome, Execution: item.Execution,
			Argument: item.Argument, Repeat: item.Repeat, RepeatCount: item.RepeatCount,
			Expected: make(map[string]CRC32Tuple, len(fixture.Consumers)),
		}
		var recipe crc32RecipeJSON
		if item.Recipe != "" {
			var ok bool
			recipe, ok = recipes[item.Recipe]
			if !ok {
				return nil, fmt.Errorf("case %s references unknown recipe %s", item.ID, item.Recipe)
			}
		}
		for _, consumer := range fixture.Consumers {
			if item.Outcome != "" {
				want, ok := fixture.Outcomes[item.Outcome]
				if !ok {
					return nil, fmt.Errorf("case %s references unknown outcome %s", item.ID, item.Outcome)
				}
				want.Outcome = item.Outcome
				row.Expected[consumer.ID] = want
				continue
			}
			value, ok := crc32Answer(recipe, consumer.Answer)
			if !ok {
				return nil, fmt.Errorf("case %s recipe %s has no %s answer", item.ID, item.Recipe, consumer.Answer)
			}
			row.Expected[consumer.ID] = CRC32Tuple{Exit: 0, Stdout: fmt.Sprintf("%d\n", value), Outcome: "returned"}
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// CRC32ManifestControls returns the separate missing/malformed native manifest
// controls; they are tool failures, not file application outcomes.
func CRC32ManifestControls() ([]CRC32ManifestControl, error) {
	fixture, err := loadCRC32Fixture()
	if err != nil {
		return nil, err
	}
	return append([]CRC32ManifestControl(nil), fixture.ManifestControls...), nil
}

// CRC32MaterializeCase creates one row's private input and returns the exact
// token to pass after the public CLI separator. Relative opaque names stay
// relative so --json, --, and shell metacharacters exercise argv handling.
func CRC32MaterializeCase(t testing.TB, dir string, row CRC32Case, recipes map[string]CRC32OracleRow) string {
	t.Helper()
	switch row.Kind {
	case "argument", "direct-argument":
		if row.RepeatCount > 0 {
			return strings.Repeat(row.Repeat, row.RepeatCount)
		}
		return row.Argument
	case "device":
		info, err := os.Stat(row.Path)
		if err != nil || info.Mode()&os.ModeDevice == 0 {
			t.Skipf("env:device unavailable at %s", row.Path)
		}
		return row.Path
	case "missing":
		return row.Path
	case "directory":
		if err := os.Mkdir(filepath.Join(dir, row.Path), 0o700); err != nil {
			t.Fatal(err)
		}
		return row.Path
	case "fifo":
		if err := syscall.Mkfifo(filepath.Join(dir, row.Path), 0o600); err != nil {
			t.Skipf("env:fifo unavailable: %v", err)
		}
		return row.Path
	case "symlink":
		recipe, ok := recipes[row.Recipe]
		if !ok {
			t.Fatalf("symlink row %s references missing recipe %s", row.ID, row.Recipe)
		}
		if err := os.WriteFile(filepath.Join(dir, row.Target), recipe.Input, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(row.Target, filepath.Join(dir, row.Path)); err != nil {
			t.Skipf("env:input symlink unavailable: %v", err)
		}
		return row.Path
	case "mode000":
		if os.Geteuid() == 0 {
			t.Skip("env:root can read a mode-000 file")
		}
		recipe, ok := recipes[row.Recipe]
		if !ok {
			t.Fatalf("unreadable row %s references missing recipe %s", row.ID, row.Recipe)
		}
		path := filepath.Join(dir, row.Path)
		if err := os.WriteFile(path, recipe.Input, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
		return row.Path
	case "file":
		recipe, ok := recipes[row.Recipe]
		if !ok {
			t.Fatalf("file row %s references missing recipe %s", row.ID, row.Recipe)
		}
		if err := os.WriteFile(filepath.Join(dir, row.Path), recipe.Input, 0o600); err != nil {
			t.Fatal(err)
		}
		return row.Path
	default:
		t.Fatalf("unknown CRC32 case kind %q", row.Kind)
		return ""
	}
}

func crc32Answer(recipe crc32RecipeJSON, answer string) (uint64, bool) {
	var value *uint64
	switch answer {
	case "crc32":
		value = recipe.CRC32
	case "low16":
		value = recipe.Low16
	case "adler32":
		value = recipe.Adler32
	case "adler_low16":
		value = recipe.AdlerLow16
	default:
		return 0, false
	}
	if value == nil {
		return 0, false
	}
	return *value, true
}

func crc32RecipeBytes(id string) ([]byte, error) {
	switch id {
	case "empty":
		return []byte{}, nil
	case "zero":
		return []byte{0x00}, nil
	case "check":
		return []byte("123456789"), nil
	case "highbit-80":
		return []byte{0x80}, nil
	case "highbit-80ff":
		return []byte{0x80, 0xff}, nil
	case "all-bytes-00-ff":
		out := make([]byte, 256)
		for index := range out {
			out[index] = byte(index)
		}
		return out, nil
	case "cap-minus-one-ff":
		return bytes.Repeat([]byte{0xff}, CRC32Limit-1), nil
	case "cap-ff":
		return bytes.Repeat([]byte{0xff}, CRC32Limit), nil
	case "cap-plus-one-ff":
		return bytes.Repeat([]byte{0xff}, CRC32Limit+1), nil
	case "wikipedia":
		return []byte("Wikipedia"), nil
	default:
		return nil, fmt.Errorf("unknown CRC32 byte recipe %q", id)
	}
}
