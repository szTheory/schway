package session

import (
	"github.com/szTheory/schway/internal/compiler/modload"
	"github.com/szTheory/schway/internal/compiler/native"
)

// ModuleDigestsForTest exposes moduleDigests to the external session_test
// package (plan 31-05, D-11 set-order control). Compiled only into test binaries.
func ModuleDigestsForTest(set modload.Set) []native.ModuleDigest { return moduleDigests(set) }
