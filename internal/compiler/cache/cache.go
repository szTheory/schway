// Package cache is a local, content-addressed store for expensive
// intermediate ARTIFACTS -- compiled binaries, instrumented (ASan/UBSan)
// binaries, per-mutant binaries -- keyed on an explicitly DECLARED input
// list (D-06-07). It is a dependency-free leaf package following
// internal/compiler/reduce's own scaffolding precedent (a directory-scoped
// package with its own _test.go sibling, importing only stdlib) -- no
// content-addressed cache existed anywhere in this tree before this
// package, so reduce is cited only for "how a brand-new package is
// scaffolded here," never for cache logic itself.
//
// The single load-bearing structural property of this package (D-06-06,
// rated one-way): it stores ARTIFACTS ONLY, never a verdict, judgement, or
// pass/fail outcome. The checker, the five-axis comparator, and the
// sanitizer classifier always re-run fresh against whatever binary --
// cached or newly built -- is on disk this invocation. A cache hit
// therefore changes what is SKIPPED (recompilation) and never what is
// ASSERTED. Caching a verdict even once would create a false-proof vector
// that no later refactor could retroactively remove from already-published
// evidence -- so this constraint is enforced structurally, by
// TestCacheExportedSurfaceStoresNoVerdict and TestCacheImportsStayIndependent
// (cache_test.go), not merely by convention.
//
// A SHA-256 digest anywhere in this package is content identity only, never
// an integrity or authenticity proof (the project's standing stance, see
// evidence.DigestClaim's sibling documentation).
package cache

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// Schema versions this package's on-disk meta.json record, independent of
// every other schema in the compiler (a new constant beside its peers --
// core.Schema1, evidence.Schema1, debugmap.Schema, ... -- never editing an
// existing one).
const Schema = "lang.cache-meta/0"

// MaxArtifactBytes and MaxMetaBytes are declared byte caps (T-06-CACHE-05):
// Get reads through them rather than accumulating an unbounded amount of
// untrusted on-disk data. Exceeding either is a fail-closed typed error,
// never silent truncation.
const (
	MaxArtifactBytes = 256 * 1024 * 1024
	MaxMetaBytes     = 1 << 20
)

// Error is this package's stable typed failure, matching the {Code}-only
// shape evidence.ValidationError/debugmap.Error already use, so callers can
// dispatch on Code identically.
type Error struct{ Code string }

func (e *Error) Error() string { return e.Code }

// Input is one declared input feeding a cache key (D-06-07): a named,
// digested fact about what produced an artifact. Digest is a content
// identity string (typically a hex SHA-256), never itself a verdict.
type Input struct {
	Name   string `json:"name"`
	Digest string `json:"digest"`
}

// Key is a content-addressed identifier over a full declared input list
// (D-06-07/D-06-09): ID is hex.EncodeToString(sha256(canonical-json(sorted
// Inputs))). Two Keys with byte-identical declared inputs always share one
// ID -- a merge by construction (FND-04 adjacency edge), never a collision.
type Key struct {
	ID     string  `json:"id"`
	Inputs []Input `json:"inputs"`
}

// ComputeKey sorts inputs by Name and hashes the canonical (sorted) form.
// It refuses a duplicate Name and refuses an empty Digest with
// Error{Code: "cache.input_undeclared"} -- ambiguity in the declared input
// list itself is refused outright, never silently accepted (D-06-11).
func ComputeKey(inputs []Input) (Key, error) {
	sorted := append([]Input(nil), inputs...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Name < sorted[j].Name })

	seen := make(map[string]bool, len(sorted))
	for _, input := range sorted {
		if input.Digest == "" {
			return Key{}, &Error{Code: "cache.input_undeclared"}
		}
		if seen[input.Name] {
			return Key{}, &Error{Code: "cache.input_undeclared"}
		}
		seen[input.Name] = true
	}

	encoded, err := json.Marshal(sorted)
	if err != nil {
		return Key{}, err
	}
	sum := sha256.Sum256(encoded)
	return Key{ID: hex.EncodeToString(sum[:]), Inputs: sorted}, nil
}

// Store is a root directory holding the sharded on-disk layout D-06-09
// specifies: <Root>/<first two hex chars of key.ID>/<key.ID>/{artifact,
// meta.json}.
type Store struct {
	Root string
}

// Open resolves the cache root via Go's os.UserCacheDir() -- NOT a
// hand-read of $XDG_CACHE_HOME -- joined with "lang-verify", so the
// cross-platform fallback (macOS/Linux/Windows) comes for free.
func Open() (*Store, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &Store{Root: filepath.Join(base, "lang-verify")}, nil
}

// Path returns the sharded directory a Key's artifact and meta.json live
// under: <Root>/<first two hex chars of key.ID>/<key.ID>/.
func (s *Store) Path(key Key) string {
	shard := key.ID
	if len(shard) > 2 {
		shard = shard[:2]
	}
	return filepath.Join(s.Root, shard, key.ID)
}

// storedMeta is the on-disk meta.json shape: the schema string, the key ID,
// and the FULL declared input list used to compute it -- the transparency
// record a reviewer can read directly (D-06-09), the analogue of
// `go build -x` / `GODEBUG=gocachehash`.
type storedMeta struct {
	Schema string  `json:"schema"`
	ID     string  `json:"id"`
	Inputs []Input `json:"inputs"`
}

// Put writes artifact and meta.json for key. Each file is written to a
// temp name in the same directory and renamed into place (os.Rename is
// atomic within one filesystem), so a crashed run between the write and
// the rename leaves no directory containing a half-written "artifact" file
// that Get would ever report as found -- Get requires meta.json to be
// present AND to match key before reporting anything as found at all.
func (s *Store) Put(key Key, artifact []byte) error {
	dir := s.Path(key)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := writeAtomic(dir, "artifact", artifact); err != nil {
		return err
	}
	meta := storedMeta{Schema: Schema, ID: key.ID, Inputs: key.Inputs}
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	return writeAtomic(dir, "meta.json", metaBytes)
}

func writeAtomic(dir, name string, data []byte) error {
	temp, err := os.CreateTemp(dir, ".tmp-"+name+"-*")
	if err != nil {
		return err
	}
	tempName := temp.Name()
	if _, err := temp.Write(data); err != nil {
		temp.Close()
		os.Remove(tempName)
		return err
	}
	if err := temp.Close(); err != nil {
		os.Remove(tempName)
		return err
	}
	return os.Rename(tempName, filepath.Join(dir, name))
}

// Get returns the artifact bytes stored under key and whether it was
// found. It reads through io.LimitReader at MaxArtifactBytes and
// strict-decodes meta.json at MaxMetaBytes; ANY anomaly -- meta.json
// absent, corrupt, or recording a declared input list that does not match
// key's -- is treated as absent, never as an error and never as a hit
// (T-06-CACHE-01: a hand-edited or partially-deleted cache directory is
// indistinguishable from a cold one by design; this bounds, but does not
// close, that accepted risk).
func (s *Store) Get(key Key) ([]byte, bool, error) {
	dir := s.Path(key)

	metaBytes, err := os.ReadFile(filepath.Join(dir, "meta.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	if len(metaBytes) > MaxMetaBytes {
		return nil, false, nil
	}
	decoder := json.NewDecoder(bytes.NewReader(metaBytes))
	decoder.DisallowUnknownFields()
	var meta storedMeta
	if err := decoder.Decode(&meta); err != nil {
		return nil, false, nil
	}
	if meta.Schema != Schema || meta.ID != key.ID || !inputsEqual(meta.Inputs, key.Inputs) {
		return nil, false, nil
	}

	file, err := os.Open(filepath.Join(dir, "artifact"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, err
	}
	defer file.Close()

	limited := io.LimitReader(file, MaxArtifactBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	if len(data) > MaxArtifactBytes {
		return nil, false, &Error{Code: "cache.artifact_too_large"}
	}
	return data, true, nil
}

func inputsEqual(a, b []Input) bool {
	if len(a) != len(b) {
		return false
	}
	for index := range a {
		if a[index].Name != b[index].Name || a[index].Digest != b[index].Digest {
			return false
		}
	}
	return true
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hashString(s string) string { return hashBytes([]byte(s)) }
