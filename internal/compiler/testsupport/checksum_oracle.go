package testsupport

import (
	"bytes"
	"io"
	"os"
	"syscall"
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
