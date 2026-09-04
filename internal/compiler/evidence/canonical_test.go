package evidence

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/codename-lang/lang/internal/compiler/syntax"
)

// The canonical projection is what SourceDigest binds, so a formatter that
// loses or alters semantics must be rejected at the point the projection is
// established rather than surfacing later as a diagnostic against text the
// user never wrote. These cases drive the seam directly so they stay valid
// once the underlying formatter defect is repaired.
func TestCanonicalRoundTripFailsClosed(t *testing.T) {
	source := readCanonicalSource(t)
	for _, test := range []struct {
		name   string
		format func(syntax.Tree) []byte
	}{
		{
			// Canonical bytes that do not reparse: previously this fed a
			// half-parsed program into check and fabricated a source
			// diagnostic at offsets in the corrupted projection.
			name: "canonical_does_not_reparse",
			format: func(tree syntax.Tree) []byte {
				return bytes.Replace(syntax.Format(tree), []byte("fn "), []byte("fn@ "), 1)
			},
		},
		{
			// Canonical bytes that reparse cleanly but are not a fixed
			// point: the digest would bind a projection the formatter does
			// not itself agree is canonical.
			name: "canonical_not_fixed_point",
			format: driftingFormat(),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			product, diagnostics, err := build(source, canonicalFacts(), test.format, syntax.Parse)
			if err == nil {
				t.Fatalf("unstable canonical projection was accepted: diagnostics=%v manifest=%+v", diagnostics, product.Manifest)
			}
			if len(diagnostics) > 0 {
				t.Fatalf("integrity failure was laundered into source diagnostics: %v", diagnostics)
			}
			if code := ErrorCode(err); code != "evidence.canonical_unstable" {
				t.Fatalf("unstable canonical projection reported %q, want evidence.canonical_unstable (err=%v)", code, err)
			}
			if product.ManifestBytes != nil || product.CanonicalSource != nil {
				t.Fatalf("rejected build still returned products: %+v", product)
			}
		})
	}
}

func TestCanonicalRoundTripAdmitsValidPrograms(t *testing.T) {
	source := readCanonicalSource(t)
	product, diagnostics, err := Build(source, canonicalFacts())
	if err != nil || len(diagnostics) > 0 {
		t.Fatalf("valid program rejected by canonical round trip: err=%v diagnostics=%v", err, diagnostics)
	}
	if len(product.ManifestBytes) == 0 {
		t.Fatal("valid program produced no manifest")
	}
	// The admitted canonical source must itself be the fixed point the check
	// asserts, so the digest binds a stable projection.
	if !bytes.Equal(product.CanonicalSource, syntax.Format(syntax.Parse(product.CanonicalSource).Tree)) {
		t.Fatal("accepted canonical source is not a formatter fixed point")
	}
}

// driftingFormat models a formatter whose second pass disagrees with its
// first: the bytes still reparse cleanly, but they are not a fixed point, so
// the digest would bind a projection the formatter does not itself call
// canonical.
func driftingFormat() func(syntax.Tree) []byte {
	passes := 0
	return func(tree syntax.Tree) []byte {
		passes++
		formatted := syntax.Format(tree)
		if passes > 1 {
			formatted = append(formatted, '\n')
		}
		return formatted
	}
}

func readCanonicalSource(t testing.TB) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "phase2", "owned_transfer.lang"))
	if err != nil {
		t.Fatalf("read owned transfer source: %v", err)
	}
	return data
}

func canonicalFacts() Facts {
	return Facts{
		CompilerIdentity: "codename-lang-stage0/test",
		ClangIdentity:    "clang-test",
		Target:           "test-target",
		Flags:            append([]string(nil), DefaultFlags...),
		Policy:           "phase1-pure-c17-v1",
	}
}
