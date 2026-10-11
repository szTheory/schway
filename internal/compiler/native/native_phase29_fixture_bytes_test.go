package native_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/szTheory/schway/internal/compiler/testsupport"
)

const phase29ApplicationFlags = "-std=c17 -Wall -Wextra -Werror -pedantic -O0"

const phase29DumpDriver = `#include <stdio.h>
#include <stdlib.h>
#include "adapter.h"

int main(int argc, char **argv) {
  int i;
  for (i = 1; i < argc; i++) {
    char *end = NULL;
    unsigned long long id = strtoull(argv[i], &end, 10);
    schway_fixture_bytes_acquire_result got = schway_fixture_bytes_acquire((uint64_t)id);
    if (got.status != SCHWAY_FIXTURE_BYTES_OK) {
      printf("%llu status %d owner %s\n", id, (int)got.status, got.owner.data == NULL ? "null" : "set");
      continue;
    }
    schway_fixture_bytes_view_result view = schway_fixture_bytes_view(got.owner);
    printf("%llu ", id);
    for (uint64_t k = 0; k < view.length; k++) printf("%02x", view.data[k]);
    printf("\n");
    schway_fixture_bytes_release(got.owner);
  }
  return 0;
}
`

func phase29BuildDriver(t *testing.T, driver string, extra ...string) string {
	t.Helper()
	clang, err := exec.LookPath("clang")
	if err != nil {
		t.Skip("clang unavailable on this host")
	}
	root := t.TempDir()
	project := testsupport.ProjectPath()
	for _, name := range []string{"adapter.h", "adapter.c"} {
		data, err := os.ReadFile(filepath.Join(project, "examples", "phase29", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	driverPath := filepath.Join(root, "dump.c")
	if err := os.WriteFile(driverPath, []byte(driver), 0o600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(root, "dump")
	args := append(strings.Fields(phase29ApplicationFlags), extra...)
	args = append(args, "-I", root, "-o", binary, driverPath, filepath.Join(root, "adapter.c"))
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if out, err := exec.CommandContext(ctx, clang, args...).CombinedOutput(); err != nil {
		if len(extra) > 0 && strings.Contains(string(out), "sanitize") {
			t.Skip("address sanitizer unavailable on this host")
		}
		t.Fatalf("compile dump driver: %v\n%s", err, out)
	}
	return binary
}

func phase29RunDriver(t *testing.T, binary string, ids ...string) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, binary, ids...).CombinedOutput()
	if err != nil {
		t.Fatalf("run dump driver: %v\n%s", err, out)
	}
	return string(out)
}

func TestPhase29FixtureBytesGeneratorsAgree(t *testing.T) {
	binary := phase29BuildDriver(t, phase29DumpDriver)
	ids := testsupport.Phase29FixtureCaseIDs()
	if len(ids) != 10 {
		t.Fatalf("case ids = %v, want ten", ids)
	}
	args := make([]string, len(ids))
	for i, id := range ids {
		args[i] = fmt.Sprint(id)
	}
	lines := strings.Split(strings.TrimRight(phase29RunDriver(t, binary, args...), "\n"), "\n")
	if len(lines) != len(ids) {
		t.Fatalf("driver printed %d lines, want %d", len(lines), len(ids))
	}
	for i, id := range ids {
		want, ok := testsupport.Phase29FixtureBytes(id)
		if !ok {
			t.Fatalf("Go generator refused case %d", id)
		}
		wantLine := fmt.Sprintf("%d %x", id, want)
		if lines[i] != wantLine {
			t.Errorf("case %d: C and Go bytes differ (C len %d, Go len %d)", id, len(lines[i]), len(wantLine))
		}
		if _, ok := testsupport.Phase29FixtureExpected(id); !ok {
			t.Errorf("case %d has no pinned expected checksum", id)
		}
	}
	if _, ok := testsupport.Phase29FixtureBytes(10); ok {
		t.Error("Go generator accepted case 10")
	}
}
