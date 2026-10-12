package cgen

import "github.com/szTheory/schway/internal/compiler/core"

// Phase31FileBytesOwnerProgram exposes cgen's own file-bytes matcher to the
// external cgen_test package. EmitApplication runs core validation first, so a
// forged program can be refused by a generic law before cgen's matcher runs;
// the seeded-fault rows call this seam so cgen's matcher must refuse alone.
func Phase31FileBytesOwnerProgram(program core.Program, entry core.Function) error {
	_, err := checkFileBytesOwnerProgram(program, entry, programApplicationShell)
	return err
}
