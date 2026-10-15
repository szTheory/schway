package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhase36PublicHelperConformance(t *testing.T) {
	const source = `module phase36.red_public

export {
  fn main
}

fn crc_step(state: U64) -> U64 {
  let result = state >> 1
  result
}

fn main(unused: Byte) -> U64 {
  let state = 1
  let first = crc_step(state)
  let result = crc_step(first)
  result
}
`
	if strings.Count(source, "unused") != 1 {
		t.Fatal("public helper witness must keep its Byte parameter unused")
	}
	path := filepath.Join(t.TempDir(), "public.schway")
	if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
		t.Fatalf("write public helper witness: %v", err)
	}
	for _, engine := range []string{"interpreter", "native"} {
		t.Run(engine, func(t *testing.T) {
			phase35RunPublicCLI(t, engine, path, "1994146192", nil)
		})
	}
}
