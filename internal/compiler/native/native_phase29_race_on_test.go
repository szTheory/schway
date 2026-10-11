//go:build race

package native_test

// phase29RaceEnabled labels the interpreter lane of the spike 008 harness and
// disables its default-mode ceiling: the race detector slows the interpreter by
// a large, host-dependent factor, so timing is only logged under -race.
const phase29RaceEnabled = true
