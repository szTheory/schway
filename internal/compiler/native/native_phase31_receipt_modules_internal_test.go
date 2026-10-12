package native

import (
	"strings"
	"testing"
)

// TestPhase31ValidateReceiptModulesTable arms WR-03 without a toolchain: the
// faithful rows pass and every malformed row is refused with its own reason.
func TestPhase31ValidateReceiptModulesTable(t *testing.T) {
	const entryDigest = "0000000000000000000000000000000000000000000000000000000000000001"
	const importDigest = "0000000000000000000000000000000000000000000000000000000000000002"
	build := func(entryPath, importPath string) BuildReceipt {
		modules := []ModuleDigest{
			{Module: "a.entry", Path: entryPath, Digest: entryDigest},
			{Module: "b.import", Path: importPath, Digest: importDigest},
		}
		return BuildReceipt{SourceDigest: entryDigest, Modules: modules, Identity: &BuildIdentityInputs{Modules: modules}}
	}
	for _, tc := range []struct {
		name         string
		entry, other string
		want         string // empty means accepted
	}{
		{"faithful", "checksum.schway", "adler32.schway", ""},
		{"faithful-nested", "src/checksum.schway", "lib/adler32.schway", ""},
		{"empty-segment", "a//b.schway", "adler32.schway", "canonical"},
		{"trailing-slash", "dir/", "adler32.schway", "canonical"},
		{"dot-segment", "./checksum.schway", "adler32.schway", "canonical"},
		{"inner-dot-segment", "dir/./x.schway", "adler32.schway", "canonical"},
		{"drive-prefix", "checksum.schway", "C:adler32.schway", "canonical"},
		{"drive-absolute", "checksum.schway", `C:\adler32.schway`, "relative"},
		{"dotdot", "checksum.schway", "../adler32.schway", ".."},
		{"absolute", "/etc/checksum.schway", "adler32.schway", "relative"},
		{"duplicate-path", "checksum.schway", "checksum.schway", "duplicated"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateReceiptModules(build(tc.entry, tc.other))
			if tc.want == "" {
				if err != nil {
					t.Fatalf("faithful receipt refused: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err = %v, want a refusal containing %q", err, tc.want)
			}
		})
	}
	t.Run("duplicate-module-name", func(t *testing.T) {
		receipt := build("checksum.schway", "adler32.schway")
		receipt.Modules[1].Module = receipt.Modules[0].Module
		receipt.Identity.Modules = receipt.Modules
		if err := validateReceiptModules(receipt); err == nil || !strings.Contains(err.Error(), "unsorted or duplicated") {
			t.Fatalf("err = %v, want the duplicate module name refusal", err)
		}
	})
}
