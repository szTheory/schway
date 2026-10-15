package main

import (
	"os"
	"strings"
	"testing"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

func TestPhase37GuideCleanCheckout(t *testing.T) {
	guidePath := testsupport.ProjectPath("examples", "crc32", "README.md")
	guide, err := os.ReadFile(guidePath)
	if err != nil {
		t.Fatalf("read CRC-32 guide: %v", err)
	}
	if !strings.Contains(string(guide), "```sh doc-test=setup") {
		t.Fatal("CRC-32 guide has no executable setup fence")
	}
}
