// Package modload is the only Phase 30 package that opens module files (D-10).
//
// Modules are whole-program composition, not separate compilation: the loader
// resolves a closed set of source files under one entry directory, and the
// result is checked, validated and emitted as one program. resolve.go holds the
// resolver every module read goes through; this file holds the one place that
// opens the confinement root. Nothing here wires itself into a CLI route; later
// Phase 30 plans compose it.
package modload

import "os"

// openModuleRoot opens dir as the confinement root for module reads. It is the
// only os.OpenRoot call in the repository outside tests.
func openModuleRoot(dir string) (*os.Root, error) {
	return os.OpenRoot(dir)
}
