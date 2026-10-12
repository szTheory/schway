package testsupport

import (
	"bytes"
	"fmt"
	"hash/adler32"
	"strings"
	"testing"
)

const checksumOracleCmdPrefix = `python3 -c 'import zlib;`

// checksumOracleRowProblems returns one message per defect in the oracle rows:
// missing provenance, or a Want that disagrees with Go hash/adler32 (the second,
// independent oracle).
func checksumOracleRowProblems(rows []ChecksumOracleRow) []string {
	var problems []string
	for _, row := range rows {
		id := row.App + "/" + row.Name
		if row.Cmd == "" {
			problems = append(problems, id+": empty Cmd")
		} else if !strings.HasPrefix(row.Cmd, checksumOracleCmdPrefix) {
			problems = append(problems, id+": Cmd does not start with "+checksumOracleCmdPrefix)
		}
		if row.Python == "" {
			problems = append(problems, id+": empty Python")
		}
		switch row.Outcome {
		case "returned":
			got := uint64(adler32.Checksum(row.Input))
			if row.App == "adler_low" {
				got %= 65536
			}
			if got != row.Want {
				problems = append(problems, fmt.Sprintf("%s: Want %d disagrees with hash/adler32 %d", id, row.Want, got))
			}
		case "FileTooLarge":
			if len(row.Input) != ChecksumLimit+1 {
				problems = append(problems, fmt.Sprintf("%s: FileTooLarge input length %d is not ChecksumLimit+1", id, len(row.Input)))
			}
			if !strings.Contains(row.Cmd, "zlib.adler32(d)") {
				problems = append(problems, id+": FileTooLarge Cmd lacks zlib.adler32(d)")
			}
		}
	}
	return problems
}

func TestPhase31OracleRows(t *testing.T) {
	rows := ChecksumOracleRows()
	if problems := checksumOracleRowProblems(rows); len(problems) != 0 {
		t.Fatalf("oracle row problems: %v", problems)
	}
	if len(rows) != 14 {
		t.Fatalf("oracle rows = %d, want 14", len(rows))
	}
	big := uint64(adler32.Checksum(bytes.Repeat([]byte{0xFF}, ChecksumLimit+1)))
	for _, row := range rows {
		if row.Outcome == "returned" && row.Want == big {
			t.Errorf("%s/%s: Want equals the 65,537x0xFF checksum %d", row.App, row.Name, big)
		}
	}
	// Boundary and empty edges, by both oracles.
	for _, row := range rows {
		if row.App == "checksum" && row.Name == "65536x0xFF" && row.Want != 2006388466 {
			t.Errorf("65536x0xFF Want = %d, want 2006388466", row.Want)
		}
		if row.Name == "empty" && row.Want != 1 {
			t.Errorf("%s empty Want = %d, want 1", row.App, row.Want)
		}
	}

	t.Run("ff65537-cmd-value", func(t *testing.T) {
		found := map[string]bool{}
		for _, row := range rows {
			if row.Name != "65537x0xFF" {
				continue
			}
			found[row.App] = true
			if len(row.Input) != ChecksumLimit+1 {
				t.Fatalf("%s input length = %d", row.App, len(row.Input))
			}
			switch row.App {
			case "checksum":
				const wantCmd = `python3 -c 'import zlib;d=b"\xff"*65537;print(len(d)>65536, zlib.adler32(d))'`
				if row.Cmd != wantCmd {
					t.Errorf("checksum Cmd = %q", row.Cmd)
				}
				// The recorded Cmd prints `True 2273841137` under Python 3.14.4.
				if got := adler32.Checksum(row.Input); got != 2273841137 {
					t.Errorf("hash/adler32 = %d, want 2273841137", got)
				}
			case "adler_low":
				if !strings.HasSuffix(row.Cmd, `zlib.adler32(d) % 65536)'`) {
					t.Errorf("adler_low Cmd = %q", row.Cmd)
				}
				// The recorded Cmd prints `True 4081`.
				if got := adler32.Checksum(row.Input) % 65536; got != 4081 {
					t.Errorf("hash/adler32 %% 65536 = %d, want 4081", got)
				}
			}
			if row.Want != 0 || row.Outcome != "FileTooLarge" {
				t.Errorf("%s: Want %d outcome %q, want 0/FileTooLarge", row.App, row.Want, row.Outcome)
			}
		}
		if !found["checksum"] || !found["adler_low"] {
			t.Fatalf("missing 65537x0xFF rows: %v", found)
		}
	})

	t.Run("empty-cmd-control", func(t *testing.T) {
		cp := append([]ChecksumOracleRow(nil), rows...)
		cp[2].Cmd = ""
		if problems := checksumOracleRowProblems(cp); len(problems) != 1 {
			t.Fatalf("problems = %v, want exactly one", problems)
		}
	})
}
