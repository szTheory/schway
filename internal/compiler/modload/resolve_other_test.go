//go:build !unix

package modload

import "testing"

// makeFIFO reports that special files are unavailable on this platform.
func makeFIFO(t *testing.T, root string) bool { return false }
