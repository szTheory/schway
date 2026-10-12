package session

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
)

// checkSourcePathModules is checkSourcePath plus the module digests (Phase 31,
// D-11). It performs the one loader read; a program without imports returns
// nil digests and the unchanged byte-taking Check, so no-import identity is
// byte-identical.
func checkSourcePathModules(path string) (CheckResult, []byte, []native.ModuleDigest, error) {
	set, diagnostics, err := loadSourcePath(path)
	if err != nil {
		return CheckResult{}, nil, nil, err
	}
	entry := set.Sources[0].Bytes
	if len(set.Program.Imports) == 0 {
		return Check(entry), entry, nil, nil
	}
	return checkLoaded(set, diagnostics), entry, moduleDigests(set), nil
}

// moduleDigests hashes the bytes the loader parsed (no re-read) into rows with
// entry-relative slash paths, sorted by module name. The entry row is named by
// its declared module header.
func moduleDigests(set modload.Set) []native.ModuleDigest {
	rows := make([]native.ModuleDigest, 0, len(set.Sources))
	for i, source := range set.Sources {
		name := source.Module
		if i == 0 {
			name = set.Program.Module
		}
		sum := sha256.Sum256(source.Bytes)
		rows = append(rows, native.ModuleDigest{Module: name, Path: source.RelPath, Digest: hex.EncodeToString(sum[:])})
	}
	sort.Slice(rows, func(a, b int) bool { return rows[a].Module < rows[b].Module })
	return rows
}
