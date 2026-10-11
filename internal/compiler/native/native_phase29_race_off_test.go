//go:build !race

package native_test

// phase29RaceEnabled is false without the race detector; see the race-enabled
// twin for why the spike 008 default-mode ceiling depends on it.
const phase29RaceEnabled = false
