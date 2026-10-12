package session

import (
	"github.com/szTheory/schway/internal/compiler/core"
	"github.com/szTheory/schway/internal/compiler/pathoracle"
)

// validateLocalOrFileBytesOwnerPaths dispatches the build-route owner-path
// replay: a program carrying the Phase 31 file-bytes owner operations is
// replayed by the file-bytes oracle, every other program by the Phase 23-25
// local-owner oracle unchanged.
func validateLocalOrFileBytesOwnerPaths(program core.Program) error {
	if pathoracle.HasFileBytesOwnerOperations(program) {
		return pathoracle.ValidateFileBytesOwnerPaths(program)
	}
	return pathoracle.ValidateLocalOwnerPaths(program)
}
