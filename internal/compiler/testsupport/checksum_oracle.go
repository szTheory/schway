package testsupport

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// ChecksumLimit is the bounded-read limit for the Phase 31 file checksum
// (D-07): a file of more than ChecksumLimit bytes is FileTooLarge.
const ChecksumLimit = 65536

// ChecksumOracleRow is one pinned (app, input) expectation. Want is the
// checksum (or its low 16 bits for adler_low) when Outcome is "returned"; it is
// 0 when Outcome is "FileTooLarge".
type ChecksumOracleRow struct {
	App, Name string
	Input     []byte
	Want      uint64
	Outcome   string
	Cmd       string
	Python    string
}

// ChecksumReadFile is the independent Go host reader for the interpreter route
// (D-07). It is written separately from the C adapter so a common-mode error
// cannot hide. It returns the file bytes with an empty alternative, or nil with
// the named FileError alternative.
func ChecksumReadFile(path string) ([]byte, string) {
	f, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, "FileUnavailable"
	}
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		_ = f.Close()
		return nil, "FileUnavailable"
	}
	data, err := io.ReadAll(io.LimitReader(f, ChecksumLimit+1))
	if err != nil {
		_ = f.Close()
		return nil, "FileUnavailable"
	}
	if err := f.Close(); err != nil {
		return nil, "FileUnavailable"
	}
	if len(data) > ChecksumLimit {
		return nil, "FileTooLarge"
	}
	if data == nil {
		data = []byte{}
	}
	return data, ""
}

// ChecksumOracleRows returns the D-12 oracle table. The values come from Python
// zlib (the recorded Cmd on each row), are cross-checked by Go hash/adler32 in
// plan 31-08, and are never derived by running Schway.
func ChecksumOracleRows() []ChecksumOracleRow {
	const python = "Python 3.14.4"
	type input struct {
		name       string
		data       []byte
		expr       string
		checksum   uint64
		low        uint64
		tooLargeBy bool
	}
	inputs := []input{
		{"empty", []byte{}, `b""`, 1, 1, false},
		{"0x00", []byte{0x00}, `b"\x00"`, 65537, 1, false},
		{"Wikipedia", []byte("Wikipedia"), `b"Wikipedia"`, 300286872, 920, false},
		{"65536x0xFF", bytes.Repeat([]byte{0xFF}, 65536), `b"\xff"*65536`, 2006388466, 3826, false},
		{"65537x0xFF", bytes.Repeat([]byte{0xFF}, 65537), `b"\xff"*65537`, 0, 0, true},
		{"0x80", []byte{0x80}, `b"\x80"`, 8454273, 129, false},
		{"0x80 0xFF", []byte{0x80, 0xFF}, `b"\x80\xff"`, 33620352, 384, false},
	}
	rows := make([]ChecksumOracleRow, 0, 2*len(inputs))
	for _, app := range []string{"checksum", "adler_low"} {
		for _, in := range inputs {
			row := ChecksumOracleRow{App: app, Name: in.name, Input: in.data, Outcome: "returned", Python: python}
			suffix := ""
			want := in.checksum
			if app == "adler_low" {
				suffix = " % 65536"
				want = in.low
			}
			row.Want = want
			if in.tooLargeBy {
				row.Outcome = "FileTooLarge"
				row.Want = 0
				row.Cmd = `python3 -c 'import zlib;d=b"\xff"*65537;print(len(d)>65536, zlib.adler32(d)` + suffix + `)'`
			} else {
				row.Cmd = `python3 -c 'import zlib;print(zlib.adler32(` + in.expr + `)` + suffix + `)'`
			}
			rows = append(rows, row)
		}
	}
	return rows
}

// ChecksumCase is one shared (input, expected result) row run on both engines
// (D-07, D-08, D-09). Exit, Stdout and Stderr are literal expectations; no
// engine is compared only against the other.
type ChecksumCase struct {
	Name, Kind string
	Content    []byte
	Exit       int
	Stdout     string
	Stderr     string
	Outcome    string
	NativeOnly bool
}

// ChecksumSecretSegment is a distinctive non-ASCII directory name every
// materialised case lives under; no error output may contain ChecksumSecret.
const (
	ChecksumSecretSegment = "phase31-secret-\u00fc-segment"
	ChecksumSecret        = "phase31-secret"
)

// ChecksumCaseMatrix returns the shared case list. Success stdout values come
// from ChecksumOracleRows, never from new literals.
func ChecksumCaseMatrix() []ChecksumCase {
	want := map[string]string{}
	var tooLarge []byte
	for _, row := range ChecksumOracleRows() {
		if row.App != "checksum" {
			continue
		}
		if row.Outcome == "returned" {
			want[row.Name] = fmt.Sprintf("%d\n", row.Want)
		} else {
			tooLarge = row.Input
		}
	}
	const (
		unavailable = "FileError.FileUnavailable\n"
		bigName     = "FileError.FileTooLarge\n"
	)
	success := func(name, oracle string, content []byte) ChecksumCase {
		return ChecksumCase{Name: name, Kind: "file", Content: content, Exit: 0, Stdout: want[oracle], Outcome: "returned"}
	}
	unavail := func(name, kind string, content []byte) ChecksumCase {
		return ChecksumCase{Name: name, Kind: kind, Content: content, Exit: 65, Stderr: unavailable, Outcome: "FileUnavailable"}
	}
	return []ChecksumCase{
		success("empty", "empty", []byte{}),
		success("zero", "0x00", []byte{0x00}),
		success("wikipedia", "Wikipedia", []byte("Wikipedia")),
		success("ff65536", "65536x0xFF", bytes.Repeat([]byte{0xFF}, 65536)),
		{Name: "ff65537", Kind: "file", Content: tooLarge, Exit: 65, Stderr: bigName, Outcome: "FileTooLarge"},
		unavail("missing", "missing", nil),
		unavail("directory", "directory", nil),
		unavail("fifo", "fifo", nil),
		unavail("devnull", "devnull", nil),
		unavail("mode000", "mode000", bytes.Repeat([]byte{0xFF}, 65537)),
	}
}

// MaterializeChecksumCase creates the input for c under dir and returns the
// path. It skips with cited env: tokens where the host cannot.
func MaterializeChecksumCase(t testing.TB, dir string, c ChecksumCase) string {
	t.Helper()
	if c.Kind == "devnull" {
		return "/dev/null"
	}
	secret := filepath.Join(dir, ChecksumSecretSegment)
	if err := os.MkdirAll(secret, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(secret, c.Name+".bin")
	switch c.Kind {
	case "file":
		if err := os.WriteFile(path, c.Content, 0o600); err != nil {
			t.Fatal(err)
		}
	case "missing":
	case "directory":
		if err := os.Mkdir(path, 0o700); err != nil {
			t.Fatal(err)
		}
	case "fifo":
		if err := syscall.Mkfifo(path, 0o600); err != nil {
			t.Skip("env:mkfifo unavailable on this host")
		}
	case "mode000":
		if os.Geteuid() == 0 {
			t.Skip("env:root mode-000 file is readable by the superuser")
		}
		if err := os.WriteFile(path, c.Content, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(path, 0); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	default:
		t.Fatalf("unknown checksum case kind %q", c.Kind)
	}
	return path
}
